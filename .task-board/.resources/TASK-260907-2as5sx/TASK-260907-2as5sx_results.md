# TASK-260907-2as5sx results

## Implementation

Added `internal/stateread` with typed `present`, `absent`, and `unreadable` outcomes for exact-path file, directory, and metadata reads. Only a proven not-exist result is absence. Read, stat, directory-list, decode, and validation failures retain the path and `manager_state_unreadable`; required paths can report typed `manager_state_absent`.

Migrated the review-cycle readers and related manager-owned state readers to this seam. Marker consumers now use `marker.ReadState`; a production AST check rejects calls to the legacy nil-on-error `marker.Read`. The status and cleanup paths retain unknown/read-failure outcomes, and purge preflights markers before its first write.

The required `rg` inventory found further absence collapses outside the review-cycle list. They are included in the inventory and fixed below. In particular, the global bin and adapter ownership ledgers no longer become empty maps on a failed read; consumer-registry callers can no longer receive an empty list for unreadable state; hybrid edits refuse unreadable manifests; and runtime collection returns an unreadable inventory error instead of skipping that subtree.

`TestKnownAbsenceSensitiveReadersUseSharedSeam` enumerates 32 production reader functions across environment profiles, context state, markers, scopes, adapter ledgers, install, UI, registry, and CLI status. It parses each function and checks its expected seam or typed-reader call, and rejects direct `os.ReadFile`, `ReadDir`, `Stat`, `Lstat`, `Open`, or `IsNotExist` calls inside the enumerated readers. Per-site tests exercise production entries for absence and mode-000 unreadability.

## Read-site inventory

| Site | Manager-owned shape | Absent outcome | Unreadable or invalid outcome |
|---|---|---|---|
| `internal/envprofile/import.go:294` `readRootSurface` (cycle-2 C2-B1) | Environment root marker and root surface | Missing marker allows the documented root-surface import path | Marker or surface read failure becomes a typed import loss with its path |
| `internal/envprofile/import.go:370` `readSkillsLedger` (cycle-1 fix `ee6743a2`) | Adapter skills ledger | No ledger means the scanned skill is unledgered | Failed or malformed ledger is a typed import loss; entries are not imported as proven unledgered |
| `internal/envprofile/envprofile.go:959` `updateLocked` (cycle-4 C4-M1) | KindPath context snapshot manifest | Missing snapshot is a source-invalid outcome | Unreadable snapshot is a source-invalid outcome carrying `manager_state_unreadable`; update does not silently no-op |
| `internal/envprofile/status.go:527` `scopeHomes` (cycle-6 `status.go:432`) | Managed environment marker | Missing marker is known unprovisioned | Unreadable marker is unknown, non-current, and has a typed marker diagnostic |
| `internal/envprofile/status.go:648` `orphanHomes` (cycle-6 `status.go:521`) | Orphan environment marker and inventory | Missing marker is not an orphan | Unreadable marker or inventory is reported as a diagnostic; unknown state is not counted as empty |
| `internal/envprofile/switch.go:703` `purgeHomes` | Marker governing recorded native surfaces | Missing marker permits the documented purge | Unreadable marker refuses purge before writes, with typed diagnostics |
| `internal/adapters/adapters.go:169` `readLedger` | Project/global adapter ownership ledger | Missing ledger starts with no managed entries | Read/decode/schema failure aborts staging with a typed diagnostic; ledger is not replaced |
| `internal/globalbins/globalbins.go:393` `readLedger` | Global-bin shim ownership ledger | Missing ledger starts fresh | Refresh reports the typed read failure; staged refresh returns it; neither rewrites the ledger |
| `internal/scopes/hybrid.go:34,144,225` hybrid load/add/remove | Machine hybrid declaration manifest | Missing manifest loads empty and may be created by add; remove reports typed absence | Load/add/remove preserve unreadable or invalid state as typed errors and do not overwrite it |
| `internal/scopes/consumers.go:97` `readConsumers` / `LoadConsumers` | Machine consumer registry | Missing registry is an empty set | Failed or invalid registry is a typed error; record/stage callers refuse to overwrite it |
| `internal/scopes/gc.go:335` `sweepRuntime` | Runtime store inventory | Missing runtime root is an empty store | Unreadable root or skill inventory returns a typed error; collection no longer skips it |
| `internal/contextstore/contextstore.go:48` `Exists` | Context-store entry metadata | Missing entry returns `(false, nil)` | Metadata failure returns a typed error instead of `(false, nil)` |

Other candidate manager-state reads were inspected during the same `internal/` and `cmd/` search:

| Sites | Shape and current failure handling |
|---|---|
| `internal/config/config.go:1006`, `internal/config/sourcepolicy.go:312`, `internal/config/sourceproviders.go:54` | Global config and machine repository policy/provider files. Missing paths have their documented absent/default result or a missing-config diagnostic; other OS and decode errors propagate. |
| `internal/hookapproval/hookapproval.go:246,391,502-513` | Approval state and operator-selected hook files. Missing approval state yields no records; unreadable approval state makes status non-current. Missing unrecorded hook files produce no posture row; unreadable files retain an unreadable posture. |
| `internal/contextpkg/contextpkg.go:425,521-536` | MCP manifests and declared context modules. Missing required inputs receive missing/invalid diagnostics; other stat/read errors propagate. |
| `internal/install/generation.go:72` | Project or manager declaration document read exactly once. Absence gets the explicit generation sentinel; open/stat/read failures propagate. |
| `internal/buildcache/cache.go:131`, `internal/buildcache/publish.go:528,620`, `internal/runtimestore/runtimestore.go:230` | Protected build-cache entries and runtime shim inventory. Exact absence maps to cache miss/no-op; other metadata/list failures become untrusted/error outcomes. |
| `internal/contextstore/contextstore.go:94,188` | Path-source inspection and copied tree reads. Missing source paths report `profile_source_path_missing`; other inspection/copy failures report unreadable/source errors. |
| `internal/envprofile/managed.go:533-710,767,1171-1259,1712` | Native seed, managed surface, and profile-collision readers. These paths already use missing-only branches and return/record other read failures; they remain covered by the AST reader inventory where they feed manager state. |

`foreignManagerHint` and XDG shadow-warning probes inspect operator-owned directories for advisory warnings. Their errors do not supply a manager-state default or authorize a state mutation, so they remain outside the manager-owned seam. The scan also includes project/source snapshots, arbitrary operator-selected files, and transaction checks; those are not manager-owned state merely because they call `os.*`.

## Design choice

A shared typed seam was feasible for the discovered manager-state file, directory, and metadata readers, and is used by the 32 enumerated functions directly or through `marker.ReadState`, `envmarker.Read`, and `contextpkg.LoadManifest`. The AST inventory is the backstop for those reader bodies.

A wrapper around every filesystem call in `internal/` and `cmd/` was weighed and rejected: that set includes operator-selected source trees, package snapshots, generation-sensitive open/read operations, and protected-cache APIs with their own ownership and identity rules. Rewriting those through a state-file result type would blur their existing contracts. The selected boundary centralizes absence classification for manager-owned state and explicitly inventories the reviewed reader functions.

The seam also now underlies marker reads used by attestation, GC, install marker inspection, UI, and CLI status. This may make §8.4.1 security-leaf readers in TASK-1ll22r/ryh3kw trivially classify absence versus failure; ownership and acceptance of those leaves remain with their assignees.

## Narrowing mutants

Mutants were applied only in `/tmp/TASK-260907-2as5sx-mutants-20260923`, then restored from the Story worktree. Each mutant changed the listed read failure back to an absence/empty result. Each focused test killed its mutant; the mutant command's actual exit was 1.

| Site narrowed | Named test that killed it | Mutant exit |
|---|---|---:|
| `readRootSurface` marker read | `TestImportUnreadableMarkerIsLoss` | 1 |
| `readSkillsLedger` | `TestImportAbsentAndUnreadableSkillsLedgerDiffer` | 1 |
| KindPath snapshot read | `TestUpdatePathSnapshotUnreadableIsNotAbsent` | 1 |
| Status scope marker | `TestStatusScopeMarkerAbsenceAndUnreadabilityDiffer` | 1 |
| Status orphan marker | `TestStatusOrphanMarkerAbsenceAndUnreadabilityDiffer` | 1 |
| `purgeHomes` marker preflight | `TestRemovePurgeMarkerAbsenceAndUnreadabilityDiffer` | 1 |
| Hybrid declaration read | `TestHybridManifestAbsenceAndUnreadabilityDiffer` | 1 |
| Consumer registry read | `TestConsumerRegistryAbsenceAndUnreadabilityDiffer` | 1 |
| Runtime skill inventory | `TestSweepRuntimeUnreadableSkillInventoryIsNotIgnored` | 1 |
| Global-bin ledger | `TestRefreshAndStageRefuseAnUnreadableOwnershipLedger` | 1 |
| Adapter ledger | `TestStageProjectRefusesAnUnreadableAdapterLedger` | 1 |
| Context-store existence probe | `TestExistsDistinguishesAbsentFromUnreadable` | 1 |

## Verification

Green commands run directly in the Story worktree:

- `go test ./internal/stateread -count=1` — exit 0.
- `go test ./internal/marker -count=1` — exit 0.
- `go test ./internal/contextlock -count=1` — exit 0.
- `go test ./internal/contextpkg -count=1` — exit 0.
- `go test ./internal/envmarker -count=1` — exit 0.
- Targeted `go test ./internal/envprofile -run 'Test(CurrentPointerAbsenceAndUnreadabilityDiffer|ScopedCurrentAbsenceAndUnreadabilityDiffer|ScopedCurrentReadFailureDoesNotFallBackToMachineCurrent|ProfileSourceAbsenceAndUnreadabilityDiffer|UnreadableDefaultProfileSourceDoesNotRecreateDefault|OverlayOwnerLockAbsenceAndUnreadabilityDiffer|ImportAbsentAndUnreadableRootSurfaceDiffer|ImportAbsentAndUnreadableSkillsLedgerDiffer|ImportAbsentAndUnreadableSkillsSurfaceDiffer|UpdatePathSnapshotUnreadableIsNotAbsent|StatusScopeMarkerAbsenceAndUnreadabilityDiffer|StatusOrphanMarkerAbsenceAndUnreadabilityDiffer|StatusUnreadableOrphanInventoryIsNotEmpty|StatusBackupInventoryAbsenceAndUnreadabilityDiffer|RemovePurgeMarkerAbsenceAndUnreadabilityDiffer|KnownAbsenceSensitiveReadersUseSharedSeam|UpdatePathMissingSnapshotIsSourceInvalid)$' -count=1` — exit 0.
- `go test ./internal/adapters -count=1` — exit 0.
- `go test ./internal/scopes -count=1` — exit 0.
- `go test ./internal/globalbins -count=1` — exit 0.
- `go test ./internal/contextstore -count=1` — exit 0.
- `go test ./internal/install -run '^TestConcurrentProjectInstallsPreserveBothConsumers$' -count=1` — exit 0.
- `go test ./internal/install -run '^TestInjectedClockAndGenerationReaderDriveInstallMarkers$' -count=1` — exit 0.
- `go test ./internal/registry -count=1` — exit 0.
- `go test ./internal/ui -count=1` — exit 0.
- Targeted `go test ./cmd/curator` marker/build-root and consumer-GC tests — exit 0.
- `go test ./internal/crossconformance -run '^TestDraftExternalStatusFailsClosedWithoutSource$' -count=1` — exit 0.
- `go vet` across the changed state, install, status, and cross-conformance packages — exit 0.
- `make lint` — exit 0, 0 issues. It emitted a non-blocking generated-file-filter warning because another temporary worktree file disappeared while golangci-lint processed an issue.
- `go build ./cmd/curator` — exit 0; the generated binary was removed.
- `git diff --check` — exit 0.

Resolved intermediate checks: the first `make lint` exited 2 and reported the stale generation-reader fake, one G304 annotation, one ineffassign, and missing exported-constant comments; these were fixed and the final lint run is green. An initial `go test ./internal/scopes -count=1` exited 1 on a missing test import, an initial `go test ./internal/contextstore -count=1` exited 1 on a stale test call after the `Exists` signature change, and an initial `go test ./internal/adapters -count=1` exited 1 on a missing `fmt` import; each was fixed and its final package test exited 0.

`go test ./internal/envprofile -count=1` did not complete within the local command bound. It was terminated after more than 10 minutes and exited 143. The targeted absence/unreadability suite above and the AST inventory test completed successfully. The hosted landing suite was not run manually; the handoff workflow owns its single configured run. Windows permission behavior was not run locally; the mode-000 tests use the repository's existing host-capability skip class.

Important findings and the required no-logbook boundary are recorded here. The campaign rules prohibit editing `LOGBOOK.md` from this Story worktree, so no logbook file was changed.

## Revision 1 — refresh onto 1511b345

### Refresh and merge

- Combined trunk `1511b345c143acfd78b5db0ab4f3176f5ce6ce94` into the candidate using the prescribed `git diff 09b25ef6 origin/main -- . ':!.task-board'` patch and `git apply --3way`. Overlaps in `internal/envprofile/envprofile.go`, `internal/envprofile/status.go`, and `internal/globalbins/globalbins.go` were resolved with both trunk behavior and the state-read migration retained. The checkpoint’s same-source `--use` activation remains in `envprofile.go`; trunk surfacing and its sink are preserved.
- `task-board worktree refresh-candidate TASK-260907-2as5sx --replay-resolutions /tmp/TASK-260907-2as5sx-replay-resolutions.json` — exit 0, `refresh_advanced`; replayed checkpoint `50fad31fa3ff7126dfe35409e4f32ff232f731d9` onto trunk. Its `internal/envprofile/envprofile.go` resolution was bound to the tool’s template with SHA-256 `684298b1c2f5823076936f33f14e69ad0850f1096f48981181e54b404c5957dc`. The refreshed Story tip is `fc8b1cfee1f0d2bd69876c28224e503cc79c3732`.
- Trunk’s new `globalbins.PublishedShims` metadata read is now routed through `stateread.Stat`; unreadable state returns the typed error, and provider posture adds a `manager_state_unreadable` status diagnostic and marks status non-current. The AST inventory now enumerates `PublishedShims` as a shared-seam reader.
- After refresh, restored the stale checkout copy of `.task-board` to refreshed HEAD. `git diff --name-only -- .task-board` is empty (exit 0), and `git diff --cached --stat` is empty (exit 0); no board paths are staged or included in the candidate.

### Verification rerun after refresh

- `go test ./internal/stateread -count=1` — exit 0.
- `go test ./internal/marker -count=1` — exit 0.
- `go test ./internal/contextlock -count=1` — exit 0.
- `go test ./internal/contextpkg -count=1` — exit 0.
- `go test ./internal/envmarker -count=1` — exit 0.
- `go test ./internal/contextstore -count=1` — exit 0.
- `go test ./internal/globalbins -count=1` — exit 0.
- `go test ./internal/adapters -count=1` — exit 0.
- `go test ./internal/envprofile -run '^TestKnownAbsenceSensitiveReadersUseSharedSeam$' -count=1` — exit 0.
- The focused `internal/envprofile` absence/unreadability regex recorded in the Verification section above — exit 0.
- `go test ./internal/scopes -count=1` — exit 0.
- `go test ./internal/install -run '^TestConcurrentProjectInstallsPreserveBothConsumers$' -count=1` — exit 0.
- `go test ./internal/install -run '^TestInjectedClockAndGenerationReaderDriveInstallMarkers$' -count=1` — exit 0.
- `go test ./internal/registry -count=1` — exit 0.
- `go test ./internal/ui -count=1` — exit 0.
- `go test ./internal/crossconformance -run '^TestDraftExternalStatusFailsClosedWithoutSource$' -count=1` — exit 0.
- `go test ./cmd/curator -run '^(TestMarkerDigestsDistinguishesAbsentAndUnreadable|TestContextExposureDistinguishesAbsentAndUnreadableBuildRoots|TestGCPrunesDeadConsumersUnderTheHomeLock|TestProviderInputsForHostScannedBinWarnsWithoutLedger|TestProviderInputsForHostPublishedBinRefuses|TestProviderPostureReportsUnreadablePublishedShimLedger|TestProfileGitReinstallHonoursUseAndTakeover)$' -count=1` — exit 0.
- Narrowing mutant for the new `PublishedShims` read changed unreadable metadata to `(false, nil)`. `go test ./internal/globalbins -run '^TestPublishedShimsTracksTheLedger$' -count=1` killed it with exit 1 (`unreadable ledger = (false, <nil>)`); after restoring the source (SHA-256 matched the saved file), the same test exited 0. The earlier per-site mutants remain as attached above and were not rerun during refresh.
- `go vet ./internal/stateread ./internal/marker ./internal/contextlock ./internal/contextpkg ./internal/contextstore ./internal/envmarker ./internal/envprofile ./internal/adapters ./internal/scopes ./internal/globalbins` — exit 0.
- `go vet ./internal/install ./internal/registry ./internal/ui ./internal/crossconformance ./cmd/curator ./internal/envregistry ./internal/contextmaterialize ./internal/envfragment ./internal/gitops` — exit 0.
- The first two post-refresh `make lint` runs exited 2 on a stale `G304` diagnostic whose path resolved into a removed sibling Story worktree. After clearing the local linter cache, `make lint` exited 0 with `0 issues`.
- `go build -o /tmp/TASK-260907-2as5sx-curator ./cmd/curator` — exit 0.
- `git diff --check` — exit 0.

The configured landing suite was not run manually; handoff owns its single run. All focused rows, the AST inventory, vet, lint, build, and diff checks above were rerun on the refreshed candidate.

## Revision 2 — marker invalid-state regression and Windows mode-bit rows

### Behavior correction

- `marker.ReadState` now keeps three marker outcomes distinct: a missing path is `KindAbsent`; a filesystem read error is `KindUnreadable` with `manager_state_unreadable`; readable but malformed or schema-invalid bytes are `KindPresent` with typed `marker.InvalidError` (`install_marker_invalid`). The schema-version probe uses the same invalid-marker type.
- `marker.Current` treats an invalid marker as non-current with no read error, allowing install staging to re-derive it. Actual read errors still propagate. `markerGeneration` preserves the typed invalid error, and `detectMovedTagsIn` treats that invalid historical generation as stale evidence while refusing a failed read.
- Status, UI, attestation, and GC keep invalid-marker diagnostics separate from failed reads. The GC invalid-marker test expects `install_marker_invalid`; its real unreadable-marker integration fixture now uses a directory at the marker path so it proves an actual read failure rather than malformed JSON.
- Added invalid-marker rows for `ReadState`, `Current`, default marker generation, moved-tag generation handling, the status surfaces, UI skills inventory, registry attestation, and GC. `TestDraftExternalExecutionPolicyInvalidMarkerReDerives` independently drives `install.Project` for the `execution_policy` mismatch while leaving the pinned 94-row matrix's full-run requirement intact.
- Added separate ledger rows for `internal/marker TestCurrentDistinguishesAbsentAndUnreadable`, `internal/marker TestReadStateDistinguishesAbsentAndUnreadable`, and `internal/ui TestSkillsUnderDistinguishesAbsentAndUnreadableMarkers`: `linux,darwin` required, `windows` tolerated as `platform-control`. Each Windows skip prints `platform-control: POSIX mode-bit unreadability`, registered in `skip-classes.tsv`.

### Narrowing evidence

- A `marker.Current` mutant that propagated the typed invalid-marker error was killed by `TestDraftExternalExecutionPolicyInvalidMarkerReDerives` (mutant test exit 1; the repair reported `install_marker_invalid` instead of succeeding). The restored implementation passed the same test normally and under `-race`.
- A `markerGeneration` mutant that relabeled invalid bytes `manager_state_unreadable` was killed by `TestMarkerGenerationDistinguishesAbsentUnreadableAndInvalid` (mutant test exit 1).
- A `detectMovedTagsIn` mutant that refused invalid historical marker state was killed by `TestMovedTagReaderTreatsInvalidMarkerAsStaleButKeepsReadFailure` (mutant test exit 1).
- Earlier mutations to the default `markerGeneration` reader and to `detectMovedTagsIn` each survived the cross-conformance test (exit 0), showing that this semantic row covers the `marker.Current` path but not those generation branches. The two focused install tests above cover the default reader and its production moved-tag consumer directly, and their corresponding mutants are killed; no coverage is claimed for those readers from the cross-conformance row alone.

### Verification run directly on the final worktree

- `go test ./internal/marker ./internal/ui ./internal/registry ./internal/scopes -count=1` — exit 0.
- `go test ./internal/install -run '^(TestMarkerGenerationDistinguishesAbsentUnreadableAndInvalid|TestMovedTagReaderTreatsInvalidMarkerAsStaleButKeepsReadFailure)$' -count=1` — exit 0.
- `go test ./internal/crossconformance -run '^TestDraftExternalExecutionPolicyInvalidMarkerReDerives$' -count=1` — exit 0.
- `go test -race ./internal/crossconformance -run '^TestDraftExternalExecutionPolicyInvalidMarkerReDerives$' -count=1` — exit 0.
- `go test ./cmd/curator -run '^TestDraftStatusInvalidMarker$' -count=1` — exit 0.
- `go test ./cmd/curator -run '^TestCompiledProjectStatusAndUntrustedRecovery$' -count=1` — exit 0 (318.469 seconds).
- `go test ./internal/envprofile -run '^TestKnownAbsenceSensitiveReadersUseSharedSeam$' -count=1` — exit 0.
- `go vet ./internal/marker ./internal/install ./internal/crossconformance ./internal/scopes ./internal/registry ./internal/ui ./cmd/curator` — exit 0.
- `bash .github/ci/ledger-consistency.sh .temp/TASK-260907-2as5sx/ledger` — exit 0, 248 rows checked across Linux, macOS, and Windows.
- `bash .github/ci/gate-selftest.sh` — exit 0, 198 passed / 0 failed.
- `go build -o .temp/TASK-260907-2as5sx/curator ./cmd/curator` — exit 0.
- `make lint` — exit 0, 0 issues.
- `git diff --check` — exit 0.

### Bounded-run limitations and corrected intermediate failures

- `go test ./internal/crossconformance -run TestDraftSourcesSemanticCases -count=1` — exit 1 after the 10 minute test timeout; the last active row was `external-evidence-mismatch-substituted`. The full 94-row matrix under `-race` was not run. A filtered attempt at `TestDraftSourcesSemanticCases/external-evidence-mismatch-execution_policy` also exited 1 because the matrix deliberately rejects partial `-run` selections (1 of 94 executed). The separate focused production-entry test is the bounded check for this rework.
- The first final-package sweep of marker/UI/registry/scopes exited 1 because two GC tests expected the old combined unreadable-or-invalid wording and one fixture wrote malformed JSON while naming it unreadable. The expectations now check the typed invalid diagnostic, and the unreadable integration case uses a directory at the marker path; the full package sweep above then exited 0.
- Windows runtime execution was unavailable locally. Ledger consistency checks the three declarations against all three platform builds, and gate self-tests passed; the platform-specific mode-bit skip was not observed on a Windows host in this run.
- `LOGBOOK.md` remains unchanged under the previously recorded Story-worktree no-logbook boundary; this revision's findings and evidence are captured in this task-scoped result and its board resource.

## Revision 3 — preserve §2.3 ordering with a valid publication-failure fixture

### Decision and behavior

- Kept the production `Install` order. `installLocked` reads the existing profile's `source.json` before it resolves and audits a new candidate so it can decide whether the profile name is already installed. A regular file at the profile-directory path makes that read fail with `ENOTDIR`; the §8.4 seam must report that read failure rather than treating it as absence, and surfacing must not run before the audit gate.
- Changed `TestSurfacingEmittedDespitePublicationFailure` to create the obstructing regular file from the observing sink after it records the first surfacing row. The source probe now sees true absence; resolution and audit pass; `op.publish` then fails at the obstructed profile directory. The test still requires the row to have reached the sink, checks that the error names the obstructed publication path, and verifies no lock was published.
- This follows the §2.3 “Declaration visibility” clause in `docs/environment-config.md`: install/update prints declaration rows “after the audit gate passes and before the lock is published or any surface is (re-)materialized.” No production ordering change or assertion removal was needed.
- Formatted `internal/envregistry/envregistry.go` and `internal/envprofile/surfacing_order_test.go`.

### Verification rerun for Revision 3

- `gofmt -l cmd internal` — empty output, exit 0.
- `go test ./internal/envprofile -run 'Surfacing|Stateread|Absent|Unreadable'` — exit 0 (54.867s).
- `go test -race ./internal/envprofile -run 'Surfacing|Stateread|Absent|Unreadable'` — exit 0 (60.858s).
- `go test ./internal/envprofile -run '^TestKnownAbsenceSensitiveReadersUseSharedSeam$' -count=1` — exit 0 (0.674s).
- `go test ./internal/crossconformance -run TestDraftSourcesSemanticCases` — exit 0 (432.830s; all 94 semantic cases).
- `bash .github/ci/gate-selftest.sh` — exit 0 (198 passed, 0 failed).
- `bash .github/ci/ledger-consistency.sh .temp/TASK-260907-2as5sx/ledger` — exit 0 (248 rows checked across Linux, macOS, and Windows).
- `make lint` — exit 0 (0 issues).
- `go build -o /tmp/TASK-260907-2as5sx-curator ./cmd/curator` — exit 0.
- `git diff --check` — exit 0.

The no-logbook boundary recorded above still applies; this task-scoped outcome and its board attachment carry the Revision 3 decision and evidence.

## Revision 4 — declare every POSIX mode-bit unreadability fixture

### Platform handling

- Swept the Story-added and Story-changed tests for `Chmod`, mode `0o000`, and unreadability. Added a Windows platform-control skip before each POSIX mode-bit fixture that did not already have one, using the registered reason `platform-control: POSIX mode-bit unreadability`.
- The build-root test that prompted this revision now makes the `assets` parent mode `000` after proving the build root absent. Its failed metadata read is therefore a POSIX permission fixture; Windows records the declared platform-control skip. This removes the Windows-only `("", nil)` outcome caused by treating a file-valued path parent as absent.
- Added 28 ledger rows. The three existing marker/UI rows already had the required Windows guard and remain in the inventory. No shared DACL helper or `internal/godriver` file was changed.

### Test inventory and ledger rows

Each row below uses `\t` to show the literal tab separators in `.github/ci/platform-cases.tsv`. The three marker/UI rows at the end were already declared before Revision 4.

```tsv
cmd/curator	TestContextExposureDistinguishesAbsentAndUnreadableBuildRoots	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves a failed build-root metadata read
internal/adapters	TestStageProjectRefusesAnUnreadableAdapterLedger	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves adapter staging preserves ledger read failures
internal/contextlock	TestReadDistinguishesAbsentAndUnreadableLock	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves context lock reads preserve failures
internal/contextstore	TestExistsDistinguishesAbsentFromUnreadable	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves context entry metadata failures
internal/envprofile	TestCurrentPointerAbsenceAndUnreadabilityDiffer	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves current pointer read failures
internal/envprofile	TestScopedCurrentAbsenceAndUnreadabilityDiffer	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves scoped current read failures
internal/envprofile	TestScopedCurrentReadFailureDoesNotFallBackToMachineCurrent	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves scoped current failures do not fall back
internal/envprofile	TestProfileSourceAbsenceAndUnreadabilityDiffer	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves profile source failures
internal/envprofile	TestUnreadableDefaultProfileSourceDoesNotRecreateDefault	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves default source failures stop recreation
internal/envprofile	TestOverlayOwnerLockAbsenceAndUnreadabilityDiffer	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves lock read failures
internal/envprofile	TestImportAbsentAndUnreadableRootSurfaceDiffer/unreadable_is_a_typed_loss	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves import root read failures
internal/envprofile	TestImportAbsentAndUnreadableSkillsLedgerDiffer/unreadable_ledger_is_a_typed_loss	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves import ledger read failures
internal/envprofile	TestImportAbsentAndUnreadableSkillsSurfaceDiffer/unreadable_skills_directory_is_a_typed_loss	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves import skill inventory read failures
internal/envprofile	TestUpdatePathSnapshotUnreadableIsNotAbsent	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves path snapshot read failures
internal/envprofile	TestStatusScopeMarkerAbsenceAndUnreadabilityDiffer/unreadable_marker_is_unknown_with_a_typed_diagnostic	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves status marker read failures
internal/envprofile	TestStatusOrphanMarkerAbsenceAndUnreadabilityDiffer/unreadable_marker_is_reported,_not_treated_as_absence	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves orphan marker read failures
internal/envprofile	TestStatusUnreadableOrphanInventoryIsNotEmpty	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves orphan inventory read failures
internal/envprofile	TestStatusBackupInventoryAbsenceAndUnreadabilityDiffer/unreadable_backup_inventory_is_unknown_and_non-current	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves backup inventory read failures
internal/envprofile	TestRemovePurgeMarkerAbsenceAndUnreadabilityDiffer/unreadable_marker_refuses_before_purge_writes	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves purge marker preflight failures
internal/globalbins	TestRefreshAndStageRefuseAnUnreadableOwnershipLedger	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves refresh and staging preserve ledger read failures
internal/scopes	TestConsumerRegistryAbsenceAndUnreadabilityDiffer	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves consumer registry failures
internal/scopes	TestSweepRuntimeAbsenceAndUnreadabilityDiffer	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves runtime root inventory failures
internal/scopes	TestSweepRuntimeUnreadableSkillInventoryIsNotIgnored	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves runtime skill inventory failures
internal/scopes	TestHybridManifestAbsenceAndUnreadabilityDiffer	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves hybrid manifest failures
internal/scopes	TestCollectStaysFailSafeAcrossConsecutivePasses/unreadable_installed_skill_directory	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves GC retains state when a skill directory cannot be read
internal/scopes	TestCollectSkipsTheBuildSweepOnUnprovableReferences/unreadable_global_skills_directory	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves GC refuses an unprovable skill inventory
internal/stateread	TestReadFileDistinguishesAbsentAndUnreadable	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves typed file read failures
internal/stateread	TestReadDirDistinguishesAbsentAndUnreadable	linux,darwin	windows	platform-control	POSIX mode-bit unreadability proves typed directory listing failures
internal/marker	TestCurrentDistinguishesAbsentAndUnreadable	linux,darwin	windows	platform-control	POSIX mode-bit unreadability is the fixture used to prove Current preserves failed reads
internal/marker	TestReadStateDistinguishesAbsentAndUnreadable	linux,darwin	windows	platform-control	POSIX mode-bit unreadability is the fixture used to prove ReadState preserves failed reads
internal/ui	TestSkillsUnderDistinguishesAbsentAndUnreadableMarkers	linux,darwin	windows	platform-control	POSIX mode-bit unreadability is the fixture used to prove the skills view preserves failed reads
```

### Verification run for Revision 4

- Focused absence/unreadability tests passed: `cmd/curator` `TestContextExposureDistinguishesAbsentAndUnreadableBuildRoots`; `internal/adapters` `TestStageProjectRefusesAnUnreadableAdapterLedger`; `internal/globalbins` `TestRefreshAndStageRefuseAnUnreadableOwnershipLedger`; `internal/contextlock` `TestReadDistinguishesAbsentAndUnreadableLock`; `internal/contextstore` `TestExistsDistinguishesAbsentFromUnreadable`; `internal/stateread` `TestReadFileDistinguishesAbsentAndUnreadable` and `TestReadDirDistinguishesAbsentAndUnreadable`; the 15 targeted `internal/envprofile` tests listed by their test names in the sweep; the four new `internal/scopes` absence/unreadability tests and both affected GC test tables; `internal/marker` Current/ReadState tests; and `internal/ui` skills inventory test. All direct `go test` commands exited 0.
- Verbose subtest checks confirmed the exact ledger names for the changed import/status/purge subtests and both GC table subtests; those commands exited 0.
- `sh .github/ci/ledger-consistency.sh /tmp/TASK-260907-2as5sx-ledger-revision4` — exit 0, 276 rows checked across Linux, macOS, and Windows.
- `sh .github/ci/gate-selftest.sh` — exit 0, 198 passed / 0 failed.
- Windows test binaries cross-compiled with exit 0 for `./cmd/curator`, `./internal/adapters`, `./internal/contextlock`, `./internal/contextstore`, `./internal/envprofile`, `./internal/globalbins`, `./internal/marker`, `./internal/scopes`, `./internal/stateread`, and `./internal/ui` using `GOOS=windows go test -c`.
- `make lint` — exit 0, 0 issues. `go build -o /tmp/TASK-260907-2as5sx-curator-revision4 ./cmd/curator` — exit 0. `gofmt -l` on all touched Go files — empty output, exit 0. `git diff --check` — exit 0.
- The Windows test binaries were cross-compiled but not executed on a Windows host. The full repository suite was not rerun in this revision; focused affected-package tests and the requested ledger/self-test/build/lint checks were run.

## Revision 5 (cleanup + republish)

- Deleted `TASK-260907-2as5sx_results.md` from the Story worktree root: it is a board resource, not a repository file. Change nothing else.
- Proof of no other change: before deletion the worktree copy was byte-identical to the published board resource (sha256 `f4002230aa192be0c891fb2e9d84f158e0651ed59d845ebdd41c6c49f027a910`, 35030 bytes, `diff -q` exit 0). After deletion `git status --short` shows the same 52 modified tracked paths plus the same 7 untracked test paths as before, minus only the removed `?? TASK-260907-2as5sx_results.md` line; `git diff --check` exits 0. Per-file sha256 hashes of the 52 tracked and 8 untracked files were recorded in this run.
- Code, tests, and Revision 4 evidence are unchanged; this revision only republishes the identical report as the board outcome resource.

## Revision 6 — deny-by-default manager read scan

### Guard and coverage

- Replaced the fixed function-name list with `TestManagerOwnedAbsenceReadsAreGuarded`, an AST scan over every non-test `.go` file under `internal/` and `cmd/`, excluding `internal/stateread/`. A second independent directory walk checks the production-file denominator before the test accepts the scan.
- The scanner resolves import aliases, local error variables, and typed method selections. It recognizes `os` and `io/fs` read calls and method values, `os.IsNotExist`, `errors.Is(err, fs.ErrNotExist)`, and `os.ErrNotExist`; it follows local and imported calls transitively to `internal/stateread`.
- It reports the measured site ratio on every run: **173 guarded via seam / 117 allowlisted / 290 scanned**, across **367 production files**. Every exception is an exact `file:function` key with a single-line reason. The 117 reviewed entries concern caller-selected config/source inputs, source trees, native surfaces, path topology, caches, and temporary workspaces. Manager-owned persistent metadata was routed through `stateread`; broadening that seam over external inputs, cache lookups, and destination topology would make it own filesystem reads outside its state boundary, so those sites remain explicitly reviewed exceptions.
- Added `TestManagerReadScannerFindsAliasedNotExistCollapse`, using aliased imports and `errors.Is(err, fs.ErrNotExist)` to prove that syntax shape is recognized. The main guard is deny-by-default: any unreviewed absence-sensitive read fails the named test.
- Ran the reviewer’s `readNewManagerState` narrowing mutant in `/tmp/TASK-260907-2as5sx-mutant.ZTYCeW`, with the current scanner source copied into that disposable tree. `TestManagerOwnedAbsenceReadsAreGuarded` failed as expected (exit 1) and named `internal/envprofile/zz_state_read_mutant.go:readNewManagerState`, which performs `os.ReadFile` then maps both not-exist and other errors to `"default"`.

### Revision 6 verification

- `go test ./internal/envprofile -run '^(TestManagerOwnedAbsenceReadsAreGuarded|TestManagerReadScannerFindsAliasedNotExistCollapse)$' -count=1 -v` — exit 0; reported 173 / 117 / 290 across 367 files.
- `go test ./internal/envprofile -run '^(TestCurrentPointerAbsenceAndUnreadabilityDiffer|TestScopedCurrentAbsenceAndUnreadabilityDiffer|TestScopedCurrentReadFailureDoesNotFallBackToMachineCurrent|TestProfileSourceAbsenceAndUnreadabilityDiffer|TestUnreadableDefaultProfileSourceDoesNotRecreateDefault|TestOverlayOwnerLockAbsenceAndUnreadabilityDiffer|TestImportAbsentAndUnreadableRootSurfaceDiffer|TestImportAbsentAndUnreadableSkillsLedgerDiffer|TestImportAbsentAndUnreadableSkillsSurfaceDiffer|TestUpdatePathSnapshotUnreadableIsNotAbsent|TestStatusScopeMarkerAbsenceAndUnreadabilityDiffer|TestStatusOrphanMarkerAbsenceAndUnreadabilityDiffer|TestStatusUnreadableOrphanInventoryIsNotEmpty|TestStatusBackupInventoryAbsenceAndUnreadabilityDiffer|TestRemovePurgeMarkerAbsenceAndUnreadabilityDiffer)$' -count=1` — exit 0 (44.302s).
- Re-ran the five review-cycle sites verbosely: `TestImportAbsentAndUnreadableRootSurfaceDiffer`, `TestImportAbsentAndUnreadableSkillsLedgerDiffer`, `TestUpdatePathSnapshotUnreadableIsNotAbsent`, `TestStatusScopeMarkerAbsenceAndUnreadabilityDiffer`, `TestStatusOrphanMarkerAbsenceAndUnreadabilityDiffer`, and `TestRemovePurgeMarkerAbsenceAndUnreadabilityDiffer` — exit 0 (28.936s); absent and unreadable subtests both ran and passed, with no platform skips.
- `go test ./internal/hookapproval -count=1` — exit 0.
- `go test ./internal/registry -run '^(TestProtectedRollbackStateDistinguishesAbsentAndUnreadable|TestSnapshotRollbackStateCorruptionFailsClosedAndMigrates)$' -count=1` — exit 0.
- `go test ./internal/transaction -run '^TestJournalInventoryDistinguishesAbsentAndUnreadable$' -count=1` — exit 0.
- `go test ./internal/audit -run '^(TestCheckSourceAudit.*|TestSourceAudit.*|TestValidateSourceAudit.*|TestParseSourceAudit.*)$' -count=1` — exit 0.
- `go vet ./...` — exit 0 (rerun after the last scanner helper edit).
- `go build -o /tmp/TASK-260907-2as5sx-curator-build ./cmd/curator` — exit 0 (rerun after the last scanner helper edit).
- Repository gofmt check, `test -z "$(gofmt -l cmd internal)" || { echo 'gofmt: files need formatting:'; gofmt -l cmd internal; exit 1; }` — exit 0.
- `golangci-lint run` — exit 0, 0 issues on final run.
- `git diff --check` — exit 0.
- The full `go test ./...` suite was not run in this revision; the affected site tests and packages above were run directly.

### Resolved intermediate failures

- The first aliased-fixture run exited 1 because the isolated fixture also inherited the production allowlist and reported its entries stale. The scanner now accepts an explicit allowlist argument; the fixture passes an empty one and the final aliased-fixture run exits 0.
- An intermediate global scan exited 1 after conservative receiver-method detection classified `archive/zip.File.Open` as a filesystem read. The scan now checks Go's resolved method package. That exposed two `os.Root.Open` sites; typed `os` method handling was added, and the final scan exits 0 with all allowlist entries in use.
- The first `golangci-lint run` exited 1 because `readModulePath` ignored `file.Close`. It now checks the close result; the final lint run exits 0.
- Earlier incremental package test attempts during scanner assembly exited 1 for incomplete imports/helpers, including a missing `stateread` import in `switch.go`. Those compile errors were corrected before the final targeted test and validation runs above.

## Revision 7 — refresh onto R5 trunk and retention-test investigation

### macOS retention assertion

- The rev6 macOS failure was not reproduced. `newFakeDeps` injects `countingGeneration` into `TestAnInFlightTransactionKeepsThePublishedCacheEntry`, so the test never calls the production `markerGeneration` reader migrated to `marker.ReadState`. The test synchronously adds each published key to its stub journal; commit error handling sets `BuildCacheRetained` from the commit outcome, which checks failed publication state and whether an in-flight journal still references the published key. The state-read seam does not participate in that path.
- Exact test command on trunk `948ae7c9` (materialized from `git archive` under `/tmp/2as5sx-rev7`): `go test ./internal/install -run '^TestAnInFlightTransactionKeepsThePublishedCacheEntry$' -count=30` — exit 0, 22.702s.
- Same command on the pre-refresh candidate — exit 0, 23.609s.
- Same command after refresh onto `948ae7c9` — exit 0, 161.439s.
- The lone CI assertion failure remains unexplained; repeated base/candidate runs did not reproduce it, and the production seam is outside its injected read path. The evidence is consistent with a transient platform/run anomaly, but does not identify its cause. I left the test and retention production path unchanged.

### Refresh and newly discovered R5 reader

- Combined `git diff 1511b345 948ae7c9 -- . ':!.task-board' ':!CHANGELOG.md'` with a three-way application, retaining both sides of the `skip-classes.tsv` overlap. Then `task-board worktree refresh-candidate TASK-260907-2as5sx` — exit 0, `refresh_advanced`, trunk/reviewed trunk `948ae7c9e4a71a4026968913e1ff646aa21e0d52`, refreshed branch `adda8e2a982c3d0ff7cba4857fb1440e9b3bca03`.
- The first post-merge deny-by-default scan found `internal/runtimestore/enforced.go:ManagedEnforcedShimsIn`, introduced by R5, using `os.ReadDir` plus a missing-only fallback. It was migrated to `stateread.ReadDir`: proven absence still yields an empty inventory; read failure returns the typed `manager_state_unreadable` error. Added `TestManagedEnforcedShimsInDistinguishesAbsentAndUnreadable` and its POSIX/Windows platform-case row.
- A narrowing mutant changed the new reader's error return to `(nil, nil)`. `go test ./internal/runtimestore -run '^TestManagedEnforcedShimsInDistinguishesAbsentAndUnreadable$' -count=1` killed it with exit 1 and named the test. After restoring production code, the same test exited 0.
- The refreshed `.task-board` checkout copy was restored to refreshed `HEAD`; the authoritative board was accessed only through `task-board`. Final `git diff --name-only -- .task-board` was empty. `CHANGELOG.md` matches trunk `948ae7c9` byte-for-byte (`git diff 948ae7c9 --exit-code -- CHANGELOG.md` — exit 0); the release-prep entry is below.

### Guard and absence/unreadable evidence

- Final guard run: `go test ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$' -count=1 -v` — exit 0; measured **175 seam-guarded / 117 reviewed exceptions / 292 scanned reads**, across **403 production files**. The scan covers non-test Go under `internal/` and `cmd/`, excluding the shared `internal/stateread` implementation.
- Fresh narrowing mutant added `readNewManagerState`, which maps both `os.ReadFile` absence and other errors to `"default"`. The named guard failed as expected with exit 1, identified `internal/envprofile/zz_rev7_manager_read_mutant.go:readNewManagerState`, and reported 292/293 readers covered. The mutant file was removed; the guard then exited 0 again.
- Reran the focused envprofile absence/unreadability suite after refresh — exit 0, 27.992s. It includes all five review-cycle sites (`readRootSurface`, `readSkillsLedger`, KindPath snapshot, status scope marker, status orphan marker) and `purgeHomes`. Their individual narrowing-mutant evidence from Revision 2 remains attached above; those six site mutants were not re-applied in Revision 7. Revision 7 freshly re-proved the deny-by-default mutant and the added R5 reader mutant.
- `go test ./internal/install -run '^(TestMarkerGenerationDistinguishesAbsentUnreadableAndInvalid|TestMovedTagReaderTreatsInvalidMarkerAsStaleButKeepsReadFailure|TestInjectedClockAndGenerationReaderDriveInstallMarkers)$' -count=1` — exit 0, 6.376s.
- `go test ./internal/runtimestore/... -count=1` — exit 0, 8.328s; the site-specific absent/unreadable test also passed independently after mutant restoration.

### Bounded verification

- `go test ./internal/install/... -count=1` was started directly and interrupted with Ctrl-C after 8m35s (exit 1) before the main install test binary returned; no pass is claimed for this aggregate. The focused install tests and the required 30-run test above completed green. The full `internal/envprofile` package suite was not rerun: Revision 6 records that `go test ./internal/envprofile -count=1` exceeded the local 10-minute command bound and exited 143; this revision reran the bounded affected-test mask and guard instead.
- `gofmt -l cmd internal` — exit 0, empty output.
- `make lint` — exit 0, 0 issues. It emitted one non-blocking generated-file-filter warning for a file concurrently removed from another Story worktree.
- `bash .github/ci/ledger-consistency.sh /tmp/2as5sx-rev7/ledger` — exit 0, 387 rows checked across Linux, macOS, and Windows.
- `bash .github/ci/gate-selftest.sh` — exit 0, 198 passed / 0 failed.
- `go build -o /tmp/2as5sx-rev7/curator ./cmd/curator` — exit 0.
- Final `git diff --check`, `git diff --cached --quiet`, and `gofmt -l cmd internal` — exit 0. No changes are staged. Windows execution was not available on this host; its platform-control skip is registered and the ledger consistency check passed.

The campaign rule prohibits editing `LOGBOOK.md` from this Story worktree. Findings and anomalies are recorded here, satisfying the logbook citation boundary for this task.

## CHANGELOG entry (for release prep)

- Manager-owned absence-sensitive reads now use typed absent and unreadable outcomes through a shared filesystem seam. The deny-by-default reader audit measures guarded and reviewed sites and fails on an unreviewed absence collapse. Status, import, snapshot, purge, cleanup, registry, and runtime inventory readers preserve read failures instead of invoking absence fallbacks; newly refreshed enforced-shim inventory reads use the same seam (environments §8.4.1).


## Revision 8 — refresh onto a48f584c

### Refresh and merge

- Applied the prescribed `git diff 948ae7c9 a48f584c -- . ':!.task-board' ':!CHANGELOG.md' | git apply --3way` with `pipefail`. It exited 1 on the overlapping `internal/marker/marker.go` hunk after applying the other hunks. Kept revision 7's `stateread` API and the incoming schema-9 upper bound (`SkillSchemaVersion > 9` is rejected).
- Invoked `task-board worktree refresh-candidate TASK-260907-2as5sx`. The wrapper did not expose that invocation's process exit code; `task-board worktree status STORY-260906-1a2i5a` exited 0 and confirmed the refreshed tip `d9cb8465` is based on `a48f584c`. The generated `resolution-template.json` had no conflict entries. A replay invocation with that template returned exit 1, `base is already current`.
- The refresh left the incoming directory feature paths absent from the live worktree, so restored the non-overlapping incoming paths from refreshed `HEAD`, excluding `.task-board`, `CHANGELOG.md`, and `internal/marker/marker.go`. The marker file retains both the schema-9 validation and revision 7's stateread seam. `CHANGELOG.md` and the checkout copy of `.task-board` have no worktree delta.
- The incoming production reader `lockedNetworkRepository` used `os.Lstat` on the manager's SkillsRoot checkout list. Migrated it to `stateread.Lstat`, preserving the missing-checkout fallback and returning typed unreadable errors. Added `TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts`.
- The incoming `manifest.expand.rejectDirectorySymlinks` Lstat probes a pinned source snapshot rather than manager-owned state. It skips exact absence for later selection diagnostics and returns other errors; added its exact function key to the scanner allowlist with the source-snapshot rationale. The first post-refresh guard run exited 1 and identified this reader. The guard passed after the reviewed entry.

### Regression and narrowing mutant

- Narrowing mutant changed `lockedNetworkRepository`'s stateread error branch to `continue`, collapsing unreadable into absence. `go test ./internal/install -run '^TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts$' -count=1` exited 1 and failed the named `unreadable_checkout_stops_fallback` subtest because it fell through to the later valid checkout. Restored production code; the same command passed with exit 0.
- Final deny-by-default inventory: `go test ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$' -count=1 -v` — exit 0; **178 guarded via seam / 118 reviewed allowlisted / 296 scanned**, across **403 production files**.

### Revision 8 verification

All listed commands below were run directly and exited 0 unless marked expected-red.

- `go test ./internal/marker -count=1`
- `go test ./internal/envprofile -run '^(TestCurrentPointerAbsenceAndUnreadabilityDiffer|TestScopedCurrentAbsenceAndUnreadabilityDiffer|TestScopedCurrentReadFailureDoesNotFallBackToMachineCurrent|TestProfileSourceAbsenceAndUnreadabilityDiffer|TestUnreadableDefaultProfileSourceDoesNotRecreateDefault|TestOverlayOwnerLockAbsenceAndUnreadabilityDiffer|TestImportAbsentAndUnreadableRootSurfaceDiffer|TestImportAbsentAndUnreadableSkillsLedgerDiffer|TestImportAbsentAndUnreadableSkillsSurfaceDiffer|TestUpdatePathSnapshotUnreadableIsNotAbsent|TestStatusScopeMarkerAbsenceAndUnreadabilityDiffer|TestStatusOrphanMarkerAbsenceAndUnreadabilityDiffer|TestStatusUnreadableOrphanInventoryIsNotEmpty|TestStatusBackupInventoryAbsenceAndUnreadabilityDiffer|TestRemovePurgeMarkerAbsenceAndUnreadabilityDiffer)$' -count=1`
- `go test ./internal/install -run '^(TestMarkerGenerationDistinguishesAbsentUnreadableAndInvalid|TestMovedTagReaderTreatsInvalidMarkerAsStaleButKeepsReadFailure|TestInjectedClockAndGenerationReaderDriveInstallMarkers)$' -count=1`
- `go test ./internal/runtimestore/... -count=1`
- `go test ./internal/hookapproval -count=1`
- `go test ./internal/registry -run '^(TestProtectedRollbackStateDistinguishesAbsentAndUnreadable|TestSnapshotRollbackStateCorruptionFailsClosedAndMigrates)$' -count=1`
- `go test ./internal/transaction -run '^TestJournalInventoryDistinguishesAbsentAndUnreadable$' -count=1`
- `go test ./internal/audit -run '^(TestCheckSourceAudit.*|TestSourceAudit.*|TestValidateSourceAudit.*|TestParseSourceAudit.*)$' -count=1`
- `go test ./internal/contextlock ./internal/contextstore ./internal/ui ./internal/stateread -count=1`
- `go test ./internal/scopes -run '^(TestConsumerRegistryAbsenceAndUnreadabilityDiffer|TestSweepRuntimeAbsenceAndUnreadabilityDiffer|TestSweepRuntimeUnreadableSkillInventoryIsNotIgnored|TestHybridManifestAbsenceAndUnreadabilityDiffer|TestCollectStaysFailSafeAcrossConsecutivePasses|TestCollectSkipsTheBuildSweepOnUnprovableReferences)$' -count=1`
- `go test ./internal/adapters -run '^TestStageProjectRefusesAnUnreadableAdapterLedger$' -count=1`
- `go test ./internal/globalbins -run '^TestRefreshAndStageRefuseAnUnreadableOwnershipLedger$' -count=1`
- `go test ./cmd/curator -run '^TestContextExposureDistinguishesAbsentAndUnreadableBuildRoots$' -count=1`
- 1kpw4w directory/marker rows: `go test ./internal/closure -count=1`; `go test ./internal/skillspec ./internal/manifest ./internal/identifiers ./internal/marker -count=1`; and `go test ./internal/crossconformance -run '^TestDraftSourcesCLIDirectoryManifestDependencyInstallRefreshAndAudit$|^TestDraftSourcesCLILocalSkillScriptDependencies$' -count=1`.
- `go build -o /tmp/TASK-260907-2as5sx-curator-rev8 ./cmd/curator`; `go vet ./...`; `make lint` (0 issues); `gofmt -l cmd internal` (empty); `git diff --check`; `git diff --cached --quiet`; `git diff --quiet HEAD -- CHANGELOG.md`; `git diff --quiet HEAD -- .task-board`.

The full `go test ./...` suite was not run; this refresh followed the bounded package and focused-row instruction. No files were staged or committed.


## Revision 9 — Windows unreadable checkout

- Reproduced the Windows CI row failure source: the old fixture put a file at `blocked` but asked `Lstat` to read `blocked/child`. Windows can report that path as not-found, allowing the absence fallback. The fixture now puts the file at the exact locked checkout path (`blocked`), which is a present but unusable manager-owned checkout on all platforms.
- Production defect fixed: `lockedNetworkRepository` now wraps an existing checkout's failed repository validation in `stateread.UnusableError`, preserving the `manager_state_unreadable` typed outcome and stopping fallback. Proven absence still continues to the later checkout.
- Named regression: `TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts` has separate absent-permits-fallback and unreadable-stops-fallback rows, and asserts both the typed kind and stable unreadable diagnostic.
- Narrowing mutant: changed the `EnsureRepo` failure branch to `continue`, collapsing the invalid present checkout into absence. `go test ./internal/install -run '^TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts$' -count=1` failed with exit 1 in `unreadable_checkout_stops_fallback` because the later checkout was returned and the error was nil. Restored production code; the same command passed with exit 0.
- `GOOS=windows go vet ./internal/install` — exit 0. The Windows test binary was not executed on this host; this is cross-target vet evidence, while the changed test row ran natively.
- `gofmt -d internal/install/draftsources.go internal/install/draftsources_test.go` — exit 0, empty output. `git diff --check` — exit 0. The full package suite was not rerun; this revision followed the bounded changed-row instruction. No CHANGELOG edit was made.

### Revision 9 verification rerun for this handoff

- The refreshed worktree already contains the portable fixture: a regular file occupies the exact locked checkout path `skillsRoot/blocked`; the later checkout is a valid repository. The production `EnsureRepo` failure is wrapped as `stateread.UnusableError`, so fallback remains limited to proven absence. No persistent source edit was needed in this rework.
- Mutant recheck: changed the `EnsureRepo` validation failure branch in `lockedNetworkRepository` to `continue`. `go test ./internal/install -run '^TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts$' -count=1` exited 1; the named `unreadable_checkout_stops_fallback` subtest received `<nil>` instead of the typed error. Restored the candidate source byte-for-byte (`cmp` exit 0).
- Restored candidate: the same focused `go test` command exited 0.
- `GOOS=windows go vet ./internal/install` exited 0. Windows execution of the test row was not available in this local session; the portable row has no platform skip or ledger addition.
- `gofmt -d internal/install/draftsources.go internal/install/draftsources_test.go` exited 0 with empty output; `git diff --check` exited 0. No CHANGELOG edit was made.
