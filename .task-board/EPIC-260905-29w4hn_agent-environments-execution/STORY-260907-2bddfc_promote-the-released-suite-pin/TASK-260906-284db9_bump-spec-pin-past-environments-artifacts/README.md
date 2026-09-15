# TASK-260906-284db9: bump-spec-pin-past-environments-artifacts

## Description
Move the committed conformance-suite pin from 0ed5c691 to the v1.0.0-rc.11 released revision 87a0d006, so the default Test and Race lanes serve the environments conformance families instead of deferring them. curator currently tests against a root that release/1.0.0-rc.9.json no longer names as required: it demands manifest 0e195ecd, which is the tagged revision, while the committed pin carries 803918bf, and committed_release_pin_advanced is still false. The ci.yml policy holds — the pin carries only a qualified released revision, and v1.0.0-rc.11 is a signed released tag. Every root-unset tolerance the four affected packages carry becomes unreachable once they are served, and a tolerance that cannot fire is a false ledger row, so those come out with it.

## Scope
curator repository, branch chore/promote-spec-pin in worktree /Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-pin, base eca87fe38957ed08f4819836cc6fe12a3efd18dd. Authority curator-spec v1.0.0-rc.11 = 87a0d0060bad64ab883d007dcdf35df7485368bf. Files: .github/workflows/ci.yml, .github/ci/root-artifacts.tsv, .github/ci/platform-cases.tsv.

## Acceptance Criteria
Once a curator-spec release publishes the environments conformance artifacts and that release is qualified, SPEC_PIN moves to it, the two root-artifacts.tsv registrations stop causing a deferral, and the three ledger rows drop their root-unset tolerance so the cases are required to pass on every hosted lane. Until then, every stage that adds a case reading those families dispatches the candidate lane against curator-spec main and records the run id as landing evidence.
