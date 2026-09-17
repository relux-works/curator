## Status
development

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(8))

## Blocked By
- TASK-260910-1wjst3

## Blocks
- (none)

## Checklist
- [ ] Approval state package: closed record { path, sha256, approved_by, approved_at } under manager home, canonicalized absolute paths, atomic writes, never read from package/project/profile data
- [ ] Manager records approved_by=manager digests whenever it writes .agents/env.sh or .agents/env.ps1
- [ ] Emitted POSIX and PowerShell hooks verify the candidate digest against the record before sourcing; A-warning (default, sources + warns) and B-enforcing (skips + warns) selectable by one option; shell_hook_env_unapproved / shell_hook_env_changed spelled exactly; one warning per shell session naming path and curator hook approve <path>; forged project-local record ignored
- [ ] Go test executes every shell-hook-trust.json case per the §8.7 recipe (two activations, sourced/diagnostic/warning count) from CURATOR_CONFORMANCE_ROOT with the root-content skip and platform-cases.tsv row; hostile-checkout unit test refuses foreign env.sh under B and warns under A
- [ ] CHANGELOG Unreleased S6 warning-release entry with migration hint; narrow go build/vet/gofmt/test transcripts in TASK-260910-1952mz_results.md
- [ ] Code written per task description and AC
- [ ] Relevant tests written for new or changed behavior and passing
- [ ] Lint clean
- [ ] Relevant build/validation commands run after changes and build not broken
- [ ] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant

## Notes
spawn selection rationale tuple: {"role":"developer","pair":"muse-spark-1.3-contributor/max","text":"Wave-1 manager implementation of the landed S6 spec (hook generation, approval state, vector execution test); muse-spark-1.3-contributor:max is the operator's producer pair; reviewer will be codex astra low"}
spawn selection rationale for muse-spark-1.3-contributor/max: Wave-1 manager implementation of the landed S6 spec (hook generation, approval state, vector execution test); muse-spark-1.3-contributor:max is the operator's producer pair; reviewer will be codex astra low
spawn agent resolution: Agent selection: muse via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (muse) (run=RUN-260917-063e54, max_parallel=20)
spawn run started: [implementer] developer (muse) (run=RUN-260917-063e54)

## Precondition Resources
- [remediation-manager-producer-rules.md](file://TASK-260910-1952mz/remediation-manager-producer-rules.md) — Campaign rules for curator manager producers/reviewers: worktree, spec source at curator-spec 0da4020, conformance root and root-content skips, warn-first, posture, validation, handoff
- [TASK-260910-1952mz_brief.md](file://TASK-260910-1952mz/TASK-260910-1952mz_brief.md) — Task brief: S6 hook trust gate, approval state package, manager-recorded digests, vector-execution test, rollout default A-warning

## Outcome Resources
- [TASK-260910-1952mz_spawn-log_-implementer--developer--muse-_RUN-260917-063e54.log](file://TASK-260910-1952mz/TASK-260910-1952mz_spawn-log_-implementer--developer--muse-_RUN-260917-063e54.log) — System spawn log captured by task-board

## Created
2026-09-10T14:43:12Z

## Last Update
2026-09-17T00:53:23Z

## Assigned To
[implementer] developer (muse)
