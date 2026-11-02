package codexskillcheck

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func skillDocument(header, body string) []byte {
	return []byte("---\n" + header + "\n---\n" + body)
}

// These independent documents exercise the bounded grammar and exact first
// issue, including adjacent valid controls. They do not establish equivalence
// for deferred YAML keys, duplicates, optional-value types or empty strings.
func TestValidate(t *testing.T) {
	valid := "name: sample-skill\ndescription: Complete description"
	cases := []struct {
		name string
		data []byte
		code string
	}{
		{"valid", skillDocument(valid, "# Instructions\nUse the tool.\n"), ""},
		{"allowed-keys", skillDocument(valid+"\nlicense: MIT\nallowed-tools: Read\nmetadata:\n  category: examples", ""), ""},
		{"trimmed", skillDocument("name: ' sample-skill '\ndescription: ' Complete description '", ""), ""},
		{"missing-frontmatter", []byte(valid), "frontmatter_invalid"},
		{"unclosed-frontmatter", []byte("---\n" + valid), "frontmatter_invalid"},
		{"malformed-yaml", skillDocument("name: [", ""), "frontmatter_invalid"},
		{"nonmapping-yaml", skillDocument("- example", ""), "frontmatter_invalid"},
		{"invalid-utf8", append(skillDocument(valid, ""), 0xff), "utf8_invalid"},
		{"unknown-key", skillDocument(valid+"\nextra: true", ""), "key_unknown"},
		{"missing-name", skillDocument("description: Complete", ""), "name_missing"},
		{"missing-description", skillDocument("name: sample-skill", ""), "description_missing"},
		{"name-type", skillDocument("name: 42\ndescription: Complete", ""), "name_type"},
		{"description-type", skillDocument("name: sample-skill\ndescription: [Complete]", ""), "description_type"},
		{"uppercase-name", skillDocument("name: Sample\ndescription: Complete", ""), "name_invalid"},
		{"underscore-name", skillDocument("name: sample_skill\ndescription: Complete", ""), "name_invalid"},
		{"leading-hyphen", skillDocument("name: '-sample'\ndescription: Complete", ""), "name_invalid"},
		{"trailing-hyphen", skillDocument("name: 'sample-'\ndescription: Complete", ""), "name_invalid"},
		{"double-hyphen", skillDocument("name: sample--skill\ndescription: Complete", ""), "name_invalid"},
		{"name-64", skillDocument("name: "+strings.Repeat("a", 64)+"\ndescription: Complete", ""), ""},
		{"name-65", skillDocument("name: "+strings.Repeat("a", 65)+"\ndescription: Complete", ""), "name_too_long"},
		{"description-1024-runes", skillDocument("name: sample\ndescription: "+strings.Repeat("é", 1024), ""), ""},
		{"description-1025-runes", skillDocument("name: sample\ndescription: "+strings.Repeat("é", 1025), ""), "description_too_long"},
		{"description-todo", skillDocument("name: sample\ndescription: '[TODO: complete this]'", ""), "description_todo"},
		{"description-inline-todo", skillDocument("name: sample\ndescription: 'Explain [TODO: markers] literally.'", ""), ""},
		{"description-left-angle", skillDocument("name: sample\ndescription: 'Use <input'", ""), "description_angle"},
		{"description-right-angle", skillDocument("name: sample\ndescription: 'Use input>'", ""), "description_angle"},
		{"unknown-before-name", skillDocument("extra: true\nname: BAD", "[TODO: body]"), "key_unknown"},
		{"name-before-description", skillDocument("name: BAD\ndescription: '[TODO: description]'", "[TODO: body]"), "name_invalid"},
		{"name-grammar-before-length", skillDocument("name: "+strings.Repeat("A", 65)+"\ndescription: Complete", ""), "name_invalid"},
		{"description-length-before-todo", skillDocument("name: sample\ndescription: '[TODO: "+strings.Repeat("a", 1024)+"]'", ""), "description_too_long"},
		{"description-todo-before-angle", skillDocument("name: sample\ndescription: '[TODO: <description>]'", ""), "description_todo"},
		{"description-before-body", skillDocument("name: sample\ndescription: 'Use <input>'", "[TODO: body]"), "description_angle"},
		{"crlf", []byte(strings.ReplaceAll(string(skillDocument(valid, "Instructions\n")), "\n", "\r\n")), ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := bytes.Clone(tc.data)
			assertResult(t, Validate(tc.data), tc.code)
			if !bytes.Equal(before, tc.data) {
				t.Fatal("validation changed input bytes")
			}
		})
	}
}

// Fence state is exercised with both marker types, optional list prefixes,
// failed closing attempts and a subsequent outside placeholder. These are
// source-projection controls, not general Markdown parsing claims.
func TestBodyTODOFenceState(t *testing.T) {
	cases := []struct {
		name string
		body string
		code string
	}{
		{"outside", "[TODO: finish]\n", "body_todo"},
		{"three-spaces", "   [TODO: finish] \t\n", "body_todo"},
		{"four-spaces", "    [TODO: literal]\n", ""},
		{"tab-prefix", "\t[TODO: literal]\n", ""},
		{"inline", "Explain [TODO: literal] here.\n", ""},
		{"list-todo", "- [TODO: literal]\n", ""},
		{"backticks", "```text\n[TODO: literal]\n```\n", ""},
		{"tildes", "~~~text\n[TODO: literal]\n~~~\n", ""},
		{"bullet-prefix", "- ```text\n[TODO: literal]\n- ```\n", ""},
		{"plus-prefix", "+ ~~~text\n[TODO: literal]\n+ ~~~\n", ""},
		{"star-prefix", "* ```text\n[TODO: literal]\n* ```\n", ""},
		{"number-dot-prefix", "12. ```text\n[TODO: literal]\n12. ```\n", ""},
		{"number-paren-prefix", "12) ~~~text\n[TODO: literal]\n12) ~~~\n", ""},
		{"unicode-decimal-prefix", "١. ```text\n[TODO: literal]\n١. ```\n", ""},
		{"unicode-decimal-close-restores", "١) ~~~text\n[TODO: literal]\n٢) ~~~\n[TODO: finish]\n", "body_todo"},
		{"close-then-todo", "```\n[TODO: literal]\n```\n[TODO: finish]\n", "body_todo"},
		{"tilde-close-then-todo", "~~~\n[TODO: literal]\n~~~\n[TODO: finish]\n", "body_todo"},
		{"wrong-marker-not-close", "```\n~~~\n[TODO: literal]\n", ""},
		{"short-marker-not-close", "````\n```\n[TODO: literal]\n", ""},
		{"suffix-not-close", "```\n``` text\n[TODO: literal]\n", ""},
		{"longer-marker-closes", "```\n````\n[TODO: finish]\n", "body_todo"},
		{"list-close-restores", "- ~~~text\n[TODO: literal]\n2) ~~~~ \t\n[TODO: finish]\n", "body_todo"},
		{"unclosed-fence-permitted", "```text\n[TODO: literal]\n", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertResult(t, Validate(skillDocument("name: sample\ndescription: Complete", tc.body)), tc.code)
		})
	}
}

func assertResult(t *testing.T, result Result, code string) {
	t.Helper()
	if result.SchemaVersion != 1 || result.Contract != "codex-quick-validate-v1" {
		t.Fatalf("unexpected result identity: %#v", result)
	}
	if code == "" {
		if result.Outcome != "valid" || result.ExitCode() != 0 || result.Issues == nil || len(result.Issues) != 0 {
			t.Fatalf("valid control rejected: %#v", result)
		}
		return
	}
	if result.Outcome != "invalid" || result.ExitCode() != 1 || !reflect.DeepEqual(result.Issues, []Issue{{Code: code}}) {
		t.Fatalf("want invalid/%s, got %#v", code, result)
	}
}

// Only SKILL.md is consulted; directory/name equality and package manifests
// are unnecessary. Both content refusal and I/O failure preserve existing data.
func TestCheckReadOnlyAndIOFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	for _, tc := range []struct{ body, code string }{{"Instructions", ""}, {"[TODO: finish]", "body_todo"}} {
		data := skillDocument("name: independent-name\ndescription: Complete", tc.body)
		if err := os.WriteFile(path, data, 0o640); err != nil {
			t.Fatal(err)
		}
		assertResult(t, Check(dir), tc.code)
		after, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(after, data) {
			t.Fatalf("input changed: %v", err)
		}
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) != 1 || entries[0].Name() != "SKILL.md" {
			t.Fatalf("unexpected input directory change: %v, %v", entries, err)
		}
	}
	missing := t.TempDir()
	result := Check(missing)
	if result.ExitCode() != 3 || result.Outcome != "setup_error" || !reflect.DeepEqual(result.Issues, []Issue{{Code: "skill_io"}}) {
		t.Fatalf("missing SKILL.md mislabeled as content verdict: %#v", result)
	}
	entries, err := os.ReadDir(missing)
	if err != nil || len(entries) != 0 {
		t.Fatalf("I/O failure changed input directory: %v, %v", entries, err)
	}
}
