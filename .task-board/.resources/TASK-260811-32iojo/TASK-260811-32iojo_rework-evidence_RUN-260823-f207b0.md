# TASK-260811-32iojo rework evidence

Status at handoff: ready for review.

## Reviewer findings resolved

1. PnP invocation now binds `--require ./.pnp.cjs` in the immutable C5 action template and the concrete protected-executor permit. The node-modules linker retains plain Node invocation. The behaviorally faithful executor test rejects a PnP invocation that omits the loader, and the real Yarn fixture resolves `is-number` through the regenerated loader.
2. The adapter now implements the exact condition strings emitted by pinned Yarn 4.9.2 `Manifest.getConditions`: `os`, `cpu`, and `libc` groups joined by ` & `; parenthesized alternatives joined by ` | `; and negated selectors prefixed with `!`. The same evaluator is used during lock selection and C4 projection. Selected and pruned evidence is tested.
3. Modern-Yarn wrapper tests now explicitly cover S01-S08 and N01-N13, including tamper/missing archive, network-capable lifecycle, compiled direct/renamed/nested payload, missing graph edge, generated JS/source map, workspace escape/drift, plugin/Git/patch, ambient cache, PnP state, and both supported linkers.
4. Behavior-affecting `.yarnrc.yml` values are closed. Telemetry and the experimental PnP ESM loader must be disabled; `defaultProtocol` must be `npm:`; and `npmRegistryServer` must be the pinned default. Their normalized effective values are included in `ConfigurationDigest`; explicit defaults and pinned-release defaults canonicalize identically.
5. Declared patch bytes, path, locator, and SHA-256 are covered by a positive identity test; patch-byte drift changes closure identity.

## Validation

Each gate below ran as a standalone process. Exit codes are literal.

| Command | Exit | Result |
| --- | ---: | --- |
| `go test -count=1 ./internal/yarnmodernsource` | 0 | focused suite passed, final run `0.768s` |
| `go test -count=1 -race ./internal/yarnmodernsource` | 0 | race suite passed, final run `3.211s` |
| `go test -count=1 -coverprofile=.temp/TASK-260811-32iojo/coverage-rework.out ./internal/yarnmodernsource` | 0 | 72.4% statement coverage |
| Real Yarn 4.9.2 `install --immutable --immutable-cache --mode=skip-build` with network/scripts/global cache disabled and preseeded PnP/install state removed | 0 | regenerated offline PnP state in `24ms` |
| `node --require ./.pnp.cjs -e "process.stdout.write(String(require('is-number')('42')))"` | 0 | output `true` |
| `golangci-lint run` | 0 | `0 issues.` |
| `go vet ./...` | 0 | passed after the final source edits |
| `go build ./...` | 0 | passed after the final source edits |
| `go test -count=1 ./...` | 0 | repository-wide uncached suite passed; `cmd/curator` 391.942s, `internal/yarnmodernsource` 11.394s |
| `git diff --check` | 0 | passed |
| `gofmt -l internal/yarnmodernsource` | 0 | empty output |
| `task-board validate` | 0 | `Board is valid. No issues found.` |

The first post-rework `golangci-lint run` correctly exited 1 with four local findings (bounds analysis, deprecated test API, capitalized error, and an unused helper). Those findings were fixed; the standalone rerun above exited 0.

## Source identity

| File | SHA-256 |
| --- | --- |
| `capture.go` | `cc9c16d47a06c9384198e8a863d1223778046f27f9cd94198b52408196875549` |
| `conformance_test.go` | `2475ffcf7c9099ac2976371defe753d4e83e731f7129d396906ff77c05ecd3b1` |
| `errors.go` | `61fc044d9291133c31146cd550512775024bd14ef2cbe3123b66262fc0137302` |
| `lock.go` | `61b65ddaf36014e3e795374ae77300416c1b0a663005db031b2374af502c5967` |
| `materialize.go` | `e2c5c599d8d9476080f4001d3886300c1810289820184278c2c902edb7701477` |

No files were staged or committed. Existing unrelated dirty-worktree changes were preserved.
