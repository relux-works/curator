# Narrowing mutant behavioral logs

Candidate: `1e6be490fee5768d0aa6808b825a20437058ef1c`. Harness exit 0; every planted Go test exit 1.

## M-denied-class

```text
--- FAIL: TestMaterializeCaptureUnreadableJournal (0.55s)
    materialize_capture_test.go:776: capture = <nil>, open <temporary-fixture-path>: permission denied; want manager_state_unreadable and no output
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.000s
FAIL
```

## M-lock-name

```text
--- FAIL: TestMaterializeCaptureAuthorizingDrift (0.61s)
    --- FAIL: TestMaterializeCaptureAuthorizingDrift/lock-identity (0.61s)
        materialize_capture_test.go:744: capture = <non-nil capture; fixture byte arrays redacted>, <nil>; want environment_materialize_capture_drift and no output
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	0.936s
FAIL
```

## M-composition-policy

```text
--- FAIL: TestMaterializeCaptureSameFragmentDrift (0.61s)
    --- FAIL: TestMaterializeCaptureSameFragmentDrift/composition-policy (0.61s)
        materialize_capture_test.go:409: capture = <non-nil capture; fixture byte arrays redacted>, <nil>; want environment_materialize_capture_drift and no output
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	0.937s
FAIL
```

## M3-lock-only

```text
--- FAIL: TestMaterializeCaptureSameFragmentDrift (0.69s)
    --- FAIL: TestMaterializeCaptureSameFragmentDrift/aba (0.69s)
        materialize_capture_test.go:409: capture = <non-nil capture; fixture byte arrays redacted>, <nil>; want environment_materialize_capture_drift and no output
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.012s
FAIL
```

## M-prepared

```text
--- FAIL: TestMaterializeCapturePendingDoesNotRecover (0.71s)
    --- FAIL: TestMaterializeCapturePendingDoesNotRecover/prepared (0.71s)
        materialize_capture_test.go:198: capture = <non-nil capture; fixture byte arrays redacted>, <nil>; want environment_materialize_recovery_required and no output
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.038s
FAIL
```

## M-corrupt-absence

```text
--- FAIL: TestMaterializeCaptureBadRecoveryEvidence (0.62s)
    --- FAIL: TestMaterializeCaptureBadRecoveryEvidence/corrupt (0.62s)
        materialize_capture_test.go:244: capture = <non-nil capture; fixture byte arrays redacted>, <nil>; want invalid transaction journal and no output
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	0.947s
FAIL
```

## M-probe

```text
--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor (1.00s)
    materialize_capture_test.go:450: vendor executable ran
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.338s
FAIL
```

## M-executable-mode

```text
--- FAIL: TestMaterializeCaptureSameFragmentDrift (0.62s)
    --- FAIL: TestMaterializeCaptureSameFragmentDrift/mode (0.62s)
        materialize_capture_test.go:409: capture = <non-nil capture; fixture byte arrays redacted>, <nil>; want environment_materialize_capture_drift and no output
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	0.941s
FAIL
```

## M-provision-absent

```text
--- FAIL: TestMaterializeCaptureDoesNotProvision (0.00s)
    --- FAIL: TestMaterializeCaptureDoesNotProvision/absent (0.00s)
        materialize_capture_test.go:467: selected unprovisioned source
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	0.328s
FAIL
```
