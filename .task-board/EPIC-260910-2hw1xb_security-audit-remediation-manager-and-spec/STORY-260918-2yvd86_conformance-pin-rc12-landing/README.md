# STORY-260918-2yvd86: conformance-pin-rc12-landing

## Description
Landing story for the rc.12 conformance-pin promotion union: the union accepted at TASK-260917-16l2md revision 5 (checkpoint 73fc8a4 on base 3c45d4b, STORY-260917-3w3lvj) could not be replayed onto the moved curator trunk because task-board worktree refresh-candidate fails with INTERNAL_ERROR change_request_checkpoint_conflict on a story carrying an empty-delta second checkpoint (TASK-260917-2ecpjv). This story forks a fresh workspace from the current trunk and lands the combined tree — the accepted union + the S6 landing (64cacfc) + the rose-air CI fix (d00fe7a) with the three additive conflicts resolved as unions and a test-only fix-up — as one reviewed story-final Change Request through the board-owner integrate path. STORY-260917-3w3lvj is closed by tree-carried evidence once this lands.

## Scope
(define story scope)

## Acceptance Criteria
The landed tree = current trunk + exactly the accepted rev-5 union hunks (per-file patch-id identical except CHANGELOG.md, cmd/curator/envstatus.go, internal/envprofile/status.go resolved as unions and cmd/curator/main.go auto-merged) + the test-only fix-up (two S6 test files planting stub providers); SPEC_PIN dced9b8; closed status-row order stated and pinned; full suite green at the rc.12 root on the hosted gate; independently reviewed; integrated by worktree integrate.
