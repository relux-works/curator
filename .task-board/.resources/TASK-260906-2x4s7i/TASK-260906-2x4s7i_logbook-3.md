# Logbook — cycle-3 review of PR #47 (curator-spec, path overlay declarability)

## `git archive` is not a safe way to snapshot this repository

My first mutant control run failed with `vector digest mismatch for fixtures/byte-exact/subst.txt`
against an *unmutated* schema. Cause: I built the scratch tree with `git archive 2f2dfa4 | tar -x`,
and `git archive` applies `export-subst` and the `.gitattributes` eol rules. `conformance/v1/fixtures/
byte-exact/` exists precisely to pin bytes that those filters would change — its own nested
`.gitattributes` outranks the root file, and the blobs were committed through
`git hash-object -w --no-filters`. Any tooling that snapshots this repo for validation must copy the
working tree (`rsync -a --exclude .git`) or check out, never `git archive`. Worth knowing before the
next reviewer loses the same twenty minutes.

## A build artifact slipped past nine green CI lanes

`2f2dfa4` committed `generate-vectors` — 5,462,002 bytes of Mach-O arm64, the default output name of
`go build ./tools/generate-vectors`. All nine PR checks pass with it present, because no lane inspects
the tracked file set. This is the second instance of the same class on this branch: cycle 1 found the
`tools/__pycache__` trap live, and `550579d` closed it for Python only by adding `__pycache__/` and
`*.py[cod]` to `.gitignore`. The Makefile uses `go run`, never `go build`, so nothing in the repository
anticipates a Go binary at the root. A `/generate-vectors` ignore line (root-anchored, so
`tools/generate-vectors/` stays tracked) closes it.

## Two of six mutant survivors were provable no-ops, and saying so mattered

`M-drive-del3` and `M-reorder` survive the corpus. Neither is an evidence gap:

* The Windows drive carve-out `^[A-Za-z]:[\\/]` and the SCP pattern's first path character `[^/\s\\]`
  have **empty intersection** — the carve-out demands a `/` or `\` right after the colon and the SCP
  class excludes exactly those two. So arm 3's copy of the carve-out cannot change a result, and its
  survival is redundancy, not blindness. Arm 2's copy *is* load-bearing.
* JSON Schema `allOf` is order-independent and each arm's preconditions are explicit `not` clauses,
  so permuting the arms is semantically inert.

Reporting an exit code alone would have made both look like gaps. The distinguishing evidence is a
**classification diff**: run every probed source under committed and mutated schemas and print only the
rows that move. Empty diff ⇒ no-op; non-empty diff ⇒ a named behaviour nothing pins. That technique
turned four of the remaining survivors from "green suite" into concrete claims
(`a/b:c` path→refused, `host:\x` refused→git, and so on) and told me exactly which case would kill each.

## The one-letter-scheme / drive-letter collision is unresolvable, and both cycles picked the same side

`C://Users/x` and `g://host/x` are the same string shape. RFC 3986 permits a one-character URI scheme;
Windows drive letters are one character. Cycle 2 required `C://…` to be a declarable path (it was
refused in all 8 columns, a regression against main), which forces `g://host/x` to be a path too. That
is safe only because none of core §6.1's four schemes is one character, so nothing legal is lost — but
it is a rule that exists nowhere in the prose and should be written down rather than rediscovered.
