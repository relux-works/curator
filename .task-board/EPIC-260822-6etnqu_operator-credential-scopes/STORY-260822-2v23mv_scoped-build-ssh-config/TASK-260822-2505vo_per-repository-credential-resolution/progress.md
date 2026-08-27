## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(5))

## Blocked By
- TASK-260822-96m5pj

## Blocks
- TASK-260822-b0wg3a

## Checklist
- [x] Per-repository resolution wired into external install; flags/env precedence over config scopes
- [x] https and local-substitution repositories skip selection; empty ssh selection fails closed with protocol code
- [x] Pinned-agent, agent-only, identity-only selections reach SSHPolicy; tilde expansion for config paths
- [x] go test green
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260822-c60921, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260822-c60921)
Work lands in a worktree off origin/main at .temp/TASK-260822-2505vo/worktree, branch task/TASK-260822-2505vo-per-repo-credentials. The checkpoint branch handoff/cocoaskills-parity-20260731 has no internal/buildrepo and no internal/install/external.go (50 commits behind main), so the named target files only exist on main. The accepted TASK-260822-96m5pj config patch was applied there verbatim first; it applies clean to main.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-c60921, pid=14968, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260822-3af710, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260822-3af710)
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-3af710, pid=80063, exit=0)
Finalize only: implementation is COMPLETE and verified in .temp/TASK-260822-2505vo/worktree (branch off origin/main 6a9b201, 14 files, index clean). Prior run ended mid-final-gate. Remaining: (1) run go test ./... -count=1 on the worktree; (2) regenerate the patch artifact via git diff against 6a9b201 and attach as resource; (3) attach results.md summarizing: per-repo resolution, precedence, https/local skip, fail-closed protocol code, three selection shapes, tilde expansion, and the CI skip-class fix for Windows lanes (.github/ci/skip-classes.tsv); (4) check remaining checklist items; (5) handoff to-review. Do not rewrite code unless the final test run fails.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260822-f380a3, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260822-f380a3)
agent completed: [implementer] developer (claude) (exit=1)
spawn run completed: claude (run=RUN-260822-f380a3, pid=68189, exit=1)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260822-25ce5c, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260822-25ce5c)
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-25ce5c, pid=40396, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [reviewer] reviewer (claude) (run=RUN-260822-13d94b, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260822-13d94b)
reviewer verdict RUN-260822-13d94b: ACCEPTED. Evidence in TASK-260822-2505vo_review.md. Every AC clause is implemented and covered by a named test; the board patch artifact byte-matches the worktree diff vs 6a9b201; the branch applies the accepted TASK-260822-96m5pj patch verbatim with only MatchBuildSSH added. Reviewer re-ran on the final tree: gofmt clean, go vet 0, golangci-lint v2.12.2 0 issues, targeted SSH tests ok in install/buildrepo/config, .github/ci/gate-selftest.sh 75 passed 0 failed, attached full-suite log 41 ok / 0 FAIL. No production regression: productionGitTool never sets SSHWrapper, so AcquireNetwork already refused external SSH before this change. Four non-blocking observations recorded for the docs/CLI follow-ups (known-hosts precedence asymmetry, no ~ expansion for the env spelling, install.Options.BuildSSH write-only inside the package, Lstat-based known_hosts default). Reviewer supplies no commit_ack: work is uncommitted on task/TASK-260822-2505vo-per-repo-credentials; the commit-owning mover lands TASK-260822-2505vo_final.patch and makes the final done transition.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-13d94b, pid=12944, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260822-2505vo_spawn-log_-implementer--developer--claude-_RUN-260822-c60921.log](file://TASK-260822-2505vo/TASK-260822-2505vo_spawn-log_-implementer--developer--claude-_RUN-260822-c60921.log) — System spawn log captured by task-board
- [TASK-260822-2505vo_spawn-log_-implementer--developer--claude-_RUN-260822-3af710.log](file://TASK-260822-2505vo/TASK-260822-2505vo_spawn-log_-implementer--developer--claude-_RUN-260822-3af710.log) — System spawn log captured by task-board
- [TASK-260822-2505vo_spawn-log_-implementer--developer--claude-_RUN-260822-f380a3.log](file://TASK-260822-2505vo/TASK-260822-2505vo_spawn-log_-implementer--developer--claude-_RUN-260822-f380a3.log) — System spawn log captured by task-board
- [TASK-260822-2505vo_spawn-log_-implementer--developer--claude-_RUN-260822-25ce5c.log](file://TASK-260822-2505vo/TASK-260822-2505vo_spawn-log_-implementer--developer--claude-_RUN-260822-25ce5c.log) — System spawn log captured by task-board
- [TASK-260822-2505vo_results.md](file://TASK-260822-2505vo/TASK-260822-2505vo_results.md)
- [TASK-260822-2505vo_full-suite-final.log](file://TASK-260822-2505vo/TASK-260822-2505vo_full-suite-final.log)
- [TASK-260822-2505vo_final.patch](file://TASK-260822-2505vo/TASK-260822-2505vo_final.patch)
- [TASK-260822-2505vo_spawn-log_-reviewer--reviewer--claude-_RUN-260822-13d94b.log](file://TASK-260822-2505vo/TASK-260822-2505vo_spawn-log_-reviewer--reviewer--claude-_RUN-260822-13d94b.log) — System spawn log captured by task-board
- [TASK-260822-2505vo_review.md](file://TASK-260822-2505vo/TASK-260822-2505vo_review.md) — Reviewer verdict: accepted; AC-by-AC evidence, independent gofmt/vet/lint/test/gate re-runs, and four non-blocking observations

## Created
2026-08-22T16:12:05Z

## Last Update
2026-08-22T19:10:55Z

## Assigned To
[reviewer] reviewer (claude)
