# THE ONLY CURRENT INSTRUCTION — BUG-261001-2n70px: curator install skips project skills at a non-git root (curator)

## Problem (evidence: TASK-261001-3qugz9 results, probe P1, citation C6)
A product folder that is NOT a git repository but holds nested repos (e.g. ~/src/op2-product) is a supported project root:
- `curator project add` / `project resolve` work there, and the schema-2 lock is written at the root.

But `curator install` (also with `--fix-gitignore`):
- prints "skipped";
- exits 0;
- installs NO project skills.

Cause: the gitignore hygiene gate (internal/gitignore/gitignore.go:48–75, 96–101, 116–138; internal/install/install.go:334; cmd/curator/main.go:794–805) runs `git check-ignore`, gets exit 128 (not a repository), and skips the install. Reporting "skipped" with exit 0 hides the failure.

## Required
1. A non-git project root materialises declared project skills through the real `curator install` entry point.
   - No `git init`; no files written outside the declared project surfaces.
   - The gitignore hygiene check applies only when the root is inside a git work tree.
   - When it is not, record that the check does not apply (a clear notice line), not a silent skip.
2. Keep git-workspace safety exactly as today for real repos:
   - an ignored-path conflict still refuses or fixes as before;
   - a git root with no .gitignore entry behaves as before.

   Distinguish "not a repository" (exit 128 with the not-a-git-repository message) from other git failures. Other failures must still fail closed: never treat an arbitrary git error as "not a repo".
3. Entry-point tests on a temp non-git root with a throwaway HOME:
   - **positive:** skill bytes present after install, exit 0, notice printed;
   - **git-repo regression rows:** unchanged behaviour;
   - **broken git:** e.g. a corrupt .git file or an unexpected git exit → refuses, non-zero.

   Mutant: restore the old skip and show the positive row fails, with real exit codes.
4. Docs: one paragraph in docs/cli.md (project install), covering non-git product folders and where the lock and bindings live. One CHANGELOG line under Unreleased.

Do NOT touch LOGBOOK.md. Never spell any employer name. Run targeted `go test` for cmd/curator, internal/install and internal/gitignore with real exit codes.

## Handoff
Update the results, then run `task-board handoff BUG-261001-2n70px --role developer`, then END YOUR TURN.
