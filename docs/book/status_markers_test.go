package book_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func statusFilter() string {
	if path := os.Getenv("BOOK_STATUS_FILTER"); path != "" {
		return path
	}
	return "filters/book.lua"
}

// Real pandoc validates every marker in a run, including adjacent multiword
// labels. Valid runs become individual badges; unknown later labels refuse
// with the Lua-filter diagnostic, rather than passing through as plain text.
func TestAdjacentStatusMarkers(t *testing.T) {
	pandoc, err := exec.LookPath("pandoc")
	if err != nil {
		t.Skip("pandoc unavailable; PDF toolchain check required separately")
	}
	for _, tc := range []struct {
		name, input, unknown string
		badges               int
	}{
		{"single", "⟦решено⟧", "", 1},
		{"adjacent", "⟦решено⟧⟦открыто⟧", "", 2},
		{"multiword", "⟦решено⟧⟦работает сегодня⟧", "", 2},
		{"punctuation", "До(⟦решено⟧),⟦открыто⟧⟦спецификация⟧.", "", 3},
		{"unknown_second", "⟦решено⟧⟦готово⟧", "готово", 0},
		{"unknown_third", "⟦решено⟧⟦открыто⟧⟦готово⟧", "готово", 0},
		{"unknown_multiword", "⟦решено⟧⟦ещё не принято⟧", "ещё не принято", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(pandoc, "-f", "markdown", "-t", "html5", "--lua-filter", statusFilter())
			cmd.Stdin = strings.NewReader(tc.input + "\n")
			out, err := cmd.CombinedOutput()
			if tc.unknown != "" {
				exit, ok := err.(*exec.ExitError)
				if !ok || exit.ExitCode() != 83 || !strings.Contains(string(out), "unknown status marker: "+tc.unknown) {
					t.Fatalf("unknown adjacent status not rejected: %s (%v)", out, err)
				}
				return
			}
			if err != nil || strings.Count(string(out), `class="status `) != tc.badges || strings.Contains(string(out), "⟦") {
				t.Fatalf("valid marker run did not become %d badges: %s (%v)", tc.badges, out, err)
			}
			if tc.name == "punctuation" && (!strings.Contains(string(out), "До(") || !strings.Contains(string(out), "</span>),") || !strings.Contains(string(out), "</span>.")) {
				t.Fatalf("marker punctuation changed: %s", out)
			}
		})
	}
}

// The real build script and pandoc accept the valid adjacent pair, but refuse
// an unknown second marker before invoking the PDF renderer or replacing prior
// PDF bytes. Only WeasyPrint is instrumented; this is not a PDF-layout test.
func TestBuildAdjacentStatusMarkerGate(t *testing.T) {
	requireFigureTools(t)
	if _, err := exec.LookPath("pandoc"); err != nil {
		t.Skip("pandoc unavailable; PDF toolchain check required separately")
	}
	for _, tc := range []struct {
		name, markers string
		unknown       bool
	}{
		{"valid", "⟦решено⟧⟦открыто⟧", false},
		{"unknown_second", "⟦решено⟧⟦готово⟧", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := figureFixture(t)
			for _, name := range []string{"build.sh", "template.html", "style.css"} {
				write(t, filepath.Join(root, name), read(t, name))
			}
			write(t, filepath.Join(root, "filters", "book.lua"), read(t, statusFilter()))
			write(t, filepath.Join(root, "ru", "00-chapter.md"), tc.markers+"\n\n![Figure](../diagrams/figure.png)\n")
			if err := os.Chmod(filepath.Join(root, "check-figures.sh"), 0o700); err != nil {
				t.Fatal(err)
			}
			renderer := filepath.Join(root, "bin", "weasyprint")
			write(t, renderer, `#!/usr/bin/env bash
printf called > "$(dirname "$0")/renderer-called"
for last; do :; done
printf rendered > "$last"
`)
			if err := os.Chmod(renderer, 0o700); err != nil {
				t.Fatal(err)
			}
			pdf := filepath.Join(root, "output.pdf")
			write(t, pdf, "prior PDF")
			cmd := exec.Command("bash", filepath.Join(root, "build.sh"), "ru", pdf)
			cmd.Env = append(os.Environ(), "PATH="+filepath.Dir(renderer)+string(os.PathListSeparator)+os.Getenv("PATH"))
			out, err := cmd.CombinedOutput()
			called := filepath.Join(root, "bin", "renderer-called")
			if !tc.unknown {
				if err != nil || read(t, pdf) != "rendered" || read(t, called) != "called" {
					t.Fatalf("valid adjacent-marker build refused: %s (%v)", out, err)
				}
				return
			}
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 83 || !strings.Contains(string(out), "unknown status marker: готово") {
				t.Fatalf("status build expected unknown-marker rejection: %s (%v)", out, err)
			}
			if read(t, pdf) != "prior PDF" {
				t.Fatal("unknown marker replaced the prior PDF")
			}
			if _, err := os.Stat(called); !os.IsNotExist(err) {
				t.Fatalf("unknown marker invoked the renderer or invocation record is unreadable: %v", err)
			}
		})
	}
}

// Removing only the adjacent-marker split must let the planted second unknown
// marker through the build, while the valid build remains accepted. A setup
// failure or unrelated child failure does not count as detecting the mutant.
func TestAdjacentStatusMarkerNarrowingControl(t *testing.T) {
	requireFigureTools(t)
	if _, err := exec.LookPath("pandoc"); err != nil {
		t.Skip("pandoc unavailable")
	}
	original := read(t, "filters/book.lua")
	needle := "  inlines = split_marker_starts(inlines)"
	if strings.Count(original, needle) != 1 {
		t.Fatal("adjacent-marker control plant no longer applies exactly once")
	}
	mutant := filepath.Join(t.TempDir(), "filter.lua")
	write(t, mutant, strings.Replace(original, needle, "", 1))
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, witness := range []string{"valid", "unknown_second"} {
		cmd := exec.Command(binary, "-test.run=^TestBuildAdjacentStatusMarkerGate$/^"+witness+"$", "-test.v")
		cmd.Env = append(os.Environ(), "BOOK_STATUS_FILTER="+mutant)
		out, err := cmd.CombinedOutput()
		if witness == "valid" {
			if err != nil || !strings.Contains(string(out), "--- PASS: TestBuildAdjacentStatusMarkerGate/valid") {
				t.Fatalf("adjacent-marker mutant broke positive setup: %s (%v)", out, err)
			}
			continue
		}
		exit, ok := err.(*exec.ExitError)
		if !ok || exit.ExitCode() != 1 || !strings.Contains(string(out), "--- FAIL: TestBuildAdjacentStatusMarkerGate/unknown_second") || !strings.Contains(string(out), "status build expected unknown-marker rejection:") {
			t.Fatalf("adjacent-marker mutant did not fail its intended build assertion: %s (%v)", out, err)
		}
	}
}
