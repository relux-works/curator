## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(3))

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] Test invoking executable identity via symlinked launcher path
- [x] Canonicalization before checks if test fails; substitution rejection still fail-closed
- [x] Evidence note if current path already canonicalizes; go test green
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches
- [x] Reword the Windows symlink skip in identity_test.go:24 to a reason skip-classes.tsv classifies (reuse "creating Windows symlink requires host support") or add a host-capability row
- [x] Add platform-cases.tsv rows requiring the new identity cases per runner, incl. the windows-only junction cases; keep make ledger-check green

## Notes
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260822-507162, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260822-507162)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260822-1e48d3, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260822-1e48d3)
Finding: internal/godriver/identity.go canonicalized the running manager with filepath.EvalSymlinks, while the rest of the driver (selectToolchain, verifySelectedRoot, mustPhysical) uses physicalPath. Identical on unix - so darwin/linux symlinked launch already canonicalized - but on Windows a manager installed under a directory junction was refused with worker_identity_invalid: cannot canonicalize the manager executable, before any substitution check ran. That is the operator launch shape treated as a fault, the inversion profiles/manager.md orders against. Fix: resolveExecutableIdentity now uses physicalPath(absolute). New tests: identity_test.go (four launch shapes resolve to one identity; substitution through a link still fails closed; real re-exec through a shim link via a test-only identity probe mode; full Build handshake with managerExecutable pointed at a link) and identity_windows_test.go (junction resolve + substitution behind a junction). Evidence: os.Executable reports the LINK path on both darwin/arm64 and windows/amd64, so canonicalization is not a no-op there. Red-to-green proved on the native Windows host: pre-fix baseline junction test exits 1, post-fix full godriver suite exits 0 (248.8s).
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-1e48d3, pid=16223, exit=0)
RUN-260822-507162 (second of two duplicate spawn runs for this task; RUN-260822-1e48d3 finished its work but ended without running the handoff, leaving the item in development with uncommitted changes in .temp/TASK-260822-27bvo4/worktree).

FINDING: os.Executable does not canonicalize on macOS or Windows - measured from a started process, not assumed - so resolveExecutableIdentity canonicalizing before the substitution checks is load-bearing, not decorative. On Windows it was also wrong: it used filepath.EvalSymlinks while the rest of the driver resolves with physicalPath. A manager installed under a directory junction (GitHub Actions tool-cache shape) was refused with worker_identity_invalid: cannot canonicalize the manager executable before any substitution check ran.

FIX: internal/godriver/identity.go:56 canonicalizes with physicalPath(absolute). Bit-identical on unix; GetFinalPathNameByHandle on Windows. Substitution battery unchanged and now applied to the physical file.

TESTS: internal/godriver/identity_test.go (5 tests) and identity_windows_test.go (2 junction tests), plus a test-only identityProbeMode in main_test.go and startRawWorkerFrom/launchPath in worker_test.go. Four launch shapes resolve to one identity; substitution through a link and behind a junction still fails closed (dangling, directory, empty, link loop, hard link, retargeted shim/junction, mutated bytes, physical file swapped for a link); a real re-exec through a link; the full Build handshake; and the worker process itself started from a link.

GATES (real exit codes, standalone): darwin go build ./... 0, gofmt -l cmd internal 0, go vet ./... 0, go test ./internal/godriver/ -count=1 0 (48.8s), golangci-lint run 0. Native Windows host (ssh win, go1.25.5 windows/amd64): go test ./internal/godriver/ -count=1 exit 0 (93.7s) with the fix; the junction test on the pre-fix identity.go exits 1 - expected red, that is the defect. GOOS=windows golangci-lint reports 4 findings, all in files this change does not touch.

BRANCH: task/TASK-260822-27bvo4-symlink-launcher-507162 in .temp/TASK-260822-27bvo4/RUN-260822-507162, uncommitted per standing orders. It is the union of both runs: peer test files and probe mode adopted wholesale, three extra cases grafted in (link loop, physical file swapped for a link, worker started from a link).

ANOMALY: host data volume at 99-100% (under 4 GiB free). The first go test ./... failed across five packages with no space left on device; unrelated to this change, but the full-repo suite is not reliably runnable here at default parallelism.
agent completed: [implementer] developer (claude) (exit=124)
spawn run completed: claude (run=RUN-260822-507162, pid=13733, exit=124)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260822-50e157, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260822-50e157)
RUN-260822-50e157 (third spawn). RUN-260822-507162 merged both earlier lanes and checked every item, then ended exit=124 (timeout) before the handoff, leaving the item in development. This run changed no product code: it re-verified the merged tree independently, added the darwin mutation check the earlier runs lacked, and ran the handoff.

VERIFIED TREE: .temp/TASK-260822-27bvo4/RUN-260822-507162, branch task/TASK-260822-27bvo4-symlink-launcher-507162, base 6a9b201, uncommitted per standing orders. Fix is one line: internal/godriver/identity.go:56 canonicalizes with physicalPath(absolute) instead of filepath.EvalSymlinks(filepath.Clean(absolute)); the substitution battery below it is untouched and now applies to the physical file.

NEW EVIDENCE (this run): the earlier runs proved red-to-green only on Windows, where the EvalSymlinks/physicalPath difference lives - on unix the two are the same function, so that mutation is a no-op on darwin and said nothing about the darwin tests. Stripping canonicalization entirely (canonical := filepath.Clean(absolute)) exits 1 with all six identity tests red. It also reddens the pre-existing substitution test, which uses no link at all: t.TempDir() returns /var/folders/... and /var is a symlink to /private/var, so without canonicalization the manager cannot identify itself through an ordinary absolute path whose parent is a link. Canonicalization is therefore load-bearing on macOS independently of any package manager, not only for the Windows junction. Log: TASK-260822-27bvo4_darwin-mutant-no-canonicalization.log.

GATES (real exit codes, each standalone, no pipes): gofmt -l ./cmd ./internal 0; go build ./... 0; go vet ./... 0; go test ./internal/godriver/ -count=1 0 (61.3s) and 0 again after mutate/restore (108.6s); the six identity tests -v all PASS with 0 skips; golangci-lint run 0 (0 issues); GOOS=windows go vet ./internal/godriver/ 0; GOOS=linux go vet ./internal/godriver/ 0.

NOT RUN: go test ./... - host data volume sat at 100% (1.0 GiB free) with two concurrent full-suite spawns, and a concurrent process purged the shared GOCACHE mid-link; three mutation attempts died in the linker (no space left on device, then twice could not import io from a vanished cache object) and only completed once moved to a private per-run GOCACHE, deleted afterwards. Windows execution not run here - RUN-260822-507162 measured it natively (go test ./internal/godriver/ exit 0 in 93.7s; pre-fix junction test exit 1) and those logs are attached. Conformance gates not run, CURATOR_CONFORMANCE_ROOT unset.

INTEGRITY: product code was mutated in place with an unconditional byte-backup restore around every attempt; the tree re-diffs to the identical 3-file/70-insertion diff and the attached patch artifact is byte-identical to the working tree.

FOR THE REVIEWER - repo geometry: internal/godriver does not exist on the main checkout branch (handoff/cocoaskills-parity-20260731 at 74fe162), which is why make ledger-check is red there on rows naming it. It exists on main (0d9fe6e), still carrying the unfixed filepath.EvalSymlinks, so nothing has landed yet. Base 6a9b201 is one board-metadata commit behind main.

Artifacts: TASK-260822-27bvo4_verification-RUN-50e157.md, TASK-260822-27bvo4_darwin-mutant-no-canonicalization.log, TASK-260822-27bvo4_symlink-launcher.patch.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-50e157, pid=14866, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [reviewer] reviewer (claude) (run=RUN-260822-1f3d1d, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260822-1f3d1d)
REVIEW VERDICT (RUN-260822-1f3d1d, reviewer): changes requested -> to-dev. The product fix is correct and every claim in the implementer notes was reproduced independently, not read back. Two CI-gate wiring gaps block acceptance; neither touches the fix or the test set.

VERIFIED INDEPENDENTLY. Gates re-run by me in .temp/TASK-260822-27bvo4/RUN-260822-507162 (real exit codes): gofmt -l ./cmd ./internal 0; go build ./... 0; go vet ./... 0; GOOS=windows go vet ./internal/godriver/ 0; GOOS=linux go vet ./internal/godriver/ 0; golangci-lint run 0 (0 issues); go test ./internal/godriver/ -count=1 0 (77.1s); the 5 new darwin-visible identity tests -v 0 with 0 skips; make ledger-check 0 (63 rows).

NATIVE WINDOWS RED-TO-GREEN, reproduced by me on ssh win (go1.25.5 windows/amd64), tree shipped to C:\curator-ci\review27 and removed after: post-fix 7 identity tests incl. both junction tests exit 0 with 0 skips; pre-fix baseline (physicalPath reverted to EvalSymlinks on the host copy, restored, go build 0) both junction tests exit 1 with worker_identity_invalid: cannot canonicalize the manager executable and EvalSymlinks returning The system cannot find the path specified - exactly the inversion the task describes; full godriver suite post-fix exit 0 in 93.781s.

DARWIN MUTATION, throwaway worktree, restored and re-diffed identical: reverting only physicalPath -> EvalSymlinks leaves all six identity tests PASSING, so the darwin lane proves nothing about this change and the load-bearing proof is Windows-only - the implementer said so plainly and that is correct. Stripping canonicalization entirely reddens all six, so canonicalization as such is load-bearing on darwin too. os.Executable was observed reporting the launch path rather than the physical file on BOTH darwin and windows.

FIX ASSESSMENT: identity.go:56 physicalPath(absolute) puts identity on the same resolver as selectToolchain/verifySelectedRoot/mustPhysical instead of keeping its own rule. Resolution-first does not widen trust: the substitution battery is untouched and now applies to the physical file, and on Windows artifactHasMultipleLinks opens the resolved path with FILE_FLAG_OPEN_REPARSE_POINT and still rejects reparse points and NumberOfLinks != 1. Worker proves identity through the same function (workerserver.go:88), so both sides canonicalize identically.

FINDING 1 (must fix) - unclassified skip reason, CI-fatal on a legitimate host condition. internal/godriver/identity_test.go:24 skips with "this account cannot create a file symbolic link". platform-case-gate.sh TIER 2 classifies every skip in the run against skip-classes.tsv and its own text says an unrecognised reason is fatal: add it to skip-classes.tsv with a class, or fix the case. The change did neither. I matched all three new skip reasons against every regex in the table the way the gate does: the junction and hard-link reasons hit host-capability :: this host cannot create; this one matches NOTHING. internal/godriver runs on the darwin and windows runners, so on any Windows host without SeCreateSymbolicLinkPrivilege this skips four subtests of TestExecutableIdentityResolvesALauncherLink plus five of TestExecutableIdentityRejectsSubstitutionThroughALauncherLink and turns CI red for the wrong reason instead of degrading the way the branch intends. It did not fire on the Windows host I used, which is why it would land unnoticed. The repo already has the classified phrasing for the identical situation: internal/buildcache/protection_windows_test.go:272 uses "creating Windows symlink requires host support: %v". Reuse that wording or add a skip-classes.tsv row.

FINDING 2 (should fix) - no platform-cases.tsv rows for the new identity cases. The fix is a Windows-only behaviour change (proved above), yet no ledger row requires any runner to execute the cases that prove it. A rename or deletion of the junction case is invisible to CI, and its own t.Skip classifies as host-capability/allow, so a runner that stops producing junctions silently drops the only case catching a regression of this fix - all while green. In-convention, not a new ask: platform-cases.tsv:107 already carries the sibling TestExecutableIdentityRejectsSubstitutionAndMutation darwin,windows. Four suggested rows are in the review artifact and are ledger-consistency-clean (the junction cases live in a //go:build windows file). Widening row 107 instead is fine; the point is that some runner must be REQUIRED to execute the Windows case.

NOT ASKED FOR: the test set is good - four launch shapes resolving to one identity, nine substitution shapes still failing closed, a real re-exec through a link, the full Build handshake with managerExecutable pointed at a link, and the worker process itself started from a link. identityProbeMode is in main_test.go so it is not in the shipped binary. The doc comment states the ordering rule and the junction reason accurately. Do not re-litigate the fix or the tests; just wire the two gate rows.

REVIEW WAS READ-ONLY: no file in the reviewed tree was modified; it still shows the same 3-file/70-insertion diff. Scratch worktree and the Windows host copy were removed.

Artifacts: TASK-260822-27bvo4_review-RUN-1f3d1d.md plus four review verification logs.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-1f3d1d, pid=79839, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260822-5ca017, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260822-5ca017)
agent completed: [implementer] developer (claude) (exit=1)
spawn run completed: claude (run=RUN-260822-5ca017, pid=68595, exit=1)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260822-185ddd, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260822-185ddd)
RUN-260822-185ddd (fourth spawn, rework after the RUN-1f3d1d changes-requested verdict). Both review findings are closed. NO PRODUCT CODE CHANGED by this run: identity.go still carries the same accepted one-line fix (physicalPath(absolute)) and the test set is unchanged. What was missing was the CI wiring that makes the new cases enforceable.

FINDING 1 CLOSED - identity_test.go:31 now skips with "creating Windows symlink requires host support: %v", the exact phrasing skip-classes.tsv already classifies as host-capability/allow and the phrasing internal/buildcache/protection_windows_test.go uses for the same situation. skip-classes.tsv itself is unchanged; no new row was needed. All four skip sites the new tests can reach were matched against the table the way platform-case-gate.sh matches them: the two hard-link ones and the junction one hit "this host cannot create", the symlink one hits the reused row. None fired on either runner below, so classification protects a host that lacks SeCreateSymbolicLinkPrivilege rather than propping up the green runs.

FINDING 2 CLOSED - nine rows added to platform-cases.tsv as a pure insertion after the existing TestExecutableIdentityRejectsSubstitutionAndMutation row (no existing row edited). Five cases required on darwin,windows with a tolerated host-capability skip on windows; the two junction cases required on windows (the ones the review asked for by name, since the fix is a Windows-only behaviour change); two Parent/* subtest rows that make a darwin subtest skip fatal while still tolerating the privilege-dependent Windows one. make ledger-check: 72 rows, exit 0 (was 63).

ROWS PROVED ENFORCEABLE, NOT MERELY CONSISTENT: platform-case-gate.sh was run against both real go test -json streams (CI_EXCLUDED_PKGS = every other ledger package, CI_DEFERRED_PKGS=internal/godriver, no conformance root here). Windows stream, CI_GATE_GOOS=windows: all seven new cases observed ok including both junction cases, zero skips among them. Darwin stream: the five darwin-required cases ok, zero skips among them. Both gate runs exit 1 on exactly one line, identical on both platforms and pre-existing: TestCandidateGoV1SourceAwareContract skips root-unset where its (untouched) row tolerates root-content, purely because CURATOR_CONFORMANCE_ROOT is unset locally. Reported red, not dressed up.

GATES (real exit codes, each standalone, no pipes). darwin/arm64 go1.25.5: gofmt -l ./cmd ./internal 0 (empty); go build ./... 0; go vet ./... 0; GOOS=windows go vet ./internal/godriver/ 0; GOOS=linux go vet ./internal/godriver/ 0; golangci-lint run 0 (0 issues); make ledger-check 0; gate-selftest.sh 0 (75 passed, 0 failed); no-broad-suppression.sh 0; go test -json ./internal/godriver/ -count=1 0. Native Windows host ssh win (DESKTOP-3PBO632, go1.25.5 windows/amd64), tree shipped to C:\curator-ci\dev185ddd and removed after: go build ./... 0; go test -json -count=1 ./internal/godriver/ 0 in 94.95s.

NOT RUN: go test ./... - host data volume at 99% (12 GiB free) and earlier runs on this task died with no space left on device at full-suite parallelism; go build ./... and go vet ./... cover the rest of the module. Conformance gates not run, CURATOR_CONFORMANCE_ROOT unset (that is the one gate failure above).

TREE: .temp/TASK-260822-27bvo4/RUN-260822-507162, branch task/TASK-260822-27bvo4-symlink-launcher-507162, base 6a9b201, uncommitted per standing orders. 4 files modified + 2 new test files, 95 insertions. The attached patch artifact was regenerated from this tree.

Artifacts: TASK-260822-27bvo4_gate-wiring-RUN-185ddd.md, TASK-260822-27bvo4_ledger-check-RUN-185ddd.log, TASK-260822-27bvo4_platform-case-gate-windows-RUN-185ddd.log, TASK-260822-27bvo4_platform-case-gate-darwin-RUN-185ddd.log, TASK-260822-27bvo4_skips-observed-windows-RUN-185ddd.tsv, TASK-260822-27bvo4_skips-observed-darwin-RUN-185ddd.tsv, and the refreshed TASK-260822-27bvo4_symlink-launcher.patch.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-185ddd, pid=43933, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [reviewer] reviewer (claude) (run=RUN-260822-abf447, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260822-abf447)
REVIEW VERDICT (RUN-260822-abf447, reviewer): ACCEPTED. Both RUN-1f3d1d findings closed. Every claim reproduced by me from a real run; no implementer log read back as evidence. Review was read-only - the tree still shows the same 4 modified files + 2 new test files, 95 insertions.

FINDING 1 CLOSED. identity_test.go:31 now skips with the exact phrasing skip-classes.tsv already classifies as host-capability/allow (creating Windows symlink requires host support), the same wording internal/buildcache/protection_windows_test.go uses; skip-classes.tsv itself unchanged and no new row needed. I matched all five skip reasons the new tests can reach - identity_test.go:31 and :175, identity_windows_test.go:39 and :115, and the pre-existing makeTestJunction at selection_windows_test.go:83 - against the table the way the gate matches them; all five classify host-capability, and none fired on either runner, so classification protects a privilege-less host rather than propping up a green run.

FINDING 2 CLOSED, AND THE ROWS BITE. Nine rows added as a pure insertion; make ledger-check exit 0 with 72 rows (was 63) and every new case name resolves against the per-GOOS build file sets. The rework sets skip_allowed_on=windows/host-capability where I had suggested -; that deviation is correct, not a shortcut: four of the five shared cases genuinely skip at the PARENT level on a Windows host without SeCreateSymbolicLinkPrivilege, so - would have re-created exactly the CI-fatal condition Finding 1 was about, and must=X/skip=X/host-capability is already the repo convention (two internal/buildcache reparse rows, one internal/scopes row). My ask was that some runner must be REQUIRED to execute the Windows case by name; that is satisfied.

ENFORCEABILITY PROVED, NOT ASSUMED. I drove platform-case-gate.sh against real go test -json streams from both runners plus two adversarial mutations of the darwin stream: darwin - all 5 darwin-visible cases ok, 0 skips among them; windows - all 7 cases ok including both junction cases, 0 skips among them; darwin stream with ResolvesALauncherLink renamed - FAIL required case never ran on darwin, by name; darwin stream with one subtest turned into a classified skip - FAIL ledger case skipped where the ledger does not tolerate it, which is the Parent/* rows doing their job; and the same windows stream against the BASE ledger from HEAD - identical single failure, so the new rows add no new red.

GATES (real exit codes, each standalone). darwin/arm64 go1.25.5: gofmt -l ./cmd ./internal 0 (empty); go vet ./internal/godriver/ 0; GOOS=windows go vet ./internal/godriver/ 0; golangci-lint run 0 (0 issues); go test -json ./internal/godriver/ -count=1 0 with 296 pass / 0 fail / 18 skip and zero skips among the identity cases; make ledger-check 0; gate-selftest.sh 0 (75 passed, 0 failed); no-broad-suppression.sh 0. Native Windows host ssh win (DESKTOP-3PBO632, go1.25.5 windows/amd64), tree shipped to C:\curator-ci\rev-abf447 and removed after: go build ./... 0; go test -json -count=1 ./internal/godriver/ 0 with 297 pass / 0 fail / 17 skip and both junction cases passing with zero skips.

THE ONE GATE RED, REPORTED AS RED: both platform-case-gate.sh runs exit 1 on exactly one line, identical on both platforms - TestCandidateGoV1SourceAwareContract skips root-unset where its untouched row tolerates root-content, because CURATOR_CONFORMANCE_ROOT is unset locally. Running the same windows stream against the base ledger reproduces the identical single failure, so it is pre-existing and independent of this change.

NOT RUN: go test ./... - host data volume at 99% (~10 GiB free) and earlier runs on this task died with no space left on device at full-suite parallelism. go build ./... and go vet ./... cover the rest of the module and the change touches only internal/godriver plus a CI table. Conformance gates not run, CURATOR_CONFORMANCE_ROOT unset.

FIX RE-CONFIRMED INTACT: identity.go:56 canonicalizes with physicalPath(absolute); on unix physicalPath is EvalSymlinks(Clean(path)) verbatim (platform_unix.go:16) so the change is bit-identical there, and on Windows it is GetFinalPathNameByHandle, which follows a junction where EvalSymlinks does not. The substitution battery in readExecutableIdentity is untouched and now applies to the physical file, and Verify() deliberately does NOT re-canonicalize - it re-Lstats the recorded physical path, which is why swapping that file for a link is still caught.

MINOR, NON-BLOCKING: identity_test.go:62 declares skipWin bool in the launch-shape table struct; no case sets it and nothing reads it - a dead field left over from merging the two implementation lanes. Lint does not flag it and it changes no behaviour. Worth a one-line sweep when the file is next touched, not another review cycle.

ACCEPTANCE EVIDENCE FOR THE COMMIT-OWNING MOVER: this reviewer run supplies no commit_ack. Tree .temp/TASK-260822-27bvo4/RUN-260822-507162, branch task/TASK-260822-27bvo4-symlink-launcher-507162, base 6a9b201, uncommitted per standing orders, matching the attached TASK-260822-27bvo4_symlink-launcher.patch. internal/godriver does not exist on the main checkout branch (handoff/cocoaskills-parity-20260731); it exists on main, still carrying the unfixed filepath.EvalSymlinks.

Artifacts: TASK-260822-27bvo4_review-RUN-abf447.md plus five review verification logs.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260822-abf447, pid=82944, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260822-27bvo4_spawn-log_-implementer--developer--claude-_RUN-260822-507162.log](file://TASK-260822-27bvo4/TASK-260822-27bvo4_spawn-log_-implementer--developer--claude-_RUN-260822-507162.log) — System spawn log captured by task-board
- [TASK-260822-27bvo4_spawn-log_-implementer--developer--claude-_RUN-260822-1e48d3.log](file://TASK-260822-27bvo4/TASK-260822-27bvo4_spawn-log_-implementer--developer--claude-_RUN-260822-1e48d3.log) — System spawn log captured by task-board
- [TASK-260822-27bvo4_symlink-launcher-canonicalization.md](file://TASK-260822-27bvo4/TASK-260822-27bvo4_symlink-launcher-canonicalization.md) — Finding, fix, tests, and red-to-green Windows evidence for manager executable identity canonicalization through a launcher link
- [TASK-260822-27bvo4_windows-identity-tests.log](file://TASK-260822-27bvo4/TASK-260822-27bvo4_windows-identity-tests.log) — windows/amd64 go test -run TestExecutableIdentity -v, exit 0 (post-fix)
- [TASK-260822-27bvo4_windows-prefix-baseline-red.log](file://TASK-260822-27bvo4/TASK-260822-27bvo4_windows-prefix-baseline-red.log) — windows/amd64 junction test against the pre-fix EvalSymlinks baseline, exit 1 (expected red)
- [TASK-260822-27bvo4_symlink-launcher-canonicalization-RUN-507162.md](file://TASK-260822-27bvo4/TASK-260822-27bvo4_symlink-launcher-canonicalization-RUN-507162.md) — Converged deliverable of the two duplicate spawn runs: physicalPath fix, launcher-link test set, darwin+native-Windows gate evidence including the pre-fix red baseline
- [TASK-260822-27bvo4_spawn-log_-implementer--developer--claude-_RUN-260822-50e157.log](file://TASK-260822-27bvo4/TASK-260822-27bvo4_spawn-log_-implementer--developer--claude-_RUN-260822-50e157.log) — System spawn log captured by task-board
- [TASK-260822-27bvo4_verification-RUN-50e157.md](file://TASK-260822-27bvo4/TASK-260822-27bvo4_verification-RUN-50e157.md) — RUN-50e157 independent re-verification of the merged tree: gate exit codes, darwin mutation check, AC mapping, and what was not run
- [TASK-260822-27bvo4_darwin-mutant-no-canonicalization.log](file://TASK-260822-27bvo4/TASK-260822-27bvo4_darwin-mutant-no-canonicalization.log) — go test -v with canonicalization stripped from resolveExecutableIdentity: all six identity tests red on darwin, exit 1
- [TASK-260822-27bvo4_symlink-launcher.patch](file://TASK-260822-27bvo4/TASK-260822-27bvo4_symlink-launcher.patch) — Full working-tree patch incl. the RUN-185ddd ledger rows and reworded skip
- [TASK-260822-27bvo4_spawn-log_-reviewer--reviewer--claude-_RUN-260822-1f3d1d.log](file://TASK-260822-27bvo4/TASK-260822-27bvo4_spawn-log_-reviewer--reviewer--claude-_RUN-260822-1f3d1d.log) — System spawn log captured by task-board
- [TASK-260822-27bvo4_review-RUN-1f3d1d.md](file://TASK-260822-27bvo4/TASK-260822-27bvo4_review-RUN-1f3d1d.md) — Reviewer verdict: fix verified independently on darwin and native Windows (red-to-green reproduced); changes requested on two CI-gate wiring gaps
- [TASK-260822-27bvo4_review-windows-prefix-red.log](file://TASK-260822-27bvo4/TASK-260822-27bvo4_review-windows-prefix-red.log) — Reviewer RUN-1f3d1d independent verification log
- [TASK-260822-27bvo4_review-windows-identity-postfix.log](file://TASK-260822-27bvo4/TASK-260822-27bvo4_review-windows-identity-postfix.log) — Reviewer RUN-1f3d1d independent verification log
- [TASK-260822-27bvo4_review-darwin-prefix-evalsymlinks.log](file://TASK-260822-27bvo4/TASK-260822-27bvo4_review-darwin-prefix-evalsymlinks.log) — Reviewer RUN-1f3d1d independent verification log
- [TASK-260822-27bvo4_review-darwin-mutant-no-canonical.log](file://TASK-260822-27bvo4/TASK-260822-27bvo4_review-darwin-mutant-no-canonical.log) — Reviewer RUN-1f3d1d independent verification log
- [TASK-260822-27bvo4_spawn-log_-implementer--developer--claude-_RUN-260822-5ca017.log](file://TASK-260822-27bvo4/TASK-260822-27bvo4_spawn-log_-implementer--developer--claude-_RUN-260822-5ca017.log) — System spawn log captured by task-board
- [TASK-260822-27bvo4_spawn-log_-implementer--developer--claude-_RUN-260822-185ddd.log](file://TASK-260822-27bvo4/TASK-260822-27bvo4_spawn-log_-implementer--developer--claude-_RUN-260822-185ddd.log) — System spawn log captured by task-board
- [TASK-260822-27bvo4_gate-wiring-RUN-185ddd.md](file://TASK-260822-27bvo4/TASK-260822-27bvo4_gate-wiring-RUN-185ddd.md) — RUN-185ddd gate wiring evidence
- [TASK-260822-27bvo4_ledger-check-RUN-185ddd.log](file://TASK-260822-27bvo4/TASK-260822-27bvo4_ledger-check-RUN-185ddd.log) — RUN-185ddd gate wiring evidence
- [TASK-260822-27bvo4_platform-case-gate-windows-RUN-185ddd.log](file://TASK-260822-27bvo4/TASK-260822-27bvo4_platform-case-gate-windows-RUN-185ddd.log) — RUN-185ddd gate wiring evidence
- [TASK-260822-27bvo4_platform-case-gate-darwin-RUN-185ddd.log](file://TASK-260822-27bvo4/TASK-260822-27bvo4_platform-case-gate-darwin-RUN-185ddd.log) — RUN-185ddd gate wiring evidence
- [TASK-260822-27bvo4_skips-observed-windows-RUN-185ddd.tsv](file://TASK-260822-27bvo4/TASK-260822-27bvo4_skips-observed-windows-RUN-185ddd.tsv) — RUN-185ddd gate wiring evidence
- [TASK-260822-27bvo4_skips-observed-darwin-RUN-185ddd.tsv](file://TASK-260822-27bvo4/TASK-260822-27bvo4_skips-observed-darwin-RUN-185ddd.tsv) — RUN-185ddd gate wiring evidence
- [TASK-260822-27bvo4_spawn-log_-reviewer--reviewer--claude-_RUN-260822-abf447.log](file://TASK-260822-27bvo4/TASK-260822-27bvo4_spawn-log_-reviewer--reviewer--claude-_RUN-260822-abf447.log) — System spawn log captured by task-board
- [TASK-260822-27bvo4_review-RUN-abf447.md](file://TASK-260822-27bvo4/TASK-260822-27bvo4_review-RUN-abf447.md) — Reviewer verdict (accepted): both gate-wiring findings closed, rows proved enforceable on darwin and native Windows
- [TASK-260822-27bvo4_review-abf447-gate-darwin.log](file://TASK-260822-27bvo4/TASK-260822-27bvo4_review-abf447-gate-darwin.log) — Reviewer RUN-abf447 verification log: gate-darwin.log
- [TASK-260822-27bvo4_review-abf447-gate-windows.log](file://TASK-260822-27bvo4/TASK-260822-27bvo4_review-abf447-gate-windows.log) — Reviewer RUN-abf447 verification log: gate-windows.log
- [TASK-260822-27bvo4_review-abf447-gate-darwin-renamed.log](file://TASK-260822-27bvo4/TASK-260822-27bvo4_review-abf447-gate-darwin-renamed.log) — Reviewer RUN-abf447 verification log: gate-darwin-renamed.log
- [TASK-260822-27bvo4_review-abf447-gate-darwin-subskip.log](file://TASK-260822-27bvo4/TASK-260822-27bvo4_review-abf447-gate-darwin-subskip.log) — Reviewer RUN-abf447 verification log: gate-darwin-subskip.log
- [TASK-260822-27bvo4_review-abf447-ledger-check.log](file://TASK-260822-27bvo4/TASK-260822-27bvo4_review-abf447-ledger-check.log) — Reviewer RUN-abf447 verification log: ledger-check.log

## Created
2026-08-22T16:12:28Z

## Last Update
2026-08-22T19:35:25Z

## Assigned To
[reviewer] reviewer (claude)
