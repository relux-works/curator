# Reviewer local evidence

Candidate tree: 43b7ed0e9efd36ffcc321f6604d72aea221e2af7.
All commands ran on this candidate with Go 1.25.5, darwin/arm64.
Filesystem fixture paths are redacted.

## focused.log

Command: `go test ./internal/config ./internal/envregistry ./internal/envprofile ./cmd/curator -run 'TestCredential(Store|Projection)|TestRC14' -count=1`

Exit: 0

```text
ok  	github.com/relux-works/curator/internal/config	0.472s
ok  	github.com/relux-works/curator/internal/envregistry	0.441s [no tests to run]
ok  	github.com/relux-works/curator/internal/envprofile	27.946s
ok  	github.com/relux-works/curator/cmd/curator	1.723s

```

## row1.log

Command: `go test -overlay .temp/review-hzbo83/attacks-overlay.json ./internal/envprofile -run '^TestReviewTransientPolicyRead$' -count=1 -v`

Exit: 0

```text
=== RUN   TestReviewTransientPolicyRead
--- PASS: TestReviewTransientPolicyRead (0.65s)
PASS
ok  	github.com/relux-works/curator/internal/envprofile	1.913s

```

## row2.log

Command: `go test -overlay .temp/review-hzbo83/attacks-overlay.json ./internal/envprofile -run '^TestReviewStatusBackendSwitch$' -count=1 -v`

Exit: 0

```text
=== RUN   TestReviewStatusBackendSwitch
--- PASS: TestReviewStatusBackendSwitch (0.78s)
PASS
ok  	github.com/relux-works/curator/internal/envprofile	2.054s

```

## row3.log

Command: `go test ./internal/envprofile -run 'TestSharedToIsolatedRemovesStaleLink|TestCredentialProjectionIdentitySubrequest|TestRC14.*(Migration|Migrates)|TestMigrate(PlanDriftRefuses|RollbackRestoresPriorState|InterruptedApplyRecovers|RecoveryRefusesUnexpectedTarget|NoSecretCopies)' -count=1 -v`

Exit: 0

```text
=== RUN   TestSharedToIsolatedRemovesStaleLink
--- PASS: TestSharedToIsolatedRemovesStaleLink (1.04s)
=== RUN   TestCredentialProjectionIdentitySubrequest
=== RUN   TestCredentialProjectionIdentitySubrequest/shared
=== RUN   TestCredentialProjectionIdentitySubrequest/isolated
--- PASS: TestCredentialProjectionIdentitySubrequest (1.32s)
    --- PASS: TestCredentialProjectionIdentitySubrequest/shared (0.76s)
    --- PASS: TestCredentialProjectionIdentitySubrequest/isolated (0.55s)
=== RUN   TestMigrateNoSecretCopies
=== RUN   TestMigrateNoSecretCopies/interrupted_journal_holds_no_credential_bytes
--- PASS: TestMigrateNoSecretCopies (2.57s)
    --- PASS: TestMigrateNoSecretCopies/interrupted_journal_holds_no_credential_bytes (1.30s)
=== RUN   TestMigratePlanDriftRefuses
--- PASS: TestMigratePlanDriftRefuses (0.87s)
=== RUN   TestMigrateRollbackRestoresPriorState
=== RUN   TestMigrateRollbackRestoresPriorState/link_phase
=== RUN   TestMigrateRollbackRestoresPriorState/post_publish
--- PASS: TestMigrateRollbackRestoresPriorState (2.06s)
    --- PASS: TestMigrateRollbackRestoresPriorState/link_phase (0.80s)
    --- PASS: TestMigrateRollbackRestoresPriorState/post_publish (1.26s)
=== RUN   TestMigrateInterruptedApplyRecovers
--- PASS: TestMigrateInterruptedApplyRecovers (1.55s)
=== RUN   TestMigrateRecoveryRefusesUnexpectedTarget
--- PASS: TestMigrateRecoveryRefusesUnexpectedTarget (1.01s)
=== RUN   TestRC14MigrationRehashesLegacyIdentities
--- PASS: TestRC14MigrationRehashesLegacyIdentities (1.55s)
=== RUN   TestRC14ResolveRepairMigratesSiblingHomes
--- PASS: TestRC14ResolveRepairMigratesSiblingHomes (4.61s)
=== RUN   TestRC14IdentityMigrationRollsBackEveryEntry
--- PASS: TestRC14IdentityMigrationRollsBackEveryEntry (4.39s)
=== RUN   TestRC14IdentityMigrationRefusesSiblingPlanDrift
--- PASS: TestRC14IdentityMigrationRefusesSiblingPlanDrift (0.90s)
=== RUN   TestRC14IdentityMigrationRefusesVersionAndPinMismatch
=== RUN   TestRC14IdentityMigrationRefusesVersionAndPinMismatch/marker-version
=== RUN   TestRC14IdentityMigrationRefusesVersionAndPinMismatch/lock-version
=== RUN   TestRC14IdentityMigrationRefusesVersionAndPinMismatch/store-pin
--- PASS: TestRC14IdentityMigrationRefusesVersionAndPinMismatch (1.42s)
    --- PASS: TestRC14IdentityMigrationRefusesVersionAndPinMismatch/marker-version (0.48s)
    --- PASS: TestRC14IdentityMigrationRefusesVersionAndPinMismatch/lock-version (0.48s)
    --- PASS: TestRC14IdentityMigrationRefusesVersionAndPinMismatch/store-pin (0.46s)
=== RUN   TestRC14IdentityMigrationRecoversInterruptedCommit
--- PASS: TestRC14IdentityMigrationRecoversInterruptedCommit (2.77s)
=== RUN   TestRC14Schema1IdentityMigrationWithoutCredentialOperations
=== RUN   TestRC14Schema1IdentityMigrationWithoutCredentialOperations/resolve
=== RUN   TestRC14Schema1IdentityMigrationWithoutCredentialOperations/migrate
--- PASS: TestRC14Schema1IdentityMigrationWithoutCredentialOperations (2.73s)
    --- PASS: TestRC14Schema1IdentityMigrationWithoutCredentialOperations/resolve (1.31s)
    --- PASS: TestRC14Schema1IdentityMigrationWithoutCredentialOperations/migrate (1.42s)
=== RUN   TestRC14IdentityMigrationPreservesFallbackCopies
--- PASS: TestRC14IdentityMigrationPreservesFallbackCopies (1.47s)
=== RUN   TestRC14UseMigratesLegacyProfileBeforeNativePublication
--- PASS: TestRC14UseMigratesLegacyProfileBeforeNativePublication (0.76s)
PASS
ok  	github.com/relux-works/curator/internal/envprofile	31.341s

```

## free-hunt.log

Command: `go test -overlay .temp/review-hzbo83/attacks-overlay.json ./internal/envprofile -run '^TestReviewDetachedBackendSwitch$' -count=1 -v`

Exit: 0

```text
=== RUN   TestReviewDetachedBackendSwitch
=== RUN   TestReviewDetachedBackendSwitch/link
=== RUN   TestReviewDetachedBackendSwitch/detached
--- PASS: TestReviewDetachedBackendSwitch (1.44s)
    --- PASS: TestReviewDetachedBackendSwitch/link (0.76s)
    --- PASS: TestReviewDetachedBackendSwitch/detached (0.68s)
PASS
ok  	github.com/relux-works/curator/internal/envprofile	1.771s

```

## controls-row1.log

Command: `python3 .temp/review-hzbo83/controls-row1.py`

Exit: 0

```text
{"mutant": "M1-auto-isolated", "exit_code": 1, "named_failure": "TestCredentialProjectionIsolationRefusals/auto", "killed": true}
{"mutant": "M2-muse-file-admitted", "exit_code": 1, "named_failure": "TestCredentialStorePolicyActivation/muse/file", "killed": true}
{"mutant": "M3-permission-defaults-file", "exit_code": 1, "named_failure": "TestCredentialStorePolicyNativeInspection/permission", "killed": true}
{"mutant": "M4-keyring-resurrects-file", "exit_code": 1, "named_failure": "TestCodexKeyringAmbient/keyring", "killed": true}
{"mutant": "M5-shared-second-observation", "exit_code": 1, "named_failure": "TestCredentialProjectionSingleObservation/shared", "killed": true}
{"mutant": "M6-error-reobservation", "exit_code": 1, "named_failure": "TestCredentialProjectionSingleObservation", "killed": true}

```

## controls-row3.log

Command: `python3 .temp/review-hzbo83/controls-row3.py`

Exit: 0

```text
{"mutant": "M7-pi-subrequest-reuses-policy", "exit_code": 1, "named_failure": "TestCredentialProjectionIdentitySubrequest/isolated", "killed": true}

```

## M1-auto-isolated.log

Expected child exit: 1; named failure required by harness.

```text
--- FAIL: TestCredentialProjectionIsolationRefusals (0.43s)
    --- FAIL: TestCredentialProjectionIsolationRefusals/auto (0.43s)
        credential_projection_test.go:118: native auto isolation must refuse: result=&{Document:[123 34 101 110 118 34 58 123 34 67 79 68 69 88 95 72 79 77 69 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 67 114 101 100 101 110 116 105 97 108 80 114 111 106 101 99 116 105 111 110 73 115 111 108 97 116 105 111 110 82 101 102 117 115 97 108 115 97 117 116 111 53 57 52 55 56 53 53 49 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 105 115 111 108 97 116 105 111 110 47 99 111 100 101 120 95 99 108 105 34 125 44 34 101 110 118 105 114 111 110 109 101 110 116 34 58 34 99 111 100 101 120 95 99 108 105 34 44 34 102 114 97 103 109 101 110 116 34 58 34 108 97 117 110 99 104 45 101 110 118 45 102 114 97 103 109 101 110 116 45 118 50 34 44 34 109 99 112 34 58 123 34 99 104 97 110 110 101 108 115 34 58 91 123 34 97 114 103 117 109 101 110 116 34 58 34 110 97 109 101 34 44 34 102 108 97 103 34 58 34 45 112 34 44 34 107 105 110 100 34 58 34 102 108 97 103 34 44 34 110 97 109 101 34 58 34 99 117 114 97 116 111 114 45 109 99 112 34 125 93 44 34 101 110 118 95 110 97 109 101 115 34 58 91 34 70 73 71 77 65 95 65 80 73 95 75 69 89 34 93 44 34 112 97 116 104 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 67 114 101 100 101 110 116 105 97 108 80 114 111 106 101 99 116 105 111 110 73 115 111 108 97 116 105 111 110 82 101 102 117 115 97 108 115 97 117 116 111 53 57 52 55 56 53 53 49 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 105 115 111 108 97 116 105 111 110 47 99 111 100 101 120 95 99 108 105 47 99 117 114 97 116 111 114 45 109 99 112 46 99 111 110 102 105 103 46 116 111 109 108 34 125 44 34 112 101 114 109 105 115 115 105 111 110 115 34 58 123 34 108 111 99 107 101 100 34 58 102 97 108 115 101 44 34 109 111 100 101 34 58 34 110 97 116 105 118 101 34 44 34 115 111 117 114 99 101 34 58 34 100 101 102 97 117 108 116 34 125 44 34 112 114 101 99 101 100 101 110 99 101 34 58 123 34 112 108 97 99 101 109 101 110 116 34 58 34 119 105 110 110 101 114 45 108 97 115 116 34 44 34 119 105 110 110 101 114 34 58 34 104 105 103 104 101 114 45 119 101 105 103 104 116 34 125 44 34 112 114 111 102 105 108 101 34 58 123 34 108 111 99 107 95 115 104 97 50 53 54 34 58 34 98 52 51 100 49 52 98 48 101 52 98 53 54 50 102 101 99 55 54 51 53 100 99 99 50 102 102 53 57 51 48 51 54 102 52 57 52 54 98 48 48 101 54 100 99 100 51 51 98 49 101 102 57 101 99 56 48 49 97 50 52 56 53 100 34 44 34 110 97 109 101 34 58 34 105 115 111 108 97 116 105 111 110 34 125 44 34 115 121 115 116 101 109 95 112 114 111 109 112 116 34 58 123 34 99 104 97 110 110 101 108 115 34 58 91 123 34 107 101 121 34 58 34 109 111 100 101 108 95 105 110 115 116 114 117 99 116 105 111 110 115 95 102 105 108 101 34 44 34 107 105 110 100 34 58 34 99 111 110 102 105 103 45 107 101 121 34 44 34 115 101 109 97 110 116 105 99 115 34 58 34 114 101 112 108 97 99 101 34 125 93 44 34 112 97 116 104 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 67 114 101 100 101 110 116 105 97 108 80 114 111 106 101 99 116 105 111 110 73 115 111 108 97 116 105 111 110 82 101 102 117 115 97 108 115 97 117 116 111 53 57 52 55 56 53 53 49 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 105 115 111 108 97 116 105 111 110 47 99 111 100 101 120 95 99 108 105 47 46 97 103 101 110 116 45 99 111 110 116 101 120 116 47 115 121 115 116 101 109 45 112 114 111 109 112 116 46 109 100 34 125 125 10] Notice:Managed home provisioned at <fixture-path>
            The tool treats this home as its own state root: sessions, trust records, and approvals accrue here, not in the native home.
            First-run steps the seeds do not cover: Log in unless the native credential is shared through; pass --skip-git-repo-check outside a git repository. Warnings:[environment_tool_version_unverified: codex_cli detected release unknown, recorded 0.153.2 mcp_env_passthrough_unlisted: passing FIGMA_API_KEY with passable_env_names absent; list the named variables to keep passing them after the flip] Provisioned:true StaleReasons:[]} err=<nil>
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.021s
FAIL

```

## M2-muse-file-admitted.log

Expected child exit: 1; named failure required by harness.

```text
--- FAIL: TestCredentialStorePolicyActivation (0.51s)
    --- FAIL: TestCredentialStorePolicyActivation/muse/file (0.08s)
        credential_policy_test.go:30: Resolve = <nil>, environment_credential_conflict: <fixture-path>: native credential target is absent; want environment_credential_store_unsupported and no fragment
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	0.847s
FAIL

```

## M3-permission-defaults-file.log

Expected child exit: 1; named failure required by harness.

```text
--- FAIL: TestCredentialStorePolicyNativeInspection (0.28s)
    --- FAIL: TestCredentialStorePolicyNativeInspection/permission (0.00s)
        credential_policy_test.go:81: Resolve = <nil>, codex_settings_unsupported: explicit Codex file storage requires qualified login/app-server/refresh behavior and strict successor fragment consumers; no native fallback; want environment_seed_unreadable
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	0.841s
FAIL

```

## M4-keyring-resurrects-file.log

Expected child exit: 1; named failure required by harness.

```text
--- FAIL: TestCodexKeyringAmbient (0.48s)
    --- FAIL: TestCodexKeyringAmbient/keyring (0.48s)
        managed_test.go:810: a keyring store is ambient, records 1
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.318s
FAIL

```

## M5-shared-second-observation.log

Expected child exit: 1; named failure required by harness.

```text
--- FAIL: TestCredentialProjectionSingleObservation (0.60s)
    --- FAIL: TestCredentialProjectionSingleObservation/shared (0.60s)
        credential_projection_test.go:51: one file observation should resolve: result=<nil> err=environment_seed_unreadable: codex native config.toml: permission denied reads=2
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.995s
FAIL

```

## M6-error-reobservation.log

Expected child exit: 1; named failure required by harness.

```text
--- FAIL: TestCredentialProjectionSingleObservation (1.79s)
    --- FAIL: TestCredentialProjectionSingleObservation/shared (0.89s)
        credential_projection_test.go:85: changed selector must refuse: result=<nil> err=environment_seed_unreadable: codex native config.toml: permission denied reads=3
    --- FAIL: TestCredentialProjectionSingleObservation/isolated (0.90s)
        credential_projection_test.go:85: changed selector must refuse: result=<nil> err=environment_seed_unreadable: codex native config.toml: permission denied reads=3
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	2.447s
FAIL

```

## M7-pi-subrequest-reuses-policy.log

Expected child exit: 1; named failure required by harness.

```text
--- FAIL: TestCredentialProjectionIdentitySubrequest (0.57s)
    --- FAIL: TestCredentialProjectionIdentitySubrequest/isolated (0.57s)
        credential_projection_test.go:201: environment_credential_conflict: legacy credentials require explicit env migrate before repair
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.019s
FAIL

```
