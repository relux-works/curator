# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(BUG-260731-11bpa4, status=development)'
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

- [ ] Reproduce the Windows-only compile/vet failure and identify the intended helper contract.
- [ ] Restore test helper coverage without deleting or skipping the Windows case.
- [ ] Publish a signed Curator PR targeting main and attach Windows plus non-Windows evidence.
- [ ] Obtain independent Opus review and land only after required CI is green.
- [ ] Board size is proportional to the spec and is the smallest decomposition that maps every requirement
- [ ] Every story and task traces to a concrete spec requirement; justified-gap elements also carry a self-verified gap record
- [ ] Beyond-literal-spec elements include a written justification naming the gap and the spec and out-of-scope checks performed before creation
- [ ] Research tasks cite an exact question the spec genuinely leaves open
- [ ] Dependencies linked
- [ ] Tasks are atomic — one clear deliverable each
- [ ] Completeness verified — nothing forgotten
- [ ] Any planning artifacts actually produced are linked as new task-scoped outcome resources; diagrams are strictly optional, never a standing deliverable
- [ ] ARCH-RESOLVED: give the call line its own two-pass escaping in WindowsShimContent (runtimestore.go:191), leaving the set "PATH=..." escaping at one pass (runtimestore.go:182). Expected rule: quadruple % on the call line (pass 1 %%%%->%%, pass 2 %%->%), keep doubling on the set line. DERIVED not executed — must be proven on windows-latest; if the runner disagrees the empirical result wins.
- [ ] ARCH-RESOLVED: drop ONLY the percent%PATH%value fixture argument in targets_windows_test.go (lines 90 and 117), which asserts the unreachable verbatim %VAR% forwarding. Keep the space, embedded-quote, Unicode and empty-string arguments, the % bearing artifact directory immutable cache % Unicode, the PATH assertion and the exit-code 37 assertion. The retained % directory is what proves the call-line escaping fix.
- [ ] ARCH-RESOLVED: document on WindowsShimContent that verbatim %VAR% argument forwarding is out of contract on Windows because %* substitutes arguments on pass 1 and call re-expands on pass 2. No separate docs board item — this is an AC line here.
- [ ] Confirm the globalbins integration point: globalbins.go:353 compares stored shim bytes to a fresh WindowsShimContent(canonical, nil). Both sides recompute so they stay consistent, but a shim already installed under a % bearing path now compares unequal and is treated as unowned. Decide the intended behaviour rather than discovering it in the field.
- [ ] Prove TestWindowsPostInstallWrappersForwardArgumentsPathAndExitCode PASSES on windows-latest in real Curator CI, with platform-cases.tsv row 61 unchanged (must_run_on=windows, skip_allowed_on=-, class=-) and both conformance tests still green.
- [ ] Code written per task description and AC
- [ ] Relevant tests written for new or changed behavior and passing
- [ ] Lint clean
- [ ] Relevant build/validation commands run after changes and build not broken
- [ ] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
## Your Task

- **ID**: BUG-260731-11bpa4
- **Title**: BUG-260731-11bpa4: curator-windows-test-vet-compile-break
- **Parent**: STORY-260720-35dck7
### Description

Curator Test (windows-latest) fails before any test runs: go vet rejects the package with vet.exe: internal\runtimestore\targets_windows_test.go:97:14: undefined: decodeHelperOutput. The windows-only test file references a helper that does not exist in the windows build, so the package does not compile there. Pre-existing on main cfffd7cd and unrelated to any protocol vector; it was masked until BUG-260731-3gm8kc PR 9 repaired the toolchain-identity gate that had been failing every Go job at step 4. Evidence: run 30615765014 job 91108467247, plus the isolated control run on branch ci/goenv-control-BUG-260731-3gm8kc which carries only the toolchain-identity repair.
### Scope

internal/runtimestore windows-only test helpers. Restore the missing helper or the call site so the windows build compiles; do not delete the case to make the gate pass.
### Acceptance Criteria

go vet and go test pass for internal/runtimestore on windows-latest in Curator CI.

## Instructions

The following instructions have been attached to this task:

### BUG-260731-11bpa4_orchestrator-context.md
> Orchestrator execution context: worktree isolation, PR9 base commit, ownership boundary vs BUG-260731-lepevi

# BUG-260731-11bpa4 — orchestrator execution context

## Isolation (mandatory)

The primary checkout `/Users/iv/Developer/ReluxWorks/curator` is on branch
`agent/link-curator-skill-registry` and is dirty with unrelated board files.
Do NOT work in it and do NOT switch its branch.

Create a task-scoped worktree instead:

```bash
git -C /Users/iv/Developer/ReluxWorks/curator worktree add \
  .temp/BUG-260731-11bpa4/worktree \
  -b task/BUG-260731-11bpa4-windows-vet bd6ba08acda3dc801512c408c759ac0ac6f79f26
```

Initialize submodules inside the worktree before running the interop/conformance
suites.

## Base commit rationale (do not silently change it)

Base on `bd6ba08acda3dc801512c408c759ac0ac6f79f26` — the head of Curator PR 9
(`task/BUG-260731-3gm8kc-lifecycle-vector-gate`), not on `main` (`cfffd7c`).

Reason: `main` still carries the broken `.github/ci/toolchain-identity.sh` gate,
which fails every Go job at step 4 before `go vet` ever runs. Branching off main
makes it impossible to demonstrate a green Windows lane for this fix. PR 9
repairs that gate in `bd6ba08`.

If PR 9 merges into `main` while you are working, rebase onto the merged main
commit and say so in your outcome artifact.

## Ownership boundary — a sibling agent is running concurrently

`BUG-260731-lepevi` (Linux lane) is being fixed in parallel in its own worktree.
It owns `internal/godriver`, `internal/transaction`, and `cmd/curator`.

You own `internal/runtimestore` only. Do not touch the sibling's files, do not
"fix" the Linux lint findings, and do not rebase onto its branch. Two separate
PRs, two separate branches.

## Scope reminder

`internal/runtimestore/targets_windows_test.go:97:14: undefined: decodeHelperOutput`.
Restore the missing helper or its call site so the Windows build compiles.
Deleting or skipping the Windows case to make the gate pass is an explicit
non-goal and will be rejected at review.

## Evidence expected

- `go vet` and `go test` green for `internal/runtimestore` on `windows-latest` in
  real Curator CI, not only locally.
- A signed PR targeting `main`.
- Windows plus non-Windows evidence attached as a task-scoped outcome resource.

## Reporting honesty

If a required checklist item is not truly satisfied, leave it unchecked and
explain why in your notes. Do not tick items to get past the handoff gate.





## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# FIRST/LAST lifecycle commands:
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(BUG-260731-11bpa4, status=analysis)'       # analyst-style work
task-board m 'set_status(BUG-260731-11bpa4, status=development)'    # implementation / testing work
task-board m 'set_status(BUG-260731-11bpa4, status=reviewing)'      # reviewer handoff
task-board m 'set_status(BUG-260731-11bpa4, status=blocked)'        # when blocked
task-board m 'set_status(BUG-260731-11bpa4, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(BUG-260731-11bpa4, item=1)'                        # check item N
task-board m 'add_checklist_item(BUG-260731-11bpa4, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(BUG-260731-11bpa4, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(BUG-260731-11bpa4, name=BUG-260731-11bpa4_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add BUG-260731-11bpa4 ./path/to/file --type outcome --name BUG-260731-11bpa4_artifact.bin -d "Description"
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
task-board m 'add_resource(BUG-260731-11bpa4, name=BUG-260731-11bpa4_results.md, content="...", type=outcome, description="Description")'
task-board resource add BUG-260731-11bpa4 ./path/to/file --type outcome --name BUG-260731-11bpa4_artifact.bin -d "Description"
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
task-board handoff BUG-260731-11bpa4 --role developer
```

## Working Directory
Board directory: `/Users/iv/Developer/ReluxWorks/curator/.task-board`

Work in the project root. Do not modify board files directly — always use the `task-board` CLI.
