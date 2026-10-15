# TASK-261015-258cwf: env-config-show-permissions-lock

## Description
curator env config show reports the effective permissions posture {mode, locked, source} per profile, from the same code as the fragment permissions member (option a, owner-approved).

## Scope
(define task scope)

## Acceptance Criteria
env config show (human and JSON) reports {mode, locked, source} per installed profile with fragment semantics from the shared code path; unreadable posture is reported, never shown as unlocked; production-entry tests incl. the read-failure mutant; docs and CHANGELOG updated; tests and lint pass
