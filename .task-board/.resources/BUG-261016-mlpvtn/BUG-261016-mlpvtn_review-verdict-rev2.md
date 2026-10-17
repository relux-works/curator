# Review B revision 2: BUG-261016-mlpvtn — self-hosted-linux-ci-flaky

Verdict: accepted. Candidate tree `5788aaf9481965f0b22ce0257a18d6175b380b01`; base `a17e58156ac30a1629e3ff36efa259bab1555a15`. Repository delta is present. Review ran sequentially without changing candidate code. Fresh origin/main fetch showed no overlap changes on the six candidate paths; local HEAD..main count was zero.

Coverage inspected before technical validation: **0 of 3 original AC rows fully driven; 6 of 6 Linux-routable job definitions checked**. The revised brief explicitly admits candidate review following a green remote gate and defers three green main runs. Historical ABI0 race reproduction and arbitrary loaded-host qualification remain stated bounds, not passing claims. Named production drivers: ci.yml -> isolate-go-state.sh -> TestGoStateInitialization/TestGoStateMissingRunnerContext; ci.yml lint -> go test -count=1 ./tools/ciworkflow/ -> TestLinuxCIWorkflow; TestLinuxCIWorkflowRejectsNarrowingPlants drives scheduling; TestLinuxRoutingKeepsForksHosted drives routing expressions. AC3 requires post-integration main CI, outside this candidate review.

The revision delta from revision 1 changes only the checker/tests and their documented shell subset. The prior `shell-cache-override-unguarded` mechanism no longer reproduces: the identical token-preserving export plant causes the full package to fail with the intended reserved-state error. TestLinuxCIWorkflowShellOverrideBehavior independently executes the actual first Go consumer and demonstrates its unsafe paths, then requires rejection. The 27 literal shell variants pass; weakening the regex to admit GOMODCACHE produces the named export regression failure. No arbitrary-shell interpretation is claimed.

## Swept surface table

| Row | Result | Attack/evidence |
| --- | --- | --- |
| B1 private state | held | Walked go-family-qualification, test, lint, gate-selftest, interop, naming-gate: checkout immediately followed by Linux initializer; no preceding Go consumer. setup-go cache disabled for self-hosted Linux; lint cache disabled/private. Full package, real consumer export, poisoned ambient state, missing context, and narrowed module-name guard executed. Logs: review-tests.log, prior-export.log, narrow-module.log. |
| B2 concurrency | held | cancel_main and split_same_ref narrowing controls refused. Exact workflow/ref group; cancel expression true only for pull_request, false for all push refs. Log: concurrency.log. |
| B3 routing and pins | held | 216 actual-expression cases across six definitions passed, covering forks/trusted PRs, pushes, dispatches and hosted platforms. All runs-on expressions and SPEC_PIN/PNPM_PIN/GOTOOLCHAIN/GOENV lines unchanged from base; Race untouched. Logs: fork-routing.log, preservation.json. |
| B4 control | held | Exact pre-fix workflow fails TestLinuxCIWorkflow with private-state ordering error (exit 1); exact fixed full package passes (exit 0). This proves the workflow contract, not an ABI linker race. Log: base-control.log. |
| B5 remote gate | held | Fresh GitHub run/jobs API reads confirm run 39016112959 completed/success on 5e1d61470ec525b73c216f50c531cf407cffd93e. Entire snapshot tree equals candidate. 20 jobs successful, 2 optional skips, all 8 self-hosted Linux instances successful with nonempty registration names. Full remote gate reused, not rerun. Sanitized per-job mapping: remote-sanitized.json; exact registration names retained privately and in authoritative provider records. |
| B6 public text hygiene | held | Complete six-file diff reviewed, added-line personal-path/IP/email/known-host scan passed. Existing unchanged historical labels are outside this delta. Log: hygiene.json. |

## Findings and notes

```json
{
  "verdict": "accepted",
  "findings": [],
  "notes": [
    "Revision 1 shell-cache-override-unguarded is fixed and its token-preserving witness now fails for the intended reason.",
    "Original AC coverage remains 0 of 3 fully driven with explicit bounds; acceptance applies to the revised workflow-side candidate brief, not final pool qualification.",
    "Three consecutive green self-hosted main runs remain a post-integration obligation; acceptance is not landing or rollout completion.",
    "Computed shell variable names and modifications inside called scripts or third-party actions remain documented checker bounds and require independent review.",
    "No numeric review/free-hunt budget was supplied. All six rows were swept; bounded free hunt revisited missing-context partial writes, poisoned ambient cache paths and current cache consumers without a new blocking finding.",
    "An initial actionlint call without the custom-label configuration failed on the unchanged existing runner label. Repeating with the producer existing allowlist passed; the initial failure is not counted as passing evidence."
  ]
}
```

## Evidence and limits

Reviewer reran exact-tree package tests, the prior export attack with the full behavioral suite, a narrowing guard mutant with the full behavioral suite, concurrency negatives, routing matrix, pre-fix control, configured actionlint, and privacy/preservation checks. All mutations were confined to ignored scratch fixtures or Go overlays. Candidate source and Story branch were not modified. Go 1.25.5 darwin/arm64 and Node v25.6.1 were used locally; local checks do not substitute for Linux runtime qualification.

Full remote execution is evidenced by the producer terminal command record and independently refreshed provider results at https://github.com/relux-works/curator/actions/runs/39016112959 . The complete source-tree equality was verified by git diff --quiet, exit 0. The producer's attached BUG-261016-mlpvtn_shell-override-rework.md supplies the run/job handoff. Product probe limits remain 15 seconds and two minutes, clamped in internal/godriver/session.go; no product flag or timeout adjustment was added.

Evidence bundle: BUG-261016-mlpvtn_review-rev2-evidence.md. This acceptance establishes only that the named attacks did not reproduce; it makes no proof-of-absence claim.
