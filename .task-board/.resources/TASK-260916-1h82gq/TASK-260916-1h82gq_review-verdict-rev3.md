# TASK-260916-1h82gq — review verdict, revision 3 (R2: capabilities + mandatory portable controls)

Reviewer run RUN-260921-271026 (claude-opus-5), 2026-09-21. Read-only review of
`CR-TASK-260916-1h82gq-3`; nothing in the Story worktree was modified (every rerun, probe and
mutant ran in disposable clones `/tmp/r2c` (reruns), `/tmp/r2m` (mutants), `/tmp/r2p` (untracked
zz probes), all committed at the candidate tree).

## 1. Candidate identity

| Fact | Value |
|---|---|
| Base OID | `d28b182ac3811bde9d8a0a64f32ca23b7a0ac768` (R1 checkpoint) |
| Candidate tree OID | `a04c8db64340bcb40488600353c1e5ceb0929097` |
| Story worktree (temp-index `read-tree HEAD` + `add -A internal/ cmd/ docs/ CHANGELOG.md .github/` + `write-tree`) | `a04c8db6…` — **equal** (untracked files included) |
| `TASK-260916-1h82gq_change-request_rev3.patch` sha256 | `67fd48926aaab2be079644b5a45989bae944600398960f5b71bff83823b516be` — **equal** to the CR record |
| Clones `/tmp/r2c`, `/tmp/r2m`, `/tmp/r2p` (`checkout --detach <base> && apply --index <rev3 patch> && commit`) → `HEAD^{tree}` | `a04c8db6…` all three |
| Gate run 35638584352 (`remote gate: run … finished: success`; Naming, Lint, Interop, Test ×3, Race ×2, Gate self-test ×3 = 11 jobs success; rose-air and candidate suite skipped as configured) | headSha `5665d4dc17a7243aae1c9fad49286e94bcd3537f`, parent `d28b182a`, `^{tree}` = `a04c8db6…` — **the gate ran the exact revision-3 tree** |

Diff shape: 28 paths, 4432 insertions / 110 deletions (`git diff --stat` on the clone). Spec pin used
for every normative check: curator-spec `87a0d0060bad64ab883d007dcdf35df7485368bf` (`profiles/manager.md`
§3/§3.1, `protocol/core.md` §4.1.1/§4.3, vector `script-host-execution-policy.json`; conformance
manifest sha256 `0e195ecd…` as in the reconciliation record).

## 2. What I read (design check against the spec at the pin)

Files read in full: `internal/scriptworker/{capabilities,exec,exec_unix,exec_windows,launcher}.go`,
`runtimestore/enforced.go`; diffs of `client.go`, `worker.go`, `protocol.go`, `scriptworker.go`,
`scriptpolicy.go`, `install/{targets,install,global}.go`, `cmd/curator/main.go`, every test file, the
ledger, CHANGELOG, `docs/script-interpreters.md`, `docs/troubleshooting.md`.

| Brief item | Where | Verdict |
|---|---|---|
| Reserved enumeration (§3.1) | `capabilities.go:527-607` | Matches the profile name for name: portable exact set (PATH, HOME, TMPDIR, TEMP, TMP, XDG_{CONFIG,CACHE,DATA,STATE}_HOME, IFS, CSK_PROJECT_ROOT), the five proxy names + lowercase spellings, RES_OPTIONS/HOSTALIASES/LOCALDOMAIN, `LD_*`; darwin `DYLD_*`; Windows USERPROFILE/APPDATA/LOCALAPPDATA/PATHEXT/COMSPEC/WINDIR/SYSTEMROOT matched case-insensitively (`EqualFold`, host keys folded in `parseHostEnvironment`, validator folds); `python3-v1` `PYTHON*`+`__PYVENV_LAUNCHER__`; `node-v1` `NODE_*`/`NPM_CONFIG_*`. Unknown identifier → `DeriveProfile` refuses `package_influence_forbidden` (`:149`) and `DeriveStaticReport` likewise. |
| Manager-built environment | `buildSessionEnvironment` `:333-422` | Empty map → manager-set values (PATH=farm, HOME=private base, TMPDIR, XDG_*, Windows TEMP/TMP/USERPROFILE/APPDATA/LOCALAPPDATA/PATHEXT, SYSTEMROOT/WINDIR always manager-set on Windows, COMSPEC host-carried) → exactly the non-reserved `env_read` names present in the host env → offline scrub. Every withheld name is reported (`WithheldEnv`), as §3.1 requires. |
| Manager-built PATH | `exec.go` `ResolveExec`/`buildPathFarm`, `capabilities.go:165` | PATH = exactly one manager-owned farm directory under the operation-private base holding the interpreter + resolved `exec` names (unix symlink, Windows byte copy — never a hard link, which is the rev-2 NTFS fix; `readInterpreterIdentity`/`godriver.HasMultipleLinks` untouched). `exec` resolves only from compiled-in `DefaultExecSearchDirs` (or the test override), skipping forbidden roots, non-regular, multi-link, non-executable files; unresolvable names reported, never taken from the caller PATH. Farm contents re-verified entry by entry; `VerifyExec` re-hashes at the launch boundary (`client.go:105`). |
| Offline configuration + scrub | `:404-414`, `proxyOrResolverName` shared by filter, scrub and validator | Derived `network` none/absent → `NetworkOffline`, scrub of the shared proxy/resolver subset, validator refuses any survivor. Declared hosts → `RecordedHosts` only; no filter state exists anywhere (grep: no host list reaches the worker request). Manager-side and worker-side `validateDerivedEnvironment` (`worker.go:267`) are the same function. |
| Private runtime area applied (R-F) | `deriveWorkingDir` `:289-308`, bindings `:347-360`, validator bindings `:676-705`, worker `runInterpreter` (`Dir`, `Env` = derived) | Working dir = canonical project root for `repo`/path-set with a project, private tmp otherwise (absent, `home-config`, empty set, global scope). Presence is read from the declared bytes (`Present["filesystem"]`), so the parser's `"repo"` default cannot widen (core §4.1.1). |
| Secrets | `:201`, no code path reads a secret | Identifiers only; nothing resolves a value. |
| Launcher (R-A) | `runtimestore/enforced.go`, `install/targets.go:297-405`, `launcher.go`, `main.go:96-107,169-188` | Enforced commands stage as a native copy of the running manager (`<bin>/<cmd>` unix, `<cmd>.exe` Windows) plus `<launcher>.curator-shim.json`; excluded from the ordinary shim specs and from the global forwarding mirror; ordinary/enforced transitions computed independently with distinct plan identifiers (`enforced-shim/`, `enforced-sidecar/`, the rev-2 duplicate-target fix) and replacement-supersedes-removal. `main()` dispatches worker mode first, then sidecar-paired launcher mode (`RunShim` → `Launch` → `runSession` re-executing the launcher copy itself as the worker). No shell, `.cmd` or symlink between manager and interpreter; the interpreter is started by `exec.Command(<canonical path>)`. |
| Permit frame (R-C) | `protocol.go` `kindPermit`, `client.go:171-190`, `worker.go:59-97` | Parent validates the manager proof, interpreter proof, then sends `permit`; the worker blocks in `awaitPermit` before `serveRun`; shutdown instead of permit = clean exit 0; foreign nonce refused. |
| R-G guard | `targets.go:407-421,449-461` | `activeScriptCommands` skips enforced commands; `activeEnforcedScriptCommands` returns `Admit`'s error directly and never formats a nil (`%!w(<nil>)` shape gone). |
| R-H | `scriptpolicy.go:93-101` | `MandatoryControls()` now returns a copy; the slice is unexported. |
| R-I docs | `docs/script-interpreters.md`, troubleshooting, CHANGELOG `### Added` | Present; the diagnostic claims match the code (`ResolveInterpreter` unbound → `control_unavailable`). |
| Admission honesty | `scriptpolicy.go:70-81` table, `Launch:155`, `skillcheck`→install, `activeEnforcedScriptCommands` | `inventory-controls-applied` and `closed-script-capability-evidence-record` remain `Owner: R3`; every production entry still refuses `script_execution_control_unavailable`. |
| Seam | `OverrideMandatoryControlsForTest` | Only writer of `testControlOverride`; callers: `derive_test.go`, `install/scriptpolicy_test.go`, `scriptpolicy/conformance_test.go` only (`grep` over non-test files: none). No env var, flag, config key or build tag reaches it. |

## 3. My own reruns (disposable clone `/tmp/r2c`, darwin/amd64, go1.26.0, bash drivers with real exit codes)

Every command below ran through a health-gated driver (`perl -e 'alarm 8; exec' <println probe>` before each
attempt, retry on `signal: killed`) because the host opened an exec-stall window 19:07–19:22Z (fresh
binaries stranded at `_dyld_start`, `syspolicyd` absent from `ps`; rules forbid daemon restarts). Everything
below was produced **outside** that window; nothing was accepted from inside it.

| Command (clone at `a04c8db6…`) | Result |
|---|---|
| `go build ./...` / `go vet` (scriptworker, scriptpolicy, install, runtimestore, cmd/curator) | exit 0 / exit 0 |
| `GOOS=windows go build ./...` / `GOOS=windows go vet` (scriptworker, runtimestore, install) | exit 0 / exit 0 |
| `gofmt -l cmd internal` | clean |
| `go test -count=1 -v ./internal/scriptworker/` | `ok 13.445s`, 48/48 top-level PASS, 0 skip (R1 + R2 rows) |
| `go test -count=1 -v ./internal/runtimestore/` | `ok 4.271s`, 26 PASS, 2 pre-existing SKIPs (candidate-launcher contract rows, unrelated) |
| `go test -count=1 -v ./internal/scriptpolicy/` without a conformance root | `ok`, 6 vector rows SKIP (no root) |
| `CURATOR_CONFORMANCE_ROOT=/tmp/r2spec/conformance/v1 go test -count=1 -v ./internal/scriptpolicy/` (spec pin `87a0d006` checked out fresh, manifest sha `0e195ecd…`) | `ok 0.423s`, **15/15 PASS, 0 skip** — incl. `TestScriptHostExecutionPolicySectionsAreAllClassified`, `TestMandatoryControlsMatchTheSuite`, `TestPreflightStillRefusesForR3ControlsAlone`, `TestProductionTableCannotBeForced`, `TestARefusalPrecedesEveryWorkerSurface` |
| same root, `./internal/skillspec/ ./internal/skillcheck/` | `ok 3.371s` / `ok 2.220s` |
| `go test -count=1 -v ./internal/install/ -run 'Test(EnforcedScriptCommandIsRefusedAtInstall|DeclaredOnlySchema8ScriptCommandInstallsUnchanged|ActiveScriptCommandsRefusesEnforcedCommand|ActiveEnforcedScriptCommandsAdmitsBehindSeam|EnforcedScriptCommandInstallsNativeLauncherBehindSeam|EnforcedToDeclaredOnlyFlipRemovesNativeLauncher)$'` | `ok 126.985s`, 6/6 PASS (native-launcher row 63.6 s, flip row 58.8 s — each builds the production binary) |
| `go test -count=1 ./internal/config/ -run ScriptInterpreter` | `ok`, `TestParseScriptInterpreters` PASS |
| `go test -count=1 ./cmd/curator/ -run 'TestHiddenWorkerModeIsNotAUserVisibleCommand$'` | `ok` |
| Built production binary `go build -o /tmp/r2rev/curator ./cmd/curator/` | exit 0 (19.4 MB) |

Not rerun by me: the full `cmd/curator`, `internal/install` and race suites (accepted from the hosted gate,
§4), `golangci-lint` (hosted `Lint` job success accepted; a local run inside a stall window is unreliable).

## 4. Hosted evidence (gate run 35638584352, artifacts `test-evidence-{ubuntu,macos,windows}-latest` downloaded from the Story worktree)

`go-test.json` per lane, filtered to the four touched packages and to the 25 R2 ledger rows the candidate adds
to `.github/ci/platform-cases.tsv` (`internal/scriptpolicy` 2, `internal/scriptworker` 17, `internal/install` 2,
`internal/runtimestore` 4):

| Lane | scriptworker | scriptpolicy | runtimestore | install | R2 ledger rows | R2 subtests |
|---|---|---|---|---|---|---|
| ubuntu-latest | pass 4.2 s, 48 top-level pass | pass, 15 | pass, 28 | pass 85.6 s, 299 pass / 2 pre-existing skip | **25/25 pass, 0 skip** | 20 pass / 0 fail / 0 skip |
| macos-latest | pass 6.1 s, 48 | pass, 15 | pass, 28 | pass 624 s, 299 / 2 skip | **25/25 pass, 0 skip** | 20 / 0 / 0 |
| windows-latest | pass 12.7 s, 48 | pass, 15 | pass, 23 + 5 pre-existing unix-only skips | pass 1400 s, 291 / 10 pre-existing skips (draft-network rows) | **25/25 pass, 0 skip** | 20 / 0 / 0 |

Windows specifics the note asked for, all `pass` on windows-latest with elapsed > 0: `TestManagerBuiltEnvironmentWithholdsReserved`
(0.14 s; its `folded` list `path`, `home`, `WINDIR`, `Pathext`, `lD_pReLoAd`, `uSeRpRoFiLe` is asserted withheld
there, so case-insensitive matching is proven on the platform), `TestPathFarmPreservesSingleLinkIdentity` (the
NTFS fix row), `TestEnforcedToDeclaredOnlyFlipRemovesNativeLauncher` (37.4 s, the rev-1 duplicate-target
bug), `TestStageEnforcedShimTransitionStagesManagerCopies`, `TestCapabilityDerivationAllFieldsAbsentDenyByDefault`
(the rev-2 SYSTEMROOT fix, both subtests). No `fail` action exists in any of the three lanes for these packages.

NTFS fix is a cause fix: `exec_windows.go` `linkFarmEntry` byte-copies (never `os.Link`); `interpreter.go`,
`godriver.HasMultipleLinks` and `readInterpreterIdentity` are not among the 28 changed paths, so the
single-link identity check is byte-identical to R1 rev5. `internal/snapshot`/`internal/gitops` are not
in the delta (`git diff --stat` = 0 lines) and `go list -deps ./internal/snapshot` names none of
scriptworker/scriptpolicy/runtimestore/install, so the run-35633381854 `TestConcurrentGetAcceptsOneImmutablePublisher`
Windows failure cannot be R2's; it did not recur on the rev-3 gate. Orchestrator files it separately, as
results.md says.

## 5. Mutants (clone `/tmp/r2m`; python single-line replacer with sha check, fresh `go build ./...` + fresh test binary per mutant, `git checkout -- .` after each; clone ended at `a04c8db6…`, status clean)

| ID | Control / gate | Narrowing mutant (production site) | Rows run | Result |
|---|---|---|---|---|
| M1 | manager-built-environment | `IFS` dropped from `reservedExactNames` (`capabilities.go:533`) | WithholdsReserved, ReservedEnvironmentNameTable | **KILLED** (`reserved IFS reached the interpreter`) |
| M2 | manager-built-environment (macOS) | darwin `DYLD_` prefix narrowed to the single exact name `DYLD_INSERT_LIBRARIES` (`:580`) | same | **KILLED** (`DYLD_FUTURE` passed through) |
| M3a | manager-built-path | builder keeps the inherited PATH after the farm (`:347`) | 4 derivation rows + WithholdsReserved | **KILLED** by the manager-side closed-form validator (5 rows red) |
| M3b | manager-built-path, both layers | M3a **plus** validator separator and farm-equality checks neutralised (`:669`, `:672`) — the shared function, so the worker layer is neutralised too | + WorkerRevalidatesDerivedEnvironment | **KILLED** at the interpreter boundary: AllFieldsAbsent (`PATH reaches the inherited entry`), ExecIsManagerResolved (`decoyonly` resolves), WorkerRevalidates (`path-doubled`/`path-elsewhere`) |
| M4 | offline scrub | pure scrub skip (`:404` → `if false`) | OfflineScrub + derivation rows | **SURVIVES** — by construction: the reserved filter already withholds every proxy/resolver name from `env_read`, and no manager-set value is a proxy name, so the scrub can only ever remove what is already absent. Same finding as the producer's M3b; the property itself is pinned by M5 and by the validator. Stated bound, not a defect. |
| M5 | offline scrub / reserved set | `LOCALDOMAIN` dropped from the shared proxy-or-resolver subset (`:549`) | OfflineScrub, WithholdsReserved | **KILLED** (both rows) |
| M6b | operation-private-runtime-area | `TMPDIR` bound to the host `os.TempDir()` (`:349`) **and** the validator's binding-value check neutralised (`:701`) | RuntimeAreaApplied, WithholdsReserved, AllFieldsAbsent | **KILLED** at the interpreter (`cwd != TMPDIR`, `HOME != Dir(TMPDIR)`) — the rows see the applied values, not just the validator |
| M7 | all-fields-absent deny-by-default | absent `filesystem` treated as present, so the parser's `"repo"` default derives the project root (`:291`) | RuntimeAreaApplied, AllFieldsAbsent | **KILLED** (`absent-derives-no-root`: cwd = project instead of private tmp) |
| M8 | declared-secrets-remain-identifiers | secret identifiers resolved from the host under the normalised name `RELEASE_TOKEN` (`:387`) | SecretsRemainIdentifiers | **SURVIVES** (`ok 1.9 s`) — the committed row plants only the literal `release-token=` and an `env_read`-legitimised `RELEASE_TOKEN`; my probe row (§6, host `RELEASE_TOKEN=SUPERSECRET-…`, no `env_read`) **kills it** (`secret material reached the interpreter via RELEASE_TOKEN`) and passes on the candidate. Coverage gap, not a product defect. |
| M9 | manager-built-path / exec | exec search consults the request host PATH before the manager list (`:161`) | ExecIsManagerResolved | **KILLED** (`resolved git = <decoy>, want the manager search directory`) |
| M10 | permit frame (worker) | worker runs the interpreter without waiting for the permit (`worker.go:59`) | WorkerWaitsForPermit, RejectsPermitWithForeignNonce | **KILLED** (`the interpreter ran before the permit`) |
| M11 | permit frame (parent) | parent skips the manager identity-proof check before permitting (`client.go:177`) | ParentWithholdsPermitOnProofMismatch | **KILLED** (forge worker earned a permit → row red) — this is R1's M9/R-D survivor, now killed |
| M12 | worker revalidation | worker's `validateDerivedEnvironment` result ignored (`worker.go:270`) | WorkerRevalidatesDerivedEnvironment | **KILLED** (7/7 subtests: the interpreter ran for mis-derived requests) |
| M13 | launcher dispatch | `main()` ignores the sidecar (`main.go:105`) — built binary | ProductionBinaryRefusesWithoutSeam, NativeLauncherWorkerModeWins | **KILLED** (launcher ran the CLI instead of refusing `control_unavailable`) |
| M14 | admission honesty | `Launch` skips the preflight (`scriptworker.go:155`) | LaunchRefusesAtPreflight, ProductionBinaryRefusesWithoutSeam | **KILLED** |
| M15 | R-G guard | enforced commands no longer skipped by `activeScriptCommands` (`targets.go:431`) | install native-launcher + flip rows | **KILLED** at `install.Project` (`active script commands must be unique`) |
| M16 | launcher staging | enforced names not excluded from the ordinary shim loop (`targets.go:281`) | install native-launcher row | **KILLED** at `install.Project` (`two staged targets claim the live path …/.agents/bin/enforced-skill-tool`) |

(M6 as first written was a compile error — `value` unused — and is not counted; M6b is its compile-safe form.)
Score: 16 executed, **14 killed**, 2 survivors, both explained (M4 by construction, M8 a row gap with the
killing row shape supplied). Every one of the note's four requested shapes was run: pass one reserved name
(M1/M2/M5), keep inherited PATH (M3a/M3b), skip the proxy scrub with network=none (M4/M5), skip runtime-area
binding (M6b).

## 6. Probes (clone `/tmp/r2p`, untracked `internal/scriptworker/zz_probe_test.go`, `?? …` only status line; deleted with the clone)

| Probe | Result |
|---|---|
| **Built production binary in worker mode holds the interpreter until `permit`** (`TestZZProbeBuiltWorkerHoldsInterpreterUntilPermit`): raw session against `/tmp/…/curator-scriptworker-prod-*/curator`, `ready` received with the launcher's own hash, parent silent 1.5 s → marker absent; `shutdown` instead of permit → worker exits 0, marker still absent | **PASS** (6.4 s) — R-C proven on the production image, not only on the test binary |
| **Launcher streams** (`TestZZProbeLauncherStdinAndStreams`): sidecar-paired copy of the built binary via `RunShim` behind the seam — exit 7 propagated, argv `a`,`b c` verbatim after the runtime entry, stdin `payload` delivered, stderr `err-line` forwarded, `HTTP_PROXY` withheld; stdin of 64 MiB+1 refused before launch with exit 1 `script_execution_worker_protocol_invalid: launcher standard input exceeds the session bound` | **PASS** (4.1 s) — confirms results.md's R-B statement ("the launcher works without pass-through"; bounded, non-interactive) |
| **Sidecar gate negatives** (`TestZZProbeSidecarGateNegatives`): unknown field, version 2, `bash-v1`, relative `runtime_entry`, schema 7, `exec: ["../evil"]`, garbage → all seven refuse `script_execution_package_influence_forbidden`; `ParseDeclaredCapabilities` refuses `/usr/bin/git`, `../git`, `git tool` | **PASS** — no committed negative row exists for `LoadShimSidecar` (residual R-c) |
| **Secrets under a normalised name** (`TestZZProbeSecretsNotResolvedUnderNormalisedName`) | **PASS on the candidate**, **FAIL under M8** (§5) |
| **Seam from the outside** (built binary, no test seam): `HOME=<tmp> CURATOR_CONFIG=<cfg with script_interpreters + bogus keys "script_execution_controls","mandatory_controls","preflight.force"> CURATOR_MANDATORY_CONTROLS=all CURATOR_SCRIPT_PREFLIGHT=complete CURATOR_FORCE_CONTROLS=1 curator skill check <enforced skill>` | rc=1, `error: script_execution_control_unavailable … cannot apply mandatory controls: inventory-controls-applied, closed-script-capability-evidence-record` — exactly the two R3 rows, nothing else named |

## 7. Findings

No blocking finding. The candidate implements the brief's R2 scope, the rulings R1–R5 hold, and the honest
partial list ("Undone: none") is accurate for the brief's scope. Non-blocking residuals for the orchestrator's
list (R3/R5 owners unless noted):

- **R-a (R3 needs it): the derived path set is not carried.** `DerivedProfile` carries `ProjectRoot` only; a
  portable `filesystem` path set derives the project root as working directory but the set itself is dropped,
  and core §4.1.1 says it "derives exactly those paths beneath the canonical project root" — the Linux
  `filesystem-write-confinement` (Landlock write rights) of R3 needs those paths in the profile/request.
- **R-b (row gap, ~15 lines): secrets row shape.** Add the §6 probe row (host carries the upper-cased,
  underscore-normalised name of a declared secret, no `env_read`; assert nothing containing the value reaches the
  interpreter) — it kills M8, which the committed row survives.
- **R-c (row gap): `LoadShimSidecar` negatives.** The gate is correct (§6) but has no committed negative row;
  a table over the seven shapes above belongs beside `TestNativeLauncherWorkerModeWinsOverShimDispatch`.
- **R-d (R5, Windows qualification with real images): farm copies of DLL-dependent images.** On Windows farm
  entries are byte copies (the only privilege-free, non-hard-link option); a copy of `python.exe`, or of Git for
  Windows' `cmd\git.exe`, cannot load its side-by-side DLLs / relative `mingw64` tree from the farm directory, so a
  bare-name spawn from inside a script would fail to start there — reasoned from the Windows DLL search
  order, not observed in a run (the interpreter itself is launched by canonical
  path and is unaffected; `sys.executable`/`process.execPath` remain canonical). `DefaultExecSearchDirs` on
  Windows is `System32` + `SystemRoot` only, so `exec: ["git"]` is unresolvable (reported) on Windows until an
  operator mapping exists. Both are stated in code comments; neither is exercised by a real interpreter yet.
- **R-e (R3 stream model, restated precisely).** stdin: a pipe/file is pre-read into the request (bound 64 MiB,
  `maxProtocolFrame`), a terminal binds the null device (`os.ModeCharDevice` test in `runEnforcedShim`);
  stdout/stderr are captured into the shared 16 MiB `defaultOutputLimit` and written after the interpreter
  exits (overflow → `script_execution_worker_protocol_invalid`). Interactive commands and streaming output need
  the pass-through binding (a protocol change: inherited descriptors/handles or a side channel). Not a blocker:
  production refuses every enforced launch until R3 anyway.
- **R-f (reporting): the per-invocation `DerivationReport` is not surfaced.** `RunShim` drops `Result.Report`;
  withheld names are reported at install time (result messages, proven by the install row) but not at
  invocation (§3.1: "MUST report every env_read entry it withholds" — install-time reporting satisfies the
  letter; R3's operator-selected diagnostic destination is the natural place for the invocation record).
- **R-g (hardening): launcher forbidden roots.** `RunShim` forbids `sidecar.RuntimeDir` and the bin
  directory; the project root and the snapshot store root are not in the list, so an operator binding that points
  into the checked-out project (a repo-local venv) is only prevented by operator hygiene. Interpreter selection is
  operator config (`script_interpreters`, path env-overridable via `CURATOR_CONFIG` like every other config
  value — same posture as go-v1's `CURATOR_GO`/`GOROOT`).
- **R-h (row gap): global-scope enforced install.** The forwarding-mirror exclusion (`global.go:438-449`) has
  no row; both install rows are project scope.
- **R-i (cosmetic):** `writeShimDiagnostic` prefixes an error whose `Error()` already carries
  `script-worker-v1 <code>: `, so launcher stderr reads `script-worker-v1 script_execution_worker_protocol_invalid: script-worker-v1 script_execution_worker_protocol_invalid: …` (seen in §6).
- **R-j (robustness, fail-closed by design):** `parseHostEnvironment` refuses the whole invocation with
  `worker_protocol_invalid` when the manager's own `os.Environ()` contains an entry without `=`.
- **R-k (interpretation, spec-consistent):** "offline environment configuration" is implemented as the absence
  of every proxy/resolver name plus the recorded `offline-environment` mode; no positive offline value is set for
  either interpreter (the spec names none). `M4` shows the scrub is redundant by construction.
- **R-l (cost bound):** each enforced command installs a full copy of the manager binary (19.4 MB on darwin/amd64)
  under `.agents/bin/`; go-v1 compiled shims are also binaries there, and `.agents` is a generated path the
  gitignore gate covers.
- **R-m (test hygiene):** `internal/install` `enforcedBinaries` builds curator + stub into
  `os.MkdirTemp("", "curator-enforced-")` and never removes it (no `TestMain` cleanup, unlike
  `internal/scriptworker`'s `curatorOnce`); ~20 MB leaks per `go test` process of the package (three from my
  reruns removed by hand; 11 older ones from the producer's runs are still in `$TMPDIR`).
- R1's R-E (worker pre-exec re-verification without a replacement row) and R-B stay as stated; R-A, R-C, R-D
  (M11 killed), R-F, R-G, R-H, R-I are closed by this revision.

## 8. What R3 must pick up (from the undone list + residuals)

1. The two table rows: `inventory-controls-applied` and `closed-script-capability-evidence-record` (probe once
   pre-worker-launch, apply exactly the available/host-conditional controls, one closed evidence record validated
   by the parent **before** the `permit` — the frame is in place: `client.go:171-190`).
2. Pass-through stream binding (R-e) if interactive/long-output commands are in R3's scope; otherwise state the
   bound in docs.
3. Carry the derived path set (R-a) for Landlock write rights.
4. Surface the invocation `DerivationReport` through the evidence/diagnostic destination (R-f).
5. Rows R-b, R-c, R-h; hardening R-g; cosmetic R-i; test hygiene R-m.
6. R5: Windows qualification with real `python.exe`/`node.exe` and a declared `exec` (R-d).

## 9. Verdict

**ACCEPT** — `accept_cr(TASK-260916-1h82gq, revision=3, evidence=TASK-260916-1h82gq_review-verdict-rev3.md)`.

Basis: exact tree proven (worktree, patch, three clones, gate commit); all four derivation cases and the nine R2
controls driven at `runSession`/`Launch`/`RunShim`/`install.Project` with the real worker process, the built
production binary as launcher/worker, and fresh-process refusal rows; 25/25 R2 ledger rows pass on ubuntu,
macos and windows with 0 skips; 14/16 reviewer mutants killed with both survivors bounded; the seam has no
production writer and the built binary refuses under every config/env attempt; R-A/R-C/R-G/R-H/R-I closed; docs
and CHANGELOG present. Reviewer checklist items 14–17 checked on the basis above.
