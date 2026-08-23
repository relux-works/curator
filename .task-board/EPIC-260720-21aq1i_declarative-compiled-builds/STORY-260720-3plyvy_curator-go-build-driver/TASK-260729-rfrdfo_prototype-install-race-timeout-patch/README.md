# TASK-260729-rfrdfo: prototype-install-race-timeout-patch

## Description
Prototype the reviewer-sanctioned test-only install race-timeout optimization in a task-owned copy of the exact TASK-260720-jrrgw9 candidate. Produce an applicable patch and focused evidence without mutating the main candidate.

## Scope
Copy the exact current jrrgw9 worktree into private task storage; capture a path-sorted SHA-256 source baseline; modify only the 13-file required allowlist from TASK-260729-3dr6hw revision 3. Add t.Parallel only to the approved 88 tests and partition atomicity injection with a separate injectClasses selector while retaining full scenario.classes success coverage. Do not touch product Go, aba_test.go, atomicity/fixture_test.go, timeouts, skips, assertions outside the explicitly retired cross-class defense-in-depth sequence, conformance root, pins, accepted worktrees, or shared caches. Run gofmt and only focused package/test commands with -count=1; no go test ./..., no full package race, no host install, stage, commit, publish or pin.

## Acceptance Criteria
A task-owned source baseline and post-manifest prove only the 13 allowlisted test files changed; an exact patch applies cleanly to the current jrrgw9 candidate; focused install and atomicity commands compile and pass without timeout overrides; assertion-retention and deliberate cross-class defense-in-depth retirement are documented; main candidate remains byte-identical.
