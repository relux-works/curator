# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260729-3nx97g, status=reviewing)'
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

- [ ] Import exact accepted rc.5 candidate bytes and accepted rc.4 build-driver generator/fixture semantics into one isolated task worktree with recorded provenance
- [ ] Regenerate complete manager-worker-v1 build-driver vectors and expected artifacts with independently checked positive and negative identities
- [ ] Prove deterministic double regeneration, complete manifest inventory and byte preservation outside the owned golden surface
- [ ] Run curator-spec validation/generator/release-candidate gates and Curator candidate metadata tests without skip using an explicit conformance root
- [ ] Attach candidate-only evidence and hand off for independent review without landing, publication, tag or pin mutation
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

- **ID**: TASK-260729-3nx97g
- **Title**: TASK-260729-3nx97g: regenerate-rc5-build-driver-goldens
- **Parent**: STORY-260720-35dck7
### Description

Carry the independently accepted schema-6 build-driver golden suite forward into the exact accepted rc.5 TASK-260728-2kp3tv candidate under execution_policy=manager-worker-v1.
### Scope

Work in a new isolated curator-spec task worktree. Reuse the accepted TASK-260720-1s1vr6 generator, fixture and expected-artifact semantics while taking the exact TASK-260728-2kp3tv rc.5 snapshot as the candidate base. Regenerate conformance/v1/vectors/build-drivers.json and conformance/v1/expected/build-driver/, include them in the deterministic manifest, update implementation-neutral generator/tests only as required by rc.5 canonical identity, and make Curator candidate metadata-artifact tests execute rather than skip. Preserve schema-1 through schema-6 declaration bytes and all unrelated rc.5 bytes. No Curator/CocoaSkills product edits, stage, commit, tag, publication, pin advancement or release claim.
### Acceptance Criteria

The rc.5 candidate publishes a complete deterministic build-driver vector and expected-artifact suite; the positive portable input requires execution_policy=manager-worker-v1 and independently recomputes cache key sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b and receipt sha256:919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd; legacy rc.4 and reserved-hardened inputs are explicit schema-invalid non-alias negatives; every prior positive, rejection and byte-edge cluster remains represented or is explicitly superseded; two clean regenerations are byte-identical; validate, Python/Go generator tests, release-candidate checks and Curator TestCandidateBuildMetadataArtifacts pass without skip against the explicit candidate root; all unrelated accepted rc.5 and frozen legacy bytes remain unchanged; evidence is candidate-only and authorizes no landing or publication.

## Instructions

No specific instructions attached. Work according to the task description and acceptance criteria above.

## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# FIRST/LAST lifecycle commands:
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(TASK-260729-3nx97g, status=analysis)'       # analyst-style work
task-board m 'set_status(TASK-260729-3nx97g, status=development)'    # implementation / testing work
task-board m 'set_status(TASK-260729-3nx97g, status=reviewing)'      # reviewer handoff
task-board m 'set_status(TASK-260729-3nx97g, status=blocked)'        # when blocked
task-board m 'set_status(TASK-260729-3nx97g, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(TASK-260729-3nx97g, item=1)'                        # check item N
task-board m 'add_checklist_item(TASK-260729-3nx97g, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(TASK-260729-3nx97g, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(TASK-260729-3nx97g, name=TASK-260729-3nx97g_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add TASK-260729-3nx97g ./path/to/file --type outcome --name TASK-260729-3nx97g_artifact.bin -d "Description"
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
task-board m 'add_resource(TASK-260729-3nx97g, name=TASK-260729-3nx97g_results.md, content="...", type=outcome, description="Description")'
task-board resource add TASK-260729-3nx97g ./path/to/file --type outcome --name TASK-260729-3nx97g_artifact.bin -d "Description"
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
