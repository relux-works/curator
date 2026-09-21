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
// snapshot. The default mints once at fixture build and serves that stamp;
// snapshotMintedAtServeTime opts into stamping created_at inside the handler,
// which is a legitimate publication while the fetch is in flight — the
// product tolerates its own latency since it sampled its clock
// (BUG-260920-2d9gfv), so only a timestamp ahead of the post-fetch clock
// plus skew reads as tampering.
type registryFixture struct {
	createdAt      time.Time
	beforeSnapshot func()
	serveTimeMint  bool
	futureOffset   time.Duration
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

// snapshotMintedAtServeTime stamps created_at when the handler runs rather
// than when the fixture is built. A real registry may publish between our
// clock read and our fetch completing; that publication is legitimate and
// must not read as tampering (BUG-260920-2d9gfv).
func snapshotMintedAtServeTime() registryOption {
	return func(f *registryFixture) { f.serveTimeMint = true }
}

// snapshotFutureBy stamps created_at the given offset ahead of the serve
// clock (truncated to whole seconds). Unlike snapshotCreatedAt, which is
// anchored at fixture-build time and can slide into the past while test
// setup runs, a serve-relative stamp stays genuinely ahead of the
// post-fetch clock no matter how slow setup was.
func snapshotFutureBy(offset time.Duration) registryOption {
	return func(f *registryFixture) { f.serveTimeMint = true; f.futureOffset = offset }
}

// crossSecondBoundary blocks until 50 ms past the next whole second, so the
// snapshot response is written in a later wall-clock second than the one in
// which the checker sampled its `now`. RFC3339 truncates created_at down to
// that later second, which is how a per-request timestamp lands ahead of a
// `now` that was sampled before the fetch (BUG-260906-1bdotx). The 50 ms
// past the boundary only guards the hook's own scheduling (a thin margin
// would do on an idle runner); the product bound covers the checker's
// latency since it sampled its clock, so the acceptance margin is positive
// by construction (BUG-260920-2d9gfv).
func crossSecondBoundary() {
	now := time.Now()
	time.Sleep(now.Truncate(time.Second).Add(time.Second + 50*time.Millisecond).Sub(now))
}

// fakeRegistry serves signed snapshot and records for one artifact status.
func fakeRegistry(t *testing.T, status, sourceIdentity, commit, contentHash string, options ...registryOption) (*httptest.Server, string) {
	t.Helper()
	public, private, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	// The default stamp is minted once, here; serve-time options re-stamp
	// inside the handler. A serve-time stamp may land in a later whole
	// second than the checker's pre-fetch `now` — that is a legitimate
	// publication during the fetch, not tampering, and the product bound
	// covers it (BUG-260920-2d9gfv).
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
			stamped := createdAt
			if fixture.serveTimeMint {
				stamped = time.Now().UTC().Add(fixture.futureOffset).Truncate(time.Second).Format(time.RFC3339)
			}
			_ = json.NewEncoder(w).Encode(sign(map[string]any{
				"schema_version": 1, "merkle_root": strings.Repeat("a", 64), "log_size": 1, "head": strings.Repeat("b", 64),
				"version": 1, "created_at": stamped,
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

// TestRegistrySnapshotMintedDuringFetchIsNotFuture is the deterministic
// production-entry reproduction for BUG-260920-2d9gfv. resolveRegistries
// samples `now` before it fetches; a registry that publishes while the fetch
// is in flight serves a created_at in a later whole second than that `now`,
// and under a literal zero clock skew any positive difference read as
// tampering. The hook makes that crossing certain instead of leaving it to
// the runner's speed: the stub signs strictly after the next second
// boundary, so created_at is always ahead of the pre-fetch `now` while
// remaining behind the post-fetch clock. It runs through the frozen v1 lane
// (schema-1 Skillfile, install.Project -> resolveRegistries ->
// registry.CheckSnapshotsWithPolicy) with the same zero-skew Go-API config
// shape the crossconformance harness uses, so it is also the legacy-lane
// proof that the fix only removes the false "future" refusal.
func TestRegistrySnapshotMintedDuringFetchIsNotFuture(t *testing.T) {
	t.Parallel()
	e, server := registryEnv(t, "audited", snapshotMintedAtServeTime(), beforeSnapshotResponse(crossSecondBoundary))
	defer server.Close()
	e.cfg.Audit.SnapshotClockSkewSeconds = 0
	result := e.install(Options{})
	if result.Status != "ok" {
		t.Fatalf("a snapshot minted while we were fetching must install: %+v", result)
	}
	if joined := strings.Join(result.Messages, "\n"); strings.Contains(joined, "too far in the future") {
		t.Fatalf("mint-during-fetch was read as a future timestamp: %s", joined)
	}
}

// TestRegistrySnapshotSlowSiblingDoesNotFlipInstantRegistry is the two-registry
// production-entry proof for BUG-260920-2d9gfv (review rev1 §2/F1): the first
// trusted registry is slow — its snapshot response crosses a whole-second
// boundary — and carries no audit record (pending), while the second is
// instant, mints created_at at serve time, and holds the audited record.
// Every timestamp here is at or behind the client's wall clock when checked,
// so the bound (post-fetch clock plus skew) must accept both registries and
// the strict install must succeed via the second. A per-registry start (the
// rev1 shape) refuses the instant registry as "too far in the future"
// whenever the sibling's fetch crossed the boundary; the function-entry
// start accepts it. It runs through the frozen v1 lane with a literal zero
// skew, like the row above.
func TestRegistrySnapshotSlowSiblingDoesNotFlipInstantRegistry(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")
	e.write(e.project, "Skillfile.json", `{
		"schema_version": 1, "agents": ["claude_code"],
		"skills": [{"name": "skill-a", "git": "git@git.example.com:skills/skill-a.git", "tag": "v1"}]
	}`)
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
	slow, slowKey := fakeRegistry(t, "pending", id, ref.Commit, contentHash, beforeSnapshotResponse(crossSecondBoundary))
	defer slow.Close()
	instant, instantKey := fakeRegistry(t, "audited", id, ref.Commit, contentHash, snapshotMintedAtServeTime())
	defer instant.Close()
	e.cfg.AuditRegistries = []config.Registry{
		{Name: "reg-a", URL: slow.URL, PublicKeys: []string{slowKey}, Enabled: true},
		{Name: "reg-b", URL: instant.URL, PublicKeys: []string{instantKey}, Enabled: true},
	}
	e.cfg.Audit.RegistryPolicy = "strict"
	e.cfg.Audit.SnapshotClockSkewSeconds = 0
	result := e.install(Options{})
	if joined := strings.Join(result.Messages, "\n"); strings.Contains(joined, "too far in the future") {
		t.Fatalf("a snapshot at or behind the post-fetch clock was refused as future: %s", joined)
	}
	if result.Status != "ok" {
		t.Fatalf("the audited registry must stay usable: %+v", result)
	}
}

// TestRegistrySnapshotTwoSecondsPastSkewStillRefuses is the negative half of
// the same bound at the production entry: a snapshot genuinely ahead of the
// client's clock after the fetch (serve clock + 2s truncated, i.e. at least
// a second past now + skew with the zero skew used here) must still be
// refused with the tampered-snapshot class, through the same v1 lane as the
// row above. The stamp is serve-relative so slow test setup cannot slide it
// into the past before `now` is sampled. Dropping the future check entirely
// admits this row.
func TestRegistrySnapshotTwoSecondsPastSkewStillRefuses(t *testing.T) {
	t.Parallel()
	e, server := registryEnv(t, "audited", snapshotFutureBy(2*time.Second))
	defer server.Close()
	e.cfg.Audit.SnapshotClockSkewSeconds = 0
	result := e.install(Options{})
	if result.Status != "failed" {
		t.Fatalf("a snapshot 2s past a zero skew must deny the install: %+v", result)
	}
	if !strings.Contains(strings.Join(result.Errors, "\n"), "every trusted audit registry served a tampered snapshot") {
		t.Fatalf("a genuinely future snapshot must exclude the registry: %+v", result.Errors)
	}
	if !strings.Contains(strings.Join(result.Messages, "\n"), "registry test-reg snapshot timestamp is too far in the future") {
		t.Fatalf("the refusal must name the future timestamp: %+v", result.Messages)
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
