package shell

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envfiles"
)

func TestPosixHookEntersNestedSwitchesAndLeavesProjects(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Git Bash rewrites Win32 paths; POSIX behavior runs on Linux and macOS")
	}
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
	powerShell, err := Hook("powershell", false)
	if err != nil || !strings.Contains(powerShell, "function global:prompt") || !strings.Contains(powerShell, "CuratorPromptWrapped") {
		t.Fatalf("PowerShell hook does not install an idempotent prompt wrapper: %v", err)
	}
}

func TestPowerShellHookRunsOnEveryPrompt(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell prompt integration is exercised on Windows")
	}
	powerShell, err := exec.LookPath("pwsh")
	if err != nil {
		t.Skip("pwsh is unavailable")
	}
	root := t.TempDir()
	project := filepath.Join(root, "project")
	nested := filepath.Join(project, "nested")
	outside := filepath.Join(root, "outside")
	for _, dir := range []string{nested, outside} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := envfiles.WriteProject(project); err != nil {
		t.Fatal(err)
	}
	hook, err := Hook("powershell", false)
	if err != nil {
		t.Fatal(err)
	}
	hookPath := filepath.Join(root, "hook.ps1")
	if err := os.WriteFile(hookPath, []byte(hook), 0o600); err != nil {
		t.Fatal(err)
	}
	scriptPath := filepath.Join(root, "test.ps1")
	script := `
$originalPath = $env:PATH
function global:prompt { return "ORIGINAL>" }
. $env:HOOK_PATH
Set-Location $env:NESTED
$firstPrompt = prompt
Write-Output "first=$($env:CSK_PROJECT_ROOT):$firstPrompt"
Set-Location $env:OUTSIDE
$secondPrompt = prompt
$restored = if ($env:PATH -eq $originalPath) { "restored" } else { "changed" }
$active = if ($env:CURATOR_ACTIVE_ENV) { $env:CURATOR_ACTIVE_ENV } else { "unset" }
Write-Output ("left={0}:{1}:{2}" -f $active, $restored, $secondPrompt)
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(powerShell, "-NoProfile", "-File", scriptPath)
	command.Env = append(os.Environ(), "HOOK_PATH="+hookPath, "NESTED="+nested, "OUTSIDE="+outside)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("PowerShell hook execution: %v\n%s", err, output)
	}
	text := string(output)
	firstLine := ""
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "first=") {
			firstLine = strings.TrimSpace(line)
			break
		}
	}
	activeRoot := strings.TrimSuffix(strings.TrimPrefix(firstLine, "first="), ":ORIGINAL>")
	projectInfo, projectErr := os.Stat(project)
	activeInfo, activeErr := os.Stat(activeRoot)
	if projectErr != nil || activeErr != nil || !os.SameFile(projectInfo, activeInfo) {
		t.Fatalf("PowerShell hook activated %q instead of %q:\n%s", activeRoot, project, text)
	}
	if !strings.Contains(text, "left=unset:restored:ORIGINAL>") {
		t.Fatalf("PowerShell hook did not restore the original environment and prompt:\n%s", text)
	}
}
