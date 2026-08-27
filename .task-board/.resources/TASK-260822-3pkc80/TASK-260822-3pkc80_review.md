# TASK-260822-3pkc80 review — cli-config-build-ssh

Reviewer run `RUN-260822-0dff48`, read-only. Verdict: **accepted**.

## Scope reviewed

Working-tree diff (uncommitted, branch `handoff/cocoaskills-parity-20260731` @ `74fe162`):

- `cmd/curator/main.go` — `cmdConfig` dispatcher, `cmdConfigBuildSSH{Add,List,Remove}`, `buildSSHAgentValue`, `formatBuildSSH`, `configUsage`/`buildSSHUsage`, top-level usage line
- `cmd/curator/main_test.go` — 6 new tests
- `internal/config/write.go` — `SetBuildSSH`, `RemoveBuildSSH`, `buildSSHEntries`
- `internal/config/write_test.go` — 3 new tests
- `internal/config/buildssh.go` — `ValidBuildSSHPath` export (file itself is TASK-260822-96m5pj's, still untracked)
- `LOGBOOK.md` — entry at line 169

`internal/config/config.go`, `internal/closure/*`, `internal/skillspec/*`, `internal/install/install_test.go`,
`internal/skillcheck/*` are dirty from parallel tasks in the same checkout and were excluded from this review.

## Gates re-run independently by the reviewer

Each standalone, real exit code read from `$?` (not `${PIPESTATUS[0]}` — bashism, empty under this zsh):

| gate | result |
| --- | --- |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `gofmt -l cmd internal` | no output |
| `go test -count=1 ./...` | 0, all packages ok |
| `go test -count=1 ./cmd/... ./internal/config/...` | 0 (`cmd/curator` 10.8s, `internal/config` 10.9s) |
| `golangci-lint run` | 0 issues |

NOT run, and not required by this task's AC: `make ci-test` / `check-ci` (need `CURATOR_CONFORMANCE_ROOT`
pointing at a materialised conformance/v1 checkout, unavailable in this session; CI runs them from the pin).

## Behavioural verification (built binary, four scratch configs under `.temp/TASK-260822-3pkc80/`)

Not just the unit tests — the real CLI was exercised against real config files:

- `add <scope> --agent` → `added build_ssh scope git.example.com: agent`, writes `"agent": true`, exit 0
- `add <scope> --agent /run/p.sock --identity ~/.ssh/p` → both recorded in the operator's spelling, exit 0
- `add --agent alpha.example.com` → **bare `--agent` does not swallow the scope**; `alpha.example.com` stays
  positional and the entry is agent-only. The `AcceptsOptionalValue` hook is sound here because the scope
  grammar (lowercase host, no leading `/` or `~/`) and the credential-path grammar are disjoint.
- re-`add` of a recorded scope → `replaced …: identity=/keys/org`, and the previous `agent: true` is **gone**
  from the file, not merged. Confirmed by reading the JSON, not only by the test.
- `remove <unconfigured>` → exit 1, message names the user config path
- `remove <last scope>` → the whole `build_ssh` field is dropped, no `{}` left behind
- `list` → `<scope>\t<fields>`, sorted; with no scopes stdout is empty and the note goes to stderr (exit 0)
- validation failures all exit 2 and leave the file byte-identical: `git.example.com/` (trailing slash),
  `.example.com`, `--known-hosts` alone, relative identity/socket, `--agent=false`, unknown flag
- `config show` marshals `BuildSSH`, so the new field is visible in the effective configuration for free

## AC → evidence

1. **add/list/remove with grammar validation and sorted output** — `cmdConfigBuildSSHAdd` calls
   `config.ValidateBuildSSH` *before* `loadConfig`, so a malformed invocation is exit 2 and is never
   attributed to the config file; `SetBuildSSH` validates again at the library boundary. `list` uses
   `cfg.BuildSSHScopes()` (`sort.Strings`). Verified live above.
2. **help documents precedence flags > env > config scopes** — `buildSSHUsage` states it, and
   `TestConfigHelpDocumentsPrecedenceAndSubcommands` pins the *ordering by string index*, not merely the
   presence of the words. That is the right shape of assertion for a claim about precedence.
3. **CLI tests incl. validation failures** — 13 rejected invocations plus a byte-comparison of the config
   file before/after, so "rejected" is proven to also mean "did not write". `remove` covers missing scope
   (exit 1) separately from malformed invocation (exit 2).
4. **go test green** — re-run by the reviewer, table above.

## Fit with the project

- `SetBuildSSH`/`RemoveBuildSSH` follow `AddProject`'s established shape exactly: `readObject` → mutate →
  `Parse` as a pre-write gate → `writeObjectAtomic` (temp file, 0600, rename). No new write path invented.
- `cmdConfig` replacing the inline `config show` case is a strict improvement — `curator config frobnicate`
  used to fall through to "unknown command \"config\"", now it names the actual mistake.
- The `--agent` optional-value flag reuses the existing `parseInterspersed` `AcceptsOptionalValue` hook
  rather than adding a parser.
- `TASK-260822-96m5pj`'s logbook warning ("the CLI has to call `ValidateBuildSSH` before `BuildSSHObject`
  so it cannot write what the validator refuses" — `BuildSSHObject` normalizes `Agent:false`+socket into
  `"agent": "<socket>"`, which reparses as `Agent:true`) is honoured: `SetBuildSSH` validates first, and
  `TestSetBuildSSHRejectsInvalidCredentialsWithoutWriting`'s "socket no agent" case pins it.

## Non-blocking observations (recorded, not rework)

1. **Live cross-task constraint.** `buildSSHUsage` commits to `CURATOR_BUILD_SSH_*` and a test pins the
   ordering, but no flags/env resolution layer exists yet. `TASK-260822-2505vo` (currently `development`)
   must adopt that prefix or update `buildSSHUsage` in the same change. The prefix matches the story text
   ("CURATOR/CSK_BUILD_SSH_*") and the repo's only env convention (`CURATOR_CONFIG`,
   `CURATOR_SYSTEM_CONFIG`, `CURATOR_REGISTRY_TOKEN`). Note the story's `CSK_` alias is *not* documented in
   the help; if the resolver accepts both, the help needs a second sentence.
2. **System-config layering.** Verified live: a system `build_ssh` shows up in `list`, `remove` of it is
   exit 1 naming the user path (legible), and the operator's first `add` masks the whole system field —
   whole-top-level-key override, identical to `projects`/`allowed_sources`, inherent to Spec §7.2 and owned
   by the grammar task, not this one. `build_ssh` is a manager key but not in `LockableKeys`, so an org can
   supply a default it cannot enforce; worth a glance from `TASK-260822-4p3dcq` when documenting.
3. **`list` output is not space-safe.** `validCredentialPath` permits spaces, so `--identity "/keys/my key"`
   lists as `git.example.com\tidentity=/keys/my key` — the tab separates scope from fields, but the fields
   are space-joined. Human-facing output with `config show` as the machine surface, so acceptable; only
   matters if someone later promises `list` is parseable.
4. **`list` ignores trailing arguments** (`config build-ssh list --json` exits 0 silently) while `remove`
   rejects them. Matches the repo's existing argless-command convention (`cmdList`, `cmdGC`), so consistent
   rather than wrong.
5. **Inherited from `AddProject`:** the pre-write `Parse` gate validates the *user* object alone, so on a
   machine whose user config omits `skills_root` and inherits it from the system config, `add` would fail
   on `skills_root`. Pre-existing pattern, identical in `AddProject`; not introduced here.

## Handoff

Task-level acceptance. Working tree is uncommitted by policy. Reviewer archetype supplies no `commit_ack`;
the commit-owning mover carries this scope into the Story/Bug commit and makes that final transition.
