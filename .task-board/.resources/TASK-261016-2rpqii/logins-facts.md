# Agent logins: measured facts, owner decisions and Curator's design (source pack, 2026-10-16)

This is the single source pack for documentation and the book about how agent tools (Claude Code, Codex CLI, Muse) keep their logins and how Curator uses that. It is private material. Anything published to a public repository must drop host names, personal names, account details, seat or lane names and any incident reference.

## 1. Measured facts (probes of 2026-10-14)

All probes ran on scratch OS users and scratch homes. The tool versions are named per row.

### Claude Code 2.1.299, macOS 15

- **Desktop login.** A login from a session with login-Keychain access (the desktop session and everything started from it, including tmux servers and daemons it launched) lands in a per-home Keychain item `Claude Code-credentials-<first 8 hex of sha256(CLAUDE_CONFIG_DIR)>`. No file is written; the native login is untouched.
- **ssh login.** A login from a session without Keychain access (ssh and everything started from it; `security show-keychain-info` exits 36 there) lands in `<home>/.credentials.json`, mode 0600. No Keychain item is written. Hiding the file gives "Not logged in".
- **The two stores do not see each other.** A home logged in from the desktop reports `loggedIn=false` to an ssh-started session, and the other way round.
- **No Keychain at all.** An OS user with no login Keychain (a launchd Background session) cannot log in: "Couldn't save your login". There is no file fallback.
- **Env token.** `CLAUDE_CODE_OAUTH_TOKEN` (from `claude setup-token`) authenticates with `authMethod=oauth_token`.
  - It wins over a login stored in the home and writes nothing to the Keychain or the home.
  - It is inference-only: Remote Control refuses it.
  - `claude auth status --json` reports `loggedIn` and `authMethod` without a model call.
- **Linux.** Claude stores the login in `<home>/.credentials.json`, so `shared` can link it to the native file.

### Muse 1.4.8, macOS

- **Default store.** By default the login is one global Keychain item `ai.meta.dev.credentials`; `auth.json` is only a fallback. Sessions record `credential_backend` as keychain, keychain_fallback_file or file.
- **No Keychain.** With no Keychain, OAuth succeeds but the save fails ("keychain write failed (os status -60006)"); there is no file fallback and login exits 1.
- **`TBH_CREDENTIAL_BACKEND=file`:**
  - the login is saved to `$XDG_CONFIG_HOME/muse/auth.json` (0600);
  - later runs use backend `file` with zero Keychain access;
  - hiding the file gives missing credentials;
  - a home whose `auth.json` links to another home's file works (this is the basis of `shared`).
- **API keys.**
  - A minted API key passed by `exec --api-key-stdin` bills the subscription.
  - `META_API_KEY` leaves the store untouched.
- Token-refresh write behaviour is unmeasured.

### Codex CLI 0.153.2

- **Store selection.** The store is selected by `cli_auth_credentials_store` in `config.toml`:
  - absent means `file` (`$CODEX_HOME/auth.json`);
  - `keyring` and `auto` use the OS keyring.
- **Override.** `-c cli_auth_credentials_store="file"` on the command line overrides the file for one invocation. This is what Curator's launcher uses; nobody rewrites `config.toml`.

## 2. Owner decisions

D1–D11 date from 2026-10-15 and D14 from 2026-10-16; the full text is in `rc7-decisions.md`, attached.

- **D7.** Login readiness is a Curator status row: `login: present | absent | unknown`, metadata only.
- **D8.** `curator profile prepare P` prepares a profile for spawns, and `curator env login <env> --profile P` runs the tool's own login inside a managed home.
- **D9.** Curator decides per profile × environment whether a tool keeps its login in the Keychain or in a file, and the launcher applies it.
- **D10.** Muse `isolated` and `shared` are admitted through `TBH_CREDENTIAL_BACKEND=file`.
- **D11.** Swarm agents do not need Remote Control, and long-lived credentials (Claude `setup-token`, a Muse minted key) are accepted for v0.
- **D14.** Credential-store selection is machine-owned:
  - the only selector is the machine knob `credential_store.<profile>.<env-id>` = `native | file | keychain` (default `native`), set with `curator env config set credential_store.<profile>.<env-id> <value>`;
  - packages, overlays, locks and repositories never select a store;
  - the reason: where a login lives depends on the machine (Keychain or not, desktop or ssh, macOS or Linux, shared or isolated), not on the profile.
- **New, 2026-10-16: Claude's env-token path must be supported.** This is the `setup-token` source of CIP-0010 delivered as `CLAUDE_CODE_OAUTH_TOKEN`. It removes the per-home login for Claude on macOS and the desktop/ssh split, and it works for users without a Keychain. It is inference-only, which D11 accepts.

## 3. Curator's design (normative in curator-spec main 5509499)

- **Isolation** is `isolation.<profile>.<env-id>` = `shared | isolated` (curator `docs/environment-config.md`, "Logging in to managed homes"):
  - `shared` links the home's credential file to the native file;
  - `isolated` means one login per home.
- **Claude on macOS** is always `isolated`, because its Keychain item is keyed to `CLAUDE_CONFIG_DIR`; `shared` is refused with `environment_shared_unsupported`.
- **Credential store, CIP-0015 login-readiness §2.** The resolution order is source → store → isolation admission → link strategy.

  | Adapter | native | file | keychain |
  |---|---|---|---|
  | Codex | native `config.toml`; absent means file; no override | `-c cli_auth_credentials_store="file"` at login and launch | `-c ... "keyring"` only on a qualified macOS backend; otherwise refuse |
  | Muse | native: global Keychain with file fallback | `TBH_CREDENTIAL_BACKEND=file` at login and launch | refuse until a force-Keychain interface is qualified |
  | Claude | session-qualified native choice | refuse: no force-file interface | refuse: no force-Keychain interface |

  Changing the selector never migrates or deletes a login.
- **`curator env login`** (CIP-0015 §3) runs only for isolated, home-owned logins, from a real terminal. It never captures credential output. Shared mode refuses and gives native-login guidance.
- **`curator profile prepare`** (CIP-0015 §4) installs if absent and enumerates every environment. It logs in only where a home-owned login is absent, and reports a readiness table.
- **CIP-0018 and D13/D14.** Profile-level Codex settings reach Codex as typed `-c` launch overrides. The credential store is not a package key: Curator fills `cli_auth_credentials_store` from the machine selector.
- **CIP-0017.** Muse gets an explicit `root_context` fragment member; the credential store for Muse goes through `TBH_CREDENTIAL_BACKEND=file`.
- **CIP-0010 (Draft).**
  - Claude `setup-token` is enrolled once into a protected per-user store under an authentication node, referenced by an opaque `source_ref`, and injected as `CLAUDE_CODE_OAUTH_TOKEN` by the final executor only. The plan and the fragment never carry the token.
  - Homes with inherited authentication need no per-home login.
  - This is the "OAuth" path the owner wants supported.

## 4. Status today (curator rc.6 installed, main 205272e7)

| Capability | Status |
|---|---|
| `isolation.<profile>.<env-id>` shared/isolated | works today |
| Claude macOS per-home login via `eval "$(curator env resolve claude_code --profile P --format shell)"; claude auth login` | works today (documented route) |
| `credential_store.<profile>.<env-id>` knob | decided (CIP-0015, D14); implementation in rc.7 |
| `curator env login`, `curator profile prepare`, login readiness row | decided (CIP-0015); implementation in rc.7 |
| Codex `-c` overrides from profile settings | decided (CIP-0018); implementation in rc.7 |
| Muse `root_context`, `TBH_CREDENTIAL_BACKEND=file` | decided (CIP-0017, D10); implementation in rc.7 |
| Claude `setup-token` env source (OAuth) | proposed (CIP-0010 Draft); the owner wants it supported; needs normative text and implementation |
| Managed home snapshots (`env materialize`) | decided (CIP-0014); implementation in rc.7 |

## 5. Practical recipes (what a human does)

- **Claude, macOS desktop, today.**

  ```bash
  eval "$(curator env resolve claude_code --profile P --format shell)"
  claude auth login
  ```

  Run it from the same kind of session as the launcher.
- **Codex.** Nothing to do when the native login is a file (the default): `shared` links it. For an isolated per-profile login, after rc.7:

  ```bash
  curator env config set credential_store.P.codex_cli file
  curator env login codex_cli --profile P
  ```

- **Muse.** Shared through `TBH_CREDENTIAL_BACKEND=file` and the `auth.json` link (D10).
- **Claude without any login in the home (planned, CIP-0010).** Enroll a `setup-token` once; spawned agents get `CLAUDE_CODE_OAUTH_TOKEN` from the executor.
