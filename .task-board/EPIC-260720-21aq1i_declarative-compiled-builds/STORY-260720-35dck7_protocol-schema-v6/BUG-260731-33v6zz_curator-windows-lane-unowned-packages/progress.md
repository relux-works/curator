## Status
development

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(8))

## Blocked By
- (none)

## Blocks
- BUG-260731-3a5q1p

## Checklist
- [x] Branch from the current PR 10 head and classify all 14 Windows failures from raw go test JSON.
- [ ] Fix cmd curator buildcache globalbins managerlock and staging Windows behavior without skips.
- [x] Add focused Windows regression coverage and preserve Linux and macOS behavior.
- [x] Publish a signed Curator PR targeting main with native windows-latest evidence.
- [ ] Attach outcome evidence and hand off to independent Opus review.
- [ ] Code written per task description and AC
- [ ] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant

## Notes
Blocked-by rationale: main is 2b6ef21 (PR 9 merged) and still carries the internal/runtimestore vet break, so Test (windows-latest) aborts at go vet before any test runs and this bug can get no Windows signal from a branch off main. Either wait for BUG-260731-11bpa4 PR 10 to land, or branch off PR 10 head as BUG-260731-27h1yc did.

Verification note: none of the 14 failing cases is in .github/ci/platform-cases.tsv, so ledger-consistency.sh and platform-case-gate.sh will not flag them. Use the raw go test -json evidence artifact, not the gate summary — the job log prints only stage exit codes. Reproduce on a native windows-latest runner; a macOS or Linux host cannot execute these paths.
Execution directive: Claude Opus 5 only. Start from published PR 10 head 8a686920fc05c92b7e5b8d8bfd9d3de39d6e5a98 in an isolated worktree so Windows tests compile. Keep ownership boundaries exactly as scope states; do not touch runtimestore, buildsource, install, atomicity, godriver or transaction. Publish signed PR to relux-works/curator main, no tags/releases.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260731-6b26c6, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260731-6b26c6)
Worktree .temp/BUG-260731-33v6zz/worktree on task/BUG-260731-33v6zz-windows-lane, based on PR10 head 8a68692. Round 1 pushed (eca32f2, signed). Root causes classified from run 30620739038 windows artifact:
- buildcache x2: ensureProtectedBase create-then-protect TOCTOU on Windows -> concurrent publisher sees an inherited DACL. Fixed by creating the directory with the owner-only protected SD attached (windows.CreateDirectory + SECURITY_ATTRIBUTES, SE_DACL_PROTECTED).
- managerlock x2: portable cases asserted the on-disk spelling as identity; Windows folds case-insensitive components by design (identity_windows.go, asserted by identity_windows_test.go). Expectations now spelled through the platform rule.
- staging x1: POSIX path literals are neither absolute nor clean on Windows. Fixtures now built from the host root.
- globalbins x1: unix runtime profile pinned on a host with no POSIX execute bit. Now uses runtimestore.Platform().
- cmd/curator x1 (upgrade case): internal/identity rejected C:\... as a malformed network source. Fix taken byte-identical from BUG-260731-27h1yc so both PRs merge cleanly; the identity unit coverage stays in that PR.
- cmd/curator x7: go-v1 refuses the host GOROOT with go_toolchain_missing "trusted GOROOT is not a real directory". Not yet diagnosed - only cmd/curator exercises the real build.Default.GOROOT path (install uses fakeToolchain, godriver injects hostFacts), and it passes on macOS. Round 1 carries TestHostGoToolchainIsSelectableOnAnInventoryPlatform which prints the selection inputs so the Windows runner names the rejected host fact.
CI dispatched: run 30623699047.
Round 6 in flight (run 30627771028). Native windows-latest progress from raw go test -json: 14 unowned failures at base -> 4 after round 5, all in cmd/curator; buildcache/managerlock/globalbins/staging green since round 2. macOS, Linux, Lint and Race all green on this branch since the rebase onto main (PR 10 and PR 11 merged).

Four peeled layers, each a real Windows defect:
1. internal/identity rejected C:\... as a malformed network source (hunk byte-identical to BUG-260731-27h1yc so both PRs merge cleanly).
2. go-v1 refused the host GOROOT. Proven on the runner: the Actions tool cache publishes Go behind a DIRECTORY JUNCTION (attributes=0x410, reparseTag=0xa0000003). Since Go 1.23 os.Lstat calls a junction ModeIrregular and filepath.EvalSymlinks follows only ModeSymlink, so it left it unresolved; EvalSymlinks also fails ENOTDIR for a junction mid-path. Fixed by resolving root and launcher through GetFinalPathNameByHandle. Anchoring on the physical target is stricter than before, not looser.
3. buildcache refused the manager home itself. Nothing in Curator ever attaches a protected DACL to a home made by os.MkdirAll, so NO publication on Windows could ever succeed; the buildcache suite only passed because its fixture calls protectTestHome. ensureProtectedBase now hardens an owned, merely-inheriting ancestor; other-principal directories are still refused.
4. That hardening then denied the owner access to its own new files, because the owner-only ACE was non-inheritable and the home is the parent of manager locks, transaction files and the global root. A container ACE now propagates, a file ACE does not, and validateWindowsDACL accepts propagation flags while still rejecting INHERITED_ACE.

OWNERSHIP: internal/godriver was excluded because BUG-260731-lepevi was concurrent. PR 11 has merged, this branch is rebased onto it, and PR 11 touched controls_other.go not session.go, so nothing collides. Flagged for review.
NOT MINE, still red: internal/godriver TestFingerprintReportsUnreadableDirectoryIdentically failed at base, was mapped to BUG-260731-lepevi, and survived that PR - it is unowned again.
PR: https://github.com/relux-works/curator/pull/13
Round 8 in flight (run 30630564858). Native windows-latest trajectory from raw go test -json: 14 -> 8 -> 4 -> 3 -> 1 failing case. buildcache, managerlock, globalbins and staging green since round 2; cmd/curator down to one test (TestCompiledProjectStatusRepairRollbackRecovery) whose last three subtests were fixture-level Windows gaps, now fixed. macOS, Linux, Lint, Race all green on this branch.

Six stacked root causes, each a real Windows defect or a Unix-shaped expectation:
1. buildcache create-then-protect TOCTOU -> atomic CreateDirectory with SE_DACL_PROTECTED.
2. identity rejected C:\... as a malformed network source (byte-identical to PR 12).
3. managerlock/staging/globalbins Unix-shaped expectations.
4. go-v1 refused the host GOROOT: the Actions tool cache publishes Go behind a DIRECTORY JUNCTION (reparseTag=0xa0000003). Since Go 1.23 os.Lstat calls a junction ModeIrregular and EvalSymlinks follows only ModeSymlink, so it stayed unresolved; EvalSymlinks also fails ENOTDIR for a junction mid-path. Fixed via GetFinalPathNameByHandle; anchoring on the physical target is stricter than before.
5. NO manager home was ever protected, so no publication on Windows could ever succeed; buildcache passed only because its fixture calls protectTestHome. ensureProtectedBase now hardens an owned, merely-inheriting ancestor.
6. That hardening then denied the owner its own new files - a non-inheritable ACE on a container gives children an empty DACL. Container ACEs now propagate; validateWindowsDACL accepts propagation flags but still rejects INHERITED_ACE.

Ledgers byte unchanged: git diff main -- .github/ is empty.
OWNERSHIP: internal/godriver was excluded because BUG-260731-lepevi was concurrent; PR 11 has MERGED, branch rebased onto it, and PR 11 touched controls_other.go not session.go, so no collision. Flagged in the PR body.
NOT MINE, unowned again: internal/godriver TestFingerprintReportsUnreadableDirectoryIdentically and internal/transaction (5) were mapped to BUG-260731-lepevi and survived PR 11, whose AC covered only the Linux lane.
PR: https://github.com/relux-works/curator/pull/13
Migration checkpoint 2026-07-31: developer RUN-260731-6b26c6 was intentionally cancelled after publishing PR 13. Resume by checking run 30630564858 and PR head, then route remaining failures/rework or independent Opus review. Do not use alternative branch task/BUG-260731-fs3dht-windows-goroot unless comparison is needed.
Host migration resumed on macbook-iv.local. Dependency BUG-260731-fs3dht is closed as superseded because PR 13 already contains the integrated GOROOT solution. Inspect the latest Windows artifact from run 30630564858, fix or classify every remaining red case, push only to PR 13 branch, then hand off to independent Opus review.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260731-1b588e, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260731-1b588e)

## Precondition Resources
(none)

## Outcome Resources
- [BUG-260731-33v6zz_spawn-log_-implementer--developer--claude-_RUN-260731-6b26c6.log](file://BUG-260731-33v6zz/BUG-260731-33v6zz_spawn-log_-implementer--developer--claude-_RUN-260731-6b26c6.log) — System spawn log captured by task-board
- [BUG-260731-33v6zz_results.md](file://BUG-260731-33v6zz/BUG-260731-33v6zz_results.md) — Windows lane fix: root causes, per-round native windows-latest evidence, verification exit codes, ownership notes
- [BUG-260731-33v6zz_spawn-log_-implementer--developer--claude-_RUN-260731-1b588e.log](file://BUG-260731-33v6zz/BUG-260731-33v6zz_spawn-log_-implementer--developer--claude-_RUN-260731-1b588e.log) — System spawn log captured by task-board

## Created
2026-07-31T09:59:06Z

## Last Update
2026-07-31T13:41:18Z

## Assigned To
[implementer] developer (claude)
