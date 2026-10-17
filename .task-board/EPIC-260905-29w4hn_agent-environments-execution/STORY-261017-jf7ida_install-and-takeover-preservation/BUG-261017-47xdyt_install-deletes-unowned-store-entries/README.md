# BUG-261017-47xdyt: install-deletes-unowned-store-entries

## Description
P1 from the preservation triage TASK-261017-yyg01d (items A and C2): project install deletes or replaces content it does not own. A: stale cleanup removes every unowned sibling under .agents/skills/ (an unmanaged skill folder, a notes-only folder, a loose file), exits 0 without notice, and leaves a foreign native link dangling. C2: an existing unowned directory at a selected canonical destination (manual SKILL.md and notes, no marker) is replaced. Rule (skillfile-sources 2, core 10-11, manager 2.5): only entries with valid preceding Curator ownership may be removed or replaced; anything else refuses before any destructive publication, naming every path, and stays byte-identical.

## Scope
internal/install staging (targets.go stageStaleSkillRemovals and stageNode, install.go stageProjectTargets, global.go review), internal/marker ownership reads; tests

## Acceptance Criteria
1. Stale cleanup removes only entries with valid preceding Curator ownership; unowned skill folders, notes folders and loose files under the canonical store refuse the operation by name before any publication and stay byte-identical, and foreign native links stay valid. 2. An existing unowned or unproven canonical destination for a selected node refuses by name with no partial change; an absent destination installs and a valid owned node updates. 3. Unreadable or invalid ownership evidence refuses and is never treated as absence or permission. 4. Mixed owned and unowned candidates: owned stale entries are still removed, so a reject-everything implementation fails. 5. Global and hybrid callers of the same helpers keep their contracts; transaction and GC invariants hold. 6. Narrowing mutants fail as intended; full remote gate green.
