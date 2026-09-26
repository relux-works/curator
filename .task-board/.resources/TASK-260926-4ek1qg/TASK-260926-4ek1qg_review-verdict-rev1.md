# TASK-260926-4ek1qg review verdict — CR rev1: CHANGES REQUESTED

Reviewed candidate tree 207e7559 against base 57463678 (4 files). Also checked PR #88 head b68a937 (b202b5d + b68a937).

## What's fine
- Go pin 0a628621 exists; comment is accurate. The sparse checkout (`/*`, `!/.task-board/`, non-cone) keeps `internal/` and the submodules.
- Go names at 0a628621 (checked with git grep on curator): TestScriptHostExecutionPolicyProductionConsumersCoverAllCases (conformance_test.go:370) and TestPreflightRefusalCases (:741) exist. TestARefusalPrecedesEveryWorkerSurface is gone (at a3abcf34 it was at :223).

## Blocking: hosted Implementations (ubuntu-latest) FAILS on head b68a937
Run 36206606817, job 108304365393:
```
FAILED tests/test_schema8_candidate_conformance.py::test_script_policy_sections_are_all_classified -
AssertionError: the root publishes sections this build does not classify:
['executable_identity_cases', 'hard_link_substitution_definition']
##[error]Process completed with exit code 1.
```
The Python manager lane (`.github/workflows/implementations.yml`, step "Checkout Python manager", pin unchanged) does not classify the two sections added by #88. The AC says "the Implementations job steps pass". The producer only proved the Go steps, so the job is not green.
macOS and Windows jobs were still pending when this was checked, so the Windows sparse checkout is still unverified.

## Required rework
1. Make the manager lane pass with #88 content. Either pin a manager commit that classifies both sections, or get a coordinated decision for that lane. Don't leave it red.
2. Re-run the PR #88 checks and cite a green Implementations run on all three OSes, including windows-latest.
3. Secondary: the ledger claim for TestPreflightRefusalCases ("mandatory-control-unavailable ... shapes") is narrower than the old "every enforced shape refused before any worker surface". State in the results that this narrowing is deliberate, or map the broader claim to its successor as well.
