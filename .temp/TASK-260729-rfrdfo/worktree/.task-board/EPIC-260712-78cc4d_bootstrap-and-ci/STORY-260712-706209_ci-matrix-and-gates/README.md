# STORY-260712-706209: ci-matrix-and-gates

## Description
GitHub Actions workflow: test job on ubuntu, macos, and windows (go test ./..., gofmt diff check, go vet), a lint job with golangci-lint, and a naming-gate job that greps the repository case-insensitively for the forbidden brand and fails on a match.

## Acceptance Criteria
- CI green on all three OS
- golangci-lint runs with a committed config
- Naming gate fails the build when the brand is introduced
