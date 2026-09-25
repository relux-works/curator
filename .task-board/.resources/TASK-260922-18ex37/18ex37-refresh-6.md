# TASK-260922-18ex37 — rework 6: remove the artefact AND refresh onto 96e3f272 (THE ONLY CURRENT INSTRUCTION; bound developer run)

Revision 5: ONE blocking finding (verdict rev5 F1) — remove `.github/workflows/ci.yml.merged.tmp`; everything else verified. Publication
of the cleanup then refused `change_request_base_authority_mismatch` because trunk moved to `96e3f272` (11jgkt internal/snapshot, 10d3l1
internal/install/draftevidence_test.go — both DISJOINT from this Story). Loop bound (notes): revision budget 1, reference = rev5 paths minus
the artefact (61 paths).
1. `task-board m 'set_status(TASK-260922-18ex37, status=development)'` if needed.
2. Delete `.github/workflows/ci.yml.merged.tmp` and any other untracked `*.tmp`/`*.orig`/`*.rej`/`*.merged*` (list them).
3. Combine trunk: `git diff ab34556e 96e3f272 -- . ':!.task-board' ':!CHANGELOG.md' | git apply --3way` (disjoint — must apply cleanly); leave
   nothing staged. VERIFY `git diff --name-only 96e3f272 -- . ':!.task-board'` over the working tree == the 61 reference paths (no trunk revert).
4. `task-board worktree refresh-candidate TASK-260922-18ex37` (3bbvrs replay via `--replay-resolutions` only; CHANGELOG → trunk bytes).
5. Append "Revision 6 — artefact removed, refresh onto 96e3f272" to results, `resource update`, `task-board handoff TASK-260922-18ex37 --role
   developer`; stay in the turn while the gate runs. A write-boundary `policy warn` block is a warning.
