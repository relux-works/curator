# Draft source expansion boundary

This is an internal implementation of unreleased `skillfile-sources-v1` §1,
not an enabled schema-2 installation path. The default manifest reader still
rejects schema 2. Frozen v1 and release qualification are unchanged.

`manifest.Expand` consumes a manifest parsed with `DraftSourcesV1`. Local paths
are literal and relative to the declaring Skillfile; Git aliases require an
already acquired tree in `ExpansionOptions.GitRoots`. It performs no network
operation. The acquisition owner supplies machine runtime/cache/snapshot/staging
boundaries through `OutputRoots`; project `.agents`, adapter outputs and `.git`
metadata are excluded internally. Boundary paths must be absolute.

Each result retains the original selection index, individual declaration,
source-relative directory, observed physical path and parsed skill manifest.
Collections enumerate immediate directories in UTF-8 byte order. Explicit
literals are checked before exclusion; invalid remaining packages fail the
whole result. Names come from validated SKILL.md YAML and must agree with an
individual selector and any optional legacy manifest name. Duplicate installed
names and case-equivalent portable destinations fail across all direct entries.

`closure.BuildExpanded` connects expansion to the existing provider-first
closure walk through an explicit `AcquireSelection` boundary. That acquisition
owner must freeze package bytes, validate the acquired spec and supply its
snapshot. Git nodes retain their repository identity/ref/commit, and all members
of one alias must agree on the commit. Local nodes leave Git fields empty.
Identical transitive Git requirements still unify, including diamonds; local
packages and repository subdirectories cannot be identified by a legacy
root-repository dependency of the same installed name. Version/source conflicts,
cycles and command-narrowing validation remain in the existing closure walk.
Legacy `closure.Build` refuses unresolved selectors before acquisition.

## Deliberate integration bounds

Expansion paths are observations, not immutable package identities. This task
adds neither a local snapshot capturer nor a lock writer, transport resolver,
publication path, root-input policy, audit bypass or marker migration. A caller
must implement those independently before enabling schema-2 CLI installation.
In particular, acquisition must revalidate membership and package inputs,
apply root-input policy and recheck destination separation before publication.
Do not pass these nodes to legacy marker or runtime-store writers. The
`AcquireSelection` callback is a trusted internal acquisition interface, not
package-provided code or attestation evidence.

Tests drive `manifest.Expand`, `closure.BuildExpanded`, and the legacy `Build`
refusal path. Acquisition fixtures are immutable temporary directories; tests
of expansion and closure are not evidence of snapshot capture, lock replay,
transport authentication or installation. Those remain separate delivery work.
