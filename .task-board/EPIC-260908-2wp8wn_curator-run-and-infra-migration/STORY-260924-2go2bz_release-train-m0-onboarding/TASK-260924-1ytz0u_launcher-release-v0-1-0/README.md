# TASK-260924-1ytz0u: launcher-release-v0-1-0

## Description
Prepare curator-agent-launcher v0.1.0: first tagged release; decide with evidence what a release needs in this repo (it has no release workflow and no tags yet: at least CHANGELOG section, version/specVersion consistency, go install @v0.1.0 working; add a release workflow only if onboarding needs prebuilt binaries — state which); land after the F-L1b Story (STORY-260922-39hxog) lands, then signed tag v0.1.0.

## Scope
(define task scope)

## Acceptance Criteria
release-prep landed with CI green; signed tag v0.1.0; go install ...@v0.1.0 verified
