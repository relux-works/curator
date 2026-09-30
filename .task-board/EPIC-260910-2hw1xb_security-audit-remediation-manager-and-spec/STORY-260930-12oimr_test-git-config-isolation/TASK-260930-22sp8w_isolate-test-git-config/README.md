# TASK-260930-22sp8w: isolate-test-git-config

## Description
Every test process that runs git uses an isolated git config (GIT_CONFIG_GLOBAL pointing to an empty test-owned file, GIT_CONFIG_NOSYSTEM=1), both via the CI test gate and in the Go test binaries themselves.

## Scope
(define task scope)

## Acceptance Criteria
1. test-gate.sh exports GIT_CONFIG_NOSYSTEM=1 and an empty GIT_CONFIG_GLOBAL. 2. Shared TestMain helper in git-using packages. 3. Hostile-global-config regression row passes. 4. gate-selftest row. 5. Helper-removed mutant killed.
