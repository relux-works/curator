# TASK-260730-2kfjm8: land-accepted-curator-composite

## Description
Create one reviewable commit from the independently accepted TASK-260720-1pvfj5 final composite at exact origin/main base 17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8, verify its tree against the accepted 374-entry manifest, and land that exact reviewed commit to origin/main without advancing the released protocol pin or including local board/research state.

## Scope
Accepted Curator composite commit, independent verification, and fast-forward main landing

## Acceptance Criteria
The exact independently accepted 374-entry Curator composite is committed without byte drift or unrelated local state, independently reviewed to done, then that exact commit is fast-forward pushed to relux-works/curator main while SPEC_PIN remains rc.3. Tag and GitHub Release are explicitly deferred until a new human command.
