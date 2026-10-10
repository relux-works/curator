# THE ONLY CURRENT INSTRUCTION — delta review of Change Request rev2 of TASK-261010-aqpf2a (reviewer, read-only)

The rev1 review (`TASK-261010-aqpf2a_review-verdict-rev1.md`) accepted the content and required exactly one change: drop the `LOGBOOK.md` hunk and keep `.research/261010_pi-opencode-tool-lockdown.md` byte-identical.

Confirm, read-only:
1. The rev2 Change Request (`TASK-261010-aqpf2a_change-request_rev2.patch` and its candidate) changes exactly one path, `.research/261010_pi-opencode-tool-lockdown.md`. `LOGBOOK.md` is untouched.
2. That file's blob in rev2 equals its blob in rev1 (506 lines); compare the blob ids or hashes.
3. The rev2 validation log is green.

If all three hold, `accept_cr`. Otherwise request changes with the exact difference. In the verdict findings block, `severity` must be one of `bypass|regression|robustness|note`; put the P-level in `severity_reason`. Same-provider review (operator rule R138). Do not run tests or builds on this host. Then END YOUR TURN.
