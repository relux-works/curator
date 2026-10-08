package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// Second-operator first-run rows (TASK-260930-3b3oyi): every row drives
// the production CLI entry run() — or the built binary for the nested
// process-tree row — and pins the exact message and exit code.

const wantConfigHint = "curator: run `curator bootstrap --skills-root <dir>` first (see docs/cli.md#bootstrap)"

func runFile(t *testing.T, configPath string, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr strings.Builder
	code := run(args, fileConfigSource(configPath), &stdout, &stderr)
	t.Logf("production run %v: exit=%d stdout=%q stderr=%q", args, code, stdout.String(), stderr.String())
	return code, stdout.String(), stderr.String()
}

func TestGlobalInitCleanHomeRefusesWithBootstrapHint(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	for _, args := range [][]string{
		{"global", "init"},
		{"global", "list"},
		{"profile", "list"},
		{"env", "status"},
	} {
		code, stdout, stderr := runFile(t, configPath, args...)
		want := "curator: global config not found: " + configPath + "\n" + wantConfigHint + "\n"
		if code != exitFail || stderr != want || stdout != "" {
			t.Fatalf("%v on a clean home = %d\nstdout %q\nstderr %q\nwant stderr %q", args, code, stdout, stderr, want)
		}
	}
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatalf("the refusal created the config: %v", err)
	}
}

// A malformed config is a failed read, not an absence: it must not get the
// bootstrap hint that is defined for the missing file.
func TestGlobalInitMalformedConfigHasNoBootstrapHint(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	code, _, stderr := runFile(t, configPath, "global", "init")
	if code != exitFail || strings.Contains(stderr, "bootstrap") || strings.Contains(stderr, "not found") {
		t.Fatalf("malformed config = %d\nstderr %q", code, stderr)
	}
}

func TestGlobalInitHelpWithoutConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	for _, args := range [][]string{
		{"global", "--help"},
		{"global", "-h"},
		{"global", "help"},
		{"profile", "--help"},
		{"profile", "help"},
		{"env", "--help"},
		{"env", "help"},
	} {
		code, stdout, stderr := runFile(t, configPath, args...)
		if code != exitOK || stderr != "" || !strings.HasPrefix(stdout, "usage: curator "+args[0]) ||
			!strings.Contains(stdout, "curator bootstrap --skills-root <dir>") {
			t.Fatalf("%v = %d\nstdout %q\nstderr %q", args, code, stdout, stderr)
		}
	}
}

func TestGlobalInitSubcommandHelpWithoutConfigRefuses(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	for _, args := range [][]string{
		{"global", "init", "--help"}, {"global", "add", "--help"},
		{"env", "status", "-h"}, {"env", "resolve", "--help"},
		{"profile", "install", "--help"},
	} {
		code, stdout, stderr := runFile(t, configPath, args...)
		want := "curator: global config not found: " + configPath + "\n" + wantConfigHint + "\n"
		if code != exitFail || stdout != "" || stderr != want {
			t.Fatalf("%v = %d stdout %q stderr %q; want refusal %q", args, code, stdout, stderr, want)
		}
	}
}

const wantPostureWarning = "warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults\n"

func TestGlobalInitEnvResolveSubcommandHelpPreservesFlagSet(t *testing.T) {
	home := t.TempDir()
	configPath := filepath.Join(home, "config.json")
	t.Setenv(postureWarnedEnv, "")
	if code, _, stderr := runFile(t, configPath, "bootstrap", "--non-interactive", "--skills-root", filepath.Join(home, "skills")); code != exitOK {
		t.Fatalf("bootstrap = %d\n%s", code, stderr)
	}
	for _, row := range []struct {
		name  string
		args  []string
		usage string
	}{
		{"env resolve", []string{"env", "resolve", "--help"}, "Usage of env resolve:\n  -dry-run\n    \treport store rebuilds without modifying state (requires --repair)\n  -format string\n    \tfragment format: json | env | shell (default \"json\")\n  -profile string\n    \tprofile name (default: the current profile)\n  -repair\n    \tprovision or repair the managed home under the mutation lock\n  -takeover\n    \ttake over the unmanaged files the repair would write (only with --repair)\ncurator: env resolve <env-id> [--profile <name>] [--repair [--dry-run]] [--takeover] [--format json|env|shell]\n" + envAliasUsage + "\n"},
		{"global add", []string{"global", "add", "--help"}, "Usage of global add:\n  -branch string\n    \tgit branch\n  -git string\n    \tgit clone URL\n  -revision string\n    \tgit revision\n  -source string\n    \tsource directory under skills_root\n  -tag string\n    \tgit tag\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			code, stdout, stderr := runFile(t, configPath, row.args...)
			if code != exitUsage || stdout != "" || stderr != wantPostureWarning+row.usage {
				t.Fatalf("subcommand help = %d stdout %q stderr %q; want exit 2 stderr %q", code, stdout, stderr, wantPostureWarning+row.usage)
			}
		})
	}
}

func TestBootstrapDocsExampleRuns(t *testing.T) {
	docs, err := os.ReadFile(filepath.Join("..", "..", "docs", "cli.md"))
	if err != nil {
		t.Fatal(err)
	}
	match := regexp.MustCompile("(?m)^curator bootstrap --if-missing --non-interactive.*$").Find(docs)
	if match == nil {
		t.Fatal("docs/cli.md lost the bootstrap example")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	args := strings.Fields(strings.ReplaceAll(string(match), `"$HOME/skills"`, filepath.Join(home, "skills")))[1:]
	configPath := filepath.Join(home, ".curator", "config.json")
	code, stdout, stderr := runFile(t, configPath, args...)
	if code != exitOK || !strings.Contains(stdout, "wrote "+configPath) {
		t.Fatalf("documented example %v = %d\nstdout %q\nstderr %q", args, code, stdout, stderr)
	}
	// The old example (no --skills-root) is a usage error.
	code, _, stderr = runFile(t, filepath.Join(t.TempDir(), "config.json"), "bootstrap", "--if-missing", "--non-interactive")
	if code != exitUsage || stderr != "curator: bootstrap requires --skills-root\n" {
		t.Fatalf("bootstrap without --skills-root = %d stderr %q", code, stderr)
	}
	// After bootstrap global init succeeds.
	if code, _, stderr := runFile(t, configPath, "global", "init"); code != exitOK {
		t.Fatalf("global init after bootstrap = %d\n%s", code, stderr)
	}
}

func TestEnvResolveStaleHintsRepair(t *testing.T) {
	source, _ := profileHome(t)
	installCLIEnvProfile(t, source)
	code, stdout, stderr := runProfile(t, source, "env", "resolve", "codex_cli", "--profile", "acme")
	if code != exitFail || stdout != "" || !strings.HasSuffix(stderr, "curator: environment_home_stale: home unprovisioned; rerun with --repair\n") {
		t.Fatalf("resolve without --repair = %d\nstdout %q\nstderr %q", code, stdout, stderr)
	}
	// --repair --dry-run already names the repair: no rerun hint.
	code, _, stderr = runProfile(t, source, "env", "resolve", "codex_cli", "--profile", "acme", "--repair", "--dry-run")
	if strings.Contains(stderr, "rerun with --repair") {
		t.Fatalf("dry-run repair = %d carries the rerun hint:\n%s", code, stderr)
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", "codex_cli", "--profile", "acme", "--repair"); code != exitOK {
		t.Fatalf("resolve --repair = %d\n%s", code, stderr)
	}
}

// TestEnvStatusCheckCurrentScopeOnly: a non-current profile's
// unprovisioned homes are printed but do not fail --check; the same
// finding on the current profile still does.
func TestEnvStatusCheckCurrentScopeOnly(t *testing.T) {
	source, _ := profileHome(t)
	writeNativeCredentials(t)
	bin := t.TempDir()
	for _, name := range []string{"curator-run", "curator-session"} {
		full := filepath.Join(bin, name)
		if runtime.GOOS == "windows" {
			full += ".exe"
		}
		if err := os.WriteFile(full, []byte(""), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	if code, _, stderr := runProfile(t, source, "profile", "use", "acme"); code != exitOK {
		t.Fatalf("use stderr:\n%s", stderr)
	}
	envs := []string{"claude_code", "codex_cli", "opencode", "pi"}
	for _, env := range envs[1:] {
		if code, _, stderr := runProfile(t, source, "env", "resolve", env, "--profile", "acme", "--repair"); code != exitOK {
			t.Fatalf("repair %s stderr:\n%s", env, stderr)
		}
	}
	// Current profile still has one unprovisioned home: --check fails.
	code, stdout, _ := runProfile(t, source, "env", "status", "--check")
	if code != exitFail || !strings.Contains(stdout, "acme claude_code: non-current, unprovisioned") {
		t.Fatalf("current-scope unprovisioned --check = %d\n%s", code, stdout)
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", envs[0], "--profile", "acme", "--repair"); code != exitOK {
		t.Fatalf("repair stderr:\n%s", stderr)
	}
	code, stdout, _ = runProfile(t, source, "env", "status", "--check")
	if code != exitOK {
		t.Fatalf("healthy current scope --check = %d, want %d\n%s", code, exitOK, stdout)
	}
	if !strings.Contains(stdout, "default codex_cli: non-current, unprovisioned") {
		t.Fatalf("the non-current profile's row is no longer printed:\n%s", stdout)
	}
}

func postureWarnings(stderr string) int {
	return strings.Count(stderr, "security_posture_permissive")
}

func TestPostureWarningOncePerProcessTree(t *testing.T) {
	home := t.TempDir()
	configPath := filepath.Join(home, "config.json")
	t.Setenv(postureWarnedEnv, "")
	t.Setenv("CURATOR_CONFIG", configPath)
	if code, _, stderr := runFile(t, configPath, "bootstrap", "--non-interactive", "--skills-root", filepath.Join(home, "skills")); code != exitOK {
		t.Fatalf("bootstrap = %d\n%s", code, stderr)
	}
	// A single command still prints the warning exactly once.
	code, _, stderr := runFile(t, configPath, "env", "status")
	if code != exitOK || postureWarnings(stderr) != 1 {
		t.Fatalf("top-level env status = %d, printed %d posture warnings:\n%s", code, postureWarnings(stderr), stderr)
	}
	// A native manager binary named curator-run stands in for the provider:
	// curator run env status → curator-run env status. Both invocations
	// drive main() and the real dispatch, without a platform-specific shell.
	plant := t.TempDir()
	binary := filepath.Join(plant, "curator-run")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.Command("go", "build", "-p", "1", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, output)
	}
	t.Setenv("PATH", plant+string(os.PathListSeparator)+os.Getenv("PATH"))
	runBinary := func(args ...string) (int, string, string) {
		t.Helper()
		command := exec.Command(binary, args...)
		var stdout, stderr strings.Builder
		command.Stdout, command.Stderr = &stdout, &stderr
		code := exitOK
		if err := command.Run(); err != nil {
			var exitError *exec.ExitError
			if !errors.As(err, &exitError) {
				t.Fatal(err)
			}
			code = exitError.ExitCode()
		}
		t.Logf("production binary %v: exit=%d stderr=%q", args, code, stderr.String())
		return code, stdout.String(), stderr.String()
	}
	// The outer process warns, the nested one must not. Successful exit
	// and the child's status output prove the provider actually ran.
	code, stdout, stderr := runBinary("run", "env", "status", "--json")
	if code != exitOK || stderr != wantPostureWarning || !strings.Contains(stdout, `"security_posture_rows"`) {
		t.Fatalf("curator run → env status = %d, printed %d posture warnings, want 1:\nstdout %s\nstderr %s", code, postureWarnings(stderr), stdout, stderr)
	}
	// The nested child alone (no marker) would warn: the suppression is
	// the marker, not a missing warning.
	code, _, stderr = runBinary("env", "status", "--json")
	if code != exitOK || stderr != wantPostureWarning {
		t.Fatalf("standalone nested command = %d, printed %d warnings:\n%s", code, postureWarnings(stderr), stderr)
	}
	for _, row := range []struct{ name, marker string }{
		{"user-set-one", "1"},
		// This test process is the binary's actual parent, so its PID + 1
		// cannot identify that parent, even if it names a live process.
		{"not-parent-pid", strconv.Itoa(os.Getpid() + 1)},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Setenv(postureWarnedEnv, row.marker)
			if code, _, stderr := runBinary("env", "status", "--json"); code != exitOK || stderr != wantPostureWarning {
				t.Fatalf("forged marker %q = %d stderr %q; want one warning %q", row.marker, code, stderr, wantPostureWarning)
			}
		})
	}
}
