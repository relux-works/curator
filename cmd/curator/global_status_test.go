package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
)

// globalPayload is the machine-readable document the machine-wide scope
// produces. It deliberately carries no `path`: the scope has no
// operator-supplied root and the manager home is never published.
type globalPayload struct {
	Alias  string            `json:"alias"`
	Skills map[string]string `json:"skills"`
	Builds []buildReport     `json:"builds"`
}

func decodeGlobalStatus(t *testing.T, payload string) globalPayload {
	t.Helper()
	var decoded globalPayload
	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		t.Fatalf("global status --json is not one JSON object: %v\n%s", err, payload)
	}
	return decoded
}

// globalScopeDeclaring creates a manager home whose machine-wide Skillfile
// declares exactly what the caller asks for, and redirects the user home the
// global scope mirrors adapters and forwarding shims into, so a real
// installation cannot reach the operator's own home. It returns the manager
// home.
func globalScopeDeclaring(t *testing.T, declarations string) string {
	t.Helper()
	root := t.TempDir()
	home := filepath.Join(root, "home")
	configPath := filepath.Join(home, "config.json")
	skillsRoot := filepath.Join(root, "skills")
	userHome := filepath.Join(root, "user")
	if err := os.MkdirAll(userHome, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CURATOR_CONFIG", configPath)
	t.Setenv("HOME", userHome)
	t.Setenv("USERPROFILE", userHome)

	if err := config.Bootstrap(configPath, skillsRoot, "", []string{"codex_cli"}, false); err != nil {
		t.Fatal(err)
	}
	if code, stdout, stderr := capture(t, "global", "init"); code != exitOK {
		t.Fatalf("global init = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	writeFile(t, filepath.Join(install.GlobalRoot(home), manifest.Name),
		`{"schema_version":1,"agents":["codex_cli"],"skills":[`+declarations+`]}`)
	return home
}

// compiledGlobalScope declares one skill that exports a schema 6 go-v1 build
// command, so the machine-wide scope activates a compiled command.
func compiledGlobalScope(t *testing.T) string {
	t.Helper()
	home := globalScopeDeclaring(t, `{"name":"build-skill","tag":"v1"}`)
	writeCompiledSkillRepo(t, filepath.Join(filepath.Dir(home), "skills", "build-skill"))
	return home
}

// legacyGlobalScope declares one script-only skill, so the machine-wide scope
// activates no compiled command at all. It pins the pre-existing surface.
func legacyGlobalScope(t *testing.T) string {
	t.Helper()
	home := globalScopeDeclaring(t, `{"name":"skill-a","tag":"v1"}`)
	skill := filepath.Join(filepath.Dir(home), "skills", "skill-a")
	writeFile(t, filepath.Join(skill, "SKILL.md"), "---\nname: skill-a\n---\n# Skill\n")
	runGit(t, skill, "init", "-q", "-b", "main")
	runGit(t, skill, "add", ".")
	runGit(t, skill, "commit", "-qm", "initial skill")
	runGit(t, skill, "tag", "v1")
	return home
}

// captureReport runs one internal reporting phase with both standard streams
// redirected. It is the in-process counterpart of capture: capture proves the
// whole command path, this proves the same phase a command path reaches, driven
// from a plan the test already acquired.
func captureReport(t *testing.T, call func() int) (int, string, string) {
	t.Helper()
	outReader, outWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	errReader, errWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	realOut, realErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = outWriter, errWriter

	var stdout, stderr []byte
	var readers sync.WaitGroup
	readers.Add(2)
	go func() { defer readers.Done(); stdout, _ = io.ReadAll(outReader) }()
	go func() { defer readers.Done(); stderr, _ = io.ReadAll(errReader) }()

	code := call()

	os.Stdout, os.Stderr = realOut, realErr
	_ = outWriter.Close()
	_ = errWriter.Close()
	readers.Wait()
	return code, string(stdout), string(stderr)
}

// globalPlan is one read-only machine-wide plan a test acquired for itself,
// through the exact acquisition phase a command run uses.
type globalPlan struct {
	result     install.Result
	unprovable bool
}

// acquireGlobalPlan runs the read-only machine-wide plan once. Every plan is a
// full fingerprint of the trusted toolchain, so a test acquires one and replays
// every rendering of the same live state from it instead of deriving an
// identical plan per assertion.
func acquireGlobalPlan(t *testing.T) globalPlan {
	t.Helper()
	cfg, code := loadConfig()
	if code != exitOK {
		t.Fatalf("loading the manager configuration = %d", code)
	}
	result, unprovable := globalStatusPlan(cfg)
	if len(result.Builds) == 0 {
		t.Fatalf("the read-only global plan describes no compiled command: %+v", result)
	}
	return globalPlan{result: result, unprovable: unprovable}
}

// replay renders one report from this plan through the same production
// classification, rendering, and fail-closed decision `curator global status`
// runs, so every replayed assertion is an assertion about production behaviour.
func (plan globalPlan) replay(t *testing.T, opts globalStatusOptions) (int, string, string) {
	t.Helper()
	cfg, code := loadConfig()
	if code != exitOK {
		t.Fatalf("loading the manager configuration = %d", code)
	}
	return captureReport(t, func() int {
		return reportGlobalStatus(cfg, opts, func(*config.Config) (install.Result, bool) {
			return plan.result, plan.unprovable
		})
	})
}

// restoreMarkerAfter copies one install marker aside and puts the exact bytes
// back when the test ends.
func restoreMarkerAfter(t *testing.T, installed string) {
	t.Helper()
	path := filepath.Join(installed, marker.Name)
	original, err := os.ReadFile(path) // #nosec G304 -- test fixture path
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if writeErr := os.WriteFile(path, original, 0o644); writeErr != nil {
			t.Fatal(writeErr)
		}
	})
}

// snapshotBuildCacheAfter copies every protected build-cache byte and permission
// bit aside and restores it when the test ends.
//
// A case that damages live cache state resets its own fixture from that
// snapshot rather than running a real machine-wide reconciliation. Reinstalling
// to repair a fixture compiles the command again and proves nothing this file
// owns; install, repair, and rollback semantics keep their own dedicated tests.
func snapshotBuildCacheAfter(t *testing.T, home string) {
	t.Helper()
	base := filepath.Join(home, "cache", "build")
	type node struct {
		relative string
		dir      bool
		mode     os.FileMode
		payload  []byte
	}
	var nodes []node
	err := filepath.WalkDir(base, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, relErr := filepath.Rel(base, path)
		if relErr != nil {
			return relErr
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			return infoErr
		}
		item := node{relative: relative, dir: entry.IsDir(), mode: info.Mode().Perm()}
		if !entry.IsDir() {
			payload, readErr := os.ReadFile(path) // #nosec G304 -- test fixture path
			if readErr != nil {
				return readErr
			}
			item.payload = payload
		}
		nodes = append(nodes, item)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		// The protected boundary is read-only by construction, so the tree is made
		// writable before it is replaced, and every recorded mode is put back
		// afterwards — deepest last-written first — so a restored read-only
		// directory never blocks the write that fills it.
		if clearErr := clearProtection(base); clearErr != nil {
			t.Fatal(clearErr)
		}
		if removeErr := os.RemoveAll(base); removeErr != nil {
			t.Fatal(removeErr)
		}
		for _, item := range nodes {
			path := filepath.Join(base, item.relative)
			if item.dir {
				if mkErr := os.MkdirAll(path, 0o700); mkErr != nil {
					t.Fatal(mkErr)
				}
			} else if writeErr := os.WriteFile(path, item.payload, 0o600); writeErr != nil {
				t.Fatal(writeErr)
			}
			// Mode bits are the whole boundary on unix and none of it on
			// Windows, where a recreated node inherits its parent's DACL and
			// lands outside the protected boundary it was snapshotted inside.
			restoreCacheProtection(t, path, item.dir)
		}
		for index := len(nodes) - 1; index >= 0; index-- {
			if chmodErr := os.Chmod(filepath.Join(base, nodes[index].relative), nodes[index].mode); chmodErr != nil {
				t.Fatal(chmodErr)
			}
		}
	})
}

// clearProtection makes a protected cache tree writable so a snapshot restore
// can replace it. Directories are opened before they are descended into, so a
// read-only entry never hides its own contents from the restore.
func clearProtection(base string) error {
	return filepath.WalkDir(base, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			if os.IsNotExist(walkErr) {
				return nil
			}
			return walkErr
		}
		if entry.IsDir() {
			return os.Chmod(path, 0o700)
		}
		return os.Chmod(path, 0o600)
	})
}

// driftExpectation is one non-current compiled state and everything the whole
// reporting contract must say about it.
type driftExpectation struct {
	state   string
	cause   string
	outcome string
}

// assertGlobalDriftDocument proves the machine-readable half of one non-current
// compiled state, together with the fail-closed exit the same invocation
// produced.
func assertGlobalDriftDocument(t *testing.T, want driftExpectation, code int, stdout string) {
	t.Helper()
	if code != exitFail {
		t.Fatalf("global status --json --check over %s = %d, want %d\n%s", want.state, code, exitFail, stdout)
	}
	drifted := decodeGlobalStatus(t, stdout)
	if len(drifted.Builds) != 1 {
		t.Fatalf("global status reported %d compiled commands:\n%s", len(drifted.Builds), stdout)
	}
	row := drifted.Builds[0]
	if row.State != want.state {
		t.Fatalf("state = %q, want %q (row %+v)", row.State, want.state, row)
	}
	if row.Cause != want.cause {
		t.Fatalf("cause = %q, want %q (row %+v)", row.Cause, want.cause, row)
	}
	if row.CacheOutcome != want.outcome {
		t.Fatalf("cache outcome = %q, want %q", row.CacheOutcome, want.outcome)
	}
	if row.Detail == "" {
		t.Fatalf("non-current row %+v carries no operator detail", row)
	}
	if drifted.Skills["build-skill"] != want.state {
		t.Fatalf("skills = %v, want build-skill %q", drifted.Skills, want.state)
	}
}

// assertGlobalDriftHuman proves the operator half of the same state, and that
// the plain report still exits zero: reporting a verdict is not itself a
// failure, and `--check` is the only surface that turns one into an exit code.
func assertGlobalDriftHuman(t *testing.T, want driftExpectation, code int, human string) {
	t.Helper()
	if code != exitOK {
		t.Fatalf("plain global status over %s = %d, want %d", want.state, code, exitOK)
	}
	if !strings.Contains(human, "global: build-skill.build-tool build ") ||
		!strings.Contains(human, "state="+want.state) {
		t.Fatalf("human global status does not report %q:\n%s", want.state, human)
	}
	if want.cause != "" && !strings.Contains(human, "cause="+want.cause) {
		t.Fatalf("human global status does not report cause %q:\n%s", want.cause, human)
	}
	if !strings.Contains(human, "cache="+want.outcome) {
		t.Fatalf("human global status does not report cache outcome %q:\n%s", want.outcome, human)
	}
}

// TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck is the end-to-end
// proof of the machine-wide compiled surface: an unchanged compiled global
// installation reports every active command as current, and each independent way
// that state can stop being current produces its own stable code — the same
// codes `curator status` reports — and a non-zero `global status --check`.
//
// Every case proves the same complete contract: the machine-readable row, the
// demoted declared-skill entry, the fail-closed exit, the operator line, and the
// zero exit of the plain report. What differs is where the plan under
// classification comes from. A case the read-only plan cannot observe — a
// rewritten install marker, a build root materialized in the installed skill —
// is replayed from one plan acquired while the installation was unchanged. A
// case the plan does observe — protected cache state — acquires its own plan
// while that state is live, because a replayed one would describe cache
// evidence that no longer exists.
func TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck(t *testing.T) {
	requireNativeControlInventoryPlatform(t)
	home := compiledGlobalScope(t)
	if code, stdout, stderr := capture(t, "global", "install"); code != exitOK {
		t.Fatalf("global install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	installed := filepath.Join(install.GlobalRoot(home), "skills", "build-skill")

	// `--json` and `--check` are combined so one invocation proves the published
	// document and the exit code it produced are the same run's verdict.
	code, stdout, stderr := capture(t, "global", "status", "--json", "--check")
	if code != exitOK {
		t.Fatalf("clean compiled global status --json --check = %d\nstderr:\n%s", code, stderr)
	}
	report := decodeGlobalStatus(t, stdout)
	if report.Alias != "global" {
		t.Fatalf("alias = %q, want %q", report.Alias, "global")
	}
	if len(report.Builds) != 1 {
		t.Fatalf("global status reported %d compiled commands, want 1:\n%s", len(report.Builds), stdout)
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
	// The machine-wide scope has no operator-supplied root, so no manager-owned
	// location may reach the machine-readable surface at all.
	if strings.Contains(stdout, home) {
		t.Fatalf("global status --json published the manager home %q:\n%s", home, stdout)
	}

	// One plan of the unchanged installation, acquired once and reused by every
	// case below that cannot change what a plan derives.
	unchanged := acquireGlobalPlan(t)

	for _, testCase := range []struct {
		name     string
		tamper   func(t *testing.T)
		want     driftExpectation
		endToEnd bool
	}{
		{
			name: "recorded build-source identity no longer matches the frozen snapshot",
			tamper: func(t *testing.T) {
				rewriteMarker(t, installed, func(object map[string]any) {
					source := object["build_source"].(map[string]any)
					source["content_sha256"] = testDigest(5)
				})
			},
			want:     driftExpectation{state: buildSourceDrift, outcome: "cache-hit"},
			endToEnd: true,
		},
		{
			name: "logical key recorded by the marker was derived from another build input",
			tamper: func(t *testing.T) {
				rewriteMarker(t, installed, func(object map[string]any) {
					builds := object["builds"].(map[string]any)
					build := builds["build-tool"].(map[string]any)
					build["cache_key"] = testDigest(5)
				})
			},
			want: driftExpectation{state: buildInputDrift, cause: causeUnattributed, outcome: "cache-hit"},
		},
		{
			name: "marker records no build for the command the closure activates",
			tamper: func(t *testing.T) {
				rewriteMarker(t, installed, func(object map[string]any) {
					object["builds"] = map[string]any{}
					delete(object, "build_source")
				})
			},
			want: driftExpectation{state: buildCommandDrift, outcome: "cache-hit"},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			restoreMarkerAfter(t, installed)
			testCase.tamper(t)

			code, stdout, _ := unchanged.replay(t, globalStatusOptions{jsonOut: true, check: true})
			assertGlobalDriftDocument(t, testCase.want, code, stdout)
			if testCase.endToEnd {
				// One case in this group also runs the whole command path, so the
				// replayed classification is pinned to what the CLI publishes.
				code, stdout, _ = capture(t, "global", "status", "--json", "--check")
				assertGlobalDriftDocument(t, testCase.want, code, stdout)
			}

			code, human, _ := unchanged.replay(t, globalStatusOptions{})
			assertGlobalDriftHuman(t, testCase.want, code, human)
		})
	}

	for _, testCase := range []struct {
		name   string
		tamper func(t *testing.T)
		want   driftExpectation
	}{
		{
			name:   "protected cache entry cannot be interpreted",
			tamper: func(t *testing.T) { corruptCacheArtifact(t, home) },
			want:   driftExpectation{state: buildCorruptCache, outcome: "corrupt"},
		},
		{
			name: "protected cache holds no entry for the recorded key",
			tamper: func(t *testing.T) {
				for _, entry := range cacheEntries(t, home) {
					if err := os.RemoveAll(entry); err != nil {
						t.Fatal(err)
					}
				}
			},
			want: driftExpectation{state: buildMissingArtifact, outcome: "would-preflight-and-build"},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			snapshotBuildCacheAfter(t, home)
			testCase.tamper(t)

			// Protected cache evidence is read by planning itself, so this state is
			// classified from a plan acquired while it is live.
			drifted := acquireGlobalPlan(t)
			code, stdout, _ := drifted.replay(t, globalStatusOptions{jsonOut: true, check: true})
			assertGlobalDriftDocument(t, testCase.want, code, stdout)

			code, human, _ := drifted.replay(t, globalStatusOptions{})
			assertGlobalDriftHuman(t, testCase.want, code, human)
		})
	}

	t.Run("build root reached agent-facing context", func(t *testing.T) {
		want := driftExpectation{state: buildContextExposed, outcome: "cache-hit"}
		exposed := filepath.Join(installed, "assets", "build-tool")
		if err := os.MkdirAll(exposed, 0o755); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := os.RemoveAll(exposed); err != nil {
				t.Fatal(err)
			}
		})

		// The exposure is in the installed skill, not in the plan, so the replay
		// classifies it as faithfully as a fresh plan would — and the whole
		// command path is run once here to pin that to what the CLI publishes.
		code, stdout, _ := unchanged.replay(t, globalStatusOptions{jsonOut: true, check: true})
		assertGlobalDriftDocument(t, want, code, stdout)
		code, stdout, _ = capture(t, "global", "status", "--json", "--check")
		assertGlobalDriftDocument(t, want, code, stdout)

		code, human, _ := unchanged.replay(t, globalStatusOptions{})
		assertGlobalDriftHuman(t, want, code, human)
	})

	// A refused toolchain is classified per active compiled command like every
	// other state, so it reuses this installation instead of building a second
	// identical one. The plan cannot be replayed here: refusing the toolchain is
	// exactly what the acquisition phase has to report.
	t.Run("trusted go toolchain cannot be resolved", func(t *testing.T) {
		t.Setenv(godriver.SelectionCuratorGo, filepath.Join(t.TempDir(), "nowhere", "bin", "go"))

		code, stdout, stderr := capture(t, "global", "status", "--json", "--check")
		if code != exitFail {
			t.Fatalf("global status --json --check with an unusable toolchain = %d, want %d\n%s",
				code, exitFail, stderr)
		}
		if !strings.Contains(stderr, "warning:") || strings.Contains(stderr, "error:") {
			t.Fatalf("global status reported the refusal as a failure rather than a warning:\n%s", stderr)
		}
		for _, want := range append([]string{
			godriver.SelectionCuratorGo, godriver.SelectionGOROOT,
			"Curator never searches PATH and never downloads a toolchain",
		}, godriver.TestedFamilies()...) {
			if !strings.Contains(stderr, want) {
				t.Fatalf("global toolchain diagnostic does not name %q:\n%s", want, stderr)
			}
		}
		refused := decodeGlobalStatus(t, stdout)
		if len(refused.Builds) != 1 {
			t.Fatalf("global status reported %d compiled commands, want 1:\n%s", len(refused.Builds), stdout)
		}
		row := refused.Builds[0]
		if row.State != buildUnusableToolchain || row.Cause == "" {
			t.Fatalf("toolchain row = %+v", row)
		}
		if row.Skill != "build-skill" || row.Command != "build-tool" || row.Detail == "" {
			t.Fatalf("toolchain row does not name the active command: %+v", row)
		}
		// Nothing the plan could not derive may be published as if it were known.
		if row.CacheKey != "" || row.Target != "" || row.ArtifactPath != "" {
			t.Fatalf("toolchain row published identities the plan never derived: %+v", row)
		}
		if refused.Skills["build-skill"] != buildUnusableToolchain {
			t.Fatalf("skills = %v", refused.Skills)
		}
		if code, human, _ := capture(t, "global", "status"); code != exitOK ||
			!strings.Contains(human, "state="+buildUnusableToolchain) {
			t.Fatalf("plain global status over an unusable toolchain = %d\n%s", code, human)
		}
	})
}

// TestGlobalStatusReportsATransitivelyResolvedCompiledCommand proves the
// machine-wide scope resolves the installed directory of a node no declaration
// names. The global scope reads one store — it consults no hybrid store, since
// hybrid declarations activate against a project — so a provider reached only
// through a dependency must still be found, classified, and able to fail
// `--check` on its own.
func TestGlobalStatusReportsATransitivelyResolvedCompiledCommand(t *testing.T) {
	requireNativeControlInventoryPlatform(t)
	home := globalScopeDeclaring(t, `{"name":"consumer","tag":"v1"}`)
	skillsRoot := filepath.Join(filepath.Dir(home), "skills")
	writeCompiledSkillRepo(t, filepath.Join(skillsRoot, "build-skill"))

	consumer := filepath.Join(skillsRoot, "consumer")
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

	if code, stdout, stderr := capture(t, "global", "install"); code != exitOK {
		t.Fatalf("global install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}

	code, stdout, stderr := capture(t, "global", "status", "--json", "--check")
	if code != exitOK {
		t.Fatalf("clean transitive compiled global status --json --check = %d\nstderr:\n%s", code, stderr)
	}
	report := decodeGlobalStatus(t, stdout)
	if len(report.Builds) != 1 || report.Builds[0].Skill != "build-skill" ||
		report.Builds[0].State != buildCurrent {
		t.Fatalf("transitive compiled command was not reported:\n%s", stdout)
	}
	if _, declared := report.Skills["build-skill"]; declared {
		t.Fatalf("the provider is a transitive node, not a declaration: %v", report.Skills)
	}

	// The provider's own compiled state must still fail --check on its own,
	// through the drift surface no declared-skill map can see.
	for _, entry := range cacheEntries(t, home) {
		if err := os.RemoveAll(entry); err != nil {
			t.Fatal(err)
		}
	}
	code, stdout, _ = capture(t, "global", "status", "--json", "--check")
	if code != exitFail {
		t.Fatalf("transitive compiled drift passed global status --check: %d\n%s", code, stdout)
	}
	if drifted := decodeGlobalStatus(t, stdout); len(drifted.Builds) != 1 ||
		drifted.Builds[0].State != buildMissingArtifact {
		t.Fatalf("transitive drift was not reported:\n%s", stdout)
	}
}

// TestGlobalStatusKeepsTheDeclaredSkillSurfaceWithoutCompiledCommands pins the
// compatibility boundary: a machine-wide scope that activates no compiled
// command prints exactly the lines it printed before this surface existed, adds
// no `builds` key to the machine-readable document, and still exits zero.
func TestGlobalStatusKeepsTheDeclaredSkillSurfaceWithoutCompiledCommands(t *testing.T) {
	home := legacyGlobalScope(t)
	if code, stdout, stderr := capture(t, "global", "install"); code != exitOK {
		t.Fatalf("global install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}

	code, stdout, stderr := capture(t, "global", "status")
	if code != exitOK {
		t.Fatalf("global status = %d\nstderr:\n%s", code, stderr)
	}
	if stdout != "global: skill-a "+stateUpToDate+"\n" {
		t.Fatalf("global status changed the pre-existing declared-skill output:\n%q", stdout)
	}
	if stderr != "" {
		t.Fatalf("a clean global status wrote to standard error:\n%s", stderr)
	}
	if code, _, _ := capture(t, "global", "status", "--check"); code != exitOK {
		t.Fatalf("clean global status --check = %d, want %d", code, exitOK)
	}

	code, stdout, _ = capture(t, "global", "status", "--json")
	if code != exitOK {
		t.Fatalf("global status --json = %d", code)
	}
	var object map[string]any
	if err := json.Unmarshal([]byte(stdout), &object); err != nil {
		t.Fatalf("global status --json is not one JSON object: %v\n%s", err, stdout)
	}
	if len(object) != 2 || object["alias"] != "global" || object["skills"] == nil {
		t.Fatalf("global status --json is not the declared-skill document:\n%s", stdout)
	}
	if strings.Contains(stdout, home) {
		t.Fatalf("global status --json published the manager home %q:\n%s", home, stdout)
	}

	// Ordinary declared-skill drift keeps its own more actionable code and still
	// fails the new fail-closed verdict.
	installed := filepath.Join(install.GlobalRoot(home), "skills", "skill-a")
	writeFile(t, filepath.Join(installed, "SKILL.md"), "tampered\n")
	code, stdout, _ = capture(t, "global", "status")
	if code != exitOK || stdout != "global: skill-a "+stateContentDrift+"\n" {
		t.Fatalf("global status over tampered content = %d\n%s", code, stdout)
	}
	if code, _, _ := capture(t, "global", "status", "--check"); code != exitFail {
		t.Fatalf("global status --check over tampered content = %d, want %d", code, exitFail)
	}
}

// TestGlobalStatusWithoutASkillfileStaysSilentAndCurrent proves the scope that
// declares nothing keeps its historical behaviour: no output, exit zero, and
// nothing for the fail-closed verdict to refuse.
func TestGlobalStatusWithoutASkillfileStaysSilentAndCurrent(t *testing.T) {
	root := t.TempDir()
	userHome := filepath.Join(root, "user")
	if err := os.MkdirAll(userHome, 0o755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(root, "home", "config.json")
	t.Setenv("CURATOR_CONFIG", configPath)
	t.Setenv("HOME", userHome)
	t.Setenv("USERPROFILE", userHome)
	if err := config.Bootstrap(configPath, filepath.Join(root, "skills"), "", []string{"codex_cli"}, false); err != nil {
		t.Fatal(err)
	}

	for _, args := range [][]string{
		{"global", "status"},
		{"global", "status", "--check"},
	} {
		code, stdout, stderr := capture(t, args...)
		if code != exitOK {
			t.Fatalf("%v = %d, want %d\nstderr:\n%s", args, code, exitOK, stderr)
		}
		if stdout != "" {
			t.Fatalf("%v printed a report for a scope that declares nothing:\n%q", args, stdout)
		}
	}
}

// TestGlobalStatusFailsCheckWhenTheUserHomeCannotBeResolved proves the
// machine-wide plan cannot silently call the scope current when it cannot
// derive the user home whose adapters and forwarding shims it must inspect.
// Plain reporting keeps the historical zero exit; the fail-closed verdict does
// not.
func TestGlobalStatusFailsCheckWhenTheUserHomeCannotBeResolved(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "home", "config.json")
	t.Setenv("CURATOR_CONFIG", configPath)
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")
	if err := config.Bootstrap(configPath, filepath.Join(root, "skills"), "", []string{"codex_cli"}, false); err != nil {
		t.Fatal(err)
	}

	code, stdout, stderr := capture(t, "global", "status")
	if code != exitOK {
		t.Fatalf("global status without a resolvable user home = %d, want %d", code, exitOK)
	}
	if stdout != "" {
		t.Fatalf("global status invented declarations without a user home:\n%q", stdout)
	}
	if !strings.Contains(stderr, "warning:") ||
		!strings.Contains(stderr, "could not resolve the user home") {
		t.Fatalf("global status did not explain the unprovable scope:\n%s", stderr)
	}
	if code, _, _ := capture(t, "global", "status", "--check"); code != exitFail {
		t.Fatalf("global status --check without a resolvable user home = %d, want %d", code, exitFail)
	}
}

// TestGlobalStatusFailsCheckWhenTheClosureCannotBeProven proves the fail-closed
// half of the contract: a plan that refuses before it can describe compiled
// state still publishes the declared-skill report and still exits zero, but
// `--check` refuses to call an unprovable scope current.
func TestGlobalStatusFailsCheckWhenTheClosureCannotBeProven(t *testing.T) {
	home := legacyGlobalScope(t)
	if code, _, stderr := capture(t, "global", "install"); code != exitOK {
		t.Fatalf("global install = %d\n%s", code, stderr)
	}
	// A declaration that cannot be resolved fails the read-only plan long before
	// any build phase, so this run cannot prove the scope activates no compiled
	// command.
	writeFile(t, filepath.Join(install.GlobalRoot(home), manifest.Name),
		`{"schema_version":1,"agents":["codex_cli"],"skills":[{"name":"skill-a","tag":"v9"}]}`)

	code, stdout, stderr := capture(t, "global", "status")
	if code != exitOK {
		t.Fatalf("global status over an unresolvable declaration = %d, want %d", code, exitOK)
	}
	if stdout != "global: skill-a "+stateUnresolvable+"\n" {
		t.Fatalf("global status suppressed the declared-skill report:\n%q", stdout)
	}
	if !strings.Contains(stderr, "warning:") {
		t.Fatalf("global status did not report the refusal as a warning:\n%s", stderr)
	}
	if code, _, _ := capture(t, "global", "status", "--check"); code != exitFail {
		t.Fatalf("global status --check over an unprovable closure = %d, want %d", code, exitFail)
	}
}

// TestGlobalStatusRejectsPositionalArguments proves the machine-wide scope
// takes no target: it is one scope, and a stray path must not be silently
// ignored.
func TestGlobalStatusRejectsPositionalArguments(t *testing.T) {
	legacyGlobalScope(t)
	if code, _, _ := capture(t, "global", "status", "somewhere"); code != exitUsage {
		t.Fatalf("global status with a positional argument = %d, want %d", code, exitUsage)
	}
}
