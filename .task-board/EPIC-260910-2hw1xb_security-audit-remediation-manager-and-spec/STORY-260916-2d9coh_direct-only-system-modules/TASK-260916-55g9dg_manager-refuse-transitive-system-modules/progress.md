## Status
development

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(8))

## Blocked By
- TASK-260916-1hrx51

## Blocks
- (none)

## Checklist
- [ ] Knobs transitive_system_modules (drop|error, default drop; lockable to error only) and system_module_waivers ({package, reason} list, default empty, not lockable) parsed, validated and written; schema cases consumed from the root with root-content skip
- [ ] Admission implemented: direct = root / active overlay / packages named by their requires.contexts; waived packages admitted; drop skips transitive system modules with context_system_module_dropped naming package and module (bytes = admitted modules only); error fails resolution with context_system_module_transitive and leaves the lock unchanged
- [ ] context-system-module-present stays always-warn; launch-fragment works.relux.curator.system-modules follows the admitted set; env status reports the policy value and every dropped module by package and path
- [ ] Five environments.json admission vector cases executed byte-exact from CURATOR_CONFORMANCE_ROOT plus unit tests; CHANGELOG E2 entry; narrow go build/vet/gofmt/test transcripts in TASK-260916-55g9dg_results.md
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
spawn queued: [implementer] developer (muse) (run=RUN-260917-8f2998, max_parallel=20)
spawn run started: [implementer] developer (muse) (run=RUN-260917-8f2998)

## Precondition Resources
- [remediation-manager-producer-rules.md](file://TASK-260916-55g9dg/remediation-manager-producer-rules.md) — Campaign rules for curator manager producers/reviewers: worktree, spec source at curator-spec 0da4020, conformance root and root-content skips, warn-first, posture, validation, handoff
- [TASK-260916-55g9dg_brief.md](file://TASK-260916-55g9dg/TASK-260916-55g9dg_brief.md) — Task brief: spec sections, current code sites, deliverable, rollout default, vectors, out of scope, handoff

## Outcome Resources
- [TASK-260916-55g9dg_spawn-log_-implementer--developer--muse-_RUN-260917-8f2998.log](file://TASK-260916-55g9dg/TASK-260916-55g9dg_spawn-log_-implementer--developer--muse-_RUN-260917-8f2998.log) — System spawn log captured by task-board

## Created
2026-09-16T10:50:06Z

## Last Update
2026-09-17T01:12:49Z

## Assigned To
[implementer] developer (muse)
