# Decision 0010: Kotlin artifact model and the `kotlin-native` driver pair

Record numbers 0007 and 0008 are taken by accepted sibling records that have not
yet landed on `main`; 0009 is claimed concurrently by `TASK-260728-12pnm1`
(Rust) and `TASK-260728-1jafds` (hardened execution profile). This record takes
0010 so that the set can land in any order. If a lower free number is available
when this record lands, it renumbers rather than contests, exactly as
`TASK-260728-2spy93` did when 0007 was claimed.

## Context

Decision 0008 reserved six additional-language driver identifiers, two of them
Kotlin — `kotlin-native-v1` and `kotlin-native-repository-v1` — and made
`TASK-260728-168smo` responsible for one thing the other two language contracts
did not have to answer:

> `TASK-260728-168smo` additionally decides the Kotlin backend **within**, not
> around, section 3; if no Kotlin backend satisfies it, both Kotlin identifiers
> are retired unused.

Decision 0008 section 3 admits exactly one artifact class,
`native-executable-v1`: one bounded regular file, named solely by the manager,
directly executable by the host program loader using only base-installation
libraries, never executed by the manager. Its section 6 item 3 additionally
requires every executable started below the worker to be a fingerprinted member
of the driver's declared trusted toolchain closure, and states that a
host-resolved tool outside the closure is not admissible.

Decision 0007 reserved the `kotlin` toolchain entry and the companion-only `jdk`
identifier, requires the primary executable to be a regular executable at the
entry's fixed relpath inside the tree being fingerprinted, and admits exactly
two resolution origins: a toolchain bundled with the manager distribution, and
trusted operator configuration in manager-owned, owner-protected state.

The first version of this record was written without a Kotlin/Native
distribution on any reachable host. Review cycle 1 obtained one and disproved
three of its load-bearing assumptions: the official distribution contains no
regular executable, its Apple dependency closure is not present in the release
archive, and an empty `KONAN_DATA_DIR` is mutated by the compiler. This revision
replaces the design those assumptions supported. Everything below marked
**MEASURED** was taken on macOS 26.5 arm64 against the checksum-verified
official release `kotlin-native-prebuilt-macos-aarch64-2.4.10.tar.gz`
(`55ded039…ed9d`, matching the published `.sha256` asset), with argv and real
exit codes recorded in `TASK-260728-168smo_command-evidence.log`.

The complete implementation-ready reference is
[`docs/kotlin-native-build-drivers.md`](../docs/kotlin-native-build-drivers.md).

This record defines no schema file, regenerates no vector, creates no release
metadata, advances no pin, and claims no platform. Both identifiers remain
reserved. Nothing here admits them to the wire.

## Decision

### 1. The artifact model: Kotlin/Native, and it is measured

**MEASURED, and therefore DECIDED.** The Kotlin family is admitted through the
Kotlin/Native backend producing a single native executable, or not at all. The
question decision 0008 section 10 asks — does *any* Kotlin backend satisfy
section 3 — is answered in the affirmative by a produced artifact rather than by
argument:

```text
konanc -produce program -target macos_arm64 -o out/app src/main.kt   exit 0
out/app.kexe   785,672 B   Mach-O 64-bit executable arm64
./out/app.kexe -> "curator-kotlin-native-probe", exit 0
otool -L      -> libSystem.B.dylib, libc++.1.dylib, libobjc.A.dylib,
                 Foundation, CoreFoundation   — all base-installation
```

One bounded regular file, directly loadable, no runtime to install, no launcher,
no classpath. Every Kotlin/JVM shape remains rejected, and the rejection is a
consequence of section 3 rather than a preference:

| Candidate | Published files | Directly loadable | Verdict |
|---|---|---|---|
| Kotlin/JVM thin JAR plus operator-installed JRE | 1 | no — needs a JVM and a classpath | rejected: `runtime-bundle` |
| Kotlin/JVM fat JAR (`-include-runtime`) | 1 | no — still needs a JVM | rejected: `runtime-bundle` |
| Kotlin/JVM plus `jlink`/`jpackage` runtime image | many | launcher plus runtime tree | rejected: `runtime-bundle` |
| Kotlin/JVM plus GraalVM `native-image` | 1 | yes | rejected: see rejected alternatives |
| **Kotlin/Native `-produce program`** | **1** | **yes, measured** | **selected** |

The fat JAR keeps its explicit row because it is the shape most often described
as self-contained. It is one file and still fails section 3's third bullet. One
file is necessary and not sufficient.

A future Kotlin/JVM driver, if a `runtime-bundle` profile is ever reviewed and
accepted, MUST use a different family segment and MUST NOT reuse either
identifier reserved here, exactly as decision 0008 section 2 requires.

### 2. Identifiers, and no new names

**DECIDED.** This record coins no identifier. It adopts, unchanged, the two
names decision 0008 section 2 reserved:

| Driver | Source mode | Receipt schema | Execution policy | Status |
|---|---|---|---|---|
| `kotlin-native-v1` | local snapshot | 3 | `manager-worker-v2` | reserved |
| `kotlin-native-repository-v1` | external repository | 4 | `manager-worker-v2` | reserved |

Reservation is not admission. Until `TASK-260728-251p01` moves them in the same
change that mints the schema admitting them, every schema MUST reject both and a
manager MUST treat both as unknown drivers. Section 12 fixes the conditions
under which they are instead retired unused.

### 3. The trusted root is an operator-curated bundle, not the vendor archive

This is the change review cycle 1 forced, and it is the centre of the record.

**MEASURED — the vendor archive cannot be the toolchain root.** Every entry in
`<dist>/bin` — `konanc`, `kotlinc-native`, `run_konan`, `cinterop`, `klib`,
`konan-lldb`, `generate-platform` — is a Bourne-Again shell script. The compiler
is one JAR, `konan/lib/kotlin-native-compiler-embeddable.jar`. There is no
regular executable anywhere in the distribution, so decision 0007 section 3
cannot be satisfied by it. `run_konan` additionally resolves `java` by name from
`PATH` unless `$JAVACMD` or `$JAVA_HOME` says otherwise, and appends `-D` and
`-J` arguments to the JVM: a wrapper taking environment-selected pipeline
inputs, which decision 0007 section 3 and decision 0008 section 6 item 3 both
forbid.

**DECIDED — `curator-kotlin-bundle-v1`.** The `kotlin` toolchain root is an
operator-curated tree with a fixed layout, resolved through decision 0007's
second admissible origin — trusted operator configuration in manager-owned,
owner-protected state — and fingerprinted whole:

```text
<kotlin_root>/                                  immutable, fingerprinted whole
  jdk/                                          a JDK runtime
    bin/java            (bin/java.exe on Windows)      <- primary_relpath
  kotlin-native/                                the unpacked official distribution
    konan/lib/kotlin-native-compiler-embeddable.jar
    konan/konan.properties
    klib/…
  konan-data/                                   the prehydrated dependency closure
    dependencies/<name>/…
    dependencies/.extracted
```

Four properties follow, and each answers one review finding:

1. **The primary executable is real.** `jdk/bin/java` is a regular executable at
   a fixed relpath inside the tree being fingerprinted — decision 0007 section 3
   satisfied literally. The "narrow reading" of section 3 proposed in the first
   version of this record is **withdrawn**; no reading is needed.
2. **There is no companion.** Because the JDK is inside the primary root, the
   companion list is **empty**, `toolchain_identities` is a one-element array,
   and one tree digest covers the entire executable closure. The `jdk`
   identifier reserved by decision 0007 is **not used** by this decision and
   stays reserved with no entry and no driver mapping. This supersedes the first
   version of this record, which made `jdk` a REQUIRED companion, and it is
   strictly stronger: an external companion root would be a second tree the
   manager trusts on the operator's word.
3. **The dependency closure is inside the fingerprint.** The Kotlin/Native
   compiler needs LLVM, a target toolchain, a sysroot and per-target extras that
   the release archive does not contain. **MEASURED**: a first run downloads
   them (`lldb-4-macos` 64,230,999 B,
   `llvm-21-aarch64-macos-essentials-97` 151,150,049 B,
   `libffi-3.3-1-macos-arm64` 17,037 B; 688 MB extracted) from
   `download.jetbrains.com`, with no integrity check reported by the compiler.
   Hydration is therefore an **operator act performed once, outside any Curator
   operation**, and its result is inside the fingerprinted tree. The manager
   never downloads a dependency, at any time, in any mode.
4. **Provenance is the operator's, and the manager proves only stability.** The
   manager reads no bundle descriptor and MUST NOT: a manifest inside the root
   describing the root would be a second trust input the manager cannot verify.
   The identity is decision 0007's — algorithm identifier, normalized native
   version, primary-executable relpath, tree digest — and it proves the tree is
   stable across and identical between operations, not that it is genuinely the
   vendor's. Verifying that remains the operator's responsibility at
   configuration time, exactly as decision 0007's security section already
   states for every toolchain. The curation procedure the operator documentation
   MUST carry is fixed in the reference: fetch the official archive, verify the
   published checksum, unpack, hydrate once with network access on the same
   platform, verify that no dependency the target needs is `remote:internal`,
   place a JDK, then make the tree read-only.

**DECIDED — the process graph.** `manager-worker-v2`'s two lower nodes bind, for
both Kotlin drivers, to:

```text
manager parent
  -> identity-verified manager-owned worker
       -> <kotlin_root>/jdk/bin/java                    (the trusted launcher)
            -> the Kotlin/Native compiler, in-process in that JVM,
               loaded by -cp from <kotlin_root>/kotlin-native
                 -> regular executables inside <kotlin_root>
                    (LLVM clang++, the assembler, the archiver, the linker)
```

**MEASURED** for the in-closure half: the LLVM driver the compiler spawns is
`<konan-data>/dependencies/llvm-21-aarch64-macos-essentials-97/bin/clang++`,
inside the curated tree. The driver MUST NOT execute `bin/konanc`,
`bin/kotlinc-native`, `bin/run_konan`, `bin/kotlinc`, or any other launcher
script from any distribution, on any platform, including for the version probe.

### 4. The operation-private overlay, and the no-download proof

Three measurements fix this section, and they correct the cycle-1 record in both
directions.

**MEASURED — a hydrated closure is not mutated by a compile.** The aggregate
SHA-256 over all 1,470 files of a hydrated `KONAN_DATA_DIR` is byte-identical
before and after a successful compile. The mutation review cycle 1 observed —
`dependencies/`, `dependencies/.extracted`, `dependencies/cache/.lock` appearing
in an empty directory — is a *hydration* write, not a compile write.

**MEASURED — the data directory must nevertheless be writable.** With the tree
made read-only the compile fails: `FileNotFoundException:
…/dependencies/cache/.lock (Permission denied)`. A read-only distribution is
fine; a read-only data directory is not.

**MEASURED — the overlay closes the gap.** With `<kotlin_root>` fully read-only
and `KONAN_DATA_DIR` pointed at an operation-private directory holding one entry
per prehydrated dependency materialised from the bundle, a copy of `.extracted`,
and a fresh writable `cache/`, the compile exits 0 and produces the artifact,
and both read-only roots are unchanged.

**DECIDED.** `KONAN_DATA_DIR` is an operation-private writable overlay,
materialised from `<kotlin_root>/konan-data` by a manager-owned mechanism that
copies or links only, adds no entry that is not in the bundle, reaches no
network, and leaves the fingerprinted root byte-unchanged. The overlay is never
fingerprinted, never published, and is discarded with the operation. The exact
materialisation mechanism is per platform and is fixed by that platform's
qualification, because a symlink farm is measured on macOS and is not
universally available on Windows.

**DECIDED — the driver fails closed rather than downloading.** The compile
vector carries the single manager-owned constant override
`-Xoverride-konan-properties=airplaneMode=true`. **MEASURED**: with it, a
missing dependency is `Cannot find a dependency locally: <name>` and exit 2,
with no download attempted. This is an in-process guarantee that does not depend
on the network being unreachable, and it composes with the network denial the
acceptance test requires. **MEASURED** independently: with the closure hydrated
and all network denied, the compile exits 0 and logs no download.

### 5. The project-metadata file: `kotlin-native-module.json`

**DECIDED**, unchanged from the first version of this record, which review
cycle 1 did not contest.

Decision 0008 section 4 requires each local driver to bind exactly one closed
**driver-defined** project-metadata file that exists directly in the build root
and is the nearest ancestor of `source_dir`. Decision 0007 section 1.3 requires
the `kotlin` entry to name exactly one file and one field. Kotlin has no such
file, and every ecosystem candidate fails:

| Candidate | Why rejected |
|---|---|
| `build.gradle.kts` / `build.gradle` | Reading it is executing it. `TASK-260729-rhjxtx` measured that the pure metadata query `gradle properties` compiles the build script as a program source unit (`_BuildScript_`) before it can answer. This is the generic Gradle escape hatch the epic exists to refuse. |
| `settings.gradle{,.kts}` | Same class, same measurement. |
| `gradle.properties` | Not executable, but a Gradle input whose `kotlin.*` keys select compiler and daemon behaviour — package-selected build behaviour under a name that merely looks inert. |
| `pom.xml` | Parsable without execution, but the Maven project model is its `<build><plugins>` graph; reading one field while ignoring that graph would let a package ship a plugin declaration Curator silently does not honour. |
| A bare marker file with no field | Fails decision 0007 section 1.3, which requires one file **and** one field. |
| `.kotlin-version` plain text | No schema version, no closure, no way to reject an added line. |

The driver therefore binds a Curator-owned file, which is exactly what
"driver-defined" permits and what `skill-build.json` already establishes as
protocol style for a Curator-owned file inside package-supplied source:

```json
{"schema_version": 1, "kotlin_version": "2.4.10"}
```

Exactly `schema_version` and `kotlin_version`, both REQUIRED,
`additionalProperties: false`. `schema_version` is the `const` integer `1` and
participates in the file-shape gate only. `kotlin_version` is the sole
`metadata_sources` field of the `kotlin` entry: a canonical `major.minor.patch`
triple in decision 0007 section 2.1's grammar, asserting the Kotlin compiler
version the sources are written against. It is never passed to the compiler and
contributes no argument.

Three consequences, each a simplification rather than a special case. The
classifier is two classes rather than seven, because Curator owns the grammar.
The security partition `F` is empty, because a version cannot name where a
toolchain comes from, so decision 0007's P1 and P2 collapse to the satisfiable
equality `C = Upstream` over Curator's own grammar. And the file is inert to the
compiler, so deleting it produces a deterministic Curator rejection and an
unchanged compiler.

`source_dir` maps to exactly one compiled program without discovery: the program
is the recursive `.kt` source set under `source_dir`, compiled in `program` mode
with the compiler's default entry point. Zero or multiple entry points is a
compiler error and therefore deterministic; the manager never names, searches
for, or infers one.

### 6. The command surfaces

**DECIDED.** Nothing Kotlin-specific is added to either command shape. The local
command is decision 0008's `buildCommandV8` with `driver` the `const`
`kotlin-native-v1`:

```json
{"type":"build","driver":"kotlin-native-v1","source_dir":"build/cmd/tool",
 "toolchain":{"id":"kotlin","version":{"kind":"at_least","min":"2.4.10"}}}
```

The external command is `repositoryBuildCommandV2` with `driver` the `const`
`kotlin-native-repository-v1`, and the descriptor target is
`skillBuildTargetV2` with the same `const`, `build_root`, `source_dir`, and the
OPTIONAL `toolchain`. `toolchain.id` MUST be `kotlin`; there is no companion to
express and none may be added.

No member is added anywhere. In particular no manifest, descriptor, or metadata
file may express a target, a Kotlin/Native target triple, an entry point, a
module name, a klib, a library, a linker option, an opt-in marker, a language or
API version flag, a compiler plugin, a plugin option, a memory model, a GC
selection, a bitcode or debug setting, a cache mode, a JVM option, a classpath,
a `KONAN_DATA_DIR`, a `konan.properties` override, or a dependency source.

### 7. The pre-compile rejection matrix: an allow-list, not a deny-list

**DECIDED.** Decision 0008 section 7 requires an exhaustive, deterministic,
pre-compile rejection matrix over every package-selected code-execution surface.
Kotlin's surfaces are an open set — a general-purpose script engine, a script
dialect the compiler itself can execute, annotation processing, compiler
plugins, C interop — so an enumerated deny-list could not be exhaustive.

The matrix is a **closed allow-list over the compiler-visible tree**, computed
from the validated immutable snapshot, before the compile phase, by a
manager-side walk that runs no compiler and reaches no network. The build root
and every directory below it admit exactly:

1. directories whose names match `^[A-Za-z0-9_][A-Za-z0-9_.-]*$`;
2. regular files whose names match `^[A-Za-z0-9_][A-Za-z0-9_.-]*\.kt$`; and
3. exactly one additional regular file, `kotlin-native-module.json`, directly in
   the build root and nowhere else.

Every other entry is rejected under `build_package_code_execution_forbidden`
with a per-surface diagnostic. There is no exception, no opt-out, no advisory
mode, and no configuration that widens it. The rule subsumes, by construction
rather than by enumeration, Gradle and Maven files and wrappers, `.kts` script
sources, kapt and KSP configuration, `META-INF/services`, C-interop `.def`
files, headers and native sources, objects, archives and shared libraries,
prebuilt `.klib`, `.jar` and `.aar`, every non-regular entry, every dotfile, and
every name beginning with `@`.

**MEASURED — the `@` row, corrected.** The first version of this record
described the response-file vector as defeated by absolute paths. That is the
right conclusion for the wrong reason, and the reason matters because an
implementer reproduces the reason rather than the conclusion. On the native
backend:

| Argument token | File on disk | Outcome |
|---|---|---|
| `@inject` | `inject` | expanded — `-version` honoured, exit 0 |
| `@./inject` | `inject` | expanded, exit 0 |
| `@/abs/path/inject` | `inject` | **expanded**, exit 0 |
| `/abs/path/@inject` | `@inject` | not expanded — `source entry is not a Kotlin file`, exit 1 |
| `./@inject` | `@inject` | not expanded, exit 1 |
| `@nonexistent` | — | `warning: argfile not found`, warning only |

Expansion is decided by the **first character of the argv token**; the `@` is
stripped and the remainder is the response-file path, absolute or not. So an
absolute package path is safe because it starts with `/`, not because it is
absolute — and a driver that ever prefixed a path with `@` would reopen the
whole surface even in absolute form. A missing response file is a warning, so a
partial mitigation fails silently.

Two independent layers close it, and the contract requires both: the normative
layer is the allow-list, which rejects the name before the compile phase as
decision 0008 section 7 demands; the structural backstop is the argv discipline
of section 8, which never emits a token whose first character is `@`.

The four argv-only plugin surfaces — `-Xplugin=`, `-Xcompiler-plugin=`,
`-P plugin:`, and `-script-templates` — are structurally unreachable, because
the closed command surface gives a package no way to supply an argument and the
driver's vector is fixed. They are named so the matrix is exhaustive over
surfaces rather than over file names, and so that any future change letting
package data reach argv is visibly a change to this paragraph.

### 8. The worker session, the argument vector, and the environment

**DECIDED — session shape.** A Kotlin operation uses **zero** graph-phase
commands and exactly one compile-phase command. Decision 0008 section 7's
"at most one" graph phase is what admits this: there is no dependency resolution
to perform, because section 7 admits no package-supplied library and section 6
admits no dependency declaration, so the source set is decided by a manager-side
directory walk rather than by asking a tool. The Kotlin compile daemon is
forbidden; no daemon argument is passed and the driver MUST fail rather than
fall back to one.

**DECIDED — argument vector.** One process, one command, no retry, no second
phase:

```text
<kotlin_root>/jdk/bin/java
  -ea -Xmx3G -XX:TieredStopAtLevel=1
  -Dfile.encoding=UTF-8 -Duser.language=en -Duser.country=US
  -Dkonan.home=<kotlin_root>/kotlin-native
  -cp <kotlin_root>/kotlin-native/konan/lib/kotlin-native-compiler-embeddable.jar
  org.jetbrains.kotlin.cli.utilities.MainKt
  konanc
  -Xoverride-konan-properties=airplaneMode=true
  -produce program
  -target <resolved default native target>
  -o <operation-private-staging>/<command>
  <absolute path to source 1>
  …
  <absolute path to source N>
```

The JVM options are the vendor launcher's own, minus everything it takes from
the environment. `-Duser.language`/`-Duser.country` are the one element added
beyond the measured vector, for locale-independent diagnostics; qualification
re-measures the vector including them. Binding rules:

- sources are the recursive `.kt` set under `source_dir`, enumerated by the
  manager, sorted by Unicode-scalar order of relative path, and passed as
  absolute paths whose first character is `/` (or a drive letter on Windows) —
  never as a directory source root, because a directory hands file discovery to
  the compiler and would compile files the allow-list walk rejected;
- no entry point, module name, opt-in, language-version, API-version,
  optimisation, debug, cache, memory-model, or plugin argument is passed, and
  none is package-derivable;
- `-target` names the host's own default native target only. **Cross-compilation
  is not admitted in this version.** **MEASURED**: the distribution states its
  own default, `konanc -list-targets` marking exactly one line `(default)`;
- `-Xoverride-konan-properties` carries exactly the constant above. Its value is
  manager-owned and no package byte can reach it.

**DECIDED — operation-private environment.** In addition to the
`manager-worker-v2` portable control set:

| Variable | Action | Why |
|---|---|---|
| `KONAN_DATA_DIR` | set to the operation-private overlay | section 4 |
| `TMPDIR`, `TMP`, `TEMP` | set to operation-private staging | **MEASURED**: the compiler writes `konan_temp<random>/` intermediates there |
| `KONAN_USE_INTERNAL_SERVER` | unset | **MEASURED**: selects a JetBrains-internal dependency host, `https://repo.labs.intellij.net/kotlin-native` |
| `JDK_JAVA_OPTIONS`, `JAVA_TOOL_OPTIONS`, `_JAVA_OPTIONS` | unset | honoured by the JVM launcher; inject arbitrary JVM options |
| `CLASSPATH` | unset | would extend the compiler classpath |
| `JAVA_HOME`, `JAVACMD`, `JAVA_OPTS` | unset | launcher-script inputs; closed structurally by never running the launcher, unset as defence in depth |
| `KOTLIN_HOME`, `KOTLIN_COMPILER`, `KOTLIN_TOOL`, `KOTLIN_RUNNER` | unset | same; `KOTLIN_COMPILER` selects the compiler main class |
| `PATH` | manager-owned minimal value | no toolchain is resolved through it |

No manager-written configuration file is placed anywhere for this driver. The
pipeline reads no configuration file, because the launcher that would is never
run and the compiler is given an explicit, complete argument vector.

### 9. Platform libraries, native interop, and the published-artifact gate

Review cycle 1 correctly rejected the first version's "standard library and
nothing else / no C interop" claim, and the corrected policy is a change of
substance rather than of wording.

**MEASURED.** The distribution ships 200-plus platform klibs per Apple target
under `klib/platform/<target>/`. Source importing them compiles and links with
no `-library` argument and no `.def` file:

```kotlin
import platform.posix.getpid
import platform.Foundation.NSProcessInfo
```

compiled with the vector of section 8 exits 0, runs, and the produced binary
gains a dynamic dependency the previous program did not have
(`/usr/lib/libresolv.9.dylib`).

**DECIDED — the surface is allowed, and its consequence is bound.** The fixed,
distribution-owned platform library surface is **admitted**: it is inside the
fingerprinted bundle, it is selected by source rather than by any package
control input, a package cannot add to it or name one, and rejecting it would
require the manager to parse Kotlin source, which it does not do. What is
rejected is everything a package could *supply*: `.def` files, headers, native
sources, objects, archives, shared libraries and prebuilt `.klib` are all
rejected by the section 7 allow-list, and `cinterop` is a second tool and a
second command that `manager-worker-v2`'s session admits in neither position.
The honest statement is therefore: **no user-defined C interop, and the
distribution's own platform bindings are available as part of the fingerprinted
toolchain.**

**DECIDED — the published-artifact dynamic dependency gate.** Because the
artifact's dynamic dependency set is a function of package source, decision 0008
section 3's third bullet cannot be discharged by any pre-compile file walk.
Before hashing and before publication, in operation-private staging, the manager
MUST read the produced file's dynamic dependency list and MUST reject the build
with `build_artifact_class_unsupported` if any entry is outside the closed
base-installation library allow-list that the platform matrix fixes for that
tuple. Reading a file's headers is not executing it, so this does not touch
decision 0008 section 3's "never executed by the manager" clause. Each qualified
tuple MUST supply that allow-list as part of its qualification; a tuple with no
allow-list cannot be qualified.

### 10. Registry entries

**DECIDED — shape.** The `kotlin` entry is primary for both drivers with an
**empty** companion list. The `jdk` identifier is not used and stays reserved.

| Field | Value |
|---|---|
| `toolchain_id` | `kotlin` |
| `fingerprint_algorithm` | `curator-kotlin-toolchain-v1` |
| companions | empty |
| `metadata_sources` | `kotlin-native-module.json` → `kotlin_version` |
| `baseline` | `at_least 2.4.10` |
| `compatibility` | **empty**; family granularity `(major, minor)` |
| `platforms` | **empty** |
| `primary_relpath` | `jdk/bin/java`; `jdk/bin/java.exe` on Windows — installed per operating system when that operating system enters `platforms` |
| `probe` | `<kotlin_root>/jdk/bin/java` with the section 8 JVM options, `-cp` the compiler JAR, `<main> konanc -version`, from a manager-owned empty working directory under the section 8 environment |
| `normalization` | `kotlin.konanc.dashversion.stdout`, below |

**Normalization, MEASURED.** `konanc -version` writes exactly
`Kotlin/Native: 2.4.10` — 22 bytes, one line — to **stdout** and exits 0, while
stderr carries JVM warnings and `info: kotlinc-native 2.4.10 (JRE 26.0.1)`. The
rule reads line 1 of stdout, bounded to the first 4 KiB, anchored whole-line:

```text
^Kotlin/Native: (0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(\S*)(?:\s.*)?$
```

Groups 1–3 are the triple; a non-empty group 4 sets the prerelease flag.
Unmatched or multiply-matched output is `build_toolchain_version_undetermined`,
never a default. The literal `Kotlin/Native: ` prefix is load-bearing: it
asserts the **backend**, not merely a version, so a JVM-backend distribution —
which writes `info: kotlinc-jvm …` to stderr with an empty stdout — cannot
satisfy the rule at all. That is the check the first version of this record
demanded and could not write.

**Why `compatibility` and `platforms` are empty and that is not an omission.**
Decision 0007 section 1.1.1 admits a family only after it has been tested
against that driver's conformance vectors; no Kotlin vectors exist, so no family
may be listed. Decision 0007 section 4 declares `primary_relpath` and `probe`
per operating system and only for operating systems in `platforms`, and its
release gate rejects a relpath declared for an operating system outside that
set — so while `platforms` is empty the entry structurally cannot carry a
relpath or a probe, and this record fixes the values qualification installs
rather than smuggling them into the registry early. An entry that looks complete
and is not is worse than one that is visibly reserved.

### 11. Platform position

**DECIDED — macOS is excluded, and the exclusion is measured, not pending.**

`macos/arm64` and `macos/amd64` are **not admissible for this driver family at
Protocol 1.0**, on two independent grounds, either one sufficient.

*Ground one — the process closure.* Under exec containment allowing only
`<kotlin_root>` and the operation-private overlay, a hello-world
`-produce program -target macos_arm64` requires exactly seven executables
outside every curatable root:

```text
/bin/bash
/usr/bin/xcode-select                     (via bash)
/usr/libexec/PlistBuddy                   (via bash)
/usr/bin/xcrun
<Xcode>/Contents/Developer/usr/bin/xcodebuild                       (via xcrun)
<Xcode>/…/Toolchains/XcodeDefault.xctoolchain/usr/bin/ld
<Xcode>/…/Toolchains/XcodeDefault.xctoolchain/usr/bin/dsymutil
```

plus the Xcode macOS SDK as sysroot data. With all seven denied the compile
fails at `CurrentXcode.bash(Xcode.kt:144)`; with all seven allowed it exits 0.
Four of them are OS-owned absolute paths hard-coded in the compiler and can
never be members of an operator-curated root; the other three are discovered at
run time from whichever Xcode the host happens to have.

No manager-fixed input closes this. **MEASURED**: with
`ignoreXcodeVersionCheck=true` and `targetToolchain`, `targetSysRoot` and
`additionalToolsDir` all overridden to absolute local paths, the compile still
fails with `Cannot run program "/usr/bin/xcrun"` at
`CurrentXcode.getToolchain(Xcode.kt:92)` ←
`AppleConfigurablesImpl.getAbsoluteTargetToolchain(Apple.kt:45)`: the Apple
toolchain path is resolved by spawning `xcrun` and the override is not
consulted. The structural cause is in the distribution's own data and code:
`konan.properties` declares the Apple toolchain, sysroot and addon as
`remote:internal`, and the only implementations of `XcodePartsProvider` are
`InternalServer` — gated on `KONAN_USE_INTERNAL_SERVER` and pointing at
`https://repo.labs.intellij.net/kotlin-native`, which no operator can reach —
and `Local`, which is the host's installed Xcode. There is no third source.

This violates decision 0008 section 6 item 3 directly, and decision 0007's
closed identifier set `{go, rust, swift, kotlin, jdk}` contains no identifier
for a platform SDK that contributes a process, so the component cannot be bound
into the toolchain identity either. Rust's data-only SDK root is not a precedent
for an executed linker.

*Ground two — cache identity.* The Apple toolchain and SDK are not in any
fingerprinted tree, so the canonical build input cannot bind them. Two hosts
with different Xcode versions would compute the same cache key for different
artifacts, which is exactly the aliasing decision 0008 section 6 item 4 and
section 8 exist to prevent. Even if the process rule were waived, the identity
rule would still exclude the platform.

An implementation MUST fail with `platform_unsupported` on macOS. It MUST NOT
answer this by resolving a host tool, shipping a shim, declaring the SDK
data-only while executing from it, or publishing a second file.

**DECIDED — Windows and Linux are candidate tuples, and the reason macOS fails
does not apply to them.** In the same `konan.properties`,
`targetToolchain.mingw_x64` and `targetSysRoot.mingw_x64` resolve to
`$toolchainDependency.mingw_x64`, and the two `linux_x64` equivalents to
`$gccToolchain.linux_x64/…`; none is `remote:internal` and none names a host
SDK, so all of them live under `$KONAN_DATA_DIR/dependencies` — inside a
curatable, fingerprintable tree. This is a reading of the distribution's own
data, **not** a measurement of those hosts, and it is not a platform claim:

- **`windows/amd64`** is the only tuple that can admit this pair before Linux
  enters the protocol's platform set at all. It is unqualified, and it is
  qualified only by the acceptance test of section 12 run on a Windows host.
  `TASK-260729-rhjxtx` measured that the reachable Windows host carries no
  Kotlin toolchain of any backend, which the curated-bundle model of section 3
  is exactly what answers. `windows/arm64` is a separate tuple and is not
  implied by `windows/amd64`.
- **`linux/*`** is excluded from the protocol until `TASK-260728-1skseh`, then
  qualified by `TASK-260728-3u1nho`.

An implementation MUST NOT resolve `cmd.exe`, `powershell.exe`, `link.exe`,
`cl.exe`, `lib.exe`, `vswhere.exe`, or a Visual Studio activation script on
Windows; the MinGW-family linking path must be shown to live inside
`<kotlin_root>`.

### 12. The acceptance test and the retirement trigger

**DECIDED.** A `(driver, operating_system, architecture)` tuple enters
`platforms` only when a run on that exact tuple, on an immutable native host,
produces all of the following, each with recorded argv and real exit code:

| # | Requirement |
|---|---|
| A1 | The bundle is curated per section 3 from a checksum-verified official archive; `primary_relpath` resolves to a regular executable; the probe and the anchored `Kotlin/Native: ` normalization of section 10 are reproduced on that host. |
| A2 | A shim-`PATH` run of the full compile records **0** `PATH` resolutions, **and** the paired control run through the shipped launcher records a non-zero count. Without the control firing the zero proves nothing. |
| A3 | Under exec containment allowing only `<kotlin_root>` and the operation-private overlay, the compile exits 0. Every spawned child is recorded by resolved absolute path and is a regular file inside those two trees. The containment control MUST be shown to fire on a deliberately excluded path. |
| A4 | With `airplaneMode=true` and all network denied, the compile exits 0, logs no download, and `<kotlin_root>` is byte-identical before and after; every write lands inside operation-private state. |
| A5 | Exactly one file is produced for publication; its exact name and any compiler-applied suffix are recorded; renaming happens inside operation-private staging only; every by-product — debug bundles, `konan_temp*`, caches — stays in staging. |
| A6 | The published file is a native executable for the tuple, runs, and every entry of its dynamic dependency list is inside that tuple's closed base-installation allow-list, which the qualification supplies. A source importing distribution platform libraries is included in the sample, because that is what moves the list. |
| A7 | The default native target is read from `konanc -list-targets` and mapped to the claim vocabulary `(operating_system, architecture)`. |
| A8 | The section 7 allow-list is exercised with at least one positive and one rejected case per subsumed surface, including a name beginning with `@`, a `.kts` file, and a `.klib`. |
| A9 | `compatibility` gains the tested family, and only that family, after the driver's conformance vectors pass against it. |

Because macOS is excluded by section 11 and Linux is outside the protocol's
platform set until `TASK-260728-1skseh`, the pair's admission rests entirely on
`windows/amd64` at the time this record lands.

**DECIDED — retirement.** Both `kotlin-native-v1` and
`kotlin-native-repository-v1` are **retired unused** if, at the moment
`TASK-260728-251p01` mints manifest schema 8, no tuple has passed A1 through A9.
Retired means what decision 0008 section 2 says: the identifiers are not
reassigned to another language, backend, artifact class, or source mode, and are
not enabled by relaxing another driver. Claim schema 4 is then minted over the
admitted set without them and both become structurally unassertable.

This record does **not** retire them now, and the reason is evidentiary rather
than optimistic. What is measured is that the *Apple* toolchain path cannot be
closed; what is not measured is any Kotlin/Native host whose toolchain is not
Apple. Decision 0008 section 10 asks whether a Kotlin backend satisfies section
3, and section 1 answers that with a produced, executed, base-installation-only
binary. Retiring on macOS evidence would assert something about `windows/amd64`
and `linux/amd64` that the evidence does not support, in the same way that
claiming them would — a fabricated negative rather than a fabricated positive.
Retirement at Protocol 1.0 is also one-way: decision 0008 section 2 closes the
identifier space, so a retired Kotlin pair cannot be reopened under any name for
the whole protocol version. The trigger above converts the remaining uncertainty
into a dated, checkable outcome owned by a named task, which is what a
reservation is for.

### 13. Capability limitations, stated rather than discovered

**DECIDED.** These are consequences of the design, not defects, and the
authoring guide MUST state all five:

1. **No third-party dependencies.** A package cannot supply or name a `.klib`,
   and the driver passes no `-library`. A build root compiles against the
   distribution's own libraries and nothing else.
2. **No user-defined C interop.** `cinterop` is a second tool and a second
   command, which `manager-worker-v2`'s session admits in neither position, and
   `.def` files, headers and native sources are rejected by the allow-list. The
   distribution's own platform bindings remain available, subject to section 9's
   published-artifact gate.
3. **The build root is not an IDE project.** The allow-list excludes every
   Gradle file, so an author keeps the IDE project outside the build root. Build
   roots are context-excluded and never runtime-copied, so this costs the agent
   nothing and costs the author a duplicated source layout.
4. **No cross-compilation.** One host builds for its own default target only.
5. **No macOS.** Section 11. On macOS the driver fails closed; it does not
   silently fall back to anything.

### 14. Downstream obligations

- `TASK-260728-1koh5v`, `TASK-260728-gmfxdg` (Curator) and
  `TASK-260728-3ar1qp`, `TASK-260728-1uj0bc` (csk): implement against
  [`docs/kotlin-native-build-drivers.md`](../docs/kotlin-native-build-drivers.md),
  including the section 9 published-artifact dynamic dependency gate, and
  implement **no** platform as claimed until section 12 passes for its tuple.
- A Windows qualification run on `windows/amd64` owns A1 through A9 and MUST NOT
  fill a registry field by assertion.
- `TASK-260728-3u1nho`: Linux qualification, after `TASK-260728-1skseh`, with
  the identical acceptance test.
- `TASK-260728-r3j8ef` and `TASK-260728-1aveb2`: cross-manager interop
  verification for this pair cannot run on macOS under section 11 and must be
  sequenced onto the qualified host.
- `TASK-260728-251p01`: integrate only if at least one tuple qualified;
  otherwise apply section 12 and mint claim schema 4 without the two
  identifiers.
- `TASK-260728-2uh7em`: authoring and operations guidance, which MUST carry
  section 13 in full, MUST carry the bundle curation procedure of section 3, and
  MUST NOT present a Gradle project as a supported layout.
- `TASK-260728-1g0z69`'s reserved `kotlin` entry is completed by this record as
  far as section 10 states; its reserved `jdk` entry is **not** claimed by this
  record and remains reserved with no driver mapping.

## Stable failure classes

No new architecture-level class is introduced. Kotlin's per-surface diagnostics
sit beneath decision 0008's existing classes:

- `build_package_code_execution_forbidden` — the section 7 allow-list, with
  per-surface diagnostics `kotlin_response_file_name_forbidden`,
  `kotlin_script_source_forbidden`, `kotlin_build_system_file_forbidden`,
  `kotlin_native_interop_input_forbidden`, `kotlin_prebuilt_library_forbidden`,
  `kotlin_non_source_entry_forbidden`, and `kotlin_non_regular_entry_forbidden`;
- `build_artifact_class_unsupported` — the section 9 published-artifact gate, and
  a platform that cannot produce a single directly loadable file;
- `build_toolchain_metadata_mismatch` — `kotlin-native-module.json` shape or
  `kotlin_version` comparison;
- `build_toolchain_untested_release`, `build_toolchain_version_undetermined`,
  `build_toolchain_untrusted`, `platform_unsupported` — decision 0007, unchanged;
- `build_descriptor_driver_unsupported` and
  `build_descriptor_schema_unsupported` — decision 0008 section 5, unchanged.

## Rejected alternatives

- **The official Kotlin/Native archive as the toolchain root, with the JDK as a
  companion.** Rejected on measurement, and this is the alternative the first
  version of this record chose. The archive contains no regular executable, so
  decision 0007 section 3 has nothing to point at; its Apple dependency closure
  is not in the archive at all; and a companion root would be a second
  operator-asserted tree outside the primary fingerprint. The curated bundle
  satisfies the invariant literally instead of reading it narrowly.
- **A "narrow reading" of decision 0007 section 3 under which the pipeline's
  primary executable may live in a companion root.** Withdrawn. It was a request
  to reinterpret an accepted invariant in order to keep a design; the design
  changed instead.
- **Retiring both identifiers now, because the only measurable host cannot
  qualify.** Rejected: the measurement shows that the *Apple* toolchain path
  cannot be closed, and shows in the same file that the Windows and Linux
  toolchain paths are declared as ordinary dependencies rather than as host
  SDKs. Retiring would assert a negative about two tuples on evidence drawn from
  a third, and retirement is one-way for the whole protocol version. Section 12's
  trigger already converts the outcome into a dated, checkable decision owned by
  a named task.
- **Admitting macOS by treating `/usr/bin/xcrun`, `/bin/bash` and the Xcode
  toolchain as part of the platform's base installation.** Rejected on two
  independent grounds. Decision 0008 section 6 item 3 requires every executable
  started below the worker to be a fingerprinted member of the closure and names
  a host-resolved tool outside it as inadmissible; and the Xcode SDK would be an
  unfingerprinted build input, so two hosts with different Xcode versions would
  alias in the cache. The base-installation clause of section 3 is about the
  *artifact's* dynamic dependencies, not about the compiler's process graph.
- **Admitting macOS by copying the Xcode toolchain and SDK into the curated
  bundle.** Rejected on measurement: `/usr/bin/xcrun`, `/bin/bash`,
  `/usr/bin/xcode-select` and `/usr/libexec/PlistBuddy` are absolute paths
  compiled into the compiler, and with every Apple property overridden to point
  inside a curated tree the compiler still spawned `/usr/bin/xcrun` to resolve
  the toolchain. The copy would also redistribute Apple-licensed material, which
  is not a decision this record may take.
- **Using `KONAN_USE_INTERNAL_SERVER` to obtain the Apple parts as ordinary
  dependencies.** Rejected: it points at `https://repo.labs.intellij.net`, a
  JetBrains-internal host, so it is not an origin any operator can use, and
  depending on it would make the contract unimplementable outside one company.
- **Letting the manager hydrate the dependency closure itself on first use.**
  Rejected: it is a manager-initiated download of hundreds of megabytes of
  executable content, over a channel with no integrity check the compiler
  reports, inside a Curator operation — the auto-install decision 0007 refuses,
  under a different name. Hydration is an operator act, performed once, outside
  every operation, and its result is inside the fingerprint.
- **A manager-read bundle descriptor recording the bundle's provenance and
  component digests.** Rejected: the manager cannot verify such a file against
  anything, so it would be a trust input that looks like a proof. The tree
  digest is the proof of stability, and provenance stays where decision 0007
  puts it — with the operator, at configuration time.
- **Pointing `KONAN_DATA_DIR` at the fingerprinted bundle directly.** Rejected
  on measurement: the compiler opens `dependencies/cache/.lock` for writing, so
  a read-only root fails, and a writable root would let an operation mutate the
  tree whose digest is the toolchain identity.
- **Relying only on network denial for the no-download guarantee.** Rejected as
  the sole layer: it is an operator or platform property rather than a driver
  property, and it fails open on a host where denial is unavailable.
  `airplaneMode=true` makes the guarantee in-process and measurable, and the
  network denial is retained as the independent second layer in A4.
- **Describing the response-file defence as "absolute paths do not expand".**
  Rejected on measurement: `@/abs/path/inject` expands. The token's first
  character is what decides, and stating the wrong reason would let an
  implementer reproduce the conclusion incorrectly.
- **A deny-list of forbidden filenames instead of the allow-list.** Rejected:
  Kotlin's code-execution surfaces are an open set, so a deny-list cannot be
  exhaustive, which decision 0008 section 7 requires. The `@` vector is the
  proof — it is a name shape, not a known filename.
- **Rejecting source-level imports of the distribution's platform libraries.**
  Rejected: enforcing it would require the manager to parse Kotlin source, and
  the surface is not package-supplied — it is inside the fingerprinted bundle
  and cannot be extended by a package. The consequence that actually needs
  closing is the artifact's dynamic dependency set, and section 9 closes it at
  the point where it is observable.
- **Claiming "standard library only, no C interop", as the first version of this
  record did.** Rejected on measurement: `import platform.posix.*` and
  `import platform.Foundation.*` compile and link with no `-library` and no
  `.def`, and change the artifact's dynamic dependencies.
- **Passing `source_dir` to the compiler as a directory source root.** Rejected:
  it hands file discovery to the compiler, so files the allow-list walk rejected
  could still be compiled, and decision 0008 section 4 forbids the manager
  inferring the source set.
- **Admitting cross-compilation via a package-selected or driver-selected
  target.** Rejected: a package-selected target is a forbidden command member,
  and a driver-selected one has no naming authority and reintroduces the
  representable-but-unserved-target gate for no capability the artifact model
  needs.
- **Kotlin/JVM plus GraalVM `native-image`.** Rejected on four independent
  grounds, any one sufficient: it needs a second compile-phase command, which
  `manager-worker-v2`'s session shape refuses; it needs a third toolchain
  outside decision 0007's closed identifier set; assembling its classpath needs
  either a package manager or a package-controlled classpath member, both
  refused; and it is a different backend, so decision 0008 section 2 requires a
  different family segment.
- **Gradle in any role, including a metadata-only query.** Rejected on
  measurement: `gradle properties` compiles the build script as a program source
  unit before answering. There is no read-only Gradle, and no `--offline`,
  `--dry-run`, or init-script variant changes what it is.
- **Maven as the metadata file, reading only one field.** Rejected: the Maven
  project model is its plugin graph; reading one field while ignoring the graph
  would let a package ship a plugin declaration Curator silently does not run.
- **Naming a `compatibility` family or a `platforms` pair now.** Rejected:
  decision 0007 section 1.1.1 admits a family only after testing against the
  driver's conformance vectors, which do not exist.

## Compatibility impact

None on the wire. This record admits no identifier, mints no schema, and moves
no frozen byte. `go-v1` and `go-repository-v1` identities, `manager-worker-v1`,
`capability-evidence-v1`, `curator-go-toolchain-v1`, the rc.4 and rc.5
conformance corpora, and the rc.5 pin are untouched. Manifest schemas 1 through
7, descriptor schema 1, receipt schemas 1 and 2, marker schemas 1 through 3, and
claim schemas 1 through 3 continue to reject both Kotlin identifiers, as
decision 0008 section 11 item 12 requires.

`kotlin-native-module.json` is a new package-authored file bound by a reserved
driver. It reaches no wire surface until manifest schema 8 is minted with
`kotlin-native-v1` admitted, and it is inert for every other driver.

Decision 0007's reserved `kotlin` entry moves from "no fields" to "shape,
metadata source, baseline, probe, normalization and identity fixed;
`primary_relpath` installed per operating system on qualification;
`compatibility` and `platforms` outstanding". Its reserved `jdk` entry is
**released**: this record does not use it, so it stays reserved with no driver
mapping and no obligation on this task. That supersedes both the "`jdk` if JVM"
expectation in decision 0007's driver-mapping table and the first version of
this record, which made `jdk` a REQUIRED companion under a native model; the JDK
is now a component of the `kotlin` root rather than a toolchain of its own.

## Security impact

Net positive, and concentrated in five places.

The curated bundle makes the entire executable closure one fingerprinted tree.
Under the previous shape the JDK was a separate operator-asserted root and the
LLVM and sysroot closure was outside every fingerprint; now a single tree digest
covers the launcher, the compiler, and every tool the compiler spawns, and the
cache key binds it.

Refusing the shipped launcher removes the `PATH` resolution of `java` and six
environment-selected pipeline inputs (`JAVACMD`, `JAVA_HOME`, `JAVA_OPTS`,
`KOTLIN_RUNNER`, `KOTLIN_COMPILER`, `KOTLIN_TOOL`) from the process graph
structurally rather than by neutralisation. `KOTLIN_COMPILER` in particular
selects the compiler main class — an environment-selected backend.

Hydration is moved out of the operation entirely. The measured download path has
no integrity check the compiler reports, so performing it inside an operation
would be a manager-initiated fetch and execution of hundreds of megabytes of
third-party executable content — the largest untrusted action the protocol
excludes. `airplaneMode=true` makes the exclusion enforceable in-process rather
than by hoping the network is unreachable.

The measured `@`-response-file vector is a real property of the Kotlin CLI on
both backends: a token whose first character is `@` is a response file, and a
missing one is only a warning, so a partial mitigation fails silently. The
contract closes it twice, and the corrected statement of *why* the argv
discipline works is what keeps the second layer from being reproduced wrongly.

The macOS exclusion is the largest security effect and it is a refusal rather
than a mitigation. A Kotlin build integration that runs on macOS necessarily
executes at least seven host binaries the manager cannot fingerprint, and binds
an Xcode SDK into the output that the cache key cannot see. Both are refused,
the platform is excluded, and nothing compensates for it.

Two exposures are recorded rather than closed. The compiler-input exposure of
decision 0008's security section is unchanged: a trusted compiler parsing
adversarial source under the portable, non-hardened controls is the same
exposure the other drivers carry. And the artifact's dynamic dependency set is
package-source-dependent through the distribution's platform libraries; section
9's published-artifact gate is what converts that from an unbounded property
into a checked one, and a tuple without a base-installation allow-list cannot be
qualified.
