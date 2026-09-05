# Producer brief: stage (a) core — rework 1

Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-a-core`, branch
`feat/agent-environments-stage-a`, head `7238412c`. Findings:
`TASK-260905-30zs8t_review-findings-stage-a-1.md` (2 blocking, 2 major, 2 minor, 1 recorded bound) —
read it in full; every finding carries the reviewer's reproduction. New signed commits on top of
`7238412c`, no rewrite. NOTE: the acquisition branch this one is based on was rebased and lands on
curator main as PR #58; when main carries it, `git fetch origin && git rebase -S origin/main` and
prove identity with `git range-diff`.

## Author decisions
- **F1 (blocking)** — `internal/envprofile/envprofile.go:480,545`: the detector must scan the package
  root, not the snapshot root. Apply the same `entry`+`resolved.Directory` join the manifest load
  already uses fifteen lines below (`envprofile.go:559-560`); better, hoist that join into one helper
  used by every consumer of a resolved member so a third call site cannot drift. Add tests that drive
  the production entry points for both shapes the reviewer used: a `--directory` install with a secret
  under the subdirectory, and a transitive `requires.contexts[].directory` member with a secret —
  each must fail installation with the blocking `context-secret-material` finding.
- **F2 (blocking)** — `EnsureDefault` (`envprofile.go:631-654`): create the store entry it pins
  (`contextstore.EnsureState` on the materialized state, not a fresh empty temp dir), and carry the
  migrated global skills as lock members per §9.4 (the skill set of the machine's global scope at
  migration time; direct machine declarations write into the lock). A fresh manager home must make
  `profile list`, `profile use default`, `profile sync`, and `profile use --clear --env <id>` all
  work — test exactly those four on an isolated `CURATOR_HOME`.
- **F3 (major)** — take the manager-home mutation lock and journal the per-entry plan through
  `internal/transaction` and `internal/managerlock`, as §9.2 step 1 requires and as every other
  manager-home mutation in this repository does. If a subset genuinely cannot be journaled in this
  stage, say which and why in the report — but the default is: do it.
- **F4 (major)** — `internal/interop/context_materialization_test.go:145`: `CURATOR_STAGE_B` is read
  nowhere. Either honour it (run the cases when set and let them fail loudly until stage (b)), or drop
  the variable and classify the skip with a registered class that states the truth ("surface not
  implemented until stage (b)") and add that class to `.github/ci/skip-classes.tsv` if it is missing.
  No skip may claim an opt-in that does not exist.
- **F5 (minor)** — reject `>` / `<` on a bare wildcard (`>*`, `<*`, `>x`, `<X`) as
  `profile_source_invalid` rather than treating them as match-everything; node-semver reads them as
  match-nothing and §1.4 admits neither reading. Add the four cases as tests.
- **F6 (minor)** — a root `weights` entry naming a `skill` or `mcp` member: apply rules 1–3 to every
  member kind (§6 says every closure member has one effective weight). If you instead reject such an
  entry, cite the exact sentence that permits the rejection — otherwise apply it.
- **F7** — the reviewer's recorded bound (scoped waivers have no production path) stays a bound; note
  in the report which stage introduces the operator surface for it.

## Gates and delivery
`go build ./...`, `go vet ./...`, `gofmt -l`, `golangci-lint run ./...` if installed,
`go test -count=1 -race` on the touched packages, the vector families with
`CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator-spec/conformance/v1`,
`bash .github/ci/gate-selftest.sh`, and the platform-case gate for the three GOOS values as `ci.yml`
runs it; `go test -count=1 -timeout 30m ./cmd/curator` once at the end. Signed commits; do not push,
tag, or open a PR. Attach `TASK-260905-30zs8t_rework-report-1.md` (finding → disposition, new test
names, gate outputs); `task-board handoff TASK-260905-30zs8t --role developer`. Never write LOGBOOK.md
or anything into the control root; the story workspace carries an empty delta by design.
