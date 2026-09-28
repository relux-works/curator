# TASK-260906-3o75d6 — rework 2 (orchestrator, binding)

Revision 2 was CHANGES_REQUESTED (`TASK-260906-3o75d6_review-verdict-rev2.md`). Everything else verified OK.
Fix exactly:
1. `cli/curator.md` ~57-58: replace "named above as onboarding triggers" with
   "named in environments section 9.5 as onboarding triggers" (check the section number against
   `protocol/environments.md` before writing it), and adapt `takeover_cli_clause` / the exact-note pin in
   `tools/validate.py` so the pin matches the new text (it must stay an exact-equality pin).
2. Drop the stray blank line before the closing fence of the example block (~line 232).
Nothing else. `python tools/validate.py` + the validator unittest file(s) you touched, each as a bounded call;
`make regenerate-check`. Append "Revision 3" to results, then `task-board handoff TASK-260906-3o75d6 --role developer`.
A `run_wrote_outside_worktree … policy warn` block is a warning — verify the status moved to `to-review`.
