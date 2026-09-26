# Review note for TASK-260922-cww1ov revision 1 (orchestrator, binding) — F-C3, FINAL leaf of STORY-260922-1cenbr

Read the task brief `cww1ov-brief.md` (precondition), the producer's `results.md`, F-C1 results +
accepted rev4 verdict, and F-C2 results + accepted rev4 verdict. F-C1 (repairs, states, TOML
admission) and F-C2 (explicit migration with journal/recovery) are CHECKPOINTED on the Story
branch; this leaf adds the production-entry coverage and its CI registration, it does not
re-implement them. Gate: run 35716709852 succeeded — verify its head resolves to the exact
candidate tree before reusing it, and do not rerun the full landing suite.

What this verdict must decide, per the brief's scope:
1. the two 0017 hazards (stale recorded link across shared→isolated; unlink of a regular file at a
   link path refused with `environment_credential_conflict`, bytes intact);
2. the dangling native target case in BOTH states — mis-targeted (the operator's Pi case: conflict,
   relink) and dangling-to-declared (pre-login: pending finding, provisioning succeeds);
3. migration end-to-end at the production entry: inspect → plan printed before mutation → apply
   under the lock, plan required, drift refusal on links AND marker digest, journal + recovery
   (interrupted apply deterministic; unknown states refuse; foreign temp files preserved), no
   secret bytes copied (byte-identity assertion);
4. every repair refusal class with the codex admission table (absent `cli_auth_credentials_store`
   ⇒ effective `file`; keyring/auto; unknown; single/double-quoted TOML);
5. `resolve --repair` never migrates silently.

Judge the COVERAGE CLAIM, not the prose: for each row, does a committed test reach the named
production entry (Go `envprofile` API or the real `curator env …` CLI), and does the stated
narrowing mutant actually kill a named test? Rerun a representative subset yourself with real exit
codes (`-count=1`, bounded timeouts, state the shell), attack at least two mutants of your own
choosing (one on a refusal bound, one on the migration/recovery path), and check the platform-cases
ledger delta: every row registered with the must/skip shape per lane, Windows symlink rows using the
existing envprofile privilege pattern, NO new skip class. Report helper-direct tests, pin checks and
inspection as bounds, never as driven rows.

Record exactly one verdict: `accept_cr(TASK-260922-cww1ov, revision=1, evidence=<your outcome
resource>)` on ACCEPT, or a changes-requested verdict routed with `set_status` naming file:line and
an executable reproduction for each finding. Do not write into the control root's LOGBOOK.md.
