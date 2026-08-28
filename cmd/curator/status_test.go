package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/adapters"
	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/scopes"
	"github.com/relux-works/curator/internal/testtoolchain"
)

const (
	cliHelperProcess = "CURATOR_TEST_CLI_HELPER"
	cliHelperConfig  = "CURATOR_TEST_CONFIG_PATH"
)

// TestMain gives the test binary the same fixed hidden worker mode the
// installed manager dispatches before command parsing, so a compiled
// installation in these tests runs the real identity-verified process graph
// instead of a mock.
func TestMain(m *testing.M) {
	if len(os.Args) == 2 && os.Args[1] == godriver.WorkerMode {
		os.Exit(godriver.RunWorker(os.Stdin, os.Stdout))
	}
	if os.Getenv(cliHelperProcess) == "1" {
		os.Exit(m.Run())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	lock, err := testtoolchain.AcquireHostGOROOT(ctx)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "acquire package host GOROOT test lock:", err)
		os.Exit(1)
	}
	code := m.Run()
	if lock != nil {
		if closeErr := lock.Close(); closeErr != nil {
			_, _ = fmt.Fprintln(os.Stderr, "release package host GOROOT test lock:", closeErr)
			code = 1
		}
	}
	os.Exit(code)
}

func TestCLIHelperProcess(t *testing.T) {
	if os.Getenv(cliHelperProcess) != "1" {
		t.Parallel()
		return
	}
	separator := -1
	for index, arg := range os.Args {
		if arg == "--" {
			separator = index
			break
		}
	}
	if separator < 0 {
		os.Exit(exitUsage)
	}
	os.Exit(run(os.Args[separator+1:], fileConfigSource(os.Getenv(cliHelperConfig)), os.Stdout, os.Stderr))
}

// capture runs one CLI invocation with a private config source and output
// buffers. It never mutates process-global environment or standard streams.
func capture(t *testing.T, configPath string, args ...string) (int, string, string) {
	t.Helper()
	userHome := filepath.Join(filepath.Dir(filepath.Dir(configPath)), "user")
	return captureWithUserHome(t, configPath, func() (string, error) { return userHome, nil }, args...)
}

func captureWithUserHome(t *testing.T, configPath string, userHome func() (string, error), args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr strings.Builder
	command := cli{config: fileConfigSource(configPath), stdout: &stdout, stderr: &stderr, userHome: userHome}
	code := command.run(args)
	return code, stdout.String(), stderr.String()
}

// captureWithEnv drives the real CLI core in an isolated helper process for
// cases whose contract is specifically about process environment selection.
func captureWithEnv(t *testing.T, configPath string, environment map[string]string, args ...string) (int, string, string) {
	t.Helper()
	commandArgs := []string{"-test.run=^TestCLIHelperProcess$", "--"}
	commandArgs = append(commandArgs, args...)
	command := exec.Command(os.Args[0], commandArgs...)
	overrides := map[string]string{cliHelperProcess: "1", cliHelperConfig: configPath}
	for key, value := range environment {
		overrides[key] = value
	}
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if _, replaced := overrides[key]; !replaced {
			command.Env = append(command.Env, entry)
		}
	}
	for key, value := range overrides {
		command.Env = append(command.Env, key+"="+value)
	}
	var stdout, stderr strings.Builder
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	code := exitOK
	if err != nil {
		var exitError *exec.ExitError
		if !errors.As(err, &exitError) {
			t.Fatal(err)
		}
		code = exitError.ExitCode()
	}
	return code, stdout.String(), stderr.String()
}

// statusPayload is the machine-readable status document one project produces.
type statusPayload struct {
	Alias  string            `json:"alias"`
	Path   string            `json:"path"`
	Skills map[string]string `json:"skills"`
	Builds []buildReport     `json:"builds"`
}

func decodeStatus(t *testing.T, payload string) statusPayload {
	t.Helper()
	var decoded statusPayload
	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		t.Fatalf("status --json is not one JSON object: %v\n%s", err, payload)
	}
	return decoded
}

// compiledProject creates a manager home, a skill repository that exports one
// schema 6 go-v1 build command from a vendored module, and a project that
// declares it. It returns the project root and the manager home.
func compiledProject(t *testing.T) (string, string) {
	t.Helper()
	return compiledProjectDeclaring(t, `{"name":"build-skill","tag":"v1"}`)
}

// compiledProjectDeclaring builds the same fixture but lets the caller choose
// what the project declares, so the transitively resolved case can declare a
// consumer instead of the compiled provider itself.
func compiledProjectDeclaring(t *testing.T, declarations string) (string, string) {
	t.Helper()
	root := t.TempDir()
	home := filepath.Join(root, "home")
	configPath := filepath.Join(home, "config.json")
	skillsRoot := filepath.Join(root, "skills")
	project := filepath.Join(root, "project")

	writeCompiledSkillRepo(t, filepath.Join(skillsRoot, "build-skill"))

	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, project, "init", "-q", "-b", "main")
	writeFile(t, filepath.Join(project, ".gitignore"), ".agents/\n.codex/skills/\nSkillfile.dev.json\n")
	writeFile(t, filepath.Join(project, manifest.Name),
		`{"schema_version":1,"agents":["codex_cli"],"skills":[`+declarations+`]}`)
	if err := config.Bootstrap(configPath, skillsRoot, "", []string{"codex_cli"}, false); err != nil {
		t.Fatal(err)
	}
	if err := config.AddProject(configPath, "app", project, []string{"codex_cli"}); err != nil {
		t.Fatal(err)
	}
	return project, home
}

// writeCompiledSkillRepo creates a skill repository that exports one schema 6
// go-v1 build command from a vendored module.
func writeCompiledSkillRepo(t *testing.T, skill string) {
	t.Helper()
	writeFile(t, filepath.Join(skill, "SKILL.md"), "---\nname: build-skill\ndescription: compiled command\n---\n"+
		"Resolve .agents/bin/build-tool.cmd, then global/bin/build-tool, then command -v build-tool or Get-Command build-tool.\n")
	writeFile(t, filepath.Join(skill, "assets", "prompt.md"), "prompt-visible asset\n")
	writeFile(t, filepath.Join(skill, "assets", "build-tool", "go.mod"), "module example.test/build-tool\n\ngo 1.23\n")
	writeFile(t, filepath.Join(skill, "assets", "build-tool", "vendor", "modules.txt"), "")
	writeFile(t, filepath.Join(skill, "assets", "build-tool", "cmd", "tool", "main.go"),
		"package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"curator\") }\n")
	writeFile(t, filepath.Join(skill, "agent-skill.json"), `{
		"schema_version": 6,
		"build_roots": ["assets/build-tool"],
		"capabilities": {},
		"commands": {
			"build-tool": {"type":"build","driver":"go-v1","source_dir":"assets/build-tool/cmd/tool"}
		}
	}`)
	runGit(t, skill, "init", "-q", "-b", "main")
	runGit(t, skill, "add", ".")
	runGit(t, skill, "commit", "-qm", "initial build skill")
	runGit(t, skill, "tag", "v1")
}

// legacyProject creates a project that declares one script-only skill, so a
// test can pin the pre-compiled-command behaviour of the same surfaces.
func legacyProject(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	configPath := filepath.Join(root, "home", "config.json")
	skillsRoot := filepath.Join(root, "skills")
	project := filepath.Join(root, "project")

	skill := filepath.Join(skillsRoot, "skill-a")
	writeFile(t, filepath.Join(skill, "SKILL.md"), "---\nname: skill-a\n---\n# Skill\n")
	runGit(t, skill, "init", "-q", "-b", "main")
	runGit(t, skill, "add", ".")
	runGit(t, skill, "commit", "-qm", "initial skill")
	runGit(t, skill, "tag", "v1")

	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit(t, project, "init", "-q", "-b", "main")
	writeFile(t, filepath.Join(project, ".gitignore"), ".agents/\n.codex/skills/\nSkillfile.dev.json\n")
	writeFile(t, filepath.Join(project, manifest.Name),
		`{"schema_version":1,"agents":["codex_cli"],"skills":[{"name":"skill-a","tag":"v1"}]}`)
	if err := config.Bootstrap(configPath, skillsRoot, "", []string{"codex_cli"}, false); err != nil {
		t.Fatal(err)
	}
	if err := config.AddProject(configPath, "app", project, []string{"codex_cli"}); err != nil {
		t.Fatal(err)
	}
	return project, filepath.Dir(configPath)
}

// reinstall runs the ordinary reconciliation path, which is also the only
// repair path Curator has.
func reinstall(t *testing.T, home string) {
	t.Helper()
	if code, stdout, stderr := capture(t, filepath.Join(home, "config.json"), "install", "app"); code != exitOK {
		t.Fatalf("restoring install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// rewriteMarker edits one installed marker in place. The install marker is
// excluded from the installed content hash, so this isolates compiled-state
// drift from ordinary content drift.
func rewriteMarker(t *testing.T, installed string, edit func(map[string]any)) {
	t.Helper()
	path := filepath.Join(installed, marker.Name)
	payload, err := os.ReadFile(path) // #nosec G304 -- test fixture path
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil {
		t.Fatal(err)
	}
	edit(object)
	rewritten, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(rewritten, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	if marker.Read(installed) == nil {
		t.Fatalf("the rewritten marker is no longer a valid marker:\n%s", rewritten)
	}
}

// refuseMarker replaces one installed marker with bytes the marker reader must
// refuse. It is the counterpart of rewriteMarker: it proves the refusal itself
// still produces a stable, distinguishable currentness code.
func refuseMarker(t *testing.T, installed, payload string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(installed, marker.Name), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	if marker.Read(installed) != nil {
		t.Fatalf("the marker reader accepted the payload, so the case proves nothing:\n%s", payload)
	}
}

// markerPayload reads one installed marker as a generic JSON object.
func markerPayload(t *testing.T, installed string) map[string]any {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join(installed, marker.Name)) // #nosec G304 -- test fixture path
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil {
		t.Fatal(err)
	}
	return object
}

// corruptCacheReceipt rewrites the canonical receipt of every protected entry,
// so the entry can no longer be interpreted at all.
func corruptCacheReceipt(t *testing.T, home string) {
	t.Helper()
	for _, entry := range cacheEntries(t, home) {
		receipt := filepath.Join(entry, buildcache.ReceiptFilename)
		if err := os.Chmod(receipt, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(receipt, []byte(`{"schema_version":1,"not":"a receipt"}`), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

// corruptCacheArtifact rewrites the published artifact bytes of every protected
// entry, so the entry no longer matches its own receipt.
func corruptCacheArtifact(t *testing.T, home string) {
	t.Helper()
	for _, entry := range cacheEntries(t, home) {
		matches, err := filepath.Glob(filepath.Join(entry, "bin", "*"))
		if err != nil || len(matches) != 1 {
			t.Fatalf("protected entry %s does not hold exactly one artifact: %v %v", entry, matches, err)
		}
		if err := os.Chmod(matches[0], 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(matches[0], []byte("not the artifact this receipt describes"), 0o700); err != nil {
			t.Fatal(err)
		}
	}
}

// cacheFingerprint digests every protected entry so a test can prove a refused
// run left the live build cache byte-for-byte unchanged.
func cacheFingerprint(t *testing.T, home string) string {
	t.Helper()
	digest := sha256.New()
	base := filepath.Join(home, "cache", "build")
	err := filepath.WalkDir(base, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, relErr := filepath.Rel(base, path)
		if relErr != nil {
			return relErr
		}
		digest.Write([]byte(relative + "\n"))
		if entry.IsDir() {
			return nil
		}
		payload, readErr := os.ReadFile(path) // #nosec G304 -- test fixture path
		if readErr != nil {
			return readErr
		}
		digest.Write(payload)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(digest.Sum(nil))
}

// requireNativeControlInventoryPlatform carves a case that installs a real
// compiled command out of a host rc5-native-control-inventory-v1 defines no
// record for.
//
// A compiled command is produced by the go-v1 driver, and on such a host the
// driver refuses before any worker starts: the portable execution policy is
// specified for macOS and Windows only. A case that needs a completed
// compilation has nothing to assert there, and asserting one anyway would claim
// execution the protocol does not grant.
//
// The carve-out is never taken on trust. The refusal is asserted positively on
// this very runner by TestCompiledInstallFollowsTheNativeControlInventoryExactly,
// exactly as internal/godriver's own platform exclusion is asserted by
// TestProbeRejectsAnUncoveredPlatformBeforeTheWorker.
//
// The predicate is read from godriver.InventoryPlatform rather than written as
// a GOOS list here, so it cannot drift from the inventory it stands for. Linux
// is the platform the rc.5 qualification vector still marks excluded, with
// until_task TASK-260728-1skseh; when the inventory gains a record for a host,
// these cases run there without an edit to this helper.
func requireNativeControlInventoryPlatform(t *testing.T) {
	t.Helper()
	if godriver.InventoryPlatform(runtime.GOOS) == "" {
		t.Skipf("rc5-native-control-inventory-v1 defines no record for host %s; "+
			"the portable execution policy is specified for macOS and Windows only", runtime.GOOS)
	}
}

// publishedCacheEntries counts protected build-cache entries without requiring
// the cache root to exist. A refused operation publishes nothing, so on a host
// the inventory does not cover the root itself is legitimately absent.
func publishedCacheEntries(t *testing.T, home string) []string {
	t.Helper()
	base := filepath.Join(home, "cache", "build", "go-v1")
	entries, err := os.ReadDir(base)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatalf("protected build cache is unreadable: %v", err)
	}
	var paths []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			paths = append(paths, filepath.Join(base, entry.Name()))
		}
	}
	return paths
}

func cacheEntries(t *testing.T, home string) []string {
	t.Helper()
	base := filepath.Join(home, "cache", "build", "go-v1")
	entries, err := os.ReadDir(base)
	if err != nil {
		t.Fatalf("protected build cache is missing: %v", err)
	}
	var paths []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			paths = append(paths, filepath.Join(base, entry.Name()))
		}
	}
	return paths
}

// compiledProjectFixture is one real compiled installation: the project the
// operator declared, the manager home that owns the protected build cache, and
// the installed skill whose marker records the compiled state.
type compiledProjectFixture struct {
	project   string
	home      string
	installed string
}

// newInstalledCompiledProject builds that fixture and installs it once, through
// the real command path.
//
// Every surface below starts from the same thing: one installed compiled
// command whose protected entry and install marker agree. Deriving it costs a
// real compilation of the vendored module, so it is derived once and handed to
// each case rather than rebuilt per test.
func newInstalledCompiledProject(t *testing.T) compiledProjectFixture {
	t.Helper()
	project, home := compiledProject(t)
	if code, stdout, stderr := capture(t, filepath.Join(home, "config.json"), "install", "app"); code != exitOK {
		t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	return compiledProjectFixture{
		project:   project,
		home:      home,
		installed: filepath.Join(project, ".agents", "skills", "build-skill"),
	}
}

// TestCompiledProjectStatusRepairRollbackRecovery drives the whole compiled
// project lifecycle over one installed fixture: currentness reporting, corrupt
// cache repair by install and by upgrade, the rollback and recovery of a commit
// that failed after a real publication, protected cache state that moves during
// a check, and the repair of cache bytes outside a provable boundary.
//
// Each scenario owns an installed fixture so the expensive compiler sessions
// can run concurrently without sharing mutable marker or cache state.
func TestCompiledProjectStatusAndUntrustedRecovery(t *testing.T) {
	t.Parallel()
	requireNativeControlInventoryPlatform(t)
	fixture := newInstalledCompiledProject(t)
	assertCompiledCurrentnessAndFailedCheck(t, fixture)
	assertProtectedCacheStateThatMovedDuringTheCheck(t, fixture)
	assertUntrustedCompiledStateIsRepaired(t, fixture)
}

func TestCompiledProjectRepairsCorruptCompiledState(t *testing.T) {
	t.Parallel()
	requireNativeControlInventoryPlatform(t)
	for _, command := range []string{"install", "upgrade"} {
		t.Run(command, func(t *testing.T) {
			t.Parallel()
			fixture := newInstalledCompiledProject(t)
			for _, corrupt := range []func(t *testing.T, home string){corruptCacheReceipt, corruptCacheArtifact} {
				assertCorruptCompiledStateIsRepaired(t, fixture, command, corrupt)
			}
		})
	}
}

func TestCompiledProjectRestoresCacheWhenCommitFails(t *testing.T) {
	t.Parallel()
	requireNativeControlInventoryPlatform(t)
	for _, command := range []string{"install", "upgrade"} {
		t.Run(command, func(t *testing.T) {
			t.Parallel()
			assertTheCacheIsRestoredWhenTheCommitFails(t, newInstalledCompiledProject(t), command)
		})
	}
}

// assertCompiledCurrentnessAndFailedCheck is the end-to-end proof of the
// compiled currentness surface: a real installation reports every planned
// command as current, and each independent way that state can stop being
// current produces its own stable code and a non-zero `status --check`.
func assertCompiledCurrentnessAndFailedCheck(t *testing.T, fixture compiledProjectFixture) {
	project, home, installed := fixture.project, fixture.home, fixture.installed

	code, stdout, stderr := capture(t, filepath.Join(home, "config.json"), "status", "app", "--json")
	if code != exitOK {
		t.Fatalf("status --json = %d\nstderr:\n%s", code, stderr)
	}
	report := decodeStatus(t, stdout)
	if len(report.Builds) != 1 {
		t.Fatalf("status reported %d compiled commands, want 1:\n%s", len(report.Builds), stdout)
	}
	current := report.Builds[0]
	if current.State != buildCurrent || current.CacheOutcome != "cache-hit" {
		t.Fatalf("current build row = %+v", current)
	}
	// The artifact name is the target's, not this host's spelling of it: a
	// windows build reports bin/build-tool.exe and a unix one bin/build-tool.
	wantArtifact, err := buildmeta.ArtifactPath("build-tool", runtime.GOOS)
	if err != nil {
		t.Fatal(err)
	}
	if current.Skill != "build-skill" || current.Command != "build-tool" ||
		current.Driver != "go-v1" || current.BuildRoot != "assets/build-tool" ||
		current.SourceDir != "assets/build-tool/cmd/tool" || current.CacheKey == "" ||
		current.ArtifactPath != wantArtifact || current.Target == "" ||
		current.BuildSource.ContentSHA256 == "" {
		t.Fatalf("current build row does not report the full planned command: %+v", current)
	}
	if report.Skills["build-skill"] != stateUpToDate {
		t.Fatalf("skills = %v", report.Skills)
	}
	// No manager-private location may reach the machine-readable surface. The
	// project path is the operator's own argument and stays.
	for _, forbidden := range []string{home, filepath.Join(project, ".agents", "bin")} {
		if strings.Contains(marshal(t, report.Builds), forbidden) {
			t.Fatalf("build rows published the private location %q:\n%s", forbidden, stdout)
		}
	}
	if code, _, _ := capture(t, filepath.Join(home, "config.json"), "status", "app", "--check"); code != exitOK {
		t.Fatalf("clean compiled status --check = %d, want %d", code, exitOK)
	}

	for name, testCase := range map[string]struct {
		tamper  func(t *testing.T)
		restore func(t *testing.T)
		want    string
		cause   string
		outcome string
		// skillState is the code the declared skill is demoted to. It defaults
		// to the build state, because compiled state demotes the skill it
		// belongs to.
		skillState string
		// snapshotCache copies the live protected cache aside and puts every
		// byte and permission bit back, which is how a case that damages cache
		// state resets its own fixture. Repairing a reporting fixture by
		// reinstalling compiles the command again and proves nothing this test
		// owns; install, repair, and rollback keep their own dedicated tests.
		snapshotCache bool
		// humanCLI additionally runs the whole plain-text command path, so the
		// rendering asserted for every case is pinned to the line the command
		// really prints for at least one case of each shape.
		humanCLI bool
	}{
		"artifact hash recorded by the marker no longer matches the entry": {
			tamper: func(t *testing.T) {
				rewriteMarker(t, installed, func(object map[string]any) {
					builds := object["builds"].(map[string]any)
					build := builds["build-tool"].(map[string]any)
					build["artifact_sha256"] = testDigest(5)
				})
			},
			// One no-cause marker drift also proves the whole plain-text path.
			want: buildArtifactDrift, outcome: "cache-hit", humanCLI: true,
		},
		"receipt recorded by the marker no longer matches the entry": {
			tamper: func(t *testing.T) {
				rewriteMarker(t, installed, func(object map[string]any) {
					builds := object["builds"].(map[string]any)
					build := builds["build-tool"].(map[string]any)
					build["receipt_sha256"] = testDigest(5)
				})
			},
			want: buildCorruptReceipt, outcome: "cache-hit",
		},
		"logical key recorded by the marker was derived from another build input": {
			tamper: func(t *testing.T) {
				rewriteMarker(t, installed, func(object map[string]any) {
					builds := object["builds"].(map[string]any)
					build := builds["build-tool"].(map[string]any)
					build["cache_key"] = testDigest(5)
				})
			},
			// One cause-bearing drift also proves the whole plain-text path.
			want: buildInputDrift, cause: causeUnattributed, outcome: "cache-hit", humanCLI: true,
		},
		"logical key was derived under a build root the marker does not record": {
			tamper: func(t *testing.T) {
				rewriteMarker(t, installed, func(object map[string]any) {
					object["build_roots"] = []any{"assets/other-tool"}
					builds := object["builds"].(map[string]any)
					build := builds["build-tool"].(map[string]any)
					build["cache_key"] = testDigest(5)
				})
			},
			want: buildInputDrift, cause: causeBuildRoot, outcome: "cache-hit",
		},
		"recorded artifact is not the one this target derives": {
			tamper: func(t *testing.T) {
				rewriteMarker(t, installed, func(object map[string]any) {
					builds := object["builds"].(map[string]any)
					build := builds["build-tool"].(map[string]any)
					build["cache_key"] = testDigest(5)
					build["artifact_path"] = foreignArtifactPath()
				})
			},
			want: buildInputDrift, cause: causeTarget, outcome: "cache-hit",
		},
		"recorded build-source identity no longer matches the frozen snapshot": {
			tamper: func(t *testing.T) {
				rewriteMarker(t, installed, func(object map[string]any) {
					source := object["build_source"].(map[string]any)
					source["content_sha256"] = testDigest(5)
				})
			},
			want: buildSourceDrift, outcome: "cache-hit",
		},
		"marker records no build for the command the closure activates": {
			tamper: func(t *testing.T) {
				rewriteMarker(t, installed, func(object map[string]any) {
					object["builds"] = map[string]any{}
					delete(object, "build_source")
				})
			},
			want: buildCommandDrift, outcome: "cache-hit",
		},
		// The two cases below damage the shared protected cache rather than the
		// marker, so each restores it from its own byte-and-mode snapshot before
		// the next case runs.
		"protected cache entry cannot be interpreted": {
			tamper:        func(t *testing.T) { corruptCacheArtifact(t, home) },
			snapshotCache: true,
			want:          buildCorruptCache, outcome: "corrupt",
		},
		"protected cache holds no entry for the recorded key": {
			tamper: func(t *testing.T) {
				for _, entry := range cacheEntries(t, home) {
					if err := os.RemoveAll(entry); err != nil {
						t.Fatal(err)
					}
				}
			},
			snapshotCache: true,
			want:          buildMissingArtifact, outcome: "would-preflight-and-build",
		},
		// The unreadable schema is one past the whole readable band, not one
		// past the written one. Schemas 3 and 4 are read by this manager, so
		// pinning `written + 1` silently stopped testing the unsupported band
		// the moment a newer schema became readable -- and asserted the
		// opposite of the spec while doing it.
		"marker schema cannot be read by this manager": {
			tamper: func(t *testing.T) {
				object := markerPayload(t, installed)
				object["schema_version"] = marker.NewestSchemaVersion + 1
				refuseMarker(t, installed, marshal(t, object))
			},
			want: stateUnsupportedMarker, outcome: "cache-hit",
		},
		// A readable schema whose document is not a valid marker must be
		// reported as an invalid document, never as one from a newer manager:
		// "upgrade the manager" is not the remedy and no upgrade exists.
		"marker at a readable schema is still not a marker document": {
			tamper: func(t *testing.T) {
				object := markerPayload(t, installed)
				object["schema_version"] = marker.NewestSchemaVersion
				delete(object, "commit")
				refuseMarker(t, installed, marshal(t, object))
			},
			want: stateInvalidMarker, outcome: "cache-hit",
		},
		"marker records a build driver outside the closed set": {
			tamper: func(t *testing.T) {
				object := markerPayload(t, installed)
				builds := object["builds"].(map[string]any)
				builds["build-tool"].(map[string]any)["driver"] = "go-v2"
				refuseMarker(t, installed, marshal(t, object))
			},
			want: buildUnsupportedDriver, outcome: "cache-hit",
		},
		"marker is not a marker document at all": {
			tamper: func(t *testing.T) { refuseMarker(t, installed, "{}") },
			want:   stateInvalidMarker, outcome: "cache-hit",
		},
		"build root reached agent-facing context": {
			tamper: func(t *testing.T) {
				if err := os.MkdirAll(filepath.Join(installed, "assets", "build-tool"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
			restore: func(t *testing.T) {
				if err := os.RemoveAll(filepath.Join(installed, "assets", "build-tool")); err != nil {
					t.Fatal(err)
				}
			},
			want: buildContextExposed, outcome: "cache-hit",
		},
		"protected cache boundary is no longer provable": {
			tamper: func(t *testing.T) {
				for _, entry := range cacheEntries(t, home) {
					breakCacheProtection(t, entry)
				}
			},
			restore: func(t *testing.T) {
				for _, entry := range cacheEntries(t, home) {
					restoreCacheProtection(t, entry, true)
				}
			},
			// One cache-boundary drift also proves the whole plain-text path.
			want: buildUntrustedCache, outcome: "would-rebuild-untrusted-cache", humanCLI: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			original, err := os.ReadFile(filepath.Join(installed, marker.Name)) // #nosec G304 -- test fixture path
			if err != nil {
				t.Fatal(err)
			}
			// Cleanup runs in reverse order, so the marker restore is registered
			// first to make it run last: live state is always put back before the
			// marker, and a cache snapshot restored afterwards would undo it.
			t.Cleanup(func() {
				if testCase.restore != nil {
					testCase.restore(t)
				}
				if writeErr := os.WriteFile(filepath.Join(installed, marker.Name), original, 0o644); writeErr != nil {
					t.Fatal(writeErr)
				}
			})
			if testCase.snapshotCache {
				snapshotBuildCacheAfter(t, home)
			}
			testCase.tamper(t)

			// `--json` and `--check` are combined so one invocation proves the
			// published document and the fail-closed exit are the same run's
			// verdict, rather than two classifications of drifting live state.
			code, stdout, stderr := capture(t, filepath.Join(home, "config.json"), "status", "app", "--json", "--check")
			if code != exitFail {
				t.Fatalf("status --json --check on %s = %d, want %d\nstdout:\n%s\nstderr:\n%s",
					name, code, exitFail, stdout, stderr)
			}
			drifted := decodeStatus(t, stdout)
			if len(drifted.Builds) != 1 {
				t.Fatalf("status reported %d compiled commands:\n%s", len(drifted.Builds), stdout)
			}
			if drifted.Builds[0].State != testCase.want {
				t.Fatalf("state = %q, want %q (row %+v)", drifted.Builds[0].State, testCase.want, drifted.Builds[0])
			}
			if drifted.Builds[0].Cause != testCase.cause {
				t.Fatalf("cause = %q, want %q (row %+v)", drifted.Builds[0].Cause, testCase.cause, drifted.Builds[0])
			}
			if drifted.Builds[0].CacheOutcome != testCase.outcome {
				t.Fatalf("cache outcome = %q, want %q", drifted.Builds[0].CacheOutcome, testCase.outcome)
			}
			if drifted.Builds[0].Detail == "" {
				t.Fatalf("non-current row %+v carries no operator detail", drifted.Builds[0])
			}
			wantSkill := testCase.skillState
			if wantSkill == "" {
				wantSkill = testCase.want
			}
			if drifted.Skills["build-skill"] != wantSkill {
				t.Fatalf("skills = %v, want build-skill %q", drifted.Skills, wantSkill)
			}

			// Every state and cause is rendered through the same method the
			// command prints, so the operator surface is proven for all of them
			// from the row this run published.
			described := drifted.Builds[0].Describe()
			if !strings.Contains(described, "state="+testCase.want) {
				t.Fatalf("the operator line does not report %q:\n%s", testCase.want, described)
			}
			if testCase.cause != "" && !strings.Contains(described, "cause="+testCase.cause) {
				t.Fatalf("the operator line does not report cause %q:\n%s", testCase.cause, described)
			}
			if !testCase.humanCLI {
				return
			}
			// One case of each shape also runs the whole plain-text path, which
			// pins the rendering above to the line the command really prints —
			// and proves the plain report still exits zero, because reporting a
			// verdict is not itself a failure.
			code, human, _ := capture(t, filepath.Join(home, "config.json"), "status", "app")
			if code != exitOK {
				t.Fatalf("plain human status on %s = %d, want %d\n%s", name, code, exitOK, human)
			}
			if !strings.Contains(human, "app: "+described) {
				t.Fatalf("the command did not print the line this row renders:\nwant: %s\ngot:\n%s", described, human)
			}
		})
	}
}

// assertCorruptCompiledStateIsRepaired proves the reconciliation path for a
// protected entry that cannot be interpreted at all: corrupt receipt bytes and
// corrupt artifact bytes are rebuilt by install and by upgrade, only after
// every gate has passed, and a run that fails leaves the previous installation
// and the live cache exactly as they were.
func assertCorruptCompiledStateIsRepaired(
	t *testing.T,
	fixture compiledProjectFixture,
	command string,
	corrupt func(t *testing.T, home string),
) {
	project, home, installed := fixture.project, fixture.home, fixture.installed
	before := marker.Read(installed)
	if before == nil || len(before.Builds) != 1 {
		t.Fatalf("the installation this case starts from records no compiled state: %+v", before)
	}

	corrupt(t, home)
	if code, stdout, _ := capture(t, filepath.Join(home, "config.json"), "status", "app", "--json"); code != exitOK ||
		decodeStatus(t, stdout).Builds[0].State != buildCorruptCache {
		t.Fatalf("status did not report corrupt compiled state:\n%s", stdout)
	}

	corruptedCache := cacheFingerprint(t, home)
	manifestPath := filepath.Join(project, manifest.Name)
	declared, err := os.ReadFile(manifestPath) // #nosec G304 -- test fixture path
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, manifestPath,
		`{"schema_version":1,"agents":["codex_cli"],"skills":[{"name":"build-skill","tag":"missing"}]}`)
	if code, stdout, _ := capture(t, filepath.Join(home, "config.json"), command, "app"); code != exitFail {
		t.Fatalf("%s with an unresolvable declaration = %d, want %d\n%s", command, code, exitFail, stdout)
	}
	if cacheFingerprint(t, home) != corruptedCache {
		t.Fatal("a refused run changed the live build cache")
	}
	if refused := marker.Read(installed); refused == nil || refused.Builds["build-tool"] != before.Builds["build-tool"] {
		t.Fatalf("a refused run changed the previous compiled state: %+v", refused)
	}
	if _, err := os.Stat(filepath.Join(installed, "SKILL.md")); err != nil {
		t.Fatalf("a refused run removed the previous installation: %v", err)
	}
	if err := os.WriteFile(manifestPath, declared, 0o644); err != nil {
		t.Fatal(err)
	}

	code, stdout, stderr := capture(t, filepath.Join(home, "config.json"), command, "app")
	if code != exitOK {
		t.Fatalf("repairing %s = %d\nstdout:\n%s\nstderr:\n%s", command, code, stdout, stderr)
	}
	if !strings.Contains(stdout, "rebuilt corrupt build cache state into a new protected entry") {
		t.Fatalf("%s did not report the repair:\n%s", command, stdout)
	}
	if code, _, _ := capture(t, filepath.Join(home, "config.json"), "status", "app", "--check"); code != exitOK {
		t.Fatalf("status --check after repairing cache state = %d, want %d", code, exitOK)
	}
	after := marker.Read(installed)
	if after == nil || after.Builds["build-tool"].CacheKey != before.Builds["build-tool"].CacheKey {
		t.Fatalf("a repair changed the logical key: %+v", after)
	}
	if len(cacheEntries(t, home)) != 1 {
		t.Fatalf("repair left %d live protected entries, want 1", len(cacheEntries(t, home)))
	}
}

// assertTheCacheIsRestoredWhenTheCommitFails is the other half of the repair
// contract, driven end to end through the real commands.
//
// The refusal case above fails a gate *before* the repair. This one lets the
// repair happen — the corrupt entry is really quarantined and a rebuilt entry
// really goes live — and then fails the durable commit that needed it. A run
// that committed nothing must also have displaced nothing, or the previous
// installation would silently change what `status` says about it: from
// `corrupt-build-cache`, which install repairs, to receipt or artifact drift
// against an entry nobody asked for.
//
// The commit is failed by taking away write access to the context store the
// first target class swaps in, which is the last moment a real installation can
// still fail without leaving a durable transaction behind.
func assertTheCacheIsRestoredWhenTheCommitFails(t *testing.T, fixture compiledProjectFixture, command string) {
	project, home, installed := fixture.project, fixture.home, fixture.installed
	// The shared fixture is current when this case starts, and a preceding
	// repair legitimately republished the entry the marker names, so the
	// compiled state this case must find unchanged is re-read here.
	before := marker.Read(installed)
	if before == nil || len(before.Builds) != 1 {
		t.Fatalf("the installation this case starts from records no compiled state: %+v", before)
	}

	corruptCacheArtifact(t, home)
	// A privileged process ignores the directory mode this case fails the
	// commit with, and skips below with the cache already corrupted. The
	// next case is handed a current fixture on that path too.
	t.Cleanup(func() {
		if t.Skipped() {
			reinstall(t, home)
		}
	})
	if code, stdout, _ := capture(t, filepath.Join(home, "config.json"), "status", "app", "--json"); code != exitOK ||
		decodeStatus(t, stdout).Builds[0].State != buildCorruptCache {
		t.Fatalf("status did not report corrupt compiled state:\n%s", stdout)
	}
	corruptArtifact := liveArtifactBytes(t, home)
	installedBefore := installedFingerprint(t, project, home)
	// Withdrawn entries are never deleted, only collected by the ordinary
	// sweep, so a shared cache root still holds what earlier cases withdrew.
	// The count this case owns is therefore measured against what it found.
	withdrawnBefore := quarantinedEntries(t, home)

	store := filepath.Join(project, ".agents", "skills")
	denyWrites(t, store)
	code, stdout, stderr := capture(t, filepath.Join(home, "config.json"), command, "app")
	restoreWrites(t, store)
	if code != exitFail {
		t.Fatalf("%s over an unwritable context store = %d, want %d\nstdout:\n%s\nstderr:\n%s",
			command, code, exitFail, stdout, stderr)
	}

	// The case only proves anything if the repair really did publish before
	// the commit failed. Publication quarantines the corrupt entry and the
	// reversal withdraws the replacement the same way, so exactly one
	// withdrawn entry is left behind for the ordinary sweep.
	if withdrawn := quarantinedEntries(t, home) - withdrawnBefore; withdrawn != 1 {
		t.Fatalf("withdrawn cache entries = %d, want exactly the one this run published", withdrawn)
	}
	// Nothing half-built survives in the shared cache root either: a
	// publication that failed or was reversed leaves no private staging
	// directory behind for the next run to trip over.
	if staged := stagedEntries(t, home); staged != 0 {
		t.Fatalf("private staging directories left in the shared cache root = %d, want 0", staged)
	}

	// Exact restoration: the entry the run displaced is live again, with
	// the bytes it had, and the installation is untouched.
	if got := liveArtifactBytes(t, home); got != corruptArtifact {
		t.Fatal("a failed commit did not restore the build cache entry it replaced")
	}
	if len(cacheEntries(t, home)) != 1 {
		t.Fatalf("a failed commit left %d live protected entries, want 1", len(cacheEntries(t, home)))
	}
	if failed := marker.Read(installed); failed == nil ||
		failed.Builds["build-tool"] != before.Builds["build-tool"] {
		t.Fatalf("a failed commit changed the previous compiled state: %+v", failed)
	}
	// The marker is only one class of what an installation commits. The
	// launcher bytes the previous install placed, its adapter mirrors, the
	// machine runtime, and the consumer ledger all have to be at the exact
	// value the run found them at.
	if got := installedFingerprint(t, project, home); got != installedBefore {
		t.Fatalf("a failed commit changed installed state\nbefore:\n%s\nafter:\n%s",
			installedBefore, got)
	}
	if code, stdout, _ := capture(t, filepath.Join(home, "config.json"), "status", "app", "--json"); code != exitOK ||
		decodeStatus(t, stdout).Builds[0].State != buildCorruptCache {
		t.Fatalf("the previous installation no longer reports the state it had:\n%s", stdout)
	}

	// The ordinary reconciliation path still repairs it afterwards.
	if code, stdout, stderr := capture(t, filepath.Join(home, "config.json"), command, "app"); code != exitOK {
		t.Fatalf("repairing %s = %d\nstdout:\n%s\nstderr:\n%s", command, code, stdout, stderr)
	}
	if code, _, _ := capture(t, filepath.Join(home, "config.json"), "status", "app", "--check"); code != exitOK {
		t.Fatalf("status --check after the repair = %d, want %d", code, exitOK)
	}
}

// installedFingerprint digests every class of state one installation commits:
// the project tree it owns, the agent adapter mirrors it publishes, and the
// machine-level runtime, hybrid store, and consumer ledger. A run that failed
// must leave the whole map at its prior value, not just the parts a single
// assertion happens to look at.
func installedFingerprint(t *testing.T, project, home string) string {
	t.Helper()
	paths := map[string]string{
		"project/agents":         filepath.Join(project, ".agents"),
		"project/adapters":       filepath.Join(project, ".claude", "skills"),
		"project/adapters-codex": filepath.Join(project, filepath.FromSlash(adapters.AgentPaths["codex_cli"])),
		"home/runtime":           filepath.Join(home, "runtime"),
		"home/hybrid":            scopes.HybridSkillsRoot(home),
		"home/consumers":         filepath.Join(home, scopes.ConsumersName),
	}
	records := make([]string, 0, len(paths))
	for name, path := range paths {
		digest := pathDigest(t, path)
		// A fingerprint of nothing would compare equal to a fingerprint of
		// nothing, so the two classes that always exist after an installation are
		// required to be there. The adapter mirrors and the hybrid store depend on
		// what the project declares and may legitimately be absent.
		if digest == "absent" && (name == "project/agents" || name == "home/consumers") {
			t.Fatalf("%s is missing, so this fingerprint would compare equal to anything", name)
		}
		records = append(records, name+"="+digest)
	}
	sort.Strings(records)
	return strings.Join(records, "\n")
}

// pathDigest reads one path exactly as it is. A symbolic link is digested by
// its destination rather than followed, so a mirror that was replaced,
// re-pointed, or removed shows up as a change instead of resolving to the same
// canonical bytes.
func pathDigest(t *testing.T, path string) string {
	t.Helper()
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return "absent"
	}
	if err != nil {
		t.Fatal(err)
	}
	switch {
	case info.Mode()&os.ModeSymlink != 0:
		target, err := os.Readlink(path)
		if err != nil {
			t.Fatal(err)
		}
		return "link:" + filepath.ToSlash(target)
	case info.IsDir():
		entries, err := os.ReadDir(path)
		if err != nil {
			t.Fatal(err)
		}
		members := make([]string, 0, len(entries))
		for _, entry := range entries {
			members = append(members, entry.Name()+"="+pathDigest(t, filepath.Join(path, entry.Name())))
		}
		sort.Strings(members)
		return "dir[" + strings.Join(members, ",") + "]"
	default:
		payload, err := os.ReadFile(path) // #nosec G304 -- test fixture path
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(payload)
		return fmt.Sprintf("file:%s:%s", info.Mode().Perm(), hex.EncodeToString(sum[:]))
	}
}

// liveArtifactBytes reads the single artifact of the single live protected
// entry, so a test can prove which bytes a reversal put back.
func liveArtifactBytes(t *testing.T, home string) string {
	t.Helper()
	entries := cacheEntries(t, home)
	if len(entries) != 1 {
		t.Fatalf("live protected entries = %d, want exactly 1", len(entries))
	}
	matches, err := filepath.Glob(filepath.Join(entries[0], "bin", "*"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("protected entry %s does not hold exactly one artifact: %v %v", entries[0], matches, err)
	}
	payload, err := os.ReadFile(matches[0]) // #nosec G304 -- test fixture path
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

// replaceCacheEntry rewrites the single live protected entry into a different
// but equally valid entry for the same logical key: new artifact bytes and the
// canonical receipt that exactly describes them.
//
// It is the one way protected state can move without ever looking broken, so it
// is the case a currentness check can only catch by re-reading the cache.
func replaceCacheEntry(t *testing.T, home string, input buildmeta.Input) {
	t.Helper()
	entries := cacheEntries(t, home)
	if len(entries) != 1 {
		t.Fatalf("live protected entries = %d, want exactly 1", len(entries))
	}
	entry := entries[0]
	relative, err := buildmeta.ArtifactPath(input.Command, input.Target.GOOS)
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte("a different artifact, described exactly by its own receipt")
	digest := sha256.Sum256(payload)
	receipt, err := buildmeta.NewReceipt(input, buildmeta.Artifact{
		Path: relative, SHA256: "sha256:" + hex.EncodeToString(digest[:]), Size: int64(len(payload)),
	})
	if err != nil {
		t.Fatal(err)
	}
	receiptBytes, err := receipt.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	artifact := filepath.Join(entry, filepath.FromSlash(relative))
	if err := os.Chmod(artifact, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(artifact, payload, 0o700); err != nil {
		t.Fatal(err)
	}
	receiptPath := filepath.Join(entry, buildcache.ReceiptFilename)
	if err := os.Chmod(receiptPath, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(receiptPath, receiptBytes, 0o600); err != nil {
		t.Fatal(err)
	}
}

// quarantinedEntries counts the cache entries that were moved aside rather than
// deleted, which is how both a replacement and its reversal withdraw one.
func quarantinedEntries(t *testing.T, home string) int {
	t.Helper()
	return cacheRootPrefixCount(t, home, ".quarantine-")
}

// stagedEntries counts the private staging directories left in the shared cache
// root. A publication owns its staging on every path, so this is always zero
// once a run has released the manager-home lock.
func stagedEntries(t *testing.T, home string) int {
	t.Helper()
	return cacheRootPrefixCount(t, home, ".stage-")
}

func cacheRootPrefixCount(t *testing.T, home, prefix string) int {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(home, "cache", "build", "go-v1"))
	if err != nil {
		t.Fatalf("protected build cache is missing: %v", err)
	}
	count := 0
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), prefix) {
			count++
		}
	}
	return count
}

// denyWrites makes one directory unwritable and proves it really is, so a case
// that depends on a write failing cannot pass vacuously — a privileged process
// ignores the mode entirely.
func denyWrites(t *testing.T, dir string) {
	t.Helper()
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, info.Mode().Perm()) })
	probe := filepath.Join(dir, ".curator-write-probe")
	if err := os.WriteFile(probe, []byte("probe"), 0o600); err == nil {
		_ = os.Remove(probe)
		if err := os.Chmod(dir, info.Mode().Perm()); err != nil {
			t.Fatal(err)
		}
		t.Skip("this process can write through a read-only directory, so a commit failure cannot be injected")
	}
}

func restoreWrites(t *testing.T, dir string) {
	t.Helper()
	if err := os.Chmod(dir, 0o755); err != nil {
		t.Fatal(err)
	}
}

// TestStatusReportsATransitivelyResolvedCompiledCommand proves the compiled
// surface is not limited to project declarations: a build that belongs to a
// node the project reaches only through a dependency is reported, classified,
// and fails --check on its own, even though no declaration names it.
func TestStatusReportsATransitivelyResolvedCompiledCommand(t *testing.T) {
	t.Parallel()
	requireNativeControlInventoryPlatform(t)
	project, home := compiledProjectDeclaring(t, `{"name":"consumer","tag":"v1"}`)
	consumer := filepath.Join(filepath.Dir(project), "skills", "consumer")
	writeFile(t, filepath.Join(consumer, "SKILL.md"), "---\nname: consumer\ndescription: consumer\n---\n")
	writeFile(t, filepath.Join(consumer, "agent-skill.json"), `{
		"schema_version": 6,
		"build_roots": [],
		"capabilities": {},
		"commands": {},
		"dependencies": {"skills": {
			"build-skill": {
				"git": "./build-skill",
				"ref": {"kind":"tag","value":"v1"},
				"mode": "runtime"
			}
		}}
	}`)
	runGit(t, consumer, "init", "-q", "-b", "main")
	runGit(t, consumer, "add", ".")
	runGit(t, consumer, "commit", "-qm", "initial consumer")
	runGit(t, consumer, "tag", "v1")

	if code, stdout, stderr := capture(t, filepath.Join(home, "config.json"), "install", "app"); code != exitOK {
		t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}

	code, stdout, stderr := capture(t, filepath.Join(home, "config.json"), "status", "app", "--json")
	if code != exitOK {
		t.Fatalf("status --json = %d\nstderr:\n%s", code, stderr)
	}
	report := decodeStatus(t, stdout)
	if len(report.Builds) != 1 || report.Builds[0].Skill != "build-skill" ||
		report.Builds[0].State != buildCurrent {
		t.Fatalf("transitive compiled command was not reported:\n%s", stdout)
	}
	if _, declared := report.Skills["build-skill"]; declared {
		t.Fatalf("the provider is a transitive node, not a declaration: %v", report.Skills)
	}
	if code, _, _ := capture(t, filepath.Join(home, "config.json"), "status", "app", "--check"); code != exitOK {
		t.Fatalf("clean transitive compiled status --check = %d, want %d", code, exitOK)
	}

	// The provider's own compiled state must still fail --check on its own,
	// through the drift surface no declared-skill map can see.
	for _, entry := range cacheEntries(t, home) {
		if err := os.RemoveAll(entry); err != nil {
			t.Fatal(err)
		}
	}
	code, stdout, _ = capture(t, filepath.Join(home, "config.json"), "status", "app", "--json")
	if code != exitOK {
		t.Fatalf("status --json after the entry was removed = %d", code)
	}
	if drifted := decodeStatus(t, stdout); len(drifted.Builds) != 1 ||
		drifted.Builds[0].State != buildMissingArtifact {
		t.Fatalf("transitive drift was not reported:\n%s", stdout)
	}
	if code, _, _ := capture(t, filepath.Join(home, "config.json"), "status", "app", "--check"); code != exitFail {
		t.Fatalf("transitive compiled drift passed --check: %d", code)
	}
}

// TestStatusReportsAnUnusableToolchainPerCompiledCommand proves the
// missing-or-incompatible Go diagnostic is not stderr-only: every active
// compiled command still gets a machine-readable row carrying the stable
// go-v1 boundary code, and `status --check` fails.
func TestStatusReportsAnUnusableToolchainPerCompiledCommand(t *testing.T) {
	t.Parallel()
	requireNativeControlInventoryPlatform(t)
	_, home := compiledProject(t)
	if code, stdout, stderr := capture(t, filepath.Join(home, "config.json"), "install", "app"); code != exitOK {
		t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	selected := map[string]string{
		godriver.SelectionCuratorGo: filepath.Join(t.TempDir(), "nowhere", "bin", "go"),
	}

	// Reporting a refusal is not itself a failure: the row carries the verdict
	// and `--check` is what turns a non-current verdict into a non-zero exit.
	code, stdout, stderr := captureWithEnv(t, filepath.Join(home, "config.json"), selected, "status", "app", "--json")
	if code != exitOK {
		t.Fatalf("status --json with an unusable toolchain = %d, want %d\n%s", code, exitOK, stderr)
	}
	if !strings.Contains(stderr, "warning:") || strings.Contains(stderr, "error:") {
		t.Fatalf("status reported the refusal as a failure rather than a warning:\n%s", stderr)
	}
	report := decodeStatus(t, stdout)
	if len(report.Builds) != 1 {
		t.Fatalf("status reported %d compiled commands, want 1:\n%s", len(report.Builds), stdout)
	}
	row := report.Builds[0]
	if row.State != buildUnusableToolchain || row.Cause == "" {
		t.Fatalf("toolchain row = %+v", row)
	}
	if row.Skill != "build-skill" || row.Command != "build-tool" || row.Detail == "" {
		t.Fatalf("toolchain row does not name the active command: %+v", row)
	}
	// Everything that was already established when the toolchain was consulted
	// is reported. Build sources are validated before the toolchain is touched
	// and the driver is closed by the schema, so a row that omitted them would
	// be hiding identity it had.
	if row.Driver != "go-v1" || row.BuildRoot != "assets/build-tool" ||
		row.SourceDir != "assets/build-tool/cmd/tool" || row.BuildSource.ContentSHA256 == "" {
		t.Fatalf("toolchain row dropped identities the plan already had: %+v", row)
	}
	// Nothing the plan could not derive may be published as if it were known.
	if row.CacheKey != "" || row.Target != "" || row.ArtifactPath != "" {
		t.Fatalf("toolchain row published identities the plan never derived: %+v", row)
	}
	if report.Skills["build-skill"] != buildUnusableToolchain {
		t.Fatalf("skills = %v", report.Skills)
	}

	code, human, _ := captureWithEnv(t, filepath.Join(home, "config.json"), selected, "status", "app")
	if code != exitOK || !strings.Contains(human, "state="+buildUnusableToolchain) ||
		!strings.Contains(human, "cause="+row.Cause) {
		t.Fatalf("human status = %d and does not report the toolchain row:\n%s", code, human)
	}
	// The human line and the JSON row carry the same identities, field for field.
	for _, want := range []string{
		"driver=" + row.Driver, "root=" + row.BuildRoot, "dir=" + row.SourceDir,
		"source=" + row.BuildSource.Algorithm + ":" + row.BuildSource.ContentSHA256,
	} {
		if !strings.Contains(human, want) {
			t.Fatalf("human status does not report %q:\n%s", want, human)
		}
	}
	// The cache outcome is the row's own planner verdict and is reported; the
	// identities that depend on the toolchain are not published at all.
	if !strings.Contains(human, "cache="+row.CacheOutcome) {
		t.Fatalf("human status does not report the planner outcome %q:\n%s", row.CacheOutcome, human)
	}
	for _, absent := range []string{"target=", "key=", "artifact="} {
		if strings.Contains(human, absent) {
			t.Fatalf("human status published %q for a command nothing was derived for:\n%s", absent, human)
		}
	}
	if code, _, _ := captureWithEnv(t, filepath.Join(home, "config.json"), selected, "status", "app", "--check"); code != exitFail {
		t.Fatalf("status --check with an unusable toolchain = %d, want %d", code, exitFail)
	}
}

// assertProtectedCacheStateThatMovedDuringTheCheck closes the other half of the
// classification window.
//
// Install markers are fingerprinted around the whole run, but a marker is only
// half the evidence a compiled verdict is derived from: the protected cache is
// inspected once, during a read-only plan that holds no lock, and an entry that
// is removed, corrupted, or loses its provable protection afterwards changes
// every verdict taken from it without touching a single marker byte.
//
// The cache really moves here, between the plan and the classification, exactly
// as a concurrent install, sweep, or permission change would move it. Nothing
// rewrites a marker, so a verdict that stays authoritative could only have come
// from evidence nobody re-read.
func assertProtectedCacheStateThatMovedDuringTheCheck(t *testing.T, fixture compiledProjectFixture) {
	project, home := fixture.project, fixture.home

	command := cli{config: fileConfigSource(filepath.Join(home, "config.json")), stderr: io.Discard}
	cfg, code := command.loadConfig()
	if code != exitOK {
		t.Fatalf("loading the configuration = %d", code)
	}

	// The same read-only plan `status` takes, and the same marker fingerprint
	// it opens the window with.
	planned := install.Project(cfg, project, "app", install.Options{DryRun: true})
	if planned.Status != "ok" || len(planned.Builds) != 1 {
		t.Fatalf("read-only plan = %+v", planned)
	}
	facts := factsList(planned.Builds)
	scope := projectStatusScope(cfg, project, "app")
	before := markerDigests(scope.stores...)

	drift, rows := statusReport(cfg, scope, facts, before)
	if len(rows) != 1 || rows[0].State != buildCurrent {
		t.Fatalf("settled compiled state = %+v", rows)
	}
	if checkFailed(drift, rows) {
		t.Fatal("a settled compiled installation failed --check")
	}

	for name, move := range map[string]func(t *testing.T, input buildmeta.Input){
		"the protected entry is corrupted": func(t *testing.T, _ buildmeta.Input) {
			corruptCacheArtifact(t, home)
		},
		"the protected entry is removed": func(t *testing.T, _ buildmeta.Input) {
			for _, entry := range cacheEntries(t, home) {
				if err := os.RemoveAll(entry); err != nil {
					t.Fatal(err)
				}
			}
		},
		"the protected boundary stops being provable": func(t *testing.T, _ buildmeta.Input) {
			for _, entry := range cacheEntries(t, home) {
				breakCacheProtection(t, entry)
			}
		},
		// The hardest case: the entry is still a valid hit for the same logical
		// key, so nothing about it looks broken — only its receipt identity moved.
		// Without the recheck this publishes a stale *drift* verdict
		// (corrupt-build-receipt against the marker) rather than a stale current
		// one, which is just as wrong and far less obvious.
		"the protected entry is replaced by another valid one": func(t *testing.T, input buildmeta.Input) {
			replaceCacheEntry(t, home, input)
		},
	} {
		t.Run(name, func(t *testing.T) {
			// Each case opens its own window over the state it is about to move:
			// the case before it put every protected cache byte back from its own
			// snapshot, so the production plan and the marker fingerprint this
			// window is taken from are acquired again here rather than carried in.
			planned := install.Project(cfg, project, "app", install.Options{DryRun: true})
			if planned.Status != "ok" || len(planned.Builds) != 1 {
				t.Fatalf("read-only plan = %+v", planned)
			}
			facts := factsList(planned.Builds)
			before := markerDigests(scope.stores...)
			markerPath := filepath.Join(project, ".agents", "skills", "build-skill", marker.Name)
			marked, err := os.ReadFile(markerPath) // #nosec G304 -- test fixture path
			if err != nil {
				t.Fatal(err)
			}

			// Every byte and permission bit this case moves is copied aside before
			// it moves, and put back when the case ends. Repairing the fixture by
			// reinstalling compiles the command again and proves nothing this test
			// owns; install, repair, and rollback keep their own dedicated cases.
			snapshotBuildCacheAfter(t, home)
			move(t, planned.Builds[0].Expectation().Input)

			// The marker is deliberately untouched, so only the cache recheck can
			// have produced the verdict below.
			after, err := os.ReadFile(markerPath) // #nosec G304 -- test fixture path
			if err != nil {
				t.Fatal(err)
			}
			if string(after) != string(marked) {
				t.Fatal("the case moved the install marker, so it proves nothing about the cache recheck")
			}

			drift, rows := statusReport(cfg, scope, facts, before)
			if len(rows) != 1 || rows[0].State != buildStateChanged {
				t.Fatalf("moved compiled state = %+v", rows)
			}
			if !strings.Contains(rows[0].Detail, "build cache") {
				t.Fatalf("the row does not say what moved: %+v", rows[0])
			}
			if drift["build-skill"] != buildStateChanged {
				t.Fatalf("skills = %v, want build-skill %q", drift, buildStateChanged)
			}
			if !checkFailed(drift, rows) {
				t.Fatal("compiled state that moved during the check passed --check")
			}
		})
	}
}

// TestStatusKeepsLegacyBehaviourWhenAPlanFailsWithoutCompiledCommands pins the
// compatibility boundary of the change above: a failure that derived no
// compiled command still reports exactly the historical error and nothing else.
func TestStatusKeepsLegacyBehaviourWhenAPlanFailsWithoutCompiledCommands(t *testing.T) {
	t.Parallel()
	project, home := legacyProject(t)
	if code, _, stderr := capture(t, filepath.Join(home, "config.json"), "install", "app"); code != exitOK {
		t.Fatalf("install = %d\n%s", code, stderr)
	}
	// A declaration that cannot be resolved fails the read-only plan long
	// before any build phase.
	writeFile(t, filepath.Join(project, manifest.Name),
		`{"schema_version":1,"agents":["codex_cli"],"skills":[{"name":"skill-a","tag":"v9"}]}`)

	code, stdout, stderr := capture(t, filepath.Join(home, "config.json"), "status", "app", "--json")
	if code != exitFail {
		t.Fatalf("status over an unresolvable declaration = %d, want %d", code, exitFail)
	}
	if strings.TrimSpace(stdout) != "[]" {
		t.Fatalf("a failed legacy status printed a document it never printed before:\n%s", stdout)
	}
	if stderr == "" {
		t.Fatal("a failed legacy status printed no error")
	}
}

// assertUntrustedCompiledStateIsRepaired proves the documented reconciliation
// path: install rebuilds candidate bytes that are outside a provable protected
// boundary instead of adopting them.
func assertUntrustedCompiledStateIsRepaired(t *testing.T, fixture compiledProjectFixture) {
	project, home, installed := fixture.project, fixture.home, fixture.installed

	// The shared fixture is current when this case starts, and a preceding
	// repair legitimately republished the entry the marker names, so the
	// compiled state this case must find preserved is re-read here.
	before := marker.Read(installed)
	if before == nil || len(before.Builds) != 1 {
		t.Fatalf("the installation this case starts from records no compiled state: %+v", before)
	}
	for _, entry := range cacheEntries(t, home) {
		breakCacheProtection(t, entry)
	}

	code, dryRun, _ := capture(t, filepath.Join(home, "config.json"), "install", "app", "--dry-run")
	if code != exitOK {
		t.Fatalf("dry run over untrusted cache state = %d\n%s", code, dryRun)
	}
	if !strings.Contains(dryRun, "outcome=would-rebuild-untrusted-cache") {
		t.Fatalf("dry run did not name the untrusted protected state:\n%s", dryRun)
	}
	if strings.Contains(dryRun, "rebuilt untrusted") {
		t.Fatalf("dry run claimed a completed repair:\n%s", dryRun)
	}

	code, stdout, stderr := capture(t, filepath.Join(home, "config.json"), "install", "app")
	if code != exitOK {
		t.Fatalf("repairing install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "rebuilt untrusted build cache state into a new protected entry") {
		t.Fatalf("install did not report the repair:\n%s", stdout)
	}
	if code, _, _ := capture(t, filepath.Join(home, "config.json"), "status", "app", "--check"); code != exitOK {
		t.Fatalf("status --check after repair = %d, want %d", code, exitOK)
	}
	after := marker.Read(installed)
	if after == nil || len(after.Builds) != 1 {
		t.Fatalf("repaired installation lost its compiled state: %+v", after)
	}
	if after.Builds["build-tool"].CacheKey != before.Builds["build-tool"].CacheKey {
		t.Fatalf("a repair changed the logical key: %q -> %q",
			before.Builds["build-tool"].CacheKey, after.Builds["build-tool"].CacheKey)
	}

	// A run that cannot resolve its declaration repairs nothing and leaves the
	// durable installation exactly as the successful run left it.
	manifestPath := filepath.Join(project, manifest.Name)
	declared, err := os.ReadFile(manifestPath) // #nosec G304 -- test fixture path
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, manifestPath,
		`{"schema_version":1,"agents":["codex_cli"],"skills":[{"name":"build-skill","tag":"missing"}]}`)
	if code, stdout, _ := capture(t, filepath.Join(home, "config.json"), "install", "app"); code != exitFail {
		t.Fatalf("install with an unresolvable declaration = %d, want %d\n%s", code, exitFail, stdout)
	}
	if err := os.WriteFile(manifestPath, declared, 0o644); err != nil {
		t.Fatal(err)
	}
	failed := marker.Read(installed)
	if failed == nil || failed.Builds["build-tool"] != after.Builds["build-tool"] {
		t.Fatalf("a failed install changed the previous compiled state: %+v", failed)
	}
	if _, err := os.Stat(filepath.Join(installed, "SKILL.md")); err != nil {
		t.Fatalf("a failed install removed the previous installation: %v", err)
	}
}

// TestGCRetainsAndReportsReferencedCompiledState proves the maintenance
// surface: a cache entry a live marker references is neither removed nor
// reported as removed.
func TestGCRetainsAndReportsReferencedCompiledState(t *testing.T) {
	t.Parallel()
	requireNativeControlInventoryPlatform(t)
	_, home := compiledProject(t)
	if code, stdout, stderr := capture(t, filepath.Join(home, "config.json"), "install", "app"); code != exitOK {
		t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	entries := cacheEntries(t, home)
	if len(entries) != 1 {
		t.Fatalf("protected build cache holds %d entries, want 1", len(entries))
	}

	code, stdout, stderr := capture(t, filepath.Join(home, "config.json"), "gc")
	if code != exitOK {
		t.Fatalf("gc = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "0 build entries removed") {
		t.Fatalf("gc did not report the compiled sweep result:\n%s", stdout)
	}
	if strings.Contains(stdout, "removed build ") {
		t.Fatalf("gc removed a referenced compiled entry:\n%s", stdout)
	}
	if _, err := os.Stat(entries[0]); err != nil {
		t.Fatalf("gc removed a referenced protected entry: %v", err)
	}
	if code, _, _ := capture(t, filepath.Join(home, "config.json"), "status", "app", "--check"); code != exitOK {
		t.Fatalf("status --check after gc = %d, want %d", code, exitOK)
	}
}

// TestCompiledInstallFollowsTheNativeControlInventoryExactly asserts the
// platform carve-out of the compiled-build surface positively, in both
// directions, on every runner — including the one where the carved-out cases
// above do not run.
//
// rc5-native-control-inventory-v1 defines control records for exactly macOS and
// Windows. The claim under test is that a compiled install succeeds on exactly
// those platforms and fails closed everywhere else, so:
//
//   - on a covered host a real compiled command installs and publishes exactly
//     one protected cache entry;
//   - on an uncovered host the same invocation is refused with
//     build_execution_control_unavailable, naming the inventory and the host,
//     publishes nothing at all, and leaves `status --check` non-zero rather
//     than reporting a compiled command as current.
//
// This is what makes the skip in requireNativeControlInventoryPlatform a
// recorded boundary rather than a quiet omission: the uncovered runner still
// proves the refusal it is carved out by.
func TestCompiledInstallFollowsTheNativeControlInventoryExactly(t *testing.T) {
	t.Parallel()
	_, home := compiledProject(t)
	code, stdout, stderr := capture(t, filepath.Join(home, "config.json"), "install", "app")

	if godriver.InventoryPlatform(runtime.GOOS) != "" {
		if code != exitOK {
			t.Fatalf("a compiled install failed on %s, which rc5-native-control-inventory-v1 covers: %d\nstdout:\n%s\nstderr:\n%s",
				runtime.GOOS, code, stdout, stderr)
		}
		if entries := publishedCacheEntries(t, home); len(entries) != 1 {
			t.Fatalf("a covered platform published %d protected cache entries, want 1", len(entries))
		} else {
			payload, err := os.ReadFile(filepath.Join(entries[0], buildcache.ExecutionReceiptFilename))
			if err != nil {
				t.Fatalf("portable build published no execution receipt: %v", err)
			}
			receipt, err := closureexec.DecodeBuildSessionReceipt(payload)
			if err != nil || receipt.Binding.AssuranceMode != closureexec.AssurancePortable || receipt.ProviderExecutionReceipt != nil {
				t.Fatalf("portable execution receipt = %+v, %v", receipt, err)
			}
			for _, capability := range receipt.Binding.ActualCapabilities {
				if strings.Contains(capability.CapabilityID, "lossless") || strings.Contains(capability.CapabilityID, "total-network") {
					t.Fatalf("portable receipt inflated capability %q", capability.CapabilityID)
				}
			}
		}
		return
	}

	if code == exitOK {
		t.Fatalf("a compiled install succeeded on %s, for which rc5-native-control-inventory-v1 defines no record\nstdout:\n%s\nstderr:\n%s",
			runtime.GOOS, stdout, stderr)
	}
	if !strings.Contains(stderr, godriver.CodeControlUnavailable) {
		t.Fatalf("the refusal did not carry %s:\nstderr:\n%s", godriver.CodeControlUnavailable, stderr)
	}
	for _, want := range []string{
		godriver.NativeControlInventoryVersion,
		"no record for host " + runtime.GOOS,
	} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("the refusal did not name %q:\nstderr:\n%s", want, stderr)
		}
	}
	// A refusal before the worker starts publishes nothing: no protected cache
	// entry, and no compiled state for a later run to trust.
	if entries := publishedCacheEntries(t, home); len(entries) != 0 {
		t.Fatalf("a refused compiled install published %d protected cache entries: %v", len(entries), entries)
	}
	// And the status surface fails closed rather than reporting the command as
	// current on a platform that cannot produce it.
	if checkCode, checkOut, _ := capture(t, filepath.Join(home, "config.json"), "status", "app", "--check"); checkCode == exitOK {
		t.Fatalf("status --check passed after a refused compiled install:\n%s", checkOut)
	}
}

// TestDryRunNeverClaimsACompletedCompilerCheck pins the compiler-free dry-run
// vocabulary: a miss is planned, never reported as preflighted or built.
func TestDryRunNeverClaimsACompletedCompilerCheck(t *testing.T) {
	t.Parallel()
	_, home := compiledProject(t)
	code, stdout, stderr := capture(t, filepath.Join(home, "config.json"), "install", "app", "--dry-run")
	if code != exitOK {
		t.Fatalf("dry run = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "outcome=would-preflight-and-build") {
		t.Fatalf("cold dry run did not plan a preflight and build:\n%s", stdout)
	}
	for _, forbidden := range []string{
		"outcome=built", "outcome=preflighted", "staged key=", "rebuilt untrusted",
	} {
		if strings.Contains(stdout, forbidden) {
			t.Fatalf("dry run claimed completed compiler work %q:\n%s", forbidden, stdout)
		}
	}
	if !strings.Contains(stdout, "dry-run; no files modified") {
		t.Fatalf("dry run did not state that nothing changed:\n%s", stdout)
	}

	// upgrade shares the install phase, so it reports the same compiler-free
	// planning vocabulary.
	code, upgraded, stderr := capture(t, filepath.Join(home, "config.json"), "upgrade", "app", "--dry-run")
	if code != exitOK {
		t.Fatalf("upgrade --dry-run = %d\nstdout:\n%s\nstderr:\n%s", code, upgraded, stderr)
	}
	if !strings.Contains(upgraded, "outcome=would-preflight-and-build") {
		t.Fatalf("upgrade --dry-run did not plan a preflight and build:\n%s", upgraded)
	}
}

// TestStatusExplainsAnUnusableGoToolchain proves the missing-or-incompatible
// toolchain diagnostic names every accepted selection mechanism and the tested
// release families, on any host, without suggesting a PATH lookup or download.
func TestStatusExplainsAnUnusableGoToolchain(t *testing.T) {
	t.Parallel()
	_, home := compiledProject(t)
	selected := map[string]string{
		godriver.SelectionCuratorGo: filepath.Join(t.TempDir(), "nowhere", "bin", "go"),
	}

	code, _, stderr := captureWithEnv(t, filepath.Join(home, "config.json"), selected, "status", "app")
	if code != exitOK {
		t.Fatalf("status with an unusable toolchain = %d, want %d\n%s", code, exitOK, stderr)
	}
	for _, want := range append([]string{
		godriver.SelectionCuratorGo, godriver.SelectionGOROOT,
		"Curator never searches PATH and never downloads a toolchain",
	}, godriver.TestedFamilies()...) {
		if !strings.Contains(stderr, want) {
			t.Fatalf("toolchain diagnostic does not name %q:\n%s", want, stderr)
		}
	}
	if code, _, _ := captureWithEnv(t, filepath.Join(home, "config.json"), selected, "status", "app", "--check"); code != exitFail {
		t.Fatalf("status --check with an unusable toolchain = %d, want %d", code, exitFail)
	}
	if code, _, _ := captureWithEnv(t, filepath.Join(home, "config.json"), selected, "install", "app"); code != exitFail {
		t.Fatalf("install with an unusable toolchain = %d, want %d", code, exitFail)
	}
}

// TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands pins backward
// compatibility: a closure with no build command produces exactly the
// historical document, with no build key at all.
func TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands(t *testing.T) {
	t.Parallel()
	_, home := legacyProject(t)
	if code, _, stderr := capture(t, filepath.Join(home, "config.json"), "install", "app"); code != exitOK {
		t.Fatalf("install = %d\n%s", code, stderr)
	}

	code, stdout, _ := capture(t, filepath.Join(home, "config.json"), "status", "app", "--json")
	if code != exitOK {
		t.Fatalf("status --json = %d", code)
	}
	var object map[string]any
	if err := json.Unmarshal([]byte(stdout), &object); err != nil {
		t.Fatalf("status --json is not one JSON object: %v\n%s", err, stdout)
	}
	if len(object) != 3 || object["alias"] == nil || object["path"] == nil || object["skills"] == nil {
		t.Fatalf("status --json changed the historical document shape:\n%s", stdout)
	}
	skills, ok := object["skills"].(map[string]any)
	if !ok || skills["skill-a"] != stateUpToDate {
		t.Fatalf("skills = %v", object["skills"])
	}
}

// TestStatusAcceptsAnUnchangedLegacyMarkerSchema pins the legacy contract: a
// marker below the current schema still describes a current schema 1 through 5
// installation and must not be reported as unsupported.
func TestStatusAcceptsAnUnchangedLegacyMarkerSchema(t *testing.T) {
	t.Parallel()
	project, home := legacyProject(t)
	if code, _, stderr := capture(t, filepath.Join(home, "config.json"), "install", "app"); code != exitOK {
		t.Fatalf("install = %d\n%s", code, stderr)
	}

	installed := filepath.Join(project, ".agents", "skills", "skill-a")
	rewriteMarker(t, installed, func(object map[string]any) {
		object["schema_version"] = marker.LegacySchemaVersion
		delete(object, "build_roots")
		delete(object, "build_source")
		delete(object, "builds")
	})

	code, stdout, stderr := capture(t, filepath.Join(home, "config.json"), "status", "app", "--check")
	if code != exitOK {
		t.Fatalf("legacy marker status --check = %d, want %d\nstdout:\n%s\nstderr:\n%s", code, exitOK, stdout, stderr)
	}
	if !strings.Contains(stdout, "app: skill-a "+stateUpToDate) {
		t.Fatalf("legacy marker was not reported as current:\n%s", stdout)
	}
}

func marshal(t *testing.T, value any) string {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(payload)
}
