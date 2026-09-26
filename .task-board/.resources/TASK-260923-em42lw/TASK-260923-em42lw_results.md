# TASK-260923-em42lw results

## Implementation

- Added `environments.permissions` parsing for portable profile keys and `native`/`yolo` values. System files accept only `native`; the system lock replaces the machine map and preserves the manager warning when it overrides a machine value.
- Threaded the effective map and system lock into environment resolution. `env resolve --format json` now emits `launch-env-fragment-v2` with the closed `permissions` member. A configured profile emits `source: profile`; silence emits native/default; a system lock emits native/global/locked.
- Added production-entry rows for profile yolo, profile native, absent/default, and locked/global. Added parser validity/refusal rows, schema cases, docs, and an Unreleased changelog entry. No launcher files changed.

## Normative fixture

Source checkout: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec`, commit `eadb1c06480f4438f01b2b0a973caf775a188f41` (after `ec8dc656`).

The copied schema is at `internal/envfragment/testdata/curator-spec-main/schemas/v1/launch-env-fragment-v2.schema.json`; SHA-256 `4d26b5e2a89452eb3c7fe6945f19dd16dda557382be15972ab9c57cb02d38d40`. The test verifies this hash, registers its copied dependencies, validates all 14 copied schema cases, and validates every production-entry JSON output. The copied cases matched the source directory byte-for-byte. Fixture README: `internal/envfragment/testdata/curator-spec-main/README.md`.

## Passing verification

All commands below were run as standalone processes from the Story worktree; each exited 0.

| Command | Result |
| --- | --- |
| `go test ./internal/config -count=1` | exit 0 |
| `go test ./internal/envfragment -count=1` | exit 0 |
| `CURATOR_TEST_CLI_HELPER=1 go test ./cmd/curator -run '^TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases$' -count=1` | exit 0; all copied valid/invalid cases |
| `CURATOR_TEST_CLI_HELPER=1 go test ./cmd/curator -run '^TestEnvResolveEmitsPermissionsV2AtProductionEntry$' -count=1` | exit 0; all four production rows |
| `CURATOR_TEST_CLI_HELPER=1 go test ./cmd/curator -run '^TestEnvResolveRepairEmitsFragment$' -count=1` | exit 0 |
| `go test ./internal/envprofile -run '^(TestResolveProvisionRepair|TestResolveFormats|TestResolveDefaultProfile)$' -count=1` | exit 0 |
| `go test ./internal/envregistry -count=1` | exit 0 |
| `go build ./cmd/curator` | exit 0 |
| `golangci-lint run` | exit 0, 0 issues |
| `git diff --check` | exit 0 |
| `gofmt -l cmd/curator/env.go cmd/curator/env_test.go cmd/curator/env_permissions_test.go internal/config/config.go internal/config/environments.go internal/config/environments_test.go internal/envfragment/envfragment.go internal/envfragment/envfragment_test.go internal/envfragment/fragment_schema_test.go internal/envprofile/managed.go internal/envregistry/envregistry.go` | exit 0, no paths listed |

The `cmd/curator` test package normally holds a host-Go-root serialization lock in `TestMain`. The focused tests use the package's helper-process mode to skip only that unrelated lock; the tests still call the production `run` entry and production resolver.

## Lattice narrowing mutants

Each mutant was applied temporarily to the resolver, run through the named production-entry row, and removed. Each command exited 1 as expected, proving the row rejects the narrowed behavior.

| Rule | Narrowing mutant | Killed by |
| --- | --- | --- |
| `locked` iff `source == global` | Emit `locked: false` for the locked/global branch | `permissions-system-lock-global` |
| Native for global/default provenance | Emit `mode: yolo` for the locked/global branch | `permissions-system-lock-global` |
| Yolo only with profile provenance | Emit `source: default` for a configured profile | `permissions-yolo-profile` |

Each mutant invocation was `CURATOR_TEST_CLI_HELPER=1 go test ./cmd/curator -run '<production test>/<named row>' -count=1`.

## Non-final attempts and their actual exits

- Initial `go test ./internal/config` attempts exited 1: first a compile mismatch in a new parser test, then acceptance of explicit `permissions: null`. Both issues were fixed; the final parser suite exited 0.
- An initial combined `cmd/curator` schema/production filter exited 1 because the schema validator's default Go regexp engine rejected the schema's ECMAScript lookahead pattern. The test now supplies an ECMAScript-compatible engine; the final schema suite exited 0.
- The first production matrix and a follow-up yolo row each exited 1 because the test configured profile `default` while profile installation selected `acme`. The fixture now configures `acme`; the final four-row matrix exited 0.
- An unfiltered `go test ./internal/envprofile` ended with exit 143 while another Story held the shared host-Go-root test lock. A lock-queued `cmd/curator` schema attempt was interrupted and exited 1. Focused envprofile tests and the CLI production/schema rows were rerun successfully.

The hosted landing gate was not run locally; `task-board handoff TASK-260923-em42lw --role developer` runs it once.

## Base validation

Fresh `origin/main` resolved to `48da2690fe79ddb24eff078c5efb13d4869aa6a9`, matching the Story worktree's recorded selected base at start. No commit was created.

## Revision 2

The revision-1 hosted ledger gate found that `.github/ci/platform-cases.tsv`
required `internal/envfragment :: TestFragmentEmissionMatchesReference`, but
the candidate had removed that test when switching the reference from v1 to
v2. Restored the exact top-level test name against the copied curator-spec
main v2 `valid-permissions-yolo-unlocked.json` fixture. It has no
`CURATOR_CONFORMANCE_ROOT` skip, so it executes wherever the package is built.
The ledger check below confirmed that it is compiled for linux, darwin, and
windows.

### Revision 2 verification

Commands were run directly from the Story worktree; the exit code is stated
for each command.

| Command | Exit | Result |
| --- | ---: | --- |
| `go test ./internal/envfragment -list '.*'` | 0 | Prints the required exact name `TestFragmentEmissionMatchesReference`. |
| `go test ./internal/envfragment -run '^TestFragmentEmissionMatchesReference$' -count=1` | 0 | Reference test executed and passed. |
| `go test ./internal/envfragment -race -count=1` | 0 | Package race run passed. |
| `sh .github/ci/ledger-consistency.sh` | 2 | Usage error: this script requires an evidence-directory argument; not counted as a passing gate. |
| `sh .github/ci/ledger-consistency.sh .temp/ci-evidence/ledger` | 0 | 241 ledger rows consistent across linux, darwin, and windows; the restored row is compiled on all three. |
| `sh .github/ci/gate-selftest.sh` | 0 | 185 passed, 0 failed. |
| `CURATOR_TEST_CLI_HELPER=1 go test ./cmd/curator -run '^TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases$' -count=1` | 0 | Copied spec-main v2 schema cases passed. |
| `CURATOR_TEST_CLI_HELPER=1 go test ./cmd/curator -run '^TestEnvResolveEmitsPermissionsV2AtProductionEntry$' -count=1` | 0 | Four production-entry permission rows passed. |
| `go test ./internal/config -count=1` | 0 | Full config package passed. |
| `go test ./internal/envregistry -count=1` | 0 | Full envregistry package passed. |
| `go test ./internal/envprofile -run '^(TestResolveProvisionRepair|TestResolveFormats|TestResolveDefaultProfile)$' -count=1` | 0 | Relevant resolver tests passed. |
| `go build ./...` | 0 | All packages built. |
| `golangci-lint run` | 0 | 0 issues. |
| `gofmt -l` on all changed Go files | 0 | No paths listed. |
| `git diff --check` | 0 | No whitespace errors. |

### Revision 2 evidence bounds

- The no-argument ledger invocation above is an actual exit 2 caused by the
  script's required evidence-directory argument. The workflow's intended
  invocation was then run and passed; no nonzero gate is reported as green.
- The schema source hash and copied-case identity are unchanged from Revision
  1 and remain recorded above. The schema-case and production-entry tests were
  rerun in this revision.
- The three narrowing-mutant outcomes recorded above were accepted from the
  already-attached Revision 1 evidence. No resolver or lattice code changed
  in this rework; Revision 2 restores only the missing reference test.
- The full landing suite is left to `task-board handoff`, which runs the
  hosted CI gate once as specified by the task brief.

## Revision 3

### Revision 2 rework: manager-config vector compatibility

Updated `internal/config/environments_conformance_test.go`. The vector harness reads the checked out conformance root revision and logs the manager-config vector counts. If a vector's expected `environments` object omits `permissions`, it accepts the effective empty object `{}` only when every other environment member still compares equal. A non-empty or non-object `permissions` value remains part of the comparison and fails. Roots whose expected object contains `permissions` use the unchanged exact comparison. Every accepted forward difference logs once for its vector and names the root revision.

### Revision 3 verification

Commands were run directly from the Story worktree; each exit code is stated.

| Command | Exit | Result |
| --- | ---: | --- |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260923-1lu2o3/worktree/.temp/TASK-260923-em42lw-pre0018/conformance/v1 go test -count=1 -run '^TestManagerConfigV2Vectors$' -v ./internal/config` | 0 | Pre-0018 root `dced9b8317e0`: 48 vectors, 22 valid and 26 invalid. All passed; the empty default was logged exactly once for each of the 22 valid vectors. |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260923-1lu2o3/worktree/.temp/TASK-260923-em42lw-pre0018/conformance/v1-mutated go test -count=1 -run '^TestManagerConfigV2Vectors/schema2-minimal-defaults$' ./internal/config` | 1 | Expected rejection proof: injected `permissions.companyA=yolo` into the pre-0018 minimal-default vector; the environment comparison failed on that non-default value. |
| `go test -count=1 -run '^TestManagerConfigV2VectorDefaultPermissionsAccommodationIsNarrow$' ./internal/config` | 0 | Empty default accepted; non-default permissions and an unrelated environment-member change rejected; a current-root expected permissions member did not use the accommodation. |
| `go test -count=1 -run '^TestManagerConfigV2VectorDefaultPermissionsAccommodationIsNarrow$' -overlay=.temp/TASK-260923-em42lw-pre0018/overlay-mutant.json ./internal/config` | 1 | Expected mutant failure: the disposable mutant accepted any present `permissions` value; the test failed because yolo was treated as the empty default. |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test -count=1 -run '^TestManagerConfigV2Vectors$' -v ./internal/config` | 1 | Current local main root `fe2d1c6470a0`: 56 vectors, 25 valid and 31 invalid. There were no forward-difference logs, but the full suite does not pass: this root also expects newer `require_source_signers`, `source_signers`, and audit registry bootstrap members that this permissions-only candidate does not implement. The targeted `schema2-permissions` row also exits 1 because those additional expected fields differ; the emitted permissions value itself matches. This is outside Revision 2's permissions-default accommodation and would require unrelated curator features or a compatible schema root. |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260923-1lu2o3/worktree/.temp/TASK-260923-em42lw-pre0018/conformance/v1 go test -count=1 ./internal/config/...` | 0 | Full config package suite passed against the pre-0018 root. |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260923-1lu2o3/worktree/.temp/TASK-260923-em42lw-pre0018/conformance/v1 go test -race -count=1 ./internal/config/...` | 0 | Full config package race run passed against the pre-0018 root. |
| `go vet ./internal/config/...` | 0 | Passed. |
| `golangci-lint run` | 0 | 0 issues. |
| `gofmt -l internal/config/environments_conformance_test.go` | 0 | No paths listed. |
| `git diff --check` | 0 | No whitespace errors. |

The current-main vector run is recorded as failing, not passing. Its newer requirements come from later curator-spec work (`source_signers` and registry trust) and exceed the assigned rework; no broad comparison relaxation or unrelated product implementation was added. The landing suite remains delegated to `task-board handoff` and will run there once.

## Revision 4 — refresh onto aa093918

Replayed the task delta onto the clean Story worktree at `aa093918`. The
`git apply --3way` reapply exited 1 because of one conflict in
`internal/config/environments_conformance_test.go`; I resolved it by keeping
trunk's conformance driver. No other files conflicted. The final tracked diff
is limited to this task's implementation, tests, dependencies, and docs; no
`CHANGELOG.md` or `LOGBOOK.md` edit is present.

The pinned `curator-spec` root is `dcc7f015e2d97edf2d52928afb6fd79ec8129e8b`.
I ran `TestManagerConfigV2Vectors` against that exact root without the
pre-0018 permissions accommodation. The vector test passes, and the pinned
vectors include effective `permissions` values, so the accommodation and its
test rows were removed by retaining trunk's exact comparison. The old
pre-0018 vector-root behavior is historical evidence only.

Permissions parsing accepts only `native` and `yolo` per profile and rejects
invalid profile/value/object shapes. The system-file permissions lock is
lockable, forces `native` for every profile, and keeps the manager §1 warning
when it replaces a differing machine value. The production `env resolve
--format json` path emits v2's required closed `permissions` object for
profile yolo, profile native, silent default, and an engaged global lock.

The v2 schema and fourteen valid/invalid cases are copied under
`internal/envfragment/testdata/curator-spec-main`. The schema hash is
`4d26b5e2a89452eb3c7fe6945f19dd16dda557382be15972ab9c57cb02d38d40`; byte
comparisons passed against the local `curator-spec` main checkout for the six
schemas and fourteen launch-env-fragment-v2 cases. Tests compile that schema,
check its published valid/invalid cases, and validate each production-entry
fragment.

### Revision 4 verification

Commands were run directly from this Story worktree. Each exit code is the
process exit code.

- Exit **0** — pinned manager-config-v2 vectors pass with trunk's strict
  comparison and no accommodation:

  ```sh
  CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260923-1lu2o3/worktree/.task-spec-pin-dcc7f015/conformance/v1 go test ./internal/config -run '^TestManagerConfigV2Vectors$' -count=1
  ```

- Exit **0** — required bounded production-entry, config-vector/schema, and
  fragment checks:

  ```sh
  CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260923-1lu2o3/worktree/.task-spec-pin-dcc7f015/conformance/v1 go test ./internal/config/... ./internal/envprofile ./cmd/curator -run 'ManagerConfigV2|Fragment|EnvResolve|Permissions' -count=1
  ```

- Exit **0** — config permission/vector checks under the race detector:

  ```sh
  CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260923-1lu2o3/worktree/.task-spec-pin-dcc7f015/conformance/v1 go test -race ./internal/config/... -run 'ManagerConfigV2|Permissions' -count=1
  ```

- Exit **0** — `go test ./internal/envfragment -run 'Fragment' -count=1`.
- Exit **0** — `go vet ./internal/config/... ./internal/envprofile ./internal/envfragment ./internal/envregistry ./cmd/curator`.
- Exit **0** — `GOOS=windows go vet ./internal/config ./cmd/curator`.
- Exit **0** — `go build ./cmd/curator`; the generated local binary was removed afterward.
- Exit **0** — `make lint`; `golangci-lint run` reported 0 issues.
- Exit **0** — `git diff --check`; no whitespace errors.
- Exit **0** — a `set -e` byte-comparison loop checked all six copied schema JSON files and all fourteen v2 case JSON files against the local `curator-spec` main checkout.
- Exit **0** — `git diff --name-only HEAD`; the tracked diff listed only task implementation, test, dependency, and documentation paths.

The first two attempts to run the pinned vector test with a relative
`CURATOR_CONFORMANCE_ROOT` exited 1: the first archive contained only the
manifest, and the second path resolved relative to `internal/config`. After
archiving `conformance/v1/**` and using the absolute Story-worktree path, the
exact pinned vector command above passed. Those setup failures are not
reported as passing gates.

### Revision 4 lattice narrowing mutants

Each mutant was a temporary one-line production change in
`internal/envprofile/managed.go`, restored before the passing verification
commands above. The intended nonzero result means the production-entry test
killed the mutant.

| Mutant | Test command | Exit | Failure observed |
| --- | --- | ---: | --- |
| `source=global` with `locked=false` | `go test ./cmd/curator -run '^TestEnvResolveEmitsPermissionsV2AtProductionEntry/permissions-system-lock-global$' -count=1` | 1 | Expected `locked=true`, got false. |
| `source=global` with `mode=yolo` | `go test ./cmd/curator -run '^TestEnvResolveEmitsPermissionsV2AtProductionEntry/permissions-system-lock-global$' -count=1` | 1 | Expected native, got yolo. |
| Absent profile knob with `source=profile` | `go test ./cmd/curator -run '^TestEnvResolveEmitsPermissionsV2AtProductionEntry/permissions-absent-default$' -count=1` | 1 | Expected default provenance, got profile. |
| `source=default` with `mode=yolo` | `go test ./cmd/curator -run '^TestEnvResolveEmitsPermissionsV2AtProductionEntry/permissions-absent-default$' -count=1` | 1 | Expected native, got yolo. |

The CHANGELOG entry text for landing is: “Emit launch-env-fragment-v2 with
effective per-profile permissions and system-lock provenance.” Per the
Revision 3 policy, this text is recorded here and no `CHANGELOG.md` edit was
made. No logbook entry was needed or made.

### Revision 4 — refresh onto aa093918: this-turn reapplication and verification

The worktree started at `aa093918aa78436245bf27f0697b2d4aa1a1fcb3` on the Story branch, with the task already set to `development` (the explicit status mutation returned exit 0). The saved-delta command was run exactly as directed with `set -o pipefail`; `git apply --3way` returned **1** because the task changes were already present in the worktree. It reported index mismatches for the existing edits in `cmd/curator/env.go`, `cmd/curator/env_test.go`, `docs/environment-config.md`, `go.mod`, `go.sum`, `internal/config/config.go`, `internal/config/environments.go`, `internal/config/environments_test.go`, `internal/envfragment/envfragment.go`, `internal/envfragment/envfragment_test.go`, `internal/envfragment/fragment_schema_test.go`, `internal/envprofile/managed.go`, and `internal/envregistry/envregistry.go`. `cmd/curator/env_permissions_test.go` and the v2 schema fixture tree were already untracked task files. The patch also reported a conflict for `internal/config/environments_conformance_test.go`; I kept the `aa093918` version, which carries trunk's 0018 conformance coverage and strict vector comparison. After application there were no unmerged index stages or conflict markers. No pre-0018 permissions accommodation or its test rows remain.

I archived `conformance/v1` directly from curator-spec commit `dcc7f015e2d97edf2d52928afb6fd79ec8129e8b` inside the Story worktree and ran the strict `TestManagerConfigV2Vectors` against that root without accommodation. The test passed. An initial run without `CURATOR_CONFORMANCE_ROOT` exited 0 but skipped the conformance test, so it is not counted as evidence. I also compared the copied v2 fixture bytes against the exact dcc7f015 archive: all 6 schema files and 14 schema cases matched. The first comparison attempt used unavailable `cmp` and exited 127; the equivalent Python byte comparison passed with exit 0.

All bounded validation below was run directly from this Story worktree; the pinned conformance root was supplied where needed so the conformance tests executed:

| Command | Exit | Result |
| --- | ---: | --- |
| `env CURATOR_CONFORMANCE_ROOT="$PWD/.tmp-spec-dcc7/conformance/v1" go test ./internal/config/... -run '^TestManagerConfigV2Vectors$' -count=1` | 0 | Strict 0018 vectors passed against dcc7f015 without the pre-0018 accommodation. |
| `env CURATOR_CONFORMANCE_ROOT="$PWD/.tmp-spec-dcc7/conformance/v1" go test ./internal/config/... ./internal/envprofile ./cmd/curator -run 'ManagerConfigV2|Fragment|EnvResolve|Permissions' -count=1` | 0 | All selected packages passed; config and envprofile passed, then cmd/curator passed. |
| `env CURATOR_CONFORMANCE_ROOT="$PWD/.tmp-spec-dcc7/conformance/v1" go test -race ./internal/config/... -run 'ManagerConfigV2|Permissions' -count=1` | 0 | Config race run passed. |
| `go vet ./internal/config/... ./internal/envprofile ./internal/envfragment ./internal/envregistry ./cmd/curator` | 0 | Passed. |
| `env GOOS=windows go vet ./internal/config ./cmd/curator` | 0 | Passed. |
| `go build ./cmd/curator` | 0 | Passed; generated binary was removed. |
| `make lint` | 0 | `golangci-lint run`, 0 issues. |
| `env CURATOR_TEST_CLI_HELPER=1 go test ./cmd/curator -run '^TestEnvResolveEmitsPermissionsV2AtProductionEntry$' -count=1` | 0 | All four production-entry rows passed after restoring the mutants. |
| Python byte comparison of copied dcc7f015 schemas and v2 cases | 0 | 6 schemas and 14 cases byte-identical. |
| `gofmt -l` on changed Go files | 0 | No paths listed. |
| `git diff --check` | 0 | Passed. |

### Revision 4 revalidated lattice mutants

The production-entry test killed one narrowing mutant for each lattice rule. Each one-line mutation was restored immediately after its test; all three expected-red commands returned **1**:

| Rule | Mutant | Production-entry assertion |
| --- | --- | --- |
| `locked` iff `source=global` | global branch emits `locked=false` | `permissions-system-lock-global` expected locked true. |
| `mode=native` for global/default source | silent default emits `mode=yolo` | `permissions-absent-default` expected native. |
| `source=default` means profile level is silent | silent default emits `source=profile` | `permissions-absent-default` expected default. |

Each was run with `env CURATOR_TEST_CLI_HELPER=1 go test ./cmd/curator -run '<test>/<row>' -count=1`. The helper flag skips only the host-Go-root serialization lock; the test still drives the production `run` command and resolver.

Final worktree checks showed no staged paths and no unmerged entries. `git diff --name-only HEAD` contains only task-owned tracked paths; `CHANGELOG.md` and `LOGBOOK.md` are untouched. The changelog entry text remains: “Emit launch-env-fragment-v2 with effective per-profile permissions and system-lock provenance.” The temporary dcc7 fixture and build binary were removed. The configured landing suite is left to the developer handoff, which runs it once.

### Revision 6 — gate fix

This revision addresses the hosted gate regressions on curator-spec pin
`dcc7f015e2d97edf2d52928afb6fd79ec8129e8b`. The spec tree was archived
from the local curator-spec checkout into the ignored Story-worktree
`.temp/pinned-spec-dcc7f015`; every conformance command below used its
`conformance/v1` directory.

Before the fix, I reproduced both reported failures with their real exit
codes:

| Command | Exit | Observed failure |
| --- | ---: | --- |
| `CURATOR_CONFORMANCE_ROOT="$PWD/.temp/pinned-spec-dcc7f015/conformance/v1" go test ./internal/config -run 'TestOverlayGapOwnersMatchFirstProductionBlocker\|TestManagerConfigV2Vectors' -count=1` | 1 | Overlay vector EffectiveJSON differences were only `source_signers` and `require_source_signers`, while the test still expected a permissions difference. |
| `CURATOR_CONFORMANCE_ROOT="$PWD/.temp/pinned-spec-dcc7f015/conformance/v1" go test ./internal/envfragment -count=1` | 1 | The v2 token was accepted by the v1 reader as the invalid identity; valid v1 rows were rejected because the reader compared against the v2 emitter token. |

The overlay ledger keeps all 28 rows because the signer fields still block
the cases: 14 schema cases fail through production `Load`, and 14 vectors
differ in `EffectiveJSON`. Its owner histogram changed from
`STORY-260916-ioemse+STORY-260922-1cenbr: 28` to
`STORY-260916-ioemse: 28`. Permissions were removed from the expected
blocker set; no row now fully passes, so none was removed.

The v1 reader now compares against the fixed v1 protocol token independently
of the v2 emitter. Against the pinned root, the v1 case runner reports
`50 driven, 0 known-gap, 0 bound, 0 skipped, 50 total`; this includes a
passing rejection of `invalid-fragment-identity.json`.

Post-fix verification (each command was run directly as a standalone process):

| Command | Exit | Result |
| --- | ---: | --- |
| `CURATOR_CONFORMANCE_ROOT="$PWD/.temp/pinned-spec-dcc7f015/conformance/v1" go test ./internal/config -run 'TestOverlayGapOwnersMatchFirstProductionBlocker\|TestManagerConfigV2Vectors' -count=1` | 0 | Overlay blockers and pinned manager vectors pass. |
| `CURATOR_CONFORMANCE_ROOT="$PWD/.temp/pinned-spec-dcc7f015/conformance/v1" go test ./internal/envfragment -count=1` | 0 | Full envfragment package passes with v1 and v2 cases. |
| `CURATOR_CONFORMANCE_ROOT="$PWD/.temp/pinned-spec-dcc7f015/conformance/v1" go test -race ./internal/config -run 'TestOverlayGapOwnersMatchFirstProductionBlocker\|TestManagerConfigV2Vectors' -count=1` | 0 | Config race run passes. |
| `CURATOR_CONFORMANCE_ROOT="$PWD/.temp/pinned-spec-dcc7f015/conformance/v1" go test -race ./internal/envfragment -count=1` | 0 | Envfragment race run passes. |
| `CURATOR_CONFORMANCE_ROOT="$PWD/.temp/pinned-spec-dcc7f015/conformance/v1" go test -v ./internal/envfragment -run '^TestFragmentAuthoritativeSchemaCases$' -count=1` | 0 | Pinned v1 schema cases: 50/50 driven, including invalid fragment identity. |
| `bash .github/ci/ledger-consistency.sh .temp/TASK-260923-em42lw-revision6-ledger` | 0 | 417 platform-ledger rows checked across linux, darwin, and windows. |
| `bash .github/ci/gate-selftest.sh` | 0 | 212 passed, 0 failed; includes ledger-consistency and platform-case ledger checks. |
| `go vet ./internal/config ./internal/envfragment ./cmd/curator` | 0 | Passed. |
| `go build -o .temp/TASK-260923-em42lw-revision6-curator ./cmd/curator` | 0 | Build passed. |
| `make lint` | 0 | `golangci-lint run`: 0 issues. |
| `gofmt -l internal/config/environments_conformance_test.go internal/envfragment/fragment_schema_test.go && git diff --check` | 0 | No gofmt output or diff errors. |

No CHANGELOG or LOGBOOK edit was made per the current gate-fix instruction.
The previously attached production-entry permissions rows and four lattice
narrowing mutants remain documented above; this revision only repairs the
two hosted gate regressions and re-derives the overlay ledger ownership.

### Revision 6 — gate fix revalidation before handoff

I reran the bounded gate-fix commands against the pinned root at
`.temp/pinned-spec-dcc7f015/conformance/v1` (curator-spec revision
`dcc7f015e2d97edf2d52928afb6fd79ec8129e8b`). Each listed command was a direct,
standalone process from the Story worktree:

| Command | Exit | Result |
| --- | ---: | --- |
| `CURATOR_CONFORMANCE_ROOT="$PWD/.temp/pinned-spec-dcc7f015/conformance/v1" go test ./internal/config -run 'TestOverlayGapOwnersMatchFirstProductionBlocker\|TestManagerConfigV2Vectors' -count=1` | 0 | Overlay blockers and manager-config vectors pass. |
| `CURATOR_CONFORMANCE_ROOT="$PWD/.temp/pinned-spec-dcc7f015/conformance/v1" go test ./internal/envfragment -count=1` | 0 | Fragment package passes. |
| `CURATOR_CONFORMANCE_ROOT="$PWD/.temp/pinned-spec-dcc7f015/conformance/v1" go test -race ./internal/config -run 'TestOverlayGapOwnersMatchFirstProductionBlocker\|TestManagerConfigV2Vectors' -count=1` | 0 | Config race run passes. |
| `CURATOR_CONFORMANCE_ROOT="$PWD/.temp/pinned-spec-dcc7f015/conformance/v1" go test -race ./internal/envfragment -count=1` | 0 | Fragment race run passes. |
| `CURATOR_CONFORMANCE_ROOT="$PWD/.temp/pinned-spec-dcc7f015/conformance/v1" go test -v ./internal/envfragment -run '^TestFragmentAuthoritativeSchemaCases$' -count=1` | 0 | 50/50 v1 cases driven; `invalid-fragment-identity.json` is rejected by the v1 reader. |
| `bash .github/ci/ledger-consistency.sh .temp/TASK-260923-em42lw-revision6-ledger` | 0 | 417 ledger rows checked across linux, darwin, and windows. |
| `bash .github/ci/gate-selftest.sh` | 0 | 212 passed, 0 failed. |
| `go build -o .temp/TASK-260923-em42lw-gatefix-curator ./cmd/curator` | 0 | Curator CLI builds. |
| `gofmt -l internal/config/environments_conformance_test.go internal/envfragment/fragment_schema_test.go` | 0 | No files need formatting. |
| `git diff --check` | 0 | No whitespace errors. |

The first gate-selftest invocation yielded after its initial 30-second output
without the session handle being recorded, so its terminal exit is unknown and
is not counted. I reran the same command with a retained session handle; the
exit 0 result above is from that complete run.

The owner histogram below counts every row in `.github/ci/conformance-gaps.tsv`
before and after the overlay re-derivation (69 rows each):

- Before: `BUG-260923-2afgyq=5`, `STORY-260910-2qmrb8=3`,
  `STORY-260910-6bo7ej=2`, `STORY-260916-1i1gfo=2`,
  `STORY-260916-ioemse=11`,
  `STORY-260916-ioemse+STORY-260922-1cenbr=28`,
  `STORY-260922-1cenbr=16`, `STORY-260925-1v7pvn=2`.
- After: `BUG-260923-2afgyq=5`, `STORY-260910-2qmrb8=3`,
  `STORY-260910-6bo7ej=2`, `STORY-260916-1i1gfo=2`,
  `STORY-260916-ioemse=39`, `STORY-260922-1cenbr=16`,
  `STORY-260925-1v7pvn=2`.

Within the 28 overlay rows specifically, ownership changes from
`STORY-260916-ioemse+STORY-260922-1cenbr=28` to
`STORY-260916-ioemse=28`. None of those rows passes fully: the pinned schema
cases still fail on signer fields, and the vectors still differ only at
`source_signers` and `require_source_signers`, so the ledger ratchet correctly
retains all 28 rows. No CHANGELOG or LOGBOOK edit was made.

## Revision 8 — Windows fragment schema

The production-entry test still runs each of the four permissions rows to completion on Windows. Whole-fragment validation now lives in a nested `schema-validation` subtest; on Windows only that child is skipped, after the emitted JSON, permissions member, lattice values, and system-lock warning have been asserted. The exact skip reason is:

> launch-env-fragment-v2 schema gap (environments §10.2): Windows host-native paths do not match pattern ^/[^\u0000]*$; permissions assertions passed

The platform ledger requires the parent and each of the four permissions rows on linux, darwin, and windows. Its wildcard child row tolerates the schema-validation skip on Windows under `platform-control`; the skip-class table admits that exact reason prefix.

### v1 precedent and curator-spec gap

There is no existing v1 production-entry test that validates the emitted fragment against the v1 schema and skips that validation on Windows. The production-entry test `TestEnvResolveRepairEmitsFragment` at `cmd/curator/env_test.go:42` checks CLI JSON and is required on Windows by `.github/ci/platform-cases.tsv:468`, but it does not validate the schema. The v1 fixture test `TestFragmentAuthoritativeSchemaCases` at `internal/envfragment/fragment_schema_test.go:281` checks static POSIX-rooted fixtures, not host-native production paths.

The copied curator-spec main v2 schema's `$defs.absolutePath` at `internal/envfragment/testdata/curator-spec-main/schemas/v1/launch-env-fragment-v2.schema.json:100-105` uses pattern `^/[^\\u0000]*$`. The `env` values reference this definition at line 25. It cannot accept a native Windows path such as the reported `C:\Users\…` path. Environments §10.2 therefore has a spec gap for Windows fragment paths. A curator-spec follow-up leaf should define the Windows path representation and update the v2 schema and conformance cases; this code does not invent a Windows form or weaken the copied schema.

### Revision 8 verification

Commands below ran directly from the refreshed Story worktree. Each exit code is the process exit code.

| Command | Exit | Result |
| --- | ---: | --- |
| `go test -json -run 'TestEnvResolveEmitsPermissionsV2AtProductionEntry|TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases' ./cmd/curator > /tmp/TASK-260923-em42lw-darwin-permissions-rev8.json` | 0 | Darwin production matrix and all 14 copied v2 schema cases pass. |
| `CI_PLATFORM_CASES=/tmp/TASK-260923-em42lw-platform-cases-rev8.tsv CI_GATE_GOOS=darwin .github/ci/platform-case-gate.sh /tmp/TASK-260923-em42lw-darwin-permissions-rev8.json /tmp/TASK-260923-em42lw-platform-gate-rev8` | 0 | Parent and all four required production rows pass; 0 skips on Darwin. |
| `GOOS=windows go vet ./cmd/curator` | 0 | Windows package and test code vet cleanly. |
| `bash .github/ci/ledger-consistency.sh .temp/TASK-260923-em42lw-revision8-ledger` | 0 | 424 ledger rows compile consistently across linux, darwin, and windows. |
| `go build -o /tmp/TASK-260923-em42lw-curator ./cmd/curator` | 0 | Curator CLI builds on the local Darwin host. |
| `gofmt -l cmd/curator/env_permissions_test.go` | 0 | No formatting output. |
| `git diff --check` | 0 | No whitespace errors. |
| `git fetch origin main` | 0 | `origin/main`, `FETCH_HEAD`, and candidate HEAD match `316438cc3f830801c44adba75750a6b28a156933`. |
| `task-board worktree refresh-candidate TASK-260923-em42lw` | 0 | Managed candidate refreshed to reviewed trunk `316438cc3f830801c44adba75750a6b28a156933`, preserving both the intervening trunk changes and task delta. |

A native Windows runtime was not available locally. The required Windows vet and cross-platform ledger checks passed; hosted Windows execution remains part of the handoff validation. No CHANGELOG or LOGBOOK edit was made per the gate-fix instruction.

## Revision 9 — reconcile with 2elcdc

Fetched `origin/main`. `HEAD`, `origin/main`, and `FETCH_HEAD` all resolve to
`316438cc3f830801c44adba75750a6b28a156933`, so there was no newer trunk to
combine and `task-board worktree refresh-candidate` was not needed.

Reconciled the two affected config files against trunk's landed isolation
behavior:

- `internal/config/environments.go`: removed the stale comment claiming the
  system isolation lock only permits `shared`; the task's permissions rule
  still documents that the system lock permits only `native`.
- `internal/config/environments_test.go`: removed the stale `isolated
  direction` refusal row. Trunk's `TestSystemV2IsolationLockAcceptsIsolated`
  remains, covering the admitted system lock value.

No other `internal/config` hunk is unrelated to this task's permissions
parsing and lock support. No `CHANGELOG.md` or `LOGBOOK.md` edit was made, as
the current gate-fix instruction directs.

### Revision 9 verification

Each validation command below ran directly from the Story worktree:

| Command | Exit | Result |
| --- | ---: | --- |
| `go test ./internal/config -count=1` | 0 | Config tests pass. |
| `go test ./internal/config -race -count=1` | 0 | Config race tests pass. |
| `go test ./cmd/curator -run 'EnvResolve|Permissions|Isolation' -count=1` | 0 | Production resolver, permissions, and isolation tests pass (67.509s). |
| `git diff --check` | 0 | No whitespace errors. |

The first CLI test session exceeded its output window before its session id
was retained, so that attempt's exit is unknown and is not counted. I reran
the exact command with a retained session id; the recorded exit 0 above is
from that complete process.

## Revision 10 (carry-forward republish)

Trunk moved to `f02ba39e`; the accepted revision 9 delta is carried
uncommitted in the Story worktree. Verified against
`TASK-260923-em42lw_change-request_rev9.patch` (39 paths) by blob-hash
comparison (patch `index` pre/post images vs `HEAD:` blobs vs worktree
`git hash-object`):

- 35 of 39 paths: trunk untouched (`HEAD:` blob equals the patch
  pre-image) and worktree byte-identical to the revision 9 post-image.
  This covers all new files (`cmd/curator/env_permissions_test.go` and
  the 22 vendored `internal/envfragment/testdata/curator-spec-main`
  files) plus `cmd/curator/env.go`, `cmd/curator/env_test.go`,
  `docs/environment-config.md`, `go.mod`, `go.sum`,
  `internal/config/config.go`, `internal/config/environments.go`,
  `internal/config/environments_conformance_test.go`,
  `internal/config/environments_test.go`,
  `internal/envfragment/envfragment.go`,
  `internal/envfragment/envfragment_test.go`,
  `internal/envfragment/fragment_schema_test.go`,
  `.github/ci/conformance-gaps.tsv`.
- 4 intersecting paths where trunk moved (commit `7de564c4`,
  stage-c-composition-path-kind-and-import, in regions disjoint from the
  task hunks): `.github/ci/platform-cases.tsv`,
  `.github/ci/skip-classes.tsv`, `internal/envprofile/managed.go`,
  `internal/envregistry/envregistry.go`. Every revision 9 `+` line is
  present in the worktree (0 missing across 29 added lines); trunk
  markers retained (`stateread` import and `ReadFile`/`Lstat` call sites,
  `DiagBackupRecordUnreadable`, trunk TSV rows — HEAD is a line
  subsequence of the worktree for the TSVs); each task addition occurs
  exactly once; no `<<<<<<<`/`>>>>>>>`/`|||||||` marker lines anywhere.
  Nothing dropped, nothing duplicated.
- `CHANGELOG.md`: revision 9 carries no CHANGELOG hunk and the worktree
  file equals trunk (`git diff HEAD -- CHANGELOG.md` empty), so there is
  nothing to revert. No stray root `TASK-*`/`BUG-*`, `test/`, or
  `ledger/` paths.
- Worktree freshness: `git diff --name-only HEAD -- . ':!.task-board'`
  lists only the 17 tracked revision 9 paths; the two untracked entries
  (`cmd/curator/env_permissions_test.go`,
  `internal/envfragment/testdata/`) are revision 9 new files. No trunk
  revert, no stale snapshot.
- No source changes were made in this revision; content is revision 9
  as accepted.

### Revision 10 verification (this turn, unpiped, real exits)

| Command | Exit | Result |
| --- | ---: | --- |
| `go test ./internal/config ./internal/envfragment -count=1` | 0 | Both packages pass. |
| `go test ./cmd/curator -run 'EnvResolve|Permissions' -count=1` | 0 | Resolver and permissions tests pass (76.918s). |

### CHANGELOG entry (for release prep)

Emit launch-env-fragment-v2 with effective per-profile permissions and system-lock provenance.

## Revision 11 (carry-forward republish)

Trunk moved to `bd3c0f43`; the accepted revision 10 delta is carried
uncommitted in the Story worktree. Verified against
`TASK-260923-em42lw_change-request_rev10.patch` (39 paths) by blob-hash
comparison (patch `index` pre/post images vs `HEAD:` blobs vs worktree
`git hash-object`):

- 38 of 39 paths: trunk untouched (`HEAD:` blob equals the patch
  pre-image) and worktree byte-identical to the revision 10 post-image.
  This covers all new files (`cmd/curator/env_permissions_test.go` and
  the 21 vendored `internal/envfragment/testdata/curator-spec-main`
  files) plus `cmd/curator/env.go`, `cmd/curator/env_test.go`,
  `docs/environment-config.md`, `go.mod`, `go.sum`,
  `internal/config/config.go`, `internal/config/environments.go`,
  `internal/config/environments_conformance_test.go`,
  `internal/config/environments_test.go`,
  `internal/envfragment/envfragment.go`,
  `internal/envfragment/envfragment_test.go`,
  `internal/envfragment/fragment_schema_test.go`,
  `.github/ci/platform-cases.tsv`, `.github/ci/skip-classes.tsv`,
  `internal/envprofile/managed.go`, `internal/envregistry/envregistry.go`.
- 1 intersecting path where trunk moved (range `f02ba39e..bd3c0f43`):
  `.github/ci/conformance-gaps.tsv`. Trunk added one comment line
  (`# Case IDs and owners are verified against v1.0.0-rc.13 tag commit
  23435129ebc4c29e5b7f75ec72a0aa0cd3f16065.`); the worktree keeps it and
  applies all 28 revision 10 added lines, each exactly once, with no
  `<<<<<<<`/`>>>>>>>`/`|||||||` markers. Nothing dropped, nothing
  duplicated.
- `CHANGELOG.md`: revision 10 carries no CHANGELOG hunk and the worktree
  file equals trunk (`git diff HEAD -- CHANGELOG.md` empty), so there is
  nothing to revert; the entry text stays verbatim in the
  "CHANGELOG entry (for release prep)" section above. No stray root
  `TASK-*`/`BUG-*`, `test/`, or `ledger/` paths.
- Worktree freshness: `git diff --name-only HEAD -- . ':!.task-board'`
  lists only the 17 tracked revision 10 paths; the two untracked entries
  (`cmd/curator/env_permissions_test.go`,
  `internal/envfragment/testdata/`) are revision 10 new files. No trunk
  revert, no stale snapshot.
- No source changes were made in this revision; content is revision 10
  as accepted.

### Revision 11 verification (this turn, unpiped, real exits)

| Command | Exit | Result |
| --- | ---: | --- |
| `go test ./internal/config ./internal/envfragment -count=1` | 0 | Both packages pass. |
| `go test ./cmd/curator -run 'EnvResolve|Permissions' -count=1` | 0 | Resolver and permissions tests pass (66.085s). |
