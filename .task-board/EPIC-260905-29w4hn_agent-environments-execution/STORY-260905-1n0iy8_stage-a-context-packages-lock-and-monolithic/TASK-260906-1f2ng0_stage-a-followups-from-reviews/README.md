# TASK-260906-1f2ng0: stage-a-followups-from-reviews

## Description
Follow-ups recorded by the stage (a) reviewers, none blocking: (FU-1) environments §1.1 names profile_source_path_missing and profile_source_path_unreadable and the absence-vs-unreadable rule, but neither diagnostic exists in the code — a missing path operand and an unreadable one both surface as profile_source_invalid with a manifest or read error; implement both with tests that prove neither fires on the other case. (FU-2) loadMachinePolicy returns an empty policy when config.UserPath() does not exist — judge against the absence-vs-unreadable rule and fix or state the bound. (FU-3) the operator warning when a global skill does not migrate into the default profile lock (cycle-2 note). (FU-4) any remaining cycle-5 follow-ups listed in TASK-260905-30zs8t_review-findings-stage-a-5.md. Schedule after stage (a) lands.

## Scope
(define task scope)

## Acceptance Criteria
Each follow-up implemented or explicitly declined with a recorded reason; tests prove the absence-vs-unreadable separation; gates green; PR reviewed and landed by fast-forward.
