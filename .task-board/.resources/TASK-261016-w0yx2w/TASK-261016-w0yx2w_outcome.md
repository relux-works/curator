# SNAP-PUBLISH: blocked publication contract; retained admission candidate

Task: TASK-261016-w0yx2w. This is a blocked producer outcome, not an accepted implementation or a review handoff.

## Decision required

The empty-target publication clause requires replacing an existing caller-owned empty directory while preserving any competing replacement. The mandatory race row permits substitution between the last identity/emptiness check and publication. A retained target descriptor does not make parent/name rename or removal conditional on that target's identity. Caller ownership and mode 0700 do not exclude another process with the same user's authority.

The exercised macOS rename/unlink interfaces do not provide that conditional operation:

- `RENAME_EXCL` safely publishes an absent name but returns EEXIST for an existing empty directory.
- `RENAME_EXCL | RENAME_SWAP` returns EINVAL.
- Rechecking the retained inode and emptiness, then removing by parent/name, deletes a substituted empty directory. No-replace rename can subsequently succeed, concealing that deletion.
- Ordinary rename atomically replaces whichever empty entry currently occupies the name.
- Exchange atomically moves even a competing nonempty user directory into the staging name, before its identity can be checked. This already changes the forbidden namespace; subsequent rollback has another replacement window.
- A held directory `flock` does not prevent these replacements. Removing with `unlinkat(retained_target, "", AT_REMOVEDIR)` or `"."` fails (ENOENT/EINVAL), rather than conditionally removing that retained entry.

These are observed limitations of the tested interfaces, not a claim to have exhausted every possible operating-system facility. A different trusted publication service or namespace-exclusivity design would be an ownership/architecture change, not an implementation detail that this producer may invent.

Recommended owner decision: revise new publication to require an **absent** destination under a permitted private parent; retain fully verified exact-prior no-op. Callers allocate a fresh nonexistent leaf. The tradeoff is that pre-created empty directories cease to be accepted.

Alternative: define and qualify an enforced exclusive-writer namespace/authority before admitting existing empty targets. Advisory locks are insufficient for the stated hostile substitution tests; the authority, ownership and failure-preservation contract must be specified. A contract revision allowing a different in-place empty-target protocol would also need its own explicit atomicity/interruption semantics and qualification. This outcome does not claim either alternative implemented.

Exact input needed: an owner decision reconciling existing-empty acceptance, replacement of its directory entry, and preservation under competing same-authority replacement. Update the normative contract/leaf accordingly, or supply a qualified platform operation/exclusivity design that meets all three. No force flag, unsafe overwrite, mock-only publication or cleanup workaround was added.

## Sources and reproducible constraint evidence

Normative curator-spec main: `15d970111277b1f68b3f7a4ddd63bc339b2c0a3f`. Read snapshots sections 2/4/5, environments section 5, and CIP-0014 destination/idempotence requirements from the fetched main revision. The read-only unit has no identified normative contradiction; the unresolved issue is the empty-entry publication mechanism and authority model.

- [Snapshot contract, pinned](https://github.com/relux-works/curator-spec/blob/15d970111277b1f68b3f7a4ddd63bc339b2c0a3f/protocol/managed-home-snapshots.md#2-destination-and-publication).
- Local macOS `renamex_np(2)` documents EXCL, SWAP, their incompatibility and the corresponding EEXIST/EINVAL outcomes. The six attached syscall witnesses execute those behaviors on macOS 27.0 / arm64.
- [Linux rename(2)](https://man7.org/linux/man-pages/man2/rename.2.html) documents the analogous distinction: NOREPLACE rejects an existing name; EXCHANGE swaps existing objects and cannot combine with NOREPLACE. Linux execution of the probe was **not run**; this is documentation evidence only.
- [Darwin VFS source](https://github.com/apple-oss-distributions/xnu/blob/main/bsd/vfs/vfs_syscalls.c) is supporting platform reference; the observed witness results, not unexecuted source inference, establish the macOS limitation reported here.

`TASK-261016-w0yx2w_publication-probe_test.go` contains isolated capability witnesses, not production acceptance tests. Place it in a scratch package under the repository's `.temp/` and run `go test -count=1 -v <scratch-package>`. It imports the repository's existing x/sys dependency. All fixture mutations are confined to private probe allocations and fixtures are removed. Its passing assertions demonstrate the unsafe candidate algorithms' effects; they do not prove a publication implementation correct.

## Candidate retained in the Story worktree

Story HEAD remains `58b2aac337f73fb313893611d9de28e9a4fcff0a`, originally equal to refreshed origin/main. No Story commit, switch, rebase, merge, index staging or trunk publication was performed.

The uncommitted first coherent unit provides:

- `OpenDestination`: read-only no-follow admission with retained ancestors, target and caller-supplied protected roots; owner/mode/ACL checks; identity/physical ancestry exclusions.
- `Destination.Revalidate`: rejects replaced ancestors/targets, competing destination creation, changed permissions/protected roots and prior-tree drift.
- `Destination.MatchPrior`: re-verifies actual bytes, modes and entry set using SNAP-ID and compares exact manifest bytes. Prior owner/ACL checks cover every actual member. A changed/missing prior manifest cannot become absence or unrelated-content fallback.
- `Destination.Close`: releases retained handles and deletes nothing.
- A deliberately refusing backend on unqualified platforms, including Windows. It is not Windows reparse/ACL/runtime support.

The coordinator must supply the complete existing manager/store/home/credential/repository/workspace inventory. Inventory discovery, source capture authority and materialize CLI integration are absent. There is no production materialize call site. Admission is an observation, never a capability authorizing pathname rename after a recheck.

Executable diff scope: only `internal/envsnapshot/destination*.go`, with module documentation in `internal/envsnapshot/README.md`. The sole additional path is `UNRESOLVED_QUESTIONS.md`, required by the repository instructions for unresolved decisions and added after stopping publication work. No shared filesystem consumer, source acquisition package, codec, SNAP-ID implementation or existing reviewer test was changed. The attached patch includes all 12 files. The 11 original files match the signed gate tree byte-for-byte; the later root decision document is separately hashed in the source identity artifact and does not change executable/test/configuration identity.

Candidate snapshot (off-branch validation only): `6d1a905a88e111eda227886581eff3d3b9011902`. The required gate's normal synthetic-commit invocation was narrowly wrapped in task scratch to retain the configured author/signing identity and verify its signature before push. The wrapper changed only the off-branch commit-tree call, not source, gate behavior, HEAD or index. The signature verified successfully. The temporary remote gate branch was removed; no remote producer branch or PR was created.

## Measured acceptance coverage

**0 of 5 AC rows driven through the production materialize entry point.** That CLI integration is unimplemented. The three library entry points above are exercised directly; no helper result is presented as CLI conformance. All five AC rows remain required, not waived as out of contract.

| AC row | Named library call sites and evidence | Unimplemented/unqualified bound |
|---|---|---|
| 1: retained no-follow, owner/ACL admission | OpenDestination; TestDestinationReadOnlyAdmission, TestDestinationBoundary, TestDestinationOwnershipAndReadFailure, TestDestinationKernelACL | Materialize integration; Windows owner/ACL/reparse backend; Linux runtime. Foreign-owner test requires an unprivileged caller. |
| 2: protected identity and ancestry | OpenDestination -> excludeProtected; TestDestinationProtectedIdentity | Actual machine inventory discovery and session allocator integration; Windows identity backend. Case-alias branch is filesystem-conditional and has no separate execution count. |
| 3: admission and safe publication | OpenDestination / Revalidate; TestDestinationReplacementRaces | Staging, no-replace publication, empty-entry replacement and failure cleanup. Empty-entry authority/atomicity decision is blocking. |
| 4: exact prior idempotence | MatchPrior -> revalidate -> inspectDestination -> verifyDirectory; TestDestinationPriorIdentity, TestDestinationPriorMissingMetadata, TestDestinationPriorKernelACL | Actual materialize no-op/output; prior observation is not a continuing lease. |
| 5: independent private output and cleanup | Existing-target read-only owner/mode/ACL validation only | No files are published; execute metadata, imported xattrs/ACLs/set-id, own-staging cleanup and interruption qualification remain outstanding. |

The brief contains a paired-test table, but no catalog-keyed surface table. This is a **brief gap**, not a silent waiver of coverage-map requirements. The following map preserves every supplied paired-test row; incomplete rows block full handoff.

| Supplied paired row | Attacking tests | Killed narrowing mutants | Out-of-contract inputs (AC clause) / bounds |
|---|---|---|---|
| Absent/empty outside protected roots versus links, owner/ACL, protected aliases | TestDestinationReadOnlyAdmission; TestDestinationBoundary; TestDestinationOwnershipAndReadFailure; TestDestinationKernelACL; TestDestinationProtectedIdentity | M2; owner-root-parent; mode-group-writable-ancestor; protected-equal-identity; acl-one-permit-kind | None waived. Admission only, macOS runtime; other platforms as declared above. |
| Retained parent/empty target versus ancestor/target replacement and competition | TestDestinationReplacementRaces | ancestor-directory-replacement; target-empty-replacement | None waived. Revalidation only; publication witness and its positive publish-once case are missing. |
| Intact identical prior versus modified bytes/mode/extras/unreadable/unknown/new generation | TestDestinationPriorIdentity; TestDestinationPriorMissingMetadata; TestDestinationPriorKernelACL | M6-saved-manifest-bytes; prior-acl-one-member | None waived. Library exact-prior observation, not materialize CLI no-op. |
| Staging failure / cleanup substitution / interruption | Six platform capability probes explain the blocking empty-target mechanism; no production staging path exists | No product-stage mutant; uncovered | None waived. Unimplemented and blocks full handoff. |

## Narrowing controls

The opt-in control harness copies the package to isolated scratch, weakens one gate and runs the named behavioral test with `go test -count=1 -json`. Exit 1 plus a named Action=fail is required. A compile error, skip, missing event or child success cannot count as a kill. No source-text scanner attests product acceptance; mutation-target matching is only control setup.

| Mutant | Gate narrowed to admit | Named failing test | Result / survivor bound |
|---|---|---|---|
| M2-linked-ancestor | A directory component named target may resolve a link; final new component remains no-follow | TestDestinationBoundary/linked-ancestor | Killed; child exit 1 |
| M6-saved-manifest-bytes | Trust saved manifest only for a file-identity failure, retaining other refusals | TestDestinationPriorIdentity/bytes | Killed; child exit 1 |
| owner-root-parent | Root-owned immediate publication parent admitted for an unprivileged caller | TestDestinationOwnershipAndReadFailure | Killed; child exit 1 |
| mode-group-writable-ancestor | Mode 0770 ancestor admitted; other unsafe modes still refuse | TestDestinationBoundary/shared-ancestor | Killed; child exit 1 |
| protected-equal-identity | Permit equality to the protected root while other ancestry guards remain | TestDestinationProtectedIdentity | Killed; child exit 1 |
| ancestor-directory-replacement | Permit changed ancestor identity for component parent | TestDestinationReplacementRaces/ancestor-directory | Killed; child exit 1 |
| target-empty-replacement | Permit a replacement target inode when its directory mode is 0700 | TestDestinationReplacementRaces/target-empty | Killed; child exit 1 |
| unrelated-one-entry | Admit unrelated nonempty target with exactly one entry | TestDestinationBoundary/unrelated | Killed; child exit 1 |
| acl-one-permit-kind, initial rights 0x06 | Permit only an ACE whose rights equal 0x06 | None; TestDestinationKernelACL passed | Survivor; child exit 0. The actual kernel fixture uses rights 0x24, so this first plant was not applied to the exercised violation. It proves no ACL bound. Superseded by corrected plant below. |
| acl-one-permit-kind, corrected rights 0x24 | Permit only the real add_file/add_subdirectory ACE mask | TestDestinationKernelACL | Killed; child exit 1 |
| prior-acl-one-member | Skip prior-member ACL check only for AGENTS.md | TestDestinationPriorKernelACL | Killed; child exit 1 |

The Linux-specific default-ACL control and Linux prior-home ACL control compile but were not executed. No Linux mutant kill is claimed. Unchanged SNAP-ID reviewer tests remain under their original names and pass in the normal package run. The pre-existing opt-in SNAP-ID control suite was not replayed; no new claim is made about that unexecuted command.

## Verification ledger

Local environment: Go 1.25.5; macOS 27.0; arm64. Logs are bundled in `TASK-261016-w0yx2w_validation.txt`; exact source identities are in `TASK-261016-w0yx2w_source-identity.json`.

| Direct command | Actual exit | Evidence and bound |
|---|---:|---|
| go test -count=1 -race -coverprofile=.temp/TASK-261016-w0yx2w/coverage.out ./internal/envsnapshot | 0 | package-05.log; all ordinary package/reviewer tests; 87.9% package statement coverage |
| go test -tags snapshotcontrols -count=1 -run '^TestDestinationNarrowingControls$' -v ./internal/envsnapshot | 0 | controls-03.log; 10 corrected plants killed, each child exit 1 |
| golangci-lint run ./internal/envsnapshot/... | 0 | lint-05.log; 0 issues |
| GOOS=linux golangci-lint run ./internal/envsnapshot/... | 0 | linux-lint-01.log; static analysis only |
| go build -o .temp/TASK-261016-w0yx2w/curator ./cmd/curator | 0 | build-02.log; current candidate identity |
| GOOS=linux GOARCH=amd64 go test -c -o .temp/TASK-261016-w0yx2w/envsnapshot-linux.test ./internal/envsnapshot | 0 | linux-compile-01.log; compilation only, executable not run |
| GOOS=windows GOARCH=amd64 go test -c -o .temp/TASK-261016-w0yx2w/envsnapshot-windows.test.exe ./internal/envsnapshot | 0 | windows-compile-01.log; compilation only, executable not run |
| git diff --check | 0 | Existing tracked diff clean; gofmt applied to every new Go file |
| git apply --reverse --check .temp/TASK-261016-w0yx2w/TASK-261016-w0yx2w_admission.patch | 0 | Entire 12-file patch matches the retained uncommitted candidate, including the post-blocker decision document |
| go test -count=1 -v ./.temp/TASK-261016-w0yx2w/publicationprobe | 0 | publication-constraint-01.log; six isolated capability witnesses, not product qualification |
| sh scripts/remote-gate.sh | 137 | Intentionally stopped after contract blocker; hosted run cancelled before execution. NOT GREEN. No local mirror substituted. |

[Hosted gate run](https://github.com/relux-works/curator/actions/runs/38890917179): actual conclusion cancelled. Last inspection showed required jobs queued; cancellation is not an external-CI-failure claim. The local gate process was stopped by its exact recorded PID after cancellation and branch cleanup, preventing its automatic cancelled-run retries. No long command was left running at turn end.

Earlier diagnostic failures are not green evidence: focused-02.log and package-01.log exited 1 because the test's real deny-delete ACL blocked fixture cleanup; cleanup was corrected. controls-01.log exited 1 because the initial ACL plant survived; controls-02.log and controls-03.log exited 0 after using the actual rights mask. lint-01.log and lint-03.log exited 1 on named lint diagnostics subsequently resolved. The lint-02 invocation shared a shell with a later build and its separate exit was not retained; it is not counted as standalone gate evidence. Final standalone lint/build commands above supply current evidence. package-02/03/04 passed at earlier source identities; they are not substituted for the current package-05 run.

## Logbook and next step

- The first required status mutation refused until the leaf estimate was recorded as Fibonacci 5; development was then entered synchronously.
- The chosen first coherent unit followed the brief's partial-implementation allowance. It remains an uncommitted candidate, not acceptance of remaining work.
- A real ACL fixture exposed an unsafe test teardown assumption; its ACL is now removed before deleting that isolated fixture. Production admission never performs cleanup writes.
- A surviving ACL plant exposed a mismatch between assumed and actual kernel rights. A bounded numeric-only kernel probe established 0x24; the corrected behavioral control is killed. No principal identifiers were persisted in public evidence.
- The next publication design step exposed the empty-entry replacement/exclusivity conflict above. Product-code changes stopped before adding a workaround; only isolated platform probes, evidence and the mandatory unresolved-question document were added afterwards.
- The full leaf, all runtime platforms, complete paired table, materialize production integration, full landing gate and reviewer handoff remain pending. No incomplete checklist item tied to those claims is checked.
- Resume only after the owner resolves the precise empty-target contract/authority decision. Reuse current green evidence only while its source/test/configuration/environment identity remains unchanged; finish publication, integration, all paired negative/positive cases and platform qualification, then rerun the full gate and hand off.
