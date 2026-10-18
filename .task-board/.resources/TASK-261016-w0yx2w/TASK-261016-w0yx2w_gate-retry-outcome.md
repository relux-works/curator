# Round 2 hosted-gate retry — blocked by hosted Linux ACL baseline

Task: TASK-261016-w0yx2w — implement retained-handle snapshot destination publication.

No implementation change is delivered by this retry. The candidate remains uncommitted at the managed Story checkpoint. Candidate signed snapshot `b5a21d362c27956aaef1da5d8ec4cb389665593e`, tree `2e1811a9302b6d040adff5cf0a4692fa09b3ffb0`, signature **G**, matches the current complete worktree tree. Every recorded source/test/configuration identity member is unchanged. Local environment remains macOS 27.0 arm64 / Go 1.25.5. No Story commit, checkout, rebase, merge, Story-index staging or CI-configuration change was made.

## Actual hosted result

[Full hosted run 39067688154, attempt 2](https://github.com/relux-works/curator/actions/runs/39067688154/attempts/2) reran the entire original matrix on the exact signed candidate. `gh run rerun 39067688154` exited **0**, which admits the trigger only. The full gate is **not green**: **16 successful jobs, 1 failed job, 3 cancelled jobs, 2 conditional skipped jobs**. The remaining three long suites were cancelled only after the Linux race failure and its bounded diagnostic established the blocker; they are not passing evidence. Terminal status is `completed/cancelled`; terminal `gh run watch 39067688154 --exit-status` exited **1**.

| Job | Actual verdict | Job ID |
|---|---|---:|
| Lint | success | 117501715766 |
| Test (windows-latest) | cancelled | 117501715851 |
| Race (macos-latest) | cancelled | 117501715863 |
| Gate self-test (windows-latest) | success | 117501715864 |
| Interop conformance gate | success | 117501715920 |
| Naming gate | success | 117501715933 |
| Go 1.27.1 driver (ubuntu-latest) | success | 117501715943 |
| Go 1.26.0 driver (macos-latest) | success | 117501715945 |
| Gate self-test (macos-latest) | success | 117501715954 |
| Test (macos-latest) | cancelled | 117501715973 |
| Race (ubuntu-latest) | failure | 117501715977 |
| Go 1.27.1 driver (macos-latest) | success | 117501715996 |
| Test (ubuntu-latest) | success | 117501715998 |
| Go 1.25.5 driver (ubuntu-latest) | success | 117501716030 |
| Go 1.26.0 driver (windows-latest) | success | 117501716045 |
| Gate self-test (ubuntu-latest) | success | 117501716083 |
| Go 1.27.1 driver (windows-latest) | success | 117501716105 |
| Go 1.25.5 driver (windows-latest) | success | 117501716116 |
| Go 1.26.0 driver (ubuntu-latest) | success | 117501716117 |
| Go 1.25.5 driver (macos-latest) | success | 117501716140 |
| Candidate suite (${{ matrix.os }}) | skipped | 117501716661 |
| Test (self-hosted macOS) | skipped | 117501716737 |

The self-hosted Linux Test job passed `bash .github/ci/test-gate.sh .temp/ci-evidence/test`, exit **0**. The GitHub-hosted Linux Race job failed `bash .github/ci/test-gate.sh .temp/ci-evidence/race`, exit **1**, with `GO_TEST_FLAGS=-race` and `GO_TEST_TIMEOUT=60m`. Its snapshot valid controls refuse with `environment_materialize_destination_invalid: directory-acl` before publication. This is a real failing required gate, not an expected-red mutant and not a race-detector success.

The normal Linux artifact has **254 test/subtest pass verdicts, 0 failures, 0 skips** in `internal/envsnapshot`, plus the passing package verdict. The Linux race artifact has **178 test/subtest pass verdicts, 76 failures, 0 skips**, plus the failing package verdict. Its one non-JSON leading line is a Go dependency-download banner; every remaining event parses and the terminal package failure is present. All original seven reviewer test names remain in the candidate, unchanged, and pass in both Linux artifacts. Named Linux successes include access/default ACL rejection, prior ACL rejection, and all four `TestPublicationNoReplaceRace` competing-entry variants in the regular Test job.

## Bounded diagnosis and external blocker

The unchanged workflow routes the normal Linux Test job through the configured self-hosted Linux runner, but its `race` job uses `runs-on: matrix.os` directly. Job API labels confirm the failed job used GitHub-hosted `ubuntu-latest` (Ubuntu 24.04 image), while the green normal Test job used self-hosted Linux. No runner storage error recurred in this attempt.

[Diagnostic run 39070224888](https://github.com/relux-works/curator/actions/runs/39070224888) uses signed detached diagnostic commit `d12ce1c3f5b80f2fffb3e94d663ec3533df06996` (signature **G**). Relative to the candidate, it adds only ACL metadata logging in the existing Linux test and replaces the workflow with one focused hosted-Linux diagnostic job. **This is not a candidate or full-gate qualification.** It executes `go test -race -count=1 -run '^TestDestinationKernelACL$' -v ./internal/envsnapshot`, actual exit **1**. Both positive controls reproduce the intended `directory-acl` refusal. The diagnostic reports a readable **44-byte `system.posix_acl_default` on a retained ancestor above the workspace**, depth 7 from the fixture, filesystem type `0xef53`; the other probed ACL attributes report `ENODATA`. It logs sizes/errno/filesystem metadata only, never ACL payloads or host pathnames. This establishes actual extended/default ACL evidence in that runner baseline; it is neither an unreadable-result-as-absence inference nor a storage-failure inference.

The current Linux admission policy intentionally refuses extended access/default ACLs rather than treating their presence as absent. The existing candidate test contract specifically rejects a default ACL even when the directory mode remains private. This retry does not broaden that policy to fit an image. The diagnostic branch was deleted and absence verified. Temporary diagnostic test bytes were restored before constructing the diagnostic tree; the complete current worktree tree and every stored identity member are again exactly the original candidate.

**Input required:** the CI owner supplies a Linux race execution environment whose retained ancestry satisfies this boundary. Preferred option: route the Linux race job to the same qualified self-hosted Linux configuration as the regular Test job, with the normal configuration review and subsequent full candidate gate. Alternative: qualify an operator-owned ephemeral runner image with an appropriate ACL baseline before tests. Both retain actual syscall/ACL/race testing. This leaf does not authorize editing global runner ACLs, changing another owner's CI configuration, weakening the product ACL gate, skipping positive controls or accepting a failed hosted status. Such configuration is outside this leaf's declared module paths. No unresolved product model is being papered over.

After that input, rerun the full required gate on the appropriate exact candidate/configuration identity, inspect every job, update evidence and checklist, then invoke developer handoff. There is no developer handoff in this blocked retry.

## Round 2 AC evidence and bounds

Coverage remains **5 of 5 leaf AC rows driven through production library calls** by unchanged local macOS evidence and the green Linux Test job. **0 of 5 materialize CLI rows driven**: SNAP-CLI is explicitly separate. This ratio does not claim a passing full gate, Linux race acceptance, cancelled macOS/Windows full-suite acceptance, Windows publication support or source authorization. The earlier Round 2 outcome retains the complete four-row paired-test coverage map and out-of-contract clauses; the brief still lacks a catalog-keyed surface table, a declared brief gap.

All five AC rows below reuse the previously attached local command `go test -count=1 -race -coverprofile=.temp/TASK-261016-w0yx2w/coverage.out ./internal/envsnapshot`, exit **0**, on the unchanged macOS identity; this retry did not rerun it. All five also execute in the fresh Linux Test job's command above, exit **0**. The full Linux race job exits **1** and its positive admission refusal blocks acceptance; no row is silently promoted to full-gate success.

| AC row | Production call and named tests | New command/result and limit |
|---|---|---|
| 1. Read-only retained no-follow owner/ACL admission | `OpenDestination` -> `retainDirectories` / `destinationDirectory`; `TestDestinationReadOnlyAdmission`, `TestDestinationBoundary`, `TestDestinationOwnershipAndReadFailure`, `TestDestinationKernelACL`, `TestDestinationPriorKernelACL` | Linux Test gate **0**; access/default ACL cases execute. Linux Race gate **1**, valid ancestry refuses due actual default ACL. Windows runtime admission refusal was not newly qualified because its suite was cancelled. |
| 2. Protected identity/ancestry exclusion | `OpenDestination` -> `excludeProtected`; `TestDestinationProtectedIdentity` | Linux Test gate **0** for every supplied protected-root class. Inventory discovery remains the coordinator's responsibility; Linux Race gate **1** before these valid controls can admit. |
| 3. Absent/private-parent admission, retained-parent no-replace publication | `OpenDestination`, `Revalidate`, `Publish` -> `publishNoReplace`; `TestDestinationPrivateParent`, `TestDestinationReplacementRaces`, `TestPublicationNoReplaceRace`, `TestPublicationRetainedParentRace` | Linux Test gate **0**, real Linux renameat2 competition cases pass; existing empty destinations refuse per the owner decision. Linux Race gate **1** before publication. |
| 4. Actual prior tree identity and exact-generation no-op | `MatchPrior` / `Publish` prior branch -> `verifyDirectory`; `TestDestinationPriorIdentity`, `TestDestinationPriorMissingMetadata`, `TestDestinationPriorKernelACL`, `TestPublicationPrivateAndPrior` | Linux Test gate **0**, prior byte/mode/extra/version/generation refusals and unchanged no-op execute. Linux Race gate **1** due baseline admission. |
| 5. Private captured copies, post-verification and staging-only cleanup | `Publish` -> `createStagingDirectory`, `createStagingFile`, `cleanupStaging`, `verifyDirectory`; `TestPublicationCapturedIdentity`, `TestPublicationStagingOnlyCleanup`, `TestPublicationKernelStagingFailure`, `TestPublicationCleanupSubstitutedRoot`, `TestPublicationPostVerification` | Linux Test gate **0**, failure/interruption/cleanup/no-success cases execute. Linux Race gate **1** before staging. Same-user mutation inside private staging remains outside snapshots section 2; abrupt-death recovery and CLI output are not claimed. |

## Narrowing controls — exact-identity evidence reused

No product gate, behavioral test or mutant definition changed. The previous corrected macOS control command `go test -tags snapshotcontrols -count=1 -run '^TestDestinationNarrowingControls$' -v ./internal/envsnapshot` remains exact-identity green evidence (harness **0**, sixteen intended child failures **1**). It was **not rerun** in this retry; no Linux mutant execution is claimed. The following table is retained from the attached Round 2 outcome, including its historical survivor and precise bound; that survivor is not recast as a kill.


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



## Command ledger and scope

| Command or observation | Real exit | Bound |
|---|---:|---|
| `gh run rerun 39067688154` | 0 | Full-matrix trigger admitted, not a green verdict. |
| Linux Test: `bash .github/ci/test-gate.sh .temp/ci-evidence/test` | 0 | Self-hosted Linux, full required job and snapshot runtime evidence. |
| Linux Race: `bash .github/ci/test-gate.sh .temp/ci-evidence/race` | 1 | GitHub-hosted Linux, actual ACL-positive-control failures. |
| `gh run watch 39067688154 --exit-status --interval 55` | 143 | Local observer deliberately stopped after a bounded interval; does not cancel or attest the hosted gate. |
| `GOOS=linux GOARCH=amd64 go test -c -o .temp/TASK-261016-w0yx2w/gate-rerun/envsnapshot-diagnostic-linux.test ./internal/envsnapshot` | 0 | Diagnostic test compiles only; no Linux execution claim. |
| Diagnostic: `go test -race -count=1 -run '^TestDestinationKernelACL$' -v ./internal/envsnapshot` | 1 | Reproduced ACL baseline refusal; diagnostic only, not a green control. |
| `gh run cancel 39067688154` | 0 | Cancellation request for this attempt only. Terminal cancellation separately read back. |
| `gh run watch 39067688154 --exit-status` | 1 | Terminal failed/cancelled gate, not success. |
| Complete candidate-tree equality check and recorded SHA256-member comparison | 0 | No retained candidate changes. |
| Signature verification for original and diagnostic snapshots | 0 | Both report G; detached snapshots only. |
| Diagnostic branch deletion and fresh `git ls-remote --heads` readback | 0 | No diagnostic branch remains. |

The first raw job-log fetch exited **1** because the CLI refused terminal escape sequences; the explicit `--allow-escape-sequences` retry exited **0**. An initial generic JSON reader exited **1** on the dependency-download banner; the bounded reader identified that single banner and parsed all subsequent events. An initial diagnostic push had an invalid shell-expanded refspec and published nothing; the corrected explicit refspec push exited **0**. None of these are represented as passing validation commands. No status-masking test pipeline was used.

There is **no retained rework diff**. The diagnostic-only exception was the temporary detached workflow override used solely to probe a hosted runner, plus logging in the existing Linux test. Neither is in the candidate. Root README and module README exceptions remain exactly as justified in the earlier implementation outcome. No full local mirror, review waiver, remote status override, relaxed ACL gate, test skip or infrastructure mutation was used.

Logbook: the reported storage repair permits the self-hosted Linux suite to pass. The exact-candidate rerun exposed a different baseline on the workflow's GitHub-hosted Linux race job: a real default ACL above the workspace. A signed isolated diagnostic reproduced it without changing product behavior or the candidate. The full gate remains red/unverified; required checklist rows remain unchecked. Persisted evidence is `TASK-261016-w0yx2w_gate-retry-evidence.json`, this outcome, and the updated validation history. No local process or hosted run is left pending.
