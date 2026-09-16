# TASK-260916-2timlf — Review verdict (Claude Fable, independent second opinion)

Reviewed: CR-TASK-260916-2timlf-1 rev 1, base/candidate 18f05497 (repository delta empty), outcome resources
TASK-260916-2timlf_report.md (24,947 B) and TASK-260916-2timlf_native-help.txt. Shell zsh, host e11-1 (Darwin 24.6.0),
2026-09-16. Read-only: no repository edits, no logins, no bypass flags executed, no secret contents read
(Keychain queried without `-w`; files inspected by `ls`/size only).

## VERDICT: ACCEPT (report usable as the basis of the two decision drafts, with the corrections below appended)

Why an empty repository delta is the right outcome: the task is a research report ("no code"); the brief directs the
producer to write the report to /tmp and attach it as a board outcome. Both deliverables are attached; AC sections
1–5 are all present and, as fact-checked below, correct except for the corrections in §2.

## 1. Fact-check table (claim → verified / corrected, with citation)

| # | Report claim | Result | Citation (worktree 18f05497 unless noted) |
|---|---|---|---|
| 1 | claude_code Passthrough darwin={} / linux .credentials.json file-link; seeds [.claude.json]; VerifiedRelease 2.1.261; CredentialScope "per CLAUDE_CONFIG_DIR (keychain service suffix sha256[0:8]) on macOS; home on Linux" | VERIFIED | internal/envregistry/envregistry.go:193-202 |
| 2 | codex_cli auth.json keyring-preferred (file-link unless native config.toml selects keyring); seed config.toml; pin 0.153.2 | VERIFIED | envregistry.go:216-221; managed.go:497-547 (`codexKeyring`, regex `cli_auth_credentials_store`) |
| 3 | pi auth.json file-link, pin 0.84.2 | VERIFIED; correction: registry seeds are [settings.json, models.json] (envregistry.go:267) — the report's "seeds []" describes the host marker, not the registry | envregistry.go:263-273 |
| 4 | macOS Claude: isolated default, shared refused (DiagSharedUnsupported) at/above pin; below pin shared default, isolated refused | VERIFIED — the report does account for CredentialScope/DiagSharedUnsupported | envregistry.go:388-408 |
| 5 | opencode isolated refused (no-op) | VERIFIED | envregistry.go:389-391 |
| 6 | Curator resolves Pi native home to ~/.pi and links ~/.pi/auth.json, while Pi 0.84.2 stores tokens in ~/.pi/agent/auth.json | VERIFIED | internal/envprofile/switch.go:78 (DefaultDir ".pi"); managed.go:128-137; ~/.local/opt/pi-0.84.2/pi/docs/providers.md:26,111,139; host: ~/.pi/auth.json 2 B, ~/.pi/agent/auth.json 127 B; managed pi/auth.json -> /Users/administrator/.pi/auth.json |
| 7 | Hazard 1: shared→isolated does not remove the old credential link | VERIFIED by inspection: effectivePassthrough returns {} for isolated (managed.go:500-502); applyPlan's removal set is built from marker Surfaces only (managed.go:1670-1677); finalizeMarker only loops over new links (managed.go:934-942); checkPassthrough compares recorded vs effective sets only (managed.go:1342-1345). Not reproduced live; needs a production-entry test |
| 8 | Hazard 2: finalizeMarker removes the wanted-link path before symlinking, without credential-specific preservation | VERIFIED: `_ = os.Remove(full)` at managed.go:936; conflicts with manager.md:2365-2368 |
| 9 | Backups skip symlinks; existing generation refused; takeover backs up replaced files | VERIFIED | managed.go:823-850, 1653-1666 |
| 10 | Spec locations: manager §7 is source audit; credentials are manager §12.4; adapters/knobs are environments §7 / §12.1 | VERIFIED | curator-spec a68854d: profiles/manager.md:1008, 2343; protocol/environments.md:944, 2240 |
| 11 | isolation lockable only toward shared; system map replaces user map whole | VERIFIED | manager.md:51-73; environments.md:2285-2297; internal/config/environments.go:570-598, 796-816 |
| 12 | secret_material_waivers waives an audit finding (pin/file/span), not credential copying | VERIFIED | environments.md:1566-1582, 2267 |
| 13 | passable_env_names bounds MCP env-name passthrough only | VERIFIED | environments.md:2130-2135, 2264 |
| 14 | Launcher flags: --profile, --system-prompt, --model, --effort, --name, --ax-profile <standard|yolo>; repeated/unknown flags usage error; --ax-profile untracked is usage error; native args after -- verbatim | VERIFIED | curator-agent-launcher b34e1e2: internal/cli/cli.go:31-46, 227-269, 296-303; README.md:113-132; SPEC.md:139-143 |
| 15 | opencode refused by launcher mapping | VERIFIED | internal/mapping/mapping.go:22-32 |
| 16 | defaults.json closed schema {model, effort} | VERIFIED | SPEC.md:324-355 |
| 17 | Composed ax document never carries a bypass flag; ax plugin refuses one from native argv (Decision 0013 D5/D3.6); interactive plan; native args appended to Argv | VERIFIED | SPEC.md:623-629; internal/plan/plan.go:227; internal/composition/composition.go:104; internal/execution/execution.go:46,74,139 |
| 18 | Legacy: `agents-infra claude|codex -d|--danger|--yolo` expand to `--dangerously-skip-permissions` / `--dangerously-bypass-approvals-and-sandbox`; --print-config shows source wrapper:-d / wrapper:--yolo | VERIFIED (reran print-config, no launch): relux-agents-infra-main 459742ea tools/agents-infra/internal/infra/claude_launch.go:12,419-427; codex_launch.go:16,567-575 | note: sibling checkout relux-agents-infra (c2da9ff) has a different layout; cite only the -main checkout |
| 19 | Module A interactive builders omit bypass flags | VERIFIED | skill-agents-management 7f0b6bc pkg/agentic/systems/claude/args.go:56-58, codex/args.go:44-46 |
| 20 | Claude 2.1.273 --help: `--dangerously-skip-permissions` "Bypass all permission checks."; `--allow-dangerously-skip-permissions` "Enable bypassing all permission checks"; `--permission-mode` choices acceptEdits, auto, bypassPermissions, manual, dontAsk, plan | VERIFIED (reran `claude --help`, lines 18, 65, 144-147) |
| 21 | Codex 0.153.4 `codex --help` and `codex exec --help` both list `--dangerously-bypass-approvals-and-sandbox`, `--approve-for-me` (workspace-write), `--dangerously-bypass-hook-trust`; `-s/--sandbox` read-only|workspace-write|danger-full-access; `-a/--ask-for-approval` on-request|never only in top-level help; no `--full-auto` | VERIFIED (reran both helps) |
| 22 | Pi 0.84.2 `pi --help` has no permission-bypass flag; only `--approve/-a` and `--no-approve/-na` (trust project-local files) | VERIFIED (reran, lines 55-56) |
| 23 | Versions: Claude 2.1.273, Codex 0.153.4, Pi 0.84.2 | VERIFIED |
| 24 | B5 evidence: Claude exit 1 "Not logged in", passthrough []; Codex exit 0 OK with auth.json link; Pi exit 1 "No API key found for anthropic"; managed homes list | VERIFIED | TASK-260908-yl5x3k_onboarding-evidence-rev2.md:594, 796-806, 808-812, 842-850, 1011-1014 |
| 25 | Claude Keychain item "Claude Code-credentials" account $USER; suffix under CLAUDE_CONFIG_DIR is accepted pinned evidence | PARTIALLY CORRECTED — see §2.1 (host evidence contradicts the residual assumption) |
| 26 | Binary contains string CLAUDE_SECURESTORAGE_CONFIG_DIR (presence only) | VERIFIED (`strings` on the claude binary) |

## 2. Corrections and disagreements

### 2.1 Material new observation: the managed Claude home on e11-1 holds a JSON credential file, and no suffixed Keychain item exists
- `~/.curator/environments/relux-root-context-ivan/claude_code/.credentials.json` exists, mode 0600, 819 B, mtime 2026-09-16 04:30 (the operator's /login inside `curator run claude_code`).
- The login Keychain holds exactly one generic password with service `Claude Code-credentials` (account `administrator`), created 2026-07-08, last modified 2026-08-17 — i.e. NOT touched by the 04:30 managed-home login, and no `Claude Code-credentials-<8hex>` item was created.
- The native `~/.claude/.credentials.json` (509 B, mtime 02:17) also exists.
Consequence: the environments §7.4 residual (environments.md:1119-1123 "a fresh login inside a managed home writes the suffixed item … requires an operator to confirm") is answered NEGATIVELY on this host at 2.1.273: the managed-home login produced a file under CLAUDE_CONFIG_DIR, not a suffixed Keychain item. The cause is unknown (Keychain write refused/failed in that session, or file store chosen for non-default config dirs); the report's row "Keychain … suffixed for CLAUDE_CONFIG_DIR; JSON fallback" states the pinned assumption and misses this host fact. It must be appended to §1 of the report and to the credential decision draft. Design impact: a `.credentials.json` file-link (the Linux row) may be viable for macOS `shared` at ≥2.1.273, which would let `DiagSharedUnsupported` be lifted — but ONLY after an authorized, disposable-account test proves (a) which store Claude reads first when both file and Keychain exist, (b) whether the file is rewritten in place on refresh, (c) whether the suffixed Keychain item is written under other conditions. Until then keep isolated-by-default and the shared refusal exactly as coded. Secrets must never be exported from Keychain to JSON by the manager.

### 2.2 Minor corrections
- Pi registry seeds are [settings.json, models.json] (envregistry.go:267); "seeds []" was the marker's view.
- The Pi root discrepancy should be stated as a concrete change: NativeHome for `pi` must resolve to `~/.pi/agent` (Pi treats PI_CODING_AGENT_DIR as the replacement for `~/.pi/agent`, so the managed-home link path `auth.json` is right and only the native target is wrong). The existing `~/.pi/auth.json` (2 B) is a manager-created artefact of the wrong target and must be handled by migration (unlink the recorded link, never delete the native file).
- Report §3 says "Exec help does not list -a": correct for the top-level flag table; `codex exec` still accepts `-c approval_policy=...` overrides, so the conflict check must also cover `-c`/`--config` keys `approval_policy`, `sandbox_mode`, `sandbox_permissions`.

### 2.3 Disagreements (none blocking)
- The report defers `--yolo` as an alias; I recommend shipping `--yolo` as an exact alias of `--permissions yolo` in the same increment (parity with the legacy wrapper is a stated driver) and rejecting `-d`/`--danger` entirely (Claude's native `-d` is debug).
- The report proposes the value spelling `native|yolo`; agreed, but the launcher must print the resolved posture on stderr in untracked mode so an operator can distinguish "native" from "nothing was requested".

## 3. FINAL recommended design

### (a) Credential modes
- Knob: keep `environments.isolation.<profile>.<env-id> = shared|isolated` (manager-config schema 2). No new knob, no per-run flag. Locking stays "toward shared only"; enforced fleet isolation is a separate policy decision.
- Strategies (registry-owned, per env × GOOS): codex_cli keyring-preferred → file-link `auth.json` (unchanged); pi file-link `auth.json` with native root corrected to `~/.pi/agent`; claude_code linux file-link (unchanged); claude_code darwin: isolated default, shared refused, with an explicit follow-up experiment per §2.1 before any change; opencode ambient only. Rejected: keychain-shared (no manager-linkable item; would require secret handling), copy-at-provision (contradicts manager §12.4 no-copy boundary; waiver is not consent).
- Fix first (curator internal/envprofile): (1) shared→isolated must unlink the recorded credential symlink when it still targets the declared native store, and refuse on a detached/regular file; (2) finalizeMarker must not `os.Remove` a regular file at a link path — refuse with a credential-conflict diagnostic; (3) marker records `isolation` and strategy per entry (currently `isolation: None` on the host marker).
- Migration for existing managed homes: an explicit inspect → plan → apply step under the manager lock (not silent in `resolve --repair`): inventory old marker + link targets + both Pi roots; preserve effective mode on upgrade; operator chooses account on isolated→shared conflicts; no secret copies; publish mode/strategy/provenance in the marker atomically.
- Spec touchpoints: environments.md §7.4 (pi target, marker fields, migration rule, §2.1 residual outcome), §7.7 diagnostics, §10.1 repair, §12.1/12.2 unchanged; manager.md §12.4 (conflict rule replaces the implicit remove). Frozen v1 protocol schemas untouched; marker/config changes are manager-config schema 2.

### (b) `curator run` permission interface
- Flag: `--permissions <native|yolo>` before `--`; default `native` (= no launcher override, forward argv verbatim). `--yolo` is an exact alias. Never read from defaults.json (closed schema rejects any permission member), never derived from profile, model, ax-profile, credentials or environment.
- Per-env native mapping (verified from installed `--help`):

| env | yolo maps to | placement |
|---|---|---|
| claude_code | `--dangerously-skip-permissions` | before native args, interactive |
| codex_cli | `--dangerously-bypass-approvals-and-sandbox` | top-level and after `exec` (both helps list it) |
| pi | none — refuse `permission_mode_unsupported` (Pi has no bypass flag; `--approve` is not equivalent) | — |
| opencode | `env_unsupported` (already) | — |

- Refusal rules (usage exit 2 unless noted): explicit yolo plus a conflicting native selector after `--` — Claude `--permission-mode`, `--allow-dangerously-skip-permissions`, duplicate `--dangerously-skip-permissions`; Codex `-a/--ask-for-approval`, `-s/--sandbox`, `--approve-for-me`, `--dangerously-bypass-*`, `-c`/`--config` `approval_policy|sandbox_mode|sandbox_permissions` in `=` and separate forms; invalid value, repeated flag; unknown env/version → fail closed. Native mode performs no inspection of argv (raw contract unchanged).
- ax interplay: `--ax-profile` unchanged and independent (tracking-only). Tracked mode + `--permissions yolo` is refused (`permission_mode_tracked_unsupported`) until a versioned ax capability admits a permission posture (SPEC §4.6, Decision 0013 D5/D3.6); no fallback to untracked.
- Provenance: one stderr line `curator-run: permissions=<native|yolo> source=<flag|alias|absent> mapped=<flag or none>`; tracked document schema unchanged; no argv or credential dumps.
- Touchpoints: launcher SPEC §3 (flag table), §4.2 (mapping table), §4.3 (defaults exclusion), §4.5 (composition placement/conflicts), §4.6 (tracked refusal), §6 (diagnostics); README options table; internal/cli parse, internal/mapping capability table, internal/composition placement, internal/execution refusal + provenance; negative tests driving the real `curator run` entry with a fake process.

## 4. What the two decision drafts must contain
- Credential draft: the §2.1 host observation with dates/sizes; the Pi root correction; the two migration hazards with file:line; the explicit no-copy/no-Keychain-export boundary; the darwin experiment protocol (disposable account, both stores present, refresh behaviour) as the gate for lifting `DiagSharedUnsupported`; the marker/config fields added; spec section list above.
- Permission draft: flag grammar and alias; mapping table with `--help` quotes and tool versions; the full refusal list including `-c` overrides; tracked-mode refusal and the ax admission precondition; the stderr provenance line; the statement that defaults.json and any config surface can never set yolo; the Pi/opencode refusals.

## 5. Review accounting
Reran independently (all exit 0): `claude --help`, `codex --help`, `codex exec --help`, `pi --help`, version probes, `agents-infra claude|codex --print-config -d|--yolo`, `security find-generic-password -s "Claude Code-credentials"` (attributes only), `security dump-keychain` grep (no `-d`), `ls -la` of native and managed credential paths, `strings` on the claude binary. Not rerun: B5 prompts, any login, any bypass launch, Linux/Windows lanes (unverified, not passing). No landing suite run (research task, empty delta).
