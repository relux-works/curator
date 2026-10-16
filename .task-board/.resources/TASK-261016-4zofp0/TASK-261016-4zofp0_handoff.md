# SNAP-LOCK developer handoff

Candidate: `1e6be490fee5768d0aa6808b825a20437058ef1c` (signed detached gate snapshot).
Tree: `b91a63aa36e9e266a7514d9a4d92356181bbfc6f`.
Story checkpoint: `0d00a109ad1da77a5506747c89f24af9442eb341`; the Story branch remains there, with implementation and tests uncommitted for Change Request capture.
Normative source: curator-spec `d2a942cf17e4eb017c45475442da1456d8dfbe6c`, snapshot sections 1/5, environment lifecycle/resolution clauses and CIP-0014 consistency steps 1–5. No specification conflict found.

## Behavior and production entry points

`CaptureMaterialize` opens the existing manager mutation lock with the normal bounded wait, observes journals without recovery, reloads machine policy and selects the current profile under that lock, and shares the read-only resolution refusal path. It invokes a local closure callback, then rechecks effective selection/policy, the stable lock name, pending journals, source currency/fragment, authorizing record identity and every captured object's bytes, size, mode, timestamps and object identity. Output is returned only after lock release. Absent manager/lock state is never provisioned.

`PendingRecovery` validates all records, including cleanup tombs, without resuming or cleaning them. Pending state returns `environment_materialize_recovery_required`. Corrupt evidence retains journal refusals; filesystem read failures retain `manager_state_unreadable`.

`verifyHomeObserved` / `assembleHomeObserved` form the shared content core. `verifyHomeContent` supplies unknown release metadata without invoking vendor discovery. Capture uses `resolveVerified(..., verifyHomeContent)` and metadata-only status uses the same core. Ordinary resolve retains release observation outside that core.

Coverage: **5 of 5 AC rows driven at the production capture-stage entry point**. **0 of 2 CLI forms qualified here**: the actual export and manifest-only commands belong to SNAP-CLI, which must integrate this capture path. Test labels for the two consumers exercise the same capture-stage operation, not an already shipped CLI. No vendor-runtime, archive, manifest or destination publication qualification is claimed.

| AC row | Production call site | Named tests and stated limits |
|---|---|---|
| Non-recovering bounded lock | `CaptureMaterialize` → `beginCaptureOperation` → `AcquireExistingHomeOnly` | `TestMaterializeCaptureCurrentAndRelease`, `TestMaterializeCaptureWaitReselectsAndTimeout`, `TestCaptureLockSharesWriterLock`, `TestAcquireExistingHomeOnlyDoesNotProvision` |
| Pending/read-failure refusal | `beginCaptureOperation` → `PendingRecovery` | `TestMaterializeCapturePendingDoesNotRecover`, `TestMaterializeCaptureBadRecoveryEvidence`, `TestMaterializeCaptureUnreadableJournal`, `TestPendingRecoveryReadOnlyPhasesAndTombs`, `TestPendingRecoveryRejectsLinkedInventory`; `TestRecoveryObserverDetectsEmptyRecovery` proves the non-entry instrument detects an actual recovery call |
| Fresh selection and generation identity | `CaptureMaterialize` selection/revalidation and `ReadObject` / `revalidateObjects` | `TestMaterializeCaptureWaitReselectsAndTimeout`, `TestMaterializeCaptureSameFragmentDrift`, `TestMaterializeCaptureAuthorizingDrift`, `TestMaterializeCaptureStoreObjectIdentity`, `TestMaterializeCaptureExternalTornGeneration`, `TestMaterializeCaptureConcurrentRepair`, `TestMaterializeCaptureSerializesProfileWriters` |
| No vendor observation in content verification | `resolveVerified(..., verifyHomeContent)` and `StatusOf` → `homeState` | `TestMaterializeCaptureAndStatusNeverProbeVendor`; a real PATH executable plant is used with nil Detect, on a successful Codex capture fixture and the status matrix |
| Bounded local capture and release before publication | `ReadObject`, `CaptureMaterialize` return/deferred release | `TestMaterializeCaptureObjectBounds`, `TestMaterializeCaptureCurrentAndRelease`, `TestMaterializeCaptureDoesNotProvision`; byte/object quotas and a capture deadline are local limits, not transport limits. Deadline checks occur between local operations; the trusted callback must perform bounded local work and use protected handles. Arbitrary blocking callback code is not preempted. Transport/publication are outside this operation |

## Paired behavior coverage map

Brief gap: no catalog surface table was supplied. This map covers the five paired behavior rows actually supplied; it does not silently invent a catalog selection.

| Supplied surface row | Positive and negative tests | Killed narrowing mutants | Out-of-contract inputs / acceptance clause |
|---|---|---|---|
| Current source / pending journal | `TestMaterializeCaptureCurrentAndRelease`; `TestMaterializeCapturePendingDoesNotRecover`, `TestMaterializeCaptureBadRecoveryEvidence`, `TestMaterializeCaptureUnreadableJournal` | M-prepared, M-corrupt-absence, M-denied-class, M-provision-absent | None within capture-stage journal semantics |
| Released lock / timeout and stale selection | `TestMaterializeCaptureWaitReselectsAndTimeout`; `TestMaterializeCaptureAuthorizingDrift` | M-lock-name, M-composition-policy | Actual CLI selection/loading adapter remains SNAP-CLI: leaf says “SNAP-CLI integrates both” |
| Cooperating writers / multi-file generation | `TestMaterializeCaptureSerializesProfileWriters`, `TestMaterializeCaptureConcurrentRepair`, `TestMaterializeCaptureExternalTornGeneration`; neighboring same-fragment witness in `TestMaterializeCaptureSameFragmentDrift/aba` | M3-lock-only | Global/profile scope sync and real repair/use are driven; no external service is waited on under the lock |
| Stable objects / same-fragment ABA and store changes | `TestMaterializeCaptureSameFragmentDrift`, `TestMaterializeCaptureStoreObjectIdentity`, `TestMaterializeCaptureAuthorizingDrift` | M3-lock-only, M-executable-mode, M-lock-name | Closure inventory and protected handle acquisition belong to SNAP-CLOSURE, as the leaf explicitly states |
| File/metadata verification / vendor executable | `TestMaterializeCaptureAndStatusNeverProbeVendor`; owned source/native fixture state comparisons in capture tests | M-probe | Actual manifest-only/export publication and vendor-runtime activation remain SNAP-CLI / later runtime qualification |

Negative cases require the intended refusal and no successful output. Source-generation matrices compare the post-refusal source, native credential roots and owned fixture state with the deliberately planted state; pending journals and transaction targets are preserved. Credential links are never dereferenced by fixture-state snapshots.

## Narrowing controls

Each control is applied in an isolated task-local copy. The behavioral Go test is a standalone process, not a static token check. The harness restores the source after each plant. All nine current plants are killed; each Go test below exits **1**, an expected failing gate, not a passing test command.

| Mutant | Gate narrowed to | Named failing test | Surviving bound |
|---|---|---|---|
| M-denied-class | Preserve typed read failures except permission denial | `TestMaterializeCaptureUnreadableJournal` | None; killed |
| M-lock-name | Preserve stable-name identity except a zero-length regular replacement | `TestMaterializeCaptureAuthorizingDrift/lock-identity` | None; killed |
| M-composition-policy | Allow exactly enabling the composition knob with all other effective inputs equal | `TestMaterializeCaptureSameFragmentDrift/composition-policy` | None; killed |
| M3-lock-only | Revalidate the profile lock object while ignoring other captured object identities | `TestMaterializeCaptureSameFragmentDrift/aba` | None; killed; the witness independently asserts equal fragments |
| M-prepared | Refuse pending phases except prepared | `TestMaterializeCapturePendingDoesNotRecover/prepared` | None; killed |
| M-corrupt-absence | Treat the planted broken journal as absent while retaining other journal gates | `TestMaterializeCaptureBadRecoveryEvidence/corrupt` | None; killed |
| M-probe | Permit Codex discovery only; other adapters remain metadata-only | `TestMaterializeCaptureAndStatusNeverProbeVendor` | None; killed |
| M-executable-mode | Compare object modes while ignoring executable bits | `TestMaterializeCaptureSameFragmentDrift/mode` | None; killed |
| M-provision-absent | Provision only absent manager homes while retaining existing-state gates | `TestMaterializeCaptureDoesNotProvision/absent` | None; killed |

Historical control: M-executable-mode initially survived with exit 0 and no named failing test. That bounded the old witness to other permission changes; it did not establish executable-bit coverage. The witness was narrowed to one execute bit and the behavioral plant rerun, producing the named failure above. The development logbook preserves this correction.

No source-text inspection gate was introduced. The vendor plant keeps the verification function names present and changes behavior; the harness executes the Go behavioral suite.

## Validation

Environment: Go 1.25.5, darwin/arm64; golangci-lint 2.12.2; Python is used only to stage isolated mutation copies and invoke existing Go tests. No test library or runtime dependency was added.

| Command | Exit | Evidence |
|---|---:|---|
| `go test -race ./internal/envprofile ./internal/managerlock ./internal/transaction -run 'TestMaterializeCapture\|TestPendingRecovery\|TestRecoveryObserver\|TestCaptureLock\|TestAcquireExistingHomeOnly' -count=1` | 0 | `race-07.log`; current candidate, includes the external torn-generation witness |
| `GOLANGCI_LINT_CACHE="$PWD/.temp/TASK-261016-4zofp0/lint-cache" make lint` | 0 | `lint-09.log`, 0 issues; isolated cache avoids stale entries from unrelated checkouts |
| `go build ./...` | 0 | `build-07.log`; empty output is recorded alongside its process exit |
| `GOOS=windows GOARCH=amd64 go test -c ./internal/envprofile -o .temp/TASK-261016-4zofp0/envprofile-windows.test.exe` | 0 | `windows-compile-02.log`; compilation only, not Windows execution |
| `go vet ./...` | 0 | `vet-05.log`; empty output is recorded alongside its process exit |
| `git diff --check` | 0 | `diff-check-05.log`; no whitespace errors |
| `sh -n scripts/remote-gate.sh` | 0 | `gate-syntax-03.log`; unchanged script identity |
| `sh scripts/remote-gate.sh` after loading the configured agent Git environment | 0 | [Hosted gate](https://github.com/relux-works/curator/actions/runs/38845304296), exact signed snapshot above |
| Mutation harness, invoking nine bounded standalone `go test ... -run ... -count=1 -timeout=2m` commands | 0 harness; 1 each Go plant | Attached harness and machine-readable mutant outcomes; expected red results remain red |

The first frozen gate (`90685b712e500a752f23430a9b5f9385740cd8b5`, run 38835465876) exited 1: Windows fixture snapshots read a held bookkeeping lock, identity replacement plants failed, and the platform-case gate rejected unknown skip descriptions. The second frozen gate (`37f08891e6d0325cc6cd2d3c9a9d604de1860719`, run 38840074132) also exited 1: the bookkeeping and skip corrections passed, but all four held-object replacement plants still received AccessDenied. Clearing readonly was insufficient; that earlier hypothesis is not accepted as the failure cause. Both gates had nineteen other enabled jobs green.

The green candidate fixture uses native Windows FileRenameInformationEx with POSIX replacement semantics and the readonly-attribute flag through the existing x/sys dependency. Production operations and CI gates are unchanged. A real sharing violation is an explicit unavailable plant, not passing coverage. Platform-specific fixture helpers are test-only additions outside the leaf's named new file. Their purpose and native API references are preserved in the logbook.

The earlier permission-denied regression command exited 1 before the read-class fix. Earlier capture/lint development failures and six superseded gate runs are not green evidence. Superseded remote gate processes were explicitly interrupted with exit 137, their workflows cancelled and temporary branches removed; the accepted gate must be the exact candidate linked above.

A previous full local package run also exited 0, but predates the last object-revalidation/test additions and is diagnostic only. Acceptance relies on the current focused race suite and exact-candidate hosted full gate rather than claiming that older run tested this tree.

Platform bounds: POSIX executable-mode, shell-executable and permission-denial witnesses explicitly skip on Windows using the existing platform-control classes. Held lock-name replacement skips if Windows sharing prevents the plant. Metadata-only status is exercised with a Codex managed fixture; no live vendor login/model/refresh behavior is qualified. Workflow-configured optional jobs are declared separately in the stored gate outcome; a skipped job is not a passing runtime qualification.

## Scope and retained tests

- Test-only platform fixture dispatch: `internal/envprofile/materialize_fixture_unix_test.go` and `materialize_fixture_windows_test.go` allow the same identity witness on supported host APIs without changing product rename behavior.
- Declared implementation: new `internal/envprofile/materialize_capture.go`, narrow `internal/envprofile/lock.go`, and read-only observer/tests in `internal/transaction/`.
- Required shared seam: `internal/envprofile/managed.go` and `status.go` separate release observation from content verification and reuse all existing resolution refusals.
- Required lock primitive: `internal/managerlock/filelock.go`, `managerlock.go`, and `capture_test.go` open existing lock state without creating directories/files and revalidate the held name.
- Recovery-entry observation in `internal/transaction/engine.go` / `types.go` is nil in production and supports the bounded non-entry/control evidence.
- `README.md` documents the internal capture seam, metadata-only observation and commands, as required by repository documentation rules.
- `scripts/remote-gate.sh` has the minimal signing-policy correction needed to run the required full gate with the configured human author; its candidate signature is verified before push.

All previously committed tests remain under their original names; existing test files are unchanged. New named tests are present in the signed gate snapshot while the Story checkout stays uncommitted. No login/store, profile-prepare, closure, destination, archive, launcher or runtime implementation is included.

The leaf declares SNAP-CLOSURE as the content callback owner and SNAP-CLI as the integrator. Those remaining integrations must supply fresh effective configuration loading, protected handles, admitted inventory/relocation/secret checks and actual publication. This handoff claims the capture transaction and its tested boundaries, not those dependent products.

The full command `sh scripts/remote-gate.sh` returned **exit 0**. All 20 enabled jobs passed, including hosted Test on Linux/macOS/Windows, Race on Linux/macOS, gate self-tests and the interop property suite. The optional candidate suite was not requested; the optional self-hosted test is not enabled for gate snapshots. Those two skipped jobs are not runtime qualification. The stored gate matrix records those exact conditions.

The fresh post-gate identity check confirmed signature `G`, candidate tree `b91a63aa36e9e266a7514d9a4d92356181bbfc6f`, unchanged Story checkpoint and an empty real index. `git diff --exit-code HEAD -- ':(glob)**/*_test.go'` returned 0, independently confirming retention of all previously tracked test files.

## Hosted leaf runtime evidence

| Platform | Pass records | Skip records | Bound |
|---|---:|---:|---|
| Linux | 54 | 0 | All supplied leaf cases ran |
| macOS | 54 | 0 | All supplied leaf cases ran |
| Windows | 46 | 6 | Two held-lock identity plants are prevented by OS sharing; two executable-mode cases, the shell vendor sentinel and POSIX unreadability are unavailable |

Counts are named terminal Go test/subtest records; parents and children are counted separately. Windows successfully executed the same-fragment ABA, store-object identity and marker-identity witnesses using the native replacement fixture. The six skipped cases are declared bounds, not passing qualification; the corresponding Unix cases ran. Exact names and verdicts are in `TASK-261016-4zofp0_platform-cases.json`. Narrowing mutants were executed on darwin/arm64, not replayed on Windows.

## DoD checklist readback

The authoritative board checklist was read back after the full gate: **25 of 25 items checked**. Items 1–5 are mapped individually in the AC table; items 21–25 are the five individually tracked paired behavior rows mapped above. Items 6–20 are evidenced by the negative matrices, nine named narrowing controls, exact-command validation table, retained-test diff, unchanged checkpoint/index, scope declarations, attached outcomes and logbook. The source-text-gate item is explicitly not applicable: no such gate was introduced; behavioral controls preserve the production function names.
