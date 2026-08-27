# TASK-260729-365r5r — literal edit allowlist

## Baseline, explicitly recorded

Source: `.temp/TASK-260729-rfrdfo/worktree` (the rfrdfo prototype state), copied
with `rsync -a --exclude=.git --exclude=.temp --exclude=.task-board` into
`.temp/TASK-260729-365r5r/worktree`. A byte-identical second copy is kept at
`.temp/TASK-260729-365r5r/worktree-baseline` and is never edited.

Chosen over the pristine `TASK-260720-jrrgw9` candidate because the 480 s
atomicity target is quoted against the rfrdfo gate timings
(`gate-race-atomicity-{1,2,3}` = 593 s / 561 s / 564 s), and the rfrdfo delta is
test-only (13 files under `internal/install`), disjoint from every file below.

Path-sorted pre-manifest: `evidence/manifest-pre.txt` (391 entries, 306 outside
`.task-board/`), SHA-256, `.git` excluded, produced by `bin/manifest.sh`.

### Restoration point

RUN-260729-bd5fd3 introduced a cross-save verdict cache
(`namespaceGraphAccepted` / `acceptNamespaceGraph` / `forgetNamespaceGraph` plus
`Engine.namespaceChecked` / `Engine.namespaceGraph` / `Engine.namespaceMu`) and
was rejected. This run restored `engine.go`, `journal.go` and `namespace.go`
from `worktree-baseline` before making any edit and re-ran the manifest:

```
diff evidence/manifest-restored.txt evidence/manifest-pre.txt   # exit 0, 391 lines
```

The tree was byte-identical to its recorded pre state before the first edit
below. The rejected design's own notes are archived unmodified under
`evidence/rejected-cache-design/`.

## Product files — allowed

1. `internal/transaction/namespace.go` — the only product file edited.
   - `type resolvedNamespacePath` — **new**; one declared path as a single pass
     sees it: the pre-split key, its per-component NFD form, and a lazily read,
     once-per-pass filesystem identity.
   - `resolveNamespacePath` — **new**; the per-path half of the sweep, taken
     once. Deliberately does not touch the filesystem.
   - `(*resolvedNamespacePath).identity` — **new**; the `identityRead` guard.
   - `validateIndependentTargetNamespaces` — `paths` becomes
     `[]resolvedNamespacePath`, the three append sites wrap their literal in
     `resolveNamespacePath`, and the pairwise loop takes `&paths[i]`. No branch,
     no early exit, no ordering change.
   - `namespacePathsOverlap` — takes `*resolvedNamespacePath`; reads identity
     through `identity()` instead of `namespaceIdentity` directly.
   - `namespaceContains` — takes `*resolvedNamespacePath`; selects the raw or
     the NFD pre-split instead of calling `namespaceComponents` per pair.
   - `namespaceComponentEqual` — loses the `normInsensitive` parameter, which
     moved to `resolveNamespacePath`; case folding stays.

   Untouched in that file: `targetNamespacePath`,
   `canonicalNamespaceTargetPath`, `canonicalNamespacePath`,
   `namespaceIdentity`, `namespaceComponents`, `existingNamespaceAncestor`.

**Not edited**: `internal/transaction/engine.go` and
`internal/transaction/journal.go` are byte-identical to `worktree-baseline`.
`saveJournal`, `validateJournal`, `(*Engine).validateJournal`, `loadJournal`,
`buildJournal` and `type Engine` are unchanged.

## Test files — allowed

2. `internal/transaction/namespace_pass_test.go` — **new file only**
   - `TestNamespaceIdentityIsReadOnceWithinOneValidationPass`
   - `TestNamespaceIdentitySnapshotDoesNotOutliveItsPass`
   - `TestValidateIndependentTargetNamespacesRejectsMalformedPaths` (9 cases)
   - `TestValidateIndependentTargetNamespacesRejectsOverlappingPaths` (7 cases)
   - `TestValidateIndependentTargetNamespacesAcceptsDisjointPaths`
   - `TestSaveJournalRejectsNamespaceAliasIntroducedBetweenSaves` (2 cases)
   - `TestRecoverRejectsDecodedTargetNamespacesAliasedWhileStopped`
   - `BenchmarkValidateIndependentTargetNamespaces`
   - helpers `namespaceTargetRecord`, `namespaceBenchmarkTargets`

No existing test file was edited. No assertion anywhere was weakened, deleted,
skipped, or retimed.

### RUN-260729-d36102 lint rework — the only later edit

Four `build` closures inside
`TestValidateIndependentTargetNamespacesRejectsOverlappingPaths` had their
unused `t *testing.T` parameter renamed to `_ *testing.T`: the `nested live
paths`, `repeated live path`, `live path is another target's backup sidecar`
and `live path is another target's cleanup tomb` cases. The other three closures
use `t` and were left alone.

Eight changed lines in one file, no assertion touched, no test added or removed,
no case removed — the table still has 7 cases and the verbose gate still reports
25 PASS / 0 SKIP. `namespace.go` was not touched (`bb332038…` before and after).
`evidence/rework-lint.md` carries the diff and the gate ledger.

## Explicitly out of scope — and confirmed untouched

- No journal schema change: no `Journal` / `TargetRecord` field added, removed,
  or retagged; nothing new is serialized.
- No cross-save cache or process-local verdict of any kind. `type Engine` gains
  no field. `namespace.go` has no package-level `var`.
- No timeout value, no `-timeout` token, no CI file, no protocol or conformance
  fixture.
- `.temp/TASK-260729-rfrdfo/worktree`, `.temp/TASK-260720-jrrgw9/worktree` and
  `.temp/TASK-260729-365r5r/worktree-baseline` are never written to.
- `.temp/TASK-260729-365r5r/equivcheck/` is a throwaway third copy used only to
  run the new tests against unmodified baseline product code. It is not a
  deliverable.
