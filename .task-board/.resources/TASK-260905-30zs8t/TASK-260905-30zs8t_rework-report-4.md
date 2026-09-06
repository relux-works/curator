# TASK-260905-30zs8t — rework report 4 (F13, F14, F15)

Branch `feat/agent-environments-stage-a`, head `b6f00e1a` on top of the
cycle-4 head `dea3f5ac` (one signed commit, `G oparin@me.com`, 5 files,
+428/−50). No rewrite, no push, no tag, no PR. Worktree
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-a-core`.

Authority: curator-spec main `f39f4a9` (`protocol/environments.md` rev 1.1
§9.1/§9.2, `cli/curator.md`, Decision 0012). Findings:
`TASK-260905-30zs8t_review-findings-stage-a-4.md` (F13/F14 blocking, F15
major) — read in full; each carries the reviewer's reproduction.

## Finding → disposition (author's call per producer-brief-stage-a-rework-4.md)

### F14 (blocking) — syntactic operand classification — FIXED

`isPathOperand` (`internal/envprofile/envprofile.go`) no longer stats the
filesystem. Classification is syntactic per §9.1: `/`, `./`, `../` (plus
`.\`, `..\`, bare `.`/`..`) and the platform absolute spellings
(Windows drive-letter `X:` and UNC `\\`) are `path`; everything else is
`git`. A directory in the operator's working directory can no longer shadow
a git identity and bypass the core §6.1 allowlist (path sources bypass it
by design). A syntactic path that does not exist is still a path: with a
requirement flag or `--directory` it is `profile_install_ref_conflict`
without ever reaching a clone.

Regression cover (all drive production entry points):

- `TestOperandShadowedByDirectoryResolvesAsGit` — plants
  `github.com/evil-org/pkg` (a valid context package) in the cwd, installs
  the operand `github.com/evil-org/pkg` under a locked allowlist
  (`github.com/relux-works`): refused with the allowlist reason, no record,
  no clone. The a9.sh shape.
- `TestAbsentPathOperandWithRequirementIsRefConflict` — absolute-absent and
  `./`-absent operands with a requirement flag are
  `profile_install_ref_conflict` with no clone attempted.
- `TestIsPathOperandIsSyntactic` — classifier table (22 rows): `/`, `./`,
  `../`, `.`, `..`, `.\`, `..\`, `C:/`, `C:\`, `D:`, `\\host\share` are
  path; `github.com/evil-org/pkg`, `https://…`, `git@…`, `ssh://…`,
  bare `pkg` are git.

### F13 (blocking) — install activation performs the §9.2 switch — FIXED

`installLocked` no longer moves the `current` pointer directly. When the
machine has no current profile (first install) or `InstallOptions.Use` is
set, it calls `useLocked` under the same held operation: every entry is
attempted, per-adapter results are collected in `Info.Activation`, the new
current is published through the journal only when the whole scope
materialized, and a partial scope returns `profile_use_partial` with the
current unchanged (the lock is still written; only the activation is
partial). Non-activating installs are unchanged (guidance line).
`cmdProfileInstall` prints the per-adapter `switched` lines on success and,
on a partial failure, the `installed profile <name>` line plus the
per-adapter failures and the `profile_use_partial` error (exit 1) — the
same shape as `profile use`.

Regression cover through the CLI (as briefed) plus library pins:

- `TestProfileInstallUseActivatesThroughSwitch` (CLI): install alpha, then
  install beta `--use`: exit 0, `installed and activated profile beta` plus
  `claude_code: switched`, `profile list` shows beta current, `CLAUDE.md`
  carries `## Context: beta 1.0.0`, marker `profile.name` is beta.
- `TestProfileInstallUsePartialLeavesCurrent` (CLI): install alpha, break
  one adapter home (directory replaced with a file), install beta `--use`:
  exit 1, `profile_use_partial`, `profile list` keeps alpha current and
  beta non-current.
- `TestInstallUseSwitchesAndAgrees` / `TestInstallUsePartialLeavesCurrent`
  (library, `internal/envprofile`): same shapes through `Install`,
  asserting `Info.Activation` (4 results, all OK / one failing) and
  `Current`. These close the `--use` zero-coverage hole at the library
  level too.

### F15 (major) — policy table over all CLI paths — FIXED

`TestProfileListMigrationHonoursSystemPolicy` is now a table over `list`,
`use default`, `sync`, `update default`, each on a fresh isolated home with
a declared global skill and a system config locking `allowed_sources` to
`github.com/relux-works`: all four refuse with `profile_source_invalid`
plus the allowlist reason. M-D re-run below kills the `use`/`sync`
subtests.

### loadMachinePolicy absent-file branch — judged, bound kept

`loadMachinePolicy` returns an empty policy when `config.UserPath()` does
not exist and fails on any other read/parse error. That matches the
absence-vs-unreadable rule: absence (no configuration file) legitimately
means no gates are configured; a failed read is never an empty policy. The
branch is not CLI-reachable (`fileConfigSource.Load` → `config.Load` →
`readObject` fails the command first with "global config not found"), so no
production path reaches a system overlay through it; it is a latent shape
for a future library caller only. The bound is stated on the function doc
(`Production always calls with home == cfg.Home()…`) and carried here.

## Mutant table (each run in a scratch rsync copy, each kills a named test)

| Mutant (gate stays present, admits one class member) | Named test that fails |
|---|---|
| F14: classifier keeps the syntactic rule but re-admits the `os.Stat` directory fallback for non-syntactic operands | `TestOperandShadowedByDirectoryResolvesAsGit` FAILS (`shadowed operand err = <nil>, want profile_source_invalid` — the planted directory installs as path) |
| F13: activation keeps the `activated` claim but drops the switch (pointer-only publish, the pre-fix shape) | `TestInstallUseSwitchesAndAgrees` FAILS (`activation results = 0, want 4`) and `TestInstallUsePartialLeavesCurrent` FAILS (`install beta --use err = <nil>, want profile_use_partial`) |
| F15/M-D: `UseWithPolicy`/`SyncWithPolicy` CLI call sites pass `Policy{}` instead of `PolicyFromConfig(cfg)` (list/update untouched) | `TestProfileListMigrationHonoursSystemPolicy/use-default` and `/sync` FAIL (migration admitted, then a clone of `example.com/skills/hello` is attempted and fails — the allowlist reason is gone); `/list` and `/update-default` still pass, as they keep the policy |

No survivors. The F13 mutant is the brief's narrowing shape taken to its
limit (drop the whole materialization while keeping the claim); a
single-adapter drop is subsumed by it and is killed by the same
marker-and-bytes agreement assertion.

## Gate outputs (real exit codes, standalone processes, this session)

- `go build ./...` — exit 0
- `go vet ./...` — exit 0
- `gofmt -l cmd internal` — exit 0, clean (empty)
- `golangci-lint run ./internal/envprofile/... ./cmd/curator/...` — exit 0, `0 issues.`
- `go test -count=1 -race ./internal/envprofile/` — exit 0, `ok … 28.057s`
- `go test -count=1 -race -run 'TestProfileInstallUse|TestProfileListMigrationHonoursSystemPolicy|TestProfileInstallListUse|TestProfileInstallRefConflict' ./cmd/curator/` — exit 0, `ok … 3.994s`
- `go test -count=1 -race ./internal/contextresolve/ ./internal/identity/ ./internal/audit/` — exit 0, all `ok`
- Vector families with `CURATOR_CONFORMANCE_ROOT=…/curator-spec/conformance/v1`, `go test -count=1 -run 'TestConformance' ./internal/interop/ -v` — exit 0: detectors, header, monolithic, resolution, versions, snapshot-acquisition all PASS; exactly the 7 known stage-deferred sub-skips (3 referenced-* + 4 mcp-*), 0 FAIL
- `bash .github/ci/gate-selftest.sh` — exit 0, `81 passed, 0 failed`
- `bash .github/ci/ledger-consistency.sh .temp/ci-evidence/ledger` — exit 0, `103 rows checked`, `ok`
- `bash .github/ci/no-broad-suppression.sh` — exit 0, `ok`
- `go test -count=1 -timeout 30m ./cmd/curator/` — exit 0, `ok … 270.412s` (full suite, this session)
- Scoped platform-case probe: `go test -count=1 -json ./internal/envprofile/ ./internal/interop/` — test exit 0; `platform-case-gate.sh` on that stream — gate exit 1, expected-scoped: every in-stream case passes and the 7 recorded skips are exactly the stage-deferred set (`stage-deferred`/`tolerated-by-ledger`); the gate's FAIL rows are all `required case never ran` for packages outside the scoped stream. The full matrix remains CI's hosted job, as in prior cycles (the shipped ledger's satisfiability on linux/darwin/windows is covered by gate-selftest, exit 0).

Reran myself in this session: everything above. Accepted from prior evidence: nothing — no gate output is cited from an earlier report.

## Files

- `internal/envprofile/envprofile.go` — syntactic `isPathOperand` (+ `isASCIIDriveLetter`); `Info.Activation`; install activation via `useLocked` with `profile_use_partial`; doc updates
- `internal/envprofile/envprofile_test.go` — `pinHomes` on the previously succeeding path-install tests (first install now materializes, so they must not observe the operator's real homes)
- `internal/envprofile/envprofile_f13f14_test.go` (new) — 2 F13 + 3 F14 tests above
- `cmd/curator/profile.go` — install prints `Info.Activation` per-adapter results; partial failure prints the installed line plus the `profile_use_partial` error
- `cmd/curator/profile_test.go` — policy table (4 subtests) + 2 F13 CLI tests

Signed commit `b6f00e1a` (`G`). No push, no tag, no PR. Nothing written
into the control root. Worktree `git status` holds only these 5 paths.
