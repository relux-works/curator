# TASK-260929-jup8re review verdict rev1 — ACCEPTED
Candidate tree abd4aee5 (= worktree), base fc499a96. Driven naming-gate.sh directly on temp trees (names assembled at runtime).
1. Content-free: short- and full-name plants -> rc=1, output `./d/x.md:1` + summary, planted word count in output = 0 (both checks).
2. Signature: prose "step name: ... ./x" rc=1; step name + space (no TAB) + timestamp rc=1; JSON `\t` echo rc=0; literal-TAB echo with full name rc=0. Binary-patch exemption unchanged (diff touches only the report line + ECHO test).
3. Selftest rows f,g,h,i present (+ j,k for a truncation arm). Mutants (sed copies, real rc):
   - M1 content printed again: output leaks planted word (leak=1) -> row (i) grep kills it. KILLED.
   - M2 signature relaxed to step name: prose probe rc=0 (want 1) -> row (h) kills it. KILLED.
   - M3 exemption removed: echo probe rc=1 -> rows f/g kill it; origin/main archive rc=1. KILLED.
4. `git archive origin/main` (fc499a96) extraction: naming-gate rc=0. Worktree rc=0.
5. Residual stated in the header comment: forged exact-signature line is exempt; prose never matches (probes above).
Observations (non-blocking): second ECHO arm (U+2026 ellipsis + timestamp tail `T hh:mm:ss[.f]Z ./`) goes beyond the brief; it still requires a machine-shaped timestamp+` ./`, and prose-with-ellipsis probe/row (k) fails as required — same forged-line bound. Fail rows (h)/(k) assert rc only, not the summary text; (i) asserts path:line. Full gate-selftest.sh not rerun by reviewer; accepted hosted gate green for rev1.
