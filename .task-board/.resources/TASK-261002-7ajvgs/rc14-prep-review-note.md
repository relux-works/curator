# Review note — TASK-261002-7ajvgs rc.14 prep (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Same-provider review (operator rule, R138); release-critical. curator-spec CR, base 045ceb20, 37 paths, gate green. The candidate is also pushed as branch `candidate-rc14`. Review it against `rc14-prep-brief.md` and RELEASE.md.

Verify, with real exit codes:
1. **Active version.** 1.0.0-rc.14 everywhere it should be (README, tools, generator). Historical rc.13/rc.9 identifiers are byte-identical. `release/1.0.0-rc.13.json` equals the v1.0.0-rc.13 tag: the #122 guard passes.
2. **rc.14 record.**
   - `release/1.0.0-rc.14.json` pins EXACTLY the generated core manifest sha256 6f832d81…; recompute it yourself.
   - The source suite digest 061ec05d… is unchanged.
   - Unsupported sets are truthful and empty; there are no hardened-execution claims.
3. **Vectors.** Many vector files changed. Confirm the changes are version-string and metadata only, with no semantic change to any case. Diff a sample of 3 files and explain any non-metadata hunk.
4. **Checks.** `make validate`, `make regenerate-check` and the tools tests pass. The Makefile and release.yml diff lists include rc.14. CHANGELOG has an accurate dated rc.14 section covering #99, #113, #115, #116, #118, #120, #121 and #122; B3 and 3ny11n are not included.
5. **Hygiene.** No LOGBOOK; never spell any employer name.

accept_cr, or changes requested with file:line.
