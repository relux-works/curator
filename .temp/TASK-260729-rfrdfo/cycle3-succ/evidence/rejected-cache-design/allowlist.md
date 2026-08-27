# TASK-260729-365r5r — literal edit allowlist

Baseline: `.temp/TASK-260729-rfrdfo/worktree` (the rfrdfo prototype state), copied
with `rsync -a --exclude=.git --exclude=.temp --exclude=.task-board` into
`.temp/TASK-260729-365r5r/worktree`. A byte-identical second copy is kept at
`.temp/TASK-260729-365r5r/worktree-baseline` and is never edited; it is what the
baseline gates measure.

Baseline chosen over the pristine `TASK-260720-jrrgw9` candidate because the
480 s atomicity target is quoted against the rfrdfo gate timings
(`gate-race-atomicity-{1,2,3}` = 593 s / 561 s / 564 s). Measuring the product
change on the same tree makes the before/after directly comparable. The rfrdfo
delta is test-only (13 files under `internal/install`), disjoint from every file
below.

## Product files — allowed

1. `internal/transaction/engine.go`
   - `type Engine` — two process-local fields: `namespaceGraph`, `namespaceChecked`
   - `(*Engine).buildJournal` — one call site: `validateJournal` → `validateFreshJournal`
2. `internal/transaction/journal.go`
   - `validateJournal` (package function) — recomposed, behavior unchanged
   - `validateJournalStructure` — new; the filesystem-free half, split out verbatim
   - `(*Engine).validateJournal` — structure always, namespaces through the verdict
   - `(*Engine).validateFreshJournal` — new; validation that never reuses a verdict
   - `(*Engine).discardNamespaceVerdict` — new
   - `(*Engine).validateTargetNamespaces` — new; the resolve-or-reuse decision
   - `(*Engine).loadJournal` — one call site: `validateJournal` → `validateFreshJournal`
3. `internal/transaction/namespace.go`
   - `type namespaceGraphDigest` — new
   - `validateIndependentTargetNamespaces` — recomposed, behavior unchanged
   - `resolveTargetNamespacePaths` — new; the target expansion, split out verbatim
   - `resolveReservedNamespacePaths` — new; the reserved expansion, split out verbatim
   - `assertNamespacePathsIndependent` — new; the pairwise sweep, split out verbatim
   - `assertNamespacePathIndependent` — new; one path against an independent set
   - `assertNamespacePairIndependent` — new; one pair, split out verbatim
   - `targetNamespaceGraph` — new; the reuse key
   - `writeNamespaceGraphComponent` — new

## Test files — allowed

4. `internal/transaction/namespace_graph_test.go` — new file only
   - `namespaceGraphJournal`, `aliasLiveTargets`, `appendNamespaceTarget`
   - `TestUnchangedTargetNamespaceGraphIsResolvedOncePerTransaction`
   - `TestRecoveryResolvesTheDecodedTargetNamespaceGraphBeforeMutating`
   - `TestSaveJournalResolvesEveryChangedTargetNamespaceGraph`
   - `TestTargetNamespaceGraphDistinguishesEveryDeclaredPath`

## Explicitly out of scope

- No journal schema change: `namespaceGraphDigest` lives on `Engine`, is never
  serialized, and no `Journal`/`TargetRecord` field is added, removed, or retagged.
- No timeout value, CI file, protocol or conformance fixture is touched.
- No existing assertion is weakened, deleted, or skipped.
- Neither `.temp/TASK-260729-rfrdfo/worktree` nor `.temp/TASK-260720-jrrgw9/worktree`
  is modified.
