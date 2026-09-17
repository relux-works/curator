# Brief — TASK-260910-gocke2: MCP declaration surfacing, allowlist warning and passthrough profiles (S4, manager)

Story `STORY-260910-1lf0m5` (bound-mcp-declaration-exposure), wave 1.
Rules: `remediation-manager-producer-rules.md` (attached). Role: developer.
This is the story's only implementation task.

## Spec (read first; it is the contract)
curator-spec `23dafa7` (landed by PR #63): `protocol/environments.md` §2.2
(`passable_env_names`: absent knob = empty list under `s4-enforce`, explicit
`null` = unbounded; empty `mcp_package_allowlist` MUST warn
`mcp_package_allowlist_empty` at `profile install`, `profile update` and
`env status`), new §2.3 (one closed-column surfacing row per MCP declaration
package — `mcp-declaration <package> <version> <transport> command=<command>
args=<json-array-no-spaces> env_names=<json-array-no-spaces>`, one LF,
ascending package-name byte order, printed after the audit gate passes and
before the lock is published or any surface is (re-)materialized; `env status`
repeats the rows; informative, never fails), §9.1/§9.2 (install and update
sequence steps), §10.3 (profiles `s4-warn`: absent knob behaves as unbounded
but every passed operator variable outside an explicitly configured list —
or any passed variable when the knob is absent — warns
`mcp_env_passthrough_unlisted` naming the variables and the knob with the
migration hint; `s4-enforce`: absent = empty, unlisted requested names are
dropped with `mcp_env_passthrough_dropped`; a manager MUST ship `s4-warn`
first), §12 posture (allowlist-empty warning row, active S4 profile with the
effective `passable_env_names`, the surfacing rows per reported scope;
warnings never make a row non-current), §12.1/§12.2. Vectors:
`conformance/v1/vectors/environments-env-passthrough.json` (default
resolution under both profiles incl. explicit null and narrowed list,
allowlist warning cases, surfacing output bytes incl. an argument with a
space, install/update order cases) and the `manager-config-v2` schema cases
for the new default. Audit: `docs/security-audit-2026-09.md` S4 in this
repository; the spec audit S4.

## Current code
`internal/envfragment/envfragment.go` `BoundEnvNames` (nil passable =
unbounded), `internal/envprofile/managed.go` ≈1486 (fragment MCP env names
from `req.Machine.PassableEnvNames`), `internal/config/environments.go`
(`PassableEnvNames` nil for null; `mcp_package_allowlist` parsing),
`internal/contextmaterialize` (`MCPEnvNames`), `cmd/curator` profile
install/update and `envstatus.go`.

## Deliverable
1. **Profiles**: implement both `s4-warn` and `s4-enforce` behind ONE
   internal option/constant; ship **`s4-warn`** as the default. Under
   `s4-warn` keep today's unbounded behaviour for an absent knob but emit
   `mcp_env_passthrough_unlisted` (once per launch/resolution, naming the
   variables and `passable_env_names`, with the migration hint "list the
   named variables to keep passing them after the flip"); under `s4-enforce`
   an absent knob is the empty list and unlisted requested names are dropped
   with `mcp_env_passthrough_dropped` naming the variables. Explicit `null`
   stays unbounded under both. Reserved names stay excluded.
2. **Allowlist warning**: `mcp_package_allowlist_empty` emitted by `profile
   install`, `profile update` and `env status` whenever the effective
   `mcp_package_allowlist` is empty, stating that every declaration package
   in the closure is admitted.
3. **Surfacing** (§2.3): print the closed-column rows at `profile install`
   and `profile update` after the audit gate and before lock publication /
   (re-)materialization, in ascending package-name byte order, exact
   formatting (compact JSON arrays, quoted strings, `-`/`[]` for http, one
   LF); `env status` repeats the rows for the current profile of each
   reported scope; no rows when the MCP set is empty; never a diagnostic.
4. **Posture** (§12): `env status` reports the active S4 profile with the
   effective `passable_env_names`, the allowlist-empty warning row when
   applicable, and the surfacing rows; warnings never make a row non-current.
5. **Vectors/tests**: a Go test executes every case of
   `environments-env-passthrough.json` from `CURATOR_CONFORMANCE_ROOT`
   (default resolution under both profiles, allowlist warning, surfacing
   bytes byte-exact incl. the space-containing argument, install/update
   order) with the root-content skip and ledger row; unit tests for the
   drop/warn paths and the AC's "install output lists stdio commands;
   warnings implemented".
6. `CHANGELOG.md` Unreleased: "S4: … warning release (`s4-warn`) …" with
   the migration hint; note `s4-enforce` follows in a later release.

## Out of scope
The closed interpreter contract for MCP launch (later revision), E3 codex
seed (`STORY-260916-1i1gfo`), S1/S3 hardened defaults, SPEC_PIN, spec edits.

## Handoff
Narrow validation per the rules (`go build`, `go vet`, `gofmt -l`, `go test`
for `./internal/config/... ./internal/envfragment/... ./internal/envprofile/...
./internal/contextmaterialize/... ./cmd/curator/...` with the conformance root
set). Attach `TASK-260910-gocke2_results.md`, tick the checklist, then
`task-board handoff TASK-260910-gocke2 --role developer`.
