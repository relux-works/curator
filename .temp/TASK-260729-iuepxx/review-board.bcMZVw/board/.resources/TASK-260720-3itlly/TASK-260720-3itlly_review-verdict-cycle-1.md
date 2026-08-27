# TASK-260720-3itlly reviewer verdict — changes requested

## Verdict

Route to `to-dev`. Tests and lint are green, and build ordering/staging cleanup are substantially implemented, but two trust failures can still surface after or escape the required pre-mutation boundary.

## R1 — Toolchain trust is finalized after live mutation (high)

`Project` and `Global` defer `releasePlan` immediately after `planBuilds`, but the deferred call executes only after consumer recording, context/runtime installation, cleanup, shims, env files, adapters, and GC. `releasePlan` calls `BuildPlan.Close`; the real `godriver.Session.Close` re-fingerprints the trusted toolchain before cleanup. Therefore a toolchain drift is reported as a failed install only after live state has changed. `TestSessionReleaseFailureIsReported` confirms the late failure but does not assert that prior project/global installation, runtime, consumer state, and live cache are unchanged.

Evidence: `internal/install/install.go:378,398-415`; `internal/install/global.go:172,188-203`; `internal/install/plan.go:169-194`; `internal/godriver/session.go:285-321`; `internal/install/stage_test.go:768-781`.

Required rework: split pre-handoff trust finalization from private-root cleanup, revalidate the toolchain after the final build and before `OnStaged` or any shared mutation, and add project/global regression coverage proving a verification failure preserves all prior live state byte-for-byte.

## R2 — Frozen source identity is not finalized before cache reuse/handoff (high)

`planBuilds` validates each snapshot once, then `planOne` calls `CacheInspector.Inspect` directly. Cache inspection is not bracketed by `buildsource.Token.Use`, and `BuildPlan.Close` closes tokens without `Recheck`. A cache hit can consequently be selected for identity A, the snapshot can change during/after inspection, and the run can still proceed to `OnStaged` and live materialization. Misses are checked inside each `godriver.Build`, but there is no final all-source recheck after every build and immediately before handoff/shared mutation.

Evidence: `internal/install/plan.go:169-184,220-227,291-315`; `internal/buildsource/buildsource.go:52-55,105-123`; `internal/install/stage.go:78-124`.

Required rework: bracket cache inspection or otherwise recheck its source token, then perform one deterministic final recheck of every planned source before staging handoff/first mutation. Add injected cache-hit and post-build mutation tests that assert failure and unchanged project/global install, consumer ledger, runtime, and live build cache.

## Independent verification

- `git diff --check` — pass
- `go test -count=1 ./internal/install/...` — pass
- `go test -count=1 ./...` — pass
- `go vet ./...` — pass
- `golangci-lint v2.1.6 run ./internal/install/...` — 0 issues

No product code was modified by the reviewer. The finding was also recorded in `LOGBOOK.md`.