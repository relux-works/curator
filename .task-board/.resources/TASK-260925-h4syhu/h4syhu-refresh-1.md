# TASK-260925-h4syhu — refresh the Story onto trunk ab34556e and publish (THE ONLY CURRENT INSTRUCTION, with h4syhu-brief.md)

Your h4syhu work is done in the Story worktree, but publication refused `change_request_base_authority_mismatch`: the Story checkpoint
d9cb8465 (187z6x) sits on a48f584c; trunk is `ab34556e` (Skillfile default-on + lock replay 3eywt2, marker v5 txgta4, CI budget 2mla0q,
3qwrnl, and possibly 11jgkt). Safety ref of the current worktree: refs/campaign/h4syhu-delta-20260926.
1. `task-board m 'set_status(TASK-260925-h4syhu, status=development)'` if needed.
2. Combine trunk: `git diff a48f584c ab34556e -- . ':!.task-board' ':!CHANGELOG.md' | git apply --3way`. Keep BOTH sides. The 2as5sx
   deny-by-default AST guard (the §8.4 absence-vs-read-failure class scan) now also scans trunk's new code (internal/install/draftsources.go
   lock replay, marker v5, etc.): every new collapse site trunk introduced must be migrated to the stateread seam or allow-listed with
   a reason, and the guard's counts/ratio updated honestly — list each. Leave nothing staged.
3. VERIFY before refresh: `git diff --name-only ab34556e` over the working tree must contain ONLY this Story's paths (2as5sx delta + your
   draftsources_test.go + any migrations above) — NOT a revert of trunk files. If any trunk file shows as reverted, stop and report.
4. `task-board worktree refresh-candidate TASK-260925-h4syhu` (187z6x checkpoint replay conflicts only via its `--replay-resolutions`
   template; CHANGELOG.md → trunk bytes).
5. Bounded runs: the guard test, your M1 kill, `go test ./internal/install -run 'LockedNetworkRepository|Draft' -count=1`, go vet on
   touched packages, `GOOS=windows go vet ./internal/install`. Real exit codes.
6. Append "Revision 1b — refresh onto ab34556e" to results, `resource update`, `task-board handoff TASK-260925-h4syhu --role developer`;
   stay in the turn while the gate runs. A write-boundary `policy warn` block is a warning.
