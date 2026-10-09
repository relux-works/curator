## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(3))

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] Fix implemented at the production boundary named in the report
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Regression test through the production entry written; its red/green proof comes from hosted CI (scratch-branch run URL in the results)
- [x] Report's negative controls for this area green in that hosted run
- [x] No local go test on the mini (R223): compile-only locally, tests on runners
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn selection rationale tuple: {"role":"developer","pair":"muse-spark-1.3-contributor/max","text":"tb-R164 developer muse-spark-1.3-contributor max; tb-R136 health-gated queue"}
spawn selection rationale for muse-spark-1.3-contributor/max: tb-R164 developer muse-spark-1.3-contributor max; tb-R136 health-gated queue
spawn agent resolution: Agent selection: muse via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (muse) (run=RUN-261008-d579e5, max_parallel=20)
spawn run RUN-261008-d579e5 failed; operator action required; failure: queued spawn preparation failed: change_request_sibling_producer_blocked: refusing producer TASK-261008-eaqg8a: TASK-261008-2uq6jo holds unresolved Change Request CR-TASK-261008-2uq6jo-2 revision 2 (state=changes_requested) in Story STORY-261008-19ygi3 (blocking_cr=CR-TASK-261008-2uq6jo-2, blocking_state=changes_requested, blocking_task=TASK-261008-2uq6jo, element_id=TASK-261008-eaqg8a, integration_scope=STORY-261008-19ygi3)
spawn selection rationale for muse-spark-1.3-contributor/max: tb-R164 developer muse-spark-1.3-contributor max; tb-R136 health-gated queue
spawn agent resolution: Agent selection: muse via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (muse) (run=RUN-261008-794f90, max_parallel=20)
spawn run started: [implementer] developer (muse) (run=RUN-261008-794f90)
N7 fix ready for review. Unchecked items 2,3,4,6 need the hosted gate (R223 bars local go test; vet+build green locally). Item 10 N/A: brief forbids LOGBOOK edits; findings are in TASK-261008-eaqg8a_results.txt.
Item 10 checked: brief orders No LOGBOOK edits, so the when-relevant condition does not trigger; findings are in TASK-261008-eaqg8a_results.txt. Items 2,3,4,6 need test execution barred by R223 / the hosted gate; handoff refused (exit 1). Stop-The-Line packet appended to the results resource; setting blocked.
spawn run child final message (run=RUN-261008-794f90, tools=59 patches=5 failed=2):
Blocked, not handed off: the producer handoff gate requires test-execution evidence (checklist items 2, 3, 4, 6) that R223 forbids me from generating locally, and the hosted gate has not run on this uncommitted candidate. Checking those items without runs would violate the Evidence Honesty Contract, and deleting them would destroy the reviewer's verification surface — so the task is now `blocked` with a Stop-The-Line packet in `TASK-261008-eaqg8a_results.txt` naming the exact orchestrator decision needed (run the hosted gate and verify items 2, 3, 4, 6 at review time, or re-scope the producer-handoff checklist).

The code work itself is finished and green within the compile-only envelope:

- Fix: `StageForwarding` stages a recognized manager-owned legacy forwarding symlink as an owned directory entry (`ReplaceEntry`/`RemoveEntry` via a new `ownedLegacyLink` helper in [globalbins.go](/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-261009-2kq6u5/worktree/internal/globalbins/globalbins.go); [stage.go](/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-261009-2kq6u5/worktree/internal/globalbins/stage.go)). Only `ownedTarget`-recognized links change kind; foreign/unrecognized links still never become targets, and regular-shim paths are byte-identical to before.
- Regression test: `TestGlobalReconcilesOwnedLegacyForwardingLink` in [global_legacy_link_test.go](/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-261009-2kq6u5/worktree/internal/install/global_legacy_link_test.go) drives `install.Global` through all four probe shapes, including both negative controls, plus a byte-exact launcher assertion on the reconciled shim (legacy-link shape skips on Windows, where the form doesn't exist).
- CHANGELOG line under Unreleased → Fixed.
- Validation: `gofmt` clean, `go vet ./...` exit 0, `go build ./...` exit 0. `go test` not run per R223.

Work is uncommitted in the story worktree as required. No code change is needed from the producer side — this needs the orchestrator's routing decision on the gate-dependent checklist items.
agent completed: [implementer] developer (muse) (exit=0)
spawn run completed: muse (run=RUN-261008-794f90, pid=89068, exit=0)
\## Stop-The-Line
### Blocker
Producer handoff is refused by the checklist gate while the task's own
rules forbid the producer from generating the missing evidence. Items
2, 3, 4, 6 all require test execution: R223 bars any local `go test` on
the mini (compile-only: `go vet ./...`, `go build ./...`), and the
Evidence Honesty Contract forbids checking a command-tied item unless
that exact command ran green with exit 0. The hosted gate — the named
arbiter — has not run on this uncommitted candidate. A same-brief retry
is deterministic: a second identical handoff refusal becomes terminal
`recovery_parked`, so the run stops here instead of retrying.
### Evidence
- `task-board handoff TASK-261008-eaqg8a --role developer` -> exit 1:
  "unchecked checklist items [2 3 4 6 10] ... handoff evidence missing".
- `gofmt -l internal/globalbins/ internal/install/` -> clean (no files).
- `go vet ./...` -> exit 0, no diagnostics.
- `go build ./...` -> exit 0, no diagnostics.
- `go test` -> NOT RUN (R223). No local suite evidence exists or may exist.
- Item 2 red half ("fails without the fix") is established by already-
  attached evidence: wave-2 report command 07, exit 1, N7 probe red
  (`source_output_overlap ... is now a link`); the adapted test keeps the
  identical assertion, production entry (install.Global), and fixture
  (Refresh-emitted legacy link). Green half, negative-control rerun, and
  hosted-gate verdict have no evidence anywhere yet.
- Item 10 was checked by the producer: the brief explicitly orders "No
  LOGBOOK edits", so the item's "when relevant" condition does not
  trigger; findings live in this resource.
### Options
- (a) Orchestrator runs the hosted gate on the candidate revision and the
  reviewer completes items 2, 3, 4, 6 against the live checklist at
  accept_cr (the review brief already assigns this: "accept_cr with the
  live checklist if it holds"). Recommended: matches the brief's design
  ("The hosted gate is the arbiter").
- (b) Orchestrator re-scopes the producer-handoff DoD so gate-dependent
  items verify at review time rather than at producer handoff.
- (c) Waive R223 for one bounded local run. NOT recommended: violates an
  explicit task constraint and the mini policy behind it.
Checking the items without runs, or deleting them to pass the gate, were
rejected: the first violates the Evidence Honesty Contract, the second
destroys the reviewer's verification surface.
### EXACT INPUT NEEDED
Orchestrator decision: run the hosted gate on this candidate and route
items 2, 3, 4, 6 to review-time verification (option a), or re-scope the
producer-handoff checklist (option b). No code change is needed from the
producer: the fix, regression test, CHANGELOG line, and compile-only
validation are attached and green.
spawn selection rationale for muse-spark-1.3-contributor/max: tb-R164 developer muse-spark-1.3-contributor max; tb-R136 health-gated queue
spawn agent resolution: Agent selection: muse via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (muse) (run=RUN-261009-3ac756, max_parallel=20)
spawn run RUN-261009-3ac756 failed; operator action required; failure: queued spawn preparation failed: worktree_base_fast_forward_blocked: 1 uncommitted path(s) in the STORY-261009-2kq6u5 workspace are also changed by the incoming authority 3d395ffe72ec979e2ef1d3d792655a7f405785cc, so the fast-forward would overwrite work that exists nowhere else (branch_oid=3b6481c15d61b08d3f0fae7c3269329555336709, branch_ref=refs/heads/task-board/story/STORY-261009-2kq6u5, checkpoint_oid=3b6481c15d61b08d3f0fae7c3269329555336709, dirty_path_count=4, execution_root=/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-261009-2kq6u5/worktree, head_oid=3b6481c15d61b08d3f0fae7c3269329555336709, incoming_path_count=25, integration_ref=refs/heads/main, overlapping_paths=CHANGELOG.md, reason=dirty_paths_overlap_incoming_delta, remediation=abort, remediation_command=commit or discard the listed paths, or task-board worktree abort STORY-261009-2kq6u5, selected_oid=3d395ffe72ec979e2ef1d3d792655a7f405785cc, story_id=STORY-261009-2kq6u5)
spawn selection rationale for muse-spark-1.3-contributor/max: tb-R164 developer muse-spark-1.3-contributor max; tb-R136 health-gated queue
spawn agent resolution: Agent selection: muse via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (muse) (run=RUN-261009-be2516, max_parallel=20)
spawn run started: [implementer] developer (muse) (run=RUN-261009-be2516)
Developer handoff: fix + regression test + CHANGELOG in worktree (uncommitted, 4 paths). Hosted evidence: green2 run 37868830965 success, red2 run 37868866996 failure with only owned-legacy-link red; URLs in TASK-261008-eaqg8a_results.md. Item 7 basis: findings recorded in the results resource; repo LOGBOOK.md untouched per brief No LOGBOOK edits. Item 10: zero local go test invocations; go vet + go build exit 0. Branch-name adaptation: gate/TASK-* used instead of scratch/* because the CI push trigger covers only main and gate/**; all 4 evidence branches deleted.
spawn run child final message (run=RUN-261009-be2516, tools=115 patches=7 failed=2):
Ready for review — TASK-261008-eaqg8a (N7) is handed off to review with checklist 10/10.

**Fix.** `install.Global` now reconciles a recognized manager-owned legacy forwarding symlink instead of failing with `source_output_overlap`. In [stage.go](/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-261009-2kq6u5/worktree/internal/globalbins/stage.go), `StageForwarding` stages such links as owned directory entries (`ReplaceEntry`/`RemoveEntry`) via a new `ownedLegacyLink` helper in [globalbins.go](/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-261009-2kq6u5/worktree/internal/globalbins/globalbins.go). Only `ownedTarget`-recognized links change kind, so the exact link target flows through the existing transaction journal/preimage/rollback machinery while foreign shims, changed preimages, and the no-follow-links rule are untouched. CHANGELOG line re-added under Unreleased → Fixed.

**Regression test.** `TestGlobalReconcilesOwnedLegacyForwardingLink` in [global_legacy_link_test.go](/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-261009-2kq6u5/worktree/internal/install/global_legacy_link_test.go) drives the production entry across all four shapes, including both negative controls.

**Evidence (all hosted, zero local `go test` per R223).** Local: `go vet ./...` and `go build ./...` exit 0, gofmt clean. Hosted full matrix: [green run 37868830965](https://github.com/relux-works/curator/actions/runs/37868830965) (success, all 4 subtests pass, 0 failures) vs [red run 37868866996](https://github.com/relux-works/curator/actions/runs/37868866996) (fix-only revert; sole failure is `owned-legacy-link`, siblings green). Full detail, including snapshot byte-identity checks, is in `TASK-261008-eaqg8a_results.md` on the board.

Two adaptations to note: branches were pushed as `gate/TASK-*` rather than `scratch/*` because the CI push trigger covers only `main` and `gate/**` (all deleted after recording); and the Windows skip reason was reworded to match an existing `platform-control` skip class after the first hosted attempt failed only the platform-case gate — no CI tables were modified.
agent completed: [implementer] developer (muse) (exit=0)
spawn run completed: muse (run=RUN-261009-be2516, pid=79136, exit=0)
spawn selection rationale tuple: {"role":"reviewer","pair":"gpt-6.1-sol/high","text":"R187/R195 reviewer sol high (cross-provider)"}
spawn selection rationale for gpt-6.1-sol/high: R187/R195 reviewer sol high (cross-provider)
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (codex) (run=RUN-261009-e838c7, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-261009-e838c7)
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-261009-e838c7, pid=45736, exit=0)
run write-boundary clearance for RUN-261009-be2516: Operator review 2026-09-23: every flagged path is board state under .task-board/ (activity streams, journals and element files of other elements) or inside OTHER Stories' managed worktrees under .temp/STORY-*/worktree, written concurrently by the orchestrator's own board mutations and by other tracked runs during this run's lifetime; no source path outside the Story worktree was written.
run write-boundary clearance for RUN-261009-e838c7: Operator review 2026-09-23: every flagged path is board state under .task-board/ (activity streams, journals and element files of other elements) or inside OTHER Stories' managed worktrees under .temp/STORY-*/worktree, written concurrently by the orchestrator's own board mutations and by other tracked runs during this run's lifetime; no source path outside the Story worktree was written.
spawn selection rationale tuple: {"role":"developer","pair":"gpt-6-astra/low","text":"bound eaqg8a-land (land queue); codex gpt-6-astra low"}
spawn selection rationale for gpt-6-astra/low: bound eaqg8a-land (land queue); codex gpt-6-astra low
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (codex) (run=RUN-261009-d356f1, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-261009-d356f1)
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-261009-d356f1, pid=84989, exit=0)

## Precondition Resources
- [w2fix-N7-brief.md](file://TASK-261008-eaqg8a/w2fix-N7-brief.md)
- [w2fix-N7-install-probe_test.go](file://TASK-261008-eaqg8a/w2fix-N7-install-probe_test.go)
- [w2fix-N7-review.md](file://TASK-261008-eaqg8a/w2fix-N7-review.md)
- [w2fix-hosted-evidence.md](file://TASK-261008-eaqg8a/w2fix-hosted-evidence.md)
- [N7-changelog.patch](file://TASK-261008-eaqg8a/N7-changelog.patch)
- [eaqg8a-integrate-land.md](file://TASK-261008-eaqg8a/eaqg8a-integrate-land.md)

## Outcome Resources
- [TASK-261008-eaqg8a_spawn-log_-implementer--developer--muse-_RUN-261008-d579e5.log](file://TASK-261008-eaqg8a/TASK-261008-eaqg8a_spawn-log_-implementer--developer--muse-_RUN-261008-d579e5.log) — System spawn log captured by task-board
- [TASK-261008-eaqg8a_spawn-log_-implementer--developer--muse-_RUN-261008-794f90.log](file://TASK-261008-eaqg8a/TASK-261008-eaqg8a_spawn-log_-implementer--developer--muse-_RUN-261008-794f90.log) — System spawn log captured by task-board
- [TASK-261008-eaqg8a_results.txt](file://TASK-261008-eaqg8a/TASK-261008-eaqg8a_results.txt)
- [TASK-261008-eaqg8a_spawn-log_-implementer--developer--muse-_RUN-261009-3ac756.log](file://TASK-261008-eaqg8a/TASK-261008-eaqg8a_spawn-log_-implementer--developer--muse-_RUN-261009-3ac756.log) — System spawn log captured by task-board
- [TASK-261008-eaqg8a_spawn-log_-implementer--developer--muse-_RUN-261009-be2516.log](file://TASK-261008-eaqg8a/TASK-261008-eaqg8a_spawn-log_-implementer--developer--muse-_RUN-261009-be2516.log) — System spawn log captured by task-board
- [TASK-261008-eaqg8a_results.md](file://TASK-261008-eaqg8a/TASK-261008-eaqg8a_results.md) — N7 developer results: fix, regression test, compile-only tail, hosted red/green run URLs
- [TASK-261008-eaqg8a_change-request_rev1.patch](file://TASK-261008-eaqg8a/TASK-261008-eaqg8a_change-request_rev1.patch) — Change Request CR-TASK-261008-eaqg8a-1 revision 1 candidate patch (repository_delta=present, 4 changed paths)
- [TASK-261008-eaqg8a_change-request_rev1-validation.log](file://TASK-261008-eaqg8a/TASK-261008-eaqg8a_change-request_rev1-validation.log) — Change Request CR-TASK-261008-eaqg8a-1 revision 1 bounded validation log
- [TASK-261008-eaqg8a_spawn-log_-reviewer--reviewer--codex-_RUN-261009-e838c7.log](file://TASK-261008-eaqg8a/TASK-261008-eaqg8a_spawn-log_-reviewer--reviewer--codex-_RUN-261009-e838c7.log) — System spawn log captured by task-board
- [TASK-261008-eaqg8a_review-verdict-rev1.md](file://TASK-261008-eaqg8a/TASK-261008-eaqg8a_review-verdict-rev1.md) — Revision 1 acceptance: swept surfaces, exact-tree hosted red/green evidence, compile-only reviewer checks
- [TASK-261008-eaqg8a_acceptance-runtime-note-rev1.md](file://TASK-261008-eaqg8a/TASK-261008-eaqg8a_acceptance-runtime-note-rev1.md) — Successful acceptance with nonblocking runtime write-attribution warning
- [TASK-261008-eaqg8a_spawn-log_-implementer--developer--codex-_RUN-261009-d356f1.log](file://TASK-261008-eaqg8a/TASK-261008-eaqg8a_spawn-log_-implementer--developer--codex-_RUN-261009-d356f1.log) — System spawn log captured by task-board
- [TASK-261008-eaqg8a_integration-preflight_RUN-261009-d356f1.md](file://TASK-261008-eaqg8a/TASK-261008-eaqg8a_integration-preflight_RUN-261009-d356f1.md) — Fresh accepted-revision and candidate preflight for runner-owned integration

## Created
2026-10-08T18:19:45Z

## Last Update
2026-10-09T03:50:23Z

## Assigned To
[implementer] developer (codex)
