package install

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/marker"
)

// isolateTempDir points every ephemeral allocation of one test at a directory
// nothing else writes to, so the complete temporary footprint of a run is
// observable. Call it after the fixture finished allocating its own temporary
// directories, otherwise those land inside the isolated root as well.
func isolateTempDir(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	// TMPDIR is the POSIX lookup; TMP and TEMP are the Windows ones.
	t.Setenv("TMPDIR", base)
	t.Setenv("TMP", base)
	t.Setenv("TEMP", base)
	return base
}

// tempEntries lists the temporary roots present in an isolated TMPDIR.
func tempEntries(t *testing.T, base string) []string {
	t.Helper()
	entries, err := os.ReadDir(base)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

// requireOnlyOperationPrivateRoots asserts that whatever a run is currently
// holding in TMPDIR belongs to the single operation-private root. Any separate
// closure scratch workspace shows up here as an extra root.
func requireOnlyOperationPrivateRoots(t *testing.T, base string) {
	t.Helper()
	for _, name := range tempEntries(t, base) {
		if !strings.HasPrefix(name, operationPrivatePrefix) {
			t.Fatalf("temporary root %q is outside the operation-private root", name)
		}
	}
}

// requireTempDirEmpty asserts a finished run left no temporary root at all.
func requireTempDirEmpty(t *testing.T, base, context string) {
	t.Helper()
	if names := tempEntries(t, base); len(names) != 0 {
		t.Fatalf("%s left temporary roots behind: %v", context, names)
	}
}

// relocate moves a prepared skill repository out of the skills root and
// returns its new location. A declaration pointing at that location can only
// resolve by cloning, which is the path a dry run used to satisfy with a
// separate scratch workspace.
func (e *env) relocate(name, origin string) string {
	e.t.Helper()
	moved := filepath.Join(origin, name)
	if err := os.MkdirAll(origin, 0o755); err != nil {
		e.t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(e.skillsRoot, name), moved); err != nil {
		e.t.Fatal(err)
	}
	return moved
}

// skillfile renders a Skillfile declaring one tagged skill, optionally naming
// the clone URL it must be fetched from.
func (e *env) skillfile(name, gitURL string) string {
	e.t.Helper()
	decl := map[string]any{"name": name, "tag": "v1"}
	if gitURL != "" {
		decl["git"] = gitURL
	}
	payload, err := json.MarshalIndent(map[string]any{
		"schema_version": 1,
		"agents":         []string{"claude_code"},
		"skills":         []map[string]any{decl},
	}, "", "  ")
	if err != nil {
		e.t.Fatal(err)
	}
	return string(payload)
}

// declareCloned writes a project Skillfile naming a skill that is absent from
// the skills root, so resolution must clone it from gitURL.
func (e *env) declareCloned(name, gitURL string) {
	e.t.Helper()
	e.write(e.project, "Skillfile.json", e.skillfile(name, gitURL))
	e.write(e.project, ".gitignore", ".agents/\n.claude/skills/\nSkillfile.dev.json\n")
}

// globalDeclare initializes the machine-wide scope and declares one skill
// resolved from the skills root.
func (e *env) globalDeclare(name string) {
	e.t.Helper()
	e.globalDeclareCloned(name, "")
}

// globalDeclareCloned is declareCloned for the machine-wide scope.
func (e *env) globalDeclareCloned(name, gitURL string) {
	e.t.Helper()
	if _, err := GlobalInit(e.home); err != nil {
		e.t.Fatal(err)
	}
	e.write(GlobalRoot(e.home), "Skillfile.json", e.skillfile(name, gitURL))
}

// TestProjectDryRunKeepsEveryEphemeralPathInOneOperationPrivateRoot proves the
// project dry run creates no closure scratch workspace of its own. The declared
// skill lives outside the skills root, so resolution has to clone and snapshot
// it — exactly the work that previously needed a separate curator-dry-run-*
// root — and it now happens inside the single operation-private root.
func TestProjectDryRunKeepsEveryEphemeralPathInOneOperationPrivateRoot(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	origin := e.relocate("build-skill", filepath.Join(t.TempDir(), "origin"))
	e.declareCloned("build-skill", origin)
	base := isolateTempDir(t)
	deps, toolchain, _, builder := newFakeDeps(t)

	var midRun []string
	toolchain.observeProbe = func() { midRun = tempEntries(t, base) }

	result := e.install(Options{DryRun: true, Build: deps})
	if result.Status != "ok" {
		t.Fatalf("dry-run of a cloned source failed: %+v", result)
	}
	if len(result.Builds) != 1 || result.Builds[0].Skill() != "build-skill" {
		t.Fatalf("planned builds = %+v, want the cloned skill", result.Builds)
	}
	if len(builder.calls) != 0 {
		t.Fatalf("dry-run invoked the builder: %v", builder.calls)
	}
	if len(midRun) != 1 || !strings.HasPrefix(midRun[0], operationPrivatePrefix) {
		t.Fatalf("temporary roots during the dry run = %v, want exactly one %s* root",
			midRun, operationPrivatePrefix)
	}
	for _, name := range midRun {
		if strings.Contains(name, "dry-run") {
			t.Fatalf("dry-run created a separate closure scratch root %q", name)
		}
	}
	requireTempDirEmpty(t, base, "the project dry run")
	requireAbsent(t, e.persistentPaths())
	requireNoLocks(t, e.home)
	requireNoLocks(t, e.project)
	if _, err := os.Lstat(filepath.Join(e.skillsRoot, "build-skill")); !os.IsNotExist(err) {
		t.Fatalf("dry-run persisted the cloned repository: %v", err)
	}
}

// TestGlobalDryRunKeepsEveryEphemeralPathInOneOperationPrivateRoot is the
// machine-wide half of the same guarantee.
func TestGlobalDryRunKeepsEveryEphemeralPathInOneOperationPrivateRoot(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	origin := e.relocate("build-skill", filepath.Join(t.TempDir(), "origin"))
	e.globalDeclareCloned("build-skill", origin)
	userHome := t.TempDir()
	base := isolateTempDir(t)
	deps, toolchain, _, builder := newFakeDeps(t)

	var midRun []string
	toolchain.observeProbe = func() { midRun = tempEntries(t, base) }

	result := Global(e.cfg, userHome, Options{Platform: "unix", DryRun: true, Build: deps})
	if result.Status != "ok" {
		t.Fatalf("global dry-run of a cloned source failed: %+v", result)
	}
	if len(result.Builds) != 1 {
		t.Fatalf("planned builds = %+v, want one", result.Builds)
	}
	if len(builder.calls) != 0 {
		t.Fatalf("global dry-run invoked the builder: %v", builder.calls)
	}
	if len(midRun) != 1 || !strings.HasPrefix(midRun[0], operationPrivatePrefix) {
		t.Fatalf("temporary roots during the global dry run = %v, want exactly one %s* root",
			midRun, operationPrivatePrefix)
	}
	for _, name := range midRun {
		if strings.Contains(name, "dry-run") {
			t.Fatalf("global dry-run created a separate closure scratch root %q", name)
		}
	}
	requireTempDirEmpty(t, base, "the global dry run")
	for _, path := range []string{
		filepath.Join(e.home, "cache"),
		filepath.Join(e.home, "runtime"),
		filepath.Join(GlobalRoot(e.home), "skills"),
		filepath.Join(GlobalRoot(e.home), "bin"),
		filepath.Join(e.skillsRoot, "build-skill"),
	} {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("global dry-run changed persistent state at %s: %v", path, err)
		}
	}
}

// TestDryRunRemovesItsOperationPrivateRootOnFailure covers the other half of
// the lifecycle: a dry run that fails inside planning still leaves nothing
// temporary behind.
func TestDryRunRemovesItsOperationPrivateRootOnFailure(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	origin := e.relocate("build-skill", filepath.Join(t.TempDir(), "origin"))
	e.declareCloned("build-skill", origin)
	base := isolateTempDir(t)
	deps, _, cache, _ := newFakeDeps(t)
	cache.byCommand["alpha"] = buildcache.Result{
		Status: buildcache.Unsupported, Reason: "platform protection is unavailable",
	}

	result := e.install(Options{DryRun: true, Build: deps})
	if result.Status != "failed" {
		t.Fatalf("dry-run status = %q, want failed: %+v", result.Status, result)
	}
	requireTempDirEmpty(t, base, "a failed project dry run")
	requireAbsent(t, e.persistentPaths())
}

// TestGlobalDryRunRemovesItsOperationPrivateRootOnFailure is the machine-wide
// half of the failure lifecycle.
func TestGlobalDryRunRemovesItsOperationPrivateRootOnFailure(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	origin := e.relocate("build-skill", filepath.Join(t.TempDir(), "origin"))
	e.globalDeclareCloned("build-skill", origin)
	userHome := t.TempDir()
	base := isolateTempDir(t)
	deps, _, cache, _ := newFakeDeps(t)
	cache.byCommand["alpha"] = buildcache.Result{
		Status: buildcache.Unsupported, Reason: "platform protection is unavailable",
	}

	result := Global(e.cfg, userHome, Options{Platform: "unix", DryRun: true, Build: deps})
	if result.Status != "failed" {
		t.Fatalf("global dry-run status = %q, want failed: %+v", result.Status, result)
	}
	requireTempDirEmpty(t, base, "a failed global dry run")
}

// TestRealRunKeepsEveryEphemeralPathInOneOperationPrivateRoot proves the same
// containment for a staging run: the session base is allocated inside the
// operation-private root and the whole root is gone once the install returned.
func TestRealRunKeepsEveryEphemeralPathInOneOperationPrivateRoot(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	e.declare("build-skill")
	base := isolateTempDir(t)
	deps, _, _, builder := newFakeDeps(t)

	var midRun []string
	builder.observe = func(StageRequest) { midRun = tempEntries(t, base) }

	result := e.install(Options{Build: deps})
	if result.Status != "ok" {
		t.Fatalf("install failed: %+v", result)
	}
	for _, name := range midRun {
		if !strings.HasPrefix(name, operationPrivatePrefix) {
			t.Fatalf("temporary root %q during staging is outside the operation-private root", name)
		}
	}
	requireTempDirEmpty(t, base, "a committed install")
}

// TestGlobalMcpFailureBlocksToolchainCacheAndBuild proves the machine-wide
// scope runs MCP verification before any compiler-facing work, exactly like a
// project install.
func TestGlobalMcpFailureBlocksToolchainCacheAndBuild(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	e.globalDeclare("build-skill")
	userHome := t.TempDir()
	deps, toolchain, cache, builder := newFakeDeps(t)
	before := captureLiveState(t, e)

	staged := 0
	result := Global(e.cfg, userHome, Options{
		Platform: "unix", Build: deps,
		OnStaged: func(Staged) error { staged++; return nil },
		VerifyMcp: func([]*closure.Node) (map[string]map[string][]string, []string, error) {
			return nil, nil, errors.New("missing MCP server \"ledger\" for build-skill")
		},
	})
	requireGateBlockedGlobalBuild(t, e, gateBlock{
		result: result, want: "missing MCP server", toolchain: toolchain,
		cache: cache, builder: builder, staged: staged, before: before,
	})
}

// TestGlobalRegistryFailureBlocksToolchainCacheAndBuild proves the machine-wide
// scope resolves registry attestation before any compiler-facing work.
func TestGlobalRegistryFailureBlocksToolchainCacheAndBuild(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	e.globalDeclare("build-skill")
	userHome := t.TempDir()
	deps, toolchain, cache, builder := newFakeDeps(t)
	before := captureLiveState(t, e)

	staged := 0
	result := Global(e.cfg, userHome, Options{
		Platform: "unix", Build: deps,
		OnStaged: func(Staged) error { staged++; return nil },
		ResolveAttest: func([]*closure.Node) (map[string]*marker.Attestation, []string, error) {
			return nil, nil, errors.New("build-skill is revoked by registry-a")
		},
	})
	requireGateBlockedGlobalBuild(t, e, gateBlock{
		result: result, want: "is revoked by", toolchain: toolchain,
		cache: cache, builder: builder, staged: staged, before: before,
	})
}

// gateBlock bundles what a failing global gate must have prevented.
type gateBlock struct {
	result    Result
	want      string
	toolchain *fakeToolchain
	cache     *fakeCache
	builder   *fakeBuilder
	staged    int
	before    liveState
}

func requireGateBlockedGlobalBuild(t *testing.T, e *env, block gateBlock) {
	t.Helper()
	if block.result.Status != "failed" {
		t.Fatalf("global status = %q, want failed: %+v", block.result.Status, block.result)
	}
	if !containsSubstring(block.result.Errors, block.want) {
		t.Fatalf("global errors = %v, want one containing %q", block.result.Errors, block.want)
	}
	if block.toolchain.probes != 0 || block.toolchain.establishes != 0 {
		t.Fatalf("a failing gate still reached the toolchain: probes=%d establishes=%d",
			block.toolchain.probes, block.toolchain.establishes)
	}
	if len(block.cache.inspected) != 0 {
		t.Fatalf("a failing gate still inspected the protected cache: %v", block.cache.inspected)
	}
	if len(block.builder.calls) != 0 {
		t.Fatalf("a failing gate still compiled: %v", block.builder.calls)
	}
	if block.staged != 0 {
		t.Fatalf("a failing gate still handed off %d staged results", block.staged)
	}
	if len(block.result.Builds) != 0 || len(block.result.Staged) != 0 {
		t.Fatalf("a failing gate still reported builds=%d staged=%d",
			len(block.result.Builds), len(block.result.Staged))
	}
	block.before.requireUnchanged(t, e, "a failing global gate")
}

// TestGlobalMarkersCarryMcpAndAttestationEvidence proves the machine-wide scope
// records the same MCP and registry evidence a project install records.
func TestGlobalMarkersCarryMcpAndAttestationEvidence(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	e.skill("skill-g")
	e.globalDeclare("skill-g")
	userHome := t.TempDir()
	deps, _, _, _ := newFakeDeps(t)

	result := Global(e.cfg, userHome, Options{
		Platform: "unix", Build: deps,
		VerifyMcp: func([]*closure.Node) (map[string]map[string][]string, []string, error) {
			return map[string]map[string][]string{"skill-g": {"ledger": {"claude_code"}}}, nil, nil
		},
		ResolveAttest: func([]*closure.Node) (map[string]*marker.Attestation, []string, error) {
			return map[string]*marker.Attestation{
				"skill-g": {Registry: "registry-a", Status: "audited", KeyID: "0123456789abcdef"},
			}, nil, nil
		},
	})
	if result.Status != "ok" {
		t.Fatalf("global install failed: %+v", result)
	}
	recorded := marker.Read(filepath.Join(GlobalRoot(e.home), "skills", "skill-g"))
	if recorded == nil {
		t.Fatal("global marker is missing")
	}
	if got := recorded.McpServers["ledger"]; len(got) != 1 || got[0] != "claude_code" {
		t.Fatalf("global marker mcp_servers = %v, want the verified finding", recorded.McpServers)
	}
	if recorded.Attestation == nil || recorded.Attestation.Registry != "registry-a" ||
		recorded.Attestation.KeyID != "0123456789abcdef" {
		t.Fatalf("global marker attestation = %+v, want the resolved one", recorded.Attestation)
	}
}
