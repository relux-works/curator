# BUG-261004-bknio5 — gc-sweeps-live-runtime-on-uncertain-marks

Verdict: **accepted**. Independent review of CR-BUG-261004-bknio5-1 revision 1. All four acceptance criteria hold; no blocking findings.
Base: `934952a45953587a1d4184b692b3fb4ee401e732`.
Candidate: `31d96981112f3a682357b8064b28ed8ba5c1d42b`.

The review used an isolated archive of the exact candidate. Only temporary copies were mutated; the Story worktree, index and branch were not modified. The pre-fix replay substitutes the base gc.go into the candidate: because the only production delta is gc.go, this is exactly base production plus the new tests.

| Swept surface | Evidence / finding |
| --- | --- |
| Production dispatch and locks | Read run -> cli.cmdGC -> collectUnderLock -> scopes.Collect; real CLI tests take home locks and execute WriteBinShim launchers. |
| Uncertain registry and markers | All 3/3 requested uncertainty cases exercised twice; base breaks the shim, candidate preserves live and orphan runtime, registry bytes and explicit reason/skip warnings. |
| Conservative consumer and build handling | 11/11 internal conservative cases pass, including permission and Unix redirect cases, across two passes; recording cache remains uncalled and consumers remain known. |
| Legacy runtime-only API | CollectRuntime now returns an explanatory refusal before sweep; conservative tests exercise it on every pass. |
| Complete reference set | CLI control removes the orphan and keeps the working shim on both passes. Existing collection, journal, grace and lock tests pass. |
| rc.14 retention roots | Exact committed pin 43bf0a2506d5c354a73bbc3ea4623d4653db10c7: 5/5 GC roots pass (receipts, journals, markers, snapshots, uncertainty). Other status/repair rows retain their explicit existing bounds, not a new status/repair conformance claim. |
| Scope and architecture | Three requested files only. Mark/prune/sweep structure and conservative semantics retained. No CHANGELOG.md or LOGBOOK.md edits. |
| Upstream freshness | Fresh origin main ca1b776fb580ec0cee0173bf150daf063023aeaa differs from base only in CHANGELOG.md; no overlap with candidate paths. |

All reviewer Go commands set GOFLAGS=-work and use -count=1. Local checks are bounded and sequential; no test was abandoned as background work.

Hosted CR gate: https://github.com/relux-works/curator/actions/runs/37173583916 . Independent `gh run view 37173583916 --exit-status` returned 0, conclusion success; 11/11 required jobs passed, two optional lanes skipped. Head 7ac0b9f3f79eb54f05e6c6374ad6bd8c977e2af3 resolves to the exact candidate tree. The attached CR validation log also records exit 0. This supersedes reliance on the producer's earlier hosted run. Full OS/race suites, build and lint are accepted from this exact-tree hosted evidence, not claimed as independently repeated locally.

The broad local command accidentally selected an older local conformance checkout first; its green is not used as rc.14 evidence. The separately exported exact committed pin was then tested successfully. The broad mask selected no runtimestore tests; no local runtimestore-suite coverage is claimed. Exact hosted coverage remains the authority for that suite.

No product findings arose in the free hunt beyond the named surface sweep. The runtime-only bypass is covered by the candidate. Permission failure and invalid file kind are separately represented by internal and CLI tests.

Independent commands and observed exits:

| Replay | Command / alteration | Exit |
| --- | --- | ---: |
| Base red | `go test ./cmd/curator -run '^TestGCPreservesRuntimeOnUncertainMarks$' -count=1 -v`; base production with new tests | 1 |
| Candidate controls | `go test -p 2 ./internal/scopes ./internal/runtimestore ./cmd/curator -run '^(TestGC|TestCollect|TestAuthoritativeGarbageCollection|TestNonDirectory|TestRecording|TestParseConsumers|TestARepeatedRegistry|TestWritersRefuse|TestStructDecoding)' -count=1 -timeout=5m -v` | 0 |
| Exact rc.14 | `go test -p 2 ./internal/scopes -run '^TestAuthoritativeGarbageCollectionRootsAreRetained$' -count=1 -timeout=3m -v` with exact-pin CURATOR_CONFORMANCE_ROOT | 0 |
| Required mutant 1 | Move Collect uncertainty guard after sweepRuntime; CLI regression above, -p 2 -timeout=3m | 1 |
| Required mutant 2 | Omit only registry-read uncertainty append, keeping registryUnknown; same CLI regression | 1 |
| Reviewer mutant | Narrow Collect guard from any uncertainty to registryUnknown only; same CLI regression | 1 |
| Patch whitespace | `git diff --check BASE CANDIDATE` | 0 |

All 3/3 mutants killed. Reorder breaks all three uncertainty rows; dropped registry uncertainty breaks only the registry row; registry-only narrowing breaks both marker rows. All three mutants preserve the passing normal-removal control. Each variant was restored from saved candidate bytes in a finally block, with equality asserted. Base regression failures include actual launcher failure (no such file or directory), not only diagnostic mismatch.

Restored final replay: `go test -p 2 ./cmd/curator ./internal/scopes -run '^(TestGCPreservesRuntimeOnUncertainMarks|TestCollectStaysFailSafe|TestAuthoritativeGarbageCollectionRootsAreRetained)' -count=1 -timeout=3m -v`, with GOFLAGS=-work and exact rc.14 root: **exit 0**. Restored source equality asserted. Test logs are attached as `BUG-261004-bknio5_review-tests-rev1.md` with personal temporary paths redacted.

Host observation: syspolicyd successive crashes/runs were 398/501 before local tests, after base and green runs, between mutants and at closure; no increase observed. The first service-name probe used an incorrect name and was corrected to com.apple.security.syspolicy before test execution. No inference about overall host health is made.

No source edits, commits, index changes, branch changes, LOGBOOK.md or CHANGELOG.md edits were made by the reviewer. Findings are persisted here and in board notes instead, following the task-specific no-logbook rule. The run is not goal-bound (spawn goal reports none). Accept revision 1 through accept_cr and route to integrating; no done transition or commit_ack belongs to this review.
