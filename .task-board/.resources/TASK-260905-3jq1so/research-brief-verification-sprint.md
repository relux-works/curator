# Research brief: verification sprint on the installed agent binaries (step 5)

## Purpose

Before the environments revision 1.1 vectors and adapter registry freeze, verify on this
machine (and the Windows host for the one Windows item) the facts the specification currently
carries at docs-confidence. Every result becomes board evidence: what was run, exact output,
version, and a verdict of **verified / falsified / not reproducible / requires operator**.
Nothing is asserted without a command and its output.

## Hard constraints

- Read-only toward the operator's real state: never modify, delete, or log out of the
  operator's Claude/Codex/pi/opencode homes, Keychain items, or auth files; never print,
  copy, or persist a token, cookie, credential JSON, or Keychain secret value — inspect
  metadata only (`security find-generic-password` WITHOUT `-w`; sizes, mtimes, key names).
- Fresh-home probes run under scratch homes below the worktree's `.temp/` (e.g.
  `CLAUDE_CONFIG_DIR`, `CODEX_HOME`, pi's agent dir, `XDG_CONFIG_HOME`) and never touch
  the real ones. Where a probe would require an interactive login, OAuth, or a permission
  prompt, stop at that point, record it as **requires operator**, and do not attempt to
  complete it.
- Bounded: each item gets at most a handful of targeted commands; no unbounded loops, no
  installs. Record the exact binary versions once: `claude --version`, `codex --version`,
  `pi --version`, `opencode --version` (if installed), `xcodebuild -version`.
- Windows item only through `ssh win` (key-based; a host with go1.25.5 exists). If unreachable,
  record it.
- Never write LOGBOOK.md or anything into the control root. Do not edit specification files.

## Items (one section each in the report)

1. **Keychain `oauth.claude.profile.*` keying.** `security dump-keychain 2>/dev/null | grep -A4
   -i 'claude'` limited to attribute lines (never `-d`); list the service/account names that
   exist (`Claude Code-credentials` keyed by the macOS account; any
   `oauth.claude.profile.<64-hex>` accounts). Does the per-profile scheme exist on 2.1.26x,
   and what selects it (search the `claude` bundle strings: `strings $(which claude) | grep -i
   'oauth.claude.profile'` and nearby identifiers)? Verdict on whether `isolated` homes could
   ever get separate credentials.
2. **`.claude.json` seed shape.** In a scratch `CLAUDE_CONFIG_DIR`, run `claude --version` and
   a non-interactive `claude -p 'say ok'` (expect "Not logged in" or similar — do not log in);
   list what files the tool creates on first run and the top-level keys of `.claude.json`
   (redact values). Which keys carry login state vs. preferences vs. project trust?
3. **codex global `AGENTS.md` cap.** Does `project_doc_max_bytes` (default 32768) apply to the
   global `$CODEX_HOME/AGENTS.md`? Evidence: `codex --help`, `codex exec --help`, the
   documentation strings in the binary (`strings` grep for `project_doc_max_bytes`,
   `AGENTS.md`), and a scratch-home experiment with a 40 KB global `AGENTS.md` where a
   deterministic marker sits after byte 32768 and `codex exec` is asked to repeat it — only if
   `codex exec` works without login in the scratch home; otherwise record requires operator.
4. **codex and pi `auth.json` write mode.** Determine whether each tool rewrites its auth
   file in place or by temp+rename: inspect metadata (inode via `stat -f %i`, mtime) of the
   real files before/after a read-only command that does not refresh tokens — only if such a
   command exists; otherwise inspect the binary/docs for `rename`/atomic-write patterns and
   codex `cli_auth_credentials_store` (`file|keyring|auto`) in `codex --help`/config docs.
   Verdict per tool: in-place / rename-over / unknown.
5. **Fresh-home first run per tool with seeds.** For claude, codex, pi (and opencode if
   installed): in a scratch home, what is the first-run behavior with an empty home, and with
   the seed files the environments text proposes (`.claude.json` minimal, codex `config.toml`
   with trust entries, pi `settings.json`)? Which walls appear (login, trust prompt,
   onboarding), and which seeds remove which wall? Stop at any interactive prompt.
6. **Xcode embedded agents honoring root context.** With `xcodebuild -version`, locate the
   CodingAssistant home (`~/Library/Developer/Xcode/...` or as the environments text names it);
   read-only listing of what files exist; do the embedded agents read `CLAUDE.md`/`AGENTS.md`
   from that home (evidence: bundle strings, docs, existing files)? Behavioral test only if it
   needs no GUI action; else requires operator.
7. **opencode `XDG_CONFIG_HOME` on Windows.** Over `ssh win`: is opencode installed; if so,
   does it honor `XDG_CONFIG_HOME` / where does it read config on Windows (`%APPDATA%`?);
   evidence from `opencode --help`, docs strings, or a scratch run.
8. **claude referenced-form approval.** Does claude 2.1.26x treat `@path` references in
   `CLAUDE.md` inside a scratch `CLAUDE_CONFIG_DIR` as approved without a prompt (the
   environments §5.3 referenced form)? Evidence from a non-interactive run in the scratch home
   if login is not required for the read; else bundle strings/docs and requires operator.
9. **codex `-p curator-mcp` layer composition.** In a scratch `CODEX_HOME`, write
   `curator-mcp.config.toml` with only `mcp_servers`, run `codex -p curator-mcp exec --help`
   or a config-dump command if one exists (`codex config`? check `--help`), and determine:
   is the layer applied; what happens with two `-p` flags (last wins? error?); does the layer
   file need other members. Record exact output.

## Deliverable

Outcome resource `TASK-TASK-260905-3jq1so_verification-sprint.md` on the task: per item — commands,
redacted output, versions, verdict, and the exact sentence(s) the environments text or the
adapter registry must change as a result (quote the docs-confidence claim, give the verified
replacement). A final table item → verdict. Then
`task-board handoff TASK-260905-3jq1so --role researcher`.
