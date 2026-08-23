# Reviewer verdict for TASK-260811-32iojo

Verdict: **changes requested -> to-dev**

Reviewer run: `RUN-260823-9899c8` (not goal-bound; the final
`task-board spawn goal` query reported no active goal).

## Blocking finding

The peer-virtualization traversal rejects every ordinary runtime dependency
cycle, including graphs with no peer dependencies. `buildPackageGraph` marks a
package as `expanding`, recursively expands all dependency children, and treats
any return to that package as `closure_graph_incomplete: peer virtualization
contains a recursive context` (`internal/yarnmodernsource/lock.go:553-563` and
`630-634`). It doesn't distinguish a valid non-ordering runtime SCC from a
genuinely recursive/non-convergent peer context.

This conflicts with the accepted architecture: runtime and peer cycles are
retained as canonical SCC evidence; only cycles in the execution projection
reject. It also rejects a supported modern Yarn graph before capture or
materialization, so the implementation does not yet satisfy the task's graph
closure and architecture-fit criteria.

## Reproduction

The reviewer created a three-workspace Yarn 4.9.2 project:

```text
root -> a -> b -> a
```

There are no remote packages, lifecycle scripts, plugins, patches, native
payloads, or build actions. With the task's closed `.yarnrc.yml` settings,
pinned Yarn 4.9.2 generated lock schema 8 successfully and then repeated:

```text
yarn install --immutable --immutable-cache --mode=skip-build
```

with exit 0 and `enableNetwork: false`. The generated authoritative lock has
exact workspace entries for both sides of the cycle.

Passing that exact lock, the three exact manifests, the same rc, Yarn version
4.9.2, and the bound target to `yarnmodernsource.Parse` returns:

```text
packages=0 edges=0 code=closure_graph_incomplete
error=closure_graph_incomplete: peer virtualization contains a recursive context
```

Scratch inputs and logs are under
`.temp/TASK-260811-32iojo-review-cycle/`; the exact adapter probe is under
`.temp/TASK-260811-32iojo-review-cycle-workspace/`.

## Required rework

1. Preserve ordinary runtime dependency cycles in the derived package graph.
   Use cycle-aware memoization/SCC handling so revisiting an in-progress base or
   already-derived instance does not itself imply an invalid peer context.
2. Keep fail-closed rejection for genuinely recursive, ambiguous, or
   non-convergent peer virtualization, but base it on peer-context derivation
   evidence rather than the presence of any dependency back-edge.
3. Add a regression with a real Yarn 4.9.2 cyclic workspace graph that passes
   Parse, capture, immutable private-cache replay, PnP reconciliation, and Node
   invocation. Also cover a remote/package cycle and a cycle containing a
   peer-bearing package so the valid and invalid boundaries remain explicit.
4. Assert the cycle is retained in canonical package/edge evidence and does not
   become an execution-order cycle unless a declared action graph actually
   introduces one.

## Validation evidence

| Command | Exit | Result |
| --- | ---: | --- |
| `go test -count=1 ./internal/yarnmodernsource` | 0 | Existing focused suite passes. |
| `golangci-lint run ./internal/yarnmodernsource` | 0 | `0 issues.` |
| Pinned Yarn 4.9.2 cyclic-workspace lock generation | 0 | Valid lock schema 8 generated with network disabled. |
| Pinned immutable cycle replay | 0 | Immutable lock/cache and skip-build install succeeds. |
| Exact adapter cycle probe | 0 process / rejected result | `Parse` returns `closure_graph_incomplete` before emitting graph evidence. |
| `git diff --check -- internal/yarnmodernsource README.md` | 0 | No tracked whitespace errors. |

The earlier peer-context fixes remain covered by the green focused suite, but
this new architecture-level failure is sufficient to reject the delivery; a
full repository gate cannot convert the false rejection into acceptance.

No product code was modified, staged, committed, reset, or cleaned by this
reviewer. Only task-scoped scratch probes and this verdict artifact were
created under `.temp/`.
