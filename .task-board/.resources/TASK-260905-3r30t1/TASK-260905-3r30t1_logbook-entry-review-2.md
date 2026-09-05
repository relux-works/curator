# Logbook — byte-exact snapshot acquisition, review cycle 2 (curator `bb14375a`)

**Date:** 2026-09-06 · **Task:** TASK-260905-3r30t1 · **Run:** RUN-260905-2d5b2b (reviewer) · **Verdict:** accepted

## Finding worth carrying forward

**The platform-path collision gate in `internal/gitops/planWrites` keys on the folded full path, so it misses
directory-level folds.** Two tree paths whose directory components fold together but whose basenames differ
(`Dir/x.txt` + `dir/y.txt`) are admitted, and on a case-folding filesystem both land in one physical directory.
Proved on one macOS host by extracting the same commit into a case-insensitive APFS destination and a
case-sensitive APFS volume created with `hdiutil`:

```
Dir/x.txt + dir/y.txt  ->  case-folding   [Dir/x.txt Dir/y.txt]        sha256:1f97a4cb…1b5e57d2
                           case-sensitive [Dir/x.txt dir/y.txt]        sha256:5419409210…4ac90ebc
```

So a snapshot content hash for one commit still depends on the destination filesystem. Not a regression —
the replaced `git archive` path had no collision detection at all, and `git checkout` merges the same way, so
curator matches git. Fix shape when someone picks it up: fold every path *prefix* in `planWrites` and refuse
when a folded prefix is already claimed by a different original, which subsumes the existing file-level rule.

## Non-obvious facts established (so nobody re-derives them)

- **git canonicalises tree modes at `ls-tree` time**: `100600`, `100664`, `100000` → `100644`; `100777` →
  `100755`. A crafted tree cannot smuggle an odd mode past a `100644`/`100755` allow-list, and written
  permission bits are deterministic regardless of what the tree object stores.
- **A stalled `cat-file` child is not the deadlock class.** `Extract` blocks for the child's lifetime and then
  returns the typed error with an empty destination. There is no independent I/O timeout anywhere in
  `internal/gitops` (`run()` has none either) and never was. Watchdog-style tests must distinguish the two or
  they will report a stall as a deadlock.
- **A Unicode NFD/NFC path pair bypasses the `strings.ToLower` pre-pass** but is still refused — by the
  mid-stream `Lstat` backstop in `writeBlobs`, which aborts cleanly (kill + bounded drain) with an 8 MiB blob
  queued. The `O_EXCL` open means the worst case is a refusal, never an overwrite.
- **Closure staging introduces no rename race.** `snapshotFor` has one call site, the package has no
  goroutines, and `ScratchRoot` is a per-invocation private dir (`private.dir("closure-")`), so the
  `ENOTEMPTY`-on-rename hazard of the new stage-and-rename shape is unreachable.

## What made the review conclusive

Mutants, not reading. Collapsing `writeBlobs.abort` back to `stdin.Close(); Wait()` reinstated the cycle-1
deadlock and was caught by two named committed tests plus an independent one; reverting `Extract` to
`git archive` failed the byte-exact vector under both `core.autocrlf` settings; narrowing the size bound, the
`.git` case comparison, and the mode allow-list each killed a named test. The producer's own mutant table also
declared a **surviving** mutant (M5, closure staging) rather than hiding it, and it reproduced exactly as
reported — honest evidence is worth noting when you see it.
