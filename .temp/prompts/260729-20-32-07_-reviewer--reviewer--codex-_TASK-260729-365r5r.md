# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260729-365r5r, status=reviewing)'
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

- [ ] Record exact source/prototype baseline and path-sorted pre-manifest
- [ ] Document literal product and test file/function allowlist before edits
- [ ] Prove all new, resumed, recovered, and externally decoded target graphs are validated before mutation
- [ ] Implement cached/hoisted namespace validation without changing journal schema or timeout behavior
- [ ] Add focused negative tests for malformed and overlapping target namespaces
- [ ] Run sequential focused non-race and race timing gates with real exit codes
- [ ] Attach task-scoped outcome and hand off for independent review
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

- **ID**: TASK-260729-365r5r
- **Title**: TASK-260729-365r5r: prototype-savejournal-namespace-validation
- **Parent**: STORY-260720-3plyvy
### Description

Build an isolated production-side prototype that removes repeated O(P^2) target-namespace validation from saveJournal while preserving fail-closed validation before any transaction mutation and all journal integrity checks.
### Scope

Use a task-owned worktree copied from TASK-260729-rfrdfo prototype state or the pristine jrrgw9 candidate as explicitly recorded. Limit product edits to internal/transaction and tests needed to prove validation timing/coverage. Do not touch the main candidate, timeout values, CI, protocol/spec behavior, or weaken/remove assertions. Run only focused transaction/atomicity/install measurements sequentially; no ./... until independent tester stage.
### Acceptance Criteria

Literal file/function allowlist and pre/post manifests; static call-path proof that every externally supplied or recovered journal target graph is validated before mutation; every saveJournal call still revalidates current filesystem namespace facts but performs at most O(P) filesystem identity/resolution reads per pass instead of repeating them for O(P^2) pairs; malformed, overlapping, and between-save symlink/alias changes fail closed; focused non-race and race evidence demonstrates a defensible atomicity margin at or below 480 seconds or explicitly rejects the prototype; independent review required before integration.

## Instructions

The following instructions have been attached to this task:

### TASK-260729-365r5r_reentry-constraint.md
> Fail-closed scope for post-driver evidence packaging

# Re-entry constraint

This run packages already completed production gates. It MUST NOT run or start baseline scripts, atomicity tests, go test, go build, go vet, benchmarks, conformance probes, detached processes, or product edits. Existing valid gate files are authoritative. It MAY run exactly one command after proving the shared process barrier is empty: /Users/iv/go/bin/golangci-lint run in the prototype worktree, with a real exit code captured. Then update task-scoped evidence/checklists and hand off to review. Any partial baseline rerun artifacts from cancelled runs are invalid and must be explicitly excluded.


### TASK-260729-365r5r_lint-rework-constraint.md
> Narrow fix for three prototype-introduced revive findings

# Lint-only rework constraint

Apply only the three mechanical prototype lint fixes identified by gate-lint-abs: in internal/transaction/namespace_pass_test.go rename the unused t *testing.T closure parameter to _ *testing.T at the nested, exact-repeat, and cross-target-backup table cases. No other product/test/source change is allowed, especially no namespace.go edit. Regenerate the two-file prototype patch and pre/post manifest evidence. After proving the shared Go barrier is empty, run only gofmt for the two changed paths, go test -count=1 ./internal/transaction, go test -count=1 -race ./internal/transaction, the focused namespace verbose tests, and /Users/iv/go/bin/golangci-lint run ./internal/transaction/... sequentially with real exits. Do not rerun atomicity, install, benchmarks, baseline scripts, or detached processes. Preserve prior accepted 66s and 84/76/75s performance evidence as behavior-neutral pre-rework evidence, explicitly recording why the parameter-only fix does not stale it. Full lint may retain one inherited untouched godriver ineffassign; acceptance requires zero introduced/transaction lint findings. No commit/stage/publish.




## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# FIRST/LAST lifecycle commands:
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(TASK-260729-365r5r, status=analysis)'       # analyst-style work
task-board m 'set_status(TASK-260729-365r5r, status=development)'    # implementation / testing work
task-board m 'set_status(TASK-260729-365r5r, status=reviewing)'      # reviewer handoff
task-board m 'set_status(TASK-260729-365r5r, status=blocked)'        # when blocked
task-board m 'set_status(TASK-260729-365r5r, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(TASK-260729-365r5r, item=1)'                        # check item N
task-board m 'add_checklist_item(TASK-260729-365r5r, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(TASK-260729-365r5r, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(TASK-260729-365r5r, name=TASK-260729-365r5r_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add TASK-260729-365r5r ./path/to/file --type outcome --name TASK-260729-365r5r_artifact.bin -d "Description"
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
task-board m 'add_resource(TASK-260729-365r5r, name=TASK-260729-365r5r_results.md, content="...", type=outcome, description="Description")'
task-board resource add TASK-260729-365r5r ./path/to/file --type outcome --name TASK-260729-365r5r_artifact.bin -d "Description"
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
