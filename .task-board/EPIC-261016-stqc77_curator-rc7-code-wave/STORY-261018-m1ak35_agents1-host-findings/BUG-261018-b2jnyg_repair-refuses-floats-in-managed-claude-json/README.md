# BUG-261018-b2jnyg: repair-refuses-floats-in-managed-claude-json

## Description
After Claude Code has run in a managed home, env resolve --repair fails with environment_repair_failed: CCJ-1 does not support float64. Claude writes floats into .claude.json (cachedGrowthBookFeatures probabilities, projects[...].lastCost, lastFpsAverage, costUSD). Any profile update then leaves the home unrepairable, and task-board refuses its automatic repair, so a reboot does not restore the session. Reproduced on agents1 (Linux, rc.7).

## Scope
(define bug scope / affected area)

## Acceptance Criteria
Repair of a managed home whose .claude.json holds floats succeeds, adds only the launch-directory entry and preserves every other value and number text; an already-satisfied file is not rewritten; the absent-file seed and existing refusals are unchanged; a regression test and plant control prove it; full local suite and hosted gate pass on the exact signed candidate.
