# TASK-260823-3c27d3: fix-windows-staged-script-executable-binding

## Description
Windows candidate lane: internal/install TestDryRunEffectBindingsSeeWhatARealOperationWrites — a real project install rejects staged script command clonable-tool as not executable. Windows has no executable bit; determine what the staging/binding layer should treat as executable there per the spec, fix with tests, land via PR green CI. Evidence in .temp/TASK-260823-1l1p8q/windows-evidence-32638424105/. Fully autonomous per the 2026-08-22 pre-authorization.

## Scope
(define task scope)

## Acceptance Criteria
The dry-run effect-binding candidate case passes on windows; merged to main with green CI.
