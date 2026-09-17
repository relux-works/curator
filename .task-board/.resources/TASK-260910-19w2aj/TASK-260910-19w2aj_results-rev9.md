# TASK-260910-19w2aj — results rev9 (rework 8: explicit refresh fetches legacy/transitive repos)

Story final leaf. Implements verdict-rev8 P1: `project resolve`/`project refresh`
now refresh applicable legacy and transitive repositories through the admitted
acquisition path. Preserves all rev2–rev8 fixes (allowlist, SkillsRoot,
exact-alias/legacy identity, full Git inventory auth, real Git marker, refresh
rollback, declared-commit expansion).

## Production changes

- `cmd/curator/project_resolve.go` (`resolveDraftPlan`): set `Fetch: true` on
  `DraftResolveConfig` with one shared dedup map; alias trees just cloned/fetched
  by `acquireDraftGitRoots` are pre-marked via `closure.MarkRepoFetched`, so the
  pinned closure never fetches them twice. Failure still precedes publication,
  so a failed fetch leaves lock/bindings/install unchanged.
- `internal/closure/resolve.go`: `DraftResolveConfig.Fetch` documented (explicit
  resolve/refresh set it; frozen install/launch never fetch); `pinGitAliases`
  dedup key unified to the shared closure `repoKey` (was an ad-hoc
  lower+cased clean path that never matched `ensureRepo`'s key).
- `internal/closure/closure.go`: new `MarkRepoFetched` (same key `ensureRepo`
  consults; nil-safe). `ensureRepo` itself unchanged: with `FetchExisting` now
  true it fetches reached legacy/transitive checkouts through the existing
  policy-gated lane (clones still gated; fetches use the checkout's own remote).
- `internal/gitops/gitops.go`: new `HasRemote`; `Fetch` is a successful no-op
  on remote-less trees, so origin-less configured checkouts are never fetched
  (previously `git fetch --all` no-op'd at the tool; now not invoked at all).
- `cmd/curator/draft_sources_test.go`: existing
  `TestProjectResolveLegacyNetworkGitThroughCLI` fixture now clones its checkout
  from a local bare repo (origin fetchable offline); manifest-declared fixture
  URL keeps the network-git identity under test. Required because explicit
  resolve now fetches that checkout (its old unreachable `origin` would hang/fail).

## New CLI regressions (all through `project resolve`/`refresh` + install, temp config)

- `TestProjectRefreshLegacyConfiguredGitBranchThroughCLI` — `{name,branch}`
  runtime-only upstream advance → refresh publishes new commit + complete runtime
  bytes; checkout HEAD stays stale; pinned dry-run green before and after.
- `TestProjectRefreshLegacyNetworkGitBranchThroughCLI` — same for
  `{name,branch,git}`; lock keeps network-git kind; declared identity never
  contacted (checkout's own local origin serves the fetch).
- `TestProjectRefreshTransitiveTagAdvanceThroughCLI` — consumer v1→v2 while the
  provider checkout lacks v2 → refresh fetches the transitive repo and locks v2
  with complete membership.
- `TestProjectRefreshFetchFailurePreservesStateThroughCLI` — bare removed →
  refresh exits 1 with `failed to fetch`; lock bytes, bindings bytes and the
  real-installed skill byte-identical; post-failure dry-run still green.
- `TestProjectResolveOriginLessSkipsFetchThroughCLI` — origin-less legacy
  resolve succeeds; git call log contains zero `fetch`.
- `TestProjectRefreshAliasFetchDedupedThroughCLI` — fresh resolve fetches 0
  times, refresh fetches exactly 1 time, lock SHA unchanged without upstream
  change.

## Evidence (zsh, `set -o pipefail` not needed; exit codes observed)

- `go build ./...`: exit 0. `go vet` on touched packages: exit 0.
  `golangci-lint run ./cmd/curator/ ./internal/closure/ ./internal/gitops/`:
  exit 0, 0 issues. `gofmt -l` on touched files: clean.
- `go test -p 1 -vet=off ./cmd/curator -run '^TestProjectRefresh'`: exit 0,
  6/6 PASS (26.6s, incl. pre-existing branch-membership test).
- `go test -p 1 -vet=off ./cmd/curator -run '^TestProjectResolve'`: exit 0,
  18/18 PASS (104.9s, incl. legacy goldens, allowlist, tamper, alias, tag/HEAD).
- `go test -p 1 -vet=off ./internal/closure/`: exit 0 (35.5s; existing
  `TestFetchExistingIsScopedToReachedClosureAndDeduplicated` passes unchanged —
  injected `FetchRepo` bypasses the remote-less guard by design).
- `go test -p 1 -vet=off ./internal/gitops/ ./internal/snapshot/`: exit 0.
- Reviewer rev8 repro (`TASK-260910-19w2aj_review-rev8-refresh-repro.py`) against
  a fresh production binary: default AND `--network` forms now fail at the
  `stale pin` assertion with `REFRESH_LOCK == REMOTE_NEW` on the FIRST refresh
  (no manual fetch) — the reported bug is gone. Probe scripts kept under
  `/tmp/rev8check/` (binary + repro copy); raw mutant logs at
  `/tmp/mutant1.log`, `/tmp/mutant2.log`, suite logs at
  `/tmp/refresh-all.log`, `/tmp/resolve-all.log`, `/tmp/closure.log`,
  `/tmp/gitops.log`.

## Mutants (both killed, raw logs above)

- Aliases-only (`FetchExisting: false`, pin still fetches): configured branch
  refresh FAILs in 2.9s with `refreshed commit = <old>, want <new>` (stale lock);
  full 4-test refresh set FAILs. Restored.
- No pre-marking (drop `MarkRepoFetched` loop): dedup test FAILs in 2.8s with
  `fresh resolve fetched 1 times, want 0` (pin refetches the just-cloned alias).
  Restored; both files verified `MUTANT`-free.

## Checklist

- [x] Explicit resolve/refresh fetches legacy + transitive repos via admitted path
- [x] Dedup: aliases never double-fetched (pre-mark + shared repoKey)
- [x] Origin-less checkouts never fetched; frozen install/launch stay offline
- [x] Fetch failure preserves lock/bindings/installed state (tested at CLI)
- [x] Allowlist/SkillsRoot/legacy-identity/rollback behavior preserved (suites green)
- [x] Narrow tests only, `-p 1`, tool calls bounded; linter clean
- [x] Workspace holds only the 1a75qd checkpoint + this leaf's delta (uncommitted)
