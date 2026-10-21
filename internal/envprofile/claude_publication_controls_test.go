package envprofile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envregistry"
)

// Source overlays schedule real tool writes without adding a production hook.
// They drive Resolve, not a second implementation of the publication algorithm.
var claudePublicationProbe func(int) error

func claudePublicationOverlay(t *testing.T, file, before, after, run string, child bool) (string, int) {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	original := filepath.Join(root, "internal", "envprofile", file)
	data, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	source := strings.ReplaceAll(string(data), "\r\n", "\n")
	if strings.Count(source, before) != 1 {
		t.Fatal("publication control anchor is not unique")
	}
	dir := t.TempDir()
	replacement := filepath.Join(dir, file)
	if err := os.WriteFile(replacement, []byte(strings.Replace(source, before, after, 1)), 0600); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(map[string]any{"Replace": map[string]string{original: replacement}})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(overlay, payload, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "test", "-overlay", overlay, "-count=1", "-timeout=4m", "-v", "-run", run, "./internal/envprofile")
	cmd.Dir = root
	cmd.Env = os.Environ()
	if child {
		cmd.Env = append(cmd.Env, "CURATOR_PUBLICATION_PROBE=1")
	}
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Start(); err != nil {
		t.Fatalf("cannot-start: %v", err)
	}
	t.Logf("owned child PID=%d", cmd.Process.Pid)
	err = cmd.Wait()
	code := 0
	if err != nil {
		exit, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}
		code = exit.ExitCode()
	}
	t.Logf("child exit %d:\n%s", code, output.String())
	return output.String(), code
}

// Parameterized class control: zero, one, two and three changed checks, both
// forms. Third mismatch refuses without publishing the seed; earlier context
// writes are not rolled back. This does not exercise the residual read/rename race.
func TestClaudePublicationRetryBounds(t *testing.T) {
	if os.Getenv("CURATOR_PUBLICATION_PROBE") != "1" {
		out, code := claudePublicationOverlay(t, "managed.go",
			"fresh, err := readClaudeSeedInput(root, homeRel)",
			"if claudePublicationProbe != nil { if err := claudePublicationProbe(attempt); err != nil { return nil, err } }; fresh, err := readClaudeSeedInput(root, homeRel)",
			"^TestClaudePublicationRetryBounds$", true)
		if code != 0 || !strings.Contains(out, "--- PASS: TestClaudePublicationRetryBounds/monolithic/changes-3") {
			t.Fatalf("retry property failed: exit %d", code)
		}
		return
	}
	for _, form := range []string{envregistry.FormReferenced, envregistry.FormMonolithic} {
		for changes := 0; changes <= 3; changes++ {
			t.Run(fmt.Sprintf("%s/changes-%d", form, changes), func(t *testing.T) {
				_, req, seed := claudeRepairFixture(t, form)
				if err := os.WriteFile(seed, []byte(`{"counter":0}`), 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(filepath.Dir(seed), "CLAUDE.md"), []byte("drift"), 0600); err != nil {
					t.Fatal(err)
				}
				calls := 0
				var last []byte
				claudePublicationProbe = func(attempt int) error {
					calls++
					if attempt < changes {
						last = []byte(fmt.Sprintf(`{"counter":%d,"number":1e999}`, attempt+1))
						return os.WriteFile(seed, last, 0600)
					}
					return nil
				}
				defer func() { claudePublicationProbe = nil }()
				_, err := Resolve(req)
				wantCalls := changes + 1
				if wantCalls > 3 {
					wantCalls = 3
				}
				if calls != wantCalls {
					t.Errorf("publication checks=%d want %d", calls, wantCalls)
				}
				if changes == 3 {
					if err == nil || !strings.Contains(err.Error(), envregistry.DiagSeedUnreadable) || !strings.Contains(err.Error(), "all 3 publication checks") {
						t.Fatalf("wrong exhaustion refusal: %v", err)
					}
					if !bytes.Equal(readClaudeState(t, seed), last) {
						t.Error("exhaustion published seed")
					}
				} else {
					if err != nil {
						t.Fatal(err)
					}
					got := claudeNumberObject(t, readClaudeState(t, seed))
					if got["counter"] != json.Number(fmt.Sprint(changes)) {
						t.Errorf("retry lost current input: %v", got)
					}
				}
				entries, err := os.ReadDir(filepath.Dir(seed))
				if err != nil {
					t.Fatal(err)
				}
				for _, entry := range entries {
					if strings.HasPrefix(entry.Name(), ".curator-write-") {
						t.Error("publication left staging output")
					}
				}
			})
		}
	}
}

// Narrowing controls retain early validation, real recovery and the diagnostic.
// Each weakens only one admitted class and requires named red plus green neighbors.
func TestClaudePublicationNarrowingMutants(t *testing.T) {
	for _, m := range []struct{ name, file, before, after, run, fail, pass, message string }{
		{"late-stale-payload", "managed.go", "if !sameClaudeSeedInput(prepared, fresh) {", "if !sameClaudeSeedInput(prepared, fresh) && string(prepared.input) != `{" + `"counter":1` + "}` {", "^TestReviewFinalPublicationUpdate$/^(referenced|monolithic)$/^(changed|unchanged)$", "TestReviewFinalPublicationUpdate/%s/changed", "TestReviewFinalPublicationUpdate/%s/unchanged", "late fresh tool update lost"},
		{"omit-completed-recovery", "lock.go", "if op.recovered {", "if op.recovered && !strings.Contains(err.Error(), envregistry.DiagSeedUnreadable) {", "^TestReviewPendingJournal$/^(referenced|monolithic)$/^(early-invalid|late-invalid|unchanged)$", "TestReviewPendingJournal/%s/late-invalid", "TestReviewPendingJournal/%s/unchanged", "completed recovery is absent from public refusal"},
	} {
		t.Run(m.name, func(t *testing.T) {
			// The report plant uses the existing error text without importing test-only
			// dependencies into production; preserve the recovered condition token.
			if m.name == "omit-completed-recovery" {
				m.after = "if op.recovered && err.Error() != `environment_seed_unreadable: managed .claude.json is not an object` {"
			}
			out, code := claudePublicationOverlay(t, m.file, m.before, m.after, m.run, false)
			if code != 1 || !strings.Contains(out, m.message) {
				t.Fatalf("mutant survived or unrelated failure: exit %d", code)
			}
			for _, form := range []string{"referenced", "monolithic"} {
				if !strings.Contains(out, "--- FAIL: "+fmt.Sprintf(m.fail, form)+" ") || !strings.Contains(out, "--- PASS: "+fmt.Sprintf(m.pass, form)+" ") {
					t.Fatalf("missing intended failure/valid neighbor for %s", form)
				}
			}
		})
	}
}
