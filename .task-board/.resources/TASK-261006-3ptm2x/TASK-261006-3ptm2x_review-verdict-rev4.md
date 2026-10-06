# TASK-261006-3ptm2x — Scope the NUL opaque gate to v1 identities: revision 4 review

Verdict: **changes_requested**, route **to-dev**. One P1 bypass remains. Revision-3 findings F1, F2 and F3 are fixed; this is ordinary implementation rework within the same task, with no external blocker or human-only decision.

Reviewed CR-TASK-261006-3ptm2x-4 revision 4, base `5364c4dfad572077317b7a60d020452c861b3ffe`, candidate tree `3db2c9b9c63419a2731a283c4369d8787c33006d`. Fresh upstream main remains the base. The run is not goal-bound (`spawn goal` returned none). The repository candidate was left unchanged, including all 37 changed paths. No commit, acknowledgement, integration or LOGBOOK edit was performed.

Normative authority: [curator-spec core §8](https://github.com/relux-works/curator-spec/blob/210d3619051248cf9b87a7bfe02e5a9142a6b7cf/protocol/core.md#8-content-hashes), fetched at freshly resolved main. A v1 reader must inspect the full skill/context snapshot before computing or trusting its content identity. Acquisition through Git does not exempt a later v1 content computation. V2 accepts NUL as file data; framing comes from the carrier or frozen shape, and comparisons include framing version.

## Numbered finding

1. **F1 / P1 / bypass — Commit-pinned v1 context re-checks skip the full-snapshot opaque guard and compute/trust v1 identities over NUL.** `internal/envprofile/store_boundary.go:211` returns from the `member.Commit` branch before the guard at :221. The comment correctly recognizes the Git tree OID as a separate algorithm, but the same snapshot then reaches `Resolve -> verifyHome -> assembleHome -> loadMaterial -> rootContext -> homePlan.surfaceHash`. `loadMaterial` (`switch.go:629`) has no version guard, and `surfaceHash` (`managed.go:400`, reached at :514 for monolithic output) invokes the core content framing with the recorded v1 lock version.

   `TestReviewV1CommitContextResolveRefusesNUL` drives the exported production `envprofile.Resolve` reader under the current v2 writer, with a canonical legacy schema-1 lock, a real commit-pinned context snapshot, and a valid schema-1 managed-home marker. Git integrity succeeds; there is no fixture hash mismatch. Two NUL shapes both return a non-empty launch fragment with **nil error**, reporting the home current:

   - `module`: `context/a.md` contains `x\x00y\n`. The maintained counter observes two v1 computations, and independent file-set instrumentation observes **one v1 hash whose input contains NUL**.
   - `excluded`: the module is clean but `assets/deep/a.bin` contains NUL. Resolve omits that file from rendering without the required full-snapshot opaque refusal, then trusts the v1 surface identity.

   The clean v1 control and both NUL-bearing v2 controls return launch fragments with nil error. V2 controls compute **zero** v1 identities. The attack expectation holds in **3/5** final subtests; the two required v1 refusals fail. This is a synthetic, internally consistent legacy-state attack, not a claim that an ordinary prior v1 install admitted these bytes.

   Required rework: guard the full context package snapshot according to the lock's content framing before any downstream v1 materialized identity is computed or trusted, including commit-pinned contexts and excluded files. Preserve the Git object-integrity check and v2 admission. Sweep other materialization/read entry points that share this path. Add a maintained production `Resolve` regression with the two NUL shapes, clean v1/v2 controls, and direct observation of refusal before hashing. Do not infer framing from the commit or digest.

   Narrowing mutant for that regression: enforce the full-snapshot guard only for state-pinned members, skipping `member.Commit != ""`. As a diagnostic in the disposable copy, a version-aware full-snapshot guard before the commit branch makes **5/5** subtests pass: both v1 NUL inputs refuse with `audit.opaque.nul-byte` and zero v1 computations; clean v1 and NUL-bearing v2 remain current. The diagnostic was restored and is not a repository fix.

   **repeat-of: none.** Revision-3 F1's filtered frozen projection is fixed. This bypass chooses scan eligibility by the acquisition pin algorithm and misses a later v1 content computation; it is a distinct mechanism. The general full-snapshot invariant recurs, so rework must include this omitted reader class within this leaf.

## Prior findings and required attacks

- Rev3 F1: full frozen snapshot scan now precedes filtering. Independent frozen helper probes cover runtime, build and unlisted roots; an exported `ResolveDraft` probe refuses excluded NUL. Moving the scan back to the projection kills these probes. Maintained resolve/refresh tests and prior-lock preservation are present and covered by the exact-tree hosted suite.
- Rev3 F2: `PinAtVersion` writes schema 2/hash_version 2 for v2, legacy `Pin` remains the explicit v1 writer, and CLI `audit --allow` passes the writer version. Both mismatch directions refuse; matching versions authorize. The previous versionless writer probe was updated to call the explicit v2 writer. Ignoring the pin version kills maintained regressions.
- Rev3 F3: the new counter catches discarded pre-refusal hashes in audit pipeline and gate. The corresponding frozen-path mutant is killed by the reviewer observation. Clean controls prove the counter is wired. Old empty-returned-digest tests are no longer claimed to prove ordering.
- Requested attacks **5/5 examined**: legacy marker collision, mixed-version audit subjects, installed status after NUL appears, v2-to-v1 marker downgrade, and v1 verdict-cache collision twin. Their maintained expectations hold. CLI/install execution is reused from hosted evidence, not a local replay.

## Surface sweep

Scanner inventory is **5/5 production scan calls**: audit gate (:205), auditSubject (:312), legacy detect/canary (:632), contextaudit's v1 Detect (:184), and the unqualified call in `opaquescan.RefuseNULV1`. `git grep opaquescan.NULPaths` over production sources was cross-checked with the new wrapper and candidate files. No version choice inspected uses digest bytes.

“Held” below means the inspected version propagation and available regression evidence hold; it does not claim exhaustive dynamic coverage for that row.

| Surface | Version source / result |
| --- | --- |
| Audit gate and pipeline | Subject.HashVersion; zero=v1; unknown refuses. Held, local baseline and order mutants |
| Legacy detect/canary | Frozen v1 helper; v2 pipeline passes no opaque paths. Held |
| Shared RefuseNULV1 | Explicit caller version; v1 scans, v2 skips, unknown refuses. Held |
| Contextaudit version dispatch | Legacy entry or recorded lock version. Held, version-narrowing mutant killed |
| Project/global/CLI/external admission | Draft=v1; otherwise writer version. Held, hosted production tests |
| Draft source audit | Explicit frozen v1; refusal precedes hash/cache/pin trust. Held |
| Frozen resolve/refresh/projection | Frozen v1; full pre-projection scan. Held |
| Install registry and staged hashing | Draft Package=v1; otherwise writer version. Held |
| Recorded marker currentness | Recorded version or frozen shape. Held; legacy collision/downgrade probes |
| CLI drift/hybrid/draft status | Recorded marker version. Held; guard precedes rehash |
| Project/global status plans | Recorded marker version. Held; project/global helper, hosted status stderr regression |
| Marker writer versions | Explicit v2 writer, frozen Package=v1. Held; upstream staging guard |
| Context store writer hashing | WriteVersion; pre-hash guard. Held |
| Strict profile audit/update | Resolved lock version; DetectAtVersion and hash guard agree. Held |
| Profile store state-pin verification | Lock version; guard precedes core hash. Held |
| Profile store commit-pin verification and v1 context reads | Git object check skips scanner; later lock=v1 context hash. **Broken, F1** |
| Managed skill loader | Lock version; guard before skill hash even for commit members. Held |
| Profile identity migration | State-pin boundary inspected; separate migration transaction not dynamically attacked this round |
| Global skill migration | Explicit v1 contextaudit before strict hash. Held |
| Registry/source-audit comparisons | Explicit recorded version or frozen v1 shape; prior snapshot admission. Held |
| Verdict cache carrier/version/exclusion | Schema-1=v1, schema-2 records version. Held; NUL refusal before lookup/store |
| Trust pin carrier/version comparison | Versioned writer and explicit subject version. Held |
| Revocations | Existing bare digest/source list; writer-version identity, no v1 fallback hashing. Held; cutover action documented |
| Materialized context surface hashes | Lock version, manager-rendered file set. **Broken for commit-pinned v1 contexts, F1** |
| Other identity schemes | Git OIDs, snapshot/build identities and plain fingerprints are separate algorithms. Held as algorithm classification; not exemptions for downstream core hashing |

The production core content-hash calls were also swept repository-wide, including `contextmaterialize.SurfaceHashWithVersion`. Its previously stated “upstream admission protects rendering” bound is insufficient for a recorded-v1 re-check of commit-pinned context state; the new probe establishes the missing case. Concurrent scan/hash mutation and exhaustive rendered-surface coverage remain unestablished.

## Evidence and bounds

Hosted [run 37480146753](https://github.com/relux-works/curator/actions/runs/37480146753) independently reports completed/success, head `69c810fc9b322c263bb38ad1011dd46edc6e1237`, whose tree is exactly the candidate. Reused its CLI/install/profile, lint, build, cross-platform and race evidence. The attached CR validation log reports configured command-shard coverage **1/1**; individual hosted test-case coverage remains unknown. No cmd/curator or install test was run locally.

Local checks ran in a disposable clone through `mini-build-lock run opaque -- env GOFLAGS=-work go test ./internal/<pkg> -run <regex> -count=1 -timeout=6m -v`, only for audit, contextaudit, opaquescan and hashing. Reviewer probes in audit drive exported closure, marker and profile APIs. The profile probe uses real Git with an empty init template and only non-executable data fixtures; it creates no fake executables or executable hooks. Source instrumentation records NUL input to the v1 in-memory hasher, without altering its result.

All 22 bounded calls observed syspolicyd running and successive crashes **26 -> 26**, total 44.844 seconds. The attached sanitized JSONL records each mask, duration, exit and output. Initial profile prototypes showed missing surfaces and then a missing modern credential fixture; those intermediate control failures are excluded from the final behavioral ratio. The final corrected controls and diagnostic establish the finding.

Valid mutation coverage is **11/11 killed**: **8/8 by maintained tests**, **3/3 by independent reviewer probes**.

| Mutant | Observed killing test |
| --- | --- |
| Frozen scan moved to projection | Reviewer frozen runtime/build/unlisted and ResolveDraft probes |
| Pin version ignored | TestV2RejectsLegacyV1Pin, TestV1RejectsV2Pin, TestLegacySchemalessPinReadsAsV1 |
| Discarded v1 hash inside pipeline refusal | TestAuditSubjectV1NULRefusalComputesNoV1Identity, TestCheckSourceAuditV1NULRefusalComputesNoV1Identity |
| Discarded v1 hash inside gate refusal | TestGateV1NULRefusalComputesNoV1Identity |
| Discarded v1 hash inside frozen refusal | Reviewer frozen zero-counter probe |
| Cache version ignored | TestV2RejectsLegacyUnversionedVerdict, TestV1IgnoresV2VerdictCarrier |
| V2 cache written in frozen v1 carrier | TestV2VerdictCarrierRecordsHashVersion |
| Opaque scan applied to v2 | TestGateV2AdmitsNULBearingTreeWhenAuditIsDisabled |
| Legacy zero subject treated as v2 | TestGateV1KeepsOpaqueBlockWithUnchangedFinding/legacy-zero |
| Context v2 dispatched to v1 detector | TestDetectAtVersionAppliesNULRuleOnlyToV1 |
| Recorded-marker v1 guard narrowed away | Reviewer legacy-marker collision probe |

All production mutations and instrumentation were restored and byte-compared to the candidate. Restored audit/contextaudit subsets pass. The original workspace's 37/37 changed paths equal the candidate blobs. CHANGELOG Unreleased is present; no LOGBOOK.md or scripts/remote-gate.sh delta; whitespace checks pass. Important findings are recorded on the board instead of LOGBOOK, as the binding host rule requires.

Portable probes and instrumentation are attached in `TASK-261006-3ptm2x_review-attack-probes-rev4.patch`; apply only to a disposable exact-candidate copy. The final profile controls were executed on macOS; the patch's modern credential setup needs the corresponding native fixture if replayed on Linux. This is an explicit portability bound, not evidence that a Linux replay has passed.

## Machine-readable finding set

```verdict-findings
{
  "findings": [
    {
      "id": "v1-commit-context-bypass",
      "row": "Profile store commit-pin verification and v1 context reads",
      "invariant": "Before computing or trusting any v1 content identity, inspect every regular file of the full context snapshot for NUL and refuse before hashing; v2 accepts NUL as data.",
      "mechanism": "storeEntryPinHashes returns from its member.Commit branch before RefuseNULV1; loadMaterial has no version guard, so Resolve -> verifyHome -> assembleHome -> rootContext -> surfaceHash computes and trusts a v1 identity from NUL-bearing commit-pinned context bytes.",
      "reproductions": [
        {
          "test_file": "internal/audit/review_profile_surface_test.go (attached reviewer patch)",
          "command": "mini-build-lock run opaque -- env GOFLAGS=-work go test ./internal/audit -run '^TestReviewV1CommitContextResolveRefusesNUL$' -count=1 -timeout=6m -v",
          "expected_failure": "module and excluded subtests return a launch fragment with nil error instead of the opaque refusal; module observes one v1 file-set hash containing NUL. Clean v1 and both NUL-bearing v2 controls pass.",
          "pinned_blobs": [
            "sha256:a2c8f671101bd866d7cffbc3cd6192121a169a7c13521df8f4e99679b4c8497b",
            "sha256:abc0c24bd9d84aca0a6b0523342ce99792074ca1a7a0821b91e5185e459df08f",
            "sha256:fdc32b638d18b35d19ae9d1a49d90ab8a4653ab95da57db36238ba0dd743c844",
            "sha256:d430dd3c0550e6bb92968a06d3da41a1f79f35d1acacaac1c15b5b2ba670a53e"
          ]
        }
      ],
      "severity": "bypass",
      "repeat-of": "none"
    }
  ],
  "notes": [
    "Revision-3 F1/F2/F3 are fixed. The new finding skips the scan by pin algorithm, whereas rev3/F1 scanned a filtered frozen projection; same general invariant, different bypass mechanism.",
    "Exhaustive rendered-surface and concurrent scan/hash mutation coverage remain unestablished."
  ],
  "surface_results": [
    {
      "row": "Audit gate and pipeline",
      "result": "held"
    },
    {
      "row": "Legacy detect/canary",
      "result": "held"
    },
    {
      "row": "Shared RefuseNULV1",
      "result": "held"
    },
    {
      "row": "Contextaudit version dispatch",
      "result": "held"
    },
    {
      "row": "Project/global/CLI/external admission",
      "result": "held"
    },
    {
      "row": "Draft source audit",
      "result": "held"
    },
    {
      "row": "Frozen resolve/refresh/projection",
      "result": "held"
    },
    {
      "row": "Install registry and staged hashing",
      "result": "held"
    },
    {
      "row": "Recorded marker currentness",
      "result": "held"
    },
    {
      "row": "CLI drift/hybrid/draft status",
      "result": "held"
    },
    {
      "row": "Project/global status plans",
      "result": "held"
    },
    {
      "row": "Marker writer versions",
      "result": "held"
    },
    {
      "row": "Context store writer hashing",
      "result": "held"
    },
    {
      "row": "Strict profile audit/update",
      "result": "held"
    },
    {
      "row": "Profile store state-pin verification",
      "result": "held"
    },
    {
      "row": "Profile store commit-pin verification and v1 context reads",
      "result": "broken"
    },
    {
      "row": "Managed skill loader",
      "result": "held"
    },
    {
      "row": "Profile identity migration",
      "result": "not-attacked",
      "detail": "Old-lock state boundary inspected; separate migration transaction not dynamically attacked this round."
    },
    {
      "row": "Global skill migration",
      "result": "held"
    },
    {
      "row": "Registry/source-audit comparisons",
      "result": "held"
    },
    {
      "row": "Verdict cache carrier/version/exclusion",
      "result": "held"
    },
    {
      "row": "Trust pin carrier/version comparison",
      "result": "held"
    },
    {
      "row": "Revocations",
      "result": "held"
    },
    {
      "row": "Materialized context surface hashes",
      "result": "broken"
    },
    {
      "row": "Other identity schemes",
      "result": "held"
    }
  ],
  "free_hunt": [
    "Commit-pinned v1 context re-check bypass demonstrated through envprofile.Resolve."
  ]
}
```
