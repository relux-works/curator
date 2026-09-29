# TASK-260910-32gki6 — review verdict, CR rev5 (reviewer: claude-opus-5-5 low)

**Verdict: ACCEPTED**

- Candidate: base 213a53e5, tree 0fd98b72 (I checked the worktree by writing it into a temporary index: `git write-tree` = 0fd98b72dc6171c210356ab5dff1ea70b6622c40). Tests ran in a disposable rsync copy under $TMPDIR, with CURATOR_CONFORMANCE_ROOT set to curator-spec 23435129 (rc.13) conformance/v1.
- Reference: environments §4 "protected state" and the §9 env resolve steps.

## Rules → production entry
- **Enclosing boundary.**
  - `Resolve` (managed.go:2329) → `validateResolveRoots`. This checks exactly `EnvRoot(home)` and `contextstore.Root(home)` through `pathboundary.ValidateLeafWithOwner`, not the Curator home or its ancestors.
  - Lock: `loadResolveInputsWithOwner`. Marker: `validateMarkerBoundary`.
  - Reads go through `stateread.Lstat`, so absence and read failure are separate cases. Every failure is returned directly and nothing is rebuilt.
- **Entry class.**
  - `validateNamedStoreBoundaries` (managed.go:2343) and `validateNamedStorePins` (managed.go:2373).
  - Under repair: `rebuildUntrustedStoreEntry` takes the op lock, re-checks the roots, lock hash, marker and the other entries, extracts a snapshot, then stages, protects, validates and renames.
  - Under dry-run: `would_rebuild`.
  - Path and local pins with no git snapshot → repair_failed.
- **Pin hash.**
  - For a git entry: `gitops.TreeObjectIDFromDir(entry)` is compared with `gitops.PinnedCommitTree`, read from the local repo object DB only (no fetch).
  - A missing commit object → `environment_store_untrusted` "pinned commit object unavailable", and it is never repaired (managed.go:2374).
  - For path/local entries: the `state_sha256` from ContentHash. A missing hash is an error.
- **Five checks.** All of them go through the shared `internal/pathboundary` walker (ownership, permissions/DACL, containment, regular types, link safety); there is no second walker.
- **Windows.** The `privatedir` Protect/ProtectTree owner-only DACL is applied at creation and at the rebuild publish. The DACL check is not relaxed, and `skip-classes.tsv` / `platform-cases.tsv` are untouched, so there are no new skips.

## Vectors (rc.13 environments-store-boundary.json, production entries Resolve / Resolve repair / dry-run / Status)
| Family | Result |
|---|---|
| resolve_cases | 20 driven, 0 gap, 0 bound |
| dry_run_cases | 4 |
| repair_cases | 10 |
| status_cases | 4 |

These counts match `conformance-case-counts.tsv`. `conformance-gaps.tsv` has no rows owned by STORY-260910-148pj1 or this task, so before/after is 0/0 and the file is unchanged.

## Reviewer runs (zsh, `set -o pipefail`, real exit codes)
- `go vet` (envprofile, pathboundary, privatedir, gitops) → 0
- `GOOS=windows go vet` (envprofile, privatedir, pathboundary) → 0
- `go test ./internal/{pathboundary,privatedir,gitops,contextlock,contextstore} ./cmd/curator -run 'Env|Boundary|Tree|Protect|Private|Lock|Store'` → rc=0 (cmd/curator took 581 s)
- `go test ./internal/envprofile -run 'StoreBoundary|Reserved|Surfacing|Guarded|Pin|Takeover|Legacy'` → rc=0
  - Note: without CURATOR_CONFORMANCE_ROOT the vector tests SKIP. The mutant runs below set it.

## Mutants (`-run 'StoreBoundary|GitPinMissing'`, each applied to a fresh copy; unmutated baseline rc=0)
| Mutant | Exit code | Failing tests | Result |
|---|---|---|---|
| M1 pin check skipped in Resolve | 1 | 11 (swapped system-prompt / root-context, GitPinMissing …) | KILLED |
| M2 missing commit object treated as pass | 1 | 1 (TestResolveRejectsGitPinMissingFromLocalObjectDatabase) | KILLED |
| M3 permissions check removed in pathboundary.checkNode | 1 | 4 (wrong-permissions-untrusted, repair-local-entry-cannot-rebuild) | KILLED |
| M4 entry class escalated (rebuild refused) | 1 | 8 (repair-rebuilds-git-entry-from-snapshot …) | KILLED |
| M5/M6 enclosing class demoted (root error ignored in Resolve and in the rebuild re-check) | 0 | 0 | SURVIVED — equivalent by redundancy, see R1 |

## Residuals (non-blocking)
- **R1 (M5/M6).** Demoting the enclosing class is not observable at the vectors, because every entry check re-walks the store root.
  - Result: repair-enclosing-refuses-no-rebuild still refuses without mutating anything, and the refusal carries the untrusted diagnostic.
  - The class distinction therefore depends on defence in depth. The row could also assert that there is no `environment_repair_failed` or rebuild text, which would kill the demotion directly.
- **R2.** The upgrade consequence stated by the producer: an unprotected Windows root created by an older Curator is refused as class (a). The orchestrator decides whether a repair command is needed.

## Scope hygiene
- The 31 paths are product, test and `.github/ci` files only.
- No CHANGELOG or LOGBOOK change; no stray files.
- The hosted gate is green on every lane (rev5, per the orchestrator note).
