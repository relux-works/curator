# TASK-260922-cww1ov brief (orchestrator, binding) — F-C3 0017 production-entry tests (FINAL leaf)

Story STORY-260922-1cenbr; F-C1 (repairs, states, TOML admission) and F-C2 (explicit migration with
journal/recovery) are checkpointed on the Story branch. This is the Story's LAST leaf: on ACCEPT the
orchestrator integrates the Story (the runtime replays the checkpoints onto current trunk at your
spawn). Read F-C1/F-C2 results.md and every verdict (F-C1 rev1/rev4, F-C2 rev1/rev3/rev4): the
reviewer probes already committed there are your baseline — this leaf makes the coverage
COMPLETE and CI-registered, it does not re-implement.

## Scope (F-C3 per the adoption's follow-up table, amended by the operator corrections)
Production-entry test suite on temporary stores (Go API `envprofile` + CLI `curator env …`) that
covers, each with a narrowing mutant proving the bound:
1. the two 0017 hazards: (a) a stale recorded link surviving shared→isolated; (b) unlink of a
   regular file at a link path (must refuse with `environment_credential_conflict`, bytes intact);
2. the dangling native target case: a recorded link whose native target does not exist —
   BOTH states: mis-targeted (operator's Pi case → conflict, relink) and dangling-to-declared
   (pre-login → pending finding, provisioning succeeds);
3. the migration end-to-end: inspect → plan (printed before mutation) → apply under the lock,
   plan-required, drift refusal (links AND marker digest), journal + recovery (interrupted apply
   completes/restores deterministically; unknown states refuse; foreign temp files preserved),
   no secret bytes copied (byte-identity assertion);
4. every repair refusal class (`environment_credential_conflict`, `environment_credential_unsupported`,
   `environment_isolated_unsupported`, `environment_repair_failed` where still emitted) with the
   codex admission table (absent key ⇒ file; keyring/auto; unknown; single/double-quoted TOML);
5. `resolve --repair` never migrates silently.
Register every row in `.github/ci/platform-cases.tsv` with the must/skip shape per lane; Windows
symlink rows follow the existing envprofile privilege pattern and the ledger vocabulary (no new
skip class); the three hosted lanes are the arbiter.

## Rulings
R1 No product change unless a row exposes a real defect — then fix it minimally and name it in
results.md (the reviewer will judge scope); no test deleted or weakened; the F-C1/F-C2 reviewer
rows stay as committed.
R2 Coverage claim must be a table: hazard/case → row name → production entry (API/CLI) → mutant →
lanes; "conflict admitted ⇒ test fails" for every refusal.
R3 CHANGELOG `### Added` (tests only → say so) and docs cross-references if any doc claims are
now proven. results.md: the table, mutant table, ledger delta, bounds. Publish only on a green
gate.
