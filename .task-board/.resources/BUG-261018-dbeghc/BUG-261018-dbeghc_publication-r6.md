
## Publication after converge — r6 retry

The candidate remains based on `892dab098aef1497df9ccacdb173faaf4f8b8418`, candidate tree `efc4e2371c2f7e21ec8b56d9ffc0273b5c98c722`. All five implementation/test/docs blobs match the previous publication evidence. Root README and CHANGELOG match current trunk `59746a4545db06607446dcfe5fc35f28118a9125` exactly. No source, test, documentation, index or producer commit was changed in this run.

The deferred-base instruction requires refresh because incoming trunk also changes `internal/envprofile/managed.go`. Inspection finds native-skills repair additions, separate from the mode changes; the tool did not install the combination. `task-board worktree refresh-candidate BUG-261018-dbeghc` exited **1** with the exact refusal:

```
change_request_candidate_drift: candidate refresh would leave tracked board paths outside its file transaction stale; synchronize the board with the prepared checkpoint before retrying (paths=.task-board/.activity/BUG-261018-uoisfe/events.ndjson, .task-board/.activity/EPIC-261016-stqc77/events.ndjson, .task-board/.activity/STORY-261018-k5yd2n/events.ndjson, .task-board/EPIC-261016-stqc77_curator-rc7-code-wave/STORY-261018-k5yd2n_native-skills-in-place/BUG-261018-uoisfe_in-place-profile-skills-miss-native-surface/README.md, .task-board/EPIC-261016-stqc77_curator-rc7-code-wave/STORY-261018-k5yd2n_native-skills-in-place/BUG-261018-uoisfe_in-place-profile-skills-miss-native-surface/progress.md, .task-board/EPIC-261016-stqc77_curator-rc7-code-wave/STORY-261018-k5yd2n_native-skills-in-place/README.md, .task-board/EPIC-261016-stqc77_curator-rc7-code-wave/STORY-261018-k5yd2n_native-skills-in-place/progress.md, .task-board/EPIC-261016-stqc77_curator-rc7-code-wave/progress.md, prepared_tree_oid=4b0d82cc6a7997ac15f1ce0d41ecedd3b52729d8, workspace_tree_oid=efc4e2371c2f7e21ec8b56d9ffc0273b5c98c722)
```

Prepared checkpoint tree: `4b0d82cc6a7997ac15f1ce0d41ecedd3b52729d8`; retained workspace tree: `efc4e2371c2f7e21ec8b56d9ffc0273b5c98c722`. Direct edits to board files and orchestrator impersonation are forbidden. Recovery requires the orchestrator/CLI owner to synchronize the tracked board/checkpoint projection through its supported flow, preserving this candidate, then resume refresh/handoff. Repeating the earlier identical drift refusal or weakening its guard is not a solution. No product decision or source workaround is needed.

Current product coverage: **4 of 5 AC rows driven**. AC1: `TestManagedSeedPreservesNativePermissions` -> `Resolve/gatherSeeds/applyPlan`; AC2: `TestManagedHomeRepairTightensDirectories` -> `Resolve/verifyHomeObserved/managedDirectory`; AC3: `TestManagedAuthoredFilesArePrivate` and `TestManagedFallbackPermissions` -> `Resolve/claudeSeed/applyPlan/copyLinkFallback/copyTree`; AC4: `TestManagedSeedRefusalsPreserved` -> `Resolve/gatherSeeds`, plus the narrowing controls. AC5 requires configured full validation after successful handoff; it is not claimed. The current retry brief has no surface table, a recorded brief gap; no catalog-wide coverage is claimed.

| Mutant | Gate narrowed to | Named failing test | Verdict / survivor bound |
| --- | --- | --- | --- |
| private-seed-0644 | Only 0600 seed copies widen | `TestManagedSeedPreservesNativePermissions/0600` | Killed; 0644 instead of 0600 |
| claude-trust-0644 | Only Claude trust output widens | `TestManagedAuthoredFilesArePrivate` | Killed; .claude.json 0644 instead of 0600 |
| single-fallback-0644 | Only single-file fallback widens | `TestManagedFallbackPermissions/tree-false` | Killed; 0644 instead of 0600 |
| executable-tree-fallback-0755 | Only 0755 source evades mask; mask token retained | `TestManagedFallbackPermissions/tree-true` | Killed; 0755 instead of 0700 |
| codex-directory-drift-admitted | Only Codex drift admitted | `TestManagedHomeRepairTightensDirectories/codex_cli` | Killed; wide home admitted |
| codex-directory-not-tightened | Only Codex directory protection skipped | `TestManagedHomeRepairTightensDirectories/codex_cli` | Killed; permission drift after repair |

Each control requires real child exit 1 and the named permission failure. No survivors. Bounds remain the named classes, excluding concurrent filesystem replacement/chmod. Out of contract: spec changes (leaf Scope); refreshing tool-owned seeds during repair (section 7.4); native passthrough/link targets and takeover backups (ownership contracts). Windows makes no new POSIX-mode claim. No reviewer test was removed or renamed; no repeat-of review verdict was issued. No rework source diff was added, inside or outside the named module. The existing docs note is explicitly in the leaf scope.

Logbook: the previously documented board/checkpoint drift refusal recurred for incoming native-skills board records. The candidate remains preserved; no shared runner, service, authentication or foreign process was changed. Full local/hosted validation was not rerun and is not represented as green on a combined tree. No post-turn-validation checklist item was added.

### Current commands and real exits

| Command | Exit | Evidence / limit |
| --- | ---: | --- |
| `go test -count=1 -timeout=5m -json -run '^Test(Managed\|Resolve)' ./internal/envprofile` | 0 | 63 named passes, including all six narrowing controls; child commands each fail with exit 1 for their intended permission reason |
| `go build ./...` | 0 | Build passes on retained base |
| `go vet ./internal/envprofile ./cmd/curator` | 0 | Static validation passes |
| `gofmt -l` on four changed Go files; empty-output assertion | 0 each | No formatting changes needed |
| `git diff --exit-code main -- README.md CHANGELOG.md` | 0 | Wave docs equal trunk |
| `git diff --cached --exit-code`; `git diff --check` | 0 each | Real index unchanged; no whitespace errors |
| `task-board worktree refresh-candidate BUG-261018-dbeghc` | 1 | Exact `change_request_candidate_drift` refusal above |

The new task-scoped JSON outcome records candidate hashes, all named passes, all six child failure outputs, command exits and log digests. Conformance and broad CLI suites were not rerun in this publication retry; their previous evidence is historical, not a current combined-tree result. Local environment remains macOS arm64 / Go 1.25.5. No new signed candidate or hosted gate result is claimed before the configured handoff snapshot.
