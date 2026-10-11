# THE ONLY CURRENT INSTRUCTION — TASK-261011-2xdx9u: docs-only contract fix R-CS2 (developer)

**Sources (read them first):**
- The contract table: `relux-works/curator` at 177cb554, file `.research/261010_platform-contract-table.md`. It is the single source of canonical values.
- The audit: `relux-works/curator` at 48789217, file `.research/261010_platform-docs-audit.md`. Use item R-CS2 in section D1, and every A-row it lists.
- The precondition `owner-decisions-20261010-audit.md`: D-GOALS, D-IDENTITY, D-EXECUTOR, D-PATH, D-DONOR-PROVISIONING, D-WIRE-NAMES, D-RESUME.

**Do this in this repository:**
1. Fix every audit row of R-CS2 to the contract table's canonical value.
2. Where the table says OWNER DECISION NEEDED and no owner decision above covers it, write the open question explicitly in the spec, with its options. Do not invent values.
3. Apply in particular: Write an operative disposition block at the top of CIP-0002 to 0007, 0010 and 0011: decided, open, superseded. Mark superseded default/Keychain/approval questions as superseded, with a pointer. Apply D-PATH as the one combined PATH recipe, reconciling CIP-0002, CIP-0004 and CIP-0007. Apply D-EXECUTOR: the first broker consumer is curator-run, and board and session host go through the dispatcher. Apply D-WIRE-NAMES. Add one cross-CIP table of schema, owner and version. Accepted headers must not silently imply runtime adoption. Keep historical evidence as historical.
4. Correct the existing diagrams that R-CS2 names, wherever they show wrong names or flows. New diagrams are a later P1 task.

**Rules:**
- Docs only: no code changes. Keep the spec's style, in English.
- No first names, local paths or session links.
- Do not run tests or builds yourself. The repository's configured Change Request validation suite runs automatically at handoff, and that is sanctioned: on this host its Go step is compile-only (rule R223). Hand off normally. Do not put board-resource links in repository files; the alignment table is a board resource only.

**Hand-off:**
- Attach `TASK-261011-2xdx9u_alignment.md` as an outcome resource: a table mapping each audit row of R-CS2 to the section changed and the canonical value or open question.
- Run `task-board handoff TASK-261011-2xdx9u --role developer` and END YOUR TURN.
