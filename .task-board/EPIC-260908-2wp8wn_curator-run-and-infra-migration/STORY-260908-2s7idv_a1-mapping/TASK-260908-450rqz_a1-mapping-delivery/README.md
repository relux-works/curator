# TASK-260908-450rqz: a1-mapping-delivery

## Description
SPEC section 4.2: closed environment to system and provider mapping.

## Scope
(define task scope)

## Acceptance Criteria
Closed environment mapping uses native pi-native from landed upstream; supported rows return exact system/provider pairs, while opencode/unknown refuse env_unsupported. Production ordering is fragment then mapping before later stages. Existing CLI/fragment behavior stays covered, only evidenced SPEC mapping/version metadata changes, and local checks pass.
