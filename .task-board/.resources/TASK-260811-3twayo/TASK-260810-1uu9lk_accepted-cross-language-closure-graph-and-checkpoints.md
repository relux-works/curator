# Cross-language source-closure graph and audit checkpoints

Status: solution-architecture decision prepared for review

Board task: `TASK-260810-1uu9lk`

Date: 2026-08-11

## Decision

Curator should use one language-neutral, selection-neutral capture multigraph,
bind exact platforms and external toolchains in a selection-specific typed-edge
overlay, derive an active graph from the capture plus that overlay, and derive
one deterministic execution DAG from the active graph. The graph schemas are
shared; acquisition, lock parsing, materialization, and supported execution
profiles remain ecosystem-specific.

The graph has ten closed node kinds and eleven closed edge kinds. It records
both selected and pruned dependency relations, source and generated inputs,
host and target execution domains, FFI and subprocess contracts, build tools,
plugins and hooks, external toolchains and SDKs, target platforms, and locally
produced outputs. Every selected relation must terminate at an immutable node.
Unknown node or edge kinds, unresolved conditions, implicit executable
extensions, or observed inputs absent from the graph fail closed.

One canonical checkpoint sequence, `C0.profile` through `C7.publish`, binds the
policy, resolution, immutable capture, artifact admission, closed graph,
execution plan, offline replay, and protected outputs. Pre-C5 executable
evidence derivation is not an exception: every manifest evaluation, Cargo
vendor invocation, mirror replay, or manager metadata invocation requires a
pre-issued derivation permit and a causal receipt bound to already admitted
inputs and a C0-selected toolchain. Each checkpoint hashes its exact
predecessor and canonical payload. The checkpoint chain, rather than an
installed tree or package-manager cache, is the closure identity.

The complete graph may contain explicitly non-ordering cycles, such as a Node
runtime dependency cycle or peer relationship. The derived execution graph may
not contain a cycle. Curator reports a deterministic `closure_build_cycle`
diagnostic and starts no action instead of collapsing or heuristically breaking
an ordering cycle.

This model does not admit a new Python adapter, verified binaries, Kotlin,
Dart, .NET, non-SwiftPM native build systems, Rust build scripts or proc macros,
Node native addons, or active Swift plugins and macros in this cycle.

## Governing requirements

The labels below are local traceability shorthand; the cited specification
headings and numbered items remain authoritative.

| Label | Authoritative requirement |
| --- | --- |
| `SCOPE` | `.spec/skill-facing-cli-source-closure.md` — **Current delivery scope**: Rust, Node/TypeScript, Swift, and SwiftPM-supported C, C++, Objective-C, and Objective-C++; Go baseline; independent Python reference; Kotlin, Dart, and .NET deferred. |
| `SCI-1` | **Source closure invariant 1** — enumerate every transitive dependency and bind immutable identity. |
| `SCI-2` | **Source closure invariant 2** — retain lock, checksum, snapshot, build metadata, and toolchain evidence for audit/checkpoints. |
| `SCI-3` | **Source closure invariant 3** — rebuild offline from captured closure without new network resolution. |
| `SCI-4` | **Source closure invariant 4** — hooks, plugins, generators, and build scripts cannot add undeclared inputs or escape the trusted boundary. |
| `SCI-5` | **Source closure invariant 5** — one explicit mixed-language graph includes build order, toolchains, target platform, and FFI boundaries. |
| `SCI-6` | **Source closure invariant 6** — unsupported ecosystems or shapes fail closed or use a narrower source-only profile. |
| `VCAP` | **Vendored compiled artifact prohibition** — dependency payloads may not contain precompiled executable/intermediate code; selected external toolchains and protected local outputs are distinct causal roles. |
| `RR-GRAPH` | **Required research** — mixed-language graph, FFI, build-order, toolchain checkpoint, and reusable conformance semantics. |
| `DELIVERY` | **Delivery completion** — atomic implementation tasks; Rust, Node/TypeScript, SwiftPM/C-family implementations; shared conformance; reviewed integration. |

## Prerequisite outcomes consumed

| Accepted outcome | Decision consumed by this architecture |
| --- | --- |
| `TASK-260810-1veyfw` — language and reference-surface inventory | Go is the only implemented adapter baseline. Swift and real Swift-to-C targets exist; currency exchange supplies a real Go-to-Swift subprocess boundary. Rust and Node require purpose-built seed fixtures. Python is an independent reference. Revision state, external system commands, generated local products, and source payloads are distinct authorities. |
| `TASK-260810-29vk09` — compiled-artifact taxonomy | Trust role precedes class. Recursive byte/structure inspection and deny dominance are shared services. `artifact-manifest-v1`, exact toolchain checkpoints, protected local-output receipts, stable diagnostics, closed limits, ELF `ET_DYN` resolution, and the absent `verified-binary-v1` seam are normative. |
| `TASK-260810-2n3sbi` — Node/TypeScript and Python closure | Share schemas, policy, sandbox semantics, diagnostics, and fixtures, not implementation state. npm, pnpm, Yarn Classic, and modern Yarn require separate lock/materialization profiles. Caches are derived. Lifecycle execution is disabled during materialization. Generated JS has shipped-input and locally-generated roles. The common checkpoint sequence is `C0`-`C7`. |
| `TASK-260810-3urqbl` — Rust/Cargo closure | Preserve lock-superset and active target graphs. Admit raw registry/Git/path origins before Cargo. Treat vendoring as pinned `cargo-vendor-transform-v1`, not provenance. Model host executable units even though v1 rejects build scripts, build dependencies, proc macros, native links, cross targets, and ambient Cargo configuration. |
| `TASK-260810-zddzh7` — SwiftPM/C-family closure | Wrap SwiftPM rather than trusting it as the boundary. Separate Swift from Clang targets. Capture exact revisions and kind-preserving mirrors. Parse module maps and verify header reads. Bind Swift/Clang/SDK/platform/C++/Objective-C identities. Reject active plugins/macros, binary targets, unsafe flags, and untrusted system libraries. |

No prerequisite outcome remains unresolved. Their differences are ecosystem
profile data, not competing security models.

## Canonical graph

### Graph envelope

The canonical model is four records rather than one target-contaminated
envelope:

```text
CaptureGraph = {
  schema_id,
  profile_id,
  policy_ids,
  root_node_ids[],
  node_ids[],
  edge_ids[],
  condition_declarations[],
  artifact_manifest_ids[],
  captured_graph_id
}

SelectionContext(S) = {
  schema_id,
  product_node_ids[],
  platform_roles,
  requested_features[],
  marker_cfg_values,
  peer_context,
  evaluator_ids[],
  selection_context_id
}

SelectionBinding(S) = {
  schema_id,
  captured_graph_id,
  selection_context_id,
  binding_node_ids[],
  binding_edge_ids[],
  selection_binding_id
}

ActiveGraph(S) = {
  schema_id,
  captured_graph_id,
  selection_context_id,
  selection_binding_id,
  node_activations[],
  edge_activations[],
  non_ordering_sccs[],
  active_graph_id
}
```

`captured_graph_id` covers the conservative lock/resolution superset, including
inactive optional, feature, target-predicate, peer, dormant extension, and
development entries that entered the captured payload. It includes condition
expressions, abstract action/tool/platform slots, and evaluator declarations,
but no requested product/feature/marker/peer/target values, concrete platform
or external-toolchain nodes, `targets` or concrete tool-binding edges, or
selected/pruned result. A lock entry may supply a selection-neutral structural
instance key, including an npm peer-resolution key, only when that key exists
in the authoritative lock independently of the current request.

The IDs resolve in canonical schema-checked record tables carried by the
closure evidence bundle; they are not ambient registry lookups. Capture owns
the records named by `node_ids`/`edge_ids`, and a binding owns only the records
named by `binding_node_ids`/`binding_edge_ids`. The graph and binding hashes use
the sorted IDs after independently recomputing every referenced record ID.

`SelectionContext(S)` names exact intrinsic `target_platform` node IDs through
the closed `platform_roles` map (`target` is required; `host` is required when
different), plus requested product, feature, marker/cfg, peer, and evaluator
values. A platform ID is computable before the overlay because it hashes only
the platform node's intrinsic payload. It is not valid merely because the ID
has the right shape: it must resolve exactly once in `SelectionBinding(S)`.

`SelectionBinding(S)` is the only authority allowed to add selection-specific
graph records. Its binding-node table contains exact `target_platform` nodes
and manager-selected external `toolchain_component` nodes: pre-C5 evidence
tools must already be checkpointed at C0, while build-only tools may first be
selected and fingerprinted during C4 closure before any build plan or action.
Its binding-edge table
uses the same closed edge schema and ID function as capture, and contains the
concrete `targets`, `uses_tool`, toolchain-scoped `requires`, and selected SDK/
interop bindings whose endpoints depend on `S`. It cannot add a package,
source, target unit, action, generated artifact, interop declaration, or
expected output, and it cannot replace a capture record. Abstract conditions
and declaration relationships remain in capture; evaluated peer placement,
resolved Rust features, selected Swift manifest, concrete target/tool bindings,
and selected/pruned state belong to the binding plus `ActiveGraph(S)`.

`active_graph_id` covers the exact selected capture records plus the complete
binding overlay and therefore the exact product, target, feature, condition,
peer, host/target, SDK, and toolchain selection used for planning. All four IDs
are required: scanning only the active build can miss rejected bytes in the
captured superset, while hashing only the superset cannot identify the exact
build. Two selections over unchanged captured bytes and declarations therefore
have the same `captured_graph_id` and different `selection_context_id`,
`selection_binding_id`, `active_graph_id`, checkpoint, plan, and output-receipt
identities.

Individual source leaves and nested archive members remain in
`artifact-manifest-v1`; they are not expanded into millions of graph nodes.
Every graph source node names the exact artifact-manifest digest and selected
leaf set or tree projection. This preserves byte-level evidence without making
the dependency graph an archive listing.

### Closed node kinds

Every node carries `kind`, one globally unique graph-local `logical_key`, and a
closed intrinsic payload. A logical key is derived from immutable declaration
evidence; it never contains another node hash and cannot be reassigned after
graph closure. Dependency, ownership, producer, consumer, tool, interop, and
platform relationships are forbidden in node payloads and live only in the
typed edge table. The kind-specific fields below are mandatory rather than an
open extension map.

| Node kind | Purpose and required identity |
| --- | --- |
| `command_product` | One declared skill-facing command or bundle product. Intrinsic identity includes manifest/schema, skill and command keys, entry-point contract kind, and declaration-fragment digest. Selected target and publication layout are expressed by `declares`, `targets`, and `publishes` edges. Multiple commands, such as `exchange` and `exchange-scraper`, are separate products joined by explicit package/runtime edges. |
| `package_instance` | One captured package instance, not merely name/version. Intrinsic identity includes ecosystem/manager, normalized source ID and origin, authoritative lock-instance key, name/version, artifact or snapshot digest, and workspace/path identity. A lock-defined npm peer instance remains distinct, but evaluated peer/feature/marker/condition/target activation is recorded only in `ActiveGraph(S)`. |
| `source_set` | One immutable root, workspace/path tree, raw package artifact, admitted package projection, or generated source input. Identity references immutable origin plus `artifact-manifest-v1`, canonical tree/member selection, source grammar/profile, and trust role. Shipped generated text is a `source_set` with class `source.generated_text`. |
| `target_unit` | The smallest compiler/package-manager unit Curator lets one tool plan as a unit. Intrinsic identity includes target name/kind, declaration-fragment digest, language set, host or target execution domain, declared feature/condition expressions, and expected output class. Package membership, sources, activation context, platform, and outputs are edges or active-graph records. Swift and C-family sources may not share one SwiftPM target; a supported Clang target may contain C/C++/Objective-C/Objective-C++ under a validated profile. |
| `action` | An executable materialize, generator, hook, build-script, plugin, macro, compiler, linker, packager, or publisher step in the derived build DAG. Intrinsic identity contains subtype, host/target domain, argv and working-directory templates, named tool/read/write slots, and environment/process/network policy IDs. It contains no executable, input, output, producer, consumer, or ordering node reference; C5 binds slots through edges and emits the exact physical command. Rejected extensions are still represented before policy rejection so they cannot remain hidden. Pre-C5 evidence derivations use the separate permit/receipt schema below. |
| `generated_artifact` | A declared locally generated source, metadata, module map, header, package-manager state, or intermediate used by a later action. Intrinsic identity includes logical path/slot, expected class/grammar, role, and declaration digest. Producer and consumers are expressed only by `produces`, `reads`, or `requires` edges. Pre-shipped generated text is a `source_set`, not this node. |
| `interop_boundary` | One typed language or process interface shared by provider and consumer. Intrinsic identity includes boundary mode, provider/consumer language classes, ABI/runtime/protocol schema, header/module/interface/symbol contract, and calling-convention/link/load semantics. Provider, consumer, platform, and toolchain/SDK relationships are typed edges. |
| `toolchain_component` | A Curator-selected compiler, runtime, manager, linker, archiver, SDK/sysroot component, or complete toolchain set outside the dependency closure. Identity includes policy selector, contained relative path, full content fingerprint, version/build output, links, platform/ABI/SDK facts, and time-of-use recheck rule. Exact external components are binding-overlay nodes, not capture nodes, because manager/target selection establishes their role. A source-provided tool is never relabeled as this kind. |
| `target_platform` | The exact build/run destination: OS, architecture, ABI/libc/runtime, target triple/cfg/tags/conditions, minimum deployment/runtime version, SDK/sysroot, language modes, and tuning. It is a binding-overlay node; capture retains only platform predicates and abstract slots. Host and target are separate nodes when they differ; v1 Rust rejects that split. |
| `output_artifact` | An immutable expected-output declaration. Intrinsic identity includes logical path/slot, expected class, role, compatibility predicate, and declaration digest. Producer, consumers, platform, and publication are edges. C6 records observed path/class/size/hash in a separate `produced_artifact_observation`; it never mutates this node. A pre-existing dependency binary cannot become this node by copying it into an output directory. |

### Node and edge reference validation

Canonical graph validation runs before any graph ID is accepted. Validation
uses the disjoint union of capture and binding tables while preserving their
different authorities:

1. recompute every `node_id` from `kind`, `logical_key`, and intrinsic payload;
2. require `logical_key` and `node_id` to be unique, rejecting exact duplicate
   records rather than accepting two authorities;
3. reject a duplicate `logical_key` or `node_id` within or across capture and
   binding tables; reject a binding node whose kind is not `target_platform` or
   external `toolchain_component`, or whose trust evidence is absent from the
   C0 evidence-tool checkpoint or C4 build-tool selector/fingerprint record;
4. recompute every `edge_id`, require one unique `edge_key` across both tables,
   and require both endpoint IDs to resolve exactly once to an allowed kind;
   binding edges may cross from a capture node to a binding node but may not
   replace or duplicate a capture semantic edge;
5. require every `SelectionContext.platform_roles` value to resolve to exactly
   one binding `target_platform` of the expected role, and require one explicit
   typed `targets` edge for every selected product/target/action/toolchain/
   output/boundary slot that declares that role; dangling, duplicate, wrong-kind,
   or role-mismatched platform references are canonical failures;
6. reject a relation copied into a node payload, an undeclared action slot, a
   slot bound by zero or multiple selected capture/binding edges, and duplicate
   semantic edges;
7. require every active node/edge activation to reference the unchanged capture
   table, every selected/pruned conditional edge exactly once, and the exact
   binding overlay named by `selection_binding_id`; binding records are all
   selected and are never represented as capture activations; and
8. sort capture and binding nodes by `(kind, logical_key, node_id)`, capture and
   binding edges by `(kind, edge_key, edge_id)`, and activation records by
   referenced ID.

Because nodes never hash endpoints and observed outputs never rewrite nodes,
action/output relations, runtime cycles, platform/tool bindings, and interop
provider/consumer pairs can only form cycles in the typed capture/binding edge
tables. They cannot form a recursive ID.

### Build tools, hooks, plugins, and macros

There is no ambiguous `build_tool` trust shortcut:

- A manager-selected `cargo`, `rustc`, `node`, `npm`, `swift`, `swiftc`,
  `clang`, linker, SDK tool, or equivalent is a `toolchain_component`.
- A build tool supplied as source by a package is a `package_instance` plus
  `source_set`, `target_unit`, build `action`, and receipted `output_artifact`.
  A later action may `uses_tool` that output only after its receipt exists.
- A pure-JavaScript TypeScript compiler is executable build logic even though
  the bytes are source text. It is invoked only by a declared host `action`.
- Rust `build.rs` and proc macros, Node lifecycle or implicit `node-gyp`, Swift
  build plugins/macros, Yarn plugins, and pnpm hook files are typed `action`
  candidates. Current profiles reject the unsupported active subtypes before
  compilation or invocation; their declarations and reachability stay in the
  graph evidence.

This distinction makes trust causal: package location cannot establish an
external toolchain, and output location cannot establish local production.

### Closed edge kinds

Semantic dependency edges point from consumer to requirement. `produces` and
`provides_interop` point from provider to product/boundary. The execution
projection reverses consumer-to-requirement relations when deriving
provider-before-consumer order.

| Edge kind | Endpoints and required fields | Ordering meaning |
| --- | --- | --- |
| `declares` | Package/product -> target, action, boundary, or source set; manifest field and evidence origin. | None by itself. |
| `resolves_to` | Package instance -> immutable source set; lock/origin/checksum mapping and artifact-manifest ref. | Capture/admission precedes package use. |
| `requires` | Product/package/target/action/boundary -> package/target/output/toolchain; scope enum `runtime`, `build`, `development`, `peer`, `optional`, `workspace`, `tool`, `toolchain`, or `package_artifact`; exact condition and dependency-kind evidence. | Ordering only for the selected scopes that consume a prior build/materialization output. Runtime/peer relations can be non-ordering. |
| `reads` | Target/action -> source set, generated artifact, output, or toolchain resource; declared path/projection and read class. | Producer precedes reader when the input is generated/local output. Immutable source/toolchain inputs already exist before execution. |
| `uses_tool` | Action -> external toolchain component or receipted local tool output; executable-relative path and invocation role. | Local tool producer precedes use; external toolchain is a precondition. |
| `targets` | Product/target/action/toolchain/output/boundary -> target platform; exact role and selection evidence. Concrete `targets` records live only in `SelectionBinding(S)`; capture keeps the unevaluated platform predicate/slot. | Platform is a precondition and part of binding/active-graph/plan identity, not capture or an endpoint node's intrinsic identity. |
| `produces` | Action -> generated artifact or output artifact; declared path/class and write-set member. | Producer precedes every consumer and publication. |
| `provides_interop` | Provider target/output/toolchain -> interop boundary; exported header/module/symbol/protocol evidence. | Provider build precedes compile/link consumer when the boundary requires provider output. |
| `consumes_interop` | Consumer target/action/output -> interop boundary; import/link/load/invoke use and ABI expectation. | Boundary provider precedes compile/link consumer; runtime-only invocation is non-ordering unless packaging requires both outputs. |
| `invokes` | Product/output/action -> product/output/tool action; runtime protocol, executable resolution, arguments, and environment contract. | Runtime-only by default. A separate selected `requires`/`produces` edge expresses build or bundle order. |
| `publishes` | Command product -> output artifact or bundle; destination layout and entry-point mapping. | All producing actions precede publication. |

Every conditional capture edge has the exact normalized
condition/marker/cfg/platform/feature/peer expression, the required versioned
evaluator identity, and the manifest/lock/metadata field that supplied it. It
does **not** have `state`, an evaluation context, a result, or a reason.

`ActiveGraph(S)` carries exactly one activation record for every conditional
capture edge: `state=selected|pruned`, the bound `selection_context_id`,
evaluator ID, normalized result, and stable reason enum. The binding overlay is
referenced once and contributes all of its concrete nodes and edges to the
active graph; copying those records into activations is invalid. Silently
deleting a pruned dependency or a target binding is invalid. A different target
or feature selection produces a different binding, active graph, and checkpoint
while retaining the same capture graph when the underlying captured bytes and
declarations are unchanged.

### No-hidden-edge invariant

For every selected capture or binding edge, both endpoints exist and carry
immutable identity. For every action, the executable, working directory,
arguments, environment, read set, process set, network policy, write set,
target, tools, and outputs are closed by its intrinsic slot schema, the selected
capture-plus-binding edge union, and C5 physical command binding.
After execution, observed reads/processes/writes and manager metadata reconcile
exactly with the plan without changing any node or edge.

The following cannot repair a missing edge:

- an installed-tree layout, hoist, virtual store, PnP file, `.build` checkout,
  Cargo cache, venv, or ambient package-manager cache;
- a package label, extension, executable bit, integrity checksum, wheel tag,
  SwiftPM target kind, or source directory named `toolchain`;
- an implicit lifecycle convention, `binding.gyp`, build backend discovery,
  module-map escape, Cargo configuration search, plugin registry, or host
  `PATH`; or
- a successful compiler or package-manager run that happened to find an
  undeclared input.

Unknown security-relevant metadata or an observed extra endpoint is
`closure_graph_incomplete`, a more specific accepted ecosystem diagnostic, or
an execution-boundary diagnostic. There is no warning-and-continue state.

## Cycle handling and deterministic build order

### Graph projections

Curator derives four canonical projections:

1. `G_capture`: all captured package/source nodes, unevaluated conditional
   declarations, and the lock/resolution superset. It has no requested target,
   feature, marker, peer, or selected/pruned state.
2. `G_active(S)`: exact selected capture subgraph unioned with
   `SelectionBinding(S)`, including product, target, feature, peer, marker,
   platform, SDK, and toolchain bindings for selection context `S`.
3. `D_build(S)`: action nodes plus ordering arcs required to materialize,
   generate, compile, link, bundle, and publish `G_active(S)`.
4. `G_runtime(S)`: entry points, receipted outputs, runtime dependencies,
   subprocess protocols, dynamic-load declarations, and target compatibility.

`G_capture`, the capture-plus-binding `G_active`, and `G_runtime` may contain
cycles. `D_build` must be a DAG.

### Permitted cycles

A cycle is permitted only when every participating edge is explicitly marked
non-ordering by the closed edge policy. Examples are:

- Node runtime/package cycles that the selected manager materializes as one
  closed package set;
- peer relationships that constrain compatible package instances but do not
  produce one another;
- a runtime subprocess/protocol relationship that does not require one process
  to be generated by the other; and
- compiler-internal source/module cycles represented inside one `target_unit`,
  where the ecosystem compiler owns the unit atomically.

Permitted cycles are not discarded. Curator records strongly connected
components as sorted node IDs plus sorted internal edge IDs and hashes that
record into `ActiveGraph(S)`. The selection-neutral capture retains the edge
declarations but no selection-dependent SCC result.

### Rejected cycles

An SCC containing more than one action, or an action self-edge, is rejected
when any edge requires prior output or execution. Examples include mutually
dependent generators, a source-built tool depending on the output it must
produce, two packages requiring each other's compiled library, or an FFI
provider that requires its consumer's linked output.

Curator does not collapse an ordering SCC, choose a package-manager incidental
order, or break a cycle by lexical preference. It emits
`closure_build_cycle` with:

- sorted SCC action IDs;
- sorted internal ordering-edge IDs;
- a `cycle_digest` over that canonical set;
- affected product and target IDs; and
- the last successful checkpoint.

No `C5.plan` checkpoint or build action is issued for the rejected graph.

### Stable topological waves

For an acyclic `D_build`, build order is derived with Kahn's algorithm:

1. derive requirement-to-consumer arcs from the closed edge rules;
2. deduplicate exact arcs and calculate in-degree;
3. select every zero-in-degree action as the next parallel wave;
4. sort each wave by canonical action node ID, never discovery order;
5. remove the whole wave and repeat; and
6. hash the ordered wave array and sorted ordering-edge table into
   `build_plan_id`.

Different filesystem enumeration, map iteration, archive member, lockfile, or
adapter emission orders therefore produce the same plan. A package manager may
execute independent actions in another interleaving only when the checkpointed
policy permits parallelism and the observable write sets are disjoint; the
canonical wave plan remains the audit order.

## Mixed-language and FFI boundaries

An `interop_boundary` is a first-class node because a language edge is more
than a package dependency. Its closed mode enum is:

| Mode | Required contract |
| --- | --- |
| `c_abi` | Headers/module maps, symbol ownership, calling convention, layout-relevant compiler/target facts, link/load mode, and C provider/consumer targets. |
| `cxx_interop` | C++ headers/modules, C++ standard/library, compiler ABI, Swift interoperability mode where used, transitive Swift target opt-in, supported API-shape profile, and target/toolchain identity. |
| `objc_runtime` | Objective-C or Objective-C++ headers/modules, runtime and SDK identity, framework/system edges, target platform, and any C++ exposure mode. |
| `native_link` | Provider output/source role, exact library/search input, symbols, linker and target ABI. Dependency prebuilt libraries remain forbidden; selected SDK libraries are external toolchain resources. |
| `dynamic_load` | Exact produced/selected module, loader/runtime ABI, target compatibility, and load path. Current dependency-native Node/Python modules are rejected before this boundary can authorize them. |
| `host_extension` | Host target, executable plugin/macro/proc-macro/build-tool identity, tool dependencies, I/O policy, and generated outputs. Current Rust and Swift profiles reject active instances; Node permits only separately declared source build actions. |
| `subprocess_protocol` | Provider command/output, consumer command/output, protocol/schema/version, argv/environment/working directory, runtime executable resolution, and publication bundle. This models currency exchange's Go-to-Swift JSON subprocess without mislabeling it FFI. |

Rules:

1. Provider and consumer attach through separate `provides_interop` and
   `consumes_interop` capture edges. Selection-specific `targets`, `uses_tool`,
   and toolchain/SDK binding edges live in `SelectionBinding(S)`; their
   endpoints must be compatible with both sides of the boundary.
2. Header, module-map, source, library, framework, executable, protocol schema,
   and generated binding inputs are graph nodes or artifact-manifest leaves.
3. Compile/link boundaries add ordering arcs from provider production to
   consumer action. Runtime-only boundaries do not invent build order; bundle
   or publication requirements express any required co-publication.
4. An FFI declaration cannot change artifact trust role. A vendored `.a`,
   `.so`, `.dylib`, `.framework`, `.xcframework`, `.node`, `.pyd`, object,
   `.swiftmodule`, `.rlib`, or equivalent still fails `VCAP`.
5. A system-library edge is accepted only when every referenced byte lies in a
   manager-selected, fingerprinted SDK/sysroot/toolchain. Arbitrary host paths
   remain `artifact_toolchain_untrusted`.

Initial profile outcomes:

| Boundary | Current decision |
| --- | --- |
| Swift -> C target | Supported with separate targets, admitted sources/headers, module evidence, target identity, and observed compiler reads. |
| Swift -> C++ | Restricted to a pinned toolchain/platform/profile with `.interoperabilityMode(.Cxx)` throughout the required Swift target chain and passing API fixture. |
| Swift -> Objective-C / Objective-C++ | Restricted initially to accepted Darwin SDK/runtime fixtures; direct C++ exposure also follows C++ restrictions. |
| Rust -> native library / build script / proc macro | Representable but rejected in `rust-source-v1`; future host/native capabilities require separate policy and conformance. |
| Node -> native addon / Wasm | Representable but rejected in the pure-source Node profile. |
| Go `exchange` -> Swift scraper subprocess | Supported graph shape after both source closures are independently admitted, built, receipted, bundled, and joined by an explicit protocol edge. |

## Deterministic identities

### Canonical encoding

Use Curator's strict canonical JSON/CCJ rules and domain-separated SHA-256:

```text
ID(label, payload) = "sha256:" + hex(
  SHA256(UTF8(label) || 0x00 || CCJ(payload))
)
```

Portable identity excludes timestamps, temporary roots, host-absolute paths,
directory enumeration order, archive member order, and display-only logs.
Paths are canonical portable UTF-8/NFC paths. Closed enums are lowercase
strings. Sets become bytewise sorted arrays. Maps have sorted keys. Integers
obey the protocol safe-integer bound. Empty collections are explicit rather
than omitted when the schema requires them.

### Identity layers

| Identity | Canonical payload |
| --- | --- |
| `artifact_manifest_id` | Accepted `artifact-manifest-v1`, including raw payload/origin, detector/policy/limit/profile IDs, complete canonical member findings, classes, and decision. |
| `node_id` | Node kind, unique immutable logical key, and closed intrinsic payload. No endpoint, producer, consumer, platform, tool, or observation reference is permitted. |
| `edge_id` | Edge kind/key, endpoint node IDs, unevaluated condition declaration, semantic fields, and origin evidence. Selection state is excluded. |
| `captured_graph_id` | Schema/policy/profile, sorted capture node/edge IDs, artifact-manifest IDs, and root declarations. Requested product/feature/marker/peer/target values and evaluation results are excluded. |
| `selection_context_id` | Requested product IDs, closed platform-role-to-intrinsic-node-ID map, features/default-feature mode, marker/cfg values, peer context, and evaluator IDs. |
| `selection_binding_id` | Capture and selection-context refs plus sorted exact binding-node and binding-edge IDs. Binding nodes are concrete target platforms and selected external toolchains; binding edges use the same typed edge encoding as capture. |
| `active_graph_id` | Captured graph, selection-context, and selection-binding refs; sorted capture node/edge activations; exact selected/pruned results/reasons; and non-ordering SCC records over the selected capture plus binding union. |
| `derivation_permit_id` | Exact predecessor evidence ID, admitted input manifests, C0 toolchain checkpoint and executable, command/cwd/environment, host/target, process/read/write/network policy, expected evidence outputs, and time-of-use recheck rule. |
| `derivation_receipt_id` | Permit ID, pre/post toolchain rechecks, actual process/read/environment/write/network audit, exit result, output evidence manifests/digests, decision, and next causal head. |
| `build_plan_id` | Active graph ref (therefore exact binding overlay), sorted action payloads, sorted ordering arcs, canonical waves, declared outputs, and execution-policy identity. |
| `checkpoint_id(Cn)` | Checkpoint schema/name, exact predecessor ID (`null` only for C0), and checkpoint payload. |
| `closure_id` | Domain `curator-source-closure-v1` plus the exact `C5.plan` checkpoint ID. Because C5 chains C0-C4, one reference binds the complete pre-execution closure. |
| `produced_artifact_observation_id` | Expected output node ID, producer action and `produces` edge IDs, and canonical observed path/class/size/SHA-256. The enclosing C6 execution receipt binds the observation and causal action audit; neither record hashes the other recursively. |
| `execution_receipt_id` | `closure_id`, C6 evidence, observed action/order/I/O/toolchain rechecks, and sorted produced-observation IDs. |
| `publication_receipt_id` | C7 payload including execution receipt, independently derived expected cache input, published outputs, protection result, and canonical receipt hash. |

Changing a source byte, lock or condition, package origin, peer context,
generated input, plugin set, graph edge, build order, policy, target, runtime,
toolchain/SDK, command, environment, sandbox rule, or expected output changes
the relevant identity and every downstream checkpoint. A self-consistent
receipt or hash outside the manager-protected boundary is not provenance.

Output hashes are exact local-production evidence. This architecture does not
claim universal cross-toolchain or cross-host byte-for-byte reproducibility.
Reuse requires the same independently derived logical input and a protected
exact hit. Where a profile claims repeated-build byte identity, its
conformance fixture must prove it for the exact checkpointed toolchain and
target.

### Selection-neutral two-target golden

`CGP05` uses one capture with an unevaluated Linux target predicate and two
exact bindings: Darwin arm64 and Linux x86_64. The concrete platform nodes and
`targets` edges appear only in their respective binding overlays. The fixture
anchor is a published domain-separated sentinel for the identical C0-C3 prefix;
it is conformance data, not a production checkpoint schema. Every other record
below is the exact production CCJ shape named by its label.

```text
name=cgp05.root
label=curator-node-v1
{"kind":"command_product","logical_key":"product:fixture-cli","payload":{"command_key":"fixture-cli","declaration_digest":"sha256:3333333333333333333333333333333333333333333333333333333333333333","entry_point_contract":"native_command","profile":"fixture-source-v1","skill_key":"fixture"}}
sha256:6c9a7ebc8ae8d0302bd3c14028dab48913bd9954ecddcc9fb67e50ecd859cea3

name=cgp05.extra
label=curator-node-v1
{"kind":"package_instance","logical_key":"pkg:extra@1.0.0","payload":{"artifact_manifest_id":"sha256:2222222222222222222222222222222222222222222222222222222222222222","ecosystem":"fixture","lock_instance_key":"extra","name":"extra","origin":"registry://fixture/extra/1.0.0","profile":"fixture-source-v1","trust_role":"dependency_input","version":"1.0.0"}}
sha256:48371dfff16352b496ef14ce4199a031b3b3d20489ca1f47f962a822d534bc50

name=cgp05.requires
label=curator-edge-v1
{"edge_key":"edge:fixture-cli-requires-extra","from_node_id":"sha256:6c9a7ebc8ae8d0302bd3c14028dab48913bd9954ecddcc9fb67e50ecd859cea3","kind":"requires","payload":{"condition":{"evaluator_id":"fixture-target-v1","expression":"target.os == linux"},"origin":{"field":"dependencies.extra[target.os=linux]","manifest_digest":"sha256:3333333333333333333333333333333333333333333333333333333333333333"},"scope":"optional"},"to_node_id":"sha256:48371dfff16352b496ef14ce4199a031b3b3d20489ca1f47f962a822d534bc50"}
sha256:7aa6d3f1f6ab5a9146e6fb4885684dcb26edc03a925478555dbbe45e5a685b02

name=cgp05.capture
label=curator-capture-graph-v1
{"artifact_manifest_ids":["sha256:2222222222222222222222222222222222222222222222222222222222222222"],"edge_ids":["sha256:7aa6d3f1f6ab5a9146e6fb4885684dcb26edc03a925478555dbbe45e5a685b02"],"node_ids":["sha256:48371dfff16352b496ef14ce4199a031b3b3d20489ca1f47f962a822d534bc50","sha256:6c9a7ebc8ae8d0302bd3c14028dab48913bd9954ecddcc9fb67e50ecd859cea3"],"policy_ids":["curator-artifact-policy-v1"],"profile_id":"fixture-source-v1","root_node_ids":["sha256:6c9a7ebc8ae8d0302bd3c14028dab48913bd9954ecddcc9fb67e50ecd859cea3"],"schema_id":"closure-capture-graph-v1"}
sha256:1bcd31f3b5b1e1e77da9256c4395d59f75802df7a6d3dcef2504448c2c04f5f2

name=cgp05.platform.darwin
label=curator-node-v1
{"kind":"target_platform","logical_key":"platform:darwin-arm64","payload":{"abi":"darwin","architecture":"arm64","libc":"libSystem","minimum_runtime":"macos-13.0","os":"darwin","sdk_id":"macos-sdk-fixture-v1","target_triple":"arm64-apple-macosx13.0"}}
sha256:730de9010560a6ad76e49fbe2d58b182ae56440f9e4117867eed3b42d700e06b

name=cgp05.platform.linux
label=curator-node-v1
{"kind":"target_platform","logical_key":"platform:linux-x86_64","payload":{"abi":"gnu","architecture":"x86_64","libc":"glibc","minimum_runtime":"glibc-2.31","os":"linux","sdk_id":"linux-sysroot-fixture-v1","target_triple":"x86_64-unknown-linux-gnu"}}
sha256:17527a5f8337dc55fb9390ac4671179d1dc14ec6433e5d2b6324314cd4fe0367

name=cgp05.selection.darwin
label=curator-selection-context-v1
{"default_features":false,"evaluator_ids":["fixture-target-v1"],"features":[],"markers":{},"peer_context":{},"platform_roles":{"target":"sha256:730de9010560a6ad76e49fbe2d58b182ae56440f9e4117867eed3b42d700e06b"},"product_node_ids":["sha256:6c9a7ebc8ae8d0302bd3c14028dab48913bd9954ecddcc9fb67e50ecd859cea3"],"schema_id":"closure-selection-context-v1"}
sha256:eb95dae28edbea38b77d2a0f7d702d0fa214e7126f714ef33ff857ff40b435e7

name=cgp05.selection.linux
label=curator-selection-context-v1
{"default_features":false,"evaluator_ids":["fixture-target-v1"],"features":[],"markers":{},"peer_context":{},"platform_roles":{"target":"sha256:17527a5f8337dc55fb9390ac4671179d1dc14ec6433e5d2b6324314cd4fe0367"},"product_node_ids":["sha256:6c9a7ebc8ae8d0302bd3c14028dab48913bd9954ecddcc9fb67e50ecd859cea3"],"schema_id":"closure-selection-context-v1"}
sha256:44523f093fb255a4b9015ac640faa9f16cc873817248754222c049c23f8849e2

name=cgp05.targets.darwin
label=curator-edge-v1
{"edge_key":"edge:fixture-cli-targets-darwin-arm64","from_node_id":"sha256:6c9a7ebc8ae8d0302bd3c14028dab48913bd9954ecddcc9fb67e50ecd859cea3","kind":"targets","payload":{"binding_role":"target","origin":{"field":"selection.platform_roles.target"}},"to_node_id":"sha256:730de9010560a6ad76e49fbe2d58b182ae56440f9e4117867eed3b42d700e06b"}
sha256:b3e04cda6e0419ff5d8281ab20dc1ab05986cbd5541a3b50f805c2a469e9578d

name=cgp05.targets.linux
label=curator-edge-v1
{"edge_key":"edge:fixture-cli-targets-linux-x86_64","from_node_id":"sha256:6c9a7ebc8ae8d0302bd3c14028dab48913bd9954ecddcc9fb67e50ecd859cea3","kind":"targets","payload":{"binding_role":"target","origin":{"field":"selection.platform_roles.target"}},"to_node_id":"sha256:17527a5f8337dc55fb9390ac4671179d1dc14ec6433e5d2b6324314cd4fe0367"}
sha256:6cb771139f53d164abcef33d460c427da94c74009989e9560960cc82ba88c430

name=cgp05.binding.darwin
label=curator-selection-binding-v1
{"binding_edge_ids":["sha256:b3e04cda6e0419ff5d8281ab20dc1ab05986cbd5541a3b50f805c2a469e9578d"],"binding_node_ids":["sha256:730de9010560a6ad76e49fbe2d58b182ae56440f9e4117867eed3b42d700e06b"],"captured_graph_id":"sha256:1bcd31f3b5b1e1e77da9256c4395d59f75802df7a6d3dcef2504448c2c04f5f2","schema_id":"closure-selection-binding-v1","selection_context_id":"sha256:eb95dae28edbea38b77d2a0f7d702d0fa214e7126f714ef33ff857ff40b435e7"}
sha256:5e2b1414ffeef8a4d8100c18e06a13ca8853515244025291db658096dcc2770f

name=cgp05.binding.linux
label=curator-selection-binding-v1
{"binding_edge_ids":["sha256:6cb771139f53d164abcef33d460c427da94c74009989e9560960cc82ba88c430"],"binding_node_ids":["sha256:17527a5f8337dc55fb9390ac4671179d1dc14ec6433e5d2b6324314cd4fe0367"],"captured_graph_id":"sha256:1bcd31f3b5b1e1e77da9256c4395d59f75802df7a6d3dcef2504448c2c04f5f2","schema_id":"closure-selection-binding-v1","selection_context_id":"sha256:44523f093fb255a4b9015ac640faa9f16cc873817248754222c049c23f8849e2"}
sha256:ae2c74d58e22e117c217c462c3eccc510d137b84dff4961a3aa28aba5a1ceb26

name=cgp05.active.darwin
label=curator-active-graph-v1
{"captured_graph_id":"sha256:1bcd31f3b5b1e1e77da9256c4395d59f75802df7a6d3dcef2504448c2c04f5f2","edge_activations":[{"edge_id":"sha256:7aa6d3f1f6ab5a9146e6fb4885684dcb26edc03a925478555dbbe45e5a685b02","evaluation":false,"reason":"condition_false","state":"pruned"}],"node_activations":[{"node_id":"sha256:48371dfff16352b496ef14ce4199a031b3b3d20489ca1f47f962a822d534bc50","state":"pruned"},{"node_id":"sha256:6c9a7ebc8ae8d0302bd3c14028dab48913bd9954ecddcc9fb67e50ecd859cea3","state":"selected"}],"non_ordering_sccs":[],"schema_id":"closure-active-graph-v1","selection_binding_id":"sha256:5e2b1414ffeef8a4d8100c18e06a13ca8853515244025291db658096dcc2770f","selection_context_id":"sha256:eb95dae28edbea38b77d2a0f7d702d0fa214e7126f714ef33ff857ff40b435e7"}
sha256:c8a8a70de7cece61ad01eabebd6171cd9af62ae11b783761cd4562cb4fb9e3e8

name=cgp05.active.linux
label=curator-active-graph-v1
{"captured_graph_id":"sha256:1bcd31f3b5b1e1e77da9256c4395d59f75802df7a6d3dcef2504448c2c04f5f2","edge_activations":[{"edge_id":"sha256:7aa6d3f1f6ab5a9146e6fb4885684dcb26edc03a925478555dbbe45e5a685b02","evaluation":true,"reason":"condition_true","state":"selected"}],"node_activations":[{"node_id":"sha256:48371dfff16352b496ef14ce4199a031b3b3d20489ca1f47f962a822d534bc50","state":"selected"},{"node_id":"sha256:6c9a7ebc8ae8d0302bd3c14028dab48913bd9954ecddcc9fb67e50ecd859cea3","state":"selected"}],"non_ordering_sccs":[],"schema_id":"closure-active-graph-v1","selection_binding_id":"sha256:ae2c74d58e22e117c217c462c3eccc510d137b84dff4961a3aa28aba5a1ceb26","selection_context_id":"sha256:44523f093fb255a4b9015ac640faa9f16cc873817248754222c049c23f8849e2"}
sha256:ad0b68d8939eb9e1ab74f8fec4b164e418096dad244ece858e02e7d4b34beabc

name=cgp05.plan.darwin
label=curator-build-plan-v1
{"action_node_ids":[],"active_graph_id":"sha256:c8a8a70de7cece61ad01eabebd6171cd9af62ae11b783761cd4562cb4fb9e3e8","declared_output_node_ids":[],"execution_policy_id":"fixture-execution-v1","ordering_edges":[],"schema_id":"closure-build-plan-v1","waves":[]}
sha256:e71959451bf642d7860caaac2c14d0ea8c18bcc27d5d743255f5afc94de5a139

name=cgp05.plan.linux
label=curator-build-plan-v1
{"action_node_ids":[],"active_graph_id":"sha256:ad0b68d8939eb9e1ab74f8fec4b164e418096dad244ece858e02e7d4b34beabc","declared_output_node_ids":[],"execution_policy_id":"fixture-execution-v1","ordering_edges":[],"schema_id":"closure-build-plan-v1","waves":[]}
sha256:3f568ae879be512fdf808f11ce9ed534337ff597c49be265dab972fa406a3b75

name=cgp05.c3-anchor
label=curator-checkpoint-fixture-anchor-v1
{"checkpoint_name":"C3.admit","fixture_id":"CGP05","schema_id":"closure-checkpoint-fixture-anchor-v1"}
sha256:73222e0fec267f7869f03fd0fa5bc94d46dff87ab2e7b6ce914894b13298782f

name=cgp05.c4.darwin
label=curator-checkpoint-v1
{"checkpoint_name":"C4.close","decision":"admit","diagnostics":[],"payload":{"active_graph_id":"sha256:c8a8a70de7cece61ad01eabebd6171cd9af62ae11b783761cd4562cb4fb9e3e8","captured_graph_id":"sha256:1bcd31f3b5b1e1e77da9256c4395d59f75802df7a6d3dcef2504448c2c04f5f2","selection_binding_id":"sha256:5e2b1414ffeef8a4d8100c18e06a13ca8853515244025291db658096dcc2770f","selection_context_id":"sha256:eb95dae28edbea38b77d2a0f7d702d0fa214e7126f714ef33ff857ff40b435e7"},"previous_checkpoint_id":"sha256:73222e0fec267f7869f03fd0fa5bc94d46dff87ab2e7b6ce914894b13298782f","schema_id":"closure-checkpoint-v1"}
sha256:af7645b4942ff42586144fdf69455166a13ad7dcab5e2ea7d08b083b0b8dd2cf

name=cgp05.c4.linux
label=curator-checkpoint-v1
{"checkpoint_name":"C4.close","decision":"admit","diagnostics":[],"payload":{"active_graph_id":"sha256:ad0b68d8939eb9e1ab74f8fec4b164e418096dad244ece858e02e7d4b34beabc","captured_graph_id":"sha256:1bcd31f3b5b1e1e77da9256c4395d59f75802df7a6d3dcef2504448c2c04f5f2","selection_binding_id":"sha256:ae2c74d58e22e117c217c462c3eccc510d137b84dff4961a3aa28aba5a1ceb26","selection_context_id":"sha256:44523f093fb255a4b9015ac640faa9f16cc873817248754222c049c23f8849e2"},"previous_checkpoint_id":"sha256:73222e0fec267f7869f03fd0fa5bc94d46dff87ab2e7b6ce914894b13298782f","schema_id":"closure-checkpoint-v1"}
sha256:f35a82ebc0b023aa466c629060545f3b4ea64ea2bb7d1eb0775a56337404370a

name=cgp05.c5.darwin
label=curator-checkpoint-v1
{"checkpoint_name":"C5.plan","decision":"admit","diagnostics":[],"payload":{"build_plan_id":"sha256:e71959451bf642d7860caaac2c14d0ea8c18bcc27d5d743255f5afc94de5a139"},"previous_checkpoint_id":"sha256:af7645b4942ff42586144fdf69455166a13ad7dcab5e2ea7d08b083b0b8dd2cf","schema_id":"closure-checkpoint-v1"}
sha256:74b020032b7466dafdf5ed33e35a57008884fe4086f53990713ba4d21d506b14

name=cgp05.c5.linux
label=curator-checkpoint-v1
{"checkpoint_name":"C5.plan","decision":"admit","diagnostics":[],"payload":{"build_plan_id":"sha256:3f568ae879be512fdf808f11ce9ed534337ff597c49be265dab972fa406a3b75"},"previous_checkpoint_id":"sha256:f35a82ebc0b023aa466c629060545f3b4ea64ea2bb7d1eb0775a56337404370a","schema_id":"closure-checkpoint-v1"}
sha256:10ee2b87c903e1b74c95568fa343c03e1ac638a93b9919a1b90dacd062931e97
```

The assertions are exact: Darwin and Linux both use capture ID
`sha256:1bcd31f3...04f5f2`; their concrete platform, selection, binding, active,
plan, C4, and C5 IDs all differ; each `SelectionContext.platform_roles.target`
resolves to its published platform node; and each binding contains the explicit
published `targets` edge. Adding requested target facts or either target edge to
capture, omitting the overlay, or accepting a dangling/wrong-kind platform
reference breaks this golden.

### Immutable expected-output and receipt golden

`CGP10` publishes every referenced codec fixture record, its domain label,
exact CCJ bytes, and derived ID. The admitted source seed is the 26 UTF-8 bytes
`int main(void){return 0;}` plus LF. The C3 fixture anchor is the same kind of
published conformance-only predecessor sentinel used by `CGP05`; all graph,
plan, C4, C5, closure, observation, execution, and publication records use the
normative schemas. The `one` and `two` branches differ only after C5.

```text
name=cgp10.artifact-manifest
label=curator-artifact-manifest-v1
{"decision":"ADMIT_INPUT","fixture_id":"CGP10","leaf":{"path":"main.c","sha256":"sha256:86004d65c4f387c95467c6cee92bc1f1f8cb04d6650be09fbd1e359834a56766","size":26},"policy_id":"curator-artifact-policy-v1","schema_id":"artifact-manifest-v1"}
sha256:8b74a3c90112a14569e008ecfc0dc3b606af11729c94d6f158df9da858cbd594

name=cgp10.product
label=curator-node-v1
{"kind":"command_product","logical_key":"product:fixture-cli","payload":{"command_key":"fixture-cli","declaration_digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","entry_point_contract":"native_command","profile":"fixture-source-v1","skill_key":"fixture"}}
sha256:cd6f8b28322d936302424dfd2e63ea6de415da146f11f864059f01de2736628d

name=cgp10.source
label=curator-node-v1
{"kind":"source_set","logical_key":"source:main.c","payload":{"artifact_manifest_id":"sha256:8b74a3c90112a14569e008ecfc0dc3b606af11729c94d6f158df9da858cbd594","grammar":"c-source-v1","origin":"fixture://CGP10/main.c","profile":"fixture-source-v1","projection":["main.c"],"trust_role":"dependency_input"}}
sha256:a9ec53c2f23441d605fde246e9b9591ddc2feba1748c82a82fd97f67765820ba

name=cgp10.action
label=curator-node-v1
{"kind":"action","logical_key":"action:compile","payload":{"action_subtype":"compiler","argv_template":["$TOOL(compiler)","$READ(src)","-o","$WRITE(cli)"],"environment_policy_id":"env-v1","execution_domain":"target","network":"none","process_policy_id":"process-v1","profile":"fixture-source-v1","read_slot_names":["src"],"tool_slot_names":["compiler"],"write_slot_names":["cli"]}}
sha256:cee4a0fa99322dc92049228dd12e4df7d45a2ca13609eb1710fbcfc02c34ce7e

name=cgp10.output
label=curator-node-v1
{"kind":"output_artifact","logical_key":"output:cli","payload":{"expected_class":"native.executable","logical_path":"bin/cli","output_role":"published_command","profile":"fixture-source-v1"}}
sha256:a0502b262aa5a9be018eb43910c58d7d6d6634b87301526727ef31a0ba180691

name=cgp10.declares
label=curator-edge-v1
{"edge_key":"edge:fixture-cli-declares-compile","from_node_id":"sha256:cd6f8b28322d936302424dfd2e63ea6de415da146f11f864059f01de2736628d","kind":"declares","payload":{"origin":{"field":"fixture.actions.compile"}},"to_node_id":"sha256:cee4a0fa99322dc92049228dd12e4df7d45a2ca13609eb1710fbcfc02c34ce7e"}
sha256:dea20bbad611eee3dea791db039e3dc419c4124743d38c8f649a58221b4b851c

name=cgp10.reads
label=curator-edge-v1
{"edge_key":"edge:compile-reads-main","from_node_id":"sha256:cee4a0fa99322dc92049228dd12e4df7d45a2ca13609eb1710fbcfc02c34ce7e","kind":"reads","payload":{"path":"main.c","read_slot":"src"},"to_node_id":"sha256:a9ec53c2f23441d605fde246e9b9591ddc2feba1748c82a82fd97f67765820ba"}
sha256:bce7373d51fd183db8a6322930792e785e92170c68a0faf0c20d4ccfd64a3fb9

name=cgp10.produces
label=curator-edge-v1
{"edge_key":"edge:compile-produces-cli","from_node_id":"sha256:cee4a0fa99322dc92049228dd12e4df7d45a2ca13609eb1710fbcfc02c34ce7e","kind":"produces","payload":{"path":"bin/cli","write_slot":"cli"},"to_node_id":"sha256:a0502b262aa5a9be018eb43910c58d7d6d6634b87301526727ef31a0ba180691"}
sha256:30704c3d7a0d0f937e27de4d3996411e1d88b3261f126150ab287230ac0a45b3

name=cgp10.publishes
label=curator-edge-v1
{"edge_key":"edge:fixture-cli-publishes-cli","from_node_id":"sha256:cd6f8b28322d936302424dfd2e63ea6de415da146f11f864059f01de2736628d","kind":"publishes","payload":{"destination":"bin/cli","entry_point":"fixture-cli"},"to_node_id":"sha256:a0502b262aa5a9be018eb43910c58d7d6d6634b87301526727ef31a0ba180691"}
sha256:0cccf7078514286313ebf8b7cf727cf48697f8362fd990352695e05a79ce855e

name=cgp10.capture
label=curator-capture-graph-v1
{"artifact_manifest_ids":["sha256:8b74a3c90112a14569e008ecfc0dc3b606af11729c94d6f158df9da858cbd594"],"edge_ids":["sha256:0cccf7078514286313ebf8b7cf727cf48697f8362fd990352695e05a79ce855e","sha256:30704c3d7a0d0f937e27de4d3996411e1d88b3261f126150ab287230ac0a45b3","sha256:bce7373d51fd183db8a6322930792e785e92170c68a0faf0c20d4ccfd64a3fb9","sha256:dea20bbad611eee3dea791db039e3dc419c4124743d38c8f649a58221b4b851c"],"node_ids":["sha256:a0502b262aa5a9be018eb43910c58d7d6d6634b87301526727ef31a0ba180691","sha256:a9ec53c2f23441d605fde246e9b9591ddc2feba1748c82a82fd97f67765820ba","sha256:cd6f8b28322d936302424dfd2e63ea6de415da146f11f864059f01de2736628d","sha256:cee4a0fa99322dc92049228dd12e4df7d45a2ca13609eb1710fbcfc02c34ce7e"],"policy_ids":["curator-artifact-policy-v1"],"profile_id":"fixture-source-v1","root_node_ids":["sha256:cd6f8b28322d936302424dfd2e63ea6de415da146f11f864059f01de2736628d"],"schema_id":"closure-capture-graph-v1"}
sha256:5e6c0363867ef33fc6094939818551f91f0c6e74adc00ee6a7235b455757113b

name=cgp10.platform
label=curator-node-v1
{"kind":"target_platform","logical_key":"platform:darwin-arm64","payload":{"abi":"darwin","architecture":"arm64","libc":"libSystem","minimum_runtime":"macos-13.0","os":"darwin","sdk_id":"macos-sdk-fixture-v1","target_triple":"arm64-apple-macosx13.0"}}
sha256:730de9010560a6ad76e49fbe2d58b182ae56440f9e4117867eed3b42d700e06b

name=cgp10.toolchain
label=curator-node-v1
{"kind":"toolchain_component","logical_key":"toolchain:fixture-cc","payload":{"component_role":"c_compiler","content_fingerprint":"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","executable_relative_path":"bin/fixture-cc","platform_abi":"darwin-arm64","policy_selector":"fixture-toolchain-v1","version_output":"fixture-cc 1.0"}}
sha256:17adb5a8a6fb4adb2f069caecc4d9a3a2b1126226c031f9a9e6d52e2f3fb73f6

name=cgp10.selection
label=curator-selection-context-v1
{"default_features":false,"evaluator_ids":[],"features":[],"markers":{},"peer_context":{},"platform_roles":{"target":"sha256:730de9010560a6ad76e49fbe2d58b182ae56440f9e4117867eed3b42d700e06b"},"product_node_ids":["sha256:cd6f8b28322d936302424dfd2e63ea6de415da146f11f864059f01de2736628d"],"schema_id":"closure-selection-context-v1"}
sha256:901c675cd9a9297cbdceac1436f4307a34da3f134df6185c6830897be8cfe016

name=cgp10.uses-tool
label=curator-edge-v1
{"edge_key":"edge:compile-uses-fixture-cc","from_node_id":"sha256:cee4a0fa99322dc92049228dd12e4df7d45a2ca13609eb1710fbcfc02c34ce7e","kind":"uses_tool","payload":{"executable_relative_path":"bin/fixture-cc","tool_slot":"compiler"},"to_node_id":"sha256:17adb5a8a6fb4adb2f069caecc4d9a3a2b1126226c031f9a9e6d52e2f3fb73f6"}
sha256:a219bbf8650a047b8308413e60320f4d40cbd67ade647082c91e4548ff12b908

name=cgp10.targets.product
label=curator-edge-v1
{"edge_key":"edge:product-targets-darwin-arm64","from_node_id":"sha256:cd6f8b28322d936302424dfd2e63ea6de415da146f11f864059f01de2736628d","kind":"targets","payload":{"binding_role":"target","origin":{"field":"selection.platform_roles.target"}},"to_node_id":"sha256:730de9010560a6ad76e49fbe2d58b182ae56440f9e4117867eed3b42d700e06b"}
sha256:c58c26a2729baefac8d67ce0a7a67ec611ff9de3524c645206b38cbb70e3951c

name=cgp10.targets.action
label=curator-edge-v1
{"edge_key":"edge:action-targets-darwin-arm64","from_node_id":"sha256:cee4a0fa99322dc92049228dd12e4df7d45a2ca13609eb1710fbcfc02c34ce7e","kind":"targets","payload":{"binding_role":"target","origin":{"field":"selection.platform_roles.target"}},"to_node_id":"sha256:730de9010560a6ad76e49fbe2d58b182ae56440f9e4117867eed3b42d700e06b"}
sha256:6970abafa15cc6f3beb4b2cdea26c3bdf4881726c4c4cdb5fb8c87f319f5fe21

name=cgp10.targets.toolchain
label=curator-edge-v1
{"edge_key":"edge:toolchain-targets-darwin-arm64","from_node_id":"sha256:17adb5a8a6fb4adb2f069caecc4d9a3a2b1126226c031f9a9e6d52e2f3fb73f6","kind":"targets","payload":{"binding_role":"target","origin":{"field":"selection.platform_roles.target"}},"to_node_id":"sha256:730de9010560a6ad76e49fbe2d58b182ae56440f9e4117867eed3b42d700e06b"}
sha256:d44dd6fe4610aecda96ccd101c8a771ba89d2764ccb5759105a84fdf5328a1c2

name=cgp10.targets.output
label=curator-edge-v1
{"edge_key":"edge:output-targets-darwin-arm64","from_node_id":"sha256:a0502b262aa5a9be018eb43910c58d7d6d6634b87301526727ef31a0ba180691","kind":"targets","payload":{"binding_role":"target","origin":{"field":"selection.platform_roles.target"}},"to_node_id":"sha256:730de9010560a6ad76e49fbe2d58b182ae56440f9e4117867eed3b42d700e06b"}
sha256:edfdc5a3c8687bba58f63dd28131c970465bd728249e7bd86caa6195c277f9b2

name=cgp10.binding
label=curator-selection-binding-v1
{"binding_edge_ids":["sha256:6970abafa15cc6f3beb4b2cdea26c3bdf4881726c4c4cdb5fb8c87f319f5fe21","sha256:a219bbf8650a047b8308413e60320f4d40cbd67ade647082c91e4548ff12b908","sha256:c58c26a2729baefac8d67ce0a7a67ec611ff9de3524c645206b38cbb70e3951c","sha256:d44dd6fe4610aecda96ccd101c8a771ba89d2764ccb5759105a84fdf5328a1c2","sha256:edfdc5a3c8687bba58f63dd28131c970465bd728249e7bd86caa6195c277f9b2"],"binding_node_ids":["sha256:17adb5a8a6fb4adb2f069caecc4d9a3a2b1126226c031f9a9e6d52e2f3fb73f6","sha256:730de9010560a6ad76e49fbe2d58b182ae56440f9e4117867eed3b42d700e06b"],"captured_graph_id":"sha256:5e6c0363867ef33fc6094939818551f91f0c6e74adc00ee6a7235b455757113b","schema_id":"closure-selection-binding-v1","selection_context_id":"sha256:901c675cd9a9297cbdceac1436f4307a34da3f134df6185c6830897be8cfe016"}
sha256:7f404718cb92e903b650594515e373cdaf7643b4908225cc7671cf262c2c1578

name=cgp10.active
label=curator-active-graph-v1
{"captured_graph_id":"sha256:5e6c0363867ef33fc6094939818551f91f0c6e74adc00ee6a7235b455757113b","edge_activations":[],"node_activations":[{"node_id":"sha256:a0502b262aa5a9be018eb43910c58d7d6d6634b87301526727ef31a0ba180691","state":"selected"},{"node_id":"sha256:a9ec53c2f23441d605fde246e9b9591ddc2feba1748c82a82fd97f67765820ba","state":"selected"},{"node_id":"sha256:cd6f8b28322d936302424dfd2e63ea6de415da146f11f864059f01de2736628d","state":"selected"},{"node_id":"sha256:cee4a0fa99322dc92049228dd12e4df7d45a2ca13609eb1710fbcfc02c34ce7e","state":"selected"}],"non_ordering_sccs":[],"schema_id":"closure-active-graph-v1","selection_binding_id":"sha256:7f404718cb92e903b650594515e373cdaf7643b4908225cc7671cf262c2c1578","selection_context_id":"sha256:901c675cd9a9297cbdceac1436f4307a34da3f134df6185c6830897be8cfe016"}
sha256:5491c83f7169f4cdd3416382f89205c9e960a13dbaa05ceffaeb428beca25ef0

name=cgp10.plan
label=curator-build-plan-v1
{"action_node_ids":["sha256:cee4a0fa99322dc92049228dd12e4df7d45a2ca13609eb1710fbcfc02c34ce7e"],"active_graph_id":"sha256:5491c83f7169f4cdd3416382f89205c9e960a13dbaa05ceffaeb428beca25ef0","declared_output_node_ids":["sha256:a0502b262aa5a9be018eb43910c58d7d6d6634b87301526727ef31a0ba180691"],"execution_policy_id":"fixture-execution-v1","ordering_edges":[],"schema_id":"closure-build-plan-v1","waves":[["sha256:cee4a0fa99322dc92049228dd12e4df7d45a2ca13609eb1710fbcfc02c34ce7e"]]}
sha256:2ecd72056ccf4a3dce3ef8d7b9288f2a895f4eadf6841f35700c7c302fe02ed2

name=cgp10.c3-anchor
label=curator-checkpoint-fixture-anchor-v1
{"checkpoint_name":"C3.admit","fixture_id":"CGP10","schema_id":"closure-checkpoint-fixture-anchor-v1"}
sha256:1150e262159ee0a4025a35c823bbe5bf256d8fa31e1caab791c25036926e2267

name=cgp10.c4
label=curator-checkpoint-v1
{"checkpoint_name":"C4.close","decision":"admit","diagnostics":[],"payload":{"active_graph_id":"sha256:5491c83f7169f4cdd3416382f89205c9e960a13dbaa05ceffaeb428beca25ef0","captured_graph_id":"sha256:5e6c0363867ef33fc6094939818551f91f0c6e74adc00ee6a7235b455757113b","selection_binding_id":"sha256:7f404718cb92e903b650594515e373cdaf7643b4908225cc7671cf262c2c1578","selection_context_id":"sha256:901c675cd9a9297cbdceac1436f4307a34da3f134df6185c6830897be8cfe016"},"previous_checkpoint_id":"sha256:1150e262159ee0a4025a35c823bbe5bf256d8fa31e1caab791c25036926e2267","schema_id":"closure-checkpoint-v1"}
sha256:e8024adb0cad0a5ffc815e97294fd240d7cd59a47e4ebd5da0d736a5bd989120

name=cgp10.c5
label=curator-checkpoint-v1
{"checkpoint_name":"C5.plan","decision":"admit","diagnostics":[],"payload":{"build_plan_id":"sha256:2ecd72056ccf4a3dce3ef8d7b9288f2a895f4eadf6841f35700c7c302fe02ed2"},"previous_checkpoint_id":"sha256:e8024adb0cad0a5ffc815e97294fd240d7cd59a47e4ebd5da0d736a5bd989120","schema_id":"closure-checkpoint-v1"}
sha256:8730ad8567f6874837242937fc00f76f8fddbd77d7e6cb3096625f4abcf1a3a6

name=cgp10.closure
label=curator-source-closure-v1
{"c5_checkpoint_id":"sha256:8730ad8567f6874837242937fc00f76f8fddbd77d7e6cb3096625f4abcf1a3a6","schema_id":"curator-source-closure-v1"}
sha256:c66440c54e898549b510fb6e6415c8918cb44899a92ce06a98d671f6928f1c9d

name=cgp10.expected-cache-input
label=curator-expected-cache-input-v1
{"closure_id":"sha256:c66440c54e898549b510fb6e6415c8918cb44899a92ce06a98d671f6928f1c9d","expected_output_node_ids":["sha256:a0502b262aa5a9be018eb43910c58d7d6d6634b87301526727ef31a0ba180691"],"schema_id":"closure-expected-cache-input-v1"}
sha256:8bd4c2b7178aef7426f453169cb6351824abf3e5dae33edbdd363dc790d260a1

name=cgp10.observation.one
label=curator-produced-artifact-observation-v1
{"class":"native.executable","expected_output_node_id":"sha256:a0502b262aa5a9be018eb43910c58d7d6d6634b87301526727ef31a0ba180691","path":"bin/cli","producer_action_id":"sha256:cee4a0fa99322dc92049228dd12e4df7d45a2ca13609eb1710fbcfc02c34ce7e","produces_edge_id":"sha256:30704c3d7a0d0f937e27de4d3996411e1d88b3261f126150ab287230ac0a45b3","sha256":"sha256:7692c3ad3540bb803c020b3aee66cd8887123234ea0c6e7143c0add73ff431ed","size":3}
sha256:5c7837de3e32a78c9c51c6a199d963ae2d3d7fb46ebfb24c062ae7f67bd065e9

name=cgp10.execution.one
label=curator-execution-receipt-v1
{"action_order":["sha256:cee4a0fa99322dc92049228dd12e4df7d45a2ca13609eb1710fbcfc02c34ce7e"],"closure_id":"sha256:c66440c54e898549b510fb6e6415c8918cb44899a92ce06a98d671f6928f1c9d","decision":"success","network":"none","produced_observation_ids":["sha256:5c7837de3e32a78c9c51c6a199d963ae2d3d7fb46ebfb24c062ae7f67bd065e9"],"schema_id":"closure-execution-receipt-v1","toolchain_rechecks":"match","write_set":["bin/cli"]}
sha256:01d8aeba69eda60b9b66e3de0d360c818ca76fe1b1cb54c2af52d0d6dbf3827e

name=cgp10.publication.one
label=curator-publication-receipt-v1
{"decision":"published","execution_receipt_id":"sha256:01d8aeba69eda60b9b66e3de0d360c818ca76fe1b1cb54c2af52d0d6dbf3827e","expected_cache_input_id":"sha256:8bd4c2b7178aef7426f453169cb6351824abf3e5dae33edbdd363dc790d260a1","protected_result":"exact_write","published_observation_ids":["sha256:5c7837de3e32a78c9c51c6a199d963ae2d3d7fb46ebfb24c062ae7f67bd065e9"],"schema_id":"closure-publication-receipt-v1"}
sha256:be40450ce12e9d10fa27a040040d79a55717ab58f7b8bf357f9fb8be76dfcd08

name=cgp10.observation.two
label=curator-produced-artifact-observation-v1
{"class":"native.executable","expected_output_node_id":"sha256:a0502b262aa5a9be018eb43910c58d7d6d6634b87301526727ef31a0ba180691","path":"bin/cli","producer_action_id":"sha256:cee4a0fa99322dc92049228dd12e4df7d45a2ca13609eb1710fbcfc02c34ce7e","produces_edge_id":"sha256:30704c3d7a0d0f937e27de4d3996411e1d88b3261f126150ab287230ac0a45b3","sha256":"sha256:3fc4ccfe745870e2c0d99f71f30ff0656c8dedd41cc1d7d3d376b0dbe685e2f3","size":3}
sha256:4ce66fd6765802f5692cc81b7229fff3d6ae5a85442da0af37192a0b15ce057a

name=cgp10.execution.two
label=curator-execution-receipt-v1
{"action_order":["sha256:cee4a0fa99322dc92049228dd12e4df7d45a2ca13609eb1710fbcfc02c34ce7e"],"closure_id":"sha256:c66440c54e898549b510fb6e6415c8918cb44899a92ce06a98d671f6928f1c9d","decision":"success","network":"none","produced_observation_ids":["sha256:4ce66fd6765802f5692cc81b7229fff3d6ae5a85442da0af37192a0b15ce057a"],"schema_id":"closure-execution-receipt-v1","toolchain_rechecks":"match","write_set":["bin/cli"]}
sha256:a19b8ee241590953b3ab80d67c837e06807b3e112cccaaf0724a1332942e79d9

name=cgp10.publication.two
label=curator-publication-receipt-v1
{"decision":"published","execution_receipt_id":"sha256:a19b8ee241590953b3ab80d67c837e06807b3e112cccaaf0724a1332942e79d9","expected_cache_input_id":"sha256:8bd4c2b7178aef7426f453169cb6351824abf3e5dae33edbdd363dc790d260a1","protected_result":"exact_write","published_observation_ids":["sha256:4ce66fd6765802f5692cc81b7229fff3d6ae5a85442da0af37192a0b15ce057a"],"schema_id":"closure-publication-receipt-v1"}
sha256:39f8595568f4d5ecad1d46b07ea5f0319b8a001c6029158f90afd28aaa8bc60d
```

Both branches retain identical intrinsic action/output nodes, capture and
binding edges, capture/binding/active graph, build plan, C4, C5, closure, and
independently derived expected-cache-input IDs. Only the produced observation
and its downstream execution/publication receipt IDs vary. Whether the `two`
branch is publishable under a stronger ecosystem reproducibility profile is a
separate policy decision; observed bytes can never rewrite the planned graph.

## Audit checkpoints

Every `Cn` includes `schema_id`, `checkpoint_name`, `previous_checkpoint_id`,
canonical payload, decision, deterministic diagnostics, and its own ID. A
failure record names the last successful checkpoint; a failed checkpoint is
not issued and cannot seed a cache entry.

| Checkpoint | Required payload and gate |
| --- | --- |
| `C0.profile` | Adapter/profile and schema versions; artifact policy, detector registry, source grammar and limit vector; requested selection context, exact intrinsic platform records, and platform roles; supported lock/manager/tool versions; configuration/environment allowlists; extension capabilities; and a complete external-toolchain checkpoint for every executable that may derive evidence before C5, including Swift/PackageDescription, Cargo, Node/manager, and mirror/metadata tools. Reject unknown formats, policy versions, targets, duplicate/dangling/wrong-kind platform records, manager extensions, absent capabilities, or an incomplete toolchain fingerprint. No process may run before C0 is committed. |
| `C1.resolve` | Frozen root/workspace declaration bytes and authoritative lock candidate; capture-edge condition declarations; parser/evaluator identities; candidate packages/products/targets/extensions/system edges; requested selection context; and the complete ordered intake/derivation journal used to obtain the candidate. Declarative parsing needs no process. For executable Swift manifest evaluation, the complete root tree—and then each complete fetched package tree—has an admitted intake receipt before that package's manifest permit. Iteration continues in canonical dependency rounds until a fixed point. Reject stale/missing lock, mutable locator, unknown kind, unresolved context, unauthorized derivation, or metadata conflict. |
| `C2.capture` | Aggregate every per-input intake receipt used by C1 plus all remaining lock-superset captures: immutable origin, expected and observed size/digest, raw bytes or canonical source snapshot in protected storage, retrieval record, workspace containment, source-control object/submodule evidence, and broker receipt. Network is allowed only to a separately permitted manager acquisition broker for exact origins; it cannot evaluate package code. Every C1 intake ID must appear unchanged. Reject missing/mismatched identity, incomplete capture, or a mutable handle. |
| `C3.admit` | Aggregate `artifact-manifest-v1` for every C2 root/path snapshot, raw package container, nested member, and relevant derived source materialization; preliminary C1 intake admissions must reproduce exactly. For Cargo, `C3a.origin-admission` admits every origin, then one derivation permit authorizes the selected Cargo/toolchain to run `cargo vendor`; its receipt binds command and observed I/O; `C3b.derived-admission` verifies the exact transform and rescans vendor output. No Cargo metadata/build runs before final C3. Any deny/unknown/incomplete or permit/receipt mismatch stops all later execution. |
| `C4.close` | Reconciled selection-neutral `G_capture`, explicit `SelectionContext(S)`, exact `SelectionBinding(S)`, and their `G_active(S)` union; artifact-manifest refs; package metadata agreement; unique origin/package/target/source mappings; explicit generated/extension/interop/system declarations; concrete platform/toolchain/SDK binding nodes and typed edges; selected/pruned capture activations; SCC evidence; and all four graph IDs. SwiftPM kind-preserving mirror replay and Cargo/Node manager metadata may run only through derivation permits whose admitted inputs and toolchains chain from final C3. Reject any missing/extra/duplicate/dangling/wrong-kind capture or binding node/edge, platform-role mismatch, active unsupported extension, metadata or derivation drift, ambiguous mapping, or undeclared FFI/system input. |
| `C5.plan` | Acyclic `D_build(S)`, stable waves/order, exact action commands resolved from immutable action slots and C4 capture/binding edges, physical executable resolution, working directories, host/target roles, tools/runtimes/SDKs, read/write/process/network/environment policies, generated lineage, expected immutable output nodes, runtime/bundle edges, fresh-root layout, time-of-use recheck plan, and build-plan ID. C5 may reference C0-prebound evidence tools and C4-bound build tools, but may add no node or edge and cannot mutate capture/binding or retroactively authorize pre-C5 use. After C5 is issued, its checkpoint ID derives `closure_id`. Reject ordering cycles, implicit or unbound actions, unsupported hook/plugin/macro/build/native units, untrusted toolchains, or unsafe target. |
| `C6.offline` | New sandbox identity; read-only closure mount; empty output/temp/home/config roots; private derived manager cache/store receipt; ambient cache/credential denial; OS network=`none`; frozen/immutable manager flags; actual build-DAG action order; process/read/environment/write/network audit; materialized tree/package map; toolchain/target rechecks; separate produced-artifact observations keyed to immutable expected output nodes; and execution receipt. Missing capture, graph drift, network/process/I/O violation, unexpected action/output, or toolchain/target drift stops publication and never mutates C4/C5 identities. |
| `C7.publish` | Recursive output inspection; exact observation/declared write-set reconciliation; runtime entry-point resolution; graph/plan/checkpoint/toolchain/target refs; sorted produced-observation records; canonical receipt; protected-cache ownership/atomic publication result; and independent exact-hit validation inputs. Pre-existing, unreceipted, drifted, or denied-class output cannot publish. Only C7 creates reusable state. |

### Evidence-derivation execution before C5

Pre-C5 execution is a closed evidence plane, not an unmodeled exception to the
build DAG. Its three canonical record types are:

| Record | Required fields and rule |
| --- | --- |
| `intake_admission_receipt` | Previous causal head; immutable origin and protected handle; raw/tree size and digest; artifact policy/profile/detector/limit IDs; complete `artifact-manifest-v1` ID and admit decision. It is produced without executing package code. |
| `derivation_permit` | Previous causal head; invocation key and subtype; exact admitted input receipt IDs; C0 external-toolchain checkpoint and executable component/relative path; host/target; argv, logical cwd, environment; process/read/write/network policies; empty or declared evidence-output roots; expected evidence schema; resource limits; and mandatory immediate time-of-use recheck. The permit is committed before launch. |
| `derivation_receipt` | Permit ID; before/after toolchain fingerprints; actual executable/argv/cwd/environment; process/read/write/network audit; exit status; deterministic diagnostics; output paths/manifests/digests; and decision. A successful receipt becomes the next causal head; a failed invocation cannot contribute evidence to a checkpoint. |

The journal is canonical by `(resolution_round, package_origin_key,
invocation_subtype)`. The initial profile executes derivations serially in that
order; a future parallel form must preserve the same explicit causal DAG and
canonical aggregate. A process may read only toolchain bytes and input receipts
already present in its permit. Newly discovered bytes return to the broker,
capture, and admission path before any process can interpret them.

SwiftPM therefore follows this exact loop:

1. freeze and admit the complete root tree;
2. issue and execute the root-manifest permit;
3. broker each newly discovered package into quarantine;
4. freeze and admit that package's complete tree before issuing its manifest
   permit; and
5. repeat until the lock/manifest graph reaches a fixed point, then aggregate
   the same intake receipts at C2/C3 and replay from admitted mirrors at C4.

A rejected root has zero root-manifest evaluations. A rejected dependency can
be discovered by an already admitted root, but has zero evaluations of its own
manifest. Cargo vendoring occurs only between C3a and C3b; Swift mirror replay
and Cargo/Node metadata occur only inside C4. Each uses a new permit and
receipt. If the selected toolchain changes after permit issuance, the immediate
recheck fails before the affected process starts. C5 is authoritative only for
the derived materialize/generate/compile/link/package/publish DAG; it never
retroactively authorizes these earlier evidence actions.

### First binding of trusted inputs

| Input class | First authoritative binding | Downstream rule |
| --- | --- | --- |
| Policy/profile/schema/limits/detectors and requested selection | C0 | Every later checkpoint or subreceipt carries the causal chain; changes invalidate admission and receipts. |
| Evidence-derivation toolchain/SDK/runtime | C0 external-toolchain checkpoint, then each pre-use permit | Immediate time-of-use recheck precedes manifest/vendor/mirror/metadata launch; C5 cannot backfill trust. |
| Build-only external toolchain/SDK/runtime | C4 manager selector, complete fingerprint, platform node, and typed binding edge | C5 resolves action slots only through these unchanged bindings; C6 immediately rechecks identity before use. |
| Root/dependency bytes, source snapshots, Git objects, origins | Per-input `intake_admission_receipt` before first interpretation; C2 is the complete aggregate | Later aggregation must contain the exact earlier receipt; no mutable manager cache can replace its protected handle. |
| Artifact classes/member manifests and admission verdicts | Same per-input receipt before executable interpretation; C3 is the complete aggregate/derived gate | C4 graph nodes reference exact manifest IDs; preliminary and aggregate results must be identical; no adapter override. |
| Manifest/lock declarations and candidate conditions | C1 final journal/checkpoint | C2/C3 must cover every byte that C1 interpreted; C4 evaluates declarations only against an explicit selection context. |
| Complete capture graph, selection context, concrete platform/toolchain binding overlay, active graph, FFI, generated/extension/system edges | C4 | C5 may order only the unchanged selected capture-plus-binding union. |
| Build-DAG commands, environment, sandbox, build order, and expected outputs | C5 | C6 rechecks at use and audits observed behavior; expected nodes remain immutable. |
| Materialization, action observations, produced bytes | Separate C6 observations/receipt | C7 may publish only this causal output set; observations cannot change C4/C5/closure identity. |
| Protected location and reusable receipt | C7 | Reuse starts from a fresh expected C0-C5 input and exact protected hit. |

## Protected execution boundary

The acquisition and execution boundaries are deliberately different:

- Acquisition may use network only through a manager-owned broker that writes
  raw immutable origins to quarantine/protected capture. It runs no package
  code and imports no ambient installed tree as authority.
- Admission uses non-executable streaming readers and safe source-tree handles.
  It finishes before package-manager materialization, except for Cargo's
  separately confined, post-origin-admission vendor transform.
- Offline replay starts with empty task-specific home/config/cache/output/temp
  roots. It exposes the read-only admitted closure and selected toolchain/SDK,
  reconstructs any manager cache only from admitted inputs, denies credentials
  and ambient configuration, and enforces network denial outside the manager.
- Before C5, only the evidence-derivation invocations enumerated above may
  start, each after its permit is committed and its affected source tree is
  admitted. They may emit only resolution/admission evidence. After C5, only
  actions in the derived build DAG may start. Every executable resolves to a
  C0-selected evidence tool, a C4-selected build-only external toolchain, or a
  previously receipted local tool, and
  receives exact read roots, environment keys, child-process allowlist, write
  paths, and output classes.
- The output namespace starts empty and is disjoint from dependency/source
  inputs. Copying or hard-linking a pre-existing binary cannot establish local
  production. Publication validates class, path, size, hash, complete input,
  and protected ownership.
- Runtime network required by an installed CLI is not build network. For
  example, the currency scraper may use its service network when invoked by a
  user, but resolution/materialization/build remain networkless and its
  subprocess contract is still checkpointed.

## Diagnostics and precedence

Accepted `artifact_*`, `rust_*`, `swiftpm_*`, and shared `closure_*` codes keep
their meanings. The graph layer adds only gaps required by this task:

| Code | Stable meaning and required fields |
| --- | --- |
| `closure_graph_schema_unsupported` | Unknown graph/checkpoint node, edge, condition, action, boundary, or schema kind; schema/profile and offending kind. |
| `closure_graph_incomplete` | A selected dependency/source/tool/action/output/extension/interop edge has no unique immutable endpoint, or an unexpected endpoint appears; product, target, edge and evidence origin. |
| `closure_graph_reference_invalid` | Duplicate logical/node/edge key within or across capture/binding tables, disallowed binding-node/edge kind, duplicate semantic edge, dangling or wrong-kind endpoint or platform role, capture replacement, relational field inside a node payload, or action slot with zero/multiple bindings; canonical key/ID, owning table, expected endpoint/slot kind, and all conflicting origins. |
| `closure_derivation_unauthorized` | A pre-C5 executable invocation lacks a committed permit, names an unadmitted input, uses a tool absent from C0, or occurs outside its allowed checkpoint phase; invocation key/subtype, causal head, input/tool IDs, and start counter. |
| `closure_derivation_drift` | A permitted pre-C5 invocation's executable, argv, cwd, environment, process/read/write/network audit, output evidence, or receipt differs from its permit; invocation key, permit/receipt IDs, exact differing field, and whether the process started. |
| `closure_build_cycle` | Ordering SCC in `D_build`; sorted action IDs, sorted internal edge IDs, cycle digest, product/target, last checkpoint. |
| `closure_interop_undeclared` | Missing, escaping, ambiguous, or incompatible FFI/module/header/link/load/subprocess contract; provider/consumer, boundary mode, target/toolchain/ABI facts, path/symbol/protocol evidence. |
| `closure_checkpoint_invalid` | Predecessor, canonical payload, graph/plan/manifest reference, or downstream chain differs from independently derived expectation; checkpoint name, expected/observed IDs, differing field class. |
| `closure_target_identity_changed` | Target OS/arch/triple/ABI/libc/SDK/minimum-runtime/condition facts differ between selection, plan, and use. |
| `closure_process_undeclared` | An action starts or resolves a child executable absent from its plan; action, executable class/path, parent process and checkpoint. |
| `closure_input_undeclared` | An action reads a filesystem input or environment/config value absent from its plan; action, operation, canonical resource class and path/key. |
| `closure_write_undeclared` | An action writes, deletes, renames, links, or mutates outside its declared write/output set; action, operation and canonical path. |

Existing shared codes cover the rest:

- origin/lock/integrity before parsing;
- artifact archive/class/admission before graph/build errors;
- `closure_metadata_mismatch`, `closure_hook_undeclared`, and
  `closure_build_dependency_unlocked` during closure/planning;
- `artifact_toolchain_untrusted`, `artifact_toolchain_identity_changed`,
  `closure_runtime_identity_changed`, and target drift before derivation or
  build-action use;
- `closure_offline_input_missing`, `closure_network_attempted`, and the new
  process/read/write codes during C6; and
- `artifact_local_output_unreceipted`, `artifact_local_output_drift`, or
  `closure_generated_output_drift` before publication.

Diagnostic precedence is:

1. immutable origin/integrity failure;
2. recursive artifact safety/classification failure;
3. toolchain/runtime/target identity failure at the affected pre-use gate;
4. unauthorized or drifted evidence derivation;
5. graph schema/reference/completeness/metadata/interop/cycle failure;
6. unsupported capability or undeclared hook/build dependency;
7. offline network/process/read/environment/write violation; then
8. output/receipt drift.

Within one precedence class, findings sort by canonical node/path, rule
priority, and code. A recognized compiled dependency remains primary even if
the same package also asks for an unsupported native build.

## Reusable conformance vectors

Every vector publishes input bytes and graph declarations, expected canonical
artifact/graph/plan/checkpoint bytes and digests, expected code and failing
checkpoint, selected/pruned edge set, action start counters, network audit,
write set, and output receipt where applicable. Adapter wrappers may add
evidence but cannot weaken the expected semantic result.

### Positive semantic vectors

| ID | Fixture | Required result |
| --- | --- | --- |
| `CGP01` | Root product -> source package A -> transitive B -> C, exact locks/origins, source-only nested archive | Complete `G_capture` and `G_active`; deterministic build waves; clean offline replay and receipt. Wrap in Go baseline, Cargo, all Node managers, and SwiftPM where representable. |
| `CGP02` | Swift product -> Swift target -> C provider through module/header `c_abi` boundary | Separate targets, explicit provider/consumer edges, Clang/Swift/SDK/platform identities, provider-before-consumer build order, offline output receipt. Extend with restricted C++/Objective-C fixtures. |
| `CGP03` | Go frontend and Swift scraper built independently, co-published, Go invokes Swift through versioned JSON subprocess protocol | Two source closures and outputs, Swift/Go action order only where publication requires it, explicit runtime invocation and protocol identity, no hidden PATH lookup. |
| `CGP04` | Immutable shipped generated JS plus a separate declared TypeScript generation action | Shipped JS is `source.generated_text`; local JS/map are `generated_artifact`/`output_artifact` with compiler/plugin/config lineage and producer-before-consumer order. |
| `CGP05` | Exact CCJ seed above for Darwin arm64 and Linux x86_64 over one target-conditional dependency declaration | Capture bytes and ID remain exactly `sha256:1bcd31f3b5b1e1e77da9256c4395d59f75802df7a6d3dcef2504448c2c04f5f2`; platform nodes and explicit `targets` edges resolve only in their binding overlays; selection, binding, active, plan, C4, and C5 bytes/IDs match the published goldens and differ; the Linux condition is selected only for Linux. |
| `CGP06` | Node packages A and B form a runtime dependency cycle but need no prior compiled output | Canonical non-ordering SCC recorded; build/materialization DAG remains acyclic and replay succeeds. |
| `CGP07` | Permute manifest map, lock record, source enumeration, archive member, node, and edge input order | Exact same canonical manifests, graph bytes/digests, SCC records, waves, checkpoint chain, diagnostic ordering, and outputs. |
| `CGP08` | Replay once with empty ambient caches and once with poisoned but inaccessible caches/config/home | Identical materialized graph, outputs, receipts, and network=`none` audit. |
| `CGP09` | Declared steps produce object, library/addon where profile permits local production, executable, and multiple published files | `ALLOW_OUTPUT` only after causal action evidence; sorted multi-output receipt and exact protected reuse. Dependency copies of the same bytes remain rejected. |
| `CGP10` | Exact fully published action/output/capture/binding/plan/checkpoint/receipt golden above, then C6 observes three-byte outputs `one` and `two` | Every label+CCJ record independently derives its published ID. Action/output nodes, all graph edges, capture/binding/active graph, plan, C4, C5, closure, and expected-cache-input remain exactly fixed; produced-observation and execution/publication receipt IDs match the two distinct published branches. An observation never rewrites a node. |
| `CGP11` | Swift root plus two dependency manifests, Cargo vendor, Swift mirror replay, and manager metadata, all successful | C0 toolchain checkpoint precedes use; each complete Swift tree admits before its own manifest permit; C3a -> vendor permit/receipt -> C3b and C3 -> mirror/metadata permit/receipt ordering is explicit; final C1-C4 checkpoints bind the complete causal journal. |

### Negative semantic vectors

Every negative vector asserts no later checkpoint, no unpermitted action, no
published output/cache entry, and a deterministic last-successful-checkpoint.

| ID | Fixture | Expected result |
| --- | --- | --- |
| `CGN01` | Selected transitive, peer, workspace, source, build-tool, or output edge has no unique node; installed layout contains an extra package | `closure_graph_incomplete` at C4; layout cannot widen or repair the graph. |
| `CGN02` | Generator A requires B output and B requires A output; repeat with an FFI provider/consumer build cycle | `closure_build_cycle` before C5 with invariant SCC/edge set and cycle digest under input permutations. |
| `CGN03` | C/Swift/C++/Objective-C consumer reaches an absolute/escaping header, undeclared module/library/framework/symbol, ABI-incompatible provider, or implicit subprocess | `closure_interop_undeclared` or accepted more-specific Swift/toolchain diagnostic at C4/C5; no compiler/plugin starts. |
| `CGN04` | Generated source/header/config/plugin/output appears without producer lineage or declared path/class | `artifact_generated_input_undeclared` before consumption; extra runtime write is `closure_generated_output_drift`/`closure_write_undeclared`. |
| `CGN05` | npm implicit node-gyp/lifecycle, pnpm hook, Yarn plugin/Git pack hook, Cargo build script/proc macro, or Swift plugin/macro becomes active | Accepted profile-specific or `closure_hook_undeclared` diagnostic before extension compilation/execution. |
| `CGN06` | Any accepted `C01-C12` compiled fixture, including renamed/suffixless/nested/role-spoofed forms, through every adapter | Same shared class, leaf manifest, and `artifact_compiled_dependency_forbidden`; no manager hook/build. |
| `CGN07` | Opaque, ambiguous, encrypted, malformed, unsafe-path/entry, unsupported, over-limit, or incompletely read payload | Accepted `F01-F14` deterministic artifact diagnostic before graph/build. |
| `CGN08` | Package supplies a compiler/runtime under `vendor/toolchain`; approved external toolchain link escapes or bytes change | Dependency compiled rejection, `artifact_toolchain_untrusted`, or `artifact_toolchain_identity_changed`; no build. |
| `CGN09` | Target OS/arch/triple/ABI/libc/SDK/tag/condition changes after C0/C4/C5, or selection names a missing/wrong-kind platform node or omits/duplicates a required `targets` edge | `closure_graph_reference_invalid` at C0/C4 for invalid binding structure or `closure_target_identity_changed` for time-of-use drift; no cross-target cache reuse. |
| `CGN10` | Declared build attempts network, unplanned process, host-file/config/environment read, or out-of-set write | `closure_network_attempted`, `closure_process_undeclared`, `closure_input_undeclared`, or `closure_write_undeclared` at C6; no publication. |
| `CGN11` | Pre-existing dependency binary copied/hard-linked into output, or produced path/class/size/hash/write set changes | `artifact_local_output_unreceipted`, `artifact_local_output_drift`, or `closure_generated_output_drift`; protected entry absent/quarantined. |
| `CGN12` | Mutate a C0-C5 payload while retaining a stale downstream checkpoint/receipt/cache key | `closure_checkpoint_invalid`; an independently derived expectation prevents hit/reuse. |
| `CGN13` | Remove one required raw artifact/mirror/vendor/store member while an ambient cache has matching package identity | `closure_offline_input_missing`; ambient state cannot satisfy C6; no action consuming the missing node starts. |
| `CGN14` | Same captured lock under changed feature/marker/peer/target evaluator, platform binding, or package metadata that would select another edge | `closure_metadata_mismatch`, feature/target-specific diagnostic, or C4 graph drift; pruned edges and concrete bindings cannot silently change or become selected. |
| `CGN15` | Duplicate logical key/node ID/edge key across capture/binding tables, dangling or wrong-kind endpoint/platform role, forbidden binding kind, capture replacement, relation embedded in a node payload, or action slot bound zero/twice; repeat inside runtime cycle, action/output pair, and FFI boundary | `closure_graph_reference_invalid` at C4 with canonical owning table and conflicting origins; input permutations choose the same primary finding; no recursive hash attempt or C5 checkpoint. |
| `CGN16` | Swift root contains a compiled/opaque/unsafe leaf; separately, an admitted root discovers a dependency whose complete tree contains that leaf | Shared artifact diagnostic. Root case has `root_manifest_eval_count=0`; dependency case permits prior root evaluation but has `rejected_dependency_manifest_eval_count=0`; no affected permit, mirror replay, build, or publication occurs. |
| `CGN17` | C0-selected Swift/PackageDescription, Cargo, Node manager, or mirror tool bytes change after permit issuance and before manifest/vendor/metadata/mirror use | `artifact_toolchain_identity_changed` or `closure_runtime_identity_changed` before the affected process; affected invocation count is zero, no receipt is issued, and no later checkpoint exists. |
| `CGN18` | Pre-C5 manifest/vendor/metadata/mirror invocation lacks a permit, names an unadmitted input, or observes extra argv/environment/process/read/write/network/output evidence | `closure_derivation_unauthorized` before launch or `closure_derivation_drift` after the permitted process; no evidence from the invocation enters C1-C4 and no build/publication follows. |

### Reuse of accepted fixture families

- `TASK-260810-29vk09` `A*`, `C*`, `F*`, `T*`, and current-capability
  `V01` supply shared byte fixtures for `CGP01`, `CGP09`, and `CGN06-CGN12`.
- `TASK-260810-2n3sbi` `S*`, `N*`, and `P*` supply manager/independent-Python
  wrappers for `CGP01`, `CGP04-CGP08`, `CGP11`, and `CGN01`,
  `CGN04-CGN14`, `CGN17-CGN18`.
- `TASK-260810-3urqbl` `R*`, `RV*`, `GV*`, `VF*`, `PV*`, `RF*`, and `RH*`
  supply Cargo origin/transform/build wrappers, including the pre-vendor
  zero-spawn boundary.
- `TASK-260810-zddzh7` `S*`, `H*`, `R*`, and `P*` supply SwiftPM target,
  module-map, mirror, extension, and compiled-payload wrappers for `CGP01-03`
  and `CGP11`, plus `CGN01-18` where applicable.
- `TASK-260810-1veyfw` supplies the extensionless Mach-O, revision-state,
  Go/Swift subprocess, Swift-to-C, missing-lock, `.pyc`/venv, and Python drift
  examples. They become fixtures or explicit unsupported migration evidence,
  not permission to expand delivery scope.

## Ecosystem instantiation

| Profile | Graph specialization | Active extension decision | Toolchain/target identity |
| --- | --- | --- | --- |
| Go baseline | Existing module/package/source graph maps to package, source-set, target, compiler/link actions and one output. Go assembly exception stays Go-profile-specific; cgo/native edges remain rejected. | Active generators/cgo/host objects rejected by existing Go policy. | Existing full GOROOT fingerprint, Go target/tuning, fixed execution policy, exact protected receipt. Shared schemas must preserve behavior. |
| `rust-source-v1` | Lock-superset and exact target/feature graph; raw registry/Git/path source sets; pinned vendor-transform action/evidence; one package/bin target. | Build scripts, build dependencies, proc macros, native links, wrappers/config, unstable features, cross/custom targets rejected. | Physical Cargo/rustc, sysroot/target stdlib, linker/SDK, native triple/cfg, features/profile. |
| Node pure-source | Manager-specific package-instance/peer/workspace/condition graph feeds common runtime/generator actions. Raw packages remain authority; manager caches/install state are derived. | Dependency lifecycle, implicit node-gyp, manager hooks/extensions, native addons/Wasm rejected. Root TypeScript generation allowed only as exact declared action. | Node and manager full fingerprints/versions; platform/arch/libc/module/ABI/component identities; TS compiler/plugin/config where used. |
| Python reference | Produces/consumes the same protocol goldens and shared semantic expectations through an independent implementation. No repository/cache/venv sharing and no new adapter task. | PEP 517/Python rules remain the accepted reference outcome; current delivery adds no Python code. | Independent interpreter/frontend/backend/tag identities when the reference runs the fixtures. |
| `swiftpm-source-v1` | Root/pinned package snapshots, mirrors, product/target/source graph, separate Swift and Clang targets, module/header/FFI boundaries, selected/dormant extension reachability. | Active build plugins/macros, binary targets, unsafe flags, untrusted system libraries, registries/submodules/LFS and unsupported destinations rejected. | SwiftPM/swiftc/PackageDescription, clang/clang++/linker, SDK/sysroot/runtime, host/target, language modes/standards/C++ interop/traits/configuration. |

## Development-ready board decomposition

### Proportionality and gap audit

The implementation story now has fourteen active leaves. This is the smallest
decomposition that keeps one reviewable deliverable per task:

- three shared responsibilities have distinct trust boundaries: artifact
  admission, graph/checkpoint semantics, and protected execution/publication;
- Cargo has a pre-Cargo immutable intake/vendor-transform deliverable separate
  from executable build/toolchain behavior;
- Node has one common runtime/build contract plus four non-interchangeable
  manager lock/materialization implementations explicitly required by the
  accepted research;
- SwiftPM has source resolution/mirror closure, C-family interop validation,
  and offline execution/publication boundaries; and
- one final task owns only cross-adapter conformance/integration and its
  task-local documentation.

Combining any adjacent pair would cross a causal trust boundary or merge
independent package-manager implementations. Splitting further into docs,
quality gates, schema-only ceremony, or per-fixture leaves would not add an
independent deliverable.

No beyond-literal-spec task was created. Before creation, the audit checked
`SCOPE`, `SCI-1` through `SCI-6`, `VCAP`, `RR-GRAPH`, `DELIVERY`, and every
explicit exclusion. It rejected tasks for a new Python adapter, Kotlin/Dart/
.NET, verified-binary admission, non-SwiftPM native systems, external-system
admission, active Rust hooks/proc macros/native/cross support, Node native
addons, and Swift plugins/macros. No genuinely open research question remains,
so no research task was created.

### Active implementation leaves

| Task | One deliverable and traceability | Direct blockers | Estimate |
| --- | --- | --- | ---: |
| `TASK-260811-2gazym` — implement shared artifact admission policy | Shared detector/container/path/limit/manifest service. `SCI-1/2/4/6`, `VCAP`; accepted taxonomy. | Accepted synthesis `TASK-260810-1dgdos` | 13 |
| `TASK-260811-i3154q` — implement canonical closure graph and checkpoints | Selection-neutral capture, concrete selection-binding overlay, active projection, intrinsic nonrecursive nodes, immutable expected outputs, observations, cycles/order, C0-C7 codecs, and exact CGP05/CGP10 bytes. `SCI-1/2/5`, `RR-GRAPH`; this decision. | Accepted synthesis | 8 |
| `TASK-260811-27xisf` — implement protected closure execution and receipts | Manager-neutral intake/admission and pre-C5 derivation permit/receipt substrate, offline build isolation, multi-output observations/receipts, protected cache. `SCI-2/3/4`, `VCAP`, `DELIVERY`. | Artifact policy; graph/checkpoints | 13 |
| `TASK-260811-2h4m0s` — implement Cargo source capture and vendor transform | Cargo lock/manifest capture predicates, exact native-target/tool binding overlay, immutable registry/Git/path intake, pre-admission, C0-bound vendor/metadata permits, and pinned transform. `SCOPE`, `SCI-1/2/3/4/6`, `VCAP`; Cargo outcome. | All three shared foundations | 13 |
| `TASK-260811-3kbf3l` — implement Rust offline build adapter | Consume exact native platform/Cargo/rustc binding records; enforce Rust target/toolchain/profile rejection, frozen build, and output receipt without capture mutation. `SCOPE`, `SCI-2/3/4/5/6`, `DELIVERY`; Cargo outcome. | Cargo capture/transform | 8 |
| `TASK-260811-3twayo` — implement Node/TypeScript runtime and build plan | Common selection-neutral Node capture plus exact target/runtime/tool binding overlay and active graph, C0 manager identity and metadata permits, runtime/lifecycle/generator contract, and Python protocol goldens. `SCOPE`, `SCI-2/3/4/5/6`; Node/Python outcome. | All three shared foundations | 8 |
| `TASK-260811-1u42b9` — implement npm profile | npm lock/workspace/tarball/private-cache/offline materializer. `SCOPE`, `SCI-1/2/3/4/6`; npm research. | Common Node contract | 8 |
| `TASK-260811-3ksxig` — implement pnpm profile | pnpm lock/importer/peer/store/hook restrictions. Same spec sections; pnpm research. | Common Node contract | 8 |
| `TASK-260811-twq9ad` — implement Yarn Classic profile | Yarn 1 lock/workspace/offline-mirror materializer. Same spec sections; Yarn Classic research. | Common Node contract | 5 |
| `TASK-260811-32iojo` — implement modern Yarn profile | Modern Yarn lock/rc/plugin/cache/linker immutable materializer. Same spec sections; modern Yarn research. | Common Node contract | 8 |
| `TASK-260811-33ukne` — implement SwiftPM source resolution and closure | Root/dependency intake-before-manifest derivation journal, exact source capture, lock/mirrors, selection-neutral target predicates, concrete destination/tool binding overlay, and permitted offline graph replay. `SCOPE`, `SCI-1/2/3/4/6`, `VCAP`; SwiftPM outcome. | All three shared foundations | 13 |
| `TASK-260811-tkurtl` — implement SwiftPM C-family interop validation | Separate target languages, module/header/system/FFI declarations, and exact platform/toolchain/SDK binding-overlay validation. `SCOPE`, `SCI-4/5/6`, `VCAP`; SwiftPM outcome. | SwiftPM resolution/closure | 13 |
| `TASK-260811-2qfnai` — implement SwiftPM offline build adapter | Consume the exact platform/Swift/Clang/SDK binding overlay for isolated native build, complete read/write evidence, and output receipts without graph mutation. `SCOPE`, `SCI-2/3/4/5/6`, `DELIVERY`; SwiftPM outcome. | C-family interop validation | 8 |
| `TASK-260811-x611eq` — integrate cross-language adapter conformance | Shared capture/binding/selection/output/checkpoint/receipt goldens—including every published CGP05/CGP10 label+CCJ byte—plus adapter wrappers, E2E offline and denial proof, task-local support docs. `SCI-1..6`, `VCAP`, `RR-GRAPH`, `DELIVERY`. | Rust build; four Node manager profiles; SwiftPM build | 13 |

Each active leaf has explicit description, scope, acceptance criteria, three
task-specific checklist gates, `task_class=code`, `review=required`, and a
Fibonacci estimate. A developer can begin any dependency-unblocked leaf without
an unresolved product or architecture question.

Four original placeholders were closed, not deleted, before implementation
because their immutable checklists crossed the new atomic boundaries:
`TASK-260811-13xlp0`, `TASK-260811-2fwvml`, `TASK-260811-ojduq3`, and
`TASK-260811-1zwhx2`. Their notes point to the replacement leaves and their
blocker links were removed so resolved elements do not retain active blockers.

### Build order from the live task DAG

After this architecture task and the final synthesis are accepted, the
implementation waves are:

| Wave | Tasks |
| ---: | --- |
| 1 | `TASK-260811-2gazym` artifact admission and `TASK-260811-i3154q` graph/checkpoints in parallel. |
| 2 | `TASK-260811-27xisf` protected execution/receipts. |
| 3 | `TASK-260811-2h4m0s` Cargo capture/transform, `TASK-260811-3twayo` Node common contract, and `TASK-260811-33ukne` SwiftPM resolution/closure in parallel. |
| 4 | `TASK-260811-3kbf3l` Rust build; npm, pnpm, Yarn Classic, and modern Yarn profiles; and `TASK-260811-tkurtl` SwiftPM interop in parallel. |
| 5 | `TASK-260811-2qfnai` SwiftPM offline build. |
| 6 | `TASK-260811-x611eq` cross-adapter conformance/integration. |

The live related-scope critical path is:

```text
TASK-260810-1uu9lk
-> TASK-260810-1dgdos
-> TASK-260811-2gazym
-> TASK-260811-27xisf
-> TASK-260811-33ukne
-> TASK-260811-tkurtl
-> TASK-260811-2qfnai
-> TASK-260811-x611eq
```

The artifact-policy and graph tasks are parallel roots because both implement
accepted contracts. Protected execution depends on both. Ecosystem tasks name
their direct shared APIs even when one shared dependency is also reachable
transitively; those are real compile/test dependencies, not scheduling
ceremony.

## Unsupported cases retained explicitly

| Excluded or limited case | Current disposition |
| --- | --- |
| New Python adapter or shared Python/Node repository/cache/runtime | No task. Export common protocol fixture schemas/goldens only. |
| Kotlin/Gradle/Maven, Dart, .NET | Deferred; no active dependency or research task. |
| Verified dependency binaries | `verified-binary-v1` absent; compiled dependency candidates reject. No implementation leaf. |
| Rust build scripts, build dependencies, proc macros, native links, custom/cross targets, nightly/unstable features | Explicit `rust-source-v1` rejection and negative fixtures. Future capability only under new approval. |
| Node native addons, native source builds, WebAssembly, lifecycle-required dependencies, unapproved manager plugins/hooks, Git/HTTP mutable sources | Explicit pure-source profile rejection. |
| Swift active plugins/macros, binary targets, unsafe flags, registries, submodules/LFS, arbitrary host libraries, non-SwiftPM C-family projects, unvalidated Objective-C/C++ targets | Explicit `swiftpm-source-v1` rejection or restricted target fixture. |
| External system commands such as `glab` or `sentry-cli` | Remain outside language source closure and require a separate future system-dependency policy; no gap-closing task in this cycle. |
| Universal output bit reproducibility | Not claimed. Exact logical identity and protected causal receipts are required; any stronger profile claim needs fixture evidence. |

## Requirement-to-evidence audit

| Requirement | Architecture evidence | Implementation evidence required |
| --- | --- | --- |
| `SCI-1` recursive enumeration and immutable identity | Selection-neutral `G_capture`, explicit concrete binding overlay and selection/activation records, package/source nodes, artifact-manifest refs, intrinsic node IDs, edge-only relations, and C1-C4. | Artifact, graph, Cargo capture, all Node manager, and SwiftPM resolution tasks; `CGP01/05/10`, `CGN01`, `CGN06/07/13-16`. |
| `SCI-2` locks/checksums/snapshots/build metadata/toolchains | Nonrecursive identity layers, concrete C0/C4 tool/platform bindings, immutable expected outputs, derivation permits/receipts, first-binding table, C0-C7, and fully published CGP10 receipt encodings. | Every shared and adapter task; exact selection/binding/output/checkpoint goldens, `CGP05/10/11`, and `CGN17/18`. |
| `SCI-3` offline reconstruction | Protected execution boundary and C6. | Protected executor plus Rust/Node/Swift build/materialization tasks; `CGP01/08`, `CGN10/13`. |
| `SCI-4` no hidden hook/generator/build inputs | Edge-only action relations, pre-C5 derivation plane, intake-before-interpretation rule, C4-C6, and derivation/execution diagnostics. | Protected executor, Node common, Rust and Swift rejection gates; `CGP04/11`, `CGN04/05/10/11/16-18`. |
| `SCI-5` mixed-language graph/order/toolchain/platform/FFI | Target/action/boundary declarations in capture; exact toolchain/platform nodes and typed bindings in the selection overlay; nonrecursive edge authority; cycle/order algorithm; interop modes. | Graph task, Swift interop/build, Rust target evidence, Node runtime, integration; `CGP02/03/05/07/10`, `CGN02/03/08/09/15`. |
| `SCI-6` conservative support | Closed enums/profile gates, unsupported table, deterministic diagnostics. | All adapter AC and negative fixtures; no best-effort mode. |
| `VCAP` global compiled rejection | Shared trust/class role, C3 gate, FFI non-bypass, output causal role. | Artifact policy and every adapter wrapper; accepted `C*`/`T*`, `CGN06/08/11`. |
| `RR-GRAPH` reusable conformance | `CGP01-11`, `CGN01-18`, accepted-family mapping, exact selection/output goldens, and stable diagnostic/checkpoint assertions. | Integration task publishes exact fixture/golden bytes and runs every wrapper. |
| `DELIVERY` atomic implementation backlog | Fourteen active traced leaves, linked DAG, estimates, task-local checklists, superseded history. | Producer/reviewer lifecycle through accepted implementation; no research-only task is treated as delivery. |

The evidence above proves architecture/decomposition completeness for
`TASK-260810-1uu9lk`; it does not claim that any new adapter is already
implemented.

## Decision log and anomalies

1. The common graph and artifact manifest are separate canonical records.
   Flattening archive leaves into graph nodes is unnecessary and would obscure
   dependency semantics; omitting the artifact-manifest reference would hide
   bytes.
2. Node package/runtime cycles require explicit non-ordering SCC support.
   Treating every dependency edge as build order would reject valid source-only
   graphs; allowing action cycles would make output causality ambiguous.
3. FFI is a boundary node, not a string edge. ABI/module/toolchain/platform
   evidence must be shareable by provider and consumer and independently
   hashable.
4. Node IDs contain only intrinsic immutable declarations. Package/source,
   producer/consumer, interop, tool, platform, and publication relationships
   have one authority: the edge table. C6 output facts are observations, not
   mutations of expected nodes.
5. Capture is selection-neutral. Requested target/features/markers/peers,
   concrete platform/toolchain nodes and typed bindings, and selected/pruned
   results live in separate selection, binding, and active records. The exact
   two-target golden makes capture contamination, hidden `targets` edges, and
   dangling/wrong-kind platform references visible.
6. Cargo vendoring is an evidence derivation inside admission, not the origin.
   Its C0 toolchain, C3a permit/receipt, and C3b treatment preserve the accepted
   zero-Cargo-spawn rule for rejected raw inputs and verify the transform.
7. SwiftPM manifest evaluation is executable resolution logic, not a passive
   parse. The root and each fetched package tree admit before that package's
   manifest permit; the iterative journal is replayed later from mirrors.
8. Pre-C5 manager execution cannot borrow authority from C5. Manifest, vendor,
   mirror, and metadata invocations bind their exact toolchain, command,
   environment, I/O policy, and observations before use.
9. Package-manager caches and installed trees are derived materialization
   evidence. None can substitute for raw origin and graph checkpoints.
10. The existing inventory's extensionless Mach-O and Swift module-map escape
   are cross-language fixtures, not adapter-specific exceptions.
11. No architecture diagram was produced. The node/edge/checkpoint tables and
   explicit build-wave sequence convey the relationships without an additional
   planning artifact.
12. Canonical goldens are not hash-only claims. Every CGP05/CGP10 hash now has
    its domain label and exact CCJ payload inline; CGP10 also publishes every
    referenced fixture record through expected cache input and both receipt
    branches so an implementation cannot invent an incompatible encoding.

## Verification

The current artifact and board were checked after all mutations:

- UTF-8, trailing-whitespace, required-heading, prerequisite-citation,
  balanced-fence, and Markdown-table structure harness:
  `architecture_artifact=pass lines=1245 tables=18 bytes=112189`;
- task-scoped independent verifier
  `TASK-260810-1uu9lk_canonical-golden-verifier.rb`, run as
  `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb`,
  canonicalizes and hashes every published record and validates all fixture
  references, cross-table node/edge kinds, platform roles, target/tool/read/
  write slots, checkpoints, and receipt branches. Its exact output is:
  `canonical_goldens=pass labeled_records=53 cgp05_target_branches=2
  cgp10_observation_branches=2` and
  `canonical_references=pass cgp05_capture_reused=true
  explicit_target_bindings=2 cgp10_all_refs_resolve=true`;
- live implementation-leaf query: exactly 14 active backlog tasks, each with
  `task_class=code`, `review=required`, a Fibonacci estimate, `Spec trace:`,
  nonempty scope and AC, exactly three task-specific checklist items, and at
  least one blocker; the nine directly affected contracts contain their F1/F2
  obligations: `implementation_leaves=pass active=14 closed_superseded=4
  refined_contracts=9`;
- related-scope planning query: the same eight acyclic waves and critical path
  recorded above: `dependency_plan=pass waves=8`;
- `task-board validate`: `Board is valid. No issues found.`
- uncached full repository gate: `go test -count=1 ./...` exited 0 across every
  package; notable timings were `cmd/curator` 358.819s,
  `internal/godriver` 67.989s, `internal/install` 100.744s, and
  `internal/install/atomicity` 105.524s.

No product code, tests, or configuration were changed by this solution-
architecture task. The full repository suite was nevertheless rerun to satisfy
the task gate and protect compatibility; the architecture, canonical-golden,
live-board, dependency-plan, board-validation, and repository-test results are
the authoritative handoff evidence.

## References

- `.spec/skill-facing-cli-source-closure.md`
- `TASK-260810-1veyfw_accepted-inventory-language-and-reference-surfaces.md`
- `TASK-260810-29vk09_accepted-compiled-artifact-taxonomy.md`
- `TASK-260810-2n3sbi_accepted-node-typescript-python-closure-research.md`
- `TASK-260810-3urqbl_accepted-rust-cargo-source-closure.md`
- `TASK-260810-zddzh7_accepted-swiftpm-mixed-c-family-source-closure.md`
- `.research/260811_cross-language-closure-canonical-golden-verifier.rb`
- Existing Go baseline: `internal/buildsource`, `internal/buildmeta`,
  `internal/buildcache`, `internal/godriver`, `internal/buildrepo`, and
  `internal/install`.
