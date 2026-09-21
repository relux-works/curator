# BUG-260920-2d9gfv rework 1 (orchestrator, binding)

Verdict rev1: CHANGES_REQUESTED (BUG-260920-2d9gfv_review-verdict-rev1.md, RUN-260921-976df9).
Root cause confirmed; the one open item is R1 itself: your per-registry `checkStart` bound
still lets the refusal class depend on fetch duration whenever more than one trusted registry
is configured (the reviewer drove it at the production entry with two registries: the instant
second registry is refused "too far in the future", 3/3 under the candidate, 3/3 pass under the
function-entry shape). Orchestrator ruling: adopt the brief's preferred function-entry shape —
the tolerance is the checker's own latency since it sampled its clock, which is ≈ the
post-fetch wall clock + skew for production callers and ~zero for injected-`now` callers with
instant fetches. Continue from the revision-1 tree in the Story workspace (no checkout/clean/
stash). Do exactly verdict §7:

1. `internal/registry/snapshot.go`: `start := time.Now()` as the first statement of
   `checkSnapshotsWithPolicy`; compare `parsed.CreatedAt.After(now.Add(clockSkew).Add(time.Since(start)))`;
   rewrite the two comments accordingly (a genuinely future timestamp — ahead of the
   post-fetch clock + skew — is still refused).
2. `internal/registry/registry_test.go`: replace the "slow sibling does not widen" row by its
   inverse — after a 1.5 s sibling, an instant registry with `created_at = now + 1 s` (behind
   the wall clock) is ACCEPTED, and one with `created_at = now + skew + 3 s` (ahead of the
   post-fetch clock) is still refused with `too far in the future`; keep the other two rows;
   optional: pre-create the catalog in the exact-bound pins so first-use fsyncs do not sit
   inside their margins.
3. `internal/install/registry_e2e_test.go`: add the two-registry production-entry row (shape
   of the reviewer's probe `BUG-260920-2d9gfv_review-rev1-two-registry-probe_test.go.txt`: slow
   first registry crossing the boundary, instant serve-time-minted second registry, strict
   policy, zero skew → `ok`, no future warning) — it must fail on the rev1 tree and pass with
   (1); restore `crossSecondBoundary` to 2 ms or keep 50 ms with a corrected comment (F2);
   refresh the three stale comments (F3).
4. CHANGELOG: "per-registry latency" → the function-entry wording; results.md "Revision 2":
   the read-only-path bound (verdict §2), the mutant that now kills the sibling row (hoist vs
   per-registry), the two-registry row output.
5. Rerun the R3 determinism command on the reworked tree (record it) and publish revision 2
   only on a green gate.
