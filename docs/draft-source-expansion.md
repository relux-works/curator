# Skillfile schema 2 source expansion

Schema 2 source declarations are parsed by default in the project path.
Schema 1 keeps its exact meaning, and no on-disk migration is implicit.
Global scope continues to require schema 1.

`manifest.Expand` consumes a parsed schema-2 manifest. Local paths are
literal and relative to the declaring Skillfile; Git aliases use an
already acquired tree in `ExpansionOptions.GitRoots`. Expansion performs
no network operation. The project resolver owns acquisition and supplies
runtime, cache, snapshot, and staging boundaries through `OutputRoots`;
project `.agents`, adapter outputs, and `.git` metadata are excluded.
Boundary paths must be absolute.

Each result retains its selection index, individual declaration,
source-relative directory, observed physical path, and parsed skill
manifest. Collections enumerate immediate directories in UTF-8 byte
order. Explicit literals are checked before exclusion; invalid remaining
packages fail the whole result. Names come from validated `SKILL.md`
YAML and must agree with an individual selector and any optional legacy
manifest name. Duplicate installed names and case-equivalent portable
destinations fail across all direct entries.

`closure.BuildExpanded` connects expansion to the provider-first closure
walk through an explicit `AcquireSelection` boundary. Acquisition freezes
package bytes, validates the acquired spec, and supplies its snapshot. Git
nodes retain repository identity, ref, and commit; all members of one
alias must agree on the commit. Local nodes leave Git fields empty.
Identical transitive Git requirements still unify, including diamonds;
local packages and repository subdirectories cannot be identified by a
legacy root-repository dependency of the same installed name.
Version/source conflicts, cycles, and command-narrowing validation remain
in the existing closure walk. Legacy `closure.Build` refuses unresolved
selectors before acquisition.

## Identity and installation boundary

Expansion paths are observations, not immutable package identities. The
project installation path authenticates local snapshots and Git commits,
freezes membership and package inputs, applies root-input policy, writes
a Skillfile lock and machine-private bindings on explicit resolve, and
consumes those frozen records during install. It does not pass schema-2
nodes to the legacy marker or runtime-store writers. The acquisition
callback is an internal trusted boundary; it is not package-provided code
or attestation evidence.

Tests cover the published schema-2 manifest corpus, `manifest.Expand`,
`closure.BuildExpanded`, and the legacy `Build` refusal path. Production
CLI tests separately cover lock creation, install, status, refresh, and
frozen replay using local fixtures.
