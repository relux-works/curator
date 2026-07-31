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

### Gates and tooling

Every gate below is a script under `.github/ci/`, called directly by
[.github/workflows/ci.yml](.github/workflows/ci.yml) and mirrored by a `make`
target for local use. CI calls the scripts rather than `make` because `make` is
not a guaranteed tool on the Windows runner. Each gate writes its raw stream and
its report under `EVIDENCE` (default `.temp/ci-evidence/`), which CI uploads per
runner, so any claim about a gate can be checked against the run that produced
it.

| Tool | What it gates | How to run it | Where its output goes |
| --- | --- | --- | --- |
| `test-gate.sh` | plans the run from the supplied conformance root, executes it, then enforces the platform-case ledger; every status is fatal | `make ci-test` | `$EVIDENCE/test/` — `go-test*.json`, `suite-plan.txt`, `platform-cases.txt`, `skips-observed.tsv` |
| `test-gate.sh` (`-race`) | the same gate under the race detector | `make race` | `$EVIDENCE/race/` |
| `suite-plan.sh` | decides, from the root alone, which packages it serves, which it cannot, and which this platform does not qualify | called by `test-gate.sh` | `$EVIDENCE/*/suite-plan.txt`, `plan-*.txt` |
| `platform-case-gate.sh` | requires every case [`platform-cases.tsv`](.github/ci/platform-cases.tsv) names on this runner, and classifies every skip against [`skip-classes.tsv`](.github/ci/skip-classes.tsv) | called by `test-gate.sh` | `$EVIDENCE/*/platform-cases.txt`, `skips-observed.tsv` |
| `ledger-consistency.sh` | proves each ledger row against the real per-GOOS builds via `go list` — no runner needed | `make ledger-check` | `$EVIDENCE/ledger/ledger-consistency.txt` |
| `excluded-packages.sh` | the one resolver of "which packages does this platform not execute, and on whose authority" | called by the two gates above | stdout (TSV) |
| `candidate-suite.sh` | rejects a non-immutable candidate revision; records a candidate root's identity as candidate-only evidence | `make candidate-verify-ref CANDIDATE_REF=…` / `make candidate-record CANDIDATE_ROOT=…` | `$EVIDENCE/candidate/candidate-suite-identity.txt` |
| `toolchain-identity.sh` | asserts the resolved Go toolchain is exactly `go.mod`'s, with `GOTOOLCHAIN=local` and `GOENV=off` read back | run by every Go-consuming job | job log |
| `no-broad-suppression.sh` | rejects bare `//nolint`, bare `//#nosec`, production-path lint exclusions, wholesale disabling, and unrecorded `gosec` exclusions | `make no-broad-suppression` | job log |
| `gate-selftest.sh` | drives every gate above against synthetic inputs and asserts a real exit code for each negative case | `make gate-selftest` | job log |

`make ci-test`, `make race` and `make check-ci` require `CURATOR_CONFORMANCE_ROOT`
to point at a materialised `<curator-spec>/conformance/v1`; they refuse to run
without it, because a gate that runs with the conformance suite unset is a
smaller gate wearing the same name.

The committed protocol-suite pin is declared once, as `SPEC_PIN` in the workflow
`env:` block, and every job reads it from there. A schema v6 candidate suite is
never committed and never a default: it enters only through the
`candidate-conformance` job, on an explicit `workflow_dispatch` that supplies a
full 40-character revision or a pre-materialised root. That job sets
`CI_REQUIRE_FULL_ROOT=1`, so a candidate must serve the whole package set, and
everything it emits is stamped in the artifact itself as candidate-only evidence
— neither a published release nor a conformance claim.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the working agreements: board-first workflow, discrete signed commits, spec-first rule.

## License

Apache License 2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).
