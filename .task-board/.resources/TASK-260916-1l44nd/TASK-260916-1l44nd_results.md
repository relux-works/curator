# TASK-260916-1l44nd results (R3 native probes, evidence record, preflight) — rev 3 (rework 2, run 3)

Worktree: `.temp/STORY-260822-2h0v9j/worktree`, branch `task-board/story/STORY-260822-2h0v9j`.
Spec pin: curator-spec `87a0d0060bad64ab883d007dcdf35df7485368bf` (`profiles/manager.md` §3.1,
vector `script-host-execution-policy.json`). Shell: bash. Role: developer, run 2 (continuation).
Run 1 (RUN-260921-70335c) implemented the slice and timed out before handoff; run 2 verified
it, fixed 2 stale draft tests, killed 6 narrowing mutants, strengthened the 12 mutation rows
with gate-detail pins, and added a production dispatch test. Status: COMPLETE — gate green
locally; hosted lanes arbitrate Linux/Windows.

## Design

- Probe design per platform (`internal/scriptworker/inventory*.go`): one
  per-invocation pre-worker-launch probe per inventory control, no labels/cache/
  config substitution. `available` must probe true or the invocation/install
  refuses `script_execution_control_unavailable`; `host-conditional` records
  present/absent and never refuses; fixed-`unavailable` is never probed and
  never rejects.
  - unix/macOS+Linux: termination = pgid+signal-0; fsize = exact-bound
    set/restore; handles = cloexec pipe test.
  - Linux host-conditional: cgroup v2 (transient child cgroup, enable
    controllers, write pids.max/memory.max, remove), Landlock (create ruleset
    + add rules on O_PATH fds, no self-restrict in parent), netns (self-reexec
    probe child with Unshareflags user+net).
  - Windows: job-object limits (kill-on-close, active-process, memory) via
    exact-limit job creation; handles via attribute-list allocation.
- Applied-control mapping: parent installs job/cgroup/rlimit mechanisms
  pre-launch; worker confirms session leadership, rlimit value, cgroup
  membership, handle hygiene, and applies Landlock (worker self-restriction
  before the readiness proof) + netns (interpreter child Unshareflags).
  Exactly the probe-available/present set is applied.
- Evidence record schema: `script-capability-evidence-v1`
  `{record_version, execution_policy, platform, controls[{name, availability,
  status, probed_at}]}`; worker returns it with the nonce acknowledgement
  (`ready`); parent validates against its own probes BEFORE `permit`
  (`validateScriptEvidence` per-rule checks plus the `scriptEvidenceEqual`
  identity comparison as backstop). `*_invalid` for shape/probe
  contradictions, `*_hardened_claim_forbidden` for foreign policy/version or
  any of the 13 deferred guarantees (7 script + 6 build). Result-only: never
  stdout/stderr/markers/receipts/cache/claims; surfaced via `Result.Evidence`
  + operator `script_diagnostics_dir`.
- R2 seam (`OverrideMandatoryControlsForTest`) removed; table complete (11/11).
- R-e stream model: kept bounded (64 MiB stdin cap, 16 MiB capture); pass-through
  deferred — bound stated in docs (R5 decides).

## Row table (production boundary per row)

Vector `script-host-execution-policy.json` at CI pin 87a0d006. Local lane: darwin/amd64,
Go 1.26.0. `go build ./...` exit 0; `go vet` on scriptworker/scriptpolicy/install/
skillcheck/config/cmd/curator exit 0; `GOOS=windows`/`GOOS=linux` build + vet exit 0;
`ledger-consistency.sh` ok (307 rows).

### Evidence cases (14/14 driven, all PASS)

| Vector case | Production entry test |
|---|---|
| valid-linux-host-conditional-unavailable | scriptworker `TestLinuxHostConditionalUnavailableSucceeds` (runSession, Linux-only; SKIP on darwin via `platform-control`) + shape row `TestScriptEvidenceValidRecordSucceeds` (host inventory) |
| available-control-reported-unavailable | `TestScriptEvidenceMutationsRefuseBeforePermit/available-control-reported-unavailable` (runSession + forge worker; refuses before permit naming "reports status", no `.forge-permit`) |
| unavailable-control-reported-applied | `.../unavailable-control-reported-applied` (macOS inventory forced on unix; "reports status") |
| missing-control-entry | `.../missing-control-entry` ("missing exactly one entry") |
| duplicate-control-entry | `.../duplicate-control-entry` ("duplicates control") |
| extra-control-entry | `.../extra-control-entry` ("outside script-worker-v1-native-control-inventory-v1") |
| unknown-record-version | `.../unknown-record-version` ("record_version") |
| host-conditional-status-contradicts-probe | `.../host-conditional-status-contradicts-probe` (Linux inventory forced; skips on Windows per ledger; "reports status") |
| cached-probe-result | `.../cached-probe-result` ("probed_at") |
| second-record-for-invocation | `TestScriptEvidenceSecondRecordRefuses` (permit earned, second record in result refuses naming "second") |
| foreign-build-record-version | `.../foreign-build-record-version` ("record_version") |
| foreign-build-execution-policy | `.../foreign-build-execution-policy` (`*_hardened_claim_forbidden`, "is not the enforced script identity") |
| deferred-script-guarantee-entry | `.../deferred-script-guarantee-entry` ("deferred guarantee") |
| deferred-build-guarantee-entry | `.../deferred-build-guarantee-entry` ("deferred guarantee") |

Each mutation row pins its own validation rule by refusal detail (run 2 addition), so no
row can pass on the identity-comparison backstop. `TestScriptEvidenceOrderMismatchRefuses`
pins the backstop itself ("contradicts this invocation's probe"). Vector-shape pins:
scriptpolicy `TestPreflightRefusalCases` + `TestScriptHostExecutionPolicySectionsAreAllClassified`
(skip without `CURATOR_CONFORMANCE_ROOT`; CI serves the pin).

### Preflight cases (5/5 driven, all PASS)

| Vector case | Production entry test |
|---|---|
| mandatory-control-unavailable-at-install | install `TestEnforcedScriptCommandIsRefusedAtInstall` (install.Project, node-v1+python3-v1; reports `script_execution_control_unavailable`, names the control, publishes no shim) + `TestActiveEnforcedScriptCommandsRefusesUnavailableHost` (writer guard) + draft-lane `TestDraftLocalEnforcedCommandRefused` phase 1 |
| mandatory-control-unavailable-at-invocation | scriptworker `TestPreflightRefusesUnavailableControlAtInvocation` (`Launch` entry; refuses, starts no worker — no forge markers) |
| linux-pids-max-probe-available-evidence-applied-invocation-succeeds | `TestLinuxPidsMaxProbeAvailableApplies` (runSession, delegated-cgroup fixture, unix lanes) |
| linux-pids-max-probe-unavailable-evidence-unavailable-invocation-succeeds | `TestLinuxPidsMaxProbeUnavailableSucceeds` (fixture without pids delegation) |
| fixed-unavailable-control-does-not-reject | `TestFixedUnavailableControlDoesNotReject` (runSession; Linux drives macOS case via override) |

### Admission / opt-in / launch (R1 completion, all PASS)

- scriptpolicy `TestEnforcedShapesProceedPastAdmission`, `TestMandatoryControlsMatchTheSuite`,
  `TestPreflightTableCompleteAdmitsEnforced`, `TestPreflightTableAccessorIsACopy`,
  `TestHostControlErrorNamesUnavailableControls`: table 11/11, no owners, seam removed.
- install `TestScriptOptInCasesAtInstallEntry` (6 opt-in shapes at install.Project),
  `TestEnforcedInstallAndLaunchAtCLIEntry` (built binary argv dispatch + installed launcher runs),
  `TestEnforcedScriptCommandInstallsNativeLauncher`,
  `TestGlobalEnforcedInstallExcludesForwardingMirror`,
  `TestEnforcedToDeclaredOnlyFlipRemovesNativeLauncher`,
  `TestDraftLocalEnforcedCommandRefused` (two-phase: fault→refusal+no side effects;
  healthy host→ok+native launcher sidecar),
  `TestDraftAuditEnforcedScriptAdmitted` (renamed; dry-run gate admits, no script-policy refusal).
- scriptworker `TestProductionBinaryLaunchesWhenHostProvides`, `TestLaunchDrivesDerivation`,
  `TestNativeLauncherWorkerModeWinsOverShimDispatch`, `TestLoadShimSidecarRejectsMalformed`,
  `TestRunShimWritesDiagnosticsRecord` (R-f), `TestRunShimToleratesUnusableDiagnosticsDir`,
  `TestHostProbeReportsClosedInventory`, `TestFixedUnavailableControlsAreNeverProbed`,
  `TestWorkerRejectsMalformedInventory`, `TestDerivedPathSetShapes` (R-a),
  `TestWindowsJobLimitsAppliedAndConfirmed` (windows-latest; SKIP on darwin),
  `TestLinuxLandlockConfinementMatchesProbe`, `TestLinuxHostProbeAndEvidenceAreConsistent`
  (ubuntu-latest; SKIP on darwin).
- cmd/curator `TestProductionBinaryDispatchesNetNSProbe` (run 2; built binary dispatches
  the hidden probe mode before launcher dispatch, exits 0; ledger-registered).
- Full-package runs: scriptworker ok (19.7s), scriptpolicy ok, skillcheck ok, config ok,
  install `TestDraft*` lane ok (484s).

### Incidents fixed in run 2

1. Four worker tests (`TestWorkerConfirmsFileSizeBound`, `TestWorkerRejectsWrongFileSizeBound`,
   `TestScriptWorkerRejectsReplayedNonce`, `TestScriptWorkerModeIsNotReachableThroughPackageData`)
   failed once (~76s hang, "worker session channel closed") during a host-thrash window that
   also SIGKILLed `go tool compile` repeatedly. All pass in isolation and in every full-package
   rerun since. Environmental flake, not a regression; no code change.
2. `TestDraftLocalEnforcedCommandRefused` + `TestDraftAuditEnforcedScriptRefused` failed
   deterministically: they encoded the pre-R3 universal enforced refusal. R3's binding R1
   ruling (enforced proceeds when the host provides controls) ends that contract, and no
   draft-specific guard exists in production (the draft lane shares staging, which publishes
   contained native launchers with the host preflight). Contract change surfaced: runtime
   test is now two-phase (host-conditional refusal + proceed path); audit test renamed to
   `TestDraftAuditEnforcedScriptAdmitted` (dry run stops before staging, so it never probes).
   Precedent: run 1 converted `TestEnforcedScriptCommandIsRefusedAtInstall` and the
   skillcheck enforced test the same way.

## Mutant table (6 narrowing mutants, all killed)

Production call sites in `internal/scriptworker/inventory.go` (`validateScriptEvidence`,
`probeScriptInventory`, `scriptEvidenceEqual`) and `internal/scriptpolicy/scriptpolicy.go`.

| Mutant | Narrowing | Killing row | Outcome |
|---|---|---|---|
| M1 order-insensitive `scriptEvidenceEqual` | identity comparison ignores entry order | `TestScriptEvidenceOrderMismatchRefuses` | KILLED: reordered record permitted, failed downstream with worker-protocol-invalid |
| M2 accept `probed_at: at-install` | timing check admits cached generation | `.../cached-probe-result` | SURVIVED code-only (identity backstop refused instead) → rows strengthened with detail pins → KILLED ("contradicts…", want "probed_at") |
| M3 probe fixed-unavailable controls | skip-branch condition never matches | `TestFixedUnavailableControlsAreNeverProbed` | KILLED: injected fault fired, preflight refused control-unavailable |
| M4 accept unavailable status for available control | status check admits contradiction | `.../available-control-reported-unavailable` | KILLED via detail pin (backstop would have masked it) |
| M5 fail open on probe error | probe error recorded as absent host-conditional | `TestPreflightRefusesUnavailableControlAtInvocation` | KILLED: Launch proceeded past preflight (no control-unavailable refusal) |
| M6 `inventory-controls-applied` unimplemented | table narrowed to 10/11 | `TestPreflightTableCompleteAdmitsEnforced` | KILLED: "not implemented; the R3 table is complete" |

All mutants reverted; tree verified mutant-free (`grep MUTANT` empty) and gofmt-clean;
targeted rows re-run green after revert.

## Ratio line (vector keys consumed)

Derivation 4/4 (kept), mandatory 11/11, evidence 14/14, preflight 5/5, opt-in 6/6
(kept + install/CLI entries), audit 0/4 (R4 owns). Sections: 11/12 consumed
(`audit_label_cases` = `not implemented: audit warning classes`, R4); the six
`unreachable` classifications are replaced by production-entry consumers.

## Windows proof status

UNVERIFIED locally (darwin host). `GOOS=windows go build ./...` + `go vet` (incl. test
files) exit 0. Windows job-object rows committed for windows-latest (native evidence
suite + `TestWindowsJobLimitsAppliedAndConfirmed`); hosted lanes arbitrate.

## Bounds

- R-e pass-through streaming deferred (docs bound).
- Linux true-probe rows prove on ubuntu-latest only; fixture-backed Linux rows run on
  unix lanes. Linux-only code reviewed statically (Landlock packed-struct layout,
  netns uid-mapping, cgroup fixture protocol); `GOOS=linux` build + vet exit 0.
- Full 305-test install package not re-run end to end (exceeds one bounded call);
  blast radius covered instead: production install diff is strictly enforced-gated
  (one host-preflight call), and every enforced-touching install test passes
  (9 script rows + full `TestDraft*` lane incl. the 2 converted tests).
- Campaign rules prohibit LOGBOOK.md edits: no logbook record; findings live here.
- Lint: `go vet` clean on all touched packages (darwin; scriptworker/install also vetted
  for windows/linux); gofmt clean. Local golangci-lint (misspell/revive/gosec) cannot run
  on this host — the binary is SIGKILLed (exit 137, host OOM) even for `--version` and
  single-package runs; environmental, not a code verdict. Mitigation: gosec is excluded
  for `_test.go` by config and every run-2 edit is tests/TSV/docs (production diff
  unchanged since run 1, which carries narrow named `#nosec` annotations); CI lint lane
  arbitrates.

## Undone

1. Hosted ubuntu-latest / windows-latest evidence (CI lanes arbitrate; cannot run here).

---

# Revision 2 (rework 1, COMPLETE — awaiting rev2 gate)

Gate rev1 FAILED (run 35667106874). Rework-1 named Race-ubuntu + Lint; the
run's own evidence shows Test-ubuntu and Test-windows failed too. All four
are addressed here; no new scope. Shell: bash. Local lane: darwin/amd64,
Go 1.26.0. Previous sections unchanged (run-2 evidence kept).

## Rev2 cause 1 — Linux Landlock `/dev/null` EINVAL (Race + Test ubuntu)

Ubuntu evidence (`test-evidence-ubuntu-latest/go-test.json`) shows the single
root cause in every failing session:

    script_execution_capability_evidence_invalid:
    cannot add the Landlock rule for "/dev/null"

The worker granted filesystem-write-confinement over `/dev/null` (a
character device) with write+truncate rights; the kernel returns EINVAL for
a path rule carrying rights the object type cannot carry. Failed rows: all
session-driving rows in both ubuntu lanes (derivation, evidence, pids-max,
host-conditional, shim/diagnostics, install launcher + CLI entry — 44
`FAIL required` lines across Test and Race ubuntu, same row set in both).

Fix (production, `internal/scriptworker`):
- `worker.go`: the session pre-binds the null-device handle in `accept()`,
  BEFORE `applyScriptControls` (Landlock never restricts an already-open
  descriptor); `runInterpreter` uses the bound handle and refuses
  fail-closed if it was not bound. Single construction site (`RunWorker`
  → `accept` → `serveRun`), so the handle is always present.
- `apply.go`: `/dev/null` removed from the `writables` grant set — the null
  standard input carries no rule at all.
- `landlock_linux.go` `addGrant`: per-object-type rights — the granted set
  is `access & handled`, and non-directories (files, devices) additionally
  shed the truncation right, which applies to directory hierarchies. A
  regular file in the derived path set therefore grants write-only; a
  derived-path file opened O_TRUNC is denied (bounded, fail-closed).
- Comments updated (`landlock_linux.go` header, `inventory.go`
  `landlockGrants`): null input binds pre-restriction, no rule.

Lane-independence question (rework-1 §1): the premise "the Test lane passed"
is contradicted by the run evidence — Test-ubuntu failed the SAME rows as
Race-ubuntu (identical Landlock refusal). The rows are already
lane-independent; both lanes failed on the one cause above, and both are
repaired by it. Probe/apply proof on the ubuntu runner comes from the rev2
gate (no Linux host here; `GOOS=linux` build + vet exit 0; no container
runtime usable — see Bounds).

## Rev2 cause 2 — Lint (5 issues, all in `internal/scriptworker`)

CI Lint (golangci-lint v2.12.2 via action v7) reported exactly:
- `capabilities.go:334` QF1002 untagged switch → tagged
  `switch filesystem.Keyword`.
- `inventory_unix.go:50` revive unused `platform` → `_` (same for
  `inventory_windows.go`, which the Linux lint run does not compile but
  shares the defect; `inventory_other.go` already used `_`).
- `landlock_linux.go:115,125` gosec G103 → `// #nosec G103 -- raw syscall
  args for landlock_*; no memory escapes` audit comments (repo style).
- `worker_test.go:726` revive unused `request` → `_`.
Local `go vet` (darwin + `GOOS=windows` + `GOOS=linux` on touched packages)
and `gofmt` are clean; golangci-lint itself cannot run on this host (see
Bounds, unchanged from run 2).

## Rev2 cause 3 — Windows enforced-launcher naming (2 tests, 1 test-side bug)

Windows evidence (`test-evidence-windows-latest/go-test.json`):
- `TestEnforcedInstallAndLaunchAtCLIEntry`: `scriptpolicy_test.go:675` —
  `...\.agents\bin\cli-enforced-tool.cmd.curator-shim.json: The system
  cannot find the file specified.`
- `TestScriptOptInCasesAtInstallEntry/schema8-explicit-opt-in`:
  `scriptpolicy_test.go:577` — `...optin-schema8-explicit-tool.cmd: The
  system cannot find the file specified.`

Root cause (test-side, not production): both tests built the enforced
launcher path with `shimName()` (ordinary-shim naming: `.cmd` on Windows),
but production publishes enforced native launchers as `<cmd>.exe` +
`<cmd>.exe.curator-shim.json` on Windows and never a `.cmd` wrapper
(`runtimestore.NewManagedEnforcedShim`; proven by the passing
`TestEnforcedScriptCommandInstallsNativeLauncher`, which also asserts the
`.cmd` absence). The CLI install itself succeeded on Windows — only the
sidecar-name expectation was wrong.

Fix (test-only, `internal/install/scriptpolicy_test.go`):
- New `nativeLauncherName()` helper (`.exe` on Windows, plain elsewhere).
- Opt-in `wantLauncher` branch and the CLI-entry test use it (unix paths
  byte-identical to before).
- Refusal test additionally asserts the native `.exe` + sidecar are absent
  on Windows (the old `.cmd`-only absence check was vacuous there).
No production naming change; no test renamed, so ledger rows are unaffected.

Note: the platform-case gate printed `ok` for
`TestScriptOptInCasesAtInstallEntry` while `go-test.json`/`observed-cases.tsv`
record `fail` for it (parent and subtest). The lane still failed correctly
via `go test exit=1`. Gate-display quirk only; out of scope, recorded here.

## Rev2 local verification (reran myself, exit codes real)

- `go build ./...` → 0; `GOOS=windows go build ./...` → 0;
  `GOOS=linux go build ./internal/scriptworker/` → 0.
- `go vet` darwin on scriptworker/scriptpolicy/install → 0;
  `GOOS=windows go vet` on scriptworker/install (compiles the Windows test
  branches) → 0; `GOOS=linux go vet` on scriptworker → 0.
- `gofmt -l` on touched dirs → empty.
- `go test ./internal/scriptworker/ ./internal/scriptpolicy/ -count=1` →
  ok (scriptworker 17.9s, scriptpolicy 1.0s). Covers the worker pre-bind
  change across every session test incl. `TestScriptWorkerBindsStandardInput`.
- `go test ./internal/install/ -run 'TestScriptOptInCasesAtInstallEntry|
  TestEnforcedInstallAndLaunchAtCLIEntry|
  TestEnforcedScriptCommandIsRefusedAtInstall|
  TestEnforcedScriptCommandInstallsNativeLauncher|
  TestActiveEnforcedScriptCommandsRefusesUnavailableHost'` → all PASS
  (196s). No `TestDraft*` or other install tests touched by this rework.
- Not rerun: full install package (unchanged rationale from run 2);
  hosted ubuntu/windows lanes (rev2 gate arbitrates).

## Rev2 undone

1. Hosted ubuntu-latest / windows-latest rev2 evidence (CI lanes arbitrate;
   cannot run here). If the rev2 gate finds a failure PAST the fixed
   Windows sidecar lookup (e.g. launcher stdout shape on Windows), that is
   new information, not a known gap.
2. No new narrowing mutants for the rework diff (rework-1: no new scope);
   run-2 mutant table (M1–M6, all killed) stands for the gates.

---

# Revision 3 (rework 2, COMPLETE — awaiting rev3 gate)

Gate rev2 FAILED on BOTH ubuntu lanes (run 35671704019; 33 rows, one
root cause). Lint green. Previous sections unchanged (run-2 + rev2
evidence kept). Shell: bash. Local lane: darwin/arm64, Go 1.26.0.

## Rev3 cause — `landlock_restrict_self` EPERM in the worker

Ubuntu evidence: every session-driving row refused with

    script_execution_capability_evidence_invalid:
    cannot enforce the Landlock ruleset

i.e. `landlock_restrict_self` returned EPERM. Three compounding defects
in `internal/scriptworker/landlock_linux.go` (rev2):

1. The worker never set `no_new_privs`, which `landlock_restrict_self`
   requires of the calling thread (absent CAP_SYS_ADMIN, which the
   runner does not grant).
2. The restriction ran in the multi-threaded worker process at pre-ready
   time: Landlock restricts the calling THREAD only, so even with the
   privilege bit the domain would have covered one arbitrary thread
   rather than the interpreter child.
3. The failure was classified as `capability_evidence_invalid` by the
   worker itself, although an apply error is the probe/apply outcome of
   ONE control, not a probe/evidence contradiction (which only the
   parent's evidence gate can verdict).

## Rev3 fix — spawn-time enforcement on a locked thread

Production (`internal/scriptworker/landlock.go` NEW shared,
`landlock_linux.go` rewritten, `landlock_other.go` stubs,
`worker.go`, `apply.go`, `cmd/curator/main.go`,
`main_test.go` dispatch):

- Enforcement moved out of the pre-ready apply into `runInterpreter`,
  on a locked OS thread, immediately before the interpreter spawn, from
  that same thread: `runtime.LockOSThread` (held until the
  single-session worker exits — returning a restricted thread to the
  pool would sandbox unrelated goroutines) → `prctl(PR_SET_NO_NEW_PRIVS,
  1)` → `landlock_create_ruleset` with the ABI-gated handled mask →
  one typed `landlock_add_rule` per grant → `landlock_restrict_self` →
  fork/exec from the locked thread, so the child inherits the domain.
  Post-restriction the worker thread only moves pipe bytes, waits, and
  writes the result frame (all on already-open descriptors, unrestricted
  under Landlock); reads were never handled, so the Go runtime, the
  dynamic linker, and the netns `/proc` confirmation are unaffected.
- ABI mask (`landlockHandledForABI`, shared + pure): execute/write need
  ABI 1, truncation needs ABI 3, below ABI 1 handles nothing, and newer
  ABIs never adopt refer/network/ioctl — the enforcement never requests
  a right the probed ABI lacks (a foreign right fails ruleset creation
  with EINVAL).
- Rule typing (`landlockRuleRights`, shared + pure): each rule carries
  only handled rights valid for its object type — truncation only on
  directories, files/devices at most write (+execute for the interpreter
  file); stdio and the null device stay pre-bound with no rule.
- Pre-ready apply keeps a construction-only confirmation
  (`confirmScriptLandlock`: same ABI mask, same typed rules over the
  same grant paths derived from one helper, ruleset discarded, never
  restricted). One atomic ruleset ⇒ all-or-nothing: a control whose
  enforcement cannot be constructed is reported `unavailable` in the
  worker's record — the probe/apply outcome of that one control — and
  the parent refuses the resulting genuine probe contradiction before
  the permit. The worker never manufactures `capability_evidence_invalid`
  for an apply error. A spawn-time enforcement failure (after the ready
  record claimed applied) refuses fail-closed with
  `script_execution_worker_protocol_invalid` and the interpreter never
  starts — the same code as "cannot start the verified interpreter".
- Probe (`probeScriptLandlock`): ABI query, then a self-reexec child in
  the new hidden mode `__curator-script-landlock-probe` (dispatched in
  `cmd/curator/main.go` and the test binary, before launcher dispatch)
  performing the FULL sequence — lock, no_new_privs, ruleset, rule over
  a manager-owned directory, restrict — and exiting 0/1. The manager and
  test binaries never restrict themselves (Landlock stacks at most 16
  layers per thread; a self-restriction would confine the tester).
- No test, and no manager path, calls restrict/enforce in-process:
  unit tests cover construction only; enforcement rows drive the built
  worker binary at the process boundary.

## Rev3 tests (all linux-gated rows ubuntu-arbitrated)

New, ledger-registered (`platform-cases.tsv` 313 rows, +6):

- `TestLandlockHandledMaskFollowsABI` (all platforms, pure): ABI 0–6 ×
  control combos. PASS on darwin.
- `TestLandlockRuleRightsFollowObjectType` (all platforms, pure): 7
  typing cases incl. the rev1 shape (truncate on a file). PASS on darwin.
- `TestLandlockConstructionConfirmsTypedRules` (linux-required): grants
  over existing paths confirm (incl. per-control subsets), a missing
  grant confirms neither, and a canary write outside the grants succeeds
  afterwards — proving construction never restricted the caller.
  SKIP on darwin via `platform-control`.
- `TestLinuxLandlockProbeFindsPresentControls` (linux-required): the
  true probe finds BOTH Landlock controls present on the hosted runner
  (kernel 6.8+, ABI 4). Without it a silent absent would disarm every
  enforcement row. SKIP on darwin via `platform-control`.
- `TestLinuxLandlockProbeChildAnswersAtTheBoundary` (linux-required):
  the re-executed child exits 0 over an owned directory, nonzero over a
  missing path; both branches in the child, never in the test binary.
  SKIP on darwin via `platform-control`.
- `TestProductionBinaryDispatchesLandlockProbe` (cmd/curator, all
  platforms): the prod binary dispatches the mode before launcher
  dispatch — exit 0 on Linux, the absent exit elsewhere, silent either
  way, mode absent from `usage`. PASS on darwin (absent branch).

Strengthened: `TestLinuxLandlockConfinementMatchesProbe` now pins
EACCES (`permission denied`) for the outside write and the outside
spawn when confined — any other failure would name a different
mechanism. Both branches kept, so the row stays green on
Landlock-less Linux hosts and passes identically under `-race` (no
timing dependence: stub writes files, reports errors).

Fixed for first enforcement: `TestRunSessionTerminatesDescendants`
published the descendant pidfile in the system temp dir — outside the
derived grants, so the write is EACCES once Landlock actually enforces
(rev3 is the first revision that reaches enforcement on ubuntu). The
test now declares a `repo` filesystem over the fixture root and
publishes the pidfile inside it (granted); the spawned program is the
granted interpreter itself. Same test name, same ledger row, behavior
identical off Linux. Audited every other stub side-effect vector:
raw-worker marker tests never claim Landlock (no enforcement);
LookPath is reads-only; install/draft/CLI tests bind the stub with no
`STUB_*` filesystem/exec vectors; the forge binary never spawns an
interpreter; diagnostics are written by the unconfined parent.

## Rev3 mutants (narrowing; N1–N2 killed here, N3–N5 ubuntu-arbitrated)

| Mutant | Narrowing | Killing row | Outcome |
|---|---|---|---|
| N1 truncation without the ABI≥3 gate (`landlock.go`) | handled mask over-requests on ABI 1–2 (EINVAL on old kernels) | `TestLandlockHandledMaskFollowsABI/abi-{1,2}-without-truncate` | KILLED locally: `handled mask = 0x1003, want 0x3` |
| N2 no per-object-type narrowing (`landlock.go`) | file grants keep truncation (the rev1 EINVAL shape) | `TestLandlockRuleRightsFollowObjectType/file-sheds-truncate` | KILLED locally: `rule rights = 0x1002, want 0x2` |
| N3 probe child skips no_new_privs (`landlock_linux.go`) | probe proves ruleset+rule but not the privilege precondition (the rev2 EPERM shape) | `TestLinuxLandlockProbeFindsPresentControls` | LINUX-COMPILE-VERIFIED, darwin suite green (never called there); predicted kill on ubuntu: child restrict → EPERM → exit 1 → absent → row fails. Note the confinement row would stay green via its absent branch — this is why the require-present row exists |
| N4 enforcement is a no-op returning nil (`landlock_linux.go`) | evidence claims applied but no domain is enforced | `TestLinuxLandlockConfinementMatchesProbe` | LINUX-COMPILE-VERIFIED, darwin suite green; predicted kill on ubuntu: outside writes/spawns succeed → EACCES pins fail |
| N5 construction always confirms (`landlock_linux.go`) | bogus grants confirm | `TestLandlockConstructionConfirmsTypedRules` | LINUX-COMPILE-VERIFIED; darwin-neutrality run inconclusive (host thrash, see incident 1); predicted kill on ubuntu: bogus-grant case confirms (true, true) |

Note: N3–N5 live in `landlock_linux.go`, which the darwin build does
not compile; each was verified with `GOOS=linux go build` while
applied. N3/N4 additionally ran the full darwin `scriptworker` suite
green while applied (proving Linux-gating); N5's neutrality run was
cut short by host thrash (incident 1) — neutrality still holds by
build-tag exclusion (the file is not compiled on darwin), and the
post-revert suite re-green covers the tree. The ubuntu lanes arbitrate
the kills. All mutants reverted; tree verified mutant-free (`grep
MUTANT` empty) and gofmt-clean; targeted rows re-run green after
revert.

## Rev3 incidents

1. Host thrash (environmental, run-2 incident-1 recurrence): during the
   N5 neutrality run the suite slowed 20× (a worker child still
   spawning at 10 min; load 5.6, ~172 MB free of 32 GB, a VM at 109%
   CPU plus a concurrent `task-board spawn wait` at 27%), then a smoke
   re-run was SIGKILLed at 78 s with near-zero CPU time. Terminated my
   own stray workers, waited, re-ran: green. Not a code signal — the
   same tests passed minutes earlier, SIGKILL is an external killer
   (Go's own timeout panics instead), and N5 cannot affect the darwin
   binary (build-tag exclusion).

## Rev3 local verification (reran myself, exit codes real)

- `go build ./...`, `GOOS=linux go build ./...`,
  `GOOS=windows go build ./...` → 0.
- `go vet` darwin on scriptworker/scriptpolicy/install/cmd/curator →
  0; `GOOS=linux go vet` on scriptworker → 0; `GOOS=windows go vet`
  on scriptworker + cmd/curator (compiles the Windows test branches)
  → 0. `gofmt -l` on touched dirs → empty.
- `go test ./internal/scriptworker/ ./internal/scriptpolicy/ -count=1`
  → ok (scriptworker ~16s, scriptpolicy ~1s), incl. the reworked
  `TestRunSessionTerminatesDescendants` (also run standalone).
- `go test ./cmd/curator/ -run
  'TestProductionBinaryDispatchesLandlockProbe|
  TestProductionBinaryDispatchesNetNSProbe'` → both PASS.
- `ledger-consistency.sh` → ok, 313 rows checked.
- Not rerun: full install package (untouched by this rework — no
  install/ file changed in rev3); hosted ubuntu/windows lanes (rev3
  gate arbitrates). No container runtime usable here (docker daemon
  unresponsive), so Linux behavior is proven by the ubuntu lanes only.
- Lint: `go vet` clean as above; golangci-lint itself cannot run on
  this host (unchanged environmental bound from run 2: the binary is
  SIGKILLed). Mitigation: the two new `#nosec G204` suppressions are
  narrow, named, and reasoned (self-reexecution in the fixed hidden
  probe mode, mirroring the netns precedent); no new `unsafe` blocks
  (existing G103 comments kept); new switches tagged; no unused
  params; `no-broad-suppression.sh` pattern satisfied by inspection
  (no bare `//nolint`, no unnamed `#nosec`).

## Rev3 undone

1. Hosted ubuntu-latest / windows-latest rev3 evidence (CI lanes
   arbitrate; cannot run here). The rev3 gate is the first run of the
   corrected enforcement on a Landlock host.
2. Mutant kills N3–N5 (ubuntu-arbitrated; predicted above).
3. No narrowing mutant for the spawn-failure `worker_protocol_invalid`
   classification: no seam forces a post-ready enforcement failure
   (the probe child proves the sequence seconds earlier), and adding a
   fault seam to the worker child would itself be new scope.

---

# Revision 4 (rework 3, COMPLETE — awaiting rev4 gate)

Gate rev3 FAILED (run 35676751063): BOTH ubuntu lanes plus Lint. The
rework premise "(Test ubuntu green)" is contradicted by the run evidence
(same miscount as rework-1): Test ubuntu job 106584765891 FAILED the
SAME 2 rows as Race ubuntu (verified from the downloaded
`test-evidence-ubuntu-latest/go-test.json`, 3 `fail` entries:
`TestOperationPrivateRuntimeAreaApplied`,
`.../paths-derive-project-root`, `TestRunSessionTerminatesDescendants`).
All three failures are deterministic and lane-independent. Previous
sections unchanged (run-2 + rev2 + rev3 evidence kept). Shell: bash.
Local lane: darwin/arm64, Go 1.26.0.

## Rev4 cause 1 — missing derived member voids BOTH Landlock controls (both ubuntu lanes)

Row `TestOperationPrivateRuntimeAreaApplied/paths-derive-project-root`
refused with

    script_execution_capability_evidence_invalid: host-conditional control
    "descendant-exec-denial" reports status "unavailable" against this
    invocation's probe

Root cause (deterministic, in the row + the worker): the row declares
`filesystem: ["rel/path"]` over a fresh TempDir where `rel/path` was
never created, so the derived path set contains a member that does not
exist. `landlockAddGrant` opens each grant `O_PATH`; the missing member
fails with ENOENT; `landlockConstructRuleset` was all-or-nothing, so
`confirmScriptLandlock` returned `(false, false)` — one missing
*writable* silently voided the *execute* denial too — and the worker
reported both installable controls `unavailable` while the probe had
found them present. The parent's evidence gate refused the
contradiction, naming the first contradicting entry
(descendant-exec-denial). The sibling `repo-derives-project-root` row
passed because its derived set (the project root itself) exists.

Fix (production, `internal/scriptworker/landlock_linux.go`,
`landlock_other.go`, `apply.go`): the exec-denial decision is now made
once per invocation from the same probe result the record cites.

- `landlockConstructRuleset` collects rule failures PER CONTROL (one
  ruleset — stacked layers would intersect, so enforcement stays a
  single domain — but a grant that cannot be ruled fails only its own
  control; a ruleset-level failure fails every installable control with
  the same cause). Confirm and enforce still build it in the same
  function, so they cannot diverge.
- `confirmScriptLandlock` returns per-control causes
  `(execErr, writeErr error)`; `apply.go` REFUSES fail-closed with
  `script_execution_worker_protocol_invalid` naming the control and the
  cause when an installable control cannot be constructed. No re-probe
  in the worker, no per-path fallback, never a status flip: the spec
  (manager.md at the pin: a host-conditional control "is applied and
  reported `applied` exactly when this invocation's probe found it
  present") leaves refusal as the only consistent outcome when the
  probe said present but the ruleset cannot be built. The worker still
  never manufactures `capability_evidence_invalid` (rework-2 stands);
  the code matches the existing spawn-time enforcement-failure
  classification.
- Off Linux the stub fails any installable claim (fail-closed); it is
  unreachable in practice (non-Linux inventories never install these
  controls; forced-Linux fixtures probe them absent).

Fix (row, `derive_test.go`): `paths-derive-project-root` creates
`rel/path` before the session — a path set grants exactly its members,
so the row provides them. A production declaration with a not-yet-
existing member now refuses with a clear diagnostic (docs +
troubleshooting added) instead of contradicting; the bound is stated in
`docs/troubleshooting.md` ("declare paths that exist at invocation
time").

New row (ubuntu-arbitrated, ledger-registered, 314 rows):
`TestLinuxMissingDerivedPathRefusesWriteConfinement` drives the old
paths-derive shape (a missing `no-such-path` member) as a NEGATIVE
production-entry row: it requires the probe to find write confinement
present, then pins code `worker_protocol_invalid`, detail `cannot
install inventory control`, the control name, the missing member, the
ABSENCE of the exec-denial name (the per-control split), and that the
interpreter never ran (STUB_MARKER absent).

## Rev4 cause 2 — descendant identity via a side file (both ubuntu lanes)

`TestRunSessionTerminatesDescendants` failed at the post-session
`ReadFile` (`descendant.pid: no such file`) after the stub's full 10 s
wait; the session itself succeeded. The pidfile side channel is the
row's assumption, and under write confinement (which grants only the
derived set) a side file outside the grants is EACCES. The exact
denied operation (spawn vs write) could not be reproduced from this
macOS host (no Linux runtime; docker daemon unresponsive), so per the
rework ruling the ROW was fixed, not the confinement:

- The stub (`testdata/stubinterp/main.go`) now publishes the descendant
  through the stdout protocol report: new `spawn_pid`/`spawn_err`
  fields; trigger renamed `STUB_SPAWN_PIDFILE` → `STUB_SPAWN_SLEEP`
  (the value carried a path; nothing carries one now); the sleep child
  writes nothing; the 10 s file-wait is gone (a successful start is
  proof of life — the child sleeps 120 s).
- The test keeps the `repo` declaration (the row still proves teardown
  UNDER confinement) and reads the pid from `result.Stdout`; a spawn
  failure now fails visibly with the spawn error instead of a mystery
  ENOENT. Local runtime 14 s → ~1 s.

Residual risk (honest): if the grandchild spawn itself were denied on
ubuntu, this row still fails — but with the spawn error in the failure
line, which is the diagnostic the next step needs. The spawn target is
the granted interpreter file itself, the same file the (working)
interpreter exec proves executable, so success is predicted.

Side-file audit (rework-3 §24): every other interpreter side-effect
vector checked. All `STUB_MARKER` rows are raw-worker tests whose
fixture claims no Linux-only control (`fixtureInventoryPresent`
defaults false; production sessions claim the probed set through
`runSession` instead), so no enforcement applies and the markers are
unaffected. `STUB_WRITE_TARGETS`/`STUB_EXEC_TRY` rows write/spawn from
the interpreter directly and assert the allow/deny outcome by design.
Diagnostics are written by the unconfined parent. No other row depends
on a confined side file.

## Rev4 cause 3 — Lint (2 issues)

`landlock_linux.go:212,215` ST1005: both `Landlock is not provided...`
error strings lowercased. No new `unsafe`, switches, params, or
suppressions in rev4 (verified by inspection; golangci-lint itself
still cannot run on this host — unchanged environmental bound).

## Rev4 mutants (narrowing; N6–N7 ubuntu-arbitrated)

Both live in `apply.go` (compiled on darwin too), so neutrality is
proven by a green full darwin suite WHILE APPLIED, not just by
build-tag exclusion.

| Mutant | Narrowing | Killing row | Outcome |
|---|---|---|---|
| N6 restore the status flip (omit failed controls, report unavailable) | pre-ready refusal removed | `TestLinuxMissingDerivedPathRefusesWriteConfinement` | LINUX-COMPILE-VERIFIED, darwin suite green while applied (22.6s); predicted kill on ubuntu: worker reports unavailable vs probe present → `capability_evidence_invalid`, row wants `worker_protocol_invalid` |
| N7 exec check reads the write cause (`execDenial && writeErr != nil`, exec cause ignored) | per-control split crossed | `TestLinuxMissingDerivedPathRefusesWriteConfinement` | LINUX-COMPILE-VERIFIED, darwin suite green while applied (24.8s); predicted kill on ubuntu: refusal names `descendant-exec-denial` → "must not void the execute denial" fires |

All mutants reverted; tree verified mutant-free (`grep MUTANT` empty)
and gofmt-clean; targeted rows re-run green after revert.

## Rev4 local verification (reran myself, exit codes real)

- `go build ./...`, `GOOS=linux go build ./...`,
  `GOOS=windows go build ./...` → 0.
- `go vet` darwin on scriptworker/scriptpolicy/install/cmd/curator →
  0; `GOOS=linux go vet` on scriptworker → 0; `GOOS=windows go vet`
  on scriptworker + cmd/curator → 0. `gofmt -l` on touched dirs → empty.
- `go test ./internal/scriptworker/ ./internal/scriptpolicy/ -count=1`
  → ok (scriptworker ~20–28s, scriptpolicy ~1–2s), incl. the reworked
  `TestRunSessionTerminatesDescendants` (~1s, was 14s) and all five
  `TestOperationPrivateRuntimeAreaApplied` subtests; the two new/
  reworked linux-only rows SKIP via the ledger vocabulary.
- `go test ./cmd/curator/ -run
  'TestProductionBinaryDispatchesLandlockProbe|
  TestProductionBinaryDispatchesNetNSProbe'` → both PASS.
- `ledger-consistency.sh` → ok, 314 rows checked.
- Not rerun: full install package (untouched by this rework — no
  install/ file changed in rev4); hosted ubuntu/windows lanes (rev4
  gate arbitrates). No container runtime usable here, so Linux behavior
  is proven by the ubuntu lanes only.

## Rev4 files

Production: `internal/scriptworker/landlock_linux.go` (per-control
construct, ST1005), `internal/scriptworker/landlock_other.go` (stub
shape), `internal/scriptworker/apply.go` (refuse, never flip).
Tests: `landlock_test.go` (per-control construction incl. mixed
bogus-exec/bogus-write cases), `preflight_test.go` (new negative row),
`derive_test.go` (paths-derive provides `rel/path`),
`client_test.go` + `worker_test.go` + `testdata/stubinterp/main.go`
(pid via stdout). Ledger: `platform-cases.tsv` (+1). Docs:
`docs/troubleshooting.md` + `CHANGELOG.md` (new refusal).

## Rev4 undone

1. Hosted ubuntu-latest / windows-latest rev4 evidence (CI lanes
   arbitrate; cannot run here).
2. Mutant kills N6–N7 (ubuntu-arbitrated; predicted above).
3. The exact denied operation behind the rev3 pidfile failure was not
   isolated from macOS (no Linux runtime); the row no longer depends
   on it, and a spawn denial would now surface verbatim in the
   failure line.

---

# Revision 5 (rework 4, COMPLETE — awaiting rev5 gate)

Gate rev4 FAILED (run 35680994406) on the two ubuntu lanes with exactly
TWO rows; Lint and the other lanes green. Previous sections unchanged
(run-2 + rev2 + rev3 + rev4 evidence kept). Shell: bash. Local lane:
darwin/arm64, Go 1.26.0. No new tests, no ledger change (314 rows),
no row touched — both fixes are production (+ comments/docs) only.

## Rev5 cause 1 — confined descendant stdio cannot open /dev/null (both ubuntu lanes)

`TestRunSessionTerminatesDescendants`: "the interpreter descendant did
not start: open /dev/null: permission denied". The confined stub
spawns its sleep child with nil stdio, so Go's spawn opens
`/dev/null` for the child (read for stdin, write for stdout/stderr);
with write confinement enforced and no rule covering the null device,
the write-opens are denied with EACCES. This is a property of the
control, not of the row: any script redirecting to `/dev/null` would
break the same way.

Fix (production, `internal/scriptworker/landlock_linux.go`): under
write confinement the ruleset constructor adds a file-typed rule for
the null device after the derived grants. Reads are never handled, so
no read rule is needed or possible (a rule may only carry handled
rights — a READ rule with reads unhandled would EINVAL); the rule
carries `WRITE_FILE` only, and `landlockAddGrant`'s object-type
narrowing keeps truncation off the non-directory, so the rev1 EINVAL
shape cannot recur. Everything else stays denied. Confirm and enforce
share the constructor, so the pre-ready confirmation and the enforced
domain cannot diverge; a host where the null device cannot be ruled
refuses fail-closed with the install refusal naming it.

Side effect (for the better): `TestLinuxLandlockConfinementMatchesProbe`
passed its exec pin on rev4 even though the `STUB_EXEC_TRY` spawn also
opens `/dev/null` for its nil stdio — the denial it pinned may have
been the stdio denial rather than the exec denial, since both surface
as "permission denied". With the null device granted, the pin now
tests the actual exec denial.

Comments updated to the new applied-rule set: `landlock_linux.go`
(constructor doc + grant site), `inventory.go` (`landlockGrants`),
`landlock.go` (rule rights + grants derivation), `worker.go` (the
pre-bound stdin handle needs no rule for the spawn itself; opens from
inside the domain are covered by the null-device rule). The stale
"(a path rule on it returns EINVAL)" claim is corrected: only
directory-typed rights on it would. Docs: `docs/script-interpreters.md`
(grants the derived path set, the operation-private area, and the null
device), `docs/troubleshooting.md` (new applied-rule set),
`CHANGELOG.md` (same line as the refusal it extends).

## Rev5 cause 2 — install refusal did not name the missing member (both ubuntu lanes)

`TestLinuxMissingDerivedPathRefusesWriteConfinement`: the refusal was
`script_execution_worker_protocol_invalid: cannot install inventory
control "filesystem-write-confinement"` with no member. Root cause:
the offending path lived only in the wrapped cause (`cannot grant
Landlock rights on "/…/no-such-path": …`), and `Diagnostic.Error()`
renders code + detail only — the wrapped cause never crosses the
worker session boundary (the failure frame carries code + detail).

Fix (production, `internal/scriptworker/apply.go`): the install
refusal detail appends the construction cause (`…control %q: %s`),
keeping the class and the per-control attribution. The cause quotes
the offending path with `%q`, like every other path a diagnostic
reports; all grants are manager-derived absolute clean NUL-free
paths, so no further sanitization applies. The wrapped cause stays on
the worker-local error chain as before.

## Rev5 verification (reran myself, exit codes real)

- `go build ./...`, `GOOS=linux go build ./...`,
  `GOOS=windows go build ./...` → 0.
- `go vet` darwin on scriptworker/scriptpolicy/install/cmd/curator →
  0; `GOOS=linux go vet` on scriptworker → 0; `GOOS=windows go vet`
  on scriptworker + cmd/curator → 0. `gofmt -l` on touched dirs →
  empty.
- Fix-2 mechanics proven on darwin by a temporary in-package probe
  (deleted after the run; tree verified without it): with only
  write-confinement installable, `applyScriptControls` refuses with
  code `worker_protocol_invalid`, detail containing `cannot install
  inventory control`, the control name, and the stub cause — and not
  the execute-denial name. PASS. The production-entry row itself is
  ubuntu-arbitrated.
- `go test ./internal/scriptworker/ ./internal/scriptpolicy/ -count=1`
  → ok (scriptworker ~21–23s, scriptpolicy ~1s).
- `go test ./cmd/curator/ -run
  'TestProductionBinaryDispatchesLandlockProbe|
  TestProductionBinaryDispatchesNetNSProbe'` → both PASS.
- `ledger-consistency.sh /tmp/rev5-ledger` → ok, 314 rows checked
  (unchanged: no test added, none renamed).
- Not rerun: full install package (untouched by this rework — no
  install/ file changed in rev5); hosted ubuntu/windows lanes (rev5
  gate arbitrates). No container runtime usable here, so Linux
  behavior is proven by the ubuntu lanes only.
- Lint: `go vet` clean as above; golangci-lint itself cannot run on
  this host (unchanged environmental bound from run 2). New code
  adds no `unsafe`, no switch, no unused param, no suppression, and
  no capitalized error string (verified by inspection).

## Rev5 mutants (narrowing; N8–N9 ubuntu-arbitrated)

| Mutant | Narrowing | Killing row | Outcome |
|---|---|---|---|
| N8 omit the null-device grant (`landlock_linux.go`, condition `&& false`) | write confinement grants the derived set but not `/dev/null` | `TestRunSessionTerminatesDescendants` | LINUX-COMPILE-VERIFIED, darwin suite green while applied (22.4s); predicted kill on ubuntu: descendant spawn fails `open /dev/null: permission denied` |
| N9 install refusal drops the cause (`apply.go`, detail without `: %s`) | refusal names the control but not the path | `TestLinuxMissingDerivedPathRefusesWriteConfinement` | LINUX-COMPILE-VERIFIED, darwin suite green while applied (20.9s); predicted kill on ubuntu: `want it to name the missing member` |

N8 lives in `landlock_linux.go` (darwin neutrality by build-tag
exclusion); N9 lives in `apply.go` (darwin neutrality proven by the
green suite — no darwin row pins the cause detail). All mutants
reverted; tree verified mutant-free (`grep MUTANT` exit 1) and
gofmt-clean.

## Rev5 files

Production: `internal/scriptworker/landlock_linux.go` (null-device
rule), `internal/scriptworker/apply.go` (cause-carrying refusal).
Comments: `inventory.go`, `landlock.go`, `worker.go`,
`client_test.go` (grant-set wording only; the row itself untouched).
Docs: `docs/troubleshooting.md` (refusal symptom + applied-rule
set), `docs/script-interpreters.md` (applied-rule set),
`CHANGELOG.md` (refusal names the path; null-device grant).

## Rev5 undone

1. Hosted ubuntu-latest / windows-latest rev5 evidence (CI lanes
   arbitrate; cannot run here).
2. Mutant kills N8–N9 (ubuntu-arbitrated; predicted above).

---

# Handoff run (rev5 tree re-verified, no code change)

Prior run completed rev5 (rework-4) in-tree but never published; this run
re-verified the identical tree and publishes it. No file changed, added, or
renamed in this run (`git status` delta unchanged: rev5 production +
comment/docs files only). Shell: bash. Local lane: darwin/arm64, Go 1.26.0.

Re-ran myself, exit codes real:
- `go build ./...`, `GOOS=linux go build ./...`,
  `GOOS=windows go build ./...` → 0.
- `go vet` darwin on scriptworker/scriptpolicy/install → 0;
  `GOOS=linux go vet` on scriptworker → 0; `gofmt -l` on touched dirs → empty.
- `go test ./internal/scriptworker/ ./internal/scriptpolicy/ -count=1` →
  ok (scriptworker 18.3s, scriptpolicy 0.4s).
- `go test ./cmd/curator/ -run
  'TestProductionBinaryDispatchesLandlockProbe|
  TestProductionBinaryDispatchesNetNSProbe'` → both PASS (7.9s).
- `ledger-consistency.sh` → ok, 314 rows checked (unchanged).
- Tree verified mutant-free (`grep MUTANT` zero hits in scriptworker/scriptpolicy).
- Not rerun: full install package (no install/ file in the rev5 delta);
  hosted ubuntu/windows lanes (rev5 gate arbitrates).

Undone (unchanged from Rev5): hosted rev5 evidence; mutant kills N8–N9.
---

# Revision 6 (rework 5, COMPLETE — awaiting rev6 gate)

Verdict rev5: CHANGES_REQUESTED (reviewer contract tests attached,
`TASK-260916-1l44nd_review-verdict-rev5.md`). Two HIGH findings, both
in `internal/scriptworker/landlock.go`; probe design, evidence record,
preflight, permit gating, and Windows job objects stand as verified.
Previous sections unchanged (run-2 + rev2–rev5 + handoff evidence
kept). Shell: bash. Local lane: darwin/amd64, Go 1.26.0. Two ledger
rows added (316 checked, was 314); two Linux enforcement rows added;
no other row touched.

## Rev6 cause 1 (F1) — wrong UAPI bit leaves truncation unrestricted

`landlockAccessFSTruncate = 1 << 12` is MAKE_SYM (0x1000); the kernel
UAPI defines TRUNCATE as `1 << 14` (0x4000). The rev5 ABI-4 write mask
was therefore WRITE_FILE | MAKE_SYM (0x1002, as the reviewer
observed), and real truncation — `truncate(path,0)`,
`open(O_RDONLY|O_TRUNC)` — was unhandled and therefore allowed
outside every grant while evidence claimed write confinement applied.
The object-typing rule then stripped the mislabeled bit from all
non-directories (the real cause of the earlier EINVAL on file grants),
which after a constant-only fix would wrongly deny granted file
truncation. The existing `os.WriteFile` denial row could not catch
either half: it never truncates an existing outside file, and it never
grants an individual file.

## Rev6 cause 2 (F2) — handled mask omits directory mutation rights

The rev5 mask also omitted REMOVE_FILE, REMOVE_DIR, MAKE_DIR,
MAKE_REG, MAKE_FIFO (and MAKE_SOCK/MAKE_CHAR/MAKE_BLOCK/MAKE_SYM,
REFER, IOCTL_DEV): Landlock permits every unhandled action, so a
confined script could unlink, rmdir, mkdir, mkfifo, and same-directory
rename outside its grants while reporting the control applied —
contradicting `docs/script-interpreters.md` ("everything else stays
denied"). The existing confinement row checks file-content writes and
exec only: 0 tests for the five outside-directory mutation actions.

## Rev6 fix — UAPI constants, full ABI-gated mask, correct typing

Production (`internal/scriptworker/landlock.go`,
`landlock_const_linux.go` NEW, `landlock_const_other.go` NEW,
`landlock_linux.go`, `landlock_other.go`):

- Every access-right constant comes from `golang.org/x/sys/unix`
  (`LANDLOCK_ACCESS_FS_*`) on Linux. x/sys exposes those symbols on
  Linux only, so the off-Linux mirror carries the same UAPI values as
  literals (they never reach a syscall there: the probe reports
  absent); the mask tests pin both against independent header
  literals on every platform, so a drift fails where it lives.
- `landlockHandledForABI` handles every filesystem right the probed
  ABI provides for the installable controls. ABI to mask table
  (write-confinement set; exec-denial adds EXECUTE 0x1):

  | ABI | Added rights | Write set | With exec |
  |---|---|---|---|
  | 1 | WRITE_FILE 0x2 | 0x2 | 0x3 |
  | 2 | REMOVE_DIR 0x10, REMOVE_FILE 0x20, MAKE_CHAR 0x40, MAKE_DIR 0x80, MAKE_REG 0x100, MAKE_SOCK 0x200, MAKE_FIFO 0x400, MAKE_BLOCK 0x800, MAKE_SYM 0x1000, REFER 0x2000 | 0x3FF2 | 0x3FF3 |
  | 3–4 | TRUNCATE 0x4000 (ABI 4 adds no filesystem right) | 0x7FF2 | 0x7FF3 |
  | 5+ | IOCTL_DEV 0x8000 | 0xFFF2 | 0xFFF3 |

  Never handled: READ_FILE/READ_DIR (reads stay unrestricted — the
  control is write confinement, and the interpreter keeps reading
  system objects outside its grants), the network set (the netns
  covers it), and rights beyond this build's UAPI (RESOLVE_UNIX on
  ABI 9+, scope flags) — documented in code and results, allowed on
  kernels that provide them.
- `landlockRuleRights` strips only the ten directory-only rights
  (REMOVE_*, MAKE_*, REFER) from non-directory grants; files keep
  WRITE_FILE | TRUNCATE | IOCTL_DEV. TRUNCATE is valid on regular
  files (truncate/ftruncate/creat/open-O_TRUNC); overwriting an
  existing file needs it in addition to WRITE_FILE.
- The null-device grant carries the file-typed write set
  (`landlockFileAccessForABI`: WRITE plus TRUNCATE/IOCTL_DEV per
  ABI), not WRITE alone: with real truncation handled, an open of
  the sink with O_TRUNC (a `> /dev/null` shape, a `DEVNULL` handle
  opened for writing) would otherwise fail with EACCES. Reads need
  no rule (unhandled rights cannot appear in a rule).
- Directory-only rights are granted only beneath the derived
  writable directories (operation-private area + derived path set);
  no writable root was broadened. Confirm and enforce share the
  constructor, so the pre-ready confirmation and the enforced domain
  cannot diverge; the probe child enforces the same shaped sequence
  (directory rule, so all rights valid). Ruleset attr stays size 8:
  every filesystem ABI shares the first u64.
- `landlock_other.go` gains a `landlockABI` stub (tests call it
  behind a Linux gate; keeps untagged test files portable).

## Rev6 rows (ubuntu Test + Race, through the real worker)

Stub protocol (`testdata/stubinterp/main.go`, `mkfifo_unix.go` NEW,
`mkfifo_other.go` NEW — the FIFO helper is build-tagged so the
double still compiles on Windows): STUB_TRUNCATE_TARGETS drives
truncate(2) + O_RDONLY|O_TRUNC per entry (truncations run before
writes, so a both-listed file ends with the write payload);
STUB_UNLINK/RMDIR/MKDIR/MKFIFO_TARGETS and parallel
STUB_RENAME_SRC/DST drive the five mutations; seven new report maps
(`truncate`, `otrunc`, `unlink`, `rmdir`, `mkdir`, `mkfifo`,
`rename`), decoded into `stubReport` (`worker_test.go`).

- `TestLinuxWriteConfinementTruncateMatchesProbe` (row a): with an
  individual-file grant (`filesystem: ["granted.txt"]`) the granted
  file truncates (truncate + O_TRUNC) and overwrites, ending with
  the write payload; an existing outside file's truncate + O_TRUNC
  are refused EACCES with content preserved (ABI 3+), or succeed
  (ABI < 3, right unhandled — asserted, not skipped); evidence
  `applied` in the same invocation.
- `TestLinuxWriteConfinementDirectoryMutationMatchesProbe` (row b):
  unlink, rmdir, mkdir, mkfifo, and same-directory rename outside
  the grants are refused EACCES with state preserved (ABI 2+), or
  succeed (ABI 1 — asserted); the same five inside a granted
  directory succeed with effects shown (victim gone, dir/FIFO
  created, rename moved with content); evidence `applied` in the
  same invocation.
- Both rows require the probe to find write confinement present
  (without enforcement they would prove nothing — the
  missing-path-row precedent) and adapt per-right to the probed ABI
  (the MatchesProbe precedent), so they are deterministic on any
  Linux host and strict on the hosted ABI-4 runner; no timing, no
  sleeps, race-clean by construction. Every stub attempt is
  presence-asserted, so a withheld `env_read` fails the row instead
  of passing vacuously.
- `TestLinuxLandlockConfinementMatchesProbe` unchanged in code
  (consistent: the denied new-file creation now refuses on the
  ungranted MAKE_REG right — still EACCES); one comment sentence
  records the mechanism. Existing-suite audit: no other ubuntu row
  creates or mutates files outside its grants (markers flow through
  raw-worker rows that claim no Landlock control, refusal rows that
  never start the interpreter, or stdout reports), so the broadened
  handled set breaks no existing row.
- Ledger: two rows, `linux` / `darwin,windows` via `platform-control`
  with the exact skip vocabulary (316 rows checked).

## Rev6 mutants (narrowing; unit kills executed, row kills ubuntu-arbitrated)

Each mutant is one omitted right in `landlock.go`; the unit kill ran
here (darwin), the row kill is the precise predicted ubuntu failure
(Linux enforcement cannot run on this host — no container runtime).

| Mutant | Narrowing | Unit kill (executed) | Predicted ubuntu row kill |
|---|---|---|---|
| N10 omit real TRUNCATE (`abi >= 3` → `>= 99`) | handled set without 0x4000 | `TestLandlockHandledMaskFollowsABI`: 8 subtests FAIL (abi-3/4/5/6/9, exec-only, write-only, no-control) | Truncate row: outside truncate + O_TRUNC succeed → "was not refused with EACCES" |
| N11 omit REMOVE_FILE | dir-only set without 0x20 | mask test: abi-2+ subtests FAIL | Mutation row: outside unlink + outside rename succeed → EACCES pin fails |
| N12 omit REMOVE_DIR | dir-only set without 0x10 | mask test: 9 subtests FAIL | Mutation row: outside rmdir succeeds → EACCES pin fails |
| N13 omit MAKE_DIR | dir-only set without 0x80 | mask test: abi-2+ subtests FAIL | Mutation row: outside mkdir succeeds → EACCES pin fails |
| N14 omit MAKE_REG | dir-only set without 0x100 | mask test: abi-2+ subtests FAIL | Mutation row: outside rename succeeds → EACCES pin fails |
| N15 omit MAKE_FIFO | dir-only set without 0x400 | mask test: abi-2+ subtests FAIL | Mutation row: outside mkfifo succeeds → EACCES pin fails |
| N16 file rule strips TRUNCATE (`&^= TRUNCATE` on non-dirs) | granted files lose truncation | `TestLandlockRuleRightsFollowObjectType`: 3 subtests FAIL (file-keeps-write-and-truncate, file-sheds-directory-mutation, device-sheds-directory-mutation) | Truncate row: granted-file truncate/otrunc/overwrite all fail → "truncate of the granted file failed" |

N10 also `GOOS=linux go build ./...` clean while applied
(LINUX-COMPILE-VERIFIED spot check; the other six are single-line
changes in the same portable file). All mutants reverted; tree
verified mutant-free (`grep MUTANT` exit 1) and the mask tests green
after restore. Note: N12's first patch missed (gofmt aligns the
const block; the pattern assumed one space) and ran green on the
clean tree — caught by the assert, rerun correctly, FAIL as above.
No other mutant harness issue.

## Rev6 verification (reran myself, exit codes real)

- `go build ./...`, `GOOS=linux go build ./...`,
  `GOOS=windows go build ./...` → 0.
- `go vet` darwin on scriptworker/scriptpolicy/install → 0;
  `GOOS=linux go vet` on scriptworker → 0; `GOOS=windows go vet`
  on scriptworker + cmd/curator → 0. `gofmt -l` on internal/ and
  docs/ → empty.
- `go test ./internal/scriptworker/ ./internal/scriptpolicy/ -count=1`
  → ok (scriptworker 16.8s: 192 PASS, 0 FAIL, 10 platform-vocabulary
  skips incl. the 2 new rows; scriptpolicy 0.4s). NOTE: the first
  combined invocation stalled 8+ min nearly idle (shared-host
  contention — an unrelated agent `go test` ran concurrently) and
  was terminated; the clean rerun above is the evidence. No code
  changed between the two.
- Reviewer contract tests (`TestReviewerTruncateUAPIIdentity`,
  `TestReviewerWriteConfinementMask`,
  `TestReviewerRegularFileTruncationGrant`, copied in from the
  verdict attachment, removed after): 3/3 PASS on the fixed tree.
- `go test ./cmd/curator/ -run
  'TestProductionBinaryDispatchesLandlockProbe|
  TestProductionBinaryDispatchesNetNSProbe'` → ok (3.5s).
- `go test ./internal/install/ -run
  '^TestEnforcedInstallAndLaunchAtCLIEntry$'` → ok (62.2s): the
  extended stub still launches end to end at the CLI entry.
- `ledger-consistency.sh` → ok, 316 rows checked (+2 new).
- `go build` of the stub on darwin/linux/windows → ok (no stray
  binaries left in the tree).
- Not rerun: full install package (only the CLI-entry row above;
  no install/ file changed); hosted ubuntu/windows lanes (rev6
  gate arbitrates). No container runtime usable here, so Linux
  enforcement is proven by the ubuntu lanes only.
- Lint: `go vet` clean as above; golangci-lint itself cannot run on
  this host (unchanged environmental bound). New code adds no
  `unsafe`, no switch, no unused param, no suppression, and no new
  error-string shape (the `landlockABI` stub mirrors the two
  pre-existing capitalized Linux-only errors verbatim); test and
  testdata additions follow the existing `#nosec`-free patterns.

## Rev6 files

Production: `internal/scriptworker/landlock.go` (full ABI-gated
mask, dir-only/file typing, file-typed null grant),
`landlock_const_linux.go` NEW (x/sys UAPI constants),
`landlock_const_other.go` NEW (off-Linux mirror),
`landlock_linux.go` (null-grant set, constructor/struct comments),
`landlock_other.go` (`landlockABI` stub), `inventory.go`
(`landlockGrants` doc). Tests: `landlock_test.go` (independent-UAPI
expectations, ABI 1/2/3/4/5/6/9, read-never-handled pin),
`preflight_test.go` (2 new rows + MatchesProbe mechanism note),
`worker_test.go` (`stubReport` fields), `testdata/stubinterp/main.go`
+ `mkfifo_unix.go`/`mkfifo_other.go` NEW (7 new STUB probes).
Ledger: `.github/ci/platform-cases.tsv` (+2 rows). Docs:
`docs/script-interpreters.md` (handled set, "everything else stays
denied" now true), `docs/troubleshooting.md` (denied-op inventory),
`CHANGELOG.md` (R3 bullet extended).

## Rev6 undone

1. Hosted ubuntu-latest / windows-latest rev6 evidence (CI lanes
   arbitrate; cannot run here).
2. Mutant row kills N10–N16 (ubuntu-arbitrated; predicted above).
3. Bounds carried from rev5 unchanged (R-e pass-through streaming
   deferred to R5; RESOLVE_UNIX/scope-flag rights beyond this
   build's UAPI stay unhandled — new, documented in code).
# Revision 7 (rework 6, COMPLETE — awaiting rev7 gate)

Verdict rev6: CHANGES_REQUESTED with ONE finding (F1 HIGH;
`TASK-260916-1l44nd_review-verdict-rev6.md`, reviewer test
`TASK-260916-1l44nd_review-rev6-abi1_test.go`). Rev5 F1/F2 stand
verified; rev6 scope otherwise accepted. Previous sections unchanged
(run-2 + rev2–rev6 kept as history — the rev6 ABI table above is
superseded by the corrected table here, not edited). Shell: bash.
Local lane: darwin/amd64, Go 1.26.0. One ledger row added (317
checked, was 316); no enforcement row added or removed; one unit
test committed.

## Rev7 cause — ABI-1 directory mutation rights left unrestricted

`landlockHandledForABI` added the whole directory-only set only for
ABI >= 2. Per the Linux v5.13 UAPI header REMOVE_DIR, REMOVE_FILE
and all seven MAKE_* rights exist since ABI 1; only REFER starts at
ABI 2 (TRUNCATE at ABI 3, IOCTL_DEV at ABI 5). The probe accepts ABI
1, so on an ABI-1 host the write mask was 0x2 instead of 0x1ff2 and
a successful application attested `applied` while unlink, rmdir,
mkdir, mkfifo, and same-directory rename outside the grants stayed
unrestricted — the same incomplete-control class as rev5 F2 on a
supported older ABI. Two tests protected the defect:
`landlock_test.go` expected ABI 1 to handle WRITE_FILE only, and
the production mutation row asserted outside-grant mutation
SUCCEEDS on ABI 1.

## Rev7 fix — ABI-1 mutation set, REFER/TRUNCATE/IOCTL still gated

Production (`internal/scriptworker/landlock.go` only):

- New `landlockMutationABI1Access`: the nine ABI-1 rights
  (REMOVE_DIR, REMOVE_FILE, MAKE_CHAR/DIR/REG/SOCK/FIFO/BLOCK/SYM),
  handled from ABI 1 alongside WRITE_FILE. Only REFER stays behind
  `abi >= 2`, TRUNCATE behind `abi >= 3`, IOCTL_DEV behind
  `abi >= 5`.
- `landlockDirectoryOnlyAccess` is now the ABI-1 mutation set plus
  REFER; rule narrowing (`landlockRuleRights`,
  `landlockFileAccessForABI`) still sheds all ten from file grants,
  so the null-device grant and granted files stay EINVAL-free on
  every ABI.
- Corrected ABI to mask table (write-confinement set; exec-denial
  adds EXECUTE 0x1):

  | ABI | Added rights | Write set | With exec |
  |---|---|---|---|
  | 1 | WRITE_FILE 0x2, REMOVE_DIR 0x10, REMOVE_FILE 0x20, MAKE_CHAR 0x40, MAKE_DIR 0x80, MAKE_REG 0x100, MAKE_SOCK 0x200, MAKE_FIFO 0x400, MAKE_BLOCK 0x800, MAKE_SYM 0x1000 | 0x1FF2 | 0x1FF3 |
  | 2 | REFER 0x2000 | 0x3FF2 | 0x3FF3 |
  | 3–4 | TRUNCATE 0x4000 (ABI 4 adds no filesystem right) | 0x7FF2 | 0x7FF3 |
  | 5+ | IOCTL_DEV 0x8000 | 0xFFF2 | 0xFFF3 |

  File-typed grants (null device, granted files): 0x2 on ABI 1–2,
  0x4002 on ABI 3–4, 0xC002 on ABI 5+. Verified against the
  production helper on this host (scratch print, removed after):
  exactly the four expected write masks plus 0x0 on ABI 0.
- On ABI >= 2 the handled set is bit-identical to rev6 (same union,
  regrouped), so the hosted ABI-4 rows behave exactly as before;
  the only behavior change is ABI 1 (0x2 → 0x1ff2). The control is
  NOT suppressed on ABI 1. Writable roots unchanged
  (operation-private area + derived path set + file-typed null
  device); no root broadened. Source comments corrected (ABI
  introduction, handled-mask doc, mutation-set doc).

Tests:

- `reviewer_abi1_test.go` NEW: the reviewer's
  `TestReviewerABI1MutationRights` committed (9 per-right pins +
  exact 0x1ff2 pin + no-REFER-on-ABI-1 pin). Whitespace-only
  difference vs the attachment: gofmt (tabs, brace style) — CI
  fails on `gofmt -l` output, so the committed copy is the
  gofmt-clean form; semantics identical.
- `landlock_test.go`: independent UAPI literals split into
  `uapiMutationABI1` + REFER; `writeABI1` is now
  WRITE|mutationABI1 (0x1ff2); subtests renamed
  `abi-1-write-and-mutation` / `abi-2-adds-refer` (the ledger pins
  the parent name only, so the rename is gate-safe); comments
  fixed.
- `preflight_test.go`:
  `TestLinuxWriteConfinementDirectoryMutationMatchesProbe` requires
  EACCES denial with state preserved on EVERY ABI — the ABI-1
  success branch is deleted, not inverted. Truncate row untouched
  (TRUNCATE is genuinely ABI 3+).
- Ledger: +1 row for `TestReviewerABI1MutationRights`
  (`linux,darwin,windows`, no skip) — 317 rows checked.

## Rev7 mutants (narrowing; executed locally, all killed)

| Mutant | Narrowing | Kill (executed, darwin) |
|---|---|---|
| M1 restore the rev6 defect (mutation set back behind `abi >= 2`) | ABI-1 write mask 0x1ff2 → 0x2 | `TestLandlockHandledMaskFollowsABI/abi-1-write-and-mutation` FAIL + `TestReviewerABI1MutationRights` 9/9 FAIL |
| M2 omit MAKE_REG from the ABI-1 set | handled set without 0x100 | mask test: 7 subtests FAIL (abi-1/2/3/4/5/6/9) |
| M3 ungate REFER (requested on ABI 1) | handled set over-requests 0x2000 on ABI 1 | mask abi-1 subtest FAIL + reviewer "must not request REFER" FAIL |

M3 proves the upper bound too: the fix handles every right ABI 1
provides and no right it does not. All mutants reverted; tree
verified mutant-free by diff against the pre-mutant copy. Honest
bound: M1's production-row kill needs an ABI-1 kernel — neither
this host nor the hosted ABI-4 runner can execute it (on ABI >= 2
M1 is behavior-identical to the fix, so the hosted mutation row
cannot kill it either). The ABI-1 enforcement proof stays at the
mask-helper-plus-UAPI level, as the reviewer noted.

## Rev7 verification (reran myself, exit codes real)

- `go build ./...`, `GOOS=linux go build ./...`,
  `GOOS=windows go build ./...` → 0.
- `go vet` darwin on scriptworker/scriptpolicy/install → 0;
  `GOOS=linux go vet` on scriptworker → 0; `GOOS=windows go vet`
  on scriptworker + cmd/curator → 0 (vet compiles the new test
  per GOOS). `gofmt -l cmd internal` → empty.
- Mask pins: `TestReviewerABI1MutationRights` +
  `TestLandlockHandledMaskFollowsABI` +
  `TestLandlockRuleRightsFollowObjectType` → PASS (reviewer 9/9).
- `go test ./internal/scriptworker/ ./internal/scriptpolicy/ -count=1`
  → ok (scriptworker 43.7s, scriptpolicy 1.6s).
- `go test ./cmd/curator/ -run
  'TestProductionBinaryDispatchesLandlockProbe|TestProductionBinaryDispatchesNetNSProbe'`
  → ok (48.8s).
- `ledger-consistency.sh` → ok, 317 rows checked (+1 new).
- Not rerun: full install package (no install/ file changed, and
  the brief asks only for touched install rows); hosted
  ubuntu/windows lanes (rev7 gate arbitrates). golangci-lint
  itself cannot run on this host (unchanged environmental bound);
  new code adds no `unsafe`, no switch, no unused param, no
  suppression. Linux enforcement on ABI 1 is not executable on
  any available lane (see mutant bound above).

## Rev7 files

Production: `internal/scriptworker/landlock.go` (ABI-1 mutation
set, REFER/TRUNCATE/IOCTL gating, comments). Tests:
`reviewer_abi1_test.go` NEW (committed reviewer contract),
`landlock_test.go` (split UAPI expectations, renamed ABI
subtests), `preflight_test.go` (mutation row denies on every
ABI). Ledger: `.github/ci/platform-cases.tsv` (+1 row). Docs and
CHANGELOG need no change: neither names an ABI version (the
"handles every right the probed ABI provides" wording is now true
on ABI 1 too).

## Rev7 undone

1. Hosted ubuntu-latest / windows-latest rev7 evidence (CI lanes
   arbitrate; cannot run here). Expected: green with zero row
   behavior change on the ABI-4 runner (handled set bit-identical
   there); the new reviewer test passes on all three lanes (pure,
   no skip).
2. ABI-1 kernel enforcement proof (no ABI-1 lane exists anywhere
   in this pipeline; mask-helper-plus-UAPI proof only).
3. Bounds carried from rev6 unchanged (R-e pass-through streaming
   deferred to R5; RESOLVE_UNIX/scope-flag rights beyond this
   build's UAPI stay unhandled).
