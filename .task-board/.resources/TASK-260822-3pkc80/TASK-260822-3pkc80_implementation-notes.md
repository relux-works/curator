# TASK-260822-3pkc80 — cli-config-build-ssh

Implementation notes, handed to review.

## Surface

```
curator config show
curator config build-ssh add <scope> [--agent [SOCKET]] [--identity PATH] [--known-hosts PATH]
curator config build-ssh list
curator config build-ssh remove <scope>
```

`config` became a real dispatcher (`cmdConfig`), replacing the inline
`config show` special case in `run()`. Top-level usage now reads
`config <subcommand>      show | build-ssh (see config build-ssh -h)`.

## Behaviour

- **add** builds a `config.BuildSSHCredential` and validates it with
  `config.ValidateBuildSSH` (the same grammar `parseBuildSSH` enforces) *before*
  reading the config, so a malformed invocation is a usage error (exit 2) and
  never attributed to the config file. It replaces the whole entry recorded
  under the scope — a merge would leave a previous agent selection able to
  authenticate — and prints `added|replaced build_ssh scope <scope>: <fields>`.
- **list** prints one `<scope>\t<fields>` line per scope, sorted by
  `Config.BuildSSHScopes()`. With no scopes configured, stdout stays empty and
  the note goes to stderr, so a caller parsing the listing sees nothing rather
  than a line that names no scope.
- **remove** of a scope that is not recorded in the user config fails (exit 1)
  with a message naming the config path. Removing the last scope drops the
  whole `build_ssh` field rather than leaving `{}`.
- Fields render in the operator's own spelling (`~/...` is not expanded on the
  way in or out); `Expanded()` stays a read-time concern.

## `--agent` optional value

`buildSSHAgentValue` implements `flag.Value` + `IsBoolFlag` +
`AcceptsOptionalValue`, the pattern `parseInterspersed` already uses for
`--audit`. It claims the following token only when it reads as a credential
path, so `add --agent git.example.com` keeps the scope positional while
`add git.example.com --agent /run/a.sock` names a socket. `--agent=false` is
rejected: the config grammar has only the affirmative spelling, since
`"agent": false` would be a second way to write an identity-only entry.

## Exit codes

- 2: malformed invocation (bad scope, relative path, entry that selects
  nothing, wrong positional count, unknown flag/subcommand).
- 1: the operation itself fails (config unreadable/unwritable, removing a
  scope that is not configured).

## New config API (internal/config)

- `SetBuildSSH(path, credential) (replaced bool, err error)` — validates,
  reads the user config, replaces the scope entry, re-`Parse`s the whole
  object, writes atomically. Unrelated fields and other scopes are untouched.
- `RemoveBuildSSH(path, scope) error`.
- `ValidBuildSSHPath(value) bool` — exported so the flag layer can tell an
  optional value apart from a positional without duplicating the path rule.

## Precedence in help

`curator config build-ssh -h` states: command-line flags override
`CURATOR_BUILD_SSH_*` environment values, which override the scopes configured
here; and that credentials are operator-owned (no manifest, descriptor,
repository, substitution, or marker can select them).

**Constraint for TASK-260822-2505vo (per-repository-credential-resolution):**
the help text now commits to the `CURATOR_BUILD_SSH_*` env prefix, matching the
repo's existing `CURATOR_CONFIG` / `CURATOR_SYSTEM_CONFIG` /
`CURATOR_REGISTRY_TOKEN` convention. The resolution task must use that prefix or
update `buildSSHUsage` (a test asserts the precedence ordering in the text).
No flags/env resolution layer exists yet — this task documents the order the
resolver will implement.

## Tests

`cmd/curator/main_test.go`
- `TestConfigBuildSSHAddListRemove` — add three scopes, verify the parsed
  credential, replacement drops the previous agent selection, exact sorted
  listing, remove leaves the other scopes alone.
- `TestConfigBuildSSHListWithoutScopesPrintsNothing`
- `TestConfigBuildSSHAddRejectsInvalidInvocations` — 13 malformed invocations
  (no scope, two scopes, uppercase host, dot segment, empty segment, scheme in
  scope, entry selecting nothing, known-hosts only, relative identity, relative
  socket, relative known-hosts, `--agent=false`, unknown flag), all exit 2, and
  the config file is byte-identical afterwards.
- `TestConfigBuildSSHAgentFlagKeepsScopePositional`
- `TestConfigBuildSSHRemoveRejectsMissingAndMalformedTargets`
- `TestConfigHelpDocumentsPrecedenceAndSubcommands` — asserts flags < env <
  config scopes in the help text.
- `TestConfigSubcommandDispatch`

`internal/config/write_test.go`
- `TestSetBuildSSHRecordsReplacesAndPreservesOtherFields`
- `TestSetBuildSSHRejectsInvalidCredentialsWithoutWriting`
- `TestRemoveBuildSSHDropsScopesAndReportsMissingOnes`

Mutation check: patching `SetBuildSSH` to merge instead of replace turns the
suite red (exit 1), so the replacement claim is actually covered.

## Verification (real exit codes, each command run standalone)

| Command | Exit |
|---|---|
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `gofmt -l cmd internal` (no output) | 0 |
| `go test -count=1 ./...` | 0 |
| `golangci-lint run` | 0 (`0 issues.`) |
| `go build -o bin/curator ./cmd/curator` | 0 |

Not run: `make check-ci` / `make ci-test`, which require
`CURATOR_CONFORMANCE_ROOT` pointing at a materialised
`<curator-spec>/conformance/v1` checkout; no such root is available in this
session. CI runs those gates from the committed pin.

Raw logs: `.temp/TASK-260822-3pkc80/` (`go-test-final.log`, `golangci-02.log`,
`go-vet-01.log`, `mutation-check.log`).
