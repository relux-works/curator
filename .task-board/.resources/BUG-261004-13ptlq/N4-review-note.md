# Review note — BUG-261004-13ptlq (N4) symlink backup restore (orchestrator, binding)
Same-provider review (R164), astra medium. Brief: n4-brief.md. Source: docs/security-audit-2026-10-inline.md §N4 (#106). Builds on N3 (505e1526).
Verify, with real exit codes (GOFLAGS=-work, targeted packages only):
1. Red first. The CLI round trip is: foreign symlink, then profile install, then profile use --takeover, then env unmanage --restore-backups. It failed on the base (environment_backup_record_unreadable) and passes on the candidate, which restores the identical link target.
2. Restore never follows the link. The external target file stays byte-identical. Links in the parent route are still refused, with no weakening.
3. The N3 mode behaviour is unchanged (0600 stays 0600).
4. A mutant that follows the link on restore is killed.
5. Cite the hosted CR gate (all OSes). No CHANGELOG.md or LOGBOOK.md edits.
accept_cr if all hold; otherwise request changes.
