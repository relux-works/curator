# STORY-260823-31jj6m: candidate-lane-qualification-fixes

## Description
Two concrete curator-side blockers keep the schema-8 rc.9 candidate-conformance matrix red (candidate branch candidate/schema-8-rc.9, evidence on TASK-260822-c0rxj7): the Ubuntu lane needs the multi-project dry-run binding fix that exists on open curator PR 14 but not on main, and the Windows lane fails candidate identity because Git for Windows shasum prefixes the digest with a backslash when the hashed path is escaped. Fix both on curator main, then re-run candidate-conformance with the same SHA and digest.

## Scope
(define story scope)

## Acceptance Criteria
Candidate-conformance run green on ubuntu and windows for the exact rc.9 candidate SHA and digest; evidence routed to the implementation stories and the blocked vector tasks.
