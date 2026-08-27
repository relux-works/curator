# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260720-3lo9jc, status=reviewing)'
```

## Your Role
# reviewer

## Description

Reviews how a task was implemented and how the solution fits into the project. Does not modify code; records one of the explicit verdict branches below.

## Deliverable

Verdict branches are explicit:

- accepted → `done`
- changes requested → `to-dev` for implementation rework or `analysis` for research/decision work, with verdict evidence for the next producer and another reviewer cycle
- genuine stop-the-line boundary → `blocked` only for a concrete external blocker or an unresolved human-only platform/product/architecture/tradeoff/approval decision, with evidence, failed assumptions/attempts, viable alternatives and tradeoffs, a recommendation, and the exact human decision or external input needed

Do not leave the task in `reviewing`, and do not use `blocked` for ordinary rework or a recoverable child/runtime failure.

## Status Transitions

- **start_status:** `reviewing`
- **end_status:** no unconditional default; the reviewer must set exactly one verdict status: `done`, `to-dev`, `analysis`, or evidence-backed `blocked`

## Definition of Done

- Implementation matches AC
- Solution fits project architecture
- Tests green
- If review does not accept the work — verdict evidence added via `task-board m 'set_notes(...)'` or an outcome resource, then status routed by the explicit verdict branches above

## Constraints

Does NOT modify code. Read-only access.

## Skills

These role skill references are a lazy catalog, not a mandate to bulk-read every
body. Before technical work, identify the skills relevant to this task's concrete
scope and read those full skill bodies. Always read any skill explicitly required
by the task, user, or project instructions.

- **project-management**: `/Users/iv/.claude/skills/project-management/SKILL.md`
- **architecture-diagrams**: `/Users/iv/.claude/skills/architecture-diagrams/SKILL.md`

## Definition of Done

- [ ] Document one complete schema 6 declaration and the fixed go-v1 prerequisites and limits
- [ ] Document context exclusion, cache and marker currentness, compiler-free dry-run, diagnostics, repair, and future-driver rules
- [ ] Verify local links and run make validate
- [ ] Code written per task description and AC
- [ ] Relevant tests written for new or changed behavior and passing
- [ ] Lint clean
- [ ] Relevant build/validation commands run after changes and build not broken
- [ ] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [ ] Implementation matches AC
- [ ] Solution fits project architecture
- [ ] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches
## Your Task

- **ID**: TASK-260720-3lo9jc
- **Title**: Document schema v6 authoring and CLI behavior
- **Parent**: STORY-260720-35dck7
### Description

Update the schema index, conformance contract, and Curator CLI guide so an author or implementation developer can use schema 6 without inferring behavior from source code. Documentation must explain the complete declaration, toolchain and vendoring prerequisites, excluded build roots, cache and marker semantics, dry-run and status output, security limits, claim transition, and the process for standardizing future drivers.
### Scope

Work in curator-spec after normative prose, schemas, and vector names stabilize. Own schemas/v1/README.md, conformance/README.md, and cli/curator.md. Include one complete schema 6 example and link exact new schemas and vector files. Do not claim that managers already ship support, add generic build flags, or move cross-manager interoperability guidance from the downstream interop story into this protocol task.
### Acceptance Criteria

The schema index lists canonical and legacy v6, receipt v1, marker v2, and claim v2 and states that schemas 1 through 5 and claim v1 remain supported historical contracts; conformance docs define rc.4 manager evidence, v1 versus v2 claim handling, build fixture and vector execution, suite hash meaning, and required failure behavior; CLI docs explain install and upgrade build phase, fixed go-v1 prerequisites, build_roots context exclusion, cache status and currentness, dry-run results without compilation, diagnostics for unsupported or corrupt inputs, repair and GC implications, and no package argv or hook option; a future driver section requires a new identifier, closed schema, threat review, fixed process graph, cache identity, and vectors; links resolve and make validate passes.

## Instructions

No specific instructions attached. Work according to the task description and acceptance criteria above.

## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# The FIRST/LAST sections above define your role-default lifecycle commands.
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(TASK-260720-3lo9jc, status=analysis)'       # analyst-style work
task-board m 'set_status(TASK-260720-3lo9jc, status=development)'    # implementation / testing work
task-board m 'set_status(TASK-260720-3lo9jc, status=reviewing)'      # reviewer handoff
task-board m 'set_status(TASK-260720-3lo9jc, status=blocked)'        # when blocked
task-board m 'set_status(TASK-260720-3lo9jc, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(TASK-260720-3lo9jc, item=1)'                        # check item N
task-board m 'add_checklist_item(TASK-260720-3lo9jc, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(TASK-260720-3lo9jc, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(TASK-260720-3lo9jc, name=TASK-260720-3lo9jc_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add TASK-260720-3lo9jc ./path/to/file --type outcome --name TASK-260720-3lo9jc_artifact.bin -d "Description"
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
task-board m 'add_resource(TASK-260720-3lo9jc, name=TASK-260720-3lo9jc_results.md, content="...", type=outcome, description="Description")'
task-board resource add TASK-260720-3lo9jc ./path/to/file --type outcome --name TASK-260720-3lo9jc_artifact.bin -d "Description"
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
