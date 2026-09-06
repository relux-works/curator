# Review brief: path overlay declarability, cycle 3

## Subject

- Branch `feat/path-overlay-declarable` at `2f2dfa4` in
  `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-path-overlay`, two signed commits past
  curator-spec main. PR https://github.com/relux-works/curator-spec/pull/47. Diff:
  `git diff origin/main..2f2dfa4`; the cycle-2 delta alone is `git diff bd39adb..2f2dfa4`.
- Read `TASK-260906-2x4s7i_review-verdict.md` (your predecessor's cycle-2 findings F5–F8), its
  `TASK-260906-2x4s7i_classification-matrix.md` and both `mutant-sweep` resources, and
  `review-brief-path-overlay-2.md`.
- Authority: `protocol/environments.md` §1, §6, §12.1; `protocol/core.md` §6.1;
  `profiles/manager.md`; `schemas/v1/manager-config-v2.schema.json`.

## What changed, and by whom

Cycle 2 returned CHANGES REQUESTED and said plainly that PR #47 was not safe to land. The repair was
made by the **orchestrator**, not by a producer, because this element's Change Request record became
unusable. Review it exactly as adversarially as a producer's work — more so, since the author is the
same party that wrote your brief.

The overlay's `allOf` was rewritten from two arms into three:

1. a URI scheme outside core §6.1's set is refused (`file:` removed from the allowlist per F6; a
   scheme is now at least two characters per F5);
2. an SCP-shaped spelling that is not a valid core §6.1 network form is refused rather than falling
   through to `path` (F7b), with a Windows drive letter carved out ahead of it;
3. a git source requires exactly one form; anything else permits none, and `directory` is refused
   there too.

Nine cases were added or inverted, and one was renamed
(`invalid-overlay-path-file-requirement-form` → `invalid-overlay-file-url-with-form`) because a
`file:` URL is no longer classified as a path.

## Review dimensions

1. **Re-run your predecessor's mutants, plus your own.** The author reports all five dying against a
   named case: the permissive host class, a ≥2-character host, dropping `directory` from the path
   refusal, re-admitting `file:`, and the scheme length `+` → `*`. Verify each independently. Then
   attack the *new* structure: mutate the Windows-drive carve-out, the SCP-shape test, the
   ordering between arms 2 and 3, and the `not/anyOf` membership. Report anything that survives.

2. **The decisions the author made, from the spec.** Two are judgement calls that must be sourced,
   not merely plausible:
   - `c:example/team-context` — a single-character host — is now a **valid git** overlay requiring a
     form, and a positive case pins it. core §6.1's host grammar is `[A-Za-z0-9][A-Za-z0-9.-]*`,
     which admits one character. But this is also exactly the spelling of a Windows drive-relative
     path. Decide whether pinning it as git is right, and say which sentence decides it. If §1 and
     §6.1 genuinely conflict here, that is a spec gap to name, not a case to publish.
   - `C:\…`, `C:/…` and `C://…` are paths because a drive letter is carved out before the SCP test.
     Verify the carve-out cannot swallow a legitimate single-character-host SCP form, and say what
     distinguishes them.

3. **Faithfulness, driven against the committed schema.** Rebuild the full matrix yourself — both
   directions, every arm — rather than reading the author's. Include the rows the author claims:
   `svn://`, `file:///x` bare and with a form, `git@my_host:x` bare and with a form, `my_host:x`,
   `C:foo` bare and with a form, all four schemes in a non-lowercase spelling, both SCP forms, a
   project-relative spelling, `.`, `..`, and a path carrying each of the five §1-forbidden members.
   Add edges the author did not: a UNC path `\\server\share\x`, a scheme-relative `//host/x`, a
   source that is only a colon, a source with a colon in a later segment (`a/b:c`).

4. **F6's consequence.** `file:` URLs are now refused outright. Confirm that is what §1 and core §6.1
   require rather than merely the opposite of the previous mistake, and that the refusal reaches the
   operator as `profile_source_kind_unsupported` in spirit — or say plainly that the schema layer
   cannot express which §1.1 diagnostic applies, if that is the case.

5. **Mechanics and gates.** Two signed commits past `origin/main`, human identity, no stray file —
   including no `tools/__pycache__`. Re-run `make validate` and `make regenerate-check` yourself.
   Confirm from the pinned manager's source that no file it consumes gained an unpassable case. Read
   `gh pr checks 47` and treat any red lane as blocking.

6. **The renamed case.** Verify the rename did not orphan an index entry or a vector name, and that
   the generator is the only thing that wrote the regenerated files.

## Constraints

Read-only outside your scratch. Never write into the control root, and do not push or modify the branch.

## Verdict contract

Attach `TASK-260906-2x4s7i_review-findings-3.md` with a `repeat-of:` line naming any earlier class that
recurs. Blocking or major → set the task to `development`. Otherwise an explicit ACCEPT at `to-review`,
and state plainly whether PR #47 is safe to land. Do not mark the task done. Then
`task-board handoff TASK-260906-2x4s7i --role reviewer`.
