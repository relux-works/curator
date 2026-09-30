# TASK-260930-1kylpg — pin named-path absence at the store boundary (THE ONLY CURRENT INSTRUCTION)

Residual R-1 from the S5 review (TASK-260910-32gki6_review-verdict-rev8.md, "Mutant … SURVIVES"). internal/pathboundary skips a transient
child that vanishes during a tree walk (uyak0e), but a NAMED target (the root, or a lock-named store entry) must fail when missing. No
row pins this today.
1. internal/pathboundary: add a row. `ValidateWithinWithOwner(root, missingTarget)` (or the production named-route entry) on a missing
   named target → *Failure (CheckRegular or the documented class). This must hold on unix and Windows.
2. internal/envprofile: add a row through the production resolve entry. Remove one lock-named store entry after lock creation, then show
   that env resolve reports environment_store_untrusted for that entry, and that the failure arises at the boundary step, not only at pin
   recomputation. If the code proves it cannot distinguish the two, state that with evidence and pin what is observable.
3. Mutant, with real exit codes: in the three named-route Lstat loops, add `return nil` on IsAbsent, and drop `path == target ||` from
   vanished(). It must survive on current main and be killed by your rows.
4. Tests only; change no production code unless a row exposes a real bug, and if one does, stop and report it in the results. No
   CHANGELOG/LOGBOOK. Never spell any employer name.

Update the results, then run `task-board handoff TASK-260930-1kylpg --role developer`, then END YOUR TURN.
