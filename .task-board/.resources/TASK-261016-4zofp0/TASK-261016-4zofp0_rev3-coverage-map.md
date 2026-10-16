# SNAP-LOCK capture-stage coverage map

Candidate: `4cf8cc94ef36be4e95888d57b48f3cf2314b9b47`.

Coverage: **5 of 5 AC rows driven at the production capture-stage entry point**. **0 of 2 CLI forms qualified here**: the actual export and manifest-only commands belong to SNAP-CLI, which must integrate this capture path. Test labels for the two consumers exercise the same capture-stage operation, not an already shipped CLI. No vendor-runtime, archive, manifest or destination publication qualification is claimed.

| AC row | Production call site | Named tests and stated limits |
|---|---|---|
| Non-recovering bounded lock | `CaptureMaterialize` → `beginCaptureOperation` → `AcquireExistingHomeOnly` | `TestMaterializeCaptureCurrentAndRelease`, `TestMaterializeCaptureWaitReselectsAndTimeout`, `TestCaptureLockSharesWriterLock`, `TestAcquireExistingHomeOnlyDoesNotProvision` |
| Pending/read-failure refusal | `beginCaptureOperation` → `PendingRecovery` | `TestMaterializeCapturePendingDoesNotRecover`, `TestMaterializeCaptureBadRecoveryEvidence`, `TestMaterializeCaptureUnreadableJournal`, `TestPendingRecoveryReadOnlyPhasesAndTombs`, `TestPendingRecoveryRejectsLinkedInventory`; `TestRecoveryObserverDetectsEmptyRecovery` proves the non-entry instrument detects an actual recovery call |
| Fresh selection and generation identity | `CaptureMaterialize` selection/revalidation and `ReadObject` / `revalidateObjects` | `TestMaterializeCaptureWaitReselectsAndTimeout`, `TestMaterializeCaptureSameFragmentDrift`, `TestMaterializeCaptureAuthorizingDrift`, `TestMaterializeCaptureStoreObjectIdentity`, `TestMaterializeCaptureExternalTornGeneration`, `TestMaterializeCaptureConcurrentRepair`, `TestMaterializeCaptureSerializesProfileWriters` |
| No vendor observation in content verification | `resolveVerified(..., verifyHomeContent)` and `StatusOf` → `homeState` | `TestMaterializeCaptureAndStatusNeverProbeVendor`; a real PATH executable plant is used with nil Detect, on a successful Codex capture fixture and the status matrix |
| Bounded local capture and release before publication | `ReadObject`, `CaptureMaterialize` return/deferred release | `TestMaterializeCaptureObjectBounds`, `TestMaterializeCaptureCurrentAndRelease`, `TestMaterializeCaptureDoesNotProvision`; byte/object quotas and a capture deadline are local limits, not transport limits. The internal callback is required to perform bounded local work and use protected handles; transport/publication are outside this operation |

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

The named narrowing controls and expected-red assertions are attached separately as task-scoped mutant outcomes. No Windows shell/execute-bit/POSIX unreadability runtime coverage is claimed; the platform-case ledger records those bounds. Lock bookkeeping is compared by name, mode and size because Windows byte-range locks forbid a second reader. Source, journal, target and native credential fixture bytes remain covered by the state comparisons. M-provision-absent is killed through TestMaterializeCaptureDoesNotProvision/absent at CaptureMaterialize.

## Convergence regression

**1 of 1 convergence requirements driven**: `TestMaterializeCaptureCredentialStoreAdmission/file` and `/keychain` drive `CaptureMaterialize` → `resolveVerified` → `admitCredentialStore`, compare the exact refusal with `Resolve`, require nil output and no closure callback, and preserve source/credential fixture state. Each case first captures real content with the default native selector. `M-codex-file-admission` admits only explicit Codex file; the named `/file` capture assertion fails. The keychain refusal remains active in the mutant.

All ten narrowing plants were rerun on the converged source, tests and toolchain; no older mutation pass is reused. The mandatory M3 plant remains killed by `TestMaterializeCaptureSameFragmentDrift/aba`. No catalog surface table was supplied; this is the same brief gap accepted in revision 2. The five supplied paired rows above are fully mapped.

The separate existing `checkIsolatedCredentialStore` admission guard remains unchanged from trunk. One direct admission call appears in the shared resolve body; a successful call graph can also execute the assembly guard, and capture intentionally performs initial and final resolve passes. No whole-call-graph exactly-once claim is made.
