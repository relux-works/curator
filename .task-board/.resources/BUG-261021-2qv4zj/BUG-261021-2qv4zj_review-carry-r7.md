# Review carry-forward: BUG-261021-2qv4zj r7 (Windows test fixtures), orchestrator-verified delta

The accepted review is BUG-261021-2qv4zj_review-r5.md by ivmbp_curator_rev1, at e32d101e. The orchestrator rebased it patch-identically to 04fab6de (carry r5).

The r6 delta 04fab6de..a6270186 is one signed commit that fixes the hosted gates. The orchestrator reviewed it:
- errcheck: cleanup errors are joined into the result through named returns (`publishSkillProjection`, `publishSkillFallback`); a missing staging path is nil for `RemoveAll`, so success paths are unchanged;
- G304: `#nosec` with a justification on the walker-selected canonical file read, following the existing repository precedent;
- test-only errcheck fixes;
- `.github/ci/platform-cases.tsv`: the four case-folding witnesses may skip only on Linux, under the existing host-capability class, while the directory-alias CLI cases must run everywhere;
- `gate-selftest.sh` gains controls for the wrong class and for a forbidden directory skip.

The producer reports lint 2.12.2 clean, gate self-test 274/0, and the affected packages green. Hosted CI on the landing PR is the remaining check.

The final rebase onto main after PR #158 was patch-identical (range-diff "=") at 08fa312e. r7 (df147f08) is test-only and was reviewed by the orchestrator. The source-anchor seed normalises CRLF, still requires exactly one match and restores the original endings, with added LF/CRLF unique, missing and duplicate coverage. The context CLI fixture reuses `writePackage` (`pathboundary.ProtectTree`, owner-only Windows DACLs). Production checks are unchanged.

```json
{"verdict":"accepted","head":"df147f0883027a5dc737b2d2556ec208dde78931","findings":[{"id":"F5","severity":"P2","file_line":"internal/envprofile (projection replace)","mechanism":"shared skill links briefly return ENOENT between backup and install during repair; never a mixed tree","reproduction":"see review-r5 atomic-reader companion","required_fix":"follow-up BUG-261021-4xrg5d"}],"basis":"carry-forward of rev1 r5 at e32d101e plus orchestrator-verified CI-gate delta a6270186 and test-only Windows fixture delta df147f08; patch-identical rebases"}
```
