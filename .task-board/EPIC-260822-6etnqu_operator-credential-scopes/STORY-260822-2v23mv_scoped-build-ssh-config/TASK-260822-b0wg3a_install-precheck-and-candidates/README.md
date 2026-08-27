# TASK-260822-b0wg3a: install-precheck-and-candidates

## Description
Before any fetch, resolve credentials for every declared ssh build repository. Unresolved + interactive terminal: prompt with detected candidates - live agent socket (key count via ssh-add -l, degrade gracefully) and *.pub files under ~/.ssh - default entry 'agent + pin first key'; persist only after an explicit scope choice (default: the repository namespace). Unresolved + non-interactive: fail closed with build_repository_ssh_credential_missing plus ready-to-run 'curator config build-ssh add' commands built from the same candidates. Discovery only lists material; nothing is used without explicit selection.

## Scope
(define task scope)

## Acceptance Criteria
Prompt flow covered with scripted stdin tests (default selection, abort, manual path); fail-closed message asserts candidate commands; dry-run reports per-repository credential source; go test green.
