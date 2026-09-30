# TASK-260930-38fjt0 integration preconditions (integration run)

Board status read: `integrating` (unchanged; no status writes made).
Worktree: exactly 3 uncommitted modifications, no commit on the branch:
- `.github/workflows/ci.yml` (+5/-1)
- `.github/ci/gate-selftest.sh` (+39)
- `docs/self-hosted-runner-setup.md` (+4/-1)
No CHANGELOG/LOGBOOK edit.

## 1. runs-on pin (ruby YAML parse, exit 0)
`test-self-hosted` runs-on == `["self-hosted", "macOS", "ARM64", "rose-air"]`.
Only self-hosted `runs-on` in ci.yml (line 241); all others are matrix/os or ubuntu-latest.
Job comment block states the lane is pinned to the rose-air runner by label.
Docs describe the required `rose-air` label and that an unlabeled runner never takes the job.

## 2. gate-selftest green (exit 0)
`bash .github/ci/gate-selftest.sh`: GATE_EXIT=0, `gate-selftest: 294 passed, 0 failed`.
Rose-air rows in log:
- `ok test-self-hosted runs-on is pinned to the rose-air label`
- `ok the self-test rejects an unpinned self-hosted runs-on`

## 3. label-removed mutant killed (independent probe, real exits)
Pinned workflow -> `runs_on_pins_rose_air` exit 0 (admitted).
Unpinned mutant (`rose-air` removed, line 241) -> exit 1 (rejected).

## 4. Landing readiness
Change no file in this run; no `worktree integrate` executed (runner lands).
Working tree left uncommitted for handoff snapshot.