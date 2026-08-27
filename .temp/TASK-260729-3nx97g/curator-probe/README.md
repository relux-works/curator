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

- **Skill packages**: `SKILL.md` plus context directories, with an implementation-neutral machine manifest (`agent-skill.json`, schemas 1 through 5) declaring commands, runtime layout, capabilities, and dependencies. The legacy `csk-skill.json` filename remains readable.
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

## Commands without profile setup

Shell profile changes are not required. After `curator install`, agents can
invoke project commands through `.agents/bin/<command>` on Unix and
`.agents\bin\<command>.cmd` on Windows. Global installation publishes
non-destructive forwarding shims to a safe user directory already on `PATH`
when one is available; otherwise Curator reports the canonical global bin
location. Installed launchers carry the project command directory and resolved
system dependency directories themselves, preserve the inherited `PATH`, and
return the child command status.

Repository bootstrap can remain idempotent and non-interactive:

```bash
curator bootstrap --if-missing --non-interactive --skills-root "$HOME/src/skills"
curator upgrade .
```

`upgrade` fetches only the selected project's direct and transitive dependency
closure. `upgrade --dry-run` plans with temporary sources and snapshots without
changing source checkouts, caches, security state, runtime state, or project
artifacts.

Interactive users who want bare command names and automatic project switching
can cache the optional hook once:

```bash
curator shell-init --install
# Add the source command printed above to .zshrc or .bashrc.
```

Automatic detection selects zsh or bash from `SHELL`, preserves Git Bash on
Windows, and otherwise selects PowerShell on Windows. The cached hook does not
start Curator during later shell launches. Curator never edits a profile
automatically. Set `CURATOR_AUTO_ENV=0` to retain global activation while
disabling project-directory scans.

## Maintenance and the build-cache grace period

`curator gc` runs one serialized maintenance pass. It acquires the exclusive
manager-home mutation lock, recovers any incomplete install transaction, and
only then marks and sweeps, so it cannot race an install, a rollback, or a
recovery, and cannot lose a consumer registry update. The same pass runs at the
end of every installation, under the lock that installation already holds.

Marking reads the live project, global, and hybrid scopes once. Runtime store
entries are marked from every supported install marker schema; compiled build
cache entries are marked from marker v2 build state and from every in-flight
transaction journal.

Anything the pass cannot prove keeps its artifacts, and keeps them across
passes. A consumer registry that exists but does not match the exact shape
Curator writes is reported and left untouched rather than rewritten; a
registered checkout is unregistered only once its scope is proven absent or
proven valid and empty. An *ambiguous* registry counts as unreadable: a document
that states `schema_version` or `consumers` more than once does not say what it
means, so it is refused by every reader and writer rather than resolved to
whichever value happens to come last. A skills directory or installed skill that
is a symbolic link or a reparse point is refused instead of followed, and any
marker that exists but cannot be read or validated blocks the build sweep. A
later pass therefore sees the same uncertainty and makes the same refusal,
instead of inheriting a registry the earlier pass had quietly emptied.

The sweep removes a protected build cache entry only when all of the following
hold: no marker and no journal references it, the cache root and the entry are
still verifiable manager-protected state, the entry is structurally exact and
self-consistent with the logical key its directory encodes, and it was
published more than **24 hours** ago — Curator's documented grace period.
Everything else is retained and reported as a maintenance warning, including
corrupt receipts, untrusted roots, symlink or reparse escapes, and ownership or
DACL failures. Entry content is never executed, adopted, or permission-repaired,
and a receipt alone is never treated as proof of provenance or of a live
consumer. Retaining an unreferenced entry is always safe: the only cost of a
removal is one rebuild.

Every decision and every removal is bound to the directory object the pass
proved, not to the pathname it proved it under. The decisive classification of a
candidate — its exact members, its receipt, its artifact bytes and size — is
read through the descriptor of the entry the pass resolved, and the rename that
retires it and the deletion behind it resolve through the proven cache root; an
entry whose parent is no longer that object is retained and reported. Exchanging
the cache-root path after validation can therefore neither redirect a removal
outside the Curator cache root nor let a planted replacement supply the verdict
for the entry that is actually being removed.

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
