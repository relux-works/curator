# TASK-261014-2f0qnv — preserve-lock implementation handoff

Candidate: ready for review. Latest direct hosted gate exited 0; candidate remains uncommitted. Earlier green and failed handoff results remain recorded below.

## Changes and identities

- Base: `84cf54b94670ae9881f91b1c147feb4b75afddee`, verified equal to refreshed `origin/main` when work began.
- Spec: `6648e18213d519cb4b6a33a49b0853d5b6eaa37d`.
- Hosted candidate: `6706a888fdcc4777fa92b8a809cab0554940f2ee`; every changed/new file is byte-equal to this snapshot. Story HEAD remains at the base.
- Added `--preserve-lock`, its CLI/API usage guards, usage/help expectations and README documentation.
- `Resolve` captures the original lock hash and refuses lock-changing repairs before store rebuild checks, repair-lock acquisition, staging or publication. The shared `repairWouldChangeLockHash` predicate currently identifies v1 locks under a v2 writer. The diagnostic names the selected environment/profile and the exact plain-repair command.
- Repair and store rebuild use `beginRepairOperation` and `beginOperationValidated`: the original hash is checked after acquisition before journal recovery, then checked again before repair. Ordinary repair retains migration and current-home behavior.
- Plain identity migration now applies the same owner-only regular-file protection as ordinary profile-record publication. The Windows full gate exposed a newly published lock with broad inherited mutation rights, which made the subsequent production read reject the manager’s own output. This is fixed in `identityMigration.publish`, without relaxing the boundary.
- Added the exact ten-vector family from the spec outside its released `conformance/v1` corpus to `internal/envprofile/testdata/environments-repair-preserve-lock.json`. The production-entry driver runs it in ordinary hosted tests without moving the released conformance pin; `CURATOR_PRESERVE_LOCK_VECTORS` selects the upstream file directly.

## Coverage map

**10 of 10 AC/vector rows driven**, through the production call sites below. No surface table was supplied in the brief: catalog-row coverage is an explicit brief gap, not an implied claim over the attack-surface catalog.

| Vector / AC row | Named tests | Production call site |
| --- | --- | --- |
| v1-v2-stale-preserve-lock-refused | TestRepairPreserveLockVectors/v1-v2-stale-preserve-lock-refused | Resolve |
| v1-v2-unprovisioned-preserve-lock-refused | TestRepairPreserveLockVectors/v1-v2-unprovisioned-preserve-lock-refused | Resolve |
| v1-v2-current-preserve-lock-refused | TestRepairPreserveLockVectors/v1-v2-current-preserve-lock-refused; TestEnvResolvePreserveLockRefusesCurrentV1AndPlainRepairMigrates | Resolve / cli.cmdEnvResolve |
| v2-stale-preserve-lock-repairs | TestRepairPreserveLockVectors/v2-stale-preserve-lock-repairs | Resolve |
| v2-unprovisioned-preserve-lock-repairs | TestRepairPreserveLockVectors/v2-unprovisioned-preserve-lock-repairs | Resolve |
| preserve-lock-without-repair-usage | TestRepairPreserveLockVectors/preserve-lock-without-repair-usage; TestEnvResolvePreserveLockRequiresRepair | Resolve / cli.cmdEnvResolve |
| v2-current-preserve-lock-lock-free | TestRepairPreserveLockVectors/v2-current-preserve-lock-lock-free | Resolve |
| v1-v2-plain-repair-migrates-identity | TestRepairPreserveLockVectors/v1-v2-plain-repair-migrates-identity; TestEnvResolvePreserveLockRefusesCurrentV1AndPlainRepairMigrates | Resolve / cli.cmdEnvResolve |
| v2-lock-changed-before-repair-lock-refused | TestRepairPreserveLockVectors/v2-lock-changed-before-repair-lock-refused; TestPreserveLockRevalidationHoldsRepairLock | Resolve |
| v1-current-status-read-only | TestRepairPreserveLockVectors/v1-current-status-read-only; TestEnvResolvePreserveLockRefusesCurrentV1AndPlainRepairMigrates | StatusOf / cli.cmdEnvStatus |

Additional ordering tests: `TestPreserveLockRefusesBeforeStoreRebuild/{pin,boundary}` drives `Resolve` with corrupted v1 entries; `TestPreserveLockMigrationRefusalPrecedesBusyLock/{stale,unprovisioned,current}` holds a real manager lock; `TestPreserveLockRaceRefusesBeforeJournalRecovery` places a malformed journal behind the changed-lock guard. `TestPreserveLockRevalidationHoldsRepairLock` drives the raced-lock read through `Resolve` and verifies that a real competing acquisition is blocked during that read, so the injection-hook invocation is not the only acquisition evidence. All refusal tests refuse without fragments and compare the entire temp manager-home snapshot. Snapshots bracket the external holder: before acquisition and after release, because Windows denies reads of its locked byte range. No entries are omitted. Fixture lock entries are created before the baseline, so preparation by the external holder is not attributed to Resolve. The v2 CLI positive is `TestEnvResolvePreserveLockRepairBindsUnchangedLock`. All homes and native inputs are temporary.

Plain migration compares closure metadata, versions, source declarations, commit pins and snapshot bytes, with divergent candidate source bytes available, and verifies that the fragment binds the migrated hash. Fragment assertions parse `profile.lock_sha256` and require its exact computed value, rather than accepting the digest elsewhere in the document. `TestPlainRepairMigrationKeepsLockPrivate` drives plain repair through `Resolve` and validates the migrated lock with the platform-native private-file backend. M10 attacks the Unix permission-bit shape; Windows DACL preservation is separately driven by this test and the CLI migration test in the hosted matrix. Current v1 status provisions every required home so the CLI's `--check --json` succeeds without a lock rewrite.

## Narrowing mutants

Every mutant retains its gate and runs behavioral tests, not a source-token checker. Expected-red commands actually exited 1; these are failing gates, not passing suites. No survivors.

| Mutant | What the gate is narrowed to | Named failing test | Survivor bound |
| --- | --- | --- | --- |
| M1: pi migration exemption | Refuses v1-to-v2 migration only for environments other than pi | TestRepairPreserveLockVectors/v1-v2-current-preserve-lock-refused (also stale/unprovisioned) | None; killed |
| M2: refusal after acquisition | Allows repair-lock acquisition before refusing migration | TestPreserveLockMigrationRefusalPrecedesBusyLock/current (also stale/unprovisioned) | None; killed |
| M3: same-framing race exemption | Allows a changed v2 hash while retaining rejection for other writer generations | TestRepairPreserveLockVectors/v2-lock-changed-before-repair-lock-refused; TestPreserveLockRaceRefusesBeforeJournalRecovery | None; killed |
| M4: CLI named-profile usage exemption | Lets invalid preserve-only argv for one named profile reach the API rather than returning the required CLI usage code | TestEnvResolvePreserveLockRequiresRepair | None; killed |
| M5: current preserve repair locks | Keeps the current-home fast path only for non-preserve invocations | TestRepairPreserveLockVectors/v2-current-preserve-lock-lock-free | None; killed |
| M6: plain-current migration exemption | Lets a current plain-repair v1 home emit without the required migration | TestRepairPreserveLockVectors/v1-v2-plain-repair-migrates-identity | None; killed |
| M7: API profile usage exemption | Lets preserve-only requests for one named profile pass the API usage guard and reach resolution | TestRepairPreserveLockVectors/preserve-lock-without-repair-usage | None; killed |
| M8: fragment pin misbinding, digest retained elsewhere | Emits a wrong pi preserve-lock digest while retaining the expected digest in another fragment field | TestRepairPreserveLockVectors/v2-stale-preserve-lock-repairs | None; killed |
| M9: pi revalidation before acquisition | Keeps the hash check but permits pi to revalidate before holding the repair lock | TestPreserveLockRevalidationHoldsRepairLock | None; killed |
| M10: reopened lock mutation permissions | Retains protection on every regular target, then reopens Unix mutation bits for lock.json only | TestPlainRepairMigrationKeepsLockPrivate | None; killed |

## Validation

Local platform: Go 1.25.5, darwin/arm64. Commands were standalone processes with real exit codes, with logs under `.temp/TASK-261014-2f0qnv/`.

| Command / evidence | Exit | Result |
| --- | ---: | --- |
| `go build ./...` (`build-07.log`) | 0 | Production builds |
| `go vet ./...` (`vet-07.log`) | 0 | Candidate validates |
| `go test ./internal/envprofile -run 'TestRepairPreserveLock\|TestPreserveLock\|TestPlainRepairMigrationKeepsLockPrivate' -count=1 -v` (`env-focused-20.log`) | 0 | Ten vectors and ordering/race tests green |
| `go test ./cmd/curator -run 'TestEnvResolvePreserveLock\|TestGlobalInitEnvResolveSubcommandHelp' -count=1` (`cli-focused-12.log`) | 0 | CLI guards, hash binding, current-v1 status, migration and retained help test green |
| Original `TestEnvStatusRegistryBoundaryPostureAndCheck` with PATH restricted to active Go and system tools (`registry-controlled-01.log`) | 0 | Original name and assertions retained; green |
| `git diff --check` (`diff-check-07.log`) | 0 | Clean diff |
| `gofmt -l cmd internal` (`gofmt-07.log`) | 0 | Empty output |
| Ten mutant `go test` commands (`mutants-all.json`) | 1 each | Expected-red, with named failures above |
| `sh scripts/remote-gate.sh` (`remote-gate-07.log`) | 0 | Exact candidate full hosted matrix green |
| Prior Windows test gate (run 38536470961) | 1 | Repository failures in locked-entry snapshots and migration DACL publication; corrected and current Windows gate green |
| Snapshot-fixture ordering attempts (`env-focused-17.log`, `env-focused-18.log`) | 1 each | External-holder preparation was after baseline; corrected and env-focused-20 is green |
| Broad `go test ./internal/envprofile/... ./cmd/curator/...` (`focused-03.log`) | 1 | Envprofile package passed; broad CLI process stopped to bound local verification. Overall non-green, full local CLI suite unrun |
| Queued/diagnostic focused attempts (`focused-04.log`, `cli-diagnostic-01.log`) | 1 each | Host-toolchain-lock wait/timeout before CLI tests; unrun |
| Earlier scoped checks (`focused-01.log`, `cli-focused-05.log`, `cli-focused-09.log`) | 1 each | Corrected test compilation/assertion/fixture errors; not green evidence |
| Registry test with inherited PATH (`cli-focused-08.log`) | 1 | Existing optional engines provider was untrusted; controlled-path rerun above is green |

The successful envprofile package result was collected before later status-fixture expansion. It is supplementary evidence; the scoped vector suite was rerun on the expanded fixture. The hosted full matrix is the arbiter under R223; no local full-repository test or race suite was substituted.

## Hosted gate

`sh scripts/remote-gate.sh` (`remote-gate-07.log`) exited **0**. Run https://github.com/relux-works/curator/actions/runs/38544107344 accepted candidate `6706a888fdcc4777fa92b8a809cab0554940f2ee`: **20 successful jobs**, including full tests on Linux, macOS and Windows and race tests on Linux/macOS. Two workflow-conditioned lanes were skipped as configured. No hosted result was synthesized and no local mirror was substituted.

The uploaded Ubuntu and Windows gate streams explicitly contain pass events for **all ten vectors and ten additional/new or retained named tests**, including real held-lock revalidation, private migrated-lock validation, the CLI plain migration and the retained help/status tests. The Windows fixes are thus driven through the production entry points on Windows, not inferred from Unix results.

Earlier runs 38526983046 and 38527407940 had the unused-stderr repository lint failure; 38528453886 had the deprecated-GOROOT fixture lint failure. Each was cancelled after the concrete failure and its gate process exited 137; each temporary gate branch was removed. Run 38529129075 was superseded to strengthen fragment-field assertions before its full matrix had a verdict; its gate process exited 137 and its temporary branch was removed. These were non-green or unverified runs, not external-CI fallbacks. Run 38532695133 was superseded after strengthening the held-lock assertion and cancelled while two jobs were still running. Its gate process exited 137 and its temporary branch was removed. It is supplementary evidence for the prior test snapshot, not proof for the strengthened candidate. Its uploaded Ubuntu evidence confirms ten vector cases and the new/retained named CLI and ordering tests passed. Run 38536470961 had a concrete Windows repository failure: snapshots read locked entries during the external hold, and the CLI plain migration produced a lock with broad inherited DACL mutation rights. The Windows test gate was red; the remaining macOS race job was cancelled to supersede the candidate, the gate process exited 137, and the temporary branch was removed. Its current Ubuntu artifact had all ten vectors and nine additional named tests green, but this is supplementary evidence for the previous source/test identity. Local snapshot-ordering attempts env-focused-17/18 exited 1 because the external holder’s new lock entry was included only after the baseline; the fixture now creates it before the baseline. The accidental earlier-test edit was reverted before the current candidate.

## Gaps, bounds and scope

- No spec/vector disagreement found in the assigned sections.
- No AC/vector row is treated as out of contract. Other environment capabilities catalogued in spec §13 are outside this leaf: its AC is specifically `env resolve --repair --preserve-lock` and this ten-case family.
- The hermetic profiles use real temp-store snapshots rather than the spec's literal golden profile bytes. Tests assert exact before/after lock-byte preservation and actual computed fragment-hash binding; they do not claim the golden fixtures' literal digest values.
- Snapshot assertions include directory/file modes, regular-file bytes, symlink destinations, stores, staging and journals. They do not claim timestamp or platform ACL equality.
- Race tests inject a valid same-framing replacement before acquisition. Arbitrary uncooperative writers after acquisition are outside the specified pre-check/acquisition race.
- No source-inspecting gate was introduced. Mutants retain diagnostic tokens and execute the production behavioral suite.
- This is initial implementation, not review rework. Changes outside `internal/envprofile` are necessary CLI flag wiring, CLI tests/help and README documentation. Existing reviewer tests retain their original names and assertions; the help expectation adds the new flag.

## Logbook entry

The original operation helper runs journal recovery immediately after acquiring the mutation lock. Checking only inside `repairUnderLock` would permit recovery writes first. The candidate adds the validation callback before recovery and uses it for both repair and store rebuild. A malformed-journal race test and a real busy-lock test make that ordering observable. A later coverage review identified that the injection hook records an attempt, so a real competing acquisition was added at the revalidation read; its narrowing mutant catches validation performed before acquisition. Hosted lint caught an unused assignment and deprecated fixture API; both were corrected. Windows verification additionally exposed locked-byte-range snapshot reads and the identity publisher’s missing regular-file privacy step. Complete snapshots now bracket the external hold, and publication reuses the ordinary publisher’s private-file protection. The named private-lock regression and M10 preserve that behavior. Inherited local PATH can add an unrelated untrusted provider to status; an unchanged test passed under a controlled toolchain/system PATH. All mutations are restored, and the candidate is left uncommitted for the managed Story snapshot.


## Successor recovery and handoff validation

No production, test or configuration bytes changed during recovery. SHA-256 verification of all ten candidate paths matched the attached identity, and the previously green snapshot `6706a888fdcc4777fa92b8a809cab0554940f2ee` and handoff snapshot `9ba1c76ada3ed140f595c3ae0c8b3b45ad2322cc` share tree `265d93f4093227dfa6e1a05ea49c72dbb5a3551e`.

The first automatic Change Request validation ran `sh scripts/remote-gate.sh` and exited **1**: hosted run https://github.com/relux-works/curator/actions/runs/38551230578 failed in `TestCompiledProjectRepairsCorruptCompiledState/install` at `status_test.go:894` (empty status result). The other jobs succeeded; all ten preserve-lock vector cases and all task-specific tests passed in the failed macOS job's uploaded JSON stream. This failed gate remains recorded as failed, despite the earlier green run. Its failure was not classified as an external CI failure; no local-mirror fallback was used. The hosted failure's cause remains unknown.

The successor personally reran these standalone commands on the unchanged candidate (Go 1.25.5, darwin/arm64):

| Command | Exit | Evidence |
| --- | ---: | --- |
| `go build ./...` | 0 | recovery-build-01.log |
| `go vet ./...` | 0 | recovery-vet-01.log |
| `go test ./internal/envprofile -run 'TestRepairPreserveLock|TestPreserveLock|TestPlainRepairMigrationKeepsLockPrivate' -count=1 -v` | 0 | recovery-env-01.log; 10 of 10 vectors driven |
| `go test ./cmd/curator -run '^TestCompiledProjectRepairsCorruptCompiledState$' -count=1 -v` (inherited PATH) | 1 | recovery-compiled-01.log; both subcases failed at status --check after repair, a different assertion from the hosted failure |
| Same compiled-state command with PATH limited to the active Go toolchain and system tools | 0 | recovery-compiled-controlled-02.log; original install and upgrade test names/assertions unchanged |
| `go test ./cmd/curator -run 'TestEnvResolvePreserveLock|TestGlobalInitEnvResolveSubcommandHelp' -count=1 -v` (same controlled PATH) | 0 | recovery-cli-01.log |
| `git diff --check` | 0 | recovery-diff-check-01.log |
| `gofmt -l` on the eight changed/new Go files | 0 | recovery-gofmt-01.log; empty output |

The earlier ten-mutant results are reused, with their named failing tests and bounds above, because the candidate source and test bytes are unchanged. No new mutant or full local suite was run by the successor. The managed Story checkpoint remains `84cf54b94670ae9881f91b1c147feb4b75afddee`; observed local main `352c5c460739ec5849d677cad62c76875ed49f81` is six commits ahead. Evidence here is for the checkpoint candidate, not current trunk. No branch switch, commit or rebase was performed; upstream convergence belongs to the orchestrator.

The required `task-board handoff TASK-261014-2f0qnv --role developer` is being retried after these diagnostics. Its generated Change Request validation outcome is the authoritative result of that new hosted gate; the earlier green run alone does not satisfy a failed handoff. This recovery note is attached before invoking the synchronous handoff.


## Second successor recovery


No repository bytes changed in this run. The candidate remains uncommitted at checkpoint `84cf54b94670ae9881f91b1c147feb4b75afddee`, with tree `265d93f4093227dfa6e1a05ea49c72dbb5a3551e`. All ten task-path SHA-256 values match the previous evidence. The earlier green snapshot, both refused handoff snapshots, and the new gate snapshot have that exact tree. Local main is six commits ahead; this evidence describes the checkpoint candidate, and the orchestrator owns convergence.

## Hosted failure diagnosis

Revision 2 validation (`sh scripts/remote-gate.sh`) exited **1**, run [38557397441](https://github.com/relux-works/curator/actions/runs/38557397441). Uploaded macOS race JSON identifies exactly one failing test: `internal/buildrepo.TestResolvedTransportTotalDeadlineBoundsSlowFetch`. Its observed elapsed time was **15.7576585s**, exceeding its unchanged **7.5s** limit. Every other hosted job that ran succeeded, including all three ordinary platform tests, Linux race and lint. The changed candidate does not touch the failing test or its production module. This is a recorded failure, not proof of an external CI fault; no mirror or gate bypass is claimed.

Revision 1 remains recorded as exit **1** for `TestCompiledProjectRepairsCorruptCompiledState/install`; its cause remains unknown. A prior green result does not erase these failed handoffs.

## Personally executed standalone commands

Platform: darwin/arm64; Go 1.25.5.

| Command | Exit | Log |
| --- | ---: | --- |
| `go build ./...` | 0 | successor-build-01.log |
| `go vet ./...` | 0 | successor-vet-01.log |
| `go test ./internal/envprofile -run 'TestRepairPreserveLock\|TestPreserveLock\|TestPlainRepairMigrationKeepsLockPrivate' -count=1 -v` | 0 | successor-env-01.log |
| `go test ./cmd/curator -run 'TestEnvResolvePreserveLock\|TestGlobalInitEnvResolveSubcommandHelp' -count=1 -v` | 0 | successor-cli-01.log |
| `git diff --check` | 0 | successor-diff-check-02.log |
| `gofmt -l` on the eight changed/new Go files | 0; empty output | successor-gofmt-02.log |

CLI validation uses PATH restricted to the active Go toolchain and system tools, matching the previous hermetic local CLI evidence. No full local suite was run (R223). **10 of 10 AC/vector rows driven**: production `Resolve`, `StatusOf`, `cli.cmdEnvResolve`, `cli.cmdEnvStatus`, with individual names and bounds in the gaps outcome. The checked-in vector file is byte-identical to the assigned upstream family. No surface table was supplied; the existing brief gap remains declared. No assigned row is out of contract. Existing reviewer test names/assertions remain present; no test was edited during recovery.

Ten earlier narrowing mutant results are reused after verifying candidate byte identity and each readable log's named behavioral failure and terminal FAIL. No new mutant execution is claimed. Mutant table and survivor bounds remain in the gaps outcome; there are no survivors.

## Current hosted validation

Fresh direct `sh scripts/remote-gate.sh` exited **0** against snapshot `5dc7bdfca400f1ca3ebf98cc1ef3f6e1a64ce56e` (same candidate tree), run [38563918010](https://github.com/relux-works/curator/actions/runs/38563918010). All 20 jobs that ran succeeded; two configured jobs were skipped. Uploaded Ubuntu, Windows and macOS race streams each confirm all ten vectors and ten additional/retained named tests passed. The macOS race stream also confirms `TestResolvedTransportTotalDeadlineBoundsSlowFetch` passed. This new green gate does not erase the two earlier failures or claim their causes were established. No external-CI fallback was used.

## Logbook entry

Recovered two automatic handoff failures without modifying unrelated modules or weakening their assertions. Identified the second failed test from the uploaded JSON instead of the truncated workflow log. Preserved all failures and exact identities, personally validated the task scope, and launched the full required hosted gate. Candidate and Story checkpoint remain unchanged.


## Assigned-spec identity check

The inspected local specification prose also carries unrelated adapter amendments. The exact `6648e18` contents were read and checked separately: environments §1.3, §10.1, §10.4, the preserve-lock status note and manager §12.5 match the relevant inspected sections; §13 was read from the assigned revision. The local ten-vector file exactly matches `6648e18`. No preserve-lock spec/vector disagreement was found; unrelated adapter amendments do not expand this leaf.


## Handoff readiness

Candidate source, tests and configuration remain byte-identical to the recorded identity. Full direct hosted gate exit 0 is observed, landing-gate property suites are green, reviewer test names/assertions are retained and green, and **10 of 10 AC/vector rows are driven**. Ten narrowing mutants are reused on the unchanged identity; their named failures and no-survivor bounds remain in the gaps table. All validation commands are recorded with real exits. No assigned AC row is out of contract; no surface table was supplied, as declared above. There is no new rework diff. Changes outside the named module remain the original necessary CLI wiring, CLI tests/help and README changes. The managed Story HEAD stays at its checkpoint and all task files remain uncommitted. The orchestrator owns upstream convergence and integration. The developer role handoff is the next board action after these resources are attached.
