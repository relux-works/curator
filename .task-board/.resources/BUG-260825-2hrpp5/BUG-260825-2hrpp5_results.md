# BUG-260825-2hrpp5 implementation evidence

## Result

Windows no longer hard-links the temporary HTTPS askpass wrapper to the running
manager executable. `copyBrokerExecutable` uses an independent byte copy on
Windows and returns only after the input and output handles have closed. Unix
retains the hard-link fast path.

Both affected tests now call the production materializer and explicitly remove
the wrapper before `t.TempDir` cleanup. On Windows, the prior implementation
fails this assertion because the wrapper shares the running executable's locked
file identity; the new copy lifecycle makes it removable.

## Changed scope

- `internal/buildrepo/httpsbroker.go`
- `internal/buildrepo/httpsbroker_test.go`
- `LOGBOOK.md`

The worktree contains pre-existing composite-story changes outside this bug;
they were preserved and not rewritten for this task.

## Validation

| Command | Exit | Result |
| --- | ---: | --- |
| `go test ./internal/buildrepo -run '^(TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts|TestHTTPSBrokerStateContainsHostAndUsernameOnly)$' -count=1` | 0 | Both regression tests pass on macOS. |
| `go test -race ./internal/buildrepo -run '^(TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts|TestHTTPSBrokerStateContainsHostAndUsernameOnly)$' -count=1` | 0 | Targeted race run passes. |
| `go test ./internal/buildrepo -count=1` | 0 | Package suite passes. |
| `go test ./... -count=1` | 0 | Full Go suite passes; slowest package was `cmd/curator` at 339.812s. |
| `GOOS=windows GOARCH=amd64 go test -c ./internal/buildrepo -o .temp/BUG-260825-2hrpp5-buildrepo-windows.test.exe` | 0 | Windows test package compiles. |
| `GOOS=linux GOARCH=amd64 go test -c ./internal/buildrepo -o .temp/BUG-260825-2hrpp5-buildrepo-linux.test` | 0 | Linux test package compiles. |
| `gofmt -l internal/buildrepo/httpsbroker.go internal/buildrepo/httpsbroker_test.go` | 0 | No unformatted files. |
| `go vet ./internal/buildrepo` | 0 | Vet clean. |
| `golangci-lint run ./internal/buildrepo` | 0 | Lint clean. |
| `go build -o .temp/BUG-260825-2hrpp5-curator ./cmd/curator` | 0 | CLI builds on macOS. |

Native Windows test execution was not available on the macOS worker. The real
Windows unlink behavior is exercised by the two regression tests on
`windows-latest`; cross-compilation proves that the platform-specific path and
tests compile.
