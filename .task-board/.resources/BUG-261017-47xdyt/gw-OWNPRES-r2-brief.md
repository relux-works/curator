# THE ONLY CURRENT INSTRUCTION — BUG-261017-47xdyt: rework against the landed ownership subset (developer)

Your previous run stopped on a contract conflict: the repair contract had to proceed over a schema-invalid marker. The owner's decision (option 3 of your record) is now normative in curator-spec main 3ca4b84f4eb3a1940fbccbe7f23e3b9987a6027e (PR #160):

- skillfile-sources §4.1 "Canonical ownership before publication";
- the source-marker section with the manifestless form in draft marker v6;
- the diagnostic `source_ownership_unproven` (§5).

The rule: a canonical node is Curator-owned only with a regular, readable, no-follow marker that parses as a JSON object, carries the source-marker schema identity and names this node.

- Invalid build or evidence fields inside such a marker mean re-derive. The existing repair contract and the external-evidence-mismatch rows stay green.
- An unparseable marker, a marker for another name, a link or a missing marker means unproven ownership. The operation refuses with `source_ownership_unproven` before any destructive step and names every such path.

Where this brief and the spec disagree, the spec wins.

## Do this

1. Rework the candidate in the existing Story workspace against that rule. The original brief `gw-OWNPRES-brief.md` stays in force: items A and C2, no generic takeover flag, native refusal and authorized-backup semantics unchanged.
2. Map the curator-spec cases into executable Go tests, row by row:
   - `conformance/draft-sources-v2/marker-ownership-cases.json` (21 cases, 17 refusals);
   - `conformance/draft-sources-v2/marker-writer-cases.json`.

   Record the mapping in the outcome. The candidate-conformance lane pins the rc.14 suite, so do not rely on it for the draft cases.
3. Keep the previous preservation controls and the five narrowing mutants. Add one mutant that treats a parseable marker naming another node as owned; it must go red for that reason.
4. Rerun the full local suite and the full hosted gate on the exact signed candidate. A runner infrastructure failure is a blocker to report, never a pass.
5. Update `BUG-261017-47xdyt_implementation.md` with a section "Round 2 against curator-spec 3ca4b84f".

## Hand-off

Run `task-board handoff BUG-261017-47xdyt --role developer`, then end your turn.
