# TASK-261016-wh6v9a — credential-store-policy-successors: revision 2 review

CHANGES REQUESTED / P2: required hosted validation remains unresolved. No reproduced defect in the narrowed Muse/SNAP-LOCK implementation.

```json
{
  "verdict": "changes_requested",
  "change_request": "CR-TASK-261016-wh6v9a-2",
  "revision": 2,
  "candidate_tree": "9bad791482638be26d3ab306653fd279791f449b",
  "findings": [
    {
      "id": "dryrun-go-probe-timeout",
      "row": "required hosted validation",
      "invariant": "The required full gate must pass the published valid compiled-cache-miss dry-run case without persistent mutation.",
      "mechanism": "On the exact candidate, internal/godriver/session.go:231 reports process_timeout; the public install.Project dry-run propagates toolchain-unavailable, rejected by internal/install/dryrun_conformance_test.go:297. The underlying timeout cause and attribution to this delta are unestablished.",
      "reproductions": [
        {
          "test_file": "internal/install/dryrun_conformance_test.go",
          "test": "TestAuthoritativeDryRunCasesMutateNothingPersistent/compiled-cache-miss-is-read-only",
          "command": "sh scripts/remote-gate.sh -> Ubuntu Test: bash .github/ci/test-gate.sh .temp/ci-evidence/test",
          "candidate_tree": "9bad791482638be26d3ab306653fd279791f449b",
          "expected_failure": "Named test fails because toolchain-unavailable/process_timeout is outside the published admitted outcomes; Go test exit 1.",
          "log": "TASK-261016-wh6v9a_review-hosted-attempt1-rev2.md",
          "hosted_job": "https://github.com/relux-works/curator/actions/runs/39004671814/job/117314584571"
        }
      ],
      "severity": "regression",
      "priority": "P2",
      "repeat-of": "none"
    },
    {
      "id": "legacy-source-diagnostic-preempted",
      "row": "required hosted validation",
      "invariant": "The frozen CLI source-identity refusal must retain build_repository_source_unavailable in the required full suite.",
      "mechanism": "The public status app path returns toolchain_mutated/cannot revalidate toolchain instead of the frozen source diagnostic, failing cmd/curator/draft_transport_test.go:814. internal/godriver/session.go:336 wraps revalidation errors; this does not prove actual toolchain mutation. Underlying cause and delta attribution remain unknown.",
      "reproductions": [
        {
          "test_file": "cmd/curator/draft_transport_test.go",
          "test": "TestReviewProbeLegacyLaneIdentityInvalid",
          "command": "sh scripts/remote-gate.sh -> Ubuntu Test: bash .github/ci/test-gate.sh .temp/ci-evidence/test",
          "candidate_tree": "9bad791482638be26d3ab306653fd279791f449b",
          "expected_failure": "Named test fails at status app: required build_repository_source_unavailable diagnostic is absent, replaced by toolchain_mutated; Go test exit 1.",
          "log": "TASK-261016-wh6v9a_review-hosted-attempt1-rev2.md",
          "hosted_job": "https://github.com/relux-works/curator/actions/runs/39004671814/job/117314584571"
        }
      ],
      "severity": "regression",
      "priority": "P2",
      "repeat-of": "none"
    }
  ],
  "notes": [
    "Revision 1 was accepted with no findings; the brief legacy-rejection sentence is incorrect. No repeat-of mechanism exists.",
    "No Muse/SNAP-LOCK interaction defect reproduced. The convergence fix retains one policy projection in homePlan, preserving capture/status no-probe behavior. All eight swept product surfaces held for the named attacks.",
    "Coverage: 7 of 8 narrowed producer AC rows production-driven, eighth explicitly structural/boundary-only; 1 of 1 convergence interaction rows driven. Deferred consumers and qualification remain in TASK-261017-0pawkw (credential-store-policy-consumer-remainder).",
    "All focused suites and 28 narrowing controls passed independently. Free hunt exercised failed Muse shared capture and restoration; no finding reproduced.",
    "The two full-suite failures did not reproduce in local macOS whole-test replays. This is not equivalent Ubuntu evidence and does not establish an external cause. The subtest-filtered dry-run diagnostic command exited 1 from the family completeness checker and is not counted as passing.",
    "Hosted attempt 1: 19 jobs success, Ubuntu Test failure, 2 configured optional skips. Hosted attempt 2 retried only Ubuntu Test and was cancelled during go test; it supplies no passing verdict. Cancellation cause is unknown. No third attempt was launched.",
    "The local remote-gate waiter was intentionally stopped by PID before the long shell limit (exit 137); hosted CI continued independently. No script exit 0 is claimed.",
    "The findings are observed required-validation regressions, not a claim that this Muse delta introduced the unchanged Go-driver/transport behavior. Do not broaden Muse policy or weaken assertions to address them."
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
      "result": "broken",
      "findings": [
        "dryrun-go-probe-timeout",
        "legacy-source-diagnostic-preempted"
      ]
    }
  ],
  "free_hunt_findings": []
}
```

# Review sweep: revision 2

Candidate tree: `9bad791482638be26d3ab306653fd279791f449b`.
Scope: accepted-to-combined interaction; revision 1 was accepted, not rejected.
No catalog table or wall-clock budget was supplied. Retain the seven revision-1 surfaces and prepend the changed observation boundary. Original coverage is 7 of 8 producer AC rows driven, with a named structural-only eighth row; revision-2 interaction is 1 of 1 driven through CaptureMaterialize and StatusOf.

| Surface | Result | Evidence and bound |
|---|---|---|
| observation boundary / cache | held | No-probe test covers Codex and Muse native/file x shared/isolated. M-policy-probe, M-release-cache and M-probe each exit 1 with named behavioral failure; baseline exits 0. No vendor discovery or stale different-release projection reproduced. |
| destination conflicts | held | TestMusePolicySwitchPreservesCredentials, TestMusePolicyExplicitMigration: detached/shared transitions preserve credential state; explicit migration remains usable. |
| read boundary | held | TestMuseAuthParentBoundaryAtResolve, TestMuseAuthRepairAdditionalStates, published link vectors: parent/link retarget and invalid entry types refuse beside valid routes. |
| error surface | held | TestMusePublishedLinkStates, TestMuseMissingNativeFilePolicy: failed metadata observations do not become absent/file fallback; no placeholder created. |
| selector paths | held | TestMuseCLIStorePolicy and retained credential config/resolve tests: native/file controls remain admitted, keychain/unqualified selection refuses; stored selector unchanged. |
| text grammar | held | TestMuseFragmentV4PublishedCases and TestMuseFragmentBoundary: 67 pinned cases and missing/unknown/conflicting fields; direct boundary and schema scope only, not external wire ingestion. |
| identity | held | TestMuseRootContextVectors, TestMuseRootContextDrift and seven root narrowing plants: missing/changed/header-only root behavior retained. |
| ordering | held | TestMuseCLIStorePolicy and renderer-drops-only-backend plant: JSON/env/shell retain all five values and registry order; actual CLI rendering executed. |

## Free hunt

TestReviewMuseCaptureRefusalNoProbe (review-only Go overlay) drives CaptureMaterialize and StatusOf after deleting a valid shared native credential. Capture refuses with environment_home_stale before its callback; status is non-current; byte/mode/link snapshot is unchanged. Restoring the credential admits capture. Explicit Detect trap never fires. Exit 0. No blocking finding reproduced.

All twelve SNAP-LOCK controls independently rerun: twelve child exits 1, all named capture assertions, no build/setup failures counted. Together with sixteen Muse/root controls: 28 killed, zero survivors. This is bounded test sensitivity evidence, not exhaustive proof.


## Required rework

Diagnose the two named full-suite failures with their owning validation/Go-driver paths and restore a completed green hosted gate for the exact candidate. Preserve the existing named tests and frozen outcomes. If an external cause can be established, retain provider evidence and rerun the affected hosted job legitimately; local passes alone do not discharge it. This is validation rework, not a request to implement deferred consumers or change Muse policy. A subsequent reviewer cycle is required. No separate harness/research leaf is requested.

## Evidence and logbook

The exact candidate's signed gate snapshot is be87653a94f23f39acbaa3e965a8f424cf769b60 (signature G). Main/base was 612960ee90bedc7a4a113eda13aaa6cf2a2cbe9f. Product source was not edited and the Story branch was not committed or moved.

- TASK-261016-wh6v9a_review-local-rev2.md: independent commands, real exits, complete local logs, merge comparison, normative authority, swept surfaces and 28 controls.
- TASK-261016-wh6v9a_review-free-hunt-rev2.go: review-only overlay test.
- TASK-261016-wh6v9a_review-hosted-attempt1-rev2.md: both failing public-path tests and exact hosted job.
- TASK-261016-wh6v9a_review-diagnostics-rev2.md: local diagnostic replays, including the rejected subtest-filter run and limits.
- TASK-261016-wh6v9a_review-hosted-final-rev2.json: both terminal hosted attempts, snapshot identity, per-job results.

2026-10-17 review logbook: convergence behavior and narrowing controls held; independent full CI exposed two unchanged Go-driver/CLI validation failures. The only bounded hosted retry was cancelled, so no green full-gate claim or external-cause inference is made. Route to to-dev for validation diagnosis and another review, not blocked and not accepted.
