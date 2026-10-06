# TASK-261007-2uet3u — rc.4 reconciliation map

Range: `git log v0.15.0-rc.3..origin/main` = 44 commits (14 product, 30 board-record with no product change).

## Entry → commit map

| rc.4 entry | Commit(s) |
|---|---|---|
| Added: go-v1 admits Go 1.26/1.27 (curator#87) | 9bc8e1a1 Qualify Go 1.26 and 1.27 (TASK-261004-1nl0kg) |
| Added: docs-only research (CIP-0002..0006 + CSK note) | 5d2b9437 (CIP-0002), 481c1fba (CIP-0003), 071826c8 (CIP-0004), 332b95d4 (CIP-0005), ddd23010 (CIP-0006), 8fb50937 (CSK gap follow-ups) — all .research-only |
| Changed: v2 writer flip + atomic v1→v2 migration | afcf1898 v1-to-v2-profile-hash-migration (CR-TASK-261003-1uzji7-5; EnableV2Writers=true, identity_migration.go, byte-exact-snapshot gap closed) |
| Changed: operator action (migration + revocations + re-pin) | afcf1898 (migration rides env migrate/resolve/use) + 7d809c5f (bare v1 revocations re-issued from v2 digest; version-scoped pins/verdicts) |
| Changed: NUL gate scoped to v1; refusal precedes v1 computation | 7d809c5f opaque-gate-v1-only (CR-TASK-261006-3ptm2x-7; text carried over from Unreleased + merged operator-action bullet) |
| Fixed: restore keeps file mode (audit N3, #106) | 505e1526 (BUG-261004-33fnzw) |
| Fixed: symlink backup restored as link (audit N4, #106) | 1fc0a93a (BUG-261004-13ptlq) |
| Security: GC skips sweep on uncertain refs (audit N1, #106) | 052f764c (CR-BUG-261004-bknio5-1) |
| Security: expanded-snapshot budget (audit N2, #106) | 371f250d (CR-BUG-261004-2v9pbz-1; + docs/repository-admission-limits.md) |
| Security: pin no longer waives strict findings (audit N5, #106) | 54bed271 (BUG-261004-16a407) |
| Known: Windows broker flake TASK-260930-fp8vx7 | carried over from rc.3; no new commit; still documented risk |
| Known: security_posture rev B + Codex seed rev B held | absence verified: no flip commit in range; codex-seed B gaps still owned in conformance-gaps.tsv |
| Known: Go board-close gap tooling-only (spm#537) | per brief; no product commit |

## Full subject list reconciled (44)

Product (14): 7d809c5f, afcf1898, 8fb50937, 332b95d4, 071826c8, 5d2b9437, 481c1fba, ddd23010, 9bc8e1a1, 1fc0a93a, 54bed271, 505e1526, 052f764c, 371f250d (subjects as in map above).
Board-record, no product change (30): b89427c1, 5364c4df, 75633c08, b99fd9c1, 8cd25160, 728f08de, fae2ff9c, 3d9aa987, 0afb18b0, 6ddf6eb5, 40bc4c6d, 77fabd45, ff8f75a0, 555c551e, 3c07e428, 0c21f196, a9a7ae68, 846a1520, 96b39fa1, 54e8595f, 6d49db30, 20513a20, 41f2ce6c, 94fa72d9, e5489b6b, 7ec86ac4, 98de13ef, a2df5887, 73474998, f3f29295.

## Notes

- Heading follows the brief literally (`## v0.15.0-rc.4 — 2026-10-07`); earlier sections use `## 0.15.0-rc.3 - date` (no v, hyphen). Deliberate per instruction; reviewer may normalize.
- Only CHANGELOG.md changed (git status: ` M CHANGELOG.md`); no LOGBOOK edits per scope.
- Docs-only: no code/test/build gate applies (no test or tool references CHANGELOG); none run.
- Logbook item N/A: nothing beyond the notes themselves; LOGBOOK edits forbidden by task.
