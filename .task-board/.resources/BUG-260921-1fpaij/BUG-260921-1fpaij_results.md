# BUG-260921-1fpaij results — gitignore spawn failure conflated with not-ignored

- Task: BUG-260921-1fpaij (parent STORY-260915-3w11un), role: developer
- Verdict: **ready for review** — fix + regression rows + CHANGELOG complete;
  narrow gates green in a Linux container (host macOS cannot exec new
  binaries; see anomaly note).

## Change

- `internal/gitignore/gitignore.go` — `Missing` now classifies: exit status 1
  (`*exec.ExitError`, code 1) = not ignored (unchanged); any other outcome
  (spawn error, `exec.ErrNotFound`, exit 128 or other) returns
  `git check-ignore failed for "<entry>": <stderr or errno>` with `%w` chain
  preserved. New `NotIgnoredError` type (byte-identical policy message) and
  `IsNotIgnored` discriminator. `Ensure` returns the typed policy error and
  propagates tool errors. No retry; stderr captured per probe.
- `internal/install/install.go` — both `Ensure` gates (managed block §3,
  devsub §4): `IsNotIgnored(err)` → `skipped` as before; any other error →
  `failed` via `failf` with the tool diagnostic. Fail-closed preserved.
- Tests: `internal/gitignore/gitignore_test.go` (+4 rows),
  `internal/install/gitignore_spawn_test.go` (new, production entry row).
- `CHANGELOG.md` — `## Unreleased` → `### Fixed` entry.

## Before / after at the production entry (`install.Project`)

Setup: valid project (Skillfile + correct `.gitignore`), `git` replaced on
PATH by the 0755-shim/0644-interpreter fixture (hosted-gate EACCES shape).

- BEFORE (conflation mutant, observed): `Status=skipped`,
  `Messages=[test: generated paths are not ignored by git; missing entries:
  .agents/, .claude/skills/; skipped]` — the exact PR #83 run 35550731645
  symptom.
- AFTER (this fix, observed): `TestProjectRefusesOnGitignoreSpawnFailure`
  passes — `Status=failed`, `Errors` carry
  `git check-ignore failed … permission denied`, and no output contains
  `generated paths are not ignored`.

## Caller table

| Caller | Git call | Before | After |
|---|---|---|---|
| `internal/install` `projectAttempt` §3 (install.go:262) managed gate | `Ensure(required, fix)` | any error → `skipped` | `NotIgnored` → `skipped` (unchanged); tool error → `failed` + diagnostic |
| `internal/install` `projectAttempt` §4 devsub gate | `Ensure([devsub.Name], fix)` | any error → `skipped` | same split as §3 |
| `cmd/curator` `cmdInit` (main.go:453), project add (main.go:1167) | `Append` only (no git spawn) | n/a | unchanged |
| `cmd/curator` install/status via `Project` (main.go:622,753) | indirect | spawn failure surfaced as `skipped` | `failed`; status path prints + `exitFail` (existing `failed` branch, BuildsComplete=false) |

## Regression rows (all observed green)

| Row | Test | Result |
|---|---|---|
| (a) spawn EACCES | `TestMissingSpawnFailureIsToolError` (gitignore) | PASS — tool error, op context, entry name, no probe leak, `permission denied`, `errors.Is EACCES`; `Ensure` likewise, never the policy message |
| (b) non-repo root | `TestMissingNonRepositoryIsToolError` (gitignore) | PASS — tool error + exit-128 chain, never policy |
| (c) exit 1 | `TestMissingExitOneIsNotIgnored` + existing `TestEnsureFailsWithoutIgnoreThenFixes`, `TestEnsurePassesWhenAlreadyIgnored` | PASS — `Missing`→entries+nil; `Ensure`→`NotIgnoredError`, stable message |
| (d) production entry | `TestProjectRefusesOnGitignoreSpawnFailure` (install) | PASS — `failed` (not `skipped`), diagnostic present, policy message absent |
| (+) git absent | `TestMissingGitAbsentIsToolError` (gitignore) | PASS — tool error, `errors.Is ErrNotFound` |
| existing | `TestGitignoreGateSkips`, `TestAdapterLedgerCommitsAfterTheMirrorsItClaims` (install) | PASS — policy path still skips; PR #83 ledger test green |

Windows skips: the two EACCES rows skip with
`executable-bit spawn refusal is exercised on the unix runners`, matching
skip-classes.tsv `platform-control` (`(is|are) exercised (on|by)`), same
string as sibling rows `gitops.TestWriteBlobsSpawnFailureIsWrapped` and
`install.TestProjectWriteBlobsSpawnFailureIsWrapped`.

## Mutant table (conflation mutant: `if err != nil { missing = append… }`)

Observed by running the mutant in the container (exit 1 everywhere below):

| Mutant vs row | Observed outcome |
|---|---|
| (a) spawn | KILLED — `Missing with a non-executable git must fail` |
| (b) non-repo | KILLED — `Missing outside a repository must fail` |
| (c) exit 1 | survives (correct — the preserved path) |
| (d) production | KILLED — `status = "skipped", want failed` + policy message (the before-text above) |
| (+) absent git | KILLED — `Missing without a git binary must fail` |

Mutant reverted after the run; restored tree re-verified green (below).

## Verification log (real exit codes)

Test execution ran in `golang:1.26-bookworm` with the worktree mounted
read-only (`-v $PWD:/src:ro`), because the host cannot exec new binaries
(see anomaly). Shell `bash`, `set -o pipefail` where piped:

- `go test ./internal/gitignore/ -count=1 -v` → exit 0, 6/6 PASS.
- `go test ./internal/install/ -run
  'TestProjectRefusesOnGitignoreSpawnFailure|TestGitignoreGateSkips|TestAdapterLedgerCommitsAfterTheMirrorsItClaims'`
  → exit 0, 3/3 PASS.
- Mutant: gitignore package → exit 1 (3 kills); install row (d) → exit 1
  (kill + before-text). Mutant reverted byte-exact (diff stat identical).
- Restored tree: gitignore package + install rows (d, skip) → exit 0.
- Host-native static gates on the exact tree: `go vet`
  `./internal/gitignore/ ./internal/install/` → 0;
  `gofmt -l internal/gitignore internal/install` → clean;
  `golangci-lint run ./internal/gitignore/ ./internal/install/` → 0;
  `GOOS=linux go build` both packages → 0.
- Shell-level git semantics (host, observed): unignored → exit 1; ignored →
  exit 0; non-repo → exit 128 + `fatal: not a git repository …`.
- Full `go test ./...` deliberately NOT run (campaign rule: the landing
  suite runs exactly once at handoff); legacy goldens untouched
  (byte-identical success path and policy message).

Anomaly (external, evidence-backed): this macOS host SIGKILLs every newly
created Mach-O (trivial Go hello-world, copied `/bin/echo`, ad-hoc
re-signed binary, and `cc` itself all `Killed: 9` after hanging; `go test`
died with `signal: killed` 3× at ~84s). Pre-existing binaries unaffected;
no memorystatus/code-signing kill in logs; `spctl` assessments disabled.
Workaround used: read-only worktree mount in Docker. No host state changed.

## Ratio line

No corpus change: no conformance vectors, goldens, or snapshots touched.
Diff: 2 product files, 2 test files, CHANGELOG (193 insertions,
2 deletions + 1 new test file).
