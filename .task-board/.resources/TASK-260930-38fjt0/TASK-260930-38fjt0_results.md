# TASK-260930-38fjt0 results
Changes (uncommitted in story worktree):
- .github/workflows/ci.yml: test-self-hosted `runs-on: [self-hosted, macOS, ARM64, rose-air]`; comment block states the lane is pinned to the rose-air runner by label.
- .github/ci/gate-selftest.sh: rows `test-self-hosted runs-on is pinned to the rose-air label` (parses the job's runs-on list, requires exact element `rose-air`) and in-script mutant row `the self-test rejects an unpinned self-hosted runs-on` (label stripped copy must be rejected).
- docs/self-hosted-runner-setup.md: labels now `self-hosted, macOS, ARM64, rose-air` + custom-label requirement.
Evidence:
- `bash .github/ci/gate-selftest.sh` → exit 0, 294 passed, 0 failed.
- Mutant: removed `rose-air` from ci.yml runs-on → exit 1, 292 passed, 2 failed (pin row FAIL "observed runs-on: [self-hosted, macOS, ARM64]"; mutant row FAIL since no label to strip). ci.yml restored, verified.
Bound: row checks the single-line flow-sequence runs-on form only; a block-sequence rewrite fails the row (fail-closed).
No CHANGELOG/LOGBOOK edits.
