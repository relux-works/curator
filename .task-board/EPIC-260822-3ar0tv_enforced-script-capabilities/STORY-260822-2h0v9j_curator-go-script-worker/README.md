# STORY-260822-2h0v9j: curator-go-script-worker

## Description
curator (Go): implement script-worker-v1 — extend the existing manager-owned worker re-execution mode to script commands, derive containment from declared capabilities (deny-by-default), mandatory portable controls + native-control inventory probing with capability-evidence records, script_execution_control_unavailable preflight, audit warning class for declared-only legacy script commands.

## Scope
(define story scope)

## Acceptance Criteria
Conformance suite for script-worker-v1 green on ubuntu/macos/windows lanes per the platform-case ledger; gates and evidence wired like the go-v1 ones.
