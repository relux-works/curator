package shell

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/envfiles"
)

func TestDetectIsCrossPlatform(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		goos     string
		expected string
	}{
		{name: "zsh", env: map[string]string{"SHELL": "/bin/zsh"}, goos: "darwin", expected: "zsh"},
		{name: "bash", env: map[string]string{"SHELL": "/usr/local/bin/bash"}, goos: "linux", expected: "bash"},
		{name: "git bash", env: map[string]string{"SHELL": `C:\Program Files\Git\bin\bash.exe`}, goos: "windows", expected: "bash"},
		{name: "windows", env: map[string]string{}, goos: "windows", expected: "powershell"},
		{name: "PowerShell environment", env: map[string]string{"PSModulePath": "modules"}, goos: "linux", expected: "powershell"},
		{name: "portable fallback", env: map[string]string{}, goos: "linux", expected: "bash"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := Detect(test.env, test.goos); actual != test.expected {
				t.Fatalf("Detect(%v, %q) = %q, want %q", test.env, test.goos, actual, test.expected)
			}
		})
	}
}

func TestPosixHookEntersNestedSwitchesAndLeavesProjects(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("native path comparison is covered by PowerShell on Windows")
	}
	for _, shellName := range []string{"bash", "zsh"} {
		t.Run(shellName, func(t *testing.T) {
			executable, err := exec.LookPath(shellName)
			if err != nil {
				t.Skip(shellName + " is unavailable")
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
			hookPath := writeHook(t, shellName, false)
			script := `set -e
original="$PATH"
cd "$FIRST/nested"
. "$HOOK_PATH"
printf 'first=%s\n' "$CSK_PROJECT_ROOT"
cd "$SECOND"
_curator_auto_env
printf 'second=%s\n' "$CSK_PROJECT_ROOT"
cd "$OUTSIDE"
_curator_auto_env
printf 'left=%s:%s\n' "${CURATOR_ACTIVE_ENV-unset}" "$([ "$PATH" = "$original" ] && printf restored || printf changed)"
`
			output := runPosix(t, executable, script, map[string]string{
				"HOOK_PATH": hookPath, "FIRST": first, "SECOND": second, "OUTSIDE": outside,
			})
			for _, expected := range []string{"first=" + first, "second=" + second, "left=unset:restored"} {
				if !strings.Contains(output, expected) {
					t.Fatalf("hook output lacks %q:\n%s", expected, output)
				}
			}
		})
	}
}

func TestPosixHookRejectsNonAbsolutePWDWithoutHanging(t *testing.T) {
	for _, shellName := range []string{"bash", "zsh"} {
		t.Run(shellName, func(t *testing.T) {
			executable, err := exec.LookPath(shellName)
			if err != nil {
				t.Skip(shellName + " is unavailable")
			}
			hookPath := writeHook(t, shellName, false)
			for _, broken := range []string{"", ".", "relative/path"} {
				t.Run(strings.ReplaceAll(broken, "/", "-"), func(t *testing.T) {
					script := `PWD="$BROKEN_PWD"
export PWD
. "$HOOK_PATH"
printf 'completed\n'
`
					output := runPosix(t, executable, script, map[string]string{
						"BROKEN_PWD": broken, "HOOK_PATH": hookPath,
					})
					if output != "completed\n" {
						t.Fatalf("unexpected output for PWD %q: %q", broken, output)
					}
				})
			}
		})
	}
}

func TestPosixHookSupportsNounsetProfiles(t *testing.T) {
	for _, shellName := range []string{"bash", "zsh"} {
		t.Run(shellName, func(t *testing.T) {
			executable, err := exec.LookPath(shellName)
			if err != nil {
				t.Skip(shellName + " is unavailable")
			}
			hookPath := writeHook(t, shellName, true)
			output := runPosix(t, executable,
				`set -u; . "$HOOK_PATH"; printf 'completed\n'`,
				map[string]string{"HOOK_PATH": hookPath},
			)
			if output != "completed\n" {
				t.Fatalf("nounset profile output: %q", output)
			}
		})
	}
}

func TestBashHookDoesNotDuplicatePromptCommand(t *testing.T) {
	executable, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is unavailable")
	}
	hookPath := writeHook(t, "bash", false)
	output := runPosix(t, executable,
		`PROMPT_COMMAND="existing"; . "$HOOK_PATH"; . "$HOOK_PATH"; printf '%s\n' "$PROMPT_COMMAND"`,
		map[string]string{"HOOK_PATH": hookPath},
	)
	if output != "_curator_auto_env;existing\n" {
		t.Fatalf("duplicated PROMPT_COMMAND: %q", output)
	}
}

func TestBashHookPreservesPromptCommandArray(t *testing.T) {
	executable, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is unavailable")
	}
	hookPath := writeHook(t, "bash", false)
	output := runPosix(t, executable,
		`PROMPT_COMMAND=("first" "second"); . "$HOOK_PATH"; . "$HOOK_PATH"; printf '%s|%s|%s|%s\n' "${#PROMPT_COMMAND[@]}" "${PROMPT_COMMAND[0]}" "${PROMPT_COMMAND[1]}" "${PROMPT_COMMAND[2]}"`,
		map[string]string{"HOOK_PATH": hookPath},
	)
	if output != "3|_curator_auto_env|first|second\n" {
		t.Fatalf("PROMPT_COMMAND array was not preserved: %q", output)
	}
}

func TestZshHookIsIdempotentAndDoesNotReenter(t *testing.T) {
	executable, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh is unavailable")
	}
	project := t.TempDir()
	nested := filepath.Join(project, "nested")
	agentsDir := filepath.Join(project, ".agents")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	env := "CSK_SOURCE_COUNT=$(( ${CSK_SOURCE_COUNT:-0} + 1 ))\n" +
		"_curator_auto_env\n" +
		"export CSK_PROJECT_ROOT=" + shellQuote(project) + "\n"
	if err := os.WriteFile(filepath.Join(agentsDir, "env.sh"), []byte(env), 0o600); err != nil {
		t.Fatal(err)
	}
	hookPath := writeHook(t, "zsh", false)
	script := `cd "$NESTED"
. "$HOOK_PATH"
. "$HOOK_PATH"
chpwd_count=0
for fn in ${chpwd_functions[@]}; do
  [ "$fn" = "_curator_auto_env" ] && chpwd_count=$((chpwd_count + 1))
done
precmd_count=0
for fn in ${precmd_functions[@]}; do
  [ "$fn" = "_curator_auto_env" ] && precmd_count=$((precmd_count + 1))
done
printf 'sources=%s chpwd=%s precmd=%s\n' "$CSK_SOURCE_COUNT" "$chpwd_count" "$precmd_count"
`
	output := runPosix(t, executable, script, map[string]string{
		"HOOK_PATH": hookPath, "NESTED": nested,
	})
	if output != "sources=1 chpwd=1 precmd=1\n" {
		t.Fatalf("zsh hook is not reentrancy-safe and idempotent: %q", output)
	}
}

func TestPosixHookDoesNotReenterWhileSourcingGlobalEnv(t *testing.T) {
	for _, shellName := range []string{"bash", "zsh"} {
		t.Run(shellName, func(t *testing.T) {
			executable, err := exec.LookPath(shellName)
			if err != nil {
				t.Skip(shellName + " is unavailable")
			}
			managerHome := t.TempDir()
			globalDir := filepath.Join(managerHome, "global")
			if err := os.MkdirAll(globalDir, 0o755); err != nil {
				t.Fatal(err)
			}
			payload := "CURATOR_GLOBAL_SOURCE_COUNT=$(( ${CURATOR_GLOBAL_SOURCE_COUNT:-0} + 1 ))\n" +
				"_curator_auto_env\n"
			if err := os.WriteFile(filepath.Join(globalDir, "env.sh"), []byte(payload), 0o600); err != nil {
				t.Fatal(err)
			}
			hookPath := writeHook(t, shellName, true)
			output := runPosix(t, executable,
				`. "$HOOK_PATH"; printf 'sources=%s\n' "$CURATOR_GLOBAL_SOURCE_COUNT"`,
				map[string]string{
					"HOOK_PATH":      hookPath,
					"CURATOR_CONFIG": filepath.Join(managerHome, "config.json"),
				},
			)
			if output != "sources=1\n" {
				t.Fatalf("global environment re-entered: %q", output)
			}
		})
	}
}

func TestPosixHookNormalizesGitBashConfigAndAvoidsDirname(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("native Git Bash paths are exercised by the Windows CI hook shape")
	}
	executable, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is unavailable")
	}
	root := t.TempDir()
	globalDir := filepath.Join(root, "C:", "manager", "global")
	if err := os.MkdirAll(globalDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(globalDir, "env.sh"), []byte("CURATOR_GLOBAL_TEST=loaded\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	hookPath := writeHook(t, "bash", true)
	payload, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), "dirname") {
		t.Fatalf("hook starts an external dirname process:\n%s", payload)
	}
	output := runPosix(t, executable,
		`cd "$ROOT"; . "$HOOK_PATH"; printf '%s\n' "$CURATOR_GLOBAL_TEST"`,
		map[string]string{
			"ROOT": root, "HOOK_PATH": hookPath, "CURATOR_CONFIG": `C:\manager\config.json`,
		},
	)
	if output != "loaded\n" {
		t.Fatalf("Git Bash drive path was not normalized: %q", output)
	}
}

func TestInstallHookAndSourceCommand(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home with ' quote")
	target, err := InstallHook("bash", home, true)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(target) != "curator.bash" {
		t.Fatalf("hook target = %q", target)
	}
	payload, err := os.ReadFile(target)
	if err != nil || !strings.Contains(string(payload), "_curator_source_global_env") {
		t.Fatalf("cached hook: %v\n%s", err, payload)
	}
	replaced, err := InstallHook("bash", home, false)
	if err != nil || replaced != target {
		t.Fatalf("replace hook = %q, %v", replaced, err)
	}
	payload, err = os.ReadFile(target)
	if err != nil || strings.Contains(string(payload), "_curator_source_global_env") {
		t.Fatalf("cached hook was not replaced: %v\n%s", err, payload)
	}
	command, err := SourceCommand("bash", target)
	if err != nil || !strings.Contains(command, `'"'"'`) {
		t.Fatalf("POSIX source command = %q, %v", command, err)
	}
	gitBash, err := SourceCommand("bash", `C:\User's Files\curator.bash`)
	if err != nil || gitBash != `. 'C:/User'"'"'s Files/curator.bash'` {
		t.Fatalf("Git Bash source command = %q, %v", gitBash, err)
	}
	powerShell, err := SourceCommand("powershell", `C:\User's Files\curator.ps1`)
	if err != nil || powerShell != `. 'C:\User''s Files\curator.ps1'` {
		t.Fatalf("PowerShell source command = %q, %v", powerShell, err)
	}
}

func TestHookVariants(t *testing.T) {
	for _, shellName := range []string{"zsh", "bash", "powershell"} {
		hook, err := Hook(shellName, true)
		if err != nil || hook == "" || !strings.Contains(hook, "global") {
			t.Fatalf("Hook(%s) = %q, %v", shellName, hook, err)
		}
		if !strings.Contains(hook, "CURATOR_AUTO_ENV") {
			t.Fatalf("%s hook cannot disable project scanning", shellName)
		}
	}
	zsh, _ := Hook("zsh", false)
	if !strings.Contains(zsh, "add-zsh-hook precmd") || strings.Contains(zsh, "PROMPT_COMMAND=") {
		t.Fatalf("zsh integration shape is wrong:\n%s", zsh)
	}
	bash, _ := Hook("bash", false)
	if !strings.Contains(bash, "PROMPT_COMMAND=") || strings.Contains(bash, "add-zsh-hook") {
		t.Fatalf("bash integration shape is wrong:\n%s", bash)
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
		powerShell, err = exec.LookPath("powershell")
	}
	if err != nil {
		t.Skip("PowerShell is unavailable")
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
		t.Fatalf("hook execution: %v\n%s", err, output)
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

func writeHook(t *testing.T, shellName string, includeGlobal bool) string {
	t.Helper()
	hook, err := Hook(shellName, includeGlobal)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "hook")
	if err := os.WriteFile(path, []byte(hook), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func runPosix(t *testing.T, executable, script string, environment map[string]string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	args := []string{"--noprofile", "--norc", "-c", script}
	if filepath.Base(executable) == "zsh" {
		args = []string{"-dfc", script}
	}
	command := exec.CommandContext(ctx, executable, args...)
	command.Env = append([]string{}, os.Environ()...)
	command.Env = append(command.Env, "SHELL="+executable)
	for name, value := range environment {
		command.Env = append(command.Env, name+"="+value)
	}
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("%s hook timed out: %v\n%s", filepath.Base(executable), ctx.Err(), output)
	}
	if err != nil {
		t.Fatalf("%s hook execution: %v\n%s", filepath.Base(executable), err, output)
	}
	return string(output)
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}
