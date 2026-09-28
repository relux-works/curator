# STORY-260922-39hxog: 0018-launcher-permission-interface

## Description
Follow-up F-L1 of adopted Decision 0018 (curator-spec decisions/0018-curator-run-permission-interface.md, Compatibility section, choices 1, 4, 5, 7; narrowed by operator correction C1 of 2026-09-22 to mode resolution, transport of the resolved mode and provenance only, NO argv grammar and no provider flag spelling - Decision 0013 D5). Control root: curator/curator-agent-launcher. Depends on F-M1 (STORY-260922-h3epwn: the LaunchRequest permission-mode member and the release the launcher pins) and F-S2 (TASK-260922-1hla8q: fragment member names + minimum transport token). Two leaves: L1a launcher SPEC + README revision, L1b implementation with the choice-5 negative rows through the real curator run entry.

## Scope
curator-agent-launcher: SPEC §3 flag table (--permissions native|yolo, --yolo exact alias, -d/--danger rejected), §4.1 fragment members + transport version precondition, §4.3 defaults v2 + item-2 precedence, §4.5 composition placement of the resolved mode, §4.6 tracked refusal from every level + headless detector {CI, GITHUB_ACTIONS} + headless/CI/tracked default native, §4.7 file family note, §6 diagnostics (permission_policy_unsupported, permission_mode_tracked_unsupported, usage), choice-4 effective-native-policy stderr line + launch-record key; README options table; internal/cli parse, internal/composition placement via the agents-management member, internal/execution refusal + provenance; go.mod pin bump to the F-M1 release. No argv grammar, no flag spelling, no ax document change.

## Acceptance Criteria
Both leaves landed on launcher main; the launcher never spells a provider flag (argv grammar test proves the bypass spelling is absent from the launcher module); tracked + effective yolo from any level refused with permission_mode_tracked_unsupported and no fallback to untracked; Decision 0013 D3.6/D5/D6.4 satisfied before any tracked bypass.
