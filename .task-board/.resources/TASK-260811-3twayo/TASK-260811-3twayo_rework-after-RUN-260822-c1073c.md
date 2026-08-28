# Rework required after reviewer RUN-260822-c1073c

Close the one remaining Python P10 protocol finding from `TASK-260811-3twayo_reviewer-verdict_RUN-260822-c1073c.md`; preserve all accepted code and tests.

Encode P10 as two target-scoped canonical outcomes, one per interpreter/platform/ABI binding, with distinct outcome identities over one selection-neutral capture. Add a separate cross-target reuse-negative that references those exact binding identities.

Both Go and Python must independently decode, validate, canonicalize, and hash the accepted shared wire shapes for capture graph, selection context, selection binding, active graph, and diagnostics. Do not use adapter-local abbreviated `python-protocol-*` substitutes for shared graph records. Preserve strict missing/unknown-field rejection for nested `lock`, `artifact`, and `build` objects.

Retain the accepted exact C4-to-C5 plan re-derivation and forged zero/subset/action/order plan negatives, plus every previously closed security and conformance path.

Run focused protocol tests, Python oracle, accepted Ruby verifier, race, vet, lint, repository compile/build, diff and board validation, and the full uncached repository suite unless a deterministic acceptance failure makes it irrelevant. Attach task-scoped evidence and use the developer handoff.
