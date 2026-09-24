# TASK-260916-1l44nd rework 6 (orchestrator, binding)

Verdict rev6: CHANGES_REQUESTED with ONE finding (TASK-260916-1l44nd_review-verdict-rev6.md;
reviewer test attached: `TASK-260916-1l44nd_review-rev6-abi1_test.go`). F1/F2 of rev5 are fixed
and verified; rev5→rev6 scope is otherwise accepted. Continue from the revision-6 tree (no
checkout/clean/stash); fix exactly:

F1 (high) — `internal/scriptworker/landlock.go:71–75` adds the directory-only set only for ABI ≥ 2,
but REMOVE_DIR, REMOVE_FILE and all seven MAKE_* rights exist since ABI 1 (Linux 5.13); only REFER
starts at ABI 2, TRUNCATE at ABI 3, IOCTL_DEV at ABI 5. On an ABI-1 host the write mask is 0x2
instead of 0x1ff2 and the control still attests `applied`. Handle REMOVE_*/MAKE_* from ABI 1;
gate only REFER (ABI 2), TRUNCATE (ABI 3), IOCTL_DEV (ABI 5); directory-only rights stay off file
grants; writable roots unchanged. Expected write masks: ABI 1 = 0x1ff2; ABI 2 = 0x3ff2; ABI 3–4 =
0x7ff2; ABI 5+ = 0xfff2 (exec-denial adds 0x1). Fix the tests that currently PROTECT the defect
(`landlock_test.go:67–75` expects WRITE_FILE only on ABI 1; `preflight_test.go:436–467` expects
outside-grant mutation to succeed on ABI 1) — the production mutation row must require denial on
ABI 1 too; commit the reviewer's ABI-1 test; correct source comments and the results.md ABI→mask
table ("Revision 7"). Do not suppress the control on ABI 1. Republish only on a green gate.
