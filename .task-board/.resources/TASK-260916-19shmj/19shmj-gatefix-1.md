# TASK-260916-19shmj — gate fix (THE ONLY CURRENT INSTRUCTION, with 19shmj-sec-brief.md)

Revision 1 (tree ce268158, base 55b94af2) FAILED the hosted gate on EVERY lane (Test ubuntu/macos/windows, Race ubuntu/macos): run 36296393245,
`TASK-260916-19shmj_change-request_rev1-validation.log`. Visible failures are in internal/envprofile: TestResolveDriftRepair,
TestResolvePassthroughLiveness, TestResolveClaudeProjectEntry, TestSharedToIsolatedRemovesStaleLink, TestCredentialLinkRegularFileRefuses,
TestStaleCredentialLinkRefusals, TestDanglingPiLinkReportedDetached, TestCredentialLinkDirectory, TestCredentialLinkUnrecordedSymlinkRefuses —
i.e. your nofollow write rule now breaks the manager's OWN recorded credential/passthrough links and drift repair. Your local broad
`go test ./internal/envprofile` never produced an exit code, so the regression went unseen. Read `gh run view 36296393245 --log-failed`
(from the worktree) for every lane. Fix: the nofollow rule must refuse following a PLANTED link at a managed write target, while the
manager's recorded links (0017 credential modes, passthrough, drift repair/takeover) keep their specified behaviour — cite the rc.13 clauses
that distinguish them. Run `go test ./internal/envprofile` split by -run groups, each with a real captured exit code.
Also: trunk moved to 38c68570 — `git fetch origin main`, combine keeping both sides; VERIFY `git diff --name-only origin/main -- . ':!.task-board'`
shows only your paths. `task-board m 'set_status(TASK-260916-19shmj, status=development)'` first; append "Revision 2 — recorded links keep
their behaviour", `resource update`, handoff; the hosted gate on the published revision is the only green proof. Stay in the turn. If the
loop detector refuses, stop and report. No CHANGELOG/LOGBOOK edit.
