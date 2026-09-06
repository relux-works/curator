# TASK-260906-284db9: bump-spec-pin-past-environments-artifacts

## Description


## Scope
Stage (b) review cycle 3 (m6): the three schema drivers TestFragmentAuthoritativeSchemaCases, TestFragmentEmissionMatchesReference and TestParseAuthoritativeEnvMarkerSchemaCases assert nothing on any automatically triggered hosted lane. ci.yml pins SPEC_PIN 0ed5c691, which predates the environments artifacts, so suite-plan defers both packages and the three cases record tolerated root-unset skips on every Test and Race lane. The only lane that serves them is Candidate suite, which is workflow_dispatch and skips on pull_request. They therefore run in CI only when someone dispatches the candidate lane by hand. Promoting SPEC_PIN is owned separately (the comment in ci.yml names TASK-260720-38l1sy after TASK-260720-25d05o qualifies the release), and the environments work is not part of a released revision yet, so this is a scheduling dependency rather than a defect in stage (b).

## Acceptance Criteria
Once a curator-spec release publishes the environments conformance artifacts and that release is qualified, SPEC_PIN moves to it, the two root-artifacts.tsv registrations stop causing a deferral, and the three ledger rows drop their root-unset tolerance so the cases are required to pass on every hosted lane. Until then, every stage that adds a case reading those families dispatches the candidate lane against curator-spec main and records the run id as landing evidence.
