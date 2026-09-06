# TASK-260906-1xitqi: prepare-and-validate-completed-main-release

## Description
Role developer. Audit public Curator channels against completed main 7320bc2 and prepare minimal fixes required for a stable release. Release v0.14.0-rc.1 supports schema 7 only; latest stable v0.13.0. Tags rc.2/rc.3 exist without public release: investigate failed workflows. Scope README Go install (replace directive likely breaks go install @latest), release config/workflow compatibility, missing channel capabilities. Use gh for hosted evidence, inspect CI green main and release failures; correct actual blockers with meaningful tests. Do not merge stage b PR or publish/tag yet: parent orchestrator owns publication after review. Select appropriate next stable version and document concrete release plan and channel verification. Use signed repository commits only; do not bypass signing. Persist evidence and hand off using tracked lifecycle.

## Scope
(define task scope)

## Acceptance Criteria
Minimal release-blocker fixes validated, exact completed-main parity inventory and version proposal documented, successful baseline main CI linked, candidate ready for reviewer; public publishing deferred to parent orchestrator.
