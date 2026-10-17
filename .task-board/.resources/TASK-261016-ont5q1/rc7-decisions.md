# rc.7 spec package: owner decisions (2026-10-14/15)

Binding decisions from the owner, recorded by the curator lane orchestrator. Every CIP in this package cites them as "owner decision 2026-10-15".

## Ownership

- D1. Curator owns agent homes: content, verification, repair, login readiness and snapshots. The launcher (agents-management) only builds argv and env from the Curator fragment and the plan. task-board and the session host only point at a home and keep their own concerns: the private session folder with tool policy, trust, resume, goals and input. Nobody else keeps a list of what a home contains or checks profile files by its own rules.
- D2. `curator-run` is the entry point for launching agents with Curator profiles, including task-board hosted sessions. Hosted launches must work for Claude Code, Codex and Muse from the start (Codex is not deferred).
- D3. The agents-infra exit is the current priority: the repository has to be opened soon, and Curator is its entry point.

## Homes and snapshots

- D4. A live session never points into a Curator managed home: homes are rewritten in place by `env resolve --repair`, `profile use` and global skill sync, with no running-session check. A session gets its own snapshot.
- D5. `curator env materialize --into <dir>` produces that snapshot (regular files plus a manifest with digests). The task-board interim copy (copy between two equal-digest resolves, byte compare, `auth.json` linked to the native file) is temporary and is removed once materialize ships.
- D6. In the swarm the same snapshot travels as a plan-covered archive: the caller materializes, the plan covers the archive digest like task resources, the dispatcher relays it, and `curator-run exec` unpacks it read-only in the agent account, verifies the digest and points `CLAUDE_CONFIG_DIR` or the XDG variables at it. The dispatcher never learns where Curator homes live. We own the dispatcher and launcher specifications too.

## Login

- D7. Login readiness is a Curator status row: `login: present | absent | unknown`, metadata only, never secret bytes.
- D8. One command prepares a profile for spawns: `curator profile prepare P` (install, resolve every participating environment, log in where a home needs its own login, report readiness). `curator env login <env> --profile P` runs the tool's own login inside the managed home.
- D9. Curator decides per profile × environment whether a tool keeps its login in the Keychain or in a file; the launcher applies it (Codex: `-c cli_auth_credentials_store="file"` in the launch; Muse: `TBH_CREDENTIAL_BACKEND=file` in the environment).
- D10. Muse `isolated` and `shared` are admitted through `TBH_CREDENTIAL_BACKEND=file` now, optimistically, without waiting for the token-refresh probe.
- D11. Swarm agents do not need Remote Control. Long-lived credentials (Claude `setup-token`, Muse minted key) are accepted for v0.

## Muse root context and Codex settings

- D12. Muse: Curator states the root-context fact with an explicit `root_context` member in the launch fragment; the launcher reads that member and stops inspecting Curator's file layout. Do it properly now; no interim launcher hack.
- D13. Codex settings live at profile level with inheritance and reach Codex as `-c key=value` at launch; nobody rewrites the `config.toml` that Codex owns.

- D14 (2026-10-16). Credential-store selection is machine-owned. CIP-0015 login-readiness section 2 stands unchanged: the only selector is the user-owned machine knob `credential_store.<profile>.<env-id>` (default `native`), and packages, overlays, locks and repositories never select a store. CIP-0018 keeps the profile-level Codex settings machinery of D13 (manifest member, inheritance, typed `-c` launch overrides, no `config.toml` rewrite) as the single transport, but `cli_auth_credentials_store` leaves the package vocabulary: Curator fills that override from the machine selector. Reason: where a login lives depends on the machine (Keychain or not, desktop or ssh, macOS or Linux, shared or isolated), not on profile content, and a package must not decide that tokens go to disk.

## Measured facts (2026-10-14, for citation)

- Claude Code 2.1.299 on macOS 15: a login from a session with Keychain access (the desktop session and everything started from it) lands in the per-home Keychain item `Claude Code-credentials-<first 8 hex of sha256(config dir)>`. A login from a session without Keychain access (ssh and everything started from it; `security show-keychain-info` exits 36) lands in `<home>/.credentials.json` (0600). The two stores do not see each other. An OS user with no login Keychain cannot log in ("Couldn't save your login"). `CLAUDE_CODE_OAUTH_TOKEN` authenticates (`authMethod=oauth_token`), wins over a login stored in the home, writes nothing, and is inference-only: Remote Control refuses it. `claude auth status --json` reports `loggedIn` and `authMethod` without a model call.
- Muse 1.4.8 (macOS 27): by default the login is one global Keychain item `ai.meta.dev.credentials`, and `auth.json` is only the fallback (sessions record `credential_backend` keychain / keychain_fallback_file / file). With no Keychain the login is not saved ("keychain write failed (os status -60006)"). With `TBH_CREDENTIAL_BACKEND=file` the login is saved to `$XDG_CONFIG_HOME/muse/auth.json` (0600), later runs use backend `file` with zero Keychain access, and a home whose `auth.json` links to another home's file works. The minted API key passed by `exec --api-key-stdin` bills the subscription. `META_API_KEY` leaves the store untouched. Token-refresh write behaviour is unmeasured.
- Codex: sharing is defined by the native `config.toml` storage (`cli_auth_credentials_store`; absent means `file`); curator-spec environments section 7.4.

## M1 leaf correction (2026-10-16)
Leaf muse-v4-producer acceptance 2 said "No applicable root module means no fact". The normative rule (environments 5.4 and 10.2) keys the fact on a verified root-context surface. A declared context with zero applicable modules still materializes a header-only surface and emits the fact; only a root that declares no context has no surface and no fact. The leaf now says so. The spec is unchanged.
M1 consumes the common effective credential policy from the L15-1 remainder (TASK-261016-hzbo83) and starts after that lands. It adds no second resolver.
