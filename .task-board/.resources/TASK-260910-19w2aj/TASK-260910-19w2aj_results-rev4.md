# TASK-260910-19w2aj results — rev4 (rework 3: two P1s from rev3 verdict)

Base HEAD `ff433a1`; work uncommitted in the Story worktree (story_final:
only the 1a75qd checkpoint plus this leaf's delta, no stray files).
`internal/marker/marker.go` untouched (frozen v1 validation unchanged);
the unchanged marker builder is fed real declared values.

## P1-a — complete Git snapshot authentication on every frozen read

- `internal/snapshot/snapshot.go`: new `AuthenticateGit(home, source,
  repo, commit)` — read-only counterpart of `Get`. Missing cache fails
  `source_snapshot_unavailable` (never extracts into place); a cache tree
  whose complete inventory (every regular file incl. runtime/build roots,
  plus permission bits, via `transaction.DigestPath` against a fresh
  read-back of the pinned commit) differs fails
  `source_snapshot_changed`. Local repo reads only; no fetch, no ref
  resolution, no live-source adoption.
- `internal/closure/resolve.go`: `OpenDraftFrozen`/`LoadDraftFrozenNodes`
  take `FrozenOptions{GitRepos, Refs}`; each distinct cached commit tree
  authenticates once, then locked subtrees are served with the existing
  containment (`serveGitSubtree`). No repo entry fails closed
  (`source_snapshot_unavailable`). `gitCacheKey` fixes the transitive
  lane: root members key on canonical repository identity (alias
  acquisition), transitive members on the member name (the legacy lane's
  `snapshotFor` key) — previously transitive frozen reads looked in the
  wrong cache dir.
- `internal/install/draftsources.go`: `draftFrozenNodes` builds the
  options from the manifest plus the machine bindings written by the
  explicit attempt (`DraftBindingsPath`, shared with `project
  resolve`/`refresh`); stale bindings fail stale. Transitive repos resolve
  to the SkillsRoot checkout the legacy lane persists.

## P1-b — accepted source identity reaches materialization

- `LoadDraftFrozenNodes` carries, for root members, the installed source
  name (legacy source default), the manifest ref kind/value and declared
  endpoint from `FrozenOptions.Refs` (built from manifest alias matched on
  canonical identity + bindings location); transitive members recover the
  exact requirement declaration (ref kind/value, endpoint, source=name)
  from their requirers' specs, matching first-resolution-wins
  unification. Empty stays empty (local snapshots declare no ref); no
  value is invented for validation.

## Tests (production entry; lock always created through `project resolve`)

- `cmd/curator/draft_sources_test.go` (reviewer fixture shape: local Git
  + fake transport + `scripts/tool.sh` command with `runtime_roots`):
  `...GitRuntimeTamperRefusedThroughCLI` (tampered runtime refuses
  dry-run AND real), `...GitMissingMemberRefusedThroughCLI`,
  `...GitRealInstallThroughCLI` (real install exit 0, marker
  `tag/v1/<commit>` + `source=review`; bare repo deleted, second real
  install exit 0 — pinned, no network).
- `internal/closure/resolve_test.go`:
  `TestOpenDraftFrozenGitRuntimeBytesAuthenticated` (tamper / missing /
  exec-bit-flip / missing-repo-fails-closed),
  `TestDraftFrozenNodesCarryDeclaredIdentity`,
  `TestDraftTransitiveGitDependencyFrozen` (lock + pinned consumption +
  recovered identity of a transitive requirement).
- Exec-bit subtest skips on Windows with the declared class verbatim
  ("Windows does not expose portable executable permission bits",
  `.github/ci/skip-classes.tsv` platform-control); Git CLI cases keep the
  existing "test transport wrapper is POSIX-only" skip.

## Evidence (zsh, `set -o pipefail`, Story worktree)

- `go build ./...` → exit 0
- `go vet ./internal/closure/ ./internal/snapshot/ ./internal/install/ ./cmd/curator/` → exit 0
- `GOOS=windows go vet ./internal/closure/` → exit 0
- `gofmt -l` (4 pkgs) → clean, exit 0; `git diff --check` → exit 0
- `golangci-lint run cmd/curator/... internal/closure/... internal/install/... internal/snapshot/...` → `0 issues`, exit 0
- `go test -p 1 -count=1 -run 'TestResolveDraft|TestOpenDraftFrozen|TestRefresh|TestDraft' ./internal/closure/` → ok, exit 0
- `go test -p 1 -count=1 -run 'TestDiamond|...|TestBuildExpanded' ./internal/closure/` (legacy complement) → ok, exit 0
- `go test -p 1 -count=1 -run 'TestDraftInstall|TestLegacyInstallUntouched|TestDraftSourcesSwitch' ./internal/install/` → ok, exit 0
- `go test -p 1 -count=1 ./internal/snapshot/` → ok, exit 0
- `go test -p 1 -count=1 -run 'TestProjectResolve' ./cmd/curator/` → 7/7 PASS, exit 0
- Mutant A (skip `AuthenticateGit`): tamper CLI test FAILS
  (`tampered dry-run install = 0, want source_snapshot_changed`) —
  narrowing to the projected hash admits runtime tamper; reverted.
- Mutant B (drop `Refs` enrichment): real-install CLI test FAILS with
  `error: review: install marker is invalid for schema 2` — the exact
  rev3 P1-b signature; reverted.
- Reviewer repro `TASK-260910-19w2aj_review-rev3-repro.py` against a
  rev4 binary: resolve 0, dry-run 0, baseline real install 0 (was 1),
  tampered dry-run 1 + tampered real install 1 with
  `source_snapshot_changed ... does not match its locked commit`
  (was 0), published runtime still `echo original`.

## Checklist

- [x] Runtime-only tamper refused dry-run AND real; missing member refused
- [x] Real Git install publishes a valid marker with declared identity
- [x] Pinned second install uses the lock with the repository deleted
- [x] Conflict / root-only branch rules retained (untouched paths green)
- [x] Legacy goldens byte-identical, switch off (tests green)
- [x] Lint clean; narrow tests green with real exit codes
- [x] Workspace holds only checkpoint + leaf delta; no commits
- [x] No live credential export or runtime-home modification

## Notes for review

- Local-snapshot real installs still cannot produce a v2 marker (no
  declared ref exists to carry); they fail exactly as before with the
  marker validation error. Truthful draft markers (marker schema 5) are
  owned by STORY-260910-1s75e1 (backlog); this leaf carries only accepted
  values and invents none.
- Windows/macOS hosted lanes not rerun locally; run 35183472040-class
  platform gate stays hosted evidence.
