# TASK-260910-32gki6 results

## Implementation

Implemented S5 manager environment-resolution checks through the CLI and manager production paths. Resolve validates the environments root, profile-store root, lock file, marker, and every lock-named store entry for operator ownership, private mutation permissions (owner-only DACL on Windows), containment, regular-file types, and lstat link safety. Enclosing-root failures refuse resolution without a fragment or rebuild. Named store-entry failures report `environment_store_untrusted`; repair revalidates the lock and other entries, rebuilds the Git entry from its locally available pinned commit snapshot, protects the staged tree, and verifies the rebuilt pin before retrying resolve. An unreadable or malformed lock is never rebuilt. Missing local Git commit objects fail closed without fetching or network access.

Resolve recomputes each named store entry's tree hash from store bytes and compares it with the expected pin. Git pins are resolved to their tree object through Curator's local Git object database; path/local pins use `state_sha256`. Resolve also compares the marker's recorded `root-context` and `system-prompt` surface hashes with the surfaces generated from the verified lock, reports drift, and emits no fragment on failure.

Windows creation/publish paths now establish owner-only DACLs for manager roots, lock and marker files, store roots, and staged/rebuilt entries. Existing paths that do not meet the boundary are refused with a path diagnostic for out-of-band repair. Manager-owned state reads use `internal/stateread`; managed writes use the no-follow and private-file helpers.

## rc.13 basis and production rows

Pinned spec: `curator-spec v1.0.0-rc.13` tag `34b02e2b045ca091d15da98057181682c390dba1`, `protocol/environments.md` §4 lines 697–757 and §9 lines 2873–2912. §4 defines all five boundary checks, the enclosing-boundary/entry failure classes, no-rebuild rule for an unreadable lock, and verification order. §9 requires store-byte pin recomputation before home currency and marker surface-hash comparison. The CI root-artifact table now requires `vectors/environments-store-boundary.json`; its four count pins cover 38 rows total.

| Rule | Production entry and test | rc.13 rows and representative vectors |
|---|---|---|
| Ownership, private permissions, containment, regular file types, and lstat link safety | `curator env resolve` → `envprofile.Resolve` → `validateResolveRoots`, `validateLockBoundary`, `validateMarkerBoundary`, `validateNamedStoreBoundaries`; `TestStoreBoundaryResolveVectorsDriveResolve` | `resolve_cases`: 20/20. Includes `wrong-ownership-untrusted`, `wrong-permissions-untrusted`, `containment-escape-untrusted`, `non-regular-component-untrusted`, `symlinked-entry-root-untrusted`, `marker-symlink-untrusted`, and enclosing-root/lock rows. |
| Recomputed Git/state pin hash and local-object availability | `Resolve` → `validateNamedStorePins` → `storeEntryPinHashes`; `TestResolveVerifiesGitStoreEntryAgainstPinnedTree`, `TestResolveRejectsGitPinMissingFromLocalObjectDatabase` | Resolve vectors include swapped system-prompt/root-context bytes, swapped updated store, and unprovisioned swapped store. Dedicated tests cover matching pin, mutated store, and unavailable local commit object. |
| Recorded root-context/system-prompt surface hashes | `Resolve` → `verifyHome.checkSurfaces`; `TestResolveReportsRecordedContextSurfaceHashDrift` | Resolve rows include swapped surfaces and stale-marker ordering; both root-context and system-prompt hash-drift subtests refuse fragment emission. |
| Dry-run and revalidated repair behavior | `Resolve` with dry-run/repair → `dryRunStoreEntryFailure` / `rebuildUntrustedStoreEntry`; `TestStoreBoundaryDryRunVectorsDriveResolveRepairPlan`, `TestStoreBoundaryRepairVectorsDriveResolveRepair` | `dry_run_cases`: 4/4; `repair_cases`: 10/10. Includes enclosing-boundary no-rebuild, Git entry rebuild from snapshot, state-pin rebuild refusal, stale-marker repair, and reapply-before-trust refusal. |
| Status diagnostics | status production entry → `validateProfileStoreState`; `TestStoreBoundaryStatusVectorsDriveStatus` | `status_cases`: 4/4, including naming the failed boundary check and enclosing-boundary failure. |

All 38 pinned rows are driven through production entries; the count ratchet passed. `.github/ci/conformance-gaps.tsv` had 0 rows owned by STORY-260910-148pj1 or TASK-260910-32gki6 before this change and has 0 afterward (0 → 0; there were no owned gap rows to remove).

The CLI acceptance test `TestEnvResolveRejectsSwappedStoreEntry` proves that `curator env resolve` detects swapped store bytes and emits no trusted fragment.

## Mutant evidence

Each mutation was temporary and restored before final validation. Every mutant command below exited **1**, as required for killed mutants.

| Rule | Mutation and killed test evidence | Exit |
|---|---|---:|
| Pin verification | Skipped the resolve pin-check call; `go test ./internal/envprofile -run '^TestResolveVerifiesGitStoreEntryAgainstPinnedTree$' -count=1` failed because mutated store bytes were trusted. | 1 |
| Missing local object | Treated “pinned commit object unavailable” as success; `go test ./internal/envprofile -run '^TestResolveRejectsGitPinMissingFromLocalObjectDatabase$' -count=1` failed because resolve emitted a fragment. | 1 |
| Recorded surface hash | Removed the recorded-versus-generated surface hash comparison; `go test ./internal/envprofile -run '^TestResolveReportsRecordedContextSurfaceHashDrift$' -count=1` failed both root-context and system-prompt drift cases. | 1 |
| Ownership | Weakened the owner check; the `wrong-ownership-untrusted` resolve vector failed because the untrusted entry was accepted. | 1 |
| Private permissions | Weakened the Unix private-permission check; the `wrong-permissions-untrusted` resolve vector failed because the untrusted entry emitted a fragment. | 1 |
| Containment | Suppressed the escaping-link containment result; after tightening the vector assertion to require `failed containment check:`, `containment-escape-untrusted` failed with `link_safety` instead of the required `containment`. | 1 |
| Regular file types | Removed the special-file rejection; `non-regular-component-untrusted` failed because Resolve returned nil error. | 1 |
| lstat link safety | Suppressed the store-entry link check; `symlinked-entry-root-untrusted` failed with `regular_types` instead of the required `link_safety`. | 1 |

The containment mutant first exposed a false positive in the test's old bare-word diagnostic match (the temporary directory name contained “containment”). The test now matches the structured failing-check phrase, passes on restored code, and kills the mutant. Early containment mutant invocations with an incorrect vector-root path or a subtest filter exited 1 before exercising the intended vector; they are not counted as mutant evidence. The incorrect-root invocation named a missing vector file; the correct root is `curator-spec/conformance/v1`.

## Validation (final restored tree)

| Command | Exit | Result |
|---|---:|---|
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envprofile -run 'Reserved|Surfacing' -count=1` | 0 | passed (53.105s) |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envprofile -run 'Store|Boundary|Pin' -count=1` | 0 | passed (184.427s) |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envprofile -run 'Resolve|Guarded' -count=1` | 0 | passed (245.614s) |
| `go test ./internal/privatedir ./internal/contextlock ./internal/contextstore -count=1` | 0 | all three packages passed |
| `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./cmd/curator -run '^TestEnvResolveRejectsSwappedStoreEntry$' -count=1` | 0 | passed |
| `GOOS=windows go vet ./...` | 0 | passed cross-platform vet/type validation |
| `golangci-lint run ./internal/privatedir ./internal/contextlock ./internal/contextstore ./internal/pathboundary ./internal/gitops ./internal/envprofile ./cmd/curator` | 0 | 0 issues |
| `git diff --check` | 0 | clean |

A preliminary targeted Windows vet exited 1 on the unavailable `windows.READ_ATTRIBUTES` constant; it was corrected to `windows.FILE_READ_ATTRIBUTES`, the targeted Windows vet passed, and the full `GOOS=windows go vet ./...` above passed. A first vector command with `CURATOR_CONFORMANCE_ROOT` set one directory too high exited 1 because the vector file was not present there; all final vector commands above use the correct `conformance/v1` root. No Windows runtime lane was run locally; runtime platform coverage remains for the hosted gate.

No CHANGELOG.md or LOGBOOK.md edits were made. Findings and evidence are recorded here. No CHANGELOG entry was added to the repository; release-prep text follows.

## CHANGELOG entry (for release prep)

Harden manager environment resolve with protected store-boundary checks, store-byte pin-hash verification, owner-only Windows DACLs, and recorded surface-hash drift detection before fragment emission.

## Gate fix 3 (Windows checked roots, vs rev3 tree bafbeddb)

Diff vs bafbeddb (tracked + the untracked store_boundary files): pathboundary.go (+ValidateLeafWithOwner), store_boundary.go (2 call sites),
nofollow.go (removed 4 ValidateRoot calls), store_boundary_conformance_test.go (+1 test).

Checked roots (environments §4), now exactly:
- `<home>/environments` (EnvRoot) — full node check (owner, owner-only perms/DACL, type, lstat link safety); route home→root link/containment only.
- `<home>/contexts` (contextstore.Root, profile store root) — same.
- `<home>/profiles/<profile>/lock.json` — full check on the lock file node only; `<home>`, `<home>/profiles`, `<home>/profiles/<profile>` walked for link safety/containment only.
- markers: ValidateWithin(EnvRoot, marker); store entries: ValidateWithin(contexts, entry) — unchanged (manager-created, below checked roots).
- NOT checked any more: the Curator home and its ancestors, the profiles dir (operator directories with inherited ACEs).

Cause of rev3 failures: (1) `ValidateRouteWithOwner(home, root)` / `ValidateWithinWithOwner(home, lock)` ran checkNode on `<home>` itself (`…\001`);
(2) the `profile_use_partial` follow-ons: rev3 added `pathboundary.ValidateRoot` on every managed-write root/component in nofollow.go
(managedPath/managedDirectory), which runs on operator roots (Curator home, tool homes) → refused on Windows. Those four calls are removed;
the write helpers keep lstat link refusal and still create missing components with privatedir (owner-only DACL on Windows). The DACL check itself is unchanged.

New row: TestStoreBoundaryCheckedRootsExcludeCuratorHome (drives Resolve): foreign-owned home / profiles dir / profile dir → pass;
world-writable home → pass; foreign-owned environments root / store root → environment_store_untrusted ownership.

Mutants (zsh, real exit codes, `go test ./internal/envprofile -run TestStoreBoundaryCheckedRootsExcludeCuratorHome`):
- roots back to ValidateRouteWithOwner(home, root): rc=1 (killed); fixed tree rc=0
- lock back to ValidateWithinWithOwner(home, lock): rc=1 (killed)

Local runs (zsh, standalone, real exit codes; CURATOR_CONFORMANCE_ROOT unset locally so pinned-vector tests skip — hosted gate is the arbiter):
- go test ./internal/envprofile -run 'Reserved|Surfacing|Store|Boundary|Pin|Guarded' rc=0
- go test ./internal/envprofile -run 'Resolve' rc=0
- go test ./internal/envprofile -run 'Status|Switch|Use|Import' rc=0
- go test ./internal/pathboundary ./internal/contextstore ./internal/privatedir rc=0
- GOOS=windows go vet ./... rc=0; go vet ./internal/... ./cmd/... rc=0
Not run locally: full envprofile package unfiltered, Windows execution (no Windows host) — Windows is unverified until the hosted lane.

Upgrade consequence (unchanged from gatefix-2): an existing environments root / store root / lock / store entry created by an older Curator
on Windows without an owner-only DACL is refused per §4 class (a) with `environment_store_untrusted: <label> failed permissions check: <path>`;
an operator's Curator home is no longer a cause of refusal.

## Gate fix 4 (base tree 61f5b184) — last 3 Windows failures
No product change; DACL check not relaxed. `git diff 61f5b184` touches only test fixtures/helpers:
1. TestLegacyPathInstallIgnoresStoreOverlap: `<home>/contexts` is `contextstore.Root(home)` = the profile store root, manager-protected per environments §4 ("the profile store root … manager-created, manager-protected"). The test's intent is that the Skillfile store-overlap *admission rule* does not apply to profile path sources (§1), not that the protected-root check is skipped. Fixture now creates the store root via `privatedir.MakeAll` (as the manager does); the seed lives under it. Code unchanged.
2. TestStoreBoundaryCheckedRootsExcludeCuratorHome: new test helper `makeWorldWritableSingleDirectoryForTest` (windows: restores the original DACL of that one path by handle, no tree walk, so manager-created links under the home are neither followed nor refused; unix/other: delegates to the existing helper). Existing tree-protect helper unchanged for the other callers.
3. TestRepairTakeoverProvisionsManagedHome: fixture created `<home>/environments/acme/claude_code` with os.MkdirAll (unprotected environments root); now `privatedir.MakeAll`, matching production creation.

Validation (zsh, real exit codes):
- `GOOS=windows go vet ./internal/envprofile` → 0; `go vet ./internal/envprofile` → 0
- `go test ./internal/envprofile -run 'Legacy|Boundary|Takeover|Store' -count=1` → exit 0 (122.5s)
- `go test ./internal/envprofile -run 'Reserved|Surfacing|Pin|Resolve|Guarded' -count=1` → exit 0 (338.2s)
- Windows not run locally (darwin host): unverified until the hosted gate.


## Revision 6 — re-apply on 6a7deb11dc92229c0380d2ea64cf302b86a51383

Re-applied accepted rev5 (refs/campaign/148pj1-rev5-20260929, parent 213a53e5) to the fresh trunk. The filtered product diff contains the same 31 paths as rev5. Per-path added/removed-line comparison matched in 29 paths; two intentional differences are recorded below.

Conflict resolutions, preserving both sides:
- .github/ci/conformance-case-counts.tsv: kept trunk's security-posture/vectors row and rev5's four S5 rows (20 resolve, 4 dry-run, 10 repair, 4 status).
- internal/envprofile/managed.go: retained trunk's hardened-MCP lock check and rev5's store-pin verification and repair handling.
- internal/envprofile/status.go: retained trunk's Codex seed test seam and rev5's boundaryOwnerLookup / profileStoreErrors fields and propagation. gofmt aligned the merged fields.
- internal/envprofile/store_boundary_conformance_test.go: added TestResolveRepairKeepsEnclosingStoreRootFailureClass in response to the rev5 review's surviving M5/M6 class-demotion finding. It requires the enclosing profile-store diagnostic, forbids repair/dry-run rebuild diagnostics, checks no fragment, and checks manager state is unchanged. The attached rev5 verdict says ACCEPTED while documenting this residual; this test answers the review-round instruction without changing production behavior.

Diff audit: git diff --name-only origin/main -- . ':!.task-board' lists 31 paths and matches the accepted rev5 path set. The per-path comparison is identical for the other 29 paths. status.go's three formatting-only field-line differences are the merge of trunk and rev5 fields described above. The conformance test file has only the named review regression as an addition. conformance-gaps.tsv is byte-identical to origin/main; Story/task-owned gap rows: 0 before to 0 after. The initial unfiltered diff also showed five concurrent .task-board paths; those were excluded from the candidate diff and not edited.

The S5 vectors were read from curator-spec commit 23435129; the checkout's environments-store-boundary.json SHA-256 matched the pinned file. The vector counts remain resolve 20, dry-run 4, repair 10, status 4.

### Revision 6 validation (zsh; each command run as a standalone process)

| Command | Exit | Result |
|---|---:|---|
| go test ./internal/envprofile -run '^TestResolveRepairKeepsEnclosingStoreRootFailureClass$' -count=1 | 0 | regression passes |
| Narrow mutant via Go overlay: skip only validateOptionalProtectedRoot for profile store root; same named test | 1 | KILLED; it received environment_repair_failed: another store entry is untrusted instead of the required enclosing-boundary diagnostic |
| CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envprofile -run 'Store|Boundary|Pin' -count=1 | 0 | 209.598s; pinned S5 vectors included |
| CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envprofile -run 'Resolve' -count=1 | 0 | 194.915s |
| CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envprofile -run 'Status|Legacy|Guarded' -count=1 | 0 | 151.547s; absence-read guard included |
| go test ./internal/pathboundary ./internal/stateread | 0 | both packages passed |
| GOOS=windows go vet ./internal/envprofile ./internal/pathboundary | 0 | passed |
| CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./cmd/curator -run 'Env' -count=1 | 0 | 545.157s; includes TestEnvResolveRejectsSwappedStoreEntry |
| golangci-lint run ./internal/envprofile ./internal/pathboundary ./internal/privatedir ./internal/gitops ./internal/contextlock ./internal/contextstore ./cmd/curator | 0 | 0 issues |
| bash .github/ci/ledger-consistency.sh "$TMPDIR/32gki6-ledger" | 0 | 468 ledger rows checked across linux, darwin, windows |
| git diff --check | 0 | clean |
| filtered 31-path / per-file delta comparison | 0 | 29 identical paths; two documented review/merge differences |
| diff <(git show origin/main:.github/ci/conformance-gaps.tsv) .github/ci/conformance-gaps.tsv | 0 | no diff |

Recovered command attempts: the first cmd/curator -run Env attempt exited 1 after 300.800s waiting for another live run's shared host-GOROOT lock; the lock holder exited, and the standalone retry above passed. The no-argument ledger script invocation exited 2 with its documented usage; supplying the required TMPDIR evidence directory passed. The first delta-comparison wrapper exited 1 from a Python syntax error; the corrected comparison above exited 0. Windows runtime execution was not available locally; hosted platform lanes remain authoritative.

No CHANGELOG.md or LOGBOOK.md edits. The release-prep CHANGELOG entry from the prior results is retained.

## Revision 7 — re-apply on fc499a96 (pathboundary.go merged with uyak0e)

Shell: zsh, each command standalone, real exit codes.

- `git diff 6a7deb11 refs/campaign/148pj1-rev6-20260929 -- . ':!.task-board' | git apply --3way`: 30 paths clean, conflict only in internal/pathboundary/pathboundary.go.
- `git diff --name-only origin/main -- . ':!.task-board'` = the 31 rev6 paths (`diff` against rev6 name list: identical, 31).
- For every path except pathboundary.go, the +/- lines of `git diff origin/main -- P` are byte-identical to `git diff 6a7deb11 rev6 -- P` (loop reported no DIFF).

### pathboundary.go resolution
- Kept uyak0e's test seam `validateWithOwnerAndEntryInfo(root, lookup, entryInfo)` (used by its pathboundary_test rows); it now delegates to `validateWithinWithOwnerAndEntryInfo(root, root, ...)`. `ValidateWithOwner` and S5's `ValidateWithinWithOwner` both use `defaultEntryInfo` (entry.Info()).
- S5 route walk (root → each component → target via os.Lstat) unchanged: any missing ROOT, component, or named TARGET (lock, marker, store entry) still fails closed (CheckRegular with the ENOENT cause). `ValidateRouteWithOwner` unchanged.
- `validateTree(root, target, operator, lookup, entryInfo)` carries uyak0e's `vanished` rule for every per-entry probe (walkErr, entryInfo, isLink, checkNode cause); the exemption is `path == target` (the walked, caller-named node), so only transient children discovered by the readdir walk may vanish (SkipDir for dirs). Failures report Root=root as in S5.
- Nothing else added.

### Focused runs
| command | exit |
|---|---|
| go test ./internal/pathboundary | 0 |
| go test ./internal/envprofile -run 'Store\|Boundary\|Pin\|Resolve\|Status\|Legacy\|PathInstall\|Guarded' | 0 (220s) |
| go test ./cmd/curator -run Env | 0 (418s) |
| GOOS=windows go vet ./internal/pathboundary ./internal/envprofile | 0 |
| go build ./... | 0 |

conformance-gaps.tsv untouched vs rev6 (rev6 delta applied identically). Windows runtime not executed locally; hosted gate is the arbiter. No CHANGELOG/LOGBOOK edits; CHANGELOG entry retained above.

## Revision 6 — carry-forward onto 450861c1 (no content change; rev5 failed only the Naming gate on a trunk board resource)

Bound run, no file changed. HEAD = origin/main = 450861c17b1bb69a1c9bd34a7beb4214fdda35e7 (zsh).
- `git diff --name-only origin/main -- . ':!.task-board' | wc -l` → 31
- `git diff | grep -c '^+<<<<<<<'` → 0 (grep exit 1 = no matches); unmerged entries in `git status --short` → 0
The prior Naming-gate failure came from a trunk board resource fixed on 450861c1 (jup8re); not caused by this delta. Hosted gate remains the arbiter.


### Carry-forward re-verification (2026-09-29T11:32:51Z)
Re-ran on HEAD = origin/main = 450861c1, no file changed (zsh): name-only count 31 (exit 0); `^+<<<<<<<` count 0; unmerged entries 0.

## Revision 8 — rev7 re-applied unchanged on 0a6382883557793a23a2719dfc6eedcf6030a39a

- `git diff fc499a96 refs/campaign/148pj1-rev7-20260929 -- . ':!.task-board' > $TMPDIR/s5.patch && git apply $TMPDIR/s5.patch` → exit 0 (clean).
- The three new files (store_boundary.go, store_boundary_conformance_test.go, internal/gitops/treehash.go) were marked `git add -N` (intent-to-add, no content staged) so `git diff` sees them; nothing committed.
- `git diff --name-only origin/main -- . ':!.task-board' | wc -l` → `31`.
- The +/- line multiset diff of rev7 vs the working tree → empty output, exit 0.
- Conflict markers `git diff | grep -c '^+<<<<<<<'` → `0`.
- No content change, so I did not rerun any tests in this revision. Earlier revisions' evidence stands, and the hosted gate decides. No CHANGELOG/LOGBOOK edit.

## Revision 8 — rev7 re-applied unchanged on 0a638288 (worktree base; origin/main 934093b0 differs from it only in README.md, SECURITY.md)
- Delta already present in fresh worktree; no file changed in this run.
- `git diff --name-only HEAD -- . ':!.task-board' | wc -l` → 31 (vs origin/main: 33 = 31 + README.md, SECURITY.md trunk-side only).
- +/- multiset diff vs `git diff fc499a96 refs/campaign/148pj1-rev7-20260929`: empty, exit 0.
- Conflict markers: `git diff | grep -c '^+<<<<<<<'` → 0.
