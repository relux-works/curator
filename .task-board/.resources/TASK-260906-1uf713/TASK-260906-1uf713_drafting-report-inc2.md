# TASK-260906-1uf713 drafting report — increment 2 (composition, path kind, onboarding import)

Successor run RUN-260906-c9dc57 to the increment-1 run. Increment 1
(commits `8adc2214`, `ff91d1a8`: schema-2 reader + CLI, AC 9 of 15)
stands; this increment lands the remaining 6 AC rows.

Branch `feat/agent-environments-stage-c` in
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c`.
Authority curator-spec `550579d`. No push, no tag, no PR.

## Commits (all signed, each building and testing green)

- `9a5853e4` overlays join the closure, precedence from policy, path diagnostics
- `234fd3ab` onboarding takeover across the mutating operations
- `4100c7a9` onboarding import through the path pipeline
- `ab5d9054` reader grammar follows the published overlay family
- `91ed6b26` repeated-overlay composition test and ledger row
- `4f513f20` heuristic warning, read-failure losses, use-takeover row

## What changed

Composition (`internal/envprofile/overlays.go`, resolution wiring in
`envprofile.go`): machine overlay declarations resolve jointly with the
root. Git overlays fetch and read their names from the manifest at the
pinned commit; path overlays snapshot immutably under their state hash.
A repeated closure name fails `environment_composition_invalid`. The
four §6 weight rules apply in order with the error-versus-warning split,
and resolution warnings surface on install and update. The §12.2
forbidding policy empties every list at resolution. `Remove` refuses a
profile named as another profile's overlay member with `profile_in_use`.
Precedence comes from the machine policy at every emission site
(switch, repair, verify, fragment, status).

Path kind (`internal/contextstore`, `envprofile.go`): missing vs
unreadable are distinct diagnostics; non-directories, nested `.git`,
links, and folded platform paths are `profile_source_invalid`; the root
`.git` is excluded. Reader grammar requires exactly one form on every
overlay (the published family accepts a revision on a path source); the
§1 refusal for a form on a path source fires at resolution.

Takeover (`switch.go`, `managed.go`, CLI): per-entry and per-repair
inventory, foreign-manager stop with the abort-or-take-over choice,
non-blocking dotfile heuristic over the closed list, replace notice,
always-backup (now on provisioning too). `--takeover` on profile
install/use/update/sync and env resolve (repair-only).

Import (`import.go`, CLI): detection per adapter over native homes,
lossy classification with the loss list, per-operation consent,
reassembly (normalization, ascending module order, name-taken before
any write, revision-pinned `requires.skills` with the foreign warning),
installation through the path pipeline with always-strict audit.
Markers and `profile list` record `imported_from_native`.
`profile import [--as] [--allow-lossy] [--use]`.

## AC coverage: 15 of 15 stage (c) rows driven through the production entry point

Composition (§6, §6.1) — production call site `Install`/`UpdateWithPolicy`
→ `resolveOverlays` → `contextresolve.Resolve`:

- overlays join beside the root: `TestPathOverlayJoinsClosure`,
  `TestGitOverlayJoinsClosure`
- repeated name: `TestOverlayDuplicateNameIsCompositionInvalid`,
  `TestOverlayRepeatedNameIsCompositionInvalid`
- unreadable overlay source: `TestMissingOverlayPathIsPathMissing`
- weights rules 1–4 in order: `TestWeightRulesApplyInOrder`,
  `TestOverlayExplicitWeightOverridesDefault`
- conflict error/warning split: `TestWeightConflictErrorAndWarning`
- `context_weights_duplicate`: `TestWeightDuplicateRefused`
- `context_weight_unknown`: `TestWeightUnknownRefused`
- `context_weights_not_root`: `TestNonRootWeightsMapRefused`
- forbidding policy empties: `TestForbiddenOverlaysResolveAlone`,
  `TestPolicyFromConfigCarriesEnvGates`
- precedence primitives independently: `TestPrecedencePrimitivesDriveEmission`
  (all four pairs through `UseWithPolicy`), `TestPrecedenceTieKeepsTopologicalOrder`
- overlay-of removal guard: `TestRemoveRefusesOverlayOfAnotherProfile`

Path kind (§1) — call site `Install` → `stateForPath`/`EnsureState`:

- missing vs unreadable: `TestMissingPathOperandIsPathMissing`,
  `TestUnreadablePathIsUnreadable`, `TestMissingOverlayPathIsPathMissing`
- non-directory, nested `.git`, symlink, fold collision:
  `TestNonDirectoryPathIsSourceInvalid`, `TestNestedGitIsSourceInvalid`,
  `TestSymlinkInPathIsSourceInvalid`, `TestFoldedPathsAreSourceInvalid`
- root `.git` excluded, state-hash pin: `TestRootGitExcluded`
- form on path declaration: `TestPathOverlayFormIsSourceInvalid`
- reader grammar per the published family:
  `TestPathOverlayDeclarationParses`, schema-case conformance green

Onboarding (§9.5) — call sites `UseWithPolicy`/`SyncWithPolicy` →
`materializeOne`, `Resolve` → `repairUnderLock`:

- unmanaged stop: `TestUseWithoutTakeoverRefusesUnmanaged`
- takeover backup + notice: `TestUseTakeoverBacksUpAndNotifies`,
  `TestRepairTakeoverProvisionsManagedHome`
- foreign stop + choice: `TestForeignSymlinkStopsSwitch`
- heuristic warning: `TestTakeoverWarnsDotfileHeuristic`
- read-only never onboard: `TestReadOnlyCommandsNeverOnboard`
- CLI flags: `TestProfileUseTakeoverRow`,
  `TestResolveTakeoverRequiresRepair`, `TestSyncTakeoverFlagParses`

Import (§9.6) — call site `Import` → `installLocked`:

- lossless path pipeline: `TestImportLosslessInstallsThroughPathPipeline`
- normalization + order: `TestImportNormalizesOrdersAndPins`,
  `TestNormalizeImportText`
- loss list + consent: `TestImportLossyStopsWithoutConsent`,
  `TestImportLossyProceedsWithConsent`, `TestImportUnreadableRootIsLoss`,
  `TestImportInvalidUTF8IsLoss`
- name-taken before write: `TestImportNameTakenBeforeAnyWrite`
- partial activation keeps the lock:
  `TestImportPartialActivationKeepsLock` (fresh machine, unmanaged
  detected file → `profile_use_partial`, lock installed)
- revision pinning + foreign warning:
  `TestImportSkillForeignPinnedByRevision`,
  `TestImportSkillMarkerRecovery`
- `imported_from_native`: `TestImportRecordsImportedFromNative`
- CLI rows: `TestProfileImportRow`, `TestProfileImportLossyRow`,
  `TestProfileImportNameTakenRow`, `TestProfileComposeAddPathRow`

Schema 2 (§12.1, §12.2): increment 1 plus `TestPolicyFromConfigCarriesEnvGates`
(overlays, default weight, precedence carried; forbidding empties),
`TestPathOverlayDeclarationParses`, and the full schema-case/vector
conformance below.

## Gate table (exact standalone commands, observed exit codes)

Root A (default lane): materialized pin
`git -C <spec> archive 0ed5c691e9208eea52f21db2fc05e226ce3516fd
conformance/v1 | tar -x -C /tmp/spec-pin-root` (exit 0).
Root B (candidate lane): curator-spec main worktree at `550579d`,
`.../worktree/conformance/v1`.

- `go build ./...` → exit 0
- `go vet ./...` → exit 0
- `gofmt -l cmd internal` → empty, exit 0
- `golangci-lint run ./...` → 0 issues, exit 0 (an earlier
  `./internal/...` run exited 1 on three revive notes — package-comment
  headers and one unused parameter — all fixed and re-verified)
- `bash .github/ci/gate-selftest.sh` → 81 passed, 0 failed, exit 0
- `bash .github/ci/no-broad-suppression.sh` → exit 0
- `bash .github/ci/ledger-consistency.sh <evidence>` → 177+ rows ok,
  exit 0 (rerun green after every ledger append)
- `go test -count=1 -race` on the new envprofile masks → exit 0
- `go test -count=1 -race ./internal/config/ ./internal/contextstore/` → exit 0
- `go test -count=1 -run '<CLI rows>' -race ./cmd/curator/` → exit 0
- `go test -count=1 -timeout 30m ./internal/envprofile/ ./internal/config/
  ./internal/contextstore/ ./internal/contextresolve/
  ./internal/contextmaterialize/` → exit 0
- `CURATOR_CONFORMANCE_ROOT=<rootB> go test -count=1 -run
  'TestManagerConfigV2|TestSystemConfigV2' ./internal/config/` → exit 0
- `CURATOR_CONFORMANCE_ROOT=<rootB> go test -count=1 -run
  'TestConformanceContextResolution|TestConformanceEnvironmentsHeader'
  ./internal/interop/` → exit 0
- `CURATOR_CONFORMANCE_ROOT=/tmp/spec-pin-root/conformance/v1 bash
  .github/ci/test-gate.sh /tmp/lane-default-evidence` → attempt 1: go
  test exit=0, platform-case gate exit=1 — the tree moved under the run
  (three tests committed after the lane's go-test compiled), so the
  ledger named cases the stream never ran. No test failed. Re-run
  against the final tree below.
- `CURATOR_CONFORMANCE_ROOT=<rootB> CI_REQUIRE_FULL_ROOT=1 bash
  .github/ci/test-gate.sh /tmp/lane-candidate-evidence` → attempt 1:
  same freshness failure on one later test; go test exit=0. Re-run below.
- `CURATOR_CONFORMANCE_ROOT=/tmp/spec-pin-root/conformance/v1 bash
  .github/ci/test-gate.sh /tmp/lane-default-evidence2` → go test
  exit=0, gate exit=1 on one row: `TestFoldedPathsAreSourceInvalid`
  skips on case-folding filesystems (this Mac) but the ledger tolerated
  no skip, and the reason matched no skip class. Real defect, fixed:
  reasons reworded into the classified vocabulary, fold/symlink/unreadable
  rows carry host-capability tolerance where the host cannot provide the
  fixture. Verified on a partial envprofile stream (every envprofile row
  ok), full re-run below.
- `CURATOR_CONFORMANCE_ROOT=<rootB> CI_REQUIRE_FULL_ROOT=1 bash
  .github/ci/test-gate.sh /tmp/lane-candidate-evidence2` → same single
  row, same fix, re-run below.
- `CURATOR_CONFORMANCE_ROOT=/tmp/spec-pin-root/conformance/v1 bash
  .github/ci/test-gate.sh /tmp/lane-default-evidence3` → exit 0:
  `suite-plan: served=69 deferred=3 excluded=0` (`defer
  internal/config` — the pin predates the environments families, so the
  root-needing cases defer under their registered tolerance while the
  hermetic stage (c) rows run), `stage served exit=0`, `stage deferred
  exit=0`, `platform-case gate: ok`.
- `CURATOR_CONFORMANCE_ROOT=<rootB> CI_REQUIRE_FULL_ROOT=1 bash
  .github/ci/test-gate.sh /tmp/lane-candidate-evidence3` → exit 0:
  `go test exit=0, platform-case gate exit=0` (full root, fails closed).
- `go test -count=1 -timeout 30m ./cmd/curator` → exit 0
  (`ok ... 282.345s`, zero `--- FAIL`) on the final tree `8560145`.
- Hosted CI was not consulted: no `gh pr checks` output exists for this
  run (nothing pushed, per the brief). The two local lane reproductions
  above stand in for it.

## Narrowing mutants (gate stays present, weakened to admit one member)

| Mutant | Narrowed gate | Named test that fails | Survivor / bound |
|---|---|---|---|
| M1 drop path-form check in `resolveOverlay` | formed-path overlay admitted | `TestPathOverlayFormIsSourceInvalid` (`err = <nil>`) | killed |
| M2 skip nested-`.git` only in `copyStateTree` | nested `.git` admitted; symlink/root-`.git` still refused (proven: those tests pass under M2) | `TestNestedGitIsSourceInvalid` | killed |
| M3 collapse absence→unreadable in `EnsureState` | — | no test failed | SURVIVOR: `pathManifestDiag` precedes the store, so the store branch fires only on a manifest/snapshot race (TOCTOU). Stated bound. |
| M3r collapse absence→invalid in `pathManifestDiag` | missing operand admitted as invalid | `TestMissingPathOperandIsPathMissing`, `TestMissingOverlayPathIsPathMissing` | killed |
| M4 skip foreign-symlink stop only | foreign symlink taken without flag; plain-unmanaged still refused (proven under M4) | `TestForeignSymlinkStopsSwitch` | killed |
| M5 consent bypass for exactly one loss | single-loss import proceeds | `TestImportLossyStopsWithoutConsent` (`err = <nil>`) | killed |
| M6 name gate only for default names | explicit-`--as` taken name proceeds (falls through to `profile_name_taken`, proving the distinct diagnostic) | `TestImportNameTakenBeforeAnyWrite` | killed |
| M7 downgrade two-requirer conflicts to warning | two-requirer disagreement admitted | `TestWeightConflictErrorAndWarning/error` (`err = <nil>`) | killed |
| M8 pin placement to winner-last | winner-first emission misordered; winner-last + tie pass under M8 | `TestPrecedencePrimitivesDriveEmission` (exactly the two winner-first subtests) | killed |
| M9 composition gate only for root repeats | overlay-vs-overlay repeat admitted (dies later at source agreement — defense in depth documented) | `TestOverlayRepeatedNameIsCompositionInvalid` | killed |
| M10 report failed walk as invalid | unreadable source admitted as invalid | `TestUnreadablePathIsUnreadable` | killed |
| M11 treat every `.git` entry as root-level (token search preserved, behavior changed) | nested `.git` admitted; `TestRootGitExcluded` passes under M11 | `TestNestedGitIsSourceInvalid` (`err = <nil>`) | killed; the behavioral suite (both tests) executes, not only a static check |

## Lessons compliance (stages a/b, not re-paid)

- Pinned root vs main: no new test reads a conformance family, so no
  new `root-artifacts.tsv` row was needed (verified by grep: zero
  `CURATOR_CONFORMANCE_ROOT` reads in the new files). All 49 new ledger
  rows are hermetic (`- -`, required on all three runners).
- Windows fixtures: swept every new fixture for platform-absolute
  literals and `filepath.IsAbs`-style assumptions. Findings: (1) the
  fold-collision fixture cannot observe two entries on case-folding
  filesystems — the test probes and skips there, running on
  case-sensitive lanes; (2) symlink-creation tests skip when the
  platform refuses; (3) permission-bit unreadability skips on Windows
  and under superuser; (4) `pinOperatorHome` sets both `HOME` and
  `USERPROFILE` so the opencode global surface resolves under test
  control on Windows too. No `filepath.IsAbs` on fixtures, no backslash
  literals in values.
- Uncalled gates: every new refusal is driven through `Install`,
  `UpdateWithPolicy`, `UseWithPolicy`, `SyncWithPolicy`, `Resolve`, or
  `Import` — never a helper alone — and each carries a narrowing
  mutant above.
- Gate table honesty: every row above is a standalone process with its
  observed exit code. The one mid-run lint failure (exit 1) is reported
  as failing with its fix. Pending rows stay pending until they land.

## Stated bounds

- `requires` never names a path source structurally (requirement
  entries always carry git identity); no dedicated test.
- Secondary fixed-home targets join inventory/backup through the
  existing switch seams; their divergent-file comparison has no probe
  path in revision 1 (targets name roles, not paths), so import treats
  them as the switch does.
- The import audit is the shared install always-strict path; no
  import-specific secret-material vector was added.
- Hard-link rejection is inherited from the store discipline (unchanged
  code); symlink, nested-`.git`, collision, and non-directory carry the
  new tests.
- The opencode global skills surface resolves below the operator home
  (`~/.agents/skills`); ledgered entries (`.csk-managed.json`) and
  store-target symlinks belong to the §9.4 migration path, never import.
- Machine configuration cannot pre-record import consent structurally
  (no knob feeds `AllowLossy`); `Policy.Takeover` is likewise
  operation-scoped and never set from configuration.
- `TestFoldedPathsAreSourceInvalid` skips on case-folding filesystems;
  the Linux lanes exercise it.
