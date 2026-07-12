package install

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/identity"
	markerpkg "github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/registry"
	"github.com/relux-works/curator/internal/snapshot"
)

// fakeRegistry serves signed snapshot and records for one artifact status.
func fakeRegistry(t *testing.T, status, sourceIdentity, commit, contentHash string) (*httptest.Server, string) {
	t.Helper()
	public, private, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	pinned := "ed25519:" + base64.StdEncoding.EncodeToString(public)
	sign := func(body map[string]any) map[string]any {
		signature := ed25519.Sign(private, registry.CanonicalBytes(body))
		body["sig"] = map[string]any{
			"key_id": registry.KeyID(public), "algorithm": "ed25519",
			"signature": base64.StdEncoding.EncodeToString(signature),
		}
		return body
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/v1/snapshot"):
			_ = json.NewEncoder(w).Encode(sign(map[string]any{
				"schema_version": 1, "merkle_root": "r", "log_size": 1, "head": "h",
				"version": 1, "created_at": time.Now().UTC().Format(time.RFC3339),
			}))
		case strings.HasSuffix(r.URL.Path, "/v1/records"):
			record := sign(map[string]any{
				"name": "skill-a", "source_identity": sourceIdentity,
				"commit": commit, "content_sha256": contentHash, "status": status,
			})
			_ = json.NewEncoder(w).Encode(map[string]any{"records": []any{record}})
		default:
			http.NotFound(w, r)
		}
	}))
	return server, pinned
}

func registryEnv(t *testing.T, status string) (*env, *httptest.Server) {
	t.Helper()
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")
	// give the declaration a git URL so the artifact has an identity
	e.write(e.project, "Skillfile.json", `{
		"schema_version": 1, "agents": ["claude_code"],
		"skills": [{"name": "skill-a", "git": "git@git.example.com:skills/skill-a.git", "tag": "v1"}]
	}`)
	// but keep resolution local: the repo already exists under skills root
	ref, err := gitops.Resolve(e.skillsRoot+"/skill-a", "tag", "v1")
	if err != nil {
		t.Fatal(err)
	}
	snap, err := snapshot.Get(e.home, "skill-a", e.skillsRoot+"/skill-a", ref.Commit)
	if err != nil {
		t.Fatal(err)
	}
	contentHash, err := hashing.ContentSHA256(snap, nil)
	if err != nil {
		t.Fatal(err)
	}
	id := identity.Canonical("git@git.example.com:skills/skill-a.git")
	server, pinned := fakeRegistry(t, status, id, ref.Commit, contentHash)
	e.cfg.AuditRegistries = []config.Registry{{Name: "test-reg", URL: server.URL, PublicKeys: []string{pinned}, Enabled: true}}
	return e, server
}

func TestRegistryRevocationDeniesInstall(t *testing.T) {
	e, server := registryEnv(t, "revoked")
	defer server.Close()
	result := e.install(Options{})
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, "\n"), "revoked by test-reg") {
		t.Fatalf("revocation must deny: %+v", result)
	}
}

func TestRegistryAttestationLandsInMarker(t *testing.T) {
	e, server := registryEnv(t, "audited")
	defer server.Close()
	result := e.install(Options{})
	if result.Status != "ok" {
		t.Fatalf("install: %+v", result)
	}
	recorded := readMarkerFor(t, e, "skill-a")
	if recorded.Attestation == nil || recorded.Attestation.Registry != "test-reg" || recorded.Attestation.Status != "audited" {
		t.Fatalf("attestation: %+v", recorded.Attestation)
	}
}

func TestStrictRegistryPolicyFailsUnknown(t *testing.T) {
	e, server := registryEnv(t, "pending") // pending resolves as unknown
	defer server.Close()
	e.cfg.Audit.RegistryPolicy = "strict"
	result := e.install(Options{})
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, "\n"), "registry_policy is strict") {
		t.Fatalf("strict policy must fail unknown: %+v", result)
	}
	e.cfg.Audit.RegistryPolicy = "advisory"
	result = e.install(Options{})
	if result.Status != "ok" {
		t.Fatalf("advisory must pass unknown: %+v", result)
	}
}

func readMarkerFor(t *testing.T, e *env, name string) *markerpkg.Marker {
	t.Helper()
	m := markerpkg.Read(e.project + "/.agents/skills/" + name)
	if m == nil {
		t.Fatalf("marker missing for %s", name)
	}
	return m
}
