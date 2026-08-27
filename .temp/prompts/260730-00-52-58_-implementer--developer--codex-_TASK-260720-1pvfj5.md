# Agent Task Assignment

## FIRST — Run Immediately

```bash
task-board m 'set_status(TASK-260720-1pvfj5, status=development)'
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
6. For board reads, use compact task-specific projections. A concrete assignment does not need routine `summary()`, `plan()`, `schema()`, or `{ full }`; request scoped schema only after an unknown call.

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

- [ ] CI pins the reviewed immutable rc.4 protocol commit
- [ ] Linux, macOS, and Windows exercise their platform-specific behavior
- [ ] Race, vet, formatting, lint, and acceptance evidence are required
- [ ] Candidate suite input is explicit and never advances or impersonates the qualified released pin
- [ ] Keep every default committed protocol pin on the previous release; supply the candidate suite only through a non-default immutable input
- [ ] Composite proven: 356 product manifest entries byte-identical to the accepted TASK-260720-jrrgw9 candidate (cmp exit 0); overlay is CI/quality files only
- [ ] Default pin lane and rc.5 candidate lane both exit 0 through .github/ci/test-gate.sh on the composite
- [ ] Gate self-test 70/70 and ledger-consistency 49 rows across linux/darwin/windows both exit 0
- [ ] Implementation matches AC
- [ ] Solution fits project architecture
- [ ] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches
- [ ] Ambiguous candidate_ref plus candidate_root dispatch fails before checkout or evidence stamping
- [ ] Focused regression keeps ref-only and root-only candidate paths valid
- [ ] Code written per task description and AC
- [ ] Relevant tests written for new or changed behavior and passing
- [ ] Lint clean
- [ ] Relevant build/validation commands run after changes and build not broken
- [ ] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
## Your Task

- **ID**: TASK-260720-1pvfj5
- **Title**: Enforce cross-platform compiled-build CI gates
- **Parent**: STORY-260720-3plyvy
### Description

Enforce cross-platform candidate compiled-build tests and Go quality gates without advancing the committed protocol-suite pin before the schema v6 protocol release is qualified.
### Scope

Own .github/workflows/ci.yml, Makefile or narrowly related quality configuration, and task-scoped verification logs. Keep the committed curator-spec checkout at the currently qualified released revision during candidate development. Provide an explicit caller-supplied full candidate revision or CURATOR_CONFORMANCE_ROOT path for schema v6 qualification; never commit a branch, mutable tag, placeholder, guessed hash, or unreleased candidate as the official pin. Run the full supplied candidate suite on Linux, macOS, and Windows using the repository Go version, keep gofmt and go vet, retain golangci-lint, and add a supported race job on at least Linux covering transaction, cache, install, and conformance packages. Ensure Windows exercises DACL or reparse checks and .cmd launchers and Unix jobs exercise permission, link, readonly-source, resource-policy, and executable behavior. TASK-260720-38l1sy owns the official released-suite pin promotion and audit only after TASK-260720-25d05o qualifies that release.
### Acceptance Criteria

Normal Curator CI keeps one immutable currently released curator-spec pin, while an explicitly supplied schema v6 candidate root or full revision runs every compiled-build case on ubuntu, macos, and windows with no case silently skipped except protocol-defined unsupported platform controls. Candidate evidence records the exact suite revision and digest and cannot be presented as a published release or conformance claim. go test -race ./... passes on the selected supported runner; go vet ./..., gofmt check, and golangci-lint pass with no broad suppression for new security code. The Windows gate executes DACL or reparse and .cmd cases; Unix gates execute ownership, no-follow, readonly-source, resource-policy, and executable cases. No README release wording or committed suite pin claims rc.4 before TASK-260720-25d05o, and the task handoff gives TASK-260720-38l1sy exact candidate CI evidence for the later released-pin audit.

## Instructions

The following instructions have been attached to this task:

### TASK-260720-1pvfj5_composite-rework.md
> Mandatory accepted-composite rework instructions

# Mandatory composite rework

The first RUN-260729-9ead43 outcome is diagnostic only and must be superseded. It validated origin/main without the accepted compiled-build implementation.

1. Use /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/worktree as the authoritative accepted product candidate. Verify it against TASK-260720-jrrgw9_production-integration-results.md, verifier4 evidence, and the final review verdict. This candidate already represents TASK-260729-2kaopg plus the accepted TASK-260720-2qqq0w implementation and accepted jrrgw9 integration.
2. Create a fresh TASK-260720-1pvfj5 rework composite; preserve every accepted product byte. Overlay only the CI/Makefile/narrow quality files that remain valid from the cancelled run. Do not use origin/main-only package inventory.
3. CI gates must exercise actual godriver, buildcache/buildsource, transaction/cache, install/atomicity, interop/conformance, DACL/reparse, readonly-source, resource-policy, executable and no-follow behavior as present in the accepted candidate. Do not downgrade an AC because origin/main lacks it.
4. Keep Go 1.25.5, the current qualified released pin, and candidate-only immutable revision/root semantics. No pin advancement or release/conformance claim.
5. Do not overlap heavy Go suites. Reuse accepted verifier4 full/race evidence where byte identity permits, and run only focused CI-script selftests/gates needed for the CI delta. Record exact composite provenance and hashes.
6. No stage, commit, publish, pin change, broad suppression, fixture weakening, timeout inflation, or product-code mutation. Attach a superseding task-scoped patch/results packet and hand off only when the exact accepted composite is proven.


### TASK-260720-1pvfj5_final-integration.md
> Final Curator integration recipe after the two accepted product blockers

Resume only after BUG-260729-r0fe02 and BUG-260729-1o0m8f are done. Use .temp/TASK-260720-1pvfj5/rework/composite as the accepted CI-overlay composite. Apply verbatim the accepted BUG-260729-1o0m8f_lint-fix.patch and BUG-260729-r0fe02_patch.diff from their board resource paths. Do not change the 15-file CI overlay, any release/candidate pin, protocol vectors, timeouts, suppression policy, or unrelated product bytes. Prove the final delta is exactly the prior accepted composite plus those two accepted patches, with the original 356-entry product manifest changed only at the seven patch-owned files. Run pinned golangci-lint v2.12.2, gofmt/vet/build/no-broad-suppression, default-pin and explicit rc.5 candidate test gates, the focused deterministic godriver cancellation gate, and exactly one final serialized full race gate with no other heavy Go test active. Reuse prior selftest/ledger evidence if scripts are byte-identical; rerun only if identity differs. Keep stale checklist item 1 unchecked as superseded by task notes; close lint item and any genuine remaining DoD, attach final integration patch/evidence/outcome, then hand off for independent review. Never stage, commit, publish, or advance a pin.


### TASK-260720-1pvfj5_final-review.md
> Independent final-review boundaries; reuse heavy evidence and preserve release pin

Review the final integrated composite and attached final-integration outcome/evidence. Treat the accepted CI overlay and independently accepted BUG-260729-1o0m8f and BUG-260729-r0fe02 patches as the exact allowed inputs. Verify the 372-to-374 manifest proof, seven owned paths, unchanged 16-file CI/quality overlay, unchanged released SPEC_PIN, explicit candidate-only rc.5 identity, pinned lint 0 issues, default/candidate gate exits, and the single serialized full race exit with no diagnostics. Reuse the attached 372s/352s/471s gate evidence; do not run another full go test, full race, or candidate/default test-gate. Narrow static/hash/log checks are allowed. The lone unchecked checklist item about pinning rc.4 is stale and explicitly superseded by current scope/AC/notes; acceptance must not advance that pin. Record an evidence-backed verdict and route the task through the reviewer branch.


### TASK-260720-1pvfj5_candidate-input-rework.md
> Focused final-review rework for candidate input provenance

# Focused final-review rework

Fix the P1 provenance defect from TASK-260720-1pvfj5_review-verdict.md in .temp/TASK-260720-1pvfj5/rework/composite. workflow_dispatch currently permits candidate_ref and candidate_root together and can stamp one revision beside another root digest. Implement fail-fast rejection of the both-nonempty combination before checkout or evidence recording. Add a focused regression proving: both inputs fail; ref-only remains valid; root-only remains valid. Preserve SPEC_PIN, candidate-only wording, the seven accepted product paths, candidate identity semantics, and every unrelated overlay byte. Do not rerun default/candidate/full/race Go suites solely for this narrow fix. Run only affected gate self-test/static workflow checks plus formatting or shell checks directly relevant to changed files. Reuse the already accepted heavy-suite evidence where bytes permit. Attach a task-scoped rework outcome with exact changed paths and command exits, then route to review.




## Board CLI Reference

You have access to `task-board` CLI for managing your work. All writes use the `m` (mutation DSL) subcommand:

```bash
# FIRST/LAST lifecycle commands:
# Use explicit status changes when the task flow requires it.
task-board m 'set_status(TASK-260720-1pvfj5, status=analysis)'       # analyst-style work
task-board m 'set_status(TASK-260720-1pvfj5, status=development)'    # implementation / testing work
task-board m 'set_status(TASK-260720-1pvfj5, status=reviewing)'      # reviewer handoff
task-board m 'set_status(TASK-260720-1pvfj5, status=blocked)'        # when blocked
task-board m 'set_status(TASK-260720-1pvfj5, status=to-review)'      # when your work is ready for review

# Track progress with checklist
task-board m 'check_item(TASK-260720-1pvfj5, item=1)'                        # check item N
task-board m 'add_checklist_item(TASK-260720-1pvfj5, text="Write tests")'    # add checklist item

# Add notes about your progress, decisions, blockers, or review findings
task-board m 'set_notes(TASK-260720-1pvfj5, text="your note here")'

# Save short text directly as outcome resources
task-board m 'add_resource(TASK-260720-1pvfj5, name=TASK-260720-1pvfj5_results.md, content="...", type=outcome, description="Description")'

# Attach an existing local file (screenshots, PDFs, logs, archives, research docs)
task-board resource add TASK-260720-1pvfj5 ./path/to/file --type outcome --name TASK-260720-1pvfj5_artifact.bin -d "Description"
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
task-board m 'add_resource(TASK-260720-1pvfj5, name=TASK-260720-1pvfj5_results.md, content="...", type=outcome, description="Description")'
task-board resource add TASK-260720-1pvfj5 ./path/to/file --type outcome --name TASK-260720-1pvfj5_artifact.bin -d "Description"
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
task-board handoff TASK-260720-1pvfj5 --role developer
```

## Working Directory
Board directory: `/Users/iv/Developer/ReluxWorks/curator/.task-board`

Work in the project root. Do not modify board files directly — always use the `task-board` CLI.
