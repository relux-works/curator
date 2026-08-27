# Review cycle 4 verdict: changes requested

**Task:** `TASK-260720-poa3ze` — Research compile-only build drivers  
**Route:** `analysis` — research and contract rework  
**Reviewed outcome:** `TASK-260720-poa3ze_compile-only-build-drivers.md` at SHA-256 `6618fae3312e0b212e2e1e7a6daa3c2c8aa1d09abdb8e165564129373b86cb61`

## Summary

The cycle-3 rework closes the returned context-isolation, cross-project rollback, Go dynamic-import, Meson/bundler, and CLI-inventory findings. The pinned repositories and their existing tests are green, and the report's JSON/digest examples recompute.

The outcome is not yet acceptable. Its cache-hit proof contradicts its own threat model: an attacker-controlled receipt can self-consistently name any regular executable artifact because neither the cache key nor any trusted value commits to the expected output hash. The affected-artifact inventory also omits the protocol-version-pinned conformance-claim schema/cases, and the proposed physical cache name needs to be separated from the protocol's implementation-specific machine-state boundary.

## Independent verification

- Current remote `main` still equals every inspected detached ref: curator-spec `57c1f56846d221ecc55786bd3c2467ec32f11730`, Curator `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`, and cocoaskills `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`.
- The board outcome and `.research/260720_compile-only-build-drivers.md` are byte-identical at the reviewed SHA-256.
- All 21 fenced JSON examples parse. Independent recomputation matches cache key `sha256:3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48`, receipt hash `sha256:750f5f7557ffe1dac1ec06b7d47cb5976e390a7dc7d57ea616c9edda91013c11`, toolchain vector `sha256:baf7c5f3b9c3f1fae3da4c356381bf74442aa7f8f0b6fb2304c9c10833d6032e`, both marker-inclusive build-source digests, and their equal legacy marker-excluding content hash.
- Exact-ref gates pass: curator-spec validates 30 schemas and 93 vectors, passes 8 Python tests and Go tool tests; Curator passes `go test ./...`; cocoaskills reports 488 passed and 18 skipped.
- Official Go documentation confirms private `go telemetry off`, `-mod=vendor` network/module-cache behavior, and `GOTOOLCHAIN=local`; primary Meson, webpack, esbuild, pip, and Deno documentation supports the revised language classifications. The Go 1.25.5 compiler/linker sources support the `go:cgo_import_dynamic` rejection.
- All three detached worktrees remain clean. `task-board validate` reports only the pre-existing 12 legacy dependency references and one unrelated orphan resource; none belongs to this task.

## Findings requiring rework

### High — an untrusted, self-consistent receipt does not prove build provenance

Section 2.1 explicitly treats cached build entries and receipts as attacker-controlled until validation. Section 5.4 then calls exact input equality "the cache-hit proof that the artifact was created only after" the required checks. The listed validation proves only internal consistency:

1. `cache_key` commits to source, command, policy, target, and toolchain input, but not to an independently known artifact digest.
2. The receipt supplies its own artifact digest and size.
3. `receipt_sha256` is only a hash of those attacker-supplied receipt bytes; it is not a signature, MAC, or value anchored before cache lookup.
4. A new install has no trusted prior marker. The manager would generate marker v2 from the accepted cache hit, so the marker does not create provenance retroactively.

A concrete counterexample using the report's displayed input leaves cache key `sha256:3fcd714a...` unchanged, replaces the artifact with the 24 bytes `attacker-chosen-artifact` (`sha256:a4f06a1304c926ed7f2326c8fd90cabc5c5bd2981e690a4351c852d91c079d88`), and canonicalizes a structurally valid receipt at `sha256:9a23f5b77e6173b0f10e7ed43cd2b21aa3b99f3a34945ec432fbb31338a6186d`. With that regular file at the derived path, every listed receipt check succeeds even though no trusted `go-v1` build produced it. The payload could equally be a valid executable or script.

Choose and specify one coherent trust model:

- make manager-created, owner-protected cache provenance part of the TCB, remove attacker-controlled cache replacement from scope, define ownership/permission/link checks, and describe receipts as corruption/currentness metadata rather than cryptographic provenance; or
- keep cache contents attacker-controlled and authenticate the receipt/artifact binding with a trust anchor whose key lifecycle and cross-implementation behavior are defined; or
- rebuild and compare any cache entry lacking authenticated provenance.

Add a negative conformance vector containing an exact-input, exact-key, self-consistent forged receipt/artifact. It must be rejected or rebuilt under the chosen model. Update threat model, receipt schema/semantics, marker currentness, dry-run, repair, cache publication, and GC consistently.

### Medium — protocol-version conformance artifacts are missing from the inventory

At the inspected ref, `schemas/v1/conformance-claim-v1.schema.json` line 9 fixes `protocol_version` to `1.0.0-rc.3`; its valid/invalid schema cases and generated conformance manifest carry the same release identity. Schema 6 and marker v2 are release/version metadata changes, but section 8 names the generator, validator, manifest, and release docs without naming the conformance-claim schema or its cases.

Add the exact claim-schema migration to the affected-artifact list and recommendation. Decide whether compatibility requires `conformance-claim-v2` or another explicit reader/writer transition; update its schema cases and generated manifest inventory. Do not leave new schema-v6 conformance results representable only as rc.3 claims.

### Medium — the physical cache namespace conflicts with the current portability boundary unless labeled illustrative

Section 5.3 says to use `<manager-home>/build-cache/go-v1/<key>/...` as an exact layout. Protocol Core lines 31–34 state that machine-home directories and cache names are implementation-specific, and the manager profile says implementations do not share machine-local state by default. The downstream Python implementation story also explicitly retains csk-specific manager-home layout.

Keep the logical key, receipt bytes, artifact-relative path, and validation semantics portable, but mark the physical cache-root/name as non-normative and implementation-specific. If the intent is instead to standardize a shared physical cache namespace, record that as an explicit compatibility-boundary change and update the affected core/profile artifacts accordingly.

## Re-review gate

Update the primary task outcome and its byte-identical `.research/` mirror; attach a new task-scoped rework outcome; close the cache-provenance contradiction with the forged-hit vector; add the conformance-claim migration; clarify the cache-layout portability boundary; recompute all displayed hashes; rerun the same exact-ref gates; and return the task to `to-review` for another reviewer cycle.
