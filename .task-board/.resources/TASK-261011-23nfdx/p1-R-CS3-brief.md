# THE ONLY CURRENT INSTRUCTION — TASK-261011-23nfdx: module diagrams M-CIP1..M-CIP2 (R-CS3, developer)

**Sources:**
- The audit: `relux-works/curator` at 48789217, `.research/261010_platform-docs-audit.md`, section B4, rows M-CIP1..M-CIP2. The rows name the exact files, kinds and participants.
- This repository's spec and README on current main, which are aligned with the contract table after the P0 fixes.
- The contract table: `relux-works/curator` at 177cb554, `.research/261010_platform-contract-table.md`.

**Do this:**
1. Author or extend every figure of rows M-CIP1..M-CIP2 as PlantUML sources under `diagrams/`.
2. Each figure must show:
   - the phase (current, SH1/SH2, v0/v1/v2);
   - the identity of every process;
   - who owns each piece of persistent state;
   - the refusal and recovery branches;
   - a link to the defining spec section.
3. Draw only what the spec says. Where the spec is silent, put a visible note ("unspecified in spec §X") instead of inventing behaviour.
3a. Also resolve the follow-ups of `TASK-261011-2xdx9u_second-review-record.md` (on that task):
   - F1: in `cips/platform-contract-disposition.md` and CIP-0002, restate the contract-table X1 rule (deduplicate by identity; preserve the existing position of the manager dispatcher directory), or list it as an open grammar detail.
   - F2: remove the private input file name from the shared document.
   - F3: label the wire-name rows that are derived by rule as "derived by rule, pending matching module vectors".
3b. Add at the top of CIP-0002 to 0011 a short operative summary table: purpose, boundary, authority owner, inputs and outputs, normative links. Keep the existing disposition blocks.
4. Link each figure from the relevant spec section and from the README. If you change an existing figure, keep it consistent with the spec and do not leave contradicting copies.
5. Keep PlantUML syntax conservative; there is no renderer on this host.

**Rules:**
- Docs only. No first names, local paths or session links.
- Do not run tests or builds yourself. The configured CR validation at handoff is sanctioned.
- Never write LOGBOOK.md.

**Hand-off:** attach `TASK-261011-23nfdx_figures.md` (a table of figure → file → spec section → audit row), then run `task-board handoff TASK-261011-23nfdx --role developer` and END YOUR TURN.
