# External build repositories: author and operator guide

This guide describes Curator's protocol 1.0.0-rc.5 support for repository-backed
CLI dependencies. The only supported external driver is the closed
`go-repository-v1` driver. An external Git repository is a source and audit
subject, not a generic build-script host.

## Ownership model

Three files have deliberately separate responsibilities:

| Owner | File | May select |
| --- | --- | --- |
| Skill author | `agent-skill.json` schema 7 | Repository identity and immutable lock; command key; repository key; logical target |
| Repository author | repository-root `skill-build.json` schema 1 | Closed driver revision; one explicit build root; one contained source directory |
| Operator | `Skillfile.dev.json` schema 2 and manager policy | Development source substitution; credentials and host verification; admitted Git and toolchain; audit, cache, and platform policy |

The command key is the executable identity. For command key `golden-tool`, the
manager derives the single artifact path as `bin/golden-tool` or
`bin/golden-tool.exe`. Repository metadata cannot choose binary names, output
paths, argv, environment, credentials, signing, hooks, plugins, generators,
fallbacks, or secondary artifacts. A package cannot delegate those choices to
a wrapper or arbitrary command. Receipts and install markers record the
manager's result; repository data cannot use them to make any of those choices.

The canonical manifest filename is `agent-skill.json`. Curator continues to
read an equal `csk-skill.json` compatibility copy, but authors should not ship
two different manifests: conflicting copies fail closed.

## Author a repository-backed command

The consuming skill declares a repository and selects a logical target. The
full lowercase object ID is the lock; `tag` is an optional additional
assertion.

<!-- example:manifest -->
```json
{
  "schema_version": 7,
  "capabilities": {},
  "build_repositories": {
    "golden-tools": {
      "git": "https://github.com/example/golden-tools.git",
      "locked_commit": {
        "object_format": "sha1",
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
  }
}
```

At the root of `golden-tools`, the repository author publishes the logical
target. This monorepo example selects the module under `tools/admin`; it does
not ask the manager to discover a module or executable.

<!-- example:descriptor -->
```json
{
  "schema_version": 1,
  "targets": {
    "golden-tool": {
      "driver": "go-repository-v1",
      "build_root": "tools/admin",
      "source_dir": "tools/admin/cmd/golden-tool"
    }
  }
}
```

`source_dir` must equal `build_root` or be contained below it. `build_root`
must contain `go.mod` directly, and that file must be the nearest ancestor
`go.mod` of `source_dir`. Select each nested module with its own logical target.
The whole admitted repository remains the identity and audit subject, while
only the selected build root is compiler-visible. External repository files do
not enter agent context or the consuming skill's runtime copy.

Use SHA-1 only with a 40-character object ID and SHA-256 only with a
64-character object ID. Do not use a branch, `HEAD`, abbreviated ID, revision
expression, or range in the declaration. For an untagged declaration the
manager acquires exactly the locked object. For a tagged declaration it
acquires exactly `refs/tags/<tag>`, peels the complete tag chain, and requires
the terminal commit to equal the lock. There is no direct-object fallback when
the tag is missing or moved.

Repository access is HTTPS or SSH under operator policy. TLS verification,
known-hosts state, authentication material, credential brokering, proxy policy,
timeouts, and the admitted Git/SSH binaries remain outside package data. A
repository that needs a package-provided credential helper or transport command
is unsupported.

## Development substitutions

Substitutions are non-committed operator input. They replace acquisition for
one declared repository; they do not change its repository key, target, driver,
command name, output, compiler policy, credentials, or signing policy.

A local worktree substitution is project-relative:

<!-- example:substitution -->
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

A network substitution uses the same restricted repository grammar and an
explicit ref. Branches are allowed here because this file is operator-owned
development state, never package provenance.

<!-- example:substitution -->
```json
{
  "schema_version": 2,
  "substitutions": {},
  "build_repository_substitutions": {
    "golden-skill": {
      "golden-tools": {
        "git": "ssh://git@github.com/example/golden-tools.git",
        "ref": {
          "kind": "branch",
          "value": "driver-development"
        }
      }
    }
  }
}
```

Local input must be an ordinary non-bare worktree with a direct, link-free
`.git` directory, files-format refs, admitted configuration, and complete
SHA-1 or SHA-256 loose objects or supported pack/index pairs. Gitfiles, linked
worktrees, bare repositories, reftable, alternates, replace refs, grafts,
shallow or promisor state, optional pack sidecars, Git LFS, links, and special
files fail closed. Curator parses Git administration and inertly copies object
bytes; it does not execute source hooks, filters, helpers, upload-pack, LFS, or
maintenance.

Receipts and markers retain both declared and effective identities. Strict
audit rejects substitutions. Advisory audit discloses the substitution and
audits the exact effective snapshot independently from the consuming skill.

## What an operator must provide

Before enabling `go-repository-v1`, establish all of the following:

- an admitted, fingerprinted Git release family and the fixed Go toolchain and
  sysroot covered by the rc.5 qualification evidence;
- HTTPS or SSH authentication and host-verification policy owned by the
  operator, never by the package;
- resource limits covering object count and size, expanded bytes, file count,
  path size, tree/tag depth, and time;
- an audit policy that treats the skill and each effective external repository
  as separate subjects; and
- owner-protected, contained, link-safe storage for frozen snapshots, receipts,
  artifacts, journals, markers, and command shims.

The required order is fixed: acquire the exact source, recompute raw-object
identities and prove the complete graph, scan every blob for Git LFS pointers,
materialize exact regular files, validate the whole snapshot, compute its
build-source digest, validate the descriptor and target, audit the external
subject, and only then look in the artifact cache. A cache miss may proceed to
the compiler. A claimed cache hit does not skip source proof or audit.

Cache keys and receipt hashes are consistency identifiers, not signatures or
trust claims. Curator recomputes a receipt cache key from its complete canonical
input and recomputes the marker receipt hash from the complete canonical
receipt. Unreadable or unprovable protected state is not current and is never
made trusted by changing permissions or adopting candidate bytes.

Command shims are manager-created PATH entries. They forward the user's
arguments, preserve the inherited PATH and child exit status, and resolve only
to the marker-selected protected artifact. They never point into a checkout,
Git object store, source snapshot, staging directory, or script runtime. Curator
does not execute the artifact during install, validation, status, repair,
rollback, or garbage collection.

## Project and global operator procedure

Start from an operator-controlled Curator configuration and a checked-in
`Skillfile.json`. The package declaration supplies the immutable repository
lock; operators must not replace it with a branch, a locally chosen object, or
an arbitrary build command.

```bash
# One-time machine configuration. Use an operator-owned directory containing
# the skill repositories named by Skillfile.json.
curator bootstrap --non-interactive --skills-root /absolute/operator/skills

# Project scope: validate the complete plan without writes, install it, and
# then require both installed and compiled state to be current.
curator install /absolute/project --dry-run --audit=strict --strict-tags
curator install /absolute/project --audit=strict --strict-tags
curator status /absolute/project --json --check
```

On Windows, pass an absolute Windows project path and skills root (for example
`C:\\Operator\\skills` and `C:\\src\\project`). Do not translate a Windows
path through a POSIX compatibility layer: repository and protected-store path
proof is native to the host.

The machine-wide scope has its own declaration and lifecycle:

```bash
curator global init
curator global install --dry-run --audit=strict --strict-tags
curator global install --audit=strict --strict-tags
curator global status --json --check
```

`status` is read-only. A non-current code is evidence to run the same declared
install again; install is Curator's repair operation and reacquires or rebuilds
only after source, audit, and protected-cache proof. Do not manually edit a
marker, receipt, cache entry, shim, snapshot, or its permissions. If an install
or repair fails, Curator retains the previously committed installation and
reports the rollback or recovery action. Re-run `status --check` after the
failure before deciding whether the old generation remains usable.

Uninstallation is declaration-first so dependency and shared-consumer state is
recomputed transactionally:

```bash
# Project scope.
curator remove golden-skill --project /absolute/project
curator install /absolute/project --audit=strict --strict-tags
curator status /absolute/project --json --check

# Global scope.
curator global remove golden-skill
curator global install --audit=strict --strict-tags
curator global status --json --check

# Reclaim only entries no live marker or journal still references.
curator gc
```

For interactive bare command names, cache the optional manager-generated shell
hook and follow the printed shell-specific activation instruction:

```bash
curator shell-init --install
```

Project and global shims remain directly invocable without modifying a shell
profile. PATH activation never points at a repository checkout or protected
cache and does not authorize a package-selected hook.

## Offline and read-only behavior

Offline outcomes are intentionally different:

| Operation | Exact protected evidence available | Required outcome |
| --- | --- | --- |
| Syntax-only validation | No | Warning `build_repository_unverified_offline`; no source, audit, cache, receipt, marker, or installation claim |
| Install/update/repair or coverage-claiming audit | No | Error `build_repository_source_unavailable` before cache lookup, compilation, or mutation |
| Untagged install with an exact protected snapshot | Yes | Revalidate the protected boundary and snapshot, repeat independent audit, then allow normal cache lookup/build processing |
| Tagged install or repair while remote tag proof is unavailable | Even if an old snapshot exists | Error `build_repository_source_unavailable`; a protected object cannot replace the operation's exact-tag assertion |
| Read-only status | Exact snapshot, receipt, artifact, marker, and shim relationship | Report current without contacting the remote merely to retest tag movement |
| Read-only status | Evidence missing or unreadable | Report non-current or currentness unknown; do not fetch, repair, adopt, sign, or execute |

The warning from syntax-only validation is not an installation failure, and it
is not proof that installation would succeed. Protected offline reuse is a
narrow manager-owned optimization for an exact untagged snapshot, not a
package-controlled fallback.

## Signing boundary

`go-repository-v1` performs no post-build signing, timestamping, or
notarization. A package signing request fails
`build_repository_package_signing_forbidden`. If the platform requires local
signing, the build fails `build_repository_signer_policy_unsupported` until a
separately reviewed operator signer profile defines the fixed signer, process
graph, network policy, identity handling, cache input, publication, and
rollback. Release-pipeline signing remains an operator concern; repository
metadata cannot request or select it.

## Troubleshooting

| Symptom or code | Meaning | Safe response |
| --- | --- | --- |
| Schema rejects `output`, `argv`, `env`, `credentials`, `signing`, hooks, plugins, generators, or fallback | Package data crossed a closed ownership boundary | Remove the field; do not move it into a script, wrapper, or alternate descriptor |
| `build_repository_ref_moved` | The exact tag no longer peels to the locked commit | Restore the tag or intentionally publish a reviewed manifest update with a new full lock |
| `build_repository_source_unavailable` | Exact source or required tag proof is unavailable | Restore operator network/access policy or the exact source; do not broaden the refspec or use an arbitrary command |
| `build_repository_unverified_offline` | Syntax is valid but source coverage was impossible | Treat it only as a warning from syntax validation; run a source-covering operation when evidence is available |
| `build_repository_incomplete_source` | The raw object graph or a bounded object is missing | Repair the source repository under operator control; do not hydrate content during admission |
| `build_repository_git_object_semantics_invalid` | A commit, tree, tag, mode, or object identity failed proof | Correct and republish the repository; do not bypass raw-object verification |
| `build_repository_git_lfs_unsupported` | An admitted blob is an LFS pointer | Commit complete ordinary Git blob bytes; LFS hydration is not supported |
| `build_repository_local_gitfile_unsupported` or `build_repository_local_linked_worktree_unsupported` | The local substitution is indirect or linked | Use a direct ordinary worktree clone |
| `build_repository_local_bare_unsupported` | The local substitution is bare | Use a non-bare ordinary worktree |
| `build_repository_local_layout_unsafe` or `build_repository_local_format_unsupported` | Links, special files, extensions, alternates, partial-clone state, or other administration are outside the admitted format | Create a clean ordinary worktree; do not patch Curator's protected state |
| `build_repository_audit_blocked` | The effective repository snapshot failed its independent policy | Resolve the audit record or policy decision; skill audit evidence cannot substitute for repository evidence |
| `build_repository_protected_boundary_untrusted` | Cache or snapshot storage cannot be proved owner-protected and link-safe | Rebuild into a new manager-created protected boundary; never adopt or chmod candidate bytes into trust |
| `build_repository_receipt_invalid` or `build_repository_artifact_invalid` | Cached evidence is corrupt or mismatched | Quarantine it and rebuild from an exact revalidated, audited snapshot |

## Future-driver admission

The external Git envelope is not a generic frontend and unsupported languages
are not equivalents to Go. Each language below remains unsupported until it
has a separately versioned closed driver, independent threat review, immutable
schema and vectors, and native qualification evidence. A future driver must
fix the entire process graph and compiler-visible input; it cannot expose a
package recipe or arbitrary command.

<!-- future-driver-table:start -->
| Language | Admission | Build scripts and plugins | Macros and generators | Dependencies and network | Native inputs | Deterministic artifact |
| --- | --- | --- | --- | --- | --- | --- |
| Rust | Unsupported; requires a separately versioned closed driver | Reject or fully model `build.rs`, Cargo aliases, config, and compiler/plugin loading | Bound procedural macros and generated source without executing repository-selected tools | Prove vendored crates, lockfile, features, registry identity, and offline Cargo behavior; deny undeclared fetches | Pin linker, sysroot, target specs, C toolchain, and native libraries | Define one manager-derived binary plus normalized metadata and reproducibility vectors |
| Swift | Unsupported; requires a separately versioned closed driver | Reject or fully model SwiftPM command/build-tool plugins and package scripts | Bound attached/freestanding macros and generated sources | Prove resolved packages and artifacts; deny undeclared registries, Git access, and downloads | Pin SDK, target triple, Clang importer/module maps, C-family libraries, and linker | Define one manager-derived executable with stable module/debug metadata handling and native vectors |
| Kotlin/JVM | Unsupported; requires a separately versioned closed driver | Reject Gradle/Maven scripts, tasks, init scripts, and plugins unless the driver fixes them completely | Bound KSP, kapt, annotation processors, compiler plugins, and generated classes/resources | Prove dependency graph, repositories, checksums, and offline cache; deny dynamic versions and fetches | Pin JDK, JVM target, JNI libraries, resource transforms, and platform tools | Define one manager-derived JAR or native artifact with normalized archives and reproducibility vectors |
| C/C++ | Unsupported; requires a separately versioned closed driver | Reject CMake/Meson/Autoconf recipes and package scripts unless replaced by a fixed manager graph | Bound preprocessor inputs, code generators, response files, and compiler/linker plugins | Prove headers, libraries, package-manager state, and offline dependency closure; deny configure-time network | Pin compiler, assembler, linker, sysroot, ABI, target, system headers/libraries, and runtime | Define one manager-derived artifact with normalized debug/build IDs and cross-platform vectors |
| .NET | Unsupported; requires a separately versioned closed driver | Reject or fully model MSBuild targets/tasks, SDK resolvers, props, and plugins | Bound source generators, analyzers, weaving, generated resources, and AOT steps | Prove NuGet lock graph, feeds, package hashes, and offline restore; deny undeclared downloads | Pin SDK/runtime packs, RID, native/AOT toolchain, P/Invoke and COM inputs, and platform libraries | Define one manager-derived assembly or native artifact with deterministic archive/metadata and native vectors |
<!-- future-driver-table:end -->

Every proposal must additionally define toolchain and sysroot identity, fixed
arguments and environment, offline link policy, audit-before-cache ordering,
receipt/marker/cache identity, signer boundary, dry-run semantics, status,
repair, rollback, garbage collection, resource limits, and platform-specific
failure vectors. It must reject produced-program execution during admission,
build validation, installation, or lifecycle operations.

See the [project overview](../README.md) for Curator's broader security and
interoperability model.
