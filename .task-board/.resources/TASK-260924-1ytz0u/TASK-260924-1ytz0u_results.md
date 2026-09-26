# TASK-260924-1ytz0u results — curator-agent-launcher v0.1.0

## Release contents and decisions

- `CHANGELOG.md` now has a `0.1.0` release section containing the accumulated unreleased entries.
- The executable reports launcher version `0.1.0`; `SPEC.md` remains `0.5.0-draft` because the protocol draft version is separate from the launcher release. `TestReleaseVersionPinned` checks these values and the documented install/version references.
- `README.md` tells operators to install `github.com/relux-works/curator-agent-launcher/cmd/curator-run@v0.1.0` and documents the Go install bin path.
- `go.mod` declares module `github.com/relux-works/curator-agent-launcher`, Go 1.25.5, and public module dependencies. There is no `replace` directive or workspace override.
- No GitHub release workflow or prebuilt binaries are needed for the documented onboarding path: operators install the versioned Go module and build it with their Go toolchain. `.github/workflows/ci.yml` already runs build, formatting, vet, tests, and race checks; it has no release or publish step.
- Prerequisite `STORY-260922-39hxog — 0018-launcher-permission-interface` is `done`.
- The orchestrator owns the signed `v0.1.0` tag and its post-landing `@v0.1.0` install verification. This work did not create a tag or commit.

## Changes

Updated `CHANGELOG.md`, `README.md`, `cmd/curator-run/main.go`, `cmd/curator-run/main_test.go`, and `cmd/curator-run/testdata/help.golden`. The test pins launcher and SPEC versions, verifies the README install command and version reference, checks the changelog release header, and exercises the `--version` output path.

## Verification

Commands were run as standalone processes from the Story worktree; no gate was piped through another command.

- PASS, exit 0: `make check`. This ran `go build ./...`, formatting check, `go vet ./...`, `go test ./... -count=1`, and `go test ./... -count=1 -race`; all packages passed.
- PASS, exit 0: fresh-cache local-module install:
  `GOPATH=/tmp/curator-agent-launcher-TASK-260924-1ytz0u-online.gXI29R/gopath GOMODCACHE=/tmp/curator-agent-launcher-TASK-260924-1ytz0u-online.gXI29R/gomodcache GOBIN=/tmp/curator-agent-launcher-TASK-260924-1ytz0u-online.gXI29R/bin GOFLAGS=-mod=mod GOPROXY=https://proxy.golang.org,direct go install ./cmd/curator-run`.
  The isolated GOPATH and module cache started empty. Go fetched `skill-agents-management v0.5.22` and `go-toml/v2 v2.4.3` from the module proxy. Running the installed binary exited 0 and printed `curator-run 0.1.0 (specification 0.5.0-draft)`.
- NOT PASS, exit 1: the prescribed empty-cache offline variant, `GOFLAGS=-mod=mod GOPROXY=off go install ./cmd/curator-run`, cannot resolve either required module from an empty `GOMODCACHE` (`module lookup disabled by GOPROXY=off`). The successful cold-cache install above verifies the local worktree source without local-only dependencies using the configured Go module proxy.
- PASS, exit 0: `git diff --check`.
- Not run: `go install ...@v0.1.0`, because the signed tag is created and verified by the orchestrator after landing. Hosted GitHub CI is part of the handoff/landing path, not a local result claimed here.

## Handoff

The five product/documentation changes remain uncommitted in the assigned Story worktree. This result records the offline-cache limitation and the successful fresh-cache install variant for review.
