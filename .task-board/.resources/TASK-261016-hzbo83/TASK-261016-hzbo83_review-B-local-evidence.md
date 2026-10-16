# Review B executed local evidence

Candidate 43b7ed0e9efd36ffcc321f6604d72aea221e2af7; Go 1.25.5, darwin/arm64.

## exact-blobs.log
```text
7 of 7 changed candidate blobs match worktree. Candidate 43b7ed0e9efd36ffcc321f6604d72aea221e2af7
```

## focused.log
```text
ok  	github.com/relux-works/curator/internal/envprofile	89.316s
ok  	github.com/relux-works/curator/internal/envregistry	0.266s
ok  	github.com/relux-works/curator/internal/config	0.301s
ok  	github.com/relux-works/curator/cmd/curator	41.403s
```

## row1.log
```text
ok  	github.com/relux-works/curator/internal/envprofile	1.155s
```

## row2.log
```text
ok  	github.com/relux-works/curator/internal/envprofile	1.315s
```

## row3.log
```text
ok  	github.com/relux-works/curator/internal/envprofile	41.872s
```

## free-hunt.log
```text
ok  	github.com/relux-works/curator/internal/envprofile	1.541s
```

## controls.log
```text
{"mutant": "M1-auto-isolated", "exit_code": 1, "named_failure": "TestCredentialProjectionIsolationRefusals/auto", "killed": true}
{"mutant": "M2-muse-file-admitted", "exit_code": 1, "named_failure": "TestCredentialStorePolicyActivation/muse/file", "killed": true}
{"mutant": "M3-permission-defaults-file", "exit_code": 1, "named_failure": "TestCredentialStorePolicyNativeInspection/permission", "killed": true}
{"mutant": "M4-keyring-resurrects-file", "exit_code": 1, "named_failure": "TestCodexKeyringAmbient/keyring", "killed": true}
{"mutant": "M5-shared-second-observation", "exit_code": 1, "named_failure": "TestCredentialProjectionSingleObservation/shared", "killed": true}
{"mutant": "M6-error-reobservation", "exit_code": 1, "named_failure": "TestCredentialProjectionSingleObservation", "killed": true}
{"mutant": "M7-pi-subrequest-reuses-policy", "exit_code": 1, "named_failure": "TestCredentialProjectionIdentitySubrequest/isolated", "killed": true}
```

## M1-auto-isolated.log
```text
--- FAIL: TestCredentialProjectionIsolationRefusals (0.51s)
    --- FAIL: TestCredentialProjectionIsolationRefusals/auto (0.51s)
        credential_projection_test.go:118: native auto isolation must refuse: result=&{Document:[redacted synthetic fragment bytes] Notice:Managed home provisioned at <temporary-fixture>
            The tool treats this home as its own state root: sessions, trust records, and approvals accrue here, not in the native home.
            First-run steps the seeds do not cover: Log in unless the native credential is shared through; pass --skip-git-repo-check outside a git repository. Warnings:[environment_tool_version_unverified: codex_cli detected release unknown, recorded 0.153.2 mcp_env_passthrough_unlisted: passing FIGMA_API_KEY with passable_env_names absent; list the named variables to keep passing them after the flip] Provisioned:true StaleReasons:[]} err=<nil>
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	0.865s
FAIL
```

## M2-muse-file-admitted.log
```text
--- FAIL: TestCredentialStorePolicyActivation (0.75s)
    --- FAIL: TestCredentialStorePolicyActivation/muse/file (0.12s)
        credential_policy_test.go:30: Resolve = <nil>, environment_credential_conflict: <temporary-fixture>: native credential target is absent; want environment_credential_store_unsupported and no fragment
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.095s
FAIL
```

## M3-permission-defaults-file.log
```text
--- FAIL: TestCredentialStorePolicyNativeInspection (0.27s)
    --- FAIL: TestCredentialStorePolicyNativeInspection/permission (0.00s)
        credential_policy_test.go:81: Resolve = <nil>, codex_settings_unsupported: explicit Codex file storage requires qualified login/app-server/refresh behavior and strict successor fragment consumers; no native fallback; want environment_seed_unreadable
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	0.622s
FAIL
```

## M4-keyring-resurrects-file.log
```text
--- FAIL: TestCodexKeyringAmbient (0.52s)
    --- FAIL: TestCodexKeyringAmbient/keyring (0.52s)
        managed_test.go:810: a keyring store is ambient, records 1
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	0.849s
FAIL
```

## M5-shared-second-observation.log
```text
--- FAIL: TestCredentialProjectionSingleObservation (0.58s)
    --- FAIL: TestCredentialProjectionSingleObservation/shared (0.58s)
        credential_projection_test.go:51: one file observation should resolve: result=<nil> err=environment_seed_unreadable: codex native config.toml: permission denied reads=2
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	0.926s
FAIL
```

## M6-error-reobservation.log
```text
--- FAIL: TestCredentialProjectionSingleObservation (2.16s)
    --- FAIL: TestCredentialProjectionSingleObservation/shared (0.95s)
        credential_projection_test.go:85: changed selector must refuse: result=<nil> err=environment_seed_unreadable: codex native config.toml: permission denied reads=3
    --- FAIL: TestCredentialProjectionSingleObservation/isolated (1.21s)
        credential_projection_test.go:85: changed selector must refuse: result=<nil> err=environment_seed_unreadable: codex native config.toml: permission denied reads=3
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	2.516s
FAIL
```

## M7-pi-subrequest-reuses-policy.log
```text
--- FAIL: TestCredentialProjectionIdentitySubrequest (0.61s)
    --- FAIL: TestCredentialProjectionIdentitySubrequest/isolated (0.61s)
        credential_projection_test.go:201: environment_credential_conflict: legacy credentials require explicit env migrate before repair
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	0.957s
FAIL
```

Synthetic fragment bytes and temporary fixture paths are redacted; assertions, test names and exit codes are retained.
