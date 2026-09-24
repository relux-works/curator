# BUG-260924-5p8b0z — integration precondition check (RUN-260924-4392f2, bound developer run)

Board status observed: `integrating` (left untouched; no status/handoff writes per binding).
`task-board worktree integrate` was NOT executed in this run per the Integration Assignment
(runner performs the bound landing synchronously). No file changed in this run.

## Landing preconditions (verified read-only)
- Worktree: `.temp/STORY-260924-2tyzhh/worktree`, branch `task-board/story/STORY-260924-2tyzhh`, base `5b326aa3` is ancestor of HEAD.
- `git status --short` contains exactly:
  - `M internal/scriptworker/exec.go`
  - `?? internal/scriptworker/exec_identity_conformance_test.go`
  - `?? internal/scriptworker/testdata/executable_identity_cases.json`
- No CHANGELOG touch (`git diff --name-only` = `internal/scriptworker/exec.go` only).
- No stray files beyond the three listed.
- Accepted CR rev 1 already in worktree: `internal/scriptworker/exec.go:79`
  `if platform == "windows" && useDefaultSearchDirs && managerEnvironment != nil`,
  so the System32 hard-link allowance applies only when System32 derives from the
  manager-captured SYSTEMROOT. Test drives production entry via
  `deriveProfileForPlatform(..., "windows")` (test:169-174).

## Fresh evidence (this run, shell `bash`, `set -o pipefail`)
- `go test ./internal/scriptworker/ -run 'TestExecutableIdentityCasesAtProductionEntry' -count=1 -v` → EXIT 0,
  8/8 subtests PASS including `windows-exec-uncaptured-systemroot-hardlinks`.
- `go vet ./internal/scriptworker/` → EXIT 0.
- `git status --short` after verification: same three paths, tree otherwise clean.

## Accepted from already-attached evidence (not rerun here)
- Mutant (drop `&& managerEnvironment != nil`) killed: review verdict rev1 §4 reran it,
  uncaptured subtest FAILS `accepted = true, want false` (EXIT 1). Not re-applied here
  per the `Change no file` binding.
- Hosted Windows lane: validation log run 35987255178 exit 0, `Test (windows-latest): success`;
  review verdict §5 confirms windows-latest gate shows parent + all 8 subtests passing.
- Full package + vet + build green per review verdict §3 and prior `integration-final.md`
  (full `internal/scriptworker` pass, `go vet` clean).

## CHANGELOG entry (for release prep; no CHANGELOG edit in this leaf)
- `scriptworker`: Windows exec hard-link allowance now requires System32 derived from
  the manager-captured SYSTEMROOT; uncaptured roots reject multiply-linked targets
  (`windows-exec-uncaptured-systemroot-hardlinks`).

## Handoff
No `handoff`, `set_status`, or `worktree integrate` executed per the binding.
Runner performs the bound landing; orchestrator delivers on refusal.
