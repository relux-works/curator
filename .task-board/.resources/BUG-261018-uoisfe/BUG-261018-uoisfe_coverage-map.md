# Native skills coverage map — rework round 7

Candidate `8c5043959ce201bacb9ee73ce8d66d0d3a1f10e4`. Coverage: **12 of 12 surface rows driven; 5 of 5 AC rows driven**.
Rows 1–11 retain their accepted implementation and tests. Row 12 enforces the five-operation admission before help and dispatch, including true repair and false dry-run. The current validation report distinguishes fresh execution from identity-checked reuse.

| Surface row | Production entry and call site | Attacking tests | Killed narrowing mutant | Out-of-contract inputs |
|---|---|---|---|---|
| 1. Absent target | `UseWithPolicy` / `SyncWithPolicy` → `materializeOne`, switch.go:690; CLI `run` → install/use/sync/update/resolve | `TestNativeSkillsUseAndSync`, `TestNativeSkillsInstallAndUpdate`, `TestNativeSkillsCLIOperations`; existing `TestGlobalAddPublishesProfileLockBeforeNativeMaterialization` and `TestGlobalInstallPublishesMigratedLockBeforeNativeMaterialization` | `pi-native-only` → `TestNativeSkillsUseAndSync` | None claimed |
| 2. Recorded manager link | `Resolve` → `nativeSkillsState`, native_skills.go:65; `repairUnderLock`, managed.go:2648 | `TestNativeSkillsRepair` (status, restore, dry-run and non-current bounds), `TestNativeSkillsUseAndSync` | `codex-link-drift` → `TestNativeSkillsRepair` | None claimed |
| 3. Foreign symlink | `UseWithPolicy` → `materializeOne` inventory, switch.go:690; five CLI operations through `run` | `TestNativeSkillsTakeover/foreign`, `TestNativeSkillsCLIOperations` | `dangling-foreign-consent` → `TestNativeSkillsTakeover/foreign` | None claimed |
| 4. Other unrecorded entry | `UseWithPolicy` → `materializeOne` inventory, switch.go:690 | `TestNativeSkillsTakeover/directory`, `/file`, `/store`, `/relative-store` | `plain-file-consent` → `TestNativeSkillsTakeover/file` | None claimed |
| 5. Path components | `UseWithPolicy` → `managedPath` via `inPlaceSkillsLedger` and `materializeOne` | `TestNativeSkillsLinkedParent/ordinary`, `/takeover` | `skills-parent-traversal` → `TestNativeSkillsLinkedParent` | None claimed |
| 6. Unnamed entries | five CLI operations through `run` → native materializer; `SyncWithPolicy` and `UseWithPolicy` | `TestNativeSkillsCLIOperations`, `TestNativeSkillsConvergenceAndSwitch`, `TestNativeSkillsEmptyProfileLeavesUnnamedParent` | `unnamed-takeover-directory` → `TestNativeSkillsConvergenceAndSwitch` | None claimed |
| 7. Adapter ledger | `UseWithPolicy` → `inPlaceSkillsLedger`, native_skills.go:31; `StatusOf` → `scopeHomes`, status.go:1023 | `TestNativeSkillsLedger`, `TestNativeSkillsUnrecordedLedger` | `one-ledger-name` → `TestNativeSkillsLedger` | None claimed |
| 8. Recorded old staging | `SyncWithPolicy` → `materializeOne` recorded-surface removal | `TestNativeSkillsConvergenceAndSwitch` | `one-old-staging-name` → `TestNativeSkillsConvergenceAndSwitch` | None claimed |
| 9. Profile switch | `UseWithPolicy` → `materializeOne` removal/replacement; scoped selection | `TestNativeSkillsConvergenceAndSwitch`, `TestNativeSkillsInstallAndUpdate` | `one-retired-name` → `TestNativeSkillsConvergenceAndSwitch` | None claimed |
| 10. Backup generation | `UseWithPolicy` → `openBackupWithDirectoryRead`, switch.go:1013 | `TestNativeSkillsBackupCollision` (collision and next-generation positive control) | `backup-directory-collision` → `TestNativeSkillsBackupCollision` | None claimed |
| 11. Unreadable marker | `UseWithPolicy` → native `envmarker.Read` before inventory and writes | `TestNativeSkillsUnreadableMarker` | `directory-marker-is-absence` → `TestNativeSkillsUnreadableMarker` | None claimed |
| 12. Flag acceptance | `cli.run` main.go:283 -> `rejectsTakeover` command_help.go:128, before every help return; repair/dry-run gate :162 | `TestNativeSkillsTakeoverEveryRegisteredCommand`; `TestNativeSkillsTakeoverArgumentDiscardRegression`; `TestNativeSkillsTakeoverActivationBoundary`; `TestNativeSkillsCLIOperations`; `TestReviewNativeTakeoverHelpBoundary`; `TestNativeSkillsTakeoverHelpCrossProduct`; all retained reviewer tests | `list-only-consent-bypass` -> both named list regressions; `claude-takeover-without-repair` -> `TestNativeSkillsTakeoverFlagBoundary/env-resolve-claude_code`; `resolve-help-missing-repair` and `resolve-help-dry-run` -> `TestReviewNativeTakeoverHelpBoundary` | None claimed; `--` ends flag parsing by the documented argument boundary |

| AC row | Production entry and call site | Attacking tests |
|---|---|---|
| 1. Five operations, three adapters | CLI run → profile install/use/sync/update and cmdEnvResolve; materializeOne (switch.go:690), syncCurrentNativeSkills (native_skills.go:209), repairUnderLock (managed.go:2648) | TestNativeSkillsCLIOperations, TestNativeSkillsUseAndSync, TestNativeSkillsInstallAndUpdate, TestNativeSkillsRepair |
| 2. Closed takeover and no-follow | UseWithPolicy → materializeOne inventory, managedPath, openBackupWithDirectoryRead | TestNativeSkillsTakeover, TestNativeSkillsLinkedParent, TestNativeSkillsBackupCollision, TestNativeSkillsUnreadableMarker, TestNativeSkillsCLIOperations |
| 3. Ledger entries retained/reported | UseWithPolicy → inPlaceSkillsLedger; StatusOf → scopeHomes/nativeSkillsState | TestNativeSkillsLedger, TestNativeSkillsUnrecordedLedger |
| 4. Recorded staging converges | SyncWithPolicy → materializeOne recorded removal | TestNativeSkillsConvergenceAndSwitch |
| 5. Twelve attack rows, temp homes | Real UseWithPolicy/SyncWithPolicy/Install/Update/Resolve/StatusOf and CLI run entry points in the twelve-row map | TestNativeSkillsNarrowingMutants and all named native/CLI tests; temporary HOME/USERPROFILE/native-home fixtures; 12/12 existing narrowing controls; local and hosted results are reported by exact candidate in the validation resource |
Every gate row (3, 4, 5, 6, 7, 11) retains its accepted narrowing control.
Row 12 additionally has list-only and two help-specific narrowing mutants. The parent mutation
harness requires a named failing behavioral test and actual child exit 1;
setup, compiler and timeout failures do not count as kills.

No surface-table row is treated as out of contract. The brief explicitly excludes
Muse, opencode and managed-home layout changes. Open spec questions are kept in
the gaps resource: Muse multi-root ownership, adapter-ledger coexistence and
opencode machine-global native-surface ownership.

Registry sweep controls: every unrelated command is driven without help and
with help and must return exit 2 for consent. The same help invocation without
consent succeeds for every command. Group topics and interactive operations use
help as their positive control; actual no-flag list/update/gc/version controls
are separately exercised. This does not claim successful live UI/login actions.


## Rework round 7 amendment

Candidate `8c5043959ce201bacb9ee73ce8d66d0d3a1f10e4`. Coverage: 12 of 12 surface rows, 5 of 5 AC rows driven. No row is treated as out of contract. Row 12 uses `cli.run` (main.go:283) -> `rejectsTakeover` (command_help.go:128), before every help return. Repair/dry-run parsing is at :144/:146; help-topic normalization :154; resolve activation gate :162. Tests add permanent `TestReviewNativeTakeoverHelpBoundary` and `TestNativeSkillsTakeoverHelpCrossProduct`. The narrowing `resolve-help-missing-repair` and `resolve-help-dry-run` plants are both killed by that reviewer witness, observing forbidden successful help, with named parent control subtests. Old list and claude controls retain their names; the claude plant now targets the shared gate. All rows 1–11 retain their accepted implementation, test names and plants; all current native controls replayed in group 8. The gaps resource has the 72-topic/595-row help table, explicit old-survivor bound, and scope/spec questions.

### Round-7 validation and candidate identity

Signed candidate `8c5043959ce201bacb9ee73ce8d66d0d3a1f10e4`, tree `769581956b4fd7bfd02c2567d7465b60335f5b4b`, signature G. Worktree remains uncommitted on checkpoint `5d8e6b07f7f485557d9bbb065c9692348ca065c2` and matches that tree; actual index is empty. Fresh trunk `b9caa20901ee859eec20e649940693033e6297d4` has no candidate-path overlap.

Full composed local suite: 85/85 served packages, 3497 top-level names. Fresh: every CLI group (15), envprofile groups 8/11, full scriptworker/scriptpolicy/crossconformance, and the three install binary consumers. Reused unchanged identity: 79 packages and unchanged envprofile/install tests, explicitly bounded in validation-r7 and reuse-identity.json. All current direct test processes exited 0; missing-group evidence control intentionally exited 1. Raw and sanitized platform-case gates, build, vet, lint, format and signed-tree integrity checks exited 0.

[Hosted full gate](https://github.com/relux-works/curator/actions/runs/39158440563) succeeded on the exact candidate: 20 successful jobs and two conditional skips, no failed job or infrastructure failure. Historical failures and superseded cancelled runs are separated from current passing evidence. The configured post-turn validation remains a subsequent gate, with no advance success claim.

See BUG-261018-uoisfe_validation-r7.md, BUG-261018-uoisfe_validation-logs-r7.txt and BUG-261018-uoisfe_evidence-manifest-r7.json for exact commands, real exits, stream coverage and hashes.
