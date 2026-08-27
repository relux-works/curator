# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260720-2g21eg, status=reviewing)'
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

- [ ] Record the base SHA and create the task worktree only after the clean local main clone is fast-forwarded and dependency handoffs are present.
- [ ] Test the exact source-aware argv, graph rejection surface, fixed environment, output verification, and never-run invariant.
- [ ] Run focused pytest, the real fixture build, python -m mypy, and attach task-scoped evidence.
- [ ] Code written per task description and AC
- [ ] Relevant tests written for new or changed behavior and passing
- [ ] Lint clean
- [ ] Relevant build/validation commands run after changes and build not broken
- [ ] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [ ] Tests written and passing
- [ ] Coverage target ~80%+ for affected code
- [ ] New task-scoped outcome artifact attached on the board for reports, logs, screenshots, or other produced evidence
- [ ] Implementation matches AC
- [ ] Solution fits project architecture
- [ ] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches
## Your Task

- **ID**: TASK-260720-2g21eg
- **Title**: Implement fixed go-v1 compile driver
- **Parent**: STORY-260720-1uv5gi
### Description

Implement the source-aware go-v1 preflight and compile engine under the rc.5 portable manager-worker-v1 execution policy, with a hidden identity-verified manager-owned worker, manager-owned argv, environment, dependency-graph validation, native-control preflight, capability evidence, and output verification, without cache or install concerns.
### Scope

Own src/csk/builds/go_v1.py, its worker re-execution and injected process executor boundary, the hidden worker entry point, the worker session protocol, and focused unit and fixture tests. Starting from the frozen snapshot and trusted toolchain descriptor, run the fixed four-node process graph of protocol core.md 4.2.1: manager parent, identity-verified manager-owned worker, fingerprinted GOROOT/bin/go, fingerprinted regular executables below GOROOT/pkg/tool/GOHOSTOS_GOHOSTARCH. The worker is an exact re-execution of the installed manager executable in one fixed hidden mode; no package file, manifest value, descriptor value, environment value, PATH lookup, shell, or user option may select it. Split the five fixed argv forms by process: the three package-independent probe forms telemetry-off, version and env stay on the manager parent and are owned by TASK-260720-3j8pp5, while the two source-aware forms list and build are issued only by the worker inside the manager-built clean environment. Drive the 13 session states from vectors/go-host-execution-policy.json session_states, from parent-package-independent-toolchain-probe through worker-domain-teardown. Parse the complete go list JSON stream in the parent, enforce module and path containment for main and transitive dependencies before issuing the authenticated build permit, and validate the staged executable. Emit exactly one capability-evidence-v1 record per operation as a result-only value. Implement macOS and Windows and fail closed on every other host; do not add a Linux control path. Do not publish caches, write markers or shims, launch the output, or claim any deferred hardened guarantee.
### Acceptance Criteria

The only source-aware argv are the normative vendor-only go list and go build forms with trimpath, buildvcs disabled, compiler gc, pgo off, internal linking, libgcc none, and a manager-derived output path, and both are issued only by the manager-owned worker, exactly one of each per session, never by the manager parent; the three package-independent probe forms remain parent-side and belong to TASK-260720-3j8pp5. One worker session performs exactly one go list, waits for parent validation of the complete package graph, accepts exactly one authenticated build permit bound to a fresh session nonce, and performs exactly one go build; any retry, second list, second build, replayed nonce, out-of-order, oversize, or unknown message tears the session down without starting a compiler. The environment starts empty and fixes private Go caches, GOPROXY and VCS off, GOWORK off, GOTOOLCHAIN local, CGO disabled, GO_EXTLINK_ENABLED 0, native target and tuning, locale, temp roots, and empty executable PATH, matching vectors/build-drivers.json fixed_environment. All 18 mandatory_controls from vectors/go-host-execution-policy.json are enforced always, including identity-verified-manager-owned-worker, pre-launch-worker-identity-verification, post-exec-identity-reverification, frozen-source-snapshot-integrity, fixed-manager-selected-process-graph, closed-standard-input-and-descriptors, worker-domain-teardown, no-artifact-execution, inventory-native-controls-applied and closed-capability-evidence-record. Native-control availability is probed once per operation before worker launch against rc5-native-control-inventory-v1, which is exhaustive over exactly macos and windows with exactly the five controls descendant-domain-termination, active-process-count-limit, aggregate-memory-limit, per-file-size-limit and inherited-handle-restriction; every control the inventory marks available for the host is applied, nothing outside the inventory is applied or reported, and a cached, inherited, configured, or host-label result is not a probe. Exactly one capability-evidence-v1 record per operation carries exactly record_version, execution_policy, platform and controls, with exactly one entry per inventory control and exactly the fields name, availability, status and probed_at; it is exposed in dry-run-plan-result, install-result and status-result and excluded from cache-key, receipt, install-marker and conformance-claim. All eight capability_evidence_record consistency_rules are enforced with their declared errors: six reject with build_execution_capability_evidence_invalid and the two hardened rules reject with build_execution_hardened_claim_forbidden. Failure boundary matches the vector exactly: a missing mandatory portable control fails with build_execution_control_unavailable before worker-launch, rejects the build and publishes nothing, while an unavailable inventory native control and a missing deferred hardened capability neither reject the build nor block publication. All six deferred_hardened_guarantees are refused by their six deferred_capability_rejection_guards and are never claimed. All 14 identity_and_protocol_cases and all 8 package_influence_cases from the vector are covered as named negatives, and the 11 capability_evidence_cases are covered as named cases. Exactly one non-DepOnly main package is accepted. Missing or inconsistent vendor data, workspace or toolchain switching, graph paths outside build root or GOROOT, load errors, cgo or native fields, any SysoFiles, nonstandard SFiles, escaped embeds, and active nonstandard cgo_import_dynamic directives fail before the build permit. Output is one bounded regular executable in staging, is hashed and permissioned, and is never run. On any host outside macOS and Windows the source-aware path fails closed with an unavailable-control error and never starts a worker. Focused pytest, a valid real Go fixture build on a native macOS or Windows host, poisoned-environment negatives, and strict mypy pass.

## Instructions

No specific instructions attached. Work according to the task description and acceptance criteria above.

## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# The FIRST/LAST sections above define your role-default lifecycle commands.
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(TASK-260720-2g21eg, status=analysis)'       # analyst-style work
task-board m 'set_status(TASK-260720-2g21eg, status=development)'    # implementation / testing work
task-board m 'set_status(TASK-260720-2g21eg, status=reviewing)'      # reviewer handoff
task-board m 'set_status(TASK-260720-2g21eg, status=blocked)'        # when blocked
task-board m 'set_status(TASK-260720-2g21eg, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(TASK-260720-2g21eg, item=1)'                        # check item N
task-board m 'add_checklist_item(TASK-260720-2g21eg, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(TASK-260720-2g21eg, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(TASK-260720-2g21eg, name=TASK-260720-2g21eg_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add TASK-260720-2g21eg ./path/to/file --type outcome --name TASK-260720-2g21eg_artifact.bin -d "Description"
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
task-board m 'add_resource(TASK-260720-2g21eg, name=TASK-260720-2g21eg_results.md, content="...", type=outcome, description="Description")'
task-board resource add TASK-260720-2g21eg ./path/to/file --type outcome --name TASK-260720-2g21eg_artifact.bin -d "Description"
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
Snapshot digest: sha256:2151664891f75b4da043154c4f94c1e63514f8799fa0e69bd474448fdf0f79da

### checklist/definition_of_done

Digest: sha256:5c35199cc2225c27f9367ed2173aced07152be9bf29d87e7acde7bceb2611701
Length: 1176 bytes
Body:
["Record the base SHA and create the task worktree only after the clean local main clone is fast-forwarded and dependency handoffs are present.","Test the exact source-aware argv, graph rejection surface, fixed environment, output verification, and never-run invariant.","Run focused pytest, the real fixture build, python -m mypy, and attach task-scoped evidence.","Code written per task description and AC","Relevant tests written for new or changed behavior and passing","Lint clean","Relevant build/validation commands run after changes and build not broken","New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables","Important findings, decisions, anomalies, or regressions recorded in logbook when relevant","Tests written and passing","Coverage target ~80%+ for affected code","New task-scoped outcome artifact attached on the board for reports, logs, screenshots, or other produced evidence","Implementation matches AC","Solution fits project architecture","Tests green","If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches"]

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

Digest: sha256:6a42f7e2f15f3723792ce9489a6927cc3857241d611af3845e5bb9e62593b992
Length: 3999 bytes
Body:
The only source-aware argv are the normative vendor-only go list and go build forms with trimpath, buildvcs disabled, compiler gc, pgo off, internal linking, libgcc none, and a manager-derived output path, and both are issued only by the manager-owned worker, exactly one of each per session, never by the manager parent; the three package-independent probe forms remain parent-side and belong to TASK-260720-3j8pp5. One worker session performs exactly one go list, waits for parent validation of the complete package graph, accepts exactly one authenticated build permit bound to a fresh session nonce, and performs exactly one go build; any retry, second list, second build, replayed nonce, out-of-order, oversize, or unknown message tears the session down without starting a compiler. The environment starts empty and fixes private Go caches, GOPROXY and VCS off, GOWORK off, GOTOOLCHAIN local, CGO disabled, GO_EXTLINK_ENABLED 0, native target and tuning, locale, temp roots, and empty executable PATH, matching vectors/build-drivers.json fixed_environment. All 18 mandatory_controls from vectors/go-host-execution-policy.json are enforced always, including identity-verified-manager-owned-worker, pre-launch-worker-identity-verification, post-exec-identity-reverification, frozen-source-snapshot-integrity, fixed-manager-selected-process-graph, closed-standard-input-and-descriptors, worker-domain-teardown, no-artifact-execution, inventory-native-controls-applied and closed-capability-evidence-record. Native-control availability is probed once per operation before worker launch against rc5-native-control-inventory-v1, which is exhaustive over exactly macos and windows with exactly the five controls descendant-domain-termination, active-process-count-limit, aggregate-memory-limit, per-file-size-limit and inherited-handle-restriction; every control the inventory marks available for the host is applied, nothing outside the inventory is applied or reported, and a cached, inherited, configured, or host-label result is not a probe. Exactly one capability-evidence-v1 record per operation carries exactly record_version, execution_policy, platform and controls, with exactly one entry per inventory control and exactly the fields name, availability, status and probed_at; it is exposed in dry-run-plan-result, install-result and status-result and excluded from cache-key, receipt, install-marker and conformance-claim. All eight capability_evidence_record consistency_rules are enforced with their declared errors: six reject with build_execution_capability_evidence_invalid and the two hardened rules reject with build_execution_hardened_claim_forbidden. Failure boundary matches the vector exactly: a missing mandatory portable control fails with build_execution_control_unavailable before worker-launch, rejects the build and publishes nothing, while an unavailable inventory native control and a missing deferred hardened capability neither reject the build nor block publication. All six deferred_hardened_guarantees are refused by their six deferred_capability_rejection_guards and are never claimed. All 14 identity_and_protocol_cases and all 8 package_influence_cases from the vector are covered as named negatives, and the 11 capability_evidence_cases are covered as named cases. Exactly one non-DepOnly main package is accepted. Missing or inconsistent vendor data, workspace or toolchain switching, graph paths outside build root or GOROOT, load errors, cgo or native fields, any SysoFiles, nonstandard SFiles, escaped embeds, and active nonstandard cgo_import_dynamic directives fail before the build permit. Output is one bounded regular executable in staging, is hashed and permissioned, and is never run. On any host outside macOS and Windows the source-aware path fails closed with an unavailable-control error and never starts a worker. Focused pytest, a valid real Go fixture build on a native macOS or Windows host, poisoned-environment negatives, and strict mypy pass.

### task/description

Digest: sha256:161866cdea23672327b99b204939a026b9ea69bcff4e09bd2a3b57715e0742c9
Length: 344 bytes
Body:
Implement the source-aware go-v1 preflight and compile engine under the rc.5 portable manager-worker-v1 execution policy, with a hidden identity-verified manager-owned worker, manager-owned argv, environment, dependency-graph validation, native-control preflight, capability evidence, and output verification, without cache or install concerns.

### task/scope

Digest: sha256:acb681e3bf91dd0de35ef58a98f27eac7f5270502cfb848d15b828becdfe6e9f
Length: 1672 bytes
Body:
Own src/csk/builds/go_v1.py, its worker re-execution and injected process executor boundary, the hidden worker entry point, the worker session protocol, and focused unit and fixture tests. Starting from the frozen snapshot and trusted toolchain descriptor, run the fixed four-node process graph of protocol core.md 4.2.1: manager parent, identity-verified manager-owned worker, fingerprinted GOROOT/bin/go, fingerprinted regular executables below GOROOT/pkg/tool/GOHOSTOS_GOHOSTARCH. The worker is an exact re-execution of the installed manager executable in one fixed hidden mode; no package file, manifest value, descriptor value, environment value, PATH lookup, shell, or user option may select it. Split the five fixed argv forms by process: the three package-independent probe forms telemetry-off, version and env stay on the manager parent and are owned by TASK-260720-3j8pp5, while the two source-aware forms list and build are issued only by the worker inside the manager-built clean environment. Drive the 13 session states from vectors/go-host-execution-policy.json session_states, from parent-package-independent-toolchain-probe through worker-domain-teardown. Parse the complete go list JSON stream in the parent, enforce module and path containment for main and transitive dependencies before issuing the authenticated build permit, and validate the staged executable. Emit exactly one capability-evidence-v1 record per operation as a result-only value. Implement macOS and Windows and fail closed on every other host; do not add a Linux control path. Do not publish caches, write markers or shims, launch the output, or claim any deferred hardened guarantee.

### task/title

Digest: sha256:21d0882a312260f5f3b166c280ff9cc38b2eade05011da3f41cc875d892afdae
Length: 36 bytes
Body:
Implement fixed go-v1 compile driver

### validation_manifest/current

Digest: sha256:4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945
Length: 2 bytes
Body:
[]

