# TASK-260728-2jaw7h reviewer verdict — cycle 1

## CHANGES REQUESTED

Route: to-dev. This is ordinary implementation rework, not an external blocker.

### 1. Release-blocking: the candidate rewrites frozen rc.5 conformance identity

The accepted predecessor and the candidate are not byte-identical at `release/1.0.0-rc.5.json`. Predecessor SHA-256: `b32ee9d35fc2fce0539d0b1c4b15f9c5239115c22fe54a72401f5da6d6646441`; candidate SHA-256: `32114a3225dc8ee0b2b0e5981c998897dc67211d41d8ae64cd6263d470b14835`. The candidate changes both `candidate_protocol_pin.manifest_sha256` and `downstream_consumption.required_manifest_sha256` from `sha256:9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf` to `sha256:3675a28c17b8fa5116ff7afcd888e62258b40e49a9510a4aafbddb3f058dafd7`. `conformance/v1/manifest.json` and `conformance/v1/schema-cases/index.json` likewise change from predecessor hashes `9ba9b8ec...` and `2faa2baa...` to `3675a28c...` and `a14068e1...`.

This contradicts the accepted version boundary and the cycle directive: rc.5 conformance bytes and pins remain frozen, while rc.6 minting belongs to `TASK-260728-251p01`. The clean scratch `release_gate.py --version 1.0.0-rc.5 --commit HEAD` passing at `eb3bb375...` establishes only that the rewritten rc.5 set is self-consistent; it does not establish preservation of the accepted release.

Required rework: restore every frozen rc.5 conformance byte and pin to the accepted predecessor. Keep the new schemas, vectors, and generators on a candidate surface that does not silently regenerate rc.5 release identity, or defer their release-manifest inclusion to the rc.6 integration task. Add a compatibility guard that compares the frozen rc.5 release/manifest identity to its accepted predecessor or canonical pin, so a self-consistent rewrite cannot pass.

### 2. Contract reference is internally stale

`docs/compiled-build-toolchain-requirements.md` says in its header that landing added the `registry` `source_ref.surface` token and distinguished classifier `matches` as `absence` versus `value`. The body still states the old contract: around line 1233 it lists only `manifest`, `descriptor`, and `source_metadata`; the value-classifier rules around lines 349-365 do not encode the new absence/value distinction. The file itself says disagreement with `protocol/core.md` is a defect.

Required rework: update those body sections to match `protocol/core.md`, `common.schema.json`, registry vectors, and diagnostic payloads; add a guard/test that prevents the reference token/classifier rules from drifting again.

### Independent evidence

Green: `tools/validate.py` (48 schemas, 593 vector files); 141 Python unit tests; `go test ./tools/...`; `go vet ./tools/...`; `gofmt -l tools`; Python compileall with an external pycache; `git diff --check`; two deterministic regeneration passes; clean scratch release gate. Boundary probes passed against Go 1.25.1 and 1.25.5 with 16 go-directive cases, 13 toolchain-directive cases, 331 closure checks, and zero failures; all five expected-red controls failed as required on both versions. All predecessor `*.schema.json` files except the intentionally additive `common.schema.json` are byte-identical, every old `$defs` member is unchanged, 20 definitions are additive, and all pre-existing corpus files other than the rewritten aggregate manifest/index are byte-identical.

The executable contract is otherwise coherent, but acceptance is impossible until both defects are corrected and reviewed again.