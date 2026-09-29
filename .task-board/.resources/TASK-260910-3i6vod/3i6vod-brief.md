# TASK-260910-3i6vod — security hardening leaf (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md` and the task description. Security epic EPIC-260910-2hw1xb, Story STORY-260928-rp2r1j. If curator-spec v1.0.0-rc.13 (SPEC_PIN
23435129) has a clause or vector for this surface, cite it and drive it; otherwise add production-entry rows that prove the rule.
Focus (docs, S7): README and SECURITY.md state plainly that installed commands run with the user's privileges under portable assurance, and that script-worker-v1 (enforced commands) and verified mode are the enforcement paths — with links to the spec sections. Docs-only; a docs link/anchor check if the repo has one.
One mutant per rule (the protection removed) survives before and is killed after (real exit codes). No CHANGELOG/LOGBOOK edit (entry text in
results under "## CHANGELOG entry (for release prep)"). New manager-state reads via internal/stateread; managed writes via the E5 nofollow
helpers. Write only inside your Story worktree. Attach results, check DoD, `task-board handoff TASK-260910-3i6vod --role developer`, END YOUR TURN.
