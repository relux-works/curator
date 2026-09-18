# Rework brief — TASK-260910-2g5v17, revision 2 (P4)

Revision 1 passed every serial scenario, both mutants, atomicity, storage
protection and docs; one medium finding remains
(`TASK-260910-2g5v17_review-verdict-rev1.md`, F1). Keep everything else
byte-identical.

**F1 — the comparison under writer serialization must carry the full import
contract.** `bundle.py:149` reads the high-water outside the write
transaction; the authoritative recheck in `store.py:1219-1237` raises a
generic `ValueError("upstream high-water advanced during import")`, does not
receive the override flag, bypasses the outcome logging at `bundle.py:196`,
and `cli.py:271` prints the generic text. Three legal interleavings (two
real `Store` connections, the outer import paused at `append_imports` while
the other commits) therefore lose the closed diagnostics:
1. persisted v1, outer offers v2, competitor imports v3 first → must refuse
   with `import_upstream_rollback` naming `key_id`, the persisted (v3) and
   the offered (v2) boundary, with the structured refusal audit event;
2. same schedule with `--accept-older-upstream` → must warn and import
   WITHOUT lowering v3 (the override applies to the authoritative
   comparison too);
3. unknown upstream, competing first imports with different v1 bodies, the
   competitor commits first → `import_upstream_inconsistent` with key_id and
   both boundaries, never overridable, with the audit event.
Correction: perform the authoritative comparison inside the serialized
write transaction (BEGIN IMMEDIATE) with the same outcome logic as the
ordinary path — persisted boundary, offered boundary, key_id, override
policy, identical no-op — so the pre-transaction read becomes advisory only
(or drop it); emit the documented refusal/warning audit event for the
comparison that actually decided; preserve atomicity and never admit an
equal-version inconsistency. Add deterministic competing-writer regression
tests for the three schedules plus the equal-identical no-op under
competition (the reviewer's `probes.py` in the verdict shows the pausing
technique; commit an equivalent deterministic harness).

Validation as before (pytest with `CURATOR_CONFORMANCE_ROOT` at the
`47c3c8c` root per the corrected rules, mypy strict); results "Revision 2"
section; tick the checklist; `task-board handoff TASK-260910-2g5v17 --role developer`.
