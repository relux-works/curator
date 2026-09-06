# Producer brief: stage (a) core — rework 3 (F10, F11, F12)

Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-a-core`, branch
`feat/agent-environments-stage-a`, head `314ae748`. Findings:
`TASK-260905-30zs8t_review-findings-stage-a-3.md` — F8 and F9 verified fixed and held; one blocking
and two major findings remain. Read them in full; each carries the reviewer's reproduction. New signed
commits on top of `314ae748`, no rewrite.

## Author decisions
- **F10 (blocking)** — the migration path must not re-parse configuration into a weaker copy. Thread
  the already-loaded `Policy` from the CLI (which has the overlaid `cfg`) into
  `EnsureDefault`/`ensureDefault`/`migrateGlobalSkills`. If a load is unavoidable in some path, it goes
  through `config.Load(config.UserPath(), …)` including the **system** layer, and a read or parse
  failure fails the migration — a failed read is never an empty policy (the absence-vs-unreadable rule
  this protocol repeats everywhere). Regression cover as the finding names: an isolated home with
  `CURATOR_SYSTEM_CONFIG` locking `allowed_sources` and `audit.revocations` so that a declared global
  skill is excluded/revoked — the migration is refused; and the mirror case that succeeds when the
  system config permits it.
- **F11 (major)** — the F9 negatives are delete-only, so three weakening mutants survive. Add the three
  narrowing tests the finding specifies: an allowlist entry `example.com/org` against the operand
  `example.com/org-evil/pkg` (same host, adjacent path segment); a revoked `skill` member and a revoked
  `mcp` member; and a canary test that drives the strict-audit member path with the canary forced to
  fail and asserts the install is refused. Re-run each as a narrowing mutant and put the killed test in
  the report's table. Also fix `TestMigratedSkillSourceIsCanonical`: it is named for the canonical
  identity but asserts the raw `file://` shim URL, so it passes under the old `canonicalGit` — make it
  assert the canonical identity.
- **F12 (major)** — a `file://` git operand is reachable from the CLI and writes a lock and a marker
  that fail the published schemas, so "test-only shim" is not true as written. Reject a `file://`
  operand and a `file://` requirement source at the same boundary that rejects a malformed network
  source, with `profile_source_invalid`. The F8 tests already prove the replacement (`insteadOf` gives
  hermetic offline network-identity fixtures) — convert the remaining `file://`-based tests to that
  shape. Only if that conversion genuinely cannot fit this stage: gate the operand at the CLI so no
  operator-reachable path produces a schema-invalid artifact, and say so explicitly — but the default
  is the full rejection.
- The reviewer judged the other two bounds defensible (MCP allowlist awaiting manager-config v2;
  `fetchRaw` first-spelling-wins). Keep them, and add the reviewer's note that a transitive requirement
  declared `git@host:org/dep` is canonicalized before transport selection.

## Gates and delivery
`go build ./...`, `go vet ./...`, `gofmt -l`, `golangci-lint run ./...` if installed,
`go test -count=1 -race` on the touched packages, the vector families with
`CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator-spec/conformance/v1`,
`bash .github/ci/gate-selftest.sh`, the platform-case gate for the three GOOS values, and
`go test -count=1 -timeout 30m ./cmd/curator` once at the end. Every fix carries a **narrowing** mutant
(one that weakens the gate rather than deleting it) that kills a named test — that was the F11 lesson.
Signed commits; do not push, tag, or open a PR. Attach `TASK-260905-30zs8t_rework-report-3.md`;
`task-board handoff TASK-260905-30zs8t --role developer`. Never write LOGBOOK.md or anything into the
control root.
