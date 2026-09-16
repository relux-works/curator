package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func packageAt(t *testing.T, root, folder, name string) string {
	t.Helper()
	dir := filepath.Join(root, filepath.FromSlash(folder))
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: "+name+"\ndescription: A skill\n---\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}
func expandFixture(t *testing.T, root, selectors string) (*Manifest, error) {
	t.Helper()
	return ParseBytesWithOptions([]byte(fmt.Sprintf(`{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[%s]}`, selectors)), filepath.Join(root, Name), draftOptions)
}
func TestExpandOrderedIndividualAndCollections(t *testing.T) {
	root := t.TempDir()
	packageAt(t, root, "skills/zulu", "last")
	packageAt(t, root, "skills/alpha", "first")
	packageAt(t, root, "skills/alpha/nested", "not-discovered")
	packageAt(t, root, "skills/excluded", "unused")
	m, err := expandFixture(t, root, `{"from":"s","directory":"skills","include":["*","alpha"],"exclude":["excluded","absent"]}`)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Expand(m, ExpansionOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var names, dirs []string
	for _, s := range got {
		names = append(names, s.Decl.Name)
		dirs = append(dirs, s.Directory)
		if s.Index != 0 || s.Spec == nil || s.Decl.Selector.Collection {
			t.Fatalf("not a skill-level selection: %+v", s)
		}
	}
	if !reflect.DeepEqual(names, []string{"first", "last"}) || !reflect.DeepEqual(dirs, []string{"skills/alpha", "skills/zulu"}) {
		t.Fatalf("%v %v", names, dirs)
	}
	m, err = expandFixture(t, root, `{"name":"last","from":"s","directory":"skills/zulu"},{"name":"old","source":"configured","tag":"v1"}`)
	if err != nil {
		t.Fatal(err)
	}
	got, err = Expand(m, ExpansionOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Index != 0 || got[1].Index != 1 || got[1].Decl.Source != "configured" || got[1].Path != "" || got[1].Spec != nil {
		t.Fatalf("%+v", got)
	}
}

func TestExpandRefusals(t *testing.T) {
	tests := []struct {
		name, selection, code string
		setup                 func(*testing.T, string)
	}{
		{"missing-excluded", `{"from":"s","directory":"skills","include":["missing"],"exclude":["missing"]}`, "source_member_missing", nil},
		{"file-excluded", `{"from":"s","directory":"skills","include":["file"],"exclude":["file"]}`, "source_member_invalid", nil},
		{"empty", `{"from":"s","directory":"skills","include":["good"],"exclude":["good"]}`, "source_selection_invalid", nil},
		{"missing-frontmatter", `{"from":"s","directory":"skills","include":["*"]}`, "source_member_invalid", func(t *testing.T, r string) {
			if err := os.Mkdir(filepath.Join(r, "skills", "bad"), 0755); err != nil {
				t.Fatal(err)
			}
		}},
		{"invalid-discovered-name", `{"from":"s","directory":"skills","include":["*"]}`, "source_member_invalid", func(t *testing.T, r string) { packageAt(t, r, "skills/bad", "../bad") }},
		{"name-mismatch", `{"name":"other","from":"s","directory":"skills/good"}`, "source_member_invalid", nil},
		{"repeated-direct", `{"from":"s","directory":"skills","include":["good"]},{"name":"good","from":"s","directory":"skills/good"}`, "source_name_conflict", nil},
		{"legacy-collision", `{"name":"good","tag":"v1"},{"from":"s","directory":"skills","include":["good"]}`, "source_name_conflict", nil},
		{"case-collision", `{"from":"s","directory":"skills","include":["*"]}`, "source_name_conflict", func(t *testing.T, r string) { packageAt(t, r, "skills/second", "GOOD") }},
		{"version-collision", `{"from":"s","directory":"skills","include":["*"]}`, "source_name_conflict", func(t *testing.T, r string) { packageAt(t, r, "skills/second", "good") }},
		{"invalid-manifest", `{"name":"good","from":"s","directory":"skills/good"}`, "source_member_invalid", func(t *testing.T, r string) {
			writeExpansionFile(t, filepath.Join(r, "skills/good/agent-skill.json"), `{"schema_version":999}`)
		}},
		{"metadata-description", `{"name":"good","from":"s","directory":"skills/good"}`, "source_member_invalid", func(t *testing.T, r string) {
			writeExpansionFile(t, filepath.Join(r, "skills/good/SKILL.md"), "---\nname: good\ndescription: 123\n---\n")
		}},
		{"metadata-duplicate", `{"name":"good","from":"s","directory":"skills/good"}`, "source_member_invalid", func(t *testing.T, r string) {
			writeExpansionFile(t, filepath.Join(r, "skills/good/SKILL.md"), "---\nname: good\nname: bad\ndescription: ok\n---\n")
		}},
		{"metadata-triggers", `{"name":"good","from":"s","directory":"skills/good"}`, "source_member_invalid", func(t *testing.T, r string) {
			writeExpansionFile(t, filepath.Join(r, "skills/good/SKILL.md"), "---\nname: good\ndescription: ok\ntriggers: [42]\n---\n")
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			packageAt(t, root, "skills/good", "good")
			writeExpansionFile(t, filepath.Join(root, "skills/file"), "file")
			if tt.setup != nil {
				tt.setup(t, root)
			}
			m, err := expandFixture(t, root, tt.selection)
			if err != nil {
				t.Fatal(err)
			}
			got, err := Expand(m, ExpansionOptions{})
			if err == nil || !strings.Contains(err.Error(), tt.code) || got != nil {
				t.Fatalf("got %+v, %v; want %s", got, err, tt.code)
			}
		})
	}
}
func writeExpansionFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestExpandPhysicalBoundariesAndPruning(t *testing.T) {
	for _, mode := range []string{"escape", "managed", "git", "snapshot", "contained"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			packageAt(t, root, "skills/good", "good")
			target := packageAt(t, root, "other", "other")
			opts := ExpansionOptions{}
			want := ""
			switch mode {
			case "escape":
				target = packageAt(t, t.TempDir(), "outside", "outside")
				want = "source_selection_invalid"
			case "managed":
				target = packageAt(t, root, ".agents/skills/output", "output")
				want = "source_output_overlap"
			case "git":
				target = packageAt(t, root, ".git/objects", "output")
				want = "source_output_overlap"
			case "snapshot":
				target = packageAt(t, root, "store/output", "output")
				opts.OutputRoots = []string{filepath.Join(root, "store")}
				want = "source_output_overlap"
			}
			if err := os.Symlink(target, filepath.Join(root, "skills/link")); err != nil {
				t.Fatal(err)
			}
			m, err := expandFixture(t, root, `{"name":"other","from":"s","directory":"skills/link"}`)
			if err != nil {
				t.Fatal(err)
			}
			_, err = Expand(m, opts)
			if want == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("%v, want %s", err, want)
			}
			if mode == "managed" || mode == "snapshot" {
				m, err = expandFixture(t, root, `{"from":"s","directory":"skills","include":["*"]}`)
				if err != nil {
					t.Fatal(err)
				}
				got, err := Expand(m, opts)
				if err != nil || len(got) != 1 || got[0].Decl.Name != "good" {
					t.Fatalf("pruning: %+v %v", got, err)
				}
			}
		})
	}
}

func TestExpandAcquiredGitAndLiteralPath(t *testing.T) {
	root := t.TempDir()
	gitTree := t.TempDir()
	packageAt(t, gitTree, ".", "root-skill")
	m, err := ParseBytesWithOptions([]byte(`{"schema_version":2,"sources":{"s":{"repository":"example.org/kit","tag":"v1"}},"skills":[{"name":"root-skill","from":"s","directory":"."}]}`), filepath.Join(root, Name), draftOptions)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Expand(m, ExpansionOptions{}); err == nil {
		t.Fatal("unacquired Git accepted")
	}
	got, err := Expand(m, ExpansionOptions{GitRoots: map[string]string{"s": gitTree}})
	if err != nil || len(got) != 1 || got[0].Directory != "." {
		t.Fatalf("%+v %v", got, err)
	}
	packageAt(t, root, "$HOME/~literal", "root-skill")
	m.Sources["s"] = Source{Path: "$HOME/~literal"}
	got, err = Expand(m, ExpansionOptions{})
	if err != nil || !strings.Contains(got[0].Path, "$HOME") {
		t.Fatalf("literal path: %+v %v", got, err)
	}
}

func TestExpandDeclaredManifestIdentity(t *testing.T) {
	for _, name := range []string{"good", "other"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			packageAt(t, root, "skill", "good")
			writeExpansionFile(t, filepath.Join(root, "skill/agent-skill.json"), fmt.Sprintf(`{"schema_version":1,"name":%q}`, name))
			m, err := expandFixture(t, root, `{"name":"good","from":"s","directory":"skill"}`)
			if err != nil {
				t.Fatal(err)
			}
			_, err = Expand(m, ExpansionOptions{})
			if name == "good" && err != nil {
				t.Fatal(err)
			}
			if name != "good" && (err == nil || !strings.Contains(err.Error(), "source_member_invalid")) {
				t.Fatalf("%v", err)
			}
		})
	}
}

func TestExpandMalformedSelectorBeforeFilesystem(t *testing.T) {
	for _, selection := range []string{
		`{"name":"good","from":"missing","directory":"."}`,
		`{"name":"good","from":"s","directory":"../escape"}`,
		`{"from":"s","directory":".","include":["**"]}`,
		`{"from":"s","directory":".","include":["a*"]}`,
		`{"from":"s","directory":".","include":["a/b"]}`,
		`{"from":"s","directory":".","include":["*"],"exclude":["*"]}`,
	} {
		t.Run(selection, func(t *testing.T) {
			m, err := expandFixture(t, filepath.Join(t.TempDir(), "absent"), selection)
			if err == nil {
				_, err = Expand(m, ExpansionOptions{})
			}
			if err == nil {
				t.Fatal("invalid selector accepted")
			}
		})
	}
}

func TestExpandOutputReadFailure(t *testing.T) {
	root := t.TempDir()
	packageAt(t, root, "skill", "good")
	writeExpansionFile(t, filepath.Join(root, "file"), "x")
	m, err := expandFixture(t, root, `{"name":"good","from":"s","directory":"skill"}`)
	if err != nil {
		t.Fatal(err)
	}
	// A path below a regular file must be refused on every platform. On Unix
	// this surfaces as ENOTDIR ("cannot inspect"); on Windows the same path
	// reports "not found", so the ancestor walk must find the file ancestor
	// and refuse it as a non-directory. Both diagnostics share the
	// source_output_overlap class.
	for _, boundary := range []string{
		filepath.Join(root, "file", "child"),
		filepath.Join(root, "file", "child", "grandchild"),
		filepath.Join(root, "file"),
	} {
		_, err = Expand(m, ExpansionOptions{OutputRoots: []string{boundary}})
		if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
			t.Fatalf("%s: %v", boundary, err)
		}
	}
	// Absent directories remain allowed.
	if _, err := Expand(m, ExpansionOptions{OutputRoots: []string{filepath.Join(root, "absent", "child")}}); err != nil {
		t.Fatalf("absent output boundary: %v", err)
	}
}

func TestExpandMetadataLinks(t *testing.T) {
	for _, file := range []string{"SKILL.md", "agent-skill.json", "csk-skill.json", "agents"} {
		t.Run(file, func(t *testing.T) {
			root := t.TempDir()
			dir := packageAt(t, root, "skill", "good")
			outside := t.TempDir()
			target := filepath.Join(outside, "manifest")
			writeExpansionFile(t, target, `{"schema_version":1}`)
			if file == "SKILL.md" {
				if err := os.Remove(filepath.Join(dir, file)); err != nil {
					t.Fatal(err)
				}
				writeExpansionFile(t, target, "---\nname: good\ndescription: external\n---\n")
			}
			if file == "agents" {
				target = outside
				writeExpansionFile(t, filepath.Join(outside, "runtime.json"), `{}`)
			}
			if err := os.Symlink(target, filepath.Join(dir, file)); err != nil {
				t.Fatal(err)
			}
			m, err := expandFixture(t, root, `{"name":"good","from":"s","directory":"skill"}`)
			if err != nil {
				t.Fatal(err)
			}
			_, err = Expand(m, ExpansionOptions{})
			if err == nil || !strings.Contains(err.Error(), "source_member_invalid") {
				t.Fatalf("%v", err)
			}
		})
	}
}
