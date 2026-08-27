# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(BUG-260731-3gm8kc, status=reviewing)'
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

- [ ] Replace brittle lifecycle vector length checking with semantic required-case validation.
- [ ] Add focused Curator regression coverage for the rc.6 compiled-cache-miss dry-run case.
- [ ] Publish a signed Curator commit and open a PR targeting main.
- [ ] Advance curator-spec implementations.yml to the exact published Curator commit.
- [ ] Attach evidence, hand off to independent Opus review, and require Curator plus PR 15 CI green.
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

- **ID**: BUG-260731-3gm8kc
- **Title**: BUG-260731-3gm8kc: curator-interop-lifecycle-vector-gate
- **Parent**: STORY-260720-35dck7
### Description

curator-spec Implementations (ubuntu/macos/windows) fails on PR 15 and identically on PR 14 at its first step, Go manager shared suite, in TestManagerLifecycleVectors. Curator internal/interop/golden_test.go:487 hard-codes: if len(vector.LauncherCases) != 2 || len(vector.BootstrapCases) != 3 || len(vector.UpgradeCases) != 3 || len(vector.DryRunCases) != 2. The rc.6 base commit 671888e expanded conformance/v1/vectors/manager-lifecycle.json dry_run_cases from 2 to 3 by adding compiled-cache-miss-is-read-only; curator-spec main still has 2, which is why the gate is green there. Verified during BUG-260731-2rhy74 RUN-260731-b4fd97: advancing the implementations.yml Curator pin from 17804ce (v0.12.5) to released main/v0.13.0 cfffd7cd does NOT fix it. Tested locally in a fresh submodule-initialized worktree at cfffd7cd against the PR 15 conformance root: go test -count=1 ./internal/interop ./internal/closure ./internal/skillspec exits 1 with the same manager lifecycle vector is incomplete message; closure and skillspec pass. Of all pushed Curator branches only main cfffd7cd and agent/seamless-manager-lifecycle a78545cd carry TestManagerLifecycleVectors and both assert != 2, so no pinnable commit exists. Not introduced by the marker-v2 work: git diff --name-only 671888e HEAD on task/BUG-260731-2rhy74-marker-v2-fixture touches no vector file, and PR 14 fails the same three jobs without that commit. PR 15 Specification, Formatting and Links checks all pass.
### Scope

Curator interop lifecycle-vector consumer plus the immutable Curator pin in curator-spec implementation conformance. Publish Curator and spec changes through their existing task branches/PRs; do not tag or release.
### Acceptance Criteria

Curator internal/interop golden gate accepts the rc.6 manager-lifecycle vector including the third dry-run case compiled-cache-miss-is-read-only, ideally by asserting required case names rather than list lengths; the change is published on a Curator commit that curator-spec can pin; .github/workflows/implementations.yml advances the Curator pin to that commit; Implementations passes on ubuntu, macOS and Windows for both PR 15 and PR 14.

## Instructions

The following instructions have been attached to this task:

### BUG-260731-3gm8kc_orchestrator-review-context.md
> Orchestrator review-routing context: prior reviewer runs died at launch, checklist items 5/8 decision, sibling CI bugs in flight

# BUG-260731-3gm8kc — orchestrator review-routing context

## Why you were spawned

The producer handed off at `to-review` with a full evidence packet
(`BUG-260731-3gm8kc_implementation-and-evidence.md`). Two earlier reviewer runs
died at launch with a transport-level `Not logged in` API error and produced no
verdict. That was an environment failure, not a review outcome — there is no
prior verdict to defer to. You are the first real review of this work.

## Isolation (mandatory)

The primary checkout `/Users/iv/Developer/ReluxWorks/curator` is on branch
`agent/link-curator-skill-registry` and is dirty with unrelated board files.
Do not switch its branch or commit in it. If you need to build or run tests,
use a read-only worktree under `.temp/BUG-260731-3gm8kc/review-worktree/`.

## What is under review

- Curator PR 9, branch `task/BUG-260731-3gm8kc-lifecycle-vector-gate`,
  head `bd6ba08acda3dc801512c408c759ac0ac6f79f26`.
  - `fee35c87` — manager-lifecycle vector gated by required case **name** instead
    of list length, plus a focused `TestManagerCompiledCacheMissDryRunVector`
    bound to the `install.BuildOutcome` vocabulary.
  - `bd6ba08` — repairs `.github/ci/toolchain-identity.sh`, which asserted
    `go env GOENV = off` while go 1.25 prints nothing for `GOENV=off`.
- curator-spec pin advanced to that commit on both PR 14
  (`release/v1.0.0-rc.6`) and PR 15 (`task/BUG-260731-2rhy74-marker-v2-fixture`).

## The decision the producer explicitly escalated to review

The producer deliberately left two checklist items unchecked and refused to tick
them to pass the handoff gate. Judge both:

1. **Item 5 — "require Curator plus PR 15 CI green."**
   Curator PR 9 CI is red in four jobs (`Lint`, `Test (ubuntu)`, `Race (ubuntu)`,
   `Test (windows)`). The producer's claim is that all four are **pre-existing on
   `main`**, proven by an isolated control branch
   `ci/goenv-control-BUG-260731-3gm8kc` (= main + only the toolchain-identity
   repair, run 30616027892) reproducing identical signatures.
   Verify that control-branch claim yourself rather than accepting it.

2. **Item 8 — "Lint clean."**
   `golangci-lint v2.12.2` exits 0 on darwin, but the repo gate is the CI `Lint`
   job, which is red on linux for two `unused` findings in `internal/godriver`
   and `internal/transaction` — neither in a file this change touches.

**Orchestrator routing fact you should use:** those pre-existing failures are
already split out as separately owned board work, and both are executing right
now in parallel with this review:
- `BUG-260731-11bpa4` (curator-windows-test-vet-compile-break) — Windows `go vet`
  `undefined: decodeHelperOutput` in `internal/runtimestore`.
- `BUG-260731-lepevi` (curator-main-ci-red-linux-lane) — the two linux `unused`
  findings plus the six `cmd/curator` compiled cases.

So "these failures are out of this bug's scope" is a testable claim, not an
excuse: either the evidence shows they are pre-existing and separately owned, or
it does not.

## Also verify

- The AC is about the **Implementations** gate: the producer claims 8/8 green on
  both PR 14 (head b07ef1d) and PR 15 (head 2629aec). Confirm against real CI.
- Whether gating by required case **name** is genuinely stronger than the length
  check it replaced. The producer supplied four negative controls, including an
  rc.3 rename that the OLD gate accepted. Re-check the teeth; a gate that only
  looks stronger is a fail.
- The producer's own flagged concern: conformance README section 4 wants a full
  immutable commit ID that passed its own required CI. The pin currently names an
  unmerged PR head whose CI is red for the pre-existing reasons above. Decide
  whether that is acceptable now or must be re-pinned after PR 9 lands.
- One out-of-scope finding the producer did not fix and did not hide:
  `internal/install TestAuthoritativeDryRunCasesMutateNothingPersistent` fails on
  an rc.6 root because the new `scope multi-project` case has no executable
  binding. Confirm it is genuinely outside this bug and, if it needs tracking,
  say so in your verdict.

## Verdict contract

Pick exactly one branch and record evidence for it:
- accepted → `done`
- changes requested → route back to `to-dev`, with a concrete, actionable list
- genuine human-only stop-the-line boundary → `blocked` with the full evidence
  packet and the exact decision needed

Do not weaken acceptance to unblock the pipeline, and do not accept on the
producer's summary alone — verify against the repository and real CI.





## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# FIRST/LAST lifecycle commands:
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(BUG-260731-3gm8kc, status=analysis)'       # analyst-style work
task-board m 'set_status(BUG-260731-3gm8kc, status=development)'    # implementation / testing work
task-board m 'set_status(BUG-260731-3gm8kc, status=reviewing)'      # reviewer handoff
task-board m 'set_status(BUG-260731-3gm8kc, status=blocked)'        # when blocked
task-board m 'set_status(BUG-260731-3gm8kc, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(BUG-260731-3gm8kc, item=1)'                        # check item N
task-board m 'add_checklist_item(BUG-260731-3gm8kc, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(BUG-260731-3gm8kc, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(BUG-260731-3gm8kc, name=BUG-260731-3gm8kc_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add BUG-260731-3gm8kc ./path/to/file --type outcome --name BUG-260731-3gm8kc_artifact.bin -d "Description"
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
task-board m 'add_resource(BUG-260731-3gm8kc, name=BUG-260731-3gm8kc_results.md, content="...", type=outcome, description="Description")'
task-board resource add BUG-260731-3gm8kc ./path/to/file --type outcome --name BUG-260731-3gm8kc_artifact.bin -d "Description"
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
