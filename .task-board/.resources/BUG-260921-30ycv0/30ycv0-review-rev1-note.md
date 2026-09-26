# Review note for BUG-260921-30ycv0 revision 1 (orchestrator, binding)

Brief 30ycv0-brief.md (rulings R1–R4): wrap/classify the `gitops.writeBlobs` spawn and wait
errors like `gitops.run` (sanitized operation context, no new path disclosure), keep the
callers' error-class mapping, no retry in product code, CHANGELOG Fixed, legacy goldens green.
Gate green: run 35545383544 — verify the gate commit resolves to the exact revision-1 tree.
Judge with your own reruns (disposable clone; bounded commands; retry once on host stalls):
1. Message shape before/after at the production entry (`install.Project` over a git source
   with the injected non-executable git — the fixture shape: PATH-resolvable 0755 script whose
   shebang interpreter is 0644) and that nothing else in `Result` changed; the mutant restoring
   the bare return fails the row.
2. Sanitization: the wrapped text discloses no more than `gitops.run` does (compare with the
   existing `git %s failed: %s` shape); no environment, no absolute temp paths beyond current
   rules; `Wait` errors and `Start` errors both wrapped; the batch-protocol error paths inside
   `writeBlobs` (short reads, missing objects) unchanged.
3. Callers: `closure`/`install` still classify the error the same way (diff the non-test
   files; only `internal/gitops/gitops.go` + CHANGELOG expected); legacy goldens green.
4. No retry; success path byte-identical (existing Extract rows); Windows skip reasons in the
   ledger vocabulary if the new rows skip there.
Record exactly one verdict: accept_cr(BUG-260921-30ycv0, revision=1, evidence=<your outcome
resource>) on ACCEPT, or changes_requested with file:line and reproduction.
