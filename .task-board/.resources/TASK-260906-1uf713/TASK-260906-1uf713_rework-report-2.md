# TASK-260906-1uf713 rework report 2: corrupt-marker import, derived lockable gate, four minors

Head: `4df4d507` on `feat/agent-environments-stage-c` (5 signed commits on `0bcea201`).
Base: curator main `b056e5da`. Authority: curator-spec `550579d`.
Worktree: `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c`. No push, PR #61 untouched.

Commits (all signed, each building green):

- `221e0082` C2-B1: corrupt-marker import fails closed with a loss
- `86c81cfc` C2-M1: derive the lockable-subset test from the knob list
- `63256470` C2-m2: drop Config.CheckMachineUse with no production caller
- `e516fc97` C2-m1+m3+m4: dotfile bound, ledger wording, remove-first
- `4df4d507` ledger: register the three marker-state import rows

## Finding -> resolution (all six)

### C2-B1 (blocking) — corrupt marker re-imports the managed root as native

Agreed, including the repeat-of-cycle-1-B3 classification and the §8.4
absence-versus-failed-read framing. `readRootSurface` skipped a managed
home only on `(marker, nil)`; any read error fell through to native
detection, so a truncated or hand-edited marker re-detected curator's own
generated root-context document (full §5.1 header) as a native surface at
exit 0 with no warning and no loss.

Fix as the class: `readRootSurface` now distinguishes the two facts. No
marker file is absence and detects the surface; a marker that exists but
cannot be read or decoded is a failed read of the evidence that classifies
the surface, so the surface is neither detected nor silently dropped — the
marker file itself joins the loss list and the consent gate decides (lossy
without `--allow-lossy`, warning-carrying with it, managed root never
reassembled either way).

Narrowing mutant (gate stays, admits exactly one adapter):

```go
// internal/envprofile/import.go — readRootSurface
-	marker, err := envmarker.Read(native)
-	if err != nil {
+	marker, err := envmarker.Read(native)
+	if err != nil && envID != "claude_code" {
```

Killed by `TestImportCorruptMarkerIsLoss` (production `Import`):
`corrupt-marker roots detected: [{envID:claude_code data:[60 33 45 45 ...
curator-root-context-v2 ...]}]` — the exact bypass bytes, a complete §5.1
header re-detected as native. The same mutant also kills
`TestImportUnreadableMarkerIsLoss`. `TestImportSkipsManagedRootContext`
(intact markers) stays green under the mutant, proving exactly one member
admitted.

Three marker outcomes, each driven through the production `Import`:

| marker state | outcome |
|---|---|
| absent | surface detected as native; lossless import installs it (`TestImportAbsentMarkerDetectsNative`) |
| corrupt (`not json at all`) | no roots detected; marker file is the loss; import stops with `environment_import_lossy` naming `.agent-environment.json`; under `--allow-lossy` proceeds with the loss as warning and no `claude_code` module reassembled (`TestImportCorruptMarkerIsLoss`) |
| unreadable (mode 000) | same as corrupt (`TestImportUnreadableMarkerIsLoss`; host-capability skip where the host reads through mode 000, same tolerance as the unreadable-root row) |

Absence-vs-failed-read sweep of this stage (manager-owned state reads):

| site | disposition |
|---|---|
| `readRootSurface` marker (`import.go`) | FIXED here: error is a loss, absence detects, valid marker skips |
| `readSkillsLedger` (`import.go`) | already fail-closed (`ee6743a2`): absence is empty, any other failure is the ledger itself as loss |
| `recoverSkill` skill-marker (`import.go:411`, single-return `marker.Read`) | fail-closed: a failed skill-marker read falls through to the git-checkout probe, and an unrecoverable entry is a loss, never a silent import |
| `isStoreEntry` Lstat/Readlink error (`import.go`) | fail-closed: false leads to recovery, and an unrecoverable entry is a loss |
| `switch.go:419` prior marker read | already fail-closed: error returns `EntryResult{OK:false}` |
| `verifyHome` marker read (`managed.go:1064`) | already fail-closed: error records `marker invalid` |
| `purgeHomes` marker read (`switch.go:696`) | conservative by construction: a failed read skips deletion rather than removing what it cannot verify |
| `status.go:432` provisioned flag, `status.go:521` orphan listing | read-only display, swept and left: a failed read reports unprovisioned / skips the orphan row. Neither admits, installs, nor deletes anything; inventing a status-time refusal there would be new behaviour outside this finding |

### C2-M1 (major) — the lockable-subset test is now derived, not listed

Agreed, including the repeat-of-cycle-1-m1 classification. The production
gate was already derived (`systemEnvKnobs` from `LockableEnvKeys`); the
test was the hand-written list, covering 5 of 12 non-lockable knobs. Now
`EnvKnobNames` is the single closed §12.1 list that the reader (`envKnob`,
`SplitEnvKnob`) and the test both derive from, and the test carries a
grammatically valid system-file payload for every one of the 18 knobs, so
the only possible refusal is the lockable gate itself. A knob added to the
reader without a payload fails (`no system-file payload for knob ...`);
the payload-count guard fails if the two drift.

Derived-gate mutant table (each keeps the gate, admits exactly one member;
all four kills observed without `CURATOR_CONFORMANCE_ROOT`, and the
`secret_material_waivers` kill additionally observed with the root set —
the config package is hermetic, and the suite-plan serves it in both
lanes):

| mutant (`parseSystemEnvironments`) | result |
|---|---|
| `&& key != "forms"` (cycle-1's survivor) | KILLED — `TestSystemV2LockableSubsetIsClassWide/non-lockable/forms`: `admitted: err = <nil>` |
| `&& key != "secret_material_waivers"` | KILLED — `.../non-lockable/secret_material_waivers`: `admitted: err = <nil>` |
| `&& key != "xdg_seed_allowlist"` | KILLED — `.../non-lockable/xdg_seed_allowlist`: `admitted: err = <nil>` |
| `&& key != "in_place_mode"` | KILLED — `.../non-lockable/in_place_mode`: `admitted: err = <nil>` |

The end-to-end arm of the finding (system-injected waivers surfacing in
`env config show` under the mutant) is closed by the same gate the test
now covers class-wide: `mergeSystemEnvironments` rule 3 can only inherit
what `parseSystemEnvironments` admits, and no single knob can be admitted
without failing the test.

### C2-m1 — the §9.5 dotfile list bound, stated in code and here

The production lookup is correct (operator home via `os.UserHomeDir`,
portable relative joins, `0bcea201`'s `pinOperatorHome` fixture asserts
real behaviour on all three runners). The *list* is POSIX-portable by
construction and stays that way: whether §9.5's closed list is
platform-specific is a spec question §9.5 does not decide (it gives
`~/.local/share/chezmoi` and "and the like" with no per-platform
statement), now filed as TASK-260906-vlrjo1. Inventing Windows locations
would record an unverified tool fact as normative. The bound is stated on
`dotfileStateDirs` in `managed.go` naming that task, and here: no Windows
location added, case not skipped into silence (row keeps
`linux,darwin,windows` with no tolerance).

### C2-m2 — `Config.CheckMachineUse` deleted, comment gone with it

Routed by deletion rather than correction: the seam gate is
`Policy.CheckMachineUse` at `useLocked` (`switch.go:207`), the single
`CurrentFile` writer the cycle-2 review verified. The `Config` method had
zero production callers, and a corrected comment would still leave the
uncalled-gate shape. The two config tests now assert what `Load` owns (the
knob plus its lock carried for `PolicyFromConfig`; unlocked value carried
with no lock). The refusal stays driven where it lives:
`TestPolicyFromConfigCarriesEnvGates` (production `PolicyFromConfig` +
`CheckMachineUse`) and the `run()`-level
`TestProfileInstallUseLockedRequireRefuses`,
`TestProfileInstallFirstActivationLockedRequireRefuses`,
`TestProfileUseLockedRequireRefuses`. No `Config.CheckMachineUse`
reference remains in code, tests, or docs.

### C2-m3 — ledger row 304 describes the machinery, not a declaration

Row 304 now reads: `a form-free path overlay joins the closure flagged
overlay at the machine default weight from a manufactured Policy
(resolution machinery only; no operator can declare it --
see TASK-260906-3x0w4y bound)`. Row 303's fixed wording is the model.
`ledger-consistency.sh` green at 194 rows.

### C2-m4 — remove-first symmetric in `applyPlan`

The provisioning seed write and the marker write remove first, as the
copied surfaces and `switch.go` already do. Both targets stay under
`EnvRoot(home)` — the manager's own tree, not the B2 class, as the
reviewer states. `claudeSeed` is deliberately excluded with the reason in
the comment: it reads the existing `.claude.json` and merges, so removing
first would destroy the tool-owned state it must preserve. (`writeStoreDoc`
writes content-addressed store documents under the manager home;
overwriting identical bytes is idempotent — swept, left.)

## AC coverage (honest ratio)

14 of 15 stage (c) rows driven through the production entry point. The one
stated bound is unchanged: the `path`-overlay declaration (spec
contradiction owned by TASK-260906-3x0w4y, curator-spec PR #47). No word
from the orchestrator that the spec fix landed, so the bound stands and
`cmdComposeList`'s dead `default: form = "path"` branch stays named as
dead. The rework adds no new driven rows and drops none: the three marker
tests deepen the already-driven import row (C2-B1 addressing modes), and
the derived lockable test deepens the already-driven system-config row
(C2-M1 class coverage).

## Bounds (honest)

- Path overlays: see M1 bound above (names TASK-260906-3x0w4y).
- Store missing-branch TOCTOU (increment-2 M3 / reviewer R9): untouched by
  this delta (different seam: `contextstore.EnsureState` vs import
  detection). The cycle-2 review re-ran that exact mutant on `0bcea201`
  and it survives with the TOCTOU-only rationale verified against both
  production callers; that verification stands.
- POSIX dotfile list: see C2-m1 (names TASK-260906-vlrjo1).
- Carried from rework 1 unchanged: `requires` never names a path source
  structurally; secondary fixed-home divergent-file comparison has no
  probe; import audit is the shared always-strict path with no
  import-specific secret-material vector; hard-link rejection inherited
  from store discipline; opencode global skills below the operator home
  and ledgered/store-target entries belong to §9.4 migration, never
  import; machine configuration cannot pre-record import consent
  (`AllowLossy` from the CLI flag only, `Policy.Takeover` operation-scoped);
  fold-collision skips on case-folding filesystems (host-capability).
- `status.go:432/521` read-only display collapse: swept, judged
  informative-only, see the sweep table.

## Gate table (exact standalone commands, observed exit codes)

All commands run as standalone processes in
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c` at `4df4d507`.
Exit codes are the process's own (no `tee`, no unguarded pipe).

- `go build ./...` → exit 0, empty output.
- `go vet ./...` → exit 0, empty output.
- `gofmt -l cmd internal` → exit 0, empty output.
- `golangci-lint run ./...` → exit 0, `0 issues.`
- `bash .github/ci/gate-selftest.sh` → exit 0, `94 passed, 0 failed`.
- `bash .github/ci/no-broad-suppression.sh` → exit 0, `no-broad-suppression: ok`.
- `bash .github/ci/ledger-consistency.sh /tmp/rework2-evidence/ledger` → exit 0, `194 rows checked`, `ok`.
- Default lane `CURATOR_CONFORMANCE_ROOT=/tmp/rework2-spec-pin/conformance/v1 bash .github/ci/test-gate.sh /tmp/rework2-evidence/lane-default` (root materialized via `git -C <spec> archive 0ed5c691e9208eea52f21db2fc05e226ce3516fd conformance/v1 | tar -x -C /tmp/rework2-spec-pin`, exit 0) → exit 0. `test-gate: go test exit=0, platform-case gate exit=0`; `suite-plan: served=69 deferred=3 excluded=0` (`defer internal/config`, `internal/envfragment`, `internal/envmarker` — the pin predates those families; `TestSystemV2LockableSubsetIsClassWide` runs hermetic and is `ok`); `platform-case gate: ok` (32 skips). All three new marker tests `ok` in the served `internal/envprofile` stage.
- Candidate lane `CURATOR_CONFORMANCE_ROOT=<story-worktree-at-550579d>/conformance/v1 CI_REQUIRE_FULL_ROOT=1 bash .github/ci/test-gate.sh /tmp/rework2-evidence/lane-candidate` → exit 0. `test-gate: go test exit=0, platform-case gate exit=0`; `suite-plan: served=72 deferred=0 excluded=0` with `CI_REQUIRE_FULL_ROOT=1 -- every package must be served`; `platform-case gate: ok` (20 skips). The schema-case families serve rather than skip (`TestManagerConfigV2SchemaCases`, `TestSystemConfigV2SchemaCases`, `TestManagerConfigV2Vectors` all `ok`); all three new marker tests and the derived lockable test `ok`.
- `go test -count=1 -timeout 30m ./cmd/curator` → exit 0, `ok ... 290.821s`.
- `go test -count=1 -race ./internal/config/` → exit 0; `go test -count=1 -race -run '<marker/import/takeover/heuristic cases>' ./internal/envprofile/` → exit 0 (both after the lanes, never alongside them).
- The two lanes ran sequentially, never concurrently and never alongside a `-race` suite.
- Hosted CI was not consulted: nothing pushed, PR #61 untouched per the brief, so no `gh pr checks` output exists for `4df4d507`. The two local lane reproductions above stand in for it.

## CLI probe (production binary built from `4df4d507`)

`cli-marker-probe.sh` (attached beside this report) drives the three
marker outcomes through `run()` against scratch homes:

```
built from 4df4d507
===== absent marker: native file is detected =====
  import exit: 1 (profile_use_partial: installed, activation needs --takeover)
  installed: absentr path 1.0.0
===== intact marker: managed root is skipped =====
  import exit: 0
  managed root imported as native: no (gate holds)
===== corrupt marker: loss naming the marker, managed root never native =====
  import exit: 1 (want non-zero)
  environment_import_lossy
  loss names: .agent-environment.json
  managed root imported as native: no
===== unreadable marker =====
  import exit: 1 (want non-zero)
  environment_import_lossy
```

The absent exit 1 is `profile_use_partial`, not a refusal: the surface was
detected and `absentr` installed through the path pipeline at version
1.0.0 with a state-hash pin (confirmed via `profile list`); only the
activation over the unmanaged file needs `--takeover`, exactly the
`TestImportPartialActivationKeepsLock` shape.

## Windows-fixture sweep

Every new test builds paths with `filepath.Join` over `t.TempDir()` bases;
no POSIX literal, no backslash in values, no `filepath.IsAbs` on fixtures.
The marker paths under test come from `filepath.Join(home, envmarker.Name)`
in production and the same join in the tests. Permission-based tests (the
new unreadable-marker row, like the unreadable-root row) probe readability
and skip under the classified `host-capability` reason on
Windows/superuser; the ledger row carries that tolerance. The rework delta
adds no `$HOME`-relative lookup, no well-known-location list beyond the
C2-m1 bound, and no new symlink fixture. Found nothing new.
