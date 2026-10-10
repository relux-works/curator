# THE ONLY CURRENT INSTRUCTION — TASK-261011-3uoh3h: docs-only contract fix R-LR1-SPEC (developer)

**Sources (read them first):**
- The contract table: `relux-works/curator` at 177cb554, file `.research/261010_platform-contract-table.md`. It is the single source of canonical values.
- The audit: `relux-works/curator` at 48789217, file `.research/261010_platform-docs-audit.md`. Use item R-LR1 in section D1, and every A-row it lists.
- The precondition `owner-decisions-20261010-audit.md`: D-GOALS, D-IDENTITY, D-EXECUTOR, D-PATH, D-DONOR-PROVISIONING, D-WIRE-NAMES, D-RESUME.

**Do this in this repository (curator-agent-launcher, SPEC.md and README only):**
1. Fix every audit row of R-LR1 to the contract table's canonical value.
2. Where the table says OWNER DECISION NEEDED and no owner decision above covers it, write the open question explicitly in the spec, with its options. Do not invent values.
3. Apply in particular: Specify in SPEC.md and README the one-shot exec mode: immutable plan input and its digest, approved transport (descriptor), caller versus executor authority, broker lease binding (generation, profile, harness, binary digest), PATH per D-PATH, signals and stop, exit codes, receipts and typed refusals. Keep the interactive scope unchanged. Add a short front-door map: direct human launch, hosted/ax launch, and agent one-shot exec. Do not claim the mode is implemented.
4. Correct the existing diagrams that R-LR1 names, wherever they show wrong names or flows. New diagrams are a later P1 task.

**Rules:**
- Docs only: no code changes. Keep the spec's style, in English.
- No first names, local paths or session links.
- Do not run tests or builds yourself. The repository's configured Change Request validation suite runs automatically at handoff, and that is sanctioned: on this host its Go step is compile-only (rule R223). Hand off normally. Do not put board-resource links in repository files; the alignment table is a board resource only.

**Hand-off:**
- Attach `TASK-261011-3uoh3h_alignment.md` as an outcome resource: a table mapping each audit row of R-LR1 to the section changed and the canonical value or open question.
- Run `task-board handoff TASK-261011-3uoh3h --role developer` and END YOUR TURN.
