# Review verdict — TASK-260827-xdbobc (style-sweep-and-delivery-prep)

- Run: RUN-260828-cae69c (reviewer, claude-opus-5)
- Change Request: CR-TASK-260827-xdbobc-1 revision 1
- Verdict: **changes requested** → `to-dev`

## Scope reviewed

Final state of the 12 shipped documentation files named in the delivery list,
plus the working-tree delta this worktree will hand the orchestrator. The delta
against HEAD (`41ab53cd`) is docs-only apart from `.gitignore` and `LOGBOOK.md`;
no Go source changed, so the producer's `go test ./cmd/curator/...` evidence is
not load-bearing for this delta and was not rerun.

## What holds

The mechanical acceptance criteria pass. I reran them independently rather than
accepting the producer's validator.

- **Dashes and guillemets.** A full non-ASCII inventory over the 12 files finds
  U+2014 only at `docs/prose-style.md:77` and `docs/prose-style.md:102`, and
  U+00AB only at `docs/prose-style.md:80`. All three are labeled negative
  examples (`Bad:` / `Bad (slop):`). Zero U+2013 anywhere. Remaining non-ASCII
  is `§`, `×`, and box-drawing characters in directory trees, all legitimate.
- **Antithesis, marketing, filler, summary closers.** Pattern sweep for
  `not just/merely/only`, `isn't about`, `powerful|seamless|robust|blazingly|
  game-changer|effortless|world-class|...`, `let's dive|it is important to
  note|in today's world|as we will see`, and `^In summary|In conclusion|To
  summarize` returns hits only inside `docs/prose-style.md` lines 73, 74, 88,
  89, 94, 95, 102, all labeled negative examples. Contrastive-precision
  sentences elsewhere (`binds to the directory object the pass proved, not to
  the pathname`) are not the blacklisted strawman construction and correctly
  survived.
- **Links.** All 69 local link targets and anchor fragments across the 12 files
  resolve: 0 broken paths, 0 broken anchors. The `docs/ci-gates.md` repairs
  from `.github/...` to `../.github/...` are correct for a file one level down.
- **Cross-links.** Present in both directions: `docs/build-ssh.md:18` to
  `build-https.md`, `docs/build-https.md:13` to `build-ssh.md`.
- **`docs/build-ssh.md:372`.** `authorizes a scope, not just a key` rewritten to
  `authorizes both a scope and a key` is accurate: the section opens with
  "Curator asks two questions per uncovered repository" and the transcript shows
  both a credential prompt and a scope prompt. Good fix.

## F1 (blocking) — unbalanced parenthesis introduced by the sweep

`docs/authoring-language-adapters.md:331`

The em-dash removal opened a parenthesis that never closes:

```
1. A package-local normative suite alongside the existing per-adapter suites
   (`conformance_test.go` in `internal/{npmsource,...,swiftpmbuild}`,
   `build_conformance_test.go` in `internal/rustsource`, and
   `swiftpmsource_test.go`/`swift_integration_test.go` in `internal/swiftpmsource`.
   Drive the real capture, bind, plan, materialize/build, and publish
```

The sentence-ending period at line 333 now sits inside an open paren, and the
following four sentences read as part of the aside. The original used one
em-dash as colon-glue and was correctly punctuated; the replacement broke it.
A paren-balance pass over the prose of all 12 files (inline code spans and link
targets stripped, fenced blocks skipped) reports exactly this one imbalance.

Fix: close the parenthesis after `internal/swiftpmsource` and keep the period
outside, or drop the parens and use a colon after `suites`.

## F2 (blocking) — the build-https rewrite narrows a fail-closed predicate

`docs/build-https.md:93`

Before: `When either standard input or standard error is not a terminal, it
reads one line from standard input instead.`

After: `When standard input or standard error is a non-terminal pipe or file,
it reads one line from standard input.`

The production path is `readBuildHTTPSToken` at `cmd/curator/main.go:2242`:

```go
if attachedToTerminal(in) && attachedToTerminal(errOut) {
    ... term.ReadPassword ...
} else {
    ... bufio.NewScanner(in) ...
}
```

`attachedToTerminal` is `term.IsTerminal(file.Fd())` at
`cmd/curator/main.go:1406`, and its own doc comment names the case the new
wording excludes:

```go
// The character-device test alone is not enough: `< /dev/null` is a character
// device, and treating it as a terminal would make a scripted run block on a
// prompt nobody can answer instead of failing closed.
```

`/dev/null` is neither a pipe nor a regular file, and neither is a socket; both
take the read-a-line branch. The new sentence describes a strictly narrower
class than the code implements, in the document that defines a credential
contract. The original sentence was exact.

The premise of the edit is also wrong. `is not a terminal` is a literal
predicate over a file descriptor, not an antithesis construction; the blacklist
targets `"it's not X, it's Y"` staged against a strawman. Nothing here needed
changing. `docs/build-ssh.md:383` still carries the original phrasing ("when
stdin or stderr is not a real terminal"), so the two credential documents now
describe the same predicate differently.

Fix: restore the original sentence verbatim and rewrap to the file's ~76-column
convention (line 93 is currently a 118-column outlier).

## F3 (blocking) — the delivery file list is not the delivery

The AC asks the outcome to confirm that no `.task-board` or unrelated file
reaches the orchestrator, and to list the exact docs-scope files. The outcome
lists 12 docs files. The actual working-tree delta the orchestrator will pick up
is 12 paths, and two of them are absent from that list:

```
 M .gitignore
 M LOGBOOK.md
 M README.md
 M docs/authoring-language-adapters.md
 M docs/build-https.md
 M docs/build-ssh.md
 M docs/ci-gates.md
 M docs/implementation-plan.md
 M docs/source-closure-adapter-conformance.md
?? docs/authoring-cli-commands.md
?? docs/cli.md
?? docs/troubleshooting.md
```

`LOGBOOK.md` belongs; the DoD requires the logbook entry. `.gitignore` does not
belong to a docs refresh. Its uncommitted addition is:

```
# Curator
.agents/
.claude/skills/
.codex/skills/
.cursor/rules/
.gemini/skills/
Skillfile.dev.json
```

The block exists at no commit on this branch (checked `2bb54a25`, `80d51ee3`,
`2874234a`, `41ab53cd`), and `git status --ignored` confirms it is masking real
install artifacts produced inside this worktree (`.agents/`,
`.claude/skills/.csk-managed.json`, `.codex/skills/.csk-managed.json`). It is a
side effect of a dogfooding run, not documentation.

The board control plane is clean: the CR diagnostic shows 518 `.task-board`
paths restored from base, so nothing from the worktree's board copy is in the
candidate. The gap is only the unlisted non-docs files.

Fix: reconcile the list against `git status --short` and either revert
`.gitignore` or state in the outcome why it ships with the docs commit. The
delivery list must be the exact set the orchestrator commits, `LOGBOOK.md`
included.

## Nit (non-blocking)

`docs/build-https.md:13` links to `build-ssh.md` as "SSH credentials for
external build repositories"; that file's title is "Operator SSH credentials for
external build repositories". The style guide asks for a precise cross-reference
target. Align the link text with the title while touching the file for F2.

## Verification commands run in this review

- Non-ASCII inventory over the 12 files (Python, per-character, fence-aware
  reporting): 63 lines, classified above.
- Blacklist pattern sweep (antithesis, marketing, filler, summary closers,
  exclamation, ellipsis), fenced blocks excluded: hits only in labeled
  `prose-style.md` examples.
- Paren-balance pass over prose paragraphs, inline code and link targets
  stripped: 1 imbalance (F1).
- Link and anchor resolver over all 12 files: 69 local targets, 0 broken.
- `git status --short`, `git status --short --ignored`, and `git show
  <commit>:.gitignore` across the four branch commits for F3.
- `sed`/`awk` reads of `cmd/curator/main.go:1401-1408` and `2238-2264` for F2.
