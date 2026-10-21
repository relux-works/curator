# CLI help and bounded probe evidence

Task: TASK-261021-i2f2uq. Review date: 2026-10-21.

Curator source: cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7. Launcher review build: dd70beb42239fc80f81bcfd512358f8d3844b615. Commands below are read-only help or rejected-syntax probes. Launcher user policy was redirected to task-local fixtures. This does not prove absence of all filesystem effects, cross-platform behavior, or hosted launch behavior.

## Help captures

### curator --version

Exit: 0

```text
curator v0.15.0-rc.7
```

### curator help

Exit: 0

```text
curator: agent environment manager

Usage:
  curator <command> [arguments]

Commands:
  bootstrap [flags]         create the machine configuration
  init [path]              create Skillfile.json and the managed gitignore block
  add <name> ...           add or replace a skill declaration, then install
  remove <name>            remove a skill declaration
  install [path] [flags]   apply Skillfile.json (see install -h)
  update                   fetch all source repositories under skills_root
  upgrade [path]           fetch the selected dependency closure, then install
  status [path] [flags]    manifest, installed, and compiled state (--check, --json, --attest)
  list                     configured projects and declared skills
  project <subcommand>     add | resolve | refresh
  skill check <dir>        validate one skill package (--locale, --json)
  global <subcommand>      init | add | remove | list | status (--check, --json) | install | adopt <command> [--dry-run] | update | upgrade
  profile <subcommand>     install | list | use | update | remove | sync | compose (see profile install -h)
  env <subcommand>         resolve | status | config | migrate | unmanage (see env resolve -h)
  run <env-id> ...         umbrella dispatch to curator-run (see §11)
  hybrid <subcommand>      add | remove | list | status
  audit [target] [flags]   run audit, pin trust, or publish a signed record
  gc                       remove unreferenced runtime entries
  shell-init [shell]       print or cache an optional hook (auto, zsh, bash, powershell)
  hook <subcommand>        approve | approvals | revoke (project env file trust)
  ui                       terminal view over installed state
  config <subcommand>      show | build-ssh | build-https (see config build-ssh -h, config build-https -h)
  --version                print the curator version

Environment operands accept claude and codex as aliases of claude_code and
codex_cli; every output keeps the canonical id.
```

### curator bootstrap --help

Exit: 0

```text
usage: curator bootstrap [--skills-root <dir>] [--default-agents <ids>] [--preferred-locale <locale>] [--force] [--if-missing] [--non-interactive]

Create the machine configuration and default directories.

```

### curator init --help

Exit: 0

```text
usage: curator init [path]

Create Skillfile.json and the managed .gitignore block in the project directory.

```

### curator add --help

Exit: 0

```text
usage: curator add <name> [--git <url>] [--branch <ref>] [--tag <ref>] [--revision <ref>] [--source <dir>] [--project <path>]

Add or replace a skill declaration, then install it.

```

### curator remove --help

Exit: 0

```text
usage: curator remove <name> [path] [--project <path>]

Remove a skill declaration.

```

### curator install --help

Exit: 0

```text
usage: curator install [path] [--all] [--dry-run] [--fix-gitignore] [--strict-tags] [--verbose] [--audit[=advisory|strict]] [--build-ssh-identity <path>] [--build-ssh-agent <socket>] [--build-ssh-known-hosts <path>]

Apply Skillfile.json to the selected project.

Skillfile schema 2 project sources:
  Schema 2 source declarations resolve with the project commands below.
  Skillfile schema 1 retains its exact meaning; no on-disk migration is implicit.

  curator project resolve [path]   freeze declared sources into Skillfile.lock.json
  curator install [path]           materialize the locked snapshot (repairs drift)
  curator status [path]            compare installed state against the lock (--check, --json)
  curator project refresh [path]   re-resolve refs, bytes, and membership explicitly
  curator install [path]           install the refreshed lock (refresh alone changes nothing live)

  Launch and status never rescan collections, advance branches, or
  replace a local snapshot: installed shims execute the frozen runtime
  until an explicit refresh plus install republishes it. A Skillfile
  edited after resolve fails source_lock_stale until refresh; run
  install again after refresh to repair drifted bytes from the lock.
```

### curator update --help

Exit: 0

```text
usage: curator update

Fetch all source repositories under skills_root.

```

### curator upgrade --help

Exit: 0

```text
usage: curator upgrade [path] [--all] [--dry-run] [--fix-gitignore] [--strict-tags] [--verbose] [--audit[=advisory|strict]] [--build-ssh-identity <path>] [--build-ssh-agent <socket>] [--build-ssh-known-hosts <path>]

Fetch the selected dependency closure, then install it.

Skillfile schema 2 project sources:
  Schema 2 source declarations resolve with the project commands below.
  Skillfile schema 1 retains its exact meaning; no on-disk migration is implicit.

  curator project resolve [path]   freeze declared sources into Skillfile.lock.json
  curator install [path]           materialize the locked snapshot (repairs drift)
  curator status [path]            compare installed state against the lock (--check, --json)
  curator project refresh [path]   re-resolve refs, bytes, and membership explicitly
  curator install [path]           install the refreshed lock (refresh alone changes nothing live)

  Launch and status never rescan collections, advance branches, or
  replace a local snapshot: installed shims execute the frozen runtime
  until an explicit refresh plus install republishes it. A Skillfile
  edited after resolve fails source_lock_stale until refresh; run
  install again after refresh to repair drifted bytes from the lock.
```

### curator status --help

Exit: 0

```text
usage: curator status [path] [--all] [--check] [--json] [--attest]

Report manifest, installed, and compiled state.

Skillfile schema 2 project sources:
  Schema 2 source declarations resolve with the project commands below.
  Skillfile schema 1 retains its exact meaning; no on-disk migration is implicit.

  curator project resolve [path]   freeze declared sources into Skillfile.lock.json
  curator install [path]           materialize the locked snapshot (repairs drift)
  curator status [path]            compare installed state against the lock (--check, --json)
  curator project refresh [path]   re-resolve refs, bytes, and membership explicitly
  curator install [path]           install the refreshed lock (refresh alone changes nothing live)

  Launch and status never rescan collections, advance branches, or
  replace a local snapshot: installed shims execute the frozen runtime
  until an explicit refresh plus install republishes it. A Skillfile
  edited after resolve fails source_lock_stale until refresh; run
  install again after refresh to repair drifted bytes from the lock.
```

### curator list --help

Exit: 0

```text
usage: curator list

List configured projects and declared skills.

```

### curator project --help

Exit: 0

```text
usage: curator project <add|resolve|refresh> [arguments]

Register projects or resolve their dependency closures.

```

### curator project add --help

Exit: 0

```text
usage: curator project add <alias> <path> [--agents <ids>]

Register a project and initialize its declarative files.

```

### curator project resolve --help

Exit: 0

```text
usage: curator project resolve [path]

Resolve transitive dependencies for a project closure.

curator project resolve: resolve transitive dependencies for a project closure

Usage:
  curator project resolve [path]
  curator project refresh [path]

Schema 1: prints the alias, project path, Skillfile path, and managed
skill and bin directories without modifying disk state.

Skillfile schema 2 project sources:
  Schema 2 source declarations resolve with the project commands below.
  Skillfile schema 1 retains its exact meaning; no on-disk migration is implicit.

  curator project resolve [path]   freeze declared sources into Skillfile.lock.json
  curator install [path]           materialize the locked snapshot (repairs drift)
  curator status [path]            compare installed state against the lock (--check, --json)
  curator project refresh [path]   re-resolve refs, bytes, and membership explicitly
  curator install [path]           install the refreshed lock (refresh alone changes nothing live)

  Launch and status never rescan collections, advance branches, or
  replace a local snapshot: installed shims execute the frozen runtime
  until an explicit refresh plus install republishes it. A Skillfile
  edited after resolve fails source_lock_stale until refresh; run
  install again after refresh to repair drifted bytes from the lock.

Schema 2 source examples (one alias or selector value per line):
  {"path": "./agents"}              local directory, relative to Skillfile.json
  {"path": "/work/shared-agents"}   local directory, absolute
  {"git": "https://example.org/kit.git", "tag": "v1.2.0"}   Git source at a tag
  {"repository": "example.org/kit", "branch": "main"}       logical Git source (needs machine policy)
  {"name": "review", "from": "team", "directory": "skills/review"}   one skill
  {"from": "team", "directory": "skills", "include": ["review", "docs"]}   collection
  {"from": "team", "directory": "skills", "include": ["*"]}   whole directory
  {"from": "team", "directory": "skills", "include": ["*"], "exclude": ["release"]}   minus one

  Machine policy (operator-owned, beside the manager configuration):
  source-policy.json maps canonical host/path entries to one or two
  endpoints plus fallback, and admits root packages via "root_inputs".
  See "Skillfile schema 2 project sources" in docs/cli.md.
```

### curator project refresh --help

Exit: 0

```text
usage: curator project refresh [path]

Refresh the frozen project source lock.

curator project resolve: resolve transitive dependencies for a project closure

Usage:
  curator project resolve [path]
  curator project refresh [path]

Schema 1: prints the alias, project path, Skillfile path, and managed
skill and bin directories without modifying disk state.

Skillfile schema 2 project sources:
  Schema 2 source declarations resolve with the project commands below.
  Skillfile schema 1 retains its exact meaning; no on-disk migration is implicit.

  curator project resolve [path]   freeze declared sources into Skillfile.lock.json
  curator install [path]           materialize the locked snapshot (repairs drift)
  curator status [path]            compare installed state against the lock (--check, --json)
  curator project refresh [path]   re-resolve refs, bytes, and membership explicitly
  curator install [path]           install the refreshed lock (refresh alone changes nothing live)

  Launch and status never rescan collections, advance branches, or
  replace a local snapshot: installed shims execute the frozen runtime
  until an explicit refresh plus install republishes it. A Skillfile
  edited after resolve fails source_lock_stale until refresh; run
  install again after refresh to repair drifted bytes from the lock.

Schema 2 source examples (one alias or selector value per line):
  {"path": "./agents"}              local directory, relative to Skillfile.json
  {"path": "/work/shared-agents"}   local directory, absolute
  {"git": "https://example.org/kit.git", "tag": "v1.2.0"}   Git source at a tag
  {"repository": "example.org/kit", "branch": "main"}       logical Git source (needs machine policy)
  {"name": "review", "from": "team", "directory": "skills/review"}   one skill
  {"from": "team", "directory": "skills", "include": ["review", "docs"]}   collection
  {"from": "team", "directory": "skills", "include": ["*"]}   whole directory
  {"from": "team", "directory": "skills", "include": ["*"], "exclude": ["release"]}   minus one

  Machine policy (operator-owned, beside the manager configuration):
  source-policy.json maps canonical host/path entries to one or two
  endpoints plus fallback, and admits root packages via "root_inputs".
  See "Skillfile schema 2 project sources" in docs/cli.md.
```

### curator skill --help

Exit: 0

```text
usage: curator skill check <dir> [--locale <locale>] [--json]

Validate a skill package.

```

### curator skill check --help

Exit: 0

```text
usage: curator skill check <dir> [--locale <locale>] [--json]

Validate one skill package.

```

### curator global --help

Exit: 0

```text
usage: curator global <init|add|remove|list|status|install|adopt|update|upgrade> [arguments]

Manage machine-wide skill installations.

usage: curator global <init|add|remove|list|status|install|adopt|update|upgrade> [flags]

Manage machine-wide skill installations. Requires the global config:
run `curator bootstrap --skills-root <dir>` first (see docs/cli.md#bootstrap).
```

### curator global init --help

Exit: 0

```text
usage: curator global init

Create the global Skillfile.json and managed directories.

```

### curator global add --help

Exit: 0

```text
usage: curator global add <name> [--git <url>] [--branch <ref>] [--tag <ref>] [--revision <ref>] [--source <dir>]

Add a global skill declaration and install it.

```

### curator global remove --help

Exit: 0

```text
usage: curator global remove <name>

Remove a global skill declaration.

```

### curator global list --help

Exit: 0

```text
usage: curator global list

List global skill declarations.

```

### curator global status --help

Exit: 0

```text
usage: curator global status [--check] [--json]

Report global installed and compiled state.

```

### curator global install --help

Exit: 0

```text
usage: curator global install [--dry-run] [--audit[=advisory|strict]]

Install global skill declarations.

```

### curator global adopt --help

Exit: 0

```text
usage: curator global adopt <command> [--dry-run]

Adopt an existing global command.

```

### curator global update --help

Exit: 0

```text
usage: curator global update

Fetch global source repositories.

```

### curator global upgrade --help

Exit: 0

```text
usage: curator global upgrade [--dry-run] [--audit[=advisory|strict]]

Fetch the global dependency closure and install it.

```

### curator profile --help

Exit: 0

```text
usage: curator profile <install|import|list|use|update|remove|sync|compose> [arguments]

Manage environment profiles.

Requires the global config: run `curator bootstrap --skills-root <dir>` first (see docs/cli.md#bootstrap).
```

### curator profile install --help

Exit: 0

```text
usage: curator profile install <source> [--directory <dir>] [--range <range>|--tag <tag>|--revision <ref>] [--as <name>] [--use] [--takeover] [--confirm-system-delta]

Install an environment profile from a source package.

```

### curator profile import --help

Exit: 0

```text
usage: curator profile import [--as <name>] [--allow-lossy] [--use]

Import the managed home as an environment profile.

```

### curator profile list --help

Exit: 0

```text
usage: curator profile list

List installed environment profiles.

```

### curator profile use --help

Exit: 0

```text
usage: curator profile use <name> [--env <env-id>] [--target <target-id>] [--clear] [--takeover]

Activate or clear an environment profile.

curator: <env-id> accepts claude and codex as aliases of claude_code and codex_cli; outputs keep the canonical id
```

### curator profile update --help

Exit: 0

```text
usage: curator profile update [<name>] [--all] [--takeover] [--confirm-system-delta]

Update installed environment profiles.

```

### curator profile remove --help

Exit: 0

```text
usage: curator profile remove <name> [--purge]

Remove an installed environment profile.

```

### curator profile sync --help

Exit: 0

```text
usage: curator profile sync [--takeover]

Synchronize active environment profiles.

```

### curator profile compose --help

Exit: 0

```text
usage: curator profile compose <profile> <add|remove|list> [arguments]

Compose an environment profile from installed profiles.

```

### curator profile compose example add --help

Exit: 0

```text
usage: curator profile compose <profile> add <source> [--range <range>|--tag <tag>|--revision <ref>] [--directory <dir>] [--weight <n>]

Add an overlay to a composed profile.

```

### curator profile compose example remove --help

Exit: 0

```text
usage: curator profile compose <profile> remove <source>

Remove an overlay from a composed profile.

```

### curator profile compose example list --help

Exit: 0

```text
usage: curator profile compose <profile> list

List members of a composed profile.

```

### curator env --help

Exit: 0

```text
usage: curator env <resolve|status|config|migrate|unmanage> [arguments]

Inspect and manage environment homes.

Requires the global config: run `curator bootstrap --skills-root <dir>` first (see docs/cli.md#bootstrap).
```

### curator env resolve --help

Exit: 0

```text
usage: curator env resolve <env-id> [--profile <name>] [--repair [--preserve-lock] [--dry-run]] [--takeover] [--format json|env|shell]

Resolve an environment fragment; --repair provisions or repairs the managed home.

curator: <env-id> accepts claude and codex as aliases of claude_code and codex_cli; outputs keep the canonical id
```

### curator env status --help

Exit: 0

```text
usage: curator env status [--check] [--json]

Report environment home state.

curator: <env-id> accepts claude and codex as aliases of claude_code and codex_cli; outputs keep the canonical id
```

### curator env config --help

Exit: 0

```text
usage: curator env config <show|set|unset> [<knob> [<value>]]

Inspect or edit machine environment configuration.

curator: environment positions in <knob> accept claude and codex as aliases of claude_code and codex_cli; the file keeps the canonical id
```

### curator env config show --help

Exit: 0

```text
usage: curator env config show [<knob>]

Print effective environment configuration.

curator: environment positions in <knob> accept claude and codex as aliases of claude_code and codex_cli; the file keeps the canonical id
```

### curator env config set --help

Exit: 0

```text
usage: curator env config set <knob> <value>

Set a machine environment configuration knob.

curator: environment positions in <knob> accept claude and codex as aliases of claude_code and codex_cli; the file keeps the canonical id
```

### curator env config unset --help

Exit: 0

```text
usage: curator env config unset <knob>

Unset a machine environment configuration knob.

curator: environment positions in <knob> accept claude and codex as aliases of claude_code and codex_cli; the file keeps the canonical id
```

### curator env migrate --help

Exit: 0

```text
usage: curator env migrate --inspect|--plan|--apply [--profile <name>] [--env <env-id>] [--expect <plan-hash>]

Migrate existing environment installations.

```

### curator env unmanage --help

Exit: 0

```text
usage: curator env unmanage [--restore-backups] [--env <env-id>] [--target <target-id>]

Restore files from managed home backups.

curator: <env-id> accepts claude and codex as aliases of claude_code and codex_cli; outputs keep the canonical id
```

### curator run --help

Exit: 0

```text
usage: curator run <env-id> [arguments] [-- <passthrough arguments>]

Dispatch to the installed curator-run provider; arguments after -- are passed unchanged.

curator: <env-id> accepts claude and codex as aliases of claude_code and codex_cli; outputs keep the canonical id
```

### curator hybrid --help

Exit: 0

```text
usage: curator hybrid <add|remove|list|status> [arguments]

Manage hybrid-scope skill declarations.

```

### curator hybrid add --help

Exit: 0

```text
usage: curator hybrid add <name> [--git <url>] [--branch <ref>|--tag <ref>|--revision <ref>] --target <target>|--targets <targets>

Add a hybrid skill declaration.

```

### curator hybrid remove --help

Exit: 0

```text
usage: curator hybrid remove <name>

Remove a hybrid skill declaration.

```

### curator hybrid list --help

Exit: 0

```text
usage: curator hybrid list

List hybrid skill declarations.

```

### curator hybrid status --help

Exit: 0

```text
usage: curator hybrid status

Report hybrid installed state.

```

### curator audit --help

Exit: 0

```text
usage: curator audit [target] [--all] [--global] [--json] [--allow <hash> --reason <reason>] [--publish <file> --registry <url> --token <token>]

Audit skills, pin trust, or publish a signed record.

```

### curator gc --help

Exit: 0

```text
usage: curator gc

Remove unreferenced runtime entries.

```

### curator shell-init --help

Exit: 0

```text
usage: curator shell-init [auto|zsh|bash|powershell] [--install] [--no-global]

Print or cache an optional shell hook.

```

### curator hook --help

Exit: 0

```text
usage: curator hook <approve|approvals|revoke> [arguments]

Manage trust approvals for project environment files.

```

### curator hook approve --help

Exit: 0

```text
usage: curator hook approve <path>

Approve a project environment file.

```

### curator hook approvals --help

Exit: 0

```text
usage: curator hook approvals

List recorded project environment file approvals.

```

### curator hook revoke --help

Exit: 0

```text
usage: curator hook revoke <path>

Revoke a project environment file approval.

```

### curator ui --help

Exit: 0

```text
usage: curator ui

Open the terminal view of installed state.

```

### curator config --help

Exit: 0

```text
usage: curator config <show|build-ssh|build-https> [arguments]

Inspect and edit the machine configuration.

curator config: inspect and edit the machine configuration

Usage:
  curator config show                  print the effective configuration
  curator config build-ssh <subcommand>
                                       operator SSH credentials for external
                                       build repositories (add | list | remove)
  curator config build-https <subcommand>
                                       operator HTTPS tokens for external
                                       build repositories (add | login | list | remove)
```

### curator config show --help

Exit: 0

```text
usage: curator config show

Print the effective machine configuration.

```

### curator config build-ssh --help

Exit: 0

```text
usage: curator config build-ssh <add|list|remove> [arguments]

Manage SSH credentials for external build repositories.

curator config build-ssh: operator SSH credentials for external build repositories

Usage:
  curator config build-ssh add <scope> [--agent [SOCKET]] [--identity PATH] [--known-hosts PATH]
  curator config build-ssh list
  curator config build-ssh remove <scope>

A scope is a lowercase host optionally followed by '/'-separated path segments.
It is matched against the canonical host/path identity of a build repository on
whole-segment boundaries, and the longest matching scope wins: scope
'git.example.com/portals' covers 'git.example.com/portals/app' and never
'git.example.com/portals-other'.

An entry names an agent, an identity file, or both:
  --agent                  the agent socket the operator's environment provides
  --agent SOCKET           a named agent socket
  --identity PATH          an identity file offered to the destination
  --agent --identity PATH  an agent pinned to that one identity
  --known-hosts PATH       a known-hosts file for this scope
Every path must be absolute or start with '~/'.

add replaces whatever was recorded under the same scope; remove reports a scope
that is not configured as an error; list prints one line per scope, sorted.

Precedence: command-line flags override CURATOR_BUILD_SSH_* environment values,
which override the scopes configured here. Credentials are operator-owned: no
manifest, descriptor, repository, substitution, or marker can select them.

An install that reaches an SSH build repository nothing covers stops before it
fetches anything. With a terminal attached it lists the agent and ~/.ssh/*.pub
files found on this host and asks which to record and under which scope; with
no terminal, and on a dry run, it prints those same candidates as ready-to-run
add commands and fails with build_repository_ssh_credential_missing.
```

### curator config build-ssh add --help

Exit: 0

```text
usage: curator config build-ssh add <scope> [--agent [SOCKET]] [--identity PATH] [--known-hosts PATH]

Add or replace SSH credentials for a repository scope.

curator config build-ssh: operator SSH credentials for external build repositories

Usage:
  curator config build-ssh add <scope> [--agent [SOCKET]] [--identity PATH] [--known-hosts PATH]
  curator config build-ssh list
  curator config build-ssh remove <scope>

A scope is a lowercase host optionally followed by '/'-separated path segments.
It is matched against the canonical host/path identity of a build repository on
whole-segment boundaries, and the longest matching scope wins: scope
'git.example.com/portals' covers 'git.example.com/portals/app' and never
'git.example.com/portals-other'.

An entry names an agent, an identity file, or both:
  --agent                  the agent socket the operator's environment provides
  --agent SOCKET           a named agent socket
  --identity PATH          an identity file offered to the destination
  --agent --identity PATH  an agent pinned to that one identity
  --known-hosts PATH       a known-hosts file for this scope
Every path must be absolute or start with '~/'.

add replaces whatever was recorded under the same scope; remove reports a scope
that is not configured as an error; list prints one line per scope, sorted.

Precedence: command-line flags override CURATOR_BUILD_SSH_* environment values,
which override the scopes configured here. Credentials are operator-owned: no
manifest, descriptor, repository, substitution, or marker can select them.

An install that reaches an SSH build repository nothing covers stops before it
fetches anything. With a terminal attached it lists the agent and ~/.ssh/*.pub
files found on this host and asks which to record and under which scope; with
no terminal, and on a dry run, it prints those same candidates as ready-to-run
add commands and fails with build_repository_ssh_credential_missing.
```

### curator config build-ssh list --help

Exit: 0

```text
usage: curator config build-ssh list

List configured SSH credential scopes.

curator config build-ssh: operator SSH credentials for external build repositories

Usage:
  curator config build-ssh add <scope> [--agent [SOCKET]] [--identity PATH] [--known-hosts PATH]
  curator config build-ssh list
  curator config build-ssh remove <scope>

A scope is a lowercase host optionally followed by '/'-separated path segments.
It is matched against the canonical host/path identity of a build repository on
whole-segment boundaries, and the longest matching scope wins: scope
'git.example.com/portals' covers 'git.example.com/portals/app' and never
'git.example.com/portals-other'.

An entry names an agent, an identity file, or both:
  --agent                  the agent socket the operator's environment provides
  --agent SOCKET           a named agent socket
  --identity PATH          an identity file offered to the destination
  --agent --identity PATH  an agent pinned to that one identity
  --known-hosts PATH       a known-hosts file for this scope
Every path must be absolute or start with '~/'.

add replaces whatever was recorded under the same scope; remove reports a scope
that is not configured as an error; list prints one line per scope, sorted.

Precedence: command-line flags override CURATOR_BUILD_SSH_* environment values,
which override the scopes configured here. Credentials are operator-owned: no
manifest, descriptor, repository, substitution, or marker can select them.

An install that reaches an SSH build repository nothing covers stops before it
fetches anything. With a terminal attached it lists the agent and ~/.ssh/*.pub
files found on this host and asks which to record and under which scope; with
no terminal, and on a dry run, it prints those same candidates as ready-to-run
add commands and fails with build_repository_ssh_credential_missing.
```

### curator config build-ssh remove --help

Exit: 0

```text
usage: curator config build-ssh remove <scope>

Remove an SSH credential scope.

curator config build-ssh: operator SSH credentials for external build repositories

Usage:
  curator config build-ssh add <scope> [--agent [SOCKET]] [--identity PATH] [--known-hosts PATH]
  curator config build-ssh list
  curator config build-ssh remove <scope>

A scope is a lowercase host optionally followed by '/'-separated path segments.
It is matched against the canonical host/path identity of a build repository on
whole-segment boundaries, and the longest matching scope wins: scope
'git.example.com/portals' covers 'git.example.com/portals/app' and never
'git.example.com/portals-other'.

An entry names an agent, an identity file, or both:
  --agent                  the agent socket the operator's environment provides
  --agent SOCKET           a named agent socket
  --identity PATH          an identity file offered to the destination
  --agent --identity PATH  an agent pinned to that one identity
  --known-hosts PATH       a known-hosts file for this scope
Every path must be absolute or start with '~/'.

add replaces whatever was recorded under the same scope; remove reports a scope
that is not configured as an error; list prints one line per scope, sorted.

Precedence: command-line flags override CURATOR_BUILD_SSH_* environment values,
which override the scopes configured here. Credentials are operator-owned: no
manifest, descriptor, repository, substitution, or marker can select them.

An install that reaches an SSH build repository nothing covers stops before it
fetches anything. With a terminal attached it lists the agent and ~/.ssh/*.pub
files found on this host and asks which to record and under which scope; with
no terminal, and on a dry run, it prints those same candidates as ready-to-run
add commands and fails with build_repository_ssh_credential_missing.
```

### curator config build-https --help

Exit: 0

```text
usage: curator config build-https <add|login|list|remove> [arguments]

Manage HTTPS credentials for external build repositories.

curator config build-https: operator HTTPS tokens for external build repositories

Usage:
  curator config build-https add <scope> (--git-credentials | --keyring | --token-env NAME) [--username NAME]
  curator config build-https login <scope> [--username NAME]
  curator config build-https list
  curator config build-https remove <scope>

A scope is a lowercase host optionally followed by '/'-separated path segments.
It is matched against the canonical host/path identity of a build repository on
whole-segment boundaries, and the longest matching scope wins: scope
'git.example.com/portals' covers 'git.example.com/portals/app' and never
'git.example.com/portals-other'.

A token is never accepted as a command-line argument, and the config never
stores one as a literal value. add names where the token is read from:
  --git-credentials  the operator's own Git HTTPS credential for the scope's
                      host, read through 'git credential fill'
  --keyring           the token already stored for this scope by login
  --token-env NAME    an environment variable read at process entry
  --username NAME     sent alongside the resolved token; defaults to "token"

login reads a token from a hidden prompt, or one line from stdin when it is
not attached to a terminal, stores it through the operator's own Git
credential machinery under a scope-namespaced entry, and selects it for the
scope the same way 'add --keyring' would.

add and login replace whatever was recorded under the same scope; remove
reports a scope that is not configured as an error, and also deletes the
scope's stored token when it selected one; list prints one line per scope,
sorted, and reports whether the token it names actually resolves right now.

Precedence: CURATOR_BUILD_HTTPS_TOKEN, optionally pinned to one host by
CURATOR_BUILD_HTTPS_HOST, covers every repository on that host ahead of the
scopes configured here; unset, or once that host is exhausted, the longest
matching build_https scope applies; an HTTPS repository nothing covers is not
an error, unlike SSH — it fetches anonymously. Credentials are operator-owned:
no manifest, descriptor, repository, substitution, or marker can select them.

Disclosure warning (Spec core §12.2): CURATOR_BUILD_HTTPS_TOKEN without
CURATOR_BUILD_HTTPS_HOST is not bound to any one repository identity — it is
offered to every private HTTPS build repository host this run reaches. Set
CURATOR_BUILD_HTTPS_HOST to bind it to one host, or prefer a build_https
scope, unless every one of those hosts is meant to receive that token.
```

### curator config build-https add --help

Exit: 0

```text
usage: curator config build-https add <scope> [--git-credentials] [--keyring] [--token-env NAME] [--username <name>]

Configure HTTPS credentials for a repository scope.

curator config build-https: operator HTTPS tokens for external build repositories

Usage:
  curator config build-https add <scope> (--git-credentials | --keyring | --token-env NAME) [--username NAME]
  curator config build-https login <scope> [--username NAME]
  curator config build-https list
  curator config build-https remove <scope>

A scope is a lowercase host optionally followed by '/'-separated path segments.
It is matched against the canonical host/path identity of a build repository on
whole-segment boundaries, and the longest matching scope wins: scope
'git.example.com/portals' covers 'git.example.com/portals/app' and never
'git.example.com/portals-other'.

A token is never accepted as a command-line argument, and the config never
stores one as a literal value. add names where the token is read from:
  --git-credentials  the operator's own Git HTTPS credential for the scope's
                      host, read through 'git credential fill'
  --keyring           the token already stored for this scope by login
  --token-env NAME    an environment variable read at process entry
  --username NAME     sent alongside the resolved token; defaults to "token"

login reads a token from a hidden prompt, or one line from stdin when it is
not attached to a terminal, stores it through the operator's own Git
credential machinery under a scope-namespaced entry, and selects it for the
scope the same way 'add --keyring' would.

add and login replace whatever was recorded under the same scope; remove
reports a scope that is not configured as an error, and also deletes the
scope's stored token when it selected one; list prints one line per scope,
sorted, and reports whether the token it names actually resolves right now.

Precedence: CURATOR_BUILD_HTTPS_TOKEN, optionally pinned to one host by
CURATOR_BUILD_HTTPS_HOST, covers every repository on that host ahead of the
scopes configured here; unset, or once that host is exhausted, the longest
matching build_https scope applies; an HTTPS repository nothing covers is not
an error, unlike SSH — it fetches anonymously. Credentials are operator-owned:
no manifest, descriptor, repository, substitution, or marker can select them.

Disclosure warning (Spec core §12.2): CURATOR_BUILD_HTTPS_TOKEN without
CURATOR_BUILD_HTTPS_HOST is not bound to any one repository identity — it is
offered to every private HTTPS build repository host this run reaches. Set
CURATOR_BUILD_HTTPS_HOST to bind it to one host, or prefer a build_https
scope, unless every one of those hosts is meant to receive that token.
```

### curator config build-https login --help

Exit: 0

```text
usage: curator config build-https login <scope> [--username <name>]

Store HTTPS credentials interactively for a repository scope.

curator config build-https: operator HTTPS tokens for external build repositories

Usage:
  curator config build-https add <scope> (--git-credentials | --keyring | --token-env NAME) [--username NAME]
  curator config build-https login <scope> [--username NAME]
  curator config build-https list
  curator config build-https remove <scope>

A scope is a lowercase host optionally followed by '/'-separated path segments.
It is matched against the canonical host/path identity of a build repository on
whole-segment boundaries, and the longest matching scope wins: scope
'git.example.com/portals' covers 'git.example.com/portals/app' and never
'git.example.com/portals-other'.

A token is never accepted as a command-line argument, and the config never
stores one as a literal value. add names where the token is read from:
  --git-credentials  the operator's own Git HTTPS credential for the scope's
                      host, read through 'git credential fill'
  --keyring           the token already stored for this scope by login
  --token-env NAME    an environment variable read at process entry
  --username NAME     sent alongside the resolved token; defaults to "token"

login reads a token from a hidden prompt, or one line from stdin when it is
not attached to a terminal, stores it through the operator's own Git
credential machinery under a scope-namespaced entry, and selects it for the
scope the same way 'add --keyring' would.

add and login replace whatever was recorded under the same scope; remove
reports a scope that is not configured as an error, and also deletes the
scope's stored token when it selected one; list prints one line per scope,
sorted, and reports whether the token it names actually resolves right now.

Precedence: CURATOR_BUILD_HTTPS_TOKEN, optionally pinned to one host by
CURATOR_BUILD_HTTPS_HOST, covers every repository on that host ahead of the
scopes configured here; unset, or once that host is exhausted, the longest
matching build_https scope applies; an HTTPS repository nothing covers is not
an error, unlike SSH — it fetches anonymously. Credentials are operator-owned:
no manifest, descriptor, repository, substitution, or marker can select them.

Disclosure warning (Spec core §12.2): CURATOR_BUILD_HTTPS_TOKEN without
CURATOR_BUILD_HTTPS_HOST is not bound to any one repository identity — it is
offered to every private HTTPS build repository host this run reaches. Set
CURATOR_BUILD_HTTPS_HOST to bind it to one host, or prefer a build_https
scope, unless every one of those hosts is meant to receive that token.
```

### curator config build-https list --help

Exit: 0

```text
usage: curator config build-https list

List configured HTTPS credential scopes.

curator config build-https: operator HTTPS tokens for external build repositories

Usage:
  curator config build-https add <scope> (--git-credentials | --keyring | --token-env NAME) [--username NAME]
  curator config build-https login <scope> [--username NAME]
  curator config build-https list
  curator config build-https remove <scope>

A scope is a lowercase host optionally followed by '/'-separated path segments.
It is matched against the canonical host/path identity of a build repository on
whole-segment boundaries, and the longest matching scope wins: scope
'git.example.com/portals' covers 'git.example.com/portals/app' and never
'git.example.com/portals-other'.

A token is never accepted as a command-line argument, and the config never
stores one as a literal value. add names where the token is read from:
  --git-credentials  the operator's own Git HTTPS credential for the scope's
                      host, read through 'git credential fill'
  --keyring           the token already stored for this scope by login
  --token-env NAME    an environment variable read at process entry
  --username NAME     sent alongside the resolved token; defaults to "token"

login reads a token from a hidden prompt, or one line from stdin when it is
not attached to a terminal, stores it through the operator's own Git
credential machinery under a scope-namespaced entry, and selects it for the
scope the same way 'add --keyring' would.

add and login replace whatever was recorded under the same scope; remove
reports a scope that is not configured as an error, and also deletes the
scope's stored token when it selected one; list prints one line per scope,
sorted, and reports whether the token it names actually resolves right now.

Precedence: CURATOR_BUILD_HTTPS_TOKEN, optionally pinned to one host by
CURATOR_BUILD_HTTPS_HOST, covers every repository on that host ahead of the
scopes configured here; unset, or once that host is exhausted, the longest
matching build_https scope applies; an HTTPS repository nothing covers is not
an error, unlike SSH — it fetches anonymously. Credentials are operator-owned:
no manifest, descriptor, repository, substitution, or marker can select them.

Disclosure warning (Spec core §12.2): CURATOR_BUILD_HTTPS_TOKEN without
CURATOR_BUILD_HTTPS_HOST is not bound to any one repository identity — it is
offered to every private HTTPS build repository host this run reaches. Set
CURATOR_BUILD_HTTPS_HOST to bind it to one host, or prefer a build_https
scope, unless every one of those hosts is meant to receive that token.
```

### curator config build-https remove --help

Exit: 0

```text
usage: curator config build-https remove <scope>

Remove an HTTPS credential scope.

curator config build-https: operator HTTPS tokens for external build repositories

Usage:
  curator config build-https add <scope> (--git-credentials | --keyring | --token-env NAME) [--username NAME]
  curator config build-https login <scope> [--username NAME]
  curator config build-https list
  curator config build-https remove <scope>

A scope is a lowercase host optionally followed by '/'-separated path segments.
It is matched against the canonical host/path identity of a build repository on
whole-segment boundaries, and the longest matching scope wins: scope
'git.example.com/portals' covers 'git.example.com/portals/app' and never
'git.example.com/portals-other'.

A token is never accepted as a command-line argument, and the config never
stores one as a literal value. add names where the token is read from:
  --git-credentials  the operator's own Git HTTPS credential for the scope's
                      host, read through 'git credential fill'
  --keyring           the token already stored for this scope by login
  --token-env NAME    an environment variable read at process entry
  --username NAME     sent alongside the resolved token; defaults to "token"

login reads a token from a hidden prompt, or one line from stdin when it is
not attached to a terminal, stores it through the operator's own Git
credential machinery under a scope-namespaced entry, and selects it for the
scope the same way 'add --keyring' would.

add and login replace whatever was recorded under the same scope; remove
reports a scope that is not configured as an error, and also deletes the
scope's stored token when it selected one; list prints one line per scope,
sorted, and reports whether the token it names actually resolves right now.

Precedence: CURATOR_BUILD_HTTPS_TOKEN, optionally pinned to one host by
CURATOR_BUILD_HTTPS_HOST, covers every repository on that host ahead of the
scopes configured here; unset, or once that host is exhausted, the longest
matching build_https scope applies; an HTTPS repository nothing covers is not
an error, unlike SSH — it fetches anonymously. Credentials are operator-owned:
no manifest, descriptor, repository, substitution, or marker can select them.

Disclosure warning (Spec core §12.2): CURATOR_BUILD_HTTPS_TOKEN without
CURATOR_BUILD_HTTPS_HOST is not bound to any one repository identity — it is
offered to every private HTTPS build repository host this run reaches. Set
CURATOR_BUILD_HTTPS_HOST to bind it to one host, or prefer a build_https
scope, unless every one of those hosts is meant to receive that token.
```

### curator version --help

Exit: 0

```text
usage: curator version

Print the curator version.

```

### curator-run --version

Exit: 0

```text
curator-run 0.3.0 (specification 0.5.1-draft)
```

### curator-run --help

Exit: 0

```text
curator-run 0.3.0 (specification 0.5.1-draft)

usage: curator-run <env-id> [--hosted | --native | --untracked]
                   [resume [SES-HANDLE]] [--resume <id>]
                   [--profile <name>] [--system-prompt <append|replace>]
                   [--model <model>] [--effort <effort>]
                   [--permissions <native|yolo> | --yolo]
                   [--name <session-name>] [--ax-profile <standard|yolo>]
                   [--network <profile>] [--network-policy <policy>]
                   [--network-known-bad <path>]
                   [--] <native args...>
       curator-run session-manager list [--json]
       curator-run session-manager show <id> [--json]
       curator-run session-manager register --manifest <path|-> [--select | --no-select]
       curator-run session-manager select <id>
       curator-run session-manager unregister <id>
       curator-run werkstatt up <id> [--attach] [--new] [--task-class C [--difficulty D]]
       curator-run werkstatt down <id> [--force]
       curator-run werkstatt recycle <id>
       curator-run werkstatt status [<id>] [--json] [--set-task T | --clear-task] [--queue-depth N]
                                    [--rebind-manager | --abandon-operation]
       curator-run werkstatt list [--json]
       curator-run migrate [--dry-run]
       curator-run --help | -h
       curator-run --version

Launcher flags are recognized only before "--". Everything after "--" is
forwarded to the tool verbatim, in order, uninspected.

options:
  --hosted                          use the selected registered session manager
  --native, --untracked             run natively; the two spellings are equal
  --board-goal                      bind the workspace board goal (hosted only)
  --a2a <identity>                  request runtime doorbell wiring (hosted only)
  --detach                          launch without attaching (hosted only)
  resume [SES-HANDLE]               hosted resume selector (latest or handle)
  --resume <id>                     hosted resume selector (provider identity)
  --profile <name>                  Curator profile forwarded to env resolve
  --system-prompt <append|replace>  engage the fragment's system-prompt channel
  --model <model>                   model passed to the spawn plane as declared
  --effort <effort>                 reasoning effort passed as declared
  --permissions <native|yolo>        permission mode passed to the spawn plane
  --yolo                             alias for --permissions yolo
  -d, --danger                       rejected; use --permissions explicitly
  --name <session-name>             ax session name (tracked machines only)
  --ax-profile <standard|yolo>      ax execution profile (tracked machines only)
  --network-policy <policy>        optimistic (default), strict, pinned:<binary SHA-256>
  --network-known-bad <path>        explicit operator known-bad list (failure refuses)
  --network <profile>               network profile for direct execution (tracked refuses)
  --help, -h                        print this usage text and exit 0
  --version                         print the launcher name and version, exit 0

Environments: claude_code (alias claude), codex_cli (alias codex), pi, muse.
opencode is currently unsupported. Aliases normalize to the canonical id
before validation or lookup; outputs carry the canonical id only.
Omitting --profile uses Curator's current profile; resolution always repairs.
Use curator run <env> --profile <p> -- <args> for umbrella discovery.
Install curator-run on trusted PATH (e.g. /usr/local/bin), not Curator's
~/.local/bin shim directory or another managed directory.

Defaults: flags > operator > machine > admitted lineup, per model/effort.
Files: $XDG_CONFIG_HOME/swarma/launcher/defaults.json (default
~/.config/swarma/launcher/defaults.json) and /etc/swarma/launcher/defaults.json.
Machine locked:true ignores operator entries for its named environments and
rejects flags overriding members it sets. Pi fallback preference is
pi-anthropic, pi-openai, pi-google; explicit models bind their own runtime.
Pi has no MCP channel. Prompt opt-in defaults off; native prompt-file
candidates can still produce warnings without opt-in.

Tracking: sibling ax.json, schema curator-run-ax-v1, enabled:true.
An existing machine ax.json wins; otherwise use the operator file.
Missing files or enabled:false select direct execution. Invalid/unreadable
configuration refuses even --help/--version. No direct fallback on ax failure.
--name defaults to <env-id>-<YYYYMMDDTHHMMSSZ> (UTC), accepts
[A-Za-z0-9][A-Za-z0-9._-]{0,63}, and has no effect when untracked.
--ax-profile is a usage error when untracked; absent uses ax's default.
Repeated/unknown flags, missing values and extra operands before -- refuse.

Host: --hosted, --native and --untracked share one host request: a repeat
or a combination refuses. Machine launcher.json selects native, hosted or auto
(default). Native skips the manager registry; explicit --native also bypasses
ax. Auto uses the selected registered manager, or launches natively with a notice.
An invalid registry refuses auto; hosted refusals never fall back to native.
Delegated launches require explicit --hosted and machine delegated_trial:true
until sign-off. Implicit auto stays native for delegated-only managers; implicit
hosted refuses session_host_missing. A missing upgrade capability prints a notice.
The v3 defaults host member is withdrawn; use migrate.
resume selectors and --resume are hosted-only; in native mode they refuse.
--hosted with --network refuses network_scope_unsupported (exit 16).

Exit codes: 0 success; 2 usage; 1 operational refusal. Direct child exit codes
propagate unchanged; signal exits are 128 + signal. Hosted refusals use the
receiver exit mapping (1, 2, 6, 16). Stable diagnostic codes and details go
to stderr; see README.md and SPEC.md section 6.
```

### curator-run session-manager --help

Exit: 2

```text
curator-run: usage: expected session-manager list|show|register|select|unregister with its documented arguments
```

### curator-run session-manager list --help

Exit: 2

```text
curator-run: usage: expected session-manager list|show|register|select|unregister with its documented arguments
```

### curator-run session-manager show --help

Exit: 2

```text
curator-run: usage: expected session-manager list|show|register|select|unregister with its documented arguments
```

### curator-run session-manager register --help

Exit: 2

```text
curator-run: usage: expected session-manager list|show|register|select|unregister with its documented arguments
```

### curator-run session-manager select --help

Exit: 2

```text
curator-run: usage: expected session-manager list|show|register|select|unregister with its documented arguments
```

### curator-run session-manager unregister --help

Exit: 2

```text
curator-run: usage: expected session-manager list|show|register|select|unregister with its documented arguments
```

### curator-run werkstatt --help

Exit: 2

```text
curator-run: usage: expected werkstatt up|down|recycle|status|list with its documented arguments
```

### curator-run werkstatt up --help

Exit: 2

```text
curator-run: usage: expected werkstatt up|down|recycle|status|list with its documented arguments
```

### curator-run werkstatt down --help

Exit: 2

```text
curator-run: usage: expected werkstatt up|down|recycle|status|list with its documented arguments
```

### curator-run werkstatt recycle --help

Exit: 2

```text
curator-run: usage: expected werkstatt up|down|recycle|status|list with its documented arguments
```

### curator-run werkstatt status --help

Exit: 2

```text
curator-run: usage: expected werkstatt up|down|recycle|status|list with its documented arguments
```

### curator-run werkstatt list --help

Exit: 2

```text
curator-run: usage: expected werkstatt up|down|recycle|status|list with its documented arguments
```

### curator-run migrate --help

Exit: 2

```text
curator-run: usage: expected migrate [--dry-run]
```

### curator-run codex --help

Exit: 0

```text
curator-run 0.3.0 (specification 0.5.1-draft)

usage: curator-run <env-id> [--hosted | --native | --untracked]
                   [resume [SES-HANDLE]] [--resume <id>]
                   [--profile <name>] [--system-prompt <append|replace>]
                   [--model <model>] [--effort <effort>]
                   [--permissions <native|yolo> | --yolo]
                   [--name <session-name>] [--ax-profile <standard|yolo>]
                   [--network <profile>] [--network-policy <policy>]
                   [--network-known-bad <path>]
                   [--] <native args...>
       curator-run session-manager list [--json]
       curator-run session-manager show <id> [--json]
       curator-run session-manager register --manifest <path|-> [--select | --no-select]
       curator-run session-manager select <id>
       curator-run session-manager unregister <id>
       curator-run werkstatt up <id> [--attach] [--new] [--task-class C [--difficulty D]]
       curator-run werkstatt down <id> [--force]
       curator-run werkstatt recycle <id>
       curator-run werkstatt status [<id>] [--json] [--set-task T | --clear-task] [--queue-depth N]
                                    [--rebind-manager | --abandon-operation]
       curator-run werkstatt list [--json]
       curator-run migrate [--dry-run]
       curator-run --help | -h
       curator-run --version

Launcher flags are recognized only before "--". Everything after "--" is
forwarded to the tool verbatim, in order, uninspected.

options:
  --hosted                          use the selected registered session manager
  --native, --untracked             run natively; the two spellings are equal
  --board-goal                      bind the workspace board goal (hosted only)
  --a2a <identity>                  request runtime doorbell wiring (hosted only)
  --detach                          launch without attaching (hosted only)
  resume [SES-HANDLE]               hosted resume selector (latest or handle)
  --resume <id>                     hosted resume selector (provider identity)
  --profile <name>                  Curator profile forwarded to env resolve
  --system-prompt <append|replace>  engage the fragment's system-prompt channel
  --model <model>                   model passed to the spawn plane as declared
  --effort <effort>                 reasoning effort passed as declared
  --permissions <native|yolo>        permission mode passed to the spawn plane
  --yolo                             alias for --permissions yolo
  -d, --danger                       rejected; use --permissions explicitly
  --name <session-name>             ax session name (tracked machines only)
  --ax-profile <standard|yolo>      ax execution profile (tracked machines only)
  --network-policy <policy>        optimistic (default), strict, pinned:<binary SHA-256>
  --network-known-bad <path>        explicit operator known-bad list (failure refuses)
  --network <profile>               network profile for direct execution (tracked refuses)
  --help, -h                        print this usage text and exit 0
  --version                         print the launcher name and version, exit 0

Environments: claude_code (alias claude), codex_cli (alias codex), pi, muse.
opencode is currently unsupported. Aliases normalize to the canonical id
before validation or lookup; outputs carry the canonical id only.
Omitting --profile uses Curator's current profile; resolution always repairs.
Use curator run <env> --profile <p> -- <args> for umbrella discovery.
Install curator-run on trusted PATH (e.g. /usr/local/bin), not Curator's
~/.local/bin shim directory or another managed directory.

Defaults: flags > operator > machine > admitted lineup, per model/effort.
Files: $XDG_CONFIG_HOME/swarma/launcher/defaults.json (default
~/.config/swarma/launcher/defaults.json) and /etc/swarma/launcher/defaults.json.
Machine locked:true ignores operator entries for its named environments and
rejects flags overriding members it sets. Pi fallback preference is
pi-anthropic, pi-openai, pi-google; explicit models bind their own runtime.
Pi has no MCP channel. Prompt opt-in defaults off; native prompt-file
candidates can still produce warnings without opt-in.

Tracking: sibling ax.json, schema curator-run-ax-v1, enabled:true.
An existing machine ax.json wins; otherwise use the operator file.
Missing files or enabled:false select direct execution. Invalid/unreadable
configuration refuses even --help/--version. No direct fallback on ax failure.
--name defaults to <env-id>-<YYYYMMDDTHHMMSSZ> (UTC), accepts
[A-Za-z0-9][A-Za-z0-9._-]{0,63}, and has no effect when untracked.
--ax-profile is a usage error when untracked; absent uses ax's default.
Repeated/unknown flags, missing values and extra operands before -- refuse.

Host: --hosted, --native and --untracked share one host request: a repeat
or a combination refuses. Machine launcher.json selects native, hosted or auto
(default). Native skips the manager registry; explicit --native also bypasses
ax. Auto uses the selected registered manager, or launches natively with a notice.
An invalid registry refuses auto; hosted refusals never fall back to native.
Delegated launches require explicit --hosted and machine delegated_trial:true
until sign-off. Implicit auto stays native for delegated-only managers; implicit
hosted refuses session_host_missing. A missing upgrade capability prints a notice.
The v3 defaults host member is withdrawn; use migrate.
resume selectors and --resume are hosted-only; in native mode they refuse.
--hosted with --network refuses network_scope_unsupported (exit 16).

Exit codes: 0 success; 2 usage; 1 operational refusal. Direct child exit codes
propagate unchanged; signal exits are 128 + signal. Hosted refusals use the
receiver exit mapping (1, 2, 6, 16). Stable diagnostic codes and details go
to stderr; see README.md and SPEC.md section 6.
```

### task-board session-manager --help

Exit: 0

```text
Register Task-board with the machine launcher

Usage:
  task-board session-manager [command]

Available Commands:
  register    Pipe the installed Task-board manifest to curator-run

Flags:
  -h, --help   help for session-manager

Global Flags:
      --board-dir string         Path to the board directory (overrides TASK_BOARD_DIR and config local.board_dir) (default ".task-board")
      --context-profile string   Agent context profile override: full or lite
      --insecure                 Skip TLS certificate verification (for self-signed certs)
      --json                     Output in JSON format
      --no-update-check          Disable automatic update check
      --remote string            Remote board server URL (overrides TASK_BOARD_REMOTE)
      --remote-board string      Remote board ID (overrides TASK_BOARD_ID)

Use "task-board session-manager [command] --help" for more information about a command.
```

### task-board session-manager register --help

Exit: 0

```text
Pipe the installed Task-board manifest to curator-run

Usage:
  task-board session-manager register --for curator-run [--select | --no-select] [flags]

Flags:
      --for string   Launcher to register with (curator-run)
  -h, --help         help for register
      --no-select    Register without changing the selection
      --select       Register and select Task-board

Global Flags:
      --board-dir string         Path to the board directory (overrides TASK_BOARD_DIR and config local.board_dir) (default ".task-board")
      --context-profile string   Agent context profile override: full or lite
      --insecure                 Skip TLS certificate verification (for self-signed certs)
      --json                     Output in JSON format
      --no-update-check          Disable automatic update check
      --remote string            Remote board server URL (overrides TASK_BOARD_REMOTE)
      --remote-board string      Remote board ID (overrides TASK_BOARD_ID)
```

## Bounded probes

### curator run session-manager register --help

Claim: Umbrella nested help stops at the generic run topic.

Asserted exit: 0

```text
usage: curator run <env-id> [arguments] [-- <passthrough arguments>]

Dispatch to the installed curator-run provider; arguments after -- are passed unchanged.

curator: <env-id> accepts claude and codex as aliases of claude_code and codex_cli; outputs keep the canonical id
```

### curator-run session-manager register --help

Claim: Direct nested registry help rejects the help flag before any registry mutation.

Asserted exit: 2

```text
curator-run: usage: expected session-manager list|show|register|select|unregister with its documented arguments
```

### curator-run werkstatt status example --help

Claim: A complete station status operand still does not admit --help.

Asserted exit: 2

```text
curator-run: usage: expected werkstatt up|down|recycle|status|list with its documented arguments
```

### curator-run --help

Claim: Top-level launcher help succeeds with absent fixture user configuration.

Asserted exit: 0

```text
curator-run 0.3.0 (specification 0.5.1-draft)

usage: curator-run <env-id> [--hosted | --native | --untracked]
                   [resume [SES-HANDLE]] [--resume <id>]
                   [--profile <name>] [--system-prompt <append|replace>]
                   [--model <model>] [--effort <effort>]
                   [--permissions <native|yolo> | --yolo]
                   [--name <session-name>] [--ax-profile <standard|yolo>]
                   [--network <profile>] [--network-policy <policy>]
                   [--network-known-bad <path>]
                   [--] <native args...>
       curator-run session-manager list [--json]
       curator-run session-manager show <id> [--json]
       curator-run session-manager register --manifest <path|-> [--select | --no-select]
       curator-run session-manager select <id>
       curator-run session-manager unregister <id>
       curator-run werkstatt up <id> [--attach] [--new] [--task-class C [--difficulty D]]
       curator-run werkstatt down <id> [--force]
       curator-run werkstatt recycle <id>
       curator-run werkstatt status [<id>] [--json] [--set-task T | --clear-task] [--queue-depth N]
                                    [--rebind-manager | --abandon-operation]
       curator-run werkstatt list [--json]
       curator-run migrate [--dry-run]
       curator-run --help | -h
       curator-run --version

Launcher flags are recognized only before "--". Everything after "--" is
forwarded to the tool verbatim, in order, uninspected.

options:
  --hosted                          use the selected registered session manager
  --native, --untracked             run natively; the two spellings are equal
  --board-goal                      bind the workspace board goal (hosted only)
  --a2a <identity>                  request runtime doorbell wiring (hosted only)
  --detach                          launch without attaching (hosted only)
  resume [SES-HANDLE]               hosted resume selector (latest or handle)
  --resume <id>                     hosted resume selector (provider identity)
  --profile <name>                  Curator profile forwarded to env resolve
  --system-prompt <append|replace>  engage the fragment's system-prompt channel
  --model <model>                   model passed to the spawn plane as declared
  --effort <effort>                 reasoning effort passed as declared
  --permissions <native|yolo>        permission mode passed to the spawn plane
  --yolo                             alias for --permissions yolo
  -d, --danger                       rejected; use --permissions explicitly
  --name <session-name>             ax session name (tracked machines only)
  --ax-profile <standard|yolo>      ax execution profile (tracked machines only)
  --network-policy <policy>        optimistic (default), strict, pinned:<binary SHA-256>
  --network-known-bad <path>        explicit operator known-bad list (failure refuses)
  --network <profile>               network profile for direct execution (tracked refuses)
  --help, -h                        print this usage text and exit 0
  --version                         print the launcher name and version, exit 0

Environments: claude_code (alias claude), codex_cli (alias codex), pi, muse.
opencode is currently unsupported. Aliases normalize to the canonical id
before validation or lookup; outputs carry the canonical id only.
Omitting --profile uses Curator's current profile; resolution always repairs.
Use curator run <env> --profile <p> -- <args> for umbrella discovery.
Install curator-run on trusted PATH (e.g. /usr/local/bin), not Curator's
~/.local/bin shim directory or another managed directory.

Defaults: flags > operator > machine > admitted lineup, per model/effort.
Files: $XDG_CONFIG_HOME/swarma/launcher/defaults.json (default
~/.config/swarma/launcher/defaults.json) and /etc/swarma/launcher/defaults.json.
Machine locked:true ignores operator entries for its named environments and
rejects flags overriding members it sets. Pi fallback preference is
pi-anthropic, pi-openai, pi-google; explicit models bind their own runtime.
Pi has no MCP channel. Prompt opt-in defaults off; native prompt-file
candidates can still produce warnings without opt-in.

Tracking: sibling ax.json, schema curator-run-ax-v1, enabled:true.
An existing machine ax.json wins; otherwise use the operator file.
Missing files or enabled:false select direct execution. Invalid/unreadable
configuration refuses even --help/--version. No direct fallback on ax failure.
--name defaults to <env-id>-<YYYYMMDDTHHMMSSZ> (UTC), accepts
[A-Za-z0-9][A-Za-z0-9._-]{0,63}, and has no effect when untracked.
--ax-profile is a usage error when untracked; absent uses ax's default.
Repeated/unknown flags, missing values and extra operands before -- refuse.

Host: --hosted, --native and --untracked share one host request: a repeat
or a combination refuses. Machine launcher.json selects native, hosted or auto
(default). Native skips the manager registry; explicit --native also bypasses
ax. Auto uses the selected registered manager, or launches natively with a notice.
An invalid registry refuses auto; hosted refusals never fall back to native.
Delegated launches require explicit --hosted and machine delegated_trial:true
until sign-off. Implicit auto stays native for delegated-only managers; implicit
hosted refuses session_host_missing. A missing upgrade capability prints a notice.
The v3 defaults host member is withdrawn; use migrate.
resume selectors and --resume are hosted-only; in native mode they refuse.
--hosted with --network refuses network_scope_unsupported (exit 16).

Exit codes: 0 success; 2 usage; 1 operational refusal. Direct child exit codes
propagate unchanged; signal exits are 128 + signal. Hosted refusals use the
receiver exit mapping (1, 2, 6, 16). Stable diagnostic codes and details go
to stderr; see README.md and SPEC.md section 6.
```

### curator-run codex --native --hosted

Claim: Conflicting host flags are rejected by syntax admission.

Asserted exit: 2

```text
curator-run: usage: --hosted, --native and --untracked share one host request and cannot be repeated or combined

usage: curator-run <env-id> [--hosted | --native | --untracked]
                   [resume [SES-HANDLE]] [--resume <id>]
                   [--profile <name>] [--system-prompt <append|replace>]
                   [--model <model>] [--effort <effort>]
                   [--permissions <native|yolo> | --yolo]
                   [--name <session-name>] [--ax-profile <standard|yolo>]
                   [--network <profile>] [--network-policy <policy>]
                   [--network-known-bad <path>]
                   [--] <native args...>
       curator-run session-manager list [--json]
       curator-run session-manager show <id> [--json]
       curator-run session-manager register --manifest <path|-> [--select | --no-select]
       curator-run session-manager select <id>
       curator-run session-manager unregister <id>
       curator-run werkstatt up <id> [--attach] [--new] [--task-class C [--difficulty D]]
       curator-run werkstatt down <id> [--force]
       curator-run werkstatt recycle <id>
       curator-run werkstatt status [<id>] [--json] [--set-task T | --clear-task] [--queue-depth N]
                                    [--rebind-manager | --abandon-operation]
       curator-run werkstatt list [--json]
       curator-run migrate [--dry-run]
       curator-run --help | -h
       curator-run --version

Launcher flags are recognized only before "--". Everything after "--" is
forwarded to the tool verbatim, in order, uninspected.

options:
  --hosted                          use the selected registered session manager
  --native, --untracked             run natively; the two spellings are equal
  --board-goal                      bind the workspace board goal (hosted only)
  --a2a <identity>                  request runtime doorbell wiring (hosted only)
  --detach                          launch without attaching (hosted only)
  resume [SES-HANDLE]               hosted resume selector (latest or handle)
  --resume <id>                     hosted resume selector (provider identity)
  --profile <name>                  Curator profile forwarded to env resolve
  --system-prompt <append|replace>  engage the fragment's system-prompt channel
  --model <model>                   model passed to the spawn plane as declared
  --effort <effort>                 reasoning effort passed as declared
  --permissions <native|yolo>        permission mode passed to the spawn plane
  --yolo                             alias for --permissions yolo
  -d, --danger                       rejected; use --permissions explicitly
  --name <session-name>             ax session name (tracked machines only)
  --ax-profile <standard|yolo>      ax execution profile (tracked machines only)
  --network-policy <policy>        optimistic (default), strict, pinned:<binary SHA-256>
  --network-known-bad <path>        explicit operator known-bad list (failure refuses)
  --network <profile>               network profile for direct execution (tracked refuses)
  --help, -h                        print this usage text and exit 0
  --version                         print the launcher name and version, exit 0

Environments: claude_code (alias claude), codex_cli (alias codex), pi, muse.
opencode is currently unsupported. Aliases normalize to the canonical id
before validation or lookup; outputs carry the canonical id only.
Omitting --profile uses Curator's current profile; resolution always repairs.
Use curator run <env> --profile <p> -- <args> for umbrella discovery.
Install curator-run on trusted PATH (e.g. /usr/local/bin), not Curator's
~/.local/bin shim directory or another managed directory.

Defaults: flags > operator > machine > admitted lineup, per model/effort.
Files: $XDG_CONFIG_HOME/swarma/launcher/defaults.json (default
~/.config/swarma/launcher/defaults.json) and /etc/swarma/launcher/defaults.json.
Machine locked:true ignores operator entries for its named environments and
rejects flags overriding members it sets. Pi fallback preference is
pi-anthropic, pi-openai, pi-google; explicit models bind their own runtime.
Pi has no MCP channel. Prompt opt-in defaults off; native prompt-file
candidates can still produce warnings without opt-in.

Tracking: sibling ax.json, schema curator-run-ax-v1, enabled:true.
An existing machine ax.json wins; otherwise use the operator file.
Missing files or enabled:false select direct execution. Invalid/unreadable
configuration refuses even --help/--version. No direct fallback on ax failure.
--name defaults to <env-id>-<YYYYMMDDTHHMMSSZ> (UTC), accepts
[A-Za-z0-9][A-Za-z0-9._-]{0,63}, and has no effect when untracked.
--ax-profile is a usage error when untracked; absent uses ax's default.
Repeated/unknown flags, missing values and extra operands before -- refuse.

Host: --hosted, --native and --untracked share one host request: a repeat
or a combination refuses. Machine launcher.json selects native, hosted or auto
(default). Native skips the manager registry; explicit --native also bypasses
ax. Auto uses the selected registered manager, or launches natively with a notice.
An invalid registry refuses auto; hosted refusals never fall back to native.
Delegated launches require explicit --hosted and machine delegated_trial:true
until sign-off. Implicit auto stays native for delegated-only managers; implicit
hosted refuses session_host_missing. A missing upgrade capability prints a notice.
The v3 defaults host member is withdrawn; use migrate.
resume selectors and --resume are hosted-only; in native mode they refuse.
--hosted with --network refuses network_scope_unsupported (exit 16).

Exit codes: 0 success; 2 usage; 1 operational refusal. Direct child exit codes
propagate unchanged; signal exits are 128 + signal. Hosted refusals use the
receiver exit mapping (1, 2, 6, 16). Stable diagnostic codes and details go
to stderr; see README.md and SPEC.md section 6.
```

### curator-run werkstatt down example --json

Claim: A flag documented only on status/list is rejected on down.

Asserted exit: 2

```text
curator-run: usage: expected werkstatt up|down|recycle|status|list with its documented arguments
```

### curator-run --help

Claim: Malformed fixture launcher policy prevents top-level help; this probe changes no real machine policy.

Asserted exit: 1

```text
curator-run: defaults_config_invalid: <review-fixture-root>/probe-config/swarma/launcher/launcher.json: json: expected object key at byte 1
```

