# Brief — TASK-260910-27yepb: cursor bound to the page boundary (R1/P1 service half, second leaf)

Story `STORY-260910-3rvvxh` (records-boundary-in-response, `EPIC-260910-16qce1`),
wave 2 of the 2026-09 security-audit remediation; second and FINAL leaf of the
story. The first leaf `TASK-260910-14dnb7` (accepted, checkpointed on the
Story branch) made every `/v1/records` and `/v1/log` envelope carry the signed
`boundary`, made the cursor carry the complete signed boundary it was issued
with (every cursor page echoes exactly that object, rotation-stable), and
validates served envelopes against the real v2 schemas. Read
`TASK-260910-14dnb7_results.md` (rev 2 section, incl. the sibling
coordination note on the cursor payload shape) before coding; build on it,
do not redo it. Rules: `remediation-registry-producer-rules.md` (attached).
Role: developer. Worktree: the managed Story worktree
`<control-root>/.temp/STORY-260910-3rvvxh/worktree` (already carrying the
checkpointed 14dnb7 commit).

## Spec (the contract)
curator-spec `dced9b8`: `profiles/registry-service.md` §2 (a cursor is bound
to the boundary of its first page; a service MUST refuse — `404
invalid_cursor` — to serve a cursor page at a boundary that differs from the
cursor's and MUST NOT re-evaluate a cursor at a newer boundary), §5, §11;
`protocol/registry.md` §9/§9.3; vectors `registry-service.json` →
`pagination.cursor_boundary_cases` (`cursor-boundary-disagreement`: 404
`invalid_cursor`, no re-evaluation) and the existing `cursor_rejections`.

## Deliverable
1. The disagreement refusal as a real gate, not an assumption: when a cursor
   page is served, the boundary the page is evaluated at MUST be the cursor's
   carried boundary — assert it structurally (the store call takes the
   cursor's boundary; a mismatch between the carried boundary and what the
   store would evaluate at — e.g. a carried boundary that is no longer
   available, or a boundary body that disagrees with the carried signed
   object's fields — is `404 invalid_cursor`); a cursor is never re-evaluated
   at a newer boundary (no fallback to `store.snapshot_boundary()` when the
   carried one is unavailable). Verify the carried boundary's own fields
   against the store (head/merkle_root/log_size at that log_size) before
   serving, so a signed-but-inconsistent carried object cannot pass.
2. Conformance: drive `pagination.cursor_boundary_cases` through the real
   endpoints in `tests/test_protocol_conformance.py` (the disagreement case
   → 404 `invalid_cursor`; the chain is not re-evaluated after an append —
   reuse the `append_after_first_page` flow); keep every existing
   `cursor_rejections` case green.
3. Tests in `tests/test_registry.py`: a forged/altered carried boundary
   (body field changed while signature re-made with a key the service
   accepts — e.g. the overlap key — but disagreeing with the store) is
   refused; a carried boundary for a pruned/unavailable log prefix is
   refused; both endpoints.
4. `CHANGELOG.md`: extend the R1 Unreleased entry with the P1 cursor binding;
   `README.md`/`SECURITY.md` sentence where pagination is described.

## Out of scope
R2 (boundary memoization), the manager client, the spec, key-rotation
policy changes.

## Checklist and handoff
Tick the checklist items you satisfy; attach `TASK-260910-27yepb_results.md`
(per-AC file:line, transcripts: `python -m pytest -q` with
`CURATOR_CONFORMANCE_ROOT`, `python -m mypy`), then
`task-board handoff TASK-260910-27yepb --role developer`; the runtime runs the
hosted gate once at handoff.
