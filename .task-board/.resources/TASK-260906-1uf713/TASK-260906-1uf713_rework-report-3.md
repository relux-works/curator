# TASK-260906-1uf713 rework report 3: path-update immutability, the second seam bypass, the transcribed lockable pin

Head: `4a8a1d42` on `feat/agent-environments-stage-c` (4 signed commits on `4df4d507`).
Base: curator main `b056e5da`. Authority: curator-spec `550579d`.
Worktree: `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c`. No push, PR #61 untouched.

Commits (all signed, each building green):

- `7fcf9c1b` C3-B2: path update resolves the root from the store snapshot, plus immutability tests
- `9342598d` C3-B1: structural require_current_profile gate plus caller-enumeration tests
- `598727c2` C3-M1: transcribed §12.2 pin proved by widening and narrowing
- `4a8a1d42` C3-m1+m2 and ledger: compose-add inert warning, four-rule weight test, ledger rows

Staged with named paths only; `git status --short` read before each commit
(no `git add -A`; the gate run leaves no artefacts in the tree — final
`git status --short` is empty).

## Finding → resolution (all six)

### C3-B2 (blocking, done first) — `profile update` re-reads a `path` source

Agreed on both consequences. `updateLocked`
(`internal/envprofile/envprofile.go`) re-resolved a `KindPath` root with
`LoadManifest(source.Path)` + `stateForPath(home, manifest.Name,
source.Path)`, so an edit to the source moved the `state_sha256` pin with
no reinstall, and every §9.6 imported profile (whose staging directory
`importLocked` deletes) broke `profile update --all` with
`profile_source_path_missing`.

Fix from §1: the `KindPath` arm now resolves the root from the snapshot the
store already holds under the old lock's `state_sha256` pin — never from
`source.Path`. It reads the root member's pin from the old lock, loads the
manifest from `contextstore.EntryDir(home, KindContext, oldLock.Root, pin)`,
fails `profile_source_invalid` on a missing pin, an unreadable snapshot, or
a snapshot/lock name mismatch, and builds the resolution `RootState` from
those store bytes. `resolveOverlays` still runs below the switch, so a
`path` root with `git` overlays is a legitimate update that moves when an
overlay moves. The `Update` doc comment that said "a path root re-resolves
against its directory" now states the snapshot discipline.

Driven before/after through the review's own probe (`probes-c3.sh` C3-B2
section, binary built from this head):

```
before (4df4d507): state hash ff6c35c7… -> 43d9f269…  SNAPSHOT MOVED
after  (4a8a1d42): state hash ff6c35c7… -> ff6c35c7…  snapshot immutable
before: profile update --all -> exit 1  profile_source_path_missing (imported staging gone)
after:  profile update --all -> exit 0  (imported: unchanged)
```

Committed tests (production entry points `Install`, `UpdateWithPolicy`,
`SyncWithPolicy`, `UseWithPolicy`, `Import`, `ListWithPolicy` — the calls
the CLI rows reach):

- `TestPathSnapshotImmutableAcrossUpdateSyncUse`: installs `pk`, edits the
  source, asserts the pin across `Update` (asserts `moved == false`),
  `Sync` and `Use`, then installs the edited tree as `pk2` and asserts a
  different pin (the test is not vacuous).
- `TestPathUpdateAllSucceedsWithImportedProfile`: imports (asserts the
  recorded `source.Path` names no existing entry, i.e. the deleted-staging
  shape), then runs the `profile update --all` loop (`ListWithPolicy`
  skipping `default`, `UpdateWithPolicy` each) and asserts every update
  succeeds.

Narrowing mutant (the old arm restored over the new tests): both tests fail
with the exact review symptoms — `update of a path root must not move the
lock: the snapshot is immutable` and `update imported:
profile_source_path_missing: path "…/.import-N" names no existing
filesystem entry`.

The first default-lane run on this rework exposed one stale oracle the
targeted `-run` invocations had missed: `TestUpdatePathMovesLock` asserted
the old re-read (`update must move the lock`). It is rewritten to the §1
oracle as `TestUpdatePathLeavesLockInPlace` (same production path, inverted
expectation, with a comment naming C3-B2); it is not ledgered. The lane was
re-run green afterwards — see the gate table. Lesson recorded: the full
touched package, not a `-run` subset, is the unit of verification.

### C3-B1 (blocking) — `profile use <name> --clear` skips the locked `require_current_profile`

Agreed, including the `repeat-of: cycle-1 B1` classification. Fixed
structurally at both layers:

1. The seam (`useLocked`, `internal/envprofile/switch.go`): the §12.2 gate
   is now the single construction point of a machine-scope switch. An
   operand beside `--clear` is refused (`profile use --clear takes no
   profile operand`) instead of silently ignored, and one gate —
   `if scope == "" { CheckMachineUse(effective) }` — covers the clear and
   non-clear branches before any write, preceding the installed check so an
   uninstalled profile is still a configuration error rather than
   not-found. `CurrentFile` keeps its single production writer (the publish
   below the gate; `SetCurrent` has no production caller), so no caller
   reaches the seam in machine scope without passing through it. The
   enumeration comment is rewritten to name `profile use` (clear or not),
   `install --use`, first-install auto-activation, `import --use`, and
   resync, plus the single-writer fact. `Policy.CheckMachineUse`'s comment
   is updated to the same enumeration.
2. The row (`cmdProfileUse`, `cmd/curator/profile.go`): `cli/curator.md`
   defines exactly two `profile use` forms, so a positional beside `--clear`
   is now a usage error (exit 2), with or without `--env`/`--target`.

Caller enumeration, each driven through `run()`:

| caller reaching the seam | run() driver | outcome under lock to `acme`/`groot` |
|---|---|---|
| `profile use <other>` (machine) | `TestProfileUseLockedRequireRefuses` (existing) | exit 1, names the knob |
| `profile install <other> --use` | `TestProfileInstallUseLockedRequireRefuses` (existing) | exit 1, moves no current |
| first-install auto-activation | `TestProfileInstallFirstActivationLockedRequireRefuses` (existing) | exit 1, names the knob |
| `profile import --as other --use` | `TestProfileImportUseLockedRequireRefuses` (new) | exit 1, moves no current |
| resync after a moving update | `TestProfileUpdateResyncPassesLockedRequire` (new; git range install then a new tag, update moves) | exit 0, `updated` (required profile passes) |
| scoped `use` / `use --clear --env` | `TestProfileScopedUseUnaffectedByLockedRequire` (new) | exit 0 both; machine `use other` still refused |
| `profile use <name> --clear` (undefined form) | `TestProfileUseClearOperandIsUsage` (new) | exit 2 usage, current stays `acme` |

Narrowing mutant to quote (the reviewer's exact shape): `if scope == ""`
→ `if scope == "" && !clearScope` in `useLocked`. Because the CLI now
rejects the undefined form as usage, no `run()` row can reach the exempted
path by construction — which is the point — so the kill is at the seam
library level: `TestMachineClearReachesSeamGate` (new, production
`UseWithPolicy`) fails with `machine clear under the lock = <nil>, want the
§12.2 refusal`. That test also proves the operand refusal, the unchanged
current, the unaffected scoped clear, and that the required profile passes
the gate (failing later only on the installed check). An additional
admit-one-name mutant shape is covered by the `run()`-level refusal rows,
which assert the per-profile refusal rather than mere existence.

### C3-M1 (major) — the lockable set pinned by one hand-written knob

Agreed, including the third-consecutive-cycle classification. The consumer
gate was already class-wide (eight consumer mutants die here as well: the
four named plus four of the same shape re-applied during this rework); the
derivation source was the unpinned list. The test now transcribes §12.2 —
the six keys the specification states (`overlays_allowed`, `precedence`,
`mcp_package_allowlist`, `passable_env_names`, `require_current_profile`,
`isolation`), written out independently — and asserts set equality with
`LockableEnvKeys` in both directions before deciding each knob's branch
from the transcription rather than the map.

Mutant table (each keeps the gate; map mutants observed on the hermetic
`./internal/config/` suite, consumer mutants likewise):

| mutant | result |
|---|---|
| widen `LockableEnvKeys += environments.secret_material_waivers` (the reviewer's one-liner) | KILLED — `LockableEnvKeys carries 7 keys, §12.2 transcribes 6` |
| widen `LockableEnvKeys += environments.current_profile` | KILLED — same set-equality (`7 vs 6`) |
| narrow `LockableEnvKeys -= environments.isolation` | KILLED — `carries 5 keys, §12.2 transcribes 6` |
| narrow consumer `parseSystemEnvironments && key != "forms"` (cycle-1) | KILLED — `non-lockable/forms admitted` |
| narrow consumer `&& key != "secret_material_waivers"` | KILLED — `non-lockable/secret_material_waivers admitted` |

The `secret_material_waivers` end-to-end arm is closed by the same gate:
`mergeSystemEnvironments` rule 3 can only inherit what
`parseSystemEnvironments` admits, and the map it derives from is now pinned
by the transcription.

### C3-m1 — `profile compose add` silent under a locked `overlays_allowed: false`

Warned, following the `list` reasoning: `cmdComposeAdd` now prints
`warning: overlays_allowed is false: the added declaration is inert;
resolution joins the root alone` when the effective policy forbids
overlays. Driven by `TestProfileComposeAddWarnsWhenOverlaysForbidden`
(`run()`; asserts the add warning and the reloaded list warning).

### C3-m2 — all four §6 weight rules disagreeing at once

Committed as `TestWeightRulesAllFourDisagree` (production `Install` over
git fixtures): leaf manifest 5, two agreeing requirers at 20, root map 30,
machine overlay 40 → lock records weight 40, `overlay: true`,
`required_by: [mid1, mid2]`, exactly the reviewer's drive.

### C3-m3 — `TestTakeoverWarnsDotfileHeuristic` drives `UseWithPolicy`

Noted, no action: it is a production entry point the CLI calls with the
same arguments, not the uncalled-gate shape.

## AC coverage (honest ratio)

15 of 15 stage (c) rows driven through the production entry point. The two
gaps the review named are closed: the `path`-immutability half of the path
row is now driven (`TestPathSnapshotImmutableAcrossUpdateSyncUse`, plus the
import-shaped `update --all` row and the rewritten `Update` oracle), and
the lockable-set row is pinned by the transcription rather than one knob
name. The one carried bound stands: the `path`-overlay declaration (spec
contradiction owned by TASK-260906-3x0w4y, curator-spec PR #47); ledger row
304 keeps describing what `TestPathOverlayJoinsClosure` actually asserts.

## Bounds (honest)

- Path overlays: see above (names TASK-260906-3x0w4y).
- POSIX dotfile list: carried (names TASK-260906-vlrjo1).
- Store missing-branch TOCTOU (increment-2 M3): untouched by this delta.
- Carried from rework 2 unchanged: `requires` never names a path source
  structurally; secondary fixed-home divergent-file comparison has no
  probe; import audit is the shared always-strict path; hard-link rejection
  inherited from store discipline; opencode global skills below the
  operator home and ledgered/store-target entries belong to §9.4 migration;
  machine configuration cannot pre-record import consent; fold-collision
  skips on case-folding filesystems (host-capability).
- Resync refusal for an already-violating machine (current != required when
  an update moves): covered by construction — resync calls the same single
  gate with the same machine-scope shape as install activation, so no
  second gate exists to bypass; the committed row proves the allowed path
  (required current) succeeds. A mutant distinguishing resync from install
  activation would need call-stack inspection, which is precisely what the
  single-gate structure removes.
- C3-B1 reviewer's mutant shape kills the seam-level test rather than a
  `run()` row: the CLI rejects the undefined form as usage, so no `run()`
  row can reach the exempted path by construction; stated here rather than
  claimed otherwise.

## Gate table (exact standalone commands, observed exit codes)

All commands run as standalone processes in
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c` on content
identical to the final head `4a8a1d42` (the four commits change nothing
beyond the reviewed working tree; final `git status --short` is empty).
Exit codes are the process's own (no `tee`, no pipe; where output was
redirected to a file, `$?` was read before any other command, so it is the
gate's own status).
The two `test-gate.sh` lanes ran sequentially, never concurrently and never
alongside a `-race` suite.

- `go build ./...` → exit 0, empty output.
- `go vet ./...` → exit 0, empty output.
- `gofmt -l cmd internal` → exit 0, empty output.
- `golangci-lint run ./...` → exit 0, `0 issues.`
- `bash .github/ci/gate-selftest.sh` → exit 0, `94 passed, 0 failed`.
- `bash .github/ci/no-broad-suppression.sh` → exit 0, `no-broad-suppression: ok`.
- `bash .github/ci/ledger-consistency.sh /tmp/rework3-evidence/ledger` → exit 0, `203 rows checked across linux darwin windows`, `ok`.
- Default lane `CURATOR_CONFORMANCE_ROOT=/tmp/rework3-spec-pin/conformance/v1 bash .github/ci/test-gate.sh /tmp/rework3-evidence/lane-default` (root materialized via `git -C /Users/iv/Developer/ReluxWorks/curator-spec archive 0ed5c691e9208eea52f21db2fc05e226ce3516fd conformance/v1 | tar -x -C /tmp/rework3-spec-pin`, exit 0) → exit 0. Final verdict line: `test-gate: go test exit=0, platform-case gate exit=0`. `suite-plan: served=69 deferred=3 excluded=0` (defer `internal/config`, `internal/envfragment`, `internal/envmarker` — the pin predates those families; `TestSystemV2LockableSubsetIsClassWide` runs hermetic and is `ok`). `platform-case gate: ok` (32 skips). All nine new ledger rows `ok` in the served `internal/envprofile` and `cmd/curator` stages.
- Candidate lane `CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator-spec/.temp/STORY-260906-1a2i5a/worktree/conformance/v1 CI_REQUIRE_FULL_ROOT=1 bash .github/ci/test-gate.sh /tmp/rework3-evidence/lane-candidate` → exit 0. Final verdict line: `test-gate: go test exit=0, platform-case gate exit=0`. `suite-plan: served=72 deferred=0 excluded=0` with `CI_REQUIRE_FULL_ROOT=1`. `platform-case gate: ok` (20 skips). The schema-case families serve rather than skip (`TestManagerConfigV2SchemaCases`, `TestSystemConfigV2SchemaCases`, `TestManagerConfigV2Vectors` all `ok`); every new test `ok`.
- `go test -count=1 -timeout 30m ./cmd/curator` → exit 0, `ok github.com/relux-works/curator/cmd/curator 287.609s`.
- `go test -count=1 ./internal/envprofile/ ./internal/config/` → exit 0, `ok internal/envprofile 57.529s`, `ok internal/config 0.853s` (the full touched packages, not a `-run` subset).
- Hosted CI was not consulted: nothing pushed, PR #61 untouched per the brief, so no `gh pr checks` output exists for `4a8a1d42`. The two local lane reproductions above stand in for it.

Stated plainly: the first default-lane attempt on this rework printed
`test-gate: go test exit=1, platform-case gate exit=0` — one failing case,
`internal/envprofile TestUpdatePathMovesLock`, the stale oracle rewritten
above. It was fixed (not skipped, not narrowed) and the lane re-run green;
the red is reported here rather than buried.

## CLI probe (review's `probes-c3.sh` re-driven against `4a8a1d42`)

```
built from 4a8a1d42
===== C3-B1 =====
  profile use other                -> exit 1  (gate holds: expect 1)
  profile use bogus --clear        -> exit 2  (usage; nothing written, no current file)
===== C3-B2 =====
  state hash before edit: ff6c35c747a8bd75180447f23a9f6f95ca0f0368f6cd29972a1d45293d900586
  state hash after update: ff6c35c747a8bd75180447f23a9f6f95ca0f0368f6cd29972a1d45293d900586
  snapshot immutable
    imported: unchanged
    profile update --all -> exit 0
```

(The script's own C3-B1/C3-M1 annotations still narrate the old
expectations — "BYPASS: observed 0", "SURVIVES" — because they are static
text around the mutant shape, not a re-applied mutant. The observed exits
above are the fix: 2 and immutable. The M1 widening mutant was re-applied
by hand during this rework and dies at the set-equality; see the mutant
table.)

## Windows-fixture sweep

Every new test builds paths with `filepath.Join` over `t.TempDir()` bases;
no POSIX literal, no backslash in values, no `filepath.IsAbs` on fixtures.
The git-based rows (`TestWeightRulesAllFourDisagree`,
`TestProfileUpdateResyncPassesLockedRequire`) reuse the `gitFileURL`
forward-slash wiring and the `newGitIdentities`/`gitContextRepo` hermetic
fixtures already green on all three runners. No new `$HOME`-relative
lookup, no well-known-location list, no new symlink or permission fixture.
Found nothing new.
