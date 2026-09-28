# Review note — TASK-260924-1nh93t F-M1e, CR revision 2 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review against `1nh93t-brief.md`, the AC and the results; disposable clone of skill-agents-management.
1. The audit table is COMPLETE: re-read launcher SPEC 0.5.0-draft §4 (curator-agent-launcher
   .temp/STORY-260922-39hxog/worktree/SPEC.md) yourself and check every clause needing the module is a row with a real public
   API. A missing row is the most expensive finding possible here (it would stop F-L1b a fourth time) — hunt for one.
2. The public classifier (`agentic.Registry.ClassifyNon…`): versioned with the same per-release grammar; unknown system /
   unverified release fails closed (typed error); prompt text after `--` never a flag; rows per environment × form ×
   placement (incl. codex `exec`); reuses internal/nativeargs (no second grammar).
3. Spelling guard green; README + one new CHANGELOG bullet, released entries untouched.
4. Rev1 failed only on host-load validation timeouts (660 s packages, 10 s probe) — confirm rev2 is the same content and its
   validation log is green. Re-apply one mutant yourself (e.g. classify an unverified release as interactive) → killed.
accept_cr or changes requested with file:line. No LOGBOOK.md.
