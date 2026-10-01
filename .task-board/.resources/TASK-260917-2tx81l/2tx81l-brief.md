# TASK-260917-2tx81l — curator: content-hash framing v2 (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md`, this task's README and STORY-260917-hbuawd's README. The normative text is curator-spec main b1a2efb
(PR #116): core.md §8 curator-content-v2, the hash_version carriers, registry.md equal-version matching, and the interim NUL-opaque
rule (already on curator main via TASK-260917-8vfgxf). Curator main already accepts that suite (TASK-260930-12i5zr, 71360fcb): the v2
families are known-gap rows owned by THIS task, active only under the candidate manifest digest 950ee74a… A new spec release is not
cut yet. Run with CURATOR_CONFORMANCE_ROOT pointed at a disposable clone of curator-spec at b1a2efb, and also at rc.13 (SPEC_PIN).
1. Implement the v2 framing in internal/hashing behind an explicit hash version: domain prefix, per-entry F, uint64be length framing,
   canonical order, empty tree. Check it against content-hashes-v2.json. v1 stays byte-identical for v1 readers.
2. Carry and compare the version everywhere the spec requires:
   - install markers (v5);
   - context locks (v2);
   - agent environment markers (v3);
   - audit records (v2);
   - manager-config waiver state (v3);
   - registry log entries / bundles / log-response (v2 / v2 / v3), with matching at internal/registry/registry.go:298.

   v1 identities never equal v2 identities. Writers emit v2 shapes under a gate the spec defines. If the spec leaves the write
   cut-over to the manager, state your choice and keep reading v1.
3. Drive the v2 families. Every known-gap row this task owns must LEAVE conformance-gaps.tsv as it becomes driven (the ratchet rejects
   passing gaps). Report before/after counts per family under the candidate digest; rc.13 stays green and unchanged.
4. Mutants, with real exit codes:
   - length framing removed → the colliding pair collides;
   - the version not compared in registry matching;
   - a v1 identity accepted where v2 is required.

   Each must survive before your change and be killed after it.
5. New state reads go through internal/stateread, managed writes through the E5 nofollow helpers. No Windows-reserved file names. No
   CHANGELOG/LOGBOOK edit: put the entry text in the results under "## CHANGELOG entry (for release prep)". Never spell any employer
   name.

Update the results, then run `task-board handoff TASK-260917-2tx81l --role developer`, then END YOUR TURN. The runner publishes the CR and
runs the gate; do not wait for it.
