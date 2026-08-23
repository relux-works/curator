# STORY-260819-2ymytg: verified-cross-platform-release

## Description
Qualify the three providers together and publish verified mode only after all common and platform evidence is accepted.

## Scope
Cross-platform compatibility matrix, upgrade and rollback, revocation, signing-chain verification, mixed fleet behavior, reproducible release evidence, operator documentation, and release publication. Do not weaken requirements to make one platform pass.

## Acceptance Criteria
macOS, Linux, and Windows providers pass the identical required verified capability suite plus their platform suites; mixed fleets do not downgrade silently; release artifacts and signatures are reproducible and verifiable; rollback and revocation fail safely; documentation accurately separates portable, verified, and future binary admission.
