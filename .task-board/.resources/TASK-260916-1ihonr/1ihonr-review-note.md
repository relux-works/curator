# Review note — TASK-260916-1ihonr launcher config family ownership (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev1 (curator-agent-launcher, base 27cc242d, tree 0b302f60, 18 paths, gate green) against `1ihonr-brief.md`:
1. defaults.json and ax.json (machine and operator files) are validated before use: symlink, foreign owner, group/world-writable, and on
   Windows a DACL granting write to another identity → refusal with a named diagnostic; absent → today's behaviour; unreadable → refusal,
   never absent. Check every load site in cmd/curator-run (no path that reads the file without the check).
2. SPEC §4.7 states the contract consistently with §4.3/§4.6; diagnostics named in the SPEC.
3. Rows per refusal + happy path; mutants (symlink check removed; permission check removed; unreadable→absent) killed with real exit codes;
   fake-ax only; Windows rows explicit (run or skip with reason).
4. CHANGELOG.md: confirm this repo's convention allows leaf entries (look at git log of CHANGELOG.md); if not, request it moved to results.
No stray files. accept_cr or changes requested with file:line. No LOGBOOK.md.
