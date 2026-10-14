# TASK-261014-2f0qnv — second successor recovery

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
