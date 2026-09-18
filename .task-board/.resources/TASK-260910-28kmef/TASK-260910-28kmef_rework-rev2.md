# Rework brief — TASK-260910-28kmef, revision 2 (R5)

Revision 1 passed every item except one (`TASK-260910-28kmef_review-verdict-rev1.md`):

**R1 — canonicalization depth errors must be `ProtocolError`.**
`signing.py` defines `CanonicalDepthError(CanonicalError)` with
`CanonicalError` inheriting only `ValueError`; `canonical_bytes` and
`canonical_document_bytes` propagate it, so a caller catching
`ProtocolError` (the explicitly required contract) misses it. Make depth /
recursion errors from BOTH canonicalization entry points catchable as
`ProtocolError` while preserving `CanonicalError` / `ValueError`
compatibility and every other mapping (e.g. `CanonicalDepthError(CanonicalError, ProtocolError)`
or a shared error-definition module that breaks the protocol→signing import
direction — your call, state it). Add direct tests that catch
`ProtocolError` for (a) an actual over-depth input and (b) an injected
`RecursionError` from the CCJ validation path, for both entry points.

**Validation note (not a defect of yours).** The reviewer ran the suite
against the `dced9b8` root the rules named and hit the pre-existing
`checkpoint_cases` expectation; the CI pin is `47c3c8c` and the rules are
corrected: run the suite against a detached curator-spec worktree at
`47c3c8cbd5da5e3fd6d99b0327d382c6a36494fe` (see the updated
`remediation-registry-producer-rules.md`) and quote the command; do not
touch the pin.

Everything else stays byte-identical to revision 1. Update the results with
a "Revision 2" section, tick the checklist, and
`task-board handoff TASK-260910-28kmef --role developer`.
