# TASK-260823-2erkhe: fix-windows-candidate-identity-digest

## Description
curator repo: the Windows candidate-conformance lane fails candidate identity before the suite gate — Git for Windows shasum emits a leading backslash before the digest when the hashed path contains Windows escaping, so the measured digest is textual backslash-782d... instead of canonical 782d... Fix the digest computation or parsing in the candidate scripts (.github/ci/candidate-suite.sh and wherever shasum output is consumed) to normalize or avoid the escape (e.g. hash via stdin), add a negative case to gate-selftest.sh so a prefixed digest is caught, and land on curator main via PR with green CI. Worktree from origin/main, not the diverged local main. Maintainer pre-authorization 2026-08-22: fully autonomous.

## Scope
(define task scope)

## Acceptance Criteria
Digest computation immune to Git for Windows shasum escaping; gate-selftest covers the case; merged to main with green CI.
