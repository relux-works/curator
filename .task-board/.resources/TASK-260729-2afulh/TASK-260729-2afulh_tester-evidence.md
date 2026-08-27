# TASK-260729-2afulh tester evidence

## Verdict

Reject the 14-file fixture trim as the atomicity timeout solution. It improves
the workload materially, but the valid count-one race gate took 493 seconds and
missed the strict 480-second acceptance bar by 13 seconds.

## Candidate integrity

- Fresh source, baseline, and prototype path-sorted SHA-256 manifests: exit 0,
  391 files in baseline and prototype.
- Fresh baseline versus the task's recorded baseline manifest: `diff` exit 0.
- Fresh baseline versus TASK-260729-rfrdfo's recorded final manifest: `diff`
  exit 0.
- Literal 14-file integrity checker: exit 0,
  `modified=14 added=0 deleted=0 unexpected=0 forbidden_touched=0`,
  `ALLOWLIST_SIZE=14`, `INTEGRITY_OK`.
- Corrected baseline-to-prototype delta checker: exit 0; only
  `./internal/install/atomicity/fixture_test.go` changed, with no additions or
  deletions.

The first delta-check one-liner exited 1 because it keyed manifest rows by
digest instead of path. The corrected command above exited 0. This was a
checker-command mistake and did not write to the candidate.

## Fixture neutrality and isolation

`bin/neutrality.sh` exited 0 and reported identical before/after counts for all
assertion, test, subtest, parallelism, temp-dir, environment, skip, and timeout
surfaces. The only count change was the intended fixture write:
`e.write(` 8 to 7.

Direct `cmp -s` processes each exited 0 for:

- `internal/install/atomicity/commit_atomicity_test.go`
- `internal/install/atomicity/activation_test.go`
- `internal/install/atomicity/doc.go`
- `internal/install/install_test.go` (inherited Patch A fixture)

No non-comment `references/info.md` occurrence remains in the atomicity
package. Patch A's separate `internal/install/install_test.go` occurrence is
retained.

## Staging measurements

The implementer's A/B scratch tests both passed as standalone `go test`
processes: before in 61.271s and after in 54.448s. Entries and non-empty chunks
were observed; staging-path `saveJournal` calls are exact arithmetic over those
observations (`3*entries + 2*chunks`) because the unexported method has no
observation hook and the literal 14-file deliverable forbids a production seam.

| scenario | phase | entries | non-empty chunks | exact staging saves |
| --- | --- | --- | --- | --- |
| project | baseline | 24 -> 20 | 14 -> 12 | 100 -> 84 |
| project | upgrade | 34 -> 28 | 19 -> 16 | 140 -> 116 |
| global | baseline | 26 -> 22 | 16 -> 14 | 110 -> 94 |
| global | upgrade | 25 -> 21 | 16 -> 14 | 107 -> 91 |

## Fresh focused gates

Every executed gate had a separate two-scan `BARRIER_OK`.

| command | real exit | wall time | evidence |
| --- | ---: | ---: | --- |
| `gofmt -l internal/install internal/install/atomicity` | 0 | 0s | empty output |
| `go build ./...` | 0 | 1s | empty output |
| `go vet ./internal/install/...` | 0 | 0s | empty output |
| `go test -count=1 -v ./internal/install/atomicity` | 0 | 273s | PASS |
| `go test -count=1 -race ./internal/install/atomicity` | 0 | 493s | `ok ... 492.231s`; no `DATA RACE` |

Race repetition 2 received SIGTERM before an exit file was written; its partial
log is not evidence. At the next safe checkpoint the tester observed
orchestrator directive `RUN-260729-b2a441:nudge:3afa0e`, which required stopping
after the first valid race run already made the strict acceptance predicate
false. Repetition 2 was not rerun and repetition 3 was not started, releasing
the shared Go slot for TASK-260729-365r5r. `gates/DRIVER-STOP-REASON` documents
the cooperative stop; the annotated `DRIVER-DONE` is manual and does not claim
natural driver completion.

The first implementer driver directory is preserved as
`gates-partial-RUN-260729-7ade05/` and is excluded from the verdict.

## Additional command honesty

- Direct `bin/report.sh`: exit 126, permission denied (prepared helper lacked
  executable bit).
- `bash bin/report.sh`: exit 0, report rendered.
- Post-stop process barrier: exit 0, both scans empty.
- Later barrier before lint: exit 1 because TASK-260729-365r5r had correctly
  claimed the shared slot. No overlapping lint was started.

## Acceptance decision

The measured staging reduction is genuine, and the race gate improved from the
inherited 561-593s range to 493s. Nevertheless 493s is 13s above the strict
480s limit and has negative margin. The fixture-only trim must be rejected as
the timeout solution and independently reviewed as evidence for the
production-side `saveJournal`/namespace-validation successor.
