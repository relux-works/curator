# BUG-260729-1o0m8f: resolve-accepted-composite-lint-findings

## Description
The accepted Curator composite is not clean under the CI-pinned golangci-lint v2.12.2: G115 in internal/protocoljson/ccj.go at the guarded rune-to-byte encoding, G602 in internal/transaction/journal.go at a short-circuit previous-entry lookup, and ineffassign in internal/godriver/builddriver_positive_conformance_test.go. Apply semantics-preserving source refactors so lint passes without suppression.

## Scope
Own only internal/protocoljson/ccj.go and focused tests, internal/transaction/journal.go and focused tests, and the single dead test assignment in internal/godriver/builddriver_positive_conformance_test.go. Start from the exact accepted TASK-260720-jrrgw9 candidate. Do not change CI files, .golangci.yml, protocol vectors, timeouts, or unrelated product behavior.

## Acceptance Criteria
golangci-lint v2.12.2 exits 0 with no new exclusion or nolint directive; the rune encoding remains byte-identical for all control characters, journal ordering validation remains fail-closed and panic-free, and the godriver test behavior is unchanged after removing dead code. Focused protocoljson, transaction, and godriver tests plus go vet on those packages pass, and an independent reviewer accepts the patch before TASK-260720-1pvfj5 resumes lint.
