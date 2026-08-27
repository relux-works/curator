# STORY-260713-9f4c2b: seamless-shell-onboarding

## Description
Make command execution independent from shell profile setup while keeping optional interactive activation safe on zsh, bash, PowerShell, and Git Bash.

## Acceptance Criteria
- Project commands remain directly addressable through portable project shims
- Global commands are published to a safe existing user PATH directory when possible
- Hook generation auto-detects the current supported shell and can cache a sourceable hook
- POSIX activation terminates for invalid working directories and cannot re-enter while sourcing an environment
- Re-loading hooks does not duplicate prompt or directory-change integration
- Windows native paths and Git Bash drive paths resolve the manager home consistently
- Skill validation warns about prompt-visible runtime paths and missing shell-neutral command resolution
- Cross-platform tests cover the regression paths
