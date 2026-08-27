# TASK-260728-168smo — rework cycle 2 handoff

Ready for review. Rework of the review cycle 1 changes-requested verdict.

## Verdict on the rework directive

The directive asked to keep the Kotlin/Native choice, the JVM rejection, the
closed package surfaces, the `@` defence and local/external equivalence, and to
replace the non-implementable official-archive-as-toolchain model with a closed
operator-curated bundle — then either pass the acceptance test for at least one
tuple or provide evidence that the pair must retire.

What happened:

- **The bundle model works and is measured.** An immutable, fingerprinted root
  containing the JDK, the distribution and a prehydrated dependency closure,
  driven through `<kotlin_root>/jdk/bin/java` with an operation-private writable
  overlay, compiles offline with both roots read-only and the closure
  byte-unchanged. Every registry, process, argv, environment, overlay, cache and
  identity field this task owns is now filled from measurement.
- **macOS cannot pass, and that is now proven rather than open.** Seven host
  executables outside every curatable root are required, no manager-fixed input
  removes them, and the cause is structural in the distribution's own data and
  code. macOS moves from "unverified" to **measured unsupported**, permanently,
  for Protocol 1.0.
- **The pair is not retired.** The evidence proves an *Apple-target* failure, and
  proves in the same file that the Windows and Linux toolchain paths are ordinary
  downloadable dependencies rather than host SDKs. Retiring on macOS evidence
  would be a fabricated negative about two tuples that were never measured, and
  retirement is one-way for the whole protocol version. The hard trigger at
  schema-8 minting is retained and sharpened.
- **Two review findings produced substantive contract changes**, not wording
  changes: the `jdk` companion is gone, and the "no C interop / stdlib only"
  claim is replaced by an allowed platform-library surface plus a new
  published-artifact dynamic dependency gate.

## What changed against cycle 1

| Cycle 1 | Cycle 2 |
|---|---|
| Official archive is the `kotlin` root; `primary_relpath` unresolved (shape (a)/(b)) | Operator-curated bundle `curator-kotlin-bundle-v1`; `primary_relpath` = `jdk/bin/java`, a regular executable inside the fingerprinted tree |
| "Narrow reading" of decision 0007 section 3 | **Withdrawn.** The invariant is satisfied literally |
| `jdk` a REQUIRED companion; `toolchain_identities` = `[kotlin, jdk]` | No companion at all; one root, one digest, one-element array. `jdk` stays reserved and unused |
| Offline closure asserted; `KONAN_DATA_DIR` an empty directory | Closure prehydrated by the operator, inside the fingerprint; `KONAN_DATA_DIR` an operation-private overlay; no-download proved twice |
| Probe/normalization/baseline open (K-7, K-8) | Measured: `konanc -version` → stdout `Kotlin/Native: 2.4.10`; anchored rule asserting the backend token; baseline `at_least 2.4.10` |
| K-4 output naming open | Measured: `-o app` → `app.kexe` plus an `app.kexe.dSYM` by-product |
| K-6 target discovery open | Measured: `konanc -list-targets` marks exactly one line `(default)` |
| "absolute paths do not expand" | Corrected: expansion is decided by the argv token's **first character**; `@/abs/path/file` does expand |
| "standard library only, no C interop" | Corrected: distribution platform libraries import with no `-library`/`.def` and change the artifact's dynamic dependencies. Surface allowed; new published-artifact gate binds the consequence |
| Every platform unverified, retirement deferred | macOS measured **unsupported** on two independent grounds; `windows/amd64` the single remaining candidate; Linux after `TASK-260728-1skseh` |

## The macOS exclusion, in one paragraph

Under exec containment allowing only the bundle roots, a hello-world
`-produce program -target macos_arm64` fails at `CurrentXcode.bash(Xcode.kt:144)`.
Iterating the allow-list to exit 0 yields exactly `/bin/bash`,
`/usr/bin/xcode-select`, `/usr/libexec/PlistBuddy`, `/usr/bin/xcrun`,
`<Xcode>/…/usr/bin/xcodebuild`, `<Xcode>/…/XcodeDefault.xctoolchain/usr/bin/ld`
and `…/dsymutil`, plus the Xcode SDK as sysroot. With
`ignoreXcodeVersionCheck=true` and every Apple property overridden to absolute
local paths, the compile still spawns `/usr/bin/xcrun` from
`AppleConfigurablesImpl.getAbsoluteTargetToolchain(Apple.kt:45)`. The cause:
`konan.properties` declares the Apple toolchain, sysroot and addon as
`remote:internal`, and `XcodePartsProvider` has exactly two implementations —
`InternalServer`, gated on `KONAN_USE_INTERNAL_SERVER` against
`https://repo.labs.intellij.net/kotlin-native`, and `Local`, the host's Xcode.
Independently, the Xcode SDK is an unfingerprinted build input, so two hosts
with different Xcode versions would alias in the cache.

## Consequence the story owner needs to see

macOS is the host this work is being done on, and it is now the one platform
this pair can never run on. Concretely:

- `TASK-260728-r3j8ef` and `TASK-260728-1aveb2` (local and external
  cross-manager Kotlin interop) cannot execute a Kotlin build on macOS and must
  be sequenced onto the qualified host.
- The four implementation tasks (`TASK-260728-1koh5v`, `TASK-260728-gmfxdg`,
  `TASK-260728-3ar1qp`, `TASK-260728-1uj0bc`) can be implemented and unit-tested
  anywhere, but their end-to-end behaviour is only observable where a tuple has
  qualified.
- The story had no Windows qualification task. One has been created
  (`TASK-260729-*`, see the board note) with a written justified-gap record;
  without it, `TASK-260728-251p01` would retire both identifiers by default
  rather than by evidence.

This is a sequencing judgement for the story owner, not a blocker on this task.

## Artifacts

| Artifact | What it is |
|---|---|
| `TASK-260728-168smo_decision-0010-kotlin-native-driver-pair.md` | decision 0010 rev 2 |
| `TASK-260728-168smo_kotlin-native-build-drivers-reference.md` | implementation-ready reference rev 2 |
| `TASK-260728-168smo_command-evidence.log` | E1–E19, argv, real exit codes, byte counts |
| `TASK-260728-168smo_probe.tar.gz` | reproduction package: sandbox profiles, enumerator, raw run logs, README |
| `TASK-260728-168smo_results.md` | this document |

Both spec documents were written into the task-owned worktree at
`.temp/TASK-260728-168smo/curator-spec-worktree`, which is clean of the baseline
`57c1f56` apart from the untracked in-flight records. Nothing was staged,
committed, published or pinned.

## Gates, real exit codes

| Gate | Exit | Note |
|---|---|---|
| curator-spec `tools/validate.py`, baseline `57c1f56` | **0** | 30 schemas, 93 vector files |
| curator-spec `tools/validate.py`, task worktree | **1** | expected-red: 3 broken links, all in two copied-in in-flight sibling docs (`docs/external-build-repositories.md` → `../release/1.0.0-rc.5.json`; `docs/portable-go-execution-policy.md` → `../conformance/v1/vectors/go-host-execution-policy.json`, twice). **0** from this task |
| link check scoped to the two documents authored here | **0** | 6 local links, 0 broken |
| curator-spec `python -m unittest discover -s tools` | **0** | 8 tests |
| curator-spec `go test ./tools/...` | **0** | |
| curator `go build ./...` | **0** | |
| curator `go vet ./...` | **0** | |
| curator `gofmt -l ./cmd ./internal` | **0** | 0 files |
| curator `go test ./...` | **0** | |

`golangci-lint` is not installed on this host and was not run. No Go, schema,
vector or release file was authored by this task.

## Host footprint

The Kotlin/Native distribution (889 MB) and its hydrated dependency closure
(688 MB) were downloaded into `.temp/TASK-260728-168smo/kn/` and **deleted after
measurement**; the archive was deleted after checksum verification. No toolchain
was installed, no PATH or shell profile was touched, `~/.konan` was never
created, and the host's free disk was restored. Only logs, sandbox profiles and
scripts remain, and they are inside the probe archive.
