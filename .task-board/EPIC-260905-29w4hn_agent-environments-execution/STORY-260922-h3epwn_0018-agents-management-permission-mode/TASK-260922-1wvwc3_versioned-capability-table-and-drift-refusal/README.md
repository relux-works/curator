# TASK-260922-1wvwc3: versioned-capability-table-and-drift-refusal

## Description
F-M1b: encode the versioned provider-capability table keyed (environment, tool release) -> permission grammar version (0018 choice 6): each mapping is re-verified per tool release; on version drift the yolo mapping fails closed first (yolo refused for unverified releases) while native still forwards verbatim with no claims; unknown future native policy forms (new codex -c keys, new claude modes) are refused as usage exit 2 and never resolved into a policy claim (choice 3); the grammar distinguishes prompt text from flags (parsing rule owned here). Then cut the release the launcher pins.

## Scope
skill-agents-management: capability table + version token, drift refusal, parsing rule, goldens, docs, CHANGELOG, release tag (next patch of v0.5.x). Depends on F-M1a.

## Acceptance Criteria
1) table present and versioned; an unpinned or newer tool release refuses yolo with a named diagnostic and still forwards native verbatim (golden pair); 2) unknown native policy forms are refused as usage exit 2 (rows for a new codex -c key and a new claude mode); 3) parsing rule rows: prompt text that looks like a flag is never parsed as one; 4) one narrowing mutant per refusal bound executed and killed; 5) release tagged and CHANGELOG names the grammar version token the launcher must cite.
