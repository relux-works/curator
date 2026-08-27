package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/relux-works/curator/internal/buildcache"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/godriver"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
)

// TestMain gives the test binary the same fixed hidden worker mode the
// installed manager dispatches before command parsing, so a compiled
// installation in these tests runs the real identity-verified process graph
// instead of a mock.
func TestMain(m *testing.M) {
	if len(os.Args) == 2 && os.Args[1] == godriver.WorkerMode {
		os.Exit(godriver.RunWorker(os.Stdin, os.Stdout))
	}
	os.Exit(m.Run())
}

// capture runs one CLI invocation with both standard streams redirected.
func capture(t *testing.T, args ...string) (int, string, string) {
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

	code := run(args)

	os.Stdout, os.Stderr = realOut, realErr
	_ = outWriter.Close()
	_ = errWriter.Close()
	readers.Wait()
	return code, string(stdout), string(stderr)
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
	t.Setenv("CURATOR_CONFIG", configPath)
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
func legacyProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	configPath := filepath.Join(root, "home", "config.json")
	t.Setenv("CURATOR_CONFIG", configPath)
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
	return project
}

// reinstall runs the ordinary reconciliation path, which is also the only
// repair path Curator has.
func reinstall(t *testing.T) {
	t.Helper()
	if code, stdout, stderr := capture(t, "install", "app"); code != exitOK {
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

// TestStatusReportsCompiledCurrentnessAndFailsCheck is the end-to-end proof of
// the compiled currentness surface: a real installation reports every planned
// command as current, and each independent way that state can stop being
// current produces its own stable code and a non-zero `status --check`.
func TestStatusReportsCompiledCurrentnessAndFailsCheck(t *testing.T) {
	project, home := compiledProject(t)
	if code, stdout, stderr := capture(t, "install", "app"); code != exitOK {
		t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	installed := filepath.Join(project, ".agents", "skills", "build-skill")

	code, stdout, stderr := capture(t, "status", "app", "--json")
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
	// No manager-private location may reach the machine-readable surface. The
	// project path is the operator's own argument and stays.
	for _, forbidden := range []string{home, filepath.Join(project, ".agents", "bin")} {
		if strings.Contains(marshal(t, report.Builds), forbidden) {
			t.Fatalf("build rows published the private location %q:\n%s", forbidden, stdout)
		}
	}
	if code, _, _ := capture(t, "status", "app", "--check"); code != exitOK {
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
	}{
		"artifact hash recorded by the marker no longer matches the entry": {
			tamper: func(t *testing.T) {
				rewriteMarker(t, installed, func(object map[string]any) {
					builds := object["builds"].(map[string]any)
					build := builds["build-tool"].(map[string]any)
					build["artifact_sha256"] = testDigest(5)
				})
			},
			want: buildArtifactDrift, outcome: "cache-hit",
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
			want: buildInputDrift, cause: causeUnattributed, outcome: "cache-hit",
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
					build["artifact_path"] = "bin/build-tool.exe"
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
		// The two cases below mutate the shared protected cache rather than the
		// marker, so each rebuilds it through the ordinary reconciliation path
		// before the next case runs.
		"protected cache entry cannot be interpreted": {
			tamper:  func(t *testing.T) { corruptCacheArtifact(t, home) },
			restore: reinstall,
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
			restore: reinstall,
			want:    buildMissingArtifact, outcome: "would-preflight-and-build",
		},
		"marker schema cannot be read by this manager": {
			tamper: func(t *testing.T) {
				object := markerPayload(t, installed)
				object["schema_version"] = marker.SchemaVersion + 1
				refuseMarker(t, installed, marshal(t, object))
			},
			want: stateUnsupportedMarker, outcome: "cache-hit",
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
					if err := os.Chmod(entry, 0o777); err != nil {
						t.Fatal(err)
					}
				}
			},
			restore: func(t *testing.T) {
				for _, entry := range cacheEntries(t, home) {
					if err := os.Chmod(entry, 0o700); err != nil {
						t.Fatal(err)
					}
				}
			},
			want: buildUntrustedCache, outcome: "would-rebuild-untrusted-cache",
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

			code, stdout, stderr := capture(t, "status", "app", "--json")
			if code != exitOK {
				t.Fatalf("status --json = %d\nstderr:\n%s", code, stderr)
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

			if code, _, _ := capture(t, "status", "app", "--check"); code != exitFail {
				t.Fatalf("status --check on %s = %d, want %d", name, code, exitFail)
			}
			code, human, _ := capture(t, "status", "app")
			if code != exitOK || !strings.Contains(human, "state="+testCase.want) {
				t.Fatalf("human status does not report %q:\n%s", testCase.want, human)
			}
			if testCase.cause != "" && !strings.Contains(human, "cause="+testCase.cause) {
				t.Fatalf("human status does not report cause %q:\n%s", testCase.cause, human)
			}
		})
	}
}

// TestInstallAndUpgradeRepairCorruptCompiledState proves the reconciliation
// path for a protected entry that cannot be interpreted at all: corrupt
// receipt bytes and corrupt artifact bytes are rebuilt by install and by
// upgrade, only after every gate has passed, and a run that fails leaves the
// previous installation and the live cache exactly as they were.
func TestInstallAndUpgradeRepairCorruptCompiledState(t *testing.T) {
	for _, command := range []string{"install", "upgrade"} {
		t.Run(command, func(t *testing.T) {
			project, home := compiledProject(t)
			if code, stdout, stderr := capture(t, "install", "app"); code != exitOK {
				t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
			}
			installed := filepath.Join(project, ".agents", "skills", "build-skill")
			before := marker.Read(installed)
			if before == nil || len(before.Builds) != 1 {
				t.Fatalf("first install did not record compiled state: %+v", before)
			}

			for _, corruption := range []struct {
				name    string
				corrupt func(t *testing.T, home string)
			}{
				{"receipt bytes", corruptCacheReceipt},
				{"artifact bytes", corruptCacheArtifact},
			} {
				t.Run(corruption.name, func(t *testing.T) {
					corruption.corrupt(t, home)
					if code, stdout, _ := capture(t, "status", "app", "--json"); code != exitOK ||
						decodeStatus(t, stdout).Builds[0].State != buildCorruptCache {
						t.Fatalf("status did not report corrupt compiled state:\n%s", stdout)
					}

					// A gate that fails must refuse before the repair: the corrupt
					// entry, the installed marker, and the installed content are all
					// still exactly as this run found them.
					corruptedCache := cacheFingerprint(t, home)
					t.Setenv(godriver.SelectionCuratorGo, filepath.Join(t.TempDir(), "nowhere", "bin", "go"))
					if code, stdout, _ := capture(t, command, "app"); code != exitFail {
						t.Fatalf("%s without a usable toolchain = %d, want %d\n%s", command, code, exitFail, stdout)
					}
					if cacheFingerprint(t, home) != corruptedCache {
						t.Fatal("a refused run changed the live build cache")
					}
					if refused := marker.Read(installed); refused == nil ||
						refused.Builds["build-tool"] != before.Builds["build-tool"] {
						t.Fatalf("a refused run changed the previous compiled state: %+v", refused)
					}
					if _, err := os.Stat(filepath.Join(installed, "SKILL.md")); err != nil {
						t.Fatalf("a refused run removed the previous installation: %v", err)
					}
					t.Setenv(godriver.SelectionCuratorGo, "")

					code, stdout, stderr := capture(t, command, "app")
					if code != exitOK {
						t.Fatalf("repairing %s = %d\nstdout:\n%s\nstderr:\n%s", command, code, stdout, stderr)
					}
					if !strings.Contains(stdout, "rebuilt corrupt build cache state into a new protected entry") {
						t.Fatalf("%s did not report the repair:\n%s", command, stdout)
					}
					if code, _, _ := capture(t, "status", "app", "--check"); code != exitOK {
						t.Fatalf("status --check after repairing %s = %d, want %d", corruption.name, code, exitOK)
					}
					after := marker.Read(installed)
					if after == nil || after.Builds["build-tool"].CacheKey != before.Builds["build-tool"].CacheKey {
						t.Fatalf("a repair changed the logical key: %+v", after)
					}
					if len(cacheEntries(t, home)) != 1 {
						t.Fatalf("repair left %d live protected entries, want 1", len(cacheEntries(t, home)))
					}
				})
			}
		})
	}
}

// TestStatusReportsATransitivelyResolvedCompiledCommand proves the compiled
// surface is not limited to project declarations: a build that belongs to a
// node the project reaches only through a dependency is reported, classified,
// and fails --check on its own, even though no declaration names it.
func TestStatusReportsATransitivelyResolvedCompiledCommand(t *testing.T) {
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

	if code, stdout, stderr := capture(t, "install", "app"); code != exitOK {
		t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}

	code, stdout, stderr := capture(t, "status", "app", "--json")
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
	if code, _, _ := capture(t, "status", "app", "--check"); code != exitOK {
		t.Fatalf("clean transitive compiled status --check = %d, want %d", code, exitOK)
	}

	// The provider's own compiled state must still fail --check on its own,
	// through the drift surface no declared-skill map can see.
	for _, entry := range cacheEntries(t, home) {
		if err := os.RemoveAll(entry); err != nil {
			t.Fatal(err)
		}
	}
	code, stdout, _ = capture(t, "status", "app", "--json")
	if code != exitOK {
		t.Fatalf("status --json after the entry was removed = %d", code)
	}
	if drifted := decodeStatus(t, stdout); len(drifted.Builds) != 1 ||
		drifted.Builds[0].State != buildMissingArtifact {
		t.Fatalf("transitive drift was not reported:\n%s", stdout)
	}
	if code, _, _ := capture(t, "status", "app", "--check"); code != exitFail {
		t.Fatalf("transitive compiled drift passed --check: %d", code)
	}
}

// TestStatusReportsAnUnusableToolchainPerCompiledCommand proves the
// missing-or-incompatible Go diagnostic is not stderr-only: every active
// compiled command still gets a machine-readable row carrying the stable
// go-v1 boundary code, and `status --check` fails.
func TestStatusReportsAnUnusableToolchainPerCompiledCommand(t *testing.T) {
	compiledProject(t)
	if code, stdout, stderr := capture(t, "install", "app"); code != exitOK {
		t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	t.Setenv(godriver.SelectionCuratorGo, filepath.Join(t.TempDir(), "nowhere", "bin", "go"))

	// Reporting a refusal is not itself a failure: the row carries the verdict
	// and `--check` is what turns a non-current verdict into a non-zero exit.
	code, stdout, stderr := capture(t, "status", "app", "--json")
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
	// Nothing the plan could not derive may be published as if it were known.
	if row.CacheKey != "" || row.Target != "" || row.ArtifactPath != "" {
		t.Fatalf("toolchain row published identities the plan never derived: %+v", row)
	}
	if report.Skills["build-skill"] != buildUnusableToolchain {
		t.Fatalf("skills = %v", report.Skills)
	}

	code, human, _ := capture(t, "status", "app")
	if code != exitOK || !strings.Contains(human, "state="+buildUnusableToolchain) ||
		!strings.Contains(human, "cause="+row.Cause) {
		t.Fatalf("human status = %d and does not report the toolchain row:\n%s", code, human)
	}
	if code, _, _ := capture(t, "status", "app", "--check"); code != exitFail {
		t.Fatalf("status --check with an unusable toolchain = %d, want %d", code, exitFail)
	}
}

// TestStatusKeepsLegacyBehaviourWhenAPlanFailsWithoutCompiledCommands pins the
// compatibility boundary of the change above: a failure that derived no
// compiled command still reports exactly the historical error and nothing else.
func TestStatusKeepsLegacyBehaviourWhenAPlanFailsWithoutCompiledCommands(t *testing.T) {
	project := legacyProject(t)
	if code, _, stderr := capture(t, "install", "app"); code != exitOK {
		t.Fatalf("install = %d\n%s", code, stderr)
	}
	// A declaration that cannot be resolved fails the read-only plan long
	// before any build phase.
	writeFile(t, filepath.Join(project, manifest.Name),
		`{"schema_version":1,"agents":["codex_cli"],"skills":[{"name":"skill-a","tag":"v9"}]}`)

	code, stdout, stderr := capture(t, "status", "app", "--json")
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

// TestInstallRepairsUntrustedCompiledStateAndPreservesTheOldInstall proves the
// documented reconciliation path: install rebuilds candidate bytes that are
// outside a provable protected boundary instead of adopting them.
func TestInstallRepairsUntrustedCompiledStateAndPreservesTheOldInstall(t *testing.T) {
	project, home := compiledProject(t)
	if code, stdout, stderr := capture(t, "install", "app"); code != exitOK {
		t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	installed := filepath.Join(project, ".agents", "skills", "build-skill")
	before := marker.Read(installed)
	if before == nil || len(before.Builds) != 1 {
		t.Fatalf("first install did not record compiled state: %+v", before)
	}
	for _, entry := range cacheEntries(t, home) {
		if err := os.Chmod(entry, 0o777); err != nil {
			t.Fatal(err)
		}
	}

	code, dryRun, _ := capture(t, "install", "app", "--dry-run")
	if code != exitOK {
		t.Fatalf("dry run over untrusted cache state = %d\n%s", code, dryRun)
	}
	if !strings.Contains(dryRun, "outcome=would-rebuild-untrusted-cache") {
		t.Fatalf("dry run did not name the untrusted protected state:\n%s", dryRun)
	}
	if strings.Contains(dryRun, "rebuilt untrusted") {
		t.Fatalf("dry run claimed a completed repair:\n%s", dryRun)
	}

	code, stdout, stderr := capture(t, "install", "app")
	if code != exitOK {
		t.Fatalf("repairing install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "rebuilt untrusted build cache state into a new protected entry") {
		t.Fatalf("install did not report the repair:\n%s", stdout)
	}
	if code, _, _ := capture(t, "status", "app", "--check"); code != exitOK {
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

	// A run that cannot pass the toolchain gate repairs nothing and leaves the
	// durable installation exactly as the successful run left it.
	t.Setenv(godriver.SelectionCuratorGo, filepath.Join(t.TempDir(), "nowhere", "bin", "go"))
	if code, stdout, _ := capture(t, "install", "app"); code != exitFail {
		t.Fatalf("install without a usable toolchain = %d, want %d\n%s", code, exitFail, stdout)
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
	_, home := compiledProject(t)
	if code, stdout, stderr := capture(t, "install", "app"); code != exitOK {
		t.Fatalf("install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	entries := cacheEntries(t, home)
	if len(entries) != 1 {
		t.Fatalf("protected build cache holds %d entries, want 1", len(entries))
	}

	code, stdout, stderr := capture(t, "gc")
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
	if code, _, _ := capture(t, "status", "app", "--check"); code != exitOK {
		t.Fatalf("status --check after gc = %d, want %d", code, exitOK)
	}
}

// TestDryRunNeverClaimsACompletedCompilerCheck pins the compiler-free dry-run
// vocabulary: a miss is planned, never reported as preflighted or built.
func TestDryRunNeverClaimsACompletedCompilerCheck(t *testing.T) {
	compiledProject(t)
	code, stdout, stderr := capture(t, "install", "app", "--dry-run")
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
	code, upgraded, stderr := capture(t, "upgrade", "app", "--dry-run")
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
	compiledProject(t)
	t.Setenv(godriver.SelectionCuratorGo, filepath.Join(t.TempDir(), "nowhere", "bin", "go"))

	code, _, stderr := capture(t, "status", "app")
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
	if code, _, _ := capture(t, "status", "app", "--check"); code != exitFail {
		t.Fatalf("status --check with an unusable toolchain = %d, want %d", code, exitFail)
	}
	if code, _, _ := capture(t, "install", "app"); code != exitFail {
		t.Fatalf("install with an unusable toolchain = %d, want %d", code, exitFail)
	}
}

// TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands pins backward
// compatibility: a closure with no build command produces exactly the
// historical document, with no build key at all.
func TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands(t *testing.T) {
	legacyProject(t)
	if code, _, stderr := capture(t, "install", "app"); code != exitOK {
		t.Fatalf("install = %d\n%s", code, stderr)
	}

	code, stdout, _ := capture(t, "status", "app", "--json")
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
	project := legacyProject(t)
	if code, _, stderr := capture(t, "install", "app"); code != exitOK {
		t.Fatalf("install = %d\n%s", code, stderr)
	}

	installed := filepath.Join(project, ".agents", "skills", "skill-a")
	rewriteMarker(t, installed, func(object map[string]any) {
		object["schema_version"] = marker.LegacySchemaVersion
		delete(object, "build_roots")
		delete(object, "build_source")
		delete(object, "builds")
	})

	code, stdout, stderr := capture(t, "status", "app", "--check")
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
