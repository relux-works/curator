# TASK-260910-31ocjt — security hardening leaf (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md` and the task description. Security epic EPIC-260910-2hw1xb, Story STORY-260928-1t6bto. If curator-spec v1.0.0-rc.13 (SPEC_PIN
23435129) has a clause or vector for this surface, cite it and drive it; otherwise add production-entry rows that prove the rule.
Focus: deliver the HTTPS askpass secret to the fetch child through a pipe or inherited fd instead of CURATOR_BUILD_HTTPS_ASKPASS_SECRET in the environment, so grandchildren (and /proc/<pid>/environ readers) cannot inherit it; works on macOS/Linux/Windows (Windows: inherited handle or named pipe with a private ACL); the secret never touches argv, env, or disk. Rows: the child receives the secret; a grandchild spawned by the askpass helper does not see it in its environment; the env var is gone everywhere.
One mutant per rule (the protection removed) survives before and is killed after (real exit codes). No CHANGELOG/LOGBOOK edit (entry text in
results under "## CHANGELOG entry (for release prep)"). New manager-state reads via internal/stateread; managed writes via the E5 nofollow
helpers. Write only inside your Story worktree. Attach results, check DoD, `task-board handoff TASK-260910-31ocjt --role developer`, END YOUR TURN.
