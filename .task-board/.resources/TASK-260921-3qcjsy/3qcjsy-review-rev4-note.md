# Review note for TASK-260921-3qcjsy revision 4 (orchestrator, binding) — adopt 0017/0018

Revision 4 = your revision-1 verdict's rework (F1 marker prose vs frozen schema, F2 codex_cli
`isolated` under `auto` fail-closed, F3 proposal-era sentences; N1–N4) re-applied unchanged on the
fresh trunk 8e65374c (BUG-260921-3cgij4 landed meanwhile: draft-sources-v1 fixtures, path-disjoint).
Revisions 2 and 3 carried the same bytes but their local gates were killed by the host (fresh
`go test` binary SIGKILLed); the configured gate now retries that step only on `signal: killed` and
rev4 is green. Verify the gate ran on the exact revision-4 tree and that rev4 − rev1 is exactly the
F1–F3/N1–N4 text delta (no schema/vector/generator change beyond rev1's accepted mechanics).

Judge: F1 — §8.2 schema-1 bullet back to strategy-only, §7.4 worded as the F-S1 marker-revision
content, 0017 Compatibility aligned; F2 — `isolated` for `codex_cli` under `file` only, `auto` joins
`keyring` in `environment_isolated_unsupported` in §7.4 paragraph+matrix, manager §12.4
paragraph+table, 0017 choice 4, results row 4; F3 — `grep -n "not an adoption\|adopting revision
amends\|stays open question" decisions/001[78]*.md` empty and each rewording correct; N1–N4 as
requested. Do NOT write outside your task-scoped resources: the LOGBOOK.md of the curator control
root is operator-owned (the rev1 review wrote 9 lines there; that write was discarded) — record
findings only in your verdict resource. Record exactly one verdict:
accept_cr(TASK-260921-3qcjsy, revision=4, evidence=<your outcome resource>) on ACCEPT, or
changes_requested with file:line and reproduction.
