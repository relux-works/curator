# TASK-260922-1hla8q: fragment-permission-members-and-transport-token

## Description
F-S2 (0018 choice 7 + Compatibility): fix the exact launch-env fragment member names that carry the profile-level permission mode and the lock engagement, and the minimum Curator/fragment version token that establishes transport support (the token names the revision that defines it); the fail-closed rule stays: unestablished transport support => would-be yolo refused with permission_policy_unsupported, native (explicit or headless/CI/tracked silence) proceeds. Mirror the closed non-interactive marker set {CI, GITHUB_ACTIONS} in environments §10.1 and record that additions come only by spec revision. The launcher follow-up F-L1 consumes these names.

## Scope
curator-spec: environments §10.1/§12.5 (fragment shaping), launcher SPEC §4.1 precondition mirror if the spec hosts it, manager-config-v2 schema for the fragment members, schema cases, CHANGELOG. No launcher code.

## Acceptance Criteria
1) member names and the minimum token value are normative with schema cases (fragment with members valid; v1 fragment carrying a member invalid; unknown value invalid); 2) the fail-closed rule text cites the token; 3) headless marker set versioned in one place and mirrored; 4) make validate green; 5) CHANGELOG entry names F-L1 as the consumer.
