package envprofile

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

// The built CLI with explicit strict/high audit rejects all nine root aliases
// before lock publication or native-context exposure on case-insensitive hosts.
// Case-sensitive directory aliases install without publishing their bytes.
// Canonical undeclared contacts also refuse, whereas research noise and declared
// contacts install and expose the expected bytes through the native Claude home.
func TestProfileInstallMarkdownCaseAliasesCLI(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "curator")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.Command("go", "build", "-o", binary, "./cmd/curator")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build exact-source CLI: %v\n%s", err, output)
	}
	probe := t.TempDir()
	if err := os.WriteFile(filepath.Join(probe, "case-probe"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	_, aliasErr := os.Stat(filepath.Join(probe, "CASE-PROBE"))
	insensitive := aliasErr == nil
	cases := []struct {
		file, published          string
		declared, allowed, alias bool
	}{
		{"SKILL.MD", "SKILL.md", false, false, true},
		{"References/guide.md", "references/guide.md", false, false, true},
		{"AGENTS/guide.md", "agents/guide.md", false, false, true},
		{"ASSETS/guide.md", "assets/guide.md", false, false, true},
		{".SKILL_TRIGGERS/guide.md", ".skill_triggers/guide.md", false, false, true},
		{"Templates/guide.md", "templates/guide.md", false, false, true},
		{"Examples/guide.md", "examples/guide.md", false, false, true},
		{"Data/guide.md", "data/guide.md", false, false, true},
		{"Scripts/guide.md", "scripts/guide.md", false, false, true},
		{"SKILL.md", "SKILL.md", false, false, false},
		{"references/guide.md", "references/guide.md", false, false, false},
		{"assets/guide.md", "assets/guide.md", false, false, false},
		{"references/guide.md", "references/guide.md", true, true, false},
		{".research/notes.md", "", false, true, false},
		{"References/guide.md", "references/guide.md", true, true, true},
	}
	for _, tc := range cases {
		name := tc.file
		if tc.declared {
			name += "-declared"
		}
		t.Run(name, func(t *testing.T) {
			if tc.alias && !insensitive && tc.file == "SKILL.MD" {
				t.Skip("test filesystem is case-sensitive: root entry alias requires case folding; copy test asserts missing canonical entry refusal")
			}
			allowed, published := tc.allowed, tc.published
			if tc.alias && !insensitive {
				allowed, published = true, ""
			}
			home := t.TempDir()
			pinOperatorHome(t, home)
			homes := pinHomes(t)
			ids := newGitIdentities(t)
			caps := `{}`
			if tc.declared {
				caps = `{"network":["api.example.net"]}`
			}
			frontmatter := "---\nname: scoped-guide\ndescription: Markdown alias fixture\n---\n# Guide\n"
			files := map[string]string{
				"SKILL.md":                    frontmatter,
				"agent-skill.json":            `{"schema_version":3,"capabilities":` + caps + `}`,
				".research/history.md":        "Source https://history.example.net/notes\n",
				".task-board/.resources/x.md": "Source https://history.example.net/board\n",
			}
			if tc.file == "SKILL.MD" {
				delete(files, "SKILL.md")
			}
			if tc.published == "SKILL.md" {
				files[tc.file] = frontmatter + "Read https://api.example.net/guide\n"
			} else {
				files[tc.file] = "Read https://api.example.net/guide\n"
			}
			skill := gitRepo(t, files, "v1.0.0")
			operand := ids.serve(skill, "https://example.com/skills/scoped-guide")
			profile := t.TempDir()
			writeManifestPackage(t, profile, `{"schema_version":1,"name":"scope-profile","version":"1.0.0","requires":{"skills":{"scoped-guide":{"git":"`+operand+`","range":"*"}}}}`, nil)
			configPath := filepath.Join(home, "config.json")
			configBytes, err := json.Marshal(map[string]any{"schema_version": 1, "skills_root": filepath.Join(home, "skills"), "projects": map[string]any{}, "audit": map[string]any{"enabled": true, "mode": "strict", "fail_on": "high", "backend": "null"}})
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(configPath, configBytes, 0600); err != nil {
				t.Fatal(err)
			}
			systemPath := filepath.Join(home, "system.json")
			if err := os.WriteFile(systemPath, []byte(`{"schema_version":1}`), 0600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("CURATOR_CONFIG", configPath)
			t.Setenv("CURATOR_SYSTEM_CONFIG", systemPath)
			command := exec.Command(binary, "profile", "install", profile)
			output, err := command.CombinedOutput()
			if !allowed {
				exit, ok := err.(*exec.ExitError)
				if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "audit.capability.network-undeclared") || !strings.Contains(string(output), tc.published) {
					t.Fatalf("CLI alias not specifically refused: %v\n%s", err, output)
				}
				for _, path := range []string{sourcePath(home, "scope-profile"), lockPath(home, "scope-profile")} {
					if _, err := os.Stat(path); !os.IsNotExist(err) {
						t.Fatalf("refused CLI published profile metadata: %v", err)
					}
				}
				// Follow the expected native skill link, which would expose the attack
				// bytes after a successful install; absence is required on refusal.
				if _, err := os.Stat(filepath.Join(homes["claude_code"], "skills", "scoped-guide")); !os.IsNotExist(err) {
					t.Fatalf("refused CLI exposed native skill context: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("CLI positive control refused: %v\n%s", err, output)
			}
			if _, err := os.Stat(lockPath(home, "scope-profile")); err != nil {
				t.Fatalf("successful CLI omitted lock: %v", err)
			}
			native := filepath.Join(homes["claude_code"], "skills", "scoped-guide", "SKILL.md")
			if published != "" {
				native = filepath.Join(homes["claude_code"], "skills", "scoped-guide", filepath.FromSlash(published))
			}
			payload, err := os.ReadFile(native)
			if err != nil {
				t.Fatalf("successful control omitted native context: %v", err)
			}
			if tc.alias && !insensitive {
				for _, path := range []string{tc.file, tc.published} {
					if _, err := os.Stat(filepath.Join(homes["claude_code"], "skills", "scoped-guide", filepath.FromSlash(path))); !os.IsNotExist(err) {
						t.Fatalf("case-sensitive alias reached native context: %s: %v", path, err)
					}
				}
				if strings.Contains(string(payload), "https://api.example.net/guide") {
					t.Fatal("case-sensitive alias contact reached native entry")
				}
			}
			if published != "" && !strings.Contains(string(payload), "https://api.example.net/guide") {
				t.Fatal("declared control lost contact bytes")
			}
		})
	}
}
