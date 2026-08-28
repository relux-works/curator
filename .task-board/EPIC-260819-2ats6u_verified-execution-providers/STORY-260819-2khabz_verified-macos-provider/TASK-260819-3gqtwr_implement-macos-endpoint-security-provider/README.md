# TASK-260819-3gqtwr: implement-macos-endpoint-security-provider

## Description
Implement the Endpoint Security portion of the signed macOS provider.

## Scope
Authorize and observe the required process and file operations; bind audit tokens, code identity, parentage, paths and file identities to the common session; handle deadlines, muting, cache invalidation, event drops, provider restart, and authenticated IPC without synthesizing missing evidence.

## Acceptance Criteria
Required process and filesystem attempts are losslessly bound to the verified session on supported hosts; deadline, event-loss, disconnect, identity, and restart failures invalidate or reject execution; adversarial and common provider tests pass.
