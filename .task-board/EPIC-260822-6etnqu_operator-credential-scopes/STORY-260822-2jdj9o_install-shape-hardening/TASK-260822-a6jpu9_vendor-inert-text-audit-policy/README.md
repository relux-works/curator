# TASK-260822-a6jpu9: vendor-inert-text-audit-policy

## Description
Determine how the external repository static audit treats non-executable text below vendor/ of third-party modules (e.g. a vendored Makefile carrying a 'curl ... | sh' line). The fixed build session runs only go list and go build with hooks/generators denied, so such text never executes. If the audit blocks on it, demote non-executable vendor text to advisory while keeping executable vendor files and every critical finding blocking; if no such audit exists, record the gap and its spec basis in task notes for a follow-up decision.

## Scope
(define task scope)

## Acceptance Criteria
Reproducing fixture with a vendored third-party Makefile; install admits it while an executable vendor script and first-party text still block - or an evidence-backed note that the audit surface does not exist yet; go test green.
