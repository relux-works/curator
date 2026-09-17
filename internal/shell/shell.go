// Package shell generates and caches the optional shell hooks of Spec §8:
// finite upward project search, PATH save and restore, and global activation.
package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/relux-works/curator/internal/hookapproval"
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

// TrustProfile selects the §8.5 rollout behavior baked into generated hook
// code. The shipped default is the warning revision; the enforcing revision
// is a later release, never this change.
type TrustProfile string

// Closed rollout profiles of Spec §8.5.
const (
	TrustProfileAWarning   TrustProfile = "A-warning"
	TrustProfileBEnforcing TrustProfile = "B-enforcing"
	// DefaultTrustProfile is the revision this release ships.
	DefaultTrustProfile = TrustProfileAWarning
)

// Closed shell-hook trust diagnostics of Spec §8.4.
const (
	DiagnosticEnvUnapproved = "shell_hook_env_unapproved"
	DiagnosticEnvChanged    = "shell_hook_env_changed"
)

// Hook returns hook code for zsh, bash, or PowerShell, generated for the
// default (warning) trust profile.
func Hook(shellName string, includeGlobal bool) (string, error) {
	return HookWithProfile(shellName, includeGlobal, DefaultTrustProfile)
}

// HookWithProfile returns hook code generated for one rollout profile.
func HookWithProfile(shellName string, includeGlobal bool, profile TrustProfile) (string, error) {
	switch profile {
	case TrustProfileAWarning, TrustProfileBEnforcing:
	default:
		return "", fmt.Errorf("unsupported trust profile: %s", profile)
	}
	switch shellName {
	case "zsh", "bash":
		return posixHook(shellName, includeGlobal, profile), nil
	case "powershell":
		return powershellHook(includeGlobal, profile), nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", shellName)
	}
}

// InstallHook atomically caches one generated hook below the manager home.
func InstallHook(shellName, home string, includeGlobal bool) (string, error) {
	return InstallHookWithProfile(shellName, home, includeGlobal, DefaultTrustProfile)
}

// InstallHookWithProfile caches one generated hook for one rollout profile.
func InstallHookWithProfile(shellName, home string, includeGlobal bool, profile TrustProfile) (string, error) {
	filename, known := hookFilenames[shellName]
	if !known {
		return "", fmt.Errorf("unsupported shell: %s", shellName)
	}
	payload, err := HookWithProfile(shellName, includeGlobal, profile)
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

func posixHook(shellName string, includeGlobal bool, profile TrustProfile) string {
	globalPart := ""
	sourceGlobal := ""
	if includeGlobal {
		globalPart = `
_curator_global_env_file() {
  local cfg="${CURATOR_CONFIG:-$HOME/.curator/config.json}"
  local home_dir
  case "$cfg" in
    [A-Za-z]:\\*|[A-Za-z]:/*|\\\\*) cfg="$(printf '%s' "$cfg" | tr '\\' '/')" ;;
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

	// The bash prompt integration stays strictly POSIX-parseable (sh, dash,
	// Git Bash): the array-aware branch runs only under bash, quoted inside
	// eval so no other interpreter parses the array syntax. Other shells take
	// the plain string branch, which is inert for them.
	integration := `if [ -n "${BASH_VERSION:-}" ]; then
  eval '
    case "$(declare -p PROMPT_COMMAND 2>/dev/null)" in
      declare\ -a*)
        _curator_prompt_present=0
        _curator_nounset=0
        case "$-" in
          *u*) _curator_nounset=1 ;;
        esac
        set +u
        for _curator_prompt_entry in "${PROMPT_COMMAND[@]}"; do
          if [ "$_curator_prompt_entry" = "_curator_auto_env" ]; then
            _curator_prompt_present=1
            break
          fi
        done
        if [ "$_curator_prompt_present" = "0" ]; then
          PROMPT_COMMAND=("_curator_auto_env" "${PROMPT_COMMAND[@]}")
        fi
        if [ "$_curator_nounset" = "1" ]; then
          set -u
        fi
        ;;
      *)
        case ";${PROMPT_COMMAND:-};" in
          *";_curator_auto_env;"*) ;;
          *) PROMPT_COMMAND="_curator_auto_env${PROMPT_COMMAND:+;$PROMPT_COMMAND}" ;;
        esac
        ;;
    esac
    unset _curator_prompt_present _curator_prompt_entry _curator_nounset'
else
  case ";${PROMPT_COMMAND:-};" in
    *";_curator_auto_env;"*) ;;
    *) PROMPT_COMMAND="_curator_auto_env${PROMPT_COMMAND:+;$PROMPT_COMMAND}" ;;
  esac
fi
`
	if shellName == "zsh" {
		integration = `autoload -Uz add-zsh-hook 2>/dev/null || true
add-zsh-hook -d precmd _curator_auto_env 2>/dev/null || true
add-zsh-hook -d chpwd _curator_auto_env 2>/dev/null || true
add-zsh-hook precmd _curator_auto_env 2>/dev/null || true
add-zsh-hook chpwd _curator_auto_env 2>/dev/null || true
`
	}

	trustPart := `
_curator_hook_approvals_file() {
  local cfg="${CURATOR_CONFIG:-$HOME/.curator/config.json}"
  local home_dir
  case "$cfg" in
    [A-Za-z]:\\*|[A-Za-z]:/*|\\\\*) cfg="$(printf '%s' "$cfg" | tr '\\' '/')" ;;
  esac
  home_dir="${cfg%/*}"
  if [ "$home_dir" = "$cfg" ]; then
    home_dir="."
  elif [ -z "$home_dir" ]; then
    home_dir="/"
  fi
  printf '%s/hook-approvals.tsv\n' "$home_dir"
}

_curator_trust_canonical() {
  local _curator_tc_path _curator_tc_dir _curator_tc_base _curator_tc_phys _curator_tc_link _curator_tc_n
  _curator_tc_path="$1"
  case "$_curator_tc_path" in
    /*) ;;
    *) _curator_tc_path="${PWD:-/}/$_curator_tc_path" ;;
  esac
  _curator_tc_dir="${_curator_tc_path%/*}"
  _curator_tc_base="${_curator_tc_path##*/}"
  if [ -z "$_curator_tc_dir" ]; then
    _curator_tc_dir="/"
  fi
  if _curator_tc_phys="$(cd -P "$_curator_tc_dir" 2>/dev/null && pwd -P)"; then
    if [ "$_curator_tc_phys" = "/" ]; then
      _curator_tc_path="/$_curator_tc_base"
    else
      _curator_tc_path="$_curator_tc_phys/$_curator_tc_base"
    fi
  fi
  if command -v readlink >/dev/null 2>&1; then
    _curator_tc_n=0
    while [ "$_curator_tc_n" -lt 32 ] && [ -L "$_curator_tc_path" ]; do
      _curator_tc_link="$(readlink "$_curator_tc_path" 2>/dev/null)" || break
      case "$_curator_tc_link" in
        /*) _curator_tc_path="$_curator_tc_link" ;;
        *) _curator_tc_path="${_curator_tc_path%/*}/$_curator_tc_link" ;;
      esac
      _curator_tc_dir="${_curator_tc_path%/*}"
      _curator_tc_base="${_curator_tc_path##*/}"
      if [ -z "$_curator_tc_dir" ]; then
        _curator_tc_dir="/"
      fi
      if _curator_tc_phys="$(cd -P "$_curator_tc_dir" 2>/dev/null && pwd -P)"; then
        if [ "$_curator_tc_phys" = "/" ]; then
          _curator_tc_path="/$_curator_tc_base"
        else
          _curator_tc_path="$_curator_tc_phys/$_curator_tc_base"
        fi
      fi
      _curator_tc_n=$((_curator_tc_n + 1))
    done
  fi
  printf '%s\n' "$_curator_tc_path"
}

_curator_trust_is_windows_shell() {
  case "$(uname -s 2>/dev/null)" in
    MINGW*|MSYS*|CYGWIN*) return 0 ;;
    *) return 1 ;;
  esac
}

# Single comparison identity of hookapproval.Canonicalize: under MSYS/Git
# Bash/Cygwin the resolved MSYS spelling is mapped to the native Windows
# spelling (C:\..., uppercase drive) via cygpath -w; elsewhere the resolved
# spelling is the identity. Prints the identity; fails when cygpath is
# absent, in which case the caller warns and refuses under both profiles,
# never sourcing silently.
_curator_trust_identity() {
  local _curator_ti_native _curator_ti_drive _curator_ti_rest
  if _curator_trust_is_windows_shell; then
    if ! command -v cygpath >/dev/null 2>&1; then
      return 1
    fi
    _curator_ti_native="$(cygpath -w "$1" 2>/dev/null | tr -d '\r')" || return 1
    if [ -z "$_curator_ti_native" ]; then
      return 1
    fi
    case "$_curator_ti_native" in
      [a-z]:*)
        _curator_ti_drive="$(printf '%s' "$_curator_ti_native" | cut -c1 | tr 'a-z' 'A-Z')"
        _curator_ti_rest="$(printf '%s' "$_curator_ti_native" | cut -c2-)"
        _curator_ti_native="$_curator_ti_drive$_curator_ti_rest"
        ;;
    esac
    printf '%s\n' "$_curator_ti_native"
  else
    printf '%s\n' "$1"
  fi
}

_curator_trust_digest() {
  local out digest tool
  for tool in shasum sha256sum openssl; do
    if command -v "$tool" >/dev/null 2>&1; then
      case "$tool" in
        shasum) out="$(shasum -a 256 "$1" 2>/dev/null)" || continue ;;
        sha256sum) out="$(sha256sum "$1" 2>/dev/null)" || continue ;;
        openssl) out="$(openssl dgst -sha256 "$1" 2>/dev/null)" || continue ;;
      esac
      case "$tool" in
        openssl) digest="${out##* }" ;;
        *) digest="${out%% *}" ;;
      esac
      case "$digest" in
        *[!0-9a-f]*|"") continue ;;
      esac
      if [ "${#digest}" -eq __CURATOR_DIGEST_LEN__ ]; then
        printf '%s\n' "$digest"
        return 0
      fi
    fi
  done
  return 1
}

_curator_trust_recorded() {
  local _curator_trust_fold
  if [ ! -e "$2" ]; then
    return 1
  fi
  if [ ! -f "$2" ] || [ ! -r "$2" ] || ! command -v awk >/dev/null 2>&1; then
    return 2
  fi
  _curator_trust_fold=0
  case "$(uname -s 2>/dev/null)" in
    MINGW*|MSYS*|CYGWIN*) _curator_trust_fold=1 ;;
  esac
  awk -F'\t' -v fold="$_curator_trust_fold" 'BEGIN { want = ARGV[1]; ARGV[1] = "" }
function _curator_valid_digest(d) {
  if (length(d) != __CURATOR_DIGEST_LEN__) { return 0 }
  return (d !~ /[^0-9a-f]/)
}
function _curator_valid_ts(ts,  rest, y, mo, d, h, mi, s, oh, om, n, dim) {
  if (ts !~ /__CURATOR_TIMESTAMP_SHAPE__/) { return 0 }
  y = substr(ts, 1, 4) + 0
  mo = substr(ts, 6, 2) + 0
  d = substr(ts, 9, 2) + 0
  h = substr(ts, 12, 2) + 0
  mi = substr(ts, 15, 2) + 0
  s = substr(ts, 18, 2) + 0
  if (mo < 1 || mo > 12 || h > 23 || mi > 59 || s > 59) { return 0 }
  dim = 31
  if (mo == 4 || mo == 6 || mo == 9 || mo == 11) { dim = 30 } else if (mo == 2) {
    dim = 28
    if ((y % 4 == 0 && y % 100 != 0) || y % 400 == 0) { dim = 29 }
  }
  if (d < 1 || d > dim) { return 0 }
  rest = substr(ts, 20)
  if (substr(rest, 1, 1) == ".") {
    rest = substr(rest, 2)
    n = 0
    while (substr(rest, n + 1, 1) ~ /[0-9]/) { n++ }
    rest = substr(rest, n + 1)
  }
  if (rest == "Z") { return 1 }
  if (rest ~ /^[+-][0-9][0-9]:[0-9][0-9]$/) {
    oh = substr(rest, 2, 2) + 0
    om = substr(rest, 5, 2) + 0
    if (oh > 23 || om > 59) { return 0 }
    return 1
  }
  return 0
}
{
  _curator_path_match = ($1 == want)
  if (fold == 1) { _curator_path_match = (tolower($1) == tolower(want)) }
  if (!_curator_path_match) { next }
  if (NF != __CURATOR_RECORD_FIELDS__) { next }
  if (!_curator_valid_digest($2)) { next }
  if ($3 != "__CURATOR_APPROVED_BY_MANAGER__" && $3 != "__CURATOR_APPROVED_BY_OPERATOR__") { next }
  if (!_curator_valid_ts($4)) { next }
  print $2
  found = 1
  exit
}
END { exit !found }' "$1" "$2" 2>/dev/null
}

_curator_trust_warned() {
  case "
${_CURATOR_HOOK_TRUST_WARNED:-}" in
    *"
$1
"*) return 0 ;;
    *) return 1 ;;
  esac
}

_curator_trust_note_warned() {
  _CURATOR_HOOK_TRUST_WARNED="${_CURATOR_HOOK_TRUST_WARNED:-}$1
"
  export _CURATOR_HOOK_TRUST_WARNED
}

_curator_trust_warn() {
  if _curator_trust_warned "$1"; then
    return 0
  fi
  _curator_trust_note_warned "$1"
  if [ "$2" = "shell_hook_env_changed" ]; then
    printf 'curator: %s: %s changed since approval; run curator hook approve %s to approve the new bytes\n' "$2" "$1" "$1" >&2
  else
    printf 'curator: %s: %s is not approved; run curator hook approve %s to approve it\n' "$2" "$1" "$1" >&2
  fi
}

_curator_trust_allow() {
  local approvals digest recorded rc candidate identity
  candidate="$(_curator_trust_canonical "$1")"
  if identity="$(_curator_trust_identity "$candidate" 2>/dev/null)"; then
    candidate="$identity"
  else
    _curator_trust_warn "$candidate" "shell_hook_env_unapproved"
    return 1
  fi
  approvals="$(_curator_hook_approvals_file)"
  if digest="$(_curator_trust_digest "$1" 2>/dev/null)"; then
    :
  else
    digest=""
  fi
  if [ -z "$digest" ]; then
    _curator_trust_warn "$candidate" "shell_hook_env_unapproved"
    return 1
  fi
  if recorded="$(_curator_trust_recorded "$candidate" "$approvals" 2>/dev/null)"; then
    rc=0
  else
    rc=$?
  fi
  if [ "$rc" -eq 2 ]; then
    _curator_trust_warn "$candidate" "shell_hook_env_unapproved"
    return 1
  fi
  if [ -z "$recorded" ]; then
    _curator_trust_warn "$candidate" "shell_hook_env_unapproved"
    if [ "${_CURATOR_TRUST_PROFILE:-__CURATOR_BAKED_PROFILE__}" = "B-enforcing" ]; then
      return 1
    fi
    return 0
  fi
  if [ "$recorded" = "$digest" ]; then
    return 0
  fi
  _curator_trust_warn "$candidate" "shell_hook_env_changed"
  if [ "${_CURATOR_TRUST_PROFILE:-__CURATOR_BAKED_PROFILE__}" = "B-enforcing" ]; then
    return 1
  fi
  return 0
}
`
	return expandTrustTokens(`# Curator shell hook
_CURATOR_TRUST_PROFILE='`+string(profile)+`'
`+globalPart+trustPart+`
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
`+sourceGlobal+`  if [ "${CURATOR_AUTO_ENV:-1}" = "0" ]; then
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
    if _curator_trust_allow "$env_file"; then
      . "$env_file"
    else
      unset CURATOR_ACTIVE_ENV
      unset CURATOR_OLD_PATH
    fi
  fi
}

`+integration+`_curator_auto_env
`, profile)
}

// expandTrustTokens renders the trust-gate placeholders of a generated hook
// from the profile and the hookapproval record grammar, so the emitted POSIX
// and PowerShell consumers validate the same closed record the Go reader
// enforces.
func expandTrustTokens(hook string, profile TrustProfile) string {
	replacer := strings.NewReplacer(
		"__CURATOR_BAKED_PROFILE__", string(profile),
		"__CURATOR_RECORD_FIELDS__", fmt.Sprintf("%d", hookapproval.RecordFieldCount),
		"__CURATOR_DIGEST_LEN__", fmt.Sprintf("%d", hookapproval.SHA256HexLength),
		"__CURATOR_APPROVED_BY_MANAGER__", hookapproval.ApprovedByManager,
		"__CURATOR_APPROVED_BY_OPERATOR__", hookapproval.ApprovedByOperator,
		"__CURATOR_TIMESTAMP_SHAPE__", hookapproval.TimestampShape,
	)
	return replacer.Replace(hook)
}

func powershellHook(includeGlobal bool, profile TrustProfile) string {
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
	trustPart := `
function Get-CuratorHookApprovalsFile {
  $cfg = if ($env:CURATOR_CONFIG) { $env:CURATOR_CONFIG } else { Join-Path $HOME ".curator/config.json" }
  $homeDir = Split-Path -Parent $cfg
  return (Join-Path $homeDir "hook-approvals.tsv")
}

function Get-CuratorTrustDigest {
  param([string]$Path)
  try {
    $hash = (Get-FileHash -Algorithm SHA256 -LiteralPath $Path -ErrorAction Stop).Hash
    if ($hash -and $hash.Length -eq 64) { return $hash.ToLowerInvariant() }
    return $null
  } catch {
    return $null
  }
}

function Get-CuratorTrustCanonical {
  param([string]$Path)
  try {
    $full = [System.IO.Path]::GetFullPath($Path)
  } catch {
    return $Path
  }
  try {
    # Resolve links component by component (EvalSymlinks semantics): a link
    # may sit at any ancestor (an aliased project directory), not only at
    # the final path, and Get-Item on the file alone never exposes those.
    # Each resolution restarts the scan over the expanded absolute path so
    # links inside the target spell out too; the hop bound keeps a link
    # cycle from hanging activation. Unlistable components are kept
    # lexically, matching the Go ancestor fallback.
    $hops = 0
    $changed = $true
    while ($changed -and $hops -lt 32) {
      $changed = $false
      $root = [System.IO.Path]::GetPathRoot($full)
      $parts = @($full.Substring($root.Length) -split '[\\/]+')
      $acc = $root
      for ($i = 0; $i -lt $parts.Count; $i++) {
        $part = $parts[$i]
        if ($part -eq '' -or $part -eq '.') { continue }
        # Track the parent directly: Split-Path -Parent returns an empty
        # string for single-component rooted paths (observed: '/tmp' -> ''),
        # which would make the Join-Path below throw a binding error.
        $parent = $acc
        $acc = Join-Path $acc $part
        # Prefer the .NET link probe: the filesystem provider cannot list
        # some root-level links (observed: Get-Item -LiteralPath /tmp fails
        # on macOS while LinkTarget reports private/tmp). Get-Item stays as
        # the fallback for runtimes without LinkTarget (Windows PowerShell).
        $target = $null
        try { $target = ([System.IO.DirectoryInfo]$acc).LinkTarget } catch { $target = $null }
        if (-not $target) {
          try { $target = ([System.IO.FileInfo]$acc).LinkTarget } catch { $target = $null }
        }
        if (-not $target) {
          try {
            $entry = Get-Item -LiteralPath $acc -ErrorAction Stop
            $target = $entry.Target
            if ($target -is [System.Array]) {
              if ($target.Count -eq 0) { $target = $null } else { $target = $target[0] }
            }
          } catch {
            $target = $null
          }
        }
        if (-not $target) { continue }
        if ($target -is [System.Array]) {
          if ($target.Count -eq 0) { continue }
          $target = $target[0]
        }
        $target = [string]$target
        if ($target -eq '') { continue }
        if (-not [System.IO.Path]::IsPathRooted($target)) {
          $target = Join-Path $parent $target
        }
        if ($i + 1 -lt $parts.Count) {
          $tail = ($parts[($i + 1)..($parts.Count - 1)] -join [string][System.IO.Path]::DirectorySeparatorChar)
          if ($tail -ne '') { $target = Join-Path $target $tail }
        }
        $full = [System.IO.Path]::GetFullPath($target)
        $changed = $true
        $hops++
        break
      }
    }
    return $full
  } catch {
    return $full
  }
}

function Get-CuratorTrustRecorded {
  param([string]$Candidate, [string]$ApprovalsFile)
  if (-not (Test-Path -LiteralPath $ApprovalsFile)) { return @{ State = "absent" } }
  try {
    $lines = Get-Content -LiteralPath $ApprovalsFile -ErrorAction Stop
  } catch {
    return @{ State = "unreadable" }
  }
  foreach ($line in $lines) {
    $fields = $line.Split([char]9)
    if ($fields.Count -ne __CURATOR_RECORD_FIELDS__) { continue }
    if ($fields[0] -cne $Candidate) { continue }
    if ($fields[1] -cnotmatch '^[0-9a-f]{__CURATOR_DIGEST_LEN__}$') { continue }
    if ($fields[2] -cne '__CURATOR_APPROVED_BY_MANAGER__' -and $fields[2] -cne '__CURATOR_APPROVED_BY_OPERATOR__') { continue }
    if ($fields[3] -cnotmatch '__CURATOR_TIMESTAMP_SHAPE__') { continue }
    $parsed = [DateTimeOffset]::MinValue
    if (-not [DateTimeOffset]::TryParse($fields[3], [ref]$parsed)) { continue }
    return @{ State = "found"; SHA256 = $fields[1] }
  }
  return @{ State = "absent" }
}

function Test-CuratorTrustWarned {
  param([string]$Candidate)
  $nl = [string][char]10
  $framed = $nl + $global:CuratorTrustWarned
  return $framed.Contains($nl + $Candidate + $nl)
}

function Write-CuratorTrustWarning {
  param([string]$Candidate, [string]$Diagnostic)
  if (Test-CuratorTrustWarned $Candidate) { return }
  $nl = [string][char]10
  $global:CuratorTrustWarned = $global:CuratorTrustWarned + $Candidate + $nl
  if ($Diagnostic -ceq "shell_hook_env_changed") {
    $message = "curator: ${Diagnostic}: $Candidate changed since approval; run curator hook approve $Candidate to approve the new bytes"
  } else {
    $message = "curator: ${Diagnostic}: $Candidate is not approved; run curator hook approve $Candidate to approve it"
  }
  [Console]::Error.WriteLine($message)
}

function Test-CuratorTrustAllow {
  param([string]$Candidate)
  $Candidate = Get-CuratorTrustCanonical $Candidate
  $approvalsFile = Get-CuratorHookApprovalsFile
  $effectiveProfile = if ($CuratorTrustProfile) { $CuratorTrustProfile } else { '__CURATOR_BAKED_PROFILE__' }
  $digest = Get-CuratorTrustDigest $Candidate
  if (-not $digest) {
    Write-CuratorTrustWarning $Candidate "shell_hook_env_unapproved"
    return $false
  }
  $recorded = Get-CuratorTrustRecorded $Candidate $approvalsFile
  if ($recorded.State -ceq "unreadable") {
    Write-CuratorTrustWarning $Candidate "shell_hook_env_unapproved"
    return $false
  }
  if ($recorded.State -ceq "absent" -or -not $recorded.SHA256) {
    Write-CuratorTrustWarning $Candidate "shell_hook_env_unapproved"
    return ($effectiveProfile -cne "B-enforcing")
  }
  if ($recorded.SHA256 -ceq $digest) {
    return $true
  }
  Write-CuratorTrustWarning $Candidate "shell_hook_env_changed"
  return ($effectiveProfile -cne "B-enforcing")
}
`
	return expandTrustTokens(`# Curator shell hook
$CuratorTrustProfile = '`+string(profile)+`'
`+globalPart+trustPart+`
function Invoke-CuratorAutoEnv {
`+sourceGlobal+`  if ($env:CURATOR_AUTO_ENV -eq "0") {
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
    if (Test-CuratorTrustAllow $envFile) {
      . $envFile
    } else {
      Remove-Item Env:\CURATOR_ACTIVE_ENV -ErrorAction SilentlyContinue
      Remove-Item Env:\CURATOR_OLD_PATH -ErrorAction SilentlyContinue
    }
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
`, profile)
}
