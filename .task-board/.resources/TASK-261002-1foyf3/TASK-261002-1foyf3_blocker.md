# TASK-261002-1foyf3 — rc14-pin-and-v2-writer-cutover: stop-line evidence

Status: implementation stopped; draft is uncommitted, not a review handoff.

## Verified input and draft

Fresh main advertisement and fetched integration base both equal 68210eccfd656e770cf9101c2b927ceda01359df; the Story worktree HEAD equals that base. No branch commit, switch, rebase, or merge was performed.

Published curator-spec v1.0.0-rc.14 annotated tag object: c13ab1bd4e6f7751e873a905ee545366e45fcf77. Peeled commit: daf15ec8e78c29148063f14905fafc0b8b682786. Checked-out conformance/v1/manifest.json SHA-256: 6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5. These were resolved from the remote tag and verified locally.

Draft sets SPEC_PIN to the peeled commit and EnableV2Writers=true; updates the pin documentation, ledger comment and one Unreleased changelog line. It removes only the rc.14 snapshot-acquisition/cases/byte-exact-snapshot gap. The exact count remains 1: one production extraction case, exercised with autocrlf=true and false. Nine rc.14 seed/posture B gaps remain. CodexSeedRevision=A and SecurityPostureRevision=A are untouched; internal/buildrepo/release_pin.go remains byte-unchanged at rc.8. The vendored rc.13 skillfile-sources corpus identity is deliberately unchanged; it does not name the default CI root.

Legacy marker-writer tests now explicitly select v1 and restore the switch. V2 marker, registry-backed install, state-store, context resolution, and environment publication tests exercise the new default. Existing NUL, version-mismatch, and non-aliasing negatives are retained. Not all these draft checks have passed; see validation below.

## Reproduced constraint

Added TestRC14MigrationRehashesLegacyIdentities in internal/envprofile/rc14_cutover_test.go. It provisions a genuine rc.13/v1 Pi managed home using production Resolve, turns on the rc.14 writer, plans a credential unlink, and invokes real ApplyMigration with the required plan hash. It reads the published environment marker and independently hashes the actual root-context files with v2 framing.

Observed failures:

- ApplyMigration publishes schema 3/hash_version 2 while retaining the context member's exact v1 state identity.
- Published root-context surface identity is sha256:15edf46dfb26fb9703e14a3ab1bbabde81d11bc4bc4aa88f411fa1e66c80a6fc, while the actual v2 identity is sha256:fb7dbbe03544b782b8687888d7132a17e8cac47dc1e47e5f3b300d49f6e202d9.

Production changedMarkers (internal/envprofile/migrate.go:1273) copies the old marker and changes only its carrier/hash version and credential records. The existing v2 migration regression starts from v2-provisioned content and relabels its marker as schema 1, so it cannot establish a genuine v1-to-v2 identity conversion. The new test deliberately avoids that assumption.

The normative rc.14 protocol/environments.md section 8.2 says hash_version versions every snapshot-state and materialized-surface hash in the marker, while schemas 1 and 2 retain v1 framing. Section 7.4 requires otherwise-needed marker publication under the migration lock/journal. Relabelling unchanged v1 values as v2 violates that model.

Rehashing only the marker does not resolve it: managed verification (internal/envprofile/managed.go:1781) requires each marker member's pin to equal its lock member's pin. State store entries are keyed by that pin. Converting member pins therefore touches the lock and store identities; changing the lock also changes generated surface headers that bind its CCJ-1 hash. The credential-migration implementation currently owns link changes and marker replacement, not a coordinated profile/store/lock/surface migration. No compensating writer exceptions, fallback hashes, or weakened identity comparisons were added.

## Options and exact decision needed

1. Recommended: specify and implement an explicit atomic profile hash migration before the global cutover. Define which operation owns state-store rekeying, context-lock-v2 publication, regenerated surfaces, and marker-v3 publication; define how an existing v1 home reaches that operation before credential migration. Then rerun the real legacy transition regression and hosted matrix.
2. Expand credential migration to own that entire identity transition in the same journal. This requires an explicit scope/ownership decision and rollback/plan-drift design across lock, store, generated surfaces and marker, rather than only credential links.
3. Preserve v1 markers on credential changes. This retains their identities, but conflicts with the present rc.14 current-writer requirement; it needs a specification decision and is not implemented here.

Required architecture input: choose the operation and ownership boundary for a genuine v1 profile's coordinated v2 identity transition, including credential-triggered marker publication. A flag-only cutover cannot satisfy the current AC safely.

## Validation

All Go commands ran directly as standalone subprocesses with GOFLAGS=-work; the WORK directory was preserved. No mini-build-lock existed when checks began. syspolicyd was running with successive crashes=374 before and after every recorded command. No crash-count increase was observed.

Validation results will be listed below with real exit codes and durations. Hosted gate, full environment/context suites, and lint are not claimed green; they were not launched after the confirmed stop-line. LOGBOOK.md was not edited, as instructed. The board outcome and notes carry this finding.

### Commands executed in this run

- **first**: `GOFLAGS=-work go test ./internal/hashing ./internal/marker ./internal/contextlock/... ./internal/envmarker/... ./internal/conformancecoverage ./internal/interop/... -count=1 -timeout=8m`; exit **1**, 10.88 seconds; syspolicyd running, successive crashes **374 -> 374**.
- **requested**: `GOFLAGS=-work go test ./internal/hashing ./internal/marker ./internal/install/... ./internal/contextlock/... ./internal/envmarker/... ./internal/conformancecoverage ./internal/interop/... -count=1 -timeout=8m`; exit **1**, 487.44 seconds; syspolicyd running, successive crashes **374 -> 374**.
- **migration-regression**: `GOFLAGS=-work go test ./internal/envprofile -run ^TestRC14MigrationRehashesLegacyIdentities$ -count=1 -timeout=2m`; exit **1**, 29.34 seconds; syspolicyd running, successive crashes **374 -> 374**.
- **build**: `GOFLAGS=-work go build ./cmd/curator`; exit **0**, 7.71 seconds; syspolicyd running, successive crashes **374 -> 374**.
- **bounded-entrypoints**: `GOFLAGS=-work go test ./internal/install ./internal/contextresolve ./internal/contextstore ./internal/envprofile ./internal/interop/environments -run ^(TestEndToEndInstall|TestV2RegistryAttestationUsesVersionedInstallPath|TestV1InstallRejectsV2RegistryEvidence|TestStateEntryIsContentKeyed|TestContentHashUsesVersion2Framing|TestMinimalResolution|TestResolvePublishesCredentialMarkerThroughLockedJournalV2WriterMode|TestConformanceSnapshotAcquisition)$ -count=1 -timeout=3m -v`; exit **0**, 10.74 seconds; syspolicyd running, successive crashes **374 -> 374**.

The first run (exit 1) preceded the explicit legacy marker test adaptations; marker tests failed their v1 writer expectations. It is diagnostic evidence, not a green result on the current draft. The requested suite ran after those adaptations: hashing, marker (including v1 reader/writer negatives), install/atomicity, contextlock, envmarker, conformancecoverage and both interop packages passed, but internal/install hit the deliberately bounded 8-minute package timeout. The suite exit is 1, not passing; no unexecuted install test is claimed as passing. No existing attached evidence was accepted as a substitute for a command run here.

The added migration regression exited 1 because the asserted v1-to-v2 identity transition is unimplemented. This is a failing regression proving the blocker, not a passing expected-red gate.

The subsequent bounded entry-point subset exited 0. It executes registry-backed v2 install, the v1-install/v2-registry rejection, ordinary Project install, context Resolve, EnsureState, v2 ContentHash, locked environment-marker publication via Resolve, and exact snapshot acquisition. It does not replace the full install/environment/context matrix. Snapshot tally is **1 driven / 1 published**, 0 known-gap, 0 bound, 0 skipped; both autocrlf settings executed.

CLI `go build ./cmd/curator` exited 0. Generated binary is retained privately under the ignored local evidence directory and is not attached or part of the repository delta. `git diff --check` exited 0. Lint, go vet, remaining full environment/context package suites, and hosted CI were not run because the confirmed ownership conflict stops the cutover. No full hosted-green claim is made. No task acceptance checklist item was checked except outcome attachment.

### Captured regression output (WORK path omitted)

```
--- FAIL: TestRC14MigrationRehashesLegacyIdentities (10.36s)
    rc14_cutover_test.go:50: migration relabelled acme's v1 state identity as v2 without rehashing
    rc14_cutover_test.go:68: migration published surface hash sha256:15edf46dfb26fb9703e14a3ab1bbabde81d11bc4bc4aa88f411fa1e66c80a6fc as v2, want sha256:fb7dbbe03544b782b8687888d7132a17e8cac47dc1e47e5f3b300d49f6e202d9
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	10.880s
FAIL
```

### Captured bounded entry-point output (WORK path omitted)

```
=== RUN   TestEndToEndInstall
=== PAUSE TestEndToEndInstall
=== RUN   TestV1InstallRejectsV2RegistryEvidence
--- PASS: TestV1InstallRejectsV2RegistryEvidence (0.63s)
=== RUN   TestV2RegistryAttestationUsesVersionedInstallPath
--- PASS: TestV2RegistryAttestationUsesVersionedInstallPath (4.48s)
=== CONT  TestEndToEndInstall
--- PASS: TestEndToEndInstall (3.46s)
PASS
ok  	github.com/relux-works/curator/internal/install	9.429s
=== RUN   TestMinimalResolution
--- PASS: TestMinimalResolution (0.00s)
PASS
ok  	github.com/relux-works/curator/internal/contextresolve	2.379s
=== RUN   TestStateEntryIsContentKeyed
--- PASS: TestStateEntryIsContentKeyed (0.06s)
=== RUN   TestContentHashUsesVersion2Framing
--- PASS: TestContentHashUsesVersion2Framing (0.00s)
PASS
ok  	github.com/relux-works/curator/internal/contextstore	0.428s
=== RUN   TestResolvePublishesCredentialMarkerThroughLockedJournalV2WriterMode
--- PASS: TestResolvePublishesCredentialMarkerThroughLockedJournalV2WriterMode (2.22s)
PASS
ok  	github.com/relux-works/curator/internal/envprofile	4.117s
=== RUN   TestConformanceSnapshotAcquisition
=== RUN   TestConformanceSnapshotAcquisition/byte-exact-snapshot
=== RUN   TestConformanceSnapshotAcquisition/byte-exact-snapshot/autocrlf=true
=== RUN   TestConformanceSnapshotAcquisition/byte-exact-snapshot/autocrlf=false
=== NAME  TestConformanceSnapshotAcquisition
    coverage.go:236: published cases snapshot-acquisition/cases: 1 driven, 0 known-gap, 0 bound, 0 skipped, 1 total
--- PASS: TestConformanceSnapshotAcquisition (0.83s)
    --- PASS: TestConformanceSnapshotAcquisition/byte-exact-snapshot (0.83s)
        --- PASS: TestConformanceSnapshotAcquisition/byte-exact-snapshot/autocrlf=true (0.14s)
        --- PASS: TestConformanceSnapshotAcquisition/byte-exact-snapshot/autocrlf=false (0.12s)
PASS
ok  	github.com/relux-works/curator/internal/interop/environments	2.063s
```

### Requested-suite terminal facts

```
ok  	github.com/relux-works/curator/internal/hashing	0.768s
ok  	github.com/relux-works/curator/internal/marker	2.433s
panic: test timed out after 8m0s
FAIL	github.com/relux-works/curator/internal/install	480.499s
ok  	github.com/relux-works/curator/internal/install/atomicity	339.810s
ok  	github.com/relux-works/curator/internal/contextlock	1.803s
ok  	github.com/relux-works/curator/internal/envmarker	2.806s
ok  	github.com/relux-works/curator/internal/conformancecoverage	3.328s
ok  	github.com/relux-works/curator/internal/interop	2.128s
ok  	github.com/relux-works/curator/internal/interop/environments	8.529s
FAIL
```
