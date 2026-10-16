# Revision 4 narrowing mutation evidence

Candidate snapshot: `3bdf4c4f77a89afc12c06c242f69cd3d16deda9e`. The isolated copy contains the current module code and its test/configuration inputs. Every mutant command runs directly using subprocess, without a pipeline, with `-count=1`. Files are restored from saved bytes after each plant.

| Mutant | Gate narrowed to | Named test that fails | Real exit / bound |
|---|---|---|---|
| M-policy-probe | Respect supplied release except for Codex, which probes the vendor | `TestMaterializeCaptureAndStatusNeverProbeVendor` | 1 expected red; killed |
| M-release-cache | Accept an unknown-release lookup from a different cached observation | `TestMaterializeCaptureCredentialPolicyReleaseObservation`; `TestMaterializeCaptureCredentialPolicyReleaseCache` | 1 expected red; killed |
| M-codex-file-admission | Admit explicit Codex file while retaining other selector refusals | `TestMaterializeCaptureCredentialStoreAdmission`; `TestMaterializeCaptureCredentialStoreAdmission/file` | 1 expected red; killed |
| M-denied-class | Return an unclassified permission failure for unreadable journal evidence | `TestMaterializeCaptureUnreadableJournal` | 1 expected red; killed |
| M-lock-name | Admit a substituted zero-length mutation-lock object | `TestMaterializeCaptureAuthorizingDrift`; `TestMaterializeCaptureAuthorizingDrift/lock-identity` | 1 expected red; killed |
| M-composition-policy | Admit a composition-policy change that only enables overlays | `TestMaterializeCaptureSameFragmentDrift`; `TestMaterializeCaptureSameFragmentDrift/composition-policy` | 1 expected red; killed |
| M3-lock-only | Retain fragment/profile-lock equality but skip other captured object identities | `TestMaterializeCaptureSameFragmentDrift`; `TestMaterializeCaptureSameFragmentDrift/aba` | 1 expected red; killed |
| M-prepared | Treat only prepared journals as not pending | `TestMaterializeCapturePendingDoesNotRecover`; `TestMaterializeCapturePendingDoesNotRecover/prepared` | 1 expected red; killed |
| M-corrupt-absence | Treat the corrupt broken journal as absent | `TestMaterializeCaptureBadRecoveryEvidence`; `TestMaterializeCaptureBadRecoveryEvidence/corrupt` | 1 expected red; killed |
| M-probe | Probe only Codex from the content verifier; preserve unknown token | `TestMaterializeCaptureAndStatusNeverProbeVendor` | 1 expected red; killed |
| M-executable-mode | Ignore only execute-bit drift while retaining other identity checks | `TestMaterializeCaptureSameFragmentDrift`; `TestMaterializeCaptureSameFragmentDrift/mode` | 1 expected red; killed |
| M-provision-absent | Provision only a previously absent manager home through lock acquisition | `TestMaterializeCaptureDoesNotProvision`; `TestMaterializeCaptureDoesNotProvision/absent` | 1 expected red; killed |

**12 of 12 killed; no survivors.** A timeout, build failure, skipped plant or failure without a named test would be unverified; none is counted here. Harness exit: 0. Initial harness copy was terminated before any tests (exit 143); the bounded replacement skipped unrelated board/doc payloads, keeping module source, tests and fixtures.

Exact commands, in plant order:

```text
go test ./internal/envprofile -run '^TestMaterializeCaptureAndStatusNeverProbeVendor$' -count=1 -timeout=2m  # exit 1
go test ./internal/envprofile -run '^TestMaterializeCaptureCredentialPolicyRelease(Observation|Cache)$' -count=1 -timeout=2m  # exit 1
go test ./internal/envprofile -run '^TestMaterializeCaptureCredentialStoreAdmission$' -count=1 -timeout=2m  # exit 1
go test ./internal/envprofile -run '^TestMaterializeCaptureUnreadableJournal$' -count=1 -timeout=2m  # exit 1
go test ./internal/envprofile -run '^TestMaterializeCaptureAuthorizingDrift/lock-identity$' -count=1 -timeout=2m  # exit 1
go test ./internal/envprofile -run '^TestMaterializeCaptureSameFragmentDrift/composition-policy$' -count=1 -timeout=2m  # exit 1
go test ./internal/envprofile -run '^TestMaterializeCaptureSameFragmentDrift/aba$' -count=1 -timeout=2m  # exit 1
go test ./internal/envprofile -run '^TestMaterializeCapturePendingDoesNotRecover/prepared$' -count=1 -timeout=2m  # exit 1
go test ./internal/envprofile -run '^TestMaterializeCaptureBadRecoveryEvidence/corrupt$' -count=1 -timeout=2m  # exit 1
go test ./internal/envprofile -run '^TestMaterializeCaptureAndStatusNeverProbeVendor$' -count=1 -timeout=2m  # exit 1
go test ./internal/envprofile -run '^TestMaterializeCaptureSameFragmentDrift/mode$' -count=1 -timeout=2m  # exit 1
go test ./internal/envprofile -run '^TestMaterializeCaptureDoesNotProvision$' -count=1 -timeout=2m  # exit 1
```

All plants are behavioral. The M-policy-probe and M-probe plants preserve the no-probe verifier and its unknown-release token while routing one adapter to actual executable discovery. No source-text checker is used as the acceptance gate.
