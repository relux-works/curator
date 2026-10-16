# Review brief SNAPLOCK (deciding review B): SNAP-LOCK — Non-recovering capture transaction (TASK-261016-4zofp0)

Read gw-SNAPLOCK-brief.md, the attached leaf file(s) and rc7-code-common.md. Review the candidate against the leaf:
1. Every acceptance criterion is met by code and tests, traced to the cited normative clauses (read them on curator-spec main).
2. Every paired test row exists, asserts the intended refusal and the unchanged forbidden state, and has its positive control; an implementation that refuses everything must fail the positive controls.
3. The plant/checker control was executed and failed on the planted violation.
4. No change outside the leaf's paths and seams without a justified note; no claim of qualification that was not executed.
5. Re-run the focused tests and the repository gate yourself.
Verdict ACCEPT or REJECT with findings P1/P2/P3. Follow the board's review contract.

## Revision 3 (after convergence onto trunk)

Revision 2 was accepted, then went stale when trunk advanced on `README.md` and `internal/envprofile/managed.go`. Revision 3 is the same delta on current trunk, plus a negative test for the credential-store admission check on the capture path. Review the accepted-to-combined delta in full. The main question is whether trunk's `admitCredentialStore` and this revision's `resolveVerified`/`verifyHomeContent` split compose correctly: exactly one admission check per resolve, the same refusal on the capture path, and no capture write before the refusal. Do not reopen revision 2 passages that the merge did not touch.

## Revision 4 (re-applied onto trunk after the L15-1 policy projection)

Trunk moved the at-or-above and isolation computation into `resolveCredentialPolicy`, which observes the vendor tool and caches its result. Revision 4 re-applies this task's `managed.go` changes on top of that. Check above all: capture and status never reach `req.detected` through the shared policy path; the policy cache cannot serve a policy computed under a different release observation; there is still a single policy derivation path; and the new vendor-probe plant was executed and failed.
