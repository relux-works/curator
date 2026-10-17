# Review acceptance receipt

`accept_cr` exited 0 and returned CR state accepted, revision 1, reviewer run
RUN-261017-f6dc4a and element status integrating. A separate fresh board read
confirmed integrating. This is acceptance, not landing.

The mutation also emitted `run_wrote_outside_worktree` under policy warn,
including unattributed other-task activity and foreign-worktree paths. The
reviewer made no direct control-root edits: scratch remained within the run
worktree; board writes used task-board. The advisory was not cleared and did
not refuse acceptance. Attribution remains for the delivery owner to inspect.

An attempted append of this receipt to the accepted verdict was refused with
change_request_resource_collision (missing ownership digest). The accepted
verdict remains unchanged; this separate receipt preserves the operational
observation without rewriting acceptance evidence.
