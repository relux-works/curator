# BUG-260729-r0fe02 — lint compatibility replay

Date: 2026-07-29
Role: developer
Worktree: `.temp/BUG-260729-r0fe02/worktree`

## Resume scope

This replay applied the exact accepted `BUG-260729-1o0m8f` lint patch on top of
the preserved cancellation worktree. It did not alter the cancellation
implementation or its tests, and it reused the already attached repeated-race
evidence as required by the combined-evidence precondition.

Accepted lint patch:

`BUG-260729-1o0m8f_lint-fix.patch`

SHA-256:

`8a07c0b239548235aea7dfa05fdb1d1cb2926971d4444d3435a9e6f8da368062`

`git apply --check` before application and `git apply --reverse --check` after
application both exited 0. A byte-for-byte `cmp` of all five lint-patch files
against the accepted `BUG-260729-1o0m8f` worktree exited 0.

## Cancellation patch integrity

The cancellation files had these SHA-256 values before the lint patch and have
the same values after all checks:

- `internal/godriver/fingerprint.go`:
  `d53cb4f4212777496da7cbf80826366f5274406c0c4bb083b1c0800b9a822d23`
- `internal/godriver/fingerprint_equivalence_test.go`:
  `fb2f286b39f08104733fba20777366ec1456287457e39ce457cb94620e9644b2`

The already attached task patch is also unchanged:

- `BUG-260729-r0fe02_patch.diff`:
  `462f1ff0326f74540eeb2815cc80542c55f47b35c6b1baef17b80b8815709c28`

## Validation ledger

Every gate was run directly as a standalone process. Exit codes below are the
real process exits.

| Gate | Exit | Outcome |
| --- | ---: | --- |
| `.temp/TASK-260720-1pvfj5/bin/golangci-lint version` | 0 | v2.12.2 built with Go 1.25.5 |
| `.temp/TASK-260720-1pvfj5/bin/golangci-lint cache clean` | 0 | cache cleared before the measured lint run |
| `.temp/TASK-260720-1pvfj5/bin/golangci-lint run` | 0 | `0 issues.` |
| `gofmt -l` on the two cancellation files and five lint-patch files | 0 | no output |
| `go test ./internal/godriver/ -run 'TestFingerprintCancellationStaysFailClosed\|TestDigestCopyDiagnosticPrecedence' -count=1 -v` | 0 | deterministic cancellation boundary and mutation-negative controls pass |
| `go test ./internal/protocoljson ./internal/transaction -run 'TestMarshalCanonicalEscapesEveryControlCharacter\|TestRequireCanonicalAcceptsEveryControlCharacterEscape\|TestValidateRemovalEntries' -count=1 -v` | 0 | accepted lint-fix behavior tests pass |
| `go vet ./internal/godriver/ ./internal/protocoljson/ ./internal/transaction/` | 0 | scoped packages compile and vet cleanly |
| byte-for-byte `cmp` of all five applied lint-patch files against the accepted lint worktree | 0 | exact accepted patch content present |
| `git apply --reverse --check BUG-260729-1o0m8f_lint-fix.patch` | 0 | accepted patch is exactly reversible from the combined candidate |

The host has no unqualified `golangci-lint` on `PATH`; the measured gate used
the preserved CI-pinned v2.12.2 binary from the accepted composite evidence.

## Reused race evidence

No race gate was rerun. The existing task outcome
`BUG-260729-r0fe02_gatelogs.tar.gz` retains the bounded repeated `-race` results,
including the count-1000 cancellation/precedence gate and the focused family
race gate, both with exit 0. The combined-evidence instruction explicitly
requires reusing that evidence to avoid overlapping the parent task's heavy
race gate.

## Not run

- `go test ./...`
- any full Curator test suite
- any new `-race` command

These were intentionally omitted by the task's combined-evidence precondition.

## Outcome

The inherited lint gate is green on the combined candidate. The accepted lint
patch is present byte-for-byte, the cancellation patch content is unchanged,
and the narrow non-race compatibility checks all pass. The developer handoff is
ready for review.
