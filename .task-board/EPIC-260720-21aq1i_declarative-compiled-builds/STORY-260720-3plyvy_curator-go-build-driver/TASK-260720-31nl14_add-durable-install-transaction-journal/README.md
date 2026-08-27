# Add a durable install transaction journal

## Description
Implement a manager-home transaction engine that stages deterministic target replacements, records preimages and desired digests durably, commits under the home lock, rolls back in reverse order, and recovers interrupted work.

## Scope
Create internal/transaction journal and target abstractions on top of the manager locks. The canonical journal records transaction id, project identity, ordered target class and identifier, expected generation or preimage digest, backup path, desired digest, referenced build keys, and per-target commit state. Require a caller-held home lock for recovery and commit. Use atomic file or directory swaps plus platform durability primitives, retain backups until the consumer target is durable, compare current desired digests before rollback, and refuse to overwrite unknown concurrent state. This task supplies fault-injectable APIs and tests but does not refactor install.Project or install.Global.

## Acceptance Criteria
Target ordering is stable by class then unsigned bytewise identifier; every state transition and backup needed for crash recovery is durable before the next swap; injected failure at each target boundary restores committed targets in exact reverse order and leaves untouched targets unchanged; a desired-digest mismatch during rollback produces implementation-corruption without overwriting current bytes; restart recovery under the home lock resumes commit or rollback deterministically by transaction id regardless of current project; journal-referenced build keys remain discoverable for GC; successful commit removes backups and journal only after durability; subprocess, fault-injection, and race tests pass.
