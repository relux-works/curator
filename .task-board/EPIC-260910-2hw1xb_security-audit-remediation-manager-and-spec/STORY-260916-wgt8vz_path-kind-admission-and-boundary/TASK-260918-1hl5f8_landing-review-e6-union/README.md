# TASK-260918-1hl5f8: landing-review-e6-union

## Description
Landing review of the E6 spec revision rebased onto the E3 landing: the accepted revision 2 of TASK-260916-3l60rn (base e8b53a0) applied onto curator-spec main 4a2fa3e required hand-composed union hunks in protocol/environments.md (the section 12 store-trust row sentence, now with the E6 path-directory clause before the E3 codex-seed rows; the section 13 conformance surfaces joining the E5, S5, E3 and E6 blocks) and tools/validate.py (both registrations). The delivery branch e6-path-kind-admission (PR #70, commit 1ca4b3d) must be reviewed before it lands.

## Scope
(define task scope)

## Acceptance Criteria
The union diff (attached) equals the accepted E6 candidate plus the intervening landing everywhere except the named hunks; each of those carries both sides without loss or contradiction; make validate and regenerate-check pass on the branch; verdict resource recorded
