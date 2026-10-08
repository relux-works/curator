# Review verdict — TASK-261004-3s6ymq rev1: ACCEPTED

Scope: the CR adds one file, `.research/261004_launcher_rc3_compat_smoke.md`. Read-only review; no builds (R193/R194); nothing re-executed. I checked the record against the task-scoped `rows.jsonl`, the harness `check.py`, and the spawn log.

## Checks

- **Patch = attached record.** The `+` lines of `rev1.patch` are byte-identical to `TASK-261004-3s6ymq_smoke.md` (diff clean apart from the hunk header).
- **Every table row maps to an executed command with its exit.** The record's 45-row table equals the 45 `rows.jsonl` entries, same names and same exits (0/1/2). The `argv` fields match the table's commands.
- **Prose claims against row stderr/stdout:**
  - `curator-run --version` prints `curator-run 0.2.0 (specification 0.5.0-draft)`, exit 0.
  - Locked yolo exits 2 with `usage: yolo conflicts with Curator's force-native permission lock`; locked native exits 0.
  - Tracked yolo exits 1 with `permission_mode_tracked_unsupported`.
  - Pi yolo exits 1 with `permission_mode_unsupported` (pi 0.84.2 has no bypass flag).
  - Exact-once yolo flags appear in the capture argv: claude `--dangerously-skip-permissions`, codex `--dangerously-bypass-approvals-and-sandbox`, muse `--yolo`.
  - Muse capture shows HOME preserved and four XDG parents under the managed home.
  - Prompt collisions (claude, codex) and MCP/profile collisions (claude `--mcp-config`, codex `--profile`, codex `--config=mcp_servers…`) exit 1 with `plan_refused … conflict with fragment channel`.
  - Wrong version exits 1 (`resolve_fragment_invalid`, `launch-env-fragment-v99`); malformed digest exits 1 (`resolve_fragment_invalid`, `lock_sha256`); legacy fragment with yolo exits 1 (`permission_policy_unsupported`).
  - Discovery: `subcommand_provider_untrusted` for the environments root, `.local/bin` (declared and published) and the global skill bin. The unmanaged user-bin and skill-bin cases (exit 0) carry `subcommand_provider_outside_trust_roots`, as the record says.
  - Initial pre-capture rows are exit 1 because the native binary or release was absent. The record labels these bounded.
- **README/CHANGELOG row.** Prose says "assertion process exit 0". The spawn log shows the assertion block succeeding with `README install/support and CHANGELOG version/Muse checks: PASS`. It is not a `rows.jsonl` entry; see N1. The launcher tag `v0.2.0` resolves to `2517d27`, the candidate the record names.
- **Secrets and personal paths.** The patch and `rows.jsonl` contain no `/Users/…` paths, e-mails or token-like strings. All paths are `/tmp/launcher-rc3-smoke/...`. Muse's scratch `auth.json` is `{}`. This is consistent with AGENTS.md and tb-R163, and no other operator's product material is included.
- **Claims within the rows.** The record says GO is bounded, native tools were absent, capture substitutes prove argv/env transport only, and MCP rows are resolver-boundary mutations. It does not claim provider behavior, real MCP installation, or release-signature authenticity.

## Non-blocking notes

- **N1.** The README/CHANGELOG assertion and the Python `check.py` assertions are not rows in `rows.jsonl`/the table. Their exit 0 is evidenced only by the spawn log, which is attached to the task. Acceptable, since the record presents them as assertion processes and not as table rows.
- **N2.** The record's relative links `../SPEC.md`, `../README.md` and `../CHANGELOG.md` resolve inside the curator repo. There `SPEC.md` does not exist, and `README.md`/`CHANGELOG.md` are curator's own, not the launcher's frozen files. These are navigation hints only (link rot), not evidence.
- **N3.** The table has no column marking pre-capture rows (absent tool, bounded) apart from the later `capture-*` rows. The prose covers it.

Verdict: accepted. GO for the v0.2.0 tag stands within the stated bounded scope; the tag was already placed on `2517d27`.
