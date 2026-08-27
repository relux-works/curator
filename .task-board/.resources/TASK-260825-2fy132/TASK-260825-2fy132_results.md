## HTTPS credential documentation

Updated README.md, CHANGELOG.md, LOGBOOK.md, and added docs/build-https.md. The page documents build_https grammar, token sources, precedence, prefetch resolution, anonymous fallback, Git credential and askpass mechanics, required .git suffix, and the Spec core §12.2 identity-unbound override warning.

## Evidence

- Primary delivery checkout: go run ./cmd/curator config build-https --help exit 0.
- Primary delivery checkout: add and replacement transcripts exit 0 and match docs.
- Worktree: go build ./... exit 0; local Markdown link check exit 0.
- Primary checkout: make lint exit 0.
- Worktree make lint exit 2 only because the isolated checkout lacks the untracked replacement agents/skills/skill-go-testing-tools/tuitestkit; this does not involve changed documentation.

## Boundary

The delivered HTTPS path resolves explicitly selected sources before the first fetch. An uncovered repository remains anonymous; it does not discover or prompt for a token during install.