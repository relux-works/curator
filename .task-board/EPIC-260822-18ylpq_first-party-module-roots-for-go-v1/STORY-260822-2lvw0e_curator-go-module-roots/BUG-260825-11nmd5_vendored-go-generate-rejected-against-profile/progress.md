## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(1))

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] Vendored generator directives are inert at build time and the carve-out is bounded to materialized vendor trees per the decision 0005 relaxation family
- [x] Regression test pins acceptance of a vendored //go:generate and continued rejection of a first-party one
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant

## Notes
Picked up from skill-project-management TASK-260822-hje0ya, which this bug blocks: task-board-tui cannot build through go-v1 while vendored clipperhouse/displaywidth gen.go carries a bare //go:generate. Fix on branch bug/BUG-260825-11nmd5-vendored-go-generate (worktree .temp/BUG-260825-11nmd5/worktree): validatePackageInputs exempts //go:generate only for a package below the build root vendor tree whose module carries no replacement, mirroring the existing SFiles carve-out one block above it.
FIX LANDED ON PR https://github.com/relux-works/curator/pull/40 (branch bug/BUG-260825-11nmd5-vendored-go-generate, commit c9fe49c). internal/godriver/graph.go validatePackageInputs now rejects a generator directive only when the package is first-party (module carries a replacement) or its directory is not strictly below <build root>/vendor, mirroring the pure-Go-assembly carve-out one block above. Tests: the case joins TestAuditedVendorAllowancesAreWithheldFromAReplacedModule, whose fixture puts the audited and the replaced package in the SAME vendor tree so the firstParty guard is what does the work rather than the path. Both directions mutation-checked locally: reverting the carve-out reddens allowed_for_an_audited_third-party_module; dropping the firstParty guard reddens withheld_from_a_replaced_module. TestBuildRejectsDirectivesPGOAndMultipleRootsBeforeBuild/generator still covers the build-root package unchanged. Local make ci-test against SPEC_PIN 0ed5c69 is green; its single failure was internal/ui [setup failed] because a fresh git worktree does not populate the agents/skills/skill-go-testing-tools submodule, which passes after git submodule update --init.
Landed: PR 40 (Let a materialized vendor tree keep its generator directives) merged with every lane green pre-merge. Context: this was the real blocker behind the first cold-cache schema-8 install — vendored third-party code (clipperhouse/displaywidth gen.go) carries an inert //go:generate and go-v1 rejected it at preflight; the hje0ya acceptance review proved the earlier green status was a poisoned cache hit built by a patched curator (see the skill-board TASK-260822-hje0ya review verdict and the new memory note). REVIEWER SCOPE: the PR 40 delta — spec compliance of the vendored-generator carve-out against decision 0005 vendored-boundary relaxations and the profile scan-surface prose, plus the regression tests; targeted tests, cite lanes.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260825-0354c9, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260825-0354c9)
REVIEW VERDICT: CHANGES REQUESTED -> to-dev. Evidence: BUG-260825-11nmd5_review-verdict.md + BUG-260825-11nmd5_cgo-bypass-probe_test.go.

What is right in PR 40 (c9fe49c, merged as 680f6a6): the carve-out condition matches profiles/manager.md 2.3 at SPEC_PIN 0ed5c69 and decision 0005, mirrors the pure-Go-assembly carve-out, and all THREE halves of the guard are pinned by a test that reddens when that half is removed - I ran the mutations myself: reverting to `if matched == 2` reddens .../allowed_for_an_audited_third-party_module; dropping firstParty reddens .../withheld_from_a_replaced_module; dropping the path half reddens TestBuildRejectsDirectivesPGOAndMultipleRootsBeforeBuild/generator. Unmutated internal/godriver is ok in 41.859s. PR 40 lanes all SUCCESS (Test ubuntu/macos/windows, Race ubuntu/macos, Lint, Gate self-test x3, Interop conformance gate, Naming gate).

BLOCKING FINDING: the carve-out makes go_forbidden_compiler_directive bypassable. scanSourceDirectives (graph.go:300) returns on the FIRST 64 KiB window matching ANY needle. Within a window cgo_import_dynamic wins; ACROSS windows it does not. A //go:generate in window 1 sets matched=2 and stops the scan, so a //go:cgo_import_dynamic in window 2 is never read. Before c9fe49c matched==2 rejected unconditionally so this was harmless; now a vendored non-replaced package with that layout BUILDS CLEAN. Proven end to end through the real Build() entry point with two passing controls - see the attached probe. The Go compiler does not backstop this: cmd/compile noder.go permits //go:cgo_import_dynamic for general use (the x/sys/unix Solaris comment), and /usr/lib/libSystem.B.dylib satisfies safeArg. Spec 2.3 is a containment predicate - rejected if it CONTAINS the bytes - so this is a conformance break, not only a hardening gap.

FIX: in scanSourceDirectives, short-circuit only on the cgo needle; record a //go:generate hit and keep reading to EOF. Exact snippet in the verdict artifact. Pin it with the attached probe or an equivalent case folded into TestAuditedVendorAllowancesAreWithheldFromAReplacedModule - TestDirectiveScanFindsExactTokenAcrossReadBoundary covers one needle only and cannot catch this.

NON-BLOCKING, recorded not for rework: (1) the firstParty guard on //go:generate is stricter than 2.3 requires, since the replaced-module sentence enumerates only SFiles and cgo_import_dynamic - fail-closed and the AC chose it deliberately; (2) the pinned conformance suite has the same hole, worth a curator-spec follow-up; (3) classifyDeclaredInput (moduleroots.go:227) never scans for //go:generate despite its doc comment, pre-existing and untouched here.

Reviewer ran read-only: review worktree .temp/BUG-260825-11nmd5/review-wt (detached at origin/main), removed after the run. The story worktree is at 903af23 and predates the fix; nothing changed there.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260825-0354c9, pid=21797, exit=0)
REWORK per the review blocking finding: scanSourceDirectives (internal/godriver/graph.go:300) returns on the FIRST window matching ANY needle — after the carve-out, a //go:generate early in a vendored file terminates the scan and a //go:cgo_import_dynamic later in the same file is never read, riding in unnoticed (matched==2 is now accepted for vendored non-replaced packages). Fix: the scan must not stop until the highest-severity needle is resolved — e.g. keep scanning after a generate match until EOF or a cgo_import_dynamic match (which always wins), or track both needles independently across windows. Add the exact regression fixture the reviewer describes: vendored file with //go:generate in window 1 and //go:cgo_import_dynamic in a later window must be REJECTED with go_forbidden_compiler_directive. Take the non-blocking observations from the verdict where cheap. Land via PR, all lanes green pre-merge. Executor: claude opus-5, effort high.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260825-6e4450, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260825-6e4450)
Vendored //go:generate is accepted again, bounded to the materialized vendor tree. internal/godriver/graph.go: the generator branch now reads `matched == 2 && !vendoredDependency(item, validation.BuildRoot)`; the new helper is true only when the result carries module metadata, that module is NOT the main module, and the package directory is strictly below <buildRoot>/vendor. The main-module guard is load-bearing: a bare path-prefix test would hand the exemption to first-party code whose declared source_dir sits below vendor/. cgo_import_dynamic allowlist and the SFiles carve-out untouched. Tests: TestPackageGraphExemptsGeneratorDirectiveOnlyBelowTheVendorTree (vendored accept; build-root reject; first-party-below-vendor reject; vendored cgo_import_dynamic still rejected) and TestBuildCompilesThroughAVendoredGeneratorDirective, which drives Build(...) end to end. Four mutants applied and reverted, all killed, including the narrowing one that drops only the Module.Main guard. Gates against SPEC_PIN 00b1688 all exit 0: godriver (plus -race), all 34 served packages, the 7 deferred ones with the root unset, golangci-lint, go vet, gofmt, gate-selftest 75/0, ledger-consistency 72 rows, no-broad-suppression. Caveat recorded in the artifact: the first cmd/curator run exited 1 on a Go build-cache testlog write limit while the binary printed PASS; the -count=1 rerun exited 0 and is the cited run. linux/windows lanes not run on this macOS host. Evidence: BUG-260825-11nmd5_implementation-and-evidence.md.
agent completed: [implementer] developer (claude) (exit=124)
spawn run completed: claude (run=RUN-260825-6e4450, pid=31533, exit=124)
spawn run RUN-260825-6e4450 failed; operator action required; failure: run exceeded --timeout 45m0s and was terminated by the launcher
RESPAWN NOTE (previous run timed out without pushing): the delta is SMALL — internal/godriver/graph.go scanSourceDirectives: do not terminate the scan on a generate match; continue until EOF or a cgo_import_dynamic match, which always wins (or track the two needles independently across windows). One regression fixture: vendored file with //go:generate in the first 64KiB window and //go:cgo_import_dynamic in a later window must be rejected with go_forbidden_compiler_directive. Skip broad refactors and full local suites — targeted go test ./internal/godriver, push, PR, lanes verify, handoff. Prior context in the review verdict resource. Note PR 39 (fix/vendored-go-generate-inert) is a stale alternative branch — ignore or close it, do not build on it.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260825-3ef162, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260825-3ef162)
REWORK PUSHED: PR https://github.com/relux-works/curator/pull/42 (branch fix/BUG-260825-11nmd5-directive-scan-shortcircuit, commit 438d557, base 680f6a6). Addresses the blocking review finding.

DEFECT REPRODUCED FIRST, on unfixed origin/main: the reviewer probe through Build() returned diagnostic code "" and the build SUCCEEDED for an audited non-replaced vendored module carrying //go:generate in window 1 and //go:cgo_import_dynamic in window 2; both controls correctly returned go_forbidden_compiler_directive. The bypass was real, not theoretical.

FIX: scanSourceDirectives now resolves by SEVERITY, not by first hit. Only //go:cgo_import_dynamic — which nothing weaker can override — ends the scan early; a //go:generate hit is recorded and the file is still read to EOF. The three verdicts are named constants (directiveNone/directiveCgoImportDynamic/directiveGenerate) so the call site reads as a verdict, not as matched==1/matched==2. Carve-out, golang.org/x/sys allowlist and SFiles exception all unchanged.

TESTS: TestDirectiveScanReportsTheStrongestDirectiveAcrossWindows (scanner: generate-before-cgo, cgo-before-generate, generate-only, generate in a later window, neither) and TestVendoredGeneratorCarveOutDoesNotHideACgoImportDynamic (drives the real Build() entry point over the exact class the carve-out serves, with same-window and cgo-only controls).

FOUR MUTANTS APPLIED AND REVERTED, ALL KILLED, including the narrowing one: gating the cgo check on matched==directiveNone keeps the keep-scanning change but lets a recorded generate suppress a later cgo hit — reddens generate_before_cgo and the e2e middle case, which is what proves the bound is severity across the whole file rather than read-one-more-window. A fourth mutant removes the carve-out entirely and reddens PR 40 allowed_for_an_audited_third-party_module, proving the hardening did not quietly undo the relaxation.

LOCAL GATES, real exit codes: go build ./... 0; go test ./internal/godriver -count=1 0 (43.3s); go test -race 0 (79.7s); go vet ./... 0; gofmt -l . 0 no output; golangci-lint run 0, 0 issues; gate-selftest.sh 0, 81/0; no-broad-suppression.sh 0. NOT RUN locally: conformance-root gates (make ci-test/race need CURATOR_CONFORMANCE_ROOT) and the linux/windows lanes — CI covers them.

WORKTREE NOTE FOR THE ORCHESTRATOR: the STORY-260822-2lvw0e story worktree (HEAD 903af23, PRE-PR-40) still carries UNCOMMITTED work from the timed-out run RUN-260825-6e4450 — a divergent second implementation of the same carve-out using a Module.Main guard instead of the firstParty guard that actually landed in PR 40, plus its own LOGBOOK entry. It is dead work against a stale base and must NOT be committed or integrated; a patch of it is saved at .temp/BUG-260825-11nmd5/evidence/story-worktree-superseded.patch. This PR is based on origin/main and is the live delta.
PR 42 LANES ALL GREEN PRE-MERGE (run 32800314357, commit 438d557): gh pr checks 42 exit 0; rollup 11 SUCCESS / 1 SKIPPED; PR MERGEABLE. Test ubuntu 1m46s, Test macos 10m23s, Test windows 25m57s, Race ubuntu 3m10s, Race macos 13m18s, INTEROP CONFORMANCE GATE 21s, Lint 28s, Gate self-test x3, Naming gate — all pass; Candidate suite skipping. The Interop conformance gate is the lane that discharges the AC clause conformance suite against SPEC_PIN stays green, since the local host cannot materialize the pinned root.

PR 42 IS LEFT OPEN AND UNMERGED: lanes are verified green pre-merge, and merging is the reviewer/orchestrator step, not the developer role. Handing off to review.

Evidence artifact updated with the lane table: BUG-260825-11nmd5_rework-scan-severity.md.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260825-3ef162, pid=87086, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260825-ff8043, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260825-ff8043)
REVIEW VERDICT (CR-BUG-260825-11nmd5-1 rev1, run RUN-260825-ff8043): CHANGES REQUESTED -> to-dev. Evidence: BUG-260825-11nmd5_review-verdict-cr1-rev1.md, BUG-260825-11nmd5_cgo-shadow-probe_test.go.txt, BUG-260825-11nmd5_cgo-shadow-probe.log.

FINDING 1 (blocking): the delta is SUPERSEDED DEAD WORK. This bug is already fixed and merged on origin/main by c9fe49c (PR 40, merged 680f6a6) plus 438d557 (PR 42, merged e027667). The CR base 903af23 is NOT an ancestor of origin/main; merge-base is 1f55f1b, main is 20 commits ahead, and the story branch carries only the board-reconciliation commit. The candidate tree 93b7702 is exactly the uncommitted Module.Main variant that the RUN-260825-3ef162 developer handoff already flagged as dead work that must not be committed or integrated. Landed main uses a different guard shape (firstParty := item.Module.Replace != nil), so accepting this revision would land a second conflicting implementation of a carve-out that already exists upstream.

FINDING 2 (blocking, proven end to end): the delta reintroduces the go_forbidden_compiler_directive bypass that PR 42 exists to close. scanSourceDirectives returns on the FIRST read window matching ANY needle; cgo wins only WITHIN a window. A //go:generate in window 1 stops the scan and a //go:cgo_import_dynamic past 64 KiB in the same file is never read. On the base that was harmless because matched==2 rejected unconditionally; the candidate turns matched==2 into an ACCEPTANCE for vendored packages, so the file rides in clean. Same probe, same fixture (118123-byte vendored value.go), three trees: base 903af23 -> scan=2, go_generator_forbidden (rejected); CANDIDATE 93b7702 -> scan=2, validatePackageGraph returns NO ERROR (admitted); origin/main e027667 -> scan=1, go_forbidden_compiler_directive (rejected). This is a conformance break, not just hardening: profiles/manager.md 2.3 is a containment predicate (rejected if it CONTAINS the bytes), and it also fails the AC clause the cgo_import_dynamic allowlist is unchanged in substance.

WHY THE NEW TESTS MISSED IT: the graph_test.go subtest vendored cgo_import_dynamic writes a three-line value.go, so everything lands in one read window. That is the control case which passes on every version including the broken one. Nothing in the delta exercises a file larger than one Read chunk. origin/main already carries the two tests that do bound the class (TestDirectiveScanReportsTheStrongestDirectiveAcrossWindows, 5 subtests; TestVendoredGeneratorCarveOutDoesNotHideACgoImportDynamic, 3 subtests) - both green when I ran them.

WHAT IS RIGHT (do not throw away): the three-part guard is genuinely pinned. Four mutants applied and reverted by me: (A) gate deleted -> killed, two subtests red; (B) drop the Module.Main half keeping the path prefix -> killed, first-party_package_below_the_vendor_tree red - this is the narrowing mutant and it proves the bound is first-party/third-party, not a path prefix; (D) extend vendoredDependency to the matched==1 branch -> killed. (C) drop the path half -> SURVIVED, but it is an EQUIVALENT mutant, not a coverage gap: validateModule runs before the directive scan and already rejects any non-main module not strictly below buildRoot/vendor with vendor_dependency_missing, so the path half is unreachable defence-in-depth. TestBuildCompilesThroughAVendoredGeneratorDirective correctly drives the real Build() entry point. The Module.Main guard is defensible on its own and is NOT the reason for this verdict; it is simply not the shape that landed.

GATES I RAN MYSELF at candidate tree 93b7702 on darwin/arm64: go build ./... 0; go vet ./internal/godriver 0; gofmt -l 0 no output; golangci-lint run ./internal/godriver/... 0, 0 issues; go test ./internal/godriver -count=1 ok 29.816s; both new tests -v PASS (3 subtests); 4 mutants as above; cgo-shadow probe across three trees. NOT re-run by me, accepted from the implementer evidence: 34 served + 7 deferred conformance packages, -race, gate-selftest, ledger-consistency, no-broad-suppression - given Finding 1 those describe a delta that should not land regardless. linux/windows lanes not run, no local runner.

REQUIRED REWORK: (1) drop this delta, do not commit 93b7702; (2) bring task-board/story/STORY-260822-2lvw0e up to origin/main so it inherits c9fe49c + 438d557 instead of re-deriving the carve-out from 903af23; (3) if the story still needs its own change here after that, it must keep main severity-resolving scanner - only //go:cgo_import_dynamic may terminate the scan, a //go:generate hit is recorded and the file is still read to EOF; (4) any re-added fixture must include the CROSS-WINDOW case, not only the same-window control. Rebase/integration sequencing is the orchestrator call, not mine.

REVIEWER HYGIENE: read-only. Probes were added, run, and removed; a detached origin/main worktree under .temp was created and removed. Working tree verified byte-identical to candidate tree 93b7702 before and after (git diff --stat 93b7702 -- . empty). git status at hand-off is the four CR files and nothing else. Branch not switched, rebased, merged, or deleted.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260825-ff8043, pid=32833, exit=0)
CLOSED BY THE ORCHESTRATOR per the cr1-rev1 verdict own Finding 1: the complete fix is already merged on origin/main — c9fe49c (vendored generator carve-out, PR 40) + 438d557 (severity-resolving scanner with the cross-window regression fixture, PR 42, e027667), both citing this bug, both landed with every lane green. The cr1-rev1 review examined a stale re-derivation from base 903af23 and its required rework is to DROP that delta (never pushed) — done, nothing to integrate. Substance review trail: cycle-1 verdict caught the scan-shortcircuit bypass; PR 42 delivers exactly the demanded fix including the cross-window fixture. Live effect verified earlier: cold-cache schema-8 install of skill-project-management builds all three binaries on stock main.

## Precondition Resources
(none)

## Outcome Resources
- [BUG-260825-11nmd5_spawn-log_-reviewer--reviewer--claude-_RUN-260825-0354c9.log](file://BUG-260825-11nmd5/BUG-260825-11nmd5_spawn-log_-reviewer--reviewer--claude-_RUN-260825-0354c9.log) — System spawn log captured by task-board
- [BUG-260825-11nmd5_review-verdict.md](file://BUG-260825-11nmd5/BUG-260825-11nmd5_review-verdict.md) — Reviewer verdict on PR 40 (c9fe49c): changes requested — the vendored //go:generate carve-out makes the //go:cgo_import_dynamic gate bypassable across read windows
- [BUG-260825-11nmd5_cgo-bypass-probe_test.go](file://BUG-260825-11nmd5/BUG-260825-11nmd5_cgo-bypass-probe_test.go) — End-to-end probe reproducing the cgo_import_dynamic bypass through Build(); two controls plus the failing cross-window case
- [BUG-260825-11nmd5_spawn-log_-implementer--developer--claude-_RUN-260825-6e4450.log](file://BUG-260825-11nmd5/BUG-260825-11nmd5_spawn-log_-implementer--developer--claude-_RUN-260825-6e4450.log) — System spawn log captured by task-board
- [BUG-260825-11nmd5_implementation-and-evidence.md](file://BUG-260825-11nmd5/BUG-260825-11nmd5_implementation-and-evidence.md) — Vendored //go:generate carve-out: change, bound rationale, mutation evidence, and per-command gate results against SPEC_PIN 00b1688
- [BUG-260825-11nmd5_spawn-log_-implementer--developer--claude-_RUN-260825-3ef162.log](file://BUG-260825-11nmd5/BUG-260825-11nmd5_spawn-log_-implementer--developer--claude-_RUN-260825-3ef162.log) — System spawn log captured by task-board
- [BUG-260825-11nmd5_rework-scan-severity.md](file://BUG-260825-11nmd5/BUG-260825-11nmd5_rework-scan-severity.md) — Rework for the review blocking finding: reproduced bypass, severity-based scan fix, 4 killed mutants incl. the narrowing one, local gate exit codes, PR 42 lane table (11 SUCCESS), worktree hazard note
- [BUG-260825-11nmd5_change-request_rev1.patch](file://BUG-260825-11nmd5/BUG-260825-11nmd5_change-request_rev1.patch) — Change Request CR-BUG-260825-11nmd5-1 revision 1 candidate patch (repository_delta=present, 4 changed paths)
- [BUG-260825-11nmd5_spawn-log_-reviewer--reviewer--claude-_RUN-260825-ff8043.log](file://BUG-260825-11nmd5/BUG-260825-11nmd5_spawn-log_-reviewer--reviewer--claude-_RUN-260825-ff8043.log) — System spawn log captured by task-board
- [BUG-260825-11nmd5_cgo-shadow-probe_test.go.txt](file://BUG-260825-11nmd5/BUG-260825-11nmd5_cgo-shadow-probe_test.go.txt) — Reviewer probe driving validatePackageGraph over a vendored file with //go:generate in window 1 and //go:cgo_import_dynamic in window 2
- [BUG-260825-11nmd5_cgo-shadow-probe.log](file://BUG-260825-11nmd5/BUG-260825-11nmd5_cgo-shadow-probe.log) — Probe run log on the candidate tree: scan returns generate, validatePackageGraph returns no error
- [BUG-260825-11nmd5_review-verdict-cr1-rev1.md](file://BUG-260825-11nmd5/BUG-260825-11nmd5_review-verdict-cr1-rev1.md) — Reviewer verdict for CR-BUG-260825-11nmd5-1 rev1: CHANGES REQUESTED. Delta is superseded by merged PR40+PR42 on origin/main and reintroduces the cross-window cgo_import_dynamic bypass

## Created
2026-08-24T21:44:43Z

## Last Update
2026-08-25T02:47:21Z

## Assigned To
[reviewer] reviewer (claude)
