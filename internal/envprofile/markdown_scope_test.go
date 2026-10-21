package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
)

// A local profile can install a repo-root skill containing research, board
// resources and build vendor docs with historical URLs. The same production
// Install entry rejects an added visible contact before publishing the profile;
// declaring that host admits it. Git identities are served by local fixtures.
func TestProfileInstallMarkdownScope(t *testing.T) {
	for _, tc := range []struct {
		name, visible, capabilities string
		allowed                     bool
	}{
		{"repository-noise", "", `{}`, true},
		{"visible-contact", "Read https://api.example.net/guide\n", `{}`, false},
		{"declared-contact", "Read https://api.example.net/guide\n", `{"network":["api.example.net"]}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			pinHomes(t)
			ids := newGitIdentities(t)
			skill := gitRepo(t, map[string]string{
				"SKILL.md":                           "---\nname: scoped-guide\ndescription: Markdown scope fixture\n---\n# Guide\n" + tc.visible,
				".research/notes.md":                 "Source https://history.example.net/notes\n",
				".task-board/.resources/x.md":        "Source https://history.example.net/board\n",
				"tools/board-cli/vendor/m/README.md": "Source https://history.example.net/vendor\n",
				"tools/board-cli/go.mod":             "module example.test/board\ngo 1.25.5\n",
				"tools/board-cli/cmd/tool/main.go":   "package main\nfunc main() {}\n",
				"agent-skill.json":                   `{"schema_version":6,"capabilities":` + tc.capabilities + `,"build_roots":["tools/board-cli"],"commands":{"scope-tool":{"type":"build","driver":"go-v1","source_dir":"tools/board-cli/cmd/tool"}}}`,
			}, "v1.0.0")
			operand := ids.serve(skill, "https://example.com/skills/scoped-guide")
			profile := t.TempDir()
			writeManifestPackage(t, profile, `{"schema_version":1,"name":"scope-profile","version":"1.0.0","requires":{"skills":{"scoped-guide":{"git":"`+operand+`","range":"*"}}}}`, nil)
			info, _, _, err := Install(home, InstallOptions{Operand: profile})
			if !tc.allowed {
				if err == nil || !strings.Contains(err.Error(), "audit.capability.network-undeclared") || !strings.Contains(err.Error(), "SKILL.md") || !strings.Contains(err.Error(), "contacts api.example.net,") {
					t.Fatalf("visible contact was not specifically rejected: %v", err)
				}
				for _, path := range []string{sourcePath(home, "scope-profile"), lockPath(home, "scope-profile")} {
					if _, err := os.Stat(path); !os.IsNotExist(err) {
						t.Fatalf("refused install published %s: %v", filepath.Base(path), err)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("Markdown scope install refused: %v", err)
			}
			if _, ok := info.Lock.Find(contextlock.KindSkill, "scoped-guide"); !ok {
				t.Fatal("successful install omitted skill")
			}
			if _, err := os.Stat(lockPath(home, "scope-profile")); err != nil {
				t.Fatalf("successful install published no lock: %v", err)
			}
		})
	}
}
