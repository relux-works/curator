# External build repositories: architecture decision and threat model

**Task:** `TASK-260720-1nvomm`  
**Date:** 2026-07-27  
**Status:** Proposed architecture for independent review  
**Scope:** Architecture only. No protocol, schema, manager, test, CLI, or release file was changed.

## 1. Decision summary

External build repositories should be added as a new, explicitly versioned protocol surface. They must not be folded into manifest schema 6 or broaden `go-v1`.

The recommended transition is:

- manifest schema 7;
- a top-level `build_repositories` map in the skill manifest;
- a new closed command driver, `go-repository-v1`;
- one fixed repository-root descriptor, `curator-build.json`, schema 1;
- build receipt schema 2 and install marker schema 3;
- a new conformance claim and release identity, recommended as protocol `1.0.0-rc.5` with claim schema 3;
- schema 1 through 6, `go-v1`, build receipt v1, and marker v1/v2 remain readable with their existing meaning.

The consuming skill owns the command name. The external repository owns only a logical target definition: driver, `build_root`, and `source_dir`. It does not name the executable or select an output location. The manager continues to derive:

```text
artifact-relative path = bin/<manifest-command-key>[.exe]
staging path           = manager-private and implementation-specific
shim name              = <manifest-command-key>
```

This directly rejects the proposal that `curator-build.json` name the final binary or its output path. A repository-controlled executable name or path would reopen the output-selection boundary that schema 6 deliberately closed. Logical target IDs provide the requested target selection without transferring path or filename control.

An external source is first-class compiler input, not an ordinary runtime dependency. A manager must resolve, freeze, validate, hash, and audit its complete Git snapshot before build-cache lookup or compiler execution. A cache hit never bypasses external-source availability from either an exact protected snapshot cache or the locked Git source.

### Justified gap

**Missing piece:** the accepted schema-6 contract can compile only source stored inside the skill snapshot. It has no way to name, lock, validate, audit, or cache compiler source from a separate Git repository.

**Concrete requirement exposed by the scope delta:** Go CLI commands must be able to use a named external build repository now, while future languages must remain separate closed drivers. Repository access must be warning-only for a syntax-only offline check, hard-fail before mutation when an effective snapshot cannot be obtained for install/audit, and must preserve audit-before-compiler, no-hooks, manager-derived output, protected-cache, and rollback boundaries.

**Consequence if left open:** implementations would have to overload `go-v1`, treat an external checkout as an untracked host input, or accept package-selected Git/build/output behavior. Any of those would make cache keys incomplete, let cache hits bypass source audit, and weaken the accepted no-hooks and manager-derived-output MUST/MUST NOT rules.

**How this proposal closes it:** schema 7 adds an exact locked external-source declaration; `curator-build.json` exposes only logical targets; `go-repository-v1` is a new closed driver; receipt v2 and marker v3 bind external provenance; and the existing protected-cache and transaction model is extended without reinterpreting rc.4 artifacts.

**Self-verification performed before proposing follow-on work:**

- Checked accepted contract sections 2 (no-hooks threat model), 3 (manifest and `build_roots`), 4 (fixed `go-v1`), 5 (source/cache/receipt identity), 6 (ordering/dry-run/rollback), 7 (future ecosystems), 8 (artifact ownership), 9 (decisions/anomalies), and 10 (v1 recommendation).
- Checked the full rejected/out-of-scope boundary: generic hooks; package argv/environment/output; unsafe build systems; cgo/assembly/host objects/external linking; physical cache path standardization; receipt-only provenance; and unreviewed future drivers.
- Result: the accepted contract does not answer external Git compiler-source ownership, locking, descriptor shape, or signing separation, and does not exclude a new explicitly versioned closed driver. It does exclude broadening `go-v1` or adding package output/signing controls, so those alternatives are rejected here.
- No research task is required: the scope delta's architecture questions are resolved in this packet. Implementation work should be created only after independent architecture acceptance.

## 2. Normative manifest shape

### 2.1 Consuming skill manifest

Illustrative schema-7 manifest:

```json
{
  "schema_version": 7,
  "build_repositories": {
    "golden-tools": {
      "git": "https://github.com/example/golden-tools.git",
      "locked_commit": {
        "algorithm": "sha1",
        "hex": "0123456789abcdef0123456789abcdef01234567"
      },
      "tag": "v1.4.0"
    }
  },
  "commands": {
    "golden-tool": {
      "type": "build",
      "driver": "go-repository-v1",
      "repository": "golden-tools",
      "target": "golden-tool"
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

Normative recommendations:

1. `build_repositories` MUST be legal only in schema 7 or later and MUST be a strict map whose keys are portable identifiers.
2. Each entry MUST contain exactly `git` and `locked_commit`, plus OPTIONAL `tag`. Branches, ranges, abbreviated object IDs, symbolic revisions such as `HEAD`, and package-selected local paths MUST be rejected.
3. `locked_commit.algorithm` MUST identify the repository object format (`sha1` or `sha256` in the first version), and `hex` MUST be the full lowercase object ID of a commit. The manager MUST verify the repository storage algorithm using Git's object-format query and MUST peel the selected object to a commit.
4. When `tag` is present, the manager MUST resolve `refs/tags/<tag>^{commit}` and require exact equality with `locked_commit`. A mismatch is a hard `build_repository_ref_moved` failure in every mode that claims source verification; the general non-strict moved-tag warning does not weaken an explicit compiler-source lock.
5. Every repository declaration MUST be selected by at least one build command. Every `go-repository-v1` command MUST select exactly one declared repository and one descriptor target.
6. The command object MUST contain exactly `type`, `driver`, `repository`, and `target`. It MUST NOT contain a program, argv, environment, output, filename, toolchain, signing identity, hook, generator, plugin, build script, or fallback.
7. The command's `driver` MUST equal the selected descriptor target's `driver`. Unsupported or mismatched drivers fail before cache lookup or compilation.
8. Existing local `go-v1` commands and `build_roots` remain available in schema 7 with their schema-6 meaning. A repository command does not declare a local build root, and an external repository never becomes agent context or runtime content.

### 2.2 Why the lock is separate from the tag

A tag is a human-facing release label and can move. The full object ID is the immutable lock. Keeping both lets the manager report a meaningful moved-tag error without accepting the moved value.

A package update changes `locked_commit` explicitly. Install, repair, status, and audit never rewrite that lock. A future authoring command may propose a new lock, but applying it is a manifest mutation outside install.

## 3. Repository descriptor

### 3.1 Fixed name and schema

The repository root MUST contain exactly one descriptor named:

```text
curator-build.json
```

The first descriptor is schema 1:

```json
{
  "schema_version": 1,
  "targets": {
    "golden-tool": {
      "driver": "go-repository-v1",
      "build_root": ".",
      "source_dir": "cmd/golden-tool"
    },
    "admin-tool": {
      "driver": "go-repository-v1",
      "build_root": "tools/admin",
      "source_dir": "tools/admin/cmd/admin"
    }
  }
}
```

`curator-build.json` is strict JSON. Unknown top-level or target fields MUST fail. Target keys are logical identifiers only; they do not become filenames, paths, or PATH entries.

### 3.2 Target and monorepo rules

For `go-repository-v1`:

1. A target contains exactly `driver`, `build_root`, and `source_dir`.
2. `build_root` and `source_dir` use a repository-relative path type that permits the single value `"."` to denote the repository root and otherwise applies the protocol portable-path rules. This does not broaden the existing portable-path type globally.
3. `source_dir` MUST equal or be below the target's explicit `build_root`.
4. `build_root` MUST contain `go.mod` directly, and that file MUST be the nearest ancestor `go.mod` of `source_dir`. An intervening module fails.
5. Multiple targets MAY share the same build root. Nested build roots are permitted only when each target explicitly selects one root and satisfies the nearest-module rule. There is no automatic module discovery.
6. Every non-standard package, module file, active Go file, embedded input, and vendored dependency in the fixed `go list` graph MUST remain below the selected target's `build_root`. No package may import files from the consuming skill snapshot, another declared build repository, a sibling module, a parent workspace, a local module cache, or the host.
7. The complete external repository snapshot, including the descriptor and files outside a selected target root, participates in source validation, audit, and conservative source identity. Only the selected target root may be compiler-visible.
8. The descriptor MUST NOT contain command names, output basenames, output directories, install destinations, shim paths, PATH edits, signing policy, credentials, compiler flags, environment, or arbitrary build-system configuration.

Allowing `"."` here is safe and useful for normal root-module repositories because the complete external snapshot is compile-only and never enters agent context or runtime copying. The schema-6 prohibition on a local `"."` build root remains unchanged because its purpose is to create a static subtree exclusion inside an otherwise agent-visible skill snapshot.

## 4. Source access and immutable extraction

### 4.1 Access outcomes

The manager distinguishes three operations:

- **Syntax-only/offline check:** validates manifest and descriptor-independent field syntax. If neither an exact protected snapshot nor the locked Git object is available, it MUST warn `build_repository_unverified_offline` and MUST NOT claim source, descriptor, audit, or build validity.
- **Install, update, repair, or audit:** MUST obtain an exact frozen snapshot from a revalidated manager-protected snapshot cache, an operator substitution, or the declared Git source. If no exact snapshot is available, the operation MUST fail before cache lookup, compilation, signing, or installation mutation.
- **Dry-run:** MAY perform network reads and removable temporary fetches under the existing dry-run rules. It MUST validate and audit the exact snapshot before reporting a cache hit or would-build plan. It MUST leave no persistent checkout, response cache, snapshot, lock, marker, receipt, or other mutation.

Remote availability is not required when a manager already holds an exact, protected, fully revalidated snapshot for the declared canonical source and locked commit. “Inaccessible” means that no such effective snapshot can be proved.

### 4.2 Credentials and transports

1. Package manifests and repository descriptors MUST NOT contain usernames with secrets, passwords, tokens, credential-helper commands, `GIT_ASKPASS`, `SSH_ASKPASS`, `core.askPass`, `core.sshCommand`, `GIT_SSH_COMMAND`, proxy commands, remote helpers, or signing identities.
2. Network declarations MUST use `https` or `ssh` (including validated SCP-form SSH). Plain `http`, unauthenticated `git`, `ext`, arbitrary remote-helper schemes, and package-declared `file` sources MUST fail.
3. Local `path` or `file` access is permitted only through an explicit operator development substitution, never through package data.
4. Git protocol policy MUST default all protocols to `never` and enable only the selected `https` or `ssh` transport for the current fetch. URL rewriting MUST NOT introduce another protocol.
5. Authentication belongs to trusted operator/system configuration. A conforming manager SHOULD use a fixed credential broker or OS credential store selected independently of package input. It MUST NOT inherit a package repository's local Git config.
6. Because Git credential-helper strings can be shell-executed, a manager MUST NOT accept an arbitrary helper string from package data. If operator policy permits a helper, its executable and arguments are trusted manager/system configuration and part of the manager's source-fetch TCB.
7. Compiler children receive none of the fetch credentials, SSH agent sockets, Git environment, or network access. Source resolution and audit finish before compiler startup.

### 4.3 Extraction, submodules, LFS, links, and special entries

The manager SHOULD materialize snapshots from Git tree and blob objects using fixed plumbing equivalent to recursive NUL-delimited `git ls-tree` plus bounded `git cat-file`, without creating a checkout. This avoids checkout filters and worktree hooks. An implementation using another mechanism MUST prove identical raw-object, no-hook, no-filter behavior.

Normative rules:

1. The selected object MUST be a commit. The manager reads only its root tree.
2. Accepted tree modes are directories and regular blobs (`100644` or `100755`). Symlink mode `120000`, gitlink/submodule mode `160000`, and every unknown or special mode MUST fail before extraction or audit.
3. Submodule initialization, recursive fetch, and `.gitmodules` URL resolution are forbidden. A gitlink fails even if `.gitmodules` is absent or appears harmless.
4. Checkout clean/smudge/process filters, `export-subst`, repository hooks, `core.sshCommand`, and package repository configuration MUST NOT run or transform blob bytes.
5. Git LFS hydration is unsupported. The manager MUST NOT run `git-lfs`, a smudge filter, or an LFS network request. A canonical Git LFS pointer blob MUST fail conservatively as `unsupported_git_lfs`; the actual content is not present in the locked Git tree and therefore is not a valid compiler input.
6. Snapshot validation rejects invalid UTF-8 protocol paths, path escapes, case/platform collisions, duplicate encoded paths, links, special files, oversized entries, and mutations during use.
7. Fetch and object-database caches are manager-protected untrusted transport state. The frozen regular-file snapshot is independently validated and hashed before it can feed audit or build planning.

## 5. Audit and source identity

### 5.1 Two independently untrusted snapshots

A repository build has:

- the consuming skill snapshot, which declares the repository and command selection; and
- the external repository snapshot, which contains the descriptor and all possible compiler inputs.

Both MUST pass source policy, static validation, and audit before build-cache lookup or any compiler command. Registry evidence for the skill does not attest the external repository, and external repository evidence does not attest the skill. An audit result must name canonical source identity, locked commit, snapshot digest, descriptor target, and effective substitution state.

Audit unavailability is a hard failure for an install or an `audit` command that claims coverage. Existing advisory severity policy may still govern findings, but it cannot convert “external compiler source was not obtained or inspected” into success.

### 5.2 Build-source identity

Reuse the byte-exact `curator-build-source-v1` file-tree algorithm for the external repository snapshot: it is already domain-separated, length-framed, link-free, and covers every regular file. Its meaning remains “all regular files in one fully validated raw snapshot.”

The new logical input must distinguish the snapshot's origin and selection:

```json
{
  "schema_version": 2,
  "driver": "go-repository-v1",
  "source": {
    "kind": "git-repository",
    "repository": "golden-tools",
    "canonical_git": "github.com/example/golden-tools",
    "locked_commit": {
      "algorithm": "sha1",
      "hex": "0123456789abcdef0123456789abcdef01234567"
    },
    "build_source": {
      "algorithm": "curator-build-source-v1",
      "content_sha256": "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
    },
    "descriptor": {
      "path": "curator-build.json",
      "target": "golden-tool"
    }
  },
  "command": "golden-tool",
  "build_root": ".",
  "source_dir": "cmd/golden-tool",
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
    "telemetry": "off-private",
    "source_kind": "locked-external-git-v1"
  }
}
```

The exact input shape is receipt schema 2. Its CCJ-1 digest is the logical cache key. The canonical source identity and locked commit are included even when two repositories have byte-identical trees, so currentness, audit evidence, and revocation do not silently change provenance.

The compiler graph remains confined to `source.build_root`; the broader all-file source digest intentionally favors safe invalidation over maximum cache hits.

### 5.3 Receipt, marker, and protected caches

Build receipt v1 and marker v2 are strict and cannot gain external-source fields in place.

- **Receipt v2** records the schema-2 logical input above and retains exact canonical receipt bytes, manager-derived artifact path, hash, and size. It still contains no self-asserted provenance or trust boolean.
- **Marker v3** records, per build command, repository ID, canonical Git identity, declared tag if any, locked commit with algorithm, descriptor target, external `build_source`, cache key, receipt hash, artifact hash, and manager-derived artifact path.
- Marker v3 also retains the skill's own source/commit/content/currentness fields. Status verifies both snapshots and the exact relationship between manifest command, repository lock, descriptor target, receipt, and protected artifact.
- Marker v1 remains historical for schema 1 through 5. Marker v2 remains valid for schema-6 in-skill builds. New writers use v3 after any schema-7 mutation; readers dispatch by marker schema and never infer v3 from fields.

Repository checkout/object caches, frozen snapshot caches, receipts, and compiled artifacts have distinct roles. Only the immutable snapshot and compiled-artifact caches need currentness references from marker v3 and transaction journals. Receipt hashes remain consistency identifiers, not signatures. Every persistent cache boundary retains the manager-created, owner-protected, link-safe, revalidate-before-reuse rules. An untrusted candidate snapshot or artifact is never adopted by fixing permissions.

## 6. Build, installation, PATH, and rollback

`go-repository-v1` reuses the accepted `go-v1` compiler policy and five fixed Go argv forms, with the external target's `source_dir` as CWD and its `build_root` as the complete non-standard dependency boundary. The different source acquisition, command schema, input object, receipt interpretation, and marker semantics justify a new driver identifier even though the Go compiler arguments remain closed.

Normative sequence:

1. Parse and validate the consuming manifest, repository declaration, locked commit syntax, and command collisions.
2. Resolve or load the exact protected external snapshot; validate raw Git tree modes and paths; parse `curator-build.json`; select one target; verify driver equality and module containment.
3. Compute external `curator-build-source-v1`.
4. Audit the consuming skill and external repository independently. Apply allowed-source, revocation, registry, and moved-lock gates.
5. Probe and fingerprint the operator-trusted Go toolchain; derive the complete receipt-v2 input and logical cache key.
6. Validate a protected cache hit, or on a real miss run the fixed `go list` graph checks and fixed `go build` into operation-private staging. Never execute the artifact.
7. Verify artifact type, size, hash, and manager-derived path. Generate canonical receipt v2 and marker v3.
8. Under the existing manager-home mutation lock, recover/revalidate, atomically publish the immutable entry, and commit contexts, marker, runtime/shim/environment targets, mirrors/adapters, removals, and consumer ledger last under one journal.
9. Roll back in reverse order while retaining the home lock. A source, audit, preflight, or build failure before publication leaves installation and live caches unchanged.

PATH and shim verification is structural, not executable:

- the command key passes existing portable-name and collision rules;
- the generated shim is manager-owned and resolves exactly to the protected artifact selected by marker v3;
- the target is one regular singly linked executable file with the receipt hash and path;
- no shim points into a Git checkout, external snapshot, staging directory, or commit-keyed script runtime;
- the project/global environment exposes only the manager-generated bin directory under existing rules;
- status/repair validates these relationships without invoking a shell, `command -v`, or the built program.

The repository descriptor never writes PATH. It cannot request aliases, secondary binaries, sidecar installation, or post-build copying. A repository needing several commands declares several logical targets, and the consuming skill maps each one to a separate manifest command key.

## 7. Platform signing and notarization boundary

Signing identities and notarization credentials are platform/operator secrets. They MUST NOT appear in a skill manifest, `curator-build.json`, repository source, environment inherited by the compiler, receipt trust boolean, or marker.

The first `go-repository-v1` release performs no manager post-signing. Any code-signature metadata emitted by the fingerprinted Go toolchain is part of the ordinary artifact bytes and toolchain identity. The manager does not invoke `codesign`, `signtool`, a timestamp service, or a notary service.

If a platform deployment requires local post-build signing, it needs a separate reviewed signer profile or driver revision with all of the following:

- selection only from locked system/manager configuration;
- an opaque OS keychain, certificate-store, HSM, or trusted-service identity, never raw package credentials;
- fixed signer executable, argv, entitlements/options, network policy, and process graph;
- signed bytes, signer-policy revision, and signer/tool identity included in cache and receipt identity;
- no package-provided entitlements, certificate selector, timestamp URL, description, or signing arguments;
- private staging before publication and the same no-execution, rollback, and protected-cache rules.

Until that profile exists, a system policy that requires local signing MUST reject source-built artifacts rather than improvise a signing command.

Release-time distribution signing and notarization are separate from local installation:

- Apple Developer ID signing, hardened-runtime configuration, timestamping, notary upload, and ticket stapling belong in the producer's release pipeline. Apple notarization is a network service with account credentials and produces/staples a ticket, so it cannot be a deterministic offline install step.
- Windows Authenticode signing and optional timestamping similarly use a certificate/private-key store and may contact a timestamp server; they belong in a controlled release pipeline unless a separate local signer profile is standardized.
- A future prebuilt-artifact driver may consume signed/notarized release artifacts, but it must verify an exact artifact digest and an independently defined signature/attestation policy. It must not conflate release provenance with the local source-build receipt.

Linux support remains a separate native-validation task. This architecture adds no synthetic Linux signing requirement and makes no Linux conformance claim until the closed driver, sandbox/resource controls, shim behavior, and vectors pass on Linux.

## 8. Future language-driver extension rule

The external Git envelope may be shared, but compile semantics are never generic. Each ecosystem requires a new target driver ID and independent protocol/security review. The manifest still selects one repository, logical target, and exact closed driver; the descriptor still cannot supply argv, environment, output, hooks, plugins, or signing.

Minimum review requirements remain those in decision 0004 plus external-source rules from this packet:

- full compiler-visible input and dependency containment;
- trusted toolchain and sysroot/runtime identity;
- fixed process graph, environment, flags, target, output, and signer boundary;
- no package-controlled hook, plugin, macro, generator, annotation processor, build task, recipe, response file, linker, or produced-program execution;
- offline dependency model and exact source identity;
- receipt/marker/cache identity and corruption/provenance semantics;
- dry-run, audit-before-build, rollback/recovery, and cross-platform vectors.

Specific consequences:

- **Rust:** Cargo remains rejected because it compiles and runs `build.rs`, and procedural macros execute at compile time. A future direct `rustc` driver must define a closed dependency/linking model and reject both surfaces.
- **Swift:** SwiftPM manifests/plugins/macros remain outside the boundary. A direct `swiftc` driver must own SDK, target, dependency graph, linker, and macro policy.
- **Kotlin/JVM:** Maven and Gradle remain prohibited. A direct JVM driver must disable annotation processing and compiler plugins, own the class/module path and packaging, and fingerprint the JDK/Kotlin toolchain and runtime model.
- **C/C++:** Make/CMake/Meson/Ninja remain prohibited. A direct compiler driver must own source enumeration, preprocessing inputs, headers/sysroot, response-file grammar, compiler/linker plugins, native libraries, and output.
- **.NET:** MSBuild remains prohibited. A direct compiler driver must disable analyzers/source generators, own references/runtime packaging, and define native-host and signing behavior.

Adding an external repository does not make an unsafe build frontend safe.

## 9. Compatibility and migration

1. Schema 1 through 6 behavior is unchanged. Schema 6 rejects `build_repositories` and `go-repository-v1`.
2. `go-v1` remains an in-skill, context-excluded build-root driver and is not reinterpreted.
3. `curator-build.json` is ignored in repositories that are not selected as external build repositories; it is never an alternate skill manifest.
4. Receipt v1 and marker v2 remain frozen. Receipt v2 and marker v3 are explicit reader/writer transitions.
5. Conformance claim v2 remains rc.4 evidence. The external-repository release needs claim v3; an rc.4 result cannot claim the new behavior.
6. An existing schema-6 skill may migrate source to an external repository only by moving to schema 7, adding an exact repository lock, adding the descriptor to the external commit, switching the command to `go-repository-v1`, and reinstalling to marker v3. There is no transparent migration or fallback.
7. Physical checkout, snapshot, receipt, artifact-cache, lock, credential, quarantine, and signing paths remain implementation-specific.

## 10. Local operator substitutions

Extend `Skillfile.dev.json` with a new schema rather than overloading package data. Recommended strict shape:

```json
{
  "schema_version": 2,
  "substitutions": {},
  "build_repository_substitutions": {
    "golden-skill": {
      "golden-tools": {
        "path": "../golden-tools"
      }
    }
  }
}
```

The outer key is the consuming skill name and the inner key is its repository ID. An entry is exactly one local Git checkout path or Git plus a development ref, matching existing operator substitution concepts.

Rules:

- the file is operator/project-owned, uncommitted, and ignored;
- a local path is resolved relative to the project, must be a Git repository, and snapshots its exact HEAD commit through the same raw-object extraction; dirty or untracked worktree bytes are not compiler inputs;
- development Git refs may include branches under existing substitution rules, but the effective full commit and source identity are recorded;
- the substitution replaces the declared source only for the named consuming skill/repository pair and never changes target/driver/output semantics;
- marker v3 records that the source was substituted plus its effective canonical/local identity, commit, and source digest;
- strict audit fails and advisory audit warns, as for current development substitutions;
- the compiler receives no additional credentials or environment;
- a substitution is an explicit operator override, so availability is evaluated against the effective substituted source rather than the original remote.

## 11. Threat model

### 11.1 Adversary-controlled inputs

- consuming skill snapshot and manifest;
- external Git URL, repository bytes, refs/tags presented by the server, object database, descriptor, paths, file modes, source, vendor tree, compiler directives, embedded files, sizes, and diagnostics;
- candidate source, receipt, marker, and artifact cache bytes until protected-state admission;
- compiled artifact;
- network failure, replay, tag movement, and malicious repository configuration.

### 11.2 Trusted computing base

- manager and protocol implementation;
- operator-selected Git executable/source-fetch policy and credential broker;
- operating system, filesystem containment, locks, protected caches, and transaction journal;
- operator-trusted Go toolchain and its fingerprint;
- canonical JSON and hash implementations;
- any future separately standardized platform signer and protected signing identity.

### 11.3 Principal threats and controls

| Threat | Required control |
|---|---|
| Mutable tag or symbolic revision changes compiler source | Required full `locked_commit`; optional tag must peel to it; mismatch hard-fails |
| Package selects local source or credentials | Local paths only in operator `Skillfile.dev.json`; credentials only from trusted system/operator configuration |
| Custom Git protocol or remote helper executes code | Only `https`/`ssh`; default protocol deny; no package Git config or URL rewrite |
| Credential helper/askpass shell execution | No package helper fields or inherited helper environment; fixed trusted broker only |
| Checkout filter, hook, or LFS smudge executes code/network | Raw tree/blob extraction; no checkout/filter/hook; LFS pointers rejected |
| Submodule introduces an unbound repository | Reject gitlinks; never initialize or fetch submodules |
| Symlink, special file, path escape, or collision | Accept only tree and regular-blob modes; protocol path and platform collision validation |
| Descriptor chooses output, flags, signing, or build system | Strict target shape; manager command key/output; new closed driver only |
| Go graph reads outside selected module or from host | Fixed vendor-only `go list`; every non-standard input below explicit target build root |
| External source bypasses audit on cache hit | Resolve protected snapshot, hash, and audit before compiled-cache lookup |
| Forged source/artifact cache | Revalidate manager-protected boundary and exact hashes; rebuild rather than adopt |
| Compiler exploit or resource exhaustion | Existing read-only source, network denial, bounded diagnostics, time/memory/disk/process limits |
| Signing credential exfiltration | No install-time signing in v1; future signer uses operator-owned opaque identity and fixed policy |
| Shim/PATH hijack or validation executes artifact | Structural no-follow/hash validation; manager-generated shim; never execute artifact |
| Partial install or cross-project lost update | Existing private build then one locked journaled manager-home transaction; consumer last; reverse rollback |

Same-principal/admin compromise remains outside the local protected-cache invariant, exactly as in decision 0004. External repositories do not strengthen receipt hashes into cryptographic provenance.

## 12. Required project impact

### 12.1 curator-spec

This should be a new schema-7 story, not an expansion of the accepted schema-6/rc.4 story.

Smallest development-ready decomposition after architecture acceptance:

1. **Core/security/decision contract** — schema-7 source model, Git lock/extraction boundary, descriptor semantics, audit equivalence, signing boundary, and compatibility. Traces to sections 1-11 of this packet.
2. **Wire schemas** — manifest v7 canonical/legacy pair, `curator-build-repository-v1`, receipt v2, marker v3, Skillfile.dev v2, claim v3, strict shared definitions and generated cases. Traces to sections 2, 3, 5, 9, and 10.
3. **Manager profile and CLI** — access outcomes, protected snapshot lifecycle, credentials/transports, dry-run/status/repair/GC, substitutions, structural PATH/shim reporting. Traces to sections 4, 6, 7, and 10.
4. **Conformance and release gates** — raw Git tree vectors, lock movement, descriptor/monorepo vectors, audit/cache/build lifecycle, signing non-inheritance, rc.5 claim transition, docs and release metadata. Traces to sections 4-9 and 11.

Dependencies: 1 precedes 2 and 3; 2 and 3 precede 4. Do not split separate ceremonial documentation or test tasks; each task owns its local tests/validation.

### 12.2 Curator (Go)

Affected implementation surfaces at the inspected current checkout include:

- `internal/skillspec`: schema-7 manifest, repository declarations, external command parsing, target-driver matching;
- `internal/gitops`, `internal/snapshot`, and `internal/identity`: locked object format/commit resolution, fixed transport/config policy, raw tree/blob extraction, source identity, protected external snapshots;
- `internal/devsub`: nested operator build-repository substitutions;
- `internal/audit` and `internal/registry`: second audit subject per external repository and provenance-aware findings/attestation behavior;
- `internal/closure` and `internal/install`: repository planning, deduplication, audit-before-cache/build, build ordering, restart/revalidation, transaction ownership;
- new closed build/toolchain/cache packages rather than adding arbitrary execution to `internal/runtimestore`;
- `internal/marker`, `internal/scopes`, `internal/runtimestore`, `internal/envfiles`, and `internal/adapters`: marker v3, snapshot/artifact GC roots, immutable artifact shims, status/repair and structural PATH checks;
- `cmd/curator` and interoperability tests: syntax-only warnings, dry-run plans, audit/install failures, JSON status, and exact vectors.

The Go implementation must not reuse the existing script runtime `<skill>/<commit>` path for artifacts or use a checkout-based path that can run filters.

### 12.3 cocoaskills (Python)

At inspected `origin/main` commit `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`, affected surfaces include:

- `src/csk/skillspec.py`, `manifest.py`: schema-7 and command/repository model;
- `git_ops.py`, `snapshot.py`, `source_identity.py`: locked object format/commit, transport/config policy, raw object extraction, protected external snapshot identity;
- `dev_substitutions.py`: nested build-repository overrides;
- `closure.py`, `installer.py`, `global_install.py`: external-source planning, independent audit, cache/build sequencing, journaled commit;
- `audit/`, `audit_registry.py`, `attest.py`: external repository audit subject and registry distinction;
- `hashing.py`, `protocol_json.py`, `skillspec.py`: build-source, receipt-v2, and descriptor canonical semantics;
- `shims.py`, `env_files.py`, `status.py`, `consumers.py`, `gc.py`, `locking.py`: marker v3, protected artifact shims, structural status, snapshot/artifact reachability, manager-home isolation;
- `cli.py` and tests: static warning versus install/audit hard failure, dry-run plans, substitution reporting, and interoperability fixtures.

The current Python snapshot path uses `git archive`; the external-repository implementation must either replace it with fixed raw-object extraction for this source kind or prove that its exact invocation cannot apply package-controlled transforms. Its dry-run and persistent registry/cache behavior must continue to satisfy the already accepted no-mutation requirements.

### 12.4 Interoperability and platform validation

After both clients implement the same released vectors:

1. Cross-language fixtures must prove identical manifest/descriptor parsing, Git object-format lock handling, snapshot bytes, build-source hash, cache input/key, receipt bytes/hash, marker v3, and status results.
2. Lifecycle fixtures must cover unavailable source with and without protected snapshot, moved tag, SHA-1/SHA-256 object formats, malicious protocol rewrite, helper/filter non-execution, submodule/LFS/link rejection, monorepo targets, cache hit/miss/corruption, build failure, publisher race, rollback, repair, and locked GC.
3. Native integration must prove the same artifact and shim relations without executing the artifact during manager activity.
4. macOS and Windows signing tests initially assert that package signing fields are rejected and signing/notary credentials are absent from compiler children. They do not perform notarization.
5. Linux remains a separate later native validation and cannot be listed in claim-v3 operating systems until it passes the same manager vectors and host resource/sandbox assertions.

Recommended manager work items after the spec story:

- one Curator implementation task;
- one cocoaskills implementation task;
- one cross-client interoperability task after both;
- one later Linux native-validation task, explicitly downstream and not a blocker for architecture review.

## 13. Rejected alternatives

- **Add `git`, `ref`, and output fields directly to a build command:** duplicates source declarations, permits ambiguous sharing, and reopens package output control.
- **Let `curator-build.json` name the binary or output path:** violates the accepted manager-derived output invariant and creates path/collision/escape surface.
- **Reuse `go-v1`:** external acquisition and source identity change the command and receipt contract; silent reinterpretation would invalidate rc.4 evidence.
- **Use tag-only or branch references:** mutable compiler input and irreproducible first install.
- **Allow Git LFS or submodules:** introduces extra repositories, credentials, network operations, filters, and identities that are not in the lock or descriptor.
- **Trust a cache hit without obtaining the external snapshot:** bypasses source validation and audit.
- **Permit package credential helpers or signing identities:** both can execute helpers or expose secrets under package control.
- **Run checkout/build-system configuration from the repository:** violates no-hooks and can execute filters, plugins, recipes, or generators.
- **Standardize physical checkout/cache paths:** machine-local layout remains an implementation boundary.
- **Perform Developer ID/Authenticode signing or notarization during the first driver:** adds secrets, mutable external services, new executable children, nondeterministic timestamps, and post-processing not covered by `go-repository-v1`.

## 14. Source register

Primary sources checked on 2026-07-27:

- Git protocol policy: https://git-scm.com/docs/git-config#Documentation/git-config.txt-protocolallow
- Git credential and askpass/helper execution: https://git-scm.com/docs/gitcredentials
- Git checkout filters and clean/smudge/process execution: https://git-scm.com/docs/gitattributes
- Git tree enumeration: https://git-scm.com/docs/git-ls-tree
- Git raw object reading and symlink behavior: https://git-scm.com/docs/git-cat-file
- Git repository object-format query: https://git-scm.com/docs/git-rev-parse
- Git revision and commit peeling semantics: https://git-scm.com/docs/gitrevisions
- Git submodule descriptor semantics: https://git-scm.com/docs/gitmodules
- Git LFS stores pointer files in Git and content externally: https://git-lfs.com/
- Go vendoring and `vendor/modules.txt` consistency: https://go.dev/ref/mod#vendoring
- Go fixed build flags including `-trimpath`, `-buildvcs`, and `-pgo`: https://go.dev/cmd/go/#hdr-Compile_packages_and_dependencies
- Go local toolchain selection: https://go.dev/doc/toolchain
- Apple Developer ID signing and notarization requirements: https://developer.apple.com/documentation/security/notarizing-macos-software-before-distribution
- Apple custom notarization workflow and network submission: https://developer.apple.com/documentation/security/customizing-the-notarization-workflow
- Microsoft SignTool certificate/private-key and timestamp surfaces: https://learn.microsoft.com/en-us/windows/win32/seccrypto/signtool
- Cargo build-script execution: https://doc.rust-lang.org/cargo/reference/build-scripts.html
- `javac` annotation-processing controls: https://docs.oracle.com/en/java/javase/23/docs/specs/man/javac.html
- Roslyn source generators as compile-time metaprogramming: https://learn.microsoft.com/en-us/dotnet/csharp/roslyn-sdk/#source-generators

## 15. Review questions

Independent review should answer only these closed checks:

1. Does schema 7 plus `go-repository-v1` preserve every schema-6/`go-v1` meaning?
2. Is required `locked_commit` sufficient and unambiguous for both SHA-1 and SHA-256 Git repositories?
3. Does raw tree/blob extraction eliminate package-controlled checkout/filter/submodule/LFS execution paths?
4. Can any manifest or descriptor field influence executable name, output location, argv, environment, credentials, signing, or process graph?
5. Are the skill and external repository both validated and audited before cache lookup and compilation on real, cache-hit, and dry-run paths?
6. Do receipt v2 and marker v3 bind external source provenance without treating hashes as authentication?
7. Are PATH/shim currentness and rollback verifiable without executing built output?
8. Is signing correctly separated into a future operator-owned closed profile and release-time notarization pipeline?
