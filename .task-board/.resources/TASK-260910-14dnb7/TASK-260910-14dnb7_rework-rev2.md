# Rework brief — TASK-260910-14dnb7, revision 2 (R1 service half)

Revision 1 was rejected with two corrections
(`TASK-260910-14dnb7_review-verdict-rev1.md`). Everything else passed review;
keep it unchanged unless a correction touches it.

## Corrections (both required)
1. **Chain boundary survives staged key rotation (High).** Pages currently
   sign the boundary with the active key on every request; a cursor chain
   that spans a staged rotation therefore serves the same snapshot body with a
   different `sig`, violating registry §9.3 / profile §2 (byte-identical
   complete signed boundary across a chain). Settled fix: the cursor carries
   the complete signed boundary object it was issued with (the `snapshot`
   member of the cursor payload becomes the full `registry-snapshot-v1`
   including `sig`; cursor size stays far below the 4096-char bound — state
   the measured size), and every cursor page echoes exactly that object.
   `/v1/snapshot` keeps using the active signer. A cursor whose carried
   boundary no longer verifies against the accepted key set (retired key) is
   invalid (`404 invalid_cursor`) — that is the existing overlap rule, now
   applied to the carried object; do not invalidate cursors still inside the
   promised overlap. Tests: both endpoints, a chain started before
   `activate-key-rotation` and continued during the overlap: byte-identical
   `boundary` on every page, verified against the original key; after the
   old key retires the cursor is refused, never re-signed. Coordinate the
   cursor representation with the sibling `TASK-260910-27yepb` by documenting
   it in the results (the sibling adds the disagreement refusal on top).
2. **Validate served envelopes with the real schemas (Medium).** Replace the
   partial hand-written validator in `tests/test_protocol_conformance.py` with
   Draft 2020-12 validation of the served envelopes and the registered
   schema-cases against the actual `records-response-v2` / `log-response-v2`
   schemas with local `$ref` resolution (`registry-snapshot-v1`,
   `registry-log-entry-v1`, `audit-record-v1` …) — add `jsonschema` to the
   `dev` extra in `pyproject.toml` (pin a floor, e.g. `jsonschema>=4.23`) and
   to the CI install if needed. Include the negative regression the reviewer
   demonstrated: a served log envelope whose `entry_hash` is `"invalid"` MUST
   be rejected by the harness.

## Validation and handoff
`python -m pytest -q` with `CURATOR_CONFORMANCE_ROOT` set and `python -m mypy`
(state interpreter and exit codes); update `TASK-260910-14dnb7_results.md`
with a "Revision 2" section (per-correction file:line, transcripts, measured
cursor length, the sibling coordination note) and hand off with
`task-board handoff TASK-260910-14dnb7 --role developer`. Worktree and rules
unchanged (`TASK-260910-14dnb7_brief.md`, `remediation-registry-producer-rules.md`).
