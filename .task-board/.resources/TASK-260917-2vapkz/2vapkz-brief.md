# TASK-260917-2vapkz — spec: content-hash framing v2 (THE ONLY CURRENT INSTRUCTION)

curator-spec repository. Read `remediation-spec-producer-rules.md` (EPIC-260910-2hw1xb resources), this task's README and the Story
README (STORY-260917-hbuawd).
1. Revise protocol core.md §8 to curator-content-v2 framing:
   - the domain prefix `curator-content-v2` followed by 0x00;
   - then, per file in canonical path order: `F || uint64be(len(path)) || path || uint64be(len(bytes)) || bytes`. Define F, and define
     entries for directories and symlinks consistently with the existing §8 model.
2. Version the identity wherever the hash is carried: markers, context locks, verdict/pin state and registry records. Use either a
   prefix on the hash string or a `hash_version` member, choose one and state it normatively. Keep frozen v1 schemas untouched: add
   v2 shapes or new optional members only as the repo's schema rules allow, and say how.
3. registry.md: a record matches only when its framing version is equal. v1 identities never equal v2 identities.
4. Record the interim rule for v1 readers: a regular file containing 0x00 in a skill or context snapshot is a blocking opaque finding
   regardless of directory.
5. Vectors:
   - the two colliding trees, which MUST have different v2 hashes and whose v1 hashes collide (document the construction);
   - the empty tree;
   - at least one ordinary tree with its exact v2 hex;
   - a registry version-mismatch non-match case.

   Wire them into the manifest/schema as the repo's tools require.
6. Run the repo validators (`make` targets / tools/validate.py, and the regeneration check) and record the real exit codes. Add a
   CHANGELOG entry under Unreleased (this repo keeps a CHANGELOG). No LOGBOOK.md. Never spell any employer name.

Update the results, then run `task-board handoff TASK-260917-2vapkz --role developer`, then END YOUR TURN. Write only inside your Story
worktree.
