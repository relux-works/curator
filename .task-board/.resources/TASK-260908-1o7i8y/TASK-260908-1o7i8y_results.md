# TASK-260908-1o7i8y — producer evidence

2026-09-16; recovery run RUN-260915-a73866. Candidate based on
26baf9777e5a406aeeca343ab5a3b0a25925165e; uncommitted Story worktree.
This replaces the earlier upstream-blocker report. v0.5.13 resolves that
blocker with BuildLaunchWithEnvironment and the admitted OwnedEnv snapshot.

## Delivered candidate

The existing resumed changes wire production run through ax configuration,
fragment resolution, mapping, defaults/Lineup, tagged interactive admission,
separate provider-limit admission, prompt selection, snapshot composition,
late checks and execution. Composer consumes OwnedEnv without rebuilding a
request. SPEC changelog records the requested erratum. The interim
not_implemented path and obsolete RemainingObligations inventory are removed.
Diagnostics render at production call sites; foreign bytes remain transport.

## Production-entry coverage

All named tests below enter cmd/curator-run.run, not only helpers.

| Obligation | Test evidence | Bound |
| --- | --- | --- |
| axconfig.Load before cli.Parse | TestProductionModeSelection; TestProductionAxProfileAndConfigDiscovery | absent/disabled/enabled, machine precedence, malformed/dangling/directory, config errors before invalid argv/help |
| fragment repair invocation, mapping, defaults/Lineup and origins | TestProductionPipelineGoldens; TestProductionDefaultBindingsAndLineup; existing defaults/main tests | real fragment parser; resolver subprocess boundary scripted; repair argv asserted, real repair not claimed |
| tagged interactive BuildLaunch and provider limits | TestProductionPipelineGoldens; TestProductionAdmissionRefusals; TestProductionDefaultBindingsAndLineup | real v0.5.13 admission and temporary store; negative model/effort/unknown/limited/read-error shapes |
| prompt selection, PrepareLaunch and warnings | TestProductionPipelineGoldens; TestProductionPromptWarningsWithoutSection; TestProductionLateChecks | selected channels and Pi discovery in both modes; Pi has no MCP |
| composition from owned snapshot | TestProductionPipelineGoldens; TestProductionExitAndStdin | 6/6 adapter/mode argv/env/transport goldens; attached payload injected after real admission because interactive plugins attach none |
| three actual late checks in both modes | TestProductionLateChecks | 20/20 refusal shapes (10 per mode), mutation occurs after composition; no child side effect |
| direct status, fake-ax document, diagnostic bytes | TestProductionExitAndStdin; TestProductionAdmissionRefusals; TestProductionPipelineGoldens | direct 37 and SIGTERM=143; ax failure=1 per SPEC; binary foreign bytes retained |
| no interim path; forbidden flags | TestProductionPipelineGoldens; TestProductionForbiddenFlags | real fake-process side effects; 18/18 forbidden flag/mode cases, 9 negative goldens |

Driven implementation rows: **8/8** above. This is not a count of every SPEC
clause. Installed integration is **0/3 real adapters** in this producer run,
reserved to the orchestrator by a2-main-wiring-brief.md.

## Validation

Standalone processes in zsh with `set -o pipefail`; no tee or masking pipeline.
The restored candidate is frozen for the post-mutant narrow suite and build.
The first narrow suite also passed (exit 0), but mutation work began while its
already-built test processes were running, so the post-restoration replay is
the authoritative candidate validation rather than relying on that overlap.

| Command | Actual exit |
| --- | --- |
| go test ./cmd/curator-run ./internal/composition ./internal/plan ./internal/execution ./internal/diagnostics -count=1 (post-restoration) | 0 |
| go build ./... | 0 |
| make fmt-check | 0 |
| git diff --check | 0 |

The attached resume-tests.log contains the five passing package results.

## Mutants

Fresh production harness: python3 .scripts/pipeline-mutants.py
.temp/TASK-260908-1o7i8y/pipeline-mutants-resume — exit 0.
**9/9 narrowing mutants killed, 0 survivors**. Each underlying named go test
returned **1 (expected failure)**: E1/E2/E3 narrow prompt/binary/MCP refusal to
direct mode; E4/E5 admit directories; E6 admits regular nonexecutable binaries;
M1 ignores enabled ax for Pi; M2 ignores machine false; M3 ignores malformed
configuration with argv. Attached logs preserve exact commands and failures.
Harness restores source bytes in finally and verifies restoration.
Bound: these nine probes are not exhaustive filesystem race or mode proofs.

Earlier continuation scratch logs report composition 9/9 and plan 13/13 mutants
killed. They were inspected but NOT rerun or counted as fresh acceptance here;
the fresh evidence is the nine production-entry mutations and narrow suite.

## Integration and operational boundaries

No install, real provider/ax launch, hosted CI, signed commit/PR, independent
reviewer acceptance or cross-platform validation was performed. The runtime
owns the one landing suite at handoff; it was not run manually. Orchestrator
owns installation, real umbrella repair/current status, model/effort and MCP
verification for Claude/Codex, native Pi without MCP, review and signed delivery.
No managed Story commit was created.

## Logbook finding

The v0.5.13 snapshot removes the previously documented ownership conflict;
reconstructing the private effective request is unnecessary. The interrupted
previous run had left code but no published Change Request. This recovery
revalidates it and attaches evidence before handoff. Campaign rules prohibit
LOGBOOK.md edits and no logbook CLI is installed: this finding is recorded in
this outcome and task notes for the orchestrator to carry into the logbook.

## Current handoff continuation — RUN-260915-6ac9e7

The orchestrator rescoped the producer checklist and resumed this preserved
candidate. No product, test, or documentation source changed in this run.
The coverage table above describes the candidate; prior mutant logs are retained
as earlier-run evidence, not claimed as freshly executed here.

Fresh standalone zsh commands (`set -o pipefail` for Go/make):
- `go test ./cmd/curator-run ./internal/composition ./internal/plan ./internal/execution ./internal/diagnostics -count=1`: exit **1**. Four packages passed; execution failed one direct PTY trial awaiting terminal INT after READY (three earlier trials passed).
- `go test ./internal/execution -count=1 -run '^TestRealPTYOwnershipBothModes$' -v`: exit **0**, 10/10 PTY trials passed. This does not explain or erase the earlier intermittent failure.
- `go build ./...`: exit **0**.
- `make fmt-check`: exit **0**.
- `git diff --check`: exit **0**.

The PTY intermittency is an unresolved validation limitation for the reviewer;
no speculative timing or signal-handling workaround was added. The handoff
runtime owns the single landing-suite execution. Installation and real adapter
launches remain orchestrator post-landing obligations on the Story; independent
acceptance remains reviewer-owned. The generic logbook item is satisfied via
board notes and outcomes under the explicit campaign override; LOGBOOK.md is
untouched, and no claim of writing that file is made.

- Diagnostic rerun `go test ./internal/execution -count=1`: exit **0** (execution-recheck.log).
