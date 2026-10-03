# THE ONLY CURRENT INSTRUCTION — BUG-261003-mjzrnv: release-history freeze tests must hold once a release is tagged (curator-spec)

## Evidence
The tag-triggered release workflow for v1.0.0-rc.14 (run 37143170748, job "Publish signed specification artifacts", step "Validate release input") failed. The tag was on main daf15ec8, the release-prep squash.

    FAIL: test_all_six_published_records_are_covered_and_unchanged (AssertionError: Items in the first set but not the second: …)
    FAIL: test_rc14_candidate_keeps_rc13_as_published_history_anchor ('release/1.0.0-rc.14.json' unexpectedly found in {… published records …})

Both tests passed in PR CI only because no rc.14 tag existed yet. Once `v1.0.0-rc.14` exists, the guard rightly treats rc.14 as published, and the tests' hard-coded "six published, rc.14 is a candidate" view breaks. The orchestrator has DELETED that tag; no release had been published. It will re-tag rc.14 on the commit that lands this fix.

## Do
1. Make these tests correct in BOTH states, before and after the active version's tag exists. Derive the expected published set from the actual release tags, or from the release records plus tags, the same way the guard does, instead of a fixed count. Keep the real protections:
   - every PUBLISHED record is byte-identical to its tag;
   - the rc.13 record is the frozen anchor;
   - the active-version record may change only while it is untagged.
   Do not weaken the guard itself.
2. Prove both states locally:
   - (a) the current repo with no rc.14 tag: tests pass;
   - (b) a temporary LOCAL lightweight tag `v1.0.0-rc.14` on HEAD: `python -m unittest tools.test_validate` and `make validate` pass. Then delete the local test tag; never push it.
   - Negative: mutate release/1.0.0-rc.13.json, and the guard still fails.
   Record real exit codes. Python jsonschema is needed: use a venv.
3. Do NOT change conformance/, schemas/ or release/*.json. The manifest digest must stay 6f832d81…; verify that `make regenerate-check` passes and the digest is unchanged. One CHANGELOG line in the rc.14 section, or Unreleased per the repo's convention. Never edit LOGBOOK.md.

Then run `task-board handoff BUG-261003-mjzrnv --role developer` and END YOUR TURN.
