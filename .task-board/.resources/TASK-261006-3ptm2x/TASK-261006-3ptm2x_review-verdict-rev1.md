# TASK-261006-3ptm2x — Scope the NUL opaque gate to v1 identities: revision 1 review

Verdict: **changes_requested**, route **to-dev**. No human decision or external blocker is needed. No acceptance mutation or commit acknowledgement was supplied. The run is not goal-bound (`spawn goal` returned none immediately before recording).

Reviewed base `5364c4dfad572077317b7a60d020452c861b3ffe`, candidate tree `ad30ed54e797ddc20d4a8a7d9ce2ac7e7b516e2f`, CR revision 1. Repository code was not changed; attack probes and mutants ran in a disposable clone materialized from the candidate tree.

Normative authority: [curator-spec core §8](https://github.com/relux-works/curator-spec/blob/main/protocol/core.md#8-content-hashes), fetched live. Version and digest form one identity; absent hash_version means v1 on frozen carriers. A v1 reader must scan all regular files before computing or trusting the identity, block on NUL, and refrain from hashing those bytes. V2 treats NUL as data.

## Numbered findings

1. **P1 — Legacy installed-tree currentness still trusts a v1 NUL collision.** `internal/marker/marker.go:1347` directly rehashes at the recorded version without a NUL guard. The same unguarded recorded-version computation appears in `cmd/curator/main.go:1320` (scope drift), `:2097` (hybrid status), and `cmd/curator/draft_status.go:114`. Reviewer probe `TestReviewLegacyMarkerNULCollisionCurrentness` first writes a valid schema-2 marker without hash_version on a clean two-record tree. After replacing the files with their NUL-bearing one-record v1 collision, `marker.Current` returns **true, nil**. The clean control passed. These are real equal-v1-digest fixtures, not assumed SHA collisions. Status's source audit scans source snapshots; it does not protect a separately modified installed tree. Fix every recorded-v1 currentness/verification entry to refuse NUL before hashing or comparing, retaining the opaque finding. Add production CLI `status --check` regressions for a v1 installation after NUL appears, including the collision-preserving edit, and for v2 NUL currentness. Reinstall and generic GateReadOnly tests do not substitute for those required status-entry tests.

2. **P1 — V1 identities are computed before the opaque refusal.** `internal/audit/audit.go:328` hashes before the opaqueReport block at :340. `auditSubject` scans but forwards its NUL paths to that hashing call. Production `CheckSourceAudit` (`internal/audit/sourceaudit.go:1088`) uses this helper, bypassing the safe early skip in `Gate`. A non-invasive counter immediately before v1 digest.Write(content) observed **one NUL-bearing v1 hash** during a draft source audit, followed by the ordinary opaque error (`TestReviewDraftAuditNeverComputesV1OverNUL`). Separately, `closure.ContentHashFor` (`internal/closure/resolve.go:565`) computes the frozen v1 projected-context hash without a prior guard; the probe with a valid SKILL.md and included assets/a.bin returns a digest, no error, and one NUL-bearing v1 hash. This is reached by frozen loading/acquisition before install's audit gate, and by draft resolve/refresh paths (`internal/install/draftsources.go:505,585,682`). Move refusal before any v1 hash; audit rejection must not carry a computed v1 ContentSHA256. Sweep the lock/store readers listed below too. Add negative production-entry tests that observe whether v1 hashing was reached; checking only no cache file does not prove that no v1 computation occurred.

3. **P2 — V2 audit verdicts use an unversioned carrier and digest-only cache identity.** `storeCachedFindings` (`internal/audit/audit.go:511`) now writes a v2 digest into the same schema-1 JSON with no hash_version. `trustDir`/:455, `verdictPath`/:459, and `loadCachedFindings`/:464 carry no framing version; loading reads only findings. The producer inventory's statement that the cache is keyed by a versioned digest is inaccurate. `TestReviewV2VerdictRecordsHashVersion` observes schema_version=1/hash_version absent after auditing a v2 NUL tree. `TestReviewV2RejectsUnversionedLegacyVerdict` seeds a legacy-shaped cache at equal digest text, and v2 consumes it (`CacheHit=true`, findings empty). The latter is a synthetic version-comparison fixture, not a claim of a discovered cross-version SHA-256 collision. Carry and validate the version with cache identity, interpret legacy absent versions as v1, and prevent v2 from consuming those records solely because digest text matches. Preserve v1 NUL cache exclusion. Review the similarly digest-only pin carrier when wiring this change, and add carrier/version-mismatch negative tests.

## Surface sweep

All **4/4** production NULPaths call sites were enumerated with `git grep ... -- '*.go'` (excluding board patch history):

| Site | Version source | Review |
| --- | --- | --- |
| audit gate, :203 | Subject.HashVersion; zero=v1, unknown refuses | Version-scoped before audit-disabled return; mixed v1/v2 probe passes |
| auditSubject, :310 | Same subject version | Scan scoped correctly; subsequent hash order fails finding 2 |
| audit detect, :541 | Legacy helper, canary/test callers | V2 pipeline calls detectWithOpaquePaths with no opaque paths |
| contextaudit Detect, :184 | Legacy v1 entry; DetectAtVersion dispatch | V1 blocks; v2 skips NUL; unknown refuses; targeted tests pass |

Identity compute/trust sweep (version is selected from state or shape, never digest bytes, except the cache/pin APIs omit version entirely):

| Surface | Version source and observed guard |
| --- | --- |
| Project/global install audit | Draft lock => v1; otherwise WriteVersion; shared gate guards source snapshots |
| CLI audit/external repository admission | WriteVersion; v2 admission wired explicitly |
| Draft source-audit and frozen ContentHashFor | Frozen v1; unsafe hash ordering, finding 2 |
| Install registry resolution, install staging | Draft Package => v1, otherwise writer; downstream source gate normally protects admitted source; does not repair frozen precomputation |
| marker.Write | Core writer v2; frozen Package shape v1; v1 write trusts supplied digest without its own NUL check |
| marker.Current and three CLI drift/status hash sites | Recorded marker version; no preceding installed-tree NUL scan, finding 1 |
| contextstore.ContentHash/EnsureState | Writer version; v1 state hashing occurs before later member auditing |
| envprofile auditMember/update strict audit | Resolved lock version; DetectAtVersion before strict member hash; correct version wiring |
| envprofile migrateGlobalSkills | Explicit v1; scan before strict audit; historical migration lane remains guarded |
| envprofile store_boundary pin verification / managed.skillsOf | Lock version; direct hash calls lack a local v1 NUL guard; producer should cover these read/repair trust surfaces, not assume prior admission persists |
| envprofile identity migration | Verifies old lock/store first, then v2 rehash; old-v1 trust inherits store-boundary gap |
| Registry identity comparisons | Explicit framing version/frozen v1 carrier rules; callers remain responsible for snapshot scanning |
| Verdict and pin state | Bare digest keys and unversioned records; finding 3 |
| Snapshot inventory/build receipts/materialized surface hashes | Distinct identity schemes; not used to guess §8 skill/context version |

The producer listed the four scanner sites and the major v1 sites, but treating the latter as unchanged did not establish compliance with the task's stronger pre-hash/pre-trust requirement.

## Evidence and bounds

Hosted run [37448065494](https://github.com/relux-works/curator/actions/runs/37448065494) is completed/success. Independently queried head `1f6841a1dca404fcd43b622ef4d91856651e549d`; its tree is exactly the candidate. Accepted hosted evidence for CLI/install/envprofile tests and lint/build. The attached validation log's measured command-shard coverage is **1/1**; individual hosted test-case coverage is unknown. No prohibited executable-producing packages were run locally.

Locally reran only audit, contextaudit, opaquescan and hashing through `mini-build-lock run opaque -- env GOFLAGS=-work go test ./internal/<pkg> -run <regex> -count=1 -timeout=6m`, in bounded sequential calls. Restored audit baseline including all five new version tests passes; contextaudit version/NUL tests and hashing collision/version tests pass. Opaquescan has no test files, not counted as behavioral coverage. Every measured call found syspolicyd running and successive crashes **26 -> 26**. Exact per-call durations and outputs are in `TASK-261006-3ptm2x_review-checks-rev1.jsonl`.

Reviewer attack coverage: **5/5** requested shapes examined (legacy marker, mixed tree, status currentness primitive after a NUL collision edit, explicit v2-to-v1 marker downgrade, v1 digest cache twin). Security expectations hold for **3/5**; legacy/currentness fail. CLI status itself was not locally executed per R194; its hash sites were inspected and the marker primitive reproduced. Seven additional synthetic probes have **2/7** passes and **5/7** failures, grouped into the three findings. An existing reviewer-policy-record test also matched the regex and passed; excluded from this seven-probe ratio.

Valid mutants killed: **9/9**. This is the local pure-package bound; install/CLI/profile wiring mutants were not run on this host, so their mutation coverage is unknown.

| Mutant | Killing test observed |
| --- | --- |
| M1 unconditional NUL scan | TestGateV2AdmitsNULBearingTreeWhenAuditIsDisabled |
| M2 remove v1 gate | TestOpaqueNULBlocksEvenWhenAV1TwinHasCachedAllow; disabled-audit deep NUL; v1 finding test |
| M3 zero version becomes v2 | TestGateV1KeepsOpaqueBlockWithUnchangedFinding/legacy-zero |
| M4 unknown version becomes v2 | TestGateRejectsUnknownHashVersion |
| M5 audit hash always v1 | TestAuditSubjectWithOpaquePathsIgnoresNULPathsForV2 identity assertion |
| M6 preserve supplied opaque paths for v2 | Same test's admission assertion |
| M7 cache blocked v1 NUL result | TestGateV1KeepsOpaqueBlockWithUnchangedFinding cache assertion |
| M8 narrow v1 scan to assets only | TestGateBlocksDeepOpaqueNULWhenAuditIsDisabledAndNamesFile (docs tree) |
| M9 contextaudit v2 dispatches to v1 scan | TestDetectAtVersionAppliesNULRuleOnlyToV1 |

Two non-compiling mutant prototypes were corrected and rerun; excluded from coverage. A malformed initial frozen-context probe was corrected with a valid required SKILL.md and an included asset; only the corrected reproduction counts. All disposable mutations were restored. No workspace source changes, LOGBOOK edits, remote-gate.sh edits, commits, branch changes, or integrations were performed. CHANGELOG Unreleased is present. The documented bare-v1-revocation cutover is visible, but it does not address the three findings.

Required next cycle: producer fixes findings 1–3, expands the actual CLI/install/status and lock-reader negative tests, reruns hosted validation on the new candidate, updates its inventory/evidence with measured coverage, and hands off a new revision for review.
