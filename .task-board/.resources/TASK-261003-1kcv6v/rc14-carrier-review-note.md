# Review note — TASK-261003-1kcv6v carrier identity review (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Same-provider review (operator rule, tb-R164). The carrier re-applies ACCEPTED TASK-261002-2ipeqa rev4 (`refs/campaign/2ipeqa-rev4-20261003` = 92ae3c19, base 64345d71, pre-rewrite) onto the rewritten main f17ea733. CR: base f17ea733, tree bdfbfa2d, 16 paths, gate green. The orchestrator's check found all 16 paths blob-identical to 92ae3c19.

Verify:
1. Blob identity 16/16, independently with `git rev-parse <tree>:<path>` on both sides.
2. The CR base is f17ea733 and the candidate descends from it; no pre-rewrite commit is reachable from the candidate commit. Check that the gate commit 62446680's ancestry contains f17ea733 and not dfaa557f.
3. The candidate lane, curator CI run 37137617861 (workflow_dispatch with candidate_ref e3a88ced and sha 6f832d81), is green once complete. If it is still running, wait for it.
4. The substance is already accepted in the 2ipeqa rev3/rev4 verdicts; cite them.

accept_cr if all hold; otherwise changes requested.
