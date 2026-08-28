# Operator-owned Rust Git derivation authority

Task: `TASK-260819-2qytt1`  
Implementation consumer: `TASK-260811-2h4m0s`  
Decision status: proposed for independent review  
Date: 2026-08-19

## Decision

Adopt one production `rustsource.Manager` per Curator operation. The manager is an opaque, package-sealed authority created by the trusted CLI composition root. It owns the artifact-policy service, protected intake store, one causal `closureexec.Executor`, centrally selected Cargo 1.91.0/Cargo 0.92.0 toolchain, a fixed hidden Curator Git-oracle worker, private Cargo home, vendor/output roots, and all immediate rechecks. A request caller supplies only raw origins and selection facts. It cannot supply a toolchain, executor, provider, portable runner, vendor/metadata runner, callback, permit, receipt, selected path set, normalizer input set, normalized manifest, transform seal, Cargo home, config bytes, or destination.

The independent oracle is a Go implementation of the pinned Cargo 0.92 Git `PathSource` projection and manifest-normalization behavior. It runs as one hidden Curator worker under a committed `closureexec.DerivationPermit`; it does not consume `cargo vendor` output and does not ask Cargo to tell Curator which bytes should exist. Its canonical `rust-git-projection-v1` output is causally receipted, decoded, rebound to the admitted raw Git tree and exact package context, and converted to package-private transform state. Only then may the manager commit the Cargo vendor permit.

This is the artifactpolicy manager-owned authorization pattern applied to Rust: public constructors may create a genuine fixed manager, but no public parameter can replace its selector, seal, classifier, oracle, executor, runner, provider, or receipt verifier. The zero value and copied foreign values without the package-private production seal authorize nothing.

## Requirement trace

| Requirement | Design evidence |
| --- | --- |
| `.spec/skill-facing-cli-source-closure.md` — Current delivery scope: Rust is required this cycle | Git remains supported; no registry/path-only deferral is introduced. |
| Source closure invariant 1-4 and 6 | Raw lock/manifest/registry/Git/path inputs admit before interpretation; immutable identities, C0 tools, offline derivations, exact I/O, and conservative rejection are manager-owned. |
| Vendored compiled artifact prohibition | Complete Git trees admit before projection, including leaves that Cargo would omit; a denied leaf yields zero oracle/vendor/metadata starts. |
| Accepted Rust contract §4 “Vendoring is a versioned transform,” “Pre-vendor admission and controlled execution,” and “Git origin-to-vendor mapping” | Oracle ID, two projection branches, normalized manifest, per-leaf dispositions, Cargo toolchain, absent destination, source replacement, and post-vendor comparison remain exact. |
| Accepted Rust contract “Closure and build procedure” steps 1-9 | Manager lifecycle implements intake, admission, oracle derivation, vendor derivation, post-vendor admission, and later metadata on one causal chain. |
| Accepted graph/checkpoint contract “Evidence-derivation execution before C5” and “Protected execution boundary” | Every executable pre-C5 step has a committed permit, immediate tool/input rechecks, exact typed outputs, a receipt, and the next causal head. |
| `CGP11`, `CGN17`, `CGN18` | Tests cover ordered admitted-input derivation, toolchain drift before start, unauthorized/foreign authority, and observed output drift. |
| Task AC: production-reachable, non-substitutable, exact Go API, positive/forgery conformance | The API and test matrix below are implementation-ready and include an external-package positive Git case. |

## Justified gap and self-verification

**Missing piece.** The accepted Rust contract defines what Cargo 0.92 projection and normalization must mean, but the implementation had no production owner that could independently derive those bytes and prove that the receipt came from a non-substitutable authority. `closureexec` proves causal issuance only relative to an executor; it does not make a caller-created executor trustworthy.

**Requirement otherwise incomplete.** The accepted Rust §4 Git transform, closure procedure steps 3-6, `CGP11`, `CGN17`, `CGN18`, and `TASK-260811-2h4m0s` positive Git acceptance all require a production-reachable derivation that is independent from vendor output and committed after raw admission.

**Consequence if left open.** Keeping the current package-private test map leaves Git permanently fail-closed in production. Exporting the map, executor, provider, runner, or receipt would let a request caller issue a self-consistent forgery. Deferring Git violates current delivery scope.

**Closure.** The sealed manager plus fixed hidden oracle provides the missing owner, commits exact derivation permits, verifies receipts against its own executor, and passes only package-private sealed derivation state to the transform.

**Sections checked before accepting the gap-closing work.** Checked the full delivery spec: Current delivery scope, Source closure invariant, Vendored compiled artifact prohibition, Required research, Discovery deliverable, and Delivery completion. Checked accepted Rust: supported profile, full closure procedure, Git mapping, diagnostics, conformance, backlog, and acceptance mapping. Checked accepted graph/checkpoints: pre-C5 derivations, protected execution, conformance, unsupported cases, and decision log. Result: Git is required, manager authority is not answered, and verified-provider implementation remains excluded. The design therefore adds no verified provider, network service, GUI, new language, binary admission, or expanded Rust profile.

No new board element is justified. The authority rework and its conformance are one coherent deliverable already owned by `TASK-260811-2h4m0s`; splitting API, oracle, and tests would create artificial partial states. No research task is justified because Cargo behavior and assurance semantics are already decided.

## Exact Go API boundary

The implementation should replace free functions that accept execution seams with the following public surface. Names may move mechanically within `internal/rustsource`, but parameter direction and authority boundaries are normative.

```go
package rustsource

type ManagerConfig struct {
    // WorkRoot is an operator-configured parent. NewManager creates one new,
    // absent, mode-0700 session below it and never executes preexisting bytes.
    WorkRoot string

    // Assurance contains policy and provider identity only. It contains no
    // provider object or resolver. Empty means portable.
    Assurance closureexec.AssuranceConfig
}

// Manager is a concrete opaque authority, not an interface.
type Manager struct {
    state *managerState
}

func NewManager(ctx context.Context, config ManagerConfig) (*Manager, error)
func (m *Manager) Capture(ctx context.Context, request RawCaptureRequest) (*Capture, error)
func (m *Manager) DeriveMetadata(ctx context.Context, capture *Capture, selection SelectionContext) (MetadataResult, error)
func (m *Manager) Close() error

type RawFile struct {
    Path string
}

type RawTree struct {
    Root string
}

type RawManifest struct {
    Path string
    File RawFile
}

type RawRegistryOrigin struct {
    SourceLocator string
    IndexRecord   RawFile
    CrateArchive  RawFile
}

// One repository origin can satisfy multiple lock packages at the same exact
// source+commit. Package keys, package paths, leaves, include rules, index
// state, submodules, and tree digests are manager derivations, not fields.
type RawGitOrigin struct {
    DeclaredURL    string
    Selector       string
    LockedCommit  string
    Repository    RawTree
}

type RawPathOrigin struct {
    DeclaredPath string
    Tree         RawTree
}

type RawCaptureRequest struct {
    Workspace RawTree
    Lock      RawFile
    Manifests []RawManifest
    Registry  []RawRegistryOrigin
    Git       []RawGitOrigin
    Paths     []RawPathOrigin
}

// Capture carries detached evidence plus an opaque private handle. The handle
// is bound to the manager instance and cannot be replayed through another one.
type Capture struct {
    Evidence CaptureEvidence
    state    *captureState
}
```

### Public fields removed

Remove from `CaptureRequest` and `MetadataRequest`: parsed `LockFile`; caller-populated `GitOrigin.Leaves`, `Tree`, `PackagePath`, `Include`, `ManifestTracked`, `Submodules`, `Dirty`, `UsesFilter`, and `IndexConflict`; `CargoToolchain`; `RecheckToolchain`; `StageCargoHome`; `VendorRunner`; `MetadataRunner`; `CargoHome`; `ConfigPath`; `ConfigBytes`; `VendorDestination`; and every package-private derivation map. Parsed lock/manifests, immutable leaf manifests, tree/index/ignore state, private paths, configs, and destinations become manager state.

Do not replace these fields with `Option` functions, interfaces, callbacks, global mutable test hooks, context values, environment lookups, or serialized “authority” blobs. Tests that need faults use same-package `_test.go` constructors over unexported state; production files expose no injection seam.

### Constructor and assurance behavior

`NewManager` validates `WorkRoot`, creates an absent private session, and selects both fixed tools through closed manager selectors. It accepts no executable path. The selectors are:

- `curator-rust-git-oracle-v1`: the running signed Curator executable containing worker implementation `cargo-git-oracle-v1:cargo-0.92.0@ea2d97820c16195b0ca3fadb4319fe512c199a43`;
- `cargo-vendor-transform-v1`: Cargo 1.91.0 / crate 0.92.0 at implementation commit `ea2d97820c16195b0ca3fadb4319fe512c199a43`.

Selection uses the same closed-selector/private-seal pattern as `artifactpolicy.SelectExternalToolchain`: no public root, path, version, fingerprint, or authorization assertion. Both complete tool roots and selected executable bytes are admitted and immediately rechecked.

Portable is the default. The manager privately constructs `closureexec.ManagerProcessRunner` and `closureexec.NewAssuredExecutor`; portable receipts retain `network=not-observed` and must never claim verified host observation. Verified mode has no provider parameter. It resolves only an operator-installed provider through the trusted execution preflight. This release ships no provider, so verified selection fails with the existing verified-provider diagnostic before session staging, intake, cache lookup, or process start. Adding a provider remains separate excluded scope.

`managerState` contains the package-global production seal, executor, selected/admitted tool handles, protected intake store, session paths, current phase, and closed flag. Every method checks the seal, pointer identity, phase, and manager ownership of `Capture.state`. `Close` is idempotent, invalidates all handles, and removes only its exact session directory after all consumers release it.

## Lifecycle and ordering

1. CLI assurance preflight selects portable by default or fails closed for unavailable/drifted verified mode.
2. `NewManager` creates a private session; centrally selects, admits, and fingerprints the oracle and Cargo toolchains; constructs one causal executor with a manager-derived initial head.
3. `Capture` copies raw lock, manifests, registry records/archives, Git repositories, workspace, and path trees into protected intake. It parses nothing executable before complete relevant-tree admission.
4. Shared artifact policy admits every complete raw origin. Any deny/unknown/incomplete result returns with `oracle_starts=0`, `cargo_vendor_starts=0`, no private Cargo home, and no vendor destination.
5. Manager derives lock/package/repository mapping and a canonical package-context input from admitted bytes. For every Git package in canonical package-key order it commits and executes exactly one oracle permit.
6. Manager verifies the receipt was issued by its executor, reads the one output, canonical-decodes it, checks every binding against the raw origin/context, computes the package-private seal, and advances the causal head. No output from a failed invocation enters C1-C4.
7. After all registry and Git expected transforms exist, the manager stages a new private Cargo home solely from admitted inputs, derives exact source replacement, commits the Cargo vendor permit, immediately rechecks Cargo/tool/input/config/destination state, and runs the confined vendor operation into an absent destination.
8. Manager admits and reconciles every vendor output against the independently derived transform. Only an exact match creates `Capture.state` and C3b evidence.
9. `DeriveMetadata` accepts only a live capture from the same manager and commits the existing unfiltered then active metadata permits on the same causal chain. It accepts selection facts but no runner/tool/config/receipt.
10. Later build consumers retain the manager/capture until C4/C6 evidence is complete, then call `Close`.

## Pinned Cargo 0.92 oracle

The hidden worker is dispatched before normal CLI parsing only when the process receives the single exact internal mode argument `__curator_rust_git_oracle_v1`. It reads only fixed logical mounts and writes one fixed output. Manual invocation is harmless: bytes have no authority without a permit and a receipt issued by the manager’s private executor.

The oracle ports, with pinned fixture coverage:

- `git_index_no_include`: controlled Git dirwalk, package-root `target/` exclusion, unconflicted tracked-index re-addition, exclude/ignore semantics, nested-package removal, mandatory manifest/eligible lock rules, and generated `.cargo-ok` staging metadata;
- `filesystem_include`: include-filtered filesystem walk, hard exclusion of depth-one package-root `target/`, no index re-addition, nested-package removal, and mandatory manifest rules;
- recursive admitted submodule handling and empty manager-controlled system/global Git configuration;
- workspace inheritance, included build-script field, explicit/inferred lib/bin/example/test/bench targets, Cargo normalized-manifest construction, and exact serialization;
- exact normalizer identity `cargo-git-manifest-normalizer-v1:cargo-0.92.0@ea2d97820c16195b0ca3fadb4319fe512c199a43`.

The worker must not import vendor output, an ambient Cargo home, host Git configuration, a network client, or a request-controlled plugin. The implementation is versioned as a descriptor; behavior change requires a new descriptor and fixtures.

### Oracle permit

| `closureexec.DerivationPermit` field | Exact binding |
| --- | --- |
| `PreviousCausalHead` | Manager’s current head after all required intake admissions and previous canonical Git-package derivations. |
| `InvocationKey` | `rust-git-projection-v1:` plus the canonical package-instance ID. |
| `InvocationSubtype` | `closureexec.DerivationManifest`. |
| `AdmittedInputReceiptIDs` | Sorted unique receipts for lock bytes, applicable root/workspace manifests, complete clean Git commit/submodule tree, controlled Git index/ignore provenance, and manager-derived package-context record. No vendor bytes. |
| `InputMounts` / `ReadRoots` | One read-only logical mount per receipt; exactly equal, sorted, and non-overlapping with output. |
| `C0CheckpointID` | C0 checkpoint containing the admitted Curator oracle toolchain and the pinned implementation descriptor. |
| `ToolchainNodeID` | Centrally selected Curator oracle executable node. |
| `ToolchainFingerprint` / `ExecutableSHA256` | Complete selected tool-root fingerprint and exact staged Curator executable digest, rechecked immediately before and after. |
| `Executable` | Fixed manager-staged relative path to Curator; never a request path or PATH lookup. |
| `CWD` | Fixed empty manager session directory. |
| `Argv` | Exactly `[]string{"__curator_rust_git_oracle_v1"}`. |
| `Environment` | Closed map containing only fixed locale/timezone, fixed logical input/output paths, and `CURATOR_OUTPUT_ROOT`; no inherited environment. |
| `HostID` / `TargetID` | Exact native C0 host/target binding; cross-target oracle execution is unsupported. |
| `AllowedProcesses` | Portable receipt makes no observed-process claim; verified permit allows only the exact oracle executable and no child process. |
| `WriteRoots` | Exactly `[]string{"rust-git-projection-v1.json"}`. |
| `ExpectedEvidence` | One declaration: path `rust-git-projection-v1.json`, schema `rust-git-projection-v1`, artifact-manifest ID derived from the exact input/context/descriptor declaration. |
| `Network` | `none`; portable receipt records `not-observed`, verified receipt must observe `none`. |
| `RecheckRule` | `immediate-exact-v1`. |
| `ResourceLimits` | Fixed positive oracle profile: one process, bounded admitted-read total, one bounded output, bounded write total, and fixed wall time; ID committed. |

### Canonical oracle output

`rust-git-projection-v1` contains only canonical fields:

- schema ID and oracle/normalizer descriptor IDs;
- package-instance ID; declared URL/selector; full commit and tree; package path;
- ordered recursive submodule identities;
- projection mode and the exact facts selecting that mode;
- sorted complete selected paths and sorted normalizer input paths;
- sorted include/exclude/ignore/nested-package decisions and a disposition/reason for every admitted origin leaf;
- manifest-tracked/index-stage/clean-tree facts;
- exact normalized `Cargo.toml` bytes as strict base64, size, and SHA-256;
- exact IDs of every admitted input and package-context record.

The manager rejects unknown fields, noncanonical order/encoding, wrong descriptor, missing/duplicate path, incomplete leaf coverage, a commit/tree/package/include mismatch, a mode that does not follow its predicate, a normalizer-input mismatch, or output digest/size mismatch. It then calls `Executor.VerifyIssuedDerivationReceipt`, binds the receipt ID into the package-private seal, and passes only the sealed state to `deriveGitTransform`.

## Non-substitutability proof

1. `RawCaptureRequest` contains data locations only; reflection over its reachable exported field graph finds no interface, function, channel, unsafe pointer, `closureexec` executor/provider/runner/permit/receipt, derived projection, or normalized bytes.
2. `Manager` is concrete and its only field is private. A request cannot provide a `Manager` implementation.
3. `NewManager` accepts no behavior object. Closed selectors and the production seal are package-owned. A caller can request a genuine manager but cannot make it use different code or trust a different receipt.
4. Verified provider resolution is outside the request and absent by default. A caller-created provider cannot be passed or registered.
5. Oracle output is produced only after admission by the fixed executable under a single-use permit. Manual/forged output has no executor-issued receipt and no private seal.
6. Vendor and metadata execution methods are private manager bridges over the same executor. Public runner interfaces disappear.
7. Exact origin/context rechecks after receipt prevent a valid old receipt from being rebound to new commit/tree/package/include/manifest bytes.
8. Complete origin admission precedes projection, so omitted/unselected compiled bytes cannot exploit the oracle.

## Implementation mapping for `TASK-260811-2h4m0s`

This is one atomic rework of the existing Rust capture deliverable; create no new task.

1. Add sealed manager lifecycle and closed operator tool selectors. Move session roots, Cargo home/config, destination, toolchain selection/rechecks, artifact policy, protected intake, executor, and causal head into `managerState`.
2. Replace `CaptureAndVendor(ctx, CaptureRequest)` and `RunPermittedMetadata(ctx, MetadataRequest)` public execution-seam APIs with manager methods and the raw-origin request types above. Delete public `VendorRunner`, `MetadataRunner`, `Permit`, callback, and toolchain/config authority fields.
3. Implement the hidden pinned oracle worker and Cargo 0.92 projection/normalizer port. Keep descriptor and schema constants exact; use retained GV01-GV03e bytes as the independent oracle corpus.
4. Build private closureexec bridges for oracle, vendor, and metadata permits/receipts. Verify exact issued receipts and causal-head ordering; do not serialize authority into public strings.
5. Convert oracle output to the existing package-private sealed `gitDerivation`; keep transform/post-vendor verification and diagnostic precedence.
6. Replace same-package-only positive transform proof with production API conformance and retain focused internal fault tests.

No new shared provider implementation is part of this rework. If a required closureexec primitive is missing, extend only the generic operator-owned constructor/output handling needed by these exact derivations inside the same task; do not add a request injection seam.

## Conformance matrix

| Test | Required assertion |
| --- | --- |
| External-package production positive: GV01 | `package rustsource_test` constructs `NewManager` with portable default and supplies only raw fixture paths. Git capture succeeds, oracle/vendor receipts exist, normalized hash is `e0cf597d…df3`, copied source and checksum match, remote/cache can be removed before later offline replay. |
| External-package production positive: GV02/GV03a/GV03c/GV03e | Exact mode, selected/normalizer paths, workspace inheritance, tracked root-target re-addition, explicit-include root-target omission, nested package/submodule behavior, and normalized/checksum hashes match retained corpus. No request field supplies an expected projection. |
| Exported API audit | Reflection/go-types test proves reachable request fields contain no function/interface and none of the forbidden authority/derived types or names. `Manager` is concrete; zero manager and foreign capture fail. |
| Executor/provider/runner forgery | External test constructs caller executors/providers/runners and canonical-looking receipts but cannot pass them to the API. A sidecar `rust-git-projection-v1.json` in any input/ambient directory is ignored or rejected as undeclared raw content; starts remain governed by manager state. |
| Receipt/output forgery | Same-package fault test substitutes a canonical-looking output without an issued receipt, a receipt from another manager/executor, wrong receipt subtype/schema/path/digest/size, or mutated bytes after receipt. Reject with `closure_derivation_unauthorized`, `closure_derivation_drift`, or `rust_git_identity_invalid`; `cargo_vendor_starts=0`. |
| Binding drift | Mutate commit, tree, submodule, package mapping/path, include declarations, manifest/index/ignore bytes, mode, selected paths, normalizer inputs/ID, or normalized bytes between intake/permit/use. Reject deterministically; no affected later start/checkpoint. |
| Toolchain drift: CGN17 | Change oracle or Cargo selected-root fingerprint/executable bytes after permit and before use. Reject before the affected process; affected start count zero. |
| Unauthorized derivation: CGN18 | Missing/stale/widened permit, extra argv/env/read/write/output/process, foreign capture, or wrong phase fails; no evidence enters C1-C4. |
| Pre-admission zero spawn: GV03b/GV03d/PV01/RF01-RF08 | Rejected dirty/compiled/missing/mismatched origin yields `oracle_starts=0`, `cargo_vendor_starts=0`, `metadata_starts=0`; Cargo home/vendor/target/results absent. |
| Oracle failure zero Cargo | Oracle may start once; any nonzero exit, resource failure, missing/extra output, noncanonical payload, or binding mismatch yields `cargo_vendor_starts=0`, `metadata_starts=0`, no vendor destination. |
| Post-vendor forgery: VF01-VF03 | Only confined vendor may have started. Independent origin/oracle expected bytes reject forged copied/normalized/checksum/config/directory output; metadata/build starts remain zero. |
| Assurance modes | Empty mode equals portable and receipts claim only portable capabilities plus `network=not-observed`. Explicit verified with no trusted provider fails before session/intake/start; provider/capability drift causes zero affected starts and no fallback. |
| Race/lifecycle | Concurrent calls serialize by causal head; permit is single-use; `Close` races are safe; capture from manager A is rejected by manager B; no state survives close. Run focused race tests. |

Retain and wrap accepted `CGP11`, `CGN17-CGN18`, `R01-R08`, `RV01-RV03`, `GV01-GV03e`, `VF01-VF03`, `PV01`, and `RF01-RF10`. Stable diagnostic precedence remains origin/integrity, artifact admission, toolchain identity, derivation authority/drift, then transform/graph failures.

## Board proportionality and dependency audit

- Existing leaf `TASK-260811-2h4m0s` owns all implementation and tests above and is already blocked by `TASK-260819-2qytt1`; this is the only required new dependency edge and it already exists.
- The current architecture task is one clear deliverable: select and specify the authority seam. It does not write implementation code.
- No additional story/task/research element maps to a distinct requirement without duplicating the existing Rust task.
- Verified provider implementation, Git deferral, registry/path-only narrowing, binary admission, extra Rust capabilities, GUI/network service, and other languages remain rejected/out of scope.

## Decision log and anomalies

1. A causal receipt is necessary but not sufficient when the caller owns the executor. Authority comes from the sealed fixed manager and closed selectors; the receipt proves use of that authority.
2. The current public `Runner`, toolchain, recheck, staging, and config fields are authority injection seams even though the package-private Git map blocks the most direct forgery. They must be removed together; sealing only the projection bytes is incomplete.
3. A hidden worker is intentionally callable but not authoritative. This keeps the algorithm testable and production-reachable without making output bytes a capability.
4. The oracle is independent from `cargo vendor` output, not independent from the accepted Cargo behavior. Its descriptor and golden corpus pin that behavior.
5. Portable mode cannot claim host-wide read/write/network observation. The pure fixed oracle and exact input/output verification are portable controls; verified-only observations stay absent.
6. No standalone diagram is needed to define types, but a single focused sequence artifact is attached because it makes the authority/causal ordering and zero-start branches materially easier to review.

## Review gates

Independent review should reject the design if any exported request/option can supply behavior, trust assertions, derived bytes, private workspace contents, or receipt authority; if the oracle can consume vendor output or ambient Git/Cargo state; if verified mode can silently fall back; if a denied raw leaf can start the oracle or Cargo; if a projection failure can start Cargo; or if any listed requirement lacks an implementation/test mapping.
