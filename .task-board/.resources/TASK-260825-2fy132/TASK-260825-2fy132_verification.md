## HTTPS credential documentation verification

### Files prepared
- README.md links the operator credential overview to docs/build-https.md.
- docs/build-https.md documents build_https grammar, sources, command output, precheck behavior, precedence, fetch broker, the `.git` redirect constraint, and the identity-unbound exposure warning.
- CHANGELOG.md records the operator documentation and exposure warning.
- LOGBOOK.md records the boundary and verification.

### Verification
- `go run ./cmd/curator config build-https --help` against the primary delivery checkout: exit 0. The command surface, precedence, warning, and `.git` guidance agree with the page.
- `CURATOR_CONFIG=<task temp config> go run ./cmd/curator config build-https add git.example.com/portals --token-env PORTALS_TOKEN --username oauth2`: exit 0, output exactly matches the documented added transcript.
- The same command with `--keyring`: exit 0, output exactly matches the documented replacement transcript.
- `make lint` against the primary delivery checkout: exit 0, `golangci-lint run` reported `0 issues.`
- Local README documentation targets exist: exit 0.

### Checkout note
The assigned Story worktree lacks the uncommitted implementation that exists in the primary delivery checkout, so the CLI was verified there. Documentation changes remain in the assigned Story worktree.