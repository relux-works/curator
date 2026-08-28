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
