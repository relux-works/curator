# THE ONLY CURRENT INSTRUCTION — BUG-261004-13ptlq (N4): unmanage cannot restore its own symlink backup (developer, code)
Source: docs/security-audit-2026-10-inline.md §N4 (#106). The acceptance criteria are on the element. N3 (mode-preserving restore plan, 505e1526) is on main. Build on its restore plan so that it carries entry TYPE as well as mode.
1. RED FIRST, as a full CLI round trip: a foreign symlink CLAUDE.md, then profile install, then profile use --takeover, then env unmanage --restore-backups. Today this exits 1 with environment_backup_record_unreadable.
2. Fix: backup and restore support the same entry types. A saved link is restored as a link, atomically, without following its target. Links in the parent route are still refused.
3. The external target file must stay byte-identical and untouched. Kill the mutant that follows the link on restore.
## Mode (tb-R181, after a healthy mini restart)
Full codex pace. Run targeted tests with GOFLAGS=-work and keep the build lock. The hosted CR gate is the arbiter. Never edit LOGBOOK.md or CHANGELOG.md. Before the handoff, append a section to your task results resource and `resource update` it, because a handoff without a NEW or UPDATED outcome builds no Change Request.
Then `task-board handoff BUG-261004-13ptlq --role developer` and END YOUR TURN.
