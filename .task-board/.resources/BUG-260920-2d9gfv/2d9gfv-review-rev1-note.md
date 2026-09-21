# Review note for BUG-260920-2d9gfv revision 1 (orchestrator, binding)

Brief 2d9gfv-brief.md (rulings R1–R4). Producer's claim: root cause = `resolveRegistries`
samples `time.Now()` before the snapshot fetch, a serve-time-minted `created_at` truncated to
seconds can exceed it, and with the zero clock skew of the bare Go-API harness config the
future check flips the registry to "tampered". Fix at the product root in
`registry.checkSnapshotsWithPolicy`: per-registry `checkStart := time.Now()` and
`parsed.CreatedAt.After(now.Add(clockSkew).Add(time.Since(checkStart)))`. Gate green: run
35545031567 — verify the gate commit resolves to the exact revision-1 tree.

Judge, with your own reruns (disposable clone; bounded commands; retry once on host stalls):
1. Root cause: rerun the pre-fix proof yourself — revert the one-line fix in a clone and run
   `TestRegistrySnapshotMintedDuringFetchIsNotFuture`; it must fail with the gate's exact class
   and message. Confirm the driven row forces the boundary crossing deterministically (not a
   sleep-and-hope), and that the negative row (`…TwoSecondsPastSkewStillRefuses`) still refuses
   with the same class/text after the fix.
2. Semantics: `now` stays the reference time for injected-`now` callers (exact-bound pins);
   the per-registry start (not function entry) is pinned by the "slow sibling" row — check
   the row really distinguishes per-registry from function-entry (mutant: hoist `checkStart`
   before the loop → the row must fail). Stale check unchanged. Read-only status path shares
   the fix — is it exercised by a row?
3. Legacy v1 lane: this is shared code; the CHANGELOG `### Fixed` entry and a legacy-lane row
   (v1 install entry, zero skew, serve-time-minted snapshot) must exist per R2; legacy goldens
   green; no other behaviour change (diff of the non-test files).
4. Determinism evidence (R3): the producer's 20-consecutive `-race` run of the
   attestation-evidence rows — check the recorded command/timing; rerun a smaller count
   yourself (`-count=5`) on the candidate.
5. Harness: zero-skew Go-API config kept and documented (R3), no widening of skew, no fixed
   `created_at` in the stub. Note BUG-260906-1bdotx's earlier fixture-only fix — is anything
   from it now redundant or contradictory?
6. Mutants observed, ratio line, Windows proof status.
Record exactly one verdict: accept_cr(BUG-260920-2d9gfv, revision=1, evidence=<your outcome
resource>) on ACCEPT, or changes_requested with file:line and reproduction.
