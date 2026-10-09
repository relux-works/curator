# TASK-261008-1pg0pd — N9 audit pin creation time: review verdict, revision 1

Verdict: accepted. No blocking findings. Accept CR-TASK-261008-1pg0pd-1 revision 1 and route to integrating; landing remains the producer's responsibility.

Reviewed base: d46f2b1f41ee24276fc1d809bcb7bcef51ad5250.
Candidate tree: d8b887bd846208b2418ae898b69b4ab574541c7d.
Patch SHA-256 independently verified: 8f4b2dfbfefd982809400a4f9c028210256644b27c429b32bcda563b3c1faed9.
Fresh origin HEAD advertisement and exact main fetch both equal the CR base; there is no upstream overlap. All four working-tree paths match the candidate bytes. Repository code and shared index were not modified by this review.

| Reviewed surface | Evidence and conclusion |
| --- | --- |
| Pin writer and production call chain | CLI run dispatch -> audit handler -> audit.PinAtVersion; legacy audit.Pin delegates to the same writer. The record gains created_at using time.Now().UTC().Format(time.RFC3339) before either v1 or v2 serialization. This satisfies report N9 and curator-spec manager section 7 at pinned revision 43bf0a2506d5c354a73bbc3ea4623d4653db10c7. |
| Regression validity and reachability | TestAuditAllowRecordsPinCreationTime drives the real CLI, with an explicit synthetic operator and reason. Both writer versions check persisted content identity, pinned flag, operator, reason, timestamp presence, UTC Z suffix, RFC3339 validity, and proximity to the command window. Hosted green: 2/2 writer cases pass. Hosted production-fix removal: 2/2 fail for absent created_at. |
| Versioned carriers and compatibility | No read-side decision code changed. Existing legacy/schemaless pins still read as v1, cross-version pins remain refused, and matching versions remain honored. TestPinAtVersionWritesVersionedCarrier, TestPinVersionMatchAuthorizes, TestV2RejectsLegacyV1Pin, TestV1RejectsV2Pin, and TestLegacySchemalessPinReadsAsV1 all pass in inspected hosted evidence. Existing test assertions were not weakened; the changed test file contains only comment corrections. |
| Audit and install controls | TestGatePinPolicyFreshAndCached has 52/52 passing cases: 13 policy rows x Gate/GateReadOnly x fresh/cached. TestGateModes, TestGateReadOnlyDoesNotWriteVerdictCache, TestRequirePinForOldSchemas, TestLocalRevocations, TestRegistryRevocationDeniesInstall, TestDraftAuditPinDoesNotWaiveStrictFindings, and TestDraftAuditRevocationBlocks pass. TestDraftAuditBrokenReportRefuses passes 4/4 missing, wrong-digest, mismatching, and forged report cases. These controls also pass in the red candidate. |
| Operator documentation and hygiene | The Unreleased/Fixed changelog accurately describes newly issued pins. Timestamp encoding follows the existing source-audit created_at convention. Candidate diff --check and gofmt checks are clean; hosted lint is green. |
| Frozen validation identity | Both green hosted commits have exactly the CR candidate tree. The red snapshot differs only in the time import and created_at writer member. Hosted evidence is about this candidate, rather than a nearby or moving worktree. |

Hosted evidence independently inspected (existing runs reused; no suites rerun by this reviewer):

- Green: https://github.com/relux-works/curator/actions/runs/37883305445 — success; head e721753b5c573df34be3347e28345be3d7dd9a99; tree d8b887bd846208b2418ae898b69b4ab574541c7d. Downloaded test-evidence-ubuntu-latest: zero failure actions and the regression/control results above.
- Red: https://github.com/relux-works/curator/actions/runs/37886218996 — expected failure; head c882545b6d493c366eb5557737a7d55bf758e824; tree 39dcefc795d282835329def450d0c979441ba405. Downloaded Ubuntu evidence contains exactly the two regression subtest failures, their parent test failure, and the resulting package failure. Both report: pin created_at = <nil>, want a defined creation timestamp. All five Test/Race jobs fail; the other executed jobs pass.
- Handoff gate: https://github.com/relux-works/curator/actions/runs/37890746424 — success; head c19a90ecc3bca2a5920424b13a4b0e90a9f2343b; identical candidate tree. 20/20 executed jobs succeed. Candidate-suite and rose-air jobs are intentionally skipped; they are not claimed as executed.

Reviewer mutant: omit-created_at, reconstructed in a disposable shared clone from the exact green snapshot by removing only the timestamp member and now-unused time import. Its reconstructed tree equals 39dcefc795d282835329def450d0c979441ba405 and its full bytes match the hosted red snapshot. Reviewer ran go vet ./internal/audit/ ./cmd/curator/ and go build ./... on that clone: both exit 0. The regression failure is taken from the existing hosted run, not a local test execution. No local go test was run (R223). Producer candidate compile evidence and hosted vet/build/lint are accepted from the attached results and gate.

Evidence digests, SHA-256:

- Green go-test.json: 1343921451bbee2ac185e4077cf88c354bfde381d1b74a3a4336199e102e482e
- Green observed-cases.tsv: e8610391b126251e5f6f6e78bcc56b303e43d68185e9d37c6dcd242b7b9c106a
- Red go-test.json: 0145b74e954044ee263b731c856ac0a8a6557c6bd8c8589f90d6f6a02b033e59
- Red observed-cases.tsv: 3cca78a92c6e6df27a6af443a56cba0278e852cd21782379b6f92506388cc612

Review observations and bounds:

The attached N9 probe is actually TestWave2AuditDecisionTable, rather than the report's CLI creation-time probe. The replacement regression exercises the correct production entry and invariant. The literal historical 20-decision execution (10 rows through two entries) is not claimed as replayed: its pin-over-finding and pre-capability-pinned expectations predate the current baseline, which blocks strict findings even with a pin. The current pinned source-audit contract and preexisting 52-case policy test establish that behavior. The source-only skill-* glob row is not separately replayed; the current hosted suite exercises Git source globs, hash revocations, and install revocation refusals. N9 changes no revocation matcher or decision code. The measured coverage above describes the inspected current suite, not every possible policy, clock, filesystem, expiry, or concurrency behavior.

No repository LOGBOOK edit was made, per the task brief. These observations persist in this task-scoped outcome. The conditional changes-requested checklist item is not applicable to this accepted review; the evidence and explicit acceptance mutation provide the verdict lifecycle.
