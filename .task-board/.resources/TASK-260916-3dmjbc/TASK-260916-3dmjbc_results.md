# TASK-260916-3dmjbc results (R1 script manager/worker invocation)

Producer: developer. Worktree: `.temp/STORY-260822-2h0v9j/worktree` at
`f0a92b8b` plus uncommitted R1 changes (no commit — handoff snapshots the
working tree). Spec pin: `87a0d0060bad64ab883d007dcdf35df7485368bf`
(protocol 1.0.0-rc.9), vector `script-host-execution-policy.json` read from
`git show` of that pin.

## Outcome: R1 complete, gate green on the candidate tree

Enforced `node-v1`/`python3-v1` script commands now enter the spec
preflight and refuse `script_execution_control_unavailable` naming the
five not-yet-implemented mandatory controls; unknown policies keep the
unchanged `script_execution_policy_unsupported` refusal; the fixed
manager-owned worker invocation path behind the preflight is real
production code proven at the process boundary. No enforced script
launches uncontained: install publishes no shim for enforced commands,
and the launch entry refuses before any worker starts.

## Design

Process graph as implemented (three nodes; no fourth `exec` node until
R2 derives `exec`):

```text
manager parent (install preflight / Launch)
  -> identity-verified manager-owned script worker
       (__curator-script-worker-v1 hidden re-execution of the installed manager)
       -> identity-verified interpreter for the declared identifier
```

- Identity seam reuse (no fork): `internal/godriver/identity.go` gains
  four exported wrappers (`ResolveManagerIdentity`,
  `MatchesExpectation`, `CanonicalPhysicalPath`, `HasMultipleLinks`)
  over the unchanged go-v1 checks. `internal/scriptworker` resolves the
  manager and the interpreter through them: canonical physical path
  (junction-aware on Windows), regular file, single link / reparse
  rejection, open-then-hash. The go-v1 diff is 30 added lines, zero
  modified lines.
- Interpreter resolution (`internal/scriptworker/interpreter.go`) takes
  only the closed identifier plus the operator-trusted
  `script_interpreters` machine-config mapping (new optional
  `internal/config` key, closed to `node-v1`/`python3-v1`, absolute
  paths only) and forbidden roots. PATH, manifest values, repository,
  runtime root, and `.agents/bin` are not inputs and cannot select the
  executable. Identity is re-proved at the launch boundary.
- Nonce protocol (`internal/scriptworker/protocol.go`):
  length-prefixed JSON frames (64 MiB bound), one `request` carrying a
  fresh 64-hex-char nonce, worker `ready` with the identity proof,
  single `result` (stdout/stderr/exit/started/overflow), explicit
  `shutdown`, typed `failure`. Every response is bound to the request
  nonce; `started` must be exactly 1.
- Manager parent (`internal/scriptworker/client.go`): resolve both
  identities, create the operation-private tmp/config/cache area,
  re-verify at the launch boundary, spawn the fixed hidden mode over
  anonymous pipes (new Unix session/process-group; Windows private Job
  Object with kill-on-close), handshake, return the child exit status,
  terminate and join the complete worker domain, remove private state.
- Worker (`internal/scriptworker/worker.go`): prove own identity,
  verify the interpreter file (at accept and again immediately before
  exec), validate every request field (package-shaped values refuse as
  `script_execution_package_influence_forbidden`), bind stdin
  explicitly (payload or null device — never closed, never inherited),
  run exactly the verified interpreter against the manager-derived
  runtime entry with verbatim args. Shebang/extension/association inert.
- Preflight (`internal/scriptpolicy`): the 11-control table in vector
  order; R1 implements fixed-process-graph,
  worker-identity-verification,
  interpreter-resolution-and-identity-verification,
  operation-private-runtime-area, explicit-standard-stream-binding,
  worker-domain-teardown. Missing (R2/R3):
  `manager-built-environment`, `manager-built-path`,
  `offline-network-configuration`, `inventory-controls-applied`,
  `closed-script-capability-evidence-record`. `Admit` returns
  `script_execution_control_unavailable` naming them for implemented
  enforced commands; unknown policies return the byte-identical
  `script_execution_policy_unsupported`. `scriptworker.Launch` checks
  the same table before anything starts. No env/flag/config switch
  opens the preflight.
- `cmd/curator/main.go` dispatches `__curator-script-worker-v1` with
  exactly one argument, mirroring the go-v1 worker dispatch.

Deferred honestly to R2/R3 (noted, not hidden): declaration-derived
environment/PATH, offline configuration, native inventory probing and
application, the closed `script-capability-evidence-v1` record, and the
native enforced-shim launcher (install currently publishes no enforced
shim, per the preflight). R1 proves the worker invocation the launcher
will start.

## Row table (production boundary per row)

| Row | Production boundary | Result |
|---|---|---|
| `install.TestEnforcedScriptCommandIsRefusedAtInstall/node-v1,python3-v1` | `install.Project` → preflight refusal + missing-control names, shim absent | PASS |
| `install.TestDeclaredOnlySchema8ScriptCommandInstallsUnchanged` | `install.Project`, declared-only unchanged | PASS |
| `install.TestActiveScriptCommandsRefusesEnforcedCommand` (+ unknown-policy arm) | shim-writer guard | PASS |
| `install.TestDraftAuditEnforcedScriptRefused`, `TestDraftLocalEnforcedCommandRefused` | draft lanes, same refusal | PASS |
| `skillcheck.TestValidateRejectsEnforcedScriptExecutionPolicy` | `skill check` validation | PASS |
| `scriptpolicy.TestScriptExecutionOptInCases` (6/6) | manifest accept + enforced classification + preflight code | PASS |
| `scriptpolicy.TestARefusalPrecedesEveryWorkerSurface` (2 interpreters + unknown-policy) | admission codes before any worker surface | PASS |
| `scriptpolicy.TestMandatoryControlsMatchTheSuite` | 11-control table vs vector, in order | PASS |
| `scriptpolicy.TestPreflightRefusalCases` (2 refusal + 3 R3-owned asserted present) | install/invoke refusal, no worker | PASS |
| `scriptpolicy.TestAdmitRefusesUnknownPolicyUnchanged`, unit rows | two-tier admission | PASS |
| `config.TestParseScriptInterpreters` | machine-config seam, closed + absolute | PASS |
| `scriptworker.TestRunSessionHappyPath` | manager→worker→interpreter at process boundary: recheck, nonce, proof, verbatim args, exit status, no env inheritance, private-area cleanup | PASS |
| `scriptworker.TestRunSessionRechecksIdentityAtLaunchBoundary` | replaced manager bytes never run | PASS |
| `scriptworker.TestClientRejectsUnknownNonce` (+ bound-nonce control, failure passthrough) | nonce binding | PASS |
| `scriptworker.TestRunSessionTerminatesDescendants` | 120 s descendant reaped before return | PASS |
| `scriptworker.TestRunSessionRejectsPATHInterpreter` (refuse + configured-wins arms) | PATH never consulted | PASS |
| `scriptworker.TestBuiltCuratorWorkerHandshake` (+ exact-arg dispatch arm) | built `./cmd/curator` in hidden mode | PASS |
| `scriptworker.TestScriptWorkerHappyPath` | raw worker: proof, verify, run, verbatim argv, cwd | PASS |
| `scriptworker.TestScriptWorkerReturnsTheChildExitStatus` | exit status + stderr verbatim | PASS |
| `scriptworker.TestScriptWorkerBindsStandardInput` (payload + null arms) | explicit stdin binding | PASS |
| `scriptworker.TestScriptWorkerIgnoresTheShebang` | shebang inert | PASS |
| `scriptworker.TestScriptWorkerRejectsMalformedSessions` (5 rows) | framing/version/nonce, interpreter never runs (marker) | PASS |
| `scriptworker.TestScriptWorkerRejectsForgedWorkerIdentity` (copied + modified) | path/hash/size proof | PASS |
| `scriptworker.TestScriptWorkerRejectsSubstitutedManager` (link-resolves + retarget-refuses) | launcher-link semantics | PASS |
| `scriptworker.TestScriptWorkerRejectsTamperedInterpreter` (wrong hash + replaced file) | pre-exec verification | PASS |
| `scriptworker.TestScriptWorkerRejectsPackageSelectedPrograms` (7 rows) | closed interpreter set, absolute regular runtime entry | PASS |
| `scriptworker.TestScriptWorkerRejectsReplayedNonce` | shutdown nonce binding | PASS |
| `scriptworker.TestScriptWorkerModeIsNotReachableThroughPackageData` | hidden mode not data-selectable | PASS |
| `scriptworker.TestResolveInterpreter*` (8 tests) | config-only resolution, unknown/unconfigured/PATH/manifest/runtime-root/symlink/hardlink/directory/relative/absent refusals, replacement re-proof, primitive reuse | PASS |
| `godriver` identity + worker subsets (post-change tree) | no go-v1 regression | PASS (see Bounds) |
| `ledger-consistency.sh` (251 rows) | new ledger rows compile on linux/darwin/windows | PASS |
| `golangci-lint` (scriptworker/scriptpolicy/config), `gofmt`, `go vet` | static gates | CLEAN |
| Cross-compile `GOOS=windows/linux scriptworker`, full `go build ./...` | build not broken | PASS |

## Mutant table (narrowing mutants, each killed by its named row)

| Mutant | Killer row | Evidence |
|---|---|---|
| M1 drop launch-boundary recheck (`identity.Verify` + `VerifyInterpreter` removed) | `TestRunSessionRechecksIdentityAtLaunchBoundary` | FAIL as required: `script_execution_worker_protocol_invalid: cannot read the ready message` (replacement output is not a session) instead of `script_execution_worker_identity_invalid`; 121 s incl. 60 s ctx bound. Tree restored, 0 modified. |
| M2 drop nonce check (`receive` accepts any nonce) | `TestClientRejectsUnknownNonce` | FAIL as required in 0.34 s (foreign nonce accepted). Tree restored. |
| M3 accept PATH interpreter (`LookPath` fallback when unconfigured) | `TestRunSessionRejectsPATHInterpreter` + `TestResolveInterpreterIgnoresUserPATH` | FAIL as required: `unconfigured resolution with PATH decoy = <nil>`. Tree restored. |
| M4 skip worker-domain teardown (no group/job kill, handle never released) | `TestRunSessionTerminatesDescendants` | FAIL as required: `descendant 46968 outlived the worker domain teardown` (16.36 s). Tree restored, 0 modified. |

No survivors. Each mutant narrows (drops one check) rather than deleting
the feature, and each killer names the production call site
(`runSession`/`receive`/`ResolveInterpreter`/teardown).

## Ratio line

Spec vector `script-host-execution-policy.json` top-level keys: 12.
Consumed by this leaf: 7 (6 full + 1 partial) —
`schema_version`, `protocol_version`, `execution_policy`,
`interpreters`, `opt_in_cases` (6/6), `mandatory_controls` (11/11
names in order + implementation table), `preflight_cases` (partial:
2/5 refusal cases driven, 3/5 host-conditional success cases asserted
present and owned by R3). Still `refusedBeforeReached` (4):
`capability_derivation_cases`, `capability_evidence_cases`,
`capability_evidence_record`, `native_control_inventory` (every enforced
shape still refused at admission — proven by
`TestARefusalPrecedesEveryWorkerSurface`). Still `notImplementedYet`
(1): `audit_label_cases` (R4). Behavioral cases driven: 6 opt-in + 2
preflight + 11 controls; 27 worker/interpreter rows at the process
boundary behind them.

## Windows proof status

Windows proof is structural, not hosted: the worker/process rows are
portable Go (new session vs Job Object behind `teardown_unix.go` /
`teardown_windows.go`, `process_alive_*_test.go` per platform, stub
interpreter compiled per `GOOS`), `GOOS=windows go build` passes,
`ledger-consistency.sh` proves all 10 new ledger rows compile into the
windows build, and 8 `internal/scriptworker` rows plus 2
`internal/scriptpolicy` rows are registered `must_run_on
linux,darwin,windows` with no skip tolerance. Hosted windows-latest
execution happens in the landing suite (handoff runtime); this report
claims no hosted Windows result. No ARM64 lane is claimed either.

## Bounds

- `internal/godriver` full package: the 600 s default-timeout run and a
  240 s `Test(A|B|C)` chunk both failed on this host with
  `go-v1 process_timeout: Go probe exceeded its deadline` in fixture
  setup — stub-process spawn latency under host memory exhaustion
  (swap 78% full, load 3–10, a trivial no-op process spawn measured at
  85 s wall time, OOM kills of unrelated processes including
  `go version` and a `python3` edit script). The godriver change is 30 added / 0 modified lines (new
  exports only, never called by godriver paths); identity + worker
  subsets reran green post-change
  (`TestExecutableIdentity*`, `TestWorker*`,
  `TestBuildAcceptsAManagerStartedThroughALauncherLink`,
  `TestBuildTerminatesTheCompleteWorkerDomain`,
  `TestPerFileSizeLimitIsReallyApplied`,
  `TestProbeRejectsAnUncoveredPlatformBeforeTheWorker`). Full-package
  green must come from the landing suite on healthy infra.
- M1's kill takes 61 s by construction (the replacement is not a
  protocol speaker, so the manager waits out the 60 s invocation ctx).
  Normal-suite time for that row is <1 s (recheck refuses fast).
- The 120 s sleeper in the teardown test exits on its own if a mutant
  run is interrupted; no strays were left (`ps` verified).
- `docs/troubleshooting.md` gains the enforced-script section;
  `CHANGELOG.md` `Unreleased → Added` carries the R1 entry. No
  behaviour change for declared-only commands, build commands, or the
  go-v1 worker (goldens/rows untouched; only additive exports).

---

# TASK-260916-3dmjbc results — Revision 2 appendix (Windows-only fixture rework)

Rework scope (orchestrator `3dmjbc-rework-1.md`): revision 1 gate FAILED
only on `Test (windows-latest)` (run 35574902088); Linux/macOS lanes
green. Three Windows-only failures, all fixture/portability, no design
change. Rulings R1–R5 unchanged. Continued from the revision-1 tree in
the Story workspace (no checkout/clean/stash).

## What changed (5 test-only edits, 0 production lines)

1. `internal/config/scriptinterpreters_test.go` —
   `TestParseScriptInterpreters`: the accepted row built its bindings
   from POSIX literals (`/opt/...`, `/usr/...`), which
   `filepath.IsAbs` rejects on Windows (no volume). The row now builds
   both bindings from `t.TempDir()` (absolute on every OS;
   `strconv.Quote` for JSON), so on Windows it is also the
   volume-absolute acceptance row. Production validation needed no
   change: `parseScriptInterpreters` gates on `filepath.IsAbs`, which
   on Windows accepts `C:\...`/UNC and rejects drive-relative
   (`C:tools\...`) and rooted-without-volume (`\tools\...`) shapes —
   two new Windows-only rejection rows pin that, mirroring the POSIX
   relative-path row. (`skills_root: /tmp/skills` in the shared base
   fixture is safe: `Parse` never `IsAbs`-checks it — verified the
   only `IsAbs` in `internal/config/config.go` is the new
   interpreter gate.)
2. `internal/scriptworker/worker_test.go` —
   `TestScriptWorkerRejectsForgedWorkerIdentity`: the copied
   (`curator-copy`) and modified (`curator-modified`) manager binaries
   are now named with the platform executable suffix (`.exe` on
   Windows) via the new `executableFixtureName` helper, mirroring the
   go-v1 worker launcher fixtures
   (`internal/godriver/identity_test.go`). Without the suffix the
   copy is not executable on Windows and the launch fails before the
   worker can refuse. Negative semantics unchanged: the rows still
   assert the refusal class (`script_execution_worker_identity_invalid`
   via `expectFailure`) and the interpreter-never-ran marker, not the
   exec error.
3. Same file — `TestScriptWorkerRejectsSubstitutedManager`: the
   installed (`curator-real`), attacker (`curator-other`), and link
   (`curator`) names take the same suffix, so the launch-through-link
   arm and the retarget-refuses arm execute on Windows.
4. `internal/scriptworker/client_test.go` —
   `TestRunSessionRechecksIdentityAtLaunchBoundary`: the
   `curator-manager` copy takes the same suffix (hygiene; the row
   refuses at the recheck before exec and already passed on Windows).
5. `internal/scriptworker/main_test.go`: the `executableFixtureName`
   helper itself.

Sweep of every other new row for the same two assumptions (absolute
paths, `.exe`): all other executed binaries already carry the suffix
(`os.Args[0]` test binary, `builtCuratorBinary`, `stubInterpreterBinary`,
PATH decoys, `goBinaryName`); remaining POSIX literals in the config
test are rejected before any path check (unknown identifier,
non-object) or never path-checked (`skills_root`); interpreter copies
that are only hashed, never executed (`evil`, `interp-hard`, `interp`,
`nested`), need no suffix because the Windows executable predicate is
`mode.IsRegular()`; `process_alive_*`/`teardown_*` are already
per-platform files; install/skillcheck/scriptpolicy rows use no
executables or absolute paths and passed on Windows in revision 1.

## Windows proof status (revision 2)

Still structural on this host (macOS): the fix is verified by
`go test` of the touched rows on darwin, `go vet` of both packages,
and a `GOOS=windows` compile of the scriptworker + config test
binaries (see commands below). Hosted windows-latest execution of the
three fixed rows happens in the landing suite (handoff runtime); no
hosted Windows result is claimed here. Linux/macOS lanes were already
green and are unaffected (POSIX behavior of every edited row is
byte-identical: no suffix appended, TempDir paths absolute as before).

## Verification (this revision, macOS arm64, zsh, `set -o pipefail`)

- `go test -count=1 ./internal/config -run TestParseScriptInterpreters -v` → exit 0 (7/7 subtests).
- `go test -count=1 ./internal/scriptworker -run 'TestScriptWorkerRejectsForgedWorkerIdentity|TestScriptWorkerRejectsSubstitutedManager' -v` → exit 0.
- `go test -count=1 ./internal/config/` → exit 0 (full package, 0.7 s).
- `go test -count=1 ./internal/scriptworker/ -run TestRunSessionRechecksIdentityAtLaunchBoundary` → exit 0.
- `go test -count=1 -timeout 8m ./internal/scriptworker/` → exit 0 (full package, 8.0 s).
- `go vet ./internal/scriptworker/ ./internal/config/` → exit 0.
- `GOOS=windows go vet ./internal/scriptworker/ ./internal/config/` → exit 0 (the Windows test builds, including the new suffix helper and Windows-only config rows, typecheck).
- Formatting: the `gofmt` binary was OOM-killed on spawn 4 times (exit 137) under the host memory pressure already recorded in revision-1 bounds, so all 4 edited files were verified clean through the identical `go/format` engine via a built checker binary (exit 0, `internal/config/scriptinterpreters_test.go`, `internal/scriptworker/main_test.go`, `worker_test.go`, `client_test.go` all `clean`).
- Windows-only config rows (`C:tools\node.exe`, `\tools\node.exe` rejected; TempDir volume-absolute accepted) follow directly from `filepath.IsAbs` volume semantics; fixture JSON decoding verified (`C:tools\node.exe`, `\tools\node.exe`).
- Mutants: accepted from revision-1 evidence (above); this revision touches test fixtures only, so no mutant rerun is owed — the gates and their killers are unchanged.

## Bounds

- The modified-binary arm flips the last byte of a copied PE on
  Windows; the Windows loader does not verify the optional-header
  checksum for user-mode execution, so the copy still starts and the
  worker refuses by hash mismatch — same as POSIX. If a future
  Windows host ever refused to load it, the failure would be loud
  (`t.Fatal` at start), not a silent pass.
- Symlink arms keep their host-capability skips; windows-latest
  runners permit symlink creation, as revision 1 showed (the rows
  failed at exec, not at link creation).

---

# TASK-260916-3dmjbc results — Revision 3 appendix (F1 rework)

Rework scope (orchestrator `3dmjbc-rework-2.md`): verdict rev2
CHANGES_REQUESTED with ONE blocking finding F1 (verdict
`TASK-260916-3dmjbc_review-verdict-rev2.md`, RUN-260921-bd4d65).
Continued from the revision-2 tree in the Story workspace (no
checkout/clean/stash). Fixed exactly F1; R-H left to R2/R3 per the
rework's same-lines rule; R-A/B/C/F restated as bounds below exactly as
the verdict lists them.

Accepted external evidence: rev2 hosted gate run 35579456976 — all 11
lanes success (Test/Race/Gate self-test ubuntu+macos+windows, Lint,
Interop, Naming), gate head tree = rev2 candidate tree (verdict §Exact
candidate identity). Not rerun here; the rev3 gate runs in the landing
suite (handoff runtime).

## F1: the fix

`readInterpreterIdentity` (`internal/scriptworker/interpreter.go`)
accepted any regular single-linked hashed file, so a POSIX `#!/bin/sh`
wrapper (pyenv/asdf/volta shims) resolved, passed both worker
verifications, and ran `/bin/sh` between the worker and the
interpreter; on Windows a `.cmd`/`.bat` binding runs through `cmd.exe`
(which re-parses the verbatim argv — Go 1.26 `os/exec/exec.go:400`),
and an extensionless absolute path is rewritten by `lookExtensions`
(`os/exec/lp_windows.go:76-109`: an extension outside PATHEXT falls
through to `lookPathExts`, which appends `.com`/`.exe`/`.bat`/`.cmd`),
so the hashed file and the executed file differ.

Fix (reviewer-verified shape, no fork):

1. `internal/godriver/identity.go`: new exported
   `NativeExecutableHeader(header []byte) bool` over the platform
   `nativeExecutableHeader` (ELF/Mach-O incl. fat binaries on unix, MZ
   on Windows) — the same judgment `validateLauncher`
   (`session.go:536-559`) applies to the Go launcher. The godriver
   diff stays additive-only (+40/−0 total: five exported wrappers,
   zero modified lines).
2. `readInterpreterIdentity` (shared by manager-side
   `ResolveInterpreter`, launch-boundary `VerifyInterpreter`, and the
   worker's accept + pre-exec `verifyInterpreterFile`):
   - native-image gate: the first 8 bytes are read from the same open
     handle that is hashed (`ReadAt`, offset untouched for the hash)
     and judged by `godriver.NativeExecutableHeader`; anything else
     refuses `script_execution_worker_identity_invalid` ("not a
     native interpreter image").
   - Windows name gate: `interpreterImageNameOK` requires the `.exe`
     extension (`platformRequiresExecutableExtension` const in
     `interpreter_unix.go`/`interpreter_windows.go`; the match
     itself in the shared `hasWindowsExecutableExtension`,
     `EqualFold` on `filepath.Ext`) so `lookExtensions` returns the
     verified path unchanged.
3. `worker.go` comment corrected: the old
   "shebang/extension/association inert" claim is replaced (the
   runtime entry's bytes stay inert — the entry travels as argv,
   never as the executed program;
   `TestScriptWorkerIgnoresTheShebang` unchanged and green).
4. `docs/troubleshooting.md`
   §`script_execution_worker_identity_invalid`: "not a native
   interpreter image" cause (POSIX `#!` shims, Windows
   `.cmd`/`.bat`, extensionless Windows paths) + `.exe` remedy; one
   clause added to the CHANGELOG R1 bullet.

## Rows (production boundary per row)

| Row | Production boundary | Result |
|---|---|---|
| `scriptworker.TestResolveInterpreterRejectsWrapperImage` (POSIX: shebang-wrapper + plain-text; Windows: cmd-wrapper, bat-wrapper, extensionless-native-image, exe-named-text) | `ResolveInterpreter` → `script_execution_worker_identity_invalid` before any worker exists | PASS (POSIX arms executed here; Windows arms compile-proven + ledger-registered) |
| `scriptworker.TestScriptWorkerRejectsWrapperImage` (POSIX: working shebang wrapper that would exec the stub; Windows: .cmd wrapper, extensionless native copy, .exe-named text; request carries the wrapper's own correct hash/size) | raw hidden-mode worker → `expectFailure(worker_identity_invalid)`, marker never written | PASS (POSIX arm executed here incl. marker check) |
| `scriptworker.TestHasWindowsExecutableExtension` (10-row table) | helper-level bound for the shared name predicate; runs on every OS | PASS |
| ledger `internal/scriptworker :: TestResolveInterpreterRejectsWrapperImage`, `:: TestScriptWorkerRejectsWrapperImage` | `must=linux,darwin,windows skip=-` like the existing worker rows | ledger-consistency ok (253 rows) |

## Mutant table (revision 3)

| Mutant | Killer row | Evidence |
|---|---|---|
| M-F1a drop the header check (`interpreter.go`: `if false && !godriver.NativeExecutableHeader(...)`) | `TestResolveInterpreterRejectsWrapperImage` + `TestScriptWorkerRejectsWrapperImage` | FAIL as required: `wrapper binding "shebang-wrapper" error = <nil>, want script_execution_worker_identity_invalid` (both POSIX arms); `worker sent "ready", want a failure`. Tree restored (`grep MUTANT` = 0), rows green again. |
| M-F1b force `hasWindowsExecutableExtension` true | `TestHasWindowsExecutableExtension` | FAIL as required (6/10 arms). Tree restored. |

No survivors. M1–M4 (rev1) and M1a/M2–M9 (rev2 reruns)
accepted from prior evidence; this revision's gates are M-F1a/M-F1b.

## Implemented-controls table (updated)

No flag flipped: `interpreter-resolution-and-identity-verification`
was already `Implemented`, and F1 makes that flag truthful
(wrappers/batch/extensionless bindings refused at resolution, at the
launch boundary, and at the worker's accept + pre-exec verifications).
The five R2/R3 missing controls are unchanged, so install/invoke still
refuse `script_execution_control_unavailable` and no enforced script
launches.

## Windows proof status (revision 3)

POSIX arms executed on this host (macOS). The Windows arms
(`.cmd`/`.bat`/extensionless/`fake.exe` at both levels) are proven
structurally here: `GOOS=windows go vet` of the scriptworker package
+ tests, the `hasWindowsExecutableExtension` table green on this
host, ledger rows registered `must=linux,darwin,windows skip=-`
with zero skip tolerance, and the `lookExtensions`/batch semantics
cited from the Go 1.26 GOROOT sources above. Hosted windows-latest
execution happens in the landing suite (handoff runtime); no hosted
Windows result is claimed here.

## Verification (this revision, macOS x86_64, bash)

- `go build ./...` → exit 0.
- `go test -count=1 -timeout 8m ./internal/scriptworker/` → exit 0 (full package incl. 3 new tests).
- `go test -count=1 ./internal/scriptpolicy/ ./internal/config/ ./internal/skillcheck/` → exit 0.
- `CURATOR_CONFORMANCE_ROOT=<87a0d006 archive>/conformance/v1 go test -count=1 ./internal/scriptpolicy/ -v` → exit 0, 13/13 PASS incl. 6/6 opt-in, 11-control table row, preflight rows.
- `go test -count=1 ./internal/godriver/ -run 'TestExecutableIdentity|TestUnixLauncherRequiresExecuteBit|TestBuildAcceptsAManagerStartedThroughALauncherLink'` → exit 0.
- `go test -count=1 ./internal/install/ -run 'TestEnforcedScriptCommandIsRefusedAtInstall|TestDeclaredOnlySchema8ScriptCommandInstallsUnchanged|TestActiveScriptCommandsRefusesEnforcedCommand|TestDraftAuditEnforcedScriptRefused|TestDraftLocalEnforcedCommandRefused'` → exit 0.
- `ledger-consistency.sh` → exit 0, 253 rows checked, both new rows `must=linux,darwin,windows skip=-`.
- `GOOS=windows go vet ./internal/scriptworker/ ./internal/godriver/` → exit 0.
- `golangci-lint run ./internal/scriptworker/... ./internal/godriver/...` → 0 issues; `gofmt -l` clean.

## Bounds (revision 3)

- R-H left to R2/R3 per the rework's same-lines rule
  (`scriptpolicy.MandatoryControls` is an exported mutable slice; no
  writer exists; hardening only). It touches `scriptpolicy.go`,
  which F1 does not otherwise touch.
- The Windows production-boundary arms execute on windows-latest in
  the landing suite; this host proves them structurally (above).
- `internal/godriver` full package accepted from the hosted lanes
  (rev2 gate green); the godriver delta is additive-only and the
  identity/launcher subsets reran green here.
- M-F1a was executed on macOS: it kills via the POSIX arms. The
  Windows `exe-named-text` arms are the header-mutant killers on
  windows-latest by construction (they pass every other gate); the
  extensionless arms kill an ext-gate mutant there.
- Non-blocking residuals R-A…R-J restated below exactly as the rev2
  verdict lists them, for the R2/R3 briefs (R-J is a review-host
  note, not a product bound).

### Verdict residuals, restated

R-A. **`scriptworker.Launch` has no production caller.** Only
`RunWorker` (hidden mode) is wired (`cmd/curator/main.go:113`).
`Launch`/`runSession`, the `config.ScriptInterpreters` →
`LaunchRequest.Interpreters` wiring and the forbidden-root list are
exercised by tests only, so the vector's
`mandatory-control-unavailable-at-invocation` case is driven at
`Launch` (helper level), not at a `cmd/curator` entry; the brief's
"installed enforced shims route through this path" (R1 item 4) is
vacuous until a launcher exists. The producer states this in
results.md §Design ("native enforced-shim launcher … deferred")
while the headline says "R1 complete" — read it as complete minus
item 4. Note for the launcher design: profile §3 says an enforced
command has "no shell, `.cmd`, or symlink shim" between the manager
and the interpreter and none of §3's launcher forms apply, so the
R2/R3 shim must be a native launcher.

R-B. **Stream model.** The worker's stdout is the protocol channel,
so interpreter stdout/stderr are captured into a shared 16 MiB
budget and returned in the `result` frame (overflow →
`script_execution_worker_protocol_invalid`), and stdin is a byte
payload copied at request time, not a live pipe/terminal. §3.1
permits a bound only on output the manager itself captures and
requires transparent forwarding for real invocations ("a script
command legitimately reads a pipe or a terminal"); the launcher of
R-A will need pass-through binding (extra inherited
descriptors/handles or a side channel), i.e. a protocol change.

R-C. **No permit step.** The worker sends `ready` and starts the
interpreter immediately (probe 2); the parent's proof check
(`client.go:147-152`) and, in R3, evidence validation run
concurrently with the interpreter, with teardown as the remedy.
go-v1 gates the build behind `permit` after validating the proof
and the evidence record. Adding a permit frame before
`runInterpreter` (worker waits for it) makes R3's evidence check a
pre-run gate as §3.1 step 6→7 reads.

R-D. **Parent-side proof checks have no process-boundary row.** M9
(drop `identity.MatchesExpectation` in `runSession`) survives every
committed row: forged identity is caught by the worker's
self-check (`worker.go:89`), which every negative row exercises; a
lying worker (test binary in a "forge" mode) would be the row
shape. Same for `result.Started != 1` and
`interpreter.matchesExpectation` on `ready`.

R-E. **Worker pre-exec re-verification** (`worker.go:113`, M6) has
no row that replaces the interpreter between accept and exec (no
worker-side hook); it is a race-narrowing recheck, stated bound.

R-F. **`operation-private-runtime-area` = roots created and torn
down only.** `PrivateTmp/Config/Cache` are validated by the worker
but not applied (no `TMPDIR/TEMP/TMP/XDG_*` binding — R2's
environment builder); `WorkingDir` is caller-selected. The
"Implemented" flag is truthful for creation/teardown, not for
application.

R-G. `activeScriptCommands` (`install/targets.go:304-307`) builds
its refusal as `fmt.Errorf("%w", Admit(...))`; once R2/R3 complete
the table `Admit` returns nil and the message degrades to
`%!w(<nil>)` — the enforced-shim writer must replace this guard,
not fall through it.

R-H. `scriptpolicy.MandatoryControls` is an exported mutable slice
(go-v1 exposes a copying function). No writer exists; hardening
only.

R-I. Docs: `script_interpreters` is documented only in
troubleshooting; `build_https` has a reference page — add a config
reference entry when the launcher lands.

R-J. Host note: unsigned fresh binaries were SIGKILLed (137) in an
exec-stall window ≈09:27–09:44Z (fresh binaries stranded ~85 s at
`_dyld_start`, RSS 12 KB, SIGKILL 137; load 3–6); all reruns above
completed after the window via health-gated drivers; hosted lanes
are the arbiter.

---

# TASK-260916-3dmjbc results — Revision 4 appendix (rework-3 Windows row fix)

Rework scope (orchestrator `3dmjbc-rework-3.md`): revision 3 gate
FAILED only on `Test (windows-latest)` (run 35586991022), one row:
`internal/scriptworker :: TestScriptWorkerRejectsTamperedInterpreter`.
The rev3 F1 Windows name gate (bindings must name the native `.exe`
image itself) trips on that row's own fixture BEFORE the tamper it
wants to prove: the interpreter copy is named extensionless (`interp`),
so the pre-tamper `ResolveInterpreter` refuses
(`worker_test.go:462: ... must name the native .exe image itself`) and
the row dies in `t.Fatal`. All other lanes green. Continued from the
revision-3 tree in the Story workspace (no checkout/clean/stash).
Rulings of `3dmjbc-brief.md` and `3dmjbc-rework-2.md` unchanged.

## What changed (the row, not the check; 0 production lines)

1. `internal/scriptworker/worker_test.go` —
   `TestScriptWorkerRejectsTamperedInterpreter`, second arm: the
   interpreter copy is named via `executableFixtureName("interp")`
   (`interp.exe` on Windows; the copied bytes are the compiled native
   stub), so the pre-tamper resolution succeeds on Windows. The
   tamper now flips the LAST byte instead of the first, keeping the
   8-byte image header intact, so the worker refusal names the
   proof mismatch (`interpreter identity proof ... does not match
   ...`) rather than the image gate. Both arms assert the
   tamper-specific reason through the new `expectFailureDetail`
   helper (code `script_execution_worker_identity_invalid` + detail
   contains `interpreter identity proof`), so the row fails on the
   wrong refusal reason instead of passing on it.
2. Same file: the `expectFailureDetail` helper itself (code +
   `strings.Contains` on the failure detail; `strings` already
   imported).
3. `internal/scriptworker/interpreter_test.go` —
   `TestResolveInterpreterRejectsSubstitution`, hard-link arm (audit
   find, same assumption class): the target/alias are named via
   `executableFixtureName`, so on Windows the row reaches the
   link-count gate it intends to prove instead of tripping the
   name gate first. (It passed either way on the code assertion;
   now it proves the intended gate on all three OSes.)

No production file touched; no ledger change (no new rows, no
renames); no docs change (the refusal class and messages are
unchanged).

## Audit: every scriptworker row that binds an interpreter fixture

Each `ResolveInterpreter` / `stubInterpreterBinary` / `copyTestFile`
site re-checked for the Windows name-gate assumption:

- Bind the stub directly (`stubinterp.exe`, native): `newWorkerFixture`
  (all `worker_test.go` rows), `newLaunchFixture` (all
  `client_test.go` rows), `TestResolveInterpreterFromOperatorConfigOnly`,
  `TestResolveInterpreterResolvesABindingThroughALink`
  (link canonicalizes to the stub), both PATH/manifest decoy rows
  (decoys never resolved; PATH decoy already `.exe`). OK.
- Copy named `filepath.Base(stub)` (keeps `.exe`):
  `TestResolveInterpreterRejectsRuntimeRoot` (forbidden-root arms
  refuse before the name gate; success arm resolves),
  `TestVerifyInterpreterDetectsReplacement` (byte-0 tamper trips
  the header gate with the asserted code on every OS — unchanged,
  platform-neutral). OK.
- Intentional negative arms (refusal expected):
  `TestResolveInterpreterRejectsWrapperImage` (incl. the
  extensionless-native-image arm, which pins the name gate) and
  `TestScriptWorkerRejectsWrapperImage`. OK.
- Manager copies all use `executableFixtureName` (rev2). OK.
- `internal/config` interpreter tests bind path strings only, never
  executed or resolved through the worker; install/skillcheck/
  scriptpolicy rows bind no interpreter files. OK.

Residual of the audit: none — every executed or resolved copy now
carries the platform suffix, and every extensionless binding left in
the package is an intentional name-gate refusal arm.

## Negative control (the detail assertion is load-bearing)

Temporarily restoring the first-byte tamper in the fixed row makes
it FAIL as required:

```text
worker_test.go:504: worker failure detail = "script-worker-v1
script_execution_worker_identity_invalid: the resolved interpreter
node-v1 is not a native interpreter image", want it to name
"interpreter identity proof"
```

i.e. a header-breaking tamper is now distinguished from the
proof-mismatch tamper the row proves. File restored afterwards;
row green again.

## Windows proof status (revision 4)

Still structural on this host (macOS): the fixed row and the
hard-link arm are verified by `go test` on darwin, `go vet`, and
`gofmt`; the Windows-specific reasoning (`.exe` name passes the
name gate, last-byte tamper preserves the `MZ` prefix the shared
`godriver.NativeExecutableHeader` judges) follows from the
production code paths, which are unchanged since the rev3 gate
where every sibling row passed on windows-latest. Hosted
windows-latest execution of the fixed row happens in the landing
suite (handoff runtime); no hosted Windows result is claimed here.
Linux/macOS lanes are unaffected (POSIX names byte-identical, no
suffix appended; the last-byte tamper is platform-neutral).

## Verification (this revision, macOS x86_64, bash, `set -o pipefail`)

- `go vet ./internal/scriptworker/ && go test -count=1 ./internal/scriptworker/` → exit 0 (full package, 8.7 s).
- `go test -count=1 -v -run 'TestScriptWorkerRejectsTamperedInterpreter|TestResolveInterpreterRejectsSubstitution|TestVerifyInterpreterDetectsReplacement|TestResolveInterpreterRejectsWrapperImage|TestScriptWorkerRejectsWrapperImage' ./internal/scriptworker/` → exit 0, 5/5 PASS.
- `go build ./...` → exit 0.
- `gofmt -l internal/scriptworker/` → empty (clean).
- Mutants: M1–M4 (rev1), M-F1a/M-F1b (rev3) accepted from prior
  evidence; this revision touches test rows only, so no production
  mutant rerun is owed. The first-byte-tamper negative control
  above is this revision's narrowing proof that the row pins the
  tamper reason.

## Bounds (revision 4)

- R-A…R-J restated in revision 3 unchanged; this revision adds no
  new residual.
- The rev3 evidence (gate run 35586991022) is accepted from the
  board's `TASK-260916-3dmjbc_change-request_rev3-validation.log`
  (single FAIL line: `internal/scriptworker ::
  TestScriptWorkerRejectsTamperedInterpreter` on windows-latest);
  not independently replayed beyond the log.

---

## Revision 4 gate proof (green) + duplicate-spawn reverification

Gate for the exact rev4 tree: run
[35590341220](https://github.com/relux-works/curator/actions/runs/35590341220)
(`gate/STORY-260822-2h0v9j/260921-104420-11521-1`, commit
`a944df03834433d590f1452afa1f3fd803b186c0`), finished `success`
(see `TASK-260916-3dmjbc_change-request_rev4-validation.log`):
`Test (ubuntu-latest)`, `Test (macos-latest)`,
`Test (windows-latest)`, `Race (ubuntu-latest)`,
`Race (macos-latest)`, `Interop conformance gate`, `Lint`,
`Naming gate`, and all three `Gate self-test` lanes green;
`Test (rose-air)` and `Candidate suite` skipped (unchanged,
bounds as before). This closes the "Windows proof status"
caveat of the Revision 4 appendix above: the fixed row
`TestScriptWorkerRejectsTamperedInterpreter` now passes on
hosted windows-latest.

Reverification note (second developer spawn, same rework-3 brief,
RUN-260921-e7b2a8): this run started concurrently with the rev4
publisher and found the work already complete. Verified, not
re-edited:

- Worktree is byte-identical to the published rev4 tree:
  `git archive HEAD` + apply of
  `TASK-260916-3dmjbc_change-request_rev4.patch`, then
  `diff -r` against the worktree (excluding `.git`,
  `.task-board`, `.temp`) → exit 0, no differences.
- rev3→rev4 patch fileset identical; the delta is the
  `worker_test.go` / `interpreter_test.go` row fix only
  (0 production lines).
- Reran on macOS x86_64, Go 1.26.0, bash, `set -o pipefail`,
  exit 0: `go build ./...`,
  `go vet ./internal/scriptworker/`, `gofmt -l` clean on the
  touched packages, `go test -count=1 ./internal/scriptworker/`
  (ok, 9.8 s), and the 5-row mask
  (`TestScriptWorkerRejectsTamperedInterpreter`,
  `TestResolveInterpreterRejectsSubstitution`,
  `TestVerifyInterpreterDetectsReplacement`,
  `TestResolveInterpreterRejectsWrapperImage`,
  `TestScriptWorkerRejectsWrapperImage`) → 5/5 PASS.
- `internal/config` / `internal/scriptpolicy` full-package runs
  were SIGKILLed by the host (load 8–14, shared runner; `signal:
  killed` at 76–90 s with zero test lines printed — the known
  exec-stall pattern, binaries killed before running anything).
  Not rerun to green locally; accepted from the green gate above,
  which ran these packages on the identical tree on all three
  OSes. These packages are untouched by rework-3.
- No code changed by this run; this section is the only outcome
  delta. Handoff republishes the identical tree.
