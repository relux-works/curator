# TASK-260825-1tgpcn: https-credential-reads

## Description
Read operator credentials through the operator's own Git credential machinery (git credential fill, approve, reject) rather than any platform-specific secret tool: that is the one mechanism identical on macOS, Windows and Linux, it speaks to whichever helper the operator configured, and it needs no dependency. Disable interactive prompting on every call. Provide: read of the operator's own entry for a host, read and write and delete of a manager-namespaced entry (a distinct username keeps it separate from the operator's own credential for the same host), and a presence-only discovery used by the prompt. A write MUST be verified by reading it back: an operator helper can report success while persisting nothing, and an unverified write would look like a configured scope that fails later mid-install.

## Scope
(define task scope)

## Acceptance Criteria
All reads and writes go through git credential with prompting disabled; the operator home is pinned for the lookup; a helper that persists nothing raises with platform guidance; namespaced entries do not collide with the operator's own; go test green.
