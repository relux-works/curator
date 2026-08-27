# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260729-rfrdfo, status=reviewing)'
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

- [ ] Task-owned baseline and post-manifest prove only the 13 allowlisted test files changed
- [ ] Patch applies cleanly to the exact current jrrgw9 candidate without mutating it
- [ ] Approved install and atomicity focused commands pass with count=1 and no timeout override
- [ ] Parallelization and injectClasses preserve core assertions and document retired cross-class sequencing
- [ ] Main candidate remains byte-identical and task-scoped evidence is attached
- [ ] Code written per task description and AC
- [ ] Relevant tests written for new or changed behavior and passing
- [ ] Lint clean
- [ ] Relevant build/validation commands run after changes and build not broken
- [ ] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [ ] Findings written to file
- [ ] Key aspects highlighted
- [ ] Fact-checking performed — claims verified, sources cited
- [ ] Findings linked on the board as a new task-scoped outcome resource
- [ ] All questions from task description answered
- [ ] Implementation matches AC
- [ ] Solution fits project architecture
- [ ] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches
## Your Task

- **ID**: TASK-260729-rfrdfo
- **Title**: TASK-260729-rfrdfo: prototype-install-race-timeout-patch
- **Parent**: STORY-260720-3plyvy
### Description

Prototype the reviewer-sanctioned test-only install race-timeout optimization in a task-owned copy of the exact TASK-260720-jrrgw9 candidate. Produce an applicable patch and focused evidence without mutating the main candidate.
### Scope

Copy the exact current jrrgw9 worktree into private task storage; capture a path-sorted SHA-256 source baseline; modify only the 13-file required allowlist from TASK-260729-3dr6hw revision 3. Add t.Parallel only to the approved 88 tests and partition atomicity injection with a separate injectClasses selector while retaining full scenario.classes success coverage. Do not touch product Go, aba_test.go, atomicity/fixture_test.go, timeouts, skips, assertions outside the explicitly retired cross-class defense-in-depth sequence, conformance root, pins, accepted worktrees, or shared caches. Run gofmt and only focused package/test commands with -count=1; no go test ./..., no full package race, no host install, stage, commit, publish or pin.
### Acceptance Criteria

A task-owned source baseline and post-manifest prove only the 13 allowlisted test files changed; an exact patch applies cleanly to the current jrrgw9 candidate; focused install and atomicity commands compile and pass without timeout overrides; assertion-retention and deliberate cross-class defense-in-depth retirement are documented; main candidate remains byte-identical.

## Instructions

The following instructions have been attached to this task:

### TASK-260729-rfrdfo_prototype-input.md
> Exact prototype inputs and non-overlap boundary

# Prototype input

Source: byte-for-byte private copy of /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/worktree. Accepted comparison: /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-2kaopg/worktree. Immutable conformance root: /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1.

Authoritative plan: TASK-260729-3dr6hw_install-race-timeout-diagnosis.md revision 3 plus cycle-2 reviewer corrections: expected rsync delta remains 20 deleting + 3 existing modified + 11 newly modified = 34 lines; saveJournal has 23 call sites; atomicity uses a distinct injectClasses selector while scenario.classes remains full 7/5 coverage; cross-class chaining is explicitly retired defense in depth, while whole-state digest, cutoff, reverse rollback, journal cleanup, and full success coverage remain.

Prototype only. Never mutate the source candidate or accepted comparison. Capture pre/post sorted SHA-256 manifests and verify the source candidate remains byte-identical. No full repository or full race command.


### TASK-260729-rfrdfo_evidence-only-constraint.md
> Fail-closed no-Go scope for cycle-2 evidence reconciliation

# Evidence-only rework constraint

Run no Go command of any kind: no go test, probe, benchmark, build, vet, lint, helper script that invokes Go, or detached process. Make no product/prototype edits. Use only stored TASK-260729-2afulh tester evidence: non-race 273s; isolated race1 exit 0 in 493s; race2/3 deliberately not completed after the strict <=480s bar was conclusively missed. Correct rejected arithmetic and route fixture-only as rejected in favor of TASK-260729-365r5r. Partial probe artifacts from cancelled RUN-260729-5e854d are invalid and must not be cited.




## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# FIRST/LAST lifecycle commands:
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(TASK-260729-rfrdfo, status=analysis)'       # analyst-style work
task-board m 'set_status(TASK-260729-rfrdfo, status=development)'    # implementation / testing work
task-board m 'set_status(TASK-260729-rfrdfo, status=reviewing)'      # reviewer handoff
task-board m 'set_status(TASK-260729-rfrdfo, status=blocked)'        # when blocked
task-board m 'set_status(TASK-260729-rfrdfo, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(TASK-260729-rfrdfo, item=1)'                        # check item N
task-board m 'add_checklist_item(TASK-260729-rfrdfo, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(TASK-260729-rfrdfo, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(TASK-260729-rfrdfo, name=TASK-260729-rfrdfo_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add TASK-260729-rfrdfo ./path/to/file --type outcome --name TASK-260729-rfrdfo_artifact.bin -d "Description"
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
task-board m 'add_resource(TASK-260729-rfrdfo, name=TASK-260729-rfrdfo_results.md, content="...", type=outcome, description="Description")'
task-board resource add TASK-260729-rfrdfo ./path/to/file --type outcome --name TASK-260729-rfrdfo_artifact.bin -d "Description"
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
