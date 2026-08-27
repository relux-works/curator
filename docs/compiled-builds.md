# Compiled commands (schema 6, `go-v1`)

Schema 6 adds one thing to a skill package: a command whose executable Curator
compiles from source in the package itself, instead of shipping a script or
borrowing a binary from the host. The driver is closed and named `go-v1`.

This document is the complete authoring and operations reference. It is written
so that a skill author can produce a valid schema 6 package, and an operator can
predict every prerequisite and failure class, without reading Curator's source.

Everything in schemas 1 through 5 keeps working exactly as before. A package
that declares no `build_roots` and no `build` command is unaffected by this
document; see [What Curator manages](../README.md#what-curator-manages).

- [Protocol status](#protocol-status)
- [A complete mixed package](#a-complete-mixed-package)
- [`build_roots`](#build_roots)
- [The `build` command object](#the-build-command-object)
- [Go source prerequisites](#go-source-prerequisites)
- [The trusted toolchain](#the-trusted-toolchain)
- [What the build is not allowed to do](#what-the-build-is-not-allowed-to-do)
- [Trust boundary](#trust-boundary)
- [Logical identity and local paths](#logical-identity-and-local-paths)
- [Operating compiled commands](#operating-compiled-commands)
- [Running a compiled command](#running-a-compiled-command)
- [Failure classes](#failure-classes)
- [Verification](#verification)

## Protocol status

Curator implements the protocol published in
[relux-works/curator-spec](https://github.com/relux-works/curator-spec):
[`protocol/core.md`](https://github.com/relux-works/curator-spec/blob/main/protocol/core.md),
[`profiles/manager.md`](https://github.com/relux-works/curator-spec/blob/main/profiles/manager.md),
the JSON Schemas under
[`schemas/v1/`](https://github.com/relux-works/curator-spec/tree/main/schemas/v1),
and the conformance suite under
[`conformance/v1/`](https://github.com/relux-works/curator-spec/tree/main/conformance/v1).

CI checks this repository against one pinned protocol commit. That pin lives in
[`.github/workflows/ci.yml`](../.github/workflows/ci.yml) and currently names
`00b1688a9b2457ca397a0bb550acf47cad8ee967`, a `1.0.0-rc.3` revision. It is the
only protocol revision this repository claims conformance against.

The compiled-build contract described here — the schema 6 manifest surface, the
`curator-build-source-v1` and `curator-go-toolchain-v1` identities, the
`manager-worker-v1` execution policy, the compiled-artifact cache and receipt
format, and install marker schema 2 — is **accepted upstream but not yet part of
a published protocol revision**. There is no released tag carrying it, so this
document does not cite section numbers or vector files that would not resolve
today, and Curator does not claim conformance for it. When that revision is
published, the pin above moves and the citations become links; until then the
normative statements in this document describe implemented behavior, and the
build-driver vectors are consumed from a candidate suite only when
`CURATOR_CONFORMANCE_ROOT` points at one.

Building from a source repository other than the skill package itself is out of
scope here. It is tracked separately as a prospective schema 7 concern, is not
part of schema 6, and is not implemented: schema 6 compiles only sources that
live inside the package snapshot being installed.

## A complete mixed package

A schema 6 package may mix all three command types freely. This is a complete,
valid example: one compiled command, one script command, one system command.

```
example-skill/
├── agent-skill.json
├── SKILL.md
├── scripts/
│   └── report
└── build/                       <- build root: holds go.mod
    ├── go.mod
    ├── vendor/                  <- required once anything outside the standard library is imported
    └── cmd/
        └── indexer/             <- source_dir: the main package
            └── main.go
```

`agent-skill.json`:

```json
{
  "schema_version": 6,
  "runtime_roots": ["scripts"],
  "build_roots": ["build"],
  "capabilities": {
    "network": "none",
    "filesystem": "repo",
    "exec": "none"
  },
  "commands": {
    "indexer": {
      "type": "build",
      "driver": "go-v1",
      "source_dir": "build/cmd/indexer"
    },
    "report": {
      "type": "script",
      "unix_path": "scripts/report"
    },
    "git": {
      "type": "system",
      "command": "git",
      "hint": "install git"
    }
  }
}
```

`build/go.mod`:

```
module example.com/indexer

go 1.25
```

`build/cmd/indexer/main.go`:

```go
package main

import "fmt"

func main() { fmt.Println("indexed") }
```

After `curator install`, all three commands are reachable the same way, through
`.agents/bin/`. The agent cannot tell which one was compiled.

The legacy manifest filename `csk-skill.json` is still read for schema 6, and a
package may ship both files as long as their JSON values are equal. New packages
should write `agent-skill.json` only.

`agents/runtime.json` is the pre-schema fallback and is consulted only when
neither manifest exists. It never accepts a `build` command.

## `build_roots`

`build_roots` is a top-level list of POSIX-relative directories, admitted from
schema 6 onward. Each entry is a Go **main module root**: the directory that
holds the `go.mod` for the packages built below it.

Rules, all enforced at manifest load, before any toolchain runs:

| Rule | Diagnostic path |
|---|---|
| Every entry is a non-empty portable relative path inside the package | `build_roots[N]` |
| Every entry exists and is a directory | `build_roots[N]` |
| No path component of an entry is a symbolic link | `build_roots[N]` |
| Entries are unique | `build_roots` |
| Entries do not nest inside one another | `build_roots` |
| Entries do not overlap `runtime_roots` in either direction | `build_roots` |
| Every entry is used by at least one `build` command | `build_roots[N]` |
| `<entry>/go.mod` exists and is a real regular file | `commands.<name>.source_dir` |

`build_roots` is rejected outright below schema 6, as is any `build` command.

### Build roots never reach agent context

A build root holds compiler input, not prompt material. Curator excludes every
declared build root from agent-facing context exactly as it excludes
`runtime_roots`: the union of the two is removed from the installed context
whitelist. Nothing under `build/` is ever delivered to an agent, at any
activation mode.

`curator skill check <dir>` warns with `skill.build_root_in_prompt_context` when
prompt-visible markdown in the package references a build-root path, because
that path will not exist where the agent reads it. Write the command name, not
the source layout.

If a build root does reach agent-facing context after installation, `status`
reports `build-context-exposed` for that command.

## The `build` command object

A `build` command admits exactly three fields, and nothing else:

```json
{ "type": "build", "driver": "go-v1", "source_dir": "build/cmd/indexer" }
```

- `type` — `"build"`.
- `driver` — `"go-v1"`. It is the only accepted value; any other string is
  rejected at load, and any other recorded or planned driver is
  `unsupported-build-driver` at status time.
- `source_dir` — the POSIX-relative directory of the `main` package to compile.
  It must sit below **exactly one** `build_roots` entry (it may be that entry
  itself), must exist, and no component of it may be a symbolic link. The
  nearest `go.mod` walking up from `source_dir` must be the one directly at that
  build root: an intervening nested module is rejected.

The command name follows the ordinary identifier rule — start with a letter or
digit, then letters, digits, dots, underscores, or hyphens, no Windows-reserved
filename — because it becomes a shim filename.

Any additional field is refused with the exact path of the offending field, for
example `commands.indexer.ldflags`. This is deliberate and load-bearing: the
recognized attempts to reach the execution boundary are named surfaces, not
typos. `args`/`argv`/`arguments`, `env`/`environ`, `flags`/`tags`/`ldflags`/
`gcflags`, `toolchain`, `output`/`out`, `executable`/`program`, `hooks`/
`pre_build`/`post_build`, `plugin`/`plugins`, and `generate`/`generators`/
`scripts` are all rejected, whether they appear in the manifest or are smuggled
in later. A package never supplies an argument, an environment value, a flag, an
output path, a toolchain choice, or a lifecycle hook.

## Go source prerequisites

The build runs offline against checked-in sources. The package must satisfy all
of the following, or the build fails closed at the named boundary.

**Vendoring is mandatory.** Every non-standard-library dependency must resolve
from a checked-in `vendor/` tree under the build root. `GOPROXY` is `off`,
`GOVCS` is `*:off`, and the module cache is an empty operation-private
directory, so nothing is ever downloaded. A package that expects module
resolution fails with `vendor_dependency_missing`. Run `go mod vendor` in the
build root and commit the result:

```bash
cd build && go mod tidy && go mod vendor && git add vendor
```

A build root whose packages import only the standard library needs no `vendor/`
directory at all — `go mod vendor` reports `no dependencies to vendor` and
writes nothing, and the fixed `-mod=vendor` invocation succeeds without it. The
example above is such a package.

**One main package.** `source_dir` must resolve to exactly one importable `main`
package. Zero gives `build_package_not_main`; more than one gives
`build_package_ambiguous`.

**No workspace.** A `go.work` file anywhere under the build root is
`workspace_dependency_forbidden`. `GOWORK` is forced to `off` regardless.

**No toolchain directive.** A `toolchain` line in the build root's `go.mod` is
`toolchain_switch_forbidden`. `GOTOOLCHAIN` is `local`, so a package cannot
select or fetch a compiler.

**Pure Go only.** `CGO_ENABLED` is `0` and the following are refused anywhere in
the package graph:

| Input | Boundary code |
|---|---|
| Active cgo input | `cgo_required` |
| Other native input (C, C++, Objective-C, Fortran, SWIG) | `go_native_input_forbidden` |
| Non-standard assembly | `go_assembly_forbidden` |
| Pre-compiled host objects (`.syso`) | `go_syso_forbidden` |
| `//go:cgo_import_dynamic` | `go_forbidden_compiler_directive` |
| An active `//go:generate` directive | `go_generator_forbidden` |
| `default.pgo` | `go_pgo_forbidden` |

**Embedded inputs must be regular files inside the build root.** Every file a
`//go:embed` directive resolves to is validated like a source input: it must be
a regular file at an already-canonical path contained by the build root. A
symlink, a device or other special file, a non-canonical path, or anything that
resolves outside the build root is `go_embed_input_escape`. Embed from the
package directory itself rather than reaching across a link or a parent
directory; `embed.FS` over checked-in files under the package is fine. The same
violation inside a standard-library package is reported as
`go_standard_input_escape`, because there it means the selected toolchain is not
intact.

**Module metadata must be exact.** A `replace` directive, a failed or missing
module record, a nested or escaped main module (`nested_build_module`), or
vendored metadata pointing outside the build root (`go_module_input_escape`) all
fail closed.

**Native target only.** The build compiles for the host `GOOS`/`GOARCH`. There is
no cross-compilation surface, and the toolchain's reported host, target, and
version must all agree with the manager's platform (`target_mismatch`).

**Internal linking only.** `GO_EXTLINK_ENABLED` is `0` and the link mode is
internal with `libgcc=none`. There is no external linker step.

The exact compiler invocation is fixed and manager-owned. It is, in full:

```
go list  -mod=vendor -deps -json -buildvcs=false -compiler=gc -pgo=off .
go build -mod=vendor -trimpath -buildvcs=false -buildmode=exe -compiler=gc -pgo=off \
         -ldflags='-linkmode=internal -libgcc=none' -o <private staging path> .
```

One `go list`, one `go build`. Nothing else runs.

## The trusted toolchain

Curator never searches `PATH` for a Go installation and never downloads one. It
consults exactly three mechanisms, in this order, and stops at the first that is
set:

1. **`CURATOR_GO`** — the explicit operator override. It must be an absolute
   path naming `<GOROOT>/bin/go`, or `<GOROOT>\bin\go.exe` on Windows. The
   `GOROOT` is derived from it by taking the parent of `bin`. Anything else is
   `untrusted_go_executable`.
2. **`GOROOT`** — the trusted root; the launcher is derived as
   `<GOROOT>/bin/go`.
3. **The GOROOT the Curator binary itself was built against** — the compiled-in
   default. This is not an operator input; it is the fallback when neither
   variable is set.

`CURATOR_GO` and `GOROOT` are the only two operator mechanisms. Set `CURATOR_GO`
when the machine has several Go installations, or when the Curator binary was
built elsewhere:

```bash
CURATOR_GO=/usr/local/go/bin/go curator install
```

```powershell
$env:CURATOR_GO = 'C:\Go\bin\go.exe'; curator install
```

Whichever mechanism selected it, the result is then verified before use:

- The root and the launcher are resolved to real paths; the root must be a real
  directory and the launcher must resolve to `<resolved root>/bin/go`, or
  `toolchain_executable_mismatch`.
- The launcher must be a regular executable file with a native executable
  header. A shell wrapper is refused as `untrusted_go_executable`.
- The launcher must not live under the project or a runtime root, so a package
  cannot ship its own compiler.
- `go telemetry off`, `go version`, and one fixed `go env -json` run in an empty
  operation-private directory with an empty `PATH`. The release family must be
  allowlisted; `1.25` is the only family tested against the `go-v1` vectors, and
  anything else is `unsupported_go_family` rather than an approximation.
- The whole GOROOT tree is fingerprinted into a portable
  `curator-go-toolchain-v1` identity. That identity, not the path, enters the
  cache key. Absolute links, links escaping GOROOT, and special files are
  refused.
- The fingerprint is re-taken after the last build child exits. A toolchain that
  changed mid-operation is `toolchain_mutated` and the operation is refused.

`curator status` surfaces any of these as `unusable-build-toolchain`, carrying
the boundary code that refused it. `install --dry-run` reports
`toolchain-unavailable`, and neither builds nor mutates anything.

## What the build is not allowed to do

The following are unsupported by design, not yet-unimplemented. Each is refused
rather than approximated.

- **Lifecycle hooks.** No pre-build, post-build, install, or activation hook
  exists on any surface.
- **Package-supplied argv or environment.** The argument vector and the compiler
  environment are fixed and manager-derived. A package cannot add a flag, a
  build tag, an ldflag, a `GOFLAGS` value, or any other variable.
- **cgo.** `CGO_ENABLED=0`, always.
- **Go workspaces.** `GOWORK=off`, and a `go.work` under the build root is a
  refusal.
- **Network access of any kind during a build.** `GOPROXY=off`, `GOSUMDB=off`,
  `GOVCS=*:off`, `GONOPROXY=none`, `GONOSUMDB=none`. Vendored sources only.
- **External linking.** `GO_EXTLINK_ENABLED=0`, internal link mode,
  `libgcc=none`.
- **Root modules other than the declared build root.** The main module must be
  the build root itself; nested or escaped main modules are refused.
- **Toolchain selection or download by the package.** `GOTOOLCHAIN=local`, and a
  `toolchain` directive in `go.mod` is a refusal.
- **Code generation.** No `go generate`, no macros, no source-producing step.
- **Compiler or manager plugins.**
- **Cross-compilation.** Native host target only.
- **Any driver other than `go-v1`.** There is no generic or extensible driver
  surface in schema 6. Additional language drivers, if they are ever admitted,
  arrive as new closed driver identifiers under their own protocol revision, and
  a manifest naming one today is rejected at load.

Resource use is bounded per operation: a build deadline, a combined compiler
output bound, an artifact size bound, and — where the host provides the control —
file size, disk, memory, and process bounds. Exceeding one is a refusal
(`process_timeout`, `process_output_limit`, `artifact_size_limit`,
`process_disk_limit`), never a partial install.

## Trust boundary

**The compiled output is untrusted, and Curator never runs it.**

- Curator does not execute the artifact during `install`, `upgrade`, `status`,
  or `gc`. There is no smoke test, no `--version` probe, no post-build
  validation step that starts the binary. The build produces bytes; those bytes
  are hashed, verified, and published. Nothing more.
- The artifact runs only when a human or an agent invokes the command, through
  the shim, after installation.
- This is the same rule schemas 1 through 5 already state for skill content:
  Curator executes no package-supplied code at install time. Schema 6 compiles
  package-supplied code, which is a strictly larger surface than reading it —
  which is why the compiler runs under a fixed environment, an identity-verified
  worker re-execution of the installed manager, a bounded process domain, and no
  network.
- The build worker is an implementation boundary, not a command. It is not
  reachable through a package file, a manifest value, an environment value, a
  `PATH` lookup, a shell, or a user option.
- Cache entry content is never executed, adopted, or permission-repaired, and a
  receipt alone is never treated as proof of provenance.
- Untrusted detail — compiler output and cache reasons — is collapsed to one
  line, stripped of non-printable bytes, path-redacted, and length-bounded
  before it is printed or serialized.

Declaring a `build` command does not grant the package anything at build time.
`capabilities` still describes what the command may do when an agent runs it,
and is unrelated to what the compiler is allowed to touch.

## Logical identity and local paths

Two different things are deliberately kept apart.

**Portable logical identity** is what the protocol defines and what any
conforming manager derives identically:

- `curator-build-source-v1` — the domain-separated content identity of the
  immutable raw package snapshot.
- `curator-go-toolchain-v1` — the location-independent identity of the
  fingerprinted GOROOT tree, carrying the algorithm, the launcher relative path
  `bin/go`, the Go version, and the tree digest.
- The **logical cache key** — one opaque `sha256:` digest over the complete
  build input: schema version, driver, build-source identity, build root,
  command name, source directory, native target and its tuning, toolchain
  identity, and the fixed manager build policy.
- The **artifact path** — `bin/<command>`, or `bin/<command>.exe` when the
  target is Windows. This is a protocol-relative path, not a location.
- The **receipt** — the canonical record of one published build.
- **Install marker schema 2** — records `build_roots`, the build-source
  identity, and per-command `driver`, `cache_key`, `receipt_sha256`,
  `artifact_sha256`, and `artifact_path`. Marker schema 1 remains readable;
  because it cannot describe a compiled command, an installation that now
  activates one reports `needs-install` against a schema 1 marker.

The build input these identities compose contains no absolute path and no
timestamp, which is what lets two managers on two machines derive the same key
for the same package, toolchain, and target.

**Curator-local paths** are this implementation's business and are not protocol:

- The protected build cache lives under the Curator home at
  `cache/build/go-v1/<key>/`, where `<key>` is the hex of the logical key. Each
  entry holds the artifact at its protocol-relative path and a Curator-local
  receipt file, `curator-receipt.ccj.json`. The filename is deliberately local
  and is not part of the portable receipt schema or the cache identity.
- Operation-private compiler state — `GOPATH`, `GOMODCACHE`, `GOCACHE`,
  `GOTMPDIR`, `HOME`, `XDG_CONFIG_HOME`, `TMPDIR`, and the Windows equivalents —
  is created per operation under a manager-private base and removed afterwards.
- Staging directories, transaction journals, and the manager-home lock are
  likewise local.

Curator never publishes a manager-home, cache, staging, or probe path in a
diagnostic. Every path in `status` output is protocol-relative. Do not parse
Curator's local layout; it is not a compatibility surface.

## Operating compiled commands

The operator-facing behavior — `status` codes and causes, `status --json`,
`--check` semantics, `global status`, dry-run outcomes, install and upgrade as
the repair path, cache reversal guarantees, and the `gc` grace period — is
documented in the README:

- [Compiled-command status, diagnostics, and repair](../README.md#compiled-command-status-diagnostics-and-repair)
- [Maintenance and the build-cache grace period](../README.md#maintenance-and-the-build-cache-grace-period)

The short version, for authors:

```bash
curator install --dry-run    # plan only: no compiler, no cache or install mutation
curator install              # gates, build on miss, atomic commit, then a swept cache
curator status --json        # per-skill codes and a builds array
curator status --check       # non-zero unless every skill and compiled command is current
curator global status --check
curator gc                   # one serialized, locked maintenance pass
```

There is no `curator build` and no `curator repair`. `install` and `upgrade` are
the reconciliation path: they rebuild missing, corrupt, drifted, or untrusted
cache state, but only after every manifest, closure, collision, requirement,
audit, registry, and moved-tag gate has passed. A failed gate, preflight, build,
or commit leaves the previous installation, its consumers, and the live cache
unchanged, and says so.

Dry-run reports one of `cache-hit`, `would-preflight-and-build`,
`would-rebuild-untrusted-cache`, `corrupt`, `unsupported`, or
`toolchain-unavailable` per active build command. It is a plan, never a
completed compiler check.

`gc` removes an unreferenced protected cache entry only when it is provably
manager-protected, structurally exact, self-consistent with the key its
directory encodes, referenced by no marker and no in-flight transaction journal,
and published more than **24 hours** ago. Everything else is retained and
reported. The only cost of retention is disk; the only cost of a removal is one
rebuild.

## Running a compiled command

A compiled command is invoked exactly like a script command. The shim points at
the immutable protected cache artifact, never at a snapshot or a private path.

On Unix, `.agents/bin/<command>` is a `/bin/sh` launcher that prepends the
required directories to `PATH` when `PATH` is set, then `exec`s the target.
Because it uses `exec`, arguments, signals, and the exit status pass through
without an intervening shell:

```bash
.agents/bin/indexer --since 2026-01-01
echo $?    # the compiled command's own exit status
```

On Windows, `.agents\bin\<command>.cmd` is a `cmd` launcher that sets `PATH` the
same way, `call`s the target with `%*`, and exits with `%ERRORLEVEL%`:

```bat
.agents\bin\indexer.cmd --since 2026-01-01
echo %ERRORLEVEL%
```

Global installation additionally publishes non-destructive forwarding shims to a
user directory already on `PATH` when one is available; otherwise Curator prints
the canonical global bin location. No shell profile change is required for
either scope.

## Failure classes

Two vocabularies, deliberately separate.

**Currentness codes** are what `status` reports per skill and per active
compiled command. They are stable and machine-readable, they appear verbatim in
`status --json`, and `--check` exits non-zero for every value except
`up-to-date` and `current`. The full table, with the `build-input-drift` cause
subcodes, is in
[the README](../README.md#compiled-command-status-diagnostics-and-repair).

**`go-v1` boundary codes** are what the driver refuses with, and they name the
exact rule that was violated. Where one surfaces depends on when it fires: a
refusal before any logical identity exists — that is, at toolchain selection,
probing, or fingerprinting — is carried as the `cause` of an
`unusable-build-toolchain` status row, while a refusal during a real build fails
that `install` or `upgrade` and is reported as its bounded, path-redacted
reason, leaving the previous installation intact. This is the complete set,
grouped by what an author or operator should do:

| Group | Codes | What to do |
|---|---|---|
| Toolchain selection | `untrusted_go_executable`, `go_toolchain_missing`, `toolchain_executable_mismatch`, `unsupported_go_family` | Point `CURATOR_GO` at a real `<GOROOT>/bin/go` of a tested release family |
| Toolchain integrity | `toolchain_mutated`, `toolchain_unreadable`, `toolchain_timeout`, `toolchain_link_absolute`, `toolchain_link_escape`, `toolchain_link_dangling`, `special_file_forbidden`, `duplicate_path`, `invalid_unicode` | Do not modify or replace the Go installation during an operation; reinstall a clean toolchain |
| Toolchain probe | `malformed_go_version`, `invalid_go_env`, `go_version_failed`, `go_env_failed`, `target_mismatch`, `telemetry_initialization_failed`, `telemetry_directory_untrusted` | The selected installation is not a stock Go release; select one that is |
| Package graph | `build_package_not_main`, `build_package_ambiguous`, `go_list_failed`, `go_list_malformed`, `go_list_incomplete`, `go_test_input_forbidden`, `build_module_missing`, `nested_build_module`, `go_standard_input_escape`, `go_source_input_escape`, `go_source_unreadable` | Fix `source_dir` so it names exactly one readable `main` package under the declared build root |
| Embedded inputs | `go_embed_input_escape` | Embed only regular files that live inside the build root; see [Go source prerequisites](#go-source-prerequisites) |
| Vendoring | `vendor_dependency_missing`, `vendor_metadata_inconsistent`, `go_module_input_escape` | Run `go mod vendor` in the build root and commit `vendor/` |
| Forbidden source | `cgo_required`, `go_native_input_forbidden`, `go_assembly_forbidden`, `go_syso_forbidden`, `go_forbidden_compiler_directive`, `go_generator_forbidden`, `go_pgo_forbidden` | Remove the non-pure-Go input; see [Go source prerequisites](#go-source-prerequisites) |
| Forbidden declaration | `workspace_dependency_forbidden`, `toolchain_switch_forbidden` | Remove `go.work` from the build root, and the `toolchain` line from `go.mod` |
| Package influence | `build_execution_package_influence_forbidden` | Remove the extra field from the build command object |
| Compilation | `go_build_failed`, `external_link_forbidden`, `libgcc_fallback_forbidden` | The code does not compile under internal-only linking; fix it, or remove the input that forces a host linker |
| Artifact verification | `artifact_output_invalid`, `artifact_mutated`, `artifact_unreadable`, `artifact_special_file`, `artifact_link`, `artifact_permissions_failed`, `source_mutated` | The build produced something other than one verifiable regular file, or the snapshot changed mid-operation; re-run on a quiescent checkout |
| Execution domain | `build_execution_worker_identity_invalid`, `build_execution_worker_protocol_invalid`, `build_execution_control_unavailable`, `build_execution_capability_evidence_invalid`, `build_execution_hardened_claim_forbidden` | Do not replace the Curator binary during an operation; reinstall Curator |
| Resource bounds | `process_timeout`, `process_output_limit`, `process_disk_limit`, `artifact_size_limit` | The build exceeds a manager bound; reduce what it compiles |
| Private state | `private_probe_failed`, `private_probe_cleanup_failed`, `private_build_failed`, `private_build_cleanup_failed`, `private_build_unreadable`, `private_build_special_file`, `process_environment_poisoned` | The manager-private base is unusable; check permissions and free space |
| Manager request | `invalid_build_request`, `invalid_resource_limits` | An internal invariant was violated; report it as a Curator bug |

Manifest problems never reach the driver: they are refused at load with the
exact JSON path of the offending value, for example `commands.indexer.driver` or
`build_roots[0]`.

## Verification

Every command below runs from the repository root.

| Command | What it proves | Artifacts |
|---|---|---|
| `make build` | The manager compiles | `bin/curator` |
| `go build ./...` | Every package compiles | — |
| `go vet ./...` | Static analysis is clean | — |
| `gofmt -l cmd internal` | Formatting is clean (empty output) | — |
| `golangci-lint run` | Lint is clean | — |
| `go test ./...` | The whole suite, including documentation examples | — |
| `go test ./internal/skillspec/` | Schema 1 through 6 manifest parsing and diagnostics | — |
| `go test ./internal/godriver/` | Toolchain selection, worker execution, package-graph guards | — |
| `go test ./internal/buildcache/ ./internal/buildmeta/ ./internal/buildsource/` | Cache protection, logical identity, snapshot identity | — |
| `go test ./cmd/curator/ -run TestDocumented` | The examples and vocabularies in this document and the README | — |
| `go test ./internal/interop/ -v` | The shared protocol suite | — |

Every command in that table runs green with `CURATOR_CONFORMANCE_ROOT` unset:
each conformance test skips itself when it is, so the local loop needs no
protocol checkout.

The conformance and interop tests read the authoritative suite from
`CURATOR_CONFORMANCE_ROOT` when it is set:

```bash
git clone https://github.com/relux-works/curator-spec ../curator-spec
CURATOR_CONFORMANCE_ROOT=$PWD/../curator-spec/conformance/v1 go test ./...
```

Against the published suite that command currently **fails** in
`internal/buildsource` and `internal/buildcache`, and it should: those two tests
require `vectors/build-drivers.json`, the published suite does not carry it yet,
and a missing vector file is a hard failure rather than a skip so a manager
cannot quietly claim unverified build-driver conformance. Point
`CURATOR_CONFORMANCE_ROOT` at a candidate suite that carries the build-driver
vectors to run them, or leave it unset to run everything else. Aligning the
committed CI pin with a published suite that carries those vectors is tracked
separately and is not something this document can assert.

CI runs the gofmt, vet, test, lint, interop, and naming gates on ubuntu, macos,
and windows; see [`.github/workflows/ci.yml`](../.github/workflows/ci.yml).

Scratch output, logs, and task-scoped worktrees belong under `.temp/`, which is
not versioned. Board artifacts live under [`.task-board/`](../.task-board/), and
the wider tool table is in [the README](../README.md#tools-and-verification).
