# SNAP-LOCK revision 2 review sweep

Candidate tree: f0b08b96c097fcf57c1927355f9874079df23e5b.
Coverage map checked: 5/5 capture-stage rows named; 0/2 CLI forms, explicitly deferred. No supplied catalog table or wall-clock budget; use the five mandatory paired rows, 25 minute review budget and 3 minute free hunt.

| R1 | Current source / pending journal | held | `go test ./internal/envprofile ./internal/transaction ./internal/managerlock -run Test(MaterializeCapture(CurrentAndRelease|PendingDoesNotRecover|BadRecoveryEvidence|UnreadableJournal|DoesNotProvision)|PendingRecovery|RecoveryObserver|AcquireExistingHomeOnly) -count=1 -timeout=2m`; exit 0; R1.log |

| R2 | Released lock / timeout and stale selection | held | `go test ./internal/envprofile ./internal/transaction ./internal/managerlock -run Test(MaterializeCapture(WaitReselectsAndTimeout|AuthorizingDrift)|CaptureLock) -count=1 -timeout=2m`; exit 0; R2.log |

| R3 | Cooperating writers / multi-file generation | held | `go test ./internal/envprofile ./internal/transaction ./internal/managerlock -run TestMaterializeCapture(SerializesProfileWriters|ConcurrentRepair|ExternalTornGeneration) -count=1 -timeout=2m`; exit 0; R3.log |

| R4 | Stable objects / same-fragment ABA and store changes | held | `go test ./internal/envprofile ./internal/transaction ./internal/managerlock -run TestMaterializeCapture(SameFragmentDrift|StoreObjectIdentity|AuthorizingDrift|ObjectBounds) -count=1 -timeout=2m`; exit 0; R4.log |

| R5 | File metadata verification / vendor executable | held | `go test ./internal/envprofile ./internal/transaction ./internal/managerlock -run TestMaterializeCaptureAndStatusNeverProbeVendor -count=1 -timeout=2m`; exit 0; R5.log |

Free hunt: TestReviewCaptureLateJournal and TestReviewCaptureRejectsMutationFlags through CaptureMaterialize, isolated exact candidate plus reviewer-only tests; exit 0. Late corrupt journal and all three mutation flags refused with expected diagnostics and unchanged fixture state. No blocking finding.

M3: isolated candidate narrowed revalidateObjects to profile-lock object only. go test ./internal/envprofile -run ^TestMaterializeCaptureSameFragmentDrift/aba$ -count=1 -timeout=2m exited 1 at the intended drift assertion; restored original bytes. Other eight producer narrowing plants reused by unchanged blob identity, not independently rerun.
