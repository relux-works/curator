# Brief — TASK-260910-2rsajv: idempotency retention with slack (R8, service)

Story `STORY-260910-2xe3n2` (registry-robustness-hardening, `EPIC-260910-16qce1`),
wave 4; second of four sequential leaves (after `TASK-260910-28kmef`, whose
accepted candidate is already checkpointed on the Story branch — build on
it). Rules: `remediation-registry-producer-rules.md` (attached). Role: developer.
Worktree: the managed Story worktree `<control-root>/.temp/STORY-260910-2xe3n2/worktree`.

## Finding (read it first)
`docs/security-audit-2026-09.md` R8 (Info): `IDEMPOTENCY_TTL_SECONDS = 24*3600`
meets the registry-service profile minimum ("at least 24 hours") with zero
slack; a client retry at the boundary plus network delay can double-append.

## Deliverable
1. `IDEMPOTENCY_TTL_SECONDS = 26 * 3600` in `app.py` with a comment citing
   the profile minimum (quote the §/sentence of `profiles/registry-service.md`
   at the pinned `dced9b8`) and the two-hour slack rationale; any place that
   documents the retention (README, SECURITY, docstrings) says 26 h.
2. Test: the existing idempotency tests adjusted, plus one that a retry
   between 24 h and 26 h after the first append is still deduplicated
   (clock injected as the suite does elsewhere) and one at > 26 h expires.
3. `CHANGELOG.md` Unreleased entry "R8: …".

## Out of scope
Everything else in the story (R5 landed, R7/R4 follow).

## Checklist and handoff
Tick the checklist items; attach `TASK-260910-2rsajv_results.md` (file:line,
transcripts: `python -m pytest -q` with `CURATOR_CONFORMANCE_ROOT`,
`python -m mypy`), then `task-board handoff TASK-260910-2rsajv --role developer`.
