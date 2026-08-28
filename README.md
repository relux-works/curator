# Curator

Curator is an agent environment manager (AEM). The software manages what an AI coding agent receives in a project environment: skill packages, transitive dependencies, executable command shims, MCP server requirements, per-agent delivery adapters, and security verification gates. Declarative configuration produces reproducible installations.

Curator implements the open [Curator Specification](https://github.com/relux-works/curator-spec). The specification defines skill packages, project manifests, installation semantics, and audit registry interactions; sections are cited across this repository as `Spec §N.M`.

## Status

Version v0.1 development is complete: all twelve phases in [docs/implementation-plan.md](docs/implementation-plan.md) are done. Continuous integration validates schemas and conformance vectors from `curator-spec` on macOS, Linux, and Windows. Work tracking lives in [.task-board/](.task-board/).

## Install

<details open>
<summary>Homebrew (macOS, Linux)</summary>

```bash
brew install relux-works/tap/curator
```

</details>

<details>
<summary>Installer script (macOS, Linux)</summary>

```bash
curl -fsSL https://raw.githubusercontent.com/relux-works/curator/main/install.sh | sh
```

</details>

<details>
<summary>Scoop (Windows)</summary>

```bash
scoop bucket add relux-works https://github.com/relux-works/scoop-bucket
scoop install curator
```

</details>

<details>
<summary>Go toolchain</summary>

```bash
go install github.com/relux-works/curator/cmd/curator@latest
```

</details>

Debian and RPM packages ship with every release alongside SBOMs and cosign signatures. macOS binaries are Developer ID signed (Relux Works, LLC). Run the verification command to check any downloaded artifact:

```bash
gh attestation verify <artifact> --owner relux-works
```

## What Curator manages

Curator controls environment dependencies and runtime delivery:

- **Skill packages**: `SKILL.md` plus context directories, with an implementation-neutral machine manifest (`agent-skill.json`, schemas 1 through 8) declaring commands, runtime layout, capabilities, and dependencies. The legacy `csk-skill.json` filename remains readable. See [Authoring CLI commands](docs/authoring-cli-commands.md) for command declarations, drivers, and worked examples.
- **Project manifests**: `Skillfile.json` with exact git references; non-committed development substitutions.
- **Resolution**: Transitive dependency closures unified to one commit and one source identity per name, with activation modes.
- **Installation**: Context and runtime separation, install markers with content hashes, a commit-keyed runtime store, command shims, managed per-agent adapters.
- **Scopes**: Project, global, and hybrid (machine-stored, per-project activation).
- **MCP requirements**: Read-only verification of declared MCP servers against agent configuration surfaces.
- **Security**: Source allowlists, declared capability boundaries, zero code execution during install, and an audit registry client (Ed25519 signed records, deny-wins federation, snapshot verification).
- **Operator credentials**: Per-repository SSH selection and scoped HTTPS token sources for external build repositories, matched by canonical source identity and never selectable by a package. Private HTTPS fetches use a manager-owned, host-pinned askpass broker; public HTTPS can remain anonymous. See [SSH credentials](docs/build-ssh.md) and [HTTPS credentials](docs/build-https.md).

## Registry client guarantees

Curator binds persisted rollback and equivocation state to the canonical registry URL, so signing-key rotation never resets the highest accepted snapshot. This durable state lives under the Curator home `state/registry` directory, outside the disposable `cache/registry` responses; upgrades migrate legacy state without lowering it, and corruption or write failure is fail-closed. A protected catalog distinguishes first use from deletion of a previously accepted registry state. Record pagination rejects repeated or oversized cursors, more than 10,000 records per artifact query, and responses larger than 16 MiB.

Registry requests use bounded per-attempt and total deadlines. GET requests retry network failures, `429`, and `503` at most twice after the first attempt. Publication retries the exact body only with its deterministic `Idempotency-Key`; other client errors and unsafe requests are never retried. Redirects are rejected so a registry cannot move a request or bearer token to another endpoint.

## Commands without profile setup

Shell profile changes are not required. After `curator install`, agents can invoke project commands through `.agents/bin/<command>` on Unix and `.agents\bin\<command>.cmd` on Windows. Global installation publishes non-destructive forwarding shims to a safe user directory already on `PATH` when one is available; otherwise Curator reports the canonical global bin location. Installed launchers carry the project command directory and resolved system dependency directories themselves, preserve the inherited `PATH`, and return the child command status.

Run bootstrap and upgrade to set up a repository:

```bash
curator bootstrap --if-missing --non-interactive --skills-root "$HOME/src/skills"
curator upgrade .
```

`curator upgrade` fetches dependency closures for the selected project. `curator upgrade --dry-run` plans updates without mutating disk state.

To enable shell integration, cache the optional hook once:

```bash
curator shell-init --install
```

Automatic detection selects zsh or bash from `SHELL`, preserves Git Bash on Windows, and otherwise selects PowerShell on Windows. The cached hook does not start Curator during later shell launches. Curator never edits a profile automatically. Set `CURATOR_AUTO_ENV=0` to retain global activation while disabling project-directory scans.

## Execution assurance

Compiled builds default to portable execution mode using an authenticated manager/worker session. Verified execution mode (`execution.mode: verified`) is an explicit non-fallback selection: because this release ships no platform provider, a missing, unhealthy, incompatible, or drifted provider fails closed rather than falling back to portable execution. Portable, verified, legacy assurance-blind, cross-provider, and capability-drifted cache entries occupy disjoint cache key identities, and every adopted or published artifact carries the exact build-session receipt used at dispatch. Language source-closure adapters (Rust, SwiftPM, npm, pnpm, Yarn Classic, and Modern Yarn) bind ecosystem toolchains and resolve immutable source snapshots. See [docs/authoring-language-adapters.md](docs/authoring-language-adapters.md) for adapter authoring contracts and [docs/source-closure-adapter-conformance.md](docs/source-closure-adapter-conformance.md) for cross-adapter conformance.

## Compiled-command status, diagnostics, and repair

`curator status` reports read-only diagnostics for project closures, reporting one status code per declared skill and one diagnostic line per active compiled command. The command accepts `--check` to fail on non-current states and `--json` for machine-readable output.

Supported currentness codes include `up-to-date`, `current`, `not-installed`, `needs-install`, `content-drift`, `unresolvable`, `invalid-marker`, `unsupported-marker`, `build-command-drift`, `build-context-exposed`, `build-source-drift`, `build-input-drift`, `unsupported-build-driver`, `unusable-build-toolchain`, `missing-build-artifact`, `corrupt-build-receipt`, `build-artifact-drift`, `corrupt-build-cache`, `untrusted-build-cache`, `unsupported-build-platform`, `build-state-changed`, and `unknown-build-state`. Cause subcodes for `build-input-drift` are `build-root`, `target`, and `unattributed`.

Run `curator install` or `curator upgrade` to reconcile drifted or missing build state. See [docs/compiled-commands.md](docs/compiled-commands.md) for status codes, cause subcodes, logical cache key derivation, and repair semantics. For troubleshooting procedures, see [docs/troubleshooting.md](docs/troubleshooting.md).

## Maintenance and the build-cache grace period

`curator gc` executes maintenance to sweep unreferenced runtime store entries and compiled build cache artifacts. Garbage collection acquires an exclusive lock on the manager home directory and enforces a 24-hour grace period for unreferenced cache items. See [docs/compiled-commands.md](docs/compiled-commands.md) for maintenance locking and garbage collection criteria.

## Commands

See [docs/cli.md](docs/cli.md) for the complete CLI reference, command options, and flags.

<details open>
<summary>Project lifecycle</summary>

```bash
curator bootstrap  # create machine configuration and default directories
curator init       # initialize Skillfile.json and managed gitignore entries
curator add        # add or replace a skill declaration and install
curator remove     # remove a skill declaration from the project
curator install    # apply Skillfile.json and materialize dependencies
curator update     # fetch source repositories under skills_root
curator upgrade    # fetch dependency closure and install updated packages
curator status     # report manifest, installed, and compiled command states
curator list       # list configured projects and declared skills
```

</details>

<details>
<summary>Project subcommands</summary>

```bash
curator project add      # register a project alias and path in the machine configuration
curator project resolve  # resolve transitive dependencies for a project closure
```

</details>

<details>
<summary>Skill package validation</summary>

```bash
curator skill check  # validate a skill package directory against schemas
```

</details>

<details>
<summary>Global scope</summary>

```bash
curator global init     # initialize global machine-wide configuration
curator global add      # add a global skill declaration
curator global remove   # remove a global skill declaration
curator global list     # list installed global skills
curator global status   # check global scope installation status
curator global install  # materialize global skills into machine storage
curator global update   # update global source repositories
curator global upgrade  # upgrade global skill packages
```

</details>

<details>
<summary>Hybrid scope</summary>

```bash
curator hybrid add     # add a machine-stored hybrid skill declaration
curator hybrid remove  # remove a hybrid skill declaration
curator hybrid list    # list stored hybrid skills and activation targets
curator hybrid status  # check hybrid declaration status across projects
```

</details>

<details>
<summary>Audit and security</summary>

```bash
curator audit  # run security audit, pin trust, or publish audit records
```

</details>

<details>
<summary>Maintenance and shell integration</summary>

```bash
curator gc          # sweep unreferenced runtime store entries and build cache
curator shell-init  # print or install shell integration hooks
curator ui          # open interactive terminal status dashboard
```

</details>

<details>
<summary>Operator configuration</summary>

```bash
curator config show         # display active configuration parameters
curator config build-ssh    # configure SSH credentials for external build repos
curator config build-https  # configure HTTPS credentials for external build repos
```

</details>

## An open protocol

Curator conforms to an open specification. Independent implementations include [cocoaskills](https://github.com/ivanopcode/cocoaskills). The registry protocol is implemented by [Curator Skill Registry](https://github.com/relux-works/curator-skill-registry).

## Development

Curator uses an in-repository task board in `.task-board/`. Automated gates run under `.github/ci/` and mirror `Makefile` targets; see [docs/ci-gates.md](docs/ci-gates.md) for gate definitions, evidence paths, platform carve-out rules, and suite consumption requirements.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for working agreements: board-first workflow, discrete signed commits, and spec-first rules.

## License

Apache License 2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).
