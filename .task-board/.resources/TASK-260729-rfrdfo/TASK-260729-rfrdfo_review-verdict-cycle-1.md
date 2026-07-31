# TASK-260729-rfrdfo — review verdict cycle 1

Date: 2026-07-29  
Role: reviewer  
Verdict: **CHANGES REQUESTED → analysis**

No product or test code was modified. No Go, test, race, build, vet, format, lint,
install, stage, commit, publish, or pin command was run during this review.

## Acceptance-blocking finding R1

Patch A is accepted as effective. Patch B is not sufficient for the authoritative
section 7 acceptance bar.

All six focused race repetitions have real exit code 0:

| Package | Repetition 1 | Repetition 2 | Repetition 3 | Result against 480s |
| --- | ---: | ---: | ---: | --- |
| `internal/install` | 232.088s | 235.124s | 226.191s | PASS by 244.876–253.809s |
| `internal/install/atomicity` | 591.280s | 560.828s | 564.022s | MISS by 111.280s, 80.828s, and 84.022s |

The atomicity result is not a marginal pass. The worst focused run leaves only
8.720s below the unchanged 600s package alarm. The authoritative plan states
that focused package runs are optimistic relative to the concurrent `./...`
race gate and deliberately requires both packages to print `≤480s`. Three
independent atomicity misses therefore fire section 9 risk 4 and reject the
current Patch B sufficiency claim even though the raw test processes exited 0.

## Independently accepted evidence

- The patch contains exactly the 13 allowlisted `_test.go` paths. The task-owned
  pre/post manifests report 13 modified, 0 added, 0 deleted, and 0 unexpected
  paths. `internal/install/aba_test.go`,
  `internal/install/atomicity/fixture_test.go`, and product Go remain untouched.
- A reviewer dry-run applied all 13 patch entries to the exact current
  `TASK-260720-jrrgw9` candidate with exit 0 and no fuzz, offset, reject, failed,
  or malformed marker. A fresh 391-file path-sorted SHA-256 manifest matched the
  producer evidence before the dry-run and remained byte-identical afterward.
- The live prototype still matches the recorded 391-file final manifest. The
  recorded third-copy apply manifest is byte-identical to the gated prototype
  post-manifest.
- The current source candidate reproduces the accepted-worktree delta at 23
  lines. The prototype reproduces the corrected 34-line post-delta: 20 deleting
  plus 14 modified, consisting of 3 pre-existing and 11 newly modified entries.
- The immutable conformance root still matches the recorded 448-file digest
  manifest byte for byte.
- All 13 recorded gate exit files contain 0; every `go test` template carries
  `-count=1`; the driver contains no timeout override and no `go test ./...`;
  no gate log contains `DATA RACE`. The gofmt and vet evidence is clean.
- Patch A marks the sanctioned 88 `internal/install` tests while retaining the
  derived 19-test sequential set. Patch B separately parallelizes the seven
  approved non-sweep atomicity tests.
- `injectClasses` is a distinct injection selector while `classes` retains the
  full project 7-class and global 5-class success-coverage lists. The scenario
  table is generated from those full lists and each global entry has a distinct
  user home.
- The whole-state digest, failing-class cutoff, reverse rollback, journal
  cleanup, and full success coverage assertions remain. The 31 ordered
  cross-class sequencing pairs intentionally fall to zero and that retired
  defense-in-depth property is documented both at the source call site and in
  the outcome.

## Required bounded rework

1. Preserve Patch A, the 88/19 split, the 13-file integrity/applicability
   evidence, the assertion map, and the explicit cross-class retirement record.
2. Do not silently widen this task beyond its exact 13-file scope. Route
   analysis for the sanctioned next lever: either create a separate task with a
   declared 14-file allowlist for the `references/info.md` fixture reduction,
   attach before/after `StagingEntries` measurement, and provide a defensible
   expected `≤480s` focused result; or conclude that the test-only boundary is
   exhausted and route the separate production-side `saveJournal`
   namespace-validation fix.
3. Do not change the timeout, skip a case, weaken an assertion, run the full
   repository race gate in this prototype task, or mutate the current candidate.
4. Any successor implementation must return through a new reviewer cycle with
   count-one focused evidence that both packages satisfy the 480s condition.

This is recoverable research and implementation planning, not a genuine
stop-the-line boundary. The correct verdict route is `analysis`, not `blocked`.

## Evidence inspected

- `TASK-260729-rfrdfo_results.md`
- `TASK-260729-rfrdfo_install-race-timeout.patch`
- `TASK-260729-rfrdfo_evidence-bundle.tgz`
- `TASK-260729-rfrdfo_reconciliation-notes.md`
- `TASK-260729-3dr6hw_install-race-timeout-diagnosis.md`
- `TASK-260729-3dr6hw_review-verdict-cycle-5.md`
- Current jrrgw9 candidate, prototype, accepted-worktree, and conformance-root
  manifests and dry-run applicability

