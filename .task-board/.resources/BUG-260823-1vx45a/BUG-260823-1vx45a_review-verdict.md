# BUG-260823-1vx45a review verdict — accepted

Reviewed Change Request `CR-BUG-260823-1vx45a-2`, revision 2, candidate tree
`877ac92abee169f5d6dfe7885aeefcc120428a69` against base
`903af23ad0d0fa21328c0a2100e17968bbac6f1e`.

## Verdict

Accepted for the orchestrator's commit/integration handoff. The reviewer does
not supply `commit_ack` and does not move the bug to `done`.

## Findings

- The revision corrects the rejected revision-1 premise. No lock handle was
  shown to survive helper process exit or parent `t.TempDir` cleanup. The
  reproduced cause is the helper protocol: one 200ms context covered setup and,
  for build-key mode, two sequential acquisitions; any deadline expiry was
  serialized as `blocked`, conflating slow uncontended work with contention.
- Every expected-`blocked` call site now uses `200ms`; every expected-`acquired`
  call site uses `30s`, including Windows-only and symlink/canonicalization
  cases. `hold-project` intentionally remains outside the deadline protocol.
- `TestSubprocessExpectedAcquiredWithTinyDeadlineReportsBlocked` drives the real
  helper subprocess against an uncontended lock with a `1ns` deadline and
  requires `blocked`. The separate assertion that acquired deadline is strictly
  wider than blocked deadline makes a collapsed-deadline mutant fail. This is a
  deterministic negative regression for the exact conflation.
- Helper-held locks are now closed synchronously before acquisition-result
  classification. Every non-nil close error reaches `t.Fatalf`; no path can
  publish it as `blocked` or `acquired`.
- The neighbouring elapsed-time assertion in
  `TestProjectOrderInversionFailsBeforeWaiting` can fail on an exceptionally
  slow host even when lock-order rejection is immediate. It is unrelated to
  this incident and should be tracked separately if observed; expanding this
  CR is not justified.

## Reviewer validation

- Exact delta `903af23..877ac92`: `git diff --check` passed; worktree bytes for
  all three changed paths equal candidate tree `877ac92...`.
- Patch SHA-256: `e4dde98d50165cba01a0bf1a32ca2ac1b9097d97ba9941318b609ddaea51f118`,
  matching the Change Request record.
- `go test -count=20 -run '^(TestSubprocessContentionAndIndependentProjects|TestSubprocessBuildKeyDeduplicationAcrossProjects|TestSubprocessExpectedAcquiredWithTinyDeadlineReportsBlocked)$' ./internal/managerlock -v`: pass on macOS arm64.
- `go test -count=5 ./internal/managerlock -v`: pass.
- `go test -race -count=3 -run '^TestSubprocess' ./internal/managerlock -v`: pass.
- `go vet ./...`: pass.
- `golangci-lint run`: pass, 0 issues.
- `go build ./...`: pass.
- Windows/amd64 test-binary cross-build: pass.
- Producer outcome records 20 consecutive native Linux/arm64 and 20 native
  Windows/amd64 subprocess runs on byte-identical files. The recorded file
  SHA-256 values match this candidate:
  `managerlock_test.go` = `ffd2bbffd30842f07960052efed39c52687bf301d9340a8f71b56b85152ae9b2`,
  `identity_windows_test.go` = `0f7a3a263e140d03032d243f80acba29923209be62ef423ed4f4b07242323b1f`.

Windows CI publication remains the orchestrator's integration gate per the
revision handoff. It was not represented as reviewer-run evidence and does not
weaken this implementation verdict.

No repository code was modified by the reviewer.
