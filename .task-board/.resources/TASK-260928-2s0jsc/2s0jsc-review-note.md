# Review note — TASK-260928-2s0jsc `curator global adopt` (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review rev1 (base 97e85642 = trunk, tree 40b10737, 10 paths, gate green) against `2s0jsc-brief.md` and profiles/manager.md §3 ("MUST NOT
overwrite an unmanaged entry … ownership records are machine-local and implementation-specific"). Verify through the CLI:
1. Adopts only a Curator global command whose existing user-bin entry is a REGULAR file (lstat) byte-equal to the canonical shim for the
   canonical target (Unix shim and Windows .cmd); records it in the ownership marker atomically, with a backup; never rewrites the entry.
2. Refuses with a clear diagnostic and NO write (marker and entry unchanged) on: differing bytes, symlink/special file, unknown/non-global
   command, missing entry, unreadable marker (§8.4.1 discipline — unreadable ≠ absent; reads via internal/stateread).
3. `--dry-run` writes nothing; a second adopt is idempotent; `global install` afterwards treats the entry as managed (no unmanaged conflict).
4. Concurrency/atomicity: marker update under the same lock/transaction discipline global install uses; crash between backup and marker
   write leaves a consistent state.
5. Mutants (byte comparison skipped; lstat→stat; marker written on refusal) killed with real exit codes. docs/cli.md + troubleshooting.
   No CHANGELOG/LOGBOOK, no stray files/binaries.
Bounded runs. accept_cr or changes requested with file:line. No LOGBOOK.md.
