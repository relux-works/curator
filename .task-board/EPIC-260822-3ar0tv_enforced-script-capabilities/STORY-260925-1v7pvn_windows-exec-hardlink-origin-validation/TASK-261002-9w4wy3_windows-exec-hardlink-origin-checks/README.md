# TASK-261002-9w4wy3: windows-exec-hardlink-origin-checks

## Description
Production resolver rejects windows-exec-noncomponent-store-hardlinks and windows-exec-unowned-file-hardlinks while preserving the component-store exception and uncaptured-SystemRoot rejection.

## Scope
(define task scope)

## Acceptance Criteria
Both hard-link origin cases rejected at the production resolver; component-store exception and SystemRoot rejection preserved; all 8 family cases counted; ledger rows removed; mutants; Windows lanes green.
