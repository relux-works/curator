# THE ONLY CURRENT INSTRUCTION — TASK-261017-sxydvh: reserved special-use hosts are not network contacts (developer)

The contract has landed and the owner has approved it. The normative source is curator-spec main c1a2bb256f10e583199f848bfb3a7a4d30e0d9bc (PR #159), environments §9.1, which §9.4 references for migration. Capability audits treat two kinds of host as non-network:

- loopback hosts;
- hosts whose normalized last label is `test`, `example`, `invalid` or `localhost` (RFC 6761).

Lookalikes, bare labels and example.com, example.net and example.org stay network contacts, so the static canary still fires. A mixed file still reports its real undeclared host. Where this brief and the spec disagree, the spec wins.

## Do this

1. Implement the rule in `internal/audit` and every audit path that classifies network contacts: member audit, migration audit and the audit cache. Today only localhost and 127.0.0.1 are exempt. Normalize the host the way the spec says before taking the last label, including case and a trailing dot.
2. Map `conformance/context-detectors/network-contacts.json` from curator-spec c1a2bb25 (40 vectors) into executable Go tests, row by row, and record the mapping in the outcome.
3. Add the negative controls the task names, plus a mutant per rule:
   - `x.test.evil.com` and a bare `test` are still flagged;
   - example.com, example.net and example.org are still flagged, and the canary `exfiltrate.example.com` still blocks;
   - in a mixed file, the real host is still reported;
   - mutant: exempt every host that contains `.test` anywhere; it must go red;
   - mutant: exempt example.com; it must go red.
4. If the audit cache keys on classification, a cached finding from before the change must not hide the new result. Prove it with a test.
5. Run the full local suite and the full hosted gate on the exact signed candidate. A runner infrastructure failure is a blocker to report, never a pass.

## Hand-off

Run `task-board handoff TASK-261017-sxydvh --role developer`, then end your turn.
