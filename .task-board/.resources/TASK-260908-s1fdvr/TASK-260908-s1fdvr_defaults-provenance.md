# TASK-260908-s1fdvr — launcher operator defaults with provenance (host e11-1, 2026-09-16)

File: ~/.config/curator-run/defaults.json (schema curator-run-defaults-v1, SPEC §4.3, operator level; no /etc/curator-run machine file, `locked: false`).

| env-id | model | effort | provenance |
|---|---|---|---|
| codex_cli | gpt-5.6-sol | xhigh | relux-agents-infra main 459742ea67e3c6b84169520b92d74fe7f73e3002 `.configs/codex-config.toml`: `model = "gpt-5.6-sol"`, `model_reasoning_effort = "xhigh"` (what agents-infra installs today) |
| claude_code | claude-fable-5-1 | low | agents-infra `.configs/claude-settings.json` carries `model: "sonnet"` (an alias, not a registry model id); the operator's live `~/.claude/settings.json` on this host sets `model: claude-fable-5-1`, `effortLevel: low`, which is the operator's current preference and a registry id the launcher's registry admits. The worker policy (Astra low / Fable low) is NOT copied into user-facing defaults except where it coincides with the operator's own setting. |
| pi | — | — | no member: SPEC §4.3 level 3 resolves through the recorded Pi runtime preference (pi-anthropic, pi-openai, pi-google) and the vendor lineup |

Written by the orchestrator (host operator action, allowed: launcher-owned configuration directory, not an agent home). Verification: `curator-run` reads it at launch; the origin line-group names `operator` for both members.
