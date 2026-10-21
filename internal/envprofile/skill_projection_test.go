package envprofile

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/whitelist"
)

func projectionFiles(t *testing.T, root string) map[string][]byte {
	t.Helper()
	files := map[string][]byte{}
	resolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	root = resolved
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(rel)] = body
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

// Every fixed exclusion is absent at nested depths; visible neighboring files
// survive. Planning writes nothing, both hash framings match CopyContext, and
// publication is shared and byte-identical through links and tree-copy fallback.
func TestSkillProjectionCoreSelection(t *testing.T) {
	for _, version := range []hashing.Version{hashing.VersionV1, hashing.VersionV2} {
		t.Run(fmt.Sprint(version), func(t *testing.T) {
			source, home := t.TempDir(), t.TempDir()
			writeGitFile(t, source, "SKILL.md", "# Guide\n")
			writeGitFile(t, source, "references/guide.md", "visible\n")
			for _, pattern := range whitelist.AlwaysExcluded {
				name := strings.ReplaceAll(pattern, "*", "fixture")
				writeGitFile(t, source, "references/nested/"+name+"/notes.md", "excluded\n")
			}
			writeGitFile(t, source, ".research/notes.md", "excluded\n")
			skill, err := projectSkill(home, source, version)
			if err != nil {
				t.Fatal(err)
			}
			if entries, err := os.ReadDir(home); err != nil || len(entries) != 0 {
				t.Fatalf("planning wrote state: %v %v", entries, err)
			}
			if err := publishSkillProjection(home, skill); err != nil {
				t.Fatal(err)
			}
			want := map[string][]byte{"SKILL.md": []byte("# Guide\n"), "references/guide.md": []byte("visible\n")}
			if got := projectionFiles(t, skill.root); !reflect.DeepEqual(got, want) {
				t.Fatalf("projection published excluded files: %v", got)
			}
			canonical := filepath.Join(t.TempDir(), "canonical")
			if _, err := whitelist.CopyContext(source, canonical, true, nil); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(projectionFiles(t, canonical), want) {
				t.Fatal("canonical copier differs")
			}
			if err := publishSkillProjection(home, skill); err != nil {
				t.Fatal(err)
			}
			for _, id := range []string{"claude_code", "codex_cli"} {
				t.Run(id, func(t *testing.T) {
					target := t.TempDir()
					rel := "skills/guide"
					copies := []envmarker.Copy{}
					plan := &homePlan{marker: &envmarker.Marker{Surfaces: map[string]envmarker.Surface{envmarker.SurfaceSkills: {Paths: []string{rel}, Copies: &copies}}}}
					if err := copyLinkFallback(target, rel, skill.root, rel, plan); err != nil {
						t.Fatal(err)
					}
					if got := projectionFiles(t, filepath.Join(target, rel)); !reflect.DeepEqual(got, want) {
						t.Fatalf("fallback published excluded files: %v", got)
					}
				})
			}
		})
	}
}

// A repository-root package with board/research/vendor inputs installs, while
// all native and managed skill surfaces expose only selected context. Exported
// scripts and declared runtime/build roots stay out; command-less scripts and
// declared visible Markdown remain readable. Both fallback writers copy exactly
// the same tree as the production links for Claude and Codex.
func TestProfileSkillProjectionPublication(t *testing.T) {
	for _, commands := range []bool{false, true} {
		t.Run(fmt.Sprintf("commands-%t", commands), func(t *testing.T) {
			home := t.TempDir()
			pinOperatorHome(t, home)
			native := pinHomes(t)
			ids := newGitIdentities(t)
			manifest := `{"schema_version":3,"capabilities":{"network":["api.example.net"]}}`
			if commands {
				manifest = `{"schema_version":6,"capabilities":{"network":["api.example.net"]},"runtime_roots":["assets/runtime"],"build_roots":["tools/board-cli","assets/build"],"commands":{"scope-run":{"type":"script","unix_path":"assets/runtime/run.sh"},"scope-build":{"type":"build","driver":"go-v1","source_dir":"tools/board-cli/cmd/tool"},"scope-extra":{"type":"build","driver":"go-v1","source_dir":"assets/build/cmd/tool"}}}`
			}
			files := map[string]string{
				"SKILL.md":         "---\nname: scoped-guide\ndescription: Context projection fixture\n---\n# Guide\n",
				"agent-skill.json": manifest, "references/guide.md": "Read https://api.example.net/guide\n",
				"assets/visible.txt": "visible asset\n", "scripts/guide.md": "script instructions\n", "scripts/run.sh": "#!/bin/sh\nprintf 'runtime control\\n'\n",
				"assets/runtime/run.sh":              "#!/bin/sh\nprintf 'runtime control\\n'\n",
				"assets/runtime/runtime.md":          "runtime instructions\n",
				".research/notes.md":                 "Read https://history.example.net/research\n",
				".task-board/.resources/brief.md":    "Read https://history.example.net/board\n",
				"tools/board-cli/vendor/m/README.md": "Read https://history.example.net/vendor\n",
				"tools/board-cli/go.mod":             "module example.test/board\ngo 1.25.5\n",
				"tools/board-cli/cmd/tool/main.go":   "package main\nfunc main() {}\n",
				"references/README.md":               "Read https://history.example.net/excluded\n", "assets/tests/a.md": "Read https://history.example.net/excluded\n",
			}
			if commands {
				files["assets/build/go.mod"] = "module example.test/build\ngo 1.25.5\n"
				files["assets/build/cmd/tool/main.go"] = "package main\nfunc main() {}\n"
				files["assets/build/guide.md"] = "Read https://history.example.net/build\n"
			}
			repo := gitRepo(t, files, "v1.0.0")
			operand := ids.serve(repo, "https://example.com/skills/scoped-guide")
			profile := t.TempDir()
			writeManifestPackage(t, profile, `{"schema_version":1,"name":"scope-profile","version":"1.0.0","requires":{"skills":{"scoped-guide":{"git":"`+operand+`","range":"*"}}}}`, nil)
			info, _, _, err := Install(home, InstallOptions{Operand: profile})
			if err != nil {
				t.Fatal(err)
			}
			member, ok := info.Lock.Find(contextlock.KindSkill, "scoped-guide")
			if !ok {
				t.Fatal("missing skill")
			}
			raw := packageRoot(newGitManager(home).entryPath(home, resolvedOf(member)), member.Directory)
			want := map[string][]byte{"SKILL.md": []byte(files["SKILL.md"]), "references/guide.md": []byte(files["references/guide.md"]), "assets/visible.txt": []byte(files["assets/visible.txt"])}
			if !commands {
				for _, p := range []string{"scripts/guide.md", "scripts/run.sh", "assets/runtime/runtime.md", "assets/runtime/run.sh"} {
					want[p] = []byte(files[p])
				}
			}
			var shared string
			for _, id := range []string{"claude_code", "codex_cli"} {
				t.Run(id, func(t *testing.T) {
					req := ResolveRequest{Home: home, Profile: info.Name, EnvID: id, Repair: true, LaunchDir: t.TempDir(), Machine: envregistry.DefaultMachineConfig(), Detect: func(envregistry.Adapter) string { return "unknown" }, NativeHomeOf: func(id string) (string, error) { return native[id], nil }}
					if _, err := Resolve(req); err != nil {
						t.Fatal(err)
					}
					for _, base := range []string{native[id], ManagedHomeDir(home, info.Name, id)} {
						rel := "skills/scoped-guide"
						live := filepath.Join(base, rel)
						if got := projectionFiles(t, live); !reflect.DeepEqual(got, want) {
							t.Fatalf("skill publication differs from projection: %v", got)
						}
						target, err := os.Readlink(live)
						if err != nil {
							t.Fatal(err)
						}
						if target == raw {
							t.Fatal("raw package root published")
						}
						if shared == "" {
							shared = target
						} else if shared != target {
							t.Fatal("projection store not shared")
						}
						copyHome := t.TempDir()
						if err := copyTree(target, copyHome, rel); err != nil {
							t.Fatal(err)
						}
						if !reflect.DeepEqual(projectionFiles(t, filepath.Join(copyHome, rel)), want) {
							t.Fatal("native fallback differs")
						}
						copies := []envmarker.Copy{}
						plan := &homePlan{marker: &envmarker.Marker{Surfaces: map[string]envmarker.Surface{envmarker.SurfaceSkills: {Paths: []string{rel}, Copies: &copies}}}}
						other := t.TempDir()
						if err := copyLinkFallback(other, rel, target, rel, plan); err != nil {
							t.Fatal(err)
						}
						if !reflect.DeepEqual(projectionFiles(t, filepath.Join(other, rel)), want) {
							t.Fatal("managed fallback differs")
						}
						// Exercise verification of real installed fallback copies too.
						if err := os.Remove(live); err != nil {
							t.Fatal(err)
						}
						if err := copyTree(target, base, rel); err != nil {
							t.Fatal(err)
						}
						marker, err := envmarker.Read(base)
						if err != nil {
							t.Fatal(err)
						}
						surface := marker.Surfaces[envmarker.SurfaceSkills]
						fallback := append([]envmarker.Copy{}, *surface.Copies...)
						fallback = append(fallback, envmarker.Copy{Path: rel, Reason: envmarker.ReasonSymlinkFallback})
						surface.Copies = &fallback
						marker.Surfaces[envmarker.SurfaceSkills] = surface
						body, err := marker.Marshal()
						if err != nil {
							t.Fatal(err)
						}
						if err := os.WriteFile(filepath.Join(base, envmarker.Name), body, 0600); err != nil {
							t.Fatal(err)
						}
					}
					req.Repair = false
					if _, err := Resolve(req); err != nil {
						t.Fatalf("projected fallback not current: %v", err)
					}
					row, err := nativeSkillsState(home, info.Name, id, req.NativeHomeOf)
					if err != nil || row == nil || row.State != "current" {
						t.Fatalf("native fallback not current: %+v %v", row, err)
					}
				})
			}
			for _, p := range []string{".research/notes.md", "tools/board-cli/vendor/m/README.md"} {
				if _, err := os.ReadFile(filepath.Join(raw, p)); err != nil {
					t.Fatal("raw snapshot changed", err)
				}
			}
		})
	}
}

// Old Curator-owned raw-root links and old skills surface hashes are stale in
// both native status and managed verification. Repair replaces them without
// takeover, preserves the lock, and converges to readable projected context.
func TestProfileSkillProjectionRepair(t *testing.T) {
	for _, rawLink := range []bool{false, true} {
		for _, id := range []string{"claude_code", "codex_cli"} {
			t.Run(fmt.Sprintf("%s/raw-link-%t", id, rawLink), func(t *testing.T) {
				pinHomes(t)
				fx := writeManagedFixture(t, "sample")
				req := fx.request(id)
				req.Repair = true
				if _, err := Resolve(req); err != nil {
					t.Fatal(err)
				}
				if _, err := materializeScopeWithNativeHome(fx.home, fx.profile, id, Policy{}, req.NativeHomeOf); err != nil {
					t.Fatal(err)
				}
				lock, _, err := readLock(fx.home, fx.profile)
				if err != nil {
					t.Fatal(err)
				}
				member, _ := lock.Find(contextlock.KindSkill, "myskill")
				raw := newGitManager(fx.home).entryPath(fx.home, resolvedOf(member))
				oldHash, err := hashing.ContentSHA256WithVersion(raw, map[string]bool{}, lock.ContentHashVersion())
				if err != nil {
					t.Fatal(err)
				}
				before, err := os.ReadFile(lockPath(fx.home, fx.profile))
				if err != nil {
					t.Fatal(err)
				}
				rel := "skills/myskill"
				for _, base := range []string{fx.native[id], ManagedHomeDir(fx.home, fx.profile, id)} {
					live := filepath.Join(base, rel)
					if rawLink {
						if err := os.Remove(live); err != nil {
							t.Fatal(err)
						}
						if err := os.Symlink(raw, live); err != nil {
							t.Fatal(err)
						}
					}
					marker, err := envmarker.Read(base)
					if err != nil {
						t.Fatal(err)
					}
					surface := marker.Surfaces[envmarker.SurfaceSkills]
					surface.ContentSHA256, _ = contextmaterialize.SurfaceHashWithVersion(map[string][]byte{rel: []byte(oldHash)}, lock.ContentHashVersion())
					marker.Surfaces[envmarker.SurfaceSkills] = surface
					body, err := marker.Marshal()
					if err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(base, envmarker.Name), body, 0600); err != nil {
						t.Fatal(err)
					}
				}
				req.Repair = false
				row, err := nativeSkillsState(fx.home, fx.profile, id, req.NativeHomeOf)
				if err != nil || row == nil || row.State != DiagSurfaceDrift {
					t.Fatalf("old native surface current: %+v %v", row, err)
				}
				if _, err := Resolve(req); err == nil || !strings.Contains(err.Error(), DiagHomeStale) {
					t.Fatalf("old managed surface current: %v", err)
				}
				req.Repair = true
				if _, err := Resolve(req); err != nil {
					t.Fatalf("repair required takeover: %v", err)
				}
				req.Repair = false
				if _, err := Resolve(req); err != nil {
					t.Fatal(err)
				}
				row, err = nativeSkillsState(fx.home, fx.profile, id, req.NativeHomeOf)
				if err != nil || row == nil || row.State != "current" {
					t.Fatalf("native repair did not converge: %+v %v", row, err)
				}
				after, err := os.ReadFile(lockPath(fx.home, fx.profile))
				if err != nil || !bytes.Equal(before, after) {
					t.Fatal("repair changed lock", err)
				}
			})
		}
	}
}

// Reuse accepts an intact projection and refuses altered bytes, extra links,
// and a linked store destination without changing either the stored tree or
// the link target. A missing projection is rebuilt from its pinned source.
func TestSkillProjectionStoreReuse(t *testing.T) {
	for _, kind := range []string{"bytes", "link", "destination", "source-changed"} {
		t.Run(kind, func(t *testing.T) {
			home, source := t.TempDir(), t.TempDir()
			writeGitFile(t, source, "SKILL.md", "# Guide\n")
			skill, err := projectSkill(home, source, hashing.VersionV2)
			if err != nil {
				t.Fatal(err)
			}
			if kind == "source-changed" {
				writeGitFile(t, source, "SKILL.md", "# Changed\n")
				if err := publishSkillProjection(home, skill); err == nil || !strings.Contains(err.Error(), "changed while projecting") {
					t.Fatalf("changed source published: %v", err)
				}
				if _, err := os.Lstat(skill.root); !os.IsNotExist(err) {
					t.Fatalf("failed publication left target: %v", err)
				}
				return
			}
			if err := publishSkillProjection(home, skill); err != nil {
				t.Fatal(err)
			}
			if err := verifySkillProjection(home, skill); err != nil {
				t.Fatal(err)
			}
			sentinel := t.TempDir()
			writeGitFile(t, sentinel, "SKILL.md", "untouched\n")
			switch kind {
			case "bytes":
				writeGitFile(t, skill.root, "SKILL.md", "changed\n")
			case "link":
				if err := os.Symlink(filepath.Join(skill.root, "SKILL.md"), filepath.Join(skill.root, "extra")); err != nil {
					t.Fatal(err)
				}
			case "destination":
				if err := os.RemoveAll(skill.root); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(sentinel, skill.root); err != nil {
					t.Fatal(err)
				}
			}
			err = publishSkillProjection(home, skill)
			want := map[string]string{"bytes": "differs from its identity", "link": "link_safety", "destination": "containment"}[kind]
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("altered projection reused or wrong refusal: %v", err)
			}
			body, err := os.ReadFile(filepath.Join(sentinel, "SKILL.md"))
			if err != nil || string(body) != "untouched\n" {
				t.Fatal("refusal changed link target", err)
			}
		})
	}
}

// Missing shared projections make both native and managed status stale without
// recreating state. Repair alone republishes the projection and restores use.
func TestProfileSkillProjectionMissingStore(t *testing.T) {
	fx := writeManagedFixture(t, "sample")
	req := fx.request("codex_cli")
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	if results, err := materializeScopeWithNativeHome(fx.home, fx.profile, req.EnvID, Policy{}, req.NativeHomeOf); err != nil || len(results) != 1 || !results[0].OK {
		t.Fatalf("native materialization: %+v %v", results, err)
	}
	lock, _, err := readLock(fx.home, fx.profile)
	if err != nil {
		t.Fatal(err)
	}
	skills, err := skillsOf(fx.home, newGitManager(fx.home), lock)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(skills[0].root); err != nil {
		t.Fatal(err)
	}
	req.Repair = false
	if _, err := Resolve(req); err == nil || !strings.Contains(err.Error(), DiagHomeStale) {
		t.Fatalf("missing projection current: %v", err)
	}
	row, err := nativeSkillsState(fx.home, fx.profile, req.EnvID, req.NativeHomeOf)
	if err != nil || row == nil || row.State != DiagSurfaceDrift {
		t.Fatalf("missing native projection current: %+v %v", row, err)
	}
	if _, err := os.Lstat(skills[0].root); !os.IsNotExist(err) {
		t.Fatalf("verification wrote projection: %v", err)
	}
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	req.Repair = false
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
}
