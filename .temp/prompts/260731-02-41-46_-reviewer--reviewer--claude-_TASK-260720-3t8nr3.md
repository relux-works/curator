# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260720-3t8nr3, status=reviewing)'
```

## Your Role
# reviewer

## Description

Reviews how a task was implemented and how the solution fits into the project. Does not modify code; records one of the explicit verdict branches below.

When the run is goal-bound, query `task-board spawn goal "$TASK_BOARD_RUN_ID"` before recording the verdict. The reviewer goal is role-derived as `reviewer_verdict/reviewer_verdict`, carries its immutable parent goal ID/revision, and is satisfied only by exactly one verdict branch with evidence. A provider exit or `reviewing` status is not a verdict.
The runner persists the branch from the accepted board status plus a new or updated task-scoped verdict artifact. Only persisted `accepted` can satisfy the parent delivery goal; `changes_requested` and `stop_the_line` finish the reviewer goal without accepting delivery.

## Deliverable

Verdict branches are explicit:

- accepted → `done`
- changes requested → `to-dev` for implementation rework or `analysis` for research/decision work, with verdict evidence for the next producer and another reviewer cycle
- genuine stop-the-line boundary → `blocked` only for a concrete external blocker or an unresolved human-only platform/product/architecture/tradeoff/approval decision, with evidence, failed assumptions/attempts, viable alternatives and tradeoffs, a recommendation, and the exact human decision or external input needed

Do not leave the task in `reviewing`, and do not use `blocked` for ordinary rework or a recoverable child/runtime failure.

## Status Transitions

- **start_status:** `reviewing`
- **end_status:** no unconditional default; the reviewer must set exactly one verdict status: `done`, `to-dev`, `analysis`, or evidence-backed `blocked`

## Constraints

Does NOT modify code. Read-only access.
- Reviewer-archetype runs must not supply `commit_ack`; record acceptance evidence for the commit-owning mover, which commits then makes the final `done` transition with `commit_ack=scope_committed`.

## Skills

These role skill references are a lazy catalog, not a mandate to bulk-read every
body. Before technical work, identify the skills relevant to this task's concrete
scope and read those full skill bodies. Always read any skill explicitly required
by the task, user, or project instructions.

- **project-management**: `/Users/iv/.claude/skills/project-management/SKILL.md`
- **architecture-diagrams**: `/Users/iv/.claude/skills/architecture-diagrams/SKILL.md`

## Definition of Done

- [ ] Record the base SHA and create the task worktree only after the clean local main clone is fast-forwarded and dependency handoffs are present.
- [ ] Prove audit-before-build, build-before-mutation, marker-v2 context isolation, consumer-last commit, rollback, and cross-project isolation.
- [ ] Run the focused project and hybrid suites plus python -m mypy and attach task-scoped evidence.
- [ ] Code written per task description and AC
- [ ] Relevant tests written for new or changed behavior and passing
- [ ] Lint clean
- [ ] Relevant build/validation commands run after changes and build not broken
- [ ] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [ ] Rework B3: anchor private materialization staging independently of TMPDIR and prove cross-filesystem auto-mode stability
- [ ] Rework N5: make post-commit GC lock contention non-fatal and cover it
- [ ] Rework N6-N7: record staging placement and canonical consumer-ledger behavior
- [ ] Rework validation: focused, strict-mypy, lint, full suite, and task-scoped evidence
- [ ] Implementation matches AC
- [ ] Solution fits project architecture
- [ ] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches
## Your Task

- **ID**: TASK-260720-3t8nr3
- **Title**: Integrate atomic project and hybrid builds
- **Parent**: STORY-260720-1uv5gi
### Description

Integrate private go-v1 builds and all existing materialization surfaces into one project-scoped plan followed by one manager-home-isolated commit, including hybrid targets and consumer-last ordering.
### Scope

Own the project and hybrid materialization path in src/csk/installer.py plus only the adapter, environment, consumer, and hybrid integration changes required to express transaction targets. Build every miss into operation-private staging outside the home lock after all gates pass. Under the home lock recover, revalidate shared generations and cache trust, publish verified entries, and journal every project, hybrid, runtime, shim, environment, adapter, stale-removal, marker, and consumer target. Reuse the build, cache, marker, shim, and transaction APIs from predecessor tasks.
### Acceptance Criteria

Active builds run provider-first and command-lexically only after audit and before any persistent mutation. A build failure leaves live cache, project, hybrid, runtime, adapters, shims, environments, markers, and consumers unchanged. Commit revalidates closure, ownership, cache winners, target preimages, and generations under the home lock; stale plans restart. New or changed installs write marker v2, exclude build_roots from context and runtime, point shims to immutable cache artifacts, preserve mixed script and system behavior, and update the consumer ledger last. Any publish or target failure rolls back committed targets in reverse order while the home lock remains held; an unreferenced immutable publication may remain only when protocol-safe. Two-project success and success-versus-rollback vectors pass. Focused project, closure, hybrid, adapter, dry-run, rollback, concurrency, and strict-mypy gates pass.

## Instructions

The following instructions have been attached to this task:

### main-windows-concurrency-flake.md
> Main push Windows 3.14 transaction liveness timeout evidence

# Main Windows concurrency failure after planner landing

- Exact main SHA: `b3a5031ed551b27a298eef486a068b5175beaacc`
- Main CI run: https://github.com/ivanopcode/cocoaskills/actions/runs/30556125542
- Failed job: Windows Python 3.14, job `90916913692`
- The identical SHA passed all Windows versions in PR run `30554363746`.
- Windows 3.11, 3.12, and 3.13 passed in the main push run.

The only failure was:

`tests/test_transactions.py::test_concurrent_project_transactions_preserve_both_consumers`

Both worker threads had an empty `errors` list, but one remained alive after the
test's fixed `thread.join(timeout=5)` budget. The assertion at
`tests/test_transactions.py:2170` failed. This is outside the planner rework and
inside the current transactional project/hybrid scope.

Treat a successful one-time failed-job rerun as flaky timing evidence, not proof
that the concurrency test is robust. While implementing the project transaction
integration, preserve the vector and assess whether the test needs deterministic
coordination or a platform-appropriate bounded wait. Do not merely skip/xfail it
or remove its liveness assertion.





## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# FIRST/LAST lifecycle commands:
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(TASK-260720-3t8nr3, status=analysis)'       # analyst-style work
task-board m 'set_status(TASK-260720-3t8nr3, status=development)'    # implementation / testing work
task-board m 'set_status(TASK-260720-3t8nr3, status=reviewing)'      # reviewer handoff
task-board m 'set_status(TASK-260720-3t8nr3, status=blocked)'        # when blocked
task-board m 'set_status(TASK-260720-3t8nr3, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(TASK-260720-3t8nr3, item=1)'                        # check item N
task-board m 'add_checklist_item(TASK-260720-3t8nr3, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(TASK-260720-3t8nr3, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(TASK-260720-3t8nr3, name=TASK-260720-3t8nr3_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add TASK-260720-3t8nr3 ./path/to/file --type outcome --name TASK-260720-3t8nr3_artifact.bin -d "Description"
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
task-board m 'add_resource(TASK-260720-3t8nr3, name=TASK-260720-3t8nr3_results.md, content="...", type=outcome, description="Description")'
task-board resource add TASK-260720-3t8nr3 ./path/to/file --type outcome --name TASK-260720-3t8nr3_artifact.bin -d "Description"
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

## Working Directory
Board directory: `/Users/iv/Developer/ReluxWorks/curator/.task-board`

Work in the project root. Do not modify board files directly — always use the `task-board` CLI.
