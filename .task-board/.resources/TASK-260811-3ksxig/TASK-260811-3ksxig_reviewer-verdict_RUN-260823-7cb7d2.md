# Reviewer verdict for TASK-260811-3ksxig

Verdict: **changes requested -> to-dev**

## Review authority

- Reviewer run: `RUN-260823-7cb7d2`
- `task-board spawn goal RUN-260823-7cb7d2`: no active goal; the run is not goal-bound.
- Reviewed producer outcomes: `TASK-260811-3ksxig_rework-report.md` and `TASK-260811-3ksxig_pnpm-patch-research.md`.
- Reviewed implementation scope: `internal/pnpmsource` and the pnpm profile documentation.
- No product or test code was modified by this reviewer.

## Rework confirmed

The prior review's YAML, patch, virtual-store-entry, and exact-peer-link changes are present. Focused tests now exercise trailing YAML documents, manager patch hashes and patched expected inventories, unclaimed `.pnpm` entries, and two peer contexts with missing/swapped links. Those tests pass.

## Required changes

### 1. Workspace dependencies disappear from the common Node capture graph

`parseImporters` now correctly emits non-root importer edges from `local:<path>` (`internal/pnpmsource/lock.go`, lines 442-458), and selection traversal follows those edges. But `buildNodeCapture` still searches for `importer:<path>` for every local root (`internal/pnpmsource/capture.go`, lines 468-483). Consequently a workspace such as `packages/cli` can have a selected `local:packages/cli -> b@1.0.0` edge in the pnpm graph while the corresponding common `nodesource.PackageInstance` has no dependency on `b`.

This breaks the canonical closure/checkpoint authority after the producer's own workspace-selection fix: C0-C5 can omit a dependency that materialization still installs and validates against the adapter-local graph. The current S01/N01 test asserts only the intermediate pnpm edge and never checks the emitted `NodeCapture` dependency.

Required rework: use one canonical importer/local identity throughout parsing, selection, and `buildNodeCapture`; assert that workspace-only runtime/dev/optional dependencies survive into the common capture and active graph; add a negative/closure assertion proving no selected workspace dependency can be omitted from checkpoint authority.

### 2. The complete installed `node_modules` layout is still not owned

`validateMaterializedTree` enumerates only `node_modules/.pnpm` (`internal/pnpmsource/materialize.go`, lines 437-490). `validateLocalRoots` deliberately skips every `node_modules` subtree (`materialize.go`, lines 779-810), and final admission inventories whatever remains without comparing it to an expected owned layout (`materialize.go`, lines 846-868). Therefore an undeclared top-level `node_modules/rogue` directory, file, or symlink can remain outside `.pnpm`, be copied into the retained materialization, and become receipted source despite having no lock/snapshot identity.

The new unclaimed-content tests cover only extra entries below `node_modules/.pnpm` (`internal/pnpmsource/conformance_test.go`, lines 637-680). They do not cover root/workspace `node_modules` entries, direct dependency links, `.modules.yaml`, or other manager-created top-level members.

Required rework: define and reconcile the entire allowed root and workspace `node_modules` layout, including exact direct links to selected snapshot/local identities and the closed set of manager metadata; reject every other member before copying/admission. Add metadata-less, valid-package, regular-file, and wrong/missing direct-link negatives outside `.pnpm`.

### 3. The implementation relies on pinned pnpm behavior without enforcing or exercising a pinned pnpm version

The patch hash and virtual-store filename algorithms are sourced from pnpm `v10.33.0`, but the adapter has no supported pnpm version/identity constraint. `newRunnerSession` proves only that the caller-provided manager binding matches its own C0 checkpoint; any version output/fingerprint can be supplied. The suite uses `fakePNPMRunner`, and both producer reports acknowledge that no real pnpm executable smoke ran because pnpm is absent.

Required rework: declare the exact supported pnpm release/profile (or a closed tested compatibility set), reject a manager identity outside it before store derivation, and run at least one real pinned-pnpm end-to-end fixture covering private-store derivation plus frozen/offline/scripts-disabled materialization. The deterministic runner remains useful for fault injection but cannot by itself prove compatibility with the external manager semantics this adapter wraps.

## Validation evidence

| Command | Result |
| --- | --- |
| `go test -count=1 -cover ./internal/pnpmsource` | pass; 80.3% statements |
| `go test -count=1 ./internal/artifactpolicy ./internal/nodesource ./internal/pnpmsource` | pass |
| `go vet ./internal/pnpmsource ./internal/artifactpolicy ./internal/nodesource` | pass |
| `golangci-lint run ./internal/pnpmsource ./internal/artifactpolicy ./internal/nodesource` | pass; 0 issues |
| `git diff --check` | pass |
| `command -v pnpm` | no pnpm executable found |

Green current tests do not satisfy the acceptance criteria because they do not observe the workspace edge loss or any unclaimed content outside `.pnpm`, and they simulate rather than execute the external pnpm behavior being claimed.

## Routing

This is ordinary implementation and test rework. There is no external or human-only blocker. Route to `to-dev`; another independent reviewer cycle is required after the fixes.
