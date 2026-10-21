# Review carry-forward: BUG-261021-2qv4zj r6 (hosted CI gates), orchestrator-verified delta

The accepted review is BUG-261021-2qv4zj_review-r5.md by ivmbp_curator_rev1, at e32d101e. The orchestrator rebased it patch-identically to 04fab6de (carry r5).

The r6 delta 04fab6de..a6270186 is one signed commit that fixes the hosted gates. The orchestrator reviewed it:
- errcheck: cleanup errors are joined into the result through named returns (`publishSkillProjection`, `publishSkillFallback`); a missing staging path is nil for `RemoveAll`, so success paths are unchanged;
- G304: `#nosec` with a justification on the walker-selected canonical file read, following the existing repository precedent;
- test-only errcheck fixes;
- `.github/ci/platform-cases.tsv`: the four case-folding witnesses may skip only on Linux, under the existing host-capability class, while the directory-alias CLI cases must run everywhere;
- `gate-selftest.sh` gains controls for the wrong class and for a forbidden directory skip.

The producer reports lint 2.12.2 clean, gate self-test 274/0, and the affected packages green. Hosted CI on the landing PR is the remaining check.

The final rebase onto main after PR #158 is again patch-identical (range-diff "=").

```json
{"verdict":"accepted","head":"08fa312e49437893b05ad87a888983601b1899c9","findings":[{"id":"F5","severity":"P2","file_line":"internal/envprofile (projection replace)","mechanism":"shared skill links briefly return ENOENT between backup and install during repair; never a mixed tree","reproduction":"see review-r5 atomic-reader companion","required_fix":"follow-up BUG-261021-4xrg5d"}],"basis":"carry-forward of rev1 r5 at e32d101e plus orchestrator-verified CI-gate delta a6270186; patch-identical rebases"}
```
