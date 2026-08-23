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
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/registry"
)

// authoritativeDryRunCase is the published no-mutation contract. Only the field
// names are mirrored; the forbidden surfaces themselves are read from the
// authoritative document at run time.
type authoritativeDryRunCase struct {
	Name                       string   `json:"name"`
	Scope                      string   `json:"scope"`
	ForbiddenPersistentEffects []string `json:"forbidden_persistent_effects"`
	LogicalCacheKey            string   `json:"logical_cache_key"`
	AllowedGoCommands          []string `json:"allowed_go_commands"`
	ForbiddenGoCommands        []string `json:"forbidden_go_commands"`
	ReportedBuildOutcomes      []string `json:"reported_build_outcomes"`
	OperationPrivateStateAfter string   `json:"operation_private_state_after"`
	ArtifactExecuted           bool     `json:"artifact_executed"`
}

func authoritativeDryRunCases(t *testing.T) []authoritativeDryRunCase {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "manager-lifecycle.json")) // #nosec G304 -- explicit authoritative conformance input
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		DryRunCases []authoritativeDryRunCase `json:"dry_run_cases"`
	}
	if err := json.Unmarshal(payload, &document); err != nil {
		t.Fatal(err)
	}
	if len(document.DryRunCases) == 0 {
		t.Fatal("the authoritative suite publishes no dry-run case to bind")
	}
	return document.DryRunCases
}

// dryRunBaseline is everything a published forbidden effect can be measured
// against: the surfaces that must stay absent and the bytes that must not move.
type dryRunBaseline struct {
	env           *env
	fetchable     string
	clonedName    string
	clonedOrigin  string
	configBytes   []byte
	registryCalls func() int
}

// armedDryRunEnv builds a machine where every published persistent surface is
// genuinely reachable: one skill whose repository can really be fetched, one
// skill that can only be resolved by cloning, the audit gate armed, and a
// trusted registry that answers over loopback. Without arming them, "absent
// afterwards" would prove nothing.
func armedDryRunEnv(t *testing.T) *dryRunBaseline {
	t.Helper()
	e := newEnv(t)

	// A repository that a real run would fetch: its origin is a working local
	// upstream, so a fetch would succeed and leave FETCH_HEAD behind.
	e.skill("fetchable")
	upstream := filepath.Join(t.TempDir(), "upstream")
	e.git(t.TempDir(), "clone", "-q", "--bare", filepath.Join(e.skillsRoot, "fetchable"), upstream)
	e.git(filepath.Join(e.skillsRoot, "fetchable"), "remote", "add", "origin", upstream)

	// A skill that is absent from the skills root, so resolution has to clone.
	e.skill("clonable")
	movedOrigin := e.relocate("clonable", filepath.Join(t.TempDir(), "outside"))

	e.cfg.Audit.Enabled = true
	calls := 0
	server, pinned := loopbackRegistry(t, &calls)
	t.Cleanup(server.Close)
	e.cfg.AuditRegistries = []config.Registry{
		{Name: "loopback", URL: server.URL, PublicKeys: []string{pinned}, Enabled: true},
	}

	// config.Bootstrap writes a real configuration document, so "configuration"
	// can be asserted byte for byte rather than by absence.
	if err := config.Bootstrap(e.cfg.Path, e.skillsRoot, "", []string{"claude_code"}, false); err != nil {
		t.Fatal(err)
	}
	configBytes, err := os.ReadFile(e.cfg.Path) // #nosec G304 -- test fixture path
	if err != nil {
		t.Fatal(err)
	}
	return &dryRunBaseline{
		env: e, fetchable: filepath.Join(e.skillsRoot, "fetchable"),
		clonedName: "clonable", clonedOrigin: movedOrigin,
		configBytes: configBytes, registryCalls: func() int { return calls },
	}
}

// loopbackRegistry is a trusted registry that answers over loopback and counts
// the requests it served, so a response cache can only stay empty because the
// run refused to persist one — not because nothing was ever fetched.
func loopbackRegistry(t *testing.T, calls *int) (*httptest.Server, string) {
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
		*calls++
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/v1/snapshot"):
			_ = json.NewEncoder(w).Encode(sign(map[string]any{
				"schema_version": 1, "merkle_root": strings.Repeat("a", 64), "log_size": 1,
				"head": strings.Repeat("b", 64), "version": 1,
				"created_at": time.Now().UTC().Format(time.RFC3339),
			}))
		case strings.HasSuffix(r.URL.Path, "/v1/records"):
			_ = json.NewEncoder(w).Encode(map[string]any{"records": []any{}, "next_cursor": nil})
		default:
			http.NotFound(w, r)
		}
	}))
	return server, pinned
}

// declareBoth writes one scope manifest naming both the fetchable skill and the
// skill that can only be cloned.
func (baseline *dryRunBaseline) declareBoth(t *testing.T, root string) {
	t.Helper()
	payload, err := json.MarshalIndent(map[string]any{
		"schema_version": 1,
		"agents":         []string{"claude_code"},
		"skills": []map[string]any{
			{"name": "fetchable", "tag": "v1"},
			{"name": baseline.clonedName, "git": baseline.clonedOrigin, "tag": "v1"},
		},
	}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	baseline.env.write(root, manifest.Name, string(payload))
}

// TestAuthoritativeDryRunCasesMutateNothingPersistent binds each published
// dry-run case to a real planning run in the scope it names, then proves every
// published forbidden effect individually. An effect with no binding fails the
// test rather than passing unasserted.
func TestAuthoritativeDryRunCasesMutateNothingPersistent(t *testing.T) {
	for _, published := range authoritativeDryRunCases(t) {
		published := published
		t.Run(published.Name, func(t *testing.T) {
			if len(published.ForbiddenPersistentEffects) == 0 {
				t.Fatalf("published dry-run case %q forbids nothing", published.Name)
			}
			if published.Scope == "multi-project" {
				runAuthoritativeCompiledDryRun(t, published)
				return
			}
			baseline := armedDryRunEnv(t)
			e := baseline.env

			var result Result
			switch published.Scope {
			case "project":
				baseline.declareBoth(t, e.project)
				e.write(e.project, ".gitignore", ".agents/\n.claude/skills/\nSkillfile.dev.json\n")
				result = Project(e.cfg, e.project, "test", Options{Platform: installPlatform(), DryRun: true, Fetch: true})
			case "global":
				if _, err := GlobalInit(e.home); err != nil {
					t.Fatal(err)
				}
				baseline.declareBoth(t, GlobalRoot(e.home))
				result = Global(e.cfg, t.TempDir(), Options{Platform: installPlatform(), DryRun: true, Fetch: true})
			default:
				t.Fatalf("published dry-run scope %q has no executable binding", published.Scope)
			}
			if result.Status != "ok" {
				t.Fatalf("the %s dry run did not plan: %+v", published.Scope, result)
			}
			// The plan really did the work whose side effects are forbidden: it
			// resolved both declared skills, one of which only exists remotely.
			if !strings.Contains(strings.Join(result.Messages, "\n"), baseline.clonedName) {
				t.Fatalf("the dry run never resolved the skill it had to clone:\n%s",
					strings.Join(result.Messages, "\n"))
			}
			// The registry surfaces can only be proven unwritten if the run
			// actually reached a registry.
			for _, effect := range published.ForbiddenPersistentEffects {
				if effect != "response-cache" && effect != "registry-state" {
					continue
				}
				if baseline.registryCalls() == 0 {
					t.Fatalf("no registry was contacted, so %q proves nothing", effect)
				}
				break
			}
			for _, effect := range published.ForbiddenPersistentEffects {
				baseline.assertNoEffect(t, effect)
			}
		})
	}
}

// runAuthoritativeCompiledDryRun is the rc.8 compatibility mapping for the
// compiled multi-project dry-run contract. The published logical key remains
// the build receipt identity; Curator's protected cache address is separately
// assurance-bound and must not replace it.
func runAuthoritativeCompiledDryRun(t *testing.T, published authoritativeDryRunCase) {
	t.Helper()
	if published.Name != "compiled-cache-miss-is-read-only" || published.OperationPrivateStateAfter != "absent" || published.ArtifactExecuted {
		t.Fatalf("unknown compiled dry-run contract: %+v", published)
	}
	if !sameStrings(published.AllowedGoCommands, []string{"telemetry-off", "version", "env"}) || !sameStrings(published.ForbiddenGoCommands, []string{"list", "build"}) {
		t.Fatalf("compiled dry-run Go command contract drifted: allowed=%v forbidden=%v", published.AllowedGoCommands, published.ForbiddenGoCommands)
	}
	for _, effect := range published.ForbiddenPersistentEffects {
		if !compiledDryRunEffectCovered(effect) {
			t.Fatalf("published compiled dry-run effect %q has no executable assertion", effect)
		}
	}
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	expectedKeyBytes, err := os.ReadFile(filepath.Join(root, "expected", "build-driver", "cache-key.txt")) // #nosec G304 -- fixed path below the explicit conformance root
	if err != nil {
		t.Fatalf("read published build-driver cache key: %v", err)
	}
	expectedKey := strings.TrimSpace(string(expectedKeyBytes))
	if published.LogicalCacheKey != expectedKey {
		t.Fatalf("manager lifecycle logical cache key = %q, want build-driver identity %q", published.LogicalCacheKey, expectedKey)
	}

	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	e.declare("build-skill")
	base := isolateTempDir(t)
	deps, toolchain, _, builder := newFakeDeps(t)
	before := captureLiveState(t, e)

	result := e.install(Options{DryRun: true, Build: deps})
	if result.Status != "ok" || len(result.Builds) != 1 {
		t.Fatalf("compiled dry run did not produce one plan: %+v", result)
	}
	planned := result.Builds[0]
	if !validPublishedSHA256(published.LogicalCacheKey) || !validPublishedSHA256(string(planned.logicalKey)) || !validPublishedSHA256(string(planned.CacheKey())) {
		t.Fatalf("invalid published, logical, or protected identity: published=%q logical=%q protected=%q", published.LogicalCacheKey, planned.logicalKey, planned.CacheKey())
	}
	if planned.logicalKey == planned.CacheKey() {
		t.Fatal("assurance-bound protected cache address aliases the preserved logical receipt key")
	}
	if !containsString(published.ReportedBuildOutcomes, string(planned.Outcome())) {
		t.Fatalf("dry-run outcome %q is not published in %v", planned.Outcome(), published.ReportedBuildOutcomes)
	}
	if toolchain.probes != 1 || toolchain.establishes != 0 || len(builder.calls) != 0 {
		t.Fatalf("dry-run commands: probes=%d sessions=%d builds=%v", toolchain.probes, toolchain.establishes, builder.calls)
	}
	if planned.ArtifactPath() != "" {
		t.Fatalf("dry run exposed an executable artifact: %s", planned.ArtifactPath())
	}
	requireTempDirEmpty(t, base, "the authoritative compiled dry run")
	before.requireUnchanged(t, e, "the authoritative compiled dry run")
	requireNoLocks(t, e.home)
	requireNoLocks(t, e.project)
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range want {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}

func validPublishedSHA256(value string) bool {
	if len(value) != len("sha256:")+64 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	for _, char := range value[len("sha256:"):] {
		if char < '0' || char > '9' {
			if char < 'a' || char > 'f' {
				return false
			}
		}
	}
	return true
}

func compiledDryRunEffectCovered(effect string) bool {
	switch effect {
	case "source-checkout", "snapshot-cache", "response-cache", "toolchain-probe-memo",
		"module-cache", "go-build-cache", "compiled-artifact-cache", "audit-state",
		"registry-state", "revocation-state", "configuration", "project-lock",
		"cache-build-lock", "manager-home-lock", "journal", "backup", "quarantine",
		"permission-repair", "context-tree", "runtime-tree", "environment-file",
		"install-marker", "command-shim", "adapter-ledger", "adapter-mirror",
		"consumer-ledger", "gc-metadata":
		return true
	default:
		return false
	}
}

func (baseline *dryRunBaseline) assertNoEffect(t *testing.T, effect string) {
	t.Helper()
	e := baseline.env
	switch effect {
	case "source-fetch":
		requireMissing(t, effect, filepath.Join(baseline.fetchable, ".git", "FETCH_HEAD"))
	case "source-clone":
		requireMissing(t, effect, filepath.Join(e.skillsRoot, baseline.clonedName),
			filepath.Join(e.home, "dev"))
	case "snapshot-cache":
		requireMissing(t, effect, filepath.Join(e.home, "cache"))
	case "response-cache":
		requireMissing(t, effect, filepath.Join(e.home, "cache", "registry"))
	case "audit-state":
		requireMissing(t, effect, filepath.Join(e.home, "audit"))
	case "registry-state":
		requireMissing(t, effect, filepath.Join(e.home, "state", "registry"))
	case "configuration":
		payload, err := os.ReadFile(e.cfg.Path) // #nosec G304 -- test fixture path
		if err != nil {
			t.Fatalf("the dry run removed the configuration: %v", err)
		}
		if string(payload) != string(baseline.configBytes) {
			t.Fatalf("the dry run rewrote the configuration:\n%s", payload)
		}
	case "runtime":
		requireMissing(t, effect, filepath.Join(e.home, "runtime"))
	case "project-artifacts":
		requireMissing(t, effect, filepath.Join(e.project, ".agents"), filepath.Join(e.project, ".claude"))
	case "global-artifacts":
		requireMissing(t, effect,
			filepath.Join(e.home, "global", "skills"), filepath.Join(e.home, "global", "bin"))
	default:
		t.Fatalf("published forbidden effect %q has no executable binding", effect)
	}
}

func requireMissing(t *testing.T, effect string, paths ...string) {
	t.Helper()
	for _, path := range paths {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("the dry run produced the forbidden persistent effect %q at %s: %v", effect, path, err)
		}
	}
}
