# STORY-260712-ac7c98: go-module-and-cli-skeleton

## Description
Initialize the Go module github.com/relux-works/curator, add cmd/curator with a version subcommand/flag (build-time ldflags with a dev fallback), and a Makefile with build, test, fmt, and lint targets.

## Acceptance Criteria
- go build ./... succeeds
- curator --version prints a version string
- make build/test/fmt/lint work locally
