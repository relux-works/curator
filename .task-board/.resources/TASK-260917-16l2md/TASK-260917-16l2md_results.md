# TASK-260917-16l2md results — SPEC_PIN → v1.0.0-rc.12 with the wave-1 manager union (E2/E4/S4)

Story `STORY-260917-3w3lvj`, role developer. Worktree: the managed Story worktree (uncommitted;
Change Request published at handoff). Conformance root for every run below:
`/tmp/spec-rc12/conformance/v1` — a read-only worktree of curator-spec at exactly
`dced9b8317e0e8af79edf2d0539b32bd22b6c85b`.

## 1. What this candidate is

One candidate that moves `SPEC_PIN` in `.github/workflows/ci.yml` from
`87a0d0060bad64ab883d007dcdf35df7485368bf` (rc.11) to
`dced9b8317e0e8af79edf2d0539b32bd22b6c85b` (rc.12) together with the UNION of
the three held wave-1 manager candidates:

- E2 `TASK-260916-55g9dg` rev-3 (`E2-candidate.patch`, base `1de6f8e`):
  transitive `class: system` modules — `transitive_system_modules`
  drop|error, `system_module_waivers`, `context_system_module_transitive`
  refusal, posture.
- E4 `TASK-260916-3oh0u8` rev-1 (`E4-candidate.patch`, base `1de6f8e`):
  provider trust roots — `provider_directories`, revision A warning / B
  refusal, resolved path in `env status`.
- S4 `TASK-260910-gocke2` rev-1 (`S4-candidate.patch`, base `c64ceaf`):
  MCP env passthrough bounds — `passable_env_names` default `[]`
  (explicit `null` = unbounded), `mcp_package_allowlist_empty` warning,
  `s4-warn`/`s4-enforce` profiles, MCP surfacing rows.

Every suite checkout (5 jobs) uses `ref: ${{ env.SPEC_PIN }}`; the only
other `ref:` in ci.yml is the sanctioned `inputs.candidate_ref` of the
non-default `candidate-conformance` job. `release.yml` carries no suite
pin (out of scope, verified).

The rc.12 manifest digest recorded in the ci.yml comment,
`sha256:ea9dd5a0030b889079cf5655517056e66a0f1d61812f8c2920b4beddc1fd24ed`,
is byte-verified: it is the sha256 of the rc.12 root's
`conformance/v1/manifest.json` AND the value the root's own
`release/1.0.0-rc.9.json` requires in both
`downstream_consumption.required_manifest_sha256` and
`candidate_protocol_pin.manifest_sha256`. Protocol stays `1.0.0-rc.9`.

## 2. Per-candidate merge notes (union reconciliation)

Method: each patch was first applied onto its own base in a scratch tree
(`/tmp/pinmerge/cand-e2`, `cand-e4`, `cand-s4`) to hold the exact
candidate bytes; then applied sequentially onto the Story worktree with
`git apply --3way` (E2, then E4, then S4), resolving every conflict as
the union. `TASK-*_results.md` payloads inside the E4/S4 patches were
excluded from application (`--exclude`) — they are board evidence, not
repository content; the worktree carries no such files.

After the merge, every overlapping file was machine-checked: each
whitespace-normalized line the candidate added against its base must be
present in the union (`/tmp/pinmerge/unioncheck.sh`). E2, E4 and S4 all
verify with zero missing lines except the intentional union rewrites
listed below.

### E2 → E4 overlaps (3 files, 8 conflict hunks)

- `internal/config/config.go` (1 hunk): `LockableKeys` — E2 added
  `environments.transitive_system_modules`, E4 added
  `environments.provider_directories`. Union carries both (8-key map
  with the six pre-existing keys).
- `internal/config/environments.go` (3 hunks):
  - `LockableEnvKeys`: same two-key union.
  - `defaultEnvironments()`: E2's `TransitiveSystemModules: "drop"` +
    `SystemModuleWaivers: []` plus E4's `ProviderDirectories: []`, each
    kept at its candidate's relative position.
  - `EnvLockKey`: the single-knob `case` arms union into one arm
    listing both `transitive_system_modules` and
    `provider_directories`.
  - The effective-JSON `render()` hunk merged cleanly via 3-way and
    already carries all three keys.
- `internal/config/environments_test.go` (4 hunks):
  - `payloads` map: E2's two payloads + E4's `provider_directories`
    payload (22 payloads, matching the 22-name `EnvKnobNames`).
  - §12.2 transcription comment + map: rewritten as the 8-key union in
    spec order, verified against the landed §12.2 sentence at rc.12
    (`protocol/environments.md` §12.2 lists exactly
    `overlays_allowed`, `precedence`, `mcp_package_allowlist`,
    `passable_env_names`, `require_current_profile`,
    `transitive_system_modules`, `isolation`, `provider_directories`).
  - Rendered-knob list: both candidates' knobs.
  - Test-function interleave: E2's four tests kept complete, then E4's
    four tests — the shared trailing `}` pair was split so each
    function closes (E2 `TestSystemTransitiveDirection` tail +
    E4 `TestProviderDirectoriesWriteRoundTrip` tail verified against
    candidate bytes).

Auto-merged without conflict (verified byte-complete by the line check):
`platform-cases.tsv` (both rows), `CHANGELOG.md` (E4 entry prepended
under its own `### Added` heading — see §2.1 — plus E2's appended
entry), `cmd/curator/envstatus.go`, `internal/envprofile/status.go`.

### E2+E4 → S4 overlaps (6 files, 8 conflict hunks)

- `CHANGELOG.md` (1 hunk): E2's and S4's appended entries union
  verbatim (E2 then S4); see §2.1 for the heading repair.
- `cmd/curator/envstatus.go` (1 hunk): E4's provider helpers
  (`describeProvider`/`derefProvider`/`attachProviderPosture`) kept
  complete, then S4's `formatPassable` — shared trailing `}` split as
  above.
- `internal/config/environments.go` (1 hunk, substantive): the
  `passable_env_names` rendering. E4's `providerDirs` block is kept and
  S4's `PassableEnvNamesSet`-aware rendering supersedes the pre-S4
  block (absent → `[]`, explicit null → `null`, list → itself), exactly
  S4's bytes including the §12.1 comment.
- `internal/config/environments_test.go` (1 hunk): E4's last test kept
  complete, then S4's `TestPassableEnvNamesPresence` — shared `}` split.
- `internal/envprofile/envprofile.go` (1 hunk, substantive): the
  Install pre-publish slot. E2's `checkAdmissionPrePublish` refusal
  runs first (fail fast), then S4's §9.1 surfacing +
  `AllowlistEmptyWarning` block verbatim. Observable behaviour is order-
  independent: on E2 refusal the function returns the error either way
  (S4 rows/warnings only surface via `Info` on success, since
  `surfacing` is collected into `Info.Surfacing` for the CLI layer, not
  printed inline); on success both run. No spec text orders them.
- `internal/envprofile/status.go` (2 hunks): import union
  (`contextlock`, `contextmaterialize`, `contextpkg` + S4's
  `envfragment`, gofmt-sorted) and `Status` struct union (E4's
  `Providers` + S4's `S4Profile`/`PassableEnvNames`/`Warnings`/
  `MCPDeclarations` with S4's comments verbatim).

Auto-merged without conflict (line-check clean): `platform-cases.tsv`
(S4 row), `cmd/curator/env.go` (S4 `machineFromConfig` + E4
`attachProviderPosture` call site — disjoint lines), `env_test.go`,
`profile.go`, `mcp.go`, `envfragment.go`, `managed.go`,
`envregistry.go`, plus S4's new files.

### §2.1 CHANGELOG structure repair

E4's patch prepends its entry under a NEW `### Added` heading above the
trunk's existing `### Added` block, which would leave two identical
section headers under one `## Unreleased`. The union keeps one
`### Added` section (E4 entry first, then trunk entries, E2, S4, pin
note); every entry's text is byte-identical to its candidate. No
diagnostic or knob spelling was touched.

### Single-candidate files

Every file touched by exactly one candidate and untouched by trunk
drift since that candidate's base is byte-identical to the candidate
(`cmp`-verified): E2's `admission.go`, `admission_test.go`,
`system_module_admission_test.go`, `admission_test.go` (envprofile),
`system_module_schema_test.go`, `environments_conformance_test.go`,
`contextmaterialize.go`, `managed.go` (pre-S4 state),
`context_materialization_test.go`; E4's `umbrella.go`, `env.go`,
`umbrella_test.go`, `umbrella_conformance_test.go`. Two files carry
trunk drift past the candidate base (`internal/envprofile/envprofile.go`
1 commit, `cmd/curator/main.go` 3 commits) and merged via 3-way;
both build and their suites pass (see §5).

## 3. SPEC_PIN families served; root-content rows and skips removed

rc.12 publishes every family the candidates' `root-content` drivers
cover (verified against `/tmp/spec-rc12`):

| Family | Driver | rc.12 content |
|---|---|---|
| E2 schema cases (7) | `internal/config TestSystemModuleSchemaSubset` | 7/7 instances in `schema-cases/index.json` |
| E2 admission cases (5) | `internal/contextmaterialize TestSystemModuleAdmissionVectors` | 5/5 `system-module-*` in `vectors/environments.json` `materialization_cases` (24 total) |
| E4 umbrella vectors (14×2 rev) | `cmd/curator TestUmbrellaProviderResolutionVectors` | `vectors/umbrella-provider-resolution.json`, 14 cases |
| S4 passthrough vectors | `internal/envprofile TestEnvironmentsEnvPassthroughVectors` | `vectors/environments-env-passthrough.json` (7 resolution + 6 allowlist + 4 schema + 6 surfacing + 2 order) |

Removed per the brief ("remove the row and the skip — the family is
served"):

- The 4 `.github/ci/platform-cases.tsv` rows the candidates added. The
  ledger is now byte-identical to trunk (`cmp` against `HEAD`).
- The 6 skip branches in the 4 drivers: each `os.IsNotExist →
  t.Skipf("…publishes no…")` becomes a fail-loud `t.Fatal(err)`, and
  each subset-empty `Skipf` is deleted so the neighbouring
  partial-publication `Fatalf` ("publishes only 0 of N …") covers the
  empty case. Stale comments describing the root-content path were
  updated to state the pin serves the family. The `CURATOR_CONFORMANCE_ROOT
  is not set` skips (class `root-unset`, a different class) are kept —
  the hosted lanes always set the root.

Why deletion, not tolerance-retirement: the repo precedent (`04550e2`)
retires a tolerance by converting must-run rows to `-`/`-`, but those
rows pre-existed as must-run requirements; the 4 rows here were added by
the candidates purely to tolerate the skips, and the brief orders their
removal twice ("remove the row and the skip"). Enforcement does not
weaken: with the skip code deleted, an absent family fails the test and
fails the lane. `gate-selftest.sh` passes 180/180 on the shipped ledger
(see §5); its `PROMOTED_PACKAGES` check for `internal/config` is green
again (zero tolerated rows there).

Deliberately NOT extended: `PROMOTED_PACKAGES` in `gate-selftest.sh`
(per-package "tolerates no skip" would also constrain legitimate
`sibling host-capability`/`platform-control` rows in `cmd/curator` and
`internal/envprofile`, so the per-package granularity does not fit this
delta; the bound is stated here rather than inferred). `root-artifacts.tsv`
is untouched: it declares none of the new families, so `suite-plan.sh`
defers nothing at either pin for these packages.

No `shell-hook-trust` driver exists in the repository (no consumer of
`vectors/shell-hook-trust.json`, no skip, no row) — nothing to remove;
the family is published by the root and un-driven, which is out of scope
for this union (a later story's driver).

## 4. Acceptance criteria — how each line is met

- Hosted gate green at `dced9b8`: narrow gates green locally at the
  exact rc.12 root (see §5); the full landing suite runs once at
  handoff via `scripts/remote-gate.sh` (not run locally per the rules).
- No root-content skip for a published family: the 6 skip branches
  above are deleted; `grep` confirms no `publishes no` skip remains in
  the 4 drivers; ledger carries no `root-content` row for them.
- No weakened test: every removal converts skip → fail; vector
  comparisons stay exact (E2 rev-2 correction 1 verified: no
  `prunePostRevisionKnobs`/`postPinKnobs` anywhere in
  `internal/config/environments_conformance_test.go`).
- E2 behaves per landed spec: drop default + error opt-in, waivers,
  lock-to-error-only, `context_system_module_transitive` refusal with
  unchanged lock, posture rows — candidate bytes preserved (§2);
  `TestManagerConfigV2Vectors` (48 cases), `TestSystemModuleSchemaSubset`
  (7), `TestSystemModuleAdmissionVectors` (5), interop monolithic cases
  all execute and pass.
- E4 behaves per landed spec: revision A shipped (warn
  `subcommand_provider_outside_trust_roots`), revision B implemented
  behind the same option, `provider_directories` knob, trust verdicts
  in `env status` — 14 cases × 2 revisions execute and pass.
- S4 behaves per landed spec: `s4-warn` shipped (absent knob unbounded
  with `mcp_env_passthrough_unlisted` warning), `s4-enforce`
  implemented behind the same option, explicit null unbounded and
  silent, `mcp_package_allowlist_empty`, surfacing rows at
  install/update/status — all 5 vector groups execute and pass.
- E2 rev-2 corrections stay closed: (1) no vector-comparison workaround
  (`grep` clean, exact comparison, suite green at the new root so no
  pin-lag failure remains); (2) explicit null waiver list refused —
  `internal/config/environments.go` parses a present
  `system_module_waivers` unconditionally (no nil bypass) with
  regression `TestSystemModuleWaiversNullRejected` passing; (3) the
  dedicated admission/schema subset drivers with explicit accounting
  remain, their root-content skips removed exactly as the verdict's
  pin-promotion path prescribes ("coordinate the sibling
  integration/root qualification" — this task is that qualification).
- CHANGELOG carries the three entries (each naming its warn-first
  step: E2 drop-default/error-opt-in, E4 revision A with B later, S4
  s4-warn with s4-enforce later) plus the conformance-pin note.

## 5. Validation transcripts (all at CURATOR_CONFORMANCE_ROOT=/tmp/spec-rc12/conformance/v1)

Static gates (exit 0 throughout):

- `go build ./...` → exit 0
- `go vet ./...` → exit 0
- `gofmt -l internal cmd` → no output, exit 0
- `bash .github/ci/gate-selftest.sh` → `gate-selftest: 180 passed, 0 failed`, exit 0
- `golangci-lint run` over the 10 union-touched package trees
  (`internal/config`, `contextresolve`, `contextmaterialize`,
  `contextaudit`, `shell`, `envprofile`, `envfragment`, `envregistry`,
  `interop/environments`, `cmd/curator`) → `0 issues.`, exit 0
- `bash .github/ci/ledger-consistency.sh` → `239 rows checked …
  ok`, exit 0

Narrow suites (brief list + two touched extras):

- `go test -count=1 ./internal/config/...` → ok 2.987s, exit 0
  (48 manager-config-v2 cases + 7 schema-subset cases execute)
- `go test -count=1 ./internal/contextresolve/... ./internal/contextmaterialize/... ./internal/contextaudit/...` →
  ok 0.825s / 1.706s / 2.413s, exit 0 (5 admission cases execute)
- `go test -count=1 ./internal/shell/...` → ok 1.377s, exit 0
- `go test -count=1 ./internal/envprofile/...` → six `-run` chunks
  partitioning all 171 tests (29/29/29/29/29/26, anchored exact-match
  regexes, `-timeout 480s` each, run strictly sequentially), exit 0 each:
  chunk1 ok 87.279s, chunk2 ok 77.387s, chunk3 ok 87.434s,
  chunk4 ok 116.458s, chunk5 ok 96.021s, chunk6 ok 176.770s.
  Partition verified: 171 unique names across the six chunk files.
- `go test -count=1 ./cmd/curator/...` → four `-run` chunks
  partitioning all 214 tests (54/54/54/52), except the three
  GOROOT-hashing `TestCompiledProject*` tests, which run solo with
  `-timeout 500s` each (timings below). Those three spend
  minutes in `godriver.digestToolchainRecords` (whole-GOROOT
  fingerprint via `install.Project`), a path no candidate touches
  (verified: no `godriver`/`install` file in any patch; no union
  symbol on the timeout stacks) and equally slow on the pristine
  `HEAD` baseline — host-load slowness, also recorded in the E4/S4
  candidates' own results notes. Timings, exit 0 each:
  `TestCompiledProjectStatusAndUntrustedRecovery` ok 283.403s (union)
  vs ok 485.310s (baseline, `--- PASS` 313.90s);
  `TestCompiledProjectRepairsCorruptCompiledState` ok 188.050s;
  `TestCompiledProjectRestoresCacheWhenCommitFails` ok 203.840s;
  chunk-1-rest (51 tests) ok 253.177s; chunk-2 (54) ok 291.361s;
  chunk-3 (54) ok 273.589s; chunk-4 (52) ok 135.805s.
  Partition verified: 3 + 51 + 54 + 54 + 52 = 214 unique names.
- Extras (touched by the union, outside the brief's narrow list):
  `go test -count=1 ./internal/interop/environments/...
  ./internal/envfragment/...` → ok 2.337s / 1.292s, exit 0.

Why chunks: a single `go test ./internal/envprofile/` exceeds the 600s
default timeout on this loaded host (cumulative ~640s of git-fixture
tests, ~2–12s each — a host-load characteristic, proven by running the
same tests on a pristine `HEAD` baseline copy: per-test timings match
the union, e.g. `TestInstallUseSwitchesAndAgrees` 6.45s baseline vs
6.99s union). No hang: the verbose log shows every test passing in
sequence until the binary timeout fires. Chunks ran sequentially with no
concurrent load; each chunk log is `ok` with exit 0.

Vector-execution proof (each `-v` run below: exit 0, every `--- PASS`,
zero `--- SKIP`):

- `TestManagerConfigV2Vectors`: 49 PASS (48 subtests + parent)
- `TestSystemModuleSchemaSubset`: 8 PASS (7/7 cases + parent)
- `TestSystemModuleAdmissionVectors`: 6 PASS (5/5 cases + parent)
- `TestUmbrellaProviderResolutionVectors`: 15 PASS (14 cases × 2
  revisions + parent)
- `TestEnvironmentsEnvPassthroughVectors`: 26 PASS (7 resolution —
  incl. `s4-warn-absent-warns-every-passed`,
  `s4-enforce-absent-drops-all`, `explicit-null-unbounded`,
  `s4-enforce-absent-passes-unlisted`,
  `explicit-null-treated-as-empty` — + 6 allowlist + 4 schema + 6
  surfacing + 2 order + parent)
- `TestConformanceEnvironmentsMonolithic` (interop): 25 PASS (24 cases
  + parent), incl. all five `system-module-*` subtests PASS.

## 6. Profiles shipped (warn-first)

- E2: `transitive_system_modules` default `drop` (non-breaking),
  `error` opt-in; no A/B split per the landed spec (review verdict
  confirms the generic warn-first wording imposes none here).
- E4: revision A shipped (PATH selection + outside-roots warning);
  revision B implemented behind the same option, flip in a later release.
- S4: `s4-warn` shipped (absent knob unbounded + warning);
  `s4-enforce` implemented behind the same option, flip in a later release.

## 7. Out of scope / stated bounds

- E1 manager, R1 client, S6 story, ax integration, proposals 0014–0018,
  tags/releases, `release.yml` — untouched per the brief.
- `docs/`: none of the three patches documents knobs under `docs/`
  (verified from the patch file lists); the only knob mention in
  `docs/` is the historical audit finding text. Nothing to union.
- `PROMOTED_PACKAGES` in `gate-selftest.sh`: not extended (see §3).
- `root-artifacts.tsv`: untouched (see §3).
- No new tests were authored for the merge itself: the union reuses the
  candidates' committed tests verbatim (plus the two small union-only
  test edits in §2: the 8-key §12.2 transcription and the joined knob
  lists), and the rc.12 vectors are the coverage.
- Spec gaps found: none. One ordering judgement is recorded in §2
  (Install pre-publish: E2 refusal before S4 surfacing; observably
  equivalent either way).

## 8. Files changed (34; +5666/−209 per diffstat)

Pin: `.github/workflows/ci.yml`. Ledger: no diff (4 candidate rows
removed → byte-identical to trunk). Release notes: `CHANGELOG.md`.
Code: `cmd/curator/{env,envstatus,main,profile,umbrella}.go`,
`internal/config/{config,environments}.go`,
`internal/contextmaterialize/{admission,contextmaterialize,mcp}.go`,
`internal/envfragment/envfragment.go`,
`internal/envprofile/{envprofile,managed,status,surfacing}.go`,
`internal/envregistry/envregistry.go`. Tests: the 4 vector drivers
(skip removal), `internal/config/environments{,_conformance}_test.go`,
`internal/contextmaterialize/{admission,system_module_admission}_test.go`,
`internal/envprofile/{admission,envpassthrough_conformance,surfacing}_test.go`,
`internal/envfragment/envfragment_test.go`,
`cmd/curator/{env,envconfig,profile_surfacing,umbrella}_test.go`,
`internal/interop/environments/context_materialization_test.go`.

## 9. Recovery rev2 — Windows hosted-gate failure fix (this run)

Rev1's Change Request failed CR validation: hosted run 35266385103,
`Test (windows-latest)` only, one test —
`cmd/curator TestEnvStatusMatrix` (`env_test.go:136`, `status --check
after repair = 1`). The provider rows read
`(subcommand_provider_untrusted, non-current)` for `curator-run.exe`
and `curator-session.exe` under a `RUNNER~1` (8.3) temp path. Every
other lane, every vector suite, lint, and the gate self-test were
green. This section records the diagnosis, the fix, and the
re-verification; §§1–8 still describe the union, unchanged except the
files listed in §9.6.

### 9.1 Root cause (refines the orchestrator's mechanism)

Under the shipped revision A, `subcommand_provider_untrusted` fires
from exactly one site — `refuseDir`
(`cmd/curator/umbrella.go`, the `publishedDirs`/`managedRoots`
match) — i.e. the provider was judged inside a manager-published or
managed directory, not merely outside the trust roots. The matched
entry is the `globalbins.Select` user-bin choice, which
`providerInputsForHost` treated unconditionally as a refused
manager-published directory. Select scans PATH for the first safe
writable directory below the user home; on Windows CI `%TEMP%`
(`C:\Users\RUNNER~1\...`) sits below `%USERPROFILE%`, so the test's
stub-provider `bin` (prepended to PATH) was selected as "the user-bin
shim directory" and both providers refused. On macOS/Linux temp trees
live outside `$HOME`, the scan skips `bin`, revision A warns, and
`--check` stays green — hence Windows-only red.

The 8.3 spelling in the failed rows is display-only (`mustAbsolute`,
uncanononicalized). It did not cause this refusal: both comparison
sides already pass through `filepath.EvalSymlinks`, which on Windows
expands 8.3 → long and normalizes case per component (verified in
GOROOT `src/path/filepath/symlink_windows.go`: `evalSymlinks` →
`toNorm`/`normBase` via `FindFirstFile`, documented "must be
unique"). Real spelling gaps remain where `EvalSymlinks` does not
normalize (case variants on insensitive unix volumes) and in its
failure fallback, so the ordered canonical-identity hardening is
implemented below regardless — it is necessary hygiene, but the
Select/scanned-bin refusal is the gate fix.

### 9.2 The fix (E4 code + a minimal globalbins seam)

1. **Refuse the selected shim dir only once the manager publishes
   there.** `providerInputsForHost` now adds the `Select` result to
   `publishedDirs` only when `userBinShimRefused`
   (`cmd/curator/umbrella.go:443`) holds: the selection is explicit
   (`Selection.Explicit`, the operator-declared publishing location)
   or `globalbins.PublishedShims` reports the ownership ledger
   (`internal/globalbins/globalbins.go:404`; presence counts, missing
   ledger = unpublished, any other read failure fails closed). A
   merely scanned PATH entry holds no manager-written content, so
   providers there warn under revision A instead of refusing. Gate
   site: `cmd/curator/umbrella.go:393`.
2. **Honor the explicit override.** `pathEnvironment`
   (`cmd/curator/umbrella.go:450`) now carries
   `CURATOR_GLOBAL_USER_BIN` to `Select`; previously the host glue
   passed a PATH-only map, so an explicit shim directory was silently
   ignored (a stricter pre-existing bug in the other direction).
3. **Canonical identity in comparisons.** `sameDir`
   (`cmd/curator/umbrella.go:454`) keeps its string fast path, then
   falls back to `os.SameFile` identity; `underDir`
   (`cmd/curator/umbrella.go:475`) keeps its prefix fast path, then
   walks ancestors by identity (the `internal/staging` `Within`
   pattern). Inspection failures resolve toward the revision-A
   warning, never toward silent trust; the spec's fail-closed sites
   stay the explicit trust-root readability checks, and no invented
   diagnostic was added to the closed §11.1 set.
4. **`Selection.Explicit`** (`internal/globalbins/globalbins.go:33`,
   set at :149) marks operator-declared selections. The globalbins
   delta is purely additive (one field, one function); `Select`'s
   `Path`/`Warning` outputs are bit-identical, so
   `internal/install` (`StageForwarding`) behavior is unchanged.

### 9.3 Revision A verified (no code change)

`activeProviderRevision = providerRevisionA`
(`cmd/curator/umbrella.go:60`), pinned by
`TestActiveRevisionIsWarningRelease`. `providerPosture` resolves
under the active revision; warned rows stay `Current=true` while
refused/missing/unreadable rows set `NonCurrent`
(`cmd/curator/envstatus.go:146-158`), exactly the §12 currency rule
("the `subcommand_provider_outside_trust_roots` warning row stays
current"). The B-looking diagnostic under the A default is the
spec-mandated manager-published-directory refusal (§11: "under
revision A the refusal applies to the PATH-selected candidate") —
explained by §9.1, not a profile-option bug.

### 9.4 Tests added; authentic pre-fix failures observed

New tests (run BEFORE the umbrella fix, then after):

- `TestProviderInputsForHostScannedBinWarnsWithoutLedger`
  (`cmd/curator/umbrella_test.go:523`): stages the Windows shape on
  any platform (provider dir below a fake `$HOME`/`%USERPROFILE%`,
  PATH pointing at it, no ledger) and requires the revision-A
  warning. Pre-fix: FAIL (refused, `exit 1`); post-fix: PASS.
- `TestProviderInputsForHostPublishedBinRefuses` (:550): publishes
  via production `globalbins.Refresh`, then requires the refusal —
  pins the security side so the repair cannot overcorrect.
  PASS pre- and post-fix (characterization).
- `TestProviderInputsForHostExplicitBinRefuses` (:584): explicit var
  plus a preferred scanned dir; requires refusal of the declared
  dir. Pre-fix: FAIL (explicit ignored, warned); post-fix: PASS.
- `TestUmbrellaTrustRootCaseVariantIdentity` (:616): case-variant
  trust root resolves silently, case-variant managed root still
  refuses; skips honestly on case-sensitive volumes. Pre-fix: FAIL
  on this host (APFS insensitive — warned instead of silent);
  post-fix: PASS.
- `TestUmbrellaTrustRootShortSpellingIdentity`
  (`cmd/curator/umbrella_trustroot_windows_test.go:39`,
  `//go:build windows`): 8.3 root ↔ long PATH entry resolve
  silently in both directions; short managed root still refuses.
  Short spellings via `x/sys/windows` `GetShortPathName` (already a
  direct dependency; precedent in `internal/transaction`,
  `internal/staging`); skips honestly where the volume carries no
  8.3 alias. Compile-verified via `GOOS=windows go vet`; executes
  on the Windows hosted lane at handoff.
- `TestPublishedShimsTracksTheLedger`,
  `TestSelectReportsExplicit`
  (`internal/globalbins/globalbins_test.go`): ledger presence
  (incl. emptied) marks published; explicit vs scanned reporting.
  PASS.

`TestEnvStatusMatrix` itself is unchanged (no weakened test, no
changed expectation): it now passes because the product warns where
revision A says to warn.

### 9.5 Validation transcripts (this run, rc.12 root)

Conformance root for every run:
`/tmp/spec-rc12/conformance/v1` at exactly `dced9b8` (verified
`git rev-parse`). Shell `bash`, `set -o pipefail`; exit codes are the
gate commands' real statuses.

Static (exit 0 throughout):

- `go build ./...` → exit 0
- `go vet ./...` → exit 0
- `gofmt -l internal cmd` → empty, exit 0
- `GOOS=windows go build ./...` → exit 0
- `GOOS=windows go vet ./cmd/curator/ ./internal/globalbins/`
  (compiles the windows-only test) → exit 0
- `golangci-lint run ./cmd/curator/... ./internal/globalbins/...`
  → `0 issues.`, exit 0
- `bash .github/ci/gate-selftest.sh` → `180 passed, 0 failed`,
  exit 0
- `bash .github/ci/ledger-consistency.sh <evidence-dir>` → `239
  rows checked across linux darwin windows`, `ok`, exit 0 (ledger
  untouched by this delta)

Suites (all `-count=1`, exit 0):

- `go test ./internal/config/... ./internal/globalbins/` → ok
  3.666s / 1.296s
- `go test ./cmd/curator/` FULL: 218 tests partitioned 54/54/54/53
  (anchored exact-match regexes, `-timeout 480s` each, strictly
  sequential) plus the 3 GOROOT-hashing `TestCompiledProject*`
  solos (`-timeout 500s`): chunk ok 217.656s / 272.080s /
  291.864s / 82.379s; solos ok 221.522s / 160.487s / 123.727s.
  Partition verified exact (`cmp` of covered names vs
  `go test -list`, 218/218).
- `TestUmbrellaProviderResolutionVectors`: 15 PASS (14 cases + parent),
  0 FAIL, 0 SKIP.
- `TestEnvStatusMatrix`: ok 15.532s.

Accepted from the rev1 evidence (§5) without rerun, stated
explicitly: `contextresolve`, `contextmaterialize`, `contextaudit`,
`shell`, `envprofile`, `interop/environments`, `envfragment` suites.
This delta touches only `cmd/curator`, `internal/globalbins`, and
`CHANGELOG.md` (below); the globalbins change is additive-only, so
no other package can observe a behavior difference. The full landing
suite reruns once at handoff via `scripts/remote-gate.sh`.

### 9.6 Files changed in rev2 (6; union §§1–8 otherwise intact)

- `cmd/curator/umbrella.go`: ledger-gated refusal (§9.2.1),
  explicit env passthrough (§9.2.2), SameFile identity (§9.2.3).
- `cmd/curator/umbrella_test.go`: 4 regression/identity tests.
- `cmd/curator/umbrella_trustroot_windows_test.go` (new): 8.3 test.
- `internal/globalbins/globalbins.go`: `Selection.Explicit`,
  `PublishedShims`.
- `internal/globalbins/globalbins_test.go`: seam tests.
- `CHANGELOG.md`: one `### Fixed` E4 note (the warn-instead-of-refuse
  rule, the honored override, identity matching).

Checklist: all 10 items remain satisfied — the union (§§1–8) is
unchanged in behavior except the E4 refusal rule, which is now more
revision-A-faithful; the Windows lane failure is fixed at its root
cause with the ordered identity hardening and 8.3 coverage on top.
## 10. Recovery rev3 — hosted-gate failure fix, round 2 (this run)

Change Request revision 2 (run `35276210791`) fixed the rev1 Windows
`TestEnvStatusMatrix` failure — `TestManagerConfigV2Vectors`,
`TestEnvStatusMatrix`, lint, interop, and both macOS lanes are green at
`dced9b8` — but left three red lanes from two test-only defects, both in
the rev2-added E4 identity tests (no production file changes in rev3):

1. `Test (ubuntu-latest)` and `Race (ubuntu-latest)`:
   `platform-case gate: FAILED` —
   `FAIL  skip with an unrecognised reason on linux: cmd/curator ::
   TestUmbrellaTrustRootCaseVariantIdentity`
   (`reason: volume is case-sensitive: stat …/caseroot: no such file or
   directory`). The skip reason matched nothing in
   `.github/ci/skip-classes.tsv`: the case-semantics rows say `case
   sensitivity` (space) or require the `test filesystem …` /
   `temporary directory …` prefix, while the test printed `volume is
   case-sensitive` (hyphen). `go test` itself exited 0 on both lanes.
2. `Test (windows-latest)`: `go test exit=1, platform-case gate
   exit=0` — one test,
   `cmd/curator TestUmbrellaTrustRootShortSpellingIdentity`
   (`umbrella_trustroot_windows_test.go:70`):
   `resolved "C:\…\runneradmin\…\curator-run.exe", want
   "C:\…\RUNNER~1\…\curator-run.exe"`.
   The planted-provider expectation was spelled through `t.TempDir()`
   (8.3 short on that runner) while the lookup resolves through the
   long-spelled PATH entry; `strings.EqualFold` folds case but cannot
   equate 8.3 with long spellings. (Windows evidence artifact
   `test-evidence-windows-latest`, `test/go-test-served.json`: the only
   `--- FAIL` in the lane; `TestUmbrellaTrustRootCaseVariantIdentity`
   itself PASSES on Windows.)

### 10.1 The fix (2 test files)

- `cmd/curator/umbrella_test.go` (`TestUmbrellaTrustRootCaseVariantIdentity`):
  the skip reason is now `test filesystem is case-sensitive: %v`,
  which the gate classifies as `host-capability / allow` under its
  exact `reason ~ regex` logic (replicated locally with awk over the
  committed `skip-classes.tsv`: both new reasons classify
  `host-capability / allow`).
- `cmd/curator/umbrella_trustroot_windows_test.go`
  (windows-only, does not compile on other platforms):
  - the planted-provider comparison uses filesystem identity
    (`os.Stat` + `os.SameFile`, new `sameProviderFile` helper
    mirroring the production `sameDir` semantics) instead of
    `EqualFold`, in both directions (short-root/long-PATH and
    long-root/short-PATH);
  - the no-8.3-alias skip is now `this host cannot create 8.3
    short-name aliases for temp paths`, classified
    `host-capability / allow` (the old wording matched nothing and
    would have failed the Windows platform-case gate on any volume
    without short-name generation).

No production file changes; no ledger, pin, CHANGELOG, or vector-driver
changes. Every other `t.Skip` in the union's test files was audited
against `skip-classes.tsv`: `no bash`, `creating Windows symlink
requires host support`, `CURATOR_CONFORMANCE_ROOT is not set`
(`root-unset`, never fires where the lane sets the root), `(is|are)
exercised (on|by)`, and `no git on PATH` (never fires in CI — git is
always on PATH there) are all recognised or non-firing.

### 10.2 E2 rev-2 corrections re-verified (still closed)

- No `prunePostRevisionKnobs`/`postPinKnobs` anywhere in
  `internal/` or `cmd/` (`grep` clean); the manager-config-v2
  comparison stays exact.
- Explicit null `system_module_waivers` is refused:
  `internal/config/environments.go:295` parses a present key
  unconditionally (no nil bypass) and `parseSystemModuleWaivers`
  rejects non-lists with `environments.system_module_waivers / must
  be a list`; regression `TestSystemModuleWaiversNullRejected`
  passes.
- The admission subset drivers execute with no skip at the rc.12
  root (see §10.3 vector proof); the ledger stays byte-identical to
  trunk (no diff in `.github/ci/platform-cases.tsv`).

### 10.3 Validation transcripts (this run, rc.12 root)

Conformance root for every run:
`/tmp/spec-rc12/conformance/v1` at exactly `dced9b8` (verified
`git rev-parse`). Shell `bash`, `set -o pipefail`; exit codes are the
gate commands' real statuses.

Static (exit 0 throughout):

- `go build ./...` → exit 0
- `go vet ./...` → exit 0
- `gofmt -l internal cmd` → empty, exit 0
- `GOOS=windows go vet ./cmd/curator/` (compiles the fixed
  windows-only test) → exit 0
- `golangci-lint run ./cmd/curator/...` → `0 issues.`, exit 0
- `bash .github/ci/gate-selftest.sh` → `gate-selftest: 180 passed,
  0 failed`, exit 0
- `bash .github/ci/ledger-consistency.sh /tmp/ledger-ev` →
  `239 rows checked across linux darwin windows`, `ok`, exit 0
  (ledger untouched by this delta)

Suites (all `-count=1`, exit 0):

- `go test ./internal/config/... ./internal/contextresolve/...
  ./internal/contextmaterialize/... ./internal/contextaudit/...` →
  ok 3.562s / 1.777s / 1.329s / 2.279s, exit 0
- `go test ./internal/shell/...` → ok 2.724s, exit 0
- `go test ./internal/envprofile/...` → six `-run` chunks
  partitioning all 171 tests (29/29/29/29/29/26, anchored
  exact-match regexes, `-timeout 480s` each, strictly sequential),
  exit 0 each: chunk ok 98.649s / 73.725s / 65.210s / 138.347s /
  129.570s / 257.764s. (A single-package run exceeds the 600s
  default timeout on this host — `panic: test timed out`, zero
  `--- FAIL` — the rev1-documented host characteristic; chunks ran
  with no concurrent load.)
- `go test ./cmd/curator/...` → four `-run` chunks partitioning
  the 215 non-GOROOT tests (54/54/54/53, `-timeout 480s` each,
  strictly sequential) plus the 3 GOROOT-hashing
  `TestCompiledProject*` solos (`-timeout 500s`), exit 0 each:
  chunk ok 302.180s / 261.745s / 304.159s / 98.799s; solos ok
  232.587s / 169.525s / 162.468s.
  (A single-package run exceeds the 600s default timeout on this
  host — `panic: test timed out`, zero `--- FAIL` — the
  rev1/rev2-documented characteristic, unrelated to this delta,
  which touches no production code.)
- `go test ./internal/interop/environments/...` → ok 6.587s,
  exit 0
- `TestManagerConfigV2Vectors` + `TestSystemModuleWaiversNullRejected`:
  50 subtests run, zero FAIL/SKIP, package ok.
- `TestUmbrella*` (13 tests incl. `TestUmbrellaProviderResolutionVectors`
  and `TestUmbrellaTrustRootCaseVariantIdentity`): all PASS, ok 4.226s.

Partition-exactness: verified — envprofile 171/171 unique names
across the six chunk regexes vs `go test -list`; cmd/curator
215/215 across the four chunk regexes plus the 3 solos = 218/218.

Accepted from the rev1/rev2 evidence (§5/§9.5) without rerun, stated
explicitly: none — every narrow package above was re-run in this run.
The full landing suite reruns once at handoff via
`scripts/remote-gate.sh`.

### 10.4 Files changed in rev3 (2; union §§1–9 otherwise intact)

- `cmd/curator/umbrella_test.go`: skip-reason rewording (1 line).
- `cmd/curator/umbrella_trustroot_windows_test.go`: SameFile
  planted-provider comparison both directions + helper +
  skip-reason rewording.

Checklist: all 10 items remain satisfied — rev3 is test-only, changes
no behavior, and repairs the two remaining red lanes at their root
cause with the ordered skip-vocabulary and identity-assertion fixes.
## 11. Rework rev4 — review corrections 1–3 (this run)

Change Request revision 3 (hosted run `35286627425`, green on every lane
at `SPEC_PIN dced9b8`) was rejected with three corrections
(`TASK-260917-16l2md_review-verdict-rev3.md`). The union fidelity, the pin
move, E2, and the E4 identity work passed; this run keeps them intact and
implements exactly the three corrections below. No other behavior changes.

### 11.1 Correction 1 (High) — S4 surfacing rows print before publication

Landed §2.3 requires the manager to PRINT the `mcp-declaration` rows after
the audit gate passes and before the lock is published or any surface is
(re-)materialized. Rev3 computed the rows in `internal/envprofile` but the
CLI printed them only after `Install`/`Update` returned (lock already
published, home already activated), and a publication failure discarded
them through an empty `Info` error return.

The fix is an operation→CLI emission seam, not a redesign:

- `InstallOptions.SurfacingSink io.Writer`
  (`internal/envprofile/envprofile.go`), `ImportOptions.SurfacingSink`
  (`internal/envprofile/import.go`, threaded into the underlying
  install), and new `UpdateOptions{Policy, SurfacingSink}` with
  `UpdateWithOptions` (`UpdateWithPolicy` delegates with a nil sink).
  `updateLocked` takes the sink; the same-source git reinstall in
  `installLocked` passes it through.
- Five emission points, each immediately after the audit gate and before
  any publication/materialization: fresh install, reinstall unchanged
  (before activation), reinstall changed (before publish/resync), update
  unchanged, update changed (before publish/resync). Helper
  `carrySurfacing` (`internal/envprofile/surfacing.go`) emits the rows
  and returns what `Info.Surfacing` must carry: with a sink the rows
  were already printed so `Info` carries none (nothing prints twice, by
  construction); without a sink `Info` carries them as before.
- Emission is non-fatal: a nil sink prints nothing, a write failure is
  dropped (§2.3: surfacing emits no diagnostic and never fails the
  operation).
- The CLI (`cmd/curator/profile.go`: install, import, update) passes
  `SurfacingSink: c.stdout` and no longer prints `Info.Surfacing`
  (always empty with a sink). Stdout bytes are unchanged, only earlier.
  `env status` is untouched: it repeats the rows for the current
  profile with no publication involved.

Committed order-observation tests (all OBSERVE runtime order at the
filesystem; every one kills the defect mutant — see §11.4):

- `internal/envprofile/surfacing_order_test.go` (new): `lockObservingSink`
  snapshots lock.json existence/bytes when the first row arrives.
  `TestInstallEmitsSurfacingBeforePublication` (lock absent at first
  row, plus same-source git reinstall through the delegation),
  `TestUpdateEmitsSurfacingBeforePublication` (old-lock byte identity
  at first row, plus unchanged-update single emission),
  `TestReinstallEmitsSurfacingBeforePublication` (changed-path reinstall
  moves with old-lock identity at first row),
  `TestSurfacingEmittedDespitePublicationFailure` (profile directory
  obstructed by a regular file so the journal fails on every platform:
  the install fails but the exact row was already emitted, and nothing
  is published), `TestSurfacingSinkWriteFailureIsNonFatal` (a refusing
  sink never fails the install).
- `cmd/curator/profile_surfacing_test.go`:
  `TestProfileInstallSurfacesBeforePublication` (the reviewer's
  `TestReviewerSurfacingBeforePublication` probe shape through
  production `run()`: stat-at-first-row, exactly-once count, row before
  the installed report, allowlist warning) and
  `TestProfileUpdateSurfacesBeforePublication` (old-lock identity at
  first row, exactly-once count, move reported, warning).
- `runSurfacingOrderCase`
  (`internal/envprofile/envpassthrough_conformance_test.go`): the proxy
  `assertSurfacingDescribesCandidateLock` (which formatted a candidate
  lock through a helper and passed while production printed late) is
  deleted. Each order vector now pins the vector order, proves a blocked
  audit emits no rows to the sink and publishes nothing, and observes a
  live operation (`observeLiveInstallEmission` /
  `observeLiveUpdateEmission`, shared with the committed tests). No
  vector weakened: both order cases still execute with the same names.

### 11.2 Correction 2 (Medium) — `provider_directories: null` rejected

`internal/config/environments.go` accepted an explicit `null` as `[]`;
the rc.12 `manager-config-v2` schema requires an array (default `[]`, no
null meaning — unlike `passable_env_names`). The `rawPD != nil` bypass is
removed, so a present `null` reaches `parseProviderDirectories` →
`stringList` and fails with the existing wrong-type error naming
`environments.provider_directories`. The system file takes the same path
(`parseSystemEnvironments` delegates to `parseEnvironments`).
Committed regression `TestProviderDirectoriesNullRejected`
(`internal/config/environments_test.go`): user-config null refused,
omitted/empty accepted, system-config null refused — mirroring the E2
waiver fix, which stays closed.

### 11.3 Correction 3 (Low) — operator docs

New `docs/environment-config.md`: the three knobs/policies with exact
spec spellings (`transitive_system_modules` + `system_module_waivers`
with a waiver example, `provider_directories` with the revision-A
warning / revision-B refusal, `passable_env_names` with explicit `null`
= unbounded and the S4 profiles), migration steps for both warn-first
flips, the absent-vs-null table, the §12.2 lockability subset, and the
surfacing/warning emission points. The historical
`docs/security-audit-2026-09.md` is untouched history. No em-dashes
(docs/prose-style.md). CHANGELOG unchanged in rev4: the three entries
and the pin note already describe exactly this behavior; these are
defect corrections inside the unlanded candidate, not new scope.

### 11.4 Mutants killed (this run, all restored after)

- S4 emission-after-publish (envprofile): moved the fresh-install
  `carrySurfacing` below `op.publish` → new install order test FAILs
  (`lock.json already exists when the first mcp-declaration row is
  written`), restored.
- S4 order-vector proxy: same mutant →
  `TestEnvironmentsEnvPassthroughVectors/order/install-surfacing-before-publish`
  FAILs, restored. (The old proxy passed under this mutant; the new
  driver does not.)
- S4 rev3 CLI behavior restored (no sink + post-return print loops) →
  both new CLI before-publication tests FAIL (install: lock exists at
  first row; update: first-row bytes are not the old lock), restored.
- E2/E4/S4 narrowing mutants from the rev3 review were not re-run: no
  rev4 change touches the admitted-set, trust-verdict, or resolve-bound
  gates (verified by diff: rev4 production delta is the sink seam, the
  one-line null guard, and docs).

### 11.5 Validation transcripts (this run, rc.12 root)

Conformance root for every run: `/tmp/spec-rc12/conformance/v1` at
exactly `dced9b8317e0e8af79edf2d0539b32bd22b6c85b` (verified
`git rev-parse`). Shell `bash`, `set -o pipefail`; exit codes are the
gate commands' real statuses. The `go-testing-tools` skill was read and
set aside as TUI-specific; tests follow this repository's own
production-entry/vector-driver conventions.

Static (exit 0 throughout):

- `go build ./...` → exit 0
- `go vet ./...` → exit 0
- `gofmt -l internal cmd` → empty, exit 0
- `golangci-lint run ./internal/config/... ./internal/envprofile/...
  ./cmd/curator/...` → `0 issues.`, exit 0 (one revive
  unused-parameter on the new failingSink fixed first: real exit 1
  reported, then green)
- `bash .github/ci/gate-selftest.sh` → `gate-selftest: 180 passed,
  0 failed`, exit 0
- `bash .github/ci/ledger-consistency.sh /tmp/ledger-ev-rev4` →
  `239 rows checked across linux darwin windows`, `ok`, exit 0
  (ledger untouched by this delta)

Suites (all `-count=1`, exit 0):

- `go test ./internal/config/... ./internal/contextresolve/...
  ./internal/contextmaterialize/... ./internal/contextaudit/...` →
  ok 1.625s / 1.814s / 2.432s / 1.217s, exit 0
- `go test ./internal/shell/... ./internal/interop/environments/...
  ./internal/envfragment/... ./internal/globalbins/...` →
  ok 1.609s / 2.930s / 1.082s / 2.676s, exit 0
- `go test ./internal/envprofile/` → six `-run` chunks partitioning
  all 176 tests (30/30/30/30/30/26, anchored exact-match regexes,
  `-timeout 480s` each, strictly sequential), exit 0 each: chunk ok
  86.040s / 79.892s / 66.260s / 84.274s / 94.126s / 149.525s.
  Partition verified exact (176/176 unique names vs `go test -list`).
- `go test ./cmd/curator/` → four `-run` chunks partitioning the 217
  non-GOROOT tests (55/55/55/52, `-timeout 480s` each, strictly
  sequential) plus the 3 GOROOT-hashing `TestCompiledProject*` solos
  (`-timeout 500s`), exit 0 each: chunk ok 271.084s / 316.829s /
  456.666s / 121.112s; solos ok 288.369s
  (`TestCompiledProjectStatusAndUntrustedRecovery`) / 213.091s
  (`TestCompiledProjectRepairsCorruptCompiledState`) / 204.499s
  (`TestCompiledProjectRestoresCacheWhenCommitFails`).
  Partition verified exact (217/217 + 3/3 vs `go test -list`).
  (Single-package runs exceed the 600s default timeout on this host —
  the rev1–rev3-documented host characteristic, unrelated to this
  delta; chunks ran with no concurrent load.)

Vector-execution proof (each `-v` run: exit 0, every `--- PASS`, zero
`--- SKIP` except where noted):

- `TestManagerConfigV2Vectors`: 49 PASS (48 subtests + parent)
- `TestSystemModuleSchemaSubset`: 8 PASS (7/7 + parent)
- `TestSystemModuleAdmissionVectors`: 6 PASS (5/5 + parent)
- `TestUmbrellaProviderResolutionVectors`: 15 PASS (14 cases × 2
  revisions + parent)
- `TestEnvironmentsEnvPassthroughVectors`: 26 PASS (7 resolution + 6
  allowlist + 4 schema + 6 surfacing + 2 order + parent), order cases
  now observing live emission (see §11.1)
- `TestConformanceEnvironmentsMonolithic` (interop): 25 PASS (24 + parent)
- New: `TestProviderDirectoriesNullRejected` PASS;
  `TestInstallEmitsSurfacingBeforePublication`,
  `TestUpdateEmitsSurfacingBeforePublication`,
  `TestReinstallEmitsSurfacingBeforePublication`,
  `TestSurfacingEmittedDespitePublicationFailure`,
  `TestSurfacingSinkWriteFailureIsNonFatal` PASS;
  `TestProfileInstallSurfacesBeforePublication`,
  `TestProfileUpdateSurfacesBeforePublication` PASS.

E2 rev-2 corrections re-verified (still closed):

- No `prunePostRevisionKnobs`/`postPinKnobs` anywhere in `internal/`
  or `cmd/` (`grep` clean); the manager-config-v2 comparison stays
  exact and green at the new root.
- Explicit null `system_module_waivers` refused
  (`TestSystemModuleWaiversNullRejected` passes).
- The admission/schema subset drivers execute with no skip at the rc.12
  root; the ledger stays byte-identical to trunk.

### 11.6 Files changed in rev4 (10; union §§1–10 otherwise intact)

Production (5):

- `internal/envprofile/envprofile.go`: sink fields/options, five
  emission points, `UpdateOptions`/`UpdateWithOptions`.
- `internal/envprofile/surfacing.go`: `emitSurfacing`/`carrySurfacing`.
- `internal/envprofile/import.go`: sink threaded through the import.
- `cmd/curator/profile.go`: the three rows pass `c.stdout`; dead
  post-return print loops removed.
- `internal/config/environments.go`: null `provider_directories`
  rejected (one guard + comment).

Tests (4):

- `internal/envprofile/surfacing_order_test.go` (new): five
  order-observation tests + shared sink/helpers.
- `internal/envprofile/envpassthrough_conformance_test.go`: order
  driver observes live emission; proxy helper deleted.
- `cmd/curator/profile_surfacing_test.go`: two CLI
  before-publication tests.
- `internal/config/environments_test.go`: null-trust-roots regression.

Docs (1): `docs/environment-config.md` (new).

Checklist: all 10 items remain satisfied — rev4 repairs the three
rejected points at their root cause (emission seam, null guard, docs)
with mutant-verified tests, and changes no other behavior.

## 12. Rework rev5 — remove accidental root build executable (this run)

Per `TASK-260917-16l2md_rework-rev5.md` and the round-2 verdict
(`TASK-260917-16l2md_review-verdict-rev4.md`, sole remaining defect):
published rev4 accidentally captured local build output — the 18,798,096-byte
Mach-O executable `curator` (mode 100755) at the repository root. All three
round-1 corrections and the rest of the union were accepted unchanged.

What was done (exactly the rework brief, no other changes):

- `git status --short` before: untracked root `curator` present
  (18,798,096 bytes, matching the verdict's blob description) alongside
  the union delta.
- `rm ./curator`; `git status --short` after: 40 paths — 37 tracked
  (union source/test/docs/CI) plus exactly the 3 legitimate untracked
  union files (`cmd/curator/umbrella_trustroot_windows_test.go`,
  `docs/environment-config.md`,
  `internal/envprofile/surfacing_order_test.go`). No other build output
  or stray file present (`ls curator` → no such file; no `bin/`
  directory created).
- §11.6 correction: "Files changed in rev4 (10)" counts only the
  source/test/docs files (5 production + 4 tests + 1 docs); the published
  rev4 delta actually had 11 paths including the binary (41 changed paths
  total vs rev3's 37 + legit additions). Rev5 restores the candidate file
  set to those 10; no source, test, docs, or CI byte changed.

Build sanity (output to `/tmp` only, never the worktree root):
`go build -o /tmp/curator-rev5-check ./cmd/curator` → exit 0; the
`/tmp` artifact was removed afterwards. No test rerun: nothing
executable changed since the green rev4 validation, and the hosted gate
re-runs at handoff.

Checklist: all items remain satisfied — rev5 is a pure hygiene cleanup.
