# Review note for TASK-260916-1xib1x revision 5 (orchestrator, binding) — R4 audit labels

Revision 5 = revision 4 = revision 3 BYTES (two gate reruns: a Windows `internal/managerlock`
timing flake, then an ubuntu `internal/install` package timeout on a 2× slow runner — both
environmental, filed as BUG-260922-6chzf9 and BUG-260922-3v8k23). So this is the FIRST review of the
rework-2 content: revision 3's tree has never been reviewed.

Revision 3 = rework 2 for your revision-2 verdict (`TASK-260916-1xib1x_review-verdict-rev2.md`,
single finding: `internal/audit/audit.go:255–259` returned on a findings-cache hit before the only
call that persists `script_policies`, so a valid pre-R4 cached verdict never gained the per-command
record; reviewer row `TestReviewerExistingVerdictGetsPolicyRecord`). Brief:
`1xib1x-rework-2.md` — refresh/backfill on a writable cache hit OR invalidate old-format verdicts,
while preserving cached-findings semantics and `GateReadOnly`'s no-write guarantee (read-only mode
must still REPORT policies from the current manifest without writing), with the reviewer's row plus
a mutant (skip the backfill ⇒ row fails).

Gate: run 35717146238 succeeded — verify its head resolves to the exact revision-5 candidate tree
and reuse it; do not rerun the full landing suite. Diff revision 2 → revision 5 and judge only that
scope; everything accepted at revision 1/2 (the four vector shapes, per-command policy record,
global-install rows) stays accepted unless the new delta broke it.

What to decide:
1. Re-run your own `TestReviewerExistingVerdictGetsPolicyRecord` against revision 5 — it must pass.
2. `GateReadOnly` must still write NOTHING on a cache hit while reporting the current policies;
   prove it (a read-only gate over an old-format verdict: record reported, store bytes unchanged).
3. The backfill must not lose or rewrite cached FINDINGS (identity/version semantics intact), and
   must not turn a clean verdict into a warning or vice versa.
4. Attack it: a mutant that skips the backfill must kill the named row; add one attack of your own
   (e.g. an old-format verdict under a manifest whose commands changed since — the record must
   reflect the CURRENT manifest, not the stale one, or the producer must state that bound).
5. Check that nothing in the delta silently weakened the revision-1/2 rows.

Bounds you may keep: the platform-cases ledger registration is R5 scope (TASK-260916-2ok97n), not
this leaf; hosted-platform claims are limited to the exact-tree gate evidence.

Record exactly one verdict: `accept_cr(TASK-260916-1xib1x, revision=5, evidence=<your outcome
resource>)` on ACCEPT, or a changes-requested verdict routed with `set_status` naming file:line and
an executable reproduction. Do not write into the control root's LOGBOOK.md.
