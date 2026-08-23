# TASK-260823-qwr5w9: land-credential-scopes-composite

## Description
Assemble the accepted patches of this epic (config field, per-repository resolution, CLI subcommand, precheck/candidates, docs, symlink canonicalization, vendor audit policy, manifest fallback regression, closure provenance, toolchain remedy) into one composite branch off origin/main, resolve overlaps (admission.go and skip-classes.tsv are touched by several patches), run the full gate set, open the PR, and land it after CI. Composite commit messages reference the protocol spec and this repository only.

## Scope
(define task scope)

## Acceptance Criteria
Composite branch applies clean; gofmt/build/vet/golangci-lint/gate-selftest/ledger/full go test green; PR opened and merged with green CI including the interop conformance gate.
