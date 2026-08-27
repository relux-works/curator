# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260720-z9j4c9, status=reviewing)'
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

- [ ] Record the base SHA, fast-forward the clean local main clone to current origin/main before creating the task-scoped worktree, and include every blocked-by handoff.
- [ ] Cover the accepted positive schema 6 shape and every static manifest rejection with focused unit tests.
- [ ] Run focused pytest plus python -m mypy and attach task-scoped evidence.
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

- **ID**: TASK-260720-z9j4c9
- **Title**: Implement schema v6 build manifest model
- **Parent**: STORY-260720-1uv5gi
### Description

Extend the Python manifest domain and skill validation with the accepted schema 6 surface: top-level build_roots and a closed build command containing only type build, driver go-v1, and source_dir. Preserve schema versions 1 through 5, canonical and legacy manifest parity, dependency closure activation, and the no-build legacy fallback.
### Scope

Target the cocoaskills repository from a clean, fast-forwarded local main and a task-scoped worktree based on all landed dependencies. Own src/csk/skillspec.py, the build-domain package initializer, validation-only changes in src/csk/skillcheck.py, and focused parser and skill-check tests. Validate static paths, root use and overlap, and module-root relationships without invoking Go. Do not implement hashing, toolchain probes, compiler execution, cache storage, or install mutation.
### Acceptance Criteria

Schema 6 accepts strict build commands alongside existing script and system commands and retains schema 5 capabilities and dependencies. build_roots are unique, pairwise disjoint, non-dot, real link-free directories that do not overlap runtime_roots; every root is used; every source_dir is below exactly one root whose direct go.mod is its nearest module file. Unknown drivers or fields, mixed shapes, schema 1 through 5 build surfaces, legacy runtime builds, root or escaped sources, missing or non-directory paths, unused or overlapping roots, runtime overlap, and nested modules fail before activation. Existing schema 1 through 5 behavior remains byte-semantically compatible. Focused pytest and strict mypy pass.

## Instructions

The following instructions have been attached to this task:

### TASK-260720-z9j4c9_execution-brief.md
> CocoaSkills schema v6 implementation brief and repository routing

# CocoaSkills schema v6 execution brief

Repository source is /Users/iv/Developer/ReluxWorks/cocoaskills-production (origin git@github.com:ivanopcode/cocoaskills.git). The clean canonical main worktree is /Users/iv/Developer/Wildberries/cocoaskills. Do not work on agent/registry-client-production because its upstream is gone. Begin only after TASK-260720-1pvfj5 is done. Record the accepted Curator handoff and final origin/main SHA. Fast-forward the clean canonical main worktree once with fetch plus ff-only, then create a task-owned worktree under /Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z9j4c9/worktree from that exact base. If another task already advanced canonical main to the same origin/main SHA, verify and reuse that base; never reset or overwrite.

Implement only the Python schema-v6 manifest/domain/validation slice described by the task and accepted curator-spec schema-6 contract. Use /Users/iv/Developer/ReluxWorks/curator-spec as the normative source and /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1pvfj5/rework/composite as the accepted Go parity reference. Preserve schema 1-5 and legacy behavior. Add focused positive/rejection/parity/dependency-activation/no-build tests, run focused pytest and python -m mypy, attach exact evidence, and route to independent review. Do not stage, commit, publish, change pins, or begin transaction/build execution beyond the schema task.


### TASK-260720-z9j4c9_review-brief.md
> Independent review of CocoaSkills schema v6 implementation

# Independent schema-v6 review

Review /Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z9j4c9/worktree at exact base 6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12 against curator-spec schema 6 and accepted Curator parity behavior. Inspect TASK-260720-z9j4c9_results.md. Verify scope stays in skillspec.py, skillcheck.py, builds/__init__.py and focused tests; schema 1-5 and canonical/legacy parity remain intact; build_roots and the closed type=build driver=go-v1 source_dir command reject every extra/invalid shape; dependency activation and no-build fallback remain compatible. Independently rerun focused pytest and mypy or audit the full evidence where redundant. Do not modify code, push, land, tag, release, or touch wb. Attach a precise verdict and route accepted work to done; otherwise route to-dev.




## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# FIRST/LAST lifecycle commands:
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(TASK-260720-z9j4c9, status=analysis)'       # analyst-style work
task-board m 'set_status(TASK-260720-z9j4c9, status=development)'    # implementation / testing work
task-board m 'set_status(TASK-260720-z9j4c9, status=reviewing)'      # reviewer handoff
task-board m 'set_status(TASK-260720-z9j4c9, status=blocked)'        # when blocked
task-board m 'set_status(TASK-260720-z9j4c9, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(TASK-260720-z9j4c9, item=1)'                        # check item N
task-board m 'add_checklist_item(TASK-260720-z9j4c9, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(TASK-260720-z9j4c9, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(TASK-260720-z9j4c9, name=TASK-260720-z9j4c9_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add TASK-260720-z9j4c9 ./path/to/file --type outcome --name TASK-260720-z9j4c9_artifact.bin -d "Description"
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
task-board m 'add_resource(TASK-260720-z9j4c9, name=TASK-260720-z9j4c9_results.md, content="...", type=outcome, description="Description")'
task-board resource add TASK-260720-z9j4c9 ./path/to/file --type outcome --name TASK-260720-z9j4c9_artifact.bin -d "Description"
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
