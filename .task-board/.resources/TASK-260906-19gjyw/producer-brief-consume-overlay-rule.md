# Producer brief: consume the landed git-source-only overlay rule

## Where and what

- Repository `~/Developer/ReluxWorks/curator`. Branch and worktree will be named at spawn; base is
  curator main **after stage (c) lands** — the orchestrator will confirm the exact OID. First run
  `git submodule update --init --recursive`.
- Authority: curator-spec main `87a0d006` — `protocol/environments.md` §1 (both source kinds), §6,
  §12.1; `protocol/core.md` §6.1 (the canonical network identity and its host grammar);
  `schemas/v1/manager-config-v2.schema.json` `$defs/overlay`, which now **encodes** the rule and is
  the precise statement of it; `conformance/v1/schema-cases/manager-config-v2/` — forty-one overlay
  cases that pin every classification arm.

## Why this exists

The `path` overlay §6 promises could not be declared: §1 forbids a requirement form on a `path`
declaration while §12.1 and the schema demanded one on every overlay. That contradiction was
reconciled on curator-spec (PR #47): **the form is required only for a `git` source**. Stage (c) was
written against the earlier authority `550579d` and still enforces form-on-every-overlay, so the
candidate lane against curator-spec main fails on all three runners — only in `internal/config`, on
exactly the 30 subcases the reconciliation added. Against `550579d` the same families pass.

This task closes that gap. It is the last thing between stage (c) and "passing its conformance subset
against curator-spec main".

## Scope

1. **One discriminator, matching the landed schema.** `internal/envprofile/envprofile.go:1288`
   `isPathOperand` currently accepts only `/`, `./`, `../`, `.` and `..` — narrower than the schema,
   which classifies by spelling: a `://` URL whose scheme is **two or more** characters and is one of
   `ssh`/`git`/`http`/`https` (any letter case) is `git`; an SCP `[user@]host:path` whose host matches
   core §6.1's `[A-Za-z0-9][A-Za-z0-9.-]*` and whose first character after the colon is not `/`,
   whitespace or a backslash is `git`, except that a Windows drive letter (`^[A-Za-z]:[\\/]`) is
   carved out ahead of it and is a path; a `://` URL with any other scheme, and an SCP-shaped spelling
   that is not a valid §6.1 network form, are **refused outright** rather than treated as local,
   because core §6.1 says invalid network forms MUST be rejected, not treated as local; everything
   else is a path.

   Read the committed schema for the exact patterns rather than transcribing this paragraph, and read
   the forty-one cases for the decided edges — a bare drive letter `C:` is refused, `C://Users/…` and
   `c:\users\…` are paths, `c:example/x` is `git` on §6.1's one-character host, a colon in a later
   segment (`packages/team:context`) is a path, `file:` is refused as neither kind.

   Put it in **one** exported helper used by every caller. **Never probe the filesystem to decide the
   kind** — that was a stage (a) finding: a local directory could shadow a git identity and bypass the
   network allowlist.

2. **The three production surfaces.** `internal/config/environments.go` (the `forms != 1` check),
   `internal/envprofile/overlays.go` `resolveOverlay`, and `cmd/curator/compose.go`
   (`profile compose add`). After this, an operator can declare a `path` overlay from machine
   configuration and from the CLI row, and it joins the closure with its weight. A `path` overlay
   carrying `range`, `tag`, `branch`, `revision` or `directory` stays `profile_source_invalid`.

3. **`profile install <git-url|path>`.** The same helper decides that operand. Check what changes for
   install when the discriminator widens — `packages/team` and `C:\…` become path operands where they
   were not — and make sure nothing that was a `git` install silently becomes a `path` install.

4. **Retire the bound.** Stage (c) carries an explicit bound saying the `path` overlay is unreachable,
   and ledger rows 303–304 describe it. Retire both, and make the ledger rows describe what their
   tests then assert.

## Conformance

The candidate lane against curator-spec main `87a0d006` with `CI_REQUIRE_FULL_ROOT=1` must be **green
on all three runners** — that is this task's acceptance, and 30 subcases in `internal/config` are
currently red. Reproduce it locally against a materialized `87a0d006` root before handing off, and
also confirm the `SPEC_PIN` root still defers the three registered packages as designed.

## Method, from twenty findings in this stage

- **Drive every refusal through `run()`**, never through a helper. Three findings in this stage were
  gates that existed, compiled, returned the right value when a test called them directly, and were
  never reached from production.
- **Prove each with a narrowing mutant** that weakens the gate to admit exactly one member, not a
  mutant that deletes it. Deleting proves less.
- **Every blocking finding in this stage lived in an addressing mode nobody tested.** Enumerate the
  addressing modes of each surface you touch and pin each.
- **A green suite is not evidence.** The reconciliation itself needed five cycles because the corpus
  published four `source` spellings and no case could fail a wrong pattern.

## Delivery

Small signed commits, human identity, nothing written into the control root. **Do not stage with
`git add -A`** — the gate run produces artefacts; stage named paths and read `git status --short`.
Do not push and do not open a PR.

Gates, each a standalone process with its observed exit code: `go build ./...`, `go vet ./...`,
`gofmt -l cmd internal`, `golangci-lint run ./...`, `gate-selftest.sh`, `no-broad-suppression.sh`,
`ledger-consistency.sh`, `test-gate.sh` against **both** roots (materialized `SPEC_PIN`, and
curator-spec main `87a0d006` with `CI_REQUIRE_FULL_ROOT=1`), and `go test -count=1 -timeout 30m
./cmd/curator`. **Run the two lanes sequentially**, never concurrently and never alongside a `-race`
suite — that contention over the machine-wide Go test lock has cost two earlier attempts.

Attach `TASK-260906-19gjyw_drafting-report.md`: the discriminator's classification matrix driven
against the committed schema's cases; the three surfaces with each addressing mode driven through
`run()`; what changed for `profile install`; the mutant table; the retired bound and ledger rows; the
two-root gate table with the candidate lane green against `87a0d006`; and honest bounds. Then
`task-board handoff TASK-260906-19gjyw --role developer`.
