# Compile-only build drivers: protocol research and `go-v1` recommendation

**Task:** `TASK-260720-poa3ze`  
**Date:** 2026-07-20  
**Status:** Research outcome for review  
**Revision:** Reworked after review cycles covering external linking/inputs, telemetry isolation, interoperable digests, compiler-visible build-source identity, agent-context separation, cross-project transaction isolation, Go dynamic-import directives, ecosystem/artifact coverage, cache provenance, conformance-claim versioning, and cache-layout portability.  
**Scope:** Read-only inspection of the three requested `origin/main` revisions; no product or specification source was modified.

## Key takeaways

1. **Add one closed command form plus explicit source isolation in manifest schema 6:** `{"type":"build","driver":"go-v1","source_dir":"build/cmd/tool"}` and top-level `"build_roots":["build"]`. Build roots are excluded from agent context and are not copied as runtime. The package can select only a declared build root/source directory and the closed driver identifier. It cannot supply a program, arguments, environment, output path, tags, linker flags, build script, or hook.
2. **Ship only `go-v1` in protocol v1.** The manager invokes an operator-trusted Go 1.23+ toolchain directly, without a shell, using fixed `go telemetry off`, `go version`, `go env`, `go list`, and `go build` vectors. The build is vendor-only and networkless; forces the internal linker with no libgcc fallback; disables cgo, PGO, workspaces, and toolchain switching; rejects package-controlled assembly and host objects throughout the active dependency graph; and conservatively rejects `//go:cgo_import_dynamic` in active non-standard Go files. It never invokes the built executable.
3. **Treat build outputs as a distinct immutable cache with an explicit trust boundary.** The current runtime identity of `skill + commit` is insufficient. A build key must cover a domain-separated, length-framed digest of every regular file in the raw snapshot—including a package-provided root `.csk-install.json`—plus command/source selection, driver revision, native target, fixed policy, and exact toolchain fingerprint. Reuse is allowed only from manager-created, owner-protected machine state whose ownership/ACL and link safety are revalidated; otherwise the manager rebuilds. The canonical receipt and marker hashes detect corruption/currentness but are not signatures or independent proof of provenance.
4. **Build before any installation mutation; serialize shared commit state.** Resolve, validate, gate, and build misses in operation-private staging outside the commit lock. Then acquire one manager-home mutation lock, recover/revalidate shared state, publish immutable cache entries, and commit all project/global/hybrid/adapter/consumer changes with one rollback journal. Keep the lock through rollback and GC. Record the consumer last so concurrent projects cannot lose updates or restore stale backups.
5. **Dry-run must not compile.** It may inspect the snapshot, the trusted toolchain identity, and existing receipts read-only, but it must not run `go list` or `go build`, create Go caches, publish artifacts, or mutate registry/audit/runtime/project state.
6. **Do not generalize this into “run a build tool.”** Cargo build scripts and procedural macros, SwiftPM manifests/plugins, Make/CMake/Meson commands, Maven/Gradle plugins, MSBuild tasks, npm lifecycle scripts, JavaScript bundler configs/loaders/plugins, Python build backends, and similar mechanisms can execute package-selected host code during a build. Those violate the invariant by construction.
7. **The current managers expose pre-existing lifecycle issues that the build change must not copy forward.** Both record a consumer before materialization and both trust an existing runtime cache directory without verifying its required paths. The Python dry-run also appears able to mutate audit-registry cache/state before returning. These are implementation findings, not reasons to weaken the proposed contract.
8. **Publish the feature as protocol `1.0.0-rc.4` with conformance claim schema 2.** Freeze claim schema 1 as historical rc.3 evidence; current rc.4 writers emit `conformance-claim-v2`, and current readers distinguish the two versions rather than rewriting the meaning of the v1 schema URL.

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

- the repository snapshot, including source, `go.mod`, `go.sum`, `vendor/`, filenames, build constraints, compiler directives, embedded files, assembly, host objects, and file sizes;
- both manifest spellings and all manifest fields;
- refs, registry data, cached source/runtime entries, and stale install markers until validated;
- build-receipt and artifact bytes as untrusted parser/hash inputs; their eligibility for reuse additionally depends on the manager-protected-state boundary below;
- any executable artifact produced from the package;
- output text emitted while the compiler parses the package.

The trusted computing base is deliberately small:

- the manager binary and protocol implementation;
- the operator-selected operating system and native Go toolchain, resolved independently of the package;
- manager-owned policy constants, staging directories, cache locks, transaction journal, and manager-created private build-cache root;
- operating-system ownership, permission/ACL, and link-safety enforcement for that private root;
- the cryptographic hash and canonical JSON implementations.

The package must not influence toolchain selection through its snapshot, manifest, current directory, `PATH`, Go environment files, workspace files, or module `toolchain` directive.

The v1 cache trust model is deliberately local and explicit: an adversary may supply arbitrary package/cache candidate bytes but cannot write manager-protected state as the manager's operating-system principal or as a trusted administrator. Arbitrary same-principal code execution, administrator/root compromise, kernel compromise, and hostile storage below the operating system are outside this install-time invariant; any of them could replace every local marker and manager binary as well as a receipt. This is an honest boundary for the two current user-space managers. A deployment that cannot enforce or accept it MUST disable persistent cache reuse and rebuild from the revalidated snapshot; authenticated cross-principal provenance is a possible future protocol feature, not something a plain SHA-256 receipt supplies.

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
package-supplied -tags, -ldflags, -gcflags, -asmflags, -toolexec, -overlay, -pgo
package-supplied environment variables or output paths
execution of the output binary
```

## 3. Manifest contract

### 3.1 Recommended schema shape

Add schema version 6 to both canonical manifest names. Preserve the existing `command` definition for schemas 1–5; changing it in place would accidentally allow build commands in older manifests. Add these definitions to `common.schema.json` and let only the v6 schemas reference `commandV6`:

```json
{
  "buildCommandV6": {
    "type": "object",
    "required": ["type", "driver", "source_dir"],
    "properties": {
      "type": {"const": "build"},
      "driver": {"const": "go-v1"},
      "source_dir": {"$ref": "#/$defs/portablePath"}
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
    "build_roots": {"$ref": "common.schema.json#/$defs/pathSet"},
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

`build_roots` is new in schema 6 and is the conservative, compiler-independent context boundary. Each entry names a raw-source directory used only for manager-owned compilation. It is excluded from agent-facing context before locale rendering, never copied into the commit-keyed runtime store, and never installed as generated output. Schema versions 1–5 continue to reject this unknown field. V1 intentionally does not allow `.` as either a build root or `source_dir`; a root-module package must move its Go module below a dedicated directory such as `build/`.

### 3.2 Valid manifest example

This example deliberately mixes a compiled command with an existing script command; command collision and activation rules apply uniformly to both.

```json
{
  "schema_version": 6,
  "runtime_roots": ["scripts"],
  "build_roots": ["build"],
  "commands": {
    "golden-tool": {
      "type": "build",
      "driver": "go-v1",
      "source_dir": "build/cmd/golden-tool"
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

1. A `build` command and `build_roots` are legal only in manifest schema 6. Legacy `agents/runtime.json` commands cannot express either surface.
2. `driver` must equal the closed identifier `go-v1`. Unsupported drivers fail; there is no fallback to a system build command.
3. Every `build_roots` entry must be a portable relative path naming a real, link-free directory in the immutable snapshot. Build roots are unique and pairwise disjoint, and no build root may equal, contain, or be contained by a `runtime_root`. `.` is not a build root.
4. `source_dir` is a portable relative path naming a real, link-free directory below exactly one declared build root. It may equal that build root. `.` and paths outside all build roots fail.
5. The unique containing build root is the command's `build_root`. It must contain `go.mod` directly, and that file must be the nearest ancestor `go.mod` of `source_dir`; an intervening nested module fails. `GOWORK=off` means a `go.work` file never expands this boundary.
6. Every declared build root must be referenced by at least one build command. Before locale rendering or context copying, the context selector excludes the complete subtree of every build root in addition to every runtime root. This static rule applies identically on cache hits and dry-runs, which do not run `go list`; build roots are also omitted from runtime-root copying. Generated output lives only in manager-owned staging/cache directories and is never prompt-visible.
7. The selected directory must resolve through fixed `go list` semantics to exactly one buildable package named `main`. Package patterns, multiple packages, and library-only packages fail.
8. A command object cannot combine script, system, and build properties. Unknown fields fail before semantics.
9. The manager derives the output basename from the manifest command key. Unix output is `bin/<command>`; Windows output is `bin/<command>.exe`. The package cannot override it.
10. Build commands participate in the existing activation, dependency-command selection, portable-name collision, and shim collision rules exactly like script commands.
11. Active build command names are processed in Unicode code-point/bytewise lexical order within each provider-first closure node. The same normalized ordering is used in receipts and markers.
12. The `curator-build-source-v1` identity from section 5.2—not the installed-tree `content_sha256` and not only `source_dir`—is part of build identity. It covers every regular file in the fully validated raw snapshot, including root `.csk-install.json`, before cache lookup. This intentionally favors safe invalidation over maximum cache hits.
13. A manager that cannot enforce the fixed environment, offline module mode, native target, `CGO_ENABLED=0`, source/context boundary, and transaction rules must reject the build command rather than approximate it.
14. The selected toolchain must be Go 1.23 or newer, must pass the closed toolchain-identity rules in section 4.2, and must be in a Go release family the manager has tested against the `go-v1` conformance vectors. An unknown future release is not accepted merely because its version compares greater than 1.23.
15. The fixed dependency listing must cover the root and every active dependency. Every non-standard package directory, module file, active Go file, native-source field, and embedded input must resolve below the command's build root; standard packages and all their listed inputs must resolve below the fingerprinted `GOROOT`. Every non-standard package must have empty `CgoFiles`, `CFiles`, `CXXFiles`, `MFiles`, `HFiles`, `FFiles`, `SFiles`, `SwigFiles`, `SwigCXXFiles`, and `SysoFiles`; every package, including the standard library, must have empty `SysoFiles`. Any violation fails before `go build`.
16. For each non-standard package, join each active `GoFiles` entry to its validated package directory, require a regular file below the build root, and reject the file if its exact bytes contain ASCII `//go:cgo_import_dynamic`. This deliberately conservative byte scan may reject a harmless string/comment containing the token; it cannot miss an active directive. Standard-library files are trusted as part of the fingerprinted toolchain and are not scanned.
17. Snapshot validation and `curator-build-source-v1` digesting occur before any build-cache lookup or Go command. The manager must build from the exact validated snapshot instance and fail if its tree or bytes change before the last compiler child exits.
18. A cache hit is usable only when its canonical receipt contains the exact current `build_root`, `source_dir`, build-source identity, and directive policy `reject-nonstandard-cgo-import-dynamic-v1`. A receipt from an earlier policy is not upgraded in place. A dry-run performs rules 1–6 and 8–14 statically, validates an exact receipt if present, and reports that rules 7, 15, and 16 would run for a miss.

No package or project field is added for arbitrary build arguments. No build-policy override is added to manager or system configuration in v1.

### 3.5 Context-selection vectors

The positive vector is compiler-free. Its build root is deliberately nested below the otherwise agent-visible `assets/` root, so it proves the new subtree exclusion rather than merely relying on the existing top-level whitelist. All possible main-module, vendored, and embedded inputs live below that build root, while an unrelated asset and ordinary skill instructions remain agent-facing:

```json
{
  "name": "build-root-excluded-from-agent-context",
  "manifest": {
    "schema_version": 6,
    "build_roots": ["assets/build-tool"],
    "commands": {
      "golden-tool": {
        "type": "build",
        "driver": "go-v1",
        "source_dir": "assets/build-tool/cmd/golden-tool"
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
  },
  "snapshot_files": [
    "SKILL.md",
    "assets/prompt.md",
    "assets/build-tool/go.mod",
    "assets/build-tool/cmd/golden-tool/main.go",
    "assets/build-tool/internal/render/render.go",
    "assets/build-tool/internal/render/template.txt"
  ],
  "expected_context_files": ["SKILL.md", "assets/prompt.md"],
  "expected_excluded_files": [
    "assets/build-tool/go.mod",
    "assets/build-tool/cmd/golden-tool/main.go",
    "assets/build-tool/internal/render/render.go",
    "assets/build-tool/internal/render/template.txt"
  ],
  "cache_hit_go_commands": [],
  "dry_run_go_commands": []
}
```

The negative vector fails before context selection, cache lookup, or any Go command. Root modules fail for the same reason because `.` is not a portable build root in v1.

```json
{
  "name": "build-source-outside-declared-build-root",
  "manifest": {
    "schema_version": 6,
    "build_roots": ["build"],
    "commands": {
      "bad-tool": {
        "type": "build",
        "driver": "go-v1",
        "source_dir": "assets/tool"
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
  },
  "expected_error": "build_source_outside_build_root",
  "cache_lookup": false,
  "go_commands": []
}
```

## 4. Fixed `go-v1` driver

### 4.1 Supported shape

`go-v1` builds one native executable from one main package:

- trusted Go 1.23 or newer, limited to release families the manager has tested for this driver;
- host target only; no cross-compilation in v1;
- Go module mode only; the module root is exactly one declared, context-excluded build root below the snapshot root;
- standard-library-only dependencies or dependencies already present in the main module’s top-level `vendor/` tree;
- cgo, PGO, package-controlled assembly, host objects, and non-standard `go:cgo_import_dynamic` directives disabled;
- no workspaces, generators, tests, plugins, overlays, external compiler, or external linker;
- one output selected by the manager.

Missing/inconsistent vendored dependencies, a newer required Go version, a package that needs cgo or non-standard assembly, a target that cannot link internally, or a non-main package are deterministic install errors.

### 4.2 Toolchain selection and identity

The manager resolves the Go executable before entering any package-controlled directory. The candidate must be bundled with the manager or selected by trusted operator configuration/environment; it must not come from the repository, `runtime_roots`, project `.agents/bin`, the user `PATH`, or the package manifest. Resolve an outside launcher symlink and require the result to be the regular executable `<resolved-goroot>/bin/go` (`go.exe` on Windows). The candidate cannot be a wrapper or an executable outside the tree being fingerprinted. With `GOROOT` initially absent from the clean environment, the later `go env GOROOT` probe must canonicalize to the same derived root.

`toolchain.content_sha256` uses the byte-exact algorithm identifier `curator-go-toolchain-v1`:

1. Walk `GOROOT` without following links. The root itself is not a record. Require each relative component to contain valid Unicode scalar values; join components with `/` without case folding or Unicode normalization; encode UTF-8; reject duplicate encoded paths and special files.
2. Symlinks must be non-dangling, relative, and resolve within `GOROOT`. Their referents therefore have their own tree records. Absolute or escaping links reject the toolchain.
3. Sort all relative path bytes in unsigned bytewise order. Initialize SHA-256 with the exact ASCII bytes `curator-go-toolchain-v1` followed by `0x00`.
4. For each entry append `kind || uint64be(path_length) || path_utf8 || uint64be(payload_length) || payload`, where `kind` is ASCII `D`, `F`, or `L`; directory payload is empty; regular-file payload is the exact file bytes; and link payload is the exact UTF-8 string returned by `readlink`, without separator or Unicode normalization.
5. Capture `go version` stdout after telemetry isolation. It must be at most 4096 bytes, end in exactly one LF (optionally preceded by CR), and otherwise contain no CR, LF, or NUL. Remove that LF and optional CR and require the result to be non-empty valid UTF-8. Append one final record using kind `V`, an empty path, and those normalized version bytes as payload.
6. Prefix the lowercase digest with `sha256:`. File/directory permission bits, ownership, timestamps, ACLs, and extended attributes are deliberately not hashed; separately require `bin/go` and every invoked `pkg/tool/<host>/` child to be regular and executable at use time. Hard links are hashed as independent regular-file path/content records.

This makes the digest relocatable while covering the selected `go` binary, compiler/linker/assembler tools, standard-library sources, internal symlink topology, and normalized version output. The manager must fail if the tree changes between fingerprinting and the last child exit. A manager-owned digest memo may be read during dry-run, but uncertain metadata or toolchain changes force a full recomputation; dry-run cannot create or refresh the memo.

This cross-language conformance vector covers directory/file ordering, an internal symlink, exact file bytes, and terminal-newline normalization:

```json
{
  "algorithm": "curator-go-toolchain-v1",
  "entries": [
    {"path": "pkg/tool-link", "type": "symlink", "target": "../bin/go"},
    {"path": "bin/go", "type": "file", "content_base64": "R08="},
    {"path": "pkg", "type": "directory"},
    {"path": "bin", "type": "directory"}
  ],
  "go_version_stdout_base64": "Z28gdmVyc2lvbiBnbzEuMjUuNSBkYXJ3aW4vYXJtNjQK",
  "normalized_go_version": "go version go1.25.5 darwin/arm64",
  "content_sha256": "sha256:baf7c5f3b9c3f1fae3da4c356381bf74442aa7f8f0b6fb2304c9c10833d6032e"
}
```

Input entry order is intentionally non-canonical; implementations must sort it. An equivalent CRLF-terminated version string has the same normalized result; missing/multiple terminal newlines are invalid. Changing a link target or file byte changes the digest, while mode/timestamp-only changes do not. The cache/receipt records the algorithm, normalized `go_version`, fixed `go_relpath: "bin/go"`, and digest. `GOTOOLCHAIN=local` makes a package `toolchain`/`go` directive a compatibility check, never permission to download or switch toolchains ([Go toolchain selection](https://go.dev/doc/toolchain)).

### 4.3 Environment construction

Start from an empty environment. Add only OS variables indispensable for process creation (for example `SYSTEMROOT` on Windows), a manager-owned empty `PATH` directory, locale normalization, operation-private directories, resolved `GOROOT`, the explicit native target, and the following Go policy:

```json
{
  "GO111MODULE": "on",
  "GOENV": "off",
  "GOFLAGS": "",
  "GOPATH": "<operation-private>/gopath",
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
  "CGO_ENABLED": "0",
  "GO_EXTLINK_ENABLED": "0",
  "GOEXPERIMENT": "",
  "GOROOT": "<resolved-trusted-goroot>",
  "GOOS": "<native-goos>",
  "GOARCH": "<native-goarch>",
  "HOME": "<operation-private>/home",
  "XDG_CONFIG_HOME": "<operation-private>/config",
  "PATH": "<operation-private>/empty-path",
  "TMPDIR": "<operation-private>/tmp",
  "LC_ALL": "C",
  "LANG": "C"
}
```

Set exactly the target tuning variable applicable to `GOARCH` (`GO386`, `GOAMD64`, `GOARM`, `GOARM64`, `GOMIPS`, `GOMIPS64`, `GOPPC64`, `GORISCV64`, or `GOWASM`) to the trusted probe result; omit the rest. On Windows, set `APPDATA`, `LOCALAPPDATA`, `USERPROFILE`, `TEMP`, and `TMP` below the operation-private root. On Darwin, the private `HOME` redirects `os.UserConfigDir()` to `HOME/Library/Application Support`; on other Unix hosts, absolute `XDG_CONFIG_HOME` does so ([`os.UserConfigDir`](https://pkg.go.dev/os#UserConfigDir)). Do not inherit `CC`, `CXX`, `PKG_CONFIG`, `AR`, `GCCGO`, `GOAUTH`, user `PATH`, `GOROOT`, or any other Go variable. `GO_EXTLINK_ENABLED` is the manager-owned value `0`, not inherited.

`GOTELEMETRY` and `GOTELEMETRYDIR` are read-only Go environment values, so setting `GOTELEMETRY=off` is not a control. `go-v1` requires Go 1.23+, initializes `off` mode with the fixed command below inside the private user-config root, and verifies both reported values. Go documents that `off` disables collection and upload, while the default `local` mode writes counters to `os.UserConfigDir()/go/telemetry` ([Go telemetry](https://go.dev/doc/telemetry); [Go 1.23 release notes](https://go.dev/doc/go1.23#telemetry)). The mode file and any initialization counters are temporary operation state and are deleted with the private root.

The source snapshot is mounted/readable but not writable by the child. Only operation-private cache/temp/output directories are writable. Standard input is closed; stdout/stderr are captured with size limits. Apply a deadline, memory/disk/process limits, and OS network denial. Environment denial is normative even when a stronger OS sandbox is available.

### 4.4 Exact process construction

Invoke the resolved Go binary directly with argument vectors—never a joined command string. With a manager-owned empty directory as CWD and the bootstrap form of the clean environment (private user/config/temp roots, `GOENV=off`, `GOTOOLCHAIN=local`, and no inherited `GOROOT` or target), run exactly once per operation:

```json
["/absolute/trusted/goroot/bin/go", "telemetry", "off"]
```

```json
["/absolute/trusted/goroot/bin/go", "version"]
```

```json
["/absolute/trusted/goroot/bin/go", "env", "-json", "GOROOT", "GOHOSTOS", "GOHOSTARCH", "GOOS", "GOARCH", "GO386", "GOAMD64", "GOARM", "GOARM64", "GOMIPS", "GOMIPS64", "GOPPC64", "GORISCV64", "GOWASM", "GOTELEMETRY", "GOTELEMETRYDIR"]
```

Require `GOTELEMETRY == "off"`, require `GOTELEMETRYDIR` below the operation-private root at the platform’s expected `os.UserConfigDir()` location, validate `GOROOT` and the minimum/allowlisted release family, freeze the native target/tuning tuple, and then add explicit `GOROOT`, target, and tuning entries to the environment.

For each command set CWD to canonical `source_dir` and run the full dependency preflight:

```json
["/absolute/trusted/goroot/bin/go", "list", "-mod=vendor", "-deps", "-json", "-buildvcs=false", "-compiler=gc", "-pgo=off", "."]
```

Parse the complete JSON stream and require exactly one non-`DepOnly` result with `Name == "main"`; no result may be incomplete or carry `Error`/`DepsErrors`. A result with `Standard && Goroot` is trusted only when its directory and every listed input remain under fingerprinted `GOROOT`; any other result and its module, source, and embedded-file paths must remain under the command's declared build root, not merely somewhere in the snapshot. Reject `SysoFiles` for every result. For every non-standard result also reject any active cgo/C/C++/Objective-C/Fortran/SWIG or assembly fields named in semantic rule 15. In particular, non-standard `SFiles` are rejected rather than parsed: the Go assembler supports `#include`, whose implementation can open a supplied absolute name directly ([assembler source](https://github.com/golang/go/blob/master/src/cmd/asm/internal/lex/input.go#L401-L424)).

`go list` exposes active non-cgo files in `GoFiles`, but its native-file fields do not reveal compiler directives embedded in ordinary Go source. For every non-standard result, the manager therefore reads each active `GoFiles` file from the frozen snapshot and rejects the exact ASCII token `//go:cgo_import_dynamic` before `go build`. The Go compiler explicitly permits this directive outside cgo-generated files, while the internal linker accepts `_ _ "library"` as a request to force a dynamic library into the result ([Go compiler directive handling](https://github.com/golang/go/blob/go1.25.5/src/cmd/compile/internal/noder/noder.go#L311-L332); [Go linker handling](https://github.com/golang/go/blob/go1.25.5/src/cmd/link/internal/ld/go.go#L105-L136)). Standard-library uses remain trusted toolchain content.

The required rejection vector has empty cgo/native fields, proving those fields alone are insufficient:

```json
{
  "name": "reject-active-nonstandard-cgo-import-dynamic",
  "package": {
    "Standard": false,
    "Dir": "build/cmd/dynamic-probe",
    "GoFiles": ["main.go"],
    "CgoFiles": [],
    "CFiles": [],
    "CXXFiles": [],
    "SFiles": [],
    "SysoFiles": []
  },
  "source_token": "//go:cgo_import_dynamic",
  "expected_error": "go_forbidden_compiler_directive",
  "go_build_started": false,
  "cache_published": false
}
```

Only after the graph passes, build with this exact vector:

```json
["/absolute/trusted/goroot/bin/go", "build", "-mod=vendor", "-trimpath", "-buildvcs=false", "-buildmode=exe", "-compiler=gc", "-pgo=off", "-ldflags=-linkmode=internal -libgcc=none", "-o", "<operation-staging>/bin/golden-tool", "."]
```

On Windows the output argument is `<operation-staging>/bin/golden-tool.exe`. All arguments except the fingerprinted tool path and manager-derived staging path are protocol constants. `.` is deliberately the only package operand. `-pgo=off` removes the default package-controlled `default.pgo` input; the Go build documentation otherwise defines PGO default as `auto` ([Go build flags](https://go.dev/cmd/go/#hdr-Compile_packages_and_dependencies)).

The single `-ldflags=...` argv element supplies two linker arguments. `-linkmode=internal` prevents `GO_EXTLINK_ENABLED` or a custom toolchain’s compiled default from selecting external mode; if the target/input requires external linking, the linker fails instead of falling back ([link-mode source](https://go.dev/src/cmd/link/internal/ld/config.go#L182-L227)). `-libgcc=none` prevents the internal linker’s documented fallback of running the external compiler to locate libgcc ([linker flags](https://go.dev/cmd/link/)). The environment still fixes `GO_EXTLINK_ENABLED=0` as defense in depth. A trace may contain an inert default `-extld=cc` argument generated by trusted `go`; with forced internal mode, no host compiler/linker may start.

The manager directly launches only the fingerprinted `go`. Its permitted transitive executable children are regular executable files below the fingerprinted `GOROOT/pkg/tool/<GOHOSTOS>_<GOHOSTARCH>/`; toolchain switching, VCS commands, generators, cgo tools, host compilers, and every other executable are outside the process graph. An observed child outside that directory is a conformance failure. Integration vectors poison `cc`, `clang`, `gcc`, `ld`, and `pkg-config` and trace child starts; executor tests assert the five manager-owned argv forms and that no output artifact is ever invoked.

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

cgo may invoke C/C++ compilers and `pkg-config`, and package source can carry `#cgo` directives ([cgo documentation](https://pkg.go.dev/cmd/cgo)). `go-v1` therefore fixes `CGO_ENABLED=0` and does not inherit compiler variables. Source files requiring cgo are excluded by Go’s normal build selection; if the remaining package cannot build, installation fails. `CGO_ENABLED=0` is not sufficient by itself: ordinary active Go source can contain `//go:cgo_import_dynamic`, so the separate semantic-rule-16 byte scan is mandatory.

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

Two source identities have deliberately different jobs:

- Existing marker `content_sha256` remains the protocol section 8 hash of installed regular files **excluding** root `.csk-install.json`. It answers whether the manager-owned installed tree is current after that marker is written.
- New `build_source` is the build-cache identity of the immutable **raw snapshot** before installation writes anything. It includes every regular file, including a package-provided root `.csk-install.json`, whether or not the file is under `source_dir`.

`build_source.content_sha256` uses byte-exact algorithm identifier `curator-build-source-v1`:

1. Fully validate the raw snapshot before hashing. Do not follow links. Every descendant must be a directory or regular file; reject symbolic links, special files, invalid/non-portable protocol paths, duplicate encoded paths, and platform path collisions.
2. Collect every regular file with no exclusion list. Root `.csk-install.json` is an ordinary included record. Convert its relative protocol path to `/` separators, without case folding or Unicode normalization, and encode valid Unicode scalar values as UTF-8.
3. Sort path UTF-8 bytes in unsigned bytewise order. Initialize SHA-256 with exact ASCII `curator-build-source-v1` followed by `0x00`.
4. For each file append `F || uint64be(path_length) || path_utf8 || uint64be(content_length) || content_bytes`. Lengths are unsigned 64-bit big-endian byte counts. Empty files have zero content length; an empty snapshot hashes only the domain prefix.
5. Encode lowercase hexadecimal and prefix `sha256:`. File modes, ownership, timestamps, ACLs, and extended attributes are not inputs. The validated snapshot instance must remain unchanged through the last build child exit.

Do not implement this as the legacy content hash with an empty exclusion set. Besides intentionally excluding the marker by default, that hash uses NUL-delimited paths and unconstrained file bytes, so its record stream is not self-delimiting when content contains NUL. The new domain prefix and length frames give build caches an unambiguous cross-language preimage while leaving installed-tree compatibility unchanged ([current protocol content hash](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/protocol/core.md#L282-L295)).

Add this strict shared definition and reference it from both the receipt input and marker v2:

```json
{
  "buildSourceIdentity": {
    "type": "object",
    "required": ["algorithm", "content_sha256"],
    "properties": {
      "algorithm": {"const": "curator-build-source-v1"},
      "content_sha256": {"$ref": "#/$defs/sha256"}
    },
    "additionalProperties": false
  }
}
```

For each active build command, construct this logical input object. This is an illustrative Darwin/arm64 instance. The build-source, toolchain, and artifact hashes use repeated-byte fixtures, while the cache and receipt hashes below are the exact derived values for those fixtures.

```json
{
  "schema_version": 1,
  "driver": "go-v1",
  "build_source": {
    "algorithm": "curator-build-source-v1",
    "content_sha256": "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
  },
  "build_root": "build",
  "command": "golden-tool",
  "source_dir": "build/cmd/golden-tool",
  "target": {
    "goos": "darwin",
    "goarch": "arm64",
    "tuning": {
      "GOARM64": "v8.0"
    }
  },
  "toolchain": {
    "algorithm": "curator-go-toolchain-v1",
    "go_relpath": "bin/go",
    "go_version": "go version go1.26.1 darwin/arm64",
    "content_sha256": "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
  },
  "policy": {
    "module_mode": "vendor",
    "network": "none",
    "workspace": false,
    "cgo": false,
    "compiler_directives": "reject-nonstandard-cgo-import-dynamic-v1",
    "target_mode": "native",
    "link_mode": "internal",
    "libgcc": "none",
    "package_assembly": false,
    "host_objects": false,
    "telemetry": "off-private"
  }
}
```

Canonicalize the input object with the protocol’s canonical JSON rules and set:

```text
cache_key = "sha256:" + lowercase_hex(SHA-256(canonical_utf8(input)))
```

Here `canonical_utf8` is exactly CCJ-1: valid UTF-8, Unicode-scalar key order, no insignificant whitespace or BOM, and no trailing LF. For the object above the exact result is `sha256:3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48`.

The driver identifier is versioned (`go-v1`), so any change to command arguments, environment, sandbox-relevant semantics, receipt interpretation, or output rules requires a new driver identifier or an explicitly versioned cache-key policy revision.

Retain the prior marker-embed case as a **negative migration/regression fixture**, not as a valid v1 command. It uses a root module, so the new `build_roots` rule rejects it before cache-key construction. The two raw snapshots share `go.mod` and this source and differ only in root `.csk-install.json`:

```go
package main

import (
	_ "embed"
	"os"
)

//go:embed .csk-install.json
var marker []byte

func main() { _, _ = os.Stdout.Write(marker) }
```

Go reads explicitly matched files into `string`/`[]byte` variables at compile time, so a dotfile named explicitly by the pattern is compiler-visible ([Go `embed` directives](https://pkg.go.dev/embed#hdr-Directives)). The all-file identity remains required as defense in depth and for any future driver that admits a broader source layout. The exact negative vector is:

```json
{
  "name": "reject-root-module-marker-embed",
  "candidate_command": {
    "type": "build",
    "driver": "go-v1",
    "source_dir": "."
  },
  "directive": "//go:embed .csk-install.json",
  "legacy_content_sha256": "sha256:829a040a1455fdf96e2731aa5c089e7e42dbcec2a51b1db3222a610f0ffb5b35",
  "variants": [
    {
      "marker_content_base64": "eyJ2YXJpYW50IjoiQSJ9Cg==",
      "build_source": {
        "algorithm": "curator-build-source-v1",
        "content_sha256": "sha256:0017492cfbcd822237a7e72239d45b59f0923b54f2ac2e0a59ecd9202cc48ad0"
      }
    },
    {
      "marker_content_base64": "eyJ2YXJpYW50IjoiQiJ9Cg==",
      "build_source": {
        "algorithm": "curator-build-source-v1",
        "content_sha256": "sha256:60fe9b764163b7d6bc38bc0cac63675398eb7167ad692fd189e221c3fc096266"
      }
    }
  ],
  "expected": {
    "legacy_content_hashes_equal": true,
    "build_source_hashes_equal": false,
    "semantic_error": "build_source_outside_build_root",
    "cache_keys_created": false,
    "go_commands": []
  }
}
```

Source validation and this digest happen before cache lookup. A cache hit therefore cannot bypass inspection of raw snapshot bytes. The task-scoped Go 1.25.5 Darwin/arm64 research smoke built the otherwise-invalid root fixtures without executing them and produced different artifact hashes (`69f8a5d047b4345d8398750d0d59c0bbd6f60101d5c40689bb31625cc10583fd` and `7ddd293aa1aa441ef83e57a26ad7b886ad9156cac4cde4a317214b870ee5ff23`), confirming why both the location restriction and all-file identity are necessary.

### 5.3 Logical entry, protected state, and non-normative layout

The portable contract standardizes the logical cache key, exact canonical receipt bytes, manager-derived artifact-relative path, artifact bytes/hash/size, and validation results. It does **not** standardize the manager-home path, physical cache-root name, driver subdirectory, receipt filename, quarantine path, lock filename, or storage backend. That follows the existing core rule that machine-home directories and cache names are implementation-specific and that managers do not share machine-local state by default ([core compatibility identifiers, lines 23–34](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/protocol/core.md#L23-L34)).

This is therefore an illustrative implementation layout only, not a portable pathname:

```text
<implementation-manager-home>/<implementation-cache-root>/go-v1/<64-hex-key>/
  <implementation-receipt-name>
  bin/<command>[.exe]
```

The two managers may retain different home/cache/metadata names. Conformance fixtures address the entry by logical key and artifact-relative path, never by a literal `build-cache` or `.csk-build.json` pathname.

Persistent reuse is valid only inside **manager-protected state**, which is part of the v1 TCB. Both implementations must enforce these semantics, using platform-specific mechanisms:

1. Resolve the cache root independently of the package. Create it exclusively as private manager state; never import a package cache or silently adopt a pre-existing directory whose protection cannot be established.
2. Verify the manager owns the protected boundary and every cache component at or below it. On POSIX, the baseline is the manager effective UID, no group/other write bit, private directories, and owner-only receipt/artifact mutation. On Windows, use a DACL that grants mutation only to the manager principal and trusted operating-system administrators. Root/administrator is part of the stated TCB.
3. Reject symbolic links, reparse-point escapes, special files, and multiply linked receipt/artifact files. Resolve children relative to a held root/directory handle with no-follow semantics and re-check the opened object; Go's rooted file API is one implementation aid because it rejects names that resolve outside the root ([Go `os.OpenInRoot`](https://pkg.go.dev/os#OpenInRoot)), while Windows DACLs provide the platform access-control boundary ([Microsoft access control](https://learn.microsoft.com/en-us/windows/win32/secauthz/access-control)).
4. Revalidate the boundary on every lookup and again under the manager-home mutation lock before publication/commit. If ownership, ACL/mode, containment, file type, or link safety cannot be proved, do not parse the candidate as a hit. A real operation rebuilds into newly established protected state; dry-run reports `would-rebuild-untrusted-cache`; status is non-current; repair rebuilds. An implementation without reliable platform checks MUST disable persistent reuse rather than fail open.
5. After atomic publication, make the entry non-writable in the ordinary case while retaining locked manager GC capability (for example, owner-only directories, read-only receipt bytes, and owner-executable artifact on POSIX). No package field can relax the protection.

Entries are immutable after atomic publication. Build into an operation-private directory outside the live cache. Publication, corruption quarantine/replacement, marker reference changes, and GC all occur while holding the manager-home mutation lock from section 6. If another project publishes the key before this transaction acquires that lock, validate the protected-state boundary and winner; discard an identical staged loser, while differing bytes for the same key are a determinism/corruption error. Never merge files into an existing entry. Per-key build locks may reduce duplicate compilation, but they are only an optimization, must be released before the manager-home lock, and cannot substitute for transaction isolation.

### 5.4 Receipt example

The receipt is canonical JSON with no timestamps or absolute paths:

```json
{
  "schema_version": 1,
  "cache_key": "sha256:3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48",
  "input": {
    "schema_version": 1,
    "driver": "go-v1",
    "build_source": {
      "algorithm": "curator-build-source-v1",
      "content_sha256": "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
    },
    "build_root": "build",
    "command": "golden-tool",
    "source_dir": "build/cmd/golden-tool",
    "target": {
      "goos": "darwin",
      "goarch": "arm64",
      "tuning": {
        "GOARM64": "v8.0"
      }
    },
    "toolchain": {
      "algorithm": "curator-go-toolchain-v1",
      "go_relpath": "bin/go",
      "go_version": "go version go1.26.1 darwin/arm64",
      "content_sha256": "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
    },
    "policy": {
      "module_mode": "vendor",
      "network": "none",
      "workspace": false,
      "cgo": false,
      "compiler_directives": "reject-nonstandard-cgo-import-dynamic-v1",
      "target_mode": "native",
      "link_mode": "internal",
      "libgcc": "none",
      "package_assembly": false,
      "host_objects": false,
      "telemetry": "off-private"
    }
  },
  "artifact": {
    "path": "bin/golden-tool",
    "sha256": "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
    "size": 1234567
  }
}
```

The stored logical receipt bytes MUST equal `CCJ-1(receipt)` exactly: UTF-8, no BOM, no insignificant whitespace, and no terminal newline. The receipt schema forbids a top-level `sig`, so CCJ-1’s signed-object omission rule cannot remove data. Define:

```text
receipt_sha256 = "sha256:" + lowercase_hex(SHA-256(exact_stored_receipt_bytes))
```

For the receipt above, the exact value is `sha256:750f5f7557ffe1dac1ec06b7d47cb5976e390a7dc7d57ea616c9edda91013c11`. Pretty-printed display bytes are not hash input. Writers canonicalize before publication; readers parse with duplicate-key/integer checks, recanonicalize, require byte equality with the stored file, and then hash those same bytes. This digest is a deterministic corruption/currentness identifier. It is not a signature, MAC, or independent attestation that `go-v1` produced the artifact; provenance comes from the separately verified manager-protected-state boundary. The receipt schema MUST NOT add a self-asserted `trusted`, `provenance`, or `manager_created` boolean: an attacker could copy it. Protection is out-of-band manager state and is checked before receipt admission.

Receipt validation is semantic as well as schema-based:

1. Verify the implementation-specific protected-state boundary from section 5.3 before treating any persistent entry as reusable. Boundary failure forces a miss/rebuild and is not repaired by self-consistent receipt bytes.
2. Fully validate the raw snapshot and recompute `input.build_source` before opening a candidate entry; it must match the expected plan identity.
3. Recompute the cache key from canonical `input`; it must match both the receipt and logical entry key.
4. Match the entire expected input object, not selected fields, including `build_root`, `source_dir`, the build-source algorithm/digest, and `policy.compiler_directives`. This is a necessary consistency check; it does not by itself prove who wrote the artifact.
5. Require the artifact path to equal the manager-derived platform path.
6. Open relative to the protected root without following links, require one regular singly linked file, bound its size, and recompute SHA-256 and byte length.
7. Reject unknown receipt fields and unsupported schema/driver/build-source versions.
8. Never execute an artifact to validate it.

The required forged-hit conformance vector closes the distinction. The `candidate.receipt` below has the exact expected input/key, its 24-byte regular executable candidate matches its own hash and size, and its canonical receipt digest is self-consistent. It is still rejected because it came from a cache boundary the manager did not create/protect. The receipt hash is over `CCJ-1(candidate.receipt)`, not over this enclosing test vector.

```json
{
  "name": "reject-self-consistent-forged-hit-outside-protected-state",
  "cache_boundary": {
    "manager_created": false,
    "owner_matches_manager_principal": false,
    "other_principals_can_write": true,
    "all_components_link_safe": true
  },
  "candidate": {
    "artifact_bytes_base64": "YXR0YWNrZXItY2hvc2VuLWFydGlmYWN0",
    "artifact_is_regular_executable": true,
    "receipt": {
      "schema_version": 1,
      "cache_key": "sha256:3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48",
      "input": {
        "schema_version": 1,
        "driver": "go-v1",
        "build_source": {
          "algorithm": "curator-build-source-v1",
          "content_sha256": "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
        },
        "build_root": "build",
        "command": "golden-tool",
        "source_dir": "build/cmd/golden-tool",
        "target": {
          "goos": "darwin",
          "goarch": "arm64",
          "tuning": {
            "GOARM64": "v8.0"
          }
        },
        "toolchain": {
          "algorithm": "curator-go-toolchain-v1",
          "go_relpath": "bin/go",
          "go_version": "go version go1.26.1 darwin/arm64",
          "content_sha256": "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
        },
        "policy": {
          "module_mode": "vendor",
          "network": "none",
          "workspace": false,
          "cgo": false,
          "compiler_directives": "reject-nonstandard-cgo-import-dynamic-v1",
          "target_mode": "native",
          "link_mode": "internal",
          "libgcc": "none",
          "package_assembly": false,
          "host_objects": false,
          "telemetry": "off-private"
        }
      },
      "artifact": {
        "path": "bin/golden-tool",
        "sha256": "sha256:a4f06a1304c926ed7f2326c8fd90cabc5c5bd2981e690a4351c852d91c079d88",
        "size": 24
      }
    },
    "receipt_sha256": "sha256:9a23f5b77e6173b0f10e7ed43cd2b21aa3b99f3a34945ec432fbb31338a6186d",
    "internal_checks": {
      "cache_key_matches": true,
      "input_matches": true,
      "artifact_hash_matches": true,
      "artifact_size_matches": true,
      "receipt_hash_matches": true
    }
  },
  "expected": {
    "result": "untrusted-provenance",
    "reuse": false,
    "marker_current": false,
    "real_operation": "rebuild-from-validated-snapshot-into-protected-state",
    "dry_run": "would-rebuild-untrusted-cache"
  }
}
```

A malformed/mismatched entry is cache corruption. Under the manager-home mutation lock, quarantine it outside the live namespace or replace it only via a fresh staged build. Do not silently use it. A dry-run reports corruption but neither locks for mutation, quarantines, nor rebuilds.

### 5.5 Install marker v2

Introduce `install-marker-v2.schema.json`. New managers read marker v1 and v2, write v2 on any installation mutation, and may continue to regard a valid v1 marker as current for schema 1–5 packages. Marker v2 raises `skill_schema_version` to 6 and requires sorted `build_roots` plus `builds` (empty arrays/objects for installs without compiled commands). It requires `build_source` exactly when `builds` is non-empty; installs without active compiled commands omit that field.

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
  "build_source": {
    "algorithm": "curator-build-source-v1",
    "content_sha256": "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
  },
  "locale": null,
  "agents": [],
  "commands": ["golden-tool"],
  "dependencies": [],
  "skill_schema_version": 6,
  "runtime_roots": [],
  "build_roots": ["build"],
  "installed_at": "2026-07-20T00:00:00Z",
  "files": ["SKILL.md", "agent-skill.json"],
  "builds": {
    "golden-tool": {
      "driver": "go-v1",
      "cache_key": "sha256:3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48",
      "receipt_sha256": "sha256:750f5f7557ffe1dac1ec06b7d47cb5976e390a7dc7d57ea616c9edda91013c11",
      "artifact_sha256": "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
      "artifact_path": "bin/golden-tool"
    }
  }
}
```

The existing `content_sha256` is still recomputed over the installed tree with the manager marker excluded. For build-enabled markers, currentness additionally requires `build_roots` and the static context exclusion to match the manifest, `build_source` to match the fully validated raw snapshot selected by the effective plan, every receipt input, and every derived cache key, and the referenced logical cache entry to remain below a currently verified manager-protected-state boundary. A package-provided marker is therefore represented by `build_source` even though installation overwrites that path with the manager marker. Marker hashes pin the selected bytes but do not authenticate an unprotected cache retroactively. A missing raw snapshot, mismatched build-source identity/root declaration, prompt-visible build-root file, untrusted cache boundary, corrupt/wrong-target/wrong-toolchain entry, or invalid receipt makes status non-current (or unknown where the existing profile distinguishes it), and repair rebuilds from a revalidated snapshot into protected state. Canonical marker writing sorts `commands`, `dependencies`, `files`, `runtime_roots`, `build_roots`, and `builds` keys.

The shim must point to the immutable build-cache artifact selected by the marker/plan, not to `<home>/runtime/<skill>/<commit>`. Shims remain manager-generated, self-contained, and independent of shell profiles.

### 5.6 Garbage collection

GC runs while holding the manager-home mutation lock and traverses only a revalidated manager-protected cache root. It marks logical build keys referenced by all valid marker v2 files and in-flight transaction journals, then sweeps unreferenced entries older than the normal grace period. If the cache boundary or a marker cannot be validated safely, GC does not adopt or execute entry content; it retains/quarantines conservatively according to the implementation-specific storage policy and reports the uncertainty. It never uses receipt content alone as proof of provenance or a live consumer.

## 6. Lifecycle, ordering, dry-run, and rollback

### 6.1 Normative ordering

For a real operation, hold a same-project operation lock from planning through handoff, and use provider-first closure order. Dry-run is lock-free and relies on the generation recheck in step 7. In addition, every real manager-home mutation uses one exclusive **manager-home mutation lock** shared by project, global, hybrid, adapter, consumer, cache-publication, recovery, and GC paths. A conforming implementation performs:

1. For a real operation, acquire the canonical project operation lock, briefly acquire the manager-home mutation lock, recover every incomplete transaction journal, and release the home lock before network, audit, fingerprinting, or compilation. A dry-run acquires neither lock and performs no recovery write.
2. Resolve refs and immutable raw snapshots; parse both manifest spellings; fully validate every snapshot entry, `build_roots`, context exclusion, and semantic path; compute `curator-build-source-v1` over every regular file before any cache lookup; and freeze the validated snapshot instances for the operation.
3. Build the dependency closure and activation plan; check command/shim/case collisions and system/MCP requirements.
4. Run source audit, registry policy, attestation, and moved-tag gates without executing package code.
5. Create an operation-private probe root; run fixed `go telemetry off`, `go version`, and `go env`; verify private `off` state and native target; fingerprint the trusted Go tree byte-for-byte; and compute every active build input/cache key using the already validated build-source identity. Within a node, sort build command names lexically.
6. Validate the implementation-specific protected-cache boundary, then validate eligible receipts/artifacts read-only and produce a complete hit/miss/corrupt/untrusted-provenance plan. A boundary failure is a forced miss, even when all receipt hashes match. Capture generation digests for every shared ledger/target consulted; these are optimistic observations, not commit authority.
7. **Dry-run re-reads the generation digests, deletes the probe root, and ends here.** If shared state changed, retry the read-only plan or report `concurrent_state_change`; never create a lock file, journal, or cache. A miss is `would-preflight-and-build`, because dry-run deliberately does not run source-aware `go list`.
8. For a real operation, retain the private root and build every miss into operation-private staging, in provider-first/node-command order, using the fixed full-graph `go list`, build-root/input/directive checks, and fixed `go build`. Do not publish anything. Any snapshot/toolchain change aborts the operation.
9. Verify and hash every staged artifact and generate canonical receipts. If any build fails, delete operation staging and leave installation, consumers, and live caches unchanged.
10. Acquire the manager-home mutation lock. Recover any journal left since step 1, then revalidate the protected-cache boundary and re-read every eligible cache entry plus all project/global/hybrid/adapter/consumer/GC state that the plan will read or write. If a protected winner entry is valid and identical, discard the staged duplicate. If any change alters closure, activation, cache trust, target ownership, expected preimage, or required build key, release the lock and restart from the earliest affected read/build step; never apply a stale plan.
11. Under that lock, atomically publish missing immutable cache entries or quarantine/replace corrupt ones only from already verified staging. If the old cache root is untrusted, create a fresh implementation-specific protected root and rebuild/publish; never make the old directory private and then adopt its candidate bytes. Create one transaction journal containing a unique transaction id, canonical project identity, ordered target list, expected generation/preimage digest, backup path, desired digest, and commit state for every mutable target.
12. Stage and commit all mutable targets in deterministic classes: project/global contexts and markers; runtime/shim/environment targets; adapter ledgers and hybrid/global mirrors; stale managed removals; then the machine-wide consumer ledger **last**. Sort canonical target identifiers bytewise within a class. Preserve backups until the last target and consumer ledger are durable.
13. If any publication or target swap fails, keep the manager-home lock and restore journaled targets in exact reverse commit order. A rollback compares the current target digest with the journal's desired digest before restoring; any mismatch is an implementation-corruption error, not permission to overwrite unknown state. No other project can commit while rollback runs, so the restored preimage can never predate another project's successful commit.
14. After success, remove backups and mark/remove the journal durably. Run runtime/snapshot/build-cache GC while still holding the manager-home mutation lock; a GC failure is a reported maintenance warning and does not revert the installation. Release the home lock, then the project lock.

Lock ordering is fixed: acquire project operation locks by canonical project identity in unsigned UTF-8 byte order (one lock for a single-project operation); use any optional cache-build lock one key at a time and release it; then acquire the manager-home mutation lock. Never acquire a project/cache lock while holding the home lock. Recovery and GC acquire only the home lock. This rule covers multi-project `--all` operations and prevents lock cycles.

Compilation stays outside the manager-home lock, so independent projects may still validate and build concurrently. Shared publication/commit/rollback/GC is deliberately serialized in v1. The Python manager already wraps install/update/upgrade and GC in a coarse home `GlobalLock`, though it lacks this journal/revalidation contract and currently holds the lock through the whole operation ([Python lock](https://github.com/ivanopcode/cocoaskills/blob/6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12/src/csk/locking.py#L14-L47); [CLI lock scope](https://github.com/ivanopcode/cocoaskills/blob/6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12/src/csk/cli.py#L542-L610)). The Go manager's consumer read-modify-write path has no corresponding shared lock ([Go consumers](https://github.com/relux-works/curator/blob/17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8/internal/scopes/consumers.go#L14-L61)). The protocol must define the isolation rule rather than inherit either implementation's accident.

The command build order affects observability only; cache keys make outputs independent. Implementations may later parallelize private builds, but normalized diagnostics and the locked commit order remain normative.

### 6.2 Dry-run

Dry-run may:

- fully validate the raw snapshot and compute `curator-build-source-v1`, including root `.csk-install.json`;
- validate `build_roots`, exclude them from the prospective context file set, and reject a `source_dir`/module root outside that static boundary without invoking Go;
- resolve an operator-trusted Go binary;
- create one temporary probe root, run the fixed package-independent `go telemetry off`/`go version`/`go env` sequence, verify that telemetry is `off` below that root, and remove the root before return;
- compute the toolchain fingerprint directly or use an already validated read-only memo;
- read receipts, artifacts, markers, registry data, and policy state;
- inspect cache-root ownership/ACL/mode and containment without changing them;
- report the build-source algorithm/digest, target, driver, key, and `cache-hit`, `would-preflight-and-build`, `would-rebuild-untrusted-cache`, `corrupt`, or `unsupported` result for each command.

For the section 5.2 fixture, a miss is reported without running source-aware Go commands:

```json
{
  "command": "golden-tool",
  "driver": "go-v1",
  "build_source": {
    "algorithm": "curator-build-source-v1",
    "content_sha256": "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
  },
  "build_root": "build",
  "source_dir": "build/cmd/golden-tool",
  "cache_key": "sha256:3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48",
  "target": {
    "goos": "darwin",
    "goarch": "arm64",
    "tuning": {
      "GOARM64": "v8.0"
    }
  },
  "result": "would-preflight-and-build"
}
```

Dry-run must not:

- invoke `go list`, `go build`, or any compiler/linker;
- retain the private telemetry/config probe root after return, or create `GOCACHE`, `GOMODCACHE`, temp build directories, fingerprint memos, audit records, registry response caches, mutation-lock files, runtime/build cache entries, transaction journals, shims, markers, or consumer records;
- quarantine a corrupt cache entry;
- repair permissions/ACLs or adopt bytes from an untrusted cache boundary;
- change atime where the platform/API allows no-atime reads.

This also means the Python registry path identified above needs a genuine read-only mode before it can host build-driver dry-runs.

### 6.3 Rollback rules

- A resolution, validation, gate, toolchain, cache-inspection, or build failure occurs before commit and leaves the installed state byte-for-byte unchanged.
- A cache publication or target-swap failure restores project/global/hybrid manager-owned targets from the journal in reverse order, including prior shims, environment files, adapters, contexts, markers, and consumer ledger, while the exclusive manager-home lock is still held.
- A transaction-owned cache entry may be removed during rollback only while the home lock is held and a fresh mark proves that no valid marker or journal references it. Retaining an unreferenced immutable entry is always safe; normal locked GC removes it after the grace period.
- Existing valid immutable cache entries are never modified during rollback.
- An entry observed outside manager-protected state is never considered pre-existing valid state or transaction-owned publication merely because its receipt is internally consistent.
- Backups are retained until rollback succeeds. An interrupted transaction is recovered under the manager-home lock before any new shared mutation, regardless of which project starts the next operation. Recovery owns journals by transaction id, not by the current project.
- No transaction releases the home lock between its first target swap and successful commit or reverse rollback. A second project's success can therefore neither be lost from the consumer ledger nor be overwritten by a stale backup.
- The built artifact is never used as a rollback program or verifier.

This is a manager-home-wide isolation requirement, not merely per-skill or per-project directory replacement.

The required concurrent vectors make the lost-update and stale-restore cases explicit:

```json
[
  {
    "name": "two-project-success-preserves-both-consumers",
    "initial_consumers": [],
    "private_builds": ["project-a", "project-b"],
    "locked_commits": [
      {"project": "project-a", "read_consumers": [], "write_consumers": ["project-a"]},
      {"project": "project-b", "read_consumers": ["project-a"], "write_consumers": ["project-a", "project-b"]}
    ],
    "expected_consumers": ["project-a", "project-b"],
    "lost_update": false
  },
  {
    "name": "successful-project-survives-other-project-rollback",
    "initial_consumers": [],
    "locked_commits": [
      {"project": "project-a", "result": "commit", "post_consumers": ["project-a"]},
      {
        "project": "project-b",
        "preimage_consumers": ["project-a"],
        "failure_at": "adapter-ledger-swap",
        "result": "reverse-rollback",
        "post_consumers": ["project-a"]
      }
    ],
    "expected_project_a_targets": "committed",
    "expected_project_b_targets": "pre-transaction",
    "stale_backup_restored_over_project_a": false
  }
]
```

## 7. Candidate language/toolchain classification

The test is not “can this ecosystem build offline?” It is “can the manager invoke a fixed trusted compiler without allowing the package to select executable host code, hooks, plugins, or argument arrays?”

| Ecosystem | v1 classification | Security rationale and possible constrained future |
|---|---|---|
| **Go** | **Accept as `go-v1`** | Direct fixed Go 1.23+ commands support one package operand. A dedicated context-excluded build root, vendor mode, private telemetry-off state, local toolchain, full dependency inspection, rejection of non-standard assembly/host objects and `go:cgo_import_dynamic`, disabled cgo/PGO/workspaces/VCS stamping, forced internal linking with no libgcc fallback, and an empty executable-search path close the identified package-selected process/input paths. Output is never launched. |
| **Rust** | Defer/reject Cargo | Cargo compiles and executes package `build.rs` before building the package, and procedural macros run during compilation with compiler-process resources. A future driver would need to reject build scripts, proc-macro dependencies, compiler wrappers, linker selection, and Cargo config; direct `rustc` still needs a dependency/linking design. [Cargo build scripts](https://doc.rust-lang.org/cargo/reference/build-scripts.html); [procedural macros](https://doc.rust-lang.org/reference/procedural-macros.html). |
| **Zig** | Defer | `zig build` evaluates package `build.zig` and can run system/project tools. Zig also deliberately evaluates `comptime` code during compilation, which needs a precise side-effect/resource model before acceptance. A future fixed `zig build-exe` subset cannot inherit build.zig/package arguments. [Zig build system](https://ziglang.org/learn/build-system/); [Zig language reference](https://ziglang.org/documentation/master/#comptime). |
| **Swift** | Defer | SwiftPM package manifests are Swift code that runs, and command/build-tool plugins execute separate processes. A future direct `swiftc` subset would need to forbid package plugins/macros, own SDK/target/link flags, and solve dependency discovery without SwiftPM. [SwiftPM plugins](https://docs.swift.org/swiftpm/documentation/packagemanagerdocs/plugins/). |
| **C/C++ direct compilers** | Defer | A future direct compiler driver could own an exact source-file enumeration, sysroot, include/preprocessor/linker flags, response-file grammar, compiler plugins, and native dependency identity, but that portable trusted-input surface is substantially larger than Go v1. Package Make/CMake/Meson files cannot provide those arrays in the current contract. |
| **Make/CMake/Meson** | Prohibit without a separate sandboxed driver | Make recipes run commands; CMake can run processes during configure and declare custom commands; Meson `run_command()` executes a selected command during setup and `custom_target(command:)` executes a package-selected argv during the build. These frontends are recipes, not fixed compilers. [GNU Make recipe execution](https://www.gnu.org/software/make/manual/html_node/Execution.html); [CMake `execute_process`](https://cmake.org/cmake/help/latest/command/execute_process.html); [CMake custom commands](https://cmake.org/cmake/help/latest/command/add_custom_command.html); [Meson `run_command`](https://mesonbuild.com/Reference-manual_functions_run_command.html); [Meson `custom_target`](https://mesonbuild.com/Reference-manual_functions_custom_target.html). |
| **Java/Kotlin** | Defer | Maven and Gradle run package-selected plugins/tasks. `javac` discovers and invokes annotation processors unless processing is disabled; Kotlin supports compiler plugins and kapt processors. A future direct `javac -proc:none` subset could own classpath/output/main wrapping, but mixed JVM builds are not v1-safe. [javac annotation processing](https://docs.oracle.com/en/java/javase/23/docs/specs/man/javac.html); [Maven plugins](https://maven.apache.org/guides/introduction/introduction-to-plugins.html); [Gradle tasks](https://docs.gradle.org/current/userguide/implementing_custom_tasks.html); [Kotlin compiler plugins](https://kotlinlang.org/docs/compiler-plugins-overview.html). |
| **.NET** | Defer | `dotnet build` delegates to MSBuild; project/imported files can select executable tasks, inline compiled tasks, and `Exec`. Roslyn source generators are compile-time metaprograms. A future direct `csc` subset must disable analyzers/generators and own references/target/runtime packaging. [MSBuild tasks](https://learn.microsoft.com/en-us/visualstudio/msbuild/msbuild-tasks); [inline tasks](https://learn.microsoft.com/en-us/visualstudio/msbuild/msbuild-inline-tasks); [`Exec`](https://learn.microsoft.com/en-us/visualstudio/msbuild/exec-task); [Roslyn compiler platform](https://learn.microsoft.com/en-us/dotnet/csharp/roslyn-sdk/). |
| **Node/TypeScript (`tsc`)** | Defer | npm install lifecycle scripts and default `node-gyp` behavior execute package-selected code. A fixed trusted TypeScript compiler could transpile a closed source graph with package discovery/configuration disabled, but output is not self-contained and dependency installation/runtime resolution remains unresolved under this contract. [npm scripts](https://docs.npmjs.com/cli/using-npm/scripts/). |
| **Node/TypeScript bundlers** | Prohibit config/plugin mode; defer any fixed no-plugin subset | Webpack accepts executable JavaScript/TypeScript configuration, loaders run in Node.js with full Node capability, and plugins receive the compilation lifecycle. Esbuild's plugin API injects JavaScript or Go callbacks into resolution/loading. Invoking a repository config, loader, or plugin directly violates no-package-code-during-install. A future pinned CLI subset would need a manager-owned entry graph/options, no config discovery, no plugins/loaders, pre-vendored dependencies, and a separately specified runtime/output model. [webpack configuration languages](https://webpack.js.org/configuration/configuration-languages/); [webpack loaders](https://webpack.js.org/concepts/loaders/); [webpack plugins](https://webpack.js.org/concepts/plugins/); [esbuild plugins](https://esbuild.github.io/plugins/). |
| **Deno** | Promising, but defer | `deno compile` is closer to a direct driver and offers cached/offline controls, but current behavior includes config-driven options, preload/require execution, npm lifecycle permissions, and framework detection that may run a build script. A future version-pinned driver needs one explicit entry file, `--no-config`, no preload/require, frozen pre-vendored dependencies, no scripts, and audited behavior. [Deno compile](https://docs.deno.com/runtime/reference/cli/compile/); [Deno packages](https://docs.deno.com/runtime/packages/). |
| **Python** | Defer | pip/build frontends delegate wheel creation to the package-selected PEP 517 backend and invoke backend hooks; build requirements may also be computed dynamically. That is package-controlled host execution. `compileall` only produces bytecode and does not solve dependencies or a self-contained command. [pip build-system interface](https://pip.pypa.io/en/latest/reference/build-system/); [PyPA packaging flow](https://packaging.python.org/en/latest/flow/). |

The protocol should document this table as rationale, not reserve generic driver names. Each future driver needs a separate threat-model review, closed schema, fixed process graph, cache identity, and conformance vectors.

## 8. Protocol artifacts affected

### 8.1 Normative protocol and schemas

| Artifact | Required change |
|---|---|
| `protocol/core.md` | Add schema 6 and `build_roots`; exclude those roots from agent context and runtime copying; define build command/source/output rules, no-hooks boundary, the distinction between legacy marker-excluding `content_sha256` and length-framed all-file `curator-build-source-v1`, logical cache/receipt identity, marker v2, artifact currentness, and the fact that physical manager-home/cache names remain implementation-specific. |
| `profiles/manager.md` | Add full raw-snapshot/build-root validation and build-source digesting before cache lookup; driver discovery; byte-exact toolchain identity; private telemetry initialization; full dependency and compiler-directive checks; fixed internal-link process/environment; build planning/order; read-only dry-run; manager-home commit/rollback/GC lock, deterministic lock/target order, global recovery journal, protected-cache ownership/ACL/link checks with rebuild-on-untrusted fallback, logical receipts, shims, repair/status, and GC. Correct consumer ordering and required-path reuse language where needed. |
| `SECURITY.md` | Clarify trusted compiler versus untrusted source/output; list package-provided marker/embed and `go:cgo_import_dynamic` inputs plus forbidden hooks/plugins/build systems (including Meson and JS bundlers), external-linker/libgcc, host-object, and assembly-include surfaces; document compiler-input DoS and sandbox expectations; state that receipt SHA-256 is not provenance, manager-protected state is in the TCB, and same-principal/admin compromise is out of v1 scope. |
| `schemas/v1/common.schema.json` | Add `buildCommandV6`, `commandV6`, strict `buildSourceIdentity`, and receipt/marker supporting definitions. `source_dir` remains a portable path; do **not** broaden the existing schema-1–5 `command` union. |
| `schemas/v1/agent-skill-v6.schema.json` | New canonical manifest schema referencing `commandV6` and admitting `build_roots`. |
| `schemas/v1/csk-skill-v6.schema.json` | New legacy-name mirror with the same `build_roots`/command semantics. |
| `schemas/v1/build-receipt-v1.schema.json` | New strict receipt schema whose canonical input requires `curator-build-source-v1` plus artifact identity; document the receipt as deterministic corruption/currentness metadata, not a signature or MAC. |
| `schemas/v1/install-marker-v2.schema.json` | New reader/writer shape with `skill_schema_version <= 6`, sorted `build_roots`, required `builds`, and `build_source` present exactly when builds are active. |
| `schemas/v1/conformance-claim-v2.schema.json` | New current claim shape with `schema_version: 2` and `protocol_version: "1.0.0-rc.4"`; keep claim v1 frozen at rc.3 so old claims retain their meaning. |
| `schemas/v1/README.md` | Index and compatibility notes for all new schemas, including claim-v1 historical/current-v2 writer and reader rules. |
| `cli/curator.md` | Document that install/upgrade may compile only closed build commands, that `--dry-run` reports per-command build root/key/cache status without compiling, and that `status --json`/`--check` report missing/corrupt receipts, artifact drift, context-visible build sources, and unsupported drivers. No package argv or generic build-hook CLI flag is added. |
| `decisions/0004-compile-only-build-drivers.md` | Record the closed-driver/no-hooks decision, dedicated context-excluded build roots, Go-only v1 scope, fixed compiler/linker/directive policy, global commit isolation, and why the other ecosystems remain deferred/prohibited. |
| `README.md`, `COMPATIBILITY.md`, `CHANGELOG.md`, `RELEASE.md` | Publish protocol `1.0.0-rc.4`, schema/marker/claim reader-writer compatibility, build-root migration, security impact, conformance level/version metadata, and release-gate expectations. |

No v1 build-policy fields should be added to `manager-config-v1` or `system-config-v1`; fixed semantics are the security feature. `protocol/registry.md` and audit-record schemas need no new artifact attestation in v1: registry audit continues to attest the untrusted source snapshot, while the local receipt records the compiled result, snapshot, and trusted toolchain inside manager-protected state. It is not a registry attestation or cryptographic provenance proof. The security text should state this distinction explicitly.

### 8.2 Conformance artifacts

Do not change the meaning of the stable `conformance-claim-v1.schema.json` URL in place. Freeze it at `schema_version: 1` / `protocol_version: "1.0.0-rc.3"` and add this current-writer schema for the release that first carries manifest schema 6:

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://relux-works.github.io/curator-spec/schemas/v1/conformance-claim-v2.schema.json",
  "title": "Curator Protocol conformance claim schema 2",
  "type": "object",
  "required": ["schema_version", "protocol_version", "implementation", "implementation_version", "classes", "suite_sha256", "operating_systems", "created_at", "result"],
  "properties": {
    "schema_version": {"const": 2},
    "protocol_version": {"const": "1.0.0-rc.4"},
    "implementation": {"$ref": "common.schema.json#/$defs/nonEmptyString"},
    "implementation_version": {"$ref": "common.schema.json#/$defs/nonEmptyString"},
    "classes": {
      "type": "array",
      "minItems": 1,
      "items": {"enum": ["core", "manager", "registry-client", "registry-service"]},
      "uniqueItems": true
    },
    "suite_sha256": {"$ref": "common.schema.json#/$defs/sha256"},
    "operating_systems": {
      "type": "array",
      "minItems": 1,
      "items": {"enum": ["linux", "macos", "windows"]},
      "uniqueItems": true
    },
    "created_at": {"$ref": "common.schema.json#/$defs/timestamp"},
    "result": {"const": "pass"}
  },
  "additionalProperties": false
}
```

The generated valid schema case is exact apart from the suite fixture digest chosen by the generator:

```json
{
  "classes": ["core", "manager"],
  "created_at": "2026-07-20T00:00:00Z",
  "implementation": "example-manager",
  "implementation_version": "1.0",
  "operating_systems": ["linux"],
  "protocol_version": "1.0.0-rc.4",
  "result": "pass",
  "schema_version": 2,
  "suite_sha256": "sha256:0000000000000000000000000000000000000000000000000000000000000000"
}
```

Claim transition semantics are normative:

1. rc.4 writers emit only claim v2. Readers dispatch by `schema_version`, may retain v1 as historical rc.3 evidence, and never treat a v1 claim as evidence for rc.4/schema-6/build-driver behavior.
2. Old readers reject v2 with the normal unsupported-version/upgrade result; no reader infers v2 from fields or rewrites `protocol_version` in a v1 object.
3. `conformance/v1/manifest.json` changes its release identity to `1.0.0-rc.4` and includes both frozen v1 schema cases and current v2 cases. A current manager-class claim requires the build-driver/context/cache/lifecycle vectors added here; passing only the rc.3 suite cannot produce v2.
4. The v2 invalid schema cases cover `schema_version: 1`, `protocol_version: "1.0.0-rc.3"`, unknown fields, duplicate classes, and `result: "fail"`. Separately, claim verification/release-gate cases reject a syntactically valid `suite_sha256` that is not the generated rc.4 suite digest, which is how missing build-driver coverage is detected. The frozen v1 cases remain unchanged.
5. The generator must split its current single version constant: `protocolVersion = "1.0.0-rc.4"` drives the manifest/current v2 case, while a separate immutable `conformanceClaimV1ProtocolVersion = "1.0.0-rc.3"` drives regeneration of the historical v1 cases. Reusing the rc.4 constant for claim v1 would silently undo the compatibility decision.

At the inspected ref, the claim/release pin occurs in the claim-v1 schema/cases, generator constant, generated manifest, validator, top-level README, and changelog; `conformance/README.md` also names claim v1 as current ([claim-v1 schema](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/schemas/v1/conformance-claim-v1.schema.json), [claim cases](https://github.com/relux-works/curator-spec/tree/57c1f56846d221ecc55786bd3c2467ec32f11730/conformance/v1/schema-cases/conformance-claim-v1), [generator](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/tools/generate-vectors/main.go#L23), [validator](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/tools/validate.py#L102-L103), [claim documentation](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/conformance/README.md#L115-L120)). The two inspected managers contain no claim emitter at `origin/main`; their required change is to consume/pass the rc.4 suite before anyone issues a v2 manager claim.

Add or update:

- `conformance/v1/schema-cases/index.json`;
- valid/invalid cases for `agent-skill-v6`, `csk-skill-v6`, `build-receipt-v1`, `install-marker-v2`, and `conformance-claim-v2`, while retaining frozen claim-v1 cases;
- a separate Go fixture with `build/go.mod`, `build/cmd/golden-tool`, a transitive embedded input, and a vendored-dependency variant, while leaving the existing script golden fixture and registry hashes intact;
- expected context-file vectors proving that `build_roots` are excluded on real installs, cache hits, and dry-runs while `SKILL.md` and unrelated eligible assets remain visible;
- a byte-level `curator-build-source-v1` vector covering the domain prefix, unsigned path ordering, uint64be lengths, empty/binary files, root `.csk-install.json`, no mode/timestamp input, invalid paths, links/special files, duplicate paths, and mutation during use;
- a structural-collision vector proving that legacy `a || NUL || x || NUL || b || NUL || y` can describe one binary file or two records, while the new length-framed build-source digests differ;
- expected CCJ-1 build-input bytes with `build_source`, cache key, exact stored receipt bytes, receipt SHA-256, marker, artifact SHA-256, and dry-run plan;
- a byte-level `curator-go-toolchain-v1` vector covering unsorted entries, regular files, directories, internal links, mode/timestamp non-inputs, LF/CRLF normalization, duplicate paths, invalid Unicode, escaping links, and a selected executable outside `GOROOT`;
- `conformance/v1/manifest.json` with `protocol_version: "1.0.0-rc.4"`, `conformance/README.md` with v1/v2 claim transition rules, and `schemas/v1/README.md` with both claim schemas;
- `conformance/v1/vectors/manager-lifecycle.json` for provider/build order, no-mutation dry-run, build failure, cache race/corruption, two-project success, success-versus-rollback, deterministic lock/target ordering, interrupted global-journal recovery, currentness, repair, and locked GC;
- `conformance/v1/vectors/build-drivers.json` for build-root/context rules, environment, all five direct argv forms, full dependency/directive validation, target, pre-cache build-source identity, cache-key, toolchain-digest, and receipt-byte semantics;
- `tools/generate-vectors/main.go` (rc.4 current constant, frozen rc.3 claim-v1 constant, plus claim-v2 cases) and its tests for the new canonical bytes/hashes; `tools/validate.py` (rc.4 manifest assertion), `tools/release_gate.py`, and their tests/manifest inventory for new schemas, vectors, decision, claim transition, and release metadata;
- release-facing `README.md`, `COMPATIBILITY.md`, `CHANGELOG.md`, and `RELEASE.md`.

Minimum negative vectors:

- build command in schema 5;
- unknown driver or any `args`/`env`/`output` field;
- missing/unused/overlapping `build_roots`; overlap with `runtime_roots`; `source_dir: "."`; source outside its build root; nested/intervening `go.mod`; build-root file appearing in expected context; source path escape/link/non-directory/no root `go.mod`/non-main/multiple-package mismatch;
- dependency absent from `vendor`, inconsistent `modules.txt`, workspace-only dependency, newer auto-toolchain request, pre-Go-1.23/unknown toolchain family, and cgo-only package;
- attempted `go:generate` file proving no generator execution;
- root and transitive dependency `.syso`; root and transitive `.s` with absolute and escaping `#include`; non-standard active C/SWIG inputs; embedded/package paths outside their declared build root; and an active non-standard Go file containing `//go:cgo_import_dynamic` while every go-list native field is empty;
- two rejected `source_dir: "."` snapshots using `//go:embed .csk-install.json` and differing only in that root file: their legacy installed-tree hashes are equal, their all-file build-source digests differ, semantic validation rejects both before cache-key construction, and no Go command runs;
- poisoned `cc`/`clang`/`gcc`/`ld`/`pkg-config`, a toolchain whose external-link default is enabled, and a target requiring external linking; assert no executable outside fingerprinted `GOROOT/pkg/tool/<host>` starts and that the build fails rather than falls back;
- poisoned `PATH`, `GOFLAGS=-toolexec=...`, `GOENV`, `GOWORK`, VCS metadata, `default.pgo`, parent workspace, and repository-local fake `go`;
- ignored `GOTELEMETRY=off`, private `go telemetry off` success/failure, private-directory validation, and deletion of telemetry state on real/dry-run exit;
- cache key mismatch, wrong target/toolchain, receipt/artifact hash mismatch, noncanonical receipt whitespace/trailing LF, link/special artifact, partial entry, and concurrent publisher;
- the exact-input/exact-key/self-consistent forged receipt/artifact from section 5.4 under an untrusted cache root; all internal hashes pass, protected-state validation fails, dry-run reports `would-rebuild-untrusted-cache`, and the real operation rebuilds rather than adopts it;
- claim v1 with rc.4, claim v2 with rc.3 or schema version 1, and an rc.3 manager result presented as rc.4 build-driver evidence;
- build 2 failure after build 1 succeeds with no persistent changes;
- target-swap failure with full reverse rollback and unchanged consumer ledger;
- concurrent projects A/B both succeeding with the consumer ledger containing both, and A succeeding before B fails with B rollback preserving every A-owned shared target;
- dry-run assertions covering every forbidden persistent path, including Python registry response/state caches.

## 9. Decision and anomaly log

These findings are important beyond the schema shape and should be tracked during implementation/review:

1. **Decision:** `go-v1` is a closed protocol driver, not an adapter to arbitrary Go commands. Any semantic change creates a new driver revision.
2. **Decision:** every v1 Go module lives below one declared `build_root`; build roots are excluded from agent context, never runtime-copied, and cannot be `.` or overlap runtime roots. This static restriction is intentionally stricter than graph-derived exclusion so cache hits and dry-runs behave identically without `go list`.
3. **Decision:** vendored, native, cgo-disabled builds are the only v1 dependency/target model.
4. **Decision:** v1 rejects non-standard assembly, all active `.syso`, and the exact `//go:cgo_import_dynamic` token in active non-standard Go files; it forces internal linking and disables libgcc fallback rather than attempting filesystem confinement around package-selected native inputs.
5. **Decision:** build output uses a separate immutable cache and receipt; `<skill>/<commit>` runtime identity is not reused for compiled artifacts.
6. **Decision:** marker v2 pins build roots, build receipts, and artifacts; old marker v1 remains readable for pre-v6 packages.
7. **Decision:** dry-run does not invoke source-aware Go commands, create mutation locks/journals, or repair state, even if an implementation could isolate their caches.
8. **Decision:** Go telemetry is disabled by `go telemetry off` in an operation-private config root; `GOTELEMETRY=off` is not used because the value is read-only.
9. **Decision:** `curator-go-toolchain-v1`, `curator-build-source-v1`, CCJ-1 cache-input bytes, and exact canonical receipt-file bytes are protocol algorithms with cross-language vectors, not implementation serializer choices.
10. **Decision:** installed-tree `content_sha256` remains marker-excluding for backward-compatible currentness, but it is never a build-cache identity. Build-source identity is domain-separated, length-framed, covers every raw regular file including root `.csk-install.json`, and is computed before cache lookup.
11. **Decision:** independent projects may validate/build concurrently, but cache publication, project/global/hybrid/adapter mutation, consumer update, rollback/recovery, and GC serialize under one manager-home mutation lock. Consumer state commits last.
12. **Decision:** Meson and Node/TypeScript bundler configs/plugins join the prohibited recipe/plugin class; they are not aliases for direct compilers and receive no generic v1 driver.
13. **Decision:** v1 cache provenance is an explicit protected-state TCB assumption. Receipt and marker hashes provide consistency/currentness, not authentication. An unverified cache boundary forces rebuild, and same-principal/admin compromise is outside the install-time threat model.
14. **Decision:** the key, receipt bytes, artifact-relative path, and validation outcomes are portable; manager-home/cache directory names, receipt filenames, locks, quarantine paths, and storage backends remain implementation-specific.
15. **Decision:** schema 6 ships as protocol `1.0.0-rc.4` with `conformance-claim-v2`; claim v1 remains frozen as rc.3 evidence and cannot claim the new manager vectors.
16. **Anomaly:** both managers update consumer state before successful materialization, contrary to the documented phase order.
17. **Anomaly:** both managers accept an existing runtime directory without verifying required paths, contrary to the manager profile.
18. **Anomaly:** both managers’ per-node/per-target replacement can leave a partially changed installation after a later failure; build work requires a manager-home-isolated journal. The Python CLI has a coarse global lock but no journal/revalidation semantics; the Go manager lacks the shared lock.
19. **Probable regression:** the Python dry-run calls a registry path that appears to create/migrate persistent cache/state. Add a failing mutation test before changing it.

## 10. Clear v1 recommendation

Adopt protocol `1.0.0-rc.4`, manifest schema 6, build receipt schema 1, install marker schema 2, and conformance claim schema 2 with **only** the `go-v1` driver defined above. Require explicit context-excluded `build_roots`, a module rooted directly below one such path, the pre-cache `curator-build-source-v1` identity, the fixed active-Go-file directive scan, manager-protected cache-state checks with rebuild-on-untrusted fallback, and manager-home-isolated commit/rollback/GC in both managers before enabling v6 packages. Keep physical cache names implementation-specific, freeze claim v1 as rc.3 evidence, and retain the legacy marker-excluding hash only for installed-tree currentness. Reject all generic hooks, build executables, argument/environment arrays, package-manager/build-system/bundler frontends, cgo, PGO, `go:cgo_import_dynamic`, package-controlled assembly/host objects, external linking, libgcc lookup, downloads, workspaces, root-module build sources, and cross-compilation.

This is narrow enough to implement and test deterministically while preserving the protocol’s defining invariant: installation compiles untrusted Go source with a fixed trusted driver, but never transfers execution control to package-provided code.

## 11. Fact-check register

All external claims were checked against primary project/vendor documentation on 2026-07-20. No secondary comparison article is used as authority.

| Claim checked | Primary source | Finding used |
|---|---|---|
| Existing content-hash exclusion | [Protocol content hash](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/protocol/core.md#L282-L295), [Go manager hashing](https://github.com/relux-works/curator/blob/17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8/internal/hashing/hashing.go#L21-L29), [Python manager hashing](https://github.com/ivanopcode/cocoaskills/blob/6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12/src/csk/hashing.py#L16-L26) | Currentness hashes intentionally omit root `.csk-install.json`; they cannot bind compiler-visible bytes in a package-provided file at that path. |
| Current agent-context exclusion | [Protocol context selection](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/protocol/core.md#L73-L97), [Go whitelist](https://github.com/relux-works/curator/blob/17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8/internal/whitelist/whitelist.go#L24-L75), [Python whitelist](https://github.com/ivanopcode/cocoaskills/blob/6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12/src/csk/whitelist.py#L14-L91) | Eligible context roots are statically known and current implementations already accept excluded runtime-root subtrees; adding `build_roots` to the same pre-copy exclusion is deterministic without invoking a compiler. |
| Current cross-project isolation | [Manager concurrency profile](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/profiles/manager.md#L50-L82), [Python global lock](https://github.com/ivanopcode/cocoaskills/blob/6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12/src/csk/locking.py#L14-L47), [Go consumer update](https://github.com/relux-works/curator/blob/17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8/internal/scopes/consumers.go#L14-L61) | The profile permits independent-project concurrency without specifying shared-target isolation. Python currently serializes broadly; Go performs an unlocked consumer read-modify-write. A normative home commit/rollback/GC lock is needed for interoperable safety. |
| Machine-state portability boundary | [Core compatibility identifiers](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/protocol/core.md#L23-L34), [compatibility policy](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/COMPATIBILITY.md#L47-L50) | Machine-home directories/cache names/layouts are implementation-specific and managers do not share local state by default; only logical cache identity and receipt/artifact semantics should be portable. |
| Current claim/release identity | [Claim-v1 schema](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/schemas/v1/conformance-claim-v1.schema.json), [schema cases](https://github.com/relux-works/curator-spec/tree/57c1f56846d221ecc55786bd3c2467ec32f11730/conformance/v1/schema-cases/conformance-claim-v1), [generator version](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/tools/generate-vectors/main.go#L23), [validator version](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/tools/validate.py#L102-L103), [no in-place schema redefinition](https://github.com/relux-works/curator-spec/blob/57c1f56846d221ecc55786bd3c2467ec32f11730/COMPATIBILITY.md#L22-L24) | Claim v1 and the generated suite are pinned to rc.3. Changing that schema in place would change historical claim meaning; claim v2 plus an rc.4 suite is the explicit transition for schema 6. |
| Protected rooted filesystem access | [Go `os.OpenInRoot`](https://pkg.go.dev/os#OpenInRoot), [Microsoft access control](https://learn.microsoft.com/en-us/windows/win32/secauthz/access-control) | Current platforms expose containment and permission/ACL mechanisms, but their concrete paths and APIs differ; implementations must fail closed when they cannot establish the protected-state boundary. |
| Go embedded file inputs | [Go `embed` directives](https://pkg.go.dev/embed#hdr-Directives) | A directive names filesystem patterns relative to the package and initializes `string`/`[]byte`/`FS` from matched bytes at compile time; an explicitly named `.csk-install.json` is therefore a build input. |
| Go vendoring and network/module cache | [Go Modules Reference](https://go.dev/ref/mod#vendoring) | `-mod=vendor` loads dependencies from main-module `vendor/`, not network/local module cache, and validates `modules.txt` consistency. |
| Go automatic toolchain switching | [Go Toolchains](https://go.dev/doc/toolchain) | Auto mode may select/download a newer toolchain; `GOTOOLCHAIN=local` keeps the bundled/selected toolchain and fails compatibility instead. |
| Go external-link selection | [Go linker mode source](https://go.dev/src/cmd/link/internal/ld/config.go#L182-L227) | Auto mode consults `GO_EXTLINK_ENABLED` or the toolchain’s compiled default; explicit internal mode errors when external linking is required instead of silently changing mode. |
| Go libgcc lookup | [Go linker flags](https://go.dev/cmd/link/) | In internal mode, an unset `-libgcc` can be obtained by running the compiler; `-libgcc=none` disables that lookup. |
| Go package graph fields | [Go `list` documentation](https://go.dev/cmd/go/#hdr-List_packages_or_modules) | `-deps` emits dependencies in postorder and exposes `Standard`, `Goroot`, `SFiles`, `SysoFiles`, cgo/native-source fields, embedded files, and load errors needed by the closed-graph check. |
| Go dynamic-import directive | [Go compiler source](https://github.com/golang/go/blob/go1.25.5/src/cmd/compile/internal/noder/noder.go#L311-L332), [Go linker source](https://github.com/golang/go/blob/go1.25.5/src/cmd/link/internal/ld/go.go#L105-L136) | The compiler permits `go:cgo_import_dynamic` in ordinary non-standard source; `_ _ "library"` makes the internal linker add a selected dynamic library. It is not represented by go-list native file arrays, so active `GoFiles` need a separate rejection scan. |
| Go assembler includes | [Go assembler source](https://github.com/golang/go/blob/master/src/cmd/asm/internal/lex/input.go#L401-L424) | A package assembly `#include` can attempt a supplied name directly; rejecting non-standard active `SFiles` avoids outside-snapshot include inputs. |
| Go telemetry control | [Go telemetry](https://go.dev/doc/telemetry), [Go 1.23 release notes](https://go.dev/doc/go1.23#telemetry) | `GOTELEMETRY`/`GOTELEMETRYDIR` are read-only reports; `go telemetry off` disables local collection/upload, and the command is available in Go 1.23+. |
| Go telemetry directory | [`os.UserConfigDir`](https://pkg.go.dev/os#UserConfigDir) | Private `HOME`, `XDG_CONFIG_HOME`, or `APPDATA` values redirect telemetry configuration on Darwin, other Unix systems, or Windows respectively. |
| Go build flags and PGO | [Go command](https://go.dev/cmd/go/#hdr-Compile_packages_and_dependencies) | `-trimpath`, `-buildvcs`, `-buildmode`, `-compiler`, `-mod`, `-ldflags`, and `-o` are fixed; PGO otherwise defaults to package-local `default.pgo`, so v1 fixes `-pgo=off`. |
| Go generators | [Go generate](https://pkg.go.dev/cmd/go#hdr-Generate_Go_files_by_processing_source) | Generation is not automatically run by build/test; the manager never invokes it. |
| cgo subprocess surface | [cgo command](https://pkg.go.dev/cmd/cgo) | cgo can select/invoke C/C++ compiler and `pkg-config` behavior; disabling it removes that driver surface. |
| Rust build-time execution | [Cargo build scripts](https://doc.rust-lang.org/cargo/reference/build-scripts.html), [Rust procedural macros](https://doc.rust-lang.org/reference/procedural-macros.html) | Cargo runs compiled build scripts; proc macros execute during compilation. |
| Zig build/comptime behavior | [Zig build system](https://ziglang.org/learn/build-system/), [Zig reference](https://ziglang.org/documentation/master/#comptime) | `build.zig` controls build graph/tools; language code can be evaluated at compile time. |
| Swift package execution | [SwiftPM plugins](https://docs.swift.org/swiftpm/documentation/packagemanagerdocs/plugins/) | Package manifests run as Swift; build command plugins execute processes. |
| Make/CMake/Meson commands | [GNU Make](https://www.gnu.org/software/make/manual/html_node/Execution.html), [CMake process](https://cmake.org/cmake/help/latest/command/execute_process.html), [CMake custom command](https://cmake.org/cmake/help/latest/command/add_custom_command.html), [Meson setup command](https://mesonbuild.com/Reference-manual_functions_run_command.html), [Meson custom target](https://mesonbuild.com/Reference-manual_functions_custom_target.html) | Recipes, configure-time commands, and custom targets are package-selected execution points; Meson is not a compile-only exception. |
| JVM build/compile plugins | [javac](https://docs.oracle.com/en/java/javase/23/docs/specs/man/javac.html), [Maven](https://maven.apache.org/guides/introduction/introduction-to-plugins.html), [Gradle](https://docs.gradle.org/current/userguide/implementing_custom_tasks.html), [Kotlin](https://kotlinlang.org/docs/compiler-plugins-overview.html) | Annotation processors, plugins, and tasks execute code during compilation/build. |
| .NET build tasks/generators | [MSBuild tasks](https://learn.microsoft.com/en-us/visualstudio/msbuild/msbuild-tasks), [inline tasks](https://learn.microsoft.com/en-us/visualstudio/msbuild/msbuild-inline-tasks), [`Exec`](https://learn.microsoft.com/en-us/visualstudio/msbuild/exec-task), [Roslyn compiler platform](https://learn.microsoft.com/en-us/dotnet/csharp/roslyn-sdk/) | Project-controlled tasks/inline code/commands and source generators execute at build time. |
| npm lifecycle behavior | [npm scripts](https://docs.npmjs.com/cli/using-npm/scripts/) | Install lifecycle scripts and `node-gyp` behavior can execute package code. |
| Node/TypeScript bundler execution | [webpack configuration languages](https://webpack.js.org/configuration/configuration-languages/), [webpack loaders](https://webpack.js.org/concepts/loaders/), [webpack plugins](https://webpack.js.org/concepts/plugins/), [esbuild plugins](https://esbuild.github.io/plugins/) | Repository config can be executable JS/TS; webpack loaders run with Node capability and plugins receive compiler hooks; esbuild plugins inject JS/Go callbacks into the build. |
| Deno compile surface | [Deno compile](https://docs.deno.com/runtime/reference/cli/compile/), [Deno packages](https://docs.deno.com/runtime/packages/) | Compile supports network/package/config/script-affecting surfaces that need a separately pinned contract. |
| Python build backends | [pip build-system interface](https://pip.pypa.io/en/latest/reference/build-system/), [PyPA flow](https://packaging.python.org/en/latest/flow/) | Frontends invoke package-selected backend hooks and may install dynamic build dependencies. |

### 11.1 Rework verification

Task-scoped smoke checks used Go 1.25.5 on Darwin/arm64 and wrote only below `.temp/TASK-260720-poa3ze/`:

- `git ls-remote origin refs/heads/main` matched every inspected local `origin/main`: curator-spec `57c1f568…`, curator `17804cea…`, and cocoaskills `6fc2fd97…`;
- marker-embed snapshots A/B differed only in root `.csk-install.json`; the current protocol algorithm produced the same `content_sha256` (`sha256:829a040a…`), while independent length-framed `curator-build-source-v1` calculations produced `sha256:0017492c…` and `sha256:60fe9b76…`. The revised semantic vector rejects both root-module commands before cache-key construction;
- the research-only fixed Go build vector compiled those otherwise-invalid `source_dir: "."` fixtures without executing them and produced distinct artifact hashes `69f8a5d0…` and `7ddd293a…`, confirming that the omitted marker bytes are compiler-visible and motivating both the dedicated build-root rule and all-file identity;
- exact-ref regression gates were rerun after the cycle-4 changes: curator-spec validated 30 schemas and 93 vector files, ran 8 Python unit tests, and passed Go tool tests; Curator passed `go test ./...`; cocoaskills reported 488 passed and 18 skipped;
- with a fresh private home, `GOTELEMETRY=off go env GOTELEMETRY` still reported `local`; fixed `go telemetry off` followed by `go env` reported `off` and a directory below the private home;
- the exact `go list -mod=vendor -deps -json -buildvcs=false -compiler=gc -pgo=off .` shape exposed `SFiles:["escape_arm64.s"]` for an absolute-include fixture and `SysoFiles:["host_arm64.syso"]` for a host-object fixture, enabling rejection before build;
- a non-standard package containing `//go:cgo_import_dynamic _ _ "libReviewerProbe.dylib"` listed `GoFiles:["main.go"]` with empty `CgoFiles`, `SFiles`, and `SysoFiles`, yet the fixed internal-link build succeeded and `otool -L` showed `libReviewerProbe.dylib`; the new active-`GoFiles` byte scan rejects it before `go build`;
- the fixed build vector with `-ldflags=-linkmode=internal -libgcc=none` succeeded even when the diagnostic environment forced `GO_EXTLINK_ENABLED=1` and placed failing `cc`, `clang`, and `gcc` executables on `PATH`; none started. This demonstrates that the explicit link mode dominates an external-default signal and that the libgcc fallback is not used for the admitted pure-Go graph;
- independent Python 3.14.4 and Go 1.25.5 implementations produced `sha256:baf7c5f3b9c3f1fae3da4c356381bf74442aa7f8f0b6fb2304c9c10833d6032e` for the toolchain vector; Go and an independent framed-byte calculation agree on the build-source vectors; CCJ-1 calculations produced cache key `sha256:3fcd714a…` and receipt hash `sha256:750f5f75…` printed in section 5;
- all 24 fenced JSON documents parse. Independent recomputation confirms that the exact forged 24-byte artifact hashes to `sha256:a4f06a13…`, its self-consistent canonical receipt hashes to `sha256:9a23f5b7…`, and its input still derives the unchanged cache key `sha256:3fcd714a…`; the negative result therefore comes only from protected-state validation, as intended;
- exact-ref inspection confirmed that claim v1, its valid/invalid cases, the generator constant, generated manifest, validator, and conformance documentation are rc.3-pinned, while neither manager has a claim emitter. All nine newly cited protected-state/claim-transition primary URLs returned HTTP 200.
