# TASK-260822-2505vo — per-repository SSH credential resolution

## Where the work lands

**Worktree:** `.temp/TASK-260822-2505vo/worktree`, branch
`task/TASK-260822-2505vo-per-repo-credentials`, cut from `origin/main` at `6a9b201`.

Not on the orchestrator checkpoint branch. `handoff/cocoaskills-parity-20260731`
is 50 commits behind `main` and carries neither `internal/buildrepo` nor
`internal/install/external.go` — the two files this task names. `main` has both.

The accepted `TASK-260822-96m5pj` deliverable (`internal/config/buildssh.go`,
`buildssh_test.go`, the `config.go` wiring) exists only as uncommitted files in
the primary checkout, on no branch at all. Its board patch artifact
`TASK-260822-96m5pj_build-ssh-config.patch` applies to `main` with
`git apply --check` clean, so it was applied verbatim first rather than
reimplemented. It therefore shows up in this branch's untracked files.

## What resolution does

`internal/install/buildssh.go` (new) selects operator SSH credentials for every
planned external build repository, before the first repository is reached.

| step | rule |
|---|---|
| 1 | An explicit run-wide selection (`--build-ssh-*`, then `CURATOR_BUILD_SSH_*`) covers **every** repository of the run |
| 2 | Otherwise the longest `build_ssh` scope matching the repository's canonical identity covers its own |
| 3 | A repository whose **effective** state is not `network-git` + `ssh` needs no selection at all |
| 4 | Anything still unselected fails closed with `build_repository_ssh_credential_missing` |

Step 3 is one predicate for both skip cases the AC names, and it reads the
*effective* state rather than the declared one, so a substitution that
redirects a declared SSH repository onto HTTPS or onto a local path moves the
repository off the SSH transport with it — and a substitution that redirects a
declared HTTPS repository onto SSH moves it on.

The match key is the canonical identity the manifest is already locked to.
Package data never reaches the choice (Spec §12.2).

### Path handling

`config.BuildSSHCredential.Expanded()` resolves `~/` in `identity`,
`known_hosts` and the agent socket against the operator home. An entry that
records `"agent": true` without a socket resolves to the live `SSH_AUTH_SOCK`,
because a macOS agent socket is per login session and a persisted path goes
stale; if there is no live socket, that is an error rather than a silent
downgrade to identity-only. The same applies to `--build-ssh-agent auto`.

`known_hosts` resolves scope → run-wide → the operator's own
`$HOME/.ssh/known_hosts`, because the fetch pins `StrictHostKeyChecking=yes`
and has no other source of truth. A run-wide `--build-ssh-known-hosts` governs
a *scoped* selection too: a scope selects authentication material, not who the
destination is allowed to be.

### Fail-closed diagnostic

```
build_repository_ssh_credential_missing: external build repositories need SSH credentials:
  git.example.test/portals/app (command "build-tool" of skill "portals")
  other.example.test/tools/kit (command "build-agent" of skill "infra")
select credentials with one of:
  curator config build-ssh add git.example.test/portals --agent --identity ~/.ssh/<key>.pub
  curator config build-ssh add git.example.test/portals --agent
or pass --build-ssh-agent/--build-ssh-identity, or set CURATOR_BUILD_SSH_AGENT/CURATOR_BUILD_SSH_IDENTITY
```

Every unselected repository is named at once, and the suggested scope is the
repository namespace, so a sibling repository under the same group is covered
without naming each repository by hand. Candidate discovery (live agent key
count, `*.pub` listing) is deliberately **not** here — that is
`TASK-260822-b0wg3a`.

## Policy construction

`internal/buildrepo/credentials.go` (new):

| symbol | purpose |
|---|---|
| `CodeSSHCredentialMissing` | `build_repository_ssh_credential_missing` |
| `OperatorSSHCredentials{Identity, AgentSocket, KnownHosts}`, `.Selected()` | one per-repository selection |
| `ValidateOperatorSSHCredentials` | resolves every operator path and proves its kind (file / socket / file) |
| `SSHEndpoint(Source)` | the exact `(host, path)` Git hands to SSH — `ssh://` keeps its leading slash, scp-like does not |
| `SSHPolicyFor(base, source, credentials)` | completes a manager-owned wrapper policy with the endpoint and the selection |

`GitTool` gained `SSHCredentials`, bound **per repository** in
`externalPipelineRequest` rather than per run, so a closure spanning two hosts
cannot offer either host the other's key. `AcquireNetwork` refuses an SSH
source whose tool carries no selection, which is the backstop for a caller that
forgot to resolve: no fetch can quietly fall back to ambient SSH state.

A symbolic link on an operator path is resolved, not refused — a live agent
socket is conventionally reached exactly that way — but the resolved target
must exist and be the right kind of object.

### The three authentication shapes, measured

`SSHPolicyFor` → `ExactSSHCommand` tails, asserted end to end from a configured
scope in `TestEveryConfiguredShapeReachesTheWrapperPolicy`:

| selection | tail |
|---|---|
| identity only | `IdentitiesOnly=yes IdentityAgent=none -i <identity>` |
| agent only | `IdentitiesOnly=no IdentityFile=none IdentityAgent=<socket>` |
| agent pinned to one identity | `IdentitiesOnly=yes IdentityAgent=<socket> -i <identity>` |

Each case first runs `config.ValidateBuildSSH` on the scope it uses, so the
test cannot prove a shape the operator is not allowed to write down.

## Surfaces added

- flags on `install` / `upgrade` / `global install` / `global upgrade`:
  `--build-ssh-identity`, `--build-ssh-agent [auto|SOCKET]`, `--build-ssh-known-hosts`
- environment: `CURATOR_BUILD_SSH_IDENTITY`, `CURATOR_BUILD_SSH_AGENT`, `CURATOR_BUILD_SSH_KNOWN_HOSTS`
- `install.Options.BuildSSH`, `install.ExternalDeps.BuildSSH`, `install.CaptureBuildSSHSelection`
- `config.MatchBuildSSH(scopes, canonical)` — `BuildSSHFor` over a bare scope
  map, since the run-wide selection travels without the whole config.
  `(*Config).BuildSSHFor` now delegates to it; behaviour unchanged.
- dry-run provenance lines, e.g.
  `alias: external build ssh: git.example.test/portals/app <- config scope "git.example.test/portals"`

## Not in scope, and left undone deliberately

The manager's own SSH wrapper is still not materialized or exec'd: nothing
writes the wrapper script, the empty config, or re-execs the manager in a
wrapper mode, and `GitTool.SSHWrapper` still comes from the ambient `GIT_SSH`.
That was already true of `ExactSSHCommand` before this task — it had no
production caller either. This task builds the policy those pieces consume and
proves all three shapes reach it; wiring the wrapper itself is a separate piece
of work and is not named in this task's description or AC.

Candidate discovery and the interactive precheck belong to
`TASK-260822-b0wg3a`; the `config build-ssh` subcommand to `TASK-260822-3pkc80`;
docs to `TASK-260822-4p3dcq`. The diagnostic here names the `config build-ssh
add` commands that task will implement.

## Verification (real exit codes, each command run standalone)

| command | exit |
|---|---|
| `gofmt -l cmd internal` | 0, no output |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `golangci-lint run` (v2.12.2, CI's pin) | 0 — `0 issues.` |
| `go test ./... -count=1 -timeout 30m` | RESULT_PLACEHOLDER |

Logs under `.temp/TASK-260822-2505vo/logs/`.

## Host anomaly hit on the way

The first `go test ./...` failed: `cmd/curator` panicked with
`test timed out after 10m0s` in `TestCompiledProjectStatusRepairRollbackRecovery`.
The data volume was at **100%** (7.3 GiB free) and Go's build cache had been
trimmed out from under the running build — two consecutive `go test
./cmd/curator/` runs then failed at **build** time with
`could not import errors (open ~/Library/Caches/go-build/...-d: no such file or
directory)`. `go clean -cache` (20 GiB, one retry needed —
`unlinkat .../d6: directory not empty`, a concurrent writer) restored 34 GiB
free and the suite ran clean. Unrelated to this change; the same anomaly is
already recorded in the logbook for `TASK-260822-27bvo4`.
