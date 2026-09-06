# TASK-260906-1uf713 rework report 4: the dead reinstall, the red Windows lane, three undriven refusals

Head: `71e6baec` on `feat/agent-environments-stage-c` (4 signed commits on `4a8a1d42`).
Base: curator main `b056e5da`. Authority: curator-spec `550579d`.
Worktree: `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c`. No push, PR #61 untouched.

Files changed (6, staged with named paths only when committing; `git status --short` read before each step, no `git add -A`):

- `internal/envprofile/envprofile.go` — C4-B1 path-reinstall fix (`reinstallPathLocked`)
- `internal/envprofile/pathkind_test.go` — C4-B1 paired reinstall assertions + C4-M1 three refusal tests
- `internal/envprofile/envprofile_policy_test.go` — C4-m1 builtin-default policy row
- `cmd/curator/envconfig_test.go` — C4-m1 builtin-default run() row
- `.github/ci/skip-classes.tsv` — C4-B2 file-case registry fix
- `.github/ci/platform-cases.tsv` — row 329 wording fix + three new refusal rows

## Finding → resolution (all five)

### C4-B1 (blocking) — the §1 reinstall is dead, and reports success anyway

Agreed, including the `repeat-of: cycle-3 C3-B2` classification and the bisection to `7fcf9c1b`.
`installLocked` routed a same-name, same-source install into `updateLocked` for every kind;
`updateLocked`'s `KindPath` arm resolves from the store snapshot and never from `source.Path`,
so the `stateForPath` result computed eleven lines above was thrown away and the old pin
re-pinned. Exit 0 with `updated profile pk` and the same lock hash.

Fix from §1 (one sentence, two halves): the snapshot is immutable and never re-read, *until
the operator reinstalls* — and `cli/curator.md` defines exactly one row that reads a `path`
source, `profile install <git-url|path>`. A same-source `KindPath` install no longer delegates
to `updateLocked`. It calls the new `reinstallPathLocked`, which mirrors `updateLocked`'s
publish-and-resync shape (new-member blocking checks, lock-hash comparison, lock plus
previous-lock publication, `resyncCurrentScopes`) but resolves the root from the freshly
computed `stateForPath` input already in hand. A `git` reinstall still delegates to
`updateLocked` — for a range requirement a reinstall *is* an update — so the fix is the
`path` case, not the delegation in general.

Paired evidence through `run()`-equivalent production entry points (`Install`,
`UpdateWithPolicy`, `SyncWithPolicy`, `UseWithPolicy` — the calls the CLI rows reach), in the
extended `TestPathSnapshotImmutableAcrossUpdateSyncUse`:

- `update` does NOT move the pin after a source edit (existing assertions, unchanged);
- `sync` and `use` do not move it (existing, unchanged);
- `install <same path> --as <same name>` DOES move the pin, reports
  `activated=false updated=true`, and the re-materialized
  `<claude-home>/CLAUDE.md` carries `EDITED AFTER INSTALL`;
- a fresh install of the edited tree as `pk2` pins the identical hash to the same-name
  reinstall, so the reinstall and fresh paths agree on content.

Narrowing mutant (path reinstall routed through `updateLocked` again via
`if false && isPath`): `TestPathSnapshotImmutableAcrossUpdateSyncUse` FAILS with
`reinstall of the edited tree pinned the same hash: the §1 reinstall is dead` (exit 1).
Doc comment and ledger row 329 corrected to describe both halves (see below).

### C4-B2 (blocking) — the Windows lane is red on an unregistered skip reason

Agreed. `internal/envprofile/import_test.go:785` skips with
`this environment can read a mode-000 file; unreadability is untestable here` (file,
correctly — the artefact is `.agent-environment.json`), while
`.github/ci/skip-classes.tsv` registered only
`this environment can (inspect|read) a mode-000 directory`. One word, red lane since rework 2.

Fix: widened the registry pattern to
`this environment can (inspect|read) a mode-000 (directory|file)` and widened the note to
`the runner ignores mode-000 permissions`. No skip text changed, no case silenced — the gate
was doing its job.

Local classifier reproduction (no Unix lane exercises the skip itself):

```
reason: this environment can read a mode-000 file; unreadability is untestable here
OLD pattern (directory only): no match (reproduces C4-B2 red)
NEW pattern (directory|file): MATCHES -> classified
sibling directory reason: still MATCHES
```

### C4-M1 (major) — the three refusals C3-B2 introduced are undriven

Agreed on all three, including that P4 is §8.4 verbatim. Three new committed tests, each
through production `UpdateWithPolicy`:

- `TestUpdatePathWithoutStatePinIsSourceInvalid`: lock root rewritten to a commit pin (the
  only validated lock shape without a state hash, so `readLock` still validates), update
  refuses `profile_source_invalid` naming `carries no state pin`.
- `TestUpdatePathMissingSnapshotIsSourceInvalid`: store entry under the pin removed, update
  refuses `profile_source_invalid` naming `path snapshot cannot be read`, never silent
  unchanged.
- `TestUpdatePathSnapshotNameMismatchIsSourceInvalid`: snapshot manifest renamed, update
  refuses `profile_source_invalid` naming `snapshot names`.

Mutant table (each keeps the gate, weakens exactly one member; `envprofile` suite):

| mutant | result |
|---|---|
| P2: drop `\|\| rootMember.StateHash == ""` from the no-state-pin refusal | KILLED — `TestUpdatePathWithoutStatePinIsSourceInvalid` FAILs (falls through to the snapshot read, wrong diagnostic) |
| P3: delete the snapshot-name check | KILLED — `TestUpdatePathSnapshotNameMismatchIsSourceInvalid` FAILs (`err = <nil>`) |
| P4: snapshot read failure returns old lock as `unchanged`, nil error (§8.4 shape) | KILLED — `TestUpdatePathMissingSnapshotIsSourceInvalid` FAILs (`err = <nil>`) |
| B1: path reinstall routes through `updateLocked` again (`if false && isPath`) | KILLED — `TestPathSnapshotImmutableAcrossUpdateSyncUse` FAILs (same-hash reinstall) |

Whether anything structural stops a fourth §8.4 site: no. The three sites read three
different manager-owned shapes (adapter ledger JSON, home marker, store snapshot) through
three different helpers, and no shared helper covers all three read shapes. The class is
closed by vigilance plus the per-site narrowing mutants now committed (each failed-read
branch has a named test asserting refusal, and P4's shape is the regression test for the
class). Silence is not the answer; a fourth site gets the same treatment.

### C4-m1 — `Policy.CheckMachineUse` narrowed with `|| name == DefaultProfile`

Covered at both levels. `TestPolicyFromConfigCarriesEnvGates` now asserts
`CheckMachineUse(DefaultProfile)` refuses under a lock to `acme`, and
`TestProfileUseLockedRequireRefuses` (run()) now asserts
`profile use default` exits fail naming `environments.require_current_profile`.
Mutant `|| name == DefaultProfile` dies at both: policy suite FAILs
(`locked require must refuse the builtin default profile: <nil>`, exit 1) and cmd suite
FAILs (`use of the builtin default profile = 0`, exit 1). Production was already correct;
this was coverage, and it is now driven.

### C4-m2 — `profile update` on an already-violating machine moves the lock, then resyncs at exit 1

Answered, not fixed. A machine in a configuration-error state (current != locked requirement
when an update moves) fails loudly: `updateLocked` publishes the new lock, then
`resyncCurrentScopes` hits the §12.2 gate and returns exit 1 with the lock moved and the
surfaces stale. The state is recoverable by switching to the required profile, and the same
shape covers the new `reinstallPathLocked` (publish, then resync). Loud failure with the
refusal naming the knob is the correct behavior for a configuration error; silent success
would be the defect. No code change.

## Bounds (honest)

- Path overlays: carried (spec contradiction owned by TASK-260906-3x0w4y; the fix has
  now landed on curator-spec main as `bd39adb` "Make the section 6 path overlay
  declarable", after this task's authority `550579d` — this tree still implements the
  `550579d` form-on-every-overlay rule, the main-lane red above is its exact
  signature, and consumption is follow-up work); ledger rows 303–304 keep describing
  what their tests assert.
- POSIX dotfile list: carried (names TASK-260906-vlrjo1); bound stated in code and report;
  case runs unskipped on all three runners.
- Store missing-branch TOCTOU (increment-2 M3): untouched by this delta.
- Carried from rework 3 unchanged: `requires` never names a path source structurally;
  secondary fixed-home divergent-file comparison has no probe; import audit is the shared
  always-strict path; hard-link rejection inherited from store discipline; opencode global
  skills below the operator home and ledgered/store-target entries belong to §9.4 migration;
  machine configuration cannot pre-record import consent; fold-collision skips on
  case-folding filesystems (host-capability).
- C4-m2 resync-after-publish shape: documented above, not a defect.
- No new §8.4 helper: stated above (vigilance + per-site mutants).
- AC ratio: 15 of 15 stage (c) rows driven through the production entry point; the path row
  now drives both halves (immutability across update/sync/use, refresh across same-name
  reinstall with re-materialization).

## Gate table (exact standalone commands, observed exit codes)

All commands run as standalone processes in
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c` on the final working
tree (uncommitted rework-4 delta on `4a8a1d42`). Exit codes are the process's own
(no `tee`, no pipe; where output was redirected to a file, `$?` was read before any
other command). The two `test-gate.sh` lanes ran sequentially, never concurrently
and never alongside a `-race` suite.

- `go build ./...` → exit 0, empty output.
- `go vet ./...` → exit 0, empty output.
- `gofmt -l cmd internal` → exit 0, empty output.
- `golangci-lint run ./...` → exit 0, `0 issues.`
- `bash .github/ci/gate-selftest.sh` → exit 0, `gate-selftest: 94 passed, 0 failed`.
- `bash .github/ci/no-broad-suppression.sh` → exit 0, `no-broad-suppression: ok`.
- `bash .github/ci/ledger-consistency.sh /tmp/rework4-evidence/ledger` → exit 0,
  `206 rows checked across linux darwin windows`, `ok` (includes the three new
  refusal rows and the reworded row 329).
- Default lane `CURATOR_CONFORMANCE_ROOT=/tmp/rework4-spec-pin/conformance/v1 bash .github/ci/test-gate.sh /tmp/rework4-evidence/lane-default` (root materialized via `git -C /Users/iv/Developer/ReluxWorks/curator-spec archive 0ed5c691e9208eea52f21db2fc05e226ce3516fd conformance/v1 | tar -x -C /tmp/rework4-spec-pin`, archive exit 0, tar exit 0) → exit 0. Final verdict line: `test-gate: go test exit=0, platform-case gate exit=0`. `suite-plan: served=69 deferred=3 excluded=0` (defer `internal/config`, `internal/envfragment`, `internal/envmarker`; `TestSystemV2LockableSubsetIsClassWide` runs hermetic and is `ok`). `platform-case gate: ok` (32 skips). All four path-snapshot rows `ok` in the served `internal/envprofile` stage.
- Candidate lane `CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator-spec/conformance/v1 CI_REQUIRE_FULL_ROOT=1 bash .github/ci/test-gate.sh /tmp/rework4-evidence/lane-candidate` (root = curator-spec checkout at `87a0d00`, which contains spec fix `bd39adb`) → exit 1, reported red with rationale below (spec evolution beyond the task authority `550579d`, not a delta regression). `suite-plan: served=72 deferred=0 excluded=0`, `platform-case gate: ok` (20 skips). The only failing package is `internal/config`: `TestManagerConfigV2SchemaCases` (16 subcases: `valid-overlay-path-*.json` rejected with `requires exactly one of range, tag, or revision`, `invalid-overlay-*-with-form.json` accepted) and `TestManagerConfigV2Vectors` (`schema2-overlay-path-*`). These are the `bd39adb` "Make the section 6 path overlay declarable" cases: a git source now requires exactly one form, a path source requires none and permits none. This tree implements the `550579d` rule (a form on every overlay), so the new valid-path-without-form cases are rejected and the new invalid-path-with-form cases are accepted. No orchestrator directive to consume `bd39adb` was recorded for this run (`task-board spawn directives` → none); per the rework-1 conditional the M1 path-overlay bound stands and names the landed commit (see Bounds). Consuming it — porting the spelling discriminator to the Go reader plus all three surfaces — is follow-up work, not this rework.
- Authority lane `CURATOR_CONFORMANCE_ROOT=/tmp/spec-550579d-wt/conformance/v1 CI_REQUIRE_FULL_ROOT=1 bash .github/ci/test-gate.sh /tmp/rework4-evidence/lane-authority-wt` (real checkout at the task authority `550579d` via `git worktree add --detach`, not an archive) → exit 0. Final verdict line: `test-gate: go test exit=0, platform-case gate exit=0`. `suite-plan: served=72 deferred=0 excluded=0` with `CI_REQUIRE_FULL_ROOT=1`. The schema-case families serve rather than skip (`TestManagerConfigV2SchemaCases`, `TestSystemConfigV2SchemaCases`, `TestManagerConfigV2Vectors` all `ok`); `TestConformanceSnapshotAcquisition` `ok`; every new test `ok`.
- Lesson: a `git archive`-materialized root must never back a `CI_REQUIRE_FULL_ROOT=1` lane. `git archive` expands `export-subst` placeholders, so `internal/interop :: TestConformanceSnapshotAcquisition/byte-exact-snapshot` fails against the archive (`fixture subst.txt on disk, 65 bytes, does not match the vector, 40 bytes; the checkout normalized it`) while passing against a real checkout. Proven pre-existing on clean HEAD `4a8a1d42` (same failure, exit 1, without this delta). The required default lane (SPEC_PIN archive, without `FULL_ROOT`) is unaffected — the rc.9 pin predates that vector and the package defers there.
- `go test -count=1 -timeout 30m ./cmd/curator` → exit 0, `ok github.com/relux-works/curator/cmd/curator 296.591s` (run solo after all lanes).
- `go test -count=1 ./internal/envprofile/ -run '<new/refusal/policy rows>'` targeted rows → exit 0 (all eight pass; full suites exercised via the lanes above, not `-run` subsets).
- Hosted CI was not consulted: nothing pushed, PR #61 untouched per the brief, so no `gh pr checks` output exists for this delta. The three local lane reproductions above stand in for it.

Stated plainly: the candidate lane against curator-spec main is red (exit 1) on the `bd39adb` spec-evolution cases only; everything else in this report is green as quoted. An unread gate reported green would be a lie; this one is read and red, with the exact failing cases and the sentence that decides the bound.
