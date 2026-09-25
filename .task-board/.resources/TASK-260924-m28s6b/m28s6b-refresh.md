# TASK-260924-m28s6b — refresh onto trunk and publish (THE ONLY CURRENT INSTRUCTION)

Your lock-replay work + rework 1 (remedy/docs/test agreement) is complete and uncommitted (safety ref refs/campaign/m28s6b-delta-20260924).
Publication refused `change_request_base_authority_mismatch`: the Story checkpoint 41dd18dd (TASK-260924-1aa9wb on 1511b345) does not
descend from trunk `5b326aa3` (R5, 6chzf9, 3mazfw landed).
1. `task-board m 'set_status(TASK-260924-m28s6b, status=development)'`.
2. CHANGELOG POLICY: the Story must not change CHANGELOG.md at all — make the working candidate's CHANGELOG.md equal trunk's bytes (this also
   drops the checkpointed 1aa9wb entry); put BOTH entries (1aa9wb default-on, m28s6b lock replay) verbatim in the results resource under
   "## CHANGELOG entry (for release prep)". Remove any stray root TASK-*/BUG-*.md file.
3. Combine trunk's incoming content: `git diff 1511b345 5b326aa3 -- . ':!.task-board' ':!CHANGELOG.md' | git apply --3way` (keep both sides), leave
   nothing staged, run `task-board worktree refresh-candidate TASK-260924-m28s6b`; if the 1aa9wb checkpoint replay conflicts on CHANGELOG.md,
   resolve it to TRUNK's bytes via the command's `--replay-resolutions` template (never hand-commit the replay worktree).
4. Bounded re-runs of your replay rows + `TestDraftDocsPinExamples` + the default-on rows. Append "Revision 2 — refresh" to results,
   `resource update`, `task-board handoff TASK-260924-m28s6b --role developer`. A `run_wrote_outside_worktree … policy warn` block is a warning.
