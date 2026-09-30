# Review note — TASK-260930-12i5zr accept the content-hash-v2 candidate suite (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev1 (base cbe52078, tree 17bfe7dd, 11 paths, gate green on rc.13) against `hv2suite-brief.md`. This commit will be pinned by
curator-spec PR #116's Implementations job, so it must pass against BOTH suites without weakening any ratchet. Verify:
1. Exact counts selected by suite identity (manifest digest). Check that:
   - no exact count became a lower bound;
   - an unknown manifest digest fails closed (it does not fall back to one table);
   - the rc.13 table is byte-for-byte what main had.

   Re-run the count mutant yourself: bump one pinned count, and the ratchet must fail under the matching root. Give the real exit
   code.
2. The frozen-shape negatives (install-marker-v4, context-lock-v1, agent-environment-marker-v2, and audit-record-v1 if present) are
   DRIVEN at the production readers. Check them in the candidate clone.
3. The 87 candidate-only gap rows:
   - they are owned by TASK-260917-2tx81l;
   - they activate only under the candidate digest;
   - under rc.13 nothing is listed and nothing is skipped silently.
4. Clone curator-spec at 526a9aa0 and run the brief's package set with CURATOR_CONFORMANCE_ROOT pointed at both roots. Give the real exit
   codes. The producer reported interrupted crossconformance runs: run `go test ./internal/crossconformance` against the candidate
   (bounded, split with -run if needed) and report.
5. No production behaviour change beyond conformance consumers; no CHANGELOG/LOGBOOK.
accept_cr, or changes requested with file:line. Never spell any employer name.
