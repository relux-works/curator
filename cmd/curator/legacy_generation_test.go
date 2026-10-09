package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
)

func TestLegacyPackageStatusBindingAndRepair(t *testing.T) {
	for _, global := range []bool{false, true} {
		scope := "project"
		if global {
			scope = "global"
		}
		t.Run(scope, func(t *testing.T) {
			home := globalScopeDeclaring(t, `{"name":"consumer","tag":"v1"}`)
			root := install.GlobalRoot(home)
			args := []string{"global"}
			if !global {
				var project string
				_, home, project, _, _ = auditDirectoryProject(t)
				root = project
				args = nil
			} else {
				repo := filepath.Join(filepath.Dir(home), "skills", "consumer")
				writeAuditDirectorySkill(t, repo, "consumer", 9, nil)
				runGit(t, repo, "init", "-q", "-b", "main")
				runGit(t, repo, "add", ".")
				runGit(t, repo, "commit", "-qm", "schema9")
				runGit(t, repo, "tag", "v1")
			}
			configPath := filepath.Join(home, "config.json")
			command := func(verb string, tail ...string) (int, string, string) {
				a := append([]string{}, args...)
				a = append(a, verb)
				if !global {
					a = append(a, "app")
				}
				a = append(a, tail...)
				return capture(t, configPath, a...)
			}
			if code, stdout, stderr := command("install"); code != exitOK {
				t.Fatalf("install: %d %s %s", code, stdout, stderr)
			}
			if code, stdout, stderr := command("status", "--check"); code != exitOK {
				t.Fatalf("initial current: %d %s %s", code, stdout, stderr)
			}
			installed := filepath.Join(root, ".agents", "skills", "consumer", marker.Name)
			if global {
				installed = filepath.Join(root, "skills", "consumer", marker.Name)
			}
			before, err := os.ReadFile(installed)
			if err != nil {
				t.Fatal(err)
			}
			repo := filepath.Join(filepath.Dir(home), "skills", "consumer")
			runGit(t, repo, "tag", "v2")
			// Retain root selection and change only the declared tag.
			payload, err := os.ReadFile(filepath.Join(root, manifest.Name))
			if err != nil {
				t.Fatal(err)
			}
			changed := string(payload)
			for i := 0; i+4 <= len(changed); i++ {
				if changed[i:i+4] == `"v1"` {
					changed = changed[:i] + `"v2"` + changed[i+4:]
					break
				}
			}
			writeFile(t, filepath.Join(root, manifest.Name), changed)
			code, stdout, stderr := command("status", "--json")
			if code != exitOK {
				t.Fatalf("status JSON: %d %s %s", code, stdout, stderr)
			}
			var state string
			if global {
				state = decodeGlobalStatus(t, stdout).Skills["consumer"]
			} else {
				state = decodeStatus(t, stdout).Skills["consumer"]
			}
			if state != stateNeedsInstall {
				t.Errorf("same-commit changed declaration: %s; want needs-install", stdout)
			}
			if code, stdout, stderr := command("status", "--check"); code == exitOK {
				t.Errorf("stale declared ref admitted: %s %s", stdout, stderr)
			}
			after, err := os.ReadFile(installed)
			if err != nil {
				t.Fatal(err)
			}
			if string(before) != string(after) {
				t.Error("status mutated installed marker")
			}
			if code, stdout, stderr := command("install"); code != exitOK {
				t.Fatalf("rebind install: %d %s %s", code, stdout, stderr)
			}
			if code, stdout, stderr := command("status", "--check"); code != exitOK {
				t.Errorf("repaired binding not current: %d %s %s", code, stdout, stderr)
			}
		})
	}
}
