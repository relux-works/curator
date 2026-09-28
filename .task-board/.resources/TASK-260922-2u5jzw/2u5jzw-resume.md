# TASK-260922-2u5jzw — F-L1b resume (THE ONLY CURRENT INSTRUCTION; `2u5jzw-brief.md` still defines the deliverable)

Your Stop-The-Line was correct and is RESOLVED: skill-agents-management **v0.5.20** (signed tag, commit 1fef4b2)
implements Decision 0018 choice 4 inside the module — under yolo it refuses the known conflicting selectors (Claude
`--permission-mode`, `--allow-dangerously-skip-permissions`, `--restricted`, duplicate bypass; Codex `-a/--ask-for-approval`,
`-s/--sandbox`, `--approve-for-me`, `--dangerously-bypass-*`, `-c/--config` approval_policy|sandbox_mode|sandbox_permissions;
`=` and separate forms, codex `exec` placement) with `ErrNativePolicyConflict` and exported
`*NativePolicyConflictError{Selector, Placement}` (errors.Is / errors.As). Grammar token is now
**`permission-grammar-v2`** for Claude and Codex (Pi stays v1). Native performs no argv inspection.

1. `task-board m 'set_status(TASK-260922-2u5jzw, status=development)'` (leave `blocked`).
2. Pin `github.com/relux-works/skill-agents-management v0.5.20` (`go get …@v0.5.20`, `go mod tidy`). Cite
   `permission-grammar-v2` wherever the launcher cites the grammar token (SPEC 0.5.0-draft on your branch says v1 in
   places? — if the launcher SPEC names the token, record the discrepancy in results as a SPEC follow-up; do NOT edit
   the SPEC in this leaf).
3. Map the module's typed conflict error to the launcher's `usage` exit (2) per SPEC §6 — by `errors.Is/As`, never by
   spelling a flag. The launcher-side conflict rows then drive the REAL `curator run` entry with a fake tool and assert the
   mapped diagnostic and exit, per selector family, including the `exec` placement.
4. Everything else exactly as `2u5jzw-brief.md`: all choice-5 row families, provider-spelling scan, narrowing mutants in a
   disposable copy, `make check`, CHANGELOG, results, checklist, `task-board handoff TASK-260922-2u5jzw --role developer`.
A `run_wrote_outside_worktree … policy warn` block is a warning — verify status `to-review`. Fake tools only.
