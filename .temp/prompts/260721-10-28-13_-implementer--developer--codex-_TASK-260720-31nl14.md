# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260720-31nl14, status=development)'
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

## Status Transitions

- **start_status:** `development`
- **end_status:** `to-review` (review handoff, not accepted done)

## Definition of Done

- Code written per task description and AC
- Relevant tests written for new or changed behavior and passing
- Lint clean (`go vet`, `go fmt` / platform equivalent)
- Relevant build/validation commands run after changes and build not broken
- Any produced implementation notes, logs, screenshots, or other deliverables linked as a new outcome resource on the board with a task-scoped name such as `TASK-260218-abc123_results.md`
- Important findings, decisions, anomalies, or regressions recorded in `logbook` when the task uncovers them

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

- [ ] Journal durability precedes each target transition
- [ ] Reverse rollback preserves unknown concurrent state
- [ ] Restart recovery and boundary fault injection pass
- [ ] Code written per task description and AC
- [ ] Relevant tests written for new or changed behavior and passing
- [ ] Lint clean
- [ ] Relevant build/validation commands run after changes and build not broken
- [ ] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
## Your Task

- **ID**: TASK-260720-31nl14
- **Title**: Add a durable install transaction journal
- **Parent**: STORY-260720-3plyvy
### Description

Implement a manager-home transaction engine that stages deterministic target replacements, records preimages and desired digests durably, commits under the home lock, rolls back in reverse order, and recovers interrupted work.
### Scope

Create internal/transaction journal and target abstractions on top of the manager locks. The canonical journal records transaction id, project identity, ordered target class and identifier, expected generation or preimage digest, backup path, desired digest, referenced build keys, and per-target commit state. Require a caller-held home lock for recovery and commit. Use atomic file or directory swaps plus platform durability primitives, retain backups until the consumer target is durable, compare current desired digests before rollback, and refuse to overwrite unknown concurrent state. This task supplies fault-injectable APIs and tests but does not refactor install.Project or install.Global.
### Acceptance Criteria

Target ordering is stable by class then unsigned bytewise identifier; every state transition and backup needed for crash recovery is durable before the next swap; injected failure at each target boundary restores committed targets in exact reverse order and leaves untouched targets unchanged; a desired-digest mismatch during rollback produces implementation-corruption without overwriting current bytes; restart recovery under the home lock resumes commit or rollback deterministically by transaction id regardless of current project; journal-referenced build keys remain discoverable for GC; successful commit removes backups and journal only after durability; subprocess, fault-injection, and race tests pass.

## Instructions

The following instructions have been attached to this task:

### TASK-260720-1zl1cj_candidate-handoff.md
> Reviewer-confirmed managerlock candidate handoff with native Windows runtime gate preserved

# Managerlock candidate handoff

TASK-260720-1zl1cj remains blocked solely because native Windows subprocess execution is unavailable without a Windows runner or an unauthorized commit/push. Review cycle 4 found no remaining code or architecture defect. Consume the candidate product diff from /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1zl1cj/worktree. Preserve the native Windows race/full-suite gate for final qualification; Windows cross-compilation is not runtime evidence. Do not mark TASK-260720-1zl1cj done and do not fabricate CI provenance.




## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# The FIRST/LAST sections above define your role-default lifecycle commands.
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(TASK-260720-31nl14, status=analysis)'       # analyst-style work
task-board m 'set_status(TASK-260720-31nl14, status=development)'    # implementation / testing work
task-board m 'set_status(TASK-260720-31nl14, status=reviewing)'      # reviewer handoff
task-board m 'set_status(TASK-260720-31nl14, status=blocked)'        # when blocked
task-board m 'set_status(TASK-260720-31nl14, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(TASK-260720-31nl14, item=1)'                        # check item N
task-board m 'add_checklist_item(TASK-260720-31nl14, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(TASK-260720-31nl14, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(TASK-260720-31nl14, name=TASK-260720-31nl14_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add TASK-260720-31nl14 ./path/to/file --type outcome --name TASK-260720-31nl14_artifact.bin -d "Description"
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
task-board m 'add_resource(TASK-260720-31nl14, name=TASK-260720-31nl14_results.md, content="...", type=outcome, description="Description")'
task-board resource add TASK-260720-31nl14 ./path/to/file --type outcome --name TASK-260720-31nl14_artifact.bin -d "Description"
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
task-board m 'set_status(TASK-260720-31nl14, status=to-review)'
```

## Working Directory
Board directory: `/Users/iv/Developer/ReluxWorks/curator/.task-board`

Work in the project root. Do not modify board files directly — always use the `task-board` CLI.
