# TASK-260819-3rqw0c: implement-macos-network-extension-provider

## Description
Implement the Network Extension portion needed for verified network guarantees.

## Scope
Attach network-flow decisions and observations to the same workload session and process identities as Endpoint Security. Cover IPv4, IPv6, TCP, UDP, DNS and local bypass paths required by the threat model, extension ordering, disconnect, sleep and resume, and policy updates.

## Acceptance Criteria
Every required network attempt is enforced and evidenced without gaps on the supported matrix; bypass, race, event-loss, and provider-health negatives fail closed; evidence joins deterministically with process and file observations.
