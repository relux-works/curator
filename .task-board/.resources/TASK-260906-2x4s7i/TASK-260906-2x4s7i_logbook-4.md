# Logbook — cycle-4 review of PR #47 (`18dca85`)

## A fix applied to two of three copies

The `file:`-is-accepted error existed in three places: the schema allowlist, the PR body, and
`CHANGELOG.md`. Cycle 2 (F6) fixed the schema. Cycle 3 (F11) fixed the PR body. Nobody looked at the
changelog, because `git log origin/main..HEAD -- CHANGELOG.md` returns only `bd39adb` and a
reviewer diffing the *rework delta* never sees the file at all. **A prose claim that was true when
written and false after a rework does not appear in any subsequent delta.** When a review finds a
factual error in one artefact, the next cycle has to grep for the same claim across the whole branch,
not re-read the diff. `grep -rn 'ssh.*git.*http' --include='*.md'` would have found it in a second.

## "Provably dead" is a claim about the committed pattern only

F13's removal of arm 3's Windows-drive carve-out is genuinely a no-op — proven analytically and
measured at 0 differences over 16,159 spellings. But the clause was still load-bearing *for a mutant*:
`M-drive-wide` used to fail `valid-overlay-git-single-letter-host.json` because the mutation hit both
copies of the clause, and with one copy gone the same mutation is arm-2-local and nothing notices.
**Removing redundant-looking structure can lower mutation coverage without changing behaviour**, and
neither the corpus nor the gates can tell you. Re-running the mutant sweep across a simplification is
the only way to see it.

## A negative case does not pin the reason a thing is invalid

`invalid-overlay-scp-backslash-path.json` (`github.com:\example\x`, no form) kills the arm-2-local
backslash mutant but not the both-sites one: under that mutant the source becomes a `git` source with
no form, so it is still invalid — for a different reason. What actually kills the both-sites mutant is
`valid-overlay-path-windows-backslash.json`, a *positive* case. **When a negative case is the evidence
for a classification decision, check which mutation it distinguishes;** an invalid-for-another-reason
case proves nothing about the arm you meant to pin.

## Reading the pin's source is worth it even when the lanes are green

`cocoaskills`' `tests/protocol_conformance_adapters.py` walks `schema-cases/index.json` generically and
validates every entry against the *implementation's own* schema copy — a consumer that would break on
any new case. It looks alarming until you read `implementations.yml`, which carries an explicit note
that the module is not run here because it authenticates one immutable rc.6 suite. **The consumer
existing is not the consumer running.** Grep found the risk; only the workflow's own prose resolved it.

## The gate that does not exist

Nine green checks shipped a 5.4 MB Mach-O binary in cycle 3. `.gitignore` now names two artefact
classes, each added after an accident, and nothing asserts the invariant. Worth knowing: CI can never
produce the artefact — every lane uses `go run`, never `go build` — so a "working tree is clean" check
would be permanently green. The gate has to read `git ls-files` and reject binary or mode-755 tracked
paths outside a small conformance allowlist.
