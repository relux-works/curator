# Review brief: protocol/environments.md draft 1

## Subject
- Worktree: `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-environments-normative`, branch `draft/environments-protocol`, base `2a861e5` (= main), head `eddd509`
- Document: `protocol/environments.md` (947 lines, new file; docs-only delta)
- Producer notes: board resource `environments-doc-draft-notes.md` on TASK-260901-2tdoy5 — read fully; it lists resolved carry-forwards and deliberate deviations to judge.

## Review dimensions
1. **Coverage vs Decision 0010** (decisions/0010-agent-environment-profiles.md on main): every revision-1 surface of the decision has normative prose; nothing out-of-scope leaked in (launcher internals, ax edits, MCP write, path-kind mechanics, schemas files).
2. **Byte-rule soundness**: generation header grammar, chapter separators, monolithic join, zero-modules shape, referenced layout (`.agent-context/modules/<profile>/<path>`), system-prompt output. Check determinism claims hold (no ambiguity two implementations could read differently), collision arguments are correct (the `modules/` literal guard), and the §8 content-hash binding is well-defined.
3. **Deliberate deviations** (producer notes section "Additional normative resolutions"): header URL choice, system-prompt output WITHOUT header/chapters (verbatim bytes to the model), flat non-recursive composition, marker-as-ledger split from §11 adapter ledger, `system_prompt_files: off|append|replace` machine setting. Verdict each: sound / needs change — these narrow or concretize the decision and must not contradict it.
4. **Cross-references**: every cited core/manager/decision section exists and says what is claimed (spot-verify all § citations).
5. **House style**: MUST/MUST NOT discipline, closed sets, diagnostics tables per section, prose density matching protocol/core.md and assurance.md; separately-versioned-surface framing (adds identities, widens nothing).
6. **Diagnostics**: stable code families complete and non-overlapping; every failure path in the prose has a code; no code without a defined condition.
7. **Consistency with the shipped repo**: no contradiction with existing schemas/vectors/manager profile; `make validate` green claim spot-checked.

## Constraints
Read-only: no edits, no commits. Shell tooling; both reference checkouts read-only (`~/Developer/ReluxWorks/curator-spec` main, decisions/0010 included). Ignore `tools/__pycache__/`.

## Verdict contract
`review-findings-environments-1.md` resource on TASK-260901-2tdoy5: severity (blocking|major|minor|nit), section, quote, what is wrong, suggested fix. Blocking/major → set task status development and say so. Otherwise ACCEPT explicitly, leave to-review. Check the task DoD items you verified. Do not mark done.
