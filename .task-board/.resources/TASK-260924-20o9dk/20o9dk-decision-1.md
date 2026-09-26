# TASK-260924-20o9dk — orchestrator decision on the three local-snapshot-v1 schema documents (THE ONLY CURRENT INSTRUCTION, with 20o9dk-brief.md)

Decision: option 1. Curator's snapshot model stores immutable trees and recomputes Inventory through Capture/OpenLocal; no production
path reads serialized local-snapshot inventory JSON, and adding an unused parser to clear a checklist is a forced fit. Keep the three
local-snapshot-v1 schema documents as EXPLICIT BOUNDS (the corpus test's `bound` class) whose reason names exactly that: "Curator never
reads serialized inventory JSON; inventory is recomputed from the stored tree (internal/snapshot Capture/OpenLocal)". Checklist item 1 is
satisfied by: every semantic case and snapshot vector has a passing production-entry row (105/105, 3/3) AND every other released case is
either driven or a stated bound with that reason — cite this decision resource when checking it. If the ledger/ratchet requires a row for
bounds, add it with owner TASK-260924-20o9dk and the reason above.
1. `task-board m 'set_status(TASK-260924-20o9dk, status=development)'`.
2. Make sure the bound reason text is exactly stated in the test classification (and ledger if required); nothing else changes.
3. Check item 1 citing this resource, then `task-board handoff TASK-260924-20o9dk --role developer`; stay in the turn while the gate runs.
   A write-boundary `policy warn` block is a warning.

ADDENDUM (orchestrator): the Story was converged onto trunk 9f0da708 (11burj landed; it changed internal/install/draftsources.go —
declaredDependencyReplaySources). Your delta was carried by 3-way merge (safety ref refs/campaign/20o9dk-delta-20260926). Before handoff:
confirm draftsources.go keeps BOTH 11burj's transitive-replay recovery and your changes (C1 object-format check etc.), rerun the focused
replay rows incl. `TestDraftSourcesPlaybookCollectionAcceptanceThroughProductionCLI` (bounded), and verify `git diff --name-only HEAD`
lists only your paths.
