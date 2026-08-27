# TASK-260825-1tgpcn — https-credential-reads

Operator HTTPS credential access for external build repositories, through the
operator's own Git credential machinery.

## What landed

`internal/gitcred` (new package, `gitcred.go` + `gitcred_test.go`).

Every read and write is `git credential fill|approve|reject`. No platform
branch and no platform secret tool: one mechanism that exists identically on
macOS, Windows and Linux, speaking to whichever helper the operator's own Git
configuration already selects (Spec §12.2 keeps that material operator-owned).

### Surface

| Call | Purpose |
| --- | --- |
| `Access.ReadHost(ctx, host)` | the operator's own HTTPS credential for a host |
| `Access.ReadScoped(ctx, scope, host)` | the manager-namespaced credential for a scope |
| `Access.StoreScoped(ctx, scope, host, secret)` | write, proved by reading it back |
| `Access.DeleteScoped(ctx, scope, host)` | remove, proved by reading it back |
| `Access.Discover(ctx, host, scopes)` | presence-only material view for the prompt |
| `NamespaceUsername`, `ScopeHost`, `OperatorHome` | addressing helpers |

`Access` carries the Git executable, the pinned operator home, the base
environment and a per-call timeout; the zero value resolves `git` on PATH and
the home of the account the manager runs as.

### Decisions

- **Namespacing.** A manager entry lives under username
  `curator-build-https:<scope>`. A distinct username is what keeps it a
  separate record from the operator's own credential for the same host, on
  every platform's helper.
- **Prompting is closed off four ways.** `GIT_TERMINAL_PROMPT=0`,
  `GCM_INTERACTIVE=never`, `-c credential.interactive=false`, and
  `-c core.askPass=` with `GIT_ASKPASS`/`SSH_ASKPASS` *removed* from the
  inherited environment — Git reads an empty askpass variable as unset and
  falls through to the next prompt source, so emptying is not enough.
- **Operator home is pinned.** `HOME` and `USERPROFILE` are both overridden
  for the call. The fetch owns a private HOME; without the pin the helper the
  operator configured is simply not found. The rest of the environment is
  inherited, because a helper is a session-bound program (desktop bus,
  keychain session, proxy).
- **A write is never trusted.** `approve` is followed by a read-back; a
  mismatch or an absence raises with the store to configure on the running
  platform, plus the other platforms' stores and the environment-variable
  alternative.
- **A read never fails a run.** Absent material, a helper error, a missing
  Git, a helper that hangs (bounded at 15s) all report "nothing here".
- **A namespaced answer is refused for a host read.** A helper asked for a
  host without a username answers with whatever record it holds — which can be
  the manager's own. Reporting that back as the operator's own credential
  would make one record look like two.
- **A scoped read requires the username it asked for.** A helper free to
  answer a near miss must not hand the operator's own credential back as the
  manager's entry.
- Values carrying a newline, a carriage return or a NUL are refused before
  they reach Git: the credential protocol is newline-delimited, so such a
  value would not be one value.

### Observed helper behaviour worth keeping

Git's built-in `store` helper **prepends** a new record. Once the manager
stores its namespaced entry for a host, a username-less `fill` for that host
answers with the manager's entry rather than the operator's. The host read
refuses that answer, so the outcome is a fail-closed "no operator credential"
rather than a wrong one, and deleting the manager entry restores the operator
read. Pinned by `TestRealGitKeepsTheManagerEntrySeparate`.

## Tests

`internal/gitcred/gitcred_test.go`, 17 tests, no skips on this host.

Two harnesses:

1. A stand-in Git — this test binary re-executed — that serves the credential
   protocol from a JSON store and records the argv, environment and payload of
   every call. It has selectable defects: a Git that cannot answer, a helper
   that approves and persists nothing, a helper that answers a username it was
   not asked about, a helper that accepts a rejection and keeps the record,
   and a helper that never answers. No shell script and no build step, so the
   same tests run on every platform.
2. The real `git` with its built-in `store` helper in a pinned temp home, with
   the machine's system configuration excluded so no test can reach a real
   keychain. This covers the whole surface end to end, and reproduces the
   read-back failure for real: with no helper configured, `git credential
   approve` exits 0 and keeps nothing.

## Validation

| Command | Result |
| --- | --- |
| `go test -count=1 -v ./internal/gitcred/` | ok, 17/17 pass, 0 skipped |
| `go test -race -count=1 ./internal/gitcred/` | ok (18.4s) |
| `go test -count=1 -timeout 30m ./cmd/curator` | ok (439.3s) |
| `go test -count=1 <all other packages>` | ok, no failures |
| `go test -timeout 30m ./...` | exit 0, 42 packages ok, no FAIL |
| `go build ./...` | exit 0 |
| `go vet ./...` | exit 0 |
| `golangci-lint run` | 0 issues |

## Not in this task

The `build_https` config field (TASK-260825-168m7o), per-repository
resolution (TASK-260825-1lausy), the broker and fetch wiring
(TASK-260825-3n4bjj) and the command surface (TASK-260825-2gyhq8). This
package deliberately depends on nothing in the repository, so each of them can
import it without a cycle.

## Interop with the config field landing beside it

`internal/config/buildhttps.go` (TASK-260825-168m7o, landed in the same working
tree while this was written) enumerates exactly two token sources, and they map
one-to-one onto the two reads here:

| `build_https` token source | call |
| --- | --- |
| `git-credentials` | `Access.ReadHost` |
| `keyring` | `Access.ReadScoped` |

`config.BuildHTTPSDefaultUsername` and `gitcred.DefaultUsername` are both
`"token"`. The two packages are deliberately not coupled: this one depends on
nothing, and the resolution layer (TASK-260825-1lausy) is where a config rule
is turned into a call.
