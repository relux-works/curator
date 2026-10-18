# THE ONLY CURRENT INSTRUCTION — BUG-261018-dkgv1p: rerun the full hosted gate (developer)

Your previous run stopped because the self-hosted Linux runners hit "no space left on device" before any behavioural test ran. The runner operator has fixed this: job temp now lives on disk and is emptied after every job, and the shared tmpfs is back to 17% use.

**Do this:** do not change the candidate unless the gate shows a real failure. Rerun the full hosted gate on your exact signed candidate (the SHA recorded in your last outcome). If it is green, complete the remaining checklist rows with the run URL and job results, then hand off with `task-board handoff BUG-261018-dkgv1p --role developer`. If a real, source-caused failure appears, fix it, re-sign, rerun, and record both runs. A runner infrastructure failure is still a blocker to report, never a pass.

End your turn after the handoff, or after recording a blocker.
