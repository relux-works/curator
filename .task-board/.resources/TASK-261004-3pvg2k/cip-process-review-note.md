# Reviewer instruction: TASK-261004-3pvg2k "introduce CIPs directory and process" (curator-spec)
Cross-provider review: the producer was codex sol high, so the reviewer is claude-sonnet-5-5 high (operator rules R121/R195).
The operator asked for this (2026-10-04): introduce Curator Improvement Proposals (CIPs) in curator-spec `cips/`. Research evidence stays in each repo's `.research/`. Numbering: 0001 is the process itself. 0002–0006 are reserved for the drafts on curator main (`.research/261004_CIP-000[2-6]*`). They are published by a later task, NOT this one.
Check:
1. `cips/README.md` (or 0001) defines: the lifecycle (draft → proposed → accepted/rejected/withdrawn → implemented/superseded); the owner and arbiter of status changes (the operator decides acceptance); the numbering rule; the required sections (summary, motivation, current state with evidence links, proposal, alternatives, compatibility/migration, security, open questions); and how a CIP relates to `decisions/` (CIP = proposal; an accepted CIP that changes normative text lands as a decision plus protocol edits).
2. The template file exists and matches the README.
3. No normative protocol text changes; frozen or released schemas are untouched; CHANGELOG/LOGBOOK rules are respected (producers never edit LOGBOOK.md).
4. Links resolve. There are no personal paths, host names or employer names (the repo is public).
5. Anything in the brief left undone, or scope creep.
Verdict through the board (`accept_cr` with the merged checklist, or changes requested with numbered findings). Do NOT edit files. Do NOT run go tests. This is a docs-only review.
