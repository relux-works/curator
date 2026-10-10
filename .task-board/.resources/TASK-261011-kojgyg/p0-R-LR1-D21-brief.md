# THE ONLY CURRENT INSTRUCTION — TASK-261011-kojgyg: docs-only contract fix R-LR1-D21 (developer)

**Sources (read them first):**
- The contract table: `relux-works/curator` at 177cb554, file `.research/261010_platform-contract-table.md`. It is the single source of canonical values.
- The audit: `relux-works/curator` at 48789217, file `.research/261010_platform-docs-audit.md`. Use item R-LR1 in section D1, and every A-row it lists.
- The precondition `owner-decisions-20261010-audit.md`: D-GOALS, D-IDENTITY, D-EXECUTOR, D-PATH, D-DONOR-PROVISIONING, D-WIRE-NAMES, D-RESUME.

**Do this in this repository (curator-spec, decisions/0021 only):**
1. Fix every audit row of R-LR1 to the contract table's canonical value.
2. Where the table says OWNER DECISION NEEDED and no owner decision above covers it, write the open question explicitly in the spec, with its options. Do not invent values.
3. Apply in particular: Amend Decision 0021 with a versioned section that adds the agent-account one-shot exec mode of curator-run, keeping the interactive human mode unchanged. Rails (D-EXECUTOR, owner-confirmed): under the agent account curator-run never composes or resolves profiles; it only verifies and executes a plan composed on the caller side, through a dedicated entry such as `curator-run exec --plan-fd N`, with no network except the broker; one launch library is shared with swarma-session-runner. State the activation gate (qualification) and what the plan, transport, receipts and refusals are, by reference to the launcher spec.
4. Correct the existing diagrams that R-LR1 names, wherever they show wrong names or flows. New diagrams are a later P1 task.

**Rules:**
- Docs only: no code changes. Keep the spec's style, in English.
- No first names, local paths or session links. No tests or builds on this host.

**Hand-off:**
- Attach `TASK-261011-kojgyg_alignment.md` as an outcome resource: a table mapping each audit row of R-LR1 to the section changed and the canonical value or open question.
- Run `task-board handoff TASK-261011-kojgyg --role developer` and END YOUR TURN.
