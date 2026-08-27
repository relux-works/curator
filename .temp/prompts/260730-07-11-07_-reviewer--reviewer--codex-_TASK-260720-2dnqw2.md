# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260720-2dnqw2, status=reviewing)'
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
- [ ] Cover canonical input, receipt, marker v1 and v2, and every metadata mismatch with exact shared hashes.
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

- **ID**: TASK-260720-2dnqw2
- **Title**: Implement canonical build metadata models
- **Parent**: STORY-260720-1uv5gi
### Description

Implement portable logical build input including the rc.5-required execution_policy member, CCJ-1 cache keys, exact canonical receipt bytes and hashes, and install-marker v2 models while keeping physical csk cache paths implementation-specific.
### Scope

Own typed metadata modules under src/csk/builds, a dedicated install-marker model module, narrowly shared CCJ-1 support in src/csk/protocol_json.py, and focused tests. Model the complete go-v1 input including build source, root, command, source directory, native target, toolchain, and the fixed policy, whose required rc.5 common.schema.json goBuildPolicyV1 members are module_mode, network, workspace, cgo, compiler_directives, target_mode, link_mode, libgcc, package_assembly, host_objects, telemetry and execution_policy. Parse and canonicalize receipt schema 1 and marker schema 2. Keep marker schema 1 readable for schema 1 through 5 installs. Take every golden byte from the caller-supplied rc.5 candidate conformance root under CURATOR_CONFORMANCE_ROOT rather than from a private reimplementation. build-receipt-v2 and install-marker-v3 belong to the go-repository-v1 external-repository line and are out of scope here; the go-v1 compiled-build surface stays receipt v1 plus marker v2. Conformance-claim emission is out of scope and stays with TASK-260720-12r55p, which targets conformance-claim-v3 under rc.5. Do not implement filesystem trust, compiler execution, cache layout, status, or installer mutation.
### Acceptance Criteria

The canonical go-v1 input requires policy.execution_policy = manager-worker-v1. CCJ-1 input bytes reproduce the shared 869-byte expected/build-driver/build-input.ccj.json and derive the shared cache key sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b. Exact stored receipt bytes reproduce the shared 1120-byte expected/build-driver/receipt.ccj.json, contain no BOM, whitespace, or trailing newline, and derive the shared receipt hash sha256:919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd. Both legacy identities from vectors/build-drivers.json cache_identity are required non-alias negatives with schema_valid false: legacy_rc4_without_execution_policy sha256:3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48 with no execution_policy, and reserved_hardened sha256:13736230d33ce59de7f7323dcd4cffd510655ad8dabd5ee9e8b6cb182ec70037 with execution_policy hardened-worker-v1; aliases is false and the three keys stay numerically distinct, so ignoring execution_policy can never produce the portable key. Readers reject duplicate keys, unsafe integers, noncanonical bytes, unknown fields, mismatched keys or input, missing or unknown execution_policy, wrong derived artifact paths, unsupported versions, and malformed identities. Marker v2 deterministically sorts roots, commands, dependencies, files, and builds; matches expected/build-driver/marker.json with schema_version 2, skill_schema_version 6, and a builds entry carrying driver go-v1 plus cache_key, receipt_sha256, artifact_path and artifact_sha256; requires build_source exactly when builds are active; and keeps valid marker v1 current for pre-v6 packages. Every golden is read from the caller-supplied rc.5 candidate root whose manifest.json hashes to sha256:b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c, and the suite digest is recorded as non-release evidence. Focused pytest and strict mypy pass.

## Instructions

No specific instructions attached. Work according to the task description and acceptance criteria above.

## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# FIRST/LAST lifecycle commands:
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(TASK-260720-2dnqw2, status=analysis)'       # analyst-style work
task-board m 'set_status(TASK-260720-2dnqw2, status=development)'    # implementation / testing work
task-board m 'set_status(TASK-260720-2dnqw2, status=reviewing)'      # reviewer handoff
task-board m 'set_status(TASK-260720-2dnqw2, status=blocked)'        # when blocked
task-board m 'set_status(TASK-260720-2dnqw2, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(TASK-260720-2dnqw2, item=1)'                        # check item N
task-board m 'add_checklist_item(TASK-260720-2dnqw2, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(TASK-260720-2dnqw2, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(TASK-260720-2dnqw2, name=TASK-260720-2dnqw2_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add TASK-260720-2dnqw2 ./path/to/file --type outcome --name TASK-260720-2dnqw2_artifact.bin -d "Description"
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
task-board m 'add_resource(TASK-260720-2dnqw2, name=TASK-260720-2dnqw2_results.md, content="...", type=outcome, description="Description")'
task-board resource add TASK-260720-2dnqw2 ./path/to/file --type outcome --name TASK-260720-2dnqw2_artifact.bin -d "Description"
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
