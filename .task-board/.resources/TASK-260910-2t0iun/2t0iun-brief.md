# TASK-260910-2t0iun — security hardening leaf (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md` and the task description. Security epic EPIC-260910-2hw1xb, Story STORY-260910-234vmx. If curator-spec v1.0.0-rc.13 (SPEC_PIN
23435129) has a clause or vector for this surface, cite it and drive it; otherwise add production-entry rows that prove the rule.
Focus: install.sh verifies the release before installing — GitHub artifact attestation (gh attestation verify) when gh is available, otherwise an independent signature of checksums.txt (minisign or cosign, with the public key pinned in the script/repo), then the checksum of the downloaded archive; any verification failure refuses to install with a clear message; document an explicit, loud opt-out only if the spec/README already defines one. Rows via a local fixture release (no network in tests).
One mutant per rule (the protection removed) survives before and is killed after (real exit codes). No CHANGELOG/LOGBOOK edit (entry text in
results under "## CHANGELOG entry (for release prep)"). New manager-state reads via internal/stateread; managed writes via the E5 nofollow
helpers. Write only inside your Story worktree. Attach results, check DoD, `task-board handoff TASK-260910-2t0iun --role developer`, END YOUR TURN.
