# TASK-261004-3pvg2k — introduce CIPs directory and process

Status: blocked before implementation by repository ownership mismatch.

## Constraint and evidence
The assigned Story worktree has origin relux-works/curator.git. The task requires changes in relux-works/curator-spec.git. The assigned worktree has no GOVERNANCE.md and its Makefile has no validate target. A read-only check of the separate curator-spec checkout confirms GOVERNANCE.md and make validate exist there. That checkout is clean, but it is outside the assigned Story workspace and candidate snapshot.

The task instruction requires every repository change in the assigned workspace and prohibits switching or creating a replacement branch through manual workflow. Editing the sibling checkout would escape the recorded candidate; copying CIP documents into curator would implement the process in the wrong repository. Neither is a valid handoff.

## Attempts and command results
- Initial set_status(development): exit 1; board refused because estimate was missing. Estimate 2 was subsequently requested to resolve that routine prerequisite.
- Scoped task query with resources field: exit 1; unsupported field; retried using checklist and notes.
- Repository identity and clean-state checks: exit 0.
- make validate: NOT RUN; no such target in the assigned repository and no changes made.
- Manifest digest validation: NOT RUN; reference text in the spec release metadata names sha256:6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5, but this is not a recomputed validation result.

No repository files were edited. No protocol, schema, conformance or release changes were made. No LOGBOOK.md edit was made, per the binding task instruction. Acceptance criteria and validation checklist remain unverified.

## Options and recommendation
1. Recommended: have the orchestrator provision and bind this Story to an isolated curator-spec worktree, then rerun the developer there. This preserves candidate ownership, docs scope and make validate evidence.
2. Change the task to target curator instead. This requires a product scope decision and does not satisfy the current acceptance criteria.

Exact external input needed: an orchestrator-provisioned curator-spec Story workspace and matching producer/candidate binding. Do not hand this implementation off as ready for review until it exists and the docs and validation have been produced there.
