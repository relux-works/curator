# TASK-260927-31gaka — rework 1 (THE ONLY CURRENT INSTRUCTION)

Review rev1 = CHANGES REQUESTED (`TASK-260927-31gaka_review-verdict-rev1.md`), one blocking finding F1: a 20 MB locally built `curator`
Mach-O binary sits at the repository root of the candidate. The code itself was judged correct.
1. `task-board m 'set_status(TASK-260927-31gaka, status=development)'`.
2. `rm curator` in the Story worktree (build binaries to $TMPDIR only). VERIFY `git status --short` and `git diff --name-only origin/main --
   . ':!.task-board'` list only the 7 leaf paths (conformance-gaps.tsv, env_credential_marker_test.go, envstatus_test.go,
   codex_seed_test.go, managed.go, status.go, envregistry.go). Change nothing else.
3. Append "Revision 2 — stray binary removed" to the results, `resource update` it, `task-board handoff TASK-260927-31gaka --role
   developer`, END YOUR TURN. No CHANGELOG/LOGBOOK edit.
