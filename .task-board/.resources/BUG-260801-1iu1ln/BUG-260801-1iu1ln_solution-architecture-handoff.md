# BUG-260801-1iu1ln solution-architecture handoff

## Outcome

Keep **BUG-260801-1iu1ln — bind-rc6-lifecycle-to-observed-csk-traces** as the single lifecycle-conformance unit and hand its existing signed cycle-10 candidate back to review under the accepted portable trust boundary. No child or sibling task is added.

The cycle-10 reviewer probe replaces a manager-owned native atomic-handoff callable and executes arbitrary same-principal `os.utime` work inside that trusted callable. Requiring the Python harness to detect every direct libc/WinAPI side effect from such a replacement is a kernel-observation or isolation guarantee, not lifecycle-field binding. The accepted v1 trust model explicitly includes the manager implementation, OS ownership/permission enforcement and manager-owned native primitives in the TCB, and excludes arbitrary same-principal code execution. The separately versioned hardened sandbox work is already tracked by **EPIC-260728-2m6dqo / STORY-260728-327soo** and is explicitly non-gating for portable delivery.

## Requirement traceability

| Remaining developer work | Concrete requirement |
| --- | --- |
| Re-review signed cycle-10 candidate `80b5b167...` against actual CocoaSkills seams and persistent state within the accepted manager TCB | Description and Acceptance Criteria: all 32 cases drive corresponding CocoaSkills seams and every normative vector field is checked against observed trace/state |
| Retain exhaustive scalar-leaf mutation rejection and fail closed when a new normative vector field lacks a binding/classification | Scope: exhaustive scalar-leaf mutation coverage or fail-closed normative-field classification |
| Treat replacement of a trusted native primitive with arbitrary same-principal direct libc/WinAPI behavior as outside this portable lifecycle harness | Accepted v1 cache trust model: manager/protocol/OS enforcement are TCB; arbitrary same-principal code execution is outside the install-time invariant |
| Preserve exact-root/full authenticated tests, strict mypy, diff/provenance, packaging and release guards already evidenced for cycle 10 | Acceptance Criteria: focused exact-root tests, strict mypy and diff checks pass |
| Preserve exact signed base and release exclusions | Scope and Acceptance Criteria: exact ba250bf base, signed integration commit, no pin/schema-v7/tag/release/claim change |

## Gap and scope audit

Sections checked: Description, Scope, Acceptance Criteria, cycle-10 reviewer verdict, explicit release-surface exclusions, and current dependency edge to TASK-260720-12r55p.

Result: no beyond-literal-spec element is justified. Extending this Bug to detect arbitrary direct native side effects would contradict the accepted trust model and duplicate the explicitly non-gating hardened sandbox scope. No research task is justified because the accepted trust decision and separate hardened work already resolve the boundary. No diagram materially improves this one-boundary decision.

## Dependencies and execution

The Bug is unblocked and remains the sole review pickup. Its existing `blocks TASK-260720-12r55p` edge correctly prevents shared-v6-vector consumption until review accepts the lifecycle binding. No dependency may be added from portable delivery to **STORY-260728-327soo — fail-closed-cross-platform-build-execution**, because that story explicitly prohibits such an edge.

Checklist items 23–27 were added before the operator trust-boundary directive was observed. Their “any mutation” interpretation is superseded by this decision: items 23–25 are dispositioned as out-of-scope when they require arbitrary direct same-principal native behavior; items 26–27 are already evidenced by the cycle-10 candidate. The remaining action is an evidence-based review under the clarified AC, not another implementation workaround.
