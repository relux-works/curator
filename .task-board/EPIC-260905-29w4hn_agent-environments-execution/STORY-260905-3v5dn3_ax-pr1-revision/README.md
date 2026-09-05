# STORY-260905-3v5dn3: ax-pr1-revision

## Description
Step 1d: revise agent-session-manager-spec PR #1 (draft/curator-environment-integration, head d7075e1) per Decision 0013 Decisions 3, 4, 7, 8 — ax start --launch-plan, SpawnPlan/Launch Plan stdin, caller_launch_plan and stdin_resume_replay capabilities, refuse-on-drift for system modules, CCJ-1 fragment digest, profile-pin as lock hash. Delivered as new commits on the PR branch; the PR is never merged by us. Worktree ~/Developer/ReluxWorks/.temp/ax-curator-integration/worktree.

## Scope
(define story scope)

## Acceptance Criteria
PR #1 branch carries the revision as signed commits on top of d7075e1; validate_spec.py, test_expected_red.sh, git diff --check exit 0; independent review ACCEPT; PR description updated; PR remains open.
