# TASK-260918-2c7dgq: landing-review-e3-union

## Description
Landing review of the E3 spec revision rebased onto the R3/P2, E5 and S5 landings: the accepted revision 2 of TASK-260916-2rnkei (base 684c9f1) applied onto curator-spec main e8b53a0 required hand-composed union hunks in protocol/environments.md (the section 12 status rows: store-trust row + codex-seed rows; the section 13 conformance surfaces: E5 write-nofollow + S5 store-boundary + E3 codex-seed blocks), in tools/validate.py (both constant blocks, both registrations) and CHANGELOG.md. The delivery branch e3-codex-seed (PR #69, commit 4a2fa3e) must be reviewed before it lands.

## Scope
(define task scope)

## Acceptance Criteria
The union diff (attached) equals the accepted E3 candidate plus the intervening landings everywhere except the named hunks; each of those carries both sides without loss or contradiction; make validate and regenerate-check pass on the branch; verdict resource recorded
