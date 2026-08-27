# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260720-3j8pp5, status=reviewing)'
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

- [ ] Record the base SHA and create the task worktree only after the clean local main clone is fast-forwarded and dependency handoffs are present.
- [ ] Test exact probe argv, clean environment, private telemetry state, target tuning, and byte-exact toolchain vectors.
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

- **ID**: TASK-260720-3j8pp5
- **Title**: Implement trusted Go toolchain identity
- **Parent**: STORY-260720-1uv5gi
### Description

Resolve an operator-trusted native Go executable before entering package-controlled directories, initialize private telemetry-off state, freeze the native target, and compute curator-go-toolchain-v1 exactly.
### Scope

Own src/csk/builds/toolchain.py and its focused tests. Capture the operator process search path before project shim augmentation, reject repository or project-managed Go candidates, resolve the real GOROOT bin/go executable, run only the three package-independent probe forms from an empty bootstrap environment, validate the accepted Go release-family allowlist and native tuning tuple, and fingerprint the complete GOROOT tree. Add no package or manager-config build-policy fields and run no go list or go build.
### Acceptance Criteria

The probe invokes direct argv for go telemetry off, go version, and the fixed go env -json field list, verifies GOTELEMETRY off and GOTELEMETRYDIR inside the operation-private root, and deletes that root on every exit. GOROOT, GOHOSTOS, GOHOSTARCH, GOOS, GOARCH, and exactly one applicable tuning variable are validated and frozen. The framed tree digest handles files, directories, internal relative links, normalized LF or CRLF version output, duplicate and invalid paths, escaping links, executable checks, and tree mutation exactly as shared vectors require. Pre-Go-1.23, unknown release families, wrappers, repository-local Go, and mismatched GOROOT fail closed. Focused pytest and strict mypy pass on POSIX and Windows-safe imports.

## Instructions

The following instructions have been attached to this task:

### TASK-260720-3j8pp5_windows-review-instructions.md
> Exact scope and gates for independent Windows re-review

# Independent review instructions

Review the current uncommitted two-file diff in `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3j8pp5/worktree`. Do not edit product files, stage, commit, push, or bulk-format. Verify the fresh `os.lstat` strategy closes GitHub Windows failures without weakening file/directory mutation detection. Re-run the deterministic fake-DirEntry regression, focused toolchain tests, full pytest with both fixture roots, strict mypy, diff check, package build, and Twine check. Map the fix explicitly to run 30503926948 / Windows job 90749459882. Treat GitHub Windows CI on an exact committed candidate as still required after local acceptance. Record exact hashes and a verdict outcome.


### TASK-260720-3j8pp5_exact-commit-review.md
> Final exact commit and CI attribution review instructions

# Final exact-commit review

Review signed commit `1d28910f5bb276ff58e2a102e06968bd7640abe3` and GitHub run `30505740935` for PR #9. Do not edit, stage, commit, push, or merge. Verify signature, two-file scope, hashes, and that all prior `tests/test_builds_toolchain.py` Windows failures are closed on Python 3.11-3.14. Inspect every failed Windows summary and confirm the only remaining failures are the same eight `tests/test_build_source.py` failures owned by `TASK-260720-3c0ss2` / PR #8. Re-run the focused toolchain suite and strict mypy from exact commit if needed. Accept `done` only if landing #9 is safe as an independently closed task despite the separately tracked source failures; otherwise route precise changes requested.




## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# FIRST/LAST lifecycle commands:
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(TASK-260720-3j8pp5, status=analysis)'       # analyst-style work
task-board m 'set_status(TASK-260720-3j8pp5, status=development)'    # implementation / testing work
task-board m 'set_status(TASK-260720-3j8pp5, status=reviewing)'      # reviewer handoff
task-board m 'set_status(TASK-260720-3j8pp5, status=blocked)'        # when blocked
task-board m 'set_status(TASK-260720-3j8pp5, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(TASK-260720-3j8pp5, item=1)'                        # check item N
task-board m 'add_checklist_item(TASK-260720-3j8pp5, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(TASK-260720-3j8pp5, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(TASK-260720-3j8pp5, name=TASK-260720-3j8pp5_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add TASK-260720-3j8pp5 ./path/to/file --type outcome --name TASK-260720-3j8pp5_artifact.bin -d "Description"
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
task-board m 'add_resource(TASK-260720-3j8pp5, name=TASK-260720-3j8pp5_results.md, content="...", type=outcome, description="Description")'
task-board resource add TASK-260720-3j8pp5 ./path/to/file --type outcome --name TASK-260720-3j8pp5_artifact.bin -d "Description"
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
