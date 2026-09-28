# TASK-260924-mcmova: define-hard-link-substitution-for-platform-system-executables

## Description
Erratum to protocol/core.md (enforced script commands, exec capability and interpreter identity rules) and profiles/manager.md: state precisely what hard-link substitution means, and that a platform-owned system executable resolved from the manager's own captured SystemRoot (canonical %SystemRoot%\System32, default search list only, target physically below it) whose extra links are the platform component store is not a substitution; everything else multiply-linked stays rejected. Add or adjust the conformance vector/case if the rule is vectorised; CHANGELOG.

## Scope
(define task scope)

## Acceptance Criteria
normative text states the rule and its bounds; no widening beyond System32 under the manager-captured SystemRoot; interpreter rule unchanged for python/node; make validate (bounded parts) and regenerate-check green
