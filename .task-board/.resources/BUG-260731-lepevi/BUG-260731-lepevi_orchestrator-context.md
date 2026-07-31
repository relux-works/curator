# BUG-260731-lepevi — orchestrator execution context

## Isolation (mandatory)

The primary checkout `/Users/iv/Developer/ReluxWorks/curator` is on branch
`agent/link-curator-skill-registry` and is dirty with unrelated board files.
Do NOT work in it and do NOT switch its branch.

Create a task-scoped worktree instead:

```bash
git -C /Users/iv/Developer/ReluxWorks/curator worktree add \
  .temp/BUG-260731-lepevi/worktree \
  -b task/BUG-260731-lepevi-linux-lane bd6ba08acda3dc801512c408c759ac0ac6f79f26
```

Initialize submodules inside the worktree before running the interop/conformance
suites.

## Base commit rationale (do not silently change it)

Base on `bd6ba08acda3dc801512c408c759ac0ac6f79f26` — the head of Curator PR 9
(`task/BUG-260731-3gm8kc-lifecycle-vector-gate`), not on `main` (`cfffd7c`).

Reason: `main` still carries the broken `.github/ci/toolchain-identity.sh` gate,
which fails every Go job at step 4 before Lint or Test findings are ever
reached. That gate is exactly what masked these two Linux failures. PR 9 repairs
it in `bd6ba08`; branching off main would hide your own evidence.

If PR 9 merges into `main` while you are working, rebase onto the merged main
commit and say so in your outcome artifact.

## Ownership boundary — a sibling agent is running concurrently

`BUG-260731-11bpa4` (Windows `go vet`) is being fixed in parallel in its own
worktree. It owns `internal/runtimestore` only.

You own the Linux lane: `internal/godriver`, `internal/transaction`, and the
`cmd/curator` compiled-status/GC expectations. Do not touch
`internal/runtimestore`, and do not rebase onto the sibling's branch. Two
separate PRs, two separate branches.

## Hard constraints from the AC

Curator CI `Lint` and `Test (ubuntu-latest)` must pass **without weakening the
unused check and without weakening the native control inventory carve-out**.

That means:
- No `//nolint`, no blanket linter exclusions, no `_ = fn` reference tricks to
  silence `unused`. Decide honestly whether
  `internal/godriver/controls_other.go:35 (*controlDomain).destroy` and
  `internal/transaction/namespace.go:310 existingNamespaceAncestor` are genuinely
  dead on Linux, and remove or properly wire them.
- The six `cmd/curator` compiled-build cases fail with
  `go-v1 build_execution_control_unavailable: the portable execution policy is
  specified for macOS and Windows only`. Do not fabricate a Linux execution
  binding and do not claim unsupported execution. Align the Linux expectation
  with the authoritative `rc5-native-control-inventory-v1` carve-out, and relate
  your decision to the open Linux qualification item named in the curator-spec
  conformance README.

## Stop-the-line

If aligning the Linux expectation would require inventing platform support that
the spec does not grant — i.e. the honest fix needs a spec/product decision
rather than code — stop before stacking workarounds, record the constraint,
options, tradeoffs, recommendation, and the exact decision needed, and set the
bug to `blocked`. Do not force-fit.

## Reporting honesty

If a required checklist item is not truly satisfied, leave it unchecked and
explain why in your notes. Do not tick items to get past the handoff gate.
