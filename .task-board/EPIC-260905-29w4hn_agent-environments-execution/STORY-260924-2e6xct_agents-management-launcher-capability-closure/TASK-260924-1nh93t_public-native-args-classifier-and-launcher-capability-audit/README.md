# TASK-260924-1nh93t: public-native-args-classifier-and-launcher-capability-audit

## Description
(1) Audit launcher SPEC 0.5.0-draft section 4 (curator-agent-launcher Story STORY-260922-39hxog branch) against the module public API and list EVERY capability the launcher must obtain from agents-management; (2) expose each missing one publicly, first a versioned classifier: given system, verified tool release and native-argument suffix, does it select a non-interactive form (reusing internal/nativeargs, same grammar version as the permission grammar); (3) release v0.5.22.

## Scope
(define task scope)

## Acceptance Criteria
audit table in results (SPEC clause, module API that satisfies it, status); public classifier with rows per environment x non-interactive form x placement, unknown/unverified release fails closed; no provider spelling outside plugins; mutants killed; go test/vet exit 0; README + one CHANGELOG bullet
