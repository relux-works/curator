# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260720-1nvomm, status=reviewing)'
```

## Your Role
# reviewer

## Description

Reviews how a task was implemented and how the solution fits into the project. Does not modify code; records one of the explicit verdict branches below.

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

- [ ] Map every accepted protocol decision and security boundary to normative prose or decision 0004
- [ ] Prove build roots are excluded from agent context and runtime copying on real, cache-hit, and dry-run paths
- [ ] Run the scoped protocol validation command and record the result
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

- **ID**: TASK-260720-1nvomm
- **Title**: Specify protocol v6 core and security contract
- **Parent**: STORY-260720-35dck7
### Description

Update the normative protocol model from the accepted TASK-260720-poa3ze contract. Define manifest schema 6 as a declarative compiled-artifact extension with build_roots and a closed go-v1 command. Build sources must be statically excluded from agent context and runtime copying on real installs, cache hits, and dry-runs. Record the no-hooks threat boundary, build-source identity, logical cache and receipt semantics, marker v2 compatibility, and the Go-only v1 decision. Preserve every schema 1 through 5 behavior.
### Scope

Work only in /Users/iv/Developer/ReluxWorks/curator-spec from origin/main commit 57c1f56846d221ecc55786bd3c2467ec32f11730. Own protocol/core.md, SECURITY.md, and new decisions/0004-compile-only-build-drivers.md. Use TASK-260720-poa3ze_compile-only-build-drivers.md at accepted SHA-256 6308d99d8bdad4445841bc1cbd230cadbed0020012d0e9d38d877b413348f681 as the contract. Do not edit schemas, vectors, generator, manager profile, CLI guide, or release metadata in this task.
### Acceptance Criteria

protocol/core.md normatively defines schema 6, build_roots, the strict build command, build-source exclusion, manager-derived artifact paths, curator-build-source-v1, logical cache and receipt identity, marker v2 implications, and schema 1 through 5 compatibility; SECURITY.md explicitly forbids package shell, argv, environment, output selection, hooks, plugins, generators, unsafe build systems, external-link fallback, and executing built output while documenting compiler-input and protected-cache trust boundaries; decision 0004 records the closed Go-only v1 choice, rejected alternatives, non-normative physical cache layout, and future-driver review rule; terminology matches the accepted contract without weakening any MUST or MUST NOT; python3 tools/validate.py passes.

## Instructions

The following instructions have been attached to this task:

### TASK-260720-1nvomm_artifact-map.puml
> Implementation planning diagram showing normative contract ownership and downstream consumers

@startuml
!theme plain
!pragma layout smetana

title Protocol schema v6 — task and artifact ownership
top to bottom direction
skinparam componentStyle rectangle

package "Normative contract" #LightBlue {
  [TASK-260720-1nvomm\nCore, security, ADR] as Core
  [TASK-260720-17llva\nManager profile] as Manager
}

package "Wire schemas" #LightGreen {
  [TASK-260720-wajgn8\nManifest v6 pair] as Manifest
  [TASK-260720-12iigs\nReceipt v1 + marker v2] as Artifacts
  [TASK-260720-2zc6k1\nClaim v2 + rc.4 suite identity] as Claim
  [TASK-260720-37ei85\nLegacy compatibility guards] as Legacy
}

package "Generated conformance" #LightYellow {
  [TASK-260720-1s1vr6\nBuild-driver fixture + vectors] as BuildVectors
  [TASK-260720-cw39jh\nManager lifecycle vectors] as LifecycleVectors
}

package "Publication gates" #LightPink {
  [TASK-260720-1u7hes\nValidation + release gates] as Gates
  [TASK-260720-3lo9jc\nSchema, conformance, CLI docs] as Docs
  [TASK-260720-q5oy3o\nrc.4 release metadata] as Release
  [TASK-260720-3ag6pi\nIntegrated verification] as Verify
}

Core -down-> Manager : operational semantics
Core -down-> Manifest : structural surface
Manager -down-> Artifacts : lifecycle metadata
Manifest -down-> Artifacts : shared definitions
Artifacts -down-> Claim : release identity
Claim -down-> Legacy : claim transition
Legacy -down-> BuildVectors : compatibility baseline
BuildVectors -down-> LifecycleVectors : logical identities
LifecycleVectors -down-> Gates : executable evidence
LifecycleVectors -down-> Docs : stable docs surface
Gates -down-> Release
Docs -down-> Release
Release -down-> Verify

@enduml



### TASK-260720-1nvomm_accepted-contract.md
> Accepted compile-only driver contract, verified at SHA-256 6308d99d8bdad4445841bc1cbd230cadbed0020012d0e9d38d877b413348f681

Attached file materialized at: `/Users/iv/Developer/ReluxWorks/curator/.temp/resources/TASK-260720-1nvomm/TASK-260720-1nvomm_accepted-contract.md`

Use this local file directly when you need to inspect the attachment.




## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# The FIRST/LAST sections above define your role-default lifecycle commands.
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(TASK-260720-1nvomm, status=analysis)'       # analyst-style work
task-board m 'set_status(TASK-260720-1nvomm, status=development)'    # implementation / testing work
task-board m 'set_status(TASK-260720-1nvomm, status=reviewing)'      # reviewer handoff
task-board m 'set_status(TASK-260720-1nvomm, status=blocked)'        # when blocked
task-board m 'set_status(TASK-260720-1nvomm, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(TASK-260720-1nvomm, item=1)'                        # check item N
task-board m 'add_checklist_item(TASK-260720-1nvomm, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(TASK-260720-1nvomm, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(TASK-260720-1nvomm, name=TASK-260720-1nvomm_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add TASK-260720-1nvomm ./path/to/file --type outcome --name TASK-260720-1nvomm_artifact.bin -d "Description"
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
task-board m 'add_resource(TASK-260720-1nvomm, name=TASK-260720-1nvomm_results.md, content="...", type=outcome, description="Description")'
task-board resource add TASK-260720-1nvomm ./path/to/file --type outcome --name TASK-260720-1nvomm_artifact.bin -d "Description"
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
