# BUG-261003-mjzrnv: freeze-guard-tests-break-on-release-tag

## Description
tools/test_validate.py ReleasedRecordImmutabilityTests assume rc.14 is an untagged candidate; once v1.0.0-rc.14 is tagged the release workflow fails (test_all_six_published_records_are_covered_and_unchanged, test_rc14_candidate_keeps_rc13_as_published_history_anchor).

## Scope
(define bug scope / affected area)

## Acceptance Criteria
Freeze tests pass with and without the active-version tag; guard still fails on a mutated published record; conformance/schemas/release records and manifest digest unchanged.
