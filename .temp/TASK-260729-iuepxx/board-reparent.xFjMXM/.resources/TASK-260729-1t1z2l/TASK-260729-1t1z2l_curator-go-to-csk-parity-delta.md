# Curator Go → CocoaSkills Go parity delta

**Board task:** `TASK-260729-1t1z2l`

**Research date:** 2026-07-29

**Scope:** read-only reconnaissance across Curator, CocoaSkills/csk, the task board, the frozen schema-6/`go-v1` wire contract, and the accepted rc.5 portable-execution amendment. No product, test, configuration, staging, commit, pin, or publication changes were made.

## 1. Executive finding

The 17 CocoaSkills Go tasks already form a coherent independent Python implementation plan. Every task has a concrete Curator source/test analogue, but the reusable boundary is **protocol behavior and test vectors, not Go code**. CocoaSkills must implement its own Python domain models, process layer, protected filesystem backends, transaction engine, and CLI integration.

Implementation is not ready to start today:

1. The CocoaSkills root tasks `TASK-260720-z9j4c9` and `TASK-260720-z2z795` are both hard-blocked by Curator gate `TASK-260720-1pvfj5`.
2. That Curator gate waits on in-flight currentness/repair work `TASK-260720-1nlmvv`, then Curator conformance `TASK-260720-jrrgw9`, documentation `TASK-260720-2qqq0w`, and cross-platform CI.
3. Shared vector consumption `TASK-260720-12r55p` has an additional independent prerequisite: `TASK-260720-3ag6pi`. Its literal rc.4 composite validates as a working tree, but the reviewer correctly left it `blocked` because no real rc.4 commit/ref exists.
4. The CocoaSkills local `main` is clean but two commits behind `origin/main`; it must be fast-forwarded before any task worktree is created.

There is also a required contract-rerouting decision. The 17 CocoaSkills task briefs still name rc.4, while the accepted Curator driver now implements the human-authorized rc.5 `manager-worker-v1` amendment from `TASK-260728-zb2s4z`. The schema-6 declaration, build-source, metadata, receipt, and marker bytes remain frozen, but execution-policy identity, worker/control behavior, and conformance-suite identity are rc.5. The existing logbook recommends rc.5 supersession because neither rc.4 nor rc.5 has been published; a board owner must explicitly retarget the literal rc.4 CocoaSkills briefs before implementation or choose the less likely rc.4 publication path.

The fastest valid start after the Curator gate closes is to launch the two CocoaSkills roots in parallel:

- `TASK-260720-z9j4c9` — schema-v6 build model.
- `TASK-260720-z2z795` — lock/journal transaction engine.

## 2. Repository and evidence provenance

### 2.1 Repositories

| Repository | Inspected state | Provenance and caveat |
| --- | --- | --- |
| Curator | `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1ljev5/worktree` | Reviewer-accepted composite is an **uncommitted task worktree** based on `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`. On 2026-07-29, `git ls-remote origin refs/heads/main` still returned that same commit. The composite includes the 14 reviewed `done` implementation tasks through safe build-cache GC. |
| Curator in-flight | `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1nlmvv/worktree` | Also based on `17804cea…`; adds currentness/status/repair diagnostics over the accepted composite. Board status was `development` after a changes-requested review. It is evidence for unresolved delta only, not an accepted implementation baseline. |
| curator-spec rc.4 candidate | `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3ag6pi/worktree` | Uncommitted composite based on `57c1f56846d221ecc55786bd3c2467ec32f11730`; `origin/main` still returned that rc.3-era commit. Individual schema/design/docs work was reviewed, but integrated verification is `blocked` until an authorized rc.4 commit/ref exists and gates are rerun without a virtual-index wrapper. |
| curator-spec rc.5 portable amendment | `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-zb2s4z/curator-spec-worktree` | Reviewer-accepted, human-authorized `manager-worker-v1` execution amendment, still uncommitted on base `57c1f568…`. Candidate manifest digest recorded by its accepted review is `sha256:58f8d2299c8f4a5ed78546913e567f637f1cc905dc5212b460f4097be7ff2af9`. The successor `TASK-260728-2kp3tv` retains local schema-6/`go-v1` identities while advancing the unreleased rc.5 candidate for external repository work. |
| CocoaSkills/csk | `/Users/iv/Developer/Wildberries/cocoaskills` | Local clean `main` at `edce8816dda44bb121d661b7c4dea942558ce408`, two commits behind local/remote `origin/main` `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`. Remote head was rechecked with `git ls-remote`. |

The Curator composite cannot be described as “landed on main.” Its audit identity is the base commit plus the accepted board handoff chain and retained worktree. The CocoaSkills implementation tasks must record the final accepted Curator composite provenance supplied at handoff, not silently substitute `origin/main`.

### 2.2 Board-state provenance

Curator story `STORY-260720-3plyvy` contained:

- 14 reviewed `done` implementation tasks;
- `TASK-260720-1nlmvv` in `development`;
- five backlog tasks: `TASK-260720-jrrgw9`, `TASK-260720-2qqq0w`, `TASK-260720-1pvfj5`, `TASK-260729-2kaopg`, and `TASK-260729-3jku56`.

CocoaSkills story `STORY-260720-1uv5gi` contained exactly 17 pre-existing implementation/verification tasks, all in `backlog`, plus this research task.

## 3. Accepted protocol surfaces

The parity contract has two layers:

- The schema-6 declarations, build-source framing, local `go-v1` metadata, receipt-v1, marker-v2, and their frozen expected bytes originate in the rc.4 candidate and remain byte-frozen in rc.5.
- The executable host boundary is the accepted rc.5 `manager-worker-v1` amendment. It distinguishes mandatory portable controls from six deferred hardened guarantees and adds the closed `rc5-native-control-inventory-v1` and `capability-evidence-v1` vocabularies.

| Surface | Binding artifacts |
| --- | --- |
| Schema-6 declaration | `schemas/v1/agent-skill-v6.schema.json`, `schemas/v1/csk-skill-v6.schema.json`, their generated cases under `conformance/v1/schema-cases/`, and `protocol/core.md` §4.2 |
| Build-root context and raw source identity | `protocol/core.md` §§3.1 and 8.1; `conformance/v1/vectors/build-drivers.json`; `conformance/v1/expected/build-driver/build-source.preimage.bin`, `build-source-sha256.txt`, `context_files.json`, and `context_sha256.txt` |
| Trusted Go identity and closed driver | Frozen identity/argv: `protocol/core.md` §8.2, `profiles/manager.md` §2.2, `decisions/0004-compile-only-build-drivers.md`, and `build-drivers.json`. Accepted execution amendment: rc.5 `protocol/core.md` §4.2.1, `profiles/manager.md` §2.2.1, `decisions/0006-portable-manager-worker-execution.md`, and `docs/portable-go-execution-policy.md`. |
| Canonical input, key, receipt, marker | `protocol/core.md` §§9–10; `schemas/v1/build-receipt-v1.schema.json`; `schemas/v1/install-marker-v2.schema.json`; `schemas/v1/conformance-claim-v2.schema.json`; expected CCJ/input/key/receipt/marker files under `conformance/v1/expected/build-driver/` |
| Planning, transaction, rollback, recovery | `profiles/manager.md` §§2.4–2.6; `conformance/v1/vectors/manager-lifecycle.json`; `cli/curator.md` §§58–231 |
| Activation, status, repair, GC | `profiles/manager.md` §§8 and 10; `cli/curator.md` §§128–231 |
| Security and platform trust | `SECURITY.md` compile-only, compiler-input, and protected-cache sections; `decisions/0004-compile-only-build-drivers.md`; negative clusters in `build-drivers.json` and `manager-lifecycle.json` |

These files are acceptable as reviewed candidate inputs, but neither rc.4 nor rc.5 is a published immutable release. Candidate tests must continue to use an explicit `CURATOR_CONFORMANCE_ROOT` and record its digest. The committed default pin remains on the released prior suite until `TASK-260720-25d05o` qualifies the chosen release and the manager-specific pin-audit tasks accept promotion.

## 4. Task-by-task parity map

Legend:

- **Accepted** means the Curator counterpart is present in the reviewer-accepted `TASK-260720-1ljev5` composite.
- **In flight** means it exists only in `TASK-260720-1nlmvv` after a changes-requested review.
- **Pending** means the Curator task remains backlog and is a handoff prerequisite, not reusable evidence.

| CocoaSkills task | Curator counterpart and state | Concrete Curator source and tests | Required Python/csk adaptation | Protocol surface |
| --- | --- | --- | --- | --- |
| `TASK-260720-z9j4c9` schema-v6 build model | `2g0e3b` **Accepted** | `internal/skillspec/{types.go,parse.go,build_test.go,conformance_test.go}`; `internal/skillcheck/{skillcheck.go,skillcheck_test.go}` | Extend `src/csk/skillspec.py` dataclasses/parser and `skillcheck.py`; preserve schemas 1–5 and both manifest names; reproduce static path/module/root validation without invoking Go. Tests: `tests/test_skillspec.py`, `test_skillcheck.py`, protocol cases. | Schema-6 schemas/cases; core §4.2 |
| `TASK-260720-3c0ss2` build-source/context boundary | `256kj1` and exclusion part of `11pfex` **Accepted** | `internal/buildsource/*`; `internal/whitelist/{whitelist.go,whitelist_test.go,conformance_test.go}`; `internal/install/install_test.go`; `internal/skillcheck/*` | Add a Python frozen-snapshot token and framed digest under `src/csk/builds/`; keep `hashing.content_sha256` marker-excluding; pass `build_roots` into whitelist/runtime exclusion in real and dry-run paths. | Core §§3.1, 8.1; source/context vectors |
| `TASK-260720-3j8pp5` Go toolchain identity | `6i3cya` **Accepted** | `internal/godriver/{session.go,fingerprint.go,identity.go}` plus `session_test.go`, `fingerprint_test.go`, Unix/Windows helpers | New `src/csk/builds/toolchain.py`; use direct executable resolution and subprocess calls, a clean environment, private telemetry state, frozen target/tuning, and byte-identical GOROOT framing. Current accepted Curator allowlist is Go family `1.25`, while protocol floor remains 1.23. | Core §8.2; manager §2.2; toolchain vectors |
| `TASK-260720-2dnqw2` canonical build metadata | `3mrm4z` and `4bd0it` **Accepted** | `internal/buildmeta/{models.go,codec.go,buildmeta_test.go}`; `internal/protocoljson/ccj.go`; `internal/marker/{marker.go,marker_v2_test.go}` | Create typed Python input/receipt/marker modules; reuse only the existing strict JSON loader/CCJ primitive; keep physical cache paths csk-specific and strict readers duplicate-key/noncanonical aware. | Receipt/marker/claim schemas; core §§9–10; expected CCJ/key/receipt/marker bytes |
| `TASK-260720-2g21eg` go-v1 compile driver | `1zntv0` **Accepted** after portable-worker review cycle 2 | `internal/godriver/{build.go,graph.go,executor.go,workerclient.go,workerserver.go,workerproto.go,controls_*.go}`; `build_test.go`, `build_conformance_test.go`, `graph_test.go`, `boundary_test.go`, worker/control tests, real fixture | Implement Python subprocess/worker boundary rather than porting Go internals. Preserve exact five argv forms, empty/fixed environment, graph rejection, native controls, staged-output validation, no output launch. The task brief’s rc.4 host-policy assumption must be retargeted to accepted rc.5 `manager-worker-v1`. | Decisions 0004 and 0006; manager §§2.2–2.2.1; driver/graph/host-policy vectors |
| `TASK-260720-2jfnz6` protected cache POSIX | `3pwg2w` **Accepted** | `internal/buildcache/{cache.go,publish.go,protection_unix.go}` and cache, publication, conformance, validation, Unix tests | New backend-neutral cache API plus POSIX backend using fd/rooted no-follow operations and ownership/mode/link checks. csk may choose a different physical namespace but must preserve logical key/receipt behavior. | Core §9; manager §2.4; protected-cache security boundary |
| `TASK-260720-8nxlgx` protected cache Windows | `3pwg2w` **Accepted** | `internal/buildcache/protection_windows.go`, `protection_windows_test.go`, `collect_windows_test.go` | Implement DACL/owner/reparse/file-ID/hard-link checks with Python `ctypes` or standard APIs; keep module import-safe elsewhere. Cross-compilation is insufficient: native Windows negative tests are required. | Same logical cache contract; Windows platform policy |
| `TASK-260720-z2z795` transaction engine | `1zl1cj` and `31nl14` **Accepted** | `internal/managerlock/*`; `internal/transaction/*`; `internal/staging/*`; extensive identity, durability, recovery, namespace, rollback, subprocess, Darwin, and Windows tests | Extend `locking.py` with canonical project/home lock hierarchy; add Python journal/transaction modules with deterministic target ordering, preimage/generation digests, crash recovery, reverse rollback, and consumer-last durability. | Manager §§2.5–2.6; lifecycle vectors |
| `TASK-260720-11yhth` command runtime activation | `11pfex` and `29hi1h` **Accepted** | `internal/closure/*`; `internal/runtimestore/{scripts.go,targets.go,*_test.go}`; staged helpers in `globalbins`, `adapters`, `envfiles`; closure and shim tests | Split script-runtime completeness checking from compiled-target activation in `shims.py`; point compiled shims at immutable artifacts; preserve existing csk project/global/user-bin conventions and mixed-command collision rules. | Core §4; manager §8; activation vectors |
| `TASK-260720-2x6mjn` side-effect-free planner | `3itlly` **Accepted** | `internal/install/{plan.go,private.go,stage.go,builddeps.go}`; `private_test.go`, `stage_test.go`, `revalidation_test.go`; snapshot locks | Introduce a pure Python planner and explicit read-only audit/registry paths. Do not reuse the current coarse `GlobalLock` routing for dry-run. Model every persistent write as a later staged target. | Manager §§2.4–2.5; compiler-free dry-run and lifecycle vectors |
| `TASK-260720-3t8nr3` project/hybrid transaction | `2284br` **Accepted** | `internal/install/{install.go,commit.go,generation.go,targets.go,atomicity/*}`; `internal/scopes/{stage.go,hybrid.go}`; adapter/env/global-bin staging; commit, ABA, atomicity, revalidation tests | Refactor `installer.py` from per-node/per-target mutation to plan → private builds → home-lock revalidation/publication → one journaled commit. Express project, hybrid, runtime, context, marker, shims, adapters, env, stale removals, and consumer ledger as transaction targets. | Manager §§2.5–2.6; project/hybrid commit/rollback vectors |
| `TASK-260720-g7kgox` global transaction | Global half of `2284br` **Accepted** | `internal/install/global.go`; `internal/globalbins/*`; `internal/adapters/*`; `internal/envfiles/*`; global cases in `install_test.go` and atomicity tests | Reuse the same planner/transaction implementation in `global_install.py`; retain csk’s global manifest and user-bin selection behavior but remove partial materialization. | Same lifecycle contract; global activation/status vectors |
| `TASK-260720-th0jdi` currentness/repair/GC | Marker and GC parts `4bd0it` + `1ljev5` **Accepted**; diagnostic/repair layer `1nlmvv` **In flight** | Accepted: `internal/marker/*`, `internal/buildcache/collect.go`, `internal/scopes/{gc.go,gc_*test.go}`, `cmd/curator/gc_test.go`. In flight: `cmd/curator/{builds.go,builds_test.go,status_test.go}`, `internal/install/{diagnostics.go,diagnostics_test.go}` and changes to install planning. | Implement only after the reworked Curator vocabulary/semantics are accepted. Extend both `status.py` and `global_install.py` status; make install/upgrade the repair path; keep GC conservative across corrupt/unknown markers, consumers, journals, and redirected roots. | Marker v2; manager §§2.6, 10; CLI status/repair/GC; lifecycle vectors |
| `TASK-260720-12r55p` shared v6 vector consumer | Existing package conformance tests **Accepted**, but integrated Curator `jrrgw9` **Pending**; literal rc.4 protocol `3ag6pi` **Blocked** | `internal/{skillspec,buildsource,buildmeta,buildcache,godriver,closure,whitelist,runtimestore}/...conformance_test.go`; `internal/interop/golden_test.go`; pending manager-wide assertions | Write independent Python assertions in `tests/test_protocol_conformance.py`; adapters may parse shared fixtures but must not duplicate product logic or copy Go code. Keep legacy rc.3 green and candidate input explicit. Before work, retarget the task to the owner-selected rc.4 or rc.5 suite; recommended rc.5 must include portable host-policy vectors while preserving frozen schema-6/local-driver bytes. | All schema-6 cases, build-driver/lifecycle vectors, expected frozen bytes, and chosen host-policy suite |
| `TASK-260720-akf5kh` user documentation | Protocol author/CLI docs `3lo9jc` **Accepted**; Curator product docs `2qqq0w` **Pending** | Candidate `README.md`, `SECURITY.md`, `cli/curator.md`, `profiles/manager.md`; accepted Curator composite `README.md`; pending final Curator docs | Adapt terminology and paths to csk (`README*`, `ARCHITECTURE*`, `SECURITY*`, authoring docs), including maintained Russian mirrors. Do not copy Curator CLI claims before final diagnostics/currentness behavior is accepted. | Authoring, CLI, security, lifecycle, activation, status/GC docs |
| `TASK-260720-3pemm6` cross-platform Go E2E | Platform-specific Curator tests are **Accepted** throughout packages; Curator `1pvfj5` and `jrrgw9` **Pending** | `internal/godriver/*_{darwin,windows,unix}_test.go`; buildcache Windows/POSIX tests; managerlock/transaction platform tests; runtimestore/shim tests; install/atomicity tests; Curator task gate evidence | Add a real vendored Go skill and black-box Python subprocess E2E. CI must run Ubuntu, macOS, and Windows across Python 3.11–3.14, set up the accepted Go family, accept an explicit candidate suite, and retain the old default pin. | Full driver/lifecycle/activation vectors |
| `TASK-260720-3s27te` integrated verification | Whole accepted composite plus `jrrgw9`, `2qqq0w`, `1pvfj5` once accepted | Entire Curator package/test graph and final platform evidence | Verification only: clean CocoaSkills worktree; full pytest, strict mypy, build/twine, diff check, candidate conformance, and three-OS CI. Route semantic failures back to owning tasks. | Every story criterion and required vector cluster |

All 17 pre-existing CocoaSkills tasks are represented exactly once above.

## 5. Adaptation and reuse boundaries

### Reuse directly

- Canonical JSON/CCJ rules and exact expected bytes.
- Schema files, generated positive/negative cases, fixtures, and vector names.
- Stable diagnostic codes only after `TASK-260720-1nlmvv` is accepted.
- Ordering rules: provider-first, unsigned UTF-8 lexical command order, deterministic target-class order, consumer last.
- Security invariants: audit before compiler; immutable/protected cache; no package-controlled args/env/output name; no output execution during install; compiler-free dry-run.
- Platform acceptance matrices and negative cases.

### Reimplement independently

- Python dataclasses/readers and error types.
- Subprocess/worker execution and resource-control binding.
- POSIX fd-rooted and Windows handle/DACL/reparse filesystem operations.
- Lock identity, journaling, durability, recovery, and rollback.
- csk-specific cache paths, global manifest, user-bin selection, adapter/environment materialization, and CLI rendering.

### Do not reuse

- Go package structure or copied Go algorithms.
- Curator physical cache/lock/journal path names.
- Test-only Curator worker dispatch or internal dependency injection.
- Unaccepted `TASK-260720-1nlmvv` diagnostics and status semantics.
- A working-tree candidate as if it were a published protocol release or manager pin.

## 6. Critical path and parallel order

### 6.1 External gate to CocoaSkills start

```text
TASK-260720-1nlmvv currentness/repair rework accepted
  ├─> TASK-260720-jrrgw9 Curator rc.4 conformance
  └─> TASK-260720-2qqq0w Curator compiled-build docs
          \                 /
           TASK-260720-1pvfj5 Curator cross-platform CI gate accepted
                    |
          +---------+---------+
          |                   |
  z9j4c9 schema model   z2z795 transaction engine
```

The two new Curator follow-ups, `TASK-260729-2kaopg` (global status) and `TASK-260729-3jku56` (repeated-install idempotence), were created from currentness review findings. Before `jrrgw9` or `1pvfj5` closes, the coordinator must either integrate their accepted behavior or explicitly prove they are outside the gate; otherwise CocoaSkills would inherit a known parity gap.

### 6.2 CocoaSkills implementation critical path

```text
z9j4c9
  ├─> 3c0ss2 ─┐
  └─> 3j8pp5 ─┴─> 2dnqw2 ─> 2jfnz6 ─> 8nxlgx ─> 11yhth ─┐
                 └─> 2g21eg ────────────────────> 2x6mjn ─┤
z2z795 ─────────────────────────────────────────> 2x6mjn ─┘
                                                          |
                       3t8nr3 -> g7kgox -> th0jdi
                                      ├─> 12r55p -> 3pemm6 ─┐
                                      └─> akf5kh ───────────┴─> 3s27te
```

Additional join: `12r55p` currently waits for literal rc.4 verification `TASK-260720-3ag6pi`; that dependency must either reach acceptance on a real rc.4 ref or be explicitly retargeted to the accepted rc.5 publication/verification path. `akf5kh` also waits for already accepted protocol docs `TASK-260720-3lo9jc`.

Recommended execution:

1. Start `z9j4c9` and `z2z795` in parallel after `1pvfj5` is `done`.
2. After the model, run `3c0ss2` and `3j8pp5` in parallel.
3. After both, run `2dnqw2` and `2g21eg` in parallel.
4. Keep cache backends serial: POSIX `2jfnz6`, then Windows `8nxlgx`.
5. Start `11yhth` after Windows cache; start `2x6mjn` once driver, Windows cache, and transaction engine are accepted.
6. Join both at `3t8nr3`, then serialize global integration and maintenance: `g7kgox` → `th0jdi`.
7. Run shared vectors and docs in parallel; then E2E; then integrated verification.

## 7. macOS, Windows, and Linux gates

| Gate class | macOS requirement | Windows requirement | Linux requirement |
| --- | --- | --- | --- |
| Static models/bytes | Native pytest/mypy for schema, framing, CCJ, receipt, marker | Import/path/case vectors; exact bytes must match | Portable parity in CI |
| Toolchain/driver | Native Go 1.25-family probe and real vendored build; process/resource controls; poisoned env negatives | Native executable/GOROOT identity, job/process controls, path/DACL/reparse behavior; real vendored build | Final real E2E; not a substitute for native macOS/Windows platform security |
| Cache | Native POSIX ownership/mode/no-follow/link/race/publication tests | Native DACL/owner/reparse/hard-link/file-ID/race/publication tests; cross-build alone does not qualify | POSIX portable tests and final E2E |
| Locks/transactions | Native file locking, Darwin case/Unicode namespace, fsync/rename/recovery/rollback | Native canonical identity, first-use case aliases, durability, replace/journal cleanup, recovery/rollback | Portable transaction/E2E coverage |
| Shims/activation | POSIX project/global argv and exit propagation | `.cmd` quoting, `%*`, `ERRORLEVEL`, injection rejection, project/global/user-bin paths | POSIX project/global launch |
| Status/repair/GC | Native protected-state inspection and conservative two-pass GC | Native protected boundary, reparse/redirect, lock serialization, repair and GC | Final status/repair/E2E behavior |
| Final story gate | Python 3.11–3.14 macOS CI plus strict mypy/build metadata | Python 3.11–3.14 Windows CI, no unexpected skip/xfail in task-owned cases | Python 3.11–3.14 Ubuntu CI; full real fixture |

Curator review history demonstrates why native Windows is mandatory: cross-compilation passed while later native runs found lock identity, DACL, journal durability, and GC containment defects. CocoaSkills should budget native Windows review cycles from the start.

## 8. Exact handoff prerequisites

Before the first CocoaSkills implementation task:

- [ ] A board owner resolves the stale rc.4 wording across the 17 CocoaSkills briefs. Recommended: authorize rc.5 supersession, preserve the frozen schema-6/local-driver byte contracts, and add the accepted `manager-worker-v1` policy/control vectors. Alternative: land and publish literal rc.4, which would diverge from the already accepted Curator driver boundary.
- [ ] `TASK-260720-1nlmvv` is reviewer-accepted, including production-reachable stable diagnostics, non-attributing input drift, atomic repair of corrupt cache state, and bounded cross-platform redaction.
- [ ] The coordinator resolves or explicitly scopes `TASK-260729-2kaopg` and `TASK-260729-3jku56` before Curator conformance/CI acceptance.
- [ ] `TASK-260720-jrrgw9`, `TASK-260720-2qqq0w`, and `TASK-260720-1pvfj5` are reviewer-accepted; the last is the actual hard blocker of both CocoaSkills roots.
- [ ] The Curator handoff records the exact accepted composite identity: base commit, task outcome chain, worktree/archive digest, candidate conformance-root digest, and macOS/Windows gate evidence.
- [ ] `/Users/iv/Developer/Wildberries/cocoaskills` is still clean, is fast-forwarded from `edce8816…` to `6fc2fd97…` (or the then-current verified `origin/main`), and each task worktree records that base before edits.
- [ ] The final accepted Go-family allowlist and native tuning contract are copied from protocol/Curator handoff evidence. Current Curator code allows `1.25`; “1.23+” is only the protocol floor.
- [ ] A macOS native runner, a native Windows runner/SSH alias `win`, and final Ubuntu CI are available; Windows security/durability tasks must not be accepted from cross-compile evidence alone.
- [ ] Candidate conformance input remains explicit and immutable; no committed suite pin moves merely because a working-tree candidate passes.

Before `TASK-260720-12r55p`:

- [ ] An authorized curator-spec commit/ref exists for the owner-selected release line.
- [ ] Literal rc.4 path: `TASK-260720-3ag6pi` is rerun against that real clean ref without virtual-index or status wrappers and reaches reviewer acceptance.
- [ ] Recommended rc.5 path: the CocoaSkills brief/dependency is retargeted to the landed accepted rc.5 snapshot; conformance covers frozen schema-6/local-`go-v1` bytes plus `manager-worker-v1`, native-control inventory, capability evidence, and failure-boundary vectors.
- [ ] `CURATOR_CONFORMANCE_ROOT`, repository SHA, complete manifest/suite digest, and previous released default pin are recorded.

Before the final CocoaSkills gate:

- [ ] Every predecessor is accepted and landed in the CocoaSkills integration base.
- [ ] Linux, macOS, and Windows real-fixture CI is green across Python 3.11–3.14 with no unexpected task-owned skip/xfail.
- [ ] Full pytest, strict mypy, `python -m build`, `python -m twine check dist/*`, candidate conformance, and `git diff --check` run as standalone gates with real exit codes.
- [ ] Release/pin promotion remains owned by `TASK-260720-25d05o` and `TASK-260720-1utsx8`; candidate qualification alone must not fabricate release evidence.

## 9. Risks and mitigations

| Risk | Evidence | Required mitigation |
| --- | --- | --- |
| CocoaSkills briefs say rc.4 while Curator implements rc.5 execution semantics | `TASK-260728-zb2s4z` and `TASK-260720-1zntv0` are accepted; `TASK-260720-3ag6pi` remains blocked and no rc.4 ref exists | Board owner explicitly retargets tasks/dependencies to rc.5 (recommended) or authorizes a literal rc.4 publication path before implementation |
| Treating an uncommitted composite as a release | Curator and curator-spec remote `main` still point to rc.3-era commits; `3ag6pi` reviewer blocked publication acceptance | Require exact task artifact/archive digest now and a real immutable ref before released-suite claims |
| Copying Curator internals into Python | Story explicitly requires an independent consumer | Share only schemas, vectors, bytes, codes, and black-box cases |
| Dry-run mutates registry/cache/state | Existing `installer.py` performs `_check_audit_registries` before dry-run return; accepted research already flagged this | Build an explicit read-only planner/audit path and whole-tree before/after tests |
| Existing coarse lock is the wrong lifecycle | `GlobalLock` covers broad existing CLI operations but lacks the accepted lock hierarchy, journal, and revalidation semantics | Implement project locks + home mutation lock + durable generic target journal before installer integration |
| Partial per-node mutation breaks rollback | Current `installer.py` materializes consumers/runtime/context incrementally | Make every live surface a staged transaction target; consumer last |
| Physical cache trust confused with receipt consistency | Protocol says hashes are consistency metadata, not provenance | Validate owner-protected boundary before parsing/adopting receipt/artifact bytes |
| Native Windows bugs hidden by cross-build | Curator review found multiple native-only DACL, identity, durability, and GC issues | Require native Windows positive/negative/race evidence on the owning tasks |
| In-flight status vocabulary copied too early | `1nlmvv` review found unreachable codes, misclassification, repair refusal, and path leaks | Wait for accepted production-path semantics and consume its final code list as a handoff artifact |
| Slow/space-heavy verification destabilizes local gates | This research run’s full Curator test exhausted temp storage during `internal/install` | Preflight free space, use task-local temp roots, preserve standalone exits, and keep CI as the authoritative full platform gate |

## 10. Fact-check record

Commands were run as standalone processes; no checklist claim is based on an unrun or green-assumed command.

| Command | Exit | Finding |
| --- | ---: | --- |
| `git ls-remote origin refs/heads/main` in Curator | 0 | `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8` |
| `git ls-remote origin refs/heads/main` in curator-spec | 0 | `57c1f56846d221ecc55786bd3c2467ec32f11730` |
| `git ls-remote origin refs/heads/main` in CocoaSkills | 0 | `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12` |
| `go test ./...` in accepted Curator `TASK-260720-1ljev5` worktree | **143, failed/terminated** | Multiple packages printed `ok`, but `internal/install` failed with `no space left on device`; the still-running command was terminated after the gate was irrecoverably red. This is not a passing repository-wide gate. |
| `df -h` after the failure | 0 | Data volume reported 100% capacity with 8.4 GiB nominally available, consistent with the temp-write failure and unsuitable for another full local gate. |
| Initial 17-ID coverage assertion | **1, failed as written** | The first document draft used abbreviated IDs in the table and the exact-ID assertion reported `missing TASK-260720-11yhth`. The table was corrected to full IDs. |
| Revised 17-ID coverage assertion | 0 | Every one of the 17 exact CocoaSkills task IDs is present. |
| Exact parity-table row-count assertion | 0 | Exactly 17 mapped task rows. |
| Board-reference existence assertion | 0 | Every cited board evidence file exists, including the rc.5 portable amendment review. |
| Initial trailing-whitespace assertion | **1, failed as written** | It found two deliberate Markdown hard-break spaces on the metadata lines. They were replaced with blank lines. |
| Revised trailing-whitespace assertion | 0 | No trailing whitespace remains. |
| `git status --short -- cmd internal go.mod go.sum README.md Makefile .github` in Curator root | 0 | Empty output; no Curator product/config surface was changed by this research task. |
| `git status --short` in CocoaSkills | 0 | Empty output; CocoaSkills remained clean. |
| `git diff --cached --quiet` in Curator and CocoaSkills | 0 for each | No staged changes. |

Prior green evidence is therefore cited as **reviewed historical evidence**, not as a green rerun in this research task:

- `TASK-260720-1ljev5_review-verdict-cycle-3.md` — accepted cache-GC composite with macOS, cross-platform, and native Windows evidence.
- `TASK-260720-1zntv0_portable-review-cycle-2-verdict.md` — accepted portable worker/driver boundary.
- `TASK-260720-31nl14_review-cycle-11-verdict.md` — accepted durable journal with native Windows evidence.
- `TASK-260720-3ag6pi_reviewer-verdict.md` — protocol content validation accepted but publication/ref gate blocked.
- `TASK-260728-zb2s4z_review-cycle-2-verdict.md` — accepted rc.5 portable execution amendment and exact candidate identity.
- `TASK-260720-1nlmvv_review-verdict-cycle-1.md` — exact in-flight currentness/repair semantic defects.

## 11. Recommendation

Do not collapse or renumber the 17 CocoaSkills tasks. The decomposition matches Curator’s actual subsystem boundaries and correctly isolates the two platform cache backends, transaction infrastructure, project/global integration, maintenance, vectors, docs, E2E, and final verification.

At Curator gate closure, hand CocoaSkills two immutable packets:

1. the accepted Curator composite provenance plus task-to-file/test inventory and native platform evidence;
2. the owner-selected candidate protocol tree/ref and complete conformance digest, explicitly reconciling frozen schema-6/local-driver bytes with the rc.5 portable execution amendment.

Then start only `z9j4c9` and `z2z795` in parallel. This preserves the no-bypass dependency rule while removing reconnaissance latency from the first implementation cycle.

## 12. References

### Board outcomes

- `.task-board/.resources/STORY-260720-1uv5gi/STORY-260720-1uv5gi_17-task-audit.md`
- `.task-board/.resources/STORY-260720-1uv5gi/STORY-260720-1uv5gi_decomposition-plan.md`
- `.task-board/.resources/STORY-260720-1uv5gi/STORY-260720-1uv5gi_planning-validation.md`
- `.task-board/.resources/TASK-260720-1ljev5/TASK-260720-1ljev5_review-verdict-cycle-3.md`
- `.task-board/.resources/TASK-260720-1zntv0/TASK-260720-1zntv0_portable-review-cycle-2-verdict.md`
- `.task-board/.resources/TASK-260720-31nl14/TASK-260720-31nl14_review-cycle-11-verdict.md`
- `.task-board/.resources/TASK-260720-3ag6pi/TASK-260720-3ag6pi_reviewer-verdict.md`
- `.task-board/.resources/TASK-260728-zb2s4z/TASK-260728-zb2s4z_review-cycle-2-verdict.md`
- `.task-board/.resources/TASK-260720-1nlmvv/TASK-260720-1nlmvv_review-verdict-cycle-1.md`

### Source roots

- Curator accepted composite: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1ljev5/worktree`
- Curator currentness/repair rework: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1nlmvv/worktree`
- Protocol candidate: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3ag6pi/worktree`
- Accepted rc.5 portable amendment: `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-zb2s4z/curator-spec-worktree`
- CocoaSkills baseline: `/Users/iv/Developer/Wildberries/cocoaskills`
