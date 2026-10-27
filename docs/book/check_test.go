package book_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// check.sh is POSIX sh with grep, find and perl; it refuses (exit 2) without a
// UTF-8 locale, which the tests treat as an unavailable prerequisite.
func requireCheckTools(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("book shell toolchain requires Unix paths")
	}
	for _, tool := range []string{"sh", "grep", "find", "sort", "awk", "perl", "wc", "tr"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("chapter check prerequisite unavailable: %s", tool)
		}
	}
}

const (
	chapterA = "# Для кого\n\nКнига ссылается на главу \"Вторая глава\" и на главы \"Вторая глава\" и \"Где мы\".\n"
	// Clean prose that sits next to several rules: a numeric range with an
	// en dash, a reference wrapped across lines, a figure, a relative link with
	// an anchor, an external link, and a reference inside a fenced code block.
	chapterB = "# Вторая глава\n\n" +
		"Уровни 0–2 и L0–L2 описаны в главе \"Для\nкого\".\n\n" +
		"![Рисунок](../diagrams/fig.svg){.tall}\n\n" +
		"См. [начало](00-dlya-kogo.md#x) и [сайт](https://example.invalid/a).\n\n" +
		"```sh\n# глава \"Нет такой\"\necho [x](nope.md)\n```\n"
	// The state chapter may carry status marks and delivery hedges.
	chapterState = "# Где мы\n\n| Часть | Статус |\n| --- | --- |\n| Дом | ⟦решено⟧, пока в плане |\n"
)

func chapterFixture(t *testing.T) string {
	t.Helper()
	requireCheckTools(t)
	root := t.TempDir()
	script := "check.sh"
	if override := os.Getenv("BOOK_CHAPTER_CHECKER"); override != "" {
		script = override
	}
	write(t, filepath.Join(root, "check.sh"), read(t, script))
	write(t, filepath.Join(root, "ru", "00-dlya-kogo.md"), chapterA)
	write(t, filepath.Join(root, "ru", "01-vtoraya.md"), chapterB)
	write(t, filepath.Join(root, "ru", "11-gde-my.md"), chapterState)
	write(t, filepath.Join(root, "diagrams", "fig.svg"), "<svg/>")
	return root
}

func runChapterCheck(t *testing.T, root string) (string, error) {
	t.Helper()
	out, err := exec.Command("sh", filepath.Join(root, "check.sh")).CombinedOutput()
	if exitCode(err) == 2 && strings.Contains(string(out), "no UTF-8 locale") {
		t.Skip("no UTF-8 locale for check.sh")
	}
	return string(out), err
}

// The clean fixture passes, which proves that numeric-range en dashes, a
// wrapped and a plural chapter reference to existing titles, an existing
// figure, an anchored relative link, an external link, code-block content
// and markers or hedges in the state chapter are not reported.
func TestChapterCheckClean(t *testing.T) {
	out, err := runChapterCheck(t, chapterFixture(t))
	if err != nil || strings.TrimSpace(out) != "check.sh: ok (3 files)" {
		t.Fatalf("clean chapters refused: %s (%v)", out, err)
	}
}

// Each rule fires on one planted violation in an isolated fixture: the
// output carries the rule's section header and file:line of the planted
// line, and the check exits 1. The planted line is appended to the second
// chapter (line 14), except where a case needs its own file. For the plural
// reference only the unknown title is reported, not the existing one.
func TestChapterCheckRules(t *testing.T) {
	const at = "ru/01-vtoraya.md:14:"
	cases := []struct {
		name, line, file, label, where, absent string
	}{
		{"status_marker", "Дом готов ⟦решено⟧.", "", "== status marker", at, ""},
		{"split_marker", "Дом готов ⟦решено.", "", "== status marker", at, ""},
		{"hedge_poka", "Это пока не сделано.", "", "== delivery hedge", at, ""},
		{"hedge_eshche", "Механизм ещё не включён.", "", "== delivery hedge", at, ""},
		{"hedge_v0", "Так было в v0.", "", "== delivery hedge", at, ""},
		{"hedge_realized", "Канал не реализован.", "", "== delivery hedge", at, ""},
		{"em_dash", "Дом — каталог.", "", "== dash in prose", at, ""},
		{"en_dash", "Дом – каталог.", "", "== dash in prose", at, ""},
		{"guillemets", "Глава «Дом».", "", "== guillemets", at, ""},
		{"antithesis", "Это не дом, а каталог.", "", "== antithesis", at, ""},
		{"filler", "Стоит отметить, что дом есть.", "", "== filler opener", at, ""},
		{"marketing", "Мощный менеджер.", "", "== marketing adjective", at, ""},
		{"task_id", "См. TASK-261027-1f3v75.", "", "== board, run or review identifier", at, ""},
		{"bug_id", "См. BUG-12.", "", "== board, run or review identifier", at, ""},
		{"decision_number", "Так решено в Decision 0028.", "", "== decision number", at, ""},
		{"decision_short", "Так решено в D-001.", "", "== decision number", at, ""},
		{"pr_number", "Это пришло в PR #42.", "", "== pull request number", at, ""},
		{"users_path", "Файл /Users/iv/notes.md.", "", "== local path", at, ""},
		{"tilde_path", "Каталог `~/Werkstatt`.", "", "== local path", at, ""},
		{"bad_reference", "Подробно в главе \"Нет такой\".", "", "== chapter reference", at + ` no chapter titled "Нет такой"`, ""},
		{"bad_plural", "См. главы \"Для кого\" и \"Нет такой\".", "", "== chapter reference", at + ` no chapter titled "Нет такой"`, `no chapter titled "Для кого"`},
		{"missing_figure", "![Рисунок](../diagrams/none.svg)", "", "== missing figure", at + " ../diagrams/none.svg", ""},
		{"broken_link", "См. [план](nope.md).", "", "== broken link", at + " nope.md", ""},
		{"missing_heading", "Текст без заголовка.", "ru/02-bez.md", "== chapter without a level-1 heading", "ru/02-bez.md", ""},
		{"marker_outside_state", "⟦решено⟧", "ru/12-slovar.md", "== status marker", "ru/12-slovar.md:2:", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := chapterFixture(t)
			if tc.file == "" {
				write(t, filepath.Join(root, "ru", "01-vtoraya.md"), chapterB+tc.line+"\n")
			} else if tc.name == "missing_heading" {
				write(t, filepath.Join(root, tc.file), tc.line+"\n")
			} else {
				write(t, filepath.Join(root, tc.file), "# Словарь\n"+tc.line+"\n")
			}
			out, err := runChapterCheck(t, root)
			if err == nil {
				t.Fatalf("check expected rejection for %s: %s", tc.name, out)
			}
			if exitCode(err) != 1 || !strings.Contains(out, tc.label+"\n") || !strings.Contains(out, tc.where) {
				t.Fatalf("wrong rejection, want %q at %q: %s (%v)", tc.label, tc.where, out, err)
			}
			if tc.absent != "" && strings.Contains(out, tc.absent) {
				t.Fatalf("existing title reported: %s", out)
			}
		})
	}
}

// Removing only the chapter-reference report from check.sh must let the
// planted unknown title through (the rule test's rejection assertion fires),
// while the clean fixture still passes.
func TestChapterCheckNarrowingControl(t *testing.T) {
	requireCheckTools(t)
	original := read(t, "check.sh")
	needle := `print "chapter reference\t$f:" . $line->($pos) . ": no chapter titled \"$t\"\n" unless $titles{$t};`
	if strings.Count(original, needle) != 1 {
		t.Fatal("chapter-reference control plant no longer applies exactly once")
	}
	mutant := filepath.Join(t.TempDir(), "check.sh")
	write(t, mutant, strings.Replace(original, needle, "1;", 1))
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, run := range []string{"^TestChapterCheckClean$", "^TestChapterCheckRules$/^bad_reference$"} {
		cmd := exec.Command(binary, "-test.run="+run, "-test.v")
		cmd.Env = append(os.Environ(), "BOOK_CHAPTER_CHECKER="+mutant)
		out, err := cmd.CombinedOutput()
		if strings.Contains(run, "Clean") {
			if err != nil {
				t.Fatalf("mutant broke the clean fixture: %s (%v)", out, err)
			}
			continue
		}
		if exitCode(err) != 1 || !strings.Contains(string(out), "check expected rejection for bad_reference") {
			t.Fatalf("chapter-reference mutant not killed by intended assertion: %s (%v)", out, err)
		}
	}
}
