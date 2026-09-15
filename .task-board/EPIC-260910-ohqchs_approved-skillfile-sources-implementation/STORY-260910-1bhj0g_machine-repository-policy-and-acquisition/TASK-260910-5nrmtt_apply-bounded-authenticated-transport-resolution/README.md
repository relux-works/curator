# TASK-260910-5nrmtt: apply-bounded-authenticated-transport-resolution

## Description
Apply bounded authenticated transport resolution. Extend current implementation after inspecting existing outcomes; this task is not authorization to start work.

## Scope
internal/gitops, buildrepo, buildsource, gitcred; existing Git acquisition lanes.

## Acceptance Criteria
Use existing SSH wrapper/HTTPS credential broker and lane grammar; at most two attempts within total deadline; fallback only on positively classified availability/auth failures. TLS, host-key, ref, identity, integrity, audit, unknown and ambiguous 404 fail closed. Verify locked content and sanitized errors; no ambient unsafe Git config or secret persistence.
