# Brief LOCK: expose the permissions lock in `curator env config show` (TASK-261015-258cwf)

Repository: curator (code + docs). Background: runtimes that have no Curator adapter (qwen, gemini, agy) need the same permissions lock facts as the fragment's `permissions` object `{mode, locked, source}` (fragment v2+, environments sections 10.2 and 12.2, Decision 0018 items 1, 3, 4, 5). Today the only reader path is a read-only `curator env resolve` of some supported environment. The agreed option (a): `curator env config show` reports the effective lock per profile.

Implement:
1. `curator env config show` (human and `--json`/existing JSON form, whatever the command already supports) gains the effective permissions posture per installed profile: `{mode, locked, source}` with exactly the fragment's semantics (`locked: true` iff `source: "global"`, and then `mode: "native"`). Derive it from the same code that fills the fragment's `permissions` member, never a second implementation. A profile whose posture cannot be read reports it as unreadable, never as unlocked.
2. `curator env config show permissions` (single knob form, if the command supports knob arguments) prints the same for the current profile.
3. Tests at the production entry (cmd/curator): locked global, unlocked profile-level, unreadable config; a mutant that reports unlocked on read failure must fail a test.
4. docs/cli.md (and docs/environment-config.md if it lists the output) document the new output; CHANGELOG `## Unreleased` gains one Added line.

Rules: read cmd/curator/envconfig.go and the fragment permissions code first; keep the JSON stable and additive; run `go test ./...` for the touched packages and the repository's lint. Leave the change uncommitted for the Change Request flow. No host names or paths in code, docs or commit text.
