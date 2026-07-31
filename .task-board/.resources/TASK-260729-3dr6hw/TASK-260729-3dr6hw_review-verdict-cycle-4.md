# TASK-260729-3dr6hw — review verdict cycle 4

Verdict: CHANGES REQUESTED → analysis.

## Blocking finding

R4-1 — The cycle-4 evidence repair mischaracterizes the censored atomicity ratio as carrying no information.

Verifier-3 records `internal/install/atomicity` at `ok ... 441.122s` without race and `FAIL ... 603.701s` under race. Because the race run timed out before the binding sweep completed and before three post-sweep tests started, the completed race duration is strictly greater than 603.701 seconds. Therefore the completed-package factor is strictly greater than `603.701 / 441.122 = 1.3685`, conventionally reported as a weak lower bound `>×1.37`.

The ratio cannot be treated as a completed factor and cannot support a comparison with `cmd/curator` or the five completed packages. It is nevertheless not information-free. The main diagnosis repeats the incorrect stronger statement in the cycle-4 change table and section 3.5, while treating the analogous censored `internal/install` ratio as a genuine floor. This internal inconsistency violates the fact-checking and exact-evidence acceptance criteria.

## Required bounded rework

1. Replace every claim that the atomicity ratio carries no information with the precise statement: it is a censored lower bound `>×1.37`, too weak for cross-package comparison or for estimating the completed factor.
2. Remove or qualify the statement that its position below `cmd/curator` is purely due to truncation. The observed value is low because the numerator is truncated; the completed factor comes from the separate phase-level evidence in section 3.4.
3. Apply the same correction to the cycle-4 rework record and all current-summary passages, preserving the accepted phase-level solve and all timing bands.
4. Preserve sections 1, 2, 4 through 9, both allowlists, the assertion/invariant tradeoff, focused validation commands, and candidate-integrity checks unchanged.

## Independently accepted evidence

- Exact verifier evidence is correct: `cmd/curator` passed race at 557.779s; `internal/install` timed out at 603.306s with `TestStrictRegistryPolicyFailsUnknown` active; atomicity timed out at 603.701s with the two sweep scenarios active; no DATA RACE marker exists.
- Static ordinal inventory confirms `TestStrictRegistryPolicyFailsUnknown` is test 73 of 107 and atomicity has 8 top-level tests with the sweep fifth.
- The section 3.5 selection rule independently yields exactly 38 `ok/ok` packages and five with `max(non-race,race) >= 10s`: curator ×1.45, godriver ×1.39, transaction ×1.25, runtimestore ×0.98, managerlock ×2.90. The all-package extrema ×0.78 and ×8.47 are correct.
- Source independently confirms `internal/transaction` has 50 tests, 33 `Prepare` sites, two `exec.Command` sites that both re-execute the test binary, and no git or go-build test invocation. The old subprocess attribution is withdrawn consistently and the target-count explanation is labelled as an unmeasured hypothesis.
- The Patch A partition is exact: 107 source tests = 88 literal allowlisted parallel tests + 19 literal sequential exclusions, with no overlap or omission.
- Patch B is implementable against the current source: `injectClasses` separates injection from full coverage, per-class scenarios retain every assertion, distinct global user homes satisfy `globalbins.Select` isolation, and the retired 31-pair cross-class residue chain is disclosed with its residual risk and fallback.
- The 13-file allowlist, 23-to-34 accepted-worktree delta expectation, pre/post SHA-256 manifest, focused package commands, two-scan non-overlap barrier, unchanged timeout, and 480-second producer bar are coherent.

No candidate file was edited and no Go command was run during this review.