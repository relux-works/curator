# TASK-260918-ryh3kw — orchestrator decision (THE ONLY CURRENT INSTRUCTION, with ryh3kw-sec-brief.md)

Decision: RE-ATTRIBUTE. The rc.13 vectors backup-record-unreadable-restore-stops and backup-record-absent-restore-nothing need the
`env unmanage --restore-backups` production entry, which is a separate surface now owned by TASK-260927-1wc76r (Story STORY-260927-22m88w). Keep both as gap rows in
.github/ci/conformance-gaps.tsv owned by TASK-260927-1wc76r with the exact blocker ("no env unmanage --restore-backups production entry"); the AC item
covering them is satisfied by that attributed row (cite this resource). Everything else of this leaf stands — finish it and hand off.
`task-board m 'set_status(TASK-260918-ryh3kw, status=development)'` first. No CHANGELOG/LOGBOOK edit.
