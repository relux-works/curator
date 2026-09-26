# BUG-260921-1fpaij results rev2 — gitignore spawn failure conflated with not-ignored

- Task: BUG-260921-1fpaij (parent STORY-260915-3w11un), role: developer
- Verdict: **ready for review** — rev1 product fix retained byte-identical;
  rev1 gate failure root-caused to two CLI fixtures; fixtures repaired;
  narrow gates green on this exact tree.
- This run: rev2. Rev1 evidence
  (`BUG-260921-1fpaij_results.md`) still describes the product change;
  this note adds the gate-failure analysis, the fixture repair, and
  fresh verification on the exact candidate tree.

## Change (rev1 retained + rev2 fixture repair)

- `internal/gitignore/gitignore.go` — `Missing` classifies: exit status 1
  (`*exec.ExitError`, code 1) = not ignored (unchanged); any other outcome
  (spawn error, `exec.ErrNotFound`, exit 128 or other) returns
  `git check-ignore failed for "<entry>": <stderr or errno>` with `%w`
  chain preserved. New `NotIgnoredError` type (byte-identical policy
  message) and `IsNotIgnored` discriminator. `Ensure` returns the typed
  policy error and propagates tool errors. No retry; stderr captured
  per probe. Byte-identical to rev1 patch.
- `internal/install/install.go` — both `Ensure` gates (managed block §3,
  devsub §4): `IsNotIgnored(err)` → `skipped` as before; any other error →
  `failed` via `failf` with the tool diagnostic. Byte-identical to rev1.
- Tests: `internal/gitignore/gitignore_test.go` (+4 rows),
  `internal/install/gitignore_spawn_test.go` (new, production entry row).
  Byte-identical to rev1.
- `CHANGELOG.md` — `## Unreleased` → `### Fixed` entry. Byte-identical
  to rev1.
- rev2 repair: `cmd/curator/main_test.go` (+8 lines) — `git init` the two
  fixture projects that the corrected contract newly refuses (below).

## Why rev1 failed the gate (evidence-backed)

- Gate run 35558711655 (rev1 head `9b42a8b4`): all 5 go-test lanes
  failed (Test ubuntu/macos/windows, Race ubuntu/macos); Lint, Naming,
  self-tests, Interop green.
- Downloaded lane artifacts (`test-evidence-*`, `race-evidence-*`) and
  read the `go-test-served.json` streams: every failing lane fails on
  exactly the same 2 tests, nothing else:
  - `cmd/curator :: TestCLIExecutionAssuranceSelectionIsPortableDefaultAndVerifiedFailClosed/portable`
    (`main_test.go:345: portable install dry-run exit=1`)
  - `cmd/curator :: TestUpgradeDryRunDoesNotCreateOrFetchSkillsRoot`
    (`main_test.go:703: upgrade --dry-run = 1`)
- Root cause: both fixtures build a project directory with NO `git init`
  (plain `MkdirAll` + Skillfile + correct `.gitignore`) and run a dry-run
  expecting exit 0. Under the old conflation, the exit-128 `check-ignore`
  outside a repo collapsed into "not ignored" → `skipped` → CLI exit 0,
  so the tests passed while the run never proceeded past the gate. Under
  the corrected contract (R1/R3: non-repo → tool error → `failed`), the
  CLI correctly exits 1. The fixtures, not the product code, contradict
  the new contract.
- Repair: `runGit(t, project, "init", "-q", "-b", "main")` in
  `writeAssuranceCLIConfig` and in
  `TestUpgradeDryRunDoesNotCreateOrFetchSkillsRoot`. This is the
  established pattern — every other project fixture in `cmd/curator`
  already does it (`compiledProjectDeclaring`, `legacyProject`,
  `hookStatusProject`, lifecycle `upgradeFixture`, draft `setupDraftCLI`,
  `TestCLIEndToEndInstallStatusAndTamperCheck`). Test intent is
  preserved: the portable dry-run now genuinely proceeds past the gate
  (exit 0), and the upgrade dry-run still creates no skills root. The two
  sibling subtests (`verified missing provider`, `unknown mode`) still
  fail for their own preflight reasons (verified green below).
- No other test in the repo depends on the old non-repo→skip path: the
  gate streams show only these 2 failures, and a grep over all
  `.gitignore`-writing fixtures confirms every other project is a git
  checkout.

## Before / after at the production entry (`install.Project`)

Setup: valid project (Skillfile + correct `.gitignore`), `git` replaced on
PATH by the 0755-shim/0644-interpreter fixture (hosted-gate EACCES shape).

- BEFORE (conflation mutant, observed this run):
  `Status=skipped`,
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
| `cmd/curator` install/upgrade via `Project` (main.go:622) | indirect | spawn failure surfaced as `skipped` (exit 0) | `failed` (exit 1); `skipped` still exit 0 |

## Regression rows (all observed green on this tree)

| Row | Test | Result |
|---|---|---|
| (a) spawn EACCES | `TestMissingSpawnFailureIsToolError` (gitignore) | PASS — tool error, op context, entry name, no probe leak, `permission denied`, `errors.Is EACCES`; `Ensure` likewise, never the policy message |
| (b) non-repo root | `TestMissingNonRepositoryIsToolError` (gitignore) | PASS — tool error + exit-128 chain, never policy |
| (c) exit 1 | `TestMissingExitOneIsNotIgnored` + legacy `TestEnsureFailsWithoutIgnoreThenFixes`, `TestEnsurePassesWhenAlreadyIgnored` | PASS — `Missing`→entries+nil; `Ensure`→`NotIgnoredError`, stable message |
| (d) production entry | `TestProjectRefusesOnGitignoreSpawnFailure` (install) | PASS — `failed` (not `skipped`), diagnostic present, policy message absent |
| (+) git absent | `TestMissingGitAbsentIsToolError` (gitignore) | PASS — tool error, `errors.Is ErrNotFound` |
| existing | `TestGitignoreGateSkips`, `TestAdapterLedgerCommitsAfterTheMirrorsItClaims`, `TestEndToEndInstall`, `TestDryRunTouchesNothing` (install) | PASS — policy path still skips; PR #83 ledger test green; success path intact |
| rev2 fixtures | `TestCLIExecutionAssuranceSelectionIsPortableDefaultAndVerifiedFailClosed` (all 3 subtests), `TestUpgradeDryRunDoesNotCreateOrFetchSkillsRoot`, `TestGlobalUpgradeDryRunDoesNotCreateSkillsRoot` (cmd/curator) | PASS |

Windows skips: the two EACCES rows skip with
`executable-bit spawn refusal is exercised on the unix runners`, matching
skip-classes.tsv `platform-control` (`(is|are) exercised (on|by)`), same
string as sibling rows `gitops.TestWriteBlobsSpawnFailureIsWrapped` and
`install.TestProjectWriteBlobsSpawnFailureIsWrapped`.

## Mutant table (conflation mutant: `if err != nil { missing = append… }`)

Observed by running the mutant on this tree (FAIL everywhere below):

| Mutant vs row | Observed outcome |
|---|---|
| (a) spawn | KILLED — `Missing with a non-executable git must fail` |
| (b) non-repo | KILLED — `Missing outside a repository must fail` |
| (c) exit 1 | survives (correct — the preserved path) |
| (d) production | KILLED — `status = "skipped", want failed` + policy message (the before-text above) |
| (+) absent git | KILLED — `Missing without a git binary must fail` |

Mutant reverted after the run; `git diff --stat` confirms the candidate
is exactly rev1 content + the 8-line fixture repair; restored tree
re-verified green (below).

## Verification log (real exit codes, host macOS, shell bash, `set -o pipefail`)

- `go test ./internal/gitignore/ -count=1 -v` → exit 0, 6/6 PASS.
- `go test ./internal/install/ -run 'TestProjectRefusesOnGitignoreSpawnFailure|TestGitignoreGateSkips|TestAdapterLedgerCommitsAfterTheMirrorsItClaims'` → exit 0, 3/3 PASS.
- `go test ./internal/install/ -run 'TestEndToEndInstall|TestDryRunTouchesNothing'` → exit 0.
- `go test ./cmd/curator/ -run 'TestCLIExecutionAssuranceSelectionIsPortableDefaultAndVerifiedFailClosed|TestUpgradeDryRunDoesNotCreateOrFetchSkillsRoot'` → exit 0 (3/3 subtests + upgrade row PASS).
- `go test ./cmd/curator/ -run '...|TestGlobalUpgradeDryRunDoesNotCreateSkillsRoot'` → exit 0.
- Mutant: gitignore package → FAIL (3 kills + (+) kill); install row (d) → FAIL (kill + before-text). Mutant reverted; restored gitignore package → exit 0; restored install rows → exit 0.
- `gofmt -l internal/gitignore internal/install cmd/curator` → clean.
- `go vet ./internal/gitignore/ ./internal/install/ ./cmd/curator/` → exit 0.
- `golangci-lint run ./internal/gitignore/ ./internal/install/` → 0 issues; `./cmd/curator/` → 0 issues.
- Full `go test ./...` deliberately NOT run (campaign rule: the landing
  suite runs exactly once at handoff). A local full-package
  `go test ./internal/install/` exceeded the 10m default timeout and was
  discarded in favor of the bounded `-run` subsets above; the gate runs
  with a 30m timeout. Legacy goldens untouched (byte-identical success
  path and policy message).

## Ratio line

No corpus change: no conformance vectors, goldens, or snapshots touched.
Diff vs base: 5 product/test files + CHANGELOG (201 insertions,
2 deletions) + 1 new test file; rev2 delta over rev1 is 8 fixture lines
in `cmd/curator/main_test.go`.
