# TASK-260909-2vy977 — host e11-1 producer evidence

## Implementation

Base/checkpoint for this managed host workspace: cb232a120c9a04c56ae5037c82347921f020e688. Candidate remains uncommitted; the real index and branch are untouched. Historical checkpoint e6827e350485c55c743c4cfcfc80ad6aeb4ed831 is preserved as historical accepted evidence, not recreated on this host.

Applied TASK-260908-25z3wj_change-request_rev1.patch first after git apply --check exited 0. Recovered defaults.go/defaults_test.go without rewriting their accepted behavior. Ported the valid rev2 lineup, native registry, production diagnostics and entry tests onto current main without touching execution, systemprompt, composition or plan source. Existing v0.5.11 go.mod/go.sum pins remain unchanged.

F4 is resolved using the explicit convention in defaults-lineup-brief.md: pi-anthropic, pi-openai, pi-google. Complete selects the first preferred runtime with driven rows, ranks only those rows through tagged vendorplugin.Lineup, and binds its contributor. It never ranks Pi vendor scores together. Configured models still bind independently to their vendor runtime. A missing preference candidate refuses with defaults_unresolvable and the exact preference/reason; declared runtime resolution failures remain failures. SPEC stays 0.3.0-draft and records the preference and changelog. Pi MCP and full pipeline integration remain outside this leaf.

Main now calls Load -> Files.Complete -> EmitGroup after fragment resolution and mapping. Config path discovery is lazy so informational flags and earlier refusals do not depend on HOME discovery. Errors and locks stop before the group. Model/effort origin is independent; configured values are never replaced with another row's defaults. The permitted not_implemented boundary remains immediately after the group. No production BuildLaunch/limits/exec wiring is claimed.

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

## F3 correction and validation accounting

Historical artifacts remain intact. Rev2 performed TWO full checks: manual make check through a tail/echo pipeline, and runtime make check (credible exit 0). Its exactly-once and 12/12 production-entry claims were inaccurate, as recorded by review-verdict-rev2.md. The manual outer exit was echo's status; it was not fail-closed propagation. Historical 34/34 mutation evidence applies to the old candidate, not automatically this port.

This run uses zsh with set -o pipefail for every direct validation below, standalone processes without tee/pipeline. Tool process transcripts carry the observed exits. No manual make check has been run. Runtime validation is reserved for handoff, with its own bounded log as the authority; no success is pre-attested here.

| Direct validation | Observed exit | Explanation |
|---|---:|---|
| GOWORK=off go mod download -json github.com/relux-works/skill-agents-management | 0 | real v0.5.11 module and checksums |
| GOWORK=off go test ./internal/defaults ./cmd/curator-run ./internal/diagnostics -count=1 (initial port) | 0 | focused packages |
| GOWORK=off go test ./cmd/curator-run -run 'TestRunDefaults\|TestRunPiRuntimePreference' -count=1 (first attempt) | 1 | test incorrectly supplied empty CLI flags; existing CLI correctly refused; fixture changed to explicit-empty file values |
| Same command after fixture correction | 0 | production-entry cases |
| GOWORK=off go test ./internal/defaults -run TestCompletedPairsAdmitTaggedBuildLaunch -count=1 (first attempt) | 1 | test omitted explicit PATH; fixed inert binary fixtures |
| GOWORK=off go test ./internal/defaults ./cmd/curator-run ./internal/diagnostics -count=1 (after fixture correction) | 0 | includes successful tagged admission |
| GOWORK=off go test ./cmd/curator-run -run TestRunPiRuntimePreference -count=1 | 0 | added unpreferred-only refusal |
| GOWORK=off go test ./cmd/curator-run -run 'TestRunDefaultsPathOrdering\|TestRunPiRuntimePreference' -count=1 | 0 | lazy path discovery and preference cases |
| File-resolution mutant batch (20 IDs, command in run transcript) | 0 | 20/20 named expected-red tests, each exit 1 |

The file batch tested unchanged defaults/fragment source plus defaults tests. Later main-only path-discovery changes do not change that package's tested inputs. The second batch covers the current lineup/main candidate. Each harness restores original bytes from memory; it never restores from Git. No source-text-inspection admission gate is introduced; the inherited sig attack retains the token and executes behavioral tests.

## Dependency verification

Fresh ls-remote exit 0 returned tag object 0ea486e46765ecf12fff7c2ac526e12da02e95ed and peeled commit a2a6e9f377f62a5872d99ecdfff0d1690e385f2a. A bare task-local fetch obtained that tag. Module download returned h1:i1qwfvOx7hnHqZexW6xE28gsGiN38veDBgyLvER7B+Y=, matching the existing pin.

Default-host git verify-tag returned **1**: cryptographically good ECDSA signature, but No principal matched. This is not reported as green. A task-local allowed-signers file was then pinned to the independently recorded prior rev2 fingerprint SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM (oparin@me.com). The extracted signature key's ssh-keygen fingerprint matched (exit 0), and git -c gpg.ssh.allowedSignersFile=<task-local-file> verify-tag v0.5.11 returned **0**, Good git signature for that principal. This proves signature validity against the historical evidence key, not enrollment in this host's global trust store. No trust configuration, tag or release was changed.

Inspected actual v0.5.11 Lineup/vendor/runtime API docs in the module cache: ranking is vendor-local; native Pi has three frozen vendor-resolved declarations. NewRegistry explicitly registers native Pi and the Google-required gemini/agy plugins. Legacy pi wrapper is not registered. No replace, pseudo-version or go.work is used.

## Scope and lifecycle

No new board elements, planning artifacts or dependencies were created. Existing task scope is SPEC §4.3; accepted file-family recovery is an input resource. Product ambiguity was resolved by the latest authorized brief, not a new inferred policy. Board decomposition/research/justified-gap clauses are inapplicable to this implementation; existing historical analysis remains linked. LOGBOOK/control-root writes are prohibited, so findings live in this task-scoped outcome and board notes.

Host: darwin/amd64, Go 1.26.0 (module requires >=1.25.5). No hosted CI, installs/restarts, real ax, runtime-home changes, commits, tags, releases or trunk integration. Foreign files and human signatures are preserved. Reviewer applies the source-owned attestcheck to supplied declarations as its contract permits; this producer did not clone that checker into the launcher or attest authenticity through it.

## Validation interruption (not a passing result)

The second mutant batch exited 1 on subprocess.TimeoutExpired: its current `go test ./cmd/curator-run -count=1 -v` exceeded 120 seconds. The harness restores source bytes in finally; it did not reach summary publication. Completed per-mutant logs remain under .temp/evidence/defaults-mutants-lineup. No total kill count is claimed for this interrupted batch. After this, even new `/bin/sh` builtin-echo and clock-read processes stopped returning output. Host recovery has been requested. Remaining validation, build and board resource attachment/handoff are not claimed executed.

Filesystem edits remained available while process execution was unavailable. The inherited unresolvable-admit-empty and top-contributor-admit probes were tightened to admit only claude_code empty candidates and exactly two contributors respectively; their earlier broad probes do not count as narrowing evidence. findrow-fuzzy now preserves exact matches and admits a single unknown-id exception. Nil-registry (claude_code only) and cross-driven-model narrowing probes were added. These five revised/new definitions are UNRUN, not certified kills.

On recovery: inspect completed logs and rerun only interrupted/unrun/revised probes. Run the relevant build and whitespace checks; preserve all real failure records. Do not manually run make check: it belongs to runtime handoff. Final configured validation remains unchecked until its exact command exits 0; do not pre-attest it merely to satisfy the handoff checklist.

The task-board resource-add process also produced no output and was explicitly interrupted (exit 130); attachment is UNCONFIRMED. Stalled diagnostic shell processes were interrupted too (exit 130; the PTY probe exited 1 on Ctrl-C). No handoff was attempted and no configured make check was run. Last confirmed board status is development, with configured-validation, mutant-certification and build rows unchecked. The CLI outage prevents recording the supported blocked transition and attaching this packet; do not misread development as active successful validation.

External constraint: process execution must recover before the public board CLI or validation can run. A builtin shell echo and a date read both remained outputless, including /bin/sh with login disabled and /tmp as cwd; shell startup bypasses and waiting did not restore execution. Root cause is unknown. Recommended option: restore host process execution, then resume this exact uncommitted managed candidate and attach/update this task-scoped packet. Rebuilding the Story or editing board records would lose provenance and is not an option. Exact external input needed: confirmation that host process execution is responsive again. No permission to change source scope, runtime defaults, tag policy or publication gates is requested.
