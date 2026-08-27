# Review cycle 3 verdict: changes requested

**Task:** `TASK-260720-poa3ze` — Research compile-only build drivers  
**Route:** `analysis` — research and contract rework  
**Reviewed outcome:** `TASK-260720-poa3ze_compile-only-build-drivers.md` at SHA-256 `fc4da2bcefa341d4d911f337011c488f41ee69495d2b45c7cde731aa1a8acbe1`

## Summary

The latest revision closes the prior marker-embed cache alias. Its separate length-framed `curator-build-source-v1` identity covers root `.csk-install.json`, is used before cache lookup, and is consistently bound into keys, receipts, markers, lifecycle, dry-run, and vectors. The report also satisfies the task-level matrix and fixed-command coverage at a broad level.

The outcome is not yet acceptable because two downstream contract requirements remain undefined, one Go linker input surface is unclassified, and the parent-story comparison inventory is incomplete.

## Independent verification

- Current remote `main` still equals every inspected immutable ref: curator-spec `57c1f56846d221ecc55786bd3c2467ec32f11730`, Curator `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`, and cocoaskills `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`.
- The board outcome and `.research/260720_compile-only-build-drivers.md` are byte-identical at the reviewed SHA-256.
- All 17 fenced JSON examples parse. Independent recomputation matches both revised build-source vectors, cache key `sha256:53baf0cc...`, and receipt hash `sha256:0d7fa98e...`.
- Exact-ref gates pass: curator-spec validates 30 schemas and 93 vectors plus 8 Python tests and Go tool tests; Curator passes `go test ./...`; cocoaskills reports 488 passed and 18 skipped.
- All three exact-ref worktrees remain clean.
- `task-board validate` still reports only the pre-existing 12 legacy dependency references and one unrelated orphan resource; none belongs to this task.

## Findings requiring rework

### High — build-source exclusion from agent context is undefined

The downstream `STORY-260720-35dck7` — Protocol schema v6 acceptance criteria require build sources to be excluded from agent context. Current protocol context selection admits entries under `agents/`, `references/`, `assets/`, `templates/`, `examples/`, and `data/`, excluding a subtree only when another rule such as `runtime_roots` applies.

The proposed semantic rule 11 says only that generated output never enters prompt context and that a build source *may* be under a runtime root. It neither requires this nor rejects a declaration such as `source_dir: "assets/tool"`. It also does not cover transitive local Go packages or embedded inputs under another context-eligible root. Those compiler inputs can therefore become prompt-visible even though the consuming schema-v6 story explicitly forbids that result.

Define one closed, implementable rule for source/context separation. It must cover the selected package, transitive non-standard package inputs, and embed inputs; explain behavior on cache hits and dry-runs where `go list` is intentionally skipped; and add positive/negative context-selection vectors. A conservative location restriction is acceptable if dynamic graph-based exclusion cannot be applied consistently.

### High — whole-install rollback has no cross-project isolation rule

Lifecycle step 10 correctly places project/global contexts, adapter ledgers, hybrid targets, and the machine-wide consumer ledger in one transaction. Rollback also discusses a concurrent transaction. However, the only serialization rule is the same-project lock used for recovery. The manager profile permits independent projects to install concurrently, while both implementations share machine-home hybrid/cache/consumer state.

Two project installs can therefore read the same consumer ledger, stage different updates, and lose one update at commit. More seriously, one transaction can restore a stale backup over another transaction's successful shared-target commit. Per-cache-key locks do not protect hybrid targets, adapter ledgers, consumer state, transaction journals, or GC.

Specify transaction isolation and deterministic lock ordering. The simplest v1 rule is a manager-home mutation/commit-and-GC lock, while source checks and cache builds may occur outside it followed by revalidation. A finer-grained design must enumerate every shared lock, acquisition order, journal ownership, rollback conflict behavior, and GC coordination. Add concurrent two-project success and success-versus-rollback vectors.

### Medium — ordinary Go source can still select a dynamic library

The dependency preflight rejects cgo/native file fields, `.syso`, and non-standard assembly, but it does not inspect Go compiler directives. The Go compiler deliberately permits `//go:cgo_import_dynamic` in ordinary non-cgo source, even with `CGO_ENABLED=0`; the linker can use `_ _ "library"` to force that package-selected dynamic library into the executable. See the primary Go compiler source at <https://go.dev/src/cmd/compile/internal/noder/noder.go> and linker source at <https://go.dev/src/cmd/link/internal/ld/go.go>.

This surface is not exposed by the listed `go list` native-file fields and is inconsistent with the report's closed vendor/native dependency claim unless it is explicitly admitted and modeled. Prefer rejecting active non-standard `go:cgo_import_dynamic` directives before `go build`, with cache-hit validation/receipt semantics and a negative vector. If it is intentionally allowed, document why it cannot cause build-time external reads/execution on every supported native target and represent its runtime dependency/security implications explicitly.

### Medium — parent-story ecosystem and artifact coverage is incomplete

`STORY-260720-x8a1p7` — Build-driver security model explicitly asks for Meson and Node/TypeScript bundlers. The matrix covers Make/CMake and npm/`tsc`, but not Meson or bundler plugin/config execution. Add concise primary-source classifications for both. Also add `cli/curator.md` to the affected-artifact inventory, or explain why its documented install/dry-run/status behavior needs no change for build plans and diagnostics.

## Re-review gate

Update the main task outcome and its byte-identical `.research/` mirror; attach a new task-scoped rework outcome; add the required context, concurrency, directive, and ecosystem evidence/vectors; recompute all displayed hashes; rerun the same exact-ref gates; and return the task to `to-review` for another reviewer cycle.
