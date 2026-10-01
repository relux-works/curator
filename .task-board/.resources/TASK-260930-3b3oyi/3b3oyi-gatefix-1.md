# TASK-260930-3b3oyi — Windows gate fix 1 (THE ONLY CURRENT INSTRUCTION, with secondfix-brief.md)

Rev1 (tree 5d5a332a) failed only on Test (windows-latest), in hosted run 36778472331. No individual test reported FAIL in go-test.json.
That points to a package-level failure: a build error, a panic, a timeout or a TestMain exit. Download the job log and artifacts from
the worktree:
`gh run view 36778472331 --log-failed`, then `gh run download 36778472331 -p 'test-evidence-windows*'`.
Find the package-level failure and fix it. Likely suspects are the posture-warning suppression env marker and the help-without-config
path on Windows.
- Do not skip on Windows.
- Do not weaken assertions.

Run `GOOS=windows go vet ./cmd/curator ./internal/...` and the affected tests locally, with real exit codes. Set status development,
update the results, run `task-board handoff TASK-260930-3b3oyi --role developer`, then END YOUR TURN. No CHANGELOG/LOGBOOK edit.
