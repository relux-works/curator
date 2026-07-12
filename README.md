# Curator

Curator is an agent environment manager (AEM): a single tool that manages what an AI coding agent gets in a project. Skills and their transitive dependencies, executable commands, MCP server requirements, per-agent delivery, and the security gates around all of it. Declarative, reproducible, verifiable.

Curator is implemented in Go and follows a published protocol specification for skill packages, project manifests, installation semantics, and the audit registry, so environments it manages interoperate with other conforming tools. The specification lives in the Relux Works organization; sections are cited across this repository as `Spec §N.M`.

## Status

v0.1 development complete: all twelve phases of [docs/implementation-plan.md](docs/implementation-plan.md) are done, including the interoperability golden gate (byte-equality against fixtures produced by an independent conforming implementation). CI runs tests on ubuntu, macos, and windows plus lint, a naming gate, and the interop gate. Work is tracked on the in-repo task board under [.task-board/](.task-board/).

## Install

```bash
# Homebrew (macOS, Linux)
brew install relux-works/tap/curator

# installer script (macOS, Linux)
curl -fsSL https://raw.githubusercontent.com/relux-works/curator/main/install.sh | sh

# Scoop (Windows)
scoop bucket add relux-works https://github.com/relux-works/scoop-bucket
scoop install curator

# Go toolchain
go install github.com/relux-works/curator/cmd/curator@latest
```

Debian and RPM packages ship with every [release](https://github.com/relux-works/curator/releases), together with SBOMs and cosign signatures. macOS binaries are Developer ID signed (Relux Works, LLC). Verify any downloaded artifact:

```bash
gh attestation verify <artifact> --owner relux-works
```

## What Curator manages

- **Skill packages**: `SKILL.md` plus context directories, with a machine manifest (`csk-skill.json`, schemas 1 through 5) declaring commands, runtime layout, capabilities, and dependencies.
- **Project manifests**: `Skillfile.json` with exact git references; non-committed development substitutions.
- **Resolution**: transitive dependency closures unified to one commit and one source identity per name, with activation modes.
- **Installation**: context and runtime separation, install markers with content hashes, a commit-keyed runtime store, command shims, managed per-agent adapters.
- **Scopes**: project, global, and hybrid (machine-stored, per-project activation).
- **MCP requirements**: read-only verification of declared MCP servers against agent configuration surfaces.
- **Security**: source allowlists, declared capabilities, no code execution at install time, and an audit registry client (Ed25519 signed records, deny-wins federation, snapshot verification).

## Development

The repository uses an in-repo task board (`.task-board/`, epics, stories, and tasks as files) and the agent tooling connected under `agents/`. Go testing follows the closed-loop tooling of `skill-go-testing-tools` (including `tuitestkit` for terminal UI phases).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the working agreements: board-first workflow, discrete signed commits, spec-first rule.

## License

Apache License 2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).
