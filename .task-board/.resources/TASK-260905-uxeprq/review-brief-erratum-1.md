# Review brief: Decision 0010 erratum

## Subject
- Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-0010-erratum`,
  branch `draft/decision-0010-erratum`, head `9198c64`, base `b4f29cd` (= main).
- Read the producer brief `producer-brief-0010-erratum.md` (precondition on
  TASK-260905-uxeprq) and the drafting report (outcome resource). Diff:
  `git diff b4f29cd..9198c64 -- decisions/0010-agent-environment-profiles.md`.

## Review dimensions
1. **Amendment, not silent edit**: an Erratum section after Status; each of the three
   items quotes the original verbatim (diff the quote against the original text),
   states why it is wrong, cites evidence, and gives the corrected statement; the
   original passages remain in place with `[Erratum 2026-09-05, item N]` markers; the
   phasing-table row is annotated without editing cells; Status mentions the erratum;
   `git diff --stat` shows exactly one file and no unrelated hunks.
2. **Facts**: re-verify on the installed binaries — `pi --version` and `pi --help`
   (no `--system-prompt-file`/`--append-system-prompt-file`; `--system-prompt <text>`,
   `--append-system-prompt <text>` present), `claude --version`/`--help` (both `-file`
   spellings present). Verify the `path` promotion claim against environments.md §1 at
   `b4f29cd` and PR #34 (`git log --oneline f8d7e7ab -1`). Verify the credentials
   corrections quote review M4 accurately (board resource
   `pre-implementation-review-v3.md` on STORY-260901-zddtn8) and that the
   `oauth.claude.profile.*` item is labeled unverified.
3. **No overreach**: the erratum corrects claims; it does not pre-empt the environments
   revision 1.1 normative rewrite (M4/M5) beyond naming it as the owner.
4. **Signed commit**: one commit, signature verifies with the human identity.

## Constraints
Read-only: no edits, commits, pushes. Never write LOGBOOK.md or anything into the
control root.

## Verdict contract
Attach `TASK-260905-uxeprq_review-findings-erratum-1.md` (outcome resource): per finding —
severity (blocking|major|minor|nit), line, quote, what is wrong, fix. Blocking or major →
route to `development`; otherwise explicit ACCEPT, leave at `to-review`. Do not mark done.
Then `task-board handoff TASK-260905-uxeprq --role reviewer`.
