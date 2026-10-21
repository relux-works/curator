# Review brief r4: skill publication matches the core §3.1 projection (BUG-261021-2qv4zj, curator)

This is a correctness re-review at the exact head in the request. Review against:
- `W-AUDITSCOPE-rework-r4-brief.md`;
- your `BUG-261021-2qv4zj_review-r3.md`;
- the producer's `BUG-261021-2qv4zj_results-r4.md`, which carries the publisher inventory.

The producer reports:
- native and managed profile skills now publish one shared §3.1 projection;
- Curator-owned raw links and copied homes are repaired without takeover;
- surface identity is versioned (`core-3.1-v1`);
- commands, runtime publication and audit ruleset 8 are unchanged.

| Surface | Check |
| --- | --- |
| F3 closure | Re-run your two r3 probes (native and managed) at the head. Excluded repository files, fixed excluded names, runtime/build roots and exported-script Markdown are absent from `skills/<name>/` for Claude and Codex, in both link and copy modes. Visible and declared controls are still published. |
| Inventory completeness | Check the publisher table against the code. Every path that places skill bytes on an agent surface must use the projection: native, managed, copy fallback, legacy hash-identity migration, status, repair, and project/global installs. A publisher still pointing at the raw root is a finding. |
| Projection identity | Reuse of a projection directory is verified against its versioned identity. A missing, extra or changed file makes it stale, and it is rebuilt rather than served. Two profiles sharing a member share one projection, and removing one profile keeps a projection the other still references. |
| Migration | A home with a raw-root link and a home with an old copied tree are both reported stale by status, which stays read-only. Repair converges them through the recoverable transaction, with no takeover prompt for Curator-owned entries. A foreign (non-Curator) link still requires explicit takeover. This is the ivmbp rc.8 swap path. |
| Runtime unaffected | Script commands and `runtime_roots` resolve exactly as before. No command depended on the raw-root link; test one exported-command package end to end. |
| Regressions and suite | F1 and F2 stay closed. The task-board-shaped package installs and none of its repository files appear on any surface. Re-run 2-3 of the new narrowing controls. Run the full default suite at the exact head with a private GOCACHE, and delete scratch afterwards. |

Give the verdict at the exact head. Attach `BUG-261021-2qv4zj_review-r4.md` as an outcome resource. Write probes and the report in neutral QA terms.
