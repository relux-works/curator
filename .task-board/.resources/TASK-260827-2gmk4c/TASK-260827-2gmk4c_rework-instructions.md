# TASK-260827-2gmk4c review verdict: changes requested

Reviewer run RUN-260827-701ce4, CR-TASK-260827-2gmk4c-1 revision 1.
Delta reviewed: `git diff 903af23ad0d0fa21328c0a2100e17968bbac6f1e fbde19647497a3dbc988462de64ca4b481263498`
(2 paths: `LOGBOOK.md`, `docs/prose-style.md`). The worktree tree matches the
candidate tree OID byte for byte (verified with `git show <oid>:docs/prose-style.md | diff -`).

## Verdict

Changes requested. Route to `to-dev`. Two of the Curator-adapted examples state
things that are false about Curator, and the task DoD item "No discrepancies
between code and description" is checked but not satisfied. Both fixes are
single-line edits inside `docs/prose-style.md`; nothing else in the delta needs
rework.

## What passes

The port itself is correct and the acceptance criteria are otherwise met.

- Structure ported in full: Voice and sentences, Paragraphs, Terminology,
  Punctuation and typography, Lists vs prose, Tone, Blacklist, Worked contrast.
- The Russian engineering prose section is dropped, and the file contains no
  Cyrillic (`grep -nP '[\x{0400}-\x{04FF}]'` returns nothing). Russian-only
  blacklist entries ("не просто X, а Y", "стоит отметить", "давайте разберёмся")
  were dropped or replaced with English equivalents ("it is important to note").
- The comparative-overview allowance (line 55) and the collapsible `<details>`
  allowance (line 57) are both present.
- Blacklist keeps all eight entries with Bad/Good pairs (lines 73 to 96).
- 106 lines, under the 200-line ceiling.
- Self-consistent typography: no em-dash, en-dash, or ellipsis characters
  anywhere in the file; no exclamation points. Guillemets appear only inside the
  Bad illustration on line 80, as in the source guide.
- Gates re-run by this reviewer on the candidate tree: `make gate-selftest`
  exit 0 (75 passed, 0 failed), `make no-broad-suppression` exit 0. These do not
  cover documentation content, so they neither confirm nor deny the findings below.

## Finding 1 (blocking): `curator repair` does not exist

`docs/prose-style.md:19` uses this as the worked example of definition-before-consequences:

> "A compiled command links a skill binary to an executable shim in `.agents/bin/`.
> `curator repair` restores broken links."

The first sentence is accurate (`README.md:71` and `internal/install/install_test.go:171`
place shims at `.agents/bin/<command>`). The second is not. Curator has no
`repair` command: the top-level dispatch in `cmd/curator/main.go:97-138` accepts
`init`, `bootstrap`, `add`, `remove`, `install`, `update`, `upgrade`, `status`,
`list`, `project`, `skill`, `global`, `hybrid`, `audit`, `gc`, `shell-init`,
`ui`, `config`, and nothing else.

The source tree says so outright. `cmd/curator/builds.go:680`:

> // Curator has no separate repair command: install and upgrade are the

And `docs/compiled-commands.md:57`, delivered by a sibling task in this same story:

> Curator provides no separate repair command: `curator install` and
> `curator upgrade` act as the reconciliation path.

So the story would land two documents that contradict each other, with the
binding style guide holding the wrong one. `TASK-260827-2gmk4c_results.md`
lists "curator repair" as a deliberate Curator adaptation, which is where the
invented command entered.

Suggested replacement that keeps the rhetorical shape and stays true:

> "A compiled command links a skill binary to an executable shim in `.agents/bin/`.
> `curator install` rebuilds a missing or drifted shim."

## Finding 2: the precise-cross-reference example points at a file Curator does not have

`docs/prose-style.md:67`:

> Cross-reference with a precise target: "see Security model in ARCHITECTURE.md",
> never "as we will see later".

`ARCHITECTURE.md` does not exist in this repository (repo root holds `README.md`,
`CONTRIBUTING.md`, `CHANGELOG.md`, `LICENSE`, `NOTICE`, and `docs/`). This is an
unadapted CocoaSkills leftover, and the task scope required examples adapted to
Curator. The rule demands a precise target while its own example names a
nonexistent one.

Suggested replacement using a section that exists in this story:

> Cross-reference with a precise target: "see Reconciliation and repair in
> docs/compiled-commands.md", never "as we will see later".

## Nit (non-blocking): the em-dash Bad example no longer shows an em-dash

`docs/prose-style.md:76-78` bans "em-dashes and en-dashes used as rhetorical
glue" and then illustrates it with a plain hyphen: "`curator` - a skill manager".
The guillemet entry two lines below does print the real glyph. Either print the
glyph here too (a quoted counter-example is arguably not prose in the guide's
own voice) or rewrite the Bad line to name the construction, for example
"Bad: an em-dash standing in for the verb, as in `curator` followed by a dash
and 'a skill manager'." Also minor: line 45 says "Do not use guillemets" while
line 79 says "Russian guillemets"; pick one spelling of the rule.

## Re-review scope

Fix findings 1 and 2 (and optionally the nit), keep the file under 200 lines,
and hand back a new CR revision. No other part of the delta, including the
`LOGBOOK.md` entry 0052, needs changes beyond dropping the `curator repair`
mention from its deliverable line.
