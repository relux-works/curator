# STORY-260928-1xu5sf: rose-air-scriptworker-identity-tests

## Description
After the rose-air rustc PATH fix (2252ebee) the Test (rose-air) lane reaches go test for the first time since 2026-09-24 and fails only internal/scriptworker TestScriptWorkerRejectsForgedWorkerIdentity and TestScriptWorkerRejectsSubstitutedManager with worker session channel closed (run 36387486082). They pass on macos-latest/ubuntu/windows.

## Scope
(define story scope)

## Acceptance Criteria
Test (rose-air) green on a main push
