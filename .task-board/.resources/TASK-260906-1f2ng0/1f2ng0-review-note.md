# Review note — TASK-260906-1f2ng0 stage (a) follow-ups, CR revision 5 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

First content review. Review revision 5 against `1f2ng0-brief.md` (FU-1..FU-4) and the results; read-only, disposable clone.
1. FU-1: at the production CLI entry, missing path operand → `profile_source_path_missing`, unreadable dir →
   `profile_source_path_unreadable`, neither fires on the other, regular-file row present.
2. FU-2: `loadMachinePolicy` — absent ⇒ default; any other read error refuses (§8.4). Tests prove both.
3. FU-3: operator warning for a global skill not migrated into the default profile lock — implemented with the spec's exact
   wording (environments §9.4), or a recorded reason why the spec does not ask for it. Check the wording against the spec.
4. FU-4: partial first install prints the installed line; compatible with 187z6x's `updated profile …` reinstall wording
   (checkpointed on the same Story branch) — check both.
5. Windows: revisions 2-4 fought the platform-case gate. Revision 5 must: leave `internal/godriver` byte-identical to base;
   give the three POSIX-only unreadable subtests their own ledger rows `linux,darwin  windows  platform-control …` with a
   registered skip reason; no broadened class (`no-broad-suppression.sh` green). Absent-vs-unreadable still proven on
   linux/darwin (read the gate's test evidence).
6. One narrowing mutant per implemented behaviour, killed — re-apply one yourself. Validation log green.
accept_cr on revision 5, or changes requested with file:line. No LOGBOOK.md.
