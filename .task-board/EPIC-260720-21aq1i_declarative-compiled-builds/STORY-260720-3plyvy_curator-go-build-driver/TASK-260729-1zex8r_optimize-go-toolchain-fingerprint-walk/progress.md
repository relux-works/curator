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
- TASK-260729-2kaopg

## Checklist
- [x] Implement directory-scoped fingerprint traversal without changing canonical records or digest
- [x] Add equivalence and fail-closed regression coverage for paths, symlinks, mutations, cancellation and errors
- [x] Measure before and after fingerprint and cmd/curator runtime on clean count-one runs
- [x] Run focused godriver, go-v1 conformance/vector, build, vet, gofmt and diff gates
- [x] Attach exact evidence and complete developer handoff for independent review
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Tests written and passing
- [x] Coverage target ~80%+ for affected code
- [x] New task-scoped outcome artifact attached on the board for reports, logs, screenshots, or other produced evidence
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
ORCHESTRATOR DECISION 2026-07-29 06:5x +04: choose diagnostic option 1 and authorize this separate godriver-owned task to change internal/godriver/fingerprint.go. This is not a human blocker. Source evidence is on TASK-260729-2kaopg in TASK-260729-2kaopg_performance-diagnostic.md; retrieve it with task-board resource get. A digest-identical measurement prototype exists at .temp/TASK-260729-2kaopg/probe/main.go and is reference-only, not accepted code. Preserve TASK-260729-2kaopg behavior/worktree and do not edit it. Implement in this task-owned worktree, prove canonical-record and digest equivalence including failures/mutations, measure performance, rerun the exact go-v1 conformance/vector gates discoverable from repository contracts, attach evidence, and hand off to independent review. This task is on the Curator Go critical path and may use the currently eligible Opus slot; RUN-260729-3bd9db own slot remains cooling until 07:16:50 +04.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260729-b1885e, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260729-b1885e)
Directory-scoped fingerprint traversal implemented in internal/godriver/fingerprint.go (worktree .temp/TASK-260729-1zex8r/worktree, base 17804ce + the shared uncommitted godriver tree). Equivalence proven by an in-package differential suite that preserves the pre-change traversal verbatim as legacyFingerprintToolchain and asserts identical digest, identical canonical record set and identical diagnostic code+detail. Real host GOROOT: 16093 records, identical digest. Benchmark A/B in one process: legacy 3.42-3.60 s/op vs scoped 0.99-1.04 s/op (3.4x). cmd/curator before/after measurement queued behind a concurrent foreign ./... run on the same host.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260729-b1885e, pid=87871, exit=0)
TEST COMPLETION DIRECTIVE after RUN-260729-b1885e: producer deliberately stopped at 9/11 because TASK-260720-1nlmvv owns the current repository-wide suite. After that suite is terminal and host has no foreign Go test, use an independent Codex tester, not another Opus, to finish evidence without product edits. First prove the producer task differs from its imported baseline only by the exact fingerprint patch and owned differential test; if currentness has advanced, create two private disposable byte-identical currentness-candidate copies and apply only TASK-260729-1zex8r_fingerprint.diff plus its owned test to the candidate copy. Run clean count-one baseline/candidate fingerprint and cmd/curator timing with identical environment and no overlapping suites, then focused godriver/equivalence, go-v1 conformance/vector, build/vet/gofmt/diff gates. Require unchanged digest/records/diagnostics and target cmd/curator <=480s. Attach exact evidence, complete only truthful checklist items and hand off to-review. No shared-cache clearing, timeout increase, product behavior edit, staging, commit, publish or pin.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [tester] tester (codex) (run=RUN-260729-ecbe05, max_parallel=20)
spawn run started: [tester] tester (codex) (run=RUN-260729-ecbe05)
TESTER STOP-THE-LINE 2026-07-29: expected-red TestScopedDirsDoesNotReadFromAReplacedDirectory proves cached scoped directory handles can read detached old bytes after canonical directory replacement; focused go test exit 1. This violates mutation/race fail-closed AC, so timing and release gates were not continued. Tester directive forbids product edits. Exact evidence/options: TASK-260729-1zex8r_tester-mutation-regression.md; regression is attached in TASK-260729-1zex8r_fingerprint_equivalence_test.go. Recommend godriver developer rework path-binding validation, then rerun all gates.
agent completed: [tester] tester (codex) (exit=0)
spawn run completed: codex (run=RUN-260729-ecbe05, pid=20428, exit=0)
REWORK DIRECTIVE after tester RUN-260729-ecbe05: tester status=blocked is not an external blocker; route to product rework. Preserve its expected-red TestScopedDirsDoesNotReadFromAReplacedDirectory and all producer equivalence evidence. Implement option 1: before every reuse of a cached directory handle for an entry open, resolve that directory again from the trusted GOROOT root, compare filesystem identity against the cached handle, and fail closed on rename/replacement/dangling/error before reading through the cached handle. Preserve legacy canonical record ordering, digest and diagnostic code/detail; add deterministic replacement, ancestor replacement, symlink, ABA-attempt and cancellation hooks. Retain scoped traversal only where path binding is proven; security equivalence overrides the speed target. Rebenchmark the real GOROOT and continue only if materially faster; if <=480s cmd/curator is not met, report measured residual and optimize another semantics-preserving portion rather than weakening revalidation. Run focused godriver/conformance gates now, but defer repository-wide/race/performance timing while foreign TASK-260720-1nlmvv race-cmd is active. Attach revised exact diff/evidence and hand off to an independent tester, then reviewer. No shared-cache clearing, timeout increase, staging, commit, publish or pin.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260729-19d434, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260729-19d434)
Developer rework after tester regression TASK-260729-1zex8r_tester-mutation-regression.md. Root cause confirmed: the previous candidate reused ancestor directory handles in the DIGEST phase, so a directory renamed out of GOROOT and replaced kept feeding detached bytes into the hash. Chosen fix = tester option 1 (restore legacy path binding) applied only where it is load-bearing: the digest phase is back to a full root-relative os.Root.Open per file (byte-for-byte legacy), and the scoped traversal is kept only for the per-entry Lstat in the collection phase, whose results are all re-anchored later (files by the digest-phase Open + os.SameFile, links by root-anchored Readlink/Stat, directories by a root-anchored Lstat). scopedDirs deleted. Measured fingerprint A/B on host GOROOT: legacy 1.560-1.587 s/op vs new 1.163-1.168 s/op = 1.35x. Gates so far all real exit 0: build, vet, gofmt, git diff --check, go test ./internal/godriver (30.1s), same with CURATOR_CONFORMANCE_ROOT (34.9s), go-v1 vector run incl. TestFingerprintImplementationMatchesRC4ToolchainVector + TestToolchainFramingMatchesRC4Vector. fingerprint.go coverage: collect 100%, digest 85.4%, descend 88.7%. cmd/curator before/after measurement in progress.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260729-19d434, pid=24610, exit=0)
TEST COMPLETION DIRECTIVE after RUN-260729-19d434: Preserve the rework candidate and do not edit product code. Independently prove the detached-directory regression is closed by the new design: digest phase restores root-relative os.Root.Open plus SameFile, collection-only scoped Lstat cannot contribute detached bytes, and link/directory checks remain root-anchored. Run the producer and tester mutation suites including directory replacement, ancestor replacement, symlink/dangling, ABA attempt, cancellation, diagnostic equivalence, canonical record ordering and digest equality. Validate the exact fingerprint patch and owned tests against the imported baseline. Wait for a host-release barrier with no foreign Go/test process before decisive measurements; then complete identical-environment clean count-one baseline/candidate fingerprint and cmd/curator timing, require candidate cmd/curator <=480s, and record exact exit codes/times. Re-run focused godriver, CURATOR_CONFORMANCE_ROOT go-v1 vectors, build, vet, gofmt, diff and affected coverage gates; do not clear shared caches, increase timeouts, run unrelated repository-wide suites, or overlap decisive timing. If all criteria are proven, attach revised task-scoped tester evidence, truthfully complete checklist and hand off to review. If any security/equivalence/performance criterion fails, attach exact evidence and route to development. No stage, commit, publish or pin.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [tester] tester (codex) (run=RUN-260729-fb2506, max_parallel=20)
spawn run started: [tester] tester (codex) (run=RUN-260729-fb2506)
Tester rework finding: digest-phase detached-directory regression is green, but new TestToolchainWalkRejectsDirectoryReplacedByFile exits 1. A listed directory can be reclassified as a file leaf after root-path replacement, skipping legacy fs.WalkDir descent and omitting descendants. Exact evidence and recommendation: TASK-260729-1zex8r_tester-collection-race.md. Performance/conformance/release gates withheld after the trust-boundary failure; route to developer rework.
agent completed: [tester] tester (codex) (exit=0)
spawn run completed: codex (run=RUN-260729-fb2506, pid=48152, exit=0)
REWORK CYCLE 2 DIRECTIVE after tester RUN-260729-fb2506: tester found a second real collection-phase race and routed to-dev. Preserve the expected-red TestToolchainWalkRejectsDirectoryReplacedByFile, TASK-260729-1zex8r_tester-collection-race.md, the now-green digest-phase root-binding fix, all prior equivalence tests, and LOGBOOK evidence. Fix only the traversal semantics: retain the listed/scoped DirEntry directory decision independently from the later root-relative canonical classification; if an entry that was listed and scoped as a directory is absent, dangling, symlinked, replaced, or no longer a directory at the trusted-root recheck, fail closed before appending any leaf record or omitting descendants. Do not let any type change in either direction reinterpret traversal, add/skip descendants, or produce a digest from mixed path generations. Add deterministic matrix coverage for directory-to-file, directory-to-symlink, file-to-directory, ancestor replacement, ABA attempts, cancellation and diagnostic code/detail equivalence against legacy fs.WalkDir. Preserve root-relative os.Root.Open plus SameFile in digest phase and every accepted go-v1 vector. Security equivalence takes priority over speed. After focused mutation/godriver/conformance/build/vet/gofmt/diff/coverage gates are green, wait for no foreign Go/test process before clean identical-environment fingerprint and cmd/curator baseline/candidate timing; do not overlap TASK-260720-1nlmvv gates or clear shared caches. Require candidate cmd/curator <=480s, attach exact revised diff/evidence and hand off first to independent tester, then reviewer. No unrelated behavior, timeout increase, stage, commit, publish or pin.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260729-e7c7cc, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260729-e7c7cc)
ORCHESTRATOR TIMING ATTRIBUTION: benchmark attempt bench-fingerprint-03 started around 2026-07-29 08:14 +04 while TASK-260720-1nlmvv internal/install focused suite was still active on the same host. It violates the no-foreign-Go host barrier and is non-evidence regardless of exit or numbers. Do not cite it for acceptance. Allow the already-running short benchmark to finish, but the next independent Codex tester must rerun clean A/B timing only after all foreign Go/test processes are absent.
Developer rework in progress (session 2). Found the tester regression TestToolchainWalkRejectsDirectoryReplacedByFile red against the prior candidate: toolchainWalk.descend decided descent from the resolved lstat instead of the listed fs.DirEntry, so a listed directory replaced by a file was recorded as an F leaf and its descendants silently dropped, where fs.WalkDir still calls fs.ReadDir and fails closed. Fixed in internal/godriver/fingerprint.go: the listed entry type now selects both the resolving handle (root Lstat for listed directories, scoped Lstat otherwise) and the descent, so OpenRoot fails exactly where fs.ReadDir failed with the same code and detail. Tester test kept verbatim under its original name plus a broader table pinning code, detail and record set. Gates green: gofmt 0, build 0, vet 0, git diff --check 0, focused godriver 0, go-v1 conformance 0, vectors 0 (same 10 baseline skips as the before-tree). Fingerprint A/B 1.69-1.83s legacy vs 1.14-1.15s scoped. cmd/curator before/after measurement running.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260729-e7c7cc, pid=53437, exit=0)
ORCHESTRATOR MEASUREMENT VERDICT 2026-07-29 08:2x +04: RUN-260729-e7c7cc reached terminal while cmdcurator-measurement.txt contained only before-30m warm-compile exit=0 and no timed test exit/result; its child measurement process then disappeared. This is incomplete and non-evidence. Keep task in development. After TASK-260720-1nlmvv releases the host, spawn an independent Codex tester to rerun clean identical-environment fingerprint and cmd/curator A/B from zero, plus the full focused security/equivalence/conformance gates, before any reviewer routing.
INDEPENDENT TEST COMPLETION DIRECTIVE 2026-07-29 after TASK-260720-1nlmvv reached done and the host-release barrier is clear: preserve product code. Use .temp/TASK-260729-1zex8r/worktree as the rework source and the independently accepted .temp/TASK-260720-1nlmvv/worktree as the integrated currentness base. First prove the rework source differs from its imported baseline only by the exact owned fingerprint production/test delta; regenerate a task-scoped exact patch because the earlier attached fingerprint.diff predates both race fixes. Create two private disposable byte-identical copies of the accepted currentness candidate, baseline and candidate, and apply only the audited current fingerprint production patch (plus owned tests where needed) to candidate. Verify digest phase root-relative os.Root.Open plus SameFile, listed DirEntry controls descent, and root-anchored classification fails closed for directory-to-file, directory-to-symlink, file-to-directory, ancestor replacement, ABA, dangling/symlink, cancellation and diagnostic code/detail equivalence without mixed-generation records. Run producer/tester mutation matrices, canonical records/digest differential, focused godriver, CURATOR_CONFORMANCE_ROOT go-v1 vectors, build, vet, changed-file gofmt, diff and affected coverage sequentially. With no foreign Go/test processes, perform clean identical-environment -count=1 A/B fingerprint measurement and default-timeout go test ./cmd/curator -count=1 on baseline and candidate; no timeout override. Candidate must exit 0 in <=480 seconds. Record baseline truthfully even if its default gate times out. Do not clear shared caches, overlap commands, run unrelated repository-wide/race/lint suites, mutate accepted worktrees, stage, commit, publish or pin. Remove only task-owned disposable timing copies after checks if evidence is preserved. Attach TASK-260729-1zex8r_tester-evidence-cycle-3.md, refreshed exact patch and raw timing logs. If all security/equivalence/performance criteria pass, complete truthful checklist and tester-handoff to review; otherwise route development with exact evidence.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [tester] tester (codex) (run=RUN-260729-b6ab4c, max_parallel=20)
spawn run started: [tester] tester (codex) (run=RUN-260729-b6ab4c)
INDEPENDENT TESTER CYCLE 3 2026-07-29: security/equivalence matrix, focused godriver, pinned conformance/vector, build, vet, gofmt, diff/patch and coverage gates all exit 0. Host GOROOT: 16093 identical records, digest sha256:ea13c6bb11293e951ab9f189144a1f660cb2f398385109c0a3f7ad4875942191. Affected fingerprint.go coverage 87.1%. Clean A/B fingerprint 1.559s/op baseline vs 1.081s/op candidate (1.44x); literal default-timeout go test -count=1 ./cmd/curator exit 0 at 564.778s baseline vs 441.177s candidate, under 480s. Ten vector/conformance cases truthfully skipped because the accepted pre-revision root publishes no newer artifacts. Attached TASK-260729-1zex8r_tester-evidence-cycle-3.md, refreshed exact patch, and raw-evidence archive. No product edit, cache clearing, timeout override, staging, commit, publish or pin. Ready for independent review.
agent completed: [tester] tester (codex) (exit=0)
spawn run completed: codex (run=RUN-260729-b6ab4c, pid=7998, exit=0)
FINAL REVIEW DIRECTIVE cycle 3: Review only the current exact cycle-3 patch and attached evidence. Validate source scope, both previously reported race closures, and byte-identical accepted-currentness baseline/candidate construction with exact patch applicability and SHA. Replay or inspect the deterministic mutation matrix, focused godriver, CURATOR_CONFORMANCE_ROOT conformance/vector, build/vet/gofmt/diff/coverage gates and raw host-barrier/exit/timing logs. Confirm canonical records, digest, diagnostics and fail-closed behavior for directory-to-file, directory-to-symlink, file-to-directory, ancestor replacement, ABA, dangling/symlink and cancellation. The first baseline benchmark compile failure caused by candidate-only test internals was detected, discarded and is non-evidence; require the documented identical standalone benchmark rerun instead. Do not rerun the approximately 15-minute cmd/curator A/B unless evidence integrity or attribution fails. Accepted => done; rejected => development with exact findings. No product edits, broad/race/full-repository suites, shared-cache clearing, timeout changes, staging, commit, publish or pin.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260729-91ea7b, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260729-91ea7b)
REVIEWER VERDICT cycle 3: ACCEPTED. Exact patch SHA a7e0906612ce6f007bfdb3776de632dd9c7a673e9b501443be5fb3eced8f1beb applies cleanly to accepted currentness. Independent mutation/equivalence replay, real-GOROOT digest, focused godriver, pinned conformance, RC4 vectors, build, vet, gofmt and patch gates pass. Raw evidence confirms 1.44x fingerprint improvement and cmd/curator 441.177s under 480s. Full evidence: TASK-260729-1zex8r_reviewer-verdict-cycle3.md
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260729-91ea7b, pid=22467, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260729-1zex8r_spawn-log_-implementer--developer--claude-_RUN-260729-b1885e.log](file://TASK-260729-1zex8r/TASK-260729-1zex8r_spawn-log_-implementer--developer--claude-_RUN-260729-b1885e.log) — System spawn log captured by task-board
- [TASK-260729-1zex8r_fingerprint.diff](file://TASK-260729-1zex8r/TASK-260729-1zex8r_fingerprint.diff) — Scoped diff of internal/godriver/fingerprint.go: directory-scoped traversal replacing full-path os.Root resolution
- [TASK-260729-1zex8r_bench-fingerprint.txt](file://TASK-260729-1zex8r/TASK-260729-1zex8r_bench-fingerprint.txt) — A/B benchmark: legacy vs directory-scoped fingerprint over the host GOROOT, same process
- [TASK-260729-1zex8r_fingerprint_equivalence_test.go](file://TASK-260729-1zex8r/TASK-260729-1zex8r_fingerprint_equivalence_test.go) — Differential suite updated with expected-red directory-to-file collection race regression
- [TASK-260729-1zex8r_spawn-log_-tester--tester--codex-_RUN-260729-ecbe05.log](file://TASK-260729-1zex8r/TASK-260729-1zex8r_spawn-log_-tester--tester--codex-_RUN-260729-ecbe05.log) — System spawn log captured by task-board
- [TASK-260729-1zex8r_tester-mutation-regression.md](file://TASK-260729-1zex8r/TASK-260729-1zex8r_tester-mutation-regression.md) — Expected-red trust-boundary regression, exact evidence, constraint, and rework options
- [TASK-260729-1zex8r_spawn-log_-implementer--developer--claude-_RUN-260729-19d434.log](file://TASK-260729-1zex8r/TASK-260729-1zex8r_spawn-log_-implementer--developer--claude-_RUN-260729-19d434.log) — System spawn log captured by task-board
- [TASK-260729-1zex8r_spawn-log_-tester--tester--codex-_RUN-260729-fb2506.log](file://TASK-260729-1zex8r/TASK-260729-1zex8r_spawn-log_-tester--tester--codex-_RUN-260729-fb2506.log) — System spawn log captured by task-board
- [TASK-260729-1zex8r_tester-collection-race.md](file://TASK-260729-1zex8r/TASK-260729-1zex8r_tester-collection-race.md) — Expected-red collection race evidence, legacy divergence, and developer rework recommendation
- [TASK-260729-1zex8r_spawn-log_-implementer--developer--claude-_RUN-260729-e7c7cc.log](file://TASK-260729-1zex8r/TASK-260729-1zex8r_spawn-log_-implementer--developer--claude-_RUN-260729-e7c7cc.log) — System spawn log captured by task-board
- [TASK-260729-1zex8r_spawn-log_-tester--tester--codex-_RUN-260729-b6ab4c.log](file://TASK-260729-1zex8r/TASK-260729-1zex8r_spawn-log_-tester--tester--codex-_RUN-260729-b6ab4c.log) — System spawn log captured by task-board
- [TASK-260729-1zex8r_tester-evidence-cycle-3.md](file://TASK-260729-1zex8r/TASK-260729-1zex8r_tester-evidence-cycle-3.md) — Independent cycle-3 tester report: security equivalence, gates, coverage, and clean A/B performance
- [TASK-260729-1zex8r_fingerprint-cycle3.patch](file://TASK-260729-1zex8r/TASK-260729-1zex8r_fingerprint-cycle3.patch) — Refreshed exact production fingerprint patch plus owned differential test
- [TASK-260729-1zex8r_cycle3-raw-evidence.tar.gz](file://TASK-260729-1zex8r/TASK-260729-1zex8r_cycle3-raw-evidence.tar.gz) — Raw cycle-3 gate, timing, coverage, host-barrier, and harness-correction logs
- [TASK-260729-1zex8r_spawn-log_-reviewer--reviewer--codex-_RUN-260729-91ea7b.log](file://TASK-260729-1zex8r/TASK-260729-1zex8r_spawn-log_-reviewer--reviewer--codex-_RUN-260729-91ea7b.log) — System spawn log captured by task-board
- [TASK-260729-1zex8r_reviewer-verdict-cycle3.md](file://TASK-260729-1zex8r/TASK-260729-1zex8r_reviewer-verdict-cycle3.md) — Independent accepted reviewer verdict with scope, trust-boundary, gate, coverage, and performance evidence

## Created
2026-07-29T02:49:22Z

## Last Update
2026-07-29T06:10:05Z

## Assigned To
[reviewer] reviewer (codex)
