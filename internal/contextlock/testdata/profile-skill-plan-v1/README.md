# Profile skill-plan carrier corpus

Unmodified fixtures from relux-works/curator-spec at
65574fc58b8a2d2aca3ac0335eac9feb7c967696, conformance/profile-skill-plan-v1.
These are candidate allocation fixtures, not a released conformance pin.

`go test ./internal/contextlock` exercises all 42 structural cases through the
Go lock/plan readers and the literal baseline CCJ digests. Semantic cases are
selected by the Go tests for the PR1 graph/binding boundary only; transaction,
live routing, companion qualification and actual PATH delivery belong to PR2/3.
The complete source corpus is kept intact so subset selection is reviewable.
