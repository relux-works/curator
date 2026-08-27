# Cross-Platform Setup And Credentials Pattern

Reusable pattern for `relux-works` skill repos that ship both:

- a repo-local agent skill
- a companion CLI binary

The goal is to keep setup and credential handling consistent across projects so
agents and humans get the same commands, the same install shape, and the same
security rules.

## 1. First-Class Platform Matrix

Default support matrix for new repos:

- `macOS arm64`
- `macOS x86_64`
- `Windows x86_64`
- `Windows arm64`

Release artifacts should be produced for:

- `darwin/amd64`
- `darwin/arm64`
- `windows/amd64`
- `windows/arm64`

Linux is optional. Do not declare first-class Linux support until both setup and
credential storage have a tested story.

## 2. Setup Entry Points

Every repo should expose thin root wrappers:

- `./setup.sh`
- `./setup.ps1`

Those wrappers should delegate to:

- `scripts/setup.sh`
- `scripts/setup.ps1`

The wrapper stays tiny. All real logic lives under `scripts/`.

## 3. Setup Responsibilities

The macOS and Windows setup scripts should perform the same logical flow:

1. Resolve repo root.
2. Ensure Go is installed.
3. Derive build metadata from git when available.
4. Build the CLI with embedded version metadata.
5. Install the binary into a user-local bin directory.
6. Install a degitized skill artifact into `~/.agents/skills/<skill-name>`.
7. Refresh `~/.claude/skills/<skill-name>` and `~/.codex/skills/<skill-name>`.
8. Write install-state metadata into the app config directory.
9. Ensure the user-local bin directory is on `PATH`.
10. Verify the installed binary with no-side-effect commands.

Platform-specific bootstrap:

- macOS: prefer `brew install go`
- Windows: prefer `winget install GoLang.Go`

## 4. Install-State Contract

Store install metadata in:

- `os.UserConfigDir()/APP/install.json`

Expected fields:

- `repoPath`
- `installedSkillPath`
- `binDir`
- `platform`
- `arch`
- `version`
- `commit`
- `buildDate`
- `installOnly`

This file is operational metadata, not secret material.

## 5. Credential Sources

Keep the same source names across projects:

- `auto`
- `keychain`
- `env_or_file`

That lets projects change internals without changing the auth UX.

## 6. Default Source Policy

Recommended `auto` behavior for new repos:

- `darwin` -> `keychain`
- `windows` -> `keychain`
- everything else -> `env_or_file` until a system keyring path is verified

Interpret `keychain` generically:

- on macOS: Keychain
- on Windows: Windows Credential Manager or another OS-native secret store

`env_or_file` remains the explicit fallback for:

- CI
- one-off local overrides
- platforms without validated system keyring support

## 7. Credential Storage Contract

### `keychain`

Use system secret storage on supported desktop platforms.

- service name: stable app identifier
- account key: stable org/account/instance key
- stored payload: serialized credential object, not only the raw token

Minimum payload:

- `email`
- `token` or `api_token`
- `auth_type`

### `env_or_file`

Resolution order:

1. environment variables
2. global config file under `os.UserConfigDir()`

Config path convention:

- `os.UserConfigDir()/APP/auth.json`

Examples:

- macOS: `~/Library/Application Support/<app>/auth.json`
- Windows: `%AppData%\\<app>\\auth.json`

Keep the file scoped by organization or account key:

```json
{
  "profiles": {
    "acme": {
      "email": "agent@example.com",
      "api_token": "secret",
      "auth_type": "api_token"
    }
  }
}
```

Never use repo-local auth files for normal operation.

## 8. Auth CLI Contract

Keep the auth command tree stable:

- `auth set-access`
- `auth whoami`
- `auth resolve`
- `auth clean` or `auth clear-access`
- `auth config-path`

Recommended behavior:

- `set-access` writes to the default storage for the current platform
- `whoami` inspects local state and may perform a live auth probe
- `whoami --check=false` skips live network validation
- `resolve` reports where credentials would load from without printing the secret
- `clean` removes the stored entry for the selected org/account
- `config-path` prints the global auth-file location

If a low-level helper like `write-config` exists, document it as support-only.

## 9. Verification Contract

Every repo using this pattern should support:

```bash
go test ./...
<tool> version
<tool> auth config-path
```

Credential smoke flow:

1. `<tool> auth set-access ...`
2. `<tool> auth whoami --check=false ...`
3. `<tool> auth resolve ...`
4. `<tool> auth clean ...`

If live auth checks exist, they belong in `whoami` and must be opt-out.

## 10. Security Rules

- Prefer system secret storage on supported desktop platforms.
- Treat env vars as override or CI input, not the main desktop UX.
- Keep fallback files under `os.UserConfigDir()`, never inside the repo.
- Never commit tokens, debug dumps, or copied secret payloads.
- Keep the auth command surface stable even if the backend changes.

## 11. Copy-Forward Checklist

When applying this to another repo:

- rename binary, app id, service name, and config directory
- keep `auto`, `keychain`, and `env_or_file`
- keep the `auth/*` command family
- add both `setup.sh` and `setup.ps1`
- write install-state metadata
- ship release artifacts for the declared platform matrix
- test platform-default source selection and resolution order
- verify setup with a no-side-effect smoke flow

## 12. Recommended Baseline For New Repos

If a new repo starts today, the baseline should be:

- cross-platform setup from source checkout
- macOS default credentials in Keychain
- Windows default credentials in Windows Credential Manager
- `env_or_file` available as explicit fallback
- stable `auth set-access -> whoami -> resolve -> clean` workflow

That is the pattern to copy forward.
