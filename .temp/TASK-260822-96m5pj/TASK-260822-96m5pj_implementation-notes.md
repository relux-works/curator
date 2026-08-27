# TASK-260822-96m5pj — build_ssh config field

## What landed

New file `internal/config/buildssh.go`, new test file `internal/config/buildssh_test.go`,
three-hunk change to `internal/config/config.go`.

### Surface

```go
type BuildSSHCredential struct {
    Scope       string // config key this credential was recorded under
    Agent       bool   // an SSH agent is selected
    AgentSocket string // explicit socket; empty with Agent set = the operator's own agent socket
    Identity    string // identity file; with Agent set it pins the single key the agent offers
    KnownHosts  string // per-scope known-hosts override
}

func (b BuildSSHCredential) Empty() bool
func (b BuildSSHCredential) Expanded() BuildSSHCredential   // resolves a leading "~/"
func ValidBuildSSHScope(value string) bool
func ValidateBuildSSH(credential BuildSSHCredential) error
func BuildSSHObject(map[string]BuildSSHCredential) map[string]any   // serialization
func (c *Config) BuildSSHFor(canonical string) (BuildSSHCredential, bool)  // longest-prefix match
func (c *Config) BuildSSHScopes() []string                  // sorted
```

`Config.BuildSSH map[string]BuildSSHCredential` and `managerKeys["build_ssh"]`.

### JSON shape

```json
"build_ssh": {
  "git.example.com":                     {"agent": true},
  "git.example.com/relux-works":         {"identity": "~/.ssh/org", "known_hosts": "~/.ssh/known_hosts_org"},
  "git.example.com/relux-works/portals": {"agent": "/run/portals.sock", "identity": "/keys/portals"}
}
```

### Grammar (fail closed)

Scope: lowercase host (label form, `<=253` bytes, no empty label, no trailing dot)
optionally followed by `/`-separated segments of `[A-Za-z0-9._-]+`; no empty, `.`
or `..` segment; `<=4096` scalars. Path case is preserved, per Spec §6.3.

Entry: only `agent`, `identity`, `known_hosts`. Unknown fields are reported
sorted. `agent` is `true` or a socket path — `false`, `null` and other types are
rejected, so identity-only has exactly one spelling. Paths must be absolute or
start with `~/` and carry no control character; present-but-empty and
present-but-non-string are faults, not "absent". At least one of
`agent`/`identity` per entry.

### Matching

`BuildSSHFor` gates on `identity.ValidCanonical` first, so a bare host, an
uppercased identity, a URL, or an SCP remote select nothing. It then reuses
`identity.MatchesPrefix` (Spec §6.1 segment-aware matching), longest scope wins.
Boundary covered by tests: `git.example.com/relux-works/portals` matches the
`portals` scope, while `.../portals-evil` and `.../portals.git-mirror` fall back
to the enclosing `relux-works` scope.

## Decisions worth review

1. **`agent` accepts `true`, not only a socket path.** The sibling CLI task
   specifies `--agent [SOCKET]` with an optional socket, and an agent socket
   path is per-login-session on macOS
   (`/private/tmp/com.apple.launchd.*/Listeners`), so persisting one into a
   machine-global config goes stale after a relogin. `agent: true` records the
   operator's *choice* of the agent and lets resolution read the live socket and
   pass it explicitly as `IdentityAgent=<socket>`. `agent: "<socket>"` still
   parses, so nothing is lost — this shape is a superset of socket-only.
   The three authentication shapes therefore map to:
   identity-only, `agent` alone, `agent` + `identity` (pinned).

2. **Rooted paths only.** A relative `identity`/`known_hosts`/agent socket is
   rejected: a machine-global config has no stable directory to resolve it
   against. Absoluteness is decided by a platform-independent rule (`/`-prefix,
   `X:` drive, or `\\` UNC) so one config file yields the same verdict wherever
   it is read; a Windows named pipe agent socket parses on every platform.

3. **Raw spelling is stored, not the expanded path.** Parsing keeps `~/...`
   verbatim so serialization round-trips byte-for-byte; `Expanded()` is what the
   install path calls (TASK-260822-2505vo).

4. **Scope iteration is sorted before validation**, so a config with several
   faults always reports the same one instead of following Go's map order.

## Concurrency incident

Two spawns ran this task in the same checkout at the same time:
RUN-260822-13cbfb (20:13) and RUN-260822-fad725 (20:14, this run). They
overwrote each other's files mid-edit. This run moved to an isolated worktree at
`.temp/TASK-260822-96m5pj/worktree`, finished and verified there, then compared
against the peer's completed result and landed the merged superset.

The two implementations converged on nearly the same design. The only
substantive fork was decision 1 above (`Agent string` in the peer's version vs
`Agent bool` + `AgentSocket string` here). The peer's test suite contributed the
`identity null` and `known_hosts empty` rejection cases and the
`TestValidateBuildSSHAcceptsEveryParsedCredential` invariant, all folded in.
Both peer files are preserved verbatim next to this note as
`buildssh.peer-run-13cbfb.go.txt` and `buildssh_test.peer-run-13cbfb.go.txt`.

## Verification (all run in the main checkout after landing)

| Command | Exit |
|---|---|
| `gofmt -l internal/config/` | clean (no output) |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `golangci-lint run` | 0 (`0 issues.`) |
| `go test -count=1 ./internal/config/` | 0 |
| `go test -race -count=1 ./internal/config/` | 0 |
| `go test ./...` | 0 (31 packages ok) |

Statement coverage of every new function in `buildssh.go` is 100.0%
(`go tool cover -func`); package total 79.9%.

## Not in this task

No CLI subcommand (TASK-260822-3pkc80), no install-path resolution
(TASK-260822-2505vo), no precheck (TASK-260822-b0wg3a), no docs
(TASK-260822-4p3dcq). `ValidateBuildSSH` and `BuildSSHObject` are exported for
the CLI task to reuse instead of re-deriving the grammar.
