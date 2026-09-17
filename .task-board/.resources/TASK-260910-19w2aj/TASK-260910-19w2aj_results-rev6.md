# TASK-260910-19w2aj rev6 results — legacy frozen identity recovery (rework-5)

## Scope
Rework-5 two P1s in frozen identity recovery for LEGACY root entries
(skillfile-sources §1 retains unchanged legacy named entries). Preserves
rev4 (SkillsRoot, exact-alias recovery, full Git inventory authentication,
real Git selector marker, refresh rollback) and all prior rework fixes.

## Changes (worktree delta vs checkpoint ff433a1, plus prior rev1-5 delta)
- [draftsources.go](/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260910-3vxe3y/worktree/internal/install/draftsources.go:112): `draftFrozenInput` now branches on the indexed declaration. Draft selector roots keep the bindings path; legacy roots (`Selector == nil`, `decl.Name == member.Name`) recover `Source/Git/Kind/Ref` from the indexed `name/source/git/tag/branch/revision` declaration, map `GitRepos` to `SkillsRoot/<decl.Source>`, and set `GitKeys` to the legacy source key. Transitive path unchanged. Stale bindings / missing repo still fail closed.
- [resolve.go](/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260910-3vxe3y/worktree/internal/closure/resolve.go:245): `FrozenOptions` gains `GitKeys` override; `OpenDraftFrozen` uses it when present, falling back to `gitCacheKey`. `gitCacheKey` comment reconciles the legacy exception (legacy network roots snapshotted under `decl.Source`, not canonical repository).
- [draft_sources_test.go](/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260910-3vxe3y/worktree/cmd/curator/draft_sources_test.go:568): new production CLI regressions through `project resolve` + install (lock created only via CLI):
  - `TestProjectResolveLegacyConfiguredGitThroughCLI/default|custom-source`: resolve → dry-run (asserts `review tag v1 `) → real install (marker `tag/v1/commit`, `Source` default `review` or `custom/review`, `Git == ""`) → pinned reinstall.
  - `TestProjectResolveLegacyNetworkGitThroughCLI/default|custom-source`: same flow with `git: https://fixture.test/review.git` and a SkillsRoot checkout (with `origin`); marker asserts `Source` and `Git == declaredURL`.
  - `TestProjectResolveLegacyMixedThroughCLI`: legacy configured-git + draft local selection; resolve → dry-run asserts both members (`legacy-review tag v1 `) → pinned dry-run. Real-install for the local member is out of scope (local draft carries no Git ref for the schema-2 marker); legacy real-install covered above.

## Evidence (narrow, tool calls < 2 min; host stalls on wide suites)
- `go vet ./cmd/curator ./internal/closure ./internal/install` → exit 0.
- `gofmt -l` on the three touched files → empty (exit 0).
- `golangci-lint run ./internal/closure/... ./internal/install/... ./cmd/curator/...` → 0 issues (exit 0; fixed one `revive` empty-block).
- `go test -p 1 ./internal/closure -run 'TestDraft|TestOpenDraft|TestResolveDraft' -count=1 -timeout=90s` → ok 9.074s, exit 0.
- `go test -p 1 ./internal/install -run 'TestDraft|TestLegacyInstall' -count=1 -timeout=90s` → ok 10.312s, exit 0.
- `go test -p 1 ./internal/closure -run 'Draft|Refresh|Legacy|Frozen|Transitive' -count=1` → ok 10.649s, exit 0.
- `go test -p 1 ./internal/install -run 'Draft|Refresh|Legacy' -count=1` → ok 13.090s, exit 0.
- `go test -p 1 ./cmd/curator -run 'TestProjectResolveLegacy' -count=1 -timeout=120s` → ok 28.390s, exit 0 (2 configured + 2 network + 1 mixed subtests).
- `go test -p 1 ./cmd/curator -run '^TestProjectResolve' -count=1 -timeout=120s` → ok 93.907s, exit 0 (all 13 top-level resolve tests incl. SkillsRoot, exact-alias, runtime-tamper, real-install, refresh).
- Reviewer rev5 repro (`TASK-260910-19w2aj_review-rev5-repro.py` with binary built from this tree at `/tmp/TASK-260910-19w2aj-cli`): both `configured` and `network` cases now `project resolve` exit 0, `install --dry-run` exit 0 (`review tag v1`), real `install` exit 0 (was: `source_member_invalid: locked review has no declaring source` and `install marker is invalid for schema 2`). Full log tail attached in handoff run output.

## Mutants (both killed)
- M1 cache-key narrowing (`cacheKey := member.Package.Repository` instead of `decl.Source`): `TestProjectResolveLegacyNetworkGitThroughCLI/custom-source` FAILs with `source_snapshot_unavailable` (wrong cache dir). Restored.
- M2 ref narrowing (drop legacy `Kind/Ref` recovery): `TestProjectResolveLegacyConfiguredGitThroughCLI/default` FAILs with `source_member_invalid: locked review has no declaring source` at dry-run. Restored. Tree verified restored via `grep` + `go vet`.

## Checklist
- [x] Implement the scoped production behavior with traceability to the accepted draft contracts.
- [x] Run task-specific positive, negative and legacy regression checks; record exact revision and evidence for independent review.
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant

## Notes / findings
- Local draft real installs carry no Git ref, so the schema-2 marker stays invalid for them; the mixed test therefore exercises dry-run only. This is pre-existing behavior outside rework-5 scope (no prior test does a local real install); flagged here, not changed.
- Workspace carries only the 1a75qd checkpoint plus this leaf's delta (3 modified tracked files from prior revs + 6 task files, no stray files). Leave uncommitted for the handoff snapshot.
