# TASK-260923-2elcdc results

## Implementation

- system-config-v2 now accepts isolated in the locked environments.isolation map while retaining whole-map manager §1 lock semantics.
- Environment resolution carries the effective isolation map, pre-overlay user values, and system policy path. Silence follows the locked direction; explicit shared against a locked isolated returns environment_isolation_lock_conflict naming the profile, environment, and policy source.
- A provisioned shared passthrough continues to fail closed with environment_credential_conflict. Refused repair leaves the link and native bytes unchanged and names the exact profile/environment F-C2 plan/apply commands.
- env migrate now receives the same effective machine isolation, allowing that explicit plan/apply command to perform the isolated transition.
- Added config, pinned-vector projection, and CLI production-entry tests.

## Validation

Every command below was run directly. Exit codes are the process exit codes.

- env CURATOR_CONFORMANCE_ROOT=/tmp/curator-spec-task-260923-2elcdc/conformance/v1 go test -count=1 ./internal/config — exit 0.
- env CURATOR_CONFORMANCE_ROOT=/tmp/curator-spec-task-260923-2elcdc/conformance/v1 go test -count=1 -v ./internal/config -run '^(TestSystemConfigV2SchemaCases|TestSystemConfigV2IsolationDirectionsFromPinnedCases)$' — exit 0. Pin: dcc7f015e2d97edf2d52928afb6fd79ec8129e8b; manifest SHA-256 cb7a98e97543282cdde75f04c55fb15acb91062f8ba6510e3d603db42f1c1ab2. The full 42-case system-config-v2 family reports 36 driven, 6 known-gap, 0 bound, 0 skipped. Both pinned shared/isolated direction cases are also projected to just the isolation knob and locked entry and pass through Load. The full direction cases remain known gaps because they bundle source_signers and permissions locks outside this task's supported subset; those existing gaps are not presented as passes.
- go test -count=1 ./cmd/curator -run '^(TestEnvResolveIsolatedSystemLockUsesDirection|TestEnvResolveExplicitSharedConflictsWithIsolatedSystemLock|TestEnvResolveLockedIsolationRequiresMigration)$' — exit 0.
- go test -count=1 ./internal/envregistry — exit 0.
- go test -count=1 ./internal/envprofile -run '^TestSharedToIsolatedRemovesStaleLink$' — exit 0.
- golangci-lint run — exit 0, 0 issues.
- go vet ./internal/config ./internal/envregistry ./internal/envprofile ./cmd/curator — exit 0.
- go build -o /tmp/curator-task-260923-2elcdc ./cmd/curator — exit 0.
- gofmt -l cmd internal — exit 0, no files listed.
- git diff --check — exit 0.
- The full repository landing suite was not run manually; the campaign handoff runtime owns that configured suite and runs it once.

## Narrowing mutation evidence

Each mutant was temporary and restored before final green validation. Baseline selectors exclude the new regression case and pass (mutant survives); the new CLI regression case fails under the mutant (mutant killed). The non-zero kill runs are expected failures.

| Mutant | Baseline survival command (exit) | New regression command (exit) |
| --- | --- | --- |
| Under an isolated lock, silently remove recorded passthrough links during repair | go test -count=1 ./internal/envprofile -run '^TestSharedToIsolatedRemovesStaleLink$' — 0 | go test -count=1 ./cmd/curator -run '^TestEnvResolveLockedIsolationRequiresMigration$' — 1 |
| Narrow explicit-shared refusal so it skips profile acme | go test -count=1 ./internal/envregistry — 0 | go test -count=1 ./cmd/curator -run '^TestEnvResolveExplicitSharedConflictsWithIsolatedSystemLock$' — 1 |
| Narrow locked isolation so silent values resolve to shared | go test -count=1 ./internal/envregistry — 0 | go test -count=1 ./cmd/curator -run '^TestEnvResolveIsolatedSystemLockUsesDirection$' — 1 |

## Scope and findings

- Changed files: cmd/curator/env.go, cmd/curator/env_test.go, cmd/curator/envmigrate.go, internal/config/config.go, internal/config/environments.go, internal/config/environments_conformance_test.go, internal/config/environments_test.go, and internal/envregistry/envregistry.go.
- No CHANGELOG.md or LOGBOOK.md edit. Findings and the release-prep entry are recorded here.
- The six pinned known gaps are pre-existing and outside this task; no support for unrelated system locks was added.

## CHANGELOG entry (for release prep)

System environments.isolation locks can enforce either shared or isolated. Silent locked settings are honored, conflicting explicit shared requests fail closed, and existing shared passthrough requires the explicit F-C2 migration.
