# TASK-260729-365r5r — review verdict, cycle 2

Verdict: BLOCKED pending an explicit scope-approval decision. The prototype is not accepted in this cycle.

## Evidence reviewed

- The recorded rfrdfo baseline, literal allowlist, path-sorted manifests, static call-path proof, current two-file patch, raw exit files, and excluded-artifact ledger were inspected read-only. No Go, lint, test, build, benchmark, atomicity, install, or detached command was run.
- The current worktree exactly matches evidence/manifest-post.txt (recomputed manifest diff exit 0). The patch hashes 3dbbbfbd06d586442bd08166142934763b217efae42f85506d9c6a258c4c50d2 and git apply --check against the never-edited twin exits 0.
- Product code is unchanged from the previously reviewed prototype: internal/transaction/namespace.go hashes bb332038c5bf41a4043f6c3f799ea3ab530b9beeac9b5688fed8d1ad0edc56be. The accepted fail-closed call paths, per-pass O(P) identity-read design, between-save revalidation coverage, and preserved 66s non-race plus 84/76/75s race atomicity evidence remain sound.
- Diffing prototype-prerework.patch against prototype.patch proves the lint cycle changed exactly four closure parameters from t *testing.T to _ *testing.T: nested live paths, repeated live path, cross-target backup sidecar, and cleanup tomb. The attached constraint and cycle-2 reviewer directive authorize exactly the first three and forbid any other source change.
- The conflict is real, not hypothetical. After exactly the three authorized renames, gate-rw-lint-transaction exited 1 on namespace_pass_test.go:144, the cleanup-tomb closure. After the fourth rename, gate-rw2-transaction=0/14s, race-transaction=0/19s, namespace-verbose=0 with 25 PASS and 0 SKIP, equivalence=0/1s, and transaction lint=0/2s. Full lint exits 1 only for the inherited godriver ineffassign; that file is byte-identical to baseline (cmp exit 0).
- The results artifact also ends by claiming no equivcheck change, while its earlier gate ledger and rework artifact correctly record four matching renames in the adapted equivcheck test copy. That sentence needs correction but is not the stop boundary.

## Failed assumption and attempted paths

The lint-only constraint assumed the capped previous report listed every introduced revive finding. It did not: max-same-issues limited the report to three. Enforcing the literal three-renames-only boundary leaves an introduced transaction lint failure. Applying the fourth identical rename clears lint but exceeds the explicit source allowlist. No autonomous rework state satisfies both requirements.

## Viable alternatives

1. Approve a narrow constraint amendment allowing the fourth cleanup-tomb t-to-_ rename in internal/transaction/namespace_pass_test.go and the matching adapted equivcheck copy. This preserves the current green scoped gates and the unchanged product/performance evidence. Recommended.
2. Enforce the original exact-three boundary and revert the fourth rename. This restores literal scope compliance but necessarily leaves transaction lint red, so delivery cannot be accepted.
3. Reject or close the prototype. This avoids amending scope but gives up the validated 5.7x-or-better margin under the 480-second race bar.

## Exact decision needed

A human/orchestrator must explicitly approve or reject the fourth cleanup-tomb parameter rename, including whether the matching equivcheck adaptation and existing round-2 gate evidence are accepted under the amended scope. If approved, correct the contradictory no-equivcheck sentence and send the unchanged prototype through a new independent reviewer cycle.