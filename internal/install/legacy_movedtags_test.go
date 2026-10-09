package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLegacyLaneMovedTagRefusesStrictSecondInstall is the F4 row at the
// real install entry: a schema-9 skill installed under a schema-1
// Skillfile records a package marker with no frozen lock, and legacy
// resolution still resolves the live tag on each install — so after the
// tag moves, a strict reinstall refuses with the moved-tag diagnostic
// (manager §2.1 keeps moved-tag evaluation mandatory), while a default
// reinstall warns and rebinds the recorded package commit.
func TestLegacyLaneMovedTagRefusesStrictSecondInstall(t *testing.T) {
	t.Parallel()
	e := newEnv(t)
	first := setupLegacyScriptRepo(e, "consumer", "ctool", "#!/bin/sh\necho consumer-ok\n")
	e.declare("consumer")

	if result := e.install(Options{}); result.Status != "ok" {
		t.Fatalf("install: %+v", result)
	}

	repo := filepath.Join(e.skillsRoot, "consumer")
	e.write(repo, "references/moved.md", "moved\n")
	e.git(repo, "add", ".")
	e.git(repo, "commit", "-qm", "move tag")
	e.git(repo, "tag", "-f", "v1")
	moved := e.gitOutput(repo, "rev-parse", "HEAD")
	if moved == first {
		t.Fatal("tag did not move")
	}

	strict := e.install(Options{StrictTags: true})
	if strict.Status != "failed" ||
		!strings.Contains(strings.Join(strict.Errors, "\n"), "moved tag for consumer: v1 "+first+" -> "+moved) {
		t.Fatalf("strict reinstall = %+v, want the moved-tag refusal", strict)
	}
	if _, err := os.Stat(filepath.Join(e.project, ".agents", "skills", "consumer")); err != nil {
		t.Fatalf("refused reinstall disturbed the installation: %v", err)
	}

	again := e.install(Options{})
	if again.Status != "ok" {
		t.Fatalf("reinstall: %+v", again)
	}
	if joined := strings.Join(again.Messages, "\n"); !strings.Contains(joined, "moved tag for consumer: v1 ") {
		t.Fatalf("default reinstall warned nothing about the moved tag:\n%s", joined)
	}
	recorded, _ := readLegacyMarker(t, filepath.Join(e.project, ".agents", "skills", "consumer"))
	if recorded.Package == nil || recorded.Package.Commit == nil || recorded.Package.Commit.Hex != moved {
		t.Fatalf("reinstalled marker binds %+v, want the moved commit %s", recorded.Package, moved)
	}
}

// TestReviewChangedDeclaredTagIsNotMovedTag is the F4 regression at the
// real install entries: a schema-9 skill installed under a schema-1
// Skillfile records a package marker with no prior ref, and legacy
// resolution re-resolves the live tag on each install — so after an
// explicit declaration change to another unchanged tag, a strict
// reinstall must succeed silently instead of refusing with the moved-tag
// diagnostic (core §10 limits strict-tag refusal to a moved tag). The
// narrowing mutant `movedtag-commit-only` drops the prior-declaration comparison
// and warns on any commit difference; it fails this test on both
// entries. The companion moved-tag test above pins that an actual move
// of the same tag still refuses.
func TestReviewChangedDeclaredTagIsNotMovedTag(t *testing.T) {
	t.Parallel()
	for _, global := range []bool{false, true} {
		label := "project"
		if global {
			label = "global"
		}
		t.Run(label, func(t *testing.T) {
			t.Parallel()
			e := newEnv(t)
			setupLegacyScriptRepo(e, "consumer", "ctool", "#!/bin/sh\necho consumer-ok\n")
			root := e.project
			if global {
				e.globalDeclare("consumer")
				root = GlobalRoot(e.home)
			} else {
				e.declare("consumer")
			}
			run := func() Result {
				if global {
					return Global(e.cfg, t.TempDir(), Options{Platform: installPlatform(), StrictTags: true})
				}
				return e.install(Options{StrictTags: true})
			}
			if first := run(); first.Status != "ok" {
				t.Fatalf("first install: %+v", first)
			}
			repo := filepath.Join(e.skillsRoot, "consumer")
			e.write(repo, "references/new.md", "new release\n")
			e.git(repo, "add", ".")
			e.git(repo, "commit", "-qm", "release two")
			e.git(repo, "tag", "v2")
			path := filepath.Join(root, "Skillfile.json")
			payload, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var doc map[string]any
			if err := json.Unmarshal(payload, &doc); err != nil {
				t.Fatal(err)
			}
			doc["skills"].([]any)[0].(map[string]any)["tag"] = "v2"
			payload, err = json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, payload, 0o600); err != nil {
				t.Fatal(err)
			}
			second := run()
			if second.Status != "ok" {
				t.Fatalf("unchanged v1 and v2 tags; explicit v1->v2 declaration must install under strict-tag policy: %+v", second)
			}
			if strings.Contains(strings.Join(second.Messages, "\n"), "moved tag") {
				t.Fatalf("unchanged distinct tag reported as moved: %+v", second)
			}
		})
	}
}

// Real Project/Global entries: keeping an unrelated tag on the old commit
// does not make movement of the installed tag an explicit declaration change.
func TestLegacyLaneSameTagMoveWithAlias(t *testing.T) {
	for _, global := range []bool{false, true} {
		for _, alias := range []string{"none", "lightweight", "annotated"} {
			scope := "project"
			if global {
				scope = "global"
			}
			t.Run(scope+"/"+alias, func(t *testing.T) {
				e := newEnv(t)
				first := setupLegacyScriptRepo(e, "consumer", "ctool", "#!/bin/sh\necho consumer-ok\n")
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
				installed := filepath.Join(root, ".agents", "skills", "consumer")
				if global {
					installed = filepath.Join(root, "skills", "consumer")
				}
				_, before := readLegacyMarker(t, installed)
				repo := filepath.Join(e.skillsRoot, "consumer")
				if alias == "lightweight" {
					e.git(repo, "tag", "archive", first)
				}
				if alias == "annotated" {
					e.git(repo, "tag", "-a", "archive", "-m", "old release", first)
				}
				e.write(repo, "references/moved.md", "moved\n")
				e.git(repo, "add", ".")
				e.git(repo, "commit", "-qm", "move v1")
				e.git(repo, "tag", "-f", "v1")
				moved := e.gitOutput(repo, "rev-parse", "HEAD")
				result := run()
				if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, "\n"), "moved tag for consumer: v1 "+first+" -> "+moved) {
					t.Fatalf("same declared v1 moved; unrelated %s tag on prior commit cannot waive strict policy: %+v", alias, result)
				}
				_, after := readLegacyMarker(t, installed)
				beforeBytes, _ := json.Marshal(before)
				afterBytes, _ := json.Marshal(after)
				if string(beforeBytes) != string(afterBytes) {
					t.Fatal("strict refusal changed marker")
				}
			})
		}
	}
}

// Deleting an unrelated, no-longer-declared tag does not move the new tag.
func TestLegacyLaneChangedTagAfterOldTagDeleted(t *testing.T) {
	for _, global := range []bool{false, true} {
		scope := "project"
		if global {
			scope = "global"
		}
		t.Run(scope, func(t *testing.T) {
			e := newEnv(t)
			setupLegacyScriptRepo(e, "consumer", "ctool", "#!/bin/sh\necho consumer-ok\n")
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
			repo := filepath.Join(e.skillsRoot, "consumer")
			e.write(repo, "references/release.md", "release\n")
			e.git(repo, "add", ".")
			e.git(repo, "commit", "-qm", "v2 release")
			e.git(repo, "tag", "v2")
			e.git(repo, "tag", "-d", "v1")
			path := filepath.Join(root, "Skillfile.json")
			payload, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var doc map[string]any
			if err := json.Unmarshal(payload, &doc); err != nil {
				t.Fatal(err)
			}
			doc["skills"].([]any)[0].(map[string]any)["tag"] = "v2"
			payload, err = json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, payload, 0600); err != nil {
				t.Fatal(err)
			}
			result := run()
			if result.Status != "ok" || strings.Contains(strings.Join(result.Messages, "\n"), "moved tag") {
				t.Fatalf("explicit declaration change to unmoved v2 must install: %+v", result)
			}
		})
	}
}
