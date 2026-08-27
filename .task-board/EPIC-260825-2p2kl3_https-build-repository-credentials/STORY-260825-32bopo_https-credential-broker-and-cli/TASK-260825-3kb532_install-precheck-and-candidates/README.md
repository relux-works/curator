# TASK-260825-3kb532: install-precheck-and-candidates

## Description
Before the first fetch, resolve credentials for every declared HTTPS build repository. On an operator terminal an unmatched repository prompts with detected candidates: the operator's existing credential for that host, and entering a token now. Discovery only lists what exists; nothing is used without an explicit selection, and nothing persists without an explicit scope choice whose narrowest option is the repository namespace. A choice marked as this-run-only MUST NOT reach the config; save only the entries the operator chose to persist. Off a terminal the run continues anonymously for HTTPS. Mirror the SSH prompt shape.

## Scope
(define task scope)

## Acceptance Criteria
Prompt default, this-run-only, and abort covered by tests; a this-run-only answer never lands in the saved config; the same latent persistence bug is checked for on the SSH prompt path and fixed if present; go test green.
