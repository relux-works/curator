# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260729-osjeay, status=analysis)'
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

- [ ] Current CI/Make/toolchain/dependency drift is source-verified
- [ ] Exact file-level producer plan and candidate input/pin invariants are attached
- [ ] macOS, Windows, Linux, race, vet, format, and lint matrix is executable
- [ ] Task-scoped outcome is independently reviewable and no product/pin edits occurred
- [ ] Implementation matches AC
- [ ] Solution fits project architecture
- [ ] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches
- [ ] Independent reviewer verifies the execution map source facts and executable command contracts under the read-only no-Go scope
- [ ] Execution-map command contracts validated by an executable no-Go stub harness (verify-recipes.sh: 7/7 expectations met, real exit 0)
- [ ] Cycle-4 rework validated by the extended no-Go stub harness (verify-recipes.sh: 21/21 expectations met, real exit 0) — supersedes the 7/7 wording in item 16, which the append-only CLI cannot reword
- [ ] Cycle-5 rework validated by the extended no-Go/no-Windows stub harness (verify-recipes.sh: 41/41 expectations met, real exit 0) - supersedes the 21/21 wording in item 17
- [ ] Findings written to file
- [ ] Key aspects highlighted
- [ ] Fact-checking performed — claims verified, sources cited
- [ ] Findings linked on the board as a new task-scoped outcome resource
- [ ] All questions from task description answered
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
## Your Task

- **ID**: TASK-260729-osjeay
- **Title**: TASK-260729-osjeay: prepare-final-curator-ci-execution-map
- **Parent**: STORY-260720-3plyvy
### Description

Read-only refresh of the final Curator compiled-build CI task against the accepted rc.5 candidate, current workflow, Makefile, platform runners, and current task dependencies so implementation can start without another discovery phase.
### Scope

Inspect TASK-260720-1pvfj5, current Curator candidate and accepted comparison, .github/workflows/ci.yml, Makefile, Go/tool versions, macOS/Windows/Linux runner constraints, and existing task evidence. Produce an exact file-level producer plan, conflict-free edit ownership, candidate-input/pin invariants, platform matrix, and narrow/full validation commands. Do not edit product/spec files, CI, pins, task 1pvfj5, or run heavy tests.
### Acceptance Criteria

Outcome identifies all stale rc.4 wording and dependency drift, exact files and YAML/Make targets, immutable rc.5 candidate evidence inputs, Linux/macOS/Windows and race/vet/gofmt/lint gates, native prerequisite handling, and the smallest producer/reviewer sequence. Every claim is source-verified and no product or pin mutation occurs.

## Instructions

The following instructions have been attached to this task:

### TASK-260729-osjeay_audit-scope.md
> Bounded read-only final Curator CI execution-map scope

Read-only CI readiness audit. Authoritative task TASK-260720-1pvfj5 currently contains stale rc.4 wording and is blocked by TASK-260720-2qqq0w and TASK-260720-jrrgw9. Exact candidate is /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/worktree; accepted comparison /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-2kaopg/worktree; immutable rc.5 conformance root /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 with manifest SHA b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c. Inspect current workflow/Make/toolchain/tasks and produce exact implementation map; do not modify TASK-260720-1pvfj5 or any product/spec/CI/pin file, do not pull/install/download, and do not run Go or heavy tests while verifier3 is active. Distinguish committed qualified release pin from explicit non-default candidate input and keep Linux validation non-gating until the existing external prerequisite is available.


### TASK-260729-osjeay_rework-cycle-1.md
> Reviewer-required executable final-CI map correction

Revise the final-CI execution map only. Address every reviewer finding: enumerate both current target-contract conflicts (full Linux candidate suite and full go test -race ./...); provide a board-owner decision packet with exact proposed scope/AC wording; select one executable candidate delivery mechanism with exact runner labels, transport/materialization, immutable identity verification, and a Windows-visible root; retain Linux native as non-gating until its named prerequisite; define exact .github/workflows/ci.yml jobs/steps and Makefile recipes/dependencies; if Linux uses a safe-package allowlist, list every package and an executable new-package drift guard; correct the conformance-root git status command/result to 3 modified and 354 untracked paths (or newly remeasure truthfully); state all commands as future producer gates, not green evidence. Preserve digest and pin invariants; do not edit product, CI, Makefile, target-task fields, spec, or pins.




## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# FIRST/LAST lifecycle commands:
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(TASK-260729-osjeay, status=analysis)'       # analyst-style work
task-board m 'set_status(TASK-260729-osjeay, status=development)'    # implementation / testing work
task-board m 'set_status(TASK-260729-osjeay, status=reviewing)'      # reviewer handoff
task-board m 'set_status(TASK-260729-osjeay, status=blocked)'        # when blocked
task-board m 'set_status(TASK-260729-osjeay, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(TASK-260729-osjeay, item=1)'                        # check item N
task-board m 'add_checklist_item(TASK-260729-osjeay, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(TASK-260729-osjeay, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(TASK-260729-osjeay, name=TASK-260729-osjeay_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add TASK-260729-osjeay ./path/to/file --type outcome --name TASK-260729-osjeay_artifact.bin -d "Description"
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
task-board m 'add_resource(TASK-260729-osjeay, name=TASK-260729-osjeay_results.md, content="...", type=outcome, description="Description")'
task-board resource add TASK-260729-osjeay ./path/to/file --type outcome --name TASK-260729-osjeay_artifact.bin -d "Description"
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
task-board handoff TASK-260729-osjeay --role researcher
```

## Working Directory
Board directory: `/Users/iv/Developer/ReluxWorks/curator/.task-board`

Work in the project root. Do not modify board files directly — always use the `task-board` CLI.
