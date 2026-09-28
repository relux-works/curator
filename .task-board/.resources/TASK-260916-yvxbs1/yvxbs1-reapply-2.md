# TASK-260916-yvxbs1 — re-apply accepted rev5 on trunk 97e85642 (THE ONLY CURRENT INSTRUCTION)

Revision 5 was ACCEPTED. TASK-260916-1zgucp (E1 signers + system delta) landed (97e85642) and the three-way merge conflicts on
internal/envprofile/envprofile.go only. Accepted content: `refs/campaign/wgt8vz-rev5-20260928` (parent 86552087, tree 3e55fd15). Your
Story worktree is fresh on trunk.
1. `task-board m 'set_status(TASK-260916-yvxbs1, status=development)'`.
2. `git diff 86552087 refs/campaign/wgt8vz-rev5-20260928 -- . ':!.task-board' > $TMPDIR/yvxbs1.patch; git apply --3way $TMPDIR/yvxbs1.patch`;
   resolve envprofile.go KEEPING BOTH SIDES (E1's update/delta/confirmation path + your path-source boundary preflight and
   default-update ordering). Add nothing else.
3. VERIFY `git diff --name-only origin/main -- . ':!.task-board'` = the 32 rev5 paths.
4. Focused runs (real exit codes): `go test ./internal/envprofile -run 'Path|Boundary|Overlay|Update|Delta|Signer|Guarded'` (split),
   `go test ./cmd/curator -run 'Profile|Path|Delta'`.
5. Append "Revision 6 — re-apply on 97e85642", `resource update`, handoff, END YOUR TURN. No CHANGELOG/LOGBOOK edit.
