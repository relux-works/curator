# TASK-260922-1s0zja: permission-mode-member-and-provider-mapping

## Description
F-M1a: add the LaunchRequest permission-mode member for LaunchModeInteractive (values native|yolo; native = pass nothing, provider stored settings decide; yolo = the single provider bypass flag) with the per-environment mapping per pinned tool release: claude_code -> --dangerously-skip-permissions, codex_cli -> --dangerously-bypass-approvals-and-sandbox, pi -> its documented flag or an explicit unsupported refusal. The flag spelling lives only here (argvguard). Positive and negative interactive goldens per tool release. Source: curator-spec decisions/0018 adoption choices 1, 3, 6 and the Compatibility section (F-M1).

## Scope
skill-agents-management: LaunchRequest type + validation, interactive exec plugins for claude/codex/pi, argvguard, internal/regress interactive goldens, docs. No capability-table versioning yet (F-M1b), no launcher change.

## Acceptance Criteria
1) member exists, validated: yolo|native only, LaunchModeInteractive only, else a named refusal diagnostic; 2) goldens per pinned tool release show the exact argv for native (no flag) and yolo (exact flag, once, before the prompt text) for claude_code and codex_cli, pi row documented; 3) negative goldens: unknown value, yolo with a non-interactive LaunchMode, yolo combined with a raw bypass flag after -- (duplicate spelling refused), each with a narrowing mutant executed and killed; 4) argvguard test proves one spelling site per flag; 5) docs + CHANGELOG entry.
