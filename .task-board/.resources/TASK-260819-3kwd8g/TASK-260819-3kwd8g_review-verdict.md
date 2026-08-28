# Reviewer verdict: changes requested

Reviewed commit: `704060526560a36e540bb27678e58edb381482da`.

## Blocking finding

The rc.7 conformance layer does not enforce several normative verified-mode relationships, so the published gates can pass while accepting records that violate explicit MUST rules. `protocol/assurance.md` requires `observed_at` to precede `expires_at` and requires the first `permit-issued` checkpoint to use a null predecessor. However, `provider-capability-receipt-v1.schema.json` validates the two timestamps only independently, `execution-checkpoint-v1.schema.json` permits either a digest or null for every phase, and `tools/validate.py::validate_wire_semantics` has no rc.7 assurance cases. `validate_assurance_vectors` checks only list shape, two cache digests, failure flags, identity uniqueness, and the empty claim list; it does not validate record freshness, checkpoint sequencing, or cross-record binding.

Adversarial review against the committed schemas and validator produced: `provider-capability-receipt-v1.schema.json schema_errors=0 semantic_error=None` after setting `observed_at=2026-07-13T00:06:00Z` and `expires_at=2026-07-13T00:05:00Z`; and `execution-checkpoint-v1.schema.json schema_errors=0 semantic_error=None` after giving the first `permit-issued` checkpoint a non-null predecessor. An execution receipt with `started_at` after `completed_at` is likewise structurally and semantically accepted.

This leaves the acceptance criteria for negative conformance vectors, validators, and release-gate coverage unsatisfied. It also weakens independent proof of fail-closed provider negotiation and cache/checkpoint separation even though the normative prose is correct.

## Required rework

1. Add semantic validation plus generated invalid vectors for capability receipt interval ordering and phase-specific checkpoint predecessor rules.
2. Add executable relational vectors and validator checks for the normative provider/capability/permit/receipt bindings: provider identity, nonce, operation, capability receipt, permit, build input, artifact, expiry/freshness, checkpoint chain, and no portable fallback. Each normative mismatch should have a named rejection outcome and fail before execution where required.
3. Add mutation tests proving the validator and release gate reject those violations, and keep generator/regeneration coverage deterministic.
4. Rerun `make validate`, `make regenerate-check`, and `make release-check VERSION=1.0.0-rc.7` in a clean checkout.

## Checks completed

`make validate` passed: 49 schemas, 464 vector files, 80 Python tests, and Go tests. `make regenerate-check` passed. The release gate passed in a clean clone at the reviewed commit. The direct source worktree release-check refusal was caused only by the pre-existing untracked `task-board.config.json`. Historical release metadata hashes remain frozen: rc.5 `75ae17fc029b4f51ca40ce768d04fd72991ec3db2602b8fe59213bee6ac34583`; rc.6 `c4ad58e76687bd563679773a60c6ce35c238d4117b7cbceb05d4f88b5300ed3f`. The normative mode selection, platform-neutral provider contract, separate host installation, compiled-artifact denial, and no released verified claim otherwise fit the requested architecture.