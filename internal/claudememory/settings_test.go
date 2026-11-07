package claudememory

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"
)

// A redirected Claude config directory must not change the native memory
// target. A relative operator home refuses rather than using the working tree.
func TestNativeFileUsesOperatorHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("CLAUDE_CONFIG_DIR", t.TempDir())
	if got, err := NativeFile(); err != nil || got != filepath.Join(home, ".claude", "CLAUDE.md") {
		t.Fatalf("native memory does not use operator home: %q, %v", got, err)
	}
	t.Setenv("HOME", "relative-home")
	t.Setenv("USERPROFILE", "relative-home")
	if got, err := NativeFile(); err == nil || got != "" {
		t.Fatal("relative operator home was resolved against working directory")
	}
}

// The merge adds only the native file, preserves existing exclusions/unrelated
// number text, and does not broaden the exclusion to project or managed roots.
func TestMergeNativeMemory(t *testing.T) {
	// These are path values only: no native directory or file is read or written.
	// Fixed literal expectations stay independent of Merge and the temp root.
	base := filepath.ToSlash(filepath.VolumeName(t.TempDir())+string(filepath.Separator)) + "curator-memory-fixture/"
	for _, tc := range []struct{ name, pattern string }{
		{"operator", "operator"},
		{"operator[1]*?{x}(y)!+@", "operator[[]1[]][*][?][{]x[}][(]y[)][!][+][@]"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			native := filepath.FromSlash(base + tc.name + "/.claude/CLAUDE.md")
			want := base + tc.pattern + "/.claude/CLAUDE.md"
			before := []byte(`{"claudeMdExcludes":["**/vendor/CLAUDE.md"],"n":1e999,"hooks":{"SessionStart":[]}}`)
			got, err := Merge(before, native)
			if err != nil {
				t.Fatal(err)
			}
			var settings map[string]json.RawMessage
			if err := json.Unmarshal(got, &settings); err != nil {
				t.Fatal(err)
			}
			var excludes []string
			if err := json.Unmarshal(settings["claudeMdExcludes"], &excludes); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(excludes, []string{"**/vendor/CLAUDE.md", want}) || string(settings["n"]) != "1e999" || string(settings["hooks"]) != `{"SessionStart":[]}` {
				t.Fatalf("merge changed unrelated settings or exclusion scope: %s", got)
			}
			again, err := Merge(got, native)
			if err != nil || !bytes.Equal(again, got) {
				t.Fatal("satisfied merge is not byte-identical")
			}
			for _, payload := range [][]byte{nil, []byte(`{}`), []byte(`{"claudeMdExcludes":[]}`)} {
				if result, err := Merge(payload, native); err != nil || !bytes.Contains(result, []byte("claudeMdExcludes")) {
					t.Fatalf("valid settings rejected: %v", err)
				}
			}
		})
	}
}

// Malformed settings/exclusion shapes refuse without yielding a replacement;
// nearby valid input is covered above so rejecting everything cannot pass.
func TestMergeRefusesMalformedSettings(t *testing.T) {
	native := filepath.Join(t.TempDir(), ".claude", "CLAUDE.md")
	for _, input := range []string{`{`, `null`, `[]`, `{"claudeMdExcludes":null}`, `{"claudeMdExcludes":"x"}`, `{"claudeMdExcludes":[1]}`, `{"claudeMdExcludes":[null]}`, "", `{"claudeMdExcludes":[],"claudeMdExcludes":[]}`} {
		if result, err := Merge([]byte(input), native); err == nil || result != nil {
			t.Fatalf("invalid settings produced a replacement: %q", input)
		}
	}
	if result, err := Merge(nil, "relative/.claude/CLAUDE.md"); err == nil || result != nil {
		t.Fatal("relative native path admitted")
	}
}
