# ACCEPT — M1 revision 1

```json
{
  "verdict": "accepted",
  "change_request": "CR-TASK-261016-ont5q1-1",
  "revision": 1,
  "candidate_tree": "dc879a2d8126093d786c98a99c18e531829407a2",
  "base_oid": "34a89308a69d4863699887359cd122d64b1e716f",
  "ac_coverage": "4 of 4 AC rows driven",
  "paired_coverage": "6 of 6 paired rows driven",
  "surfaces": [
    {
      "row": "R1",
      "result": "held",
      "evidence": "TASK-261016-ont5q1_review-sweep-rev1.md"
    },
    {
      "row": "R2",
      "result": "held",
      "evidence": "TASK-261016-ont5q1_review-sweep-rev1.md"
    },
    {
      "row": "R3",
      "result": "held",
      "evidence": "TASK-261016-ont5q1_review-sweep-rev1.md"
    },
    {
      "row": "R4",
      "result": "held",
      "evidence": "TASK-261016-ont5q1_review-sweep-rev1.md"
    },
    {
      "row": "R5",
      "result": "held",
      "evidence": "TASK-261016-ont5q1_review-sweep-rev1.md"
    },
    {
      "row": "R6",
      "result": "held",
      "evidence": "TASK-261016-ont5q1_review-sweep-rev1.md"
    }
  ],
  "findings": [],
  "free_hunt_findings": [],
  "notes": [
    "No catalog surface table or review budgets were supplied. Used all six mandatory paired rows as the floor and a bounded additional free hunt.",
    "Fresh origin main 26d8ddbf1278ac025a3333aee1053adbdb37f1bf is two commits ahead of the base. Candidate/upstream changed paths do not overlap. Upstream changes CI; this review qualifies the exact candidate only, not an unvalidated integration tree.",
    "No live vendor login, refresh, 1.4.8 exclusion qualification or Windows ACL qualification was performed or inferred. These are declared bounds, not absent problems.",
    "The six-file delta is justified: status behavior plus existing public CLI/schema/root test seams and environment documentation. Reuses the common credential-policy resolver; no parallel resolver or marker expansion.",
    "The older conformance runtime notes use a generic keychain diagnostic; current environments section 7.4 explicitly refines it to environment_credential_store_unsupported, which implementation and tests use."
  ],
  "hosted_gate": {
    "url": "https://github.com/relux-works/curator/actions/runs/39034928676",
    "head": "16b8c4679aa2fe7164898e681432703b332bbce2",
    "exit": 0,
    "succeeded_jobs": 20,
    "conditional_skips": 2
  }
}
```

# Independent M1 review evidence, revision 1

Task: TASK-261016-ont5q1 — implement-rc7-muse-v4-producer.
Candidate tree: `dc879a2d8126093d786c98a99c18e531829407a2`.
Base: `34a89308a69d4863699887359cd122d64b1e716f`.
Signed hosted snapshot: `16b8c4679aa2fe7164898e681432703b332bbce2` (tree exactly equals candidate; signature `G`).

## Contract and coverage

Read attached leaf, decisions and producer gap analysis before implementation review. Coverage is **4 of 4 AC rows driven** and **6 of 6 paired rows driven**, with the actual named tests and call sites checked against source. The baseline implementation is reused appropriately. The six-file delta adds status distinctions and strengthens existing test seams. No new resolver, marker field, credential migration or consumer capability was introduced.

Normative authority: curator-spec main `1059b94b4135089d842bf017070b8ab47d79a536`, read via fresh GitHub API; local environments text compared byte-identical. Read CIP-0017 implementation leaf 5, environments §§5.4, 7.1, 7.4, 8.2–8.4.1, 10.1–10.3, 12, v4 schema and conformance notes. Current environments §7.4's explicit store diagnostic refines the older generic wording in conformance notes.

| AC | Production path | Named tests inspected and independently replayed |
|---|---|---|
| 1 | `cmdEnvResolve -> Resolve -> buildFragment -> renderFragment` | `TestMuseCLIFragmentAndFakeLaunch`, `TestMuseCLIStorePolicy`, `TestMuseCredentialPolicyMatrix`, `TestMuseFragmentV4PublishedCases` |
| 2 | `cmdEnvResolve -> Resolve -> verifyHome/checkSurfaces`; repair/apply/copy fallback; `cmdEnvStatus -> StatusOf` | `TestMuseRootContextVectors`, `TestMuseRootContextDrift`, `TestMuseRootContextRefusals`, `TestMuseRootContextCopyFallback` |
| 3 | public config/resolve -> operation-local credential policy and `inspectMuseAuth` | `TestMuseCLIStorePolicy`, `TestMuseCredentialPolicyMatrix`, `TestMuseMissingNativeFilePolicy`, `TestMusePolicySwitchPreservesCredentials` |
| 4 | `Resolve -> inspectMuseAuth/applyPlan`; `StatusOf -> homeState/verifyHomeContent` | `TestMusePublishedLinkStates`, `TestMuseAuthRepairAdditionalStates`, `TestMuseCLIFragmentAndFakeLaunch`, `TestMuseRootContextStatus`, `TestMuseCredentialPolicyStatus` |

## Executed checks

Go 1.25.5, darwin/arm64. Synthetic isolated fixtures only. All local commands below exited 0. No local Muse row was skipped. Both narrowing harnesses passed by requiring each expected-red subprocess to exit 1 with its named behavioral failure; all 21 controls were killed, including `renderer-drops-only-backend`.

```sh
CURATOR_CONFORMANCE_ROOT="$PWD/.temp/TASK-261016-ont5q1/spec/conformance/v1" go test ./internal/envfragment ./internal/envprofile ./cmd/curator -run '^TestMuse' -count=1 -timeout=8m -v
go test ./cmd/curator -run '^TestMuseRootContextVectors$' -count=1 -timeout=3m -v
python3 .temp/m1-review/prepare-attacks.py
go test -overlay "$PWD/.temp/m1-review/overlay.json" ./cmd/curator -run '^TestM1Review' -count=1 -timeout=5m -v
python3 .temp/m1-review/prepare-boundary.py
go test -overlay "$PWD/.temp/m1-review/boundary-overlay.json" ./cmd/curator -run '^TestM1ReviewBoundary$' -count=1 -timeout=3m -v
git diff --check
git diff dc879a2d8126093d786c98a99c18e531829407a2 --exit-code
```

The additional scripts are attached as task-scoped `review-prepare-attacks-rev1.py` and `review-prepare-boundary-rev1.py`. Materialize them under the command paths above; they produce disposable Go overlays from candidate files. They never edit repository source. The first attacks forged missing marker/surface evidence, same-byte link retarget, repeated resolve/status, and a changed unsupported selector. The second corrupts six proposed fragment shapes immediately before the production boundary and verifies the CLI refuses each with diagnostic text, no fragment and unchanged marker, beside a successful unpoisoned provisioning control.

Full six-row results and limits: `TASK-261016-ont5q1_review-sweep-rev1.md`. Logs: `review-focused-rev1.log`, `review-attacks-rev1.log`, `review-boundary-rev1.log` (all task-prefixed board resources). The targeted row-2 replay is redundant to the full focused evidence and passed locally; no additional claim relies on it.

## Bounds and integration handoff

Fresh protected main `26d8ddbf1278ac025a3333aee1053adbdb37f1bf` is two commits ahead of the CR base; changed-path intersection is empty. Upstream modifies CI. This review is bound to the exact CR candidate and signed gate snapshot, not a future combined integration tree. The delivery owner retains exact-head integration validation obligations.

No runtime login/refresh, model-visible exclusion, exact 1.4.8 exclusion replay, consumer launch or Windows ACL qualification is claimed. The status note correctly distinguishes historical observations and unknowns. No blocking finding reproduced; `held` records only the named attacks.

## Terminal full-gate evidence

`sh scripts/remote-gate.sh` exited **0**. The independent hosted run is https://github.com/relux-works/curator/actions/runs/39034928676 . All 20 executable jobs succeeded on signed snapshot `16b8c4679aa2fe7164898e681432703b332bbce2`; two conditional jobs (dispatch-only candidate suite and optional self-hosted macOS) were skipped, not executed. No retry, mirror or override was used. Lint, three platform tests, both race jobs, driver matrix, conformance and platform gates passed. Final tracked worktree still equals the candidate.

The full gate outlasted the initial 20-minute review target; attacks were already finished and no extra attack scope was added during that wait. Completion waited synchronously for hosted evidence, which is attached as `TASK-261016-ont5q1_review-hosted-rev1.json` and `TASK-261016-ont5q1_review-remote-gate-rev1.log`.

Reviewer goal lookup immediately before verdict: this run is not goal-bound. Acceptance is recorded only through `accept_cr` for revision 1, with no `commit_ack`. No repository source was modified and no commit was made on the Story branch. The non-acceptance checklist item is not applicable to this accepted verdict. Acceptance does not claim landing or release qualification.
