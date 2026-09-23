# TASK-260908-1bfk8y revision 1 review

Verdict: ACCEPT. Exact candidate b69de3555c6fcc84bc0b2d5522231640b4a38248 against base 09b25ef6629b41455d91dcb252ab4e4034e12750. Reviewed the producer results and independently inspected and exercised the candidate. No code modified.

## Scope and factual checks

The entire delta is the rationale comment in .github/ci/platform-exclusions.tsv:10-16. Python compared the base and candidate non-comment byte streams: identical. git diff against the candidate was empty before verification and after the independent consumer checks. git diff --check passed. No behavior changed; no CHANGELOG entry or new code tests needed.

.github/workflows/ci.yml:52 pins 87a0d0060bad64ab883d007dcdf35df7485368bf. Independently read conformance/v1/vectors/conformance-claim-v3-qualification.json from that exact curator-spec Git object: Linux excluded until TASK-260728-1skseh; macOS and Windows pending downstream native evidence. This supports the corrected rationale. Workflow :215-217 supplies the pinned root to test-gate.sh; test-gate.sh:70 calls suite-plan.sh.

## Consumer analysis

excluded-packages.sh:61 reads column three into defaults; :41 detects the vector; :65-74 uses the vector exclusively when present and the defaults otherwise. Live callers are suite-plan.sh:89 and ledger-consistency.sh:90. suite-plan.sh:113 onward writes the excluded and assertion plans; test-gate.sh:137 forwards the resolved exclusion list. platform-case-gate.sh:133 consumes that list, not column three. Repository Go search found no direct default_excluded_on or platform-exclusions.tsv consumers. Keeping the column is correct.

Independent helper checks passed 5/5: exact pinned vector on Linux and Darwin, vector-absent root on Linux and Darwin, and an adversarial vector permitting Linux despite the table's Linux fallback. The last check returned no exclusions, proving vector precedence. These are helper-level checks; committed gate-selftest.sh:560-581 additionally drives suite-plan.sh for exclusion, assertion, Darwin non-exclusion, and pre-vector fallback. No new gate behavior is introduced, and no source mutation was performed in this read-only review.

## Bounds and follow-up

The existing excluded-packages.sh:12-14 header repeats the obsolete pin claim. Producer already recorded it. Non-blocking follow-up outside the explicitly restricted TSV-prose scope; not introduced by this change. No LOGBOOK.md edit, per campaign instructions; this board artifact preserves the observation.

Ledger consistency checks declarations across platform build inventories; it does not execute foreign-platform tests. Hosted/full landing suite was not rerun or claimed independently verified here. No cross-platform runtime claim is made.

## Independent validation results

Executed in bash with pipefail, capturing the command status explicitly:
- bash .github/ci/gate-selftest.sh: exit 0; 185 passed, 0 failed.
- bash .github/ci/ledger-consistency.sh .temp/TASK-260908-1bfk8y-review/ledger: exit 0; 241 rows checked across linux/darwin/windows; OK.
- git diff --check BASE CANDIDATE: exit 0.
- Five consumer checks above: 5/5 passed, each subprocess exit 0.

Attached full independent self-test and ledger logs, plus consumer-check output. No producer test results substituted for these reruns. The run goal query returned no active goal (not goal-bound). All review checklist requirements satisfied; the rejection-routing item is not applicable because this verdict accepts the revision. Acceptance routes to integrating, not done; producer integration remains outstanding.
