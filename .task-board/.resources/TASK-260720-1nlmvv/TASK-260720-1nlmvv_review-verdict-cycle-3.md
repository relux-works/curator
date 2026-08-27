# TASK-260720-1nlmvv review verdict — cycle 3

## Verdict: CHANGES REQUESTED

Route: `to-dev`. No product code, tests, staging, commits, or published state were modified during review.

## Blocking finding

### P1 — cache compensation is still not closed over failures inside Publish and Revert

The new compensation is registered only after `Publisher.Publish` returns success. `buildcache.Store.Publish` can already have changed the live cache before returning an error: it quarantines an unusable predecessor at `internal/buildcache/publish.go:171-178`, then after selecting the staged winner it can fail winner validation or cache-root sync at lines 182-190 and return an empty `PublicationResult`. It can also exhaust the selection loop after the predecessor was quarantined and return the line-199 error. `publishWinners` returns on the error at `internal/install/commit.go:710-716` and appends the publication only at lines 718-720, so `runCommit` sees no compensation record for that mutation. A failed install or upgrade can therefore leave the predecessor quarantined, a replacement live or the logical slot absent.

The inverse path is also not fail-closed. `Store.Revert` withdraws the published winner at `internal/buildcache/publish.go:238-240` before restoring the predecessor and syncing at lines 244-248. If either later step fails, the live slot has already changed and no attempt restores the withdrawn winner. `runCommit` only joins that error at `internal/install/commit.go:586-588`; it does not set `BuildCacheRetained`, so presentation can still claim the live build cache is unchanged.

The focused tests do not cover these boundaries. Post-publication faults are injected only after a successful fake or real Publish return. `TestRevertFailsClosed` covers refusals before mutation, not restore or sync failure after withdrawal. This leaves the explicit cycle-3 criterion — cache publication reversible on every post-publication failure for install and upgrade — unsatisfied.

## Required rework

Make every mutation inside Publish observable to the caller even when Publish returns an error, or compensate internally before returning. Make Revert atomic/recoverable across withdraw, restore and sync failures, or preserve a durable recovery record and report the actual retained/changed state without an unchanged-cache claim. Add deterministic fault seams and production-path tests for quarantine, winner selection, winner validation, directory sync, withdrawal, predecessor restore and reversal sync. Exercise both install and upgrade and assert exact predecessor cache bytes plus installed bytes, marker and consumer state.

## Other cycle-2 blockers rechecked

- PASS: build-boundary commit errors use the bounded redaction sink.
- PASS: toolchain-refusal rows carry `driver=go-v1` and the validated build-source identity; the end-to-end CLI test passed.
- PASS: final status classification re-reads receipt and artifact evidence and maps removal, corruption, replacement and protection loss to `build-state-changed`.

## Independent gates

- `go test ./internal/buildcache -run ^(TestRevertRestoresExactlyWhatAPublicationDisplaced|TestRevertFailsClosed)$ -count=1` — PASS.
- Focused `internal/install` cycle-3 tests for commit compensation, redaction and toolchain inventory — PASS.
- `go test ./cmd/curator -run ^TestStatusReportsAnUnusableToolchainPerCompiledCommand$ -count=1 -timeout 15m` — PASS.
- `go vet ./internal/buildcache ./internal/install ./cmd/curator` — PASS.
- `git diff --check` — PASS.

The green gates confirm this is an uncovered semantic failure boundary, not a mechanical regression in the covered paths.