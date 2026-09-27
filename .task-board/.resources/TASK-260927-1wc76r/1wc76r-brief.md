# TASK-260927-1wc76r — env unmanage --restore-backups (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md` and the task description. Find the normative text in curator-spec v1.0.0-rc.13 (environments: unmanage,
takeover backups / backup records; cite clauses) and the vectors (backup-record-unreadable-restore-stops, backup-record-absent-restore-nothing,
and every other unmanage/restore vector). Implement `curator env unmanage --restore-backups` through the CLI production entry, reusing the
existing takeover/backup record code (0017 cww1ov, dotfile manager table bi6ouz) and the internal/stateread seam (§8.4.1: unreadable ≠ absent).
Drive every such vector through the production entry; remove the gap rows owned by this task (TASK-260918-ryh3kw attributed the two
backup-record vectors to it) — report before/after. Mutants: unreadable record treated as absent; restore performed without a record — killed.
No CHANGELOG/LOGBOOK edit (entry text in results). Focused bounded runs (host memory). Attach results, check DoD,
`task-board handoff TASK-260927-1wc76r --role developer`. A write-boundary `policy warn` block is a warning.
