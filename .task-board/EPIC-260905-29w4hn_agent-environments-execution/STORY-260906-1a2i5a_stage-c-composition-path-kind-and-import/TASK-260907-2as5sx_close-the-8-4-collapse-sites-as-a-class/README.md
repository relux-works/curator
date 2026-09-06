# TASK-260907-2as5sx: close-the-8-4-collapse-sites-as-a-class

## Description
Close the section 8.4 absence-versus-failed-read collapse as a class rather than one site at a time. Five sites have now been found in three separate review cycles.

## Scope
Section 8.4 makes absence and a failed read different facts, so a fallback defined for absence must not fire on a read failure. Sites found so far: readSkillsLedger (fixed in ee6743a2), readRootSurface (cycle-2 C2-B1), the KindPath snapshot read (cycle-4 C4-M1), and two more the cycle-6 reviewer observed without treating them as blocking: status.go:432 and status.go:521, alongside purgeHomes. The stage (c) producer answered honestly that nothing structural prevents a sixth: the sites read different manager-owned shapes through different helpers and no shared helper covers them. That answer was accepted for the leaf and is the reason for this task.

## Acceptance Criteria
Either a shared read seam distinguishes absence from failure for every read of manager-owned state, or a test enumerates those read sites and asserts each distinguishes the two. The five known sites are covered either way, each proven by a narrowing mutant that collapses one of them and kills a named test. If a structural seam is judged infeasible, the report says why with the alternatives weighed rather than asserting vigilance.
