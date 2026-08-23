# Reviewer verdict for TASK-260811-32iojo

Verdict: **changes requested -> to-dev**

Reviewer run: `RUN-260823-b91de2` (not goal-bound; `task-board spawn goal` reported no active goal).

## Blocking finding

The adapter does not implement a valid modern Yarn peer context. A real Yarn
4.9.2 project with three local workspaces (`root`, `host`, and `plugin`), where
`root` depends on both workspaces and `plugin` declares `host` as a required
peer, installs successfully with network disabled. Yarn generates a virtual PnP
locator for `plugin`, proving that the peer provider is resolved by Yarn's peer
context.

The same authoritative `yarn.lock`, manifests, and closed `.yarnrc.yml` fail in
`Parse` with `closure_graph_incomplete: dependency descriptor is absent`.
`buildEdges` sends every peer through ordinary descriptor lookup
(`internal/yarnmodernsource/lock.go:508-525`), while `resolveDependency`
(`lock.go:763-767`) looks for `host@*` in the selector index instead of binding
the provider selected by the consuming context. This means the current positive
N10 coverage contains only condition selection; its peer tests cover unresolved
required and unresolved optional peers, but no valid resolved peer graph.

There is a second boundary behind the parse failure: the real Yarn loader
contains a `virtual:<hash>#workspace:packages/plugin` locator and a virtual
package location below `.yarn/__virtual__`, while `reconcilePnPRuntimeState`
constructs only one base locator per `Package` and rejects any additional
observed locator (`internal/yarnmodernsource/materialize.go:850-867`). A parser-
only workaround would therefore still fail materialization.

This violates the accepted common Node contract that peer contexts distinguish
package instances and the required N10 peer-context vector. It also means a
supported pure-source modern Yarn graph cannot currently materialize or invoke
offline even though pinned Yarn 4.9.2 can resolve it without network access.

## Required rework

1. Model resolved Yarn peer contexts as distinct immutable package instances;
   do not resolve a peer declaration as an ordinary `name@range` descriptor.
   Bind the exact provider context and virtual locator identity into capture and
   selection evidence.
2. Reconcile generated PnP virtual locators, locations, and dependency maps
   against those peer-context instances instead of rejecting every virtual
   entry as an extra package.
3. Add a positive real Yarn 4.9.2 N10 regression for a supplied required peer
   that passes Parse, capture, OS-denied immutable materialization, PnP state
   reconciliation, and Node invocation. Include two provider contexts if the
   supported profile can express them, proving their identities do not alias.
4. Preserve the existing fail-closed unresolved-required-peer and explicit
   unresolved-optional-peer behavior.

## Condition rework accepted

The requested condition repair is correct and independently reproduced:

- malformed `&&`, `||`, unknown selectors, trailing input, optional,
  optional-peer, and unreachable conditions reject with
  `closure_lock_format_unsupported` before graph/config identity;
- repeated unary `!`, parentheses, left-to-right `&`, `|`, and `^`, and exact
  `os`/`cpu`/`libc` selectors match the pinned grammar probes;
- `ConditionGrammarID` participates in the configuration identity; and
- selection propagates evaluator errors instead of converting them into prune
  evidence.

Prior strict lock/rc grammar, workspace reconciliation, patch bijection, peer
metadata authority, preseeded-state rejection, and baseline real PnP fixes also
remain green.

## Validation evidence

| Command | Exit | Result |
| --- | ---: | --- |
| Real pinned Yarn 4.9.2 local-workspace peer install with `enableNetwork: false` | 0 | Yarn succeeds and emits a `virtual:` PnP locator. |
| Reviewer overlay peer probe | 0 | Confirms the adapter rejects that valid graph with `closure_graph_incomplete` before runtime reconciliation. |
| Prior condition probe | 0 | `&&` rejects with empty identity; `!!` and `!!!` select per pinned semantics. |
| Prior 26-case strict lock/rc grammar probe | 0 | 26/26 malformed variants reject without canonical aliasing. |
| Prior workspace/lock/peer-authority probes | 0 | Previously fixed cases remain rejected at the expected boundary. |
| Real baseline Yarn PnP verified-executor test under macOS `sandbox-exec` network denial | 0 | Non-peer baseline materializes and invokes successfully. |
| `go test -count=1 ./internal/yarnmodernsource` | 0 | Focused suite passes. |
| `go test -count=1 -race ./internal/yarnmodernsource` | 0 | Race gate passes. |
| `golangci-lint run ./internal/yarnmodernsource` | 0 | `0 issues.` |
| `go vet ./internal/yarnmodernsource` | 0 | Vet passes. |
| `go build ./internal/yarnmodernsource` | 0 | Focused build passes. |
| `gofmt -l internal/yarnmodernsource` and `git diff --check` | 0 | Formatting and whitespace checks pass. |
| `go test -count=1 ./...` | 0 | Full uncached repository suite passes; `cmd/curator` 408.735s and `internal/yarnmodernsource` 11.371s. |

Scratch fixtures and logs are under
`.temp/TASK-260811-32iojo-review-6/`. No product code was modified, staged,
committed, reset, or cleaned by this reviewer.
