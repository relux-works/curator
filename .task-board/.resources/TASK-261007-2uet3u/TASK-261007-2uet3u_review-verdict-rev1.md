# TASK-261007-2uet3u — curator v0.15.0-rc.4 release notes review, revision 1

Verdict: accepted. No findings requiring rework.

Reviewed CR-TASK-261007-2uet3u-1 revision 1: base b89427c1dee8c9f3863d06118de7275fc7d5978d; candidate tree 67d613f21f8303e261f828b85ba26d2f84096513. Fresh git ls-remote origin refs/heads/main equals the base and origin/main. The working CHANGELOG matches the candidate. Run goal query reported no goal binding.

## Swept surfaces

| Surface | Evidence and result |
|---|---|
| Scope and format | Exact CR delta: CHANGELOG.md only, 81 insertions / 8 deletions; no LOGBOOK changes. Empty Unreleased precedes the requested dated v0.15.0-rc.4 heading and all five required subsections. The explicit heading instruction takes precedence over rc.3 punctuation. |
| History completeness | Read all 44 subjects and non-board changed paths. 14/14 non-board commits represented (8 implementation and 6 research); 30/30 board-record commits excluded from product claims. Producer reconciliation outcome is attached and consistent. |
| v2 cutover and migration | afcf1898; EnableV2Writers=true, identity_migration.go, migration-plan-v3, migrate/resolve/use production paths and rc14_cutover_test.go support atomic rehash, rollback, plan drift refusal and recovery. Frozen marker framing qualifications and conformance gap closure checked. |
| v1-only NUL gate and operator actions | 7d809c5f; audit gate scans only v1 before hashing; revocations match the selected framing digest without legacy recomputation. audit --allow calls PinAtVersion with WriteVersion. Pin and verdict readers check framing versions; old pins cannot authorize v2. Re-pin and re-issue bare revocations are correct; source revocations remain supported. |
| N1–N5 | GC uncertainty prevents runtime/build sweeping; admission budgets expanded paths and canonical bytes; unmanage writes original modes or symlink entries atomically with parent-route refusal; strict decisions apply findings despite a pin, cached and fresh, after revocations. No inflated claims found. |
| Go qualification | 9bc8e1a1 adds 1.26/1.27 families and per-family CI; future families still refused. |
| Research | Six research-only commits, correctly described as documentation with no product behavior changes. |
| Known issues | Windows risk retained from rc.3 rather than described as fixed. B flips absent from history and retained gap ownership corroborates the deferral. spm#537 tooling-only statement is explicitly supplied by the approved task brief, not inferred as a product change. |

## Entry-to-commit mapping

| Entry | Commits |
|---|---|
| Added: Go families | 9bc8e1a1 |
| Added: research CIP-0002 / 0003 / 0004 / 0005 / 0006 / CSK | 5d2b9437 / 481c1fba / 071826c8 / 332b95d4 / ddd23010 / 8fb50937 |
| Changed: writer cutover and identity migration | afcf1898 |
| Changed: operator migration, re-pin, revocation instructions | afcf1898, 7d809c5f |
| Changed: both NUL paragraphs, cache/pin framing and pinned context readers | 7d809c5f |
| Fixed: N3 original modes | 505e1526 |
| Fixed: N4 symlink backups | 1fc0a93a |
| Security: N1 GC | 052f764c |
| Security: N2 expanded-snapshot limits | 371f250d |
| Security: N5 strict findings with pins | 54bed271 |
| Known issues | Windows/B deferrals retained from rc.3 and absence of flips; Go tooling issue supplied in brief. These are risk disclosures, not claims of new product commits. |

## Validation and limits

Reviewer reran git diff --check on the exact CR delta: exit 0.
Reviewer reran 10 selected audit test functions covering framing pins, legacy rejection, mixed v1/v2 NUL scope, cache versions and pre-hash refusal: go test -count=1 -timeout=2m ./internal/audit with a named -run mask; exit 0, package 0.543s.
Reviewer reran 5 migration test functions: TestRC14MigrationRehashesLegacyIdentities, TestRC14IdentityMigrationRollsBackEveryEntry, TestRC14IdentityMigrationRefusesSiblingPlanDrift, TestRC14IdentityMigrationRecoversInterruptedCommit, TestRC14UseMigratesLegacyProfileBeforeNativePublication; go test -count=1 -timeout=2m ./internal/envprofile with that exact name mask; exit 0, package 34.443s.
Reviewer reran TestQualifiedGoFamilies and TestUnknownGoFamiliesRemainRefused: go test -count=1 -timeout=1m ./internal/godriver -run '^Test(QualifiedGoFamilies|UnknownGoFamiliesRemainRefused)$'; exit 0, package 6.236s.
Selected top-level functions passed: 17/17; this is not a total-suite coverage ratio.

Accepted existing evidence rather than rerunning the full remote gate: TASK-261007-2uet3u_change-request_rev1-validation.log reports hosted run 37539137569 success, including platform tests/races, conformance, lint, naming and all Go-family qualification jobs. Independently checked its gate commit 7cb3873f2cfa01bdfb18b17826c72794a164cd46 has the exact candidate tree. Test-case coverage of that full gate remains unknown; rose-air and candidate-suite jobs were skipped as recorded. Windows incident root cause remains unproved.

No code or repository files modified by reviewer. No new anomaly warrants a logbook edit; task scope explicitly forbids it. Non-acceptance checklist branch is not applicable. Acceptance routes to integrating via accept_cr; reviewer neither commits nor closes delivery.

## Full reconciled subjects

```
b89427c1 Record STORY-261006-1fpobd board state
7d809c5f STORY-261006-1fpobd: STORY-261006-1fpobd: opaque-gate-v1-only
5364c4df Record EPIC-261004-19s6jb board state
75633c08 Record STORY-261005-2xothk board state
b99fd9c1 Record STORY-261003-3bx9x0 board state
8cd25160 Record EPIC-261004-19s6jb board state
728f08de Record STORY-261005-2z4y92 board state
fae2ff9c Record STORY-261003-3bx9x0 board state
afcf1898 STORY-261003-3bx9x0: STORY-261003-3bx9x0: v1-to-v2-profile-hash-migration
3d9aa987 Record STORY-261002-2x32ly board state
0afb18b0 Record STORY-261004-1lk8e9 board state
6ddf6eb5 Record STORY-261004-3e03l7 board state
40bc4c6d Record STORY-261004-1lk8e9 board state
77fabd45 Record STORY-261004-3e03l7 board state
8fb50937 STORY-261004-3e03l7: STORY-261004-3e03l7: design-csk-gap-follow-ups
ff8f75a0 Record STORY-261004-2b8pnx board state
555c551e Record STORY-261004-2b8pnx board state
332b95d4 STORY-261004-2b8pnx: STORY-261004-2b8pnx: design-audit-backends-and-secret-transport
3c07e428 Record STORY-261004-7fglii board state
0c21f196 Record STORY-261004-7fglii board state
071826c8 STORY-261004-7fglii: STORY-261004-7fglii: shell-hook-path-and-backend-env-hardening
a9a7ae68 Record STORY-261004-1ffwh8 board state
846a1520 Record STORY-261004-1ffwh8 board state
5d2b9437 STORY-261004-1ffwh8: STORY-261004-1ffwh8: design-launch-project-views
96b39fa1 Record STORY-261004-1pwxri board state
54e8595f Record STORY-261004-1pwxri board state
481c1fba STORY-261004-1pwxri: STORY-261004-1pwxri: claude-macos-managed-home-login-modes
6d49db30 Record STORY-261004-3v2zm2 board state
20513a20 Record STORY-261004-3v2zm2 board state
ddd23010 STORY-261004-3v2zm2: STORY-261004-3v2zm2: legacy-provider-settings-and-mcp-optouts-105
9bc8e1a1 Qualify Go 1.26 and 1.27 in the go-v1 driver (TASK-261004-1nl0kg)
41f2ce6c Record STORY-261004-3kpcxq board state
1fc0a93a Restore a symlink backup as a link without following it (BUG-261004-13ptlq)
94fa72d9 Record STORY-261004-1aqcha board state
54bed271 A trust pin no longer waives strict findings (BUG-261004-16a407)
e5489b6b Record STORY-261004-3m2pt7 board state
505e1526 Restore unmanaged backups without widening file mode (BUG-261004-33fnzw)
7ec86ac4 Record STORY-261004-3oognx board state
98de13ef Record STORY-261004-3oognx board state
052f764c STORY-261004-3oognx: STORY-261004-3oognx: inline-audit-2026-10-remediation
a2df5887 Record STORY-261004-o4s9aq board state
73474998 Record STORY-261004-o4s9aq board state
371f250d STORY-261004-o4s9aq: STORY-261004-o4s9aq: audit-n2-expanded-snapshot-budget
f3f29295 Record STORY-261002-2327ef board state
```
