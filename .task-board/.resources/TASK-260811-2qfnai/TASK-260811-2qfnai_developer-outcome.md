# Developer outcome — TASK-260811-2qfnai SwiftPM offline build adapter

Task: `TASK-260811-2qfnai` — implement offline planning, build, and publication
for the `swiftpm-source-v1` adapter.

Status: implemented, tested, gates green, handed to review.

## What was delivered

New package `internal/swiftpmbuild` implementing the C5 plan / C6 offline build
/ C7 publication boundary of `swiftpm-source-v1`.

| File | Deliverable |
| --- | --- |
| `errors.go` | Shared diagnostic vocabulary. No shared cause is renamed at this phase boundary; `closure_input_undeclared`, `closure_write_undeclared`, `closure_process_undeclared`, and `closure_network_attempted` are re-raised unchanged. |
| `types.go` | Closed build tool-slot set, binding projection, plan, command, object-slot, and result records. |
| `binding.go` | Exact C4 overlay resolution and the single-resolution rule. |
| `plan.go` | Republished C4', product link action, expected output, deterministic C5 plan, portable command identity, closure and publication authority. |
| `build.go` | Protected-cache probe, offline build permit, command/read/write reconciliation, output reinspection, sorted observations, C6/C7 checkpoints, protected publication and exact reuse. |
| `readset.go` | The observed-read provider: `swiftpminterop.ReadSetProvider` backed by a real offline compile pass. |

## Binding overlay (AC: platform, SwiftPM, swiftc, PackageDescription, Clang, linker, SDK)

`Config.Slots` names only the **binding role** each build slot must resolve to.
Every physical identity — fingerprint, executable digest, relative path, ABI,
policy selector, version — is read back from the accepted C4 binding node, so
this stage restates no toolchain fact. The single exception is the linker,
which is the only component the build stage itself selects; it is published as
a new binding node `swiftpm.build.component.<role>` with its own `targets` and
`requires` edges and its own `ToolchainSelector` in the binding authority.

Fail-closed cases, all before any process starts:

- missing slot role, unknown component role, two slots resolving to one node,
  wrong-kind node, dangling binding node, binding node the selection binding
  never references, duplicate component role, a slot the selection does not use;
- an incomplete linker identity (`artifact_toolchain_untrusted`);
- physical drift at the immediate time-of-use recheck
  (`artifact_toolchain_identity_changed`), rechecked again inside the executor's
  toolchain-recheck callback;
- `SlotClangCXX` is required exactly when a selected target carries C++ or
  Objective-C++ source, and rejected otherwise.

`validateActionBindings` proves the AC's single-resolution rule over the whole
republished active graph: every selected action binds exactly one target
platform and resolves each declared tool slot exactly once, and every
`uses_tool` edge names the same executable as its bound component.

## Identity preservation through C5

`preservesAcceptedIdentities` proves the build stage only *adds*: every accepted
capture node, capture edge, binding node, and binding edge identity from the
interop closure is present unchanged in the republished tables. The build C4'
chains from the accepted interop C4, which itself chains from the source C4; a
foreign capture, a foreign selection context, or an unchained interop closure is
rejected with `closure_checkpoint_invalid`.

## Offline execution (AC: fresh roots, mirrors, forced pins, network none, prebuilts disabled)

The committed permit is built from the accepted capture only:

- the offline build root is the admitted root tree plus the frozen
  `Package.resolved` and a generated kind-preserving `mirrors.json`; it is
  re-admitted through the shared recursive artifact classifier and the capture
  store, bound to the frozen lock digest and the command identity;
- every admitted mirror is mounted read-only at its isolated mount path;
- the work copy is private and retained only for reconciliation;
- `Network: "none"`, `RecheckRule: "immediate-exact-v1"`, explicit resource
  limits, exact expected evidence, exact read/write roots, exact allowed
  processes;
- exact argv: `--disable-experimental-prebuilts`, `--force-resolved-versions`,
  `--disable-netrc`, `--build-system native`, isolated `--cache-path`,
  `--config-path`, `--security-path`, `--scratch-path`, one `--configuration`,
  one `--triple`, one `--product`;
- `HOME` and every `SWIFTPM_*` directory point inside the private work copy;
- the output root must be empty before the permitted action
  (`artifact_local_output_unreceipted`).

The planned command keeps `{execution-root}` as a placeholder so the portable
command identity excludes temporary paths; the concrete root is substituted only
when the permit is committed. Two isolated execution roots produce the same
`CommandID` (proved by `TestPlannedCommandIdentityIsPortable`).

`reconcileCommand` proves the issued receipt describes exactly the committed
command, and rejects portable receipts that inflate assurance, verified receipts
that observed a network attempt or an undeclared read/write/process, a receipt
whose before/after toolchain fingerprints differ, and a receipt that omits the
declared product.

## Publication (AC: sorted observations and receipts, no mutation of expected records)

Every declared output is reinspected from the retained private work copy, hashed,
validated against the immutable graph records, sorted by exact observation
identity, staged, and published atomically through
`closureexec.ProtectedStore`. The C6 execution receipt reports the plan's own
stable wave order, `network=none`, the exact C4/C5 write set, and the sorted
observation identities; C6 chains from C5 and C7 chains from C6.

An exact protected-cache hit, derived independently from the expected cache
input, short-circuits before any process starts and returns the same artifact
path and cache-input identity (`TestOfflineBuildPublishesAndReusesExactly`).

## The observed-read provider (the key architectural role)

`ReadSetObserver` implements `swiftpminterop.ReadSetProvider`. It performs one
offline native SwiftPM build from the accepted capture, under an isolated home
with network denied and resolution frozen, and answers every per-target read-set
request from the dependency files the compilers themselves emitted
(`<Target>.build/*.d`). This is what makes the C-family header closure
adversarially complete: the proof stops depending on statically reproducing
Clang's search behaviour and starts depending on what the selected compiler
actually read.

Mode separation is strict and unchanged for tkurtl's portable path:

- **portable**: the operating-system boundary cannot confine reads, so a
  compiler-emitted dependency file is corroboration and not proof. The observer
  returns `Observed: false` and the interop stage keeps its reject-by-default
  portable verdict. No process is started at all.
- **verified**: the observer runs the pass, requires
  `AssuranceVerified` and `audit.network == "none"` on the issued receipt, and
  returns the observed reads with the issued derivation receipt identity. The
  build stage additionally refuses to plan a verified build whose accepted
  interop closure is not `Reads.Mode == "observed"`
  (`swiftpm_header_input_undeclared`).

Observed reads are mapped back exactly: a read of the private work copy is
rewritten to the admitted protected tree it was copied from; a read inside the
private build tree is a locally produced output of the same permitted,
network-denied derivation and is not part of the admitted source closure; every
other read is reported verbatim so the interop binding resolver classifies it
against the selected roots and fails closed on anything undeclared.

The Make-style dependency grammar is parsed exactly, including line
continuations, backslash-escaped spaces, escaped colons, and Windows drive
letters.

## Upstream fix in `internal/swiftpminterop` (please review this deliberately)

Implementing publication surfaced a **real defect at the boundary between the
accepted interop closure and the accepted shared publication contract**, not a
local inconvenience:

- `closuregraph.PublicationEvidence.ValidateForPublication` requires
  `execution.WriteSet` to equal the graph-derived set of paths produced by
  selected actions, and requires observations to cover exactly the C5
  `DeclaredOutputNodeIDs` (which contain `output_artifact` nodes only);
- `closureexec.ProtectedStore.Publish` additionally requires the staged tree and
  the sorted observation paths to equal `execution.WriteSet`.

Together these mean: **every path produced by a selected action must be a
declared output artifact with a real observation.** The accepted interop closure
declared one `generated_artifact` object *per target*, which (a) can never be
observed and therefore made the whole SwiftPM closure unpublishable, and (b)
does not exist: SwiftPM's native build system emits **one object per source
file**, verified against the real toolchain (Apple Swift 6.3.2) for Swift and
Clang targets, flat and nested.

Fixed at source rather than worked around locally:

- `swiftpminterop` now declares one `output_artifact` per source file
  (`OutputRole: "intermediate"`, `ExpectedClass: "native.object"`) at
  `.curator/objects/<package>/<target>/<package-relative-source>.o`, one
  `produces` edge per source bound to a per-source write slot
  (`ObjectWriteSlot(index)`), and one `targets` binding edge per selected object;
- `swiftpmbuild` resolves each declared object to the exact file SwiftPM
  produced — a Clang target mirrors the source path below `<Target>.build`, a
  Swift target flattens it to the source base name — requires a unique
  resolution, and fails closed with `artifact_local_output_unreceipted` when an
  object is absent or ambiguous;
- both are published with real observed bytes and digests.

Nothing is synthesized and no compensating record is invented. The two existing
interop tests that named the old logical keys were updated; the rest of the
tkurtl suite is unchanged and green. tkurtl's portable reject-by-default read
verdict is untouched.

Two smaller upstream additions, both additive:

- `swiftpmsource.Capture.RootInput()` and `Capture.OfflineMirrors()` expose the
  admitted root tree and the admitted kind-preserving mirrors with their exact
  isolated mount paths, so the build stage replays from the same admitted bytes
  instead of rediscovering an origin;
- `swiftpminterop.TargetInterop.ObjectNodeIDs` records the declared per-source
  object outputs in source order.

## Process boundary

`internal/swiftpmbuild` starts no process of its own and does not import
`os/exec`. `guard_test.go` extends the existing cross-adapter guard to the new
surface over `swiftpmbuild`, `swiftpminterop`, `swiftpmsource`, and
`closureexec` without widening its allowlist: `acquisition.go` and
`portable_runner.go` remain the only permitted crossings.

## Tests

New: `internal/swiftpmbuild/{fixture,swiftpmbuild,conformance,guard,swift_integration}_test.go`.

- **Manager/authority**: incomplete authority, non-absolute roots, output root
  outside the execution root, protected store inside the execution root, absent
  manager and plan.
- **CGP05 target binding**: the binding names the platform and every selected
  build tool identity; the selection-neutral capture carries no platform or
  toolchain node; every selected action binds one platform and resolves each
  declared tool slot exactly once; the link action resolves `build-driver` and
  `linker` exactly once each.
- **Binding rejection**: missing role, unknown role, two slots on one node,
  linker role duplicating an accepted binding, an unused slot, incomplete linker
  identity, drifted component, failing recheck.
- **Identity preservation**: every accepted node/edge survives; exactly two
  capture nodes are added; C4' chains from the interop C4; C5 chains from C4';
  a foreign capture is rejected.
- **Planned command**: exact offline/frozen/isolated argv, no temporary path in
  the planned environment, exact output path, portable command identity.
- **S01/S02/S03/S10**: one declared object per source for every selected target;
  a multi-source C-family target declares and publishes every object; a pruned
  target contributes no action and no declared output.
- **R01/R05/R06**: admitted mirrors mounted read-only with network denied; a
  lock pin without a captured mirror fails closed; an unfrozen lock fails closed.
- **Offline build and receipts**: publication, exact reuse from an
  independently derived expected input with zero process starts, exactly one
  launched process with exactly the committed argv, `PATH` confined to the
  execution root, C6/C7 chaining, three observations with real digests.
- **Fail-closed without publication**: missing declared object, missing product,
  pre-existing output root, graph drift, ambiguous produced object, output drift.
- **Reconciliation**: inflated portable assurance, drifted command, toolchain
  drift, missing declared product, verified network attempt, verified
  undeclared read/write/process, unsupported assurance mode.
- **Read-set provider**: portable mode reports not-observed and starts no
  process; incomplete observer authority is rejected; build-tree reads are
  separated from admitted source reads; the dependency grammar is parsed
  exactly.
- **Real toolchain** (`//go:build darwin || linux`, skips without `swift`):
  the exact planned argv reaches SwiftPM and builds; the planned scratch
  directory reproduces SwiftPM's triple naming; every declared object slot
  resolves to a real produced object for flat and nested Clang sources and for
  multi-source Swift targets; the emitted dependency files parse to a non-empty
  observed read set.

## Gates

All commands run directly as standalone processes; exit codes are the real ones.

| Gate | Command | Exit |
| --- | --- | ---: |
| Repository suite (excluding `cmd/curator`) | `go test -timeout 30m -count=1 $(go list ./... \| grep -v cmd/curator)` | 0 |
| CLI suite | `go test -timeout 30m -count=1 ./cmd/curator/...` | 0 |
| Focused suite | `go test -count=1 ./internal/swiftpmbuild/ ./internal/swiftpminterop/ ./internal/swiftpmsource/ ./internal/closureexec/ ./internal/closuregraph/ ./internal/artifactpolicy/` | 0 |
| Race (focused) | `go test -race -count=1 ./internal/swiftpmbuild/ ./internal/swiftpminterop/ ./internal/swiftpmsource/ ./internal/closureexec/` | 0 |
| Race (graph/policy) | `go test -race -count=1 ./internal/closuregraph/ ./internal/artifactpolicy/` | 0 |
| Lint | `golangci-lint run ./...` (v2.12.2) | 0, `0 issues.` |
| Format | `gofmt -l ./cmd ./internal` | 0, empty |
| Vet | `go vet ./...` | 0 |
| Whitespace | `git diff --check` | 0 |
| Canonical goldens | `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | 0, `canonical_goldens=pass labeled_records=53 cgp05_target_branches=2 cgp10_observation_branches=2` / `canonical_references=pass cgp05_capture_reused=true explicit_target_bindings=2 cgp10_all_refs_resolve=true` |
| Board | `task-board --no-update-check validate` | 0, `Board is valid. No issues found.` |

Logs are attached as task resources. Digests:

- `full-suite-01.log` `4ffa076e12d34ba0390715af9ad39dd02ba96d2cbf207be4fd5e8e89485befda`
- `cmd-curator-01.log` `dfe81ab681e7824c94efcb22a2b184881146275cc53e2f19ea13fd9d76feb457`
- `focused-01.log` `720f9ac0ee3e4a8af07cef19cddf9487ed92961df1544ea218c17df29e116960`
- `race-01.log` `e9d87493ef5cf7542470de7819c7a13ff30d6b0b3f895330ca86c473295c173d`
- `lint-01.log` `e92606b0bf483111dff0a120c315ea165821348f31365020e2468a0059095c47`
- `canonical-01.log` `1847364de63c9d8e706d54739f97d05e8406cca5f9cd7aec4bb12c028998b75a`
- `gofmt-01.log`, `vet-01.log`, `diffcheck-01.log` are empty
  (`e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`).

## Design decisions worth a reviewer's attention

1. **Where produced objects are declared.** Interop declares the compile action
   and its per-source object outputs; the build stage declares only the product
   link action and the product. An action's declared write slot must be bound
   exactly once, so production cannot be attached to an accepted action from a
   later stage — the declaration has to live where the action lives.
2. **Per-source, not per-target, objects.** Verified against the real toolchain.
   The previous per-target declaration was both unobservable and factually
   wrong; collapsing N objects into one artifact would have required
   synthesizing bytes, which is exactly the kind of compensating hack the
   stop-the-line rule forbids.
3. **Verified-only observed reads.** Portable execution cannot confine reads, so
   claiming observation there would inflate assurance. The observer is
   fail-closed by construction and the interop portable path is unchanged.
4. **Two build passes in the verified pipeline.** The observation pass runs
   before the interop closure (it is the read-set provider); the publication
   build runs after C5. Both are committed, network-denied, isolated
   derivations. This is inherent to making the read set an input to the closure.
5. **Object-resolution fail-closed rule.** A declared object that resolves to
   zero or more than one produced file is `artifact_local_output_unreceipted`;
   the adapter never guesses.
6. **`--disable-sandbox` is deliberately not passed.** Curator owns the outer
   isolation; SwiftPM's own sandbox stays on as defence in depth.

## Not done / out of scope

- `TASK-260811-x611eq` (cross-adapter conformance) was not started.
- No files were staged or committed; the working tree carries this delta on top
  of the preserved tkurtl delta.
- A lossless verified `EnforceObserveProvider` does not exist on any supported
  platform yet (`closureexec.NewOSBoundary` fails closed), so the verified
  observed-read path is exercised through the seam and its policy gates rather
  than against a real OS-enforced boundary. That is the pre-existing platform
  state, recorded honestly rather than papered over.
