## Status
reviewing

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(8))

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] Download and classify the five masked Windows failures from the gate evidence artifact.
- [x] Fix buildsource install and atomicity Windows behavior without skips or ledger weakening.
- [x] Add focused Windows regression tests and preserve macOS/Linux behavior.
- [x] Publish signed Curator PR targeting main with native Windows CI evidence.
- [x] Attach outcome evidence and hand off to independent Opus review.
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [ ] Implementation matches AC
- [ ] Solution fits project architecture
- [ ] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
Execution directive: Claude Opus 5 only. Work in an isolated branch/worktree from current relux-works/curator main (PR 9 is merged). Do not modify internal/runtimestore (BUG-260731-11bpa4) or the Linux inventory files owned by PR 11. PR 10 run 30619686990 artifact is the first Windows evidence; reproduce on windows-latest or ssh win, fix real behavior without skips, publish signed PR to main, then hand off to independent Opus review. No tags or releases.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260731-6c8de0, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260731-6c8de0)
SCOPE CORRECTION from solution-architect on BUG-260731-11bpa4 (recorded as a note, not a description rewrite, because this bug is already in development — the owner should re-estimate deliberately).

This bug description names five failing cases and "five packages beyond internal/runtimestore". Real CI evidence disagrees materially. Source: artifact test-evidence-windows-latest id 8789366268 from run 30620739038 job Test (windows-latest) on PR 10, parsed from test/go-test.json.

Actual failing top-level cases in THIS bug three packages: internal/install 60 (the description names only TestEndToEndInstall), internal/install/atomicity 8, internal/buildsource 2. That is 70, not 5 — roughly a 14x underestimate against the recorded scope. Repo-wide there are 91 failing cases across 11 packages, i.e. ten packages beyond internal/runtimestore, not five.

Estimate of 5 points was set against the understated figure and is very likely too low. Re-estimate before committing to a delivery expectation.

Ownership map for the rest: internal/runtimestore 1 -> BUG-260731-11bpa4; internal/transaction 5 + internal/godriver 1 -> BUG-260731-lepevi; cmd/curator 8 + internal/managerlock 2 + internal/buildcache 2 + internal/staging 1 + internal/globalbins 1 -> newly created BUG-260731-33v6zz, which previously had no owner. Full map attached to BUG-260731-11bpa4 as BUG-260731-11bpa4_windows-ownership-map.md.
Classified the five masked Windows packages from the gate evidence artifact (run 30620739038, test-evidence-windows-latest, .temp/ci-evidence/test/go-test.json).

Root causes, in-scope packages:
1. internal/install (60) + internal/install/atomicity (8): every fixture pinned Platform: "unix". On a Windows host a unix script runtime is validated for a POSIX execute bit no file carries (runtimestore/scripts.go:172) and a native compiled artifact is refused against a unix shim (runtimestore/targets.go:96). 69 of 70 failures are one error string: "validate staged script runtime: script command is not executable" / "compiled artifact target OS does not match shim platform".
2. internal/identity (product defect, surfaced by two internal/install dry-run cases): identity.Parse sent C:\repos\skill to the scp fallback, which forbids a backslash in the path part, so the drive colon read as a host separator and the local checkout came back "malformed or unsupported network source". curator could not install from a local path on Windows at all. Canonical hid it by discarding the error.
3. internal/buildsource (2): bad:name opens an NTFS alternate data stream so the entry that landed was the portable name "bad"; and a frozen root cannot be renamed on Windows because os.OpenRoot takes no FILE_SHARE_DELETE.
4. internal/install generation rename case: Windows holds an open declaration document against replacement.

Fixes published as GitHub-signed commit 2a02da9 on task/BUG-260731-27h1yc-windows-lane, PR https://github.com/relux-works/curator/pull/12 targeting main. Branch is stacked on PR 10 (8a68692) because without its go vet fix the Windows job aborts before the test gate and no Windows evidence is producible.

Ledger untouched: .github/ci/platform-cases.tsv and .github/ci/skip-classes.tsv unchanged; no case deleted, skipped or platform-excluded. TestGlobalInstall and the global rollback sweep now assert on Windows a class they previously skipped there.

macOS local: internal/install, internal/install/atomicity, internal/buildsource, internal/identity, internal/runtimestore, internal/globalbins all ok. Awaiting native windows-latest evidence.
RESULT — acceptance criterion met on a native windows-latest runner.

Test (windows-latest) run 30626331508 (head c7bc890), artifact test-evidence-windows-latest:
  internal/buildsource        2 failing on main -> PASS (0 failures)
  internal/install           60 failing on main -> PASS (0 failures)
  internal/install/atomicity  8 failing on main -> PASS (0 failures)
Baseline measured on main @ 3a047d5, run 30624569953: 91 failures repo-wide, 70 of them in these three packages.

All eight required platform-cases.tsv rows for these packages report ok in the gate report:
  buildsource TestFrozenTokenRejectsMutation / TestFrozenTokenRejectsRootReplacement / TestWithValidatedOrdersCallbackAndRejectsMutation
  install TestDryRunTouchesNothing / TestEndToEndInstall
  install/atomicity TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder / TestStaleAdapterRemovalRollsBackToTheExactPriorEntry / TestAdapterMirrorLinksAreJournaledAndRestoredExactly

Ledger: platform-cases.tsv and skip-classes.tsv byte identical to main. skips-observed.tsv on the Windows runner identical to the main baseline (only a t.TempDir nonce differs, in an unrelated internal/scopes row). Strengthened, not relaxed: TestGlobalInstall and the global rollback sweep now assert the PATH-visible forwarding shim and user-bin mirror ledger on Windows, which they previously bypassed behind a runtime.GOOS != "windows" guard.

Job still red overall on 19 cases owned elsewhere: cmd/curator, internal/managerlock, internal/buildcache, internal/staging, internal/globalbins (BUG-260731-33v6zz) and internal/transaction (BUG-260731-lepevi). Predicted by the ownership map; out of this bug scope by construction.

Other lanes green in the same run: Test ubuntu, Test macos, Race ubuntu, Race macos, Lint, Naming gate, Interop conformance gate, all three Gate self-test jobs.

Deviation from the recorded directive: branched from PR 10 head 8a68692 rather than main, because on main go vet still aborted the Windows job before the test gate and no Windows evidence was producible. PR 10 has since merged, main was merged in (a164dca), and the PR diff against main is now this task alone. internal/runtimestore and the Linux inventory files were never touched.

One product change beyond the named packages: internal/identity/identity.go. identity.Parse refused every local Windows checkout (C:\repos\skill reached the scp fallback, which forbids a backslash in the path part, so the drive colon read as a host separator). Two internal/install dry-run cases fail inside it. Real user-facing defect: curator could not install from a local path on Windows at all.

Artifacts: BUG-260731-27h1yc_results.md, BUG-260731-27h1yc_windows-platform-case-gate.txt, BUG-260731-27h1yc_windows-skips-observed.tsv. Logbook entry 2026-07-31 1537.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260731-6c8de0, pid=21348, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [reviewer] reviewer (claude) (run=RUN-260731-5000ae, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260731-5000ae)
Migration checkpoint 2026-07-31: reviewer RUN-260731-5000ae was intentionally cancelled before host transfer, not a review verdict. On ivan-macbook-air-m1, inspect PR 12 current CI/evidence and spawn a fresh independent Claude Opus 5 reviewer.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [reviewer] reviewer (claude) (run=RUN-260731-a902cb, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260731-a902cb)

## Precondition Resources
(none)

## Outcome Resources
- [BUG-260731-27h1yc_spawn-log_-implementer--developer--claude-_RUN-260731-6c8de0.log](file://BUG-260731-27h1yc/BUG-260731-27h1yc_spawn-log_-implementer--developer--claude-_RUN-260731-6c8de0.log) — System spawn log captured by task-board
- [BUG-260731-27h1yc_results.md](file://BUG-260731-27h1yc/BUG-260731-27h1yc_results.md) — Windows lane fix: root causes, baseline vs after evidence from native windows-latest runs, ledger parity proof
- [BUG-260731-27h1yc_windows-platform-case-gate.txt](file://BUG-260731-27h1yc/BUG-260731-27h1yc_windows-platform-case-gate.txt) — platform-case gate report from Test (windows-latest) run 30626331508: all 8 required ledger rows for buildsource/install/install-atomicity report ok
- [BUG-260731-27h1yc_windows-skips-observed.tsv](file://BUG-260731-27h1yc/BUG-260731-27h1yc_windows-skips-observed.tsv) — skips-observed.tsv from the same Windows run; identical to the main baseline, proving no skip was added
- [BUG-260731-27h1yc_spawn-log_-reviewer--reviewer--claude-_RUN-260731-5000ae.log](file://BUG-260731-27h1yc/BUG-260731-27h1yc_spawn-log_-reviewer--reviewer--claude-_RUN-260731-5000ae.log) — System spawn log captured by task-board
- [BUG-260731-27h1yc_spawn-log_-reviewer--reviewer--claude-_RUN-260731-a902cb.log](file://BUG-260731-27h1yc/BUG-260731-27h1yc_spawn-log_-reviewer--reviewer--claude-_RUN-260731-a902cb.log) — System spawn log captured by task-board

## Created
2026-07-31T09:29:57Z

## Last Update
2026-07-31T13:22:52Z

## Assigned To
[reviewer] reviewer (claude)
