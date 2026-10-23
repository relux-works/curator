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

When `global install` encounters a profile-owned native skill, the profile
remains its sole native publisher. The legacy mirror verifies the environment
marker, profile lock, projected content and live entry, then leaves that entry
out of its adapter ledger. The requested global skill must have the same
canonical source, commit and selected directory; disagreement refuses with
`context_range_conflict` (environments §§1.4 and 9.4). A matching name or content
alone never establishes ownership. Foreign entries still refuse without
takeover. This does not transfer an existing legacy-only entry to a profile.

Direct constraints come from the current global Skillfile and the live profile
closure. A previous lock member is a resolved result, so removing a declaration
retires that member and its native links on the next install unless a live
profile requirement still needs it.

No-flag global install and global add serialize profile and legacy publication
under the same manager-home lock after private builds succeed and their
declaration generation is revalidated. Planning, compilation and verification
retain only the global project operation lock, so a slow compiler leaves profile
sync available. Global add stages its prospective declaration without writing
the Skillfile until the paired transaction. A returned legacy refusal restores the prior
profile lock, rendered documents, native and managed surfaces, backup inventory,
and parent directory modes. An in-place `environment_surface_unmanaged_conflict`
retains the extended lock and its install record under environments §9.4
so first-use `profile sync --takeover` can recover, while restoring the other
mutable profile state. Compensation checks recorded postimages before changing
any target and again before each restoration. Parent directory modes and owner
identities are checkpointed at publication too: a later operator chmod or owner
change is preserved and reported as a conflict, including an empty directory
created by the operation. File modes, owner identities and symlink targets also
participate in postimage comparisons. Parent rollback finishes before profile
recovery records revert. An operator replacement is preserved and reported as a
conflict. Failed compensation stops automatic
generation restarts and retains the extended profile lock and install record;
resolve the conflicting or unwritable path, then explicitly actualize the
pending lock with `profile sync --takeover` and retry the global operation.
Other failures do not inherit the handled-conflict exception. Admitted
immutable store entries may remain. The compensation covers returned errors;
it does not add a cross-phase process-crash journal to the existing individual
profile and legacy transaction recovery mechanisms.

Tests run only against temporary homes. Run the behavioral and mutation controls
with `go test ./internal/envprofile ./cmd/curator -run '^TestNativeSkills' -count=1`.
Build with `go build -o .temp/curator ./cmd/curator`. Gate commands and evidence
locations are documented in [ci-gates.md](ci-gates.md).

`global remove` holds the same canonical global operation lock as install and
add while it reads and writes the declaration. Removal waits for an in-flight
paired installation to hand off; an installation waits for in-flight removal
before resolving its profile. Removal still edits declarations only: the next
install reconciles the installed surfaces, preserving skills required by a
live profile. This prevents a successful paired install from combining a
pre-removal profile closure with a post-removal legacy closure.
