# TASK-260729-3dr6hw — reviewer verdict, cycle 1

Date: 2026-07-29
Role: reviewer
Verdict: **changes requested → analysis**

## Outcome

The diagnosis is technically strong but does not yet satisfy the task's
candidate-integrity acceptance criterion. The exact timeout evidence, package
and scenario inventory, static hot-path attribution, process-global exclusions,
assertion-preservation map, and test-only optimization direction were
independently checked and are suitable to preserve in the next revision.

No Go command or candidate edit was performed by this reviewer.

## Blocking finding — the proposed candidate-integrity checks cannot work on this candidate

Section 8 requires the future reviewer to accept only when `git diff --stat`
shows no path outside `internal/install/` and no non-test Go file. The actual
candidate is already an intentionally dirty delivery worktree at HEAD
`17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`: the independent review found 37
tracked modified files plus many untracked files, including product files
outside `internal/install/`. Therefore those raw HEAD-relative checks fail
before the proposed patch exists and cannot distinguish the patch from the
pre-existing candidate.

The report's 448-file digest check does not repair this gap. Verifier-3's
`authoritative-digests*.txt` covers the immutable conformance root, while the
candidate source is tracked separately by
`candidate-source-delta-post.txt` and `candidate-delta-digests-post.txt`.
An unchanged HEAD also says nothing about edits inside an already dirty
worktree. As written, an unrelated candidate edit could be hidden among the
existing delta while all proposed conformance-root checks still pass.

This is an ordinary research-plan defect, not a stop-the-line boundary.

## Required rework

1. Replace the raw HEAD-relative integrity gate with a pre/post content baseline
   for the actual candidate. Reuse verifier-3's accepted comparison worktree
   `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-2kaopg/worktree`
   and its candidate delta/digest artifacts, or record a task-owned,
   path-sorted SHA-256 manifest immediately before implementation. The post-patch
   comparison must prove that the only bytes changed from that captured
   candidate baseline are the literal required test-file allowlist.
2. Give the producer and reviewer literal commands for generating and comparing
   that baseline, including added/deleted-path detection. Keep the immutable
   448-file conformance digest check as a separate invariant; do not present it
   as candidate-source integrity.
3. Tighten the required file allowlist. `internal/install/aba_test.go` is
   explicitly unchanged and must not appear as a patch target. Likewise,
   `internal/install/atomicity/fixture_test.go` belongs only to the optional,
   unmeasured `references/info.md` lever and must be excluded from the smallest
   required patch unless that optional lever is separately justified and
   adopted.
4. Correct the runtime-model wording: the two `internal/install` “independent
   derivations” are algebraically the same uniform-cost extrapolation, not
   independent corroboration. Retain the estimate as an explicitly
   assumption-based projection and include cross-package contention from the
   full `go test -race ./...` gate in the uncertainty discussion. This does not
   invalidate the proposed parallelization, but it must not be overstated.
5. Update the task-scoped diagnosis outcome and task notes/log record, then hand
   off for a fresh reviewer cycle. No product or test code change and no Go test
   execution is needed for this rework.

## Independently verified evidence to preserve

- Race log: `cmd/curator` passed in 557.779s;
  `internal/install` timed out at 603.306s with
  `TestStrictRegistryPolicyFailsUnknown` active for 3s; atomicity timed out at
  603.701s with project/global sweep parents active for 8m28s and class
  subtests active for 52s/44s. No `DATA RACE` marker is present.
- Non-race log: `internal/install` passed in 341.415s and atomicity in 441.122s.
- Static inventory: `internal/install` has 107 top-level tests; exactly 70
  precede `registry_e2e_test.go`. Atomicity has eight top-level tests and sweep
  class sets of seven project classes and five global classes.
- The three runnable alarm goroutines independently land in
  `namespaceComponents → namespaceContains → namespacePathsOverlap →
  validateIndependentTargetNamespaces → saveJournal`, with target counts 8,
  19, and 20. `saveJournal` invokes the namespace validation through both
  `validateJournal` layers, and staging writes call `saveJournal` repeatedly.
- The proposed `t.Parallel` exclusions correctly identify the process-global
  `t.Setenv` cases, the unsynchronized `afterDocumentOpen` hook users, and the
  helper-process test. The atomicity scenario namespaces are based on distinct
  `t.TempDir` roots, and the whole-state snapshot assertion directly checks
  rollback residue after every class injection.
- Existing verifier evidence is honestly mixed: the full non-race suite is
  green, while the required race suite is red solely on the two package
  timeouts. No new tests were authorized or run during diagnosis or review.

## Sources checked

- `.temp/TASK-260720-jrrgw9/verifier3/go-test-race-all.log`
- `.temp/TASK-260720-jrrgw9/verifier3/go-test-all.log`
- `.temp/TASK-260720-jrrgw9/verifier3/TASK-260720-jrrgw9_final-verifier3-results.md`
- `.temp/TASK-260720-jrrgw9/verifier3/candidate-source-delta-post.txt`
- `.temp/TASK-260720-jrrgw9/verifier3/candidate-delta-digests-post.txt`
- Candidate `internal/install/**/*_test.go`
- Candidate `internal/transaction/{namespace,journal,staging}.go`

