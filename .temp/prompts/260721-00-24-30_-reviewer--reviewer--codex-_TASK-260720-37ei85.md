# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260720-37ei85, status=reviewing)'
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

- [ ] Add explicit compatibility assertions for both manifest names at schema versions 1 through 5
- [ ] Guard install marker v1 and conformance claim v1 historical semantics
- [ ] Run validation and two deterministic regeneration passes
- [ ] Guard manager and system config plus registry and audit schemas against build-policy or provenance expansion
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

- **ID**: TASK-260720-37ei85
- **Title**: Add legacy schema compatibility guards
- **Parent**: STORY-260720-35dck7
### Description

Add explicit generator tests and regeneration checks that prevent schema 6 work from changing the structural meaning or generated evidence of schemas 1 through 5, install marker v1, conformance claim v1, fixed manager and system configuration surfaces, or registry and audit evidence boundaries. This closes the preservation requirement before behavioral build vectors are added.
### Scope

Work in curator-spec after all new metadata schemas and cases land. Own focused compatibility assertions in tools/generate-vectors/main_test.go, any non-product test fixture needed to compare frozen legacy semantics, and only the newly added rc.4 compatibility sentence in protocol/core.md needed to accurately describe schema-v1 extension behavior. Do not rewrite legacy schemas or duplicate new build behavior vectors.
### Acceptance Criteria

Tests assert that common command for schemas 1 through 5 still contains only script and system; both schema-v1 manifests preserve deployed additionalProperties extension behavior and assign no build semantics to an incidental build_roots field; each agent-skill and csk-skill schema 2 through 5 rejects build_roots; every schema 1 through 5 rejects type build; install-marker-v1 remains historical shape and claim-v1 remains schema 1 plus protocol rc.3; manager-config-v1 and system-config-v1 gain no driver, argv, environment, toolchain, output-path, hook, or build-policy override surface; registry and audit-record schemas gain no local artifact attestation or receipt-provenance field; regeneration retains legacy schema-case names and expected validity while adding new cases; origin/main comparison distinguishes intentional rc.4 inventory/hash changes from frozen wire semantics; go test ./tools/generate-vectors, make regenerate, make validate, and make regenerate-check pass twice without a diff.

## Instructions

The following instructions have been attached to this task:

### TASK-260720-37ei85_compatibility-decision.md
> Orchestrator resolution derived from the user-approved compatibility requirement

# Compatibility decision

The user-provided epic objective explicitly requires schemas 1 through 5 and existing behavior to remain compatible. Frozen schema v1 intentionally has additionalProperties true. Therefore schema v1 retains that deployed extension behavior: an incidental build_roots field may validate structurally but has no build semantics, while type build remains rejected. Schemas 2 through 5 reject build_roots and type build. Legacy schemas remain byte-frozen. Correct only new rc.4 prose that inaccurately claimed v1 rejects build_roots. This resolves the stop-the-line conflict without a legacy wire change or new reader rule.





## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# The FIRST/LAST sections above define your role-default lifecycle commands.
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(TASK-260720-37ei85, status=analysis)'       # analyst-style work
task-board m 'set_status(TASK-260720-37ei85, status=development)'    # implementation / testing work
task-board m 'set_status(TASK-260720-37ei85, status=reviewing)'      # reviewer handoff
task-board m 'set_status(TASK-260720-37ei85, status=blocked)'        # when blocked
task-board m 'set_status(TASK-260720-37ei85, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(TASK-260720-37ei85, item=1)'                        # check item N
task-board m 'add_checklist_item(TASK-260720-37ei85, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(TASK-260720-37ei85, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(TASK-260720-37ei85, name=TASK-260720-37ei85_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add TASK-260720-37ei85 ./path/to/file --type outcome --name TASK-260720-37ei85_artifact.bin -d "Description"
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
task-board m 'add_resource(TASK-260720-37ei85, name=TASK-260720-37ei85_results.md, content="...", type=outcome, description="Description")'
task-board resource add TASK-260720-37ei85 ./path/to/file --type outcome --name TASK-260720-37ei85_artifact.bin -d "Description"
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
