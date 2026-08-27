# TASK-260825-1tgpcn — Review Verdict (cycle 1)

**Verdict: CHANGES REQUESTED → to-dev.** One confirmed correctness defect in the
package's own stated guarantee; everything else in the delivery is verified and
accepted as-is. The rework is one line of product code plus one harness variant.

## What was verified and holds

Reviewed `internal/gitcred/gitcred.go` + `gitcred_test.go` (new package, nothing
else in the repo imports it yet — confirmed by grep, so the delivery is a leaf
exactly as the results artifact claims).

Against the AC:

- **Every read and write is `git credential fill|approve|reject`** — no platform
  branch anywhere in the package; pinned by `TestReadHostGoesThroughGitCredentialFill`
  argv/payload assertions.
- **Prompting disabled on every call** — `GIT_TERMINAL_PROMPT=0`,
  `GCM_INTERACTIVE=never`, `-c credential.interactive=false`, `-c core.askPass=`
  with `GIT_ASKPASS`/`SSH_ASKPASS` removed (not emptied). The removal-not-empty
  reasoning is correct: git's prompt source chain treats a set-but-empty
  `core.askpass` as "stop the askpass chain", and the terminal fallback is closed
  by `GIT_TERMINAL_PROMPT=0`. Pinned by `TestEveryCallDisablesInteractivePrompting`.
- **Operator home pinned** — `HOME` and `USERPROFILE` both overridden,
  case-insensitive suppression for Windows spellings; the real-git test proves the
  store lands in the pinned home's `.git-credentials`.
- **Non-persisting helper caught by read-back with platform guidance** — both
  harnesses cover it: the stand-in git's `silent-approve` defect, and real git
  with no helper configured (approve exits 0, keeps nothing).
  `assertPlatformGuidance` checks all three platform stores plus the env-var
  alternative, and that the secret does not leak into the message.
- **No collision with the operator's own entry** — distinct username
  `curator-build-https:<scope>`, plus the two refusal guards (host read refuses a
  namespaced answer; scoped read refuses another username). The real-git test
  pins the `store`-helper-prepends behaviour and proves fail-closed rather than
  wrong-credential.
- **Validation re-run by the reviewer**: `go test -count=1 ./internal/gitcred/`
  ok; `go test -race -count=1 ./internal/gitcred/` ok (18.3s);
  `go build ./...` ok; `go vet ./...` ok; `golangci-lint run internal/gitcred/...`
  0 issues; `gofmt -l internal/gitcred/` clean. Logbook entry 2026-08-25 0158
  present and accurate (except the 15s-bound claim, below).

Architecture fit: dependency-free leaf, zero-value-usable `Access`, consumed
later by the resolution/broker/CLI tasks — fits the plan and creates no cycle.

## The defect: the 15s bound does not hold for the case it was written for

`DefaultTimeout`'s own doc comment says: *"A helper that talks to a locked
keychain can hang indefinitely; the manager treats that as no material."* That
guarantee is not delivered.

`Access.call` uses `exec.CommandContext` with a 15s context and leaves
`cmd.WaitDelay` at zero. The context kill terminates **git only**. A real
credential helper is git's child: it inherits git's stderr, which here is a pipe
back to the manager (both `cmd.Stdout` and `cmd.Stderr` are non-`*os.File`
writers, so os/exec wires pipes with copy goroutines). When the helper hangs —
the locked-keychain / no-desktop-session case — git blocks waiting on it, the
timeout kills git, the orphaned helper keeps the stderr write end open, and with
`WaitDelay == 0` `cmd.Run()` blocks until the helper exits. Potentially forever.
`ReadHost`, `ReadScoped`, `StoreScoped`, `DeleteScoped` and `Discover` (once per
scope) all sit on this path, so one wedged helper wedges the manager — the exact
"blocking a run on a dialog nobody is watching" outcome this package exists to
prevent, except it cannot even be dismissed.

**Confirmed by repro** (`.temp/TASK-260825-1tgpcn/waitdelay-repro/main.go`):
a stand-in git that spawns a 20s grandchild holding inherited stderr, run with
the exact `call` exec pattern and a 500ms context timeout:

```
waitDelay=false elapsed=20.3s err=signal: killed   <- hang lasts the grandchild's lifetime
waitDelay=true  elapsed=1.5s  err=signal: killed   <- bounded
```

The existing `TestACallIsBounded` does not catch this because `modeHang` hangs
the stand-in **git itself**; killing it closes its pipes directly. The unbounded
case needs the hanging process to be a *grandchild* that outlives the kill.

## Requested change (precise, minimal)

1. In `Access.call`, after constructing the command, set `cmd.WaitDelay` (the
   resolved `timeout`, or a small constant like 2s — either is fine; the point
   is that `Run` returns within a bound after the kill instead of waiting on
   pipes an orphaned helper holds).
2. Extend the stand-in git with a hang mode that spawns a grandchild which
   inherits stderr and sleeps past the deadline while the fake git waits on it
   (re-exec the test binary once more, or start it detached), and assert the
   elapsed bound the way `TestACallIsBounded` already does. This pins the fix so
   a future refactor to `exec.CommandContext` alone cannot regress it silently.

No other change is requested. Do not touch the environment construction, the
refusal guards, the read-back, or the message text — all verified correct.

## Reviewer evidence

- Repro source + module: `.temp/TASK-260825-1tgpcn/waitdelay-repro/`
- Reviewer validation commands and results as listed above, run on this host
  (darwin/arm64, go1.25.5, git resolved from PATH).
