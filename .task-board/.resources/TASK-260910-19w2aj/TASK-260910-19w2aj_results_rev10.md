# TASK-260910-19w2aj rev10 outcome (rework 9, story_final)

Base: trunk 0945447816cb + lock checkpoint 68daaef + rev9 candidate patch applied (`git apply --binary`, exit 0; status lists the rev9 paths). Product code unchanged from rev9; fixture-only fix.

## Fix (rework 9)
Gate run 35223050543 failed `TestProjectRefreshFetchFailurePreservesStateThroughCLI` on Windows: the shared `setupLegacyBranchRefresh` helper builds a script skill (`unix_path`-only command), and `commandRel` (`internal/runtimestore/runtimestore.go:256`) resolves to the empty `WinPath` on Windows, so the pre-refresh real install fails `script command tool path is not portable` (`internal/runtimestore/scripts.go:109`) before reaching the assertion. Sibling refresh tests only use `--dry-run`; every script-skill real-install test skips on Windows (`test transport wrapper is POSIX-only`).

Fixture-only change in `cmd/curator/draft_sources_test.go` (new file from rev9 patch, untracked): parameterized the helper as `setupLegacyBranchRefreshWithSkill(t, payload, writeSkill)` (default wrapper keeps `writeCLIScriptSkill` for the branch-membership tests that assert cached `scripts/tool.sh` bytes); the fetch-failure test now uses `writeCLISkill` (empty commands, the same portable shape as the `TestProjectResolveLegacy*` fixtures that real-install on Windows). The test needs no runtime bytes: it asserts lock/bindings/installed-SKILL.md preservation plus post-failure `--dry-run`.

Audit: every remaining unskipped-on-Windows real `install` in this file now uses a portable fixture (script-skill real installs are all behind the POSIX-only skip; dry-runs never hit `validateScriptSpec`).

## Evidence (darwin, narrow, -p 1 -count=1)
- `go vet ./cmd/curator/` -> exit 0
- `go test ./cmd/curator/ -run TestProjectRefreshFetchFailurePreservesStateThroughCLI -v` -> PASS (6.87s), exit 0
- sibling regressions `TestProjectRefreshLegacyConfiguredGitBranchThroughCLI|TestProjectRefreshLegacyNetworkGitBranchThroughCLI|TestProjectRefreshTransitiveTagAdvanceThroughCLI` -> all PASS, exit 0
- `TestProjectResolveLegacy*` -> ok (51.012s), exit 0
- `gofmt -l cmd/curator/` clean; `golangci-lint run cmd/curator/...` -> 0 issues, exit 0
- Windows lane cannot run on this host; portability follows structurally (no script commands -> `validateScriptSpec` loop vacuous on every platform) plus the Legacy-shape precedent that already installs on Windows.

## Preserved
All rev2-rev8 accepted fixes untouched (SkillsRoot, exact-alias recovery, legacy entries, frozen-tree expansion, allowlist propagation, refresh fetch, rollback). No spec edits, no new skip class, switch-gated draft lane only.
