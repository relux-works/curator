# TASK-260811-32iojo rework evidence

Run: `RUN-260823-b89237`

Reviewer input: `TASK-260811-32iojo_review-verdict_RUN-260823-a470e2.md`

## Rework outcome

- `Parse` now requires a bijection between captured root/workspace manifests and authoritative `workspace:` lock entries.
- Workspace lock identity is closed over exact name, path, Yarn-owned `0.0.0-use.local` version sentinel, canonical selectors, language/link identity, and absence of checksum/conditions.
- Workspace `dependencies`, `devDependencies`, `optionalDependencies`, `peerDependencies`, and optional metadata are reconciled against the admitted manifest before graph construction. Missing, duplicate, unmatched, path/name/version-drifted, selector-drifted, or metadata-drifted entries return `closure_lock_stale` with no graph/config identity.
- `.yarnrc.yml` now uses a single-document typed grammar. Required and optional strings, booleans, and integers reject type confusion; `supportedArchitectures` rejects unknown nested keys, non-string/empty/ambient/duplicate/malformed selectors; duplicate YAML settings and trailing documents reject.
- Prior accepted patch, peer, PnP runtime-state, cache, lifecycle, and protected-invocation behavior was preserved.

Pinned Yarn 4.9.2 was probed locally. It emits workspace lock version `0.0.0-use.local` independently of manifest semver; the implementation binds that Yarn-owned sentinel while retaining manifest semver as the workspace package version.

## Regression coverage

- Positive root plus child workspace graph with Yarn-generated descriptor shape.
- Missing root/child entry, duplicate entry, unmatched/path drift, name drift, version drift, and selector drift.
- Positive and negative dependency, development, optional, peer, and peer-optional metadata projections.
- Quoted/scalar boolean impostors, quoted integer, unknown nested architecture key, non-string selector, ambient `current`, duplicate selector, duplicate setting, and second YAML document.
- Rejected workspace variants assert zero packages, edges, and configuration identity before manager execution.

## Validation

All commands were standalone gates; no pipe or `tee` obscured status.

| Gate | Exit | Evidence |
| --- | ---: | --- |
| Reviewer reproduction probe | 0 | All unsafe probes now return `closure_lock_stale` or `closure_lock_format_unsupported`. |
| Real pinned Yarn 4.9.2 + staged Node PnP invocation under `sandbox-exec` network denial | 0 | `TestN01RealPinnedYarnPnPInvokeThroughVerifiedExecutor` passed. |
| Focused modern Yarn suite | 0 | `.temp/TASK-260811-32iojo/go-test-focused-02.log` |
| Focused race suite | 0 | `.temp/TASK-260811-32iojo/go-test-race-02.log` |
| Focused coverage | 0 | 75.0%; `.temp/TASK-260811-32iojo/go-test-cover-02.log` |
| `golangci-lint run` | 0 | `0 issues.` |
| `go vet ./...` | 0 | `.temp/TASK-260811-32iojo/go-vet-01.log` |
| `go build ./...` | 0 | `.temp/TASK-260811-32iojo/go-build-01.log` |
| `go test -count=1 ./...` | 0 | Direct tracked process; `.temp/TASK-260811-32iojo/go-test-full-02.log` |
| `gofmt -l internal/yarnmodernsource` | 0 | Empty output. |
| `git diff --check` | 0 | Empty output. |
| `task-board validate` | 0 | `Board is valid. No issues found.` |

An earlier full-suite invocation also produced a green log, but its shell wrapper did not retain the gate process exit status. It is intentionally not counted above; the independently rerun `go-test-full-02` gate is the authoritative exit-0 evidence.

## Source identities

- `internal/yarnmodernsource/lock.go`: `d53a79f6c698fb0ad51fad13f2ed1ef953cf9f054a24a98604edede1ea1c3c63`
- `internal/yarnmodernsource/conformance_test.go`: `2edc362c56b7542df81f03b8b559b4f56c5c83000d0334b940b60bd4dcd4c4d5`

The pre-existing worktree is intentionally dirty and `internal/yarnmodernsource/` remains untracked as part of the broader uncommitted adapter delivery. No files were staged, committed, reset, cleaned, or destructively modified.
