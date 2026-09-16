# BUG-260916-3aco9f: profile-install-reinstall-drops-use-and-takeover

## Description
https://github.com/relux-works/curator/issues/73 — profile install on an already-recorded source takes the reinstall delegation of installLocked (prior == source), which never reads options.Use or policy.Takeover and returns before resyncCurrentScopes when the lock hash is unchanged; the retry after a takeover stop (foreign-manager symlink or unmanaged conflict) is a silent no-op that exits 0 with updated profile <name>. Reviewer evidence: TASK-260906-1uf713_review-findings-stage-c-5.md; listed by the launcher/migration campaign report as a trunk defect reproducible on main.

## Scope
internal/envprofile/envprofile.go installLocked / reinstallPathLocked and the git reinstall delegation; conformance vector for stop-then-retry

## Acceptance Criteria
--use and --takeover are honoured on every install shape including a reinstall of a recorded source; an unchanged lock still switches under --use; the retry after a takeover stop performs the takeover with notice and backup exactly like profile use --takeover; a conformance vector covers the stop-then-retry sequence; issue #73 closed with the landed commit
