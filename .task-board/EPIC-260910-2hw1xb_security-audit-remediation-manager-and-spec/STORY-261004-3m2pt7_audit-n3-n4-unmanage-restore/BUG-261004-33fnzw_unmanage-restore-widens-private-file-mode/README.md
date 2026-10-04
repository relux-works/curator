# BUG-261004-33fnzw: unmanage-restore-widens-private-file-mode

## Description
Report: docs/security-audit-2026-10-inline.md, finding N3 (priority: medium). internal/envprofile/unmanage.go loads backups as map[string][]byte, losing mode, and always writes 0644. A 0600 CLAUDE.md saved by `profile use --takeover` (backup kept 0600) is restored by `env unmanage --restore-backups --env claude_code` as 0644, exit 0. environments §8.3 says hand-maintained context backups may hold secrets.

## Scope
internal/envprofile (unmanage.go restore plan, switch.go backup), cmd/curator env unmanage regressions

## Acceptance Criteria
1. The restore plan carries each backup entry's type and mode; restore writes atomically and never widens the original permissions.
2. Red-first CLI round-trip regressions (install → use --takeover → unmanage --restore-backups): 0600 stays 0600; 0644 stays 0644; an executable mode is preserved where the type is supported.
3. Unix permission assertions; Windows ACL behaviour stated explicitly (tested or bounded with reason).
