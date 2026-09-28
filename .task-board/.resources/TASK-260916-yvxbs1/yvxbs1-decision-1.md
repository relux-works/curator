# TASK-260916-yvxbs1 — orchestrator decision (THE ONLY CURRENT INSTRUCTION, with yvxbs1-sec-brief.md)

You are right: the pinned rc.13 text and vector win. The DoD item "path-kind package containing a class: system module is refused" was the
orchestrator's error and is REMOVED; it is replaced by "class: system modules in path-kind packages follow rc.13 §3: a trusted direct path
overlay's system module is admitted (path-overlay-system-module-admitted), a transitive one refused through the existing admission".
The task description/AC sentence about refusing system modules is superseded by this decision (rc.13 §3) — cite this decision when you
check "Code written per task description and AC". The "hosted gate green" DoD item was also removed (handoff runs the gate; it cannot be
checked before handoff) — the rule still holds: after handoff, wait for the gate; if red, fix and republish.
Also: the 3 ownership boundary vectors bounded "because this host cannot chown" — drive them instead through an owner-lookup seam in the
boundary validator (test-injected uid/owner, production uses the real Lstat owner), so the rule is proven on every lane; bound only what
genuinely needs a second real user, with an owner. Kill one mutant on the ownership rule.
Then: `task-board m 'set_status(TASK-260916-yvxbs1, status=development)'`; trunk is eca2bf27 — `git fetch origin main` and combine if your
base is older (VERIFY `git diff --name-only origin/main -- . ':!.task-board'` lists only your paths); focused runs split by -run groups with
real exit codes; check the DoD items; handoff and WAIT for the gate (never interrupt); hand off only green. No CHANGELOG/LOGBOOK edit.
