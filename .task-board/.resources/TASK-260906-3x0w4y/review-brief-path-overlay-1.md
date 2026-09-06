# Review brief: make the section 6 `path` overlay declarable (cycle 1)

## Subject

- Story branch `task-board/story/STORY-260905-2z9pw4` at `535fda6`, one signed commit past curator-spec
  main. Diff: `git diff origin/main..535fda6`. Change Request revision to accept: the one recorded for
  TASK-260906-3x0w4y.
- Read `producer-brief-path-overlay.md` and `TASK-260906-3x0w4y_drafting-report.md`.
- Authority: `protocol/environments.md` §1 (both source kinds), §6, §12.1; `protocol/core.md` §6.1
  (the canonical git identity §1 incorporates by reference); `schemas/v1/manager-config-v2.schema.json`.

## Why this exists

Stage (c)'s review drove all three production surfaces in curator and found no operator can declare
the `path` overlay §6 promises: §1 makes a form on a `path` declaration `profile_source_invalid`,
while §12.1 and the overlay schema require a form on every overlay, and the published cases give a
path-shaped source a `revision`. The fix makes the form requirement conditional on the source kind.

## Review dimensions

1. **The discriminator — the orchestrator's own concern, test it hard.** The schema now decides
   git-versus-path by a regex over the `source` spelling. The producer justifies it from §1's
   incorporation of core §6.1 and declares bounds, among them: *`C:\foo` (drive-letter) and `a:b/c`
   classify as git.*

   Treat that first bound as a probable **defect, not a bound**. Windows is a supported platform for
   this manager — `ci.yml` runs a `windows-latest` lane and stage (b) shipped a Windows-only defect
   this epic had to fix twice. §1 says a `path` source is "an operator-local package directory named
   by an absolute path, or by a project-relative path" and does not say POSIX-only. So under the
   committed pattern a Windows operator's `C:\Users\operator\context` overlay is classified git,
   required to carry a form, and then refused by §1 for carrying one — reproducing on Windows exactly
   the undeclarability this task exists to remove.

   Decide, with the sentences that decide it: is a Windows absolute path a legal `path` source under
   §1? If it is, the pattern must classify `<letter>:\…` and `<letter>:/…` as a path before the SCP
   arm, and the missing cases are a finding. If §1 genuinely restricts `path` to POSIX spellings, say
   which sentence does that, and the bound stands. Do not accept "the spec's examples are POSIX" as
   an answer — examples are not a restriction.

2. **The discriminator's other edges.** Verify the producer's classification table yourself against
   core §6.1 rather than re-reading its probe output: the four schemes in any case, the SCP
   `[user@]host:path` form, `file:///x`, a bare relative segment, `.`, `..`. Ask what a `source` that
   is neither — an empty string, a URL with an unknown scheme, a path containing a colon — does, and
   whether the resulting classification is safe in both directions. A misclassification toward `git`
   makes a real path overlay undeclarable; a misclassification toward `path` lets a git overlay skip
   its form requirement. Say which direction each edge fails in.

3. **Faithfulness to §1 and §6.** The amended §12.1 knob row and the schema must together permit
   exactly what §1 and §6 permit and refuse exactly what they refuse. Drive the matrix against the
   committed schema yourself with a JSON Schema validator: path bare, path plus each of `range`,
   `tag`, `revision`, `directory`, `branch`, path plus `weight`, git bare, git plus a form. Confirm
   §1's five forbidden members are all refused on a `path` source and that `weight` stays legal on
   both kinds. Confirm §1 and §6 prose is untouched.

4. **The two declared survivors.** The report declares two surviving mutants: a shape-1 discriminator
   with the arms swapped, and reverting the knob-row prose. Judge whether each bound is honest, and in
   particular whether the second one — "the Values-cell wording is review-enforced" — is acceptable
   for a normative table, or whether a case should pin it.

5. **The consumed-file constraint.** `.github/workflows/implementations.yml` runs the *pinned* Go
   manager's interop suite against this PR's conformance root. Verify no file that pin consumes gained
   a case it cannot pass, by checking which files it reads rather than by trusting the report. This
   repository has already lost a lane to exactly that (PR #41).

6. **Mechanics.** Exactly one signed commit past current main, human identity, no `LOGBOOK.md` and no
   stray file — note `make validate` generates `tools/__pycache__`, now ignored, and a recent batch
   did stage one by accident. `make validate` and `make regenerate-check` re-run by you, not trusted
   from the report.

## Constraints

Read-only outside your scratch (`.temp/` inside the workspace). Never write into the control root.

## Verdict contract

Attach `TASK-260906-3x0w4y_review-findings-1.md`. Blocking or major → set the task to `development`.
Otherwise an explicit ACCEPT at `to-review` with `accept_cr` on the recorded revision. Do not mark the
task done. Then `task-board handoff TASK-260906-3x0w4y --role reviewer`.
