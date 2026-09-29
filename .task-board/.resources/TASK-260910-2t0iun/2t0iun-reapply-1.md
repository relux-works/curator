# TASK-260910-2t0iun — re-apply accepted rev4 on trunk (THE ONLY CURRENT INSTRUCTION)

Revision 4 was ACCEPTED: installer release verification. After that, TASK-260910-3i6vod landed on trunk. It created `SECURITY.md` and
edited `README.md`, and your rev4 also creates `SECURITY.md`, so the files now conflict add/add. Your accepted content is saved as
`refs/campaign/234vmx-rev4-20260929`: parent 3f60f7f0, tree 3db884ae, 6 paths. Your Story worktree is fresh on trunk.
1. `task-board m 'set_status(TASK-260910-2t0iun, status=development)'`.
2. Apply the accepted delta:
   `git diff 3f60f7f0 refs/campaign/234vmx-rev4-20260929 -- . ':!.task-board' > $TMPDIR/2t0iun.patch; git apply --3way $TMPDIR/2t0iun.patch`.
   Build SECURITY.md from trunk's file: keep every section trunk has, verbatim, and add your rev4 sections (installer verification, the
   opt-out, reporting if yours had it) in a sensible place. Do not duplicate headings: when both files have the same heading, merge the
   content under one heading. For README.md, keep both trunk's and your changes. Add nothing else.
3. VERIFY, and paste the output into the results:
   - `git diff --name-only origin/main -- . ':!.task-board'` = the 6 rev4 paths;
   - for install.sh, installer_script_test.go and the two ledgers, the +/- lines are identical to rev4;
   - every line of trunk's SECURITY.md is still present: `git show origin/main:SECURITY.md | grep -vxFf SECURITY.md` prints nothing;
   - every non-heading line that rev4's SECURITY.md added is present.
4. Run `go test ./internal/install -run InstallScript` and record the real exit code.
5. Append "Revision 5 — re-apply on <trunk sha> (SECURITY.md merged with 3i6vod)" to the results, run `resource update`, then
   `task-board handoff TASK-260910-2t0iun --role developer`, then END YOUR TURN. The runner publishes the CR and runs the gate; do not wait
   for it. No CHANGELOG/LOGBOOK edit. Never spell any employer name.
