package book_test

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// The Bash book toolchain is qualified here on Unix; Windows product tests
// are independent of this optional documentation renderer.
func requireFigureTools(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("book shell toolchain requires Unix paths")
	}
	for _, tool := range []string{"bash", "grep", "sed", "sort", "basename", "awk", "shasum", "cut"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("book checker prerequisite unavailable: %s", tool)
		}
	}
}

func figureFixture(t *testing.T) string {
	t.Helper()
	requireFigureTools(t)
	root := t.TempDir()
	script := "check-figures.sh"
	if override := os.Getenv("BOOK_FIGURE_CHECKER"); override != "" {
		script = override
	}
	write(t, filepath.Join(root, "check-figures.sh"), read(t, script))
	source := "@startuml\nSource -> Target\n@enduml\n"
	write(t, filepath.Join(root, "src", "figure.puml"), source)
	write(t, filepath.Join(root, "src", "MANIFEST.md"), fmt.Sprintf("| figure | Book original | PlantUML | %x |\n", sha256.Sum256([]byte(source))))
	// The checker verifies inventory/provenance, not PNG decoding or visual quality.
	write(t, filepath.Join(root, "diagrams", "figure.png"), "render fixture")
	write(t, filepath.Join(root, "ru", "00-chapter.md"), "![Figure](../diagrams/figure.png)\n")
	return root
}

// These fixtures prove the checker accepts consistent inventories and rejects
// missing sources/renders/rows, stale source hashes and orphan renders with the
// intended diagnostic. They do not prove that a render matches its source bytes.
func TestFigureCheckContracts(t *testing.T) {
	cases := []struct {
		name string
		edit func(*testing.T, string)
		want string
	}{
		{"valid", func(*testing.T, string) {}, ""},
		{"missing_render", func(t *testing.T, r string) { mustRemove(t, filepath.Join(r, "diagrams", "figure.png")) }, "missing render:"},
		{"missing_source", func(t *testing.T, r string) { mustRemove(t, filepath.Join(r, "src", "figure.puml")) }, "missing source:"},
		{"missing_manifest", func(t *testing.T, r string) { write(t, filepath.Join(r, "src", "MANIFEST.md"), "# Empty\n") }, "missing manifest row:"},
		{"unreferenced_source", func(t *testing.T, r string) { write(t, filepath.Join(r, "src", "extra.puml"), "@startuml\n@enduml\n") }, "source without manifest row:"},
		{"changed_hash", func(t *testing.T, r string) { write(t, filepath.Join(r, "src", "figure.puml"), "changed\n") }, "source hash mismatch:"},
		{"orphan_render", func(t *testing.T, r string) { write(t, filepath.Join(r, "diagrams", "extra.png"), "orphan") }, "render without source:"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := figureFixture(t)
			tc.edit(t, r)
			out, err := exec.Command("bash", filepath.Join(r, "check-figures.sh")).CombinedOutput()
			if tc.want == "" {
				if err != nil || strings.TrimSpace(string(out)) != "figures: ok" {
					t.Fatalf("valid inventory refused: %s (%v)", out, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("contract expected rejection for %s", tc.name)
			}
			if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 || !strings.Contains(string(out), tc.want) {
				t.Fatalf("wrong rejection, want %q: %s (%v)", tc.want, out, err)
			}
		})
	}
}

func mustRemove(t *testing.T, path string) {
	t.Helper()
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
}

// Each bounded mutant removes only the named rejection property. The existing
// negative witness must fail with its own assertion, while its valid control
// still passes; unrelated process/setup failures do not count as a killed mutant.
func TestFigureCheckNarrowingControls(t *testing.T) {
	requireFigureTools(t)
	original := read(t, "check-figures.sh")
	cases := []struct{ name, witness, needle string }{
		{"missing_render", "missing_render", `  [ -f "$here/diagrams/$stem.png" ] || { echo "missing render: diagrams/$stem.png"; fail=1; }`},
		{"source_manifest", "unreferenced_source", `  [ -n "$r" ] || { echo "source without manifest row: $stem"; fail=1; continue; }`},
		{"source_hash", "changed_hash", `    [ "$got" = "$want" ] || { echo "source hash mismatch: $stem"; fail=1; }`},
		{"orphan_render", "orphan_render", `  [ -f "$here/src/$stem.puml" ] || { echo "render without source: diagrams/$stem.png"; fail=1; }`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if strings.Count(original, tc.needle) != 1 {
				t.Fatal("control plant no longer applies exactly once")
			}
			mutant := filepath.Join(t.TempDir(), "mutant.sh")
			// The row gate's removed continue must still skip hash validation:
			// otherwise a missing row causes an unrelated hash refusal.
			replacement := ":"
			if tc.name == "source_manifest" {
				replacement = `  [ -n "$r" ] || continue`
			}
			write(t, mutant, strings.Replace(original, tc.needle, replacement, 1))
			binary, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			for _, witness := range []string{"valid", tc.witness} {
				cmd := exec.Command(binary, "-test.run=^TestFigureCheckContracts$/^"+witness+"$", "-test.v")
				cmd.Env = append(os.Environ(), "BOOK_FIGURE_CHECKER="+mutant)
				out, err := cmd.CombinedOutput()
				if witness == "valid" {
					if err != nil {
						t.Fatalf("mutant broke positive setup: %s (%v)", out, err)
					}
					continue
				}
				if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 || !strings.Contains(string(out), "contract expected rejection for "+witness) {
					t.Fatalf("mutant not killed by intended assertion: %s (%v)", out, err)
				}
			}
		})
	}
}

// With a real pandoc, all five editorial labels become badges and an unknown
// label is rejected. This validates the filter, not typography or PDF layout.
func TestStatusMarkers(t *testing.T) {
	pandoc, err := exec.LookPath("pandoc")
	if err != nil {
		t.Skip("pandoc unavailable; PDF toolchain check required separately")
	}
	labels := []string{"работает сегодня", "решено", "спецификация", "предложено", "открыто"}
	for _, label := range append(labels, "готово") {
		t.Run(label, func(t *testing.T) {
			filter := "filters/book.lua"
			if override := os.Getenv("BOOK_STATUS_FILTER"); override != "" {
				filter = override
			}
			cmd := exec.Command(pandoc, "-f", "markdown", "-t", "html5", "--lua-filter", filter)
			cmd.Stdin = strings.NewReader("Текст ⟦" + label + "⟧.\n")
			out, err := cmd.CombinedOutput()
			if label == "готово" {
				if err == nil || !strings.Contains(string(out), "unknown status marker: готово") {
					t.Fatalf("unknown status not rejected: %s (%v)", out, err)
				}
				return
			}
			if err != nil || !strings.Contains(string(out), `class="status `) || strings.Contains(string(out), "⟦") {
				t.Fatalf("valid status did not become a badge: %s (%v)", out, err)
			}
		})
	}
}

// The status gate control proves the unknown-label witness detects a fallback
// that silently renders unknown labels, with an accepted-label positive control.
func TestStatusMarkerNarrowingControl(t *testing.T) {
	if _, err := exec.LookPath("pandoc"); err != nil {
		t.Skip("pandoc unavailable")
	}
	original := read(t, "filters/book.lua")
	needle := `error("unknown status marker: " .. label)`
	if strings.Count(original, needle) != 1 {
		t.Fatal("status control plant no longer applies exactly once")
	}
	mutant := filepath.Join(t.TempDir(), "filter.lua")
	write(t, mutant, strings.Replace(original, needle, `class = "open"`, 1))
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, label := range []string{"решено", "готово"} {
		cmd := exec.Command(binary, "-test.run=^TestStatusMarkers$/^"+label+"$", "-test.v")
		cmd.Env = append(os.Environ(), "BOOK_STATUS_FILTER="+mutant)
		out, err := cmd.CombinedOutput()
		if label == "решено" {
			if err != nil {
				t.Fatalf("mutant broke positive setup: %s (%v)", out, err)
			}
		} else if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 || !strings.Contains(string(out), "unknown status not rejected:") {
			t.Fatalf("status mutant not killed by intended assertion: %s (%v)", out, err)
		}
	}
}

// A real shell with instrumented renderer stand-ins proves that the build calls
// the figure gate before renderer invocation: a missing render preserves prior
// PDF bytes and starts neither renderer. Real PDF rendering is checked separately.
func TestBuildFigureGate(t *testing.T) {
	for _, missing := range []bool{false, true} {
		t.Run(fmt.Sprintf("missing_%t", missing), func(t *testing.T) {
			root := figureFixture(t)
			script := "build.sh"
			if override := os.Getenv("BOOK_BUILD_SCRIPT"); override != "" {
				script = override
			}
			write(t, filepath.Join(root, "build.sh"), read(t, script))
			if err := os.Chmod(filepath.Join(root, "check-figures.sh"), 0o700); err != nil {
				t.Fatal(err)
			}
			for name, body := range map[string]string{
				"pandoc": `#!/usr/bin/env bash
printf called > "$(dirname "$0")/pandoc-called"
while (( $# )); do
  if [ "$1" = --output ]; then printf html > "$2"; exit 0; fi
  shift
done
exit 1
`,
				"weasyprint": `#!/usr/bin/env bash
printf called > "$(dirname "$0")/weasyprint-called"
for last; do :; done
printf rendered > "$last"
`,
			} {
				p := filepath.Join(root, "bin", name)
				write(t, p, body)
				if err := os.Chmod(p, 0o700); err != nil {
					t.Fatal(err)
				}
			}
			pdf := filepath.Join(root, "output.pdf")
			write(t, pdf, "prior PDF")
			if missing {
				mustRemove(t, filepath.Join(root, "diagrams", "figure.png"))
			}
			cmd := exec.Command("bash", filepath.Join(root, "build.sh"), "ru", pdf)
			cmd.Env = append(os.Environ(), "PATH="+filepath.Join(root, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))
			out, err := cmd.CombinedOutput()
			if !missing {
				if err != nil || read(t, pdf) != "rendered" {
					t.Fatalf("valid build failed: %s (%v)", out, err)
				}
				return
			}
			if err == nil {
				t.Fatal("build contract expected rejection before rendering")
			}
			if !strings.Contains(string(out), "missing render:") || read(t, pdf) != "prior PDF" {
				t.Fatalf("wrong rejection or PDF changed: %s (%v)", out, err)
			}
			for _, name := range []string{"pandoc-called", "weasyprint-called"} {
				if _, err := os.Stat(filepath.Join(root, "bin", name)); !os.IsNotExist(err) {
					t.Fatalf("renderer invoked or unreadable invocation record: %s (%v)", name, err)
				}
			}
		})
	}
}

// Removing the build's figure-gate invocation must fail the missing-render
// preservation witness while leaving the valid instrumented build passing.
func TestBuildFigureGateNarrowingControl(t *testing.T) {
	requireFigureTools(t)
	original := read(t, "build.sh")
	needle := `"$here/check-figures.sh"`
	if strings.Count(original, needle) != 1 {
		t.Fatal("build control plant no longer applies exactly once")
	}
	mutant := filepath.Join(t.TempDir(), "build.sh")
	write(t, mutant, strings.Replace(original, needle, ":", 1))
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, witness := range []string{"missing_false", "missing_true"} {
		cmd := exec.Command(binary, "-test.run=^TestBuildFigureGate$/^"+witness+"$", "-test.v")
		cmd.Env = append(os.Environ(), "BOOK_BUILD_SCRIPT="+mutant)
		out, err := cmd.CombinedOutput()
		if witness == "missing_false" {
			if err != nil {
				t.Fatalf("mutant broke valid setup: %s (%v)", out, err)
			}
		} else if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 || !strings.Contains(string(out), "build contract expected rejection before rendering") {
			t.Fatalf("build mutant not killed by intended assertion: %s (%v)", out, err)
		}
	}
}
