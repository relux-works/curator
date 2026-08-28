# Successor-session prompt: finish TASK-260811-33ukne, then pause

You are the successor Orchestrator in
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-agent-skill`.
Use the `project-management` skill and its tracked producer/reviewer workflow.

## Boundary

- Resume only `TASK-260811-33ukne` (`implement-swiftpm-source-resolution-and-closure`).
- It is intentionally still `development`; do not mark it `to-review` or `done`
  without the normal evidence and review lifecycle.
- The prior developer run `RUN-260823-27bddf` was durably cancelled at a safe
  checkpoint after its post-nudge gates completed. Cancellation was requested
  by the user so the work could move to this session; it is not an implementation
  failure.
- Do not start `TASK-260811-tkurtl` (SwiftPM C-family),
  `TASK-260811-2qfnai` (offline build), or
  `TASK-260811-x611eq` (cross-language conformance).
- Stop after this task reaches an accepted review, or after recording a precise
  blocker/changes-requested state. The parent system goal remains active; do not
  call it complete or blocked merely because this session pauses.
- Every spawned producer or reviewer must use Codex `gpt-5.6-sol` with reasoning
  effort `medium`. Kotlin remains out of scope. Vendored compiled binaries remain
  forbidden.

## Authoritative context

Read from the board before acting:

1. `TASK-260811-33ukne_accepted-solution-architecture.md` (precondition).
2. `TASK-260811-33ukne_implementation-evidence_RUN-260823-27bddf.md` (latest
   outcome).
3. Prior reviewer run `RUN-260823-5da0b5` and the earlier changes-requested runs
   `RUN-260823-2a9dae` and `RUN-260823-5178c3`.
4. Final test resources from `RUN-260823-27bddf`, especially
   `TASK-260811-33ukne_full-go_RUN-260823-27bddf.log`, plus race and lint.

The final evidence reports green focused tests, race tests, uncached full
`go test -timeout 30m -count=1 ./...` (`full-go-04`), vet, build, pinned
golangci-lint v2.12.2, canonical-golden verification, binary-deny tests, Kotlin
exclusion, gofmt, `git diff --check`, and `task-board validate`.

## Important implementation state to verify

- Production Git acquisition and mirror verification launch the actual C0-bound
  Git executable through `Config.GitToolRoot`; there is no hashed shell wrapper
  and no ambient-Git authority.
- `ToolIdentity.ProcessFamily` binds exact child executable paths and SHA-256
  digests (including `git-upload-pack` when present); the aggregate C0 fingerprint
  and every pre-start recheck cover that family.
- Acquisition/mirror execution uses the shared manager-owned permit/receipt seam
  in `internal/closureexec/acquisition.go`; SwiftPM adapter production code must
  not own a direct `exec.Command*` seam.
- Portable evidence remains honest (`not-observed` where appropriate); verified
  requests fail closed without portable fallback.
- The narrow artifact-policy authorization admits only the intended
  `source-control-mirror-v1` chain and cannot authorize arbitrary stores or
  verified binaries.
- Drift, escape, tamper, or malformed authorization must reject before any
  process starts. Rust's cross-adapter guard may allow only the two shared seams:
  `portable_runner.go` and `acquisition.go`.
- Canonical schemas, C1/C2/C3 receipt reachability, exact revision/tree, and
  deterministic admitted mirror closure must remain intact.

## Safe continuation procedure

1. Inspect the task, attached resources, current diff, and dirty-worktree state.
   Preserve all existing user and prior-run changes. Do not stage, commit, reset,
   clean, or overwrite unrelated work.
2. Run the developer spawn preflight, then start one tracked recovery/checkpoint
   producer with the exact model policy:
   `task-board q 'project_config(view=spawn-preflight, role=developer, agent=codex)'`
   followed by
   `task-board spawn TASK-260811-33ukne --role developer --background --agent codex --model gpt-5.6-sol --reasoning-effort medium --timeout 4h`.
   Tell it to inspect the latest evidence and code, make only necessary fixes,
   refresh evidence if anything changes, and leave a clean lifecycle handoff.
   Observe it through `task-board spawn observe`; do not ingest the giant raw log.
3. When producer evidence is complete, run
   `task-board handoff TASK-260811-33ukne --role developer`.
4. Run reviewer preflight and spawn exactly one independent tracked reviewer:
   `task-board q 'project_config(view=spawn-preflight, role=reviewer, agent=codex)'`
   and
   `task-board spawn TASK-260811-33ukne --role reviewer --background --agent codex --model gpt-5.6-sol --reasoning-effort medium --timeout 4h`.
   The reviewer must explicitly examine the former wrapper/process-family issue,
   the Rust guard allowlist, canonical schemas, final full/race/lint evidence,
   artifact binary denial, and board validity.
5. If accepted, verify the task is `done`, run `task-board validate`, record the
   pause checkpoint, and stop. If changes are requested, allow one narrowly
   focused tracked producer/reviewer loop, persist exact evidence, and stop
   without starting any of the three subsequent tasks.

The handoff objective is lifecycle completion and independent acceptance of the
already-green SwiftPM implementation—not expansion of scope.
