# TASK-261016-wh6v9a — revision-3 bounded validation evidence

This artifact records the identity comparison, local no-probe test and the
completed Ubuntu Test job. The whole hosted gate verdict is recorded separately
in the revision-3 handoff and gate resources; no whole-gate verdict is inferred
from this subset.

Fresh command: `go test ./internal/envprofile -run '^TestMaterializeCaptureAndStatusNeverProbeVendor$' -count=1 -v -timeout=3m`; real exit **0**.

Fresh evidence download: `gh run download 39010003372 --name test-evidence-ubuntu-latest --dir .temp/TASK-261016-wh6v9a/revision-3/ubuntu-evidence`; real exit **0**.

Hosted source: https://github.com/relux-works/curator/actions/runs/39010003372

The Ubuntu Test job runs `bash .github/ci/test-gate.sh .temp/ci-evidence/test`.
Its uploaded go-test.json contains explicit pass events for both previously
failing review tests, the complete no-probe test and both committed narrowing
harnesses. All sixteen expected-red controls have named failing tests and
actual child exit 1. Parent harness pass is not reported as a green child.

Local execution platform: Go 1.25.5, darwin/arm64. No local whole-suite,
standalone build, standalone lint or separate SNAP-LOCK overlay-harness replay
is claimed in revision 3. Historical executions remain in revision-2 resources;
current full hosted validation is reported by its own terminal resource.

## identity-01.json

```text
{
  "revision": 3,
  "base": "a17e58156ac30a1629e3ff36efa259bab1555a15",
  "revision2_tree": "9bad791482638be26d3ab306653fd279791f449b",
  "clean_convergence_tree": "4f88746642b992950ba314473747e321218eba00",
  "different_from_clean_convergence": [
    "LOGBOOK.md"
  ],
  "logbook_change": "exact original task blocks moved to top, newest first",
  "changed_product_files_in_this_run": [],
  "remote_gate_delta": "",
  "toolchain": "go version go1.25.5 darwin/arm64",
  "gate_snapshot": "809e3154f350986be78944503a9182a1d5d09940",
  "candidate_tree": "da51dd68abb173fc894d8369e28662d6a17886d7",
  "signature": "G",
  "index_tree": "9637ce7e34e9aa74fbd467cd3433a2863823d578"
}
```

## tests-retained-01.json

```text
{
  "revision2_test_functions_retained": 3637,
  "missing": [],
  "new_test_functions_vs_revision2": [
    [
      "internal/scriptworker/landlock_probe_linux_test.go",
      "TestLandlockProbeCancellationKillsProcessGroup"
    ],
    [
      "internal/scriptworker/landlock_probe_linux_test.go",
      "TestLandlockProbeChildCannotReprobe"
    ],
    [
      "internal/scriptworker/landlock_probe_linux_test.go",
      "TestLandlockProbeNarrowingMutants"
    ],
    [
      "internal/scriptworker/landlock_probe_other_test.go",
      "TestLandlockProbeCancellationKillsProcessGroup"
    ],
    [
      "internal/scriptworker/probe_reexec_test.go",
      "TestLandlockProbeModeTerminatesAffectedTestBinaries"
    ]
  ]
}
```

## no-probe-01.log

```text
=== RUN   TestMaterializeCaptureAndStatusNeverProbeVendor
=== RUN   TestMaterializeCaptureAndStatusNeverProbeVendor/codex
=== RUN   TestMaterializeCaptureAndStatusNeverProbeVendor/codex/export
=== RUN   TestMaterializeCaptureAndStatusNeverProbeVendor/codex/manifest-only
=== RUN   TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared
=== RUN   TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared/export
=== RUN   TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared/manifest-only
=== RUN   TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared
=== RUN   TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared/export
=== RUN   TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared/manifest-only
=== RUN   TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated
=== RUN   TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated/export
=== RUN   TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated/manifest-only
=== RUN   TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated
=== RUN   TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated/export
=== RUN   TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated/manifest-only
--- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor (5.12s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/codex (0.84s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/codex/export (0.14s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/codex/manifest-only (0.17s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared (1.07s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared/export (0.14s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared/manifest-only (0.15s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared (1.10s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared/export (0.16s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared/manifest-only (0.14s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated (1.16s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated/export (0.16s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated/manifest-only (0.15s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated (0.95s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated/export (0.16s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated/manifest-only (0.15s)
PASS
ok  	github.com/relux-works/curator/internal/envprofile	5.479s
```

## ubuntu-named-results-01.json

```text
[
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/internal/envprofile",
    "Test": "TestMaterializeCaptureAndStatusNeverProbeVendor/codex/export",
    "Elapsed": 0.02
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/internal/envprofile",
    "Test": "TestMaterializeCaptureAndStatusNeverProbeVendor/codex/manifest-only",
    "Elapsed": 0.02
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/internal/envprofile",
    "Test": "TestMaterializeCaptureAndStatusNeverProbeVendor/codex",
    "Elapsed": 0.12
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/internal/envprofile",
    "Test": "TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared/export",
    "Elapsed": 0.03
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/internal/envprofile",
    "Test": "TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared/manifest-only",
    "Elapsed": 0.03
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/internal/envprofile",
    "Test": "TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared",
    "Elapsed": 0.16
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/internal/envprofile",
    "Test": "TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared/export",
    "Elapsed": 0.03
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/internal/envprofile",
    "Test": "TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared/manifest-only",
    "Elapsed": 0.03
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/internal/envprofile",
    "Test": "TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared",
    "Elapsed": 0.16
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/internal/envprofile",
    "Test": "TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated/export",
    "Elapsed": 0.02
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/internal/envprofile",
    "Test": "TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated/manifest-only",
    "Elapsed": 0.03
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/internal/envprofile",
    "Test": "TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated",
    "Elapsed": 0.14
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/internal/envprofile",
    "Test": "TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated/export",
    "Elapsed": 0.02
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/internal/envprofile",
    "Test": "TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated/manifest-only",
    "Elapsed": 0.02
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/internal/envprofile",
    "Test": "TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated",
    "Elapsed": 0.16
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/internal/envprofile",
    "Test": "TestMaterializeCaptureAndStatusNeverProbeVendor",
    "Elapsed": 0.74
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/internal/install",
    "Test": "TestAuthoritativeDryRunCasesMutateNothingPersistent/compiled-cache-miss-is-read-only",
    "Elapsed": 4.36
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/cmd/curator",
    "Test": "TestReviewProbeLegacyLaneIdentityInvalid",
    "Elapsed": 3.11
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/cmd/curator",
    "Test": "TestMuseCredentialPolicyNarrowingMutants/fragment-reobserves-muse-release",
    "Elapsed": 2.31
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/cmd/curator",
    "Test": "TestMuseCredentialPolicyNarrowingMutants/muse-keychain-admitted",
    "Elapsed": 2.74
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/cmd/curator",
    "Test": "TestMuseCredentialPolicyNarrowingMutants/missing-shared-source-admitted",
    "Elapsed": 2.04
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/cmd/curator",
    "Test": "TestMuseCredentialPolicyNarrowingMutants/isolated-muse-native-link",
    "Elapsed": 2.12
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/cmd/curator",
    "Test": "TestMuseCredentialPolicyNarrowingMutants/renderer-drops-only-backend",
    "Elapsed": 2.35
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/cmd/curator",
    "Test": "TestMuseCredentialPolicyNarrowingMutants/header-only-root-fact-dropped",
    "Elapsed": 1.53
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/cmd/curator",
    "Test": "TestMuseCredentialPolicyNarrowingMutants/boundary-admits-HOME",
    "Elapsed": 0.14
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/cmd/curator",
    "Test": "TestMuseCredentialPolicyNarrowingMutants/boundary-admits-missing-literal",
    "Elapsed": 0.19
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/cmd/curator",
    "Test": "TestMuseCredentialPolicyNarrowingMutants/boundary-admits-false-root",
    "Elapsed": 0.19
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/cmd/curator",
    "Test": "TestMuseCredentialPolicyNarrowingMutants",
    "Elapsed": 13.63
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/cmd/curator",
    "Test": "TestMuseRootContextNarrowingMutants/configured-form-ignored",
    "Elapsed": 1.36
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/cmd/curator",
    "Test": "TestMuseRootContextNarrowingMutants/referenced-fallback",
    "Elapsed": 1.72
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/cmd/curator",
    "Test": "TestMuseRootContextNarrowingMutants/no-context-admitted",
    "Elapsed": 1.55
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/cmd/curator",
    "Test": "TestMuseRootContextNarrowingMutants/wrong-target",
    "Elapsed": 1.64
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/cmd/curator",
    "Test": "TestMuseRootContextNarrowingMutants/wrong-bytes",
    "Elapsed": 1.56
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/cmd/curator",
    "Test": "TestMuseRootContextNarrowingMutants/changed-root-admitted",
    "Elapsed": 1.52
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/cmd/curator",
    "Test": "TestMuseRootContextNarrowingMutants/removed-root-admitted",
    "Elapsed": 1.5
  },
  {
    "Action": "pass",
    "Package": "github.com/relux-works/curator/cmd/curator",
    "Test": "TestMuseRootContextNarrowingMutants",
    "Elapsed": 10.85
  }
]
```

## hosted-muse-control-exits-01.log

```text
muse_credential_mutants_test.go:77: fragment-reobserves-muse-release: expected-red go test exit 1; killed by TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared
muse_credential_mutants_test.go:77: muse-keychain-admitted: expected-red go test exit 1; killed by TestMuseCredentialPolicyMatrix/keychain/shared
muse_credential_mutants_test.go:77: missing-shared-source-admitted: expected-red go test exit 1; killed by TestMuseMissingNativeFilePolicy/shared
muse_credential_mutants_test.go:77: isolated-muse-native-link: expected-red go test exit 1; killed by TestMuseCredentialPolicyMatrix/file/isolated
muse_credential_mutants_test.go:77: renderer-drops-only-backend: expected-red go test exit 1; killed by TestMuseCLIStorePolicy
muse_credential_mutants_test.go:77: header-only-root-fact-dropped: expected-red go test exit 1; killed by TestMuseRootContextVectors/zero-applicable
muse_credential_mutants_test.go:77: boundary-admits-HOME: expected-red go test exit 1; killed by TestMuseFragmentBoundary/HOME
muse_credential_mutants_test.go:77: boundary-admits-missing-literal: expected-red go test exit 1; killed by TestMuseFragmentBoundary/missing-literal
muse_credential_mutants_test.go:77: boundary-admits-false-root: expected-red go test exit 1; killed by TestMuseFragmentBoundary/false-root
muse_root_context_test.go:291: configured-form-ignored: expected-red go test exit 1; killed by TestMuseRootContextVectors/referenced-refused
muse_root_context_test.go:291: referenced-fallback: expected-red go test exit 1; killed by TestMuseRootContextVectors/referenced-refused
muse_root_context_test.go:291: no-context-admitted: expected-red go test exit 1; killed by TestMuseRootContextVectors/no-context
muse_root_context_test.go:291: wrong-target: expected-red go test exit 1; killed by TestMuseRootContextVectors/monolithic
muse_root_context_test.go:291: wrong-bytes: expected-red go test exit 1; killed by TestMuseRootContextVectors/zero-applicable
muse_root_context_test.go:291: changed-root-admitted: expected-red go test exit 1; killed by TestMuseRootContextDrift/changed
muse_root_context_test.go:291: removed-root-admitted: expected-red go test exit 1; killed by TestMuseRootContextDrift/removed
```

Ubuntu source stream SHA-256: `83f7d8ec8bee56452314d4d3a162a12415565c56d4b492b2ff76908cf230505f`.
