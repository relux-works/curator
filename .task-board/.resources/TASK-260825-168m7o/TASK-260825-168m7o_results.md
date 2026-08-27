# TASK-260825-168m7o: build-https-config-field — results

## What shipped

- `internal/config/buildhttps.go` (new): `build_https` field parsing,
  validation, matching, and serialization, mirroring
  `internal/config/buildssh.go`.
- `internal/config/buildssh.go`: extracted the longest segment-aware
  prefix match into a generic `longestScope[T any]` helper so `build_ssh`
  and `build_https` share the exact matching code instead of each keeping
  its own copy. `MatchBuildSSH` now calls it; no behavior change.
- `internal/config/config.go`: added `Config.BuildHTTPS`, wired
  `parseBuildHTTPS(data["build_https"])` into `Parse`, and added
  `build_https` to `managerKeys`. `build_https` is **not** added to
  `LockableKeys` — same rule as `build_ssh`: it is an operator credential
  selection, not an org policy a system config can force.
- `internal/config/buildhttps_test.go` (new): parse/serialize roundtrip,
  literal-secret rejection, exactly-one-source rejection, grammar
  rejections, longest-prefix match including a segment-boundary case,
  manager-key and not-lockable checks, bare-scope-map matching.

## Config shape

```json
"build_https": {
  "git.example.com": {"token": "git-credentials"},
  "git.example.com/relux-works": {"token": "keyring", "username": "oauth2"},
  "ci.example.com/group": {"token_env": "MY_CI_TOKEN"}
}
```

- Scope: exactly the `build_ssh` grammar (`ValidBuildSSHScope`,
  `BuildSSHScopeRule`), reused directly — not restated.
- `token`: one of two enumerated source names, never a literal secret:
  - `TokenSourceGitCredentials` = `"git-credentials"` — the operator's own
    Git HTTPS credential for the scope's host.
  - `TokenSourceKeyring` = `"keyring"` — the manager-namespaced entry a
    future `curator config build-https login` command stores through the
    same Git credential machinery.
  - Any other string is rejected with "...; secrets never live in the
    config".
- `token_env`: names an environment variable (validated as an identifier
  `^[A-Za-z_][A-Za-z0-9_]*$`) read at process entry — the one field that
  can hold operator-chosen text, since it never carries the secret itself.
- Exactly one of `token` or `token_env` is required per entry.
- `username`: optional, defaults to `"token"` via
  `BuildHTTPSCredential.EffectiveUsername()`.
- Unlike `build_ssh`, an unmatched scope is not a config-time or
  resolution-time error: anonymous HTTPS is a legitimate transport.

## Design decisions and why

- **Token source names.** No enum existed anywhere in this codebase yet.
  Per the task-board notes, a sibling manager at
  `/Users/iv/Developer/intranet/cocoaskills` (read-only reference, not
  copied, never named in code/comments/commits) already ships this exact
  surface with `TOKEN_SOURCES = ("git-credentials", "keyring")`. Reused
  those two names so the config surface here composes cleanly with the
  still-unimplemented sibling tasks (`TASK-260825-1lausy` resolution,
  `TASK-260825-1tgpcn` credential reads, `TASK-260825-2gyhq8` CLI) that
  will read the operator's own Git credential vs. the manager-namespaced
  one, matching the two-source language in those task descriptions.
- **Reuse, not duplication, of the SSH grammar/matcher.** Scope validity
  calls `ValidBuildSSHScope` directly. The longest-prefix match is a new
  shared generic helper (`longestScope[T any]`) in `buildssh.go` that both
  `MatchBuildSSH` and `MatchBuildHTTPS` call — one algorithm, two credential
  types.
- **`build_https` is a manager key but not a lockable key**, consistent
  with `build_ssh` and the epic's stated ratified rule: an org-enforced
  system config can supply it as a default but cannot lock the operator's
  own choice.

## Validation run

```
go build ./...                        # PASS
go vet ./...                          # PASS
go test ./internal/config/... -v      # PASS (all build_https + build_ssh + existing config tests)
golangci-lint run ./...               # 0 issues
```

## Downstream tasks this unblocks

- `TASK-260825-1lausy` (per-repository HTTPS resolution) can now read
  `Config.BuildHTTPS` / `Config.BuildHTTPSFor` / `MatchBuildHTTPS`.
- `TASK-260825-2gyhq8` (`curator config build-https` command) can call
  `ValidateBuildHTTPS` / `BuildHTTPSObject` the same way the SSH CLI uses
  `ValidateBuildSSH` / `BuildSSHObject` (no `SetBuildHTTPS` /
  `RemoveBuildHTTPS` write helpers were added yet — out of this task's
  scope; that task will need them, mirroring `internal/config/write.go`).
