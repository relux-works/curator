# Conservative Node/TypeScript and Python source closure

Status: research decision ready for review under `TASK-260810-2n3sbi`.

## Context and decision scope

Curator needs a Node/TypeScript adapter that has the same security properties as
the Go baseline: every recursive input is known and immutable before execution,
the captured closure can be replayed without a network or ambient cache, and
only declared build steps may turn admitted source into protected outputs. The
Python protocol implementation is an independent ecosystem reference. It is
not assumed to live in this repository, share a package cache, expose reusable
resolver code, or be a prerequisite for the Node implementation.

This decision is subordinate to the accepted
[source-closure specification](../.spec/skill-facing-cli-source-closure.md) and
[compiled-artifact policy](260811_compiled-artifact-taxonomy-and-deny-policy.md).
It also uses the accepted
[language and reference-surface inventory](260811_inventory-language-and-reference-surfaces.md),
which establishes the repository boundary and records that the observed Python
script environments are not presently source-only closures.

This is a policy and conformance decision, not a claim that a Node adapter or a
new Python implementation already exists. It defines the evidence an
implementation must produce before it can claim support.

## Executive decision

Use a **shared security-policy and evidence layer with separate ecosystem
closure profiles and implementations**.

1. Share the artifact classifier, trust roles, canonical dependency/build graph,
   immutable-origin rules, checkpoint schemas, sandbox and protected-output
   contract, diagnostic namespace, and byte-level conformance fixtures.
2. Implement Node resolution and materialization independently for npm, pnpm,
   Yarn Classic, and modern Yarn. Their lock formats, cache layouts, lifecycle
   behavior, workspace semantics, and install modes are not interchangeable.
3. Keep Python resolution and PEP 517 behavior in the independent Python
   implementation. Mirror the shared protocol and fixtures; do not share code,
   repository state, virtual environments, package-manager caches, or mutable
   package indexes with the Node adapter.
4. Support a conservative pure-source profile first. Recursively inspect raw
   package containers before installation. Reject precompiled Node addons,
   Python extension modules, bytecode, WebAssembly, native objects/libraries,
   opaque payloads, and every other shared deny class.
5. Disable package-manager lifecycle/build execution during materialization.
   Permit only separately declared, policy-approved build or generator nodes in
   an offline protected build. Implicit or newly discovered hooks fail closed.
6. Treat package-manager caches as derived acceleration, never as closure
   authority. The authoritative bundle is immutable raw artifacts plus their
   origin, hashes, canonical inspection manifests, lock evidence, graph, target,
   manager/runtime identities, and declared build plan.
7. Prove offline behavior by replaying from that bundle in a fresh sandbox with
   empty ambient caches and kernel-enforced network denial. An ecosystem's
   `--offline` option is a useful secondary check, not the proof boundary.

The result deliberately shares policy outcomes without coupling delivery:

```text
                 shared, versioned protocol and policy
        +------------------------------------------------+
        | graph schema | artifact manifest | diagnostics |
        | checkpoints | sandbox contract   | fixtures    |
        +----------------------+-------------------------+
                               |
              +----------------+----------------+
              |                                 |
    Node ecosystem profile             Python ecosystem profile
    npm / pnpm / Yarn adapters          independent implementation
    Node + manager identity             interpreter + frontend/backend
    lifecycle / TS build plan           PEP 517 build plan
```

No edge in this design requires the implementations to be co-located. They
interoperate through canonical records and expected fixture results.

## Closure model and proof obligations

### Canonical graph

For target identity `T`, let `G_T` contain:

- one node for every root, workspace/local source root, resolved package
  instance, raw package artifact, declared build dependency, build backend or
  source build tool, trusted runtime/toolchain, generator step, and published
  output;
- typed edges for runtime, development/build, peer, optional/conditional,
  workspace/local, package artifact, hook/generator, output, and FFI use;
- the exact condition that selected or pruned each target-dependent edge; and
- immutable identities for all nodes that may contribute bytes or behavior.

A package name and version are not necessarily a unique graph node. Peer
contexts, target markers, registry/origin, workspace root, and artifact digest
can distinguish instances. The canonical graph is sorted and hashed; an
adapter-specific installed-tree layout is not the graph authority.

### Admission predicate

For each selected target, support is granted only when all of the following are
true:

```text
source_closure(T) = lock_authoritative(T)
                 AND graph_complete(T)
                 AND every artifact has immutable origin and verified digest
                 AND recursive artifact admission succeeds
                 AND build dependencies and hooks are declared and closed
                 AND runtime, manager, backend, and toolchain identities bind
                 AND offline replay from an empty ambient state succeeds
                 AND every published output has a protected causal receipt
```

`UNKNOWN`, missing metadata, a parser gap, mutable locator, unselected target,
unavailable artifact, integrity mismatch, dynamic undeclared input, or
incomplete inspection makes the predicate false. There is no best-effort
closure.

### Why the construction proves recursive immutability

The proof is compositional rather than manager-specific:

1. A version-pinned manager adapter parses the authoritative root lock and
   manifests into `G_T`; every selected edge must end at a graph node.
2. Every external package node is bound to a raw artifact by immutable origin,
   exact size, and cryptographic digest. A registry URL without a digest, a Git
   branch/tag, or an external live directory is insufficient.
3. The shared artifact policy recursively walks the raw container and every
   recognized nested container. One rejected, opaque, unsafe, or uninspected
   member rejects the package and all ancestors.
4. Parsed package metadata from the admitted artifact is reconciled with the
   lock graph. A missing transitive edge, unexpected package, changed peer or
   marker context, or name/version mismatch rejects the graph.
5. Build and generator edges are separately enumerated. The sandbox supplies
   only graph nodes and selected trusted toolchains, denies the network, starts
   with an empty output root, and verifies the observed write set.
6. Offline replay starts without registry state or user caches. Therefore a
   successful replay can only consume the captured nodes. Its receipt binds the
   closure, graph, target, execution policy, runtime/toolchains, and outputs.

The conclusion is limited and auditable: the adapter proves that every input
inside its closed grammars and execution boundary came from the captured
closure. It does not claim that admitted source is benign or that a later CLI
runtime will never intentionally use the network.

## Node and TypeScript closure profile

### Package-manager comparison

| Manager/profile | Lock and integrity authority | Captured package bytes | Frozen/offline materialization | Conservative caveats |
| --- | --- | --- | --- | --- |
| npm | Root `package-lock.json` or publishable `npm-shrinkwrap.json`; bind lockfile version, every selected `resolved` locator and `integrity`, root manifest, npm version, and tree-shaping config | Registry `.tgz` bytes keyed by the lock integrity plus Curator SHA-256; never the installed `node_modules` tree alone | `npm ci --ignore-scripts --offline` in a private cache derived from captured tarballs; manifest/lock disagreement must fail | `prefer-offline` may use the network and is not admissible; the npm cache is explicitly a cache, not durable closure storage; implicit `binding.gyp`/`node-gyp` must be detected |
| pnpm | Versioned `pnpm-lock.yaml` parser; bind package snapshot, resolution/integrity, peer context, importer/workspace data, overrides, patches, target settings, and exact pnpm config | Raw package tarballs in Curator CAS; a private pnpm content-addressed store is derived evidence | Populate only from the captured set, then run frozen, offline, ignore-scripts materialization with a read-only/frozen private store | `pnpm fetch` skips some local `file:` dependencies; `.pnpmfile.*` can mutate manifests/resolution and is not disabled by `ignoreScripts`; side-effects cache contains hook-produced outputs and must be off |
| Yarn Classic 1.x | Root `yarn.lock`, root/workspace manifests, resolved URL and integrity, exact Yarn version/config; dependency-subtree lockfiles are not authority | Source `.tgz` files in a task-private offline mirror, each independently hashed and recursively scanned | `yarn install --frozen-lockfile --offline --ignore-scripts` with an empty ordinary cache and only the captured mirror | Do not use checksum-update modes; workspace roots and hoisting/layout config are checkpoint inputs |
| Modern Yarn/Berry | Versioned `yarn.lock` plus manifests, `.yarnrc.yml`, plugin declarations, patch data, exact Yarn release, cache key/compression/linker settings, and checksums | Exact normalized ZIPs in a task-private `.yarn/cache` or manager-neutral raw artifacts from which that cache is deterministically derived | `enableNetwork: false`, immutable lock/cache, and `yarn install --mode=skip-build`; OS network denial remains mandatory | Yarn plugins can add resolvers, fetchers, linkers, commands, and hooks; Git sources can invoke pack/build behavior; PnP loaders and install state are generated outputs, not lock authority |

These profiles share the same output schema, not the same lock parser. A lock
entry is necessary but not sufficient: it must select immutable bytes, and the
bytes and their embedded metadata must independently agree with the graph.

`package-lock.json` records a complete install tree in current npm lockfile
formats, and `npm ci` rejects a manifest/lock mismatch without rewriting the
lock. npm also documents that its cache is strictly a cache and may refetch
missing or corrupted content. These behaviors support frozen validation but do
not turn a mutable cache into source evidence
([npm package-lock](https://docs.npmjs.com/cli/v11/configuring-npm/package-lock-json/),
[npm ci](https://docs.npmjs.com/cli/v11/commands/npm-ci/),
[npm cache](https://docs.npmjs.com/cli/v11/commands/npm-cache/)).

pnpm documents that `--offline` uses only packages already in its store and
fails when an artifact is absent, while `--frozen-lockfile` refuses a stale or
missing lock. Its `fetch` command prepares the virtual store from the lockfile,
but skips local `file:` dependencies; those roots therefore need a distinct
captured workspace/local-source node
([pnpm install](https://pnpm.io/cli/install),
[pnpm fetch](https://pnpm.io/cli/fetch)).

Yarn Classic distinguishes a source-tarball offline mirror from its
implementation-specific unpacked cache. Modern Yarn stores normalized package
archives and offers immutable lock/cache checks and network disablement, but
its linker and generated-state choices remain versioned inputs
([Yarn Classic offline mirror](https://classic.yarnpkg.com/blog/2016/11/24/offline-mirror/),
[Yarn Classic install](https://classic.yarnpkg.com/en/docs/cli/install),
[modern Yarn install](https://yarnpkg.com/cli/install),
[modern Yarn configuration](https://yarnpkg.com/configuration/yarnrc),
[modern Yarn caching](https://yarnpkg.com/features/caching)).

### Package tarballs, workspaces, and graph reconciliation

- Capture the raw npm-compatible tarball before extraction. Bind the manager's
  SRI/checksum record and Curator's SHA-256 over the same raw bytes.
- Recursively inspect the gzip/tar layers and every nested archive before any
  manager or package script sees the content. Validate paths, duplicates,
  links, limits, opaque data, and compiled formats under the shared artifact
  policy. npm's `pack` output is a `.tgz`, but that label is never an admission
  shortcut ([npm pack](https://docs.npmjs.com/cli/v11/commands/npm-pack/)).
- Parse the admitted package's `package.json` and reconcile name, version,
  dependency classes, conditions, `os`/`cpu`/`libc`, entry points, scripts, and
  package-manager-specific metadata with the lock graph.
- Capture every workspace as an immutable tree with a portable canonical digest
  and manifest. Reject a workspace/local dependency that resolves outside the
  declared capture root, changes after checkpoint, or is omitted from the
  graph.
- Record optional and platform-pruned packages and the exact condition rather
  than silently deleting them. A closure is target-specific; another OS,
  architecture, libc, Node version, feature, or peer context gets another graph
  and checkpoint.
- Reject `bundleDependencies`/`bundledDependencies` in profile v1. npm packages
  can embed dependency trees in their tarball; recursive scanning detects their
  bytes, but the adapter cannot infer an independently immutable package origin
  and graph identity merely from their installed layout
  ([npm package manifest](https://docs.npmjs.com/cli/v11/configuring-npm/package-json/)).
- Reject Git, hosted-Git, HTTP archive, mutable tag/range-only, and external
  path locators in profile v1 unless a later version defines an immutable fetch
  and packaging identity equivalent to registry tarballs. A commit hash alone
  does not define stable archive bytes or suppress repository build hooks.

### Lifecycle hooks and manager extensions

Materialization always disables lifecycle scripts. This includes dependency
`preinstall`, `install`, `postinstall`, and `prepare` behavior. The adapter also
detects manager-specific implicit or pre-resolution execution:

- npm can synthesize `node-gyp rebuild` for a package containing `binding.gyp`
  when the package does not define its own install/preinstall script and does
  not opt out with `gypfile: false`. Treat that as a hook edge, never as an
  innocuous absence of `scripts.install`
  ([npm rebuild](https://docs.npmjs.com/cli/v10/commands/npm-rebuild/)).
- pnpm's `.pnpmfile.cjs`/`.pnpmfile.mjs` hooks can alter manifests and add custom
  resolvers or fetchers; pnpm documents that `ignoreScripts` does not suppress
  this file. Profile v1 rejects it, custom fetchers/resolvers, undeclared
  patches, and a populated side-effects cache
  ([pnpm hook file](https://pnpm.io/pnpmfile),
  [pnpm build settings](https://pnpm.io/settings/build)).
- Modern Yarn plugins can add resolvers, fetchers, linkers, commands, and hooks.
  Profile v1 accepts only the manager-owned, fingerprinted built-in plugin set
  named by the profile; local or downloaded plugins reject. Yarn Git sources
  can automatically execute package build/pack lifecycle behavior, which is a
  second reason they are unsupported
  ([Yarn extensibility](https://yarnpkg.com/features/extensibility),
  [Yarn lifecycle scripts](https://yarnpkg.com/advanced/lifecycle-scripts)).

A manifest-declared root build is not executed merely because it is named
`build`, `prepare`, or `postinstall`. It may be promoted into a separate build
node only when policy records the exact command/arguments, working directory,
environment, runtime and executable resolution, read set, write set, order,
target, and output classes. Package-manager convenience invocation must not add
implicit pre/post scripts. Dependency lifecycle scripts remain disabled; a
dependency that cannot function without one is unsupported in the pure-source
profile.

Any attempted unplanned process, file read outside the closure/toolchain,
write outside the declared empty output root, graph mutation, or network access
fails before publication. This execution control, rather than the fact that a
script appeared in `package.json`, is what prevents a declared string from
silently adding undeclared inputs.

### Generated JavaScript and TypeScript

There are two intentionally different admission cases:

1. **Shipped generated JavaScript.** Transpiled, bundled, or minified JS that is
   already present in an immutable package tarball may be admitted as
   `source.generated_text` when the full bytes satisfy the Node source grammar,
   every accompanying source map/config file is classified, and that exact JS
   is the declared runtime input. Missing upstream TypeScript is an audit
   limitation, not a claim that Curator reproduced the upstream publication.
2. **Locally generated JavaScript.** A TypeScript compiler, bundler, code
   generator, and plugin must be an admitted source dependency or an explicitly
   selected external toolchain. Bind its package digest/version, Node runtime,
   `tsconfig` and all config, source set, command, environment, plugin set,
   target, outputs, and output hashes. The generated JS/source maps are
   `local_build_output` and publish only with a protected receipt.

A package-provided TypeScript compiler is executable build logic even if it is
pure JavaScript. It can run only as a declared build node in the networkless
sandbox. `eval`, dynamic imports, or source semantics are not statically proven
safe; containment is supplied by the declared read/write/process/network
policy.

### Native addons and other executable payloads

- Reject every dependency `.node` file by bytes and role, including renamed or
  nested prebuilds. Node documents native addons as dynamically linked shared
  objects ([Node C++ addons](https://nodejs.org/api/addons.html)).
- Reject native executables, objects, static/dynamic libraries, WebAssembly,
  V8 code caches/startup snapshots, and opaque generated blobs under the shared
  compiled-artifact policy. `os`, `cpu`, `libc`, N-API, or package metadata
  cannot change the dependency-input trust role.
- Profile v1 does not build native-addon source. A future mixed C/C++ addon
  profile may emit a `.node` as a protected local output only after it models
  the complete C-family closure, `node-gyp`/generator inputs, compiler/linker,
  Node headers, N-API or Node module ABI, target libc/OS/architecture, and FFI
  edge. Until then it returns `closure_native_build_unsupported` before build.

### Node cache and runtime identity

The portable closure contains raw package artifacts in a Curator-owned
content-addressed store. It does not import an ambient npm cache, global pnpm
store, Yarn cache, `node_modules`, PnP install state, or side-effects cache as
trusted source. A manager-specific cache/store may be reconstructed in a fresh
task namespace from already admitted artifacts, then frozen and hashed as a
derived materialization receipt. Cache poisoning therefore cannot change the
closure without changing a captured digest or causing reconciliation failure.

The Node execution checkpoint binds at least:

- full fingerprint and selected path of the external Node runtime and package
  manager, exact version output, adapter/profile and lock-format version;
- `process.version`, `process.platform`, `process.arch`, and the complete
  `process.versions` map, including `modules`, `napi`, `v8`, and linked runtime
  components when present;
- target OS, architecture, libc/ABI, conditions, module system, linker/layout,
  workspace set, case-sensitivity assumptions, and canonical relevant
  environment/config; and
- for TypeScript generation, compiler/plugin/config identities and target/module
  options.

Node exposes platform, architecture, and component/ABI versions through
`process`; those values are evidence, while the independently computed
toolchain fingerprint protects the executable bytes
([Node process API](https://nodejs.org/api/process.html)).

## Python closure profile

### Lock and requirements comparison

| Input format | What it can establish | Conservative admission rule | Limitation that must fail closed |
| --- | --- | --- | --- |
| `pylock.toml` | Standardized reproducible-install graph with environments, extras/groups, dependency edges, and candidate wheel/sdist/archive identities and hashes | Preferred interchange format when a versioned parser selects exactly one target result and every selected artifact is captured and scanned | Do not infer that PEP 517 static and dynamic build requirements are closed merely because install packages are locked; bind a separate build graph |
| Hash-complete pip requirements | Exact requirement lines, hashes, includes/constraints, and global pip options | Require every selected transitive dependency explicitly pinned or immutably referenced, every archive SHA-256-hashed, `--require-hashes`, and no resolver work during replay | Ordinary requirements, exact top-level pins, constraints alone, index selection, ranges, and unhashed transitives are not a lock |
| Tool-specific lock (`uv.lock`, PDM, Poetry, Pipenv, other) | Often exact versions, dependency edges, markers, and artifact hashes according to that tool | Accept only through a separately versioned adapter/exporter that binds the exact frontend version, lock schema, settings, target selection, and resulting canonical graph | Never parse all formats as one generic lock; never assume pip will honor another tool's lock or that a tool-specific format is stable |

The standardized `pylock.toml` specification records package dependencies and
artifact locations/hashes for reproducible installation
([`pylock.toml` specification](https://packaging.python.org/en/latest/specifications/pylock-toml/)).
Pip requirements files are pip input files that can contain index, link,
constraint, editable, binary/source, and nested-requirement options; their
syntax alone does not promise a solved graph
([pip requirements format](https://pip.pypa.io/en/stable/reference/requirements-file-format/)).

When pip requirements are used as the closure authority, Curator applies the
same all-or-nothing discipline as pip's hash-checking mode: every dependency
must be present, pinned, and hashed. Curator additionally requires SHA-256,
captures raw bytes before replay, and rejects unsafe source classes
([pip secure installs](https://pip.pypa.io/en/stable/topics/secure-installs/)).
Tool-specific locks remain valid candidates, not portable assumptions. For
example, uv calls `uv.lock` a universal cross-platform lock but documents its
format as tool-internal rather than a stable public interface; a pinned adapter
or standardized export is required
([uv project layout](https://docs.astral.sh/uv/concepts/projects/layout/),
[uv lock internals](https://docs.astral.sh/uv/reference/internals/metadata/)).

### Target selection and dependency graph

Select extras, dependency groups, Python requirement, implementation, platform,
architecture, ABI, and environment-marker values before claiming closure. Bind
both selected and pruned edges with their marker expressions. Wheel candidate
selection additionally binds the ordered compatible-tag calculation and the
frontend/tag-library identity. One universal lock may yield multiple distinct
target closures; each target has a separate canonical graph digest and offline
receipt.

Reconcile each admitted package's core metadata with the lock. Name/version,
`Requires-Python`, dependency/extra markers, entry points, and build metadata
must agree with the selected node. A lock-selected artifact whose embedded
metadata changes the graph returns `closure_metadata_mismatch` rather than
quietly invoking a resolver.

### Source distributions and build backends

A Python sdist is a gzip-compressed tar archive, not evidence that every member
is source. Capture its immutable origin, size, and SHA-256; enforce the modern
sdist structural requirements; recursively inspect all members; and reconcile
the root name/version and `PKG-INFO`/`pyproject.toml`
([source distribution specification](https://packaging.python.org/en/latest/specifications/source-distribution-format/)).

Profile v1 supports an sdist build only when all of these conditions hold:

1. It uses PEP 517 through a present, admitted `pyproject.toml`; legacy direct
   `setup.py` invocation and editable installation are unsupported.
2. Every static `[build-system].requires` package, build frontend, external
   backend package, and backend dependency is separately locked, captured,
   scanned, and represented in the build graph. An in-tree `backend-path` must
   resolve inside the immutable sdist root and its source must be classified.
3. The expected results of `get_requires_for_build_wheel` and any corresponding
   sdist hook are recorded in the checkpoint before execution. The admitted
   backend may run the discovery hook in a read-only, networkless sandbox only
   after its own closure exists; returned requirements must exactly match the
   predeclared locked set. A new requirement returns
   `closure_build_dependency_unlocked` and cannot trigger a fetch.
4. Config settings, backend hook name/order, frontend/backend versions, Python
   runtime, environment, target, source epoch/normalization policy, read set,
   write set, and expected output class are bound.
5. The locally built wheel is recursively inspected before publication.
   `METADATA` is reconciled with the graph and every file is checked against
   `RECORD`. A native/opaque output, undeclared write/process, network attempt,
   or metadata drift rejects publication.

PEP 517 allows backends to return extra build requirements from
`get_requires_for_build_wheel`, and allows an in-tree backend through
`backend-path`. Both are code/input edges, not harmless implementation details
([PEP 517](https://peps.python.org/pep-0517/)). The project metadata
specification likewise distinguishes static build requirements and dynamic
project metadata
([`pyproject.toml` specification](https://packaging.python.org/en/latest/specifications/pyproject-toml/)).

For the pure-Python profile, the build sandbox does not expose a C/C++/Rust
compiler or arbitrary host headers/libraries. A package that invokes them is an
unsupported native build, even if its sdist itself contains only source text.

### Pure and native wheels

A wheel is a ZIP container with distribution metadata and a `RECORD` table. Its
filename tags and `Root-Is-Purelib` field guide compatibility/layout; they do not
override the byte classifier. The adapter must:

- verify the locked outer digest and size, ZIP structure and safe paths;
- recursively classify every member, including nested archives;
- parse `WHEEL`, `METADATA`, and entry-point metadata under a closed grammar;
- verify every required `RECORD` digest and size and reject unrecorded members;
  and
- reconcile name, version, requirements, Python requirement, tags, and selected
  target with the graph.

A wheel admits as a source container only when every member is allowed
source-like text/metadata under the shared policy. The current inert-binary data
allowlist is empty, so a nominally pure wheel containing an unrecognized image,
font, database, or other opaque binary still rejects. The wheel format requires
per-file hashes in `RECORD`, but a valid hash proves identity, not source class
([wheel format](https://packaging.python.org/en/latest/specifications/binary-distribution-format/)).

Reject any wheel containing `.so`, `.pyd`, `.dll`, `.dylib`, object/library,
WebAssembly, bytecode, or another compiled deny class, including a file hidden
under a text suffix. A `py3-none-any` name is not sufficient evidence of purity.
Python documents extension modules as shared libraries such as `.so` or `.pyd`
([Python extension modules](https://docs.python.org/3/extending/building.html)).

Installed virtual environments and wheelhouses are not portable closure
authority. A wheelhouse/source bundle may carry the exact admitted raw artifacts
for offline replay, but the environment is rebuilt with `--no-index`, without
dependency resolution, and with bytecode generation disabled. Shipped `.pyc`
files are rejected as dependency bytecode; replay-created caches are disposable
local outputs and are not published. Pip documents the `--no-index --no-deps`
wheelhouse pattern for repeatable installs, while Curator applies its stricter
source-only admission rule to every wheel
([pip repeatable installs](https://pip.pypa.io/en/latest/topics/repeatable-installs/),
[Python bytecode compilation](https://docs.python.org/3/library/py_compile.html)).

### Python runtime and backend identity

The Python execution checkpoint binds at least:

- the selected interpreter and standard-library/toolchain root fingerprint,
  executable relative path, full version/build output, implementation name and
  version, `sys.implementation.cache_tag`, platform, architecture, pointer
  width, ABI/extension suffixes, and relevant `sysconfig` values;
- the ordered compatible-tag set used for candidate selection, plus the exact
  packaging frontend and tag-library versions;
- lock adapter/parser, installer/frontend, PEP 517 backend and all build-tool
  package identities; extras, groups, markers, config settings, target, and
  canonical relevant environment; and
- source-closure, artifact-manifest, graph, hook/build-plan, sandbox-policy, and
  output-receipt digests.

`sys.implementation`, `cache_tag`, executable and version data are defined by
Python's runtime API; compatibility tags encode interpreter, ABI, and platform
and are selected in preference order
([Python `sys`](https://docs.python.org/3/library/sys.html),
[platform compatibility tags](https://packaging.python.org/en/latest/specifications/platform-compatibility-tags/),
[`packaging.tags`](https://packaging.pypa.io/en/latest/tags.html)). When a
Python installation supplies standardized build-details metadata, include it as
additional interpreter-build evidence, not as a substitute for fingerprinting
the selected bytes
([Python build details](https://packaging.python.org/en/latest/specifications/build-details/)).

## Shared checkpoints and evidence

Every adapter emits the same checkpoint sequence. Ecosystem-specific evidence
lives in typed fields; it does not change the security meaning.

| Checkpoint | Required evidence and decision |
| --- | --- |
| `C0.profile` | Shared policy/artifact versions, adapter/profile/manager identity, supported lock schema, source grammars, target identity, config/environment allowlist. Reject unknown formats or manager extensions. |
| `C1.resolve` | Root/workspace manifests and hashes, authoritative lock hash, selected/pruned graph with conditions, exact package instances, peer/build/runtime edges, parser output digest. Reject stale lock, missing edge, mutable locator, external path, or metadata conflict. |
| `C2.capture` | For every external node: immutable origin, expected and observed size/digests, raw bytes in protected CAS, retrieval record. Network is permitted only in this capture phase; no package code or hook runs. |
| `C3.admit` | Shared recursive `artifact-manifest-v1` for every raw artifact and local root, detector/limit/profile identity, canonical member hashes and verdict. Any compiled, opaque, unsafe, unsupported, encrypted, or incomplete member rejects. |
| `C4.close` | Reconciled canonical dependency/build graph and digest; every selected edge resolves to one admitted node or approved toolchain. Bind target and manager-specific layout semantics without treating layout as authority. |
| `C5.plan` | Explicit hook/generator/build DAG, command/arguments, executable resolution, read/write/process/network policy, environment, toolchains/runtimes, build order, expected dynamic-hook results and outputs. Undeclared or implicit execution rejects. |
| `C6.offline` | Fresh sandbox identity, empty ambient-cache/home proof, private derived cache/store receipt, network=`none`, frozen/install flags, materialized-tree digest, attempted network/process/I/O violations, and exact input set. Success is required. |
| `C7.publish` | Locally built output manifest, graph/build/toolchain/runtime/checkpoint digests, observed write set, file classes/sizes/hashes, canonical receipt and protected-cache publication. Runtime entry points must resolve only to admitted source or receipted output. |

Checkpoint `C(n)` consumes and binds the digest of every earlier checkpoint.
Changing a lock, target, manager option, parser, detector, runtime, build backend,
toolchain, hook plan, or package byte invalidates all downstream receipts.

### Offline replay protocol

Offline replay is the executable proof of closure:

1. Complete `C0` through `C5` during an authorized capture session. Retrieval
   writes only raw immutable artifacts into the protected CAS; package code is
   never executed.
2. Create a new private sandbox with a read-only closure mount, empty writable
   output root, empty task-specific home, no global/user manager configuration,
   no ambient npm/pnpm/Yarn/pip caches, and no inherited credential agents.
3. Enforce network denial outside the package manager. Configure the manager's
   own offline/network-disabled and frozen/immutable modes as defense in depth.
4. Derive a task-private manager cache/store or wheelhouse only from captured
   admitted artifacts. Record its construction and digest; make it read-only
   before installation when the manager permits.
5. Materialize with all lifecycle/build scripts disabled and dependency
   resolution forbidden. Reconcile the resulting source tree/package map with
   `G_T` and fail on a missing or extra instance.
6. Run only `C5` build nodes, in order, under the declared process/read/write
   policy. Inspect and receipt outputs before publication.
7. Repeat with poisoned ambient caches present but inaccessible. The closure
   digest, materialized graph, diagnostics, and output hashes must remain the
   same. Remove one required raw artifact and verify a deterministic
   `closure_offline_input_missing` before any build node starts.

The replay is successful only with an actual network-denial audit and empty
ambient state. `npm --offline`, `pnpm --offline`, Yarn `enableNetwork: false`,
or pip `--no-index` alone cannot prove that subprocesses, hooks, plugins, or
other tools stayed offline.

## Stable closure diagnostics

Artifact-content failures retain the already accepted shared codes such as
`artifact_compiled_dependency_forbidden`,
`artifact_opaque_dependency_forbidden`,
`artifact_generated_input_undeclared`, and
`artifact_inspection_unavailable`. Closure/build failures add the following
global codes. Manager-specific causes are structured `ecosystem`, `manager`,
`phase`, and `reason` fields, not one-off code names.

| Code | Required meaning and key fields |
| --- | --- |
| `closure_lock_missing` | No authoritative root lock for the selected profile; include root and expected formats. |
| `closure_lock_format_unsupported` | Lock/schema/manager/profile version has no closed parser; include format and exact manager identity. |
| `closure_lock_stale` | Manifest, workspace, configuration, or embedded package metadata disagrees with the frozen lock; include both digests and field/edge. |
| `closure_integrity_missing` | Selected external package lacks an admissible cryptographic size/digest binding. |
| `closure_integrity_mismatch` | Observed raw artifact size/digest differs from lock/capture evidence; include expected/observed values and origin. |
| `closure_origin_unpinned` | Locator is mutable or does not define stable retrievable bytes; include scheme and sanitized locator. |
| `closure_graph_incomplete` | A selected dependency, peer, build, workspace, optional, generator, FFI, or artifact edge has no uniquely admitted node. |
| `closure_local_path_escape` | Workspace, path dependency, backend path, patch, or config resolves outside its declared immutable capture root. |
| `closure_bundled_dependency_unsupported` | Node dependency payload embeds a package tree whose independent origin/graph identity cannot be established under profile v1. |
| `closure_manager_plugin_undeclared` | Resolver/fetcher/linker/plugin/hook/patch/config extension is absent from the approved profile or checkpoint. |
| `closure_hook_undeclared` | Lifecycle, implicit install/build, backend, generator, or convenience pre/post hook is absent from the exact build plan. |
| `closure_build_dependency_unlocked` | A PEP 517 hook, Node build tool, or other build step requests an input not already locked, captured, admitted, and expected. |
| `closure_native_build_unsupported` | Pure-source profile encounters a native-addon/extension/FFI build edge before execution; include source package, target, and required toolchain class. |
| `closure_offline_input_missing` | Replay cannot find an exact captured artifact or derived cache member; include package node and digest. |
| `closure_network_attempted` | Any install/build process attempts network access after capture; include process/build-node and destination class when safely available. |
| `closure_generated_output_drift` | Generated path, class, size, digest, metadata, or declared write set differs from the build plan/receipt. |
| `closure_runtime_identity_changed` | Manager, Node/Python runtime, backend, compiler, linker, or selected toolchain differs between checkpoint and use. |
| `closure_metadata_mismatch` | Artifact/package/wheel metadata would alter identity, dependencies, markers, tags, entry points, or target selection relative to the canonical graph. |

Diagnostic precedence is deterministic. A recognized compiled member uses
`artifact_compiled_dependency_forbidden` even if it would also imply an
unsupported native build. Raw integrity/origin failure precedes parsing;
archive safety/content failure precedes graph/build planning; graph failure
precedes offline materialization; execution violations precede output drift.
Every failure records the last successful checkpoint, package node, canonical
virtual path/container chain when applicable, and confirms whether any hook or
build process started.

## Explicit unsupported cases

Profile v1 fails closed for these cases rather than emulating a partially
closed install:

### Node/TypeScript

- a missing/stale root lock, unsupported lock schema, missing integrity, mutable
  locator, registry response not bound to exact raw bytes, or unresolved peer;
- Git/hosted-Git dependencies, arbitrary HTTP archives, external path roots,
  and embedded bundled dependencies without a future immutable-origin profile;
- npm implicit `binding.gyp`/`node-gyp`, dependency lifecycle scripts required
  for correctness, pnpm `.pnpmfile.*` or side-effects cache, non-approved Yarn
  plugins, Git pack hooks, and any other undeclared manager extension;
- native addon source builds, prebuilt `.node` files, native libraries/objects,
  Wasm, V8 cached code, package-provided compiled tools, and opaque blobs;
- generated JS whose shipped bytes are not the declared immutable runtime
  input, or local generation without complete compiler/plugin/config lineage;
- an ambient/global cache, shared pnpm store, mutable `node_modules`, or Yarn
  install-state/PnP file presented as closure authority; and
- target/platform branches that cannot be resolved and recorded before capture.

### Python

- ordinary unpinned/unhashed requirements, ranges, index-only resolution,
  constraints without a full locked graph, editable installs, VCS branches, and
  external mutable paths;
- a tool-specific lock without a pinned decoder/frontend and canonical export;
- legacy direct `setup.py` builds, an escaping `backend-path`, unlocked static
  build requirements, or PEP 517 dynamic requirements that differ from the
  predeclared locked result;
- native wheels, extension modules, compiled libraries/objects, `.pyc`, Wasm,
  opaque wheel members, and native extension builds from sdists;
- wheel `RECORD` gaps/mismatches, metadata/lock drift, or purity inferred only
  from a wheel filename/tag or `Root-Is-Purelib`;
- a venv, pip cache, manager cache, or platform-specific wheelhouse presented as
  proof without the independently captured raw-artifact and graph evidence; and
- runtime marker/tag selection that cannot be reproduced from a fingerprinted
  interpreter and versioned frontend.

A future policy may add bounded immutable Git/archive origins, centrally parsed
inert data, approved manager extensions, or source-built native/FFI profiles.
Those are policy-version changes with new fixtures, not adapter flags.

## Conformance fixtures

Each fixture asserts the canonical graph and artifact-manifest digests, stable
decision/code, failing node/path, checkpoint boundary, hook/build-started flag,
network audit, materialized tree, and protected-output receipt where relevant.
The same semantic vector is wrapped in each supported manager; an ecosystem
wrapper may add evidence but may not weaken the expected result.

### Shared closure and offline vectors

| ID | Fixture | Expected result |
| --- | --- | --- |
| `S01` | Root -> source package A -> B -> C, with exact immutable artifacts and a nested source archive | All nodes/members enumerate; clean offline replay succeeds from an empty ambient state |
| `S02` | Remove or tamper with C after checkpoint | `closure_offline_input_missing` or `closure_integrity_mismatch`; no hook/build starts |
| `S03` | Seed ambient cache with same package identity but different bytes | Ambient entry is inaccessible/ignored; captured digest and output remain unchanged |
| `S04` | Declared source build tries DNS/TCP/Unix proxy or invokes a downloader | `closure_network_attempted`; no output/cache publication |
| `S05` | Identical ELF/PE/Mach-O/Wasm leaf nested in npm tarball and Python wheel/sdist | Same shared artifact class, primary diagnostic, and leaf manifest digest |
| `S06` | Same normalized members in permuted archive/lock order | Same graph, artifact manifest, decision, and checkpoint digests |
| `S07` | A selected transitive edge is absent, while an extra installed package is present | `closure_graph_incomplete`; layout cannot repair or widen the lock graph |
| `S08` | Replay once with empty caches and once with poisoned inaccessible caches | Identical materialized graph/output; audited network remains `none` |

### Node/npm/pnpm/Yarn vectors

| ID | Fixture | Expected result |
| --- | --- | --- |
| `N01` | Three-level pure JS/TS registry graph, independently encoded for npm, pnpm, Yarn 1, and modern Yarn | Each adapter emits the same canonical package graph and succeeds offline from captured tarballs/archives |
| `N02` | Root/workspace manifest changed without matching lock update | `closure_lock_stale`; manager must not rewrite the lock |
| `N03` | Tarball content disagrees with SRI/checksum or Curator SHA-256 | `closure_integrity_mismatch` before extraction/install |
| `N04` | Dependency has install/postinstall script that writes or downloads | Materialization skips it; explicit attempt is `closure_hook_undeclared`; marker proves no hook ran |
| `N05` | npm package has `binding.gyp` and no explicit install script | Detect implicit `node-gyp rebuild`; `closure_native_build_unsupported` before install |
| `N06` | `.node`, native library, Wasm, or V8 cache direct, renamed, and nested at multiple archive depths | Shared compiled diagnostic; no manager hook/build starts |
| `N07` | Package ships minified JS/source map; root separately compiles TS with a fully declared compiler/config | Shipped JS admits as exact generated text; local output publishes only with lineage receipt |
| `N08` | TS generator adds undeclared input/plugin or produces an extra file | `closure_build_dependency_unlocked`, `closure_hook_undeclared`, or `closure_generated_output_drift` as appropriate |
| `N09` | Workspace path escapes root or changes after checkpoint | `closure_local_path_escape` or integrity/runtime drift before build |
| `N10` | Optional/peer package selects differently for OS/arch/libc/Node target | Separate target graph; exact selected/pruned reasons audited; unresolved context rejects |
| `N11` | npm bundled dependency, pnpm `.pnpmfile`, side-effects cache, Yarn local plugin, or Git dependency | Stable profile-specific unsupported/undeclared diagnostic; no extension code runs |
| `N12` | Required archive absent from private cache/store while ambient manager cache contains it | `closure_offline_input_missing`; ambient cache cannot satisfy replay |
| `N13` | Modern Yarn PnP loader/install state or `node_modules` is pre-seeded | Ignore and regenerate as receipted output; pre-seeded bytes cannot become source authority |

### Python vectors

| ID | Fixture | Expected result |
| --- | --- | --- |
| `P01` | `pylock.toml` and equivalent hash-complete requirements for A -> B -> C using source-only wheels | Same canonical graph; valid `RECORD`; offline replay with no resolver/network succeeds |
| `P02` | Exact top-level pin but missing/unhashed transitive B or C | `closure_integrity_missing` or `closure_graph_incomplete`; no install |
| `P03` | Source-only wheel with valid `RECORD`, then member/hash/size/unrecorded-file variants | Positive admits; variants return archive/integrity/metadata failure deterministically |
| `P04` | Native `.so`/`.pyd` wheel with misleading `py3-none-any` name, renamed member, and nested compiled member | `artifact_compiled_dependency_forbidden`; tag does not override bytes |
| `P05` | Modern sdist whose locked backend/static requirements build a pure source-only wheel | Capture and admission precede declared networkless PEP 517 build; output/metadata/RECORD receipt passes |
| `P06` | `get_requires_for_build_wheel` returns a new or changed dependency | `closure_build_dependency_unlocked`; no fetch and no wheel build |
| `P07` | Legacy setup-only project, editable/VCS dependency, or escaping `backend-path` | Explicit unsupported/path diagnostic before backend execution |
| `P08` | Backend tries network, reads host package state, invokes undeclared compiler, or writes outside output root | Network/hook/native-build/write-set failure; no publication |
| `P09` | Built wheel `METADATA` changes dependencies/name/version or output bytes drift | `closure_metadata_mismatch` or `closure_generated_output_drift` |
| `P10` | Markers and wheel tags select different artifacts for two interpreter/platform/ABI identities | Two separately bound target graphs; reuse across identity is rejected |
| `P11` | Source-only sdist attempts to emit a native extension | `closure_native_build_unsupported` before tool invocation or compiled-output rejection; no publication |
| `P12` | Direct/nested `.pyc`, opaque member, or unrecognized binary data inside a nominally pure wheel/sdist | Shared bytecode/opaque diagnostic; purity label is irrelevant |
| `P13` | Tool-specific lock parsed with another frontend/version or changed lock schema | `closure_lock_format_unsupported` or `closure_runtime_identity_changed` |

Every negative fixture also asserts that no later checkpoint was issued and no
package-manager cache, installed tree, wheel, addon, or executable was
published. Positive offline fixtures must be run with an OS-level network deny,
not simulated by an empty test server.

## Implementation-ready policy split

| Layer | Ownership recommendation | Node-specific work | Python relationship |
| --- | --- | --- | --- |
| Artifact admission | Shared Curator policy/service | Call it on npm tarballs, Yarn ZIPs, workspace trees, generated outputs | Independent implementation must produce the same manifest/verdict or consume the same service protocol |
| Canonical graph/evidence | Shared versioned schema and digest rules | Encode npm/pnpm/Yarn instance, peer, optional, workspace, hook, generator, and runtime evidence | Encode Python install/build/backend/marker/tag edges without sharing repository state |
| Resolution/lock parser | Separate ecosystem adapters | One pinned parser/materializer per manager/lock generation | Keep Python frontend/lock adapters in the independent implementation |
| Build policy | Shared execution contract; separate profiles | Node lifecycle suppression and explicit TS/JS generators | PEP 517 frontend/backend hooks and expected dynamic build requirements |
| Offline materialization | Shared sandbox and proof semantics; manager-specific cache builders | Private npm cache, pnpm store, Yarn mirror/cache | Private raw-artifact bundle/wheelhouse and derived venv |
| Runtime/toolchain identity | Shared checkpoint fields plus typed ecosystem payloads | Node/package-manager/process versions and TS tool lineage | Python/interpreter/frontend/backend/tag identity |
| Diagnostics and fixtures | Shared codes, ordering, fixture meanings | Manager wrappers N01-N13 | Python implementation runs P01-P13 independently and compares canonical outcomes |

Recommended implementation order:

1. Land the shared closure graph/checkpoint schemas and closure diagnostic enums
   on top of the accepted artifact policy.
2. Build one manager-neutral raw-artifact CAS importer and offline sandbox proof
   harness.
3. Add npm first, then pnpm, Yarn Classic, and modern Yarn as separate lock and
   materialization profiles. Do not hide their differences behind one parser.
4. Add the pure JS/TS build-plan runner and generated-output receipts. Leave
   native addons explicitly unsupported.
5. Publish shared semantic fixtures plus manager wrappers. Supply their schemas
   and expected digests to the independent Python implementation; a new Python
   repository or implementation is not a Node delivery prerequisite.
6. Require cross-implementation compatibility only at the protocol/fixture
   boundary. Version negotiation must fail closed when a peer does not support
   the policy, graph, artifact-manifest, or diagnostic schema in use.

## Key findings and decisions

- **Lock closure and source closure are different proofs.** Locks select a graph;
  exact raw artifacts, recursive member inspection, metadata reconciliation,
  build closure, and offline replay establish the source boundary.
- **Caches are derived state.** npm explicitly declines to treat its cache as
  durable storage; pnpm stores are a trust domain; Yarn formats caches for its
  own version/linker. None can replace manager-neutral immutable evidence.
- **Lifecycle suppression is necessary but incomplete.** npm has an implicit
  `binding.gyp` path, pnpm hook files operate outside ordinary script controls,
  and Yarn plugins/Git packaging extend resolution or execution. The adapter
  needs a closed extension registry and explicit build DAG.
- **Generated JS can be source evidence.** Exact shipped JavaScript is eligible
  as immutable generated text without pretending its upstream TypeScript build
  was reproduced. Curator-generated JS instead needs full causal lineage.
- **A pure label does not make a source payload.** A wheel, `none-any` tag,
  `Root-Is-Purelib`, package extension, integrity checksum, or platform selector
  never overrides recursive byte classification.
- **Python has two closures.** Installation dependencies and PEP 517 build
  dependencies/hooks must both close. Dynamic backend requirements are accepted
  only when their expected result was already locked and captured.
- **Runtime identity is part of dependency identity.** Node ABI/components and
  Python implementation/ABI/tags affect selection and behavior even in profiles
  that prohibit native dependency payloads.
- **Share semantics, not location.** The independent Python implementation and
  the Node adapter need matching canonical records and fixtures, not imports,
  common virtual environments, caches, or repository layout.

## Acceptance mapping

| Task requirement | Evidence in this document |
| --- | --- |
| Compare Node package-manager and Python packaging closure models | Node manager and Python lock/requirements comparison tables; package/container, graph, and cache sections |
| Analyze lifecycle hooks, build backends, generated code, native addons, wheels, and offline behavior | Dedicated Node lifecycle/generated/native/cache sections; Python sdist/backend/wheel/runtime sections; offline replay protocol |
| Recommend shared or separate policy layers | Executive decision, architecture sketch, and implementation-ready policy split |
| Prove recursive immutable source closure and offline behavior | Admission predicate, compositional proof, checkpoint chain, and OS-denied clean replay protocol |
| Reject native/compiled payloads and undeclared hooks | Shared artifact-policy dependency, Node/Python deny rules, hook policies, diagnostics, and negative fixtures |
| Define checkpoints and diagnostics | `C0`-`C7` table and stable closure diagnostic table with precedence |
| Record unsupported cases and conformance fixtures | Explicit unsupported lists and `S01`-`S08`, `N01`-`N13`, `P01`-`P13` fixture tables |
| Avoid repository-co-location assumptions | Context, executive decision, implementation split, and protocol-only Python handoff |

## Fact-check record and sources

Policy statements above are normative recommendations. Ecosystem behavior was
checked on 2026-08-11 against current primary documentation. The principal
sources are linked at the claims they support; this list records the review
surface:

- Local policy and repository evidence:
  [source-closure specification](../.spec/skill-facing-cli-source-closure.md),
  [accepted artifact taxonomy](260811_compiled-artifact-taxonomy-and-deny-policy.md),
  and
  [accepted inventory](260811_inventory-language-and-reference-surfaces.md).
- npm:
  [package-lock](https://docs.npmjs.com/cli/v11/configuring-npm/package-lock-json/),
  [`npm ci`](https://docs.npmjs.com/cli/v11/commands/npm-ci/),
  [configuration, including offline/script behavior](https://docs.npmjs.com/cli/using-npm/config/),
  [lifecycle scripts](https://docs.npmjs.com/cli/using-npm/scripts/),
  [`npm rebuild`/implicit node-gyp](https://docs.npmjs.com/cli/v10/commands/npm-rebuild/),
  [cache](https://docs.npmjs.com/cli/v11/commands/npm-cache/),
  [`npm pack`](https://docs.npmjs.com/cli/v11/commands/npm-pack/), and
  [`package.json`](https://docs.npmjs.com/cli/v11/configuring-npm/package-json/).
- pnpm:
  [install/frozen/offline behavior](https://pnpm.io/cli/install),
  [fetch](https://pnpm.io/cli/fetch),
  [hook file](https://pnpm.io/pnpmfile),
  [build settings](https://pnpm.io/settings/build), and
  [store settings](https://pnpm.io/settings/store).
- Yarn:
  [Classic lockfile](https://classic.yarnpkg.com/en/docs/yarn-lock),
  [Classic install](https://classic.yarnpkg.com/en/docs/cli/install),
  [Classic offline mirror](https://classic.yarnpkg.com/blog/2016/11/24/offline-mirror/),
  [modern install](https://yarnpkg.com/cli/install),
  [configuration](https://yarnpkg.com/configuration/yarnrc),
  [caching](https://yarnpkg.com/features/caching),
  [lifecycle scripts](https://yarnpkg.com/advanced/lifecycle-scripts),
  [manifest/build controls](https://yarnpkg.com/configuration/manifest), and
  [extensibility](https://yarnpkg.com/features/extensibility).
- Node:
  [native addons](https://nodejs.org/api/addons.html) and
  [runtime/process identity](https://nodejs.org/api/process.html).
- Python packaging:
  [`pylock.toml`](https://packaging.python.org/en/latest/specifications/pylock-toml/),
  [requirements files](https://pip.pypa.io/en/stable/reference/requirements-file-format/),
  [hash-checking mode](https://pip.pypa.io/en/stable/topics/secure-installs/),
  [repeatable/offline installs](https://pip.pypa.io/en/latest/topics/repeatable-installs/),
  [`pyproject.toml`](https://packaging.python.org/en/latest/specifications/pyproject-toml/),
  [PEP 517](https://peps.python.org/pep-0517/),
  [sdists](https://packaging.python.org/en/latest/specifications/source-distribution-format/),
  [wheels](https://packaging.python.org/en/latest/specifications/binary-distribution-format/),
  [platform tags](https://packaging.python.org/en/latest/specifications/platform-compatibility-tags/),
  [`packaging.tags`](https://packaging.pypa.io/en/latest/tags.html),
  [extension modules](https://docs.python.org/3/extending/building.html),
  [`sys`](https://docs.python.org/3/library/sys.html), and
  [build details](https://packaging.python.org/en/latest/specifications/build-details/).
- Illustrative tool-specific Python locks:
  [uv project layout](https://docs.astral.sh/uv/concepts/projects/layout/),
  [uv lock internals](https://docs.astral.sh/uv/reference/internals/metadata/),
  [PDM lockfile](https://pdm-project.org/latest/usage/lockfile/),
  [Poetry lock usage](https://python-poetry.org/docs/basic-usage/), and
  [Pipenv lock format](https://pipenv.pypa.io/en/stable/pipfile.html).

Local availability probes were run as standalone processes for context, not as
conformance claims:

| Command | Exit | Observation |
| --- | ---: | --- |
| `node --version` | 0 | `v25.6.1` |
| `npm --version` | 0 | `11.10.1` |
| `yarn --version` | 0 | `1.22.22` (Classic) |
| `python3 --version` | 0 | `Python 3.14.4` |
| `python3 -m pip --version` | 0 | `pip 26.1` |
| `pnpm --version` | 127 | `pnpm` is not installed locally; pnpm claims were source-verified and no local pnpm conformance run is claimed |

The document does not claim that manager flags alone prove isolation, that a
hash proves source class, that the existing independent Python implementation
already meets this policy, or that source-built native extensions are supported
by the pure-source profiles. Those would exceed the checked evidence.
