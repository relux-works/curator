# Cache-identity rework evidence

**Task:** `TASK-260720-poa3ze` — Research compile-only build drivers  
**Reviewed finding:** A package-provided root `.csk-install.json` can be compiled through an explicit `//go:embed` directive even though the current installed-tree hash omits that file.  
**Revised outcome:** `.research/260720_compile-only-build-drivers.md`

## Key correction

The revised recommendation separates two identities:

- `content_sha256` remains the current marker-excluding installed-tree currentness hash.
- `curator-build-source-v1` is a new domain-separated, uint64be length-framed SHA-256 preimage over every regular file in the fully validated raw snapshot, including root `.csk-install.json`.

The build-source digest is computed before any cache lookup or Go command. It is bound into the canonical cache input, cache key, receipt validation, marker v2/currentness, lifecycle order, dry-run output, affected-artifact inventory, and conformance plan. The framing is deliberately not the legacy NUL-delimited record stream, which is not self-delimiting for binary file content.

## Reproduction and exact vectors

Two `source_dir: "."` snapshots shared `go.mod` and `main.go` and differed only in root `.csk-install.json`. Their source used:

```go
//go:embed .csk-install.json
var marker []byte
```

Observed identities:

| Check | Variant A | Variant B |
|---|---|---|
| Current marker-excluding `content_sha256` | `sha256:829a040a1455fdf96e2731aa5c089e7e42dbcec2a51b1db3222a610f0ffb5b35` | same |
| `curator-build-source-v1` | `sha256:0017492cfbcd822237a7e72239d45b59f0923b54f2ac2e0a59ecd9202cc48ad0` | `sha256:60fe9b764163b7d6bc38bc0cac63675398eb7167ad692fd189e221c3fc096266` |
| Fixture cache key | `sha256:55925b710d927c74ed073be34274f97ce6181353ec66f9e3474ce0a842b73ff4` | `sha256:fa27415b5ea6b5ffb90ce09638afec2141b001a04e2adcc24c501a4a7ad212a2` |
| Go 1.25.5 smoke artifact SHA-256 | `69f8a5d047b4345d8398750d0d59c0bbd6f60101d5c40689bb31625cc10583fd` | `7ddd293aa1aa441ef83e57a26ad7b886ad9156cac4cde4a317214b870ee5ff23` |

Go and Perl implementations independently produced the two build-source digests. The fixed driver compiled both artifacts but never executed them.

The revised illustrative fixtures recompute to:

- cache key: `sha256:53baf0cc688ba0d000c683b56de47ce782b04f7db7a1f95054f34b151d6c7858`
- exact stored receipt hash: `sha256:0d7fa98e678a3d9ef04e33f08f02fdda1d8ae13e3d536afe3a28f7516f4dcc69`
- fenced JSON: 17 blocks, all accepted by `jq`

## Source and regression verification

Remote `main` still equals the inspected immutable revision in all three repositories:

- curator-spec: `57c1f56846d221ecc55786bd3c2467ec32f11730`
- Curator: `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`
- cocoaskills: `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`

Exact-ref gates:

- `uv run --isolated --with-requirements requirements-dev.txt make validate` — 30 schemas and 93 vectors validated; 8 Python tests passed; Go tool tests passed.
- `CURATOR_CONFORMANCE_ROOT=<exact-spec-ref>/conformance/v1 go test ./...` — passed.
- `uv run --isolated --with 'pytest>=9' --with 'cryptography>=44' python -m pytest -q` — 488 passed, 18 skipped.

No product or specification source file was modified. Only the research mirror, task-managed board outcomes, and task-scoped `.temp/` evidence are in scope for this handoff.
