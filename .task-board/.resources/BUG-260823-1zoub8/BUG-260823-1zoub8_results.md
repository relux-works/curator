# BUG-260823-1zoub8 developer outcome

## Delivered change

- Added one `errBuildSSHAgentUnset` diagnostic carrying `build_repository_ssh_credential_missing` and the operator remedy that `SSH_AUTH_SOCK` must be set.
- Both production paths (`runWideCredentials` and `scopeCredentials` through `matchScope`) now return that exact diagnostic without path-specific wrapping.
- Added `TestAgentUnsetDiagnosticIsIdenticalForRunWideAndScopedSelections`, which drives `resolveBuildSSH` through both selections and asserts exact message equality.

## HTTPS sibling check

The HTTPS credential surface landed on current `origin/main` after this Story worktree was created. It has no equivalent duplicate condition to align: an absent run-wide token selects anonymous HTTPS, while configured `token_env`, Git credential, and keyring failures are intentionally distinct missing-source conditions with source-specific remedies. The SSH files are byte-identical between this worktree base and `origin/main`, so this two-file patch applies without importing unrelated HTTPS commits.

## Validation evidence

- Expected-red regression: `go test ./internal/install -run '^TestAgentUnsetDiagnosticIsIdenticalForRunWideAndScopedSelections$' -count=1` exited 1 before the production fix; scoped output had a path-specific wrapper and no protocol code.
- Targeted regression after fix: same command exited 0.
- Package suite: `go test ./internal/install -count=1` exited 0 (`48.883s`).
- First full suite: `go test ./... -count=1` exited 1 because the worktree's pinned `agents/skills/skill-go-testing-tools` submodule was not initialized; `internal/ui` could not resolve the local replacement. This is the worktree setup anomaly already documented in `LOGBOOK.md`.
- Setup repair: `git submodule update --init --recursive agents/skills/skill-go-testing-tools` exited 0 and checked out the pinned revision `21585d0e937cae47e54a788d8ae36b1780eae47f`.
- Full suite rerun: `go test ./... -count=1` exited 0; all packages passed, including `internal/install` and `internal/ui`.
- Lint: `golangci-lint run` exited 0 with `0 issues.`
- Build: `go build -o .temp/BUG-260823-1zoub8/curator ./cmd/curator` exited 0.
- Formatting/diff: `gofmt -l` check and `git diff --check` exited 0.

No files were staged or committed.
