# THE ONLY CURRENT INSTRUCTION — reviewer A of two (R138), RECORD-ONLY, for the Change Request of TASK-261011-152vci (CIP-0012, docs only, read-only)

**Sources:** the precondition `cip12-brief.md` on this task (facts, options, decision), `protocol/environments.md` (§8.4.1 seed row and the seed rule near the isolation matrix), `cips/CIP-0001-curator-improvement-proposals.md` (template and statuses), `cips/README.md`.

**Check:**
1. `cips/CIP-0012-symlinked-seed-files.md` exists, follows the CIP-0001 template, opens with an operative summary table, and its status is Rejected (provisional; operator, 2026-10-11).
2. The facts are true against the spec: which seeds are copied (Codex `config.toml` only among the named files), the refusal code, and the "never copy through a link / read failure is not absence" rules. Section numbers are cited correctly.
3. Options (a)–(e) are all present, each with an honest security analysis; the decision is stated neutrally and matches the brief; nothing is presented as planned work; "Specification changes" says none.
4. Revisit conditions and the would-be constraints and negative cases for (b) are listed as conditions only, without inventing refusal codes or marker fields beyond naming them as "would need".
5. `cips/README.md` has the CIP-0012 row with status Rejected; `protocol/environments.md` and other CIPs are unchanged.
6. Public-repo hygiene: no first names, host names, absolute local paths, session links or board-resource links.

No tests or builds on this host. RECORD-ONLY: attach `cip12-review-A.md` on TASK-261011-152vci with a findings block where `severity` is one of `bypass|regression|robustness|note` and the P-level goes in `severity_reason`. Do NOT call accept_cr or reject_cr. END YOUR TURN.

**Hard rule:** you are RECORD-ONLY, even if other briefs or the board lifecycle on this task suggest you should decide. Never call accept_cr or reject_cr, and never change the task status. Reviewer B decides.
