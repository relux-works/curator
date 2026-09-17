## Status
development

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
- [ ] provider_directories knob (closed list of absolute paths, default [], lockable) parsed/validated/written; schema cases consumed from the root with root-content skip
- [ ] Resolution implements both revisions behind one option with revision A shipped: A keeps PATH selection and warns subcommand_provider_outside_trust_roots (resolved path, roots consulted, provider_directories hint); B resolves install dir then provider_directories, never PATH, refuses PATH-only/published/managed candidates with subcommand_provider_untrusted; subcommand_provider_missing names roots consulted; subcommand_provider_root_unreadable on a read failure, never absence or fallback; outcomes disjoint
- [ ] env status / curator status report the resolved provider path and trust verdict per discovered curator-<name>, always curator-run and curator-session, missing and unreadable reported; refused/failed provider row non-current
- [ ] Go test executes every umbrella-provider-resolution.json case for both revisions from CURATOR_CONFORMANCE_ROOT (root-content skip + ledger row); hostile PATH-planted curator-run refused under B and warned under A; CHANGELOG E4 warning-release entry; narrow transcripts in TASK-260916-3oh0u8_results.md
- [ ] Code written per task description and AC
- [ ] Relevant tests written for new or changed behavior and passing
- [ ] Lint clean
- [ ] Relevant build/validation commands run after changes and build not broken
- [ ] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant

## Notes
spawn selection rationale tuple: {"role":"developer","pair":"muse-spark-1.3-contributor/max","text":"Wave-1 manager implementation of a landed curator-spec revision (config knobs, resolution/admission logic, posture, vector-execution tests); muse-spark-1.3-contributor:max is the operator's producer pair; reviewer will be codex astra low"}
spawn selection rationale for muse-spark-1.3-contributor/max: Wave-1 manager implementation of a landed curator-spec revision (config knobs, resolution/admission logic, posture, vector-execution tests); muse-spark-1.3-contributor:max is the operator's producer pair; reviewer will be codex astra low
spawn agent resolution: Agent selection: muse via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (muse) (run=RUN-260917-121615, max_parallel=20)
spawn run started: [implementer] developer (muse) (run=RUN-260917-121615)

## Precondition Resources
- [remediation-manager-producer-rules.md](file://TASK-260916-3oh0u8/remediation-manager-producer-rules.md) — Campaign rules for curator manager producers/reviewers: worktree, spec source at curator-spec 0da4020, conformance root and root-content skips, warn-first, posture, validation, handoff
- [TASK-260916-3oh0u8_brief.md](file://TASK-260916-3oh0u8/TASK-260916-3oh0u8_brief.md) — Task brief: spec sections, current code sites, deliverable, rollout default, vectors, out of scope, handoff

## Outcome Resources
- [TASK-260916-3oh0u8_spawn-log_-implementer--developer--muse-_RUN-260917-121615.log](file://TASK-260916-3oh0u8/TASK-260916-3oh0u8_spawn-log_-implementer--developer--muse-_RUN-260917-121615.log) — System spawn log captured by task-board

## Created
2026-09-16T10:50:08Z

## Last Update
2026-09-17T01:12:46Z

## Assigned To
[implementer] developer (muse)
