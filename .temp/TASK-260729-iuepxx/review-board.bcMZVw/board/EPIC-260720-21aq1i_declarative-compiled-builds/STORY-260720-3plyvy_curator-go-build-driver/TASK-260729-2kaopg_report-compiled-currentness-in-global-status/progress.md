## Status
development

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(5))

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [ ] Define and document the global status --json and --check contract using the existing stable currentness vocabulary
- [ ] Report one diagnostic per active compiled command while preserving declared-skill output compatibility
- [ ] Cover unchanged, source/toolchain/cache/context drift and no-compiled-command global installations
- [ ] Run focused CLI tests and applicable repository gates and attach task-scoped evidence
- [ ] Code written per task description and AC
- [ ] Relevant tests written for new or changed behavior and passing
- [ ] Lint clean
- [ ] Relevant build/validation commands run after changes and build not broken
- [ ] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant

## Notes
EXECUTION DIRECTIVE 2026-07-29: Implement the global-status currentness surface only in an isolated task worktree. TASK-260720-1nlmvv deliberately excludes this branch, so consume its stable vocabulary/classification interfaces without editing that task worktree or duplicating its ordinary status logic. Make an explicit, consistent decision for global status --json and --check, preserve pre-existing declared-skill output when no compiled commands are active, and keep any plan/audit/registry reads strictly read-only. Add focused CLI tests for current and drifted compiled commands plus compatibility cases, update README only for this contract, run applicable Go and formatting gates, attach exact evidence, and hand off to-review. Do not stage, commit, publish, update pins or install host software.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260729-375d85, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260729-375d85)
Implemented on the preserved TASK-260720-1nlmvv tree, mirrored into an own worktree at .temp/TASK-260729-2kaopg/worktree (base origin/main 17804ce, nothing staged or committed; the predecessor tree at .temp/TASK-260720-1nlmvv/worktree was read but never modified).

CONTRACT DECISION (documented in README): curator global status gains --json and --check with the same stable vocabulary and the same fail-closed rule as curator status, and it runs the same read-only plan that curator global install --dry-run runs. That plan is the only way to derive the current logical build input, so reusing cmd/curator/builds.go rather than inventing a second vocabulary requires it. It brings the read-only audit and registry gates into the command; no compiler runs and no installation target, cache entry or trust state is written.

Two deliberate deviations from the project scope, both documented: (1) the machine-wide JSON document carries alias, skills and an optional builds array, and no path, because the scope has no operator-supplied root and the manager home is never published; (2) plain global status keeps its historical always-report / always-exit-zero contract - the declared-skill map is still read straight from install markers, never from the plan, so a plan refusal is a warning: on stderr and only --check turns it into a non-zero exit. --check fails closed twice: for every non-current code, and when the plan refused before it could describe every compiled command.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260729-375d85, pid=19765, exit=0)
CONTINUATION DIRECTIVE after RUN-260729-375d85 incomplete handoff: Preserve .temp/TASK-260729-2kaopg/worktree exactly; implementation and focused gates are reported green, but no full-suite result, checklist completion, task-scoped outcome or developer handoff exists. Resume only after this Opus slot cooldown expires at 05:09:20 Asia/Tbilisi. First inspect the worktree and confirm no prior go test process remains. Run the outstanding TestGlobalStatusReportsATransitivelyResolvedCompiledCommand and the full go test ./... as foreground standalone commands, waiting for real exit codes; do not leave required gates in the background or end while they run. If a full-suite failure is unrelated or load/disk-induced, reproduce the package in isolation and attribute it precisely. Recheck gofmt/build/vet/lint only as needed, attach concise exact evidence, complete checklist truthfully and invoke task-board handoff as the final board command. No new scope, stage, commit, publish or pin.

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260729-2kaopg_spawn-log_-implementer--developer--claude-_RUN-260729-375d85.log](file://TASK-260729-2kaopg/TASK-260729-2kaopg_spawn-log_-implementer--developer--claude-_RUN-260729-375d85.log) — System spawn log captured by task-board

## Created
2026-07-28T23:34:29Z

## Last Update
2026-07-29T00:42:50Z

## Assigned To
[implementer] developer (claude)
