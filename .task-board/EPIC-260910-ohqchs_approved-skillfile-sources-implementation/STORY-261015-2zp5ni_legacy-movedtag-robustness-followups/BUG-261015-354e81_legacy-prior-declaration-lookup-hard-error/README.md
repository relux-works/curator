# BUG-261015-354e81: legacy-prior-declaration-lookup-hard-error

## Description
Recovered 2026-10-15 from the parked predecessor lane's unpublished board state (untrusted origin, text last updated before 2026-10-10); re-checked against curator main ea072eae. A failure to look up the prior declared ref is a hard install error even without StrictTags (install.go: cannot establish prior declared ref for ...). Confirmed the error path is still present. Decide the intended behaviour and add the cases (a draft-lane marker without a generations record re-installed under a schema-1 Skillfile; a transitive dependency). Prior id BUG-261009-2oe6u4.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
(define fix acceptance criteria)
