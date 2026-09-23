# TASK-260908-1bfk8y — stale pin rationale in platform-exclusions.tsv (curator)

Control root: /Users/administrator/Developer/ReluxWorks/curator/curator; work only in your assigned Story
worktree `.temp/STORY-260907-2bddfc/worktree`. Read `campaign-producer-rules.md` first. Landing gate =
hosted CI (runtime runs it once at handoff).

Finding F1 of the TASK-260906-284db9 review: `.github/ci/platform-exclusions.tsv` lines ≈9–13 say
"The committed released pin is such a root: it publishes no qualification vector at all, so without a
fallback the default lane would lose an exclusion…". That claim is false for the committed pin
(`SPEC_PIN` in `.github/workflows/ci.yml`, currently `87a0d0060bad64ab883d007dcdf35df7485368bf`, whose
`conformance/v1/vectors/conformance-claim-v3-qualification.json` exists — verify with
`git -C /Users/administrator/Developer/ReluxWorks/curator/curator-spec ls-tree -r --name-only <pin> | grep qualification`)
and was already false for the previous pin.

Deliverable: correct the rationale so it states what `default_excluded_on` is actually for (a fallback for
a SUPPLIED root that predates the vector, e.g. an older `CURATOR_CONFORMANCE_ROOT`; the committed pin is
not such a root) or remove it if the fallback has no remaining consumer — decide from
`.github/ci/platform-case-gate.sh` / `excluded-packages.sh` (who reads `default_excluded_on`, and what
happens when the vector is present). If a consumer exists, keep the column and fix the prose; if none,
drop the column and the reading code together, with the gate self-test updated. Do not change any
exclusion row. Run the gate scripts' self-tests (`bash .github/ci/gate-selftest.sh` or the narrow
equivalent) with real exit codes. CHANGELOG entry only if the column or behaviour changes.

Attach `TASK-260908-1bfk8y_results.md` (before/after text, the consumer analysis with file:line, self-test
exit code) and hand off with `task-board handoff TASK-260908-1bfk8y --role developer`.
