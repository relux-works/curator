# BUG-261018-xait57: trust-seed-uses-empty-project-path

## Description
env resolve claude_code --profile P --repair run from a launch directory writes .claude.json projects[""].hasTrustDialogAccepted=true instead of an entry for the launch directory, so task-board then refuses claude_goal_workspace_untrusted. Seen twice on agents1 (Linux, rc.7); ivmbp needed the same one-off fix earlier.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
(define fix acceptance criteria)
