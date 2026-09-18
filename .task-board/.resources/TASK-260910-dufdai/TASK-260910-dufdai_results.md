# TASK-260910-dufdai — source-aware build receipts and cache (receipt v3), handoff evidence

Candidate: uncommitted working tree of `.temp/STORY-260910-20sx61/worktree` on top of the Story checkpoint `832facf` (17ps6u rev4 / hwxr26 rev10 both checkpointed). Shell: zsh, `set -o pipefail`; every gate below ran as a standalone `go test` / `go vet` / `golangci-lint` process redirected to a log, exit codes quoted from `$?`. No commits on the Story branch.

## What it implements (skillfile-sources.md protocol §4 "Build receipt schema 3", repository-transport §"external build receipt inputs")

Receipt schema 3 wraps the existing closed driver input as `input:{schema_version:3,package,build}`; `cache_key` is SHA-256 over the CCJ-1 bytes of that whole input; receipt/cache hashes are recomputed over the wrapper. It is selected only by a build input that carries a frozen package identity; every package-less input keeps the schema-1 (local) / receipt-2 (external) bytes, keys and namespaces byte-identical.

- `internal/buildmeta/package.go` (new): closed source-types-v1 `Package`/`PackageCommit` model with disjoint arms, `Validate`, canonical `Object()` and a closed raw-shape parser (exact field set per arm, string-typed members, commit format/length, canonical repository, portable directory). Test pins its canonical bytes equal to `sourcelock.Package.Canonical()` for all three arms. buildmeta stays a leaf: importing sourcelock created a test-only import cycle (sourcelock tests → manifest → buildrepo → buildmeta), so the receipt model owns its copy and `internal/install` translates lock → receipt identity losslessly (`receiptPackage`).
- `internal/buildmeta/models.go`, `codec.go`: `Input.Package *Package` (nil = legacy), `SourceAwareSchemaVersion = 3`, `Input.ReceiptSchemaVersion()/SourceAware()`, `logicalValue` (wrapper vs legacy value), receipts minted with the version their input selects, `Receipt.Validate` binds the receipt version to the input shape, `DecodeReceipt` selects the parser by the receipt version before any lossy decoding (schema 1 → driver input only; schema 3 → exact `{schema_version,package,build}` wrapper with the closed package parser; anything else unsupported), `DecodeInput` accepts either logical shape by its own version. `DecodeExpectedReceipt`'s whole-input equality now covers the package.
- `internal/buildcache`: distinct receipt-3 namespace `cache/build/go-v1-receipt-3` beside the legacy `cache/build/go-v1` (`Namespace`, `SourceAwareNamespace`, `pathsIn`); `Inspect`/`Publish` select the namespace from the expectation/publication input; `PublicationResult.SourceAware` lets `Revert` find the same slot; `Sweep` sweeps both roots with one referenced-key set and reports an unprovable boundary once. Because `closureexec.AssuredBuildCacheInput` hashes `Input.CacheKey()`, the assured key of a source-aware input is derived from the wrapper too; the execution receipt's `build_input_sha256` is therefore the exact receipt-3 input digest (assurance permits/receipts keep their versioned shapes).
- `internal/buildrepo`: `PipelineRequest.Package`, `CompileRequest.Package` (the compiler session binds it into the execution receipt), `PipelineResult.ReceiptSchemaVersion`; `receiptInput` wraps the unchanged receipt-2 driver input (`legacyReceiptInput`: every declared/effective identity, locked commit, tag, transport, substitution, build_source, descriptor target, target, toolchain, policy, assurance) under `{schema_version:3,package,build}`; `ArtifactsDir(version)` maps receipt 2 → `artifacts`, receipt 3 → `artifacts-receipt-3`; `DiskProtectedStore.LookupArtifact/StoreArtifact` derive version and namespace from the input shape and enforce the closed record shape for that version; `artifactPathFromInput` and GC's snapshot reachability read the driver input through the wrapper; `Collect` sweeps both artifact namespaces. New gate in `RunPipeline`: the session-derived build input must bind exactly the requested package (`reflect.DeepEqual`) or the run is refused with `build_repository_receipt_invalid` before cache adoption or dispatch — "an implementation unable to bind that input MUST reject execution".
- `internal/install` (shared, additive): `draftBuildPackages(lock)` selects the receipt package of every member whose marker is schema 5 (the local-snapshot arm, exactly what `buildDraftMarker` stages; Git draft members keep their accepted legacy receipts until their marker migrates with them — the accepted 17ps6u bound); `buildPlanRequest.packages` → `planOne` sets `Input.Package`; `StageRequest.Package` flows to the production builder (`builddeps.go`) and the assured builder check (`assurance.go`), so the execution receipt and its validation both bind the wrapper digest; `planExternalBuilds(..., packages, ...)` → `plannedExternal.pkg` → `PipelineRequest.Package`, `externalGoAdapter.Compile` → `StageRequest.Package`, `externalBuildInput` binds the package; external cache paths (`external.go` ×3, `targets.go`) use `buildrepo.ArtifactsDir(result.ReceiptSchemaVersion)`; `externalMarkerBuild` records the pipeline's receipt version; `buildDraftMarker` binds `receipt_schema_version 3` + `execution_policy` on every marker-5 build entry for both drivers and any skill schema. `global.go` passes nil packages (byte-identical).
- `internal/marker` (shared, additive; the sibling's comment invited this narrowing): marker-5 build entries must carry receipt version 3 on both arms (`validV5Build` via `validDriverBuild`); v2/v3/v4 rules unchanged (`validV3Build` = versions 1/2).
- `internal/godriver`: no change needed — its build input is validated only, never keyed or persisted; toolchain readiness/failure behaviour untouched (asserted at `install.Project`).

Not added: prebuilt download, compiler bootstrap, spec edits, Git aliases/mirrors/ports, registry changes.

## Shared edits (outside the literal scope line, minimal and additive)
`internal/install/{plan,stage,builddeps,assurance,external,targets,install,global,draftruntime}.go` (plumbing of the package identity and the namespace-aware external paths), `internal/marker/marker.go` (marker-5 receipt-3 binding), `internal/buildmeta/*` (receipt model shared by the build arms). Test-only call-site updates: `internal/install/{buildhttps,buildsshprecheck,drafttransport,stage}_test.go`, `internal/buildrepo/pipeline_test.go` (fakes now bind `request.Package`; `fakeBuilder.dropPackage` is the negative seam). Sibling files (`draftaudit*.go`, `sourceaudit.go`, `runtimestore/sourcev1.go`, `scopes/gc.go`) untouched.

## Changed paths
```
 M internal/buildcache/cache.go            M internal/install/draftruntime.go
 M internal/buildcache/collect.go          M internal/install/external.go
 M internal/buildcache/publish.go          M internal/install/global.go
 M internal/buildmeta/codec.go             M internal/install/install.go
 M internal/buildmeta/models.go            M internal/install/plan.go
 M internal/buildrepo/gc.go                M internal/install/stage.go
 M internal/buildrepo/pipeline.go          M internal/install/stage_test.go
 M internal/buildrepo/pipeline_test.go     M internal/install/targets.go
 M internal/buildrepo/protected.go         M internal/marker/marker.go
 M internal/install/assurance.go           ?? internal/buildcache/sourceaware_test.go
 M internal/install/builddeps.go           ?? internal/buildmeta/package.go
 M internal/install/buildhttps_test.go     ?? internal/buildmeta/receipt_v3_test.go
 M internal/install/buildsshprecheck_test.go ?? internal/buildrepo/receipt_v3_test.go
 M internal/install/drafttransport_test.go ?? internal/install/draftbuild_test.go
                                           ?? internal/marker/marker_v5_builds_test.go
```

## Tests (production entries named; positive + negative per acceptance clause)

Production entry `install.Project` with `DraftSourcesV1=true`, real protected build cache under the manager home, fake toolchain/builder, fake external acquisition (`internal/install/draftbuild_test.go`):
- `TestDraftBuildsPublishReceipt3OnBothArms` (+): one marker-5 member with a local `go-v1` and an external `go-repository-v1` command; both entries only under the receipt-3 namespaces; receipts are `{3, cache_key, {3, package, build}, artifact}` with `input.package` byte-equal to the lock/marker package; the local build is the unchanged schema-1 input; the external build is the unchanged receipt-2 input with declared/effective identities, locked commit, transport, substituted=false, build_source, descriptor target, assurance; external key = SHA-256(CCJ-1(wrapper)); marker-5 records bind receipt version 3, retained external fields, and the protected receipt hash; reinstall takes exact cache hits on both arms without compiling (build-bearing nodes are re-staged from the cache on every lane, so the context line reads "installed" — same as v1).
- `TestDraftBuildKeyFollowsThePackageIdentity` (+/−): context-only edit stays pinned (no rebuild); explicit refresh changes the package → both arms re-keyed, both rebuilt, marker rebound with receipt 3; the previous local entry stays in the receipt-3 namespace, the previous external entry is collected as unreferenced (existing external GC behaviour).
- `TestDraftBuildRefusesASessionThatBindsTheLegacyDigest` (−, both arms): an execution receipt over the context-only digest fails the install closed; no entry under any of the four namespaces, no marker.
- `TestDraftBuildToolchainFailureBehaviorIsPreserved` (−): toolchain establish failure fails before any cache lookup/compile/persistent state, inventory row keeps its driver and no key — as on v1.
- `TestLegacyBuildsKeepReceipt1WithTheSwitchOff` (legacy control): frozen v1 Git skill → schema-1 receipt (9-field input, no package) under `cache/build/go-v1`, no receipt-3 namespace, legacy marker.

Production entry `buildrepo.RunPipeline` / `DiskProtectedStore` (the exact call sites of `planExternalBuilds`/`stageExternalBuilds`), `internal/buildrepo/receipt_v3_test.go`:
- `TestExternalReceipt3WrapsTheReceipt2InputOnTheExternalArm` (+): namespace, shape, byte-identical inner receipt-2 input, key derivation, cache hit on rerun, package-less request is a miss and publishes to the legacy namespace, package mutation is a miss.
- `TestExternalReceipt3RefusesEveryEvidenceFieldMismatch` (−, 36 rows + receipt schema version + cache key): on-disk receipt mutated in package (snapshot, kind, extra field, removed), wrapper schema, build schema, declared identity value/kind, declared transport, locked commit hex/format, declared tag, effective identity value/kind, commit, object format, transport, substituted, substitution, build_source digest/algorithm, repository, descriptor target/path, command, build root, source dir, driver, target goarch/tuning, toolchain digest/version, policy execution/source_kind/network, assurance mode/removed, top-level receipt schema (wrapper input kept), cache key → `LookupArtifact` refuses with `build_repository_receipt_invalid` and the pipeline rebuilds (`would-rebuild-untrusted-cache`).
- `TestExternalReceipt3EntriesNeverCrossNamespaces` (−): receipt-2 record at the receipt-3 key and receipt-3 record at the legacy key both refused.
- `TestExternalReceipt3RequiresTheSessionToBindTheWrappedInput` (−): package-blind session refused at compile (nothing published); protected entry whose execution receipt binds the legacy digest refused at lookup and rebuilt.
- `TestCollectSweepsTheReceipt3Namespace` (+/−).

`internal/buildmeta/receipt_v3_test.go` (codec, all three package arms): wrapper shape/key/round-trip/expected-input equality (+); legacy bytes/key/hash golden unchanged (+); every input mutation changes the key (10 rows); 26 closed-shape refusal rows (receipt 1 over wrapper, receipt 2, wrapper schema/unknown/missing fields, package null/empty/unknown kind/foreign arm/extra field/commit null/extra/uppercase/length/format, repository `.git`/uppercase host, directory escape/glob, snapshot-into-git-arm, non-string field, build schema 3, package smuggled into build, inner/legacy cache keys) + re-labelled legacy receipt, smuggled package in a schema-1 receipt, duplicate wrapper key, non-canonical input; invalid packages refused by `Validate`/`CacheKey`/`NewReceipt` (7 rows).

`internal/buildcache/sourceaware_test.go` (real protected store): receipt-3 namespace publication/inspect/reuse/revert (+); no legacy hit satisfies a source-aware lookup and vice versa, different package is a miss, entries planted across the namespace boundary refused (−); sweep of both namespaces (+/−).

`internal/marker/marker_v5_builds_test.go`: v5 local/external/both records with receipt 3 round-trip (+); 12 refusal rows (receipt 1/2/absent on either arm, missing policy, repository state on local, hash-only external record, missing source/effective, commit length, substitution bool) (−); legacy v4 marker still refuses receipt 3.

Windows: no new skips; the new fixtures use no POSIX-only scripts (fake builder/toolchain, canonical bytes only). Real-platform evidence for windows/ubuntu comes from the hosted gate at handoff; not verified here.

## Commands and exit codes (all run by me in this session, standalone processes)
| command | exit |
|---|---|
| `go build ./...` | 0 |
| `gofmt -l internal cmd` | 0, no output |
| `git diff --check` | 0 |
| `go vet ./internal/buildmeta ./internal/buildcache ./internal/buildrepo ./internal/marker ./internal/install ./internal/scopes ./cmd/...` | 0 |
| `golangci-lint run ./internal/buildmeta/... ./internal/buildcache/... ./internal/buildrepo/... ./internal/marker/... ./internal/install/...` | 0 (0 issues) |
| `go test -p 1 -count=1 ./internal/buildmeta ./internal/buildcache ./internal/marker` (whole packages) | 0 |
| `go test -p 1 -count=1 ./internal/buildrepo` (whole package, 78.7s) | 0 |
| `go test -p 1 -count=1 ./internal/sourcelock ./internal/scopes ./internal/closureexec ./internal/godriver` | 0 |
| `go test -p 1 -count=1 -run 'TestDraftBuild\|TestLegacyBuildsKeepReceipt1\|TestDraftLocal\|TestLegacyInstallUntouchedWhenDraftOff\|TestCacheConformance' ./internal/install` | 0 (49s) |
| `go test -p 1 -count=1 -timeout 600s -run 'TestDraft\|TestCacheConformance\|TestDryRun\|TestStaging\|TestStaged\|TestCacheHit\|TestSecondBuild\|TestGlobal\|TestEndToEnd\|Toolchain\|External\|BuildSSH\|BuildHTTPS\|Revision\|TestLegacy' ./internal/install` | 0 (284s) |
| `go test -p 1 -count=1 -timeout 600s -run 'Build\|Status\|Draft\|Marker\|Gc\|Collect' ./cmd/curator` | 0 (434s) |

Not run: the full `internal/install` and `cmd/curator` packages and the whole repository suite (host stalls; the remote gate runs the configured landing suite once at handoff). Byte-identity of the frozen v1 lane rests on the pinned buildmeta goldens (`TestGoV1GoldenInputKeyReceiptAndHash` + `TestLegacyReceiptBytesAreUnchangedWithoutAPackage`), the untouched `legacyReceiptInput` map and `TestExternalReceiptV2CacheKeyVector`, `TestLegacyBuildsKeepReceipt1WithTheSwitchOff`, and the green legacy subsets above.

## Narrowing mutants (applied, tested, restored from byte copies; `go build ./...` green after each restore)
| mutant | test | exit |
|---|---|---|
| M1 `buildmeta.parsePackage` drops the exact-field-set check (extra/foreign package fields admitted) | `TestSourceAwareReceiptReaderRefusesEveryShapeConfusion` | 1 (killed) |
| M7 `parseSourceAwareInput` drops the exact wrapper field check | same | 1 (killed) |
| M2 `buildcache.namespaceName` always returns the legacy namespace | `-run 'SourceAware\|Receipt3' ./internal/buildcache` | 1 (killed) |
| M3 `DiskProtectedStore.LookupArtifact` accepts receipt schema 2 or 3 regardless of the input shape | `TestExternalReceipt3RefusesEveryEvidenceFieldMismatch` | first run 0 (survived: input/key equality already refused the planted legacy record) → added row "receipt schema version" (wrapper input kept, version re-labelled 2) → rerun 1 (killed); control without mutant 0 |
| M4 `RunPipeline` drops the session package-binding check | `TestExternalReceipt3RequiresTheSessionToBindTheWrappedInput` | 1 (killed) |
| M5 `marker.validV5Build` also admits the v3/v4 receipt versions | `TestMarkerV5BuildsBindReceiptVersion3OnBothArms` | 1 (killed) |
| M6 `install.draftBuildPackages` returns nil (draft members keyed like legacy) | `TestDraftBuildsPublishReceipt3OnBothArms` | 1 (killed) |

Two earlier mutant candidates were equivalent and are reported as such, not as kills: removing `parsed.Validate()` at the end of `parsePackage` and collapsing `DecodeReceipt`'s version switch are both caught by `Receipt.Validate` (package validation, receipt-version-to-input binding) — the second gate, not the mutated one, refused.

## Bounds / findings
- Receipt-3 applies to every member whose marker is schema 5, i.e. the local-snapshot arm (both build drivers). Git draft members still install with their accepted legacy markers and legacy receipts (17ps6u's recorded bound); migrating them means marker-5 for Git packages (attestation rules) and is not part of this leaf's scope line.
- `buildcache.Quarantine(key, lock)` keeps its legacy-namespace address (no production caller); `Revert` follows the recorded namespace.
- A first mutant pass restored files with `git checkout`, which reverted my uncommitted edits in six files; all were reapplied from the recorded edit scripts and the whole-package/install rows were rerun green before the second mutant pass (which restores from byte copies). The exit codes above are from runs on the final tree.
- No `logbook` CLI on this host and LOGBOOK.md edits are forbidden by campaign rules; findings recorded here and in board notes.

Ready for review (final leaf of STORY-260910-20sx61 handoff, developer role).

---

# Revision 2 — hosted-gate Windows failure of the external arm (rework 1)

Base checkpoint unchanged: `832facfc799f6749fdd6c1428cf096009444c738` (uncommitted candidate on the Story branch).

## Root cause (read from the code paths; Windows cannot run here)
The refusal `build_repository_protected_boundary_untrusted: protected directory cannot be proved private: protected object DACL is inheritable` is `nativeProtectedDir`'s initial proof (`internal/buildrepo/protection_windows.go`), which trusts only directories `createWindowsPrivateDirs` itself created (`os.Mkdir` success → `secureWindowsPath`). Two parents of the external arm were never created that way:

1. **Staging store root (the first-install failure at draftbuild_test.go:177/299).** `stageExternalBuilds` used `private.dir("external-cache-")` — an `os.MkdirTemp` directory with an inherited DACL — directly as `DiskProtectedStore.Root`. The first `StoreSnapshot` → `prepare()` → `protectedDir(root, true)` found the root existing (`Mkdir` → `IsExist`, so no securing) and refused it. The same code ran on the legacy external arm; it was never observed because every legacy external-arm install test skips on Windows ("test transport wrapper is POSIX-only"), while the Windows store unit tests use `filepath.Join(t.TempDir(), "cache")` and let the store create its root.
2. **Final-root parents (suspect (a) of the brief; would have failed on the reinstall/lookup).** The transaction `plan.Replace("05-external-cache", …, filepath.Join(finalRoot, ArtifactsDir(version), key), …)` relies on the commit's generic `makeMissingDirectories` (`os.Mkdir(path, 0o755)`, inherited DACL) to create `external-build-cache`, `snapshots` and the artifact namespace; the store's later `protectedDir(parent, false)` refuses such parents on Windows. Pre-existing for the legacy `artifacts` parent for the same reason.

## Pre-existing gap fixed
Both parents are inside the scope line (current build driver integration). The external arm could not publish into a protected store on Windows before this leaf: item 1 refused the very first snapshot publication, item 2 would refuse any later lookup. Fixed here at the root; Windows verification is not weakened.

## Change
- `internal/buildrepo/protected.go:41-63` — new `DiskProtectedStore.PrepareNamespaces(receiptSchemaVersions ...int)`: creates the root, `snapshots` and the artifact namespace of each named version through `protectedDir(name, true)` (the store's own private creation: non-inheritable owner-private DACL on Windows, `MkdirAll 0700` on unix). Unknown versions are refused; an existing parent the store cannot prove private is refused, never repaired (same proof as before).
- `internal/install/external.go:210-227` — the staging store root is `filepath.Join(privateDir, "store")`, created by `store.PrepareNamespaces()` before the first pipeline run, so the store proves a directory it created itself.
- `internal/install/external.go:280-311` — `stagedExternal.prepareFinalNamespaces(finalRoot)` runs at the end of `stageExternalBuilds` (after `VerifyToolchain`, only when at least one non-existing entry will be published) and prepares the final root, `snapshots` and the artifact namespace of exactly the receipt schema versions being published. The commit's generic scaffolding then finds every protected parent present. Dry-run never reaches this (stageExternalBuilds is not called in dry-run); a refused install with no publication prepares nothing.
- unix behaviour: parents are now created by the store at 0700 (previously the commit's scaffolding made `external-build-cache`/`artifacts…` 0755, which the 0o022 proof tolerated). Entry contents, receipts, keys and namespaces are unchanged.
- Shared edits: none beyond `internal/install/external.go` (my scope's external-arm integration); no edits to staging/commit, 17ps6u or hwxr26 files.

## Tests (production entries)
- `internal/buildrepo/receipt_v3_test.go:426` `TestPrepareNamespacesCreatesEveryProtectedParentPrivately` (+): root/snapshots/both artifact namespaces exist, 0700 on unix, each proved by `protectedDir(path,false)`; idempotent; unknown version refused; exactly three root entries.
- `internal/buildrepo/receipt_v3_test.go:459` `TestPrepareNamespacesRefusesAForeignParent` (−): a world-writable pre-existing namespace is refused with `build_repository_protected_boundary_untrusted` and left unrepaired. Skips on Windows with the declared reason class `… is exercised on the unix runners` (platform-control; the Windows DACL refusal is asserted by the existing `TestWindowsProtectedSecurityDescriptorRejectsWrongOwnerAndDACL`).
- `internal/install/draftbuild_test.go:447` `TestDraftBuildFinalRootParentsAreStoreCreated` at `install.Project` (+): after a receipt-3 external publication the final root, `snapshots` and `artifacts-receipt-3` are store-created (0700 on unix), the legacy namespace is NOT prepared, `PrepareNamespaces` proves the parents, and a reinstall is a hit; sub-test `foreign parent refused` (−): a pre-created world-writable `artifacts-receipt-3` makes the install fail with the protected-boundary code before publishing anything or writing a marker (declared unix-only skip as above).
- `TestDraftBuildsPublishReceipt3OnBothArms` and `TestDraftBuildKeyFollowsThePackageIdentity` keep running on Windows unchanged (no skip added); they are the Windows regression for this fix through the hosted gate.
- Point 3 of the rework: the new tests use no script fixtures (fake builder/toolchain, canonical bytes) and no exec-bit assertions, so no `unix_path`/`win_path` declarations or exec-bit skip reasons apply.

## Commands and exit codes (zsh, standalone processes, this session, final tree)
| command | exit |
|---|---|
| `go build ./...` | 0 |
| `go vet ./internal/buildrepo ./internal/install` | 0 |
| `gofmt -l internal` | 0, no output |
| `git diff --check` | 0 |
| `golangci-lint run ./internal/buildrepo/... ./internal/install/...` | 0 (0 issues) |
| `go test -p 1 ./internal/buildrepo -run 'Protected\|ReceiptV3\|Receipt3\|PrepareNamespaces' -count=1 -timeout=120s` | 0 |
| `go test -p 1 ./internal/install -run 'TestDraftBuild' -count=1 -timeout=300s` | 0 (35.9s) |
| `go test -p 1 ./internal/install -run 'TestDraftBuild\|TestLegacyBuildsKeepReceipt1\|External\|BuildSSH\|BuildHTTPS' -count=1 -timeout=300s` | 0 (38.7s) |

## Narrowing mutants (applied, tested, restored from byte copies; `go build ./...` green after restore)
| mutant | test | exit |
|---|---|---|
| M8 `prepareFinalNamespaces` prepares only root+snapshots (drops the artifact namespace, so the commit's 0755 scaffolding creates it) | `TestDraftBuildFinalRootParentsAreStoreCreated` | 1 (killed: `artifacts-receipt-3 mode = 755, want 0700`) |
| M9 `protectedDir(create=true)` repairs a group/world-writable existing parent with chmod 0700 instead of refusing (both the generic and the unix native path) | `TestPrepareNamespacesRefusesAForeignParent`; `TestDraftBuildFinalRootParentsAreStoreCreated/foreign_parent_refused` | 1, 1 (both killed: install Status:ok with etool `outcome=corrupt`) |

Bound: reverting the staging root to the raw `MkdirTemp` directory is not killable on unix (MkdirTemp yields 0700 and the owner proof passes); its kill is Windows-only and rests on the hosted gate's `TestDraftBuildsPublishReceipt3OnBothArms`. Windows is verified by the hosted gate only; not run here.

Ready for review (revision 2, developer role).

# Revision 3 — Windows, second layer: the published external entry itself (rework 2)

Base checkpoint unchanged: `832facfc799f6749fdd6c1428cf096009444c738` (uncommitted candidate on the Story branch).

## Root cause (read from the code paths; Windows cannot run here)
Gate run 35329411348 (Windows) classified the external entry the first install published as `outcome=corrupt` on reinstall and the sweep reported `retained external receipt … is unreadable`, while the local arm hit. The exact Windows proof that fails is `validateWindowsSecurityDescriptor` (`internal/buildrepo/protection_windows.go`): `control&SE_DACL_PROTECTED == 0` → "protected object DACL is inheritable", reached from `nativeProtectedDir` on the entry directory (`LookupArtifact` → `protectedDir(entry,false)`) and from `readNativeProtectedFile` on `receipt.json` (the GC sweep's `readProtectedFile`).

Why the entry loses the proof: the transaction engine does not move the staged entry. `internal/transaction/staging.go` `createStagingEntry`/`copyStagingFile` recreate every directory (`os.Mkdir(dst, 0)` + `Chmod(mode)`) and every file (`OpenFile(O_CREATE|O_EXCL, 0)` + `Chmod(mode)` + byte copy) as a FRESH object in the transaction's sidecar, then rename that sidecar into the live path. On unix the copied modes (0700/0600, owner = euid, nlink 1) ARE the proof, so the copy is a clean hit. On Windows a freshly created object inherits its parent's DACL: `SE_DACL_PROTECTED` is clear and the ACE set is the inherited one, so the private-DACL proof that `secureWindowsPath` had applied in the staging store is gone. Rev2 made the parents store-created; the entry (directory, `artifact`, `receipt.json`, `execution-receipt.ccj.json`; and the snapshot entry with `files/…` and `snapshot.json`) was still a transaction-created copy. Hard-link counts are 1 (fresh files) and the receipt fields are byte-identical (copied bytes), so those suspects are excluded by construction.

The legacy receipt-2 external arm shares this path exactly (same `transactionPlan`, same sidecar copy) and was never exercised on Windows (its install tests skip: "test transport wrapper is POSIX-only"); it gets the same treatment below.

## Change (in scope: internal/buildrepo + current build-driver integration in internal/install; no edits to internal/transaction)
- `internal/buildrepo/protected.go:358-414` — new `DiskProtectedStore.AdoptArtifact(receiptSchemaVersion, key)`: proves the store root and the namespace parent WITHOUT creating either, requires the entry to be an existing real directory whose tree holds only directories and regular files (a symlink/reparse point is refused), re-secures the entry tree with the store's own `secureProtectedTree` (Windows: `secureWindowsPath` on every object — owner, protected owner-only DACL, then `validateWindowsHandle`; unix: the walk-only mode, no chmod), reads the receipt through the protected reader, requires `cacheKey(receipt.input) == key` and `ReceiptSchemaVersionOf(input) == version`, and finally runs the ordinary exact `LookupArtifact(key, input, mutate=false)`. The entry is trusted only by that lookup; nothing is written and a refused entry is left in place (no quarantine) so the evidence survives. `validateWindowsHandle` is untouched.
- `internal/buildrepo/protected.go:416-429` — `AdoptSnapshot(key)`: same for the snapshot namespace, proved by `LoadSnapshot(key, false)`.
- `internal/buildrepo/protected.go:431-466` — shared `adoptionEntry` (root/parent proof, shape walk, secure).
- `internal/install/external.go:313-360` — `externalAdoption` + `stagedExternal.adoptions(finalRoot)`: the exact set `transactionPlan` publishes (one artifact per newly built command in its receipt namespace, each snapshot once; nothing for a final-root hit), artifacts before snapshots.
- `internal/install/commit.go:518-522, 696, 721-742` — `scopeTargets.adoptions` and `adoptExternalPublications`, called in `runCommit` immediately after `Journal.Commit`, still under the home lock and BEFORE `collectAfterCommit` (the sweep that refused the unreadable receipt). Both arms (receipt-2 legacy and receipt-3) go through it because both are in `transactionPlan`. The journal commit is durable at that point, so a failed adoption is a warning naming the kind and key (`external build cache artifact <key> was published but could not be adopted: …`), never a revert and never a write.
- Shared edits: `internal/install/install.go:784` and `internal/install/global.go:406` — one additive line each (`targets.adoptions = request.external.adoptions(request.externalStoreRoot)`) next to the existing `transactionPlan` merge. No 17ps6u/hwxr26 files touched.
- unix behaviour is unchanged for the entries themselves (walk-only securing; the lookup already passed); the only new unix effect is the extra verification pass after commit.

## Tests
- `internal/buildrepo/adopt_test.go` `TestAdoptArtifactMakesATransactionCopyACleanHit` (+, both namespaces `artifacts` and `artifacts-receipt-3`): `transactionCopy` reproduces the engine's fresh-object copy (Mkdir 0 + Chmod, O_CREATE|O_EXCL + Chmod + bytes) of a staged entry into a prepared final namespace; on Windows the unadopted copy MUST be refused with a "DACL" error (the gate's failure, asserted as the negative) and on unix it is already a hit (mode proof); after `AdoptArtifact` the exact lookup is a clean hit with the original bytes, adoption is idempotent, and no quarantine/objects were created. This is the requested unit test "an entry adopted from a staged store is a clean hit", exercising Windows securing on Windows and the unix mode on unix.
- `TestAdoptArtifactRefusesWhatItCannotProve` (−, 11 rows; 10 on Windows): unknown schema version; malformed key; absent store root (nothing created); unprepared namespace; absent entry; entry is a file; entry copied under another key; receipt-3 entry copied into the legacy namespace; tampered artifact bytes (`build_repository_artifact_invalid`); tampered receipt; missing execution receipt (`build_repository_receipt_invalid`); symlink inside the entry (unix only, Windows symlink creation needs a privilege — not a skip, the row is simply not appended). Each refusal is an error, no quarantine, no root created.
- `TestAdoptSnapshotMakesATransactionCopyACleanHit` (+/−): absent store, absent entry, Windows DACL refusal before adoption / unix hit, clean load after adoption, tampered file refused with `build_repository_object_semantics_invalid` and left in place.
- `internal/install/draftbuild_adopt_test.go` `TestExternalAdoptionsListEveryPublishedEntryOnce` (+/−): the adoption set equals the transaction's publication set (artifact per built command in its namespace, snapshot once, existing hit excluded, ordering); `TestAdoptExternalPublicationsWarnsWithoutCreatingAnything` (−): unprovable publications produce the scoped warning per entry and create no store root; an empty set warns nothing.
- Production regression on Windows: `TestDraftBuildsPublishReceipt3OnBothArms` (:279 reinstall hit), `TestDraftBuildKeyFollowsThePackageIdentity` (:311) and `TestDraftBuildFinalRootParentsAreStoreCreated` (:479) keep running on Windows unchanged (no skip). No script fixtures or exec-bit assertions were added (no `unix_path`/`win_path` or skip reasons apply).

## Commands and exit codes (zsh, `set -o pipefail`, standalone processes, this session, final tree)
| command | exit |
|---|---|
| `go build ./...` | 0 |
| `go vet ./internal/buildrepo ./internal/install` | 0 |
| `gofmt -l internal/buildrepo internal/install` | 0, no output |
| `git diff --check` | 0 |
| `golangci-lint run ./internal/buildrepo/... ./internal/install/...` | 0 (0 issues; one QF1001 De Morgan finding fixed in `adoptionEntry` before this run) |
| `go test -p 1 ./internal/buildrepo -count=1 -timeout=180s` (whole package, background + tail) | 0 (91.8s) |
| `go test -p 1 ./internal/install -run 'TestDraftBuild\|TestLegacyBuildsKeepReceipt1\|TestExternalAdoptions\|TestAdoptExternalPublications' -count=1 -timeout=400s` (background + tail) | 0 (40.5s) |
| `go test -p 1 ./internal/buildrepo -run 'TestAdopt' -count=1 -timeout=120s` | 0 |

## Narrowing mutants (applied, tested, restored from byte copies; `cmp` against the copies and `go build ./...` green after restore)
| mutant | test | exit |
|---|---|---|
| M1 `AdoptArtifact` key check narrowed to `derived == ""` (no longer compares to the entry key) | `TestAdoptArtifactRefusesWhatItCannotProve/entry_copied_under_another_key` | 1 (killed) |
| M2 `AdoptArtifact` returns nil without the exact `LookupArtifact` | `…/tampered_artifact_bytes`, `…/missing_execution_receipt` | 1 (killed) |
| M3 `adoptionEntry` skips `secureProtectedTree` | `TestAdopt*` | 0 on macOS — SURVIVOR on unix by design (unix securing is walk-only); killed on Windows by the "unadopted copy is refused with DACL / adopted copy is a hit" pair in `TestAdoptArtifactMakesATransactionCopyACleanHit` and `TestAdoptSnapshotMakesATransactionCopyACleanHit`, and in production by the three Windows reinstall rows above — hosted-gate evidence only |
| M4 `AdoptSnapshot` returns nil without `LoadSnapshot` | `TestAdoptSnapshotMakesATransactionCopyACleanHit` | 1 (killed) |
| M5 `adoptions()` omits snapshots | `TestExternalAdoptionsListEveryPublishedEntryOnce` | 1 (killed) |
| M6 `adoptions()` returns nil (adoption never reaches production) | `TestExternalAdoptionsListEveryPublishedEntryOnce` killed it (exit 1); `TestDraftBuildsPublishReceipt3OnBothArms` passed under it on macOS | bound: on unix the production reach of `adoptExternalPublications` is only proved by the helper-level test; its production kill is the Windows reinstall hit at `draftbuild_test.go:279`, i.e. exactly what gate run 35329411348 failed on |

## Bounds
- Windows behaviour (DACL loss on the transaction copy and its repair by adoption) is verified by the hosted gate only; not run here.
- Adoption runs after the durable journal commit, so it cannot fail closed before the write; it fails closed before TRUST (an unprovable entry is never a hit and is reported). Quarantine is deliberately not triggered by adoption (mutate=false) so the reviewer/next lookup sees the evidence; the next mutating lookup quarantines it exactly as before.
- `internal/transaction` was not changed (out of scope); the engine's fresh-object copy remains the publication mechanism for every other class.

Ready for review (revision 3, developer role).
