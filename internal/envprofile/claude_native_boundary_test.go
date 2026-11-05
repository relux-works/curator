package envprofile

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

// The parent overlay plants after the REAL os.Rename target Lstat (Unix), or
// source CreateFile (Windows). The candidate overlay plants after its native
// call preparation, before returning the prepared call to publishClaudeSeed.
// Both execute Resolve and require the plant marker plus exact merged content.
// This measures preparation ordering, not writes inside the native rename call.
func TestClaudeSeedNativeBoundary(t *testing.T) {
	if os.Getenv("CURATOR_CLAUDE_BOUNDARY_CHILD") == "1" {
		claudeSeedNativeBoundaryResolve(t)
		return
	}
	for _, lane := range []string{"candidate", "parent", "direct-rename-reversion", "late-preparation"} {
		t.Run(lane, func(t *testing.T) {
			out, code := claudeSeedNativeBoundaryOverlay(t, lane)
			want := 0
			if lane != "candidate" {
				want = 1 // expected red, never count a build/setup error as loss.
			}
			if code != want {
				t.Errorf("boundary lane %s exit=%d want=%d", lane, code, want)
			}
			for _, form := range []string{"referenced", "monolithic"} {
				for _, variant := range []string{"changed", "normal"} {
					verdict := "PASS"
					if want == 1 && variant == "changed" {
						verdict = "FAIL"
					}
					name := "TestClaudeSeedNativeBoundary/" + form + "/" + variant
					if !strings.Contains(out, "--- "+verdict+": "+name+" ") ||
						!strings.Contains(out, "boundary exercised: "+form+"/"+variant) {
						t.Errorf("missing exercised boundary and %s for %s", verdict, name)
					}
				}
			}
			if want == 1 && !strings.Contains(out, "native-boundary tool update lost") {
				t.Error("expected-red lane lacks intended loss assertion")
			}
		})
	}
}

func claudeSeedNativeBoundaryResolve(t *testing.T) {
	t.Helper()
	for _, form := range []string{envregistry.FormReferenced, envregistry.FormMonolithic} {
		for _, variant := range []string{"changed", "normal"} {
			t.Run(form+"/"+variant, func(t *testing.T) {
				_, req, seed := claudeRepairFixture(t, form)
				if err := os.WriteFile(seed, []byte(`{"counter":1}`), 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(filepath.Dir(seed), "CLAUDE.md"), []byte("drift"), 0600); err != nil {
					t.Fatal(err)
				}
				marker := filepath.Join(t.TempDir(), "plant")
				t.Setenv("CURATOR_CLAUDE_BOUNDARY_TARGET", seed)
				t.Setenv("CURATOR_CLAUDE_BOUNDARY_MARKER", marker)
				payload := ""
				counter := json.Number("1")
				if variant == "changed" {
					payload = `{"counter":2,"tool_update":"must survive","number":1e999}`
					counter = json.Number("2")
				}
				t.Setenv("CURATOR_CLAUDE_BOUNDARY_PAYLOAD", payload)
				_, err := Resolve(req)
				if err != nil {
					t.Fatalf("Resolve native-boundary repair: %v", err)
				}
				if got := string(readClaudeState(t, marker)); got != "native preparation reached" {
					t.Fatalf("plant not exercised: %q", got)
				}
				if os.Getenv("CURATOR_CLAUDE_BOUNDARY_TARGET") != "" {
					t.Fatal("plant did not disarm")
				}
				t.Logf("boundary exercised: %s/%s", form, variant)
				entry := map[string]any{"hasTrustDialogAccepted": true}
				if form == envregistry.FormReferenced {
					entry["hasClaudeMdExternalIncludesApproved"] = true
				}
				want := map[string]any{"counter": counter, "projects": map[string]any{req.LaunchDir: entry}}
				if variant == "changed" {
					want["tool_update"] = "must survive"
					want["number"] = json.Number("1e999")
				}
				if got := claudeNumberObject(t, readClaudeState(t, seed)); !reflect.DeepEqual(got, want) {
					t.Errorf("native-boundary tool update lost: got %#v want %#v", got, want)
				}
				req.Repair = false
				if _, err := Resolve(req); err != nil {
					t.Errorf("verify repaired seed: %v", err)
				}
			})
		}
	}
}

// Test-only source injection; no production hook, polling, sleeps or fake I/O.
// Each caller supplies the return shape of the production function it overlays.
func claudeNativeBoundaryPlant(errorReturn string) string {
	return strings.ReplaceAll(`if target == os.Getenv("CURATOR_CLAUDE_BOUNDARY_TARGET") {
		if probeErr := os.Unsetenv("CURATOR_CLAUDE_BOUNDARY_TARGET"); probeErr != nil { RETURN_ERROR }
		if payload := os.Getenv("CURATOR_CLAUDE_BOUNDARY_PAYLOAD"); payload != "" {
			if probeErr := os.WriteFile(target, []byte(payload), 0600); probeErr != nil { RETURN_ERROR }
		}
		if probeErr := os.WriteFile(os.Getenv("CURATOR_CLAUDE_BOUNDARY_MARKER"), []byte("native preparation reached"), 0600); probeErr != nil { RETURN_ERROR }
	}`, "RETURN_ERROR", errorReturn)
}

func claudeSeedNativeBoundaryOverlay(t *testing.T, lane string) (string, int) {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	replacements := map[string]string{}
	put := func(path string, data []byte) {
		t.Helper()
		replacement := filepath.Join(dir, filepath.Base(path))
		if err := os.WriteFile(replacement, data, 0600); err != nil {
			t.Fatal(err)
		}
		replacements[path] = replacement
	}
	plant := func(path, before, after string) {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		source := strings.ReplaceAll(string(data), "\r\n", "\n")
		if strings.Count(source, before) != 1 {
			t.Fatalf("boundary plant anchor is not unique: %s", filepath.Base(path))
		}
		put(path, []byte(strings.Replace(source, before, after, 1)))
		t.Logf("applied %s plant: %s", lane, filepath.Base(path))
	}
	module := filepath.Join(root, "internal", "envprofile")
	if lane == "candidate" || lane == "late-preparation" {
		if lane == "late-preparation" {
			// Keep the comparison AND direct native rename. Delay preparation
			// until after the comparison, admitting just this late-write class.
			plant(filepath.Join(module, "managed.go"), "rename, err = prepareClaudeSeedRename(temp, full)",
				`rename = func() error {
					native, prepareErr := prepareClaudeSeedRename(temp, full)
					if prepareErr != nil { return prepareErr }
					return native()
				}`)
		}
		file := "claude_seed_rename_unix.go"
		if runtime.GOOS == "windows" {
			file = "claude_seed_rename_windows.go"
		}
		plant(filepath.Join(module, file), "return func() error {",
			claudeNativeBoundaryPlant("return nil, probeErr")+"; return func() error {")
	} else {
		if lane == "parent" {
			// Exact historical function, extracted from 3fd0ab21. Keep the
			// witness executable in hosted shallow checkouts without fetching
			// repository history or changing the workflow.
			historical, err := os.ReadFile(filepath.Join(module, "testdata", "claude-seed-publish-3fd0ab21.go.txt"))
			if err != nil {
				t.Fatal(err)
			}
			historical = bytes.ReplaceAll(historical, []byte("\r\n"), []byte("\n"))
			if fmt.Sprintf("%x", sha256.Sum256(historical)) != "ee6597b496d4f7fcde1e41f4ac4b3492d69cdb0dca1c1cdbdaf33a9b105ebbc5" {
				t.Fatal("historical function identity changed")
			}
			path := filepath.Join(module, "managed.go")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			source := strings.ReplaceAll(string(data), "\r\n", "\n")
			start := strings.Index(source, "func publishClaudeSeed(")
			end := strings.Index(source, "func readClaudeSeedInput(")
			if start < 0 || end <= start {
				t.Fatal("historical publication boundaries not found")
			}
			put(path, []byte(source[:start]+string(historical)+source[end:]))
			t.Log("applied exact parent publication function: 3fd0ab21 / ee6597b4")
		} else {
			plant(filepath.Join(module, "managed.go"), "err = rename()", "err = replaceManagedEntry(temp, full); _ = rename")
		}
		if runtime.GOOS == "windows" {
			plant(filepath.Join(module, "nofollow_open_windows.go"), "defer windows.CloseHandle(handle)",
				"defer windows.CloseHandle(handle); "+claudeNativeBoundaryPlant("return probeErr"))
		} else {
			// Instrument the pinned Go implementation AFTER its actual target
			// Lstat. The product still calls the real os.Rename -> native rename.
			cmd := exec.Command("go", "env", "GOROOT")
			goRoot, err := cmd.Output()
			if err != nil {
				t.Fatalf("find active toolchain root: %v", err)
			}
			plant(filepath.Join(strings.TrimSpace(string(goRoot)), "src", "os", "file_unix.go"), "fi, err := Lstat(newname)", `fi, err := Lstat(newname)
			if newname == Getenv("CURATOR_CLAUDE_BOUNDARY_TARGET") {
				if e := Unsetenv("CURATOR_CLAUDE_BOUNDARY_TARGET"); e != nil { return e }
				if p := Getenv("CURATOR_CLAUDE_BOUNDARY_PAYLOAD"); p != "" {
					if e := WriteFile(newname, []byte(p), 0600); e != nil { return e }
				}
				if e := WriteFile(Getenv("CURATOR_CLAUDE_BOUNDARY_MARKER"), []byte("native preparation reached"), 0600); e != nil { return e }
			}`)
		}
	}
	data, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(overlay, data, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "test", "-overlay", overlay, "-count=1", "-timeout=4m", "-v", "-run", "^TestClaudeSeedNativeBoundary$", "./internal/envprofile")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "CURATOR_CLAUDE_BOUNDARY_CHILD=1")
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	if err := cmd.Start(); err != nil {
		t.Fatalf("cannot-start boundary lane: %v", err)
	}
	t.Logf("command=%q PID=%d", cmd.Args, cmd.Process.Pid)
	err = cmd.Wait()
	code := 0
	if err != nil {
		exit, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}
		code = exit.ExitCode()
	}
	t.Logf("lane=%s exit=%d\n%s", lane, code, out.String())
	return out.String(), code
}
