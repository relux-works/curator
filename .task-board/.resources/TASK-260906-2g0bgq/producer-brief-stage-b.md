# Producer brief: implementation stage (b) — managed homes, seeds, read-only resolve, fragment, untracked launch

## Where and what

- Repository `~/Developer/ReluxWorks/curator` (Go 1.25.5). Worktree
  `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-b`, branch `feat/agent-environments-stage-b`,
  base = curator main once stage (a) has landed (`the stage (a) head `834b40f6` (feat/agent-environments-stage-a, PR #59, landing on curator main by fast-forward; rebase onto main with -S once it lands and prove identity with git range-diff)`). First run
  `git submodule update --init --recursive`.
- Authority: curator-spec main `f39f4a9` — `protocol/environments.md` revision 1.1: §5.3 (referenced
  form and the `@path` external-include rule), §5.5 (system-prompt output, managed homes only), §5.7,
  §5.8 (MCP launch-channel files per adapter, `codex_cli` fixed layer path, CCJ-1 bytes), §7.1 (XDG
  seeds allowlist), §7.2 (forms), §7.3 (system-prompt channels with `argument`/`with`/`name`), §7.4
  (credential passthrough, passthrough strategies, provisioning seeds, `isolated` support matrix and
  `environment_isolated_unsupported`), §7.5–§7.9, §8.1 (modes; the claude_code root-context surface is
  always a copied regular file), §8.2 (marker `surfaces` with copies and reasons), §8.3 (versioned
  backups), §8.4 (drift), §9.5 (onboarding detect/backup/takeover — only what stage (b) needs),
  §10.1 (`env resolve`, lock-free verification, `environment_home_stale`, `--repair` with the mutation
  lock), §10.2 (`launch-env-fragment-v1` with `profile.lock_sha256`, `precedence`, `system_prompt`,
  `mcp`, `path_prepend`), §10.3 (profile-influence boundary), §10.4, §12 (status rows incl.
  `environment_passthrough_detached`, `environment_seed_shadowed`, `environment_seed_unreadable`);
  `schemas/v1/launch-env-fragment-v1.schema.json` and `agent-environment-marker-v1.schema.json`;
  `conformance/v1/expected/environments/*` (the `referenced-*`, `system-prompt-composed` and `mcp-*`
  sets — the sets stage (a) left skipped), `conformance/v1/vectors/environments.json`;
  `profiles/manager.md` §12.2/§12.4/§12.5/§12.7; `cli/curator.md` `env resolve` and `env status` rows;
  Decision 0012 D6 and D8; Decision 0013 D6 (what the launcher expects from the fragment).
- Curator's own code: everything stage (a) added (`internal/envprofile`, `contextaudit`, `contextstore`,
  `pkgversion`, the materialization and marker packages), plus `internal/adapters`, `marker`,
  `transaction`, `managerlock`, `scopes`, `globalbins`, `config`, `interop`.

## Scope — exactly the epic's stage (b) list

1. **Managed homes** (§8.1, §8.2, §7.4): provision a managed home per profile × environment under the
   manager-owned environments root; materialize the managed surfaces (root context in both forms,
   system prompt when the chain carries applicable system modules, MCP files); record every surface in
   the marker with its content hash, its form, and — for a copy — the reason (the claude_code
   root-context surface is always copied); `linked` and `copied` in-place modes for native homes.
2. **Provisioning seeds and passthrough** (§7.4): the closed per-adapter seed class (non-credential,
   one-time, never hashed) with the verified seed shapes the sprint recorded (`.claude.json` with
   `hasCompletedOnboarding` and the project trust entry, plus
   `hasClaudeMdExternalIncludesApproved` when the referenced form is materialized; codex `config.toml`
   entries; pi `settings.json`); the passthrough strategies per adapter with their liveness row
   (`environment_passthrough_detached`); `isolated` refused with `environment_isolated_unsupported`
   where §7.4 says so; the XDG seed allowlist and reconciliation with `environment_seed_shadowed`;
   `environment_seed_unreadable` stops provisioning (absence vs unreadable).
3. **Read-only `env resolve` and the fragment** (§10.1–§10.4): lock-free verification over exactly the
   marker's surfaces (link-target identity into an immutable store entry suffices for symlinked
   surfaces); a stale home reports `environment_home_stale` with reasons and emits **no** fragment;
   `--repair` takes the mutation lock with a bounded wait and a distinct lock-acquisition diagnostic,
   provisions or repairs, then emits; the fragment is the closed `launch-env-fragment-v1` with
   `profile.lock_sha256`, the `precedence` object, `env`, `system_prompt`, `mcp` (path, sorted
   `env_names` union, channel descriptor with `argument`/`with`/`name`) and `path_prepend`; `--format
   json|env|shell`; the §10.3 boundary enforced (registry-declared names only, values below the root).
4. **MCP launch-channel materialization** (§5.8, §7.8, Decision 0012 D6): one inert, hashed,
   marker-recorded file per adapter format, managed homes only, `codex_cli` at
   `<home>/curator-mcp.config.toml` with the fixed `args` spelling; the allowlist over MCP package
   canonical source identities (`mcp_package_not_allowed`); `env_names` grammar and reserved-name
   exclusion.
5. **Untracked `curator run`**: the launcher is a separate repository — in curator, stage (b) delivers
   only what the launcher consumes (the fragment above) plus the `curator run` umbrella dispatch of
   §11 (`curator-<name>` on `PATH`, missing provider names the executable and installation guidance).
   Do **not** implement the launcher itself here.
6. **`env status`** (§12): the profile × environment × surface matrix read-only, with the rows §12 and
   §7.4/§7.5/§7.1 name (stale, drifted, shadow-inert with the acknowledgment downgrade, passthrough
   detached, seed shadowed, isolated unsupported, size advisory).

Out of scope: composition/weights CLI, `path` kind and onboarding import, config schema 2 surfaces,
ax integration (stages (c) and (d)).

## Conformance subset

The `referenced-claude-code-composed`, `referenced-opencode`, `referenced-opencode-zero-modules`,
`system-prompt-composed` and the three `mcp-*` expected sets must pass byte for byte through
`CURATOR_CONFORMANCE_ROOT`, replacing the stage-deferred skips stage (a) registered — remove that skip
class usage for exactly the cases this stage implements and leave the class only where a case is still
deferred. Add the fragment and marker schema-driven tests for the members this stage introduces.

## Delivery

Small signed commits, each building and testing green. Gates: `go build ./...`, `go vet ./...`,
`gofmt -l`, `golangci-lint run ./...` if installed, `go test -count=1 -race` on the touched packages,
the vector families through `CURATOR_CONFORMANCE_ROOT`, `bash .github/ci/gate-selftest.sh`, the
platform-case gate for the three GOOS values as `ci.yml` runs it, and `go test -count=1 -timeout 30m
./cmd/curator` once at the end. Every new required case in `.github/ci/platform-cases.tsv`; every skip
in a registered class with a truthful reason (no phantom environment variables — that was a cycle-1
finding in stage (a)). Do not push, tag, or open a PR. Attach `TASK-260906-2g0bgq_drafting-report.md` (surface →
package map, vector sets passed with counts, gate outputs, anything deferred with the reason);
`task-board handoff TASK-260906-2g0bgq --role developer`. Never write LOGBOOK.md or anything into the control
root; the story workspace carries an empty delta by design.

## Facts you must not re-derive

The verification sprint (board resource `TASK-260905-3jq1so_verification-sprint.md`) established:
Claude Code keys its Keychain item by a service-name suffix over `CLAUDE_CONFIG_DIR`; codex does not
truncate the global `AGENTS.md`; codex and pi rewrite `auth.json` in place; `@path` includes outside
the launch directory are dropped without the project approval key, and the same guard skips a linked
user-level `CLAUDE.md`; codex `-p` takes exactly one value and silently ignores a missing layer file.
Use them; do not re-probe unless a spec sentence is ambiguous, and label anything new as
docs-confidence in the report.
