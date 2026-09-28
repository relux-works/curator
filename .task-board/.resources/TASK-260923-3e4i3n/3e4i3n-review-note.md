# Review note — TASK-260923-3e4i3n F-M1c (skill-agents-management), CR revision 2 (orchestrator, binding)

Review against `3e4i3n-brief.md` and the results. Read-only; disposable clone for anything you run.
1. Decision 0018 choice 4 conflict table (curator-spec main `decisions/0018-curator-run-permission-interface.md`)
   implemented exactly: Claude `--permission-mode`, `--allow-dangerously-skip-permissions`, `--restricted`, duplicate
   `--dangerously-skip-permissions`; Codex `-a/--ask-for-approval`, `-s/--sandbox`, `--approve-for-me`, any
   `--dangerously-bypass-*`, `-c/--config` keys `approval_policy`, `sandbox_mode`, `sandbox_permissions` — refused
   under yolo in `=` and separate forms, aliases, codex `exec` placement. Nothing missing, nothing extra.
2. **Native mode performs no argv inspection** (raw contract unchanged) — a refusal under native is a finding.
3. Exported stable typed errors usable with errors.Is/As by the launcher; README Permission-mode no longer says
   "later leaf"; CHANGELOG 0.5.20; version constant bumped per repo convention (v0.5.19 is the operator's).
4. Rows per selector × placement executed; non-conflicting selectors forwarded; mutant table present. Apply one
   mutant of your own (drop one selector from the table, or inspect under native) → killed?
5. Gate green (runtime validation log). No real provider CLI invoked by tests.
Findings → changes requested with file:line; else accept_cr. No LOGBOOK.md.
