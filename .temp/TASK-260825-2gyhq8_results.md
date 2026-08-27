# TASK-260825-2gyhq8: config-build-https-command — results

## What changed

- `internal/config/write.go`: added `SetBuildHTTPS`/`RemoveBuildHTTPS`, mirroring
  the existing `SetBuildSSH`/`RemoveBuildSSH` pair (validate-before-write,
  replace-whole-entry, last-scope-drops-the-field).
- `cmd/curator/main.go`: added `curator config build-https add|login|list|remove`,
  alongside the existing `build-ssh` subcommand.
  - `add <scope> (--git-credentials | --keyring | --token-env NAME) [--username NAME]`
    selects exactly one config source; validated before the config file is
    touched, mirroring build-ssh add.
  - `login <scope> [--username NAME]` reads a token via a hidden
    `term.ReadPassword` prompt when both stdin and stderr are real terminals,
    otherwise one line from stdin (never a CLI argument); stores it through
    `gitcred.Access.StoreScoped` (a scope-namespaced entry, verified by
    read-back); then selects `token=keyring` for the scope the same way
    `add --keyring` would.
  - `list` prints one line per scope, sorted, plus a live `present=true/false`
    probe: `ReadScoped` for a keyring source, `ReadHost` for git-credentials,
    `os.Getenv` for token_env.
  - `remove` drops the config scope and, only when it selected `token=keyring`,
    deletes the stored token too (verified); a git-credentials scope is never
    touched, since that credential is the operator's own, not the manager's.
  - `buildHTTPSUsage` documents precedence (`CURATOR_BUILD_HTTPS_TOKEN`,
    optionally host-pinned by `CURATOR_BUILD_HTTPS_HOST`, ahead of configured
    scopes) and the core Spec §12.2 disclosure warning for the identity-unbound
    override.

## Built on top of

The working tree already carried uncommitted sibling work from this story
when this task started: `internal/config/buildhttps.go` (the `build_https`
config field, TASK-260825-168m7o) and `internal/gitcred/` (the credential
read/write machinery, TASK-260825-1tgpcn). This task consumes both rather
than reimplementing them.

## Tests

- `internal/config/write_test.go`: `TestSetBuildHTTPSRecordsReplacesAndPreservesOtherFields`,
  `TestSetBuildHTTPSRejectsInvalidCredentialsWithoutWriting`,
  `TestRemoveBuildHTTPSDropsScopesAndReportsMissingOnes`.
- `cmd/curator/main_test.go`: add/list/remove round trip, empty list, add
  invalid-invocation matrix, remove invalid-invocation matrix, login
  invalid-invocation matrix, a full login→list→remove round trip against a
  real, isolated Git credential store (`credential.helper=store` under a
  temp `HOME`), a second-login-replaces-the-token case, a
  remove-never-touches-the-operator's-own-credential case, and a help-text
  precedence/disclosure test.

## Validation

| Command | Result |
| --- | --- |
| `go build ./...` | exit 0 |
| `go vet ./...` | exit 0 |
| `golangci-lint run ./internal/config/... ./cmd/curator/...` | 0 issues |
| `go test ./internal/config/... ./internal/gitcred/... -v` | ok, all green |
| `go test ./cmd/curator/...` (`-timeout 30m`, ~9m runtime per project memory) | ok, exit 0 |

No token ever reaches argv: `add` only names a source (an enumerated string
or an env var name), and `login` reads the secret from a hidden prompt or
stdin, never a flag or positional argument.
