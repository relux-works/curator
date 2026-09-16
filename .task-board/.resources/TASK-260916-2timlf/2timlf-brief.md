# Brief — TASK-260916-2timlf (Astra research report: credential modes + permission interface)

Read-only research in the curator control root; you may read the sibling repositories at /Users/administrator/Developer/ReluxWorks/curator/curator-spec, /Users/administrator/Developer/ReluxWorks/curator/curator-agent-launcher, /Users/administrator/Developer/ReluxWorks/skill-agents-management (agents-infra), and the B5 evidence resource TASK-260908-yl5x3k_onboarding-evidence-rev2.md (board). Do not edit any repository file; write the report to `/tmp` and attach it as `TASK-260916-2timlf_report.md`.

Ground truth already established on host e11-1 (verify, then cite):
- envregistry: claude_code Passthrough darwin={} / linux=.credentials.json file-link; codex_cli auth.json file-link (keyring-preferred variant exists); pi auth.json file-link. Managed homes under ~/.curator/environments/<profile>/<env>. Managed claude home marker: passthrough [] seeds [.claude.json]; the operator had to /login inside `curator run claude_code`; codex worked immediately (symlinked auth.json); pi's native auth.json is empty.
- curator-run flags: --profile, --system-prompt, --model, --effort, --name, --ax-profile <standard|yolo> (ax tracking only; usage error untracked); everything after `--` is forwarded verbatim.
- Legacy agents-infra: `agents-infra claude|codex [-d|--danger|--yolo]` (check the exact mapping in skill-agents-management).
Verify the native flags from the installed tools: `claude --help`, `codex --help` / `codex exec --help`, `pi --help` (Pi 0.84.2 at ~/.local/opt/pi-0.84.2) — quote the exact flag names and semantics. Note where each tool stores credentials on macOS (Claude: Keychain "Claude Code-credentials" and/or .credentials.json; Codex auth.json; Pi auth.json) and whether CLAUDE_CONFIG_DIR changes the Keychain item.

Deliver the report structured exactly as the task AC lists (1–5). Keep it under ~40 KB. Then tick the checklist and `task-board handoff TASK-260916-2timlf --role developer`.
