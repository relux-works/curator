# BUG-261016-mlpvtn \u2014 self-hosted Linux CI isolation

Producer validation after convergence; candidate remains UNCOMMITTED at Story
checkpoint `a17e58156ac30a1629e3ff36efa259bab1555a15`. This checkpoint contains
Landlock recursion fix `6e85a912`. No Story branch commit, switch, rebase, merge,
host change, product deadline change, repository-variable write, or test skip
was made in this run. The existing repository variable was already true.
The normal `sh scripts/remote-gate.sh` entry point created signed detached
snapshot `3cf597414f1285714c976361ff1a8d6f72de725c` (verified G signature).
Every changed file matches that snapshot byte-for-byte.

## Candidate scope

- `.github/workflows/ci.yml`: initialize private Go state immediately after
  checkout in six Linux-routable job definitions, before every Go consumer;
  disable setup-go and lint-action caching on self-hosted Linux; serialize
  workflow/ref and cancel superseded pull requests only.
- `.github/ci/isolate-go-state.sh`: fresh mktemp directory beneath RUNNER_TEMP,
  GOCACHE/GOMODCACHE/GOPATH/private lint state published through GITHUB_ENV.
- `tools/ciworkflow/isolation_test.go`: workflow gate and narrowing plants,
  actual initializer execution and Go environment consumption, missing-context
  rejection without partial state, executed fork/trusted routing expressions.
- `docs/self-hosted-runner-setup.md`: isolation contract, scheduling limits,
  fixed probe deadlines and rollout obligation.
- `README.md` and `UNRESOLVED_QUESTIONS.md`: tool entry points and open rollout bounds.

Implementation preserved from the preceding run; this run adds no repository
source delta. SPEC_PIN, PNPM_PIN, GOTOOLCHAIN=local, GOENV=off, hosted
macOS/Windows routing, Race routing and the fork guard retain the pre-fix values.
This is resumed producer work, not reviewer rework. No committed reviewer test
was deleted or renamed. The initializer, test package, README and unresolved
question entry are supporting changes to the requested workflow/docs scope;
there is no unrelated module change.

## Coverage and bounds

**0 of 3 original AC rows fully driven; 6 of 6 Linux-routable job definitions
checked.** This reports the original AC literally; it does not turn a structural
control into a reproduction of the historical linker failure.

| AC | Production entry point / named tests | Evidence and bound |
| --- | --- | --- |
| 1 | ci.yml -> isolate-go-state.sh; TestLinuxCIWorkflow, TestGoStateInitialization, TestGoStateMissingRunnerContext | go-family-qualification, test, lint, gate-selftest, interop and naming-gate checked. Pre-fix workflow fails the contract check; fixed workflow passes. Two initializer invocations use disjoint paths consumed by go env despite poisoned ambient caches. Historical ABI0 linker-race reproduction remains unestablished. |
| 2 | ci.yml concurrency; TestLinuxCIWorkflow, TestLinuxCIWorkflowRejectsNarrowingPlants | Same workflow/ref serialized, only PR cancellation admitted. This cannot bound other repositories or qualify arbitrarily loaded hosts. The revised brief reports the host incident resolved and task caps installed; these host facts are supplied by the orchestrator, not independently measured here. |
| 3 | main push CI | 0 of 3 consecutive self-hosted main runs established by this producer; post-integration obligation belongs to the orchestrator. |

Out of producer handoff scope: AC3's rollout sequence, explicitly assigned by
brief revision: \u201cThe three consecutive green main runs are a post-integration
obligation of the orchestrator, not a handoff prerequisite.\u201d No other AC row
is excluded; AC1 historical race and AC2 host-wide load remain stated bounds.
Revision 2 explicitly admits handoff after a green full remote gate on the
converged snapshot.

Brief gap: no formal attack-surface table was supplied; a catalog-derived
coverage map is not claimed. Current tests cover the present workflow shape
and JS-compatible expression subset, not arbitrary future Actions syntax or
third-party action internals. TestLinuxRoutingKeepsForksHosted drives actual
workflow expressions with Node for fork/trusted PRs, pushes, dispatches and
matrix operating systems. Real scheduling is recorded by the remote jobs.
The isolation check is invoked by the unconditional lint step
`go test -count=1 ./tools/ciworkflow/`.

## Probe deadline finding

`internal/install/builddeps.go` calls ConfigFromEnvironment and godriver.Probe.
`internal/godriver/session.go` fixes probe time at at most 15 seconds and
fingerprinting at at most two minutes; larger Config values are clamped.
ConfigFromEnvironment reads trusted toolchain selectors, with no documented
CI override for those bounds. No product timeout knob was invented.
The earlier failures are attributed by the revised brief to the recursing
Landlock probe on the old base, not to a demonstrated cold-cache-only problem.
This snapshot includes the separately landed recursion fix.

## Narrowing mutants and control

The pre-fix control exited **1, expected red** in TestLinuxCIWorkflow because
private state was absent. Fixed workflow exits 0. This demonstrates the check,
not an actual ABI0 reproduction. Exact source/test/dependency identities and
all reused log digests were verified against the previously attached
BUG-261016-mlpvtn_blocked-evidence.md. Go 1.25.5 darwin/arm64, Node v25.6.1,
workflow-lint configuration, initializer and test bytes are unchanged.
No old full-suite or product build result is reused for the converged base.


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


## Fresh validation

All commands ran directly as standalone processes; no tee or masked gate exit.

| Command | Exit | Evidence |
| --- | ---: | --- |
| go test -count=1 -timeout=3m -v ./tools/ciworkflow/ | 0 | Five named top-level tests and narrowing subtests green; ciworkflow.log |
| go build ./... | 0 | Fresh full build after convergence; build.log |
| actionlint -config-file <existing custom-label allowlist> .github/workflows/ci.yml | 0 | actionlint.log, actionlint 1.7.12 |
| bash -n .github/ci/isolate-go-state.sh | 0 | bash-syntax.log |
| git diff --check | 0 | diff-check.log |

Prior narrowing commands are reused, not replayed: all eight exited 1 with the
named intended failure, and the pre-fix control separately exited 1. Reuse
integrity is recorded in prior-evidence-integrity.json and source-identity.json.
Logs are stored in ignored task-local retry-3 artifacts. Raw logs and exact
runner registrations may contain host identifiers or personal paths and are
kept privately; public board resources use neutral runner aliases and direct
GitHub job links. The job API preserves exact registration names independently.

## Remote gate: green

`sh scripts/remote-gate.sh` exited **0** on the first attempt.
Run [39008448258](https://github.com/relux-works/curator/actions/runs/39008448258)
is completed/success on snapshot `3cf597414f1285714c976361ff1a8d6f72de725c`:
**20 successful jobs, 0 failures, 2 expected optional-lane skips**.
All eight self-hosted Linux job instances passed, including all three Go
families, the main test gate, lint, gate self-tests, interop and naming.
Both hosted Race jobs and the hosted macOS/Windows matrix also passed.
The full gate was run afresh by this producer after convergence; an older
aggregate result is not substituted. No cold-cache deadline failure was
observed in this run. This does not establish arbitrary-loaded-host behavior.
Windows used its existing unchanged 120m per-package suite limit.

The final API read independently confirmed completed/success and the exact
snapshot SHA. Current changed-file bytes and the unchanged Story checkpoint
were verified again after the gate. The gate branch was cleaned by the normal
entry point; the Story candidate remains uncommitted for snapshot handoff.

Exact per-job runner registrations are preserved privately in the ignored
`retry-3/BUG-261016-mlpvtn_runner-map.tsv` task artifact and in the linked GitHub
job records. They are not copied into this public board because registration
names disclose host identifiers. The following aliases identify registration
strings without publishing those names.

| Job | GitHub job | Runner alias | Lane | Verdict |
| --- | --- | --- | --- | --- |
| Race (macos-latest) | [117325626778](https://github.com/relux-works/curator/actions/runs/39008448258/job/117325626778) | runner-4a884fbb82 | hosted | success |
| Gate self-test (macos-latest) | [117325626844](https://github.com/relux-works/curator/actions/runs/39008448258/job/117325626844) | runner-097307ae08 | hosted | success |
| Naming gate | [117325626851](https://github.com/relux-works/curator/actions/runs/39008448258/job/117325626851) | runner-5ba02dc1e7 | self-hosted Linux | success |
| Lint | [117325626869](https://github.com/relux-works/curator/actions/runs/39008448258/job/117325626869) | runner-5ba02dc1e7 | self-hosted Linux | success |
| Race (ubuntu-latest) | [117325626910](https://github.com/relux-works/curator/actions/runs/39008448258/job/117325626910) | runner-7af87a6716 | hosted | success |
| Test (ubuntu-latest) | [117325626914](https://github.com/relux-works/curator/actions/runs/39008448258/job/117325626914) | runner-c55353a08c | self-hosted Linux | success |
| Interop conformance gate | [117325626938](https://github.com/relux-works/curator/actions/runs/39008448258/job/117325626938) | runner-c55353a08c | self-hosted Linux | success |
| Go 1.27.1 driver (macos-latest) | [117325626948](https://github.com/relux-works/curator/actions/runs/39008448258/job/117325626948) | runner-468d7c866a | hosted | success |
| Go 1.25.5 driver (ubuntu-latest) | [117325626961](https://github.com/relux-works/curator/actions/runs/39008448258/job/117325626961) | runner-5ba02dc1e7 | self-hosted Linux | success |
| Go 1.26.0 driver (ubuntu-latest) | [117325626966](https://github.com/relux-works/curator/actions/runs/39008448258/job/117325626966) | runner-c55353a08c | self-hosted Linux | success |
| Go 1.27.1 driver (ubuntu-latest) | [117325626968](https://github.com/relux-works/curator/actions/runs/39008448258/job/117325626968) | runner-c55353a08c | self-hosted Linux | success |
| Go 1.26.0 driver (windows-latest) | [117325626971](https://github.com/relux-works/curator/actions/runs/39008448258/job/117325626971) | runner-d6e77b9a16 | hosted | success |
| Go 1.26.0 driver (macos-latest) | [117325626977](https://github.com/relux-works/curator/actions/runs/39008448258/job/117325626977) | runner-98eb28b3da | hosted | success |
| Go 1.27.1 driver (windows-latest) | [117325626983](https://github.com/relux-works/curator/actions/runs/39008448258/job/117325626983) | runner-dce9803240 | hosted | success |
| Gate self-test (ubuntu-latest) | [117325627001](https://github.com/relux-works/curator/actions/runs/39008448258/job/117325627001) | runner-5ba02dc1e7 | self-hosted Linux | success |
| Go 1.25.5 driver (windows-latest) | [117325627005](https://github.com/relux-works/curator/actions/runs/39008448258/job/117325627005) | runner-187ae02e6d | hosted | success |
| Gate self-test (windows-latest) | [117325627016](https://github.com/relux-works/curator/actions/runs/39008448258/job/117325627016) | runner-b5b28c5df5 | hosted | success |
| Test (macos-latest) | [117325627025](https://github.com/relux-works/curator/actions/runs/39008448258/job/117325627025) | runner-d5febfd69a | hosted | success |
| Go 1.25.5 driver (macos-latest) | [117325627079](https://github.com/relux-works/curator/actions/runs/39008448258/job/117325627079) | runner-8a9b316ef7 | hosted | success |
| Test (windows-latest) | [117325627090](https://github.com/relux-works/curator/actions/runs/39008448258/job/117325627090) | runner-c32bf36a65 | hosted | success |

## Log integrity

Fresh command logs (SHA256); their corresponding real exits are recorded in the validation table and adjacent .exit records.

```json
{
  "remote-gate.log": "c43bb28f79da5b4514bad5fa7cb02e189130887e98e6d96b527355df63162071",
  "ciworkflow.log": "0b4072755840e9a89f47dd0752c539ec0f929f2fe3cbae89084da0faadf2d0e5",
  "build.log": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "actionlint.log": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "bash-syntax.log": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "diff-check.log": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
}
```

Post-integration owner obligation: obtain three consecutive green self-hosted main runs before declaring rollout qualified. This producer makes no main-run qualification claim.
