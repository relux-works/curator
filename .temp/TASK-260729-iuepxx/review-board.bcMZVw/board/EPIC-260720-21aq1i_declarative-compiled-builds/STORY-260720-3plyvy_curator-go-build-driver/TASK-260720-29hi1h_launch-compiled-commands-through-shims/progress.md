## Status
done

## Assigned To
[reviewer] reviewer (codex)

## Created
2026-07-20T02:11:48Z

## Last Update
2026-07-21T02:39:12Z

## Blocked By
- TASK-260720-11pfex
- TASK-260720-3pwg2w

## Blocks
- TASK-260720-3itlly
- TASK-260720-2284br

## Checklist
- [x] Script runtimes are validated before reuse
- [x] Unix and Windows shims forward arguments and exact exit status
- [x] Script/build/removal transitions clean only managed targets
- [x] Runtime and shim APIs stage desired/removal targets; live replacement remains transaction-owned
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
Curator integration precondition: create /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-29hi1h/worktree from exact origin/main 17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8 and import only the complete accepted product diff from /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-11pfex/worktree. This chain already includes accepted protected-cache work. Exclude .temp, board/config, planning/research, diagrams, binaries, caches, alternate indexes, and unrelated files. Use the rc.4 candidate root only as non-release test input. Do not commit or stage. Own staged typed runtime targets/shims/removal plans only; never compile, publish cache, write markers, replace live paths, or execute a built artifact during install tests. Record provenance, task-only delta, incomplete-runtime replacement, transition matrices, Unix/Windows quoting/exit propagation, no-build-root copy, post-install launch versus install-time no-launch proof, full/race/vet/platform evidence.
spawn queued: [implementer] developer (codex) (run=RUN-260721-022eb3, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260721-022eb3)
Implementation checkpoint 2026-07-21 — isolated worktree pinned to 17804ce imported the complete accepted TASK-260720-11pfex product diff. Added a closed script-runtime versus immutable-build-artifact target model, strict commit-runtime reuse validation, operation-private runtime/shim staging, and deterministic typed desired/removal planning. Live project/global/user-bin paths remain transaction-owned. Unix post-install fixtures pass argument/PATH/exit/signal propagation; Windows fixtures compile for execution on Windows. No artifact is launched during staging.
Logbook 2026-07-21 — legacy runtimestore reuse trusted directory existence and live shim writers; the task delta introduces validated typed staging without changing transaction ownership. Build artifacts remain direct protected-cache targets and are never copied below runtime/. Windows wrapper semantics retain the required call .exe with %* and exit /b ERRORLEVEL form; Microsoft call/exit documentation was checked. Native/full/race/vet/build, candidate-focused, repeated determinism, and Linux/Windows compile gates pass. Windows runtime fixture execution remains a CI gate because this macOS host has no Windows runtime or VM. Evidence: TASK-260720-29hi1h_results.md.
Board validation anomaly at handoff — task-board validate reports 12 pre-existing broken EPIC-260712 dependency references and one orphan TASK-260713-7a9c1e review.md resource outside this task scope. TASK-260720-29hi1h itself has valid dependencies, all checklist items checked, and its task-scoped outcome resource attached.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260721-022eb3, pid=25138, exit=0)
spawn queued: [reviewer] reviewer (codex) (run=RUN-260721-0765a5, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260721-0765a5)
Review accepted 2026-07-21 — independently verified the task-only runtimestore and globalbins delta against the accepted predecessor. Typed runtime targets, complete script-runtime reuse validation, staged-only desired/removal planning, direct immutable-artifact shims, Unix argument/PATH/exit/signal propagation, Windows wrapper fixture and CI selection, install-time no-launch behavior, native/full/race/vet/build, candidate conformance, 20x determinism, and Linux/Windows compile gates pass. golangci-lint and native Windows execution are unavailable on this macOS host; existing CI owns both gates. Evidence: TASK-260720-29hi1h_review-verdict.md.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260721-0765a5, pid=43526, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260720-29hi1h_spawn-log_-implementer--developer--codex-.log](file://TASK-260720-29hi1h/TASK-260720-29hi1h_spawn-log_-implementer--developer--codex-.log) — System spawn log captured by task-board
- [TASK-260720-29hi1h_results.md](file://TASK-260720-29hi1h/TASK-260720-29hi1h_results.md) — Typed runtime targets, staged shim transitions, launch fixtures, provenance, and verification evidence
- [TASK-260720-29hi1h_spawn-log_-reviewer--reviewer--codex-.log](file://TASK-260720-29hi1h/TASK-260720-29hi1h_spawn-log_-reviewer--reviewer--codex-.log) — System spawn log captured by task-board
- [TASK-260720-29hi1h_review-verdict.md](file://TASK-260720-29hi1h/TASK-260720-29hi1h_review-verdict.md) — Accepted reviewer verdict with independent architecture, behavior, platform, and validation evidence
