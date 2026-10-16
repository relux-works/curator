# SNAP-LOCK capture-stage coverage map

Candidate: `3bdf4c4f77a89afc12c06c242f69cd3d16deda9e`.

Coverage: **5 of 5 AC rows driven at the production capture-stage entry point**. **0 of 2 CLI forms qualified here**: the actual export and manifest-only commands belong to SNAP-CLI, which must integrate this capture path. Test labels for the two consumers exercise the same capture-stage operation, not an already shipped CLI. No vendor-runtime, archive, manifest or destination publication qualification is claimed.

| AC row | Production call site | Named tests and stated limits |
|---|---|---|
| Non-recovering bounded lock | `CaptureMaterialize` \u2192 `beginCaptureOperation` \u2192 `AcquireExistingHomeOnly` | `TestMaterializeCaptureCurrentAndRelease`, `TestMaterializeCaptureWaitReselectsAndTimeout`, `TestCaptureLockSharesWriterLock`, `TestAcquireExistingHomeOnlyDoesNotProvision` |
| Pending/read-failure refusal | `beginCaptureOperation` \u2192 `PendingRecovery` | `TestMaterializeCapturePendingDoesNotRecover`, `TestMaterializeCaptureBadRecoveryEvidence`, `TestMaterializeCaptureUnreadableJournal`, `TestPendingRecoveryReadOnlyPhasesAndTombs`, `TestPendingRecoveryRejectsLinkedInventory`; `TestRecoveryObserverDetectsEmptyRecovery` proves the non-entry instrument detects an actual recovery call |
| Fresh selection and generation identity | `CaptureMaterialize` selection/revalidation and `ReadObject` / `revalidateObjects` | `TestMaterializeCaptureWaitReselectsAndTimeout`, `TestMaterializeCaptureSameFragmentDrift`, `TestMaterializeCaptureAuthorizingDrift`, `TestMaterializeCaptureStoreObjectIdentity`, `TestMaterializeCaptureExternalTornGeneration`, `TestMaterializeCaptureConcurrentRepair`, `TestMaterializeCaptureSerializesProfileWriters` |
| No vendor observation in content verification | `resolveVerified(..., verifyHomeContent)` and `StatusOf` \u2192 `homeState` | `TestMaterializeCaptureAndStatusNeverProbeVendor`; a real PATH executable plant is used with nil Detect, on a successful Codex capture fixture and the status matrix |
| Bounded local capture and release before publication | `ReadObject`, `CaptureMaterialize` return/deferred release | `TestMaterializeCaptureObjectBounds`, `TestMaterializeCaptureCurrentAndRelease`, `TestMaterializeCaptureDoesNotProvision`; byte/object quotas and a capture deadline are local limits, not transport limits. The internal callback is required to perform bounded local work and use protected handles; transport/publication are outside this operation |

## Paired behavior coverage map

Brief gap: no catalog surface table was supplied. This map covers the five paired behavior rows actually supplied; it does not silently invent a catalog selection.

| Supplied surface row | Positive and negative tests | Killed narrowing mutants | Out-of-contract inputs / acceptance clause |
|---|---|---|---|
| Current source / pending journal | `TestMaterializeCaptureCurrentAndRelease`; `TestMaterializeCapturePendingDoesNotRecover`, `TestMaterializeCaptureBadRecoveryEvidence`, `TestMaterializeCaptureUnreadableJournal` | M-prepared, M-corrupt-absence, M-denied-class, M-provision-absent | None within capture-stage journal semantics |
| Released lock / timeout and stale selection | `TestMaterializeCaptureWaitReselectsAndTimeout`; `TestMaterializeCaptureAuthorizingDrift` | M-lock-name, M-composition-policy | Actual CLI selection/loading adapter remains SNAP-CLI: leaf says \u201cSNAP-CLI integrates both\u201d |
| Cooperating writers / multi-file generation | `TestMaterializeCaptureSerializesProfileWriters`, `TestMaterializeCaptureConcurrentRepair`, `TestMaterializeCaptureExternalTornGeneration`; neighboring same-fragment witness in `TestMaterializeCaptureSameFragmentDrift/aba` | M3-lock-only | Global/profile scope sync and real repair/use are driven; no external service is waited on under the lock |
| Stable objects / same-fragment ABA and store changes | `TestMaterializeCaptureSameFragmentDrift`, `TestMaterializeCaptureStoreObjectIdentity`, `TestMaterializeCaptureAuthorizingDrift` | M3-lock-only, M-executable-mode, M-lock-name | Closure inventory and protected handle acquisition belong to SNAP-CLOSURE, as the leaf explicitly states |
| File/metadata verification / vendor executable | `TestMaterializeCaptureAndStatusNeverProbeVendor`; owned source/native fixture state comparisons in capture tests | M-probe | Actual manifest-only/export publication and vendor-runtime activation remain SNAP-CLI / later runtime qualification |

Negative cases require the intended refusal and no successful output. Source-generation matrices compare the post-refusal source, native credential roots and owned fixture state with the deliberately planted state; pending journals and transaction targets are preserved. Credential links are never dereferenced by fixture-state snapshots.

The named narrowing controls and expected-red assertions are attached separately as task-scoped mutant outcomes. No Windows shell/execute-bit/POSIX unreadability runtime coverage is claimed; the platform-case ledger records those bounds. Lock bookkeeping is compared by name, mode and size because Windows byte-range locks forbid a second reader. Source, journal, target and native credential fixture bytes remain covered by the state comparisons. M-provision-absent is killed through TestMaterializeCaptureDoesNotProvision/absent at CaptureMaterialize.

## Convergence regression

**1 of 1 convergence requirements driven**: `TestMaterializeCaptureCredentialStoreAdmission/file` and `/keychain` drive `CaptureMaterialize` \u2192 `resolveVerified` \u2192 `admitCredentialStore`, compare the exact refusal with `Resolve`, require nil output and no closure callback, and preserve source/credential fixture state. Each case first captures real content with the default native selector. `M-codex-file-admission` admits only explicit Codex file; the named `/file` capture assertion fails. The keychain refusal remains active in the mutant.

All twelve narrowing plants are replayed for revision 4; their measured exits and named failures are in the mutation outcome. No older green result is reused. The mandatory M3 same-fragment witness remains part of that run.

## L15-1 release-observation convergence

**2 of 2 additional convergence properties driven:**

1. `CaptureMaterialize` and `StatusOf` use the content-only verification boundary across credential-policy derivation: `TestMaterializeCaptureAndStatusNeverProbeVendor` drives both real entry points with nil Detect and a PATH executable sentinel. `M-policy-probe` narrows the supplied-release guard to all adapters except Codex; the successful capture/status test fails because the executable runs. `M-probe` independently attacks direct content verification while preserving the unknown-release token in source.
2. Release-keyed cache separation: `TestMaterializeCaptureCredentialPolicyReleaseObservation` drives `CaptureMaterialize` with a prior known-release projection and an invalid native selector; it requires the intended refusal, no output/callback and unchanged fixture state. `TestMaterializeCaptureCredentialPolicyReleaseCache` additionally drives the shared production policy function across below-pinned, unknown and pinned observations, including an earlier refusal. Its macOS Claude isolation distinction is platform-specific; other platforms check projection identity and same-observation reuse. `M-release-cache` admits unknown-release reuse of a different observation and both tests fail.

There is one derivation implementation, `computeCredentialPolicy(adapter, release)`. Supplied observations flow through `resolveCredentialPolicyObserved`; the ordinary wrapper discovers a release only when no scoped observation exists. `verifyHomeObserved` scopes the supplied release across assembly and passthrough checks; nested assembly restores the previous observation. Cache success and error entries include profile, adapter and release. Explicit unavailable credential stores pass through the same native-inspection/admission path before any capture callback. Capture intentionally performs verification before and after the callback, so this is not a whole-call-graph exactly-once claim.

Out-of-contract rows: full closure inventory and protected-handle discovery (leaf dependency: SNAP-CLOSURE supplies the callback); actual export and manifest-only CLI publication/destination handling (leaf dependency: SNAP-CLI integrates both); vendor-runtime activation and archive transport (not part of the capture transaction). Windows POSIX sentinel and chmod/unreadability/held-object-replacement variants retain their explicit platform skips; no execution claim is made for skipped tests. Capture quotas/deadline apply to a cooperating bounded local callback, not arbitrary callback preemption.
