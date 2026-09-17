# TASK-260910-19w2aj results — rev11 (rework 10)

## Scope
Rework 10 only (rev4 verdict P1): `internal/gitops` remote-enumeration
failure vs absence. Rev9 candidate already applied in-tree (rework 9
Windows-portable fixture preserved); no re-apply attempted
(`git apply --check` fails: paths already exist).

Base: `68daaef` (lock leaf 1a75qd) on
`task-board/story/STORY-260910-3vxe3y`. Work left uncommitted.
`git status --short` lists exactly the 11 rev9+rev10 paths:
M cmd/curator/main.go, M internal/closure/closure.go,
M internal/gitops/gitops.go, M internal/install/install.go,
M internal/snapshot/snapshot.go,
?? cmd/curator/draft_sources_test.go, ?? cmd/curator/project_resolve.go,
?? internal/closure/resolve.go, ?? internal/closure/resolve_test.go,
?? internal/install/draftsources.go, ?? internal/install/draftsources_test.go.

## Production change (rework 10)
- [gitops.go](/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260910-3vxe3y/worktree/internal/gitops/gitops.go:123): `HasRemote(repo string) (bool, error)` — failed `git remote` returns `(false, err)`; only a successful empty enumeration returns `(false, nil)`.
- [gitops.go](/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260910-3vxe3y/worktree/internal/gitops/gitops.go:137): `Fetch` propagates the enumeration error; the origin-less no-op happens only after a successful empty enumeration. No other `HasRemote` callers exist (only `Fetch`).
- Rework 9 preserved: `TestProjectRefreshFetchFailurePreservesStateThroughCLI` uses `writeCLISkill` (portable, no `unix_path` script), per the existing comment at `cmd/curator/draft_sources_test.go:1351-1358`.

## Tests (all rerun by producer, serial, `-p 1`, shell `set -o pipefail`)
- `go vet ./internal/gitops/` → exit 0
- `go vet ./cmd/curator/` → exit 0
- `gofmt -l internal/gitops/gitops.go cmd/curator/draft_sources_test.go` → clean
- `golangci-lint run ./internal/gitops/...` → 0 issues, exit 0
- `golangci-lint run ./cmd/curator/...` → 0 issues, exit 0
- `go test -p 1 -count=1 -run TestFetchAndSubmodules ./internal/gitops/ -v` → PASS (2.76s), exit 0
- `go test -p 1 -count=1 ./internal/gitops/` → ok 12.939s, exit 0
- `go test -p 1 -count=1 -run TestProjectRefreshRemoteEnumerationFailurePreservesStateThroughCLI ./cmd/curator/ -v` → PASS (7.32s; 125s wall on first compile, 8.371s cached), exit 0
- `go test -p 1 -count=1 -run TestProjectRefreshFetchFailurePreservesStateThroughCLI ./cmd/curator/ -v` → PASS (6.83s), exit 0
- `go test -p 1 -count=1 -run TestProjectResolveOriginLessSkipsFetchThroughCLI ./cmd/curator/ -v` → PASS (1.25s), exit 0
- `go test -p 1 -count=1 -run 'TestProjectRefresh(LegacyConfiguredGitBranch|LegacyNetworkGitBranch|TransitiveTagAdvance|AliasFetchDeduped)ThroughCLI' ./cmd/curator/ -v` → 4/4 PASS, exit 0
- One transient infra failure observed and retried serially: first concurrent run failed with `acquire package host GOROOT test lock: context deadline exceeded` (300s, `go vet` running concurrently); serial retry passed. No product change for this.

## New regression (production entry, temp config, CLI-driven)
- `TestProjectRefreshRemoteEnumerationFailurePreservesStateThroughCLI` (`cmd/curator/draft_sources_test.go`): legacy `{name,branch}` via `setupLegacyBranchRefreshWithSkill(..., writeCLISkill)` → real `install app` → snapshot lock/bindings/installed bytes → prepend POSIX git wrapper failing only bare `git remote` (exit 73) → `project refresh app` must exit `exitFail` with `failed to fetch` in stderr, lock/bindings/skill bytes unchanged, follow-up `install --dry-run` green. Skips on Windows via `requireRealGit` → verbatim `test transport wrapper is POSIX-only` (skip class `.github/ci/skip-classes.tsv:60`).
- Positive control kept and green: `TestProjectResolveOriginLessSkipsFetchThroughCLI` (origin-less checkout resolves with zero fetch lines).

## Mutant
- Collapsing mutant (`return false, nil` on remote error, `/tmp/gitops.go.bak` holds the pre-mutant tree): new test FAILs as required — `broken remote enumeration refresh = 0, stderr "", want a failed fetch`, exit 1. Restored afterwards and re-verified PASS (exit 0).

## Accepted from prior evidence (not rerun)
Deterministic plan bytes, frozen offline install/launch, full-inventory Git authentication, real Git selector marker, SkillsRoot/transitive root, exact-alias + legacy frozen identity, checkout-vs-commit expansion, allowlist propagation, alias/legacy/transitive refresh fetch + rollback, legacy goldens, conflict/root-only rules — per rev9 candidate and prior verdicts; this leaf reran only the refresh/fetch slice above plus lint.

## Checklist
- [x] HasRemote distinguishes absence (false,nil) from failure (false,err); Fetch propagates; origin-less skip only after successful empty enumeration
- [x] CLI negative test at production entry (exit 73 wrapper), state-preservation asserts, origin-less control kept
- [x] Narrowing mutant killed
- [x] Windows fixture fix preserved; new POSIX test uses declared skip sentence
- [x] Prior accepted fixes untouched
- [x] Narrow tests green with exit codes; lint clean
- [x] Workspace holds only checkpoint + leaf delta; work uncommitted
