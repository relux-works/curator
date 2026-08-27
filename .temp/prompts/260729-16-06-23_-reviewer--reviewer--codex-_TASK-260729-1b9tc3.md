# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260729-1b9tc3, status=reviewing)'
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

For an enforced Bug/Story `done` transition, a reviewer-archetype run must not supply `commit_ack`. Record acceptance evidence and hand it to the commit-owning mover; after committing its scope, that mover makes the final `done` transition with `commit_ack=scope_committed`.

For board reads, use compact task-specific projections. A concrete review does not need routine `summary()`, `plan()`, `schema()`, or `{ full }`; request scoped schema only after an unknown call.

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

- [ ] Immutable rc.5 root and all relevant vector clusters are inventoried with digests
- [ ] Python fixture loader and independent consumer boundaries are designed without copied Go logic
- [ ] Positive, rejection, identity, lifecycle, platform, and non-alias cases map to exact tests
- [ ] Literal first-wave pytest implementation gates are attached in a task-scoped outcome
- [ ] Board size is proportional to the spec and is the smallest decomposition that maps every requirement
- [ ] Every story and task traces to a concrete spec requirement; justified-gap elements also carry a self-verified gap record
- [ ] Beyond-literal-spec elements include a written justification naming the gap and the spec and out-of-scope checks performed before creation
- [ ] Research tasks cite an exact question the spec genuinely leaves open
- [ ] Dependencies linked
- [ ] Tasks are atomic — one clear deliverable each
- [ ] Completeness verified — nothing forgotten
- [ ] Any planning artifacts actually produced are linked as new task-scoped outcome resources; diagrams are strictly optional, never a standing deliverable
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [ ] Implementation matches AC
- [ ] Solution fits project architecture
- [ ] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches
## Your Task

- **ID**: TASK-260729-1b9tc3
- **Title**: TASK-260729-1b9tc3: design-csk-rc5-conformance-consumer
- **Parent**: STORY-260720-1uv5gi
### Description

Design the exact independent Python conformance consumer and fixture strategy for CocoaSkills against the immutable rc.5 build-driver, lifecycle, execution-policy, schema, receipt, marker, and claim artifacts.
### Scope

Read-only design and executable test blueprint. No CocoaSkills product/test edits, no copied Go logic, no pin or publication changes, no broad test execution.
### Acceptance Criteria

Outcome maps every relevant rc.5 vector cluster to Python test modules and product boundaries, defines fixture loading and digest provenance, negative/non-alias coverage, platform gating and skip policy, and literal first-wave pytest commands for implementation.

## Instructions

The following instructions have been attached to this task:

### TASK-260729-1b9tc3_consumer-scope.md
> Immutable rc.5 root and CocoaSkills consumer design scope

# rc.5 Python consumer design

Authoritative immutable root: /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1. Record complete root and key vector digests before analysis. Read accepted parity map TASK-260729-1t1z2l and current CocoaSkills repository at /Users/iv/Developer/Wildberries/cocoaskills. Design an independent Python loader/assertion architecture for build-drivers.json, manager-lifecycle.json, go-host-execution-policy.json, schema cases, expected build-driver bytes, receipt/marker/claim identities, error clusters, cache outcomes, and manager-worker launch semantics. Map cases to exact existing or proposed pytest files and product modules. Preserve explicit CURATOR_CONFORMANCE_ROOT candidate input and previous default pin. Linux driver execution remains deferred; macOS and Windows are current. Do not edit either repository or copy Go implementation logic.


### TASK-260729-1b9tc3_review-instructions.md
> Independent review scope for CocoaSkills rc.5 conformance consumer design

Independently review the CocoaSkills rc.5 conformance consumer design. Read TASK-260729-1b9tc3_rc5-conformance-consumer-design.md and both baseline/candidate logs. Verify immutable rc.5 root and digests, Python loader boundaries, that the design consumes authoritative JSON/goldens rather than copying Curator Go logic, complete mapping of positive/rejection/identity/lifecycle/platform/non-alias cases, literal first-wave pytest gates, current CocoaSkills file/module fit, and truthful rc.2-to-rc.5 drift. Do not edit repositories, design, briefs, pins, or run broad tests. Attach a task-scoped review verdict and route accepted to done only if all claims and checklist items are supported; otherwise list exact corrections.




## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# FIRST/LAST lifecycle commands:
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(TASK-260729-1b9tc3, status=analysis)'       # analyst-style work
task-board m 'set_status(TASK-260729-1b9tc3, status=development)'    # implementation / testing work
task-board m 'set_status(TASK-260729-1b9tc3, status=reviewing)'      # reviewer handoff
task-board m 'set_status(TASK-260729-1b9tc3, status=blocked)'        # when blocked
task-board m 'set_status(TASK-260729-1b9tc3, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(TASK-260729-1b9tc3, item=1)'                        # check item N
task-board m 'add_checklist_item(TASK-260729-1b9tc3, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(TASK-260729-1b9tc3, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(TASK-260729-1b9tc3, name=TASK-260729-1b9tc3_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add TASK-260729-1b9tc3 ./path/to/file --type outcome --name TASK-260729-1b9tc3_artifact.bin -d "Description"
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
task-board m 'add_resource(TASK-260729-1b9tc3, name=TASK-260729-1b9tc3_results.md, content="...", type=outcome, description="Description")'
task-board resource add TASK-260729-1b9tc3 ./path/to/file --type outcome --name TASK-260729-1b9tc3_artifact.bin -d "Description"
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
