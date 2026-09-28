# TASK-260916-1ihonr results — launcher config family ownership check

## Changes

- Added `internal/configfile` for checked reads used by both `defaults.Load` and `axconfig.Load`. It distinguishes absence from failed inspection/read, rejects a final symlink and non-regular file, compares file ownership to its configuration directory, and checks POSIX group/other write bits. Unix opens use `O_NOFOLLOW`; Windows opens the reparse point itself and checks the owner SID and DACL.
- Windows DACL policy refuses null/missing DACLs, malformed/uninspectable ACLs, and write/delete/security-control grants to untrusted identities. Owner, OWNER RIGHTS/CREATOR OWNER, LocalSystem, and Builtin Administrators are trusted identities.
- Added SPEC §4.7 and §6 contract text, an Unreleased CHANGELOG entry, production-entry-point rows for machine/operator `defaults.json` and `ax.json`, symlink/group-writable stderr goldens, unreadable and different-owner rows, a valid machine/operator path through fake ax, and a Windows DACL fixture test.
- No live `ax` call or interactive permission bypass was used.

## Local validation (all commands run directly)

| Command | Exit |
|---|---:|
| `go test ./cmd/curator-run -run '^TestRunConfigSecurity' -count=1` | 0 |
| `go test ./internal/configfile ./internal/defaults ./internal/axconfig -count=1` | 0 |
| `go vet ./cmd/curator-run ./internal/configfile ./internal/defaults ./internal/axconfig` | 0 |
| `go build -o /tmp/curator-run ./cmd/curator-run` | 0 |
| `make fmt-check` | 0 |
| `git diff --check` | 0 |
| `GOOS=windows GOARCH=amd64 go test -c -o /tmp/configfile-windows.test.exe ./internal/configfile` | 0 (compile only) |
| `GOOS=windows GOARCH=amd64 go test -c -o /tmp/defaults-windows.test.exe ./internal/defaults` | 0 (compile only) |
| `GOOS=windows GOARCH=amd64 go build ./internal/configfile ./internal/defaults ./internal/axconfig` | 0 |

Windows runtime evidence is unavailable here. The Windows DACL test binary compiled but was not executed on Windows. A full Windows build/test of `cmd/curator-run` currently exits 1 in existing Unix-only `internal/execution/process.go` (missing Unix ioctl/signals and `SysProcAttr` fields); `GOOS=windows go test -c ./internal/axconfig` also exits 1 because existing `internal/axconfig/fifo_test.go` calls Unix-only `syscall.Mkfifo`. The loader packages themselves build for Windows. These are reported as unverified, not passing.

## Narrowing mutants

Each mutant was applied temporarily, its focused test command was run directly, and the production source was restored before final green checks. Each mutant test exited 1 as expected:

| Mutant | Focused command | Exit |
|---|---|---:|
| Follow a symlink (disable final-link and no-follow checks) | `go test ./cmd/curator-run -run '^TestRunConfigSecurityGoldens/symlinked-operator-defaults$' -count=1` | 1 |
| Ignore owner mismatch | `go test ./cmd/curator-run -run '^TestRunConfigSecurityRejectsDifferentDirectoryOwner$' -count=1` | 1 |
| Ignore group/world write bits | `go test ./cmd/curator-run -run '^TestRunConfigSecurityRefusalRows$' -count=1` | 1 |
| Treat unreadable as absent | `go test ./cmd/curator-run -run '^TestRunConfigSecurityRefusalRows$' -count=1` | 1 |
| Admit foreign DACL write grants | `go test ./internal/configfile -run '^TestForeignDACLWriteGrantPolicy$' -count=1` | 1 |

The task estimate was set to Fibonacci 5 after the required development transition was refused for missing estimate. A fresh `origin` advertisement and exact-ref fetch agreed on `main` at `27cc242d393afb471b62c230101caf890dbd2fb7`, matching the recorded Story base/tip.
