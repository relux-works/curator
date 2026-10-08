# THE ONLY CURRENT INSTRUCTION — TASK-261004-3s6ymq: republish the smoke record (researcher)
Your revision 1 (one new file, `.research/261004_launcher_rc3_compat_smoke.md`, the launcher v0.2.0 / curator rc.3 compatibility smoke record) was reviewed as ACCEPTED in substance (see `TASK-261004-3s6ymq_review-verdict.md`), but the board could not record the acceptance: the Story workspace was converged onto today's main and the Change Request record went stale, so reviewers were handed "revision 0". Republish the same content.
**Do, in the workspace as it is now (do NOT checkout, reset or converge):**
1. Apply the attached `TASK-261004-3s6ymq_change-request_rev1.patch` (`git apply` of one added file). Do not edit the record's content.
2. Confirm the file is byte-identical to the resource `TASK-261004-3s6ymq_smoke.md`.
3. No tests, no builds (R193/R194/R223). Results resource: one line stating the re-application.
Then `task-board handoff TASK-261004-3s6ymq --role researcher` and END YOUR TURN.
