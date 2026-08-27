# TASK-260729-365r5r — static call-path proof

Tree: `.temp/TASK-260729-365r5r/worktree` (prototype). Every claim below is a
grep over that tree; the commands are quoted so a reviewer can re-run them.

Superseded: `evidence/rejected-cache-design/` holds the proof and allowlist for
the cross-save verdict cache that the orchestrator rejected on RUN-260729-bd5fd3.
That design is not in this tree — see §3.

## 1. Every target graph is validated before any mutation

There are exactly four ways a target graph enters the engine, and all four pass
through `(*Engine).validateJournal` before the engine writes anything.

```
grep -n '^func (engine \*Engine) [A-Z]' internal/transaction/engine.go
  48: Prepare      270: Commit      285: Recover      859: ReferencedBuildKeys
```

| origin of the graph | entry point | validated at |
| --- | --- | --- |
| new graph built from a caller `Plan` | `Prepare` → `buildJournal` | `engine.go:234` `engine.validateJournal(journal)`, the last statement before `buildJournal` returns |
| resumed graph, named by id | `Commit` → `loadJournal` | `journal.go:176` `engine.validateJournal(&journal)` |
| recovered graph, swept from the journal root | `Recover` → `loadJournal` | `journal.go:176` |
| externally decoded bytes on disk | any `loadJournal` caller | `journal.go:176`, after `json.Decode` and **before** the canonical-bytes check at `journal.go:186` |
| every durable write of any graph | `saveJournal` | `journal.go:72` `engine.validateJournal(journal)`, the **first** statement of the function |

```
grep -rn 'loadJournal(' internal/ cmd/ | grep -v _test.go
  journal.go:135 (definition)  engine.go:276 (Commit)  engine.go:296 (Recover)  engine.go:871 (ReferencedBuildKeys)
```

`saveJournal` is fail-closed by construction: `validateJournal` is its first
statement, so `ensureJournalRoot`, `CreateTemp`, `Write`, `Sync`,
`durableReplaceFile` and `durableRenameNoReplace` are all unreachable for a
graph that failed validation. Nothing is created and nothing is renamed. There
are 23 `saveJournal` call sites across `engine.go` (16) and `staging.go` (7);
none of them can bypass this, because the check lives inside the callee.

```
grep -rn 'saveJournal' internal/ --include='*.go' | grep -v _test.go | wc -l   # 25
```

The 25 lines are the 23 call sites, the definition at `journal.go:71`, and one
prose occurrence in the `resolvedNamespacePath` doc comment at `namespace.go:34`
that the prototype itself adds. On the baseline tree the same command returns 24.
Call-site arithmetic: `engine.go` 64, 105, 259, 325, 332, 339, 359, 401, 413,
443, 538, 547, 588, 658, 772, 823; `staging.go` 26, 33, 56, 141, 161, 218, 244.

## 2. Namespace validation is reached on every one of those passes

```
grep -n 'validateIndependentTargetNamespaces' internal/transaction/*.go | grep -v _test.go
  journal.go:344  validateIndependentTargetNamespaces(journal.Targets)
  journal.go:354  validateIndependentTargetNamespaces(journal.Targets, targetNamespacePath{... engine.journalRoot})
  namespace.go:87 (definition)
```

`(*Engine).validateJournal` (journal.go:350) calls the package-level
`validateJournal` — whose last act (journal.go:344) is the bare target-graph
sweep — and then repeats the sweep with the manager's journal root added as a
reserved namespace. Both sweeps are unconditional. This is unchanged from the
baseline: **no call site was made conditional, skippable, or cached.** The
prototype deliberately leaves the two-sweep structure alone; folding them into
one staged pass is a further 2x that belongs to a separate change, because it
moves error attribution between the two distinct journal errors.

## 3. There is no cross-save state

```
sed -n '/^type Engine struct/,/^}/p' internal/transaction/engine.go
  mu sync.Mutex; journalRoot string; hooks Hooks; syncStagedParent func(string) error
```

`Engine` is byte-identical to the baseline: no verdict field, no digest, no
graph key, no mutex for one. `internal/transaction/engine.go` and
`internal/transaction/journal.go` are byte-identical to
`.temp/TASK-260729-365r5r/worktree-baseline`; the whole product delta is
`internal/transaction/namespace.go`.

```
grep -n '^var' internal/transaction/namespace.go    # (empty — no package-level state)
```

Every `resolvedNamespacePath` is created by `resolveNamespacePath` and stored
only in the `paths` slice that `validateIndependentTargetNamespaces` allocates
at namespace.go:88. The slice is a local, it is never returned, never stored on
`Engine`, and never escapes the call. When the function returns, the whole
snapshot is garbage. A second `saveJournal` therefore resolves and re-reads the
filesystem from scratch — which is what
`TestSaveJournalRejectsNamespaceAliasIntroducedBetweenSaves` and
`TestNamespaceIdentitySnapshotDoesNotOutliveItsPass` assert behaviorally.

## 4. At most O(P) filesystem identity reads per pass

```
grep -rn 'namespaceIdentity(' internal/
  namespace.go:81  (the single call)     namespace.go:257 (definition)
```

`namespaceIdentity` — the only `os.Stat` / `os.Lstat` in the sweep — has exactly
one caller, `(*resolvedNamespacePath).identity` (namespace.go:79), and that call
sits behind the `identityRead` guard:

```go
func (resolved *resolvedNamespacePath) identity() (os.FileInfo, error) {
	if !resolved.identityRead {
		resolved.identityInfo, resolved.identityErr = namespaceIdentity(resolved.targetNamespacePath)
		resolved.identityRead = true
	}
	return resolved.identityInfo, resolved.identityErr
}
```

`identityRead` is set on the same receiver whose fields are written, and the
receiver is a pointer into `paths`. The pairwise loop takes `&paths[leftIndex]`
and `&paths[rightIndex]`, so the guard is shared by every pair a path takes part
in. The number of `namespaceIdentity` calls in one pass is therefore at most
`len(paths)` = P, down from the baseline's 2 per surviving pair — up to
`P*(P-1)` for a fully disjoint graph, which is the common case, because a valid
graph is exactly the one where no pair short-circuits.

The remaining per-pass filesystem work is unchanged and already O(P): one
`canonicalNamespaceTargetPath` (an `EvalSymlinks` walk), one
`namespaceCaseInsensitive` and one `namespaceNormalizationInsensitive` per
declared path, all in the build loop at namespace.go:89-137, none of them inside
the pairwise sweep. Tomb paths still inherit their parent candidate's case and
normalization flags rather than re-probing, exactly as before.

## 5. The pairwise comparison itself is unchanged, only pre-computed

`namespaceContains` used to call `namespaceComponents(key)` for both sides of
every pair, and `namespaceComponentEqual` used to apply `norm.NFD.String` to
both components of every comparison. Both are pure functions of a single path,
so `resolveNamespacePath` now computes them once per path:

- `volume`, `parts` = `namespaceComponents(candidate.key)` — identical input
- `volumeNFD`, `partsNFD` = `norm.NFD.String` applied per component — the
  baseline normalized per component too, so pre-normalizing asks the same
  question of the same bytes

Case folding stays inside `namespaceComponentEqual` because it depends on the
**pair** (`left.caseInsensitive || right.caseInsensitive`), not on the path.
`normInsensitive` likewise still selects between the raw and NFD splits at
comparison time, so a pair where neither side sits on a normalization-insensitive
filesystem still compares raw components.

## 6. Error precedence and short-circuit order are preserved

`resolveNamespacePath` deliberately does **not** read the filesystem. The
baseline only asks for a path's identity once a pair has survived the
containment test, so an eager read would surface an inspection failure for a
path that a containment overlap would have rejected first. Keeping the read lazy
means the first `namespaceIdentity` call for a given path happens at exactly the
pair the baseline would have made it at, with the same error, in the same order
(`left` checked before `right`, `os.IsNotExist` tolerated, anything else
returned as `inspect <owner> <kind> path: ...`).

Behavioral confirmation is **prepared but NOT YET RUN**. `equivcheck/` is a third
copy of the tree whose `internal/transaction/namespace.go` is byte-identical to
`worktree-baseline` (verified: `diff -q` exit 0) carrying an adapted
`namespace_pass_test.go`. The adaptation drops exactly the two white-box tests
that name prototype-only symbols — `TestNamespaceIdentityIsReadOnceWithinOne`
`ValidationPass` and `TestNamespaceIdentitySnapshotDoesNotOutliveItsPass`, which
cannot compile against the baseline — and keeps all five fail-closed
behavioral tests plus the benchmark unchanged. Running them green against
baseline product code is what would show the prototype neither adds nor removes
a rejection.

That gate is `gate-equivalence` in `bin/run-gates.sh`. It has no `.exit` file, so
per this task's evidence protocol it has not passed. See
`TASK-260729-365r5r_results.md` §"Not run" for the full list of outstanding gates.
