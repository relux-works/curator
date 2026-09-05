# TASK-260905-3jq1so — review findings, sprint cycle 1

Reviewer run: `$TASK_BOARD_RUN_ID` (reviewer, read-only). Subject: outcome resource
`TASK-260905-3jq1so_verification-sprint.md` (researcher RUN-260905-518f79, window 07:58:28Z–08:23:56Z).
Reproduction ran under `.temp/TASK-260905-3jq1so/review/` scratch homes only. Binary versions re-recorded:
`claude 2.1.261`, `codex-cli 0.153.2`, `pi 0.84.2`, opencode not installed, `Xcode 26.5 (17F42)`.

## 1. Reproduced items (verdict-changing)

### Item 1 — Keychain keying. AGREE (falsified + verified)

```
security dump-keychain 2>/dev/null | grep -A12 'oauth.claude.profile' | grep -E '"(svce|labl|acct)"'
    "acct"<blob>="oauth.claude.profile.3cd05c28…"
    "svce"<blob>="com.steipete.codexbar.cache"
security find-generic-password -s 'Claude Code-credentials'     (no -w)
    "acct"<blob>="iv"   "svce"<blob>="Claude Code-credentials"
security dump-keychain | grep '"svce"<blob>="Claude Code'  -> Claude Code-credentials (1), Claude Code-doctor-probe (1)
grep -a -c -F 'oauth.claude.profile' ~/.local/share/claude/versions/2.1.261   -> 0
function Sx(n=""){let e=process.env.CLAUDE_SECURESTORAGE_CONFIG_DIR,t=e!==void 0?!e:!process.env.CLAUDE_CONFIG_DIR,
  r=e!==void 0?e.normalize("NFC"):be(),c=t?"":`-${a("sha256").update(r).digest("hex").substring(0,8)}`;
  return`Claude Code${Vt().OAUTH_FILE_SUFFIX}${n}${c}`}
```
Attack: is `Sx` the production call site or a dead helper? The bundle's keychain read is
`let a=Sx(N5),o=tv(),s=KQ(`security find-generic-password -a "${o}" -w -s "${a}"`…)` — the suffixed
service name feeds the real lookup. `OAUTH_FILE_SUFFIX` is `""` / `"-local-oauth"` / `"-custom-oauth"`
(build variants), consistent with the plain `Claude Code-credentials` item seen. Note: the researcher
labels `be()` as the config-dir reader; minified `be` is reused for several functions in the dump, so that
name is not independently confirmed here, but the env-var gate `t` (suffix only when `CLAUDE_CONFIG_DIR`
or `CLAUDE_SECURESTORAGE_CONFIG_DIR` is set) is unambiguous. Verdict stands.

### Item 3 — global `$CODEX_HOME/AGENTS.md` cap. AGREE (falsified)

Fresh files, 41015 bytes each, marker at byte 33066 (> 32768), scratch git project:
```
CODEX_HOME=<scratch> codex debug prompt-input 'hello'                       exit 0
   ZEBRA(global) True  KOALA(project) False  filler lines 1366  (= 759 global + 607 project ≈ 32768 B)
CODEX_HOME=<scratch> codex -c project_doc_max_bytes=1000 debug prompt-input 'hello'
   ZEBRA True  KOALA False  filler 778  (= 759 global + 19 ≈ 1000 B project)
control, global AGENTS.md removed:  ZEBRA False  KOALA False  filler 607
```
Narrowed gate (1000 B) truncated the project doc and left the global file whole; the control proves the
global file is the only ZEBRA source. Verdict stands.

### Item 8 — `@path` referenced-form approval. AGREE (falsified), with one gap

Bundle (cached dump of 2.1.261):
```
function xZ(e){return Ap(e,he())}
function aD(e,t,r,o,d=0,p,_){… let C=o&&(t!=="User"||egr()); … if(d>0&&!C&&!xZ(I))return[];
  if(t==="User"&&!C)try{let V=await ae().lstat(e);if(d===0&&V.isSymbolicLink()||(V.nlink??1)>1&&V.isFile())return[]}catch{}
  … for(let V of F){if(!xZ(V)&&!C)continue; …aD(V,t,r,o,d+1,e)…}}
function egr(){return M1()!=="local-agent"}
… D=o||I.hasClaudeMdExternalIncludesApproved||!1 … where I=es() = t.projects[<cwd key>] (per-project entry)
'Yes, allow external imports' present (2)
```
Scratch `-p` run (fake `ANTHROPIC_API_KEY`, scratch home, `@/etc/hosts` in both user and project
`CLAUDE.md`): no prompt, exit 124 from API-retry timeout; debug log carries no include lines. Same as the
researcher: injected content unobservable without a logged-in run.

**Gap (minor, not verdict-changing):** the same guard also drops the **user-level `CLAUDE.md` itself** when
it is a symlink (depth 0) or a hard link (`nlink > 1`) and approval is not set. The environments text's
referenced/linked forms must therefore state: with `hasClaudeMdExternalIncludesApproved` unset, a
symlinked or hard-linked `$CLAUDE_CONFIG_DIR/CLAUDE.md` is silently skipped, not just its external `@`
targets. Add this to the item 8 implied change.

### Item 9 — codex `-p` layer composition. AGREE (verified)

```
$ codex mcp list                        -> No MCP servers configured yet.            exit 0
$ codex -p curator-mcp mcp list         -> curator-probe /usr/bin/true enabled        exit 0
$ codex -p curator-mcp -p second mcp list -> error: '--profile <CONFIG_PROFILE_V2>' cannot be used multiple times   exit 2
$ codex -p nonexistent mcp list         -> No MCP servers configured yet.            exit 0
$ codex -p nonexistent --strict-config exec --skip-git-repo-check 'say ok' -> proceeds to 401 (missing layer not an error even under --strict-config)
$ layer with unknown field + --strict-config exec -> Error loading config.toml: unknown configuration field `mcp_servers.x.model`  exit 1
```
Extra fact for the registry: a missing layer is ignored even with `--strict-config`, so stat-before-launch
is the only detection; `--strict-config` does validate the layer's contents. Verdict stands.

## 2. Spot-checks (items 2, 4, 5, 6)

- Item 2: scratch home `claude-home/.claude.json` re-read; top-level keys match the report exactly
  (`firstStartTime, firstStartVersion, hasResetAutoModeOptInForDefaultOffer, machineID, migrationVersion,
  opusProMigrationComplete, seenNotifications, sonnet1m45MigrationComplete, userID`). `es()` confirms the
  trust/approval keys live under `projects.<cwd>`. Evidence-backed. OK.
- Item 4: pi `dist/core/auth-storage.js` lines 13/28/65/140 use `writeFileSync` (mode 0600), no `rename`;
  `settings-manager.js:93` likewise. codex strings: `cli_auth_credentials_store = "file"` and
  `login/src/auth/storage.rs` present (4 hits). The codex in-place verdict rests on upstream source for the
  path named in the binary, clearly labelled as such. Acceptable; "unknown" would also have been
  defensible for codex, but the citation is explicit and version-matched by file path.
- Item 5: each wall has a command and exit code; "requires operator" used only for OAuth/API login. The
  falsified `trust_level` seed claim is backed by a positive control (`mcp_servers` from the same file
  visible in `codex mcp list`). OK.
- Item 6: framework strings reproduced (`Set CLAUDE_CONFIG_DIR to …`, `Override CODEX_HOME`,
  `IDEChatOverrideAgenticHomeDirectory`, `Short hash (first 8 chars)`); no IDEIntelligenceAgents dir under
  `~/Library` (depth 4). "Requires operator" is legitimate: the home does not exist until Xcode launches an
  agent (GUI). OK.
- Item 7: `ssh win` unreachable — recorded, no change derived. OK.

## 3. Compliance

- Report and scratch tree grepped for token/JWT/`sk-ant` shapes (excluding the two `strings` dumps):
  no hits. Keychain inspected without `-w`/`-d`. Real `.claude.json` inspected by key names only.
- Real-file metadata now: `~/.claude.json` (rewritten continuously by live claude sessions, including this
  one — not attributable), `~/.pi/agent/auth.json` mtime 1782216150 and `settings.json` mtime 1787098899
  (both far before the run: untouched), `~/.codex/auth.json` inode 734845569 size 4178
  **mtime 1788596600 = 08:23:20Z, inside the run window (36 s before completion)**. The report's before/
  after stat (1788560952) brackets only the `codex login status` command. Same inode and size are
  consistent with codex's own in-place token refresh (item 4), and the researcher's `codex exec` probes
  ran under scratch `CODEX_HOME`s which cannot reach the real file. The logwork file is a summary only, so
  the writer cannot be attributed read-only. **Unknown cause, recorded as such** (cf. Evidence That
  Counts: failure to read ≠ absence). Not a demonstrated violation; the brief itself sanctioned
  `codex login status` against the real home. Severity: minor observation, no rework required; future
  briefs should bracket the whole run with stat, not one command.
- No repository file was changed (candidate tree == base), which is correct for a research-only leaf.

## 4. Implied-change judgement

- Item 1: lifting the "isolated homes cannot get separate credentials" claim for claude_code/macOS on
  bundle + Keychain-listing evidence is acceptable **with** a pinned-release gate (2.1.261) and the
  requires-operator residual (an actual login in a scratch `CLAUDE_CONFIG_DIR` was never performed; the
  suffixed item's creation is inferred from the same function feeding the read path). Registry wording
  `credential_scope: per CLAUDE_CONFIG_DIR (keychain suffix sha256[0:8])` follows from the evidence.
- Item 3: dropping the cap claim for the global doc follows directly from the narrowed-gate experiment.
- Item 8: the §5.3 replacement follows; extend it with the symlink/hard-link skip above.
- Item 9: "single `-p`, silent-if-missing, stat before launch" follows; add "even under `--strict-config`".
- Items 4/5/6: caveats are proportionate; none overreach.

## Verdict table

| # | Researcher verdict | Reviewer | Severity of disagreement |
| --- | --- | --- | --- |
| 1 | falsified / verified | agree | — |
| 2 | verified | agree | — |
| 3 | falsified | agree (reproduced, narrowed gate) | — |
| 4 | verified in-place | agree | — |
| 5 | verified / trust seed falsified | agree | — |
| 6 | requires operator | agree | — |
| 7 | not reproducible | agree | — |
| 8 | falsified | agree, plus symlink/hard-link gap | minor (text addition) |
| 9 | verified | agree, plus `--strict-config` fact | minor (text addition) |

**ACCEPT.** Two minor additive text notes (items 8, 9) and one compliance observation (codex `auth.json`
mtime inside the run window, cause unknown) — none blocking, to be folded into the environments 1.1
edit by the next producer.
