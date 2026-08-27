# TASK-260729-rhjxtx — host-verified Rust, Swift, and Kotlin Native toolchain boundary probe

Developer handoff evidence. Consumed by `TASK-260728-12pnm1` (Rust),
`TASK-260728-1yhuqi` (Swift), and `TASK-260728-168smo` (Kotlin).

Boundary inputs: decision 0008 (`TASK-260728-2spy93`, accepted) fixes the
artifact class, the reserved driver identifiers, and `manager-worker-v2`.
Decision 0007 (`TASK-260728-1g0z69`, in review) is treated as **provisional**:
this task supplies measurements for the `rust`, `swift`, and `kotlin` registry
entries it reserves, and changes nothing normative.

Nothing in `curator-spec`, this repository's Go module, its release pins, or its
dependencies was modified. No toolchain was installed, downloaded, updated,
activated, or switched on either host.

---

## 1. What was produced

| Artifact | What it is |
|---|---|
| `TASK-260729-rhjxtx_probe.tar.gz` | the probe program: a standalone Go module (`main`, `normalize.go`, `classify.go`, `harness.go`, `model.go`, `rust.go`, `swift.go`, `kotlin.go`) plus 34 unit tests |
| `TASK-260729-rhjxtx_fixture-macos.json` | machine-readable evidence, macOS 26.5 arm64 |
| `TASK-260729-rhjxtx_fixture-windows.json` | machine-readable evidence, Windows 10.0.19045 amd64 |
| `TASK-260729-rhjxtx_command-evidence.log` | the same two runs rendered with every argv, exit code, and output excerpt |
| `TASK-260729-rhjxtx_windows-inventory.log` | what was searched for on the Windows host and found absent |
| `TASK-260729-rhjxtx_gate-log.txt` | exit codes of every gate command run for this task |

Fixture format: `toolchain-probe-fixture-v1`.

Reproduce:

```
toolchainboundary -rust ROOT -swift ROOT [-kotlin ROOT]
                  [-rust-shim PATH] [-kotlinc-jvm PATH]
                  -work DIR -out FILE
```

Exit 0 = every case that ran matched its expected classification; 1 = at least
one diverged; 2 = the probe could not run. An absent toolchain is **not** a
failure: it is an evidence record plus a guidance input, and its cases are
`not_run` rather than passed.

---

## 2. The central distinction the probe enforces

Decision 0007 §4.2.1.2 records the mistake that forced the Go boundary probe to
be rewritten: an exit status from a build command answers the *conjunction* of
two independent questions and cannot be decomposed back into them.

| Question | Property of | Host-dependent? |
|---|---|---|
| **representability** — can the ecosystem hold this value at all? | the value | no |
| **host gate** — will *this* toolchain accept it? | the (value, host) pair | yes |

Every case here is therefore measured with **at least two commands**: a cheap
metadata-only command and a build or compile command. `classify.go` derives the
gate class from the observations alone and cannot see the expectation, so a
divergence is possible; that is what makes this a measurement rather than a
restatement. `classify_test.go::TestBuildExitAloneCannotSeparateTheTwo` pins the
regression: the Rust future-MSRV case and the Rust malformed-MSRV case share
build exit code 101 and must still classify differently.

Six gate classes are used: `representability`, `host_version`,
`target_representability`, `target_host`, `package_selector`, `backend_identity`.

---

## 3. Host evidence

### macOS 26.5, arm64 (Darwin 25.5.0) — primary

| Toolchain | Availability | Resolved | Root |
|---|---|---|---|
| `rust` | available | 1.91.0 | `~/.rustup/toolchains/1.91.0-aarch64-apple-darwin` |
| `swift` | available | 6.3.2 (Apple) | `/Applications/Xcode_26_5.app/…/XcodeDefault.xctoolchain` |
| `kotlin` | **unavailable** | — | none on the host |

19 cases: **15 matched, 0 divergences, 4 not run**, probe exit **0**.

### Windows 10.0.19045.6456, amd64, SSH alias `win` — reachable secondary

| Toolchain | Availability |
|---|---|
| `rust` | **unavailable** |
| `swift` | **unavailable** |
| `kotlin` | **unavailable** |

19 cases: **0 matched, 0 divergences, 19 not run**, probe exit **0**.

The host carries no Rust, Swift, Kotlin, JDK, or Go toolchain — neither on
`PATH` nor at any of the twelve candidate roots probed read-only
(`TASK-260729-rhjxtx_windows-inventory.log`). Nothing was installed. The probe
binary was cross-compiled on macOS, verified by SHA-256 on the remote host
(`c33a840b2eaaf9c4bc92ffb2dc22f738344b7b1918054496d8d633fb70fc3493`), run, and
then removed together with its working directory.

**Consequence for the three design tasks:** no Windows platform claim for any of
the six reserved driver identifiers can be evidence-backed today. Decision 0008
§9 already starts each identifier with an empty qualified-platform set; this run
confirms the Windows half of that set cannot be filled from the currently
reachable estate, and `TASK-260728-2bu2q6` will need a Windows host with real
toolchains before it can emit a Windows claim.

---

## 4. Rust — `TASK-260728-12pnm1`

### 4.1 Registry entry fields, measured

| Field | Measured value |
|---|---|
| version probe | `<root>/bin/rustc -vV`, exit 0, **stdout** |
| normalization | `^release: (0\|[1-9][0-9]*)\.(0\|[1-9][0-9]*)\.(0\|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?$` → `(1, 91, 0)`, prerelease false |
| native target | `<root>/bin/rustc --print host-tuple` → `aarch64-apple-darwin`, exit 0 |
| companion probe | `<root>/bin/cargo --version` → `cargo 1.91.0 (ea2d97820 2025-10-10)` |
| metadata sources | `Cargo.toml` `package.rust-version`; `rust-toolchain.toml`; legacy `rust-toolchain` |

The verbose form is read rather than `rustc --version`, because the short form
puts the commit hash and date on the same line as the version. A test pins that
the `release:` rule does **not** match the short form, so the two can never be
swapped.

The Rust host tuple carries no platform-version component, so it is directly
usable as a native-target identity. Swift's is not — see §5.2.

### 4.2 Representability versus host gate

| Case | Cheap command | Build | Class |
|---|---|---|---|
| `rust-version = "1.85"` | `cargo metadata` exit 0 | `cargo build` exit **0** | none |
| `rust-version = "1.99"` | `cargo metadata` exit **0**, reports `rust_version: "1.99"` | `cargo build` exit **101**, `rustc 1.91.0 is not supported by the following package` | **host_version** |
| `rust-version = "not-a-version"` | `cargo metadata` exit **101**, `unexpected prerelease field, expected a version like "1.32"` | exit 101, same diagnostic | **representability** |

A future MSRV is fully representable; only the host relation rejects it. That is
the Rust analogue of Go's `go 1.99.0` case, and it is why a classifier built on
a build exit code alone would report a representable future release as
unrepresentable.

**Obligation for the Rust contract.** Read `packages[].rust_version` from
`cargo metadata --no-deps --format-version 1 --offline` in the graph phase and
compare it against the resolved toolchain identity. Cargo's own rejection is a
build-time error, so a driver that waits for it has entered the compile phase.

### 4.3 Native target: a preflight trap

| Target | `rustc --print target-libdir` | `<root>/lib/rustlib/<t>/lib` | `cargo build --target` |
|---|---|---|---|
| `aarch64-apple-darwin` | exit 0, path | present | builds |
| `x86_64-unknown-linux-gnu` | **exit 0, path** | **absent** | exit 101 **after** `Compiling probe-positive` |
| `not-a-real-target` | exit 1, `could not find specification for target` | absent | exit 101, `failed to run rustc to learn about target-specific information` |

`rustc --print target-libdir` returns exit 0 and a plausible path for a target
whose standard library is not installed. It is a representability answer wearing
a host-gate costume, and using it as a preflight would admit a build that then
fails with `E0463: can't find crate for 'std'` — the only case in this corpus
where upstream's rejection arrives **after compilation has started**
(`upstream_rejects_after_compilation: true`).

**Obligation for the Rust contract.** Membership in `rustc --print target-list`
(298 entries on the probed toolchain) is representability. Admission requires the
manager to stat `<root>/lib/rustlib/<target>/lib` inside the fingerprinted tree.
A driver that delegates target admission to cargo has already run a compiler on
package source.

### 4.4 `rust-toolchain.toml` is a live selector — measured both ways

Same directory, same file, two resolutions:

| File | Via rustup shim `~/.cargo/bin/rustc` | Via trusted root `<root>/bin/rustc` |
|---|---|---|
| `path = "/nonexistent/trusted/root"` | exit **1**, `invalid toolchain: the path '/nonexistent/trusted/root' has no bin/ directory` | exit **0**, `rustc 1.91.0` |
| `channel = "nightly"` | exit **1**, `info: syncing channel updates for 'nightly-aarch64-apple-darwin'` then a download failure | exit **0**, `rustc 1.91.0` |

The second row is the strongest single measurement in this run: a package-owned
file made the shim **attempt to download and install a toolchain**. The attempt
was made observable and impossible by pointing `RUSTUP_DIST_SERVER` and
`RUSTUP_UPDATE_ROOT` at a closed loopback port; `rustup toolchain list` was
identical before and after, so nothing was installed.

This is decision 0007 §3 measured rather than argued. `toolchain.path` is
`forbidden` because it names *where* a toolchain comes from; `toolchain.channel`
is `compared` precisely because a directly resolved root makes it inert.

### 4.5 Cargo picks `rustc` off `PATH`

Under a manager-owned minimal `PATH` that excludes the trusted root's `bin`,
every cargo invocation failed before doing anything:

```
error: could not execute process `rustc -vV` (never executed)
Caused by: No such file or directory (os error 2)
```

**Obligation for the Rust contract.** The driver MUST pass `RUSTC` explicitly as
the absolute path to the fingerprinted compiler. Otherwise the second node of the
`manager-worker-v2` process graph — "the driver's fingerprinted trusted launcher"
and the executables below it — is selected by `PATH` order rather than by the
fingerprinted closure, which is exactly what decision 0008 §6 item 3 forbids.
This is a real defect the probe found in its own first run.

---

## 5. Swift — `TASK-260728-1yhuqi`

### 5.1 Registry entry fields, measured

| Field | Measured value |
|---|---|
| version probe | `<root>/usr/bin/swiftc -print-target-info`, exit 0, **stdout**, JSON |
| normalization | applied to `compilerVersion`: `^(?:Apple )?Swift version (…)\.(…)(?:\.(…))?(-[0-9A-Za-z.]+)? \([^()]*\)$`, plus a `DEVELOPMENT\|SNAPSHOT` marker → `(6, 3, 2)`, prerelease false |
| native target | `target.unversionedTriple` = `arm64-apple-macosx` |
| metadata sources | `Package.swift` tools-version header; `.swift-version` |

**`swift --version` must not be the probe.** It splits one banner across two
streams: stderr carries `swift-driver version: 1.148.6 ` with **no trailing
newline**, stdout carries `Apple Swift version 6.3.2 (…)` and `Target: …`.
Anything that merges the streams sees them concatenated into a single line, so an
anchored rule over merged output and one over stdout disagree about the first
line. A test pins that the merged line does not match.
`swiftc -print-target-info` is JSON on stdout with no such hazard, and it carries
the target identity in the same call.

### 5.2 The versioned triple is not an identity

```
target.triple            arm64-apple-macosx26.0
target.unversionedTriple arm64-apple-macosx
target.moduleTriple      arm64-apple-macos
```

Host macOS is 26.5 and the selected SDK is 26.5, yet the triple reads `26.0`: the
platform component is the deployment target, not the host or the SDK. SwiftPM
compiles the manifest with a third value again (`arm64-apple-macosx14.0`, visible
in the recorded `dump-package` argv).

**Obligation for the Swift contract.** Bind `unversionedTriple` as the native
target. Binding `triple` would put a value into the cache key, receipt, and
marker that moves when the deployment default moves, without the toolchain
identity changing.

### 5.3 Representability versus host gate

| Case | `swift package tools-version` | `swift build` | Class |
|---|---|---|---|
| `// swift-tools-version:6.0` | exit 0, `6.0.0` | exit **0** | none |
| `// swift-tools-version:99.0` | exit **0**, `99.0.0` | exit 1, `using Swift tools version 99.0.0 but the installed version is 6.3.2` | **host_version** |
| `// swift-tools-version:not-a-version` | exit **1**, `misspelt or otherwise invalid` | exit 1, same | **representability** |

Same shape as Rust and Go. Unlike cargo, SwiftPM's own gate fires before it
compiles anything, so this preflight removes a wasted process rather than closing
a security gap.

### 5.4 Reading the manifest body executes package code — and the header does not

This is the measurement that decides how much of a Swift package a driver may
learn.

| Reader | Fixture: manifest whose body is `fatalError("PROBE-MANIFEST-EXECUTED")` |
|---|---|
| `swift package tools-version` | exit **0**, `6.0.0` |
| `swift package dump-package` | exit **1**, `Fatal error: PROBE-MANIFEST-EXECUTED` |

A second fixture, a manifest with a Swift **syntax error**, gives the same split:
`tools-version` exits 0 and returns `6.0.0`, `dump-package` exits 1 with a
compiler diagnostic. The recorded `dump-package` stderr contains the full
`swiftc` argv it used to build `Package.swift` into a `<name>-manifest`
executable, including `-plugin-path …/host/plugins/testing`.

So: `swift package tools-version` reads the header line and neither compiles nor
runs the manifest. Every other SwiftPM read command compiles and executes it.

**Obligation for the Swift contract.** The graph phase is limited to the
tools-version header. Products, targets, and configurations must come from the
manager-owned wire surface, not from the manifest — reading the manifest is
`build_package_code_execution_forbidden` by construction, not by policy choice.

### 5.5 Native target: the same trap, a different early gate

| Target | `-print-target-info` | `swiftc -target … -c` |
|---|---|---|
| `x86_64-unknown-linux-gnu` | **exit 0**, full JSON | exit 1, `unable to load standard library for target` |
| `not-a-real-triple` | exit 1, `unknown target` | exit 1, same |

`-print-target-info` accepts a target this toolchain cannot build for, exactly as
`rustc --print target-libdir` does.

**Obligation for the Swift contract.** The early gate is to stat the
`paths.runtimeLibraryPaths` that the same command reports: for a target the
toolchain cannot serve they do not exist inside the trusted tree. The probe
measures this and it resolves the case before any compiler runs.

### 5.6 `.swift-version` is inert against a directly resolved toolchain

With `.swift-version` containing `5.9.0`, `<root>/usr/bin/swift --version`
returned byte-identical output inside and outside the directory. The file is
honored by `swiftly`, which decision 0007 §3 already forbids as a resolution
origin. This is what lets the field be classified `compared` rather than
`forbidden`: the manager reads it as an assertion and never honors it.

Recorded honestly: `TOOLCHAINS=nonexistent.toolchain.id xcrun --find swiftc`
exited **0** and returned the default toolchain rather than failing. An unknown
selector value silently falls back, so a Swift driver must not treat "the
selector did not error" as evidence that no selector was applied — it must
neither set nor inherit `TOOLCHAINS` or `DEVELOPER_DIR`.

---

## 6. Kotlin Native — `TASK-260728-168smo`

### 6.1 Absent on both hosts

| Host | Searched | Result |
|---|---|---|
| macOS 26.5 arm64 | `kotlinc-native`, `konanc`, `~/.konan`, `/opt/homebrew/opt/kotlin/{bin,libexec/bin}` | **absent** |
| Windows 10.0.19045 amd64 | `kotlinc-native`, `konanc`, `%USERPROFILE%\.konan`, `C:\Program Files\Kotlin`, `C:\kotlinc`, sdkman | **absent** |

The Homebrew `kotlin` formula on the macOS host ships `kotlinc`, `kotlinc-jvm`,
`kotlinc-js`, `kotlinc-wasm`, `kapt`, `kotlin`, `kotlinr` — no native front end.
Kotlin/Native ships as a separate `kotlin-native-prebuilt-*` distribution.

Nothing was installed. Reported as evidence with the manager-owned guidance input
`toolchain.kotlin-native.absent.primary-source`, whose primary source is the
JetBrains Kotlin release distribution and whose `auto_install` is `false`.

### 6.2 The one Kotlin control that a host without Kotlin/Native can still run

```
$ /opt/homebrew/bin/kotlinc -version
exit 0
stdout: (empty)
stderr: info: kotlinc-jvm 2.4.10 (JRE 26.0.1)
```

Two hazards in one line:

1. the version goes to **stderr**, and **stdout is empty**. A probe that reads
   stdout reports the version as undetermined on a host where the compiler is
   present and working. A test pins this.
2. the banner names the **backend**. `kotlinc`, `kotlinc-jvm`, `kotlinc-js`,
   `kotlinc-wasm`, and `kotlinc-native` are near-identical banners, and only the
   native one can produce `native-executable-v1`.

The probe classifies this as `backend_identity` and reports the toolchain as
`wrong_backend` rather than `available` — a distinct state from `unavailable`,
because "a Kotlin compiler is here and it is the wrong one" is a different
operator problem from "nothing is here".

**Obligation for the Kotlin decision.** The `kotlin` registry entry's `probe`
must name the **stderr** stream, and its `normalization` must require the
`kotlinc-native` backend token rather than any `kotlinc` banner. Otherwise a
JVM front end silently satisfies a Kotlin requirement and the build produces a
JVM archive, which decision 0008 §3 rejects as `runtime-bundle`.

### 6.3 What is honestly not established

Four Kotlin cases are `not_run` on both hosts and are **not** presented as
passing:

| Case | Status |
|---|---|
| `kotlin.positive.version-and-target` | not run — no Kotlin/Native toolchain |
| `kotlin.negative.metadata-malformed` | not run |
| `kotlin.negative.metadata-future` | not run |
| `kotlin.negative.target-unknown` | not run |

The `kotlin.konanc.version.UNVERIFIED` normalization rule carries `UNVERIFIED` in
its identifier and `verified_against_real_output: false` in the fixture. It has
never been run against real output and must not be copied into a registry entry
as though it had been. No Kotlin metadata file or field is proposed here: naming
one without a qualified host would be exactly the fabricated platform claim
decision 0007 §1.3 defers these entries to avoid.

`TASK-260728-168smo` needs a host with a JetBrains Kotlin/Native distribution
before it can fill the entry, and per decision 0008 §2, if no Kotlin backend
satisfies the artifact class both Kotlin identifiers are retired unused.

---

## 7. Cross-language summary for the shared registry

| Property | `rust` | `swift` | `kotlin` |
|---|---|---|---|
| version probe | `rustc -vV` | `swiftc -print-target-info` | `kotlinc-native -version` (unverified) |
| stream carrying the version | stdout | stdout (JSON) | **stderr** |
| normalization verified against real output | yes | yes (Apple release form only) | **no** |
| native target source | `--print host-tuple` | `target.unversionedTriple` | unknown |
| native target stable as an identity | yes | only the unversioned form | unknown |
| metadata readable without executing package code | yes | **header only** | unknown |
| representability check for a target | `--print target-libdir` exit | `-print-target-info` exit | unknown |
| host gate for a target | stat `lib/rustlib/<t>/lib` | stat `paths.runtimeLibraryPaths` | unknown |
| upstream rejects a bad target only after compiling | **yes** | no | unknown |
| package file can redirect the compiler | **yes**, via the rustup shim | via `swiftly` / `TOOLCHAINS`, not measured to fire | unknown |
| package file can trigger an install attempt | **yes**, measured | not measured | Gradle wrapper, by construction |

Three properties generalize across all three languages and belong in the shared
contract rather than in each driver:

1. **A version probe must name its stream.** Three of the four probes measured
   here write the version somewhere other than "the obvious place": Rust's short
   form mixes in a commit hash, Swift splits a banner across two streams, Kotlin
   writes to stderr with an empty stdout.
2. **A compiler's target-print command answers representability, not admission.**
   Both Rust and Swift accept a target with exit 0 and then fail to build it. The
   admission check is a manager-side stat inside the fingerprinted tree in both
   cases.
3. **Every ecosystem ships a package-owned selector, and a directly resolved root
   is what neutralizes it.** Measured for Rust in both directions. The design
   consequence is the same for all three: resolve the concrete root, never the
   shim, and classify the selector file as an assertion.

---

## 8. Verification

| Command | Exit |
|---|---|
| `gofmt -l .` (probe module — empty output) | 0 |
| `go vet ./...` (probe module) | 0 |
| `go test -count=1 ./...` (probe module, 34 test functions) | 0 |
| `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run ./...` (probe module) — `0 issues.` | 0 |
| `toolchainboundary … -out fixture-macos.json` | 0 |
| `toolchainboundary -out fixture-windows.json` (on `win`) | 0 |
| determinism re-run — verdicts identical across two independent macOS runs | 0 |
| `go build ./...` (curator module) | 0 |
| `make vet` (curator module) | 0 |
| `make test` (curator module, full suite) | 0 |
| **`make check` (curator module)** | **2 — see below** |
| `git ls-files '*.go' \| xargs gofmt -l` — all tracked Go files formatted | 0 |
| `git status --porcelain` (tracked) — no tracked file changed, nothing staged | empty |

Full transcript with exit codes: `TASK-260729-rhjxtx_gate-log.txt`.

**`make check` fails, and it is not this task's failure — but it is reported as
a failure rather than as a pass.** The target is `check: vet test` followed by
`test -z "$(gofmt -l .)"`. Its `vet` and `test` stages both pass (recorded
separately above at exit 0). Its final `gofmt` stage scans the whole working
tree including `.temp/`, and finds seven unformatted files:

```
.temp/TASK-260720-1zntv0/cycle2/curator/vendor/github.com/lucasb-eyer/go-colorful/hsluv.go
.temp/TASK-260720-1zntv0/cycle2/curator/vendor/github.com/mattn/go-isatty/isatty_windows.go
.temp/TASK-260720-1zntv0/cycle2/curator/vendor/github.com/mattn/go-localereader/localereader_unix.go
.temp/TASK-260720-1zntv0/cycle2/curator/vendor/github.com/mattn/go-localereader/localereader_windows.go
.temp/TASK-260729-3jmqgl/worktree/prototypes/macos-hardened-probes/internal/probe/environment.go
.temp/TASK-260729-3jmqgl/worktree/prototypes/macos-hardened-probes/internal/probe/probe.go
.temp/TASK-260729-3jmqgl/worktree/prototypes/macos-hardened-probes/internal/spec/spec.go
```

Four are vendored third-party sources under `TASK-260720-1zntv0`'s scratch
directory and three are `TASK-260729-3jmqgl`'s worktree. **Zero belong to this
task**, and every tracked `.go` file in the repository is formatted. This task
therefore did not break `make check` and did not fix it; formatting another
task's scratch tree is outside this scope.

**Lint.** `golangci-lint` is not installed on this host, so the reproducible
`go run …@latest` invocation used by prior reviews in this repository was used.
Its first pass reported five issues in the probe module — one `errcheck`, one
`gosec`, three `revive` builtin shadowings. Four were fixed. The fifth, `gosec`
G204 on the child-process construction, carries an inline `//nolint` with a
written justification: launching an operator-named toolchain command is the
entire purpose of this program, and it never derives an executable path from
package data, `PATH`, or an inherited environment variable. The second pass
reports `0 issues.` at exit 0, and the probe was rebuilt and re-run afterwards
with an unchanged result (15 matched, 0 divergences, 4 not run, exit 0).

### Honest limitations

- **Kotlin/Native was never measured.** Four of nineteen cases are `not_run` on
  both hosts, the normalization rule is marked `UNVERIFIED`, and no metadata file
  or field is proposed. §6.3.
- **Windows has no toolchains.** All nineteen cases are `not_run` there. The
  Windows run demonstrates the probe is portable and that absence is reported as
  evidence; it establishes no Windows platform claim for any driver.
- **Only Apple's Swift was measured.** The open-source Swift release and
  development-snapshot banner forms are covered by the rule and by unit tests,
  but the test vectors are labelled unmeasured; a Swift contract must re-measure
  them on a host carrying a swift.org toolchain.
- **`TOOLCHAINS` was not measured to fire.** An unknown value fell back silently
  rather than selecting or failing, so the Swift selector row in §7 says "not
  measured to fire", not "does not select".
- **`rustc --print target-list` membership** was measured for six targets, not
  for all 298.
- **Rust MSRV inheritance** through `workspace.package.rust-version` is declared
  `compared` but was not exercised by a workspace fixture.

### Scope compliance

Only `.temp/TASK-260729-rhjxtx/` (gitignored) was written locally. No
`curator-spec` file, schema, vector, pin, dependency, or normative document was
touched. No package-controlled path, environment override, URL, channel, install
command, or trust root was introduced: the probe's guidance records carry a
primary-source URL and prose for the operator, never a command, and
`auto_install` is `false` in all three. On the Windows host, only
`C:\Users\admin\probe-rhjxtx\` was created, and it was removed after the run.
