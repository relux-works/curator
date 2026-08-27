# TASK-260822-96m5pj — review verdict: ACCEPTED

Reviewer run `RUN-260822-b910b6`. Read-only; nothing in the working tree was
modified by this review (probe tests ran against a throwaway copy of the source
tree under `.temp/TASK-260822-96m5pj/review/`, since deleted).

## Scope reviewed

- `internal/config/buildssh.go` (new, 291 lines)
- `internal/config/buildssh_test.go` (new, 401 lines, 13 test functions)
- `internal/config/config.go` (`managerKeys`, `Config.BuildSSH`, `Parse` wiring)

`internal/skillspec/parse.go` and `parse_test.go` are also dirty in the tree but
carry no `build_ssh` reference; they predate this task and are out of scope.

## Acceptance criteria — all met

| AC | Evidence |
|----|----------|
| Parse/serialize roundtrip | `TestBuildSSHSerializationRoundtrip` marshals `BuildSSHObject(cfg.BuildSSH)` back to disk and requires `reflect.DeepEqual` on the reparsed map. PASS. |
| Grammar rejection: empty segment | `ValidBuildSSHScope("git.example.com//portals") == false`; `TestParseBuildSSHRejections/empty_segment`. PASS. |
| Grammar rejection: uppercase host | `Git.Example.com` rejected in both the unit table and the parse table. PASS. |
| Boundary bleed `portals` vs `portals-evil` | `TestBuildSSHLongestPrefixMatch`: `.../portals-evil` falls back to the `relux-works` scope, `.../portals.git-mirror` likewise, `git.example.community/...` selects nothing. PASS. |
| Longest-prefix match | Same test: three nested scopes, deepest wins; `TestBuildSSHMatchCarriesCredential` checks the selected credential's payload, not just the scope. PASS. |
| Unknown entry fields rejected | `unsupported field(s): port` and the sorted two-field variant. PASS. |
| `go test ./internal/config` green | 13 build_ssh tests PASS; see `config-buildssh-verbose.log`. |

## Independent verification (real exit codes, this reviewer's runs)

| Command | Exit |
|---------|------|
| `gofmt -l internal cmd` | 0, no output |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `golangci-lint run` | 0 — "0 issues." |
| `go test -count=1 ./internal/config` (13 build_ssh tests) | 0 |
| `go test -race -count=1 ./internal/config` | 0 |
| `go test ./... -count=1` | 0, 31 packages ok, 0 FAIL |
| `go tool cover -func` on `buildssh.go` | 100.0% on all 9 functions |

The attached `TASK-260822-96m5pj_build-ssh-config.patch` was diffed against a
freshly regenerated `git diff internal/config`: identical modulo blob hashes.
The evidence artifact is faithful to the tree.

## Architecture fit

- Matching reuses `identity.MatchesPrefix` (`internal/identity/identity.go:125`),
  the same segment-aware matcher the §6.1 source allowlist uses, behind an
  `identity.ValidCanonical` guard. No second hand-rolled prefix rule was
  introduced — this is the right call, since the two rules must never diverge.
- `BuildSSHFor` returns `(_, false)` for a non-canonical value, a bare host, an
  empty identity, an `ssh://` URL and scp form, so the caller fails closed
  instead of falling back to ambient SSH state. Verified by probe.
- The config write path (`internal/config/write.go`) preserves the raw object,
  so `build_ssh` survives `AddProject`/`Bootstrap` without extra wiring.
- `applySystem` treats `build_ssh` as a whole-key manager default: a system
  config supplies it only when the user config omits the key. Covered by
  `TestBuildSSHIsAManagerKey`.
- Scope stays in this task's lane. CLI (`TASK-260822-3pkc80`), resolution
  (`2505vo`), precheck (`b0wg3a`) and docs (`4p3dcq`) are separate siblings, so
  the absent CLI/docs surface is correct here, not a gap.
- Constraint honoured: no other manager implementation is named in either new
  file (grep for cocoaskills/csk/brew/npm/pip/cargo/asdf/mise returns only the
  `\\.\pipe\openssh-ssh-agent` Windows socket literal).

## Findings — none blocking

1. **LOW — non-deterministic error field when two paths in one entry are bad.**
   `buildssh.go:251` iterates `map[string]*string{"identity":…, "known_hosts":…}`,
   so for `{"agent": true, "identity": "rel/id", "known_hosts": "rel/kh"}` the
   reported field flips run to run. Measured over 200 loads in a scratch copy:
   `identity` 172, `known_hosts` 28. This contradicts the file's own stated
   invariant at `buildssh.go:208` ("Sorted, so a config with several faults
   always reports the same one"); `TestParseBuildSSHReportsOneFaultDeterministically`
   only pins the *scope* ordering, which does hold. Fail-closed behaviour is
   unaffected — such a config is always rejected, only the message varies. Fix
   is a two-element ordered slice instead of a map. Fold into `TASK-260822-3pkc80`
   or `4p3dcq` rather than spending a rework cycle on it.
2. **INFO — `BuildSSHObject` normalizes rather than round-trips an invalid
   struct.** A hand-built `{Agent: false, AgentSocket: "/run/a.sock"}` serializes
   to `"agent": "/run/a.sock"`, which reparses as `Agent: true`. Harmless today:
   the parser can never produce that struct and `ValidateBuildSSH` rejects it.
   `TASK-260822-3pkc80` must call `ValidateBuildSSH` before `BuildSSHObject` so
   the CLI never writes a credential the validator would refuse.
3. **INFO — `build_ssh` is deliberately not in `LockableKeys`.** Documented by
   the implementer as a policy call: an organization cannot lock operator SSH
   credential scopes from a system config. Reasonable default (credentials are
   operator-owned per §12.2), but it is a policy decision worth confirming with
   the spec owner before the story closes.

## Verdict

Accepted → `done`. AC met in full, DoD met, independent verification green on
every gate, design fits the existing identity/config layering. Nothing is
committed — working tree only, per the no-auto-commit rule; the commit-owning
mover still has to commit `internal/config/{buildssh.go,buildssh_test.go,config.go}`.
