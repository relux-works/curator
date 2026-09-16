# Skillfile implementation — wave note (2026-09-16)

Landed and in place:
- curator-spec main a68854d: protocol/skillfile-sources.md, protocol/repository-transport.md, draft-sources-v1 schemas/conformance, docs/skillfile-sources.md (the accepted contract; read it from /Users/administrator/Developer/ReluxWorks/curator/curator-spec).
- curator main: PR72 remote gate (`scripts/remote-gate.sh`) — the configured landing suite for this repository now runs on GitHub (gate/<story>/<stamp> branches) instead of the local host. Do not run the full local suite; narrow package tests only (the host stalls on new test binaries for minutes — keep `go test` invocations to the packages you touched, with `-run` masks where sensible).
- STORY-260910-197y84: parser (TASK-260910-24cuys, checkpointed on the Story branch) + collections (TASK-260910-3kvq02). Until this Story integrates, the other Skillfile Stories' worktrees branch from a main WITHOUT the parser; tasks blocked on 24cuys wait for that landing.

Worker policy: producers muse-spark-1.3-contributor:max, reviewers gpt-6-astra:low. Hand off with an evidence resource that quotes real exit codes; the handoff publishes the Change Request and triggers the remote gate once.
