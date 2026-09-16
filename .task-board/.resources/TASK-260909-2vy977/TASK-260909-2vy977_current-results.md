# TASK-260909-2vy977 current recovery evidence

## Implementation
Preserved the recovered CR1 and ported rev2 candidate, including the authorized ordered Pi preference. No product source edits in RUN-260915-0ac3b7. Changed only the mutant harness to run its named behavioral test with anchored -run, avoiding repeatedly running unrelated entry tests. Original bytes are restored after each probe. HEAD remains cb232a120c9a04c56ae5037c82347921f020e688; candidate is uncommitted.

## Validation (this recovery)
All direct checks used zsh with set -o pipefail, standalone commands, no tee or pipelines.
- GOWORK=off go test ./internal/defaults ./internal/diagnostics ./cmd/curator-run -count=1: exit 0.
- GOWORK=off python3 .scripts/defaults-mutants.py .temp/evidence/defaults-mutants-current: exit 0; 42 of 42 named probes killed. Every underlying go test exited 1 (expected failing mutant tests, not green tests). Summary and individual logs attached. This rerun is justified by changed harness, interrupted prior batches, revised probes and missing aggregate certification; not reuse of historical 34/34.
- GOWORK=off go build ./...: exit 0 after probes restored source.
- git diff --check: exit 0.
- Fresh git ls-remote exact v0.5.11/tag peeled refs: exit 0, object 0ea486e46765ecf12fff7c2ac526e12da02e95ed and commit a2a6e9f377f62a5872d99ecdfff0d1690e385f2a.
- git verify-tag of that object with preserved task-local historical allowed-signers: exit 0, ECDSA SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM, oparin@me.com. This is historical-key verification, not global host trust enrollment. Earlier default trust verification exit 1 remains recorded.

Manual make check count across this host recovery chain: ZERO. Configured make check is reserved for runtime handoff; no success is claimed in advance. Historical rev2 had TWO full suite executions (manual and runtime), contrary to its old claim; both records retained. Runtime publication and validation are separate from the checks above and must be read from the new runtime log.

## Coverage — fixed denominator from rev2

**11 of 12 AC rows driven at cmd/curator-run.run.** Row 12 is dependency inspection, not a production-entry test. Tests are source in this uncommitted candidate for the managed snapshot; they are NOT claimed already committed. The integration path owns signed commits.

| # | Owned AC row (same denominator as rev2) | Production call site / named test | Status |
|---|---|---|---|
| 1 | Real tagged lineup fallback | run -> Files.Complete -> EmitGroup; TestRunLineupEnvsPrintGroupBeforeRefusal | driven |
| 2 | Levels 1–2 file/flag resolution preserved | run -> Load -> Files.Complete; TestRunDefaultsPerMember/operator-model-machine-effort and flag-model-machine-effort | driven; exhaustive accepted helper matrix remains a bound |
| 3 | Only missing members filled | run -> Files.Complete -> EmitGroup; TestRunDefaultsPerMember (explicit-empty, flag-effort, unknown-not-replaced) | driven |
| 4 | Own row recommendation, not display recommendation | same; TestRunDefaultsPerMember/own-row-recommendation; TestRunLineupEnvsPrintGroupBeforeRefusal | driven |
| 5 | No-effort and unknown-model semantics | same; TestRunDefaultsPerMember/no-effort, pi-google-override, unknown-not-replaced | driven |
| 6 | No completion past failure | run -> Load/Complete; TestRunDefaultsFailuresStopBeforeGroup; TestRunDefaultsPathOrdering/path-failure | driven |
| 7 | Per-member stderr origin group | run -> EmitGroup; TestRunDefaultsPerMember; TestRunPiResolvesNativeLineup | driven |
| 8 | Group ordered before pending plan boundary | run -> Complete -> EmitGroup -> not_implemented; TestRunLineupEnvsPrintGroupBeforeRefusal | driven; not an actual plan invocation |
| 9 | Locks/failures refuse without launch | run -> Load/Complete -> diagnostics.Emit; TestRunDefaultsFailuresStopBeforeGroup; TestRunDiagnosticsContract | driven |
| 10 | Typed failure/no invented model | run -> Complete; TestRunPiRuntimePreference/no-preferred-driven-row and unpreferred-only; TestRunDiagnosticsContract/defaults_unresolvable; TestRunDefaultsPerMember/unknown-not-replaced | driven |
| 11 | Three environments with real supported module | run -> Complete -> EmitGroup; TestRunLineupEnvsPrintGroupBeforeRefusal, TestRunPiRuntimePreference | driven, 3 of 3 environment rows |
| 12 | Verified real tag, no override | go.mod/go.sum, ls-remote, mod download, tag signature verification | bound: inspection/commands, no production call site |

Supplementary API evidence: TestCompletedPairsAdmitTaggedBuildLaunch takes Complete's actual runtime/model/effort into real tagged vendorplugin.BuildLaunch for claude_code, codex_cli and all three Pi vendor bindings. Inert temporary provider executables are only located, not launched. This is API admission evidence, not main wiring or real-provider execution. TestCompleteAmbiguousRuntime and TestCompleteRuntimeResolutionFailure remain helper/API tests and are not silently counted as run coverage.

Stated downstream bounds retained: ErrEffortMissing -> plan_refused with --effort wording, one BuildLaunch/no-retry call, provider-limit enforcement, composition, prompt handling and actual exec/ax belong to the plan/integration tasks. Default-resolution does not silently retry or change the configured pair. NewRegistry error branches are not reachable with the fixed valid production plugin set; registration/API tests do not prove main integration. Cross-platform lanes and real provider execution were not run.


## Boundaries and handoff
The 11/12 ratio denotes candidate test reachability, not already committed code: signed commits belong to integration. Helper-direct registry ambiguity/broken declaration tests and tagged BuildLaunch API tests are bounded evidence, not main execution. All three environment defaults resolve positively before the expressly permitted pending-plan refusal. Pi runtime preference is the authorized pure convention, never a vendor score comparison. Main pipeline, real providers/ax and cross-platform lanes are unverified and outside this leaf.
No new board decomposition, research or planning artifacts were created; those generic checklist clauses are inapplicable. Existing scope/dependency resources are preserved. LOGBOOK writes are explicitly prohibited, so findings are attached here. Historical interruption packets are retained separately; the current host executed the reported commands successfully.
No committed replace/pseudo-version/go.work, control-root source writes, runtime-home changes, tags, releases, installs or restarts. Full configured validation remains unchecked until it actually runs green. This evidence does not attest publication before the runtime performs it.
