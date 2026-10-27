#!/usr/bin/env bash
# Shell equivalent of internal/testgitenv.Isolate, for standalone fixtures.
# Source this file and call isolate_test_env with a fresh owned directory before
# running any fixture command. The caller owns cleanup and the process lifetime;
# this function never touches or restores caller repositories or native homes.

isolate_test_env() {
	local root key upper windows
	root="$1"
	# Refuse reuse and canonicalize before creating config or exporting paths.
	(umask 077; mkdir "$root") || return 2
	root="$(cd "$root" && pwd -P)" || return 2
	windows="$(uname -s)" || return 2
	case "$windows" in
		MINGW*|MSYS*|CYGWIN*) windows=1 ;;
		*) windows=0 ;;
	esac
	for key in $(compgen -e); do
		upper="$(printf '%s' "$key" | tr '[:lower:]' '[:upper:]')"
		case "$upper" in
			GIT_CONFIG*|GIT_DIR|GIT_WORK_TREE|GIT_INDEX_FILE|GIT_COMMON_DIR|GIT_OBJECT_DIRECTORY|\
			GIT_ALTERNATE_OBJECT_DIRECTORIES|GIT_NAMESPACE|GIT_TEMPLATE_DIR|GIT_CEILING_DIRECTORIES|\
			GIT_DISCOVERY_ACROSS_FILESYSTEM|GIT_IMPLICIT_WORK_TREE|GIT_GRAFT_FILE|GIT_SHALLOW_FILE|\
			GIT_NO_REPLACE_OBJECTS|GIT_REPLACE_REF_BASE|GIT_PREFIX|GIT_QUARANTINE_PATH|\
			GIT_REFERENCE_BACKEND|GIT_DEFAULT_HASH|GIT_DEFAULT_REF_FORMAT|\
			SSH_AUTH_SOCK|HOME|XDG_CONFIG_HOME|CODEX_HOME|CLAUDE_CONFIG_DIR|PI_CODING_AGENT_DIR|\
			XDG_DATA_HOME|XDG_STATE_HOME|XDG_CACHE_HOME)
				unset "$key" || return 2 ;;
			USERPROFILE)
				if [ "$windows" -eq 1 ]; then unset "$key" || return 2; fi ;;
		esac
	done
	(umask 077; mkdir "$root/home" "$root/xdg" "$root/tmp" && : >"$root/gitconfig") || return 2
	HOME="$root/home"
	XDG_CONFIG_HOME="$root/xdg"
	GIT_CONFIG_GLOBAL="$root/gitconfig"
	GIT_CONFIG_NOSYSTEM=1
	TMPDIR="$root/tmp"
	TMP="$TMPDIR"
	TEMP="$TMPDIR"
	export HOME XDG_CONFIG_HOME GIT_CONFIG_GLOBAL GIT_CONFIG_NOSYSTEM
	export TMPDIR TMP TEMP
	if [ "$windows" -eq 1 ]; then
		USERPROFILE="$HOME"
		export USERPROFILE
	fi
}
