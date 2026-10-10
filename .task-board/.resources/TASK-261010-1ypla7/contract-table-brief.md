# THE ONLY CURRENT INSTRUCTION — research TASK-261010-1ypla7: the platform contract table (researcher)

**Why.** The documentation audit (TASK-261010-2iqn63; read its newest change-request candidate of `.research/261010_platform-docs-audit.md`, or its outcome resource) found 37 contradictions across the book, the module specs, the CIPs and the launcher. Before parallel fixes start, all of them need ONE table to follow. Without it, the repositories drift again.

**Inputs.**
- The audit, sections A and D.
- The precondition `owner-decisions-20261010-audit.md`: D-GOALS and D-IDENTITY.
- Earlier owner decisions recorded in the specs and CIPs:
  - final names: `swarma-credential-broker`, `swarma-user-manager` with binaries `swarma-user-manager` and `swarma-user-launcher`, `swarma-dispatcher`, `swarma-session-host`, `swarma-session-runner`;
  - `curator-run` for one-shot runs under agent accounts; agent accounts are `worker-<label>`;
  - everything under `/opt/swarma`, sockets in `/opt/swarma/run/<base>/`;
  - CIP-0003…0007 accepted, with the CIP-0007 conditions; CIP-0002 operator input;
  - CIP-0009 rev3 variants A/B/C, with DNSSEC as the default and the invitation as the fallback;
  - SSH host CA plus user CA;
  - credential source (a) per-home now, v2 requires an explicit entry.
- The specs at the pinned commits the audit uses.

**Output.** `.research/261010_platform-contract-table.md` with one row per contract item:
- names: repository, binary, service account, socket path;
- signed request domains and versions: dispatcher, broker, user manager;
- identifier kinds: root correlation ID versus per-effect IDs, generation, epoch, launch ticket;
- size limits: helper request, launcher request, decoded plan;
- process ownership: service, session runner, PTY, Codex app-server, credential renewal, stop;
- credential provider × source × channel × phase, and the first broker consumer;
- goal authority, per D-GOALS (task-board is OPTIONAL; no coupling of the session host to goals or to the orchestrator);
- identity revocation, per D-IDENTITY;
- capacity versus quota, v0 versus v1;
- launch-scoped PATH, as one combined recipe;
- donor trust bootstrap.

**Columns:**
1. canonical value;
2. the authoritative document and section that will own it;
3. every other document that must match;
4. the current mismatches, by audit row ID;
5. status: decided (cite the decision) or OWNER DECISION NEEDED (state the options; do not invent).

Cite every value to file and line at a pinned commit. No first names, no local paths, no secrets: this file is public. Keep it under 64 KiB. No `LOGBOOK.md` edits, no tests or builds on this host. Then run `task-board handoff TASK-261010-1ypla7 --role researcher` and END YOUR TURN.
