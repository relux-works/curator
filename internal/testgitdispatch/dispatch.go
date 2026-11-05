// Package testgitdispatch installs test-owned acquisition dispatchers.
package testgitdispatch

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// Config describes fixture dispatch, which observes the declared command before substituting local Git
// objects. It proves URL selection and pin bytes, not wire encryption or auth.
type Config struct {
	Binary          string
	Routes          map[string]string
	Refusals        map[string]string
	EnforceHeadless bool
	Direct          []string
}

// Configure changes only the synthetic dispatch sidecar, never operator Git configuration.
func Configure(t *testing.T, path string, change func(*Config)) {
	t.Helper()
	raw, err := os.ReadFile(path) // #nosec G304 -- test-owned dispatch sidecar
	if err != nil {
		t.Fatal(err)
	}
	var config Config
	if err := json.Unmarshal(raw, &config); err != nil {
		t.Fatal(err)
	}
	change(&config)
	raw, err = json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
}

var buildOnce sync.Once
var helperBinary string
var buildError error

// Build builds a small non-race helper once per test process, outside source.
func Build() error {
	buildOnce.Do(func() {
		_, source, _, _ := runtime.Caller(0)
		dir, err := os.MkdirTemp("", "curator-git-helper-")
		if err != nil {
			buildError = err
			return
		}
		helperBinary = filepath.Join(dir, "fixture")
		if runtime.GOOS == "windows" {
			helperBinary += ".exe"
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, "go", "build", "-race=false", "-o", helperBinary, filepath.Join(filepath.Dir(source), "testdata", "helper", "main.go")) // #nosec G204 -- fixed Go compiler builds checked-in fixture source into a test-owned directory
		if out, err := cmd.CombinedOutput(); err != nil {
			buildError = fmt.Errorf("build Git fixture: %w: %s", err, out)
		}
	})
	return buildError
}

// Cleanup removes the once-built helper after the package tests exit.
func Cleanup() {
	if helperBinary != "" {
		_ = os.RemoveAll(filepath.Dir(helperBinary))
	}
}

// Install creates a new executable atomically. An installed executable is immutable.
func Install(t *testing.T, target string) {
	t.Helper()
	if err := Build(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(target); err == nil {
		return
	}
	data, err := os.ReadFile(helperBinary) // #nosec G304 -- own once-built helper
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.CreateTemp(filepath.Dir(target), ".helper-*")
	if err != nil {
		t.Fatal(err)
	}
	path := file.Name()
	defer func() { _ = os.Remove(path) }()
	if _, err = file.Write(data); err != nil {
		t.Fatal(err)
	}
	if err = file.Chmod(0700); err != nil {
		t.Fatal(err)
	}
	if err = file.Close(); err != nil {
		t.Fatal(err)
	}
	if err = os.Rename(path, target); err != nil {
		t.Fatal(err)
	}
}

// New installs a dispatcher for explicit acquisition routes in this test.
func New(t *testing.T) string {
	t.Helper()
	binary, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	name := "git"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	Install(t, filepath.Join(dir, name))
	path := filepath.Join(dir, "declared-fixture.json")
	raw, _ := json.Marshal(Config{Binary: binary, Routes: map[string]string{}})
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return path
}

// Command bounds fixture setup/read calls and keeps ordinary Git off the dispatcher.
func Command(t *testing.T, args ...string) *exec.Cmd {
	t.Helper()
	binary, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	operation := args
	for len(operation) >= 2 && (operation[0] == "-C" || operation[0] == "-c") {
		operation = operation[2:]
	}
	if len(operation) == 0 || (operation[0] != "clone" && operation[0] != "fetch") {
		for i := 0; i < 10; i++ {
			raw, err := os.ReadFile(filepath.Join(filepath.Dir(binary), "declared-fixture.json")) // #nosec G304 -- fixture sidecar beside test-owned binary
			if err != nil {
				break
			}
			var config Config
			if json.Unmarshal(raw, &config) != nil || config.Binary == binary {
				break
			}
			binary = config.Binary
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)
	cmd := exec.CommandContext(ctx, binary, args...) // #nosec G204 -- resolved Git or test-owned dispatcher, arguments are fixture commands
	cmd.WaitDelay = time.Second
	cmd.Env = HeadlessEnv(os.Environ())
	return cmd
}

// HeadlessEnv makes fixture Git calls independent of interactive operator settings.
func HeadlessEnv(base []string) []string {
	out := make([]string, 0, len(base)+4)
	for _, entry := range base {
		key, _, _ := strings.Cut(entry, "=")
		switch key {
		case "GIT_TERMINAL_PROMPT", "GIT_ASKPASS", "SSH_ASKPASS", "SSH_ASKPASS_REQUIRE", "GIT_SSH_COMMAND":
			continue
		}
		out = append(out, entry)
	}
	return append(out, "GIT_TERMINAL_PROMPT=0", "GIT_ASKPASS=", "SSH_ASKPASS=", "SSH_ASKPASS_REQUIRE=never", "GIT_SSH_COMMAND=ssh -o BatchMode=yes -o StrictHostKeyChecking=yes")
}
