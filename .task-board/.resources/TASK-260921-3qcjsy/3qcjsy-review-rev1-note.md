# Review note for TASK-260921-3qcjsy revision 1 (orchestrator, binding) — adopt 0017/0018

Repository curator-spec. Brief 3qcjsy-brief.md (R1 0017 options/open questions, R2 0018 with the
2026-09-16 amendment, R3 amendments + follow-ups). Gate green (local `make validate`) — verify
the gate ran on the exact revision-1 tree. The patch is large (~170 paths): the two decision
documents, `profiles/manager.md`, `protocol/environments.md`, `cli/curator.md`, schemas
`manager-config-v2` / `system-config-v2` + regenerated schema-cases and `vectors/manager-config-v2.json`,
`release/1.0.0-rc.9.json`, tools (generator + validate), COMPATIBILITY, UNRESOLVED_QUESTIONS, CHANGELOG.

Judge, with your own reruns (disposable clone; `make validate`; retry once on host stalls):
1. Decision documents: Status adopted with a dated note; every option and open question has a
   recorded choice and the choices match results.md's two tables exactly; the choices follow the
   brief's rulings (recommended options verbatim; fail-closed/least-privilege where none) —
   flag any choice that widens privilege or contradicts the manager profile.
2. Normative amendments: are they the amendments the decisions themselves call for, no more?
   Enumerate every schema change (new/changed properties in manager-config-v2 and
   system-config-v2, lockability direction, grammar) and check each against the decision text;
   regenerated cases must come from the generator (`tools/generate-vectors`), not hand edits;
   `release/1.0.0-rc.9.json` / manifest changes must be consistent with the release policy of the
   repository (a released revision must not be silently rewritten — say whether this edit is
   allowed by the repo's own rules or needs a new release entry).
3. COMPATIBILITY.md / UNRESOLVED_QUESTIONS.md updates truthful; cross-references resolve;
   CHANGELOG entry.
4. Follow-up leaves per repository (results.md §) have one-line ACs and map 1:1 to what the
   adoption defers (manager repairs + migration, launcher permission interface, fleet
   enforcement, macOS claude_code experiment, capability token values).
5. Anything in the patch outside the brief's scope → list it; block only if it changes
   normative behaviour the decisions do not cover.
Record exactly one verdict: accept_cr(TASK-260921-3qcjsy, revision=1, evidence=<your outcome
resource>) on ACCEPT, or changes_requested with file:line and reproduction.
