# Producer brief: stage (a) core — rework 4 (F13, F14, F15)

Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-a-core`, branch
`feat/agent-environments-stage-a`, head `dea3f5ac`. Findings:
`TASK-260905-30zs8t_review-findings-stage-a-4.md` — F10/F11/F12 verified fixed and held; two blocking
and one major remain. Read them in full; each carries the reviewer's reproduction script. New signed
commits on top of `dea3f5ac`, no rewrite.

## Author decisions
- **F14 (blocking)** — classify the operand **syntactically**, never by probing the filesystem:
  `/`, `./`, `../` and the platform absolute spellings (Windows drive-letter and UNC prefixes) are
  `path`; everything else is `git`. A directory in the operator's working directory must not be able to
  shadow a git identity and install local bytes with the network allowlist bypassed. Regression cover:
  (1) an operand whose spelling is a git identity but which also names an existing directory resolves
  as `git` and is refused by a locked allowlist; (2) `/tmp/<absent>` and `./<absent>` with `--range`
  produce `profile_install_ref_conflict` with the right diagnostic, not a git-shaped error.
- **F13 (blocking)** — author's call, and the call is: **make install activation perform the §9.2
  switch**. `profile install --use` (and first-install activation) attempts every entry, collects
  per-adapter results, records the new current **only when the whole scope materialized**, and reports
  `profile_use_partial` when it did not. Do not keep the silent pointer move with an "activated" claim.
  Regression cover through the CLI: `--use` on a machine already current on another profile leaves the
  marker and the materialized bytes in agreement; a forced per-adapter failure leaves the previous
  current and reports the partial result. Note the reviewer's observation that `--use` has zero
  coverage anywhere today — this is the test that closes that hole.
- **F15 (major)** — the CLI policy threading the F10 fix depends on is pinned for `profile list` only,
  so dropping it from `use`, `sync` and `update` leaves the suite green. Extend
  `TestProfileListMigrationHonoursSystemPolicy` into a table over `list`, `use`, `sync` and `update`
  (each on a fresh isolated home with a declared global skill and a system config that forbids it),
  then re-run the M-D mutant and record the killed test in the table.
- The reviewer's non-blocking note on `loadMachinePolicy` returning an empty policy when
  `config.UserPath()` does not exist: judge it against the absence-vs-unreadable rule and either fix it
  or state the bound explicitly in the package doc and the report.

## Gates and delivery
The full set as before: `go build ./...`, `go vet ./...`, `gofmt -l`, `golangci-lint run ./...` if
installed, `go test -count=1 -race` on the touched packages, the vector families with
`CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator-spec/conformance/v1`,
`bash .github/ci/gate-selftest.sh`, the platform-case gate for the three GOOS values, and
`go test -count=1 -timeout 30m ./cmd/curator` once at the end. Every fix carries a **narrowing** mutant
that kills a named test. Signed commits; do not push, tag, or open a PR. Attach
`TASK-260905-30zs8t_rework-report-4.md`; `task-board handoff TASK-260905-30zs8t --role developer`.
Never write LOGBOOK.md or anything into the control root.
