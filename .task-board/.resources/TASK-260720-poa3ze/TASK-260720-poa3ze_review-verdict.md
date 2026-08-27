# Review verdict: changes requested

**Task:** `TASK-260720-poa3ze` — Research compile-only build drivers  
**Route:** `analysis` (research/design rework)  
**Reviewed outcome:** `TASK-260720-poa3ze_compile-only-build-drivers.md`

## Review result

The revised outcome materially closes the previous review findings. It now defines a closed internal-link Go command, dependency-graph rejection rules for native inputs, effective private telemetry state, and byte-exact toolchain/cache/receipt algorithms. It also covers the requested schema examples, lifecycle, dry-run, rollback, language classification, affected artifacts, and a clear v1 recommendation.

It is not yet acceptable because the proposed cache input can alias snapshots that compile to different artifacts.

## Verification performed

- The reviewed `origin/main` revisions match the current remote heads:
  - curator-spec `57c1f56846d221ecc55786bd3c2467ec32f11730`
  - curator `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`
  - cocoaskills `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`
- All 14 fenced JSON examples parse.
- Independent recomputation matched the displayed cache key, receipt hash, and toolchain fingerprint vectors.
- Exact-ref gates are green:
  - curator-spec: 30 schemas and 93 vectors validated; 8 Python tests and Go tool tests passed.
  - Curator: `go test ./...` passed.
  - cocoaskills: 488 passed, 18 skipped.
- All three exact-ref worktrees remained clean.
- Primary Go documentation/source checks support the revised Go, telemetry, vendoring, cgo, linker, PGO, and language-toolchain claims.

## Finding requiring rework

### High — the cache key omits a compiler-visible file

Section 5.2 says that `snapshot_sha256` is “the existing protocol content hash over the complete raw snapshot” and makes that field the sole snapshot-content component of the canonical build input. The first half is factually incorrect:

- The current protocol explicitly hashes regular files **excluding the marker itself**: [curator-spec protocol/core.md, lines 282–295](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/protocol/core.md#L282-L295).
- The Go manager defaults to excluding root `.csk-install.json`: [Curator hashing.go, lines 21–29](https://github.com/relux-works/curator/blob/17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8/internal/hashing/hashing.go#L21-L29).
- The Python manager has the same default exclusion: [cocoaskills hashing.py, lines 16–26](https://github.com/ivanopcode/cocoaskills/blob/6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12/src/csk/hashing.py#L16-L26).

That exclusion is correct for installed-tree currentness because the manager-generated marker carries the hash. It is not a complete build-source identity.

Go permits an explicitly named dotfile in an embed directive. A root main package can therefore compile the package-provided marker bytes:

```go
import _ "embed"

//go:embed .csk-install.json
var payload []byte
```

The `embed` package documentation confirms that matching files are read and embedded at compile time and that explicit file patterns are supported: [Go embed directives](https://pkg.go.dev/embed#hdr-Directives). The protocol and proposed semantic rules do not forbid a package snapshot from containing that root file.

Consequently, snapshots A and B may differ only in root `.csk-install.json`. The current protocol content hashes are equal, so all otherwise equal build inputs produce the same cache key and a receipt whose input also appears valid. Yet a fresh compile produces different artifacts. Installing B after A can therefore reuse A’s artifact and receipt. This violates deterministic cache identity and the receipt’s claimed binding to the build inputs.

## Required revision

Choose and specify one closed rule:

1. **Recommended:** define a separate, byte-exact build-source digest over every regular file in the raw snapshot, including root `.csk-install.json`, while retaining the existing marker-excluding content hash for installed-tree currentness; or
2. reject every source file omitted by the build identity before cache lookup or any Go command. For the current protocol, that includes rejecting a package-provided root `.csk-install.json`.

Then:

- bind the new build-source identity/rule into canonical cache input, cache key, receipt validation, marker/currentness references, lifecycle ordering, dry-run output, and the affected-artifact list;
- ensure source validation/digesting occurs before cache reuse;
- add conformance vectors with `source_dir="."` and `//go:embed .csk-install.json`: two snapshots differing only in that file must produce different cache keys and artifacts, or both must be rejected before cache lookup/build;
- recompute the JSON/cache/receipt fixtures and rerun the exact-ref gates.

## Re-review gate

Update the task outcome and byte-identical `.research/` mirror, attach the revised task-scoped outcome, and return the task to `to-review`. Acceptance requires cache and receipt identity to cover every compiler-visible package byte (or a pre-cache semantic rejection that makes omitted bytes unreachable).
