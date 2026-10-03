# Review note — TASK-261002-2ipeqa rev4 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Same-provider delta review (operator rule, R138). The rev4 tree must equal rev3 tree 9706982a; check that first.

SCOPE IS RECONCILED. The orchestrator's earlier "only one gap row" wording was wrong; the binding scope is the full five-file delta you reviewed in rev3:
- the rc.14 gap row owned by TASK-261002-1foyf3;
- the snapshot RunOutcomes adapter;
- the expanded regressions;
- 12 lifecycle count rows, including 8 for the historical candidates;
- the CHANGELOG line.

Your rev3 sweep found no product defect, and the exact-tree evidence (candidate run 37017428885 on 3 OSes, the default matrix and the spec's exact Implementations argv) may be reused.

If the tree is identical and nothing else changed: accept_cr, citing your rev3 sweep. Otherwise: changes requested.
