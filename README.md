# Curator

Curator is an agent environment manager (AEM): a single tool that manages what an AI coding agent gets in a project. Skills and their transitive dependencies, executable commands, MCP server requirements, per-agent delivery, and the security gates around all of it. Declarative, reproducible, verifiable.

Curator is implemented in Go and follows the [Curator Specification](https://github.com/relux-works/curator-spec), an open protocol for skill packages, project manifests, installation semantics, and the audit registry; sections are cited across this repository as `Spec §N.M`.

## Status

v0.1 development complete: all twelve phases of [docs/implementation-plan.md](docs/implementation-plan.md) are done. CI consumes the authoritative schemas and conformance vectors from `curator-spec` on ubuntu, macos, and windows, plus lint and the naming gate. Work is tracked on the in-repo task board under [.task-board/](.task-board/).

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

## Registry client guarantees

Curator binds persisted rollback and equivocation state to the canonical
registry URL, so signing-key rotation never resets the highest accepted
snapshot. This durable state lives under the Curator home `state/registry`
directory, outside the disposable `cache/registry` responses; upgrades migrate
legacy state without lowering it, and corruption or write failure is
fail-closed. A protected catalog distinguishes first use from deletion of a
previously accepted registry state. Record pagination rejects repeated or oversized cursors, more than
10,000 records per artifact query, and responses larger than 16 MiB.

Registry requests use bounded per-attempt and total deadlines. GET requests
retry network failures, `429`, and `503` at most twice after the first attempt.
Publication retries the exact body only with its deterministic
`Idempotency-Key`; other client errors and unsafe requests are never retried.
Redirects are rejected so a registry cannot move a request or bearer token to
another endpoint.

## An open protocol

The specification is an open protocol, not an internal contract: any manager
built from it interoperates with the same skills, the same project manifests,
and the same audit registries. That matters when internal security policies
rule out adopting an external binary and require an in-house implementation
instead. One such independent implementation of the protocol is
[cocoaskills](https://github.com/ivanopcode/cocoaskills) (Python); Curator's
conformance against the shared wire formats is enforced directly from the
versioned protocol suite in CI; this repository carries no private copy of the
expected protocol values.

The registry-service profile is implemented by
[Curator Skill Registry](https://github.com/relux-works/curator-skill-registry),
which serves signed audit and revocation records plus a verifiable transparency
log for any conforming Curator manager.

## Development

The repository uses an in-repo task board (`.task-board/`, epics, stories, and tasks as files) and the agent tooling connected under `agents/`. Go testing follows the closed-loop tooling of `skill-go-testing-tools` (including `tuitestkit` for terminal UI phases).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the working agreements: board-first workflow, discrete signed commits, spec-first rule.

## License

Apache License 2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).
