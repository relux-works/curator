# Review note for TASK-260908-1bfk8y revision 1 (orchestrator, binding) — stale pin rationale

Read the brief `1bfk8y-brief.md` (precondition) and the producer's
`TASK-260908-1bfk8y_results.md`. This is a comment-only change to
`.github/ci/platform-exclusions.tsv`: the producer KEPT the `default_excluded_on` column (claiming a
live consumer) and rewrote the rationale so it no longer says the committed pin publishes no
qualification vector.

Verify, independently:
1. **The consumer claim.** Find every reader of `default_excluded_on` (grep the `.github/ci/` scripts
   and any Go consumer) and confirm at least one live code path uses it, and that the new prose
   describes what that path actually does — in particular whether the vector really is authoritative
   on the default lane when present, and what happens when `CURATOR_CONFORMANCE_ROOT` supplies an
   older root. If no live consumer exists, keeping the column is a finding (the brief allowed
   removing it together with its reader).
2. **The factual claim.** `SPEC_PIN` at `.github/workflows/ci.yml:52` is
   `87a0d0060bad64ab883d007dcdf35df7485368bf`; confirm yourself that
   `conformance/v1/vectors/conformance-claim-v3-qualification.json` exists at that pin (via the
   curator-spec checkout at `/Users/administrator/Developer/ReluxWorks/curator/curator-spec`) and that
   the new text matches the vector's actual content for the platform rows it discusses.
3. **Data integrity.** Non-comment lines must be byte-identical to the previous revision (prove it
   yourself, e.g. `git show HEAD:<file> | grep -v '^#' | diff - <(grep -v '^#' <file>)`); no
   exclusion row may change.
4. **Gates.** `bash .github/ci/gate-selftest.sh` (or the narrow equivalent) with a real exit code; the
   hosted gate evidence for the exact candidate tree may be reused.

This is a small leaf: the finding bar is factual accuracy of the new prose and the consumer analysis,
not style. Record exactly one verdict: `accept_cr(TASK-260908-1bfk8y, revision=1,
evidence=<your outcome resource>)` on ACCEPT, or a changes-requested verdict routed with
`set_status`. Do not write into the control root's LOGBOOK.md.
