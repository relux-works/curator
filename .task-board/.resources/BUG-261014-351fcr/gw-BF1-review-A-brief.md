# THE ONLY CURRENT INSTRUCTION — reviewer A of two (R138), RECORD-ONLY, for BUG-261014-351fcr (BF1)

**Sources:** the precondition `gw-BF1-brief.md`; curator-spec main `protocol/core.md` §8 and `protocol/environments.md` §9.1; this repository.

**Check:**
1. Both audit call sites in `migrateGlobalSkills` run under `hashing.WriteVersion()`, through the same selection the `global add --revision` path uses; no second version rule, no path exemption.
2. The migrated lock records hash version 2 under a v2 write version.
3. Tests would fail on regression: reverting either call site to V1, or dropping the lock hash version, fails a test; the forced-v1 fixture still refuses with the opaque-file finding.
4. The network-undeclared follow-up has a test and a precise record in the gaps file; it is not silently fixed or silently skipped.
5. No live-home access in code or tests; the remote-gate result is cited; no private identifiers or local paths.

RECORD-ONLY: attach `gw-BF1-review-A.md`. Do NOT call accept_cr or reject_cr. END YOUR TURN.

**Hard rule:** you are RECORD-ONLY; reviewer B decides.
