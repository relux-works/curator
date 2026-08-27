# TASK-260729-365r5r — cycle-3 evidence correction

Scope of this run: **evidence-only**, under
`TASK-260729-365r5r_scope-amendment-cycle3.md`.

## 1. What the amendment authorized

- The fourth cleanup-tomb closure-parameter rename (`t *testing.T` →
  `_ *testing.T`) in `internal/transaction/namespace_pass_test.go`, plus the
  mechanically matching adaptation in the `equivcheck` copy, is **APPROVED**.
  It is the same lint-only correction as the first three; it was hidden
  initially by golangci-lint's default `max-same-issues: 3` truncation.
- Preserve the current product and test source bytes.
- Run no Go, lint, build, test, benchmark or detached command.
- Correct only the contradictory closing sentence in
  `TASK-260729-365r5r_results.md`.
- Update the board resource through `task-board`, hand off to review, do not
  mark done.

## 2. Commands run in this cycle

**None that build, test, lint, or otherwise exercise the toolchain.** The only
shell work was read-only inspection and one board write:

| command | purpose | exit |
| --- | --- | ---: |
| `task-board m 'set_status(...)'` | required first command | 0 |
| `task-board q ...` | read task/checklist state | 0 |
| `shasum -a 256 <4 files>` | verify source bytes unchanged | 0 |
| `cat gates/gate-rw2-*.exit` | read authoritative gate exits | 0 |
| `diff` / `diff -q` | verify the results.md delta and board sync | 0 / 1 (expected: 1 = files differ before the resource update) |
| `task-board resource update` | publish the corrected artifact | 0 |

No `go build`, `go test`, `go vet`, `golangci-lint`, benchmark, atomicity,
install, conformance, baseline script or detached process was started. No
process barrier check was needed because nothing was queued behind one.

## 3. Source bytes — preserved, verified

```
bb332038…  worktree/internal/transaction/namespace.go            (unchanged, == manifest-post)
3611f04f…  worktree/internal/transaction/namespace_pass_test.go  (unchanged, == manifest-post)
c86e3fbb…  equivcheck/internal/transaction/namespace_pass_test.go (unchanged, adapted copy)
997d53df…  equivcheck/internal/transaction/namespace.go          (unchanged, baseline product)
```

All four match the hashes recorded by RUN-260729-d36102. **Zero product or test
source bytes moved in this cycle.**

## 4. The correction applied

One file edited: `TASK-260729-365r5r_results.md`, closing block of §10
(`What RUN-260729-d36102 changed`). Two prose changes, no numbers altered:

1. `Exactly one file, eight lines, four parameter renames` →
   `Inside the manifested prototype worktree: exactly one file, eight lines,
   four parameter renames`.
   Reason: the sentence sat under a heading covering the whole run, but its
   supporting `shasum` block only covers the manifested worktree. Without the
   qualifier the corrected paragraph below it would have contradicted the line
   two paragraphs above.
2. The clause `no equivcheck/` — the defect cycle-2 review flagged — was
   removed and replaced with an accurate statement:
   `equivcheck/internal/transaction/namespace_pass_test.go` (`26dd7405…` →
   `c86e3fbb…`) carries **the same four closure-parameter renames**
   (lines 70/78/87/95, the identical four cases), applied under the explicit
   orchestrator directive already recorded in §8, and **`gate-rw2-equivalence`
   exited 0** (1 s, `ok … 0.741s`). `equivcheck`'s product `namespace.go`
   stayed at baseline `997d53df…`, so the gate still compares prototype
   behavior against baseline product code.

Change 1 is a scoping qualifier the amendment did not name explicitly. It is
disclosed here rather than made silently: correcting sentence 2 alone would have
left sentence 1 stating something the corrected text directly refutes. Reverting
it is a one-phrase edit if review reads the amendment more narrowly.

Nothing else in `results.md` was touched. §10's `What the reviewer should
decide` item 3 (*whether the fourth rename was in scope*) is left as written —
the amendment answers it, and this artifact plus
`TASK-260729-365r5r_scope-amendment-cycle3.md` carry that answer rather than
rewriting the handoff narrative.

## 5. Authoritative gate exits, re-read not re-run

From `.temp/TASK-260729-365r5r/gates/*.exit`, verbatim:

```
gate-rw2-gofmt.exit             0
gate-rw2-transaction.exit       0
gate-rw2-race-transaction.exit  0
gate-rw2-namespace-verbose.exit 0
gate-rw2-lint-transaction.exit  0
gate-rw2-lint-full.exit         1   (inherited godriver ineffassign only — §10)
gate-rw2-equivalence.exit       0
```

`gate-rw2-lint-full` is reported truthfully as **failing (exit 1)**. Its single
issue is the inherited `internal/godriver` `ineffassign` on a file byte-identical
to the never-edited twin; the rework constraint's acceptance bar is zero
*introduced/transaction* findings, which `gate-rw2-lint-transaction` = 0 meets.

Performance evidence (non-race atomicity **66 s**; race **84 / 76 / 75 s**
against the **480 s** bar) is unchanged and remains valid: no source byte moved
this cycle, so there is nothing that could stale it.

## 6. Status

Handed off to review. Not accepted, not done. Independent review is still
required before integration, per the AC.
