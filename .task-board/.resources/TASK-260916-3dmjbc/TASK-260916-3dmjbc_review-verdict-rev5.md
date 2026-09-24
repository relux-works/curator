# TASK-260916-3dmjbc review verdict — revision 5 (CR-TASK-260916-3dmjbc-5)

Reviewer: Claude (claude-opus-5), RUN-260921-f30f70, 2026-09-21. Read-only review; no edits to the
Story worktree. Verdict: **ACCEPTED** — `accept_cr(TASK-260916-3dmjbc, revision=5, evidence=this file)`.

Revision 5 is byte-identical to revision 4 (the second developer spawn republished the same tree,
results.md §"Revision 4 gate proof"), so the rev4 review note (`3dmjbc-review-rev4-note.md`) is the
brief for this verdict. Scope judged: the rev2 → rev5 delta (F1 rework + the rework-3 Windows row
fix). Everything else was accepted at rev2 (`TASK-260916-3dmjbc_review-verdict-rev2.md`) and is not
re-opened: no other bytes changed.

## Exact candidate identity

- Story worktree `.temp/STORY-260822-2h0v9j/worktree` at HEAD `f0a92b8b` + uncommitted delta; temp-index
  `write-tree` = `8324fdfc2394ea31ae8e6a54e041fa44cc95ad41` = the CR candidate tree.
- `TASK-260916-3dmjbc_change-request_rev5.patch` sha256 `25abb3a7…7b38bc8` (matches the CR);
  `cmp` against `…_rev4.patch` → byte-identical. Applied on `f0a92b8b` in four disposable clones
  (`/tmp/3dmjbc-rev5-{cand,mut,probe,delta}`), each `write-tree` = `8324fdfc…`.
- Hosted gates: rev4 run 35590341220 (`headSha a944df03`) and rev5 run 35595535426 (`headSha a0689295`,
  the CR's own validation log) — both commits are "remote gate snapshot" children of `f0a92b8b` and
  both resolve `^{tree}` = `8324fdfc…`. Both runs `success` on all 11 lanes (Test/Race/Gate self-test
  ubuntu+macos+windows, Lint, Interop, Naming); rose-air and candidate-suite skipped (unchanged bound).
- rev2 tree reconstructed from `…_rev2.patch` in the delta clone = `0b582d53…` (the rev2 verdict's value).
- Spec pin `87a0d006` (protocol 1.0.0-rc.9), `conformance/v1/manifest.json` sha256 `0e195ecd…`, served
  to the scriptpolicy rows via `CURATOR_CONFORMANCE_ROOT`.

## (1) rev2 → rev5 delta is exactly the briefed scope

`git diff <rev2> <rev5>`: 10 files, 269+/27−.

| File | Change | Judgement |
|---|---|---|
| `internal/godriver/identity.go` | +`NativeExecutableHeader(header []byte) bool` wrapping the unexported `nativeExecutableHeader` (comment updated) | Shared primitive **exported, not forked**: `session.go:555` (`validateLauncher`) and `scriptworker/interpreter.go:147` call the same platform function (`platform_unix.go:22` ELF/Mach-O/fat, `platform_windows.go:97` MZ). Base → rev5 godriver delta is 40+/0− (five exported wrappers, zero modified lines, no test change). |
| `internal/scriptworker/interpreter.go` | `readInterpreterIdentity`: name gate `interpreterImageNameOK` (line 111, before the link/open steps) + native-image gate (lines 141–150) reading 8 bytes with `ReadAt` from the **same open handle** that is hashed (`SameFile`-checked); `hasWindowsExecutableExtension` = `EqualFold(filepath.Ext, ".exe")` | Both gates sit on the single shared path: `ResolveInterpreter` (86), `VerifyInterpreter` (190, launch boundary), worker accept (`worker.go:95→247`) and worker pre-exec (`worker.go:113→247`). Order on the wire is unchanged. |
| `interpreter_unix.go` / `interpreter_windows.go` | `platformRequiresExecutableExtension` const false/true | Matches Go 1.26 `os/exec` semantics: `lookExtensions` returns a `.exe`-suffixed absolute path unchanged (extension in the PATHEXT list, or `hasExt`+exists in `findExecutable`); the worker env carries no `PATHEXT` so the default list applies. Spec §3.1 names `PATHEXT`, `.cmd` wrappers and file associations as forbidden intermediaries. |
| `worker.go` | comment only (the runtime entry's bytes stay inert; wrappers refused) | No behaviour change; `TestScriptWorkerIgnoresTheShebang` unchanged and green. |
| `interpreter_test.go` | +`TestResolveInterpreterRejectsWrapperImage` (POSIX: shebang-wrapper, plain-text; Windows: cmd-wrapper, bat-wrapper, extensionless-native-image, exe-named-text), +`TestHasWindowsExecutableExtension` (10 arms), hard-link arm names its fixtures with the platform suffix | Each Windows gate has a dedicated killer arm: `extensionless-native-image` (name gate only), `exe-named-text` (header gate only). |
| `worker_test.go` | +`TestScriptWorkerRejectsWrapperImage` (raw hidden-mode worker, request carries the wrapper's own correct hash/size, marker never written), +`expectFailureDetail`, tamper row: `.exe`-named copy + last-byte tamper + detail pinned to `interpreter identity proof` | The detail pin is load-bearing (control C-tamper below). |
| `.github/ci/platform-cases.tsv` | 2 rows `must=linux,darwin,windows skip=-` | Registered; hosted ledger-consistency ok (253 rows). |
| `CHANGELOG.md`, `docs/troubleshooting.md` | one clause in the R1 bullet; §`script_execution_worker_identity_invalid` cause/remedy for non-native images and the Windows `.exe` rule | Accurate against the code. |

Nothing outside this list changed; R-H (copying accessor) was not folded in — permitted by rework-2's
same-lines rule and restated as a bound in results.md.

## (2) The rev2 wrapper probe against rev5 — refused at both sites

Reviewer-owned throwaway tests in `/tmp/3dmjbc-rev5-probe/internal/scriptworker/zz_probe_test.go`
(never in the worktree), `go test -count=1 -v -run TestZZProbe ./internal/scriptworker/` → `ok`, 3/3 PASS:

| Probe | Result |
|---|---|
| `TestZZProbeWrapperRefusedAtManagerSite` — `python3-v1` bound to a 0755 `#!/bin/sh` wrapper (`echo wrapper >> marker; exec <stub> "$@"`), driven through `runSession` with an `afterResolve` hook | `script_execution_worker_identity_invalid: the resolved interpreter python3-v1 is not a native interpreter image`; the hook never ran (refused at `ResolveInterpreter`, before any worker process exists); marker absent. **The rev2 defect no longer reproduces.** |
| `TestZZProbeWrapperRefusedAtWorkerSite` — raw hidden-mode worker sent the wrapper path with the wrapper's own correct sha256/size (a lying manager) | worker `failure` frame, same code and detail; marker absent after 500 ms. |
| `TestZZProbeNativeImageStillRuns` — the native stub through `runSession` with `STUB_EXIT=7` | runs, exit status 7 returned, marker written (the gate does not refuse legitimate native images). |

## (3) Shared primitive, go-v1 unchanged

See table row 1. Reruns: `go test -count=1 -v ./internal/godriver/ -run 'TestExecutableIdentity|TestBuildAcceptsAManagerStartedThroughALauncherLink|TestUnixLauncherRequiresExecuteBit|Launcher|NativeExecutable'`
→ 15/15 PASS locally; full `internal/godriver` package `pass` in all five hosted lanes (test+race ubuntu/macos,
test windows). No godriver test file is in the CR.

## (4) Windows evidence (rev5 gate run 35595535426, artifact `test-evidence-windows-latest`)

Extracted from `test/go-test.json`: **51/51 `internal/scriptworker` rows `pass`, 0 skip** (POSIX lanes: 47/47);
present and passing: `TestResolveInterpreterRejectsWrapperImage` + arms `bat-wrapper`, `cmd-wrapper`,
`exe-named-text`, `extensionless-native-image`; `TestScriptWorkerRejectsWrapperImage` + arms `cmd-wrapper`,
`exe-named-text`, `extensionless-native-image`; `TestHasWindowsExecutableExtension`;
`TestScriptWorkerRejectsTamperedInterpreter` (the rev3 Windows failure, now pass);
`TestResolveInterpreterRejectsSubstitution` (no "cannot create a hard link" log line → the hard-link arm
executed on NTFS); `TestVerifyInterpreterDetectsReplacement`. `observed-cases.tsv` lists them,
`skips-observed.tsv` has no scriptworker entry, `ledger/ledger-consistency.txt` = ok (253 rows).
Touched packages (`scriptworker`, `scriptpolicy`, `config`, `godriver`, `install`, `skillcheck`, `cmd/curator`)
`pass` on all five lanes.

## (5) Implemented-controls table still truthful

`scriptpolicy.MandatoryControls` unchanged since rev2: 6 Implemented (fixed-process-graph,
worker-identity-verification, interpreter-resolution-and-identity-verification, operation-private-runtime-area,
explicit-standard-stream-binding, worker-domain-teardown), 5 missing (R2: manager-built-environment,
manager-built-path, offline-network-configuration; R3: inventory-controls-applied,
closed-script-capability-evidence-record). With F1 fixed the `interpreter-resolution-and-identity-verification`
flag is now backed at all four call sites by rows on three OSes; the R-B/R-F qualifications on the stream and
runtime-area flags stand as recorded at rev2. `MissingControls()` is still non-empty, so install/`skill check`
keep refusing `script_execution_control_unavailable` and no production entry launches an enforced script.

## Reruns (macOS x86_64, Go 1.26.0, bash drivers with per-step rc logging, candidate clone at tree `8324fdfc…`)

| Command | Result |
|---|---|
| `go build -o bin/curator ./cmd/curator/` | rc=0 |
| `go vet ./internal/scriptworker/ ./internal/scriptpolicy/ ./internal/config/ ./internal/godriver/` | rc=0 |
| `GOOS=windows go vet ./internal/scriptworker/ ./internal/godriver/` | rc=0 |
| `gofmt -l` over the touched packages | clean |
| `go test -count=1 -v ./internal/scriptworker/` | rc=0, 47/47 PASS (7.2 s) |
| `CURATOR_CONFORMANCE_ROOT=<pin> go test -count=1 -v ./internal/scriptpolicy/` | first attempt `signal: killed` with 0 rows (host exec stall, see host note); retry rc=0, 24/24 PASS incl. 6/6 `opt_in_cases`, 2/2 preflight refusal, 11-control table row |
| `go test -count=1 ./internal/config/ ./internal/skillcheck/` | rc=0 |
| `go test -count=1 -v ./internal/install/ -run 'TestEnforcedScriptCommandIsRefusedAtInstall|TestDeclaredOnlySchema8ScriptCommandInstallsUnchanged|TestActiveScriptCommandsRefusesEnforcedCommand|TestDraftAuditEnforcedScriptRefused|TestDraftLocalEnforcedCommandRefused'` | rc=0, 7/7 |
| godriver identity/launcher mask (above) | rc=0, 15/15 |

## Mutants (mutant clone; python anchor-checked line replacer, `git checkout --` after each, tree `8324fdfc…` at start and end)

| # | Mutant (file:line) | Rows | Result |
|---|---|---|---|
| M-F1a | drop the native-image gate (`interpreter.go:147` → `if false && …`) | `TestResolveInterpreterRejectsWrapperImage`, `TestScriptWorkerRejectsWrapperImage` | **killed**, 5 FAIL lines: `wrapper binding "shebang-wrapper" error = <nil>`, `"plain-text" error = <nil>`, `worker sent "ready", want a failure` |
| N2 (narrowing) | refuse only `#!` images instead of requiring a native image | same + `TestScriptWorkerHappyPath` | **killed** by the `plain-text` arm at `ResolveInterpreter`; the POSIX worker-site arm (shebang only) does not distinguish this narrowing — see bound B1 |
| M-W (narrowing) | gate kept at the manager site, bypassed inside the worker process (`reviewerWorkerBypass` set in `RunWorker`) | same | **killed** by `TestScriptWorkerRejectsWrapperImage/shebang-wrapper` while the resolution row stays green — the worker-site row is independently effective |
| M-F1b | `hasWindowsExecutableExtension` always true | `TestHasWindowsExecutableExtension` | **killed** (6/10 arms) — helper-level; the production-boundary killers are the Windows `extensionless-native-image` arms, hosted-proven above |
| C-tamper (test-side control) | first-byte tamper restored in `TestScriptWorkerRejectsTamperedInterpreter` | that row | **fails as required**: `detail = "… not a native interpreter image", want it to name "interpreter identity proof"` — the detail pin distinguishes the two gates |

Producer's M-F1a/M-F1b reproduce. 4 of 4 narrowing/deleting production mutants killed.

## CLI (rev5 `curator` built from the candidate clone, `env HOME=<tmp> CURATOR_CONFIG=<tmp>/config.json`)

Ran after the host exec-stall window cleared (12:47Z; from ≈12:25Z every fresh Go binary — the CLI, a
`go test` binary — and briefly `/bin/bash` script execs were stranded at `_dyld_start` and SIGKILLed 137 at
~85 s; rev2 verdict R-J shape). All evidence above was produced before the window.

| Invocation | Result |
|---|---|
| `skill check <schema-8 skill, execution_policy script-worker-v1, interpreter python3-v1>` | rc=1, `error: script_execution_control_unavailable (commands.tool.execution_policy): this manager cannot apply mandatory controls: manager-built-environment, manager-built-path, offline-network-configuration, inventory-controls-applied, closed-script-capability-evidence-record; enforced launch is refused until the control set is complete` |
| same, `interpreter node-v1` | rc=1, identical refusal |
| same, `execution_policy script-worker-v9` | rc=1, `skill.manifest_invalid …: commands.tool.execution_policy: must be "script-worker-v1"` (parser refuses before `Admit`, as at rev2) |
| `curator __curator-script-worker-v1 extra` | rc=2, `unknown command` + usage (the mode takes exactly one argv) |
| `curator __curator-script-worker-v1 </dev/null` | rc=3, one `failure` frame `script_execution_worker_protocol_invalid: worker session channel closed`, nothing else |

Unchanged from rev2 (the CLI path bytes are identical); no production entry launches an enforced script.

## Bounds

- B1. The POSIX worker-site row (`TestScriptWorkerRejectsWrapperImage`) has only the shebang arm; the
  non-shebang non-native class at the worker is covered by the resolution row on POSIX and by the Windows
  `exe-named-text` worker arm. Both sites share `readInterpreterIdentity`, so a site-specific narrowing needs
  a deliberate fork (M-W shape) and is caught; a shebang-only narrowing is caught at resolution (N2).
- B2. The header gate reads the first 8 bytes of the same handle that is hashed; the worker's pre-exec
  re-verification (M6, R-E) still has no replacement row — unchanged bound.
- B3. Windows semantics (`lookExtensions`, batch via `cmd.exe`) verified from the Go 1.26 GOROOT sources at
  rev2 and by the hosted windows-latest rows; not executed on this host.
- B4. Residuals R-A…R-J from the rev2 verdict remain for R2/R3/R5 exactly as restated in results.md
  "Revision 3"; R-H (mutable `MandatoryControls` slice) untouched, as permitted.
- B5. rose-air (ARM64) and the candidate suite skipped in the gate — unverified, not passing.

## Verdict routing

`accept_cr(TASK-260916-3dmjbc, revision=5, evidence=TASK-260916-3dmjbc_review-verdict-rev5.md)`.
