package testgitenv

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Inspect the actual child environment before Go's TestMain can isolate it.
// Values are never logged. This probe performs no Git or native-home writes.
func init() {
	if os.Getenv("CURATOR_TEST_RAW_GIT_PREFIX") != "1" {
		return
	}
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(strings.ToUpper(key), "GIT_") && key != "GIT_CONFIG_GLOBAL" && key != "GIT_CONFIG_NOSYSTEM" {
			fmt.Fprintf(os.Stderr, "raw child inherited Git name outside allowlist: %q\n", key)
			os.Exit(31)
		}
	}
	if os.Getenv("GIT_CONFIG_GLOBAL") == "" || os.Getenv("GIT_CONFIG_NOSYSTEM") != "1" {
		os.Exit(32)
	}
	fmt.Println("raw child Git environment has exactly the two fixture defaults")
	os.Exit(0)
}

func nonidentifierGitNames(t *testing.T) []string {
	t.Helper()
	var nonce [8]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		t.Fatal(err)
	}
	return []string{"GIT_F7-DASH", "GIT_F7.DOT", "GIT_F7 SPACE", "GIT_F7\nNEWLINE", fmt.Sprintf("GIT_F7_%X-RANDOM", nonce)}
}

// Go accepts and removes complete process-environment names, including names
// that Bash cannot unset. Its cleanup restores each inherited synthetic value.
func TestCompleteGitPrefixNamesInGo(t *testing.T) {
	keys := append([]string{"GIT_F7_UNKNOWN_VALID"}, nonidentifierGitNames(t)...)
	for _, key := range keys {
		t.Setenv(key, "synthetic")
	}
	cleanup := Isolate()
	for _, key := range keys {
		if _, exists := os.LookupEnv(key); exists {
			t.Errorf("Git name survived: %q", key)
		}
	}
	cleanup()
	for _, key := range keys {
		if value, exists := os.LookupEnv(key); !exists || value != "synthetic" {
			t.Errorf("Git name was not restored: %q", key)
		}
	}
}

// Each nonidentifier must cause the intended pre-command refusal, or reach a
// raw child with exactly the two allowed Git defaults. Ordinary names must
// succeed, create two distinct real commits, and allow fixture-local tracing.
// Every row preserves a complete outside-decoy preimage. No skip is permitted.
func TestShellCompleteGitPrefixNames(t *testing.T) {
	shell := os.Getenv("CURATOR_TEST_PREFIX_BASH")
	if shell == "" {
		shell = "bash"
	}
	version, err := exec.Command(shell, "-c", `printf '%s\n' "$BASH_VERSION"`).CombinedOutput()
	if err != nil {
		t.Fatalf("required Bash: %v\n%s", err, version)
	}
	t.Logf("Bash %s", strings.TrimSpace(string(version)))
	helper, err := filepath.Abs(filepath.Join("..", "..", ".github", "ci", "isolate-test-env.sh"))
	if err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ name, key, value, enumerator, refusal string }{
		{name: "clean"},
		{name: "ordinary-trace", key: "GIT_TRACE"},
		{name: "ordinary-random", key: strings.ReplaceAll(nonidentifierGitNames(t)[4], "-", "_")},
		{name: "multiline-value", key: "GIT_F7_RECORD_VALUE", value: "synthetic\nGIT_F7-DASH=still-a-value"},
		{name: "enumeration-failed", enumerator: "printf 'F7_PARTIAL=synthetic\\0'; exit 1", refusal: "complete environment enumeration unavailable"},
		{name: "enumeration-truncated", enumerator: "printf 'F7_PARTIAL=unterminated'; exit 0", refusal: "complete environment enumeration unavailable"},
	}
	for i, key := range nonidentifierGitNames(t) {
		cases = append(cases, struct{ name, key, value, enumerator, refusal string }{name: fmt.Sprintf("nonidentifier-%d", i), key: key, refusal: "unsupported inherited Git name"})
	}
	for _, row := range cases {
		t.Run(row.name, func(t *testing.T) {
			key := row.key
			decoy := newDecoy(t)
			trace := filepath.Join(decoy, "outside.trace")
			if err := os.WriteFile(trace, []byte("synthetic outside sentinel\n"), 0600); err != nil {
				t.Fatal(err)
			}
			before := decoyImage(t, decoy)
			work := t.TempDir()
			marker := filepath.Join(work, "fixture-command-dispatched")
			cmd := exec.Command(shell, "-c", `
. "$1" || exit 2
isolate_test_env "$2" || exit 2
printf dispatched >"$4" || exit 4
CURATOR_TEST_RAW_GIT_PREFIX=1 "$3" || exit 31
test ! -s "$GIT_CONFIG_GLOBAL" && test "$GIT_CONFIG_NOSYSTEM" = 1 || exit 4
repo="$HOME/ancestry"
git init -q -b main "$repo" || exit 4
git -C "$repo" -c user.name=Fixture -c user.email=fixture@example.invalid commit -qm main --allow-empty || exit 4
main="$(git -C "$repo" rev-parse HEAD)" || exit 4
git -C "$repo" switch -q -c topic || exit 4
git -C "$repo" -c user.name=Fixture -c user.email=fixture@example.invalid commit -qm topic --allow-empty || exit 4
topic="$(git -C "$repo" rev-parse HEAD)" || exit 4
test -n "$main" && test -n "$topic" && test "$main" != "$topic" || exit 4
GIT_TRACE="$HOME/local.trace" git -C "$repo" rev-parse HEAD >/dev/null || exit 4
grep -q 'built-in: git rev-parse HEAD' "$HOME/local.trace" || exit 4
printf 'two distinct real commits and fixture-local trace completed\n'
`, "_", helper, filepath.Join(work, "owned"), executable, marker)
			cmd.Dir = work
			// Clear inherited boundary inputs, then supply only this row's
			// explicit absolute HOME and synthetic Git probe value.
			for _, entry := range os.Environ() {
				name, _, _ := strings.Cut(entry, "=")
				if strings.HasPrefix(strings.ToUpper(name), "GIT_") || strings.HasPrefix(name, "XDG_") || name == "HOME" || name == "USERPROFILE" || name == "SSH_AUTH_SOCK" || name == "SSH_ASKPASS" || name == "GOFLAGS" || name == "CODEX_HOME" || name == "CLAUDE_CONFIG_DIR" || name == "PI_CODING_AGENT_DIR" || name == "CURATOR_TEST_RAW_GIT_PREFIX" {
					continue
				}
				cmd.Env = append(cmd.Env, entry)
			}
			cmd.Env = append(cmd.Env, "HOME="+decoy, "USERPROFILE="+decoy)
			if key != "" {
				value := trace
				if row.value != "" {
					value = row.value
				}
				cmd.Env = append(cmd.Env, key+"="+value)
			}
			if row.enumerator != "" {
				// Bounded fault injection in the enumeration dependency. This
				// does not claim executable provenance or a PATH sandbox.
				shim := filepath.Join(work, "bin")
				if err := os.Mkdir(shim, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(shim, "env"), []byte("#!/bin/sh\n"+row.enumerator+"\n"), 0700); err != nil {
					t.Fatal(err)
				}
				for i, entry := range cmd.Env {
					if strings.HasPrefix(entry, "PATH=") {
						cmd.Env[i] = "PATH=" + shim + string(os.PathListSeparator) + strings.TrimPrefix(entry, "PATH=")
					}
				}
			}
			out, runErr := cmd.CombinedOutput()
			requireDecoyImage(t, before, decoyImage(t, decoy))
			if exit, ok := runErr.(*exec.ExitError); ok && exit.ExitCode() == 2 && row.refusal != "" {
				if !strings.Contains(string(out), "isolate-test-env: "+row.refusal) {
					t.Fatalf("unrelated refusal: %s", out)
				}
				if _, err := os.Lstat(marker); !os.IsNotExist(err) {
					t.Fatal("fixture command dispatched after refusal")
				}
				t.Logf("pre-command refusal for %q", key)
				return
			}
			if row.refusal != "" && row.enumerator != "" {
				t.Fatalf("enumeration failure did not refuse before dispatch: %v\n%s", runErr, out)
			}
			if runErr != nil || !strings.Contains(string(out), "raw child Git environment has exactly the two fixture defaults") || !strings.Contains(string(out), "two distinct real commits and fixture-local trace completed") {
				t.Fatalf("complete prefix boundary for %q: %v\n%s", key, runErr, out)
			}
		})
	}
}
