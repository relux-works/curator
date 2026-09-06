# Review verdict — TASK-260906-1hn93j, Change Request revision 2

**Verdict: ACCEPT.** No blocking finding, no major finding. Three minors, all
one-clause editorial items or spec follow-ups; none makes a published clause
wrong.

Full findings, clause → source tables, mutant evidence and mechanics:
`TASK-260906-1hn93j_review-findings-cli-2.md`.

repeat-of: none blocking or major. Minor 1 is the residue of cycle-1 finding 1's
class (takeover shape stated inconsistently), reduced to one stale list item that
the amendment's own operative sentence overrides in the same section; the
defect cycle 1 rejected — a published operator surface documenting an
unreachable refusal branch — is gone.

## Required statement: the `repository_delta=empty` Change Request

`CR-TASK-260906-1hn93j-2` reports `repository_delta: empty`, 0 changed paths,
patch sha256 `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`
(the SHA-256 of the empty string).

**This is a base-recording artifact, not an absence of work, and I am not
accepting an empty delta.**

```
CR base OID                 e01de3f5731555457c8d3c7de6bec58b8e768f32
git rev-parse e01de3f^{tree}  f3b96cd32716842b02eb09b5c0f627010cfa3678
CR candidate tree             f3b96cd32716842b02eb09b5c0f627010cfa3678
```

The recorded base OID *is* the producer's own commit, so base tree and candidate
tree are the same object and the snapshot records "nothing past the
already-committed head" by construction. Cycle 1 shows the identical shape
(`CR-…-1`, base `f013e0c`), and six other stories on this board carry accepted
CRs with the same `repository_delta=empty`.

The leaf's repository change exists, is committed, and is what I reviewed:

```
git diff f39f4a9..e01de3f    4 files, +49/-13
  CHANGELOG.md             +16
  cli/curator.md           +23/-6
  profiles/manager.md       +6/-4
  protocol/environments.md +10/-3
```

Exactly one signed, single-parent commit past the Story base
`f39f4a9309f41a9208da817eba9129cf5a9f8dc0`, human identity
`Ivan Oparin <oparin@me.com>`, good ECDSA signature on the same key as
`f39f4a9`, no `LOGBOOK.md`, no stray file, clean tree. So this acceptance is an
ordinary acceptance of a committed four-file change whose CR snapshot was cut
one commit too late — **not** a judgement that "no repository change was the
right outcome". Had the CR genuinely carried no work, the leaf's whole purpose
(publish the takeover and import operator surface) would have been undelivered
and the verdict would have been `changes_requested`.

Worth flagging to the orchestrator as a board-mechanics issue independent of
this leaf: a CR snapshot taken after the producer commits records a zero-path
patch every time, so the CR's `repository_delta` field cannot be used to decide
whether a producer did anything.

## Why ACCEPT

1. **The §9.5 amendment is faithful and minimal.** It fixes exactly the three
   things the rework brief required — flag-carried shape with an antecedent for
   "the operation", a closed enumeration of exactly the five named onboarding
   triggers, per-file scope that never selects a scope of its own — and nothing
   more. `git diff -U0` shows two hunks across both normative documents, neither
   inside a diagnostics table; the only two diagnostic tokens in the whole diff
   pre-exist in §8.5 and §9.7. §8.3's "a second takeover of the same path" and
   the "specific unmanaged file" reading both still read true. Every one of the
   26 `takeover` hits across `protocol/`, `profiles/`, `decisions/`, `cli/` and
   `core/` is consistent with the new shape except the one named in Minor 1;
   `decisions/0010:405` ("the takeover **flag**") corroborates it.
2. **The manager sentence agrees without re-legislating.** It states the
   manager-side obligation, cites environments §9.5, and correctly leaves the
   closed enumeration in §9.5.
3. **Six rows for five operations is right.** `profile use --clear` is a
   published form of the enumerated `profile use` and §9.3 says it
   re-materializes the scope, so it writes and is covered; the `env resolve`
   row gates the flag on `--repair` exactly as the enumeration does; no row
   published the flag outside the closed set.
4. **The import row is right on both counts.** `[--use]` matches §9.1
   word-for-word and is imported by §9.6's explicit cross-reference — cycle-1
   finding 3 resolved. And §9.6's "The import writes nothing into any native
   home by itself" does settle why the row carries no `[--takeover]`; the
   activation seam it leaves is Minor 2, recorded as a spec follow-up.
5. **Surface-index discipline holds.** Every published clause maps to a
   sentence; the takeover rows' silence on scope is now itself sourced.
6. **Gates re-run by me, not read.** `make validate` exit 0 (60 schemas, 1017
   vectors, 227 tests OK, `go test` ok) and `make regenerate-check` exit 0,
   byte-clean, with `.temp/venv/bin` on `PATH`.
7. **The gates were attacked, not trusted.** Three mutants on full tree copies:
   a token-preserving *widening* of the §9.5 closed enumeration → all stages
   green; a token-preserving rename of `--takeover` across all eight
   `cli/curator.md` occurrences → all stages green; and a live control (broken
   local link in `cli/curator.md`) → `validate.py` exit 1. So the green gates
   are a measured absence of coverage over this batch's content, not evidence
   for it, and the harness is demonstrably live. Stated bound: no committed
   test asserts a row's text, a flag spelling, or the enumeration; there is no
   gate to narrow because the deliverable is normative prose and no production
   entry point exists yet.

**AC coverage: 7 of 7 rows pass**, each with its driving check named in the
findings document, plus the four widened-scope rework-brief rows.

## Minors carried forward (not rework items)

1. `protocol/environments.md:1774` and `profiles/manager.md:2285` still list
   "an explicit takeover" as a member of the *mutating profile operations*,
   which the amendment says it is not. One-clause fix; the operative sentence
   overrides it and nothing downstream is affected.
2. The closed set excludes `profile import`'s §9.1 activation half and §9.4's
   `global add`/`install` in-place materialization, both of which can meet
   unmanaged files. Escapes exist (`profile sync|use --takeover`), and widening
   the set was explicitly forbidden by the producer brief — so this is the
   orchestrator's call, recorded as a spec follow-up for STORY-260905-2z9pw4.
3. Editorial: the ~25-word takeover clause repeats verbatim in six rows and
   carries the enumeration's members but not its closure (a single note under
   the table would be strictly better); `env resolve --takeover` without
   `--repair` is behaviourally determined but syntactically unstated (correctly
   left uninvented); the takeover example sits outside the `profile use` example
   group it now illustrates.
