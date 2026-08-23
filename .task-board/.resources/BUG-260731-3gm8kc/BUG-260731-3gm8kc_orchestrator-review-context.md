# BUG-260731-3gm8kc — orchestrator review-routing context

## Why you were spawned

The producer handed off at `to-review` with a full evidence packet
(`BUG-260731-3gm8kc_implementation-and-evidence.md`). Two earlier reviewer runs
died at launch with a transport-level `Not logged in` API error and produced no
verdict. That was an environment failure, not a review outcome — there is no
prior verdict to defer to. You are the first real review of this work.

## Isolation (mandatory)

The primary checkout `/Users/iv/Developer/ReluxWorks/curator` is on branch
`agent/link-curator-skill-registry` and is dirty with unrelated board files.
Do not switch its branch or commit in it. If you need to build or run tests,
use a read-only worktree under `.temp/BUG-260731-3gm8kc/review-worktree/`.

## What is under review

- Curator PR 9, branch `task/BUG-260731-3gm8kc-lifecycle-vector-gate`,
  head `bd6ba08acda3dc801512c408c759ac0ac6f79f26`.
  - `fee35c87` — manager-lifecycle vector gated by required case **name** instead
    of list length, plus a focused `TestManagerCompiledCacheMissDryRunVector`
    bound to the `install.BuildOutcome` vocabulary.
  - `bd6ba08` — repairs `.github/ci/toolchain-identity.sh`, which asserted
    `go env GOENV = off` while go 1.25 prints nothing for `GOENV=off`.
- curator-spec pin advanced to that commit on both PR 14
  (`release/v1.0.0-rc.6`) and PR 15 (`task/BUG-260731-2rhy74-marker-v2-fixture`).

## The decision the producer explicitly escalated to review

The producer deliberately left two checklist items unchecked and refused to tick
them to pass the handoff gate. Judge both:

1. **Item 5 — "require Curator plus PR 15 CI green."**
   Curator PR 9 CI is red in four jobs (`Lint`, `Test (ubuntu)`, `Race (ubuntu)`,
   `Test (windows)`). The producer's claim is that all four are **pre-existing on
   `main`**, proven by an isolated control branch
   `ci/goenv-control-BUG-260731-3gm8kc` (= main + only the toolchain-identity
   repair, run 30616027892) reproducing identical signatures.
   Verify that control-branch claim yourself rather than accepting it.

2. **Item 8 — "Lint clean."**
   `golangci-lint v2.12.2` exits 0 on darwin, but the repo gate is the CI `Lint`
   job, which is red on linux for two `unused` findings in `internal/godriver`
   and `internal/transaction` — neither in a file this change touches.

**Orchestrator routing fact you should use:** those pre-existing failures are
already split out as separately owned board work, and both are executing right
now in parallel with this review:
- `BUG-260731-11bpa4` (curator-windows-test-vet-compile-break) — Windows `go vet`
  `undefined: decodeHelperOutput` in `internal/runtimestore`.
- `BUG-260731-lepevi` (curator-main-ci-red-linux-lane) — the two linux `unused`
  findings plus the six `cmd/curator` compiled cases.

So "these failures are out of this bug's scope" is a testable claim, not an
excuse: either the evidence shows they are pre-existing and separately owned, or
it does not.

## Also verify

- The AC is about the **Implementations** gate: the producer claims 8/8 green on
  both PR 14 (head b07ef1d) and PR 15 (head 2629aec). Confirm against real CI.
- Whether gating by required case **name** is genuinely stronger than the length
  check it replaced. The producer supplied four negative controls, including an
  rc.3 rename that the OLD gate accepted. Re-check the teeth; a gate that only
  looks stronger is a fail.
- The producer's own flagged concern: conformance README section 4 wants a full
  immutable commit ID that passed its own required CI. The pin currently names an
  unmerged PR head whose CI is red for the pre-existing reasons above. Decide
  whether that is acceptable now or must be re-pinned after PR 9 lands.
- One out-of-scope finding the producer did not fix and did not hide:
  `internal/install TestAuthoritativeDryRunCasesMutateNothingPersistent` fails on
  an rc.6 root because the new `scope multi-project` case has no executable
  binding. Confirm it is genuinely outside this bug and, if it needs tracking,
  say so in your verdict.

## Verdict contract

Pick exactly one branch and record evidence for it:
- accepted → `done`
- changes requested → route back to `to-dev`, with a concrete, actionable list
- genuine human-only stop-the-line boundary → `blocked` with the full evidence
  packet and the exact decision needed

Do not weaken acceptance to unblock the pipeline, and do not accept on the
producer's summary alone — verify against the repository and real CI.
