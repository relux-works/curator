# TASK-260916-3gcc00: reconcile-script-worker-v1-delivery

## Description
Research/reconciliation: establish with evidence whether the Curator Go script-worker-v1 scope of this Story (manager-owned worker re-execution for script commands, deny-by-default capability containment, mandatory portable controls plus native-control inventory probing with capability-evidence records, script_execution_control_unavailable preflight, audit warning class for declared-only legacy script commands) is delivered on curator main 559447ef. Map every scope clause to the landed code (internal/scriptpolicy, internal/skillspec, internal/skillcheck, cmd/curator, conformance corpus), the landing PRs (#33 Admit schema-8 module roots and the script execution surface, #34 Refuse an enforced script command instead of installing it uncontained, and later), the schema-8 conformance vectors of curator-spec rc.9, and the CI platform-case gate. For each clause: delivered (file, test, vector), partially delivered (exact gap), or missing. Produce TASK-<id>_reconciliation.md with the table and a recommendation: close the Story as delivered, or the exact residual tasks to create. No code changes.

## Scope
(define task scope)

## Acceptance Criteria
Every scope clause of STORY-260822-2h0v9j mapped to landed evidence or a named gap; recommendation stated; no code changes.
