# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260823-omp8zt, status=analysis)'
```

## Your Role
# researcher

## Description

Researches questions, collects information, highlights key aspects, performs fact-checking of findings. Writes structured findings documents.

## Deliverable

`.research/{YYMMDD}_{topic}.md` or `artifacts/RESEARCH.md`
Final human-facing wording must say "ready for review" or "handed off to review", not "done", "complete", "finished", "final", or "готово", when the board status is `to-review`.

## Standing Orders

### Evidence Honesty Contract

1. Run each validation or gate command directly as a standalone process. Do not pipe it through `tee`; do not use a pipe chain unless `pipefail` is enabled and the gate command's real status is preserved.
2. Report the real exit code of every validation or gate command.
3. Report expected-red gates truthfully as failing: when a command is expected to fail (for example, `go test` in a package-less module), give its real non-zero exit code and a one-line expected-failure rationale; never present it as passing.
4. Check a checklist item tied to a command only after that exact command has actually run green with exit code 0. If it did not run or did not exit 0, leave the item unchecked.
5. For board reads, use compact task-specific projections. A concrete assignment does not need routine `summary()`, `plan()`, `schema()`, or `{ full }`; request scoped schema only after an unknown call.

## Status Transitions

- **start_status:** `analysis`
- **end_status:** `to-review` (review handoff, not accepted done)

## Constraints

None — full read/write access to research artifacts.

## Skills

These role skill references are a lazy catalog, not a mandate to bulk-read every
body. Before technical work, identify the skills relevant to this task's concrete
scope and read those full skill bodies. Always read any skill explicitly required
by the task, user, or project instructions.

- **project-management**: `/Users/iv/.claude/skills/project-management/SKILL.md`

## Definition of Done

- [ ] Landing-sequencing question answered with the exact CI mechanism cited (pins, candidate suite, SPEC_PIN)
- [ ] cocoaskills change list enumerated with per-item scope estimates
- [ ] curator-spec residual changes beyond in-flight tasks enumerated or explicitly ruled out
- [ ] Findings written to file
- [ ] Key aspects highlighted
- [ ] Fact-checking performed — claims verified, sources cited
- [ ] Findings linked on the board as a new task-scoped outcome resource
- [ ] All questions from task description answered
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
## Your Task

- **ID**: TASK-260823-omp8zt
- **Title**: TASK-260823-omp8zt: analyze-spec-and-cocoaskills-impact
- **Parent**: STORY-260823-2zafr5
### Description

Research, no code changes. Answer three questions with evidence: (1) SEQUENCING — curator-spec CI runs pinned implementations against the conformance suite (Implementations jobs; see Pin landed agent-skill implementations #12, and the candidate-suite mechanism: a new-schema suite enters only through the candidate-conformance workflow_dispatch and is never a default) — determine whether landing schema-8 vectors on main breaks the pinned-implementation jobs, and what the correct landing order is: vector staging, candidate suite, pin bumps, SPEC_PIN in curator CI, release/rc versioning, COMPATIBILITY.md. (2) COCOASKILLS — enumerate the concrete changes it needs for schema 8: manifest parsing of execution_policy on script commands, script-worker-v1 containment obligations, module-roots bijection validation (its go_v1.py:980 currently raises vendor_metadata_inconsistent on any Module.Replace), audit labeling; give per-item scope estimates against the existing stories (cocoaskills board STORY-260822-2evh3p and STORY-260822-27ze8z). (3) CURATOR-SPEC RESIDUALS — anything the in-flight tasks do not cover (registry/audit record surface, profiles, release process). Deliverable: add_resource impact-analysis.md on this task. Maintainer pre-authorization 2026-08-22: fully autonomous, no human approval gates.
### Acceptance Criteria

impact-analysis.md resource on this task answers all three questions with evidence and concrete change lists; the landing task and implementation stories can cite it directly.

## Instructions

No specific instructions attached. Work according to the task description and acceptance criteria above.

## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# FIRST/LAST lifecycle commands:
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(TASK-260823-omp8zt, status=analysis)'       # analyst-style work
task-board m 'set_status(TASK-260823-omp8zt, status=development)'    # implementation / testing work
task-board m 'set_status(TASK-260823-omp8zt, status=reviewing)'      # reviewer handoff
task-board m 'set_status(TASK-260823-omp8zt, status=blocked)'        # when blocked
task-board m 'set_status(TASK-260823-omp8zt, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(TASK-260823-omp8zt, item=1)'                        # check item N
task-board m 'add_checklist_item(TASK-260823-omp8zt, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(TASK-260823-omp8zt, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(TASK-260823-omp8zt, name=TASK-260823-omp8zt_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add TASK-260823-omp8zt ./path/to/file --type outcome --name TASK-260823-omp8zt_artifact.bin -d "Description"
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
task-board m 'add_resource(TASK-260823-omp8zt, name=TASK-260823-omp8zt_results.md, content="...", type=outcome, description="Description")'
task-board resource add TASK-260823-omp8zt ./path/to/file --type outcome --name TASK-260823-omp8zt_artifact.bin -d "Description"
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
task-board handoff TASK-260823-omp8zt --role researcher
```

## Working Directory
Board directory: `/Users/iv/Developer/ReluxWorks/curator/.task-board`

Work in the project root. Do not modify board files directly — always use the `task-board` CLI.
