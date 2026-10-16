# SNAP-LOCK revision 3 narrowing controls

Harness: `python3 .temp/SNAPLOCK-rev3/run-mutants.py`, exit 0. Every child command is standalone `go test ./internal/envprofile -run PATTERN -count=1 -timeout=2m`; each has real exit 1 at a named behavioral failure (expected red). No survivor. The first exploratory admission run used an earlier assertion order; the reported evidence is the second run, which fails the capture assertion before checking Resolve.

| Mutant | Gate narrowed to | Named failing test | Surviving bound |
| --- | --- | --- | --- |
| M-codex-file-admission | Admit explicit Codex file only; native and every other selector retain their gates | TestMaterializeCaptureCredentialStoreAdmission; TestMaterializeCaptureCredentialStoreAdmission/file | None; killed |
| M-denied-class | Read errors still refuse except permission denial | TestMaterializeCaptureUnreadableJournal | None; killed |
| M-lock-name | Allow a zero-length regular lock-name replacement | TestMaterializeCaptureAuthorizingDrift; TestMaterializeCaptureAuthorizingDrift/lock-identity | None; killed |
| M-composition-policy | Allow composition policy enablement alone | TestMaterializeCaptureSameFragmentDrift; TestMaterializeCaptureSameFragmentDrift/composition-policy | None; killed |
| M3-lock-only | Retain profile-lock object validation but omit other captured object identities | TestMaterializeCaptureSameFragmentDrift; TestMaterializeCaptureSameFragmentDrift/aba | None; killed |
| M-prepared | Refuse every pending journal phase except prepared | TestMaterializeCapturePendingDoesNotRecover; TestMaterializeCapturePendingDoesNotRecover/prepared | None; killed |
| M-corrupt-absence | Treat the planted broken journal as absent | TestMaterializeCaptureBadRecoveryEvidence; TestMaterializeCaptureBadRecoveryEvidence/corrupt | None; killed |
| M-probe | Allow only Codex vendor discovery | TestMaterializeCaptureAndStatusNeverProbeVendor | None; killed |
| M-executable-mode | Ignore only execute permission bits in object mode equality | TestMaterializeCaptureSameFragmentDrift; TestMaterializeCaptureSameFragmentDrift/mode | None; killed |
| M-provision-absent | Provision absent manager homes only | TestMaterializeCaptureDoesNotProvision; TestMaterializeCaptureDoesNotProvision/absent | None; killed |

## Exact commands and actual exits

`go test ./internal/envprofile -run '^TestMaterializeCaptureCredentialStoreAdmission$' -count=1 -timeout=2m`: exit 1; expected failure at the named refusal assertion.

`go test ./internal/envprofile -run '^TestMaterializeCaptureUnreadableJournal$' -count=1 -timeout=2m`: exit 1; expected failure at the named refusal assertion.

`go test ./internal/envprofile -run '^TestMaterializeCaptureAuthorizingDrift/lock-identity$' -count=1 -timeout=2m`: exit 1; expected failure at the named refusal assertion.

`go test ./internal/envprofile -run '^TestMaterializeCaptureSameFragmentDrift/composition-policy$' -count=1 -timeout=2m`: exit 1; expected failure at the named refusal assertion.

`go test ./internal/envprofile -run '^TestMaterializeCaptureSameFragmentDrift/aba$' -count=1 -timeout=2m`: exit 1; expected failure at the named refusal assertion.

`go test ./internal/envprofile -run '^TestMaterializeCapturePendingDoesNotRecover/prepared$' -count=1 -timeout=2m`: exit 1; expected failure at the named refusal assertion.

`go test ./internal/envprofile -run '^TestMaterializeCaptureBadRecoveryEvidence/corrupt$' -count=1 -timeout=2m`: exit 1; expected failure at the named refusal assertion.

`go test ./internal/envprofile -run '^TestMaterializeCaptureAndStatusNeverProbeVendor$' -count=1 -timeout=2m`: exit 1; expected failure at the named refusal assertion.

`go test ./internal/envprofile -run '^TestMaterializeCaptureSameFragmentDrift/mode$' -count=1 -timeout=2m`: exit 1; expected failure at the named refusal assertion.

`go test ./internal/envprofile -run '^TestMaterializeCaptureDoesNotProvision$' -count=1 -timeout=2m`: exit 1; expected failure at the named refusal assertion.

No static source-text gate was introduced. M-probe retains the verification function/token and changes behavior; all plants execute the behavioral Go suite. Plants run on darwin/arm64; no Windows mutation execution is claimed. The harness stages an isolated task-local copy and restores every mutated file.
