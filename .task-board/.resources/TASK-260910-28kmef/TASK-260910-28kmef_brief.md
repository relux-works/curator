# Brief — TASK-260910-28kmef: pathological JSON returns `400 invalid_json`, never `500` (R5, service)

Story `STORY-260910-2xe3n2` (registry-robustness-hardening, `EPIC-260910-16qce1`),
wave 4 of the 2026-09 security-audit remediation; first of four sequential
leaves (then `TASK-260910-2rsajv`, `TASK-260910-3u9t1e`, `TASK-260910-2c7s0u`).
Rules: `remediation-registry-producer-rules.md` (attached; `main` is now
`c7ef32c`, the R3/P2 landing, and the CI protocol-suite pin is already
`dced9b8` — do not move it). Role: developer.
Worktree: the managed Story worktree `<control-root>/.temp/STORY-260910-2xe3n2/worktree`.

## Finding (read it first)
`docs/security-audit-2026-09.md` R5 (Low): `protocol.load_json` (and the
canonicalization / `_validate_ccj` path in `signing.py`) recurse; a deeply
nested document raises `RecursionError`, which the generic handler reports as
`500 internal_error` instead of `400 invalid_json`.

## Deliverable
1. `load_json` and the canonical-bytes / CCJ validation entry points convert
   `RecursionError` into `ProtocolError` (message naming the nesting bound),
   so every request path that parses or canonicalizes client JSON answers
   `400 invalid_json` with the existing error envelope; no other exception
   mapping changes, and internal (non-request) callers keep raising
   `ProtocolError` as they do for other malformed input. Prefer an explicit
   depth bound checked during parsing over relying on the interpreter's
   recursion limit if the code structure allows it cheaply; state the bound
   in the docstring and `README.md` limits section (if one exists).
2. Tests: an HTTP test posting a document nested deeper than the bound to
   `submit` (and to every other endpoint that parses a body, if any)
   asserts `400` + `invalid_json`, never `500`; a unit test for
   `load_json` and for the canonicalization path; the store's own
   JSON never trips the bound (existing suite green).
3. `CHANGELOG.md` Unreleased entry "R5: …".

## Out of scope
R4/R7/R8 (the next leaves), rate limiting, request size limits, spec edits.

## Checklist and handoff
Tick the checklist items you satisfy; attach `TASK-260910-28kmef_results.md`
(per-AC file:line, transcripts: `python -m pytest -q` with
`CURATOR_CONFORMANCE_ROOT`, `python -m mypy`), then
`task-board handoff TASK-260910-28kmef --role developer`; the runtime runs
the hosted gate once at handoff.
