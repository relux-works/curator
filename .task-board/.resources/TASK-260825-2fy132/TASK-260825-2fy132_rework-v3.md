## HTTPS credential documentation rework

### Correction

The operator page now matches the shipped resolution path in the primary delivery checkout. `internal/install/resolveBuildHTTPS` is reached from `internal/install/external.go` before external fetches. It resolves only an explicit `CURATOR_BUILD_HTTPS_TOKEN` override, then the longest matching `build_https` scope, then anonymous HTTPS. No `InteractiveBuildHTTPSResolver`, `operatorBuildHTTPSResolver`, or HTTPS candidate-prompt source exists in the project. The page therefore documents the absence of candidate discovery and terminal prompting instead of an SSH-like interaction.

The config grammar, token sources, command syntax and output examples, precheck of explicitly selected unavailable sources, `.git`/301 redirect rule, manager-owned host-pinned askpass broker, precedence, and the `Spec core §12.2` identity-unbound exposure warning remain documented.

### Validation

- `go run ./cmd/curator config build-https --help` in the primary delivery checkout: exit 0; command syntax and precedence/disclosure surface inspected.
- `go test ./cmd/curator -run 'TestConfig.*BuildHTTPS|Test.*BuildHTTPS' -count=1` in the primary delivery checkout: exit 0.
- `go test ./internal/config ./internal/install ./cmd/curator -run 'Test(BuildHTTPS|ConfigBuildHTTPS)' -count=1` in the primary delivery checkout: exit 0. This includes grammar rejection, longest-scope precedence, host pinning, unavailable selected-source failure, anonymous fallback, and command-output coverage.
- `make lint` in the story worktree: exit 0 (`golangci-lint run`, 0 issues).
- `git diff --check` in the story worktree: exit 0.
- The Curator Specification URL cited by the page returned HTTP 200.

No repository documentation-link checker is configured in the Makefile or CI scripts; all local Markdown links introduced by this change resolve to existing repository files.
