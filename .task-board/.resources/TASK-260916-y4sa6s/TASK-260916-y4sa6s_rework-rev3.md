# Rework brief — TASK-260916-y4sa6s, revision 3 (E1)

Revision 2 closed R2–R7; the reviewer keeps R1 partially open
(`TASK-260916-y4sa6s_review-verdict-rev2.md`, "Required correction"). One
focused correction; everything else stays byte-identical.

## Correction (required)
The vector/oracle representation of MCP declarations cannot express optional
field ABSENCE: `tools/validate.py` (`e1_check_snapshots`, ~:5274–5301)
requires every snapshot to carry exactly the six keys and both optional
arrays, while §9.2 (as written in rev 2) compares the CCJ-1 bytes of the
declaration with "no field narrowed out" and "absent differs from present",
and the frozen `agent-mcp-v1` schema makes `env_names` and `environments`
optional (absent selector = every adapter). Result: a narrowed comparator
that pads absent optional arrays with `[]` survives the whole E1 vector gate
(0 cases with absent optional fields).

Do exactly this:
1. Represent snapshots as real declaration objects valid under
   `schemas/v1/agent-mcp-v1.schema.json` (validate them through the real
   Draft 2020-12 entry in the gate), preserving actual field presence — no
   padding of irrelevant transport fields with `null`, no mandatory six-key
   shape; per-transport consistency checks stay.
2. Add cases with A warning + migration hint / B refusal / flagged
   acceptance for: absent → present `environments` selector; present →
   absent selector; absent `env_names` → explicit `[]` (and the reverse).
   Keep the URL-only, selector-only and array-order cases.
3. Add the narrowing test the reviewer describes: a comparator that
   normalizes absent optional fields to `[]` MUST fail on the new cases
   (`tools/test_validate.py`).
4. Do NOT change the settled canonical-byte rule, the prose of §9.2, or the
   frozen MCP schema to fit the fixture abstraction; if the §9.2 wording
   needs a clarifying half-sentence about absence-vs-present, quote it in
   the evidence.
5. `make regenerate` (manifest, index, rc.9 pin), `make validate`, the
   disposable-copy regeneration proof; evidence "Revision 3" section;
   `TASK-260916-y4sa6s_spec-patch_rev3.patch`; the curator repository delta
   stays EMPTY (no LOGBOOK.md); `task-board handoff TASK-260916-y4sa6s --role doc-writer`.

Worktree and rules unchanged (`TASK-260916-y4sa6s_brief.md`,
`remediation-spec-producer-rules.md`).
