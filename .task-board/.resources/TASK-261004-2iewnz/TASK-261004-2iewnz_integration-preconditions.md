# Integration preconditions — 2026-10-05

TASK-261004-2iewnz — research-105-design-and-implementation-plan; STORY-261004-3v2zm2 — legacy-provider-settings-and-mcp-optouts-105.

The current integration assignment supersedes the older direct-integration instruction. No integrate, checkpoint, generic handoff, status mutation, or repository edit was performed. The bound runner owns landing after this run exits.

Fresh read-only evidence:
- Board get query: exit 0; task and Story remain integrating.
- task-board worktree integrating --json: exit 0; CR-TASK-261004-2iewnz-1 revision 1 is accepted, kind story_final, not deferred, awaiting_landing. Candidate tree: 1bc879cf3414fe2e17d1c5b6f5961f9a68390788. Fresh protected main OID reported: 9bc8e1a1eace93377e41c36465b26a58ee5ce05a. Candidate not yet on trunk.
- Standalone Python comparison: exit 0, 4/4 assertions passed. Each of the two research files was compared byte-for-byte with git show <candidate-tree>:<file>; git diff HEAD --name-only was empty; git ls-files --others --exclude-standard contained exactly those two files.
- CIP SHA-256: f0d8aa67899a565fe83d93e0a31802f8861b47f9594f1234c35e56b1c9b1ace0.
- Companion evidence SHA-256: 48af5bd302b47bfc8f931ed639bf3afd6dc782b2a01caa001afaaa64a45e003d.

No product tests or historical R1 scripts rerun: this run checks landing preconditions for unchanged accepted bytes. Product Go gates remain out of scope under hosted-evidence mode. LOGBOOK.md untouched. No credentials accessed.

Limits: these observations do not attest successful landing or replace the runner transaction gates. The global read-only report also showed one unrelated uncommitted board activity entry; its effect on landing is unverified. An initial unsupported schema(operation=change_request) discovery query exited 1; schema discovery then exited 0. No gate refusal was bypassed.
