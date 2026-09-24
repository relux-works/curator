# Review note for TASK-260916-1l44nd revision 7 (orchestrator, binding)

Revision 7 = rework 6 for your revision-6 verdict (TASK-260916-1l44nd_review-verdict-rev6.md; brief
1l44nd-rework-6.md): the single finding — ABI-1 write mask handled WRITE_FILE only. Expected: the
REMOVE_*/MAKE_* rights handled from ABI 1; only REFER gated at ABI 2, TRUNCATE at 3, IOCTL_DEV at 5
(masks 0x1ff2 / 0x3ff2 / 0x7ff2 / 0xfff2, exec-denial +0x1); directory-only rights off file grants;
writable roots unchanged; the tests that protected the defect corrected (`landlock_test.go`,
`preflight_test.go` ABI-1 branch requires denial); your ABI-1 test committed; comments and the
results.md ABI→mask table corrected. Gate green on all lanes: run 35693783982 — verify the gate
commit resolves to the exact revision-7 tree and that rev6→rev7 is exactly this scope. Rerun your
`TestReviewerABI1MutationRights` against rev7 (must pass) and the mask builder per ABI against the
`unix` constants. Everything else was verified at rev5/rev6 — do not re-open it unless rev7 changed
those bytes. Record exactly one verdict: accept_cr(TASK-260916-1l44nd, revision=7, evidence=<your
outcome resource>) on ACCEPT, or changes_requested with file:line and reproduction. Do not write
into the control root's LOGBOOK.md.
