# TASK-260917-e6nwdj: landing-review-s5-union-with-e5

## Description
Landing review of the S5 spec revision rebased onto the E5 and R3/P2 landings: the accepted revision 2 of TASK-260910-39fzpq (base 684c9f1) applied onto curator-spec main 9912db7 required four hand-composed hunks in protocol/environments.md where S5 and E5 meet (link-target currency in section 10.1, the repair paragraph, the non-current list, the section 13 conformance surfaces). The delivery branch s5-store-boundary (PR #68, commit e8b53a0) must be reviewed before it lands.

## Scope
(define task scope)

## Acceptance Criteria
The union diff (attached) equals the accepted S5 candidate plus the intervening landings everywhere except the four named hunks; each of those hunks carries both sides (S5 semantics, E5 section 8.3.1 references) without loss or contradiction; make validate and regenerate-check pass on the branch; verdict resource recorded
