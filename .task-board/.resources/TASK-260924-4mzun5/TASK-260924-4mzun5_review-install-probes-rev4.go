package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Real Project/Global entries: keeping an unrelated tag on the old commit
// does not make movement of the installed tag an explicit declaration change.
func TestReviewR4SameTagMoveWithAlias(t *testing.T) {
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
func TestReviewR4ChangedTagAfterOldTagDeleted(t *testing.T) {
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
