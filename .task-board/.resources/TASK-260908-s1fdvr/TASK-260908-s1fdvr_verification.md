# TASK-260908-s1fdvr — B4 defaults verification

2026-09-16, host e11-1. Evidence-only producer run under zsh, in the assigned STORY-260908-k88yk0 launcher worktree. No code or configuration edits. No installs, login, profile changes, or ax calls.

## Defaults and provenance

Read ~/.config/curator-run/defaults.json successfully (exit 0):

```json
{
  "schema": "curator-run-defaults-v1",
  "locked": false,
  "defaults": {
    "claude_code": { "model": "claude-fable-5-1", "effort": "low" },
    "codex_cli": { "model": "gpt-5.6-sol", "effort": "xhigh" }
  }
}
```

The parsed object matches the recorded settings exactly: schema curator-run-defaults-v1, unlocked, two operator members, no pi member. Python equality assertion exited 0.

Quoted provenance resource TASK-260908-s1fdvr_defaults-provenance.md (retrieved through task-board, exit 0):

# TASK-260908-s1fdvr — launcher operator defaults with provenance (host e11-1, 2026-09-16)

File: ~/.config/curator-run/defaults.json (schema curator-run-defaults-v1, SPEC §4.3, operator level; no /etc/curator-run machine file, `locked: false`).

| env-id | model | effort | provenance |
|---|---|---|---|
| codex_cli | gpt-5.6-sol | xhigh | relux-agents-infra main 459742ea67e3c6b84169520b92d74fe7f73e3002 `.configs/codex-config.toml`: `model = "gpt-5.6-sol"`, `model_reasoning_effort = "xhigh"` (what agents-infra installs today) |
| claude_code | claude-fable-5-1 | low | agents-infra `.configs/claude-settings.json` carries `model: "sonnet"` (an alias, not a registry model id); the operator's live `~/.claude/settings.json` on this host sets `model: claude-fable-5-1`, `effortLevel: low`, which is the operator's current preference and a registry id the launcher's registry admits. The worker policy (Astra low / Fable low) is NOT copied into user-facing defaults except where it coincides with the operator's own setting. |
| pi | — | — | no member: SPEC §4.3 level 3 resolves through the recorded Pi runtime preference (pi-anthropic, pi-openai, pi-google) and the vendor lineup |

Written by the orchestrator (host operator action, allowed: launcher-owned configuration directory, not an agent home). Verification: `curator-run` reads it at launch; the origin line-group names `operator` for both members.

## Direct launch verification

Each command below ran directly as a separate foreground process, without a pipe or tee. Output quotes are combined stdout/stderr as captured by the execution tool. Every launch exited 0. These are fresh launches, not inferred from B5 evidence. Version mode requires no login; it establishes defaults resolution and native dispatch, not authenticated inference or MCP connectivity.

### curator run codex_cli -- --version

Exit 0:
```text
warning: environment_tool_version_unverified: codex_cli detected 0.153.4, recorded 0.153.2
curator-run: defaults: model=gpt-5.6-sol (operator) effort=xhigh (operator)
codex-cli 0.153.4
```

### curator run claude_code -- --version

Exit 0:
```text
warning: environment_tool_version_unverified: claude_code detected 2.1.273, recorded 2.1.261
curator-run: defaults: model=claude-fable-5-1 (operator) effort=low (operator)
2.1.273 (Claude Code)
```

### curator run pi -- --version

Exit 0:
```text
curator-run: defaults: model=claude-fable-5 (lineup) effort=high (lineup)
0.84.2
```

## Scope and checklist mapping

Items 1–3: exact defaults/provenance match and 3/3 fresh native launches print the required origin lines with exit 0. B5 onboarding itself is prior evidence (TASK-260908-yl5x3k_onboarding-evidence-rev2.md), not repeated here. Historical agents-infra commit provenance is quoted from the attached source, not independently re-audited.

Items 4 and 6: this task-scoped artifact is attached before handoff. The final board command is the required developer handoff, which owns the landing suite and stores its validation evidence. No manual duplicate landing suite.

Item 5: satisfied as N/A per explicit verification-only brief; no code changes, so no new unit tests or build needed. The three real launch commands are the relevant validation.

Item 7 / logbook: tool-version drift warnings remain for Codex and Claude; all required origins still resolve. Findings recorded here and in board notes because campaign rules prohibit LOGBOOK.md edits. No new product decision or forced fit.

Independent review and signed integration (if applicable) remain the orchestrator's responsibility; this producer does not attest its own independent acceptance. Worktree was clean before the evidence artifact, HEAD  recorded separately by command output. Evidence files live under ignored .temp; no repository delta intended.
