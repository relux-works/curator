# BUG-260928-2bfhgj: rose-air-scriptworker-forged-identity-channel-closed

## Description
Test (rose-air) run 36387486082 (head 2252ebee): internal/scriptworker TestScriptWorkerRejectsForgedWorkerIdentity (worker_test.go:546) and TestScriptWorkerRejectsSubstitutedManager (worker_test.go:601) fail with 'cannot read a worker message: worker session channel closed' — the worker session ends before the expected refusal message arrives. Green on hosted macOS/ubuntu/windows. rose-air is a self-hosted macOS runner (check arch/OS version/codesign/SIP/sandbox differences). Evidence artifact test-evidence-rose-air.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
Root cause identified with evidence; fix in product or test (no weakening of the forged-identity / substituted-manager refusal); Test (rose-air) green on main
