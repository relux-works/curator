# TASK-261017-sxydvh — revision 2 review verdict

Verdict: **accepted**. This is acceptance of the candidate, not landing.

```json
{"verdict":"accepted","revision":2,"findings":[],"notes":["The existing detector surface is HTTP(S) references in supported script/manifest files; this review does not claim general network discovery.","The full local suite was independently checked from attached fresh evidence, not replayed in this review. Targeted behavioral suites and narrowing mutants were rerun."]}
```

## Candidate and contract

Reviewed base `402fdbc9b77803cb2d6b11f4f1cfd83961e6897b`, candidate tree `7ced0d2be8a93bfd6063b22e70f83c8e37a7e29a`. All 11 changed working-tree files match candidate blobs exactly. Normative environments §§9.1/9.4 and all 40 fixture rows were read from curator-spec `c1a2bb256f10e583199f848bfb3a7a4d30e0d9bc`; the committed fixture matches its blob byte-for-byte. Nonempty repository delta is appropriate: classifier/parser, ruleset invalidation, tests, IDNA dependencies and documentation implement the approved change.

The refresh README delta was inspected first. CLIHELP and snapshot-publication text remain present; the network-vector paragraph appears once beside the audit commands, and ruleset 4 retains the ruleset-2/3 explanation. No implementation/test/dependency change was introduced by the refresh.

## Acceptance coverage

**4 of 5 AC rows driven** by named production-entry tests; AC5 is validation-only, not a product behavior. **5 of 5 AC rows have evidence.**

| AC | Driver and production call site |
| --- | --- |
| 1 | TestNetworkContactVectors and TestProfileNetworkContactVectors: audit.Gate/GateReadOnly (audit.go:196/202), Install → strictAuditMember (envprofile.go:1557/1769), EnsureDefaultWithPolicy → migrateGlobalSkills → strictAuditMember (2068) |
| 2 | Same vector suites; TestNetworkContactNormalizationBoundaries; reviewer TestReviewerNetworkFreeHunt through Gate |
| 3 | 40/40 pinned vectors through Gate/GateReadOnly; 40/40 install and 40/40 migration terminals |
| 4 | TestNetworkContactCacheInvalidatesRulesetThree through Gate and production auditSubject; TestNetworkContactNarrowingMutants runs behavioral child tests |
| 5 | Fresh local evidence verification and two exact-tree successful hosted runs; signatures independently verified G |

## Swept surface table

Every row has one result. Exact commands and complete rerun logs are in `TASK-261017-sxydvh_review-evidence-rev2.jsonl`, entries sweep.jsonl and row-00 through row-08.

| Surface | Result | Attack / evidence |
| --- | --- | --- |
| host normalization | held | Case, trailing-dot, IDN/punycode vectors; authority/port/credential and malformed-path boundaries; row-00 plus normalization.log |
| exemption set | held | Loopback, all reserved suffixes, bare-label negative controls through Gate; row-01 |
| lookalikes | held | Nested suffixes, hyphen, IDN lookalikes and bare labels; explicit test.com free-hunt control; row-02 |
| canary | held | example.com/net/org and subdomains, declared/undeclared canary, CanaryPasses; row-03 |
| mixed file | held | Both host orders and cross-file exact file/host evidence; row-04 |
| migration audit | held | All 40 vectors through real EnsureDefaultWithPolicy and Install, including no-publication assertions on refusal; row-05 |
| audit cache | held | v1/v2, stale deny and stale allow, miss then hit, old-record byte preservation; row-06 |
| vectors | held | 40-row corpus independently compared to pinned spec; all 40 Gate vectors and 80 pipeline terminals rerun; row-07 and row-05 |
| mutants | held | Six controls: full-url-percent-path, contains-test, exempt-example-com, exempt-bare-test, reserved-clears-file, reuse-ruleset-three. Every child exits 1 with its intended behavioral assertion; outer harness exits 0; row-08 |
| hosted gate | held | Native provider reads confirm full success; candidate mismatch ruled out by signed commit tree comparison; managed-hosted.json |

Free hunt: 12 additional Gate probes cover test.com, nested uppercase suffixes, normalized canary, both userinfo/host directions, Unicode separator/fullwidth normalization, bare trailing-dot test, non-exempt IPv4 neighbor, localhost lookalike and mixed invalid/real hosts. All expected decisions held. No additional findings. The normalization boundary suite separately retains the independent exec finding. These are bounded observations, not proof of absence.

## Validation evidence

Reviewer reran all sweep commands with `-count=1`; all exited 0. No process or pending evidence write remains. Code was not modified; the extra reviewer test used a Go overlay stored in task scratch.

Reused full-local evidence is from the refreshed signed candidate `5ceef5a6656555f87460ec4c05875d18c6e0b5fb`, tree equal to this CR. All seven attached bundle hashes and 294 embedded file hashes verified. Independently parsed 112 successful command receipts and complete package streams: 85 packages, 54,256 events, no accepted fail event, 3,469 pass / 31 skip top-level terminals. The independent `go test -list` inventory has 3,499/3,499 named tests with observed terminal events. The diagnostic failed setup attempt remains excluded, not relabeled as passing. Build/vet/lint and gate receipts are retained in the fresh validation bundle. Evidence-rejection controls include missing receipt and digest-consistent truncated stream, both expected-red.

Full standalone hosted run: https://github.com/relux-works/curator/actions/runs/39096450283 — completed/success on `5ceef5a6656555f87460ec4c05875d18c6e0b5fb`.

Managed handoff hosted run: https://github.com/relux-works/curator/actions/runs/39105409112 — completed/success on `3f0fa49b0bf99bde9aa19b349d5d22bfa13331b9`. Its tree equals `7ced0d2be8a93bfd6063b22e70f83c8e37a7e29a`; signature G. The managed validation log ends with real exit 0. Required jobs succeeded; optional rose-air and candidate-suite jobs were skipped as configured. Both provider reads were performed afresh in this review.

## Logbook

Revision 1 was superseded before product review, so there is no prior blocking finding or repeated mechanism. The refreshed README combines the upstream features without duplication. Review found no blocking findings; accept revision 2 and leave manual signed PR delivery to its owner.
