# TASK-260825-3n4bjj: https-broker-and-fetch-wiring

## Description
Materialize the manager credential broker for one pinned HTTPS fetch and wire it into the fetch: a manager-owned wrapper next to the SSH wrapper, a manager-owned state file carrying the pinned host and username and never a secret, GIT_ASKPASS and core.askPass both pointed at the wrapper (core.askPass overrides the environment variable, so both must be set), and the resolved secret passed only in the fetch children's environment. The broker answers exactly the two prompts Git asks and only for the pinned host; a foreign host, a foreign prompt, unreadable state, or absent material exits without printing a byte. Keep the fetch hardening as it is: pinned URL, TLS verified, redirects disabled.

## Scope
(define task scope)

## Acceptance Criteria
Private HTTPS fetch authenticates end to end against a real repository; foreign host and prompt and state cases fail closed with tests; anonymous HTTPS remains byte-identical to today when no credential is selected; secret excluded from any diagnostic representation; go test green.
