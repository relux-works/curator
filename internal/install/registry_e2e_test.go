package install

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/conformancecoverage"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/identity"
	markerpkg "github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/registry"
	"github.com/relux-works/curator/internal/snapshot"
	"github.com/relux-works/curator/internal/stateread"
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
	merkleRoot     string
	logSize        int
	version        int
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

func snapshotView(root string, logSize, version int) registryOption {
	return func(f *registryFixture) {
		f.merkleRoot = root
		f.logSize = logSize
		f.version = version
	}
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
	server, pinned, _ := fakeRegistryWithSigner(t, status, sourceIdentity, commit, contentHash, options...)
	return server, pinned
}

func fakeRegistryWithSigner(t *testing.T, status, sourceIdentity, commit, contentHash string, options ...registryOption) (*httptest.Server, string, ed25519.PrivateKey) {
	return fakeRegistryWithSignerVersion(t, status, sourceIdentity, commit, contentHash, hashing.VersionV1, options...)
}

func fakeRegistryWithSignerVersion(t *testing.T, status, sourceIdentity, commit, contentHash string, hashVersion hashing.Version, options ...registryOption) (*httptest.Server, string, ed25519.PrivateKey) {
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
	fixture := registryFixture{
		createdAt:  time.Now().UTC().Truncate(time.Second),
		merkleRoot: strings.Repeat("a", 64), logSize: 1, version: 1,
	}
	for _, option := range options {
		option(&fixture)
	}
	createdAt := fixture.createdAt.Format(time.RFC3339)
	pinned := "ed25519:" + base64.StdEncoding.EncodeToString(public)
	var pageBoundary map[string]any
	sign := func(body map[string]any) map[string]any {
		signature := ed25519.Sign(private, registry.CanonicalBytes(body))
		body["sig"] = map[string]any{
			"key_id": registry.KeyID(public), "algorithm": "ed25519",
			"signature": base64.StdEncoding.EncodeToString(signature),
		}
		return body
	}
	snapshot := func() map[string]any {
		stamped := createdAt
		if fixture.serveTimeMint {
			stamped = time.Now().UTC().Add(fixture.futureOffset).Truncate(time.Second).Format(time.RFC3339)
		}
		return sign(map[string]any{
			"schema_version": 1, "merkle_root": fixture.merkleRoot, "log_size": fixture.logSize, "head": strings.Repeat("b", 64),
			"version": fixture.version, "created_at": stamped,
		})
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/v1/snapshot"):
			if fixture.beforeSnapshot != nil {
				fixture.beforeSnapshot()
			}
			pageBoundary = snapshot()
			_ = json.NewEncoder(w).Encode(pageBoundary)
		case strings.HasSuffix(r.URL.Path, "/v1/records"):
			if pageBoundary == nil {
				pageBoundary = snapshot()
			}
			record := map[string]any{
				"name": "skill-a", "source_identity": sourceIdentity,
				"commit": commit, "content_sha256": contentHash, "status": status,
			}
			// Version 1 serves the unchanged rc.13 record bytes; only the v2
			// fixture adds the audit-record-v2 version pair.
			if hashVersion == hashing.VersionV2 {
				record["schema_version"] = 2
				record["hash_version"] = 2
			}
			record = sign(record)
			_ = json.NewEncoder(w).Encode(map[string]any{"records": []any{record}, "next_cursor": nil, "boundary": pageBoundary})
		default:
			http.NotFound(w, r)
		}
	}))
	return server, pinned, private
}

func registryEnv(t *testing.T, status string, options ...registryOption) (*env, *httptest.Server) {
	e, server, _, _ := registryEnvWithSigner(t, status, options...)
	return e, server
}

func registryEnvWithSigner(t *testing.T, status string, options ...registryOption) (*env, *httptest.Server, ed25519.PrivateKey, string) {
	return registryEnvWithVersions(t, status, hashing.VersionV1, hashing.VersionV1, options...)
}

// These fixtures select the package writer switch and therefore run serially;
// parallel install tests start after their writer selection has been restored.
func registryEnvWithVersions(t *testing.T, status string, artifactVersion, recordVersion hashing.Version, options ...registryOption) (*env, *httptest.Server, ed25519.PrivateKey, string) {
	t.Helper()
	prior := hashing.EnableV2Writers
	hashing.EnableV2Writers = artifactVersion == hashing.VersionV2
	t.Cleanup(func() { hashing.EnableV2Writers = prior })
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
	contentHash, err := hashing.ContentSHA256WithVersion(snap, nil, artifactVersion)
	if err != nil {
		t.Fatal(err)
	}
	id := identity.Canonical("git@git.example.com:skills/skill-a.git")
	server, pinned, private := fakeRegistryWithSignerVersion(t, status, id, ref.Commit, contentHash, recordVersion, options...)
	e.cfg.AuditRegistries = []config.Registry{{Name: "test-reg", URL: server.URL, PublicKeys: []string{pinned}, Enabled: true}}
	return e, server, private, pinned
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

func TestV1InstallRejectsV2RegistryEvidence(t *testing.T) {
	e, server, _, _ := registryEnvWithVersions(t, "audited", hashing.VersionV1, hashing.VersionV2)
	defer server.Close()
	e.cfg.Audit.RegistryPolicy = "strict"
	result := e.install(Options{})
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, "\n"), "registry_policy is strict") {
		t.Fatalf("v2 evidence must not authorize a v1 artifact: %+v", result)
	}
}

func TestV2RegistryAttestationUsesVersionedInstallPath(t *testing.T) {
	e, server, _, _ := registryEnvWithVersions(t, "audited", hashing.VersionV2, hashing.VersionV2)
	defer server.Close()
	result := e.install(Options{})
	if result.Status != "ok" {
		t.Fatalf("matching v2 registry evidence must install: %+v", result)
	}
	recorded := readMarkerFor(t, e, "skill-a")
	if recorded.SchemaVersion != markerpkg.SchemaV5 || recorded.HashVersion != hashing.VersionV2 ||
		recorded.Attestation == nil || recorded.Attestation.Status != "audited" {
		t.Fatalf("v2 marker did not retain matching registry evidence: %+v", recorded)
	}
}

func TestRegistryFirstUseReportsAndPersistsTOFUPosture(t *testing.T) {
	e, server := registryEnv(t, "audited")
	defer server.Close()
	result := e.install(Options{})
	if result.Status != "ok" {
		t.Fatalf("install: %+v", result)
	}
	joined := strings.Join(result.Messages, "\n")
	if strings.Count(joined, "registry_bootstrap_tofu") != 1 || !strings.Contains(joined, server.URL) || !strings.Contains(joined, "bootstrap_checkpoint") {
		t.Fatalf("first-use warning = %s, want one URL-naming checkpoint hint", joined)
	}
	state := installedRegistryState(t, e.home, server.URL)
	if state["bootstrap_source"] != "first-use" {
		t.Fatalf("first-use state = %+v", state)
	}
}

func TestRegistryCheckpointPersistsBeforeTamperedNetworkSnapshot(t *testing.T) {
	root := strings.Repeat("a", 64)
	e, server, private, pinned := registryEnvWithSigner(t, "audited", snapshotView(strings.Repeat("c", 64), 7, 7))
	defer server.Close()
	e.cfg.AuditRegistries[0].BootstrapCheckpoint = writeBootstrapCheckpoint(t, e.home, private, pinned, root, 8, 8, true)

	result := e.install(Options{})
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Messages, "\n"), "snapshot version moved backward") {
		t.Fatalf("below-checkpoint network view must be refused: %+v", result)
	}
	state := installedRegistryState(t, e.home, server.URL)
	if state["highest_version"] != float64(8) || state["merkle_root"] != root || state["bootstrap_source"] != "checkpoint" {
		t.Fatalf("checkpoint high-water was not retained before network refusal: %+v", state)
	}
}

func TestRegistryInvalidFirstUseCheckpointFailsClosed(t *testing.T) {
	e, server, private, pinned := registryEnvWithSigner(t, "audited")
	defer server.Close()
	e.cfg.AuditRegistries[0].BootstrapCheckpoint = writeBootstrapCheckpoint(t, e.home, private, pinned, strings.Repeat("a", 64), 8, 8, false)

	result := e.install(Options{})
	joined := strings.Join(append(append([]string(nil), result.Errors...), result.Messages...), "\n")
	if result.Status != "failed" || !strings.Contains(joined, "bootstrap checkpoint configuration error") || !strings.Contains(joined, "signature failed verification") {
		t.Fatalf("invalid first-use checkpoint was not refused with its path: %+v", result)
	}
	sum := sha256.Sum256([]byte(server.URL))
	statePath := filepath.Join(e.home, "state", "registry", "snapshot-"+hex.EncodeToString(sum[:])[:16]+".json")
	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Fatalf("invalid first-use checkpoint persisted registry high-water: %v", err)
	}
}

func TestRegistryRebootstrapRegressionKeepsExistingHighWater(t *testing.T) {
	e, server, private, pinned := registryEnvWithSigner(t, "audited", snapshotView(strings.Repeat("a", 64), 1, 1))
	defer server.Close()
	first := e.install(Options{})
	if first.Status != "ok" {
		t.Fatalf("initial install: %+v", first)
	}
	prior := installedRegistryState(t, e.home, server.URL)
	e.cfg.AuditRegistries[0].BootstrapCheckpoint = writeBootstrapCheckpoint(t, e.home, private, pinned, strings.Repeat("c", 64), 1, 1, true)

	result := e.install(Options{})
	joined := strings.Join(result.Messages, "\n")
	if result.Status != "ok" || !strings.Contains(joined, "registry_checkpoint_regression (error)") {
		t.Fatalf("regressing rebootstrap should retain trusted state and report an error: %+v", result)
	}
	after := installedRegistryState(t, e.home, server.URL)
	if after["highest_version"] != prior["highest_version"] || after["head"] != prior["head"] || after["merkle_root"] != prior["merkle_root"] || after["bootstrap_source"] != "first-use" {
		t.Fatalf("regressing checkpoint changed high-water: before=%+v after=%+v", prior, after)
	}
}

func writeBootstrapCheckpoint(t *testing.T, directory string, private ed25519.PrivateKey, pinned, root string, logSize, version int, validSignature bool) string {
	t.Helper()
	public, err := registry.ParsePublicKey(pinned)
	if err != nil {
		t.Fatal(err)
	}
	body := map[string]any{
		"schema_version": 1, "merkle_root": root, "log_size": logSize,
		"head": strings.Repeat("b", 64), "version": version,
		"created_at": time.Now().UTC().Truncate(time.Second).Format(time.RFC3339),
	}
	signature := ed25519.Sign(private, registry.CanonicalBytes(body))
	if !validSignature {
		signature = make([]byte, ed25519.SignatureSize)
	}
	body["sig"] = map[string]any{
		"key_id": registry.KeyID(public), "algorithm": "ed25519",
		"signature": base64.StdEncoding.EncodeToString(signature),
	}
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	checkpointDir := filepath.Join(directory, "registry-checkpoints")
	if err := os.Mkdir(checkpointDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := pathboundary.ProtectTree(checkpointDir); err != nil {
		t.Fatalf("protect registry checkpoint fixture directory: %v", err)
	}
	path := filepath.Join(checkpointDir, "registry-checkpoint.json")
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := pathboundary.ProtectTree(checkpointDir); err != nil {
		t.Fatalf("protect registry checkpoint fixture file: %v", err)
	}
	return path
}

func installedRegistryState(t *testing.T, home, registryURL string) map[string]any {
	t.Helper()
	path := filepath.Join(home, "state", "registry", "snapshot-"+registryStateDigest(registryURL)+".json")
	payload, err := os.ReadFile(path) // #nosec G304 -- test-owned manager state path
	if err != nil {
		t.Fatal(err)
	}
	var state map[string]any
	if err := json.Unmarshal(payload, &state); err != nil {
		t.Fatal(err)
	}
	return state
}

type registryClientBootstrapVector struct {
	Name                 string  `json:"name"`
	Phase                string  `json:"phase"`
	PriorState           string  `json:"prior_state"`
	CheckpointConfigured bool    `json:"checkpoint_configured"`
	CheckpointVersion    int     `json:"checkpoint_version"`
	CandidateSameBody    bool    `json:"candidate_same_body"`
	FirstNetworkVersion  int     `json:"first_network_version"`
	Accepted             bool    `json:"accepted"`
	StateChanged         bool    `json:"state_changed"`
	StoredVersion        int     `json:"stored_version"`
	Diagnostic           *string `json:"diagnostic"`
	Posture              *string `json:"posture"`
	CheckCurrent         bool    `json:"check_current"`
	SignatureValid       bool    `json:"signature_valid"`
	RegistryExcluded     bool    `json:"registry_excluded"`
	ResolutionChanged    bool    `json:"resolution_changed"`
	Policy               string  `json:"policy"`
}

// TestRegistryBootstrapVectorsDriveInstall binds every pinned bootstrap and
// rebootstrap vector to the production install.Project -> resolveRegistries
// path. The test supplies signed registry responses and checkpoints while the
// pinned vectors select trust state, versions, and expected posture.
func TestRegistryBootstrapVectorsDriveInstall(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "registry-client.json")) // #nosec G304 -- explicit pinned conformance input
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		BootstrapCases []registryClientBootstrapVector `json:"bootstrap_cases"`
	}
	if err := json.Unmarshal(payload, &document); err != nil {
		t.Fatal(err)
	}
	var cases []registryClientBootstrapVector
	for _, vector := range document.BootstrapCases {
		if vector.Phase != "compare" {
			cases = append(cases, vector)
		}
	}
	if len(cases) != 10 {
		t.Fatalf("pinned registry-client bootstrap vectors = %d, want 10", len(cases))
	}
	conformancecoverage.RunOutcomes(t, "registry-client/bootstrap-cases", cases,
		func(vector registryClientBootstrapVector) string { return vector.Name },
		func(caseT *testing.T, vector registryClientBootstrapVector) conformancecoverage.Observation {
			return runRegistryClientBootstrapVector(caseT, vector)
		})
}

func runRegistryClientBootstrapVector(t *testing.T, vector registryClientBootstrapVector) conformancecoverage.Observation {
	t.Helper()
	priorRoot := strings.Repeat("a", 64)
	checkpointRoot := priorRoot
	networkRoot := priorRoot
	networkVersion := vector.FirstNetworkVersion
	if vector.Phase == "rebootstrap" {
		networkVersion = vector.StoredVersion
		if vector.CheckpointConfigured && vector.SignatureValid && vector.CheckpointVersion > vector.StoredVersion {
			networkVersion = vector.CheckpointVersion
		}
	}
	if vector.Phase == "rebootstrap" && vector.CheckpointConfigured && !vector.CandidateSameBody {
		checkpointRoot = strings.Repeat("c", 64)
	}
	if vector.Phase == "bootstrap" && vector.CheckpointConfigured && networkVersion == vector.CheckpointVersion && !vector.CandidateSameBody {
		networkRoot = strings.Repeat("c", 64)
	}
	if vector.Phase == "bootstrap" && vector.CheckpointConfigured && networkVersion > vector.CheckpointVersion {
		networkRoot = strings.Repeat("c", 64)
	}
	if vector.Phase == "rebootstrap" && vector.CheckpointConfigured && vector.SignatureValid && vector.CheckpointVersion > vector.StoredVersion {
		networkRoot = checkpointRoot
	}

	e, server, private, pinned := registryEnvWithSigner(t, "audited", snapshotView(networkRoot, networkVersion, networkVersion))
	defer server.Close()
	if vector.Policy != "" {
		e.cfg.Audit.RegistryPolicy = vector.Policy
	}
	if vector.PriorState == "present" {
		seedRegistryHighWater(t, e.home, []string{server.URL}, []string{priorRoot})
	}
	var before map[string]any
	if vector.PriorState == "present" {
		before = installedRegistryState(t, e.home, server.URL)
	}
	if vector.CheckpointConfigured {
		e.cfg.AuditRegistries[0].BootstrapCheckpoint = writeBootstrapCheckpoint(
			t, e.home, private, pinned, checkpointRoot,
			vector.CheckpointVersion, vector.CheckpointVersion, vector.SignatureValid,
		)
	}

	result := e.install(Options{})
	joined := strings.Join(append(append([]string(nil), result.Errors...), result.Messages...), "\n")
	if vector.Phase == "bootstrap" && (result.Status == "ok") != vector.Accepted {
		return conformancecoverage.Observation{FailureReason: fmt.Sprintf("bootstrap install status=%s, accepted=%v: %s", result.Status, vector.Accepted, joined)}
	}
	if vector.Phase == "bootstrap" && (result.Status == "failed") != vector.RegistryExcluded {
		return conformancecoverage.Observation{FailureReason: fmt.Sprintf("bootstrap registry exclusion disagrees with vector: status=%s excluded=%v: %s", result.Status, vector.RegistryExcluded, joined)}
	}
	if vector.Phase == "rebootstrap" && result.Status != "ok" {
		return conformancecoverage.Observation{FailureReason: fmt.Sprintf("rebootstrap changed registry resolution: %+v", result)}
	}
	if vector.ResolutionChanged {
		return conformancecoverage.Observation{FailureReason: "pinned bootstrap vectors must not change artifact resolution"}
	}
	if vector.Diagnostic != nil && !strings.Contains(joined, *vector.Diagnostic) {
		return conformancecoverage.Observation{FailureReason: fmt.Sprintf("install output omitted %q: %s", *vector.Diagnostic, joined)}
	}
	if vector.Posture != nil && !strings.Contains(joined, *vector.Posture) {
		return conformancecoverage.Observation{FailureReason: fmt.Sprintf("install output omitted posture %q: %s", *vector.Posture, joined)}
	}
	if result.Status == "ok" {
		marker := readMarkerFor(t, e, "skill-a")
		if marker.Attestation == nil || marker.Attestation.Status != "audited" {
			return conformancecoverage.Observation{FailureReason: "bootstrap handling changed the audited resolution"}
		}
	}

	statePath := filepath.Join(e.home, "state", "registry")
	posture := registry.ReadBoundaryPostureWithPolicy(statePath, []registry.Registry{{
		Name: "test-reg", URL: server.URL, PublicKeys: []string{pinned},
		BootstrapCheckpoint: e.cfg.AuditRegistries[0].BootstrapCheckpoint,
	}}, vector.Policy)
	if len(posture) != 1 {
		return conformancecoverage.Observation{FailureReason: fmt.Sprintf("status rows = %d, want one", len(posture))}
	}
	row := posture[0]
	if (row.Diagnostic == "") != vector.CheckCurrent {
		return conformancecoverage.Observation{FailureReason: fmt.Sprintf("status current=%v, want %v: %+v", row.Diagnostic == "", vector.CheckCurrent, row)}
	}
	if vector.Diagnostic != nil && row.BootstrapDiagnostic != *vector.Diagnostic {
		return conformancecoverage.Observation{FailureReason: fmt.Sprintf("status bootstrap diagnostic=%q, want %q: %+v", row.BootstrapDiagnostic, *vector.Diagnostic, row)}
	}
	if vector.Posture != nil && row.BootstrapDiagnostic != *vector.Posture {
		return conformancecoverage.Observation{FailureReason: fmt.Sprintf("status bootstrap posture=%q, want %q: %+v", row.BootstrapDiagnostic, *vector.Posture, row)}
	}

	var after map[string]any
	stateFile := filepath.Join(statePath, "snapshot-"+registryStateDigest(server.URL)+".json")
	metadata, err := stateread.Lstat(stateFile)
	if err != nil {
		return conformancecoverage.Observation{FailureReason: fmt.Sprintf("registry state could not be inspected: %v", err)}
	}
	if metadata.Kind == stateread.KindPresent {
		after = installedRegistryState(t, e.home, server.URL)
	} else if metadata.Kind != stateread.KindAbsent {
		return conformancecoverage.Observation{FailureReason: "registry state inspection was unreadable"}
	}
	stateChanged := false
	if before == nil {
		stateChanged = after != nil
	} else if after != nil {
		stateChanged = before["highest_version"] != after["highest_version"] || before["head"] != after["head"] ||
			before["merkle_root"] != after["merkle_root"] || before["log_size"] != after["log_size"] ||
			before["bootstrap_source"] != after["bootstrap_source"]
	}
	if stateChanged != vector.StateChanged {
		return conformancecoverage.Observation{FailureReason: fmt.Sprintf("state changed=%v, want %v: before=%+v after=%+v", stateChanged, vector.StateChanged, before, after)}
	}
	if before != nil && before["highest_version"] != float64(vector.StoredVersion) {
		return conformancecoverage.Observation{FailureReason: fmt.Sprintf("stored version=%v, want %d", before["highest_version"], vector.StoredVersion)}
	}
	return conformancecoverage.Observation{}
}

func registryStateDigest(registryURL string) string {
	sum := sha256.Sum256([]byte(registryURL))
	return hex.EncodeToString(sum[:])[:16]
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

type registryClientMirrorVector struct {
	Name              string  `json:"name"`
	GroupSize         int     `json:"group_size"`
	Policy            string  `json:"policy"`
	RootsEqual        bool    `json:"roots_equal"`
	SameLogSize       bool    `json:"same_log_size"`
	Compared          bool    `json:"compared"`
	CheckCurrent      bool    `json:"check_current"`
	Diagnostic        *string `json:"diagnostic"`
	Severity          *string `json:"severity"`
	Accepted          bool    `json:"accepted"`
	RegistryExcluded  bool    `json:"registry_excluded"`
	ResolutionChanged bool    `json:"resolution_changed"`
}

// TestRegistryMirrorComparisonVectorsDriveInstall binds every published
// mirror-comparison vector to the real install.Project -> resolveRegistries
// path. The vectors decide group size, root equality, boundary equality,
// policy, and expected diagnostic; this test supplies signed registry views.
func TestRegistryMirrorComparisonVectorsDriveInstall(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "registry-client.json")) // #nosec G304 -- explicit pinned conformance input
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		BootstrapCases []registryClientMirrorVector `json:"bootstrap_cases"`
	}
	if err := json.Unmarshal(payload, &document); err != nil {
		t.Fatal(err)
	}
	var cases []registryClientMirrorVector
	for _, vector := range document.BootstrapCases {
		if strings.HasPrefix(vector.Name, "divergence-") {
			cases = append(cases, vector)
		}
	}
	if len(cases) != 5 {
		t.Fatalf("pinned registry-client mirror vectors = %d, want 5", len(cases))
	}
	conformancecoverage.RunOutcomes(t, "registry-client/mirror-comparison-cases", cases,
		func(vector registryClientMirrorVector) string { return vector.Name },
		func(caseT *testing.T, vector registryClientMirrorVector) conformancecoverage.Observation {
			return runRegistryClientMirrorVector(caseT, vector)
		})
}

func runRegistryClientMirrorVector(t *testing.T, vector registryClientMirrorVector) conformancecoverage.Observation {
	t.Helper()
	rootA := strings.Repeat("a", 64)
	rootB := rootA
	if !vector.RootsEqual {
		rootB = strings.Repeat("c", 64)
	}
	firstSize, secondSize := 8, 8
	firstVersion, secondVersion := 8, 8
	if !vector.SameLogSize {
		secondSize, secondVersion = 9, 9
	}
	e, firstServer := registryEnv(t, "audited", snapshotView(rootA, firstSize, firstVersion))
	defer firstServer.Close()
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
	first := e.cfg.AuditRegistries[0]
	first.MirrorGroup = "prod"
	registries := []config.Registry{first}
	urls := []string{first.URL}
	roots := []string{rootA}
	if vector.GroupSize == 2 {
		secondServer, secondKey := fakeRegistry(t, "audited", id, ref.Commit, contentHash,
			snapshotView(rootB, secondSize, secondVersion))
		defer secondServer.Close()
		registries = append(registries, config.Registry{
			Name: "mirror", URL: secondServer.URL, PublicKeys: []string{secondKey}, Enabled: true, MirrorGroup: "prod",
		})
		urls = append(urls, secondServer.URL)
		roots = append(roots, rootB)
	}
	e.cfg.AuditRegistries = registries
	e.cfg.Audit.RegistryPolicy = vector.Policy
	seedRegistryHighWater(t, e.home, urls, roots)

	result := e.install(Options{})
	if !vector.Accepted || vector.RegistryExcluded || vector.ResolutionChanged {
		return conformancecoverage.Observation{FailureReason: "published comparison vector is outside the report-only accepted-view cases"}
	}
	if result.Status != "ok" {
		return conformancecoverage.Observation{FailureReason: fmt.Sprintf("install was changed by report-only comparison: %+v", result)}
	}
	marker := readMarkerFor(t, e, "skill-a")
	if marker.Attestation == nil || marker.Attestation.Status != "audited" {
		return conformancecoverage.Observation{FailureReason: "mirror comparison changed the audited registry resolution"}
	}
	joined := strings.Join(result.Messages, "\n")
	if vector.Diagnostic == nil {
		if strings.Contains(joined, "registry_view_divergence") {
			return conformancecoverage.Observation{FailureReason: "a non-divergent or incomparable vector emitted registry_view_divergence"}
		}
	} else {
		if !strings.Contains(joined, *vector.Diagnostic) {
			return conformancecoverage.Observation{FailureReason: fmt.Sprintf("install output omitted %q: %s", *vector.Diagnostic, joined)}
		}
		if vector.Severity != nil && !strings.Contains(joined, "("+*vector.Severity+")") {
			return conformancecoverage.Observation{FailureReason: fmt.Sprintf("diagnostic severity did not match %q: %s", *vector.Severity, joined)}
		}
		if !strings.Contains(joined, "mirror group prod") || !strings.Contains(joined, "test-reg") || !strings.Contains(joined, "mirror") {
			return conformancecoverage.Observation{FailureReason: fmt.Sprintf("diagnostic omitted group or disagreeing registries: %s", joined)}
		}
	}
	postureRegistries := make([]registry.Registry, 0, len(registries))
	for _, configured := range registries {
		postureRegistries = append(postureRegistries, registry.Registry{
			Name: configured.Name, URL: configured.URL, PublicKeys: configured.PublicKeys,
			MirrorGroup: configured.MirrorGroup,
		})
	}
	posture := registry.ReadBoundaryPostureWithPolicy(filepath.Join(e.home, "state", "registry"), postureRegistries, vector.Policy)
	wantComparison := "not-compared"
	if vector.Compared {
		if vector.RootsEqual {
			wantComparison = "agree"
		} else {
			wantComparison = "diverged"
		}
	}
	for _, row := range posture {
		if row.LastMirrorComparison != wantComparison || row.MirrorGroup != "prod" {
			return conformancecoverage.Observation{FailureReason: fmt.Sprintf("persisted mirror posture = %+v, want %s for prod", row, wantComparison)}
		}
		rowCurrent := row.Diagnostic == ""
		if rowCurrent != vector.CheckCurrent {
			return conformancecoverage.Observation{FailureReason: fmt.Sprintf("status current=%v, want %v: %+v", row.Diagnostic == "", vector.CheckCurrent, row)}
		}
		if row.BootstrapSource != "first-use" || row.BootstrapDiagnostic != "registry_bootstrap_tofu" {
			return conformancecoverage.Observation{FailureReason: fmt.Sprintf("bootstrap posture = %+v, want first-use TOFU", row)}
		}
	}
	return conformancecoverage.Observation{}
}

func seedRegistryHighWater(t *testing.T, home string, urls, roots []string) {
	t.Helper()
	stateDir := filepath.Join(home, "state", "registry")
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		t.Fatal(err)
	}
	states := make([]string, 0, len(urls))
	for index, registryURL := range urls {
		sum := sha256.Sum256([]byte(registryURL))
		name := "snapshot-" + hex.EncodeToString(sum[:])[:16] + ".json"
		states = append(states, name)
		state, err := json.Marshal(map[string]any{
			"highest_version":  8,
			"head":             strings.Repeat("b", 64),
			"merkle_root":      roots[index],
			"log_size":         8,
			"bootstrap_source": "first-use",
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(stateDir, name), state, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	catalog, err := json.Marshal(map[string]any{"schema_version": 1, "states": states})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "known-registries.json"), catalog, 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestRegistrySnapshotSurvivesASecondBoundaryDuringFetch is the regression
// test for BUG-260906-1bdotx. resolveRegistries samples `now` before it
// fetches; a fixture that stamped created_at inside the handler produced a
// timestamp in a later whole second than that `now` whenever the fetch crossed
// a second boundary, and the e2e config carries a literal zero clock skew, so
// any positive difference read as tampering. The hook makes that crossing
// certain instead of leaving it to the runner's speed.
func TestRegistrySnapshotSurvivesASecondBoundaryDuringFetch(t *testing.T) {
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
	// Frozen v1 lane, like the rows above: the fixtures serve rc.13 audit
	// records keyed by v1 content hashes, and versioned resolution refuses
	// cross-version records. This test selects the package writer switch
	// and therefore runs serially; parallel install tests start after its
	// writer selection has been restored.
	priorWriter := hashing.EnableV2Writers
	hashing.EnableV2Writers = false
	t.Cleanup(func() { hashing.EnableV2Writers = priorWriter })
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
	e, server := registryEnv(t, "audited", snapshotCreatedAt(time.Now().Add(-time.Minute)))
	defer server.Close()
	e.cfg.Audit.SnapshotClockSkewSeconds = 0
	result := e.install(Options{})
	if result.Status != "ok" {
		t.Fatalf("a past-dated snapshot must install under a literal zero skew: %+v", result)
	}
}
