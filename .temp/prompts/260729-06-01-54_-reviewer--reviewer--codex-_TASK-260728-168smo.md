# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260728-168smo, status=reviewing)'
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

- [ ] Decision selects exactly one v1 artifact model: JVM runtime bundle or Kotlin/Native executable, with the alternative explicitly deferred
- [ ] Closed local and repository driver recipes reject generic Gradle tasks, arbitrary commands, plugins, annotation processors and package-selected toolchains
- [ ] Required JDK/Kotlin or Kotlin/Native identities, offline inputs, artifact layout, platform matrix, cache and signing boundaries are specified
- [ ] Code written per task description and AC
- [ ] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [ ] Board size is proportional to the spec and is the smallest decomposition that maps every requirement
- [ ] Every story and task traces to a concrete spec requirement; justified-gap elements also carry a self-verified gap record
- [ ] Beyond-literal-spec elements include a written justification naming the gap and the spec and out-of-scope checks performed before creation
- [ ] Research tasks cite an exact question the spec genuinely leaves open
- [ ] Dependencies linked
- [ ] Tasks are atomic — one clear deliverable each
- [ ] Completeness verified — nothing forgotten
- [ ] Any planning artifacts actually produced are linked as new task-scoped outcome resources; diagrams are strictly optional, never a standing deliverable
- [ ] Implementation matches AC
- [ ] Solution fits project architecture
- [ ] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches
## Your Task

- **ID**: TASK-260728-168smo
- **Title**: TASK-260728-168smo: select-kotlin-artifact-model-and-driver-pair
- **Parent**: STORY-260728-2mnlp0
### Description

Select the Kotlin CLI artifact/runtime model and design closed local and external driver contracts before any implementation.
### Scope

Compare Kotlin/JVM executable JAR plus trusted runtime/bundle options with Kotlin/Native executables; choose driver identifiers; define trusted compiler/JDK or native toolchain requirements; local build_roots versus external skill-build.json targets; source/dependency layout; Gradle, Maven, KSP, compiler plugins, scripts, annotations, native interop, network and launcher policy; cache and platform identity.
### Acceptance Criteria

A reviewed decision selects one implementable artifact model and paired driver identifiers, defines deterministic launch and distribution semantics, exhaustively allows/rejects package-selected build behavior without a generic Gradle escape hatch, integrates toolchain preflight, and fixes macOS/Windows/Linux qualification requirements.

## Instructions

No specific instructions attached. Work according to the task description and acceptance criteria above.

## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# FIRST/LAST lifecycle commands:
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(TASK-260728-168smo, status=analysis)'       # analyst-style work
task-board m 'set_status(TASK-260728-168smo, status=development)'    # implementation / testing work
task-board m 'set_status(TASK-260728-168smo, status=reviewing)'      # reviewer handoff
task-board m 'set_status(TASK-260728-168smo, status=blocked)'        # when blocked
task-board m 'set_status(TASK-260728-168smo, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(TASK-260728-168smo, item=1)'                        # check item N
task-board m 'add_checklist_item(TASK-260728-168smo, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(TASK-260728-168smo, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(TASK-260728-168smo, name=TASK-260728-168smo_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add TASK-260728-168smo ./path/to/file --type outcome --name TASK-260728-168smo_artifact.bin -d "Description"
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
task-board m 'add_resource(TASK-260728-168smo, name=TASK-260728-168smo_results.md, content="...", type=outcome, description="Description")'
task-board resource add TASK-260728-168smo ./path/to/file --type outcome --name TASK-260728-168smo_artifact.bin -d "Description"
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
