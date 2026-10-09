package install

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/sourcelock"
	"github.com/relux-works/curator/internal/stateread"
)

// Neither an absent history nor a self-minted manifest/lock proves that the
// declaration changed. Exercise both production entries, including dry runs.
func TestLegacyLaneRejectsUnboundPreviousGeneration(t *testing.T) {
	for _, global := range []bool{false, true} {
		for _, fault := range []string{"missing", "unreadable", "malformed", "manifest", "lock", "rehashed-lock"} {
			scope := "project"
			if global {
				scope = "global"
			}
			t.Run(scope+"/"+fault, func(t *testing.T) {
				e := newEnv(t)
				first := setupLegacyScriptRepo(e, "consumer", "ctool", "#!/bin/sh\necho consumer-ok\n")
				store := filepath.Join(e.project, ".agents", "skills")
				if global {
					e.globalDeclare("consumer")
					store = filepath.Join(GlobalRoot(e.home), "skills")
				} else {
					e.declare("consumer")
				}
				userHome := t.TempDir()
				run := func(dry bool) Result {
					opts := Options{Platform: installPlatform(), StrictTags: true, DryRun: dry}
					if global {
						return Global(e.cfg, userHome, opts)
					}
					return e.install(opts)
				}
				if result := run(false); result.Status != "ok" {
					t.Fatalf("initial install: %+v", result)
				}
				markerPath := filepath.Join(store, "consumer", marker.Name)
				before, err := os.ReadFile(markerPath)
				if err != nil {
					t.Fatal(err)
				}
				path := legacyGenerationPath(store, "consumer")
				payload, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				switch fault {
				case "missing", "unreadable":
					if err := os.Remove(path); err != nil {
						t.Fatal(err)
					}
					if fault == "unreadable" {
						if err := os.Mkdir(path, 0o700); err != nil {
							t.Fatal(err)
						}
					}
				case "malformed":
					e.write(filepath.Dir(path), filepath.Base(path), "{")
				default:
					var generation legacyGeneration
					if err := json.Unmarshal(payload, &generation); err != nil {
						t.Fatal(err)
					}
					switch fault {
					case "manifest", "rehashed-lock":
						generation.Manifest = bytes.ReplaceAll(generation.Manifest, []byte(`"v1"`), []byte(`"v2"`))
						if fault == "rehashed-lock" {
							lock, err := sourcelock.Parse(generation.Lock)
							if err != nil {
								t.Fatal(err)
							}
							digest, err := sourcelock.ManifestDigest(generation.Manifest)
							if err != nil {
								t.Fatal(err)
							}
							forged, err := sourcelock.New(digest, lock.Members)
							if err != nil {
								t.Fatal(err)
							}
							generation.Lock, err = json.Marshal(forged.Object())
							if err != nil {
								t.Fatal(err)
							}
						}
					case "lock":
						generation.Lock = bytes.ReplaceAll(generation.Lock, []byte(first), []byte(strings.Repeat("f", len(first))))
					}
					payload, err = json.Marshal(generation)
					if err != nil {
						t.Fatal(err)
					}
					e.write(filepath.Dir(path), filepath.Base(path), string(payload))
				}
				repo := filepath.Join(e.skillsRoot, "consumer")
				e.git(repo, "tag", "archive", first)
				e.write(repo, "references/moved.md", "moved\n")
				e.git(repo, "add", ".")
				e.git(repo, "commit", "-qm", "move v1")
				e.git(repo, "tag", "-f", "v1")
				for _, dry := range []bool{true, false} {
					result := run(dry)
					code := stateread.DiagUnreadable
					if fault == "missing" {
						code = stateread.DiagAbsent
					}
					if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, "\n"), code) {
						t.Fatalf("dry=%v: unproven prior declaration must refuse with %s: %+v", dry, code, result)
					}
					after, err := os.ReadFile(markerPath)
					if err != nil || !bytes.Equal(before, after) {
						t.Fatalf("refusal changed marker: %v", err)
					}
				}
			})
		}
	}
}

// Dependency refs come from the old consumer's locked manifest, even when the
// provider's old tag has moved and an unrelated alias still names that commit.
func TestLegacyLaneDependencyTagMovement(t *testing.T) {
	for _, global := range []bool{false, true} {
		for _, changedDeclaration := range []bool{false, true} {
			scope, operation := "project", "same-tag"
			if global {
				scope = "global"
			}
			if changedDeclaration {
				operation = "changed-declaration"
			}
			t.Run(scope+"/"+operation, func(t *testing.T) {
				e := newEnv(t)
				first := setupLegacyDirectoryRepos(e, "skills/backend", "skills/helper")
				provider := filepath.Join(e.skillsRoot, "backend")
				consumer := filepath.Join(e.skillsRoot, "consumer")
				changeRequirements := func(ref string) {
					writeLegacyPackageSkill(e, consumer, "consumer", 9, map[string]any{
						"backend": map[string]any{"git": legacyDirectoryGit, "ref": map[string]any{"kind": "tag", "value": ref}, "directory": "skills/backend"},
						"helper":  map[string]any{"git": legacyDirectoryGit, "ref": map[string]any{"kind": "tag", "value": ref}, "directory": "skills/helper"},
					})
					e.git(consumer, "add", ".")
					e.git(consumer, "commit", "-qm", "declare provider tag")
				}
				changeRequirements("v1")
				e.git(consumer, "tag", "-f", "v1")
				e.git(provider, "tag", "v1")
				root := e.project
				if global {
					e.globalDeclare("consumer")
					root = GlobalRoot(e.home)
				} else {
					e.declare("consumer")
				}
				userHome := t.TempDir()
				run := func() Result {
					if global {
						return Global(e.cfg, userHome, Options{Platform: installPlatform(), StrictTags: true})
					}
					return e.install(Options{StrictTags: true})
				}
				if result := run(); result.Status != "ok" {
					t.Fatalf("initial install: %+v", result)
				}
				e.git(provider, "tag", "archive", first)
				e.write(provider, "skills/backend/references/next.md", "next\n")
				e.git(provider, "add", ".")
				e.git(provider, "commit", "-qm", "next provider")
				if changedDeclaration {
					e.git(provider, "tag", "v2")
					e.git(provider, "tag", "-d", "v1")
					changeRequirements("v2")
					e.git(consumer, "tag", "v2")
					path := filepath.Join(root, "Skillfile.json")
					payload, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}
					e.write(root, "Skillfile.json", string(bytes.ReplaceAll(payload, []byte(`"v1"`), []byte(`"v2"`))))
				} else {
					e.git(provider, "tag", "-f", "v1")
				}
				result := run()
				if changedDeclaration {
					if result.Status != "ok" || strings.Contains(strings.Join(result.Messages, "\n"), "moved tag") {
						t.Fatalf("changed dependency declaration must install: %+v", result)
					}
				} else if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, "\n"), "moved tag for backend: v1 "+first) {
					t.Fatalf("same dependency tag moved despite archive alias: %+v", result)
				}
			})
		}
	}
}
