# THE ONLY CURRENT INSTRUCTION — TASK-261006-3ptm2x: scope the NUL opaque gate to v1 identities (developer, code)
**Why.** curator-spec `protocol/core.md` §8, "Interim rule for v1 readers": "It is a compatibility guard for v1 framing; v2 hashes file bytes, including `0x00`, as ordinary data." Curator main has written v2 identities since the rc.4 core (fae2ff9c; `internal/hashing.EnableV2Writers = true`). But the NUL opaque block still runs unconditionally:
- `internal/audit/audit.go` `gate()` (comment: "NUL-bearing files are opaque under the current v1 content framing"), `auditSubject`/`auditSubjectWithOpaquePaths` and `detect`/`detectWithOpaquePaths`;
- `internal/contextaudit`.
This blocks real installs today, for example a vendored archive containing NUL bytes.

**Do.**
1. **Design note.** Write a short note in your validation resource. Enumerate every call site of `opaquescan.NULPaths` and every place a v1 identity is computed or trusted: legacy markers with no hash_version, v1 readers, frozen v1 shapes, verdict cache keys, context locks. For each site, state the identity version in force at that point and how the code knows it. Never infer the version from digest bytes.
2. **Implement.** Block on NUL exactly when a v1 identity is computed or trusted, and never for a v2 computation or verification. Keep the finding id, severity and messages for the v1 case unchanged. Keep the verdict-cache exclusion for v1 NUL results.
3. **Tests, at the production entry** (`cli.run` / install / audit / status, not just unit tests):
   - (a) a NUL-bearing skill installs, audits and reports current under v2 with a v2 identity;
   - (b) the same tree read through a legacy v1 marker or v1 path still gets the blocking opaque finding;
   - (c) negative: no path computes a v1 identity over NUL;
   - (d) contextaudit follows the same rule.
   List the mutants that would kill each test.
4. Add a CHANGELOG `Unreleased` entry. Do NOT edit LOGBOOK.md or scripts/remote-gate.sh.

**Host rules (attached host-rules.md: R193/R194).**
- Run targeted tests ONLY for `internal/audit`, `internal/contextaudit`, `internal/opaquescan` and `internal/hashing`, through `~/.local/bin/mini-build-lock run opaque -- env GOFLAGS=-work go test ./internal/<pkg> -run '<regex>' -count=1 -timeout=6m`.
- Do NOT run cmd/curator or internal/install locally: the hosted gate is the arbiter. If a sandbox blocks the lock wrapper, list the commands instead.

Then `task-board handoff TASK-261006-3ptm2x --role developer` and END YOUR TURN.
