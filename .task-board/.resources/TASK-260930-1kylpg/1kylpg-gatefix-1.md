# TASK-260930-1kylpg — gate fix 1 (THE ONLY CURRENT INSTRUCTION, with r1rows-brief.md)

Rev1 failed only on Windows (run 36726325163):
- internal/pathboundary TestNamedRoutesRejectMissingTarget (named_absence_test.go:31) got
  `permissions: … boundary check failed at <t.TempDir()>\001: DACL grants mutation rights to another identity` instead of the missing-
  target *Failure. On Windows, an ordinary t.TempDir() is not owner-only, so the enclosing-root DACL check fires before the absence
  check.
- Fix the FIXTURE, not the check. Create the root through the product's private-directory helper (internal/privatedir, or whatever
  the existing pathboundary/envprofile Windows fixtures use; grep for them) so that it carries the owner-only protected DACL. Then the
  missing named target is the first failure on every OS. Do not skip on Windows, and do not relax the assertion.
- (TestPrivateHTTPSBrokerAuthenticatesRealGitRepository also failed in that run. It is a known intermittent Windows flake in
  internal/buildrepo and is tracked separately. Ignore it.)

Run `GOOS=windows go vet ./internal/pathboundary ./internal/envprofile`, `go test ./internal/pathboundary -run NamedRoutes`, and the
envprofile row, with the real exit codes. Keep the mutant evidence. Set status development, update the results, run
`task-board handoff TASK-260930-1kylpg --role developer`, then END YOUR TURN.
