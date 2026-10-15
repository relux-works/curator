# BUG-261015-2dqj4r: legacy-movedtag-source-change-negative-row

## Description
Recovered 2026-10-15 from the parked predecessor lane's unpublished board state (untrusted origin, text last updated before 2026-10-10); re-checked against curator main ea072eae. The sameLegacyPackageSource guard in detectMovedTagsIn (internal/install/install.go) has no negative row: when a skill's repository or directory changes while the same tag name is declared and the commits differ, the correct behaviour is no moved-tag diagnostic, and no test fails if the guard is removed. Confirmed: no test references the guard. Prior id BUG-261009-3ihpfr.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
(define fix acceptance criteria)
