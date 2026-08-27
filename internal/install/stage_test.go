package install

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/marker"
)

// fakeSession stands in for a trusted Go session. Its root models the
// operation-private staging area the driver owns and deletes on Release.
type fakeSession struct {
	target     buildmeta.Target
	toolchain  buildmeta.Toolchain
	root       string
	released   int
	releaseErr error
	verified   int
	verifyErr  error
	// lateVerifyErr is reported by every verification after the first, so any
	// re-check taken during teardown surfaces as a failed installation.
	lateVerifyErr error
}

func (session *fakeSession) Target() buildmeta.Target       { return session.target }
func (session *fakeSession) Toolchain() buildmeta.Toolchain { return session.toolchain }

func (session *fakeSession) VerifyToolchain(context.Context) error {
	session.verified++
	if session.verified > 1 && session.lateVerifyErr != nil {
		return session.lateVerifyErr
	}
	return session.verifyErr
}

func (session *fakeSession) Release() error {
	session.released++
	if err := os.RemoveAll(session.root); err != nil {
		return err
	}
	return session.releaseErr
}

type fakeToolchain struct {
	t             *testing.T
	target        buildmeta.Target
	toolchain     buildmeta.Toolchain
	probes        int
	establishes   int
	session       *fakeSession
	probeErr      error
	establishErr  error
	releaseErr    error
	verifyErr     error
	lateVerifyErr error
	// observeProbe runs while the plan is mid-flight, so a test can record the
	// complete temporary footprint a dry run holds at its widest point.
	observeProbe func()
}

func (toolchain *fakeToolchain) Probe(context.Context) (buildmeta.Target, buildmeta.Toolchain, error) {
	toolchain.probes++
	if toolchain.observeProbe != nil {
		toolchain.observeProbe()
	}
	if toolchain.probeErr != nil {
		return buildmeta.Target{}, buildmeta.Toolchain{}, toolchain.probeErr
	}
	return toolchain.target, toolchain.toolchain, nil
}

func (toolchain *fakeToolchain) Establish(context.Context) (BuildSession, error) {
	toolchain.establishes++
	if toolchain.establishErr != nil {
		return nil, toolchain.establishErr
	}
	toolchain.session = &fakeSession{
		target:        toolchain.target,
		toolchain:     toolchain.toolchain,
		root:          toolchain.t.TempDir(),
		releaseErr:    toolchain.releaseErr,
		verifyErr:     toolchain.verifyErr,
		lateVerifyErr: toolchain.lateVerifyErr,
	}
	return toolchain.session, nil
}

type fakeCache struct {
	byCommand map[string]buildcache.Result
	inspected []string
	// observe runs inside the lookup, modelling any state change that races a
	// cache decision the planner is about to trust.
	observe func(buildcache.Expectation)
}

func (cache *fakeCache) Inspect(expect buildcache.Expectation) buildcache.Result {
	cache.inspected = append(cache.inspected, expect.Input.Command)
	if cache.observe != nil {
		cache.observe(expect)
	}
	if result, present := cache.byCommand[expect.Input.Command]; present {
		return result
	}
	return buildcache.Result{Status: buildcache.Miss, Reason: "cache entry is absent"}
}

type fakeBuilder struct {
	t       *testing.T
	calls   []string
	staged  []string
	failOn  map[string]error
	observe func(StageRequest)
}

func (builder *fakeBuilder) Stage(_ context.Context, request StageRequest) (StagedArtifact, error) {
	builder.t.Helper()
	builder.calls = append(builder.calls, request.Command)
	if builder.observe != nil {
		builder.observe(request)
	}
	if err := builder.failOn[request.Command]; err != nil {
		return StagedArtifact{}, err
	}
	session, ok := request.Session.(*fakeSession)
	if !ok {
		return StagedArtifact{}, errors.New("staging outside a trusted session")
	}
	relative, err := buildmeta.ArtifactPath(request.Command, session.target.GOOS)
	if err != nil {
		return StagedArtifact{}, err
	}
	path := filepath.Join(session.root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return StagedArtifact{}, err
	}
	payload := []byte("artifact:" + request.Command)
	if err := os.WriteFile(path, payload, 0o700); err != nil {
		return StagedArtifact{}, err
	}
	builder.staged = append(builder.staged, path)
	digest := sha256.Sum256(payload)
	return StagedArtifact{
		Path: path,
		Metadata: buildmeta.Artifact{
			Path:   relative,
			SHA256: "sha256:" + hex.EncodeToString(digest[:]),
			Size:   int64(len(payload)),
		},
	}, nil
}

type fixedClock struct{ at time.Time }

func (clock fixedClock) Now() time.Time { return clock.at }

// countingGeneration proves the moved-tag gate reads persistent installation
// state through the injected reader and never opens it directly.
type countingGeneration struct{ reads int }

func (generation *countingGeneration) InstalledMarker(installedDir string) *marker.Marker {
	generation.reads++
	return marker.Read(installedDir)
}

func testTarget() buildmeta.Target {
	return buildmeta.Target{GOOS: "linux", GOARCH: "amd64", Tuning: map[string]string{"GOAMD64": "v1"}}
}

func testToolchain() buildmeta.Toolchain {
	return buildmeta.Toolchain{
		Algorithm:     buildmeta.ToolchainAlgorithm,
		GoRelpath:     buildmeta.ToolchainGoRelpath,
		GoVersion:     "1.25.0",
		ContentSHA256: "sha256:" + strings.Repeat("a", 64),
	}
}

func newFakeDeps(t *testing.T) (BuildDeps, *fakeToolchain, *fakeCache, *fakeBuilder) {
	t.Helper()
	toolchain := &fakeToolchain{t: t, target: testTarget(), toolchain: testToolchain()}
	cache := &fakeCache{byCommand: map[string]buildcache.Result{}}
	builder := &fakeBuilder{t: t, failOn: map[string]error{}}
	deps := BuildDeps{
		Toolchain:  toolchain,
		Cache:      cache,
		Builder:    builder,
		Clock:      fixedClock{at: time.Unix(1_700_000_000, 0).UTC()},
		Generation: &countingGeneration{},
	}
	return deps, toolchain, cache, builder
}

// buildSkill creates a schema v6 skill repository exporting one build command
// per name, all below a single module root.
func (e *env) buildSkill(name string, commands ...string) {
	e.t.Helper()
	e.buildSkillWithRequirement(name, "", commands...)
}

// buildSkillWithRequirement is buildSkill plus a full-mode requirement on a
// provider skill, so the closure orders the provider first.
func (e *env) buildSkillWithRequirement(name, requires string, commands ...string) {
	e.t.Helper()
	dir := filepath.Join(e.skillsRoot, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		e.t.Fatal(err)
	}
	e.git(dir, "init", "-q", "-b", "main")
	e.write(dir, "SKILL.md", "---\nname: "+name+"\ndescription: d\n---\n# "+name+"\n")
	e.write(dir, "src/go.mod", "module example.com/"+name+"\n")
	declared := map[string]any{}
	for _, command := range commands {
		e.write(dir, "src/cmd/"+command+"/main.go", "package main\n\nfunc main() {}\n")
		declared[command] = map[string]any{
			"type": "build", "driver": "go-v1", "source_dir": "src/cmd/" + command,
		}
	}
	spec := map[string]any{
		"schema_version": 6,
		"build_roots":    []string{"src"},
		"capabilities":   map[string]any{},
		"commands":       declared,
	}
	if requires != "" {
		spec["dependencies"] = map[string]any{"skills": map[string]any{
			requires: map[string]any{
				"git":  "./" + requires,
				"ref":  map[string]any{"kind": "tag", "value": "v1"},
				"mode": "full",
			},
		}}
	}
	payload, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		e.t.Fatal(err)
	}
	e.write(dir, "agent-skill.json", string(payload))
	e.git(dir, "add", ".")
	e.git(dir, "commit", "-qm", "init")
	e.git(dir, "tag", "v1")
}

// persistentPaths enumerates every scope-level path a plan or staging failure
// must leave absent or unchanged.
func (e *env) persistentPaths() []string {
	return []string{
		filepath.Join(e.project, ".agents"),
		filepath.Join(e.project, ".claude"),
		filepath.Join(e.home, "consumers.json"),
		filepath.Join(e.home, "runtime"),
		filepath.Join(e.home, "cache", "build"),
		filepath.Join(e.home, "hybrid"),
		filepath.Join(e.home, "global"),
	}
}

func requireAbsent(t *testing.T, paths []string) {
	t.Helper()
	for _, path := range paths {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("persistent path %s exists or is unreadable: %v", path, err)
		}
	}
}

// treeDigest renders a stable byte-for-byte fingerprint of a directory tree.
func treeDigest(t *testing.T, root string) string {
	t.Helper()
	var lines []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		relative, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		if info.IsDir() {
			lines = append(lines, "d "+filepath.ToSlash(relative))
			return nil
		}
		payload, readErr := os.ReadFile(path) // #nosec G304 -- test fixture
		if readErr != nil {
			return readErr
		}
		digest := sha256.Sum256(payload)
		lines = append(lines, fmt.Sprintf("f %s %d %s", filepath.ToSlash(relative), info.Size(), hex.EncodeToString(digest[:])))
		return nil
	})
	if err != nil {
		t.Fatalf("digest %s: %v", root, err)
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

// snapshotDir locates the frozen commit snapshot of one skill in the machine
// snapshot cache, so a test can mutate a source the planner already froze.
func (e *env) snapshotDir(name string) string {
	e.t.Helper()
	root := filepath.Join(e.home, "cache", name)
	var found []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name() == "snapshot" {
			found = append(found, path)
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		e.t.Fatalf("walk %s: %v", root, err)
	}
	if len(found) != 1 {
		e.t.Fatalf("snapshot directories for %s = %v, want exactly one", name, found)
	}
	return found[0]
}

// mutateSnapshot changes a frozen build source in place. Only the tree content
// matters here: any added file breaks the identity the plan committed to.
func (e *env) mutateSnapshot(name string) {
	e.t.Helper()
	e.write(e.snapshotDir(name), "src/injected.go", "package injected\n")
}

// liveState is the byte-for-byte fingerprint of everything an installation may
// mutate. A failure before the first live write must leave it identical.
type liveState struct {
	project   string
	runtime   string
	cache     string
	consumers string
	global    string
}

func captureLiveState(t *testing.T, e *env) liveState {
	t.Helper()
	consumers, err := os.ReadFile(filepath.Join(e.home, "consumers.json"))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return liveState{
		project:   treeDigest(t, filepath.Join(e.project, ".agents")),
		runtime:   treeDigest(t, filepath.Join(e.home, "runtime")),
		cache:     treeDigest(t, filepath.Join(e.home, "cache", "build")),
		consumers: string(consumers),
		global:    treeDigest(t, GlobalRoot(e.home)),
	}
}

func (before liveState) requireUnchanged(t *testing.T, e *env, context string) {
	t.Helper()
	after := captureLiveState(t, e)
	if after.project != before.project {
		t.Fatalf("%s: installed project tree changed", context)
	}
	if after.runtime != before.runtime {
		t.Fatalf("%s: runtime store changed", context)
	}
	if after.cache != before.cache {
		t.Fatalf("%s: live build cache changed", context)
	}
	if after.consumers != before.consumers {
		t.Fatalf("%s: consumer ledger changed", context)
	}
	if after.global != before.global {
		t.Fatalf("%s: global scope changed", context)
	}
}

// seedLiveCache writes a prior live build-cache entry that no failing run may
// touch.
func (e *env) seedLiveCache() {
	e.t.Helper()
	e.write(filepath.Join(e.home, "cache", "build"), "go-v1/existing/artifact", "prior")
}

func TestDryRunPlansBuildsWithoutToolchainSessionOrPersistentState(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	e.declare("build-skill")
	deps, toolchain, cache, builder := newFakeDeps(t)

	result := e.install(Options{DryRun: true, Build: deps})
	if result.Status != "ok" {
		t.Fatalf("dry-run failed: %+v", result)
	}
	if toolchain.probes != 1 || toolchain.establishes != 0 {
		t.Fatalf("dry-run probes=%d establishes=%d, want exactly one probe and no session",
			toolchain.probes, toolchain.establishes)
	}
	if len(builder.calls) != 0 {
		t.Fatalf("dry-run invoked the builder: %v", builder.calls)
	}
	if len(cache.inspected) != 1 || cache.inspected[0] != "alpha" {
		t.Fatalf("dry-run cache inspection = %v, want exactly [alpha]", cache.inspected)
	}
	if len(result.Builds) != 1 {
		t.Fatalf("planned builds = %d, want 1", len(result.Builds))
	}
	planned := result.Builds[0]
	if planned.Outcome() != BuildWouldPreflightAndBuild {
		t.Fatalf("planned outcome = %q, want %q", planned.Outcome(), BuildWouldPreflightAndBuild)
	}
	line := planned.Describe()
	for _, want := range []string{
		"build-skill.alpha build",
		"source=curator-build-source-v1:sha256:",
		"root=src dir=src/cmd/alpha",
		"target=linux/amd64+GOAMD64=v1",
		"key=sha256:",
		"outcome=would-preflight-and-build",
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("dry-run line %q is missing %q", line, want)
		}
	}
	if string(planned.CacheKey()) == "" || !strings.Contains(line, "key="+string(planned.CacheKey())) {
		t.Fatalf("dry-run line %q does not report the exact planned key %q", line, planned.CacheKey())
	}
	if !containsLine(result.Messages, "test: "+line) {
		t.Fatalf("dry-run messages omit the plan line: %v", result.Messages)
	}
	requireAbsent(t, e.persistentPaths())
	requireNoLocks(t, e.home)
	requireNoLocks(t, e.project)
}

// requireNoLocks asserts a read-only phase left no cross-process lock behind.
func requireNoLocks(t *testing.T, root string) {
	t.Helper()
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir // git's own transient locks are not manager state
		}
		if !info.IsDir() && strings.HasSuffix(path, ".lock") {
			t.Fatalf("dry-run left a lock at %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
}

func TestGlobalDryRunPlansBuildsWithoutSessionOrPersistentState(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	if _, err := GlobalInit(e.home); err != nil {
		t.Fatal(err)
	}
	e.write(GlobalRoot(e.home), "Skillfile.json", `{
		"schema_version": 1, "agents": ["claude_code"],
		"skills": [{"name": "build-skill", "tag": "v1"}]}`)
	deps, toolchain, cache, builder := newFakeDeps(t)

	result := Global(e.cfg, t.TempDir(), Options{Platform: "unix", DryRun: true, Build: deps})
	if result.Status != "ok" {
		t.Fatalf("global dry-run failed: %+v", result)
	}
	if toolchain.probes != 1 || toolchain.establishes != 0 {
		t.Fatalf("global dry-run probes=%d establishes=%d, want one probe and no session",
			toolchain.probes, toolchain.establishes)
	}
	if len(builder.calls) != 0 {
		t.Fatalf("global dry-run invoked the builder: %v", builder.calls)
	}
	if len(cache.inspected) != 1 || len(result.Builds) != 1 {
		t.Fatalf("global dry-run cache=%v builds=%d, want one of each", cache.inspected, len(result.Builds))
	}
	if !containsLine(result.Messages, "global: "+result.Builds[0].Describe()) {
		t.Fatalf("global dry-run messages omit the plan line: %v", result.Messages)
	}
	for _, path := range []string{
		filepath.Join(e.home, "runtime"),
		filepath.Join(e.home, "cache", "build"),
		filepath.Join(GlobalRoot(e.home), "skills"),
		filepath.Join(GlobalRoot(e.home), "bin"),
	} {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("global dry-run changed persistent state at %s: %v", path, err)
		}
	}
}

func TestGlobalStagingFailureLeavesGlobalScopeUnchanged(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-g")
	e.buildSkill("build-skill", "alpha")
	if _, err := GlobalInit(e.home); err != nil {
		t.Fatal(err)
	}
	e.write(GlobalRoot(e.home), "Skillfile.json", `{
		"schema_version": 1, "agents": ["claude_code"],
		"skills": [{"name": "skill-g", "tag": "v1"}]}`)
	userHome := t.TempDir()
	baseline, _, _, _ := newFakeDeps(t)
	if result := Global(e.cfg, userHome, Options{Platform: "unix", Build: baseline}); result.Status != "ok" {
		t.Fatalf("baseline global install failed: %+v", result)
	}
	e.write(GlobalRoot(e.home), "Skillfile.json", `{
		"schema_version": 1, "agents": ["claude_code"],
		"skills": [{"name": "skill-g", "tag": "v1"}, {"name": "build-skill", "tag": "v1"}]}`)
	globalBefore := treeDigest(t, GlobalRoot(e.home))
	runtimeBefore := treeDigest(t, filepath.Join(e.home, "runtime"))
	deps, _, _, builder := newFakeDeps(t)
	builder.failOn["alpha"] = errors.New("go build failed")

	result := Global(e.cfg, userHome, Options{Platform: "unix", Build: deps})
	if result.Status != "failed" {
		t.Fatalf("global install status = %q, want failed", result.Status)
	}
	if got := treeDigest(t, GlobalRoot(e.home)); got != globalBefore {
		t.Fatal("global scope changed after a staging failure")
	}
	if got := treeDigest(t, filepath.Join(e.home, "runtime")); got != runtimeBefore {
		t.Fatal("runtime store changed after a global staging failure")
	}
	if _, err := os.Lstat(filepath.Join(e.home, "cache", "build")); !os.IsNotExist(err) {
		t.Fatalf("global staging failure touched the live build cache: %v", err)
	}
}

func TestDryRunReportsCacheHitWithoutBuilding(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	e.declare("build-skill")
	deps, _, cache, builder := newFakeDeps(t)
	cache.byCommand["alpha"] = buildcache.Result{Status: buildcache.Hit, ArtifactPath: "/protected/bin/alpha"}

	result := e.install(Options{DryRun: true, Build: deps})
	if result.Status != "ok" {
		t.Fatalf("dry-run failed: %+v", result)
	}
	if len(builder.calls) != 0 {
		t.Fatalf("cache hit invoked the builder: %v", builder.calls)
	}
	if got := result.Builds[0].Outcome(); got != BuildCacheHit {
		t.Fatalf("outcome = %q, want %q", got, BuildCacheHit)
	}
	if got := result.Builds[0].ArtifactPath(); got != "/protected/bin/alpha" {
		t.Fatalf("hit artifact path = %q", got)
	}
	requireAbsent(t, e.persistentPaths())
}

func TestCacheHitPerformsNoSourceAwareGoCommand(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha", "beta")
	e.declare("build-skill")
	deps, _, cache, builder := newFakeDeps(t)
	cache.byCommand["alpha"] = buildcache.Result{Status: buildcache.Hit, ArtifactPath: "/protected/bin/alpha"}

	result := e.install(Options{Build: deps})
	if result.Status != "ok" {
		t.Fatalf("install failed: %+v", result)
	}
	if len(builder.calls) != 1 || builder.calls[0] != "beta" {
		t.Fatalf("builder calls = %v, want only the miss [beta]", builder.calls)
	}
	if len(result.Staged) != 1 || result.Staged[0].Command() != "beta" {
		t.Fatalf("staged = %+v, want only beta", result.Staged)
	}
}

func TestStagingRunsProviderFirstAndCommandLexical(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("provider", "p-beta", "p-alpha")
	e.buildSkillWithRequirement("consumer", "provider", "c-beta", "c-alpha")
	e.declare("consumer")
	deps, _, _, builder := newFakeDeps(t)

	result := e.install(Options{Build: deps})
	if result.Status != "ok" {
		t.Fatalf("install failed: %+v", result)
	}
	want := []string{"p-alpha", "p-beta", "c-alpha", "c-beta"}
	if strings.Join(builder.calls, ",") != strings.Join(want, ",") {
		t.Fatalf("build order = %v, want provider-first and command-lexical %v", builder.calls, want)
	}
	var stagedOrder []string
	for _, staged := range result.Staged {
		stagedOrder = append(stagedOrder, staged.Command())
	}
	if strings.Join(stagedOrder, ",") != strings.Join(want, ",") {
		t.Fatalf("staged order = %v, want %v", stagedOrder, want)
	}
}

func TestStagedOutputsStayPrivateAndAreReleased(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	e.declare("build-skill")
	deps, toolchain, _, builder := newFakeDeps(t)

	var observedDuringRun []string
	result := e.install(Options{Build: deps, OnStaged: func(staged Staged) error {
		// Trust is finalized before the handoff, not while the plan is released.
		if toolchain.session == nil || toolchain.session.verified != 1 {
			t.Fatalf("session verification at handoff = %+v, want exactly one completed check", toolchain.session)
		}
		if toolchain.session.released != 0 {
			t.Fatal("the plan was released before the staged outputs were handed off")
		}
		for _, build := range staged.Builds() {
			if _, err := os.Stat(build.Path()); err != nil {
				t.Fatalf("staged artifact %s is not readable during the run: %v", build.Path(), err)
			}
			observedDuringRun = append(observedDuringRun, build.Path())
		}
		return nil
	}})
	if result.Status != "ok" {
		t.Fatalf("install failed: %+v", result)
	}
	if len(observedDuringRun) != 1 {
		t.Fatalf("OnStaged observed %d outputs, want 1", len(observedDuringRun))
	}
	session := toolchain.session
	if session == nil || session.released != 1 {
		t.Fatalf("session release count = %+v, want exactly one release", session)
	}
	// The release path is teardown, so the pre-handoff check stays the only
	// toolchain verdict the whole run took.
	if session.verified != 1 {
		t.Fatalf("session verification count = %d, want exactly one pre-handoff check", session.verified)
	}
	for _, path := range builder.staged {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("staged artifact %s survived the run: %v", path, err)
		}
		if strings.HasPrefix(path, e.project) || strings.HasPrefix(path, e.home) {
			t.Fatalf("staged artifact %s was written inside an installation scope", path)
		}
	}
	receipt := result.Staged[0].Receipt()
	if receipt.CacheKey != result.Builds[0].CacheKey() {
		t.Fatalf("staged receipt key %q does not match the planned key %q", receipt.CacheKey, result.Builds[0].CacheKey())
	}
	if err := receipt.Validate(); err != nil {
		t.Fatalf("staged receipt is invalid: %v", err)
	}
	// Nothing publishes yet, so no live cache entry may appear.
	if _, err := os.Lstat(filepath.Join(e.home, "cache", "build")); !os.IsNotExist(err) {
		t.Fatalf("staging published into the live build cache: %v", err)
	}
}

func TestSecondBuildFailurePreservesPriorInstallationAndLiveCache(t *testing.T) {
	e := newEnv(t)
	e.skill("script-skill")
	e.buildSkill("build-skill", "alpha", "beta")
	e.declare("script-skill")
	deps, _, _, _ := newFakeDeps(t)

	// A prior, purely script-based installation is the state that must survive.
	if first := e.install(Options{Build: deps}); first.Status != "ok" {
		t.Fatalf("baseline install failed: %+v", first)
	}
	liveCache := filepath.Join(e.home, "cache", "build")
	if err := os.MkdirAll(filepath.Join(liveCache, "go-v1", "existing"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(liveCache, "go-v1", "existing", "artifact"), []byte("prior"), 0o600); err != nil {
		t.Fatal(err)
	}
	projectBefore := treeDigest(t, filepath.Join(e.project, ".agents"))
	runtimeBefore := treeDigest(t, filepath.Join(e.home, "runtime"))
	cacheBefore := treeDigest(t, liveCache)
	consumersBefore, err := os.ReadFile(filepath.Join(e.home, "consumers.json"))
	if err != nil {
		t.Fatal(err)
	}

	e.declare("script-skill", "build-skill")
	failing, _, _, builder := newFakeDeps(t)
	builder.failOn["beta"] = errors.New("go build failed")

	result := e.install(Options{Build: failing})
	if result.Status != "failed" {
		t.Fatalf("install status = %q, want failed", result.Status)
	}
	if strings.Join(builder.calls, ",") != "alpha,beta" {
		t.Fatalf("builder calls = %v, want alpha then beta", builder.calls)
	}
	for _, path := range builder.staged {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("staging for %s survived a later build failure: %v", path, err)
		}
	}
	if got := treeDigest(t, filepath.Join(e.project, ".agents")); got != projectBefore {
		t.Fatalf("installed project tree changed after a staging failure")
	}
	if got := treeDigest(t, filepath.Join(e.home, "runtime")); got != runtimeBefore {
		t.Fatalf("runtime store changed after a staging failure")
	}
	if got := treeDigest(t, liveCache); got != cacheBefore {
		t.Fatalf("live build cache changed after a staging failure")
	}
	consumersAfter, err := os.ReadFile(filepath.Join(e.home, "consumers.json"))
	if err != nil || string(consumersAfter) != string(consumersBefore) {
		t.Fatalf("consumer record changed after a staging failure: %v", err)
	}
}

// installBaseline performs a script-only installation and seeds a live build
// cache entry, so the state a later failure must preserve actually exists.
func (e *env) installBaseline() liveState {
	e.t.Helper()
	deps, _, _, _ := newFakeDeps(e.t)
	if first := e.install(Options{Build: deps}); first.Status != "ok" {
		e.t.Fatalf("baseline install failed: %+v", first)
	}
	e.seedLiveCache()
	return captureLiveState(e.t, e)
}

func TestToolchainDriftAfterTheFinalBuildBlocksHandoffAndPreservesLiveState(t *testing.T) {
	e := newEnv(t)
	e.skill("script-skill")
	e.buildSkill("build-skill", "alpha")
	e.declare("script-skill")
	before := e.installBaseline()

	e.declare("script-skill", "build-skill")
	deps, toolchain, _, builder := newFakeDeps(t)
	toolchain.verifyErr = errors.New("toolchain tree changed during operation")
	handoffs := 0

	result := e.install(Options{Build: deps, OnStaged: func(Staged) error {
		handoffs++
		return nil
	}})
	if result.Status != "failed" {
		t.Fatalf("install status = %q, want failed when the toolchain drifts during the build", result.Status)
	}
	if !containsSubstring(result.Errors, "toolchain tree changed during operation") {
		t.Fatalf("errors do not report the drift: %v", result.Errors)
	}
	if len(builder.calls) != 1 {
		t.Fatalf("builder calls = %v, want the drift to be caught after the build", builder.calls)
	}
	if toolchain.session == nil || toolchain.session.verified != 1 {
		t.Fatalf("session verification count = %+v, want exactly one pre-handoff check", toolchain.session)
	}
	if handoffs != 0 {
		t.Fatal("staged outputs were handed off after the toolchain drifted")
	}
	for _, path := range builder.staged {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("staging survived toolchain drift: %s", path)
		}
	}
	before.requireUnchanged(t, e, "toolchain drift")
}

func TestGlobalToolchainDriftAfterTheFinalBuildPreservesGlobalScope(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-g")
	e.buildSkill("build-skill", "alpha")
	if _, err := GlobalInit(e.home); err != nil {
		t.Fatal(err)
	}
	e.write(GlobalRoot(e.home), "Skillfile.json", `{
		"schema_version": 1, "agents": ["claude_code"],
		"skills": [{"name": "skill-g", "tag": "v1"}]}`)
	userHome := t.TempDir()
	baseline, _, _, _ := newFakeDeps(t)
	if result := Global(e.cfg, userHome, Options{Platform: "unix", Build: baseline}); result.Status != "ok" {
		t.Fatalf("baseline global install failed: %+v", result)
	}
	e.write(GlobalRoot(e.home), "Skillfile.json", `{
		"schema_version": 1, "agents": ["claude_code"],
		"skills": [{"name": "skill-g", "tag": "v1"}, {"name": "build-skill", "tag": "v1"}]}`)
	e.seedLiveCache()
	before := captureLiveState(t, e)

	deps, toolchain, _, builder := newFakeDeps(t)
	toolchain.verifyErr = errors.New("toolchain tree changed during operation")
	handoffs := 0

	result := Global(e.cfg, userHome, Options{Platform: "unix", Build: deps, OnStaged: func(Staged) error {
		handoffs++
		return nil
	}})
	if result.Status != "failed" {
		t.Fatalf("global install status = %q, want failed on toolchain drift", result.Status)
	}
	if !containsSubstring(result.Errors, "toolchain tree changed during operation") {
		t.Fatalf("global errors do not report the drift: %v", result.Errors)
	}
	if handoffs != 0 {
		t.Fatal("global staged outputs were handed off after the toolchain drifted")
	}
	if toolchain.session == nil || toolchain.session.verified != 1 {
		t.Fatalf("global session verification count = %+v, want exactly one pre-handoff check", toolchain.session)
	}
	if len(builder.calls) != 1 {
		t.Fatalf("global builder calls = %v, want the drift to be caught after the build", builder.calls)
	}
	before.requireUnchanged(t, e, "global toolchain drift")
}

func TestCacheHitOnlyPlanStillFinalizesToolchainTrust(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	e.declare("build-skill")
	deps, toolchain, cache, builder := newFakeDeps(t)
	cache.byCommand["alpha"] = buildcache.Result{Status: buildcache.Hit, ArtifactPath: "/protected/bin/alpha"}
	toolchain.verifyErr = errors.New("toolchain tree changed during operation")

	result := e.install(Options{Build: deps})
	if result.Status != "failed" {
		t.Fatalf("install status = %q, want failed even when every command was a cache hit", result.Status)
	}
	if len(builder.calls) != 0 {
		t.Fatalf("a cache-hit-only plan compiled something: %v", builder.calls)
	}
	if !containsSubstring(result.Errors, "toolchain tree changed during operation") {
		t.Fatalf("errors do not report the drift: %v", result.Errors)
	}
	requireAbsent(t, e.persistentPaths())
}

func TestCacheInspectionIsBracketedByTheFrozenSource(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha", "beta")
	e.declare("build-skill")
	deps, _, cache, builder := newFakeDeps(t)
	// The lookup returns a reusable entry for identity A while the snapshot
	// changes underneath it. Bracketing rejects that decision inside the
	// read-only planning phase, so the second command is never planned and the
	// miss is never compiled.
	cache.byCommand["alpha"] = buildcache.Result{Status: buildcache.Hit, ArtifactPath: "/protected/bin/alpha"}
	cache.observe = func(buildcache.Expectation) { e.mutateSnapshot("build-skill") }
	handoffs := 0

	result := e.install(Options{Build: deps, OnStaged: func(Staged) error {
		handoffs++
		return nil
	}})
	if result.Status != "failed" {
		t.Fatalf("install status = %q, want failed when the source changes during cache inspection", result.Status)
	}
	if !containsSubstring(result.Errors, "build-source snapshot mutated") {
		t.Fatalf("errors do not report the mutated source: %v", result.Errors)
	}
	if !containsSubstring(result.Errors, "build-skill.alpha") {
		t.Fatalf("errors do not name the command whose lookup raced the source: %v", result.Errors)
	}
	if len(cache.inspected) != 1 {
		t.Fatalf("cache inspections = %v, want planning to stop at the first raced lookup", cache.inspected)
	}
	if len(result.Builds) != 0 {
		t.Fatalf("plan reported %d commands, want none accepted after a raced lookup", len(result.Builds))
	}
	if len(builder.calls) != 0 || handoffs != 0 {
		t.Fatalf("planning proceeded past a mutated source: builds=%v handoffs=%d", builder.calls, handoffs)
	}
	requireAbsent(t, e.persistentPaths())
}

func TestSourceMutationDuringStagingBlocksHandoffOfACacheHit(t *testing.T) {
	e := newEnv(t)
	e.skill("script-skill")
	e.buildSkill("build-skill", "alpha", "beta")
	e.declare("script-skill")
	before := e.installBaseline()

	e.declare("script-skill", "build-skill")
	deps, toolchain, cache, builder := newFakeDeps(t)
	// alpha is reused from the protected cache; beta is compiled, and the
	// snapshot both decisions rest on changes while beta is building.
	cache.byCommand["alpha"] = buildcache.Result{Status: buildcache.Hit, ArtifactPath: "/protected/bin/alpha"}
	builder.observe = func(request StageRequest) {
		if request.Command == "beta" {
			e.mutateSnapshot("build-skill")
		}
	}
	handoffs := 0

	result := e.install(Options{Build: deps, OnStaged: func(Staged) error {
		handoffs++
		return nil
	}})
	if result.Status != "failed" {
		t.Fatalf("install status = %q, want failed when a planned source changes during staging", result.Status)
	}
	if !containsSubstring(result.Errors, "build-source snapshot mutated") {
		t.Fatalf("errors do not report the mutated source: %v", result.Errors)
	}
	if !containsSubstring(result.Errors, "build-skill") {
		t.Fatalf("errors do not name the mutated node: %v", result.Errors)
	}
	if len(builder.calls) != 1 || builder.calls[0] != "beta" {
		t.Fatalf("builder calls = %v, want only the miss [beta]", builder.calls)
	}
	if handoffs != 0 {
		t.Fatal("staged outputs were handed off after a planned source changed")
	}
	if toolchain.session == nil || toolchain.session.verified != 1 {
		t.Fatalf("session verification count = %+v, want exactly one pre-handoff check", toolchain.session)
	}
	before.requireUnchanged(t, e, "source mutation during staging")
}

func TestToolchainFailureBlocksEveryPersistentMutation(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	e.declare("build-skill")
	deps, toolchain, cache, builder := newFakeDeps(t)
	toolchain.establishErr = errors.New("trusted toolchain unavailable")

	result := e.install(Options{Build: deps})
	if result.Status != "failed" {
		t.Fatalf("install status = %q, want failed", result.Status)
	}
	if len(cache.inspected) != 0 {
		t.Fatalf("cache inspected before a trusted toolchain: %v", cache.inspected)
	}
	if len(builder.calls) != 0 {
		t.Fatalf("builder ran without a trusted toolchain: %v", builder.calls)
	}
	requireAbsent(t, e.persistentPaths())
}

func TestCorruptCacheEntryFailsBeforeAnyPersistentMutation(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	e.declare("build-skill")
	deps, _, cache, builder := newFakeDeps(t)
	cache.byCommand["alpha"] = buildcache.Result{Status: buildcache.Corrupt, Reason: "receipt is not canonical"}

	result := e.install(Options{Build: deps})
	if result.Status != "failed" {
		t.Fatalf("install status = %q, want failed", result.Status)
	}
	if len(builder.calls) != 0 {
		t.Fatalf("builder ran for a corrupt cache entry: %v", builder.calls)
	}
	if len(result.Builds) != 1 || result.Builds[0].Outcome() != BuildCorrupt {
		t.Fatalf("plan did not report the corrupt outcome: %+v", result.Builds)
	}
	if !containsSubstring(result.Errors, "receipt is not canonical") {
		t.Fatalf("errors do not explain the corrupt entry: %v", result.Errors)
	}
	requireAbsent(t, e.persistentPaths())
}

func TestUnsupportedCacheProtectionFailsClosed(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	e.declare("build-skill")
	deps, _, cache, builder := newFakeDeps(t)
	cache.byCommand["alpha"] = buildcache.Result{
		Status: buildcache.Unsupported, Reason: "platform protection is unavailable",
	}

	result := e.install(Options{Build: deps})
	if result.Status != "failed" {
		t.Fatalf("install status = %q, want failed", result.Status)
	}
	if len(builder.calls) != 0 {
		t.Fatalf("builder ran without protected cache support: %v", builder.calls)
	}
	if result.Builds[0].Outcome() != BuildUnsupported {
		t.Fatalf("outcome = %q, want %q", result.Builds[0].Outcome(), BuildUnsupported)
	}
	requireAbsent(t, e.persistentPaths())
}

func TestUntrustedCacheEntryIsRebuiltAndNeverReused(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	e.declare("build-skill")
	deps, _, cache, builder := newFakeDeps(t)
	cache.byCommand["alpha"] = buildcache.Result{
		Status:       buildcache.UntrustedProvenance,
		Reason:       "cache entry is group-writable",
		ArtifactPath: "/protected/bin/alpha",
	}

	result := e.install(Options{Build: deps})
	if result.Status != "ok" {
		t.Fatalf("install failed: %+v", result)
	}
	if result.Builds[0].Outcome() != BuildWouldRebuildUntrustedCache {
		t.Fatalf("outcome = %q, want %q", result.Builds[0].Outcome(), BuildWouldRebuildUntrustedCache)
	}
	if result.Builds[0].ArtifactPath() != "" {
		t.Fatalf("untrusted entry exposed a reusable artifact path %q", result.Builds[0].ArtifactPath())
	}
	if len(builder.calls) != 1 || builder.calls[0] != "alpha" {
		t.Fatalf("untrusted entry was not rebuilt: %v", builder.calls)
	}
}

func TestScriptOnlyInstallPerformsNoToolchainOrCacheWork(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")
	deps, toolchain, cache, builder := newFakeDeps(t)

	result := e.install(Options{Build: deps})
	if result.Status != "ok" {
		t.Fatalf("install failed: %+v", result)
	}
	if toolchain.probes != 0 || toolchain.establishes != 0 {
		t.Fatalf("script-only install touched the toolchain: probes=%d establishes=%d",
			toolchain.probes, toolchain.establishes)
	}
	if len(cache.inspected) != 0 || len(builder.calls) != 0 {
		t.Fatalf("script-only install touched the build cache or builder: %v %v", cache.inspected, builder.calls)
	}
	if len(result.Builds) != 0 || len(result.Staged) != 0 {
		t.Fatalf("script-only install reported build work: %+v %+v", result.Builds, result.Staged)
	}
}

func TestInjectedClockAndGenerationReaderDriveInstallMarkers(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")
	deps, _, _, _ := newFakeDeps(t)
	generation := &countingGeneration{}
	deps.Generation = generation
	deps.Clock = fixedClock{at: time.Date(2031, 3, 4, 5, 6, 7, 0, time.UTC)}

	if result := e.install(Options{Build: deps}); result.Status != "ok" {
		t.Fatalf("install failed: %+v", result)
	}
	installed := marker.Read(filepath.Join(e.project, ".agents", "skills", "skill-a"))
	if installed == nil {
		t.Fatal("install marker is missing")
	}
	if installed.InstalledAt != "2031-03-04T05:06:07Z" {
		t.Fatalf("installed_at = %q, want the injected clock value", installed.InstalledAt)
	}
	if generation.reads == 0 {
		t.Fatal("the moved-tag gate did not read installation state through the injected reader")
	}
}

func TestPlannedBuildAccessorsAreImmutable(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	e.declare("build-skill")
	deps, _, _, _ := newFakeDeps(t)

	result := e.install(Options{DryRun: true, Build: deps})
	if result.Status != "ok" {
		t.Fatalf("dry-run failed: %+v", result)
	}
	planned := result.Builds[0]
	target := planned.Target()
	target.Tuning["GOAMD64"] = "tampered"
	if planned.Target().Tuning["GOAMD64"] != "v1" {
		t.Fatal("PlannedBuild.Target exposed the plan's own tuning map")
	}

	// Builds returns an independent slice on every call.
	plan := BuildPlan{scope: "test", builds: []PlannedBuild{planned}}
	copied := plan.Builds()
	copied[0] = PlannedBuild{}
	if plan.Builds()[0].Command() != "alpha" {
		t.Fatal("BuildPlan.Builds shares storage with the plan")
	}
}

// TestReleaseTakesNoToolchainVerdictAfterLiveMutation pins the post-handoff
// boundary. Releasing the plan is deferred across every live write, so a
// verification taken there would report drift against state the run can no
// longer preserve. The fake fails every check after the first, so any re-check
// during teardown fails an installation that had already committed.
func TestReleaseTakesNoToolchainVerdictAfterLiveMutation(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	e.declare("build-skill")
	deps, toolchain, _, builder := newFakeDeps(t)
	toolchain.lateVerifyErr = errors.New("toolchain tree changed during operation")

	result := e.install(Options{Build: deps})
	if result.Status != "ok" {
		t.Fatalf("install status = %q, want ok: %+v", result.Status, result)
	}
	if containsSubstring(result.Errors, "toolchain tree changed during operation") {
		t.Fatalf("teardown reported a toolchain verdict after live mutation: %v", result.Errors)
	}
	session := toolchain.session
	if session == nil || session.verified != 1 {
		t.Fatalf("session verification count = %+v, want exactly one pre-handoff check", session)
	}
	if session.released != 1 {
		t.Fatalf("session release count = %d, want exactly one release", session.released)
	}
	// The run really did commit, so the pre-handoff check was the last point at
	// which drift could still be reported for free.
	if _, err := os.Stat(filepath.Join(e.project, ".agents", "skills", "build-skill")); err != nil {
		t.Fatalf("installation did not commit: %v", err)
	}
	for _, path := range builder.staged {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("staged artifact %s survived the release: %v", path, err)
		}
	}
}

// TestGlobalReleaseTakesNoToolchainVerdictAfterLiveMutation is the global-scope
// half of the same boundary.
func TestGlobalReleaseTakesNoToolchainVerdictAfterLiveMutation(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	if _, err := GlobalInit(e.home); err != nil {
		t.Fatal(err)
	}
	e.write(GlobalRoot(e.home), "Skillfile.json", `{
		"schema_version": 1, "agents": ["claude_code"],
		"skills": [{"name": "build-skill", "tag": "v1"}]}`)
	userHome := t.TempDir()
	deps, toolchain, _, builder := newFakeDeps(t)
	toolchain.lateVerifyErr = errors.New("toolchain tree changed during operation")

	result := Global(e.cfg, userHome, Options{Platform: "unix", Build: deps})
	if result.Status != "ok" {
		t.Fatalf("global install status = %q, want ok: %+v", result.Status, result)
	}
	if containsSubstring(result.Errors, "toolchain tree changed during operation") {
		t.Fatalf("global teardown reported a toolchain verdict after live mutation: %v", result.Errors)
	}
	session := toolchain.session
	if session == nil || session.verified != 1 {
		t.Fatalf("global session verification count = %+v, want exactly one pre-handoff check", session)
	}
	if session.released != 1 {
		t.Fatalf("global session release count = %d, want exactly one release", session.released)
	}
	if _, err := os.Stat(filepath.Join(GlobalRoot(e.home), "skills", "build-skill")); err != nil {
		t.Fatalf("global installation did not commit: %v", err)
	}
	for _, path := range builder.staged {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("global staged artifact %s survived the release: %v", path, err)
		}
	}
}

// TestSessionReleaseFailureWarnsWithoutFailingACommittedInstall covers the only
// failure teardown can still produce: a private root it could not remove. It is
// reported, but it does not retract an installation that already succeeded.
func TestSessionReleaseFailureWarnsWithoutFailingACommittedInstall(t *testing.T) {
	e := newEnv(t)
	e.buildSkill("build-skill", "alpha")
	e.declare("build-skill")
	deps, toolchain, _, _ := newFakeDeps(t)
	toolchain.releaseErr = errors.New("cannot remove operation-private probe")

	result := e.install(Options{Build: deps})
	if result.Status != "ok" {
		t.Fatalf("install status = %q, want ok: %+v", result.Status, result)
	}
	if !containsSubstring(result.Messages, "cannot remove operation-private probe") {
		t.Fatalf("release failure was swallowed: %v", result.Messages)
	}
	if len(result.Errors) != 0 {
		t.Fatalf("a teardown failure became an install error: %v", result.Errors)
	}
	if _, err := os.Stat(filepath.Join(e.project, ".agents", "skills", "build-skill")); err != nil {
		t.Fatalf("installation did not commit: %v", err)
	}
}

func TestDefaultToolchainProbeRemovesItsProbeRootOnFailure(t *testing.T) {
	base := t.TempDir()
	t.Setenv("TMPDIR", base)
	t.Setenv("CURATOR_GO", filepath.Join(base, "absent-go"))
	private, toolchain := defaultToolchain(t)

	if _, _, err := toolchain.Probe(context.Background()); err == nil {
		t.Fatal("probe with an absent toolchain unexpectedly succeeded")
	}
	requireNoToolchainBases(t, private, "a failed probe")
	requireOnlyOperationPrivateRoots(t, base)
}

// TestDefaultToolchainProbeRemovesItsProbeRootOnSuccess is the success half of
// the same guarantee: a probe that resolves a real trusted toolchain still owns
// no private state once it returns. Probing runs go version and go env only, so
// it needs no worker re-execution and stays valid in this test binary.
func TestDefaultToolchainProbeRemovesItsProbeRootOnSuccess(t *testing.T) {
	base := t.TempDir()
	t.Setenv("TMPDIR", base)
	private, toolchain := defaultToolchain(t)

	target, resolved, err := toolchain.Probe(context.Background())
	if err != nil {
		t.Skipf("no trusted Go toolchain is resolvable here: %v", err)
	}
	if target.GOOS == "" || resolved.GoVersion == "" {
		t.Fatalf("probe returned an incomplete identity: target=%+v toolchain=%+v", target, resolved)
	}
	requireNoToolchainBases(t, private, "a successful probe")
	requireOnlyOperationPrivateRoots(t, base)
}

func TestDefaultToolchainEstablishRemovesItsPrivateRootOnFailure(t *testing.T) {
	base := t.TempDir()
	t.Setenv("TMPDIR", base)
	t.Setenv("CURATOR_GO", filepath.Join(base, "absent-go"))
	private, toolchain := defaultToolchain(t)

	if _, err := toolchain.Establish(context.Background()); err == nil {
		t.Fatal("establish with an absent toolchain unexpectedly succeeded")
	}
	requireNoToolchainBases(t, private, "a failed session")
	requireOnlyOperationPrivateRoots(t, base)
}

// defaultToolchain builds the real trusted-toolchain boundary over a private
// root the test owns, and removes that root when the test ends.
func defaultToolchain(t *testing.T) (*privateRoot, *goToolchain) {
	t.Helper()
	private := &privateRoot{prefix: operationPrivatePrefix}
	t.Cleanup(func() {
		if err := private.remove(); err != nil {
			t.Errorf("remove the operation-private root: %v", err)
		}
	})
	return private, &goToolchain{private: private}
}

// requireNoToolchainBases asserts the toolchain left no base inside the
// operation-private root, on the success and the failure path alike.
func requireNoToolchainBases(t *testing.T, private *privateRoot, context string) {
	t.Helper()
	if !private.created {
		return
	}
	entries, err := os.ReadDir(private.path)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "go-probe-base-") || strings.HasPrefix(entry.Name(), "go-build-base-") {
			t.Fatalf("toolchain base %s survived %s", entry.Name(), context)
		}
	}
}

func containsLine(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsSubstring(values []string, want string) bool {
	for _, value := range values {
		if strings.Contains(value, want) {
			return true
		}
	}
	return false
}
