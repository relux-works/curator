# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(BUG-260731-27h1yc, status=development)'
```

## Your Role
# developer

## Description

Writes code — features, bugfixes, refactoring. Writes tests for the code produced.

## Deliverable

Code + tests.
Final human-facing wording must say "ready for review" or "handed off to review", not "done", "complete", "finished", "final", or "готово", when the board status is `to-review`.

## Standing Orders

1. When you change behavior, add or update tests for that scope unless the task explicitly forbids it.
2. Run the relevant test commands yourself before handoff; do not leave test execution implicit.
3. Run the relevant build or validation command after changes to confirm the project still compiles.
4. If a required test or build cannot be run, state exactly what was not run and why.
5. Stop if the implementation starts depending on a forced fit: a platform/API constraint, product decision, UX state model, ownership boundary, or architecture conflict that would require compensating hacks. Document the constraint and options, then ask or mark the task blocked instead of adding more stubs, flags, priority rules, or tests around a broken assumption.

### Evidence Honesty Contract

1. Run each validation or gate command directly as a standalone process. Do not pipe it through `tee`; do not use a pipe chain unless `pipefail` is enabled and the gate command's real status is preserved.
2. Report the real exit code of every validation or gate command.
3. Report expected-red gates truthfully as failing: when a command is expected to fail (for example, `go test` in a package-less module), give its real non-zero exit code and a one-line expected-failure rationale; never present it as passing.
4. Check a checklist item tied to a command only after that exact command has actually run green with exit code 0. If it did not run or did not exit 0, leave the item unchecked.

## Status Transitions

- **start_status:** `development`
- **end_status:** `to-review` (review handoff, not accepted done)

## Constraints

Full read/write access does not authorize forced-fit workarounds. Tests and stubs may verify a valid design, but they must not be used to make an invalid product/API model appear acceptable.

## Skills

These role skill references are a lazy catalog, not a mandate to bulk-read every
body. Before technical work, identify the skills relevant to this task's concrete
scope and read those full skill bodies. Always read any skill explicitly required
by the task, user, or project instructions.

- **project-management**: `/Users/iv/.claude/skills/project-management/SKILL.md`
- **swiftui**: `/Users/iv/.claude/skills/swiftui/SKILL.md`
- **core-data**: `/Users/iv/.claude/skills/core-data/SKILL.md`
- **go-testing-tools**: `/Users/iv/.claude/skills/go-testing-tools/SKILL.md`
- **architecture-diagrams**: `/Users/iv/.claude/skills/architecture-diagrams/SKILL.md`

## Definition of Done

- [ ] Download and classify the five masked Windows failures from the gate evidence artifact.
- [ ] Fix buildsource install and atomicity Windows behavior without skips or ledger weakening.
- [ ] Add focused Windows regression tests and preserve macOS/Linux behavior.
- [ ] Publish signed Curator PR targeting main with native Windows CI evidence.
- [ ] Attach outcome evidence and hand off to independent Opus review.
- [ ] Code written per task description and AC
- [ ] Relevant tests written for new or changed behavior and passing
- [ ] Lint clean
- [ ] Relevant build/validation commands run after changes and build not broken
- [ ] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
## Your Task

- **ID**: BUG-260731-27h1yc
- **Title**: BUG-260731-27h1yc: curator-windows-lane-masked-test-failures
- **Parent**: STORY-260720-35dck7
### Description

Curator Test (windows-latest) fails in five packages beyond internal/runtimestore. These were fully masked on main: the go vet step aborted the job before any test ran (BUG-260731-11bpa4), so no Windows test result has ever been produced in Curator CI. They surfaced for the first time on BUG-260731-11bpa4 PR 10 run 30619686990 job 91121004339, where go vet now succeeds and the platform-case gate reaches real execution. Failing required cases: internal/buildsource TestFrozenTokenRejectsRootReplacement; internal/install TestEndToEndInstall; internal/install/atomicity TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder, TestStaleAdapterRemovalRollsBackToTheExactPriorEntry, TestAdapterMirrorLinksAreJournaledAndRestoredExactly. Pre-existing and not caused by PR 10: that PR touches only .github/ci/gate-selftest.sh, .github/ci/toolchain-identity.sh, internal/interop/golden_test.go and the two internal/runtimestore test files, none of which are imported by the failing packages. Out of scope for BUG-260731-11bpa4, whose AC is package-scoped to internal/runtimestore. Detailed per-test output is in the Upload gate evidence artifact of that job (.temp/ci-evidence/test/go-test.json), not in the job log, which prints only stage exit codes.
### Scope

Curator Windows lane for internal/buildsource, internal/install and internal/install/atomicity. Diagnose each failure on a native windows-latest runner and fix the real Windows behavior or the Windows expectation. Do not delete, skip or platform-exclude a required case to make the gate pass, and do not weaken the platform-case ledger.
### Acceptance Criteria

Curator Test (windows-latest) reports no required-case failures for internal/buildsource, internal/install or internal/install/atomicity, with the platform-case ledger unchanged or strengthened rather than relaxed.

## Instructions

No specific instructions attached. Work according to the task description and acceptance criteria above.

## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# FIRST/LAST lifecycle commands:
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(BUG-260731-27h1yc, status=analysis)'       # analyst-style work
task-board m 'set_status(BUG-260731-27h1yc, status=development)'    # implementation / testing work
task-board m 'set_status(BUG-260731-27h1yc, status=reviewing)'      # reviewer handoff
task-board m 'set_status(BUG-260731-27h1yc, status=blocked)'        # when blocked
task-board m 'set_status(BUG-260731-27h1yc, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(BUG-260731-27h1yc, item=1)'                        # check item N
task-board m 'add_checklist_item(BUG-260731-27h1yc, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(BUG-260731-27h1yc, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(BUG-260731-27h1yc, name=BUG-260731-27h1yc_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add BUG-260731-27h1yc ./path/to/file --type outcome --name BUG-260731-27h1yc_artifact.bin -d "Description"
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
task-board m 'add_resource(BUG-260731-27h1yc, name=BUG-260731-27h1yc_results.md, content="...", type=outcome, description="Description")'
task-board resource add BUG-260731-27h1yc ./path/to/file --type outcome --name BUG-260731-27h1yc_artifact.bin -d "Description"
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

## Completion Discipline

Keep working until the task reaches a terminal handoff for your role. If no objective blocker remains, do not stop while the board item is still parked in `analysis`, `development`, `testing`, or `reviewing`.

Before your final status change:
- satisfy the task acceptance criteria and relevant checklist items
- attach outcome evidence for the work you produced
- run the relevant verification commands when the task changes code, tests, docs, or config

Use `blocked` only for either a concrete external blocker you cannot resolve autonomously or an unresolved human-only platform/product/architecture/tradeoff/approval decision. Record the constraint, evidence, failed assumptions/attempts, viable alternatives and tradeoffs, recommendation, and exact human decision or external input needed. Recoverable failures and ordinary rework are not `blocked`.

Status language is literal:
- `to-review` means your role has handed work to review; it does not mean the board task is accepted or done.
- In your final response, say "ready for review" or "handed off to review" when the final board status is `to-review`.
- Do not say "done", "complete", "finished", "final", or "готово" as the overall task state unless the board status is actually `done`.

## LAST — Run For Role Handoff

When you have completed all role work and the task is ready for its role handoff, run this as your **final board command**:

```bash
task-board handoff BUG-260731-27h1yc --role developer
```

## Working Directory
Board directory: `/Users/iv/Developer/ReluxWorks/curator/.task-board`

Work in the project root. Do not modify board files directly — always use the `task-board` CLI.
