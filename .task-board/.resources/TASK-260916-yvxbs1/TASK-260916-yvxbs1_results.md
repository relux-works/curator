# TASK-260916-yvxbs1 — implementation results

## Scope and pinned rules

Implemented curator path-kind MCP refusal and path-directory boundary validation against curator-spec v1.0.0-rc.13, pinned at commit `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`.

The normative rules are in `protocol/environments.md`:

- §2.2 refuses MCP declarations carried by path-kind roots, overlays, and onboarding imports with `mcp_declaration_path_source_refused`, naming the package and declaration.
- §3 admits system modules from direct packages, including direct path roots/overlays, and applies the existing `transitive_system_modules` policy to transitive modules. Under `error`, a transitive module fails with `context_system_module_transitive`.
- §4 requires a path source directory to meet the protected-boundary contract at every resolve and before materialization. Ownership, private mutation permissions/DACL, containment, regular-file type, and link safety failures use `environment_store_untrusted`; resolve emits no fragment, status is non-current, and the manager does not rebuild the operator-owned path source.
- The pinned vectors are `conformance/v1/vectors/environments-path-kind-admission.json`.

`yvxbs1-decision-1.md` supersedes the task description’s blanket system-module refusal: rc.13 §3 governs. A trusted direct path overlay’s system module is admitted, while a transitive one is refused through the existing admission. `internal/pathboundary` is the shared protected-boundary helper for the later S5 leaf; production supplies the platform owner lookup and the validator exposes an injected owner lookup for production-entry tests.

## Rules → rows → vectors

| Rule / production entry | Pinned vector rows | Result |
|---|---|---:|
| MCP source admission through `Install` (including path root, onboarding import, and overlay) | `git-mcp-declaration-admitted`; `path-root-mcp-declaration-refused`; `path-import-mcp-declaration-refused`; `path-overlay-mcp-declaration-refused`; `path-mcp-declaration-admitted` (published non-conforming admission, now refused) | 5/5 driven |
| §4 protected boundary through `Install`, `Import`, `Resolve`, and `Status` | `path-overlay-world-writable-untrusted`; `path-overlay-no-system-world-writable-untrusted`; `path-overlay-symlinked-component-untrusted`; `path-overlay-containment-escape-untrusted`; `path-overlay-non-regular-component-untrusted`; `path-overlay-wrong-ownership-untrusted`; `path-import-no-system-wrong-ownership-untrusted`; `path-overlay-untrusted-reported-current`; `path-overlay-untrusted-rebuilds` | 14/14 driven (includes the rows below) |
| §3 direct/transitive system-module admission through production `Install` and `Resolve` | `path-overlay-system-module-admitted`; `path-transitive-system-module-refused` | included in 14/14 |
| Path root/import/overlay normal admission | `path-root-no-system-modules-admitted`; `path-import-no-system-modules-admitted`; `path-overlay-no-system-modules-admitted` | included in 14/14 |
| Read-only resolve behavior | `path-overlay-dry-run-untrusted-no-rebuild`; `path-overlay-dry-run-intact-plans-nothing`; `path-overlay-dry-run-reports-would-rebuild` | 3/3 driven |

All vector families reported zero known gaps, bounds, or skips. The boundary rows drive production entries and assert the required diagnostics and named failing checks. Ownership rows use injected foreign identities through the boundary validator; production uses the real platform owner lookup. A direct trusted system overlay emits its system-prompt bytes; a transitive system module fails install through the existing admission.

## Conformance-gap counts

The assigned checkpoint eca2bf27 had 70 actual TSV data rows (comments and the column header excluded), with 0 rows owned by STORY-260916-wgt8vz or TASK-260916-yvxbs1. Current origin/main 6bd98d49 has 68 data rows and 0 Story/task-owned rows; the candidate also has 68 and 0. The relevant before/after counts are 68→68 total rows against the current main baseline and 0→0 owned rows. No E6-owned passing gap row remained to remove, and the candidate has no conformance-gaps.tsv delta versus origin/main.

Two rows removed between the older checkpoint and current main are Codex seed rows owned by STORY-260916-1i1gfo; they are outside this task's ledger ownership. The earlier 72/70 tally counted two non-data TSV lines and is superseded here by the explicit data-row count.

## Mutation evidence

Each valid mutant was applied temporarily, its target vector failed with exit 1, and the modified file was restored and byte-compared with its pre-mutation copy. The unmutated pinned suites passed with exit 0 before and after these probes.

| Rule | Mutant | Production-entry test result |
|---|---|---|
| §2.2 MCP refusal | Replaced `refusePathMCPDeclaration` with a no-op | Full MCP vector family failed, exit 1; the path root/import/overlay declarations were admitted. |
| §4 protected permissions | Removed the mutation-permission/DACL check from the boundary walk | Full boundary vector family failed, exit 1, on world-writable path rows. |
| §4 ownership | Bypassed the owner identity equality check | Full boundary vector family failed, exit 1, on wrong-owner overlay/import and untrusted-rebuild rows. |
| §3 direct admission | Removed the direct-package admission branch | Full boundary vector family failed, exit 1; trusted direct overlay emitted no system-prompt section. |
| §3 transitive refusal | Softened the `TransitiveError` check to return without refusal | Full boundary vector family failed, exit 1; the transitive system module was admitted by install. |

The MCP mutant command was `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envprofile -run '^TestPathKindMCPVectorsDriveInstallEntry$' -count=1` — exit 1. Each of the four §3/§4 boundary mutants used `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envprofile -run '^TestPathKindBoundaryVectorsDriveProductionEntries$' -count=1` — exit 1.

The final mutant probes were buildable and killed by behavior assertions. Two initial mutation drafts were discarded because they did not compile (owner variable unused; direct-set variable unused); they are not counted as kills. Earlier filtered-subtest probes also exited 1 with the coverage helper’s filtered-family warning; the table records the full-family runs.

## Local validation and real exit codes

Each command below ran directly as a standalone process. No hosted-green claim is inferred from local runs.

- `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envprofile -run '^(TestPathKindMCPVectorsDriveInstallEntry|TestPathKindBoundaryVectorsDriveProductionEntries|TestPathKindDryRunVectorsDriveReadOnlyResolve)$' -count=1 -v` — exit 0; MCP 5/5, boundary 14/14, dry-run 3/3.
- `go test ./internal/envprofile -run 'Resolve' -count=1` — exit 0.
- `go test ./internal/envprofile -run 'Status' -count=1` — exit 0.
- `go test ./internal/envprofile -run 'Path' -count=1` — exit 0.
- `go test ./internal/envprofile -run 'Overlay' -count=1` — exit 0.
- `go test ./internal/envprofile -run 'Boundary' -count=1` — exit 0.
- `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envprofile -run 'CodexSeed|EnvironmentsCodexSeed' -count=1` — exit 0.
- `go test ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$' -count=1` — exit 0.
- `go test ./internal/envprofile -run '^(TestPathOverlayBoundaryRejectsEscapingAndInternalLinks|TestPathOverlayTrustedSystemModuleRemainsAdmitted|TestPathOverlayFailureBlocksResolveAndMakesStatusNonCurrent|TestPathSourceMutationsRejectUntrustedOverlayBeforeDefaultWrites|TestUpdateDefaultIsBlocked)$' -count=1` — exit 0.
- `go test ./cmd/curator -run '^(TestProfileInstallRefusesPathMCPDeclarationThroughCLI|TestProfileListMigrationHonoursSystemPolicy/update-default)$' -count=1` — exit 0.
- `go test ./internal/pathboundary ./internal/contextpkg ./internal/contextresolve ./internal/contextmaterialize -count=1` — exit 0.
- `golangci-lint run` — exit 0, 0 issues.
- `go vet ./internal/pathboundary ./internal/contextpkg ./internal/contextresolve ./internal/contextmaterialize ./internal/envprofile ./cmd/curator` — exit 0.
- `go build -o "$TMPDIR/TASK-260916-yvxbs1-curator" ./cmd/curator` — exit 0.
- `GOOS=windows GOARCH=amd64 go test -c ./internal/pathboundary -o "$TMPDIR/TASK-260916-yvxbs1-pathboundary.test.exe"` — exit 0.
- `GOOS=windows GOARCH=amd64 go test -c ./internal/envprofile -o "$TMPDIR/TASK-260916-yvxbs1-envprofile.test.exe"` — exit 0. These are compile checks, not Windows runtime evidence.
- `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 bash .github/ci/ledger-consistency.sh "$TMPDIR/TASK-260916-yvxbs1-ledger-final"` — exit 0; 465 rows checked across linux, darwin, and windows.
- `git diff --check` — exit 0. Compared with current `origin/main`, the changed-path list contains only task source/test paths and the shared CI ledgers.

## Prior hosted failures and current gate

The attached revision-2 validation log (run 36328835395) is red. The macOS and Ubuntu test/race lanes report the default-profile update-order failures; Windows reports that its world-writable fixture did not create a boundary and that TestPathOverlayFailureBlocksResolveAndMakesStatusNonCurrent returned environment_home_stale instead of environment_store_untrusted. This prior log is evidence of a failed revision, not green proof.

The current candidate now validates path sources in Resolve before home-currency checks, has Windows helpers that add a real Everyone-write DACL for the permission vector, and preserves default-profile update refusal ordering. On this host, TestPathOverlayFailureBlocksResolveAndMakesStatusNonCurrent, TestValidateRejectsGroupOrWorldWritableComponents, TestUpdateDefaultIsBlocked, and the CLI update-default subtest all pass. Windows code/tests cross-compile, but Windows runtime behavior remains unverified locally.

A new handoff must publish the current candidate and run the hosted gate. Read its attached validation log and accept only an all-green result; the red revision-2 result remains part of this history.

## Worktree and release prep

No `CHANGELOG.md` or `LOGBOOK.md` was edited. The worktree remains uncommitted. The user-specified combine check was run against latest `origin/main`; the candidate differs there only in this leaf’s files and CI ledgers.

## CHANGELOG entry (for release prep)

- Refuse MCP declarations from path-kind packages and validate path source directories against the protected-boundary contract, including ownership, permissions, symlink, and containment checks.

## Revalidation performed on 2026-09-27

The pinned spec checkout was verified at 23435129ebc4c29e5b7f75ec72a0aa0cd3f16065. The three separate production-entry runs on this candidate passed: MCP 5/5, path boundary 14/14, and read-only resolve 3/3; each reported 0 known gaps, 0 bounds, and 0 skips. The owner, symlink, escape, and permission rows were all driven on this macOS host.

Additional standalone commands passed with exit 0:

- go test ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$' -count=1
- go test ./internal/pathboundary ./internal/contextpkg ./internal/contextresolve ./internal/contextmaterialize -count=1
- go test ./internal/envprofile -run '^(TestPathRootAndImportedMCPDeclarationsRefused|TestPathOverlayMCPDeclarationRefused|TestPathOverlayBoundaryRejectsEscapingAndInternalLinks|TestPathOverlayFailureBlocksResolveAndMakesStatusNonCurrent|TestPathSourceMutationsRejectUntrustedOverlayBeforeDefaultWrites)$' -count=1
- go test ./internal/envprofile -run '^TestUpdateDefaultIsBlocked$' -count=1
- go test ./cmd/curator -run 'TestProfileListMigrationHonoursSystemPolicy/update-default' -count=1
- go test ./internal/pathboundary -run '^TestValidateRejectsGroupOrWorldWritableComponents$' -count=1
- go test ./internal/envprofile -run '^TestPathOverlayFailureBlocksResolveAndMakesStatusNonCurrent$' -count=1
- GOMAXPROCS=2 make lint
- go build -o "$TMPDIR/TASK-260916-yvxbs1-curator" ./cmd/curator
- GOMAXPROCS=2 GOOS=windows GOARCH=amd64 go test -c -o "$TMPDIR/TASK-260916-yvxbs1-envprofile.test.exe" ./internal/envprofile
- gofmt -d over changed Go sources/tests; git diff --check origin/main -- . ':!.task-board'

The build, lint, test, cross-compile, formatting, and diff-check commands all exited 0. The Windows command was compile-only, not Windows runtime evidence. The new outcome does not claim hosted-green status before the handoff gate runs.

The origin/main baseline selection for the three E6 vector test names exited 0 with “no tests to run”, confirming that those consumers were absent there; this is not counted as a passing behavior test. In this candidate, six narrowed mutants were killed by production-entry tests (each expected-red go test exited 1): make refusePathMCPDeclaration a no-op; bypass checkAdmissionPrePublish; accept foreign owners; ignore symlinks; suppress escaping-target classification; and disable group/world-write rejection. Each temporary mutation was restored. The final candidate passed the three vector families after restoration.

The fetched-base comparison shows only the E6 tracked delta relative to origin/main. The task checklist is checked; q confirmed all ten items done. No changelog or logbook file was edited.

## Checklist disposition

The task checklist items are checked on the board. This result resource records the rc.13 decision for the system-module AC, vector coverage, before/after gap counts, mutation results, validation commands, and prior hosted-gate findings. The task is ready for developer handoff; hosted validation remains the handoff gate.

## Revision 3 hosted gate correction (2026-09-27)

Revision 3 handoff published candidate `79e6a8ced9a3d818b25211083a70b2f440544342`; hosted run 36337460311 failed only the Windows test lane. The `Action=="fail"` rows in Windows `go-test.json` name `internal/envprofile::TestInstallSurfacesSystemModuleWarning` and `cmd/curator::TestProfileInstallWarnsOnSystemModule`. Their fixtures used `t.TempDir()` path packages without a private Windows DACL, so the protected-boundary validator correctly rejected the fixtures before the tests reached their expected warning assertions. This was a test-fixture trust setup issue; product behavior was not weakened.

Both fixtures now call the existing `pathboundary.ProtectTree` helper after writing the package files. The directly rerun commands all exited 0:

- `go test ./internal/envprofile -run '^TestInstallSurfacesSystemModuleWarning$' -count=1`
- `go test ./cmd/curator -run '^TestProfileInstallWarnsOnSystemModule$' -count=1`
- `gofmt -d internal/envprofile/envprofile_test.go cmd/curator/profile_test.go`
- `go build -o "$TMPDIR/TASK-260916-yvxbs1-curator" ./cmd/curator`
- `GOMAXPROCS=2 make lint` (0 issues)
- `GOMAXPROCS=2 GOOS=windows GOARCH=amd64 go test -c -o "$TMPDIR/TASK-260916-yvxbs1-envprofile.test.exe" ./internal/envprofile`
- `GOMAXPROCS=2 GOOS=windows GOARCH=amd64 go test -c -o "$TMPDIR/TASK-260916-yvxbs1-curator.test.exe" ./cmd/curator`

The two Windows commands are compile checks only. Revision 3 remains recorded as a failed hosted gate; the corrected candidate requires a new handoff gate, whose attached validation log will determine hosted status.
## Revision 5 — re-applied on `86552087`

### Rebase and behavior

Fetched `origin main` and applied `refs/campaign/wgt8vz-rev4-20260928` onto trunk `86552087f9a5ba27a9148ee84f579cc8b0b1fcbe`. The five expected conflicts were resolved by keeping trunk’s E4/E3 and §8.4.1 state while retaining E6. In particular, update/status/seed reads retain the `stateread` seams; the builtin `default` update refusal still precedes the source-record read; and path-source boundary checks remain at the relevant production entries.

Review of the §4 call sites found that standalone `curator gc` acquired the manager-home lock but did not preflight current path sources. Added `envprofile.PreflightCurrentPathSources` and call it from `collectUnderLock` before transaction recovery or sweeping. `TestGCFailsClosedForUntrustedCurrentPathSource` removes an installed current path source, then verifies GC returns `environment_store_untrusted` naming `regular_types` and leaves a dead consumer registered. The regression test is in `platform-cases.tsv` for linux, darwin, and windows.

Compared with `origin/main`, the candidate changes only this leaf’s source/test files and the three CI ledgers. `conformance-gaps.tsv` has no candidate delta. The baseline and candidate each have 68 data rows and 0 rows owned by STORY-260916-wgt8vz or TASK-260916-yvxbs1 (68→68 total; 0→0 owned); there were no owned passing gap rows to remove. `root-artifacts.tsv` retains trunk’s Codex-seed and read-failure rows and adds the path-kind vector family. No CHANGELOG.md or LOGBOOK.md edit.

### rc.13 rules and production-entry rows

Spec checkout pin: `curator-spec` tag `v1.0.0-rc.13`, commit `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`.

- §2.2: path-kind roots, overlays, and onboarding imports carrying an MCP declaration are refused with `mcp_declaration_path_source_refused`, naming package and declaration.
- §3: per `yvxbs1-decision-1.md`, directly named path roots/overlays admit system modules after §4 boundary validation; transitive modules use existing admission and refuse under `transitive_system_modules=error` with `context_system_module_transitive`.
- §4: path-source directories use the ownership, permissions/DACL, containment, regular-type, and lstat link-safety checks at resolve and under the manager-home lock for mutating operations. A failed boundary is `environment_store_untrusted`; resolve emits no fragment, status is non-current, and source directories are never rebuilt. `internal/pathboundary` is the shared helper intended for the later S5 leaf; production uses real owner lookup and tests inject owner identity.

| Rule / production entry | rc.13 rows | Current result |
|---|---|---:|
| §2.2 MCP admission through Install/Import and overlay installation | `mcp_kind_cases` | 5/5 driven; 0 gap, bound, or skipped |
| §4 path-boundary / §3 system-admission vector rows through Install, Import, Resolve, repair, and Status | `path_boundary_cases` | 14/14 driven; 0 gap, bound, or skipped |
| §4 dry-run resolve | `dry_run_cases` | 3/3 driven; 0 gap, bound, or skipped |

The path-boundary family includes the trusted direct overlay system-module admission and the transitive system-module refusal. It also drives permission/DACL, symlink, wrong-owner, containment escape, special-file, no-fragment/non-current, and no-rebuild observations. Ownership vectors use the injected owner seam; production uses the real platform lstat owner lookup. The production boundary call sites are Install/Import, Update, Use, Sync, Resolve, repair-under-lock, Status, and the newly guarded GC path. Update/Use/Sync and GC have focused non-vector tests; the `path-overlay-untrusted-rebuilds` vector exercises repair Resolve.

### Revision 5 validation and exit codes

Passed with exit 0:

- `go test ./internal/pathboundary`.
- `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envprofile -run '^(TestPathKindMCPVectorsDriveInstallEntry|TestPathKindBoundaryVectorsDriveProductionEntries|TestPathKindDryRunVectorsDriveReadOnlyResolve)$' -count=1 -v` — 5/5 MCP, 14/14 boundary/system, 3/3 dry-run.
- `go test ./internal/envprofile -run 'Path' -count=1`; `-run 'Boundary|Overlay'`; `-run 'MCP|Mcp|System'`; and `-run 'Update|Resolve|Status|Guarded'` — each exit 0.
- `go test ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$' -count=1`.
- CLI tests `TestProfileInstallRefusesPathMCPDeclarationThroughCLI`, `TestProfileInstallWarnsOnSystemModule`, and `TestProfileListMigrationHonoursSystemPolicy/update-default`, each individually.
- `go test ./cmd/curator -run '^TestGC(PrunesDeadConsumersUnderTheHomeLock|FailsClosedForUntrustedCurrentPathSource)$' -count=1`; the new GC test also passed after restoring the mutant.
- `go build ./...`; `GOMAXPROCS=2 make lint` (0 issues).
- `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 bash .github/ci/ledger-consistency.sh "$TMPDIR/TASK-260916-yvxbs1-ledger-r5-final"` — 467 rows checked across linux/darwin/windows.
- Windows cross-compiles for `./internal/envprofile` and `./cmd/curator` with `GOOS=windows GOARCH=amd64` — exit 0; these are compile checks, not Windows runtime evidence.
- `gofmt -d` across changed Go sources and `git diff --check origin/main -- . ':!.task-board'` — exit 0.

Two non-passing local attempts are recorded accurately:

- `go test ./cmd/curator -run 'Profile|Path|Status' -count=1` exited 1 after Go’s 10-minute timeout in `TestProfileInstallActivationMatrix/collection/same-source-unchanged-lock/neither`, blocked in a remote Git fetch. The narrowed leaf CLI tests above pass; no E6 assertion was reported by this broad timeout.
- The first run of `go test ./cmd/curator -run '^TestGCFailsClosedForUntrustedCurrentPathSource$' -count=1` exited 1 at compile time because the existing test helper had not yet passed the new policy argument. Updated the helper and reran the GC tests successfully.

### Mutation evidence and handoff

The prior revision-4 mutation matrix remains attached above: MCP refusal removed; protected-permission and owner-equality checks bypassed; direct system admission removed; and transitive refusal softened. Those production-entry tests exited 1 under their mutants. The prior baseline check reported no tests to run for the E6 vector consumer names on `origin/main` (exit 0, not counted as a behavior pass). For the new GC call site, removing `PreflightCurrentPathSources` made the new CLI regression test fail as expected (exit 1: GC returned 0); restored `main.go` byte-for-byte and reran the test with exit 0.

The hosted gate has not run yet; the handoff runner publishes this revision and runs it after this turn. Windows runtime behavior remains unverified locally.

## Revision 6 — re-applied on trunk `97e85642`

### Re-apply and review response

Re-applied the accepted revision-5 delta from `refs/campaign/wgt8vz-rev5-20260928` onto trunk `97e856425b1aeafe533e86e33e7f9dfd873d9b50` with `git apply --3way`. The only conflict was `internal/envprofile/envprofile.go`; it was resolved retaining trunk's E1 update/delta and confirmation flow, the path-source preflight, and `profile update default`'s refusal ordering. `fetchLockedSources(manager, oldLock)` and the E1 confirmation tests remain present. Against `origin/main`, the candidate has the 32 re-applied E6 paths plus the new regression test; no trunk-only paths were changed.

Read the pinned source directly from curator-spec tag `v1.0.0-rc.13` at `23435129`: `protocol/environments.md` §2.2 (path-root, overlay, and import MCP refusal), §3 (direct system-module admission and existing transitive admission), and §4 (five protected-boundary checks at every resolve and under the mutation lock for install/update/use/sync/repair/GC). The pinned vector file is `conformance/v1/vectors/environments-path-kind-admission.json`; the local corpus matches the tag's corpus. `internal/pathboundary` remains the shared S5 boundary helper.

| Rule and production entry | rc.13 vector rows | Revision-6 result |
|---|---|---:|
| §2.2 MCP refusal through install/import and overlay install | MCP family | 5/5 driven |
| §4 protected boundary, including §3 direct/transitive system modules through install/import/resolve/status | boundary family | 14/14 driven |
| §4 read-only resolve behavior | dry-run family | 3/3 driven |

The vector consumers report no known gaps, bounds, or skips. Boundary call sites remain install/import, update, use, sync, resolve, repair-under-lock, status, and standalone GC under the manager-home lock. A trusted direct path overlay's system module is admitted; the transitive system module is refused by the existing admission.

### Gap ledger and narrowing mutant

Compared with revision-6 base `origin/main` (`97e85642`), `.github/ci/conformance-gaps.tsv` has 12 data rows before and 12 after; rows owned by `STORY-260916-wgt8vz` or `TASK-260916-yvxbs1` are 0 before and 0 after. There were no Story-owned passing gaps to remove. The file has no candidate delta.

The rev5 review's M3 narrowing residual is closed by `TestResolveRejectsGroupWritableOnlyPathOverlay`, a production-entry Resolve regression that changes only the source overlay's group-write bit and requires `environment_store_untrusted` with the permissions check.

| Probe | Real exit | Outcome |
|---|---:|---|
| With `owner_unix.go` mutated from mask `0o022` to `0o002`, existing `go test ./internal/envprofile -run '^TestPathOverlayFailureBlocksResolveAndMakesStatusNonCurrent$' -count=1` | 0 | Mutant survived the old world-writable-only row. |
| Unmutated `go test ./internal/envprofile -run '^TestResolveRejectsGroupWritableOnlyPathOverlay$' -count=1` | 0 | New regression passes. |
| Same new test with the `0o022`→`0o002` mutant | 1 | Mutant killed: Resolve returned `environment_home_stale` instead of the required `environment_store_untrusted` permissions refusal. |
| Restored production check; focused boundary group below | 0 | Regression passes with the `0o022` check restored. |

The other rev5 mutation results remain in the revision-5 section above; they were not re-mutated during this re-apply. The rev5 review's Windows-only DACL narrowing and its non-blocking policy/overlay residuals were not changed in this revision; Windows runtime validation remains for the hosted lane.

### Revision-6 local validation

Each command ran as a standalone process and exited 0:

- `go test ./internal/pathboundary -count=1`.
- `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envprofile -run '^(TestPathKindMCPVectorsDriveInstallEntry|TestPathKindBoundaryVectorsDriveProductionEntries|TestPathKindDryRunVectorsDriveReadOnlyResolve)$' -count=1`.
- `go test ./internal/envprofile -run '^(TestResolveRejectsGroupWritableOnlyPathOverlay|TestPathOverlayBoundaryRejectsEscapingAndInternalLinks|TestPathOverlayTrustedSystemModuleRemainsAdmitted|TestPathOverlayFailureBlocksResolveAndMakesStatusNonCurrent|TestPathSourceMutationsRejectUntrustedOverlayBeforeDefaultWrites|TestUpdateDefaultIsBlocked|TestManagerOwnedAbsenceReadsAreGuarded)$' -count=1`.
- `go test ./internal/contextpkg ./internal/contextresolve -count=1`.
- `go test ./cmd/curator -run 'TestProfileInstallRefusesPathMCPDeclarationThroughCLI|TestProfileInstallWarnsOnSystemModule|TestProfileUpdateSystemDeltaConfirmationGolden|TestProfileUpdateMCPDeltaConfirmationGolden|TestProfileUpdateAllDeltaVectorsAtCLI' -count=1` (193.308s).
- `go test ./cmd/curator -run '^TestProfileListMigrationHonoursSystemPolicy/update-default$' -count=1`.
- `go build ./...`.
- `GOMAXPROCS=2 make lint` (0 issues).
- `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 bash .github/ci/ledger-consistency.sh "$TMPDIR/TASK-260916-yvxbs1-ledger-r6-final"` (468 rows checked across Linux, Darwin, and Windows).
- `gofmt -d internal/envprofile/envprofile.go internal/envprofile/path_source_group_write_unix_test.go` (no output).
- `git diff --check origin/main -- . ':!.task-board'`.

The local run did not execute Windows tests; the ledger check verified their platform compilation/coverage accounting, and the handoff runner will run the hosted gate after this turn. No CHANGELOG or LOGBOOK file was edited. This revision adds a task-scoped group-write finding to these results instead of editing LOGBOOK.

## Revision 7 (carry-forward republish)

Revision 6 was ACCEPTED on content (`TASK-260916-yvxbs1_review-verdict-rev6.md`). Trunk moved to `e4f4fe86`, so the orchestrator converged the accepted delta uncommitted into the Story worktree. This revision changes nothing; it verifies the carry-forward per `yvxbs1-carry-7.md` and republishes.

### Per-path verification (rev6 patch = 33 paths)

Method: materialized rev6 base `97e85642` for all 33 patch paths in `/tmp`, applied `TASK-260916-yvxbs1_change-request_rev6.patch` with `patch -p1` (rc=0), then compared against the worktree. Trunk `97e85642..e4f4fe86` (excl `.task-board`) touched 5 of the 33 paths (intersecting): `.github/ci/conformance-case-counts.tsv`, `.github/ci/root-artifacts.tsv`, `cmd/curator/main.go`, `internal/envprofile/managed.go`, `internal/envprofile/switch.go`.

- 28 non-intersecting paths: byte-identical (`cmp`) to base+rev6, including all 16 new files (`internal/envprofile/path_kind_*`, `path_source_*`, `pathsource.go`, all 8 `internal/pathboundary/*`). Nothing dropped.
- Intersecting paths, both sides present, no conflict markers (`<<<<<<<`/`>>>>>>>` absent in all 5):
  - `conformance-case-counts.tsv`: rev6 path-kind-admission rows (26-28) + trunk `environments-write-nofollow/cases 11` row; worktree-vs-base+rev6 diff is exactly that one trunk row.
  - `root-artifacts.tsv`: trunk write-nofollow row preserved; rev6 path-kind-admission row added.
  - `cmd/curator/main.go`: rev6 hunks content-identical (line-shifted by trunk's `cmdGlobalAdopt`); trunk global-adopt + rev6 `PreflightCurrentPathSources` both present.
  - `internal/envprofile/managed.go`: rev6's 3 `validateProfilePathSources` hunks present on trunk's nofollow-refactored code (`storeRoot`, `managedRelative`, `atomicManagedFile`, `DiagWriteWouldFollowLink` all present).
  - `internal/envprofile/switch.go`: rev6's `useLocked` preflight + `SyncWithPolicy` preflight + `materializeScope` validate hunks present; worktree-vs-HEAD diff is purely additive (+22 lines).

### CHANGELOG policy

`CHANGELOG.md` is absent from the rev6 patch and unmodified in the worktree (equals trunk) — no hunk to revert. The entry text remains verbatim in the "## CHANGELOG entry (for release prep)" section above for the release-prep leaf. No stray root `TASK-*`/`BUG-*` files; no `test/` or `ledger/` paths.

### Worktree freshness (step 4)

`git diff --name-only HEAD -- . ':!.task-board'` lists exactly the 17 tracked rev6 paths (comm-verified: no extras, no missing); untracked entries are exactly the 16 new rev6 paths. No trunk revert present.

### Revision-7 validation (standalone processes, real exit codes)

- `go test ./internal/envprofile -run 'Path|Boundary|Nofollow|Update|Guarded' -count=1` → ok 109s, exit 0 (includes trunk's Nofollow rows on the merged tree).
- `go build ./...` → exit 0.

Checklist: all 14 items were already checked; none left unchecked. No code changed in this revision, so no new tests were required.
