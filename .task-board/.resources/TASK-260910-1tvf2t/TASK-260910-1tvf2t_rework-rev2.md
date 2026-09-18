# Rework brief — TASK-260910-1tvf2t, revision 2 (S2)

Revision 1 was rejected with three corrections
(`TASK-260910-1tvf2t_review-verdict-rev1.md`, F1–F3). The brief-conformance
matrix passed otherwise; keep everything else byte-identical.

- **F1 — pins must require present, typed evidence (rule 7).** The
  `bootstrap_cases` gate uses `is True` / `is not True` predicates, so a
  scenario with the discriminating input REMOVED (e.g. `signature_valid`,
  `candidate_same_body`, `same_log_size`, `roots_equal`) still passes under
  its name — 0/5 of the reviewer's removal probes refused through
  `validate.main()`. Require each phase's discriminating inputs to exist
  with exact types, require explicit `False` where false is the
  discriminator, and add missing/null/wrong-type negatives plus
  self-consistent replacement negatives (incl. a main-entry negative) so all
  five probes refuse. Keep the 15 scenario names and the valid vectors.
- **F2 — status severity is explicit per row.** manager §10 blanket-treats
  "every other bootstrap row as a warning row", which downgrades the
  `registry_checkpoint_regression` ERROR (registry §5.1) and mislabels a
  successful checkpoint bootstrap. State the status outcome per row:
  successful checkpoint bootstrap (current, informational), TOFU (warning
  row), checkpoint regression (error, non-current), divergence (warning
  under advisory / non-current error under strict); pin the refusal's
  status mapping in a vector case. No new diagnostic.
- **F3 — cache-backed first fixation stated once.** registry §8 says
  caching/offline grace never bootstrap rollback state, yet requires TOFU
  posture when the fixing view came from the cache; §5 and the diagnostic
  table describe network-only first fixation. Make §5, §8 and the
  diagnostic condition agree: settle that a cached snapshot MAY establish
  first-use high-water only when it was itself accepted online under §5
  (i.e. it is the persisted state, not a fresh fixation) and that a
  never-validated cache entry never fixes state — or, if you find the
  landed text implies otherwise, choose the reading that keeps "caching
  cannot bypass checkpoint validation, lower high-water, or recover
  known-lost state" and pin the permitted path's posture in a vector.
- Minor: CHANGELOG "no behavior change without a checkpoint or a mirror
  group" overlooks the new TOFU warning — fix the sentence.

## Validation and handoff
`make validate` and the regeneration proof (exit codes); evidence "Revision
2" section; `TASK-260910-1tvf2t_spec-patch_rev2.patch` = `git diff HEAD` of
the worktree (base `4a2fa3e`) with new files via `git add -N`; EMPTY curator
delta; `task-board handoff TASK-260910-1tvf2t --role doc-writer`. Worktree
and rules unchanged.
