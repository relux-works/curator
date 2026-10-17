# Review brief L151C (deciding review B): credential store policy, remaining workstreams (TASK-261016-wh6v9a)

Read gw-L151C-brief.md, the attached leaf and rc7-code-common.md. Review the candidate against the requirements the brief lists:
1. Each delivered requirement is met by code and tests, traced to the cited normative clauses (read them on curator-spec main). Any narrowing is stated honestly, and no deferred item is claimed.
2. The landed operation-local projection is reused; no second policy derivation path exists.
3. Every paired test row exists, asserts the intended refusal and the unchanged forbidden state, and has its positive control; an implementation that refuses everything must fail the positive controls.
4. The plant controls were executed and failed on the planted violation.
5. No change outside the leaf's paths and seams without a justified note.
6. Re-run the focused tests and the repository gate yourself.
Verdict ACCEPT or REJECT with findings P1/P2/P3. Follow the board's review contract.

## Revision 2 (after convergence onto trunk with SNAP-LOCK)

Revision 1 was accepted, then went stale when SNAP-LOCK landed on the same files. Review the accepted-to-combined delta. Check above all: the Muse policy code never observes the vendor tool on the capture or status path; there is still one policy derivation path; SNAP-LOCK's no-probe test and plants still pass on the combined tree. Do not reopen revision 1 passages that the merge did not touch.

## Revision 3

The only open item from revision 2 was the hosted Ubuntu Test failure, attributed to the Landlock probe recursion load (fixed on main 6e85a91; the Story was converged onto it). Verify the fresh full gate on the exact revision 3 candidate, the LOGBOOK.md placement, and that the delta from revision 2 is limited to the convergence and the logbook move unless the producer documents a reproduced defect.
