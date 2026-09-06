# TASK-260906-1uf713 rework report 5: the silent reinstall retry, the reissued bound, three coverage minors

Head: `8b8aa041` on `feat/agent-environments-stage-c` (3 signed commits on `71e6baec`).
Base: curator main `b056e5da`. Authority: curator-spec `550579d`.
Worktree: `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c`. No push, PR #61 untouched.

Commits (each signed `G`, human identity, staged with named paths only, `git status --short` clean at the end):

- `f2212d79` Stage (c) rework C5-M1: reinstall honours --use/--takeover with run()-level combos
  (`internal/envprofile/envprofile.go`, `cmd/curator/import_cli_test.go`, cmd ledger row)
- `743fd842` Stage (c) rework C5-m1+m4: bare/non-current pin and four blocking-gate tests
  (`internal/envprofile/pathkind_test.go`, 5 ledger rows; also carries pathkind's one-line skip prefix)
- `8b8aa041` Stage (c) rework C5-m2/m3+ledger: classified skips and reissued path-overlay bound
  (`internal/envprofile/import_test.go`, bound reword of rows 303-304)

Per-commit verification (stash-checked, each on its own tree): `go build ./...` exit 0,
targeted `go test -run` exit 0, `ledger-consistency.sh` exit 0 (207 rows at C1, 212 at C2/C3).
Full gates below ran on content byte-identical to `8b8aa041` (the ledger file was rebuilt
incrementally and diffed equal to the lane-run copy before committing).

## Finding → resolution (all five)

### C5-M1 (major) — the retry the §9.5 stop invites was a silent no-op reporting success

Agreed, including the `repeat-of: cycle-1 M1 / cycle-2 C2-m2` (false comment) and
`repeat-of: cycle-4 C4-B1` (success for no work) classifications. Reproduced first on stock
`71e6baec` behavior via the narrowing mutant below, which restores exactly the old
identical-lock branch.

Decision from §9.1 and §9.5: **honour the flags.** `cli/curator.md` accepts `[--use]
[--takeover]` on the install row unconditionally; §9.1 activates on install when the machine
has no current or `--use` is passed; §9.5 says takeover "is carried by a mutating operation
and covers only the specific unmanaged files that carrying operation would write". Refusing
the combination would break idempotent retry; dropping it silently is what made C5-M1. So a
same-source path reinstall now:

- activates through the `useLocked` seam when the machine records no current (first-install
  semantics) or `--use` is passed and the root is not current — with the operation's
  takeover flag and under the locked `require_current_profile` gate, like every other
  machine-scope switch;
- otherwise re-materializes scopes already on the profile via `resyncCurrentScopes`,
  exactly as an update does (a `--takeover` without `--use` activates nothing by itself,
  since a reinstall that switches nothing writes no native surface);
- always reports `updated` (a reinstall is reported as an update); `activated` reports
  whether the install row's activation ran.

The corrected doc comment (`envprofile.go`, replacing the false
"`--use` … takes the fresh-install path, not this one") now reads:

> A reinstall is still the install row (cli/curator.md), so it honours the install row's
> flags: --use activates the installed root when it is not current (a machine with no
> current activates the same way a first install does), and --takeover covers the unmanaged
> files the activation or the resync would write (environments §9.5). Both ride the
> useLocked seam, so the locked require_current_profile gate applies here exactly as on
> every other machine-scope switch. Without --use and with a current recorded, a reinstall
> only re-materializes scopes already on the profile, exactly as an update does. The
> reinstall always reports updated, never a fresh install: activated reports whether the
> install row's activation ran.

The `resyncScopedScopes` split is behavior-identical refactoring: the scoped half of
`resyncCurrentScopes`, so a reinstall activation (which switches the machine scope itself)
still converges scoped currents without materializing the machine scope twice.

All four combinations through `run()`, on a non-current path root with an unmanaged
blocker (`TestProfileInstallReinstallHonoursUseAndTakeover`, 5 subtests, all passing):

- `--use` alone → exit 1 naming `environment_surface_unmanaged_conflict`, current stays
  `other`, no backup (the switch was attempted and failed loudly, not dropped);
- `--takeover` alone → exit 0 `updated profile tk`, current stays `other`, unmanaged bytes
  untouched (nothing to take over without a switch);
- neither → exit 0 `updated profile tk`, current stays `other`;
- `--use --takeover` → exit 0, replace notice naming `.agent-environment-backup`,
  current moves to `tk`, backup generation 1 holds the operator bytes, `CLAUDE.md`
  materialized from the store;
- plus a current-root reinstall with an edited source under `--use --takeover` → exit 0
  `updated profile tk`, still current, edited bytes re-materialized.

Narrowing mutant (identical-lock branch drops the activation, restoring stock behavior for
exactly the already-installed case): the `use with takeover` subtest FAILS at
`import_cli_test.go:249` (no replace notice; current stays `other`, no backup). Killed.

### C5-m1 (minor) — the C4-B1 regression test pinned only the `--as` addressing mode

Closed with `TestPathReinstallBareAndNonCurrentMovesPin` (Install level): a bare
`Install(home, {Operand: source})` reinstall moves the pin, and a reinstall of a
non-current profile moves its pin without switching the machine.

- my1 (`if isPath` → `if isPath && options.As != ""`): FAILS —
  `bare reinstall of the edited tree pinned the same hash: the §1 reinstall is dead
  through the default form`. Killed.
- my4 (`if isPath` → `if machine, _ := Current(home); isPath && machine == name`):
  FAILS — `non-current reinstall of the edited tree pinned the same hash: the §1
  reinstall is dead for a non-current profile`. Killed.

### C5-m2 (minor) — two stage-(c) skip reasons still unregistered

Fixed by spelling them like their classifying siblings (no registry change, no silenced
case, no artefact lie):

- `internal/envprofile/import_test.go:781` (marker file): now
  `this environment can read a mode-000 file: chmod refused: %v`;
- `internal/envprofile/pathkind_test.go:359` (source directory): now
  `this environment can read a mode-000 directory: chmod refused: %v`
  (this line rode in commit `743fd842` with the rest of that file).

Enumeration: every `t.Skip` in the twelve stage-touched test files (15 sites) run through
an awk port of `platform-case-gate.sh`'s own `classify()` against the amended registry:

- 12 classify (`host-capability` mode-000/symlink/case-folding, `root-unset`
  conformance root) — including both fixed lines, verified `host-capability`;
- bare `chmod refused` no longer occurs anywhere;
- 3 remaining UNCLASSIFIED are all `no git on PATH` (cmd `envconfig_test.go:496`,
  `profile_test.go:333`, `internal/envprofile/envprofile_test.go:428`) — pre-existing on
  `origin/main`, guarding git-fixture branches dead on all three runners (git is on PATH
  there), out of scope as cycle 5 already ruled.

No new test in this rework introduces a skip.

### C5-m3 — the AC's "against curator-spec main" clause; the bound retired and reissued

Per the orchestrator's decision: the M1 bound is kept but reissued. The spec contradiction
is gone (curator-spec `bd39adb` makes the overlay form git-only); what survives is an
implementation gap (this tree still enforces the `550579d` form-on-every-overlay rule).
Ledger rows 303–304 reworded to say exactly that, naming TASK-260906-19gjyw (which exists,
status `backlog`) instead of the spec task TASK-260906-3x0w4y. No production code changed
for this: `compose.go`'s form rule, the config reader, and `cmdComposeList`'s dead
`default: form = "path"` branch still implement the task authority `550579d`, truthfully
described by the reworded rows. Nothing implemented here; nothing claimed as driven.

Candidate-root check (curator-spec main `87a0d00`, `CI_REQUIRE_FULL_ROOT=1`,
`go test -count=1 ./internal/config/`): exit 1 with exactly the same 30 failing subcases
as cycle 5 (`TestManagerConfigV2SchemaCases`, `TestManagerConfigV2Vectors`, all `bd39adb`
cases), no more and no fewer. The gap is unchanged and still owned by TASK-260906-19gjyw.

### C5-m4 (minor) — the copied blocking-audit gate gets its mutants; reachability answered

Four new tests (Install/UpdateWithPolicy level, machine path overlays declared directly in
`Policy`, which is how the resolution machinery is reachable in tests):

- `TestUpdateBlocksSecretOverlayMember` / `TestReinstallBlocksSecretOverlayMember`: a new
  overlay member carrying secret material (assembled at runtime via `directorySecret()`,
  no literal) is refused with `profile_update_blocked` naming the blocking finding, old
  lock stands;
- `TestUpdateBlocksRevokedOverlayMember` / `TestReinstallBlocksRevokedOverlayMember`: a
  clean overlay revoked by content hash is refused with `profile_update_blocked` naming
  the revocation, old lock stands. (A path member carries no network identity — its lock
  source is empty, verified by driving it — so the revocation names the snapshot's
  content hash, the production mechanism for local sources per §9.1.)

Mutant table (each keeps the gate, narrows exactly one member kind; `envprofile` suite):

| mutant | result |
|---|---|
| R-g1: reinstall `report.Blocking()` → `&& resolved.Kind != KindContext` | KILLED — `TestReinstallBlocksSecretOverlayMember` FAILs (`profile_source_invalid` naming `context-secret-material`, want `profile_update_blocked`) |
| U-g1: update twin of the above | KILLED — `TestUpdateBlocksSecretOverlayMember` FAILs (same shape) |
| R-g2: reinstall `strictAuditMember` error ignored for `KindContext` | KILLED — `TestReinstallBlocksRevokedOverlayMember` FAILs |
| U-g2: update twin of the above | KILLED — `TestUpdateBlocksRevokedOverlayMember` FAILs (`profile_source_invalid` naming the revoked content hash, want `profile_update_blocked`) |

Reachability answer: **yes, a blocking-but-not-strict finding is reachable** — a revoked
but clean member passes the deterministic detector and is refused by the strict member
audit, proven on stock code by both revoked tests. And neither narrowing is silently
exploitable: when the lock moves, `auditAndStore` re-audits every member, so a narrowed
new-member gate still ends refused (under `profile_source_invalid`); the committed tests
pin refusal at the correct gate with the correct diagnostic. On the structural question:
still no shared helper — the sites read three different manager-owned shapes (ledger JSON,
home marker, store snapshot, overlay audit) — so the class remains closed by vigilance
plus per-site narrowing mutants, now seven pins (P2/P3/P4 carried, arms untouched by this
delta; four new here). Rework-4's P2/P3/P4 mutants were not re-run; their `updateLocked`
arms are byte-unchanged by this delta and the suites covering them are green.

## Gate table (exact standalone commands, observed exit codes)

All in `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c` on the final tree
(`8b8aa041`, clean). No `tee`, no pipes hiding status. The two `test-gate.sh` lanes ran
sequentially, never concurrently and never alongside a `-race` suite.

- `go build ./...` → exit 0, empty output.
- `go vet ./...` → exit 0, empty output.
- `gofmt -l cmd internal` → exit 0, empty output.
- `golangci-lint run ./...` → exit 0, `0 issues.`
- `bash .github/ci/gate-selftest.sh` → exit 0, `gate-selftest: 94 passed, 0 failed`.
- `bash .github/ci/no-broad-suppression.sh` → exit 0, `no-broad-suppression: ok`.
- `bash .github/ci/ledger-consistency.sh /tmp/ledger-final` → exit 0,
  `212 rows checked across linux darwin windows`, `ok` (206 carried + 6 new).
- Default lane `CURATOR_CONFORMANCE_ROOT=/tmp/rework5-spec-pin/conformance/v1 bash .github/ci/test-gate.sh /tmp/rework5-evidence/lane-default` (root materialized via `git -C <curator-spec> archive 0ed5c691… conformance/v1 | tar -x -C …`, archive exit 0) → exit 0. Verdict: `test-gate: go test exit=0, platform-case gate exit=0`. `suite-plan: served=69 deferred=3 excluded=0`. Platform-case gate: 32 skips, `ok`.
- Authority lane `CURATOR_CONFORMANCE_ROOT=/tmp/spec-550579d-wt/conformance/v1 CI_REQUIRE_FULL_ROOT=1 bash .github/ci/test-gate.sh /tmp/rework5-evidence/lane-authority` (real checkout at `550579d` via `git worktree add --detach`, not an archive, per the export-subst lesson) → exit 0. Verdict: `test-gate: go test exit=0, platform-case gate exit=0`. `suite-plan: served=72 deferred=0 excluded=0`. Platform-case gate: 20 skips, `ok`. All six new tests observed passing in both lanes' `observed-cases.tsv` (11 lines incl. subtests).
- Candidate root `CURATOR_CONFORMANCE_ROOT=<curator-spec main 87a0d00> CI_REQUIRE_FULL_ROOT=1 go test -count=1 ./internal/config/` → exit 1, exactly 30 failing subcases, all `bd39adb` — the known TASK-260906-19gjyw consumption gap (see C5-m3).
- `go test -count=1 -timeout 30m ./cmd/curator` → exit 0, `ok … 310.113s` (run solo after all lanes).
- Hosted CI: not consulted — nothing pushed, PR #61 untouched per the brief, so no `gh pr checks` output exists for this delta. Stated plainly.

## Bounds (honest)

- Path overlays: reissued implementation-consumption gap, TASK-260906-19gjyw (was: spec contradiction, TASK-260906-3x0w4y). Ledger rows 303–304 reworded; `cmdComposeList`'s dead `default: form = "path"` branch unchanged and still named here.
- POSIX dotfile list: carried (TASK-260906-vlrjo1); case runs unskipped on all runners.
- Git same-source reinstall still delegates to `updateLocked` with no `--use`/`--takeover` activation handling — pre-existing `origin/main` shape (cycle 5 reported it as code reading), undrivable hermetically here, and this rework was scoped to path roots. Left unchanged, stated.
- C4-m2 resync-after-publish on an already-violating machine: carried documented answer (fail loudly, recover by switching).
- No shared §8.4 helper: vigilance + per-site mutants (seven pins now).
- `purgeHomes` same-shape site on `origin/main` (cycle-5 observation): pre-existing, out of delta, untouched.
- Store missing-branch TOCTOU, `requires`-never-path, secondary-target divergent comparison, always-strict import audit, hard-link discipline, opencode global skills / ledgered entries (§9.4), no pre-recorded import consent, fold-collision skips: carried unchanged from rework 4.
- AC ratio: 15 of 15 stage (c) rows driven through the production entry point. The two rows cycle 5 read as gaps are now pinned: the path row's default addressing mode and non-current reinstall (C5-m1 test) and the `profile install --takeover` retry after the §9.5 stop (C5-M1 test).
