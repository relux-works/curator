# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260728-wy3dsw, status=reviewing)'
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

## Status Transitions

- **start_status:** `reviewing`
- **end_status:** no unconditional default; the reviewer must set exactly one verdict status: `done`, `to-dev`, `analysis`, or evidence-backed `blocked`

## Definition of Done

- Implementation matches AC
- Solution fits project architecture
- Tests green
- If review does not accept the work — verdict evidence added via `task-board m 'set_notes(...)'` or an outcome resource, then status routed by the explicit verdict branches above

## Constraints

Does NOT modify code. Read-only access.

## Skills

These role skill references are a lazy catalog, not a mandate to bulk-read every
body. Before technical work, identify the skills relevant to this task's concrete
scope and read those full skill bodies. Always read any skill explicitly required
by the task, user, or project instructions.

- **project-management**: `/Users/iv/.claude/skills/project-management/SKILL.md`
- **architecture-diagrams**: `/Users/iv/.claude/skills/architecture-diagrams/SKILL.md`

## Definition of Done

- [ ] Validate exact Git, offline, audit, cache, transaction, status, repair, and GC examples
- [ ] Document stable typed failures and package-controlled behavior rejections
- [ ] Docs updated and consistent with current code
- [ ] No discrepancies between code and description
- [ ] Result linked as a new task-scoped outcome resource
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [ ] Implementation matches AC
- [ ] Solution fits project architecture
- [ ] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches
## Your Task

- **ID**: TASK-260728-wy3dsw
- **Title**: TASK-260728-wy3dsw: external-repository-manager-profile-and-lifecycle-contract
- **Parent**: STORY-260728-10wxx2
### Description

Specify the rc.5 manager profile and CLI lifecycle for resolving, validating, auditing, compiling, installing, reporting, repairing, and collecting external build repositories. Pin the exact clean Git process, raw-object proof, LFS and special-file rejection, snapshot/cache identity, mixed planning, transaction, PATH/shim verification, offline semantics, and status/repair/GC behavior.
### Scope

curator-spec manager-profile prose, CLI contracts, lifecycle state machine, stable diagnostics, and normative command examples. Reuse the accepted go-v1 trusted toolchain/build session without changing its schema-6 meaning.
### Acceptance Criteria

The profile fixes supported transports and canonical identity, exact init/fetch/ref flows, SSH isolation, local-substitution admission, pack/index and cat-file grammar, commit/tag/tree/blob recomputation, complete-object and all-blob LFS proof, audit ordering, cache and receipt keys, marker currentness, rollback, repair and GC roots; tagged declarations always exact-fetch the tag as the sole path; syntax-only offline warning and install/audit failure are unambiguous; no repository-controlled hooks, helpers, filters, argv, env, output, credentials, signing, or lazy network reads remain possible; profile examples and validation pass.

## Instructions

No specific instructions attached. Work according to the task description and acceptance criteria above.

## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# The FIRST/LAST sections above define your role-default lifecycle commands.
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(TASK-260728-wy3dsw, status=analysis)'       # analyst-style work
task-board m 'set_status(TASK-260728-wy3dsw, status=development)'    # implementation / testing work
task-board m 'set_status(TASK-260728-wy3dsw, status=reviewing)'      # reviewer handoff
task-board m 'set_status(TASK-260728-wy3dsw, status=blocked)'        # when blocked
task-board m 'set_status(TASK-260728-wy3dsw, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(TASK-260728-wy3dsw, item=1)'                        # check item N
task-board m 'add_checklist_item(TASK-260728-wy3dsw, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(TASK-260728-wy3dsw, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(TASK-260728-wy3dsw, name=TASK-260728-wy3dsw_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add TASK-260728-wy3dsw ./path/to/file --type outcome --name TASK-260728-wy3dsw_artifact.bin -d "Description"
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
task-board m 'add_resource(TASK-260728-wy3dsw, name=TASK-260728-wy3dsw_results.md, content="...", type=outcome, description="Description")'
task-board resource add TASK-260728-wy3dsw ./path/to/file --type outcome --name TASK-260728-wy3dsw_artifact.bin -d "Description"
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


# Execution Continuation Context

Prompt mode: full
Snapshot digest: sha256:dc74fb21059ca31536efa7bfed61fbd64243a5bf4c458f913291454d56f94683

### checklist/definition_of_done

Digest: sha256:7876d9cbfb4c63dba64f8d042d0d467dc1253edbae505a9ee2af3bf97d4825d3
Length: 604 bytes
Body:
["Validate exact Git, offline, audit, cache, transaction, status, repair, and GC examples","Document stable typed failures and package-controlled behavior rejections","Docs updated and consistent with current code","No discrepancies between code and description","Result linked as a new task-scoped outcome resource","Important findings, decisions, anomalies, or regressions recorded in logbook when relevant","Implementation matches AC","Solution fits project architecture","Tests green","If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches"]

### execution_policy/digest

Digest: sha256:631dca31732b29a626b42e07097ae1301645ac06894e609be1e3f52549f88596
Length: 71 bytes
Body:
sha256:7095dca731199937f8161a647cc0f5e332ffcb6774192a511250d19105e1733b

### final_audit_assignment/current

Digest: sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945
Length: 2 bytes
Body:
[]

### goal/revision

Digest: sha256:74234e98afe7498fb5daf1f36ac2d78acc339464f950703b8c019892f982b90b
Length: 4 bytes
Body:
null

### next_action/reason

Digest: sha256:205f87c0be0dc5b66d88aef7dc6935832fe5f6bfd275ca07da3a91392244f026
Length: 22 bytes
Body:
board_status:to-review

### post_rework_validation/current

Digest: sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945
Length: 2 bytes
Body:
[]

### producer_resolutions/current

Digest: sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945
Length: 2 bytes
Body:
[]

### reviewer_authorization/current

Digest: sha256:2e2f92c499978b499f71fc5adf3448d4e094f35055ad2a7303212bcc1bff5570
Length: 13 bytes
Body:
{"run_id":""}

### reviewer_findings/current

Digest: sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945
Length: 2 bytes
Body:
[]

### role/body

Digest: sha256:0571aca5102774ae78178dd47e2a0adc0c391b00bd079254def33430ddbda6a0
Length: 2044 bytes
Body:
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

## Status Transitions

- **start_status:** `reviewing`
- **end_status:** no unconditional default; the reviewer must set exactly one verdict status: `done`, `to-dev`, `analysis`, or evidence-backed `blocked`

## Definition of Done

- Implementation matches AC
- Solution fits project architecture
- Tests green
- If review does not accept the work — verdict evidence added via `task-board m 'set_notes(...)'` or an outcome resource, then status routed by the explicit verdict branches above

## Constraints

Does NOT modify code. Read-only access.

### role/metadata

Digest: sha256:ce29a04c3230e295e239ca92ba538b520993f88af99fc1d3c3c84b2ac2869809
Length: 148 bytes
Body:
{"archetype":"reviewer","description":"Reviews implementation quality and project fit","end_status":"","name":"reviewer","start_status":"reviewing"}

### skill/Domain skills as needed to evaluate implementation quality

Digest: sha256:e257caafc9ad9c1ceeff028d093e2a5b3455964373f4b8e4af86dc46a8cec4d4
Length: 79 bytes
Body:
{"name":"Domain skills as needed to evaluate implementation quality","path":""}

### skill/architecture-diagrams

Digest: sha256:4d77caf872bab194004877589bf7b82672514a42abce5c5e1150027e66b119f8
Length: 97 bytes
Body:
{"name":"architecture-diagrams","path":"/Users/iv/.claude/skills/architecture-diagrams/SKILL.md"}

### skill/project-management

Digest: sha256:8d1bab60e478d47f7d2ae88871b9a86b0911c06c1bdfa38e7c8c627b91c9fbae
Length: 91 bytes
Body:
{"name":"project-management","path":"/Users/iv/.claude/skills/project-management/SKILL.md"}

### task/acceptance_criteria

Digest: sha256:1350763a623340ba7bc360ea33497318da8f5802e49562d4044dc5fcf45c1324
Length: 639 bytes
Body:
The profile fixes supported transports and canonical identity, exact init/fetch/ref flows, SSH isolation, local-substitution admission, pack/index and cat-file grammar, commit/tag/tree/blob recomputation, complete-object and all-blob LFS proof, audit ordering, cache and receipt keys, marker currentness, rollback, repair and GC roots; tagged declarations always exact-fetch the tag as the sole path; syntax-only offline warning and install/audit failure are unambiguous; no repository-controlled hooks, helpers, filters, argv, env, output, credentials, signing, or lazy network reads remain possible; profile examples and validation pass.

### task/description

Digest: sha256:329ba84d9e3f436b845b795c3dcb35cc0923cb0d446d6a840721407e86b4b4dd
Length: 387 bytes
Body:
Specify the rc.5 manager profile and CLI lifecycle for resolving, validating, auditing, compiling, installing, reporting, repairing, and collecting external build repositories. Pin the exact clean Git process, raw-object proof, LFS and special-file rejection, snapshot/cache identity, mixed planning, transaction, PATH/shim verification, offline semantics, and status/repair/GC behavior.

### task/scope

Digest: sha256:4fbf90d07ac3ae9898a28e51b72aaf84d556325f94e79f390acc0dcf7cc9a511
Length: 223 bytes
Body:
curator-spec manager-profile prose, CLI contracts, lifecycle state machine, stable diagnostics, and normative command examples. Reuse the accepted go-v1 trusted toolchain/build session without changing its schema-6 meaning.

### task/title

Digest: sha256:044e2514d57517224a97b6f81b4a2036410d18db266ddd2ba35bf0d257b05491
Length: 78 bytes
Body:
TASK-260728-wy3dsw: external-repository-manager-profile-and-lifecycle-contract

### validation_manifest/current

Digest: sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945
Length: 2 bytes
Body:
[]

