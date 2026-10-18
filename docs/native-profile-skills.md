# Native profile skills

For the effective current profile of each scope, `profile install --use`,
`profile use`, `profile sync`, `profile update`, and `env resolve --repair`
materialize Claude Code, Codex CLI, and pi skills at `<native home>/skills/<name>`.
The links point to immutable profile store entries and the environment marker
records their paths. Status reports native skill drift alongside the scope;
repair restores the current profile's native entries. Repair of an explicit
non-current profile only affects its managed home. A dry-run never writes.

Unrecorded foreign links require `--takeover`; unrecorded files, directories,
and links into the store also require consent. Takeover backs up each entry
before replacement, preserving a symlink's original link text. A symlinked
parent below the native home always refuses traversal. Names outside the
profile remain untouched. Previously recorded `.agent-context/skills` entries
are removed during the next materialization.

An entry named by the native adapter ledger remains owned by the legacy global
scope. Profile switching leaves it in place, excludes it from environment
ownership, and reports it as non-current. A failed ownership-state read refuses
writes. Muse, opencode, and managed-home discovery paths retain their existing
behavior. Muse multi-root ownership, opencode's machine-global surface, and
adapter-ledger precedence require separate specification clarification.

Tests run only against temporary homes. Run the behavioral and mutation controls
with `go test ./internal/envprofile ./cmd/curator -run '^TestNativeSkills' -count=1`.
Build with `go build -o .temp/curator ./cmd/curator`. Gate commands and evidence
locations are documented in [ci-gates.md](ci-gates.md).
