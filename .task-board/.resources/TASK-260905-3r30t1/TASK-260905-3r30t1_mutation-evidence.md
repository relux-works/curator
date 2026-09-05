# TASK-260905-3r30t1 — narrowing-mutant and AC-coverage evidence (head a46abc80)

Run in a scratch copy of the curator worktree (`rsync` of `feat/byte-exact-acquisition` @ a46abc80 into this story
worktree's `.temp/TASK-260905-3r30t1/mut/`, `.git` excluded). The curator worktree itself was not edited. Each mutant
was applied to `internal/gitops/gitops.go`, `go test -count=1 -run <tests> ./internal/gitops` run directly, then the
file reverted (`diff -q` proved revert; clean `go test -count=1 ./internal/gitops` exit 0 afterwards).

## AC coverage: 9 of 9 AC rows driven through the production entry point (`gitops.Extract`, called from
`internal/snapshot/snapshot.go:48` and `internal/closure/closure.go:413`)

| AC row | Named committed test | Entry point |
|---|---|---|
| ls-tree -r -z + cat-file --batch extraction, exact bytes | TestExtractProducesExactTree, TestExtractReproducesByteExactVector | gitops.Extract |
| refuse symlinks (120000) | TestArchiveRejectsLinks | gitops.Extract |
| refuse gitlinks (160000) | TestExtractRefusesSubmodules | gitops.Extract |
| refuse escapes / empty names | TestExtractRefusesEscapingPaths | gitops.Extract |
| refuse oversize blobs | TestExtractRefusesOversizeBlob | gitops.Extract |
| refuse duplicate platform paths | TestExtractRefusesDuplicatePlatformPaths, TestExtractRefusesExistingDestinationEntries | gitops.Extract |
| preserve 100755 | TestExtractPreservesExecutableBit | gitops.Extract |
| vector sha256:500ea934…2bced0 under autocrlf true/false, $Format:%H$ intact | TestExtractReproducesByteExactVector; TestConformanceSnapshotAcquisition (CURATOR_CONFORMANCE_ROOT) | gitops.Extract |
| old git-archive behavior gone (negative) | TestExtractIgnoresWorkingTreeConversion | gitops.Extract |

Stated bound: callers are switched (grep shows the two call sites; no `git archive` invocation remains in
`internal/gitops` non-test code), but the snapshot/closure packages exercise Extract only through their existing
suites, not through a byte-exact assertion of their own.

## Narrowing mutants (gate stays present; admits exactly one member of its refused class)

| Mutant | Narrows the gate to | Failing named test | Survivor bound |
|---|---|---|---|
| M1 `120000` refused only when path contains `/` | admits top-level symlinks | TestArchiveRejectsLinks (exit 1) | — |
| M2 `160000`/commit refused only when path contains `/` | admits top-level gitlinks | none (exit 0) — SURVIVOR | the generic `kind != "blob"` clause still refuses it with the same "unsupported entry type" class; the submodule-specific wording is not asserted. Bound: submodule-specific branch is redundant defense, not independently proven. |
| M3 drop `component == "."` | admits `sub/./x` | TestExtractRefusesEscapingPaths (exit 1) | — |
| M3b drop `HasPrefix(name, "/")` | admits `/abs` | none (exit 0) — SURVIVOR | `strings.Split("/abs","/")` yields an empty first component, refused by the component check. Bound: leading-slash check is redundant defense. |
| M4 `> max` → `> max+1` | admits a blob exactly max+1 bytes | TestExtractRefusesOversizeBlob (exit 1) | — |
| M5 Lstat collision refused only for `100755` | admits 0644 collisions | TestExtractRefusesDuplicatePlatformPaths, TestExtractRefusesExistingDestinationEntries (exit 1) | — |
| M6 exec bit only for top-level paths | nested 100755 written as 0644 | none (exit 0) — SURVIVOR | no test commits a nested executable. Bound: exec-bit preservation is proven for top-level files only (fixture `lf.txt`). |
| M6b exec bit only for nested paths | top-level 100755 written as 0644 | TestExtractPreservesExecutableBit (exit 1) | — |

Prior delete-style mutants (drafting report) remain as existence evidence only.

## Source-text-inspecting gates (checklist item 19)

No gate in this change inspects source text: the refusals inspect `ls-tree` mode/type fields, tree paths, and
`cat-file --batch` size headers. The CI platform-case gate matches test *output* (skip reason) against
`skip-classes.tsv`; rework report 1 attacked it with a skip-reason mutant (`.temp/rework1/skipmutant.json`). Declared
not applicable — stated bound.

## Commands

```
go build ./...                                   # exit 0 (scratch copy)
go test -count=1 -run <mutant tests> ./internal/gitops   # per-mutant exit codes as tabled
go test -count=1 ./internal/gitops               # after revert, exit 0
```
