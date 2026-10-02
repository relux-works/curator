# Review — restore-rc13-record-and-freeze-history

Candidate: CR-TASK-261002-1ig5ev-1 revision 1, tree e2e2da29e55a9f431d4083a47b384215be73b0e2; base e41c561b3300a0a4d3425437fddc5a7048d7b11e.

## Swept surfaces

| Surface | Review evidence |
| --- | --- |
| Published history | Exact candidate/tag diff for rc.13 exits 0. Inventory contains 6/6 existing published records. rc.6 has no own tag; the first release-tag snapshot is rc.7. Missing/unreadable inventory fails closed. |
| Generator | Production main now calls writeCandidateMetadata; no release record writer remains. Current core/source digests live in conformance/candidate.json, a minimal separate candidate mechanism. |
| Freeze guard | Called first by validate.main and by make regenerate-check. Independent CLI negative in an isolated temporary fixture: baseline exits 0; appending whitespace separately to rc.5, rc.6, rc.7, rc.8, rc.9, rc.13 exits 1 for 6/6 and identifies the mutated record. |
| Consumers | implementations.yml and all conformance/v1 bytes unchanged. Inspected curator's exact pinned e4a6a8d5 revision: selectedSuiteIdentity hashes CURATOR_CONFORMANCE_ROOT/manifest.json, not release metadata; bd03456b... remains the Muse candidate digest, with 103 count rows in its ledger. Hosted cross-platform consumer jobs were not rerun; compatibility conclusion is scoped to unchanged inputs and inspected dispatch. |
| Tests | Added coverage includes all six byte mutations, deletion, unreadable/malformed inventory, missing latest evidence, candidate identity/digests, and real generator preservation including a future sentinel and missing rc.13. |
| Hygiene | 13 intended paths; one Unreleased changelog line; no LOGBOOK, released-schema delta, or rc.14 record. git diff --check exits 0; gofmt lists no files. |
| Freshness | Fresh origin main advertisement and fetched OID both equal CR base e41c561b. No overlapping upstream delta. |

Initial system-Python make validate exited 2 because jsonschema was absent; an old temporary environment also lacked it. Created /tmp/1ig5ev-review-env and installed requirements-dev.txt successfully. Initial isolated negative fixture omitted assurance.py and failed during import; the corrected fixture copies all tools, passes the baseline, and produces the guard failures reported above.

Guard bound: release tags available locally are the authority; latest rc.13 evidence is mandatory. No assertion about remote tags absent from the checkout. Both CI/release workflows fetch full history. Released schema bytes remain unchanged.

## Final validation and verdict

**Accepted, revision 1. No blocking findings.**

- Reviewer reran `make validate` in the requirements-installed environment: exit 0; 73 schemas / 1,294 vector files, 669 Python tests OK, Go tools tests OK. Wall time included the host-wide syspolicy execution stall; pending commands were retained until recovery (crash counter 351 → 353), with no background-and-exit handoff. The completed Python run reported 1203.791 seconds.
- Reviewer reran `make regenerate-check`: exit 0. A wrapper measured unchanged raw bytes AND mtimes for 6/6 release files across the real generator invocation.
- Independent narrowed mutant in a temporary copy restricted the guard to rc.13. A whitespace mutation of rc.5 then returned CLI exit 0, violating the negative harness requirement of exit 1. Thus the all-record negative distinguishes a narrowed guard from complete coverage.
- Final tracked-file comparison to the candidate tree excluding the new untracked candidate file: exit 0. The new candidate file was compared directly to its candidate-tree blob and matches. The initial broad git-diff wrapper returned 1 solely because Git reports the untracked candidate path as deleted relative to the tree; direct byte comparison resolves that bookkeeping limitation. No review source edits remain.
- Initial tag/candidate comparison exit 0, formatter and whitespace checks clean, no extra repository files or LOGBOOK writes.

All acceptance evidence above was rerun or inspected by this reviewer; no producer test result was substituted. Full hosted cross-platform implementation executions remain outside this local review scope. The final conditional checklist item (nonacceptance routing) is not applicable to this accepted verdict.
