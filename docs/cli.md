# Curator CLI Reference

Every synopsis and flag in this reference was verified verbatim against `./bin/curator` built via `make build` from the repository tree.

## Shared flags

Curator flags use consistent names and behavior across command groups.

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

## Environment initialization and skill lifecycle

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
curator bootstrap --if-missing --non-interactive
```

The command initializes `$HOME/.curator/config.json` without overwriting existing files.

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

For schema-2 projects with the draft opt-in switch set, `install`
materializes the locked snapshot and repairs drifted bytes from the
lock without touching the lock itself; see [Draft Skillfile
sources](#draft-skillfile-sources-opt-in-unreleased). Run `install -h`
with the switch set for the workflow summary.

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

The command resolves updated revisions, updates `Skillfile.lock`, and materializes skill packages.

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

For schema-2 projects with the draft opt-in switch set, `status`
compares installed state against the frozen lock only — it never
rescans collections, advances branches, or replaces a local snapshot;
see [Draft Skillfile sources](#draft-skillfile-sources-opt-in-unreleased).
Run `status -h` with the switch set for the workflow summary.

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

The command updates `Skillfile.lock` with resolved dependency commits and content hashes.

For frozen v1 projects the command prints the alias, project path,
Skillfile path, and managed skill and bin directories without modifying
disk state. For schema-2 projects with the draft opt-in switch set (see
[Draft Skillfile sources](#draft-skillfile-sources-opt-in-unreleased)),
it runs the explicit attempt: it acquires every Git source alias,
freezes local bytes and Git commits into a locked plan, and publishes
`Skillfile.lock.json` plus the machine bindings transactionally after
every gate succeeds. With `CURATOR_DRAFT_SOURCES_V1=1`, `curator project
resolve -h` prints the draft workflow; with the switch off the `-h` and
`--help` spellings keep the frozen v1 behavior (they resolve the current
project exactly as before). The bare word `help` is never a help flag:
it resolves as a project alias or path.

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

Refresh a draft lock after editing sources or after upstream refs moved:

```bash
curator project refresh .
curator install .
```

Refresh alone changes nothing live: run `curator install` afterwards to
materialize the refreshed lock. Launch and status never refresh on
their own. Run `curator project refresh -h` for the workflow (draft
switch on; with the switch off the flag spellings keep the frozen v1
behavior, as for `resolve`).

## Draft Skillfile sources (opt-in, unreleased)

Skillfile schema 2 with local and Git sources is draft functionality
behind the operator-owned opt-in switch. It is not part of the released
v1 behavior: package data can neither set nor observe the switch, and
with the switch off schema 2 fails closed while frozen v1 behaves
byte-identically.

```bash
export CURATOR_DRAFT_SOURCES_V1=1
```

### Workflow

Five existing verbs move a schema-2 project; no new command is added:

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
see [Draft source and transport
diagnostics](troubleshooting.md#draft-source-and-transport-diagnostics).

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
or hijack the clone, so the draft lane does not honour proxy
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

Initialize global configuration:

```bash
curator global init
```

The command creates global configuration files under the Curator home directory.

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

## Operator configuration

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
