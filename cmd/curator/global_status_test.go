package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

// reinstallGlobal runs the machine-wide reconciliation path, which is also the
// only repair path the global scope has.
func reinstallGlobal(t *testing.T) {
	t.Helper()
	if code, stdout, stderr := capture(t, "global", "install"); code != exitOK {
		t.Fatalf("restoring global install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
}

// TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck is the end-to-end
// proof of the machine-wide compiled surface: an unchanged compiled global
// installation reports every active command as current, and each independent
// way that state can stop being current produces its own stable code — the same
// codes `curator status` reports — and a non-zero `global status --check`.
func TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck(t *testing.T) {
	home := compiledGlobalScope(t)
	if code, stdout, stderr := capture(t, "global", "install"); code != exitOK {
		t.Fatalf("global install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	installed := filepath.Join(install.GlobalRoot(home), "skills", "build-skill")

	code, stdout, stderr := capture(t, "global", "status", "--json")
	if code != exitOK {
		t.Fatalf("global status --json = %d\nstderr:\n%s", code, stderr)
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
	if current.Skill != "build-skill" || current.Command != "build-tool" ||
		current.Driver != "go-v1" || current.BuildRoot != "assets/build-tool" ||
		current.SourceDir != "assets/build-tool/cmd/tool" || current.CacheKey == "" ||
		current.ArtifactPath != "bin/build-tool" || current.Target == "" ||
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
	if code, _, _ := capture(t, "global", "status", "--check"); code != exitOK {
		t.Fatalf("clean compiled global status --check = %d, want %d", code, exitOK)
	}

	for name, testCase := range map[string]struct {
		tamper  func(t *testing.T)
		restore func(t *testing.T)
		want    string
		cause   string
		outcome string
	}{
		"recorded build-source identity no longer matches the frozen snapshot": {
			tamper: func(t *testing.T) {
				rewriteMarker(t, installed, func(object map[string]any) {
					source := object["build_source"].(map[string]any)
					source["content_sha256"] = testDigest(5)
				})
			},
			want: buildSourceDrift, outcome: "cache-hit",
		},
		"logical key recorded by the marker was derived from another build input": {
			tamper: func(t *testing.T) {
				rewriteMarker(t, installed, func(object map[string]any) {
					builds := object["builds"].(map[string]any)
					build := builds["build-tool"].(map[string]any)
					build["cache_key"] = testDigest(5)
				})
			},
			want: buildInputDrift, cause: causeUnattributed, outcome: "cache-hit",
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
		"protected cache entry cannot be interpreted": {
			tamper:  func(t *testing.T) { corruptCacheArtifact(t, home) },
			restore: reinstallGlobal,
			want:    buildCorruptCache, outcome: "corrupt",
		},
		"protected cache holds no entry for the recorded key": {
			tamper: func(t *testing.T) {
				for _, entry := range cacheEntries(t, home) {
					if err := os.RemoveAll(entry); err != nil {
						t.Fatal(err)
					}
				}
			},
			restore: reinstallGlobal,
			want:    buildMissingArtifact, outcome: "would-preflight-and-build",
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
	} {
		t.Run(name, func(t *testing.T) {
			original, err := os.ReadFile(filepath.Join(installed, marker.Name)) // #nosec G304 -- test fixture path
			if err != nil {
				t.Fatal(err)
			}
			testCase.tamper(t)
			t.Cleanup(func() {
				// Live state is restored before the marker, so a restore that runs
				// a real reconciliation cannot leave a rewritten marker behind.
				if testCase.restore != nil {
					testCase.restore(t)
				}
				if writeErr := os.WriteFile(filepath.Join(installed, marker.Name), original, 0o644); writeErr != nil {
					t.Fatal(writeErr)
				}
			})

			code, stdout, stderr := capture(t, "global", "status", "--json")
			if code != exitOK {
				t.Fatalf("global status --json = %d\nstderr:\n%s", code, stderr)
			}
			drifted := decodeGlobalStatus(t, stdout)
			if len(drifted.Builds) != 1 {
				t.Fatalf("global status reported %d compiled commands:\n%s", len(drifted.Builds), stdout)
			}
			row := drifted.Builds[0]
			if row.State != testCase.want {
				t.Fatalf("state = %q, want %q (row %+v)", row.State, testCase.want, row)
			}
			if row.Cause != testCase.cause {
				t.Fatalf("cause = %q, want %q (row %+v)", row.Cause, testCase.cause, row)
			}
			if row.CacheOutcome != testCase.outcome {
				t.Fatalf("cache outcome = %q, want %q", row.CacheOutcome, testCase.outcome)
			}
			if row.Detail == "" {
				t.Fatalf("non-current row %+v carries no operator detail", row)
			}
			if drifted.Skills["build-skill"] != testCase.want {
				t.Fatalf("skills = %v, want build-skill %q", drifted.Skills, testCase.want)
			}

			if code, _, _ := capture(t, "global", "status", "--check"); code != exitFail {
				t.Fatalf("global status --check on %s = %d, want %d", name, code, exitFail)
			}
			code, human, _ := capture(t, "global", "status")
			if code != exitOK || !strings.Contains(human, "global: build-skill.build-tool build ") ||
				!strings.Contains(human, "state="+testCase.want) {
				t.Fatalf("human global status does not report %q:\n%s", testCase.want, human)
			}
			if testCase.cause != "" && !strings.Contains(human, "cause="+testCase.cause) {
				t.Fatalf("human global status does not report cause %q:\n%s", testCase.cause, human)
			}
		})
	}
}

// TestGlobalStatusReportsAnUnusableToolchainPerCompiledCommand proves the
// machine-wide scope reports a refused toolchain the same way a project does:
// one machine-readable row per active compiled command carrying the stable
// go-v1 boundary code, the guidance on standard error as a warning, and a
// non-zero `--check`.
func TestGlobalStatusReportsAnUnusableToolchainPerCompiledCommand(t *testing.T) {
	compiledGlobalScope(t)
	if code, stdout, stderr := capture(t, "global", "install"); code != exitOK {
		t.Fatalf("global install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	t.Setenv(godriver.SelectionCuratorGo, filepath.Join(t.TempDir(), "nowhere", "bin", "go"))

	code, stdout, stderr := capture(t, "global", "status", "--json")
	if code != exitOK {
		t.Fatalf("global status with an unusable toolchain = %d, want %d\n%s", code, exitOK, stderr)
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
	report := decodeGlobalStatus(t, stdout)
	if len(report.Builds) != 1 {
		t.Fatalf("global status reported %d compiled commands, want 1:\n%s", len(report.Builds), stdout)
	}
	row := report.Builds[0]
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
	if report.Skills["build-skill"] != buildUnusableToolchain {
		t.Fatalf("skills = %v", report.Skills)
	}
	if code, _, _ := capture(t, "global", "status", "--check"); code != exitFail {
		t.Fatalf("global status --check with an unusable toolchain = %d, want %d", code, exitFail)
	}
}

// TestGlobalStatusReportsATransitivelyResolvedCompiledCommand proves the
// machine-wide scope resolves the installed directory of a node no declaration
// names. The global scope reads one store — it consults no hybrid store, since
// hybrid declarations activate against a project — so a provider reached only
// through a dependency must still be found, classified, and able to fail
// `--check` on its own.
func TestGlobalStatusReportsATransitivelyResolvedCompiledCommand(t *testing.T) {
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

	code, stdout, stderr := capture(t, "global", "status", "--json")
	if code != exitOK {
		t.Fatalf("global status --json = %d\nstderr:\n%s", code, stderr)
	}
	report := decodeGlobalStatus(t, stdout)
	if len(report.Builds) != 1 || report.Builds[0].Skill != "build-skill" ||
		report.Builds[0].State != buildCurrent {
		t.Fatalf("transitive compiled command was not reported:\n%s", stdout)
	}
	if _, declared := report.Skills["build-skill"]; declared {
		t.Fatalf("the provider is a transitive node, not a declaration: %v", report.Skills)
	}
	if code, _, _ := capture(t, "global", "status", "--check"); code != exitOK {
		t.Fatalf("clean transitive compiled global status --check = %d, want %d", code, exitOK)
	}

	// The provider's own compiled state must still fail --check on its own,
	// through the drift surface no declared-skill map can see.
	for _, entry := range cacheEntries(t, home) {
		if err := os.RemoveAll(entry); err != nil {
			t.Fatal(err)
		}
	}
	code, stdout, _ = capture(t, "global", "status", "--json")
	if code != exitOK {
		t.Fatalf("global status --json after the entry was removed = %d", code)
	}
	if drifted := decodeGlobalStatus(t, stdout); len(drifted.Builds) != 1 ||
		drifted.Builds[0].State != buildMissingArtifact {
		t.Fatalf("transitive drift was not reported:\n%s", stdout)
	}
	if code, _, _ := capture(t, "global", "status", "--check"); code != exitFail {
		t.Fatalf("transitive compiled drift passed global status --check: %d", code)
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
