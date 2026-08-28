# Rework required after reviewer RUN-260822-d8c978

Implement only the two remaining findings from `TASK-260811-3twayo_reviewer-verdict_RUN-260822-d8c978.md`; preserve every closed portion.

1. Make Node output reconciliation prove the exact C5 plan derived from the supplied C4 active graph before trusting `DeclaredOutputNodeIDs`. Either consume already validated publication authority or independently call the canonical derivation and compare exact plan identity. Add fail-closed negatives for forged zero-output, subset-output, action-set, and ordering substitutions while preserving genuine selected/pruned output cases.
2. Strengthen the independent Python P01-P13 protocol oracle. P10 must represent separately bound target closures with distinct canonical graph/outcome identities and reject cross-target reuse. Both Go and Python must derive and validate real canonical graph/binding/diagnostic records from fixture inputs, and independently reject missing or unknown fields in nested `lock`, `artifact`, and `build` schemas. Do not merely hash shared expected objects.

Do not regress canonical root/target ordering, exact C0 tool-record binding, mandatory executable SHA, graph-bound output declarations, generator chaining/cycles, lifecycle/native rejection, or selected/pruned/feature/peer/runtime/manager coverage.

Run focused, forged-plan, race, vet, lint, repository compile/build, canonical verifier, Python oracle, diff, and board validation gates, plus the full uncached repository suite unless a deterministic acceptance failure makes it irrelevant. Attach task-scoped evidence and use the developer handoff.
