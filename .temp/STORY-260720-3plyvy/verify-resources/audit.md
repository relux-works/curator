# Curator Go build driver — decomposition audit 2

**Story:** `STORY-260720-3plyvy` — Curator Go build driver  
**Role:** solution architect  
**Audit date:** 2026-07-20  
**Curator baseline:** `origin/main` at `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`  
**Accepted contract:** `TASK-260720-poa3ze_compile-only-build-drivers.md`, SHA-256 `6308d99d8bdad4445841bc1cbd230cadbed0020012d0e9d38d877b413348f681`  
**Protocol implementation gate:** `TASK-260720-3ag6pi` — Verify schema v6 protocol release candidate

## Audit verdict

The existing 18-task decomposition remains the correct cardinality. It covers the accepted Curator contract in 10 acyclic phases with 42 explicit prerequisite links. No duplicate, clarification, or additional research task is justified. Six task briefs had evidenced development-readiness defects; their existing descriptions, scopes, acceptance criteria, or checklists were corrected in place.

The story-local dependency graph remains unchanged. During this audit, the parallel interoperability audit added the correct cross-story edge from candidate Curator CI to manager-release qualification: `TASK-260720-3pvihp` is now blocked by `TASK-260720-1pvfj5`. The official released-suite pin can still move only after protocol-release qualification. Encoding `TASK-260720-25d05o` as a direct prerequisite of candidate CI would recreate the manager/spec release cycle. The downstream `TASK-260720-38l1sy` join is correctly blocked by both `TASK-260720-1pvfj5` and `TASK-260720-25d05o`.

## Evidence-backed corrections

1. `TASK-260720-1zl1cj` — Add cross-platform manager operation locks: fixed package ownership to new `internal/managerlock`; `internal/transaction` is explicitly reserved for `TASK-260720-31nl14`.
2. `TASK-260720-3mrm4z` — Canonicalize build inputs and receipts: fixed the ambiguous `internal/buildcache or smaller package` ownership. It now owns reusable CCJ-1 work in `internal/protocoljson` plus logical models/codecs in new `internal/buildmeta`; filesystem/protection remains solely with `TASK-260720-3pwg2w`.
3. `TASK-260720-29hi1h` — Launch compiled commands through managed shims: changed the deliverable to typed targets and staged runtime/shim outputs. It may emit desired/removal targets but cannot mutate live project/global paths; `TASK-260720-2284br` owns the transaction.
4. `TASK-260720-3itlly` — Stage build plans before installation mutation: made the boundary explicit—immutable plans and private staged results only; live mutation adapters stay downstream.
5. `TASK-260720-1zntv0` — Implement go-v1 preflight and artifact build: added the accepted contract's missing executable gates for a read-only frozen source/vendor tree, operation-private writable roots, network/module/VCS denial, deadlines, bounded output/artifact/disk, memory/process controls where supported, and unexpected-child rejection. A fourth checklist item makes these platform gates non-optional.
6. `TASK-260720-1pvfj5` — Enforce cross-platform compiled-build CI gates: removed the contradiction between its README and release-order note. This task now owns candidate-suite and Go quality evidence without advancing the committed released-suite pin, and its accepted handoff gates manager-release qualification `TASK-260720-3pvihp`. `TASK-260720-38l1sy` owns the official pin promotion/audit after `TASK-260720-25d05o` qualifies the published protocol release.

## Atomic ownership map

| Task | Atomic deliverable | Primary ownership |
|---|---|---|
| `TASK-260720-2g0e3b` — Parse and validate schema v6 build manifests | Strict schema-v6 domain/parser | `internal/skillspec` |
| `TASK-260720-256kj1` — Implement immutable build-source identity | Frozen raw-snapshot identity and repair | new `internal/buildsource`, focused `internal/snapshot` hooks |
| `TASK-260720-1zl1cj` — Add cross-platform manager operation locks | Cross-process lock primitives | new `internal/managerlock` |
| `TASK-260720-11pfex` — Activate build commands and exclude build roots | Activation, collision, context exclusion, author warnings | `internal/closure`, `internal/whitelist`, `internal/skillcheck` |
| `TASK-260720-3mrm4z` — Canonicalize build inputs and receipts | Portable logical models and strict CCJ-1 codecs | `internal/buildmeta`, `internal/protocoljson` |
| `TASK-260720-6i3cya` — Establish a trusted Go toolchain session | Package-independent trusted Go identity/session | `internal/godriver` session layer |
| `TASK-260720-3pwg2w` — Protect and publish immutable build cache entries | Protected filesystem cache backend | `internal/buildcache` |
| `TASK-260720-31nl14` — Add a durable install transaction journal | Generic durable journal/recovery engine | `internal/transaction` |
| `TASK-260720-1zntv0` — Implement go-v1 preflight and artifact build | Source-aware fixed driver plus host policy | `internal/godriver` build/platform layer |
| `TASK-260720-4bd0it` — Implement install marker v2 and build currentness | Marker compatibility and build currentness | `internal/marker` |
| `TASK-260720-29hi1h` — Launch compiled commands through managed shims | Typed staged runtime and shim targets | `internal/runtimestore`, focused shim/global-bin helpers |
| `TASK-260720-3itlly` — Stage build plans before installation mutation | Read-only plan and private build staging | planning/staging layer in `internal/install` |
| `TASK-260720-2284br` — Commit installations atomically across scopes | Live cache publication and cross-scope target transaction | `internal/install` entrypoints plus target adapters |
| `TASK-260720-1ljev5` — Collect compiled artifacts safely | Locked marker/journal-aware build GC | `internal/scopes`, focused cache helpers |
| `TASK-260720-1nlmvv` — Expose build diagnostics and repair behavior | CLI status/currentness/repair presentation | `cmd/curator`, result models |
| `TASK-260720-2qqq0w` — Document Curator compiled-build authoring | Curator author/operator documentation | `README.md`, maintained Curator docs |
| `TASK-260720-jrrgw9` — Verify rc.4 build-driver conformance end to end | Integrated authoritative-vector evidence | `internal/interop`, test-only fixtures/helpers |
| `TASK-260720-1pvfj5` — Enforce cross-platform compiled-build CI gates | Candidate cross-platform and Go quality gates | `.github/workflows/ci.yml`, `Makefile`/quality config |

Sequential same-package work is intentional: `TASK-260720-6i3cya` establishes the package-independent `internal/godriver` session before `TASK-260720-1zntv0` adds source-aware execution; `TASK-260720-3itlly` produces immutable plan/staging values before `TASK-260720-2284br` integrates live mutation. Their prerequisite links prevent conflicting parallel ownership.

## Accepted-contract coverage

| Contract surface | Owning task(s) |
|---|---|
| Schema 6, strict build shape, build-root/module semantics, schemas 1–5 compatibility | `TASK-260720-2g0e3b` |
| Closure activation, narrowing, collisions, context/runtime exclusion, skill warnings | `TASK-260720-11pfex` |
| All-file length-framed raw-source identity before cache lookup; snapshot mutation detection | `TASK-260720-256kj1` |
| Exact CCJ-1 input, cache key, receipt bytes/hash, target/toolchain/policy identity | `TASK-260720-3mrm4z` |
| Trusted Go selection, telemetry-off private probe, native target, release-family allowlist, toolchain digest | `TASK-260720-6i3cya` |
| Vendor-only fixed `go list`/`go build`, graph/directive/native-input rejection, read-only source, denied network/process escape, bounded host resources | `TASK-260720-1zntv0` |
| Protected manager-local immutable cache, ownership/ACL/link checks, forged/corrupt/stale miss behavior | `TASK-260720-3pwg2w` |
| Marker-v1 compatibility, marker-v2 build state, exact compiled currentness | `TASK-260720-4bd0it` |
| Built-command staged shims, runtime repair, forwarded arguments and exact exit/signal behavior | `TASK-260720-29hi1h` |
| All validation/trust gates before build; compiler-free and persistent-mutation-free dry-run; build-two failure isolation | `TASK-260720-3itlly` |
| Project/home locks, durable journal, shared-state revalidation, cache publication, reverse rollback, consumer-last atomic commit | `TASK-260720-1zl1cj`, `TASK-260720-31nl14`, `TASK-260720-2284br` |
| Marker/journal-aware protected cache GC | `TASK-260720-1ljev5` |
| Stable human/JSON diagnostics, status check, install/upgrade repair | `TASK-260720-1nlmvv` |
| Complete author/operator docs without unlanded release claims | `TASK-260720-2qqq0w` |
| Shared positive/negative vectors, fixtures, unit/integration/fault/race/launch evidence | `TASK-260720-jrrgw9` |
| Linux/macOS/Windows candidate suite, race, vet, gofmt, lint, and later qualified released-pin audit | `TASK-260720-1pvfj5`, downstream `TASK-260720-38l1sy` |

## Dependency audit

- 18 unique child tasks, 10 phases, 42 `blocked_by` links, no cycle reported by the canonical planner.
- Each phase-1 Curator foundation task is blocked by `TASK-260720-3ag6pi`; no task can privately invent unfinished rc.4 schemas or vectors.
- The critical path remains `TASK-260720-256kj1` → `TASK-260720-3mrm4z` → `TASK-260720-3pwg2w` → `TASK-260720-29hi1h` → `TASK-260720-3itlly` → `TASK-260720-2284br` → `TASK-260720-1ljev5` → `TASK-260720-1nlmvv` → `TASK-260720-2qqq0w` → `TASK-260720-1pvfj5`.
- `TASK-260720-jrrgw9` gates the independent Curator shared-suite consumer `TASK-260720-1673lr`; test ownership is not duplicated.
- `TASK-260720-1pvfj5` gates pre-release manager qualification `TASK-260720-3pvihp` and the later Curator pin audit `TASK-260720-38l1sy`; `TASK-260720-25d05o` independently gates that same released-pin audit. This is the correct candidate-evidence path plus two-evidence release join and avoids a release cycle.
- No redundant transitive prerequisite was added. Every existing link corresponds to a consumed model, API, staged artifact, or evidence gate.

## New task-scoped diagrams

- `TASK-260720-1zntv0_go-v1-execution-boundary.puml` and `.svg`: fixed process graph, read-only inputs, operation-private writable roots, denied surfaces, and supported host controls. Attached as outcomes to `TASK-260720-1zntv0`.
- `TASK-260720-1pvfj5_candidate-release-ci-gates.puml` and `.svg`: candidate evidence versus manager/spec/protocol release evidence and official Curator pin promotion. Attached as outcomes to `TASK-260720-1pvfj5`.

Both sources pass PlantUML 1.2026.6 `-checkonly`, render to valid SVG and PNG, and were visually inspected. The first uses Smetana because the installed Graphviz binary cannot load `libltdl.7.dylib`; this local tool anomaly does not block the artifacts.

## Verification

- `task-board q 'list(type=task, parent=STORY-260720-3plyvy) { full }'`: all 18 tasks have non-placeholder descriptions, scopes, acceptance criteria, and task-specific checklists.
- `task-board q 'plan(STORY-260720-3plyvy, mode=children)'`: 18 elements, 10 phases, no cycle.
- `task-board plan STORY-260720-3plyvy --save`: canonical snapshot `.planning/260720_065314_story-260720-3plyvy.md`.
- `task-board validate`: no story-local broken link or orphan resource; the same 12 legacy `EPIC-260712-*` prose dependency strings and one unrelated orphan `TASK-260713-7a9c1e/review.md` remain outside this story.
- Baseline inspection used `git show origin/main` for package/workflow ownership. `origin/main` remains `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`; no product code, tests, documentation, or configuration was changed by this planning audit.

No unresolved product, security, platform, or architecture decision remains. Development stays intentionally blocked only on the accepted upstream protocol implementation/verification gate.
