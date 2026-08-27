# Kotlin build drivers: `kotlin-native-v1` and `kotlin-native-repository-v1`

Implementation-ready reference for the pair selected by
[decision 0010](../decisions/0010-kotlin-native-driver-pair.md), inside the
boundary of [decision 0008](../decisions/0008-additional-language-driver-boundary.md)
and the toolchain contract of
[decision 0007](../decisions/0007-compiled-build-toolchain-preflight.md) and
[`compiled-build-toolchain-requirements.md`](compiled-build-toolchain-requirements.md).

**Both identifiers are reserved, not admitted, and `platforms` is empty.**
Nothing in this document is a platform claim.

Evidence markers: **MEASURED** — recorded with argv and real exit code in
`TASK-260728-168smo_command-evidence.log`, taken on macOS 26.5 arm64 with
OpenJDK 26.0.1 against the checksum-verified official release
`kotlin-native-prebuilt-macos-aarch64-2.4.10.tar.gz`
(`55ded039bb56a69aec9df354a92b42df9e916104e3c53d8d9852d9cc6617ed9d`, equal to
the published `.sha256` asset). **UNVERIFIED** — a named per-tuple obligation
carried into section 11. Absent marker — a consequence of an accepted decision.

macOS is measured **unsupported** for this pair (section 11.2). Every macOS
measurement below therefore characterises the *pipeline*, so that a Windows or
Linux qualification is mechanical rather than a second design pass; none of it
admits macOS.

## 1. The toolchain root: `curator-kotlin-bundle-v1`

### 1.1 Why the vendor archive is not the root

**MEASURED.** The official distribution contains no regular executable:

| Path | Type |
|---|---|
| `bin/konanc`, `bin/kotlinc-native`, `bin/run_konan`, `bin/cinterop`, `bin/klib`, `bin/konan-lldb`, `bin/generate-platform` | Bourne-Again shell scripts |
| `konan/lib/kotlin-native-compiler-embeddable.jar` | the compiler, 83,901,273 B |

`run_konan` takes `JAVACMD`, `JAVA_HOME` and `JAVA_OPTS` from the environment,
falls back to resolving the bare name `java` through `PATH`, and appends every
`-D` and `-J` argument to the JVM. Decision 0007 section 3 requires a regular
executable at a fixed relpath inside the fingerprinted tree, and decision 0008
section 6 item 3 forbids environment-selected pipeline inputs, so neither the
archive root nor any script inside it can be the primary executable.

### 1.2 The bundle layout

The `kotlin` root is an operator-curated tree resolved through decision 0007's
second admissible origin — trusted operator configuration in manager-owned,
owner-protected state — with this exact layout:

```text
<kotlin_root>/
  jdk/
    bin/java                       primary_relpath (Unix)
    bin/java.exe                   primary_relpath (Windows)
  kotlin-native/
    konan/lib/kotlin-native-compiler-embeddable.jar
    konan/konan.properties
    klib/…
    bin/…                          present, never executed
  konan-data/
    dependencies/<name>/…          the prehydrated closure
    dependencies/.extracted
```

The whole tree is fingerprinted, is immutable for the life of the configuration,
and MUST be read-only to the account the manager runs as. The manager reads no
descriptor inside it; the tree digest is the identity.

### 1.3 Curation procedure (operator, once, outside every operation)

1. Download the official `kotlin-native-prebuilt-<platform>-<version>` archive
   and verify it against the release's published `.sha256` asset.
2. Unpack into `<kotlin_root>/kotlin-native`.
3. Read `kotlin-native/konan/konan.properties` and confirm that for the target
   this host will build, none of `targetToolchain.<target>`,
   `targetSysRoot.<target>` or `additionalToolsDir.<target>` resolves to a
   `remote:internal` dependency. **MEASURED**: every Apple target does resolve to
   `remote:internal`, which is why macOS is unsupported (section 11.2);
   `linux_x64` and `mingw_x64` do not.
4. Hydrate once, with network access, on this platform: run one throwaway
   `-produce program` compile with `KONAN_DATA_DIR` pointed at
   `<kotlin_root>/konan-data`. **MEASURED** on macOS arm64: this fetches
   `lldb-4-macos` (64,230,999 B), `llvm-21-aarch64-macos-essentials-97`
   (151,150,049 B) and `libffi-3.3-1-macos-arm64` (17,037 B) from
   `download.jetbrains.com` and extracts 688 MB. The compiler reports no
   integrity check for these downloads; verifying them is part of curation, not
   of any Curator operation.
5. Place a JDK at `<kotlin_root>/jdk`. Any JDK the compiler runs on is
   admissible; its identity is covered by the bundle digest, not by a separate
   probe.
6. Make the whole tree read-only, then register it in operator configuration.

Nothing in this procedure is performed by a manager, in any mode, at any time.

### 1.4 The `kotlin` registry entry

| Field | Value |
|---|---|
| `toolchain_id` | `kotlin` |
| `fingerprint_algorithm` | `curator-kotlin-toolchain-v1` |
| companions | **empty** |
| `metadata_sources` | `kotlin-native-module.json` → `kotlin_version` |
| `baseline` | `at_least 2.4.10` |
| `compatibility` | **empty**; family granularity `(major, minor)` |
| `platforms` | **empty** |
| `primary_relpath` | `jdk/bin/java`, `jdk/bin/java.exe` on Windows — declared per operating system, and only once that operating system is in `platforms` |
| `probe` | section 1.5 |
| `normalization` | `kotlin.konanc.dashversion.stdout`, section 1.5 |

The `jdk` identifier decision 0007 reserved is **not used**: the JDK is a
component of the `kotlin` root, so there is no companion entry, no second
probe, and no second tree digest. `toolchain_identities` in the canonical build
input is therefore a one-element ordered array.

Decision 0007 section 4 declares `primary_relpath` and `probe` per operating
system and only for operating systems in `platforms`, and its release gate
rejects a relpath declared outside that set. While `platforms` is empty the
entry structurally cannot carry either; the values above are what a
qualification installs, not registry content today.

### 1.5 Probe and normalization

**MEASURED**, exit 0:

```text
<kotlin_root>/jdk/bin/java
  -ea -Xmx3G -XX:TieredStopAtLevel=1
  -Dfile.encoding=UTF-8 -Duser.language=en -Duser.country=US
  -Dkonan.home=<kotlin_root>/kotlin-native
  -cp <kotlin_root>/kotlin-native/konan/lib/kotlin-native-compiler-embeddable.jar
  org.jetbrains.kotlin.cli.utilities.MainKt konanc -version
```

| Stream | Content |
|---|---|
| stdout | `Kotlin/Native: 2.4.10` — 22 bytes, one line |
| stderr | four JVM "restricted method" warnings, then `info: kotlinc-native 2.4.10 (JRE 26.0.1)` |

The probe runs from a manager-owned empty working directory under the section 5
environment. Normalization `kotlin.konanc.dashversion.stdout` reads line 1 of
**stdout**, bounded to the first 4 KiB, anchored whole-line:

```text
^Kotlin/Native: (0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(\S*)(?:\s.*)?$
```

Groups 1–3 are the triple; a non-empty group 4 sets the prerelease flag, and a
prerelease host never satisfies any requirement. Output that the rule does not
match, or matches more than once, is `build_toolchain_version_undetermined` —
never a default.

The literal `Kotlin/Native: ` prefix asserts the **backend**. A Kotlin/JVM
distribution writes `info: kotlinc-jvm <version>` to stderr with an empty
stdout, so it cannot satisfy this rule at all, and a misconfigured root that
holds the wrong distribution fails the probe rather than reaching a compile that
cannot produce `native-executable-v1`.

Resolution origins are decision 0007 section 3's two and only two. The ambient
or user `PATH`, `JAVA_HOME`, `KOTLIN_HOME`, `KONAN_DATA_DIR`, `SDKMAN_DIR`, a
runtime root, project `.agents/bin`, a shim, a manifest or descriptor value, an
inherited environment variable, and any version-manager wrapper are all
forbidden origins.

## 2. Identity

`curator-kotlin-toolchain-v1`. A resolved identity is the algorithm identifier,
the normalized native version string, the primary-executable relpath, and the
tree digest of `<kotlin_root>`. Toolchain location is not portable identity.

Because the JDK, the compiler and the whole dependency closure are inside that
one tree, the digest covers every executable the pipeline can start. There is no
second root, no cross-root closure hash, and no component the cache key cannot
see.

Fingerprinting proves that the tree is stable across an operation and identical
across operations. It does **not** prove upstream authenticity: the download in
section 1.3 step 4 carries no integrity check the compiler reports, so
verification is the operator's responsibility at curation time. A
`content_sha256` in a receipt MUST NOT be read as provenance.

## 3. The trusted launcher and the process graph

```text
manager parent
  -> identity-verified manager-owned worker
       -> <kotlin_root>/jdk/bin/java
            -> Kotlin/Native compiler, in-process, loaded by -cp from
               <kotlin_root>/kotlin-native
                 -> regular executables inside <kotlin_root>
                    (LLVM clang++, assembler, archiver, linker)
```

**Normative.** The driver MUST NOT execute `bin/konanc`, `bin/kotlinc-native`,
`bin/run_konan`, `bin/kotlinc`, `bin/cinterop`, `bin/klib`, `bin/konan-lldb`,
`bin/generate-platform`, or any other launcher script from any distribution, on
any platform, including for the version probe.

**MEASURED** for the in-closure half: the LLVM driver the compiler spawns is
`<konan-data>/dependencies/llvm-21-aarch64-macos-essentials-97/bin/clang++`.

**Per-tuple obligation K-2.** Every executable the compiler spawns must resolve
inside `<kotlin_root>` or the operation-private overlay of section 5.2. Any
other executable fails that tuple; there is no toolchain identifier for a
platform SDK that contributes a process, and Rust's data-only SDK root is not a
precedent for an executed linker. **MEASURED on macOS: this obligation fails,
and section 11.2 is the consequence.**

## 4. Source layout and the module file

### 4.1 Local mode

`build_roots` is the schema-6/7 model, unchanged: a portable relative path other
than `.`, a real link-free directory in the immutable raw snapshot, unique and
pairwise disjoint, never equal to or nested with a runtime root, referenced by
at least one build command, statically excluded from agent context and from the
runtime copy before cache lookup, identically for real builds, exact cache hits,
and dry-runs.

The build root MUST contain `kotlin-native-module.json` **directly**, and that
file MUST be the nearest ancestor of `source_dir`. The manager does not search
for it, does not walk upward, and does not infer it.

### 4.2 External mode

`skill-build.json` schema 2, target `driver: "kotlin-native-repository-v1"`,
with `build_root`, `source_dir` and the OPTIONAL `toolchain`. The command and
descriptor drivers MUST be equal. The whole repository snapshot is the
validation, identity and audit subject; only the selected build root is
compiler-visible; no external repository byte is agent-facing or runtime-copied.
`kotlin-native-module.json` MUST exist directly in the descriptor's
`build_root`. Against a schema-1 descriptor the command fails
`build_descriptor_driver_unsupported`, with no fallback.

### 4.3 `kotlin-native-module.json`

```json
{"schema_version": 1, "kotlin_version": "2.4.10"}
```

Exactly two members, both REQUIRED, `additionalProperties: false`.

| Member | Type | Meaning |
|---|---|---|
| `schema_version` | `const` integer `1` | file-shape gate only |
| `kotlin_version` | canonical `major.minor.patch` | the Kotlin compiler version the sources are written against |

`kotlin_version` matches decision 0007 section 2.1's grammar exactly:
`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`, each component at most
`999999`, no prefix, prerelease, build metadata, leading zero, or wildcard.

The file is never passed to the compiler, contributes no argument, and selects
nothing.

### 4.4 `source_dir` to program

The program is the recursive `.kt` set under `source_dir`, compiled in `program`
mode with the compiler's **default** entry point. The manager never names,
searches for, or infers an entry point. Zero or multiple entry points is a
compiler error and is therefore deterministic. No new command member is
introduced, and the consuming manifest command key remains the sole naming
authority for the artifact.

## 5. Environment and the operation-private overlay

### 5.1 Environment

On top of the `manager-worker-v2` portable control set and the mandatory
portable controls of `protocol/core.md` section 4.2.1:

| Variable | Action | Reason |
|---|---|---|
| `KONAN_DATA_DIR` | set to the overlay of 5.2 | the bundle is read-only; the compiler needs a writable data dir |
| `TMPDIR`, `TMP`, `TEMP` | set to operation-private staging | **MEASURED**: intermediates land in `$TMPDIR/konan_temp<random>/` |
| `KONAN_USE_INTERNAL_SERVER` | unset | **MEASURED**: selects the dependency host `https://repo.labs.intellij.net/kotlin-native` |
| `JDK_JAVA_OPTIONS` | unset | honoured by the JVM launcher; injects arbitrary JVM options |
| `JAVA_TOOL_OPTIONS` | unset | same |
| `_JAVA_OPTIONS` | unset | same family |
| `CLASSPATH` | unset | would extend the compiler classpath |
| `JAVA_HOME`, `JAVACMD`, `JAVA_OPTS` | unset | `run_konan` inputs; closed structurally by never running it, unset as defence in depth |
| `KOTLIN_HOME`, `KOTLIN_COMPILER`, `KOTLIN_TOOL`, `KOTLIN_RUNNER` | unset | same; `KOTLIN_COMPILER` selects the compiler main class |
| `LIBCLANG_DISABLE_CRASH_RECOVERY` | unset | set only by the launcher, which is never run |
| `PATH` | manager-owned minimal value | no toolchain is resolved through it |
| `LANG`, `LC_ALL` | `C`/`POSIX` per the portable policy | locale-independent diagnostics |

No manager-written configuration file is placed anywhere for this driver. Rust
needed one because Cargo discovers ancestor `.cargo/config.toml`; this pipeline
reads no configuration file, because the launcher that would is never run and
the compiler is given an explicit, complete argument vector.

### 5.2 The overlay

**MEASURED, three facts that together fix this section.**

| Run | Outcome |
|---|---|
| hydrated `KONAN_DATA_DIR`, compile | exit 0; the tree is byte-identical before and after (aggregate SHA-256 over 1,470 files unchanged) |
| the same tree made read-only, compile | exit 2 — `FileNotFoundException: …/dependencies/cache/.lock (Permission denied)` |
| bundle read-only, `KONAN_DATA_DIR` = overlay | exit 0, artifact produced, both read-only roots unchanged |

So the closure is not mutated by a compile, but the data directory must be
writable, and an operation-private overlay satisfies both.

**Normative.** For each operation the manager materialises a private writable
directory whose `dependencies/` holds one entry per dependency present in
`<kotlin_root>/konan-data/dependencies`, a copy of `.extracted`, and a fresh
empty writable `cache/`. The materialisation mechanism MUST copy or link only,
MUST add no entry that is not in the bundle, MUST reach no network, and MUST
leave `<kotlin_root>` byte-unchanged. The overlay is never fingerprinted, never
published, and is removed with the operation. The exact mechanism is per
platform and is fixed by that platform's qualification — a symlink farm is
**MEASURED** on macOS and is not universally available on Windows.

### 5.3 No download, twice

The compile vector carries `-Xoverride-konan-properties=airplaneMode=true`, a
manager-owned constant no package byte can reach. **MEASURED**: with it, a
dependency missing from the data directory is
`Cannot find a dependency locally: <name>` and exit 2, with no download
attempted. **MEASURED** independently: with the closure hydrated and every
network operation denied, the compile exits 0 and logs no download line.

Both layers are required. The override is a driver property that holds on any
host; network denial is an operator or platform property that may be
unavailable. Section 11.1 A4 requires both.

## 6. The worker session and the argument vector

**Session.** Zero graph-phase commands, exactly one compile-phase command.
Decision 0008 section 7's "at most one" graph phase admits this. No retry, no
second phase, no daemon, no additional executable, no shell, no VCS operation,
no dependency download, no generator, no test, no run, no tool request.

The Kotlin compile daemon is **forbidden**: `kotlin-daemon.jar` and
`kotlin-daemon-client.jar` ship in the distribution, and a daemon is a
persistent process outside the session shape. The driver passes no daemon
argument and MUST fail rather than fall back to one.

**Compile vector.**

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
  -o <staging>/<command>
  <abs source 1> … <abs source N>
```

Binding rules:

1. The JVM options are `run_konan`'s own, minus everything it takes from the
   environment. `-Duser.language`/`-Duser.country` are the one element added
   beyond the measured vector, for locale determinism; qualification re-measures
   the vector including them.
2. **Sources are enumerated by the manager**, recursively under `source_dir`,
   sorted by Unicode-scalar order of relative path, and passed as **absolute**
   paths. A directory is never passed as a source root: that hands file
   discovery to the compiler and would compile files the section 7 walk
   rejected.
3. **No argv token may begin with `@`** — see 7.1; this is what makes the
   absolute-path form a structural backstop rather than a convention.
4. No entry point, module name, opt-in marker, `-language-version`,
   `-api-version`, optimisation, debug, cache, memory-model, bitcode, library,
   or plugin argument is passed, and none is package-derivable.
5. `-target` names the host's own default native target only.
   **Cross-compilation is not admitted.** **MEASURED**: `konanc -list-targets`
   marks exactly one line `(default)` — on this host `macos_arm64 (default)` —
   which is obligation K-6's source of truth for a tuple.
6. **MEASURED, obligation K-4 for this tuple**: `-o <staging>/app` produces
   `app.kexe` plus an `app.kexe.dSYM/` by-product directory. The suffix is
   compiler-applied; the rename to the published name happens **inside
   operation-private staging only**, before hashing. K-4 is re-measured per
   tuple because the suffix is platform-dependent.

## 7. Pre-compile rejection matrix

Decision 0008 section 7 requires an exhaustive, deterministic, pre-compile
rejection matrix. Kotlin's code-execution surfaces are an open set — a
general-purpose script engine, a script dialect the compiler can execute,
annotation processing, compiler plugins, C interop — so the matrix is a **closed
allow-list**, computed by a manager-side walk of the validated snapshot that
runs no compiler and reaches no network.

### 7.1 The allow-list

| # | Admitted | Pattern |
|---|---|---|
| A1 | directories | name matches `^[A-Za-z0-9_][A-Za-z0-9_.-]*$` |
| A2 | Kotlin sources | name matches `^[A-Za-z0-9_][A-Za-z0-9_.-]*\.kt$`, regular file |
| A3 | the module file | exactly `kotlin-native-module.json`, regular file, **directly** in the build root and nowhere else |

Anything else is `build_package_code_execution_forbidden`. The walk is total: it
classifies every entry it encounters, follows no symlink, and has no depth,
count, or size exemption that would leave an entry unclassified.

The leading-character class excludes every dotfile **and every name beginning
with `@`**.

**MEASURED — the response-file channel, and the corrected reason.** On the
native backend the compiler strips a leading `@` from an argv token and reads
the remainder as a response-file path:

| Argument token | File on disk | Expanded | Outcome |
|---|---|---|---|
| `@inject` | `inject` (containing `-version`) | yes | exit 0, `Kotlin/Native: 2.4.10` on stdout |
| `@./inject` | `inject` | yes | exit 0, same |
| `@/abs/path/inject` | `inject` | **yes** | exit 0, same |
| `/abs/path/@inject` | `@inject` | no | `error: source entry is not a Kotlin file`, exit 1 |
| `./@inject` | `@inject` | no | same, exit 1 |
| `@nonexistent` | — | n/a | `warning: argfile not found`, warning only |

Expansion is therefore decided by the **first character of the argv token**, not
by whether the path is absolute. An absolute package path is safe because it
starts with `/`, and a driver that ever prefixed a path with `@` would reopen
the surface even in absolute form. Because a missing response file is only a
warning, a partial mitigation fails silently.

Two independent layers close it, and both are required:

- **normative** — A1/A2/A3 reject the name before the compile phase, which is
  what decision 0008 section 7 demands;
- **structural backstop** — rule 3 of section 6: the driver never emits a token
  whose first character is `@`.

### 7.2 Surfaces subsumed, with per-surface diagnostics

Each row is a diagnostic beneath `build_package_code_execution_forbidden`, not a
separate check; the allow-list is what rejects them.

| Surface | Example entries | Diagnostic |
|---|---|---|
| response-file name | `@anything` | `kotlin_response_file_name_forbidden` |
| Kotlin script source | `*.kts`, `build.gradle.kts`, `settings.gradle.kts` | `kotlin_script_source_forbidden` |
| build-system file | `build.gradle`, `settings.gradle`, `gradle.properties`, `gradlew`, `gradlew.bat`, `gradle/`, `.gradle/`, `pom.xml`, `.mvn/`, `mvnw` | `kotlin_build_system_file_forbidden` |
| annotation processing / plugin config | kapt or KSP directories, `META-INF/services`, `*.pro` | `kotlin_build_system_file_forbidden` |
| C interop input | `*.def`, `*.h`, `*.c`, `*.m`, `*.mm`, `*.cpp`, `*.s`, `*.S`, `*.o`, `*.a`, `*.so`, `*.dylib`, `*.dll`, `*.lib` | `kotlin_native_interop_input_forbidden` |
| prebuilt library | `*.klib`, `*.jar`, `*.aar` | `kotlin_prebuilt_library_forbidden` |
| any other file | anything not A2 or A3 | `kotlin_non_source_entry_forbidden` |
| non-regular entry | symlink, device, socket, FIFO | `kotlin_non_regular_entry_forbidden` |

`.kts` is rejected as a build-model surface rather than described as immediate
execution. **MEASURED** on the JVM backend: a `.kts` file reaching the compiler
*without* `-script` is compiled into a class and not executed. It is inert
alone, and it is the payload half of the `@` vector; the allow-list rejects it
either way.

### 7.3 Surfaces closed structurally rather than by the walk

Argv-only, and unreachable because the closed command surface gives a package no
way to supply an argument and the driver's vector is fixed. They are listed so
the matrix is exhaustive over *surfaces*, and so that any future change letting
package data reach argv is visibly a change to this section.

| Surface | Flag | Why unreachable |
|---|---|---|
| compiler plugin by classpath | `-Xplugin=<path>` | not in the vector; no package-controlled argv |
| compiler plugin, new form | `-Xcompiler-plugin=…`, `-Xcompiler-plugin-order=…` | same |
| plugin option | `-P plugin:<id>:<opt>=<v>` | same |
| script templates | `-script-templates`, `-Xdefault-script-extension`, `-Xscript-resolver-environment` | same |
| explicit script execution | `-script`, `-expression`/`-e` | same, **and** reachable only through the `@` vector 7.1 rejects |
| toolchain re-pointing | `-Xkonan-data-dir`, further `-Xoverride-konan-properties` keys | same; the driver's own single override is a manager-owned constant |
| JVM passthrough | `-J…`, `-D…` | `run_konan` features; it is never run |

### 7.4 Network

No dependency resolution exists to perform: section 4 admits no dependency
declaration and 7.1 admits no prebuilt library, so there is nothing to fetch.
Section 5.3 closes the compiler's own dependency fetch twice.

## 8. Stage B — metadata disposition

Runs after local snapshot validation, or after exact external acquisition and
audit, and before any artifact-cache candidate is read or any compiler child is
started. Ordered steps per decision 0007 section 4: recompute the effective
requirement now that the descriptor requirement is readable, re-evaluate the
resolved version against the narrowed interval, gate on file shape, then
evaluate dispositions.

### 8.1 File-shape gate

`build_toolchain_metadata_mismatch`, before cache lookup, if any of:

1. `kotlin-native-module.json` is absent from the build root, is not a regular
   file, or is present anywhere other than directly in the build root;
2. it is not well-formed JSON, or is not a JSON object;
3. `schema_version` is absent or is not the integer `1`;
4. `kotlin_version` is absent;
5. any member other than those two is present;
6. a member is duplicated.

### 8.2 Disposition table

| File | Field | Disposition | Rule |
|---|---|---|---|
| `kotlin-native-module.json` | `kotlin_version` | `compared` | canonical triple; strictly above the resolved compiler triple ⇒ `build_toolchain_metadata_mismatch` |
| `kotlin-native-module.json` | `schema_version` | file-shape gate only | not a metadata source |

There is **no `forbidden` class**, because the field's value space is a version
and nothing else — no spelling of it names *where* a toolchain comes from.

### 8.3 The classifier is two classes

| Class | Condition | Disposition |
|---|---|---|
| C1 | matches the canonical triple grammar exactly | `compared` |
| C2 | anything else | `build_toolchain_metadata_mismatch` (8.1 step 4 or the grammar) |

Total by construction, with C2 as the mandatory catch-all. Because Curator owns
the grammar there is no document layer, no ecosystem grammar layer, and no
edition floor for the layers to be independent of. Decision 0007's alignment
properties reduce accordingly: the security partition `F` is empty, so P1
(`C ⊆ Upstream`) and P2 (`Upstream \ F ⊆ C`) collapse to the satisfiable
equality `C = Upstream`, where `Upstream` is Curator's own canonical grammar.
Both hold trivially and are checked as such. There is consequently no ecosystem
boundary probe to extend for Kotlin: decision 0007's obligation to measure two
independent acceptance layers is satisfied vacuously, because the ecosystem
supplies neither layer. The host-version gate decision 0007 requires each
ecosystem to name is `kotlin_version` versus the resolved triple in 8.2, and it
is deliberately outside the grammar.

## 9. Identity, cache, receipt, marker, claim

`curator-build-source-v1` is reused unchanged for both source modes. The
protected external snapshot key of `protocol/core.md` section 9.4 is unchanged.

The logical cache key is the SHA-256 of `CCJ-1` over the complete build input,
which binds:

| Input | Local | External |
|---|---|---|
| `receipt_schema_version` | `3` | `4` |
| `driver` | `const kotlin-native-v1` | `const kotlin-native-repository-v1` |
| source state | `curator-build-source-v1` over the raw snapshot | repository snapshot identity per decision 0005 |
| consuming command name | yes | yes |
| build root and `source_dir` | yes | yes |
| native target | resolved default target | same |
| `toolchain_identities` | ordered one-element array `[kotlin]` | same |
| policy object | closed, below | same |

Policy object, closed to exactly two members:

```json
{"execution_policy": "manager-worker-v2", "compile_profile": "kotlin-native-program-v1"}
```

`execution_policy` is the `const` decision 0008 section 2's closed table binds to
these identifiers. `compile_profile` is a `const` naming the session shape, the
source-enumeration rule, the fixed argument vector including the
`airplaneMode=true` override, and the default-entry-point mapping together, so
that a future semantic change to any of them cannot happen without a
cache-identity change.

The effective toolchain requirement and the `compatibility` set are gates, not
build inputs, so the wire `toolchain` object never enters a cache key, receipt,
marker, or claim. What enters is the resolved identity.

Install marker v4 records `driver`, `receipt_schema_version` and
`execution_policy` per entry; a reader validates both recorded values against
decision 0008 section 2's closed tables and rejects a mismatch rather than
inferring. Top-level `build_source` is REQUIRED exactly when at least one active
local build command of any admitted local driver exists.

Claim schema 4 asserts these identifiers **only if** section 11 admitted at
least one tuple. Otherwise decision 0010 section 12 applies and they are
unassertable.

## 10. Artifact

`native-executable-v1`, and nothing else:

- exactly one bounded regular file, produced into operation-private staging,
  hashed there, published immutably under the manager-home mutation lock;
- named solely by the manager from the consuming manifest command key, as
  `bin/<command>` on Unix and `bin/<command>.exe` on Windows;
- directly executable by the host program loader using only base-installation
  libraries;
- never executed by the manager during validation, installation, status, repair,
  rollback, or garbage collection.

By-products stay in staging and are discarded with it: **MEASURED**, the
`*.dSYM` bundle the compiler emits beside the executable, `$TMPDIR/konan_temp*`
intermediates, `.klib` intermediates, and the compiler cache. None enters cache
identity, the receipt, the marker, the shim relationship, or publication.

Bit-reproducibility is not required. A linker-applied ad-hoc signature is
compiler output, not a manager signing step, and must be produced by the fixed
vector without selecting a signing identity, credential, or network interaction.

### 10.1 Platform libraries and the published-artifact dynamic dependency gate

**MEASURED.** The distribution ships 200-plus platform klibs per target under
`klib/platform/<target>/`. Source importing them compiles and links with no
`-library` and no `.def`:

```kotlin
import platform.posix.getpid
import platform.Foundation.NSProcessInfo
```

exits 0 and the produced binary gains `/usr/lib/libresolv.9.dylib`, a dynamic
dependency the same program without those imports does not have.

**Policy.** The fixed, distribution-owned platform library surface is
**allowed**: it is inside the fingerprinted bundle, a package cannot extend it
or name one, and rejecting it would require the manager to parse Kotlin source.
Everything a package could *supply* remains rejected by 7.1 — `.def` files,
headers, native sources, objects, archives, shared libraries, prebuilt `.klib` —
and `cinterop` is a second tool the session admits in neither position. The
accurate capability statement is **no user-defined C interop**, not "no C
interop".

**Normative gate.** Because the artifact's dynamic dependency set is a function
of package source, decision 0008 section 3's base-installation clause cannot be
discharged by any pre-compile walk. In operation-private staging, before hashing
and before publication, the manager MUST read the produced file's dynamic
dependency list and MUST fail with `build_artifact_class_unsupported` if any
entry is outside the closed base-installation library allow-list the platform
matrix fixes for that tuple. Reading a file's headers is not executing it.

Each qualified tuple MUST supply that allow-list; a tuple with no allow-list
cannot be qualified. **MEASURED** on macos/arm64 for reference only — the tuple
is unsupported — the observed set is `libSystem.B.dylib`, `libc++.1.dylib`,
`libobjc.A.dylib`, `libresolv.9.dylib`, `Foundation` and `CoreFoundation`.

### 10.2 Signing and credential boundary

Unchanged from decision 0008 section 9, restated because it is easy to reopen:

- neither driver performs manager post-build signing, timestamping, or
  notarization;
- no manifest, descriptor, module file, or repository byte may select a signing
  identity, certificate, entitlement, provisioning profile, keychain, or
  notarization credential — no such field exists, and none may be added;
- `codesign`, `productsign`, `notarytool`, `stapler` and `signtool.exe` MUST NOT
  appear anywhere in the process graph;
- a platform policy that requires a locally signed binary MUST reject the build
  until the separately versioned and reviewed signer profile of
  `protocol/core.md` section 12.2 exists. It MUST NOT be answered by a
  self-signed identity, an ad-hoc `codesign` invocation, or a
  quarantine-attribute removal;
- credentials, host-verification state, transport executables, proxy policy,
  timeouts and authentication modes stay operator-owned and MUST NOT appear in a
  manifest, descriptor, repository, module file, compiler environment, receipt
  trust field, or marker.

## 11. Platform matrix and qualification

`platforms` is **empty** for both identifiers. No tuple is claimed.

### 11.1 The acceptance test

A tuple is admitted only when a run on that exact tuple, on an immutable native
host, produces all of A1–A9 with recorded argv and real exit codes.

| # | Requirement |
|---|---|
| A1 | The bundle is curated per section 1.3 from a checksum-verified official archive; `primary_relpath` resolves to a regular executable; the probe and the anchored `Kotlin/Native: ` normalization of 1.5 are reproduced on that host |
| A2 | A shim-`PATH` run of the full compile records **0** `PATH` resolutions, **and** the paired control through the shipped launcher records a non-zero count. Without the control firing the zero proves nothing |
| A3 | Under exec containment allowing only `<kotlin_root>` and the overlay, the compile exits 0; every spawned child is recorded by resolved absolute path and lies inside those two trees; the containment control is shown to fire on a deliberately excluded path — obligation K-2 |
| A4 | With `airplaneMode=true` **and** all network denied, the compile exits 0, logs no download, and `<kotlin_root>` is byte-identical before and after; every write lands in operation-private state — obligation K-3 |
| A5 | Exactly one file is produced for publication; its exact name and compiler-applied suffix are recorded; renaming happens inside staging only; every by-product stays in staging — obligation K-4 |
| A6 | The published file is a native executable for the tuple, runs, and every entry of its dynamic dependency list is inside that tuple's closed base-installation allow-list, which this run supplies. The sample MUST include a source importing distribution platform libraries, because that is what moves the list — section 10.1 |
| A7 | The default native target is read from `konanc -list-targets` and mapped to `(operating_system, architecture)` — obligation K-6 |
| A8 | 7.1 is exercised with at least one positive and one rejected case per row of 7.2, including a name beginning with `@`, a `.kts` file, and a `.klib` |
| A9 | `compatibility` gains the tested family, and only that family, after the driver's conformance vectors pass against it |

### 11.2 macOS — measured unsupported

`macos/arm64` and `macos/amd64` are **not admissible**, on two independent
grounds, either one sufficient.

**Process closure.** **MEASURED**: under exec containment allowing only the
bundle roots, a hello-world `-produce program -target macos_arm64` fails at
`CurrentXcode.bash(Xcode.kt:144)`. Iterative enumeration yields the complete
external set, after which the compile exits 0:

```text
/bin/bash
/usr/bin/xcode-select                        (via bash)
/usr/libexec/PlistBuddy                      (via bash)
/usr/bin/xcrun
<Xcode>/Contents/Developer/usr/bin/xcodebuild                      (via xcrun)
<Xcode>/…/Toolchains/XcodeDefault.xctoolchain/usr/bin/ld
<Xcode>/…/Toolchains/XcodeDefault.xctoolchain/usr/bin/dsymutil
```

plus the Xcode macOS SDK as sysroot data. **MEASURED**: no manager-fixed input
removes them — with `ignoreXcodeVersionCheck=true` and `targetToolchain`,
`targetSysRoot` and `additionalToolsDir` all overridden to absolute local paths,
the compile still fails with `Cannot run program "/usr/bin/xcrun"` at
`CurrentXcode.getToolchain(Xcode.kt:92)` ←
`AppleConfigurablesImpl.getAbsoluteTargetToolchain(Apple.kt:45)`. The cause is
structural: `konan.properties` declares the Apple toolchain, sysroot and addon
as `remote:internal`, and `XcodePartsProvider` has exactly two implementations —
`InternalServer`, gated on `KONAN_USE_INTERNAL_SERVER` and pointing at
`https://repo.labs.intellij.net/kotlin-native`, and `Local`, the host's Xcode.

**Cache identity.** The Apple toolchain and SDK are in no fingerprinted tree, so
the build input cannot bind them and two hosts with different Xcode versions
would alias in the cache.

An implementation MUST fail `platform_unsupported` on macOS. It MUST NOT resolve
a host tool, ship a shim, declare the SDK data-only while executing from it, or
publish a second file.

### 11.3 Windows — unverified, candidate

`windows/amd64` is the only tuple that can admit this pair before Linux enters
the protocol's platform set. It is **unqualified**: no Windows host was
reachable for this task, and `TASK-260729-rhjxtx` measured that the reachable
Windows host carries no Kotlin toolchain of any backend — which section 1.3's
curation procedure is exactly what answers.

The reason macOS fails does not apply. **MEASURED** by reading the same
`konan.properties`: `targetToolchain.mingw_x64` and `targetSysRoot.mingw_x64`
both resolve to `$toolchainDependency.mingw_x64`, and
`llvm.mingw_x64.user = llvm-21-x86_64-windows-essentials-150`; none is
`remote:internal` and none names a host SDK, so all of them live under
`$KONAN_DATA_DIR/dependencies`. This is a reading of distribution data, not a
measurement of a Windows host, and it is not a claim.

An implementation MUST NOT resolve `cmd.exe`, `powershell.exe`, `link.exe`,
`cl.exe`, `lib.exe`, `vswhere.exe`, or a Visual Studio activation script; A3
must show the MinGW-family linking path inside `<kotlin_root>`. `windows/arm64`
is a separate tuple and is not implied by `windows/amd64`.

### 11.4 Linux — excluded, then candidate

Excluded from the protocol until `TASK-260728-1skseh`, then qualified by
`TASK-260728-3u1nho` with the identical A1–A9 test. The same properties reading
applies: `targetToolchain.linux_x64 = $gccToolchain.linux_x64/…`,
`targetSysRoot.linux_x64 = $gccToolchain.linux_x64/…/sysroot`,
`llvm.linux_x64.user = llvm-21-x86_64-linux-essentials-116`.

### 11.5 Retirement

If no tuple has passed A1–A9 when `TASK-260728-251p01` mints manifest schema 8,
both identifiers are retired unused: not reassigned to another language,
backend, artifact class, or source mode, and not enabled by relaxing another
driver. Claim schema 4 is then minted without them and both become structurally
unassertable.

### 11.6 Open per-tuple obligations

| # | Question | Failure consequence |
|---|---|---|
| K-2 | every spawned child inside `<kotlin_root>` or the overlay | that tuple is excluded (macOS: **fails**) |
| K-3 | no download, bundle byte-unchanged | that tuple is excluded |
| K-4 | exact produced filename and suffix | blocks A5 |
| K-6 | default native target token and mapping | blocks A7 |
| K-9 | the tuple's closed base-installation library allow-list | blocks A6 |
| K-10 | the overlay materialisation mechanism available on that platform | blocks A4 |

## 12. Diagnostics

| Code | Fires when |
|---|---|
| `build_package_code_execution_forbidden` | any 7.1 allow-list rejection, with the 7.2 per-surface diagnostic in the payload |
| `build_toolchain_metadata_mismatch` | 8.1 file-shape gate, or `kotlin_version` strictly above the resolved triple |
| `build_toolchain_requirement_invalid` | `toolchain.id` not `kotlin`; malformed requirement literal |
| `build_toolchain_requirement_unsatisfiable` | empty intersection of baseline, manifest and descriptor requirements |
| `build_toolchain_version_undetermined` | probe stdout unmatched or multiply matched by the anchored rule |
| `build_toolchain_untrusted` | a declared `<kotlin_root>` that does not exist, or whose `primary_relpath` is missing or not a regular executable |
| `build_toolchain_untested_release` | resolved family not in `compatibility` — which, while `compatibility` is empty, is **every** resolved version |
| `platform_unsupported` | host `(os, arch)` not in `platforms` — which, while `platforms` is empty, is **every** host, and which is the permanent outcome on macOS |
| `build_artifact_class_unsupported` | the 10.1 published-artifact gate, or a platform that cannot produce a single directly loadable file |
| `build_descriptor_driver_unsupported` | schema-1 descriptor named by a `kotlin-native-repository-v1` command |
| `build_descriptor_schema_unsupported` | unsupported descriptor version |
| `build_execution_control_unavailable` | any mandatory portable control unavailable, unchanged under `manager-worker-v2` |

The two rows marked "every" are the honest current state, not a defect: while
both sets are empty, an implementation that resolves a Kotlin toolchain at all
fails closed, which is the correct behaviour for a reserved identifier.

## 13. Capability limitations

Authoring guidance (`TASK-260728-2uh7em`) MUST carry all five:

1. **No third-party dependencies.** No `.klib` may be supplied or named, and the
   vector passes no `-library`. A build root compiles against the distribution's
   own libraries only.
2. **No user-defined C interop.** `cinterop` is a second tool and a second
   command; `manager-worker-v2` admits neither. The distribution's own platform
   bindings are available and are subject to the 10.1 gate.
3. **The build root is not an IDE project.** 7.1 excludes Gradle files. An
   author keeps the IDE project outside the build root; build roots are
   context-excluded and never runtime-copied, so this costs the agent nothing
   and costs the author a duplicated source layout.
4. **No cross-compilation.** One host builds for its own default target only.
5. **No macOS.** Section 11.2. The driver fails closed there and falls back to
   nothing.

## 14. Conformance vector inventory

Vectors are authored by `TASK-260728-251p01` only if 11.5 does not retire the
identifiers.

| Group | Cases | Notes |
|---|---|---|
| reserved-identifier rejection | 10 | each frozen manifest, descriptor, receipt, marker and claim schema rejects both identifiers |
| command shape | 8 | `buildCommandV8` and `repositoryBuildCommandV2` with the `kotlin-native` consts; missing member; extra member; `toolchain.id` not `kotlin` |
| descriptor | 4 | schema-2 positive; schema-1 with a Kotlin driver ⇒ `build_descriptor_driver_unsupported`; unknown schema; command/descriptor driver mismatch |
| module file shape | 12 | six 8.1 gate cases, plus duplicate member, non-object, wrong `schema_version` type, absent file, file outside the build root, file not a regular file |
| `kotlin_version` classifier | 10 | C1 equal, below, above ⇒ mismatch; C2 for prefix, prerelease, build metadata, leading zero, two-component, wildcard, empty |
| allow-list | 22 | one positive and one rejected case per row of 7.2, plus the `@`-name case, the `.kts` case, a nested directory, a rejected directory name, a symlink, and the module file placed outside the root |
| requirement intersection | 6 | baseline ∩ manifest ∩ descriptor, including the non-empty-but-excludes-resolved case decision 0007 section 4 names |
| policy and identity | 6 | policy object closed to its two members; `execution_policy` const; `compile_profile` const; single-element `toolchain_identities`; v1/v2 non-aliasing; receipt schema 3 vs 4 |
| artifact gate | 4 | a dependency inside the allow-list publishes; one outside it ⇒ `build_artifact_class_unsupported`; a second produced file is never published; the `.dSYM` by-product stays in staging |
| platform | 4 | empty `platforms` ⇒ `platform_unsupported` on every host; macOS ⇒ `platform_unsupported` permanently; empty `compatibility` ⇒ `build_toolchain_untested_release` on every resolved version |

Every vector declares the `compatibility` set and the `platforms` set as fixture
input, so outcomes are deterministic across managers, exactly as decision 0007
section 1.1.1 requires.
