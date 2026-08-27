# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(BUG-260825-11nmd5, status=reviewing)'
```

## Operational Constraints (Headless Run)

Read this before you run any gate or attach any outcome. A spawned run is
usually headless: the session ends when your turn ends.

- **Never background a long command and end your turn.** The session terminates
  with the turn and the process dies with it, while the run is still recorded as
  completed — a successful-looking run with no evidence attached.
- **A single shell call is time-bounded (~10 minutes).** Waiting out a longer
  check inside one call fails too. Split long verification into bounded
  sequential calls (package subsets, `-run` masks) and state explicitly what you
  reran yourself versus accepted from already-attached evidence.
- **Attach evidence before you end the lifecycle, not after.** "I will attach it
  when the run finishes" is unfulfillable by construction: once your turn ends
  there is no session left to attach anything.

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
- When your prompt carries a `Change Request Under Review` section, record acceptance with `task-board m 'accept_cr(ID, revision=N, evidence=ID_review-verdict.md)'`. It parks the element at `to-review` as the accepted handoff — accepted work never stays in `reviewing` — and the orchestrator then makes the `done` transition with `commit_ack=scope_committed`. `accept_cr` takes no `commit_ack`, and only the reviewer run that was handed the revision may call it.

## Skills

These role skill references are a lazy catalog, not a mandate to bulk-read every
body. Before technical work, identify the skills relevant to this task's concrete
scope and read those full skill bodies. Always read any skill explicitly required
by the task, user, or project instructions.

- **project-management**: `/Users/iv/.claude/skills/project-management/SKILL.md`
- **architecture-diagrams**: `/Users/iv/.claude/skills/architecture-diagrams/SKILL.md`

## Definition of Done

- [ ] Vendored generator directives are inert at build time and the carve-out is bounded to materialized vendor trees per the decision 0005 relaxation family
- [ ] Regression test pins acceptance of a vendored //go:generate and continued rejection of a first-party one
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

- **ID**: BUG-260825-11nmd5
- **Title**: BUG-260825-11nmd5: vendored-go-generate-rejected-against-profile
- **Parent**: STORY-260822-2lvw0e
### Description

go-v1 rejects any active Go file containing //go:generate, including audited third-party vendored code. curator-spec profiles/manager.md is normative that the directive is inert and that its presence in vendored GoFiles does not fail preflight, and decision 0005-vendored-go-boundary-relaxation names clipperhouse/displaywidth and skill-project-management as the motivating case. Commit 9ba552f Restore Go generator rejection dropped the vendor carve-out from internal/godriver/graph.go. Discovered while adopting schema-8 module roots in skill-project-management (its board TASK-260822-hje0ya): task-board-tui cannot build because bubbletea to charmbracelet/x/ansi to clipperhouse/displaywidth ships a bare //go:generate in gen.go, and no released version drops it.
### Acceptance Criteria

validatePackageGraph exempts //go:generate only for packages below the build root vendor tree; build-root and first-party packages are still rejected with go_generator_forbidden; the cgo_import_dynamic allowlist is unchanged; conformance suite against SPEC_PIN stays green.

## Instructions

No specific instructions attached. Work according to the task description and acceptance criteria above.

## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# FIRST/LAST lifecycle commands:
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(BUG-260825-11nmd5, status=analysis)'       # analyst-style work
task-board m 'set_status(BUG-260825-11nmd5, status=development)'    # implementation / testing work
task-board m 'set_status(BUG-260825-11nmd5, status=reviewing)'      # reviewer handoff
task-board m 'set_status(BUG-260825-11nmd5, status=blocked)'        # when blocked
task-board m 'set_status(BUG-260825-11nmd5, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(BUG-260825-11nmd5, item=1)'                        # check item N
task-board m 'add_checklist_item(BUG-260825-11nmd5, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(BUG-260825-11nmd5, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(BUG-260825-11nmd5, name=BUG-260825-11nmd5_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add BUG-260825-11nmd5 ./path/to/file --type outcome --name BUG-260825-11nmd5_artifact.bin -d "Description"
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
task-board m 'add_resource(BUG-260825-11nmd5, name=BUG-260825-11nmd5_results.md, content="...", type=outcome, description="Description")'
task-board resource add BUG-260825-11nmd5 ./path/to/file --type outcome --name BUG-260825-11nmd5_artifact.bin -d "Description"
```

If you revise the same artifact later, use `task-board m 'update_resource(...)'` or `task-board resource update ...` instead of creating a silent overwrite.

If you discover important findings, decisions, anomalies, regressions, or non-obvious constraints while working, record them in `logbook` as well as on the board.

This ensures your results persist on the board and are accessible to other agents and the coordinator. Spawn completion is expected to produce at least one new task-scoped outcome artifact before the task can cleanly remain in `to-review`.

## Evidence That Counts

A passing suite means nothing unless something in it would have failed.

- Any behavior that GATES, REFUSES, VALIDATES, AUTHORIZES, or ATTESTS ships with negative tests that fail when the gate admits what it must reject. A positive test proves the gate is reachable, not that it works.
- Prove a bound by NARROWING the gate, not only by deleting it. A delete-only mutant proves the gate exists and says nothing about the class it covers.
- Prove behavior by driving the real entry point — launch, materialize, resolve, publish — and name the production call site. A helper that is unit-tested but called from nowhere promises nothing.
- Standard negative shapes to test and to look for: forged or self-minted evidence; absent evidence treated as satisfied; the check present but uncalled from production; a bypass path around the check; a capability claim that does not reproduce.
- An absence and a failure to read are different facts. A failed, partial, or malformed read is never a legitimate absence, and a fallback defined for absence must not fire on a read failure.
- Prove, or report nothing. Where a property cannot be established, report unknown instead of inferring it from a proxy signal; callers act on a plausible guess.

Shapes vocabulary and the incident record behind each rule: `references/negative-evidence.md` in the project-management skill.

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

## Story Workspace

You are running in an isolated Git worktree for STORY-260822-2lvw0e.

- Workspace path: `.temp/STORY-260822-2lvw0e/worktree` (this is your working directory)
- Branch: `task-board/story/STORY-260822-2lvw0e`, forked from `main`
- Authoritative board: `/Users/iv/Developer/ReluxWorks/curator/.task-board` (already exported as `TASK_BOARD_DIR`)

Make every repository change here. Do not switch, rebase, merge, or delete this branch, and do not run `task-board` against the `.task-board` copy inside this worktree — it is a checkout artifact, not the board. Integration into trunk is the orchestrator's step, not yours.

## Change Request Under Review

- Change Request: `CR-BUG-260825-11nmd5-1` revision `1` (state `ready`)
- Element: `BUG-260825-11nmd5`, integration scope: `STORY-260822-2lvw0e`
- Branch: `task-board/story/STORY-260822-2lvw0e`
- Base OID: `903af23ad0d0fa21328c0a2100e17968bbac6f1e`
- Candidate tree OID: `93b7702087ea83985166633bda09b723ffde0504`
- Repository delta: `present`
- Patch resource: `BUG-260825-11nmd5_change-request_rev1.patch` (sha256 `7769d872fa2e1f9e57d930c288ee3e383ec679222b93d7a12588227ee3dd72e9`)

```bash
git -C . diff 903af23ad0d0fa21328c0a2100e17968bbac6f1e 93b7702087ea83985166633bda09b723ffde0504          # the exact reviewable delta
git -C . diff --stat 903af23ad0d0fa21328c0a2100e17968bbac6f1e 93b7702087ea83985166633bda09b723ffde0504   # its shape
```

Changed paths (4):

- `LOGBOOK.md`
- `internal/godriver/build_test.go`
- `internal/godriver/graph.go`
- `internal/godriver/graph_test.go`

### Recording the verdict

Acceptance is a named mutation, not a status change:

```bash
task-board m 'accept_cr(BUG-260825-11nmd5, revision=1, evidence=BUG-260825-11nmd5_review-verdict.md)'
```

`accept_cr` sets the Change Request to `accepted`, stamps this run as the reviewer, and parks the element at `to-review` (BUG-260825-11nmd5) — which is the accepted handoff, not `done`. The orchestrator checkpoints or integrates the accepted revision and makes the `done` transition with `commit_ack`. Never supply `commit_ack` yourself.

Changes requested and Stop-The-Line are unchanged: route to `to-dev`/`analysis` or evidence-backed `blocked`. Leaving the element in `reviewing` is still not a verdict and still gets this run cancelled and respawned.
