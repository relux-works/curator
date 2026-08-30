# Review verdict — TASK-260827-xdbobc (style-sweep-and-delivery-prep)

- Run: RUN-260828-240382 (reviewer, claude-opus-5)
- Change Request: CR-TASK-260827-xdbobc-2 revision 2
- Base OID: `da28717bb99b46d431d459f90c719c937142356f` (verified ancestor of `origin/main`)
- Candidate tree OID: `c6f717f727d833525b1083fd0a41ebd6f1774a84`
- Verdict: **accepted**

## Scope reviewed

The rework of F1, F2, F3 and the non-blocking nit raised by RUN-260828-cae69c,
plus an independent re-run of every acceptance criterion against the final state
of the 12 shipped documentation files.

I bounded the rework surface first rather than trusting the outcome resource.
Splitting both change-request patches by path and comparing section by section,
rev1 and rev2 differ in exactly four paths and nowhere else:

| Path | rev1 | rev2 | Why |
| --- | --- | --- | --- |
| `.gitignore` | present | absent | F3, dogfooding artifact reverted |
| `LOGBOOK.md` | present | changed | rework round entry |
| `docs/authoring-language-adapters.md` | present | changed | F1 |
| `docs/build-https.md` | present | changed | F2 + nit |

Non-docs paths differing between the two revisions: 0. No Go source moved in
this round, so `go test ./cmd/curator/...` is not load-bearing for the rework
delta and I did not rerun the full package.

## F1 — resolved

`docs/authoring-language-adapters.md:331-333`. The parenthesis now closes after
`internal/swiftpmsource` with the period outside:

```
1. A package-local normative suite alongside the existing per-adapter suites
   (`conformance_test.go` in `internal/{npmsource,...,swiftpmbuild}`,
   `build_conformance_test.go` in `internal/rustsource`, and
   `swiftpmsource_test.go`/`swift_integration_test.go` in `internal/swiftpmsource`).
```

A paren-balance pass over the prose of all 12 files (inline code spans stripped,
link targets collapsed, fenced blocks skipped) now reports 0 imbalanced
paragraphs, down from 1.

The sentence is also factually accurate: all nine cited test files exist on
disk (`internal/{npmsource,pnpmsource,yarnclassicsource,yarnmodernsource,
swiftpminterop,swiftpmbuild}/conformance_test.go`,
`internal/rustsource/build_conformance_test.go`,
`internal/swiftpmsource/swiftpmsource_test.go`,
`internal/swiftpmsource/swift_integration_test.go`).

## F2 — resolved

`docs/build-https.md:93-98`. The original predicate is restored verbatim and the
paragraph rewrapped to the file's convention (max 78 columns, no 118-column
outlier):

```
`login` reads one token from a hidden terminal prompt. When either standard
input or standard error is not a terminal, it reads one line from standard
input instead. It stores the token through the operator's Git credential
machinery under a scope-namespaced entry, then records a `keyring` source for
that scope. A token is never a command-line argument and is never written as a
literal configuration value.
```

That matches production. `readBuildHTTPSToken` at `cmd/curator/main.go:2241`
branches on `attachedToTerminal(in) && attachedToTerminal(errOut)`, and
`attachedToTerminal` at `cmd/curator/main.go:1406` is `term.IsTerminal(fd)`. The
document again describes the same class the code implements, including
`< /dev/null` and sockets, which the rev1 "non-terminal pipe or file" wording
excluded. `docs/build-ssh.md:383` and this sentence now describe the predicate
consistently.

## F3 — resolved

`.gitignore` is byte-identical to HEAD in the candidate tree, and the dogfooding
residue it was masking is gone: `.agents/` absent,
`.claude/skills/.csk-managed.json` and `.codex/skills/.csk-managed.json` removed.
`.claude/skills/skill-go-testing-tools` and the `.codex` equivalent remain, but
those are tracked symlinks committed at HEAD, not new artifacts, and
`git check-ignore` reports them unignored.

The delivery set the orchestrator will commit is exactly 11 paths, all markdown,
and the outcome resource reproduces this block verbatim:

```
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

`git diff --name-only HEAD` plus `git ls-files --others --exclude-standard`
confirm no non-markdown file, no `.task-board` path, and nothing else. The CR
diagnostic's 518 restored board paths are the control plane behaving as designed.

## Nit — resolved

`docs/build-https.md:13` and `docs/build-ssh.md:18` now name their targets by
exact title: "Operator SSH credentials for external build repositories" and
"HTTPS credentials for external build repositories", matching line 1 of each
file. Cross-links resolve in both directions.

## Acceptance criteria, re-verified independently

- **Zero blacklist hits outside labeled negative examples.** A per-character
  non-ASCII inventory over all 12 files finds U+2014 only at
  `docs/prose-style.md:77` and `:102`, U+00AB/U+00BB only at
  `docs/prose-style.md:80`; all three sit under `Bad:` / `Bad (slop):` labels.
  Zero U+2013 anywhere. Remaining non-ASCII is `§` (60), box-drawing characters
  in directory trees (49), and `×` in a dimension expression, all legitimate.
  A fence-aware pattern sweep for antithesis, marketing, filler, summary closers,
  exclamation, and ellipsis returns 20 hits: 9 in labeled `prose-style.md`
  examples, 5 ellipses inside inline code spans (`go vet ./...`, `C:\...`,
  `CANDIDATE_REF=...`, a path glob, an error-string literal), and 3 uses of
  "rather than" for comparative precision (`docs/build-ssh.md:247`,
  `docs/source-closure-adapter-conformance.md:169,218`) which are not the
  blacklisted strawman construction. No true positives.
- **Links resolve.** A resolver over whole-file text, so wrapped links are not
  missed, checks 69 local link targets and anchor fragments across the 12 files:
  0 broken paths, 0 broken anchors.
- **Exact delivery file list in the outcome.** Present and correct, see F3.

## Definition of Done

- Blacklist-clean sweep with file:line evidence, links resolve, delivery file
  list produced: yes, all re-verified above.
- `build-ssh.md` and `build-https.md` cross-link each other: yes, one line each.
- Docs consistent with current code, no discrepancies: the rework surface was
  checked against `cmd/curator/main.go:1406,2241` (F2) and against the nine
  adapter test files on disk (F1). Both accurate.
- Result linked as a task-scoped outcome resource: `TASK-260827-xdbobc_results.md`.
- Logbook: `LOGBOOK.md` carries a rework-round entry naming F1, F2, F3 and the nit.
- Tests green: the delta is markdown-only. The only tests that assert
  documentation content are in `cmd/curator/builds_test.go`, and I ran them
  against the final tree rather than accepting the producer's claim:
  `go test ./cmd/curator/ -run
  'TestEveryCurrentnessCodeIsDocumented|TestInputCausesAreDistinctAndDocumented'
  -count=1` → `ok github.com/relux-works/curator/cmd/curator 3.242s`.

## Non-blocking observations, for a future touch only

`docs/build-https.md:13` (109 columns) and `docs/build-ssh.md:18` (106) are the
two longest lines in two files otherwise hard-wrapped near 76. The other nine
documents use one unwrapped line per paragraph, so there is no repository-wide
column convention to violate and `docs/prose-style.md` sets none.

The outcome resource's 12-item delivery list names `docs/prose-style.md`, which
is committed at HEAD and not in the working-tree delta, and omits
`docs/compiled-commands.md`, which ships in the candidate. The list is the swept
document set rather than the delta, so the sentence claiming it "match[es] the
exact working-tree delta" overstates. The verbatim `git status --short` block
immediately below it is exact, so the orchestrator has the unambiguous set
either way. Both `prose-style.md` and `compiled-commands.md` were included in my
sweep and are clean.

## Verification commands run in this review

- `git diff --stat` / `--name-status` between base `da28717b`, candidate
  `c6f717f7`, `HEAD`, and the working tree; `git merge-base --is-ancestor` for
  the base.
- Python section-split of `TASK-260827-xdbobc_change-request_rev1.patch` and
  `_rev2.patch` by `diff --git` header, compared per path.
- Per-character non-ASCII inventory over the 12 files.
- Fence-aware blacklist pattern sweep.
- Paren-balance pass over prose paragraphs.
- Link and anchor resolver over whole-file text.
- `git status --short --ignored`, `git ls-files --others --exclude-standard`,
  `git check-ignore -v`, `git diff HEAD -- .gitignore`.
- `grep` of `cmd/curator/main.go:1398-1408` and `2241-2265` for F2.
- `go test ./cmd/curator/ -run '...IsDocumented|...AreDistinctAndDocumented'`.
