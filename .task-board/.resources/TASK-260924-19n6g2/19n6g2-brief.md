# TASK-260924-19n6g2 — curator-spec v1.0.0-rc.13 release prep (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md`. spec main now carries: #90 (skillfile-sources residual clarifications, 7b06bb3), #96 (ACCEPTED
skillfile-sources revision 1 + repository transport revisions 1 and 2, Decision 0022, 5746367 — cocoaskills-authored) and #88 (script-host
hard-link erratum, TASK-260924-mcmova, + Implementations CI per-implementation claims, TASK-260926-4ek1qg), main = 2c39c428. rc.13 = everything on main since v1.0.0-rc.12 (dced9b83).
1. Follow EXACTLY how rc.10–rc.12 were cut (git log/show of those release-prep commits; tools/release_gate.py,
   tools/verify_release_commit.py, release metadata files, CHANGELOG section, COMPATIBILITY/README version references). Do not invent a new
   process.
2. CHANGELOG: turn "Unreleased" into the v1.0.0-rc.13 section (keep every entry; highlight: skillfile-sources revision 1 and repository
   transport revisions 1–2 ACCEPTED as a separately pinnable suite conformance/skillfile-sources-v1, consumable against the rc.10 core suite;
   manifest v9 `directory` NOT included; script-host hard-link erratum; Implementations CI checks each manager against its claim —
   cocoaskills as a partial client on core rc.10 + skillfile-sources-v1).
3. Run the release gate recipes locally in bounded parts (real exit codes); regenerate-check clean.
4. Deliver the release-prep change as your Story candidate (the orchestrator lands it via PR with all required checks and then creates the
   signed tag v1.0.0-rc.13 on that main commit — NOT you). DoD items must be checkable before the tag exists.
No LOGBOOK.md. Attach results (what changed, gate outputs, the exact tag command the orchestrator should run and the release-workflow
check to watch), check DoD, `task-board handoff TASK-260924-19n6g2 --role developer`. A write-boundary `policy warn` block is a warning.
