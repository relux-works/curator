# Revision 4 surface sweep
Candidate: b060ac7b7ccf6dc93322770fa710e78cee3814dc.

| Surface | Result | Attacks / bounds |
| --- | --- | --- |
| seed modes | held | Resolve with 0600, 0400, 0640, 0644 seeds; transformed TOML; private-seed narrowing mutant killed. targeted.log |
| directory modes | held | All five adapters created at 0700 via Resolve; targeted.log. |
| repair | held | Existing 0755 routes, read-only and dry-run refusal, content retention; Codex-specific detection and protection weakening killed. targeted.log |
| other managed writes | held | Resolve Claude trust/context/marker checks; trust narrowing killed. Fallback writer tests and narrowing controls pass; forcing OS symlink failure through Resolve remains a stated bound. targeted.log |
| refusals | held | Resolve rejects symlink, unreadable file, FIFO without publishing home; valid seed controls pass. targeted.log |
| platforms | held | Local macOS behavioral suite; exact-tree hosted macOS/Linux/Windows success independently queried. POSIX tests Unix-only; Windows DACL backend unchanged. hosted.json |
| tests | held | Six narrowing overlays each require named behavioral failure and exit 1; no delete-only or static-only control. targeted.log |
| hosted gate | held | Run 39182785617 success; head 96b3684539d3228ff26e4b6d2df076d2da651069, tree and signature independently checked. Two conditional lanes skipped as recorded, not represented as run. |

## Verdict

```json
{"verdict":"accepted","revision":4,"findings":[],"notes":["The initial expanded 0777 fixture failed because the pre-existing trust boundary refuses group/other-writable roots before stale detection. A dedicated Resolve(repair=true) attack confirmed environment_store_untrusted and unchanged 0777 mode. This is a preserved security boundary, not evidence that arbitrary wider roots are repairable; the docs wording should be read with that boundary.","Full local suite was not rerun on revision 4. The latest r7 brief explicitly requires affected-path tests, build and vet locally and assigns the full configured gate to post-turn validation. Those producer checks plus the exact-tree full hosted gate satisfy that revised validation route; the original literal full-local AC5 is not claimed.","The gaps artifact still describes publication as a current blocker. That historical observation is superseded by the actual ready revision 4 and successful revision-4 validation record."]}
```

Repository delta is present and scoped to the five declared files. No source or index edits, commits, branch changes or direct control-root writes were made by this review. Go overlays only extend test fixtures under .temp; production blobs remain exact candidate blobs.

## AC coverage and evidence ownership

**4 of 5 AC rows driven by named behavioral tests**, with row 5 a validation requirement rather than product behavior. **5 of 5 assessed under the latest r7 validation instructions**; original full-local execution remains explicitly unclaimed.

1. `TestManagedSeedPreservesNativePermissions` -> `Resolve/gatherSeeds/applyPlan/atomicManagedFile` (source modes and transformed Codex config).
2. `TestManagedHomeRepairTightensDirectories` -> `Resolve/verifyHomeObserved/applyPlan/managedDirectory` (creation, stale detection, dry-run, repair, contents).
3. `TestManagedAuthoredFilesArePrivate` -> `Resolve/claudeSeed/applyPlan`; `TestManagedFallbackPermissions` -> real fallback writer. Fallback dispatch under an OS-induced symlink failure is not newly proven.
4. `TestManagedSeedRefusalsPreserved` -> `Resolve/gatherSeeds`; `TestManagedModesNarrowingMutants` -> six behavioral controls, including conditional 0600-to-0644 widening. Each requires exit 1, the named failed test, and the permission-specific reason.
5. Producer r7 affected-path/build/vet evidence reused from `BUG-261018-dbeghc_publication-r7-evidence.json`; full hosted run independently queried with `gh run view 39182785617 --repo relux-works/curator --json headSha,conclusion,jobs`. Commit signature G and candidate tree equality checked with git. Hosted URL: https://github.com/relux-works/curator/actions/runs/39182785617 . Conditional skipped jobs are not claimed as executed.

## Reviewer commands and results

All log names below are keys in `BUG-261018-dbeghc_review-evidence-rev4.json`.

- `go test -count=1 -timeout=4m -v -run '^TestManaged(SeedPreservesNativePermissions|HomeRepairTightensDirectories|AuthoredFilesArePrivate|SeedRefusalsPreserved|FallbackPermissions|ModesNarrowingMutants)$' ./internal/envprofile`: exit 0, targeted.log. All six narrowing controls killed for their intended reasons.
- `go test -count=1 -timeout=3m -v -run '^Test(ResolveLockContention|ResolvePublishesCredentialMarkerThroughLockedJournal|RepairNeverRefreshesSeeds|ClaudeRepairSatisfiedIsNoOp|ClaudeRepairRefusesUnsafeState|ResolveRenderedDocumentReplacesStoreSymlinkWithoutFollowing)$' ./internal/envprofile`: exit 0, neighbors.log. Includes the incoming Claude no-op behavior, locked journal, and unsafe-state refusals.
- Free hunt: overlay of the candidate test file only, extending seed modes with 0700, 0740, 0444, 0755: all eight source modes pass. Initial directory extension to 0777 exits 1 at the existing trust boundary (expanded.log), not a product permission leak. The complete initial test source is retained as initial-expanded-test.txt.
- Overlay directory extension to 0751, `go test -overlay <overlay.json> -count=1 -timeout=2m -v -run '^TestManagedHomeRepairTightensDirectories$' ./internal/envprofile`: exit 0, expanded-repair.log.
- Overlay `TestReviewWritableRootRefused` through Resolve repair, same command with `-run '^TestReviewWritableRootRefused$'`: exit 0, writable-root.log. Exact added test is in expanded_test.go. To reproduce overlays, map the absolute candidate path internal/envprofile/managed_modes_unix_test.go to the saved test content with Go's Replace overlay map.
- `git diff --check`: exit 0. All five changed worktree files compared byte-for-byte to tree b060ac7b7ccf6dc93322770fa710e78cee3814dc.

Free hunt produced no blocking findings. Held means the named attacks did not reproduce, not proof of absence. Concurrent source replacement/chmod, ACL preservation beyond POSIX bits, and forced filesystem fallback dispatch are not established by these tests. Local platform macOS arm64; Linux and Windows evidence is reused from the independently verified exact-tree hosted run, not represented as local execution. The run is not goal-bound, confirmed again before verdict.

Logbook outcome: review accepts revision 4 after the converged Claude implementation, checks exact signed hosted identity, and records the pre-existing writable-root boundary and stale publication note. Acceptance is not landing.
