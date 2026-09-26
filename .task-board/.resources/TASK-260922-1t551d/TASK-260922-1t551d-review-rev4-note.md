# Review note for TASK-260922-1t551d revision 4 (orchestrator, binding) — F-C1

Revision 4 = your revision-1 verdict's rework (TASK-260922-1t551d_review-verdict-rev1.md) plus two
gate-driven fixes: rework 1 (R1 liveness of the link target without reading bytes; R2 real TOML
parsing of `cli_auth_credentials_store` covering quoted spellings and tables, absent ⇒ file,
keyring/auto ⇒ `environment_isolated_unsupported`, other ⇒ `environment_credential_unsupported`,
invalid/unreadable TOML ⇒ fail closed); rework 2 (orchestrator ruling refining R2: MIS-TARGETED =
recorded target ≠ the strategy's declared native path ⇒ detached + `environment_credential_conflict`
+ fix-first relink of a recorded symlink; DANGLING-TO-DECLARED = target == declared path but absent
(pre-login) ⇒ provisioning/repair succeed, `status`/`resolve` surface a detached-pending finding
("log in to <tool>"), never `environment_repair_failed`; EACCES/inspection failure ⇒ conflict);
rework 3 (Windows skip reason in the ledger vocabulary). Gate green on all lanes: run 35690797842 —
verify the gate commit resolves to the exact revision-4 tree.

Judge with your own reruns (disposable clone; temporary stores; bounded commands):
1. Rerun your two probes (`TestReviewerDanglingExpectedTarget`, `TestReviewerLiteralCodexSelector`)
   against rev4 — the dangling expected-target case must now be REPORTED (finding), the three
   literal selectors must refuse with the right classes; confirm the two rows and the narrowing
   mutant that distinguish mis-targeted from dangling-to-declared; the operator's Pi case
   (`~/.pi/auth.json` recorded vs `~/.pi/agent/auth.json` declared) is mis-targeted ⇒ conflict.
2. The pre-existing Linux rows that rev2 broke (`TestResolveClaudeProjectEntry`,
   `TestClaudeSeedMergePreservesToolState`) pass because provisioning to the declared path
   succeeds while status reports pending — check the wording and that no secret bytes are read.
3. TOML parser: which library, how absence vs invalid vs unreadable are distinguished, rows for
   `'keyring'`, `"auto"`, multi-line strings, key inside `[table]`; mutants per refusal class.
4. Nothing accepted at rev1 regressed (stale-link removal, regular-file refusal, Pi agent root,
   marker untouched, no credential reads); ledger rows; CHANGELOG/troubleshooting claims accurate.
Record exactly one verdict: accept_cr(TASK-260922-1t551d, revision=4, evidence=<your outcome
resource>) on ACCEPT, or changes_requested with file:line and reproduction. Do not write into the
control root's LOGBOOK.md.
