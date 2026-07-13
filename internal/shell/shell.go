// Package shell generates and caches the optional shell hooks of Spec §8:
// finite upward project search, PATH save and restore, and global activation.
package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var hookFilenames = map[string]string{
	"zsh":        "curator.zsh",
	"bash":       "curator.bash",
	"powershell": "curator.ps1",
}

// Detect returns the best supported shell for an environment and platform.
// A nil environment reads the current process environment; an empty goos uses
// runtime.GOOS. SHELL wins on Windows so Git Bash keeps POSIX integration.
func Detect(environment map[string]string, goos string) string {
	getenv := func(name string) string {
		if environment == nil {
			return os.Getenv(name)
		}
		return environment[name]
	}
	configured := strings.ReplaceAll(strings.TrimSpace(getenv("SHELL")), `\`, "/")
	if index := strings.LastIndex(configured, "/"); index >= 0 {
		configured = configured[index+1:]
	}
	configured = strings.TrimSuffix(strings.ToLower(configured), ".exe")
	if configured == "zsh" || configured == "bash" {
		return configured
	}
	if goos == "" {
		goos = runtime.GOOS
	}
	if goos == "windows" || getenv("PSModulePath") != "" {
		return "powershell"
	}
	return "bash"
}

// Hook returns hook code for zsh, bash, or PowerShell.
func Hook(shellName string, includeGlobal bool) (string, error) {
	switch shellName {
	case "zsh", "bash":
		return posixHook(shellName, includeGlobal), nil
	case "powershell":
		return powershellHook(includeGlobal), nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", shellName)
	}
}

// InstallHook atomically caches one generated hook below the manager home.
func InstallHook(shellName, home string, includeGlobal bool) (string, error) {
	filename, known := hookFilenames[shellName]
	if !known {
		return "", fmt.Errorf("unsupported shell: %s", shellName)
	}
	payload, err := Hook(shellName, includeGlobal)
	if err != nil {
		return "", err
	}
	hooksDir := filepath.Join(home, "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return "", err
	}
	temporary, err := os.CreateTemp(hooksDir, "."+filename+".*.tmp")
	if err != nil {
		return "", err
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if _, err := temporary.WriteString(payload); err != nil {
		_ = temporary.Close()
		return "", err
	}
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return "", err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return "", err
	}
	if err := temporary.Close(); err != nil {
		return "", err
	}
	target := filepath.Join(hooksDir, filename)
	if err := os.Rename(temporaryPath, target); err != nil {
		if removeErr := os.Remove(target); removeErr != nil && !os.IsNotExist(removeErr) {
			return "", err
		}
		if err := os.Rename(temporaryPath, target); err != nil {
			return "", err
		}
	}
	return target, nil
}

// SourceCommand returns the profile line that sources a cached hook.
func SourceCommand(shellName, hookPath string) (string, error) {
	switch shellName {
	case "zsh", "bash":
		hookPath = posixReadableWindowsPath(hookPath)
		return ". '" + strings.ReplaceAll(hookPath, "'", `'"'"'`) + "'", nil
	case "powershell":
		return ". '" + strings.ReplaceAll(hookPath, "'", "''") + "'", nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", shellName)
	}
}

func posixReadableWindowsPath(path string) string {
	drivePath := len(path) >= 3 &&
		((path[0] >= 'A' && path[0] <= 'Z') || (path[0] >= 'a' && path[0] <= 'z')) &&
		path[1] == ':' && (path[2] == '\\' || path[2] == '/')
	if drivePath || strings.HasPrefix(path, `\\`) {
		return strings.ReplaceAll(path, `\`, "/")
	}
	return path
}

func posixHook(shellName string, includeGlobal bool) string {
	globalPart := ""
	sourceGlobal := ""
	if includeGlobal {
		globalPart = `
_curator_global_env_file() {
  local cfg="${CURATOR_CONFIG:-$HOME/.curator/config.json}"
  local home_dir
  case "$cfg" in
    [A-Za-z]:\\*|[A-Za-z]:/*|\\\\*) cfg="${cfg//\\//}" ;;
  esac
  home_dir="${cfg%/*}"
  if [ "$home_dir" = "$cfg" ]; then
    home_dir="."
  elif [ -z "$home_dir" ]; then
    home_dir="/"
  fi
  if [ -f "$home_dir/global/env.sh" ]; then
    printf '%s/global/env.sh\n' "$home_dir"
  fi
}

_curator_source_global_env() {
  local global_env
  global_env="$(_curator_global_env_file 2>/dev/null || true)"
  if [ -n "$global_env" ] && [ "${CURATOR_ACTIVE_GLOBAL_ENV:-}" != "$global_env" ]; then
    CURATOR_ACTIVE_GLOBAL_ENV="$global_env"
    export CURATOR_ACTIVE_GLOBAL_ENV
    . "$global_env"
  fi
}
`
		sourceGlobal = "  _curator_source_global_env\n"
	}

	integration := `case "$(declare -p PROMPT_COMMAND 2>/dev/null)" in
  declare\ -a*)
    _curator_prompt_present=0
    for _curator_prompt_entry in "${PROMPT_COMMAND[@]}"; do
      if [ "$_curator_prompt_entry" = "_curator_auto_env" ]; then
        _curator_prompt_present=1
        break
      fi
    done
    if [ "$_curator_prompt_present" = "0" ]; then
      PROMPT_COMMAND=("_curator_auto_env" "${PROMPT_COMMAND[@]}")
    fi
    unset _curator_prompt_present
    unset _curator_prompt_entry
    ;;
  *)
    case ";${PROMPT_COMMAND:-};" in
      *";_curator_auto_env;"*) ;;
      *) PROMPT_COMMAND="_curator_auto_env${PROMPT_COMMAND:+;$PROMPT_COMMAND}" ;;
    esac
    ;;
esac
`
	if shellName == "zsh" {
		integration = `autoload -Uz add-zsh-hook 2>/dev/null || true
add-zsh-hook -d precmd _curator_auto_env 2>/dev/null || true
add-zsh-hook -d chpwd _curator_auto_env 2>/dev/null || true
add-zsh-hook precmd _curator_auto_env 2>/dev/null || true
add-zsh-hook chpwd _curator_auto_env 2>/dev/null || true
`
	}

	return `# Curator shell hook
` + globalPart + `
_curator_find_env() {
  local dir="${PWD:-}"
  case "$dir" in
    /*) ;;
    *) return 1 ;;
  esac
  while :; do
    if [ -f "$dir/.agents/env.sh" ]; then
      printf '%s/.agents/env.sh\n' "$dir"
      return 0
    fi
    if [ "$dir" = "/" ]; then
      break
    fi
    dir="${dir%/*}"
    if [ -z "$dir" ]; then
      dir="/"
    fi
  done
  return 1
}

_curator_auto_env() {
  local env_file
` + sourceGlobal + `  if [ "${CURATOR_AUTO_ENV:-1}" = "0" ]; then
    if [ -n "${CURATOR_ACTIVE_ENV:-}" ]; then
      PATH="${CURATOR_OLD_PATH:-$PATH}"
      export PATH
      unset CURATOR_ACTIVE_ENV
      unset CURATOR_OLD_PATH
    fi
    return 0
  fi
  env_file="$(_curator_find_env 2>/dev/null || true)"
  if [ -n "${CURATOR_ACTIVE_ENV:-}" ] && [ "$CURATOR_ACTIVE_ENV" != "$env_file" ]; then
    PATH="${CURATOR_OLD_PATH:-$PATH}"
    export PATH
    unset CURATOR_ACTIVE_ENV
    unset CURATOR_OLD_PATH
  fi
  if [ -n "$env_file" ] && [ "${CURATOR_ACTIVE_ENV:-}" != "$env_file" ]; then
    CURATOR_OLD_PATH="$PATH"
    export CURATOR_OLD_PATH
    # Mark the environment before sourcing. zsh can run chpwd hooks for a cd
    # used by env.sh while it resolves the project root.
    CURATOR_ACTIVE_ENV="$env_file"
    export CURATOR_ACTIVE_ENV
    . "$env_file"
  fi
}

` + integration + `_curator_auto_env
`
}

func powershellHook(includeGlobal bool) string {
	globalPart := ""
	sourceGlobal := ""
	if includeGlobal {
		globalPart = `
function Get-CuratorGlobalEnvFile {
  $cfg = if ($env:CURATOR_CONFIG) { $env:CURATOR_CONFIG } else { Join-Path $HOME ".curator/config.json" }
  $homeDir = Split-Path -Parent $cfg
  $candidate = Join-Path $homeDir "global/env.ps1"
  if (Test-Path $candidate) { return $candidate }
  return $null
}

function Invoke-CuratorGlobalEnv {
  $globalEnv = Get-CuratorGlobalEnvFile
  if ($globalEnv -and $env:CURATOR_ACTIVE_GLOBAL_ENV -ne $globalEnv) {
    . $globalEnv
    $env:CURATOR_ACTIVE_GLOBAL_ENV = $globalEnv
  }
}
`
		sourceGlobal = "  Invoke-CuratorGlobalEnv\n"
	}
	return `# Curator shell hook
` + globalPart + `
function Invoke-CuratorAutoEnv {
` + sourceGlobal + `  if ($env:CURATOR_AUTO_ENV -eq "0") {
    if ($env:CURATOR_ACTIVE_ENV) {
      $env:PATH = $env:CURATOR_OLD_PATH
      Remove-Item Env:\CURATOR_ACTIVE_ENV -ErrorAction SilentlyContinue
      Remove-Item Env:\CURATOR_OLD_PATH -ErrorAction SilentlyContinue
    }
    return
  }
  $dir = Get-Location
  $envFile = $null
  while ($dir) {
    $candidate = Join-Path $dir ".agents/env.ps1"
    if (Test-Path $candidate) { $envFile = $candidate; break }
    $parent = Split-Path -Parent $dir
    if ($parent -eq $dir) { break }
    $dir = $parent
  }
  if ($env:CURATOR_ACTIVE_ENV -and $env:CURATOR_ACTIVE_ENV -ne $envFile) {
    $env:PATH = $env:CURATOR_OLD_PATH
    Remove-Item Env:\CURATOR_ACTIVE_ENV -ErrorAction SilentlyContinue
    Remove-Item Env:\CURATOR_OLD_PATH -ErrorAction SilentlyContinue
  }
  if ($envFile -and $env:CURATOR_ACTIVE_ENV -ne $envFile) {
    $env:CURATOR_OLD_PATH = $env:PATH
    $env:CURATOR_ACTIVE_ENV = $envFile
    . $envFile
  }
}
if (-not $global:CuratorPromptWrapped) {
  $global:CuratorOriginalPrompt = (Get-Item Function:prompt -ErrorAction SilentlyContinue).ScriptBlock
  function global:prompt {
    Invoke-CuratorAutoEnv
    if ($global:CuratorOriginalPrompt) {
      return & $global:CuratorOriginalPrompt
    }
    return "PS $($executionContext.SessionState.Path.CurrentLocation)> "
  }
  $global:CuratorPromptWrapped = $true
}
Invoke-CuratorAutoEnv
`
}
