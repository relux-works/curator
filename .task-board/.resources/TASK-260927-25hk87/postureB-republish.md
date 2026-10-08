# THE ONLY CURRENT INSTRUCTION — TASK-260927-25hk87: republish on the converged base (developer, handoff only)
The orchestrator converged your workspace onto fresh main. It dropped your `scripts/remote-gate.sh` hunk, because main already carries the equivalent privacy edit (1fc0a93a). Nothing else changed. Snapshot: refs/campaign/postureB-snapshot-20261005.
1. Check that `git diff --stat` shows your posture-B changes and does NOT touch scripts/remote-gate.sh.
2. Append "Converged; remote-gate hunk dropped (on main via N4)" to your results resource and run `resource update`.
3. `task-board handoff TASK-260927-25hk87 --role developer` and END YOUR TURN. Do not edit code. Do not edit remote-gate.sh, LOGBOOK.md or CHANGELOG.md.
