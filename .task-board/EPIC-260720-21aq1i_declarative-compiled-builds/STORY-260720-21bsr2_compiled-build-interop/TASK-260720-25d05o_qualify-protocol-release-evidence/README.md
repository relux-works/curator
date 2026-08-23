# Qualify the released schema v6 protocol suite

## Description
Establish a second immutable evidence gate after released manager revisions pass the in-tree specification workflow and before Curator or csk advances its protocol-suite pin.

## Scope
Perform a read-only qualification after TASK-260720-vs6den and the actual curator-spec schema v6 release. Resolve the annotated signed release tag to its full commit, verify the published conformance archive and checksum against the repository suite manifest, and collect required specification CI including both pinned managers and the black-box parity gate. Attach a task-scoped evidence record. Do not edit manager workflows, suite pins, claims, tags, release notes, checksums, or artifacts. A release candidate plan, branch, mutable tag, local archive, or successful pre-release workflow is not a published protocol release.

## Acceptance Criteria
The outcome records the actual protocol version and tag, full commit ID, release URL and timestamp, tag or artifact signature result required by governance, conformance archive checksum, generated suite SHA-256, and passing required Linux, macOS, and Windows implementation-conformance records. The archive suite is byte-identical to the tagged repository suite and its manifest includes the compiled-build fixture and cases. Curator and csk refs in the release gate match TASK-260720-3pvihp and pass without skipped or xfailed compiled-build cases. No manager suite pin changed before this qualification. If the release, signature, artifact, digest, or CI evidence is missing or inconsistent, the task records the exact missing external item and transitions to blocked without guessing a ref or fabricating conformance evidence.
