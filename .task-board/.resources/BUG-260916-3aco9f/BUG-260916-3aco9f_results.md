# BUG-260916-3aco9f results

## Changes

Added CLI-driven profile install matrix coverage in `cmd/curator/profile_install_matrix_test.go`. The tests call `run()` and exercise `cmdProfileInstallInstall` → `envprofile.Install` → `installLocked`, including the same-source delegation to `reinstallPathLocked` and the Git update/reinstall activation path.

The matrix covers path roots, Git roots, and Git collection members selected with `--directory contexts/tk`; each has first-install, same-source unchanged-lock, and same-source changed-lock cases, with no flags, `--use`, `--takeover`, and both flags. Unchanged-lock cases compare the actual CLI-reported lock hash before and after. Changed-source cases require the hash to move.

Added a stop-then-retry CLI conformance test for all three source shapes. The first `--use` attempt stops on an unmanaged Claude surface after publishing the lock. Retrying the same source with `--use --takeover` switches to the target, prints the replacement notice and backup path, backs up the original bytes, and keeps the lock hash unchanged.

Expanded the project-scope preservation row across all three source shapes: a same-source unchanged-lock `--use` switches the machine current profile while the independent `codex_cli` project current and its materialized bytes remain unchanged.

The previous attached revision-1 landing validation failed on Windows at the unrelated `internal/registry::TestSnapshotZeroClockSkewIsLiteral` case. The checker correctly includes its own initialization latency; the test fixture had created and synced its cache after sampling a 500 ms timing edge. The test now primes both cache catalogs before sampling `now`, following the neighboring exact-skew test's setup. This is test-only and changes no production behavior.

## Install matrix

Every row below is an individual `run()`-driven subtest. “Expected refusal” is a passing result: `--use` attempts activation, reports the unmanaged-surface conflict and partial use, leaves the old current profile and unmanaged bytes intact, and makes no backup. “No switch” rows confirm that `--takeover` alone does not activate.

| Shape | Install state | Flags | Status |
|---|---|---|---|
| Path root | First install | none | PASS — lock installed; old current remains; no takeover |
| Path root | First install | `--use` | PASS — expected refusal; old current and bytes remain |
| Path root | First install | `--takeover` | PASS — no switch; no backup |
| Path root | First install | `--use --takeover` | PASS — target becomes current; notice and backup verified |
| Path root | Same source, unchanged lock | none | PASS — lock hash unchanged; no switch |
| Path root | Same source, unchanged lock | `--use` | PASS — expected refusal; lock hash unchanged |
| Path root | Same source, unchanged lock | `--takeover` | PASS — lock hash unchanged; no switch |
| Path root | Same source, unchanged lock | `--use --takeover` | PASS — lock hash unchanged; target becomes current with backup |
| Path root | Same source, changed lock | none | PASS — lock hash changes; old current remains |
| Path root | Same source, changed lock | `--use` | PASS — lock advances; activation refusal preserves old current and bytes |
| Path root | Same source, changed lock | `--takeover` | PASS — lock advances; takeover alone does not switch |
| Path root | Same source, changed lock | `--use --takeover` | PASS — lock advances; target becomes current with backup |
| Git root | First install | none | PASS — lock installed; old current remains; no takeover |
| Git root | First install | `--use` | PASS — expected refusal; old current and bytes remain |
| Git root | First install | `--takeover` | PASS — no switch; no backup |
| Git root | First install | `--use --takeover` | PASS — target becomes current; notice and backup verified |
| Git root | Same source, unchanged lock | none | PASS — lock hash unchanged; no switch |
| Git root | Same source, unchanged lock | `--use` | PASS — expected refusal; lock hash unchanged |
| Git root | Same source, unchanged lock | `--takeover` | PASS — lock hash unchanged; no switch |
| Git root | Same source, unchanged lock | `--use --takeover` | PASS — lock hash unchanged; target becomes current with backup |
| Git root | Same source, changed lock | none | PASS — lock hash changes; old current remains |
| Git root | Same source, changed lock | `--use` | PASS — lock advances; activation refusal preserves old current and bytes |
| Git root | Same source, changed lock | `--takeover` | PASS — lock advances; takeover alone does not switch |
| Git root | Same source, changed lock | `--use --takeover` | PASS — lock advances; target becomes current with backup |
| Git collection | First install | none | PASS — lock installed; old current remains; no takeover |
| Git collection | First install | `--use` | PASS — expected refusal; old current and bytes remain |
| Git collection | First install | `--takeover` | PASS — no switch; no backup |
| Git collection | First install | `--use --takeover` | PASS — target becomes current; notice and backup verified |
| Git collection | Same source, unchanged lock | none | PASS — lock hash unchanged; no switch |
| Git collection | Same source, unchanged lock | `--use` | PASS — expected refusal; lock hash unchanged |
| Git collection | Same source, unchanged lock | `--takeover` | PASS — lock hash unchanged; no switch |
| Git collection | Same source, unchanged lock | `--use --takeover` | PASS — lock hash unchanged; target becomes current with backup |
| Git collection | Same source, changed lock | none | PASS — lock hash changes; old current remains |
| Git collection | Same source, changed lock | `--use` | PASS — lock advances; activation refusal preserves old current and bytes |
| Git collection | Same source, changed lock | `--takeover` | PASS — lock advances; takeover alone does not switch |
| Git collection | Same source, changed lock | `--use --takeover` | PASS — lock advances; target becomes current with backup |

Matrix total: 36/36 applicable CLI rows passed.

## Stop-then-retry and scopes

| Shape | Stop then retry | Project scope preservation |
|---|---|---|
| Path root | PASS — stop after publishing; unchanged-lock takeover retry switches and backs up | PASS — machine current switches to `tk`; `codex_cli=project` and its bytes remain |
| Git root | PASS — stop after publishing; unchanged-lock takeover retry switches and backs up | PASS — machine current switches to `tk`; `codex_cli=project` and its bytes remain |
| Git collection | PASS — stop after publishing; unchanged-lock takeover retry switches and backs up | PASS — machine current switches to `tk`; `codex_cli=project` and its bytes remain |

## Applicability boundary

`curator profile install` accepts Git and local context-package path sources; its Git repository/collection forms are included above. Skillfile schema-2 source declarations are consumed by the separate top-level `curator install` command, which has no `--use` or `--takeover` flags and does not enter `internal/envprofile`. There is no valid profile-install flag row for that separate API, so it is recorded as not applicable rather than represented by a forced-fit test.

The stop-then-retry vector is a CLI test in the curator repository. The earlier attached investigation found protocol vectors for lower-level environment operations but no profile-install lifecycle vector family to extend; the configured `curator-spec` checkout was not present at its documented path in this run. No spec repository was modified.

Issue #73 remains open at this producer handoff. No issue post or close action was made, per the task instruction; the orchestrator owns closing it with the landed commit.

## Mutation evidence

A temporary narrowing mutant changed the unchanged-lock path reinstall activation branch to skip activation (`activate && false`). The targeted CLI test failed as expected: it received exit 0 and “updated profile” instead of the unmanaged-surface refusal. The mutant was reverted; the same targeted row then passed. This kills the path-shape activation-drop mutant. The earlier attached results also report a Git activation-drop mutant killed at `git-root/same-source-unchanged-lock/use` (expected-red exit 1); that mutant was not rerun in this turn. The existing eight-row Git reinstall suite was rerun green below.

## Verification

All commands below were run directly and their exit codes are real.

| Command | Exit | Result |
|---|---:|---|
| `go test ./cmd/curator -run '^TestProfileInstallActivationMatrix$/^path-root$' -count=1` (first run) | 1 | Test-helper defect: selected the current-marker column instead of the lock-hash column. Fixed the helper to read column 6; rerun below is green. |
| `go test ./cmd/curator -run '^TestProfileInstallActivationMatrix$/^path-root$' -count=1` | 0 | All 12 path-root matrix rows |
| `go test ./cmd/curator -run '^TestProfileInstallActivationMatrix$/^git-root$' -count=1` | 0 | All 12 Git-root matrix rows |
| `go test ./cmd/curator -run '^TestProfileInstallActivationMatrix$/^git-collection$' -count=1` | 0 | All 12 Git-collection matrix rows |
| `go test ./cmd/curator -run '^TestProfileInstallStopThenRetryConformanceVector$' -count=1` | 0 | Stop/retry on all three shapes |
| `go test ./cmd/curator -run '^TestProfileInstallReinstallUsePreservesProjectScope$' -count=1` | 0 | Scope preservation on all three shapes |
| `go test ./cmd/curator -run '^TestProfileGitReinstallHonoursUseAndTakeover$' -count=1` | 0 | Existing eight-row Git reinstall suite |
| `go test ./cmd/curator -run '^TestProfileInstallActivationMatrix$/^path-root$/^same-source-unchanged-lock$/^use$' -count=1` with activation-drop mutant | 1 | Expected red; mutant killed |
| Same targeted path row after reverting mutant | 0 | Green |
| `go test ./internal/registry -run '^TestSnapshotZeroClockSkewIsLiteral$' -count=1` | 0 | Stable fixture test |
| `go build -o /tmp/curator-profile-install-validation ./cmd/curator` | 0 | CLI builds |
| `go vet ./cmd/curator ./internal/registry` | 0 | Clean |
| `golangci-lint run ./cmd/curator ./internal/registry` | 0 | 0 issues |
| `gofmt -d cmd/curator/profile_install_matrix_test.go internal/registry/registry_test.go` | 0 | No formatting diff |

The attached revision-1 remote gate ran as `scripts/remote-gate.sh` and exited 1 because the Windows test job failed at `internal/registry::TestSnapshotZeroClockSkewIsLiteral`; Mac and Ubuntu test jobs and the other reported gates were green. That result is not treated as passing. The corrected registry fixture was rerun locally as listed above. No full landing suite was run manually; `task-board handoff` owns its single configured-suite run.
## CHANGELOG entry (for release prep)

Profile installs now honor `--use` and `--takeover` when reinstalling an already-recorded source, including unchanged-lock retries after an unmanaged-surface stop.
