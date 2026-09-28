# TASK-260922-1zfqq0: launcher-spec-permission-interface-revision

## Description
F-L1a: revise the launcher SPEC and README for the 0018 permission interface without any provider flag spelling: §3 flag table (--permissions native|yolo; --yolo as an exact alias of the yolo mode, same increment; -d/--danger rejected), §4.1 fragment members (names from F-S2) and the transport version precondition, §4.3 defaults v2 + item-2 precedence (flag > profile > global > default-interactive > default-headless), §4.5 composition placement of the resolved mode (passed to the agents-management LaunchRequest member, never spelled), §4.6 tracked refusal from every level (permission_mode_tracked_unsupported, no fallback to untracked), the item-5 headless detector with the closed marker set {CI, GITHUB_ACTIONS} (versioned here, mirrored in environments §10.1), headless/CI/tracked silence => native (source=default-headless), lock-engaged rule (visible yolo => usage, silence => native), §4.7 file family note, §6 diagnostics incl. permission_policy_unsupported, choice-4 effective-native-policy stderr line spelling + launch-record extension key. Source: curator-spec decisions/0018 Compatibility section + choices 1, 4, 5, 7.

## Scope
curator-agent-launcher: SPEC.md, README.md, CHANGELOG. No Go code (F-L1b).

## Acceptance Criteria
1) every listed section revised and internally consistent with Decision 0013 D5 (no flag spelling anywhere in the launcher docs; a doc grep row proves --dangerously is absent); 2) diagnostics table lists permission_policy_unsupported, permission_mode_tracked_unsupported and the usage rows with exit codes; 3) headless marker set stated once as closed and versioned; 4) choice-4 line and key spelled; 5) CHANGELOG entry.
