# Verdict — CR-TASK-260906-2cfxfv-2, revision 2

**ACCEPT.** Full evidence: `TASK-260906-2cfxfv_review-findings-2.md`.

`repeat-of:` cycle-1 F1 (a guard whose note claims more than its scan enforces)
recurs as F4, **minor**, on a guard that did not exist at cycle 1. F1, F1b, F2
and F3 are fixed and verified fixed.

## Why an empty `repository_delta` is the right outcome for this leaf

The Change Request snapshots the story worktree
`curator-spec/.temp/STORY-260905-1n0iy8/worktree` — the **curator-spec**
repository. This leaf's scope names a different repository:

> curator repository, branch `feat/interop-root-artifacts` in worktree
> `/Users/iv/Developer/ReluxWorks/.worktrees/curator-interop-coverage`, base
> `fb916acd`. Files: `.github/ci/root-artifacts.tsv`,
> `.github/ci/platform-cases.tsv`, `internal/interop` …

Both producer briefs direct the commits there explicitly. Nothing in the task
description, the scope or the acceptance criteria asks for any change to
curator-spec, and the reviewed change touches no curator-spec file — correctly.
So the zero-path patch is a **workspace-binding artefact**, not an absent
deliverable, and "no repository change" is right for *this* repository.

The deliverable exists and is complete in the curator repository:

- head `e3a97d7b98688d42acae7e04fa064441d222a5cc`, branch
  `feat/interop-root-artifacts`, PR https://github.com/relux-works/curator/pull/63;
- 4 commits, every one `%G? = G`, author Ivan Oparin <oparin@me.com>;
- `origin/main` (`fb916acd`) **is an ancestor** of that head — the reviewed,
  signed objects fast-forward with no rebase and no re-review;
- 11 files, +560/−64 against `origin/main`; working tree clean.

I verified that tree first-hand, not the producer's logs. Everything in
`TASK-260906-2cfxfv_review-findings-2.md` was re-derived in a throwaway
`--shared` clone; the producer's worktree and the control root were never
written to.

## What was proven

- **AC met.** Seven negative runs — each declared artefact removed from the
  candidate root in turn — all `exit=1`, package not served, missing artefact
  **named** at the plan. Positive run green. Roots materialized as plain
  checkouts and verified file-by-file against `manifest.json` (1047/1047 and
  691/691, zero extra). Never `git archive`.
- **Default lane loses nothing.** On the SPEC_PIN root `./internal/interop` went
  from 25 PASS + 6 SKIP to 25 PASS + 0 SKIP; the six that vanished are exactly
  the six that were skipping. Across both roots, 49 cases moved and **0** were
  lost.
- **The gate change is safe and load-bearing.** Both recorded lane streams
  replayed through old and new gates: `skips-observed.tsv` **byte-identical**.
  Attacked with four real shipped ledger rows: fires exactly where it must,
  silent everywhere else, narrows exactly one class. And it is the *only* thing
  standing between a served package and the one skip wording that evades the
  static scan (F1 + F1b compose).
- **Nothing is decoration.** 7 gate mutants (5 narrowing, 1 widening, 1 delete
  control) each name a specific assertion that goes red; 9 injected skip shapes
  — 8 killed, 1 the documented bound, which is fatal behaviourally in all four
  wordings tested.
- **Gates green on the reviewed head:** build, vet, gofmt, golangci-lint,
  `gate-selftest.sh` (130 passed, 0 failed), `no-broad-suppression.sh`,
  `ledger-consistency.sh` (231 rows, three platforms) — all exit 0. Hosted: no
  red lane; `Test (windows-latest)` still pending at the time of writing.

## Non-blocking, carried forward

- **F4 (minor)** — `TestNoRootReadHereEscapesTheDeclaredArtefactGuard`'s ledger
  note and `contract_test.go:228` claim `rootPath` is the only way this package
  reaches the root. A read via `os.Getenv("CURATOR_CONFORMANCE_ROOT")` bypasses
  all four contract scans (measured; the case passes while reading an undeclared
  path). Its worst outcome is a **red** mid-case failure, never a silent skip, so
  it is not the failure class this leaf closes. Fix the `so` clause, or add one
  `ast.Inspect` clause for the env var, on the next touch of that file.
- **F5 (companion, cross-repo)** — curator-spec's Implementations lane runs an
  explicit package list containing `./internal/interop` but not
  `./internal/interop/environments`, and its coverage ledger declares no interop
  rows. The current Go pin `a3abcf34` predates these cases, so nothing is lost
  today; when that pin advances past `e3a97d7b`, add the new package to
  `.github/workflows/implementations.yml` or give it ledger rows.
- **F6 (nit)** — the platform-case gate prints a `FATAL-served-package` case
  again as `tol …` in its Tier-1 section. Cosmetic; the gate still exits 1.

## Landing

PR #63 is safe to land. **Integration must land curator `e3a97d7b` / PR #63 —
not an empty curator-spec merge.** `accept_cr` routes this element to
`integrating`; the bound producer run owns that step.
