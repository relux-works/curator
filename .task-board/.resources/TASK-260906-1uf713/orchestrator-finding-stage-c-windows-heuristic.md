# Orchestrator finding (blocking), added to stage (c) rework 1: the dotfile heuristic does not fire on Windows

## Observed

PR #61, run 34033182525, `Test (windows-latest)` — the only failing lane, one failing case:

```
--- FAIL: TestTakeoverWarnsDotfileHeuristic (14.60s)
    ... {Adapter:opencode Home:C:\Users\RUNNER~1\AppData\Local\Temp\TestTakeoverWarnsDotfileHeuristic.../xdg/opencode OK:true ... Warnings:[]}
        {Adapter:pi       Home:C:\Users\RUNNER~1\AppData\Local\Temp\TestTakeoverWarnsDotfileHeuristic.../pi       OK:true ... Warnings:[]}
    carry no environment_foreign_manager_suspected naming chezmoi
```

Every other lane on this head is green: Test and Race on ubuntu and macOS, Lint, Naming, Interop, Gate
self-test ×3.

The cycle-1 reviewer did not see this — it recorded its verdict while the Windows lane was still
pending and said so explicitly, making no claim that every lane was green. So this is a new finding,
and it belongs to this rework rather than the next cycle.

## Why it matters beyond the one case

This is the **third consecutive stage** to ship a Windows-only defect that no Unix lane and no local
run could see. Stage (a) wrote native Windows paths into git config values, where a backslash is an
escape. Stage (b) fed POSIX literals to `filepath.IsAbs`, which needs a volume name on Windows. Now
stage (c)'s §9.5 dotfile-manager heuristic does not fire there.

The pattern is not three unlucky bugs; it is that path-shaped fixtures and path-shaped lookups are
written POSIX-first and validated only on POSIX. Treat it as such.

## Required

1. Determine whether the defect is in the **production lookup** or in the **fixture**. §9.5 describes
   the heuristic as "the presence of a well-known dotfile-manager state location (a closed, documented
   list per manager — `~/.local/share/chezmoi`, `~/.config/home-manager`, and the like)". Those
   spellings are POSIX. Decide from §9.5 whether the closed list is meant to be platform-specific —
   chezmoi on Windows keeps its state under `%LOCALAPPDATA%`, which is a different path entirely — or
   whether the list is POSIX-only by construction and the *test* is what must be platform-aware.
   Say which, with the sentence that decides it. If §9.5 does not decide it, that is a spec gap: state
   it as one rather than picking whichever makes the test pass.
2. Whichever way it resolves, the heuristic must remain **non-blocking** — §9.5 is explicit that it
   never blocks — and the case must run and assert something real on all three runners, not be skipped
   into silence. A registered skip is acceptable only if the behaviour genuinely does not exist on that
   platform, and then the class and reason must be registered and the ledger row must carry it.
3. Sweep the rest of the stage (c) delta for the same class before you hand off: a POSIX literal fed to
   a `filepath` operation, a hardcoded `/` in a path comparison, a `~/`-relative lookup assumed to
   resolve the same way, a well-known-location list that is POSIX-only. Report what you swept for and
   what you found — including "nothing else", if that is the answer.
4. The rework report must carry `gh pr checks 61` output verbatim for the head you hand off, or say
   plainly that the head was not pushed and CI therefore not consulted. The previous report's CI row
   was honest about this; keep that standard.
