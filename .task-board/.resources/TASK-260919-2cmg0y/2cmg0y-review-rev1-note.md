# Review note for TASK-260919-2cmg0y revision 1 (orchestrator, binding)

Brief 2cmg0y-brief.md (rulings R1–R4). This is the FINAL leaf of STORY-260919-37szes: on
ACCEPT the orchestrator integrates the whole Story. Provenance: the producer's candidate was
captured as `TASK-260919-2cmg0y_rev1-candidate.patch` on the pre-refresh branch (tip 4f213e77),
the Story branch was then replayed onto trunk d4fe8347 (CHANGELOG union-merged: the Story's
entries + trunk's 30ycv0 entry), and the same bytes were re-applied (results.md "Re-apply"
section) — verify: the CR patch equals the captured patch for every path except CHANGELOG.md
(compare per-file patch-ids), and the gate commit (run 35563543978) resolves to the exact
revision-1 tree.

Judge with your own reruns (disposable clone; bounded commands; retry once on host stalls):
1. R1 cache: per-transaction, invalidated by the per-write boundary recheck — find the seam
   the producer cites and prove the invalidation with the committed swapped-symlink row; run
   mutants "drop the invalidation" and "cache across transactions/engines"; confirm no cache
   when the recheck is absent; every namespace/boundary negative row green.
2. R2 fsync: the producer recorded "no change, bound" — confirm no durability path changed
   (diff `durability_*.go`, `files.go`, `staging.go`, `journal.go` sync calls).
3. R3: no proof weakened — the sweep `TestDraftFailureAtEveryTargetClassRestoresPriorState`
   and every rollback row unchanged in assertions; no new skip; ledger vocabulary unchanged.
4. R4 measurement: the AC is ≥ 2× on the Windows hosted gate for the sweep. Extract the
   elapsed of the sweep from `test-evidence-windows-latest` of gate run 35563543978 and of
   the baseline (run 35555033097, BUG-2d9gfv rev2, same Story) plus the runner-speed proxy
   (sum of top-level `internal/install` tests present in both) and state the raw and
   normalized ratio yourself; the producer's numbers are claims until you reproduce them.
5. Correctness of `canonicalNamespacePath` semantics under the cache: same input → same
   resolution or the same error class; missing-prefix resolution unchanged; Windows path
   case/separator handling unchanged (rows run on windows-latest).
Record exactly one verdict: accept_cr(TASK-260919-2cmg0y, revision=1, evidence=<your outcome
resource>) on ACCEPT, or changes_requested with file:line and reproduction.
