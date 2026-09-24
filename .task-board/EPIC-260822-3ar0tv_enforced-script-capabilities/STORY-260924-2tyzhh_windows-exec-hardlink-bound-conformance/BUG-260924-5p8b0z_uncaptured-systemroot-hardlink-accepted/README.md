# BUG-260924-5p8b0z: uncaptured-systemroot-hardlink-accepted

## Description
Hosted windows run 35977701956 (candidate TASK-260922-18ex37 rev3 at spec pin dcc7f015): internal/scriptworker TestExecutableIdentityCasesAtProductionEntry case windows-exec-uncaptured-systemroot-hardlinks: production resolver accepted = true, want false (reason systemroot-not-manager-captured; platform_owned=true). The hard-link allowance must apply only when System32 is derived from the manager-captured SYSTEMROOT; otherwise reject.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
the case is driven at the production entry and rejected on windows; other executable_identity_cases unchanged; gap-ledger row removed; mutant (drop the captured-SystemRoot condition) killed on the Windows lane
