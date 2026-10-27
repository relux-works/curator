package book_test

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

func mustRemove(t *testing.T, path string) {
	t.Helper()
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	if exit, ok := err.(*exec.ExitError); ok {
		return exit.ExitCode()
	}
	return -1
}

// The Bash book toolchain is qualified here on Unix; Windows product tests
// are independent of this optional documentation renderer.
func requireFigureTools(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("book shell toolchain requires Unix paths")
	}
	for _, tool := range []string{"bash", "grep", "sed", "sort", "basename", "awk", "shasum", "cut", "mktemp", "cmp", "wc", "tr"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("book checker prerequisite unavailable: %s", tool)
		}
	}
}

// The render pass and the kit need python3 3.11+ (tomllib) and exactly the
// Graphviz version pinned in kit/render.py.
func requireRenderTools(t *testing.T) {
	t.Helper()
	requireFigureTools(t)
	if err := exec.Command("python3", "-c", "import sys; sys.exit(sys.version_info < (3, 11))").Run(); err != nil {
		t.Skip("python3 3.11+ unavailable; the figure render pass is not exercised")
	}
	pin := regexp.MustCompile(`(?m)^GRAPHVIZ_PIN = "([^"]+)"$`).FindStringSubmatch(read(t, "kit/render.py"))
	if pin == nil {
		t.Fatal("kit/render.py has no GRAPHVIZ_PIN")
	}
	out, err := exec.Command("dot", "-V").CombinedOutput()
	if err != nil || !strings.Contains(string(out), "version "+pin[1]+" ") {
		t.Skipf("Graphviz %s unavailable (found %q); the figure render pass is not exercised", pin[1], strings.TrimSpace(string(out)))
	}
}

const fixtureChapter = "# Глава\n\n![Рисунок](../diagrams/figure.svg)\n"

func manifestRow(stem, origin, sha, box string) string {
	return fmt.Sprintf("| %s | `%s.toml` | %s | %s | %s | Глава |\n", stem, stem, origin, sha, box)
}

const manifestHead = "| Рисунок | Исходник | Происхождение | SHA-256 копии | Поле | Главы |\n| --- | --- | --- | --- | --- | --- |\n"

// A book tree in a temp dir: the checker under test, one TOML source with its
// SVG render, a manifest row and one chapter linking the figure without a
// box class. The structural pass reads inventories only, so the SVG need not
// be a real render here.
func figureFixture(t *testing.T) string {
	t.Helper()
	requireFigureTools(t)
	root := t.TempDir()
	script := "check-figures.sh"
	if override := os.Getenv("BOOK_FIGURE_CHECKER"); override != "" {
		script = override
	}
	write(t, filepath.Join(root, "check-figures.sh"), read(t, script))
	write(t, filepath.Join(root, "src", "figure.toml"), "layout = \"stack\"\n")
	write(t, filepath.Join(root, "src", "MANIFEST.md"), manifestHead+manifestRow("figure", "книга", "—", "обычное"))
	write(t, filepath.Join(root, "diagrams", "figure.svg"), "<svg/>")
	write(t, filepath.Join(root, "ru", "00-chapter.md"), fixtureChapter)
	return root
}

// These fixtures prove the structural pass accepts a consistent inventory and
// rejects each named inconsistency with its own diagnostic and exit 1:
// a chapter link without a render or without a source, a source or link
// without a manifest row, a link class that differs from the manifest box or
// from another link to the same figure, a render without a source, a source
// without a render and a changed upstream copy. They do not prove that a
// render matches its source; the render pass covers that.
func TestFigureCheckContracts(t *testing.T) {
	copySource := "@startuml\nA -> B\n@enduml\n"
	cases := []struct {
		name string
		edit func(*testing.T, string)
		want string
	}{
		{"valid", func(*testing.T, string) {}, ""},
		{"valid_tall", func(t *testing.T, r string) {
			write(t, filepath.Join(r, "src", "MANIFEST.md"), manifestHead+manifestRow("figure", "книга", "—", "tall"))
			write(t, filepath.Join(r, "ru", "00-chapter.md"), "# Глава\n\n![Рисунок](../diagrams/figure.svg){.tall}\n")
		}, ""},
		{"valid_copy", func(t *testing.T, r string) {
			write(t, filepath.Join(r, "src", "copy.puml"), copySource)
			write(t, filepath.Join(r, "diagrams", "copy.png"), "png")
			row := fmt.Sprintf("| copy | `copy.puml` | копия | %x | обычное | Глава |\n", sha256.Sum256([]byte(copySource)))
			write(t, filepath.Join(r, "src", "MANIFEST.md"), manifestHead+manifestRow("figure", "книга", "—", "обычное")+row)
		}, ""},
		{"missing_render", func(t *testing.T, r string) { mustRemove(t, filepath.Join(r, "diagrams", "figure.svg")) }, "missing render: diagrams/figure.svg"},
		{"missing_render_copy", func(t *testing.T, r string) {
			write(t, filepath.Join(r, "src", "copy.puml"), copySource)
			row := fmt.Sprintf("| copy | `copy.puml` | копия | %x | обычное | Глава |\n", sha256.Sum256([]byte(copySource)))
			write(t, filepath.Join(r, "src", "MANIFEST.md"), manifestHead+manifestRow("figure", "книга", "—", "обычное")+row)
			write(t, filepath.Join(r, "ru", "01-copy.md"), "# Копия\n\n![Копия](../diagrams/copy.png)\n")
		}, "missing render: diagrams/copy.png"},
		{"missing_source", func(t *testing.T, r string) { mustRemove(t, filepath.Join(r, "src", "figure.toml")) }, "missing source: src/figure.toml"},
		{"missing_row", func(t *testing.T, r string) { write(t, filepath.Join(r, "src", "MANIFEST.md"), manifestHead) }, "missing manifest row: figure"},
		{"source_without_row", func(t *testing.T, r string) {
			write(t, filepath.Join(r, "src", "extra.toml"), "layout = \"stack\"\n")
			write(t, filepath.Join(r, "diagrams", "extra.svg"), "<svg/>")
		}, "source without manifest row: extra"},
		{"box_class_mismatch", func(t *testing.T, r string) {
			write(t, filepath.Join(r, "ru", "00-chapter.md"), "# Глава\n\n![Рисунок](../diagrams/figure.svg){.tall}\n")
		}, "manifest box differs from the link class: figure"},
		{"mixed_box_links", func(t *testing.T, r string) {
			write(t, filepath.Join(r, "ru", "01-other.md"), "# Другая\n\n![Рисунок](../diagrams/figure.svg){.tall}\n")
		}, "linked both with and without {.tall}: figure.svg"},
		{"orphan_render", func(t *testing.T, r string) { write(t, filepath.Join(r, "diagrams", "extra.svg"), "<svg/>") }, "render without source: diagrams/extra.svg"},
		{"source_without_render", func(t *testing.T, r string) {
			write(t, filepath.Join(r, "src", "extra.toml"), "layout = \"stack\"\n")
			write(t, filepath.Join(r, "src", "MANIFEST.md"), manifestHead+manifestRow("figure", "книга", "—", "обычное")+manifestRow("extra", "книга", "—", "обычное"))
		}, "source without render: src/extra.toml"},
		{"changed_copy", func(t *testing.T, r string) {
			write(t, filepath.Join(r, "src", "copy.puml"), "changed\n")
			write(t, filepath.Join(r, "diagrams", "copy.png"), "png")
			row := fmt.Sprintf("| copy | `copy.puml` | копия | %x | обычное | Глава |\n", sha256.Sum256([]byte(copySource)))
			write(t, filepath.Join(r, "src", "MANIFEST.md"), manifestHead+manifestRow("figure", "книга", "—", "обычное")+row)
		}, "upstream copy changed: copy"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := figureFixture(t)
			tc.edit(t, r)
			out, err := exec.Command("bash", filepath.Join(r, "check-figures.sh"), "--structural").CombinedOutput()
			if tc.want == "" {
				if err != nil || strings.TrimSpace(string(out)) != "figures: structural ok" {
					t.Fatalf("valid inventory refused: %s (%v)", out, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("contract expected rejection for %s", tc.name)
			}
			if exitCode(err) != 1 || !strings.Contains(string(out), tc.want) {
				t.Fatalf("wrong rejection, want %q: %s (%v)", tc.want, out, err)
			}
		})
	}
}

// Each bounded mutant removes only the named rejection from check-figures.sh.
// Its negative witness must then fail with the "expected rejection" assertion,
// while the valid inventory still passes; a setup failure does not count.
func TestFigureCheckNarrowingControls(t *testing.T) {
	requireFigureTools(t)
	original := read(t, "check-figures.sh")
	cases := []struct{ name, witness, needle, replacement string }{
		{"missing_render", "missing_render_copy", `    [ -f "$here/diagrams/$ref" ] || { echo "missing render: diagrams/$ref (used by a chapter)"; fail=1; }`, ":"},
		{"source_manifest", "source_without_row", `    [ -n "$r" ] || { echo "source without manifest row: $stem"; fail=1; continue; }`, `    [ -n "$r" ] || continue`},
		{"box_class", "box_class_mismatch", `    [ "$want" = "$boxes" ] || { echo "manifest box differs from the link class: $stem (manifest $(col "$r" 6), link $boxes)"; fail=1; }`, ":"},
		{"orphan_render", "orphan_render", `    [ -f "$here/src/$src" ] || { echo "render without source: diagrams/$file"; fail=1; }`, ":"},
		{"copy_hash", "changed_copy", `      [ "$got" = "$want" ] || { echo "upstream copy changed: $stem ($got != $want)"; fail=1; }`, ":"},
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if strings.Count(original, tc.needle) != 1 {
				t.Fatal("control plant no longer applies exactly once")
			}
			mutant := filepath.Join(t.TempDir(), "mutant.sh")
			write(t, mutant, strings.Replace(original, tc.needle, tc.replacement, 1))
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
				if exitCode(err) != 1 || !strings.Contains(string(out), "contract expected rejection for "+witness) {
					t.Fatalf("mutant not killed by intended assertion: %s (%v)", out, err)
				}
			}
		})
	}
}

// A seven-step stack fits the tall box but not the normal one (checked by the
// kit itself in the cases below).
func tallStack() string {
	var b strings.Builder
	b.WriteString("layout = \"stack\"\nalt = \"Проверка поля\"\n\n")
	for i := 0; i < 7; i++ {
		fmt.Fprintf(&b, "[nodes.n%d]\nrole = \"service\"\nname = \"Шаг %d\"\ntext = \"описание шага\"\n\n", i, i+1)
	}
	for i := 0; i < 6; i++ {
		fmt.Fprintf(&b, "[[links]]\nfrom = \"n%d\"\nto = \"n%d\"\n\n", i, i+1)
	}
	return b.String()
}

// With the real kit, python3 and the pinned Graphviz, the render pass accepts
// a fresh render in its box and refuses a stale SVG, a figure whose link puts
// it in a box it does not fit, and a Graphviz other than the pin. Freshness is
// byte equality with a new render; legibility is the kit's --require-fit.
func TestFigureRenderPass(t *testing.T) {
	requireRenderTools(t)
	kitFiles := []string{"render.py", "theme.toml", "metrics.json"}
	cases := []struct {
		name, link, box, want string
		edit                  func(*testing.T, string)
	}{
		{"fresh_tall", "{.tall}", "tall", "", nil},
		{"stale", "{.tall}", "tall", "stale render: diagrams/figure.svg", func(t *testing.T, r string) {
			p := filepath.Join(r, "diagrams", "figure.svg")
			write(t, p, read(t, p)+"<!-- edited -->\n")
		}},
		{"does_not_fit_normal", "", "обычное", "does not fit the normal box: src/figure.toml", nil},
		{"wrong_graphviz", "{.tall}", "tall", "render pass needs Graphviz", func(t *testing.T, r string) {
			dot := filepath.Join(r, "bin", "dot")
			write(t, dot, "#!/bin/sh\necho 'dot - graphviz version 1.0.0 (0)' >&2\n")
			if err := os.Chmod(dot, 0o700); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := figureFixture(t)
			for _, f := range kitFiles {
				write(t, filepath.Join(r, "kit", f), read(t, filepath.Join("kit", f)))
			}
			src := filepath.Join(r, "src", "figure.toml")
			write(t, src, tallStack())
			write(t, filepath.Join(r, "src", "MANIFEST.md"), manifestHead+manifestRow("figure", "книга", "—", tc.box))
			write(t, filepath.Join(r, "ru", "00-chapter.md"), "# Глава\n\n![Рисунок](../diagrams/figure.svg)"+tc.link+"\n")
			if out, err := exec.Command("python3", filepath.Join(r, "kit", "render.py"), src, filepath.Join(r, "diagrams", "figure.svg")).CombinedOutput(); err != nil {
				t.Fatalf("fixture render failed: %s (%v)", out, err)
			}
			if tc.edit != nil {
				tc.edit(t, r)
			}
			cmd := exec.Command("bash", filepath.Join(r, "check-figures.sh"))
			cmd.Env = append(os.Environ(), "PATH="+filepath.Join(r, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))
			out, err := cmd.CombinedOutput()
			if tc.want == "" {
				if err != nil || strings.TrimSpace(string(out)) != "figures: ok" {
					t.Fatalf("fresh fitting render refused: %s (%v)", out, err)
				}
				return
			}
			if exitCode(err) != 1 || !strings.Contains(string(out), tc.want) {
				t.Fatalf("render pass expected %q: %s (%v)", tc.want, out, err)
			}
		})
	}
}

// The committed figures, manifest and vendored kit pass both passes of the
// real checker: every src/*.toml has a fresh SVG that fits the box its
// manifest row names. The chapters are replaced by one synthetic chapter that
// links each manifest figure with that box, so this holds while chapter text
// is edited; the build checks the real chapter links.
func TestCommittedFigures(t *testing.T) {
	requireRenderTools(t)
	root := t.TempDir()
	write(t, filepath.Join(root, "check-figures.sh"), read(t, "check-figures.sh"))
	for _, dir := range []string{"kit", "src", "diagrams"} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			if !e.IsDir() {
				write(t, filepath.Join(root, dir, e.Name()), read(t, filepath.Join(dir, e.Name())))
			}
		}
	}
	var chapter strings.Builder
	chapter.WriteString("# Рисунки\n\n")
	rows := regexp.MustCompile(`(?m)^\| ([a-z0-9-]+) \| `+"`"+`[a-z0-9-]+\.toml`+"`"+` \|[^|]*\|[^|]*\| (tall|обычное) \|`).FindAllStringSubmatch(read(t, "src/MANIFEST.md"), -1)
	if len(rows) == 0 {
		t.Fatal("no figure rows in src/MANIFEST.md")
	}
	for _, m := range rows {
		class := ""
		if m[2] == "tall" {
			class = "{.tall}"
		}
		fmt.Fprintf(&chapter, "![%s](../diagrams/%s.svg)%s\n\n", m[1], m[1], class)
	}
	write(t, filepath.Join(root, "ru", "00-figures.md"), chapter.String())
	out, err := exec.Command("bash", filepath.Join(root, "check-figures.sh")).CombinedOutput()
	if err != nil || strings.TrimSpace(string(out)) != "figures: ok" {
		t.Fatalf("committed figures refused: %s (%v)", out, err)
	}
}

func stubTools(t *testing.T, root string, weasyprint string) {
	t.Helper()
	for name, body := range map[string]string{
		"pandoc": `#!/usr/bin/env bash
printf called > "$(dirname "$0")/pandoc-called"
while (( $# )); do
  if [ "$1" = --output ]; then printf html > "$2"; exit 0; fi
  shift
done
exit 1
`,
		"weasyprint": weasyprint,
	} {
		p := filepath.Join(root, "bin", name)
		write(t, p, body)
		if err := os.Chmod(p, 0o700); err != nil {
			t.Fatal(err)
		}
	}
}

const weasyprintOK = `#!/usr/bin/env bash
printf called > "$(dirname "$0")/weasyprint-called"
for last; do :; done
printf rendered > "$last"
`

// buildFixture adds the build script, the chapter check and stand-in
// renderers to a figure fixture.
func buildFixture(t *testing.T, weasyprint string) string {
	t.Helper()
	requireCheckTools(t)
	root := figureFixture(t)
	script := "build.sh"
	if override := os.Getenv("BOOK_BUILD_SCRIPT"); override != "" {
		script = override
	}
	write(t, filepath.Join(root, "build.sh"), read(t, script))
	write(t, filepath.Join(root, "check.sh"), read(t, "check.sh"))
	for _, f := range []string{"render.py", "theme.toml", "metrics.json"} {
		write(t, filepath.Join(root, "kit", f), read(t, filepath.Join("kit", f)))
	}
	stubTools(t, root, weasyprint)
	return root
}

// A real shell with instrumented renderer stand-ins proves the build order:
// the chapter check and the structural figure pass run before any renderer,
// the render pass runs unless BOOK_SKIP_FIGURE_RENDER=1 (and then warns), an
// image WeasyPrint failed to load fails the build, and every refusal leaves
// the prior PDF bytes in place. The valid build leaves no HTML in the book
// tree and cleans its private temp directory. Real PDF rendering is checked
// separately.
func TestBuildGates(t *testing.T) {
	cases := []struct {
		name, want   string
		skipRender   bool
		renderersRun bool
		edit         func(*testing.T, string)
		weasyprint   string
	}{
		{"valid", "", true, true, nil, weasyprintOK},
		{"orphan_render", "render without source:", true, false, func(t *testing.T, r string) {
			write(t, filepath.Join(r, "diagrams", "extra.svg"), "<svg/>")
		}, weasyprintOK},
		{"hygiene", "dash in prose", true, false, func(t *testing.T, r string) {
			write(t, filepath.Join(r, "ru", "00-chapter.md"), fixtureChapter+"\nДом — каталог.\n")
		}, weasyprintOK},
		{"render_gate", "render pass needs Graphviz", false, false, func(t *testing.T, r string) {
			dot := filepath.Join(r, "bin", "dot")
			write(t, dot, "#!/bin/sh\necho 'dot - graphviz version 1.0.0 (0)' >&2\n")
			if err := os.Chmod(dot, 0o700); err != nil {
				t.Fatal(err)
			}
		}, weasyprintOK},
		{"image_failure", "an image failed to load", true, true, nil, `#!/usr/bin/env bash
printf called > "$(dirname "$0")/weasyprint-called"
echo "WARNING: Failed to load image at ../diagrams/figure.svg" >&2
for last; do :; done
printf rendered > "$last"
`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := buildFixture(t, tc.weasyprint)
			if tc.edit != nil {
				tc.edit(t, root)
			}
			pdf := filepath.Join(root, "output.pdf")
			write(t, pdf, "prior PDF")
			tmp := t.TempDir()
			cmd := exec.Command("bash", filepath.Join(root, "build.sh"), "ru", pdf)
			cmd.Env = append(os.Environ(), "PATH="+filepath.Join(root, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"), "TMPDIR="+tmp)
			if tc.skipRender {
				cmd.Env = append(cmd.Env, "BOOK_SKIP_FIGURE_RENDER=1")
			}
			out, err := cmd.CombinedOutput()
			checkRenderers := func() {
				for _, name := range []string{"pandoc-called", "weasyprint-called"} {
					_, statErr := os.Stat(filepath.Join(root, "bin", name))
					if tc.renderersRun != (statErr == nil) {
						t.Fatalf("renderer %s ran=%v, want %v: %s", name, statErr == nil, tc.renderersRun, out)
					}
				}
			}
			if tc.want == "" {
				checkRenderers()
				if err != nil || read(t, pdf) != "rendered" || !strings.Contains(string(out), "BOOK_SKIP_FIGURE_RENDER=1, figure freshness and fit are not checked") {
					t.Fatalf("valid build failed: %s (%v)", out, err)
				}
				html, _ := filepath.Glob(filepath.Join(root, "*", "*.html"))
				top, _ := filepath.Glob(filepath.Join(root, "*.html"))
				left, _ := os.ReadDir(tmp)
				if len(html)+len(top) != 0 || len(left) != 0 {
					t.Fatalf("temporary output left behind: tree %v %v, TMPDIR %v", html, top, left)
				}
				return
			}
			if err == nil {
				t.Fatal("build contract expected rejection")
			}
			if !strings.Contains(string(out), tc.want) || read(t, pdf) != "prior PDF" {
				t.Fatalf("wrong rejection or PDF changed, want %q: %s (%v)", tc.want, out, err)
			}
			checkRenderers()
		})
	}
}

// Removing one gate from build.sh must let its witness through (the build
// succeeds and the "expected rejection" assertion fires) while the valid
// build still passes.
func TestBuildGateNarrowingControls(t *testing.T) {
	requireCheckTools(t)
	original := read(t, "build.sh")
	cases := []struct{ name, witness, needle, replacement string }{
		{"structural", "orphan_render", `bash "$here/check-figures.sh" --structural || { echo "build.sh: figure check failed" >&2; exit 1; }`, ":"},
		{"chapter_check", "hygiene", `bash "$here/check.sh" || { echo "build.sh: chapter check failed" >&2; exit 1; }`, ":"},
		{"image_load", "image_failure", `if grep -q 'Failed to load image' "$tmp/weasyprint.log"; then`, "if false; then"},
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if strings.Count(original, tc.needle) != 1 {
				t.Fatal("build control plant no longer applies exactly once")
			}
			mutant := filepath.Join(t.TempDir(), "build.sh")
			write(t, mutant, strings.Replace(original, tc.needle, tc.replacement, 1))
			for _, witness := range []string{"valid", tc.witness} {
				cmd := exec.Command(binary, "-test.run=^TestBuildGates$/^"+witness+"$", "-test.v")
				cmd.Env = append(os.Environ(), "BOOK_BUILD_SCRIPT="+mutant)
				out, err := cmd.CombinedOutput()
				if witness == "valid" {
					if err != nil {
						t.Fatalf("mutant broke valid setup: %s (%v)", out, err)
					}
					continue
				}
				if exitCode(err) != 1 || !strings.Contains(string(out), "build contract expected rejection") {
					t.Fatalf("build mutant not killed by intended assertion: %s (%v)", out, err)
				}
			}
		})
	}
}

// With real pandoc and the book filter, a figure keeps exactly the box class
// its link carries (the filter adds none by stem), a {.landscape} image puts
// its figure on a landscape page, and a short preposition binds to the next
// word with a no-break space. This checks the HTML, not the PDF layout.
func TestFilterFiguresAndTypography(t *testing.T) {
	pandoc, err := exec.LookPath("pandoc")
	if err != nil {
		t.Skip("pandoc unavailable; PDF toolchain check required separately")
	}
	run := func(t *testing.T, md string) string {
		t.Helper()
		cmd := exec.Command(pandoc, "-f", "markdown", "-t", "html5", "--lua-filter", statusFilter())
		cmd.Stdin = strings.NewReader(md)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("pandoc failed: %s (%v)", out, err)
		}
		return string(out)
	}
	t.Run("tall_from_link", func(t *testing.T) {
		out := run(t, "![Рисунок](../diagrams/context-order.svg){.tall}\n")
		if !strings.Contains(out, `class="tall"`) {
			t.Fatalf("link class lost: %s", out)
		}
	})
	t.Run("no_class_by_stem", func(t *testing.T) {
		out := run(t, "![Рисунок](../diagrams/context-order.svg)\n")
		if strings.Contains(out, "tall") {
			t.Fatalf("filter added a box class by stem: %s", out)
		}
	})
	t.Run("landscape", func(t *testing.T) {
		out := run(t, "![Рисунок](../diagrams/x.svg){.landscape}\n")
		if !regexp.MustCompile(`<figure[^>]*class="landscape"`).MatchString(out) {
			t.Fatalf("landscape figure class missing: %s", out)
		}
		if strings.Contains(run(t, "![Рисунок](../diagrams/x.svg)\n"), "landscape") {
			t.Fatal("landscape class added without the link class")
		}
	})
	t.Run("nbsp", func(t *testing.T) {
		out := run(t, "Файл лежит в доме.\n")
		if !strings.Contains(out, "в доме") || strings.Contains(out, "Файл лежит") {
			t.Fatalf("preposition binding wrong: %q", out)
		}
	})
}
