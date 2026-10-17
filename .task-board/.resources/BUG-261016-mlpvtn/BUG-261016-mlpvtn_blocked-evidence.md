# BUG-261016-mlpvtn \u2014 self-hosted Linux CI isolation: blocked evidence

The candidate remains UNCOMMITTED on the managed Story branch at checkpoint
`612960ee90bedc7a4a113eda13aaa6cf2a2cbe9f`. No Story commit, switch, rebase,
merge, host mutation, or repository-variable write was performed by this successor.
A developer review handoff is NOT admitted: the required remote gate is not green.

## Candidate and scope

- `.github/workflows/ci.yml`: initialize private state immediately after checkout
  in all six Linux-routable jobs; disable setup-go caching and lint action caching
  on self-hosted Linux; serialize workflow/ref and cancel only PR events.
- `.github/ci/isolate-go-state.sh`: mktemp directory under RUNNER_TEMP, exporting
  GOCACHE, GOMODCACHE, GOPATH and lint cache through GITHUB_ENV.
- `tools/ciworkflow/isolation_test.go`: production workflow gate, negative
  workflow fixtures, real shell initialization and Go environment consumption,
  absent-context rejection, and expression execution for fork routing.
- `docs/self-hosted-runner-setup.md`, `README.md`, `UNRESOLVED_QUESTIONS.md`:
  isolation contract, invocation, deadline finding and unresolved rollout bounds.

Changes recovered from the predecessor were preserved. This is the first producer
attempt, not reviewer rework; there are no newly removed/renamed reviewer tests.
Existing gate scripts and test names were preserved. Scope outside the requested
workflow/docs module is the initializer, its Go tests, required README tools
entry, and unresolved-questions record; each supports this workflow change.
SPEC_PIN, PNPM_PIN, GOTOOLCHAIN=local, GOENV=off, hosted macOS/Windows routing,
the fork guard, and Race routing retain their pre-fix values.

## Coverage and bounds

**0 of 3 AC rows fully driven.** AC1's isolation sub-contract is driven, but its
historical ABI0 reproduction requirement is not. No AC row is declared out of
contract: the remaining obligations below stay open.

| AC | Production entry / named tests | Evidence and remaining bound |
| --- | --- | --- |
| 1 | ci.yml -> isolate-go-state.sh; TestLinuxCIWorkflow, TestGoStateInitialization, TestGoStateMissingRunnerContext | 6 of 6 current LINUX_SELF_HOSTED-routed jobs checked: go-family-qualification, test, lint, gate-selftest, interop, naming-gate. Pre-fix plant fails; private state and no shared module cache verified. Actual historical ABI0 linker failure was not reproduced; a contract-check failure is not that reproduction. |
| 2 | ci.yml concurrency; TestLinuxCIWorkflow and TestLinuxCIWorkflowRejectsNarrowingPlants | Same workflow/ref overlap bounded; no cross-repository or whole-host scheduling guarantee. Loaded-host qualification is unproven; real fingerprint deadline failure remains. |
| 3 | GitHub main push workflow | 0 of 3 required consecutive green self-hosted main runs established for this candidate; needs integration and orchestrator rollout. Producer does not claim that changing the variable early satisfies AC3. |

Brief gap: no formal surface table was supplied, so a catalog-derived coverage
map was not available. This is explicit, not an inferred waiver. Static checking
covers the present workflow expression shape, not every future GitHub Actions
syntax/action behavior. GitHub scheduling is evidenced remotely, not by Node.

## Probe finding and stop boundary

`internal/install/builddeps.go:183,191` calls ConfigFromEnvironment and
`godriver.Probe`. `internal/godriver/session.go` fixes the probe bound at 15s
and the fingerprint bound at 2m; larger Config values are clamped.
ConfigFromEnvironment reads CURATOR_GO/GOROOT only. No existing documented CI
knob can increase these bounds. No product flag, skip, deadline increase, or
alternate hosted routing was introduced.

Remote run [38991835985](https://github.com/relux-works/curator/actions/runs/38991835985)
ran source/test/workflow snapshot `9205dea4ae3423c621938896ef24968e177cca20`.
Go 1.26 Linux job [117275706601](https://github.com/relux-works/curator/actions/runs/38991835985/job/117275706601)
failed with exit 1 in TestRealTrustedGoProbe after 120.38s:
`go-v1 toolchain_timeout: toolchain fingerprint deadline exceeded`.
This is an observed fixed-deadline failure; current pool load/filesystem cause
is unproven. Workflow/ref concurrency cannot reserve host-wide capacity.

The candidate source, tests and workflow match that snapshot byte-for-byte.
Two subsequent documentation additions record this fingerprint failure; they
were not included in that remote snapshot. There is no exact-current-tree green
remote-gate claim. The predecessor's remote-gate shell log stops before a terminal
exit: its aggregate result is UNKNOWN/unrun as acceptance evidence. The older
aborted attempt exited 137 and is not green evidence.

Successor attempted `gh run rerun 38991835985 --job 117275706601`; exit 1,
`job 117275706601 cannot be rerun` while the overall workflow was still active.
No full suite was replayed merely to seek a lucky green, and no other operator's
job or process was cancelled. The remaining Windows job was still in progress
at the latest observation; no passing result is inferred for it.

Linux job runner mapping is captured in ignored task-local `remote-jobs-03.tsv`;
exact runner names are intentionally not copied into this public board resource
because they expose host identifiers. The GitHub job records preserve the
mapping independently. Linux Test, Lint, Gate self-test, Interop, Naming and
Go 1.25/1.27 qualification jobs reported success. These are individual job
results, not an aggregate passing gate.

Required external input: owner authorization for pool capacity/filesystem
investigation and remediation outside the workflow-only brief, or an explicit
product deadline-policy/qualification decision. Recommend the capacity path
first, retaining the current trust/deadline policy. After resolution, rerun the
failed qualification, obtain a genuinely green candidate gate, then route review
and the three-main-run rollout through the orchestrator. Hosted routing and a
new product timeout flag are not permitted substitutes.

## Narrowing mutants

All listed red commands exited 1 with the intended named failure. No survivor
was observed among these eight executed mutants; this is not an exhaustive
mutation claim. The pre-fix control is additional to the mutants.

| Mutant | What it narrows the gate to | Named failing test | Result / survival bound |
| --- | --- | --- | --- |
| Echo initialization command in one job | Five jobs execute initialization; one only retains the command token | TestLinuxCIWorkflow | Killed |
| One setup-go cache:true with original condition retained in comment | Allow shared cache restore in one job | TestLinuxCIWorkflow | Killed |
| Cancel on pushes as well as PRs | Protect only dispatches from cancellation | TestLinuxCIWorkflow | Killed |
| Add run_id to concurrency group | Serialize only a single run, allowing same-ref overlap | TestLinuxCIWorkflow | Killed |
| Invoke isolation test only on PR | Gate remains but pushes bypass it | TestLinuxCIWorkflow | Killed |
| Checker accepts cache:true | Cache guard remains but admits one unsafe cache value | TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache | Killed |
| Initializer shares only GOMODCACHE under RUNNER_TEMP | Other state stays private; same-runner jobs share module cache | TestGoStateInitialization | Killed by executing the real shell entry point |
| Add PR bypass OR while preserving original fork-guard token | Fork PR reaches self-hosted labels in one job | TestLinuxRoutingKeepsForksHosted | Killed by executing the actual expression with Node |

The last two behavioral mutants preserve the source tokens used by structural
checks and execute the behavioral suite, rather than only the static checker.

## Validation

| Command | Exit | Result |
| --- | ---: | --- |
| `go test -count=1 -timeout=3m -v ./tools/ciworkflow/` | 0 | All named isolation/routing tests passed. |
| `go build ./...` | 0 | Full project build. |
| `go vet ./tools/ciworkflow/` | 0 | Changed Go package. |
| `actionlint -config-file .temp/BUG-261016-mlpvtn/actionlint.yaml .github/workflows/ci.yml` | 0 | Existing custom-label allowlist; actionlint 1.7.12. |
| `shellcheck .github/ci/isolate-go-state.sh` | 0 | Initializer shell lint. |
| `bash .github/ci/gate-selftest.sh` | 0 | Configured agent Git environment; 267 passed, 0 failed; 2 rustup-absent host-dependent cases explicitly skipped. |
| `git diff --check` | 0 | Whitespace validation. |
| `CI_WORKFLOW_UNDER_TEST=<pre-fix fixture> go test -count=1 ./tools/ciworkflow/ -run ^TestLinuxCIWorkflow$` | 1 | EXPECTED RED: missing private state; fixture equals checkpoint workflow. |
| `CI_ISOLATION_SCRIPT_UNDER_TEST=<shared-module fixture> go test -count=1 ./tools/ciworkflow/ -run ^TestGoStateInitialization$` | 1 | EXPECTED RED: two jobs share GOMODCACHE. |
| `CI_WORKFLOW_UNDER_TEST=<fork-bypass fixture> go test -count=1 ./tools/ciworkflow/ -run ^TestLinuxRoutingKeepsForksHosted$` | 1 | EXPECTED RED: fork reaches self-hosted routing. |
| `CI_WORKFLOW_UNDER_TEST=<current workflow> go test -count=1 ./.temp/BUG-261016-mlpvtn/checker-mutant/ -run ^TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache$` | 1 | EXPECTED RED: narrowed checker admits cache:true. |
| `CI_WORKFLOW_UNDER_TEST=<narrow-init.yml> go test -count=1 ./tools/ciworkflow/ -run ^TestLinuxCIWorkflow$` | 1 | EXPECTED RED: named TestLinuxCIWorkflow failure, not a setup error. |
| `CI_WORKFLOW_UNDER_TEST=<narrow-cache.yml> go test -count=1 ./tools/ciworkflow/ -run ^TestLinuxCIWorkflow$` | 1 | EXPECTED RED: named TestLinuxCIWorkflow failure, not a setup error. |
| `CI_WORKFLOW_UNDER_TEST=<narrow-cancel.yml> go test -count=1 ./tools/ciworkflow/ -run ^TestLinuxCIWorkflow$` | 1 | EXPECTED RED: named TestLinuxCIWorkflow failure, not a setup error. |
| `CI_WORKFLOW_UNDER_TEST=<narrow-group.yml> go test -count=1 ./tools/ciworkflow/ -run ^TestLinuxCIWorkflow$` | 1 | EXPECTED RED: named TestLinuxCIWorkflow failure, not a setup error. |
| `CI_WORKFLOW_UNDER_TEST=<narrow-invocation.yml> go test -count=1 ./tools/ciworkflow/ -run ^TestLinuxCIWorkflow$` | 1 | EXPECTED RED: named TestLinuxCIWorkflow failure, not a setup error. |

Fresh successor commands above ran directly, without tee or status-masking pipelines.
Local environment: Go 1.25.5, darwin/arm64; Node v25.6.1. Remote Linux lint
was green on the unchanged source/test/workflow snapshot; the full local lint
job was not replayed. Local Go build is not loaded-Linux qualification.

Source SHA256 identity:
```json
{
  ".github/workflows/ci.yml": "f2d5daededfda50de425aa790338e8118334ac9002790df1b3fef45940c0378e",
  ".github/ci/isolate-go-state.sh": "1647682f0de005c2cc2def56f8ee683d2d58ce070ce20b314cf48287c6637cf0",
  "tools/ciworkflow/isolation_test.go": "89406e206e340dd7746e3d5b1f9fab2103067e713e50c5295d95188ada9b8dc6",
  "README.md": "9aa2153c622311bff029c1c556b8ca15224f321c6579ebd7cb95f5d3bc347bf9",
  "docs/self-hosted-runner-setup.md": "d68f0c3b525112022d34dfe1bbc2b6b4569268e2bbc2db4826bbaa8d97f40f28",
  "UNRESOLVED_QUESTIONS.md": "73fb96220957c8c29789de3c2e30ec31c14b474fef5406034e6779dc21058a2d"
}
```

Validation log digests:
```json
{
  "ciworkflow-tests-04.log": "aeaba0c91d7234e450bc066113bad1777c15d297e6dcd1fd3ee98703bb27ccba",
  "go-build-all-02.log": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "go-vet-02.log": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "actionlint-04.log": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "shellcheck-03.log": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "gate-selftest-03.log": "c93a0ee72c60a3fd5c6037cbb35e7c2e0e742d589ba95cc671deb92d8bdac0d6",
  "diff-check-05.log": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "control-pre-fix-03.log": "f9310079f30cf6fa94caa7175be3c9526bfc865aa1cd9f0c32cb44fec53d4c09",
  "shared-mod-mutant-02.log": "b9fba538ee3d4b069b0dcadc9df70299fd7e54e6efbc9c9c879e1955803dc8be",
  "fork-mutant-03.log": "359eaf01a00fcd9a77da94cf231d28cb843f060a48a9e8d865cd630935e1ca09",
  "checker-mutant-03.log": "920bf2c3dc7b76c091c160d44956810eabe6c6ae4b01423f97c2251ce87bd02f",
  "narrow-init-01.log": "a611962f1da2d5891b68096535a2a2d8c3611925a5058fecaede5087f678e0cd",
  "narrow-cache-01.log": "959742889f943ffd36a9c97350330ce5cfbaf61634d2dd4eb67c88502a1c6a5e",
  "narrow-cancel-01.log": "d8befb53a432a0685234db7a57be71add4974183c792ef6e9c438985a9739e10",
  "narrow-group-01.log": "7d9d32d49f998ce693caf33b33685b1681dae5c2a57df1abf16eb67ef7c49bd2",
  "narrow-invocation-01.log": "87cc58325687ce4fcd444247a1f4250800f679017ef9a1d5e502dbaecb1ac0e4"
}
```

Logs remain under ignored `.temp/BUG-261016-mlpvtn/`; no raw runner log containing
host paths or identifiers was attached to the public board. The evidence resource
and separate task-scoped logbook record persist the decision and limits.
