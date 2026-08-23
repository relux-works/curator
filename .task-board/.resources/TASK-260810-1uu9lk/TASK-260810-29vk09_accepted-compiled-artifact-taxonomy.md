# Compiled artifact taxonomy and fail-closed deny policy

Status: research decision revised for review under `TASK-260810-29vk09`.

## Context

Curator needs one artifact-admission rule for every language-aware skill CLI
adapter. The repository delivery input requires each supported adapter to
capture a complete, immutable, recursively auditable source closure or fail
before build or installation. It also forbids vendored precompiled executable
code until a separate binary-admission capability exists
([local source-closure specification](../.spec/skill-facing-cli-source-closure.md)).

This document defines the common taxonomy, trust boundaries, deterministic
decision procedure, recursive-container rules, diagnostics, audit evidence,
future verified-binary seam, and conformance vectors. It is intentionally an
adapter-independent policy. Rust, Node/TypeScript, SwiftPM/C-family, Go, and any
future adapter may narrow their supported source profiles, but none may admit a
class that this policy rejects.

Revision note: reviewer finding R1 is addressed by replacing the former blanket
`ET_DYN` -> dynamic-library mapping with an evidence-based ELF resolution table,
an explicit compiled ambiguity class, raw resolution observations, and shared
dynamic-PIE, static-PIE, shared-object, renamed, suffixless, and ambiguous
conformance fixtures. The dependency-input rejection is unchanged.

## Executive decision

1. Classify an artifact from its bytes, validated structure, filesystem node,
   container ancestry, and declared use. A suffix, MIME label, package tag,
   executable bit, or checksum is evidence but is never sufficient to allow it.
2. Apply the artifact's **trust role** before applying its class. Compiled bytes
   in a dependency payload are rejected; the same class may be used only when
   Curator selected it as an external trusted-toolchain component or proved it
   was produced locally from the admitted closure in the protected build
   pipeline.
3. Recursively inspect every recognized container before any package-manager
   hook, generator, installer, linker, loader, interpreter, or compiler can
   consume it. An archive is admitted only when every reachable member is
   admitted and its container metadata is safe.
4. Make **deny dominate**: if any sound detector recognizes a compiled or
   executable intermediate class, or if competing interpretations cannot be
   resolved, reject. Unknown, encrypted, malformed, unsupported, partially
   inspected, or resource-limit-exceeding content is rejected.
5. Emit one stable cross-adapter diagnostic vocabulary and a canonical manifest
   that binds the policy, limits, detector identities, immutable origin,
   container chain, leaf hashes, classification, and decision.
6. Reserve verified precompiled dependencies for a separate
   `verified-binary-v1` capability. A valid signature alone is not admission;
   the future verifier must bind exact bytes, provenance, signer/builder policy,
   target compatibility, and its own audit receipt.

The important security consequence is that package labels cannot create trust.
A SwiftPM `systemLibrary` declaration, npm `os`/`cpu` metadata, a wheel
`*-none-any.whl` name, a JAR signature, or a dependency path named `toolchain/`
cannot move bytes out of the untrusted dependency role.

## Terms and invariants

### Artifact

An artifact is a regular-file byte sequence, filesystem node, directory/bundle,
archive member, or compressed stream that can enter or affect the captured
closure or build. Every artifact has a canonical virtual path, exact byte size
and SHA-256 when it has bytes, origin identity, trust role, and zero or more
detector observations.

### Source-like text

Source-like text is a complete byte sequence accepted by a versioned parser or
lexer for a declared source, script, manifest, configuration, interface, or
documentation grammar. Merely being printable or UTF-8 is insufficient.
Interpreted scripts are source, even when they have a shebang or executable
mode. Minified or transpiled JavaScript and generated Swift interfaces may be
source-like text when their exact shipped bytes and immutable origin are the
declared build/runtime input.

If Curator runs a generator, the generator, all inputs, invocation, environment,
and outputs must instead be declared in the protected build graph. A generated
file already present in an immutable package is treated as a captured input;
its unavailable upstream generator is a reproducibility caveat for the
ecosystem policy, not permission to hide an opaque binary.

### Compiled or executable intermediate

This is a byte representation intended for a native loader, linker, virtual
machine, runtime bytecode loader, or compiler backend without rebuilding its
implementation from the admitted source boundary. It includes relocatable
objects and serialized compiler modules even when they are not independently
launchable.

The ELF ABI, for example, distinguishes relocatable, executable, and shared
object types and describes all three as binary program representations
([ELF object files](https://refspecs.linuxfoundation.org/elf/gabi4%2B/ch4.intro.html),
[ELF `e_type`](https://refspecs.linuxfoundation.org/elf/gabi4%2B/ch4.eheader.html)).
The historical `ET_DYN` label is not a sufficient use classification: GNU
binutils now describes it as either a position-independent executable or a
shared object
([binutils `ET_DYN` clarification](https://gcc.gnu.org/pipermail/binutils/2020-May/111252.html)).
Microsoft likewise defines PE images, COFF objects, and archive libraries as
executable/linkable structures
([PE/COFF specification](https://learn.microsoft.com/en-us/windows/win32/debug/pe-format)).

### Admission invariant

For every dependency payload `P`:

```text
admit(P) = origin_verified(P)
        AND inspection_complete(P, policy_version, limits)
        AND every reachable node/member has decision ADMIT_INPUT
```

There is no warning-and-continue result. `UNKNOWN`, detector failure, incomplete
inspection, or an unavailable required parser becomes `REJECT`.

## Trust roles and non-bypassable boundaries

The caller assigns a role from provenance and pipeline state; artifact metadata
cannot self-assert one.

| Trust role | How it is established | Compiled-class decision | Required evidence |
| --- | --- | --- | --- |
| `dependency_input` | Bytes arrived through a root package, transitive dependency, vendored tree, package cache import, or archive member before the protected build | **Reject** everywhere, across all adapters | Immutable package origin, raw payload digest, recursive manifest, policy verdict |
| `external_toolchain` | Curator policy selects a manager/OS-controlled root outside the dependency closure before the build | Allowed only as toolchain, never copied into or relabeled as source | Policy selector, resolved root/paths, version, platform, complete toolchain fingerprint, mutation/containment checks |
| `local_build_output` | A write is causally produced after admission in a clean, protected staging namespace by a declared build step | Allowed as intermediate or output; never retroactively becomes an input | Source-closure digest, complete build input, toolchain identity, command/policy/target, output path/size/hash/class, protected publication receipt |
| `verified_binary_candidate` | A future manager-owned capability receives exact candidate bytes through an explicit binary edge | **Reject today**; future verifier may return a policy-scoped allow receipt | See `verified-binary-v1` seam below |

Boundary rules:

- A package-provided compiler, SDK, linker, runtime, binary generator, `bin/`
  directory, SwiftPM system-library declaration, or `PATH` entry remains
  `dependency_input` and is rejected if compiled.
- Arbitrary host libraries are not trusted toolchains. Only components within a
  Curator-selected and fingerprinted toolchain/SDK/sysroot boundary qualify.
- The output namespace starts empty and is disjoint from the read-only captured
  source snapshot. Moving, copying, or hard-linking a pre-existing dependency
  binary into it does not establish local production.
- Declared hooks may compile admitted source only when the adapter's build
  policy permits the hook and its inputs, outputs, environment, network state,
  toolchains, and ordering are explicit. The artifact rule does not authorize
  arbitrary hooks.
- Intermediate objects may exist during a protected local build. They are not
  reusable across runs unless covered by the same protected receipt rules as a
  published output.

This distinction matches the existing Go baseline: its logical build input
includes source, target, toolchain, and closed execution policy
(`internal/buildmeta/models.go:61-121`), while protected-cache reuse requires an
exact protected hit and validates receipt input plus artifact size and digest
(`internal/buildcache/cache.go:28-56, 178-214`). The current receipt hash is
explicitly a consistency identifier, not a signature or provenance proof
(`internal/buildmeta/models.go:118-121`, `internal/buildmeta/codec.go:162-164`).

## Shared artifact taxonomy and decisions

Decision vocabulary:

- `ADMIT_INPUT`: eligible to enter the source closure after origin and graph
  checks.
- `DESCEND`: a structural container; its final result is the conjunction of all
  member results.
- `REJECT`: not admitted and build/install must not begin.
- `ALLOW_TOOLCHAIN`: usable only under the independently established
  `external_toolchain` role.
- `ALLOW_OUTPUT`: usable only under the causally established
  `local_build_output` role and valid protected receipt.

| Stable class | Examples and recognition | `dependency_input` | Toolchain/local-output handling |
| --- | --- | --- | --- |
| `source.authored_text` | Declared Go, Rust, Swift, C-family, JS/TS, Python, shell, assembly, headers, module maps, or another supported grammar; full parse/lex succeeds | `ADMIT_INPUT` when origin and declaration are bound | May be a toolchain resource or generated output, but remains hashed |
| `source.generated_text` | Transpiled/bundled/minified JS, generated `.swiftinterface`, generated C/Rust/Go source, source maps, generated manifests | `ADMIT_INPUT` when the exact bytes are the declared immutable input; adapters may require stronger generator evidence | Locally generated form requires declared generator lineage |
| `text.metadata` | Lockfiles, manifests, licenses, README/docs, checksums, SBOM text, declarative configuration accepted by a closed grammar | `ADMIT_INPUT`; a manifest may not override a byte-level deny | Hash and parser identity are recorded |
| `data.known_inert` | A non-code binary format supported by a manager-owned, versioned allowlist and strict parser | Default policy-v1 allowlist is empty, so `REJECT`; adding a format requires a central policy revision, never an adapter exception | Toolchain resource or receipted output only under its role |
| `native.executable` | ELF `ET_EXEC` or an ELF `ET_DYN` resolved as a PIE/executable by the rule below, Mach-O executable, PE image, fat/universal slice containing one, or another validated native image; suffixes such as `.exe` are hints | `REJECT` | `ALLOW_TOOLCHAIN` or `ALLOW_OUTPUT` only |
| `native.object` | ELF `ET_REL`, Mach-O object, COFF object, LLVM native-object wrapper, Go `.syso`, or equivalent relocatable input | `REJECT` | `ALLOW_TOOLCHAIN` or `ALLOW_OUTPUT` only |
| `native.library.static` | `ar`/COFF archive libraries, `.a`, `.lib`, Rust `.rlib`, or a container of native objects intended for linking | `REJECT` | `ALLOW_TOOLCHAIN` or `ALLOW_OUTPUT` only |
| `native.library.dynamic` | ELF `ET_DYN` resolved as a shared object by the rule below, Mach-O `.dylib`/bundle, PE DLL/import library, `.so`, or equivalent loadable/linkable image | `REJECT` | `ALLOW_TOOLCHAIN` or `ALLOW_OUTPUT` only |
| `native.elf.et_dyn_ambiguous` | Structurally valid ELF `ET_DYN` whose PIE/executable versus shared-object use lacks decisive evidence or has conflicting evidence | `REJECT` with `artifact_compiled_dependency_forbidden` as the primary code | `ALLOW_TOOLCHAIN` or `ALLOW_OUTPUT` only under the independently established role; retain the ambiguous class and raw ELF facts |
| `apple.framework` | A `.framework` bundle by portable path shape or bundle metadata, including resource-only variants | `REJECT` as a whole; do not infer safety from a missing expected binary | Trusted SDK component or local output only |
| `apple.xcframework` | `.xcframework` bundle, directory, or archive by path shape/metadata | `REJECT` as a whole | Trusted SDK component or local output only |
| `native.extension.node` | `.node` addon or a native image referenced as a Node addon | `REJECT` | Local source build may emit one as a protected output |
| `native.extension.python` | CPython/shared extension such as `.so` or `.pyd`, including ABI-tagged names | `REJECT` | Local source build may emit one as a protected output |
| `vm.jvm_bytecode` | Valid `.class` (`CAFEBABE` plus structural validation), DEX, or equivalent JVM/Android executable bytecode | `REJECT` | Trusted toolchain component or protected local output only |
| `vm.python_bytecode` | `.pyc` recognized by a supported interpreter header or any dependency file claiming a `.pyc` role | `REJECT` | Protected local cache/output only; never a source input |
| `vm.javascript_code_cache` | V8/Node cached-data, startup snapshot, bytecode blob, or equivalent engine-specific serialized executable state | `REJECT`; an unrecognized variant is also opaque and rejected | Protected runtime/toolchain state or local output only |
| `ir.webassembly` | Valid WebAssembly module/component binary or a file claiming `.wasm`; binary preamble begins `\0asm` for core modules | `REJECT` | Protected local output only unless future binary verification allows it |
| `ir.compiler_serialized` | LLVM bitcode, Swift `.swiftmodule`/`.swiftdoc`, Clang PCH/PCM, C++ BMI, Rust `.rmeta`, or equivalent serialized AST/IR/module cache | `REJECT` | Trusted toolchain resource or protected local output only |
| `container.archive` | ZIP/ZIP64, tar, gzip-wrapped tar, package tarball, wheel, JAR/WAR/AAR, or another supported archive grammar | `DESCEND`; admit the container only if every member admits and metadata is safe | Toolchain/output role still records the container digest |
| `container.compressed_stream` | Supported gzip stream or another explicitly registered compression envelope | `DESCEND` into exactly one bounded byte stream | Same role-specific handling as its decoded child |
| `fs.directory` | Real directory or synthetic archive directory | `DESCEND` in canonical path order | Included in manifest structure |
| `fs.symlink_or_hardlink` | Filesystem or archive link entry | `REJECT` in policy v1, including an apparently in-root target | A fingerprinted external toolchain may use independently validated contained links; output publication must not |
| `fs.special` | Device, FIFO, socket, mount/reparse point, sparse mapping needing external data, or unknown node kind | `REJECT` | Reject unless a future toolchain policy explicitly models it; never publish as output |
| `opaque.unknown` | Bytes, encoding, grammar, parser version, or compound/polyglot interpretation not fully understood by the closed registry | `REJECT` | Not usable as toolchain/output without the corresponding role validator |

Apple describes Mach-O as its native executable format and its `__text` section
as compiled machine code
([Mach-O overview](https://developer.apple.com/library/archive/documentation/Performance/Conceptual/CodeFootprint/Articles/MachOOverview.html)).
Apple's current XCFramework documentation calls an XCFramework a binary package
that contains static or dynamic frameworks/libraries
([XCFramework bundle](https://developer.apple.com/documentation/Xcode/creating-a-multi-platform-binary-framework-bundle?language=objc));
framework documentation explains that framework executable code is a dynamic
shared library
([framework binding](https://developer.apple.com/library/archive/documentation/MacOSX/Conceptual/BPFrameworks/Concepts/FrameworkBinding.html)).

Node documents C++ addons as dynamically linked shared objects
([Node addons](https://nodejs.org/api/addons.html)). Python documents a C
extension as a shared library such as `.so` or `.pyd`
([Python extension modules](https://docs.python.org/3.12/extending/building.html))
and `.pyc` as generated bytecode
([Python `py_compile`](https://docs.python.org/3/library/py_compile.html)).

The JVM specification defines a class file as the binary representation of JVM
instructions and fixes its magic value to `0xCAFEBABE`
([JVM class format](https://docs.oracle.com/javase/specs/jvms/se24/html/jvms-4.html)).
The WebAssembly binary specification defines its module preamble and binary
grammar
([WebAssembly binary modules](https://webassembly.github.io/spec/core/binary/modules.html)).
LLVM bitcode is serialized LLVM IR and can also be wrapped inside ELF, COFF, or
Mach-O objects
([LLVM bitcode format](https://llvm.org/docs/BitCodeFormat.html)). Swift
describes `swiftmodule` as an opaque, compiler-version-tied archive and contrasts
it with a textual module interface
([Swift module stability](https://www.swift.org/blog/abi-stability-and-more/)).
These facts justify classifying non-launchable bytecode, objects, and serialized
IR as deny classes rather than looking only for stand-alone executables.

### Ambiguous cases and conservative outcomes

| Case | Decision and reason |
| --- | --- |
| Executable-mode shell/Python/JS text with a valid declared grammar | Admit as source-like text; mode/shebang does not make it precompiled. Build/hook policy still controls execution. |
| Minified/bundled JS without its original TypeScript | Eligible as `source.generated_text` only when the exact JS is the immutable declared runtime input. Do not claim source-level reproducibility that the evidence does not provide. |
| `.swiftinterface` text | Eligible generated interface text. It cannot make a missing/rejected compiled Swift library admissible. Binary `.swiftmodule` remains rejected. |
| Textual assembly, WebAssembly text, or LLVM textual IR | Eligible only when a central source profile recognizes and fully parses the grammar as a declared source input consumed by a trusted toolchain. Otherwise reject as opaque. Their compiled/binary forms remain deny classes. |
| Text file named `.so`, `.dll`, `.a`, `.class`, `.wasm`, `.pyc`, or `.node` | Reject as `artifact_type_ambiguous`; a deny-indicating name/use and benign byte interpretation conflict. This intentionally rejects linker scripts masquerading as `.so` until a central explicit profile exists. |
| ELF/PE/Mach-O/JVM/Wasm bytes named `.txt` or with no suffix | Reject the content-derived compiled class. |
| Valid ELF `ET_DYN` without decisive PIE/executable or shared-object evidence | Classify `native.elf.et_dyn_ambiguous`; reject with `artifact_compiled_dependency_forbidden`, not the generic ambiguity code, because every remaining interpretation is compiled executable/linkable code. |
| Polyglot matching source and archive/native interpretations | Reject as ambiguous unless a compiled detector matches, in which case `artifact_compiled_dependency_forbidden` dominates. |
| JAR/WAR/AAR | Treat as an archive, not as proof of bytecode. Any `.class`, DEX, native library, nested compiled member, unsafe entry, or opaque member rejects. A genuinely source-only JAR is eligible only under a central source-container profile. Oracle confirms JAR is ZIP-based ([JAR specification](https://docs.oracle.com/en/java/javase/26/docs/specs/jar/jar.html)). |
| Python wheel | Treat as a ZIP container; filename compatibility tags and `Root-Is-Purelib` are hints. Inspect all members and verify wheel `RECORD`; any native extension/compiled/opaque member rejects. The wheel specification defines both ZIP structure and per-file hashes ([wheel format](https://packaging.python.org/en/latest/specifications/binary-distribution-format/)). |
| Python sdist, npm `.tgz`, Cargo `.crate`, source ZIP | Treat as nested compression/archive containers and inspect all members. “Source” in the package format is not an admission shortcut. Python's sdist spec, for example, defines a `.tar.gz` container and extraction constraints ([sdist format](https://packaging.python.org/en/latest/specifications/source-distribution-format/)). |
| Resource-only `.framework` | Reject. The bundle class is globally denied, and Apple notes frameworks can contain resource varieties or even be resource-only; allowing a special case would create an adapter/bundle-metadata bypass. |
| PNG/PDF/font/database/other binary data | Policy v1 rejects because the manager-owned inert-data registry is empty. A later central revision may add strictly parsed non-code formats; PDF scripts, font bytecode, embedded objects, and unknown chunks require explicit decisions first. |
| Base64/encrypted/steganographic executable represented inside source text | The text may classify as source because arbitrary program semantics are outside this scanner. Offline build isolation, declared hooks/I/O, and runtime policy must contain reconstruction. Encrypted or encoded *container members* that the container declares are rejected. |
| Signed binary in a dependency | Reject today. Signature presence does not change the trust role or invoke an absent capability. |

### ELF `ET_DYN` class resolution

The detector must retain the raw ELF type and then resolve use separately.
`ET_DYN` alone cannot select `native.executable` or
`native.library.dynamic`: GNU toolchains emit both PIE executables and shared
objects with this type. GCC documents `-pie` as a dynamically linked
position-independent executable, `-static-pie` as a position-independent
executable that does not need a dynamic linker, and `-shared` as a shared
object
([GCC link options](https://gcc.gnu.org/onlinedocs/gcc-13.3.0/gcc/Link-Options.html)).
GNU binutils identifies an `ET_DYN` PIE by parsing `DT_FLAGS_1` and testing
`DF_1_PIE`, rather than by trusting `e_type` or the suffix
([binutils `readelf` PIE detection](https://sourceware.org/pipermail/binutils/2021-June/116921.html)).

Other observations are useful but not individually conclusive. The generic ELF
ABI says `PT_INTERP` is meaningful for executable files but may occur in shared
objects. It defines `DT_SONAME` as ignored for executables and optional for
shared objects, so its presence is library evidence but its absence is not
executable evidence
([ELF program headers](https://gabi.xinuos.com/elf/07-pheader.html),
[ELF dynamic tags](https://gabi.xinuos.com/elf/08-dynamic.html)). `e_entry` is
always recorded but is not decisive because a linker entry point can be set
explicitly. Names, modes, and suffixes remain hints only.

Before this resolution, `elf-v1` must structurally validate the ELF header,
program-header table, load ranges, any `PT_INTERP`, the `PT_DYNAMIC` range, and
the complete dynamic table. It records these normalized observations:

- raw `e_type`, `e_entry`, `e_machine`, class, data encoding, OS ABI, ABI
  version, and `e_flags`;
- the ordered program-header types and ranges, including the count and exact
  bounded value of `PT_INTERP` and presence of `PT_DYNAMIC`;
- raw `DT_FLAGS_1`, whether `DF_1_PIE` is set, `DT_SONAME`, executable-only
  dynamic tags when present, and unexpected/duplicate tags;
- manager-resolved use edges: `execute`, `link_or_load`, both, or none, with
  the manifest/build-graph field that established each edge. A package label
  or filename by itself is not a resolved use edge.

Apply this table in order after structural validation:

| Priority | Valid `ET_DYN` evidence | Stable class and variant |
| ---: | --- | --- |
| 1 | `DT_FLAGS_1` has `DF_1_PIE` | `native.executable`; variant `elf.pie.interpreter` when `PT_INTERP` exists, otherwise `elf.pie.no_interpreter`. A contrary link/load edge is recorded as misuse but cannot relabel an explicit PIE as a library. |
| 2 | No `DF_1_PIE`; one manager-resolved `execute` edge plus executable-use structural support such as `PT_INTERP`; no `DT_SONAME` or link/load edge | `native.executable`; variant `elf.et_dyn.executable_by_use` |
| 3 | No `DF_1_PIE`, execute edge, `PT_INTERP`, or other executable-use structural fact; `DT_SONAME` or one manager-resolved `link_or_load` edge | `native.library.dynamic`; variant `elf.shared_object` |
| 4 | No rule above, or the non-decisive structural/use facts conflict | `native.elf.et_dyn_ambiguous`; reason enum such as `insufficient_evidence`, `interp_soname_conflict`, or `use_conflict` |

For `dependency_input`, **every row returns `REJECT` with
`artifact_compiled_dependency_forbidden` as the primary diagnostic**. The
ambiguous row may carry secondary classification detail, but it must not switch
the primary result to `artifact_type_ambiguous` and must never admit. Dynamic
PIE, no-interpreter/static PIE, shared-object, renamed, and suffixless forms
therefore differ only in honest class/variant evidence, not in security outcome.

## Deterministic classification and admission algorithm

The shared service, not an adapter, owns this ordered procedure.

1. **Freeze identity.** Read from the captured immutable snapshot, not a live
   mutable package cache. Record origin locator, lock/checksum identity, raw
   size, SHA-256, and role before parsing.
2. **Validate the node and path.** Reject unsupported nodes and paths before
   content parsing. Do not follow dependency links.
3. **Recognize bundle/container context.** Detect framework/XCFramework roots,
   archive/compression grammars, and package containers from path plus validated
   structure. A renamed archive is still a container.
4. **Run the closed detector set.** Each versioned detector either returns
   `NO_MATCH`, a structurally validated class with evidence, or `ERROR`.
   `ERROR` rejects. Magic bytes are a candidate signal, not proof; LLVM's own
   bitcode documentation warns tools not to decide validity from magic alone.
5. **Resolve observations.** Apply format-specific rules, including the ordered
   ELF `ET_DYN` table above, before selecting the stable class. A compiled-class
   match wins. Multiple incompatible non-compiled interpretations reject as
   ambiguous. A deny-indicating suffix, manifest role, or load/link reference
   with otherwise benign bytes also rejects as ambiguous. No detector may turn
   another detector's deny into an allow.
6. **Descend containers.** Enumerate every logical entry, validate metadata,
   sort by canonical virtual path, and stream each member through this same
   procedure. Nested archives are detected by bytes even when their names lack
   an archive suffix.
7. **Apply source/data profile.** A source/text allow requires full grammar
   consumption under the declared encoding. The manager-owned profile is part
   of the policy identity. Adapters may remove allowed grammars but may not add
   compiled or opaque classes.
8. **Combine.** One rejected member rejects every ancestor and package. An
   archive admits only if all entries admit, all declared package hashes agree,
   and the entire traversal stayed within limits.
9. **Commit evidence.** Canonically encode and hash the manifest. Build planning
   consumes the admitted manifest digest, not a rescanned mutable pathname.

All entries are inspected to collect bounded audit evidence when safe to do so.
The primary diagnostic is the first event after sorting by canonical virtual
path, then rule priority, then class identifier. Structural inability to
continue at an ancestor is reported at that ancestor. This makes error choice
independent of archive member order, host directory enumeration, and adapter.

## Recursive archive policy

### Supported container registry

Policy v1 must at minimum understand:

- ZIP and ZIP64, including wheel and JAR conventions;
- POSIX ustar/pax and GNU tar members;
- gzip as a bounded single-stream envelope, including `.tar.gz`, npm `.tgz`,
  and source/package tarballs;
- native archive/library formats needed to identify and reject `.a`, `.lib`,
  `.rlib`, and import libraries.

An unavailable decoder, unsupported compression/encryption method, split or
multi-volume archive, malformed/truncated index, checksum mismatch, or trailing
data that creates a second interpretation rejects. Adding bzip2, xz, zstd, 7z,
DMG, package installers, or another format requires a detector-registry and
policy-version change with fixtures. Until then it is
`artifact_archive_unsupported`, not an adapter-specific best effort.

### Safe virtual paths and entries

Before extraction or hashing an entry:

- require valid UTF-8 in normalized NFC form with `/` as the only separator;
- reject NUL/control characters, backslashes, absolute/UNC/drive paths, empty,
  `.` or `..` components, and Windows device/alternate-stream spellings;
- cap the path at 4096 UTF-8 bytes and each component at 255 bytes;
- reject exact duplicates and collisions under the portable comparison key
  (NFC, Unicode simple case-fold, Windows trailing-dot/space folding);
- accept only regular files and directories; reject symlinks, hard links,
  devices, FIFOs, sockets, sparse/external extents, reparse/mount entries, and
  unknown types;
- ignore ownership, timestamps, ACLs, xattrs, setuid/setgid/sticky bits, and
  other host metadata for identity, while recording their presence and
  rejecting any feature that changes content resolution or execution.

This is intentionally stricter than general extraction libraries. Python's
official `tarfile` guidance identifies absolute/out-of-tree paths, links,
special files, duplicate names, case-fold shadowing, file counts, expanded
sizes, and denial-of-service limits as concerns that require verification even
with extraction filters
([Python `tarfile` security guidance](https://docs.python.org/3/library/tarfile.html)).
Go's ZIP reader likewise exposes insecure non-local paths as a distinct error
([Go `archive/zip`](https://pkg.go.dev/archive/zip)).

Inspection uses a streaming, non-executable reader rooted in a new private
temporary namespace. It never extracts over an existing tree, follows a link,
mounts an image, invokes an archive's helper, or trusts last-entry-wins
semantics. The normalized manifest, not the extracted filesystem's accidental
result, is authoritative.

### Closed resource limits

The proposed immutable defaults for `curator-artifact-policy-v1` are:

| Limit name | Value | Accounting rule |
| --- | ---: | --- |
| `max_raw_payload_bytes` | 512 MiB | Raw bytes across one package payload before expansion |
| `max_single_leaf_bytes` | 256 MiB | Bytes emitted for any regular leaf |
| `max_total_emitted_bytes` | 2 GiB | Every byte emitted by every decompressor, charged again at each nested layer |
| `max_archive_depth` | 8 | Root container is depth 1; compression envelopes count |
| `max_container_count` | 1,024 | All archive/compression nodes in one payload |
| `max_entry_count` | 100,000 | Files plus explicit/synthetic directories across all containers |
| `max_expansion_ratio` | 200:1 | Per stream and aggregate emitted/raw bytes; nonempty output from zero declared compressed bytes rejects |
| `max_path_bytes` | 4,096 | Canonical UTF-8 virtual path including container separators |
| `max_component_bytes` | 255 | Each canonical path component |
| `max_recorded_findings` | 1,000 | Canonical evidence retains total count and digest even if display details truncate |

All arithmetic is checked for overflow before addition. Declared sizes are used
for early refusal, but actual streamed counts are authoritative and may only
tighten the result. Exceeding any limit returns
`artifact_inspection_limit_exceeded` with `limit_name`, `limit`, and `observed`;
it never admits a partially inspected package. The exact limit vector is part
of the policy ID and audit checkpoint, so an adapter cannot silently raise it.

An operational deadline or I/O cancellation may stop work earlier, but it
returns `artifact_inspection_unavailable` and rejects. It is not represented as
a canonical content classification and cannot produce a reusable admission
receipt.

### Detection limits and compensating controls

The scanner proves only the closed statement that every reachable artifact
boundary it understands was classified under this policy. It does **not** prove
that admitted source is benign, that a parser/runtime has no vulnerability, or
that code cannot calculate/download/JIT/decrypt new executable bytes.

Specifically, it does not search arbitrary byte offsets for every possible
embedded format, interpret base64/source literals, execute macros, solve build
scripts, mount filesystem images, or follow runtime/network data. Recognized
containers and native wrappers are recursively parsed; everything else that is
not fully recognized source/text is rejected as opaque. The compensating
controls are complete declared build graphs, a read-only captured source root,
an empty disjoint output root, no network, manager-selected toolchains,
restricted hooks/plugins/generators, write-set verification, and protected
output receipts.

## Stable diagnostics

Diagnostic codes are lowercase snake case, global to all adapters, and stable
within the policy major version. Operator detail may improve without changing
the code. This follows the existing Go driver's convention that callers branch
on a stable code and humans read detail (`internal/godriver/errors.go:8-30`).

| Code | Required meaning and fields |
| --- | --- |
| `artifact_origin_unverified` | Immutable origin/checksum/lock binding missing or mismatched; include package identity and observed digest |
| `artifact_compiled_dependency_forbidden` | A deny-class compiled/executable/intermediate artifact was found under `dependency_input`; include class, variant/resolution reason, virtual path, SHA-256, size, detector, raw format facts (including ELF `e_type` and `ET_DYN` discriminators), and container chain |
| `artifact_binary_admission_unavailable` | A caller requested binary admission but `verified-binary-v1` is absent or disabled; never downgrade to the generic source path |
| `artifact_type_ambiguous` | Conflicting valid interpretations or deny-indicating name/use with nonmatching bytes; include all observations |
| `artifact_opaque_dependency_forbidden` | No complete allowed interpretation exists; include attempted detectors and non-sensitive failure summaries |
| `artifact_archive_invalid` | A recognized container is malformed, truncated, has inconsistent indexes/sizes/checksums, or has disallowed trailing data |
| `artifact_archive_unsupported` | Recognized archive/compression/encryption/multi-volume feature has no policy-v1 decoder |
| `artifact_archive_encrypted` | An encrypted member/stream prevents complete pre-execution inspection |
| `artifact_archive_unsafe_path` | Absolute, traversal, nonportable, duplicate, or colliding virtual path; include canonical collision key where safe |
| `artifact_archive_unsafe_entry` | Link, special node, sparse/external extent, reparse/mount, or unsupported member kind |
| `artifact_inspection_limit_exceeded` | A closed count/size/depth/ratio/path limit was crossed; include limit vector identity and exact field |
| `artifact_inspection_unavailable` | Read, parser, cancellation, or operational failure prevented a complete decision; no admission receipt may be cached |
| `artifact_generated_input_undeclared` | Build-time generation or a generated input lacks declared generator/input/output lineage required by its adapter profile |
| `artifact_toolchain_untrusted` | A purported tool is package-provided, outside a selected root, mutable, unapproved, incomplete, or otherwise lacks toolchain role evidence |
| `artifact_toolchain_identity_changed` | Toolchain bytes/path/version changed between checkpoint and use |
| `artifact_local_output_unreceipted` | Compiled/intermediate bytes in the output namespace lack causal build and protected-publication evidence |
| `artifact_local_output_drift` | Protected artifact path, size, digest, class, or complete receipt input differs from expectation |
| `artifact_policy_internal_error` | Closed fail-safe for an invariant violation; always reject and never cache |

Example human rendering:

```text
curator artifact_compiled_dependency_forbidden:
dependency @scope/pkg@1.2.3 member package/prebuilds/linux-x64/addon.node
classified native.extension.node by elf-v1 (sha256:..., 48211 bytes)
inside package.tgz!package/prebuilds/linux-x64/addon.node
```

The machine record must not require consumers to parse that sentence. It carries
the structured fields below. Attacker-controlled names are escaped, display
length is bounded, and host-absolute temporary paths are excluded.

## Required audit and checkpoint evidence

### Dependency artifact manifest (`artifact-manifest-v1`)

Canonical, schema-checked evidence contains:

- schema and policy ID/version; complete limit vector;
- detector registry ID and each detector/parser implementation identity;
- adapter/profile ID, package manager, package name/version, immutable origin,
  lock/checksum record, raw payload SHA-256 and size;
- trust role assigned by Curator;
- every node in canonical virtual-path order: node kind, original encoded name,
  canonical path/collision key, byte size/hash, container parent/chain,
  declared mode/use, detector observations, selected class, decision and rule;
- for every ELF node, the raw header facts and validated program/dynamic-table
  observations listed in the `ET_DYN` section; for `ET_DYN`, also the ordered
  resolution rule, selected variant or ambiguity reason, and provenance of each
  manager-resolved use edge;
- archive declared versus observed sizes/checksums, compression/encryption
  method, and accumulated limits;
- all diagnostics in deterministic order; when display findings are capped,
  retain the total count plus a digest over the complete canonical finding set;
- final package decision and a SHA-256 digest over the canonical manifest.

An allow receipt is valid only for the exact raw payload and exact policy,
profile, detector registry, origin, and limit vector. Timestamps and absolute
host paths may appear in operational logs but are excluded from portable cache
identity.

### External toolchain checkpoint

Record the manager policy selector, resolved root and executable relative path,
version output, platform/architecture/ABI/SDK identity, complete content
fingerprint, contained-link validation, environment search resolution, and the
time-of-use recheck. Package manifests cannot add trusted roots. The existing Go
fingerprinter already uses a domain-separated canonical tree digest and rejects
escaping links and special files (`internal/godriver/fingerprint.go:19-72,
166-222`); the shared contract should preserve those properties.

### Local build-output receipt

Record at least:

- source-closure and artifact-manifest digests;
- dependency graph/build-plan digest and declared write set;
- adapter/driver and execution-policy identity;
- complete toolchain/SDK/runtime identity;
- target OS, architecture, ABI, tuning, minimum deployment/runtime versions;
- canonical commands, relevant environment, network=`none`, sandbox policy,
  generator/hook decisions, and build order;
- clean staging-root identity and observed produced paths;
- every published output's canonical relative path, stable class, SHA-256, and
  size;
- canonical receipt hash and protected-cache publication/validation result.

Reuse requires an independently derived expected input and exact protected hit.
A content hash or self-consistent receipt outside the manager-protected boundary
does not prove provenance. This is already the posture of the Go cache and must
remain true for multi-output and mixed-language adapters.

## Future `verified-binary-v1` capability seam

Verified binaries must never be represented as source or enabled by an adapter
flag. Add a separate graph node/edge and a manager-owned interface conceptually
like:

```go
type BinaryAdmissionVerifier interface {
    Verify(context.Context, BinaryCandidate, BinaryPolicy) (VerifiedBinaryReceipt, error)
}
```

Current construction supplies no implementation and the central classifier
returns `artifact_binary_admission_unavailable` for every compiled
`dependency_input`. A future release may construct the verifier only when an
explicit Curator capability and administrator policy both enable it.

`BinaryCandidate` must bind the exact immutable bytes (SHA-256 and size),
canonical format/class, source package and origin, intended graph edge, target
OS/architecture/ABI/minimum runtime, and the raw signature/attestation bundle.
The verifier must evaluate, rather than merely record:

1. format validity and target compatibility for every slice/member;
2. subject/artifact digest and size equality;
3. signature authenticity against versioned trusted roots and allowed signer
   identities, including threshold, timestamp/transparency, expiry and
   revocation policy where applicable;
4. authenticated provenance tying the subject to an allowed builder, source
   repository/revision, build type, parameters, and complete materials;
5. Curator policy constraints for package, version, origin, builder, signer,
   source, target, allowed dependencies, and freshness/rollback state;
6. recursive handling of bundled/nested binaries—one verified outer signature
   does not automatically authorize unenumerated inner subjects;
7. time-of-check/time-of-use identity by returning a protected handle or
   publishing the exact verified bytes before any loader/linker can consume
   them.

The receipt must include verifier and policy IDs/versions, trust-root set,
candidate identity, normalized platform facts, signature and provenance
subjects, builder/signer identities, source/material digests, transparency or
timestamp evidence, each check and result, decision/diagnostic, and a canonical
receipt digest. The receipt becomes a separate build-checkpoint input.

This seam is consistent with established provenance/verification models:
SLSA requires provenance to identify output artifacts by cryptographic digest
and describe how they were produced, with stronger levels authenticating the
provenance and protecting it from tenant forgery
([SLSA build requirements](https://slsa.dev/spec/v1.2/build-requirements)).
Sigstore bundles carry signature verification material plus transparency-log or
timestamp evidence for short-lived certificates
([Sigstore bundle format](https://docs.sigstore.dev/about/bundle/)). TUF target
metadata binds target path, length, hashes, delegated trust, and signed metadata
state
([TUF specification](https://theupdateframework.github.io/specification/)).
These are useful inputs/models, not a decision to adopt any one system or a
claim that signature/provenance alone satisfies Curator policy.

## Conformance cases

Every adapter must run the same byte fixtures through the shared classifier and
assert the stable class, decision, primary code, canonical path, and manifest
digest. Adapter tests may add ecosystem wrappers but may not replace the shared
vectors.

### Admission vectors

| ID | Fixture | Expected result |
| --- | --- | --- |
| `A01` | Authored source plus lock/manifest/license text, immutable origin | `ADMIT_INPUT` |
| `A02` | Executable-mode shell script with valid grammar and shebang | `source.authored_text`, admit; hook execution remains separately controlled |
| `A03` | Minified JS and source map shipped as exact immutable inputs | `source.generated_text`/`text.metadata`, admit under Node profile |
| `A04` | Generated `.swiftinterface` plus rebuilding Swift source, no compiled library | Text inputs admit; build remains source-only |
| `A05` | Source-only ZIP -> nested `.tar.gz` -> source files, all safe | All containers descend; root admits with complete chain evidence |
| `A06` | Source-only wheel with valid `RECORD`, no compiled/opaque members, under an enabled Python source-container profile | Container admits; wheel name alone is not asserted as proof |
| `A07` | Manager-selected external compiler/SDK outside closure with stable full fingerprint | `ALLOW_TOOLCHAIN`; not present in source manifest as a dependency |
| `A08` | Object, library, addon, and executable produced after admission by declared steps in clean staging | `ALLOW_OUTPUT` only with complete receipt and protected publication |

### Compiled-payload rejection vectors

| ID | Fixture | Expected class / primary code |
| --- | --- | --- |
| `C01` | Valid ELF `ET_EXEC` and `ET_REL`, each with correct, wrong, and absent suffix | `native.executable` and `native.object`; `artifact_compiled_dependency_forbidden` |
| `C01a` | Pinned GNU dynamic PIE (`-fPIE -pie`): `ET_DYN`, `DF_1_PIE`, and `PT_INTERP`; repeat as executable-looking, `.so`, renamed `.dat`, and no-suffix paths | `native.executable`, variant `elf.pie.interpreter`; same primary code and identical content-derived class for every name |
| `C01b` | Pinned GNU static PIE (`-fPIE -static-pie`): `ET_DYN`, `DF_1_PIE`, and no `PT_INTERP`; repeat with wrong and absent suffix | `native.executable`, variant `elf.pie.no_interpreter`; same primary code |
| `C01c` | Ordinary shared object (`-fPIC -shared -Wl,-soname,libcase.so`): `ET_DYN`, no `DF_1_PIE`, valid `DT_SONAME`; repeat renamed and suffixless | `native.library.dynamic`, variant `elf.shared_object`; same primary code |
| `C01d` | Structurally valid `ET_DYN` with no `DF_1_PIE`, no `DT_SONAME`, and no resolved use edge; plus a variant with `PT_INTERP` but no resolved execute edge | `native.elf.et_dyn_ambiguous`, reason `insufficient_evidence`; same primary code, never `artifact_type_ambiguous` as primary |
| `C01e` | `ET_DYN` with conflicting non-decisive facts (`PT_INTERP` plus `DT_SONAME`, or both resolved use edges) | `native.elf.et_dyn_ambiguous`, conflict reason; same primary code |
| `C01f` | Legacy/no-flag `ET_DYN` paired separately with a manager-resolved execute edge plus `PT_INTERP`, or a manager-resolved link/load edge | `native.executable` or `native.library.dynamic` by rules 2/3; same primary code; the exact edge origin is audited |
| `C02` | Valid PE image, DLL, COFF object, and COFF archive/import library | Native class; same code |
| `C03` | Thin/fat Mach-O executable, object, dylib, and bundle | Native class; same code; every fat slice inventoried |
| `C04` | `.a`, `.lib`, and `.rlib` nested at archive depths 1, 2, and 8 | Static library; same code with complete container chain |
| `C05` | `.framework` and `.xcframework`, including renamed/nested and resource-only bundle | Apple bundle class; same code |
| `C06` | ELF/PE/Mach-O `.node` addon and ABI-tagged Python `.so`/`.pyd` | Specialized native-extension class; same code |
| `C07` | Valid JVM `.class` direct and inside JAR-inside-ZIP; DEX inside AAR | VM bytecode; same code |
| `C08` | Valid WebAssembly binary direct, renamed `.dat`, and nested in npm/wheel source container | `ir.webassembly`; same code |
| `C09` | Python `.pyc`, V8 cached-data/snapshot, LLVM `.bc` and wrapper, Swift `.swiftmodule`, Clang PCM/PCH | Corresponding VM/IR class; same code |
| `C10` | Compiled bytes with executable mode cleared, or text extension | Content-derived deny unchanged |
| `C11` | Compiled dependency bytes identical to an approved toolchain component | Still reject because role is `dependency_input` |
| `C12` | Same compiled fixture presented through Go, Rust, Node, SwiftPM, and Python adapter harnesses | Identical shared class/code/manifest leaf; no adapter override |

### Archive, ambiguity, and fail-closed vectors

| ID | Fixture | Expected code |
| --- | --- | --- |
| `F01` | Absolute, drive, UNC, backslash, `..`, NUL/control, overlong path | `artifact_archive_unsafe_path` |
| `F02` | Duplicate path and case/NFC/trailing-dot collision | `artifact_archive_unsafe_path` with deterministic collision key |
| `F03` | Symlink/hardlink (even internal), device, FIFO, socket, sparse/external extent | `artifact_archive_unsafe_entry` |
| `F04` | Encrypted ZIP member | `artifact_archive_encrypted` |
| `F05` | Unsupported compression, multi-volume ZIP, filesystem image | `artifact_archive_unsupported` |
| `F06` | Truncated header/index, CRC/hash/size disagreement, illegal trailing polyglot | `artifact_archive_invalid` or compiled deny if a compiled detector soundly matches |
| `F07` | Nested source archive at depth 9 | `artifact_inspection_limit_exceeded`, `max_archive_depth` |
| `F08` | 100,001 entries, 257 MiB leaf, >2 GiB emitted, or >200:1 stream | Same code with the exact exceeded limit |
| `F09` | Parser read error/cancellation and incomplete traversal | `artifact_inspection_unavailable`; no cached allow receipt |
| `F10` | Text bytes named `.so`/`.node`/`.wasm` | `artifact_type_ambiguous` |
| `F11` | Valid compiled bytes named `.txt` | `artifact_compiled_dependency_forbidden` |
| `F12` | Valid UTF-8 not accepted by a declared grammar and unknown binary blob | `artifact_opaque_dependency_forbidden` |
| `F13` | Pure source archive missing immutable origin/checksum evidence | `artifact_origin_unverified` |
| `F14` | Archive order permutations of the same normalized members | Same decision, primary diagnostic, canonical manifest bytes/digest |

### Trust-boundary and future-capability vectors

| ID | Fixture | Expected result |
| --- | --- | --- |
| `T01` | Package places `rustc`, `node`, `swiftc`, or `clang` under `vendor/toolchain/bin` | Compiled dependency reject; name/path cannot establish toolchain role |
| `T02` | Package declares arbitrary host `/usr/local/libfoo` as a system library | `artifact_toolchain_untrusted` unless already inside an independently approved toolchain/sysroot |
| `T03` | Toolchain link escapes root, special node appears, or bytes mutate during fingerprint/use | Existing/specialized toolchain diagnostic; no build |
| `T04` | Pre-existing binary copied/hard-linked into nominal output directory | `artifact_local_output_unreceipted` |
| `T05` | Protected output digest/size/path or complete input differs from receipt | `artifact_local_output_drift` |
| `V01` | Signed binary while capability absent | `artifact_binary_admission_unavailable` |
| `V02` | Capability present, but signature subject digest/size differs | Future verifier reject; no fallback to source path |
| `V03` | Valid signature but provenance missing/untrusted, builder/source/material policy mismatch | Future verifier reject |
| `V04` | Provenance passes but target OS/arch/ABI/minimum runtime mismatches | Future verifier reject |
| `V05` | Outer bundle attested but inner binary absent from authorized subjects | Future verifier reject |
| `V06` | All signature, provenance, identity, target, nested-subject, and policy checks pass; exact bytes protected before use | Future `verified-binary-v1` receipt and explicit binary graph edge; never `source` |

Each negative case also asserts that no package-manager hook/build process ran,
no output/cache entry was published, and the diagnostic/audit record contains
the triggering virtual path and container chain.

## Implementation-ready recommendations

1. Add a language-neutral `artifactpolicy` package with closed class/role/code
   enums, canonical manifest schema, streaming limit accountant, path validator,
   and detector registry. Keep adapter packages as callers.
2. Place inspection after immutable package capture/checksum verification and
   before dependency extraction into a build snapshot or execution of any
   package script.
3. Make container readers return virtual members to the same classifier. Do not
   use package-manager-installed trees as the only evidence because extraction
   may overwrite duplicates, normalize paths, or omit archive metadata.
4. Build shared positive/rejection fixtures once, including renamed and nested
   forms, then wrap them in every adapter's conformance suite.
5. Extend the current build metadata/receipt model to bind artifact-manifest and
   graph digests and sorted multi-output records while preserving exact-input
   validation and protected-cache semantics.
6. Keep `BinaryAdmissionVerifier` absent in this cycle. Land only the explicit
   interface/capability boundary if implementation needs a stable seam; do not
   add an allow flag, dummy verifier, signature-presence exception, or adapter
   escape hatch.
7. Version changes to detector coverage, inert-data formats, path rules, or
   limits as policy changes that invalidate prior admission receipts.

## Key findings and decisions

- **Container labels are not content evidence.** JAR and wheel are ZIP-based;
  npm/source packages commonly add compression layers. Recursive inspection is
  required even when ecosystem metadata says “pure” or a package checksum
  matches.
- **Objects and bytecode belong in the deny set.** They are linkable/loadable
  program representations despite not being stand-alone executables.
- **ELF `ET_DYN` is a representation type, not a complete use class.** An
  explicit `DF_1_PIE` resolves a PIE; `DT_SONAME` or a validated graph edge can
  establish shared-object use; missing or conflicting evidence stays a named
  compiled ambiguity. All branches retain the same dependency rejection.
- **Trust is causal, not locational.** Only Curator selection establishes an
  external toolchain, and only observed protected production establishes a
  local output. A package path or manifest cannot establish either role.
- **Unknown is rejection, not absence of evidence.** Parser gaps, encrypted
  members, resource exhaustion, and ambiguous/polyglot content all prevent the
  proof required by source closure.
- **Current receipt hashes are not future binary provenance.** The repository
  explicitly treats them as consistency metadata. Verified binary admission
  needs an independent policy-scoped signature/provenance receipt.
- **The scanner is not a malware detector.** Its finite boundary is compensated
  by offline execution, declared inputs/hooks/write sets, manager-selected
  toolchains, and protected output publication.

## Acceptance mapping

| Task requirement | Evidence in this document |
| --- | --- |
| Deterministic decision for every artifact class | Trust-role matrix, complete taxonomy including `native.elf.et_dyn_ambiguous` and catch-all `opaque.unknown`, ordered ELF resolution, and ordered general algorithm |
| Reject compiled dependency payloads across all adapters | `dependency_input` invariant, deny-dominates rule, native/VM/IR classes, and `C12` |
| Distinguish external toolchains and locally built protected outputs | Trust roles, checkpoint/receipt requirements, and `T01`-`T05` |
| Recursive detection and rejection | Container registry, path/entry policy, closed limits, and `F01`-`F14` |
| Ambiguous cases and conservative defaults | Ambiguity table; empty inert-data allowlist; opaque/encrypted/unsupported rejection |
| Stable diagnostics and audit evidence | Diagnostic table and three evidence schemas |
| Future verified-binary capability boundary | `verified-binary-v1` seam and `V01`-`V06` |
| Conformance requirements | Admission, compiled, archive/fail-closed, trust-boundary, and future-capability vector tables |

## Fact-check record and references

The policy decisions above are normative recommendations. Format/ecosystem
facts were checked on 2026-08-11 against these primary or project-authoritative
sources:

- Curator source-closure invariant and prohibition:
  [repository specification](../.spec/skill-facing-cli-source-closure.md),
  especially lines 39-78.
- Existing stable diagnostics, canonical build inputs, toolchain fingerprints,
  receipts, and protected-cache checks: `internal/godriver/errors.go`,
  `internal/godriver/fingerprint.go`, `internal/buildmeta/models.go`,
  `internal/buildmeta/codec.go`, and `internal/buildcache/cache.go`.
- Native formats:
  [ELF object-file introduction](https://refspecs.linuxfoundation.org/elf/gabi4%2B/ch4.intro.html),
  [ELF header/type definitions](https://refspecs.linuxfoundation.org/elf/gabi4%2B/ch4.eheader.html),
  [current ELF program-header rules](https://gabi.xinuos.com/elf/07-pheader.html),
  [current ELF dynamic-tag rules](https://gabi.xinuos.com/elf/08-dynamic.html),
  [GCC PIE/static-PIE/shared link options](https://gcc.gnu.org/onlinedocs/gcc-13.3.0/gcc/Link-Options.html),
  [GNU binutils `ET_DYN` clarification](https://gcc.gnu.org/pipermail/binutils/2020-May/111252.html),
  [GNU `readelf` `DF_1_PIE` detection](https://sourceware.org/pipermail/binutils/2021-June/116921.html),
  [Microsoft PE/COFF](https://learn.microsoft.com/en-us/windows/win32/debug/pe-format),
  [Apple Mach-O overview](https://developer.apple.com/library/archive/documentation/Performance/Conceptual/CodeFootprint/Articles/MachOOverview.html),
  and [Apple XCFramework documentation](https://developer.apple.com/documentation/Xcode/creating-a-multi-platform-binary-framework-bundle?language=objc).
- Native extensions and Python packaging:
  [Node addons](https://nodejs.org/api/addons.html),
  [Python C extensions](https://docs.python.org/3.12/extending/building.html),
  [wheel format](https://packaging.python.org/en/latest/specifications/binary-distribution-format/),
  [source distribution format](https://packaging.python.org/en/latest/specifications/source-distribution-format/),
  and [Python bytecode compilation](https://docs.python.org/3/library/py_compile.html).
- VM/IR and archive formats:
  [JVM class format](https://docs.oracle.com/javase/specs/jvms/se24/html/jvms-4.html),
  [JAR format](https://docs.oracle.com/en/java/javase/26/docs/specs/jar/jar.html),
  [WebAssembly binary modules](https://webassembly.github.io/spec/core/binary/modules.html),
  [LLVM bitcode](https://llvm.org/docs/BitCodeFormat.html),
  [Swift module stability](https://www.swift.org/blog/abi-stability-and-more/),
  and [Node/V8 cached-data identity](https://nodejs.org/api/v8.html#v8cacheddataversiontag).
- Package containers:
  [npm pack](https://docs.npmjs.com/cli/v11/commands/npm-pack/), which emits
  `.tgz` tarballs, and
  [Cargo packaging](https://doc.rust-lang.org/cargo/reference/publishing.html),
  which produces compressed `.crate` source packages.
- Archive safety:
  [Python tar extraction/security guidance](https://docs.python.org/3/library/tarfile.html)
  and [Go ZIP reader](https://pkg.go.dev/archive/zip).
- Future provenance/verification seam:
  [SLSA 1.2 build requirements](https://slsa.dev/spec/v1.2/build-requirements),
  [Sigstore bundle format](https://docs.sigstore.dev/about/bundle/), and
  [The Update Framework specification](https://theupdateframework.github.io/specification/).

No claim in this document says that the future capability already exists, that
a current Curator receipt authenticates provenance, or that a file suffix or
magic prefix alone proves a valid artifact. Those would contradict the checked
sources and the fail-closed decision procedure.
