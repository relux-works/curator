package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/testtoolchain"
)

// Only an admitted Git Bash file can serve the Windows hook test; a WSL
// launcher, an unrelated bin directory, and a directory named bash.exe cannot.
func TestWindowsGitBashSelection(t *testing.T) {
	root := t.TempDir()
	wsl := filepath.Join(root, "System32")
	unrelated := filepath.Join(root, "other", "bin")
	gitBin := filepath.Join(root, "Git", "bin")
	gitUsrBin := filepath.Join(root, "Git", "usr", "bin")
	for _, dir := range []string{wsl, unrelated, gitBin, gitUsrBin} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, dir := range []string{wsl, unrelated} {
		if err := os.WriteFile(filepath.Join(dir, "bash.exe"), []byte("not Git Bash"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	path := strings.Join([]string{wsl, unrelated, gitBin, gitUsrBin}, string(os.PathListSeparator))
	if got := windowsGitBash(path); got != "" {
		t.Fatalf("non-Git launcher selected: %s", got)
	}
	fakeDirectory := filepath.Join(gitBin, "bash.exe")
	if err := os.Mkdir(fakeDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := windowsGitBash(path); got != "" {
		t.Fatalf("directory selected as Git Bash: %s", got)
	}
	for _, dir := range []string{gitUsrBin, gitBin} {
		candidate := filepath.Join(dir, "bash.exe")
		if dir == gitBin {
			if err := os.Remove(fakeDirectory); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(candidate, []byte("Git Bash fixture"), 0o755); err != nil {
			t.Fatal(err)
		}
		if got := windowsGitBash(path); got != candidate {
			t.Fatalf("Git Bash selection = %q, want %q", got, candidate)
		}
	}
}

// Removing either selection restriction must fail its specific refusal witness,
// rather than an unrelated build or setup failure.
func TestWindowsGitBashSelectionNarrowingMutant(t *testing.T) {
	for _, tc := range []struct{ name, before, after, failure string }{
		{"Git-directory", `if !strings.EqualFold(filepath.Base(directory), "bin") || !strings.EqualFold(filepath.Base(parent), "Git") {`, `if false {`, "non-Git launcher selected:"},
		{"regular-file", `err == nil && info.Mode().IsRegular()`, `err == nil && info != nil`, "directory selected as Git Bash:"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source, err := filepath.Abs("hook_test.go")
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(source)
			if err != nil {
				t.Fatal(err)
			}
			data = []byte(strings.ReplaceAll(string(data), "\r\n", "\n"))
			if strings.Count(string(data), tc.before) != 1 {
				t.Fatal("Git Bash mutation anchor is not unique")
			}
			directory := t.TempDir()
			replacement := filepath.Join(directory, "hook_test.go")
			if err := os.WriteFile(replacement, []byte(strings.Replace(string(data), tc.before, tc.after, 1)), 0o600); err != nil {
				t.Fatal(err)
			}
			data, err = json.Marshal(map[string]any{"Replace": map[string]string{source: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			overlay := filepath.Join(directory, "overlay.json")
			if err := os.WriteFile(overlay, data, 0o600); err != nil {
				t.Fatal(err)
			}
			command := exec.Command(testtoolchain.Go(t), "test", "-overlay", overlay, "-count=1", "-timeout=3m", "-v", "-run", "^TestWindowsGitBashSelection$", ".")
			// The child runs only path-selection fixtures; the parent owns the GOROOT lock.
			command.Env = append(os.Environ(), cliPathChild+"=1")
			output, err := command.CombinedOutput()
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 1 || !strings.Contains(string(output), "--- FAIL: TestWindowsGitBashSelection") || !strings.Contains(string(output), tc.failure) || strings.Contains(string(output), "build failed") {
				t.Fatalf("unrestricted Bash mutant missed intended refusal: %v\n%s", err, output)
			}
		})
	}
}
