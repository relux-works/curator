package shell

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envfiles"
)

func TestPosixHookEntersNestedSwitchesAndLeavesProjects(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is unavailable")
	}
	root := t.TempDir()
	first := filepath.Join(root, "first")
	second := filepath.Join(root, "second")
	outside := filepath.Join(root, "outside")
	for _, dir := range []string{filepath.Join(first, "nested"), second, outside} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := envfiles.WriteProject(first); err != nil {
		t.Fatal(err)
	}
	if err := envfiles.WriteProject(second); err != nil {
		t.Fatal(err)
	}
	hook, err := Hook("bash", false)
	if err != nil {
		t.Fatal(err)
	}
	script := `set -e
original="$PATH"
cd "$FIRST/nested"
eval "$HOOK"
printf 'first=%s\n' "$CSK_PROJECT_ROOT"
cd "$SECOND"
_curator_auto_env
printf 'second=%s\n' "$CSK_PROJECT_ROOT"
cd "$OUTSIDE"
_curator_auto_env
printf 'left=%s:%s\n' "${CURATOR_ACTIVE_ENV-unset}" "$([ "$PATH" = "$original" ] && printf restored || printf changed)"
`
	command := exec.Command(bash, "-c", script)
	command.Env = append(os.Environ(), "HOOK="+hook, "FIRST="+first, "SECOND="+second, "OUTSIDE="+outside)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("hook execution: %v\n%s", err, output)
	}
	text := string(output)
	for _, expected := range []string{"first=" + first, "second=" + second, "left=unset:restored"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("hook output lacks %q:\n%s", expected, text)
		}
	}
}

func TestHookVariants(t *testing.T) {
	for _, shellName := range []string{"zsh", "bash", "powershell"} {
		hook, err := Hook(shellName, true)
		if err != nil || hook == "" {
			t.Fatalf("Hook(%s) = %q, %v", shellName, hook, err)
		}
		if !strings.Contains(hook, "global") {
			t.Fatalf("%s hook does not include global activation", shellName)
		}
	}
	withoutGlobal, err := Hook("bash", false)
	if err != nil || strings.Contains(withoutGlobal, "_curator_global_env_file") {
		t.Fatalf("--no-global hook = %q, %v", withoutGlobal, err)
	}
	if _, err := Hook("fish", false); err == nil {
		t.Fatal("unsupported shell must fail")
	}
}
