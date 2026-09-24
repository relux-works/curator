# BUG-260922-6chzf9 — make the managerlock tiny-deadline row deterministic (curator)

Control root: /Users/administrator/Developer/ReluxWorks/curator/curator; work only in your assigned Story
worktree (STORY-260923-vkxt08 ci-flake-fixes). Read `campaign-producer-rules.md` first. Landing gate =
hosted CI, run once by the runtime at handoff.

## Why this is urgent
`internal/managerlock` `TestSubprocessExpectedAcquiredWithTinyDeadlineReportsBlocked`
(`internal/managerlock/managerlock_test.go:525-534`) has now failed the hosted **windows-latest** lane at
least FOUR times on candidates that do not touch `internal/managerlock` (gate runs 35702368558,
35709048693, 35735392681, 35852658095), each time with
`managerlock_test.go:532: uncontended helper with tiny deadline = "acquired", want blocked`. Every
failure costs the campaign a full republish cycle (~40–60 min).

## The defect
The row starts an UNCONTENDED helper with a 1 ns deadline and asserts the helper reports `blocked`. It
assumes the deadline always expires before an uncontended acquisition completes. On a fast Windows
runner the acquisition wins the race and the helper reports `acquired`. The row tests a timing
accident, not the contract.

## Deliverable
1. Read the production code the helper exercises (`runHelperWithDeadline`, the `try-project` helper
   path, and the deadline handling in `internal/managerlock`) and state precisely what contract the row
   is meant to guard: "an acquisition attempt whose deadline has ALREADY expired reports blocked and
   takes no lock", or something else — cite the code.
2. Make the row deterministic WITHOUT weakening that contract. Acceptable shapes (pick the one the code
   supports, say why): (a) an already-expired deadline checked BEFORE any lock attempt (if production
   does that, assert it with a zero/negative deadline, not a race); (b) CONTEND the lock — hold it from
   another process, then assert the tiny-deadline helper reports blocked; (c) inject the clock. Do not
   delete the row, do not loosen it to accept both outcomes, do not add a retry loop.
3. If the analysis shows production itself may take the lock after its deadline expired (a real
   defect, not just a test race), fix production minimally and name it.
4. Prove determinism: run the row with `-count=200` locally (state the shell and exit code), and keep
   the sibling `TestSubprocessExpectedAcquired…`/blocked rows green. Add a narrowing mutant that breaks
   the guarded contract and is killed by the rewritten row.
5. CHANGELOG (unreleased, Fixed — test determinism). Attach `BUG-260922-6chzf9_results.md` and hand off
   with `task-board handoff BUG-260922-6chzf9 --role developer`.
