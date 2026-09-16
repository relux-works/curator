package manifest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

var draftOptions = ParseOptions{DraftSourcesV1: true}

func TestDraftPublishedSchemaCases(t *testing.T) {
	files, err := filepath.Glob("testdata/draft-sources-v1/skillfile-v2/*.json")
	if err != nil || len(files) != 41 {
		t.Fatalf("published corpus: %d files, %v", len(files), err)
	}
	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			payload, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			root := writeManifest(t, string(payload))
			_, err = LoadWithOptions(root, draftOptions)
			want := strings.HasPrefix(filepath.Base(file), "valid-")
			if (err == nil) != want {
				t.Fatalf("valid=%v: %v", want, err)
			}
		})
	}
}

func TestDraftCapabilityAdmission(t *testing.T) {
	for _, version := range []string{"0", "1", "2", "3", "2.5", "null", `"2"`, "1e100"} {
		payload := []byte(`{"schema_version":` + version + `,"skills":[]}`)
		for _, opted := range []bool{false, true} {
			_, err := ParseBytesWithOptions(payload, "/missing/Skillfile.json", ParseOptions{DraftSourcesV1: opted})
			want := version == "1" || version == "2" && opted
			if (err == nil) != want {
				t.Errorf("version=%s opted=%v err=%v", version, opted, err)
			}
		}
	}
	root := writeManifest(t, `{"schema_version":2,"skills":[]}`)
	if _, err := Load(root); err == nil {
		t.Fatal("legacy Load accepted draft")
	}
	if _, err := ParseBytes([]byte(`{"schema_version":2,"skills":[]}`), "Skillfile.json"); err == nil {
		t.Fatal("legacy ParseBytes accepted draft")
	}
	if _, err := Parse(map[string]any{"schema_version": float64(2), "skills": []any{}}, "Skillfile.json"); err == nil {
		t.Fatal("legacy Parse accepted draft")
	}
}

func TestDraftPreservesEveryLegacyField(t *testing.T) {
	payload := `{"schema_version":1,"project":{"alias":"Demo iOS"},"agents":["codex_cli","claude_code"],"locale":"ru","skills":[{"name":"one","source":"références/文書","git":"http://example.org/one","branch":"legacy branch"},{"name":"two","revision":"abc123"},{"name":"three","tag":"v1"}]}`
	legacy, err := ParseBytes([]byte(payload), "/project/Skillfile.json")
	if err != nil {
		t.Fatal(err)
	}
	draft, err := ParseBytesWithOptions([]byte(strings.Replace(payload, `"schema_version":1`, `"schema_version":2`, 1)), "/project/Skillfile.json", draftOptions)
	if err != nil {
		t.Fatal(err)
	}
	draft.SchemaVersion = 1
	if !reflect.DeepEqual(legacy, draft) {
		t.Fatalf("legacy fields changed:\n%+v\n%+v", legacy, draft)
	}
	if draft.Skills[0].Source != "références/文書" || draft.Skills[1].Source != "two" {
		t.Fatal("configured-root source meaning changed")
	}
}

func TestDraftSourceUnionAndSelectors(t *testing.T) {
	payload := `{"schema_version":2,"sources":{"rel":{"path":"../$HOME/~kit"},"abs":{"path":"/missing/kit"},"url":{"git":"ssh://git@EXAMPLE.org/team/kit.git","branch":"feature/work"},"logical":{"repository":"example.org/team/kit","revision":"` + strings.Repeat("a", 64) + `"}},"skills":[{"name":"one","from":"rel","directory":"."},{"from":"logical","directory":"skills","include":["*","two"],"exclude":["three"]}]}`
	m, err := LoadWithOptions(writeManifest(t, payload), draftOptions)
	if err != nil {
		t.Fatal(err)
	}
	if m.Sources["rel"].Path != "../$HOME/~kit" || m.Sources["abs"].Path != "/missing/kit" || m.Sources["url"].Identity != "example.org/team/kit" || m.Sources["url"].Git != "ssh://git@EXAMPLE.org/team/kit.git" || m.Sources["logical"].Repository != "example.org/team/kit" {
		t.Fatalf("sources=%+v", m.Sources)
	}
	if m.Skills[0].Source != "" || m.Skills[0].Ref != (Ref{}) || m.Skills[0].Selector.Directory != "." || !m.Skills[1].Selector.Collection || !reflect.DeepEqual(m.Skills[1].Selector.Include, []string{"*", "two"}) || !reflect.DeepEqual(m.Skills[1].Selector.Exclude, []string{"three"}) {
		t.Fatalf("selectors=%+v", m.Skills)
	}
}

func TestDraftNegativeAdmission(t *testing.T) {
	sources := []string{
		`null`, `[]`, `{"Bad/Alias":{"path":"."}}`, `{"s":null}`, `{"s":{"path":".","tag":"v1"}}`,
		`{"s":{"git":"https://example.org/a","repository":"example.org/a","tag":"v1"}}`,
		`{"s":{"repository":"EXAMPLE.org/a","tag":"v1"}}`, `{"s":{"repository":"example.org/a.git","tag":"v1"}}`,
		`{"s":{"repository":"example.org/a","tag":"v1","branch":"main"}}`,
		`{"s":{"git":"http://example.org/a","tag":"v1"}}`, `{"s":{"git":"https://token@example.org/a","tag":"v1"}}`,
		`{"s":{"git":"ssh://git@example.org:22/a","tag":"v1"}}`,
		`{"s":{"repository":"example.org/a","revision":"abc123"}}`, `{"s":{"repository":"example.org/a","branch":"bad..ref"}}`,
		`{"s":{"repository":"example.org/a","tag":"foo.lock"}}`, `{"s":{"repository":"example.org/a","branch":null}}`,
	}
	for _, source := range sources {
		_, err := ParseBytesWithOptions([]byte(`{"schema_version":2,"sources":`+source+`,"skills":[]}`), "/nonexistent/Skillfile.json", draftOptions)
		if err == nil {
			t.Errorf("accepted sources %s", source)
		}
	}
	selections := []string{
		`{"name":"a","from":"unknown","directory":"."}`, `{"name":"a","from":"S","directory":"."}`,
		`{"name":"a","from":"s","directory":".","git":null}`, `{"name":"a","from":"s","directory":".","include":["*"]}`,
		`{"from":"s","directory":".","include":[]}`, `{"from":"s","directory":".","include":["*","*"]}`,
		`{"from":"s","directory":".","include":["*"],"exclude":null}`, `{"from":"s","directory":".","include":["*"],"exclude":["a","a"]}`,
	}
	for _, dir := range []string{"", "..", "../a", "/a", "a//b", "a/./b", `a\b`, "C:/a", "NUL", "a/COM1.txt", "a.", "a ", "a*", "a?", "a[b]"} {
		encoded, _ := json.Marshal(dir)
		selections = append(selections, `{"name":"a","from":"s","directory":`+string(encoded)+`}`)
	}
	for _, selection := range selections {
		_, err := ParseBytesWithOptions([]byte(`{"schema_version":2,"sources":{"s":{"path":"/nonexistent"}},"skills":[`+selection+`]}`), "/nonexistent/Skillfile.json", draftOptions)
		if err == nil {
			t.Errorf("accepted selection %s", selection)
		}
	}
	for _, payload := range []string{
		`{"schema_version":2,"schema_version":1,"skills":[]}`,
		`{"schema_version":2,"sources":{"s":{"path":"a","path":"b"}},"skills":[]}`,
		`{"schema_version":1,"sources":{},"skills":[]}`,
		`{"schema_version":2,"DraftSourcesV1":true,"skills":[]}`,
		`{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"name":"a","tag":"v1"},{"name":"a","from":"s","directory":"."}]}`,
	} {
		if _, err := ParseBytesWithOptions([]byte(payload), "/missing/Skillfile.json", draftOptions); err == nil {
			t.Errorf("accepted %s", payload)
		}
	}
}

func TestDraftValidRefsAndEndpoints(t *testing.T) {
	for _, endpoint := range []string{`"git":"https://EXAMPLE.org/a.git"`, `"git":"ssh://git@example.org/a"`, `"git":"git@example.org:a"`, `"repository":"example.org/a"`} {
		for _, ref := range []string{`"tag":"v1"`, `"branch":"feature/x"`, `"revision":"` + strings.Repeat("a", 40) + `"`, `"revision":"` + strings.Repeat("b", 64) + `"`} {
			payload := `{"schema_version":2,"sources":{"s":{` + endpoint + `,` + ref + `}},"skills":[]}`
			if _, err := ParseBytesWithOptions([]byte(payload), "/missing/Skillfile.json", draftOptions); err != nil {
				t.Errorf("%s: %v", payload, err)
			}
		}
	}
}

func TestDraftRefGrammarBounds(t *testing.T) {
	for _, kind := range []string{"tag", "branch"} {
		for _, value := range []string{"@", "/a", "a/", "a.", "a//b", "a..b", "a@{b", "a b", "a\x7fb", "a~b", "a^b", "a:b", "a?b", "a*b", "a[b", `a\b`, ".a", "x/.a", "x/a.lock", strings.Repeat("é", 256)} {
			raw, _ := json.Marshal(value)
			payload := `{"schema_version":2,"sources":{"s":{"repository":"example.org/a","` + kind + `":` + string(raw) + `}},"skills":[]}`
			if _, err := ParseBytesWithOptions([]byte(payload), "/missing/Skillfile.json", draftOptions); err == nil {
				t.Errorf("accepted %s %q", kind, value)
			}
		}
		for _, value := range []string{"a", "é", strings.Repeat("é", 255), "heads/a", "a.locked"} {
			raw, _ := json.Marshal(value)
			payload := `{"schema_version":2,"sources":{"s":{"repository":"example.org/a","` + kind + `":` + string(raw) + `}},"skills":[]}`
			if _, err := ParseBytesWithOptions([]byte(payload), "/missing/Skillfile.json", draftOptions); err != nil {
				t.Errorf("rejected %s %q: %v", kind, value, err)
			}
		}
	}
}
