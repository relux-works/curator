# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(BUG-260731-1rldqv, status=development)'
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

- [ ] Reproduce and isolate each Windows failure family on ssh win or a faithful Windows harness.
- [ ] Fix command-shim digest, replacement, and provenance behavior without weakening corruption guards.
- [ ] Add focused Windows regression tests and keep Linux/macOS behavior green.
- [ ] Publish a signed commit to the CocoaSkills PR 16 branch only on ivanopcode/cocoaskills.
- [ ] Attach evidence, hand off to independent Opus review, and require the full PR 16 matrix green.
- [ ] Code written per task description and AC
- [ ] Relevant tests written for new or changed behavior and passing
- [ ] Lint clean
- [ ] Relevant build/validation commands run after changes and build not broken
- [ ] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
## Your Task

- **ID**: BUG-260731-1rldqv
- **Title**: BUG-260731-1rldqv: windows-transactional-install-regression
- **Parent**: STORY-260720-35dck7
### Description

CocoaSkills PR 16 CI run 30594273278 fails 34-45 tests on every windows-latest cell (3.11: 45 failed/1157 passed; 3.13: 45 failed/1157 passed; 3.14: 34 failed/1168 passed) while all four ubuntu and all four macOS cells and mypy strict are green. Signatures: (1) TransactionCorruptionError "transaction target changed while digesting: C:\...\.csk-materialization-plan-...\home\runtime\<skill>\<sha>\bin\<cmd>.cmd" - every changed target is a .cmd command shim, never .csk-install.json; (2) "[WinError 5] Access is denied"; (3) "cache_publication_invalid: publication artifact source is not ...". Failing files: test_activation_modes, test_audit_cli, test_closure_install, test_dev_substitution, test_gc, test_global_install, test_hybrid_scope, test_install, test_mcp_dependencies, test_status. Provenance established during BUG-260731-2rhy74 RUN-260731-b4fd97: CocoaSkills main b3a5031 is green on all four windows cells (run 30556125542); the transaction engine commits 721ca47 and edbc871 are already on main; c4131bd feat(installer) make project installs transactional (TASK-260720-3t8nr3) is NOT on main; the marker-v2 commit 8a02e17 touches only ci.yml, install_marker.py, installer.py and three test files and nothing in transactions.py, shims.py, runtime materialization or cache publication, and all four of its marker tests PASS on every completed windows cell. PR 16 = main + c4131bd + 8a02e17, therefore the regression entered with c4131bd. It was never seen before because the prior PR 16 run 30589736936 had every windows cell cancelled. Likely area: src/csk/transactions.py digest_path/_digest_file use os.path.samestat between lstat and fstat plus st_size and st_mode re-checks, which do not hold for freshly written .cmd shims on Windows.
### Scope

CocoaSkills transactional project materialization on Windows, including command-shim digest stability, ownership/provenance checks, atomic replacement, and the existing PR 16 matrix. Do not weaken corruption detection or change unrelated marker-v2 semantics.
### Acceptance Criteria

Root cause identified in the transactional materialization path for Windows; the digest/publication guards no longer report spurious change or access-denied on command shims; the failing tests pass on windows-latest for Python 3.11, 3.12, 3.13 and 3.14 without weakening the corruption guard; ubuntu and macOS stay green; PR 16 CI matrix fully green.

## Instructions

No specific instructions attached. Work according to the task description and acceptance criteria above.

## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# FIRST/LAST lifecycle commands:
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(BUG-260731-1rldqv, status=analysis)'       # analyst-style work
task-board m 'set_status(BUG-260731-1rldqv, status=development)'    # implementation / testing work
task-board m 'set_status(BUG-260731-1rldqv, status=reviewing)'      # reviewer handoff
task-board m 'set_status(BUG-260731-1rldqv, status=blocked)'        # when blocked
task-board m 'set_status(BUG-260731-1rldqv, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(BUG-260731-1rldqv, item=1)'                        # check item N
task-board m 'add_checklist_item(BUG-260731-1rldqv, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(BUG-260731-1rldqv, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(BUG-260731-1rldqv, name=BUG-260731-1rldqv_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add BUG-260731-1rldqv ./path/to/file --type outcome --name BUG-260731-1rldqv_artifact.bin -d "Description"
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
task-board m 'add_resource(BUG-260731-1rldqv, name=BUG-260731-1rldqv_results.md, content="...", type=outcome, description="Description")'
task-board resource add BUG-260731-1rldqv ./path/to/file --type outcome --name BUG-260731-1rldqv_artifact.bin -d "Description"
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
task-board handoff BUG-260731-1rldqv --role developer
```

## Working Directory
Board directory: `/Users/iv/Developer/ReluxWorks/curator/.task-board`

Work in the project root. Do not modify board files directly — always use the `task-board` CLI.
