# BUG-261004-2v9pbz — expanded snapshot budget bypass via repeated blobs

Developer candidate ready for review. Changes remain uncommitted in the Story worktree.

## Implementation and acceptance evidence

1. The shared `proveRepository` walk reserves emitted-file count, aggregate file-content bytes, and canonical framing bytes before copying/appending any file. Canonical framing includes the frozen 24-byte header, type byte, both uint64 lengths, UTF-8 path bytes, and content. Canonical storage is allocated exactly once at the reserved size. Unique-object count/byte limits remain separate; their aggregate addition now uses an overflow-safe subtraction check. `docs/repository-admission-limits.md` documents this policy.
2. `MaxTreeEntries` (default 400,000) counts every expanded child entry, including repeated directories and empty directories, before retaining its path. The walk checks cancellation at each tree and child entry, including cached-object walks.
3. RED FIRST through real Git + `AdmitLocal`: one 4096-byte blob at 64 paths and a repeated-tree DAG were admitted before the fix (negative assertions failed, exit 1). One-blob positive and 64-distinct negative controls passed before and after. Restored-candidate boundary tests admit exactly at the canonical/file/object/entry limits and refuse one past them; an empty-tree DAG proves file count cannot substitute for entry accounting.
4. `AcquireNetwork` is exercised through the existing POSIX transport fixture, real Git fetch and batch reader: the one-blob control is admitted; 64 aliases refuse with `build_repository_incomplete_source`. The argv log proves the shared raw-object proof was reached. This fixture does not contact an external server; Windows execution is left to the hosted gate.
5. Frozen canonical bytes for the in-budget one-blob and empty-DAG snapshots, SHA-1/SHA-256 network/local canonical parity, and the targeted rc.14 buildrepo conformance tests passed. Selected normative families report 163/163 driven cases, zero known gaps/bounds/skips. This is targeted coverage, not a full repository conformance claim.

## Direct validation commands and actual exit codes

All Go test runs used `GOFLAGS=-work` and `-count=1`. No validation was piped through tee. Commands returned their own statuses.

Regression command R:
```sh
GOFLAGS=-work go test -count=1 ./internal/buildrepo -run 'Test(AdmitLocal|AcquireNetwork)ExpandedSnapshotBudget$' -v
```
- Before production changes: exit **1**. Local aliases, local repeated-tree DAG, and network aliases were wrongly admitted; controls passed.
- After production changes: exit **0**.

Boundary command B:
```sh
GOFLAGS=-work go test -count=1 ./internal/buildrepo -run 'Test(AdmitLocal(ExpandedSnapshotBudget|SnapshotBudgetBoundaries|ExpandedTreeEntryBudget)|AcquireNetworkExpandedSnapshotBudget|SnapshotBudget.*)$' -v
```
- First expanded boundary run: exit **1** due to fixture errors (header counted as 23 rather than 24; attempted overwrite of Git's existing read-only empty blob).
- After correcting those fixtures: exit **0**.

Mutation commands:
```sh
GOFLAGS=-work go test -count=1 ./internal/buildrepo -run '^TestSnapshotBudgetRefusesBeforeEmission$' -v
GOFLAGS=-work go test -count=1 ./internal/buildrepo -run 'Test(AdmitLocal|AcquireNetwork)ExpandedSnapshotBudget$' -v
```
- Move reserve check after content copy/file append: exit **1**; the walker retained 4 files instead of 3. This supplements the entry regressions because discarded partial output is invisible at the admission API.
- Reserve only uncached blob objects: exit **1**; local aliases, local DAG and network aliases were admitted.
- Required mutants killed: **2/2**. Production source restored byte-for-byte from a pre-mutation copy, not from Git.

Targeted conformance-only command C (with the selected rc.14 `CURATOR_CONFORMANCE_ROOT`):
```sh
GOFLAGS=-work go test -count=1 ./internal/buildrepo -run 'Test(NetworkAndLocalSHA1SHA256RawObjectParity|RawObjectAndLFSPinnedConformanceFixtures|PackIndexConformanceAndExactSSHWrapper|ExternalRepositoryAcquisitionConformance|LocalPackedHeadAndSnapshotMaterialization|LocalConfigAndAdministrationAdversarialBoundaries|ReleasedSkillBuildSchemaCases|CanonicalRepositorySourceVectors|ReleasedSourceIdentityVectors)$' -v
```
- Exit **0**.

Latest restored-candidate command T (same rc.14 root):
```sh
GOFLAGS=-work go test -count=1 ./internal/buildrepo -run 'Test(AdmitLocal(ExpandedSnapshotBudget|SnapshotBudgetBoundaries|ExpandedTreeEntryBudget|EmptyTreeDAGEntryBudget|ReleasesSealedPrivateRoot|ReleasesPrivateRootOnRefusal)|AcquireNetworkExpandedSnapshotBudget|SnapshotBudget.*|NetworkAndLocalSHA1SHA256RawObjectParity|RawObjectAndLFSPinnedConformanceFixtures|PackIndexConformanceAndExactSSHWrapper|ExternalRepositoryAcquisitionConformance|LocalPackedHeadAndSnapshotMaterialization|LocalConfigAndAdministrationAdversarialBoundaries|ReleasedSkillBuildSchemaCases|CanonicalRepositorySourceVectors|ReleasedSourceIdentityVectors)$' -v
```
- Exit **0**, 56.297 seconds, 18 top-level tests, no skips; all 24 budget regression/control scenarios passed. Attached task-scoped log contains this latest run with its machine-specific WORK path removed.
- Selected root manifest SHA-256: `6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5` (released rc.14 CI conformance identity).

Other validation:
```sh
GOFLAGS=-work golangci-lint run ./internal/buildrepo
GOFLAGS=-work go build ./internal/buildrepo
git diff --check
```
- Package lint: exit **0**, zero issues (both initial and restored-candidate runs).
- Package build after restoration: exit **0**.
- Whitespace validation: exit **0**.
- gofmt commands: exit **0**.

## Bounds of this evidence

Only the named targeted buildrepo checks were rerun locally; no previously attached test evidence was reused. The full local suite was intentionally not run under the binding produce-mode instruction. Hosted Change Request validation is the arbiter and remains pending; hosted green is not claimed. No Windows gate or heavy OOM experiment was run locally. LOGBOOK.md and CHANGELOG.md were not edited, as instructed; findings are preserved in this outcome and board notes instead.

## Handoff checklist reconciliation

The initial developer handoff exited 1 because checklist items 2 and 8 required hosted-green evidence and a logbook entry. Those generic items conflict with the binding instruction to run only targeted checks, leave full validation to the hosted gate, and never edit LOGBOOK.md/CHANGELOG.md. They were replaced with explicit local-verification and board-outcome criteria; no hosted success or logbook edit is claimed. Acceptance criteria remain unchanged, and hosted validation remains pending for review.
