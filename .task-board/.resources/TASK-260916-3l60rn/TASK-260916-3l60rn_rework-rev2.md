# Rework brief — TASK-260916-3l60rn, revision 2 (E6)

Revision 1 was rejected with two corrections
(`TASK-260916-3l60rn_review-verdict-rev1.md`). Everything else passed; keep
it byte-identical.

## Corrections (both required)
- **R1 — no-rebuild is stated consistently.** §4 makes a `path` source
  directory that fails the boundary contract an entry-class
  `environment_store_untrusted` with "no rebuild", but §10.1 still says
  dry-run evaluation of an entry-class failure reports
  `would-rebuild-untrusted-store` and the §10.4 table describes that
  diagnostic as "a real operation would rebuild it". Exclude path-source
  directories from the rebuild branch explicitly in §4, §10.1 and §10.4:
  dry-run on such a failure reports `environment_store_untrusted` with no
  rebuild planned (the operator repairs the directory), never
  `would-rebuild-untrusted-store`; keep the store-entry rules and the
  operator-repair rule as settled. Add a pinned dry-run conformance case for
  the path-directory failure and a rule-7 replacement test.
- **R2 — the boundary check applies regardless of content class (pin it).**
  Both no-system cases pass all five checks and all five failure cases
  carry system modules, so a validator narrowed to
  `if case["carries_system_modules"]` survives (`validate.main()` exit 0,
  21/21 tests green). Add pinned refusal scenarios for an untrusted
  no-system overlay AND an untrusted no-system onboarding import (plus their
  trusted controls), and rule-7 replacement tests; the reviewer's narrowing
  MUST fail the production validator and the regression suite.

## Validation and handoff
`make validate` and the regeneration proof (exit codes); evidence "Revision
2" section; `TASK-260916-3l60rn_spec-patch_rev2.patch` = `git diff HEAD` of
the worktree (base `e8b53a0`) with new files via `git add -N`; EMPTY curator
delta; `task-board handoff TASK-260916-3l60rn --role doc-writer`. Worktree
and rules unchanged.
