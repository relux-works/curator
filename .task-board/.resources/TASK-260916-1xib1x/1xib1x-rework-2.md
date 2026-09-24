# TASK-260916-1xib1x rework 2 (orchestrator, binding) — R4

Verdict rev2: CHANGES_REQUESTED with ONE finding (TASK-260916-1xib1x_review-verdict-rev2.md;
reviewer row `TestReviewerExistingVerdictGetsPolicyRecord`): `internal/audit/audit.go:255–259`
returns on a findings-cache hit before the only call that persists `script_policies` (:263–264),
and the cache identity/version was not changed — a valid pre-R4 cached verdict never gains the
per-command policy record. Fix: either refresh/backfill the persisted record on a writable cache
hit, or bump the verdict cache identity/version so old-format verdicts are regenerated — keep the
cached-findings semantics and `GateReadOnly`'s no-write guarantee (read-only mode must still
report policies from the current manifest without writing); commit the reviewer's row plus a
mutant (skip the backfill → row fails). Continue from the revision-2 tree (no checkout/clean/
stash); append "Revision 3" to results.md; republish only on a green gate.
