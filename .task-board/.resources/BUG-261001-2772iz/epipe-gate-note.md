# Gate note — BUG-261001-2772iz rev1 (orchestrator)

Hosted run 36837352988 failed on Test (ubuntu) and Race (ubuntu). The fix regressed the HAPPY path:
- `internal/buildrepo TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts/password` → httpsbroker_test.go:148: `code=1 output="", want code=0 output="broker-secret\n"`.
- The askpass child no longer receives the secret on Linux. The likely cause: the write is now gated on a condition that a Linux child never satisfies, or the pipe closes too early.

Fix the happy path on every OS:
- The secret must still be delivered when the child asks.
- Keep the refusal-path EPIPE classification.

Run locally before handoff, with real exit codes:
- `go test ./internal/buildrepo -run HTTPSCredentialBroker -count=20`
- the crossconformance askpass rows
