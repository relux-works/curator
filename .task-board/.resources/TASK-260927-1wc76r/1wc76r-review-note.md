# Review note — TASK-260927-1wc76r env unmanage --restore-backups (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review against `1wc76r-brief.md` and curator-spec v1.0.0-rc.13 (unmanage + takeover backup records; cite clauses). Through the CLI: backups
recorded at takeover are restored; an unreadable backup record stops the restore with the typed read-failure diagnostic (stateread seam,
§8.4.1) — never treated as absent; an absent record restores nothing; no partial writes on failure (check atomicity/journal as the spec
requires). The two backup-record vectors and every other unmanage vector pass; their gap rows (attributed by ryh3kw) removed. Kill both
listed mutants yourself. Consistent with 0017 takeover code and bi6ouz dotfile table. No CHANGELOG/LOGBOOK, no stray files, gate green.
accept_cr or changes requested with file:line. No LOGBOOK.md.
