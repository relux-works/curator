# TASK-260728-d8ktna: qualify-rc5-external-repository-release-and-pins

## Description
Run final rc.5 cross-manager qualification, audit released shared-suite and implementation pins, and publish claim-v3 evidence for external build repositories on supported platforms. Keep Linux and unimplemented language drivers out of the claim.

## Scope
Released artifact pin verification, clean shared runner execution, curator-spec/Curator/csk compatibility matrix, security regression audit, macOS/Windows evidence review, claim-v3 and release metadata, rollback plan, and final integrated outcome.

## Acceptance Criteria
Exact released spec, Curator, csk, fixture, Git, Go, Python, macOS, and Windows revisions are pinned and reproducible; all required shared and native gates pass from clean state; independent review confirms source access/audit/build/install/PATH and failure boundaries; schema 1-6 and rc.4 compatibility remains green; claim v3 names only go-repository-v1 on evidenced macOS/Windows targets, excludes Linux until STORY-260728-1eye8p is accepted, and excludes every future language driver until its own reviewed contract and implementation exist; rollback and revocation instructions are complete.
