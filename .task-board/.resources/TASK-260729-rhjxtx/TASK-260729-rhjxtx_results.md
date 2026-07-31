# TASK-260729-rhjxtx — Rust, Swift and Kotlin Native toolchain boundary probe

Developer handoff, **rework cycle 1**. Supersedes the cycle-0 handoff.

Host-measured evidence for the three language-driver contracts
(`TASK-260728-12pnm1` Rust, `TASK-260728-1yhuqi` Swift, `TASK-260728-168smo`
Kotlin/Native). Nothing here is normative and nothing here is a platform claim.

## What changed in this cycle

All four findings in `TASK-260729-rhjxtx_review-verdict-cycle-1.md` are closed.

| # | Finding | Resolution |
|---|---|---|
| 1 | `.swift-version` carried two different expected classes across hosts | One canonical case table in `cases.go`; `finish()` re-applies it to every emitted case; two regressions compare declared against executed |
| 2 | Swift `upstream_rejects_after_compilation` was inferred from a non-zero exit | Replaced with a measured `-v` job-trace line, carried in the fixture as `build_compilation_evidence`; the boolean is now derived from that line and cannot be set independently |
| 3 | The supplied-root Kotlin/Native branch returned "corpus is not authored" | Branch implemented and unit-tested: stderr version/backend normalization, `-list-targets` parsing, native-target identity, positive version+target case, unknown-target and unserved-target controls; metadata boundary recorded as an explicit measured absence plus one new Gradle case that needs no Kotlin/Native |
| 4 | Windows cross-compile/transfer/execute/cleanup commands were not recorded | `TASK-260729-rhjxtx_windows-commands.log`: 12 steps, exact argv, per-command exit code, both digests |

**The corrected fact from finding 2.** The cycle-0 handoff claimed "Rust is the
only case where upstream rejection arrives after compilation starts." That was
an artefact of the inference, and the measurement contradicts it. With `swiftc
-v`, the driver's own job trace shows it **does** execute a
`swift-frontend -frontend -c -primary-file …` job for
`x86_64-unknown-linux-gnu` before failing on the standard library, and executes
**no** frontend job for `not-a-real-triple`. Rust and Swift therefore both
reject an unserved-but-known target only after starting a compiler; only the
unknown-triple case is refused by the driver alone.

## Hosts

| | macOS | Windows |
|---|---|---|
| OS | 26.5, arm64 | 10.0.19045, amd64, via SSH alias `win` |
| Rust | 1.91.0, `~/.rustup/toolchains/1.91.0-aarch64-apple-darwin` | absent |
| Swift | Apple Swift 6.3.2, Xcode 26.5 toolchain | absent |
| Kotlin/Native | **absent** | absent |
| Kotlin JVM | kotlinc-jvm 2.4.10 (measured as the wrong backend) | absent |
| Gradle | 9.2.1 (measured only to price reading Kotlin metadata) | absent |
| Result | 19 cases, 16 matched, **0 divergences**, 3 not run, exit 0 | 19 cases, 19 not run, exit 0 |

Nothing was installed, downloaded, updated, activated, or switched on either
host. The remote working directory was removed and its absence confirmed.

## Per-language findings

### Rust — `TASK-260728-12pnm1`

1. **Version.** `rustc -vV` `release:` line on stdout. `rustc --version` embeds
   the commit hash on the same line; the verbose form is the narrower surface.
2. **Native target.** `rustc --print host-tuple`. Carries no platform-version
   component, so it is directly usable as a native-target identity.
3. **Metadata.** `Cargo.toml`, read with `cargo metadata --no-deps
   --format-version 1 --offline`. `packages[].rust_version` is surfaced without
   compiling anything, so the MSRV host gate is decidable in the graph phase.
4. **Target admission is manager-owned.** `rustc --print target-libdir --target
   x86_64-unknown-linux-gnu` exits **0** and prints a path for a target whose
   standard library is absent. The manager must stat
   `<root>/lib/rustlib/<target>/lib` inside the fingerprinted tree. Measured:
   cargo's own rejection arrives after `Compiling probe-positive v0.1.0`.
5. **`rust-toolchain.toml` is a live selector.** `path=` redirected the rustc
   shim (`invalid toolchain: the path … has no bin/ directory`, exit 1);
   `channel = "nightly"` made it attempt a download (`syncing channel updates
   for 'nightly-aarch64-apple-darwin'`). Measured against
   `RUSTUP_DIST_SERVER=http://127.0.0.1:1`, so the attempt is observable and
   cannot succeed. The directly resolved root is inert against both.
6. **Defect the probe found on itself.** Cargo resolves `rustc` by name from
   PATH. Under a manager-owned minimal PATH it fails with "could not execute
   process `rustc -vV` (never executed)". A Rust driver **must** pass `RUSTC`
   explicitly, or the second node of its process graph is chosen by PATH order
   rather than by the fingerprinted closure.

### Swift — `TASK-260728-1yhuqi`

1. **Version.** `compilerVersion` from `swiftc -print-target-info` (JSON on
   stdout). `swift --version` splits one banner across **two** streams —
   `swift-driver version: 1.127.14` on stderr without a trailing newline, the
   Apple version line on stdout — so anything merging the streams sees them
   concatenated into one line and an anchored rule stops matching.
2. **Native target.** `target.triple` carries a deployment-version component
   supplied by the SDK (`arm64-apple-macosx26.0` on a 26.5 host).
   `target.unversionedTriple` is the identity form.
3. **Metadata, header only.** `swift package tools-version` returns exit 0 and
   the normalized version for a manifest carrying a Swift syntax error **and**
   for one whose body is `fatalError`. It reads the header line and neither
   compiles nor runs the manifest. A Swift graph phase is bounded there.
4. **`dump-package` is code execution.** Measured: it compiled the manifest and
   the `fatalError` fixture reported `Fatal error: PROBE-MANIFEST-EXECUTED`.
   Never invoke `dump-package`, `describe`, `resolve`, or `build` to learn what
   a package declares.
5. **Target admission is manager-owned**, same shape as Rust.
   `-print-target-info -target x86_64-unknown-linux-gnu` exits **0**; compiling
   fails with `unable to load standard library for target`. The early gate is
   to stat the `paths.runtimeLibraryPaths` the same command reports.
6. **Where the compiler boundary actually is** (corrected this cycle). Measured
   from the `-v` job trace, not inferred:

   | case | frontend job spawned | `upstream_rejects_after_compilation` |
   |---|---|---|
   | `not-a-real-triple` | no | false |
   | `x86_64-unknown-linux-gnu` | yes | true |

7. **`.swift-version` is inert.** Honored by the `swiftly` version manager, not
   by a directly resolved toolchain; the file changed nothing. That is what
   admits the field as `compared` rather than `forbidden`.

### Kotlin/Native — `TASK-260728-168smo`

**Honest status first: no Kotlin/Native distribution exists on either host.**
Three of the five Kotlin cases are `not_run`, the `kotlinc-native` normalization
rule is named `kotlin.konanc.version.UNVERIFIED`, and the `-list-targets` parser
has never seen real output. The supplied-root branch is now genuinely runnable
and unit-tested against a **synthetic** root; that verifies the probe's own
logic and is explicitly not evidence about Kotlin/Native.

Measured on macOS without any Kotlin/Native distribution:

1. **The Kotlin compiler that is present is the wrong backend.** `kotlinc
   -version` writes `info: kotlinc-jvm 2.4.10 (JRE 26.0.1)` to **stderr** with
   an empty stdout. A probe reading stdout gets nothing; a probe accepting any
   `kotlinc` banner would admit a front end that cannot produce a
   `native-executable-v1` artifact. The Kotlin entry's probe must name the
   stderr stream and its normalization must require the backend token.
2. **Kotlin has no legal project-metadata file for section 4.** This is the
   load-bearing input for the Kotlin contract, and it is recorded as an
   explicit absence rather than a placeholder:
   - `kotlinc-native` consumes `.kt` sources and an output path. It defines no
     manifest. (UNVERIFIED — needs a qualified host to confirm from `-help`.)
   - The ecosystem's manifest is a Gradle script. **Measured**: answering the
     pure metadata query `gradle properties` compiled the build script as a
     program source unit (`_BuildScript_`) before it could answer. On this host
     it then crashed, because Gradle 9.2.1 rejects the installed JDK's class
     file version — the compilation is the finding; finishing is incidental.
     `build.gradle.kts` was **not** measured: resolving the Kotlin DSL requires
     network fetches this probe refuses to perform.
   - Consequence for `TASK-260728-168smo`: it must either mint a curator-owned
     metadata file for the build root, or accept that decision 0008 section 10's
     retirement clause applies to both Kotlin identifiers. This probe proposes
     no candidate file and no candidate field.
3. **What the supplied-root branch will measure when a host exists**: version
   and backend from the stderr banner; native target from the `-list-targets`
   entry marked `(default)`; a positive version+target case; an unknown-target
   representability control; and an unserved-target host gate whose target is
   *selected from the distribution's own output* rather than asserted. Every one
   of them degrades to `not_run` with a stated reason if the real output does
   not match the declared shape.

## Cross-language conclusions

1. **Representability and host admission are different questions, and every
   ecosystem answers only the first cheaply.** Rust and Swift both accept a
   target they cannot build for, with exit 0. Admission is a manager-side stat
   inside the fingerprinted tree in both cases.
2. **Three of four version probes need an explicit stream.** `swift --version`
   splits across two; `kotlinc -version` is stderr-only; `rustc --version` mixes
   in the commit hash. A normalization rule that does not name its stream is not
   reproducible.
3. **Every ecosystem has a package-owned selector, and they differ in kind.**
   Rust's `rust-toolchain.toml` actively redirects and can trigger a download.
   Swift's `.swift-version` is inert against a directly resolved toolchain.
   Kotlin's is the whole build script.
4. **Only Rust and Swift have a metadata file readable without executing
   package code.** Kotlin does not, and that is a decision input, not a gap in
   this probe.
5. **A compiler's exit code cannot tell you where the boundary is.** Both Swift
   target cases exit 1. Only the measured job trace separates "the driver
   refused the triple" from "a frontend ran and failed".

## Gates

Real exit codes, each command run standalone. Full transcript in
`TASK-260729-rhjxtx_gate-log.txt`.

| Gate | Exit |
|---|---|
| probe `gofmt -l .` | 0 (empty) |
| probe `go build ./...` | 0 |
| probe `go vet ./...` | 0 |
| probe `go test -count=1 ./...` | 0 |
| probe `golangci-lint run ./...` | 0 (`0 issues.`) |
| probe executed-vs-declared regression against real local toolchains | 0 |
| curator `go build ./...` | 0 |
| curator `make vet` | 0 |
| curator `make test` | 0 |
| curator `make check` | **2** |
| curator `gofmt -l .` (attribution) | 0 |

`make check` fails at its `gofmt -l .` stage on four files, all of them vendored
third-party sources under `.temp/TASK-260720-1zntv0/cycle2/curator/vendor/` —
another task's scratch tree. Zero are from this task, and no tracked project
file was modified.

## Artifacts

| Artifact | What it is |
|---|---|
| `_probe.tar.gz` | The standalone Go module: 9 sources, 4 test files, 49 test functions |
| `_fixture-macos.json` | `toolchain-probe-fixture-v1`, macOS run |
| `_fixture-windows.json` | `toolchain-probe-fixture-v1`, Windows run |
| `_command-evidence.log` | Both runs rendered: every argv, env delta, exit code, output excerpt, fixture digest, manager obligation |
| `_windows-commands.log` | The 12 exact Windows steps with per-command exit codes and both digests |
| `_windows-inventory.log` | Read-only Windows inventory: 12 PATH names, 12 candidate roots, all absent |
| `_gate-log.txt` | Gate transcript with real exit codes |
| `_results.md` | This document |

## Scope kept

No normative curator-spec file, implementation source, release pin, dependency,
schema, or vector was touched. The probe is a standalone module under `.temp/`
with its own `go.mod`. No package-controlled paths, environment overrides, URLs,
channels, install commands, or trust roots were introduced. Kotlin means Kotlin
Native throughout. Linux was not probed and remains later non-gating validation.
