# Identity review (orchestrator, binding) — same tree, re-published with tree-bound evidence

You (claude-opus-5-5) already ACCEPTED this leaf's content at the previous revision; `accept_cr` was
refused only with `validation_not_bound_to_tree` (the Change Request had been published by the old
board binary). The producer re-published the SAME tree under the new binary, so the landing suite ran
again and its evidence is now bound to the tree.

Do exactly this:
1. Prove the new revision's candidate tree is byte-identical to the revision you accepted: compare the
   two `*_change-request_rev*.patch` resources (and the recorded candidate tree ids) — identical path set
   and identical content. Any difference is a finding.
2. Confirm the new revision's validation evidence is green and tree-bound (read its validation log).
3. Do NOT re-review the content; your previous verdict resource carries the content judgement — cite it.
4. Record exactly one verdict: `accept_cr(<TASK-ID>, revision=<new N>, evidence=<your outcome resource>)`
   on identity + green gate; otherwise changes-requested naming the difference. No LOGBOOK.md writes.
