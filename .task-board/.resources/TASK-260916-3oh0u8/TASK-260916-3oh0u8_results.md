# TASK-260916-3oh0u8 results

Role: developer
Candidate: Story worktree on trunk, base d41da0f; no commit made.
Spec pin: curator-spec 23435129ebc4c29e5b7f75ec72a0aa0cd3f16065 (rc.13), matching SPEC_PIN at .github/workflows/ci.yml:43.
Normative clauses: environments.md sections 11, 11.1, 12, 12.1, 12.2; profiles/manager.md section 1.

## Changes in this candidate

- cmd/curator/umbrella.go: revision B now refuses a candidate whose symlink target resolves outside all trust roots; refusal does not fall through to a later root. Revision A's provider_directories hint uses the canonical target directory. The optional PATH override is a CLI test seam; production uses host PATH.
- cmd/curator/main.go and cmd/curator/envstatus.go: curator status now reports provider rows in text and JSON, and --check fails for missing, refused, unreadable, or otherwise failed rows. Shared config-to-posture construction keeps env status and curator status aligned.
- cmd/curator/umbrella_test.go and cmd/curator/status_test.go: added symlink-escape coverage and top-level status coverage for warning, missing, unreadable, and --check behavior.
- .github/ci/platform-cases.tsv: registered the new resolver, symlink, and status tests.
- internal/config already contains the knob implementation on trunk; no config files were changed here. The implementation is at internal/config/environments.go:338, 731-761, 1120, 1138 and internal/config/config.go:66. It parses unique absolute POSIX or Windows-drive paths, writes the default empty list, and includes the system-lockable key.

## Rules, vectors, and production entries

| Rule | Production entry and evidence |
| --- | --- |
| Revision A keeps PATH selection and warns with the resolved path, roots, and provider_directories hint. | resolveRevisionA and warningText in cmd/curator/umbrella.go:136-152, 367-369. Pinned vector suite includes the S6 planted PATH case; end-to-end A warning test passed. |
| Revision B selects install directory then provider_directories and never dispatches a PATH-only candidate. | resolveRevisionB in cmd/curator/umbrella.go:162-194. All 14 published cases ran against both revisions (28 resolver comparisons); S6 PATH plant was refused under B and warned under A. |
| Published/managed candidates are refused; unreadable roots are not absence or fallback; missing and refusal outcomes stay distinct. | cmd/curator/umbrella.go:137-147, 162-194. Published, managed, missing, and unreadable vector cases passed; unreadable and refused provider posture rows are non-current. |
| A symlink inside a trust root cannot escape it under B. | New containment check at cmd/curator/umbrella.go:176-183 and regression at cmd/curator/umbrella_test.go:134-177. A still selects and warns; B refuses before trying a later root. |
| Status reports every discovered provider and always run/session, including missing and unreadable outcomes. | env status uses providerPostureForConfig at cmd/curator/envstatus.go:163-183. curator status emits rows and checks currency at cmd/curator/main.go:838-844, 908-914, 948-963. CLI test at cmd/curator/status_test.go:1826-1914 covers warning, missing, unreadable, and --check. |
| Config schema is consumed from the pinned root. | TestManagerConfigV2SchemaCases and TestSystemConfigV2SchemaCases at internal/config/environments_conformance_test.go:102-132 index and execute all cases. The root includes valid-provider-directories, valid-provider-directories-windows-drive, invalid-provider-directories-relative, and invalid-provider-directories-duplicate manager cases, plus the corresponding system negative cases. |

The rc.13 root publishes the provider vectors, so this candidate does not use a root-content skip for that family. TestUmbrellaProviderResolutionVectors skips only when CURATOR_CONFORMANCE_ROOT is unset; with the pinned root it fails if the vector file or cases are absent. The ledger already records 14 cases at .github/ci/conformance-case-counts.tsv:61.

## Conformance-gap ledger

Baseline d41da0f: 69 data rows, 0 owned by STORY-260916-2otjbn or TASK-260916-3oh0u8.
Candidate: 69 data rows, 0 owned by this Story/task.
No owned row passed and therefore none required removal. Other owners' rows were left unchanged.

## Narrowing mutants

All overlays were temporary files under TMPDIR; no mutants were written into the worktree.

| Mutant | Before/new gate evidence |
| --- | --- |
| Remove the B canonical-target containment check. | Existing vector/selection/refusal mask survived: exit 0. The new TestUmbrellaTrustedRootSymlinkEscape then failed on B dispatching the escaped symlink: exit 1. This demonstrates the old coverage gap and the new regression kill. |
| Weaken the revision-A outside-roots warning to silent resolution. | TestUmbrellaProviderResolutionVectors failed on warning cases: exit 1. |
| Dispatch the revision-B PATH probe result. | TestUmbrellaProviderResolutionVectors failed on the S6 plant and other PATH-only/refused cases: exit 1. |
| Treat unreadable trust roots as absence. | TestUmbrellaUnreadableRootNeverAbsence failed because revision A warned from PATH: exit 1. |
| Remove manager-published/managed directory refusal. | TestUmbrellaRefusedDirectoriesBothRevisions and TestUmbrellaSymlinkIntoManagedRefused failed: exit 1. |
| Keep status rows but suppress --check failure for non-current providers. | TestCuratorStatusProviderPostureAndCheck failed because missing rows returned exit 0 instead of exit 1; mutant test command exited 1. |

Overlay setup attempts that did not represent behavioral mutants also exited 1 (unused local variables in two initial overlays and one incorrect overlay filename); the overlays were corrected before the kills above and those setup failures are not counted as gate evidence.

## Local validation transcripts

Passing commands:

- CURATOR_CONFORMANCE_ROOT=/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T//TASK-260916-3oh0u8-rc13.tVXCPS/root/conformance/v1 go test ./cmd/curator -run '^(TestUmbrellaTrustedRootSymlinkEscape|TestUmbrellaProviderResolutionVectors|TestCuratorStatusProviderPostureAndCheck|TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands|TestUmbrellaHostilePathPlantWarnsUnderAEndToEnd|TestUmbrellaHostilePathPlantRefusedUnderB|TestUmbrellaRefusedDirectoriesBothRevisions|TestUmbrellaUnreadableRootNeverAbsence|TestProviderPostureRows)$' -count=1 — exit 0.
- CURATOR_CONFORMANCE_ROOT=/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T//TASK-260916-3oh0u8-rc13.tVXCPS/root/conformance/v1 go test ./cmd/curator -run '^TestUmbrellaProviderResolutionVectors$' -count=1 -v — exit 0; 14 driven, 0 known-gap, 0 bound, 0 skipped, 14 total; each case compared A and B.
- CURATOR_CONFORMANCE_ROOT=/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T//TASK-260916-3oh0u8-rc13.tVXCPS/root/conformance/v1 go test ./internal/config -run 'TestManagerConfigV2SchemaCases|TestSystemConfigV2SchemaCases|TestManagerConfigV2Vectors' -count=1 — exit 0.
- CURATOR_CONFORMANCE_ROOT=/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T//TASK-260916-3oh0u8-rc13.tVXCPS/root/conformance/v1 go test ./internal/config -run '^(TestManagerConfigV2SchemaCases|TestSystemConfigV2SchemaCases)$' -count=1 -v — exit 0; manager schema 78 driven / 29 known-gap / 0 skipped of 107; system schema 36 driven / 6 known-gap / 0 skipped of 42. Provider-directory cases passed; remaining known gaps belong to other findings.
- go test ./internal/config/... — exit 0.
- go test ./internal/config -run '^TestProviderDirectories' -count=1 — exit 0.
- go test ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$' -count=1 — exit 0.
- go build ./... — exit 0.
- go vet ./... — exit 0.
- gofmt -l cmd internal — exit 0, no output.
- golangci-lint run — exit 0, 0 issues.
- git diff --check — exit 0.

Additional commands and limits:

- CURATOR_CONFORMANCE_ROOT=/var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T//TASK-260916-3oh0u8-rc13.tVXCPS/root/conformance/v1 go test ./cmd/curator/... — exit 1: package test timed out at 10 minutes in unrelated TestDraftTransportIdentityInvalidSurfacesThroughCLI while fingerprintToolchain/digestToolchainRecords read the host Go toolchain. No provider assertion failure was reported. The focused provider/status suite above passed afterward; the full package suite remains locally unverified.
- The updated status test first waited on the package host-GOROOT lock held by another concurrent run and exited 1 after 301.623 seconds. After that run exited, the focused suite passed. No lock was bypassed.
- gofmt -l . — exit 0, but listed 24 archived Go files under .task-board/.resources. None were candidate source files; gofmt -l cmd internal was clean.
- Early focused iterations exited 1 for a Darwin /var versus /private/var expected-path spelling and two status-test fixture-order/config-version mistakes. The assertions/fixture were corrected; final focused command exits 0 as recorded above.

## Handoff notes and release preparation

The archived refs/campaign/2otjbn-full-20260927 patch was inspected as reference and not applied. The current task resources contain no standalone revision-1 review-verdict artifact; the attached hold and revision-1 validation history explain the old rc.11 pin lag. Current SPEC_PIN is rc.13.

No CHANGELOG.md or LOGBOOK.md edit was made, per the newer campaign rules. Findings, decisions, and anomalies are recorded here.

## CHANGELOG entry (for release prep)

E4: ship the warning release (revision A), keeping PATH selection while warning with subcommand_provider_outside_trust_roots, the resolved provider path, consulted trust roots, and a provider_directories migration hint. Revision B will refuse PATH-only providers in a later release. This is the warn-first step for the proposal 0016 path_prepend dependency.

The hosted gate for this new candidate has not run yet; the task-board handoff/runtime gate is the arbiter.

## Revision 3 — re-apply on eca2bf27

Re-applied the accepted rev2 patch from `refs/campaign/2otjbn-rev2-20260927` onto the fresh trunk base `eca2bf27`. The two expected conflicts were resolved by retaining both sides: trunk's registry boundary posture and rev2's provider posture in `cmd/curator/envstatus.go` and `cmd/curator/main.go`. No code outside those two conflict resolutions was introduced. The other four changed paths (`.github/ci/platform-cases.tsv`, `cmd/curator/status_test.go`, `cmd/curator/umbrella.go`, `cmd/curator/umbrella_test.go`) compare byte-for-byte with rev2 (`git diff --quiet refs/campaign/2otjbn-rev2-20260927 -- <four paths>`: exit 0). The worktree delta is exactly the six rev2 paths; no commit was made.

The attached `TASK-260916-3oh0u8_review-verdict-rev2.md` currently says **ACCEPTED** and names `TestUmbrellaTrustedRootSymlinkEscape`; the review-round note says rev2 was rejected. I found no rejection rationale in the attached verdict. To answer the requested regression/mutant evidence, I reran that named test and performed a temporary narrowing mutant against its canonical-target check. The mutant changed the containment input from the resolved canonical provider path to the lexical symlink path; the test then failed because revision B dispatched the escaped provider (mutant command exit 1). I restored the source byte-for-byte, reran the regression green, and verified `cmd/curator/umbrella.go` still matches rev2.

### Revision 3 validation transcripts

- `go build ./...` — exit 0 (after restoring the mutant).
- `env CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./cmd/curator -run '^TestUmbrellaProviderResolutionVectors$' -count=1 -v` — exit 0; 14 driven, 0 known-gap, 0 bound, 0 skipped; all 14 cases compared both revisions.
- `go test ./cmd/curator -run '^TestUmbrella(HostilePathPlantWarnsUnderAEndToEnd|HostilePathPlantRefusedUnderB|TrustedRootSymlinkEscape)$' -count=1` — exit 0.
- `go test ./cmd/curator -run '^TestCuratorStatusProviderPostureAndCheck$' -count=1` — exit 0.
- `go test ./cmd/curator -run '^TestUmbrellaTrustedRootSymlinkEscape$' -count=1` — exit 1 while the temporary lexical-path mutant was present (expected kill); exit 0 after restoring the accepted implementation.
- The requested broad mask `env CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./cmd/curator -run 'Umbrella|Provider|EnvStatus|Status|Boundary|Attest'` — exit 1 after Go's 10-minute test timeout in `TestEnvStatusMissingAndUnreadableKeepRecord`, inside envprofile/contextlock canonical JSON processing. No provider assertion failed before timeout. The provider vector, hostile PATH, symlink, and status-provider tests were rerun in the bounded masks above and passed.
- `git diff --check HEAD -- . ':!.task-board'` — exit 0. The four non-conflicting paths compare byte-for-byte with rev2 (exit 0). Final changed-path listing: `.github/ci/platform-cases.tsv`, `cmd/curator/envstatus.go`, `cmd/curator/main.go`, `cmd/curator/status_test.go`, `cmd/curator/umbrella.go`, `cmd/curator/umbrella_test.go`.

This re-apply changed no `CHANGELOG.md`, `LOGBOOK.md`, configuration, spec pin, or conformance ledger. The hosted gate for this re-applied candidate has not yet run; handoff will publish it for the configured gate.
