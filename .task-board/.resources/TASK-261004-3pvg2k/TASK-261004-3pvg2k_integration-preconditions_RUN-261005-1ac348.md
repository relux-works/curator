# Integration preconditions

Task: introduce-cips-directory-and-process. Run: RUN-261005-1ac348.

Read-only board inspection confirms status integrating, CR revision 1 accepted, kind story_final, producer developer/implementer, current run holds workspace lease. Base 43bf0a2506d5c354a73bbc3ea4623d4653db10c7; accepted candidate cd8e895fe3cf3b98eb65114c70b3382bc0850f49.

Standalone Python precondition check exit 0: HEAD equals accepted base; tracked contents and all three untracked CIP files match accepted candidate; exact candidate delta is GOVERNANCE.md, README.md and three cips Markdown files. Manifest SHA-256 unchanged: 6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5. No repository files modified.

Earlier broad git-diff diagnostic exited 1 because Git treats candidate files absent from the index as deleted despite their presence as untracked files. Corrected check compares tracked paths with git diff and untracked bytes with git show; it exited 0. Two initial query syntax probes exited 1 and were corrected through schema/get overview.

make validate was not rerun: this integration run made no changes; validation and acceptance remain those of accepted revision 1, not newly claimed test results. Runner owns final authority freshness, validation reuse and landing gates.

Board worktree status reports board_debt indeterminate because the authoritative board is outside the control root. This is a reported limitation, not proof of landing refusal or success.

Per latest Integration Assignment, no status mutation, handoff, checkpoint or integrate invoked. Runner performs bound synchronous landing after this run exits. Landing outcome is pending.