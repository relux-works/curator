# Review brief: TASK-261021-vb3z6f, Draft CIP-0023 final-form exact-plan hosted launches (curator-spec)

Review the branch `exact-plan-final-form` at the exact head in the request, against `W-EXACT-brief.md`.

| Surface | Attack |
| --- | --- |
| Closure of the gap | The successor capability really carries detach, doorbell and the session record, so `werkstatt up` can leave `launch-delegated/1`. Walk one `werkstatt up` and one resume end to end through the text. |
| Snapshot binding | The exact plan is bound to the snapshot digest; the host never composes. Check consistency with CIP-0014 and the CIP-0016 body plus its amendment (now on main, or on branch `cip16-delegated-launches`). |
| Retirement | Per-runtime retirement of the delegated route: what is advertised, how curator-run selects, the refusal codes, and how the joint three-runtime rule is reconciled. |
| Project context | The CIP-0002 composition (stable home, lease, admitted generation) as a snapshot input. The CIP-0002 candidate is on branch `cip02-ready-for-acceptance` (head 752307c6); check that the assumed interface matches it, and report mismatches. |
| Lifecycle | Resume, recycle and GC of snapshots referenced by live sessions, including crash and unknown-liveness cases. |
| Scope and rules | No published wire schemas; owner questions carry recommendations; validator, links and prose rules; no personal paths. |
