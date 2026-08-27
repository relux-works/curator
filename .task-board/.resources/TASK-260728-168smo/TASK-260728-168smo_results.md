# TASK-260728-168smo — rework cycle 3 handoff

Ready for review.

## Verdict on the cycle-2 directive

The directive asked for one of two outcomes: qualify `windows/amd64` through
A1–A9 on the available root-capable Windows host and fill every registry field,
or retire both Kotlin identifiers now. **The tuple qualified.** The host was
used, the acceptance test was run end to end against checksum-verified upstream
inputs in a private temporary directory, and every field decision 0007 section
1.3 obliges this task to supply is now filled from that run rather than
asserted.

Everything the cycle-2 verdict asked to preserve is preserved: Kotlin/Native is
still the only selected artifact class, every Kotlin/JVM shape is still
deferred, `curator-kotlin-bundle-v1` is still the single-root trust model,
local and repository drivers still share one closed contract, and macOS is
still permanently unsupported.

## The four blocking findings

### 1 — No qualified tuple, so the required registry entry was missing

**Closed by qualification.** `windows/amd64` passed A1–A7 on
Windows 10 Pro 10.0.19045.6456, AMD64. A8 and A9 are discharged by the
conformance corpus rather than by a host run, because the allow-list walk
executes no compiler and the compatibility set is manager policy; both are
named as `TASK-260728-251p01` obligations.

The entry now reads: `platforms = {(windows, amd64)}`,
`compatibility = {(2, 4)}` with `(major, minor)` granularity,
`primary_relpath` `jdk\bin\java.exe` declared for `windows` and for no other
operating system, `probe` two vectors, two normalizations, `baseline`
`at_least 2.4.10`, empty companion list, `metadata_sources`
`kotlin-native-module.json` → `kotlin_version`.

`compatibility` carries one honest condition rather than a hidden one: decision
0007 section 1.1.1 admits a family only after it passes the driver's conformance
vectors, and those vectors land in the same change that mints manifest schema 8.
So the entry ships with `{(2, 4)}` in a change where the vectors pass, and if
they do not, the entry has no admissible family and both identifiers retire
unused. That is the only retirement branch left; the platform branch is
satisfied.

### 2 — The default native target was used but never probed

**Closed by making it a declared Stage A probe vector.** Decision 0007 section
1.1 declares `probe` as *the exact package-independent argument vectors* —
plural, as the `go` entry already uses — and Stage A step 6 already exists to
compare the toolchain's own reported host target against the native target. The
`kotlin` entry now declares P1 `konanc -version` and P2 `konanc -list-targets`,
with exact argv, stream, 4 KiB bound, line grammar, an exactly-one-`(default)`
rule, a closed token-to-claim map, ordering, and typed failures routed to the
two sites decision 0007 section 5 already declares — no new diagnostic code, no
new firing site, and no undeclared worker command, because both vectors are
manager-parent registry probes exactly like Go's three bootstrap vectors.

Measured: P2 exits 0, stdout 131 bytes, stderr 0 bytes, eight lines, exactly one
marked `mingw_x64 (default)`, which maps to `(windows, amd64)` and equals the
host pair. Both vectors also exit 0 against an **empty** `KONAN_DATA_DIR` and
leave 0 entries in it, so Stage A can run them before any cache lookup or
compiler work without touching the closure.

### 3 — The rejection matrix contradicted its own allow-list

**Closed in favour of the classifier.** Inert directories are admitted. Because
the walk is total, an admitted directory can contain only admitted directories
and `.kt` files, so `gradle/`, `kapt/`, `ksp/`, `META-INF/` and `services/` are
containers that can never carry a build script, a wrapper JAR, a plugin service
registration or any other non-`.kt` file. Dot-leading names stay rejected,
directories included.

A build-system directory-name deny-list was considered and refused for the same
reason the design refuses a filename deny-list: it cannot be exhaustive
(`gradle-8`, `Gradle`, `gradle.d`), it needs a case-folding policy the protocol
does not define, and it closes nothing the file-level rule does not already
close.

Section 7.2 is now an **ordered, total naming** of entries the allow-list has
already rejected: seven rows, first match wins, last row is the catch-all, so
exactly one diagnostic fires per rejected entry and the corpus can require a
case per row. A8 and the vector inventory were rewritten to match, and the
inventory now has an explicit admitted-cases group including an inert `gradle/`
and an inert `META-INF/services/`.

### 4 — The macOS "exactly seven" inventory was internally inconsistent

**Closed by reading the archive and stating only what the method proves.** The
cycle-2 archive is correct and the prose was wrong. `run7`, seeded empty,
established `/bin/bash` and `/usr/bin/xcrun` and its raw output additionally
shows `/usr/bin/xcode-select`, `/usr/libexec/PlistBuddy` and `xcodebuild`
denied. `run8`, seeded with the first four, reached exit 0 after adding `ld` and
`dsymutil`, so its final `allowed-externals.txt` legitimately holds **six**. The
**union observed** is seven; the set **sufficient** for a successful compile is
six; the difference is path-dependent, because an allowed `xcode-select` removes
the `xcrun xcodebuild -version` fallback. "Exactly seven" conflated the two and
is withdrawn.

Iterative denial proves required-on-the-path and sufficiency. It does not prove
completeness, so no completeness claim is made for macOS — and none is needed,
because one external executable already violates decision 0008 section 6 item 3
and ground two (cache identity) is independent of the process count.

The method gap is closed where it actually matters. **A3 now requires a
kernel-level process trace that records every process start in the window by
resolved absolute image path**, and forbids iterative denial as the basis for
*admitting* a tuple. On `windows/amd64` that trace was taken with ETW
`Microsoft-Windows-Kernel-Process` and `WINEVENT_KEYWORD_PROCESS`: 13 process
starts in the window, exactly **two** children below the compiler JVM —
`clang++.exe` and `ld.lld.exe`, both regular files inside `<kotlin_root>` — and
the tracer's completeness control fired on a deliberately external `where.exe`.
Sufficiency is enough to exclude a tuple; only completeness is enough to
qualify one.

## What the host run measured

| # | Result |
|---|---|
| Inputs | `kotlin-native-prebuilt-windows-x86_64-2.4.10.zip` 222,016,219 B, computed digest equals the published `.sha256`; Temurin `jdk-21.0.11+10` 205,073,954 B, computed digest equals the Adoptium API checksum |
| Vendor archive | 13 `bin` entries, all scripts; **0** `*.exe` anywhere in the distribution — decision 0007 section 3 cannot be satisfied by it on Windows either |
| Bundle | 27,867 files, 2,456,792,320 B, digest `63d96ff7…dae3`; the digest was reproduced after the tree was disturbed and rebuilt, so curation is byte-reproducible |
| P1 | exit 0, stdout 23 B `Kotlin/Native: 2.4.10\r\n`, stderr 50 B; macOS is the same line with LF, 22 B — absorbed by the line rule |
| P2 | exit 0, stdout 131 B, stderr 0 B, one `(default)` line, `mingw_x64` ⇒ `(windows, amd64)` |
| Hydration | exactly four dependencies, 465,614,911 B, from `download.jetbrains.com`, no integrity check reported |
| `.extracted` | **required**: absent ⇒ `Cannot find a dependency locally` exit 2 with zero downloads; restored 103 B ⇒ exit 0 |
| Bundle mutation | a compile against a writable bundle data dir adds exactly one file, `dependencies/cache/.lock`, 0 B; removing it restores the digest byte for byte |
| Write-denied bundle | compile fails `java.io.IOException: Access denied`; exec still works; append denied |
| Overlay | directory junctions, fresh writable `cache/`, `.extracted` copy ⇒ exit 0, bundle byte-unchanged |
| Process closure | 2 children, both in-bundle; tracer control fired |
| Artifact | `-o app` ⇒ `app.exe`, 570,368 B, PE32+ `0x20b`, machine `0x8664`, runs, exit 0, **no by-product**; staging held only the JVM's `hsperfdata_*` |
| Imports | plain program and platform-library program both import exactly `KERNEL32.dll`, `msvcrt.dll` |
| `PATH` | every compile ran with `PATH` = one empty manager-owned directory, exit 0; control `konanc.bat` exits 9009 on the `PATH` resolution of `java` |
| `airplaneMode=true` | fail-closed, exit 2, zero dependency-download lines |
| `@` tokens | the six-row expansion table reproduces the macOS semantics exactly; expansion is decided by the token's first character |

## Two changes of substance beyond the four findings

**The base-installation allow-list is what was measured, not what is
plausible.** For `windows/amd64` it is exactly `{KERNEL32.dll, msvcrt.dll}`.
`USER32.dll`, `ADVAPI32.dll`, `GDIPLUS.dll`, `OPENGL32.dll` and `WINHTTP.dll`
are all obviously part of a base Windows installation and are all reachable
through the distribution's other platform klibs, and none of them is in the set,
because no measured artifact imported one. Widening it is a re-qualification
run of A6 whose sample actually produces the import. A closed set smaller than
reality fails safe and is visibly extendable; a closed set built from assertion
fails open once, silently.

**Read-only is an ownership obligation, not a per-user deny — obligation
K-11.** Denying writes to the running user stopped the compiler writing into the
bundle. It did **not** stop a delete, because the same account held Full Control
through `BUILTIN\Administrators`, which grants `FILE_DELETE_CHILD` on the
parent. Curation therefore requires the bundle to be owned by an account the
manager does not run as, with the manager's account granted read and execute
only. This was found by a probe that actually deleted a bundle file, which is
also how the `.extracted` requirement surfaced.

## Board

No board element was created, retired, or re-linked in this cycle.
`TASK-260729-2vfvgi` (`run-windows-kotlin-driver-pair-qualification`), created in
cycle 2, keeps its scope with a narrower meaning: this record qualifies the
tuple against the reference pipeline, and that task re-runs A1–A7 against the
landed manager implementations and records any divergence as a qualification
regression. Decision 0010 section 14 states that explicitly.

Sequencing consequence for the story owner is unchanged and now firmer:
`TASK-260728-r3j8ef` and `TASK-260728-1aveb2` cannot run a Kotlin build on macOS
and must be sequenced onto the Windows host.

## Artifacts

| Name | Contents |
|---|---|
| `TASK-260728-168smo_decision-0010-kotlin-native-driver-pair.md` | decision rev 3 |
| `TASK-260728-168smo_kotlin-native-build-drivers-reference.md` | reference rev 3 |
| `TASK-260728-168smo_command-evidence.log` | cycle-3 measurement record, Windows W1–W16 plus the retained macOS record |
| `TASK-260728-168smo_probe.tar.gz` | reproduction package: every PowerShell suite run on the host, the raw host logs, the ETW parse, and a README with the exact command sequence |
| `TASK-260728-168smo_gate-log-cycle3.txt` | gate transcript with real exit codes |
| `TASK-260728-168smo_results.md` | this document |

## Gates

Real exit codes, each run standalone. See the gate log for the transcript.

| Gate | Exit |
|---|---|
| curator-spec `tools/validate.py` (venv, jsonschema 4.25.1) — baseline `57c1f56` | 0 |
| curator-spec `tools/validate.py` — task worktree | 1, **expected-red**, one broken link, attributed to the copied-in in-flight sibling `docs/external-build-repositories.md` |
| link sweep scoped to the two documents authored here | 0 — 6 local links, 0 broken |
| curator-spec `python -m unittest discover -s tools` | 0 — 8 tests |
| curator-spec `go test ./tools/...` | 0 |
| curator `go build ./...` | 0 |
| curator `go vet ./...` | 0 |
| curator `gofmt -l ./cmd ./internal` | 0 — 0 files |
| curator `go test ./...` | 0 |
| `golangci-lint` | **not run** — not installed on this host |

This task authored two Markdown documents and no Go, schema, vector, or release
file. No staging, commit, publication, pin, or toolchain installation.

## Host footprint

The Windows host was used through `ssh win` only. Everything lived in one
private directory, `C:\kn168`, which was removed after measurement: the two
verified archives, the 2.4 GB curated bundle and its 465 MB hydrated closure,
the overlays, the ETW traces, and the produced binaries. No software was
installed, no persistent `PATH` or profile was touched, `~/.konan` was never
created, the temporary write-deny ACL was lifted, the scoped outbound firewall
rule created for the network-denial measurement was deleted and its absence
verified, and free disk space was restored.
