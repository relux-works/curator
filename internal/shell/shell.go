// Package shell prints the shell hooks of Spec §14.1: upward search for the
// project env file with PATH save and restore, plus optional global env
// sourcing.
package shell

import "fmt"

// Hook returns the hook code for a shell: zsh, bash, or powershell.
func Hook(shellName string, includeGlobal bool) (string, error) {
	switch shellName {
	case "zsh", "bash":
		return posixHook(includeGlobal), nil
	case "powershell":
		return powershellHook(includeGlobal), nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", shellName)
	}
}

func posixHook(includeGlobal bool) string {
	globalPart := ""
	sourceGlobal := ""
	if includeGlobal {
		globalPart = `
_curator_global_env_file() {
  local cfg="${CURATOR_CONFIG:-$HOME/.curator/config.json}"
  local home_dir
  home_dir="$(dirname "$cfg")"
  if [ -f "$home_dir/global/env.sh" ]; then
    printf '%s/global/env.sh\n' "$home_dir"
  fi
}

_curator_source_global_env() {
  local global_env
  global_env="$(_curator_global_env_file 2>/dev/null || true)"
  if [ -n "$global_env" ] && [ "$CURATOR_ACTIVE_GLOBAL_ENV" != "$global_env" ]; then
    . "$global_env"
    CURATOR_ACTIVE_GLOBAL_ENV="$global_env"
    export CURATOR_ACTIVE_GLOBAL_ENV
  fi
}
`
		sourceGlobal = "  _curator_source_global_env\n"
	}
	return `# Curator shell hook
` + globalPart + `
_curator_find_env() {
  local dir="$PWD"
  while [ "$dir" != "/" ]; do
    if [ -f "$dir/.agents/env.sh" ]; then
      printf '%s/.agents/env.sh\n' "$dir"
      return 0
    fi
    dir="$(dirname "$dir")"
  done
  return 1
}

_curator_auto_env() {
` + sourceGlobal + `  local env_file
  env_file="$(_curator_find_env 2>/dev/null || true)"
  if [ -n "$CURATOR_ACTIVE_ENV" ] && [ "$CURATOR_ACTIVE_ENV" != "$env_file" ]; then
    PATH="$CURATOR_OLD_PATH"
    export PATH
    unset CURATOR_ACTIVE_ENV
    unset CURATOR_OLD_PATH
  fi
  if [ -n "$env_file" ] && [ "$CURATOR_ACTIVE_ENV" != "$env_file" ]; then
    CURATOR_OLD_PATH="$PATH"
    export CURATOR_OLD_PATH
    . "$env_file"
    CURATOR_ACTIVE_ENV="$env_file"
    export CURATOR_ACTIVE_ENV
  fi
}

case "$SHELL" in
  *zsh*) autoload -Uz add-zsh-hook 2>/dev/null || true; add-zsh-hook chpwd _curator_auto_env 2>/dev/null || true ;;
esac
PROMPT_COMMAND="_curator_auto_env${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
_curator_auto_env
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
` + sourceGlobal + `  $dir = Get-Location
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
    . $envFile
    $env:CURATOR_ACTIVE_ENV = $envFile
  }
}
Invoke-CuratorAutoEnv
`
}
