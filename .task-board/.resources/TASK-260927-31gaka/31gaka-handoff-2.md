# TASK-260927-31gaka — hand off and END THE TURN (THE ONLY CURRENT INSTRUCTION; bound run)

The orchestrator's earlier instruction to wait for the hosted gate inside your run was WRONG and caused both blocks: the runner publishes
the Change Request and runs the gate only AFTER your turn ends (see the corrected "Handoff and the hosted gate" section of
campaign-producer-rules.md). Do ONLY:
1. `task-board m 'set_status(TASK-260927-31gaka, status=development)'`.
2. Append "Revision 1 handoff (orchestrator-corrected procedure)" to TASK-260927-31gaka_results.md and `task-board resource update` it.
3. `task-board handoff TASK-260927-31gaka --role developer`, then END YOUR TURN immediately. Do not poll for the CR, do not set blocked.
Change no code. No CHANGELOG/LOGBOOK edit.
