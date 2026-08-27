# STORY-260713-b4e219: seamless-manager-lifecycle

## Description
Make repository bootstrap, dependency upgrades, dry-run planning, and command execution work without shell-profile setup or unrelated source mutations.

## Acceptance Criteria
- Runtime launchers carry project, manager-runtime, and declared system dependency paths on Unix and Windows
- Launchers preserve inherited PATH, arguments, and child exit status
- Bootstrap can create configuration only when absent without parsing or rewriting an existing file
- Upgrade fetches only the selected direct and transitive dependency closure
- Multi-project upgrade deduplicates shared repository fetches
- Project and global dry runs leave persistent source, cache, security, runtime, configuration, and install state unchanged
- Authoritative manager lifecycle vectors pass on all supported platforms
