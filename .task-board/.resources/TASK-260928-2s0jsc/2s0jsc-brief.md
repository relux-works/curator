# TASK-260928-2s0jsc — `curator global adopt <command>` (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md` and the task description. Context: TASK-260923-xq4pjj replaced the hand-written ~/.local/bin/task-board
with bytes equal to runtimestore.UnixShimContent for the canonical target, but internal/globalbins refuses any existing entry absent from
.curator-managed.json (unmanagedConflict), and only Curator writes that marker — there is no way to take an already-canonical shim under
management. profiles/manager.md §3: a manager "MUST NOT overwrite an unmanaged entry … Forwarding locations and their ownership records are
machine-local and implementation-specific" — so this is a curator-only command, no spec change.
Implement `curator global adopt <command> [--dry-run]`: the command must be a Curator global command with a canonical target; the existing
user-bin entry must be a regular file (lstat; symlink/special → refuse) whose bytes equal the canonical shim (Unix shim / Windows .cmd) for
that target; on match, back it up (cp -p equivalent under the Curator backups root), record it in the ownership marker atomically, print
what was adopted; on any mismatch refuse with a clear diagnostic naming the path and the reason and write nothing. `global install` then
treats it as managed. Tests (unix + windows): adopt succeeds on byte-identical; refuses differing bytes / symlink / unknown command /
missing entry with no write; idempotent second adopt; dry-run writes nothing; mutant (byte comparison skipped) killed. docs/cli.md and
docs/troubleshooting.md (the "unmanaged conflict" recovery). Manager-state reads via internal/stateread. No CHANGELOG/LOGBOOK edit (entry in
results). Update the results resource, `task-board handoff TASK-260928-2s0jsc --role developer`, END YOUR TURN. Write only inside your
Story worktree.
