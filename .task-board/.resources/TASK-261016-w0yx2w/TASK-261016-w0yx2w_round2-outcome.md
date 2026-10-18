Current disposition: **ready for review** after the green Round 3 retry. The signed candidate is `57de425979331d8fc21d91a7f7b349d5af2b5840`; [full hosted run 39075205524](https://github.com/relux-works/curator/actions/runs/39075205524) passed (standalone gate exit 0, 20 successful jobs, 2 conditional skips). The Story candidate remains uncommitted for Change Request handoff. The original Round 2 blocker report below is retained verbatim as historical evidence and is superseded by the Round 3 terminal results.

# Round 2 — SNAP-PUBLISH implementation; external validation blocker

Task: TASK-261016-w0yx2w — implement retained-handle snapshot destination publication.

The library implementation is retained uncommitted. This is a blocked producer outcome, not a review handoff or accepted delivery. The full remote gate is not green. No product/platform decision remains open: curator-spec main `15d2fc2be22c2b013faa8f637101947b1ff20830` resolves the staging boundary and is the normative source read for this round (snapshots sections 2/4/5, environments section 5, CIP-0014 destination/idempotence rows).

## External blocker and recovery

[Hosted run 39067688154](https://github.com/relux-works/curator/actions/runs/39067688154) used signed snapshot `b5a21d362c27956aaef1da5d8ec4cb389665593e`, tree `2e1811a9302b6d040adff5cf0a4692fa09b3ffb0`. Its Linux vet, lint, interop and driver jobs failed with `no space left on device`, before qualifying the candidate. This is explicit provider evidence, not an inference from missing status. The failed build/vet stages and failing driver fixtures name the same storage error. No repository-caused diagnostic was established from these failures; that does not qualify the unexecuted tests.

Failed job IDs: interop `117498647115`; lint `117498647121`; default Linux test/vet `117498647173`; Go 1.25.5 driver/vet `117498647291`; Go 1.27.1 driver fixtures `117498647424`.

An individual-job retry was attempted using `gh run rerun 39067688154 --job 117498647115`. The CLI exited 0 but reported `job 117498647115 cannot be rerun`; no retry execution or green verdict is claimed. The overall run was still active. Repeating the entire matrix against exhausted shared storage would not resolve the external condition.

The local gate process was stopped by its recorded exact PID (exit 137); its TERM handler performs cleanup but does not exit, so TERM alone did not stop it. Cancellation was requested for this run only. The temporary gate branch is confirmed absent. Hosted readback is terminal `completed/cancelled`; no required full-gate success is claimed. No validation process is being left for a subsequent turn.

**Input needed:** the runner operator must restore sufficient free temporary/build storage without deleting active jobs' data, then rerun the failed/cancelled jobs on the exact signed snapshot or execute `sh scripts/remote-gate.sh` against the retained candidate. Reuse genuine green jobs only for unchanged source/test/configuration/environment identity. The full gate and missing Linux/Windows runtime evidence remain blocking; no local mirror substitutes for hosted requirements. This work does not authorize shared-runner cleanup, service interruption, unrelated data deletion or a weakened CI configuration.

Alternatives considered: rerunning a single failed job (not admitted while active); local focused checks (executed, narrower OS/contract); a full local mirror (cannot satisfy required remote statuses or Linux/Windows runtime identity). Recommendation: repair runner capacity and rerun the exact candidate, without changing the product implementation or required deadlines. No additional product decision is required.

## Implementation and boundary

- `OpenDestination`, `Revalidate`, `MatchPrior`: retained no-follow admission, caller owner/private mode/ACL checks, physical protected identity/ancestry, absent-only new destinations, full actual-tree prior verification. Existing empty directories refuse.
- `Destination.Publish(context.Context, Publication)`: clones and validates captured manifest/descriptor/content bytes before writing, stages beside the absent name using an unpredictable invocation-created private directory, creates independent 0700 directories and 0600/verified-0700 files through retained descriptors, syncs files/directories, then calls `publishNoReplace`.
- macOS backend: `renameatx_np(RENAME_EXCL)`; Linux backend: `renameat2(RENAME_NOREPLACE)`. No overwrite/copy fallback. Every kind of competing name refuses and is preserved, including an empty directory created after the final revalidation.
- Published success is gated by retained target identity and SNAP-ID verification against the captured manifest, plus owner/ACL and retained-chain rechecks. Byte/size/mode/metadata/entry-set mismatches and post-rename cancellation return no `Verified` result and leave the published entry for the operator. No destructive rollback.
- Cleanup walks only the invocation's recorded staging objects in reverse through retained parent handles. It does not enumerate an arbitrary tree, match a filename prefix, clean the destination, follow a substituted link, or recursively delete a path. Observed substitutions refuse cleanup. Same-user mutation inside private staging remains outside the normative guarantee; no conditional-unlink claim is made. Cooperative context interruption cleans unpublished staging; abrupt process death may leave a private partial stage and emits no success.
- Unsupported platforms, including Windows, explicitly refuse admission/publication. Windows owner/ACL/reparse support is not claimed.

The capture coordinator remains responsible for source authorization, complete protected-root inventory and verified execute metadata. Publication accepts captured bytes, not source filesystem paths or authority inferred from a self-minted manifest. Source capture/closure and SNAP-CLI integration are separate leaves, explicitly excluded by the current brief. This library does not emit CLI output, discover inventory, mutate native/managed sources, or authorize launch.

## Round 2 acceptance matrix

**5 of 5 leaf AC rows driven through the production library API on macOS**, with the OS bounds below. **0 of 5 rows driven through the materialize CLI**; SNAP-CLI is expressly separate. This is measured library coverage, not full cross-platform acceptance. Tests are present in the uncommitted managed candidate for the Change Request snapshot; they are not represented as committed on the Story branch.

All rows below use `go test -count=1 -race -coverprofile=.temp/TASK-261016-w0yx2w/coverage.out ./internal/envsnapshot`, exit **0**, on macOS 27.0 arm64 / Go 1.25.5. Narrowing controls use `go test -tags snapshotcontrols -count=1 -run '^TestDestinationNarrowingControls$' -v ./internal/envsnapshot`, corrected run exit **0** (16 killed plants; each named child command exits **1**, intentionally red).

| AC row | Production call site and named tests | Result and bounds |
|---|---|---|
| 1. No-write validation; retained ancestors/target; owner/ACL/no-follow refusal | `OpenDestination` -> `retainDirectories` / `destinationDirectory`; `TestDestinationReadOnlyAdmission`, `TestDestinationBoundary`, `TestDestinationOwnershipAndReadFailure`, `TestDestinationKernelACL`, `TestDestinationPriorKernelACL` | Local exit 0. Real macOS ACLs and read failures; foreign-owner case executed as an unprivileged caller. Linux ACL runtime not qualified by the failed gate; Windows refuses. |
| 2. Protected equal/inside/containing identity | `OpenDestination` -> `excludeProtected`; `TestDestinationProtectedIdentity` | Local exit 0 for every supplied root class and similar-prefix positive siblings. Case aliases are exercised only on a folding filesystem. Inventory discovery/session allocator remain coordinator work. |
| 3. Absent/private-parent or exact-prior only; no-replace publication and replacement races | `OpenDestination` / `Revalidate` / `Publish` -> `publishNoReplace`; `TestDestinationPrivateParent`, `TestDestinationReplacementRaces`, `TestPublicationNoReplaceRace`, `TestPublicationRetainedParentRace`, `TestPublicationPrivateAndPrior` | Local exit 0. Empty target is negative under the owner decision. Four real competing creations run between final validation and rename. Moved-ancestor test demonstrates retained-parent writes and refusal of success without outside traversal. Linux backend compiled/static-linted; runtime gate not green. |
| 4. SNAP-ID actual prior verification and exact-generation no-op | `MatchPrior` -> `revalidate` -> `inspectDestination` -> `verifyDirectory`; `Publish` prior branch; `TestDestinationPriorIdentity`, `TestDestinationPriorMissingMetadata`, `TestDestinationPriorKernelACL`, `TestPublicationPrivateAndPrior` | Local exit 0. Bytes/mode/extras/unreadable/unknown version/new generation refuse; no-op and generation refusal preserve manifest inode/mtime. Observation does not lease against subsequent same-user mutation. |
| 5. Private independent copies, execute modes, failure/interruption and staging-only cleanup | `Publish` -> `createStagingDirectory` / `createStagingFile` / `cleanupStaging` / `verifyDirectory`; all `TestPublication*`, especially `TestPublicationCapturedIdentity`, `TestPublicationStagingOnlyCleanup`, `TestPublicationKernelStagingFailure`, `TestPublicationCleanupSubstitutedRoot`, `TestPublicationPostVerification` | Local exit 0. Kernel permission failure, context cancellation, outside sibling/link preservation and post-publication mismatch are exercised. SNAP-ID checks exact semantic modes/nlink. Fresh macOS provenance is OS-created, not imported; no other xattrs admitted by the test. Execute authorization belongs to capture. No process-kill recovery, Windows private-ACL qualification or CLI success-output qualification claimed. |

## Coverage map

The brief has a paired-test table but no catalog-keyed surface table: this is a **brief gap**, not a waived map. The following map preserves all four supplied paired rows, in order, following the coverage-map template. Relevant catalog surfaces are destination conflicts, managed-output pruning, read boundary, identity and error surface.

| Supplied surface / paired row | Attacking tests | Killed narrowing mutants | Out-of-contract inputs / exact clause |
|---|---|---|---|
| Absent private destination versus links, owner/ACL, protected identity aliases | `TestDestinationReadOnlyAdmission`, `TestDestinationBoundary`, `TestDestinationKernelACL`, `TestDestinationOwnershipAndReadFailure`, `TestDestinationProtectedIdentity`, `TestDestinationPrivateParent` | M2; owner-root-parent; mode-group-writable-ancestor; protected-equal-identity; empty-target; private-parent; acl-one-permit-kind | Existing empty destination explicitly refuses under owner decision 2026-10-17; unqualified platforms refuse under Round 2 Windows qualification clause. |
| Stable retained parent versus ancestor/target replacement and competing creation | `TestDestinationReplacementRaces`, `TestPublicationNoReplaceRace`, `TestPublicationRetainedParentRace` | ancestor-directory-replacement; target-identical-replacement; publication-replaces-empty | No target competition waived. Windows support remains explicit refusal under Round 2. |
| Exact prior versus altered actual bytes/modes/extras/unreadable/version/generation | `TestDestinationPriorIdentity`, `TestDestinationPriorMissingMetadata`, `TestDestinationPriorKernelACL`, `TestPublicationPrivateAndPrior` | M6-saved-manifest-bytes; prior-acl-one-member; unrelated-one-entry | No prior-verification cases waived. |
| Staging failure versus unsafe cleanup, interruption and partial published output | `TestPublicationStagingOnlyCleanup`, `TestPublicationKernelStagingFailure`, `TestPublicationCleanupSubstitutedRoot`, `TestPublicationPostVerification` | cleanup-foreign-stage-link; cleanup-omits-one-file (corrected test); post-verify-one-byte-mismatch | Same-user mutation **inside private staging** is outside snapshots section 2's guarantee (Round 2). Process death may retain private staging; no crash-recovery/CLI-emission claim. |

## Mutant evidence

Every current listed plant below is killed on macOS with a named behavioral failure and child exit **1**. The harness itself exits **0** only after verifying those failures. Build/setup failure and a missing/skipped test cannot count. Linux-specific ACL/no-replace plants are present but **not executed**; no Linux mutant kill is claimed.

| Mutant | What the gate is narrowed to admit | Named failing test | Result / surviving bound |
|---|---|---|---|
| cleanup-foreign-stage-link | Cleanup removes exactly one foreign stage-looking sibling link in addition to recorded objects | `TestPublicationStagingOnlyCleanup/io-failure` | Killed, exit 1 |
| cleanup-omits-one-file, initial test selection | Leave only the recorded AGENTS.md cleanup member behind | None; `TestPublicationStagingOnlyCleanup/io-failure` passed | **Survivor**, child exit 0: interruption occurred before AGENTS.md was staged. This selection proves nothing about that member. |
| cleanup-omits-one-file, corrected test selection | Same weakening; interrupt after the entire stage is populated | `TestPublicationStagingOnlyCleanup/before-rename` | Killed, exit 1 |
| post-verify-one-byte-mismatch | Only `file-identity` mismatch may return success after publication; other verification refusals remain | `TestPublicationPostVerification/bytes` | Killed, exit 1 |
| publication-replaces-empty | Weaken no-replace to ordinary rename, retaining refusal of a nonempty competing directory but admitting replacement of an empty one | `TestPublicationNoReplaceRace/empty` | Killed, exit 1; macOS backend only executed |
| M2-linked-ancestor | Allow only a directory component named target to follow a link | `TestDestinationBoundary/linked-ancestor` | Killed, exit 1 |
| M6-saved-manifest-bytes | Trust saved manifest only after an actual-file identity failure | `TestDestinationPriorIdentity/bytes` | Killed, exit 1 |
| owner-root-parent | Admit root-owned immediate parent for an unprivileged caller | `TestDestinationOwnershipAndReadFailure` | Killed, exit 1 |
| mode-group-writable-ancestor | Admit only mode 0770 as an otherwise forbidden writable ancestor | `TestDestinationBoundary/shared-ancestor` | Killed, exit 1 |
| protected-equal-identity | Allow equality while other ancestry checks remain | `TestDestinationProtectedIdentity` | Killed, exit 1 |
| ancestor-directory-replacement | Permit replacement only for retained component parent | `TestDestinationReplacementRaces/ancestor-directory` | Killed, exit 1 |
| target-identical-replacement | Permit changed target inode only when replacement mode is 0700 | `TestDestinationReplacementRaces/target-identical` | Killed, exit 1 |
| empty-target | Treat just an empty existing target as absent | `TestDestinationReadOnlyAdmission` | Killed, exit 1 |
| private-parent | Additionally admit mode 0755 immediate parent | `TestDestinationPrivateParent` | Killed, exit 1 |
| unrelated-one-entry | Admit only an unrelated single-entry target | `TestDestinationBoundary/unrelated` | Killed, exit 1 |
| acl-one-permit-kind | Admit only the real kernel permit ACE with rights mask 0x24 | `TestDestinationKernelACL` | Killed, exit 1 |
| prior-acl-one-member | Skip only AGENTS.md prior-member ACL check | `TestDestinationPriorKernelACL` | Killed, exit 1; Linux variant is unrun |

## Command ledger

All validation commands ran directly, without tee or a status-masking pipe. Paths below are repository-relative. No full prior-run evidence was reused as current qualification: this round changes source/tests, so focused checks ran anew. Unchanged reviewer tests were included in the package run under their original names (`TestReviewerNullChannel`, `TestReviewerManifestOnly`, `TestReviewerPortableLeftDrift`, `TestReviewerRecapture`, `TestReviewerFreeHunt`, `TestReviewerUnicodePositive`, `TestReviewerGreekCollision`). Their source is unchanged.

| Exact command | Actual exit | Scope / evidence |
|---|---:|---|
| `go test -count=1 -run '^TestPublication' -v ./internal/envsnapshot` | 1 | Initial diagnostic run failed only because a fresh macOS object carried OS-generated provenance; corrected the test's imported-metadata assertion. |
| `go test -count=1 -race -coverprofile=.temp/TASK-261016-w0yx2w/coverage.out ./internal/envsnapshot` | 0 | macOS full package, original reviewer tests included; 86.9% statement coverage. |
| `go test -tags snapshotcontrols -count=1 -run '^TestDestinationNarrowingControls$' -v ./internal/envsnapshot` (initial) | 1 | Harness truthfully rejects the surviving cleanup plant described above. |
| Same control command (corrected selection) | 0 | 16 named macOS mutant kills, each nested command expected-red exit 1. |
| `golangci-lint run ./internal/envsnapshot/...` | 0 | macOS, 0 issues. |
| `GOOS=linux GOARCH=amd64 golangci-lint run ./internal/envsnapshot/...` | 0 | Linux static analysis only, 0 issues. |
| `go build -o .temp/TASK-261016-w0yx2w/curator ./cmd/curator` | 0 | Local production command compiles; not materialize CLI execution. |
| `GOOS=linux GOARCH=amd64 go test -c -o .temp/TASK-261016-w0yx2w/envsnapshot-linux.test ./internal/envsnapshot` | 0 | Linux compilation only; no runtime verdict. |
| `GOOS=windows GOARCH=amd64 go test -c -o .temp/TASK-261016-w0yx2w/envsnapshot-windows.test.exe ./internal/envsnapshot` | 0 | Windows compilation only; no runtime verdict. |
| `sh scripts/remote-gate.sh` | 137 | Intentionally stopped after confirmed remote storage failures; not green, not an expected-red product control. Hosted run cancellation is separately recorded. |
| `git apply --check --reverse .temp/TASK-261016-w0yx2w/TASK-261016-w0yx2w_round2-candidate.patch` | 0 | Exported patch corresponds to current uncommitted candidate. |
| `git diff --check` | 0 | Whitespace check. |
| `git log -1 --format='%H %G?' b5a21d362c27956aaef1da5d8ec4cb389665593e` | 0 | Signature reports G; detached validation commit only. |

`go test -tags snapshotcontrols -run '^TestNarrowingControls$'` for unchanged SNAP-ID controls was not rerun; this outcome claims only the executed destination/publication control command. No local full repository suite, full Linux runtime suite, Windows runtime suite or new CLI integration was executed. Cancelled/partial logs are not passing evidence.

## Scope, state and logbook

HEAD stays at `1724b4a5` (recorded in the attached source identity); no Story commit, switch, rebase, merge or index staging was performed. A temporary separate index exports the patch and confirms that the candidate tree equals the signed gate tree. `HEAD..main` is 0 against the local main ref; that is not a fresh remote-main attestation. The orchestrator owns convergence and integration.

Code/tests stay under `internal/envsnapshot/destination*.go` and `publication*.go`, with module README updates. The sole additional changed path is root `README.md`, required by the repository's tools-documentation rule to explain focused publication/control commands and link the platform bounds. `UNRESOLVED_QUESTIONS.md`'s previous task-specific staging question was removed because the owner resolved it; that restores its tracked bytes and leaves no diff there. No existing SNAP-ID source, reviewer test, consumer or workflow configuration changes.

Logbook: Round 2 applies owner option A and the absent-only rule. The initial metadata test was corrected for OS-created provenance; the initially surviving cleanup plant's selection was corrected and re-executed, retaining both observations. Required hosted validation was attempted and halted on independently logged shared-storage exhaustion. The candidate and all local evidence are preserved; cross-platform/full-gate obligations remain open. The signed-commit sentence in the attached Round 2 brief conflicts with the managed Story assignment's explicit no-commit/handoff checkpoint rule, so the Story candidate remains uncommitted; only the gate creates a signed detached snapshot.

Next lifecycle: restore runner storage, run/reuse exact-identity hosted evidence as appropriate, update this outcome/validation, satisfy the full gate and OS bounds, then `task-board handoff TASK-261016-w0yx2w --role developer`. No handoff is attempted while those preconditions are unmet.


## Round 3 — Linux race routing and full hosted qualification

The current round changes only `.github/workflows/ci.yml`,
`docs/self-hosted-runner-setup.md` and `tools/ciworkflow/isolation_test.go`
beyond the retained Round 2 candidate. These exceptions are explicitly
required by the Round 3 CI-owner instruction. Snapshot source, snapshot tests,
SNAP-ID and all original reviewer tests remain byte-identical to Round 2.
The workflow change routes Linux `race` exactly like `test` and
`go-family-qualification`: enabled pushes, dispatches and same-repository PRs
use qualified self-hosted Linux; fork PRs stay hosted. The race job also gets
the mandatory existing private-Go-state initializer and self-hosted cache
restore prohibition. macOS routing is unchanged. The runner setup document
states the access/default ACL requirement and the known fork-image limit;
a hosted fork control refusal is not a pass.

The existing routing test now explicitly requires all three jobs and executes
the real race expression for 36 event/variable/repository/platform combinations.
The two new narrowing plants preserve the routing tokens and run the behavioral
suite in a subprocess, requiring its named assertion failure and real exit 1.
The actual GitHub job API independently confirms the Linux race job used labels
`self-hosted, relux-linux, curator`; local expression evaluation alone is not
claimed as scheduler evidence.

| Mutant | What it narrows the gate to | Named test that fails | Result / surviving bound |
|---|---|---|---|
| `fork_admitted` | Admit fork PRs to self-hosted Linux while keeping variable and platform guards | `TestLinuxRoutingKeepsForksHosted/race` | Killed: child exit 1 on fork PR, enabled Linux |
| `dispatch_hosted` | Lose only the enabled trusted dispatch route while retaining pushes and PR behavior | `TestLinuxRoutingKeepsForksHosted/race` | Killed: child exit 1 on dispatch, enabled Linux |

No current local LF/CRLF routing plant survived. The earlier Windows setup failures below are not kills. Historical snapshot mutant results and the
corrected cleanup survivor bounds above are retained without reinterpretation.
Their product source/test/definition and local toolchain identity are unchanged;
they are reused, not rerun. No Linux snapshot mutant execution is claimed.

| Round 3 command | Actual exit | Bound |
|---|---:|---|
| `go test -count=1 -v ./tools/ciworkflow/` | 0 | macOS; real workflow parser, routing evaluation, initializer behavior and two intended-red routing children |
| `golangci-lint run ./tools/ciworkflow/...` | 0 | macOS static analysis; no issues |
| `bash .github/ci/gate-selftest.sh` (first invocation) | 1 | Fixture signing setup failed because agent Git environment was missing; 266 passed, 1 failed, 2 reported capability skips. Not green evidence. |
| `bash .github/ci/gate-selftest.sh` (retry with agent Git environment) | 0 | 267 passed, 0 failed, 2 explicit local rustup-absence capability skips; hosted job results recorded separately. |
| `sh scripts/remote-gate.sh` with agent Git environment | 0 | Signed detached candidate `57de425979331d8fc21d91a7f7b349d5af2b5840`, tree `2e20b225f2794c9e524ee69332ed156424c612d8`; [hosted run 39075205524](https://github.com/relux-works/curator/actions/runs/39075205524) |
| `git apply --check --reverse .temp/TASK-261016-w0yx2w/TASK-261016-w0yx2w_round3-retry-candidate.patch` | 0 | Full exported patch matches candidate |
| `git diff --check` | 0 | Whitespace validation |
| `git log -1 --format='%H %T %G?' 57de425979331d8fc21d91a7f7b349d5af2b5840` | 0 | Signature G, signed gate snapshot only |
| Independent temporary-index full-tree equality and stored SHA256 comparison | 0 | Worktree equals signed gate tree; only workflow differs among Round 2 identity members |

Checklist note: the obsolete empty-target acceptance item was replaced with the binding absent-only/no-replace/post-verification criterion before checking it. No superseded empty-target success claim is made.

### Round 3 AC coverage and qualification bounds

Coverage: **5 of 5 leaf AC rows driven through production library entry
points**, with the binding absent-only and staging-boundary decisions applied.
**0 of 5 CLI rows driven**; SNAP-CLI remains a separate leaf. The four-row paired
coverage map in Round 2 remains applicable and is not expanded into an invented
catalog table: the brief has no catalog-keyed attack-surface table, a declared
brief gap. No leaf row is silently treated as out of contract.

| AC row | Production call sites and named tests | Current command, exit and bound |
|---|---|---|
| 1. Read-only retained no-follow ownership/ACL admission | `OpenDestination` -> `retainDirectories` / `destinationDirectory`; `TestDestinationReadOnlyAdmission`, `TestDestinationBoundary`, `TestDestinationOwnershipAndReadFailure`, `TestDestinationKernelACL`, `TestDestinationPriorKernelACL` | Linux and macOS regular/race `bash .github/ci/test-gate.sh` exit 0 on the exact candidate; actual kernel ACL paths, not schema vectors. Windows supports explicit refusal only. |
| 2. Protected identity and ancestry exclusions | `OpenDestination` -> `excludeProtected`; `TestDestinationProtectedIdentity` | Linux and macOS regular/race `bash .github/ci/test-gate.sh` exit 0 on the exact candidate; tests supply protected roots. Coordinator-owned inventory discovery is outside this library leaf. |
| 3. Absent/private-parent admission and no-replace publication | `OpenDestination`, `Revalidate`, `Publish` -> `publishNoReplace`; `TestDestinationPrivateParent`, `TestDestinationReplacementRaces`, `TestPublicationNoReplaceRace`, `TestPublicationRetainedParentRace` | Linux and macOS regular/race `bash .github/ci/test-gate.sh` exit 0 on the exact candidate; real renameat2/renameatx_np races, including a competing empty destination. Existing empty targets refuse. |
| 4. Actual prior tree identity and exact generation no-op | `MatchPrior` / `Publish` -> `verifyDirectory`; `TestDestinationPriorIdentity`, `TestDestinationPriorMissingMetadata`, `TestDestinationPriorKernelACL`, `TestPublicationPrivateAndPrior` | Linux and macOS regular/race `bash .github/ci/test-gate.sh` exit 0 on the exact candidate; bytes, modes, extras, unreadable/version/generation failures preserve target; unchanged prior is a no-op. |
| 5. Private captured bytes, complete post-verification and staging-only cleanup | `Publish` -> `createStagingDirectory`, `createStagingFile`, `cleanupStaging`, `verifyDirectory`; `TestPublicationCapturedIdentity`, `TestPublicationStagingOnlyCleanup`, `TestPublicationKernelStagingFailure`, `TestPublicationCleanupSubstitutedRoot`, `TestPublicationPostVerification` | Linux and macOS regular/race `bash .github/ci/test-gate.sh` exit 0 on the exact candidate; mismatch returns no success and leaves the published entry. Same-user mutation inside private staging is excluded by snapshots section 2; abrupt-death recovery and CLI success output are not claimed. |

Windows production `OpenDestination` and `Destination.Publish` are driven by
`TestDestinationUnqualifiedPlatform`, asserting `platform-unqualified` refusal
and an absent destination. This is not Windows owner/ACL/reparse/no-replace
runtime qualification. No platform policy was weakened to fit a CI image.
The original seven reviewer tests remain under their original names and byte
identity. Runtime execution evidence is recorded below only when obtained.

### Round 3 first full gate and CRLF rework

[First run 39071503518](https://github.com/relux-works/curator/actions/runs/39071503518)
ran on signed snapshot `0c9a7d1f0b7a18867a9c926897e8df10cd46cb84`,
tree `96e0f4146dbd91e190ad5d1b3161b1bf0589959c`. The real command
`sh scripts/remote-gate.sh` exited **1**: **19 successful jobs, one failed
Windows Test job, two conditional skipped jobs**. This is a source-caused
failure, not infrastructure, not an expected-red plant, and not a green gate.
The Windows `tools/ciworkflow` package failed because the new mutation harness
searched for a raw LF-only `race` job header. The Windows checkout used CRLF,
so both plants stopped at `race job not found` before exercising the behavior.
Those two Windows plants are **not kills**; their bound is no Windows harness
qualification on that first candidate. The original named failures are
`TestLinuxRoutingRejectsRacePlants/fork_admitted` and
`TestLinuxRoutingRejectsRacePlants/dispatch_hosted`.

The correction matches the header with either line ending. Named positive
controls `TestLinuxRoutingRejectsRacePlants/LF/positive` and
`TestLinuxRoutingRejectsRacePlants/CRLF/positive` require the real behavioral
child to exit **0**. Both LF and CRLF variants of `fork_admitted` and
`dispatch_hosted` then require child exit **1** from
`TestLinuxRoutingKeepsForksHosted/race`, including the expected assertion
text. Setup/parse/no-job failures cannot count as kills. These six cases and
the full local package passed in `go test -count=1 -v ./tools/ciworkflow/`,
exit **0**; `golangci-lint run ./tools/ciworkflow/...` also exited **0**.
Only `tools/ciworkflow/isolation_test.go` changed between the two signed
candidates. No snapshot source/test/configuration or admission policy changed.

The first run's snapshot packages all passed, with no snapshot failures or
skips: Linux regular/race **254/254**, macOS regular/race **252/252**, Windows
regular **104** test/subtest verdicts. All seven original reviewer names passed
in each Linux/macOS stream. Windows `TestDestinationUnqualifiedPlatform`
passed through both production refusal entry points; it does not qualify
Windows publication. Both Linux jobs used self-hosted Linux; macOS jobs and
Windows used their hosted images. The first-run evidence resource retains
every job verdict and package test name.

A strict first artifact reader exited **1** on the Linux race stream's one
leading Go dependency-download banner. A bounded classifier then exited **0**:
that single known banner was identified and every later line parsed, with a
terminal package pass. The retained helper command
`python3 .temp/TASK-261016-w0yx2w/round3-read-evidence.py test-linux race-linux race-macos`
exited **0**; separate `test-macos` and `test-windows` reads exited **0**.
Missing/truncated logs are not counted as passing evidence.

The second full gate uses signed snapshot
`57de425979331d8fc21d91a7f7b349d5af2b5840` (signature **G**),
tree `2e20b225f2794c9e524ee69332ed156424c612d8`, at
[run 39075205524](https://github.com/relux-works/curator/actions/runs/39075205524).
It is **green**, with the standalone remote-gate process exiting **0**. Story HEAD remains the checkpoint and its
candidate remains uncommitted; the gate's detached signed commits do not move
that branch. A full rerun is required by the changed harness identity and the
Round 3 exact-candidate instruction. The earlier green workflow self-test is
reused locally: its script/workflow/configuration are unchanged; the changed
Go harness has been tested afresh and passed in the full hosted gate.

Logbook (Round 3): apply the CI-owner routing decision without relaxing ACL
admission; the real Linux regular and race jobs pass. The first full run finds
a CRLF bug in this round's own mutation harness. Preserve that real failure,
correct it with explicit LF/CRLF controls, re-sign only the detached candidate,
and rerun the full hosted matrix. The second full gate passed before the developer handoff.

Historical first-candidate Windows plant accounting:

| Mutant | What it was intended to narrow | Named behavioral failure | Result / bound |
|---|---|---|---|
| `fork_admitted`, first Windows candidate | Admit fork PRs while retaining variable/platform guards | None: `TestLinuxRoutingRejectsRacePlants/fork_admitted` stopped at `race job not found` | Survivor for gate-evidence accounting: plant not applied; no Windows kill claimed |
| `dispatch_hosted`, first Windows candidate | Keep enabled dispatch on hosted Linux | None: `TestLinuxRoutingRejectsRacePlants/dispatch_hosted` stopped at `race job not found` | Survivor for gate-evidence accounting: plant not applied; no Windows kill claimed |

The current named LF/CRLF controls supersede that broken harness only where
actually executed. A green detached gate snapshot does not itself satisfy the
runner pool's separate three-consecutive-green-main rollout qualification.


### Round 3 retry terminal hosted evidence

[Run 39075205524](https://github.com/relux-works/curator/actions/runs/39075205524) reports **20 successful jobs and 2 conditional skips** on the exact signed retry snapshot. Every job and step verdict was inspected. The remote-gate process exited **0**; no local mirror or synthetic status is used. The skipped optional self-hosted macOS and candidate-suite jobs are not passes.

| Hosted job | Terminal result |
|---|---|
| Race (macos-latest) | success |
| Gate self-test (ubuntu-latest) | success |
| Go 1.26.0 driver (ubuntu-latest) | success |
| Gate self-test (macos-latest) | success |
| Gate self-test (windows-latest) | success |
| Naming gate | success |
| Go 1.26.0 driver (macos-latest) | success |
| Race (ubuntu-latest) | success |
| Lint | success |
| Go 1.27.1 driver (ubuntu-latest) | success |
| Test (ubuntu-latest) | success |
| Test (windows-latest) | success |
| Test (macos-latest) | success |
| Go 1.27.1 driver (windows-latest) | success |
| Interop conformance gate | success |
| Go 1.25.5 driver (macos-latest) | success |
| Go 1.25.5 driver (windows-latest) | success |
| Go 1.26.0 driver (windows-latest) | success |
| Go 1.27.1 driver (macos-latest) | success |
| Go 1.25.5 driver (ubuntu-latest) | success |
| Test (self-hosted macOS) | skipped |
| Candidate suite (${{ matrix.os }}) | skipped |

| Snapshot stream | Passing test/subtest verdicts | Failures / skips | Runtime bound |
|---|---:|---:|---|
| test-linux | 254 | 0 / 0 | Qualified platform publication and all seven original reviewer names |
| race-linux | 254 | 0 / 0 | Qualified platform publication and all seven original reviewer names |
| test-macos | 252 | 0 / 0 | Qualified platform publication and all seven original reviewer names |
| race-macos | 252 | 0 / 0 | Qualified platform publication and all seven original reviewer names |
| test-windows | 104 | 0 / 0 | Explicit unqualified-platform refusal only |

The exact hosted commands were `bash .github/ci/test-gate.sh .temp/ci-evidence/test` on Linux, macOS and Windows, and `bash .github/ci/test-gate.sh .temp/ci-evidence/race` with `GO_TEST_FLAGS=-race` on Linux/macOS; each corresponding hosted step exited **0**. The artifact parser `python3 .temp/TASK-261016-w0yx2w/round3-read-retry-evidence.py test-linux race-linux test-macos race-macos test-windows` exited **0** after checking complete terminal JSON package results, no snapshot failures/skips, required reviewer names and all named LF/CRLF routing controls. Linux race has one classified dependency-download banner; every later line parses.

On all five streams, `TestLinuxRoutingKeepsForksHosted/race` and both LF/CRLF positive controls pass. All four named LF/CRLF `fork_admitted`/`dispatch_hosted` controls pass by requiring the intended behavioral child failure and child exit **1**. The current routing plants have no survivor on Linux, macOS or Windows; the first candidate's unapplied Windows plants remain historical survivors above. Snapshot mutation controls were reused only under their unchanged recorded macOS source/test/configuration/toolchain identity; Linux snapshot mutation runs are not claimed.

The source-identity check, full temporary-index tree comparison, signature verification, reverse-patch check and `git diff --check` all exit **0**. Story HEAD remains `1724b4a52b4c2d163c0a3e6ab7c2ad2bbe54ac0c`, with the 21-path candidate uncommitted. The detached signed tree is `2e20b225f2794c9e524ee69332ed156424c612d8`. The canonical outcome, validation ledger, current patch, source identity and current hosted evidence are attached before handoff.
