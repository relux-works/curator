# TASK-260822-3nvx91 review verdict

Verdict: changes requested (`to-dev`).

## Technical review

The implementation itself is accepted in substance:

- `protocol/core.md` section 4.2.3 admits schema-8 declared module roots and defines the effective replacement set, annotation reconciliation, admitted directory form, declaration/directive bijection, containment, scan surface, vendor-copy authority, unchanged external dependency/cache identity, and the pre-`go build` failure boundary.
- The prose explicitly rejects path escape/non-portable declarations, redirects and versioned replacement forms, undeclared directives, unused declarations, nested/overlapping module roots, build-root/runtime-root overlap, and platform/Windows-folding collisions.
- Schema 8 is extended in place through `$defs.buildCommandV8`; schema 6 and 7 remain wired to the frozen `$defs.buildCommandV6`. No schema 9 was introduced. This matches the TASK-260822-1mwy10 numbering coordination.
- Generator/schema cases, validator guards, and tests cover the changed structural behavior without changing schema 1-7 generated inventories.
- The solution fits the existing spec/schema/generator architecture.

Reviewer reran these gates in the story worktree:

- `PATH=<task-venv>/bin:$PATH PYTHONDONTWRITEBYTECODE=1 make validate` — pass: 52 schemas, 686 vector files, 95 Python tests, Go tool tests.
- `git diff --check` — pass.
- `test -z "$(gofmt -l tools)"` — pass.
- `lychee --no-progress --max-retries 3 --retry-wait-time 2 --accept 200,206,429 '**/*.md'` — pass: 40 OK, 0 errors, 1 excluded.

Evidence logs are under `.temp/TASK-260822-3nvx91/` in the Curator checkout.

## Required delivery rework

No code or prose changes are requested. The acceptance criterion says the prose and schema must be committed on the story branch, but `/Users/iv/Developer/ReluxWorks/curator-spec/.temp/STORY-260822-1pm1c9/prose-worktree` remains on branch `spec/module-roots-prose` at `ebfed81` with all task files modified/untracked and no task commit.

The commit-owning mover must:

1. Review and commit exactly this task scope on `spec/module-roots-prose` without AI attribution.
2. Push the branch as already directed in the task notes.
3. Attach the resulting commit hash (and any post-commit gate evidence) to TASK-260822-3nvx91.
4. Return the task to `to-review` for a fresh reviewer cycle.

