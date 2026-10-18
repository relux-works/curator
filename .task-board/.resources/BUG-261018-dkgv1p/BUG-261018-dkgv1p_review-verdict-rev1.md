# Review B: subcommand-help-flag-dispatches-command

Candidate: CR revision 1, tree `796a42f78bf9fdad6dedd31e571653be096577a8`, signed commit `233c8873889a9f25cd0e2d040a39bcf564a6540c`.

## Coverage checked before implementation inspection

5 of 6 AC rows are driven by named production-entry tests; AC6 is an operational CI/signature check. AC1–3: `TestRegisteredSubcommandHelpNoWrites`, built executable `main -> run -> cli.run`. AC4: `TestCommandHelpRunPassthrough`, `cli.run -> runProvider -> cmdUmbrella`. AC5: `TestCommandHelpNarrowingMutants`, overlays inherited into the executable build and the same public entry. AC6: exact-head hosted CI and signature verification.

No supplied surface table or numerical review/free-hunt budget exists. This review explicitly uses catalog ordering and error-surface rows for argument dispatch and configuration bypass, plus a separate validation identity check. It does not invent a supplied table or claim exhaustive attack coverage.

## Swept surface table

| Surface | Result | Executed attack and evidence |
| --- | --- | --- |
| ordering | held | Both help spellings at command-path positions and after invalid operands, dedicated `help` route, all 72 registry topics/490 invocations; 3 narrowing mutants (init exclusion, leading-only recognition, crossing `--`) killed by named behavioral failures with child exit 1; init positive control and real provider argv roundtrip passed. `BUG-261018-dkgv1p_review-behavior.log`, command below, exit 0. |

Reviewer command: `go test ./cmd/curator -count=1 -timeout=5m -run '^(TestRegisteredSubcommandHelpNoWrites|TestCommandHelp.*)$' -v`. Exit 0; 38.054s. No helper override in this parent invocation. Mutant overlays preserve production files and run the built CLI, not a static checker.

| Additional surface | Result | Executed attack and evidence |
| --- | --- | --- |
| error surface | held | Malformed config, repeated/mixed help flags, concrete profile operand, unknown help topic through the executable: 11 successful dedicated-help calls; unknown topic exit 2 with unchanged state. `python3 .temp/review-dkgv1p/BUG-261018-dkgv1p_review-probe.py`, exit 0; attached probe and log. |

Free hunt: repeated mixed flags at each nested position with a literal `-h` after `--`, and concrete profile name rather than the registry placeholder. No additional findings reproduced. These results assert only the named attacks, not absence of defects.

## Candidate and validation identity

Fresh `git fetch origin main` confirms upstream remains the CR base `1724b4a52b4c2d163c0a3e6ab7c2ad2bbe54ac0c`; HEAD is zero commits behind local main. Every changed working file was compared byte-for-byte with its candidate-tree blob (8 of 8 equal). Candidate commit resolves to the exact CR tree and signature status `G`. `git diff --check` and local executable build exit 0. The registry now supplies actual top-level handlers; the help guard runs before those handlers and config loading. External provider discovery remains the fallback.

Reused full-suite evidence: [CI attempt 4](https://github.com/relux-works/curator/actions/runs/39067488895/attempts/4), exact candidate SHA. Reviewer independently read terminal API results and downloaded the full attempt logs, exit 0: 20 jobs success, 2 optional lanes skipped, lint reports 0 issues. CI Test lanes cover Linux, macOS and Windows; race lanes Linux and macOS. The attached sanitized CI manifest records results and full log digest. Full suite was not rerun locally; narrow behavioral, mutation and extra adversarial probes were rerun by this reviewer. This is reuse of the exact unchanged signed tree and its original hosted environment, not an assertion that local and CI environments are identical.

## Verdict

```json
{
  "verdict": "accepted",
  "revision": 1,
  "findings": [],
  "notes": [
    "No attack-surface table or numeric budgets supplied; the explicitly documented supplemental sweep covers ordering and error surface.",
    "Tree snapshots measure persistent paths, sizes, modes and bytes, not transient writes removed before exit or accesses outside the isolated roots.",
    "Dynamic external provider names and hidden worker modes are outside the registered built-in subcommand contract.",
    "The source-text-gate rule is inapplicable: the guard parses argv; the mutation harness runs behavior."
  ]
}
```

Repository delta is present and appropriate: shared dispatch, dedicated help, registry-driven integration tests, narrowing controls, updated old help expectations, and CLI documentation. No production changes made by reviewer. Acceptance means reviewed, not landed.

## Task-scoped logbook

Review B independently reproduced all three intended mutant failures and found no bypass in mixed help positions with malformed configuration. Verified fresh upstream, signed tree, successful hosted results and readable hosted logs. No direct control-root files edited; this logbook travels as a board outcome.
