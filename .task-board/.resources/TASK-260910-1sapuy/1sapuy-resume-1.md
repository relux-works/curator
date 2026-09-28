# TASK-260910-1sapuy — resume on top of the security_posture model (THE ONLY CURRENT INSTRUCTION, with 1sapuy-sec-brief.md)

Your block was right: the effective security_posture was missing. TASK-260927-4pv4au (security_posture revision A) is ACCEPTED and
CHECKPOINTED on this Story's branch — your Story worktree now contains it (HEAD = the 4pv4au checkpoint commit). The revision-B flip was
moved to another Story. You are the LAST leaf of STORY-260910-2qmrb8: when you land, the Story lands (4pv4au + your change).
1. `task-board m 'set_status(TASK-260910-1sapuy, status=development)'`.
2. Implement the unreachable-registry pair of vectors/security-posture.json through `curator install` / `update`, using the effective-
   posture API 4pv4au exposes (do not re-derive posture): `unreachable-registry-permissive-warns` (prominent gate notice naming every
   artifact resolved without registry evidence) and `unreachable-registry-hardened-refuses` (refusal). Cite profiles/manager.md §7.1 and the
   registry evidence clauses. Remove the Story-owned gap rows these cases held; one mutant per rule killed (real exit codes); stateread for
   manager-state reads; the permissive warning rules of 4pv4au stay (stderr only, never on launch paths).
3. No CHANGELOG/LOGBOOK edit (entry in results). Update the results resource, handoff, END YOUR TURN. Write only inside your Story worktree.
