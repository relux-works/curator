# SNAP-LOCK revision 2 re-handoff

Task: TASK-261016-4zofp0 — implement-rc7-snapshot-nonrecovering-capture.

## Candidate and scope

Story base: `dccac3199aa5c543c589c129f5ef2bc437523a2a`.
Signed hosted snapshot: `a718fda37792b57b094d9cbca8984104d8e320da`.
Snapshot tree: `f0b08b96c097fcf57c1927355f9874079df23e5b`.
Signature verification returned `G`. The Story branch remains at its checkpoint; no Story commit, checkout, rebase or merge was performed by this run.

The operator withdrew revision 1 before review and converged the Story onto current trunk. This run restored exactly the supplied README paragraph immediately after the validation paragraph, once, while keeping every current trunk line. The new login-reference, PlantUML and Lychee text remains intact. No behavior changes were made. The historical refresh-blocker record was not edited.

All 15 non-README candidate files have exactly the same Git blobs as revision 1's green snapshot `1e6be490fee5768d0aa6808b825a20437058ef1c`. A separate whole-tree comparison excluding README, docs, diagrams and board data exited 0, establishing unchanged source, tests and validation configuration. The candidate still comprises the same 16 paths. The README delta is exactly the requested two-line insertion, confirmed against current HEAD.

Existing scope exceptions remain those declared in TASK-261016-4zofp0_handoff.md: shared managed/status content verification, the existing-state managerlock primitive, recovery-entry instrumentation, test-only platform fixtures, README documentation and the remote gate signing-policy correction. This refresh adds no new exception.

## Validation performed in this run

Environment: Go 1.25.5, darwin/arm64. Commands were standalone processes with direct output redirection, without tee.

| Command | Actual exit | Evidence |
| --- | ---: | --- |
| `go test -race ./internal/envprofile ./internal/managerlock ./internal/transaction -run 'TestMaterializeCapture\|TestPendingRecovery\|TestRecoveryObserver\|TestCaptureLock\|TestAcquireExistingHomeOnly' -count=1` | 0 | focused-01.log; all three packages green |
| `go build ./...` | 0 | build-01.log |
| `lychee --offline --include-fragments --no-progress README.md docs/environment-config.md docs/agent-logins.md` | 0 | links-01.log; 0 errors |
| `git diff --check` | 0 | diff-check-02.log |
| `sh -n scripts/remote-gate.sh` | 0 | gate-syntax-02.log |
| `git diff --exit-code HEAD -- ':(glob)**/*_test.go'` | 0 | retained-tests-01.log; all previously tracked tests retained unchanged |
| `git diff --exit-code 1e6be490fee5768d0aa6808b825a20437058ef1c a718fda37792b57b094d9cbca8984104d8e320da -- . ':!README.md' ':!docs/**' ':!diagrams/**' ':!.task-board/**'` | 0 | behavior-identity-01.log |
| `sh scripts/remote-gate.sh` after loading the configured agent Git environment | 0 | [Hosted gate](https://github.com/relux-works/curator/actions/runs/38882703220); all 20 enabled jobs green; two optional jobs skipped |

No new tests were added for restoring a documentation paragraph. Existing behavioral tests were rerun; the full gate supplies current candidate build, lint, property and runtime coverage.

## Existing evidence reused, with unchanged identities

The source/test/configuration equality above and the same Go version and host platform bind the existing nine narrowing controls to this unchanged implementation. They were executed in revision 1 and were not replayed in this README-only re-publication. Exact harness, expected-red logs, commands and exits remain in TASK-261016-4zofp0_mutant_harness.py, TASK-261016-4zofp0_mutants.json and TASK-261016-4zofp0_mutant_logs.md. Each named behavioral Go command exited 1 on its planted violation; the harness exited 0. This is reuse of recorded evidence, not a claim of new execution.

| Mutant | Gate narrowed to | Named failing test from retained evidence | Surviving bound |
| --- | --- | --- | --- |
| M-denied-class | Other read failures still refuse; permission denial admitted | TestMaterializeCaptureUnreadableJournal | None; killed |
| M-lock-name | Allow zero-length regular lock-name replacement | TestMaterializeCaptureAuthorizingDrift/lock-identity | None; killed |
| M-composition-policy | Allow composition policy enablement alone | TestMaterializeCaptureSameFragmentDrift/composition-policy | None; killed |
| M3-lock-only | Lock identity remains checked; other object identities ignored | TestMaterializeCaptureSameFragmentDrift/aba | None; killed |
| M-prepared | Refuse every pending phase except prepared | TestMaterializeCapturePendingDoesNotRecover/prepared | None; killed |
| M-corrupt-absence | Treat the planted corrupt journal as absent | TestMaterializeCaptureBadRecoveryEvidence/corrupt | None; killed |
| M-probe | Permit Codex vendor discovery only | TestMaterializeCaptureAndStatusNeverProbeVendor | None; killed |
| M-executable-mode | Ignore only execute permission bits | TestMaterializeCaptureSameFragmentDrift/mode | None; killed |
| M-provision-absent | Provision absent manager homes only | TestMaterializeCaptureDoesNotProvision/absent | None; killed |

The M3 plant is the leaf's required control. No source-text gate was introduced. No prior reviewer test was removed or renamed. Revision 1 received no reviewer verdict; its withdrawal was an operator convergence action, not a rejection or repeat-of class.

## Coverage and declared bounds

Coverage remains **5 of 5 capture-stage AC rows driven**, **0 of 2 CLI forms qualified**. The unchanged detailed coverage map is TASK-261016-4zofp0_coverage-map.md and the original handoff retains each paired positive/negative witness and production call site:

| AC row | Production call site | Named executable evidence |
| --- | --- | --- |
| Bounded non-recovering lock | CaptureMaterialize → beginCaptureOperation → AcquireExistingHomeOnly | TestMaterializeCaptureCurrentAndRelease; TestMaterializeCaptureWaitReselectsAndTimeout; TestAcquireExistingHomeOnlyDoesNotProvision |
| Pending/read failure refusal | beginCaptureOperation → PendingRecovery | TestMaterializeCapturePendingDoesNotRecover; TestMaterializeCaptureBadRecoveryEvidence; TestMaterializeCaptureUnreadableJournal; TestRecoveryObserverDetectsEmptyRecovery |
| Reselect/revalidate generation | CaptureMaterialize; ReadObject; revalidateObjects | TestMaterializeCaptureSerializesProfileWriters; TestMaterializeCaptureConcurrentRepair; TestMaterializeCaptureExternalTornGeneration; TestMaterializeCaptureSameFragmentDrift; TestMaterializeCaptureAuthorizingDrift; TestMaterializeCaptureStoreObjectIdentity |
| Metadata-only verification | resolveVerified(..., verifyHomeContent); StatusOf → homeState | TestMaterializeCaptureAndStatusNeverProbeVendor |
| Bound capture/release before transport | ReadObject; CaptureMaterialize return/deferred release | TestMaterializeCaptureObjectBounds; TestMaterializeCaptureCurrentAndRelease; TestMaterializeCaptureDoesNotProvision |

Brief gap: no catalog surface table was supplied; the supplied five-row paired behavior table is fully mapped in the retained coverage-map resource. No row is silently waived.

Out-of-contract rows retain their explicit leaf clauses: closure inventory and protected handle acquisition belong to SNAP-CLOSURE ("supplies the content capture callback"); actual export/manifest-only publication and fresh CLI configuration loading belong to SNAP-CLI ("integrates both"). No archive, destination, launcher, credential-provider or live vendor-runtime qualification is claimed. The local callback must remain bounded and use protected handles; arbitrary blocking callback code is not preempted. Windows POSIX executable-mode, vendor-shell sentinel, unreadability and held-lock replacement limits remain the declared platform bounds in the original handoff. Historical platform counts are not presented as new run counts.

## Re-publication logbook

- Restored the specified README paragraph once and preserved every converged trunk line.
- Verified all 15 carried non-README file blobs and the whole behavioral/configuration tree against the previously green revision.
- Repeated the focused race suite, build, local link validation and whitespace/shell/retained-test checks with exit 0.
- Ran the required full hosted gate on the signed exact candidate through its actual process return: exit 0. Twenty enabled jobs passed; two workflow-optional jobs skipped. No fallback was used.
- Preserved prior refresh-blocker history and all earlier outcome evidence.

## Current hosted leaf evidence

Artifacts from this exact run were downloaded with exit 0 for Linux, macOS and Windows. Terminal verdicts were deduplicated by package/test across aggregate and served JSON streams. Parent tests and subtests remain separate records. These are current-run records, not historical counts reused from revision 1.

| Platform | Pass | Skip | Bound |
| --- | ---: | ---: | --- |
| Linux | 54 | 0 | All leaf records green |
| macOS | 54 | 0 | All leaf records green |
| Windows | 46 | 6 | Existing declared platform limits; skipped cases are not passing qualification |

Windows skips: `TestMaterializeCaptureAndStatusNeverProbeVendor`, `TestMaterializeCaptureAuthorizingDrift/lock-identity`, `TestMaterializeCaptureSameFragmentDrift/mode`, `TestMaterializeCaptureStoreObjectIdentity/mode`, `TestMaterializeCaptureUnreadableJournal`, `TestCaptureLockRejectsReplacedIdentity`. These are the POSIX shell/vendor, permission-mode, unreadability and held-lock replacement bounds. The actual ABA/store/marker identity witnesses ran on Windows. Exact named terminal records are in `TASK-261016-4zofp0_rev2-platform-cases.json`.

One intermediate API status observation encountered a network timeout; its partial/empty response was not used as absence or as a gate verdict. A bounded retry succeeded. No CI fallback was used.

## Post-gate identity and handoff preconditions

The final temporary-index tree is `f0b08b96c097fcf57c1927355f9874079df23e5b`, exactly equal to the signed green snapshot. Story HEAD remains `dccac3199aa5c543c589c129f5ef2bc437523a2a`; the real index is empty and the snapshot signature verifies as `G`. A fresh remote query exited 0 and confirmed the temporary gate branch absent. Candidate files remain uncommitted for managed Change Request capture.

The authoritative checklist was read back with **30 of 30 items checked**. The reviewer-rejection conditional is not activated: revision 1 was withdrawn before any review verdict; no repeat-of rework class exists. Original coverage-map and out-of-contract declarations remain explicit above. The full gate includes the current property suites, lint and platform-case gates; all enabled jobs are green. All previously tracked test files remain unchanged under their original names.

Whitespace and shell-syntax checks were repeated as standalone commands because the first shell invocation did not preserve their individual exit statuses; both repeated commands returned 0. This was a correction to evidence collection, with no repository delta.

Current supporting resources: `TASK-261016-4zofp0_rev2-identity.json`, `TASK-261016-4zofp0_rev2-platform-cases.json`, `TASK-261016-4zofp0_rev2-gate.json` and `TASK-261016-4zofp0_rev2-validation-logs.md`. Existing detailed coverage and mutant resources remain unchanged and are explicitly reused under the identity proof.

## Handoff audit observation

The first `task-board handoff TASK-261016-4zofp0 --role developer` returned exit 0. A fresh compact status query returned `to-review`. The CLI also emitted `run_wrote_outside_worktree` under policy `warn`, listing unattributed changes in the shared control root, including other tasks' board records/resources and shared source paths. Its own-chain section attributed this run's eleven board/resource paths. This run performed no direct control-root edits; its project changes were confined to the managed worktree and its outcome staging files were written to a temporary directory as the assignment requires. The warning is preserved as an audit anomaly, not represented as a refusal or as proof that another actor authored each unattributed path. No automatic-approval rejection occurred.
