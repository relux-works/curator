# BUG-260731-11bpa4 — orchestrator execution context

## Isolation (mandatory)

The primary checkout `/Users/iv/Developer/ReluxWorks/curator` is on branch
`agent/link-curator-skill-registry` and is dirty with unrelated board files.
Do NOT work in it and do NOT switch its branch.

Create a task-scoped worktree instead:

```bash
git -C /Users/iv/Developer/ReluxWorks/curator worktree add \
  .temp/BUG-260731-11bpa4/worktree \
  -b task/BUG-260731-11bpa4-windows-vet bd6ba08acda3dc801512c408c759ac0ac6f79f26
```

Initialize submodules inside the worktree before running the interop/conformance
suites.

## Base commit rationale (do not silently change it)

Base on `bd6ba08acda3dc801512c408c759ac0ac6f79f26` — the head of Curator PR 9
(`task/BUG-260731-3gm8kc-lifecycle-vector-gate`), not on `main` (`cfffd7c`).

Reason: `main` still carries the broken `.github/ci/toolchain-identity.sh` gate,
which fails every Go job at step 4 before `go vet` ever runs. Branching off main
makes it impossible to demonstrate a green Windows lane for this fix. PR 9
repairs that gate in `bd6ba08`.

If PR 9 merges into `main` while you are working, rebase onto the merged main
commit and say so in your outcome artifact.

## Ownership boundary — a sibling agent is running concurrently

`BUG-260731-lepevi` (Linux lane) is being fixed in parallel in its own worktree.
It owns `internal/godriver`, `internal/transaction`, and `cmd/curator`.

You own `internal/runtimestore` only. Do not touch the sibling's files, do not
"fix" the Linux lint findings, and do not rebase onto its branch. Two separate
PRs, two separate branches.

## Scope reminder

`internal/runtimestore/targets_windows_test.go:97:14: undefined: decodeHelperOutput`.
Restore the missing helper or its call site so the Windows build compiles.
Deleting or skipping the Windows case to make the gate pass is an explicit
non-goal and will be rejected at review.

## Evidence expected

- `go vet` and `go test` green for `internal/runtimestore` on `windows-latest` in
  real Curator CI, not only locally.
- A signed PR targeting `main`.
- Windows plus non-Windows evidence attached as a task-scoped outcome resource.

## Reporting honesty

If a required checklist item is not truly satisfied, leave it unchecked and
explain why in your notes. Do not tick items to get past the handoff gate.
