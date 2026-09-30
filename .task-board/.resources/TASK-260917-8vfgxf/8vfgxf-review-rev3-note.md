# Review note — TASK-260917-8vfgxf rev3 re-review (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Re-review rev3 (base 0e3169bb, tree 0a10a669, 10 paths, gate green) against your rev2 verdict (F1, F2) and `8vfgxf-rework-1.md`. Verify:
1. F1: the NUL-opaque check runs unconditionally, before the `!cfg.Audit.Enabled` early return, or at every admission site. With the
   DEFAULT config (audit disabled), each of these refuses both colliding trees and a deep docs/ or assets/ NUL file:
   - install.Install (install.go:580);
   - the global lane (global.go:198);
   - the CLI paths (main.go:1896/2198);
   - strictAuditMember (envprofile.go:1712).

   Re-run your temporary Gate probe with Enabled=false: it must now refuse. The rest of the audit stays opt-in.
2. F2: the blocking error names the file, and the install-lane row asserts it.
3. Mutants, run yourself with real exit codes, each killed on the install lane with the default config:
   - rule gated behind Audit.Enabled again;
   - rule limited to one directory;
   - finding demoted to a warning.
4. No other behaviour change in the audit; no Windows-reserved names; the stateread guard is still green; no CHANGELOG/LOGBOOK.
accept_cr, or changes requested with file:line. Never spell any employer name.
