# Shell activation regression audit

## Findings

- POSIX upward search stops only at `/`; an empty or relative `PWD` can repeat forever.
- Project activation marks the environment active after sourcing `env.sh`; zsh directory hooks can re-enter during the internal root-resolution `cd`.
- Re-loading the POSIX hook prepends another `PROMPT_COMMAND` entry and can add duplicate zsh hooks.
- The POSIX global lookup uses an external `dirname` and does not normalize a native Windows drive path supplied to Git Bash.
- Hook generation requires an explicit shell and has no atomic cached form, which encourages invoking the manager binary from every shell startup.
- Global commands exist only below the manager home, so their availability currently depends on shell activation.
- Skill validation does not detect prompt-visible runtime-only paths or missing cross-platform shim resolution guidance.

## Required behavior

Shell activation is an optional human convenience. Agent command execution must use installed shims and must not require a user profile. Optional hooks must be finite, reentrancy-safe, idempotent, cached when installed, and portable across zsh, bash, PowerShell, and Git Bash.

## Verification

- All Go packages pass with the authoritative conformance root.
- `go vet ./...` passes.
- golangci-lint v2 reports zero issues.
- Formatting, whitespace, and repository naming gates pass.
- The specification validator passes 25 schemas and 79 vector files.
