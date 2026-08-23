# TASK-260811-32iojo developer rework evidence

Status at handoff: ready for review.

Run: `RUN-260823-5141d9`

## Review findings resolved

1. `CaptureAndAdmit` now discovers every regular file below `.yarn/patches`
   and requires an exact bijection with lock-declared, byte-bound patch inputs.
   An undeclared captured patch returns
   `closure_manager_plugin_undeclared` before any manager process can start.
2. Required peer descriptors must resolve during graph construction. An
   unresolved optional peer remains as an explicit unselected edge with the
   deterministic reason `optional_peer_unresolved`; it is not silently omitted.
3. PnP validation now parses the Yarn 4.9.2 `RAW_RUNTIME_STATE` and reconciles
   the exact selected package locators, pruned-package absence, dependency map,
   workspace locations, and private-cache ZIP locations. A nonfunctional
   `module.exports = {}` loader fails materialization and cannot publish.
4. The positive PnP test stages exact Node bytes and the pinned
   `@yarnpkg/cli-dist` 4.9.2 entry point, runs immutable/immutable-cache/
   skip-build installation through the verified executor contract under
   `sandbox-exec` with `(deny network*)`, and then invokes a dependency-reading
   `index.js` through `Invoke` with `--require ./.pnp.cjs`.

## Additional real-tool findings

- Yarn 4.9.2 rejects the legacy `YARN_IGNORE_SCRIPTS` environment setting.
  The profile now uses the supported `YARN_ENABLE_SCRIPTS=0` binding while the
  root rc also binds `enableScripts: false`.
- Immutable replay requires the exact Yarn cache filename, not an arbitrary
  safe `.zip` basename. Cache names now bind the Yarn locator SHA-512 slug and
  the first ten checksum hex characters; supplied normalized archives must use
  that exact name.
- Yarn immutable install requires its generated `languageName` and `linkType`
  lock identity. The closed lock parser now requires `unknown`/`soft` for
  workspaces and `node`/`hard` for supported package entries and includes both
  fields in canonical lock identity.

## Validation

Every gate below ran as a standalone process. Exit codes are literal.

| Command | Exit | Evidence |
| --- | ---: | --- |
| `node .../@yarnpkg/cli-dist/bin/yarn.js --version` | 0 | `4.9.2` |
| `CURATOR_TEST_YARN_MODERN_JS=... go test -count=1 ./internal/yarnmodernsource` | 0 | focused suite including real Yarn/Node vector passed in `2.767s` |
| `CURATOR_TEST_YARN_MODERN_JS=... go test -count=1 -race ./internal/yarnmodernsource` | 0 | race suite passed in `5.820s` |
| `CURATOR_TEST_YARN_MODERN_JS=... go test -count=1 -coverprofile=.temp/TASK-260811-32iojo/coverage-rework-3.out ./internal/yarnmodernsource` | 0 | `73.1%` statement coverage |
| `gofmt -l internal/yarnmodernsource` | 0 | empty output |
| `golangci-lint run` | 0 | `0 issues.` |
| `go vet ./...` | 0 | passed |
| `go build ./...` | 0 | passed |
| `go test -count=1 ./...` | 0 | all packages passed; `cmd/curator` `410.430s`, `internal/yarnmodernsource` `8.140s` |
| `git diff --check` | 0 | passed |

Development red runs were retained honestly:

| Command / iteration | Exit | Cause and disposition |
| --- | ---: | --- |
| focused `go test` after integration scaffolding | 1 | unused `sort` import; removed |
| real protected integration, first run | 1 | Yarn rejected legacy `ignoreScripts`; replaced with `enableScripts` |
| real protected integration, second run | 1 | fixture lock omitted Yarn-owned language/link fields; parser and fixture closed them |
| real protected integration, third run | 1 | fixture lock formatting was not immutable-canonical; corrected fixture bytes |
| real protected integration, fourth run | 1 | derived cache filename did not match Yarn locator identity; implemented exact slug/checksum naming |
| real protected integration, fifth run | 1 | PnP parser incorrectly treated the single-quoted JS payload as a JSON-escaped string; parser now consumes the line-continuation-stripped JSON payload directly |
| `CURATOR_TEST_YARN_MODERN_JS=... go test -run '^TestN01RealPinnedYarnPnPInvokeThroughVerifiedExecutor$' -v ./internal/yarnmodernsource` | 0 | real immutable install and dependency import passed in `1.86s` |

## Source identities

| File | SHA-256 |
| --- | --- |
| `internal/yarnmodernsource/capture.go` | `4f619467ac7f36bb3b0f7d71af26a2919482b2056fb7af6b503330c0eaf30fb6` |
| `internal/yarnmodernsource/conformance_test.go` | `ec22133bc7f758725fe872d91e9eeca6fe69c6f6ac3366a02d974acd9a0ea3fd` |
| `internal/yarnmodernsource/errors.go` | `61fc044d9291133c31146cd550512775024bd14ef2cbe3123b66262fc0137302` |
| `internal/yarnmodernsource/lock.go` | `78b0f90c3ace872876848356cb397c32795b4cda84e534deda5514ffec7a801d` |
| `internal/yarnmodernsource/materialize.go` | `182f4c70c3dc4258d7cf513a77bca9e1f7de254251d11803eaf30ffc23dda1d6` |

No files were staged or committed. Existing unrelated dirty-worktree changes
were preserved.
