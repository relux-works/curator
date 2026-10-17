# TASK-261016-wh6v9a \u2014 credential-store-policy-successors revision-2 execution evidence
Local Go 1.25.5 darwin/arm64. Red logs are expected failures or the pre-fix regression, never green gate evidence.

## identity.json

```text
{
  "head": "612960ee90bedc7a4a113eda13aaa6cf2a2cbe9f",
  "gate_snapshot": "178649841e0af556ff0827e8dded65194fb34ebc",
  "tree": "9bad791482638be26d3ab306653fd279791f449b",
  "retained_test_functions": 3609,
  "missing_test_functions": [],
  "index_unchanged": true,
  "remote_gate_unchanged": true
}
```

## accepted-to-combined.json

```text
[
  {
    "path": "LOGBOOK.md",
    "merge_exit": 0,
    "equal": false,
    "result": "revision-2 convergence regression and trunk gate-fix attribution"
  },
  {
    "path": "README.md",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "cmd/curator/muse_credential_mutants_test.go",
    "merge_exit": 0,
    "equal": false,
    "result": "new Muse-only reobservation mutant; Unix sentinel bound"
  },
  {
    "path": "cmd/curator/muse_root_context_test.go",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "cmd/curator/muse_test.go",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "docs/environment-config.md",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/envfragment.go",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/muse_test.go",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/README.md",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/schemas/v1/launch-env-fragment-v1.schema.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/schemas/v1/launch-env-fragment-v2.schema.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/schemas/v1/launch-env-fragment-v4.schema.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/index.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-HOME.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-additions-on-v1.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-additions-on-v2.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-additions-on-v3.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-backend-absolute.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-backend-empty.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-backend-keychain.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-claude_code-backend.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-claude_code-file.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-claude_code-root.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-codex-file-disagreement.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-codex-file-missing-override.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-codex-keychain-disagreement.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-codex-keychain-missing-override.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-codex-native-override.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-codex-unknown-override.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-codex_cli-backend.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-codex_cli-root.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-foreign-parent.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-mcp.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-missing-TBH_CREDENTIAL_BACKEND.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-missing-XDG_CACHE_HOME.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-missing-XDG_CONFIG_HOME.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-missing-XDG_DATA_HOME.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-missing-XDG_STATE_HOME.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-opencode-backend.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-opencode-file.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-opencode-root.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-pi-backend.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-pi-file.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-pi-root.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-relative-path.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-root-boolean.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-root-empty.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-root-extra.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-root-false.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-root-null.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-root-string.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-store-empty.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-store-extra.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-store-keychain.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-store-missing.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-store-native.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-store-null.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-store-override.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-system-prompt.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-unknown-revision.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-unknown-top.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-v3-backend.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-v3-credential_store.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-v3-root_context.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-v4-on-v1.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-v4-on-v2.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-v4-on-v3.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/invalid-wrong-suffix.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/valid-claude_code-native.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/valid-claude_code.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/valid-codex-file.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/valid-codex-keychain.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/valid-codex_cli-native.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/valid-codex_cli.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/valid-muse-no-root.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/valid-muse-root.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/valid-opencode-native.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/valid-opencode.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/valid-pi-native.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envfragment/testdata/curator-spec-muse/v4-cases/valid-pi.json",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envprofile/credential_policy.go",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envprofile/credential_policy_test.go",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envprofile/managed.go",
    "merge_exit": 0,
    "equal": false,
    "result": "bounded policy-in-plan fix plus gofmt"
  },
  {
    "path": "internal/envprofile/muse.go",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envprofile/muse_credential_policy_test.go",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envprofile/muse_test.go",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "internal/envregistry/envregistry.go",
    "merge_exit": 0,
    "equal": true,
    "result": "matches clean three-way merge of accepted revision"
  },
  {
    "path": "scripts/remote-gate.sh",
    "result": "no delta; signed snapshot fix already on trunk",
    "equal": true
  },
  {
    "path": "internal/envprofile/materialize_capture_test.go",
    "result": "only new tracked path: extended existing no-probe test with Muse cases"
  }
]
```

## no-probe-before-01.log

```text
=== RUN   TestMaterializeCaptureAndStatusNeverProbeVendor
=== RUN   TestMaterializeCaptureAndStatusNeverProbeVendor/export
=== RUN   TestMaterializeCaptureAndStatusNeverProbeVendor/manifest-only
=== NAME  TestMaterializeCaptureAndStatusNeverProbeVendor
    materialize_capture_test.go:498: vendor executable ran
--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor (1.24s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/export (0.43s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/manifest-only (0.17s)
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.624s
FAIL
```

## no-probe-after-01.log

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
--- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor (7.40s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/codex (1.27s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/codex/export (0.14s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/codex/manifest-only (0.13s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared (1.51s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared/export (0.14s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared/manifest-only (0.14s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared (1.61s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared/export (0.14s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared/manifest-only (0.14s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated (1.47s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated/export (0.15s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated/manifest-only (0.14s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated (1.54s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated/export (0.14s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated/manifest-only (0.13s)
PASS
ok  	github.com/relux-works/curator/internal/envprofile	7.740s
```

## focused-01.log

```text
ok  	github.com/relux-works/curator/internal/envprofile	69.421s
ok  	github.com/relux-works/curator/internal/envfragment	0.248s
ok  	github.com/relux-works/curator/internal/envregistry	0.234s
ok  	github.com/relux-works/curator/internal/config	0.269s
```

## cli-mutants-01.log

```text
=== RUN   TestCredentialStoreCLIConfigRoundTrip
--- PASS: TestCredentialStoreCLIConfigRoundTrip (1.12s)
=== RUN   TestCredentialStoreCLIConfigRefusesMalformed
--- PASS: TestCredentialStoreCLIConfigRefusesMalformed (0.00s)
=== RUN   TestCredentialStoreCLIPromotion
--- PASS: TestCredentialStoreCLIPromotion (0.00s)
=== RUN   TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery
--- PASS: TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery (1.11s)
=== RUN   TestReviewDottedCredentialProfile
=== RUN   TestReviewDottedCredentialProfile/team
=== RUN   TestReviewDottedCredentialProfile/team.prod
--- PASS: TestReviewDottedCredentialProfile (0.05s)
    --- PASS: TestReviewDottedCredentialProfile/team (0.03s)
    --- PASS: TestReviewDottedCredentialProfile/team.prod (0.03s)
=== RUN   TestReviewDottedCredentialProfileExisting
=== RUN   TestReviewDottedCredentialProfileExisting/show
=== RUN   TestReviewDottedCredentialProfileExisting/unset
--- PASS: TestReviewDottedCredentialProfileExisting (0.02s)
    --- PASS: TestReviewDottedCredentialProfileExisting/show (0.01s)
    --- PASS: TestReviewDottedCredentialProfileExisting/unset (0.01s)
=== RUN   TestCredentialStoreCLIDottedProfileRouting
=== RUN   TestCredentialStoreCLIDottedProfileRouting/team/codex
=== RUN   TestCredentialStoreCLIDottedProfileRouting/team/codex_cli
=== RUN   TestCredentialStoreCLIDottedProfileRouting/team/claude
=== RUN   TestCredentialStoreCLIDottedProfileRouting/team/claude_code
=== RUN   TestCredentialStoreCLIDottedProfileRouting/team.prod/codex
=== RUN   TestCredentialStoreCLIDottedProfileRouting/team.prod/codex_cli
=== RUN   TestCredentialStoreCLIDottedProfileRouting/team.prod/claude
=== RUN   TestCredentialStoreCLIDottedProfileRouting/team.prod/claude_code
=== RUN   TestCredentialStoreCLIDottedProfileRouting/Team..prod/codex
=== RUN   TestCredentialStoreCLIDottedProfileRouting/Team..prod/codex_cli
=== RUN   TestCredentialStoreCLIDottedProfileRouting/Team..prod/claude
=== RUN   TestCredentialStoreCLIDottedProfileRouting/Team..prod/claude_code
=== RUN   TestCredentialStoreCLIDottedProfileRouting/codex.prod/codex
=== RUN   TestCredentialStoreCLIDottedProfileRouting/codex.prod/codex_cli
=== RUN   TestCredentialStoreCLIDottedProfileRouting/codex.prod/claude
=== RUN   TestCredentialStoreCLIDottedProfileRouting/codex.prod/claude_code
--- PASS: TestCredentialStoreCLIDottedProfileRouting (0.15s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team/codex (0.02s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team/codex_cli (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team/claude (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team/claude_code (0.02s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team.prod/codex (0.03s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team.prod/codex_cli (0.04s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team.prod/claude (0.02s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team.prod/claude_code (0.01s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/Team..prod/codex (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/Team..prod/codex_cli (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/Team..prod/claude (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/Team..prod/claude_code (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/codex.prod/codex (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/codex.prod/codex_cli (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/codex.prod/claude (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/codex.prod/claude_code (0.00s)
=== RUN   TestCredentialStoreCLIInvalidProfileRouting
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/.team/set
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/.team/show
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/.team/unset
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/team./set
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/team./show
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/team./unset
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/team/prod/set
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/team/prod/show
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/team/prod/unset
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/CON.prod/set
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/CON.prod/show
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/CON.prod/unset
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/ppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/set
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/ppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/show
=== RUN   TestCredentialStoreCLIInvalidProfileRouting/ppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/unset
--- PASS: TestCredentialStoreCLIInvalidProfileRouting (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/.team/set (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/.team/show (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/.team/unset (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/team./set (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/team./show (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/team./unset (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/team/prod/set (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/team/prod/show (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/team/prod/unset (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/CON.prod/set (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/CON.prod/show (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/CON.prod/unset (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/ppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/set (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/ppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/show (0.00s)
    --- PASS: TestCredentialStoreCLIInvalidProfileRouting/ppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppppp/unset (0.00s)
=== RUN   TestMuseCredentialPolicyNarrowingMutants
=== RUN   TestMuseCredentialPolicyNarrowingMutants/fragment-reobserves-muse-release
    muse_credential_mutants_test.go:73: fragment-reobserves-muse-release: expected-red go test exit 1; killed by TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared
=== RUN   TestMuseCredentialPolicyNarrowingMutants/muse-keychain-admitted
    muse_credential_mutants_test.go:73: muse-keychain-admitted: expected-red go test exit 1; killed by TestMuseCredentialPolicyMatrix/keychain/shared
=== RUN   TestMuseCredentialPolicyNarrowingMutants/missing-shared-source-admitted
    muse_credential_mutants_test.go:73: missing-shared-source-admitted: expected-red go test exit 1; killed by TestMuseMissingNativeFilePolicy/shared
=== RUN   TestMuseCredentialPolicyNarrowingMutants/isolated-muse-native-link
    muse_credential_mutants_test.go:73: isolated-muse-native-link: expected-red go test exit 1; killed by TestMuseCredentialPolicyMatrix/file/isolated
=== RUN   TestMuseCredentialPolicyNarrowingMutants/renderer-drops-only-backend
    muse_credential_mutants_test.go:73: renderer-drops-only-backend: expected-red go test exit 1; killed by TestMuseCLIStorePolicy
=== RUN   TestMuseCredentialPolicyNarrowingMutants/header-only-root-fact-dropped
    muse_credential_mutants_test.go:73: header-only-root-fact-dropped: expected-red go test exit 1; killed by TestMuseRootContextVectors/zero-applicable
=== RUN   TestMuseCredentialPolicyNarrowingMutants/boundary-admits-HOME
    muse_credential_mutants_test.go:73: boundary-admits-HOME: expected-red go test exit 1; killed by TestMuseFragmentBoundary/HOME
=== RUN   TestMuseCredentialPolicyNarrowingMutants/boundary-admits-missing-literal
    muse_credential_mutants_test.go:73: boundary-admits-missing-literal: expected-red go test exit 1; killed by TestMuseFragmentBoundary/missing-literal
=== RUN   TestMuseCredentialPolicyNarrowingMutants/boundary-admits-false-root
    muse_credential_mutants_test.go:73: boundary-admits-false-root: expected-red go test exit 1; killed by TestMuseFragmentBoundary/false-root
--- PASS: TestMuseCredentialPolicyNarrowingMutants (23.31s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/fragment-reobserves-muse-release (4.62s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/muse-keychain-admitted (2.88s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/missing-shared-source-admitted (3.15s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/isolated-muse-native-link (2.78s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/renderer-drops-only-backend (4.01s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/header-only-root-fact-dropped (3.75s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/boundary-admits-HOME (0.70s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/boundary-admits-missing-literal (0.67s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/boundary-admits-false-root (0.75s)
=== RUN   TestMuseRootContextVectors
=== RUN   TestMuseRootContextVectors/monolithic
=== RUN   TestMuseRootContextVectors/zero-applicable
=== RUN   TestMuseRootContextVectors/no-context
=== RUN   TestMuseRootContextVectors/referenced-refused
--- PASS: TestMuseRootContextVectors (5.14s)
    --- PASS: TestMuseRootContextVectors/monolithic (1.32s)
    --- PASS: TestMuseRootContextVectors/zero-applicable (1.30s)
    --- PASS: TestMuseRootContextVectors/no-context (1.34s)
    --- PASS: TestMuseRootContextVectors/referenced-refused (1.18s)
=== RUN   TestMuseRootContextDrift
=== RUN   TestMuseRootContextDrift/changed
=== RUN   TestMuseRootContextDrift/removed
--- PASS: TestMuseRootContextDrift (2.95s)
    --- PASS: TestMuseRootContextDrift/changed (1.50s)
    --- PASS: TestMuseRootContextDrift/removed (1.45s)
=== RUN   TestMuseRootContextStatus
--- PASS: TestMuseRootContextStatus (1.10s)
=== RUN   TestMuseRootContextNarrowingMutants
=== RUN   TestMuseRootContextNarrowingMutants/configured-form-ignored
    muse_root_context_test.go:291: configured-form-ignored: expected-red go test exit 1; killed by TestMuseRootContextVectors/referenced-refused
=== RUN   TestMuseRootContextNarrowingMutants/referenced-fallback
    muse_root_context_test.go:291: referenced-fallback: expected-red go test exit 1; killed by TestMuseRootContextVectors/referenced-refused
=== RUN   TestMuseRootContextNarrowingMutants/no-context-admitted
    muse_root_context_test.go:291: no-context-admitted: expected-red go test exit 1; killed by TestMuseRootContextVectors/no-context
=== RUN   TestMuseRootContextNarrowingMutants/wrong-target
    muse_root_context_test.go:291: wrong-target: expected-red go test exit 1; killed by TestMuseRootContextVectors/monolithic
=== RUN   TestMuseRootContextNarrowingMutants/wrong-bytes
    muse_root_context_test.go:291: wrong-bytes: expected-red go test exit 1; killed by TestMuseRootContextVectors/zero-applicable
=== RUN   TestMuseRootContextNarrowingMutants/changed-root-admitted
    muse_root_context_test.go:291: changed-root-admitted: expected-red go test exit 1; killed by TestMuseRootContextDrift/changed
=== RUN   TestMuseRootContextNarrowingMutants/removed-root-admitted
    muse_root_context_test.go:291: removed-root-admitted: expected-red go test exit 1; killed by TestMuseRootContextDrift/removed
--- PASS: TestMuseRootContextNarrowingMutants (28.45s)
    --- PASS: TestMuseRootContextNarrowingMutants/configured-form-ignored (4.63s)
    --- PASS: TestMuseRootContextNarrowingMutants/referenced-fallback (3.99s)
    --- PASS: TestMuseRootContextNarrowingMutants/no-context-admitted (3.89s)
    --- PASS: TestMuseRootContextNarrowingMutants/wrong-target (4.09s)
    --- PASS: TestMuseRootContextNarrowingMutants/wrong-bytes (3.87s)
    --- PASS: TestMuseRootContextNarrowingMutants/changed-root-admitted (4.36s)
    --- PASS: TestMuseRootContextNarrowingMutants/removed-root-admitted (3.62s)
=== RUN   TestMuseCLIFragmentAndFakeLaunch
--- PASS: TestMuseCLIFragmentAndFakeLaunch (1.07s)
=== RUN   TestMuseFragmentV3PublishedCases
    muse_test.go:221: CURATOR_CONFORMANCE_ROOT is not set
--- SKIP: TestMuseFragmentV3PublishedCases (0.00s)
=== RUN   TestMuseCLIAuthParentBoundary
=== RUN   TestMuseCLIAuthParentBoundary/resolve
    muse_test.go:467: production CLI refusal exit=1; no fragment emitted
=== RUN   TestMuseCLIAuthParentBoundary/resolve-repair
    muse_test.go:467: production CLI refusal exit=1; no fragment emitted
--- PASS: TestMuseCLIAuthParentBoundary (1.09s)
    --- PASS: TestMuseCLIAuthParentBoundary/resolve (0.01s)
    --- PASS: TestMuseCLIAuthParentBoundary/resolve-repair (0.01s)
=== RUN   TestMuseFragmentV4PublishedCases
=== RUN   TestMuseFragmentV4PublishedCases/valid-muse-root.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-muse-no-root.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-missing-XDG_CONFIG_HOME.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-missing-XDG_DATA_HOME.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-missing-XDG_STATE_HOME.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-missing-XDG_CACHE_HOME.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-missing-TBH_CREDENTIAL_BACKEND.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-root-null.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-root-empty.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-root-boolean.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-root-false.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-root-string.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-root-extra.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-backend-keychain.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-backend-absolute.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-backend-empty.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-store-native.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-store-keychain.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-store-null.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-store-empty.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-store-extra.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-store-override.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-store-missing.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-HOME.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-foreign-parent.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-wrong-suffix.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-relative-path.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-unknown-top.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-unknown-revision.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-system-prompt.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-mcp.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-claude_code.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-claude_code-root.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-claude_code-backend.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-claude_code-native.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-claude_code-file.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-codex_cli.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-codex_cli-root.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-codex_cli-backend.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-codex_cli-native.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-codex-file.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-codex-file-disagreement.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-codex-file-missing-override.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-codex-keychain.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-codex-keychain-disagreement.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-codex-keychain-missing-override.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-codex-native-override.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-codex-unknown-override.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-opencode.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-opencode-root.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-opencode-backend.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-opencode-native.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-opencode-file.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-pi.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-pi-root.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-pi-backend.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-pi-native.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-pi-file.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-v4-on-v1.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-additions-on-v1.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-v4-on-v2.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-additions-on-v2.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-v4-on-v3.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-additions-on-v3.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-v3-root_context.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-v3-credential_store.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-v3-backend.json
--- PASS: TestMuseFragmentV4PublishedCases (0.05s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-muse-root.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-muse-no-root.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-missing-XDG_CONFIG_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-missing-XDG_DATA_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-missing-XDG_STATE_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-missing-XDG_CACHE_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-missing-TBH_CREDENTIAL_BACKEND.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-root-null.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-root-empty.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-root-boolean.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-root-false.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-root-string.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-root-extra.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-backend-keychain.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-backend-absolute.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-backend-empty.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-store-native.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-store-keychain.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-store-null.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-store-empty.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-store-extra.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-store-override.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-store-missing.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-HOME.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-foreign-parent.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-wrong-suffix.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-relative-path.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-unknown-top.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-unknown-revision.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-system-prompt.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-mcp.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-claude_code.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-claude_code-root.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-claude_code-backend.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-claude_code-native.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-claude_code-file.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-codex_cli.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-codex_cli-root.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-codex_cli-backend.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-codex_cli-native.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-codex-file.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-codex-file-disagreement.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-codex-file-missing-override.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-codex-keychain.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-codex-keychain-disagreement.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-codex-keychain-missing-override.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-codex-native-override.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-codex-unknown-override.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-opencode.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-opencode-root.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-opencode-backend.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-opencode-native.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-opencode-file.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-pi.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-pi-root.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-pi-backend.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-pi-native.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-pi-file.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-v4-on-v1.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-additions-on-v1.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-v4-on-v2.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-additions-on-v2.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-v4-on-v3.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-additions-on-v3.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-v3-root_context.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-v3-credential_store.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-v3-backend.json (0.00s)
=== RUN   TestMuseCLIStorePolicy
--- PASS: TestMuseCLIStorePolicy (1.15s)
PASS
ok  	github.com/relux-works/curator/cmd/curator	67.112s
```

## snap-plants-01.log

```text
{"mutant": "M-policy-probe", "test": "^TestMaterializeCaptureAndStatusNeverProbeVendor$", "exit": 1, "verdict": "killed", "failures": ["--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor (8.45s)", "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/codex (1.50s)", "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared (1.58s)", "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared (1.89s)", "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated (1.76s)", "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated (1.72s)"]}
{"mutant": "M-release-cache", "test": "^TestMaterializeCaptureCredentialPolicyRelease(Observation|Cache)$", "exit": 1, "verdict": "killed", "failures": ["--- FAIL: TestMaterializeCaptureCredentialPolicyReleaseObservation (1.32s)", "--- FAIL: TestMaterializeCaptureCredentialPolicyReleaseCache (0.47s)"]}
{"mutant": "M-probe", "test": "^TestMaterializeCaptureAndStatusNeverProbeVendor$", "exit": 1, "verdict": "killed", "failures": ["--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor (7.87s)", "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/codex (1.33s)", "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared (1.51s)", "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared (1.67s)", "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated (1.75s)", "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated (1.61s)"]}
```

## M-policy-probe.log

```text
--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor (8.45s)
    --- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/codex (1.50s)
        materialize_capture_test.go:564: vendor executable ran
    --- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared (1.58s)
        materialize_capture_test.go:564: vendor executable ran
    --- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared (1.89s)
        materialize_capture_test.go:564: vendor executable ran
    --- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated (1.76s)
        materialize_capture_test.go:564: vendor executable ran
    --- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated (1.72s)
        materialize_capture_test.go:564: vendor executable ran
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	8.777s
FAIL
```

## M-release-cache.log

```text
--- FAIL: TestMaterializeCaptureCredentialPolicyReleaseObservation (1.32s)
    materialize_capture_test.go:976: invalid native selector reached capture
--- FAIL: TestMaterializeCaptureCredentialPolicyReleaseCache (0.47s)
    materialize_capture_test.go:1011: release unknown: policy=&{selection:native effectiveStore:native launchEnv:map[] isolation:shared nativeStore: links:map[] strategies:map[] records:[]} err=<nil>; want fresh isolated
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	2.120s
FAIL
```

## M-probe.log

```text
--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor (7.87s)
    --- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/codex (1.33s)
        materialize_capture_test.go:564: vendor executable ran
    --- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared (1.51s)
        materialize_capture_test.go:564: vendor executable ran
    --- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared (1.67s)
        materialize_capture_test.go:564: vendor executable ran
    --- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated (1.75s)
        materialize_capture_test.go:564: vendor executable ran
    --- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated (1.61s)
        materialize_capture_test.go:564: vendor executable ran
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	8.195s
FAIL
```

## mutants.json

```text
[
  {
    "mutant": "M-policy-probe",
    "test": "^TestMaterializeCaptureAndStatusNeverProbeVendor$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor (8.45s)",
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/codex (1.50s)",
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared (1.58s)",
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared (1.89s)",
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated (1.76s)",
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated (1.72s)"
    ]
  },
  {
    "mutant": "M-release-cache",
    "test": "^TestMaterializeCaptureCredentialPolicyRelease(Observation|Cache)$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCaptureCredentialPolicyReleaseObservation (1.32s)",
      "--- FAIL: TestMaterializeCaptureCredentialPolicyReleaseCache (0.47s)"
    ]
  },
  {
    "mutant": "M-probe",
    "test": "^TestMaterializeCaptureAndStatusNeverProbeVendor$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor (7.87s)",
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/codex (1.33s)",
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared (1.51s)",
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared (1.67s)",
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated (1.75s)",
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated (1.61s)"
    ]
  }
]
```

## pinned-muse-01.log

```text
=== RUN   TestReviewDottedCredentialProfile
=== RUN   TestReviewDottedCredentialProfile/team
=== RUN   TestReviewDottedCredentialProfile/team.prod
--- PASS: TestReviewDottedCredentialProfile (0.05s)
    --- PASS: TestReviewDottedCredentialProfile/team (0.03s)
    --- PASS: TestReviewDottedCredentialProfile/team.prod (0.02s)
=== RUN   TestReviewDottedCredentialProfileExisting
=== RUN   TestReviewDottedCredentialProfileExisting/show
=== RUN   TestReviewDottedCredentialProfileExisting/unset
--- PASS: TestReviewDottedCredentialProfileExisting (0.07s)
    --- PASS: TestReviewDottedCredentialProfileExisting/show (0.02s)
    --- PASS: TestReviewDottedCredentialProfileExisting/unset (0.06s)
=== RUN   TestMuseCredentialPolicyNarrowingMutants
=== RUN   TestMuseCredentialPolicyNarrowingMutants/fragment-reobserves-muse-release
    muse_credential_mutants_test.go:77: fragment-reobserves-muse-release: expected-red go test exit 1; killed by TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared
=== RUN   TestMuseCredentialPolicyNarrowingMutants/muse-keychain-admitted
    muse_credential_mutants_test.go:77: muse-keychain-admitted: expected-red go test exit 1; killed by TestMuseCredentialPolicyMatrix/keychain/shared
=== RUN   TestMuseCredentialPolicyNarrowingMutants/missing-shared-source-admitted
    muse_credential_mutants_test.go:77: missing-shared-source-admitted: expected-red go test exit 1; killed by TestMuseMissingNativeFilePolicy/shared
=== RUN   TestMuseCredentialPolicyNarrowingMutants/isolated-muse-native-link
    muse_credential_mutants_test.go:77: isolated-muse-native-link: expected-red go test exit 1; killed by TestMuseCredentialPolicyMatrix/file/isolated
=== RUN   TestMuseCredentialPolicyNarrowingMutants/renderer-drops-only-backend
    muse_credential_mutants_test.go:77: renderer-drops-only-backend: expected-red go test exit 1; killed by TestMuseCLIStorePolicy
=== RUN   TestMuseCredentialPolicyNarrowingMutants/header-only-root-fact-dropped
    muse_credential_mutants_test.go:77: header-only-root-fact-dropped: expected-red go test exit 1; killed by TestMuseRootContextVectors/zero-applicable
=== RUN   TestMuseCredentialPolicyNarrowingMutants/boundary-admits-HOME
    muse_credential_mutants_test.go:77: boundary-admits-HOME: expected-red go test exit 1; killed by TestMuseFragmentBoundary/HOME
=== RUN   TestMuseCredentialPolicyNarrowingMutants/boundary-admits-missing-literal
    muse_credential_mutants_test.go:77: boundary-admits-missing-literal: expected-red go test exit 1; killed by TestMuseFragmentBoundary/missing-literal
=== RUN   TestMuseCredentialPolicyNarrowingMutants/boundary-admits-false-root
    muse_credential_mutants_test.go:77: boundary-admits-false-root: expected-red go test exit 1; killed by TestMuseFragmentBoundary/false-root
--- PASS: TestMuseCredentialPolicyNarrowingMutants (14.74s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/fragment-reobserves-muse-release (2.54s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/muse-keychain-admitted (1.74s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/missing-shared-source-admitted (1.27s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/isolated-muse-native-link (1.28s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/renderer-drops-only-backend (3.79s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/header-only-root-fact-dropped (3.08s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/boundary-admits-HOME (0.39s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/boundary-admits-missing-literal (0.31s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/boundary-admits-false-root (0.33s)
=== RUN   TestMuseRootContextVectors
=== RUN   TestMuseRootContextVectors/monolithic
=== RUN   TestMuseRootContextVectors/zero-applicable
=== RUN   TestMuseRootContextVectors/no-context
=== RUN   TestMuseRootContextVectors/referenced-refused
--- PASS: TestMuseRootContextVectors (3.85s)
    --- PASS: TestMuseRootContextVectors/monolithic (0.96s)
    --- PASS: TestMuseRootContextVectors/zero-applicable (1.03s)
    --- PASS: TestMuseRootContextVectors/no-context (0.94s)
    --- PASS: TestMuseRootContextVectors/referenced-refused (0.92s)
=== RUN   TestMuseRootContextDrift
=== RUN   TestMuseRootContextDrift/changed
=== RUN   TestMuseRootContextDrift/removed
--- PASS: TestMuseRootContextDrift (2.52s)
    --- PASS: TestMuseRootContextDrift/changed (1.16s)
    --- PASS: TestMuseRootContextDrift/removed (1.36s)
=== RUN   TestMuseRootContextStatus
--- PASS: TestMuseRootContextStatus (0.65s)
=== RUN   TestMuseRootContextNarrowingMutants
=== RUN   TestMuseRootContextNarrowingMutants/configured-form-ignored
    muse_root_context_test.go:291: configured-form-ignored: expected-red go test exit 1; killed by TestMuseRootContextVectors/referenced-refused
=== RUN   TestMuseRootContextNarrowingMutants/referenced-fallback
    muse_root_context_test.go:291: referenced-fallback: expected-red go test exit 1; killed by TestMuseRootContextVectors/referenced-refused
=== RUN   TestMuseRootContextNarrowingMutants/no-context-admitted
    muse_root_context_test.go:291: no-context-admitted: expected-red go test exit 1; killed by TestMuseRootContextVectors/no-context
=== RUN   TestMuseRootContextNarrowingMutants/wrong-target
    muse_root_context_test.go:291: wrong-target: expected-red go test exit 1; killed by TestMuseRootContextVectors/monolithic
=== RUN   TestMuseRootContextNarrowingMutants/wrong-bytes
    muse_root_context_test.go:291: wrong-bytes: expected-red go test exit 1; killed by TestMuseRootContextVectors/zero-applicable
=== RUN   TestMuseRootContextNarrowingMutants/changed-root-admitted
    muse_root_context_test.go:291: changed-root-admitted: expected-red go test exit 1; killed by TestMuseRootContextDrift/changed
=== RUN   TestMuseRootContextNarrowingMutants/removed-root-admitted
    muse_root_context_test.go:291: removed-root-admitted: expected-red go test exit 1; killed by TestMuseRootContextDrift/removed
--- PASS: TestMuseRootContextNarrowingMutants (21.36s)
    --- PASS: TestMuseRootContextNarrowingMutants/configured-form-ignored (3.06s)
    --- PASS: TestMuseRootContextNarrowingMutants/referenced-fallback (3.08s)
    --- PASS: TestMuseRootContextNarrowingMutants/no-context-admitted (3.19s)
    --- PASS: TestMuseRootContextNarrowingMutants/wrong-target (3.55s)
    --- PASS: TestMuseRootContextNarrowingMutants/wrong-bytes (3.16s)
    --- PASS: TestMuseRootContextNarrowingMutants/changed-root-admitted (3.53s)
    --- PASS: TestMuseRootContextNarrowingMutants/removed-root-admitted (1.78s)
=== RUN   TestMuseCLIFragmentAndFakeLaunch
--- PASS: TestMuseCLIFragmentAndFakeLaunch (1.02s)
=== RUN   TestMuseProfileSeedSecretAndCredentialRefusals
=== RUN   TestMuseProfileSeedSecretAndCredentialRefusals/secret
=== RUN   TestMuseProfileSeedSecretAndCredentialRefusals/auth
--- PASS: TestMuseProfileSeedSecretAndCredentialRefusals (0.15s)
    --- PASS: TestMuseProfileSeedSecretAndCredentialRefusals/secret (0.15s)
    --- PASS: TestMuseProfileSeedSecretAndCredentialRefusals/auth (0.00s)
=== RUN   TestMuseFragmentV3PublishedCases
=== RUN   TestMuseFragmentV3PublishedCases/valid.json
    muse_test.go:275: production Resolve + CLI emission; schema/reader oracle also checked
=== RUN   TestMuseFragmentV3PublishedCases/invalid.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/valid-permissions-yolo-unlocked.json
    muse_test.go:275: production Resolve + CLI emission; schema/reader oracle also checked
=== RUN   TestMuseFragmentV3PublishedCases/valid-permissions-native-locked.json
    muse_test.go:275: production Resolve + CLI emission; schema/reader oracle also checked
=== RUN   TestMuseFragmentV3PublishedCases/valid-permissions-native-silent.json
    muse_test.go:275: production Resolve + CLI emission; schema/reader oracle also checked
=== RUN   TestMuseFragmentV3PublishedCases/invalid-permissions-absent.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-permissions-unknown-mode.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-permissions-yolo-locked.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-permissions-yolo-silent.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-permissions-locked-profile-source.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-permissions-global-unlocked.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-permissions-unknown-source.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-permissions-unknown-field.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-permissions-missing-mode.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/valid-claude_code.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-extra-parent-claude_code.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/valid-codex_cli.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-extra-parent-codex_cli.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/valid-opencode.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-extra-parent-opencode.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/valid-pi.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-extra-parent-pi.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-missing-XDG_CONFIG_HOME.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-wrong-parent-XDG_CONFIG_HOME.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-missing-XDG_DATA_HOME.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-wrong-parent-XDG_DATA_HOME.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-missing-XDG_STATE_HOME.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-wrong-parent-XDG_STATE_HOME.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-missing-XDG_CACHE_HOME.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-wrong-parent-XDG_CACHE_HOME.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-HOME-added.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-HOME-substitution.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-unverified-system-prompt.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-unverified-mcp.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-v2-token.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== RUN   TestMuseFragmentV3PublishedCases/invalid-unknown-environment.json
    muse_test.go:280: static schema/reader oracle; not a production emission row
=== NAME  TestMuseFragmentV3PublishedCases
    coverage.go:236: published cases launch-env-fragment-v3/schema-cases: 36 driven, 0 known-gap, 0 bound, 0 skipped, 36 total
--- PASS: TestMuseFragmentV3PublishedCases (1.41s)
    --- PASS: TestMuseFragmentV3PublishedCases/valid.json (0.47s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/valid-permissions-yolo-unlocked.json (0.05s)
    --- PASS: TestMuseFragmentV3PublishedCases/valid-permissions-native-locked.json (0.05s)
    --- PASS: TestMuseFragmentV3PublishedCases/valid-permissions-native-silent.json (0.05s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-permissions-absent.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-permissions-unknown-mode.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-permissions-yolo-locked.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-permissions-yolo-silent.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-permissions-locked-profile-source.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-permissions-global-unlocked.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-permissions-unknown-source.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-permissions-unknown-field.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-permissions-missing-mode.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/valid-claude_code.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-extra-parent-claude_code.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/valid-codex_cli.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-extra-parent-codex_cli.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/valid-opencode.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-extra-parent-opencode.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/valid-pi.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-extra-parent-pi.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-missing-XDG_CONFIG_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-wrong-parent-XDG_CONFIG_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-missing-XDG_DATA_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-wrong-parent-XDG_DATA_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-missing-XDG_STATE_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-wrong-parent-XDG_STATE_HOME.json (0.01s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-missing-XDG_CACHE_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-wrong-parent-XDG_CACHE_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-HOME-added.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-HOME-substitution.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-unverified-system-prompt.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-unverified-mcp.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-v2-token.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-unknown-environment.json (0.00s)
=== RUN   TestMuseCLIAuthParentBoundary
=== RUN   TestMuseCLIAuthParentBoundary/resolve
    muse_test.go:467: production CLI refusal exit=1; no fragment emitted
=== RUN   TestMuseCLIAuthParentBoundary/resolve-repair
    muse_test.go:467: production CLI refusal exit=1; no fragment emitted
--- PASS: TestMuseCLIAuthParentBoundary (1.40s)
    --- PASS: TestMuseCLIAuthParentBoundary/resolve (0.01s)
    --- PASS: TestMuseCLIAuthParentBoundary/resolve-repair (0.01s)
=== RUN   TestMuseFragmentV4PublishedCases
=== RUN   TestMuseFragmentV4PublishedCases/valid-muse-root.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-muse-no-root.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-missing-XDG_CONFIG_HOME.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-missing-XDG_DATA_HOME.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-missing-XDG_STATE_HOME.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-missing-XDG_CACHE_HOME.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-missing-TBH_CREDENTIAL_BACKEND.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-root-null.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-root-empty.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-root-boolean.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-root-false.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-root-string.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-root-extra.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-backend-keychain.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-backend-absolute.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-backend-empty.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-store-native.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-store-keychain.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-store-null.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-store-empty.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-store-extra.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-store-override.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-store-missing.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-HOME.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-foreign-parent.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-wrong-suffix.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-relative-path.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-unknown-top.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-unknown-revision.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-system-prompt.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-mcp.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-claude_code.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-claude_code-root.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-claude_code-backend.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-claude_code-native.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-claude_code-file.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-codex_cli.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-codex_cli-root.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-codex_cli-backend.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-codex_cli-native.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-codex-file.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-codex-file-disagreement.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-codex-file-missing-override.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-codex-keychain.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-codex-keychain-disagreement.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-codex-keychain-missing-override.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-codex-native-override.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-codex-unknown-override.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-opencode.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-opencode-root.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-opencode-backend.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-opencode-native.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-opencode-file.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-pi.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-pi-root.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-pi-backend.json
=== RUN   TestMuseFragmentV4PublishedCases/valid-pi-native.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-pi-file.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-v4-on-v1.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-additions-on-v1.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-v4-on-v2.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-additions-on-v2.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-v4-on-v3.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-additions-on-v3.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-v3-root_context.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-v3-credential_store.json
=== RUN   TestMuseFragmentV4PublishedCases/invalid-v3-backend.json
--- PASS: TestMuseFragmentV4PublishedCases (0.05s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-muse-root.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-muse-no-root.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-missing-XDG_CONFIG_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-missing-XDG_DATA_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-missing-XDG_STATE_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-missing-XDG_CACHE_HOME.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-missing-TBH_CREDENTIAL_BACKEND.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-root-null.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-root-empty.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-root-boolean.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-root-false.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-root-string.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-root-extra.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-backend-keychain.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-backend-absolute.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-backend-empty.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-store-native.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-store-keychain.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-store-null.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-store-empty.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-store-extra.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-store-override.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-store-missing.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-HOME.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-foreign-parent.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-wrong-suffix.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-relative-path.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-unknown-top.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-unknown-revision.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-system-prompt.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-mcp.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-claude_code.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-claude_code-root.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-claude_code-backend.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-claude_code-native.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-claude_code-file.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-codex_cli.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-codex_cli-root.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-codex_cli-backend.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-codex_cli-native.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-codex-file.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-codex-file-disagreement.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-codex-file-missing-override.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-codex-keychain.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-codex-keychain-disagreement.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-codex-keychain-missing-override.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-codex-native-override.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-codex-unknown-override.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-opencode.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-opencode-root.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-opencode-backend.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-opencode-native.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-opencode-file.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-pi.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-pi-root.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-pi-backend.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-pi-native.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-pi-file.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-v4-on-v1.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-additions-on-v1.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-v4-on-v2.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-additions-on-v2.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-v4-on-v3.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-additions-on-v3.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-v3-root_context.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-v3-credential_store.json (0.00s)
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-v3-backend.json (0.00s)
=== RUN   TestMuseCLIStorePolicy
--- PASS: TestMuseCLIStorePolicy (1.56s)
PASS
ok  	github.com/relux-works/curator/cmd/curator	49.202s
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
--- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor (6.63s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/codex (1.28s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/codex/export (0.21s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/codex/manifest-only (0.20s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared (1.47s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared/export (0.20s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared/manifest-only (0.18s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared (1.63s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared/export (0.23s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared/manifest-only (0.20s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated (1.12s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated/export (0.13s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated/manifest-only (0.13s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated (1.12s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated/export (0.14s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated/manifest-only (0.14s)
=== RUN   TestMaterializeCaptureCredentialPolicyReleaseObservation
--- PASS: TestMaterializeCaptureCredentialPolicyReleaseObservation (0.90s)
=== RUN   TestMaterializeCaptureCredentialPolicyReleaseCache
--- PASS: TestMaterializeCaptureCredentialPolicyReleaseCache (0.27s)
=== RUN   TestMuseCredentialPolicyMatrix
=== RUN   TestMuseCredentialPolicyMatrix//shared
=== RUN   TestMuseCredentialPolicyMatrix//isolated
=== RUN   TestMuseCredentialPolicyMatrix/native/shared
=== RUN   TestMuseCredentialPolicyMatrix/native/isolated
=== RUN   TestMuseCredentialPolicyMatrix/file/shared
=== RUN   TestMuseCredentialPolicyMatrix/file/isolated
=== RUN   TestMuseCredentialPolicyMatrix/keychain/shared
=== RUN   TestMuseCredentialPolicyMatrix/keychain/isolated
--- PASS: TestMuseCredentialPolicyMatrix (5.18s)
    --- PASS: TestMuseCredentialPolicyMatrix//shared (0.80s)
    --- PASS: TestMuseCredentialPolicyMatrix//isolated (0.70s)
    --- PASS: TestMuseCredentialPolicyMatrix/native/shared (1.06s)
    --- PASS: TestMuseCredentialPolicyMatrix/native/isolated (0.81s)
    --- PASS: TestMuseCredentialPolicyMatrix/file/shared (0.76s)
    --- PASS: TestMuseCredentialPolicyMatrix/file/isolated (0.57s)
    --- PASS: TestMuseCredentialPolicyMatrix/keychain/shared (0.22s)
    --- PASS: TestMuseCredentialPolicyMatrix/keychain/isolated (0.26s)
=== RUN   TestMuseMissingNativeFilePolicy
=== RUN   TestMuseMissingNativeFilePolicy/shared
=== RUN   TestMuseMissingNativeFilePolicy/isolated
--- PASS: TestMuseMissingNativeFilePolicy (0.94s)
    --- PASS: TestMuseMissingNativeFilePolicy/shared (0.56s)
    --- PASS: TestMuseMissingNativeFilePolicy/isolated (0.38s)
=== RUN   TestMusePolicySwitchPreservesCredentials
=== RUN   TestMusePolicySwitchPreservesCredentials/link
=== RUN   TestMusePolicySwitchPreservesCredentials/detached-file
--- PASS: TestMusePolicySwitchPreservesCredentials (1.22s)
    --- PASS: TestMusePolicySwitchPreservesCredentials/link (0.62s)
    --- PASS: TestMusePolicySwitchPreservesCredentials/detached-file (0.60s)
=== RUN   TestMusePolicyExplicitMigration
--- PASS: TestMusePolicyExplicitMigration (0.82s)
=== RUN   TestMusePublishedLinkStates
=== RUN   TestMusePublishedLinkStates/live-resolve
=== RUN   TestMusePublishedLinkStates/live-repair
=== RUN   TestMusePublishedLinkStates/write-through-refresh-resolve
=== RUN   TestMusePublishedLinkStates/write-through-refresh-repair
=== RUN   TestMusePublishedLinkStates/missing-link-resolve
=== RUN   TestMusePublishedLinkStates/missing-link-repair
=== RUN   TestMusePublishedLinkStates/temp-rename-fork-resolve
=== RUN   TestMusePublishedLinkStates/temp-rename-fork-repair
=== RUN   TestMusePublishedLinkStates/retargeted-link-resolve
=== RUN   TestMusePublishedLinkStates/retargeted-link-repair
=== RUN   TestMusePublishedLinkStates/dangling-target-resolve
=== RUN   TestMusePublishedLinkStates/dangling-target-repair
=== RUN   TestMusePublishedLinkStates/metadata-unreadable-resolve
=== RUN   TestMusePublishedLinkStates/metadata-unreadable-repair
=== RUN   TestMusePublishedLinkStates/target-unreadable-resolve
=== RUN   TestMusePublishedLinkStates/target-unreadable-repair
=== NAME  TestMusePublishedLinkStates
    coverage.go:236: published cases environments-muse/cases: 16 driven, 0 known-gap, 0 bound, 0 skipped, 16 total
--- PASS: TestMusePublishedLinkStates (10.29s)
    --- PASS: TestMusePublishedLinkStates/live-resolve (0.60s)
    --- PASS: TestMusePublishedLinkStates/live-repair (0.60s)
    --- PASS: TestMusePublishedLinkStates/write-through-refresh-resolve (0.60s)
    --- PASS: TestMusePublishedLinkStates/write-through-refresh-repair (0.72s)
    --- PASS: TestMusePublishedLinkStates/missing-link-resolve (0.59s)
    --- PASS: TestMusePublishedLinkStates/missing-link-repair (0.73s)
    --- PASS: TestMusePublishedLinkStates/temp-rename-fork-resolve (0.75s)
    --- PASS: TestMusePublishedLinkStates/temp-rename-fork-repair (0.67s)
    --- PASS: TestMusePublishedLinkStates/retargeted-link-resolve (0.71s)
    --- PASS: TestMusePublishedLinkStates/retargeted-link-repair (0.67s)
    --- PASS: TestMusePublishedLinkStates/dangling-target-resolve (0.63s)
    --- PASS: TestMusePublishedLinkStates/dangling-target-repair (0.52s)
    --- PASS: TestMusePublishedLinkStates/metadata-unreadable-resolve (0.52s)
    --- PASS: TestMusePublishedLinkStates/metadata-unreadable-repair (0.71s)
    --- PASS: TestMusePublishedLinkStates/target-unreadable-resolve (0.68s)
    --- PASS: TestMusePublishedLinkStates/target-unreadable-repair (0.58s)
=== RUN   TestMuseForkRefusedRegression
--- PASS: TestMuseForkRefusedRegression (1.16s)
=== RUN   TestMuseAuthParentBoundaryAtResolve
=== RUN   TestMuseAuthParentBoundaryAtResolve/config/outside/resolve
=== RUN   TestMuseAuthParentBoundaryAtResolve/config/outside/resolve-repair
=== RUN   TestMuseAuthParentBoundaryAtResolve/config/inside/resolve
=== RUN   TestMuseAuthParentBoundaryAtResolve/config/inside/resolve-repair
=== RUN   TestMuseAuthParentBoundaryAtResolve/config/muse/outside/resolve
=== RUN   TestMuseAuthParentBoundaryAtResolve/config/muse/outside/resolve-repair
=== RUN   TestMuseAuthParentBoundaryAtResolve/config/muse/inside/resolve
=== RUN   TestMuseAuthParentBoundaryAtResolve/config/muse/inside/resolve-repair
--- PASS: TestMuseAuthParentBoundaryAtResolve (4.26s)
    --- PASS: TestMuseAuthParentBoundaryAtResolve/config/outside/resolve (0.56s)
    --- PASS: TestMuseAuthParentBoundaryAtResolve/config/outside/resolve-repair (0.62s)
    --- PASS: TestMuseAuthParentBoundaryAtResolve/config/inside/resolve (0.47s)
    --- PASS: TestMuseAuthParentBoundaryAtResolve/config/inside/resolve-repair (0.52s)
    --- PASS: TestMuseAuthParentBoundaryAtResolve/config/muse/outside/resolve (0.42s)
    --- PASS: TestMuseAuthParentBoundaryAtResolve/config/muse/outside/resolve-repair (0.55s)
    --- PASS: TestMuseAuthParentBoundaryAtResolve/config/muse/inside/resolve (0.46s)
    --- PASS: TestMuseAuthParentBoundaryAtResolve/config/muse/inside/resolve-repair (0.67s)
=== RUN   TestMuseDirectoriesAndIsolation
--- PASS: TestMuseDirectoriesAndIsolation (0.55s)
=== RUN   TestMuseAuthRepairAdditionalStates
=== RUN   TestMuseAuthRepairAdditionalStates/empty-directory
=== RUN   TestMuseAuthRepairAdditionalStates/nonempty-directory
=== RUN   TestMuseAuthRepairAdditionalStates/relative-link
=== RUN   TestMuseAuthRepairAdditionalStates/target-directory
=== RUN   TestMuseAuthRepairAdditionalStates/readlink-error
=== RUN   TestMuseAuthRepairAdditionalStates/missing-link-missing-target
=== RUN   TestMuseAuthRepairAdditionalStates/config-parent-link
--- PASS: TestMuseAuthRepairAdditionalStates (4.55s)
    --- PASS: TestMuseAuthRepairAdditionalStates/empty-directory (0.61s)
    --- PASS: TestMuseAuthRepairAdditionalStates/nonempty-directory (0.45s)
    --- PASS: TestMuseAuthRepairAdditionalStates/relative-link (0.77s)
    --- PASS: TestMuseAuthRepairAdditionalStates/target-directory (0.85s)
    --- PASS: TestMuseAuthRepairAdditionalStates/readlink-error (0.86s)
    --- PASS: TestMuseAuthRepairAdditionalStates/missing-link-missing-target (0.52s)
    --- PASS: TestMuseAuthRepairAdditionalStates/config-parent-link (0.48s)
=== RUN   TestMuseGlobalReconciliationUsesExplicitProvisioning
--- PASS: TestMuseGlobalReconciliationUsesExplicitProvisioning (0.53s)
=== RUN   TestMuseStatusOptionalProvisioning
=== RUN   TestMuseStatusOptionalProvisioning/absent
=== RUN   TestMuseStatusOptionalProvisioning/marker-unreadable
=== RUN   TestMuseStatusOptionalProvisioning/backup-unreadable
=== RUN   TestMuseStatusOptionalProvisioning/marker-malformed
--- PASS: TestMuseStatusOptionalProvisioning (1.69s)
    --- PASS: TestMuseStatusOptionalProvisioning/absent (0.08s)
    --- PASS: TestMuseStatusOptionalProvisioning/marker-unreadable (0.09s)
    --- PASS: TestMuseStatusOptionalProvisioning/backup-unreadable (0.08s)
    --- PASS: TestMuseStatusOptionalProvisioning/marker-malformed (0.08s)
=== RUN   TestMusePublishedRegistryLayout
    muse_test.go:540: registry/layout production checks exercised; exec/serve permission mappings remain the separately owned launcher integration gap
--- PASS: TestMusePublishedRegistryLayout (0.83s)
PASS
ok  	github.com/relux-works/curator/internal/envprofile	40.064s
```

## packages-01.log

```text
ok  	github.com/relux-works/curator/internal/envfragment	0.233s
ok  	github.com/relux-works/curator/internal/envregistry	0.154s
ok  	github.com/relux-works/curator/internal/config	0.999s
```

## build-01.log

```text
(empty output; exit 0 recorded by process runner)
```

## vet-01.log

```text
(empty output; exit 0 recorded by process runner)
```

## lint-01.log

```text
0 issues.
```

## whitespace-01.log

```text
(empty output; exit 0 recorded by process runner)
```
