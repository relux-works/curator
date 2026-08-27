# Review verdict: changes requested

**Task:** `TASK-260720-poa3ze` — Research compile-only build drivers  
**Route:** `analysis` (research/design rework)  
**Reviewed outcome:** `TASK-260720-poa3ze_compile-only-build-drivers.md`

## What passed review

- The outcome is English, task-scoped, attached to the board, and mirrored under `.research/`.
- The inspected `origin/main` objects are current and match the report: curator-spec `57c1f56846d221ecc55786bd3c2467ec32f11730`, curator `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`, and cocoaskills `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`.
- Direct inspection confirms the report's manager findings: both managers record the consumer before materialization, both initially reuse an existing runtime-root directory without verifying required paths, their replacement scope is not a whole-install transaction, and the Python registry path mutates persistent state before its dry-run return.
- The report covers the requested manifest shape, cache/receipt/marker model, lifecycle and rollback, language matrix, affected protocol artifacts, and clear v1 recommendation. All 12 fenced JSON examples parse with `jq`.
- Exact-ref validation is green: curator-spec validated 30 schemas and 93 vector files, its 8 Python tests and Go tool tests passed; curator `go test ./...` passed after initializing its pinned `skill-go-testing-tools` submodule; cocoaskills reported 488 passed and 18 skipped.

## Findings requiring rework

### 1. High — `go-v1` does not yet have a closed executable/input graph

The report promises no external linker (section 4.1), says merely not to inherit `GO_EXTLINK_ENABLED` (section 4.3), and constructs `go build` without an internal-link constraint (section 4.4). That is insufficient:

- In link mode `auto`, Go uses `GO_EXTLINK_ENABLED` when set and otherwise may use the value compiled into the toolchain; a value of `1` selects external linking. This matters because the report explicitly permits patched/custom trusted toolchains. See the [Go linker mode source](https://go.dev/src/cmd/link/internal/ld/config.go) and [Go build bootstrap documentation](https://go.dev/src/make.bash).
- Even in internal mode, the Go linker documents that an unset `-libgcc` default is obtained by running the compiler. The implementation invokes the external compiler to locate `libgcc` when host-object inputs leave unresolved symbols. See [Go linker flags](https://go.dev/cmd/link/) and [linker implementation](https://go.dev/src/cmd/link/internal/ld/lib.go?m=text).
- Package-controlled `.syso` host objects are not rejected by the proposed semantics. Package-controlled Go assembly is also an undeclared external input surface: the assembler supports `#include`, and its implementation first opens the supplied include name directly, so an absolute or escaping include can read outside the snapshot. See [Go assembler documentation](https://go.dev/cmd/asm/) and [assembler include implementation](https://github.com/golang/go/blob/master/src/cmd/asm/internal/lex/input.go#L401-L424). Without mandatory read confinement or a source restriction, the artifact can depend on bytes absent from `snapshot_sha256`.

This breaks both the report's no-external-linker claim and its deterministic cache-input claim.

Required revision:

1. Close the linker path explicitly, for example with manager-owned `GO_EXTLINK_ENABLED=0` plus fixed linker arguments equivalent to `-linkmode=internal -libgcc=none`, and reject targets/toolchains for which that mode cannot build. The final exact argv must be stated and tested.
2. Inspect the full active dependency graph (for example, fixed `go list -deps -json -buildvcs=false .` semantics), not only one root package. Reject package-controlled host objects (`SysoFiles`). Either reject non-standard assembly (`SFiles`) in v1 or make read confinement to the snapshot, trusted `GOROOT`, and operation-private directories mandatory on every supported host; a best-effort sandbox is not enough for cache identity.
3. Add negative conformance vectors that poison `cc`/`clang`, supply `.syso`, use an escaping/absolute assembler include, and exercise a toolchain whose external-link default is enabled. Assert that no external executable starts and no outside-snapshot input is accepted.

### 2. High — `GOTELEMETRY=off` is not a valid telemetry control

The proposed environment sets `GOTELEMETRY=off`, but Go exposes `GOTELEMETRY` as a read-only/non-settable Go environment value. On the reviewed host (`go1.25.5`), `GOTELEMETRY=off go env GOTELEMETRY` still reports `local`. Go's documentation says telemetry mode is changed with `go telemetry off` and stored below `os.UserConfigDir()`. See [Go telemetry configuration](https://go.dev/doc/telemetry) and [the `go` command documentation](https://go.dev/cmd/go/).

Required revision: define a real fixed policy—such as a minimum compatible Go version plus a package-independent `go telemetry off` step in an operation-private user-config root—or explicitly rely on a fresh private root's default `local` mode and account for its temporary writes. Do not claim that the ignored environment entry disables collection. Include this step/state in dry-run and process-graph semantics.

### 3. Medium — toolchain and receipt digests are not reproducibly specified

Section 4.2 calls for a "canonical path-and-content SHA-256" over `GOROOT`, link targets, and `go version` output, but does not define traversal order, entry/type framing, path encoding, symlink-target encoding, mode treatment, stdout newline handling, or whether a selected `go` executable outside `GOROOT` is included. The marker also uses `receipt_sha256` without stating the exact bytes hashed. Two conforming managers can therefore produce different cache keys for the same toolchain and receipt.

Required revision: specify the exact byte-level record algorithm (or require a trusted supplied toolchain digest), either require the resolved executable to be the fingerprinted `GOROOT/bin/go` or hash it separately, define `receipt_sha256` over exact canonical bytes, and add cross-language conformance vectors for files, links, ordering, and newline cases.

## Re-review gate

Revise the outcome resource and `.research/` mirror, rerun JSON/source-link checks, and return the task to `to-review`. Acceptance requires an actually closed `go-v1` process/input graph, effective telemetry isolation, and byte-exact interoperable digest rules.
