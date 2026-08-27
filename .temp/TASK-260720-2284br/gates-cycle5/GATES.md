# TASK-260720-2284br rework cycle 4 (R5) gates

Worktree: /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1zntv0/worktree
Driver start: 2026-07-28T19:50:05
Driver done:  2026-07-28T20:08:03
golangci-lint: golangci-lint has version 2.4.0 built with go1.25.5 from (unknown, modified: ?, mod sum: "h1:qz6O6vr7kVzXJqyvHjHSz5fA3D+PM8v96QU5gxZCNWM=") on (unknown)

Every exit below is the real status of a standalone process, written to a
durable .exit file as that gate's last action. A gate with no .exit file was
killed or never finished and is not evidence.

| gate | exit | seconds |
| --- | --- | --- |
| gate-build | 0 | 0 |
| gate-diff-check | 0 | 0 |
| gate-gofmt | 0 | 0 |
| gate-lint-repo-clean | 0 |  |
| gate-lint-repo | 0 | 2 |
| gate-race-activation | 0 | 98 |
| gate-race-concurrency | 0 | 59 |
| gate-race-loaders | 0 | 3 |
| gate-race-r5 | 0 | 75 |
| gate-race-revalidation | 0 | 108 |
| gate-reviewer-overlays | 0 | 10 |
| gate-test-atomicity | 0 | 406 |
| gate-test-godriver-rerun1 | 0 |  |
| gate-test-godriver | 1 | 51 |
| gate-test-install | 0 | 190 |
| gate-test-loaders | 0 | 1 |
| gate-test-r5 | 0 | 15 |
| gate-test-rest | 0 | 29 |
| gate-test-revalidation | 0 | 30 |
| gate-vet | 0 | 1 |

## Red gate, not claimed green

gate-test-godriver exited 1 in the driver run:
one subtest, TestBuildStopsBeforeBuildForEveryPreflightRejectionClass/escaped_embed,
hit the sibling-owned 15s wall-clock probe deadline (go-v1 process_timeout) with no
assertion failure anywhere, on a machine at load average 12+ from unrelated work.
Isolated rerun of the same tree: exit 0 in
28.034s.
internal/godriver is owned by TASK-260720-6i3cya and is untouched by this task.

## Negative control (pre-fix relation restored)

| variant | exit | outcome |
| --- | --- | --- |
| digest before the read | 1 | 4 install-level ABA cases FAIL committing the transient declaration; in-place binding case FAILS; byte-identical control PASSES |
| digest after the read | 1 | rename binding case FAILS; in-place case passes |
