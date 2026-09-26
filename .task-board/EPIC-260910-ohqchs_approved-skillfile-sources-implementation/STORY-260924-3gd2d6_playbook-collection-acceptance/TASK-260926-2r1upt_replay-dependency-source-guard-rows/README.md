# TASK-260926-2r1upt: replay-dependency-source-guard-rows

## Description
Residual of TASK-260924-11burj review (verdict rev2, bounds): the production fix declaredDependencyReplaySources in internal/install/draftsources.go (fresh-machine replay of a transitive network-Git member recovers the dependency declared source from the replayed requirer) has only a whole-change mutant kill; its individual guards (directory mismatch -> source_snapshot_changed, identity mismatch -> source_snapshot_changed, and any other branch) have no dedicated killing rows. Add one production-entry row per guard and kill a per-guard mutant.

## Scope
(define task scope)

## Acceptance Criteria
one row per guard in declaredDependencyReplaySources; each per-guard mutant survives before and is killed after (real exit codes); no CHANGELOG edit
