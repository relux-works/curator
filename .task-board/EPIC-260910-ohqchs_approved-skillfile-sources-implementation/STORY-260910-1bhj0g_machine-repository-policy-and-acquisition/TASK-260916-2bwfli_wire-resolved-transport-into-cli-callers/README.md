# TASK-260916-2bwfli: wire-resolved-transport-into-cli-callers

## Description
Wire AcquireNetworkResolved and the SSH credential manager dispatch (delivered by TASK-260910-5nrmtt as library code with no cmd caller) into the production acquisition callers behind the draft/opt-in switch: the Skillfile v2 source acquisition path selects the resolved executor when a machine repository policy (TASK-260910-1o9x1f) is present, legacy lane byte-identical when the switch is off; end-to-end test through the cmd/curator entry point with the fake transport; docs/draft-transport-resolution.md caller section. Final leaf of the story.

## Scope
(define task scope)

## Acceptance Criteria
Production caller exists and is exercised by a cmd-level test for both switch states; legacy lane unchanged (golden); narrow tests and remote gate green.
