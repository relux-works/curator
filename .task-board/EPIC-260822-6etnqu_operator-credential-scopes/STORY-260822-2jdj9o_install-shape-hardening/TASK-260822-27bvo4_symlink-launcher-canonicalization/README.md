# TASK-260822-27bvo4: symlink-launcher-canonicalization

## Description
Verify the manager executable identity path when the binary is started through a package-manager symlink (e.g. a shim directory on PATH): per profiles/manager.md the manager resolves its own executable to a canonical regular installed file and rejects symlink/reparse/hard-link substitution - resolution comes first, rejection targets substitution, not the operator's launch shape. Inspect os.Executable() usage in internal/godriver (build.go) and any single-link/canonical checks downstream. Write a test that invokes identity verification with a symlinked executable path; fix by canonicalizing before checks if it fails.

## Scope
(define task scope)

## Acceptance Criteria
Test proves symlinked launch resolves (or documents with evidence that the current path already canonicalizes); substitution controls still fail closed; go test green.
