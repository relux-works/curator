# TASK-261016-wh6v9a — credential-store-policy-successors: revision 3 review

ACCEPT for the explicitly narrowed producer slice. No blocking P1/P2/P3 finding reproduced.

```json
{
  "verdict": "accepted",
  "change_request": "CR-TASK-261016-wh6v9a-3",
  "revision": 3,
  "candidate_tree": "da51dd68abb173fc894d8369e28662d6a17886d7",
  "findings": [],
  "notes": [
    "Revision 1 was accepted; the legacy rejection sentence in the generic brief is inaccurate. Revision 2 requested only full-gate revalidation.",
    "did not reproduce: dryrun-go-probe-timeout (revision 2). Exact Ubuntu artifact contains pass for TestAuthoritativeDryRunCasesMutateNothingPersistent/compiled-cache-miss-is-read-only.",
    "did not reproduce: legacy-source-diagnostic-preempted (revision 2). Exact Ubuntu artifact contains pass for TestReviewProbeLegacyLaneIdentityInvalid.",
    "Independent clean merge of revision 2 onto current base produces 4f88746642b992950ba314473747e321218eba00. Candidate differs only by the requested verbatim relocation of both task LOGBOOK blocks, with the newer task block first. All 87 other candidate paths remain byte-identical to revision 2. No new product/test changes or weakened assertions.",
    "Coverage: 7 of 8 narrowed producer AC rows production-driven; row 8 has named schema/boundary tests and an explicit external wire-ingestion bound. The capture/status interaction is 1 of 1 driven. Production call-site map is retained in the revision-3 handoff.",
    "Fresh local focused/CLI/free-hunt suites passed; 16 Muse/root narrowing controls killed by named behavioral test failures. Twelve standalone SNAP-LOCK controls are reused revision-2 evidence, not fresh execution. No exhaustive absence claim.",
    "Full hosted gate on exact signed snapshot succeeded: 20 jobs success, 2 workflow-configured optional skips. The owned local waiter was stopped at the bounded shell window (exit 137); hosted completion was observed separately. No local-script exit 0 claimed.",
    "Windows vendor sentinel and its Muse reobservation control remain explicitly Unix-only. No Windows vendor-execution proof inferred.",
    "Accepted scope remains managed Muse file policy, shared/isolated admission, v4 composition and verified root fact. Readiness/login, qualified Codex v5/application, real hosted override plant and adopted external-source policy remain in TASK-261017-0pawkw (credential-store-policy-consumer-remainder).",
    "Fresh origin/main equals base a17e58156ac30a1629e3ff36efa259bab1555a15. Final worktree tree equals candidate. Reviewer changed no product code, Story branch or real index. Acceptance is not landing.",
    "No review/free-hunt budget or separate catalog table was supplied. Retained all eight prior product rows and the required hosted-validation row; fresh named attacks and bounded inherited evidence are stated in the sweep.",
    "Initial gh run log reads while workflow was running were unavailable, not absence. The completed Ubuntu artifact supplied explicit test pass events."
  ],
  "surface_results": [
    {
      "row": "observation boundary / cache",
      "result": "held"
    },
    {
      "row": "destination conflicts",
      "result": "held"
    },
    {
      "row": "read boundary",
      "result": "held"
    },
    {
      "row": "error surface",
      "result": "held"
    },
    {
      "row": "selector paths",
      "result": "held"
    },
    {
      "row": "text grammar",
      "result": "held"
    },
    {
      "row": "identity",
      "result": "held"
    },
    {
      "row": "ordering",
      "result": "held"
    },
    {
      "row": "required hosted validation",
      "result": "held"
    }
  ],
  "free_hunt_findings": []
}
```

# Revision 3 review sweep

Candidate: da51dd68abb173fc894d8369e28662d6a17886d7.
Revision-3 scope is convergence plus logbook relocation and fresh required CI.
Coverage checked before source inspection: 7 of 8 narrowed producer AC rows driven; row 8 has named boundary/schema tests and explicit external-consumer bound. No new AC row is omitted. Revision-2 interaction: 1 of 1 driven.
Clean merge independently recomputed with git merge-tree: 4f88746642b992950ba314473747e321218eba00. Only LOGBOOK.md differs. All 87 other candidate paths are byte-identical to revision 2. Both log blocks retain their content.

| Surface | Result | Fresh attack evidence / limits |
|---|---|---|
| observation boundary / cache | held | focused.log: TestMaterializeCaptureAndStatusNeverProbeVendor and CredentialPolicyRelease tests; real vendor sentinel, native/file x shared/isolated. CLI plant restores only Muse release observation and is killed. |
| destination conflicts | held | focused.log: TestMusePolicySwitchPreservesCredentials and TestMusePolicyExplicitMigration; refusal versus explicit migration controls through Resolve/ApplyMigration. |
| read boundary | held | focused.log: TestMuseAuthParentBoundaryAtResolve and TestMuseAuthRepairAdditionalStates; bad links/types against Resolve. |
| error surface | held | focused.log: TestMusePublishedLinkStates and TestMuseMissingNativeFilePolicy; failed/missing shared source refuses without placeholder. |
| selector paths | held | cli.log: TestMuseCLIStorePolicy; native/file positive controls and keychain refusal via real CLI; admission narrowing plant killed. |
| text grammar | held | cli.log: published v3/v4 corpus (pinned CI schemas) and boundary mutants admitting HOME, absent literal, false root are killed; external consumer ingestion remains unclaimed. |
| identity | held | cli.log: TestMuseRootContextVectors/Drift/Status and seven root narrowing plants, including header-only context; real Resolve/StatusOf entry points. |
| ordering | held | cli.log: TestMuseCLIStorePolicy; renderer plant removes only backend while preserving source tokens, named behavioral test fails. |

## Free hunt

free-hunt.log: TestReviewMuseCaptureRefusalNoProbe drives CaptureMaterialize and StatusOf through valid shared file, removed file, then restored file. Refusal is environment_home_stale, callback never runs, state stays unchanged, vendor trap never fires. Exit 0. No finding reproduced.

Fresh local controls: 16 Muse/root narrowing plants killed by named behavioral failures (9 policy/fragment and 7 root). Twelve standalone SNAP-LOCK overlay controls are inherited from revision-2 review, not rerun or represented as fresh evidence. Product/test definitions are unchanged; changed upstream scriptworker behavior is exercised by the fresh whole gate.

| required hosted validation | held | Fresh run 39018771945: 20 success, 2 configured optional skips; both prior Ubuntu failures have explicit pass events on exact candidate. |

## Evidence and review logbook

- TASK-261016-wh6v9a_review-local-rev3.md: exact commands, real exits, independent merge comparison, fresh focused/CLI suites, 16 narrowing controls, and capture/refusal/recovery attack.
- TASK-261016-wh6v9a_review-hosted-rev3.json: terminal hosted job results, exact signed snapshot and both previous failing tests' pass events.
- TASK-261016-wh6v9a_review-free-hunt-rev2.go: unchanged review-only test replayed via Go overlay; no repository source mutation.

2026-10-17 review logbook: revision-3 convergence matches clean merge plus the requested task logbook relocation. Both revision-2 Ubuntu findings did not reproduce. Independent exact-tree full CI succeeded. The reviewer accepts only the already narrowed Muse producer scope; consumer remainder stays preserved. This task-scoped logbook entry travels as an outcome resource, without a control-root edit.
