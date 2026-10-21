# Rework: TASK-261021-jr7tu7, repository-transport covered lanes (curator-spec)

Review r1 (`TASK-261021-jr7tu7_review-r1.md`) requested changes, P1 R1: the text of `protocol/repository-transport.md` is digest-frozen. Its prose is pinned by the rc.14 accepted manifest and the candidate pin, and only the Skillfile-source prose is draft-owned. Any edit there breaks the production release gate: 3 failures and 1 error in the full Python suite.

## Fix

- **Do not edit `protocol/repository-transport.md` at all.** Revert every change to it, and to the frozen conformance manifest, back to `origin/main` byte for byte.
- Keep the clarification **only on the lane side**: the one-sentence back-links in each lane's own section (manager §12.3, §11, environments context/MCP, Skillfile sources), but only in files that are **not** digest-frozen. Find out which files the accepted manifest and the release gate pin, and touch none of them. If the lane sections that need the sentence are themselves frozen, change nothing there and report it: the clarification then needs a proper new revision of the transport protocol, which is the owner's call.
- Verify with the **full** normal-.git Python suite (`make validate` and the release-gate tests) at your head: 879 tests, exit 0, plus the production `release_gate.validate_skillfile_sources_manifest` check.

## Hand-off

Branch `repository-transport-lanes`, a new signed commit, no force-push. Results r3 with the list of frozen files you found. Handover with the head.
