# TASK-261006-3ptm2x — Scope the NUL opaque gate to v1 identities: revision 7 review

Verdict: **accepted**. No findings. Revision 7 fixes the previous `v1-commit-context-bypass`; the complete prior surface table was swept again. Acceptance routes the task to **integrating**, not done. No repository code, LOGBOOK, gate script, commit, acknowledgement or integration was changed by this reviewer.

Reviewed CR-TASK-261006-3ptm2x-7, base `5364c4dfad572077317b7a60d020452c861b3ffe`, candidate tree `a3a1e98329723729ba5d30fa8f0fbd95a7f94ab4`. Fresh upstream main equals the base. `spawn goal` reports that this reviewer run is not goal-bound. All **39/39** changed paths in the original workspace and the restored disposable copy match candidate blobs.

Normative source: [curator-spec core §8](https://github.com/relux-works/curator-spec/blob/210d3619051248cf9b87a7bfe02e5a9142a6b7cf/protocol/core.md#8-content-hashes), read through the primary GitHub API after fresh main resolution. Recorded framing or a frozen carrier determines the version; content comparisons include the version. The v1 guard inspects all regular files before computing or trusting a content identity; v2 admits NUL as file data. No inspected version decision infers framing from digest bytes.

## Previous verdict answered

**Revision 4 / v1-commit-context-bypass: fixed.** `storeEntryPinHashes` now scans the full store entry at the lock's version before the commit/state branch split (`store_boundary.go:222`). Expected-pin lookup remains first and computes only a Git object lookup or a pin string, preserving missing-object repair diagnostics without computing a core v1 identity. Commit integrity still uses Git tree OIDs. The later materialized content identity follows the lock version.

`loadMaterial` adds an independent full-entry guard before reading manifest/module data (`switch.go:644`). This protects `assembleHome`, `materializeScope`, and `preflightInPlaceScope`, including paths without prior pin verification. `skillsOf` guards the selected skill root before lock-version hashing. MCP manifest loading has no core content-hash computation. State-pin, extraction, rebuilding, publication, migration and status siblings were inspected against the producer's **16/16** branch inventory; contextlock hashes canonical lock bytes, not core tree framing.

The maintained production `envprofile.Resolve` regression is observed green: both legacy v1 NUL shapes refuse at `pin_hash` with the stable opaque finding and zero v1 computations; clean v1 emits a fragment and exercises the counter; both v2 NUL controls provision and re-resolve successfully with zero v1 computations. That is **5/5** Resolve shapes. The switch-materialization regression is **3/3**: module NUL and excluded NUL refuse before hashing, clean v1 materializes.

Revision 7's fixture change is appropriate: one native home, launch directory and XDG directory are reused across provision and bare Resolve; a synthetic live credential target supports Linux's file-link passthrough. Product code is unchanged from the bypass fix. Exact-tree hosted Ubuntu evidence confirms the three previously failing fixture leaves no longer fail; local macOS success alone is not presented as Linux proof.

## Required attack shapes

All **5/5** requested classes were examined:

| Attack | Evidence / outcome |
| --- | --- |
| Legacy marker without hash_version over a v1 NUL collision | Independent production `marker.Current` probe refuses; clean legacy and v2 controls pass |
| Mixed tree/version subjects | Maintained `TestGateMixedVersionsRefuseOnlyTheV1Subject` passes locally |
| Installed status after NUL appears | Maintained CLI `TestStatusCheckRefusesV1NULAppearingAfterInstall` and install collision/fresh-file tests inspected; exact-tree hosted suite reused |
| V2-to-v1 downgrade | Independent production marker-reader probe rejects a v5 marker changed to hash_version 1 |
| Verdict cache keyed on a v1 digest | Maintained v1 NUL/cache tests pass locally; refusal precedes lookup/storage, and cross-version cache mutants fail |

Production CLI v2 install/status/audit regressions inspect the v2 marker and currentness. Project/global installation tests also recompute the recorded v2 identity. Their execution comes from the hosted gate, not a prohibited local run. Both clean and NUL-bearing controls are present. Zero-computation tests observe actual hashing calls, rather than merely an empty returned digest.

## Complete surface sweep

Scanner inventory: **5/5 production calls** inspected. `git grep opaquescan.NULPaths` identifies audit gate, auditSubject, legacy detect/canary and contextaudit Detect; `rg` additionally confirms the unqualified wrapper call in new `opaquescan.RefuseNULV1`. Repository-wide core tree/file-set hash calls and Subject construction sites were cross-checked. These five scan sites are:

- `audit.go:205`: Subject version resolved first; zero=v1; v2 skips; unknown refuses.
- `audit.go:312`: same Subject dispatch before the pipeline hash.
- `audit.go:632`: frozen legacy detector/canary helper; v2 pipeline uses `detectWithOpaquePaths` without NUL findings.
- `contextaudit.go:184`: frozen `Detect` is v1; `DetectAtVersion` dispatches v2 to its non-opaque path.
- `opaquescan/v1guard.go:27`: explicit caller version; v1 scans, v2 skips, unknown refuses.

**25/25 prior surface rows inspected.** “Held” describes the inspected production control flow and cited evidence; it does not mean every branch or interleaving was dynamically tested.

| Surface | Version source and review result |
| --- | --- |
| Audit gate and pipeline | Subject.HashVersion; zero=v1, unknown refuses; guard before hash/cache. Held, local tests and order mutants |
| Legacy detect/canary | Frozen v1 helper; v2 subject pipeline supplies no opaque paths. Held |
| Shared RefuseNULV1 | Explicit version; full regular-file scan for v1. Held, local version tests |
| Contextaudit version dispatch | Legacy entry or resolved lock version. Held, local version-narrowing mutant |
| Project/global/CLI/external admission | Draft=v1; otherwise writer version; every production Subject site checked. Held, hosted entry tests |
| Draft source audit | Explicit frozen v1; live refusal before hash and stored trust. Held, local source-entry counter |
| Frozen resolve/refresh/projection | Frozen v1; full snapshot scanned before whitelist/runtime/build filtering. Held, independent ResolveDraft and excluded-root probes |
| Install registry and staged hashing | Draft Package=v1; otherwise writer version; pre-hash guards and upstream full-snapshot admission. Held, hosted tests |
| Recorded marker currentness | Recorded version or frozen schema. Held, independent collision/downgrade probe |
| CLI drift/hybrid/draft status | Recorded marker version; guard before rehash. Held, hosted status regression |
| Project/global status plans | Recorded marker version; guard consulted for OperationStatus. Held, hosted entry regressions |
| Marker writer versions | New core v5=v2; frozen Package=v1. Held, upstream guarded staging |
| Context store writer hashing | WriteVersion; guard before hash. Held, inspected; hosted maintained store tests |
| Strict profile audit/update | Resolved lock version; audit and revocation identity use it consistently. Held, local counter tests |
| Profile store state-pin verification | Lock version; full entry guard before state hash. Held, local guarded-reader controls |
| Profile store commit-pin verification and v1 context reads | Lock version; guard before Commit return, independent material loader guard. Held, fixed prior bypass and narrowing mutants |
| Managed skill loader | Lock version; package-root guard before skill hash. Held, local guarded-reader controls |
| Profile identity migration | Old state validated through guarded store boundary; new state explicitly v2. Held by inspection and hosted suite; migration-specific adversarial replay not performed this round |
| Global skill migration | Explicit v1 contextaudit before later strict audit. Held by inspection |
| Registry/source-audit comparisons | Explicit version or frozen v1 shape; upstream snapshot guard before recorded trust. Held |
| Verdict cache carrier/version/exclusion | Frozen schema-1=v1; schema-2 records v2; version compared, digest keyed by path; v1 NUL refuses before cache. Held, local carrier and cache mutants |
| Trust pin carrier/version comparison | Explicit PinAtVersion; legacy pins=v1; version compared, digest keyed by path; CLI passes writer version. Held, local mismatch tests and hosted CLI writer test |
| Revocations | Writer/lock-version identity; no fallback v1 hash on v2 path; existing bare-list cutover documented. Held by inspection |
| Materialized context surface hashes | Lock-versioned file sets; full source guard before rendering/hash; skill hashes separately guarded. Held, Resolve and switch tests, prior bypass mutant |
| Other identity schemes | Git OIDs, plain file/lock hashes, snapshot/build identities are distinct algorithms; none exempts downstream v1 core hashing. Held by classification |

## Verification and mutation evidence

Independently verified [hosted run 37509619107](https://github.com/relux-works/curator/actions/runs/37509619107): completed/success, head `19488bff34a010e15fb86d1fca44c841bddfbef8`; its Git tree equals the candidate. Required non-skipped job coverage is **20/20 success**; two optional jobs are skipped. The configured command-shard evidence is **1/1 green**. Individual hosted test-case coverage remains unknown. Reused hosted CLI/install, full-suite, lint, vet/build, cross-platform, interop and race evidence. No CLI, install or fake-executable test was run locally.

Local tests ran sequentially in a disposable exact-candidate clone through `mini-build-lock run opaque -- env GOFLAGS=-work go test ./internal/<pkg> -run <mask> -count=1 -timeout=6m -v`. Packages: audit, contextaudit, opaquescan, hashing and envprofile (explicitly allowed by the rework-3 instruction). Targeted masks exclude the executable-writing vendor fixtures. Profile fixtures use real Git; after the initial baseline, an empty Git template prevents sample hooks. Inputs are synthetic test material, not another operator's product material.

**25 bounded calls**, **64.563 seconds total**. Syspolicyd was running before/after each, successive crashes **26 -> 26 throughout**. All baseline/restoration calls pass. The attached sanitized JSONL records masks, outputs, exits, durations and health counts.

Valid mutation coverage: **13/13 killed**, **11/13 by maintained tests**, **2/13 by independent reviewer probes**. Two preliminary materialization mutations failed compilation due to an unused import; they are explicitly marked invalid and excluded. Corrected, compiling versions were rerun and killed behaviorally.

| Mutant | Observed killing test |
| --- | --- |
| Move full-entry pin guard below Commit return | TestResolveRefusesV1CommitPinnedContextNUL; both shapes fail pin-check routing |
| Above mutant plus remove material-loader guard (full old bypass) | Same Resolve test; both shapes return fragments/nil error and compute two v1 identities |
| Remove material-loader guard alone | TestSwitchMaterializeRefusesV1CommitPinnedContextNUL; both NUL shapes materialize and compute one v1 identity |
| Compute/discard v1 hash before pipeline refusal | TestAuditSubjectV1NULRefusalComputesNoV1Identity and TestCheckSourceAuditV1NULRefusalComputesNoV1Identity |
| Compute/discard v1 hash before gate refusal | TestGateV1NULRefusalComputesNoV1Identity |
| Ignore pin version | TestV2RejectsLegacyV1Pin, TestV1RejectsV2Pin, TestLegacySchemalessPinReadsAsV1 |
| Ignore cache version | TestV2RejectsLegacyUnversionedVerdict and TestV1IgnoresV2VerdictCarrier |
| Write v2 cache in frozen v1 carrier | TestV2VerdictCarrierRecordsHashVersion |
| Scan/block v2 audit subjects | TestGateV2AdmitsNULBearingTreeWhenAuditIsDisabled |
| Read legacy zero Subject as v2 | TestGateV1KeepsOpaqueBlockWithUnchangedFinding/legacy-zero |
| Route v2 contextaudit to v1 Detect | TestDetectAtVersionAppliesNULRuleOnlyToV1 |
| Move frozen scan to filtered projection | Independent frozen runtime/build/unlisted and production ResolveDraft probes |
| Narrow recorded-marker guard away from v1 | Independent legacy collision/currentness probe |

All mutation edits were byte-restored. Restored audit/profile/contextaudit masks and independent probes pass. Original workspace and restored disposable-copy checks both match all 39 changed candidate blobs. `git diff --check` passes. CHANGELOG Unreleased is updated; no LOGBOOK.md or scripts/remote-gate.sh delta. Important review observations are attached to the board instead of LOGBOOK, following the binding host rule.

Bounds: this review establishes the listed snapshot/version paths and named negative shapes. Concurrent mutation between scan and hash and exhaustive generated-surface interleavings are not dynamically established. Raw v1 hashing primitives intentionally remain available for compatibility and collision fixtures; production callers carry the pre-hash guards. No unresolved finding is inferred from those stated bounds.

## Machine-readable verdict

```verdict-findings
{
  "findings": [],
  "notes": [
    "Revision-4 v1-commit-context-bypass fixed and independently mutation-tested.",
    "13/13 valid mutants killed; two build-invalid preliminary attempts excluded.",
    "25/25 surface rows inspected; hosted exact candidate tree green.",
    "Migration-specific attack replay and concurrent scan/hash mutation not established this round."
  ]
}
```
