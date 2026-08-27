# TASK-260728-168smo — Kotlin artifact model and driver pair

Developer handoff. Ready for review.

Selects the Kotlin CLI artifact/runtime model and designs the closed local and
external driver contracts, inside the boundary of decision 0008
(`TASK-260728-2spy93`, artifact class and driver namespace) and decision 0007
(`TASK-260728-1g0z69`, toolchain requirement and preflight), using the honest
absence recorded by `TASK-260729-rhjxtx` as the starting point rather than
working around it.

Nothing here is a platform claim. Both identifiers remain reserved.

## Deliverables

| Artifact | What it is |
|---|---|
| `_decision-0010-kotlin-native-driver-pair.md` | The decision: artifact-model selection with the four rejected Kotlin shapes, identifiers, the minted project-metadata file, command surfaces, the measured process graph, the allow-list rejection matrix, session and environment, registry entries, platform position, the retirement trigger, capability limits, 14 rejected alternatives |
| `_kotlin-native-build-drivers-reference.md` | Implementation-ready reference: both registry entries, identity, trusted launcher, source layout, the module file, environment table, the argument vector, the allow-list with per-surface diagnostics, Stage B classifier, cache/receipt/marker/claim identity, artifact rules, the Q1–Q9 acceptance test, diagnostics, 82 conformance vectors |
| `_command-evidence.log` | Every argv, real exit code, byte counts and bounded output, in order, for all seven measurement groups |
| `_gate-log.txt` | Gate transcript with real exit codes and attribution for the two expected-red gates |
| `_probe.tar.gz` | The reproducible probe tree: source fixtures, the `@`-injection fixtures, the 15 logging shims, and the runner |

Decision and reference are also written into the task worktree at
`.temp/TASK-260728-168smo/curator-spec-worktree/{decisions/0010-kotlin-native-driver-pair.md,docs/kotlin-native-build-drivers.md}`.
Nothing was staged, committed, pinned, or published.

## The question this task had to answer, and the answer

Decision 0008 section 10 made this task responsible for one thing:

> `TASK-260728-168smo` additionally decides the Kotlin backend **within**, not
> around, section 3; if no Kotlin backend satisfies it, both Kotlin identifiers
> are retired unused.

**Answer: Kotlin/Native, and no JVM shape.** Section 3 admits exactly one
artifact class — one bounded regular file, directly executable by the host
program loader. Four candidate Kotlin shapes were evaluated against it:

| Candidate | Files | Directly loadable | Verdict |
|---|---|---|---|
| thin JAR plus operator JRE | 1 | no | `runtime-bundle` |
| fat JAR (`-include-runtime`) | 1 | **no** | `runtime-bundle` |
| `jlink`/`jpackage` image | many | launcher plus tree | `runtime-bundle` |
| GraalVM `native-image` | 1 | yes | rejected on four independent grounds |
| **Kotlin/Native `-produce program`** | **1** | **yes** | **selected** |

The fat JAR is the row worth stating explicitly, because it is the shape usually
described as self-contained: it is one file and it is still not loadable without
a JVM, so one file is necessary and not sufficient.

GraalVM is rejected because it needs a second compile-phase command that
`manager-worker-v2`'s session shape refuses, a third toolchain outside decision
0007's closed identifier set, and a classpath that only a package manager or a
package-controlled command member could assemble — and because it is a different
backend, which decision 0008 section 2 requires to carry a different family
segment.

`jdk` stays a **companion**, and it is required under a *native* artifact model —
which corrects decision 0007's "`jdk` if JVM" expectation. The reason is that the
Kotlin/Native compiler is JVM-hosted, so the JDK is a build input, not an
execution-time dependency of the produced artifact. That distinction is what
keeps the selection inside section 3.

## Four measured findings that decided the design

Everything below was measured on this host against the **Kotlin/JVM** CLI
(`kotlinc-jvm 2.4.10`) and OpenJDK 26.0.1, because `TASK-260729-rhjxtx`'s finding
holds and was re-confirmed: **no Kotlin/Native distribution exists on any
reachable host.** Where a measurement is used to reason about Kotlin/Native, the
inference is labelled and carried as a named obligation, never as a claim.

### 1. A filename alone is an arbitrary-code-execution channel

The Kotlin CLI expands a bare `@`-prefixed argument as a response file. With a
build-root file named `@inject` containing `-script`, and a `.kts` file:

| Argument form | Expanded | Outcome |
|---|---|---|
| `@inject` | **yes** | package code **executed**, marker file written, exit **0** |
| `<abs>/@inject` | no | `error: source entry is not a Kotlin file`, exit 1 |
| `./@inject` | no | same, exit 1 |
| `@atfile` containing `-verbose` | yes | compiler ran verbose, exit 0 |
| `@nonexistent` | n/a | `warning: argfile not found`, exit **0** |

So a package that ships a file named `@x` in a directory the driver enumerates
executes arbitrary Kotlin at build time — through the compiler alone, with no
Gradle, and with a zero exit code. A *missing* response file is only a warning,
so a partial mitigation fails silently.

This is why the rejection matrix is a **closed allow-list over name shapes**
rather than a deny-list of filenames: the vector is a name shape, not a name.
The build root admits only directories, `*.kt` regular files, and one
`kotlin-native-module.json`, all with a leading-character class that excludes
dotfiles and `@`. Absolute-path argv is kept as a measured structural backstop,
not as the rule, because decision 0008 section 7 requires the surface rejected
*before* the compile phase.

A `.kts` reaching the compiler *without* `-script` was measured to be compiled
into a class and **not** executed. That nuance is recorded rather than
overstated: `.kts` is inert alone and is the payload half of the `@` vector.

### 2. The shipped launcher is inadmissible; bypassing it closes everything

Under a `PATH` of 15 logging shims:

| Run | `PATH` resolutions | Exit | Artifact |
|---|---|---|---|
| control: `<kotlin_root>/bin/kotlinc` | **4** — `uname`, `dirname`, `java -version`, `java <compile>` (plus `bash` from the shebang) | 127 | none |
| pinned: `<jdk_root>/bin/java -cp <kotlin_root>/…` | **0** | 0 | produced |

The launcher is a bash script that resolves the compiler process as the bare
name `java`, and it reads `JAVACMD`, `JAVA_HOME`, `JAVA_OPTS`, `KOTLIN_RUNNER`,
`KOTLIN_COMPILER`, and `KOTLIN_TOOL` from the inherited environment —
`KOTLIN_COMPILER` selects the compiler main class, i.e. an environment-selected
backend. That is exactly what decision 0007 section 3 forbids of a primary
executable and decision 0008 section 6 item 3 forbids of the process graph.

The driver therefore never runs any launcher script, on any platform, including
for the version probe. `<jdk_root>/bin/java` is a `Mach-O 64-bit executable
arm64` — a regular executable at a fixed relpath inside the fingerprinted tree.
Six environment channels are closed *structurally* rather than by
neutralisation, because the script that reads them is never run.

This is the Kotlin analogue of the Rust linker/SDK pinning result, and it lands
in the same place: the whole process closure sits inside fingerprinted,
operator-trusted roots.

### 3. Kotlin has no lawful project-metadata file, so the driver mints one

Decision 0008 section 4 requires exactly one closed driver-defined
project-metadata file; decision 0007 section 1.3 requires "exactly one file and
field". Every ecosystem candidate fails: Gradle scripts because reading one is
executing one (`TASK-260729-rhjxtx` measured `gradle properties` compiling the
build script as `_BuildScript_`), `gradle.properties` because its keys select
compiler and daemon behaviour, `pom.xml` because the Maven model *is* its plugin
graph, a bare marker because decision 0007 needs a field.

The driver therefore binds `kotlin-native-module.json`:

```json
{"schema_version": 1, "kotlin_version": "2.4.0"}
```

Two members, both REQUIRED, `additionalProperties: false`, never passed to the
compiler, selecting nothing. Three properties follow, and each is a
simplification rather than a special case: the Stage B classifier is **two**
classes rather than Rust's seven, because Curator owns the grammar and there is
no ecosystem grammar or edition floor to be independent of; the security
partition `F` is empty, so decision 0007's P1 and P2 collapse to the satisfiable
equality `C = Upstream`; and the file is inert to the compiler, so deleting it
is a deterministic Curator rejection rather than a build difference.

### 4. The JDK probe stream, chosen by measurement

| Command | stdout | stderr |
|---|---|---|
| `java --version` | `openjdk 26.0.1 2026-04-21` (148 B) | empty |
| `java -version` | empty | quoted banner (158 B) |
| `java --version` with `JDK_JAVA_OPTIONS` set | **unchanged, 148 B** | 2,205 B incl. `NOTE: Picked up …` |

`--version`/stdout is pinned: one stream, and byte-stable even when the
environment carries an injection variable. `JDK_JAVA_OPTIONS` and
`JAVA_TOOL_OPTIONS` were both measured to be honoured, so the environment table
unsets them; the stdout choice is the layer that holds if that unset is missed.

## The platform position, and why it is empty

`platforms` holds the **empty set** for both identifiers, and `compatibility` is
empty for both registry entries. Neither is an omission:

- decision 0007 section 1.1.1 admits a `compatibility` family only after it has
  passed **that driver's** conformance vectors, and no Kotlin vectors exist;
- decision 0008 section 9 starts every reserved identifier with an empty
  qualified-platform set;
- macOS is not claimed merely because the design was reasoned about on a macOS
  host. There is no Kotlin/Native distribution here to qualify.

The `kotlin` and `jdk` entries are therefore **reserved and incomplete** after
this record, which is the state decision 0007 section 1.3 prescribes for a
reserved entry with no qualified host. What this task supplies instead is the
part that makes qualification mechanical rather than a second design pass: entry
shape, metadata source, probe stream discipline, the anchored `jdk`
normalization rule (measured), the `kotlin` normalization requirement that the
rule must assert the *native backend token* and not merely a version, and a
decision rule for every outstanding field.

Eight obligations are named rather than guessed — launcher shape, child-process
containment, distribution self-containment, output filename, `.kts` under the
native backend, default target spelling, the `kotlin` probe, and both baselines —
each with the consequence of failing it attached. The `Q1–Q9` acceptance test is
identical on every candidate platform and includes the requirement that the
control run must fire: without it, a zero `PATH`-resolution count proves nothing.

Per-platform, the sharp statements are:

- **macOS**: if `clang`, `ld`, `libtool`, `xcrun`, `dsymutil`,
  `install_name_tool`, or `codesign` is spawned from outside a fingerprinted
  root, macOS is **not admitted**. Decision 0007's identifier set is closed to
  `{go, rust, swift, kotlin, jdk}`, and Rust's data-only SDK root is not a
  precedent for an *executed* linker.
- **Windows**: `cmd.exe`, `powershell.exe`, `link.exe`, `cl.exe`, `lib.exe`,
  `vswhere.exe` and VS activation scripts are all forbidden; the MinGW linking
  path must be shown to live inside the Kotlin root. `amd64` and `arm64` are
  separate tuples.
- **Linux**: excluded until `TASK-260728-1skseh`, then `TASK-260728-3u1nho` runs
  the identical test plus the base-installation dependency check.

And the retirement trigger is fixed now rather than left to drift: if no tuple
has passed `Q1–Q9` when `TASK-260728-251p01` mints manifest schema 8, both
identifiers are retired unused and claim schema 4 is minted over six identifiers
rather than eight — exactly the outcome decision 0008 section 8 anticipated in
writing.

## Capability limitations, stated rather than discovered

The design is deliberately the narrowest of the four languages, and the
authoring guidance task is required to carry all four:

1. **No third-party dependencies.** A package cannot supply or name a `.klib`
   and the vector passes no `-library`; a build root compiles against the
   distribution's standard library only. This is what removes the
   dependency-resolution graph phase entirely — the session is **zero** graph
   commands plus one compile command, which is the case decision 0008 section
   7's "at most one" was written to admit.
2. **No C interop.** `cinterop` is a second tool and a second command; the
   session admits neither.
3. **The build root is not an IDE project**, because the allow-list excludes
   Gradle files. Authors keep the IDE project outside the build root. Build roots
   are context-excluded and never runtime-copied, so this costs the agent nothing
   and costs the author a duplicated source layout.
4. **No cross-compilation.** One host builds for its own target only — which also
   removes the representable-but-unserved-target gate that Rust and Swift both
   needed.

## One narrow reading of decision 0007, recorded rather than slipped in

Under the JVM-hosted launcher shape the `kotlin` root contributes no launcher
process, and the pipeline's primary executable is the companion JDK's
`bin/java`. This is a completion of a reserved entry within the mandate decision
0007 section 1.3 grants ("MUST each supply … every field … plus its companion
list"), not a widening of its section 3: every executed binary is still a regular
file at a fixed relpath inside a fingerprinted, operator-trusted root, nothing is
package-selected, no wrapper is run, and no auto-install is introduced. It is
strictly stronger than honouring the shipped launcher would be, and the
measurement is what makes that checkable rather than argued.

If decision 0007's owner rejects the reading, the consequence is the section 10
retirement, not a compensating wrapper. Flagging it here so the reviewer decides
deliberately rather than by omission.

## Gates

Real exit codes, each command run standalone, no pipeline masking. Full
transcript in `_gate-log.txt`.

| Gate | Exit |
|---|---|
| spec `tools/validate.py`, clean 57c1f56 baseline worktree | 0 |
| spec `tools/validate.py`, task worktree | **1** |
| spec `python3 -B -m unittest discover -s tools -p 'test_*.py'` | 0 |
| spec `go test ./tools/...` | 0 |
| scoped link check over the two authored documents | 0 (4 local links, 0 broken) |
| whole-tree link check (attribution) | **1** (3 broken, 0 from this task) |
| curator `go build ./...` | 0 |
| curator `go vet ./...` | 0 |
| curator `go test ./...` | 0 |
| curator `gofmt -l ./cmd ./internal` | 0 |
| curator `make check` | **2** |

`golangci-lint` was **not run**: the binary is not installed on this host
(`command -v golangci-lint` exit 1). Same state the Rust contract task recorded.

Both non-zero gates are expected-red and attributed, and neither is from this
task:

- **spec `validate.py` exit 1** fails at its link check on
  `docs/external-build-repositories.md` → `../release/1.0.0-rc.5.json` and two
  occurrences of `docs/portable-go-execution-policy.md` →
  `../conformance/v1/vectors/go-host-execution-policy.json`. Both documents were
  **copied** into the task worktree unmodified from another task's in-flight
  tree so the link graph reachable from decisions 0005–0008 could be evaluated;
  neither target exists at base commit 57c1f56. The clean baseline worktree at
  the same commit exits 0, and the scoped check over the two documents this task
  authored reports 4 local links and 0 broken.
- **curator `make check` exit 2** fails at its `gofmt -l .` stage on four
  vendored third-party files under
  `.temp/TASK-260720-1zntv0/cycle2/curator/vendor/`, another task's scratch tree.
  This task authored **0** Go files (verified by `find`), and
  `gofmt -l ./cmd ./internal` over the tracked tree exits 0.

## Scope kept

No normative curator-spec file was modified. `decisions/0010-…` and
`docs/kotlin-native-build-drivers.md` are new files; decisions 0004 through 0008
and three `docs/` documents were **copied** into the task worktree unmodified so
the link graph could be evaluated. No schema, vector, release pin, dependency,
generated corpus, or release metadata was touched. No frozen artifact was
altered; nothing was staged, committed, pinned, or published.

Nothing was installed, downloaded, updated, activated, or switched on any host.
No Kotlin/Native distribution was obtained, and none was simulated: the three
Kotlin/Native questions that need one are named obligations, not filled fields.

## Open item for the reviewer

Decision number `0010` is reserved by this record. `0009` is claimed
concurrently by `TASK-260728-12pnm1` (Rust) and `TASK-260728-1jafds` (hardened
execution profile), and `TASK-260728-1yhuqi` (Swift) had authored no record at
the time of writing. If a lower free number is available when this lands, it
renumbers rather than contests, as `TASK-260728-2spy93` did when `0007` was
claimed.
