package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
)

// All five CLI operations carry takeover to native skills, preserve unnamed entries, and back up exact foreign link text.
func TestNativeSkillsCLIOperations(t *testing.T) {
	for _, op := range []string{"install", "use", "sync", "update", "repair"} {
		t.Run(op, func(t *testing.T) {
			source, _ := profileHome(t)
			pinOperatorHome(t)
			writeNativeCredentials(t)
			repo := t.TempDir()
			if err := os.WriteFile(filepath.Join(repo, "SKILL.md"), []byte("# myskill\n"), 0644); err != nil {
				t.Fatal(err)
			}
			profileInstallGit(t, repo, "init", "--quiet")
			profileInstallGit(t, repo, "add", ".")
			profileInstallGit(t, repo, "commit", "--quiet", "-m", "fixture")
			profileInstallGit(t, repo, "tag", "v1.0.0")
			gitconfig := filepath.Join(t.TempDir(), "gitconfig")
			if err := os.WriteFile(gitconfig, []byte("[url \""+gitFileURL(repo)+"\"]\n\tinsteadOf = https://example.com/native-skill\n"), 0644); err != nil {
				t.Fatal(err)
			}
			t.Setenv("GIT_CONFIG_GLOBAL", gitconfig)
			t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
			pkg := t.TempDir()
			writeContextPackage(t, pkg, "native-profile", "1.0.0", "native\n")
			manifest := `{"schema_version":1,"name":"native-profile","version":"1.0.0","context":{"modules":[{"path":"a.md"}]},"requires":{"skills":{"myskill":{"git":"https://example.com/native-skill","tag":"v1.0.0"}}}}`
			if err := os.WriteFile(filepath.Join(pkg, "agent-context.json"), []byte(manifest), 0644); err != nil {
				t.Fatal(err)
			}
			if op != "install" {
				if code, _, stderr := runProfile(t, source, "profile", "install", pkg, "--use"); code != exitOK {
					t.Fatal(stderr)
				}
			}
			native := t.TempDir()
			t.Setenv("CLAUDE_CONFIG_DIR", native)
			if err := os.Mkdir(filepath.Join(native, "skills"), 0755); err != nil {
				t.Fatal(err)
			}
			target := filepath.Join(native, "skills", "myskill")
			if err := os.Symlink("../../retired/missing", target); err != nil {
				t.Fatal(err)
			}
			unnamed := filepath.Join(native, "skills", "unnamed")
			if err := os.WriteFile(unnamed, []byte("untouched\n"), 0644); err != nil {
				t.Fatal(err)
			}
			var args []string
			switch op {
			case "install":
				args = []string{"profile", "install", pkg, "--use"}
			case "use":
				args = []string{"profile", "use", "native-profile"}
			case "sync":
				args = []string{"profile", "sync"}
			case "update":
				if err := os.WriteFile(filepath.Join(pkg, "context", "a.md"), []byte("updated\n"), 0644); err != nil {
					t.Fatal(err)
				}
				args = []string{"profile", "update", "native-profile"}
			case "repair":
				args = []string{"env", "resolve", "claude_code", "--profile", "native-profile", "--repair"}
			}
			// Retry with consent is the nearby positive control for the same target.
			if code, _, stderr := runProfile(t, source, args...); code != exitFail || !strings.Contains(stderr, "environment_foreign_manager_detected") {
				t.Fatalf("without takeover = %d: %s", code, stderr)
			}
			if got, err := os.Readlink(target); err != nil || got != "../../retired/missing" {
				t.Fatal("refusal replaced link")
			}
			if _, err := os.Lstat(filepath.Join(native, ".agent-environment-backup")); !os.IsNotExist(err) {
				t.Fatal("refusal created backup")
			}
			if code, _, stderr := runProfile(t, source, append(args, "--takeover")...); code != exitOK {
				t.Fatalf("takeover = %d: %s", code, stderr)
			}
			if link, err := os.Readlink(target); err != nil || !filepath.IsAbs(link) {
				t.Fatalf("native link missing: %q %v", link, err)
			}
			marker, err := envmarker.Read(native)
			if err != nil || marker == nil {
				t.Fatalf("native marker: %v", err)
			}
			paths := marker.Surfaces[envmarker.SurfaceSkills].Paths
			if len(paths) != 1 || paths[0] != "skills/myskill" {
				t.Fatalf("native record = %v", paths)
			}
			if link, err := os.Readlink(filepath.Join(native, ".agent-environment-backup", "1", "skills", "myskill")); err != nil || link != "../../retired/missing" {
				t.Fatalf("backup link = %q %v", link, err)
			}
			if got, err := os.ReadFile(unnamed); err != nil || string(got) != "untouched\n" {
				t.Fatal("unnamed entry changed")
			}
		})
	}
}

// Non-mutating and unrelated CLI rows refuse takeover; resolve additionally requires repair and forbids dry-run takeover.
func TestNativeSkillsTakeoverFlagBoundary(t *testing.T) {
	for _, args := range [][]string{
		{"profile", "list"}, {"profile", "remove", "default"}, {"profile", "import"},
		{"env", "status"}, {"env", "resolve", "claude_code"}, {"env", "resolve", "claude_code", "--repair", "--dry-run"},
		{"env", "unmanage", "claude_code"}, {"status"},
	} {
		t.Run(strings.Join(args, "-"), func(t *testing.T) {
			source, _ := profileHome(t)
			pinOperatorHome(t)
			code, _, stderr := runProfile(t, source, append(args, "--takeover")...)
			if code != exitUsage || (!strings.Contains(stderr, "takeover") && !strings.Contains(stderr, "profile list takes no arguments")) {
				t.Fatalf("takeover admitted or unrelated failure = %d: %s", code, stderr)
			}
		})
	}
}
