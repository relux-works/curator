# Environment machine configuration

Operator reference for the `environments` knobs of `manager-config`
schema 2, including permission mode (environments §10.1/§10.2/§12.1/§12.2),
the transitive system-module policy (E2), provider trust roots (E4), and
the MCP environment passthrough bound (S4). Knob names, values, and
diagnostics follow `protocol/environments.md` and manager profile §1. The
historical audit narrative lives in
[docs/security-audit-2026-09.md](security-audit-2026-09.md) and stays
history; this file is the current usage reference.

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
    "passable_env_names": ["FIGMA_API_KEY"],
    "permissions": {"companyA": "yolo"}
  }
}
```

`env status` reports the effective state of every knob below.

## Security posture

Schema-2 machine configurations default to `security_posture: hardened`
(revision B). Revision A shipped its migration warning in v0.15.0-rc.3.
Hardened defaults select strict audit and registry policy, reject unwaived
transitive system modules, and require source signers. An empty source
allowlist refuses install/update; an empty MCP allowlist refuses profile
composition when MCP declarations are present. Configure the required
allowlists and signer keys before those operations.

An explicit `security_posture: permissive` retains the former per-knob
defaults and emits `security_posture_permissive` once per manager operation.
Explicit knobs and locked system values still override posture defaults.
Schema-1 machines retain their frozen permissive behavior.

## Permission mode

`permissions` maps profile names to `native` or `yolo`. An absent profile
entry is silent; it leaves the launcher's own global default available.
For example:

```json
"permissions": {
  "interactive": "yolo",
  "restricted": "native"
}
```

`env resolve --format json` emits `launch-env-fragment-v2` and its required
closed `permissions` member. A configured profile entry is emitted with
`source: "profile"` and `locked: false`. An absent entry is emitted as
`mode: "native"`, `source: "default"`, and `locked: false`; that native mode
is a silence placeholder, not an explicit profile override.

`curator env config show` adds a `permissions_posture` object keyed by
installed profile, alongside the unchanged configured `permissions` map.
`curator env config show permissions` prints the effective `{mode, locked,
source}` object for the machine's current profile. Both forms use JSON and
the fragment's permission resolver, without requiring a runtime adapter.
Configuration or profile-state read failures exit non-zero with a diagnostic
and no posture output, never an unlocked fallback; a missing current profile
makes the single-knob form unavailable. See [CLI usage](cli.md#curator-env-config-show).

A system configuration may lock `environments.permissions` only toward
`native`. The system file's map replaces the machine map whole, and a
differing machine value produces the manager §1 warning naming the system
file. While the lock is engaged, `env resolve` emits
`mode: "native"`, `source: "global"`, and `locked: true` for the resolved
profile, including when that profile has no entry in the system map. The
launcher receives lock engagement in the fragment; this Curator setting
does not write launcher configuration.

```json
{
  "schema_version": 2,
  "locked": ["environments.permissions"],
  "environments": {"permissions": {}}
}
```
## Logging in to managed homes

Curator never copies, stores, or reads login tokens. A managed home gets its
login in one of two ways, chosen per profile × environment with
`curator env config set isolation.<profile>.<env-id> shared|isolated`:

- **`shared`**: the home reuses your native login through a link that Curator
  re-checks on every resolve. You log in once, natively.
- **`isolated`**: the home has no link and logs in by itself. You log in once
  **per home**, that is once per profile × environment, and every later launch
  into that home, including spawned agents, reuses that login.

| Environment | Default | `shared` | `isolated` |
|---|---|---|---|
| `claude_code` on macOS | `isolated` | refused with `environment_shared_unsupported`: Claude Code keys its Keychain item to `CLAUDE_CONFIG_DIR`, so there is nothing to link without copying tokens | one login per home |
| `claude_code` on Linux | `shared` | `.credentials.json` links to the native file | one login per home |
| `codex_cli` | `shared` | `auth.json` links to the native file under `file` storage (the default when `cli_auth_credentials_store` is absent); under `keyring` or `auto` the login is ambient | admitted only under `file` storage; otherwise `environment_isolated_unsupported` |
| `pi` | `shared` | `auth.json` links to the native `~/.pi/agent/auth.json` | one login per home |
| `muse` | `shared` | `config/muse/auth.json` links to the native file | refused with `environment_isolated_unsupported` until a fresh-login probe verifies it |
| `opencode` | `shared` (ambient) | its login lives outside the swapped config home, so nothing is linked | not applicable: refused, since the login is outside the home |

For `codex_cli`, sharing follows the storage selected by your **native**
`config.toml`. A managed `config.toml` that selects differently changes
nothing, and Curator never rewrites either file to align them.

**Claude Code on macOS needs one login per home.** A new or never-used home is
not logged in, and a launch into it, interactive or spawned, stops with
Claude Code's "Not logged in" message. Log in once before the first spawn:

```bash
eval "$(curator env resolve claude_code --profile <name> --format shell)"
claude auth login
```

`--format shell` prints the `CLAUDE_CONFIG_DIR` export for that home. Each
profile has its own home, so each profile you use with Claude Code needs its
own login.

**Log in from the same kind of session the launches will run in.** Where
Claude Code stores the login depends on whether the session can use the login
Keychain, and the two stores do not see each other (observed on Claude Code
2.1.299, macOS 15):

| Session the login runs in | Login stored in | Visible to launches from |
|---|---|---|
| A session that can use the login Keychain: the desktop session and everything started from it, including tmux servers and daemons it launched | the home's Keychain item `Claude Code-credentials-<hash>` | sessions that can use the Keychain |
| A session that cannot: ssh, and anything started from an ssh session; `security show-keychain-info` exits 36 there | `<home>/.credentials.json`, mode 0600 | any session of that user |
| An OS user with no login Keychain | nothing: the login fails with "Couldn't save your login" | none |

A home logged in from a desktop terminal reports "Not logged in" to a launch
from an ssh session, and the other way round. A daemon keeps the Keychain
access of the session that started it, so log a home in from the same kind of
session as the launcher that will use it; `security show-keychain-info` tells
the two apart. An OS user without a login Keychain cannot log in
this way at all; such launches need a token supplied in the environment
(`CLAUDE_CODE_OAUTH_TOKEN`, from `claude setup-token`), which writes nothing
to the home or the Keychain.

## Managed-home credential records

Managed-home `.agent-environment.json` markers use schema 2 for new
provisions. Each recorded credential strategy carries its effective
isolation, strategy, byte owner, backend, verified backend release, and the
mutation that last wrote the record. A record with an ambient backend or
strategy, or an isolated per-home Keychain strategy, has no `path`; other
records carry their home-relative credential path. This lets linkless
credentials remain visible without treating them as filesystem entries.

Existing schema-1 markers remain byte-for-byte untouched when the only
change would add schema-2 credential metadata. If another successful
mutation already needs to replace a schema-1 marker, Curator publishes the
complete schema-2 record in that same manager-home transaction. Marker
publication runs under the manager-home lock with journaled sibling staging
and atomic rename, so an interrupted write is recovered on the next
operation.

## Transitive system modules (E2)

`transitive_system_modules` is exactly `drop` or `error`. Its default is
`error` under hardened posture and `drop` under permissive posture.
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
| `transitive_system_modules` | hardened: `error`; permissive: `drop` | n/a (string knob) | rejected: must be `drop` or `error` |
| `system_module_waivers` | `[]` | `[]` | rejected: must be a list |
| `provider_directories` | `[]` | `[]` | rejected: must be a list of strings |
| `passable_env_names` | s4-warn: unbounded + warning; s4-enforce: `[]` | `[]` (nothing passes) | hardened: refused; permissive: unbounded, silent |

Lockable subset (§12.2): `transitive_system_modules` (only toward
`error`), `provider_directories`, and `passable_env_names`.
`system_module_waivers` MUST NOT be lockable.
