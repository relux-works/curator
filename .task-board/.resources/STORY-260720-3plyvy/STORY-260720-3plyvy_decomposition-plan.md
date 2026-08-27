# Curator Go build driver — development decomposition

**Board story:** `STORY-260720-3plyvy`  
**Implementation baseline:** Curator `origin/main` at `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`  
**Accepted contract:** `TASK-260720-poa3ze_compile-only-build-drivers.md`, SHA-256 `6308d99d8bdad4445841bc1cbd230cadbed0020012d0e9d38d877b413348f681`  
**Protocol implementation gate:** `TASK-260720-3ag6pi`

## Architecture decisions carried into the tasks

1. Curator selects Go from an absolute `CURATOR_GO`, then `GOROOT/bin/go`, then `runtime.GOROOT()/bin/go`. It never searches user `PATH`, project shims, package content, or repository-local executables.
2. Portable cache identity stays logical. Curator's physical manager-local layout is `<curator-home>/cache/build/go-v1/<64-hex-key>/...`, protected as implementation trusted computing base state.
3. The all-file, length-framed `curator-build-source-v1` identity includes the root package marker and remains separate from the existing marker-excluding installed-content hash.
4. Schema and root checks, static context exclusion, skill check, closure, collision, audit, registry, attestation, and moved-tag gates precede package-aware Go commands.
5. Dry-run may use only removable operation-private toolchain probe state. It acquires no locks, never calls `go list` or `go build`, and mutates no persistent cache or installation state.
6. Real operations serialize per project, build misses privately, then enter the manager-home critical section for recovery, revalidation, immutable cache publication, journaled target swaps, and consumer-last commit.
7. Generated executables remain untrusted: installation verifies but never launches them. Managed Unix and Windows shims launch the marker-selected immutable artifact later and preserve arguments and exit status.

## Development task map

| Phase | Task | Atomic deliverable | Primary ownership |
|---:|---|---|---|
| 1 | `TASK-260720-2g0e3b` | Strict schema v6 manifest parser | `internal/skillspec` |
| 1 | `TASK-260720-256kj1` | Immutable build-source identity | new `internal/buildsource`, snapshot validation |
| 1 | `TASK-260720-1zl1cj` | Cross-platform operation locks | new manager-state locking package |
| 2 | `TASK-260720-11pfex` | Build activation and build-root context exclusion | closure, whitelist, skillcheck |
| 2 | `TASK-260720-3mrm4z` | Canonical build input and receipt codecs | protocol JSON and build metadata |
| 3 | `TASK-260720-6i3cya` | Trusted Go toolchain session | new `internal/godriver` session layer |
| 3 | `TASK-260720-3pwg2w` | Protected immutable build cache | new `internal/buildcache` storage |
| 3 | `TASK-260720-31nl14` | Durable target journal and recovery engine | new `internal/transaction` |
| 4 | `TASK-260720-1zntv0` | Fixed go-v1 preflight and private build | `internal/godriver` package-aware layer |
| 4 | `TASK-260720-4bd0it` | Marker v2 and compiled currentness | `internal/marker` |
| 4 | `TASK-260720-29hi1h` | Compiled runtime targets and shims | `internal/runtimestore`, shim integrations |
| 5 | `TASK-260720-3itlly` | Read-only planning and private build staging | `internal/install` orchestration |
| 6 | `TASK-260720-2284br` | Atomic project/global/hybrid commit | install plus target adapters |
| 7 | `TASK-260720-1ljev5` | Safe build-cache collection | `internal/scopes` maintenance |
| 8 | `TASK-260720-1nlmvv` | CLI currentness, diagnostics, and repair | `cmd/curator`, result models |
| 9 | `TASK-260720-2qqq0w` | Compiled-command authoring and operator docs | README and repository docs |
| 9 | `TASK-260720-jrrgw9` | rc.4 end-to-end conformance evidence | interop and lifecycle tests |
| 10 | `TASK-260720-1pvfj5` | Cross-platform release-quality CI gates | CI and quality configuration |

The board contains the authoritative 42 prerequisite links and critical path. The three phase-1 tasks are each blocked by upstream integrated protocol verification `TASK-260720-3ag6pi`; no Curator implementation task should infer or privately clone unfinished portable vectors.

## Story completeness matrix

| Story requirement or acceptance criterion | Owning task(s) |
|---|---|
| Parse valid schema v6 and reject unsafe declarations | `TASK-260720-2g0e3b` |
| Skill check, closure activation, command collisions, and static build-root exclusion | `TASK-260720-11pfex` |
| Immutable raw snapshot and source digest before cache lookup | `TASK-260720-256kj1` |
| Cache identity includes source, build policy, toolchain, and native target | `TASK-260720-3mrm4z`, `TASK-260720-6i3cya` |
| Missing or incompatible Go fails clearly; environment is isolated | `TASK-260720-6i3cya`, `TASK-260720-1nlmvv` |
| Readonly vendored modules, fixed commands, source-aware graph checks, no execution during install | `TASK-260720-1zntv0` |
| Protected manager-local cache, strict receipts, corrupt or stale artifact rebuild | `TASK-260720-3pwg2w`, `TASK-260720-4bd0it`, `TASK-260720-3itlly` |
| Cross-platform locks, durable journal, crash recovery, reverse rollback | `TASK-260720-1zl1cj`, `TASK-260720-31nl14`, `TASK-260720-2284br` |
| Marker v1 compatibility, marker v2 build state, exact currentness | `TASK-260720-4bd0it` |
| Runtime-store validation and managed shims forward arguments and exit status | `TASK-260720-29hi1h` |
| Dry-run performs no build, lock, cache, or install mutation | `TASK-260720-1zl1cj`, `TASK-260720-3itlly`, `TASK-260720-1nlmvv` |
| Build failures preserve the previous project/global/hybrid installation | `TASK-260720-3itlly`, `TASK-260720-2284br` |
| Atomic cache publication and consumer-last install commit | `TASK-260720-2284br` |
| Marker- and journal-aware protected cache GC | `TASK-260720-1ljev5` |
| Human and JSON diagnostics plus install/upgrade repair | `TASK-260720-1nlmvv` |
| README and authoring/operator guidance | `TASK-260720-2qqq0w` |
| Fixtures, unit/integration/fault/concurrency/launch tests | `TASK-260720-jrrgw9` |
| Existing tests plus race, vet, formatting, lint, and Linux/macOS/Windows behavior | `TASK-260720-1pvfj5` |

## Baseline findings and risk controls

- The shared checkout is `c06aa1a15e4093410a686ff0ce4f579fba59dec1`, not current `origin/main`. Analysis therefore used a detached worktree at the required baseline and left the user's branch untouched.
- On that baseline, installation updates the consumer ledger before all materialization steps. The atomic commit task explicitly moves this target last.
- The existing runtime store reuses an existing directory without validating every required runtime root and script entry. The shim/runtime task owns validation and repair.
- Project, global, hybrid, shim, adapter, env-file, and consumer writes are piecemeal and have no manager-home lock or durable journal. Lock, journal, staging, and atomic-commit work are separated into four prerequisite-linked tasks.
- Existing Go dry-run behavior is substantially read-only; new tests must preserve that property while proving compiler and lock absence for compiled commands.
- Current CI consumes the rc.3 suite and has no explicit race gate. The CI task may move to rc.4 only after the upstream reviewed immutable commit exists.
- The accepted contract resolves the product and security choices needed for implementation. No additional human clarification or research blocker was found; the upstream verification task is the only intentional external gate.
- System Graphviz is currently unusable because Homebrew `dot` cannot load `libltdl.7.dylib`. PlantUML 1.2026.6 with Smetana validates and renders the attached diagrams, so this is a local tooling anomaly rather than a product blocker.

## Planning artifacts and verification

- Canonical task-board plan: `.planning/260720_061939_story-260720-3plyvy.md` — 18 tasks in 10 phases.
- Component ownership source: `diagrams/plantuml/component/STORY-260720-3plyvy_component-map.puml`.
- Install lifecycle source: `diagrams/plantuml/activity/STORY-260720-3plyvy_install-lifecycle.puml`.
- Rendered SVG and PNG views are under `diagrams/artefacts/plantuml/`.
- PlantUML `-checkonly`, SVG render, PNG render, and visual inspection passed for both sources.
- Detached `origin/main` baseline passed `go test ./...`, `go test -race ./...`, and `go vet ./...` after initializing the testing-tools submodule.

Each task has one owned deliverable, a scoped description, explicit acceptance criteria, and three task-specific checklist items. Implementation code is intentionally outside this architecture handoff.
