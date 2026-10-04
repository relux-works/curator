# BUG-261004-16a407: trust-pin-overrides-strict-findings

## Description
N5 (High, static; reported by the cocoaskills comparison 2026-10-04, confirmed by reading). internal/audit/audit.go decideWithPins returns allow for any pinned content before Decide runs, so in strict mode a pinned tree with a verifiable finding at or above fail_on is allowed. Spec manager §7 (rc.14): strict mode requires an operator pin for pre-capability schemas, and a verifiable finding at or above fail_on blocks in strict mode, with no pin exception; revocation still blocks first (correct). A clarifying spec sentence (pins waive the missing-capability requirement only) belongs with the fix.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
1. Red-first regression through the production entry point: strict mode, pinned schema-3+ tree with a verifiable finding at or above fail_on -> block; pinned pre-capability tree without such findings -> allow; unpinned pre-capability tree -> require_pin; revocation still blocks pinned content.
2. Cached-verdict path (loadCachedFindings) and fresh path behave identically.
3. Advisory mode unchanged (warn).
4. curator-spec manager §7 gets one clarifying sentence on pins vs strict findings (spec-first or lockstep).
