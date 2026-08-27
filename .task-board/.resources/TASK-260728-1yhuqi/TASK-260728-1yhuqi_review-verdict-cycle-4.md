# TASK-260728-1yhuqi — review verdict, cycle 4

**VERDICT: CHANGES REQUESTED**

Route: `analysis` for security-contract and execution-policy rework.

Run: `RUN-260729-906f3c` (reviewer). `task-board spawn goal` reported that the
run is not goal-bound.

## Summary

Cycle-3 closes the two defects from cycle 3 at the mechanism level: the closed
per-job grammar rejects all 16 unknown-channel vectors, deleting the measured
plugin channel prevents `@Observable` from loading `ObservationMacros`, ordinary
Swift still builds, and the independent probe replay is green.

Acceptance is nevertheless withheld because the replacement pipeline no longer
fits the accepted `manager-worker-v2` session it continues to bind into the
canonical build input, receipt, marker, and claim. It also does not provide the
pre-compile package-macro rejection required by accepted Decision 0008.

## Blocking finding 1 — the compile phase is multiple commands, but
## `manager-worker-v2` admits exactly one

Accepted Decision 0008 is explicit:

- one v2 session performs at most one graph command and then “exactly one
  compile phase of exactly one driver-defined command”;
- the session admits no additional executable or tool request;
- a driver that cannot map to that shape is not admitted under v2; widening the
  shape requires another execution-policy identity, a new claim schema version,
  and review.

The cycle-3 Decision 0011 repair instead says:

- `compile_argv` is not executed;
- after the permit, the manager executes every verified plan job sequentially;
- `swift-frontend` and `clang` are started directly by the manager.

This is observable, not editorial. Independent replay produced:

- `S46`: 3 executed jobs for the default two-source build;
- `S47`: 2 executed jobs for the one-source ordinary-Swift build;
- the graph phase is a separate `swiftc -###` command.

The compile command cardinality therefore varies with the emitted plan and is
never the single driver-defined compile command Decision 0008 permits. The
parentage also changes from
`worker -> trusted launcher -> tool executables` to direct manager execution of
the plan tools during compile.

Decision 0011 nevertheless keeps
`policy.execution_policy = "manager-worker-v2"`. That makes the execution-policy
label, process graph, and identity semantics disagree. Calling a sequence of
two or three commands “one compile phase” does not satisfy the closed command
cardinality that caused v2 to be minted.

## Blocking finding 2 — macro use reaches a compile child instead of being
## rejected before the compile permit

Decision 0008 requires the deterministic pre-compile matrix to reject every
package-selected compiler macro before the compile phase, under
`build_package_code_execution_forbidden`. A surface that cannot be rejected
there disqualifies the driver.

The replacement successfully removes the load capability, but it does not
reject macro-selecting source before compile:

- `S48` records graph exit 0;
- the closed plan verifier accepts the plan;
- the manager grants the compile permit and starts the first frontend job;
- only that compile child parses `@Observable` and exits 1 with the missing
  implementation diagnostic.

Zero plugin-load remarks is an important security improvement, but it is not
the required Stage-B rejection of the package-selected macro surface. The
contract's diagnostic table assigns `build_package_code_execution_forbidden`
only to a failed deletion or failed E1–E5 assertion, not to macro-selecting
source. The submitted matrix therefore claims a pre-compile rejection it does
not implement or probe.

## Required rework

1. Preserve the supported SwiftPM rejection, source modes, toolchain/SDK
   closure, line-1 classifier, native-target admission, closed plan grammar, and
   all prior-cycle fixes.
2. Choose and prove one policy-consistent branch:
   - define a conservative, total manager-owned pre-compile source admission
     rule that rejects every possible macro selection before permit, then retain
     the one `swiftc` compile command required by `manager-worker-v2`; or
   - if no such rule is defensible, do not label multi-job plan execution as
     `manager-worker-v2`. Reopen the accepted architecture through its proper
     review path and mint the required new execution-policy/claim identity, or
     reject the Swift driver for this protocol version.
3. Make a macro-selecting package fail before any compile job starts, with the
   stable `build_package_code_execution_forbidden` class and a Swift-specific
   detail.
4. Add expected-red evidence that:
   - a multi-command compile sequence cannot be admitted under
     `manager-worker-v2`;
   - macro-selecting source cannot receive a compile permit.
5. Re-run the native/degraded probes, every expected-red control, structural
   checks, and applicable spec gates, then hand off for another independent
   review.

## Independent evidence

- Probe archive SHA-256:
  `801d2efe3424e72e2441db4636c1eb62191d0777898d0e0d33983f626a2516a3`.
- Extracted archive: 19 non-test Go files, 8 test files.
- `gofmt -l .`: empty, exit 0.
- `go vet ./...`: exit 0.
- `go test -count=1 ./...`: exit 0.
- `go build`: exit 0.
- Native replay: 23/23 matched, 32 closure checks with 0 verdicts, 14/14
  expected-red controls, 56/56 structural checks, executed P2 admission true,
  green true, exit 0.
- Controls C1 through C14 replayed individually: every control exited 1 as
  required.
- Cycle-3 checks: 10 plugin-channel deletions / 0 survivors; `@Observable`
  failed with 0 load remarks; live plan 101 tokens / 0 verifier rejections; all
  16 unknown-channel vectors rejected.
- Board decision/reference SHA-256 values match the copies in the task
  worktree.
- `git diff --check` in the task worktree: exit 0.
- Spec validation reproduced the disclosed unrelated broken-link failure in
  `docs/external-build-repositories.md`; no new validation claim is made.
- No producer artifact, product code, release byte, staging state, commit, pin,
  publication, installation, or platform claim was modified by this review.

