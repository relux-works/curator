# TASK-260728-168smo review verdict — cycle 1

## Verdict

**Changes requested → `analysis`.**

The Kotlin/Native artifact choice is consistent with decision 0008's
`native-executable-v1` boundary, the JVM alternatives are explicitly deferred,
and the local/external command surfaces, package-tree allow-list, response-file
defence, cache/receipt/marker separation, and signing boundary are useful.
However, the proposed driver pair is not yet implementable: its toolchain and
offline-input closure is contradicted by the official Kotlin/Native
distribution, and the registry/process/platform fields that this task owns are
still alternatives or placeholders.

## Blocking findings

### 1. The proposed Kotlin primary has no valid primary executable

Accepted decision 0007 section 3 requires the primary executable to be a
regular executable at the entry's fixed relpath inside the tree being
fingerprinted. The checksum-verified official Kotlin/Native 2.4.10 distribution
contains shell launchers at `bin/konanc`, `bin/kotlinc-native`, and
`bin/run_konan`; the actual compiler is
`konan/lib/kotlin-native-compiler-embeddable.jar` with main class
`org.jetbrains.kotlin.cli.utilities.MainKt`, launched as tool `konanc` by a
JDK. A JAR is not a regular executable, while `<jdk_root>/bin/java` belongs to
the companion entry, not the `kotlin` primary root.

The submitted "shape (a) or shape (b)" and its proposed narrow reading do not
complete the registry entry. They leave `kotlin.primary_relpath` unresolved and
contradict the accepted per-entry primary-executable invariant. Shape (b) is
not the official 2.4.10 distribution shape.

### 2. The offline input and process closure is false for the official distribution

The official macOS arm64 archive
`kotlin-native-prebuilt-macos-aarch64-2.4.10.tar.gz` was downloaded to `/tmp`
only and verified against the release checksum:

```text
expected 55ded039bb56a69aec9df354a92b42df9e916104e3c53d8d9852d9cc6617ed9d
actual   55ded039bb56a69aec9df354a92b42df9e916104e3c53d8d9852d9cc6617ed9d
```

Its `konan/konan.properties` declares
`dependenciesUrl = https://download.jetbrains.com/kotlin/native` and names
separate LLVM, target-toolchain, sysroot, additional-tools, and LLDB
dependencies. Those executable/data trees are not present as a complete
offline closure in the release archive and are outside the submitted ordered
identity `[kotlin, jdk]`.

A direct host-native compile used the submitted JVM/main-class shape, an empty
`KONAN_DATA_DIR`, `airplaneMode=true`, `-produce program`,
`-target macos_arm64`, and one absolute `.kt` source. It exited 2:

```text
Downloading native dependencies (LLVM, sysroot etc).
Cannot find a dependency locally: lldb-4-macos.
```

It produced no executable and mutated the supposedly empty-and-unchanged data
directory by creating:

```text
dependencies/
dependencies/.extracted/
dependencies/cache/
dependencies/cache/.lock
```

Therefore Q4 cannot pass as written. The process graph and
`curator-kotlin-toolchain-v1` identity also omit the downloaded dependency
closure whose linker/tool executables the compiler uses. This is not a missing
test; it is a contradiction in the trusted-input, offline, fingerprint, and
cache identity model.

### 3. The task-owned registry and platform contract remains incomplete

The reference still leaves the following task-owned values as K obligations:

- `kotlin.primary_relpath`;
- Kotlin probe argv, output stream, native-token normalization, and baseline;
- JDK baseline;
- companion shape (`[jdk]` or empty);
- exact compile argv and output name;
- initial tested compatibility family;
- every platform tuple.

Both `platforms` and `compatibility` are empty, so the proposed driver rejects
every host and every resolved version. A delayed retirement decision at schema
8 does not make this an implementable v1 contract and transfers this task's
accepted-decision obligations to a later qualification task. The official
distribution is available for read-only probing even though it was not
preinstalled on the host.

### 4. The native-interop statement is not closed

The submitted reference says the build uses the distribution standard library
"and nothing else" and labels the capability "No C interop". The official
distribution contains target platform `.klib` libraries, and Kotlin's official
documentation states that platform libraries such as POSIX, Win32, and Apple
framework bindings are available from the compiler distribution. Package
source can import those bindings without supplying a `.def`, a `.klib`, or a
`-library` argument.

The contract must explicitly allow this fixed distribution-owned platform
library surface and bind its link/dynamic-dependency consequences into the
toolchain/platform proof, or reject it with an enforceable pre-compile rule.
The policy for source-level native interop imports and annotations cannot be
replaced by rejecting only package files and argv.

## Required rework

1. Probe a checksum-verified official Kotlin/Native distribution and select one
   exact launcher/root/companion shape. Remove the unmeasured alternative shape
   and fill every registry and compile-vector field.
2. Resolve the accepted decision 0007 primary-executable conflict. Under the
   current architecture, retire both Kotlin identifiers if the Kotlin primary
   cannot own a regular executable. A proposal to change that invariant needs a
   separately reviewed upstream architecture decision, not a "narrow reading".
3. Define a complete, offline, operator-trusted Kotlin/Native dependency
   closure. If a curated prehydrated bundle is proposed, specify its immutable
   root layout, provenance, fingerprint algorithm, writable overlay policy,
   dependency lock/cache behaviour, exact child executable containment, and
   cache/receipt identity. Demonstrate that compilation performs no download.
4. Admit at least one evidence-backed tuple through Q1–Q9 or retire the pair
   now. Record macOS, Windows, and Linux as exact supported/unsupported tuples
   with reasons; do not leave every tuple for downstream discovery.
5. Close the distribution platform-library/native-interop/import/annotation
   policy and correct the "stdlib only / no C interop" capability statement.
6. Update the decision/reference/results and rerun the scoped documentation
   gates. Preserve the sound JVM rejection, closed Gradle/Maven/KSP/plugin/
   script/package-toolchain rules, `@` filename defence, and local/external
   equivalence.

## Evidence and sources

- Accepted local contracts:
  `decisions/0007-compiled-build-toolchain-preflight.md` sections 1, 3, and
  downstream obligations; `decisions/0008-additional-language-driver-boundary.md`
  sections 2, 3, and 10.
- Official release:
  <https://github.com/JetBrains/kotlin/releases/tag/v2.4.10>
- Official Kotlin/Native command-line documentation:
  <https://kotlinlang.org/docs/native-get-started.html>
- Official Kotlin/Native platform-library documentation:
  <https://kotlinlang.org/docs/native-platform-libs.html>
- Independent probe performed on macOS arm64 with OpenJDK 26.0.1. No repository
  code, toolchain installation, staging, commit, publication, or pin change was
  performed.

