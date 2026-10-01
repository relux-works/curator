# TASK-260728-20ao7p results

## Changes (uncommitted in story worktree)
- cmd/curator/native_blackbox_test.go: TestNativeBlackboxExternalBuildLifecycle. Builds the real binary once (testcli.Binary), local git tool repo via Skillfile.dev.json build_repository_substitutions (no network), schema-7 skill with build_repositories + locked_commit. Steps: install (outcome=would-preflight-and-build, artifact+receipt in build root, shim + skill published) -> shim run prints `blackbox-tool-ok` exit 0 -> reinstall (outcome=cache-hit, same key, artifact mtime and receipt unchanged) -> remove + reconcile (no build row; shim, skill, build root absent). Windows shim `.cmd`. No skips.
- docs/external-build-repositories.md: author guide (build_repositories in schema 7, locked commit + tag, local dev substitution, admission -> audit -> cache -> compiler), links spec guide.
- README.md: link to the guide.

## Validation
- `go vet ./cmd/curator`: exit 0; gofmt clean.
- `go test ./cmd/curator -run NativeBlackbox -count=1 -v`: PASS (195.5s test, ok 374s). Run piped through tail, so exit status taken from the `ok`/PASS line.
- Windows/macOS hosted lanes: not run locally; hosted gate is the arbiter. ssh relux / ssh win rc.5 qualification not run (out of this brief's scope).

## Changelog entry text
- Added a native lifecycle black-box test (install, shim run, cache-hit reinstall, remove) for external build repositories, and an author guide for `build_repositories`.
