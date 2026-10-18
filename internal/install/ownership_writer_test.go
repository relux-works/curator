package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/sourcelock"
)

// Published writer acceptance rows exercise Project on both Skillfile lanes;
// reinstallation keeps the existing marker bytes. Invalid/unreadable present
// manifests are not absence. Forbidden manifestless declarations exercise the
// production marker.Write boundary, not package admission or legacy fallback.
func TestInstallDraftMarkerWriterCases(t *testing.T) {
	data, err := os.ReadFile("testdata/ownership-r2/marker-writer-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		ID    string `json:"id"`
		Input struct {
			Skillfile int `json:"skillfile_schema_version"`
			Manifest  struct {
				State   string `json:"state"`
				Version int    `json:"schema_version"`
			} `json:"manifest"`
			Declarations []string `json:"package_declarations"`
		} `json:"input"`
		Expected struct {
			Status string `json:"status"`
			Schema int    `json:"marker_schema_version"`
			Skill  *int   `json:"marker_skill_schema_version"`
		} `json:"expected"`
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 10 {
		t.Fatalf("writer corpus changed: %d", len(cases))
	}
	for _, c := range cases {
		t.Run(c.ID, func(t *testing.T) {
			if len(c.Input.Declarations) != 0 {
				dir := t.TempDir()
				path := filepath.Join(dir, marker.Name)
				mustOwnershipWrite(t, path, "preceding marker bytes")
				m := &marker.Marker{Manifestless: true, Name: "review", Package: &marker.Package{Kind: "local-snapshot", Snapshot: "sha256:" + strings.Repeat("a", 64)}, LockSHA256: "sha256:" + strings.Repeat("b", 64), ContentSHA256: "sha256:" + strings.Repeat("c", 64), InstalledAt: "2000-01-01T00:00:00Z"}
				switch c.Input.Declarations[0] {
				case "command":
					m.Commands = []string{"tool"}
				case "runtime":
					m.RuntimeRoots = []string{"scripts"}
				case "build":
					m.BuildRoots = []string{"build"}
				case "legacy-agents-runtime":
					m.Commands = []string{"legacy-tool"}
					m.RuntimeRoots = []string{"agents"}
				default:
					t.Fatal("unmapped declaration")
				}
				if err := marker.Write(dir, m); err == nil {
					t.Fatal("manifestless declaration admitted")
				}
				got, err := os.ReadFile(path)
				if err != nil || string(got) != "preceding marker bytes" {
					t.Fatalf("refusal wrote marker: %s %v", got, err)
				}
				return
			}
			schema := c.Input.Skillfile
			if schema == 0 {
				schema = 2
			}
			var project, home, source string
			var run func() Result
			if schema == 1 {
				e := newEnv(t)
				source = filepath.Join(e.skillsRoot, "review")
				mustOwnershipWrite(t, filepath.Join(source, "SKILL.md"), "---\nname: review\ndescription: context\n---\n# Review\n")
				e.git(source, "init", "-q", "-b", "main")
				if c.Input.Manifest.State == "valid" {
					mustOwnershipWrite(t, filepath.Join(source, "agent-skill.json"), writerManifest(c.Input.Manifest.Version))
				}
				e.git(source, "add", ".")
				e.git(source, "commit", "-qm", "fixture")
				e.git(source, "tag", "v1")
				e.declare("review")
				project, home = e.project, e.home
				run = func() Result { return e.install(Options{}) }
			} else {
				payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
				project, home = draftSweepProject(t, payload)
				source = filepath.Join(project, "skills", "review")
				mustOwnershipWrite(t, filepath.Join(source, "SKILL.md"), "---\nname: review\ndescription: context\n---\n# Review\n")
				if c.Input.Manifest.State == "valid" {
					mustOwnershipWrite(t, filepath.Join(source, "agent-skill.json"), writerManifest(c.Input.Manifest.Version))
				}
				if c.Input.Manifest.State == "invalid" {
					mustOwnershipWrite(t, filepath.Join(source, "agent-skill.json"), "{")
				}
				if c.Input.Manifest.State == "unreadable" {
					if err := os.Mkdir(filepath.Join(source, "agent-skill.json"), 0o755); err != nil {
						t.Fatal(err)
					}
				}
				// Resolve is part of the real admission entry. Invalid present manifests
				// may refuse here rather than waiting until target publication.
				if c.Expected.Status == "refused" {
					m, err := manifest.ParseBytes([]byte(payload), filepath.Join(project, "Skillfile.json"))
					if err != nil {
						t.Fatal(err)
					}
					before := snapshotDraftState(t, home, project)
					_, err = closure.RefreshDraft(closure.DraftResolveConfig{ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: []byte(payload)}, sourcelock.PathIn(project), DraftBindingsPath(home, project))
					if err == nil || !strings.Contains(err.Error(), "agent-skill.json") {
						t.Fatalf("want named present-manifest admission refusal, got %v", err)
					}
					assertOwnershipUnchanged(t, before, home, project)
					return
				}
				refreshDraftLocked(t, project, home, payload, nil)
				run = func() Result { return draftSweepInstall(t, home, project, Options{}) }
			}
			if r := run(); r.Status != "ok" {
				t.Fatalf("install: %+v", r)
			}
			path := filepath.Join(project, ".agents", "skills", "review", marker.Name)
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var got map[string]json.RawMessage
			if err := json.Unmarshal(before, &got); err != nil {
				t.Fatal(err)
			}
			var version int
			if err := json.Unmarshal(got["schema_version"], &version); err != nil || version != c.Expected.Schema {
				t.Fatalf("schema=%d want %d: %v", version, c.Expected.Schema, err)
			}
			if c.Expected.Skill == nil {
				if string(got["skill_schema_version"]) != "null" {
					t.Fatalf("absence fabricated: %s", got["skill_schema_version"])
				}
			} else {
				var skill int
				if err := json.Unmarshal(got["skill_schema_version"], &skill); err != nil || skill != *c.Expected.Skill {
					t.Fatalf("skill=%d: %v", skill, err)
				}
			}
			if r := run(); r.Status != "ok" {
				t.Fatalf("reinstall: %+v", r)
			}
			after, err := os.ReadFile(path)
			if err != nil || string(after) != string(before) {
				t.Fatalf("reinstall rewrote marker: %v", err)
			}
		})
	}
}

func writerManifest(version int) string {
	data, _ := json.Marshal(map[string]any{"schema_version": version, "capabilities": map[string]any{}, "commands": map[string]any{}, "dependencies": map[string]any{"skills": map[string]any{}}})
	return string(data)
}
