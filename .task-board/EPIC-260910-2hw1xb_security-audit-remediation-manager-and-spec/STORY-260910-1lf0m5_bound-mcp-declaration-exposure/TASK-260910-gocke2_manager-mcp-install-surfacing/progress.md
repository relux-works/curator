## Status
development

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(13))

## Blocked By
- TASK-260910-2ohnjo

## Blocks
- (none)

## Checklist
- [ ] Both S4 profiles behind one option, s4-warn shipped: absent knob unbounded + mcp_env_passthrough_unlisted naming variables and knob with migration hint; s4-enforce absent = empty, unlisted names dropped with mcp_env_passthrough_dropped; explicit null unbounded; reserved names excluded
- [ ] mcp_package_allowlist_empty emitted by profile install, profile update and env status when the allowlist is empty
- [ ] §2.3 surfacing rows printed at install and update after the audit gate and before lock publication, byte-exact closed columns and order, repeated by env status; no rows for an empty MCP set; env status posture shows active profile and effective passable_env_names
- [ ] Go test executes every environments-env-passthrough.json case from CURATOR_CONFORMANCE_ROOT (root-content skip + ledger row); unit tests for warn/drop and install output; CHANGELOG S4 warning-release entry; narrow transcripts in TASK-260910-gocke2_results.md
- [ ] Code written per task description and AC
- [ ] Relevant tests written for new or changed behavior and passing
- [ ] Lint clean
- [ ] Relevant build/validation commands run after changes and build not broken
- [ ] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant

## Notes
spawn selection rationale tuple: {"role":"developer","pair":"muse-spark-1.3-contributor/max","text":"Wave-1 manager implementation of the landed S4 spec (passthrough profiles, surfacing rows, allowlist warning, vector-execution test); muse-spark-1.3-contributor:max is the operator's producer pair; reviewer will be codex astra low"}
spawn selection rationale for muse-spark-1.3-contributor/max: Wave-1 manager implementation of the landed S4 spec (passthrough profiles, surfacing rows, allowlist warning, vector-execution test); muse-spark-1.3-contributor:max is the operator's producer pair; reviewer will be codex astra low
spawn agent resolution: Agent selection: muse via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (muse) (run=RUN-260917-63187b, max_parallel=20)
spawn run started: [implementer] developer (muse) (run=RUN-260917-63187b)

## Precondition Resources
- [remediation-manager-producer-rules.md](file://TASK-260910-gocke2/remediation-manager-producer-rules.md) — Campaign rules for curator manager producers/reviewers
- [TASK-260910-gocke2_brief.md](file://TASK-260910-gocke2/TASK-260910-gocke2_brief.md) — Task brief: S4 passthrough profiles (s4-warn default), allowlist warning, §2.3 surfacing rows, posture, vector-execution test

## Outcome Resources
- [TASK-260910-gocke2_spawn-log_-implementer--developer--muse-_RUN-260917-63187b.log](file://TASK-260910-gocke2/TASK-260910-gocke2_spawn-log_-implementer--developer--muse-_RUN-260917-63187b.log) — System spawn log captured by task-board

## Created
2026-09-10T14:44:02Z

## Last Update
2026-09-17T01:49:15Z

## Assigned To
[implementer] developer (muse)
