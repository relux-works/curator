# Reviewer verdict for TASK-260811-3twayo

Verdict: changes requested -> to-dev

Reviewer run: RUN-260822-615b9e
Goal check: task-board spawn goal reported that this run is not goal-bound.
Reviewed outcome: TASK-260811-3twayo_node-runtime-build-contract.md

## Findings

1. High - The required C0 Node and selected-manager authority is absent. Bind constructs both toolchain nodes, but lines 266-280 of internal/nodesource/nodesource.go mark every tool as ToolchainBoundAtC4 and return only C4Selectors. No exact C0 checkpoint is accepted or returned, and nodesource has no pre-C5 metadata derivation permit or receipt integration. This contradicts the task scope requiring Node and manager fingerprints at C0 and permits and receipts for pre-C5 metadata. Rework must bind the evidence executables through exact C0 checkpoint membership and route every executable C1-C4 metadata derivation through closureexec permits, immediate rechecks, and receipts.

2. High - Declared TypeScript or generator lineage is not connected to a closed graph or build plan. BuildCapture at lines 92-159 emits only package and command-product records. BuildGeneratedAction at lines 306-350 returns an action, compiler node, output nodes, and edges as a disconnected helper; no API adds those records before C4, projects them into ActiveGraph, or produces closuregraph.BuildPlan/C5 evidence. TargetNodeID is only checked for syntactic validity at line 307 and is never used to emit a targets edge, so target lineage is not proven. Rework must integrate declared actions and immutable expected outputs before C4, bind host/target and tools in SelectionBinding, project the exact active records, and generate/validate a deterministic BuildPlan without adding graph records at C5.

3. High - Common capture identity is still dependent on manager parser ordering. Package records are sorted, but dependencies are iterated in caller order and their positional index becomes EdgeKey at lines 120-135. Equivalent npm, pnpm, Yarn Classic, and modern Yarn normalized inputs with different dependency ordering therefore produce different edge and capture identities, despite the manager-independent canonical capture requirement. Sort and uniquely validate normalized dependency declarations by canonical semantic fields before assigning stable edge keys, and add a permutation test across manager profiles.

4. Medium - Exact output-set validation has an acceptance hole. ValidateObservedOutputs at lines 354-363 compares counts and membership only. Duplicate declared paths are not rejected, so declarations containing the same path twice can accept an observed map containing that path plus one undeclared extra path. Observed digest validity is also unchecked. Reject duplicate declarations and invalid observations, then reconcile each exact declared path/class/grammar and observed content identity through the shared output/receipt contract.

5. High - The named task conformance is not demonstrated. internal/nodesource/nodesource_test.go contains four broad tests and no Node-bound CGP05 exact-record, CGP10-CGP11, CGN09, CGN15-CGN18, N04-N10, or N13 cases. Existing shared closuregraph/closureexec goldens pass independently, but nodesource is not wired into them. The Python script emits only three ad hoc hash vectors, not the independent Python protocol golden corpus required by the AC. Add end-to-end Node common-profile vectors that prove checkpoint boundaries, zero process starts on negative cases, selected/pruned conditions, distinct plan identities, lifecycle suppression, shipped/local generated code, native/compiled rejection, runtime drift, exact outputs, and independent Python compatibility without shared Go code/state.

## Verification

- go test -count=1 ./internal/nodesource ./internal/closuregraph ./internal/closureexec ./internal/artifactpolicy: passed.
- go vet ./internal/nodesource: passed.
- git diff --check -- internal/nodesource: passed.
- go test -run ^$ ./...: passed repository-wide compilation.
- No product code was modified by this reviewer.

The focused tests are green, but they do not exercise the missing security and integration contracts above.