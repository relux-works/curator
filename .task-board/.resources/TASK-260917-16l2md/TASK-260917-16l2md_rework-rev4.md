# Rework brief — TASK-260917-16l2md, revision 4 (rc.12 pin promotion union)

Revision 3 (hosted gate green on every lane at `SPEC_PIN dced9b8`) was
rejected with three corrections (`TASK-260917-16l2md_review-verdict-rev3.md`;
the reviewer's probes `TestReviewerSurfacingBeforePublication` and
`TestReviewerProviderDirectoriesNullRejected` are attached as task outcomes).
The union fidelity, the pin move, E2 and the E4 identity work passed; keep
them byte-identical.

## Corrections (all required)
1. **S4 — print the §2.3 surfacing rows BEFORE publication (High).**
   Environments §2.3 (rc.12) requires the manager to PRINT the MCP
   declaration rows after audit succeeds and before the lock is published or
   any surface materialized; the candidate computes the rows in
   `internal/envprofile` but the CLI prints them only after `Install`/
   `Update` return (lock already published, home already activated; a
   publication failure can even discard the rows). Fix the operation→CLI
   emission seam so the output happens at the required point for install,
   changed-path reinstall and update (before activation/resync), informative
   surfacing stays non-fatal, nothing prints twice; add committed tests that
   OBSERVE runtime order (e.g. a stdout writer that stats the lock when the
   first `mcp-declaration` row arrives; old-lock identity on update; a
   failure injected after the emission point). Replace the proxy
   `runSurfacingOrderCase` coverage with observation of the event order —
   do not weaken the vectors.
2. **E4 — `provider_directories: null` is a type error (Medium).**
   `internal/config/environments.go` accepts an explicit `null` as `[]`; the
   rc.12 `manager-config-v2` schema requires an array (default `[]`, no null
   meaning — unlike `passable_env_names`). Parse every present value and
   reject `null` with the existing wrong-type configuration error naming the
   knob; committed `Load` regressions for user and system configuration
   (omitted and empty list still accepted). Same absence-vs-null class as
   the E2 waiver fix.
3. **Operator docs (Low).** Add concise documentation under `docs/` for the
   three knobs/policies (`transitive_system_modules` + `system_module_waivers`
   with an example, `provider_directories` and the revision-A warning /
   revision-B refusal, `passable_env_names` default `[]` with explicit
   `null` = unbounded and the S4 profiles), migration steps and the
   absent-vs-null semantics; keep the historical audit report as history.

## Validation and handoff
Narrow gates at the rc.12 root (`CURATOR_CONFORMANCE_ROOT=/tmp/spec-rc12/conformance/v1`):
`go build ./... && go vet ./... && gofmt -l internal cmd && go test -count=1 ./internal/config/... ./internal/envprofile/... ./cmd/curator/...`
plus the packages of the earlier rounds; update `TASK-260917-16l2md_results.md`
with a "Revision 4" section (per-correction file:line, transcripts); hand off
with `task-board handoff TASK-260917-16l2md --role developer`. Worktree and
rules unchanged.
