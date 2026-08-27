# Curator implementation plan

This is the working plan for building Curator, an agent environment manager (AEM) in Go. Iterations consult this document and the task board; the plan is updated as decisions land. Protocol sections are cited as `Spec §N.M` against the 1.0.0-rc.2 [Curator Specification](https://github.com/relux-works/curator-spec).

## 1. Goal and definition of done

Curator v0.1 is a conforming independent implementation of the protocol: it works with the same skill packages, the same project manifests, the same installed layouts, and the same audit registries as the reference implementation, per the eight conformance requirements of Spec §17.1. Done means:

1. All conformance-relevant behavior is implemented and covered by tests, including the golden interoperability suite (Section 6 below).
2. `curator` installs a real project from `Skillfile.json` end to end on macOS, Linux, and Windows.
3. Markers, hashes, and runtime layouts produced by Curator are byte-compatible with the golden fixtures.
4. The task board reflects reality: every epic closed or explicitly deferred.

Out of scope for v0.1: publisher-side artifact signing, version ranges, lockfile, format translation, running MCP servers, a registry service implementation (client only). These stay on the board as future epics.

## 2. Engineering decisions

| Decision | Choice | Rationale |
|---|---|---|
| Language | Go 1.23+ | Static binaries, first-class Windows, stdlib `crypto/ed25519` |
| Module | `github.com/relux-works/curator` | Public repository |
| CLI framework | stdlib `flag` with a small subcommand dispatcher | The CLI surface is informative in the spec; avoid framework lock-in. Revisit only if UX demands it |
| Git operations | shell out to system `git` with `GIT_ALLOW_PROTOCOL=file:git:http:https:ssh`, URLs passed after `--` | Matches Spec §8.2 transport restrictions exactly; no libgit2/go-git divergence risk |
| JSON | stdlib `encoding/json` with strict unknown-field rejection where the spec requires it | Spec §5.1, §6.1 |
| TOML (codex configs) | `github.com/BurntSushi/toml` | Read-only MCP surface parsing (Spec §11.1) |
| JSONC (opencode) | minimal in-repo comment stripper, then stdlib JSON | One surface, trivial grammar; no dependency |
| Ed25519 | stdlib `crypto/ed25519` | Spec §13.2 |
| Hashing | stdlib `crypto/sha256`; content hash per Spec §8.5 byte layout | Byte compatibility is a conformance requirement |
| TUI (later phase) | `bubbletea` + `lipgloss`; tests with `tuitestkit` from skill-go-testing-tools | Closed-loop TUI testing |
| Testing | stdlib `testing`, table-driven; authoritative vectors from `curator-spec/conformance/v1` | Deterministic and shared by independent implementations |
| Lint/CI | `gofmt`, `go vet`, `golangci-lint`, GitHub Actions matrix (ubuntu, macos, windows) | Windows is first class |
| Layout | `cmd/curator/` (main), `internal/<domain>/` packages mirroring protocol domains | Spec-traceable package boundaries |
| Machine home | `~/.curator/` (env `CURATOR_CONFIG`, `CURATOR_SYSTEM_CONFIG`; system config `/etc/curator/config.json`, `%ProgramData%\curator`) | The machine home is tool-specific state (config, caches, runtime store), not a wire format; two managers on one machine must not share it. Shared wire formats stay exactly as specified: `Skillfile.json`, canonical `agent-skill.json`, legacy `csk-skill.json`, markers, adapter ledgers |

Package sketch: `internal/skillspec`, `internal/manifest`, `internal/devsub`, `internal/config`, `internal/identity`, `internal/gitops`, `internal/snapshot`, `internal/hashing`, `internal/closure`, `internal/whitelist`, `internal/locale`, `internal/marker`, `internal/runtimestore`, `internal/shims`, `internal/adapters`, `internal/mcp`, `internal/scopes` (global, hybrid, consumers, gc), `internal/registry`, `internal/audit`, `internal/shell`, `internal/cli`.

## 3. Working agreements

- Tasks live on the in-repo board `.task-board/` (epics, stories, tasks as directories with `README.md` and `progress.md`; artifacts in `.task-board/.resources/<ID>/`). Every change starts from a task; progress and status are updated as work proceeds.
- Commits are discrete and meaningful: one logical step per commit, imperative subject, body explaining what and why. Signed (SSH key, verified on GitHub) as Ivan Oparin <oparin@me.com>.
- The name of the alternative protocol implementation appears exactly once: the README open-protocol section. A CI check greps case-insensitively (zero matches outside README.md, exactly one inside). The protocol is cited as `Spec §N.M` against [curator-spec](https://github.com/relux-works/curator-spec); protocol file names (`Skillfile.json`, `agent-skill.json`, legacy `csk-skill.json`, `.csk-install.json`, `.csk-managed.json`) are part of the wire format and are used as-is.
- Spec first: when implementation and spec disagree, stop and resolve the spec question before coding around it.
- Every error message that the spec words normatively (allowlist refusal, MCP hint, conflicts with chains) keeps the same information content, not necessarily the same string.
- Style for prose (docs, board cards): plain technical English, no em dashes, no guillemets.

## 4. Phases

Phases map to board epics one to one. A phase is done when its acceptance criteria hold and its tests pass in CI on all three OS targets.

### Phase 0: bootstrap and CI (EPIC bootstrap-and-ci)

`go.mod`, `cmd/curator` hello skeleton with `--version`, Makefile (`build`, `test`, `lint`), GitHub Actions (test matrix + lint), golangci-lint config, `CONTRIBUTING.md` (working agreements from Section 3). Acceptance: green CI on all OS, `curator --version` prints a version.

### Phase 1: protocol formats (EPIC protocol-formats)

Parsing and validation, pure functions over bytes, no filesystem side effects beyond reads:

- `internal/skillspec`: `agent-skill.json` schemas 1 through 5, the `csk-skill.json` read alias, dual-file conflict detection, and every rule of Spec §4 (schema gating, unknown-field rejection for v2+, portable identifiers and paths, runtime roots, commands, capabilities, dependencies, and the legacy `agents/runtime.json` fallback).
- `internal/manifest`: `Skillfile.json` schema 1 (§6.1), add/remove declaration editing.
- `internal/devsub`: `Skillfile.dev.json` (§6.2).
- `internal/config`: user config (§7.1), enforced system config with locked keys (§7.2), env overrides.
- Error taxonomy: typed errors carrying the offending field path.

Test strategy: table-driven cases ported from the spec text, one test per MUST. Fixture corpus of valid and invalid manifests under `testdata/`.

### Phase 2: sources and identity (EPIC sources-and-identity)

- `internal/identity`: canonical source identity and segment-aware prefix matching (§8.2), allowlist gate semantics.
- `internal/gitops`: clone (transport allowlist, dash-URL refusal), fetch, ref resolution (tag, revision, branch incl. origin preference), `git archive` snapshot extraction with path-escape and link rejection, submodule detection.
- `internal/snapshot`: commit-keyed snapshot cache (§8.2).
- `internal/hashing`: content hash, byte-exact per §8.5.

Acceptance includes hash equality against the authoritative shared vectors.

### Phase 3: closure and activation (EPIC closure-and-activation)

`internal/closure`: pending-queue expansion from direct declarations (synthetic `<project>` consumer), unification to one commit and one identity with conflict errors carrying requirement chains, cycle detection, deterministic provider-first ordering, edge-based activation modes and command narrowing, requirement-command validation, active-command collision detection (§8.3, §8.4). Dev substitutions replace whole names and skip unification.

Test strategy: graph fixtures (diamonds, cycles, cross-version, cross-source, narrowing) with exact error assertions.

### Phase 4: install engine (EPIC install-engine)

- `internal/whitelist`: context copy rules (§4.2) incl. scripts-for-commandless rule and runtime_roots exclusion; atomic staging.
- `internal/locale`: consistency, fallback warning, frontmatter and `agents/openai.yaml` rendering (§4.3), applied only under a selected locale.
- `internal/marker`: marker payload, up-to-date checks incl. content re-hash tamper detection, atomic directory replacement (§8.5).
- `internal/runtimestore` and `internal/shims`: commit-keyed runtime store, self-contained Unix and Windows command launchers, stale shim removal (§8.6).
- `internal/adapters`: managed ledger `.csk-managed.json`, symlink/copy/auto, unmanaged-conflict refusal, native-discovery agents (§10).
- gitignore gate (§6.3), env files (§14.2), phase order orchestration (§8.1) with dry run.
- Gates wiring: system command PATH checks, migration warnings, moved-tag detection (warning and strict modes).

This is the largest phase; split into stories per bullet on the board.

### Phase 5: scopes and GC (EPIC scopes-and-gc)

Global scope under the machine home (§9.2) with home-level adapters; hybrid scope with targets, shadowing, machine-locale store rendering, reachability split (§9.3); consumers registry and runtime GC (§8.7).

### Phase 6: MCP requirements (EPIC mcp-requirements)

`internal/mcp`: per-agent configuration surfaces (§11.1) incl. TOML, JSONC, `enabled:false`, `disabledMcpjsonServers`; any/all semantics with hint-bearing errors; static stdio PATH warnings and project-only-pending warnings (§11.3); marker recording.

### Phase 7: registry client (EPIC registry-client)

`internal/registry`: canonical bytes, record parsing and statuses, pinned key parsing, Ed25519 verification, artifact matching, deny-wins resolution with the warning taxonomy (§13.1 through §13.3); snapshot verification with protected durable high-water state, fail-closed legacy migration, equal-version equivocation detection, and staleness bounds (§13.4); record cache TTL, offline grace, snapshot-bound pagination limits, bounded retries, and total deadlines (§13.5 and §13.9); install wiring (revoked fails, strict policy fails unknown, attestation into marker, all-tampered fails); `status --attest` re-check from markers; idempotent `audit --publish` submission (§13.10).

Test strategy: a local httptest registry serving signed fixtures; keys generated in tests; rollback, key rotation, equal-version equivocation, freeze, tamper, offline, retry safety, cursor cycles, and resource-limit scenarios. The authoritative registry-client vectors execute in the external conformance gate.

### Phase 8: audit gate (EPIC audit-gate)

Minimal conforming audit layer (§12): enabled/mode/fail_on decision semantics (§12.2), local revocations by hash and source glob, operator pins, canary-as-blocking, a small deterministic detector set (capability-vs-observed for network hosts and exec names in scripts), verdict cache. Backends beyond a null backend are deferred; the decision semantics MUST hold.

### Phase 9: shell and CLI polish (EPIC shell-and-cli)

`shell-init` hooks for zsh, bash, powershell with upward search and PATH restore (§14.1); the full informative CLI surface (§15): bootstrap, init, add, remove, install (all flags), update, upgrade, status (`--check`, `--json`, `--attest`), list, project, config show, skill check, global *, hybrid *, audit (incl. `--allow`, `--publish`), gc; exit code discipline.

### Phase 10: interoperability golden suite (EPIC interop-golden-suite)

The authoritative suite is owned by `curator-spec/conformance/v1`: skill packages with expected context file lists, marker objects, content hashes, signed registry records, snapshots, logs, and negative cases. CI checks out the released suite and passes its path through `CURATOR_CONFORMANCE_ROOT`; Curator keeps orchestration tests but no implementation-owned expected values.

### Phase 11: terminal UI (EPIC tui)

A `curator ui` view over installed state: projects, skills, activation, attestations; built on bubbletea, tested closed-loop with `tuitestkit` (reducer tests, golden screens). Strictly after the core conforms.

## 5. Dependency order

Phase 0 -> 1 -> 2 -> 3 -> 4 -> 5 -> {6, 7, 8 in any order} -> 9 -> 10 gate -> 11. Phases 6 and 7 only depend on 1 and 4 wiring points; start them in parallel when the install engine stabilizes.

## 6. Test strategy summary

- Unit: table-driven per MUST clause, typed error assertions.
- Integration: temp-dir projects with real git repositories built in test setup; full install cycles; idempotence (second install is `up-to-date`); tamper detection.
- Interop: the external versioned protocol suite (Phase 10) as the conformance gate.
- TUI: tuitestkit closed-loop (Phase 11).
- Windows: everything runs in the CI matrix; path handling tested with `win_path`, `.cmd` shims, `%ProgramData%` system config.

## 7. Risks and watchpoints

- Byte-compatibility of the content hash and marker JSON is the highest-risk surface; build Phase 2 hashing against the shared protocol vectors first.
- JSON field ordering in markers: the spec fixes content, not key order; markers are compared as parsed objects, hashes as bytes.
- Windows symlink permissions: adapter `auto` mode must degrade to copy without failing (Spec §10.1).
- Git version drift across CI runners: pin behaviors by testing outcomes, not stdout strings.
- Scope creep into registry service or TUI before the conformance gate; the board order enforces the gate.
