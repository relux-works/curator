# Brief LOGDOC: public reference for agent logins (TASK-261016-2rpqii)

Read the attached `logins-facts.md` first. It is the source pack (measured facts, owner decisions, Curator's design, status, recipes). It is private material; this repository is PUBLIC.

## Deliver

1. **`docs/agent-logins.md`** (English, follow `docs/prose-style.md`) for an operator who wants to know where each agent tool keeps its login and how Curator handles it. Cover:
   - **Per tool:** Claude Code, Codex CLI and Muse, each on macOS and Linux. Where the login is stored, what changes it (desktop vs ssh session, Keychain availability, `cli_auth_credentials_store`, `TBH_CREDENTIAL_BACKEND`), what fails and with which message. Every measured fact carries the tool version it was observed on.
   - **Isolation:** `isolation.<profile>.<env-id>` = `shared | isolated`, what each does per tool, and the defaults. Reuse and link the existing table in `docs/environment-config.md` instead of duplicating it.
   - **Credential store:**
     - the machine knob `credential_store.<profile>.<env-id>` = `native | file | keychain`, default `native`, set with `curator env config set credential_store.<profile>.<env-id> <value>`;
     - it is machine-owned (packages, overlays, locks and repositories never select it), and why;
     - the per-tool mapping table (`-c cli_auth_credentials_store=...`, `TBH_CREDENTIAL_BACKEND=file`, Claude refusals);
     - changing it never migrates or deletes a login.
   - **Login routes:**
     - today: the `env resolve --format shell` + `claude auth login` route;
     - rc.7: `curator env login` and `curator profile prepare` with the readiness row;
     - planned: the Claude `setup-token` source (`CLAUDE_CODE_OAUTH_TOKEN`), with its limit (inference only; Remote Control refuses it).
   - **Recipes:** macOS desktop, macOS over ssh, a Linux server, a spawned agent.
   - **Status column:** every capability is marked "works today", "rc.7" or "planned", and the marking must match the code on `main`. Check the CLI on `main` before you write "works today"; do not document an unimplemented command as available.
   - **Links:** the normative text (curator-spec `protocol/login-readiness.md`, `protocol/environments.md`, CIP-0010, CIP-0015, CIP-0017, CIP-0018), linked by repository-relative URL on github.com/relux-works/curator-spec.
2. **Diagrams** in `diagrams/plantuml/`, a small focused file each:
   - where each tool's login lives (store per tool × session kind);
   - the resolution order source → store → isolation admission → link strategy;
   - the shared vs isolated home layout (link to the native file vs the home's own login);
   - the rc.7 `profile prepare` flow.

   Render them (PlantUML, SVG) next to the sources or under `diagrams/rendered/`, whichever is simpler, and say in the README tools section how to render. Embed the rendered diagrams in `docs/agent-logins.md`.
3. **Visibility:**
   - a short, prominent pointer in `README.md` next to the existing "Important" note on managed-home logins, including the one-line `credential_store` command and its default;
   - a link from `docs/environment-config.md` "Logging in to managed homes".
4. **Hygiene:** no host names, personal names, account details, seat or lane names, incident references or local paths. Write generic roles ("the operator", "a desktop session", "an ssh session").

## Checks

- The Markdown link check the repository uses (lychee, or what CI runs).
- `git diff --check`.
- PlantUML renders without errors.
- List the exact commands and results in the handoff.

Leave the change uncommitted for the Change Request flow.
