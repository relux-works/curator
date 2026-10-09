TASK-260924-4mzun5 rework-2 results (developer) — F1–F5 answered
==============================================================

Decision (retained, reviewer-confirmed in rev1 verdict): RECORD.
skillfile-sources §4 requires marker v6 for a core schema-9 installation
even under a schema-1 Skillfile; the v6 carrier records the directory and
lock_sha256 "binds the installed selection and declared ref through the
validated lock and matching manifest". The draft-sources-v2 suite
(agent-skill-v9 / csk-skill-v9, install-marker v6) is consumed
byte-identical from curator-spec 7eaeb73f.

Round 2 carries the v6 migration through runtime, lock, builds, moved
tags, audit, and status. One staged migration per install
(internal/install/legacypackage.go: planLegacyLane) derives packages,
source-v1 runtime leaves, receipt-3 build identities, and the shared
effective-lock digest from the resolved closure, so marker, runtime,
receipts, and audit bind one staged identity.

Finding → fix → test
--------------------
F1 (runtime publication disagrees with package markers):
- Fix: migrated legacy-lane nodes publish script runtimes under the
  source-v1 key of the same staged package the marker records
  (install.go projectAttempt, global.go globalAttempt/stageGlobalTargets).
  GC already marks that key from the marker, so runtime/marker/GC agree.
  Status drift for legacy package markers compares the recorded package
  commit against the live resolution (cmd/curator/main.go
  scopeStatusDrift); full package/lock comparison runs at install time
  via marker.Current.
- Test: reviewer's TestReviewLegacyRuntimeMatchesPackageMarker is now the
  permanent TestLegacyLaneRuntimeMatchesPackageMarker (project entry:
  v6 marker, source-v1 runtime present, no commit-keyed twin, shim
  launches, second install up-to-date, scopes.CollectRuntime keeps the
  live tree) plus TestGlobalLaneRuntimeMatchesPackageMarker (global
  entry) and TestCLIStatusReportsLegacyPackageMarkerCurrent (status
  up-to-date + --check exit 0).

F2 (invented per-node lock_sha256 preimage):
- Fix: deleted legacyMarkerPackage/legacyLockDigest. The schema-1 lane
  now constructs and validates the skillfile-lock schema-1 document in
  memory (legacyEffectiveLock: manifest CCJ-1 over the declaring
  Skillfile bytes, one member per closure node with directory/package/
  content hash, root selection indexes) and records lock.LockSHA256 —
  the specified §3 digest, computed by sourcelock.New, never a new
  preimage. Staged once per install, before audit/registry/compiler
  work; dry runs stage packages but no lock (no marker is written).
- Test: TestLegacyLaneRecordsDependencyDirectory asserts all migrated
  markers share one digest AND that it equals an independently
  reconstructed lock built only from exported primitives
  (ManifestDigest + closure.Build + NetworkGit/ConfiguredGitPackage +
  ContentHashFor + sourcelock.New). The rev1 per-node binding fails both
  pins. TestLegacyLockPackageArms pins the staged arms and the
  subdir-without-identity / malformed-commit fail-closeds.

F3 (blanket refusal of migrated nodes with builds):
- Fix: refusal deleted. Migrated nodes flow through the package
  receipt/build pipeline on both lanes (legacy BuildPackages into
  planBuilds and planExternalBuilds, project + global); buildMarker
  records their builds with receipt 3 + execution policy on every
  driver, exactly as buildDraftMarker does (core §4.4 permits builds
  from schema-9 providers; §4 specifies receipt-3 package build
  records). marker.Write refuses anything else under a package marker.
- Test: TestLegacyLaneBuildsPublishReceipt3 (real install entry, fake
  toolchain/builder + real protected cache: v6 marker, receipt-3
  record, receipt-3-namespace-only entry, receipt input.package ==
  marker package, one compilation, reinstall takes exact cache hits and
  rebuilds nothing) and TestLegacyLaneBuildMarkerRebindsTamperedReceipt
  (bogus cache_key in the marker file is not adopted: reinstall
  re-derives the true receipt-3 record from the cache). The unit refusal
  pin was replaced by TestBuildMarkerRecordsReceipt3ForMigratedLegacyNode
  (receipt-3 upgrade + fail-closed without a staged migration).

F4 (moved-tag policy skipped for legacy package markers):
- Fix: detectMovedTagsIn takes the lane. Draft package markers still
  skip (frozen lock binds the ref; moves observable only at refresh).
  Legacy package markers compare the recorded package commit against
  the live declared tag binding and warn on difference (manager §2.1
  keeps moved-tag evaluation mandatory; strict mode refuses as before).
  Package presence alone proves nothing; a package without a locked
  commit stays silent.
- Test: TestLegacyLaneMovedTagRefusesStrictSecondInstall (real
  second-install strict refusal naming old→new commits, installation
  undisturbed; default reinstall warns and rebinds the marker commit).

F5 (audit record/cache binds only directory):
- Fix: Subject carries the complete source-types schema-1 package
  identity (audit.SkillPackage constructor for the Git arms; draft lane
  uses the frozen lock member; CheckSourceAudit binds its frozen
  package). The verdict record stores the stated package; cache
  equality covers repository/commit/directory plus the stored commit.
  Stated↔unstated never match either way; malformed/absent identity on
  the compared side is a miss, never a hit.
- Test: TestVerdictCacheBindsFullPackageIdentity (same-identity hit incl.
  CacheHit-true auditSubject row; different directory/repository/commit/
  kind, unstated subject, and legacy record misses; local-snapshot arm),
  TestSkillPackageArms (network/configured/sha256/invalid arms), and
  full-identity assertions on install-entry records
  (TestLegacyLaneRecordsDependencyDirectory) and CLI audit records
  (TestCLIAuditBindsDependencyDirectory).

Files changed (round 2; production)
------------------------------------
- internal/install/legacypackage.go — rewritten: legacyMigrated,
  legacyLockPackage, legacyLanePlan/planLegacyLane, legacyEffectiveLock.
- internal/install/install.go — staged migration (7b), audit subjects,
  build/runtime/marker plumbing, moved-tag lane flag, auditPackageForNode.
- internal/install/global.go — same carry-through for the global lane;
  runtimeKeys/legacy on globalTargetRequest.
- internal/install/generation.go — readManifestDocument also returns the
  parsed payload bytes (single-read discipline preserved).
- internal/install/draftruntime.go, targets.go — comment-only lane notes.
- internal/audit/audit.go — Subject.Package, SkillPackage, package+commit
  cache equality, loadCachedFindings takes the subject.
- internal/audit/sourceaudit.go — source-audit verdict binds the frozen
  package.
- cmd/curator/main.go — CLI audit subjects; status drift for legacy
  package markers.
- CHANGELOG.md — Unreleased entry extended (operator-visible).

Files changed (round 2; tests)
-------------------------------
- internal/install/legacy_runtime_test.go (new: F1 rows),
  legacy_builds_test.go (new: F3 rows),
  legacy_movedtags_test.go (new: F4 row),
  legacy_directory_test.go (F2/F5 assertions, rewritten unit pins),
  generation_read_state_test.go (moved-tag lane flag).
- internal/audit/package_test.go (new: F5 rows), directory_test.go,
  audit_test.go, opaque_v1_test.go, opaque_v1_rework_test.go
  (loadCachedFindings signature).
- cmd/curator/audit_directory_test.go (F5 CLI package assertions, status
  currentness row).

NOT touched: internal/skillspec/testdata/draft-sources-v1 and
internal/marker/testdata/draft-sources-v1 (zero diff, trunk content).
draft-sources-v2 copies re-verified 24/24 byte-identical to spec 7eaeb73f
(23 skillspec conformance files + install-marker-v6 schema).

Compile-only tail (R223: no local go test; hosted gate is the arbiter)
-----------------------------------------------------------------------
- go build ./... → exit 0
- go vet ./... → exit 0 (compiles all tests, incl. the new rows)
- gofmt -l internal cmd → clean
- golangci-lint run ./internal/install/ ./internal/audit/
  ./cmd/curator/ → 0 issues, exit 0
- go test → NOT RUN (R223 forbids local suite execution on the mini)

Checklist note: item 20 (Tests green) is deliberately left unchecked —
no suite ran locally per R223; the hosted gate on this candidate is the
arbiter. Items 18/19 (matches AC / fits architecture) were verified by
direct re-read of the round-2 delta against the verdict and spec.

Coverage bounds (stated, not silent)
-------------------------------------
- Legacy external (go-repository-v1) builds flow through the same
  `packages` parameter the draft lane pins, but round 2 adds no
  dedicated legacy external-arm row; the acceptance row covers the local
  go-v1 arm plus the shared receipt-3 binding.
- Status compares the legacy package commit at the legacy surface's
  granularity (live ref resolution + commit); the full package/lock
  comparison runs at install time via marker.Current.
