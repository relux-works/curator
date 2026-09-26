# BUG-260921-1fpaij results (rework-1 implementation — narrowed classifier)

> Supersedes the prior board rev2 attempt ("gate-failure analysis, fixture
> repair, verification", 6-path patch incl. `main_test.go`), which kept the
> strict classifier and repaired fixtures instead. This revision implements the
> binding rework-1 ruling: narrow the classifier to spawn-only errors,
> `main_test.go` untouched (5 paths: 4 modified + 1 new).

## Rule (orchestrator rework 1; narrows brief R1)

- A git that **cannot be executed** — any `cmd.Run()` failure that is NOT an
  `*exec.ExitError` (`*exec.Error`/`exec.ErrNotFound`, `*fs.PathError` such as
  `fork/exec …: permission denied`) — is returned from `Missing`/`Ensure` as a
  wrapped `git check-ignore failed for "<entry>": …` error carrying the tool
  diagnostic (entry name only; no probe path, environment, or root).
- EVERY exit status git itself reports keeps the pre-existing behaviour (entry
  reported as not ignored): 1 (not ignored) AND 128/other (not a repository,
  fatal git error). "Git's own verdicts, including 'not a repository', are
  policy outcomes as before."
- No retry. Success path and the `generated paths are not ignored by git; …`
  message byte-identical (`NotIgnoredError`).

## Before / after at the production entry (`install.Project`)

- Before (base): any `git check-ignore` failure, including a spawn failure,
  → `skipped` with `generated paths are not ignored by git; missing entries: …`
  (PR #83 run 35550731645, macos-latest,
  `TestAdapterLedgerCommitsAfterTheMirrorsItClaims` skipped exactly so while
  `.gitignore` was correct).
- After: spawn failure → `failed` with
  `git check-ignore failed for ".claude/skills/": fork/exec … permission denied`;
  the policy message never appears. Non-repository roots still yield the
  pre-existing `skipped` + policy message, as on main (CLI exit 0 either way;
  only `failed` exits 1 — `cmd/curator/main.go` `cmdInstallMode`).

## Caller table (`Missing`/`Ensure` callers; all keep fail-closed behaviour)

| Caller | Spawn failure (non-`NotIgnoredError`) | Git verdict (`NotIgnoredError`, any exit) |
|---|---|---|
| `internal/install` `projectAttempt` managed gate (install.go:262) | `failed` with `git check-ignore failed …` | `skipped` with policy message (unchanged) |
| `internal/install` `projectAttempt` devsub gate (install.go:291) | `failed` with `git check-ignore failed …` | `skipped` with policy message (unchanged) |
| `cmd/curator` | n/a — only `gitignore.Append` is used there | n/a |

## Rows (candidate tree; shell: bash, `set -o pipefail` semantics via `PIPESTATUS`)

| # | Row | Command / test | Result |
|---|---|---|---|
| a | injected non-executable git → tool error | `internal/gitignore` `TestMissingSpawnFailureIsToolError` | PASS (0.04s) |
| b | non-repository root → pre-existing not-ignored, no error | `internal/gitignore` `TestMissingNonRepositoryIsNotIgnored` | PASS (0.07s) |
| c | exit 1 unchanged | `internal/gitignore` `TestMissingExitOneIsNotIgnored` + 2 pre-existing rows | PASS |
|GitAbsent| git absent → `exec.ErrNotFound` tool error | `internal/gitignore` `TestMissingGitAbsentIsToolError` | PASS (0.05s) |
| d | production entry refuses with spawn diagnostic, never policy msg | `internal/install` `TestProjectRefusesOnGitignoreSpawnFailure` (+ `TestGitignoreGateSkips` green) | PASS (0.31s) |
| gate1 | `TestCLIExecutionAssuranceSelectionIsPortableDefaultAndVerifiedFailClosed` (all 3 subtests) on unmodified non-repo fixtures | `go test ./cmd/curator/ -run …` | PASS |
| gate2 | `TestUpgradeDryRunDoesNotCreateOrFetchSkillsRoot` on unmodified non-repo fixture | same run | PASS |
| rework | `go test ./internal/gitignore/ ./cmd/curator/ -run 'Gitignore\|DryRun\|ExecutionAssurance' -count=1` | exit 0 both packages (gitignore rows matched by exact names above; the mask string matches no test name in that package) | PASS |
| lint/vet/build | `go build ./...`, `go vet` (3 touched pkgs), `gofmt -l` (clean), `golangci-lint run ./internal/gitignore/... ./internal/install/...` | 0 issues, all exit 0 | PASS |
| install A–C | `go test ./internal/install/ -run '^Test[A-C]'` (57 tests) | PASS (105s) | PASS |

Full `internal/install` (295 tests) and `internal/crossconformance` (37 goldens):
NOT rerun to green locally — shared-host OOM, evidence below. Reran-myself list
ends here; everything above was observed by me on the candidate tree. The
authoritative full landing suite runs on fresh hosted runners at handoff.

## Mutant table (both applied to `internal/gitignore/gitignore.go`, then reverted; restore verified byte-identical via `diff`)

| Mutant | Expected | Observed |
|---|---|---|
| M1 conflation restored (`if err != nil { missing = append… }`) | FAIL rows a, d, GitAbsent; PASS b, c + pre-existing | exact: FAIL `TestMissingSpawnFailureIsToolError`, `TestMissingGitAbsentIsToolError`, `TestProjectRefusesOnGitignoreSpawnFailure` (exit 1); PASS `TestMissingExitOneIsNotIgnored`, `TestMissingNonRepositoryIsNotIgnored`, both pre-existing rows |
| M2 revision-1 strictness (`ExitCode() == 1` gate) | FAIL converted row b + the two CLI gate tests | exact: FAIL `TestMissingNonRepositoryIsNotIgnored`; CLI run FAILs with `main_test.go:703: upgrade --dry-run = 1` — the precise gate failure cited in rework 1, reproduced locally |

M1 kills the under-narrowed bound (spawn failure must be an error); M2 kills the
over-narrowed bound (exit 128 must stay a policy outcome). The classifier sits
exactly on the rework-1 boundary: error iff non-`*exec.ExitError`.

## Ratio line

No corpus change: `crossconformance` sources untouched; goldens exercise install
only through paths where git spawns normally (9 draftsources semantic test files
import `internal/install`), under which this change is a no-op. Local golden run
blocked by host OOM (below); hosted lanes are authoritative.

## Supplementary-run blocker (environmental, evidence-backed)

- `go test ./internal/install/` (full, default 600s timeout): FAIL by timeout
  while a second heavy suite ran concurrently (CI uses `-timeout 30m` per
  `.github/ci/test-gate.sh`; the 600s default is my invocation's artifact).
- Follow-ups all die with `signal: killed` / exit 137, including a ZERO-WORK run:
  `/tmp/install.test -test.list '^TestD'` (prebuilt 19MB binary, prints names
  only, executes no test and no product code): 79.8s wall, ~0.00s CPU, then
  SIGKILL. D1 verbose run: killed before the first `=== RUN` line.
- Even the tiny `internal/gitignore` package (6/6 green earlier on the identical
  tree) is now SIGKILLed at 81s. Host: 32GB box, 12k–46k free pages (50–180MB)
  across 8 probes / 8 min, load 3–7, with other agents' suites (`godriver.test`,
  `scriptworker.test`, mutant binaries) resident. macOS kills new mappings under
  this pressure; no Go test binary of any size can start reliably.
- Attribution is airtight: the same tree ran install chunk 1 (57 tests) green
  1h earlier; nothing but host load changed since. No source edit stands between
  the green rows and the kills (mutant restore verified with `diff`).

## Tree

- 4 files changed, 1 added; `cmd/curator/main_test.go` untouched (revision-1
  local fixture `git init`s reverted: the cited gate lines 345/703 match the
  tree WITHOUT them, so the published fixtures stay non-repo and the two CLI
  tests pin the rework ruling directly).
- Files: `internal/gitignore/gitignore.go`, `internal/gitignore/gitignore_test.go`,
  `internal/install/install.go`, `internal/install/gitignore_spawn_test.go` (new),
  `CHANGELOG.md` (`## Unreleased` → `### Fixed`).
