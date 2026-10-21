# TASK-261018-adogt7: r4 handoff admission blocker

## Stop-The-Line
### Blocker
The current r4 producer contract defers full exact-candidate validation to configured post-handoff execution, but the live handoff admission requires checklist rows 6, 18 and 22 already checked. Row 6 explicitly requires the full local suite and hosted gate green on the exact candidate; row 22 depends on a later review verdict. This producer cannot truthfully attest those future outcomes. The scope permits no source/test/docs changes and does not authorize weakening a shared guard or changing the delivery contract.
### Evidence
`task-board handoff TASK-261018-adogt7 --role developer` ran against the authoritative board with the configured agent Git environment and exited **1**:

```text
warning: could not compare board content for run RUN-261018-7aa5d2; no progress recorded for this command
cannot hand off TASK-261018-adogt7: unchecked checklist items [6 18 22] (Full local suite and hosted gate green on the exact candidate; Implementation matches AC; If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches): handoff evidence missing
```

This refusal precedes role handoff: no new CR revision or green hosted gate is claimed. The snapshot remains uncommitted at checkpoint `892dab098aef1497df9ccacdb173faaf4f8b8418`, tree `4cf404c658fdb199f5da12568ad7dbfe7a351239`. Seven direct standalone affected-check/build/vet/lint/invariant commands exit 0. All 56 schema, 24 lifecycle, 2 identity, 6 metadata, 4 agreement controls and ten narrowing mutants pass their intended harness assertions. Fresh evidence and gaps were attached before the refused handoff. No guard, checklist truth value, source, test or documentation was changed to bypass the refusal. Refresh would not fix these explicitly named checklist prerequisites and was not run.
### Options
Recommended: the orchestrator resolves the phase mismatch through the supported board workflow, assigning the post-handoff full-gate and review obligations to the phase that actually owns them while retaining their acceptance requirements, then retries producer handoff on this unchanged tree. This preserves the r4 contract and truthful evidence. Alternative: explicitly revise the producer brief to require pre-handoff full local/hosted validation and define how the future review-dependent row is satisfied; that restores a different ordering at the cost of a full validation round. Marking future checks green, direct `set_status(to-review)`, deleting admission guards, or modifying shared tooling from this constrained product run are not valid alternatives.
### EXACT INPUT NEEDED
An orchestrator-owned resolution of live checklist rows 6, 18 and 22 versus the r4 post-handoff validation contract, using supported CLI operations and retaining actual full-gate/review acceptance obligations. Then rerun the normal developer handoff on this retained uncommitted candidate. No product or source decision is needed; local checks are green.
