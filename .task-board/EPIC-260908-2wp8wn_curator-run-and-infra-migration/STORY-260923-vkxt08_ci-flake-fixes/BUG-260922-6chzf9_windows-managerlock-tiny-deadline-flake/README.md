# BUG-260922-6chzf9: windows-managerlock-tiny-deadline-flake

## Description
Hosted windows-latest flake: internal/managerlock TestSubprocessExpectedAcquiredWithTinyDeadlineReportsBlocked fails with uncontended helper with tiny deadline = acquired, want blocked (gate runs 35702368558 for TASK-260922-1t2w1q rev2 and 35709048693 for TASK-260916-1xib1x rev3, both candidates untouched managerlock). The row assumes a tiny deadline always elapses before an uncontended acquire completes; on a fast Windows runner the acquire wins. Make the row deterministic (contend the lock or inject a clock) without weakening the blocked-report semantics.

## Scope
internal/managerlock test determinism (and a minimal production fix only if the analysis proves one); CHANGELOG. No other package.

## Acceptance Criteria
1) the guarded contract is stated with code citations; 2) the row is deterministic (-count=200 green) without weakening the contract, no retry loop, no either-outcome assertion; 3) a narrowing mutant of the contract is killed by the rewritten row; 4) sibling managerlock rows stay green; 5) CHANGELOG Fixed entry
