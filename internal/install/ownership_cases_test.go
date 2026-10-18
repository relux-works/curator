package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/marker"
)

type ownershipCase struct {
	ID    string `json:"id"`
	Input struct {
		Selected []string                  `json:"selected"`
		Entries  map[string]ownershipEntry `json:"canonical_entries"`
		Recheck  map[string]ownershipEntry `json:"recheck_canonical_entries"`
		Parent   string                    `json:"canonical_parent_type"`
		Versions []int                     `json:"recognized_source_marker_versions"`
	} `json:"input"`
	Expected struct {
		Status  string   `json:"status"`
		Paths   []string `json:"paths"`
		Removed []string `json:"removed"`
	} `json:"expected"`
}
type ownershipEntry struct {
	Type   string `json:"node_type"`
	Marker *struct {
		Type     string          `json:"file_type"`
		Readable bool            `json:"readable"`
		JSON     json.RawMessage `json:"json"`
		Raw      string          `json:"raw"`
	} `json:"marker"`
}

// Every published ownership row drives Project and its serialized target
// stagers. Refusals preserve canonical/lock/adapter/runtime/shim state. This
// reader recognizes v6, so the v6-reader-unaware row uses an unknown carrier
// (77) to exercise that same unsupported-reader refusal; v6 itself is repaired.
func TestInstallDraftOwnershipCases(t *testing.T) {
	payload, err := os.ReadFile("testdata/ownership-r2/marker-ownership-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []ownershipCase
	if err := json.Unmarshal(payload, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 21 {
		t.Fatalf("ownership corpus changed: %d", len(cases))
	}
	for _, c := range cases {
		t.Run(c.ID, func(t *testing.T) {
			selected := append([]string{"control"}, c.Input.Selected...)
			skills := make([]map[string]string, 0, len(selected))
			for _, name := range selected {
				skills = append(skills, map[string]string{"name": name, "from": "s", "directory": "skills/" + name})
			}
			manifest, err := json.Marshal(map[string]any{"schema_version": 2, "sources": map[string]any{"s": map[string]string{"path": "."}}, "skills": skills})
			if err != nil {
				t.Fatal(err)
			}
			project, home := draftSweepProject(t, string(manifest))
			for _, name := range selected {
				writeDraftScriptSkill(t, filepath.Join(project, "skills", name), name, name+"-tool", "#!/bin/sh\necho ok\n", nil)
			}
			refreshDraftLocked(t, project, home, string(manifest), nil)
			if r := draftSweepInstall(t, home, project, Options{}); r.Status != "ok" {
				t.Fatalf("control: %+v", r)
			}
			store := filepath.Join(project, ".agents", "skills")
			for name, entry := range c.Input.Entries {
				live := filepath.Join(store, name)
				if err := os.RemoveAll(live); err != nil {
					t.Fatal(err)
				}
				switch entry.Type {
				case "fifo":
					ownershipSpecial(t, live)
				default:
					mustOwnershipWrite(t, filepath.Join(live, "notes.md"), "preceding bytes\n")
					if entry.Marker != nil {
						m := entry.Marker
						path := filepath.Join(live, marker.Name)
						raw := m.JSON
						if m.Raw != "" {
							raw = []byte(m.Raw)
						}
						if c.ID == "v6-reader-unaware" {
							raw = []byte(strings.Replace(string(raw), `"schema_version": 6`, `"schema_version": 77`, 1))
						}
						switch m.Type {
						case "symlink":
							target := filepath.Join(project, "foreign-marker")
							mustOwnershipWrite(t, target, string(raw))
							if err := os.Symlink(target, path); err != nil {
								t.Fatal(err)
							}
						case "fifo":
							ownershipSpecial(t, path)
						default:
							mustOwnershipWrite(t, path, string(raw))
							if !m.Readable {
								// Windows mode bits do not revoke reads. A directory is an actual
								// unusable marker there; Unix additionally tests regular EACCES.
								if runtime.GOOS == "windows" {
									if err := os.Remove(path); err != nil {
										t.Fatal(err)
									}
									if err := os.Mkdir(path, 0o755); err != nil {
										t.Fatal(err)
									}
								} else {
									if err := os.Chmod(path, 0); err != nil {
										t.Fatal(err)
									}
									t.Cleanup(func() { _ = os.Chmod(path, 0o600) })
								}
							}
						}
					}
					if entry.Type == "symlink" {
						target := filepath.Join(project, "foreign-node")
						if err := os.Rename(live, target); err != nil {
							t.Fatal(err)
						}
						if err := os.Symlink(target, live); err != nil {
							t.Fatal(err)
						}
					}
				}
			}
			if c.Input.Parent == "symlink" {
				target := filepath.Join(project, "foreign-store")
				if err := os.Rename(store, target); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, store); err != nil {
					t.Fatal(err)
				}
			}
			foreign := map[string]string{}
			for _, name := range []string{"foreign-marker", "foreign-node", "foreign-store"} {
				foreign[name] = draftEntryDigest(filepath.Join(project, name))
			}
			before := snapshotDraftState(t, home, project)
			opts := Options{}
			if len(c.Input.Recheck) != 0 {
				changed := false
				opts.OnStaged = func(Staged) error {
					if changed {
						return nil
					}
					changed = true
					for name := range c.Input.Recheck {
						if err := os.Remove(filepath.Join(store, name, marker.Name)); err != nil {
							return err
						}
					}
					before = snapshotDraftState(t, home, project)
					return nil
				}
			}
			r := draftSweepInstall(t, home, project, opts)
			if c.Expected.Status == "refused" {
				diagnostic := strings.Join(r.Errors, ";")
				if r.Status != "failed" || !strings.Contains(diagnostic, marker.DiagOwnershipUnproven) {
					t.Fatalf("want ownership refusal: %+v", r)
				}
				for _, path := range c.Expected.Paths {
					if !strings.Contains(diagnostic, filepath.Join(project, filepath.FromSlash(path))) {
						t.Errorf("unnamed path %s: %s", path, diagnostic)
					}
				}
				assertOwnershipUnchanged(t, before, home, project)
				for name, digest := range foreign {
					if got := draftEntryDigest(filepath.Join(project, name)); got != digest {
						t.Fatalf("foreign target %s changed on refusal", name)
					}
				}
			} else {
				if r.Status != "ok" {
					t.Fatalf("owned/absent control: %+v", r)
				}
				for _, path := range c.Expected.Removed {
					if _, err := os.Lstat(filepath.Join(project, filepath.FromSlash(path))); !os.IsNotExist(err) {
						t.Fatalf("owned stale retained: %v", err)
					}
				}
				for _, name := range selected {
					if marker.Read(filepath.Join(store, name)) == nil {
						t.Fatalf("selected %s was not rederived", name)
					}
				}
			}
		})
	}
}
