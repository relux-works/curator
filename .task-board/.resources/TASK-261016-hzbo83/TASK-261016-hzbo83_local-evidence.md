# Local verification evidence

Bound to candidate c82d1cd07ba9d550b6c10c0fe9910ec27e28d592. No hosted verdict is inferred from these logs. Temporary fixture paths are redacted below; large emitted document byte arrays are abbreviated, without changing test names, diagnostics or status.

## Focused tests, exit 0

```text
ok  	github.com/relux-works/curator/internal/envprofile	67.980s
ok  	github.com/relux-works/curator/internal/envregistry	0.467s
ok  	github.com/relux-works/curator/internal/config	0.667s
ok  	github.com/relux-works/curator/cmd/curator	41.032s
```

## Projection and identity tests, exit 0

```text
ok  	github.com/relux-works/curator/internal/envprofile	16.588s
```

## Build, exit 0 (empty output expected)

```text
```

## Lint with task-local cache, exit 0

```text
golangci-lint run
0 issues.
```

## Focused coverage, exit 0

```text
ok  	github.com/relux-works/curator/internal/envprofile	53.230s	coverage: 33.7% of statements
```

## Policy-module statement coverage

```text
credential_policy.go: 63 of 77 statements covered (81.8%). Focused slice coverage only; package-wide coverage is not a claim.
```

## Current portable controls harness, exit 0; every child exit 1 expected red

```text
{"mutant": "M1-auto-isolated", "exit_code": 1, "named_failure": "TestCredentialProjectionIsolationRefusals/auto", "killed": true}
{"mutant": "M2-muse-file-admitted", "exit_code": 1, "named_failure": "TestCredentialStorePolicyActivation/muse/file", "killed": true}
{"mutant": "M3-permission-defaults-file", "exit_code": 1, "named_failure": "TestCredentialStorePolicyNativeInspection/permission", "killed": true}
{"mutant": "M4-keyring-resurrects-file", "exit_code": 1, "named_failure": "TestCodexKeyringAmbient/keyring", "killed": true}
{"mutant": "M5-shared-second-observation", "exit_code": 1, "named_failure": "TestCredentialProjectionSingleObservation/shared", "killed": true}
{"mutant": "M6-error-reobservation", "exit_code": 1, "named_failure": "TestCredentialProjectionSingleObservation", "killed": true}
{"mutant": "M7-pi-subrequest-reuses-policy", "exit_code": 1, "named_failure": "TestCredentialProjectionIdentitySubrequest/isolated", "killed": true}
```

## Expected-red mutant logs (all child exits 1)

### M1-auto-isolated

```text
--- FAIL: TestCredentialProjectionIsolationRefusals (0.47s)
    --- FAIL: TestCredentialProjectionIsolationRefusals/auto (0.47s)
        credential_projection_test.go:118: native auto isolation must refuse: result=&{Document:<emitted fragment bytes abbreviated> Notice:Managed home provisioned at <temporary-fixture>
            The tool treats this home as its own state root: sessions, trust records, and approvals accrue here, not in the native home.
            First-run steps the seeds do not cover: Log in unless the native credential is shared through; pass --skip-git-repo-check outside a git repository. Warnings:[environment_tool_version_unverified: codex_cli detected release unknown, recorded 0.153.2 mcp_env_passthrough_unlisted: passing FIGMA_API_KEY with passable_env_names absent; list the named variables to keep passing them after the flip] Provisioned:true StaleReasons:[]} err=<nil>
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	0.829s
FAIL
```

### M2-muse-file-admitted

```text
--- FAIL: TestCredentialStorePolicyActivation (0.70s)
    --- FAIL: TestCredentialStorePolicyActivation/muse/file (0.07s)
        credential_policy_test.go:30: Resolve = <nil>, environment_credential_conflict: <temporary-fixture>: native credential target is absent; want environment_credential_store_unsupported and no fragment
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	0.930s
FAIL
```

### M3-permission-defaults-file

```text
--- FAIL: TestCredentialStorePolicyNativeInspection (0.42s)
    --- FAIL: TestCredentialStorePolicyNativeInspection/permission (0.00s)
        credential_policy_test.go:81: Resolve = <nil>, codex_settings_unsupported: explicit Codex file storage requires qualified login/app-server/refresh behavior and strict successor fragment consumers; no native fallback; want environment_seed_unreadable
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	0.679s
FAIL
```

### M4-keyring-resurrects-file

```text
--- FAIL: TestCodexKeyringAmbient (0.66s)
    --- FAIL: TestCodexKeyringAmbient/keyring (0.66s)
        managed_test.go:810: a keyring store is ambient, records 1
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	0.977s
FAIL
```

### M5-shared-second-observation

```text
--- FAIL: TestCredentialProjectionSingleObservation (0.79s)
    --- FAIL: TestCredentialProjectionSingleObservation/shared (0.78s)
        credential_projection_test.go:51: one file observation should resolve: result=<nil> err=environment_seed_unreadable: codex native config.toml: permission denied reads=2
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.071s
FAIL
```

### M6-error-reobservation

```text
--- FAIL: TestCredentialProjectionSingleObservation (2.18s)
    --- FAIL: TestCredentialProjectionSingleObservation/shared (1.12s)
        credential_projection_test.go:85: changed selector must refuse: result=<nil> err=environment_seed_unreadable: codex native config.toml: permission denied reads=3
    --- FAIL: TestCredentialProjectionSingleObservation/isolated (1.06s)
        credential_projection_test.go:85: changed selector must refuse: result=<nil> err=environment_seed_unreadable: codex native config.toml: permission denied reads=3
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	2.484s
FAIL
```

### M7-pi-subrequest-reuses-policy

```text
--- FAIL: TestCredentialProjectionIdentitySubrequest (0.85s)
    --- FAIL: TestCredentialProjectionIdentitySubrequest/isolated (0.85s)
        credential_projection_test.go:201: environment_credential_conflict: legacy credentials require explicit env migrate before repair
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.085s
FAIL
```

