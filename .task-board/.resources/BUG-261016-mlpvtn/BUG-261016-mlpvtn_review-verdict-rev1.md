# Review B: BUG-261016-mlpvtn — self-hosted-linux-ci-flaky

Candidate: f4be4e3a8887893d2e90374ad487588a9d43f293; base: a17e58156ac30a1629e3ff36efa259bab1555a15. CR revision 1. Review is sequential; no product code changed.

Coverage inspected before code: producer reports 0 of 3 original AC rows fully driven, 6 of 6 current Linux job definitions checked. The historical ABI0 race is explicitly unestablished; host-wide qualification is bounded; three main successes belong to post-integration rollout under the revised brief. These bounds are not misrepresented as passes. Review B supplies six surfaces; no numeric review/free-hunt budget was supplied.

## Surface sweep

| Row | Result | Attack and evidence |
| --- | --- | --- |
| B1 private state before every Go consumer | broken | All six current definitions initialize correctly and cache actions disable shared restore/save. Token-preserving shell export in the first Go consumer overrides all three paths; the full five-test suite remains green. Finding shell-cache-override-unguarded. |
| B2 concurrency and push cancellation | held | Fresh TestLinuxCIWorkflowRejectsNarrowingPlants rejects cancel_main and split_same_ref with intended concurrency error; exact expression admits cancellation only for pull_request. |
| B3 fork, hosted platforms, Race and pins | held | Fresh TestLinuxRoutingKeepsForksHosted evaluates 216 event/enablement/fork/platform cases across six definitions, including fork=true/false controls. Exact diff preserves every runs-on, Race definition and SPEC_PIN/PNPM_PIN/GOTOOLCHAIN/GOENV value. |
| B4 pre-fix control | held | Re-executed TestLinuxCIWorkflow on base workflow: exit 1, private Go state must precede every Go consumer. Exact candidate full package exit 0. Control demonstrates structural protection, not historical ABI0 reproduction. |
| B5 actual self-hosted remote gate | held | Fresh GitHub run/jobs API read confirms run 39008448258 completed/success at 3cf597414f1285714c976361ff1a8d6f72de725c, 20 successful jobs, 2 optional skips, 8 successful self-hosted Linux instances with nonempty runner registrations. git diff of that snapshot against candidate is empty. Full remote gate reused, not rerun; API and complete snapshot identity independently verified. |
| B6 public text hygiene | held | Added-line path/IP/known-host/email scan plus full six-file diff inspection found no new personal names, host names, addresses or local paths. Existing untouched runner-name text is outside this delta. Runner evidence stays private with neutral aliases and authoritative job IDs in this artifact. |

## Verdict: changes_requested

```json
{
  "findings": [
    {
      "id": "shell-cache-override-unguarded",
      "row": "B1",
      "invariant": "The shipped workflow check must fail when a self-hosted-routable job runs Go without job-private GOCACHE and GOMODCACHE.",
      "mechanism": "tools/ciworkflow/isolation_test.go:68-77 checks YAML env maps only. Shell assignments in run bodies can override the initialized state without changing any searched-for token; the separate initializer test never executes the consuming workflow body.",
      "reproductions": [
        {
          "test_file": "tools/ciworkflow/isolation_test.go",
          "driver": "BUG-261016-mlpvtn_reproduce.py",
          "command": "From a scratch directory containing snapshot/ extracted from candidate tree: python3 BUG-261016-mlpvtn_reproduce.py",
          "candidate": "f4be4e3a8887893d2e90374ad487588a9d43f293",
          "mutation": "Prepend export GOCACHE=$RUNNER_TEMP/shared-build GOMODCACHE=$RUNNER_TEMP/shared-mod GOPATH=$RUNNER_TEMP/shared-path to the existing first toolchain identity step, preserving its Go command, initializer and all guard tokens.",
          "expected_failure": "TestLinuxCIWorkflow must reject the shared-path consumer. Instead the full package including behavioral tests exits 0. Executing the injected shell statement after the real initializer makes the real Go binary report the three shared paths.",
          "logs": ["BUG-261016-mlpvtn_shell-override-suite.log", "BUG-261016-mlpvtn_shell-override-result.log"]
        }
      ],
      "severity": "bypass",
      "repeat-of": "none"
    }
  ],
  "notes": [
    "The current unmutated workflow correctly initializes all six definitions; the blocking finding is a false acceptance by the newly shipped regression gate, not evidence of current shared caches.",
    "Original historical ABI0 reproduction and arbitrary host-load qualification remain explicitly unestablished; revised brief permits candidate review after green remote gate and defers three main runs.",
    "No numeric review or free-hunt budget was provided. All six supplied review surfaces were swept. No additional blocking mechanism was found in the bounded free hunt.",
    "Initial scratch test invocation raced archive extraction and failed setup; it was rerun after extraction and passed. Initial behavioral witness used a local goenv shim that rewrote GOPATH; the final portable reproduction selects the real Go binary using GOROOT/bin and succeeds. Neither preliminary failure is counted as evidence."
  ]
}
```

Required rework: add a named regression for shell cache reassignment in the consuming workflow step, with a narrowing/token-preserving mutant and a valid control. Enforce the supported workflow shell/environment contract or exercise actual consumer environment propagation so this mutant fails. Do not add a product timeout flag or expand into a separate harness/research task. Preserve current isolation, routing and scheduling behavior.

## Validation and evidence reuse

Reviewer reran the complete tools/ciworkflow package on exact candidate blobs, pre-fix control, actionlint, the token-preserving mutant with the complete package, and actual shell-to-Go environment witness. All mutations live only in ignored scratch fixtures. The candidate source tree remains unchanged. No full product gate was rerun: GitHub run/job reads and exact tree equality validate reuse of the producer's successful full remote run.

The run evidence is https://github.com/relux-works/curator/actions/runs/39008448258 . Exact registration names were read privately; sanitized mapping accompanies this verdict. Probe bounds were checked in internal/godriver/session.go (15 seconds and two minutes, clamped), with no product flag added in the diff. Privacy checks cover added text, not unrelated historical content.
