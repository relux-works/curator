# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260730-2kfjm8, status=reviewing)'
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

- [ ] Exact accepted composite base equals current origin/main 17804ce and no unrelated local state is included
- [ ] Commit tree reproduces the accepted 374-entry manifest and released protocol pin remains rc.3
- [ ] Independent reviewer accepts the exact commit before main push
- [ ] Reviewed commit is fast-forward pushed to relux-works/curator main and v0.13.0 GitHub release is created
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

- **ID**: TASK-260730-2kfjm8
- **Title**: TASK-260730-2kfjm8: land-accepted-curator-composite
- **Parent**: STORY-260720-3plyvy
### Description

Create one reviewable commit from the independently accepted TASK-260720-1pvfj5 final composite at exact origin/main base 17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8, verify its tree against the accepted 374-entry manifest, and land that exact reviewed commit to origin/main without advancing the released protocol pin or including local board/research state.
### Scope

Accepted Curator composite commit, independent verification, and fast-forward main landing
### Acceptance Criteria

The exact independently accepted 374-entry Curator composite is committed without byte drift or unrelated local state, independently reviewed to done, then that exact commit is fast-forward pushed to relux-works/curator main while SPEC_PIN remains rc.3. Tag and GitHub Release are explicitly deferred until a new human command.

## Instructions

The following instructions have been attached to this task:

### TASK-260730-2kfjm8_commit-brief.md
> Create exact Curator v0.13.0 landing commit from accepted composite

# Curator landing commit

Work only in /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1pvfj5/rework/composite. Fetch origin/main and require it remains exactly 17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8, equal to worktree HEAD; fail on drift. The worktree is the independently accepted TASK-260720-1pvfj5 final composite. Create branch release/curator-v0.13.0-candidate at this HEAD, stage the complete accepted working-tree delta only, and verify no path outside the 230 modified/new status entries is added. Verify the resulting committed tree reproduces TASK-260720-1pvfj5_candidate-input-rework-manifest.txt across all 374 entries, exactly one SPEC_PIN remains 00b1688a9b2457ca397a0bb550acf47cad8ee967, and candidate identity remains candidate-only. Create one commit with message Add declarative compiled skill builds. Do not push, tag, create GitHub release, modify bytes, rerun heavy suites, or touch the dirty primary worktree. Attach commit SHA, tree SHA, staged path list/hash, and verification exits; route to review.


### TASK-260730-2kfjm8_review-brief.md
> Independent review of exact Curator landing commit

# Independent Curator landing review

Review exact local commit cfffd7cd7be33aba9ca44d26993ce5ab19b5fa4d on branch release/curator-v0.13.0-candidate in /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1pvfj5/rework/composite. Parent must be current origin/main 17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8 and commit tree must be 2ce3f14440c5ae8104ef2d9c1fa73908a84553fc. Independently verify exact 230-path delta and all 374 entries against TASK-260720-1pvfj5_candidate-input-rework-manifest.txt, SPEC_PIN rc.3 unchanged, candidate-only CI semantics, clean worktree, one commit, and no remote branch/tag/release. Reuse accepted heavy test evidence by byte identity; do not rerun heavy suites or modify/push/tag/release. Attach verdict. If accepted, route task to done despite the explicitly superseded release clause in checklist item 4; main landing will occur only after done and release remains deferred.




## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# FIRST/LAST lifecycle commands:
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(TASK-260730-2kfjm8, status=analysis)'       # analyst-style work
task-board m 'set_status(TASK-260730-2kfjm8, status=development)'    # implementation / testing work
task-board m 'set_status(TASK-260730-2kfjm8, status=reviewing)'      # reviewer handoff
task-board m 'set_status(TASK-260730-2kfjm8, status=blocked)'        # when blocked
task-board m 'set_status(TASK-260730-2kfjm8, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(TASK-260730-2kfjm8, item=1)'                        # check item N
task-board m 'add_checklist_item(TASK-260730-2kfjm8, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(TASK-260730-2kfjm8, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(TASK-260730-2kfjm8, name=TASK-260730-2kfjm8_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add TASK-260730-2kfjm8 ./path/to/file --type outcome --name TASK-260730-2kfjm8_artifact.bin -d "Description"
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
task-board m 'add_resource(TASK-260730-2kfjm8, name=TASK-260730-2kfjm8_results.md, content="...", type=outcome, description="Description")'
task-board resource add TASK-260730-2kfjm8 ./path/to/file --type outcome --name TASK-260730-2kfjm8_artifact.bin -d "Description"
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
