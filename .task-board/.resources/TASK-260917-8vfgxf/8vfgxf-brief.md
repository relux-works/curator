# TASK-260917-8vfgxf — interim NUL-opaque rule (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md`, this task's README and the Story README (STORY-260928-9vt338), and the content-hash story
STORY-260917-hbuawd's README for context. Security epic EPIC-260910-2hw1xb.

Rule (ships before content-hash v2): until curator uses the v2 framing, any REGULAR file containing a 0x00 byte, inside a skill or
context snapshot, in ANY directory, is a blocking opaque audit finding. The current v1 content hash cannot tell such a tree apart from its
colliding twin.
1. Find where the manager decides opaque/blocking audit findings for skill and context snapshots (install, update, resolve admission;
   cite the functions). Apply the rule at the production entry, whatever the directory or file extension. Symlinks and directories
   are unchanged.
2. Rows at the production entry:
   - the two colliding trees (two different trees with the same v1 content hash; build them from the v1 framing, as the Story README
     describes). Both must be refused with the opaque/blocking finding, naming the file;
   - a NUL-free tree installs as before;
   - a NUL file deep in a docs/ or assets/ path is still refused.
3. If curator-spec rc.13 already has wording or vectors for the interim rule, cite and drive them. Otherwise say so in the results
   (TASK-260917-2vapkz adds them to the spec).
4. Mutants, with real exit codes:
   - the rule limited to one directory;
   - the finding demoted to a warning.

   Each must survive before your change and be killed after it.
5. New manager-state reads go through internal/stateread. No CHANGELOG/LOGBOOK edit: put the entry text in the results under
   "## CHANGELOG entry (for release prep)". Never spell any employer name.

Update the results, then run `task-board handoff TASK-260917-8vfgxf --role developer`, then END YOUR TURN. The runner publishes the CR and
runs the gate; do not wait for it.
