# TASK-260729-365r5r — static call-path proof

Claim: after the prototype, every target namespace graph that can reach a
transaction mutation is resolved against the filesystem first, with no reuse of
a previously recorded verdict.

## 1. Only two functions admit a target graph into the engine

`Journal` values are produced in exactly two places in `internal/transaction`:

| producer | file | how the graph originates |
|---|---|---|
| `(*Engine).buildJournal` | `engine.go:115` | constructed from the caller's `Plan` |
| `(*Engine).loadJournal` | `journal.go:135` | decoded from stored bytes |

There is no third producer: `cloneJournal` (`engine.go:893`) copies an existing
in-engine journal and creates no new declaration, and `Journal` has no exported
constructor. Grepping the package for `&Journal{` returns `buildJournal` only;
`var journal Journal` returns `loadJournal` only.

Both producers call `(*Engine).validateFreshJournal`, which clears any held
verdict before validating:

- `engine.go:240` — `buildJournal` returns the journal only after it passes.
- `journal.go:181` — `loadJournal` returns the journal only after it passes.

## 2. Every mutating entry point goes through one of those two

`Engine` exports exactly five methods (`New`, `Prepare`, `Commit`, `Recover`,
`ReferencedBuildKeys`). The four that take a `HomeLock` all hold `engine.mu`
for their whole body, so the verdict fields are never touched concurrently.

```
Prepare  -> buildJournal -> validateFreshJournal   (before saveJournal, before staging)
Commit   -> loadJournal  -> validateFreshJournal   (before resume)
Recover  -> loadJournal  -> validateFreshJournal   (before resume, per journal id)
ReferencedBuildKeys -> loadJournal -> validateFreshJournal   (read-only)
```

Every mutating helper — `resume`, `commit`, `commitTarget`, `rollback`,
`rollbackTarget`, `cleanupCommitted`, `removePreparedSidecars`,
`resumeRecordedRemoval`, `removeRecordedSidecar`, `finishRecordedRemoval`,
`discardPreparing`, `stageTarget`, `createStagingEntry`, `copyStagingFile`,
`abortPreparation` — is unexported and reachable only from those four, so no
mutation is reachable from a graph that has not been freshly resolved.

Recovery is the case the acceptance criteria call out explicitly:
`Recover` calls `loadJournal` for each id and returns its error before `resume`
runs, so a stored graph that no longer resolves cleanly is refused with the
journal and every sidecar untouched. `TestRecoveryResolvesTheDecodedTarget
NamespaceGraphBeforeMutating` asserts exactly that.

## 3. What the held verdict can and cannot answer

`(*Engine).validateTargetNamespaces` (`journal.go:395`) answers from the held
verdict only when `namespaceChecked` is set **and** `targetNamespaceGraph`
returns the identical key. The key is a SHA-256 over, per target, the `Kind`
that decides whether a final component is resolved plus the four declared paths
`LivePath`, `StagedPath`, `BackupPath`, `RollbackPath`, and over the engine's
reserved `journalRoot`. Those are precisely — and only — the fields
`resolveTargetNamespacePaths` and `resolveReservedNamespacePaths` read.

Every component is length-prefixed before hashing, so two different graphs
cannot encode to the same bytes;
`TestTargetNamespaceGraphDistinguishesEveryDeclaredPath` asserts one mutation
per field plus a component-boundary collision case.

The verdict is cleared, not merely overwritten, before any resolution starts, so
a graph that fails resolution is never left recorded as accepted. The negative
test repeats each rejection twice for that reason.

## 4. The semantic delta, stated plainly

Before: every `saveJournal` re-resolved the graph, so an unchanged declaration
was swept against the filesystem once per durable journal write — 16 call sites
in `engine.go` and 7 in `staging.go`, several inside per-entry and per-chunk
loops.

After: an unchanged declaration is swept once per graph. Within one transaction
the engine therefore no longer re-detects a filesystem change made *after* the
graph was admitted and *before* the transaction finishes.

This is bounded, and it is not the property the check was carrying:

- Re-resolution never guaranteed independence *at* mutation time. The last
  sweep always preceded the mutation, so the same window existed before.
- The window is spanned by the caller's manager-home lock, which every entry
  point asserts via `requireHomeLock` before anything else.
- The moment the same bytes re-enter the engine — commit after restart, or
  recovery — the graph is resolved again from scratch, because both decode
  paths use `validateFreshJournal`.

`TestUnchangedTargetNamespaceGraphIsResolvedOncePerTransaction` pins both
halves: the elision is real (an aliased pair does not fail a repeat save), and
the fail-closed boundary is real (the same engine refuses the same bytes the
instant they are decoded).
