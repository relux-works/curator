# TASK-261006-3ptm2x — Scope the NUL opaque gate to v1 identities: revision 3 review

Verdict: **changes_requested**; route **to-dev**. These are implementation and regression-test changes, with no external blocker or human-only decision. No acceptance, commit acknowledgement, integration, or code changes were performed. The run is not goal-bound (`spawn goal` returned none).

Reviewed CR-TASK-261006-3ptm2x-3 revision 3, base `5364c4dfad572077317b7a60d020452c861b3ffe`, candidate tree `48cba51636dbd1adf22cdefb91ef2277f594f9db`. Fresh `git ls-remote` confirms main remains the base, so there is no upstream overlap. Mutations and probes ran in a disposable clone of this exact tree.

Normative authority: [curator-spec core §8](https://github.com/relux-works/curator-spec/blob/210d3619051248cf9b87a7bfe02e5a9142a6b7cf/protocol/core.md#8-content-hashes), fetched from freshly resolved main. V1 readers inspect every regular file in the snapshot and refuse NUL before computation or trust, without omitting the offending file. V2 treats NUL as data. New content-hash carriers record version 2, and identity comparisons include the version even when digest text is equal.

## Numbered findings

1. **F1 / P2 — The frozen-v1 guard scans an already filtered projection.** `closure.ContentHashFor` copies the whitelist, subtracts runtime/build roots, then scans `destination` at `internal/closure/resolve.go:569`. It never inspects the omitted regular files in the frozen skill snapshot. `TestReviewFrozenV1RefusesExcludedNUL` returns a digest and nil error for all three shapes: NUL under an explicitly excluded runtime root, under an excluded build root, and under a non-whitelisted root. More importantly, `TestReviewResolveDraftRefusesExcludedNUL` drives the exported production resolve entry with a valid local schema-4 skill and NUL in `assets/runtime/deep/a.bin`; ResolveDraft succeeds and returns a frozen v1 lock member. This is a scope violation, not a claim that the projected digest itself contains NUL. The later install source audit is not an earlier resolve refusal. Scan the full frozen skill snapshot before filtering or producing its v1 context identity, retain the projection for hashing, and add resolve/refresh tests including prior-lock preservation. Narrowing mutant: move the full scan back to `destination`. **repeat-of: none** — revision 1 finding 2 was missing pre-hash protection of included bytes; this is a distinct omission through projection.

2. **F2 / P2 — Trust pins still use the unversioned v1 carrier for v2 approval.** `audit.Pin` (`audit.go:432`) writes schema 1 without hash_version; `isPinned` (`:451`) reads only `pinned`; `decideWithPins` (`:383`) supplies only digest text. CLI `audit --allow` reaches this production writer (`cmd/curator/main.go:2150`). A new v2 approval has no version (`TestReviewNewV2PinCarrierVersion`). A seeded legacy-shaped pin at equal digest text authorizes a strict schema-2 v2 subject over a NUL tree (`TestReviewV2RejectsLegacyPin`: Gate returns no errors). This is the same synthetic equal-text/version-pair technique used by the accepted cache regression, not evidence of a discovered SHA-256 collision. Cryptographic improbability does not satisfy §8's explicit version comparison requirement. Carry the selected version through pin writing and reading, keep legacy absent-version pins v1-only, and test the CLI writer plus mismatched-version refusal and matching-version controls. Update the validation note's claim that the unchanged pin carrier is acceptable and CHANGELOG's claim that old pins automatically re-prompt. Narrowing mutant: ignore the recorded pin version. **repeat-of: revision 1 finding 3** by mechanism (unversioned carrier and digest-only content trust), now reproduced at the pin surface that the prior verdict explicitly asked the producer to review. The verdict-cache part of that prior finding is fixed.

3. **F3 / P2 — The maintained tests do not establish refusal before hashing.** The current production audit code is correctly ordered. However, its negative evidence is insufficient for the task's no-v1-computation requirement and the previous review's explicit request to observe hashing. The valid M-order mutant inserts an unused `ContentSHA256WithVersion(..., VersionV1)` immediately inside the opaque refusal branch, preserving the empty report digest, refusal message and cache exclusion. All three maintained tests pass: `TestAuditSubjectV1NULBlockCarriesNoComputedIdentity`, `TestCheckSourceAuditV1NULRefusesWithoutVerdict`, and `TestGateV1KeepsOpaqueBlockWithUnchangedFinding`. A reviewer-only counter immediately before v1 digest consumption observes one NUL-bearing hash through production `CheckSourceAudit` and fails this mutant; the same probe passes on the restored candidate. The closure test's empty returned digest is likewise not an observation of whether a discarded digest was computed. Add maintained observation of hash invocation at the protected production entries and kill the semantics-preserving hash-before-refuse mutant. Correct both validation resources' claims that empty digest/error shape proves order. **repeat-of: none** — the prior finding concerned an actual production ordering defect; this finding concerns the missing regression against that defect class.

Machine-readable findings are attached as `TASK-261006-3ptm2x_review-findings-rev3.json`. The previous verdict is legacy prose; its numbered references above are explicit manual mappings, not a claim that a previous findings array exists. Rework these findings within this leaf, then run another reviewer cycle.

## Prior findings and hosted failures

- Revision 1 finding 1: fixed. The legacy-marker collision probe now returns the opaque refusal; clean v1 and NUL-bearing v2 controls are current. The v2-to-v1 marker downgrade is refused. Recorded-version guards precede all three CLI drift recomputations.
- Revision 1 finding 2: included-byte production hash ordering is fixed at audit, closure, store, staging and recorded readers. F1 identifies the remaining frozen snapshot scope gap; F3 identifies the ineffective order regression.
- Revision 1 finding 3: verdict carriers fixed and mismatch tests kill the version-ignore mutant. Pin carrier remains F2.
- Revision 2's five hosted failure leaves: status plan guards now reach project/global status, CLI stderr has a producer regression, and draft tests expect the earlier resolve refusal with lock preservation. The exact revision-3 hosted gate is green. The updated draft tests cover included assets, not the F1 excluded-file shapes.

## Surface sweep

Scanner coverage is **5/5 production scan calls**: three qualified calls in audit, one in contextaudit, plus the new unqualified `NULPaths` call inside opaquescan.RefuseNULV1. The producer's four-call inventory omits this new scanner wrapper. `git grep` over the checked-in files was cross-checked with `rg` including new candidate files. No version decision uses digest bytes.

| Surface | Version source | Result |
| --- | --- | --- |
| Audit gate, auditSubject | Subject.HashVersion; zero=v1, unknown refuses | Correct scan dispatch; v1 refusal precedes hash |
| Audit detect | Legacy v1 canary/helper | Versioned pipeline uses detectWithOpaquePaths |
| Shared RefuseNULV1 | Explicit caller version | V1 scans recursively; v2 skips; unknown refuses |
| Contextaudit Detect / DetectAtVersion | Legacy entry / lock version | Correct; v2-as-v1 mutant killed |
| Project/global install audit | Draft lock => v1; otherwise WriteVersion | Correct wiring; hosted tests accepted |
| CLI audit / external repository audit | WriteVersion | Correct v2 wiring; hosted tests accepted |
| Draft source audit | Explicit frozen v1 | Correct pre-hash refusal; order test gap F3 |
| Frozen ContentHashFor / resolve / refresh | Frozen v1 shape | F1: full-snapshot guard incomplete |
| Install registry hash / staged hash | Draft Package => v1; otherwise writer | Local guard before hash; hosted evidence accepted |
| marker.Current | Recorded version / frozen shape | Collision and downgrade probes pass |
| Scope drift / hybrid / draft status | Recorded marker version | Guards before recompute; hybrid reports drift state |
| Project/global status plans | Recorded marker version | New helper refuses v1 NUL; CLI/install hosted evidence accepted |
| marker.Write | V2 writer or frozen Package; v1 preserves supplied digest | V2 rehashes; v1 staging guards upstream |
| contextstore.ContentHash / EnsureState | Writer version | Guard before writer-selected hash |
| Strict profile audit / update | Resolved lock version | DetectAtVersion and pre-hash guard agree |
| Profile store state-pin verification | Lock version | Guard before §8 state hash; Git commit tree IDs are a separate scheme |
| Managed skillsOf | Lock version | Guard before member hash |
| Old profile identity migration | Old-lock verification then explicit v2 rehash | Inherits guarded store boundary |
| Global skill migration | Explicit v1 audit | DetectAtVersion precedes strict hash |
| Registry / source-audit binding comparisons | Explicit recorded version or frozen v1 shape | Upstream audited snapshot; no digest inference |
| Verdict cache | Subject version; schema-1 legacy=v1 | Correct versioned carrier and mismatch refusal |
| Trust pin carrier / approval comparison | Digest-only API | F2 |
| Revocations | Existing bare-digest operator list or source pattern | Documented writer-version cutover; no v1 fallback hashing |
| Materialized home surface hashes | Lock version; manager-rendered in-memory set | Uses §8 framing, so not a distinct hash algorithm. Snapshot admission/read guards supply protection upstream; arbitrary direct helper input is outside the production-entry proof here |
| Snapshot inventory, build receipts, marker file fingerprints, Git OIDs | Separate identity contracts | Not used to infer content framing |

This is an inspected call-site table, not a universal runtime guarantee. Concurrent mutation between scan and hash and exhaustive profile/rendered-surface mutation coverage were not established by this review.

## Evidence and bounds

Hosted [run 37467565397](https://github.com/relux-works/curator/actions/runs/37467565397) was independently queried: completed/success, head `5a2f11e909450a8e97d7fcca25d638f6f0060e7b`. Its tree equals the candidate exactly. Reused its production CLI/install/profile, lint, build, cross-platform and race evidence. The attached validation log's measured configured command-shard coverage is **1/1**; hosted individual test-case coverage remains unknown. Neither cmd/curator nor install tests were executed locally.

Locally reran only internal/audit, internal/contextaudit, internal/opaquescan and internal/hashing, including pure reviewer probes in audit calling production APIs. Every Go test went through `mini-build-lock run opaque -- env GOFLAGS=-work go test ./internal/<pkg> -run <regex> -count=1 -timeout=6m -v`. No fake executables or executable fixture scripts were created. Every call measured syspolicyd running with successive crashes **26 -> 26**. Per-call durations, masks, exit codes and sanitized outputs are attached in `TASK-261006-3ptm2x_review-checks-rev3.jsonl` (19 bounded calls, 27.039 seconds total). These totals include intentional failures and two invalid prototypes, which are excluded from behavioral ratios.

Requested adversarial shapes: **5/5** examined (legacy marker, mixed v1/v2 tree, installed status after NUL appears, v2-to-v1 downgrade, cached v1 collision twin); their maintained/currentness expectations hold. CLI status execution is accepted from hosted evidence, not claimed as a local replay. Additional reviewer probes: **2/6 top-level expectations hold** on the candidate (legacy/currentness/downgrade and observed source-audit pre-hash refusal), **4/6 fail** (frozen helper scope, production ResolveDraft scope, legacy pin mismatch, new v2 pin carrier).

Valid maintained-test mutation coverage: **7/8 killed**, **1/8 survives**. The surviving M-order is F3. Reviewer instrumentation kills that survivor, but this is not counted as maintained coverage.

| Mutant | Observed killing test |
| --- | --- |
| M-v2-unconditional | TestGateV2AdmitsNULBearingTreeWhenAuditIsDisabled |
| M-legacy-default-v2 | TestGateV1KeepsOpaqueBlockWithUnchangedFinding/legacy-zero |
| M-unknown-v2 | TestGateRejectsUnknownHashVersion |
| M-cache-version-ignored | TestV2RejectsLegacyUnversionedVerdict; TestV1IgnoresV2VerdictCarrier |
| M-v2-frozen-carrier | TestV2VerdictCarrierRecordsHashVersion |
| M-scan-top-level-only | TestRefuseNULV1ScansNestedFiles |
| M-context-v2-as-v1 | TestDetectAtVersionAppliesNULRuleOnlyToV1 |
| M-order (hash then discard before refusal) | Survives all three maintained refusal tests; reviewer TestReviewSourceAuditNeverHashesNULV1 kills it |

The first unconditional-scan prototype failed compilation due to an unused variable; corrected and rerun. The first expanded probe referenced a nonexistent Lock.Sources field; corrected to the real lock and rerun. Neither invalid prototype counts as evidence of behavior. Production mutations were restored and byte-compared against the candidate; the restored relevant audit tests and legacy probe are green. Portable reviewer probes plus the hash counter are attached in `TASK-261006-3ptm2x_review-attack-probes-rev3.patch`; apply only in a disposable copy. The mutant patch is the one-line unused hash invocation specified under F3.

CHANGELOG Unreleased is present. Candidate diff passes whitespace checks and contains no LOGBOOK.md or scripts/remote-gate.sh changes. Findings are recorded here and in board notes instead of LOGBOOK, as the binding host rule requires. Workspace source was left unchanged; no commits, branch changes, rebase, merge, background process left running, or integration were performed.
