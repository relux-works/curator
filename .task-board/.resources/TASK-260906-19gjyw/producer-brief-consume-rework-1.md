# Producer brief: consumption rework 1

## Verdict

Review cycle 1 returned CHANGES REQUESTED — one major, one minor, one trivial. **No product code
needs to change.** Read `TASK-260906-19gjyw_review-findings-1.md`.

The reviewer verified your reasoning independently and it holds at every step: `canonicalGit` returns
a bare `host/path` verbatim when `identity.ValidCanonical`, `ensureRepo` clones it over https, and the
planted-directory attack reproduces — under a mutant that drops the widening, `Install` of
`github.com/evil-org/pkg` with a planted `./github.com/evil-org/pkg` and an allowlist that excludes it
**succeeds**, installing the planted bytes past the machine allowlist. With the widening it is refused
and clones nothing. F14 is genuinely defended, and the discriminator agreed with the committed schema
across 121 spellings and three engines with zero disagreements.

Both anomalies you reported were confirmed independently: `export-subst` really is set on
`conformance/v1/fixtures/byte-exact/subst.txt` and `git archive` really yields 65 bytes against 40 on
disk; and `go test -run '^(TestX/plain)$'` really prints "no tests to run" and exits 0.

## F1 (major) — the test named for the property checks none of its class

`TestInstallNeverDemotesANetworkIdentityToAPath` **survives** the mutant that deletes the widening,
while logging `checked 5 network-identity operands`.

The property loop filters on `identity.Parse(...)` being non-empty. `identity.Parse` returns `("", nil)`
for a bare canonical identity — no scheme, no colon, no `@` — so `github.com/evil-org/pkg`,
`github.com/relux-works/pkg` and `example.com/org/pkg` are all `continue`d. Those three are **exactly
and only** the operands the widening exists for. The five it does check are decided by
`ClassifySource` alone and are unaffected by the widening in either direction. The
`if checked == 0 { t.Fatal }` guard and the `checked %d` log are the coverage-ratio mechanism, and the
number they report is 5 for a class of size 0.

The same false claim then propagated into three artifacts: the test's own comment (which says the
mutant "fails here"), the drafting report's M10 row (3 tests where 2 fail), and — worst — a newly
registered `.github/ci/platform-cases.tsv` row asserting a property no test asserts. AC row 6 requires
the ledger rows to be truthful; that one is not.

Filter the loop on the predicate the widening actually uses, so the three bare canonical identities
are checked and the mutant kills it. Then correct the report row and the ledger wording. Apply the
mutant yourself afterwards and show the kill.

This is the epic's most persistent class — a test sitting over the hole and asserting something
adjacent — and it is worth noticing that it survived a producer, a self-review and a mutant table
this time because the mutant *was* applied and the harness *was* correct; the gap was that the
property's own filter excluded its subject. When a property test reports a count, treat the count as
a claim to check, not as reassurance.

## F2 (minor) — the "complete" install kind-change table omits a class

The reviewer enumerated the change itself by restoring stage (c)'s `isPathOperand` beside
`installOperandKind` and driving 57 operands through both. Every row of your table reproduces, and one
class is missing: `packages/team:context` and `a/b:c` were **git → refused** at `canonicalGit` before,
and are **path** now. That is refuse→accept, the one direction the table never shows, and
`installOperandCases` has no colon-in-a-later-segment operand, so nothing pins it. The risk is low —
the spelling carries no network identity and the schema does decide it is a path — but §3 asserts
completeness and the AC asserts silence.

Add the row and add `packages/team:context → identity.SourcePath` to `installOperandCases`.

## F3 (trivial) — a doc comment sits on the wrong function

Around `internal/envprofile/envprofile.go:1284`, the `installOperandKind` doc block runs without a
break into `// sourceKindRefusal …` and sits above `func sourceKindRefusal`. `installOperandKind` has
no doc comment at all, so godoc renders its entire rationale — including the widening's justification,
which is the one comment in this change a reader most needs — as documentation for the wrong function.
`golangci-lint` is clean, so nothing catches it.

## On the epic-wide mutant question — settled, no action for you

The review brief asked whether earlier mutant results in this epic could be false survivors. The
reviewer reports it as **unknown**, because the sweep artifacts record verdicts but not the literal
`-run` patterns.

The orchestrator's answer, which you do not need to act on but should understand: **the error is
one-directional.** A filter that matches nothing exits 0, which a harness reads as *survived*. It
cannot manufacture a *kill*, because a kill requires a named test to actually fail. So every "killed"
verdict in this epic remains sound, and the only possible corruption is over-reporting survivors —
which costs redundant pins, never missed coverage. That is the safe direction, and it is why this is a
note rather than a re-audit.

## Delivery

Small signed commits on `38702164` in `/Users/iv/Developer/ReluxWorks/.worktrees/curator-overlay-consume`.
Human identity. **Do not write `LOGBOOK.md`** — the logbook belongs to the orchestrator, and the commit
that wrote it on your previous run was dropped before publication. Do not push and do not touch PR #62.
**Do not stage with `git add -A`.**

Gates, each a standalone process with its exit code: `go build ./...`, `go vet ./...`,
`gofmt -l cmd internal`, `golangci-lint run ./...`, `gate-selftest.sh`, `no-broad-suppression.sh`,
`ledger-consistency.sh`, and `go test -count=1 -race` on the touched packages. The two `test-gate`
lanes only if you touch anything a conformance root sees — this rework should not. **Run them
sequentially**, and materialize any root as a plain checkout verified against `manifest.json`, never
with `git archive`.

Attach `TASK-260906-19gjyw_rework-report-1.md`: the corrected property loop with the mutant applied
and the kill shown; the added kind-change row and its pinned operand; the corrected ledger wording;
the moved doc comment; and the gate table. Then `task-board handoff TASK-260906-19gjyw --role developer`.
