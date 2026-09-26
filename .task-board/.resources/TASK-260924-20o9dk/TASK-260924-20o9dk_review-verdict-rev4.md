# TASK-260924-20o9dk — review verdict rev4: ACCEPTED

Reviewer: claude-opus-5-5 (low). Candidate tree b5db0d85 on base 9f0da708. I verified it in a disposable `git archive` extraction under $TMPDIR. Shell: zsh with pipefail.

## Checks
1. **Corpus.** The vendored corpus is byte-identical to curator-spec 5746367 (`diff -r` on conformance/ and schemas/skillfile-sources-v1 → IDENT). The pin file reads 574636785c9d…. The only testdata directory left is skillfile-sources-v1; no draft-sources-v1 vendoring remains.
2. **Harness tests.** Pin, CorpusCounts, SchemaCases, SnapshotVectors, SemanticCoverage and ReplayRejectsLockedObjectFormatMismatch pass (exit 0). The registry, config and snapshot unit packages pass (exit 0). Batch4 alone passes in 102 s, so the split keeps each call under the ~5-minute limit.
3. **Production changes, read.**
   - project_resolve.go: an existing checkout is fetched with `gitops.FetchURLIsolated`, per resolved target and with fallback. The stored origin is not used.
   - sourcepolicy.go: an SCP-like endpoint combined with an alias port is refused with `repository_policy_invalid`, both at parse time and at plan time. `ConnectionURL` also refuses to render it, as a second layer.
   - registry.go: a revocation is admitted on a repository+commit match, even when the grant is exact-only.
   - draftsources.go `verifyLockedGitMember`: compares `rev-parse --show-object-format` with `member.Package.Commit.ObjectFormat` (fixes C1).
   - **C2:** `driveV2DeclaredMirror` requires `operation == source-resolution` plus `declared_tag`, and compares `got` with `c.Expected`.

## Mutants (my own runs, each file restored with cmp afterwards)
- **M1** — object-format check skipped (`if false &&`): TestDraftSourcesReplayRejectsLockedObjectFormatMismatch FAILS (exit 1). KILLED.
- **M2** — advisory ignores repository+commit revocation (`if !match(record)`): the production-entry row `Batch3/attestation-evidence-revoked-identity-commit-advisory` FAILS with "Resolve = unknown, want revoked". TestResolveExact also FAILS. KILLED.
  - My first attempt at this mutant was malformed: `&& false` widens admission instead of narrowing it, and the row passed. I discarded that attempt.
- **M3** — SCP alias port rendered (parse refusal removed + `ConnectionURL` renders the port): `Batch4/v2-scp-alias-port-refused` FAILS with "want repository_policy_invalid". KILLED.
  - Bound: removing only the parse refusal is killed by the config unit tests, not by the corpus row, because `ConnectionURL` still fails closed before I/O (the second layer).
- **M4** — refresh fetches the stored origin (`FetchIsolated`): `Batch3/v2-refresh-current-endpoint-existing-checkout` FAILS with "project refresh = 1". KILLED.

## Not rerun by me
- The full set of 5 semantic batches. I rely on the orchestrator note that the hosted gate is green on the three OSes.
- The literal-isolation gate-fix tests and the gitshim seam: I read them in the diff only.

## Residuals (non-blocking)
- M3 above: the parse-time SCP refusal is killed only at unit level.
