# BUG-261004-13ptlq: unmanage-cannot-restore-own-symlink-backup

## Description
Report: docs/security-audit-2026-10-inline.md, finding N4 (priority: medium). Takeover of a foreign symlink correctly backs up the link itself (switch.go), but readBackupTree (unmanage.go) accepts only regular files, so `env unmanage --restore-backups` exits 1 with environment_backup_record_unreadable: backup entry is not a regular file. Not a repeat of E5 (E5 was writing through a symlink).

## Scope
internal/envprofile (unmanage.go readBackupTree/restore, switch.go backup types), cmd/curator env unmanage regressions

## Acceptance Criteria
1. Backup and restore support the same entry types; a saved link is restored as a link, atomically, without following its target.
2. Links in the parent route are still refused (no weakening).
3. Red-first full CLI round-trip: foreign symlink → profile use --takeover → env unmanage --restore-backups restores the identical link target; the external target file is byte-identical and untouched.
4. Lands after the N3 leaf (same restore plan), rebased on it.
