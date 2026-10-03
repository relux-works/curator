# Review verdict: accepted

BUG-261003-mjzrnv — freeze-guard-tests-break-on-release-tag. CR revision 1.
Candidate tree: 9beabb03ae1e07f17e1717902896c7ebbaca00aa.
Base/HEAD: daf15ec8e78c29148063f14905fafc0b8b682786. Fresh origin main advertisement matched this base at review preflight; no upstream overlap.
Run goal queried: not goal-bound. No code changes made by reviewer.

## Swept surfaces

| Surface | Result and evidence |
| --- | --- |
| Dynamic published inventory | Held. Independent git tag/ls-tree inventory; coverage 6/6 untagged and 7/7 tagged, including rc.6 first published under rc.7. Each published record is checked against raw tag bytes. |
| Frozen rc.13 anchor | Held. Exact rc.13 tag assertion retained; whitespace mutation rejected by real tools/validate.py with exit 1 in both states. |
| Active rc.14 lifecycle | Held. Whitespace mutation accepted untagged (exit 0), rejected tagged (exit 1 with published-record diagnostic). |
| Missing evidence | Held. Candidate exported into a temporary initialized repository containing zero tags: release-record suite exits 1 (4 failures, 2 errors), real validator exits 1 for absent baseline. The helper alone can return an empty set, but the coverage/anchor tests and production baseline check prevent suite-level empty success. |
| Narrowed guard | Held. In-memory mutation restricts production published_release_records to rc.13 only; changed byte-drift test fails on 5 older records untagged and 6 non-anchor records tagged. Negative harness exits 0 only after asserting these failures. |
| Scope/architecture | Held. Exact delta is tools/test_validate.py plus one CHANGELOG bullet. Production guard, conformance/, schemas/, release/, LOGBOOK.md unchanged. Existing production main calls the guard first. |
| Regeneration/digest | Held. make regenerate-check exits 0. Manifest SHA-256 remains 6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5. |
| Cleanup | Held. Candidate-tree diff exits 0 after all mutation tests and regeneration; temporary tag absent; no pushes or commits. git diff --check exits 0. |

Free hunt: no blocking findings after sweeping the complete two-file delta, no-tags behavior, entry-point mutants, restoration, and protected paths.

## Independently executed verification

Python 3.14.6 venv at /tmp/BUG-261003-mjzrnv-review-venv with requirements-dev.txt (jsonschema 4.25.1). All verification was rerun by reviewer; no producer pass was substituted.

| Tag state | Command | Actual exit | Result |
| --- | --- | --- | --- |
| Untagged | PYTHONPATH=tools python -m unittest tools.test_validate | 0 | 562 tests, 337.641s |
| Untagged | make validate | 0 | 672 Python tests, 399.415s; Go tests pass |
| Tagged | PYTHONPATH=tools python -m unittest tools.test_validate | 0 | 562 tests, 348.046s |
| Tagged | make validate | 0 | 672 Python tests, 396.735s; Go tests pass |
| Untagged | make regenerate-check | 0 | Published-history check, generator and diff check pass |

Tagged runs used temporary LOCAL lightweight v1.0.0-rc.14 on HEAD. Candidate tests remained the exact uncommitted CR tree; release records at HEAD and candidate are identical. Each tagged invocation removed the tag in finally; it was never pushed.

Environment observations: the initial direct unittest invocation without PYTHONPATH exited 1 (pre-existing ModuleNotFoundError: assurance); adding the tools import path resolves it, and standard Makefile discovery works without that adjustment. Initial plain git tag creation exited 128 because ambient tag signing requests a message; explicitly setting tag.gpgSign=false for the requested local lightweight test tag resolved it. Neither observation required a source change.

Raw logs, negative harness, command runner, and final scope/digest checks are attached as BUG-261003-mjzrnv_review-evidence-rev1.tar.gz. The expected negative failures in those logs are gate-defeat evidence, not failures of the unmodified candidate.

No ordinary implementation rework or external blocker found. Accept revision 1 and route to integrating via accept_cr; producer integration remains outstanding. The non-acceptance checklist branch is not applicable. LOGBOOK.md was deliberately left untouched per the binding brief; observations are preserved here and on the board.
