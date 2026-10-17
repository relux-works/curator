# BUG-261016-mlpvtn — self-hosted Linux CI isolation rework

This revision answers review finding `shell-cache-override-unguarded` from
BUG-261016-mlpvtn_review-verdict-rev1.md. The original production workflow is
unchanged from the reviewed candidate: all six Linux-routable job definitions
initialize fresh Go state before setup-go or any Go consumer; self-hosted Linux
cache restore/save stays disabled. The rework changes only the existing
`tools/ciworkflow` checker and its regression tests, plus the matching contract
paragraph in `docs/self-hosted-runner-setup.md`. The documentation change is
outside the checker module because the accepted inline-shell subset must be
visible to workflow maintainers. No product/host/deadline setting was changed.

The checker now reserves GOCACHE, GOMODCACHE and GOPATH in inline run bodies.
It rejects direct references after literal quote/escape/continuation removal,
including export, assignment, env, unset and environment-file writes. Reads
and comments are conservatively refused too. Computed variable names, called
script internals and third-party action internals are stated analysis bounds,
not a claim of arbitrary-shell interpretation or an AC exemption. The lexical normalization covers deletion of literal quotes, backslashes and line continuations; ANSI-C/hex-generated names and eval-built names are computed-name bounds, not demonstrated coverage.

`TestLinuxCIWorkflowRejectsShellCacheOverrides` drives the workflow checker on
27 shell override variants and an unchanged production-workflow control.
`TestLinuxCIWorkflowShellOverrideBehavior` executes the actual first Go
consumer after the real initializer, using the review's token-preserving
export plant. The actual Go binary consumes all three unsafe paths and the
same workflow body is rejected by the gate. Production call site remains
ci.yml lint -> go test -count=1 ./tools/ciworkflow/ -> TestLinuxCIWorkflow ->
checkIsolation -> checkInlineGoState.

All five earlier named tests are retained byte-for-byte. No separately
committed reviewer test was present to rename or delete.

## Coverage and stated bounds

**0 of 3 original AC rows fully driven; 6 of 6 current Linux-routable job
definitions checked.** The revision's shell-cache regression is driven by
27 of 27 literal shell variants plus the real Go consumer witness.

| Original AC | Production call site and tests | Bound |
| --- | --- | --- |
| 1 | ci.yml -> isolate-go-state.sh; TestLinuxCIWorkflow, TestGoStateInitialization, TestGoStateMissingRunnerContext; new shell regression pair | Contract and real Go env isolation established; historical ABI0 failure reproduction remains unestablished. |
| 2 | ci.yml concurrency; TestLinuxCIWorkflowRejectsNarrowingPlants | Serializes workflow/ref; cannot bound other repositories or qualify arbitrary host load. Recursion fix is included in checkpoint; host remediation is supplied by orchestrator. |
| 3 | main-push CI | Three consecutive green self-hosted main runs remain the orchestrator's post-integration obligation per revised brief; 0 of 3 established here. |

No additional AC row is excluded. The explicit AC3 rollout exemption is:
“The three consecutive green main runs are a post-integration obligation of
 the orchestrator, not a handoff prerequisite.” The literal-shell bounds
above limit this checker, not the required production isolation behavior.

Brief gap: the producer brief supplies no catalog-derived attack-surface
table. The six Review B requirements are mapped below without claiming they
are catalog row identifiers. Historical ABI0 reproduction and host-wide load
qualification remain unestablished, as in the previous reviewed evidence.

## Probe deadline finding

internal/install/builddeps.go -> ConfigFromEnvironment -> godriver.Probe.
internal/godriver/session.go clamps the probe to 15 seconds and toolchain
fingerprinting to two minutes. ConfigFromEnvironment admits only trusted
selectors; there is no documented CI timeout knob. No new flag or deadline
change was introduced. The candidate contains Landlock recursion fix 6e85a912.

## Review B coverage map

| Review requirement | Tests / evidence | Narrowing attack and result | Scope limit |
| --- | --- | --- | --- |
| B1 private state before Go | TestLinuxCIWorkflow; TestGoStateInitialization; TestGoStateMissingRunnerContext; TestLinuxCIWorkflowRejectsShellCacheOverrides; TestLinuxCIWorkflowShellOverrideBehavior | One skipped initializer, shared setup cache, shared module cache, review export body and weakened module-name guard all killed | Literal inline-shell contract; script/action internals reviewed separately. |
| B2 workflow/ref concurrency, push protection | TestLinuxCIWorkflowRejectsNarrowingPlants | cancel-main and split-ref killed by TestLinuxCIWorkflow | Does not bound other refs/repos or jobs within a run. |
| B3 fork/platform/Race/pins | TestLinuxRoutingKeepsForksHosted; exact before/after comparison of routing/pins and untouched jobs | Token-preserving fork-bypass killed by routing test | Executes JS-compatible expression subset; actual scheduling evidenced separately. |
| B4 real pre-fix control | TestLinuxCIWorkflow on checkpoint workflow | Expected exit 1: private Go state must precede every Go consumer; candidate package exit 0 | Does not reproduce ABI0 linker failure. |
| B5 green self-hosted remote gate | Fresh full remote-gate run and job API, exact signed snapshot identity | Runner provenance is provider evidence, not a new product gate or mutant claim | Fresh remote-gate exit 0; 20 successful jobs, including all 8 self-hosted Linux instances; 2 optional skips. |
| B6 public text hygiene | Added-line scan plus manual full diff inspection | No source behavior introduced for this property; no narrowing mutant claimed | Old untouched historical text is outside changed scope; raw runner registrations kept private. |

## Narrowing mutants

Every listed mutant ran the **entire** tools/ciworkflow package, including
initializer and Go-consumer behavioral tests, using a workflow/script fixture
or Go build overlay. The gate stayed present in all cases. All commands below
exited **1 (expected red)** for the named intended failure. No survivors among
these 11 executed mutants; this does not claim an exhaustive shell analysis.
The harness itself exited 0 only after validating all intended nonzero exits.

| Mutant | What it narrows the gate to | Named failing test | Result / survival bound |
| --- | --- | --- | --- |
| echo-init | One job retains initializer token but only echoes it | TestLinuxCIWorkflow | Killed |
| restore-cache | One setup-go restores shared cache; original condition remains in comment | TestLinuxCIWorkflow | Killed |
| cancel-main | Only dispatches protected from cancellation | TestLinuxCIWorkflow | Killed |
| split-ref | Concurrency admits overlap by adding run_id | TestLinuxCIWorkflow | Killed |
| skip-push-gate | Push bypasses the still-present isolation test | TestLinuxCIWorkflow | Killed |
| fork-bypass | OR admits fork PR while keeping original guard token | TestLinuxRoutingKeepsForksHosted | Killed |
| shell-override | First real consumer exports shared caches; initializer and all existing tokens remain | TestLinuxCIWorkflow | Killed; full behavioral suite ran |
| shared-module-cache | Only GOMODCACHE is shared beneath RUNNER_TEMP | TestGoStateInitialization | Killed; real initializer executes |
| checker-admits-cache | Guard accepts cache:true while other checks remain | TestLinuxCIWorkflowRejectsNarrowingPlants/one_setup_restores_shared_cache | Killed |
| checker-admits-module-override | Reserved-name gate still protects GOCACHE/GOPATH, admits only GOMODCACHE | TestLinuxCIWorkflowRejectsShellCacheOverrides/GOMODCACHE/export | Killed |
| checker-admits-split-name | Keep protected tokens but weaken quote normalization | TestLinuxCIWorkflowRejectsShellCacheOverrides/GOCACHE/split_name | Killed |

The separate pre-fix control also exited 1 in TestLinuxCIWorkflow, for missing
private initialization. No expected-red command is presented as a pass.
Mutant definitions and command/exit records accompany this handoff; local raw
logs stay in ignored task scratch because they may contain personal paths.

## Fresh local validation

| Standalone command | Real exit |
| --- | ---: |
| go test -count=1 -timeout=3m -v ./tools/ciworkflow/ | 0 |
| go build ./... | 0 |
| go vet ./tools/ciworkflow/ | 0 |
| actionlint -config-file <existing custom runner label allowlist> .github/workflows/ci.yml | 0 |
| bash -n .github/ci/isolate-go-state.sh | 0 |
| git diff --check | 0 |
| Each of 11 mutant full-package go test commands | 1, intended rejection |
| Pre-fix control full-package go test command | 1, intended missing-private-state rejection |

Go 1.25.5 darwin/arm64; Node v25.6.1; actionlint 1.7.12. Full remote CI is run
fresh because the checker/test/documentation identity changed. Prior full-gate
success is not substituted. No previous local test result is substituted for
the seven current named top-level tests. No test/library dependency was added.

## Candidate files

- .github/workflows/ci.yml
- .github/ci/isolate-go-state.sh
- tools/ciworkflow/isolation_test.go
- docs/self-hosted-runner-setup.md
- README.md
- UNRESOLVED_QUESTIONS.md

## Remote gate: success

The fresh standalone `sh scripts/remote-gate.sh` command exited **0**.
Run [39016112959](https://github.com/relux-works/curator/actions/runs/39016112959)
completed/success on signed snapshot
`5e1d61470ec525b73c216f50c531cf407cffd93e` (signature G).
All **20 executed jobs passed**, including **8 self-hosted Linux instances**;
2 optional lanes were skipped. All eight Linux instances successfully ran the
private-state initializer. Hosted macOS/Windows Test and both hosted Race
jobs passed. No cold-cache probe-deadline failure was observed in this run.
This does not establish arbitrary-loaded-host qualification.

The full candidate tree equals the remote snapshot tree, not merely selected
files. Story HEAD remains checkpoint
`a17e58156ac30a1629e3ff36efa259bab1555a15`; the candidate is UNCOMMITTED.
The process stayed live in exec session 28040 until its real exit, with
bounded polling calls; nothing was detached or left running at handoff.

Exact per-job registration names remain in ignored task scratch at
`rework-2/private-runner-map.tsv` and the authoritative GitHub job records.
Public board text uses neutral aliases because registrations disclose host
identifiers. The following mapping covers all executed jobs.

| Job | Provider job | Runner alias | Pool | Result |
| --- | --- | --- | --- | --- |
| Test (windows-latest) | [117348444502](https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444502) | runner-dcec810569 | hosted | success |
| Gate self-test (windows-latest) | [117348444548](https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444548) | runner-6d52d8b4db | hosted | success |
| Gate self-test (macos-latest) | [117348444594](https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444594) | runner-6f2f4fc4dc | hosted | success |
| Test (macos-latest) | [117348444603](https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444603) | runner-7ef5cdc29b | hosted | success |
| Naming gate | [117348444644](https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444644) | runner-c55353a08c | self-hosted Linux | success |
| Race (macos-latest) | [117348444646](https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444646) | runner-29963431fe | hosted | success |
| Go 1.27.1 driver (ubuntu-latest) | [117348444653](https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444653) | runner-6dd10446cf | self-hosted Linux | success |
| Gate self-test (ubuntu-latest) | [117348444657](https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444657) | runner-5ba02dc1e7 | self-hosted Linux | success |
| Go 1.26.0 driver (windows-latest) | [117348444669](https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444669) | runner-1e347592ef | hosted | success |
| Race (ubuntu-latest) | [117348444685](https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444685) | runner-550dfc0c08 | hosted | success |
| Go 1.27.1 driver (macos-latest) | [117348444689](https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444689) | runner-0f393e3742 | hosted | success |
| Interop conformance gate | [117348444712](https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444712) | runner-c55353a08c | self-hosted Linux | success |
| Test (ubuntu-latest) | [117348444790](https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444790) | runner-4d6e1c9351 | self-hosted Linux | success |
| Go 1.26.0 driver (ubuntu-latest) | [117348444796](https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444796) | runner-6dd10446cf | self-hosted Linux | success |
| Lint | [117348444831](https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444831) | runner-5ba02dc1e7 | self-hosted Linux | success |
| Go 1.25.5 driver (windows-latest) | [117348444911](https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444911) | runner-50a7478ab9 | hosted | success |
| Go 1.26.0 driver (macos-latest) | [117348444936](https://github.com/relux-works/curator/actions/runs/39016112959/job/117348444936) | runner-b7faed0803 | hosted | success |
| Go 1.25.5 driver (ubuntu-latest) | [117348445000](https://github.com/relux-works/curator/actions/runs/39016112959/job/117348445000) | runner-c55353a08c | self-hosted Linux | success |
| Go 1.25.5 driver (macos-latest) | [117348445012](https://github.com/relux-works/curator/actions/runs/39016112959/job/117348445012) | runner-2f60fe2bd8 | hosted | success |
| Go 1.27.1 driver (windows-latest) | [117348445085](https://github.com/relux-works/curator/actions/runs/39016112959/job/117348445085) | runner-0022734e54 | hosted | success |

The initial rework logbook was attached while the aggregate gate was pending;
this terminal packet supersedes that pending observation without rewriting
its history. The “Implementation matches AC” checklist records conformance to the
current revised producer brief: private Go state, workflow/ref concurrency,
negative control and fork protection, recorded fixed probe bounds, runner
documentation, review regression plus narrowing mutant, and a fresh green
remote gate on the converged candidate. It does not attest all three original
AC rows as fully driven; that measured ratio remains 0 of 3 for the explicitly
stated bounds above. Candidate review is admitted by revised brief section 2;
three green main runs remain an orchestrator obligation after integration.
No original AC coverage is silently waived or inferred from this single run.
