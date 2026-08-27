# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260720-2g21eg, status=development)'
```

## Your Role
# developer

## Description

Writes code — features, bugfixes, refactoring. Writes tests for the code produced.

## Deliverable

Code + tests.
Final human-facing wording must say "ready for review" or "handed off to review", not "done", "complete", "finished", "final", or "готово", when the board status is `to-review`.

## Standing Orders

1. When you change behavior, add or update tests for that scope unless the task explicitly forbids it.
2. Run the relevant test commands yourself before handoff; do not leave test execution implicit.
3. Run the relevant build or validation command after changes to confirm the project still compiles.
4. If a required test or build cannot be run, state exactly what was not run and why.
5. Stop if the implementation starts depending on a forced fit: a platform/API constraint, product decision, UX state model, ownership boundary, or architecture conflict that would require compensating hacks. Document the constraint and options, then ask or mark the task blocked instead of adding more stubs, flags, priority rules, or tests around a broken assumption.

### Evidence Honesty Contract

1. Run each validation or gate command directly as a standalone process. Do not pipe it through `tee`; do not use a pipe chain unless `pipefail` is enabled and the gate command's real status is preserved.
2. Report the real exit code of every validation or gate command.
3. Report expected-red gates truthfully as failing: when a command is expected to fail (for example, `go test` in a package-less module), give its real non-zero exit code and a one-line expected-failure rationale; never present it as passing.
4. Check a checklist item tied to a command only after that exact command has actually run green with exit code 0. If it did not run or did not exit 0, leave the item unchecked.

## Status Transitions

- **start_status:** `development`
- **end_status:** `to-review` (review handoff, not accepted done)

## Constraints

Full read/write access does not authorize forced-fit workarounds. Tests and stubs may verify a valid design, but they must not be used to make an invalid product/API model appear acceptable.

## Skills

These role skill references are a lazy catalog, not a mandate to bulk-read every
body. Before technical work, identify the skills relevant to this task's concrete
scope and read those full skill bodies. Always read any skill explicitly required
by the task, user, or project instructions.

- **project-management**: `/Users/iv/.claude/skills/project-management/SKILL.md`
- **swiftui**: `/Users/iv/.claude/skills/swiftui/SKILL.md`
- **core-data**: `/Users/iv/.claude/skills/core-data/SKILL.md`
- **go-testing-tools**: `/Users/iv/.claude/skills/go-testing-tools/SKILL.md`
- **architecture-diagrams**: `/Users/iv/.claude/skills/architecture-diagrams/SKILL.md`

## Definition of Done

- [ ] Record the base SHA and create the task worktree only after the clean local main clone is fast-forwarded and dependency handoffs are present.
- [ ] Test the exact source-aware argv, graph rejection surface, fixed environment, output verification, and never-run invariant.
- [ ] Run focused pytest, the real fixture build, python -m mypy, and attach task-scoped evidence.
- [ ] Implementation matches AC
- [ ] Solution fits project architecture
- [ ] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches
- [ ] Tests written and passing
- [ ] Coverage target ~80%+ for affected code
- [ ] New task-scoped outcome artifact attached on the board for reports, logs, screenshots, or other produced evidence
- [ ] Code written per task description and AC
- [ ] Relevant tests written for new or changed behavior and passing
- [ ] Lint clean
- [ ] Relevant build/validation commands run after changes and build not broken
- [ ] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
## Your Task

- **ID**: TASK-260720-2g21eg
- **Title**: Implement fixed go-v1 compile driver
- **Parent**: STORY-260720-1uv5gi
### Description

Implement the source-aware go-v1 preflight and compile engine under the rc.5 portable manager-worker-v1 execution policy, with a hidden identity-verified manager-owned worker, manager-owned argv, environment, dependency-graph validation, native-control preflight, capability evidence, and output verification, without cache or install concerns.
### Scope

Own src/csk/builds/go_v1.py, its worker re-execution and injected process executor boundary, the hidden worker entry point, the worker session protocol, and focused unit and fixture tests. Starting from the frozen snapshot and trusted toolchain descriptor, run the fixed four-node process graph of protocol core.md 4.2.1: manager parent, identity-verified manager-owned worker, fingerprinted GOROOT/bin/go, fingerprinted regular executables below GOROOT/pkg/tool/GOHOSTOS_GOHOSTARCH. The worker is an exact re-execution of the installed manager executable in one fixed hidden mode; no package file, manifest value, descriptor value, environment value, PATH lookup, shell, or user option may select it. Split the five fixed argv forms by process: the three package-independent probe forms telemetry-off, version and env stay on the manager parent and are owned by TASK-260720-3j8pp5, while the two source-aware forms list and build are issued only by the worker inside the manager-built clean environment. Drive the 13 session states from vectors/go-host-execution-policy.json session_states, from parent-package-independent-toolchain-probe through worker-domain-teardown. Parse the complete go list JSON stream in the parent, enforce module and path containment for main and transitive dependencies before issuing the authenticated build permit, and validate the staged executable. Emit exactly one capability-evidence-v1 record per operation as a result-only value. Implement macOS and Windows and fail closed on every other host; do not add a Linux control path. Do not publish caches, write markers or shims, launch the output, or claim any deferred hardened guarantee.
### Acceptance Criteria

The only source-aware argv are the normative vendor-only go list and go build forms with trimpath, buildvcs disabled, compiler gc, pgo off, internal linking, libgcc none, and a manager-derived output path, and both are issued only by the manager-owned worker, exactly one of each per session, never by the manager parent; the three package-independent probe forms remain parent-side and belong to TASK-260720-3j8pp5. One worker session performs exactly one go list, waits for parent validation of the complete package graph, accepts exactly one authenticated build permit bound to a fresh session nonce, and performs exactly one go build; any retry, second list, second build, replayed nonce, out-of-order, oversize, or unknown message tears the session down without starting a compiler. The environment starts empty and fixes private Go caches, GOPROXY and VCS off, GOWORK off, GOTOOLCHAIN local, CGO disabled, GO_EXTLINK_ENABLED 0, native target and tuning, locale, temp roots, and empty executable PATH, matching vectors/build-drivers.json fixed_environment. All 18 mandatory_controls from vectors/go-host-execution-policy.json are enforced always, including identity-verified-manager-owned-worker, pre-launch-worker-identity-verification, post-exec-identity-reverification, frozen-source-snapshot-integrity, fixed-manager-selected-process-graph, closed-standard-input-and-descriptors, worker-domain-teardown, no-artifact-execution, inventory-native-controls-applied and closed-capability-evidence-record. Native-control availability is probed once per operation before worker launch against rc5-native-control-inventory-v1, which is exhaustive over exactly macos and windows with exactly the five controls descendant-domain-termination, active-process-count-limit, aggregate-memory-limit, per-file-size-limit and inherited-handle-restriction; every control the inventory marks available for the host is applied, nothing outside the inventory is applied or reported, and a cached, inherited, configured, or host-label result is not a probe. Exactly one capability-evidence-v1 record per operation carries exactly record_version, execution_policy, platform and controls, with exactly one entry per inventory control and exactly the fields name, availability, status and probed_at; it is exposed in dry-run-plan-result, install-result and status-result and excluded from cache-key, receipt, install-marker and conformance-claim. All eight capability_evidence_record consistency_rules are enforced with their declared errors: six reject with build_execution_capability_evidence_invalid and the two hardened rules reject with build_execution_hardened_claim_forbidden. Failure boundary matches the vector exactly: a missing mandatory portable control fails with build_execution_control_unavailable before worker-launch, rejects the build and publishes nothing, while an unavailable inventory native control and a missing deferred hardened capability neither reject the build nor block publication. All six deferred_hardened_guarantees are refused by their six deferred_capability_rejection_guards and are never claimed. All 14 identity_and_protocol_cases and all 8 package_influence_cases from the vector are covered as named negatives, and the 11 capability_evidence_cases are covered as named cases. Exactly one non-DepOnly main package is accepted. Missing or inconsistent vendor data, workspace or toolchain switching, graph paths outside build root or GOROOT, load errors, cgo or native fields, any SysoFiles, nonstandard SFiles, escaped embeds, and active nonstandard cgo_import_dynamic directives fail before the build permit. Output is one bounded regular executable in staging, is hashed and permissioned, and is never run. On any host outside macOS and Windows the source-aware path fails closed with an unavailable-control error and never starts a worker. Focused pytest, a valid real Go fixture build on a native macOS or Windows host, poisoned-environment negatives, and strict mypy pass.

## Instructions

No specific instructions attached. Work according to the task description and acceptance criteria above.

## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# FIRST/LAST lifecycle commands:
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(TASK-260720-2g21eg, status=analysis)'       # analyst-style work
task-board m 'set_status(TASK-260720-2g21eg, status=development)'    # implementation / testing work
task-board m 'set_status(TASK-260720-2g21eg, status=reviewing)'      # reviewer handoff
task-board m 'set_status(TASK-260720-2g21eg, status=blocked)'        # when blocked
task-board m 'set_status(TASK-260720-2g21eg, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(TASK-260720-2g21eg, item=1)'                        # check item N
task-board m 'add_checklist_item(TASK-260720-2g21eg, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(TASK-260720-2g21eg, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(TASK-260720-2g21eg, name=TASK-260720-2g21eg_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add TASK-260720-2g21eg ./path/to/file --type outcome --name TASK-260720-2g21eg_artifact.bin -d "Description"
```

## Spawn Run Control

Tracked background spawn runs expose `TASK_BOARD_RUN_ID` in the child environment.
If your work is long-running, check for operator directives at safe checkpoints:

```bash
task-board spawn status "$TASK_BOARD_RUN_ID"
task-board spawn directives "$TASK_BOARD_RUN_ID"
```

Current runtimes do not support direct inbound push into your active session.
Treat directives as cooperative checkpoint signals:
- persist your current notes/artifacts before acting on `cancel`-style requests
- only honor pause/reroute intent at a safe checkpoint
- if no directive is present, continue normally

## IMPORTANT: Saving Results

When you produce work products (research documents, design docs, screenshots, logs, archives, implementation notes), you MUST save them as outcome resources with names that include the task ID:

```bash
task-board m 'add_resource(TASK-260720-2g21eg, name=TASK-260720-2g21eg_results.md, content="...", type=outcome, description="Description")'
task-board resource add TASK-260720-2g21eg ./path/to/file --type outcome --name TASK-260720-2g21eg_artifact.bin -d "Description"
```

If you revise the same artifact later, use `task-board m 'update_resource(...)'` or `task-board resource update ...` instead of creating a silent overwrite.

If you discover important findings, decisions, anomalies, regressions, or non-obvious constraints while working, record them in `logbook` as well as on the board.

This ensures your results persist on the board and are accessible to other agents and the coordinator. Spawn completion is expected to produce at least one new task-scoped outcome artifact before the task can cleanly remain in `to-review`.

## Stop-The-Line: No Forced Fits

Do not keep implementing when autonomous work starts requiring a forced fit. A forced fit is any path where the task conflicts with a platform/API constraint, product decision, UX state model, ownership boundary, or architecture, and the remaining "solution" is mostly compensating hacks.

Warning signs:
- each fix needs another flag, stub, priority rule, mock-only behavior, or special-case test
- the tests can pass only because the test harness avoids the real platform behavior
- the implementation depends on an assumption you can no longer defend
- the user-facing behavior cannot be described cleanly without contradicting the product model

When this happens, stop product-code changes before adding another workaround layer. Attach or note:
- the constraint and evidence
- the failed assumptions/attempts
- the viable options and tradeoffs
- the recommended option
- the exact human/product/architecture decision needed

Then set the board item to `blocked` and ask only for that exact decision or external input. This stop applies only to a concrete external blocker or an unresolved human-only platform/product/architecture/tradeoff/approval decision; recoverable failures and ordinary rework stay autonomous. Tests and stubs are not proof that a forced-fit design is correct; use them only after the state model and platform assumptions are valid.

## Completion Discipline

Keep working until the task reaches a terminal handoff for your role. If no objective blocker remains, do not stop while the board item is still parked in `analysis`, `development`, `testing`, or `reviewing`.

Before your final status change:
- satisfy the task acceptance criteria and relevant checklist items
- attach outcome evidence for the work you produced
- run the relevant verification commands when the task changes code, tests, docs, or config

Use `blocked` only for either a concrete external blocker you cannot resolve autonomously or an unresolved human-only platform/product/architecture/tradeoff/approval decision. Record the constraint, evidence, failed assumptions/attempts, viable alternatives and tradeoffs, recommendation, and exact human decision or external input needed. Recoverable failures and ordinary rework are not `blocked`.

Status language is literal:
- `to-review` means your role has handed work to review; it does not mean the board task is accepted or done.
- In your final response, say "ready for review" or "handed off to review" when the final board status is `to-review`.
- Do not say "done", "complete", "finished", "final", or "готово" as the overall task state unless the board status is actually `done`.

## LAST — Run For Role Handoff

When you have completed all role work and the task is ready for its role handoff, run this as your **final board command**:

```bash
task-board handoff TASK-260720-2g21eg --role developer
```

## Working Directory
Board directory: `/Users/iv/Developer/ReluxWorks/curator/.task-board`

Work in the project root. Do not modify board files directly — always use the `task-board` CLI.
