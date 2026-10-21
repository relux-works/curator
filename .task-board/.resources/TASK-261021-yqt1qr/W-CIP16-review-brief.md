# Review brief: TASK-261021-yqt1qr, CIP-0016 delegated-launch amendment (curator-spec)

Review the branch `cip16-delegated-launches` at the exact head named in the request, against the producer brief `W-CIP16-brief.md` and its source `front-door-appendix-A.md` (both board resources).

| Surface | Attack |
| --- | --- |
| Fidelity | Every element of Appendix A is present with the same meaning: division of work, refusals on a fragment-digest mismatch, no added content and no Remote Control, the stated limits (no snapshot isolation, "a host never composes" holds only for the daemon, provenance replaced), the retained rules, per-runtime activation, retirement and backend selection. Find any silent weakening or strengthening. |
| Owner status | The amendment is dated 2026-10-21 and recorded as operator-accepted. The two conditions (home lease, one profile per Werkstatt) are clearly marked pending and are not normative. |
| Consistency | No contradiction with CIP-0014, the CIP-0016 body, `protocol/local-hosted-launches.md`, the platform contract disposition, or the curator-run FRONT-DOOR §5.8 activation table. |
| Scope | No new wire schemas. Only the files the brief allows changed. README and CHANGELOG follow how earlier amendments did it. |
| Repo rules | Prose rules (no em dashes or guillemets), links and anchors resolve, the validator passes (rerun it), no personal paths or names. |
