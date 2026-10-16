# TASK-261016-2en3yu: implement-rc7-credential-store-policy

## Description
L15-1 slice 1 (scope narrowed by the convener on 2026-10-16): machine-owned credential-store selector in the successor manager-config (curator-spec 15d9701) with its CLI (curator env config set credential_store.<profile>.<env-id> native|file|keychain), typed selector transport into the production resolve path, and fail-closed refusal of every selection that has no qualified application yet. The rest of the L15-1 leaf moves to the follow-up task L15-1b.

## Scope
(define task scope)

## Acceptance Criteria
The slice's grammar, CLI, transport and unsupported-selection refusals are implemented with paired negative tests and positive controls, narrowing controls killed, and the full remote gate green; nothing outside the slice is claimed.
