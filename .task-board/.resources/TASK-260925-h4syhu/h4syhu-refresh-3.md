# TASK-260925-h4syhu — refresh 3 onto trunk 3bdcfe07 (THE ONLY CURRENT INSTRUCTION)

Revision 3 was ACCEPTED. Trunk moved from 9f0da708 to `3bdcfe07`: 1f2ng0 (platform-cases.tsv, skip-classes.tsv, cmd/curator/profile.go,
internal/envprofile), 4hd81z (scriptpolicy test), and cww1ov — 0017 credential modes, which ALSO introduced/extended the stateread seam and
migrated readers (internal/stateread, envprofile, envregistry, runtimestore, install/draftsources.go, cmd/curator/main.go). This Story (2as5sx
guard + h4syhu rows) overlaps heavily. Safety ref: refs/campaign/h4syhu-rev3-delta-20260926.
1. `task-board m 'set_status(TASK-260925-h4syhu, status=development)'` if needed.
2. Combine: `git diff 9f0da708 3bdcfe07 -- . ':!.task-board' ':!CHANGELOG.md' ':!LOGBOOK.md' | git apply --3way`. Resolve every overlap so BOTH
   survive: ONE stateread seam (reconcile cww1ov's and this Story's APIs; no duplicate helpers), every reader migrated once, the 2as5sx
   deny-by-default guard scanning the merged code with honest counts/ratio (new collapse sites from 1f2ng0/cww1ov migrated or allow-listed
   with reason). List every conflict and its resolution. Leave nothing staged.
3. Just before step 4: `git fetch origin main`; if origin/main moved again (306v4m is landing: .github/ci/gate-selftest.sh,
   install-rust-toolchain.sh, docs/self-hosted-runner-setup.md — disjoint), combine that diff too.
4. VERIFY `git diff --name-only origin/main -- . ':!.task-board'` over the working tree lists only this Story's paths; then
   `task-board worktree refresh-candidate TASK-260925-h4syhu` (187z6x replay via `--replay-resolutions` only).
5. Bounded runs: the guard test, M1 kill, stateread tests, `go test ./internal/envprofile ./internal/install -run 'LockedNetworkRepository|StateRead|Guard' -count=1`
   in parts, go vet, `GOOS=windows go vet ./internal/install ./internal/envprofile`. Real exit codes.
6. Append "Revision 4 — refresh onto trunk incl. cww1ov", `resource update`, `task-board handoff TASK-260925-h4syhu --role developer`;
   stay in the turn while the gate runs. A write-boundary `policy warn` block is a warning.
