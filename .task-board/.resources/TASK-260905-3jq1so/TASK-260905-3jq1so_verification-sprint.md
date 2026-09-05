# TASK-260905-3jq1so — verification sprint on the installed agent binaries

Date: 2026-09-05. Host: this Mac (Darwin 25.5.0, arm64). All fresh-home probes ran under
`.temp/TASK-260905-3jq1so/` scratch homes or `/tmp` scratch project dirs. No real home, auth file, or
Keychain secret value was modified, printed, or persisted. Keychain was inspected with
`security dump-keychain` / `find-generic-password` **without** `-d`/`-w` (attribute lines only).
Where a real file was inspected, only `stat` metadata and JSON **key names** were read.

## Binary versions (recorded once)

| Tool | Command | Output | Exit |
| --- | --- | --- | ---: |
| claude | `claude --version` | `2.1.261 (Claude Code)` — bun-compiled Mach-O at `~/.local/share/claude/versions/2.1.261` | 0 |
| codex | `codex --version` | `codex-cli 0.153.2` — `~/.codex/packages/standalone/releases/0.153.2-aarch64-apple-darwin/bin/codex` | 0 |
| pi | `pi --version` | `0.84.2` — node script `/opt/homebrew/lib/node_modules/@earendil-works/pi-coding-agent/dist/cli.js` | 0 |
| opencode | `which opencode` | not installed on this Mac | 1 |
| xcodebuild | `xcodebuild -version` | `Xcode 26.5`, `Build version 17F42` (`/Applications/Xcode_26_5.app`) | 0 |

Verdict vocabulary: **verified** / **falsified** / **not reproducible** / **requires operator**.

---

## 1. Keychain `oauth.claude.profile.*` keying

Commands (attribute lines only, hex account redacted to 8…8):

```
security dump-keychain 2>/dev/null | grep -A12 'oauth.claude.profile' | grep -E '"(svce|labl|acct)"'
    "acct"<blob>="oauth.claude.profile.3cd05c28…0c177b8a"
    "svce"<blob>="com.steipete.codexbar.cache"
security find-generic-password -s 'Claude Code-credentials'      # no -w
    "acct"<blob>="iv"
    "svce"<blob>="Claude Code-credentials"
# other Claude-named items: svce="Claude Code-doctor-probe" acct="iv"; svce="Claude Safe Storage" acct="Claude Key" (Claude desktop app)
grep -a -c -F 'oauth.claude.profile' ~/.local/share/claude/versions/2.1.261   -> 0
```

Service-name builder extracted from the bundle (offset ~157959420, verbatim minified JS):

```
var N5="-credentials";
function A_(){let n=process.env.CLAUDE_SECURESTORAGE_CONFIG_DIR;if(n!==void 0)return(n||l(o(),".claude")).normalize("NFC");return be()}
function Sx(n=""){let e=process.env.CLAUDE_SECURESTORAGE_CONFIG_DIR,t=e!==void 0?!e:!process.env.CLAUDE_CONFIG_DIR,
  r=e!==void 0?e.normalize("NFC"):be(),
  c=t?"":`-${a("sha256").update(r).digest("hex").substring(0,8)}`;
  return`Claude Code${Vt().OAUTH_FILE_SUFFIX}${n}${c}`}
function tv(){... n=process.env.USER||u().username ... return n}   // keychain account
```

Findings:
- The `oauth.claude.profile.<64-hex>` account is **not a Claude Code item**. Its service is
  `com.steipete.codexbar.cache` (CodexBar, a third-party menubar app). Claude Code 2.1.261 contains no
  such string.
- Claude Code keys its Keychain item as service `Claude Code{OAUTH_FILE_SUFFIX}-credentials`, account
  `$USER`. **When `CLAUDE_CONFIG_DIR` (or `CLAUDE_SECURESTORAGE_CONFIG_DIR`) is set**, the service gets
  a suffix `-<first 8 hex of sha256(config dir path, NFC-normalized)>`. With no such env var the
  service is the plain `Claude Code-credentials` (the only one present in this login keychain).
- Consequence: every distinct `CLAUDE_CONFIG_DIR` gets its own Keychain item, so an `isolated` home
  has separate credentials by construction; it must log in on its own (the plain item is not
  consulted). Xcode's `IDEIntelligenceAgents` replicates the same scheme
  ("CryptographicUtils: Short hash (first 8 chars)", see item 6).

**Verdict: falsified** (the per-profile `oauth.claude.profile.*` scheme is not Claude Code's) and
**verified** (per-config-dir Keychain items exist and are selected by `CLAUDE_CONFIG_DIR`).

Implied text change: replace any claim that credentials are keyed by `oauth.claude.profile.<hash>` with:
"Claude Code stores OAuth credentials in the login Keychain as service
`Claude Code-credentials` (account = `$USER`). When `CLAUDE_CONFIG_DIR` is set, the service name is
suffixed with `-` plus the first 8 hex chars of sha256 of the config-dir path, so each isolated home
owns a separate Keychain item and needs its own login."
Registry: adapter `claude` → `credential_scope: per CLAUDE_CONFIG_DIR (keychain suffix sha256[0:8])`.

---

## 2. `.claude.json` seed shape

```
CLAUDE_CONFIG_DIR=<scratch> claude --version        -> 2.1.261 (Claude Code), exit 0, creates nothing
cd /tmp && CLAUDE_CONFIG_DIR=<scratch> claude -p 'say ok'  (CLAUDECODE/ANTHROPIC_* unset)
    Not logged in · Please run /login                        exit 1
files created by the -p run:
    .claude.json  backups/.claude.json.backup.<ts>  projects/-private-tmp/<uuid>.jsonl  projects/-private-tmp/memory/  sessions/
top-level keys written on first run (types):
    firstStartTime:str firstStartVersion:str machineID:str userID:str migrationVersion:int(14)
    opusProMigrationComplete:bool sonnet1m45MigrationComplete:bool hasResetAutoModeOptInForDefaultOffer:bool seenNotifications:dict
```

Key-name inventory of the real `~/.claude.json` (names only, values never read) sorted by role:
- **login state**: `oauthAccount` (sub-keys `accountUuid, emailAddress, organizationUuid, organizationRole,
  workspaceRole, billingType, seatTier, subscriptionCreatedAt, userRateLimitTier, …`), `hasAvailableSubscription`,
  `userID`, `bridgeOauth*`. Tokens themselves are not in `.claude.json` (Keychain / `.credentials.json`, item 1).
- **onboarding / preferences**: `hasCompletedOnboarding`, `lastOnboardingVersion`, `numStartups`,
  `installMethod`, `autoUpdates`, `theme` (default `dark`, in-code defaults object), `firstStartTime`,
  `migrationVersion`, many `hasSeen*/hasShown*` flags, caches (`cached*`, `*Cache`).
- **project trust** (`projects.<abs path>`): `hasTrustDialogAccepted`, `hasClaudeMdExternalIncludesApproved`,
  `hasClaudeMdExternalIncludesWarningShown`, `allowedTools`, `mcpServers`, `enabledMcpjsonServers`,
  `disabledMcpjsonServers`, `mcpContextUris` (+ per-session metrics). In-code default project entry:
  `{allowedTools:[],mcpContextUris:[],mcpServers:{},enabledMcpjsonServers:[],disabledMcpjsonServers:[],hasTrustDialogAccepted:false,hasClaudeMdExternalIncludesApproved:false,hasClaudeMdExternalIncludesWarningShown:false}`.

A seeded `.claude.json` containing only `hasCompletedOnboarding` and `projects.<path>` survived the
first run unchanged (claude merged its own first-run keys around it; verified by re-reading keys).

**Verdict: verified.** Implied text: the minimal seed is
`{"hasCompletedOnboarding":true,"projects":{"<abs project path>":{"hasTrustDialogAccepted":true}}}`;
add `"hasClaudeMdExternalIncludesApproved":true` in the project entry only if external `@` includes
are wanted (item 8). Note the project key is the **realpath-independent literal path used as cwd**
(`/tmp/...` seed matched a `/tmp` cwd; `projects/` session dir was named `-private-tmp`).

---

## 3. codex global `AGENTS.md` cap

```
codex --help / codex exec --help: no project_doc flag; -p documented as "Layer $CODEX_HOME/<name>.config.toml on top of the base user config"
strings codex | grep project_doc_max_bytes   -> 'project_doc_max_bytes = 32768' (3× in embedded config docs), 'project_doc_fallback_filenames = []'
strings codex | grep AGENTS.md -> 'Failed to read global AGENTS.md instructions from `', 'project doc exceeds remaining budget; truncating' (core/src/agents_md.rs:157, remaining_bytes)
```

Experiment (no login needed: `codex debug prompt-input` renders the model-visible prompt as JSON):

```
$CODEX_HOME/AGENTS.md = 41011 bytes, marker "MARKER-AFTER-CAP: ZEBRA-7731" at byte offset 33046 (> 32768)
cd /tmp/cx-proj-3jq1so (git repo)
CODEX_HOME=<scratch> codex debug prompt-input 'hello'          exit 0
   contains ZEBRA-7731: True   filler lines: 661 (= whole file)
control: same 41011-byte file as project AGENTS.md with marker KOALA-4412:
   global ZEBRA present: True   project KOALA present: False   filler lines: 1190 (661 global + 529 ≈ 32768 B project)
narrowed: -c project_doc_max_bytes=1000
   global ZEBRA present: True   project KOALA present: False   filler lines: 677 (661 global + 16 ≈ 1000 B project)
```

**Verdict: falsified.** `project_doc_max_bytes` applies only to the project-doc chain
(`AGENTS.override.md`/`AGENTS.md` from cwd up); the global `$CODEX_HOME/AGENTS.md` is injected in full
(40 KB passed untruncated even with the cap narrowed to 1000). Implied text change: replace
"the 32 KiB `project_doc_max_bytes` cap also applies to the global `AGENTS.md`" with "the cap applies to
project docs only; the global `$CODEX_HOME/AGENTS.md` is not truncated (codex 0.153.2, verified with a
41 KB file via `codex debug prompt-input`)". Registry: `codex.global_context_cap: none`.

---

## 4. codex and pi `auth.json` write mode

codex:
```
stat ~/.codex/auth.json  before: inode=734845569 size=4178 mtime=1788560952 mode=-rw-------
codex login status               -> Logged in using ChatGPT   exit 0
stat ~/.codex/auth.json  after:  inode=734845569 size=4178 mtime=1788560952   (untouched — status is read-only, so write mode is not observable locally)
strings: cli_auth_credentials_store = "file" (default in embedded docs); enum File|Keyring|Auto (+Ephemeral); 'login/src/auth/storage.rs'
strings show temp+rename only for other stores ("failed to atomically replace secrets file", "Codex Apps cache"), none for auth.json
```
Source of the matching version line (openai/codex `codex-rs/login/src/auth/storage.rs`, `FileAuthStorage::save`, fetched 2026-09-05):
```rust
let mut options = OpenOptions::new();
options.truncate(true).write(true).create(true);
#[cfg(unix)] { options.mode(0o600); }
let mut file = options.open(auth_file)?; file.write_all(json_data.as_bytes())?; file.flush()?;
```
→ **codex: in-place truncate+write** (same inode), mode 0600. `cli_auth_credentials_store` = `file` (default) | `keyring` | `auto`.

pi (`dist/core/auth-storage.js`, lines 13, 28, 65, 140):
```js
const AUTH_FILE_WRITE_OPTIONS = { encoding: "utf-8", mode: 0o600 };
writeFileSync(this.authPath, next, AUTH_FILE_WRITE_OPTIONS); chmodSync(this.authPath, 0o600);   // under proper-lockfile lock
```
`stat ~/.pi/agent/auth.json`: inode=574141555 size=2 mode=-rw------- (content is `{}` — no provider creds stored; key list `[]`).
→ **pi: in-place writeFileSync** under a `proper-lockfile` lock (no temp+rename). Settings (`core/settings-manager.js:93`) likewise.

**Verdict: verified (in-place for both; codex from upstream source of the same file path named in the binary, pi from installed source).**
Implied text: "codex and pi rewrite `auth.json` in place (truncate + write, mode 0600); a snapshot taken mid-write can observe a truncated file — copy under the tool's lock or when idle. Neither uses temp+rename." Registry: `codex.auth_write: in-place`, `pi.auth_write: in-place (lockfile)`.

---

## 5. Fresh-home first run per tool with seeds

claude (scratch `CLAUDE_CONFIG_DIR`, `-p`):
- empty home → `Not logged in · Please run /login`, exit 1. No trust or onboarding wall is reachable in `-p`; the
  login wall comes first. Seeded `.claude.json` (onboarding + trust) → identical message. **Login wall: requires operator** (OAuth).
- With a deliberately invalid `ANTHROPIC_API_KEY` (no real secret) the run passes settings/skills/memory loading and fails
  only at the API (`authentication_error`, 11 retries, killed by 45 s timeout, exit 124) — proves no interactive wall in `-p` mode.

codex (scratch `CODEX_HOME`, `codex exec`, stdin closed):
- empty home, git cwd or `--skip-git-repo-check` → runs, fails with `401 Unauthorized` (exit 101/1). **Login wall: requires operator.**
- non-git cwd without the flag → `Not inside a trusted directory and --skip-git-repo-check was not specified.` exit 1.
- trust seed `[projects."<path>"] trust_level = "trusted"` in `config.toml` (tested with `/tmp/...`, its realpath
  `/private/tmp/...`, a `mktemp -d` dir, and via `-c projects."<path>".trust_level="trusted"`) **did not lift** that wall
  (config.toml was demonstrably parsed: an `mcp_servers` entry in the same file showed up in `codex mcp list`).
  Only a git repo or `--skip-git-repo-check` lifts it. **Falsified** for the "trust seed removes the exec wall" claim.
- first run creates: `installation_id`, `.sandbox_migration`, `skills/.system/*` (bundled skills), `sessions/`,
  `shell_snapshots/`, `tmp/`, sqlite state (`state_5`, `logs_2`, `goals_1`, `memories_1`, `queue_1`, `thread_history_1`). No `auth.json` is created.

pi (`PI_CODING_AGENT_DIR=<scratch>`, `pi --offline -p 'say ok'`, `getAgentDir()` in `dist/config.js:412`):
- empty dir → `No API key found for the selected model. Use /login …` exit 1; creates `auth.json` (`{}`), `models-store.json`, `sessions/<cwd>/`.
- `settings.json` seed `{"defaultProvider":"anthropic","defaultModel":"claude-sonnet-4-5"}` → same wall (only credentials remove it).
  pi has no trust dialog; `--no-approve`/`TrustOverride` only governs project-local files. **Login wall: requires operator.**

opencode: not installed → **not reproducible** here.

Implied text: "In non-interactive mode each tool's first wall is authentication; claude reaches it before any trust/onboarding
prompt, codex's only other wall is the git/`--skip-git-repo-check` check which `projects.*.trust_level` does not lift for `exec`,
pi has no trust wall." Registry: `codex.exec_flags: --skip-git-repo-check required outside git`.

---

## 6. Xcode embedded agents honoring root context

```
xcodebuild -version -> Xcode 26.5 / 17F42
ls ~/Library/Developer/Xcode/ -> DerivedData UserData *.plist XcodeCloud*   (no CodingAssistant/IntelligenceAgents dir)
find ~/Library -maxdepth 4 -iname 'com.apple.dt.IDEIntelligenceAgents*' -> nothing
defaults read com.apple.dt.Xcode | grep -E 'IDEChat|IDEIntelligence' -> IDEChatHasPreviouslyInstalledCodex, IDEIntelligenceHasInstalledAtLeastOnce, IDEChatIsBuiltInChatGPT* (values redacted)
strings /Applications/Xcode_26_5.app/Contents/PlugIns/IDEIntelligenceAgents.framework/Versions/A/IDEIntelligenceAgents:
  "Found claude at path: %s"  "Set CLAUDE_CONFIG_DIR to internal config directory: %s"  "Set CLAUDE_CONFIG_DIR to Claude Agent config directory: %s"
  "Set CLAUDE_CONFIG_DIR to UserDefaults override path: %s"  "IDEChatOverrideAgenticHomeDirectory"  "Created MCP config at: %{public}s"
  "--append-system-prompt --dangerously-skip-permissions --include-partial-messages --permission-prompt-tool"  "APPLE_CLAUDE_CODE_PROXY_PORT"
  "CryptographicUtils: Generating Claude directory hash for path" / "Short hash (first 8 chars)"   (matches item 1 Keychain scheme)
  "Add a CLAUDE.md file to this workspace, if one does not already exist"  ".claude/settings.local.json"
  "Override CODEX_HOME: %s"  "# ~/.codex/rules/default.rules"  "Generate an AGENTS.md scaffold in the current directory."  "AGENTS.md already exists here. Skipping /init"
  "com.apple.dt.IDEIntelligenceAgents/plugins/"
```

Findings: Xcode 26.5 does not bundle claude/codex; it launches the installed `claude` / `codex` binaries with
`CLAUDE_CONFIG_DIR` / `CODEX_HOME` pointed at an Xcode-internal directory (overridable via the
`IDEChatOverrideAgenticHomeDirectory` user default) and passes the system prompt via `--append-system-prompt`.
Workspace `CLAUDE.md` / `AGENTS.md` are referenced as workspace files (`/init`-style commands), and the
`.claude/settings.local.json` of the workspace is read. The internal home directory does not exist on this
machine (the agents were never launched here), so whether a `CLAUDE.md`/`AGENTS.md` placed in that home is
read cannot be listed or tested without opening Xcode.

**Verdict: requires operator** (launch an Xcode coding agent once, then list the internal config dir and test a
home-level `CLAUDE.md`); bundle evidence **verified** for env-var injection and the sha256[0:8] Keychain hash.
Implied text: name the home as "Xcode-internal `CLAUDE_CONFIG_DIR`/`CODEX_HOME` (path set by Xcode; override
`IDEChatOverrideAgenticHomeDirectory`)", not `~/Library/Developer/Xcode/CodingAssistant`.

---

## 7. opencode `XDG_CONFIG_HOME` on Windows

```
ssh -o BatchMode=yes -o ConnectTimeout=15 win 'hostname …'  -> ssh: connect to host <tailscale ip> port 22: Operation timed out
ssh -o BatchMode=yes -o ConnectTimeout=40 win 'hostname'    -> same
```
**Verdict: not reproducible** (host unreachable during the sprint; `~/.ssh/config` `Host win` entry exists, key-based).
No text change can be derived; keep the claim at docs-confidence and re-run
`ssh win 'where opencode & echo %APPDATA% & echo %XDG_CONFIG_HOME%'` when the host is up.

---

## 8. claude referenced-form (`@path`) approval

Bundle code (2.1.261, offsets ~166207999 / 166215394, verbatim minified):
```
function egr(){return M1()!=="local-agent"}
async function Sgs(...){... I=es(), D=o||I.hasClaudeMdExternalIncludesApproved||!1 ... aD(<user CLAUDE.md>,"User",C,D) ...}
async function aD(e,t,r,o,d=0,p,_){... let C=o&&(t!=="User"||egr()); ...
   if(d>0&&!C&&!xZ(I))return[];            // nested file outside project: skipped unless approved
   for(let V of F){if(!xZ(V)&&!C)continue; U.push(...await aD(V,t,r,o,d+1,e))}}   // @include outside project: skipped unless approved
function xZ(e){return Ap(e,he())}          // "is inside the project cwd"
function m8e(e){... r.type!=="User"&&r.parent&&!xZ(r.path) ...}   // files counted as "external" for the warning dialog
async function yXn(e,t){let r=es();if(r.hasClaudeMdExternalIncludesApproved||r.hasClaudeMdExternalIncludesWarningShown)return!1;return L_n(...)}
dialog strings: "Yes, allow external imports" / "No, disable external imports"
```
Scratch runs (fake `ANTHROPIC_API_KEY`, `--debug-file`): with `@/etc/hosts` in the scratch-home `CLAUDE.md` (approved and
unapproved) and in the project `CLAUDE.md` (unapproved), `claude -p` never blocked on a prompt; it proceeded to the API
(exit 124 only from the timeout on retries). The debug log carries no include line, so which content was injected is not
observable without a valid model call.

**Verdict: falsified** for "referenced form is approved without a prompt". Precisely: an `@path` whose target resolves
**outside the project cwd** is silently **skipped** unless `projects.<cwd>.hasClaudeMdExternalIncludesApproved` is true
(the interactive dialog sets it; `-p` never asks). `@path` targets **inside** the project need no approval. The type of the
including file (User vs Project) does not exempt it (`egr()` is true outside the `local-agent` mode).
Implied text change: environments §5.3 must say "referenced files outside the project are loaded only if the seed sets
`hasClaudeMdExternalIncludesApproved: true` for that project; otherwise they are dropped without error" and the seed in item 2
must include that key for the referenced form. Behavioral confirmation of injected content: **requires operator** (logged-in run).

---

## 9. codex `-p curator-mcp` layer composition

```
$CODEX_HOME/curator-mcp.config.toml:  [mcp_servers.curator-probe] command = "/usr/bin/true"
$ codex mcp list                                   -> No MCP servers configured yet.            exit 0
$ codex -p curator-mcp mcp list                    -> curator-probe  /usr/bin/true  enabled     exit 0
$ codex -p curator-mcp -p second mcp list          -> error: the argument '--profile <CONFIG_PROFILE_V2>' cannot be used multiple times   exit 2
$ codex -p nonexistent mcp list                    -> No MCP servers configured yet.  (silently ignored)   exit 0
$ codex -p curator-mcp exec --help                 -> normal exec help (flag accepted before the subcommand)
$ codex exec -p curator-mcp --skip-git-repo-check 'say ok'   -> accepted, ran to the 401 auth failure
$ codex -p bad --strict-config mcp list            -> Error: `--strict-config` is not supported for `codex mcp`   exit 1
```
**Verdict: verified.** A layer file containing only `[mcp_servers.*]` is applied; no other members are required; `-p`
accepts exactly one value (a second is a clap error, not last-wins); a missing layer file is ignored without error (exit 0);
`-p` is accepted both before and after `exec`. Implied text: "`-p <name>` layers `$CODEX_HOME/<name>.config.toml` and may
be given once; a missing layer is silently skipped, so the adapter must stat the file before launch."
Registry: `codex.profile_flag: -p (single, silent-if-missing)`.

---

## Final table

| # | Item | Verdict | Implied change |
| --- | --- | --- | --- |
| 1 | Keychain `oauth.claude.profile.*` | falsified (that account is CodexBar's); per-`CLAUDE_CONFIG_DIR` suffix scheme verified | keychain text + `credential_scope` |
| 2 | `.claude.json` seed shape | verified | minimal seed + key roles |
| 3 | codex global `AGENTS.md` cap | falsified (global file never truncated) | drop cap claim for global doc |
| 4 | codex / pi `auth.json` write mode | verified: both in-place | snapshot copy caveat |
| 5 | fresh-home first run + seeds | verified (login wall first); codex trust seed falsified for `exec` | flags text |
| 6 | Xcode embedded agents root context | requires operator (home dir absent); env-var injection verified from bundle | rename the home |
| 7 | opencode `XDG_CONFIG_HOME` on Windows | not reproducible (host down) | none yet |
| 8 | claude `@path` referenced-form approval | falsified (external targets need the approval key; no prompt in `-p`) | §5.3 + seed key |
| 9 | codex `-p` layer composition | verified | single `-p`, silent-if-missing |

Scratch artifacts: `.temp/TASK-260905-3jq1so/` (claude-home*, codex-home, pi-home, cached `strings` dumps, `ctx.py`).
`/tmp/cx-proj-3jq1so`, `/tmp/cx-nogit-3jq1so`, `/tmp/cl-proj-3jq1so` were created for the probes and removed afterwards.
