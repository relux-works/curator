# Review brief: consume the landed overlay rule (cycle 1)

## Subject

- Repository `~/Developer/ReluxWorks/curator`, worktree
  `/Users/iv/Developer/ReluxWorks/.worktrees/curator-overlay-consume`, branch
  `feat/consume-overlay-rule` at `38702164`, two signed commits past `origin/main` (`7c74a492`).
  PR https://github.com/relux-works/curator/pull/62. Diff: `git diff origin/main..HEAD`.
- **All eleven hosted lanes are green on this head, and the candidate dispatch against curator-spec
  `87a0d006` with `CI_REQUIRE_FULL_ROOT=1` is green on all three runners** (run 34071813375: ubuntu
  3m38s, windows 32m17s, macos 8m7s). That closes the thirty `internal/config` subcases this task
  exists for. Confirm it from the run's own evidence artefacts rather than from this sentence.
- Read `producer-brief-consume-overlay-rule.md` and `TASK-260906-19gjyw_drafting-report.md`.
- Authority: curator-spec `87a0d006` — `protocol/environments.md` §1, §6, §12.1; `protocol/core.md`
  §6.1; `schemas/v1/manager-config-v2.schema.json` `$defs/overlay` and the forty-one overlay cases.
- Change Request revision to accept: the one recorded for TASK-260906-19gjyw.

## Two things the producer did that you should judge, not assume

**It refused to check an acceptance row.** The DoD asked for the candidate lane green on three
runners; the producer could run only darwin and reported linux and windows **UNKNOWN**, declining to
infer them from `ledger-consistency.sh` and cross-`GOOS` vet — "an argument, not a measurement". The
orchestrator supplied the measurement by pushing and dispatching, and recorded that on the board.
Judge whether the producer's portable argument was sound *and* whether the dispatch actually measures
what the row asks.

**It deliberately did not follow the landed classification in one place.** See below. That is the
single most important thing in this review.

## Review dimensions

1. **The install widening — attack this hardest.** The landed rule classifies a bare canonical
   identity (`github.com/example/x`) as a **path**, because `packages/team-context` is a declarable
   path overlay and the two spellings are syntactically identical; the published case
   `valid-overlay-path-relative` requires exactly that. The producer argues this is correct for an
   overlay `source` and wrong for a `profile install` operand, because `canonicalGit` accepts an
   already-canonical `host/path` and `ensureRepo` clones it over https — so following the
   classification there would turn a network install into a local one and hand a planted
   `./github.com/evil-org/pkg` directory the F14 machine-allowlist bypass. It therefore keeps one
   **stated, install-only widening**: a spelling the discriminator calls `path` that is also
   `identity.ValidCanonical` stays `git`.

   Verify every part of that reasoning independently. Does `canonicalGit` really accept a bare
   canonical identity? Does `ensureRepo` really clone it? Construct the planted-directory attack
   against a build *without* the widening and show what happens; then show the widening stops it.
   Then attack the widening itself: what spelling is now `git` at install that an operator would
   expect to be a local directory, and can that be abused in the other direction? Enumerate the
   install kind changes yourself rather than trusting the report's table.

2. **One discriminator, and the schema it transcribes.** `identity.ClassifySource` must agree with the
   committed `$defs/overlay` on every one of the forty-one published cases and on the edges the spec
   reviews decided: `C:\…`, `C:/…`, `C://…`, a bare `C:` (refused), `c:\users\…`, `c:example/x`
   (git, one-character §6.1 host), `packages/team:context`, `github.com:\example\x` (refused),
   `file:` (refused), `svn://` (refused), each scheme in a non-lowercase spelling. Drive the matrix
   against the Go helper and against the schema, and report any disagreement. Confirm it never
   touches the filesystem.

3. **The ECMA-262 whitespace trap.** The schema patterns are evaluated by an ECMA engine whose `\s`
   covers the vertical tab, the Unicode space separators and U+FEFF; Go's is `[\t\n\f\r ]`. The
   producer spelled the class out and says a mutant swapping Go's in is killed. Re-apply it. Then look
   for the same dialect gap elsewhere in the transcription — anchors, character classes, greediness,
   case-insensitivity — and say whether anything else differs between the two engines here.

4. **The dropped Windows-drive carve-out.** `identity.Parse` lost its `len(host) == 1` carve-out on
   the reasoning that `driveRE` returns `C:/x` and `C:\x` local before that rule is reached and the
   corpus decides `c:example/x` is git. Verify the ordering claim by construction, not by reading, and
   check what else used that carve-out.

5. **The three surfaces.** A `path` overlay must now be declarable and reach the closure with its
   weight from the config reader, `resolveOverlay` and `profile compose add`; a `path` overlay
   carrying a form or `directory` must stay `profile_source_invalid` at both gates; a source of
   neither kind must be refused before any clone or snapshot. Drive each through `run()`.

6. **The retired bound.** Stage (c)'s bound said the overlay was unreachable pending this work. Confirm
   it and the ledger rows are retired truthfully and that nothing now claims a coverage it lacks.

7. **The two anomalies the producer reported.** Both bear on evidence, not code, and both are worth
   confirming because the whole epic's method rests on them: that `git archive` corrupts a
   materialized conformance root through `.gitattributes` filters, and that `go test -run` splits its
   pattern on unbracketed slashes so an anchored `Parent/child` filter matches nothing and exits 0.
   If the second is true, ask whether any *earlier* mutant result in this epic could have been a false
   survivor for the same reason.

## Constraints

Scratch under a throwaway copy or the worktree's `.temp/`; never write into the producer's worktree or
the control root. **Run the two `test-gate.sh` lanes sequentially.** Materialize conformance roots as
plain checkouts verified against `manifest.json`, not with `git archive`.

## Verdict contract

Attach `TASK-260906-19gjyw_review-findings-1.md`. Blocking or major → set the task to `development`.
Otherwise an explicit ACCEPT at `to-review` with `accept_cr` on the recorded revision, stating what the
hosted lanes and the candidate dispatch reported. Do not mark the task done. Then
`task-board handoff TASK-260906-19gjyw --role reviewer`.
