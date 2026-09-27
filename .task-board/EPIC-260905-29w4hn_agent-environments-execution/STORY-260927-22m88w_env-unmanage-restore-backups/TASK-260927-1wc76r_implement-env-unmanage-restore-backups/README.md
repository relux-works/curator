# TASK-260927-1wc76r: implement-env-unmanage-restore-backups

## Description
Implement curator env unmanage --restore-backups per curator-spec v1.0.0-rc.13 (environments: unmanage and backup records; cite the clauses). Restore backups recorded at takeover; an unreadable backup record stops the restore with the typed read-failure diagnostic (§8.4.1 discipline, stateread seam); an absent record restores nothing. Drive the rc.13 vectors backup-record-unreadable-restore-stops and backup-record-absent-restore-nothing (and every other unmanage vector) through the production entry and remove their gap rows.

## Scope
(define task scope)

## Acceptance Criteria
env unmanage --restore-backups through the CLI; the two backup-record vectors and all unmanage vectors pass with no gap rows; mutants (unreadable treated as absent; restore without record) killed; no CHANGELOG edit
