# Architectural Security Audit — Curator Manager, 2026-09-10

**Scope.** This audit covers the Curator manager implementation (`relux-works/curator`)
against the Curator Protocol specification (`relux-works/curator-spec`,
`1.0.0-rc.9` band). It is an **architectural** audit: it examines trust
boundaries, default postures, and cross-component composition. It is not a
line-by-line code review; a follow-up code-level pass is planned per finding.

**Method.** The full protocol documents (`protocol/core.md`, `registry.md`,
`environments.md`, `assurance.md`), `SECURITY.md`, and the manager profile
were read in full. The implementation was surveyed at the level of packages and
trust boundaries: `internal/registry`, `internal/gitops`, `internal/buildrepo`,
`internal/closureexec`, `internal/install`, `internal/transaction`,
`internal/envprofile`, `internal/shell`, `internal/runtimestore`,
`internal/gitcred`, `internal/whitelist`, `internal/scriptpolicy`,
`internal/contextaudit`, `cmd/curator` (umbrella dispatch, shims,
shell integration), and `install.sh`.

**Companion documents.** The specification-side findings live in
[`curator-spec/docs/security-audit-2026-09.md`](https://github.com/relux-works/curator-spec/blob/main/docs/security-audit-2026-09.md).
The registry-service findings live in
[`curator-skill-registry/docs/security-audit-2026-09.md`](https://github.com/relux-works/curator-skill-registry/blob/main/docs/security-audit-2026-09.md).
Finding IDs are shared across the three documents.

**Remediation tracking.** The decomposition into epics, stories, and tasks on
the shared task board is in [Appendix A](#appendix-a-remediation-decomposition).
Epic: `EPIC-260910-2hw1xb`.

## 1. Executive summary

The implementation is unusually disciplined for its domain. Protocol
boundaries — raw-object admission, manager/worker sessions, the protected build
cache, the SSH wrapper argv, the HTTPS askpass broker — are implemented close
to the letter of the spec, and the fail-closed behavior the spec demands is
present where sampled. The principal risks are not in cryptography or
parsing. They concentrate in:

1. **The default-open posture** of the audit and registry gates (advisory
   audit, advisory registry policy, empty allowlists) — the strong mechanisms
   the protocol defines are opt-in.
2. **Points where package-controlled or project-controlled data meets
   execution**: the shell hook auto-sourcing `.agents/env.sh`, MCP declaration
   packages controlling launch-time command execution with operator secret
   passthrough.
3. **A cross-repo gap in the registry trust composition**: record pages are
   not bound to a snapshot boundary and this client never replays the
   transparency log, so a key-holding registry can hide a revocation while
   serving an honest, advancing snapshot (finding **R1**, shared with the
   registry audit).

## 2. Confirmed strengths

These were verified in code and should be preserved by future changes:

- **Registry client** (`internal/registry`): bounded retries with
  idempotency-key discipline, redirect rejection, canonical base64 checks,
  `key_id`-bound verification against pinned keys, fail-closed rollback state
  with a catalog that detects state deletion, migration that refuses to lower
  the high-water mark, `state/registry` protected at 0700/0600 with fsync.
- **External build repositories** (`internal/buildrepo`): clean git
  environment (`GIT_CONFIG_GLOBAL/SYSTEM` pinned to manager-owned empty files,
  `GIT_TERMINAL_PROMPT=0`, `GIT_PROTOCOL_FROM_USER=0`), `fsckObjects`,
  `http.followRedirects=false`, `sslVerify=true`, proxies disabled, SSH argv
  with pinned options (`ExactSSHCommand`), a host-pinned askpass broker that
  answers only exact prompts, independent re-hashing of consumed objects,
  LFS-pointer rejection, bounded limits.
- **`manager-worker-v1`** (`internal/closureexec`, `internal/godriver`):
  session nonce, worker identity proof, exactly one list plus one build permit,
  bounded output, replay-copy verification after execution, snapshot
  re-validation before publication.
- **Transactions and ledgers** (`internal/transaction`, `internal/adapters`,
  `internal/globalbins`): staging plus journal plus reverse-order rollback,
  refusal to overwrite unmanaged files, ownership ledgers for shims.
- **`script-worker-v1` honesty**: schema-8 enforced commands fail closed as
  `script_execution_policy_unsupported` (`internal/scriptpolicy`) — exactly
  the spec rule.
- **Umbrella dispatch** (`cmd/curator/umbrella.go`): providers resolved
  inside manager-published directories are refused
  (`subcommand_provider_untrusted`), symlink-aware.
- **Snapshot extraction** (`internal/gitops`): path safety, case-folding
  probe, symlink refusal, size bounds, `--` operands, `GIT_ALLOW_PROTOCOL`
  restricting `ext::`.

## 3. Findings

Severity ratings are architectural (impact × likelihood under the protocol
threat model, assuming a malicious skill/project/registry input and an honest
operator machine unless stated otherwise).

### S6 / I1. Shell hook auto-sources project-controlled `.agents/env.sh` (High)

`internal/shell/shell.go`: the cached zsh/bash hook walks up from `$PWD` on
every `chpwd`/`precmd` and sources the first `.agents/env.sh` it finds, with
no digest check and no approval gate (`CURATOR_AUTO_ENV=1` by default).
`internal/envfiles` writes these files for managed projects, but **any project
checkout can ship its own `.agents/env.sh`**, so `cd` into a hostile
repository is arbitrary shell-code execution in the interactive session. This is
direnv **without** `direnv allow`. PowerShell follows the same pattern.

No other project-controlled file class auto-executes on `cd` in these
shells; this surface is unique to the hook.

*Fix direction:* record the digest of every written env file on the manager's
project trust surface; make the hook verify the digest (or manager ownership)
before sourcing and warn + skip on mismatch; add a one-time operator approval
for foreign files. Tracked as `STORY-260910-2awkzu`.

### S4. MCP declaration packages: launch-time execution with unbounded env passthrough (High)

The environments capability resolves MCP declaration packages
(`agent-mcp.json`) whose `stdio` `command`+`args` execute **in the agent
tool at launch**. The spec admits a binary allowlist bounds nothing (`npx`,
`uvx`, `node`, `sh` admit any program through `args`). The implementation
defaults keep both exposure doors open:

- `mcp_package_allowlist` defaults to empty = every network identity is
  admitted (`internal/config/environments.go`);
- `passable_env_names` defaults to `nil` = unbounded
  (`internal/envregistry`), and the reserved-name check excludes only
  manager variables — not operator secrets.

A hostile declaration package can therefore run `command: npx` with arbitrary
args and collect `env_names: ["GITHUB_TOKEN", ...]` values in the launched
environment. Root-context modules (also package data) can prompt the agent to
invoke the server, compounding the exposure. The `context-secret-material`
detector scans args/URLs for *leaked tokens* but does not bound `command`.

*Fix direction:* default `passable_env_names` to empty (opt-in per name); a
loud install-time warning when the MCP allowlist is empty; surface the
resolved `command`+`args` to the operator at profile install/update. Tracked
as `STORY-260910-1lf0m5`.

### R1 (client side). Records pages are not bound to a snapshot boundary (High, cross-repo)

`internal/registry/http.go` fetches `/v1/records` and verifies each record's
Ed25519 signature, and `internal/registry/snapshot.go` enforces monotonic
snapshot versions against persisted high-water state — but the client never
reads `/v1/log` (no code path exists) and the records response carries no
snapshot boundary. A key-holding registry can therefore serve an **honest,
advancing** `/v1/snapshot` while evaluating `/v1/records` at an **old**
boundary, hiding an appended `revoked` record and re-serving a stale `audited`
one. Rollback state does not catch this; the response cache TTL does not help
(the server controls responses). This composes with advisory defaults (S1)
into practical revocation hiding.

*Fix direction:* the spec adds a boundary to records/log responses; the client
rejects or warns when a page boundary is below its persisted high-water; an
optional strict mode replays the log for revoked decisions. Tracked as
`STORY-260910-25yc0h` (spec + client) and `STORY-260910-3rvvxh` (service).

### S1 + S3. Default-open gates; revocation hiding via DoS under advisory policy (Medium)

`internal/config/config.go` defaults: `Mode: "advisory"`,
`FailOn: "high"`, `RegistryPolicy: "advisory"`. The source allowlist and MCP
allowlist default to empty (= allow all). Under these defaults a cloned
hostile project installs without a single blocking gate, and an attacker on
the network can suppress delivery of a revoked record (unreachable registry →
artifact unknown → advisory permits install). Deny-wins requires *seeing* the
revocation; hiding it is enough. Each individual choice is defensible; the
composition makes revocation best-effort under defaults.

*Fix direction:* specify a recommended hardened-defaults profile; surface
unreachable-registry prominently during install; document the residual.
Tracked as `STORY-260910-2qmrb8`.

### S5. Profile store lacks the protected-boundary discipline of the build cache (Medium)

The build cache has a normative ownership/permission/containment contract
(core §9.3). The environments store, locks, and markers have no equivalent:
locks and markers are records, not signatures, and `env resolve` treats
link-target identity as sufficient currency — a same-user swap of store bytes
(the recorded-residual window is acknowledged in environments §10.1) passes
without detection. Prompt material (system prompt, root context) is the
sharpest surface a profile carries and is currently the *least* protected
store under the manager home.

*Fix direction:* normative ownership/permission validation for the
environments root and store at resolve; verify recorded surface hashes for the
small `system-prompt` and `root-context` files. Tracked as
`STORY-260910-148pj1`.

### S2. TOFU bootstrap and equivocation (Medium)

First snapshot pinning is trust-on-first-use with no authenticated bootstrap
checkpoint; recovery-after-loss is specified but initial pinning is not.
Per-machine monotonicity permits divergent-but-advancing views across clients
(equivocation) with no client-side detection.

*Fix direction:* signed bootstrap checkpoint interchange; optional
cross-registry Merkle-root comparison warning. Tracked as
`STORY-260910-6bo7ej`.

### I2. HTTPS askpass secret travels through the child environment (Low)

`internal/buildrepo/httpsbroker.go` delivers the resolved secret via
`CURATOR_BUILD_HTTPS_ASKPASS_SECRET` in the fetch process environment. The
broker design is strong (host-pinned, exact prompts, 0600 state), but the
secret is visible to every descendant of the fetch child. Deliver it through a
pipe or inherited fd. Tracked as `TASK-260910-31ocjt`.

### I3. `install.sh` verifies only same-origin checksums (Medium, supply chain)

`checksums.txt` is fetched from the same release as the artifact, so it
protects against channel corruption but not against a compromised release. The
README suggests `gh attestation verify` manually; the installer should do it.
Tracked as `TASK-260910-2t0iun`.

### S7. Portable assurance is not a sandbox (By-design; document)

Declared-only script commands (all of them today) execute with a bare
launcher `exec`; capability declarations are documentation. Compiled artifacts
run with user privileges at first invocation. This is honest and specified —
but the top-level README should say it plainly, with `script-worker-v1` and
verified mode named as the enforcement paths. Tracked as
`TASK-260910-3i6vod`.

### Informational observations (no action required beyond code-level pass)

- `gitops.run` inherits `os.Environ()` for skill-source clones (operator
  environment is trusted; external build repos use the clean path) — document
  the asymmetry.
- `detectRelease` executes tool binaries from `PATH` during `env status`
  (operator PATH is trusted).
- Registry HTTP client minor items: non-atomic remove+rename cache write on
  Windows (benign cache miss); float64 timestamps in the record cache
  (~256 ns precision loss, irrelevant at TTL scale).
- Marker files are 0644 (non-secret records; acceptable).
- `scriptpolicy` fail-closed behavior is conformant and correct.

## 4. Priority summary

| # | Finding | Severity | Board |
|---|---|---|---|
| S6/I1 | Shell hook auto-source without approval | High | `STORY-260910-2awkzu` |
| S4 | MCP env passthrough / launch execution | High | `STORY-260910-1lf0m5` |
| R1 | Records not bound to snapshot (client) | High | `STORY-260910-25yc0h` |
| S1+S3 | Advisory defaults; revocation hiding | Medium | `STORY-260910-2qmrb8` |
| S5 | Profile store boundary | Medium | `STORY-260910-148pj1` |
| S2 | TOFU / equivocation | Medium | `STORY-260910-6bo7ej` |
| I3 | Installer supply chain | Medium | `TASK-260910-2t0iun` |
| I2 | Askpass secret via env | Low | `TASK-260910-31ocjt` |
| S7 | Sandbox posture docs | Info | `TASK-260910-3i6vod` |

## 5. Code-level follow-up plan

For each accepted finding, the next phase is a line-by-line pass in this
order: `internal/shell` (S6 fix), `internal/registry` (fuzzing, TOCTOU of
state writes, boundary design for R1), `internal/buildrepo/local.go` (raw
object reader framing, limits, path collisions), `internal/closureexec`
(worker protocol replay/framing), and a permissions audit across the manager
home (S5).

---

## Appendix A. Remediation decomposition

Tracked on the shared Curator board. Epic **`EPIC-260910-2hw1xb` —
security-audit-remediation-manager-and-spec**. The full audit document is
attached to the epic as a board resource.

| Story | Finding(s) | Tasks |
|---|---|---|
| `STORY-260910-2awkzu` shell-hook-project-env-approval-gate | S6 (High) | `TASK-260910-1wjst3` spec-hook-trust-rule · `TASK-260910-1952mz` manager-hook-digest-pin · `TASK-260910-3ungjy` hook-approval-command |
| `STORY-260910-1lf0m5` bound-mcp-declaration-exposure | S4 (High) | `TASK-260910-2ohnjo` spec-bound-env-passthrough · `TASK-260910-gocke2` manager-mcp-install-surfacing |
| `STORY-260910-25yc0h` records-boundary-binding | R1/P1 (High) | `TASK-260910-1b1ens` spec-records-boundary-envelope · `TASK-260910-2n0233` client-boundary-high-water-check |
| `STORY-260910-2qmrb8` secure-defaults-for-audit-gates | S1+S3 (Medium) | `TASK-260910-2qtiho` spec-hardened-defaults-profile · `TASK-260910-1sapuy` install-unreachable-registry-warning |
| `STORY-260910-148pj1` profile-store-protected-boundary | S5 (Medium) | `TASK-260910-39fzpq` spec-environment-store-boundary · `TASK-260910-32gki6` manager-envresolve-boundary-check |
| `STORY-260910-6bo7ej` tofu-and-equivocation-mitigations | S2 (Medium) | `TASK-260910-1tvf2t` spec-bootstrap-checkpoint · `TASK-260910-2vnjej` client-cross-registry-root-check |
| `STORY-260910-234vmx` supply-chain-and-credential-hardening | I2, I3, S7 | `TASK-260910-2t0iun` installer-attestation-verification · `TASK-260910-31ocjt` askpass-secret-via-pipe · `TASK-260910-3i6vod` top-level-sandbox-posture-docs |

Registry-service findings are decomposed under epic
**`EPIC-260910-16qce1`** (see the registry audit document in
`curator-skill-registry`).
