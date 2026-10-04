# BUG-261004-16a407 — trust-pin-overrides-strict-findings: logbook note

The production audit entry (`Gate` and dry-run `GateReadOnly`) allowed pinned findings on fresh and cached verdict paths because `decideWithPins` returned allow before applying `Decide`. Removing that early return restricts pins to satisfying the legacy-schema pin requirement. Revocation remains earlier in the pipeline.

The new production regression matrix has 52 cases (13 policy scenarios × 2 cache states × 2 gate entries). The original code fails 28 cases (exit 1); the fix passes 52/52 (exit 0). Restoring early allow fails 28 cases; bypassing only cached or only fresh decisions fails 14 cases each (all mutant tests exit 1).

The explicit task instruction forbids editing LOGBOOK.md and CHANGELOG. This task-scoped note records the finding and decision on the board without modifying those files. The spec clarification is attached separately for its owner; the spec repository remains untouched.

Hosted validation exposed `TestDraftAuditPinAdmits`, an existing install test defending the pin bypass. The macOS served stage and Go suite exited 1. The test now requires refusal through production install.Project, fresh and cached, with no source-audit binding established. Restoring the old early allowance fails both updated install subcases (exit 1). The initial hosted attempt was cancelled after diagnosis; the updated test tree is being validated in a new hosted run.

The install absence assertion was tightened to fail on unexpected read errors. The exact candidate regression passed 52 audit cases plus 2 install.Project cases (exit 0), and exact-candidate build/vet/isolated-cache lint all exited 0. Hosted run 37170471503 passed 11/11 required jobs on snapshot cc3ba67211bc533558c3a6d0fe8bb6bce2ad7a2f. Every Test/Race lane reported go test exit 0 and platform-case gate exit 0. A host command-execution delay resolved; no host settings were changed by this run. Source and tests still match the hosted snapshot (diff exit 0).
