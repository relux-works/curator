# THE ONLY CURRENT INSTRUCTION — TASK-261011-1pvp8b: module diagrams M-LR1..M-LR2 (R-LR2, developer)

**Sources:**
- The audit: `relux-works/curator` at 48789217, `.research/261010_platform-docs-audit.md`, section B4, rows M-LR1..M-LR2. The rows name the exact files, kinds and participants.
- This repository's spec and README on current main, which are aligned with the contract table after the P0 fixes.
- The contract table: `relux-works/curator` at 177cb554, `.research/261010_platform-contract-table.md`.

**Do this:**
1. Author or extend every figure of rows M-LR1..M-LR2 as PlantUML sources under `diagrams/`.
2. Each figure must show:
   - the phase (current, SH1/SH2, v0/v1/v2);
   - the identity of every process;
   - who owns each piece of persistent state;
   - the refusal and recovery branches;
   - a link to the defining spec section.
3. Draw only what the spec says. Where the spec is silent, put a visible note ("unspecified in spec §X") instead of inventing behaviour.
3a. Include the target agent-account one-shot exec path (curator-run exec): the contract amendment landed (Decision 0021 amendment v1; launcher SPEC at 1d608047).
4. Link each figure from the relevant spec section and from the README. If you change an existing figure, keep it consistent with the spec and do not leave contradicting copies.
5. Keep PlantUML syntax conservative; there is no renderer on this host.

**Rules:**
- Docs only. No first names, local paths or session links.
- Do not run tests or builds yourself. The configured CR validation at handoff is sanctioned.
- Never write LOGBOOK.md.

**Hand-off:** attach `TASK-261011-1pvp8b_figures.md` (a table of figure → file → spec section → audit row), then run `task-board handoff TASK-261011-1pvp8b --role developer` and END YOUR TURN.
