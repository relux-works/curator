# TASK-260920-1sbj7o: install-external-refusal-masks-identity-invalid-class

## Description
Minor (from TASK-260910-1xya7x review rev3 F9): install-level external refusals report build_repository_source_unavailable for a build_repository_identity_invalid plan; the true class is pinned at ValidateTransportPlan. Surface the true class through install.Project and the CLI remediation table.

## Scope
internal/install external planning error classification; cmd/curator remediation rows

## Acceptance Criteria
install.Project and curator install/status report build_repository_identity_invalid for an invalid identity plan; the corresponding crossconformance rows assert the exact class; remediation table row present.
