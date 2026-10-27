package testgitenv

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Replay the actual standalone ancestry prefix, including its entry boundary.
// Clean and hostile runs must both create their two real fixture commits while
// every HEAD/index/ref/reflog/object/config byte in the decoy stays unchanged.
func TestShellGateFixtureIsolation(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Fatal("gate fixture qualification requires bash")
	}
	script, err := filepath.Abs(filepath.Join("..", "..", ".github", "ci", "gate-selftest.sh"))
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(script)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(body), "\n")
	end := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "TOPIC_COMMIT=") {
			end = i + 1
			break
		}
	}
	if end < 0 {
		t.Fatal("standalone ancestry fixture boundary not found")
	}
	prefix := strings.Join(lines[:end], "\n") + `
test -n "$MAIN_COMMIT" && test -n "$TOPIC_COMMIT" && test "$MAIN_COMMIT" != "$TOPIC_COMMIT" || exit 3
test "$FAIL" -eq 0 || exit 3
printf 'ancestry fixture completed\n'
`
	cases := []string{"clean", "repository", "storage", "config-count", "config-parameters", "config-global"}
	for _, mode := range cases {
		t.Run(mode, func(t *testing.T) {
			decoy := newDecoy(t)
			before := decoyImage(t, decoy)
			cmd := exec.Command("bash", "-c", prefix, script)
			cmd.Env = append(os.Environ(), "SSH_AUTH_SOCK="+filepath.Join(decoy, "absent-agent"))
			switch mode {
			case "repository":
				cmd.Env = append(cmd.Env, "GIT_DIR="+filepath.Join(decoy, ".git"), "GIT_WORK_TREE="+decoy, "GIT_INDEX_FILE="+filepath.Join(decoy, ".git", "index"))
			case "storage":
				cmd.Env = append(cmd.Env, "GIT_COMMON_DIR="+filepath.Join(decoy, ".git"), "GIT_OBJECT_DIRECTORY="+filepath.Join(decoy, ".git", "objects"))
			case "config-count":
				cmd.Env = append(cmd.Env, "GIT_CONFIG_COUNT=3", "GIT_CONFIG_KEY_0=commit.gpgsign", "GIT_CONFIG_VALUE_0=true", "GIT_CONFIG_KEY_1=gpg.format", "GIT_CONFIG_VALUE_1=ssh", "GIT_CONFIG_KEY_2=user.signingkey", "GIT_CONFIG_VALUE_2="+filepath.Join(decoy, "absent-key"))
			case "config-parameters":
				cmd.Env = append(cmd.Env, "GIT_CONFIG_PARAMETERS='commit.gpgsign=true' 'gpg.format=ssh' 'user.signingkey="+filepath.Join(decoy, "absent-key")+"'")
			case "config-global":
				config := filepath.Join(t.TempDir(), "hostile.gitconfig")
				if err := os.WriteFile(config, []byte("[commit]\n gpgsign = true\n[gpg]\n format = ssh\n[user]\n signingkey = /absent-fixture-signing-key\n"), 0600); err != nil {
					t.Fatal(err)
				}
				cmd.Env = append(cmd.Env, "GIT_CONFIG_GLOBAL="+config)
			}
			output, childErr := cmd.CombinedOutput()
			requireDecoyImage(t, before, decoyImage(t, decoy))
			if childErr != nil || !strings.Contains(string(output), "ancestry fixture completed") {
				t.Fatalf("shell ancestry %s: %v\n%s", mode, childErr, output)
			}
		})
	}
}

// The shell equivalent must scrub the same independent selector inventory and
// dynamic config channels as Go, with absolute physical exports even when its
// owned root is supplied through a relative path. This guards against drift.
func TestShellBoundaryParity(t *testing.T) {
	script, err := filepath.Abs(filepath.Join("..", "..", ".github", "ci", "isolate-test-env.sh"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("bash", "-c", `
. "$1" || exit 2
isolate_test_env environment || exit 2
shift
for key do
  if printenv "$key" >/dev/null; then printf 'selector survived: %s\n' "$key"; exit 3; fi
done
case "$HOME:$XDG_CONFIG_HOME:$GIT_CONFIG_GLOBAL:$TMPDIR:$TMP:$TEMP" in /*:/*:/*:/*:/*:/*) ;; *) exit 4 ;; esac
test "$GIT_CONFIG_NOSYSTEM" = 1 && test ! -s "$GIT_CONFIG_GLOBAL" || exit 5
printf 'shell boundary isolated\n'
`, "_", script)
	cmd.Args = append(cmd.Args, repositorySelectors...)
	cmd.Args = append(cmd.Args, nativeRootOverrides...)
	cmd.Args = append(cmd.Args, "GIT_CONFIG_COUNT", "GIT_CONFIG_PARAMETERS", "GIT_CONFIG_KEY_41", "GIT_CONFIG_VALUE_41", "SSH_AUTH_SOCK")
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), "GIT_CONFIG_COUNT=42", "GIT_CONFIG_PARAMETERS=poison", "GIT_CONFIG_KEY_41=poison", "GIT_CONFIG_VALUE_41=poison", "SSH_AUTH_SOCK=absent")
	for _, key := range repositorySelectors {
		cmd.Env = append(cmd.Env, key+"=poison")
	}
	for _, key := range nativeRootOverrides {
		cmd.Env = append(cmd.Env, key+"=poison")
	}
	output, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(output), "shell boundary isolated") {
		t.Fatalf("shell selector parity: %v\n%s", err, output)
	}
}
