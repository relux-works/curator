# TASK-260720-29hi1h implementation evidence

## Provenance

- Isolated worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-29hi1h/worktree`
- Base: exact `origin/main` commit `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`
- Imported predecessor: complete accepted product diff from `TASK-260720-11pfex`
- Candidate conformance root: `TASK-260720-3ag6pi` rc.4 candidate used only as non-release test input
- No commit or staging performed

## Task-only implementation

- Added a closed `RuntimeTarget` model with typed `ScriptTarget` and `CompiledTarget` variants. Compiled targets can be constructed only from an exact native protected-cache hit and retain cache-key and receipt-hash identity.
- Added `PrepareScriptRuntime`, which validates every declared runtime root and active script path before reuse. Missing, unsafe, non-executable, incomplete, or unmanaged commit-keyed state produces a complete operation-private staged replacement while leaving the live runtime untouched.
- Script staging copies only declared runtime roots, or only explicitly selected single-script files when there are no roots. Snapshot build roots and compiled artifacts are not representable as script-runtime inputs and never enter `runtime/<skill>/<commit>`.
- Added typed manager-owned project, global-canonical, and safe-forwarding shim destinations plus deterministic `DesiredTarget` and `RemovalTarget` planning. Only staged paths are written; live project/global/user-bin replacement and deletion remain transaction-owned by `TASK-260720-2284br`.
- Unix staged shims use `exec` with quoted immutable/runtime paths and `"$@"`. Windows wrappers use delayed-expansion-disabled batch files, conditional PATH inheritance, `call "<artifact>.exe" %*`, and `exit /b %ERRORLEVEL%`.
- All three compiled shim roles point directly to the same validated immutable cache artifact rather than chaining through or copying build roots.
- Added candidate launcher-contract coverage and safe global-bin selection integration coverage.

## Behavioral evidence

- Incomplete script runtime: staged replacement is validated; the incomplete live tree remains unchanged.
- Script-to-build and build-to-script: deterministic desired replacements; no unmanaged removal targets.
- Global publication removal and full command removal: only typed manager-owned paths are emitted, in deterministic order; an adjacent user-owned file remains unchanged.
- Unix post-install fixture: project, global canonical, and safe-forwarding shims forward spaces, embedded quotes, percent signs, Unicode, and empty arguments; preserve empty or inherited PATH; return exit code 37; and propagate SIGTERM through `exec`.
- Windows post-install fixture source covers spaces, embedded quotes, `%PATH%`-shaped percent text, Unicode, empty arguments, percent-bearing paths, empty inherited PATH, and exit code 37 through a copied `.exe` helper.
- Install-time/staging sentinel proves the compiled artifact is not launched while targets and shims are staged.

## Verification

- `go test ./...`: pass
- `go test -race ./internal/runtimestore ./internal/globalbins`: pass
- `go test -count=10 ./internal/runtimestore ./internal/globalbins`: pass
- Candidate-root focused tests for runtimestore, globalbins, buildcache, buildmeta, closure, whitelist, and skillcheck: pass
- `go vet ./...`: pass
- `go build ./...`: pass
- `gofmt -l internal/runtimestore/*.go internal/globalbins/*.go`: clean
- `git diff --check`: pass
- Linux amd64 full test compile (`go test -exec=/usr/bin/true ./...`) and `go build ./...`: pass
- Windows amd64 full test compile (`go test -exec=/usr/bin/true ./...`) and `go build ./...`: pass
- Focused package coverage: runtimestore 72.5%, globalbins 73.5%; core new transition planner 82.5%, script-spec validation 81.0%, Unix/Windows content generation 100%

## Platform execution limitation

The Windows post-install fixture was cross-compiled but could not be executed on this macOS host because no Windows execution runtime or VM is installed (`wine`, Docker, Parallels, and Multipass are unavailable). The `_windows_test.go` fixture is present for execution by Windows CI; all Windows package tests and binaries compile successfully.
