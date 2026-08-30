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

- **Skill packages**: `SKILL.md` files, context directories, and `agent-skill.json` manifests.
- **Project manifests**: `Skillfile.json` files with git references and local development substitutions.
- **Resolution**: Transitive dependency closures unified to one commit and one source identity per skill name.
- **Installation**: Separated context and runtime paths, install markers with content hashes, a commit-keyed runtime store, command shims, and managed per-agent adapters.
- **Scopes**: Project, global, and hybrid machine declarations.
- **MCP requirements**: Verification of declared MCP servers against agent configuration surfaces.
- **Security**: Source allowlists, declared capability boundaries, zero code execution during install, and Ed25519-signed audit registry client checks.
- **Operator credentials**: Per-repository SSH selection and scoped HTTPS token sources for external build repositories, matched by canonical source identity and never selectable by a package. Private HTTPS fetches use a manager-owned, host-pinned askpass broker; public HTTPS can remain anonymous. See [SSH credentials](docs/build-ssh.md) and [HTTPS credentials](docs/build-https.md).

## Commands without profile setup

Agents invoke project commands through `.agents/bin/<command>` on Unix and `.agents\bin\<command>.cmd` on Windows. Global installation publishes non-destructive forwarding shims to a safe user directory already on `PATH` when one is available; otherwise Curator reports the canonical global bin location.

Run bootstrap and upgrade to set up a repository:

```bash
curator bootstrap --if-missing --non-interactive --skills-root "$HOME/src/skills"
curator upgrade .
```

`curator upgrade` fetches dependency closures for the selected project. `curator upgrade --dry-run` plans updates without mutating disk state.

To enable shell integration, cache the optional hook:

```bash
curator shell-init --install
```

The command outputs the profile source command for your shell (`.zshrc` or `.bashrc` from `SHELL`, Git Bash preserved on Windows, PowerShell otherwise). Curator detects shell environments automatically and does not edit shell profile files directly. The cached hook does not start Curator during later shell launches. Set `CURATOR_AUTO_ENV=0` to disable automatic directory scans.

## Compiled-command status, diagnostics, and repair

`curator status` reports read-only diagnostics for declared skills and compiled `go-v1` commands. The command accepts `--check` to fail on non-current states and `--json` for machine-readable output. Reported status codes include `up-to-date` and `current` for clean states, and drift states (`not-installed`, `needs-install`, `content-drift`, `unresolvable`, `invalid-marker`, `unsupported-marker`, `build-command-drift`, `build-context-exposed`, `build-source-drift`, `build-input-drift`, `unsupported-build-driver`, `unusable-build-toolchain`, `missing-build-artifact`, `corrupt-build-receipt`, `build-artifact-drift`, `corrupt-build-cache`, `untrusted-build-cache`, `unsupported-build-platform`, `build-state-changed`, `unknown-build-state`), with subcodes (`build-root`, `target`, `unattributed`). Run `curator install` or `curator upgrade` to reconcile drifted or missing build state. See [docs/compiled-commands.md](docs/compiled-commands.md) for diagnostic status codes, logical cache key derivation, and repair semantics.

## Maintenance and the build-cache grace period

`curator gc` executes maintenance to sweep unreferenced runtime store entries and compiled build cache artifacts. Garbage collection acquires an exclusive lock on the manager home directory and enforces a 24-hour grace period for unreferenced cache items. See [docs/compiled-commands.md](docs/compiled-commands.md) for maintenance locking and garbage collection criteria.

## Registry client guarantees

Curator binds persisted rollback and equivocation state to the canonical registry URL. Key rotations preserve the highest accepted snapshot. State lives in `state/registry` under the Curator home directory. GET requests retry network failures up to two times with bounded deadlines. Requests reject HTTP redirects.

## An open protocol

Curator conforms to an open specification. Independent implementations include [cocoaskills](https://github.com/ivanopcode/cocoaskills). The registry protocol is implemented by [Curator Skill Registry](https://github.com/relux-works/curator-skill-registry).

## Development

Curator uses an in-repository task board in `.task-board/`. Automated gates run under `.github/ci/` and mirror `Makefile` targets; see [docs/ci-gates.md](docs/ci-gates.md) for gate definitions, evidence paths, and platform carve-out rules.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for working agreements: board-first workflow, discrete signed commits, and spec-first rules.

## License

Apache License 2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).
