# TASK-260924-19n6g2: spec-release-v1-0-0-rc-13

## Description
Prepare and publish curator-spec v1.0.0-rc.13 from spec main after PR #88 (erratum) lands: follow exactly how rc.10-rc.12 were cut (tools/release_gate.py, tools/verify_release_commit.py, release metadata, CHANGELOG section), release-prep commit on main via PR with all required checks green, then a signed tag v1.0.0-rc.13 on that main commit and the release workflow green.

## Scope
(define task scope)

## Acceptance Criteria
release-prep PR merged with required checks green; signed tag v1.0.0-rc.13 on a main commit; release workflow green
