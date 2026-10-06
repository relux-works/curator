# Validation — TASK-261006-3ptm2x rework 1 (rev2): NUL opaque gate scoped to v1 identities

Supersedes the revision-1 validation note. Revision-1 review
(`TASK-261006-3ptm2x_review-verdict-rev1.md`, RUN-261006-a8c933) requested
changes with three numbered findings; this revision fixes all three,
adapts the reviewer's seven attack probes into maintained regression
tests, and names the killing mutants. No version is ever inferred from
digest bytes: every rule below keys off a recorded `hash_version`, a lock
declaration, the lane, the writer switch, or the frozen shape.

## 1. Design note — complete call-site inventory

### 1a. `opaquescan.NULPaths` direct sites (4, unchanged from rev1)

| # | Site | Version in force | How known |
|---|------|------------------|-----------|
| 1 | internal/audit/audit.go gate() | per-subject: v1 scans, v2 skips, unknown refuses | Subject.HashVersion; 0 reads as v1 |
| 2 | internal/audit/audit.go auditSubject() | same per-subject rule | same Subject.HashVersion |
| 3 | internal/audit/audit.go detect() | v1-only by construction | canary (NUL-free fixture) and test callers only |
| 4 | internal/contextaudit/contextaudit.go Detect() | v1-only legacy entry | no production callers; versioned callers use DetectAtVersion |

### 1b. `opaquescan.RefuseNULV1` guard sites (new shared pre-hash refusal)

The guard scans only for v1, skips v2 entirely, refuses unknown
versions, and carries the shared id `audit.opaque.nul-byte` plus the
offending files.

| # | Site | Version in force | How known |
|---|------|------------------|-----------|
| 1 | marker.Current | recorded marker version | markerHashVersion(recorded): explicit HashVersion else shape default |
| 2 | cmd scopeStatusDrift | recorded marker version | recorded.ContentHashVersion() |
| 3 | cmd hybrid status | recorded marker version | recorded.ContentHashVersion() |
| 4 | cmd classifyDraftMember | recorded marker version | recorded.ContentHashVersion() |
| 5 | closure.ContentHashFor | frozen v1 | literal VersionV1; scans the FULL frozen snapshot before filtering or hashing (rev4 F1 correction: the rev2 text said "projected tree") |
| 6 | envprofile strictAuditMember | lock-declared version | resolvedHashVersion(resolved); 0 reads as v1 |
| 7 | envprofile storeEntryPinHashes | lock-declared version | lock.ContentHashVersion(); state pins only (commit pins use git tree OIDs, a different scheme) |
| 8 | envprofile skillsOf | lock-declared version | lock.ContentHashVersion() |
| 9 | contextstore.ContentHash | writer version | hashing.WriteVersion() |
| 10 | install resolveRegistries | draft lane v1, else writer | draftLock != nil |
| 11 | install targets.go stageNode | draft package v1, else writer | expected.Package != nil |

audit.go needs no guard call: gate() pre-scans v1 subjects before the
audit-enabled return, and auditSubjectWithOpaquePaths now refuses on
handed v1 NUL paths before hashing (same pre-hash property).

### 1c. Every v1 identity compute/trust site and its version source

| Site | Identity | Version source | Guard |
|------|----------|----------------|-------|
| audit gate content hash | ContentSHA256WithVersion(snapshot, version) | Subject.HashVersion (lane) | gate pre-scan + reorder double cover |
| draft source audit (CheckSourceAudit) | live detection hash | explicit v1 (only caller is the draft lane) | reorder; binding logic unreachable on NUL |
| closure.ContentHashFor | frozen package context hash | frozen v1 | 1b.5 |
| marker.Current | installed content identity | recorded.ContentHashVersion() | 1b.1 |
| 3 CLI drift/status recomputes | recorded installed hash | recorded.ContentHashVersion() | 1b.2–4, map to non-current states |
| install registry attestation hash | node snapshot identity | draft→v1 else writer | gates precede + 1b.10 |
| install staging hash | staged context identity | Package→v1 else writer | gates precede + 1b.11 |
| strictAuditMember revocation hash | member snapshot identity | lock declaration | DetectAtVersion precedes + 1b.6 |
| storeEntryPinHashes | recorded state pin | lock declaration | 1b.7 |
| skillsOf | lock member identity | lock declaration | 1b.8 |
| contextstore.ContentHash/EnsureState | store state hash | writer | 1b.9 |
| identity migration old-lock verify | old v1 pins | lock declaration | via guarded store boundary (1b.7); rehash is v2 |
| marker.Write v1 path | preserves caller digest, no local hash | caller (guarded staging) | none needed; no bytes hashed here |
| registry Matches/Resolve | compares caller digests | versioned carriers; unversioned legacy reads v1 | computations guarded upstream |
| verdict cache load/store | (version, digest) record identity | subject version; legacy schema-1 reads v1 | §2 finding 3 |
| source-audit binding | live v1 vs locked v1 compare | frozen v1 | live hash runs over scanned bytes only |
| contextaudit DetectAtVersion | secret-detector verdict | lock declaration | pre-existing: v1 scans, v2 skips, unknown refuses |
| migrateGlobalSkills | commit-pinned member audit | explicit v1 | DetectAtVersion precedes + 1b.6 |

Out of scope (distinct schemes, never §8 tree identities): snapshot
inventory digests, draft runtime keys (SHA-256 over CCJ-1 struct bytes),
marker file fingerprints (plain SHA-256 of the marker document),
materialized surface hashes (reviewer-scoped-out; manager-rendered
in-memory bytes), git tree object IDs. The test-only fixture setup may
still hash NUL trees directly through `hashing` — that is why the guard
lives at production call sites, not inside the hash functions.

## 2. Revision-1 findings answered

### Finding 1 (P1) — legacy installed-tree currentness trusted a v1 NUL collision

Fix: pre-hash `RefuseNULV1` at `marker.Current` (returns the opaque
error, never current) and at the three CLI recompute sites
(scopeStatusDrift → content-drift, hybrid status → content-drift,
classifyDraftMember → unresolvable). The opaque text surfaces in the
same CLI invocation through the install-status refusal.

- Regression tests: `TestCurrentLegacyMarkerNULCollisionRefuses`
  (real equal-v1-digest fixtures, clean control), `TestReadStateRefusesV2ToV1Downgrade`
  (marker, local green); `TestV1ReinstallRefusesNULAppearingAfterInstall`
  with a self-verifying collision-preserving edit plus status-plan rows
  (install, hosted); `TestStatusCheckRefusesV1NULAppearingAfterInstall`
  incl. direct drift-classifier rows plus the v2 currentness control
  `TestStatusCheckAdmitsV2NULBearingInstall` (cmd, hosted).
- Mutants: M4 guard deletion in `marker.Current` — executed, killed
  (`current=true` reproduces the original bug). Narrowing: a guard that
  scanned only top-level entries is killed by the nested `assets/a.bin`
  fixtures at every level (`TestRefuseNULV1ScansNestedFiles` pins the
  shared helper; CLI/install collision edits use nested paths).

### Finding 2 (P1) — v1 identities computed before the opaque refusal

Fix: `auditSubjectWithOpaquePaths` refuses on v1 NUL paths before
hashing and the refusal carries no digest; `closure.ContentHashFor`
scans the full frozen snapshot before filtering or hashing (rev4 F1
correction: the rev2 text said "the projection"); guards added at the
lock/store readers (strictAuditMember, storeEntryPinHashes, skillsOf),
the store hash (contextstore.ContentHash), and the install registry and
staging hashes.

- Regression tests: `TestAuditSubjectV1NULBlockCarriesNoComputedIdentity`
  (empty-digest assertion; rev4 F3 correction: revision 3 proved this
  shape does NOT establish order — a discarded hash-before-refuse mutant
  passes it — so rev4 adds `hashing.CountV1Hashes` observation at every
  guarded entry and the "no instrumentation needed" claim is withdrawn),
  `TestCheckSourceAuditV1NULRefusesWithoutVerdict`,
  `TestGateMixedVersionsRefuseOnlyTheV1Subject` (audit, local green);
  `TestContentHashForRefusesNULBeforeHashing` (closure, local green);
  `TestStrictAuditMemberRefusesV1NULBeforeHashing`,
  `TestValidateNamedStorePinsRefusesV1NULBeforeHashing`,
  `TestSkillsOfRefusesV1NULBeforeHashing` (envprofile, local green);
  `TestContentHashFollowsWriterVersionNULRule` (contextstore, local green).
- Mutants: M1 reorder revert — executed, killed (mutant yields a warn
  carrying a computed v1 digest). Narrowing (rev4 F3 correction: the
  rev2 claim below was wrong — revision 3's M-order mutant survives
  every error/digest-shape assertion): hashing before the refusal is
  now killed by `hashing.CountV1Hashes` observation at each guarded
  entry (zero v1 computations over a NUL tree, with a clean control
  proving the seam is wired), not by digest shape.

### Finding 3 (P2) — v2 verdicts in an unversioned carrier, digest-only cache identity

Fix: v2 verdicts are stored as schema 2 with `hash_version: 2`; v1
verdicts keep the byte-identical frozen schema-1 shape; loading
compares (version, digest) and reads absent-schema records as v1.

- Regression tests: `TestV2VerdictCarrierRecordsHashVersion`,
  `TestV2RejectsLegacyUnversionedVerdict`,
  `TestV1IgnoresV2VerdictCarrier`, `TestV1HonorsLegacySchema1Verdict`
  (audit, local green; the equal-digest-text fixtures are synthetic
  version-pair fixtures, not collision claims).
- Mutants: M2 v2-stored-unversioned — executed, killed. M3
  version-compare removed — executed, killed by both mismatch tests
  (mutant shows CacheHit=true consuming the foreign record).

## 3. Local evidence (all through ~/.local/bin/mini-build-lock, GOFLAGS=-work)

Every command below exited 0 unless noted as an intentional mutant kill
(FAIL). syspolicyd stayed running with successive crashes 26 → 26
across the whole session (checked before/after every long call).

- gofmt -l on internal/ and cmd/: clean, exit 0.
- go vet (compiles tests): audit, opaquescan, marker, closure,
  contextstore — exit 0; envprofile, install, contextaudit, hashing,
  cmd/curator — exit 0.
- go test ./internal/audit -count=1 (full package): ok 0.978s, exit 0.
- go test ./internal/marker -count=1 (full package): ok 1.672s, exit 0.
- go test ./internal/opaquescan ./internal/hashing ./internal/contextaudit
  -count=1: all ok, exit 0.
- go test ./internal/contextstore -count=1: ok, exit 0.
- go test ./internal/closure -run 'Frozen|ContentHash|Draft': ok, exit 0.
- go test ./internal/envprofile -run 'NUL|Opaque|Colliding': ok 15.963s,
  exit 0 (includes the pre-existing v1-blocking and v2-admission
  suites plus the three new guard tests).
- Mutants M1–M4: each FAILed its killing test as intended, then
  reverted (grep MUTANT = 0) and the owning package re-ran green.
- go build ./...: exit 0.
- golangci-lint run on the nine touched trees: 0 issues, exit 0.

## 4. Not run locally (hosted gate is the arbiter)

Per R193/R194 and the task brief, cmd/curator and internal/install
suites were written and compile-verified (vet exit 0) but not executed
here; scripts/remote-gate.sh exceeds headless single-call bounds and
was not run or edited. The orchestrator runs on the hosted gate:

- env GOFLAGS=-work go test ./internal/install -run 'NUL|Opaque|V1Reinstall|DraftLane' -count=1 -timeout=6m
- env GOFLAGS=-work go test ./cmd/curator -run 'NulOpaque|OpaqueNUL|StatusCheck' -count=1 -timeout=6m
- the full hosted gate (scripts/remote-gate.sh) as the final arbiter

## 5. Expectation changes to pre-existing tests (justified, not weakened)

Two envprofile path-source tests now see the AC-mandated earlier
refusal: the store content hash (a v1 compute site) refuses before
hashing, so the error carries the shared `audit.opaque.nul-byte` id
instead of the detector class the later audit layers would have
reported, and the overlay-update case reports `profile_source_invalid`
at source load instead of the member-audit update-blocked diagnostic.
Both tests still require a refused operation, the profile diag, the
opaque id, the named file, and (for update) an unmoved lock. The
git-lane suites keep the detector-class refusal unchanged
(TestInstallBlocksBothV1CollidingSkillTrees passes unmodified).

## 6. Stated bounds and review notes

- Pin carrier (`audit.Pin`/`isPinned`) — rev4 F2 correction: the rev2
  "deliberately unchanged" claim is withdrawn. Revision 3 showed the
  unversioned carrier lets a legacy v1 pin authorize a v2 read at
  equal digest text, repeating the finding-3 mechanism. Pins are now a
  versioned carrier (schema 1 frozen v1 shape authorizes v1 only;
  schema 2 carries `hash_version: 2`), `audit --allow` records the
  writer framing explicitly (no digest inference), and approval
  compares (version, digest). Legacy v1 pins keep authorizing v1
  reads; v2 identities need a fresh pin.
- Hybrid `status` rows carry states only, so a v1 NUL tree reports
  content-drift there without finding text; every other status surface
  prints the opaque refusal from the install-status error.
- Scan-then-hash races (TOCTOU) between the guard and the hash are out
  of scope, as before; the managed install transaction still pins
  admitted inputs for its per-write recheck.
- CHANGELOG Unreleased extended. LOGBOOK.md and scripts/remote-gate.sh
  intentionally untouched per host rules (verified in git status).

## 7. Files changed (rework on top of the rev1 candidate)

Production: internal/opaquescan/v1guard.go (new shared guard),
internal/audit/audit.go (reorder, versioned cache, shared id),
internal/marker/marker.go, internal/closure/resolve.go,
internal/contextstore/contextstore.go, internal/envprofile/envprofile.go,
internal/envprofile/store_boundary.go, internal/envprofile/managed.go,
internal/install/install.go, internal/install/targets.go,
cmd/curator/main.go, cmd/curator/draft_status.go, CHANGELOG.md.
Tests: internal/audit/opaque_v1_rework_test.go,
internal/opaquescan/v1guard_test.go, internal/marker/nul_opaque_v1_test.go,
internal/closure/nul_opaque_v1_test.go,
internal/contextstore/nul_opaque_v1_test.go,
internal/envprofile/nul_opaque_v1_guard_test.go,
internal/install/nul_opaque_v1_rework_test.go,
cmd/curator/nul_opaque_v1_rework_test.go; updated call sites in
internal/audit/audit_test.go and internal/envprofile/nul_opaque_test.go.
