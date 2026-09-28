# TASK-260922-1ejkxv — F-S3: fleet-enforced isolated credential mode (curator-spec)

Control root: curator-spec; work only in your assigned Story worktree
`.temp/STORY-260922-188t6n/worktree`. Read `campaign-producer-rules.md` first. Gate = the board's
validation command (runtime runs it once at handoff).

## Source of truth
Decision 0017 adoption choice 6: "**Not in this adoption — follow-up.** The system schema still
locks only toward `shared`; enforcing `isolated` fleet-wide needs a reviewed policy revision, not a
knob reinterpretation. Named follow-up." This leaf is that revision.

## What exists
- `protocol/environments.md` §12.1 (the `isolation.<profile>.<env-id>` knob — per 0017 choice 5 it
  stays the ONLY config field for the mode) and §12.2 (the lockable set and the direction a lock may
  force).
- `profiles/manager.md` §1 (the system `locked` list semantics).
- `system-config-v2` schema + its cases (find them under `schemas/` and
  `conformance/v1/schema-cases/`; follow the existing naming).

## Deliverable
1. Admit the `isolated` lock direction in `system-config-v2` with exactly the manager §1 locked-list
   semantics that `shared` already has — same shape, same precedence, no new knob.
2. Normative text: environments §12.2 states both directions and what each forces; a profile that
   requests `shared` under an engaged `isolated` lock is refused with a NAMED diagnostic (reuse an
   existing §7.7/§10.4 code if one fits, otherwise add one and register it in the diagnostics
   table); silence resolves to the locked direction. State what happens to an already-provisioned
   shared passthrough when the lock engages (refuse and tell the operator, or migrate — choose the
   fail-closed option consistent with 0017's "never silently migrate" rule) and cross-reference the
   migration path (F-C2's explicit migration in curator).
3. Schema cases: valid system config locking `isolated`; valid locking `shared` (unchanged);
   invalid unknown direction; invalid both directions for one key. Keep every existing case green.
4. `make validate` + regenerate-check green; CHANGELOG (unreleased).
5. results.md MUST name the curator follow-up leaf that enforces the lock (manager side) with the
   exact schema key and diagnostic name.

## Boundaries
No curator code. No marker-schema work (F-S1, TASK-260922-1nf6o6), no fragment work (F-S2,
TASK-260922-1hla8q). Frozen v1 protocol schemas untouched. Hand off with
`task-board handoff TASK-260922-1ejkxv --role developer` after attaching
`TASK-260922-1ejkxv_results.md`.
