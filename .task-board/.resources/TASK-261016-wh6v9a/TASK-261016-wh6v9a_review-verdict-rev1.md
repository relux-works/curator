# TASK-261016-wh6v9a — credential-store-policy-successors: review verdict, revision 1

ACCEPT for the explicitly narrowed producer slice. No blocking P1/P2/P3 findings reproduced. This is acceptance for integration, not a landing claim.

```json
{
  "verdict": "accepted",
  "change_request": "CR-TASK-261016-wh6v9a-1",
  "revision": 1,
  "candidate_tree": "ad5e8e7f45c56718d5889c9c675d4a33765b3b27",
  "findings": [],
  "notes": [
    "Scope accepted only for managed Muse effective-file policy, shared/isolated admission, v4 composition and verified root fact. Readiness/login, qualified Codex v5/application, real hosted plant and adopted external sources remain in TASK-261017-0pawkw (credential-store-policy-consumer-remainder).",
    "7 of 8 narrowed AC rows are production-driven. Row 8 has named schema/boundary tests and an explicit external wire-consumer bound. This is appropriate for a producer-only leaf; no external launcher ingestion or runtime qualification is claimed.",
    "The brief supplied neither a catalog surface table nor review/free-hunt budgets. The reviewer supplied the explicit seven-row sweep below; all rows held for the named attacks, without claiming absence of all defects.",
    "The gate script signed and verified the exact candidate snapshot. Its local waiter was intentionally stopped before the bounded shell window (exit 137), then the same hosted run was observed through bounded calls to terminal success. No local-script exit 0 is claimed.",
    "Checklist conditional rejection-routing item is not triggered by this acceptance; acceptance evidence and routing are recorded instead."
  ],
  "surface_results": [
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
    }
  ],
  "free_hunt_findings": []
}
```

# Review sweep — revision 1

Candidate: ad5e8e7f45c56718d5889c9c675d4a33765b3b27.
No catalog surface table or time budget was supplied; this review maps the narrowed producer requirements to catalog surfaces. Scope is the producer slice, with consumer work explicitly deferred. Seven of eight AC rows are production-driven; row 8 has named boundary/schema tests and an explicit wire-consumer bound.

| Surface | Result | Executed attacks and bounds |
|---|---|---|
| destination conflicts | held | TestMusePolicySwitchPreservesCredentials and TestMusePolicyExplicitMigration: detached file and shared-to-isolated repair refusal versus explicit migration; native bytes preserved. |
| read boundary | held | TestMuseAuthParentBoundaryAtResolve and TestMuseAuthRepairAdditionalStates: parent symlink, retarget, relative link, directory target; existing valid route controls. |
| error surface | held | TestMusePublishedLinkStates and TestMuseAuthRepairAdditionalStates: metadata/readlink errors refuse, not absence; missing native source guidance and no placeholder in TestMuseMissingNativeFilePolicy. |
| selector paths | held | CLI config set/show/unset and malformed/alias/dotted-profile cases, TestMuseCLIStorePolicy; forced keychain refusal beside native/file controls. Unqualified Codex activation still refuses. |
| text grammar | held | Unknown/missing/conflicting Muse policy and v1-v3 tokens in 67 frozen schema cases; CheckBoundary negatives; malformed machine config through CLI. Wire-consumer ingestion is explicitly out of this producer slice. |
| identity | held | TestMuseRootContextVectors and TestMuseRootContextDrift: header-only, no-context, missing/drifted surface and repair paths; no successful fact from a stale surface. |
| ordering | held | TestMuseCLIStorePolicy and TestMuseFragmentBoundary/valid: JSON/env/shell keep all five values and env/shell registry order with backend last. Keys are a fixed registry set, not locale-derived names. |

## Free hunt

TestReviewMuseIsolatedCredentialSwitch (overlay-only review test) passed: native-to-file on an isolated home preserves its vendor credential, while isolated-to-shared bare resolve/repair emits no fragment and preserves all credential scope state. Repair gives environment_credential_conflict. No new product source file was written.

Eight candidate narrowing mutants ran in the focused CLI command, including renderer-only backend omission and header-only root omission. The harness requires exit 1 and the named behavioral test failure, rejecting setup/compile failures. No blocking findings reproduced.

The supporting remote-gate script was also attacked in an isolated scratch repository with its exact candidate bytes: override signingkey to a missing public key, invoke the real shell script and real Git signing, and instrument push reachability. It refused with gate exit 1 on the missing key, and no push was reached. The real signed hosted snapshot is the positive control. This is an executed signing failure, not a forged verification result.

# Reviewer validation — revision 1

Candidate tree: ad5e8e7f45c56718d5889c9c675d4a33765b3b27.
Base and fresh origin/main: c29663024abfe6dc0353749018fd65b2bf1833e1.
Live curator-spec main: 1059b94b4135089d842bf017070b8ab47d79a536.
Reviewer signed snapshot: 593936a30bfd70885a4c1e952acfc72273175799; its tree equals the candidate exactly.
Local toolchain: Go 1.25.5, darwin/arm64.

Commands executed independently by reviewer:

- `go test ./internal/envfragment ./internal/envprofile ./internal/envregistry ./internal/config -run 'Test(Muse|CredentialStore|CredentialProjection)' -count=1 -timeout=5m`: exit 0, focused-01.log.
- `go test ./cmd/curator -run 'TestMuse(CredentialPolicyNarrowingMutants|CLI|Fragment|RootContext)' -count=1 -timeout=8m`: exit 0, cli-01.log. Includes eight named narrowing controls; their harness requires the named test failure and child exit 1, not an arbitrary failure.
- `CURATOR_CONFORMANCE_ROOT=<exact CI conformance root>/conformance/v1 go test ./internal/envprofile -run '^TestMuse(PublishedLinkStates|AuthParentBoundaryAtResolve|AuthRepairAdditionalStates|PolicySwitchPreservesCredentials|PolicyExplicitMigration)$' -count=1 -v -timeout=4m`: exit 0, links-01.log. Published link-state vectors actually executed, not skipped.
- `go test ./cmd/curator -run 'TestCredential(Store|Projection)|TestReview.*Credential' -count=1 -timeout=4m`: exit 0, selector-01.log.
- `go test -overlay <review overlay.json> ./internal/envprofile -run '^TestReviewMuseIsolatedCredentialSwitch$' -count=1 -v -timeout=3m`: exit 0, free-hunt-01.log. Overlay appends only a review test to the exact candidate test file; product code is unchanged.
- `git diff --check`: exit 0.
- `sh scripts/remote-gate.sh`, with configured agent Git environment: signed and verified the snapshot, dispatched hosted run 38977912347. Local waiter intentionally stopped with exit 137 before the bounded shell window; this is not a green script exit. Remote CI was not cancelled. Terminal hosted verdict is recorded separately after bounded reads.

Hosted run: https://github.com/relux-works/curator/actions/runs/38977912347

No producer test log is counted as reviewer execution. Producer evidence is contextual only.

Terminal hosted result: **success**; 20 jobs success, 2 configured optional jobs skipped. Skipped: Test (optional-self-hosted), Candidate suite (${{ matrix.os }}). No skipped job is claimed as executed. The hosted result is independent of the intentionally stopped local waiter.
Final working-tree snapshot still equals ad5e8e7f45c56718d5889c9c675d4a33765b3b27.

## Architecture and coverage assessment

`Resolve` / repair verify the home before `buildFragment`, which consumes the cached operation-local `resolveCredentialPolicy` projection. Native selector admission, isolation, credential links and fragment backend therefore have one owner. Successful fragment construction calls `envfragment.CheckBoundary`; no new policy derivation was introduced in CLI code. The supporting signed-gate-script delta is declared and was exercised with both a real signed positive and a missing-key negative that did not reach push.

The seven production-driven rows cover mapping/selector preservation, captured native sharing, missing-source refusal, isolated non-sharing, keychain refusal, repair versus explicit migration, and v4 rendering/root facts. Named tests and production call sites are in the producer handoff; all affected focused groups were rerun here. The 67-case successor corpus (12 positive/55 negative), three added schema versions and index were independently byte-checked against live curator-spec main. Published link-state vectors also match that authority.

No product code was edited. The extra isolation-switch test was applied only through a Go overlay. All review writes are task-local scratch or board resource mutations. The candidate is still uncommitted at its original checkpoint. The conditional rejection checklist row is N/A because this verdict accepts the revision.

Evidence: TASK-261016-wh6v9a_review-validation-rev1.txt; TASK-261016-wh6v9a_review-sweep-rev1.md; hosted run https://github.com/relux-works/curator/actions/runs/38977912347.
