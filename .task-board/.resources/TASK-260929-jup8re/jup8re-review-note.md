# Review note — TASK-260929-jup8re naming gate: content-free reports + CI-echo exemption (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev1 (base fc499a96, tree abd4aee5, 2 paths, gate green including the Naming gate on a tree that contains the echoed records of
BUG-260928-uyak0e) against `jup8re-brief.md`. Verify:
1. On failure the gate never prints line content, for both checks: only path:line and the summary. Plant the short word in a temp tree
   and confirm that the output does not contain it.
2. The exemption needs the full signature: the step name, then a TAB or a literal backslash-t, then an RFC3339 Z timestamp, then a space
   and "./". Lines without it are scanned in full. The binary-patch exemption is unchanged.
   - Probe prose with the step name but no timestamp: it must fail.
   - Probe a JSON line with the signature: it must pass.
3. Selftest rows f–i exist, and each fail row fails for the naming reason. Re-run the three mutants yourself with real exit codes:
   content printed again; signature relaxed; exemption removed.
4. `bash .github/ci/naming-gate.sh` on a `git archive origin/main` extraction exits 0.
5. The residual is stated: a forged signature line is exempt. Accept it only if prose can never match.
accept_cr, or changes requested with file:line. No LOGBOOK.md. Never spell either name anywhere, including in your verdict.
