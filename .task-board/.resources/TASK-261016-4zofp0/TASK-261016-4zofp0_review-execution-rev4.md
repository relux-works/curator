# Revision 4 independent review execution

Candidate tree: `24299c092bc02d004545ff7103ffc192b541db37`. Environment: Go 1.25.5 darwin/arm64. All 17 changed working blobs matched before tests.

AC coverage inspected before implementation review: 5 of 5 capture-stage rows; 1 of 1 credential-store convergence row; 2 of 2 release-policy convergence properties. 0 of 2 CLI forms qualified. No separate catalog rows or numeric review/free-hunt budget were supplied; the five paired rows are the sweep floor.

## R1 — held

Command: `go test ./internal/envprofile ./internal/transaction ./internal/managerlock -run Test(MaterializeCaptureCurrentAndRelease|MaterializeCapturePendingDoesNotRecover|MaterializeCaptureBadRecoveryEvidence|MaterializeCaptureUnreadableJournal|MaterializeCaptureDoesNotProvision|PendingRecovery|RecoveryObserver|AcquireExistingHomeOnly) -count=1 -timeout=3m`. Exit 0.

```text
ok  	github.com/relux-works/curator/internal/envprofile	7.852s
ok  	github.com/relux-works/curator/internal/transaction	0.687s
ok  	github.com/relux-works/curator/internal/managerlock	0.171s
```

## R2 — held

Command: `go test ./internal/envprofile ./internal/transaction ./internal/managerlock -run Test(MaterializeCaptureWaitReselectsAndTimeout|CaptureLock) -count=1 -timeout=3m`. Exit 0.

```text
ok  	github.com/relux-works/curator/internal/envprofile	1.174s
ok  	github.com/relux-works/curator/internal/transaction	0.159s [no tests to run]
ok  	github.com/relux-works/curator/internal/managerlock	0.177s
```

## R3 — held

Command: `go test ./internal/envprofile ./internal/transaction ./internal/managerlock -run Test(MaterializeCaptureSerializesProfileWriters|MaterializeCaptureConcurrentRepair|MaterializeCaptureExternalTornGeneration) -count=1 -timeout=3m`. Exit 0.

```text
ok  	github.com/relux-works/curator/internal/envprofile	4.652s
ok  	github.com/relux-works/curator/internal/transaction	0.167s [no tests to run]
ok  	github.com/relux-works/curator/internal/managerlock	0.158s [no tests to run]
```

## R4 — held

Command: `go test ./internal/envprofile ./internal/transaction ./internal/managerlock -run Test(MaterializeCaptureSameFragmentDrift|MaterializeCaptureStoreObjectIdentity|MaterializeCaptureAuthorizingDrift|MaterializeCaptureObjectBounds) -count=1 -timeout=3m`. Exit 0.

```text
ok  	github.com/relux-works/curator/internal/envprofile	15.301s
ok  	github.com/relux-works/curator/internal/transaction	0.161s [no tests to run]
ok  	github.com/relux-works/curator/internal/managerlock	0.149s [no tests to run]
```

## R5 — held

Command: `go test ./internal/envprofile ./internal/transaction ./internal/managerlock -run Test(MaterializeCaptureAndStatusNeverProbeVendor|MaterializeCaptureCredential) -count=1 -timeout=3m`. Exit 0.

```text
ok  	github.com/relux-works/curator/internal/envprofile	4.224s
ok  	github.com/relux-works/curator/internal/transaction	0.166s [no tests to run]
ok  	github.com/relux-works/curator/internal/managerlock	0.154s [no tests to run]
```

## Narrowing plants

All twelve plants rerun independently in an isolated copy of the verified candidate. Harness exit 0; each behavioral command exits 1 at its named intended assertion. No product files modified.

```json
[
  {
    "mutant": "M-policy-probe",
    "test": "^TestMaterializeCaptureAndStatusNeverProbeVendor$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor (1.62s)"
    ]
  },
  {
    "mutant": "M-release-cache",
    "test": "^TestMaterializeCaptureCredentialPolicyRelease(Observation|Cache)$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCaptureCredentialPolicyReleaseObservation (1.03s)",
      "--- FAIL: TestMaterializeCaptureCredentialPolicyReleaseCache (0.45s)"
    ]
  },
  {
    "mutant": "M-codex-file-admission",
    "test": "^TestMaterializeCaptureCredentialStoreAdmission$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCaptureCredentialStoreAdmission (1.95s)",
      "--- FAIL: TestMaterializeCaptureCredentialStoreAdmission/file (1.03s)"
    ]
  },
  {
    "mutant": "M-denied-class",
    "test": "^TestMaterializeCaptureUnreadableJournal$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCaptureUnreadableJournal (0.82s)"
    ]
  },
  {
    "mutant": "M-lock-name",
    "test": "^TestMaterializeCaptureAuthorizingDrift/lock-identity$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCaptureAuthorizingDrift (0.95s)",
      "--- FAIL: TestMaterializeCaptureAuthorizingDrift/lock-identity (0.95s)"
    ]
  },
  {
    "mutant": "M-composition-policy",
    "test": "^TestMaterializeCaptureSameFragmentDrift/composition-policy$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCaptureSameFragmentDrift (0.91s)",
      "--- FAIL: TestMaterializeCaptureSameFragmentDrift/composition-policy (0.91s)"
    ]
  },
  {
    "mutant": "M3-lock-only",
    "test": "^TestMaterializeCaptureSameFragmentDrift/aba$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCaptureSameFragmentDrift (0.85s)",
      "--- FAIL: TestMaterializeCaptureSameFragmentDrift/aba (0.85s)"
    ]
  },
  {
    "mutant": "M-prepared",
    "test": "^TestMaterializeCapturePendingDoesNotRecover/prepared$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCapturePendingDoesNotRecover (0.89s)",
      "--- FAIL: TestMaterializeCapturePendingDoesNotRecover/prepared (0.89s)"
    ]
  },
  {
    "mutant": "M-corrupt-absence",
    "test": "^TestMaterializeCaptureBadRecoveryEvidence/corrupt$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCaptureBadRecoveryEvidence (0.85s)",
      "--- FAIL: TestMaterializeCaptureBadRecoveryEvidence/corrupt (0.85s)"
    ]
  },
  {
    "mutant": "M-probe",
    "test": "^TestMaterializeCaptureAndStatusNeverProbeVendor$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor (1.40s)"
    ]
  },
  {
    "mutant": "M-executable-mode",
    "test": "^TestMaterializeCaptureSameFragmentDrift/mode$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCaptureSameFragmentDrift (0.69s)",
      "--- FAIL: TestMaterializeCaptureSameFragmentDrift/mode (0.69s)"
    ]
  },
  {
    "mutant": "M-provision-absent",
    "test": "^TestMaterializeCaptureDoesNotProvision$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCaptureDoesNotProvision (0.01s)",
      "--- FAIL: TestMaterializeCaptureDoesNotProvision/absent (0.00s)"
    ]
  }
]
```

## Supplemental attacks

`go test -overlay .temp/review-r4/overlay.json ./internal/envprofile -run ^TestReviewCapture -count=1 -timeout=2m`: exit 0. Replayed prior reviewer late-corrupt-journal and mutation-flag attacks with valid controls; no finding.

```text
ok  	github.com/relux-works/curator/internal/envprofile	3.219s
```

## Normative read

Read curator-spec main `1059b94b4135089d842bf017070b8ab47d79a536` via hosting API: snapshots sections 1/5, environments 9.2/10.1, CIP-0014 consistency steps 1–5 and CIP-0015 machine selector. No contradiction identified in the capture-stage delta.

## Prior verdict correction

Revision 2 verdict is accepted, findings empty. The generic brief describing a prior rejection is inaccurate; no finding class to repeat.

## Additional revision-4 attack and trunk regressions

`go test -overlay .temp/review-r4/overlay.json ./internal/envprofile -run ^TestReviewCaptureLateNativePolicy$ -count=1 -timeout=2m`: exit 0. Changes native credential metadata during the capture callback; final verification refuses with environment_materialize_capture_drift, nil output and unchanged planted state, beside a successful capture.

```text
ok  	github.com/relux-works/curator/internal/envprofile	1.065s
```

`go test ./internal/envprofile -run "Test(CredentialProjection|CredentialStorePolicy)" -count=1 -timeout=3m`: exit 0.

```text
ok  	github.com/relux-works/curator/internal/envprofile	8.335s
```

All 17 changed isolated-copy blobs matched the candidate tree after mutant restoration. All working candidate blobs remain unchanged.
