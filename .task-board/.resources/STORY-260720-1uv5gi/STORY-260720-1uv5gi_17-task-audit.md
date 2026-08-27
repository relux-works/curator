# STORY-260720-1uv5gi — 17-task decomposition audit

Date: 2026-07-20

## Evidence baseline

- Accepted contract: `TASK-260720-poa3ze_compile-only-build-drivers.md`, SHA-256 `6308d99d8bdad4445841bc1cbd230cadbed0020012d0e9d38d877b413348f681`.
- csk repository: `/Users/iv/Developer/intranet/cocoaskills`.
- Remote `origin/main`: `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`, verified again with `git ls-remote`.
- Local `main`: clean at `edce8816dda44bb121d661b7c4dea942558ce408`, two commits behind. Every implementation task retains the mandatory clean fast-forward-before-worktree checklist gate.

## Audit verdict

The existing 17 tasks are retained. No task was duplicated, closed, split, or added. The decomposition covers the accepted schema-v6 contract, resolves to an acyclic 12-phase plan, and leaves every implementation surface with one primary owner and executable acceptance gates.

Two evidenced wording defects were corrected:

1. `TASK-260720-12r55p` — Consume shared schema v6 vectors previously implied that a landed candidate rc.4 commit could advance the repository's committed protocol-suite pin. Its scope and AC now require the exact candidate suite through `CURATOR_CONFORMANCE_ROOT`, record the suite SHA, and preserve the last qualified released pin.
2. `TASK-260720-3pemm6` — Add cross-platform Go build end-to-end tests previously required an “actual immutable rc.4 conformance pin” while its release-order note assigned that pin to downstream `TASK-260720-1utsx8`. Its scope and AC now limit CI changes to Go/platform test execution, use the exact candidate suite out of tree, and leave the committed pin unchanged until release qualification.

This correction preserves the established release sequence: candidate verification by `TASK-260720-3ag6pi`, csk candidate consumption and platform evidence, protocol release qualification by `TASK-260720-25d05o`, then the sole csk pin owner `TASK-260720-1utsx8`.

## Atomic ownership and dependency audit

| Layer | Tasks | Ownership result |
|---|---|---|
| Manifest and transaction foundations | `TASK-260720-z9j4c9`, `TASK-260720-z2z795` | Parallel, disjoint primary modules: schema/static validation versus locks/journals/recovery. |
| Source and toolchain identity | `TASK-260720-3c0ss2`, `TASK-260720-3j8pp5` | Parallel, disjoint framed identities and platform-safe validation boundaries. |
| Portable metadata and compiler | `TASK-260720-2dnqw2`, `TASK-260720-2g21eg` | Parallel typed metadata/marker ownership versus fixed process execution and graph validation. |
| Protected cache | `TASK-260720-2jfnz6`, `TASK-260720-8nxlgx` | Portable interface plus POSIX backend precedes the Windows DACL/reparse backend. |
| Activation and planning | `TASK-260720-11yhth`, `TASK-260720-2x6mjn` | Parallel shim/runtime ownership versus audit-first planning, read-only registry paths, and dry-run routing. |
| Installation | `TASK-260720-3t8nr3`, `TASK-260720-g7kgox` | Project/hybrid integration precedes global integration; shared mutable files are not edited concurrently. |
| Maintenance | `TASK-260720-th0jdi` | One owner for marker-v2 currentness, reinstall repair, status, journals, and locked GC. |
| Evidence and docs | `TASK-260720-12r55p`, `TASK-260720-akf5kh`, `TASK-260720-3pemm6`, `TASK-260720-3s27te` | Lower-level shared vectors, user docs, real cross-platform E2E, and integrated verification remain distinct deliverables. |

The canonical plan still contains 17 elements, 12 phases, and this 12-task critical path:

`TASK-260720-z9j4c9 -> TASK-260720-3c0ss2 -> TASK-260720-2dnqw2 -> TASK-260720-2jfnz6 -> TASK-260720-8nxlgx -> TASK-260720-11yhth -> TASK-260720-3t8nr3 -> TASK-260720-g7kgox -> TASK-260720-th0jdi -> TASK-260720-12r55p -> TASK-260720-3pemm6 -> TASK-260720-3s27te`.

The similarly named interoperability task `TASK-260720-31zeo2` is not a duplicate of `TASK-260720-12r55p`: the story task implements the Python adapters and low-level schema/build/lifecycle vector assertions, while the downstream interoperability task enforces exact once-only external-suite case coverage through real csk flows after both managers and the suite contract exist. `TASK-260720-1utsx8` separately owns the released-suite pin audit.

## Accepted-contract coverage

| Contract area | Owning tasks |
|---|---|
| Strict schema 6, both manifest names, closed `go-v1`, build-root semantics, schemas 1–5 compatibility | `TASK-260720-z9j4c9`, `TASK-260720-12r55p` |
| Raw build-source identity and build-root exclusion from context/runtime | `TASK-260720-3c0ss2`, `TASK-260720-3t8nr3` |
| Trusted Go selection, private telemetry-off state, native target, exact toolchain fingerprint | `TASK-260720-3j8pp5` |
| CCJ-1 input/key, receipt v1, marker v1/v2 compatibility and exact canonical bytes | `TASK-260720-2dnqw2` |
| Five fixed Go argv forms, empty/fixed environment, vendored graph checks, native-input/directive rejection, never-run output | `TASK-260720-2g21eg` |
| csk-specific manager-home builds namespace, protected-state provenance, atomic immutable publication on POSIX and Windows | `TASK-260720-2jfnz6`, `TASK-260720-8nxlgx` |
| Existing runtime required-path validation plus self-contained Unix and Windows command activation | `TASK-260720-11yhth` |
| Audit-before-build ordering, provider/command ordering, generation checks, compiler-free persistent dry-run purity | `TASK-260720-2x6mjn` |
| Project, hybrid, global commit, consumer-last durability, rollback, recovery, cross-project isolation | `TASK-260720-z2z795`, `TASK-260720-3t8nr3`, `TASK-260720-g7kgox` |
| Status/currentness, reinstall repair, protected cache GC and journal roots | `TASK-260720-th0jdi` |
| Shared positive/minimum-negative vectors and legacy rc.3 coverage | `TASK-260720-12r55p` |
| Real Go build/launch, cache/rebuild, dry-run, rollback, concurrency and platform shims | `TASK-260720-3pemm6` |
| Author/operator/security/architecture docs and maintained Russian mirrors | `TASK-260720-akf5kh` |
| Full pytest, strict mypy, package build/twine, diff check, cross-platform CI and AC/vector evidence matrix | `TASK-260720-3s27te` |

No unresolved product or architecture question remains, so no clarification or research blocker was created.

## Diagram and board verification

- `TASK-260720-z9j4c9_csk-build-components.puml` and its SVG remain attached to `TASK-260720-z9j4c9`; source SHA-256 `a2d3bc2dc9fc1044f7cc1b85349a7e92040ac737a536f6a96c65657a971465f1`.
- `TASK-260720-z2z795_csk-build-lifecycle.puml` and its SVG remain attached to `TASK-260720-z2z795`; source SHA-256 `893a3b1d6db2c5cac1ddc9414fea0df8504d1d088cd794efcaba45cb67653aa9`.
- PlantUML 1.2026.6 `-checkonly` passes for both sources. The attached resource bytes match the repository artifacts, and both PNG renders were visually inspected for readable ownership and lifecycle flow.
- `task-board q 'plan(STORY-260720-1uv5gi, mode=children)'` reports 17 tasks, 12 phases, and no cycle.
- All 17 tasks have non-placeholder description, scope, AC, and at least three task-specific checklist gates. `TASK-260720-3pemm6` has a fourth explicit gate that preserves the released default pin and records candidate-suite evidence separately.
- `task-board validate` still reports only the 12 unrelated legacy `EPIC-260712-*` prose dependency references and one unrelated orphan `TASK-260713-7a9c1e` resource.

## Logbook anomaly

The installed task-board 0.20.1 persisted both `set_details` changes even though the command was invoked with `task-board m --dry-run`. A follow-up structured query showed the new scope and AC on disk. The intended corrections are valid, so no rollback was needed, but this violates the documented all-mutations dry-run contract and is recorded for tool maintenance.

No csk product, test, documentation, configuration, branch, or worktree state was changed by this solution-architecture audit.
