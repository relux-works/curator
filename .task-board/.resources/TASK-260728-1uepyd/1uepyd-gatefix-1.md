# TASK-260728-1uepyd — Windows gate fix 1 (THE ONLY CURRENT INSTRUCTION, with 1uepyd-brief.md)

Rev1 (tree a149aaa7) is green everywhere except Test (windows-latest) (run 36765274974):
`TestExternalRepositoryAcquisitionConformance/cases: failing published case external-repository/acquisition/cases/sha1-untagged-https
… build_repository_identity_invalid: trusted Git version probe failed`.
Your new consumer sets up the trusted Git tool in a way that fails on Windows. The existing internal/buildrepo tests that exercise
AcquireNetwork / ValidateGitTool pass on Windows, so reuse exactly their fixture setup: the trusted git path, the .exe handling, PATHEXT
and the wrapper image. Grep for how TestTaggedAcquisitionUsesOnlyExactTagAndChecksTerminalCommit and TestNetworkAndLocalSHA1SHA256RawObjectParity
obtain their Git tool.
- Fix the fixture, not the probe.
- Do not skip on Windows.
- Do not add a gap row for a fixture problem.
- If a case is truly platform-inapplicable per the vector, use the platform-case ledger with an exact reason.

Run `GOOS=windows go vet ./internal/buildrepo` and `go test ./internal/buildrepo -run ExternalRepositoryAcquisition`, with real exit
codes. Keep the mutants. Set status development, update the results, run `task-board handoff TASK-260728-1uepyd --role developer`,
then END YOUR TURN.
