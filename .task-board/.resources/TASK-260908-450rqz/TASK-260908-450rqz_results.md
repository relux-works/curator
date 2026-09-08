# F1 version rework — TASK-260908-450rqz

Ready for review; independent reviewer acceptance remains mandatory.

This rework supersedes the original version choice. Reviewer RUN-260908-c1ad07 withheld acceptance only for F1: the settled goal requires 0.3.0-draft. Exactly six current occurrences changed in SPEC.md, README.md, main.go and main_test.go. Historical rows and the Pi E1/E2 explanation remain intact. All seven candidate paths were compared against CR1 tree 9ff5c7588e3ef9705be5bfa92d20d321f4619e07; only these six substitutions differ. No new behavior, dependencies, stages or commits.

## Checks run in this rework

| Command | Real exit | Evidence |
|---|---:|---|
| go test ./cmd/curator-run ./internal/mapping -run 'TestSpecVersionPinned|TestRunInformationalFlags|TestRunMapping|TestRunUnknownResolvedMapping|TestRunUnknownFragmentStillRefusesResolution|TestRunResolveFailurePrecedesMapping|TestResolve|TestKnownUnsupportedIsNotUnknown' -count=1 -v | 0 | focused-01.log |
| go run ./cmd/curator-run --version | 0 | version-01.log: curator-run 0.0.0-dev (specification 0.3.0-draft) |
| git diff --check | 0 | diff-check-01.log |
| Python exact CR1 comparison and executable-version assertion | 0 | scope-check-01.log |

Configured make check is delegated to the mandatory publication handoff once; its actual result will be captured in the new CR validation resource. No separate broad suite or mutant rerun was warranted by the six literal replacements. The accepted prior behavioral evidence below is retained, not claimed as rerun. Historical command results describe the original producer run. Behavioral coverage remains 9 of 10 AC rows driven, with the native-tail downstream bound unchanged. Tests remain uncommitted candidate files for parent-owned signed delivery.

Initial discovery included rg against absent optional skill directories (exit 2); this was not a validation gate. No failed read was used as evidence of passing validation. No LOGBOOK/control-root writes per explicit scope.

## Retained original producer evidence (version statement corrected)

# A1 mapping producer evidence

Status: ready for review. Independent Astra-medium acceptance and parent-owned signed PR delivery remain pending. Candidate is uncommitted on the managed Story branch, based on freshly fetched launcher main `84e659e1bda41c0b70fad72e9e29b3c7ad474a7d` (fragment PR5). No prior WIP/CR existed for this task.

## Scope and evidence

`internal/mapping.Resolve` is one closed switch using the existing fragment identifiers, not a second adapter registry. `run` calls it immediately after successful `resolver.Resolve`, before the existing not_implemented stub. Unsupported mappings return exit 1 with env_unsupported; no later statement executes. The existing registry still recognizes opencode (`fragment.HomeVariable`) independently of launch support.

| Environment | System | Provider |
|---|---|---|
| claude_code | claude-code | claude |
| codex_cli | codex | codex |
| pi | pi-native | pi |
| opencode / other resolved ID | env_unsupported | no pair |

SPEC 0.3.0-draft corrects only the mapping erratum and corresponding version metadata. E1/E2 come from accepted A0 `TASK-260908-qblycn_a0-verification-findings.md`: old pi uses agents-infra and replaces home; old frozen runtime registry lacks Pi. Accepted design TASK-260908-ggxfte is status done (rev2 accepted). GitHub read confirms skill-agents-management PR23 MERGED at `a2a6e9f377f62a5872d99ecdfff0d1690e385f2a`; source at that commit identifies pinative system pi-native and frozen pi-anthropic/pi-openai/pi-google runtimes. No module import, pseudo-version, release/tag, real ax, model invocation, install, runtime-home, or LOGBOOK/control-root write.

## Measured behavioral AC coverage

**9 of 10 rows driven through production `cmd/curator-run.run`; one native-tail downstream bound.** Unknown successfully resolved ID uses a resolver-interface test double, explicitly not a claim that today's parser can resolve unknown registry members.

| Row | Named test | Production call site / bound |
|---|---|---|
| claude_code pair | TestRunMapping/claude_code | run → real fragment Resolver.Resolve → mapping.Resolve |
| codex_cli pair | TestRunMapping/codex_cli | same |
| pi pair | TestRunMapping/pi | same |
| opencode refusal and no later stub | TestRunMapping/opencode | same; exact env_unsupported family, no not_implemented |
| unknown resolved ID refusal | TestRunUnknownResolvedMapping | run → resolver-interface success → mapping.Resolve; future resolver boundary only |
| actual unknown fragment | TestRunUnknownFragmentStillRefusesResolution | run → real Resolver.Resolve; resolve_fragment_invalid, not mapping |
| resolution failure before mapping | TestRunResolveFailurePrecedesMapping | run → real Resolver.Resolve with opencode failure, no env_unsupported or later stub |
| info skips resolution | TestRunInformationalFlags | run → cli.Parse; forbidden resolver |
| usage skips resolution | TestRunUsageErrorsExit2 | run → cli.Parse; forbidden resolver |
| verbatim native tail | TestNativeTailVerbatim plus TestRunMapping input preservation | cli.Parse native result is checked byte/element-wise; run preserves input argv, but no downstream launch stage exists to observe delivered native argv |

Existing TestRunParsedLaunchResolvesThenRefuses, TestRunResolveFailuresExit1, TestRunProductionResolverAgainstFakeCurator and executable Unicode-boundary tests retain their original failure families. Later defaults/plan/exec stages do not exist; no-launch evidence is bounded to the current refusal/stub flow. Tests are candidate files for the parent's signed commit; producer commits are forbidden.

## Narrowing mutants

Harness: `python3 .scripts/mapping-mutants.py .temp/TASK-260908-450rqz/mutants`. Each mutation runs `go test ./cmd/curator-run ./internal/mapping -count=1 -v` as a standalone process. It restores exact candidate bytes in finally and requires the named FAIL line, not just any nonzero exit.

| Mutant | What the gate is narrowed to | Named failing test | Real exit | Survival bound |
|---|---|---|---:|---|
| M1 | permits exactly opencode; all other unsupported remain refused | TestRunMapping/opencode | 1 | killed |
| M2 | permits exactly future_env; all other unsupported remain refused | TestRunUnknownResolvedMapping | 1 | killed |
| M3 | mapping error return skips only opencode | TestRunMapping/opencode | 1 | killed |
| M4 | Pi pair reverts to legacy pi system | TestRunMapping/pi | 1 | killed; pair regression, not refusal-gate evidence |

No survivors. M1/M2 prove the new closed mapping gate; M3 proves the production refusal return. No new source-text gate exists, so token-preserving source-gate attack criterion is N/A. Earlier fragment/CLI mutants were not rerun: their code is unchanged, and their behavioral suites ran green.

## Commands and real exits

| Command | Exit | Evidence |
|---|---:|---|
| go test ./internal/mapping ./internal/cli ./internal/fragment ./cmd/curator-run -count=1 (twice; second after docs/version update) | 0 / 0 | focused-01.log, focused-02.log |
| python3 .scripts/mapping-mutants.py .temp/TASK-260908-450rqz/mutants | 0 | mutants/summary.tsv; each expected-red Go gate exited 1 |
| make check | 0 | make-check-01.log; build, formatting, vet, test and race |
| git diff --check | 0 | direct tool result |

All commands run directly without tee. make check ran once by producer before publication. Any additional mandatory board publication validation is board-owned evidence, not inferred from these logs.

Operational recovery: first set_status development exited 1 because estimate was required; set_estimate Fibonacci 3 then set_status exited 0. Initial GitHub repo spelling agents-management returned a repository lookup failure; corrected skill-agents-management query exited 0. Two guessed upstream source paths failed lookup; corrected pinative/pinative.go and vendorplugin/runtime.go reads succeeded. Unknown update_checklist_item schema query exited 1; supported remove/add operations split the mixed producer/reviewer checklist item truthfully, retaining independent acceptance before closure. No failed read was treated as absence.

Producer checklist source-text and LOGBOOK entries are N/A under the explicit no-source-gate/no-control-root scope. Reviewer acceptance is not claimed. Parent must route this candidate to an independent Astra-medium reviewer before closure and signed landing.
