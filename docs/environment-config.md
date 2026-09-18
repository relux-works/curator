# Environment machine configuration

Operator reference for the `environments` knobs of `manager-config`
schema 2: the transitive system-module policy (E2), the provider trust
roots (E4), and the MCP environment passthrough bound (S4). Knob names,
values, and diagnostics are spelled exactly as `protocol/environments.md`
at the pinned conformance revision spells them. The historical audit
narrative lives in [docs/security-audit-2026-09.md](security-audit-2026-09.md)
and stays history; this file is the current usage reference.

All three knobs sit under one `environments` object:

```json
{
  "schema_version": 2,
  "skills_root": "/srv/curator/skills",
  "projects": {},
  "environments": {
    "transitive_system_modules": "drop",
    "system_module_waivers": [],
    "provider_directories": [],
    "passable_env_names": ["FIGMA_API_KEY"]
  }
}
```

`env status` reports the effective state of every knob below.

## Transitive system modules (E2)

`transitive_system_modules` is exactly `drop` (default) or `error`.
A package is **direct** when it is the root, an active overlay, or named
by the root's or an active overlay's `requires.contexts`; every other
`context` member is **transitive**. Only the `class: system` modules of
direct packages, and of transitive packages admitted by a waiver,
reach the system-prompt output and the fragment `system_prompt` section.

- Under `drop`, a non-admitted system module is skipped at
  materialization with the warning `context_system_module_dropped`
  (names the package and the module path). Install, update, and
  resolution never fail for admission.
- Under `error`, the first such module in emitted order fails
  resolution with `context_system_module_transitive` (names the package
  and the module path): `profile install` fails and `profile update`
  leaves the old lock in place.

The `context-system-module-present` finding is unaffected by the policy:
it reports every `class: system` module of every member at install and
update, admitted or not.

`system_module_waivers` admits transitive packages' system modules by
name. Each entry carries `package` (a portable identifier naming a
`context` member of the lock) and `reason` (free text recording why the
operator admits it). An entry naming no lock member has no effect.

```json
"system_module_waivers": [
  {"package": "platform-base", "reason": "fleet-owned base prompt, reviewed 2026-09"}
]
```

Fleet rules (§12.2): a system file locks `transitive_system_modules`
only toward `error`; `system_module_waivers` is not lockable: a lock
MUST NOT admit a transitive package's system modules.

## Provider trust roots (E4)

An unknown CLI subcommand dispatches to a `curator-<name>` executable.
The provider trust roots, in search order, are exactly the install
directory (the directory holding the running manager executable) and the
machine-configuration `provider_directories` list in listed order. Every
entry MUST be an absolute path (POSIX-absolute or Windows
drive-absolute) and SHOULD name a directory only the operator
administers. The ambient `PATH` is not a trust root.

This release ships **revision A** (the warning release): the manager
searches the ambient `PATH` as before. A `PATH`-selected provider inside
a trust root resolves silently; one outside the trust roots still
resolves but warns `subcommand_provider_outside_trust_roots`, naming the
resolved path, the trust roots consulted, and the migration hint: list
the provider's directory in `provider_directories`. **Revision B** (a
later release) removes `PATH` selection: trust-root matches resolve,
`PATH`-only providers are refused with
`subcommand_provider_untrusted`.

Refusals and failures, both revisions:

- A provider inside a directory the manager itself publishes onto
  `PATH` (the user-bin shim directory, a managed skill bin directory,
  or below the environments root) is refused with
  `subcommand_provider_untrusted`, wherever else it was found.
- A trust root the manager cannot read fails the lookup with
  `subcommand_provider_root_unreadable`, naming the first unreadable
  root: unreadable is never absence, and nothing falls through to a
  later root, to `PATH`, or to missing.
- No match is `subcommand_provider_missing`, naming the trust roots
  consulted.

Migration: when the warning names a provider directory the operator
trusts, add it to `provider_directories`; the warning goes silent and
the provider keeps resolving after the revision-B flip. `env status`
reports the resolved absolute path and trust verdict per provider
(always for `curator-run` and `curator-session`); under revision A the
outside-trust-roots warning row stays current, while a refused,
missing, or unreadable provider row is non-current under `env status
--check`.

## MCP environment passthrough (S4)

`passable_env_names` bounds which operator environment variables an MCP
server launch may receive: the declaration's `env_names` name variables,
never values, and only a listed name's operator value may reach a
launch. Values never appear in any package, lock, marker, fragment, or
materialized file.

This release ships profile **`s4-warn`** (the warning release):

- An absent knob behaves as unbounded (pre-S4 behavior is kept), but
  every launch passing an operator variable warns
  `mcp_env_passthrough_unlisted`, naming the variables and the knob
  with the migration hint: list the named variables to keep passing
  them after the flip. With a configured list, the warning fires per
  passed variable outside it.
- An explicit `null` passes unbounded with no warning (an explicit,
  lockable-away operator choice).

Profile **`s4-enforce`** (the flip release, a later release) makes an
absent knob the empty list: a requested name outside the effective list
is dropped with `mcp_env_passthrough_dropped`, naming the dropped
variables. An explicit `null` stays unbounded with no diagnostic, and a
system file MAY lock `passable_env_names` to a list so that `null` is
unavailable.

Migration: read the `mcp_env_passthrough_unlisted` warnings, add the
named variables to `passable_env_names`, and re-run until the warnings
stop; the same list then bounds the launch after the flip. `env status`
reports the active profile with the effective allowlist.

Declaration visibility: at `profile install` and `profile update` the
manager prints one `mcp-declaration` surfacing row per MCP declaration
package after the audit gate passes and before the lock is published or
any surface is (re-)materialized; `env status` repeats the rows for the
current profile of each scope it reports. Surfacing is informative: it
emits no diagnostic and never fails the operation. Separately,
`profile install`, `profile update`, and `env status` MUST emit
`mcp_package_allowlist_empty`, stating that every declaration package
in the closure is admitted, whenever the MCP package allowlist is
empty.

## Absent versus null

Only absence takes a knob's default. `passable_env_names` is the one
knob of the three with a `null` meaning:

| Knob | Absent | Empty list | Explicit `null` |
|---|---|---|---|
| `transitive_system_modules` | `drop` | n/a (string knob) | rejected: must be `drop` or `error` |
| `system_module_waivers` | `[]` | `[]` | rejected: must be a list |
| `provider_directories` | `[]` | `[]` | rejected: must be a list of strings |
| `passable_env_names` | s4-warn: unbounded + warning; s4-enforce: `[]` | `[]` (nothing passes) | unbounded, silent |

Lockable subset (§12.2): `transitive_system_modules` (only toward
`error`), `provider_directories`, and `passable_env_names`.
`system_module_waivers` MUST NOT be lockable.
