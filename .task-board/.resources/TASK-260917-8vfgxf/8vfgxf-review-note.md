# Review note — TASK-260917-8vfgxf interim NUL-opaque rule (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev2 (base 0e3169bb, tree 6dd496cf, 8 paths, gate green on every lane; rev1 failed only on a Windows-reserved filename and the
stateread guard) against `8vfgxf-brief.md` and `8vfgxf-gatefix-1.md`. Verify:
1. The rule is enforced at the production entry (install / update / resolve admission of skill and context snapshots). Any REGULAR
   file containing 0x00, in any directory and with any extension, is a blocking opaque finding that names the file. Symlinks and
   directories are unaffected. Cite the call sites and confirm that no admission path bypasses it.
2. Rows:
   - the two colliding trees are built from the v1 framing, have equal v1 hashes, and are both refused;
   - a NUL-free tree installs;
   - a deep docs/ or assets/ NUL file is refused.

   rc.13 wording or vectors are cited, or their absence is stated.
3. Re-run the mutants yourself (rule limited to one directory; finding demoted to a warning) with real exit codes, and confirm each
   is killed.
4. The stateread guard fix routes the read through internal/stateread, or moves the allowlist entry without widening it. No
   Windows-reserved names anywhere. No CHANGELOG/LOGBOOK; no stray files.
accept_cr, or changes requested with file:line. Never spell any employer name.
