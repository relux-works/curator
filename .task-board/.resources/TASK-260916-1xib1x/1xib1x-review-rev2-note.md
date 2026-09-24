# Review note for TASK-260916-1xib1x revision 2 (orchestrator, binding) — R4

Revision 2 = rework 1 for your revision-1 verdict (TASK-260916-1xib1x_review-verdict-rev1.md; brief
1xib1x-rework-1.md): (1) per-command audit identity — every script command recorded with the exact
`script-worker-v1` or explicit absence, independent of warning eligibility, exposed in the
production audit record (CLI line(s) + stored verdict), mixed and no-warning skills tested, the bare
"audit clean" CLI assertion corrected; (2) global-install coverage — the four vector rows + the
declared-only-with-hosts negative row through `install.Global` with its default gate, warning on the
same message line as the exact label, no errors, legacy shims unchanged; your global-wiring mutant
(`Subject.Commands` → nil in global.go) must now be killed. Gate green: run 35704285880 — verify
the gate commit resolves to the exact revision-2 tree and that rev1→rev2 is this scope. Rerun your
5 mutants (expect 5/5 killed) and inspect the audit record shape for a mixed skill. Everything else
was verified at rev1. Record exactly one verdict: accept_cr(TASK-260916-1xib1x, revision=2,
evidence=<your outcome resource>) on ACCEPT, or changes_requested with file:line and reproduction.
Do not write into the control root's LOGBOOK.md.
