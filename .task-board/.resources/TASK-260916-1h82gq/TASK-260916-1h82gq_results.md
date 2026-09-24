# TASK-260916-1h82gq results (R2 capabilities + mandatory portable controls) — rev 3 (final) + Revision 2 rework appendix

Continued from the timed-out run without touching its tree (no checkout/clean/stash).
Worktree state at handoff: 18 modified + 10 new paths, uncommitted, diff stat
`1189 insertions(+), 110 deletions(-)`.

## Gates observed green this run (zsh, darwin, GOFLAGS=-p=1 GOMAXPROCS=2)

| Gate | Result |
|---|---|
| `go build ./...` | exit 0 |
| `go vet ./internal/scriptworker/ ./internal/scriptpolicy/ ./internal/install/` | exit 0 |
| `go test -count=1 ./internal/scriptpolicy/` | ok 0.539s |
| `go test -count=1 ./internal/scriptworker/` | ok 22.852s |
| `go test -count=1 ./internal/runtimestore/` | ok 8.305s |
| install `TestEnforcedScriptCommandIsRefusedAtInstall` | PASS 1.9s |
| install `TestDeclaredOnlySchema8ScriptCommandInstallsUnchanged` + `TestActiveScriptCommandsRefusesEnforcedCommand` + `TestActiveEnforcedScriptCommandsAdmitsBehindSeam` | ok 4.654s |
| install `TestEnforcedScriptCommandInstallsNativeLauncherBehindSeam` | PASS 87.5s |
| install `TestEnforcedToDeclaredOnlyFlipRemovesNativeLauncher` | PASS 92.7s |
| `gofmt -l cmd internal` | clean (after `gofmt -w` on 4 inherited R2 files: alignment only) |
| `golangci-lint run` (v2.12.2, touched pkgs) | 0 issues (after 2 fixes: package-comment re-homed into `DeriveProfile` doc per repo convention; unused `fixture` param renamed in `farm-outside-base`) |
| `TestWorkerRevalidatesDerivedEnvironment` after lint fixes | ok 1.574s |

Box note: this host (32 GB, ~65 MB free at worst, user VMs + a concurrent
curator-spec suite running) SIGKILLs large test binaries intermittently
(`signal: killed` at ~80s). Every row above re-ran to a clean PASS; the two
~90s install rows are slow because each test process builds the full curator
binary via `enforcedBinaries` (one build per `go test` invocation, shared via
`enforcedOnce` within it).

## Design (as implemented in the worktree)

- Environment builder (`internal/scriptworker/capabilities.go`): empty bootstrap,
  then manager-set values (PATH=farm, HOME/TMPDIR/XDG_\*, CSK_PROJECT_ROOT,
  Windows TEMP/TMP/USERPROFILE/APPDATA/LOCALAPPDATA/PATHEXT + SYSTEMROOT/WINDIR/
  COMSPEC passthrough), then exactly the non-reserved `env_read` names present in
  the host environment. Reserved enumeration: portable exact set + proxy/resolver
  subset + `LD_*`, macOS `DYLD_*`, Windows set (case-insensitive), `python3-v1`
  `PYTHON*`/`__PYVENV_LAUNCHER__`, `node-v1` `NODE_*`/`NPM_CONFIG_*`. Unknown
  interpreter identifier (no reserved set) is refused.
- PATH builder (`internal/scriptworker/exec.go`): manager-owned farm directory holding
  exactly the resolved interpreter plus manager-resolved declared exec names (hard
  link, copy fallback); inherited PATH discarded. Exec resolution is manager-owned
  fixed directories only, with forbidden-root exclusion, regular-file/single-link/
  executable checks, and launch-boundary hash re-verification (`VerifyExec`).
- Offline config: derived network none (absent or empty) => offline mode + proxy/
  resolver scrub (shared `proxyOrResolverName` subset) plus closed-form validation.
  Declared hosts are recorded reporting-only, never a filter.
- Runtime area binding: manager-selected working directory (canonical project root for
  repo/path-set filesystem declarations, else the private tmp root) + private
  tmp/config/cache roots bound through the platform environment; closed-form
  validator (`validateDerivedEnvironment`) runs on the manager before launch AND in
  the worker before the interpreter starts.
- Launcher (`internal/scriptworker/launcher.go`, `cmd/curator/main.go`, install +
  runtimestore): enforced commands install as native launchers (manager copy + JSON
  sidecar contract), no shell/`.cmd`/symlink shim; enforced commands stay out of the
  user-bin forwarding mirror; worker mode dispatches before shim dispatch;
  install records the static derivation in result messages.
- Permit frame (`protocol.go`, `client.go`, `worker.go`): parent validates the worker
  proof (nonce-bound) and sends permit before the interpreter starts; foreign nonce,
  proof mismatch, or wrong started-count refuses.
- `activeScriptCommands` guard replaced (install/targets.go): enforced set computed
  explicitly instead of the `%w(<nil>)` shape.
- Preflight seam: `scriptpolicy.OverrideMandatoryControlsForTest` forces the table
  complete in tests through the same production predicate; production has no setter
  (no env var/flag/config/build tag); fresh-process refusal proven by
  `TestProductionBinaryRefusesWithoutSeam` (built binary: `skill check` + executed
  launcher).

## Row table (production boundary per row)

Vector `script-host-execution-policy.json` at CI pin 87a0d006.

Derivation cases (4/4 driven at the production entry):
| Case | Row | Boundary |
|---|---|---|
| all-fields-absent-deny-by-default | TestCapabilityDerivationAllFieldsAbsentDenyByDefault | runSession -> real worker process + stub interpreter |
| declared-network-hosts-are-reporting-only | TestCapabilityDerivationNetworkHostsReportingOnly | same |
| declared-exec-is-manager-resolved | TestCapabilityDerivationExecIsManagerResolved | same |
| declared-secrets-remain-identifiers | TestCapabilityDerivationSecretsRemainIdentifiers | same |

Controls (R2 set; R1 rows intact):
| Control | Row(s) | Boundary |
|---|---|---|
| manager-built-environment | TestManagerBuiltEnvironmentWithholdsReserved, TestReservedEnvironmentNameTable, TestInterpreterWithoutReservedSetIsRefused | derived session + unit table |
| manager-built-path | TestCapabilityDerivationExecIsManagerResolved | derived session |
| offline-network-configuration | TestOfflineScrubWhenNetworkNone | derived session |
| operation-private-runtime-area | TestOperationPrivateRuntimeAreaApplied, TestWorkerRevalidatesDerivedEnvironment | derived session |
| permit/explicit streams | TestWorkerWaitsForPermit, TestWorkerRejectsPermitWithForeignNonce, TestParentWithholdsPermitOnProofMismatch, TestParentRejectsWrongStartedCount | session protocol |
| launcher | TestLaunchDrivesDerivationBehindSeam, TestNativeLauncherWorkerModeWinsOverShimDispatch, TestProductionBinaryRefusesWithoutSeam, install native-launcher rows, runtimestore enforced rows | built curator binary + install entry |
| admission honesty | TestPreflightStillRefusesForR3ControlsAlone, TestProductionTableCannotBeForced, TestARefusalPrecedesEveryWorkerSurface, TestPreflightRefusalCases | Admit/install/invoke |

## Mutant table (source mutants executed this run; each reverted, tree verified clean)

| ID | Control | Narrowing mutant (production call site) | Killer row | Verdict |
|---|---|---|---|---|
| M1 | manager-built-environment | `LD_*` prefix dropped from `reservedEnvironmentName` (capabilities.go) | TestManagerBuiltEnvironmentWithholdsReserved | KILLED (exit 1, reserved names reach interpreter) |
| M2 | manager-built-path | derived PATH gains `:/inherited` (`buildSessionEnvironment`) | TestCapabilityDerivationExecIsManagerResolved | KILLED (manager self-check refuses) |
| M3 | offline-network-configuration | proxy value planted in map + scrub loop skipped | TestOfflineScrubWhenNetworkNone | KILLED via closed-form validator |
| M3b | offline-network-configuration | pure scrub skip, no plant | (none) | SURVIVES — bound: reserved filter + validator fully cover the property; scrub is defense-in-depth removal (removal-vs-refusal unobservable at this boundary) |
| M4 | operation-private-runtime-area | TMPDIR bound to `/host-tmp-mutant` | TestOperationPrivateRuntimeAreaApplied | KILLED (self-check refuses) |
| M5 | permit frame | parent skips manager-proof validation before permit (client.go) | TestParentWithholdsPermitOnProofMismatch/forged-manager-proof | KILLED (wrong-error red) |
| M6 | manager-built-environment | `DeriveProfile` reserved-set gate removed (unknown interpreter proceeds) | TestInterpreterWithoutReservedSetIsRefused | KILLED (derivation-layer half red; runSession half still refuses via ResolveInterpreter) |
| M7 | manager-built-path | caller PATH prepended to exec search dirs (`DeriveProfile`) | TestCapabilityDerivationExecIsManagerResolved | KILLED (git resolves from decoy dir) |
| M8 | all-fields-absent-deny-by-default | absent network derives online | TestCapabilityDerivationAllFieldsAbsentDenyByDefault | KILLED (mode recorded-hosts) |
| M9 | declared-secrets-remain-identifiers | secret identifiers resolve from host env (injection, buildSessionEnvironment) | TestCapabilityDerivationSecretsRemainIdentifiers | KILLED (closed-form validator refuses the injected entry; 3 earlier attempts SIGKILLed by box pressure, 4th completed red) |

Post-mutant verification: `grep MUTANT internal/` empty; `git diff --stat`
identical to session start (18 files, 1189+/110-, 28 paths); `go build ./...`
exit 0; `go test -count=1 ./internal/scriptworker/` ok 17.664s after all
reverts.

## Ratio line

Vector keys: derivation 4/4 driven at production entry; mandatory controls R2-set
implemented (9/11, R3 owns the 2 evidence rows) with rows at the entry;
opt_in_cases 6/6 classification kept; evidence (14), preflight (5), audit (4)
untouched (R3/R4). Ledger: 24 R2 rows registered in platform-cases.tsv
(scriptpolicy 2, scriptworker 16, install 2, runtimestore 4; verified by count).

## Windows proof status

UNVERIFIED locally (darwin host). Windows rows (case-insensitive reserved names,
PATHEXT/TEMP/TMP bindings, farm entry extension rule, no-`.cmd` launcher) are
committed and registered for windows-latest; hosted evidence comes from CI lanes.

## Bounds

- Stream model: piped stdin is pre-read into the session payload (bounded); terminal
  stdin binds the null device. Live pass-through streaming stays R3-or-later; the
  launcher works without it (stated in code + docs).
- Declared network hosts are reporting-only; no host filtering is promised.
- Production still refuses every enforced install/invocation with
  `script_execution_control_unavailable` (two R3 evidence controls unavailable).
- M3b survivor documented above (redundant layer, intentional).

## Undone

None. All brief scope is implemented, gated, and mutant-attacked; the only
non-local proof is Windows hosted lanes (CI), recorded above as unverified.
Campaign rules prohibit LOGBOOK.md edits, so findings live in this outcome
instead of the logbook.

---

## Revision 2 (rework-1: Windows gate failures) — FINAL

Gate run 35627712332 failed ONLY `Test (windows-latest)` (Linux/macOS green).
Evidence stream `test-evidence-windows-latest/go-test.json` (downloaded from the
run) confirms all four rework causes with exact messages:

- Cause 1 (scriptworker, 39 fails incl. subtests): every session refuses with
  `script_execution_worker_identity_invalid: the resolved interpreter python3-v1
  has multiple filesystem links or is a reparse point` (e.g. client_test.go:108).
  Root cause: `buildPathFarm` uses `os.Link` (hard link) for farm entries on
  Windows (`internal/scriptworker/exec_windows.go`), which raises the verified
  interpreter's NTFS link count to 2; the unchanged launch-boundary check
  (`readInterpreterIdentity` → `godriver.HasMultipleLinks` →
  `GetFileInformationByHandle.NumberOfLinks != 1`) then refuses. Unix uses
  symlinks (no link-count change), so only Windows fails. Same root downstream:
  `newWorkerFixture` links the stub then the worker re-verifies
  (`TestWorkerWaitsForPermit`: `worker sent "failure", want ready`), and the
  forge worker never launches so `TestParentWithholdsPermitOnProofMismatch`
  waits 10 s for a record that can never arrive.
  Fix: Windows farm entries are byte copies, never hard links (the copy
  fallback already exists and `farmEntryResolves` verifies copies by hash);
  the identity check is NOT relaxed. Plus a regression row pinning that farm
  construction preserves single-link identity, and a fail-fast assertion in
  the forge row (see below).
- Cause 2 (runtimestore, 1 fail): `enforced_test.go:108: staged launcher mode
  = -rw-rw-rw-, want owner-only executable`. Windows has no POSIX exec bits;
  `os.WriteFile(0o700)` materializes 0666. Fix: platform-aware assertion —
  POSIX keeps `Perm() == 0o700`; Windows asserts a regular `.exe` file
  (native launcher, never `.cmd`), matching the go-v1 staging convention
  (`runtime.GOOS != "windows"` guard as in `godriver/build_test.go:64`).
  Production staging unchanged (it already mirrors go-v1: copy + `0o700`,
  executability via `.exe` on Windows).
- Cause 3 (install, 1 fail): flip reinstall
  `Errors:[duplicate staged target 80-removal/shim/flip-skill-tool]`.
  Root cause: `stageRuntimeAndShims` records enforced removals with identifier
  `"shim/"+command` for BOTH the launcher and the sidecar
  (`internal/install/targets.go`); `staging.Plan.Validate` rejects the
  duplicate `Class + Identifier`. On unix the launcher removal is suppressed
  by the `replaced[]` map (ordinary shim takes the same path), so only
  Windows (`.exe` launcher vs `.cmd` shim, distinct paths, both removals
  land) trips it. Fix: distinct removal identifiers mirroring the desired
  side (`enforced-shim/<cmd>` / `enforced-sidecar/<cmd>`). The flip test
  itself is the Windows pin (it runs on windows-latest via the ledger).
- Cause 4 (snapshot, 1 fail): `snapshot_test.go:61: worker 14 ... The process
  cannot access the file because it is being used by another process.`
  NOT R2's: the R2 tree contains zero paths under `internal/snapshot` or
  `internal/gitops`, and `go list -deps ./internal/snapshot` contains none
  of the R2-touched packages (`scriptworker`, `scriptpolicy`, `runtimestore`,
  `install`), so R2 cannot have changed the publisher path. Pre-existing
  Windows file-locking flake under 16 racing publishers; no code change —
  orchestrator files separately.

Fixes implemented (same 28 paths as rev1, no new files except ledger/test
rows; `git diff --stat` tracked: 1202 insertions, 110 deletions):

- Cause 1: `internal/scriptworker/exec_windows.go` — `linkFarmEntry` now
  always byte-copies, never `os.Link`; the single-link identity check is
  untouched. `exec.go` farm comments updated (symlinks on unix, copies on
  Windows, never hard links). New ledger row
  `TestPathFarmPreservesSingleLinkIdentity` (linux,darwin,windows) builds a
  farm through the production constructor and pins `HasMultipleLinks ==
  false` + `VerifyInterpreter` green afterwards.
- Cause 2: `internal/runtimestore/enforced_test.go` — destinations derive
  for the host platform; assertion is `IsRegular` everywhere, `Perm() ==
  0o700` on POSIX, native `.exe` name on Windows. Production staging
  unchanged.
- Cause 3: `internal/install/targets.go` — new `enforcedPlanIdentifier`
  gives launcher and sidecar distinct plan identities
  (`enforced-shim/<cmd>` / `enforced-sidecar/<cmd>`) on BOTH the desired
  and removal sides (desired behavior byte-identical). Windows naming
  audit: ordinary enumeration takes `.cmd` only on Windows (no `.exe`
  double-claim), launcher dispatch derives sidecar as exe+suffix on both
  platforms, forwarding mirror still excludes enforced — no other site
  stages the launcher. The ledger flip test is the Windows pin.
- Forge fail-fast: `TestParentWithholdsPermitOnProofMismatch` now asserts
  the error names `identity proof` (both forged-proof diagnostics contain
  it; no pre-launch refusal does), so a pre-launch refusal fails
  immediately with its detail instead of a 10 s wait.
- Cause 4: no code change (evidence above).

Gates observed green this run (zsh, darwin, default GOFLAGS):

| Gate | Result |
|---|---|
| `go build ./...` | exit 0 |
| `GOOS=windows go build ./...` | exit 0 |
| `go vet` (scriptworker, scriptpolicy, install, runtimestore) | exit 0 |
| `GOOS=windows go vet` (scriptworker, runtimestore, install) | exit 0 |
| `gofmt -l` (scriptworker, runtimestore, install, cmd) | clean |
| `golangci-lint run` v2.12.2 (same 3 pkgs) | 0 issues |
| `go test -count=1 ./internal/scriptpolicy/ ./internal/runtimestore/` | ok 0.564s / 5.440s |
| `go test -count=1 ./internal/scriptworker/` (incl. new row + all rework-named downstream rows) | ok 11.421s |
| install launcher + flip rows | PASS 55.8s + 55.7s |
| install admission rows (refusal/declared-only/active x2) | ok 3.505s |

Mutant probe (mechanism check, reverted): unix farm entry temporarily
switched to `os.Link` (simulating the old Windows behavior) →
`TestPathFarmPreservesSingleLinkIdentity` FAILS with `multiple=true`;
restored byte-identical, row green again. `grep MUTANT internal/` empty.

Ratio line delta: ledger R2 rows 24 → 25 (new farm row). All other rev3
lines unchanged.

Windows proof status: still UNVERIFIED locally (darwin host) — the three
fixes change Windows-only behavior (farm copy), Windows-only assertions,
and a Windows-only plan path (unix flip suppresses the launcher removal
via `replaced[]`, so the darwin flip row guards regressions but cannot
prove the fix). Hosted windows-latest evidence comes from the gate rerun.

Bounds: per-copy farm cost on Windows (one interpreter + exec copies per
invocation; hash-verified by `farmEntryResolves`) replaces the free hard
link — correctness over speed, documented in code. Brief's honest-partial
list unchanged (none). Campaign rules prohibit LOGBOOK.md edits, so
findings live in this outcome instead of the logbook.

---

## Revision 3 (rework-2: Windows SYSTEMROOT injection) — FINAL

Gate run 35633381854 failed ONLY `Test (windows-latest)`, and only in
`TestCapabilityDerivationAllFieldsAbsentDenyByDefault` (both subtests): the
interpreter's environment keys contain `SYSTEMROOT` while the row expects the
manager-built set without it. Linux/macOS lanes green.

- Cause: `runInterpreter` (`internal/scriptworker/worker.go`) sets `command.Env`
  to exactly the derived environment. The absent-row's synthetic
  `HostEnvironment` carries no `SYSTEMROOT`, so the conditional host-carry path
  in `buildSessionEnvironment` omits it — and Go's `os/exec` on Windows then
  injects `SYSTEMROOT=<worker-process value>` (documented: "On Windows, ...
  adds SYSTEMROOT if missing"). Production behind a real manager environment
  always carried `SYSTEMROOT` through the host-carry path, so only
  synthetic-host-env rows observed the injection.
- Fix (2 files, no other change):
  - `internal/scriptworker/capabilities.go` — `buildSessionEnvironment`,
    Windows only: `SYSTEMROOT` and `WINDIR` are now always manager-set. The
    request host value wins; when the host does not carry one, the manager's
    own process value stands (`os.LookupEnv` — the same source the worker
    bootstrap reads in `indispensableScriptEnvironment`). `COMSPEC` stays
    host-carried only: nothing injects it. Reserved-name handling, manager-set
    on Windows: `SYSTEMROOT` (always), `WINDIR` (always), `COMSPEC` (when the
    host carries it) — all reserved, never from `env_read`.
  - `internal/scriptworker/derive_test.go` — absent-row Windows `wantKeys` +=
    `SYSTEMROOT`, `WINDIR` (per-OS expected set).
- Why both guaranteed: the spec reserves both as manager-owned and the tree
  already set `WINDIR` deliberately (host-carried); guaranteeing both closes
  the derived set on Windows. `COMSPEC` deliberately not guaranteed (no
  injector exists for it).
- Blast radius (verified by inspection): no other row asserts an exact
  environment key set (only the absent-row compares `gotKeys`/`wantKeys`);
  `validateDerivedEnvironment` forbids no extras;
  `TestManagerBuiltEnvironmentWithholdsReserved` on Windows supplies both names
  via the host and still sees the host values stand (host wins over fallback).
  Worker fixtures build through the same builder with an empty host env — on
  Windows they now carry real `SYSTEMROOT`/`WINDIR`, which validation accepts.

Gates observed green this run (zsh, darwin, go1.26.0):

| Gate | Result |
|---|---|
| `go build ./...` | exit 0 |
| `GOOS=windows go build ./...` | exit 0 |
| `go vet` (scriptworker, scriptpolicy, install) | exit 0 |
| `GOOS=windows go vet` (same pkgs) | exit 0 |
| `gofmt -l cmd internal` | clean |
| `golangci-lint run ./internal/scriptworker/` | 0 issues |
| `go test -count=1 ./internal/scriptworker/ ./internal/scriptpolicy/` | ok 13.673s / 0.361s |
| Throwaway Windows-builder probe (platform="windows" driven directly on darwin: fallback, host-precedence, COMSPEC-absent) | PASS, file deleted after |

install + runtimestore rows: NOT rerun this revision — accepted from the
Revision 2 evidence above. The Revision 3 diff cannot change darwin behavior
(Windows-only block; the production call site passes `platform =
runtime.GOOS`) and touches neither package.

Windows proof status: still UNVERIFIED locally (darwin host) — the fallback's
real-value leg (`os.LookupEnv` on windows-latest) and the per-OS row
expectation are proven by the gate rerun. The throwaway probe proves the
builder logic; the hosted lane proves the values exist there.

Bounds: brief's honest-partial list unchanged (none). Campaign rules prohibit
LOGBOOK.md edits, so findings live in this outcome instead of the logbook.
