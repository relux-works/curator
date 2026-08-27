# Compile-only build drivers: protocol research and `go-v1` recommendation

**Task:** `TASK-260720-poa3ze`  
**Date:** 2026-07-20  
**Status:** Research outcome for review  
**Scope:** Read-only inspection of the three requested `origin/main` revisions; no product or specification source was modified.

## Key takeaways

1. **Add one closed command form in manifest schema 6:** `{"type":"build","driver":"go-v1","source_dir":"..."}`. The package can select only a source directory and the closed driver identifier. It cannot supply a program, arguments, environment, output path, tags, linker flags, build script, or hook.
2. **Ship only `go-v1` in protocol v1.** The manager invokes an operator-trusted Go toolchain directly, without a shell, using fixed `go list` and `go build` argument vectors, vendor-only modules, no network, `GOTOOLCHAIN=local`, `GOWORK=off`, and `CGO_ENABLED=0`. It never invokes the built executable.
3. **Treat build outputs as a distinct immutable cache.** The current runtime identity of `skill + commit` is insufficient. A build key must cover the complete snapshot, command/source selection, driver revision, native target, fixed policy, and exact toolchain fingerprint. Every artifact gets a canonical receipt, and install marker v2 pins the receipt and artifact hashes.
4. **Build before any installation mutation.** Resolve, validate, gate, and inspect caches first; build all misses in an operation-private staging area; then commit cache entries and all manager-owned project state as one rollback-capable transaction. Record the consumer only after the transaction commits; run garbage collection last.
5. **Dry-run must not compile.** It may inspect the snapshot, the trusted toolchain identity, and existing receipts read-only, but it must not run `go list` or `go build`, create Go caches, publish artifacts, or mutate registry/audit/runtime/project state.
6. **Do not generalize this into “run a build tool.”** Cargo build scripts and procedural macros, SwiftPM manifests/plugins, Make/CMake commands, Maven/Gradle plugins, MSBuild tasks, npm lifecycle scripts, Python build backends, and similar mechanisms can execute package-selected host code during a build. Those violate the invariant by construction.
7. **The current managers expose pre-existing lifecycle issues that the build change must not copy forward.** Both record a consumer before materialization and both trust an existing runtime cache directory without verifying its required paths. The Python dry-run also appears able to mutate audit-registry cache/state before returning. These are implementation findings, not reasons to weaken the proposed contract.

## Research basis

The repositories were fetched immediately before inspection. All code claims below refer to immutable objects, not the working trees.

| Repository | Requested ref | Inspected commit | Commit subject |
|---|---|---|---|
| `relux-works/curator-spec` | `origin/main` | [`57c1f56846d221ecc55786bd3c2467ec32f11730`](https://github.com/relux-works/curator-spec/tree/57c1f56846d221ecc55786bd3c2467ec32f11730) | Pin landed agent-skill implementations (#12) |
| `relux-works/curator` | `origin/main` | [`17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`](https://github.com/relux-works/curator/tree/17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8) | Pin landed rc.3 protocol |
| `ivanopcode/cocoaskills` | `origin/main` | [`6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`](https://github.com/ivanopcode/cocoaskills/tree/6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12) | Pin landed rc.3 protocol |

Primary language and build-tool documentation was used for the security classification. The source register at the end identifies the exact claims checked.

## 1. Current protocol and manager behavior

### 1.1 Normative baseline

The current protocol already establishes the correct security boundary:

- Skill packages and manifests are untrusted, and an installer **must not execute package-provided code while resolving, validating, auditing, or installing** ([core protocol, lines 63–69](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/protocol/core.md#L63-L69)).
- Manifests are strict JSON validated by schema and semantic rules; unknown fields fail ([core protocol, lines 9–20](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/protocol/core.md#L9-L20)).
- Schema versions 1–5 expose only `script` and `system` commands ([core protocol, lines 123–165](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/protocol/core.md#L123-L165); [common schema](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/schemas/v1/common.schema.json#L49-L72)). A script is copied; a system command is resolved. There is no compilation model.
- Dependency closure uses deterministic provider-first topological order ([core protocol, lines 262–280](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/protocol/core.md#L262-L280)).
- The profile requires all non-mutating checks before materialization, a side-effect-free dry-run, rollback on materialization failure, and consumer/GC work last ([manager profile, lines 50–80](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/profiles/manager.md#L50-L80)).
- A runtime cache is currently keyed by skill and commit. Existing entries may be reused only after required paths are verified ([manager profile, lines 91–112](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/profiles/manager.md#L91-L112)).
- The install marker contains source/context identity but no build identity or artifact receipt ([install-marker-v1 schema](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/schemas/v1/install-marker-v1.schema.json)).
- Capability declarations and successful audits are visibility and policy gates, not a sandbox and not proof that package code is safe to execute ([security model, lines 13–38](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/SECURITY.md#L13-L38)).

The build-driver design therefore must preserve a stronger rule than “the manifest declared execution”: installation cannot hand control to package-selected programs at all.

### 1.2 Go manager (`curator`) at `origin/main`

The Go manager supports schemas 1–5 and only script/system commands ([skill types](https://github.com/relux-works/curator/blob/17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8/internal/skillspec/types.go)). Runtime installation copies declared roots or one script and writes a shim; it has no compiler path ([installer, lines 353–420](https://github.com/relux-works/curator/blob/17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8/internal/install/install.go#L353-L420)). The marker model has no build receipt ([marker model, lines 47–70](https://github.com/relux-works/curator/blob/17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8/internal/marker/marker.go#L47-L70)).

Observed issues relevant to the new lifecycle:

- Consumer state is written before any node is materialized ([installer, lines 338–377](https://github.com/relux-works/curator/blob/17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8/internal/install/install.go#L338-L377)), although the profile places consumer records after successful materialization.
- Runtime installation returns immediately when `<home>/runtime/<skill>/<commit>` exists, without validating the active command paths ([runtime store, lines 23–64](https://github.com/relux-works/curator/blob/17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8/internal/runtimestore/runtimestore.go#L23-L64)). This is weaker than the profile’s required-path check.
- Context replacement is rollback-capable per directory, but the enclosing install is not one transaction across runtime, every context, shims, environment files, adapters, and consumer state ([installer, lines 353–455](https://github.com/relux-works/curator/blob/17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8/internal/install/install.go#L353-L455); [marker replacement](https://github.com/relux-works/curator/blob/17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8/internal/marker/marker.go#L299-L379)). A later failure can therefore leave earlier nodes changed.

The Go dry-run does return before the materialization block and uses read-only audit/registry modes; this is the behavior to retain.

### 1.3 Python manager (`cocoaskills`) at `origin/main`

The Python manager has the same schemas 1–5 and script/system-only model ([skill specification](https://github.com/ivanopcode/cocoaskills/blob/6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12/src/csk/skillspec.py#L131-L238)). It copies runtime roots/scripts into `<home>/runtime/<skill>/<commit>` and writes shims ([shim/runtime implementation](https://github.com/ivanopcode/cocoaskills/blob/6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12/src/csk/shims.py#L46-L120)). It also has no build receipt.

Observed issues relevant to the new lifecycle:

- Consumer state is written immediately after the dry-run return but before runtime/context materialization ([installer, lines 175–209](https://github.com/ivanopcode/cocoaskills/blob/6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12/src/csk/installer.py#L175-L209)).
- An existing runtime directory is accepted without required-path validation ([shims, lines 46–83](https://github.com/ivanopcode/cocoaskills/blob/6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12/src/csk/shims.py#L46-L83)).
- Materialization is per node and per target, not a project-wide rollback transaction ([installer, lines 193–280](https://github.com/ivanopcode/cocoaskills/blob/6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12/src/csk/installer.py#L193-L280)).
- **Inferred dry-run defect:** `_check_audit_registries` is called before the dry-run return, and its implementation creates/migrates persistent registry cache/state without a read-only flag ([call site](https://github.com/ivanopcode/cocoaskills/blob/6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12/src/csk/installer.py#L175-L191); [registry check](https://github.com/ivanopcode/cocoaskills/blob/6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12/src/csk/installer.py#L455-L533)). This should be verified with a mutation assertion and corrected independently of build-driver work.

### 1.4 Consequence for implementation planning

The two managers agree closely enough that one protocol contract is feasible. However, compilation must not simply be inserted into each existing per-node materialization loop. Doing so would amplify existing partial-install and cache-trust weaknesses. Transaction staging, cache validation, and consumer ordering are prerequisites of the implementation design.

## 2. No-hooks threat model

### 2.1 Assets and adversary control

Treat all of the following as attacker-controlled input:

- the repository snapshot, including source, `go.mod`, `go.sum`, `vendor/`, filenames, build constraints, compiler directives, embedded files, and file sizes;
- both manifest spellings and all manifest fields;
- refs, registry data, cached source/runtime/build entries, receipts, and stale install markers until validated;
- any executable artifact produced from the package;
- output text emitted while the compiler parses the package.

The trusted computing base is deliberately small:

- the manager binary and protocol implementation;
- the operator-selected operating system and native Go toolchain, resolved independently of the package;
- manager-owned policy constants, staging directories, cache locks, and transaction journal;
- the cryptographic hash and canonical JSON implementations.

The package must not influence toolchain selection through its snapshot, manifest, current directory, `PATH`, Go environment files, workspace files, or module `toolchain` directive.

### 2.2 Security invariant

During install, update, repair, status, GC, and dry-run, a conforming manager:

- **MUST NOT** invoke a package-provided executable, script, interpreter entry point, shared library, compiler plugin, generator, test, package-manager hook, build backend, or arbitrary build recipe;
- **MUST NOT** load package-produced code into the manager or compiler process as a plugin or macro;
- **MUST NOT** use a shell to construct a build command;
- **MUST NOT** launch the newly built artifact, including for version discovery, smoke testing, post-processing, or receipt generation;
- **MAY** pass untrusted source bytes to the fixed, trusted `go-v1` compiler pipeline described below;
- **MAY** perform fixed data transformations, hashing, copying, permission changes, and schema validation in manager code.

“Compile-only” means the manager owns every executable and argument in the build process. It does not claim that a compiler is immune to malicious inputs. Compiler vulnerabilities and resource-exhaustion inputs remain threats, so the child process must receive filesystem, network, time, output-size, process-count, memory, and disk controls where the host supports them. The resulting binary remains untrusted package code and is executed only later, when a user explicitly invokes the installed command.

### 2.3 Explicitly forbidden operations

The following are never part of `go-v1`:

```text
sh -c, cmd.exe /c, powershell, make, cmake, ninja
go generate, go run, go test, go install, go get
go mod download, go mod tidy, go mod vendor
package-supplied -tags, -ldflags, -gcflags, -asmflags, -toolexec, -overlay
package-supplied environment variables or output paths
execution of the output binary
```

## 3. Manifest contract

### 3.1 Recommended schema shape

Add schema version 6 to both canonical manifest names. Preserve the existing `command` definition for schemas 1–5; changing it in place would accidentally allow build commands in older manifests. Add these definitions to `common.schema.json` and let only the v6 schemas reference `commandV6`:

```json
{
  "snapshotDirectory": {
    "oneOf": [
      {"const": "."},
      {"$ref": "#/$defs/portablePath"}
    ]
  },
  "buildCommandV6": {
    "type": "object",
    "required": ["type", "driver", "source_dir"],
    "properties": {
      "type": {"const": "build"},
      "driver": {"const": "go-v1"},
      "source_dir": {"$ref": "#/$defs/snapshotDirectory"}
    },
    "additionalProperties": false
  },
  "commandV6": {
    "oneOf": [
      {"$ref": "#/$defs/scriptCommand"},
      {"$ref": "#/$defs/systemCommand"},
      {"$ref": "#/$defs/buildCommandV6"}
    ]
  }
}
```

The v6 schema is the v5 shape with `schema_version: 6` and `commands.additionalProperties` pointing to `common.schema.json#/$defs/commandV6`:

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://relux-works.github.io/curator-spec/schemas/v1/agent-skill-v6.schema.json",
  "title": "agent-skill.json schema 6",
  "type": "object",
  "required": ["schema_version", "capabilities"],
  "properties": {
    "schema_version": {"const": 6},
    "runtime_roots": {"$ref": "common.schema.json#/$defs/pathSet"},
    "commands": {
      "type": "object",
      "propertyNames": {"$ref": "common.schema.json#/$defs/identifier"},
      "additionalProperties": {"$ref": "common.schema.json#/$defs/commandV6"}
    },
    "capabilities": {"$ref": "common.schema.json#/$defs/capabilities"},
    "dependencies": {"$ref": "common.schema.json#/$defs/dependenciesV5"}
  },
  "additionalProperties": false
}
```

`csk-skill-v6.schema.json` is an exact legacy-name mirror with its own `$id` and title, as today.

### 3.2 Valid manifest example

This example deliberately mixes a compiled command with an existing script command; command collision and activation rules apply uniformly to both.

```json
{
  "schema_version": 6,
  "runtime_roots": ["scripts"],
  "commands": {
    "golden-tool": {
      "type": "build",
      "driver": "go-v1",
      "source_dir": "cmd/golden-tool"
    },
    "skill-helper": {
      "type": "script",
      "unix_path": "scripts/skill-helper",
      "win_path": "scripts/skill-helper.cmd"
    }
  },
  "capabilities": {
    "network": "none",
    "filesystem": "repo",
    "exec": "none",
    "secrets": "none",
    "env_read": []
  },
  "dependencies": {}
}
```

Root-module commands use `"source_dir": "."`; no empty-string alias is permitted.

### 3.3 Invalid package-controlled arguments

This object must fail schema validation because `args` is unknown. The same applies to `env`, `flags`, `tags`, `output`, `toolchain`, `build_script`, and `post_build`.

```json
{
  "type": "build",
  "driver": "go-v1",
  "source_dir": "cmd/golden-tool",
  "args": ["-tags", "package_choice"]
}
```

### 3.4 Semantic validation rules

After JSON Schema succeeds, the manager must enforce all of these rules:

1. A `build` command is legal only in manifest schema 6. Legacy `agents/runtime.json` commands cannot express it.
2. `driver` must equal the closed identifier `go-v1`. Unsupported drivers fail; there is no fallback to a system build command.
3. `source_dir` is either exactly `.` or a portable relative path. After path joining and canonicalization it must be a real directory inside the immutable snapshot. Every component must satisfy the existing no-link/no-special-file snapshot rules.
4. The nearest ancestor `go.mod` must exist and its directory must remain inside the snapshot. `GOWORK=off` means a `go.work` file never expands the source boundary.
5. The selected directory must resolve through fixed `go list` semantics to exactly one buildable package named `main`. Package patterns, multiple packages, and library-only packages fail.
6. A command object cannot combine script, system, and build properties. Unknown fields fail before semantics.
7. The manager derives the output basename from the manifest command key. Unix output is `bin/<command>`; Windows output is `bin/<command>.exe`. The package cannot override it.
8. Build commands participate in the existing activation, dependency-command selection, portable-name collision, and shim collision rules exactly like script commands.
9. Active build command names are processed in Unicode code-point/bytewise lexical order within each provider-first closure node. The same normalized ordering is used in receipts and markers.
10. The raw snapshot content hash—not only `source_dir`—is part of build identity. This intentionally favors safe invalidation over maximum cache hits.
11. A build source may be under a `runtime_root`, but generated output never is: outputs live only in manager-owned staging/cache directories and are never injected into prompt context.
12. A manager that cannot enforce the fixed environment, offline module mode, native target, `CGO_ENABLED=0`, and transaction rules must reject the build command rather than approximate it.

No package or project field is added for arbitrary build arguments. No build-policy override is added to manager or system configuration in v1.

## 4. Fixed `go-v1` driver

### 4.1 Supported shape

`go-v1` builds one native executable from one main package:

- host target only; no cross-compilation in v1;
- Go module mode only; the module root must be in the snapshot;
- standard-library-only dependencies or dependencies already present in the main module’s top-level `vendor/` tree;
- cgo disabled;
- no workspaces, generators, tests, plugins, overlays, or external linker;
- one output selected by the manager.

Missing/inconsistent vendored dependencies, a newer required Go version, a package that needs cgo, or a non-main package are deterministic install errors.

### 4.2 Toolchain selection and identity

The manager resolves the Go executable before entering any package-controlled directory. The candidate must be bundled with the manager or selected by trusted operator configuration/environment; it must not come from the repository, `runtime_roots`, project `.agents/bin`, or package manifest. The manager resolves symlinks and records an absolute path and `GOROOT`.

Run only these package-independent probes, with the sanitized environment below:

```json
["/absolute/trusted/go", "version"]
```

```json
["/absolute/trusted/go", "env", "-json", "GOROOT", "GOHOSTOS", "GOHOSTARCH", "GOOS", "GOARCH", "GO386", "GOAMD64", "GOARM", "GOARM64", "GOMIPS", "GOMIPS64", "GOPPC64", "GORISCV64", "GOWASM"]
```

The native target is frozen from this trusted probe and explicitly passed to subsequent commands. A protocol-quality toolchain fingerprint is not just a version label: compute a canonical path-and-content SHA-256 over all regular files and link targets in the resolved `GOROOT`, plus the exact `go version` stdout. This covers patched/custom compilers and standard-library sources that could otherwise share a version string. An implementation may memoize that digest in manager-owned state, but it must invalidate the memo on any uncertain metadata change; dry-run may read but not create that memo.

The cache/receipt records both `go_version` and `toolchain.content_sha256`. `GOTOOLCHAIN=local` makes a package `toolchain`/`go` directive a compatibility check, never permission to download or switch toolchains ([Go toolchain selection](https://go.dev/doc/toolchain)).

### 4.3 Environment construction

Start from an empty environment. Add only OS variables indispensable for process creation (for example `SYSTEMROOT` on Windows), a trusted minimal `PATH`, locale normalization, operation-private directories, resolved `GOROOT`, the explicit native target, and the following Go policy:

```json
{
  "GO111MODULE": "on",
  "GOENV": "off",
  "GOFLAGS": "",
  "GOMODCACHE": "<operation-private>/gomodcache",
  "GOCACHE": "<operation-private>/gocache",
  "GOTMPDIR": "<operation-private>/gotmp",
  "GOPROXY": "off",
  "GOSUMDB": "off",
  "GOPRIVATE": "",
  "GONOPROXY": "none",
  "GONOSUMDB": "none",
  "GOVCS": "*:off",
  "GOWORK": "off",
  "GOTOOLCHAIN": "local",
  "GOTELEMETRY": "off",
  "CGO_ENABLED": "0",
  "GOEXPERIMENT": "",
  "GOROOT": "<resolved-trusted-goroot>",
  "GOOS": "<native-goos>",
  "GOARCH": "<native-goarch>",
  "HOME": "<operation-private>/home",
  "TMPDIR": "<operation-private>/tmp",
  "LC_ALL": "C",
  "LANG": "C"
}
```

Set exactly the target tuning variable applicable to `GOARCH` (`GO386`, `GOAMD64`, `GOARM`, `GOARM64`, `GOMIPS`, `GOMIPS64`, `GOPPC64`, `GORISCV64`, or `GOWASM`) to the trusted probe result; omit the rest. On Windows, set `TEMP` and `TMP` to the same private temporary directory. Do not inherit `CC`, `CXX`, `PKG_CONFIG`, `AR`, `GCCGO`, `GO_EXTLINK_ENABLED`, `GOAUTH`, user `PATH`, or any other Go variable.

The source snapshot is mounted/readable but not writable by the child. Only operation-private cache/temp/output directories are writable. Standard input is closed; stdout/stderr are captured with size limits. Apply a deadline, memory/disk/process limits, and OS network denial. Environment denial is normative even when a stronger OS sandbox is available.

### 4.4 Exact process construction

Set the child working directory to the canonical `source_dir`. Invoke the resolved Go binary directly with argument vectors—never a joined command string.

Preflight:

```json
["/absolute/trusted/go", "list", "-mod=vendor", "-json", "."]
```

Parse exactly one JSON package result and require `Name == "main"`. Validate its module directory remains within the snapshot. Then build:

```json
["/absolute/trusted/go", "build", "-mod=vendor", "-trimpath", "-buildvcs=false", "-buildmode=exe", "-o", "<operation-staging>/bin/golden-tool", "."]
```

On Windows the final operand is `<operation-staging>/bin/golden-tool.exe`. All arguments except the absolute trusted tool path and manager-derived staging path are protocol constants. `.` is deliberately the only package operand.

After a successful child exit, require one regular output file, not a link or directory. Check it is within staging, impose a maximum size, set manager-defined executable permissions, hash it, and never run it. Compiler output is diagnostic data only and must be bounded/redacted before presentation.

### 4.5 Module and network behavior

The Go module reference states that `-mod=vendor` loads dependencies from the main module’s `vendor` directory instead of the network or local module cache and checks `vendor/modules.txt` consistency with `go.mod` ([Go Modules Reference](https://go.dev/ref/mod#vendoring)). Therefore:

- `go-v1` always supplies `-mod=vendor` to both fixed commands;
- `GOPROXY=off`, `GOSUMDB=off`, and `GOVCS=*:off` provide defense in depth;
- the operation-private module cache begins empty and is discarded;
- the manager never repairs, downloads, tidies, or vendors dependencies;
- standard-library-only modules need no populated vendor tree; external dependencies must already be correctly vendored in the snapshot;
- inconsistent `go.mod`/`vendor/modules.txt`, missing packages, remote-only replacements, or any attempted lookup fail the build;
- `GOWORK=off` prevents a repository or parent workspace from changing module resolution;
- `GOTOOLCHAIN=local` prevents automatic toolchain download.

`go generate` is not automatically run by `go build` or `go test` ([Go generate documentation](https://pkg.go.dev/cmd/go#hdr-Generate_Go_files_by_processing_source)), and the driver never invokes it.

### 4.6 cgo

cgo may invoke C/C++ compilers and `pkg-config`, and package source can carry `#cgo` directives ([cgo documentation](https://pkg.go.dev/cmd/cgo)). `go-v1` therefore fixes `CGO_ENABLED=0` and does not inherit compiler variables. Source files requiring cgo are excluded by Go’s normal build selection; if the remaining package cannot build, installation fails.

There is no “best effort” cgo mode. A future cgo driver would need a separate identifier, declared trusted compiler/linker set, sysroot identity, fixed flags, stronger sandbox, and its own security review.

### 4.7 Determinism claim

The v1 guarantee is deterministic **driver semantics and cache identity**, not universal bit-for-bit equality across different Go distributions or targets. For the same validated snapshot, driver, native target tuple, fixed policy, and byte-identical toolchain fingerprint, the manager constructs the same environment and arguments. `-trimpath` removes local source paths, and `-buildvcs=false` prevents ambient VCS metadata from entering the binary. The receipt records the actual artifact hash, so unexpected divergence is detected rather than normalized away.

## 5. Cache identity and receipts

### 5.1 Why `skill + commit` is insufficient

The current runtime key does not distinguish:

- Go toolchain upgrades or patched distributions;
- OS/architecture/tuning changes;
- cgo/module/network policy revisions;
- driver semantic revisions;
- command/source selection;
- a corrupted or incomplete existing cache directory.

Compiled artifacts therefore use a separate content-addressed cache. Existing script runtime storage may remain, but its required-path validation must be fixed.

### 5.2 Build input identity and cache key

For each active build command, construct this logical input object. This is an illustrative Darwin/arm64 instance; hashes are placeholders of the correct shape.

```json
{
  "schema_version": 1,
  "driver": "go-v1",
  "snapshot_sha256": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "command": "golden-tool",
  "source_dir": "cmd/golden-tool",
  "target": {
    "goos": "darwin",
    "goarch": "arm64",
    "tuning": {
      "GOARM64": "v8.0"
    }
  },
  "toolchain": {
    "go_version": "go version go1.26.1 darwin/arm64",
    "content_sha256": "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
  },
  "policy": {
    "module_mode": "vendor",
    "network": "none",
    "workspace": false,
    "cgo": false,
    "target_mode": "native"
  }
}
```

`snapshot_sha256` is the existing protocol content hash over the complete raw snapshot. Canonicalize the input object with the protocol’s canonical JSON rules and set:

```text
cache_key = "sha256:" + lowercase_hex(SHA-256(canonical_utf8(input)))
```

The driver identifier is versioned (`go-v1`), so any change to command arguments, environment, sandbox-relevant semantics, receipt interpretation, or output rules requires a new driver identifier or an explicitly versioned cache-key policy revision.

### 5.3 Immutable cache layout

Use the key without its `sha256:` prefix in the filesystem:

```text
<manager-home>/build-cache/go-v1/<64-hex-key>/
  .csk-build.json
  bin/<command>[.exe]
```

Entries are immutable after atomic publication. Build into an operation-private sibling directory, acquire a per-key lock, then rename. If another process wins, discard the loser and validate the winner. Never merge files into an existing entry.

### 5.4 Receipt example

The receipt is canonical JSON with no timestamps or absolute paths:

```json
{
  "schema_version": 1,
  "cache_key": "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
  "input": {
    "schema_version": 1,
    "driver": "go-v1",
    "snapshot_sha256": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    "command": "golden-tool",
    "source_dir": "cmd/golden-tool",
    "target": {
      "goos": "darwin",
      "goarch": "arm64",
      "tuning": {
        "GOARM64": "v8.0"
      }
    },
    "toolchain": {
      "go_version": "go version go1.26.1 darwin/arm64",
      "content_sha256": "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
    },
    "policy": {
      "module_mode": "vendor",
      "network": "none",
      "workspace": false,
      "cgo": false,
      "target_mode": "native"
    }
  },
  "artifact": {
    "path": "bin/golden-tool",
    "sha256": "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
    "size": 1234567
  }
}
```

Receipt validation is semantic as well as schema-based:

1. Recompute the cache key from canonical `input`; it must match both the receipt and directory name.
2. Match the entire expected input object, not selected fields.
3. Require the artifact path to equal the manager-derived platform path.
4. Open without following links, require one regular file, bound its size, and recompute SHA-256 and byte length.
5. Reject unknown receipt fields and unsupported schema/driver versions.
6. Never execute an artifact to validate it.

A malformed/mismatched entry is cache corruption. Under the per-key lock, quarantine it outside the live namespace or replace it only via a fresh staged build. Do not silently use it. A dry-run reports corruption but neither quarantines nor rebuilds.

### 5.5 Install marker v2

Introduce `install-marker-v2.schema.json`. New managers read marker v1 and v2, write v2 on any installation mutation, and may continue to regard a valid v1 marker as current for schema 1–5 packages. Marker v2 raises `skill_schema_version` to 6 and requires `builds` (an empty object for installs without compiled commands).

Illustrative complete marker:

```json
{
  "schema_version": 2,
  "name": "golden-skill",
  "source": "skills/golden-skill",
  "ref_kind": "tag",
  "ref": "v1.0.0",
  "commit": "0123456789abcdef0123456789abcdef01234567",
  "content_sha256": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "locale": null,
  "agents": [],
  "commands": ["golden-tool"],
  "dependencies": [],
  "skill_schema_version": 6,
  "runtime_roots": [],
  "installed_at": "2026-07-20T00:00:00Z",
  "files": ["SKILL.md", "agent-skill.json"],
  "builds": {
    "golden-tool": {
      "driver": "go-v1",
      "cache_key": "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
      "receipt_sha256": "sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
      "artifact_sha256": "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
      "artifact_path": "bin/golden-tool"
    }
  }
}
```

Marker currentness requires every active build command to match the marker and a valid cache receipt/artifact. A missing, corrupt, wrong-target, or wrong-toolchain cache entry makes status non-current and repair rebuilds it. Canonical marker writing sorts `commands`, `dependencies`, `files`, and `builds` keys.

The shim must point to the immutable build-cache artifact selected by the marker/plan, not to `<home>/runtime/<skill>/<commit>`. Shims remain manager-generated, self-contained, and independent of shell profiles.

### 5.6 Garbage collection

GC marks build cache keys referenced by all valid marker v2 files and in-flight transaction journals. It then sweeps unreferenced entries older than the normal grace period. If a marker cannot be parsed safely, GC retains potentially related entries and reports the uncertainty. It never uses receipt content alone as proof of a live consumer.

## 6. Lifecycle, ordering, dry-run, and rollback

### 6.1 Normative ordering

Serialize installs for the same project and use provider-first closure order. A conforming implementation performs:

1. Resolve refs and immutable snapshots; parse both manifest spellings; validate schemas and semantic paths.
2. Build the dependency closure and activation plan; check command/shim/case collisions and system/MCP requirements.
3. Run source audit, registry policy, attestation, and moved-tag gates without executing package code.
4. Resolve/fingerprint the trusted Go toolchain and compute every active build input/cache key. Within a node, sort build command names lexically.
5. Validate cache receipts/artifacts read-only and produce a complete hit/miss/corrupt plan.
6. **Dry-run ends here.**
7. Create one operation-private staging tree and transaction journal. For every cache miss in provider-first/node-command order, run fixed preflight/build commands. Do not publish anything yet.
8. Verify and hash every staged artifact; generate canonical receipts. If any build fails, delete operation staging and leave installation, consumer records, and permanent caches unchanged.
9. Acquire per-key cache locks, revalidate races, and atomically publish missing immutable cache entries.
10. Stage every manager-owned project/global target: contexts/markers, shims, runtime copies, environment files, adapter ledgers, hybrid targets, cleanup sets, and consumer ledger update.
11. Commit those targets using an explicit rollback journal and deterministic order. Existing targets are backed up until every swap and the consumer-ledger update succeeds.
12. On success, remove backups and transaction state. Run stale-cache/runtime GC last; GC failure is a reported maintenance warning and does not corrupt the committed installation.

The command build order affects observability only; cache keys make outputs independent. Independent builds may later run in parallel, but result/diagnostic ordering and commit ordering remain the normative sequence.

### 6.2 Dry-run

Dry-run may:

- parse/hash the snapshot;
- resolve an operator-trusted Go binary;
- run package-independent `go version`/`go env` probes, or use an already validated fingerprint;
- read receipts, artifacts, markers, registry data, and policy state;
- report the target, driver, key, and `cache-hit`, `would-build`, `corrupt`, or `unsupported` result for each command.

Dry-run must not:

- invoke `go list`, `go build`, or any compiler/linker;
- create `GOCACHE`, `GOMODCACHE`, temp build directories, fingerprint memos, audit records, registry response caches, runtime/build cache entries, transaction journals, shims, markers, or consumer records;
- quarantine a corrupt cache entry;
- change atime where the platform/API allows no-atime reads.

This also means the Python registry path identified above needs a genuine read-only mode before it can host build-driver dry-runs.

### 6.3 Rollback rules

- A resolution, validation, gate, toolchain, cache-inspection, or build failure occurs before commit and leaves the installed state byte-for-byte unchanged.
- A cache publication or target-swap failure restores project/global manager-owned targets from the journal in reverse order, including prior shims, environment files, adapters, contexts, markers, and consumer ledger.
- Newly published cache entries are removed on rollback if still unreferenced and owned by this transaction. If a concurrent successful transaction has begun referencing an identical valid entry, it is retained and will be governed by GC.
- Existing valid immutable cache entries are never modified during rollback.
- Backups are retained until rollback succeeds. An interrupted transaction is recovered under the same project lock on the next manager start before any new plan runs.
- The built artifact is never used as a rollback program or verifier.

This is a project-wide transaction requirement, not merely per-skill directory replacement.

## 7. Candidate language/toolchain classification

The test is not “can this ecosystem build offline?” It is “can the manager invoke a fixed trusted compiler without allowing the package to select executable host code, hooks, plugins, or argument arrays?”

| Ecosystem | v1 classification | Security rationale and possible constrained future |
|---|---|---|
| **Go** | **Accept as `go-v1`** | Direct `go list`/`go build` supports one fixed package operand. Vendor mode, local toolchain, disabled cgo/workspaces/VCS stamping, cleared environment, and no `go generate` close known package-selected execution paths. Output is never launched. |
| **Rust** | Defer/reject Cargo | Cargo compiles and executes package `build.rs` before building the package, and procedural macros run during compilation with compiler-process resources. A future driver would need to reject build scripts, proc-macro dependencies, compiler wrappers, linker selection, and Cargo config; direct `rustc` still needs a dependency/linking design. [Cargo build scripts](https://doc.rust-lang.org/cargo/reference/build-scripts.html); [procedural macros](https://doc.rust-lang.org/reference/procedural-macros.html). |
| **Zig** | Defer | `zig build` evaluates package `build.zig` and can run system/project tools. Zig also deliberately evaluates `comptime` code during compilation, which needs a precise side-effect/resource model before acceptance. A future fixed `zig build-exe` subset cannot inherit build.zig/package arguments. [Zig build system](https://ziglang.org/learn/build-system/); [Zig language reference](https://ziglang.org/documentation/master/#comptime). |
| **Swift** | Defer | SwiftPM package manifests are Swift code that runs, and command/build-tool plugins execute separate processes. A future direct `swiftc` subset would need to forbid package plugins/macros, own SDK/target/link flags, and solve dependency discovery without SwiftPM. [SwiftPM plugins](https://docs.swift.org/swiftpm/documentation/packagemanagerdocs/plugins/). |
| **C/C++** | Defer | Make recipes run via a shell; CMake can run processes during configure and declare custom build commands. A future direct compiler driver could own an exact source-file enumeration, sysroot, preprocessor/linker flags, and disable plugins, but portable dependency and linker identity is substantially larger than Go v1. [GNU Make recipe execution](https://www.gnu.org/software/make/manual/html_node/Execution.html); [CMake `execute_process`](https://cmake.org/cmake/help/latest/command/execute_process.html); [CMake custom commands](https://cmake.org/cmake/help/latest/command/add_custom_command.html). |
| **Java/Kotlin** | Defer | Maven and Gradle run package-selected plugins/tasks. `javac` discovers and invokes annotation processors unless processing is disabled; Kotlin supports compiler plugins and kapt processors. A future direct `javac -proc:none` subset could own classpath/output/main wrapping, but mixed JVM builds are not v1-safe. [javac annotation processing](https://docs.oracle.com/en/java/javase/23/docs/specs/man/javac.html); [Maven plugins](https://maven.apache.org/guides/introduction/introduction-to-plugins.html); [Gradle tasks](https://docs.gradle.org/current/userguide/implementing_custom_tasks.html); [Kotlin compiler plugins](https://kotlinlang.org/docs/compiler-plugins-overview.html). |
| **.NET** | Defer | `dotnet build` delegates to MSBuild; project/imported files can select executable tasks, inline compiled tasks, and `Exec`. Roslyn source generators are compile-time metaprograms. A future direct `csc` subset must disable analyzers/generators and own references/target/runtime packaging. [MSBuild tasks](https://learn.microsoft.com/en-us/visualstudio/msbuild/msbuild-tasks); [inline tasks](https://learn.microsoft.com/en-us/visualstudio/msbuild/msbuild-inline-tasks); [`Exec`](https://learn.microsoft.com/en-us/visualstudio/msbuild/exec-task); [Roslyn compiler platform](https://learn.microsoft.com/en-us/dotnet/csharp/roslyn-sdk/). |
| **Node/TypeScript** | Defer | npm install lifecycle scripts and default `node-gyp` behavior execute package-selected code. A fixed trusted TypeScript compiler could transpile a closed source graph, but output is not self-contained and dependency installation/runtime resolution remains unsafe and non-deterministic under this contract. [npm scripts](https://docs.npmjs.com/cli/using-npm/scripts/). |
| **Deno** | Promising, but defer | `deno compile` is closer to a direct driver and offers cached/offline controls, but current behavior includes config-driven options, preload/require execution, npm lifecycle permissions, and framework detection that may run a build script. A future version-pinned driver needs one explicit entry file, `--no-config`, no preload/require, frozen pre-vendored dependencies, no scripts, and audited behavior. [Deno compile](https://docs.deno.com/runtime/reference/cli/compile/); [Deno packages](https://docs.deno.com/runtime/packages/). |
| **Python** | Defer | pip/build frontends delegate wheel creation to the package-selected PEP 517 backend and invoke backend hooks; build requirements may also be computed dynamically. That is package-controlled host execution. `compileall` only produces bytecode and does not solve dependencies or a self-contained command. [pip build-system interface](https://pip.pypa.io/en/latest/reference/build-system/); [PyPA packaging flow](https://packaging.python.org/en/latest/flow/). |

The protocol should document this table as rationale, not reserve generic driver names. Each future driver needs a separate threat-model review, closed schema, fixed process graph, cache identity, and conformance vectors.

## 8. Protocol artifacts affected

### 8.1 Normative protocol and schemas

| Artifact | Required change |
|---|---|
| `protocol/core.md` | Add schema 6; define build command semantics, no-hooks boundary, build/source/output rules, cache/receipt identity, marker v2, and artifact currentness. |
| `profiles/manager.md` | Add driver discovery, fixed `go-v1` process/environment, build planning/order, read-only dry-run, whole-install transaction/rollback, build cache, locks, receipts, shims, repair/status, and GC. Correct consumer ordering and required-path reuse language where needed. |
| `SECURITY.md` | Clarify trusted compiler versus untrusted source/output; list forbidden hooks/plugins/build systems; document compiler-input DoS and sandbox expectations. |
| `schemas/v1/common.schema.json` | Add `snapshotDirectory`, `buildCommandV6`, `commandV6`, receipt/marker supporting definitions. Do **not** broaden the existing `command` union. |
| `schemas/v1/agent-skill-v6.schema.json` | New canonical manifest schema referencing `commandV6`. |
| `schemas/v1/csk-skill-v6.schema.json` | New legacy-name mirror. |
| `schemas/v1/build-receipt-v1.schema.json` | New strict receipt schema for canonical input and artifact identity. |
| `schemas/v1/install-marker-v2.schema.json` | New reader/writer shape with `skill_schema_version <= 6` and required `builds`. |
| `schemas/v1/README.md` | Index and compatibility notes for all new schemas. |

No v1 build-policy fields should be added to `manager-config-v1` or `system-config-v1`; fixed semantics are the security feature. `protocol/registry.md` and audit-record schemas need no new artifact attestation in v1: registry audit continues to attest the untrusted source snapshot, while the local receipt binds the compiled result to that snapshot and trusted toolchain. The security text should state this distinction explicitly.

### 8.2 Conformance artifacts

Add or update:

- `conformance/v1/schema-cases/index.json`;
- valid/invalid cases for `agent-skill-v6`, `csk-skill-v6`, `build-receipt-v1`, and `install-marker-v2`;
- a separate Go fixture with `go.mod`, `cmd/golden-tool`, and a vendored-dependency variant, rather than changing the existing script golden fixture and all of its registry hashes;
- expected canonical build input, cache key, receipt, marker, artifact SHA-256, and dry-run plan;
- `conformance/v1/manifest.json` and `conformance/README.md`;
- `conformance/v1/vectors/manager-lifecycle.json` for provider/build order, no-mutation dry-run, build failure, cache race/corruption, interrupted transaction recovery, rollback, currentness, repair, and GC;
- preferably a focused `conformance/v1/vectors/build-drivers.json` for environment/argv/target/cache-key semantics;
- release-facing `README.md`, `COMPATIBILITY.md`, and `CHANGELOG.md`.

Minimum negative vectors:

- build command in schema 5;
- unknown driver or any `args`/`env`/`output` field;
- source path escape/link/non-directory/no `go.mod`/non-main/multiple-package mismatch;
- dependency absent from `vendor`, inconsistent `modules.txt`, workspace-only dependency, newer auto-toolchain request, and cgo-only package;
- attempted `go:generate` file proving no generator execution;
- poisoned `PATH`, `GOFLAGS=-toolexec=...`, `GOENV`, `GOWORK`, VCS metadata, parent workspace, and repository-local fake `go`;
- cache key mismatch, wrong target/toolchain, receipt/artifact hash mismatch, link/special artifact, partial entry, and concurrent publisher;
- build 2 failure after build 1 succeeds with no persistent changes;
- target-swap failure with full reverse rollback and unchanged consumer ledger;
- dry-run assertions covering every forbidden persistent path, including Python registry response/state caches.

## 9. Decision and anomaly log

These findings are important beyond the schema shape and should be tracked during implementation/review:

1. **Decision:** `go-v1` is a closed protocol driver, not an adapter to arbitrary Go commands. Any semantic change creates a new driver revision.
2. **Decision:** vendored, native, cgo-disabled builds are the only v1 dependency/target model.
3. **Decision:** build output uses a separate immutable cache and receipt; `<skill>/<commit>` runtime identity is not reused for compiled artifacts.
4. **Decision:** marker v2 pins build receipts/artifacts; old marker v1 remains readable for pre-v6 packages.
5. **Decision:** dry-run does not invoke source-aware Go commands, even if an implementation could isolate their caches.
6. **Anomaly:** both managers update consumer state before successful materialization, contrary to the documented phase order.
7. **Anomaly:** both managers accept an existing runtime directory without verifying required paths, contrary to the manager profile.
8. **Anomaly:** both managers’ per-node/per-target replacement can leave a partially changed installation after a later failure; build work requires a whole-install journal.
9. **Probable regression:** the Python dry-run calls a registry path that appears to create/migrate persistent cache/state. Add a failing mutation test before changing it.

## 10. Clear v1 recommendation

Adopt manifest schema 6, build receipt schema 1, and install marker schema 2 with **only** the `go-v1` driver defined above. Implement the transaction/cache hardening in both managers before enabling v6 packages. Reject all generic hooks, build executables, argument/environment arrays, package-manager frontends, cgo, downloads, workspaces, and cross-compilation.

This is narrow enough to implement and test deterministically while preserving the protocol’s defining invariant: installation compiles untrusted Go source with a fixed trusted driver, but never transfers execution control to package-provided code.

## 11. Fact-check register

All external claims were checked against primary project/vendor documentation on 2026-07-20. No secondary comparison article is used as authority.

| Claim checked | Primary source | Finding used |
|---|---|---|
| Go vendoring and network/module cache | [Go Modules Reference](https://go.dev/ref/mod#vendoring) | `-mod=vendor` loads dependencies from main-module `vendor/`, not network/local module cache, and validates `modules.txt` consistency. |
| Go automatic toolchain switching | [Go Toolchains](https://go.dev/doc/toolchain) | Auto mode may select/download a newer toolchain; `GOTOOLCHAIN=local` keeps the bundled/selected toolchain and fails compatibility instead. |
| Go build flags | [Go command](https://pkg.go.dev/cmd/go) | `-trimpath`, `-buildvcs`, `-buildmode`, `-mod`, and `-o` are manager-owned fixed flags. |
| Go generators | [Go generate](https://pkg.go.dev/cmd/go#hdr-Generate_Go_files_by_processing_source) | Generation is not automatically run by build/test; the manager never invokes it. |
| cgo subprocess surface | [cgo command](https://pkg.go.dev/cmd/cgo) | cgo can select/invoke C/C++ compiler and `pkg-config` behavior; disabling it removes that driver surface. |
| Rust build-time execution | [Cargo build scripts](https://doc.rust-lang.org/cargo/reference/build-scripts.html), [Rust procedural macros](https://doc.rust-lang.org/reference/procedural-macros.html) | Cargo runs compiled build scripts; proc macros execute during compilation. |
| Zig build/comptime behavior | [Zig build system](https://ziglang.org/learn/build-system/), [Zig reference](https://ziglang.org/documentation/master/#comptime) | `build.zig` controls build graph/tools; language code can be evaluated at compile time. |
| Swift package execution | [SwiftPM plugins](https://docs.swift.org/swiftpm/documentation/packagemanagerdocs/plugins/) | Package manifests run as Swift; build command plugins execute processes. |
| Make/CMake commands | [GNU Make](https://www.gnu.org/software/make/manual/html_node/Execution.html), [CMake process](https://cmake.org/cmake/help/latest/command/execute_process.html), [CMake custom command](https://cmake.org/cmake/help/latest/command/add_custom_command.html) | Recipes/configure/custom commands are package-selected execution points. |
| JVM build/compile plugins | [javac](https://docs.oracle.com/en/java/javase/23/docs/specs/man/javac.html), [Maven](https://maven.apache.org/guides/introduction/introduction-to-plugins.html), [Gradle](https://docs.gradle.org/current/userguide/implementing_custom_tasks.html), [Kotlin](https://kotlinlang.org/docs/compiler-plugins-overview.html) | Annotation processors, plugins, and tasks execute code during compilation/build. |
| .NET build tasks/generators | [MSBuild tasks](https://learn.microsoft.com/en-us/visualstudio/msbuild/msbuild-tasks), [inline tasks](https://learn.microsoft.com/en-us/visualstudio/msbuild/msbuild-inline-tasks), [`Exec`](https://learn.microsoft.com/en-us/visualstudio/msbuild/exec-task), [Roslyn compiler platform](https://learn.microsoft.com/en-us/dotnet/csharp/roslyn-sdk/) | Project-controlled tasks/inline code/commands and source generators execute at build time. |
| npm lifecycle behavior | [npm scripts](https://docs.npmjs.com/cli/using-npm/scripts/) | Install lifecycle scripts and `node-gyp` behavior can execute package code. |
| Deno compile surface | [Deno compile](https://docs.deno.com/runtime/reference/cli/compile/), [Deno packages](https://docs.deno.com/runtime/packages/) | Compile supports network/package/config/script-affecting surfaces that need a separately pinned contract. |
| Python build backends | [pip build-system interface](https://pip.pypa.io/en/latest/reference/build-system/), [PyPA flow](https://packaging.python.org/en/latest/flow/) | Frontends invoke package-selected backend hooks and may install dynamic build dependencies. |
