package contextresolve

import (
	"strings"
	"testing"
)

// These cases prove per-source key authorization at Resolve, assuming the
// supplied cryptographic evidence. Real public signatures are checked in
// envprofile. The last two controls retain the normative commit/revision
// contract and expose the remaining release-tag-only policy gap explicitly.
func TestReleaseKeySourceAllowlist(t *testing.T) {
	platform := Signer{Type: "ssh", Key: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIF9OaCfCki1tEczuA4hE2DaaToEpFteweRW3IBJkwbgT"}
	curator := Signer{Type: "ssh", Key: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIF78aQWisJvEInj43zlVAGDUrzLpj36KEd6FPNqe4BvK"}
	commitKey := Signer{Type: "ssh", Key: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOZGXWazfNwZqHl7TqTq9MZ35JM3SQ3SIIky4LUTPO+r"}
	valid := func(s Signer) SignatureEvidence { return SignatureEvidence{Present: true, Valid: true, Signer: s} }
	const skillSource = "github.com/relux-works/skill-pdf"
	const curatorSource = "github.com/relux-works/curator"
	for _, tc := range []struct {
		name, source, diagnostic string
		evidence                 CandidateSignatures
		revision                 bool
	}{
		{name: "platform-skill-tag", source: skillSource, evidence: CandidateSignatures{Tag: valid(platform)}},
		{name: "curator-family-tag", source: curatorSource, evidence: CandidateSignatures{Tag: valid(curator)}},
		{name: "commit-key-tag-refused", source: skillSource, evidence: CandidateSignatures{Tag: valid(commitKey)}, diagnostic: DiagSourceSignerRejected},
		{name: "unsigned-refused", source: skillSource, diagnostic: DiagSourceUnsigned},
		{name: "platform-on-curator-refused", source: curatorSource, evidence: CandidateSignatures{Tag: valid(platform)}, diagnostic: DiagSourceSignerRejected},
		{name: "curator-on-platform-refused", source: skillSource, evidence: CandidateSignatures{Tag: valid(curator)}, diagnostic: DiagSourceSignerRejected},
		{name: "legacy-allowed-commit-still-authorizes", source: skillSource, evidence: CandidateSignatures{Commit: valid(platform)}},
		{name: "supported-signed-revision-preserved", source: curatorSource, evidence: CandidateSignatures{Commit: valid(curator)}, revision: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			commit := strings.Repeat("1", 40)
			source := &signerStubSource{
				stubSource: &stubSource{commits: map[string]string{"root@v1.0.0": commit}, manifests: map[string]*Package{commit: {Version: "1.0.0"}}},
				evidence:   tc.evidence,
			}
			input := Input{
				Root:          Requirement{Name: "root", Source: tc.source, Tag: "v1.0.0"},
				SourceSigners: map[string][]Signer{skillSource: {platform}, curatorSource: {curator}},
			}
			if tc.revision {
				input.Root.Tag, input.Root.Revision = "", commit
			}
			result, err := Resolve(source, input)
			if tc.diagnostic != "" {
				if result != nil || resolveError(t, err).Diagnostic != tc.diagnostic {
					t.Fatalf("refusal result=%+v error=%v; want %s and no lock", result, err, tc.diagnostic)
				}
				return
			}
			if err != nil || result == nil || result.Lock == nil || len(result.Lock.Members) != 1 || result.Lock.Members[0].Commit != commit {
				t.Fatalf("accepted source did not retain selected pin: result=%+v error=%v", result, err)
			}
		})
	}
}
