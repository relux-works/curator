# Rework brief — TASK-260910-33j1hu, revision 2 (R3+P2)

Revision 1 was rejected with one finding
(`TASK-260910-33j1hu_review-verdict-rev1.md`, F1): the validator gate requires
the seven `checkpoint_cases` names and recomputes each verdict from the case's
own inputs, but nothing pins that a name still represents its mandatory
scenario — replacing any negative case with an internally consistent passing
case of the same name survives the whole gate (0/4 mutants rejected). The
normative text, vectors and everything else passed; keep them byte-identical.

## Correction (required)
In `tools/validate.py`, pin each required scenario's discriminating input
predicates by name (configured/absent checkpoint, signature valid/invalid,
version relationship below/equal/above, equal-body vs different-body, prefix
reproduced vs not) — or, equivalently, assert complete semantic branch coverage
over the published cases in addition to the required names — so that a
self-consistent replacement scenario under a required name is refused by the
production `validate.main()` entry. Add negative tests in
`tools/test_validate.py` that replace each of the four negative scenarios with
an internally consistent passing case (keeping the name) and require the
validator to reject them; keep the existing positive controls. Reuse the
reviewer's reproduction (attached as task outcomes) as the shape of those
tests where useful.

## Validation and handoff
`make validate` and the disposable-copy regeneration proof (or the
`GIT_INDEX_FILE` form) with exit codes; evidence "Revision 2" section;
`TASK-260910-33j1hu_spec-patch_rev2.patch` = `git diff HEAD` of the worktree
(base `dced9b8`; NOT `origin/main`, which has moved) with new files via
`git add -N`; `task-board handoff TASK-260910-33j1hu --role doc-writer`.
Worktree and rules unchanged.
