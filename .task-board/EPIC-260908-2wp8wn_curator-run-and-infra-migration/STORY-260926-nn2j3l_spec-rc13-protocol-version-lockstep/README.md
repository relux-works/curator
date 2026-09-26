# STORY-260926-nn2j3l: spec-rc13-protocol-version-lockstep

## Description
curator-spec v1.0.0-rc.13 release prep (PR #97, TASK-260924-19n6g2) regenerates every core vector with protocol_version 1.0.0-rc.13 (rc.10-rc.12 kept rc.9 labels). curator main hard-codes 1.0.0-rc.9 in internal/scriptpolicy TestScriptExecutionPolicyIdentityMatchesTheSuite, so the spec Implementations Go lane fails on PR #97. Lockstep: curator must accept the rc.13 suite, then the spec PR pins that curator commit.

## Scope
(define story scope)

## Acceptance Criteria
(define acceptance criteria)
