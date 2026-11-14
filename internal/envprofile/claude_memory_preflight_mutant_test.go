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

	"github.com/relux-works/curator/internal/testtoolchain"
)

// Removing only the early settings preflight must expose real pending recovery
// writes before malformed-input refusal. Both valid-settings recovery controls
// still pass; setup failures and unrelated child failures are not a kill.
func TestClaudeNativeMemoryPreflightNarrowingMutant(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, "internal", "envprofile", "managed.go")
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	mutated, err := claudeMemoryPreflightMutantSource(data)
	if err != nil {
		t.Fatal(err)
	}
	run := func(overlay string) ([]byte, error) {
		args := []string{"test", "-count=1", "-timeout=2m", "-v", "-run", "^TestReviewPendingJournal$/^(referenced|monolithic)$/^(settings-invalid|unchanged)$", "./internal/envprofile"}
		if overlay != "" {
			args = append([]string{"test", "-overlay", overlay}, args[1:]...)
		}
		cmd := exec.Command(testtoolchain.Go(t), args...)
		cmd.Dir = root
		return cmd.CombinedOutput()
	}
	if output, err := run(""); err != nil {
		t.Fatalf("unmodified preflight controls: %v\n%s", err, output)
	}
	dir := t.TempDir()
	replacement := filepath.Join(dir, "managed.go")
	if err := os.WriteFile(replacement, mutated, 0600); err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{source: replacement}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(path, overlay, 0600); err != nil {
		t.Fatal(err)
	}
	output, err := run(path)
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "early refusal ran recovery") {
		t.Fatalf("early preflight plant survived or failed elsewhere: %v\n%s", err, output)
	}
	for _, form := range []string{"referenced", "monolithic"} {
		if !strings.Contains(string(output), "--- FAIL: TestReviewPendingJournal/"+form+"/settings-invalid ") || !strings.Contains(string(output), "--- PASS: TestReviewPendingJournal/"+form+"/unchanged ") {
			t.Fatalf("missing intended refusal/control for %s:\n%s", form, output)
		}
	}
	t.Logf("expected-red exit 1; early settings refusal ran recovery:\n%s", output)
}

// Include the repair-lock boundary so the same guard in another writer is not
// selected. Normalize Git checkout line endings before counting or replacing.
func claudeMemoryPreflightMutantSource(data []byte) ([]byte, error) {
	source := strings.ReplaceAll(string(data), "\r\n", "\n")
	const guard = "\t\tif _, err := prepareClaudeSettings(root, homeRel); err != nil {\n\t\t\treturn err\n\t\t}"
	const boundary = "\t\t// This callback holds the repair lock and precedes journal recovery,\n" +
		"\t\t// store rebuilding and legacy identity migration. Later refreshes do\n" +
		"\t\t// not replace this early refusal gate."
	anchor := guard + "\n" + boundary
	if count := strings.Count(source, anchor); count != 1 {
		return nil, fmt.Errorf("early preflight mutation anchor is not unique: count=%d", count)
	}
	return []byte(strings.Replace(source, anchor, "\t\t// early settings preflight deliberately omitted\n"+boundary, 1)), nil
}

// Actual-source LF/CRLF controls remove only the early guard and retain its
// boundary comment, including when an unrelated writer has the same guard.
// Missing/duplicate anchors refuse with no output or mutation of the input.
func TestClaudeNativeMemoryPreflightMutationAnchor(t *testing.T) {
	data, err := os.ReadFile("managed.go")
	if err != nil {
		t.Fatal(err)
	}
	lf := strings.ReplaceAll(string(data), "\r\n", "\n")
	// This independent expected edit deliberately omits the boundary from its
	// search: the fixture must distinguish the gate from an unrelated copy.
	guard := "\t\tif _, err := prepareClaudeSettings(root, homeRel); err != nil {\n\t\t\treturn err\n\t\t}"
	const marker = "// This callback holds the repair lock and precedes journal recovery,"
	if strings.Count(lf, guard) != 1 || strings.Count(lf, marker) != 1 {
		t.Fatal("source fixture does not expose exactly one early guard and boundary")
	}
	want := strings.Replace(lf, guard, "\t\t// early settings preflight deliberately omitted", 1)
	other := "\nfunc unrelatedSettingsWriter() {\n" + guard + "\n}\n"
	for _, tc := range []struct{ name, source, want, diagnostic string }{
		{name: "LF", source: lf, want: want},
		{name: "CRLF", source: strings.ReplaceAll(lf, "\n", "\r\n"), want: want},
		{name: "unrelated-guard", source: other + lf, want: other + want},
		{name: "unrelated-guard-CRLF", source: strings.ReplaceAll(other+lf, "\n", "\r\n"), want: other + want},
		{name: "missing-guard", source: strings.Replace(lf, guard, "", 1), diagnostic: "count=0"},
		{name: "missing-boundary", source: strings.Replace(lf, marker, "// different boundary", 1), diagnostic: "count=0"},
		{name: "duplicate", source: lf + lf, diagnostic: "count=2"},
		{name: "duplicate-CRLF", source: strings.ReplaceAll(lf+lf, "\n", "\r\n"), diagnostic: "count=2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := []byte(tc.source)
			before := bytes.Clone(input)
			got, err := claudeMemoryPreflightMutantSource(input)
			if tc.diagnostic != "" {
				if err == nil || !strings.Contains(err.Error(), "early preflight mutation anchor is not unique: "+tc.diagnostic) || got != nil {
					t.Fatalf("anchor refusal: output=%d bytes, err=%v, want %s", len(got), err, tc.diagnostic)
				}
			} else if err != nil || string(got) != tc.want {
				t.Fatalf("anchor changed the wrong source span: err=%v", err)
			}
			if !bytes.Equal(input, before) {
				t.Fatal("anchor processing changed its input")
			}
		})
	}
}
