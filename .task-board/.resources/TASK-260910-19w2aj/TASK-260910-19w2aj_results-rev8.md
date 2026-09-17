# TASK-260910-19w2aj rev8 results — draft closure allowlist (rework 7)

Story: STORY-260910-3vxe3y (final leaf, story_final)
Base revision: 45e397ea8a30a307531ac3dd5b75b43110a5cccb (TASK-260910-1a75qd checkpoint on task-board/story/STORY-260910-3vxe3y)
Worktree: .temp/STORY-260910-3vxe3y/worktree (uncommitted delta, 9 paths: 3 modified + 6 new)

## Fix (rework 7 P1)

Production draft resolve dropped the machine acquisition allowlist for
legacy and transitive lanes:

- `closure.DraftResolveConfig` had no `AllowedSources`.
- `closure.ResolveDraft` built `closure.Options` without it, so
  `ensureRepo`/`gateSource` saw empty allowlist = allow-all.
- Alias roots were already gated in
  `cmd/curator/project_resolve.go:acquireDraftGitRoots`; legacy named
  roots and transitive dependencies cloned denied providers.

Change (preserves all rev2–rev6 fixes):

- `internal/closure/resolve.go`: added `DraftResolveConfig.AllowedSources`
  with Spec §8.2 semantics (empty allows all); `ResolveDraft` carries it
  into `closure.Options` for `BuildExpanded` legacy/transitive
  acquisitions.
- `cmd/curator/project_resolve.go:resolveDraftPlan`: passes
  `cfg.AllowedSources` (production CLI config) into
  `DraftResolveConfig` alongside the existing `SkillsRoot` fix.
  Alias gating unchanged.

Production call sites:

- `cmd/curator/project_resolve.go:113-131` (`resolveDraftPlan`)
- `internal/closure/resolve.go:163-174` (`closureOpts`)
- `internal/closure/closure.go:358-393,485-502` (`ensureRepo`/`gateSource`)

## Tests (production entry, no helper injection)

New committed CLI regressions in `cmd/curator/draft_sources_test.go`
(drive `project resolve` + `install --dry-run` through `capture()` with a
temp config; lock always created through the CLI):

- `TestProjectResolveDraftAllowlistLegacyThroughCLI/denied`: restrictive
  `allowed_sources=["allowed.test"]`, legacy root
  `{name:review, tag:v1, git:https://denied.test/review.git}` with no
  SkillsRoot checkout. Must fail `exitFail` with `source not allowed`,
  leave lock/bindings/installed unchanged (no `Skillfile.lock.json`, no
  bindings, no `skills-root/review`, no `.agents/skills/review`), and the
  logging-git recorder must contain no `denied.test`.
- `.../allowed`: `allowed_sources=["denied.test"]` + SkillsRoot `review`
  checkout (tag v1). Must resolve `exitOK`, lock contains `review`,
  `install --dry-run` green.
- `TestProjectResolveDraftAllowlistTransitiveThroughCLI/denied`: local
  consumer requires `provider` at `https://denied.test/provider.git` tag
  v1, no SkillsRoot checkout, restrictive config. Same refusals and
  recorder/state assertions.
- `.../allowed`: same provider URL allowlisted + SkillsRoot `provider`
  checkout. Must resolve with zero acquisitions, lock contains
  transitive `provider`, dry-run green.

Recorder: `installLoggingGitForDraftCLI` (POSIX `git` wrapper logging
`$*` to a file, then `exec` real git); skipped on Windows with the
declared reason `test transport wrapper is POSIX-only`.
Config edit: `setCLIAllowedSources` rewrites `allowed_sources` in the
temp `config.json` through the normal `fileConfigSource` load path.

## Evidence (narrow, `set -o pipefail`, shell `bash`, `-p 1`)

- `go test -p 1 ./cmd/curator -run TestProjectResolveDraftAllowlistLegacyThroughCLI -count=1 -v` → PASS (3.688s), EXIT:0
  - `denied` PASS (0.55s), `allowed` PASS (2.58s)
- `go test -p 1 ./cmd/curator -run TestProjectResolveDraftAllowlistTransitiveThroughCLI -count=1 -v` → PASS (4.081s), EXIT:0
  - `denied` PASS (0.70s), `allowed` PASS (2.59s)
- Mutant (drop `AllowedSources` from `closureOpts`, i.e. alias-only
  policy): both `/denied` subtests FAIL as required, EXIT:1
  - legacy denied: `failed to clone review from https://denied.test/review.git ... Could not resolve host: denied.test`, want `source not allowed`
  - transitive denied: `failed to clone provider from https://denied.test/provider.git ...`, want `source not allowed`
  - restored file afterwards; `grep AllowedSources` confirms fix present.
- Regressions (all EXIT:0):
  - `TransitiveProviderUsesConfiguredRoot + IgnoresProcessCWD` → PASS (6.020s)
  - `LegacyConfiguredGit + LegacyNetworkGit` (default + custom-source) → PASS (48.523s)
  - `LocalCreatesLock + GitAliasSelection` → PASS (11.163s)
  - `GitTagDiffersFromHEAD + GitCollectionPinnedToTag` → PASS (25.078s)
  - `go test -p 1 ./internal/closure -run TestResolveDraft|TestRefreshDraft|TestBuildExpanded` → ok (20.735s)
- `go vet ./internal/closure ./cmd/curator` → VET_EXIT:0
- `golangci-lint run internal/closure/... cmd/curator/...` → 0 issues, LINT_EXIT:0
- `gofmt -l` on touched files → clean.

## Checklist

- [x] Implement the scoped production behavior with traceability to the accepted draft contracts.
- [x] Run task-specific positive, negative and legacy regression checks; record exact revision and evidence for independent review.
- [x] Implementation matches AC (allowlist enforced for legacy/transitive with existing semantics; alias/refresh/rollback/SkillsRoot/alias-recovery/legacy/marker fixes preserved).
- [x] Solution fits project architecture (reuses `identity.Allowed`/`gateSource`/`ensureRepo`; no new policy semantics).
- [x] Tests green (narrow CLI + closure runs above).
- [x] Code written per task description and AC.
- [x] Relevant tests written for new or changed behavior and passing (2 new CLI tests, 4 subtests).
- [x] Lint clean.
- [x] Relevant build/validation commands run after changes and build not broken (vet + golangci-lint + gofmt).
- [x] New outcome artifact attached on the board with a task-scoped name (this file).
- [x] Important findings recorded (none beyond the fixed P1; no new anomalies).

## Workspace

`git status --short` at handoff:

- ` M cmd/curator/main.go`
- ` M internal/install/install.go`
- ` M internal/snapshot/snapshot.go`
- `?? cmd/curator/draft_sources_test.go`
- `?? cmd/curator/project_resolve.go`
- `?? internal/closure/resolve.go`
- `?? internal/closure/resolve_test.go`
- `?? internal/install/draftsources.go`
- `?? internal/install/draftsources_test.go`

Uncommitted as required; no stray files added by this leaf beyond the
rev6 candidate paths plus the allowlist delta. If handoff refuses
`change_request_base_authority_mismatch`, the orchestrator handles it —
do not retry.
