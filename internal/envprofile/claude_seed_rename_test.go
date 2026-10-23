//go:build darwin || linux || windows

package envprofile

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

// The overlay instruments the real os.Lstat, including os.Rename's hidden
// destination lookup. It arms an external writer only after the last comparison.
// Native publication must make that hook unreachable. Restoring one stat must
// lose the injected write and fail that assertion, while the before-read write
// still survives. This proves removal of the metadata gap, not atomic CAS.
func TestClaudeSeedNativeRenameBoundary(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	managedPath := filepath.Join(root, "internal", "envprofile", "managed.go")
	managed, err := os.ReadFile(managedPath)
	if err != nil {
		t.Fatal(err)
	}
	goroot, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		t.Fatalf("locate toolchain for native-boundary instrumentation: %v", err)
	}
	statPath := filepath.Join(strings.TrimSpace(string(goroot)), "src", "os", "stat.go")
	stat, err := os.ReadFile(statPath)
	if err != nil {
		t.Fatal(err)
	}
	const statAnchor = "func Lstat(name string) (FileInfo, error) {"
	if bytes.Count(stat, []byte(statAnchor)) != 1 || bytes.Count(managed, []byte("err = rename()")) != 1 {
		t.Fatal("native-boundary instrumentation anchor is not unique")
	}
	// All hook state and payloads belong to disposable fixtures. Instrumenting
	// the standard-library boundary also catches a regression to os.Rename.
	stat = bytes.Replace(stat, []byte(statAnchor), []byte(statAnchor+`
	if name == Getenv("CURATOR_TEST_CLAUDE_GAP_PATH") && Getenv("CURATOR_TEST_CLAUDE_GAP_ARMED") == "1" {
		_ = Setenv("CURATOR_TEST_CLAUDE_GAP_ARMED", "0")
		if err := WriteFile(name, []byte(Getenv("CURATOR_TEST_CLAUDE_GAP_PAYLOAD")), 0600); err != nil {
			_ = Setenv("CURATOR_TEST_CLAUDE_GAP_ERROR", err.Error())
		}
		_ = Setenv("CURATOR_TEST_CLAUDE_GAP_HIT", "1")
	}
`), 1)
	for _, mutant := range []bool{false, true} {
		name := "native"
		publication := "err = rename()"
		if mutant {
			name = "stat-mutant"
			publication = "_, _ = os.Lstat(full)\n\t\t\terr = rename()"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			managedReplacement := filepath.Join(dir, "managed.go")
			statReplacement := filepath.Join(dir, "stat.go")
			armed := "_ = os.Setenv(\"CURATOR_TEST_CLAUDE_GAP_ARMED\", \"1\")\n\t\t\t" + publication + "\n\t\t\t_ = os.Setenv(\"CURATOR_TEST_CLAUDE_GAP_ARMED\", \"0\")"
			if err := os.WriteFile(managedReplacement, bytes.Replace(managed, []byte("err = rename()"), []byte(armed), 1), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(statReplacement, stat, 0600); err != nil {
				t.Fatal(err)
			}
			mapping, err := json.Marshal(map[string]any{"Replace": map[string]string{managedPath: managedReplacement, statPath: statReplacement}})
			if err != nil {
				t.Fatal(err)
			}
			overlay := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(overlay, mapping, 0600); err != nil {
				t.Fatal(err)
			}
			command := exec.Command("go", "test", "-overlay", overlay, "-count=1", "-timeout=3m", "-v", "-run", "^TestClaudeSeedNativeRenameBoundaryChild$", "./internal/envprofile")
			command.Dir = root
			command.Env = append(os.Environ(), "CURATOR_TEST_CLAUDE_GAP_CHILD=1")
			var buffer bytes.Buffer
			command.Stdout, command.Stderr = &buffer, &buffer
			if err := command.Start(); err != nil {
				t.Fatal(err)
			}
			t.Logf("owned child PID=%d", command.Process.Pid)
			err = command.Wait()
			output := buffer.String()
			if !mutant {
				if err != nil {
					t.Fatalf("native-boundary witness failed: %v\n%s", err, output)
				}
			} else {
				exit, ok := err.(*exec.ExitError)
				if !ok || exit.ExitCode() != 1 || !strings.Contains(output, "post-comparison stat overwrote the external seed update") {
					t.Fatalf("stat mutant did not fail its intended assertion: %v\n%s", err, output)
				}
				for _, form := range []string{envregistry.FormReferenced, envregistry.FormMonolithic} {
					if !strings.Contains(output, "--- FAIL: TestClaudeSeedNativeRenameBoundaryChild/"+form+"/native-gap ") ||
						!strings.Contains(output, "--- PASS: TestClaudeSeedNativeRenameBoundaryChild/"+form+"/before-comparison ") {
						t.Fatalf("stat mutant missed the gap failure or its valid neighbor:\n%s", output)
					}
				}
			}
			t.Log(output)
		})
	}
}

// This subprocess-only witness uses real temporary files and native renames.
// A before-comparison update must be merged, malformed input must be refused
// byte-exact with staging removed, and unchanged input must publish trust.
// The after-comparison hook must be unreachable on the unmodified native path.
func TestClaudeSeedNativeRenameBoundaryChild(t *testing.T) {
	if os.Getenv("CURATOR_TEST_CLAUDE_GAP_CHILD") != "1" {
		t.Skip("helper process: run through TestClaudeSeedNativeRenameBoundary's instrumented subprocess")
	}
	for _, form := range []string{envregistry.FormReferenced, envregistry.FormMonolithic} {
		for _, variant := range []string{"native-gap", "before-comparison", "invalid-before-comparison"} {
			t.Run(form+"/"+variant, func(t *testing.T) {
				root := t.TempDir()
				seed := filepath.Join(root, ".claude.json")
				launch := filepath.Join(root, "project")
				if err := os.WriteFile(seed, []byte(`{"counter":1,"number":0.125}`), 0600); err != nil {
					t.Fatal(err)
				}
				prepared, err := prepareClaudeSeed(root, ".", launch, form)
				if err != nil {
					t.Fatal(err)
				}
				payload := `{"counter":2,"number":1e999,"external":"must survive"}`
				if variant == "invalid-before-comparison" {
					payload = `{"broken":`
				}
				t.Setenv("CURATOR_TEST_CLAUDE_GAP_PATH", seed)
				t.Setenv("CURATOR_TEST_CLAUDE_GAP_PAYLOAD", payload)
				t.Setenv("CURATOR_TEST_CLAUDE_GAP_HIT", "")
				t.Setenv("CURATOR_TEST_CLAUDE_GAP_ERROR", "")
				armed := "0"
				if variant != "native-gap" {
					armed = "1"
				}
				t.Setenv("CURATOR_TEST_CLAUDE_GAP_ARMED", armed)
				_, err = publishClaudeSeed(prepared, root, ".", launch, form)
				if hookErr := os.Getenv("CURATOR_TEST_CLAUDE_GAP_ERROR"); hookErr != "" {
					t.Fatal(hookErr)
				}
				if variant == "invalid-before-comparison" {
					if err == nil || !strings.Contains(err.Error(), envregistry.DiagSeedUnreadable) || string(readClaudeState(t, seed)) != payload {
						t.Fatalf("invalid fresh seed was not refused byte-exact: %v", err)
					}
				} else {
					if err != nil {
						t.Fatal(err)
					}
					object := claudeNumberObject(t, readClaudeState(t, seed))
					hit := os.Getenv("CURATOR_TEST_CLAUDE_GAP_HIT") == "1"
					if hit && (object["counter"] != json.Number("2") || object["number"] != json.Number("1e999") || object["external"] != "must survive") {
						t.Fatal("post-comparison stat overwrote the external seed update")
					}
					if variant == "native-gap" {
						if hit {
							t.Fatal("post-comparison stat hook was reachable")
						}
						if object["counter"] != json.Number("1") || object["number"] != json.Number("0.125") {
							t.Fatal("unchanged control lost existing values")
						}
					}
					entry := object["projects"].(map[string]any)[launch].(map[string]any)
					if entry["hasTrustDialogAccepted"] != true || (form == envregistry.FormReferenced && entry["hasClaudeMdExternalIncludesApproved"] != true) {
						t.Fatal("native publication did not install the required trust fields")
					}
				}
				if variant != "native-gap" && os.Getenv("CURATOR_TEST_CLAUDE_GAP_HIT") != "1" {
					t.Fatal("before-comparison write was not applied")
				}
				if leftovers, err := filepath.Glob(filepath.Join(root, ".curator-write-*.tmp")); err != nil || len(leftovers) != 0 {
					t.Fatalf("publication leaked staged files: %v, %v", leftovers, err)
				}
			})
		}
	}
}

// The prepared native operation replaces an absent or regular destination;
// a directory destination and an invalid path cannot destroy either entry.
func TestClaudeSeedPreparedRename(t *testing.T) {
	for _, variant := range []string{"absent", "regular", "directory", "invalid-path"} {
		t.Run(variant, func(t *testing.T) {
			root := t.TempDir()
			source, target := filepath.Join(root, "staged"), filepath.Join(root, "seed")
			if err := os.WriteFile(source, []byte("new"), 0600); err != nil {
				t.Fatal(err)
			}
			switch variant {
			case "regular":
				if err := os.WriteFile(target, []byte("old"), 0600); err != nil {
					t.Fatal(err)
				}
			case "directory":
				if err := os.Mkdir(target, 0700); err != nil {
					t.Fatal(err)
				}
			case "invalid-path":
				target += "\x00"
			}
			rename, release, err := prepareClaudeSeedRename(source, target)
			if err == nil {
				err = rename()
				release()
			}
			if variant == "directory" || variant == "invalid-path" {
				if err == nil || string(readClaudeState(t, source)) != "new" {
					t.Fatalf("invalid destination was not refused with source intact: %v", err)
				}
				if variant == "directory" {
					info, statErr := os.Stat(target)
					if statErr != nil || !info.IsDir() {
						t.Fatalf("rename changed directory destination: %v", statErr)
					}
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if string(readClaudeState(t, target)) != "new" {
				t.Fatal("native rename did not publish source bytes")
			}
			if _, err := os.Stat(source); !os.IsNotExist(err) {
				t.Fatalf("native rename retained the source: %v", err)
			}
		})
	}
}
