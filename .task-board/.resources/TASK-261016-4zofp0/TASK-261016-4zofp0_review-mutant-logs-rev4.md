# Independently rerun narrowing-plant logs

Candidate tree: `24299c092bc02d004545ff7103ffc192b541db37`. Each test command used -count=1 -timeout=2m. Generated temporary paths and fragment byte dumps are redacted; assertions and diagnostics retained.

## M-codex-file-admission

```text
--- FAIL: TestMaterializeCaptureCredentialStoreAdmission (1.95s)
    --- FAIL: TestMaterializeCaptureCredentialStoreAdmission/file (1.03s)
        materialize_capture_test.go:180: capture = &{[123 34 101 110 118 34 58 123 34 67 79 68 69 88 95 72 79 77 69 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 77 97 116 101 114 105 97 108 105 122 101 67 97 112 116 117 114 101 67 114 101 100 101 110 116 105 97 108 83 116 111 114 101 65 100 109 105 115 115 105 111 110 102 105 108 101 49 49 53 56 49 49 50 51 52 55 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 99 97 112 116 117 114 101 47 99 111 100 101 120 95 99 108 105 34 125 44 34 101 110 118 105 114 111 110 109 101 110 116 34 58 34 99 111 100 101 120 95 99 108 105 34 44 34 102 114 97 103 109 101 110 116 34 58 34 108 97 117 110 99 104 45 101 110 118 45 102 114 97 103 109 101 110 116 45 118 50 34 44 34 109 99 112 34 58 123 34 99 104 97 110 110 101 108 115 34 58 91 123 34 97 114 103 117 109 101 110 116 34 58 34 110 97 109 101 34 44 34 102 108 97 103 34 58 34 45 112 34 44 34 107 105 110 100 34 58 34 102 108 97 103 34 44 34 110 97 109 101 34 58 34 99 117 114 97 116 111 114 45 109 99 112 34 125 93 44 34 101 110 118 95 110 97 109 101 115 34 58 91 34 70 73 71 77 65 95 65 80 73 95 75 69 89 34 93 44 34 112 97 116 104 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 77 97 116 101 114 105 97 108 105 122 101 67 97 112 116 117 114 101 67 114 101 100 101 110 116 105 97 108 83 116 111 114 101 65 100 109 105 115 115 105 111 110 102 105 108 101 49 49 53 56 49 49 50 51 52 55 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 99 97 112 116 117 114 101 47 99 111 100 101 120 95 99 108 105 47 99 117 114 97 116 111 114 45 109 99 112 46 99 111 110 102 105 103 46 116 111 109 108 34 125 44 34 112 101 114 109 105 115 115 105 111 110 115 34 58 123 34 108 111 99 107 101 100 34 58 102 97 108 115 101 44 34 109 111 100 101 34 58 34 110 97 116 105 118 101 34 44 34 115 111 117 114 99 101 34 58 34 100 101 102 97 117 108 116 34 125 44 34 112 114 101 99 101 100 101 110 99 101 34 58 123 34 112 108 97 99 101 109 101 110 116 34 58 34 119 105 110 110 101 114 45 108 97 115 116 34 44 34 119 105 110 110 101 114 34 58 34 104 105 103 104 101 114 45 119 101 105 103 104 116 34 125 44 34 112 114 111 102 105 108 101 34 58 123 34 108 111 99 107 95 115 104 97 50 53 54 34 58 34 51 57 57 52 54 57 54 52 54 102 48 54 100 53 48 50 48 101 57 50 101 54 99 100 52 48 51 98 98 52 48 57 53 52 57 54 51 49 52 51 52 51 54 97 57 102 56 102 50 55 54 48 49 101 51 100 51 56 98 51 57 55 99 55 34 44 34 110 97 109 101 34 58 34 99 97 112 116 117 114 101 34 125 44 34 115 121 115 116 101 109 95 112 114 111 109 112 116 34 58 123 34 99 104 97 110 110 101 108 115 34 58 91 123 34 107 101 121 34 58 34 109 111 100 101 108 95 105 110 115 116 114 117 99 116 105 111 110 115 95 102 105 108 101 34 44 34 107 105 110 100 34 58 34 99 111 110 102 105 103 45 107 101 121 34 44 34 115 101 109 97 110 116 105 99 115 34 58 34 114 101 112 108 97 99 101 34 125 93 44 34 112 97 116 104 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 77 97 116 101 114 105 97 108 105 122 101 67 97 112 116 117 114 101 67 114 101 100 101 110 116 105 97 108 83 116 111 114 101 65 100 109 105 115 115 105 111 110 102 105 108 101 49 49 53 56 49 49 50 51 52 55 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 99 97 112 116 117 114 101 47 99 111 100 101 120 95 99 108 105 47 46 97 103 101 110 116 45 99 111 110 116 101 120 116 47 115 121 115 116 101 109 45 112 114 111 109 112 116 46 109 100 34 125 125 10] [60 33 45 45 10 99 117 114 97 116 111 114 45 114 111 111 116 45 99 111 110 116 101 120 116 45 118 50 10 114 111 111 116 58 32 99 97 112 116 117 114 101 32 49 46 48 46 48 32 115 116 97 116 101 32 115 104 97 50 53 54 58 98 57 50 53 101 55 97 100 54 52 97 99 102 98 51 53 55 53 101 101 102 51 100 101 53 99 55 49 51 57 101 49 98 52 55 98 99 55 54 101 50 51 100 51 53 55 53 52 99 53 52 52 99 101 100 51 50 98 102 100 49 102 56 53 10 109 101 109 98 101 114 58 32 99 97 112 116 117 114 101 32 49 46 48 46 48 32 115 116 97 116 101 32 115 104 97 50 53 54 58 98 57 50 53 101 55 97 100 54 52 97 99 102 98 51 53 55 53 101 101 102 51 100 101 53 99 55 49 51 57 101 49 98 52 55 98 99 55 54 101 50 51 100 51 53 55 53 52 99 53 52 52 99 101 100 51 50 98 102 100 49 102 56 53 32 119 101 105 103 104 116 32 49 48 48 10 112 114 101 99 101 100 101 110 99 101 58 32 119 105 110 110 101 114 61 104 105 103 104 101 114 45 119 101 105 103 104 116 32 112 108 97 99 101 109 101 110 116 61 119 105 110 110 101 114 45 108 97 115 116 10 108 111 99 107 58 32 115 104 97 50 53 54 58 51 57 57 52 54 57 54 52 54 102 48 54 100 53 48 50 48 101 57 50 101 54 99 100 52 48 51 98 98 52 48 57 53 52 57 54 51 49 52 51 52 51 54 97 57 102 56 102 50 55 54 48 49 101 51 100 51 56 98 51 57 55 99 55 10 103 101 110 101 114 97 116 101 100 58 32 67 117 114 97 116 111 114 32 80 114 111 116 111 99 111 108 32 101 110 118 105 114 111 110 109 101 110 116 115 32 114 101 118 105 115 105 111 110 32 49 32 40 104 116 116 112 115 58 47 47 103 105 116 104 117 98 46 99 111 109 47 114 101 108 117 120 45 119 111 114 107 115 47 99 117 114 97 116 111 114 45 115 112 101 99 41 10 110 111 116 105 99 101 58 32 103 101 110 101 114 97 116 101 100 32 102 105 108 101 59 32 100 105 114 101 99 116 32 101 100 105 116 115 32 97 114 101 32 117 110 115 117 112 112 111 114 116 101 100 32 97 110 100 32 97 114 101 32 100 101 116 101 99 116 101 100 32 97 115 32 100 114 105 102 116 59 32 117 112 100 97 116 101 32 116 104 101 32 115 111 117 114 99 101 32 112 114 111 102 105 108 101 32 114 101 112 111 115 105 116 111 114 121 32 111 114 32 105 116 115 32 99 111 109 112 111 115 101 100 32 112 114 111 102 105 108 101 115 32 105 110 115 116 101 97 100 10 45 45 62 10 10 45 45 45 10 10 35 35 32 67 111 110 116 101 120 116 58 32 99 97 112 116 117 114 101 32 49 46 48 46 48 10 10 104 101 108 108 111 10]}, <nil>, callback=true; want codex_settings_unsupported before capture
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	2.271s
FAIL
```

## M-composition-policy

```text
--- FAIL: TestMaterializeCaptureSameFragmentDrift (0.91s)
    --- FAIL: TestMaterializeCaptureSameFragmentDrift/composition-policy (0.91s)
        materialize_capture_test.go:457: capture = <non-nil output>, <nil>; want environment_materialize_capture_drift and no output
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.247s
FAIL
```

## M-corrupt-absence

```text
--- FAIL: TestMaterializeCaptureBadRecoveryEvidence (0.85s)
    --- FAIL: TestMaterializeCaptureBadRecoveryEvidence/corrupt (0.85s)
        materialize_capture_test.go:292: capture = <non-nil output>, <nil>; want invalid transaction journal and no output
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.197s
FAIL
```

## M-denied-class

```text
--- FAIL: TestMaterializeCaptureUnreadableJournal (0.82s)
    materialize_capture_test.go:824: capture = <nil>, open <temporary-path> permission denied; want manager_state_unreadable and no output
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.147s
FAIL
```

## M-executable-mode

```text
--- FAIL: TestMaterializeCaptureSameFragmentDrift (0.69s)
    --- FAIL: TestMaterializeCaptureSameFragmentDrift/mode (0.69s)
        materialize_capture_test.go:457: capture = <non-nil output>, <nil>; want environment_materialize_capture_drift and no output
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.023s
FAIL
```

## M-lock-name

```text
--- FAIL: TestMaterializeCaptureAuthorizingDrift (0.95s)
    --- FAIL: TestMaterializeCaptureAuthorizingDrift/lock-identity (0.95s)
        materialize_capture_test.go:792: capture = <non-nil output>, <nil>; want environment_materialize_capture_drift and no output
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.268s
FAIL
```

## M-policy-probe

```text
--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor (1.62s)
    materialize_capture_test.go:498: vendor executable ran
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.963s
FAIL
```

## M-prepared

```text
--- FAIL: TestMaterializeCapturePendingDoesNotRecover (0.89s)
    --- FAIL: TestMaterializeCapturePendingDoesNotRecover/prepared (0.89s)
        materialize_capture_test.go:246: capture = <non-nil output>, <nil>; want environment_materialize_recovery_required and no output
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.215s
FAIL
```

## M-probe

```text
--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor (1.40s)
    materialize_capture_test.go:498: vendor executable ran
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.740s
FAIL
```

## M-provision-absent

```text
--- FAIL: TestMaterializeCaptureDoesNotProvision (0.01s)
    --- FAIL: TestMaterializeCaptureDoesNotProvision/absent (0.00s)
        materialize_capture_test.go:515: selected unprovisioned source
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	0.341s
FAIL
```

## M-release-cache

```text
--- FAIL: TestMaterializeCaptureCredentialPolicyReleaseObservation (1.03s)
    materialize_capture_test.go:908: invalid native selector reached capture
--- FAIL: TestMaterializeCaptureCredentialPolicyReleaseCache (0.45s)
    materialize_capture_test.go:943: release unknown: policy=&{selection:native isolation:shared nativeStore: links:map[] strategies:map[] records:[]} err=<nil>; want fresh isolated
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.807s
FAIL
```

## M3-lock-only

```text
--- FAIL: TestMaterializeCaptureSameFragmentDrift (0.85s)
    --- FAIL: TestMaterializeCaptureSameFragmentDrift/aba (0.85s)
        materialize_capture_test.go:457: capture = <non-nil output>, <nil>; want environment_materialize_capture_drift and no output
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.175s
FAIL
```