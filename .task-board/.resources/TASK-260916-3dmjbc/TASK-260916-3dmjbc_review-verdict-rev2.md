# TASK-260916-3dmjbc review verdict — revision 2 (CR-TASK-260916-3dmjbc-2)

Reviewer: Claude (claude-opus-5), RUN-260921-bd4d65, 2026-09-21. Read-only review; no
edits to the Story worktree. Verdict: **CHANGES REQUESTED → `to-dev`** on one defect (F1)
in the interpreter identity gate this leaf claims implemented; everything else is accepted
as delivered and the remaining items are non-blocking residuals for the orchestrator.

## Exact candidate identity

- Story worktree `.temp/STORY-260822-2h0v9j/worktree` at HEAD `f0a92b8b` + uncommitted delta;
  temp-index `write-tree` = `0b582d53b8af7ef69af4d9e8fd349799b85cfcd3` = CR candidate tree.
- Patch `TASK-260916-3dmjbc_change-request_rev2.patch` sha256 `5d538bb9…a4809fedd45aea`
  (matches the CR). Applied on `f0a92b8b` in three disposable clones (`/tmp/3dmjbc-rev2-{cand,mut,probe}`),
  each committing to tree `0b582d53…`.
- rev1 → rev2 delta reconstructed from the two patches: 4 test files only
  (`internal/config/scriptinterpreters_test.go`, `internal/scriptworker/{client,main,worker}_test.go`),
  45+/11−, 0 production lines — the producer's claim holds.
- Hosted gate run 35579456976: `headSha 88ba9af6` → `^{tree}` = `0b582d53…` (parent `f0a92b8b`),
  all 11 lanes success (Test/Race/Gate self-test ubuntu+macos+windows, Lint, Interop, Naming);
  rose-air and candidate-suite skipped.
- Spec pin `87a0d006` (protocol 1.0.0-rc.9) extracted by `git clone --no-checkout` + `checkout --detach`;
  `conformance/v1/manifest.json` sha256 `0e195ecd…` (the reconciliation's value).

## Reruns (macOS x86_64, Go 1.26.0, bash drivers with per-step rc logging, all on tree `0b582d53…`)

| Command | Result |
|---|---|
| `go build -o …/curator ./cmd/curator/` | rc=0 |
| `go vet ./internal/scriptworker/ ./internal/scriptpolicy/ ./internal/config/` | rc=0 |
| `go test -count=1 -v ./internal/scriptworker/` | rc=0, 27/27 PASS, 7.6 s |
| `go test -count=1 -v ./internal/scriptpolicy/ ./internal/config/ ./internal/skillcheck/` (root unset) | rc=0 (6 conformance rows SKIP, root-unset as the ledger declares) |
| `CURATOR_CONFORMANCE_ROOT=<pin>/conformance/v1 go test -count=1 -v ./internal/scriptpolicy/` | rc=0, 14/14 PASS incl. 6/6 opt-in, 2/2 preflight refusal, 11-control table row |
| `go test -count=1 -v ./internal/install/ -run 'TestEnforcedScriptCommandIsRefusedAtInstall|TestDeclaredOnlySchema8ScriptCommandInstallsUnchanged|TestActiveScriptCommandsRefusesEnforcedCommand|TestDraftAuditEnforcedScriptRefused|TestDraftLocalEnforcedCommandRefused'` | rc=0, 5/5 (+2 subtests), 4.7 s |
| `go test -count=1 -run 'TestExecutableIdentity|TestBuildAcceptsAManagerStartedThroughALauncherLink' ./internal/godriver/` | rc=0, 5 rows (go-v1 identity unchanged; full package accepted from the five hosted lanes) |

Hosted evidence read from the five `{test,race}-evidence-<os>` artifacts (`go-test.json`): every
`internal/scriptworker` row (27) and every touched scriptpolicy/config/install/skillcheck row is
`pass` on ubuntu-latest, macos-latest and **windows-latest** (test + race lanes; Windows 78 rows
incl. the two Windows-only config rejection rows), **zero `skip`** among them — the symlink arms ran
on windows-latest. Package level: `internal/godriver` pass on all five lanes (the producer could
not run it locally), `internal/install`, `cmd/curator` pass. Ledger `observed-cases.tsv` lists the
new rows on all three OSes; `skips-observed.tsv` has none of them.

## What I verified against the spec (§3.1 at the pin) and the brief

1. **Admission honesty (R2)** — `scriptpolicy.Admit`: unknown policy → `script_execution_policy_unsupported`
   (unchanged text); `script-worker-v1` → preflight → `script_execution_control_unavailable` naming the
   five R2/R3 controls in vector order. `MissingControls()` is derived from the 11-row table that
   `TestMandatoryControlsMatchTheSuite` binds to the vector bytes in order and asserts non-empty. No
   env/flag/config/build-tag switch reaches the table (grep over scriptworker+scriptpolicy: only the
   `SYSTEMROOT/WINDIR` bootstrap lookup on Windows). `scriptworker.Launch` re-checks the same table
   before anything starts. Install (`install.Project` v1 + draft lanes), `skill check`, and the shim
   writer (`activeScriptCommands`) all refuse; the shim writer refuses every enforced command
   unconditionally (it wraps `Admit` in a non-nil error even when `Admit` returns nil), so no
   production path can publish an enforced shim. My M8 mutant (table flipped complete) is killed by
   the conformance row, the install row and `TestLaunchRefusesAtPreflight` — see mutant table.
2. **Process graph / order** — manager parent → `exec.CommandContext(identity.Path, "__curator-script-worker-v1")`
   (new session on Unix, private Job Object with kill-on-close on Windows, empty bootstrap env) →
   worker `RunWorker` (wired in `cmd/curator/main.go:113`, exactly one argv) → `exec.Command(request.InterpreterPath, runtimeEntry, args…)`
   with explicit stdin (payload or null device), bounded captured stdout/stderr, verbatim args, exit
   status from `ProcessState`. Three nodes; no `exec` node (R2).
3. **Identity seam reuse** — `godriver/identity.go` diff is +30/−0: four exported wrappers over the
   unchanged `resolveExecutableIdentity`/`matches`/`physicalPath`/`artifactHasMultipleLinks`. Manager
   identity: canonical physical path, regular file, single link, open-then-hash, `Verify()` at the
   launch boundary (`SameFile` + hash + size). Worker self-proof: `ResolveManagerIdentity(os.Executable())`
   → `MatchesExpectation(request path/sha/size)` before anything else; parent re-checks the `ready`
   proof. Interpreter: `ResolveInterpreter(identifier, mapping, forbiddenRoots)` takes ONLY the
   closed identifier and the operator mapping — no PATH/manifest/repo/runtime-root input exists in
   the signature; forbidden roots refuse `.agents/bin`/runtime/snapshot locations even when the
   operator names them; `VerifyInterpreter` at the launch boundary; worker verifies the file at
   accept and again immediately before exec.
4. **Nonce** — 32 random bytes → 64 hex; worker requires exact length + hex; every worker frame is
   compared to the session nonce by the parent (`receive`); the worker binds `shutdown` to it.
5. **Teardown** — Unix: `kill(-pid, SIGKILL)` on the worker's session/group, then `Wait`; Windows: worker
   `Kill` + job close (kill-on-close) — same shape as go-v1 (`terminateWorkerDomain`). Private area
   removed after teardown (defer order verified).
6. **Config seam** — `script_interpreters` is an optional closed object (`node-v1`, `python3-v1` only),
   non-empty string ≤ 4096 chars, `~` expansion, `filepath.IsAbs` (Windows rejects drive-relative and
   rooted-without-volume — rows present), not lockable, follows the existing curator-specific manager
   keys (`execution`, `build_ssh`, `build_https`). Symlink bindings resolve to the physical file
   (Homebrew-style links), hard links / reparse points / directories / relative / absent refuse.
7. **No behaviour change** for declared-only, build and go-v1: godriver additive only, hosted godriver
   green, `TestDeclaredOnlySchema8ScriptCommandInstallsUnchanged` green; CHANGELOG `Unreleased → Added`
   entry and `docs/troubleshooting.md` section present and accurate.
8. **results.md** claims checked: row table matches the executed rows; the M1–M4 kills reproduce
   (below); ratio line 7/12 consumed (6 full + `preflight_cases` partial 2/5), 4 refused-before-reached,
   1 not-implemented matches `scriptHostExecutionPolicySections`; six `opt_in_cases` still `consumedHere`;
   Windows proof is now hosted, not structural (revision 2 gate).

## F1 — required fix: the interpreter identity gate admits wrapper images and, on Windows, executes a file other than the one it hashed

`internal/scriptworker/interpreter.go:95-143` (`readInterpreterIdentity`) accepts any regular,
single-linked, hashed file (`interpreterExecutable` = `mode&0o111` on POSIX, `mode.IsRegular()` on
Windows). The go-v1 primitive this leaf was told to reuse rejects exactly this class:
`godriver/session.go:536-559` `validateLauncher` → `nativeExecutableHeader` ("selected Go launcher is a
wrapper rather than a native executable"). Consequences, all at the worker's single exec site
(`worker.go:267` `exec.Command(request.InterpreterPath, …)`):

- **POSIX (reproduced, probe below):** binding `python3-v1` to a `#!/bin/sh` wrapper resolves,
  hashes, passes both worker verifications and runs — the executed program is `/bin/sh` chosen by the
  wrapper's shebang, then whatever the wrapper execs. The fixed graph gains nodes the manager never
  identity-verified, and the wrapper consults the session environment/PATH to find its target
  (pyenv/asdf/volta shims are exactly this shape and are what an operator is likely to configure).
- **Windows `.cmd`/`.bat` binding (from Go's own source and docs, not run here):** Go executes batch
  files through `cmd.exe` ("cmd.exe (and thus, all batch files)", `os/exec/exec.go:400`), so
  `node.cmd` (npm-style shim) launches `cmd.exe` between the worker and the interpreter and
  re-parses the verbatim arguments with cmd.exe's rules — the profile forbids both
  ("never through … `cmd.exe /c` … a `.cmd` … wrapper"; "forwarding arguments without reinterpretation").
- **Windows extensionless binding (from Go 1.26 `os/exec/lp_windows.go` `findExecutable`, verified
  on this host's GOROOT):** the worker's environment carries no `PATHEXT`, so Go uses
  `.com;.exe;.bat;.cmd`; an absolute path without extension is never executed as itself — Go
  substitutes `<path>.com`/`.exe`/`.bat`/`.cmd` if a sibling exists (else fails). The file that was
  hashed (`C:\tools\node`) and the file that runs (`C:\tools\node.exe`) differ: "per-invocation
  identity verification of the resolved interpreter executable file" is defeated for that shape.
  `cmd.Path` cannot override this (`Start` re-applies `lookExtensions`).

Fix shape (narrow): in `readInterpreterIdentity` (both the manager resolution and the worker's
pre-exec verification share it) require a native image header through the shared godriver
primitive (export `nativeExecutableHeader`/`validateLauncher`-style check, do not fork it) and, on
Windows, require the `.exe` extension so `lookExtensions` returns the path unchanged. Rows: POSIX
shebang wrapper binding → `script_execution_worker_identity_invalid` at `ResolveInterpreter` AND at the
worker (`expectFailure`, marker never written); Windows `.cmd`/`.bat` and extensionless bindings
refused (register on windows-latest like the existing rows); mutant: drop the header check → both rows
fail. Keep the diagnostics table/docs in step (troubleshooting §identity-invalid: "not a native
interpreter image").

Why this blocks rather than trails: this leaf marks `interpreter-resolution-and-identity-verification`
**Implemented** in the preflight table that R2/R3 will complete; nothing later revisits it, and on
Windows the verified file and the executed file can differ. It is unreachable from production today
(preflight refuses every launch), so the fix is small and self-contained.

## Probes (reviewer-owned throwaway tests in `/tmp/3dmjbc-rev2-probe`, untracked, never in the worktree)

Both probes ran green on the candidate tree after the host stall cleared (`go test -count=1 -v -run TestZZProbe ./internal/scriptworker/` rc=0, 3.7 s):

| Probe | What it does | Observed |
|---|---|---|
| `TestZZProbeWrapperInterpreterIsAccepted` | binds `python3-v1` to a 0755 `#!/bin/sh` wrapper (`echo wrapper >> marker; exec <stub> "$@"`), then drives `runSession` (the manager half, real worker process) | `ResolveInterpreter` accepted it (`sha256:b9bc8bd7…`, 257 bytes); the session ran with exit 0; `marker` = `"wrapper\n"`; the stub reported itself as the executable — i.e. `/bin/sh` (chosen by the wrapper's shebang) ran between the verified worker and the interpreter. **F1 reproduced at the process boundary on macOS.** |
| `TestZZProbeWorkerRunsBeforeManagerVerifiesReady` | raw worker (real process), sends the request and never reads `ready` | the interpreter's marker appeared within the 10 s window (0.08 s) — the worker does not wait for a manager go-ahead after `ready` (R-C). |


## Mutants (mutant clone, `git checkout --` between mutants, tree verified `0b582d53…` at the end)

Drivers: `mutants.sh`/`mutants2.sh` (bash, health-gated `for attempt in 1..40; sleep 15`, python line-replacer with anchor + sha check, `go build ./...` before each run, `git checkout --` after, `write-tree` logged). `CURATOR_CONFORMANCE_ROOT` served for M8. Start and end trees `0b582d53…`.

| # | Narrowing mutant (file:line) | Rows run | Result | Kill message |
|---|---|---|---|---|
| M1a | drop the manager launch-boundary recheck `identity.Verify()` only (`client.go:78`; interpreter recheck kept — narrower than the producer's M1) | `TestRunSessionRechecksIdentityAtLaunchBoundary` | **killed** (60.7 s, the 60 s ctx bound as the producer reported) | `replaced-manager error = …worker_protocol_invalid: cannot read the ready message, want …worker_identity_invalid` |
| M2 | `receive` accepts any nonce (`client.go:286` → `if false`) | `TestClientRejectsUnknownNonce`, `TestScriptWorkerHappyPath` | **killed** | `foreign-nonce error = <nil>, want script_execution_worker_protocol_invalid` |
| M3 | PATH fallback when the identifier is unconfigured (`interpreter.go:56`, `os.Stat` over `PATH` dirs, no new imports) | `TestRunSessionRejectsPATHInterpreter`, `TestResolveInterpreterIgnoresUserPATH` | **killed** (2 rows) | `unconfigured launch with PATH decoy = <nil>, want script_execution_control_unavailable` |
| M4 | no process-group kill in teardown (`teardown_unix.go:44` → `_ = pid`) | `TestRunSessionTerminatesDescendants` | **killed** (16.6 s) | `descendant 67518 outlived the worker domain teardown` |
| M5 | worker skips its self-proof against the request expectation (`worker.go:89`) | `TestScriptWorkerRejectsForgedWorkerIdentity`, `TestScriptWorkerRejectsSubstitutedManager` | **killed** (2 rows) | `worker sent "ready", want a failure` (copied binary, modified binary, retargeted link all run) |
| M6 | worker skips the pre-exec interpreter re-verification (`worker.go:113`) | whole `internal/scriptworker` package | **survives** (ok 11.2 s) | no row replaces the file between accept and exec — stated bound R-E |
| M7 | worker accepts any interpreter identifier (`worker.go:160` → `if false`) | `TestScriptWorkerRejectsPackageSelectedPrograms` | **killed** (`shell-interpreter`, `empty-interpreter` arms) | `worker sent "ready", want a failure` |
| M8 | preflight table reports every control implemented (`scriptpolicy.go:86` → `if false`) | scriptpolicy conformance + unit rows, `install.TestEnforcedScriptCommandIsRefusedAtInstall`, `scriptworker.TestLaunchRefusesAtPreflight` | **killed** in 3 packages / 6 top-level rows | `Admit code = "", want "script_execution_control_unavailable"`; `the preflight reports every mandatory control implemented`; install: `did not report script_execution_control_unavailable: enforced-skill.enforced-skill-tool: %!w(<nil>)` (the shim writer still refused — second layer, R-G message shape); `Launch error = …package_influence_forbidden …, want …control_unavailable` |
| M9 | manager drops the `ready` identity-proof check (`client.go:147`) | whole `internal/scriptworker` package | **survives** (ok 10.4 s) | no row has a worker that lies about its identity — stated bound R-D |

Producer's M1–M4 reproduce (M1 in its narrower form). Ratio: 7 of 9 narrowing mutants killed; the two survivors are the second-line rechecks/proof checks named above, not the operative gates.


## CLI (built `curator` from the candidate tree)

All via `env HOME=<tmp> CURATOR_CONFIG=<tmp>/config.json <candidate curator> …` after the stall cleared:

| Invocation | Result |
|---|---|
| `curator skill check <schema-8 skill, execution_policy script-worker-v1, interpreter node-v1>` | rc=1, `error: script_execution_control_unavailable (commands.tool.execution_policy): this manager cannot apply mandatory controls: manager-built-environment, manager-built-path, offline-network-configuration, inventory-controls-applied, closed-script-capability-evidence-record; enforced launch is refused until the control set is complete` |
| same with `interpreter python3-v1` | rc=1, identical refusal |
| same with `execution_policy script-worker-v9` | rc=1, `skill.manifest_invalid …: commands.tool.execution_policy: must be "script-worker-v1"` — an unknown policy is refused by the parser before `Admit`; `Admit`'s `script_execution_policy_unsupported` arm is defence-in-depth reachable only with a hand-built command (its rows are helper-level by construction) |
| `curator __curator-script-worker-v1 extra` | rc=2, `unknown command` + usage (the mode takes exactly one argv) |
| `curator __curator-script-worker-v1 </dev/null` | rc=3, one `failure` frame `script_execution_worker_protocol_invalid` on stdout, nothing else |
| `printf GARBAGE \| curator __curator-script-worker-v1` | rc=3, one `failure` frame `script_execution_worker_protocol_invalid` |

No CLI surface launches an enforced script: there is no launch subcommand at all (R-A), install refuses at preflight, the shim writer refuses unconditionally.


## Non-blocking residuals for the orchestrator (R2/R3/R5)

R-A. **`scriptworker.Launch` has no production caller.** Only `RunWorker` (hidden mode) is wired
  (`cmd/curator/main.go:113`). `Launch`/`runSession`, the `config.ScriptInterpreters` → `LaunchRequest.Interpreters`
  wiring and the forbidden-root list are exercised by tests only, so the vector's
  `mandatory-control-unavailable-at-invocation` case is driven at `Launch` (helper level), not at a
  `cmd/curator` entry; the brief's "installed enforced shims route through this path" (R1 item 4) is
  vacuous until a launcher exists. The producer states this in results.md §Design ("native
  enforced-shim launcher … deferred") while the headline says "R1 complete" — read it as complete
  minus item 4. Note for the launcher design: profile §3 says an enforced command has "no shell,
  `.cmd`, or symlink shim" between the manager and the interpreter and none of §3's launcher forms
  apply, so the R2/R3 shim must be a native launcher.
R-B. **Stream model.** The worker's stdout is the protocol channel, so interpreter stdout/stderr are
  captured into a shared 16 MiB budget and returned in the `result` frame (overflow → `script_execution_worker_protocol_invalid`),
  and stdin is a byte payload copied at request time, not a live pipe/terminal. §3.1 permits a bound
  only on output the manager itself captures and requires transparent forwarding for real invocations
  ("a script command legitimately reads a pipe or a terminal"); the launcher of R-A will need
  pass-through binding (extra inherited descriptors/handles or a side channel), i.e. a protocol change.
R-C. **No permit step.** The worker sends `ready` and starts the interpreter immediately (probe 2); the
  parent's proof check (`client.go:147-152`) and, in R3, evidence validation run concurrently with the
  interpreter, with teardown as the remedy. go-v1 gates the build behind `permit` after validating the
  proof and the evidence record. Adding a permit frame before `runInterpreter` (worker waits for it)
  makes R3's evidence check a pre-run gate as §3.1 step 6→7 reads.
R-D. **Parent-side proof checks have no process-boundary row.** M9 (drop `identity.MatchesExpectation`
  in `runSession`) survives every committed row: forged identity is caught by the worker's self-check
  (`worker.go:89`), which every negative row exercises; a lying worker (test binary in a "forge" mode)
  would be the row shape. Same for `result.Started != 1` and `interpreter.matchesExpectation` on `ready`.
R-E. **Worker pre-exec re-verification** (`worker.go:113`, M6) has no row that replaces the interpreter
  between accept and exec (no worker-side hook); it is a race-narrowing recheck, stated bound.
R-F. **`operation-private-runtime-area` = roots created and torn down only.** `PrivateTmp/Config/Cache`
  are validated by the worker but not applied (no `TMPDIR/TEMP/TMP/XDG_*` binding — R2's environment
  builder); `WorkingDir` is caller-selected. The "Implemented" flag is truthful for creation/teardown,
  not for application.
R-G. `activeScriptCommands` (`install/targets.go:304-307`) builds its refusal as `fmt.Errorf("%w", Admit(...))`;
  once R2/R3 complete the table `Admit` returns nil and the message degrades to `%!w(<nil>)` — the
  enforced-shim writer must replace this guard, not fall through it.
R-H. `scriptpolicy.MandatoryControls` is an exported mutable slice (go-v1 exposes a copying function).
  No writer exists; hardening only.
R-I. Docs: `script_interpreters` is documented only in troubleshooting; `build_https` has a
  reference page — add a config reference entry when the launcher lands.
R-J. Host note: unsigned fresh binaries were SIGKILLed (137) in an exec-stall window ≈09:27–09:44Z (fresh binaries stranded ~85 s at `_dyld_start`, RSS 12 KB, SIGKILL 137; load 3–6); all
  reruns above completed after the window via health-gated drivers; hosted lanes are the arbiter.

## Verdict routing

`set_status(TASK-260916-3dmjbc, status=to-dev)` with this file as evidence. Rework scope: F1 only
(interpreter native-image + Windows `.exe` requirement, rows on all three OSes, one mutant), append
"Revision 3" to results.md, republish; R-A…R-J are for the orchestrator's residual list, not for this
rework unless the orchestrator says so.
