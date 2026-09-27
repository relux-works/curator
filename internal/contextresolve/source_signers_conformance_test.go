package contextresolve

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
)

type sourceSignerVectorFile struct {
	VerificationCases []sourceSignerVector `json:"verification_cases"`
}

type sourceSignerVector struct {
	Name                 string                         `json:"name"`
	Conforming           *bool                          `json:"conforming"`
	SourceKind           string                         `json:"source_kind"`
	Source               *string                        `json:"source"`
	Allowlist            []sourceSignerVectorSigner     `json:"allowlist"`
	RequireSourceSigners bool                           `json:"require_source_signers"`
	Candidates           []sourceSignerVectorCandidate  `json:"candidates"`
	Expected             *sourceSignerVectorExpected    `json:"expected"`
	Observed             *sourceSignerVectorObservation `json:"observed"`
}

type sourceSignerVectorSigner struct {
	Type        string `json:"type"`
	Key         string `json:"key"`
	Fingerprint string `json:"fingerprint"`
}

type sourceSignerVectorCandidate struct {
	Version         string                      `json:"version"`
	Tag             string                      `json:"tag"`
	Commit          string                      `json:"commit"`
	Form            string                      `json:"form"`
	TagSignature    *sourceSignerVectorEvidence `json:"tag_signature"`
	CommitSignature *sourceSignerVectorEvidence `json:"commit_signature"`
}

type sourceSignerVectorEvidence struct {
	Signer sourceSignerVectorSigner `json:"signer"`
	Valid  bool                     `json:"valid"`
}

type sourceSignerVectorExpected struct {
	Verdict     string  `json:"verdict"`
	Diagnostic  string  `json:"diagnostic"`
	Selected    *string `json:"selected"`
	LockWritten bool    `json:"lock_written"`
}

type sourceSignerVectorObservation struct {
	Verdict  string  `json:"verdict"`
	Selected *string `json:"selected"`
}

type vectorVerifierSource struct {
	*stubSource
	evidenceByCommit map[string]CandidateSignatures
	verifiedCommits  []string
}

func (s *vectorVerifierSource) VerifyCandidate(_, _, _, _, commit string, _ []Signer) (CandidateSignatures, error) {
	s.verifiedCommits = append(s.verifiedCommits, commit)
	return s.evidenceByCommit[commit], nil
}

// TestSourceSignerVectorsAtResolve drives every rc.13 signer verification
// case through Resolve, including the adversarial observations that the
// protocol requires the manager to reject.
func TestSourceSignerVectorsAtResolve(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "environments-source-signers.json")) // #nosec G304 -- explicit conformance root
	if err != nil {
		t.Fatal(err)
	}
	var vectors sourceSignerVectorFile
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	if len(vectors.VerificationCases) == 0 {
		t.Fatal("pinned rc.13 source signer vector publishes no verification cases")
	}
	for _, vector := range vectors.VerificationCases {
		t.Run(vector.Name, func(t *testing.T) {
			verifySourceSignerVector(t, vector)
		})
	}
}

func verifySourceSignerVector(t *testing.T, vector sourceSignerVector) {
	t.Helper()
	base := &stubSource{commits: map[string]string{}, manifests: map[string]*Package{}}
	verifier := &vectorVerifierSource{stubSource: base, evidenceByCommit: map[string]CandidateSignatures{}}
	input := Input{
		Root:                 Requirement{Kind: contextlock.KindContext, Name: "root"},
		RequireSourceSigners: vector.RequireSourceSigners,
	}
	if vector.SourceKind == "path" {
		if len(vector.Candidates) != 1 || vector.Candidates[0].Version == "" {
			t.Fatalf("path vector has no single versioned candidate: %+v", vector.Candidates)
		}
		input.RootState = &StatePackage{
			StateHash: strings.Repeat("f", 64),
			Manifest:  &Package{Version: vector.Candidates[0].Version},
		}
	} else {
		if vector.Source == nil || *vector.Source == "" {
			t.Fatal("git signer vector has no source identity")
		}
		input.Root.Source = *vector.Source
		for _, candidate := range vector.Candidates {
			commit := candidate.Commit
			if commit == "" {
				commit = strings.Repeat("1", 40)
			}
			base.manifests[commit] = &Package{Version: candidate.Version}
			if candidate.Tag != "" {
				base.commits["root@"+candidate.Tag] = commit
			}
			verifier.evidenceByCommit[commit] = CandidateSignatures{
				Tag:    vectorSignature(candidate.TagSignature),
				Commit: vectorSignature(candidate.CommitSignature),
			}
		}
		if vector.Allowlist != nil {
			allowed := make([]Signer, 0, len(vector.Allowlist))
			for _, signer := range vector.Allowlist {
				allowed = append(allowed, Signer(signer))
			}
			input.SourceSigners = map[string][]Signer{*vector.Source: allowed}
		}
		if len(vector.Candidates) == 0 {
			t.Fatal("git signer vector publishes no candidates")
		}
		switch vector.Candidates[0].Form {
		case "tag":
			input.Root.Tag = vector.Candidates[0].Tag
		case "revision":
			input.Root.Revision = vector.Candidates[0].Commit
		default:
			input.Root.Range = "*"
		}
	}

	result, resolveErr := Resolve(verifier, input)
	if vector.Conforming != nil && !*vector.Conforming {
		if resolveErr == nil {
			t.Fatalf("resolver accepted adversarial claim %q (observed %s)", vector.Name, vector.Observed.Verdict)
		}
		resolutionErr := resolveError(t, resolveErr)
		if resolutionErr.Diagnostic != DiagSourceUnsigned {
			t.Fatalf("diagnostic = %q, want %q", resolutionErr.Diagnostic, DiagSourceUnsigned)
		}
		if vector.Name == "fallback-selection" && (len(verifier.verifiedCommits) != 1 || verifier.verifiedCommits[0] != vector.Candidates[0].Commit) {
			t.Fatalf("verification commits = %v, want only highest candidate %s", verifier.verifiedCommits, vector.Candidates[0].Commit)
		}
		return
	}
	if vector.Expected == nil {
		t.Fatal("conforming vector has no expected result")
	}
	accepted := vector.Expected.Verdict == "accepted"
	if accepted != (resolveErr == nil) {
		t.Fatalf("Resolve error = %v, want accepted=%t", resolveErr, accepted)
	}
	if accepted {
		if result == nil || vector.Expected.LockWritten != (result.Lock != nil) {
			t.Fatalf("lock result = %+v, want lock_written=%t", result, vector.Expected.LockWritten)
		}
		member, ok := result.Members[Key(contextlock.KindContext, "root")]
		if !ok || vector.Expected.Selected != nil && member.Version != *vector.Expected.Selected {
			t.Fatalf("selected root = %+v, want %v", member, vector.Expected.Selected)
		}
	} else {
		resolutionErr := resolveError(t, resolveErr)
		if resolutionErr.Diagnostic != vector.Expected.Diagnostic {
			t.Fatalf("diagnostic = %q, want %q", resolutionErr.Diagnostic, vector.Expected.Diagnostic)
		}
		if result != nil {
			t.Fatalf("refused vector produced result %+v", result)
		}
	}
	if (vector.SourceKind == "path" || vector.Allowlist == nil && !vector.RequireSourceSigners) && len(verifier.verifiedCommits) != 0 {
		t.Fatalf("unconfigured or path source was verified: %v", verifier.verifiedCommits)
	}
}

func vectorSignature(vector *sourceSignerVectorEvidence) SignatureEvidence {
	if vector == nil {
		return SignatureEvidence{}
	}
	return SignatureEvidence{
		Present: true,
		Valid:   vector.Valid,
		Signer: Signer{
			Type: vector.Signer.Type, Key: vector.Signer.Key, Fingerprint: vector.Signer.Fingerprint,
		},
	}
}
