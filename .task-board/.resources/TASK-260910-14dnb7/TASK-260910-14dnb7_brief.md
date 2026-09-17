# Brief — TASK-260910-14dnb7: committed boundary in /v1/records and /v1/log envelopes (R1 service half)

Story `STORY-260910-3rvvxh` (records-boundary-in-response, `EPIC-260910-16qce1`),
wave 2 of the 2026-09 security-audit remediation; first of two leaves (the
sibling `TASK-260910-27yepb` binds the cursor to the boundary and follows on
the same Story branch). Rules: `remediation-registry-producer-rules.md`
(attached). Role: developer.

## Finding (read it first)
`docs/security-audit-2026-09.md` "R1 (service half)": `records_page`/`log_page`
evaluate at a committed boundary and cursors bind it, but the envelope never
states which boundary was served, so a key-holding registry can serve an
advancing `/v1/snapshot` while answering `/v1/records` at an older boundary.
The spec revision landed as curator-spec `dced9b8` (`protocol/registry.md`
§9.3 "Page boundary", §5, §9 endpoint table; `profiles/registry-service.md`
§2/§5/§11; schemas `records-response-v2`, `log-response-v2`; vectors
`registry-service.json` → `pagination.boundary_emitted_on_every_page`,
`pagination.chain_boundary_byte_identical`, `pagination.cursor_boundary_cases`).

## Deliverable
1. Every successful `/v1/records` GET and `/v1/log` response carries the
   REQUIRED `boundary` member: the registry snapshot object
   (`registry-snapshot-v1`: `schema_version`, `merkle_root`, `log_size`,
   `head`, `version`, `created_at`, `sig`) at which the page was evaluated —
   the same object `/v1/snapshot` would return for that committed boundary,
   signed with the service key. All pages of one cursor chain carry a
   byte-identical `boundary` (CCJ-1 bytes equal): the cursor already carries
   the boundary (`_cursor_state`), so cursor pages re-derive the identical
   signed object (including the immutable `created_at` of that boundary per
   registry §5 — never a refreshed timestamp).
2. Envelopes validate against `records-response-v2` / `log-response-v2`
   (`additionalProperties: false`): exactly `records|entries`, `next_cursor`,
   `boundary`. Keep `_encode_cursor`/`_cursor_state` semantics (the sibling
   task adds the disagreement refusal; do not pre-empt it beyond what falls
   out naturally, and say what you did).
3. Tests (`tests/test_registry.py` + `tests/test_protocol_conformance.py`):
   boundary present on first and cursor pages for both endpoints with
   byte-identical CCJ-1 bytes across the chain; the boundary verifies against
   the service key and equals `/v1/snapshot` for the same boundary; a page
   served after an append still carries the chain's original boundary (the
   existing `append_after_first_page` vector, now asserting the boundary);
   served envelopes validate against the v2 schema-cases' schema (load the
   schema from `CURATOR_CONFORMANCE_ROOT/../../schemas/v1/` or the case's
   registration — follow how the existing tests locate spec files) and the
   new `registry-service.json` `pagination` flags are driven.
4. `.github/workflows/ci.yml`: protocol-suite `ref` → `dced9b8317e0e8af79edf2d0539b32bd22b6c85b`
   (settled); every pre-existing conformance test must stay green at the new
   pin — if a pre-existing test needs a change because the vectors moved
   between the old pin and `dced9b8`, make the minimal change and explain it.
5. `CHANGELOG.md` Unreleased entry "R1: …" naming the `boundary` member, the
   v2 envelope schemas and the spec revision; `README.md` mention where the
   API is described.

## Out of scope
Cursor/boundary disagreement refusal and its vector (`TASK-260910-27yepb`),
R2 caching, the manager client, the spec.

## Checklist and handoff
Tick the checklist items you satisfy; attach `TASK-260910-14dnb7_results.md`
(per-AC file:line, transcripts: `python -m pytest -q` with
`CURATOR_CONFORMANCE_ROOT`, `python -m mypy`), then
`task-board handoff TASK-260910-14dnb7 --role developer`; the runtime runs the
hosted gate once at handoff.
