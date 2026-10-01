# THE ONLY CURRENT INSTRUCTION — TASK-261001-3fgu9f rework 1 (curator-agent-launcher)

Review verdict rev1 is CHANGES REQUESTED (`TASK-261001-3fgu9f_review-verdict-rev1.md`). Do exactly this, nothing else:
1. **Remove LOGBOOK.md from the change entirely.** Producers never create or edit it. Keep its facts in the task results; one CHANGELOG line is fine.
2. **Revert the agents-management pin bump back to v0.5.22.**
   - go.mod and go.sum go back to the base bytes.
   - The two claude goldens (`pipeline-claude_code-false.golden`, `pipeline-claude_code-true.golden`) go back to the base bytes.
   - CHANGELOG.md: drop the pin sentence and the prompt-suggestion sentence.
   - README.md (lines ~14 and ~51) and SPEC.md (~386) must name the real pin, v0.5.22.

   The pin bump and `CLAUDE_CODE_ENABLE_PROMPT_SUGGESTION` belong to a later, separate task (curator#102).
3. Keep all muse work unchanged:
   - the mapping row;
   - the v3 reader;
   - the tests;
   - the bounds.
4. Run `go test -p 1 ./...` at v0.5.22. Every package must be ok; record the real exit code. Then `git diff d0920353 -- go.mod go.sum '*.golden'` must be empty.

Update the results, then run `task-board handoff TASK-261001-3fgu9f --role developer`, then END YOUR TURN. Never spell any employer name.
