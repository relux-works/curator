# BUG-261002-ot3ea1 — global-upgrade-sweeps-in-use-builds

## Revision 5 — drop the CHANGELOG change

Applied the binding sweep-drop-changelog.md instruction. CHANGELOG.md is byte-identical to base c085b4d2. Release notes remain owned by the central rc.3 release-notes step. All other repository changes are unchanged from the starting revision 4 candidate; LOGBOOK.md was not edited. Changes remain uncommitted.

Commands run directly and actual exit codes:

- `git diff c085b4d2 -- . ':!CHANGELOG.md' > /tmp/BUG-261002-ot3ea1-rev5-before.diff`: 0.
- `git checkout c085b4d2 -- CHANGELOG.md`: 0.
- `git diff --quiet c085b4d2 -- CHANGELOG.md`: 0 (base identity verified).
- `git diff --stat c085b4d2`: 0; 17 files changed, 637 insertions, 29 deletions; CHANGELOG.md absent.
- `git diff c085b4d2 -- . ':!CHANGELOG.md' > /tmp/BUG-261002-ot3ea1-rev5-after.diff`: 0.
- `cmp /tmp/BUG-261002-ot3ea1-rev5-before.diff /tmp/BUG-261002-ot3ea1-rev5-after.diff`: 0 (every other path unchanged).

No Go test, build, vet, lint or mutant command was rerun in revision 5, as explicitly directed by sweep-drop-changelog.md. The instruction states that revision 4's hosted gate was green; that is accepted prior evidence, not a gate executed or independently reverified in this run. No new command-based checklist items were checked.
