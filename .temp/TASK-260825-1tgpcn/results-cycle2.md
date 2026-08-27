# TASK-260825-1tgpcn — rework after review cycle 1

Scope: exactly the change `TASK-260825-1tgpcn_review-verdict.md` requested and
nothing else. The environment construction, the refusal guards, the read-back
verification and the message text were reviewed and accepted; none of them was
touched.

## The defect

`Access.call` bounded the wrong process. `exec.CommandContext` kills Git; a
credential helper is Git's *child*. Because `cmd.Stdout`/`cmd.Stderr` are not
`*os.File`, os/exec wires OS pipes with copy goroutines, and the helper
inherits the write end of Git's stderr. When the helper is the wedged one the
timeout exists for — a locked keychain, a session with no desktop behind it —
the kill removes Git, the orphan keeps that pipe open, and with
`cmd.WaitDelay == 0` `cmd.Run` returns only when the orphan does. Every entry
point (`ReadHost`, `ReadScoped`, `StoreScoped`, `DeleteScoped`, and `Discover`
once per scope) sits on that path.

## The fix

`internal/gitcred/gitcred.go`

- new `drainDelay = 2 * time.Second`, documented as what bounds the wait
  *after* the deadline has killed the call;
- `cmd.WaitDelay = drainDelay` in `Access.call`.

One credential call now costs at most `Timeout + drainDelay`.

`internal/gitcred/gitcred_test.go`

- `modeHangGrandchild` — the stand-in Git spawns a helper and blocks on it,
  the way Git blocks on a helper that never answers;
- `modeSleeper` — that helper: the test binary re-executed once more, handed
  the stand-in Git's own stderr (the pipe the manager is reading), sleeping
  `sleeperLifetime` (30s). It records nothing, so it is not counted as a
  credential call;
- `TestACallIsBoundedWhenTheHelperOutlivesGit` — 200ms deadline, asserts the
  call returns in under `sleeperLifetime/2`.

The existing `TestACallIsBounded` is untouched and still passes; it cannot
cover this case because it wedges the stand-in Git itself, and killing that
closes its own pipes.

## Regression proof

The new test was verified to fail with the fix removed, on this host
(darwin/arm64, go1.25.5):

| Case | `cmd.WaitDelay` | Elapsed | Result |
| --- | --- | ---: | --- |
| grandchild helper outlives the kill | unset | 30.008s | FAIL — call outlived the orphan |
| grandchild helper outlives the kill | `drainDelay` (2s) | 2.21s | PASS |
| stand-in Git itself hangs (existing test) | `drainDelay` | 0.21s | PASS |

## Validation (each run standalone, real exit codes)

| Command | Exit |
| --- | ---: |
| `go build ./...` | 0 |
| `go test -count=1 ./internal/gitcred/` | 0 |
| `go test -race -count=1 ./internal/gitcred/` | 0 (21.3s) |
| `go test -timeout 30m -count=1 ./...` | 0 (42 packages ok) |
| `go vet ./...` | 0 |
| `golangci-lint run ./internal/gitcred/...` | 0 (0 issues) |
| `gofmt -l internal/ cmd/` | 0 (no output) |

`golangci-lint run ./...` over the whole tree exits **1** with 5 issues, all
outside this task: `cmd/curator/main.go:2179,2181` (errcheck) and
`internal/install/buildhttps.go:46,63,87` (revive, missing `GoString`
comments). Both files were written at 02:16–02:17 by sibling tasks still in
flight (`internal/install/buildhttps.go` did not exist when this rework
started at 02:10); they are not touched by this delivery and were left for
their owners. The package-scoped lint for this task is clean.

## Logbook

The reviewer asked for the false "(15s bound)" claim in entry `2026-08-25 0158`
to be corrected. The logbook is append-only, so it is corrected by a new entry
rather than by rewriting that one: `2026-08-25 0223` opens by naming the claim
it supersedes, and records the mechanism, the fix sites, why the existing test
could not see it, the measured 30.008s vs 2.21s, and a NOTE that the same hole
exists anywhere in this repo an `exec.CommandContext` child may spawn its own
child over non-`*os.File` pipes.
