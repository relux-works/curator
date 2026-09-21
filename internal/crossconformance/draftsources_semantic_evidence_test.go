package crossconformance

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/registry"
	"github.com/relux-works/curator/internal/sourcelock"
)

// Attestation-evidence semantic rows (skillfile-sources §4): a missing,
// unreadable, malformed, stale, revoked, or mismatching required
// evidence record must refuse install/repair/refresh and report status
// noncurrent, preserving prior state. Each driven row works in layers:
//
//  0. registry.Resolve with a static per-condition fetch (every OS).
//  1. install.Project over a real httptest registry (every OS): fresh
//     install refuses, repair install refuses, prior state preserved.
//  2. closure.RefreshDraft (every OS): resolution does not consult
//     registry evidence, so refresh succeeds but touches no
//     install-gated state, and the install after it still refuses —
//     refresh cannot launder revoked evidence.
//  3. Compiled-CLI status (unix): status is nonzero and read-only.
//
// wrong-name and wrong-context are driven through the exact draft §4
// entry: frozen §13.3 OR-matching (content, or identity+commit; name
// uncompared) still accepts them as audited behind registry.Resolve,
// which the legacy lane keeps byte-identically, while the draft lane
// resolves through registry.ResolveExact and refuses both shapes.

type attestCondition int

const (
	attestGood attestCondition = iota
	attestAbsent
	attestUnreadable
	attestMalformed
	attestStale
	attestRevoked
	attestWrongRepository
	attestWrongCommit
	attestWrongKey
	attestWrongName
	attestWrongContext
)

type stubSigner struct {
	private ed25519.PrivateKey
	pinned  string
}

func newStubSigner(t *testing.T) *stubSigner {
	t.Helper()
	public, private, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &stubSigner{private: private, pinned: "ed25519:" + base64.StdEncoding.EncodeToString(public)}
}

func (s *stubSigner) sign(body map[string]any) map[string]any {
	record := map[string]any{}
	for key, value := range body {
		if key != "sig" {
			record[key] = value
		}
	}
	signature := ed25519.Sign(s.private, registry.CanonicalBytes(record))
	public, err := registry.ParsePublicKey(s.pinned)
	if err != nil {
		panic(err)
	}
	record["sig"] = map[string]any{
		"key_id":    registry.KeyID(public),
		"algorithm": "ed25519",
		"signature": base64.StdEncoding.EncodeToString(signature),
	}
	return record
}

// attestStub serves one signed registry over httptest with a
// programmable evidence condition. Records are minted from the query
// keys, exactly as a real registry answers.
type attestStub struct {
	t         *testing.T
	server    *httptest.Server
	URL       string
	pinned    string
	good      *stubSigner
	rogue     *stubSigner
	mu        sync.Mutex
	condition attestCondition
}

func newAttestStub(t *testing.T) *attestStub {
	t.Helper()
	stub := &attestStub{t: t, good: newStubSigner(t), rogue: newStubSigner(t), condition: attestGood}
	stub.pinned = stub.good.pinned
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/snapshot", stub.serveSnapshot)
	mux.HandleFunc("/v1/records", stub.serveRecords)
	stub.server = httptest.NewServer(mux)
	t.Cleanup(stub.server.Close)
	stub.URL = stub.server.URL
	return stub
}

func (s *attestStub) set(condition attestCondition) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.condition = condition
}

func (s *attestStub) get() attestCondition {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.condition
}

func (s *attestStub) serveSnapshot(w http.ResponseWriter, _ *http.Request) {
	created := time.Now().UTC().Truncate(time.Second)
	if s.get() == attestStale {
		created = created.Add(-8 * 24 * time.Hour)
	}
	body := map[string]any{
		"schema_version": 1,
		"version":        1,
		"log_size":       0,
		"head":           strings.Repeat("ab", 32),
		"merkle_root":    strings.Repeat("cd", 32),
		"created_at":     created.Format("2006-01-02T15:04:05Z"),
	}
	writeStubJSON(w, s.good.sign(body))
}

func (s *attestStub) serveRecords(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	identity, commit, content := query.Get("source_identity"), query.Get("commit"), query.Get("content_sha256")
	condition := s.get()
	if condition == attestUnreadable {
		http.Error(w, "fixture outage", http.StatusInternalServerError)
		return
	}
	var records []map[string]any
	mint := func(mut func(map[string]any), signer *stubSigner) {
		body := map[string]any{
			"name": "review", "source_identity": identity, "commit": commit, "content_sha256": content,
			"status": registry.StatusAudited, "audit": map[string]any{"auditor": "team"},
		}
		if mut != nil {
			mut(body)
		}
		records = append(records, signer.sign(body))
	}
	switch condition {
	case attestAbsent:
		// No records.
	case attestMalformed:
		records = append(records, map[string]any{"name": "review", "status": "bogus"})
	case attestRevoked:
		mint(func(body map[string]any) { body["status"] = registry.StatusRevoked }, s.good)
	case attestWrongRepository:
		mint(func(body map[string]any) {
			body["source_identity"] = "unrelated.test/kit"
			body["commit"] = strings.Repeat("b", 40)
			body["content_sha256"] = "sha256:" + strings.Repeat("e", 64)
		}, s.good)
	case attestWrongCommit:
		mint(func(body map[string]any) {
			body["commit"] = strings.Repeat("b", 40)
			body["content_sha256"] = "sha256:" + strings.Repeat("e", 64)
		}, s.good)
	case attestWrongKey:
		mint(nil, s.rogue)
	case attestWrongName:
		mint(func(body map[string]any) { body["name"] = "other" }, s.good)
	case attestWrongContext:
		mint(func(body map[string]any) {
			body["content_sha256"] = "sha256:" + strings.Repeat("e", 64)
		}, s.good)
	default:
		mint(nil, s.good)
	}
	writeStubJSON(w, map[string]any{"records": records, "next_cursor": nil})
}

func writeStubJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	payload, err := json.Marshal(value)
	if err != nil {
		http.Error(w, "fixture encode", http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(payload)
}

func (s *attestStub) registries() []registry.Registry {
	return []registry.Registry{{Name: "one", URL: s.URL, PublicKeys: []string{s.pinned}}}
}

func (s *attestStub) configRegistries() []config.Registry {
	return []config.Registry{{Name: "one", URL: s.URL, PublicKeys: []string{s.pinned}, Enabled: true}}
}

// draftGitProject resolves one Git-selected review package through the
// production closure and writes its lock plus machine bindings,
// returning the project, home, and fixture repository.
func draftGitProject(t *testing.T, gitURL string) (project, home, repo string) {
	t.Helper()
	payload := `{"schema_version":2,"sources":{"s":{"git":"` + gitURL + `","tag":"v1"}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
	project, home = draftProject(t, payload, nil)
	repo = t.TempDir()
	writeDraftSkill(t, filepath.Join(repo, "skills", "review"), "review")
	runDraftGit(t, repo, "init", "-q", "-b", "main")
	runDraftGit(t, repo, "add", ".")
	runDraftGit(t, repo, "commit", "-qm", "fixture")
	runDraftGit(t, repo, "tag", "v1")
	m := draftManifest(t, project, payload)
	plan, err := closure.ResolveDraft(closure.DraftResolveConfig{
		ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: []byte(payload),
		Expansion: manifest.ExpansionOptions{GitRoots: map[string]string{"s": repo}},
	})
	if err != nil {
		t.Fatalf("ResolveDraft: %v", err)
	}
	if err := sourcelock.Write(sourcelock.PathIn(project), plan.Lock); err != nil {
		t.Fatalf("write lock: %v", err)
	}
	bindings, err := sourcelock.NewBindings(plan.Lock.LockSHA256, map[string]sourcelock.SourceBinding{"s": {Location: repo}})
	if err != nil {
		t.Fatal(err)
	}
	if err := sourcelock.WriteBindings(install.DraftBindingsPath(home, project), bindings); err != nil {
		t.Fatal(err)
	}
	return project, home, repo
}

func draftStrictRegistryConfig(home string, stub *attestStub) *config.Config {
	cfg := draftInstallConfig(home)
	cfg.Audit.RegistryPolicy = "strict"
	cfg.AuditRegistries = stub.configRegistries()
	cfg.Audit.CacheTTLSeconds = 0
	cfg.Audit.OfflineGraceSeconds = 0
	return cfg
}

var draftEvidenceConditions = map[string]attestCondition{
	"attestation-evidence-absent":           attestAbsent,
	"attestation-evidence-unreadable":       attestUnreadable,
	"attestation-evidence-malformed":        attestMalformed,
	"attestation-evidence-stale":            attestStale,
	"attestation-evidence-revoked":          attestRevoked,
	"attestation-evidence-wrong-repository": attestWrongRepository,
	"attestation-evidence-wrong-commit":     attestWrongCommit,
	"attestation-evidence-wrong-key":        attestWrongKey,
	"attestation-evidence-wrong-name":       attestWrongName,
	"attestation-evidence-wrong-context":    attestWrongContext,
}

func init() {
	for id, condition := range draftEvidenceConditions {
		id, condition := id, condition
		registerDraftSemantic(id, func(t *testing.T, c draftSemanticCase) { driveAttestationEvidence(t, c, condition) })
	}
}

func driveAttestationEvidence(t *testing.T, c draftSemanticCase, condition attestCondition) {
	stub := newAttestStub(t)
	// Layer 0: the production resolution entry over a static fetch.
	regs := stub.registries()
	identity, commit, content := "example.org/kit", strings.Repeat("a", 40), "sha256:"+strings.Repeat("d", 64)
	static := func(payloads []map[string]any, err error) registry.FetchFn {
		return func(_, _, _, _ string) ([]map[string]any, error) { return payloads, err }
	}
	mintStatic := func(mut func(map[string]any), signer *stubSigner) map[string]any {
		body := map[string]any{"name": "review", "source_identity": identity, "commit": commit, "content_sha256": content,
			"status": registry.StatusAudited, "audit": map[string]any{"auditor": "team"}}
		if mut != nil {
			mut(body)
		}
		return signer.sign(body)
	}
	var resolution registry.Resolution
	switch condition {
	case attestAbsent:
		resolution = registry.Resolve(regs, identity, commit, content, static(nil, nil))
	case attestUnreadable:
		resolution = registry.Resolve(regs, identity, commit, content, static(nil, fmt.Errorf("fixture outage")))
	case attestMalformed:
		resolution = registry.Resolve(regs, identity, commit, content, static([]map[string]any{{"name": "review", "status": "bogus"}}, nil))
	case attestStale:
		tampered, _ := registry.CheckSnapshotsWithPolicy(regs, t.TempDir(), func(_ string) (map[string]any, error) {
			return stub.good.sign(map[string]any{"schema_version": 1, "version": 1, "log_size": 0,
				"head": strings.Repeat("ab", 32), "merkle_root": strings.Repeat("cd", 32),
				"created_at": time.Now().UTC().Add(-8 * 24 * time.Hour).Truncate(time.Second).Format("2006-01-02T15:04:05Z")}), nil
		}, time.Now(), 0, 0)
		if !tampered[stub.URL] {
			t.Fatal("stale snapshot not excluded")
		}
		resolution = registry.Resolution{Result: registry.ResultUnknown}
	case attestRevoked:
		resolution = registry.Resolve(regs, identity, commit, content, static([]map[string]any{mintStatic(func(body map[string]any) {
			body["status"] = registry.StatusRevoked
		}, stub.good)}, nil))
	case attestWrongRepository:
		resolution = registry.Resolve(regs, identity, commit, content, static([]map[string]any{mintStatic(func(body map[string]any) {
			body["source_identity"] = "unrelated.test/kit"
			body["commit"] = strings.Repeat("b", 40)
			body["content_sha256"] = "sha256:" + strings.Repeat("e", 64)
		}, stub.good)}, nil))
	case attestWrongCommit:
		resolution = registry.Resolve(regs, identity, commit, content, static([]map[string]any{mintStatic(func(body map[string]any) {
			body["commit"] = strings.Repeat("b", 40)
			body["content_sha256"] = "sha256:" + strings.Repeat("e", 64)
		}, stub.good)}, nil))
	case attestWrongKey:
		resolution = registry.Resolve(regs, identity, commit, content, static([]map[string]any{mintStatic(nil, stub.rogue)}, nil))
	case attestWrongName:
		payloads := []map[string]any{mintStatic(func(body map[string]any) { body["name"] = "other" }, stub.good)}
		if legacy := registry.Resolve(regs, identity, commit, content, static(payloads, nil)); legacy.Result != registry.ResultAudited {
			t.Fatalf("legacy Resolve(wrong-name) = %s, want the frozen audited outcome", legacy.Result)
		}
		resolution = registry.ResolveExact(regs, "review", identity, commit, content, static(payloads, nil))
	case attestWrongContext:
		payloads := []map[string]any{mintStatic(func(body map[string]any) {
			body["content_sha256"] = "sha256:" + strings.Repeat("e", 64)
		}, stub.good)}
		if legacy := registry.Resolve(regs, identity, commit, content, static(payloads, nil)); legacy.Result != registry.ResultAudited {
			t.Fatalf("legacy Resolve(wrong-context) = %s, want the frozen audited outcome", legacy.Result)
		}
		resolution = registry.ResolveExact(regs, "review", identity, commit, content, static(payloads, nil))
	}
	wantResult := registry.ResultUnknown
	if condition == attestRevoked {
		wantResult = registry.ResultRevoked
	}
	if resolution.Result != wantResult {
		t.Fatalf("Resolve = %s, want %s", resolution.Result, wantResult)
	}

	// Layer 1: fresh install and repair install over the live stub.
	project, home, _ := draftGitProject(t, "https://example.org/kit.git")
	stub.set(condition)
	freshBefore := draftInstallStateDigest(t, project, home)
	cfg := draftStrictRegistryConfig(home, stub)
	fresh := install.Project(cfg, project, "test", install.Options{DraftSourcesV1: true, Platform: draftPlatform()})
	if fresh.Status != "failed" {
		t.Fatalf("fresh install = %+v, want refusal", fresh)
	}
	wantErr := "is not audited by any trusted registry"
	if condition == attestRevoked {
		wantErr = "is revoked by"
	}
	if condition == attestStale {
		wantErr = "tampered snapshot"
	}
	if !strings.Contains(strings.Join(fresh.Errors, ";"), wantErr) {
		t.Fatalf("fresh install errors = %v, want %q", fresh.Errors, wantErr)
	}
	if after := draftInstallStateDigest(t, project, home); after != freshBefore {
		t.Fatal("refused fresh install published state")
	}
	// Baseline: the same tree installs cleanly under good evidence,
	// proving the refusal above is the evidence, not the fixture.
	stub.set(attestGood)
	repairProject, repairHome, repairRepo := draftGitProject(t, "https://example.org/kit.git")
	repairCfg := draftStrictRegistryConfig(repairHome, stub)
	if result := install.Project(repairCfg, repairProject, "test", install.Options{DraftSourcesV1: true, Platform: draftPlatform()}); result.Status != "ok" {
		t.Fatalf("baseline install = %+v", result)
	}
	recorded := marker.Read(filepath.Join(repairProject, ".agents", "skills", "review"))
	if recorded == nil || recorded.Attestation == nil || recorded.Attestation.Status != registry.StatusAudited {
		t.Fatalf("baseline marker attestation = %+v, want audited", recorded)
	}
	stub.set(condition)
	repairBefore := draftInstallStateDigest(t, repairProject, repairHome)
	repair := install.Project(repairCfg, repairProject, "test", install.Options{DraftSourcesV1: true, Platform: draftPlatform()})
	if repair.Status != "failed" || !strings.Contains(strings.Join(repair.Errors, ";"), wantErr) {
		t.Fatalf("repair install = %+v, want the %q refusal", repair, wantErr)
	}
	if after := draftInstallStateDigest(t, repairProject, repairHome); after != repairBefore {
		t.Fatal("refused repair published state")
	}

	// Layer 2: refresh does not consult registry evidence, so it
	// succeeds, touches no install-gated state, and the install after
	// it still refuses: refresh cannot launder the evidence.
	payload, err := os.ReadFile(filepath.Join(repairProject, "Skillfile.json"))
	if err != nil {
		t.Fatal(err)
	}
	gatedBefore := draftRefreshPreservedDigest(t, repairProject, repairHome)
	refreshDraftLockedWithRoots(t, repairProject, repairHome, string(payload), map[string]string{"s": repairRepo})
	if after := draftRefreshPreservedDigest(t, repairProject, repairHome); after != gatedBefore {
		t.Fatal("refresh under bad evidence touched install-gated state")
	}
	again := install.Project(repairCfg, repairProject, "test", install.Options{DraftSourcesV1: true, Platform: draftPlatform()})
	if again.Status != "failed" || !strings.Contains(strings.Join(again.Errors, ";"), wantErr) {
		t.Fatalf("install after refresh = %+v, want the %q refusal", again, wantErr)
	}

	// Layer 3: compiled-CLI status is nonzero and read-only.
	driveEvidenceCLIStatus(t, c, stub, condition, wantErr)
}

// draftInstallStateDigest hashes installed state excluding read-through
// registry caches and operation-private locks, which are not
// publication. A refused install must not touch anything else,
// including the machine bindings.
func draftInstallStateDigest(t *testing.T, project, home string) string {
	t.Helper()
	return treeDigestFiltered(t, project, nil) + treeDigestFiltered(t, home, func(rel string) bool {
		return rel == "cache" || strings.HasPrefix(rel, "cache"+string(filepath.Separator)) ||
			rel == "state" || strings.HasPrefix(rel, "state"+string(filepath.Separator))
	})
}

// draftRefreshPreservedDigest hashes install-gated state: everything
// draftInstallStateDigest covers except the machine bindings, which
// refresh legitimately re-publishes as resolution state.
func draftRefreshPreservedDigest(t *testing.T, project, home string) string {
	t.Helper()
	return treeDigestFiltered(t, project, nil) + treeDigestFiltered(t, home, func(rel string) bool {
		return rel == "cache" || strings.HasPrefix(rel, "cache"+string(filepath.Separator)) ||
			rel == "state" || strings.HasPrefix(rel, "state"+string(filepath.Separator)) ||
			rel == "source-bindings" || strings.HasPrefix(rel, "source-bindings"+string(filepath.Separator))
	})
}

// driveEvidenceCLIStatus proves the status half through the compiled
// CLI: attested install reports up-to-date, and the evidence
// condition flips status nonzero without any mutation.
func driveEvidenceCLIStatus(t *testing.T, c draftSemanticCase, stub *attestStub, condition attestCondition, wantErr string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("test transport wrapper is POSIX-only")
	}
	configPath, project, home, pathEnv := driveAttestedCLIInstall(t, stub)
	stub.set(condition)
	before := draftInstallStateDigest(t, project, home)
	code, stdout, stderr := runCurator(t, home, configPath, pathEnv, "status", "app")
	if code == 0 {
		t.Fatalf("status under %s succeeded:\n%s\n%s", c.ID, stdout, stderr)
	}
	if after := draftInstallStateDigest(t, project, home); after != before {
		t.Fatal("status mutated state")
	}
	_ = wantErr
}

// mergeRegistryConfig adds the stub registry and a strict no-cache
// audit policy to a bootstrapped CLI config.
func mergeRegistryConfig(t *testing.T, configPath string, stub *attestStub) {
	t.Helper()
	payload, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	doc["audit_registries"] = []any{map[string]any{
		"name": "one", "url": stub.URL, "public_keys": []any{stub.pinned}, "enabled": true,
	}}
	audit, _ := doc["audit"].(map[string]any)
	if audit == nil {
		audit = map[string]any{}
		doc["audit"] = audit
	}
	audit["registry_policy"] = "strict"
	audit["cache_ttl_seconds"] = 0
	audit["offline_grace_seconds"] = 0
	merged, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, merged, 0o644); err != nil {
		t.Fatal(err)
	}
}
