# Tool readiness

- `git --version`: `git version 2.50.1 (Apple Git-155)`
- `go version`: `go version go1.25.5 darwin/arm64`
- `make --version`: `GNU Make 3.81`
- Standalone `golangci-lint`: unavailable; the project’s established reproducible fallback is `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`.
- Target platform: Go module supporting native Darwin plus Linux and Windows compile graphs.

All available tools produced their expected version output before validation.
