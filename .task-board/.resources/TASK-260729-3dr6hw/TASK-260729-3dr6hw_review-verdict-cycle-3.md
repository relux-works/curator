# TASK-260729-3dr6hw review verdict — cycle 3

Verdict: **CHANGES REQUESTED to analysis**. No candidate file was edited and no Go command was run.

## Independently accepted

- The verifier-3 inventory is exact: `cmd/curator` passed at 557.779s; `internal/install` timed out at 603.306s with `TestStrictRegistryPolicyFailsUnknown` active for 3s; `internal/install/atomicity` timed out at 603.701s with both sweep scenarios active for 8m28s. The race log has no `DATA RACE` marker.
- Source order and the alarm dump confirm 107 `internal/install` tests, 72 completed tests, and `TestStrictRegistryPolicyFailsUnknown` as test 73. Atomicity has eight top-level tests and the reported project/global class positions match source.
- Cycle-2 corrections 1–4 are resolved:
  - the accepted-worktree delta has 20 `*deleting` plus 3 modified entries; the two install conformance files are candidate-only, so the corrected post-patch expectation is 11 new comparable-file entries and 34 total;
  - `saveJournal` has 16 call sites in `internal/transaction/engine.go` and 7 in `staging.go`, including the two copy-loop sites at 141 and 161;
  - Patch B now has a separate `injectClasses` selector while retaining full `classes` for the coverage assertion, generated per-class scenarios, and distinct global user homes;
  - the 31-to-0 cross-class sequence reduction is disclosed as intentionally retired defence in depth, with the retained direct snapshot, rollback-order, journal-cleanup, and full-success assertions named.
- The 88/19 Patch A partition exactly equals the candidate's 107 test functions minus the 19 named process-global/helper hazards. No name is missing or duplicated.
- Patch B is conceptually implementable against `commit_atomicity_test.go`: the injection loop and coverage assertion are distinct reads at lines 163 and 206; `env.userHome` exists; `globalbins.Select` prefers each scenario's own user bin and rejects sibling homes in its fallback.
- The 13-file producer allowlist, 220–340s / 340–460s projected package bands, unchanged 600s alarm, focused producer commands, and pre/post candidate manifest are suitable handoff inputs. They remain projections and must be validated by the downstream producer/verifier.

## Required correction

Section 3.5, newly added in cycle 3, is not factually sound as written.

1. It says the table lists every completed package large enough for a meaningful ratio and excludes only sub-10s packages, but omits `internal/runtimestore`: verifier-3 records 18.630s non-race and 18.244s race, or approximately **×0.98**. Add it or state and justify a different selection rule; withdraw the word “every” if the table remains selective.
2. It attributes the low factors of `cmd/curator`, `internal/godriver`, **and `internal/transaction`** to time spent in `git` / `go build` subprocesses. Static source contradicts that attribution for `internal/transaction`: its roughly 50 tests are overwhelmingly in-process transaction/filesystem tests; the only `exec.Command` uses found are two test-binary re-executions in `subprocess_test.go`, with no `git` or `go build` invocation. Reframe the same-run evidence as descriptive corroboration only, or use the transaction result to discuss target-count/workload shape. Do not claim a subprocess mechanism without package-level timing/profile evidence.

The corrected section may still say the exact-candidate gate shows the two timed-out install packages have materially larger race inflation than the completed packages. It must distinguish that observation from the unprofiled causal explanation.

## Evidence

- `.temp/TASK-260720-jrrgw9/verifier3/go-test-race-all.log`
- `.temp/TASK-260720-jrrgw9/verifier3/go-test-all.log`
- `.temp/TASK-260720-jrrgw9/verifier3/candidate-source-delta-post.txt`
- `.temp/TASK-260720-jrrgw9/verifier3/candidate-delta-digests-post.txt`
- `.temp/TASK-260720-jrrgw9/worktree/internal/install/atomicity/commit_atomicity_test.go`
- `.temp/TASK-260720-jrrgw9/worktree/internal/install/atomicity/fixture_test.go`
- `.temp/TASK-260720-jrrgw9/worktree/internal/globalbins/globalbins.go`
- `.temp/TASK-260720-jrrgw9/worktree/internal/transaction/*_test.go`
- `.temp/TASK-260720-jrrgw9/worktree/internal/transaction/engine.go`
- `.temp/TASK-260720-jrrgw9/worktree/internal/transaction/staging.go`

## Routing

This is one bounded research/evidence repair. Update the existing diagnosis outcome in place, attach a new cycle-4 rework record, and return for independent review. Preserve the accepted plan, timings, allowlists, risk disclosures, and cycle-2 corrections. `Tests green` remains unchecked because this diagnosis task forbids Go execution and delivers no implementation; focused/race validation belongs to the downstream patch producer.
