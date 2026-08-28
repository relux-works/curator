# Authoring a language source-closure adapter

This guide is for contributors adding a new language-utility source-closure
adapter to Curator. It describes tested behaviour only. It is not a support
matrix or a migration guide: use [Cross-adapter source-closure
conformance](source-closure-adapter-conformance.md) for supported profiles,
unsupported cases, diagnostics, and migration of an existing command.

The delivery requirements originate in [`.spec/skill-facing-cli-source-closure.md`](../.spec/skill-facing-cli-source-closure.md),
especially **Source closure invariant**, **Vendored compiled artifact
prohibition**, and **Delivery completion**. The canonical record and checkpoint
implementation is [`internal/closuregraph`](../internal/closuregraph), the
process authority is [`internal/closureexec`](../internal/closureexec), recursive
byte admission is [`internal/artifactpolicy`](../internal/artifactpolicy), and
the integration contract is
[`internal/crossconformance`](../internal/crossconformance). Do not create an
adapter-local substitute for any of those authorities.

## The contract to implement

An adapter supports a build only when it establishes the complete predicate:

```text
supported_build(selection) =
    immutable_recursive_capture(selection)
    AND complete_recursive_artifact_admission
    AND closed_selection_and_build_graph
    AND trusted_toolchain_and_target_binding
    AND authorized_declared_execution_only
    AND offline_replay_from_empty_ambient_state
    AND causally_receipted_protected_outputs
```

False, unknown, unsupported, partially inspected, or non-reproducible evidence
fails before the affected step. A lock entry, checksum, package-manager cache,
installed tree, filename suffix, or successful unmanaged build is not a proxy
for a missing conjunct. This is the predicate in
[the conformance document's **The one predicate** section](source-closure-adapter-conformance.md#the-one-predicate),
proved by the shared services in `internal/artifactpolicy`,
`internal/closuregraph`, and `internal/closureexec`.

### C0-C7 is the adapter lifecycle

Implement the closed checkpoint chain through the interfaces and payloads in
[`internal/closuregraph/adapter.go`](../internal/closuregraph/adapter.go) and
[`internal/closuregraph/checkpoint.go`](../internal/closuregraph/checkpoint.go).
Each successful checkpoint hashes its canonical payload and exact predecessor;
a failed checkpoint is not issued and cannot seed downstream identity.
`ValidateCheckpointChain` in `internal/closuregraph/checkpoint.go` proves the
sequence and permits Cargo's additional C3a/C3b admission gates.

| Checkpoint | What the adapter must establish | Identity and evidence carried forward |
| --- | --- | --- |
| `C0.profile` | Select the adapter profile, schemas, artifact policy, source grammars, limits, configuration policy, capabilities, platform roles, and every external tool allowed to derive evidence before C5. | `selection_context_id`, exact platform nodes, and evidence-toolchain nodes. Proof: `closuregraph.C0ProfilePayload`. |
| `C1.resolve` | Freeze root/workspace declarations, the authoritative lock candidate, unevaluated conditions and their evaluator identities, candidate records, and the causal derivation journal. | Lock candidate, selection context, candidate graph records, and journal entries. Proof: `closuregraph.C1ResolvePayload`. |
| `C2.capture` | Protect every immutable origin input before interpretation or derivation. | Intake receipts, immutable origin IDs, protected handles, and broker receipts. Proof: `closuregraph.C2CapturePayload` and `closureexec.CaptureStore`. |
| `C3.admit` | Recursively classify and admit every captured dependency byte through the shared policy. | Intake receipts, artifact-manifest IDs, and any permitted derivation receipts. Cargo additionally uses `C3a.origin-admission` and `C3b.derived-admission`. Proof: `closuregraph.C3AdmitPayload`, `internal/artifactpolicy`, and `internal/rustsource`. |
| `C4.close` | Reconcile the unchanged capture with one exact selection overlay and its active projection. | `captured_graph_id`, `selection_context_id`, `selection_binding_id`, and `active_graph_id`. Proof: `closuregraph.C4ClosePayload` and `closuregraph.GraphBundle.Validate`. |
| `C5.plan` | Derive one immutable, deterministic, acyclic build plan from the active graph. C5 does not add or repair graph records. | `build_plan_id`. Proof: `closuregraph.C5PlanPayload` and `internal/closuregraph/plan.go`. |
| `C6.offline` | Execute only the committed plan with empty ambient state and reconcile the observed process, reads, writes, network result, tools, and produced artifacts. | A separate `execution_receipt_id` and immutable produced-artifact observations. Proof: `closuregraph.C6OfflinePayload`, `closuregraph.ExecutionReceipt`, and `internal/closureexec`. |
| `C7.publish` | Publish only causally produced, reinspected outputs through the protected store; exact reuse requires an independently derived expected cache input. | `publication_receipt_id`. Proof: `closuregraph.C7PublishPayload`, `closuregraph.PublicationReceipt`, and `internal/artifactpolicy` local-output admission. |

### Capture first; bind the exact target later

Implement `closuregraph.CaptureAdapter` so capture is selection-neutral. It may
contain immutable declarations, the conservative lock/resolution superset,
conditions, abstract action/tool/platform slots, and artifact-manifest
references. It must not contain an exact target platform, a concrete external
toolchain component, a `targets` or `uses_tool` edge, requested target values,
or selected/pruned results. The interface contract is in
[`internal/closuregraph/adapter.go`](../internal/closuregraph/adapter.go), and
the negative shape is enforced by
`crossconformance.CheckSelectionNeutralCapture` in
[`internal/crossconformance/suite.go`](../internal/crossconformance/suite.go).

Implement `closuregraph.SelectionAdapter` as the only authority that introduces
the exact target and concrete tool identities. Its binding overlay may add only
the binding node and edge kinds accepted by `closuregraph.GraphBundle.Validate`;
exact destination must resolve once and be reached by an explicit `targets`
edge wherever the path emits binding records. It may not replace a capture
record or add a package, source, target unit, action, generated artifact,
interop declaration, or expected output. These restrictions are proved by
`internal/closuregraph/validation.go` and
`crossconformance.CheckBindingOwnsTargetAuthority`.

For the same source inputs, two destinations must reuse one exact capture while
their selection-bound identity, active projection, plan, and downstream
checkpoint identities diverge. `crossconformance.CheckTargetDivergence` is the
shared proof. This is why an adapter publishes one capture and a target matrix
of bindings, plans, executions, and receipts rather than recapturing sources
per destination.

## Scaffold the package in causal order

1. Define a versioned profile ID, accepted manager/lock schemas, source
   grammars, capability set, stable diagnostic constants, and the narrow
   supported boundary. Reuse shared diagnostic codes without renaming their
   cause. Existing shapes are demonstrated by `internal/rustsource/errors.go`,
   the four `internal/*source/errors.go` Node-manager packages, and
   `internal/swiftpmsource/errors.go`; the global vocabulary is documented in
   [**Diagnostics**](source-closure-adapter-conformance.md#diagnostics).
2. Parse only an authoritative committed lock or frozen resolution record. Bind
   every recursive instance to immutable origin and integrity evidence, and
   reject missing, stale, unsupported, mutable, ambiguous, or incomplete
   authority. The worked implementations are `internal/rustsource/parse.go`,
   `internal/npmsource/lock.go`, `internal/pnpmsource/lock.go`,
   `internal/yarnclassicsource/lock.go`,
   `internal/yarnmodernsource/lock.go`, and
   `internal/swiftpmsource/lock.go`.
3. Intake raw origins into `closureexec.CaptureStore` before executing a
   manifest, vendor transform, mirror replay, or manager metadata query. Run
   recursive byte and structure admission through `artifactpolicy.Service`;
   never infer admission from a manager's installed layout. The intake-before-
   interpretation contract is proved by `internal/closureexec/intake.go` and by
   the capture paths in `internal/rustsource`, the four Node manager packages,
   and `internal/swiftpmsource`.
4. Emit the selection-neutral capture, exact binding overlay, active graph, and
   deterministic plan with `internal/closuregraph`. Preserve selected and
   pruned conditions in evidence, reject unresolved references and execution
   cycles, and bind every action slot exactly once. The record validation and
   projection rules live in `internal/closuregraph/validation.go`,
   `projection.go`, and `plan.go`.
5. Declare every executable step, including generators. A generator declaration
   binds its tool, configuration, input set, output paths/classes, target, and
   producer-before-consumer edges; generated bytes without that lineage reject.
   The action and generated/output record contracts are in
   `internal/closuregraph/action_template.go`, `node.go`, and `edge.go`; the
   Node example is `internal/nodesource` generated-output planning.
6. Issue the C0-C7 chain and retain all permits, receipts, observations, and
   protected publication evidence needed to reproduce each identity. The
   checkpoint codecs in `internal/closuregraph/checkpoint.go` and execution
   models in `internal/closureexec/models.go` are the normative schemas.

## Process seams and guards

Adapter production code must contain zero direct `exec.Command*` calls and
must not import `os/exec`. The only production files permitted to launch a
child process are
[`internal/closureexec/acquisition.go`](../internal/closureexec/acquisition.go)
and
[`internal/closureexec/portable_runner.go`](../internal/closureexec/portable_runner.go).
Those manager-owned seams commit an acquisition or derivation permit before
start and return the corresponding receipt. The permit/receipt validation is
implemented by `internal/closureexec` and exercised from the accepted adapter
packages.

When adding a package, extend guard *coverage*, not the allowlist. Add the new
production directory to an adapter-level guard modelled on
[`internal/swiftpmbuild/guard_test.go`](../internal/swiftpmbuild/guard_test.go),
[`internal/swiftpminterop/guard_test.go`](../internal/swiftpminterop/guard_test.go),
or `internal/rustsource/build_test.go`. Scan the new surface together with
`internal/closureexec`; keep the allowlist exactly
`acquisition.go` and `portable_runner.go`. Also keep
[`internal/crossconformance/guard_test.go`](../internal/crossconformance/guard_test.go)
green: its integration oracle starts no process and its production files import
only the standard library.

Tests may launch real fixtures, but a test-only launch is not adapter authority.
Production behavior must still drive one of the two shared seams and reconcile
the returned receipt. This boundary is proved by the guard tests above and by
`closureexec.Executor` receipt reconciliation in
`internal/closureexec/executor.go`.

## Source-text analysis: reject by default

Do not emulate a language toolchain frontend to claim a complete source
closure. A source scanner may admit only constructs for which its accepted
grammar proves that every dependency-bearing interpretation is visible; every
other construct fails closed. This follows **Source closure invariant 4 and 6**
in `.spec/skill-facing-cli-source-closure.md` and is exercised most
adversarially by `internal/swiftpminterop`.

The SwiftPM C-family work exposed four independent axes that a source-text
analyzer must close:

- **Spelling:** alternative preprocessing spellings, line splices, stringized
  and pasted tokens, directive channels, and any other accepted spelling must
  be proven equivalent or rejected. Proof:
  `internal/swiftpminterop/headers.go` and its parser/module-map tests.
- **Position:** every syntactic position where expansion can change dependency
  meaning must be enumerated; the accepted SwiftPM analyzer includes Objective-C
  `@` keyword positions and expanded module-import names. Proof:
  `internal/swiftpminterop/headers.go` and
  `internal/swiftpminterop/parser_test.go`.
- **Build-setting kind:** source text is not the only frontend input. Enumerate
  every supported build-setting kind, reject unknown kinds, and reject any kind
  whose effect on macro, include, link, language, or isolation behavior cannot
  be proved safe. Proof: `internal/swiftpminterop/buildsettings.go` and
  `buildsettings_test.go`.
- **Macro-oracle input:** close every route that can define a macro, including
  both source `#define` directives and build settings that become compiler
  `-D` arguments; analyze both the macro name and replacement body through the
  same rejection logic. Proof: `internal/swiftpminterop/headers.go`,
  `buildsettings.go`, and the `H24`/`H25` (`Q1`-`Q6`) vectors in
  `buildsettings_test.go`.

Closing only one axis is not evidence for the others. If a contributor cannot
state and test the complete spelling × position × build-setting-kind ×
macro-oracle-input surface for the pinned frontend, the safe result is a stable
unsupported or undeclared-input diagnostic, not another heuristic. The
reject-by-default outcome is preserved by the portable path in
`internal/swiftpminterop/readset_test.go`.

### Boundary with observed reads

Static rejection and observed-read evidence have different authority.
`swiftpminterop.ReadSetProvider` reports `not-observed` in portable assurance;
it starts no process and does not upgrade compiler dependency output into proof.
Verified mode requires an observed read set plus the issued derivation receipt,
and rejects missing or incomplete observer authority. This separation is
implemented in `internal/swiftpminterop/boundaries.go` and proved by
`internal/swiftpminterop/readset_test.go`.

[`internal/swiftpmbuild/readset.go`](../internal/swiftpmbuild/readset.go) is the
accepted OBSERVED-READ provider. It performs a permitted, isolated, network-
denied offline build, consumes the selected compilers' dependency files, maps
work-copy reads back to admitted protected roots, separates derived build-tree
reads, and leaves all other reads for the binding resolver to classify. Its
verified path is proof only when `internal/closureexec` supplies the required
provider capability and causal receipt; portable mode remains reject-by-default.
The build and reconciliation boundary is proved by
`internal/swiftpmbuild/conformance_test.go` and
`internal/swiftpmbuild/swiftpmbuild_test.go`.

Do not use an observed-read provider to excuse an incomplete static contract.
The static adapter defines what may be attempted and the verified provider
supplies adversarial completeness for what the selected toolchain actually
read. Both authorities must agree before C4/C5. This ordering is implemented by
`internal/swiftpminterop` consuming the provider before closure and by
`internal/swiftpmbuild` refusing verified planning without observed reads.

## What a new path publishes

### Canonical vectors

Publish the adapter's projection of both accepted canonical vectors without
inventing a new encoding:

- **CGP05, two targets:** fixed source inputs produce one byte-identical
  selection-neutral capture. Concrete platform nodes and `targets` edges appear
  only in each binding overlay. Platform, selection, binding, active graph,
  plan, C4, and C5 identities diverge between destinations. The exact record
  family is checked by `internal/crossconformance/corpus.go` and
  `validate.go`; see [the conformance corpus description](source-closure-adapter-conformance.md#the-cross-adapter-conformance-suite).
- **CGP10, two observation branches:** artifact manifest, action/output nodes,
  graph edges, capture, binding, active graph, plan, C4, C5, closure, and
  expected-cache-input records stay fixed. Produced observations and execution
  and publication receipts diverge for the distinct output bytes; an
  observation never rewrites a graph node. The independent verification is in
  `internal/crossconformance/ccj.go`, `corpus.go`, and `validate.go`.

The independent CCJ-1 scanner/emitter in `internal/crossconformance` must derive
the same identities without importing `internal/closuregraph` or any adapter.
Regenerate and commit
[`internal/crossconformance/testdata/cross-adapter-protocol-export.json`](../internal/crossconformance/testdata/cross-adapter-protocol-export.json)
when the accepted corpus, delivered paths, obligations, or rejection matrix
changes. `internal/crossconformance/export.go`, `export_test.go`, and
`guard_test.go` prove this publication boundary.

### Seven obligations

Add the path to `crossconformance.DeliveredPaths` and make its real production
entry points discharge all seven obligations in
[`internal/crossconformance/suite.go`](../internal/crossconformance/suite.go):

| Obligation | Author evidence to provide |
| --- | --- |
| `capture.selection_neutral` | A capture census with no selection-only node or edge kind. |
| `capture.stable_across_targets` | Two exact destinations over one input set reuse the same capture identity. |
| `binding.target_authority` | Only the overlay introduces the exact target and tools, with an explicit `targets` edge where records are emitted. |
| `binding.diverges_per_target` | Selection-bound identity, active projection, and plan all differ between destinations. |
| `records.deterministic` | Fresh repeated capture/projection reproduces every identity. |
| `evidence.causal_chain` | Every checkpoint names its predecessor, C5 adds no graph record, and each pre-C5 executable derivation returns a causal receipt. |
| `artifact.shared_admission` | The shared deny payload has the same class, decision, primary diagnostic, and leaf digest; the new profile admits only its declared source grammars. |

`crossconformance.Coverage` must remain complete; never mark an obligation as
covered without driving the adapter's production API. The consumer-oriented
meaning and current path matrix stay in [**The cross-adapter conformance
suite**](source-closure-adapter-conformance.md#the-cross-adapter-conformance-suite)
rather than being duplicated here.

### Rejection matrix

Drive every applicable cross-drivable entry returned by
`crossconformance.RejectionVectors` through the new adapter. Each outcome must
return a code from the vector's closed set, start no affected process, and
publish nothing; `crossconformance.CheckRejection` enforces all three facts.
In particular, recognized compiled dependency bytes must retain the primary
`artifact_compiled_dependency_forbidden` result even when the package also
declares an unsupported hook or native build. That precedence is implemented by
`internal/artifactpolicy` and asserted by the adapter conformance suites.

Do not copy or fork the matrix. Its graph, artifact, identity, execution, and
output vectors (and the delegated `network-attempted`, `undeclared-write`, and
`output-drift` ownership) are listed in
[the conformance document's rejection matrix](source-closure-adapter-conformance.md#the-cross-adapter-conformance-suite)
and defined once in `internal/crossconformance/suite.go`. If a new ecosystem
cannot construct a delegated vector without forging sealed evidence, add its
owning package's negative suite and compile-time diagnostic reference following
`internal/crossconformance/guard_test.go`; do not mint fake receipts in the
integration harness.

## Ecosystem-specific work

The shared contract does not make manager formats interchangeable. Use these
packages as worked examples and keep the new ecosystem's authority explicit.

| Concern | Author requirement and proof |
| --- | --- |
| Rust/Cargo | Require the committed `Cargo.lock`, admit raw registry/Git/path origins before the pinned vendor transform, and bind the native target plus Cargo/rustc identities. Rust v1 rejects build scripts, build dependencies, proc macros, native links, package Cargo config, and cross/custom targets. Proof: `internal/rustsource/parse.go`, `manager_vendor.go`, `graph.go`, and `build_conformance_test.go`. |
| npm | Parse only the supported `package-lock.json` or `npm-shrinkwrap.json` schemas, verify every tarball against lock SRI, and materialize from the private cache with `--offline --ignore-scripts`. Reject lifecycle dependencies and implicit `binding.gyp`/`node-gyp`. Proof: `internal/npmsource/lock.go`, `capture.go`, `materialize.go`, and `conformance_test.go`. |
| pnpm | Preserve importer, peer-instance, patch, integrity, and private-store semantics from the supported `pnpm-lock.yaml`; disable scripts and reject `.pnpmfile.*`, side-effects cache, and unbound patches. Proof: `internal/pnpmsource/lock.go`, `patch.go`, `materialize.go`, and `conformance_test.go`. |
| Yarn Classic | Require the root Yarn v1 lock plus the exact accepted Yarn/config identity; populate and replay only the captured offline mirror with scripts disabled. Proof: `internal/yarnclassicsource/lock.go`, `capture.go`, `materialize.go`, and `conformance_test.go`. |
| Modern Yarn | Bind the supported modern lock, `.yarnrc.yml`, cache key/compression, linker, and built-in plugin set; disable lifecycle scripts and reject unapproved plugins, custom fetchers/resolvers/linkers, or mutable cache semantics. Proof: `internal/yarnmodernsource/lock.go`, `capture.go`, `materialize.go`, and `conformance_test.go`. |
| SwiftPM and C family | Freeze top-level `Package.resolved`, capture kind-preserving mirrors, evaluate every package manifest through a permit/receipt, keep Swift and Clang targets separate, and bind SwiftPM, Swift, Clang, linker, SDK, and exact destination identities. Reject binary targets, active plugins/macros, unsafe settings, escaping headers/module maps, and untrusted system libraries. Proof: `internal/swiftpmsource`, `internal/swiftpminterop`, and `internal/swiftpmbuild` conformance suites. |

Across ecosystems, dependency payloads are source-only under the profile and
are recursively inspected by `internal/artifactpolicy`; lifecycle scripts and
manager extensions are disabled or rejected by the manager package; every
generator is an explicit `closuregraph` action; and every destination gets its
own binding, plan, execution, and publication receipt over the shared capture.
The consumer support boundary and unsupported cases remain authoritative in
[**Supported profiles**](source-closure-adapter-conformance.md#supported-profiles)
and [**Explicit unsupported cases**](source-closure-adapter-conformance.md#explicit-unsupported-cases).

## Tests and evidence

Before proposing a path, provide all of the following evidence:

1. A package-local normative suite alongside the existing per-adapter suites
   (`conformance_test.go` in `internal/{npmsource,pnpmsource,yarnclassicsource,yarnmodernsource,swiftpminterop,swiftpmbuild}`,
   `build_conformance_test.go` in `internal/rustsource`, and
   `swiftpmsource_test.go`/`swift_integration_test.go` in `internal/swiftpmsource`).
   Drive the real capture, bind, plan, materialize/build, and publish
   entry points. Include positive fixtures and negative fixtures for lock,
   integrity, artifact, graph, target/toolchain, hook/generator, offline I/O,
   and receipt failures applicable to the profile. Those package suites prove
   the current profile-specific behavior.
2. Negative proof for every gate. Narrow or mutate the gate so the relevant
   test fails; do not rely on a delete-only mutant or a helper that production
   never calls. This expectation is embodied by the negative adapter suites and
   by `crossconformance.CheckRejection`, which requires a real error, zero
   affected starts, and zero publication.
3. Cross-adapter coverage for the seven obligations, CGP05/CGP10 projections,
   every applicable rejection vector, guard coverage, and a byte-exact protocol
   export. The normative entry point is `go test -count=1
   ./internal/crossconformance`; its completeness and committed-export checks
   are in `suite_test.go`, `rejection_test.go`, and `export_test.go`.
4. Focused package coverage and race evidence, followed by the repository CI
   gates. The repository's tested gates are `make ci-test`, `make race`,
   `make ledger-check`, `make gate-selftest`, `go vet ./...`, the `gofmt` check,
   and `golangci-lint run`; their exact platform planning and evidence outputs
   are defined by [`.github/workflows/ci.yml`](../.github/workflows/ci.yml),
   [`.github/ci/test-gate.sh`](../.github/ci/test-gate.sh), and the `Makefile`.
   Record the actual command and exit code. A coverage measurement is evidence,
   not permission to leave an unproved gate; no repository-wide percentage can
   replace the obligations and negative vectors above.
5. Store raw local logs under `.temp/<TASK-ID>/`, attach a task-scoped developer
   outcome and material artifacts to the board, and commit only protocol
   artifacts that are part of the public contract. The repository workflow and
   evidence location are defined in [`CONTRIBUTING.md`](../CONTRIBUTING.md),
   while the committed independent interface is the cross-adapter protocol
   export named above.

If any required property remains unknown, report it as unknown and reject the
profile shape. Do not add an allow flag, synthetic receipt, local parser
exception, or test-only authority to make `supported_build` appear true. That
fail-closed rule comes from `.spec/skill-facing-cli-source-closure.md` **Source
closure invariant 6** and is enforced end to end by the package suites and
`internal/crossconformance`.
