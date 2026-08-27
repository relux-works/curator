# TASK-260825-2gyhq8 review verdict: ACCEPTED

Reviewer run RUN-260824-09ae2e (the prior reviewer run RUN-260824-2f4cf5 completed
its code reading but ended before its own test runs finished, so no verdict was
recorded; this run re-verified everything first-person).

## Verdict

**accepted → done.** All four acceptance criteria are met, verified against the
code, the tests, and the real spec text, with every gate re-run by the reviewer.

## AC verification

1. **add, login, list, remove behave with validation failures covered.**
   - `add` selects exactly one source via `pickBuildHTTPSSource` (zero or
     several → usage error), validates the credential before the config file is
     read, and replaces whole entries. Invalid-invocation matrix pinned in
     `TestConfigBuildHTTPSAddRejectsInvalidInvocations` (no scope, two scopes,
     uppercase host, no source, two sources, source+env, invalid env name,
     unknown flag), with a config-file-unchanged assertion.
   - `login` validates the scope and loads the config *before* the token
     prompt, stores through `gitcred.Access.StoreScoped` (proved by read-back),
     then selects `token=keyring`. Full round trip against a real isolated
     `credential.helper=store` in
     `TestConfigBuildHTTPSLoginStoresThroughTheOperatorHelperAndSelectsIt`,
     including second-login-replaces and remove-drops-both.
   - `list` prints one sorted line per scope with a live `present=` probe
     (`ReadScoped`/`ReadHost`/`os.Getenv` per source); empty set goes to
     stderr so a parser sees empty stdout.
   - `remove` errors on an unconfigured scope, deletes the stored token only
     for a keyring scope, and never touches the operator's own git-credential
     (pinned by `TestConfigBuildHTTPSRemoveNeverTouchesTheOperatorsOwnGitCredential`).
2. **No token reaches argv.** No flag on the build-https surface carries a
   literal secret: `add` takes only enumerated source flags and an env var
   *name*; `login` reads via hidden `term.ReadPassword` when stdin+stderr are
   terminals, else one line from stdin. The config layer additionally rejects a
   literal secret pasted into `token` ("secrets never live in the config").
   The only `--token` flag in main.go is the pre-existing audit-publish
   registry token, untouched.
3. **Help documents precedence and exposure.** Checked against
   curator-spec `protocol/core.md` §12.2 directly: the unbound
   `CURATOR_BUILD_HTTPS_TOKEN` is documented as offered to every private HTTPS
   build host, with `CURATOR_BUILD_HTTPS_HOST` as the required host binding and
   scoped config as the preferred selection; operator-ownership (no manifest/
   descriptor/repository/substitution/marker can select credentials) is stated.
   Ordering (override > scopes; warning after precedence) is test-pinned in
   `TestConfigBuildHTTPSHelpDocumentsPrecedenceAndDisclosure`.
4. **go test green** — see gates below, all reviewer-run.

## Gates (all first-person, 2026-08-25)

| Gate | Result |
| --- | --- |
| `go build ./...` | exit 0 |
| `go vet ./...` | exit 0 |
| `gofmt -l` (cmd/curator, internal/config, internal/gitcred) | clean |
| `golangci-lint run` (same three packages) | 0 issues |
| `go test ./internal/config/... ./internal/gitcred/... -count=1` | ok |
| `go test ./cmd/curator/... -timeout 30m -count=1` | ok, 518.6s |

Log: `.temp/TASK-260825-2gyhq8/review-cmd-curator-test-01.log`.

## Architecture fit

The CLI mirrors the build-ssh command surface one-for-one (dispatch,
`parseInterspersed`, exit-code conventions, symmetric remove);
`SetBuildHTTPS`/`RemoveBuildHTTPS` mirror the build-ssh Set/Remove pair
(validate-before-write, Parse round-trip re-check, atomic write,
last-scope-drops-the-field). It consumes `internal/gitcred` (TASK-260825-1tgpcn)
and `internal/config/buildhttps.go` (TASK-260825-168m7o) rather than
reimplementing either, exactly as the task notes required.

## Minor non-blocking observations (no rework requested)

1. `login --username` is only length-validated at the post-prompt config write
   (>4096 runes fails `SetBuildHTTPS`'s Parse re-check), so in that absurd edge
   the token is already stored in the keyring before the write fails, leaving a
   scope-namespaced orphan entry. Benign (a retry replaces it), but pre-prompt
   username validation would make the doomed-invocation guard complete.
2. A username containing spaces or control characters is accepted into the
   config and rejected fail-closed at broker time (`validHTTPSBrokerField`,
   sibling task's surface). Add-time rejection would be friendlier; no
   security impact.
3. `TestConfigBuildHTTPSAddListRemove` probes the developer's real credential
   helper for the keyring `present=` column (read-only, non-interactive via
   gitcred's environment hardening, deterministic in practice); hermetic
   `buildHTTPSGitHome` isolation there too would be stricter.

## Out-of-scope observation for the orchestrator

LOGBOOK.md entry "0130 — build_https config field lands..." (TASK-260825-168m7o,
already accepted) names the external reference project by absolute path, which
the epic policy forbids in any comment, commit, document, or board artifact.
This task's own entry (0234) is clean. Worth scrubbing before the scope is
committed.

## Commit handling

Reviewer-archetype run: no `commit_ack` supplied. The working tree carries this
task's scope uncommitted (shared checkout with sibling tasks in flight); the
commit-owning mover commits the scope per the board's delivery flow.
