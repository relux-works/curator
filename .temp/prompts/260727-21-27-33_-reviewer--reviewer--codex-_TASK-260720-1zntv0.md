# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260720-1zntv0, status=reviewing)'
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

- [ ] Complete package graph and source inputs pass all rejection clusters
- [ ] Only the fixed go list and go build commands can execute
- [ ] Built outputs are verified but never launched during installation
- [ ] Readonly source, network denial, and supported host resource controls are executable platform gates
- [ ] Code written per task description and AC
- [ ] Relevant tests written for new or changed behavior and passing
- [ ] Lint clean
- [ ] Relevant build/validation commands run after changes and build not broken
- [ ] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [ ] Board size is proportional to the spec and is the smallest decomposition that maps every requirement
- [ ] Every story and task traces to a concrete spec requirement; justified-gap elements also carry a self-verified gap record
- [ ] Beyond-literal-spec elements include a written justification naming the gap and the spec and out-of-scope checks performed before creation
- [ ] Research tasks cite an exact question the spec genuinely leaves open
- [ ] Dependencies linked
- [ ] Tasks are atomic — one clear deliverable each
- [ ] Completeness verified — nothing forgotten
- [ ] Any planning artifacts actually produced are linked as new task-scoped outcome resources; diagrams are strictly optional, never a standing deliverable
- [ ] Implementation matches AC
- [ ] Solution fits project architecture
- [ ] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches
## Your Task

- **ID**: TASK-260720-1zntv0
- **Title**: Implement go-v1 preflight and artifact build
- **Parent**: STORY-260720-3plyvy
### Description

Implement the source-aware go-v1 pipeline that validates the complete active package graph with one fixed go list command, builds with one fixed go build command, verifies the staged executable, and never transfers control to package code.
### Scope

Continue in internal/godriver using the trusted session and frozen build source. Set CWD to canonical source_dir and expose a build-tagged host-execution policy boundary. Present the validated snapshot and vendored module tree read-only; allow writes only below operation-private Go cache, temp, telemetry, and output roots; enforce the empty fixed environment and network/module/VCS denial on every platform; and apply supported host deadlines plus bounded output, artifact/disk, memory, and process controls without widening the executable graph. Parse the complete fixed go list stream; require one non-DepOnly main package and validate every standard input below fingerprinted GOROOT and every non-standard module, source, embed, and dependency input below build_root. Reject errors, incomplete packages, cgo and native fields, every SysoFiles entry, non-standard SFiles, and the exact ASCII //go:cgo_import_dynamic token in active non-standard GoFiles. Build only with the normative internal-link, no-libgcc argv into private staging. Verify one bounded regular output, apply manager permissions, hash it, and expose artifact metadata. Do not publish caches, write markers, or execute the artifact.
### Acceptance Criteria

Valid standard-library and correctly vendored transitive fixtures build with the exact environment and argv from the read-only frozen snapshot and produce verified artifact metadata; only operation-private roots are writable, network/module/VCS access is denied, and each supported platform adapter demonstrably applies its available deadline, output, artifact/disk, memory, and process controls without permitting an extra executable. Non-main or multiple root packages, nested or escaped modules, missing or inconsistent vendor data, workspace or toolchain switching, cgo, C or SWIG inputs, root or transitive syso, non-standard assembly, escaped embed inputs, go:cgo_import_dynamic, generator and PGO attempts, poisoned variables, source writes, external-link or libgcc fallback, unexpected child executables, source or toolchain mutation, resource-limit or oversized output, links, and nonzero exits fail before publication. No package-selected executable and no built output is ever started. Mock executor and host-policy interaction tests, readonly-source and network-denial assertions, and bounded real-toolchain integration tests pass on the supported CI platforms.

## Instructions

No specific instructions attached. Work according to the task description and acceptance criteria above.

## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# The FIRST/LAST sections above define your role-default lifecycle commands.
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(TASK-260720-1zntv0, status=analysis)'       # analyst-style work
task-board m 'set_status(TASK-260720-1zntv0, status=development)'    # implementation / testing work
task-board m 'set_status(TASK-260720-1zntv0, status=reviewing)'      # reviewer handoff
task-board m 'set_status(TASK-260720-1zntv0, status=blocked)'        # when blocked
task-board m 'set_status(TASK-260720-1zntv0, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(TASK-260720-1zntv0, item=1)'                        # check item N
task-board m 'add_checklist_item(TASK-260720-1zntv0, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(TASK-260720-1zntv0, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(TASK-260720-1zntv0, name=TASK-260720-1zntv0_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add TASK-260720-1zntv0 ./path/to/file --type outcome --name TASK-260720-1zntv0_artifact.bin -d "Description"
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
task-board m 'add_resource(TASK-260720-1zntv0, name=TASK-260720-1zntv0_results.md, content="...", type=outcome, description="Description")'
task-board resource add TASK-260720-1zntv0 ./path/to/file --type outcome --name TASK-260720-1zntv0_artifact.bin -d "Description"
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
Snapshot digest: sha256:f45a541f32cf70e5abb19aaf2858e8016d190c79dc312a121f66c44f06c384a4

### checklist/definition_of_done

Digest: sha256:6942fd688da6238903bea69eed200bf7bec181e9943712a6441b4eb114d99c74
Length: 1634 bytes
Body:
["Complete package graph and source inputs pass all rejection clusters","Only the fixed go list and go build commands can execute","Built outputs are verified but never launched during installation","Readonly source, network denial, and supported host resource controls are executable platform gates","Code written per task description and AC","Relevant tests written for new or changed behavior and passing","Lint clean","Relevant build/validation commands run after changes and build not broken","New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables","Important findings, decisions, anomalies, or regressions recorded in logbook when relevant","Board size is proportional to the spec and is the smallest decomposition that maps every requirement","Every story and task traces to a concrete spec requirement; justified-gap elements also carry a self-verified gap record","Beyond-literal-spec elements include a written justification naming the gap and the spec and out-of-scope checks performed before creation","Research tasks cite an exact question the spec genuinely leaves open","Dependencies linked","Tasks are atomic — one clear deliverable each","Completeness verified — nothing forgotten","Any planning artifacts actually produced are linked as new task-scoped outcome resources; diagrams are strictly optional, never a standing deliverable","Implementation matches AC","Solution fits project architecture","Tests green","If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches"]

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

Digest: sha256:0799aeeb3183ef446e385d513ab92646249b84f6f1c542052526e84d70d8a3fc
Length: 1173 bytes
Body:
Valid standard-library and correctly vendored transitive fixtures build with the exact environment and argv from the read-only frozen snapshot and produce verified artifact metadata; only operation-private roots are writable, network/module/VCS access is denied, and each supported platform adapter demonstrably applies its available deadline, output, artifact/disk, memory, and process controls without permitting an extra executable. Non-main or multiple root packages, nested or escaped modules, missing or inconsistent vendor data, workspace or toolchain switching, cgo, C or SWIG inputs, root or transitive syso, non-standard assembly, escaped embed inputs, go:cgo_import_dynamic, generator and PGO attempts, poisoned variables, source writes, external-link or libgcc fallback, unexpected child executables, source or toolchain mutation, resource-limit or oversized output, links, and nonzero exits fail before publication. No package-selected executable and no built output is ever started. Mock executor and host-policy interaction tests, readonly-source and network-denial assertions, and bounded real-toolchain integration tests pass on the supported CI platforms.

### task/description

Digest: sha256:d8f53482735adbe2b1f6b774a9aed1f7c899074bc5dcf3c18db919c7868fb4c2
Length: 239 bytes
Body:
Implement the source-aware go-v1 pipeline that validates the complete active package graph with one fixed go list command, builds with one fixed go build command, verifies the staged executable, and never transfers control to package code.

### task/scope

Digest: sha256:3eee55456796554bd643486c5b23e50e453280661c935813fbe8fef00b0e1132
Length: 1203 bytes
Body:
Continue in internal/godriver using the trusted session and frozen build source. Set CWD to canonical source_dir and expose a build-tagged host-execution policy boundary. Present the validated snapshot and vendored module tree read-only; allow writes only below operation-private Go cache, temp, telemetry, and output roots; enforce the empty fixed environment and network/module/VCS denial on every platform; and apply supported host deadlines plus bounded output, artifact/disk, memory, and process controls without widening the executable graph. Parse the complete fixed go list stream; require one non-DepOnly main package and validate every standard input below fingerprinted GOROOT and every non-standard module, source, embed, and dependency input below build_root. Reject errors, incomplete packages, cgo and native fields, every SysoFiles entry, non-standard SFiles, and the exact ASCII //go:cgo_import_dynamic token in active non-standard GoFiles. Build only with the normative internal-link, no-libgcc argv into private staging. Verify one bounded regular output, apply manager permissions, hash it, and expose artifact metadata. Do not publish caches, write markers, or execute the artifact.

### task/title

Digest: sha256:27e6dedae5dba1fd239181af6f5a11ee3100fb8e7b2126c9326291ade51420ad
Length: 44 bytes
Body:
Implement go-v1 preflight and artifact build

### validation_manifest/current

Digest: sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945
Length: 2 bytes
Body:
[]

