package closure

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/manifest"
)

func selectionProject(t *testing.T, root string, names []string, requirements map[string]map[string]any) *manifest.Manifest {
	t.Helper()
	for _, name := range names {
		dir := filepath.Join(root, "skills", name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: "+name+"\ndescription: Test\n---\n"), 0644); err != nil {
			t.Fatal(err)
		}
		spec := map[string]any{"schema_version": 4, "capabilities": map[string]any{}, "commands": map[string]any{}, "dependencies": map[string]any{"skills": requirements}}
		data, err := json.Marshal(spec)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "agent-skill.json"), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	m, err := manifest.ParseBytesWithOptions([]byte(`{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["*"]}]}`), filepath.Join(root, manifest.Name), manifest.ParseOptions{DraftSourcesV1: true})
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// Fixtures are immutable temporary trees. This boundary supplies already acquired
// local nodes; it makes no claim to exercise the future local snapshot capture.
func fixtureAcquisition(s manifest.Selection) (*Node, error) {
	return &Node{Name: s.Decl.Name, Snapshot: s.Path, Spec: s.Spec}, nil
}

func TestBuildExpandedDiamondAndLegacy(t *testing.T) {
	h := newHarness(t)
	h.skill("provider", nil, nil)
	m := selectionProject(t, t.TempDir(), []string{"zulu", "alpha"}, map[string]map[string]any{"provider": requirement("provider", "context")})
	nodes, err := BuildExpanded(Options{SkillsRoot: h.skillsRoot, Home: h.home}, m, manifest.ExpansionOptions{}, fixtureAcquisition, nil)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, n := range nodes {
		names = append(names, n.Name)
	}
	if !reflect.DeepEqual(names, []string{"provider", "alpha", "zulu"}) {
		t.Fatal(names)
	}
	if len(nodes[0].Edges) != 2 || nodes[1].Edges[0].Consumer != ProjectEdge || nodes[2].Decl.Selector.Directory != "skills/zulu" {
		t.Fatalf("skill-level edges lost: %+v", nodes)
	}
	m.Skills = append(m.Skills, manifest.Decl{Name: "legacy", Source: "legacy", Ref: manifest.Ref{Kind: "tag", Value: "v1"}})
	h.skill("legacy", nil, nil)
	nodes, err = BuildExpanded(Options{SkillsRoot: h.skillsRoot, Home: h.home}, m, manifest.ExpansionOptions{}, fixtureAcquisition, nil)
	if err != nil || len(nodes) != 4 {
		t.Fatalf("legacy mix: %v %v", nodes, err)
	}
}

func TestBuildExpandedRefusals(t *testing.T) {
	for _, mode := range []string{"version", "source", "case", "local-identity", "cycle", "missing-acquisition", "invalid-acquisition", "unexpanded"} {
		t.Run(mode, func(t *testing.T) {
			h := newHarness(t)
			repo := h.skill("provider", nil, nil)
			req := map[string]map[string]any{"provider": requirement("provider", "context")}
			m := selectionProject(t, t.TempDir(), []string{"alpha", "zulu"}, req)
			opts := Options{SkillsRoot: h.skillsRoot, Home: h.home}
			want := ""
			acquire := AcquireSelection(fixtureAcquisition)
			switch mode {
			case "version":
				if err := os.WriteFile(filepath.Join(repo, "new"), []byte("new"), 0644); err != nil {
					t.Fatal(err)
				}
				h.git(repo, "add", ".")
				h.git(repo, "commit", "-qm", "second")
				h.git(repo, "tag", "v2")
				m.Skills = append(m.Skills, manifest.Decl{Name: "provider", Source: "provider", Ref: manifest.Ref{Kind: "tag", Value: "v2"}})
				want = "version conflict"
			case "source":
				m.Skills = append(m.Skills, manifest.Decl{Name: "provider", Source: "provider", Git: "https://example.org/a", Ref: manifest.Ref{Kind: "tag", Value: "v1"}})
				acquire = func(s manifest.Selection) (*Node, error) {
					n, _ := fixtureAcquisition(s)
					r := n.Spec.Requirements["provider"]
					r.Git = "https://example.org/b"
					n.Spec.Requirements["provider"] = r
					return n, nil
				}
				want = "source conflict"
			case "case":
				m.Skills = append(m.Skills, manifest.Decl{Name: "PROVIDER", Source: "provider", Ref: manifest.Ref{Kind: "tag", Value: "v1"}})
				want = "source_name_conflict"
			case "local-identity":
				acquire = func(s manifest.Selection) (*Node, error) {
					n, _ := fixtureAcquisition(s)
					r := n.Spec.Requirements["provider"]
					delete(n.Spec.Requirements, "provider")
					r.Name = "alpha"
					n.Spec.Requirements["alpha"] = r
					return n, nil
				}
				want = "source_name_conflict"
			case "cycle":
				// Cycle detection in the legacy dependency subgraph stays active.
				h.skill("cycle", nil, map[string]map[string]any{"cycle": requirement("cycle", "context")})
				m.Skills = append(m.Skills, manifest.Decl{Name: "cycle", Source: "cycle", Ref: manifest.Ref{Kind: "tag", Value: "v1"}})
				want = "dependency cycle"
			case "missing-acquisition":
				acquire = nil
				want = "source_selection_invalid"
			case "invalid-acquisition":
				acquire = func(manifest.Selection) (*Node, error) { return &Node{Name: "forged"}, nil }
				want = "source_member_invalid"
			case "unexpanded":
				_, err := Build(opts, m, nil)
				if err == nil || !strings.Contains(err.Error(), "source_selection_invalid") {
					t.Fatalf("%v", err)
				}
				return
			}
			nodes, err := BuildExpanded(opts, m, manifest.ExpansionOptions{}, acquire, nil)
			if err == nil || !strings.Contains(err.Error(), want) || nodes != nil {
				t.Fatalf("%+v %v; want %s", nodes, err, want)
			}
		})
	}
}

func TestBuildExpandedGitIdentity(t *testing.T) {
	for _, mode := range []string{"valid", "split-commit", "wrong-repository", "missing-commit", "local-with-git"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			m := selectionProject(t, root, []string{"alpha", "zulu"}, map[string]map[string]any{})
			source := manifest.Source{Git: "https://example.org/kit", Identity: "example.org/kit", Ref: manifest.Ref{Kind: "tag", Value: "v1"}}
			if mode != "local-with-git" {
				m.Sources["s"] = source
			}
			acquire := func(s manifest.Selection) (*Node, error) {
				n, _ := fixtureAcquisition(s)
				n.Repo = root
				n.Identity = source.Identity
				n.Resolved.Kind = "tag"
				n.Resolved.Ref = "v1"
				n.Resolved.Commit = strings.Repeat("a", 40)
				if mode == "split-commit" && s.Decl.Name == "zulu" {
					n.Resolved.Commit = strings.Repeat("b", 40)
				}
				if mode == "wrong-repository" {
					n.Identity = "example.org/other"
				}
				if mode == "missing-commit" {
					n.Resolved.Commit = ""
				}
				return n, nil
			}
			nodes, err := BuildExpanded(Options{}, m, manifest.ExpansionOptions{GitRoots: map[string]string{"s": root}}, acquire, nil)
			if mode == "valid" {
				if err != nil || len(nodes) != 2 {
					t.Fatalf("%v %v", nodes, err)
				}
			} else if err == nil || nodes != nil {
				t.Fatalf("%v %v", nodes, err)
			}
		})
	}
}
