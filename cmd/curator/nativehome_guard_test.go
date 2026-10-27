package main

import (
	"context"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/testhomeguard"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The outer supervisor checks even direct os.Exit and test-goroutine panics.
func exitGuardedTests(code int) { os.Exit(code) }

// Both child controls exercise this package's actual TestMain exit boundary.
func TestNativeHomeGuard(t *testing.T) {
	// Reuse the parent's prepared Go toolchain; taking its lock again would
	// deadlock the control child before it reaches the package-exit guard.
	testhomeguard.Control(t, "TestNativeHomeGuardChild", cliHelperProcess+"=1")
}

// This named child resolves native defaults only within synthetic fixture homes.
func TestNativeHomeGuardChild(t *testing.T) {
	testhomeguard.Child(t, func() (string, error) {
		return envprofile.NativeHome(envprofile.Adapter{ID: "claude_code", EnvVar: "CLAUDE_CONFIG_DIR", DefaultDir: ".claude", Target: "CLAUDE.md"})
	})
}

// F1: a test's explicit synthetic HOME and provider override reach the actual
// CLI helper unchanged, including the production default native-home resolver.
// The ordinary ambient-isolation control above checks the opposite boundary.
func TestNativeHomeExplicitHelper(t *testing.T) {
	origin := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestNativeHomeExplicitHelperChild$", "-test.v")
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		switch key {
		case "HOME", "USERPROFILE", "XDG_CONFIG_HOME", "CLAUDE_CONFIG_DIR", "CODEX_HOME", "PI_CODING_AGENT_DIR", "CURATOR_TEST_EXPLICIT_HOME":
			continue
		}
		cmd.Env = append(cmd.Env, entry)
	}
	cmd.Env = append(cmd.Env, "HOME="+origin, "USERPROFILE="+origin, "XDG_CONFIG_HOME="+filepath.Join(origin, ".config"), "CODEX_HOME="+filepath.Join(origin, "codex"), cliHelperProcess+"=1", "CURATOR_TEST_EXPLICIT_HOME="+origin)
	out, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(out), "explicit fixture HOME preserved") {
		t.Fatalf("explicit helper failed: %v\n%s", err, out)
	}
	if b, err := os.ReadFile(filepath.Join(origin, ".claude", "CLAUDE.md")); err != nil || string(b) != "synthetic helper context" {
		t.Fatal("helper did not write its intended fixture home")
	}
}

// This child deliberately supplies a fixture HOME; TestMain must preserve it.
func TestNativeHomeExplicitHelperChild(t *testing.T) {
	origin := os.Getenv("CURATOR_TEST_EXPLICIT_HOME")
	if origin == "" {
		return
	}
	if os.Getenv("HOME") != origin || os.Getenv("USERPROFILE") != origin || os.Getenv("XDG_CONFIG_HOME") != filepath.Join(origin, ".config") || os.Getenv("CODEX_HOME") != filepath.Join(origin, "codex") {
		t.Fatal("explicit helper environment replaced")
	}
	path, err := envprofile.NativeHome(envprofile.Adapter{ID: "claude_code", EnvVar: "CLAUDE_CONFIG_DIR", DefaultDir: ".claude", Target: "CLAUDE.md"})
	if err != nil || path != filepath.Join(origin, ".claude") {
		t.Fatal("explicit native HOME replaced")
	}
	if err := os.MkdirAll(path, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "CLAUDE.md"), []byte("synthetic helper context"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Log("explicit fixture HOME preserved")
}
