# Decision 0008: additional-language driver, version, and artifact boundary

## Context

Protocol `1.0.0-rc.5` admits exactly two compiled drivers, `go-v1` for source
inside the consuming skill snapshot and `go-repository-v1` for source in a
locked external Git repository. Decision 0004 section "Future-driver rule",
decision 0005 section "Credential, signing, and future-driver ownership", and
`protocol/core.md` section 12.3 all require a new closed identifier and an
independent review for every additional language, and forbid admitting one by
widening an existing driver or adding a generic fallback.

Rust, Swift, and Kotlin driver pairs are now being designed in parallel with a
shared toolchain-requirement contract. Each of those designs would otherwise
have to pick, on its own, a manifest schema version, a receipt numbering rule,
an artifact shape, a descriptor evolution, a process-graph story, and a platform
claim. Whichever landed first would silently fix those choices for the others,
and the resulting wire surface would be an accident rather than a decision.

The three candidate languages also differ from Go in one decisive way. `go
build` executes no package-selected code. `cargo build` runs `build.rs` and
expands procedural macros; SwiftPM compiles and executes `Package.swift` as a
program and can run plugins and macros; the mainstream Kotlin build path is
Gradle, which is a general-purpose script engine. `SECURITY.md` already names
all three of those surfaces as things a conforming manager MUST NOT invoke. The
question is therefore not how to accommodate them but whether each driver can be
defined without them at all.

Kotlin raises one further question the other languages do not. Its dominant
output is a JVM archive that cannot run without a separately installed runtime,
which does not fit the single-artifact identity that receipts, markers, shims,
currentness, and garbage collection are built on.

This decision fixes the version, source-ownership, artifact, and execution
boundary that the three driver contracts, the toolchain contract, the wire
integration, and the candidate qualification must all satisfy. It defines no
compiler pipeline, adds no schema file, regenerates no vector, creates no
release metadata, advances no pin, and makes no platform claim.

## Decision

### 1. Version boundary

The additional drivers are a new protocol surface named `1.0.0-rc.6`. It is
reserved by this decision and minted only by the wire integration task; rc.5
remains the current candidate and its conformance-manifest digest and pins are
unchanged by this decision.

| Surface | Current | Next | Status |
|---|---|---|---|
| manifest `agent-skill-vN` / `csk-skill-vN` | 7 | 8 | reserved |
| repository descriptor `skill-build.json` | 1 | 2 | reserved |
| build receipt, local source mode | 1 | 3 | reserved |
| build receipt, external source mode | 2 | 4 | reserved |
| install marker | 3 | 4 | reserved |
| conformance claim | 3 | 4 | reserved |
| `Skillfile.dev.json` | 2 | 2 | unchanged |
| execution policy | `manager-worker-v1` | `manager-worker-v1` | unchanged |
| native-control inventory | `rc5-native-control-inventory-v1` | unchanged | unchanged |
| capability-evidence record | `capability-evidence-v1` | unchanged | unchanged |
| build-source identity | `curator-build-source-v1` | unchanged | unchanged |
| Go toolchain identity | `curator-go-toolchain-v1` | unchanged | frozen, Go only |

Manifest schemas 1 through 7, `skill-build.json` schema 1, build receipt
schemas 1 and 2, install marker schemas 1 through 3, conformance claim schemas 1
through 3, `Skillfile.dev.json` schema 2, and every rc.4 and rc.5 conformance
byte are frozen. Schemas 1 through 7 MUST reject every one of the six driver
identifiers below and every schema-8-only field. A reader or writer MUST NOT
reinterpret, widen, relabel, or infer an rc.6 meaning from an rc.4 or rc.5
object.

Build receipt schema versions are allocated per source mode per admitting
protocol version. rc.5 allocated 1 to the local Go driver and 2 to the external
Go driver. This boundary allocates 3 to the local source mode and 4 to the
external source mode for the drivers admitted at manifest schema 8. In a
schema-8 manifest `go-v1` still writes receipt schema 1 and `go-repository-v1`
still writes receipt schema 2; neither is re-versioned, re-hashed, or migrated.

`Skillfile.dev.json` schema 2 is deliberately not re-versioned. An operator
development substitution selects source acquisition only. It cannot select a
driver, a toolchain, an artifact class, or a target, so admitting more drivers
changes nothing it can express.

### 2. Six closed driver identities

The driver identifier space is closed by enumeration, never by grammar,
detection, or a family prefix. Protocol 1.0 admits exactly these values and
rejects every other:

| Driver | Language family | Source mode | Receipt schema | Status |
|---|---|---|---|---|
| `go-v1` | Go | local snapshot | 1 | admitted in rc.4 |
| `go-repository-v1` | Go | external repository | 2 | admitted in rc.5 |
| `rust-v1` | Rust | local snapshot | 3 | reserved |
| `rust-repository-v1` | Rust | external repository | 4 | reserved |
| `swift-v1` | Swift | local snapshot | 3 | reserved |
| `swift-repository-v1` | Swift | external repository | 4 | reserved |
| `kotlin-native-v1` | Kotlin | local snapshot | 3 | reserved |
| `kotlin-native-repository-v1` | Kotlin | external repository | 4 | reserved |

The identifier form is `<family>-v<n>` for the local source mode and
`<family>-repository-v<n>` for the external source mode, matching the deployed
Go pair. The Kotlin family segment carries the `native` backend qualifier
because Kotlin has two candidate backends and only the native one can satisfy
section 3; a later JVM-oriented driver, if it is ever reviewed and accepted,
MUST use a different family segment and MUST NOT reuse these two identifiers.

Reservation is not admission. Each identifier becomes valid wire only when its
own driver contract is accepted and integrated. A reserved identifier whose
contract is rejected is retired unused: it MUST NOT be reassigned to a different
language, backend, artifact class, or source mode, and it MUST NOT be enabled by
relaxing another driver.

Every schema that expresses a driver MUST express it as a `const` inside a
`oneOf` over the admitted identifiers. A driver field MUST NOT be an open
`enum`, a pattern, a bare string, or a value derived from a file extension,
project-metadata filename, directory layout, or any other language detection. No
manifest, descriptor, repository, substitution, or receipt may contain a
`language`, `toolchain_family`, `build_system`, `backend`, or comparable
selector.

### 3. Artifact class: `native-executable-v1` only

This version admits exactly one artifact class, `native-executable-v1`:

- exactly one bounded regular file, produced into operation-private manager
  staging, hashed there, and published immutably under the manager-home
  mutation lock;
- named solely by the manager from the consuming manifest command key, as
  `bin/<command>` on Unix and `bin/<command>.exe` on Windows, exactly as
  `protocol/core.md` sections 4.2 and 4.2.2 already require;
- directly executable by the host program loader using only libraries the target
  platform provides in its base installation, as fixed by the driver's own
  platform matrix; and
- never executed by the manager during validation, installation, status, repair,
  rollback, or garbage collection.

No driver may publish a second file. Debug information, separate symbol files,
program databases, import libraries, module or interface files, resource
bundles, shared libraries, sysroot copies, and incremental-build state are
compiler by-products. They remain in operation-private staging, are discarded
with it, and MUST NOT enter cache identity, the receipt, the marker, the shim
relationship, or publication.

The `runtime-bundle` class is rejected for this version. A runtime bundle is any
artifact that requires a manager-generated launcher, an interpreter or virtual
machine, a classpath or module path, a runtime image, or more than one published
file. It is rejected because:

- receipts, markers, shims, currentness, and garbage collection are all built on
  exactly one artifact path and one artifact digest, and a bundle would require
  every one of those identities to be redefined;
- a manager-generated launcher is manager-authored executable content whose
  contents would be derived from package data such as a main class and a
  classpath, which reintroduces the install-time execution surface the protocol
  exists to exclude;
- the required runtime would be an execution-time dependency the manager cannot
  fingerprint, cannot bind into cache identity, and cannot verify at install
  time, so a marker could claim currentness for an artifact that no longer runs;
  and
- `protocol/core.md` section 12.1 requires a shim to point exactly at the
  marker-selected protected artifact, which a bundle cannot satisfy without
  widening that rule.

A driver and platform pair that can only produce an artifact needing a
manager-published or manager-installed sidecar runtime file MUST NOT be admitted
for that platform in this version. It fails with
`build_artifact_class_unsupported` rather than gaining a sidecar, a launcher, or
an installer step. A future runtime-bundle profile requires its own artifact
class identity, its own receipt, marker, and claim schema versions, a launcher
generation contract, a runtime identity and verification contract, and its own
review. It MUST NOT be admitted by widening `native-executable-v1` or any of the
eight driver identifiers.

This version does not require bit-reproducible artifacts. Cache identity is
keyed on the canonical build input, and the artifact digest records what a
specific operation produced. A toolchain step that legitimately embeds
non-reproducible bytes, such as a linker-applied ad-hoc signature, is compiler
output and not a manager signing step, but it MUST be produced by the driver's
fixed argument vector without selecting a signing identity, credential, or
network interaction.

### 4. Local source ownership: context-excluded build roots

Local drivers reuse the schema-6 and schema-7 `build_roots` model without
change. A local build root MUST be a portable relative path other than `.`, MUST
name a real link-free directory in the immutable raw skill snapshot, MUST be
unique and pairwise disjoint, MUST NOT equal, contain, or be contained by a
runtime root, and MUST be referenced by at least one local build command. Build
roots are statically excluded from agent context and from the runtime copy
before cache lookup or any compiler discovery, and that exclusion applies
identically to real builds, exact cache hits, and dry-runs.

A schema-8 local build command has exactly this package-controlled surface, with
`driver` drawn from the closed local set:

```json
{"type":"build","driver":"rust-v1","source_dir":"build/cmd/tool"}
```

The object MUST contain exactly `type`, `driver`, and `source_dir`. No driver
may add a package-controlled member. In particular a manifest MUST NOT express a
binary, product, crate, module, target-name, feature, profile, configuration,
optimization level, argv, environment, flag, tag, linker option, toolchain path,
output name or path, install destination, alias, PATH edit, signing identity,
credential, hook, plugin, macro, generator, recipe, post-build action, fallback,
or secondary artifact.

Each local driver MUST bind exactly one closed driver-defined project-metadata
file that MUST exist directly in the build root and MUST be the nearest ancestor
of `source_dir`, exactly as `go-v1` binds `go.mod`. A manager MUST NOT discover,
search for, or infer that file, the module, the target, the command, or the
output. Every non-standard module, package, source, embedded, and vendored input
selected by the driver's fixed graph phase MUST remain below the command's build
root.

Each local driver MUST define a deterministic, non-discovering mapping from
`source_dir` to exactly one compiled program. A driver that cannot define such a
mapping without a new package-controlled member MUST be rejected for this
version. Widening the command object is not an option, because the consuming
manifest command key is the sole executable name and the sole naming authority.

### 5. External source ownership: `skill-build.json` schema 2

The external envelope of decision 0005 is unchanged. The consuming manifest owns
`build_repositories` and the command key; the repository owns only the
descriptor; the manager owns acquisition, validation, audit, naming,
publication, and every process.

`skill-build.json` remains the sole descriptor filename, at the repository root
and nowhere else, with no alias and no implementation-specific name. Schema 2
changes exactly one thing: the target object's `driver` becomes a `oneOf` over
the four admitted external identifiers. The target object still contains exactly
`driver`, `build_root`, and `source_dir`, and the document still contains
exactly `schema_version` and a non-empty `targets` map. No member is added.

A descriptor MUST NOT express a build, install, test, or run command, argv,
environment, feature, profile, configuration, product, target name, output name
or path, toolchain version, channel, path or download location, platform list,
signing identity, credential, hook, plugin, macro, generator, recipe, post-build
action, fallback, or secondary artifact.

The repository owns the descriptor schema version and the consuming manifest
owns nothing about it. A manager that supports schema 8 MUST read descriptor
schemas 1 and 2. Descriptor schema 1 remains frozen and can express only
`go-repository-v1` targets. A `rust-repository-v1`, `swift-repository-v1`, or
`kotlin-native-repository-v1` command therefore requires a schema-2 descriptor;
against a schema-1 descriptor it MUST fail closed with
`build_descriptor_driver_unsupported`, and against an unsupported descriptor
version with `build_descriptor_schema_unsupported`. Neither failure may fall
back to another target, another driver, `go-v1`, a script, a system command, or
a generic build facility.

The command and descriptor drivers MUST be equal, and both MUST name the same
external identifier. The whole repository snapshot remains the validation,
identity, and audit subject; only the selected build root is compiler-visible;
and no external repository byte is agent-facing or runtime-copied. The
schema-6 prohibition on a local `build_root` equal to `.` is unaffected by the
descriptor's admission of `.` as a repository root.

### 6. Toolchain boundary

The manifest toolchain requirement, trusted resolution, version grammar and
comparison, two-stage preflight ordering, diagnostics, and installation-guidance
catalog are defined by the shared toolchain contract, decision 0007
(`TASK-260728-1g0z69`), and integrated by `TASK-260728-2jaw7h`. This decision
does not restate or redefine them. It fixes only the boundary properties the
version and artifact model depends on:

1. Every new driver's canonical build input MUST bind a complete trusted
   toolchain identity produced by that contract. `curator-go-toolchain-v1` stays
   frozen and Go-only; no other driver may reuse, extend, or alias it.
2. The trusted toolchain is operator-owned. No manifest, descriptor, repository,
   substitution, or environment value may supply a toolchain path, URL, channel,
   version pin, mirror, trust root, installer, or package-manager command, and no
   version of any driver may auto-install a toolchain.
3. Every executable started below the worker MUST be a fingerprinted member of
   the driver's declared trusted toolchain closure. When a compiler requires a
   platform linker, SDK, sysroot, runtime library, or archiver that is not inside
   the fingerprinted distribution, the driver MUST bind that component into its
   toolchain identity or MUST reject that platform. A host-resolved tool outside
   the closure is not admissible.
4. Because the toolchain identity is inside the canonical build input, two
   toolchains never alias in the cache, the receipt, the marker, or a claim.
5. Host availability and version preflight runs before source acquisition, and
   the project-metadata compatibility cross-check runs after local validation or
   exact external acquisition and audit and before any compiler work. This
   ordering is owned by the toolchain contract and is restated here only because
   the artifact and audit boundary assumes it.

### 7. Execution policy and process graph

All eight drivers run under the existing portable `manager-worker-v1` execution
policy. No new execution-policy identity is minted, the mandatory portable
control set of `protocol/core.md` section 4.2.1 is unchanged, the exhaustive
`rc5-native-control-inventory-v1` inventory is unchanged because no control is
added, removed, or re-scoped, and the closed `capability-evidence-v1` record is
unchanged.

The fixed process graph is restated once, generically, without changing what it
requires:

```text
manager parent
  -> identity-verified manager-owned worker
       -> the driver's fingerprinted trusted launcher
            -> fingerprinted regular executables inside that driver's
               fingerprinted trusted toolchain closure
```

The two lower nodes were always per-operation values bound by the toolchain
identity inside the build input rather than constants of the policy; for `go-v1`
and `go-repository-v1` they resolve exactly as before to `<GOROOT>/bin/go` and
`<GOROOT>/pkg/tool/<GOHOSTOS>_<GOHOSTARCH>/`. Every `go-v1` and
`go-repository-v1` canonical build input, logical cache key, receipt byte
sequence, marker record, and claim is therefore unchanged by this decision.

One worker session performs exactly one read-only graph phase of at most one
driver-defined command, waits while the parent validates the complete graph,
accepts exactly one authenticated build permit, and performs exactly one
compile phase of exactly one driver-defined command. The session admits no
retry, second graph phase, second compile phase, additional executable, shell,
VCS operation, dependency download, generator, test, run, or tool request. A
driver that cannot map its pipeline onto that session shape MUST NOT be admitted
in this version; widening the session shape requires a new execution-policy
identity, a new claim schema version, and its own review.

Before the compile phase, each driver MUST apply an exhaustive, deterministic,
pre-compile rejection matrix computed from the validated snapshot and its graph
phase, in the same position where `go-v1` rejects `SysoFiles`, native inputs,
and the non-standard `//go:cgo_import_dynamic` directive. The matrix MUST reject
every package-selected code-execution surface for that language, including build
scripts, procedural and compiler macros, compiler and build-system plugins,
annotation processors, source generators, manifest programs, build tasks and
recipes, response files, package-selected linkers and native libraries, and
network or registry access. A surface that cannot be rejected deterministically
before the compile phase MUST cause the driver to be rejected, and MUST NOT be
answered by a runtime allowance, an advisory warning, or a sandbox promise. The
shared semantic class for such a rejection is
`build_package_code_execution_forbidden`; each driver contract defines its own
per-surface diagnostics beneath it.

### 8. Cache, receipt, marker, and claim identity

`curator-build-source-v1` is unchanged and is reused as the source identity for
every driver in both source modes. It hashes bytes, not language, so no new
source-identity algorithm is created. The protected external snapshot key of
`protocol/core.md` section 9.4 is likewise unchanged.

Each logical cache key remains the SHA-256 of `CCJ-1` over the complete build
input. Every new build input MUST bind its receipt schema version, driver,
source state for its mode, consuming command name, build root and source
directory selection, native target, complete trusted toolchain identity, and a
closed per-driver policy object that includes the execution-policy identity.
Inputs from different drivers cannot alias, because `driver` and the policy
object differ.

Build receipt schema 3 covers the local source mode and schema 4 covers the
external source mode, each as a strict `oneOf` discriminated by the `driver`
`const`, and each carrying that driver's own toolchain identity, native target,
and closed policy object. Receipt schemas 1 and 2 keep their bytes and meanings.

Install marker schema 4 permits `skill_schema_version` through 8 and represents
local-only, external-only, and mixed command sets across receipt schemas 1
through 4. Every build entry MUST explicitly record its driver, its
`receipt_schema_version`, and its `execution_policy`, and a reader MUST validate
the recorded receipt version against the closed driver table of section 2 and
reject a mismatch rather than infer an absent value from a driver name. Marker
schemas 1 through 3 keep their shapes.

Marker v4 generalizes exactly one schema-6 rule. Top-level `build_source` is
REQUIRED exactly when at least one active local build command of any admitted
local driver exists, and MUST otherwise be absent; marker v3's `go-v1`-only
wording does not survive into v4, and marker v3 itself is unchanged. External
entries continue to bind source per entry and MUST NOT use the consuming skill's
raw snapshot as external compiled-source identity.

Conformance claim schema 4 pins `protocol_version` to the rc.6 candidate and
admits driver assertions for the eight identifiers, each requiring `driver`,
`language`, `execution_policy`, and `operating_systems`. It admits only
`manager-worker-v1`, so a hardened claim remains structurally impossible and
needs a later claim version. Claim schemas 1 through 3 keep their bytes.

### 9. Mixed commands, platform claims, credentials, and signing

Manifest schema 8 MAY mix script commands, system commands, and build commands
of any admitted driver in one manifest and one closure node. Activation,
dependency-command selection, portable-name collision, shim collision, and
provider-first closure rules are unchanged, and active build command names are
still processed in Unicode-scalar lexical order within a closure node. Each
command independently derives its own artifact name, build input, cache key,
receipt, marker entry, and shim, whichever driver it names.

This decision makes no platform claim. Each of the six reserved identifiers
starts with an empty qualified-platform set. macOS and Windows remain the
platforms of the portable policy and Linux remains excluded until
`TASK-260728-1skseh`. A driver and platform tuple may enter a claim only when
its driver contract is accepted, its conformance vectors exist, and immutable
native evidence for that exact tuple exists; qualification verifies this. A
platform that cannot satisfy section 3's artifact class or section 6's toolchain
closure is excluded from the claim rather than shipped with a compensating
sidecar, host-resolved tool, or downgraded control.

Credentials, host-verification state, transport executables, proxy policy,
timeouts, and authentication modes stay operator-owned for every driver and MUST
NOT appear in a manifest, descriptor, repository, compiler environment, receipt
trust field, or marker.

No driver admitted by this boundary performs manager post-build signing,
timestamping, or notarization, and no package data may select a signing
identity, certificate, entitlement, or notarization credential. A platform
policy that requires a locally signed binary MUST reject the build until the
separately versioned and reviewed signer profile of `protocol/core.md` section
12.2 exists.

### 10. Downstream obligations

- `TASK-260728-1g0z69`, then `TASK-260728-2jaw7h`: the shared toolchain
  requirement, resolution, comparison, two-stage preflight, diagnostics, and
  guidance catalog, satisfying section 6 items 1 through 5 and adding no
  package-controlled installation data.
- `TASK-260728-12pnm1` (Rust), `TASK-260728-1yhuqi` (Swift),
  `TASK-260728-168smo` (Kotlin): one accepted contract per pair, each defining
  the local project-metadata file and `source_dir` mapping of section 4, the
  descriptor target semantics of section 5, the fingerprinted toolchain closure
  of section 6 item 3, the single graph and compile commands and the exhaustive
  pre-compile rejection matrix of section 7, the closed policy object and native
  target of section 8, and the per-platform proof that section 3's artifact
  class is met. `TASK-260728-168smo` additionally decides the Kotlin backend
  within, not around, section 3; if no Kotlin backend satisfies it, both Kotlin
  identifiers are retired unused.
- `TASK-260728-251p01`: integrate only the accepted contracts into manifest
  schema 8, descriptor schema 2, receipt schemas 3 and 4, marker schema 4, claim
  schema 4, the profiles, `SECURITY.md`, `COMPATIBILITY.md`, `CHANGELOG.md`, and
  the generated positive and negative corpus, keeping schemas 1 through 7 and
  every Go identity byte-stable.
- `TASK-260728-2bu2q6`: qualify the candidate, recompute every identity, and
  emit only evidence-backed driver and platform claims.
- `STORY-260728-327soo` continues to own the six deferred hardened guarantees.
  None of them may be named, claimed, or implied by any driver admitted here.

### 11. Enforcement while the reservation stands

`tools/validate.py` carries a deterministic boundary gate that runs on every
validation. It requires this decision to fix the closed identifier set, both
artifact classes, and the boundary failure classes; it forbids this decision
from naming a deferred hardened guarantee; it rejects any occurrence of a
reserved identifier on a surface file outside `decisions/`, because a decision
record is where an identifier is proposed, reserved, and retired while every
other surface is admission; it requires every driver-bearing schema definition
to close `driver` with a `const` over the admitted identifiers; it requires the
published artifact to stay a single closed file with no bundle member; it keeps
the reserved schema slots unallocated; and it proves against the compiled
validators that each frozen manifest, descriptor, receipt, marker, and claim
schema rejects each of the six reserved identifiers.

An integration task admits a driver by extending the admitted set in that gate
together with the schemas, never by weakening it.

## Stable failure classes

These architecture-level outcomes are interoperable semantic classes and MUST
remain distinguishable from each other and from a cache hit, an audit success, a
source unavailability, or a generic fallback:

- `build_descriptor_schema_unsupported`;
- `build_descriptor_driver_unsupported`;
- `build_artifact_class_unsupported`; and
- `build_package_code_execution_forbidden`.

The stable failure classes of decision 0005 and the `build_execution_*`
diagnostics of decision 0006 continue to apply unchanged to every driver.

## Rejected alternatives

- **One driver identifier per language covering both source modes.** Rejected:
  the local and external modes have different source identities, receipts, audit
  subjects, and lifecycle rules, exactly as `go-v1` and `go-repository-v1` do,
  and collapsing them would make the receipt version unreadable from the wire.
- **A generic `native-v1` driver parameterized by a language field, or language
  detection from project-metadata filenames.** Rejected: it is the generic
  fallback that decisions 0004 and 0005 and `protocol/core.md` section 12.3
  forbid, and detection would let repository layout choose a compiler.
- **One receipt schema version per driver, giving six new receipt schemas.**
  Rejected: it multiplies frozen artifacts without adding discrimination, since
  the driver `const` already discriminates inside one schema and the marker
  records the receipt version explicitly.
- **Reusing receipt schemas 1 and 2 for the new drivers.** Rejected: it would
  change frozen rc.4 and rc.5 schema bytes and let a reader that only knows Go
  believe it understands a Rust, Swift, or Kotlin receipt.
- **Admitting a `runtime-bundle` artifact class so Kotlin/JVM could ship.**
  Rejected for the four reasons in section 3. The honest consequence is that
  Kotlin is admitted only through a native backend or not at all, and that
  outcome is `TASK-260728-168smo`'s to establish, not to negotiate away.
- **Letting a driver publish a sidecar runtime or redistributable library
  alongside the executable on platforms that need one.** Rejected: it is the
  runtime-bundle class under a different name, and it would make the artifact
  digest describe only part of what runs.
- **Adding package-controlled members such as a binary, product, feature,
  profile, or configuration to the local command or the descriptor target.**
  Rejected: the consuming command key is the sole executable name, and any
  additional selector hands output and pipeline control back to untrusted data.
- **Minting `manager-worker-v2` for the additional drivers.** Rejected: the
  mandatory control set, worker session discipline, native-control inventory,
  and evidence record are identical, so a new identity would change every `go-v1`
  cache key and invalidate the frozen rc.5 candidate for no security gain.
- **Versioning the native-control inventory or the capability-evidence record
  alongside the new drivers.** Rejected: inventory membership is unchanged, and
  it never enters a build input, artifact, or hashed identity, so a rename would
  carry no semantic content.
- **Bumping `Skillfile.dev.json` to schema 3.** Rejected: substitution selects
  source acquisition only and cannot express a driver, toolchain, target, or
  artifact class.
- **Deferring the version and artifact boundary until the three driver contracts
  are written.** Rejected: each contract would then choose its own versions,
  artifact shape, and descriptor evolution, and the first one merged would define
  the wire by accident.
- **Allowing each driver to state its own execution-policy identity.** Rejected:
  policy identity is a security-strength statement, and per-driver identities
  would let one driver's weaker contract be read as the portable policy.

## Compatibility impact

This decision changes no bytes. It adds no schema, no vector, no generated case,
and no release metadata; it does not alter the rc.5 conformance manifest digest
or any pin. Manifest schemas 1 through 7, `skill-build.json` schema 1, receipt
schemas 1 and 2, marker schemas 1 through 3, claim schemas 1 through 3,
`Skillfile.dev.json` schema 2, and all rc.4 and rc.5 conformance bytes keep their
exact contents and meanings, and every `go-v1` and `go-repository-v1` identity is
unchanged.

When the reserved versions are minted, they are explicit reader and writer
version transitions. Readers never infer them from fields, MUST reject an
unsupported version rather than downgrade it, and MUST write the version the
active feature requires. Schemas 1 through 7 MUST reject the six reserved driver
identifiers and every schema-8-only field.

## Security impact

The security posture of the portable policy is unchanged, and no hardened
guarantee is added, implied, or claimed. The six deferred guarantees remain
deferred to `STORY-260728-327soo` and MUST NOT appear in a mandatory-control set,
the native-control inventory, or a capability-evidence record.

The exposure that does change is compiler input. Admitting three more compiler
front ends under the same portable, non-hardened controls widens the untrusted
parsing, code-generation, and resource-consumption surface, and each of the three
languages ships a mainstream build path whose normal operation executes
package-selected code. That is why section 7 requires a deterministic
pre-compile rejection matrix rather than a runtime allowance, why section 6 item
3 requires every started executable to be inside a fingerprinted closure, and why
a surface that cannot be rejected before compilation disqualifies its driver.
Each driver contract MUST state its own compiler-input denial-of-service and
vulnerability exposure honestly and MUST NOT rely on containment this protocol
does not yet provide.

Section 3 also closes a currentness gap rather than opening one: a single
self-contained executable is fully described by one digest the manager computed,
whereas a bundle plus an external runtime could be reported current while the
runtime it needs is gone. Artifact execution remains a later user action, and the
compiled result remains untrusted package code.
