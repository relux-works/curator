# BUG-260924-5p8b0z — integration precondition check (bound developer run)

Board status observed: `integrating` (left untouched; no status/handoff writes per binding).
`task-board worktree integrate` was NOT executed in this run per the binding
Integration Assignment (runner performs the bound landing synchronously).

## Landing preconditions (verified read-only, worktree unchanged except verification)
- Worktree: `.temp/STORY-260924-2tyzhh/worktree`, base `5b326aa3`.
- `git status --short` contains exactly:
  - `M internal/scriptworker/exec.go`
  - `?? internal/scriptworker/exec_identity_conformance_test.go`
  - `?? internal/scriptworker/testdata/executable_identity_cases.json`
- No CHANGELOG touch (`git diff --name-only` = `internal/scriptworker/exec.go` only; no CHANGELOG in untracked set).
- No stray files beyond the three listed (product fix + committed test + fixture).
- No trunk intersection on the fix: control-root `curator/internal/scriptworker/exec.go:76`
  still reads `if platform == "windows" && useDefaultSearchDirs {` (unfixed);
  base `5b326aa3` is an ancestor of control-root HEAD `a48f584c`.

## Fix (accepted CR rev 1, already in worktree)
- `internal/scriptworker/exec.go:79`: hard-link trust root now requires
  `platform == "windows" && useDefaultSearchDirs && managerEnvironment != nil`,
  so the System32 hard-link allowance applies only when System32 derives from
  the manager-captured SYSTEMROOT. Nil (ambient) environments stay untrusted.

## Evidence (all run in this session, shell `bash`, `set -o pipefail`)
- `go test ./internal/scriptworker/ -run TestExecutableIdentityCasesAtProductionEntry -count=1 -v` → EXIT 0,
  8/8 subtests PASS including `windows-exec-uncaptured-systemroot-hardlinks`
  (driven at production entry via `deriveProfileForPlatform(..., "windows")`).
- `go test ./internal/scriptworker/ -count=1` (full package) → EXIT 0.
- `go vet ./internal/scriptworker/` → EXIT 0; `go build ./...` → EXIT 0.
- Mutant (drop `&& managerEnvironment != nil`): uncaptured subtest FAILS with
  `production resolver accepted = true, want false (reason systemroot-not-manager-captured; platform_owned=true)`,
  EXIT 1 → mutant killed. Fix restored afterwards and narrow gate re-run green (EXIT 0).
- Gap ledger: no file under the worktree retains an uncaptured-systemroot gap row;
  only references are the new fixture + test. The test's preserved-gap switch covers
  only `windows-exec-noncomponent-store-hardlinks` and `windows-exec-unowned-file-hardlinks`.

## DoD mapping
- [x] windows-exec-uncaptured-systemroot-hardlinks rejected at the production entry on windows;
      other identity cases unchanged; gap row removed; mutant killed (results above)
- [x] Implementation matches AC (conjunctive captured-SystemROOT condition)
- [x] Solution fits project architecture (resolver-level trust root, no new surface)
- [x] Tests green (narrow + full package, vet, build)
- [ ] If review does not accept the work — N/A (rev 1 accepted; this is the bound integration check)
- [x] Code written per task description and AC (in worktree, uncommitted for landing)
- [x] Relevant tests written for new/changed behavior and passing
- [x] Lint clean (`go vet` exit 0)
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached (this file)
- [x] Findings recorded here (logbook edit forbidden; this resource is the record)

## CHANGELOG entry (for release prep)
- `scriptworker`: Windows exec hard-link allowance now requires System32 derived from
  the manager-captured SYSTEMROOT; uncaptured roots reject multiply-linked targets
  (`windows-exec-uncaptured-systemroot-hardlinks`).

## Handoff
No `handoff` or `set_status` executed in this run per the binding assignment.
Runner performs the bound landing; orchestrator delivers on refusal.
