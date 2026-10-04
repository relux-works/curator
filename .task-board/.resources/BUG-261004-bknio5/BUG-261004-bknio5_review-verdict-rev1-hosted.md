# BUG-261004-bknio5 — gc-sweeps-live-runtime-on-uncertain-marks

Verdict: accepted, CR-BUG-261004-bknio5-1 revision 1. No actionable findings in the reviewed delta. Hosted-evidence mode: no local Go builds, tests, vet, lint, base replay, or mutation execution performed. This is independent source review and direct inspection of downloaded hosted evidence, not a claim of independent executable red/mutant runs.

## Identity and gate

Base: 934952a45953587a1d4184b692b3fb4ee401e732.
Candidate tree: 31d96981112f3a682357b8064b28ed8ba5c1d42b.
The CR validation resource identifies [run 37173583916](https://github.com/relux-works/curator/actions/runs/37173583916), snapshot commit 7ac0b9f3f79eb54f05e6c6374ad6bd8c977e2af3. Independently resolving that commit tree gives the exact candidate tree above (exit 0). gh run view --exit-status returned 0: 11/11 required jobs successful, including lint, three OS tests, two race lanes, interop, naming, and three gate self-tests. Optional self-hosted and candidate-suite jobs were skipped and are not claimed.

Downloaded all five test/race artifacts with gh run download (exit 0). Independently parsed go-test.json: 4/4 new CLI subtests pass in each of 5/5 lanes; both touched packages have terminal pass events in 5/5 lanes. No fail events. TestCollectStaysFailSafeAcrossConsecutivePasses: 7/7 subtests pass on both POSIX platforms (normal and race); Windows 6/7 pass and its POSIX-mode unreadable-directory test is explicitly skipped. The new unreadable-marker CLI row passes on Windows too. TestAuthoritativeGarbageCollectionRootsAreRetained has 11 pass events including parent/subtests and zero skips in every lane. The workflow pins rc.14 commit 43bf0a2506d5c354a73bbc3ea4623d4653db10c7 and supplies CURATOR_CONFORMANCE_ROOT. These GC observations are not a claim of complete product conformance.

Artifact go-test.json SHA-256:
- test Ubuntu: 6154bf823a69260ac5675683e92a2f39339ec8ff5a75b90befe6444e54a19d1a
- test macOS: 2505b251343c35948ecd6b3a7f6ca22d78885a5b00ee37a42019741332dced71
- test Windows: 3358239f22b362e0631479df9b237603b5eb0f2569b8f8f99667a7993826d25c
- race Ubuntu: 9bc0ebdf09c386d9d6722d69e362c66e1b665f62c31cfbde7ad434eed0158490
- race macOS: b68155b732dbcfc5c26cba494a35029ad9856aaaa9599aa0fc1d834cc187ba22

Initial strict JSON parser exited 1 on dependency-download text in the Ubuntu race stream. Revised event parser separated non-JSON lines, then validated all five streams and package terminal events (exit 0). This was an evidence-reader issue, not a test failure.

## Swept surfaces and acceptance criteria

| Surface | Review and evidence |
| --- | --- |
| AC1: runtime/cache safety | Collect returns on any marked.uncertain before sweepRuntime, protected cache Sweep, and external cache Collect. CollectRuntime also refuses. Recording-cache assertions remain and now verify an unmarked runtime survives. |
| Conservative references | markScopes and pruneConsumers logic unchanged: unknown registry stays byte-preserved; uncertain consumers remain registered across passes. CLI asserts registry bytes after both passes. |
| AC2: production entry | run([gc]) -> cmdGC -> real home lock -> collectUnderLock -> Collect. Three uncertainty rows each run a WriteBinShim launcher before corruption and before/after each of two passes. Directory at marker path induces a real read failure, without permission/root ambiguity. 3/3 uncertainty shapes, 6/6 uncertain passes per lane. |
| AC3: complete-reference control | Fourth CLI row keeps a live marked shim working and removes genuinely unreferenced runtime on pass one; second pass retains that result. 1/1 control, 2/2 passes per lane. |
| AC4: diagnostics | Initial warnings retain the concrete source error; added warning explicitly names runtime sweep skipped and incomplete live references. CLI checks source diagnostic plus both sweep warnings on every uncertain pass. |
| Architecture and bypass | Existing marking/pruning and lock boundary reused. Runtime-only API receives same conservative guard; internal two-pass cases call it and assert refusal. No new policy or side-channel state. |
| Scope | Exact diff is three authorized files, 133 additions/7 deletions. No CHANGELOG.md, LOGBOOK.md, unrelated code, or repository changes made by reviewer. |

AC coverage: 4/4, bounded to the requested behavior and hosted evidence.

## Red-first reasoning against base (not executed)

Base Collect calls sweepRuntime before checking uncertainty. Truncated registry yields no consumer-derived runtime marks; invalid and unreadable markers are omitted, so their live runtime is unmarked. Thus all 3/3 new uncertainty rows would lose the runtime, fail post-GC launcher execution and orphan-retention assertions, and miss the runtime-skip warning. Repeated-pass launcher checks remain meaningful because errors use Errorf rather than aborting before the second invocation. The complete-reference control retains its valid marker and deletes the orphan on base and candidate. Internal added unmarked-runtime assertions likewise fail against the base sweep ordering; CollectRuntime refusal assertions fail against its unguarded base implementation. Historical producer red exit codes are in the attached producer packet; not independently rerun or represented as current execution.

## Mutation review (inferred kills, not executed)

1. Required reorder mutant: move uncertainty check after runtime sweep. All three CLI uncertainty rows lose their executable and orphan; killed by assertions.
2. Required narrowed-source mutant: omit registry-read uncertainty while retaining registryUnknown. Truncated-registry row sweeps with empty runtime marks and loses shim/orphan; killed.
3. Reviewer-selected threshold mutant: change the Collect guard from len(uncertain)>0 to >1. Each isolated CLI fixture contributes exactly one uncertainty (other scopes absent), so all three uncertainty rows sweep and fail. This specifically proves the required single-error boundary, rather than only deleting a guard.

Inference coverage 3/3 named mutants; executed by this reviewer 0/3, as required by hosted-evidence mode. No fabricated red/mutant exit codes.

Run goal query returned no active goal (not goal-bound). Evidence is attached before the acceptance mutation; acceptance routes to integrating, not done. Producer-side integration remains outstanding.
