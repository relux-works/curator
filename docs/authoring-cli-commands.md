# Authoring CLI Commands for Skills

Skill authors can deliver CLI utilities alongside skill instructions. Curator manages the build process, source isolation, runtime storage, and executable shims in project environments.

## Command placement options

You can place command source code in two distinct locations based on project structure and ownership.

### Embedded build roots

Embedded build roots store source code directly inside the skill package repository. You declare relative directory paths under the `build_roots` array in `agent-skill.json`.

Embedded placement fits single-repository skills where the CLI utility is developed and versioned together with the skill instructions.

### External build repositories

External build repositories store source code in a separate Git repository. You declare repository sources under the `build_repositories` map in `agent-skill.json`.

External placement fits multi-repository projects where CLI tools live in upstream repositories, are shared across multiple skills, or require separate versioning and access controls.

## Per-language admission matrix

Curator validates and builds CLI commands against explicit language driver contracts defined in `internal/godriver/` and `internal/skillspec/`.

### The `go-v1` and `go-repository-v1` compiled build drivers

Curator supports compiled builds through two closed driver identities: `go-v1` (for embedded build roots) and `go-repository-v1` (for external build repositories). They are currently the only compiled build drivers implemented in Curator.

```json
{
  "type": "build",
  "driver": "go-v1",
  "source_dir": "src/cmd/mytool"
}
```

The driver enforces strict compilation constraints before and during execution:

- **Vendored dependencies**: Builds enforce `-mod=vendor`. Source trees must include a valid `vendor/modules.txt` file when third-party modules exist. Network access is disabled during compilation.
- **Closed build parameters**: Argument vectors are hardcoded in `internal/godriver/build.go` (`listArguments`, `buildArgumentPrefix`). A build command in `agent-skill.json` admits only `type`, `driver`, `source_dir`, and, from schema 8, `modules`. Any other key, including `ldflags`, `gcflags`, environment overrides, or hook fields, is rejected during manifest parsing with `skill.manifest_invalid`: `commands.<name>.<field>: field is not supported for build commands` (`internal/skillspec/parse.go`, `rejectUnknownBuildFields`). The driver re-validates the presented command surface at the execution boundary and rejects any influence attempt with `build_execution_package_influence_forbidden` (`internal/godriver/build.go`, `validatePackageCommandSurface`); that check is defense in depth, because the parser has already refused the field.
- **Cgo prohibition**: Active cgo inputs are rejected with `cgo_required`. C, C++, Objective-C, Header, Fortran, and Swig source files in non-standard packages are rejected with `go_native_input_forbidden`.
- **PGO prohibition**: Profile-guided optimization files (`default.pgo`) are rejected with `go_pgo_forbidden`.
- **Code generator prohibition**: Source generator directives (`//go:generate`) outside third-party vendor directories are rejected with `go_generator_forbidden`. Compiler directives such as `//go:cgo_import_dynamic` are rejected with `go_forbidden_compiler_directive`.
- **Test file prohibition**: Test packages and test source files selected during build graph analysis are rejected with `go_test_input_forbidden`.
- **Assembly file constraints**: Non-standard assembly files (`.s`) outside third-party vendor directories or inside replaced modules are rejected with `go_assembly_forbidden`.
- **External linking prohibition**: Builds enforce internal linking (`-ldflags=-linkmode=internal -libgcc=none`). Linker modes requiring host external linkers are rejected with `external_link_forbidden`, and host linker fallbacks are rejected with `libgcc_fallback_forbidden`.
- **Toolchain switching prohibition**: Directives requesting another toolchain version in `go.mod` are rejected with `toolchain_switch_forbidden`.
- **Workspace prohibition**: The presence of `go.work` files inside the build root is rejected with `workspace_dependency_forbidden`.
- **Platform support**: Compilation is supported on macOS (`darwin`) and Windows (`windows`). On Linux (`linux`) and other unsupported platforms, Curator refuses execution before worker launch with `build_execution_control_unavailable` because native execution control records are defined only for macOS and Windows.

## Script commands

Script commands run interpreted source files directly without a compilation step.

```json
{
  "type": "script",
  "unix_path": "scripts/run.sh",
  "win_path": "scripts/run.cmd"
}
```

Curator resolves script commands into executable shims during installation:

- **Admitted interpreters**: Schema 8 script specifications admit `node-v1` and `python3-v1` in `internal/skillspec/types.go`. Every shell identifier (such as `bash` or `sh`) is deliberately absent from `ScriptInterpreters`; admitting a shell interpreter is a specification revision, not a manager configuration option.
- **Mutual requirement for execution policy and interpreter**: In schema 8, `execution_policy` and `interpreter` are mutually required (`internal/skillspec/parse.go`). Specifying one without the other triggers validation error `requires 'execution_policy'` or `requires 'interpreter'`.
- **Executable shim resolution**: `curator install` generates shims in `.agents/bin/`. On POSIX systems, when project path entries are present, shims use `#!/bin/sh` with `exec` to prepend path entries to `PATH` and replace the launcher process with the target executable. When no path entries are required, `curator install` creates a relative symlink instead (`runtimestore.WriteBinShim`). On Windows systems, shims generate `.cmd` batch scripts with double-expansion escaping.
- **Target resolution contract**: Executables sit in Curator's manager runtime store (`$HOME/.curator/runtime/<skill>/<commit>/...`) for script files, or in the local build cache (`$HOME/.curator/cache/build/<driver>/<hash>`) / external build artifact store (`$HOME/.curator/external-build-cache/artifacts/<sha256-hex>/artifact`) for compiled binaries.
- **Enforced script policy**: Schema 8 manifests permit `execution_policy: "script-worker-v1"`. Curator validates manifest syntax but `scriptpolicy.Admit()` refuses installation with diagnostic `script_execution_policy_unsupported`: "this manager does not implement script-worker-v1, and the policy forbids installing the command declared-only, downgrading it, or ignoring the field". Curator currently lacks a script worker process; per specification, a manager without script-worker support rejects the command rather than installing it declared-only.

## Worked examples

The following minimal examples demonstrate valid configurations for each supported command path.

### Example 1: Embedded `go-v1` skill

An embedded Go skill includes build source code inside its repository tree.

Create `agent-skill.json`:

```json
{
  "schema_version": 6,
  "capabilities": {},
  "build_roots": [
    "src"
  ],
  "commands": {
    "mytool": {
      "type": "build",
      "driver": "go-v1",
      "source_dir": "src/cmd/mytool"
    }
  }
}
```

The manifest declares `src` as a build root and `src/cmd/mytool` as the command source directory.

Directory layout:

```text
embedded_go/
├── SKILL.md
├── agent-skill.json
└── src/
    ├── go.mod
    ├── vendor/
    │   └── modules.txt
    └── cmd/
        └── mytool/
            └── main.go
```

The directory tree separates skill metadata from Go source code.

Execution target contract:

`curator install` creates an executable entry point shim in `.agents/bin/mytool` pointing to the compiled binary in Curator's local build cache (`$HOME/.curator/cache/build/go-v1/<hash>`). The launcher forwards command-line arguments to the content-addressed binary.

Run `curator skill check` to validate the skill package:

```bash
curator skill check ./embedded_go
```

The check verifies manifest fields, build root paths, and prompt-visible instructions.

### Example 2: External `go-repository-v1` skill

An external Go skill references a separate Git repository for build source code.

Create `agent-skill.json`:

```json
{
  "schema_version": 7,
  "capabilities": {},
  "build_repositories": {
    "remote_tool": {
      "git": "https://github.com/example/remote-tool.git",
      "locked_commit": {
        "object_format": "sha1",
        "hex": "da39a3ee5e6b4b0d3255bfef95601890afd80709"
      }
    }
  },
  "commands": {
    "remote_cmd": {
      "type": "build",
      "driver": "go-repository-v1",
      "repository": "remote_tool",
      "target": "remote_cmd"
    }
  }
}
```

Locked commits require exact SHA-1 or SHA-256 commit hashes. Including the `.git` suffix in HTTPS repository URLs is recommended operational advice to prevent HTTP 301 redirects, which Curator's protected fetch rejects (the parser itself trims `.git` for identity derivation in `internal/buildrepo/buildrepo.go`).

> [!NOTE]
> Operators configure credentials for private external build repositories using `curator config build-ssh` or `curator config build-https`. For details on authentication sources, transport rules, and HTTP redirect constraints, see [docs/build-ssh.md](build-ssh.md) and [docs/build-https.md](build-https.md).

Directory layout:

```text
external_go_repo/
├── SKILL.md
└── agent-skill.json
```

The skill package contains only manifest and instruction files; source code is fetched from the external repository.

Execution target contract:

`curator install` creates an executable launcher in `.agents/bin/remote_cmd` pointing to the compiled binary in Curator's external build repository artifact store (`$HOME/.curator/external-build-cache/artifacts/<sha256-hex>/artifact`).

Run `curator skill check` to validate the skill package:

```bash
curator skill check ./external_go_repo
```

The check confirms repository syntax, locked commit formatting, and target selections.

### Example 3: Script-command skill

A script-command skill exposes interpreted scripts directly.

Create `agent-skill.json`:

```json
{
  "schema_version": 2,
  "runtime_roots": [
    "scripts"
  ],
  "commands": {
    "run-script": {
      "type": "script",
      "unix_path": "scripts/run.sh",
      "win_path": "scripts/run.cmd"
    }
  }
}
```

The manifest lists `scripts` under `runtime_roots` and defines POSIX and Windows script entry points.

Directory layout:

```text
script_skill/
├── SKILL.md
├── agent-skill.json
└── scripts/
    ├── run.sh
    └── run.cmd
```

Script files sit inside declared runtime root directories.

Execution target contract:

`curator install` creates an executable launcher in `.agents/bin/run-script` pointing to the target script stored in Curator's runtime store (`$HOME/.curator/runtime/<skill>/<commit>/scripts/run.sh`).

Run `curator skill check` to validate the skill package:

```bash
curator skill check ./script_skill
```

The check verifies that script paths reside within declared runtime roots.

## Planned language drivers

Drivers for Kotlin (`kotlin-native-v1`, `kotlin-native-repository-v1`), Swift (`swift-v1`, `swift-repository-v1`), and Rust (`rust-v1`, `rust-repository-v1`) represent reserved driver identities in specification records, but are planned board work and not currently shipped or admitted by Curator. Reservation of a driver identity is explicitly not admission.

A manifest declaring any of these planned drivers today fails validation during manifest parsing in `internal/skillspec/parse.go`. For example, specifying `driver: "kotlin-native-v1"` yields:

```text
error: skill.manifest_invalid (agent-skill.json): commands.<name>.driver: must be 'go-v1' or 'go-repository-v1'
```

Curator rejects invalid driver declarations before starting build execution or fetching repository sources.
