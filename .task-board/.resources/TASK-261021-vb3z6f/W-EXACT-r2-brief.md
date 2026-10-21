# Rework round 2: TASK-261021-vb3z6f, Draft CIP-0023 (curator-spec)

The review (`TASK-261021-vb3z6f_review-r1.md`) requested changes: one P1.

**F1:** your stable-home lease is capture-only, so it can be released after admission and resource transfer. The CIP-0002 candidate (branch `cip02-ready-for-acceptance`; it is being reworked now, so fetch its current head) requires the original Werkstatt home lease, owned by the session, to stay active until termination and authority/state reconciliation. Your interface defines no separate retained session lease.

## Fix

- Define two distinct leases explicitly:
  - the **session lease** on the original Werkstatt home, held from admission until termination and reconciliation, exactly as CIP-0002 requires (busy-update refusal, no repair or switch underneath);
  - the **capture lease**, held only while the snapshot is taken, which pins the snapshot roots for GC separately.
- State the order, crash and unknown-liveness behaviour, and who releases each lease. A capture-lease release never releases the session lease.
- Cross-check every lease sentence against the current CIP-0002 candidate head and name that head in your results. If CIP-0002 and CIP-0023 still disagree, list the exact sentences instead of guessing.
- Add vectors or test-plan rows: session lease retained after capture release; busy update refused while the session lease is held; GC keeps snapshot roots under a live capture or session.
- Rerun `make validate` in a normal-.git clone.

## Hand-off

A new signed commit on the same branch `exact-plan-final-form` (no force-push), results r2 with a delta summary, and a handover with the new head.
