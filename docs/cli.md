# Curator CLI Reference

Every synopsis and flag in this reference was verified verbatim against `./bin/curator` built via `make build` from the repository tree.

## Shared flags

Curator flags use consistent names and behavior across command groups.

For every built-in command (including nested verbs), `-h` or `--help` anywhere
before `--` prints dedicated command help and exits 0 without modifying files or
loading machine configuration. `curator help <command>` prints the same help.
For example, `curator init -h` and `curator help env config set` are read-only.
Arguments after `--` are operands; `curator run codex -- -h` forwards the literal
`-h` to the run provider.

- `--all`: operate on all configured projects.
- `--audit`: run the audit gate in advisory or strict mode.
- `--dry-run`: plan work without modifying files.
- `--check`: exit non-zero unless every skill or compiled command is up to date.
- `--json`: format status, audit, or validation output as machine-readable JSON.
- `--branch string`: set git branch for skill repositories.
- `--git string`: set git clone URL for skill repositories.
- `--project string`: project alias or path.
- `--revision string`: set git revision for skill repositories.
- `--tag string`: set git tag for skill repositories.
- `--source string`: set source directory under `skills_root`.

## Environment profiles

### curator profile update

```bash
curator profile update [<name>|--all] [--revision <ref>|--tag <tag>|--range <range>] [--takeover] [--confirm-system-delta]
```

Without a requirement flag, re-resolve the declared requirements of the named
profile, the current machine profile when no name is given, or every installed
non-builtin profile with `--all`. Exact revision roots stay pinned.

With one requirement flag, repin one explicitly named Git profile in place.
Only the root requirement changes; the installed Git URL and package directory
stay fixed. Resolution, strict audit, system-module admission, overlay
resolution, missing-pin rules and scope rematerialization use the ordinary
update pipeline. Existing running seats can retain their homes; removing the
profile is unnecessary.

- `--revision <ref>`: replace the root requirement with an exact Git revision.
- `--tag <tag>`: replace it with a Git tag.
- `--range <range>`: replace it with a version range over tags.
- `--confirm-system-delta`: acknowledge system-module or MCP declaration changes
  for this operation, as on ordinary update.
- `--takeover`: consent to replacing unmanaged files that rematerialization
  would write, with the ordinary backups.

At most one requirement flag is allowed. Requirement flags on a path or builtin
profile, with `--all`, without an explicit name, or with an empty value are usage
errors (`profile_install_ref_conflict`, exit 2). Update accepts no URL or directory
replacement. Installing a different URL, directory or requirement under an
existing name still refuses `profile_name_taken`; use update to change only the
requirement.

A refused repin preserves the stored source/requirement, lock and homes. Source
and successor lock are published together under the manager mutation lock.
Acquisition and audit caches may refresh while checking a candidate. If the
candidate resolves to the existing lock, print `<name>: unchanged` and leave the
stored requirement and homes unchanged, even if the flag uses another form.

## Environment initialization and skill lifecycle

<a id="bootstrap"></a>

### curator bootstrap

`curator bootstrap` creates machine configuration and default directories.

Synopsis:

```bash
curator bootstrap [flags]
```

Flags:

- `--default-agents string`: comma-separated default agents (default: `codex_cli`).
- `--force`: overwrite an existing configuration.
- `--if-missing`: create configuration only when absent.
- `--non-interactive`: fail instead of prompting for missing values.
- `--preferred-locale string`: preferred locale setting.
- `--skills-root string`: directory containing skill repositories.

Run bootstrap to create missing machine configuration:

```bash
curator bootstrap --if-missing --non-interactive --skills-root "$HOME/skills"
```

The command initializes `$HOME/.curator/config.json` without overwriting existing files. `--non-interactive` requires `--skills-root`; without it bootstrap exits 2 (`bootstrap requires --skills-root`).

### curator init

`curator init` initializes project declarative files in the specified directory.

Synopsis:

```bash
curator init [path]
```

Run init in the current directory:

```bash
curator init .
```

The command creates `Skillfile.json` and appends managed entries to `.gitignore`.

### curator add

`curator add` adds or replaces a skill declaration, then installs the skill package.

Synopsis:

```bash
curator add <name> [options]
```

Flags:

- `--branch string`: git branch.
- `--git string`: git clone URL.
- `--project string`: project alias or path.
- `--revision string`: git revision.
- `--source string`: source directory under `skills_root`.
- `--tag string`: git tag.

Add a skill package from a git clone URL:

```bash
curator add helper --git https://github.com/example/helper.git --tag v1.0.0
```

The command updates `Skillfile.json` and installs the skill package into `.agents/skills/`.

### curator remove

`curator remove` removes a skill declaration from the project.

Synopsis:

```bash
curator remove <name> [path]
```

Flags:

- `--project string`: project alias or path.

Remove a declared skill:

```bash
curator remove helper
```

The command updates `Skillfile.json` and removes the installed skill directory.

### curator install

`curator install` applies `Skillfile.json` and materializes project dependencies.

A non-git product folder containing nested repositories is also a supported
project root. Install materializes its declared skills and prints a notice that
gitignore hygiene does not apply; it does not initialize Git or create a
`.gitignore`, even with `--fix-gitignore`. In Git work trees, the existing ignore
requirements still apply, and unexpected Git failures refuse installation.
The schema-2 `Skillfile.lock.json` lives at the project root; machine-local
source bindings live under the manager home's `source-bindings/` directory.

Synopsis:

```bash
curator install [path] [flags]
```

Flags:

- `--all`: operate on all configured projects.
- `--audit`: run the audit gate in advisory or strict mode.
- `--build-ssh-agent string`: agent socket for external SSH build repositories, or `auto` for environment agent (or `CURATOR_BUILD_SSH_AGENT`).
- `--build-ssh-identity string`: identity file for external SSH build repositories (or `CURATOR_BUILD_SSH_IDENTITY`).
- `--build-ssh-known-hosts string`: host keys external SSH build repositories are verified against (or `CURATOR_BUILD_SSH_KNOWN_HOSTS`).
- `--dry-run`: plan work without modifying files.
- `--fix-gitignore`: append missing managed gitignore entries.
- `--strict-tags`: fail if an installed tag moved to another commit.
- `--verbose`: print detailed progress.

Run install in dry-run mode:

```bash
curator install --dry-run
```

The command prints planned installation steps without modifying disk files.

For schema-2 projects, `install` materializes the locked snapshot and
repairs drifted bytes from the lock without touching the lock itself;
see [Skillfile schema 2 project sources](#skillfile-schema-2-project-sources).
Run `install -h` for the workflow summary.

### curator update

`curator update` fetches all source repositories under `skills_root`.

Synopsis:

```bash
curator update
```

Run update across source repositories:

```bash
curator update
```

The command refreshes git repositories stored in the manager source root.

### curator upgrade

`curator upgrade` fetches the selected dependency closure and installs updated skill packages.

Synopsis:

```bash
curator upgrade [path] [flags]
```

Flags:

- `--all`: operate on all configured projects.
- `--audit`: run the audit gate in advisory or strict mode.
- `--build-ssh-agent string`: agent socket for external SSH build repositories, or `auto` for environment agent (or `CURATOR_BUILD_SSH_AGENT`).
- `--build-ssh-identity string`: identity file for external SSH build repositories (or `CURATOR_BUILD_SSH_IDENTITY`).
- `--build-ssh-known-hosts string`: host keys external SSH build repositories are verified against (or `CURATOR_BUILD_SSH_KNOWN_HOSTS`).
- `--dry-run`: plan work without modifying files.
- `--fix-gitignore`: append missing managed gitignore entries.
- `--strict-tags`: fail if an installed tag moved to another commit.
- `--verbose`: print detailed progress.

Upgrade project dependencies:

```bash
curator upgrade .
```

For schema-2 projects, upgrade consumes the committed lock without resolving
tags or branches again. It re-materializes missing snapshots from their
declared source at the locked revision and leaves the lock unchanged. Use
`curator project refresh` to explicitly resolve refs and replace the lock.

#### Refreshing existing source checkouts

For schema-1 (legacy) Skillfiles, `upgrade`, `global upgrade` and every
other closure fetch refresh each reached checkout under `skills_root`, and
each existing `Skillfile.dev.json` git-substitution checkout, from exactly
one endpoint. They never run `git fetch --all`, never contact a remote other
than the one chosen below, and never rewrite a checkout's remotes. The stored
origin is read from the checkout's local configuration, without network
access and before any `insteadOf` rewrite. A checkout without an origin is
not fetched and resolves from its local state.

- A declared network `git` URL passes the `allowed_sources` gate first, with
  the clone path's wording. The checkout's stored origin must have the same
  canonical identity; a different spelling of that identity (`.git` suffix,
  host case, scp or HTTPS form) is fine. The declared URL itself is then
  fetched over the operator's transport, the one clone uses.
- With no declared `git` URL, the stored origin is the refresh source. A
  network origin passes the `allowed_sources` gate and is fetched over the
  operator's transport. A local origin (a path or `file://`) is fetched with
  user and system git configuration isolated, by its absolute physical path:
  a relative origin is read from the checkout, as git reads it. Git is never
  handed the raw spelling, because git reads a bare word such as `mirror`, or
  a `file://` URL, as the name of a configured remote before a location.
  The physical path is handed to git byte for byte: a symlink target whose
  name ends in a space, tab or newline is fetched as named, never as the
  sibling without that byte.
- A declared local path must name the stored origin's location, or the
  checkout itself. A relative declaration is read from the working directory.
  Every path is resolved as the filesystem walks it, with each symlink
  resolved before a `..` that follows it, and two existing locations compare
  by filesystem identity. The stored origin is then fetched as above.

The choice and its checks run for every requirer of a checkout. When several
projects are upgraded together (`upgrade --all`), a checkout an earlier
project already refreshed is still gated and compared for each later
project; only a second fetch of the identical endpoint is skipped.

Refusals, each made before anything is fetched unless noted:

- `source_origin_mismatch`: the stored origin names another repository than
  the declaration, or one side is local and the other a network URL. The
  diagnostic names the declared URL, the stored origin, both identities and
  the checkout. Repoint the checkout's origin or fix the declaration.
- `source_path_whitespace`: a declared local path, or a stored local origin,
  begins or ends with whitespace. Git reads those bytes as part of the path,
  while a declaration is trimmed (the clone path clones the trimmed
  spelling), so the two readings can name different directories. Also
  applies before a clone or a dev-substitution clone. Remove the whitespace
  from the declaration or the origin. Whitespace inside a spelling, and in
  the directories a spelling resolves to, is kept exactly.
- `source_origin_ambiguous`: on a platform whose absolute paths do not begin
  with `/` (Windows), a local origin's absolute path is also the name of a
  configured remote of the checkout. Rename that remote.
- `source_mixed_spelling`: one resolution declares the same canonical
  identity with two different raw URLs. The diagnostic names both spellings,
  the identity and both requirer chains. Declare one spelling.
- `source_url_redirected`: git reports that the server redirected a clone or
  fetch (`warning: redirecting to <url>`). The diagnostic names the declared
  URL or stored origin and the redirect target: declare the target, or set
  the origin to it. A redirected clone is removed; the refs a redirected
  fetch already wrote stay in the checkout, and the resolution stops.

`curator update` and `curator global update` are not closure fetches: they
still refresh every repository under the source root through its stored
remotes.

### curator status

`curator status` displays manifest, installed, and compiled command states.

Synopsis:

```bash
curator status [path] [flags]
```

Flags:

- `--all`: operate on all configured projects.
- `--attest`: re-check installed skills against trusted registries.
- `--check`: exit non-zero unless every skill is up to date.
- `--json`: machine-readable output.

Check project status in JSON format:

```bash
curator status --json
```

The command prints status diagnostics for declared skills and compiled commands.

For schema-2 projects, `status` compares installed state against the
frozen lock only — it never rescans collections, advances branches, or
replaces a local snapshot; see [Skillfile schema 2 project sources](#skillfile-schema-2-project-sources).
Run `status -h` for the workflow summary.

`curator status` also reports the shell-hook trust posture (Manager profile
§8.6): one `shell-hook-trust:` row per known project env file — every
recorded path plus the `.agents/env.sh` / `.agents/env.ps1` of the reported
projects — with its trust state (`approved`, `shell_hook_env_unapproved`, or
`shell_hook_env_changed`), the absolute path, and, for recorded files, the
recorded `approved_by` value. Untrusted rows name the `curator hook approve`
command that records the current bytes. A recorded file whose bytes are
missing or unreadable keeps its row with an explicit `file is missing` /
`file is unreadable` qualifier (no digest comparison is claimed), and
`--check` treats it as non-current, as it does a changed file; an
unapproved file whose bytes read is a warning row and never fails the
check. The `--json` document carries the same rows under `shell_hook_trust`
(with the `file` qualifier where it applies) and any approval-state warnings
(malformed lines, unreadable state) under `shell_hook_trust_warnings`.
`curator env status` reports the same posture rows with the same `--check`
semantics.
For the profile × environment home rows, only profiles that are current in
some reported scope decide the `--check` exit code; rows of other installed
profiles (for example `home unprovisioned` on a profile never used) are
printed as informational.

Both `curator status` and `curator env status` also report one registry page
boundary row per enabled registry with pinned keys. Each row names the
registry URL, the persisted high-water version and `log_size` when present,
and whether the stored page-boundary posture is verified. The rows appear
after the shell-hook and §12 posture rows; JSON uses `registry_posture`.
First use with no high-water is informational and passes `--check`. An
unreadable or missing-after-use rollback state is reported as unavailable and
fails `--check`; status reads do not create or repair that state.

`curator status --attest` validates the signed `boundary` on every fetched
records page before using its records. A rejected page leaves that registry
unknown for the attestation and prints the `registry_page_boundary_*`
diagnostic with the registry URL in the result detail. The read-only command
does not advance high-water state or write the record cache. A pre-existing
cache entry remains on disk after a rejected chain but is not served for that
operation; it is eligible for a later outage fallback only when its signed
boundary matches the persisted high-water.

For managed-home provisioning, an unreadable marker is reported as unknown,
not unprovisioned: the scope row carries `provisioned_known: false` and an
`environment_marker_unreadable` diagnostic, and `--check` returns non-zero.
Unreadable backup inventories show `backups: unknown` with
`environment_backup_record_unreadable`; unreadable orphan inventory is
reported in the JSON `diagnostics` list instead of appearing empty.
An unreadable profile lock is reported as `environment_store_untrusted` with
unknown currency. Resolve, repair, and update refuse it without rebuilding or
replacing state from that lock. A passthrough entry that cannot be inspected
or whose link target cannot be read is reported as
`environment_passthrough_unreadable`; resolve emits no fragment and repair
leaves the entry untouched. A proven absent or replaced passthrough entry
remains `environment_passthrough_detached`.

### curator list

`curator list` lists configured projects and declared skills.

Synopsis:

```bash
curator list
```

List configured projects and declared skills:

```bash
curator list
```

The command displays project skills and their configured source revisions across all configured projects.

## Project subcommands

### Project-context admission

`project register`, `inspect`, `package`, `preview`, `approve`, `revoke` and
`rebind` are the explicit operator controls for protected checkout registration
and per-item admission. See [project admission](project-admission.md) for exact
commands, request grammar, preview output, platform bounds and unsupported
surfaces. These controls perform no composition or provider launch.

### curator project add

`curator project add` registers a new project alias and path in the machine configuration and initializes `Skillfile.json` and `.gitignore`.

Synopsis:

```bash
curator project add <alias> <path> [flags]
```

Flags:

- `--agents string`: comma-separated target agents.

Add a new project mapping:

```bash
curator project add myproject ./myproject --agents codex_cli
```

The command registers the project mapping and creates the initial project manifest and `.gitignore` file.

### curator project resolve

`curator project resolve` resolves transitive dependencies for a project closure.

Synopsis:

```bash
curator project resolve [path]
```

Resolve project dependencies:

```bash
curator project resolve .
```

Schema-2 projects use the committed `Skillfile.lock.json` described below.
Schema-1 projects retain their existing behavior.

For schema-1 projects the command prints the alias, project path,
Skillfile path, and managed skill and bin directories without modifying
disk state. Schema-2 projects use the workflow described in
[Skillfile schema 2 project sources](#skillfile-schema-2-project-sources);
it acquires every Git source alias, freezes local bytes and Git commits
into a locked plan, and publishes `Skillfile.lock.json` plus the machine
bindings transactionally after every gate succeeds.
Commit `Skillfile.lock.json` with `Skillfile.json`, like
`package-lock.json`. The lock contains portable package identities; the
machine-local source bindings stay outside the project.
`curator project resolve -h` prints the workflow. Schema 1 keeps its
exact meaning; no on-disk migration is implicit. The bare word `help` is
never a help flag: it resolves as a project alias or path.

### curator project refresh

`curator project refresh` re-runs the explicit attempt for a project
closure: refs are resolved anew, local bytes and collection membership
are re-frozen, and the lock plus machine bindings are replaced
atomically only after every gate succeeds. Failure preserves prior
state. For frozen v1 projects it behaves exactly like `project
resolve`.

Synopsis:

```bash
curator project refresh [path]
```

Refresh a schema-2 lock after editing sources or after upstream refs moved:

```bash
curator project refresh .
curator install .
```

Refresh alone changes nothing live: run `curator install` afterwards to
materialize the refreshed lock. Launch and status never refresh on
their own. Run `curator project refresh -h` for the workflow.

## Skillfile schema 2 project sources

Skillfile schema 2 is the default project reader and install path. Schema
1 retains its exact meaning, and no on-disk migration is implicit.

### Workflow

Five existing verbs move a schema-2 project; no new command is added:

Commit `Skillfile.lock.json` with `Skillfile.json`, like
`package-lock.json`. When a local snapshot is missing, install and upgrade
re-read the declared path bytes or fetch the exact locked Git revision. They
continue only when package identity and `content_sha256` match the lock, do
not resolve a tag or branch again, and leave the lock bytes unchanged.

```bash
curator project resolve .   # freeze declared sources into Skillfile.lock.json
curator install .           # materialize the locked snapshot (repairs drift)
curator status .            # compare installed state against the lock (--check, --json)
curator project refresh .   # re-resolve refs, bytes, and membership explicitly
curator install .           # install the refreshed lock (refresh alone changes nothing live)
```

Launch and status never rescan collections, advance branches, or
replace a local snapshot: installed shims execute the frozen runtime
until an explicit refresh plus install republishes it. A Skillfile
edited after resolve fails `source_lock_stale` until refresh; running
install again after refresh repairs drifted bytes from the lock while
the lock bytes stay identical. Every stable `source_*` and
`repository_*` failure prints a sanitized remediation naming the fix;
see [Skillfile source and transport
diagnostics](troubleshooting.md#skillfile-source-and-transport-diagnostics).

### Local sources

A `path` source freezes admitted filesystem bytes, including dirty,
staged, and untracked files; it never means Git HEAD. Relative paths
resolve against the declaring Skillfile's directory:

```json
{
  "schema_version": 2,
  "sources": {
    "local": {"path": "./pkgs"}
  },
  "skills": [
    {"name": "review", "from": "local", "directory": "review"}
  ]
}
```

An absolute path names shared machine content instead:

```json
{"path": "/work/shared-agents"}
```

Shared manifests with absolute paths stay intentionally
machine-specific, but the lock never copies those paths into identity
fields. Keep authored packages out of managed output (`.agents`,
adapter directories, the manager home): selecting a package inside
managed output fails `source_output_overlap`. Selecting the project
root as the package requires explicit `root_inputs` for that alias in
machine `source-policy.json` (see below).

### Git sources

A Git source pins exactly one of `tag`, `branch`, or `revision`
(branch only in the root project). Every member from one alias uses
the same resolved commit:

```json
{"git": "https://example.org/kit.git", "tag": "v1.2.0"}
```

```json
{
  "schema_version": 2,
  "sources": {
    "team": {"git": "https://example.org/kit.git", "tag": "v1.2.0"}
  },
  "skills": [
    {"name": "review", "from": "team", "directory": "skills/review"}
  ]
}
```

The logical spelling names the canonical identity and resolves
endpoints through machine policy:

```json
{"repository": "example.org/kit", "branch": "main"}
```

Without a policy entry for that identity, a logical declaration fails
`repository_endpoint_unavailable`; a URL declaration without an entry
attempts the declared URL once with no fallback.

### Collections

A collection selects immediate child directories of one source
directory. `include` names literal folders or `"*"`; `exclude` removes
literals afterwards. Every remaining directory must carry valid
SKILL.md frontmatter or the whole operation fails:

```json
{"from": "team", "directory": "skills", "include": ["review", "docs"]}
```

```json
{"from": "team", "directory": "skills", "include": ["*"]}
```

```json
{"from": "team", "directory": "skills", "include": ["*"], "exclude": ["release"]}
```

Overlapping installed names fail `source_name_conflict`.

### Machine policy setup

`source-policy.json` is operator-owned and lives beside the manager
configuration (next to `config.json`, outside package-controlled
trees). It maps canonical `host/path` entries to one or two endpoints
plus fallback, and admits root packages via `root_inputs`:

```json
{
  "schema_version": 1,
  "repositories": {
    "example.org/kit": {
      "endpoints": [
        {"url": "git@example.org:kit.git", "authentication": "team-ssh"},
        {"url": "https://example.org/kit.git", "authentication": "team-https"}
      ],
      "fallback": "availability-auth"
    }
  },
  "root_inputs": {
    "project": ["SKILL.md", "agent-skill.json", "references", "scripts", "build"]
  }
}
```

Without `pin`, endpoints are attempted in list order; `pin` selects
exactly one listed URL and forbids fallback. `fallback:
"availability-auth"` permits the next endpoint only after a positively
classified availability or authentication failure — TLS, host-key,
identity, ref, audit, and unclassified failures never fall back.
Named authentication providers (`source-providers.json` beside the
same configuration) apply to the external build lane under
`CURATOR_DRAFT_TRANSPORT_RESOLUTION`, not to Skillfile sources:
`project resolve` and `project refresh` clone and fetch with
credentials from the invoking environment only (SSH agent via
`SSH_AUTH_SOCK`, `GIT_ASKPASS` — the only HTTPS credential channel on
this lane: it answers git's username and password prompts (or embed the
username in the endpoint URL), it supplies credentials, it cannot
redirect the endpoint); user and system Git configuration — credential
helpers, `insteadOf`/`pushInsteadOf`, URL rewriting and includes,
`core.sshCommand` — is never consulted, and no interactive prompt
occurs. Git's environment on this lane is built from an explicit
allow-list (`PATH`; `HOME`/`USERPROFILE` only so OpenSSH finds its
default `known_hosts` and default identity files; `TMPDIR`/`TMP`/`TEMP`;
`TZ`; Windows process essentials; `SSH_AUTH_SOCK`; `GIT_ASKPASS`):
every other ambient name — `GIT_SSH`/`GIT_SSH_COMMAND`,
`GIT_PROXY_COMMAND`, `GIT_EXEC_PATH`, every other `GIT_*` override, and
the proxy environment — is dropped by construction and cannot redirect
or hijack the clone, so this source acquisition path does not honour proxy
environment. SSH runs a curator-owned command with an empty ssh config
(`-F` an empty file, `BatchMode`, `StrictHostKeyChecking=yes`,
`ProxyCommand=none`, `ProxyJump=none`, forwarding disabled, host-key
validation never disabled): host aliases, `ProxyCommand`, `IdentityFile`
and every other `~/.ssh/config` entry are not read — use an agent (or a
default-named key) and the literal host name; unknown hosts fail closed
(seed `known_hosts` with `ssh-keyscan` or a first manual `ssh`). The
`ssh` binary itself resolves through the honoured `PATH` (the lane's
tooling bound). The `authentication` identifier on a Skillfile endpoint is
validated but unused on this lane. Never put secrets in these
objects. An invalid or unreadable policy fails
`repository_policy_invalid` and is never treated as absent. On
Windows the isolation also drops Git for Windows' own system
configuration (`http.sslBackend`, `http.sslCAInfo`); whether HTTPS
verification still works there depends on the build's compiled
defaults and is unverified.

For a package at the project root, each `root_inputs` entry is a
source-relative, portable, link-free path disjoint from outputs; every
listed path must exist and cover SKILL.md, the manifest, and every
required context, runtime, and build input. Changing `root_inputs`
requires explicit refresh. Normal nested packages need no list.

## Skill package validation

### curator skill check

`curator skill check` validates a skill package directory against schema specifications.

Synopsis:

```bash
curator skill check <dir> [flags]
```

Flags:

- `--locale string`: validate against a locale.
- `--json`: output JSON.

Validate a local skill package directory:

```bash
curator skill check ./my-skill --json
```

The command verifies manifest schema conformance and prints validation errors.

## Global scope

Global commands manage machine-wide skill installations shared across projects.

### curator global init

Synopsis:

```bash
curator global init
```

Prerequisite: `curator bootstrap --skills-root <dir>` (see [curator bootstrap](#bootstrap)). Every `global`, `profile`, and `env` subcommand needs the global config that bootstrap writes; on a clean home they exit 1 with `global config not found: <path>` followed by the hint to run bootstrap first. Group-level help (`curator global --help`, `curator profile --help`, `curator env --help`, or `<group> help`) works without config. Subcommand flag help (for example `curator env resolve --help`) requires config and retains its flag usage and exit code 2.

Initialize the global Skillfile after bootstrap:

```bash
curator bootstrap --skills-root "$HOME/skills"
curator global init
```

The command creates the global Skillfile under the Curator home directory.

### curator global add

Synopsis:

```bash
curator global add <name> [options]
```

Options:

- `--branch string`: git branch.
- `--git string`: git clone URL.
- `--revision string`: git revision.
- `--source string`: source directory under `skills_root`.
- `--tag string`: git tag.

Add a global skill declaration:

```bash
curator global add helper --git https://github.com/example/helper.git
```

The command records the declaration in global configuration and installs the skill.

### curator global remove

Synopsis:

```bash
curator global remove <name>
```

Remove a global skill declaration:

```bash
curator global remove helper
```

The command removes the global skill declaration and deletes installed files.

### curator global list

Synopsis:

```bash
curator global list
```

List installed global skills:

```bash
curator global list
```

The command prints declared global skills and their installed revisions.

### curator global status

Synopsis:

```bash
curator global status [flags]
```

Flags:

- `--check`: exit non-zero unless every skill is up to date.
- `--json`: machine-readable output.

Check global scope status:

```bash
curator global status --check
```

The command validates global skill installations and exits non-zero on drift.

### curator global install

Synopsis:

```bash
curator global install [flags]
```

Flags:

- `--audit`: run the audit gate in advisory or strict mode.
- `--build-ssh-agent string`: agent socket for external SSH build repositories, or `auto` for environment agent (or `CURATOR_BUILD_SSH_AGENT`).
- `--build-ssh-identity string`: identity file for external SSH build repositories (or `CURATOR_BUILD_SSH_IDENTITY`).
- `--build-ssh-known-hosts string`: host keys external SSH build repositories are verified against (or `CURATOR_BUILD_SSH_KNOWN_HOSTS`).
- `--dry-run`: plan work without modifying files.
- `--fix-gitignore`: append missing managed gitignore entries.
- `--strict-tags`: fail if an installed tag moved to another commit.
- `--verbose`: print detailed progress.

Install global skills:

```bash
curator global install
```

The command materializes global skills into the machine home store.

### curator global adopt

Synopsis:

```bash
curator global adopt <command> [--dry-run]
```

Take an existing user-bin forwarding shim under Curator management when it is
a regular file whose bytes exactly match the canonical shim for an installed
global command. A successful adoption copies the original with its mode and
modification time preserved into the Curator backups directory, then records
the command in `.curator-managed.json`. The source entry is never rewritten.
Symlinks, special files, missing commands, and different bytes are refused.

Preview and adopt a command:

```bash
curator global adopt task-board --dry-run
curator global adopt task-board
```

`--dry-run` performs the same checks without creating a backup or changing the
ownership marker.

### curator global update

Synopsis:

```bash
curator global update
```

Update global source repositories:

```bash
curator global update
```

The command fetches latest remote state for global skill repositories.

### curator global upgrade

Synopsis:

```bash
curator global upgrade [flags]
```

Flags:

- `--audit`: run the audit gate in advisory or strict mode.
- `--build-ssh-agent string`: agent socket for external SSH build repositories, or `auto` for environment agent (or `CURATOR_BUILD_SSH_AGENT`).
- `--build-ssh-identity string`: identity file for external SSH build repositories (or `CURATOR_BUILD_SSH_IDENTITY`).
- `--build-ssh-known-hosts string`: host keys external SSH build repositories are verified against (or `CURATOR_BUILD_SSH_KNOWN_HOSTS`).
- `--dry-run`: plan work without modifying files.
- `--fix-gitignore`: append missing managed gitignore entries.
- `--strict-tags`: fail if an installed tag moved to another commit.
- `--verbose`: print detailed progress.

Upgrade global skill packages:

```bash
curator global upgrade
```

The command resolves updated revisions for global skills and installs them.

## Hybrid scope

Hybrid commands manage machine-stored skills activated per project.

### curator hybrid add

Synopsis:

```bash
curator hybrid add <name> [options]
```

Options:

- `--branch string`: git branch.
- `--git string`: git clone URL.
- `--revision string`: git revision.
- `--tag string`: git tag.
- `--target string`: target alias, absolute path, or glob.
- `--targets string`: comma-separated targets (alias, absolute path, or glob).

Add a hybrid skill declaration:

```bash
curator hybrid add helper --git https://github.com/example/helper.git --target project-a
```

The command stores the skill package in machine storage for selective activation.

### curator hybrid remove

Synopsis:

```bash
curator hybrid remove <name>
```

Remove a hybrid skill declaration:

```bash
curator hybrid remove helper
```

The command removes the hybrid skill from machine storage.

### curator hybrid list

Synopsis:

```bash
curator hybrid list
```

List hybrid skills:

```bash
curator hybrid list
```

The command displays stored hybrid skills and activation targets.

### curator hybrid status

Synopsis:

```bash
curator hybrid status
```

Check hybrid state status:

```bash
curator hybrid status
```

The command reports status (`installed`, `content-drift`, or `not-installed`) for declared hybrid skills across project targets.

## Audit and security

### curator audit

`curator audit` runs security audits, pins trust, or publishes signed audit records.

Synopsis:

```bash
curator audit [target] [flags]
```

Flags:

- `--all`: audit all configured projects and global skills.
- `--allow string`: pin trust for a content hash.
- `--global`: audit global skills.
- `--json`: machine-readable output.
- `--publish string`: signed audit record (JSON file) to submit.
- `--reason string`: reason for `--allow`.
- `--registry string`: registry base URL for `--publish`.
- `--token string`: auditor token for `--publish` (or `CURATOR_REGISTRY_TOKEN`).

Run audit check on current project:

```bash
curator audit --all --json
```

The command evaluates security policies and fails if unvetted code is detected.

Script commands additionally report two warning classes that never fail
the audit: `script-command-declared-only` for commands without
`execution_policy: "script-worker-v1"`, and
`script-command-unfiltered-declared-network` for enforced commands with
declared `network` hosts (reporting-only). Every script command also gets
an informational audit-record entry — `audit info: <skill>: command
'<name>' execution_policy=<script-worker-v1|(none)>` — carried under
`script_policies` in `--json` output and in the stored verdict, so a
reviewer can separate enforced commands from declared-only ones even when
no warning fires. See
[Troubleshooting](troubleshooting.md#script-audit-labels).

## Maintenance and shell integration

### curator gc

`curator gc` sweeps unreferenced runtime store entries and compiled build cache artifacts.

Synopsis:

```bash
curator gc
```

Run garbage collection pass:

```bash
curator gc
```

The command acquires the manager home lock and removes unreferenced artifacts older than 24 hours.

### curator shell-init

`curator shell-init` prints or installs shell integration hooks.

Synopsis:

```bash
curator shell-init [shell] [flags]
```

Flags:

- `--install`: cache the hook and print its optional profile source command.
- `--no-global`: skip global env sourcing.

Print zsh integration hook:

```bash
curator shell-init zsh --install
```

The command outputs shell code for environment auto-switching.

### curator hook approve

`curator hook approve <path>` records the digest of a project env file's
current bytes as `approved_by: operator` (the direnv-allow equivalent,
Manager profile §8.3), so the shell hook sources it without warning.
Re-run it after the file changes: the old record never authorizes the new
bytes. The path resolves to the canonical absolute identity the hook uses.
The command fails without recording when the file is absent or unreadable,
with distinct exit text for the two.

Synopsis:

```bash
curator hook approve <path>
```

Approve the env file a hook warning names:

```bash
curator hook approve ~/projects/app/.agents/env.sh
```

### curator hook approvals

`curator hook approvals` lists every approval record read-only: one
tab-separated path, sha256, approved_by, approved_at row per record, sorted
by path. It never mutates state; a malformed record is reported and skipped.

Synopsis:

```bash
curator hook approvals
```

### curator hook revoke

`curator hook revoke <path>` removes the approval record for the path. On a
path with no record it leaves state unchanged and reports that there was
nothing to revoke (exit 0).

Synopsis:

```bash
curator hook revoke <path>
```

### curator env migrate

`curator env migrate` runs the explicit credential migration step (Spec
environments §7.4/§10.1, manager §12.4): a mode change or a native-root
correction that leaves a recorded credential link behind is never applied
by `env resolve --repair` — repair re-links an absent path and refuses
anything else, pointing here. Migration inventories the old marker, every
recorded link target, and both Pi roots (`~/.pi/auth.json` and
`~/.pi/agent/auth.json`); preserves the effective mode; relinks recorded
symlinks at the declared native path and unlinks stale recorded links;
and never copies, moves, or rewrites credential bytes.

Synopsis:

```bash
curator env migrate --inspect|--plan|--apply [--profile <name>] [--env <env-id>] [--expect <plan-hash>]
```

- `--inspect` prints the read-only inventory per profile and environment:
  the marker, the effective mode, every recorded link with its current
  target and classification, and both Pi roots. It takes no lock and
  changes nothing.
- `--plan` prints the exact operations with the plan hash, plus every
  conflict with the operator choice it needs. Conflicts block `--apply`,
  not `--plan`.
- `--apply` requires the `--expect` hash of a prior `--plan` (without it
  the command is a usage error and the library refuses before any
  mutation). Under the manager-home mutation lock it recovers any
  leftover journal from an interrupted apply back to the prior state,
  re-inventories, prints the locked, revalidated plan *before* the first
  mutation (an undeliverable print fails before anything changes),
  refuses when `--expect` no longer matches the inventory (plan drift —
  the hash covers the marker identities too) or when conflicts remain,
  then journals the intent and executes exactly the printed operations —
  each relink as a temp-link + atomic rename — publishing changed
  markers as one journaled transaction. A syscall failure mid-op or a
  failed marker publication reverts the links; a kill between mutations
  leaves the journal for the next `--apply --expect <plan-hash>` to
  recover.

Migrate one Pi home, printing the plan before applying it:

```bash
curator env migrate --plan --env pi
curator env migrate --apply --expect <plan-hash> --env pi
```

Conflicts that need the operator — an isolated→shared account choice, a
regular file at a link path, two live Pi credentials — refuse with
`environment_credential_conflict` naming the exact out-of-band decision;
resolve them and re-run. The v1 marker still records path and strategy
only.

### curator env unmanage

`curator env unmanage` returns marker-recorded in-place surfaces to native
ownership (Spec environments §8.3, §8.4.1, and §9.2). By default it visits
every registered environment; `--env` limits the operation to one. It
removes only surfaces recorded by each environment marker, clears the
recorded current profile for the selected scope, and removes the marker.

Synopsis:

```bash
curator env unmanage [--restore-backups] [--env <env-id>] [--target <target-id>]
```

With `--restore-backups`, the newest readable takeover generation is copied
back to its home-relative paths before the marker is removed. A missing
backup inventory restores nothing. An inventory that cannot be established
stops before native-home mutation with
`environment_backup_record_unreadable`. Without the flag, generations are
left in place and their path is printed as a notice. Credential files and
unrecorded files are never touched. `--target` is accepted by the command
syntax, but secondary fixed-home writes remain deferred until their home
resolution and surface-write path are implemented.

Unmanage one environment and restore its newest takeover generation:

```bash
curator env unmanage --restore-backups --env claude_code
```

### curator ui

`curator ui` opens an interactive terminal view over installed environment state.

Synopsis:

```bash
curator ui
```

Launch terminal interface:

```bash
curator ui
```

The command displays interactive status dashboards for project skills.

## Profile graph contract

### curator profile graph

The command accepts these selectors:

```text
curator profile graph <profile> [--json] [--source <dir>] [--env <id>]
```

This read-only, offline command emits the machine contract
[`curator-profile-graph/1`](../internal/profilegraph/schemas/curator-profile-graph-v1.schema.json).
The [small command golden](../internal/profilegraph/testdata/small/graph.json)
and [shared DAG/overlay golden](../internal/profilegraph/testdata/dag-overlays/graph.json)
are captured through real command dispatch over synthetic immutable snapshots.
The original illustrative graph remains `small/contract-example.json`.
The executable grammar is in `cmd/curator/command_help.go`; this command does
not change the separately versioned environment protocol.

`<profile>` selects an installed profile by name. Its stored lock, recorded
composition and exact local snapshots are authoritative. A newer checkout,
registry entry or branch tip must never replace a locked input. With
`--source <dir>`, the positional name labels a preview of that package tree;
an installed profile with the same name is ignored. Preview uses install's
manifest, requirement, directory, selector and precedence semantics but a
read-only local resolver. Only already available Git objects and state
snapshots may satisfy requirements. A missing revision or an unresolved
floating requirement refuses; no fetch, clone, update, installation or lock
publication is permitted. Local Git reads disable lazy fetching and optional
locks. No hooks, builds, package scripts, runtime launches or subprocess
credential helpers may execute.

The source preview needs no bootstrap or installed profile. Installed mode
reads only the admitted profile record and environment projection; it must not
invoke provisioning or credential checks while obtaining those inputs.

`--env` selects one registered environment; `claude` and `codex` normalize to
`claude_code` and `codex_cli`. Omission selects all environments in the admitted
environment catalogue, sorted by canonical id, including environments with
no emitted surface. An unknown id refuses before reading package content.
`--json` writes exactly one UTF-8 JSON object and LF to stdout. Without it,
the text view reports the same graph and per-document totals. Diagnostics go
to stderr. Success is exit 0, invalid arguments exit 2, and input, safety,
resolution or resource failures exit 1 with a named diagnostic class. A
failure emits no partial graph. No output file, temporary file, cache, home,
lock, ref, index, config, current pointer or telemetry record is written.

### Sources, nodes and relationships

Each node has a stable `id`, an `identity` tuple, a `kind`, source provenance
when applicable, a nullable owning `layer`, content measurements, per-document
`emissions` and per-environment `loads`. Source identities preserve the logical
locator, declared version, package directory selector, exact commit or state digest and relative
file path. `authoring_source` is a separately declared editing destination;
a cached snapshot alone cannot prove ownership. Unknown ownership is explicit.
The output uses logical root labels and relative paths, never absolute home
paths, credentials or URL userinfo. Root-label bindings are caller-held input
selectors, not instructions for a consumer to search the filesystem.

| Node kinds | Meaning |
| --- | --- |
| `profile`, `root-package` | Final selected composition and its root snapshot |
| `overlay` | A declaration occurrence, including source, requirement form, directory, resolved revision, effective weight and emitted order |
| `profile-package`, `context-package` | Nested profile or required context snapshot; one shared dependency can have several incoming edges |
| `instruction-module`, `section`, `quote-block` | Manifest module occurrence and Markdown source spans; sections retain heading level/text |
| `skill`, `skill-listing`, `skill-body`, `skill-reference` | Dependency, decoded listing payload, body and referenced resource measured separately |
| `project-instruction` | Only a project instruction file explicitly admitted by Curator's selected scope |
| `emitted-document`, `generated` | Environment output and renderer-owned header, marker, separator or wrapper bytes |

Edges are directed: `root`, `overlay`, `requires-profile`, `requires-context`,
`requires-skill`, `contains`, `section`, `quote`, `lists`, `body`, `references`,
`project`, `emits`, `contributes`, and `requested-load`. Each carries a zero-based
order within that relationship, and dependency edges retain the declared
requirement form/value. Overlay order is the final renderer member position,
not merely its position in configuration. Its declaration slot is also kept.
Equal source/revision pairs with different directories, weights or slots remain
distinct. Shared dependency edges do not create duplicate package nodes; module
occurrences reached through distinct render slots remain distinct even when
their content hashes match. Real dependency cycles refuse as `graph_cycle`.

Every selected context/profile package carries its effective weight and renderer order in
`composition`; version is null when no declared package version is available.

Node ids are `n:` followed by SHA-256 of the identity tuple: concatenate each
UTF-8 component as its decimal byte length, `:`, then its bytes, without a
separator. The tuple is `[kind, logical-profile-name, ...coordinates]`.
Package coordinates are locator, directory, revision kind, revision value.
An overlay appends its declaration slot and weight. Modules append their
owning layer id, manifest slot and relative path; sections and quotes append
the module id and source start/end byte offsets. Skill parts append skill id,
part kind and relative locator. Documents append environment and relative
output path; generator nodes append document id, part kind and ordinal.
The profile has no extra coordinates. A skill package uses package coordinates;
a project file uses its logical root label and relative path. Generated bytes
have no source file identity. Structural nodes use null source/content metrics
where there is no corresponding file. Root labels and all coordinate orderings
are deterministic. Content hashes identify blobs, never occurrence identities.
Changing a pin changes its snapshot ids; moving an absolute checkout does not.
Nodes sort by id, edges by `(from, kind, order, to)`, documents by
`(environment, path)`. Duplicate ids or tuples, dangling ids and a layer id
that does not name a package/overlay refuse as `graph_invalid_input`.

### Byte accounting and load evidence

`content` measures exact original bytes, with an unprefixed lowercase SHA-256.
Text lines equal LF count plus one for a nonempty unterminated final line;
empty text has zero lines. CRLF is two bytes with one LF line delimiter.
Binary content has a byte count and hash but null lines. `unread` content has
null metrics/hash, `absent` has zero bytes/lines and null hash, and structural
`not-applicable` content has null metrics/hash. These cases never impersonate
an empty measured file. Source byte ranges are relative to the source file.
All ranges use zero-based half-open UTF-8 byte offsets, never character indexes.

Every emitted document has exact `bytes`, `lines`, `sha256`, environment,
logical root/path, form and origin. Installed mode compares existing managed
documents to rendering from the exact lock; differences refuse as
`graph_document_drift`. A missing provisioned document may be represented by
an in-memory render with origin `rendered`; it must never be called installed.
Source mode always renders in memory. Referenced form includes the root and
every emitted fragment as separate documents, as does an admitted system prompt.
No document row means no emitted surface, not a missing zero-byte file.

`documents[].coverage` is the sole additive ledger. Its entries partition
`[0, bytes)` in order without gaps, overlap or zero-length spans. Each names
one leaf contributor; the sum of `end-start` must equal the exact file size.
For empty documents coverage is empty. Header comments, chapter markers,
separators and generated imports belong to `generated` nodes. Current
`contextmaterialize.Join` preserves validated module bytes and inserts one LF
between parts. Neither a final source LF nor generator framing is silently
charged to another module. A transformed/wrapped part retains source metrics
and maps only its actual emitted bytes; wrappers are separate contributors.

`nodes[].emissions` provides inclusive projections onto each document. Within
one projection ranges are ordered, disjoint and merged when adjacent. Its
bytes/lines/hash measure the concatenation of those slices; its byte count
equals their union length. Ancestor, section and quote ranges may overlap other
nodes. Never sum inclusive projections or line counts to obtain document totals.
An emitted-document node covers the whole document. Module projections include
their section and quote bytes but exclude renderer framing. Empty emission
lists explicitly mean no emitted contribution. The renderer, not filename or
text matching, establishes layer and module attribution.

Section parsing recognizes ATX and setext headings, retains the preamble as
module-owned bytes, and ignores heading-like content inside backtick or tilde
fences. A section extends from its heading to the next heading of the same or
shallower level, or module end. Child sections overlap ancestors. Quote-block
nodes cover contiguous Markdown blockquote spans including their markers;
they do not assert that a quotation is an operative instruction. Semantic quote,
duplicate, conflict, token and trimming analysis belongs to consumers. Consumers
can read admitted text and Markdown; they need not parse Curator manifests,
locks or composition headers to reconstruct ownership or costs.

Load classes `always`, `on-demand` and `listed-only` describe only a known
Curator rendering rule for that environment. Every such claim cites
`basis: rendering-rule`; otherwise use `UNKNOWN`. State is separate:
`emitted`, `available`, `conditional`, `excluded` or `UNKNOWN`. Conditional
claims carry the known selector/trigger; an excluded module has no emission
and must not be mislabeled on-demand. A `requested-load` edge records an
explicit declared subsequent read, never proof of runtime injection. There
are no observed prompt counts or runtime loading claims in this contract.

Skills carry origin `lock`, `managed-inventory`, `runtime-discovered` or
`UNKNOWN`. Runtime-discovered means a safely admitted Curator-owned inventory
projection, never session scraping or unrestricted home discovery. A listing's
canonical content is decoded name, LF, decoded description, LF; raw YAML is
not that payload. Listing bytes are not charged as emitted instructions unless
Curator actually emits them. `materialization: listed-only` means only a
listing is represented; `loaded` means that part is embedded in a Curator
document, not that a model has read it. Unknown exposure stays `UNKNOWN`.
Bodies/references need not be opened to report their declared presence. Binary
references cannot have text lines. No body bytes inflate listing cost.
Project files are entries already selected by admitted Curator project scope
metadata. With no such selection, the graph has no project-instruction nodes.
The retained contract example explicitly admits one project input. Arbitrary
parent chains,
memory, runtime extras and missing-file discovery are not inferred. Package
projections include their contained modules, not dependency subtrees. Only
the final profile projection includes the complete composition.

`inputs` lists every consumed content/projection file by logical root and
relative path, byte count and SHA-256, sorted by `(root, path)`. The input
digest hashes this UTF-8 transcript: `curator-profile-graph/1` plus LF, mode
plus LF, profile name plus LF, then each sorted canonical environment id plus
LF, then an empty line, then each input as `root NUL path NUL sha256 LF`.
Fields in this transcript must contain no NUL, CR or LF. Duplicate input keys
refuse. Include lock bytes and admitted source/composition/environment
projections, not unrelated configuration. File hashes preserve original bytes;
projection serialization is UTF-8 JSON with sorted keys, no whitespace and no
trailing LF. `generator.version` is the actual Curator build version, independent
of schema version; the example uses an explicitly synthetic version.
JSON Schema enforces shape. Hashes, reference integrity, identity tuples,
acyclic dependencies, load evidence and exact accounting are additional
producer obligations, not properties that JSON Schema alone can prove.

### Safety and bounded reads

Reads use a positive allowlist of manifests, locks, immutable local snapshots,
declared module files, skill listing metadata, explicit references and selected
project/output files. The only machine configuration read is a restricted
composition/environment projection. Do not deserialize unrelated settings.
No recursive crawl of a home or filesystem root is allowed.

The deny-list wins over declarations and explicit selectors. Under any home,
never open content in auth, credential, session, history or log directories;
never open auth/credential files, private keys, token stores, shell histories,
session transcripts, SQLite/database files (including WAL/SHM sidecars), logs,
or shell snapshots. Matching is case-insensitive on path components and known
file families. Match each basename case-insensitively, including directory
components: `.ssh`, `.gnupg`, `auth*` (including `auth.json`), `*credential*`,
`*token*`, `*secret*`, `*oauth*`, `*session*`, `history*`, `log`, `logs`,
`*.log`, `*.sqlite*`, `*.db*`, `*-wal`, `*-shm`, `*_history`, `.env` and `.env*`
(including `.envrc`), `*password*`, `*passwd*`, `.netrc`, `.npmrc`, `.pypirc`,
`.git-credentials`, `*.pem`, `*.key`, `*.p12`, `*.pfx`, `*keystore*`, `*.jks`,
`id_*`, `*keychain*`, `*keyring*`, `shell-snapshots` and `shell_snapshots`.
The database rules require a literal dot followed by `db` or `sqlite` anywhere
in the basename, then any suffix (including empty): `a.db`, `a.db3`,
`a.db-wal`, `a.db.backup` and `a.sqlite3` match; `database.md`,
`database-notes.md` and `sqlite-notes.md` do not. `auth*` remains prefix-only,
so `my-auth.md` is loadable. Apply the same payload restrictions to source and cache
roots so a source selector cannot bypass them. Benign names matching these
families, such as `token-budget.md` or `author.md`, are denied too; declarations
provide no safe-name exception. Rename such modules to an unambiguous name.
Metadata classification may inspect path kind; denied payloads must never be
opened, hashed, included or echoed in errors. An explicitly declared denied
input refuses as `graph_path_denied`. Whole-tree source/state walks skip denied
siblings and directories before opening them; they contribute neither hashes
nor input rows. The remaining tree must still match any installed state pin;
a pin that included excluded payloads refuses rather than borrowing that pin.
Do not inspect auth/keychain state to resolve uncertainty.

Resolve every path component beneath explicitly admitted managed/source/project
roots before opening. A symlink whose target leaves those roots, including an
intermediate directory link, refuses as `graph_symlink_escape`. Internal links
may resolve only to allowlisted files and must be cycle-checked. Enforce the
boundary during open to prevent symlink replacement races. FIFOs, sockets,
devices and other non-regular payloads refuse as `graph_invalid_input`.

Version 1 caps serialized stdout at 8 MiB including its final LF, consumed
input bytes at 32 MiB, each file at 4 MiB, nodes at 20,000, edges and inspected
directory entries at 50,000 each, environments/documents at 256 each, and
dependency/link traversal depth at 128. Limits apply to text and JSON alike.
Check bounded lengths before allocation/read and budget all traversals, not
just final output. Exceeding any limit refuses as `graph_limit_exceeded`,
never truncates costs or emits a partial graph. Missing input is
`graph_missing_input`; malformed or inconsistent input is `graph_invalid_input`.
`graph_load_unknown` is a nonfatal explanation of uncertainty. All diagnostics
must avoid secret-bearing paths or values.

### Decisions

| Decision | Reason | Date |
| --- | --- | --- |
| Contract only, existing CLI reference and internal schema fixture package | Allow consumer review without advertising an executable command | 2026-11-09 |
| Closed `curator-profile-graph/1` schema; build version separate | Reject silent shape drift and distinguish generator changes | 2026-11-09 |
| Length-prefixed tuple hashes identify occurrences; content hashes identify blobs | Stable across checkout locations without collapsing distinct declarations | 2026-11-09 |
| Package DAG with ordered occurrence edges and explicit owning layer | Preserve shared dependencies, nested profiles and overlay ownership | 2026-11-09 |
| Half-open byte unions; additive leaf coverage separate from inclusive projections | Exact module cost with overlapping sections and quotations | 2026-11-09 |
| Generator owns headers, separators and wrappers; source metrics stay separate | Reconcile concatenation and transformation without hidden overhead | 2026-11-09 |
| Physical LF lines with unterminated tail; binary/unread/null distinct | Avoid character offsets and fabricated text measurements | 2026-11-09 |
| `UNKNOWN` means Curator has no applicable rendering evidence | Inventory and emission cannot prove actual model injection | 2026-11-09 |
| Separate load class, conditional/excluded state and skill materialization/origin | Keep selection, listing, embedded content and runtime extras distinct | 2026-11-09 |
| Structured Markdown spans, no semantic quote verdict or token estimates | Publish composition truth while keeping analysis in consumers | 2026-11-09 |
| Locked installed inputs versus offline source preview; detect installed drift | Never attribute current source bytes to an older installed lock | 2026-11-09 |
| Logical relative locators and sorted input transcript digest | Reproduce provenance without exporting home paths or unrelated config | 2026-11-09 |
| Fixed read/output budgets, deny-list and named symlink refusal | Bound offline work and prevent sensitive or out-of-root reads | 2026-11-09 |
| Separate package, occurrence and content identities (D02) | Directory, declaration slot and weight preserve distinct overlay occurrences; DAG reuse retains one shared package | 2026-11-10 |
| Declared requirement separate from resolved provenance (D03) | Keep requirement form, nullable version, revision kind, directory and explicit unknown ownership without guessing an editing destination | 2026-11-10 |
| Atomic exit protocol and seven fatal classes (D11) | Exit 0 emits one graph, exit 2 is usage, exit 1 emits one named fatal diagnostic and no graph; only `graph_load_unknown` is nonfatal | 2026-11-10 |
| V1 limits may rise but never fall (Q3) | Lower limits or changed semantics require `/2`; injectable lower budgets are test controls only | 2026-11-10 |
| Declared unread references have null bytes and hash (D06) | No metadata-derived size or invented content measurement in v1 | 2026-11-10 |
| Instrument the shared emitter and reuse resolve input/pin algorithms | Derive module attribution from emitted parts, preserving byte accounting without reimplementing composition | 2026-11-10 |
| Read local Git objects directly with bounded no-follow reads | Avoid Git configuration, hooks, lazy fetching, helpers, locks and writes during graph capture | 2026-11-10 |
| Deny basename families before payload reads; skip undeclared siblings and refuse declared ones | Protect `.env*` (including `.envrc`), embedded token/secret/credential/OAuth/password/passwd names, credential dotfiles (`.netrc`, `.npmrc`, `.pypirc`, `.git-credentials`), keys and containers (`*.p12`, `*.pfx`, `*keystore*`, `*.jks`), database extensions (`*.db*`, `*.sqlite*`), sidecars, sessions, histories, logs and keychain/keyring files on every read route. Database matching requires the literal extension marker `.db` or `.sqlite`, so `database.md` stays loadable. Deny benign matches such as `token-budget.md`, `.environment`, `my-password-notes.md`, `keystore-notes.md` and `a.db.backup` too; `.p12`, `.pfx` and `.jks` suffixes also deny ordinary documents with those extensions. Exact credential dotfiles leave `.npmrc-notes.md` loadable; `auth*` stays prefix-only and permits `my-auth.md`. An explicit declaration cannot prove safe content, and renaming is safer than a bypass | 2026-11-10 |
| Match path overlays through the shared resolver state seed, package key and state pin | Path/state lock members intentionally have empty source/directory. Reuse package projection and versioned tree hashing, retain declaration slot and weight, and reject mismatched names, state bytes or weights. Installed mode validates the bounded declaration tree against the pin but renders locked snapshot bytes | 2026-11-10 |

### Open contract questions

A safe source of runtime-discovered inventory remains unadmitted. The producer
does not discover runtime extras; their schema origin remains representable
for a future explicitly admitted projection. This does not block locked skills
or source preview.

## Operator configuration

### Claude settings permissions

The `env config` knob table adds this machine-owned profile setting:

| Knob | Values | Default | Scope |
| --- | --- | --- | --- |
| `claude_settings_mode.<profile>` | `native`, `read-only` | `native` | Managed `claude_code` homes on Unix; not system-lockable |

```sh
curator env config set claude_settings_mode.example read-only
curator env resolve claude_code --profile example --repair
curator env config show claude_settings_mode.example
curator env config unset claude_settings_mode.example
```

`read-only` creates missing `settings.json` as `{}` and publishes it, and any
present `cowork_settings.json` / legacy `.config.json`, as regular `0400` files.
Existing bytes are preserved. Status and resolve without repair report write-bit
drift as non-current; repair atomically replaces the entry with its final mode,
without a writable interval on the live inode. Unset stops enforcement and
preserves existing files. Invalid values report
`environments.claude_settings_mode.<profile>: must be one of native, read-only`.
See the [environments contract](environment-config.md#read-only-claude-settings-admission-contract-a)
for platform limits, refusal diagnostics, ownership and atomic update rules.

### curator env config show

```bash
curator env config show
curator env config show permissions
```

Both forms print JSON, including for interactive use; there is no separate
`--json` flag. The full form retains the existing environments knobs and
adds `permissions_posture`, keyed by installed profile name:

```json
{
  "permissions_posture": {
    "interactive": {"mode": "yolo", "locked": false, "source": "profile"},
    "restricted": {"mode": "native", "locked": false, "source": "default"}
  }
}
```

A global permissions lock applies to every installed profile. The existing `permissions` map retains
its configured mode strings; `permissions_posture` uses the same resolver
as the fragment's `permissions` member. An absent profile entry reports
`{"mode":"native","locked":false,"source":"default"}`. A system lock
reports `{"mode":"native","locked":true,"source":"global"}`, even for
a profile absent from the system map.

`show permissions` prints just that object for the machine's current
profile. Without a current profile it exits 1 with an unavailable diagnostic.
Unreadable or malformed configuration exits 1 with a diagnostic and no
posture output. Failed profile-inventory or current-pointer reads also
exit 1 and report the posture as unreadable; they never imply an unlocked
policy. Reading posture neither provisions a runtime home nor needs an
adapter for the caller's runtime.

### curator config show

Synopsis:

```bash
curator config show
```

Display active configuration:

```bash
curator config show
```

The command prints current JSON configuration parameters.

### curator config launcher

```bash
curator config launcher
```

Export the effective selected machine configuration's launcher metadata as one
JSON object. Configuration discovery, trust and system-policy validation use the
same path as ordinary Core commands; `CURATOR_CONFIG` selects the same source.
The command takes no arguments. It reads no credential store, repairs no
managed home and starts no runtime. Invalid configuration produces a nonzero
exit and no projection.

The closed transport is `curator.launcher-config/1`, with `config_schema`,
`config_path` (the actual selected source) and `extensions`. Older supported
schemas export an empty `extensions` object. Schema 7 may export typed `proxy`
and `claude_auth` references; see [machine configuration](environment-config.md).
The output contains references, never credential material. Do not use it as an
authentication or proxy qualification result.

### curator config claude-auth

```bash
curator config claude-auth plan --keychain-service SERVICE --keychain-account ACCOUNT --out CANDIDATE
curator config claude-auth plan --file /absolute/token-file --out CANDIDATE
curator config claude-auth apply --keychain-service SERVICE --keychain-account ACCOUNT --expect-source-sha256 SOURCE --expect-result-sha256 RESULT --expect-policy-sha256 POLICY
curator config claude-auth restore --backup BACKUP --expect-current-sha256 CURRENT
```

Use the same store flags for plan and apply. Exactly one file reference or a
complete Keychain service/account pair is required; references are metadata.
The command never opens a token file or looks up a Keychain item. The selected
machine path follows ordinary discovery, including `CURATOR_CONFIG`. The
writer supports Unix owner/mode and no-follow semantics; other platforms
refuse this operation.

Plan supports schema 2 to 7 and schema 7 to 7, preserving existing proxy and
unrelated declarations. It refuses other schemas, including newer carriers,
and allocates no new schema version. A missing, nonregular, linked,
non-private, differently owned or system configuration source refuses. JSON
grammar and reference validation reject unknown/duplicate fields, inline
credentials, mixed stores and unsafe file paths before candidate/backup
creation. Both source and candidate must pass the ordinary loader and system
overlay with equal effective non-extension settings.

Plan creates only the explicitly named candidate with exclusive creation and
mode 0600. Its JSON output reports the selected path and selection source,
source/candidate/system-policy digests, schema transition, selected auth
references and preserved top-level key names. Other configuration values
appear only in that private candidate. Inspect its diff privately. A policy
digest binds the selected policy's identity and bytes, or its absence.

Apply recomputes the candidate and requires the exact three digests printed
by plan. It writes one exclusive 0600 exact-source backup named
`<config>.bak-<UTC timestamp>`, syncs the backup and its parent directory,
then writes and syncs one sibling temporary file before atomic rename.
No receipt, lock or journal is created, and a failed rename has no
remove-and-retry fallback. Preflight refusals create no backup. A failure
after backup publication but before rename leaves the source unchanged and
retains/reports the recovery copy. A successful rename reports `applied`
even if subsequent readback, parent-directory sync, output or cancellation
fails; `readback_verified` and `durable` report the actual post-rename checks.
An applied result with `durable=false` has unconfirmed directory durability.

Restore validates the named sibling backup's bytes, digest, regular type,
owner and mode, and binds the current source to `CURRENT`. It validates
ordinary loading/non-extension equivalence and equality of extensions other
than `claude_auth`. A later proxy change refuses even when `CURRENT` is its
freshly observed digest. Restore rechecks source, policy and backup before
replacement, and retains the backup. A changed source or
backup refuses; a successful rename reports `restored` with the same
durability/readback semantics. The optional library-level expected-backup
digest can additionally bind caller-held provenance.

Apply and restore require an exclusive finite configuration custody window.
Expected hashes and the last pre-rename recheck prove observed freshness;
rename is **not atomic compare-and-swap** against a nonparticipating writer.
`env config set` does not participate in a shared lock. A concurrent write
after the final recheck can therefore be overwritten. Directory fsync after
rename is the reported crash-durability boundary; it does not close that
concurrency gap. Configuration editing alone neither installs a launcher
nor demonstrates a successful authenticated runtime launch.

### curator config build-ssh

`curator config build-ssh` configures SSH credentials for external build repositories.

Synopsis:

```bash
curator config build-ssh add <scope> [--agent [SOCKET]] [--identity PATH] [--known-hosts PATH]
curator config build-ssh list
curator config build-ssh remove <scope>
```

Flags:

- `--agent`: use default agent socket or specify named socket path.
- `--identity`: set identity private key file path.
- `--known-hosts`: set host keys verification file path.

Add SSH credentials for a repository scope:

```bash
curator config build-ssh add git.example.com/portals --identity ~/.ssh/id_ed25519
```

The command binds specified SSH identity files to the matching host scope.

### curator config build-https

`curator config build-https` configures HTTPS credential resolution for external build repositories.

Synopsis:

```bash
curator config build-https add <scope> (--git-credentials | --keyring | --token-env NAME) [--username NAME]
curator config build-https login <scope> [--username NAME]
curator config build-https list
curator config build-https remove <scope>
```

Flags:

- `--git-credentials`: read token via Git credential helper.
- `--keyring`: use stored keyring token from login.
- `--token-env`: read token from specified environment variable.
- `--username`: username sent with token (default: `token`).

Add HTTPS token source for a repository scope:

```bash
curator config build-https add git.example.com/portals --token-env GITHUB_TOKEN
```

The command binds the environment token source to the repository host scope.

## Version

Synopsis:

```bash
curator --version
```

Print version information:

```bash
curator --version
```

The command prints Curator version string and build details.
