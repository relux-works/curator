# Delta review: TASK-261021-vb3z6f rework r2, Draft CIP-0023 (curator-spec)

Review r1 by rev1 (`TASK-261021-vb3z6f_review-r1.md`) found P1 F1: a capture-only lease could be released, while CIP-0002 requires the original-home session lease to be retained until termination and reconciliation. The fix is at the new head in the request (branch `exact-plan-final-form`, one new commit on 0c1f830f).

1. Review the delta `0c1f830f..<head>`, then re-sweep the surfaces it touches: lease ordering, crash and unknown liveness, busy-update refusal, both GC roots.
2. Cross-check every lease sentence against the CIP-0002 candidate head f202f41b (branch `cip02-ready-for-acceptance`), which you reviewed and accepted. Any remaining disagreement is a finding.
3. The test-plan rows: the session lease survives a capture release; a busy update refuses; GC keeps roots. They must discriminate, not pass trivially.
4. Rerun make validate in a normal-.git clone.
