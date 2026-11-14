package claudememory

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"
)

// Adding the rules exclusion upgrades a file-only document without duplicating
// either owned pattern or changing unrelated number text. The completed
// document is a byte-identical positive control, including its whitespace.
func TestMergeNativeRulesUpgrade(t *testing.T) {
	base := filepath.ToSlash(filepath.VolumeName(t.TempDir())+string(filepath.Separator)) + "curator-memory-fixture/operator"
	native := filepath.FromSlash(base + "/.claude/CLAUDE.md")
	for _, prior := range [][]string{{base + "/.claude/CLAUDE.md"}, {base + "/.claude/rules/**"}, {base + "/.claude/CLAUDE.md", base + "/.claude/rules/**"}} {
		list, err := json.Marshal(prior)
		if err != nil {
			t.Fatal(err)
		}
		input := append([]byte("{ \"n\":7.50, \"claudeMdExcludes\":"), list...)
		input = append(input, []byte(" }\n")...)
		got, err := Merge(input, native)
		if err != nil {
			t.Fatal(err)
		}
		var settings struct {
			Excludes []string        `json:"claudeMdExcludes"`
			N        json.RawMessage `json:"n"`
		}
		if err := json.Unmarshal(got, &settings); err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{base + "/.claude/CLAUDE.md", base + "/.claude/rules/**"} {
			count := 0
			for _, pattern := range settings.Excludes {
				if pattern == want {
					count++
				}
			}
			if count != 1 {
				t.Fatalf("native rules upgrade missing or duplicate exclusion %q: %s", want, got)
			}
		}
		if string(settings.N) != "7.50" {
			t.Fatal("upgrade changed unrelated number")
		}
		if len(prior) == 2 && !bytes.Equal(input, got) {
			t.Fatal("complete settings rewritten")
		}
		again, err := Merge(got, native)
		if err != nil || !bytes.Equal(got, again) {
			t.Fatal("upgraded settings not idempotent")
		}
	}
}
