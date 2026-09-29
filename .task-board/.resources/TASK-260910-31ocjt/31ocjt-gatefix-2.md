# TASK-260910-31ocjt — gate fix 2 (THE ONLY CURRENT INSTRUCTION, with 31ocjt-brief.md)

Rev3 (tree 70a33e5f) compiles now, but internal/crossconformance TestIntegrationSurfaceStartsNoProcessOutsideTheSharedSeams fails on
macOS and both race lanes (run 36492403893): `guard_test.go:50: draftsources_broker_e2e_test.go bypasses the instrumented process boundary`.
Your updated e2e test starts a process directly (exec.Command or similar) instead of going through the package's shared, instrumented process
seam that guard_test.go enforces. Route that process start through the same seam the other crossconformance tests use (read guard_test.go
for what it accepts); do not add an exemption to the guard. Keep the test's assertions (broker receives the secret via the pipe/fd; the
environment does not carry it). Run `go test ./internal/crossconformance` and `go test -race ./internal/crossconformance ./internal/buildrepo`
with real exit codes. Set status development; update results; handoff; END YOUR TURN. No CHANGELOG/LOGBOOK edit.
