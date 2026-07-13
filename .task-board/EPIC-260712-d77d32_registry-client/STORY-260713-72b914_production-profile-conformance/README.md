# STORY-260713-72b914: production profile conformance

## Description

Implement the Curator Protocol 1.0 registry-client production profile and run
the authoritative generated client vectors against the Go implementation.

## Scope

- Preserve the deployed registry wire format.
- Enforce snapshot rollback and equivocation state across signing-key rotation.
- Detect deletion of previously established rollback state through a protected
  registry-state catalog.
- Classify retry safety by HTTP method, outcome, and idempotency protection.
- Enforce pagination cursor, record-count, and response-size bounds.
- Execute the pinned `registry-client.json` suite in CI on every supported OS.

## Acceptance Criteria

- Every generated registry-client vector exercises implementation behavior.
- Existing registry and interoperability tests remain green.
- Snapshot state remains keyed by canonical registry URL, not signing key.
- Unsafe publication retries and over-limit responses fail closed.
- Retry execution is bounded to three attempts and preserves exact publication
  bytes and idempotency identity.
