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

// registryFixture describes how the fake registry stamps and serves its
// snapshot. A published snapshot is minted once and then served; minting
// created_at inside the handler instead makes the timestamp advance past the
// `now` the checker already sampled, which is the defect this file carried.
type registryFixture struct {
	createdAt      time.Time
	beforeSnapshot func()
}

type registryOption func(*registryFixture)

// snapshotCreatedAt stamps the served snapshot at an explicit instant.
func snapshotCreatedAt(at time.Time) registryOption {
	return func(f *registryFixture) { f.createdAt = at.UTC().Truncate(time.Second) }
}

// beforeSnapshotResponse runs just before the snapshot response is written.
func beforeSnapshotResponse(hook func()) registryOption {
	return func(f *registryFixture) { f.beforeSnapshot = hook }
}

// crossSecondBoundary blocks until just past the next whole second, so the
// snapshot response is written in a later wall-clock second than the one in
// which the checker sampled its `now`. RFC3339 truncates created_at down to
// that later second, which is how a per-request timestamp lands ahead of a
// `now` that was sampled before the fetch (BUG-260906-1bdotx).
func crossSecondBoundary() {
	now := time.Now()
	time.Sleep(now.Truncate(time.Second).Add(time.Second + 2*time.Millisecond).Sub(now))
}

// fakeRegistry serves signed snapshot and records for one artifact status.
func fakeRegistry(t *testing.T, status, sourceIdentity, commit, contentHash string, options ...registryOption) (*httptest.Server, string) {
	t.Helper()
	public, private, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	// Mint the snapshot timestamp once, here, and never again: the served
	// snapshot must not move relative to the clock reading the checker takes
	// before it fetches.
	fixture := registryFixture{createdAt: time.Now().UTC().Truncate(time.Second)}
	for _, option := range options {
		option(&fixture)
	}
	createdAt := fixture.createdAt.Format(time.RFC3339)
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
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/v1/snapshot"):
			if fixture.beforeSnapshot != nil {
				fixture.beforeSnapshot()
			}
			_ = json.NewEncoder(w).Encode(sign(map[string]any{
				"schema_version": 1, "merkle_root": strings.Repeat("a", 64), "log_size": 1, "head": strings.Repeat("b", 64),
				"version": 1, "created_at": createdAt,
			}))
		case strings.HasSuffix(r.URL.Path, "/v1/records"):
			record := sign(map[string]any{
				"name": "skill-a", "source_identity": sourceIdentity,
				"commit": commit, "content_sha256": contentHash, "status": status,
			})
			_ = json.NewEncoder(w).Encode(map[string]any{"records": []any{record}, "next_cursor": nil})
		default:
			http.NotFound(w, r)
		}
	}))
	return server, pinned
}

func registryEnv(t *testing.T, status string, options ...registryOption) (*env, *httptest.Server) {
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
	server, pinned := fakeRegistry(t, status, id, ref.Commit, contentHash, options...)
	e.cfg.AuditRegistries = []config.Registry{{Name: "test-reg", URL: server.URL, PublicKeys: []string{pinned}, Enabled: true}}
	return e, server
}

func TestRegistryRevocationDeniesInstall(t *testing.T) {
	t.Parallel()
	e, server := registryEnv(t, "revoked")
	defer server.Close()
	result := e.install(Options{})
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, "\n"), "revoked by test-reg") {
		t.Fatalf("revocation must deny: %+v", result)
	}
}

func TestRegistryAttestationLandsInMarker(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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

// TestRegistrySnapshotSurvivesASecondBoundaryDuringFetch is the regression
// test for BUG-260906-1bdotx. resolveRegistries samples `now` before it
// fetches; a fixture that stamped created_at inside the handler produced a
// timestamp in a later whole second than that `now` whenever the fetch crossed
// a second boundary, and the e2e config carries a literal zero clock skew, so
// any positive difference read as tampering. The hook makes that crossing
// certain instead of leaving it to the runner's speed.
func TestRegistrySnapshotSurvivesASecondBoundaryDuringFetch(t *testing.T) {
	t.Parallel()
	e, server := registryEnv(t, "audited", beforeSnapshotResponse(crossSecondBoundary))
	defer server.Close()
	result := e.install(Options{})
	if result.Status != "ok" {
		t.Fatalf("a snapshot minted before the run must survive a slow fetch: %+v", result)
	}
	if joined := strings.Join(result.Messages, "\n"); strings.Contains(joined, "too far in the future") {
		t.Fatalf("boundary crossing was read as a future timestamp: %s", joined)
	}
}

// TestRegistryFutureSnapshotDeniesInstallThroughResolveRegistries drives the
// production install path with a genuinely future-dated snapshot. A future
// snapshot is what a rollback or equivocation attempt looks like, so this must
// deny the install, not warn: it names the call site the gate has to be
// reachable from (install.resolveRegistries -> registry.CheckSnapshotsWithPolicy).
func TestRegistryFutureSnapshotDeniesInstallThroughResolveRegistries(t *testing.T) {
	t.Parallel()
	e, server := registryEnv(t, "audited", snapshotCreatedAt(time.Now().Add(time.Hour)))
	defer server.Close()
	result := e.install(Options{})
	if result.Status != "failed" {
		t.Fatalf("a future-dated snapshot must deny the install: %+v", result)
	}
	if !strings.Contains(strings.Join(result.Errors, "\n"), "every trusted audit registry served a tampered snapshot") {
		t.Fatalf("future-dated snapshot must exclude the registry: %+v", result.Errors)
	}
	if !strings.Contains(strings.Join(result.Messages, "\n"), "registry test-reg snapshot timestamp is too far in the future") {
		t.Fatalf("the refusal must name the future timestamp: %+v", result.Messages)
	}
}

// TestRegistrySnapshotWithinTheBoundIsAcceptedThroughInstall is the positive
// half of the same bound at the production entry point: an ordinarily
// past-dated snapshot installs. It does not sit at the edge -- the exact edge
// is pinned in internal/registry -- and its job here is to stop the refusal
// above being satisfied by a gate that rejects every snapshot.
func TestRegistrySnapshotWithinTheBoundIsAcceptedThroughInstall(t *testing.T) {
	t.Parallel()
	e, server := registryEnv(t, "audited", snapshotCreatedAt(time.Now().Add(-time.Minute)))
	defer server.Close()
	e.cfg.Audit.SnapshotClockSkewSeconds = 0
	result := e.install(Options{})
	if result.Status != "ok" {
		t.Fatalf("a past-dated snapshot must install under a literal zero skew: %+v", result)
	}
}
