# TASK-260822-96m5pj — config build_ssh field

## What landed

`internal/config/buildssh.go` (new, 261 lines) + `internal/config/buildssh_test.go` (new,
13 tests) + wiring in `internal/config/config.go`.

### Grammar (fail closed)

`build_ssh` is a top-level object mapping a **scope** to a credential entry.

A scope is a segment prefix of the canonical `host/path` repository identity of
spec section 6.3:

- host: lowercase DNS name, `^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?(?:\.…)*$`, max 253 bytes
- optional `/`-separated path segments matching `[A-Za-z0-9._-]+`
- path case is preserved (section 6.3 keeps the path case sensitive), so
  `git.example.com/RELUX` is a legal scope while `GIT.EXAMPLE.COM` is not
- rejected: empty scope, empty segment (`h//p`), trailing slash, leading slash,
  `.`/`..` segments, uppercase host, `git..example.com`, `git.example.com.`,
  leading `-`, explicit port, a scheme, an scp-style `user@host:path`, > 4096 runes

An entry carries only `agent`, `identity`, `known_hosts`; any other key is
reported as `has unsupported field(s): …`. Each value must be a string that is
absolute or starts with `~/` (Windows `C:\`, `C:/`, `\\host\share` accepted,
platform-independently) and carries no control character. A present-but-empty
string is a fault, not an absent field — dropping it silently would hand the
repository the ambient SSH state the operator meant to replace. At least one of
`agent`/`identity` must be set, so a `known_hosts`-only entry is rejected.
`build_ssh` absent, `null`, or `{}` all parse to no credentials.

Scope faults are reported in lexicographic scope order, so a config with several
bad scopes always fails the same way rather than following Go's map iteration.

### API

| symbol | purpose |
|---|---|
| `BuildSSHCredential{Scope, Agent, Identity, KnownHosts}` | one operator selection |
| `(*Config).BuildSSHFor(canonical) (BuildSSHCredential, bool)` | longest segment-prefix match |
| `(*Config).BuildSSHScopes() []string` | sorted scopes, for `config build-ssh list` |
| `BuildSSHObject(map) map[string]any` | serialization back to the config JSON shape |
| `ValidBuildSSHScope(string) bool`, `ValidateBuildSSH(cred) error` | standalone validators for writers |
| `(BuildSSHCredential).Expanded()`, `.Empty()` | `~/` resolution; "selects nothing" test |

`managerKeys` gained `build_ssh`, so a system config may ship it as a default
(`TestBuildSSHIsAManagerKey` covers that path). It was deliberately **not** added
to `LockableKeys`: locking operator credentials from an org config is a policy
decision nobody asked for.

### Matching

`BuildSSHFor` reuses `identity.MatchesPrefix` — the same segment-aware matcher the
source allowlist uses — instead of hand-rolling a second prefix rule. It first
requires `identity.ValidCanonical`, so a local source (empty identity), a bare
host, an uppercase host, an `ssh://` URL, or a path with `..` selects nothing and
the caller must fail closed rather than fall back to ambient SSH state.

Boundary case from the AC, with scopes `git.example.com`,
`git.example.com/relux-works`, `git.example.com/relux-works/portals`:

| identity | selected scope |
|---|---|
| `…/relux-works/portals` | `…/relux-works/portals` |
| `…/relux-works/portals/sub` | `…/relux-works/portals` |
| `…/relux-works/portals-evil` | `…/relux-works` |
| `…/relux-works/portals.evil` | `…/relux-works` |
| `…/relux-works-evil/portals` | `git.example.com` |
| `git.example.community/relux-works/portals` | none |

## Verification (real exit codes)

| command | exit |
|---|---|
| `go build ./...` | 0 |
| `go test ./internal/config/ -count=1` | 0 (13 build_ssh tests, all PASS) |
| `go test ./internal/config/ -count=1 -race` | 0 |
| `go test ./...` | 0 (31 packages ok, 0 FAIL) |
| `go vet ./...` | 0 |
| `golangci-lint run` | 0 issues |
| `gofmt -l cmd internal` | 0, no output |

`go tool cover -func`: **100.0% of statements in every `buildssh.go` function**
(`parseBuildSSH`, `ValidBuildSSHScope`, `ValidateBuildSSH`, `BuildSSHFor`,
`BuildSSHScopes`, `BuildSSHObject`, `validCredentialPath`, `Empty`, `Expanded`).

Full-suite log: `.temp/TASK-260822-96m5pj/full-test-01.log`.

## Notes for the next task (TASK-260822-2505vo)

- `Expanded()` covers the "identity/known_hosts values from config expand `~`"
  requirement; it expands `Agent` too, since an agent socket path is equally a
  filesystem path an operator may spell with `~/`.
- All three authentication shapes the sibling AC names are representable and
  tested: identity-only, agent-only, and agent pinned to one identity.
- `BuildSSHFor` returning `false` is exactly the fail-closed trigger for
  `build_repository_ssh_credential_missing`; it never falls back.
- No naming of any other manager implementation appears in code, tests, or this
  artifact, per the epic's naming policy.

## Anomaly

`internal/config/buildssh.go` was rewritten under this session twice while the
task was in flight (three distinct designs within minutes), and `config.go`
ended up with a duplicated `buildSSH, err := parseBuildSSH(...)` statement — the
same edit applied twice — which does not compile. That was reconciled onto the
current design and the tree builds clean. If two producers were dispatched for
this task, only one should keep writing.
