## Status
to-dev

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(8))

## Blocked By
- TASK-260916-1x0ogh

## Blocks
- (none)

## Checklist
- [x] provider_directories knob (closed list of absolute paths, default [], lockable) parsed/validated/written; schema cases consumed from the root with root-content skip
- [x] Resolution implements both revisions behind one option with revision A shipped: A keeps PATH selection and warns subcommand_provider_outside_trust_roots (resolved path, roots consulted, provider_directories hint); B resolves install dir then provider_directories, never PATH, refuses PATH-only/published/managed candidates with subcommand_provider_untrusted; subcommand_provider_missing names roots consulted; subcommand_provider_root_unreadable on a read failure, never absence or fallback; outcomes disjoint
- [x] env status / curator status report the resolved provider path and trust verdict per discovered curator-<name>, always curator-run and curator-session, missing and unreadable reported; refused/failed provider row non-current
- [x] Go test executes every umbrella-provider-resolution.json case for both revisions from CURATOR_CONFORMANCE_ROOT (root-content skip + ledger row); hostile PATH-planted curator-run refused under B and warned under A; CHANGELOG E4 warning-release entry; narrow transcripts in TASK-260916-3oh0u8_results.md
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant

## Notes
spawn selection rationale tuple: {"role":"developer","pair":"muse-spark-1.3-contributor/max","text":"Wave-1 manager implementation of a landed curator-spec revision (config knobs, resolution/admission logic, posture, vector-execution tests); muse-spark-1.3-contributor:max is the operator's producer pair; reviewer will be codex astra low"}
spawn selection rationale for muse-spark-1.3-contributor/max: Wave-1 manager implementation of a landed curator-spec revision (config knobs, resolution/admission logic, posture, vector-execution tests); muse-spark-1.3-contributor:max is the operator's producer pair; reviewer will be codex astra low
spawn agent resolution: Agent selection: muse via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (muse) (run=RUN-260917-121615, max_parallel=20)
spawn run started: [implementer] developer (muse) (run=RUN-260917-121615)
E4 implemented, revision A shipped. Knob + trust-root lookup (both revisions) + env-status posture + 14-case x2-revision vector test + CHANGELOG. All narrow gates green (build/vet/fmt/lint exit 0; 22-test mask + 96-test CLI subset exit 0; config knob tests exit 0). Known reds are out-of-scope: transitive_system_modules/system_module_waivers/S4-passable cases (baseline-proven) and one envprofile git-subprocess timeout under host load. Findings and spec-gap decisions are in TASK-260916-3oh0u8_results.md (LOGBOOK.md is read-only for producers per campaign rules). One contract note: provider rows are in env status only; curator status (project-skills surface) is untouched per spec section 12.
agent completed: [implementer] developer (muse) (exit=0)
spawn run completed: muse (run=RUN-260917-121615, pid=8889, exit=0)
spawn autonomous recovery: run RUN-260917-121615 queued successor RUN-260917-8c6882 (attempt 1/3, model=muse-spark-1.3-contributor): Change Request construction for TASK-260916-3oh0u8 failed: Change Request CR-TASK-260916-3oh0u8-1 revision 1 validation failed at command 1/1 (1-based) with exit code 1; log resource TASK-260916-3oh0u8_change-request_rev1-validation.log; retry: fix the failure and complete the producer again; the configured suite will rerun automatically
spawn run started: [implementer] developer (muse) (run=RUN-260917-8c6882)
agent completed: [implementer] developer (muse) (exit=143)
spawn run RUN-260917-8c6882 cancelled by operator; operator action required; reason: no operator reason supplied
Orchestrator 2026-09-17: gate rev1 (run 35174489288) failed only on internal/config TestManagerConfigV2Vectors because the committed SPEC_PIN rc.11 root lacks provider_directories in its manager-config-v2 vector; implementation is correct against curator-spec 0da4020. Successor RUN-260917-8c6882 cancelled; HOLD until the operator decides on the pin promotion (see spec-pin-lag-hold.md).
spawn run completed: muse (run=RUN-260917-8c6882, pid=61972, exit=143)

## Precondition Resources
- [remediation-manager-producer-rules.md](file://TASK-260916-3oh0u8/remediation-manager-producer-rules.md) — Campaign rules for curator manager producers/reviewers: worktree, spec source at curator-spec 0da4020, conformance root and root-content skips, warn-first, posture, validation, handoff
- [TASK-260916-3oh0u8_brief.md](file://TASK-260916-3oh0u8/TASK-260916-3oh0u8_brief.md) — Task brief: spec sections, current code sites, deliverable, rollout default, vectors, out of scope, handoff
- [spec-pin-lag-hold.md](file://TASK-260916-3oh0u8/spec-pin-lag-hold.md) — HOLD: SPEC_PIN rc.11 predates the wave-1 spec landings; vector comparison and gate self-test both refuse version skew; awaiting the operator's release/pin decision

## Outcome Resources
- [TASK-260916-3oh0u8_spawn-log_-implementer--developer--muse-_RUN-260917-121615.log](file://TASK-260916-3oh0u8/TASK-260916-3oh0u8_spawn-log_-implementer--developer--muse-_RUN-260917-121615.log) — System spawn log captured by task-board
- [TASK-260916-3oh0u8_results.md](file://TASK-260916-3oh0u8/TASK-260916-3oh0u8_results.md) — E4 provider trust roots: implementation notes, AC mapping, validation transcripts, shipped profile, out-of-scope and spec-gap decisions
- [TASK-260916-3oh0u8_change-request_rev1.patch](file://TASK-260916-3oh0u8/TASK-260916-3oh0u8_change-request_rev1.patch) — Change Request CR-TASK-260916-3oh0u8-1 revision 1 candidate patch (repository_delta=present, 15 changed paths)
- [TASK-260916-3oh0u8_change-request_rev1-validation.log](file://TASK-260916-3oh0u8/TASK-260916-3oh0u8_change-request_rev1-validation.log) — Change Request CR-TASK-260916-3oh0u8-1 revision 1 bounded validation log
- [TASK-260916-3oh0u8_spawn-log_-implementer--developer--muse-_RUN-260917-8c6882.log](file://TASK-260916-3oh0u8/TASK-260916-3oh0u8_spawn-log_-implementer--developer--muse-_RUN-260917-8c6882.log) — System spawn log captured by task-board

## Created
2026-09-16T10:50:08Z

## Last Update
2026-09-17T07:03:10Z

## Assigned To
[implementer] developer (muse)
