package install

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/registry"
	"github.com/relux-works/curator/internal/sourcelock"
)

// draftEvidenceStub serves one signed registry whose records echo the
// query keys, with a programmable name/content override. It is the
// install-package counterpart of the crossconformance attest stub,
// narrowed to the draft §4 exact-matching bound.
type draftEvidenceStub struct {
	t       *testing.T
	server  *httptest.Server
	pinned  string
	private ed25519.PrivateKey
	mutate  func(body map[string]any)
}

func newDraftEvidenceStub(t *testing.T, mutate func(body map[string]any)) *draftEvidenceStub {
	t.Helper()
	public, private, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	stub := &draftEvidenceStub{
		t:       t,
		pinned:  "ed25519:" + base64.StdEncoding.EncodeToString(public),
		private: private,
		mutate:  mutate,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/snapshot", stub.serveSnapshot)
	mux.HandleFunc("/v1/records", stub.serveRecords)
	stub.server = httptest.NewServer(mux)
	t.Cleanup(stub.server.Close)
	return stub
}

func (s *draftEvidenceStub) sign(body map[string]any) map[string]any {
	record := map[string]any{}
	for key, value := range body {
		if key != "sig" {
			record[key] = value
		}
	}
	signature := ed25519.Sign(s.private, registry.CanonicalBytes(record))
	public, err := registry.ParsePublicKey(s.pinned)
	if err != nil {
		s.t.Fatal(err)
	}
	record["sig"] = map[string]any{
		"key_id":    registry.KeyID(public),
		"algorithm": "ed25519",
		"signature": base64.StdEncoding.EncodeToString(signature),
	}
	return record
}

func (s *draftEvidenceStub) serveSnapshot(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.sign(map[string]any{
		"schema_version": 1, "version": 1, "log_size": 0,
		"head": strings.Repeat("ab", 32), "merkle_root": strings.Repeat("cd", 32),
		"created_at": time.Now().UTC().Truncate(time.Second).Format("2006-01-02T15:04:05Z"),
	}))
}

func (s *draftEvidenceStub) serveRecords(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query()
	body := map[string]any{
		"name": "review", "source_identity": query.Get("source_identity"),
		"commit": query.Get("commit"), "content_sha256": query.Get("content_sha256"),
		"status": registry.StatusAudited, "audit": map[string]any{"auditor": "team"},
	}
	if s.mutate != nil {
		s.mutate(body)
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"records": []any{s.sign(body)}, "next_cursor": nil})
}

// TestDraftEvidenceExactMatch is the production-entry bound for draft §4
// (install.Project -> resolveRegistries -> registry.ResolveExact): exact
// evidence installs and lands its attestation in the marker, while a
// wrong-name, wrong-repository, wrong-commit, or wrong-context record is
// refused fail-closed under a strict policy with the shared typed refusal,
// preserving the prior lock and install and revealing no registry endpoint
// or key material. Narrowing either the repository or commit comparison to
// a non-empty check admits its corresponding single-field mismatch row.
func TestDraftEvidenceExactMatch(t *testing.T) {
	run := func(t *testing.T, mutate func(body map[string]any)) (Result, *draftEvidenceStub, map[string][]byte) {
		t.Helper()
		project, home, _, _ := setupGitInstall(t)
		skillsRoot := t.TempDir()
		installOpts := Options{Platform: installPlatform()}
		prior := Project(draftTestConfig(home, skillsRoot), project, "test", installOpts)
		if prior.Status != "ok" {
			t.Fatalf("seed install = %+v, want ok", prior)
		}
		installed := filepath.Join(project, ".agents", "skills", "review")
		statePaths := []string{
			sourcelock.PathIn(project),
			DraftBindingsPath(home, project),
			filepath.Join(installed, marker.Name),
			filepath.Join(installed, "SKILL.md"),
			filepath.Join(installed, "references", "info.md"),
		}
		priorState := make(map[string][]byte, len(statePaths))
		for _, path := range statePaths {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read seeded state %s: %v", path, err)
			}
			priorState[path] = data
		}
		stub := newDraftEvidenceStub(t, mutate)
		cfg := draftTestConfig(home, skillsRoot)
		cfg.Audit.RegistryPolicy = "strict"
		cfg.AuditRegistries = []config.Registry{{Name: "one", URL: stub.server.URL, PublicKeys: []string{stub.pinned}, Enabled: true}}
		cfg.Audit.CacheTTLSeconds = 0
		cfg.Audit.OfflineGraceSeconds = 0
		return Project(cfg, project, "test", Options{Platform: installPlatform()}), stub, priorState
	}
	t.Run("exact-admits", func(t *testing.T) {
		project, home, _, _ := setupGitInstall(t)
		stub := newDraftEvidenceStub(t, nil)
		cfg := draftTestConfig(home, t.TempDir())
		cfg.Audit.RegistryPolicy = "strict"
		cfg.AuditRegistries = []config.Registry{{Name: "one", URL: stub.server.URL, PublicKeys: []string{stub.pinned}, Enabled: true}}
		cfg.Audit.CacheTTLSeconds = 0
		cfg.Audit.OfflineGraceSeconds = 0
		result := Project(cfg, project, "test", Options{Platform: installPlatform()})
		if result.Status != "ok" {
			t.Fatalf("exact evidence install = %+v, want ok", result)
		}
		recorded := marker.Read(project + "/.agents/skills/review")
		if recorded == nil || recorded.Attestation == nil || recorded.Attestation.Status != registry.StatusAudited {
			t.Fatalf("marker attestation = %+v, want audited", recorded)
		}
	})
	for _, condition := range []struct {
		name   string
		mutate func(body map[string]any)
	}{
		{"wrong-name", func(body map[string]any) { body["name"] = "other" }},
		{"wrong-context", func(body map[string]any) {
			body["content_sha256"] = "sha256:" + strings.Repeat("e", 64)
		}},
		{"wrong-repository-only", func(body map[string]any) {
			body["source_identity"] = "unrelated.test/kit"
		}},
		{"wrong-commit-only", func(body map[string]any) {
			body["commit"] = strings.Repeat("e", 40)
		}},
	} {
		t.Run(condition.name+"-refuses", func(t *testing.T) {
			result, stub, priorState := run(t, condition.mutate)
			if result.Status != "failed" {
				t.Fatalf("%s install = %+v, want refusal", condition.name, result)
			}
			joined := strings.Join(result.Errors, ";")
			if !strings.Contains(joined, "is not audited by any trusted registry") {
				t.Fatalf("%s errors = %q, want the strict typed refusal", condition.name, joined)
			}
			if len(result.Attestations) != 0 {
				t.Fatalf("%s refusal exposed attestations: %+v", condition.name, result.Attestations)
			}
			diagnostics := strings.Join(append(append([]string(nil), result.Errors...), result.Messages...), ";")
			for _, secret := range []string{stub.server.URL, stub.pinned, "ed25519:", "127.0.0.1"} {
				if strings.Contains(diagnostics, secret) {
					t.Fatalf("%s refusal leaks %q: %q", condition.name, secret, diagnostics)
				}
			}
			for path, before := range priorState {
				after, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("%s refusal removed prior state %s: %v", condition.name, path, err)
				}
				if string(after) != string(before) {
					t.Fatalf("%s refusal changed prior state %s", condition.name, path)
				}
			}
		})
	}
}
