
## commands.md

```text
# Independent reviewer execution, revision 2

Candidate: 9bad791482638be26d3ab306653fd279791f449b.
Base: 612960ee90bedc7a4a113eda13aaa6cf2a2cbe9f.
Signed snapshot: be87653a94f23f39acbaa3e965a8f424cf769b60, signature G.
Platform: darwin/arm64, Go 1.25.5.

All commands below were run independently on the candidate. Producer logs are not substituted for reviewer runs.

| Command | Exit | Log |
|---|---:|---|
| go test ./internal/envprofile -run '^TestMaterializeCapture(AndStatusNeverProbeVendor|CredentialPolicyRelease)' -count=1 -v -timeout=3m | 0 | probe.log |
| CURATOR_CONFORMANCE_ROOT=<pinned-root>/conformance/v1 go test ./internal/envprofile ./internal/envfragment ./internal/envregistry ./internal/config -run 'Test(Muse|CredentialStore|CredentialProjection|MaterializeCapture)' -count=1 -v -timeout=6m | 0 | focused.log |
| CURATOR_CONFORMANCE_ROOT=<pinned-root>/conformance/v1 go test ./cmd/curator -run 'TestMuse(CredentialPolicyNarrowingMutants|CLI|Fragment|RootContext)|TestCredential(Store|Projection)|TestReview.*Credential' -count=1 -v -timeout=8m | 0 | cli.log |
| python3 .temp/review-wh6v9a/rev2/snap-harness.py | 0 | snap.log; three named expected-red children |
| python3 .temp/review-wh6v9a/rev2/snap-remaining.py | 0 | snap-remaining.log; nine named expected-red children |
| go test -overlay <review-overlay.json> ./internal/envprofile -run '^TestReviewMuseCaptureRefusalNoProbe$' -count=1 -v -timeout=3m | 0 | free-hunt.log |
| git diff --check | 0 | no diagnostics |

The conformance root is the existing checkout at CI pin 43bf0a2506d5c354a73bbc3ea4623d4653db10c7 under .temp/TASK-261016-wh6v9a/ci-spec. The overlay adds only the attached review test to materialize_capture_test.go; production blobs are unchanged.

Coverage: original producer slice 7 of 8 AC rows production-driven, eighth named structural/boundary-only; convergence interaction 1 of 1 driven. Both bounds remain explicit. Readiness/login, qualified Codex v5/application, real hosted plant and external sources remain in TASK-261017-0pawkw (credential-store-policy-consumer-remainder).

Curator-spec main remained 1059b94b4135089d842bf017070b8ab47d79a536. Three normative documents were independently retrieved at that SHA and compared byte-for-byte with the reviewed local copies. The cited policies are consistent with the accepted narrowed scope.

Merge audit: only LOGBOOK.md, managed.go and the Muse mutant test differ from the clean accepted-to-trunk merge, plus the new no-probe test expansion in materialize_capture_test.go. scripts/remote-gate.sh has no candidate delta: trunk owns the signing fix. The three-way script conflict is intentionally resolved to unchanged trunk, as the brief requires.

Windows no-probe sentinel remains a stated Unix-only limitation. Local focused suites reported no skips. No runtime login/refresh or external consumer qualification is claimed.

```

## sweep.md

```text
# Review sweep: revision 2

Candidate tree: `9bad791482638be26d3ab306653fd279791f449b`.
Scope: accepted-to-combined interaction; revision 1 was accepted, not rejected.
No catalog table or wall-clock budget was supplied. Retain the seven revision-1 surfaces and prepend the changed observation boundary. Original coverage is 7 of 8 producer AC rows driven, with a named structural-only eighth row; revision-2 interaction is 1 of 1 driven through CaptureMaterialize and StatusOf.

| Surface | Result | Evidence and bound |
|---|---|---|
| observation boundary / cache | held | No-probe test covers Codex and Muse native/file x shared/isolated. M-policy-probe, M-release-cache and M-probe each exit 1 with named behavioral failure; baseline exits 0. No vendor discovery or stale different-release projection reproduced. |
| destination conflicts | held | TestMusePolicySwitchPreservesCredentials, TestMusePolicyExplicitMigration: detached/shared transitions preserve credential state; explicit migration remains usable. |
| read boundary | held | TestMuseAuthParentBoundaryAtResolve, TestMuseAuthRepairAdditionalStates, published link vectors: parent/link retarget and invalid entry types refuse beside valid routes. |
| error surface | held | TestMusePublishedLinkStates, TestMuseMissingNativeFilePolicy: failed metadata observations do not become absent/file fallback; no placeholder created. |
| selector paths | held | TestMuseCLIStorePolicy and retained credential config/resolve tests: native/file controls remain admitted, keychain/unqualified selection refuses; stored selector unchanged. |
| text grammar | held | TestMuseFragmentV4PublishedCases and TestMuseFragmentBoundary: 67 pinned cases and missing/unknown/conflicting fields; direct boundary and schema scope only, not external wire ingestion. |
| identity | held | TestMuseRootContextVectors, TestMuseRootContextDrift and seven root narrowing plants: missing/changed/header-only root behavior retained. |
| ordering | held | TestMuseCLIStorePolicy and renderer-drops-only-backend plant: JSON/env/shell retain all five values and registry order; actual CLI rendering executed. |

## Free hunt

TestReviewMuseCaptureRefusalNoProbe (review-only Go overlay) drives CaptureMaterialize and StatusOf after deleting a valid shared native credential. Capture refuses with environment_home_stale before its callback; status is non-current; byte/mode/link snapshot is unchanged. Restoring the credential admits capture. Explicit Detect trap never fires. Exit 0. No blocking finding reproduced.

All twelve SNAP-LOCK controls independently rerun: twelve child exits 1, all named capture assertions, no build/setup failures counted. Together with sixteen Muse/root controls: 28 killed, zero survivors. This is bounded test sensitivity evidence, not exhaustive proof.

```

## authority.txt

```text
environments.md immutable main bytes verified
login-readiness.md immutable main bytes verified
profile-codex-settings.md immutable main bytes verified

```

## merge-comparison.json

```text
{
  "different_from_clean_merge": [
    "LOGBOOK.md",
    "cmd/curator/muse_credential_mutants_test.go",
    "internal/envprofile/managed.go",
    "scripts/remote-gate.sh"
  ],
  "additional_paths": [
    "internal/envprofile/materialize_capture_test.go"
  ],
  "candidate": "9bad791482638be26d3ab306653fd279791f449b"
}

```

## probe.log

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
--- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor (5.80s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/codex (0.82s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/codex/export (0.15s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/codex/manifest-only (0.14s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared (1.16s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared/export (0.14s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared/manifest-only (0.13s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared (1.18s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared/export (0.15s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared/manifest-only (0.17s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated (1.29s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated/export (0.14s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated/manifest-only (0.14s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated (1.36s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated/export (0.16s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated/manifest-only (0.16s)
=== RUN   TestMaterializeCaptureCredentialPolicyReleaseObservation
--- PASS: TestMaterializeCaptureCredentialPolicyReleaseObservation (0.88s)
=== RUN   TestMaterializeCaptureCredentialPolicyReleaseCache
--- PASS: TestMaterializeCaptureCredentialPolicyReleaseCache (0.22s)
PASS
ok  	github.com/relux-works/curator/internal/envprofile	7.228s

```

## focused.log

```text
=== RUN   TestCredentialStorePolicyActivation
=== RUN   TestCredentialStorePolicyActivation/claude_code/file
=== RUN   TestCredentialStorePolicyActivation/claude_code/keychain
=== RUN   TestCredentialStorePolicyActivation/codex_cli/file
=== RUN   TestCredentialStorePolicyActivation/codex_cli/keychain
=== RUN   TestCredentialStorePolicyActivation/muse/keychain
=== RUN   TestCredentialStorePolicyActivation/pi/file
=== RUN   TestCredentialStorePolicyActivation/pi/keychain
=== RUN   TestCredentialStorePolicyActivation/opencode/file
=== RUN   TestCredentialStorePolicyActivation/opencode/keychain
--- PASS: TestCredentialStorePolicyActivation (0.74s)
    --- PASS: TestCredentialStorePolicyActivation/claude_code/file (0.00s)
    --- PASS: TestCredentialStorePolicyActivation/claude_code/keychain (0.00s)
    --- PASS: TestCredentialStorePolicyActivation/codex_cli/file (0.00s)
    --- PASS: TestCredentialStorePolicyActivation/codex_cli/keychain (0.00s)
    --- PASS: TestCredentialStorePolicyActivation/muse/keychain (0.00s)
    --- PASS: TestCredentialStorePolicyActivation/pi/file (0.00s)
    --- PASS: TestCredentialStorePolicyActivation/pi/keychain (0.00s)
    --- PASS: TestCredentialStorePolicyActivation/opencode/file (0.00s)
    --- PASS: TestCredentialStorePolicyActivation/opencode/keychain (0.00s)
=== RUN   TestCredentialStorePolicyNativeInspection
=== RUN   TestCredentialStorePolicyNativeInspection/absent_key
=== RUN   TestCredentialStorePolicyNativeInspection/keyring
=== RUN   TestCredentialStorePolicyNativeInspection/auto
=== RUN   TestCredentialStorePolicyNativeInspection/unknown
=== RUN   TestCredentialStorePolicyNativeInspection/malformed
=== RUN   TestCredentialStorePolicyNativeInspection/permission
--- PASS: TestCredentialStorePolicyNativeInspection (0.38s)
    --- PASS: TestCredentialStorePolicyNativeInspection/absent_key (0.00s)
    --- PASS: TestCredentialStorePolicyNativeInspection/keyring (0.00s)
    --- PASS: TestCredentialStorePolicyNativeInspection/auto (0.00s)
    --- PASS: TestCredentialStorePolicyNativeInspection/unknown (0.00s)
    --- PASS: TestCredentialStorePolicyNativeInspection/malformed (0.00s)
    --- PASS: TestCredentialStorePolicyNativeInspection/permission (0.00s)
=== RUN   TestCredentialProjectionSingleObservation
=== RUN   TestCredentialProjectionSingleObservation/shared
=== RUN   TestCredentialProjectionSingleObservation/isolated
--- PASS: TestCredentialProjectionSingleObservation (1.84s)
    --- PASS: TestCredentialProjectionSingleObservation/shared (0.94s)
    --- PASS: TestCredentialProjectionSingleObservation/isolated (0.90s)
=== RUN   TestCredentialProjectionIsolationRefusals
=== RUN   TestCredentialProjectionIsolationRefusals/file
=== RUN   TestCredentialProjectionIsolationRefusals/keyring
=== RUN   TestCredentialProjectionIsolationRefusals/auto
--- PASS: TestCredentialProjectionIsolationRefusals (1.59s)
    --- PASS: TestCredentialProjectionIsolationRefusals/file (0.47s)
    --- PASS: TestCredentialProjectionIsolationRefusals/keyring (0.68s)
    --- PASS: TestCredentialProjectionIsolationRefusals/auto (0.44s)
=== RUN   TestCredentialProjectionStatusObservation
--- PASS: TestCredentialProjectionStatusObservation (0.99s)
=== RUN   TestCredentialProjectionIdentitySubrequest
=== RUN   TestCredentialProjectionIdentitySubrequest/shared
=== RUN   TestCredentialProjectionIdentitySubrequest/isolated
--- PASS: TestCredentialProjectionIdentitySubrequest (1.68s)
    --- PASS: TestCredentialProjectionIdentitySubrequest/shared (0.72s)
    --- PASS: TestCredentialProjectionIdentitySubrequest/isolated (0.95s)
=== RUN   TestMaterializeCaptureCurrentAndRelease
--- PASS: TestMaterializeCaptureCurrentAndRelease (0.69s)
=== RUN   TestMaterializeCaptureCredentialStoreAdmission
=== RUN   TestMaterializeCaptureCredentialStoreAdmission/file
=== RUN   TestMaterializeCaptureCredentialStoreAdmission/keychain
--- PASS: TestMaterializeCaptureCredentialStoreAdmission (1.37s)
    --- PASS: TestMaterializeCaptureCredentialStoreAdmission/file (0.78s)
    --- PASS: TestMaterializeCaptureCredentialStoreAdmission/keychain (0.59s)
=== RUN   TestMaterializeCapturePendingDoesNotRecover
=== RUN   TestMaterializeCapturePendingDoesNotRecover/prepared
=== RUN   TestMaterializeCapturePendingDoesNotRecover/committing
=== RUN   TestMaterializeCapturePendingDoesNotRecover/cleanup
=== RUN   TestMaterializeCapturePendingDoesNotRecover/rolling_back
=== RUN   TestMaterializeCapturePendingDoesNotRecover/rollback_cleanup
--- PASS: TestMaterializeCapturePendingDoesNotRecover (4.25s)
    --- PASS: TestMaterializeCapturePendingDoesNotRecover/prepared (0.80s)
    --- PASS: TestMaterializeCapturePendingDoesNotRecover/committing (0.74s)
    --- PASS: TestMaterializeCapturePendingDoesNotRecover/cleanup (0.80s)
    --- PASS: TestMaterializeCapturePendingDoesNotRecover/rolling_back (1.09s)
    --- PASS: TestMaterializeCapturePendingDoesNotRecover/rollback_cleanup (0.83s)
=== RUN   TestMaterializeCaptureBadRecoveryEvidence
=== RUN   TestMaterializeCaptureBadRecoveryEvidence/corrupt
=== RUN   TestMaterializeCaptureBadRecoveryEvidence/not-directory
--- PASS: TestMaterializeCaptureBadRecoveryEvidence (1.52s)
    --- PASS: TestMaterializeCaptureBadRecoveryEvidence/corrupt (0.70s)
    --- PASS: TestMaterializeCaptureBadRecoveryEvidence/not-directory (0.82s)
=== RUN   TestMaterializeCaptureWaitReselectsAndTimeout
--- PASS: TestMaterializeCaptureWaitReselectsAndTimeout (1.02s)
=== RUN   TestMaterializeCaptureSameFragmentDrift
=== RUN   TestMaterializeCaptureSameFragmentDrift/aba
=== RUN   TestMaterializeCaptureSameFragmentDrift/mode
=== RUN   TestMaterializeCaptureSameFragmentDrift/bytes
=== RUN   TestMaterializeCaptureSameFragmentDrift/missing
=== RUN   TestMaterializeCaptureSameFragmentDrift/closed-handle
=== RUN   TestMaterializeCaptureSameFragmentDrift/policy
=== RUN   TestMaterializeCaptureSameFragmentDrift/composition-policy
=== RUN   TestMaterializeCaptureSameFragmentDrift/store-bytes
--- PASS: TestMaterializeCaptureSameFragmentDrift (7.60s)
    --- PASS: TestMaterializeCaptureSameFragmentDrift/aba (0.89s)
    --- PASS: TestMaterializeCaptureSameFragmentDrift/mode (1.09s)
    --- PASS: TestMaterializeCaptureSameFragmentDrift/bytes (0.85s)
    --- PASS: TestMaterializeCaptureSameFragmentDrift/missing (0.94s)
    --- PASS: TestMaterializeCaptureSameFragmentDrift/closed-handle (0.94s)
    --- PASS: TestMaterializeCaptureSameFragmentDrift/policy (1.00s)
    --- PASS: TestMaterializeCaptureSameFragmentDrift/composition-policy (0.78s)
    --- PASS: TestMaterializeCaptureSameFragmentDrift/store-bytes (1.12s)
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
--- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor (6.43s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/codex (1.23s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/codex/export (0.19s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/codex/manifest-only (0.21s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared (1.18s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared/export (0.17s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared/manifest-only (0.19s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared (1.24s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared/export (0.17s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared/manifest-only (0.16s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated (1.46s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated/export (0.21s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated/manifest-only (0.20s)
    --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated (1.32s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated/export (0.15s)
        --- PASS: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated/manifest-only (0.18s)
=== RUN   TestMaterializeCaptureDoesNotProvision
=== RUN   TestMaterializeCaptureDoesNotProvision/absent
=== RUN   TestMaterializeCaptureDoesNotProvision/empty
--- PASS: TestMaterializeCaptureDoesNotProvision (0.00s)
    --- PASS: TestMaterializeCaptureDoesNotProvision/absent (0.00s)
    --- PASS: TestMaterializeCaptureDoesNotProvision/empty (0.00s)
=== RUN   TestMaterializeCaptureObjectBounds
--- PASS: TestMaterializeCaptureObjectBounds (0.69s)
=== RUN   TestMaterializeCaptureSerializesProfileWriters
=== RUN   TestMaterializeCaptureSerializesProfileWriters/use
=== RUN   TestMaterializeCaptureSerializesProfileWriters/sync
--- PASS: TestMaterializeCaptureSerializesProfileWriters (2.39s)
    --- PASS: TestMaterializeCaptureSerializesProfileWriters/use (1.11s)
    --- PASS: TestMaterializeCaptureSerializesProfileWriters/sync (1.27s)
=== RUN   TestMaterializeCaptureConcurrentRepair
--- PASS: TestMaterializeCaptureConcurrentRepair (1.39s)
=== RUN   TestMaterializeCaptureStoreObjectIdentity
=== RUN   TestMaterializeCaptureStoreObjectIdentity/stable
=== RUN   TestMaterializeCaptureStoreObjectIdentity/identity
=== RUN   TestMaterializeCaptureStoreObjectIdentity/mode
=== RUN   TestMaterializeCaptureStoreObjectIdentity/bytes
--- PASS: TestMaterializeCaptureStoreObjectIdentity (3.28s)
    --- PASS: TestMaterializeCaptureStoreObjectIdentity/stable (0.84s)
    --- PASS: TestMaterializeCaptureStoreObjectIdentity/identity (1.04s)
    --- PASS: TestMaterializeCaptureStoreObjectIdentity/mode (0.65s)
    --- PASS: TestMaterializeCaptureStoreObjectIdentity/bytes (0.74s)
=== RUN   TestMaterializeCaptureAuthorizingDrift
=== RUN   TestMaterializeCaptureAuthorizingDrift/policy-read
=== RUN   TestMaterializeCaptureAuthorizingDrift/selection
=== RUN   TestMaterializeCaptureAuthorizingDrift/marker-identity
=== RUN   TestMaterializeCaptureAuthorizingDrift/lock-identity
--- PASS: TestMaterializeCaptureAuthorizingDrift (3.69s)
    --- PASS: TestMaterializeCaptureAuthorizingDrift/policy-read (1.06s)
    --- PASS: TestMaterializeCaptureAuthorizingDrift/selection (0.93s)
    --- PASS: TestMaterializeCaptureAuthorizingDrift/marker-identity (0.93s)
    --- PASS: TestMaterializeCaptureAuthorizingDrift/lock-identity (0.76s)
=== RUN   TestMaterializeCaptureUnreadableJournal
--- PASS: TestMaterializeCaptureUnreadableJournal (0.91s)
=== RUN   TestMaterializeCaptureExternalTornGeneration
--- PASS: TestMaterializeCaptureExternalTornGeneration (1.12s)
=== RUN   TestMaterializeCaptureCredentialPolicyReleaseObservation
--- PASS: TestMaterializeCaptureCredentialPolicyReleaseObservation (0.89s)
=== RUN   TestMaterializeCaptureCredentialPolicyReleaseCache
--- PASS: TestMaterializeCaptureCredentialPolicyReleaseCache (0.24s)
=== RUN   TestMuseCredentialPolicyMatrix
=== RUN   TestMuseCredentialPolicyMatrix//shared
=== RUN   TestMuseCredentialPolicyMatrix//isolated
=== RUN   TestMuseCredentialPolicyMatrix/native/shared
=== RUN   TestMuseCredentialPolicyMatrix/native/isolated
=== RUN   TestMuseCredentialPolicyMatrix/file/shared
=== RUN   TestMuseCredentialPolicyMatrix/file/isolated
=== RUN   TestMuseCredentialPolicyMatrix/keychain/shared
=== RUN   TestMuseCredentialPolicyMatrix/keychain/isolated
--- PASS: TestMuseCredentialPolicyMatrix (5.44s)
    --- PASS: TestMuseCredentialPolicyMatrix//shared (0.71s)
    --- PASS: TestMuseCredentialPolicyMatrix//isolated (0.80s)
    --- PASS: TestMuseCredentialPolicyMatrix/native/shared (0.83s)
    --- PASS: TestMuseCredentialPolicyMatrix/native/isolated (0.81s)
    --- PASS: TestMuseCredentialPolicyMatrix/file/shared (0.78s)
    --- PASS: TestMuseCredentialPolicyMatrix/file/isolated (0.70s)
    --- PASS: TestMuseCredentialPolicyMatrix/keychain/shared (0.48s)
    --- PASS: TestMuseCredentialPolicyMatrix/keychain/isolated (0.33s)
=== RUN   TestMuseMissingNativeFilePolicy
=== RUN   TestMuseMissingNativeFilePolicy/shared
=== RUN   TestMuseMissingNativeFilePolicy/isolated
--- PASS: TestMuseMissingNativeFilePolicy (0.93s)
    --- PASS: TestMuseMissingNativeFilePolicy/shared (0.33s)
    --- PASS: TestMuseMissingNativeFilePolicy/isolated (0.60s)
=== RUN   TestMusePolicySwitchPreservesCredentials
=== RUN   TestMusePolicySwitchPreservesCredentials/link
=== RUN   TestMusePolicySwitchPreservesCredentials/detached-file
--- PASS: TestMusePolicySwitchPreservesCredentials (1.48s)
    --- PASS: TestMusePolicySwitchPreservesCredentials/link (0.79s)
    --- PASS: TestMusePolicySwitchPreservesCredentials/detached-file (0.69s)
=== RUN   TestMusePolicyExplicitMigration
--- PASS: TestMusePolicyExplicitMigration (0.95s)
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
--- PASS: TestMusePublishedLinkStates (13.68s)
    --- PASS: TestMusePublishedLinkStates/live-resolve (0.86s)
    --- PASS: TestMusePublishedLinkStates/live-repair (0.59s)
    --- PASS: TestMusePublishedLinkStates/write-through-refresh-resolve (0.69s)
    --- PASS: TestMusePublishedLinkStates/write-through-refresh-repair (0.61s)
    --- PASS: TestMusePublishedLinkStates/missing-link-resolve (1.06s)
    --- PASS: TestMusePublishedLinkStates/missing-link-repair (1.14s)
    --- PASS: TestMusePublishedLinkStates/temp-rename-fork-resolve (0.70s)
    --- PASS: TestMusePublishedLinkStates/temp-rename-fork-repair (0.79s)
    --- PASS: TestMusePublishedLinkStates/retargeted-link-resolve (0.80s)
    --- PASS: TestMusePublishedLinkStates/retargeted-link-repair (0.97s)
    --- PASS: TestMusePublishedLinkStates/dangling-target-resolve (0.71s)
    --- PASS: TestMusePublishedLinkStates/dangling-target-repair (0.59s)
    --- PASS: TestMusePublishedLinkStates/metadata-unreadable-resolve (0.72s)
    --- PASS: TestMusePublishedLinkStates/metadata-unreadable-repair (0.72s)
    --- PASS: TestMusePublishedLinkStates/target-unreadable-resolve (0.74s)
    --- PASS: TestMusePublishedLinkStates/target-unreadable-repair (1.97s)
=== RUN   TestMuseForkRefusedRegression
--- PASS: TestMuseForkRefusedRegression (3.42s)
=== RUN   TestMuseAuthParentBoundaryAtResolve
=== RUN   TestMuseAuthParentBoundaryAtResolve/config/outside/resolve
=== RUN   TestMuseAuthParentBoundaryAtResolve/config/outside/resolve-repair
=== RUN   TestMuseAuthParentBoundaryAtResolve/config/inside/resolve
=== RUN   TestMuseAuthParentBoundaryAtResolve/config/inside/resolve-repair
=== RUN   TestMuseAuthParentBoundaryAtResolve/config/muse/outside/resolve
=== RUN   TestMuseAuthParentBoundaryAtResolve/config/muse/outside/resolve-repair
=== RUN   TestMuseAuthParentBoundaryAtResolve/config/muse/inside/resolve
=== RUN   TestMuseAuthParentBoundaryAtResolve/config/muse/inside/resolve-repair
--- PASS: TestMuseAuthParentBoundaryAtResolve (4.52s)
    --- PASS: TestMuseAuthParentBoundaryAtResolve/config/outside/resolve (0.63s)
    --- PASS: TestMuseAuthParentBoundaryAtResolve/config/outside/resolve-repair (0.62s)
    --- PASS: TestMuseAuthParentBoundaryAtResolve/config/inside/resolve (0.62s)
    --- PASS: TestMuseAuthParentBoundaryAtResolve/config/inside/resolve-repair (0.49s)
    --- PASS: TestMuseAuthParentBoundaryAtResolve/config/muse/outside/resolve (0.70s)
    --- PASS: TestMuseAuthParentBoundaryAtResolve/config/muse/outside/resolve-repair (0.47s)
    --- PASS: TestMuseAuthParentBoundaryAtResolve/config/muse/inside/resolve (0.51s)
    --- PASS: TestMuseAuthParentBoundaryAtResolve/config/muse/inside/resolve-repair (0.48s)
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
--- PASS: TestMuseAuthRepairAdditionalStates (4.52s)
    --- PASS: TestMuseAuthRepairAdditionalStates/empty-directory (0.82s)
    --- PASS: TestMuseAuthRepairAdditionalStates/nonempty-directory (0.64s)
    --- PASS: TestMuseAuthRepairAdditionalStates/relative-link (0.71s)
    --- PASS: TestMuseAuthRepairAdditionalStates/target-directory (0.67s)
    --- PASS: TestMuseAuthRepairAdditionalStates/readlink-error (0.67s)
    --- PASS: TestMuseAuthRepairAdditionalStates/missing-link-missing-target (0.49s)
    --- PASS: TestMuseAuthRepairAdditionalStates/config-parent-link (0.52s)
=== RUN   TestMuseGlobalReconciliationUsesExplicitProvisioning
--- PASS: TestMuseGlobalReconciliationUsesExplicitProvisioning (0.52s)
=== RUN   TestMuseStatusOptionalProvisioning
=== RUN   TestMuseStatusOptionalProvisioning/absent
=== RUN   TestMuseStatusOptionalProvisioning/marker-unreadable
=== RUN   TestMuseStatusOptionalProvisioning/backup-unreadable
=== RUN   TestMuseStatusOptionalProvisioning/marker-malformed
--- PASS: TestMuseStatusOptionalProvisioning (1.23s)
    --- PASS: TestMuseStatusOptionalProvisioning/absent (0.07s)
    --- PASS: TestMuseStatusOptionalProvisioning/marker-unreadable (0.08s)
    --- PASS: TestMuseStatusOptionalProvisioning/backup-unreadable (0.09s)
    --- PASS: TestMuseStatusOptionalProvisioning/marker-malformed (0.08s)
=== RUN   TestMusePublishedRegistryLayout
    muse_test.go:540: registry/layout production checks exercised; exec/serve permission mappings remain the separately owned launcher integration gap
--- PASS: TestMusePublishedRegistryLayout (0.52s)
PASS
ok  	github.com/relux-works/curator/internal/envprofile	82.722s
=== RUN   TestMuseFragmentBoundary
=== RUN   TestMuseFragmentBoundary/valid
=== RUN   TestMuseFragmentBoundary/missing-data
=== RUN   TestMuseFragmentBoundary/wrong-parent
=== RUN   TestMuseFragmentBoundary/HOME
=== RUN   TestMuseFragmentBoundary/channel
=== RUN   TestMuseFragmentBoundary/missing-policy
=== RUN   TestMuseFragmentBoundary/native-policy
=== RUN   TestMuseFragmentBoundary/missing-literal
=== RUN   TestMuseFragmentBoundary/keychain-literal
=== RUN   TestMuseFragmentBoundary/false-root
--- PASS: TestMuseFragmentBoundary (0.02s)
    --- PASS: TestMuseFragmentBoundary/valid (0.00s)
    --- PASS: TestMuseFragmentBoundary/missing-data (0.00s)
    --- PASS: TestMuseFragmentBoundary/wrong-parent (0.00s)
    --- PASS: TestMuseFragmentBoundary/HOME (0.00s)
    --- PASS: TestMuseFragmentBoundary/channel (0.00s)
    --- PASS: TestMuseFragmentBoundary/missing-policy (0.00s)
    --- PASS: TestMuseFragmentBoundary/native-policy (0.00s)
    --- PASS: TestMuseFragmentBoundary/missing-literal (0.00s)
    --- PASS: TestMuseFragmentBoundary/keychain-literal (0.00s)
    --- PASS: TestMuseFragmentBoundary/false-root (0.00s)
PASS
ok  	github.com/relux-works/curator/internal/envfragment	0.319s
=== RUN   TestMuseHasNoAdmittedRootContextForm
--- PASS: TestMuseHasNoAdmittedRootContextForm (0.00s)
PASS
ok  	github.com/relux-works/curator/internal/envregistry	0.289s
=== RUN   TestCredentialStoreConfigGrammar
=== RUN   TestCredentialStoreConfigGrammar/empty
=== RUN   TestCredentialStoreConfigGrammar/all_selectors
=== RUN   TestCredentialStoreConfigGrammar/null_map
=== RUN   TestCredentialStoreConfigGrammar/array_map
=== RUN   TestCredentialStoreConfigGrammar/null_profile
=== RUN   TestCredentialStoreConfigGrammar/invalid_profile
=== RUN   TestCredentialStoreConfigGrammar/unknown_id
=== RUN   TestCredentialStoreConfigGrammar/alias_id
=== RUN   TestCredentialStoreConfigGrammar/null_selector
=== RUN   TestCredentialStoreConfigGrammar/unknown_selector
=== RUN   TestCredentialStoreConfigGrammar/wrong_type
--- PASS: TestCredentialStoreConfigGrammar (0.01s)
    --- PASS: TestCredentialStoreConfigGrammar/empty (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/all_selectors (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/null_map (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/array_map (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/null_profile (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/invalid_profile (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/unknown_id (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/alias_id (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/null_selector (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/unknown_selector (0.00s)
    --- PASS: TestCredentialStoreConfigGrammar/wrong_type (0.00s)
=== RUN   TestCredentialStoreConfigVersionBoundary
--- PASS: TestCredentialStoreConfigVersionBoundary (0.00s)
=== RUN   TestCredentialStoreConfigMachineOwnership
--- PASS: TestCredentialStoreConfigMachineOwnership (0.00s)
PASS
ok  	github.com/relux-works/curator/internal/config	0.292s

```

## cli.log

```text
=== RUN   TestCredentialStoreCLIConfigRoundTrip
--- PASS: TestCredentialStoreCLIConfigRoundTrip (0.94s)
=== RUN   TestCredentialStoreCLIConfigRefusesMalformed
--- PASS: TestCredentialStoreCLIConfigRefusesMalformed (0.00s)
=== RUN   TestCredentialStoreCLIPromotion
--- PASS: TestCredentialStoreCLIPromotion (0.00s)
=== RUN   TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery
--- PASS: TestCredentialStoreCLIResolveRefusesUnqualifiedDelivery (0.95s)
=== RUN   TestReviewDottedCredentialProfile
=== RUN   TestReviewDottedCredentialProfile/team
=== RUN   TestReviewDottedCredentialProfile/team.prod
--- PASS: TestReviewDottedCredentialProfile (0.00s)
    --- PASS: TestReviewDottedCredentialProfile/team (0.00s)
    --- PASS: TestReviewDottedCredentialProfile/team.prod (0.00s)
=== RUN   TestReviewDottedCredentialProfileExisting
=== RUN   TestReviewDottedCredentialProfileExisting/show
=== RUN   TestReviewDottedCredentialProfileExisting/unset
--- PASS: TestReviewDottedCredentialProfileExisting (0.00s)
    --- PASS: TestReviewDottedCredentialProfileExisting/show (0.00s)
    --- PASS: TestReviewDottedCredentialProfileExisting/unset (0.00s)
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
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team/codex (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team/codex_cli (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team/claude (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team/claude_code (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team.prod/codex (0.01s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team.prod/codex_cli (0.01s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team.prod/claude (0.01s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/team.prod/claude_code (0.00s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/Team..prod/codex (0.01s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/Team..prod/codex_cli (0.03s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/Team..prod/claude (0.05s)
    --- PASS: TestCredentialStoreCLIDottedProfileRouting/Team..prod/claude_code (0.01s)
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
--- PASS: TestMuseCredentialPolicyNarrowingMutants (20.98s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/fragment-reobserves-muse-release (3.83s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/muse-keychain-admitted (2.50s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/missing-shared-source-admitted (2.68s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/isolated-muse-native-link (2.68s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/renderer-drops-only-backend (3.51s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/header-only-root-fact-dropped (3.97s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/boundary-admits-HOME (0.50s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/boundary-admits-missing-literal (0.68s)
    --- PASS: TestMuseCredentialPolicyNarrowingMutants/boundary-admits-false-root (0.62s)
=== RUN   TestMuseRootContextVectors
=== RUN   TestMuseRootContextVectors/monolithic
=== RUN   TestMuseRootContextVectors/zero-applicable
=== RUN   TestMuseRootContextVectors/no-context
=== RUN   TestMuseRootContextVectors/referenced-refused
--- PASS: TestMuseRootContextVectors (5.33s)
    --- PASS: TestMuseRootContextVectors/monolithic (1.41s)
    --- PASS: TestMuseRootContextVectors/zero-applicable (1.36s)
    --- PASS: TestMuseRootContextVectors/no-context (1.36s)
    --- PASS: TestMuseRootContextVectors/referenced-refused (1.20s)
=== RUN   TestMuseRootContextDrift
=== RUN   TestMuseRootContextDrift/changed
=== RUN   TestMuseRootContextDrift/removed
--- PASS: TestMuseRootContextDrift (3.08s)
    --- PASS: TestMuseRootContextDrift/changed (1.50s)
    --- PASS: TestMuseRootContextDrift/removed (1.57s)
=== RUN   TestMuseRootContextStatus
--- PASS: TestMuseRootContextStatus (0.76s)
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
--- PASS: TestMuseRootContextNarrowingMutants (26.30s)
    --- PASS: TestMuseRootContextNarrowingMutants/configured-form-ignored (3.34s)
    --- PASS: TestMuseRootContextNarrowingMutants/referenced-fallback (4.33s)
    --- PASS: TestMuseRootContextNarrowingMutants/no-context-admitted (3.62s)
    --- PASS: TestMuseRootContextNarrowingMutants/wrong-target (3.54s)
    --- PASS: TestMuseRootContextNarrowingMutants/wrong-bytes (3.55s)
    --- PASS: TestMuseRootContextNarrowingMutants/changed-root-admitted (3.79s)
    --- PASS: TestMuseRootContextNarrowingMutants/removed-root-admitted (4.12s)
=== RUN   TestMuseCLIFragmentAndFakeLaunch
--- PASS: TestMuseCLIFragmentAndFakeLaunch (1.12s)
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
--- PASS: TestMuseFragmentV3PublishedCases (1.34s)
    --- PASS: TestMuseFragmentV3PublishedCases/valid.json (0.49s)
    --- PASS: TestMuseFragmentV3PublishedCases/invalid.json (0.00s)
    --- PASS: TestMuseFragmentV3PublishedCases/valid-permissions-yolo-unlocked.json (0.04s)
    --- PASS: TestMuseFragmentV3PublishedCases/valid-permissions-native-locked.json (0.04s)
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
    --- PASS: TestMuseFragmentV3PublishedCases/invalid-wrong-parent-XDG_STATE_HOME.json (0.00s)
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
--- PASS: TestMuseCLIAuthParentBoundary (2.08s)
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
--- PASS: TestMuseFragmentV4PublishedCases (0.06s)
    --- PASS: TestMuseFragmentV4PublishedCases/valid-muse-root.json (0.01s)
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
    --- PASS: TestMuseFragmentV4PublishedCases/invalid-foreign-parent.json (0.01s)
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
--- PASS: TestMuseCLIStorePolicy (3.40s)
PASS
ok  	github.com/relux-works/curator/cmd/curator	66.868s

```

## snap.log

```text
{"mutant": "M-policy-probe", "test": "^TestMaterializeCaptureAndStatusNeverProbeVendor$", "exit": 1, "verdict": "killed", "failures": ["--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor (7.66s)", "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/codex (1.43s)", "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared (1.57s)", "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared (1.70s)", "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated (1.48s)", "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated (1.48s)"]}
{"mutant": "M-release-cache", "test": "^TestMaterializeCaptureCredentialPolicyRelease(Observation|Cache)$", "exit": 1, "verdict": "killed", "failures": ["--- FAIL: TestMaterializeCaptureCredentialPolicyReleaseObservation (1.05s)", "--- FAIL: TestMaterializeCaptureCredentialPolicyReleaseCache (0.29s)"]}
{"mutant": "M-probe", "test": "^TestMaterializeCaptureAndStatusNeverProbeVendor$", "exit": 1, "verdict": "killed", "failures": ["--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor (8.30s)", "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/codex (1.55s)", "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared (1.82s)", "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared (1.66s)", "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated (1.71s)", "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated (1.55s)"]}

```

## snap-remaining.log

```text
{"mutant": "M-codex-file-admission", "test": "^TestMaterializeCaptureCredentialStoreAdmission$", "exit": 1, "verdict": "killed", "failures": ["--- FAIL: TestMaterializeCaptureCredentialStoreAdmission (1.39s)", "--- FAIL: TestMaterializeCaptureCredentialStoreAdmission/file (0.79s)"]}
{"mutant": "M-denied-class", "test": "^TestMaterializeCaptureUnreadableJournal$", "exit": 1, "verdict": "killed", "failures": ["--- FAIL: TestMaterializeCaptureUnreadableJournal (0.70s)"]}
{"mutant": "M-lock-name", "test": "^TestMaterializeCaptureAuthorizingDrift/lock-identity$", "exit": 1, "verdict": "killed", "failures": ["--- FAIL: TestMaterializeCaptureAuthorizingDrift (0.84s)", "--- FAIL: TestMaterializeCaptureAuthorizingDrift/lock-identity (0.84s)"]}
{"mutant": "M-composition-policy", "test": "^TestMaterializeCaptureSameFragmentDrift/composition-policy$", "exit": 1, "verdict": "killed", "failures": ["--- FAIL: TestMaterializeCaptureSameFragmentDrift (0.71s)", "--- FAIL: TestMaterializeCaptureSameFragmentDrift/composition-policy (0.71s)"]}
{"mutant": "M3-lock-only", "test": "^TestMaterializeCaptureSameFragmentDrift/aba$", "exit": 1, "verdict": "killed", "failures": ["--- FAIL: TestMaterializeCaptureSameFragmentDrift (0.98s)", "--- FAIL: TestMaterializeCaptureSameFragmentDrift/aba (0.98s)"]}
{"mutant": "M-prepared", "test": "^TestMaterializeCapturePendingDoesNotRecover/prepared$", "exit": 1, "verdict": "killed", "failures": ["--- FAIL: TestMaterializeCapturePendingDoesNotRecover (0.89s)", "--- FAIL: TestMaterializeCapturePendingDoesNotRecover/prepared (0.89s)"]}
{"mutant": "M-corrupt-absence", "test": "^TestMaterializeCaptureBadRecoveryEvidence/corrupt$", "exit": 1, "verdict": "killed", "failures": ["--- FAIL: TestMaterializeCaptureBadRecoveryEvidence (0.77s)", "--- FAIL: TestMaterializeCaptureBadRecoveryEvidence/corrupt (0.77s)"]}
{"mutant": "M-executable-mode", "test": "^TestMaterializeCaptureSameFragmentDrift/mode$", "exit": 1, "verdict": "killed", "failures": ["--- FAIL: TestMaterializeCaptureSameFragmentDrift (0.67s)", "--- FAIL: TestMaterializeCaptureSameFragmentDrift/mode (0.67s)"]}
{"mutant": "M-provision-absent", "test": "^TestMaterializeCaptureDoesNotProvision$", "exit": 1, "verdict": "killed", "failures": ["--- FAIL: TestMaterializeCaptureDoesNotProvision (0.02s)", "--- FAIL: TestMaterializeCaptureDoesNotProvision/absent (0.01s)"]}

```

## snap-all-results.json

```text
[
  {
    "mutant": "M-policy-probe",
    "test": "^TestMaterializeCaptureAndStatusNeverProbeVendor$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor (7.66s)",
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/codex (1.43s)",
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared (1.57s)",
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared (1.70s)",
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated (1.48s)",
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated (1.48s)"
    ]
  },
  {
    "mutant": "M-release-cache",
    "test": "^TestMaterializeCaptureCredentialPolicyRelease(Observation|Cache)$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCaptureCredentialPolicyReleaseObservation (1.05s)",
      "--- FAIL: TestMaterializeCaptureCredentialPolicyReleaseCache (0.29s)"
    ]
  },
  {
    "mutant": "M-probe",
    "test": "^TestMaterializeCaptureAndStatusNeverProbeVendor$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor (8.30s)",
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/codex (1.55s)",
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared (1.82s)",
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared (1.66s)",
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated (1.71s)",
      "--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated (1.55s)"
    ]
  },
  {
    "mutant": "M-codex-file-admission",
    "test": "^TestMaterializeCaptureCredentialStoreAdmission$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCaptureCredentialStoreAdmission (1.39s)",
      "--- FAIL: TestMaterializeCaptureCredentialStoreAdmission/file (0.79s)"
    ]
  },
  {
    "mutant": "M-denied-class",
    "test": "^TestMaterializeCaptureUnreadableJournal$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCaptureUnreadableJournal (0.70s)"
    ]
  },
  {
    "mutant": "M-lock-name",
    "test": "^TestMaterializeCaptureAuthorizingDrift/lock-identity$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCaptureAuthorizingDrift (0.84s)",
      "--- FAIL: TestMaterializeCaptureAuthorizingDrift/lock-identity (0.84s)"
    ]
  },
  {
    "mutant": "M-composition-policy",
    "test": "^TestMaterializeCaptureSameFragmentDrift/composition-policy$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCaptureSameFragmentDrift (0.71s)",
      "--- FAIL: TestMaterializeCaptureSameFragmentDrift/composition-policy (0.71s)"
    ]
  },
  {
    "mutant": "M3-lock-only",
    "test": "^TestMaterializeCaptureSameFragmentDrift/aba$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCaptureSameFragmentDrift (0.98s)",
      "--- FAIL: TestMaterializeCaptureSameFragmentDrift/aba (0.98s)"
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
      "--- FAIL: TestMaterializeCaptureBadRecoveryEvidence (0.77s)",
      "--- FAIL: TestMaterializeCaptureBadRecoveryEvidence/corrupt (0.77s)"
    ]
  },
  {
    "mutant": "M-executable-mode",
    "test": "^TestMaterializeCaptureSameFragmentDrift/mode$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCaptureSameFragmentDrift (0.67s)",
      "--- FAIL: TestMaterializeCaptureSameFragmentDrift/mode (0.67s)"
    ]
  },
  {
    "mutant": "M-provision-absent",
    "test": "^TestMaterializeCaptureDoesNotProvision$",
    "exit": 1,
    "verdict": "killed",
    "failures": [
      "--- FAIL: TestMaterializeCaptureDoesNotProvision (0.02s)",
      "--- FAIL: TestMaterializeCaptureDoesNotProvision/absent (0.01s)"
    ]
  }
]

```

## free-hunt.log

```text
=== RUN   TestReviewMuseCaptureRefusalNoProbe
--- PASS: TestReviewMuseCaptureRefusalNoProbe (1.65s)
PASS
ok  	github.com/relux-works/curator/internal/envprofile	1.985s

```

## M-codex-file-admission.log

```text
--- FAIL: TestMaterializeCaptureCredentialStoreAdmission (1.39s)
    --- FAIL: TestMaterializeCaptureCredentialStoreAdmission/file (0.79s)
        materialize_capture_test.go:180: capture = &{[123 34 101 110 118 34 58 123 34 67 79 68 69 88 95 72 79 77 69 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 77 97 116 101 114 105 97 108 105 122 101 67 97 112 116 117 114 101 67 114 101 100 101 110 116 105 97 108 83 116 111 114 101 65 100 109 105 115 115 105 111 110 102 105 108 101 51 52 49 48 49 48 51 54 57 56 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 99 97 112 116 117 114 101 47 99 111 100 101 120 95 99 108 105 34 125 44 34 101 110 118 105 114 111 110 109 101 110 116 34 58 34 99 111 100 101 120 95 99 108 105 34 44 34 102 114 97 103 109 101 110 116 34 58 34 108 97 117 110 99 104 45 101 110 118 45 102 114 97 103 109 101 110 116 45 118 50 34 44 34 109 99 112 34 58 123 34 99 104 97 110 110 101 108 115 34 58 91 123 34 97 114 103 117 109 101 110 116 34 58 34 110 97 109 101 34 44 34 102 108 97 103 34 58 34 45 112 34 44 34 107 105 110 100 34 58 34 102 108 97 103 34 44 34 110 97 109 101 34 58 34 99 117 114 97 116 111 114 45 109 99 112 34 125 93 44 34 101 110 118 95 110 97 109 101 115 34 58 91 34 70 73 71 77 65 95 65 80 73 95 75 69 89 34 93 44 34 112 97 116 104 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 77 97 116 101 114 105 97 108 105 122 101 67 97 112 116 117 114 101 67 114 101 100 101 110 116 105 97 108 83 116 111 114 101 65 100 109 105 115 115 105 111 110 102 105 108 101 51 52 49 48 49 48 51 54 57 56 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 99 97 112 116 117 114 101 47 99 111 100 101 120 95 99 108 105 47 99 117 114 97 116 111 114 45 109 99 112 46 99 111 110 102 105 103 46 116 111 109 108 34 125 44 34 112 101 114 109 105 115 115 105 111 110 115 34 58 123 34 108 111 99 107 101 100 34 58 102 97 108 115 101 44 34 109 111 100 101 34 58 34 110 97 116 105 118 101 34 44 34 115 111 117 114 99 101 34 58 34 100 101 102 97 117 108 116 34 125 44 34 112 114 101 99 101 100 101 110 99 101 34 58 123 34 112 108 97 99 101 109 101 110 116 34 58 34 119 105 110 110 101 114 45 108 97 115 116 34 44 34 119 105 110 110 101 114 34 58 34 104 105 103 104 101 114 45 119 101 105 103 104 116 34 125 44 34 112 114 111 102 105 108 101 34 58 123 34 108 111 99 107 95 115 104 97 50 53 54 34 58 34 55 52 56 54 54 52 97 101 54 50 98 49 57 50 97 57 50 97 53 53 54 54 55 50 51 52 54 98 97 53 56 56 54 54 102 54 55 51 99 54 51 51 55 102 54 99 57 49 51 55 49 100 53 51 51 101 48 101 51 99 50 55 50 50 34 44 34 110 97 109 101 34 58 34 99 97 112 116 117 114 101 34 125 44 34 115 121 115 116 101 109 95 112 114 111 109 112 116 34 58 123 34 99 104 97 110 110 101 108 115 34 58 91 123 34 107 101 121 34 58 34 109 111 100 101 108 95 105 110 115 116 114 117 99 116 105 111 110 115 95 102 105 108 101 34 44 34 107 105 110 100 34 58 34 99 111 110 102 105 103 45 107 101 121 34 44 34 115 101 109 97 110 116 105 99 115 34 58 34 114 101 112 108 97 99 101 34 125 93 44 34 112 97 116 104 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 77 97 116 101 114 105 97 108 105 122 101 67 97 112 116 117 114 101 67 114 101 100 101 110 116 105 97 108 83 116 111 114 101 65 100 109 105 115 115 105 111 110 102 105 108 101 51 52 49 48 49 48 51 54 57 56 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 99 97 112 116 117 114 101 47 99 111 100 101 120 95 99 108 105 47 46 97 103 101 110 116 45 99 111 110 116 101 120 116 47 115 121 115 116 101 109 45 112 114 111 109 112 116 46 109 100 34 125 125 10] [60 33 45 45 10 99 117 114 97 116 111 114 45 114 111 111 116 45 99 111 110 116 101 120 116 45 118 50 10 114 111 111 116 58 32 99 97 112 116 117 114 101 32 49 46 48 46 48 32 115 116 97 116 101 32 115 104 97 50 53 54 58 98 57 50 53 101 55 97 100 54 52 97 99 102 98 51 53 55 53 101 101 102 51 100 101 53 99 55 49 51 57 101 49 98 52 55 98 99 55 54 101 50 51 100 51 53 55 53 52 99 53 52 52 99 101 100 51 50 98 102 100 49 102 56 53 10 109 101 109 98 101 114 58 32 99 97 112 116 117 114 101 32 49 46 48 46 48 32 115 116 97 116 101 32 115 104 97 50 53 54 58 98 57 50 53 101 55 97 100 54 52 97 99 102 98 51 53 55 53 101 101 102 51 100 101 53 99 55 49 51 57 101 49 98 52 55 98 99 55 54 101 50 51 100 51 53 55 53 52 99 53 52 52 99 101 100 51 50 98 102 100 49 102 56 53 32 119 101 105 103 104 116 32 49 48 48 10 112 114 101 99 101 100 101 110 99 101 58 32 119 105 110 110 101 114 61 104 105 103 104 101 114 45 119 101 105 103 104 116 32 112 108 97 99 101 109 101 110 116 61 119 105 110 110 101 114 45 108 97 115 116 10 108 111 99 107 58 32 115 104 97 50 53 54 58 55 52 56 54 54 52 97 101 54 50 98 49 57 50 97 57 50 97 53 53 54 54 55 50 51 52 54 98 97 53 56 56 54 54 102 54 55 51 99 54 51 51 55 102 54 99 57 49 51 55 49 100 53 51 51 101 48 101 51 99 50 55 50 50 10 103 101 110 101 114 97 116 101 100 58 32 67 117 114 97 116 111 114 32 80 114 111 116 111 99 111 108 32 101 110 118 105 114 111 110 109 101 110 116 115 32 114 101 118 105 115 105 111 110 32 49 32 40 104 116 116 112 115 58 47 47 103 105 116 104 117 98 46 99 111 109 47 114 101 108 117 120 45 119 111 114 107 115 47 99 117 114 97 116 111 114 45 115 112 101 99 41 10 110 111 116 105 99 101 58 32 103 101 110 101 114 97 116 101 100 32 102 105 108 101 59 32 100 105 114 101 99 116 32 101 100 105 116 115 32 97 114 101 32 117 110 115 117 112 112 111 114 116 101 100 32 97 110 100 32 97 114 101 32 100 101 116 101 99 116 101 100 32 97 115 32 100 114 105 102 116 59 32 117 112 100 97 116 101 32 116 104 101 32 115 111 117 114 99 101 32 112 114 111 102 105 108 101 32 114 101 112 111 115 105 116 111 114 121 32 111 114 32 105 116 115 32 99 111 109 112 111 115 101 100 32 112 114 111 102 105 108 101 115 32 105 110 115 116 101 97 100 10 45 45 62 10 10 45 45 45 10 10 35 35 32 67 111 110 116 101 120 116 58 32 99 97 112 116 117 114 101 32 49 46 48 46 48 10 10 104 101 108 108 111 10]}, <nil>, callback=true; want codex_settings_unsupported before capture
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.843s
FAIL

```

## M-composition-policy.log

```text
--- FAIL: TestMaterializeCaptureSameFragmentDrift (0.71s)
    --- FAIL: TestMaterializeCaptureSameFragmentDrift/composition-policy (0.71s)
        materialize_capture_test.go:457: capture = &{[123 34 101 110 118 34 58 123 34 67 79 68 69 88 95 72 79 77 69 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 77 97 116 101 114 105 97 108 105 122 101 67 97 112 116 117 114 101 83 97 109 101 70 114 97 103 109 101 110 116 68 114 105 102 116 99 111 109 112 111 115 105 116 105 111 110 45 112 111 108 105 99 121 49 56 54 57 55 54 56 52 49 51 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 99 97 112 116 117 114 101 47 99 111 100 101 120 95 99 108 105 34 125 44 34 101 110 118 105 114 111 110 109 101 110 116 34 58 34 99 111 100 101 120 95 99 108 105 34 44 34 102 114 97 103 109 101 110 116 34 58 34 108 97 117 110 99 104 45 101 110 118 45 102 114 97 103 109 101 110 116 45 118 50 34 44 34 109 99 112 34 58 123 34 99 104 97 110 110 101 108 115 34 58 91 123 34 97 114 103 117 109 101 110 116 34 58 34 110 97 109 101 34 44 34 102 108 97 103 34 58 34 45 112 34 44 34 107 105 110 100 34 58 34 102 108 97 103 34 44 34 110 97 109 101 34 58 34 99 117 114 97 116 111 114 45 109 99 112 34 125 93 44 34 101 110 118 95 110 97 109 101 115 34 58 91 34 70 73 71 77 65 95 65 80 73 95 75 69 89 34 93 44 34 112 97 116 104 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 77 97 116 101 114 105 97 108 105 122 101 67 97 112 116 117 114 101 83 97 109 101 70 114 97 103 109 101 110 116 68 114 105 102 116 99 111 109 112 111 115 105 116 105 111 110 45 112 111 108 105 99 121 49 56 54 57 55 54 56 52 49 51 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 99 97 112 116 117 114 101 47 99 111 100 101 120 95 99 108 105 47 99 117 114 97 116 111 114 45 109 99 112 46 99 111 110 102 105 103 46 116 111 109 108 34 125 44 34 112 101 114 109 105 115 115 105 111 110 115 34 58 123 34 108 111 99 107 101 100 34 58 102 97 108 115 101 44 34 109 111 100 101 34 58 34 110 97 116 105 118 101 34 44 34 115 111 117 114 99 101 34 58 34 100 101 102 97 117 108 116 34 125 44 34 112 114 101 99 101 100 101 110 99 101 34 58 123 34 112 108 97 99 101 109 101 110 116 34 58 34 119 105 110 110 101 114 45 108 97 115 116 34 44 34 119 105 110 110 101 114 34 58 34 104 105 103 104 101 114 45 119 101 105 103 104 116 34 125 44 34 112 114 111 102 105 108 101 34 58 123 34 108 111 99 107 95 115 104 97 50 53 54 34 58 34 98 51 99 98 97 55 57 51 56 54 52 55 101 54 55 52 50 100 48 102 102 55 54 48 57 102 48 99 48 97 50 55 51 49 53 55 102 98 57 101 56 99 51 98 51 49 101 54 100 51 52 53 99 54 52 98 54 97 99 100 51 97 53 101 34 44 34 110 97 109 101 34 58 34 99 97 112 116 117 114 101 34 125 44 34 115 121 115 116 101 109 95 112 114 111 109 112 116 34 58 123 34 99 104 97 110 110 101 108 115 34 58 91 123 34 107 101 121 34 58 34 109 111 100 101 108 95 105 110 115 116 114 117 99 116 105 111 110 115 95 102 105 108 101 34 44 34 107 105 110 100 34 58 34 99 111 110 102 105 103 45 107 101 121 34 44 34 115 101 109 97 110 116 105 99 115 34 58 34 114 101 112 108 97 99 101 34 125 93 44 34 112 97 116 104 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 77 97 116 101 114 105 97 108 105 122 101 67 97 112 116 117 114 101 83 97 109 101 70 114 97 103 109 101 110 116 68 114 105 102 116 99 111 109 112 111 115 105 116 105 111 110 45 112 111 108 105 99 121 49 56 54 57 55 54 56 52 49 51 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 99 97 112 116 117 114 101 47 99 111 100 101 120 95 99 108 105 47 46 97 103 101 110 116 45 99 111 110 116 101 120 116 47 115 121 115 116 101 109 45 112 114 111 109 112 116 46 109 100 34 125 125 10] [60 33 45 45 10 99 117 114 97 116 111 114 45 114 111 111 116 45 99 111 110 116 101 120 116 45 118 50 10 114 111 111 116 58 32 99 97 112 116 117 114 101 32 49 46 48 46 48 32 115 116 97 116 101 32 115 104 97 50 53 54 58 98 57 50 53 101 55 97 100 54 52 97 99 102 98 51 53 55 53 101 101 102 51 100 101 53 99 55 49 51 57 101 49 98 52 55 98 99 55 54 101 50 51 100 51 53 55 53 52 99 53 52 52 99 101 100 51 50 98 102 100 49 102 56 53 10 109 101 109 98 101 114 58 32 99 97 112 116 117 114 101 32 49 46 48 46 48 32 115 116 97 116 101 32 115 104 97 50 53 54 58 98 57 50 53 101 55 97 100 54 52 97 99 102 98 51 53 55 53 101 101 102 51 100 101 53 99 55 49 51 57 101 49 98 52 55 98 99 55 54 101 50 51 100 51 53 55 53 52 99 53 52 52 99 101 100 51 50 98 102 100 49 102 56 53 32 119 101 105 103 104 116 32 49 48 48 10 112 114 101 99 101 100 101 110 99 101 58 32 119 105 110 110 101 114 61 104 105 103 104 101 114 45 119 101 105 103 104 116 32 112 108 97 99 101 109 101 110 116 61 119 105 110 110 101 114 45 108 97 115 116 10 108 111 99 107 58 32 115 104 97 50 53 54 58 98 51 99 98 97 55 57 51 56 54 52 55 101 54 55 52 50 100 48 102 102 55 54 48 57 102 48 99 48 97 50 55 51 49 53 55 102 98 57 101 56 99 51 98 51 49 101 54 100 51 52 53 99 54 52 98 54 97 99 100 51 97 53 101 10 103 101 110 101 114 97 116 101 100 58 32 67 117 114 97 116 111 114 32 80 114 111 116 111 99 111 108 32 101 110 118 105 114 111 110 109 101 110 116 115 32 114 101 118 105 115 105 111 110 32 49 32 40 104 116 116 112 115 58 47 47 103 105 116 104 117 98 46 99 111 109 47 114 101 108 117 120 45 119 111 114 107 115 47 99 117 114 97 116 111 114 45 115 112 101 99 41 10 110 111 116 105 99 101 58 32 103 101 110 101 114 97 116 101 100 32 102 105 108 101 59 32 100 105 114 101 99 116 32 101 100 105 116 115 32 97 114 101 32 117 110 115 117 112 112 111 114 116 101 100 32 97 110 100 32 97 114 101 32 100 101 116 101 99 116 101 100 32 97 115 32 100 114 105 102 116 59 32 117 112 100 97 116 101 32 116 104 101 32 115 111 117 114 99 101 32 112 114 111 102 105 108 101 32 114 101 112 111 115 105 116 111 114 121 32 111 114 32 105 116 115 32 99 111 109 112 111 115 101 100 32 112 114 111 102 105 108 101 115 32 105 110 115 116 101 97 100 10 45 45 62 10 10 45 45 45 10 10 35 35 32 67 111 110 116 101 120 116 58 32 99 97 112 116 117 114 101 32 49 46 48 46 48 10 10 104 101 108 108 111 10]}, <nil>; want environment_materialize_capture_drift and no output
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.051s
FAIL

```

## M-corrupt-absence.log

```text
--- FAIL: TestMaterializeCaptureBadRecoveryEvidence (0.77s)
    --- FAIL: TestMaterializeCaptureBadRecoveryEvidence/corrupt (0.77s)
        materialize_capture_test.go:292: capture = &{[123 34 101 110 118 34 58 123 34 67 79 68 69 88 95 72 79 77 69 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 77 97 116 101 114 105 97 108 105 122 101 67 97 112 116 117 114 101 66 97 100 82 101 99 111 118 101 114 121 69 118 105 100 101 110 99 101 99 111 114 114 117 112 116 51 52 52 52 48 54 57 53 52 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 99 97 112 116 117 114 101 47 99 111 100 101 120 95 99 108 105 34 125 44 34 101 110 118 105 114 111 110 109 101 110 116 34 58 34 99 111 100 101 120 95 99 108 105 34 44 34 102 114 97 103 109 101 110 116 34 58 34 108 97 117 110 99 104 45 101 110 118 45 102 114 97 103 109 101 110 116 45 118 50 34 44 34 109 99 112 34 58 123 34 99 104 97 110 110 101 108 115 34 58 91 123 34 97 114 103 117 109 101 110 116 34 58 34 110 97 109 101 34 44 34 102 108 97 103 34 58 34 45 112 34 44 34 107 105 110 100 34 58 34 102 108 97 103 34 44 34 110 97 109 101 34 58 34 99 117 114 97 116 111 114 45 109 99 112 34 125 93 44 34 101 110 118 95 110 97 109 101 115 34 58 91 34 70 73 71 77 65 95 65 80 73 95 75 69 89 34 93 44 34 112 97 116 104 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 77 97 116 101 114 105 97 108 105 122 101 67 97 112 116 117 114 101 66 97 100 82 101 99 111 118 101 114 121 69 118 105 100 101 110 99 101 99 111 114 114 117 112 116 51 52 52 52 48 54 57 53 52 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 99 97 112 116 117 114 101 47 99 111 100 101 120 95 99 108 105 47 99 117 114 97 116 111 114 45 109 99 112 46 99 111 110 102 105 103 46 116 111 109 108 34 125 44 34 112 101 114 109 105 115 115 105 111 110 115 34 58 123 34 108 111 99 107 101 100 34 58 102 97 108 115 101 44 34 109 111 100 101 34 58 34 110 97 116 105 118 101 34 44 34 115 111 117 114 99 101 34 58 34 100 101 102 97 117 108 116 34 125 44 34 112 114 101 99 101 100 101 110 99 101 34 58 123 34 112 108 97 99 101 109 101 110 116 34 58 34 119 105 110 110 101 114 45 108 97 115 116 34 44 34 119 105 110 110 101 114 34 58 34 104 105 103 104 101 114 45 119 101 105 103 104 116 34 125 44 34 112 114 111 102 105 108 101 34 58 123 34 108 111 99 107 95 115 104 97 50 53 54 34 58 34 53 49 49 49 57 48 49 99 53 51 52 101 102 102 50 48 101 55 55 100 53 57 49 51 99 98 101 49 101 52 50 97 56 55 57 101 53 100 98 55 99 52 56 52 52 98 97 51 53 53 98 51 50 98 98 97 57 98 50 102 49 53 50 48 34 44 34 110 97 109 101 34 58 34 99 97 112 116 117 114 101 34 125 44 34 115 121 115 116 101 109 95 112 114 111 109 112 116 34 58 123 34 99 104 97 110 110 101 108 115 34 58 91 123 34 107 101 121 34 58 34 109 111 100 101 108 95 105 110 115 116 114 117 99 116 105 111 110 115 95 102 105 108 101 34 44 34 107 105 110 100 34 58 34 99 111 110 102 105 103 45 107 101 121 34 44 34 115 101 109 97 110 116 105 99 115 34 58 34 114 101 112 108 97 99 101 34 125 93 44 34 112 97 116 104 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 77 97 116 101 114 105 97 108 105 122 101 67 97 112 116 117 114 101 66 97 100 82 101 99 111 118 101 114 121 69 118 105 100 101 110 99 101 99 111 114 114 117 112 116 51 52 52 52 48 54 57 53 52 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 99 97 112 116 117 114 101 47 99 111 100 101 120 95 99 108 105 47 46 97 103 101 110 116 45 99 111 110 116 101 120 116 47 115 121 115 116 101 109 45 112 114 111 109 112 116 46 109 100 34 125 125 10] [60 33 45 45 10 99 117 114 97 116 111 114 45 114 111 111 116 45 99 111 110 116 101 120 116 45 118 50 10 114 111 111 116 58 32 99 97 112 116 117 114 101 32 49 46 48 46 48 32 115 116 97 116 101 32 115 104 97 50 53 54 58 98 57 50 53 101 55 97 100 54 52 97 99 102 98 51 53 55 53 101 101 102 51 100 101 53 99 55 49 51 57 101 49 98 52 55 98 99 55 54 101 50 51 100 51 53 55 53 52 99 53 52 52 99 101 100 51 50 98 102 100 49 102 56 53 10 109 101 109 98 101 114 58 32 99 97 112 116 117 114 101 32 49 46 48 46 48 32 115 116 97 116 101 32 115 104 97 50 53 54 58 98 57 50 53 101 55 97 100 54 52 97 99 102 98 51 53 55 53 101 101 102 51 100 101 53 99 55 49 51 57 101 49 98 52 55 98 99 55 54 101 50 51 100 51 53 55 53 52 99 53 52 52 99 101 100 51 50 98 102 100 49 102 56 53 32 119 101 105 103 104 116 32 49 48 48 10 112 114 101 99 101 100 101 110 99 101 58 32 119 105 110 110 101 114 61 104 105 103 104 101 114 45 119 101 105 103 104 116 32 112 108 97 99 101 109 101 110 116 61 119 105 110 110 101 114 45 108 97 115 116 10 108 111 99 107 58 32 115 104 97 50 53 54 58 53 49 49 49 57 48 49 99 53 51 52 101 102 102 50 48 101 55 55 100 53 57 49 51 99 98 101 49 101 52 50 97 56 55 57 101 53 100 98 55 99 52 56 52 52 98 97 51 53 53 98 51 50 98 98 97 57 98 50 102 49 53 50 48 10 103 101 110 101 114 97 116 101 100 58 32 67 117 114 97 116 111 114 32 80 114 111 116 111 99 111 108 32 101 110 118 105 114 111 110 109 101 110 116 115 32 114 101 118 105 115 105 111 110 32 49 32 40 104 116 116 112 115 58 47 47 103 105 116 104 117 98 46 99 111 109 47 114 101 108 117 120 45 119 111 114 107 115 47 99 117 114 97 116 111 114 45 115 112 101 99 41 10 110 111 116 105 99 101 58 32 103 101 110 101 114 97 116 101 100 32 102 105 108 101 59 32 100 105 114 101 99 116 32 101 100 105 116 115 32 97 114 101 32 117 110 115 117 112 112 111 114 116 101 100 32 97 110 100 32 97 114 101 32 100 101 116 101 99 116 101 100 32 97 115 32 100 114 105 102 116 59 32 117 112 100 97 116 101 32 116 104 101 32 115 111 117 114 99 101 32 112 114 111 102 105 108 101 32 114 101 112 111 115 105 116 111 114 121 32 111 114 32 105 116 115 32 99 111 109 112 111 115 101 100 32 112 114 111 102 105 108 101 115 32 105 110 115 116 101 97 100 10 45 45 62 10 10 45 45 45 10 10 35 35 32 67 111 110 116 101 120 116 58 32 99 97 112 116 117 114 101 32 49 46 48 46 48 10 10 104 101 108 108 111 10]}, <nil>; want invalid transaction journal and no output
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.167s
FAIL

```

## M-denied-class.log

```text
--- FAIL: TestMaterializeCaptureUnreadableJournal (0.70s)
    materialize_capture_test.go:892: capture = <nil>, open <fixture-path> permission denied; want manager_state_unreadable and no output
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.027s
FAIL

```

## M-executable-mode.log

```text
--- FAIL: TestMaterializeCaptureSameFragmentDrift (0.67s)
    --- FAIL: TestMaterializeCaptureSameFragmentDrift/mode (0.67s)
        materialize_capture_test.go:457: capture = &{[123 34 101 110 118 34 58 123 34 67 79 68 69 88 95 72 79 77 69 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 77 97 116 101 114 105 97 108 105 122 101 67 97 112 116 117 114 101 83 97 109 101 70 114 97 103 109 101 110 116 68 114 105 102 116 109 111 100 101 50 50 54 57 49 54 53 50 57 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 99 97 112 116 117 114 101 47 99 111 100 101 120 95 99 108 105 34 125 44 34 101 110 118 105 114 111 110 109 101 110 116 34 58 34 99 111 100 101 120 95 99 108 105 34 44 34 102 114 97 103 109 101 110 116 34 58 34 108 97 117 110 99 104 45 101 110 118 45 102 114 97 103 109 101 110 116 45 118 50 34 44 34 109 99 112 34 58 123 34 99 104 97 110 110 101 108 115 34 58 91 123 34 97 114 103 117 109 101 110 116 34 58 34 110 97 109 101 34 44 34 102 108 97 103 34 58 34 45 112 34 44 34 107 105 110 100 34 58 34 102 108 97 103 34 44 34 110 97 109 101 34 58 34 99 117 114 97 116 111 114 45 109 99 112 34 125 93 44 34 101 110 118 95 110 97 109 101 115 34 58 91 34 70 73 71 77 65 95 65 80 73 95 75 69 89 34 93 44 34 112 97 116 104 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 77 97 116 101 114 105 97 108 105 122 101 67 97 112 116 117 114 101 83 97 109 101 70 114 97 103 109 101 110 116 68 114 105 102 116 109 111 100 101 50 50 54 57 49 54 53 50 57 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 99 97 112 116 117 114 101 47 99 111 100 101 120 95 99 108 105 47 99 117 114 97 116 111 114 45 109 99 112 46 99 111 110 102 105 103 46 116 111 109 108 34 125 44 34 112 101 114 109 105 115 115 105 111 110 115 34 58 123 34 108 111 99 107 101 100 34 58 102 97 108 115 101 44 34 109 111 100 101 34 58 34 110 97 116 105 118 101 34 44 34 115 111 117 114 99 101 34 58 34 100 101 102 97 117 108 116 34 125 44 34 112 114 101 99 101 100 101 110 99 101 34 58 123 34 112 108 97 99 101 109 101 110 116 34 58 34 119 105 110 110 101 114 45 108 97 115 116 34 44 34 119 105 110 110 101 114 34 58 34 104 105 103 104 101 114 45 119 101 105 103 104 116 34 125 44 34 112 114 111 102 105 108 101 34 58 123 34 108 111 99 107 95 115 104 97 50 53 54 34 58 34 51 100 48 53 53 57 53 102 100 102 100 54 49 98 100 97 54 51 56 102 54 50 101 51 56 55 101 98 53 48 52 53 54 100 57 57 102 57 52 49 48 52 49 48 56 49 54 49 50 98 51 56 99 99 100 49 53 50 98 56 97 52 48 55 34 44 34 110 97 109 101 34 58 34 99 97 112 116 117 114 101 34 125 44 34 115 121 115 116 101 109 95 112 114 111 109 112 116 34 58 123 34 99 104 97 110 110 101 108 115 34 58 91 123 34 107 101 121 34 58 34 109 111 100 101 108 95 105 110 115 116 114 117 99 116 105 111 110 115 95 102 105 108 101 34 44 34 107 105 110 100 34 58 34 99 111 110 102 105 103 45 107 101 121 34 44 34 115 101 109 97 110 116 105 99 115 34 58 34 114 101 112 108 97 99 101 34 125 93 44 34 112 97 116 104 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 77 97 116 101 114 105 97 108 105 122 101 67 97 112 116 117 114 101 83 97 109 101 70 114 97 103 109 101 110 116 68 114 105 102 116 109 111 100 101 50 50 54 57 49 54 53 50 57 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 99 97 112 116 117 114 101 47 99 111 100 101 120 95 99 108 105 47 46 97 103 101 110 116 45 99 111 110 116 101 120 116 47 115 121 115 116 101 109 45 112 114 111 109 112 116 46 109 100 34 125 125 10] [60 33 45 45 10 99 117 114 97 116 111 114 45 114 111 111 116 45 99 111 110 116 101 120 116 45 118 50 10 114 111 111 116 58 32 99 97 112 116 117 114 101 32 49 46 48 46 48 32 115 116 97 116 101 32 115 104 97 50 53 54 58 98 57 50 53 101 55 97 100 54 52 97 99 102 98 51 53 55 53 101 101 102 51 100 101 53 99 55 49 51 57 101 49 98 52 55 98 99 55 54 101 50 51 100 51 53 55 53 52 99 53 52 52 99 101 100 51 50 98 102 100 49 102 56 53 10 109 101 109 98 101 114 58 32 99 97 112 116 117 114 101 32 49 46 48 46 48 32 115 116 97 116 101 32 115 104 97 50 53 54 58 98 57 50 53 101 55 97 100 54 52 97 99 102 98 51 53 55 53 101 101 102 51 100 101 53 99 55 49 51 57 101 49 98 52 55 98 99 55 54 101 50 51 100 51 53 55 53 52 99 53 52 52 99 101 100 51 50 98 102 100 49 102 56 53 32 119 101 105 103 104 116 32 49 48 48 10 112 114 101 99 101 100 101 110 99 101 58 32 119 105 110 110 101 114 61 104 105 103 104 101 114 45 119 101 105 103 104 116 32 112 108 97 99 101 109 101 110 116 61 119 105 110 110 101 114 45 108 97 115 116 10 108 111 99 107 58 32 115 104 97 50 53 54 58 51 100 48 53 53 57 53 102 100 102 100 54 49 98 100 97 54 51 56 102 54 50 101 51 56 55 101 98 53 48 52 53 54 100 57 57 102 57 52 49 48 52 49 48 56 49 54 49 50 98 51 56 99 99 100 49 53 50 98 56 97 52 48 55 10 103 101 110 101 114 97 116 101 100 58 32 67 117 114 97 116 111 114 32 80 114 111 116 111 99 111 108 32 101 110 118 105 114 111 110 109 101 110 116 115 32 114 101 118 105 115 105 111 110 32 49 32 40 104 116 116 112 115 58 47 47 103 105 116 104 117 98 46 99 111 109 47 114 101 108 117 120 45 119 111 114 107 115 47 99 117 114 97 116 111 114 45 115 112 101 99 41 10 110 111 116 105 99 101 58 32 103 101 110 101 114 97 116 101 100 32 102 105 108 101 59 32 100 105 114 101 99 116 32 101 100 105 116 115 32 97 114 101 32 117 110 115 117 112 112 111 114 116 101 100 32 97 110 100 32 97 114 101 32 100 101 116 101 99 116 101 100 32 97 115 32 100 114 105 102 116 59 32 117 112 100 97 116 101 32 116 104 101 32 115 111 117 114 99 101 32 112 114 111 102 105 108 101 32 114 101 112 111 115 105 116 111 114 121 32 111 114 32 105 116 115 32 99 111 109 112 111 115 101 100 32 112 114 111 102 105 108 101 115 32 105 110 115 116 101 97 100 10 45 45 62 10 10 45 45 45 10 10 35 35 32 67 111 110 116 101 120 116 58 32 99 97 112 116 117 114 101 32 49 46 48 46 48 10 10 104 101 108 108 111 10]}, <nil>; want environment_materialize_capture_drift and no output
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.036s
FAIL

```

## M-lock-name.log

```text
--- FAIL: TestMaterializeCaptureAuthorizingDrift (0.84s)
    --- FAIL: TestMaterializeCaptureAuthorizingDrift/lock-identity (0.84s)
        materialize_capture_test.go:860: capture = &{[123 34 101 110 118 34 58 123 34 67 79 68 69 88 95 72 79 77 69 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 77 97 116 101 114 105 97 108 105 122 101 67 97 112 116 117 114 101 65 117 116 104 111 114 105 122 105 110 103 68 114 105 102 116 108 111 99 107 45 105 100 101 110 116 105 116 121 51 51 53 52 55 56 56 49 55 55 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 99 97 112 116 117 114 101 47 99 111 100 101 120 95 99 108 105 34 125 44 34 101 110 118 105 114 111 110 109 101 110 116 34 58 34 99 111 100 101 120 95 99 108 105 34 44 34 102 114 97 103 109 101 110 116 34 58 34 108 97 117 110 99 104 45 101 110 118 45 102 114 97 103 109 101 110 116 45 118 50 34 44 34 109 99 112 34 58 123 34 99 104 97 110 110 101 108 115 34 58 91 123 34 97 114 103 117 109 101 110 116 34 58 34 110 97 109 101 34 44 34 102 108 97 103 34 58 34 45 112 34 44 34 107 105 110 100 34 58 34 102 108 97 103 34 44 34 110 97 109 101 34 58 34 99 117 114 97 116 111 114 45 109 99 112 34 125 93 44 34 101 110 118 95 110 97 109 101 115 34 58 91 34 70 73 71 77 65 95 65 80 73 95 75 69 89 34 93 44 34 112 97 116 104 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 77 97 116 101 114 105 97 108 105 122 101 67 97 112 116 117 114 101 65 117 116 104 111 114 105 122 105 110 103 68 114 105 102 116 108 111 99 107 45 105 100 101 110 116 105 116 121 51 51 53 52 55 56 56 49 55 55 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 99 97 112 116 117 114 101 47 99 111 100 101 120 95 99 108 105 47 99 117 114 97 116 111 114 45 109 99 112 46 99 111 110 102 105 103 46 116 111 109 108 34 125 44 34 112 101 114 109 105 115 115 105 111 110 115 34 58 123 34 108 111 99 107 101 100 34 58 102 97 108 115 101 44 34 109 111 100 101 34 58 34 110 97 116 105 118 101 34 44 34 115 111 117 114 99 101 34 58 34 100 101 102 97 117 108 116 34 125 44 34 112 114 101 99 101 100 101 110 99 101 34 58 123 34 112 108 97 99 101 109 101 110 116 34 58 34 119 105 110 110 101 114 45 108 97 115 116 34 44 34 119 105 110 110 101 114 34 58 34 104 105 103 104 101 114 45 119 101 105 103 104 116 34 125 44 34 112 114 111 102 105 108 101 34 58 123 34 108 111 99 107 95 115 104 97 50 53 54 34 58 34 52 53 52 54 51 57 102 57 54 51 57 99 53 97 102 101 98 54 100 98 97 48 56 55 49 55 55 97 99 48 97 102 55 102 57 54 102 100 54 48 53 102 57 99 100 101 51 98 49 49 57 52 53 50 97 56 97 100 51 51 101 50 52 99 34 44 34 110 97 109 101 34 58 34 99 97 112 116 117 114 101 34 125 44 34 115 121 115 116 101 109 95 112 114 111 109 112 116 34 58 123 34 99 104 97 110 110 101 108 115 34 58 91 123 34 107 101 121 34 58 34 109 111 100 101 108 95 105 110 115 116 114 117 99 116 105 111 110 115 95 102 105 108 101 34 44 34 107 105 110 100 34 58 34 99 111 110 102 105 103 45 107 101 121 34 44 34 115 101 109 97 110 116 105 99 115 34 58 34 114 101 112 108 97 99 101 34 125 93 44 34 112 97 116 104 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 77 97 116 101 114 105 97 108 105 122 101 67 97 112 116 117 114 101 65 117 116 104 111 114 105 122 105 110 103 68 114 105 102 116 108 111 99 107 45 105 100 101 110 116 105 116 121 51 51 53 52 55 56 56 49 55 55 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 99 97 112 116 117 114 101 47 99 111 100 101 120 95 99 108 105 47 46 97 103 101 110 116 45 99 111 110 116 101 120 116 47 115 121 115 116 101 109 45 112 114 111 109 112 116 46 109 100 34 125 125 10] [60 33 45 45 10 99 117 114 97 116 111 114 45 114 111 111 116 45 99 111 110 116 101 120 116 45 118 50 10 114 111 111 116 58 32 99 97 112 116 117 114 101 32 49 46 48 46 48 32 115 116 97 116 101 32 115 104 97 50 53 54 58 98 57 50 53 101 55 97 100 54 52 97 99 102 98 51 53 55 53 101 101 102 51 100 101 53 99 55 49 51 57 101 49 98 52 55 98 99 55 54 101 50 51 100 51 53 55 53 52 99 53 52 52 99 101 100 51 50 98 102 100 49 102 56 53 10 109 101 109 98 101 114 58 32 99 97 112 116 117 114 101 32 49 46 48 46 48 32 115 116 97 116 101 32 115 104 97 50 53 54 58 98 57 50 53 101 55 97 100 54 52 97 99 102 98 51 53 55 53 101 101 102 51 100 101 53 99 55 49 51 57 101 49 98 52 55 98 99 55 54 101 50 51 100 51 53 55 53 52 99 53 52 52 99 101 100 51 50 98 102 100 49 102 56 53 32 119 101 105 103 104 116 32 49 48 48 10 112 114 101 99 101 100 101 110 99 101 58 32 119 105 110 110 101 114 61 104 105 103 104 101 114 45 119 101 105 103 104 116 32 112 108 97 99 101 109 101 110 116 61 119 105 110 110 101 114 45 108 97 115 116 10 108 111 99 107 58 32 115 104 97 50 53 54 58 52 53 52 54 51 57 102 57 54 51 57 99 53 97 102 101 98 54 100 98 97 48 56 55 49 55 55 97 99 48 97 102 55 102 57 54 102 100 54 48 53 102 57 99 100 101 51 98 49 49 57 52 53 50 97 56 97 100 51 51 101 50 52 99 10 103 101 110 101 114 97 116 101 100 58 32 67 117 114 97 116 111 114 32 80 114 111 116 111 99 111 108 32 101 110 118 105 114 111 110 109 101 110 116 115 32 114 101 118 105 115 105 111 110 32 49 32 40 104 116 116 112 115 58 47 47 103 105 116 104 117 98 46 99 111 109 47 114 101 108 117 120 45 119 111 114 107 115 47 99 117 114 97 116 111 114 45 115 112 101 99 41 10 110 111 116 105 99 101 58 32 103 101 110 101 114 97 116 101 100 32 102 105 108 101 59 32 100 105 114 101 99 116 32 101 100 105 116 115 32 97 114 101 32 117 110 115 117 112 112 111 114 116 101 100 32 97 110 100 32 97 114 101 32 100 101 116 101 99 116 101 100 32 97 115 32 100 114 105 102 116 59 32 117 112 100 97 116 101 32 116 104 101 32 115 111 117 114 99 101 32 112 114 111 102 105 108 101 32 114 101 112 111 115 105 116 111 114 121 32 111 114 32 105 116 115 32 99 111 109 112 111 115 101 100 32 112 114 111 102 105 108 101 115 32 105 110 115 116 101 97 100 10 45 45 62 10 10 45 45 45 10 10 35 35 32 67 111 110 116 101 120 116 58 32 99 97 112 116 117 114 101 32 49 46 48 46 48 10 10 104 101 108 108 111 10]}, <nil>; want environment_materialize_capture_drift and no output
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.175s
FAIL

```

## M-policy-probe.log

```text
--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor (7.66s)
    --- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/codex (1.43s)
        materialize_capture_test.go:564: vendor executable ran
    --- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared (1.57s)
        materialize_capture_test.go:564: vendor executable ran
    --- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared (1.70s)
        materialize_capture_test.go:564: vendor executable ran
    --- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated (1.48s)
        materialize_capture_test.go:564: vendor executable ran
    --- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated (1.48s)
        materialize_capture_test.go:564: vendor executable ran
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	8.008s
FAIL

```

## M-prepared.log

```text
--- FAIL: TestMaterializeCapturePendingDoesNotRecover (0.89s)
    --- FAIL: TestMaterializeCapturePendingDoesNotRecover/prepared (0.89s)
        materialize_capture_test.go:246: capture = &{[123 34 101 110 118 34 58 123 34 67 79 68 69 88 95 72 79 77 69 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 77 97 116 101 114 105 97 108 105 122 101 67 97 112 116 117 114 101 80 101 110 100 105 110 103 68 111 101 115 78 111 116 82 101 99 111 118 101 114 112 114 101 112 97 114 101 100 51 53 50 54 51 51 52 49 56 49 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 99 97 112 116 117 114 101 47 99 111 100 101 120 95 99 108 105 34 125 44 34 101 110 118 105 114 111 110 109 101 110 116 34 58 34 99 111 100 101 120 95 99 108 105 34 44 34 102 114 97 103 109 101 110 116 34 58 34 108 97 117 110 99 104 45 101 110 118 45 102 114 97 103 109 101 110 116 45 118 50 34 44 34 109 99 112 34 58 123 34 99 104 97 110 110 101 108 115 34 58 91 123 34 97 114 103 117 109 101 110 116 34 58 34 110 97 109 101 34 44 34 102 108 97 103 34 58 34 45 112 34 44 34 107 105 110 100 34 58 34 102 108 97 103 34 44 34 110 97 109 101 34 58 34 99 117 114 97 116 111 114 45 109 99 112 34 125 93 44 34 101 110 118 95 110 97 109 101 115 34 58 91 34 70 73 71 77 65 95 65 80 73 95 75 69 89 34 93 44 34 112 97 116 104 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 77 97 116 101 114 105 97 108 105 122 101 67 97 112 116 117 114 101 80 101 110 100 105 110 103 68 111 101 115 78 111 116 82 101 99 111 118 101 114 112 114 101 112 97 114 101 100 51 53 50 54 51 51 52 49 56 49 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 99 97 112 116 117 114 101 47 99 111 100 101 120 95 99 108 105 47 99 117 114 97 116 111 114 45 109 99 112 46 99 111 110 102 105 103 46 116 111 109 108 34 125 44 34 112 101 114 109 105 115 115 105 111 110 115 34 58 123 34 108 111 99 107 101 100 34 58 102 97 108 115 101 44 34 109 111 100 101 34 58 34 110 97 116 105 118 101 34 44 34 115 111 117 114 99 101 34 58 34 100 101 102 97 117 108 116 34 125 44 34 112 114 101 99 101 100 101 110 99 101 34 58 123 34 112 108 97 99 101 109 101 110 116 34 58 34 119 105 110 110 101 114 45 108 97 115 116 34 44 34 119 105 110 110 101 114 34 58 34 104 105 103 104 101 114 45 119 101 105 103 104 116 34 125 44 34 112 114 111 102 105 108 101 34 58 123 34 108 111 99 107 95 115 104 97 50 53 54 34 58 34 52 50 51 57 57 49 49 51 57 55 50 50 54 54 101 57 56 54 51 55 100 98 48 97 100 52 56 98 48 100 48 97 53 101 99 98 48 52 99 101 53 98 52 97 97 97 100 101 101 57 102 97 54 99 49 102 54 48 56 49 48 102 101 48 34 44 34 110 97 109 101 34 58 34 99 97 112 116 117 114 101 34 125 44 34 115 121 115 116 101 109 95 112 114 111 109 112 116 34 58 123 34 99 104 97 110 110 101 108 115 34 58 91 123 34 107 101 121 34 58 34 109 111 100 101 108 95 105 110 115 116 114 117 99 116 105 111 110 115 95 102 105 108 101 34 44 34 107 105 110 100 34 58 34 99 111 110 102 105 103 45 107 101 121 34 44 34 115 101 109 97 110 116 105 99 115 34 58 34 114 101 112 108 97 99 101 34 125 93 44 34 112 97 116 104 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 77 97 116 101 114 105 97 108 105 122 101 67 97 112 116 117 114 101 80 101 110 100 105 110 103 68 111 101 115 78 111 116 82 101 99 111 118 101 114 112 114 101 112 97 114 101 100 51 53 50 54 51 51 52 49 56 49 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 99 97 112 116 117 114 101 47 99 111 100 101 120 95 99 108 105 47 46 97 103 101 110 116 45 99 111 110 116 101 120 116 47 115 121 115 116 101 109 45 112 114 111 109 112 116 46 109 100 34 125 125 10] [60 33 45 45 10 99 117 114 97 116 111 114 45 114 111 111 116 45 99 111 110 116 101 120 116 45 118 50 10 114 111 111 116 58 32 99 97 112 116 117 114 101 32 49 46 48 46 48 32 115 116 97 116 101 32 115 104 97 50 53 54 58 98 57 50 53 101 55 97 100 54 52 97 99 102 98 51 53 55 53 101 101 102 51 100 101 53 99 55 49 51 57 101 49 98 52 55 98 99 55 54 101 50 51 100 51 53 55 53 52 99 53 52 52 99 101 100 51 50 98 102 100 49 102 56 53 10 109 101 109 98 101 114 58 32 99 97 112 116 117 114 101 32 49 46 48 46 48 32 115 116 97 116 101 32 115 104 97 50 53 54 58 98 57 50 53 101 55 97 100 54 52 97 99 102 98 51 53 55 53 101 101 102 51 100 101 53 99 55 49 51 57 101 49 98 52 55 98 99 55 54 101 50 51 100 51 53 55 53 52 99 53 52 52 99 101 100 51 50 98 102 100 49 102 56 53 32 119 101 105 103 104 116 32 49 48 48 10 112 114 101 99 101 100 101 110 99 101 58 32 119 105 110 110 101 114 61 104 105 103 104 101 114 45 119 101 105 103 104 116 32 112 108 97 99 101 109 101 110 116 61 119 105 110 110 101 114 45 108 97 115 116 10 108 111 99 107 58 32 115 104 97 50 53 54 58 52 50 51 57 57 49 49 51 57 55 50 50 54 54 101 57 56 54 51 55 100 98 48 97 100 52 56 98 48 100 48 97 53 101 99 98 48 52 99 101 53 98 52 97 97 97 100 101 101 57 102 97 54 99 49 102 54 48 56 49 48 102 101 48 10 103 101 110 101 114 97 116 101 100 58 32 67 117 114 97 116 111 114 32 80 114 111 116 111 99 111 108 32 101 110 118 105 114 111 110 109 101 110 116 115 32 114 101 118 105 115 105 111 110 32 49 32 40 104 116 116 112 115 58 47 47 103 105 116 104 117 98 46 99 111 109 47 114 101 108 117 120 45 119 111 114 107 115 47 99 117 114 97 116 111 114 45 115 112 101 99 41 10 110 111 116 105 99 101 58 32 103 101 110 101 114 97 116 101 100 32 102 105 108 101 59 32 100 105 114 101 99 116 32 101 100 105 116 115 32 97 114 101 32 117 110 115 117 112 112 111 114 116 101 100 32 97 110 100 32 97 114 101 32 100 101 116 101 99 116 101 100 32 97 115 32 100 114 105 102 116 59 32 117 112 100 97 116 101 32 116 104 101 32 115 111 117 114 99 101 32 112 114 111 102 105 108 101 32 114 101 112 111 115 105 116 111 114 121 32 111 114 32 105 116 115 32 99 111 109 112 111 115 101 100 32 112 114 111 102 105 108 101 115 32 105 110 115 116 101 97 100 10 45 45 62 10 10 45 45 45 10 10 35 35 32 67 111 110 116 101 120 116 58 32 99 97 112 116 117 114 101 32 49 46 48 46 48 10 10 104 101 108 108 111 10]}, <nil>; want environment_materialize_recovery_required and no output
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.230s
FAIL

```

## M-probe.log

```text
--- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor (8.30s)
    --- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/codex (1.55s)
        materialize_capture_test.go:564: vendor executable ran
    --- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-shared (1.82s)
        materialize_capture_test.go:564: vendor executable ran
    --- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-shared (1.66s)
        materialize_capture_test.go:564: vendor executable ran
    --- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-native-isolated (1.71s)
        materialize_capture_test.go:564: vendor executable ran
    --- FAIL: TestMaterializeCaptureAndStatusNeverProbeVendor/muse-file-isolated (1.55s)
        materialize_capture_test.go:564: vendor executable ran
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	8.634s
FAIL

```

## M-provision-absent.log

```text
--- FAIL: TestMaterializeCaptureDoesNotProvision (0.02s)
    --- FAIL: TestMaterializeCaptureDoesNotProvision/absent (0.01s)
        materialize_capture_test.go:583: selected unprovisioned source
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	0.361s
FAIL

```

## M-release-cache.log

```text
--- FAIL: TestMaterializeCaptureCredentialPolicyReleaseObservation (1.05s)
    materialize_capture_test.go:976: invalid native selector reached capture
--- FAIL: TestMaterializeCaptureCredentialPolicyReleaseCache (0.29s)
    materialize_capture_test.go:1011: release unknown: policy=&{selection:native effectiveStore:native launchEnv:map[] isolation:shared nativeStore: links:map[] strategies:map[] records:[]} err=<nil>; want fresh isolated
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.667s
FAIL

```

## M3-lock-only.log

```text
--- FAIL: TestMaterializeCaptureSameFragmentDrift (0.98s)
    --- FAIL: TestMaterializeCaptureSameFragmentDrift/aba (0.98s)
        materialize_capture_test.go:457: capture = &{[123 34 101 110 118 34 58 123 34 67 79 68 69 88 95 72 79 77 69 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 77 97 116 101 114 105 97 108 105 122 101 67 97 112 116 117 114 101 83 97 109 101 70 114 97 103 109 101 110 116 68 114 105 102 116 97 98 97 52 48 50 57 49 55 56 49 52 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 99 97 112 116 117 114 101 47 99 111 100 101 120 95 99 108 105 34 125 44 34 101 110 118 105 114 111 110 109 101 110 116 34 58 34 99 111 100 101 120 95 99 108 105 34 44 34 102 114 97 103 109 101 110 116 34 58 34 108 97 117 110 99 104 45 101 110 118 45 102 114 97 103 109 101 110 116 45 118 50 34 44 34 109 99 112 34 58 123 34 99 104 97 110 110 101 108 115 34 58 91 123 34 97 114 103 117 109 101 110 116 34 58 34 110 97 109 101 34 44 34 102 108 97 103 34 58 34 45 112 34 44 34 107 105 110 100 34 58 34 102 108 97 103 34 44 34 110 97 109 101 34 58 34 99 117 114 97 116 111 114 45 109 99 112 34 125 93 44 34 101 110 118 95 110 97 109 101 115 34 58 91 34 70 73 71 77 65 95 65 80 73 95 75 69 89 34 93 44 34 112 97 116 104 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 77 97 116 101 114 105 97 108 105 122 101 67 97 112 116 117 114 101 83 97 109 101 70 114 97 103 109 101 110 116 68 114 105 102 116 97 98 97 52 48 50 57 49 55 56 49 52 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 99 97 112 116 117 114 101 47 99 111 100 101 120 95 99 108 105 47 99 117 114 97 116 111 114 45 109 99 112 46 99 111 110 102 105 103 46 116 111 109 108 34 125 44 34 112 101 114 109 105 115 115 105 111 110 115 34 58 123 34 108 111 99 107 101 100 34 58 102 97 108 115 101 44 34 109 111 100 101 34 58 34 110 97 116 105 118 101 34 44 34 115 111 117 114 99 101 34 58 34 100 101 102 97 117 108 116 34 125 44 34 112 114 101 99 101 100 101 110 99 101 34 58 123 34 112 108 97 99 101 109 101 110 116 34 58 34 119 105 110 110 101 114 45 108 97 115 116 34 44 34 119 105 110 110 101 114 34 58 34 104 105 103 104 101 114 45 119 101 105 103 104 116 34 125 44 34 112 114 111 102 105 108 101 34 58 123 34 108 111 99 107 95 115 104 97 50 53 54 34 58 34 98 52 54 53 53 98 100 51 54 54 50 53 52 49 51 101 53 99 50 52 99 49 55 54 102 99 101 97 55 53 100 51 99 52 97 100 52 54 51 98 53 56 56 51 55 54 54 53 99 54 98 99 99 56 99 56 100 49 54 102 51 56 102 101 34 44 34 110 97 109 101 34 58 34 99 97 112 116 117 114 101 34 125 44 34 115 121 115 116 101 109 95 112 114 111 109 112 116 34 58 123 34 99 104 97 110 110 101 108 115 34 58 91 123 34 107 101 121 34 58 34 109 111 100 101 108 95 105 110 115 116 114 117 99 116 105 111 110 115 95 102 105 108 101 34 44 34 107 105 110 100 34 58 34 99 111 110 102 105 103 45 107 101 121 34 44 34 115 101 109 97 110 116 105 99 115 34 58 34 114 101 112 108 97 99 101 34 125 93 44 34 112 97 116 104 34 58 34 47 118 97 114 47 102 111 108 100 101 114 115 47 99 122 47 106 113 98 116 104 116 107 115 53 53 122 98 107 112 99 100 107 102 121 107 52 98 108 56 48 48 48 48 103 112 47 84 47 84 101 115 116 77 97 116 101 114 105 97 108 105 122 101 67 97 112 116 117 114 101 83 97 109 101 70 114 97 103 109 101 110 116 68 114 105 102 116 97 98 97 52 48 50 57 49 55 56 49 52 47 48 48 49 47 101 110 118 105 114 111 110 109 101 110 116 115 47 99 97 112 116 117 114 101 47 99 111 100 101 120 95 99 108 105 47 46 97 103 101 110 116 45 99 111 110 116 101 120 116 47 115 121 115 116 101 109 45 112 114 111 109 112 116 46 109 100 34 125 125 10] [60 33 45 45 10 99 117 114 97 116 111 114 45 114 111 111 116 45 99 111 110 116 101 120 116 45 118 50 10 114 111 111 116 58 32 99 97 112 116 117 114 101 32 49 46 48 46 48 32 115 116 97 116 101 32 115 104 97 50 53 54 58 98 57 50 53 101 55 97 100 54 52 97 99 102 98 51 53 55 53 101 101 102 51 100 101 53 99 55 49 51 57 101 49 98 52 55 98 99 55 54 101 50 51 100 51 53 55 53 52 99 53 52 52 99 101 100 51 50 98 102 100 49 102 56 53 10 109 101 109 98 101 114 58 32 99 97 112 116 117 114 101 32 49 46 48 46 48 32 115 116 97 116 101 32 115 104 97 50 53 54 58 98 57 50 53 101 55 97 100 54 52 97 99 102 98 51 53 55 53 101 101 102 51 100 101 53 99 55 49 51 57 101 49 98 52 55 98 99 55 54 101 50 51 100 51 53 55 53 52 99 53 52 52 99 101 100 51 50 98 102 100 49 102 56 53 32 119 101 105 103 104 116 32 49 48 48 10 112 114 101 99 101 100 101 110 99 101 58 32 119 105 110 110 101 114 61 104 105 103 104 101 114 45 119 101 105 103 104 116 32 112 108 97 99 101 109 101 110 116 61 119 105 110 110 101 114 45 108 97 115 116 10 108 111 99 107 58 32 115 104 97 50 53 54 58 98 52 54 53 53 98 100 51 54 54 50 53 52 49 51 101 53 99 50 52 99 49 55 54 102 99 101 97 55 53 100 51 99 52 97 100 52 54 51 98 53 56 56 51 55 54 54 53 99 54 98 99 99 56 99 56 100 49 54 102 51 56 102 101 10 103 101 110 101 114 97 116 101 100 58 32 67 117 114 97 116 111 114 32 80 114 111 116 111 99 111 108 32 101 110 118 105 114 111 110 109 101 110 116 115 32 114 101 118 105 115 105 111 110 32 49 32 40 104 116 116 112 115 58 47 47 103 105 116 104 117 98 46 99 111 109 47 114 101 108 117 120 45 119 111 114 107 115 47 99 117 114 97 116 111 114 45 115 112 101 99 41 10 110 111 116 105 99 101 58 32 103 101 110 101 114 97 116 101 100 32 102 105 108 101 59 32 100 105 114 101 99 116 32 101 100 105 116 115 32 97 114 101 32 117 110 115 117 112 112 111 114 116 101 100 32 97 110 100 32 97 114 101 32 100 101 116 101 99 116 101 100 32 97 115 32 100 114 105 102 116 59 32 117 112 100 97 116 101 32 116 104 101 32 115 111 117 114 99 101 32 112 114 111 102 105 108 101 32 114 101 112 111 115 105 116 111 114 121 32 111 114 32 105 116 115 32 99 111 109 112 111 115 101 100 32 112 114 111 102 105 108 101 115 32 105 110 115 116 101 97 100 10 45 45 62 10 10 45 45 45 10 10 35 35 32 67 111 110 116 101 120 116 58 32 99 97 112 116 117 114 101 32 49 46 48 46 48 10 10 104 101 108 108 111 10]}, <nil>; want environment_materialize_capture_drift and no output
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.316s
FAIL

```
