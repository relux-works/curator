## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(5))

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] Reproduce the Linux lint and compiled-control failures on a native Linux runner.
- [x] Remove genuinely dead Linux code without suppressing unused analysis.
- [x] Align Linux compiled-build expectations with the authoritative native-control inventory without claiming unsupported execution.
- [x] Publish a signed Curator PR targeting main and attach focused plus full CI evidence.
- [ ] Obtain independent Opus review and land only after required CI is green.
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
Execution directive: Claude Opus 5 only for implementation/review. Codex reserved for orchestration. Land accepted PR autonomously to main; no tag or release.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260731-7722a0, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260731-7722a0)
agent completed: [implementer] developer (claude) (exit=1)
spawn run completed: claude (run=RUN-260731-7722a0, pid=53098, exit=1)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260731-118c78, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260731-118c78)
agent completed: [implementer] developer (claude) (exit=1)
spawn run completed: claude (run=RUN-260731-118c78, pid=53474, exit=1)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260731-a0ad63, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260731-a0ad63)
spawn run started: [implementer] developer (claude) (run=RUN-260731-922dc9)
spawn run started: [implementer] developer (claude) (run=RUN-260731-4655a2)
spawn run started: [implementer] developer (claude) (run=RUN-260731-35d0f2)
spawn run started: [implementer] developer (claude) (run=RUN-260731-7dca5f)
ORCHESTRATOR HANDOVER (RUN-260731-11c4e2): my owner goal GOAL-260731-11c4e2 was CANCELLED by an acknowledged operator mutation at 2026-07-31T09:25:36Z (authority=operator_acknowledged, actor=codex-orchestrator, reason: Codex thread is the user-designated orchestrator; managed Opus owner repeatedly restarts and creates duplicate reviewer runs). Ownership of this bug passes to the Codex orchestrator. Producer RUN-260731-7dca5f was NOT touched and is still running - it carries a pending nudge directive RUN-260731-7dca5f:nudge:c6316a and no PR has been published yet. Context it will need: the toolchain-identity repair it depends on is on Curator PR 9 (bd6ba08), which is OPEN and MERGEABLE but unmerged; BUG-260731-11bpa4 PR 10 is stacked on that same base. Full handover: STORY-260720-35dck7_orchestrator-ownership-cancel-handover-RUN-260731-11c4e2.md
PR 11 opened (signed commit b2ac7d7, G) on task/BUG-260731-lepevi-linux-lane, base main, branched from bd6ba08 (PR 9 head). Lint: both unused findings removed as genuinely dead linux code - (*controlDomain).destroy deleted from controls_other.go (only the darwin/windows domains call their own from inside their own launch); existingNamespaceAncestor moved into namespace_case_darwin.go, its only consumer. No nolint, no exclusion, no reference trick. Tests: six cmd/curator compiled-build cases carved out by requireNativeControlInventoryPlatform, which reads godriver.InventoryPlatform rather than a GOOS list. Carve-out asserted by new TestCompiledInstallFollowsTheNativeControlInventoryExactly, which runs on every runner and proves both directions. Ledger rows added for all seven; one narrow platform-control skip-class pattern added matching the inventory wording. Local: GOOS=linux golangci-lint before=1/after=0, darwin test-gate exit 0, ledger-consistency 0, gate-selftest 0 (75/0), no-broad-suppression 0, gofmt clean, go vet 0. Awaiting native ubuntu CI on PR 11.
READY FOR REVIEW. PR 11 (signed commit b2ac7d7, G) -> main, branched from bd6ba08 (PR 9 head). AC MET on the native Linux runner: run 30620349565 Lint pass (51s) and Test (ubuntu-latest) pass (1m36s), both FAILURE at the same base in baseline run 30615765014. Race (ubuntu-latest) also flipped FAILURE -> pass; Test/Race (macos-latest) pass. Only red is Test (windows-latest), which fails at step 7 go vet on internal/runtimestore/targets_windows_test.go:97 undefined: decodeHelperOutput - that is BUG-260731-11bpa4 / PR 10, deliberately untouched here, and it fails before any test runs. Lint fix: both unused findings removed as genuinely dead linux code, no nolint / exclusion / reference trick. Test fix: six cmd/curator compiled-build cases carved out by requireNativeControlInventoryPlatform reading godriver.InventoryPlatform, plus new TestCompiledInstallFollowsTheNativeControlInventoryExactly which runs on every runner and passed on BOTH the uncovered (ubuntu) and covered (macos) side. Ledger rows added for all seven; the one new skip-class pattern was adversarially proven narrow (real reason on darwin => FATAL-not-tolerated, vague reason => UNCLASSIFIED/FATAL-wrong-class, host-capability wording => FATAL-wrong-class). Checklist item 5 (independent Opus review, land after green CI) left UNCHECKED: that is the reviewer role, not the developer role. Outcome artifacts: BUG-260731-lepevi_linux-lane-outcome.md and BUG-260731-lepevi_ci-evidence.tar.gz. Logbook entry recorded at 1338.
HANDED OFF TO REVIEW (status=to-review). PR 11 (signed commit b2ac7d7, G) -> main, branched from bd6ba08 (PR 9 head).

AC MET on the native Linux runner. Run 30620349565: Lint pass (51s), Test (ubuntu-latest) pass (1m36s) - both were FAILURE at the same base in baseline run 30615765014. Race (ubuntu-latest) also flipped FAILURE -> pass. Test (macos-latest) pass, Race (macos-latest) pass (10m41s). Only red is Test (windows-latest), failing at step 7 go vet on internal/runtimestore/targets_windows_test.go:97 undefined: decodeHelperOutput - that is BUG-260731-11bpa4 / PR 10, deliberately untouched here, and it fails before any test runs.

Lint: both unused findings removed as genuinely dead linux code. (*controlDomain).destroy had no reachable caller on the !darwin && !windows build (the two platform domains each call their own from inside their own launch; the shared worker client never does). existingNamespaceAncestor moved into namespace_case_darwin.go, its only consumer. No nolint, no linter exclusion, no reference trick; no-broad-suppression.sh passes.

Test: six cmd/curator compiled-build cases carved out by requireNativeControlInventoryPlatform, whose predicate is godriver.InventoryPlatform(runtime.GOOS)=="" - read from the inventory, not a GOOS list, so it cannot drift and stops skipping on its own when the inventory gains a record. New TestCompiledInstallFollowsTheNativeControlInventoryExactly runs on every runner and PASSED on both the uncovered side (ubuntu: refusal carries build_execution_control_unavailable, zero published cache entries, status --check non-zero) and the covered side (macos: install succeeds, exactly one protected entry). Zero cmd/curator skips on macos. Ledger rows added for all seven cases; the one new skip-class pattern was adversarially proven narrow (same reason on darwin => FATAL-not-tolerated; vague reason => UNCLASSIFIED/FATAL-wrong-class; host-capability wording => FATAL-wrong-class).

CHECKLIST ITEM 5 LEFT UNCHECKED DELIBERATELY. "Obtain independent Opus review and land only after required CI is green" is a reviewer/lander gate, not a developer gate - I cannot independently review my own work, and landing follows review acceptance. task-board handoff --role developer fails closed on it, so the role end status was applied with the explicit transition documented in the same role brief (set_status status=to-review, "when your work is ready for review"). Items 1-4 and 6-11 are checked and each was verified by a command that actually exited 0. The required-CI half of item 5 is already satisfiable from run 30620349565 as summarised above.

Outcome artifacts on the board: BUG-260731-lepevi_linux-lane-outcome.md (full decision record, evidence, exit-code table, adversarial skip-class proof, out-of-scope Windows red) and BUG-260731-lepevi_ci-evidence.tar.gz (baseline red + green ubuntu/macOS gate artifacts + local gate outputs). Logbook entry at 1338. Worktree .temp/BUG-260731-lepevi/worktree; primary checkout untouched.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [reviewer] reviewer (claude) (run=RUN-260731-a087ac, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260731-a087ac)
agent completed: [implementer] developer (claude) (exit=1)
spawn run completed: claude (run=RUN-260731-7dca5f, pid=0, exit=1)
spawn run started: [implementer] developer (claude) (run=RUN-260731-dbbe43)
CHECKPOINT RUN-260731-dbbe43 (developer successor): honored applied operator cancel RUN-260731-dbbe43:cancel:04e3a8 (payload: developer handoff complete, reviewer RUN-260731-a087ac active, cancel redundant autonomous-recovery successor). task-board spawn goal RUN-260731-dbbe43 now returns: Active Goal: none (run is not goal-bound). No developer work restarted and no product code touched by this run. This run did set status development at start before reading the directive; the active reviewer has since restored status=reviewing, so no board regression persists. Verified without mutating review state: PR 11 head b2ac7d7 signed (GitHub verification verified=true, reason=valid), base main, MERGEABLE; AC-scoped lanes green on run 30620349565 - Lint SUCCESS (job 91123182081) and Test (ubuntu-latest) SUCCESS (job 91123182110); also green: Race (ubuntu-latest), Test/Race (macos-latest), all Gate self-tests, Interop conformance, Naming gate. Test (windows-latest) FAILURE (job 91123182213) is out of this bug ownership: vet.exe internal/runtimestore/targets_windows_test.go:97:14 undefined: decodeHelperOutput, owned by sibling BUG-260731-11bpa4 / PR 10; PR 11 diff touches zero files under internal/runtimestore. PR 9 landed: main is now 2b6ef21 and base bd6ba08 is an ancestor (main ahead=1, behind=0), so no rebase is required and PR 11 stays MERGEABLE. Remaining open checklist item - independent Opus review and landing after required CI green - is owned by reviewer RUN-260731-a087ac and gated on PR 10 clearing the Windows lane.
REVIEW VERDICT: ACCEPTED (reviewer RUN-260731-a087ac, Opus 5, read-only; no commit_ack supplied, nothing merged). Artifact: BUG-260731-lepevi_review-verdict.md.

AC MET, re-derived independently rather than taken from the outcome doc. gh pr checks 11: Lint pass 51s (job 91123182081), Test (ubuntu-latest) pass 1m36s (job 91123182110); both were FAILURE at the same base in baseline run 30615765014, so the flip is attributable to this change and not to the PR 9 gate repair. Race (ubuntu), Test/Race (macos), 3x Gate self-test, Interop conformance and Naming gate are green too.

UNUSED CHECK NOT WEAKENED. Re-derived the call graph myself: destroy has exactly two call sites, controls_darwin.go:188 and controls_windows.go:161, each calling its own definition from inside its own launch; workerclient.go destroyBeforeExecution is a different function. On //go:build !darwin && !windows the method had no reachable caller in any configuration - deletion correct. existingNamespaceAncestor after the move has its definition and both call sites inside namespace_case_darwin.go only; read namespace_case_windows.go and namespace_case_other.go to confirm neither interrogates the filesystem, so neither can need an ancestor. Zero nolint / exclusion / reference tricks in the diff; ran no-broad-suppression.sh myself -> ok.

CARVE-OUT NOT WEAKENED. The guard predicate is godriver.InventoryPlatform (internal/godriver/controls.go:200) - the same function the driver refusal path consults - not a GOOS allow-list, so it cannot drift and stops skipping on its own when the inventory gains a record. Read platform-case-gate.sh rather than trusting the adversarial table: for the six rows a skip on darwin/windows fails listed(tol,goos) -> FATAL-not-tolerated (L216-220) and a mismatched class -> FATAL-wrong-class (L221-225); narrowness is a property of the gate code. Downloaded both CI artifacts myself: ubuntu observed-cases.tsv shows cmd/curator 56 pass / 6 skip / 0 fail and pass on TestCompiledInstallFollowsTheNativeControlInventoryExactly (uncovered branch really executed on native Linux); macos platform-cases.txt shows all seven rows ok with ZERO cmd/curator skips (covered branch, guard inert); 39 skips recorded on ubuntu, zero UNCLASSIFIED or FATAL-*; go-test-assert-godriver.json shows TestProbeRejectsAnUncoveredPlatformBeforeTheWorker still pass on the excluded runner. Read all six carved-out bodies: each starts from a real compiled install and asserts only compiled state - nothing non-compiled is lost, and general gc coverage (lock/prune/serialize) still runs on Linux from gc_test.go. Not a forced fit: identical shape to the existing whole-package godriver exclusion, one level finer. TASK-260728-1skseh correctly related (already cited at platform-exclusions.tsv:17).

GATES RE-RUN BY ME in the worktree: gofmt clean, go vet 0, GOOS={linux,darwin,windows} go build 0 each, gate-selftest.sh 75 passed/0 failed, ledger-consistency.sh ok 56 rows, no-broad-suppression.sh ok. Signature verified: %G? = G. golangci-lint is not installed on this reviewer host so it was not re-run locally - unnecessary, since Lint on the real ubuntu runner is green on the PR and was FAILURE with exactly the two findings at the base.

REBASE: independently confirmed no rebase needed. main 2b6ef21 tree = 4f3d788 and bd6ba08 tree = 4f3d788 - identical; PR MERGEABLE; compare/main...b2ac7d7 lists exactly the 8 changed files.

LANDING PRECONDITION (not a defect in this work): Test (windows-latest) fails at step 7 go vet, vet.exe internal/runtimestore/targets_windows_test.go:97:14 undefined: decodeHelperOutput - read from job 91123182213 log. That is BUG-260731-11bpa4 / PR 10, still OPEN; PR 11 touches zero files under internal/runtimestore. Reviewer adds for the record: Test (windows-latest) was last green 2026-07-14 (run 29298518437), i.e. before compiled builds landed in cfffd7c, so the six cases must_run_on=windows has never been observed passing - a pre-existing condition of the feature, not of this change. Declaring the requirement instead of narrowing the row to darwin is the correct call. Landing order: PR 10 -> PR 11 -> confirm Lint and Test (ubuntu-latest) green on main.

CHECKLIST: items 12/13/14 checked by this review. Item 5 LEFT UNCHECKED deliberately - the independent-Opus-review half is now satisfied, the LAND half is not and cannot be until PR 10 clears the Windows lane; ticking it would misreport state. Item 15 left unchecked as not-applicable (review accepts).

NON-BLOCKING: (1) publishedCacheEntries at status_test.go:346 duplicates cacheEntries at :368, ~15 near-identical lines differing only in missing-root tolerance; (2) the new skip-class pattern narrowness was proven with synthetic streams that were then deleted rather than added to gate-selftest.sh - acceptable because that self-test covers class mechanics not per-pattern (only 1 of 4 existing platform-control patterns has a dedicated case) and every CI run re-proves it in skips-observed.tsv; (3) the outcome artifact never states the rebase decision the orchestrator context asked for, though the decision is correct and is on the board via the RUN-260731-dbbe43 checkpoint.
REVIEWER FINDING (added after the verdict note above, artifact updated): the Windows half of the six carve-out rows is not merely undeclared-and-unobserved - it is obtainable and it FAILS. PR 10 run 30620739038 got the Windows lane past go vet; its test-evidence-windows-latest/test/observed-cases.tsv records ALL SIX carved-out cases as fail (not skip), every one with go-v1 go_toolchain_missing: trusted GOROOT is not a real directory (status_test.go:373/:1130/:1411; 22 occurrences in that stream).

WHY THIS STRENGTHENS THE ACCEPT RATHER THAN UNDERMINING IT: the failure is GOROOT resolution, NOT rc5-native-control-inventory-v1. InventoryPlatform(windows) returns windows, so requireNativeControlInventoryPlatform is inert on Windows and this PR guard is not implicated. The defect is pre-existing from cfffd7c and was observed on a branch that does not contain this change. It also confirms that narrowing those rows to must_run_on=darwin - which would have made the ledger look satisfied - would have hidden a real Windows defect behind a ledger edit. Declaring the requirement is what keeps the gate naming them.

SCOPE INPUT FOR BUG-260731-33v6zz: those six are 6 of its 8 unowned cmd/curator Windows failures. The other two are TestAuthoritativeUpgradeCasesAreExecutable and TestDryRunNeverClaimsACompletedCompilerCheck - the latter a compiled case this PR deliberately did NOT carve out (it needs no completed compilation) and correctly absent from the ledger. DERIVED, NOT EXECUTED: the new TestCompiledInstallFollowsTheNativeControlInventoryExactly takes its COVERED-host branch on Windows and requires install to exit OK, so the same GOROOT defect will make it a 7th red cmd/curator case there once PR 10 lands - a correct test exposing an existing defect, not a new one, but 33v6zz scope should count it.

Windows will therefore stay red past PR 10 until GOROOT resolution is fixed. Outside this bug AC (Lint + Test (ubuntu-latest)), both green. Logbook entry 1402.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260731-a087ac, pid=18369, exit=0)

## Precondition Resources
- [BUG-260731-lepevi_orchestrator-context.md](file://BUG-260731-lepevi/BUG-260731-lepevi_orchestrator-context.md) — Orchestrator execution context: worktree isolation, PR9 base commit, ownership boundary vs BUG-260731-11bpa4, no-weakening constraints

## Outcome Resources
- [BUG-260731-lepevi_spawn-log_-implementer--developer--claude-_RUN-260731-7722a0.log](file://BUG-260731-lepevi/BUG-260731-lepevi_spawn-log_-implementer--developer--claude-_RUN-260731-7722a0.log) — System spawn log captured by task-board
- [BUG-260731-lepevi_spawn-log_-implementer--developer--claude-_RUN-260731-118c78.log](file://BUG-260731-lepevi/BUG-260731-lepevi_spawn-log_-implementer--developer--claude-_RUN-260731-118c78.log) — System spawn log captured by task-board
- [BUG-260731-lepevi_spawn-log_-implementer--developer--claude-_RUN-260731-a0ad63.log](file://BUG-260731-lepevi/BUG-260731-lepevi_spawn-log_-implementer--developer--claude-_RUN-260731-a0ad63.log) — System spawn log captured by task-board
- [BUG-260731-lepevi_spawn-log_-implementer--developer--claude-_RUN-260731-922dc9.log](file://BUG-260731-lepevi/BUG-260731-lepevi_spawn-log_-implementer--developer--claude-_RUN-260731-922dc9.log) — System spawn log captured by task-board
- [BUG-260731-lepevi_spawn-log_-implementer--developer--claude-_RUN-260731-4655a2.log](file://BUG-260731-lepevi/BUG-260731-lepevi_spawn-log_-implementer--developer--claude-_RUN-260731-4655a2.log) — System spawn log captured by task-board
- [BUG-260731-lepevi_spawn-log_-implementer--developer--claude-_RUN-260731-35d0f2.log](file://BUG-260731-lepevi/BUG-260731-lepevi_spawn-log_-implementer--developer--claude-_RUN-260731-35d0f2.log) — System spawn log captured by task-board
- [BUG-260731-lepevi_spawn-log_-implementer--developer--claude-_RUN-260731-7dca5f.log](file://BUG-260731-lepevi/BUG-260731-lepevi_spawn-log_-implementer--developer--claude-_RUN-260731-7dca5f.log) — System spawn log captured by task-board
- [BUG-260731-lepevi_linux-lane-outcome.md](file://BUG-260731-lepevi/BUG-260731-lepevi_linux-lane-outcome.md) — Linux lane fix: native-runner baseline and full green result (ubuntu Lint/Test/Race, macOS Test/Race), both unused findings removed as genuinely dead, the six compiled-build cases carved out against rc5-native-control-inventory-v1 with the carve-out asserted on every runner, adversarial proof the new skip class is narrow, gate exit codes, and the out-of-scope Windows red
- [BUG-260731-lepevi_ci-evidence.tar.gz](file://BUG-260731-lepevi/BUG-260731-lepevi_ci-evidence.tar.gz) — Focused plus full evidence: baseline red ubuntu Lint log and Test gate artifact at bd6ba08 (run 30615765014); green ubuntu and macOS Test gate artifacts at PR 11 (run 30620349565) incl. skips-observed.tsv, observed-cases.tsv, platform-cases.txt and ledger-consistency.txt; local before/after golangci-lint, darwin test-gate, gate-selftest, and the adversarial skip-class streams
- [BUG-260731-lepevi_spawn-log_-reviewer--reviewer--claude-_RUN-260731-a087ac.log](file://BUG-260731-lepevi/BUG-260731-lepevi_spawn-log_-reviewer--reviewer--claude-_RUN-260731-a087ac.log) — System spawn log captured by task-board
- [BUG-260731-lepevi_spawn-log_-implementer--developer--claude-_RUN-260731-dbbe43.log](file://BUG-260731-lepevi/BUG-260731-lepevi_spawn-log_-implementer--developer--claude-_RUN-260731-dbbe43.log) — System spawn log captured by task-board
- [BUG-260731-lepevi_successor-run-checkpoint-RUN-260731-dbbe43.md](file://BUG-260731-lepevi/BUG-260731-lepevi_successor-run-checkpoint-RUN-260731-dbbe43.md) — Cancelled-successor checkpoint: operator cancel honored, read-only verification of PR 11 signature/CI, PR 9 landed and no rebase required (behind_by=0), remaining gate ownership
- [BUG-260731-lepevi_review-verdict.md](file://BUG-260731-lepevi/BUG-260731-lepevi_review-verdict.md) — Reviewer verdict (RUN-260731-a087ac, Opus 5): ACCEPTED. Independent re-derivation of AC evidence, dead-code call graph, carve-out design vs platform-case-gate.sh, re-run gates, plus a new reviewer finding: all six carved-out cases FAIL on Windows with go-v1 go_toolchain_missing (GOROOT), unrelated to the inventory - scope input for BUG-260731-33v6zz.

## Created
2026-07-31T08:23:36Z

## Last Update
2026-07-31T10:04:01Z

## Assigned To
[reviewer] reviewer (claude)
