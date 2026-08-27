# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(BUG-260729-r0fe02, status=reviewing)'
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

- [ ] Reproduce the exact race failure and record competing fail-closed outcomes
- [ ] Implement the narrow deterministic contract without broad error acceptance
- [ ] Run focused repeated race and non-race godriver gates with exact exits
- [ ] Attach a task-scoped patch and outcome for independent review
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

- **ID**: BUG-260729-r0fe02
- **Title**: BUG-260729-r0fe02: stabilize-godriver-cancellation-race-contract
- **Parent**: STORY-260720-3plyvy
### Description

The accepted Curator composite intermittently fails under the mandatory race gate: internal/godriver TestFingerprintCancellationStaysFailClosed/cancelled_between_the_walk_and_the_digest can observe fail-closed toolchain_mutated while the assertion permits only nil or toolchain_timeout. Determine the intended cancellation-vs-mutation precedence and make the contract deterministic without weakening fail-closed behavior.
### Scope

Own only internal/godriver fingerprint cancellation behavior and its focused tests, plus task-scoped evidence. Start from the exact accepted TASK-260720-jrrgw9 candidate. Do not change CI, protocol vectors, timeouts, product manifests outside godriver, or broad error acceptance.
### Acceptance Criteria

A focused -race repetition reliably proves the cancellation boundary contract; legitimate concurrent mutation remains toolchain_mutated, cancellation at the defined phase boundary returns the specified stable fail-closed result, and the test no longer flakes. Non-race focused godriver tests pass, the accepted product behavior is not weakened, and the patch is independently reviewed before TASK-260720-1pvfj5 resumes its final race gate.

## Instructions

The following instructions have been attached to this task:

### BUG-260729-r0fe02_resume.md
> Resume instructions after enforced heavy-test barrier

# Resume after enforced test barrier

RUN-260729-eb4053 was cancelled only because its focused race overlapped TASK-260720-1pvfj5 final-race2. The task worktree and evidence are preserved at .temp/BUG-260729-r0fe02.

Review and continue the existing narrow fingerprint.go + fingerprint_equivalence_test.go patch. Compare against the exact accepted TASK-260720-jrrgw9 godriver bytes. Run only focused internal/godriver tests: deterministic cancellation tests, negative mutation control, and a bounded repeated -race gate. Do not run ./... or any Curator full/race suite. Ensure cancellation raised by this package maps to toolchain_timeout while genuine filesystem mutation remains toolchain_mutated. Attach patch/results and handoff for review.


### BUG-260729-r0fe02_combined-evidence.md
> Precondition for lint-compatible cancellation evidence after BUG-260729-1o0m8f acceptance

Resume only to close the inherited lint gate. Apply the exact accepted BUG-260729-1o0m8f_lint-fix.patch from .temp/resources/BUG-260729-1o0m8f/ on top of the existing BUG-260729-r0fe02 worktree. Do not alter the cancellation implementation or tests. Prove the cancellation task patch hash/content is unchanged, run exact pinned golangci-lint v2.12.2 and only narrow non-race/focused checks needed for combined compatibility. Reuse the already attached repeated race evidence; do not run a full go test or full race suite. Update the task-scoped outcome/logbook, check the remaining lint item, and hand off for review.




## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# FIRST/LAST lifecycle commands:
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(BUG-260729-r0fe02, status=analysis)'       # analyst-style work
task-board m 'set_status(BUG-260729-r0fe02, status=development)'    # implementation / testing work
task-board m 'set_status(BUG-260729-r0fe02, status=reviewing)'      # reviewer handoff
task-board m 'set_status(BUG-260729-r0fe02, status=blocked)'        # when blocked
task-board m 'set_status(BUG-260729-r0fe02, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(BUG-260729-r0fe02, item=1)'                        # check item N
task-board m 'add_checklist_item(BUG-260729-r0fe02, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(BUG-260729-r0fe02, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(BUG-260729-r0fe02, name=BUG-260729-r0fe02_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add BUG-260729-r0fe02 ./path/to/file --type outcome --name BUG-260729-r0fe02_artifact.bin -d "Description"
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
task-board m 'add_resource(BUG-260729-r0fe02, name=BUG-260729-r0fe02_results.md, content="...", type=outcome, description="Description")'
task-board resource add BUG-260729-r0fe02 ./path/to/file --type outcome --name BUG-260729-r0fe02_artifact.bin -d "Description"
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
