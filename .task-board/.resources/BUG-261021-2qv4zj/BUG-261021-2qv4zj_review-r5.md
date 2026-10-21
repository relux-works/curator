# BUG-261021-2qv4zj review r5

Head: `e32d101e1f29cc98be75f285554bd017d1e00568`  
Round base: `a2244ec13e223e3d6f5b96cc11ba22829bc6a2a2`  
Original base: `2c89fd46443d461e25bd377fb94a0c6754282838`

Verdict: **accepted**, with non-blocking P2 F5. F4 closes in the exercised cases. One non-blocking P2, F5, records a reader-availability limitation of the existing transaction replacement contract.

The board assignment names this reviewer. All eleven precondition resources, the prior review and companion probes, and the producer's r5 results were read before review. Both detached checkouts equal the requested SHA; all ten introduced commits verify `G`. No source was edited. Reviewer probes and controlled input/fallback instrumentation use Go overlays.

| Surface | Held or broken | Check run | Evidence |
| --- | --- | --- | --- |
| F4 closure | Held | Unchanged r4 probe: missing/extra/changed files for both adapters repair and resolve current; intact and missing-directory controls pass. | `original-probes.log`, exit 0; ten repair cases. |
| Complete replacement and recovery | Held | Production Resolve in child processes exits at before-backup, after-backup and after-install. Fresh production Resolve recovers all six adapter/boundary cases to the original complete byte map and versioned identity. Failure before replacement preserves the exact old tree, then retry succeeds. | `independent-recovery.log`, exit 0; original rollback/recovery tests also pass. |
| Continuous reader availability | Broken, P2 F5 | Two profiles share a projection; read through both links at the transaction boundaries. Old bytes before backup, ENOENT after backup, complete new bytes after install. No mixed tree observed. | `atomic-readers.log`, expected exit 1; four intended unavailable-reader observations across both adapters. |
| Bounded source refusal | Held | Through production Resolve, change the pinned source at the copy boundary after planning. Require ErrSkillProjectionIdentity through the repair wrapper, exactly one copy, no prepared transaction, unchanged old projection, removed staging, and successful retry after restoring the input. | `independent-refusal-corrected.log`, exit 0; instrumented input change only. |
| Read-only behavior | Held | Plain production Resolve reports stale; StatusOf executes. Compare manager-tree file bytes, link targets, modes and mtimes before and after. | `independent-refusal-corrected.log`, exit 0; both adapters. |
| Shared lifecycle | Held | Two profiles share one projection. First repair commits exactly one replacement; the second resolves current and explicit repair adds none. Remove the first profile with purge, run production maintenance, then resolve/read the survivor. | `independent-lifecycle.log`, exit 0. |
| F3 native/managed publication | Held | Unchanged r3 probes build the exact-source CLI, install excluded repository inputs plus a declared visible control, and inspect both adapters' native and managed surfaces. | `original-probes.log`, exit 0. |
| Copy fallback | Held | Repeat CLI/Resolve publication probes with only projection-link creation forced unavailable; assert actual directories before checking excluded/visible bytes. | `independent-copy.log`, exit 0; controlled fallback, not an actual Windows run. |
| Core selection and migration | Held | Production publication/repair tests cover all fixed exclusions, runtime/build roots, exported versus command-less scripts, old markers, raw-root links and recorded copies. Review every operation-parameter propagation into native/global/identity migration paths. | `focused.log`, exit 0; full suite supplements migration and global paths. |
| F1/F2 and audit cache | Held | Real CLI alias and context-module admission controls; inspect shared WalkContext and ReadModule consumers and unchanged ruleset 8/kind equality. | `focused.log`; audit/cache tests in the full suite. |
| Runtime and other publishers | Held | Trace install stageNode/targets and runtimestore independently of profile links; full-suite production runtime/command tests exercise the frozen stores. | Reviewed diff and full-suite log. |
| Narrowing controls | Held | All fifteen projection overlays rerun. Omitted rebuild fails exactly six required cases; staged-identity omission and native/managed raw-root controls reach their intended assertions. | `focused.log`, exit 0 for the harness; expected nested failures are controls. |
| Documentation and producer claims | Held with P2 qualification | Re-derived staged verification, typed error chain, one rebuild, shared reuse, separate publication/cache identities and held-operation propagation. The word “atomically” does not imply uninterrupted link readability under the existing engine. | Independent probes and diff; F5 below. |
| Public repository hygiene | Held | Review full task diff and added-line scan for personal paths, host/user identifiers and organization-specific disclosures; whitespace and all introduced signatures. | `hygiene.json`, `signatures.log`; no finding. |
| Build, vet, default suite | Held after prerequisite-corrected reruns | Normal .git clone, private cache and short task-local TMPDIR; no overlay in default suite. All 3,600 top-level names ran exactly once. | Build/vet exit 0; initial full run exit 1; all 20 environment-caused failed outcomes pass in corrected reruns. |

## F5: existing consumers temporarily lose the projection during rebuild

`internal/envprofile/skill_projection.go:109-116` targets the existing shared directory with the transaction engine. The engine first renames the live directory to backup (`internal/transaction/engine.go:421`) and later renames the staged directory to the live path (`:478`). Between those operations both existing skill links return ENOENT. A process interruption there extends the absence until recovery; the independent crash probe confirms recovery succeeds.

Reproduction: map the attached atomic-reader probe to `internal/envprofile/reviewer_atomic_test.go` with a Go overlay, then run `env -u GIT_DIR -u GIT_WORK_TREE go test -overlay "$OVERLAY" -count=1 -v ./internal/envprofile -run '^TestReviewerR5AtomicReaders$'`. Both profiles share a real projection; before-backup and after-install reads are positive controls. Four after-backup reads report absence.

This is P2, not a blocking mixed-tree finding. The orchestrator clarified in the assigned topic (reply to the review notice, event `$ZrMNhINYvUaKgolV6o41BVkOgySUnMxe_c_SxpE--og`) that the existing backup/install absence is permitted by the current no-half-written-or-mixed-tree contract. It occurs during explicit repair, and no partial file set was published. Follow-up: retain complete projection versions and atomically replace each consumer link, with continuous-reader checks; meanwhile qualify the documentation's availability meaning.

## Validation and limits

Go 1.25.5, Darwin arm64. Every test unsets GIT_DIR/GIT_WORK_TREE. Private HOME, GOCACHE and GOPATH/module cache; PATH excludes operator provider CLIs. Default suite has no overlay. No actual Windows runtime or case-sensitive filesystem run is claimed. Pinned core section 3.1 was read directly from spec commit `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`; this narrow round does not claim a new external-corpus conformance run.

Two reviewer setup failures receive no passing credit. First, the initial read-only snapshot used transaction.DigestPath on a fixture manager tree containing valid symlinks; it refused before StatusOf. A no-follow snapshot corrected the instrument, and both production entry points pass. Second, concurrent `go test -list` waited on the CLI package's TestMain host-GOROOT lock and timed out after 300 seconds. The follow-up inventory passed and lists 85 packages / 3,600 top-level names. These are disclosed reviewer harness issues, not product findings. The restricted PATH also omitted required language tools. The full invocation recorded 20 failed test/subtest outcomes in three packages: one MCP golden mismatch caused by missing npx, two Rust cases caused by missing rustc, and seventeen CI routing parent/subtest outcomes caused by missing node. The installed Node 26.12.0/npm 11.20.0 and project-pinned Rust 1.91.0 were made available in follow-up PATHs. No source or expected output was changed. The original full invocation retains exit 1; it is not relabeled green. Corrected runs and name-by-name reconciliation are recorded below.

Full default invocation: **85 packages**, all **3,600 top-level names exactly once**, no missing/extra names. Initial outcomes: **11,764 passes, 199 skips, 20 failures** (77 package passes, 3 package failures, 5 packages without test files). The 20 failed names are each independently mapped to a passing outcome in prerequisite-corrected runs: MCP golden alone, the entire rustsource package, and the entire ciworkflow package. All three reruns exit 0, with **zero remaining failed names**. This is complete suite coverage with disclosed environment correction, not a claim that the first full command exited 0. Conditional skips remain skips.

| Reviewer run | Exit | Seconds |
| --- | ---: | ---: |
| `build` | 0 | 1.54s |
| `vet` | 0 | 3.23s |
| `original-probes` | 0 | 41.74s |
| `focused` | 0 | 144.71s |
| `independent-recovery` | 0 | 14.20s |
| `independent-refusal` | 1 | 8.06s |
| `independent-refusal-corrected` | 0 | 8.42s |
| `independent-copy` | 0 | 81.74s |
| `independent-lifecycle` | 0 | 28.27s |
| `atomic-readers` | 1 | 6.76s |
| `inventory` | 1 | 301.89s |
| `inventory-final` | 0 | 4.19s |
| `full-suite` | 1 | 1517.31s |
| `mcp-golden-node` | 0 | 3.97s |
| `rustsource-toolchain` | 0 | 113.20s |
| `ciworkflow-node` | 0 | 40.57s |

The bounded atomic-reader probe intentionally exits 1 and supplies F5 evidence; its failure is separate from default-suite accounting. The first refusal/snapshot run and first inventory run are the disclosed harness failures, not accepted checks. The fifteen narrowing controls all reached their intended failures, including exactly six omitted-rebuild cases.

Exact commands, exit records, raw logs and overlay sources are retained under `.out/BUG-261021-2qv4zj-review-r5-evidence/`. Only neutral review material is attached publicly. Attachment presence is verified separately before handover; task-local scratch is removed after handover.


```json
{"verdict":"accepted","head":"e32d101e1f29cc98be75f285554bd017d1e00568","findings":[{"id":"F5","severity":"P2","file_line":"internal/envprofile/skill_projection.go:109","mechanism":"The existing transaction engine renames the live shared projection to backup before installing the staged tree. Existing consumer links temporarily return ENOENT; an interruption extends absence until recovery. Complete old/new trees and recovery hold, but reader availability is not continuous.","reproduction":"Run attached TestReviewerR5AtomicReaders through a Go overlay. Two profiles share a projection for each adapter. Both links read old bytes before backup, fail with ENOENT at PointAfterBackup, and read complete new bytes after install: four intended unavailable-reader observations.","required_fix":"Non-blocking follow-up under the orchestrator clarified contract: retain complete projection versions and atomically replace consumer links, with continuous-reader coverage; qualify documentation so recoverable whole-directory replacement is not described as gap-free availability."}]}
```
